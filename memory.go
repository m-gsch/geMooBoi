package main

import (
	"github.com/hajimehoshi/ebiten"
)

/*
0000-3FFF 16KB ROM Bank 00 (in cartridge, fixed at bank 00)
4000-7FFF 16KB ROM Bank 01..NN (in cartridge, switchable bank number)
8000-9FFF 8KB Video RAM (VRAM) (switchable bank 0-1 in CGB Mode)
A000-BFFF 8KB External RAM (in cartridge, switchable bank, if any)
C000-CFFF 4KB Work RAM Bank 0 (WRAM)
D000-DFFF 4KB Work RAM Bank 1 (WRAM) (switchable bank 1-7 in CGB Mode)
E000-FDFF Same as C000-DDFF (ECHO) (typically not used)
FE00-FE9F Sprite Attribute Table (OAM)
FEA0-FEFF Not Usable
FF00-FF7F I/O Ports
FF80-FFFE High RAM (HRAM)
FFFF Interrupt Enable Register
*/
const (
	JOYP = 0xFF00
	DIV  = 0xFF04
	TIMA = 0xFF05
	TMA  = 0xFF06
	TAC  = 0xFF07
	IF   = 0xFF0F
	LCDC = 0xFF40
	STAT = 0xFF41
	SCY  = 0xFF42
	SCX  = 0xFF43
	LY   = 0xFF44
	LYC  = 0xFF45
	DMA  = 0xFF46
	WY   = 0xFF4A
	WX   = 0xFF4B
	BGP  = 0xFF47
	OBP0 = 0xFF48
	OBP1 = 0xFF49
	BOOT = 0xFF50
	IE   = 0xFFFF
)

type memoryBankController struct {
	cType     byte
	ROMbank   byte
	RAMbank   byte
	RAMenable bool
	mode      byte
}

var mbc memoryBankController

func readAddress(addr uint16) byte {

	switch {
	case addr >= 0x4000 && addr < 0x8000: // are we reading from the rom memory bank
		return cartridgeMemory[uint(addr)+uint(mbc.ROMbank-1)*0x4000]
	case addr == JOYP:
		return getJoypadState()
	case addr == IF:
		return memory[IF] | 0xE0 //Only the 5 lower bits of this register are (R/W), the others return '1' always when read
	case addr == DIV:
		return byte(dividerCounter >> 8)
	case addr == STAT:
		return memory[STAT] | 0x80 // Bit 7 is unused and always return '1'
	default:
		return memory[addr]
	}

}

//When RAM access is disabled, all writes to the external RAM area 0xA000-0xBFFF are ignored, and reads
//return 0xFF. Pan Docs recommends disabling RAM when it’s not being accessed to protect the contents
func writeAddress(addr uint16, b byte) {

	switch {
	case addr < 0x2000:
		switch mbc.cType {
		case 0x2, 0x3:
			if b&0xF == 0xA {
				mbc.RAMenable = true
			} else {
				mbc.RAMenable = false
			}

		}
	case addr >= 0x2000 && addr < 0x4000:
		switch mbc.cType {
		case 0x1, 0x2, 0x3: //MBC1
			b &= 0x1F

			if b == 0 {
				b++
			}

			mbc.ROMbank &= 0x60
			mbc.ROMbank |= b
		case 0x5, 0x6: //MBC2
			b &= 0xF

			if b == 0 {
				b++
			}

			mbc.ROMbank = b
		case 0x0F, 0x10, 0x11, 0x12, 0x13: //MBC3
			b &= 0x7F

			if b == 0 {
				b++
			}

			mbc.ROMbank = b
		}
	case addr >= 0x4000 && addr < 0x6000:
		switch mbc.cType {
		case 0x1, 0x2, 0x3:
			b &= 0x3
			if mbc.mode == 0x1 {
				mbc.RAMbank = b
			} else {
				mbc.ROMbank &= 0x1F
				mbc.ROMbank |= b << 5
			}
		}
	case addr >= 0x6000 && addr < 0x8000:
		switch mbc.cType {
		case 0x2, 0x3:
			mbc.mode = b & 0x1
		}
	case addr >= 0xFEA0 && addr < 0xFEFF:
	case addr >= 0xE000 && addr < 0xFE00:
		memory[addr] = b
		writeAddress(addr-0x2000, b)
	case addr == TAC:
		oldFreq := getClockFreq()
		memory[TAC] = b
		newFreq := getClockFreq()
		if newFreq != oldFreq {
			setClockFreq()
		}
	case addr == DIV:
		dividerCounter = 0
	case addr == LY:
		memory[LY] = 0
	case addr == DMA:
		dmaTransfer(b)
	case addr == BOOT: //After BOOT is done writes are ignored
	default:
		memory[addr] = b
	}
}

func clockEnabled() bool {
	b := readAddress(TAC) >> 2 & 0x1

	if b == 0 {
		return false
	}

	return true
}

func getClockFreq() byte {
	return readAddress(TAC) & 0x3
}

func setClockFreq() {
	clockFreq := getClockFreq()
	switch clockFreq {
	case 0:
		timerCounter = 1024
	case 1:
		timerCounter = 16
	case 2:
		timerCounter = 64
	case 3:
		timerCounter = 256
	}
}

func updateDividerRegister() {
	dividerCounter += uint16(cyclesPassed)
}

func reqInterrupt(id uint) {
	req := readAddress(IF) & 0x1F
	req |= 0x1 << id
	writeAddress(IF, req)
}

func checkInterrupts() {
	if interruptMaster {
		req := readAddress(IF) & 0x1F
		enabled := readAddress(IE)
		if req > 0 {
			for i := uint(0); i < 5; i++ {
				if req>>i&0x1 == 0x1 && enabled>>i&0x1 == 0x1 {
					serveInterrupt(i)
				}
			}
		}
	}
}

func setLCDStatus() {
	status := readAddress(STAT)
	if lcdEnabled() {
		currentLine := readAddress(LY)
		currentMode := status & 0x3
		var mode byte
		reqInt := false
		if currentLine >= 144 {
			mode = 1
			status &^= 0x2
			status |= 0x1
			if status>>4&0x1 == 0x1 {
				reqInt = true
			}
		} else {
			mode2bounds := 456 - 80
			mode3bounds := mode2bounds - 172
			switch {
			case scanlineCounter >= mode2bounds:
				mode = 2
				status &^= 0x1
				status |= 0x2
				if status>>5&0x1 == 0x1 {
					reqInt = true
				}
			case scanlineCounter >= mode3bounds:
				mode = 3
				status |= 0x3
			default:
				mode = 0
				status &^= 0x3
				if status>>3&0x1 == 0x1 {
					reqInt = true
				}
			}
		}
		if reqInt && mode != currentMode {
			reqInterrupt(1)
		}
		if currentLine == readAddress(LYC) {
			status |= 0x4
			if status>>6&0x1 == 0x1 {
				reqInterrupt(1)
			}
		} else {
			status &^= 0x4
		}
		writeAddress(STAT, status)
	} else {
		scanlineCounter = 456
		memory[LY] = 0
		status &^= 0x7 //Bits 0-2 return '0' when the LCD is off
		writeAddress(STAT, status)
	}

}

func lcdEnabled() bool {
	b := readAddress(LCDC) >> 7 & 0x1

	if b == 0 {
		return false
	}

	return true
}

func dmaTransfer(b byte) {
	addr := uint16(b) << 8
	for i := uint16(0); i < 0xA0; i++ {
		writeAddress(0xFE00+i, readAddress(addr+i))
	}
}

func getJoypadState() byte {
	joyp := memory[JOYP]

	joyp = ^joyp

	if joyp>>4&0x1 == 0x1 {
		newJoyp := joypadState & 0x0F
		joyp &= 0xF0
		joyp |= newJoyp

	} else if joyp>>5&0x1 == 0x1 {
		newJoyp := joypadState >> 4
		joyp &= 0xF0
		joyp |= newJoyp
	}
	return ^joyp
}

func setJoypadState() {

	keys := []ebiten.Key{ebiten.KeyRight,
		ebiten.KeyLeft,
		ebiten.KeyUp,
		ebiten.KeyDown,
		ebiten.KeyA,
		ebiten.KeyS,
		ebiten.KeySpace,
		ebiten.KeyEnter}

	var newJoypadState byte

	for i, key := range keys {
		if ebiten.IsKeyPressed(key) {
			newJoypadState |= 0x1 << uint(i)
		}
	}

	if joypadState^newJoypadState > 0 {
		reqInterrupt(4)
	}

	joypadState = newJoypadState
}
