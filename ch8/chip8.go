package ch8

import (
	"io"
	"log"
	"math/rand"
	"os"
	"path/filepath"
)

const (
	instructionsPerFrame = 10
	fps                  = 60

	// FrameDelay represents the time between two frames. It is used to time a 60hz loop.
	FrameDelay = 1000 / fps
)

// RenderingMode represents the different rendering modes of the super chip and xo-chip variants.
type RenderingMode byte

const (
	_ RenderingMode = iota

	// HiresRendering represents the hires rendering mode.
	HiresRendering

	// LoresRendering represents the lores rendering mode.
	LoresRendering
)

// CPU represents the inner state of the Chip 8.
type CPU struct {
	Spec           Spec
	registers      [16]byte
	programCounter uint16
	memory         [4096]byte
	stack          [16]uint16
	stackPointer   uint
	indexRegister  uint16
	SoundTimer     byte
	DelayTimer     byte

	// DisplayBuffer is a 2D array of booleans representing all the pixels in the Chip8 display.
	// The size of the buffer is set to the hires mode of the super and xo-chip variants. Only use
	// the 64x32 part when working with lores mode or original spec.
	// A true value represents an on pixel.
	DisplayBuffer [64][128]bool

	// RenderingMode represents the rendering mode of the interpreter. In lores mode only the top left
	// part of the buffer is accessible. For original spec, only lores mode is available.
	RenderingMode RenderingMode

	// DisplayUpdated a flag that is raised when the state of the DisplayBuffer is changed.
	// Check this flag to redraw the screen only when necessary.
	DisplayUpdated bool

	// Keypad is an array of 16 booleans representing the state of the 16 keys present in chip-8's keypad.
	// ith index of this array represents the key of original chip-8 with the value i.
	// For example, if Keypad[0xA] is true, the A button is pressed.
	Keypad [16]bool

	// beep is the sound played from the Chip8.
	beep Beep
}

// NewCPU creates a new Chip8 with default values.
func NewCPU(spec Spec, beep Beep) (ch8 CPU) {
	fontSet := [80]byte{
		0xF0, 0x90, 0x90, 0x90, 0xF0, // 0
		0x20, 0x60, 0x20, 0x20, 0x70, // 1
		0xF0, 0x10, 0xF0, 0x80, 0xF0, // 2
		0xF0, 0x10, 0xF0, 0x10, 0xF0, // 3
		0x90, 0x90, 0xF0, 0x10, 0x10, // 4
		0xF0, 0x80, 0xF0, 0x10, 0xF0, // 5
		0xF0, 0x80, 0xF0, 0x90, 0xF0, // 6
		0xF0, 0x10, 0x20, 0x40, 0x40, // 7
		0xF0, 0x90, 0xF0, 0x90, 0xF0, // 8
		0xF0, 0x90, 0xF0, 0x10, 0xF0, // 9
		0xF0, 0x90, 0xF0, 0x90, 0x90, // A
		0xE0, 0x90, 0xE0, 0x90, 0xE0, // B
		0xF0, 0x80, 0x80, 0x80, 0xF0, // C
		0xE0, 0x90, 0x90, 0x90, 0xE0, // D
		0xF0, 0x80, 0xF0, 0x80, 0xF0, // E
		0xF0, 0x80, 0xF0, 0x80, 0x80, // F
	}

	ch8 = CPU{
		Spec:           spec,
		programCounter: 0x200, // chip 8 programs are loaded from 512 bytes in.
		beep:           beep,
		RenderingMode:  LoresRendering,
	}

	for i := 0; i < len(fontSet); i++ {
		ch8.memory[0x000+i] = fontSet[i] // addresses 0x000 to 0x080 reserved for fonts
	}

	//ch8.memory[0x1FF] = 2 // value of 1-3 here will force the quirks test rom to bypass the menu screen

	return
}

// LoadProgram reads a file from the provided path and loads its contents to the Chip8 memory.
func (ch8 *CPU) LoadProgram(programPath string) error {
	safePath := filepath.Clean(programPath)
	file, err := os.Open(safePath)
	if err != nil {
		return err
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatalf("Could not close the rom file: %v", err)
		}
	}(file)

	buffer, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	copy(ch8.memory[0x200:], buffer) // Load the program from 512 bytes in.
	return nil
}

// ClearProgram clears the loaded program.
func (ch8 *CPU) ClearProgram() {
	for i := 0x200; i < len(ch8.memory); i++ {
		ch8.memory[i] = 0x0
	}
}

// Tick emulates what the chip 8 does in 1/60 of a second.
func (ch8 *CPU) Tick(speed int) {
	if ch8.DelayTimer > 0 {
		ch8.DelayTimer--
	}

	if ch8.SoundTimer > 0 {
		ch8.beep.Play()
		ch8.SoundTimer--
	} else {
		ch8.beep.Pause()
	}

	ch8.emulateCycle(instructionsPerFrame, speed)
}

func (ch8 *CPU) emulateCycle(cycles, speed int) {
	runFor := cycles * speed
	for i := 0; i < runFor; i++ {
		opcode := ch8.readOpcode()
		ch8.programCounter += 2

		var (
			c = byte((opcode & 0xF000) >> 12)
			x = byte((opcode & 0x0F00) >> 8)
			y = byte((opcode & 0x00F0) >> 4)
			d = byte(opcode & 0x000F)

			nnn = opcode & 0x0FFF
			nn  = byte(opcode & 0x00FF)
			n   = byte(opcode & 0x000F)
		)

		switch c {
		case 0x0:
			switch y {
			case 0xC:
				ch8.scrollDownN(n)
			case 0xE:
				switch d {
				case 0x0:
					ch8.clearScreen()
				case 0xE:
					ch8.returnFromSubroutine()
				}
			case 0xF:
				switch d {
				case 0xB:
					ch8.scrollRightFour()
				case 0xC:
					ch8.scrollLeftFour()
				case 0xD:
					// exit interpreter
					return
				case 0xE:
					ch8.switchToLores()
				case 0xF:
					ch8.switchToHires()
				}
			}
		case 0x1:
			ch8.jump(nnn)
		case 0x2:
			ch8.call(nnn)
		case 0x3:
			ch8.skipIfEqualVxNn(x, nn)
		case 0x4:
			ch8.skipIfNotEqualVxNn(x, nn)
		case 0x5:
			// maybe check if d == 0?
			ch8.skipIfEqualVxVy(x, y)
		case 0x6:
			ch8.loadXNN(x, nn)
		case 0x7:
			ch8.addNnVx(x, nn)
		case 0x8:
			switch d {
			case 0x0:
				ch8.setVxVy(x, y)
			case 0x1:
				ch8.orVxVy(x, y)
			case 0x2:
				ch8.andVxVy(x, y)
			case 0x3:
				ch8.xorVxVy(x, y)
			case 0x4:
				ch8.addVxVy(x, y)
			case 0x5:
				ch8.subVxVy(x, y)
			case 0x6:
				ch8.rightShiftVx(x, y)
			case 0x7:
				ch8.subVyVx(y, x)
			case 0xE:
				ch8.leftShiftVx(x, y)
			}
		case 0x9:
			ch8.skipIfNotEqualVxVy(x, y)
		case 0xA:
			ch8.loadIndexRegisterNNN(nnn)
		case 0xB:
			ch8.jumpWithOffset(nnn)
		case 0xC:
			ch8.randomAndNn(x, nn)
		case 0xD:
			ch8.draw(x, y, n)
		case 0xE:
			switch nn {
			case 0x9E:
				ch8.skipIfVxPressed(x)
			case 0xA1:
				ch8.skipIfVxNotPressed(x)
			}
		case 0xF:
			switch nn {
			case 0x07:
				ch8.setVxDelayTimer(x)
			case 0x0A:
				ch8.delayUntilKey(x)
			case 0x15:
				ch8.setDelayTimerVx(x)
			case 0x18:
				ch8.setSoundTimerVx(x)
			case 0x1E:
				ch8.addIndexVx(x)
			case 0x29:
				ch8.setIVx(x)
			case 0x33:
				ch8.vxToBCD(x)
			case 0x55:
				ch8.writeVxVi(x)
			case 0x65:
				ch8.writeViVx(x)
			}
		default:
			log.Fatalf("Unimplemented opcode: %#x", opcode)
		}
	}
}

func (ch8 *CPU) readOpcode() (opcode uint16) {
	firstByte := uint16(ch8.memory[ch8.programCounter])
	secondByte := uint16(ch8.memory[ch8.programCounter+1])

	opcode = (firstByte << 8) | secondByte
	return
}

func (ch8 *CPU) clearScreen() {
	ch8.DisplayUpdated = true
	ch8.DisplayBuffer = [64][128]bool{}
}

func (ch8 *CPU) returnFromSubroutine() {
	if ch8.stackPointer == 0 {
		log.Fatalln("Stack underflow!")
		return
	}

	ch8.stackPointer--
	ch8.programCounter = ch8.stack[ch8.stackPointer]
}

func (ch8 *CPU) jump(nnn uint16) {
	ch8.programCounter = nnn
}

func (ch8 *CPU) call(nnn uint16) {
	ch8.stack[ch8.stackPointer] = ch8.programCounter
	ch8.stackPointer++
	ch8.programCounter = nnn
}

func (ch8 *CPU) skipIfEqualVxNn(x, nn byte) {
	xVal := ch8.registers[uint(x)]
	if xVal == nn {
		ch8.programCounter += 2
	}
}

func (ch8 *CPU) skipIfNotEqualVxNn(x, nn byte) {
	xVal := ch8.registers[uint(x)]
	if xVal != nn {
		ch8.programCounter += 2
	}
}

func (ch8 *CPU) addVxVy(x, y byte) {
	xVal, yVal := ch8.registers[uint(x)], ch8.registers[uint(y)]
	sum := xVal + yVal

	ch8.registers[uint(x)] = sum

	if sum < xVal || sum < yVal {
		// overflow happened
		ch8.registers[0xF] = 1
	} else {
		ch8.registers[0xF] = 0
	}
}

func (ch8 *CPU) loadXNN(x, nn byte) {
	ch8.registers[uint(x)] = nn
}

func (ch8 *CPU) loadIndexRegisterNNN(nnn uint16) {
	ch8.indexRegister = nnn
}

func (ch8 *CPU) draw(x, y, n byte) {
	if ch8.Spec == Original {
		// TODO Display wait quirk
	}

	var xLimit, yLimit byte
	switch ch8.RenderingMode {
	case HiresRendering:
		xLimit, yLimit = 128, 64
	case LoresRendering:
		xLimit, yLimit = 64, 32
	}

	xCoordinate := ch8.registers[uint(x)] % xLimit
	yCoordinate := ch8.registers[uint(y)] % yLimit
	ch8.registers[0xF] = 0x0

	for i := byte(0); i < n; i++ {
		currentSpriteByte := ch8.memory[uint(ch8.indexRegister+uint16(i))]
		for j := uint16(7); j <= 7; j-- { // start from 7
			currentSpriteBit := (currentSpriteByte >> j) & 1
			if currentSpriteBit == 1 && ch8.DisplayBuffer[uint(yCoordinate)][uint(xCoordinate)] {
				ch8.DisplayBuffer[uint(yCoordinate)][uint(xCoordinate)] = false
				ch8.registers[0xF] = 0x1
				ch8.DisplayUpdated = true
			} else if currentSpriteBit == 1 {
				ch8.DisplayBuffer[uint(yCoordinate)][uint(xCoordinate)] = true
				ch8.DisplayUpdated = true
			}

			if ch8.Spec == Original || ch8.Spec == Super {
				xCoordinate++
				if xCoordinate >= xLimit {
					break
				}
			} else {
				xCoordinate = (xCoordinate + 1) % xLimit
			}
		}

		xCoordinate = ch8.registers[uint(x)] % xLimit // reset x coordinate for the next row of sprites

		if ch8.Spec == Original || ch8.Spec == Super {
			yCoordinate++
			if yCoordinate >= yLimit {
				break
			}
		} else {
			yCoordinate = (yCoordinate + 1) % yLimit
		}
	}
}

func (ch8 *CPU) skipIfEqualVxVy(x, y byte) {
	if ch8.registers[uint(x)] == ch8.registers[uint(y)] {
		ch8.programCounter += 2
	}
}

func (ch8 *CPU) addNnVx(x, nn byte) {
	ch8.registers[uint(x)] += nn
}

func (ch8 *CPU) setVxVy(x, y byte) {
	ch8.registers[uint(x)] = ch8.registers[uint(y)]
}

func (ch8 *CPU) orVxVy(x, y byte) {
	ch8.registers[uint(x)] |= ch8.registers[uint(y)]
	if ch8.Spec == Original {
		ch8.registers[0xF] = 0x0
	}
}

func (ch8 *CPU) andVxVy(x, y byte) {
	ch8.registers[uint(x)] &= ch8.registers[uint(y)]
	if ch8.Spec == Original {
		ch8.registers[0xF] = 0x0
	}
}

func (ch8 *CPU) subVxVy(x, y byte) {
	xVal, yVal := ch8.registers[uint(x)], ch8.registers[uint(y)]

	result := xVal - yVal
	ch8.registers[uint(x)] = result

	// check for overflow
	if xVal < yVal {
		ch8.registers[0xF] = 0x0
	} else {
		ch8.registers[0xF] = 0x1
	}
}

func (ch8 *CPU) rightShiftVx(x, y byte) {
	if ch8.Spec == Original {
		ch8.registers[uint(x)] = ch8.registers[uint(y)]
	}

	shiftedOut := ch8.registers[uint(x)] & 1 // the least significant bit is going to be shifted out
	ch8.registers[uint(x)] >>= 1
	ch8.registers[0xF] = shiftedOut
}

func (ch8 *CPU) subVyVx(y, x byte) {
	yVal, xVal := ch8.registers[uint(y)], ch8.registers[uint(x)]

	result := yVal - xVal
	ch8.registers[uint(x)] = result

	if yVal < xVal {
		ch8.registers[0xF] = 0x0
	} else {
		ch8.registers[0xF] = 0x1
	}
}

func (ch8 *CPU) leftShiftVx(x, y byte) {
	if ch8.Spec == Original {
		ch8.registers[uint(x)] = ch8.registers[uint(y)]
	}

	shiftedOut := (ch8.registers[uint(x)] >> 7) & 1 // The most significant bit is going to be shifted out
	ch8.registers[uint(x)] <<= 1
	ch8.registers[0xF] = shiftedOut
}

func (ch8 *CPU) skipIfNotEqualVxVy(x, y byte) {
	if ch8.registers[uint(x)] != ch8.registers[uint(y)] {
		ch8.programCounter += 2
	}
}

func (ch8 *CPU) jumpWithOffset(nnn uint16) {
	var offset uint16
	if ch8.Spec == Original {
		offset = uint16(ch8.registers[0x0])
	} else if ch8.Spec == Super {
		register := nnn & 0xF00 >> 8
		offset = uint16(ch8.registers[register])
	}
	ch8.programCounter = nnn + offset // take another look
}

func (ch8 *CPU) randomAndNn(x, nn byte) {
	random := byte(rand.Intn(255))
	ch8.registers[uint(x)] = random & nn
}

func (ch8 *CPU) skipIfVxPressed(x byte) {
	value := uint16(ch8.registers[uint(x)&0xF]) // only keep the lowest 4 bits
	if ch8.Keypad[value] {
		ch8.programCounter += 2
	}
}

func (ch8 *CPU) skipIfVxNotPressed(x byte) {
	value := uint16(ch8.registers[uint(x)&0xF]) // only keep the lowest 4 bits
	if !ch8.Keypad[value] {
		ch8.programCounter += 2
	}
}

func (ch8 *CPU) setVxDelayTimer(x byte) {
	ch8.registers[uint(x)] = ch8.DelayTimer
}

func (ch8 *CPU) setDelayTimerVx(x byte) {
	ch8.DelayTimer = ch8.registers[uint(x)]
}

func (ch8 *CPU) setSoundTimerVx(x byte) {
	ch8.SoundTimer = ch8.registers[uint(x)]
}

func (ch8 *CPU) addIndexVx(x byte) {
	ch8.indexRegister += uint16(ch8.registers[uint(x)])
}

func (ch8 *CPU) writeVxVi(x byte) {
	for i := uint16(0); i <= uint16(x); i++ {
		ch8.memory[(uint(ch8.indexRegister + i))] = ch8.registers[uint(i)]
	}

	if ch8.Spec == Original {
		ch8.indexRegister += uint16(x + 1)
	}
}

func (ch8 *CPU) writeViVx(x byte) {
	for i := uint16(0); i <= uint16(x); i++ {
		ch8.registers[uint(i)] = ch8.memory[uint(ch8.indexRegister+i)]
	}

	if ch8.Spec == Original {
		ch8.indexRegister += uint16(x + 1)
	}
}

func (ch8 *CPU) xorVxVy(x, y byte) {
	ch8.registers[uint(x)] ^= ch8.registers[uint(y)]
	if ch8.Spec == Original {
		ch8.registers[0xF] = 0x0
	}
}

func (ch8 *CPU) setIVx(x byte) {
	character := ch8.registers[uint(x)&0xF] // only keep the lowest 4 bits
	ch8.indexRegister = uint16(character) * 5
}

func (ch8 *CPU) vxToBCD(x byte) {
	value := ch8.registers[uint(x)]
	for i := uint16(2); i < 3; i-- {
		digit := value % 10
		ch8.memory[uint(ch8.indexRegister+i)] = digit
		value /= 10
	}
}

func (ch8 *CPU) delayUntilKey(x byte) {
	keyIsPressed := false
	for keyIndex, key := range ch8.Keypad {
		// The original variant did not continue execution until the key was released
		if key {
			ch8.memory[uint(x)] = byte(keyIndex)
			keyIsPressed = true
			break
		}
	}

	// Decrement the program counter if no key is pressed, essentially pausing execution.
	if !keyIsPressed {
		ch8.programCounter -= 2
	}
}

func (ch8 *CPU) switchToLores() {
	ch8.RenderingMode = LoresRendering
	ch8.DisplayUpdated = true
}

func (ch8 *CPU) switchToHires() {
	ch8.RenderingMode = HiresRendering
	ch8.DisplayUpdated = true
}

func (ch8 *CPU) scrollDownN(n byte) {
	for y := len(ch8.DisplayBuffer) - 1; y >= 0; y-- {
		for x, pixel := range ch8.DisplayBuffer[y] {
			if pixel {
				if y+int(n) < len(ch8.DisplayBuffer) {
					ch8.DisplayBuffer[y][x] = false
					ch8.DisplayBuffer[y+int(n)][x] = true
				} else {
					// Turn of the pixel that goes out of the display
					ch8.DisplayBuffer[y][x] = false
				}
				ch8.DisplayUpdated = true
			}
		}
	}
}

func (ch8 *CPU) scrollRightFour() {
	for y, row := range ch8.DisplayBuffer {
		for x := len(row) - 1; x >= 0; x-- {
			if row[x] {
				if x+4 < len(ch8.DisplayBuffer[y]) {
					ch8.DisplayBuffer[y][x] = false
					ch8.DisplayBuffer[y][x+4] = true
				} else {
					// Move pixel out of screen
					ch8.DisplayBuffer[y][x] = true
				}
				ch8.DisplayUpdated = true
			}
		}
	}
}

func (ch8 *CPU) scrollLeftFour() {
	for y, row := range ch8.DisplayBuffer {
		for x := 0; x < len(row); x++ {
			if row[x] {
				if x-4 < 0 {
					// Pixel moves out of the screen
					ch8.DisplayBuffer[y][x] = false
				} else {
					ch8.DisplayBuffer[y][x] = false
					ch8.DisplayBuffer[y][x-4] = true
				}
			}
		}
	}
}
