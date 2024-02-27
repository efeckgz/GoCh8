package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/efeckgz/GoCh8/ch8"
	"github.com/efeckgz/GoCh8/ch8sdl"
)

func main() {
	romPathArg := flag.String("rom", "", "Path to the chip 8 program")
	colorArg := flag.String("color", "green", "The color scheme for Chip 8")
	specArg := flag.String("spec", "original", "The specification of Chip 8 to emulate.")
	speedArg := flag.Int("speed", 1, "The speed of emulation")

	colorArg = trimAndLower(colorArg)
	specArg = trimAndLower(specArg)
	flag.Parse()

	checkArgumentAndAsk("Rom path", romPathArg)
	color := ch8sdl.ParseColorScheme(colorArg)
	spec := ch8.ParseChip8Spec(specArg)

	ch8sdl.RunSDL(spec, *romPathArg, color, *speedArg)
}

// trimAndLower is a function that removes whitespace from a string and converts it to lowercase.
// Arguments are converted before pattern matching so that the user is not restricted to only one way of
// providing the same argument.
func trimAndLower(s *string) *string {
	*s = strings.Join(strings.Fields(*s), "")
	*s = strings.ToLower(*s)
	return s
}

// checkArgumentAndAsk is a function that checks if a necessary argument is empty and asks the user to provide it.
func checkArgumentAndAsk(argumentName string, argument *string) {
	if *argument == "" {
		fmt.Printf("Please provide the following argument: %s ", argumentName)
		_, err := fmt.Scanln(argument)
		if err != nil {
			log.Fatalf("Scan error: %v", err)
		}
	}
}
