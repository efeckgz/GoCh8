package ch8sdl

import "github.com/veandco/go-sdl2/mix"

type sound struct {
	chunk *mix.Chunk
}

func newSound(chunk *mix.Chunk) sound {
	return sound{chunk: chunk}
}

func (s sound) Play() {
	s.chunk.Play(-1, 0)
}

func (s sound) Pause() {
	mix.HaltChannel(-1)
}
