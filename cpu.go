package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
)

type CPU struct {
	memory   [4096]uint8
	pc       uint16
	v        [16]uint8
	i        uint16
	sp       uint8
	dt       uint8
	st       uint8
	stack    [16]uint16
	keyboard Keyboard
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

	if (cpu.pc + 1) >= 4096 {
		return
	}

	// 1. Fetch
	opcode := fetch(cpu)
	cpu.pc += 2

	// 2. Decode and execute
	decode(cpu, opcode, window)

	// 3. Update timers
	cpu.dt--
	cpu.st--
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
	return uint16(cpu.memory[cpu.pc])<<8 | uint16(cpu.memory[cpu.pc+1])
}

func decode(cpu *CPU, opcode uint16, window *pixelgl.Window) {
	fmt.Printf("opcode: %x\n", opcode)
	switch opcode & 0xF000 {
	case 0x00E0:
		clearDisplay(cpu, window)
	case 0x00EE:
		// TODO: returnFromSubroutine()
	case 0x1000:
		jumpToAddress(cpu, opcode&0x0FFF)
	case 0x2000:
		// TODO: callSubroutine(opcode & 0x0FFF)
	case 0x3000:
		// TODO: skipIfEqual(cpu, opcode)
	case 0x4000:
		// TODO: skipIfNotEqual(cpu, opcode)
	case 0x5000:
		// TODO: skipIfEqualRegisters(cpu, opcode)
	case 0x6000:
		setRegister(cpu, opcode)
	case 0x7000:
		// TODO: addToRegister(cpu, opcode)
	case 0x8000:
		// TODO: setRegisterToRegister(cpu, opcode)
	case 0x8001:
		// TODO: setRegisterToRegisterOr(cpu, opcode)
	case 0x8002:
		// TODO: setRegisterToRegisterAnd(cpu, opcode)
	case 0x8003:
		// TODO: setRegisterToRegisterXor(cpu, opcode)
	case 0x8004:
		// TODO: addRegisterToRegister(cpu, opcode)
	case 0x8005:
		// TODO: subtractRegisterFromRegister(cpu, opcode)
	case 0x8006:
		// TODO: shiftRegisterRight(cpu, opcode)
	case 0x8007:
		// TODO: subtractRegisterFromRegister(cpu, opcode)
	case 0x800E:
		// TODO: shiftRegisterLeft(cpu, opcode)
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
	case 0xE09E:
		// TODO: skipIfKeyPressed(cpu, opcode)
	case 0xE0A1:
		// TODO: skipIfKeyNotPressed(cpu, opcode)
	case 0xF007:
		// TODO: setRegisterToDelayTimer(cpu, opcode)
	case 0xF00A:
		// TODO: waitForKeyPress(cpu, opcode)
	case 0xF015:
		// TODO: setDelayTimer(cpu, opcode)
	case 0xF018:
		// TODO: setSoundTimer(cpu, opcode)
	case 0xF01E:
		// TODO: addToIndexRegister(cpu, opcode)
	case 0xF029:
		// TODO: setIndexRegisterToSprite(cpu, opcode)
	case 0xF033:
		// TODO: storeBinaryCodedDecimal(cpu, opcode)
	case 0xF055:
		// TODO: storeRegisters(cpu, opcode)
	case 0xF065:
		// TODO: loadRegisters(cpu, opcode)
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
	fmt.Printf("jumping to address: %x\n from pc: %x\n", address, cpu.pc)
	cpu.pc = address

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
