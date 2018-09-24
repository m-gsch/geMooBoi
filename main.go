package main

//Execute $CF41 incorrect, 12 times correct 72993 instructions
//13 times 254950 instructions incorrect(should not be called)
//line 668082 second file == 0

//LDH  A,(#FF44h) line 16970 LY==0x90 (144) screen printed!!!
import (
	"os"
)

var memory [0x10000]byte
var instNumber int
var instructionDEBUG byte
var cyclesPassed int
var frequency = 4096
var timerCounter = clockSpeed / frequency
var dividerCounter int
var interruptMaster bool
var scanlineCounter = 456
var screenData [160][144][4]byte

func main() {

	start()
}

func start() {
	initValues()
	f, err := os.Open("tetris.gb")

	if err != nil {
		panic(err)
	}

	defer f.Close()
	f.Read(memory[:0x8000])

	showWindow()
	/* 	for {
		//regs.flag.NZ = !regs.flag.Z //Shitty workaround
		//regs.flag.NC = !regs.flag.C //Shitty workaround
		////73895
		////82647 af=0080/90C0
		//instNumber++
		//instructionDEBUG = readAddress(regs.PC)
		//regs.PC++
		//
		//decodeIns(instructionDEBUG)
		updateState()

	} */
}

func initValues() {
	regs.A = initA
	regs.flag.Z = initFlagZ
	regs.flag.N = initFlagN
	regs.flag.H = initFlagH
	regs.flag.C = initFlagC
	regs.B = initB
	regs.C = initC
	regs.D = initD
	regs.E = initE
	regs.H = initH
	regs.L = initL
	regs.PC = initPC
	regs.SP = initSP
}

func updateState() {

	cyclesInUpdate := 0
	for cyclesInUpdate < 69905 {
		regs.flag.NZ = !regs.flag.Z //Shitty workaround
		regs.flag.NC = !regs.flag.C //Shitty workaround
		instructionDEBUG = readAddress(regs.PC)
		regs.PC++
		decodeIns(instructionDEBUG)
		cyclesInUpdate += cyclesPassed

		updateTimers()

		updateGraphics()

		checkInterrupts()

		cyclesPassed = 0 //Testing
	}
	//Render screen
}

func updateTimers() {
	updateDividerRegister()

	if clockEnabled() {
		timerCounter -= cyclesPassed

		if timerCounter <= 0 {
			setClockFreq()

			if readAddress(TIMA) == 0xFF {
				writeAddress(TIMA, readAddress(TMA))
				reqInterrupt(2)

			} else {
				writeAddress(TIMA, readAddress(TIMA)+1)
			}
		}
	}
}

func serveInterrupt(id uint) {
	interruptMaster = false
	req := readAddress(IF)
	req &^= 0x1 << id
	writeAddress(IF, req)

	pcL, pcH := uint16ToBytes(regs.PC)
	regs.SP--
	writeAddress(regs.SP, pcH)
	regs.SP--
	writeAddress(regs.SP, pcL)
	switch id {
	case 0:
		regs.PC = 0x40
	case 1:
		regs.PC = 0x48
	case 2:
		regs.PC = 0x50
	case 4:
		regs.PC = 0x60
	}
}

func updateGraphics() {
	setLCDStatus()

	if lcdEnabled() {
		scanlineCounter -= cyclesPassed
		if scanlineCounter <= 0 {
			memory[LY]++
			currentLine := readAddress(LY) - 1

			scanlineCounter = 456

			switch {
			case currentLine < 144:
				drawScanline()
			case currentLine == 144:
				reqInterrupt(0)
			case currentLine > 153:
				memory[LY] = 0
			}

		}
	}
}

func drawScanline() {
	control := readAddress(LCDC)
	if control&0x1 == 0x1 {
		renderTiles()
	}

	if control>>1&0x1 == 0x1 {
		//renderSprites
	}
}

//Probably some shit wrong
func renderTiles() {
	var tileData uint16
	var tileMap uint16

	scrollY := readAddress(SCY)
	scrollX := readAddress(SCX)

	if readAddress(LCDC)>>4&0x1 == 0x1 {
		tileData = 0x8000
	} else {
		tileData = 0x9000
	}

	if readAddress(LCDC)>>3&0x1 == 0x1 {
		tileMap = 0x9C00
	} else {
		tileMap = 0x9800
	}

	ly := readAddress(LY)
	yPos := ly + scrollY

	tileRow := uint16(yPos/8) * 32

	for i := 0; i < 160; i++ {

		xPos := i + int(scrollX)
		tileCol := uint16(xPos / 8)

		tileAddr := tileMap + tileRow + tileCol

		var tileLocation uint16
		if readAddress(LCDC)>>4&0x1 == 0x1 {
			tileID := uint16(readAddress(tileAddr))
			tileLocation = tileData + tileID*16
		} else {
			tileID := int8(readAddress(tileAddr))
			tileLocation = uint16(int16(tileData) + int16(tileID)*16)
		}
		line := yPos % 8
		line *= 2

		data1 := readAddress(tileLocation + uint16(line))
		data2 := readAddress(tileLocation + uint16(line) + 1)

		colorBit := xPos % 8
		colorBit -= 7
		colorBit *= -1

		colorNum := data2 >> uint(colorBit) & 0x1
		colorNum <<= 1
		colorNum |= data1 >> uint(colorBit) & 0x1

		col := getColor(colorNum)
		var red byte
		var green byte
		var blue byte

		switch col {
		case 0:
			red = 0xFF
			green = 0xFF
			blue = 0xFF
		case 1:
			red = 0xCC
			green = 0xCC
			blue = 0xCC
		case 2:
			red = 0x77
			green = 0x77
			blue = 0x77
		}

		pixels[(i*4)+(160*4*(int(ly-1)))] = red
		pixels[(i*4)+1+(160*4*(int(ly-1)))] = green
		pixels[(i*4)+2+(160*4*(int(ly-1)))] = blue
		pixels[(i*4)+3+(160*4*(int(ly-1)))] = 0xff

	}
}

func getColor(colorNum byte) byte {

	palette := readAddress(BGP)

	var hi uint
	var lo uint

	switch colorNum {
	case 0:
		hi = 1
		lo = 0
	case 1:
		hi = 3
		lo = 2
	case 2:
		hi = 5
		lo = 4
	case 3:
		hi = 7
		lo = 6
	}

	color := palette >> hi & 0x1
	color <<= 1
	color |= palette >> lo & 0x1

	return color
}
