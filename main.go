package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

//Graphical bug in Kirby when screen wraps around (only background, window seems to be working correctly)
//Sprite rendering index out of range BGB test

//Define BOOTROM
const (
	BOOTROM = "bootrom.bin"
)

var memory [0x10000]byte
var cartridgeMemory []byte
var instNumber int
var instructionDEBUG byte
var cyclesPassed int
var frequency = 4096
var timerCounter = clockSpeed / frequency
var dividerCounter uint16 //Initial value 0xABCC
var interruptMaster bool
var scanlineCounter = 456
var joypadState byte

//MBC - Memory Bank Controller
var MBC int
var currentROMBank uint16 = 1
var gameTitle string

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Execute with arguments.")
		return
	}
	loadCartridge()
	loadBootRom()
	showWindow()
}

func loadBootRom() {
	f, err := os.Open("roms" + string(os.PathSeparator) + BOOTROM)

	if err != nil {
		panic(err)
	}

	defer f.Close()
	f.Read(memory[:0x100])
}

func loadCartridge() {
	var ROMsize [1]byte
	var err error
	var f *os.File

	if filepath.IsAbs(os.Args[1]) {
		f, err = os.Open(os.Args[1])
	} else {
		f, err = os.Open("roms" + string(os.PathSeparator) + os.Args[1])
	}

	if err != nil {
		panic(err)
	}

	defer f.Close()

	f.ReadAt(ROMsize[:], 0x148)

	switch ROMsize[0] {
	case 0:
		cartridgeMemory = make([]byte, 0x8000)
	case 1:
		cartridgeMemory = make([]byte, 0x10000)
	case 2:
		cartridgeMemory = make([]byte, 0x20000)
	case 3:
		cartridgeMemory = make([]byte, 0x40000)
	case 4:
		cartridgeMemory = make([]byte, 0x80000)
	case 5:
		cartridgeMemory = make([]byte, 0x100000)
	case 6:
		cartridgeMemory = make([]byte, 0x200000)
	case 7:
		cartridgeMemory = make([]byte, 0x400000)
	}

	f.Read(cartridgeMemory)

	titleBytes := make([]byte, 16)
	copy(titleBytes, cartridgeMemory[0x134:0x143])

	gameTitle = strings.Title(strings.ToLower(string(titleBytes)))

	switch cartridgeMemory[0x147] {
	case 1, 2, 3:
		MBC = 1
	case 5, 6:
		MBC = 2

	}

	copy(memory[:0x4000], cartridgeMemory[:0x4000])
}

var counter = 0

func updateState() {

	setJoypadState()

	cyclesInUpdate := 0
	for cyclesInUpdate < 69905 {
		regs.flag.NZ = !regs.flag.Z //Shitty workaround
		regs.flag.NC = !regs.flag.C //Shitty workaround

		instructionDEBUG = readAddress(regs.PC)

		/* if regs.PC > 0x0100 {
			fmt.Printf("Ins: %02X %02X\n", instructionDEBUG, readAddress(regs.PC+1))
			fmt.Printf("af= %02X%02X\n", regs.A, getRegF())
			fmt.Printf("bc= %02X%02X\n", regs.B, regs.C)
			fmt.Printf("de= %02X%02X\n", regs.D, regs.E)
			fmt.Printf("hl= %02X%02X\n", regs.H, regs.L)
			fmt.Printf("sp= %04X\n", regs.SP)
			fmt.Printf("pc= %04X\n", regs.PC)
			fmt.Printf("lcdc= %02X\n", readAddress(LCDC))
			fmt.Printf("stat= %02X\n", readAddress(STAT))
			fmt.Printf("ly= %02X\n", readAddress(LY))
			fmt.Printf("ie= %02X\n", readAddress(IE))
			fmt.Printf("if= %02X\n", readAddress(IF))
			fmt.Scanln()
		} */
		regs.PC++

		decodeIns(instructionDEBUG)
		cyclesInUpdate += cyclesPassed

		updateTimers()

		updateGraphics()

		checkInterrupts()

		cyclesPassed = 0 //Testing
	}
	//Fill screen white if lcd is disabled
	if !lcdEnabled() {
		for i := range pixels {
			pixels[i] = 0xFF
		}
	}

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
		renderSprites()
	}
}

//Probably some shit wrong
func renderTiles() {
	var tileData uint16
	var tileMap uint16

	scrollY := int(readAddress(SCY))
	scrollX := int(readAddress(SCX))
	windowY := int(readAddress(WY))
	windowX := int(readAddress(WX)) - 7
	ly := int(readAddress(LY))

	y := ly - 1 //To create pixel array

	var isWindow bool
	// is the window enabled?
	if readAddress(LCDC)>>5&0x1 == 0x1 && windowY <= ly {
		// is the current scanline we're drawing
		// within the windows Y pos?
		isWindow = true
	}

	if readAddress(LCDC)>>4&0x1 == 0x1 {
		tileData = 0x8000
	} else {
		tileData = 0x9000
	}

	if isWindow {
		if readAddress(LCDC)>>6&0x1 == 0x1 {
			tileMap = 0x9C00
		} else {
			tileMap = 0x9800
		}
	} else {
		if readAddress(LCDC)>>3&0x1 == 0x1 {
			tileMap = 0x9C00
		} else {
			tileMap = 0x9800
		}
	}

	var yPos int
	if isWindow {
		yPos = ly - windowY
	} else {
		yPos = ly + scrollY
	}

	tileRow := uint16(yPos/8) * 32

	for i := 0; i < 160; i++ {
		var xPos int
		if isWindow && i >= windowX {
			xPos = i - windowX
		} else {
			xPos = i + scrollX
		}
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

		color := getColor(colorNum, BGP)
		var red byte
		var green byte
		var blue byte

		switch color {
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

		pixels[(i*4)+(160*4*y)] = red
		pixels[(i*4)+1+(160*4*y)] = green
		pixels[(i*4)+2+(160*4*y)] = blue

	}
}

func renderSprites() {

	var using16bit bool

	if readAddress(LCDC)>>2&0x1 == 0x1 {
		using16bit = true
	}

	for i := 0; i < 40; i++ {
		index := uint16(i * 4)
		yPos := int(readAddress(0xFE00+index)) - 16
		xPos := int(readAddress(0xFE00+index+1)) - 8
		tileLocation := readAddress(0xFE00 + index + 2)
		attributes := readAddress(0xFE00 + index + 3)

		ly := int(readAddress(LY))
		y := ly - 1 //To create pixel array

		var ysize int

		if using16bit {
			ysize = 16
		} else {
			ysize = 8
		}

		// does this sprite intercept with the scanline?
		if ly >= yPos && ly < yPos+ysize {

			line := ly - yPos

			// read the sprite in backwards in the y axis
			if attributes>>6&0x1 == 0x1 {
				line -= ysize - 1
				line *= -1
			}

			line *= 2
			dataAddress := 0x8000 + uint16(tileLocation)*16 + uint16(line)
			data1 := readAddress(dataAddress)
			data2 := readAddress(dataAddress + 1)

			// its easier to read in from right to left as pixel 0 is
			// bit 7 in the colour data, pixel 1 is bit 6 etc...
			for tilePixel := 7; tilePixel >= 0; tilePixel-- {
				colorBit := tilePixel

				if attributes>>5&0x1 == 0x1 {
					colorBit -= 7
					colorBit *= -1
				}
				// the rest is the same as for tiles
				colorNum := data2 >> uint(colorBit) & 0x1
				colorNum <<= 1
				colorNum |= data1 >> uint(colorBit) & 0x1

				if colorNum == 0 {
					continue //color 0 is transparent for sprites
				}

				var paletteAddr uint16
				if attributes>>4&0x1 == 0x1 {
					paletteAddr = OBP1
				} else {
					paletteAddr = OBP0
				}

				color := getColor(colorNum, paletteAddr)
				var red byte
				var green byte
				var blue byte

				switch color {
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

				xPix := 7 - tilePixel

				pixel := xPos + xPix

				pixels[(pixel*4)+(160*4*y)] = red
				pixels[(pixel*4)+1+(160*4*y)] = green
				pixels[(pixel*4)+2+(160*4*y)] = blue
			}
		}
	}
}

func getColor(colorNum byte, addr uint16) byte {

	palette := readAddress(addr)

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
