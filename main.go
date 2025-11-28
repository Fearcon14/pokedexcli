package main

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/Fearcon14/pokedexCLI/internal/pokecache"
)

func main() {
	cfg := &config{
		Cache: pokecache.NewCache(5 * time.Second),
	}
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Pokedex > ")
		if !scanner.Scan() {
			break
		}
		text := scanner.Text()
		cleaned := cleanInput(text)
		if len(cleaned) == 0 {
			continue
		}
		input := cleaned[0]
		if command, ok := commands[input]; ok {
			err := command.callback(cfg, cleaned[1:])
			if err != nil {
				fmt.Println(err)
			}
		} else {
			fmt.Println("Unknown command:", input)
		}
	}
}
