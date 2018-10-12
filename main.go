package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/hajimehoshi/ebiten"
)

//Graphics rendering needs work, it is very ugly
//Use type interface for registers (?)

var memory [0x10000]byte
var cartridgeMemory []byte
var instNumber int
var instructionDEBUG byte
var cyclesPassed int
var frequency = 4096
var timerCounter = clockSpeed / frequency
var dividerCounter uint16
var interruptMaster bool
var scanlineCounter = 456
var joypadState byte

var gameTitle string

func init() {
	if len(os.Args) < 2 {
		fmt.Println("Execute with arguments.")
		return
	}
	regs.PC = 0x100
	regs.A = 0x01
	setRegF(0x13)
	regs.B = 0x00
	regs.C = 0x13
	regs.D = 0x00
	regs.E = 0xD8
	regs.H = 0x01
	regs.L = 0x4D
	regs.SP = 0xFFFE
	memory[0xFF10] = 0x80
	memory[0xFF11] = 0xBF
	memory[0xFF12] = 0xF3
	memory[0xFF14] = 0xBF
	memory[0xFF16] = 0x3F
	memory[0xFF19] = 0xBF
	memory[0xFF1A] = 0x7F
	memory[0xFF1B] = 0xFF
	memory[0xFF1C] = 0x9F
	memory[0xFF1E] = 0xBF
	memory[0xFF20] = 0xFF
	memory[0xFF23] = 0xBF
	memory[0xFF24] = 0x77
	memory[0xFF25] = 0xF3
	memory[0xFF26] = 0xF1
	memory[LCDC] = 0x91
	memory[BGP] = 0xFC
	memory[OBP0] = 0xFF
	memory[OBP1] = 0xFF
	memory[BOOT] = 0xFF //BOOT_OFF
	dividerCounter = 0xABCC
	loadCartridge()

	//Fill screen with white pixels
	for i := range pixels {
		pixels[i] = 0xFF
	}
}

func main() {
	if err := ebiten.Run(update, initScreenWidth, initScreenHeight, initScreenScale, gameTitle); err != nil {
		log.Fatal(err)
	}
}

func loadCartridge() {
	var mbcHeader [3]byte
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

	//[0x147] Cartridge type
	//ROM Size = 32kB << [0x148]
	//[0x149]Size of external RAM
	f.ReadAt(mbcHeader[:], 0x147)

	mbc = memoryBankController{cType: mbcHeader[0], ROMbank: 1, RAMbank: 0, RAMenable: false, mode: 1}

	cartridgeMemory = make([]byte, 0x8000<<mbcHeader[1])
	fmt.Printf("Cartridge type: %0X\n", mbcHeader[0])
	fmt.Printf("Size: %d\n", 0x8000<<mbcHeader[1])

	f.Read(cartridgeMemory)

	titleBytes := make([]byte, 16)
	copy(titleBytes, cartridgeMemory[0x134:0x144])

	gameTitle = strings.Title(strings.ToLower(string(titleBytes)))

	copy(memory[:0x4000], cartridgeMemory[:0x4000])
}

func updateState() {

	setJoypadState()

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

	scrollY := readAddress(SCY)
	scrollX := readAddress(SCX)
	windowY := readAddress(WY)
	windowX := readAddress(WX)

	if windowX < 7 { //If WX is set <7 it's treated as if it was 7
		windowX = 7
	}

	windowX -= 7
	ly := readAddress(LY) - 1
	y := int(ly)

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

	var yPos byte
	if isWindow {
		yPos = ly - windowY
	} else {
		yPos = ly + scrollY
	}

	tileRow := uint16(yPos/8) * 32

	for i := 0; i < 160; i++ {
		var xPos byte
		if isWindow && i >= int(windowX) {
			xPos = byte(i) - windowX
		} else {
			xPos = byte(i) + scrollX
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

		colorBit := int(xPos) % 8
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

		ly := readAddress(LY) - 1
		y := int(ly) //To create pixel array

		var ysize int

		if using16bit {
			ysize = 16
		} else {
			ysize = 8
		}

		// does this sprite intercept with the scanline?
		if int(ly) >= yPos && int(ly) < yPos+ysize {

			line := int(ly) - yPos

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

				x := xPos + xPix
				if x < 0 {
					continue
				}
				if attributes>>7&0x1 == 0x1 && pixels[(x*4)+(160*4*y)] != 0xff {
					//Bit priority set sprite behind bg expect if bg color is 0
					//Needs to know if it's color 0 right now I don't save that
					continue
				}
				pixels[(x*4)+(160*4*y)] = red
				pixels[(x*4)+1+(160*4*y)] = green
				pixels[(x*4)+2+(160*4*y)] = blue
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
