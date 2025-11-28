package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"

	"github.com/Fearcon14/pokedexCLI/internal/pokecache"
)

type config struct {
	NextURL     *string
	PreviousURL *string
	Cache       *pokecache.Cache
	Pokedex     map[string]Pokemon
}

type cliCommand struct {
	name        string
	description string
	callback    func(*config, []string) error
}

type locationAreaResponse struct {
	Count    int     `json:"count"`
	Next     *string `json:"next"`
	Previous *string `json:"previous"`
	Results  []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"results"`
}

type locationAreaDetailResponse struct {
	PokemonEncounters []struct {
		Pokemon struct {
			Name string `json:"name"`
		} `json:"pokemon"`
	} `json:"pokemon_encounters"`
}

type pokemonDetailResponse struct {
	Name           string `json:"name"`
	BaseExperience int    `json:"base_experience"`
}

type Pokemon struct {
	Name           string
	BaseExperience int
}

var commands = map[string]cliCommand{
	"exit": {
		name:        "exit",
		description: "Exit the Pokedex",
		callback:    commandExit,
	},
	"help": {
		name:        "help",
		description: "Show the help menu",
		callback:    commandHelp,
	},
	"map": {
		name:        "map",
		description: "Get the next page of locations",
		callback:    commandMap,
	},
	"mapb": {
		name:        "mapb",
		description: "Get the previous page of locations",
		callback:    commandMapb,
	},
	"explore": {
		name:        "explore",
		description: "Explore a location",
		callback:    commandExplore,
	},
	"catch": {
		name:        "catch",
		description: "Catch a Pokemon",
		callback:    commandCatch,
	},
}

func commandExit(cfg *config, args []string) error {
	_ = args
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(cfg *config, args []string) error {
	_ = args
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	fmt.Println("help: Displays a help message")
	fmt.Println("exit: Exits the Pokedex")
	fmt.Println("map: Get the next page of locations")
	fmt.Println("mapb: Get the previous page of locations")
	fmt.Println("explore <location-name>: Explore a location and see Pokemon")
	fmt.Println("catch <pokemon-name>: Attempt to catch a Pokemon")
	return nil
}

func commandMap(cfg *config, args []string) error {
	_ = args
	url := "https://pokeapi.co/api/v2/location-area/"
	if cfg.NextURL != nil {
		url = *cfg.NextURL
	}

	res, err := fetchLocationAreas(url, cfg.Cache)
	if err != nil {
		return err
	}

	cfg.NextURL = res.Next
	cfg.PreviousURL = res.Previous

	for _, area := range res.Results {
		fmt.Println(area.Name)
	}

	return nil
}

func commandMapb(cfg *config, args []string) error {
	_ = args
	if cfg.PreviousURL == nil {
		fmt.Println("You're on the first page")
		return nil
	}

	res, err := fetchLocationAreas(*cfg.PreviousURL, cfg.Cache)
	if err != nil {
		return err
	}

	cfg.NextURL = res.Next
	cfg.PreviousURL = res.Previous

	for _, area := range res.Results {
		fmt.Println(area.Name)
	}

	return nil
}

func commandExplore(cfg *config, args []string) error {
	_ = args
	if len(args) != 1 {
		return fmt.Errorf("explore command requires a location name")
	}
	location := args[0]
	url := fmt.Sprintf("https://pokeapi.co/api/v2/location-area/%s", location)
	res, err := fetchLocationAreaDetail(url, cfg.Cache)
	if err != nil {
		return err
	}

	fmt.Printf("Exploring %s...\n", location)
	fmt.Println("Found Pokemon:")
	for _, encounter := range res.PokemonEncounters {
		fmt.Printf("  - %s\n", encounter.Pokemon.Name)
	}

	return nil
}

func commandCatch(cfg *config, args []string) error {
	_ = args
	if len(args) != 1 {
		return fmt.Errorf("catch command requires a Pokemon name")
	}
	pokemonName := args[0]

	url := fmt.Sprintf("https://pokeapi.co/api/v2/pokemon/%s", pokemonName)
	pokemon, err := fetchPokemon(url, cfg.Cache)
	if err != nil {
		return err
	}

	fmt.Printf("Throwing a Pokeball at %s...\n", pokemonName)

	// Calculate catch chance based on base experience
	// Higher base experience = harder to catch
	// Base experience typically ranges from ~50 to ~300+
	// We'll use: catchThreshold = max(0, 100 - baseExperience/10)
	// This gives us a range where low exp (50) = 95% chance, high exp (300) = 70% chance
	catchThreshold := 100 - (pokemon.BaseExperience / 10)
	if catchThreshold < 0 {
		catchThreshold = 0
	}
	if catchThreshold > 100 {
		catchThreshold = 100
	}

	randomValue := rand.Intn(100)

	if randomValue < catchThreshold {
		fmt.Printf("%s was caught!\n", pokemonName)
		if _, ok := cfg.Pokedex[pokemonName]; ok {
			fmt.Printf("%s is already in your Pokedex!\n", pokemonName)
		} else {
			cfg.Pokedex[pokemonName] = Pokemon{
				Name:           pokemon.Name,
				BaseExperience: pokemon.BaseExperience,
			}
		}
	} else {
		fmt.Printf("%s escaped!\n", pokemonName)
	}

	return nil
}

func fetchLocationAreas(url string, cache *pokecache.Cache) (*locationAreaResponse, error) {
	if cachedData, ok := cache.Get(url); ok {
		var locationRes locationAreaResponse
		err := json.Unmarshal(cachedData, &locationRes)
		if err != nil {
			return nil, err
		}
		return &locationRes, nil
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "PokedexCLI")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	cache.Add(url, body)

	var locationRes locationAreaResponse
	err = json.Unmarshal(body, &locationRes)
	if err != nil {
		return nil, err
	}

	return &locationRes, nil
}

func fetchLocationAreaDetail(url string, cache *pokecache.Cache) (*locationAreaDetailResponse, error) {
	if cachedData, ok := cache.Get(url); ok {
		var locationDetail locationAreaDetailResponse
		err := json.Unmarshal(cachedData, &locationDetail)
		if err != nil {
			return nil, err
		}
		return &locationDetail, nil
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "PokedexCLI")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	cache.Add(url, body)

	var locationDetail locationAreaDetailResponse
	err = json.Unmarshal(body, &locationDetail)
	if err != nil {
		return nil, err
	}

	return &locationDetail, nil
}

func fetchPokemon(url string, cache *pokecache.Cache) (*pokemonDetailResponse, error) {
	if cachedData, ok := cache.Get(url); ok {
		var pokemon pokemonDetailResponse
		err := json.Unmarshal(cachedData, &pokemon)
		if err != nil {
			return nil, err
		}
		return &pokemon, nil
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "PokedexCLI")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	cache.Add(url, body)

	var pokemon pokemonDetailResponse
	err = json.Unmarshal(body, &pokemon)
	if err != nil {
		return nil, err
	}

	return &pokemon, nil
}
