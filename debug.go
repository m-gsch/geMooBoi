package main

import "fmt"

func printDebugInfo() {

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

}
