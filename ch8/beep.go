package ch8

// Beep represents the beep sound of the chip-8.
type Beep interface {
	Play()
	Pause()
}
