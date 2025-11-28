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
	ID             int    `json:"id"`
	Name           string `json:"name"`
	BaseExperience int    `json:"base_experience"`
	Height         int    `json:"height"`
	Weight         int    `json:"weight"`
	Stats          []struct {
		BaseStat int `json:"base_stat"`
		Effort   int `json:"effort"`
		Stat     struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"stat"`
	} `json:"stats"`
	Types []struct {
		Slot int `json:"slot"`
		Type struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"type"`
	} `json:"types"`
	Abilities []struct {
		IsHidden bool `json:"is_hidden"`
		Slot     int  `json:"slot"`
		Ability  struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"ability"`
	} `json:"abilities"`
	Moves []struct {
		Move struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"move"`
		VersionGroupDetails []struct {
			LevelLearnedAt int `json:"level_learned_at"`
			VersionGroup   struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"version_group"`
			MoveLearnMethod struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"move_learn_method"`
		} `json:"version_group_details"`
	} `json:"moves"`
	Sprites struct {
		BackDefault      string `json:"back_default"`
		BackFemale       string `json:"back_female"`
		BackShiny        string `json:"back_shiny"`
		BackShinyFemale  string `json:"back_shiny_female"`
		FrontDefault     string `json:"front_default"`
		FrontFemale      string `json:"front_female"`
		FrontShiny       string `json:"front_shiny"`
		FrontShinyFemale string `json:"front_shiny_female"`
	} `json:"sprites"`
	Order   int `json:"order"`
	Species struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"species"`
	Forms                  []interface{} `json:"forms"`
	GameIndices            []interface{} `json:"game_indices"`
	HeldItems              []interface{} `json:"held_items"`
	LocationAreaEncounters string        `json:"location_area_encounters"`
	IsDefault              bool          `json:"is_default"`
}

type Pokemon struct {
	ID             int
	Name           string
	BaseExperience int
	Height         int
	Weight         int
	Stats          []struct {
		BaseStat int
		Effort   int
		Stat     struct {
			Name string
			URL  string
		}
	}
	Types []struct {
		Slot int
		Type struct {
			Name string
			URL  string
		}
	}
	Abilities []struct {
		IsHidden bool
		Slot     int
		Ability  struct {
			Name string
			URL  string
		}
	}
	Moves []struct {
		Move struct {
			Name string
			URL  string
		}
		VersionGroupDetails []struct {
			LevelLearnedAt int
			VersionGroup   struct {
				Name string
				URL  string
			}
			MoveLearnMethod struct {
				Name string
				URL  string
			}
		}
	}
	Sprites struct {
		BackDefault      string
		BackFemale       string
		BackShiny        string
		BackShinyFemale  string
		FrontDefault     string
		FrontFemale      string
		FrontShiny       string
		FrontShinyFemale string
	}
	Order   int
	Species struct {
		Name string
		URL  string
	}
	Forms                  []interface{}
	GameIndices            []interface{}
	HeldItems              []interface{}
	LocationAreaEncounters string
	IsDefault              bool
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
	"inspect": {
		name:        "inspect",
		description: "Inspect a Pokemon",
		callback:    commandInspect,
	},
	"pokedex": {
		name:        "pokedex",
		description: "Show the Pokedex",
		callback:    commandPokedex,
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
	fmt.Println("inspect <pokemon-name>: Inspect a Pokemon")
	fmt.Println("pokedex: Show the Pokedex")
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
			cfg.Pokedex[pokemonName] = convertToPokemon(pokemon)
		}
	} else {
		fmt.Printf("%s escaped!\n", pokemonName)
	}

	return nil
}

func commandInspect(cfg *config, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("inspect command requires a Pokemon name")
	}
	pokemonName := args[0]
	if _, ok := cfg.Pokedex[pokemonName]; !ok {
		return fmt.Errorf("you have not caught that pokemon: %s", pokemonName)
	}
	fmt.Printf("Name: %s\n", cfg.Pokedex[pokemonName].Name)
	fmt.Printf("Height: %d\n", cfg.Pokedex[pokemonName].Height)
	fmt.Printf("Weight: %d\n", cfg.Pokedex[pokemonName].Weight)
	fmt.Printf("Stats:\n")
	for _, stat := range cfg.Pokedex[pokemonName].Stats {
		fmt.Printf("  - %s: %d\n", stat.Stat.Name, stat.BaseStat)
	}
	fmt.Printf("Types:\n")
	for _, t := range cfg.Pokedex[pokemonName].Types {
		fmt.Printf("  - %s\n", t.Type.Name)
	}
	return nil
}

func commandPokedex(cfg *config, args []string) error {
	_ = args
	fmt.Println("Your Pokedex:")
	for _, pokemon := range cfg.Pokedex {
		fmt.Printf("  - %s\n", pokemon.Name)
	}
	return nil
}

func convertToPokemon(p *pokemonDetailResponse) Pokemon {
	pokemon := Pokemon{
		ID:                     p.ID,
		Name:                   p.Name,
		BaseExperience:         p.BaseExperience,
		Height:                 p.Height,
		Weight:                 p.Weight,
		Order:                  p.Order,
		LocationAreaEncounters: p.LocationAreaEncounters,
		IsDefault:              p.IsDefault,
		Forms:                  p.Forms,
		GameIndices:            p.GameIndices,
		HeldItems:              p.HeldItems,
	}

	// Copy Sprites
	pokemon.Sprites = struct {
		BackDefault      string
		BackFemale       string
		BackShiny        string
		BackShinyFemale  string
		FrontDefault     string
		FrontFemale      string
		FrontShiny       string
		FrontShinyFemale string
	}{
		BackDefault:      p.Sprites.BackDefault,
		BackFemale:       p.Sprites.BackFemale,
		BackShiny:        p.Sprites.BackShiny,
		BackShinyFemale:  p.Sprites.BackShinyFemale,
		FrontDefault:     p.Sprites.FrontDefault,
		FrontFemale:      p.Sprites.FrontFemale,
		FrontShiny:       p.Sprites.FrontShiny,
		FrontShinyFemale: p.Sprites.FrontShinyFemale,
	}

	// Copy Species
	pokemon.Species = struct {
		Name string
		URL  string
	}{
		Name: p.Species.Name,
		URL:  p.Species.URL,
	}

	// Copy stats
	pokemon.Stats = make([]struct {
		BaseStat int
		Effort   int
		Stat     struct {
			Name string
			URL  string
		}
	}, len(p.Stats))
	for i, stat := range p.Stats {
		pokemon.Stats[i] = struct {
			BaseStat int
			Effort   int
			Stat     struct {
				Name string
				URL  string
			}
		}{
			BaseStat: stat.BaseStat,
			Effort:   stat.Effort,
			Stat: struct {
				Name string
				URL  string
			}{
				Name: stat.Stat.Name,
				URL:  stat.Stat.URL,
			},
		}
	}

	// Copy types
	pokemon.Types = make([]struct {
		Slot int
		Type struct {
			Name string
			URL  string
		}
	}, len(p.Types))
	for i, t := range p.Types {
		pokemon.Types[i] = struct {
			Slot int
			Type struct {
				Name string
				URL  string
			}
		}{
			Slot: t.Slot,
			Type: struct {
				Name string
				URL  string
			}{
				Name: t.Type.Name,
				URL:  t.Type.URL,
			},
		}
	}

	// Copy abilities
	pokemon.Abilities = make([]struct {
		IsHidden bool
		Slot     int
		Ability  struct {
			Name string
			URL  string
		}
	}, len(p.Abilities))
	for i, ability := range p.Abilities {
		pokemon.Abilities[i] = struct {
			IsHidden bool
			Slot     int
			Ability  struct {
				Name string
				URL  string
			}
		}{
			IsHidden: ability.IsHidden,
			Slot:     ability.Slot,
			Ability: struct {
				Name string
				URL  string
			}{
				Name: ability.Ability.Name,
				URL:  ability.Ability.URL,
			},
		}
	}

	// Copy moves
	pokemon.Moves = make([]struct {
		Move struct {
			Name string
			URL  string
		}
		VersionGroupDetails []struct {
			LevelLearnedAt int
			VersionGroup   struct {
				Name string
				URL  string
			}
			MoveLearnMethod struct {
				Name string
				URL  string
			}
		}
	}, len(p.Moves))
	for i, move := range p.Moves {
		pokemon.Moves[i].Move = struct {
			Name string
			URL  string
		}{
			Name: move.Move.Name,
			URL:  move.Move.URL,
		}
		pokemon.Moves[i].VersionGroupDetails = make([]struct {
			LevelLearnedAt int
			VersionGroup   struct {
				Name string
				URL  string
			}
			MoveLearnMethod struct {
				Name string
				URL  string
			}
		}, len(move.VersionGroupDetails))
		for j, vgd := range move.VersionGroupDetails {
			pokemon.Moves[i].VersionGroupDetails[j] = struct {
				LevelLearnedAt int
				VersionGroup   struct {
					Name string
					URL  string
				}
				MoveLearnMethod struct {
					Name string
					URL  string
				}
			}{
				LevelLearnedAt: vgd.LevelLearnedAt,
				VersionGroup: struct {
					Name string
					URL  string
				}{
					Name: vgd.VersionGroup.Name,
					URL:  vgd.VersionGroup.URL,
				},
				MoveLearnMethod: struct {
					Name string
					URL  string
				}{
					Name: vgd.MoveLearnMethod.Name,
					URL:  vgd.MoveLearnMethod.URL,
				},
			}
		}
	}

	return pokemon
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
