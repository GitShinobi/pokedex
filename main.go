package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"pokedex/internal/pokecache"
	"time"
)

type cliCommand struct {
	name        string
	description string
	callback    func(*config, string) error
}

type config struct {
	registry map[string]cliCommand
	location locationAreaResponse
	cache    pokecache.Cache
	pokemons map[string]Pokemon
}
type locationArea struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type locationAreaResponse struct {
	Next     string         `json:"next"`
	Previous string         `json:"previous"`
	Results  []locationArea `json:"results"`
}
type locationEndpointResponse struct {
	Encounters []encounters `json:"pokemon_encounters"`
}
type pokemon_info struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}
type encounters struct {
	Pokemon pokemon_info `json:"pokemon"`
}

type Pokemon struct {
	Id int `json:"id"`
	Name string `json:"name"`
	Experience int `json:"base_experience"`
	Height int `json:"height"`
	Weight int `json:"weight"`
	Stats []stats `json:"stats"`
	Types []types `json:"types"`
}
type stats struct {
	Base int `json:"base_stat"`
	Effort int `json:"effort"`
	Stat pokemon_info `json:"stat"`
}
type types struct {
	Slot int `json:"slot"`
	Type pokemon_info `json:"type"`
}

func main() {
	configVar := config{
		map[string]cliCommand{
			"exit": {
				name:        "exit",
				description: "Exit the Pokedex",
				callback:    commandExit,
			},
			"help": {
				name:        "help",
				description: "Displays a help message",
				callback:    commandHelp,
			},
			"map": {
				name:        "map",
				description: "Display the initial/next 20 locations",
				callback:    commandMap,
			},
			"explore": {
				name:        "explore",
				description: "Display the list of all the Pokémon located ",
				callback:    commandExplore,
			},
			"mapb": {
				name:        "mapb",
				description: "Display the previous 20 locations",
				callback:    commandMapb,
			},
			"catch": {
				name:        "catch",
				description: "catch pokemon",
				callback:    commandCatch ,
			},
			"inspect": {
				name:        "inspect",
				description: "Display if caugth pokemon",
				callback:    commandInspect ,
			},
			"pokedex": {
				name:        "pokedex",
				description: "Display all caugth pokemon",
				callback:    commandPokedex  ,
			},
		},
		locationAreaResponse{},
		*pokecache.NewCache(5 * time.Second),
		map[string]Pokemon{},
	}
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Pokedex > ")
		if !scanner.Scan() {
			break
		}
		input := scanner.Text()
		if len(input) == 0 {
			continue
		}
		cleanedInput := cleanInput(input)
		cmd := cleanedInput[0]
		cmd_struct, ok := configVar.registry[cmd]
		if ok {
			var arg string
			if cmd == "explore" || cmd == "catch" || cmd == "inspect" {
				if len(cleanedInput) <= 1 {
					fmt.Printf("command %s arg required\n",cmd)
					continue
				}
				arg = cleanedInput[1]
			}

			if err := cmd_struct.callback(&configVar, arg); err != nil {
				fmt.Println(err)
			}
		} else {
			fmt.Println("Unknown command")
		}

	}

}

func commandExit(*config, string) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(configVar *config, arg string) error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	for cmd := range (*configVar).registry {
		fmt.Printf("%s: %s\n", cmd, (*configVar).registry[cmd].description)
	}
	return nil
}
func commandMap(configVar *config, arg string) error {
	var url string
	if (*configVar).location.Next == "" {
		url = "https://pokeapi.co/api/v2/location-area"
	} else {
		url = (*configVar).location.Next
	}
	body, ok := (*configVar).cache.Get(url)
	if !ok {
		resp, err := http.Get(url)
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("request failed: %s", resp.Status)
		}
		defer resp.Body.Close()
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		(*configVar).cache.Add(url, body)
	} else {
		fmt.Println("cached")
	}

	var maplist locationAreaResponse
	if err := json.Unmarshal(body, &maplist); err != nil {
		return err
	}
	(*configVar).location.Previous = maplist.Previous
	(*configVar).location.Next = maplist.Next
	locations := maplist.Results
	for _, location := range locations {
		fmt.Println(location.Name)
	}
	return nil
}
func commandMapb(configVar *config, arg string) error {
	if (*configVar).location.Previous == "" {
		return fmt.Errorf("you're on the first page")
	}
	url := (*configVar).location.Previous
	body, ok := (*configVar).cache.Get(url)
	if !ok {
		resp, err := http.Get(url)
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("request failed: %s", resp.Status)
		}
		defer resp.Body.Close()
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		(*configVar).cache.Add(url, body)
	} else {
		fmt.Println("cached")
	}
	var maplist locationAreaResponse
	if err := json.Unmarshal(body, &maplist); err != nil {
		return err
	}
	(*configVar).location.Previous = maplist.Previous
	(*configVar).location.Next = maplist.Next
	locations := maplist.Results
	for _, location := range locations {
		fmt.Println(location.Name)
	}
	return nil
}
func commandExplore(configVar *config, endpoint string) error {
	url := fmt.Sprintf("https://pokeapi.co/api/v2/location-area/%s", endpoint)
	body, ok := (*configVar).cache.Get(url)
	if !ok {
		resp, err := http.Get(url)
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("request failed: %s", resp.Status)
		}
		defer resp.Body.Close()
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		(*configVar).cache.Add(url, body)
	} else {
		fmt.Println("cached")
	}
	var encounters locationEndpointResponse
	if err := json.Unmarshal(body, &encounters); err != nil {
		return err
	}
	fmt.Printf("Exploring %s...\n", endpoint)
	fmt.Println("Found Pokemon:")
	for _, Encounter := range encounters.Encounters {
		fmt.Printf(" - %s\n",Encounter.Pokemon.Name)

	}
	return nil
}
func commandCatch(configVar *config, name string) error {
	fmt.Printf("Throwing a Pokeball at %s...\n",name)
	url := fmt.Sprintf("https://pokeapi.co/api/v2/pokemon/%s", name)
	body, ok := (*configVar).cache.Get(url)
	if !ok {
		resp, err := http.Get(url)
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("request failed: %s", resp.Status)
		}
		defer resp.Body.Close()
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		(*configVar).cache.Add(url, body)
	} else {
		fmt.Println("cached")
	}
	var pokemon Pokemon
	if err := json.Unmarshal(body, &pokemon); err != nil {
		return err
	}
	res := rand.Intn(pokemon.Experience)
	if res >= pokemon.Experience / 2 {
		(*configVar).pokemons[name] = pokemon
		fmt.Printf("%s was caught!\n", name)

	}else{
		fmt.Printf("%s escaped!\n", name)
	}
	
	return nil
}

func commandInspect(configVar *config, name string) error {
	pokemon,ok := (*configVar).pokemons[name]
	if !ok{
		fmt.Println("you have not caught that pokemon")
		return nil
	}
		fmt.Printf("Name: %s\n",pokemon.Name)
		fmt.Printf("Height: %d\n",pokemon.Height)
		fmt.Printf("Weight: %d\n",pokemon.Weight)
		fmt.Println("Stats:")
		for _ , stat := range pokemon.Stats{
			fmt.Printf("\t-effort: %d\n",stat.Effort)
			fmt.Printf("\t-speed: %d\n",stat.Base)
			break
		}
		fmt.Println("Types:")
		for _ , Types := range pokemon.Types{
			fmt.Printf("\t- %s\n",Types.Type.Name)
			break
		}
		return nil
}
func commandPokedex(configVar *config, name string) error {
	pokemon := (*configVar).pokemons
	for k := range pokemon{
			fmt.Printf("\t- %s\n",k)
		}
	if len(pokemon)==0{
			fmt.Printf("no pokemon caougth")
	}
	return nil
}
