package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"slices"
	"strconv"
	"time"
)

type intComparison int

// This is for the numerical fields: GEN, HEIGHT, WEIGHT
const (
	lesser  intComparison = 0
	equals  intComparison = 1
	greater intComparison = 2
)

type typeComparison int

// This const is just for the Pokemon TYPE1 and TYPE2 fields
const (
	no_type_match  typeComparison = 0
	type_match     typeComparison = 1
	wrong_position typeComparison = 2
)

type guess struct {
	Gen    intComparison
	Type1  typeComparison
	Type2  typeComparison
	Height intComparison
	Weight intComparison
}

const numberOfPokemon = 1181

type pokemon struct {
	Name   string
	Gen    int
	Type1  string
	Type2  string
	Height int //This value should be the height * 10 to get rid of the nasty decimal
	Weight int //This value should be the height * 10 to get rid of the nasty decimal
}

type outputData struct {
	Name    string
	Entropy float64
}

var allGuesses [242]guess = generateAllGuesses()

func checkIntValue(guessPokemonValue int, queryingPokemonValue int, expectedOperation intComparison) bool {
	switch expectedOperation {
	case lesser:
		return queryingPokemonValue < guessPokemonValue
	case equals:
		return queryingPokemonValue == guessPokemonValue
	case greater:
		return queryingPokemonValue > guessPokemonValue
	default:
		log.Fatal("Expected OPERATION CAN NOT USE THE VALUE", expectedOperation)
		return false
	}

}

func outputCompare(a, b outputData) int {
	const epsilon = 1e-9
	diff := a.Entropy - b.Entropy
	if math.Abs(diff) < epsilon {
		return 0
	} else if diff > 0 {
		return -1
	} else {
		return 1
	}
}

func checkTypeValues(guessCurrentType string, queryingCurrentType string, queryingOtherType string, expectedOperation typeComparison) bool {
	switch expectedOperation {
	case no_type_match:
		return guessCurrentType != queryingCurrentType && guessCurrentType != queryingOtherType
	case type_match:
		return guessCurrentType == queryingCurrentType
	case wrong_position:
		return guessCurrentType == queryingOtherType
	default:
		log.Fatal("Expected OPERATION CAN NOT USE THE VALUE", expectedOperation)
		return false
	}

}

func checkIfPokemonMatchesComparison(guessPokemon pokemon, queryPokemon pokemon, g guess) bool {
	return checkIntValue(guessPokemon.Gen, queryPokemon.Gen, g.Gen) &&
		checkTypeValues(guessPokemon.Type1, queryPokemon.Type1, queryPokemon.Type2, g.Type1) &&
		checkTypeValues(guessPokemon.Type2, queryPokemon.Type2, queryPokemon.Type1, g.Type2) &&
		checkIntValue(guessPokemon.Height, queryPokemon.Height, g.Height) &&
		checkIntValue(guessPokemon.Weight, queryPokemon.Weight, g.Weight)
}

func generateAllGuesses() [242]guess {
	// 3^5 -1 guesses (the guess where everything is correct won't be counted)
	var guesses [242]guess
	i := 0
	for gen := lesser; gen <= greater; gen++ {
		for type1 := no_type_match; type1 <= wrong_position; type1++ {
			for type2 := no_type_match; type2 <= wrong_position; type2++ {
				for height := lesser; height <= greater; height++ {
					for weight := lesser; weight <= greater; weight++ {
						if gen != 1 || type1 != 1 || type2 != 1 || height != 1 || weight != 1 {
							guesses[i] = guess{
								Gen:    gen,
								Type1:  type1,
								Type2:  type2,
								Height: height,
								Weight: weight,
							}
							i++
						}

					}
				}
			}
		}
	}
	return guesses
}

func getGetEntropyForGuess(guessPokemon pokemon, listOfValidPokemon []pokemon, g guess) float64 {
	count := 0
	size := len(listOfValidPokemon)
	for _, p := range listOfValidPokemon {
		if checkIfPokemonMatchesComparison(guessPokemon, p, g) {
			count++
		}
	}

	if count > 0 {
		return float64(count) / float64(size) * math.Log2(float64(count)/float64(size))
	}
	return 0
}

func getEntropy(guessPokemon pokemon, pokemons []pokemon) float64 {
	var entropy float64 = 0
	for _, g := range allGuesses {

		entropy += getGetEntropyForGuess(guessPokemon, pokemons, g)
	}
	return entropy
}

func getGuessTotal(guessPokemon pokemon, listOfValidPokemon []pokemon, g guess) []pokemon {
	var output []pokemon
	for _, p := range listOfValidPokemon {
		if checkIfPokemonMatchesComparison(guessPokemon, p, g) {
			output = append(output, p)
		}
	}
	return output
}

func getEntropyRec(guessPokemon pokemon, pokemons []pokemon, numberOfGuesses int) float64 {
	var entropy float64
	var totalPokemons = pokemons
	numberOfGuesses--
	for _, g := range allGuesses {
		var pokemonsSubset = getGuessTotal(guessPokemon, pokemons, g)
		if len(pokemonsSubset) > 0 {
			prob := float64(len(pokemonsSubset)) / float64(len(totalPokemons))
			entropy += -1 * prob * math.Log2(prob)
			var bestSubGuess float64
			if numberOfGuesses > 0 {
				for _, p := range pokemonsSubset {
					var currentSubEntropy = getEntropyRec(p, pokemonsSubset, numberOfGuesses)
					if bestSubGuess < currentSubEntropy {
						bestSubGuess = currentSubEntropy
					}
				}
				entropy += prob * bestSubGuess
			}
		}

	}
	return entropy
}

func calculateEntropy(pokemon pokemon, pokemons []pokemon, numberOfGuesses int) outputData {
	entropy := getEntropyRec(pokemon, pokemons, numberOfGuesses)
	return outputData{Name: pokemon.Name, Entropy: entropy}
}

func writeToCSV(output []outputData){
	// Create a new CSV file
    file, err := os.Create("results.csv")
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()

    // Initialize the CSV writer
    writer := csv.NewWriter(file)
    defer writer.Flush()

    // Write the header
    header := []string{"Pokemon Name", "Entropy", "Average Pokemon Remaining"}
    if err := writer.Write(header); err != nil {
        log.Fatal(err)
    }

    // Write the data
    for _, record := range output {
        row := []string{
            record.Name,
            strconv.FormatFloat(record.Entropy, 'f', -1, 64),
            strconv.FormatFloat((numberOfPokemon * math.Pow(2, (-1 * record.Entropy))), 'f', -1, 64),
        }
        if err := writer.Write(row); err != nil {
            log.Fatal(err)
        }
    }
}

func main() {
	start := time.Now()
	fileBytes, _ := os.ReadFile("pokedex.json")
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(fileBytes), &data); err != nil {
		fmt.Println("Error:", err)
		return
	}

	var pokemons [numberOfPokemon]pokemon
	i := 0
	for name, values := range data {
		pokemonData := values.([]interface{})
		mon := pokemon{
			Name:  name,
			Gen:   int(math.Round(pokemonData[0].(float64))),
			Type1: pokemonData[1].(string),
			Type2: pokemonData[2].(string),
			//get rid of pesky decimal place
			Height: int(math.Round(pokemonData[3].(float64) * 10)),
			Weight: int(math.Round(pokemonData[4].(float64) * 10)),
		}

		pokemons[i] = mon
		i++
	}

	fmt.Println(len(pokemons))

	var output []outputData

	for i, p := range pokemons {
		fmt.Printf("Calculating index %v out of %v\n", i, numberOfPokemon)
		output = append(output, calculateEntropy(p, pokemons[:], 1))
	}

	slices.SortFunc(output, outputCompare)

	writeToCSV(output)

	elapsed := time.Since(start)
	fmt.Printf("Entropy function took %s\n", elapsed)
}
