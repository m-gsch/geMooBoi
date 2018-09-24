package main

import (
	"fmt"
)

//TODO: STOP, HALT
var tableR = [8]*byte{&regs.B, //B
	&regs.C, //C
	&regs.D, //D
	&regs.E, //E
	&regs.H, //H
	&regs.L, //L
	nil,     //HL,SP,AF
	&regs.A} //A

var tableCC = [5]*bool{&regs.flag.NZ, //NZ
	&regs.flag.Z,  //Z
	&regs.flag.NC, //NC
	&regs.flag.C}  //C

func decodeIns(opcode byte) {
	x := opcode >> 6
	y := opcode >> 3 & 0x7
	p := y >> 1
	z := opcode & 0x7
	ins := opcode >> 3

	switch opcode {
	case 0x00:
		nop()
	case 0x10:
		stop()
	case 0x76:
		halt()
	case 0xCB:
		prefixCB()
	case 0xF3:
		di()
	case 0xFB:
		ei()
	case 0xD3, 0xDB, 0xDD, 0xE3, 0xE4, 0xEB, 0xEC, 0xED, 0xF4, 0xFC, 0xFD: // Invalid
		fmt.Printf("Wrong instruction:%x\n PC:%x\n", opcode, regs.PC)
	case 0x01, 0x11, 0x21, 0x31:
		ldR16d16(tableR[2*p], tableR[2*p+1])
	case 0x02, 0x12:
		ldaR16A(tableR[2*p], tableR[2*p+1])
	case 0x08:
		lda16SP()
	case 0x0A, 0x1A:
		ldAaR16(tableR[2*p], tableR[2*p+1])
	case 0x22:
		ldaHLIA()
	case 0x32:
		ldaHLDA()
	case 0x2A:
		ldAaHLI()
	case 0x3A:
		ldAaHLD()
	case 0x03, 0x13, 0x23, 0x33:
		incR16(tableR[2*p], tableR[2*p+1])
	case 0x09, 0x19, 0x29, 0x39:
		addHLR16(tableR[2*p], tableR[2*p+1])
	case 0x0B, 0x1B, 0x2B, 0x3B:
		decR16(tableR[2*p], tableR[2*p+1])
	case 0xE8:
		addSPd8()
	case 0x07:
		rlca()
	case 0x17:
		rla()
	case 0x0F:
		rrca()
	case 0x1F:
		rra()
	case 0x27:
		daa()
	case 0x37:
		scf()
	case 0x2F:
		cpl()
	case 0x3F:
		ccf()
	case 0x20, 0x28, 0x30, 0x38:
		jrCCd8(*tableCC[y-4])
	case 0x18:
		jrCCd8(true)
	case 0xC2, 0xCA, 0xD2, 0xDA:
		jpCCa16(*tableCC[y])
	case 0xC3:
		jpCCa16(true)
	case 0xE9:
		jpaHL()
	case 0xC6:
		addAd8()
	case 0xCE:
		adcAd8()
	case 0xD6:
		subAd8()
	case 0xDE:
		sbcAd8()
	case 0xE6:
		andAd8()
	case 0xEE:
		xorAd8()
	case 0xF6:
		orAd8()
	case 0xFE:
		cpAd8()
	case 0xE0:
		ldha8A()
	case 0xF0:
		ldhAa8()
	case 0xE2:
		ldhaCA()
	case 0xF2:
		ldhAaC()
	case 0xEA:
		lda16A()
	case 0xFA:
		ldAa16()
	case 0xF8:
		ldHLSPd8()
	case 0xF9:
		ldSPHL()
	case 0xC1, 0xD1, 0xE1, 0xF1:
		popR16(tableR[2*p], tableR[2*p+1])
	case 0xC5, 0xD5, 0xE5, 0xF5:
		pushR16(tableR[2*p], tableR[2*p+1])
	case 0xC7, 0xD7, 0xE7, 0xF7, 0xCF, 0xDF, 0xEF, 0xFF:
		rst(uint16(y * 8))
	case 0xC4, 0xD4, 0xCC, 0xDC:
		callCCa16(*tableCC[y])
	case 0xCD:
		callCCa16(true)
	case 0xC0, 0xD0, 0xC8, 0xD8:
		retCC(*tableCC[y])
	case 0xC9:
		ret()
	case 0xD9:
		reti()
	default:
		switch {
		case x == 0 && z == 4:
			incR8(tableR[y])
		case x == 0 && z == 5:
			decR8(tableR[y])
		case x == 0 && z == 6:
			ldR8d8(tableR[y])
		case x == 1:
			ldR8R8(tableR[y], tableR[z])
		case ins == 0x10:
			addAR8(tableR[z])
		case ins == 0x11:
			adcAR8(tableR[z])
		case ins == 0x12:
			subAR8(tableR[z])
		case ins == 0x13:
			sbcAR8(tableR[z])
		case ins == 0x14:
			andAR8(tableR[z])
		case ins == 0x15:
			xorAR8(tableR[z])
		case ins == 0x16:
			orAR8(tableR[z])
		case ins == 0x17:
			cpAR8(tableR[z])
		}
	}
}

func incR8(reg *byte) {
	//Duration: 4/12
	//Byte length: 1
	//Flags: Z:Z N:0 H:H C:-
	var val byte
	var addr uint16
	c := make(chan int, 1)
	if reg == nil {
		go clockTicks(12, c)
		addr = bytesToUint16(regs.L, regs.H)
		val = readAddress(addr)
	} else {
		go clockTicks(4, c)
		val = *reg
	}
	if (((val & 0xf) + 0x1) & 0x10) == 0x10 {
		regs.flag.H = true
	} else {
		regs.flag.H = false
	}
	regs.flag.N = false
	val++
	if val == 0x00 {
		regs.flag.Z = true
	} else {
		regs.flag.Z = false
	}
	if reg == nil {
		writeAddress(addr, val)
	} else {
		*reg = val
	}
	<-c
}

func decR8(reg *byte) {
	//Duration: 4/12
	//Byte length: 1
	//Flags: Z:Z N:1 H:H C:-
	var val byte
	var addr uint16
	c := make(chan int, 1)
	if reg == nil {
		go clockTicks(12, c)
		addr = bytesToUint16(regs.L, regs.H)
		val = readAddress(addr)
	} else {
		go clockTicks(4, c)
		val = *reg
	}
	if ((val & 0xf) - 0x1) > val {
		regs.flag.H = true
	} else {
		regs.flag.H = false
	}
	val--
	regs.flag.N = true
	if val == 0x00 {
		regs.flag.Z = true
	} else {
		regs.flag.Z = false
	}
	if reg == nil {
		writeAddress(addr, val)
	} else {
		*reg = val
	}
	<-c
}

func ldR8d8(reg *byte) {
	//Duration: 8/12
	//Byte length: 2
	//Flags: Z:- N:- H:- C:-
	c := make(chan int, 1)
	if reg == nil {
		go clockTicks(12, c)
		addr := bytesToUint16(regs.L, regs.H)
		d8 := readAddress(regs.PC)
		regs.PC++
		writeAddress(addr, d8)
	} else {
		go clockTicks(8, c)
		*reg = readAddress(regs.PC)
		regs.PC++
	}
	<-c
}

func ldR8R8(destReg, srcReg *byte) {
	//Duration: 4/8
	//Byte length: 1
	//Flags: Z:- N:- H:- C:-
	c := make(chan int, 1)
	switch {
	case destReg == nil:
		go clockTicks(8, c)
		addr := bytesToUint16(regs.L, regs.H)
		writeAddress(addr, *srcReg)
	case srcReg == nil:
		go clockTicks(8, c)
		addr := bytesToUint16(regs.L, regs.H)
		*destReg = readAddress(addr)
	default:
		go clockTicks(4, c)
		*destReg = *srcReg
	}
	<-c
}

func addAR8(reg *byte) {
	//Duration: 4/8
	//Byte length: 1
	//Flags: Z:Z N:0 H:H C:C
	var val byte
	var addr uint16
	c := make(chan int, 1)
	if reg == nil {
		go clockTicks(8, c)
		addr = bytesToUint16(regs.L, regs.H)
		val = readAddress(addr)
	} else {
		go clockTicks(4, c)
		val = *reg
	}
	if (((regs.A & 0xf) + (val & 0xf)) & 0x10) == 0x10 {
		regs.flag.H = true
	} else {
		regs.flag.H = false
	}
	if regs.A > regs.A+val {
		regs.flag.C = true
	} else {
		regs.flag.C = false
	}
	regs.flag.N = false
	regs.A += val
	if regs.A == 0x00 {
		regs.flag.Z = true
	} else {
		regs.flag.Z = false
	}
	<-c
}

func addAd8() {
	//Duration: 8
	//Byte length: 2
	//Flags: Z:Z N:0 H:H C:C
	c := make(chan int, 1)
	go clockTicks(8, c)
	val := readAddress(regs.PC)
	regs.PC++
	if (((regs.A & 0xf) + (val & 0xf)) & 0x10) == 0x10 {
		regs.flag.H = true
	} else {
		regs.flag.H = false
	}
	if regs.A > regs.A+val {
		regs.flag.C = true
	} else {
		regs.flag.C = false
	}
	regs.flag.N = false
	regs.A += val
	if regs.A == 0x00 {
		regs.flag.Z = true
	} else {
		regs.flag.Z = false
	}
	<-c
}

func adcAR8(reg *byte) {
	//Duration: 4/8
	//Byte length: 1
	//Flags: Z:Z N:0 H:H C:C
	var val byte
	var addr uint16
	c := make(chan int, 1)
	if reg == nil {
		go clockTicks(8, c)
		addr = bytesToUint16(regs.L, regs.H)
		val = readAddress(addr)
	} else {
		go clockTicks(4, c)
		val = *reg
	}
	var carry byte
	if regs.flag.C {
		carry++
	}
	result := uint16(regs.A) + uint16(val) + uint16(carry)
	if (regs.A&0xf)+(val&0xf)+carry > 0xF {
		regs.flag.H = true
	} else {
		regs.flag.H = false
	}
	if result > 0xFF {
		regs.flag.C = true
	} else {
		regs.flag.C = false
	}
	regs.flag.N = false
	regs.A = uint8(result & 0x00FF)
	if regs.A == 0x00 {
		regs.flag.Z = true
	} else {
		regs.flag.Z = false
	}
	<-c
}

func adcAd8() {
	//Duration: 8
	//Byte length: 2
	//Flags: Z:Z N:0 H:H C:C
	c := make(chan int, 1)
	go clockTicks(8, c)
	val := readAddress(regs.PC)
	regs.PC++
	var carry byte
	if regs.flag.C {
		carry++
	}
	result := uint16(regs.A) + uint16(val) + uint16(carry)
	if (regs.A&0xf)+(val&0xf)+carry > 0xF {
		regs.flag.H = true
	} else {
		regs.flag.H = false
	}
	if result > 0xFF {
		regs.flag.C = true
	} else {
		regs.flag.C = false
	}
	regs.flag.N = false
	regs.A = uint8(result & 0x00FF)
	if regs.A == 0x00 {
		regs.flag.Z = true
	} else {
		regs.flag.Z = false
	}
	<-c
}

func subAR8(reg *byte) {
	//Duration: 4/8
	//Byte length: 1
	//Flags: Z:Z N:1 H:H C:C
	var val byte
	var addr uint16
	c := make(chan int, 1)
	if reg == nil {
		go clockTicks(8, c)
		addr = bytesToUint16(regs.L, regs.H)
		val = readAddress(addr)
	} else {
		go clockTicks(4, c)
		val = *reg
	}
	if ((regs.A & 0xf) - (val & 0xf)) > regs.A {
		regs.flag.H = true
	} else {
		regs.flag.H = false
	}
	if regs.A < regs.A-val {
		regs.flag.C = true
	} else {
		regs.flag.C = false
	}
	regs.A -= val
	regs.flag.N = true
	if regs.A == 0x00 {
		regs.flag.Z = true
	} else {
		regs.flag.Z = false
	}
	<-c
}

func subAd8() {
	//Duration: 8
	//Byte length: 2
	//Flags: Z:Z N:1 H:H C:C
	c := make(chan int, 1)
	go clockTicks(8, c)
	val := readAddress(regs.PC)
	regs.PC++
	if ((regs.A & 0xf) - (val & 0xf)) > regs.A {
		regs.flag.H = true
	} else {
		regs.flag.H = false
	}
	if regs.A < regs.A-val {
		regs.flag.C = true
	} else {
		regs.flag.C = false
	}
	regs.A -= val
	regs.flag.N = true
	if regs.A == 0x00 {
		regs.flag.Z = true
	} else {
		regs.flag.Z = false
	}
	<-c
}

func sbcAR8(reg *byte) {
	//Duration: 4/8
	//Byte length: 1
	//Flags: Z:Z N:1 H:H C:C
	var val uint16
	var addr uint16
	c := make(chan int, 1)
	if reg == nil {
		go clockTicks(8, c)
		addr = bytesToUint16(regs.L, regs.H)
		val = uint16(readAddress(addr))
	} else {
		go clockTicks(4, c)
		val = uint16(*reg)
	}
	var carry uint16
	if regs.flag.C {
		carry++
	}
	if carry+(val&0xf) > uint16(regs.A&0xF) {
		regs.flag.H = true
	} else {
		regs.flag.H = false
	}
	if val+carry > uint16(regs.A) {
		regs.flag.C = true
	} else {
		regs.flag.C = false
	}
	regs.A = byte(uint16(regs.A) - val - carry)
	regs.flag.N = true
	if regs.A == 0x00 {
		regs.flag.Z = true
	} else {
		regs.flag.Z = false
	}
	<-c
}

func sbcAd8() {
	//Duration: 8
	//Byte length: 2
	//Flags: Z:Z N:1 H:H C:C
	c := make(chan int, 1)
	go clockTicks(8, c)
	val := uint16(readAddress(regs.PC))
	regs.PC++
	var carry uint16
	if regs.flag.C {
		carry++
	}
	if carry+(val&0xf) > uint16(regs.A&0xF) {
		regs.flag.H = true
	} else {
		regs.flag.H = false
	}
	if val+carry > uint16(regs.A) {
		regs.flag.C = true
	} else {
		regs.flag.C = false
	}
	regs.A = byte(uint16(regs.A) - val - carry)
	regs.flag.N = true
	if regs.A == 0x00 {
		regs.flag.Z = true
	} else {
		regs.flag.Z = false
	}
	<-c
}

func andAR8(reg *byte) {
	//Duration: 4/8
	//Byte length: 1
	//Flags: Z:Z N:0 H:1 C:0
	var val byte
	var addr uint16
	c := make(chan int, 1)
	if reg == nil {
		go clockTicks(8, c)
		addr = bytesToUint16(regs.L, regs.H)
		val = readAddress(addr)
	} else {
		go clockTicks(4, c)
		val = *reg
	}
	regs.A &= val
	if regs.A == 0x00 {
		regs.flag.Z = true
	} else {
		regs.flag.Z = false
	}
	regs.flag.N = false
	regs.flag.H = true
	regs.flag.C = false
	<-c
}

func andAd8() {
	//0xe6
	//Duration: 8
	//Byte length: 2
	//Flags: Z:Z N:0 H:1 C:0
	c := make(chan int, 1)
	go clockTicks(8, c)
	val := readAddress(regs.PC)
	regs.PC++
	regs.A &= val
	if regs.A == 0x00 {
		regs.flag.Z = true
	} else {
		regs.flag.Z = false
	}
	regs.flag.N = false
	regs.flag.H = true
	regs.flag.C = false
	<-c
}

func xorAR8(reg *byte) {
	//Duration: 4/8
	//Byte length: 1
	//Flags: Z:Z N:0 H:0 C:0
	var val byte
	var addr uint16
	c := make(chan int, 1)
	if reg == nil {
		go clockTicks(8, c)
		addr = bytesToUint16(regs.L, regs.H)
		val = readAddress(addr)
	} else {
		go clockTicks(4, c)
		val = *reg
	}
	regs.A ^= val
	if regs.A == 0x00 {
		regs.flag.Z = true
	} else {
		regs.flag.Z = false
	}
	regs.flag.N = false
	regs.flag.H = false
	regs.flag.C = false
	<-c
}

func xorAd8() {
	//Duration: 8
	//Byte length: 2
	//Flags: Z:Z N:0 H:0 C:0
	c := make(chan int, 1)
	go clockTicks(8, c)
	val := readAddress(regs.PC)
	regs.PC++
	regs.A ^= val
	if regs.A == 0x00 {
		regs.flag.Z = true
	} else {
		regs.flag.Z = false
	}
	regs.flag.N = false
	regs.flag.H = false
	regs.flag.C = false
	<-c
}

func orAR8(reg *byte) {
	//Duration: 4/8
	//Byte length: 1
	//Flags: Z:Z N:0 H:0 C:0
	var val byte
	var addr uint16
	c := make(chan int, 1)
	if reg == nil {
		go clockTicks(8, c)
		addr = bytesToUint16(regs.L, regs.H)
		val = readAddress(addr)
	} else {
		go clockTicks(4, c)
		val = *reg
	}
	regs.A |= val
	if regs.A == 0x00 {
		regs.flag.Z = true
	} else {
		regs.flag.Z = false
	}
	regs.flag.N = false
	regs.flag.H = false
	regs.flag.C = false
	<-c
}

func orAd8() {
	//Duration: 8
	//Byte length: 2
	//Flags: Z:Z N:0 H:0 C:0
	c := make(chan int, 1)
	go clockTicks(8, c)
	val := readAddress(regs.PC)
	regs.PC++
	regs.A |= val
	if regs.A == 0x00 {
		regs.flag.Z = true
	} else {
		regs.flag.Z = false
	}
	regs.flag.N = false
	regs.flag.H = false
	regs.flag.C = false
	<-c
}

func cpAR8(reg *byte) {
	//Duration: 4/8
	//Byte length: 1
	//Flags: Z:Z N:1 H:H C:C
	var val byte
	var addr uint16
	c := make(chan int, 1)
	if reg == nil {
		go clockTicks(8, c)
		addr = bytesToUint16(regs.L, regs.H)
		val = readAddress(addr)
	} else {
		go clockTicks(4, c)
		val = *reg
	}
	if ((regs.A & 0xf) - (val & 0xf)) > regs.A {
		regs.flag.H = true
	} else {
		regs.flag.H = false
	}
	if regs.A < regs.A-val {
		regs.flag.C = true
	} else {
		regs.flag.C = false
	}
	regs.flag.N = true
	if regs.A-val == 0x00 {
		regs.flag.Z = true
	} else {
		regs.flag.Z = false
	}
	<-c
}

func cpAd8() {
	//Duration: 8
	//Byte length: 2
	//Flags: Z:Z N:1 H:H C:C
	c := make(chan int, 1)
	go clockTicks(8, c)
	val := readAddress(regs.PC)
	regs.PC++
	if ((regs.A & 0xf) - (val & 0xf)) > regs.A {
		regs.flag.H = true
	} else {
		regs.flag.H = false
	}
	if regs.A < regs.A-val {
		regs.flag.C = true
	} else {
		regs.flag.C = false
	}
	regs.flag.N = true
	if regs.A-val == 0x00 {
		regs.flag.Z = true
	} else {
		regs.flag.Z = false
	}
	<-c
}

func jpCCa16(cc bool) {
	//Duration: 16/12
	//Byte length: 3
	//Flags: Z:- N:- H:- C:-
	c := make(chan int, 1)
	addrl := readAddress(regs.PC)
	regs.PC++
	addrh := readAddress(regs.PC)
	regs.PC++
	addr := bytesToUint16(addrl, addrh)
	if cc {
		go clockTicks(16, c)
		regs.PC = addr
	} else {
		go clockTicks(12, c)
	}
	<-c

}

func jpaHL() {
	//Duration: 4
	//Byte length: 1
	//Flags: Z:- N:- H:- C:-
	c := make(chan int, 1)
	go clockTicks(4, c)
	addr := bytesToUint16(regs.L, regs.H)
	regs.PC = addr
	<-c
}

func jrCCd8(cc bool) {
	//Duration: 12/8
	//Byte length: 2
	//Flags: Z:- N:- H:- C:-
	c := make(chan int, 1)
	d8 := int8(readAddress(regs.PC))
	regs.PC++
	if cc {
		go clockTicks(12, c)
		regs.PC += uint16(d8)
	} else {
		go clockTicks(8, c)
	}
	<-c
}

func daa() {
	//Duration: 4
	//Byte length: 1
	//Flags: Z:Z N:- H:0 C:C
	c := make(chan int, 1)
	go clockTicks(4, c)
	a := uint16(regs.A)
	if !regs.flag.N {
		if regs.flag.H || (a&0xF) > 9 {
			a += 0x06
		}

		if regs.flag.C || a > 0x9F {
			a += 0x60
		}
	} else {
		if regs.flag.H {
			a -= 0x06
			a &= 0xFF
		}

		if regs.flag.C {
			a -= 0x60
		}
	}

	regs.flag.H = false
	regs.flag.Z = false

	if a&0x100 == 0x100 {
		regs.flag.C = true
	}

	a &= 0xFF

	if a == 0x00 {
		regs.flag.Z = true
	}

	regs.A = byte(a)

	<-c
}

func scf() {
	//Duration: 4
	//Byte length: 1
	//Flags: Z:- N:0 H:0 C:1
	c := make(chan int, 1)
	go clockTicks(4, c)
	regs.flag.N = false
	regs.flag.H = false
	regs.flag.C = true
	<-c
}

func cpl() {
	//Duration: 4
	//Byte length: 1
	//Flags: Z:- N:1 H:1 C:-
	c := make(chan int, 1)
	go clockTicks(4, c)
	regs.flag.N = true
	regs.flag.H = true
	regs.A = ^regs.A
	<-c
}

func ccf() {
	//Duration: 4
	//Byte length: 1
	//Flags: Z:- N:0 H:0 C:C
	c := make(chan int, 1)
	go clockTicks(4, c)
	regs.flag.N = false
	regs.flag.H = false
	regs.flag.C = !regs.flag.C
	<-c
}

func rlca() {
	//Duration: 4
	//Byte length: 1
	//Flags: Z:0 N:0 H:0 C:C
	c := make(chan int, 1)
	go clockTicks(4, c)
	regs.flag.Z = false
	regs.flag.N = false
	regs.flag.H = false
	if regs.A>>7 != 0x00 {
		regs.flag.C = true
	} else {
		regs.flag.C = false
	}
	regs.A = regs.A<<1 | regs.A>>7
	<-c
}

func rla() {
	//Duration: 4
	//Byte length: 1
	//Flags: Z:0 N:0 H:0 C:C
	c := make(chan int, 1)
	go clockTicks(4, c)
	regs.flag.Z = false
	regs.flag.N = false
	regs.flag.H = false
	var b byte = 0x0
	if regs.flag.C {
		b = 0x1
	}
	if regs.A>>7 != 0x00 {
		regs.flag.C = true
	} else {
		regs.flag.C = false
	}
	regs.A = regs.A<<1 | b
	<-c
}

func rrca() {
	//Duration: 4
	//Byte length: 1
	//Flags: Z:0 N:0 H:0 C:C
	c := make(chan int, 1)
	go clockTicks(4, c)
	regs.flag.Z = false
	regs.flag.N = false
	regs.flag.H = false
	if regs.A<<7 != 0x00 {
		regs.flag.C = true
	} else {
		regs.flag.C = false
	}
	regs.A = regs.A>>1 | regs.A<<7
	<-c
}

func rra() {
	//Duration: 4
	//Byte length: 1
	//Flags: Z:0 N:0 H:0 C:C
	c := make(chan int, 1)
	go clockTicks(4, c)
	regs.flag.Z = false
	regs.flag.N = false
	regs.flag.H = false
	var carry byte
	if regs.flag.C {
		carry = 0x80
	}
	if regs.A<<7 != 0x00 {
		regs.flag.C = true
	} else {
		regs.flag.C = false
	}
	regs.A = regs.A >> 1
	regs.A |= carry
	<-c
}

func incR16(regH, regL *byte) {
	//Duration: 8
	//Byte length: 1
	//Flags: Z:- N:- H:- C:-
	c := make(chan int, 1)
	go clockTicks(8, c)
	if regH == nil {
		regs.SP++
	} else {
		r16 := bytesToUint16(*regL, *regH)
		r16++
		*regL, *regH = uint16ToBytes(r16)
	}
	<-c
}

func addHLR16(regH, regL *byte) {
	//Duration: 8
	//Byte length: 1
	//Flags: Z:- N:0 H:H C:C
	c := make(chan int, 1)
	go clockTicks(8, c)
	var r16 uint16
	if regH == nil {
		r16 = regs.SP
	} else {
		r16 = bytesToUint16(*regL, *regH)
	}
	hl := bytesToUint16(regs.L, regs.H)
	regs.flag.N = false
	if (((r16 & 0xfff) + (hl & 0xfff)) & 0x1000) == 0x1000 {
		regs.flag.H = true
	} else {
		regs.flag.H = false
	}
	if hl > hl+r16 {
		regs.flag.C = true
	} else {
		regs.flag.C = false
	}
	hl += r16
	regs.L, regs.H = uint16ToBytes(hl)
	<-c
}

func decR16(regH, regL *byte) {
	//Duration: 8
	//Byte length: 1
	//Flags: Z:- N:- H:- C:-
	c := make(chan int, 1)
	go clockTicks(8, c)
	if regH == nil {
		regs.SP--
	} else {
		r16 := bytesToUint16(*regL, *regH)
		r16--
		*regL, *regH = uint16ToBytes(r16)
	}
	<-c
}

func addSPd8() {
	//Duration: 16
	//Byte length: 2
	//Flags: Z:0 N:0 H:H C:C
	c := make(chan int, 1)
	go clockTicks(16, c)
	d8 := int8(readAddress(regs.PC))
	regs.PC++
	regs.flag.Z = false
	regs.flag.N = false
	result := uint32(int16(regs.SP) + int16(d8))
	if ((regs.SP ^ uint16(d8) ^ (uint16(result & 0xFFFF))) & 0x100) == 0x100 {
		regs.flag.C = true
	} else {
		regs.flag.C = false
	}
	if ((regs.SP ^ uint16(d8) ^ (uint16(result & 0xFFFF))) & 0x010) == 0x010 {
		regs.flag.H = true
	} else {
		regs.flag.H = false
	}
	regs.SP = uint16(result & 0x0000FFFF)
	<-c
}

func ldaR16A(regH, regL *byte) {
	//Duration: 8
	//Byte length: 1
	//Flags: Z:- N:- H:- C:-
	c := make(chan int, 1)
	go clockTicks(8, c)
	addr := bytesToUint16(*regL, *regH)
	writeAddress(addr, regs.A)
	<-c
}

func ldAaR16(regH, regL *byte) {
	//Duration: 8
	//Byte length: 1
	//Flags: Z:- N:- H:- C:-
	c := make(chan int, 1)
	go clockTicks(8, c)
	addr := bytesToUint16(*regL, *regH)
	regs.A = readAddress(addr)
	<-c
}

func ldaHLIA() {
	//Byte length: 1
	//Flags: Z:- N:- H:- C:-
	c := make(chan int, 1)
	go clockTicks(8, c)
	addr := bytesToUint16(regs.L, regs.H)
	writeAddress(addr, regs.A)
	addr++
	regs.L, regs.H = uint16ToBytes(addr)
	<-c
}

func ldaHLDA() {
	//Byte length: 1
	//Flags: Z:- N:- H:- C:-
	c := make(chan int, 1)
	go clockTicks(8, c)
	addr := bytesToUint16(regs.L, regs.H)
	writeAddress(addr, regs.A)
	addr--
	regs.L, regs.H = uint16ToBytes(addr)
	<-c
}

func ldAaHLI() {
	//Duration: 8
	//Byte length: 1
	//Flags: Z:- N:- H:- C:-
	c := make(chan int, 1)
	go clockTicks(8, c)
	addr := bytesToUint16(regs.L, regs.H)
	regs.A = readAddress(addr)
	addr++
	regs.L, regs.H = uint16ToBytes(addr)
	<-c
}

func ldAaHLD() {
	//Duration: 8
	//Byte length: 1
	//Flags: Z:- N:- H:- C:-
	c := make(chan int, 1)
	go clockTicks(8, c)
	addr := bytesToUint16(regs.L, regs.H)
	regs.A = readAddress(addr)
	addr--
	regs.L, regs.H = uint16ToBytes(addr)
	<-c
}

func lda16A() {
	//Duration: 16
	//Byte length: 3
	//Flags: Z:- N:- H:- C:-
	c := make(chan int, 1)
	go clockTicks(16, c)
	addrl := readAddress(regs.PC)
	regs.PC++
	addrh := readAddress(regs.PC)
	regs.PC++
	addr := bytesToUint16(addrl, addrh)
	writeAddress(addr, regs.A)
	<-c
}

func ldAa16() {
	//Duration: 16
	//Byte length: 3
	//Flags: Z:- N:- H:- C:-
	c := make(chan int, 1)
	go clockTicks(16, c)
	addrl := readAddress(regs.PC)
	regs.PC++
	addrh := readAddress(regs.PC)
	regs.PC++
	addr := bytesToUint16(addrl, addrh)
	regs.A = readAddress(addr)
	<-c
}

func ldha8A() {
	//Duration: 12
	//Byte length: 2
	//Flags: Z:- N:- H:- C:-
	c := make(chan int, 1)
	go clockTicks(12, c)
	offset := readAddress(regs.PC)
	regs.PC++
	addr := 0xff00 + uint16(offset)
	writeAddress(addr, regs.A)
	<-c
}

func ldhAa8() {
	//Duration: 12
	//Byte length: 2
	//Flags: Z:- N:- H:- C:-
	c := make(chan int, 1)
	go clockTicks(12, c)
	offset := readAddress(regs.PC)
	regs.PC++
	addr := 0xff00 + uint16(offset)
	regs.A = readAddress(addr)
	<-c
}

func ldhaCA() {
	//Duration: 8
	//Byte length: 1
	//Flags: Z:- N:- H:- C:-
	c := make(chan int, 1)
	go clockTicks(8, c)
	addr := 0xff00 + uint16(regs.C)
	writeAddress(addr, regs.A)
	<-c
}

func ldhAaC() {
	//Duration: 8
	//Byte length: 1
	//Flags: Z:- N:- H:- C:-
	c := make(chan int, 1)
	go clockTicks(8, c)
	addr := 0xff00 + uint16(regs.C)
	regs.A = readAddress(addr)
	<-c
}

func ldR16d16(regH, regL *byte) {
	//Duration: 12
	//Byte length: 3
	//Flags: Z:- N:- H:- C:-
	c := make(chan int, 1)
	go clockTicks(12, c)
	if regH == nil {
		d16L := readAddress(regs.PC)
		regs.PC++
		d16H := readAddress(regs.PC)
		regs.PC++
		regs.SP = bytesToUint16(d16L, d16H)
	} else {
		*regL = readAddress(regs.PC)
		regs.PC++
		*regH = readAddress(regs.PC)
		regs.PC++
	}
	<-c
}

func lda16SP() {
	//Duration: 20
	//Byte length: 3
	//Flags: Z:- N:- H:- C:-
	c := make(chan int, 1)
	go clockTicks(20, c)
	spl, sph := uint16ToBytes(regs.SP)
	addrl := readAddress(regs.PC)
	regs.PC++
	addrh := readAddress(regs.PC)
	regs.PC++
	addr := bytesToUint16(addrl, addrh)
	writeAddress(addr, spl)
	writeAddress(addr+1, sph)
	<-c
}

func ldHLSPd8() {
	//Duration: 12
	//Byte length: 2
	//Flags: Z:0 N:0 H:H C:C
	c := make(chan int, 1)
	go clockTicks(12, c)

	val := int8(readAddress(regs.PC))
	regs.PC++

	regs.flag.Z = false
	regs.flag.N = false

	res := uint32(int16(regs.SP) + int16(val))
	if ((regs.SP ^ uint16(val) ^ uint16(res&0xFFFF)) & 0x010) == 0x010 {
		regs.flag.H = true
	} else {
		regs.flag.H = false
	}
	if ((regs.SP ^ uint16(val) ^ uint16(res&0xFFFF)) & 0x100) == 0x100 {
		regs.flag.C = true
	} else {
		regs.flag.C = false
	}

	regs.L, regs.H = uint16ToBytes(uint16(0x00FFFF & res))

	<-c
}

func ldSPHL() {
	//Duration: 8
	//Byte length: 1
	//Flags: Z:- N:- H:- C:-
	c := make(chan int, 1)
	go clockTicks(8, c)
	regs.SP = bytesToUint16(regs.L, regs.H)
	<-c
}

func pushR16(regH, regL *byte) {
	//Duration: 16
	//Byte length: 1
	//Flags: Z:- N:- H:- C:-
	c := make(chan int, 1)
	go clockTicks(16, c)
	if regH == nil {
		regs.SP--
		writeAddress(regs.SP, regs.A)
		regs.SP--
		writeAddress(regs.SP, getRegF())
	} else {
		regs.SP--
		writeAddress(regs.SP, *regH)
		regs.SP--
		writeAddress(regs.SP, *regL)
	}
	<-c
}

func popR16(regH, regL *byte) {
	//Duration: 12
	//Byte length: 1
	//Flags: Z:- N:- H:- C:-
	c := make(chan int, 1)
	go clockTicks(12, c)
	if regH == nil {
		setRegF(readAddress(regs.SP))
		regs.SP++
		regs.A = readAddress(regs.SP)
		regs.SP++
	} else {
		*regL = readAddress(regs.SP)
		regs.SP++
		*regH = readAddress(regs.SP)
		regs.SP++
	}
	<-c
}

func rst(addr uint16) {
	//Duration: 16
	//Byte length: 1
	//Flags: Z:- N:- H:- C:-
	c := make(chan int, 1)
	go clockTicks(16, c)
	pcL, pcH := uint16ToBytes(regs.PC)
	regs.SP--
	writeAddress(regs.SP, pcH)
	regs.SP--
	writeAddress(regs.SP, pcL)
	regs.PC = addr
	<-c
}

func callCCa16(cc bool) {
	//Duration: 24/12
	//Byte length: 3
	//Flags: Z:- N:- H:- C:-
	c := make(chan int, 1)
	addrl := readAddress(regs.PC)
	regs.PC++
	addrh := readAddress(regs.PC)
	regs.PC++
	addr := bytesToUint16(addrl, addrh)
	if cc {
		go clockTicks(24, c)
		pcL, pcH := uint16ToBytes(regs.PC)
		regs.SP--
		writeAddress(regs.SP, pcH)
		regs.SP--
		writeAddress(regs.SP, pcL)
		regs.PC = addr
	} else {
		go clockTicks(12, c)
	}
	<-c
}

func retCC(cc bool) {
	//Duration: 20/8
	//Byte length: 1
	//Flags: Z:- N:- H:- C:-
	c := make(chan int, 1)
	if cc {
		go clockTicks(20, c)
		addrL := readAddress(regs.SP)
		regs.SP++
		addrH := readAddress(regs.SP)
		regs.SP++
		regs.PC = bytesToUint16(addrL, addrH)
	} else {
		go clockTicks(8, c)
	}
	<-c
}

func ret() {
	//Duration: 16
	//Byte length: 1
	//Flags: Z:- N:- H:- C:-
	c := make(chan int, 1)
	go clockTicks(16, c)
	addrL := readAddress(regs.SP)
	regs.SP++
	addrH := readAddress(regs.SP)
	regs.SP++
	regs.PC = bytesToUint16(addrL, addrH)
	<-c
}

func reti() {
	//Duration: 16
	//Byte length: 1
	//Flags: Z:- N:- H:- C:-
	c := make(chan int, 1)
	go clockTicks(16, c)
	addrL := readAddress(regs.SP)
	regs.SP++
	addrH := readAddress(regs.SP)
	regs.SP++
	regs.PC = bytesToUint16(addrL, addrH)
	interruptMaster = true
	<-c
}

func nop() {
	//Duration: 4
	//Byte length: 1
	//Flags: Z:- N:- H:- C:-
	c := make(chan int, 1)
	go clockTicks(4, c)
	<-c
}

func stop() {
	//Duration: 4
	//Byte length: 2
	//Flags: Z:- N:- H:- C:-
	c := make(chan int, 1)
	go clockTicks(4, c)
	<-c
}

func halt() {
	//Duration: 4
	//Byte length: 1
	//Flags: Z:- N:- H:- C:-
	c := make(chan int, 1)
	go clockTicks(4, c)
	<-c
}

func prefixCB() {
	//Duration: 4
	//Byte length: 1
	//Flags: Z:- N:- H:- C:-
	c := make(chan int, 1)
	go clockTicks(4, c)
	b := readAddress(regs.PC)
	regs.PC++
	<-c
	decodeCB(b)
}

func di() {
	//Duration: 4
	//Byte length: 1
	//Flags: Z:- N:- H:- C:-
	c := make(chan int, 1)
	go clockTicks(4, c)
	interruptMaster = false
	<-c
}

func ei() {
	//Duration: 4
	//Byte length: 1
	//Flags: Z:- N:- H:- C:-
	c := make(chan int, 1)
	go clockTicks(4, c)
	interruptMaster = true
	<-c
}
