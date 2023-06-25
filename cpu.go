package main

import (
	"io/ioutil"
	"os"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
)

type CPU struct {
	memory    [4096]uint8
	pc        uint16
	v         [16]uint8
	i         uint16
	sp        uint8
	dt        uint8
	st        uint8
	stack     [16]uint16
	keyboard  Keyboard
	superchip bool
}

type Sprite struct {
	x uint8
	y uint8
}

type Keyboard struct {
	keys [16]uint8
}

func (cpu *CPU) loadProgram(program []uint8) {
	for i := 0; i < len(program); i++ {
		cpu.memory[i+512] = program[i]
	}
}

const PIXEL_SIZE = 20.0

func cpuLoop(cpu *CPU, window *pixelgl.Window) {

	// Fallback to prevent index out of bounds
	if (cpu.pc + 1) >= 4096 {
		return
	}

	// 1. Fetch
	opcode := fetch(cpu)

	// 2. Decode and execute
	decode(cpu, opcode, window)

	// 3. Update timers
	if cpu.dt > 0 {
		cpu.dt--
	}

	if cpu.st > 0 {
		cpu.st--
	}
}

func getProgram() []uint8 {
	path := os.Args[1]
	dat, err := ioutil.ReadFile(path)

	if err != nil {
		panic(err)
	}

	program := make([]uint8, len(dat))

	for i := 0; i < len(dat); i++ {
		program[i] = dat[i]
	}

	return program
}

func initCpu() {
	pixelgl.Run(runWindow)
}

func fetch(cpu *CPU) uint16 {
	opcode := uint16(cpu.memory[cpu.pc])<<8 | uint16(cpu.memory[cpu.pc+1])
	cpu.pc += 2
	return opcode
}

func decode(cpu *CPU, opcode uint16, window *pixelgl.Window) {
	switch opcode & 0xF000 {
	case 0x0000:
		if opcode == 0x00EE {
			returnFromSubroutine(cpu)
		}
	case 0x1000:
		jumpToAddress(cpu, opcode&0x0FFF)
	case 0x2000:
		callSubroutine(cpu, opcode&0x0FFF)
	case 0x3000:
		skipIfEqual(cpu, opcode)
	case 0x4000:
		// TODO: skipIfNotEqual(cpu, opcode)
	case 0x5000:
		// TODO: skipIfEqualRegisters(cpu, opcode)
	case 0x6000:
		setRegister(cpu, opcode)
	case 0x7000:
		addToRegister(cpu, opcode)
	case 0x8000:
		switch opcode & 0x000F {
		case 0x0000:
			setRegisterToRegister(cpu, opcode)
		case 0x0001:
			setRegisterToRegisterOr(cpu, opcode)
		case 0x0002:
			setRegisterToRegisterAnd(cpu, opcode)
		case 0x0003:
			setRegisterToRegisterXor(cpu, opcode)
		case 0x0004:
			addRegisterToRegister(cpu, opcode)
		case 0x0005:
			subtractRegisterFromRegister(cpu, opcode, false)
		case 0x0006:
			shiftRegisterRight(cpu, opcode)
		case 0x0007:
			// swap registers
			subtractRegisterFromRegister(cpu, opcode, true)
		case 0x000E:
			shiftRegisterLeft(cpu, opcode)
		}
	case 0x9000:
		// TODO: skipIfNotEqualRegisters(cpu, opcode)
	case 0xA000:
		setIndexRegister(cpu, opcode)
	case 0xB000:
		// TODO: jumpToAddressPlusRegister(cpu, opcode)
	case 0xC000:
		// TODO: setRegisterToRandom(cpu, opcode)
	case 0xD000:
		drawSprite(cpu, opcode, window)
	case 0xE000:
		switch opcode & 0x00FF {
		case 0x009E:
			// TODO: skipIfKeyPressed(cpu, opcode)
		case 0x00A1:
			// TODO: skipIfKeyNotPressed(cpu, opcode)
		}
	case 0xF000:
		switch opcode & 0x00FF {
		case 0x0007:
			// TODO: setRegisterToDelayTimer(cpu, opcode)
		case 0x000A:
			// TODO: waitForKeyPress(cpu, opcode)
		case 0x0015:
			// TODO: setDelayTimer(cpu, opcode)
		case 0x0018:
			// TODO: setSoundTimer(cpu, opcode)
		case 0x001E:
			addToIndexRegister(cpu, opcode)
		case 0x0029:
			// TODO: setIndexRegisterToSprite(cpu, opcode)
		case 0x0033:
			storeBinaryCodedDecimal(cpu, opcode)
		case 0x0055:
			storeRegisters(cpu, opcode)
		case 0x0065:
			loadRegisters(cpu, opcode)
		}
	}
}

func clearDisplay(cpu *CPU, window *pixelgl.Window) {
	imd := imdraw.New(nil)
	imd.Color = colornames.Black

	imd.Push(pixel.V(0, 0), pixel.V(64*PIXEL_SIZE, 32*PIXEL_SIZE))
	imd.Rectangle(0)
	imd.Draw(window)
}

func setRegister(cpu *CPU, opcode uint16) {
	cpu.v[(opcode&0x0F00)>>8] = uint8(opcode & 0x00FF)
}

func addToRegister(cpu *CPU, opcode uint16) {
	cpu.v[(opcode&0x0F00)>>8] += uint8(opcode & 0x00FF)
}

func skipIfEqual(cpu *CPU, opcode uint16) {
	if cpu.v[(opcode&0x0F00)>>8] == uint8(opcode&0x00FF) {
		cpu.pc += 2
	}
}

func setRegisterToRegister(cpu *CPU, opcode uint16) {
	value := cpu.v[(opcode&0x00F0)>>4]
	cpu.v[(opcode&0x0F00)>>8] = value
}

func setRegisterToRegisterOr(cpu *CPU, opcode uint16) {
	value := cpu.v[(opcode&0x0F00)>>8] | cpu.v[(opcode&0x00F0)>>4]
	cpu.v[(opcode&0x0F00)>>8] = value
}

func setRegisterToRegisterAnd(cpu *CPU, opcode uint16) {
	value := cpu.v[(opcode&0x0F00)>>8] & cpu.v[(opcode&0x00F0)>>4]
	cpu.v[(opcode&0x0F00)>>8] = value
}

func setRegisterToRegisterXor(cpu *CPU, opcode uint16) {
	value := cpu.v[(opcode&0x0F00)>>8] ^ cpu.v[(opcode&0x00F0)>>4]
	cpu.v[(opcode&0x0F00)>>8] = value
}

func addRegisterToRegister(cpu *CPU, opcode uint16) {
	vx := cpu.v[(opcode&0x0F00)>>8]
	vy := cpu.v[(opcode&0x00F0)>>4]
	vf := uint16(vx) + uint16(vy)

	cpu.v[(opcode&0x0F00)>>8] = vx + vy

	if vf > 255 {
		cpu.v[15] = 1
	} else {
		cpu.v[15] = 0
	}
}

func subtractRegisterFromRegister(cpu *CPU, opcode uint16, swap bool) {

	vx := cpu.v[(opcode&0x0F00)>>8]
	vy := cpu.v[(opcode&0x00F0)>>4]

	if swap {
		vx, vy = vy, vx
	}

	cpu.v[(opcode&0x0F00)>>8] = vx - vy

	if vx > vy {
		cpu.v[15] = 0
	} else {
		cpu.v[15] = 1
	}
}

func shiftRegisterRight(cpu *CPU, opcode uint16) {
	leastSignificantBit := cpu.v[(opcode&0x0F00)>>8] & 0x1
	cpu.v[(opcode&0x0F00)>>8] = cpu.v[(opcode&0x0F00)>>8] >> 1
	cpu.v[15] = leastSignificantBit
}

func shiftRegisterLeft(cpu *CPU, opcode uint16) {
	mostSignificantBit := cpu.v[(opcode&0x0F00)>>8] & 0x80
	cpu.v[(opcode&0x0F00)>>8] = cpu.v[(opcode&0x0F00)>>8] << 1
	cpu.v[15] = mostSignificantBit
}

func addToIndexRegister(cpu *CPU, opcode uint16) {
	cpu.i += uint16(cpu.v[(opcode&0x0F00)>>8])
}

func storeBinaryCodedDecimal(cpu *CPU, opcode uint16) {
	value := cpu.v[(opcode&0x0F00)>>8]
	cpu.memory[cpu.i] = value / 100
	cpu.memory[cpu.i+1] = (value / 10) % 10
	cpu.memory[cpu.i+2] = (value % 100) % 10
}

func storeRegisters(cpu *CPU, opcode uint16) {
	for i := uint16(0); i <= ((opcode & 0x0F00) >> 8); i++ {
		cpu.memory[cpu.i+i] = cpu.v[i]
	}
}

func loadRegisters(cpu *CPU, opcode uint16) {
	for i := uint16(0); i <= ((opcode & 0x0F00) >> 8); i++ {
		cpu.v[i] = cpu.memory[cpu.i+i]
	}
}

func setIndexRegister(cpu *CPU, opcode uint16) {
	cpu.i = opcode & 0x0FFF
}

func drawSprite(cpu *CPU, opcode uint16, window *pixelgl.Window) {
	vx := cpu.v[(opcode&0x0F00)>>8]
	vy := cpu.v[(opcode&0x00F0)>>4]
	nibble := opcode & 0x000F
	memory := cpu.memory[cpu.i : cpu.i+uint16(nibble)]

	for i := uint8(0); i < uint8(len(memory)); i++ {
		sprite := memory[i]

		for j := uint8(0); j < 8; j++ {
			if (sprite & (0x80 >> uint(j))) != 0 {
				rect := pixel.R(
					float64(vx+j)*PIXEL_SIZE,
					float64(window.Bounds().Max.Y)-float64(vy+i+1)*PIXEL_SIZE, // Adjusted Y-coordinate
					float64(vx+j+1)*PIXEL_SIZE,
					float64(window.Bounds().Max.Y)-float64(vy+i)*PIXEL_SIZE, // Adjusted Y-coordinate
				)

				imd := imdraw.New(nil)
				imd.Color = colornames.White

				imd.Push(rect.Min, rect.Max)
				imd.Rectangle(0)
				imd.Draw(window)
			}
		}
	}
}

func jumpToAddress(cpu *CPU, address uint16) {
	cpu.pc = address
}

func callSubroutine(cpu *CPU, address uint16) {
	cpu.stack[cpu.sp] = cpu.pc
	cpu.sp++
	cpu.pc = address
}

func returnFromSubroutine(cpu *CPU) {
	cpu.sp--
	cpu.pc = cpu.stack[cpu.sp]
}

func runWindow() {

	program := getProgram()

	cfg := pixelgl.WindowConfig{
		Title:  "Chip-8 Emulator",
		Bounds: pixel.R(0, 0, 64*PIXEL_SIZE, 32*PIXEL_SIZE),
	}

	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	cpu := CPU{}

	cpu.loadProgram(program)

	for !win.Closed() {
		cpuLoop(&cpu, win)
		win.Update()
	}
}
