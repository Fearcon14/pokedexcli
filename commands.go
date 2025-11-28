package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/Fearcon14/pokedexCLI/internal/pokecache"
)

type config struct {
	NextURL     *string
	PreviousURL *string
	Cache       *pokecache.Cache
}

type cliCommand struct {
	name        string
	description string
	callback    func(*config) error
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
}

func commandExit(cfg *config) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(cfg *config) error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	fmt.Println("help: Displays a help message")
	fmt.Println("exit: Exits the Pokedex")
	fmt.Println("map: Get the next page of locations")
	fmt.Println("mapb: Get the previous page of locations")
	return nil
}

func commandMap(cfg *config) error {
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

func commandMapb(cfg *config) error {
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
