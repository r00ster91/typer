package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/agnivade/levenshtein"
	"rsc.io/quote"
)

// The source of texts to be typed.
var text = []string{
	quote.Glass(), quote.Go(), quote.Opt(), quote.Hello(),
	// Source for the following sentences: https://tour.golang.org/
	"Go provides concurrency features as part of the core language.",
	"A function can take zero or more arguments.",
	"A function can return any number of results.",
	"A struct is a collection of fields.",
	"Struct fields are accessed using a dot.",
	"Go's return values may be named.",
	"A var statement can be at package or function level.",
	"A map maps keys to values.",
}

type Result struct {
	totalTime time.Duration
	distance  int
	score     int
}

// Pluralizes the string if required.
func pluralize(str string, n int) string {
	if n != 1 {
		return str + "s"
	} else {
		return str
	}
}

// Prints the result, including time taken to type the text, distance and score.
func (result Result) Print(text string) {
	fmt.Println("Finished in", result.totalTime.String()+"!")

	if result.distance != 0 {
		fmt.Println("Off by", result.distance, pluralize("character", result.distance))

		if result.score == 0 {
			fmt.Println("No score")
		} else {
			fmt.Println("Score:", result.score)
		}
	} else {
		fmt.Println("Perfect Score -", result.score)
	}

	previousScore := scores[text]
	if result.score > previousScore {
		fmt.Println("NEW HIGHSCORE!")
		scores[text] = result.score
	}
}

type Scores map[string]int

// Holds scores corresponding to the respective text.
// This is saved locally and loaded on start.
var scores Scores

// Saves the scores to a local file.
func (scores Scores) Save() (err error) {
	scoresJson, err := json.Marshal(scores)

	if err != nil {
		return
	}

	perm := os.FileMode(0644) // Read write permissions
	err = ioutil.WriteFile("scores.json", scoresJson, perm)

	return
}

// Loads the scores from a local file.
func (scores Scores) Load() (err error) {
	scoresJson, err := ioutil.ReadFile("scores.json")

	if err != nil {
		return nil
	}

	json.Unmarshal(scoresJson, &scores)

	return
}

// Determines whether the first text has been typed.
var firstRun = true

func handleCtrlC() {
	if scores.Save() != nil {
		fmt.Println("Failed to save scores")
	}

	fmt.Println("See you later!")
	os.Exit(0)
}

func main() {
	scores = make(Scores)

	err := scores.Load()

	if err != nil {
		fmt.Println("Failed to load scores")
		os.Exit(0)
	}

	// Exit gracefully on Ctrl+C
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			handleCtrlC()
		}
	}()

	for {
		fmt.Println("Type the following text as quickly as you can!")
		countdown()

		result, text := play()

		_, exists := scores[text]
		if !exists {
			scores[text] = result.score
		}

		result.Print(text)

		if firstRun {
			fmt.Println("\nKeep playing to see text-specific scores and records!")
		}

		fmt.Println("\nPress Enter to type another text or Ctrl+C to abort")
		readLine()

		firstRun = false
	}
}

var lastRandInt int
var randIntSrc = rand.NewSource(time.Now().UnixNano())
var rng = rand.New(randIntSrc)

// Gets a random integer guaranteed to be different from the previously generated integer.
// This function has an undefined time complexity and may never terminate.
func getNewRandInt(n int) int {
	randInt := rng.Intn(n)
	for randInt == lastRandInt {
		randInt = rng.Intn(n)
	}
	lastRandInt = randInt
	return randInt
}

// Counts down.
func countdown() {
	defer fmt.Println()

	countdownStart := 3
	for countdownStart > 0 {
		fmt.Println(countdownStart, "...")
		if firstRun {
			time.Sleep(time.Millisecond * 1000)
		} else {
			time.Sleep(time.Millisecond * 750)
		}
		countdownStart--
	}
	fmt.Println("Go!")
}

var reader = bufio.NewReader(os.Stdin)

// Reads in a line from the terminal.
func readLine() string {
	string, err := reader.ReadString('\n')

	if err != nil {
		// Although there could be other causes, we will just assume here that the user pressed Ctrl+C
		handleCtrlC()
	}

	return string
}

const prefix = "> "

// Calculates the score from the distance.
func getScore(distance int) int {
	score := 10 - distance
	if score < 0 {
		return 0
	} else {
		return score * 100
	}
}

// Plays a game round.
func play() (Result, string) {
	textToType := text[getNewRandInt(len(text))]
	fmt.Print(
		prefix,
		textToType,
		"\r", // move the cursor to the start
		prefix,
	)

	startTime := time.Now()
	input := readLine()
	endTime := time.Now()

	fmt.Println()

	distance := levenshtein.ComputeDistance(strings.TrimSpace(input), textToType)
	totalTime := endTime.Sub(startTime)

	score := getScore(distance)

	result := Result{totalTime, distance, score}

	return result, textToType
}
