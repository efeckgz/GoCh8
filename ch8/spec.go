package ch8

import (
	"fmt"
)

// Spec type specifies the specification of chip8 to emulate
type Spec int

const (
	// Original represents the Cosmac-Vip spec of Chip 8.
	Original Spec = iota

	// Super represents the Super-chip8 spec of Chip 8.
	Super

	// Xo represents the XO-Chip spec of Chip 8.
	Xo
)

// Specs maps the names of each spec to its corresponding Spec value.
var Specs = map[string]Spec{
	"original": Original,
	"super":    Super,
	"xo":       Xo,
}

// ParseChip8Spec is a function that parses the string passed as a cli argument by the user to one of the
// as a Spec for use in emulator.
func ParseChip8Spec(specArgument *string) Spec {
	spec, ok := Specs[*specArgument]
	if !ok {
		fmt.Println("No spec specified, derfaulting to Original.")
		return Original
	}
	return spec
}
