package ch8sdl

type colorScheme byte

const (
	_ colorScheme = iota
	// Black and white color scheme
	Black

	// Yellow color scheme
	Yellow

	// Green color scheme
	Green
)

var colorSchemes = map[string]colorScheme{
	"black":  Black,
	"yellow": Yellow,
	"green":  Green,
}

// ParseColorScheme takes the string passed by the user as a cli arguments
// and returns the appropriate color. Default is Black.
func ParseColorScheme(arg *string) colorScheme {
	color, ok := colorSchemes[*arg]
	if !ok {
		return Black
	}
	return color
}
