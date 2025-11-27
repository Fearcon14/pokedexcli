package main

import (
	"fmt"
	"bufio"
	"os"
)

func main() {
	for {
		fmt.Print("Pokedex > ")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		text := scanner.Text()
		cleaned := cleanInput(text)
		if len(cleaned) == 0 {
			continue
		}
		command := cleaned[0]
		fmt.Printf("Your command was: %s\n", command)
	}
}
