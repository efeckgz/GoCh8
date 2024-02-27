package ch8sdl

// ColorScheme represents the different color schemes that can be used in the emulator.
type ColorScheme byte

const (
	_ ColorScheme = iota
	// Black and white color scheme
	Black

	// Yellow color scheme
	Yellow

	// Green color scheme
	Green
)

var colorSchemes = map[string]ColorScheme{
	"black":  Black,
	"yellow": Yellow,
	"green":  Green,
}

// ParseColorScheme takes the string passed by the user as a cli arguments
// and returns the appropriate color. Default is Black.
func ParseColorScheme(arg *string) ColorScheme {
	color, ok := colorSchemes[*arg]
	if !ok {
		return Green
	}
	return color
}
