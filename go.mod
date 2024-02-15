module github.com/efeckgz/GoCh8

go 1.22

require (
	github.com/efeckgz/GoCh8/ch8 v0.0.0-00010101000000-000000000000
	github.com/veandco/go-sdl2 v0.4.38
)

replace Chip8 => github.com/efeckgz/GoCh8 v0.0.0

replace github.com/efeckgz/GoCh8/ch8 => ./ch8
