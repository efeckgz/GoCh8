package ch8sdl

import (
	"github.com/efeckgz/GoCh8/ch8"
	// embed used for embedding the beep file into the binary.
	_ "embed"
	"fmt"
	"log"
	"time"

	"github.com/veandco/go-sdl2/mix"
	"github.com/veandco/go-sdl2/sdl"
)

const (
	windowWidth  = 640
	windowHeight = 320

	colorAlpha = 255
)

var (
	bgR byte = 0
	bgG byte = 0
	bgB byte = 0
	fgR byte = 255
	fgG byte = 255
	fgB byte = 255
)

//go:embed assets/beep.wav
var beepBytes []byte

// RunSDL runs the emulator using SDL.
func RunSDL(spec ch8.Spec, romPath string, color ColorScheme, speed int) {
	if color == Yellow {
		bgR = 154
		bgG = 102
		bgB = 1
		fgR = 255
		fgG = 204
		fgB = 1
	}

	if color == Green {
		fgR = 0
		fgG = 255
		fgB = 0
	}

	// renderer is used to draw the cpu's display buffer. window is only used for cleaning up.
	renderer, window, beep := setup(spec)
	defer cleanup(window, renderer, beep)

	sound := newSound(beep) // convert the *mix.Chunk to a Beep interface
	cpu := ch8.NewCPU(spec, sound)
	err := cpu.LoadProgram(romPath)
	if err != nil {
		log.Fatalf("Error loading program: %v\n", err)
	}

	fmt.Println(`controls: 
    Keyboard				CHIP-8
	|1| |2| |3| |4|			|1| |2|	|3| |C|		
	|Q| |W| |E| |R|			|4| |5| |6| |D|
	|A| |S| |D| |F|			|7| |8| |9| |E|
	|Z| |X| |C| |V|			|A| |0| |B| |F|`)

	running := true
	for running {
		frameStart := time.Now()

		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				running = false
			case *sdl.KeyboardEvent:
				handleKeyboardInput(e, &cpu)
			}
		}

		cpu.Tick(speed)
		if cpu.DisplayUpdated {
			drawFromBuffer(cpu.DisplayBuffer, cpu.RenderingMode, renderer, bgR, bgG, bgB, fgR, fgG, fgB)
			cpu.DisplayUpdated = false
		}

		frameTime := time.Since(frameStart)
		if ch8.FrameDelay > frameTime.Milliseconds() {
			sdl.Delay(uint32(ch8.FrameDelay - frameTime.Milliseconds())) // better than time.Sleep()
		}
	}
}

// setup is a function that sets up a SDL window, renderer and the beeper for use in chip8.
func setup(spec ch8.Spec) (*sdl.Renderer, *sdl.Window, *mix.Chunk) {
	beepRWops, err := sdl.RWFromMem(beepBytes)
	if err != nil {
		log.Fatalln("Could not read the beep binary: ", err)
	}

	err = sdl.Init(sdl.INIT_VIDEO | sdl.INIT_AUDIO | sdl.INIT_EVENTS)
	if err != nil {
		log.Fatalf("Failed to initialize sdl: %v", err)
	}

	specOnWindow := ""
	switch spec {
	case ch8.Original:
		specOnWindow = "Original"
	case ch8.Super:
		specOnWindow = "Super-chip 1.1"
	case ch8.Xo:
		specOnWindow = "XO-Chip"
	}

	window, err := sdl.CreateWindow(
		fmt.Sprintf("Chip-8 Interpreter (%s)", specOnWindow),
		sdl.WINDOWPOS_UNDEFINED,
		sdl.WINDOWPOS_UNDEFINED,
		windowWidth, windowHeight,
		sdl.WINDOW_SHOWN)
	if err != nil {
		log.Fatalf("Failed to create the window: %v", err)
	}

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		log.Fatalf("Could not create renderer for window: %v", err)
	}

	if err := mix.Init(mix.INIT_MP3 | mix.INIT_FLAC | mix.INIT_OGG); err != nil {
		log.Fatalf("Failed to initialize SDL Mixer: %v", err)
	}

	if err := mix.OpenAudio(mix.DEFAULT_FREQUENCY, mix.DEFAULT_FORMAT, mix.DEFAULT_CHANNELS, 1024); err != nil {
		log.Fatalf("Failed to open audio: %v", err)
	}

	beep, err := mix.LoadWAVRW(beepRWops, false)
	if err != nil {
		log.Fatalln("Could not load wavrw: ", err)
	}

	return renderer, window, beep
}

// cleanup is a function that is responsible for cleaning up after setup.
func cleanup(window *sdl.Window, renderer *sdl.Renderer, beep *mix.Chunk) {
	err := renderer.Destroy()
	if err != nil {
		log.Fatalln("The renderer could not be destroyed.")
	}
	err = window.Destroy()
	if err != nil {
		log.Fatalln("The window could not be destroyed.")
	}
	beep.Free()
	mix.CloseAudio()
	sdl.Quit()
}

// handleKeyboardInput is a function that maps the SDL key events to chip8 keypad and sets the keypad key states
// accordingly when the appropriate key is pressed.
func handleKeyboardInput(key *sdl.KeyboardEvent, cpu *ch8.CPU) {
	switchKeyState := func(keypadIndex uint) {
		if key.State == sdl.PRESSED {
			cpu.Keypad[keypadIndex] = true
		} else if key.State == sdl.RELEASED {
			cpu.Keypad[keypadIndex] = false
		}
	}

	switch key.Keysym.Sym {
	case sdl.K_1:
		switchKeyState(0x1)
	case sdl.K_2:
		switchKeyState(0x2)
	case sdl.K_3:
		switchKeyState(0x3)
	case sdl.K_4:
		switchKeyState(0xC)
	case sdl.K_q:
		switchKeyState(0x4)
	case sdl.K_w:
		switchKeyState(0x5)
	case sdl.K_e:
		switchKeyState(0x6)
	case sdl.K_r:
		switchKeyState(0xD)
	case sdl.K_a:
		switchKeyState(0x7)
	case sdl.K_s:
		switchKeyState(0x8)
	case sdl.K_d:
		switchKeyState(0x9)
	case sdl.K_f:
		switchKeyState(0xE)
	case sdl.K_z:
		switchKeyState(0xA)
	case sdl.K_x:
		switchKeyState(0x0)
	case sdl.K_c:
		switchKeyState(0xB)
	case sdl.K_v:
		switchKeyState(0xF)
	}
}

// drawFromBuffer is a function that draws the contents of the chip8's display buffer to the SDL window.
func drawFromBuffer(displayBuffer [64][128]bool, renderingMode ch8.RenderingMode, renderer *sdl.Renderer, bgR, bgG, bgB, fgR, fgG, fgB byte) {
	var xLimit, yLimit, pixelSize int
	switch renderingMode {
	case ch8.LoresRendering:
		xLimit, yLimit = 64, 32
		pixelSize = 10
	case ch8.HiresRendering:
		xLimit, yLimit = 128, 64
		pixelSize = 5
	}

	err := renderer.SetDrawColor(bgR, bgG, bgB, colorAlpha)
	if err != nil {
		log.Fatalln("Could not set the drawing color to black.")
	}

	err = renderer.Clear()
	if err != nil {
		log.Fatalln("Could not clear the rendering area with the drawing color.")
	}

	for i := 0; i < yLimit; i++ {
		for j := 0; j < xLimit; j++ {
			pixel := displayBuffer[i][j]
			if pixel {
				err := renderer.SetDrawColor(fgR, fgG, fgB, colorAlpha)
				if err != nil {
					log.Fatalln("Could not set the drawing color to white.")
				}
				err = renderer.FillRect(&sdl.Rect{X: int32(j * pixelSize), Y: int32(i * pixelSize), W: int32(pixelSize), H: int32(pixelSize)})
				if err != nil {
					log.Fatalln("Could not draw the current sprite.")
				}
			}
		}
	}

	renderer.Present()
}
