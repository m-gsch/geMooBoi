package main

func decodeCB(opcode byte) {
	reg := opcode & 0x7
	nBit := opcode >> 3 & 0x7
	ins := opcode >> 3

	switch ins {
	case 0x0:
		rlc(tableR[reg])
	case 0x1:
		rrc(tableR[reg])
	case 0x2:
		rl(tableR[reg])
	case 0x3:
		rr(tableR[reg])
	case 0x4:
		sla(tableR[reg])
	case 0x5:
		sra(tableR[reg])
	case 0x6:
		swap(tableR[reg])
	case 0x7:
		srl(tableR[reg])
	case 0x8, 0x9, 0xA, 0xB, 0xC, 0xD, 0xE, 0xF:
		bit(tableR[reg], nBit)
	case 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17:
		res(tableR[reg], nBit)
	case 0x18, 0x19, 0x1A, 0x1B, 0x1C, 0x1D, 0x1E, 0x1F:
		set(tableR[reg], nBit)
	}
}

func rlc(reg *byte) {
	c := make(chan int, 1)
	var val byte
	var addr uint16
	if reg == nil {
		go clockTicks(16, c)
		addr = bytesToUint16(regs.L, regs.H)
		val = readAddress(addr)
	} else {
		go clockTicks(8, c)
		val = *reg
	}
	regs.flag.N = false
	regs.flag.H = false
	if val>>7 != 0x00 {
		regs.flag.C = true
	} else {
		regs.flag.C = false
	}
	rot := val<<1 | val>>7
	if rot == 0x00 {
		regs.flag.Z = true
	} else {
		regs.flag.Z = false
	}
	if reg == nil {
		writeAddress(addr, rot)
	} else {
		*reg = rot
	}
	<-c
}

func rrc(reg *byte) {
	c := make(chan int, 1)
	var val byte
	var addr uint16
	if reg == nil {
		go clockTicks(16, c)
		addr = bytesToUint16(regs.L, regs.H)
		val = readAddress(addr)
	} else {
		go clockTicks(8, c)
		val = *reg
	}
	regs.flag.N = false
	regs.flag.H = false
	if val&0x1 != 0x00 {
		regs.flag.C = true
	} else {
		regs.flag.C = false
	}
	rot := val>>1 | (val&0x1)<<7
	if rot == 0x00 {
		regs.flag.Z = true
	} else {
		regs.flag.Z = false
	}
	if reg == nil {
		writeAddress(addr, rot)
	} else {
		*reg = rot
	}
	<-c
}

func rl(reg *byte) {
	c := make(chan int, 1)
	var val byte
	var addr uint16
	var b byte = 0x0
	if reg == nil {
		go clockTicks(16, c)
		addr = bytesToUint16(regs.L, regs.H)
		val = readAddress(addr)
	} else {
		go clockTicks(8, c)
		val = *reg
	}
	regs.flag.N = false
	regs.flag.H = false
	if regs.flag.C {
		b = 0x1
	}
	if val>>7 != 0x00 {
		regs.flag.C = true
	} else {
		regs.flag.C = false
	}
	rot := val<<1 | b
	if rot == 0x00 {
		regs.flag.Z = true
	} else {
		regs.flag.Z = false
	}
	if reg == nil {
		writeAddress(addr, rot)
	} else {
		*reg = rot
	}
	<-c
}

func rr(reg *byte) {
	c := make(chan int, 1)
	var val byte
	var addr uint16
	var b byte = 0x0
	if reg == nil {
		go clockTicks(16, c)
		addr = bytesToUint16(regs.L, regs.H)
		val = readAddress(addr)
	} else {
		go clockTicks(8, c)
		val = *reg
	}
	regs.flag.N = false
	regs.flag.H = false
	if regs.flag.C {
		b = 0x80
	}
	if val&0x1 != 0x00 {
		regs.flag.C = true
	} else {
		regs.flag.C = false
	}
	rot := val>>1 | b
	if rot == 0x00 {
		regs.flag.Z = true
	} else {
		regs.flag.Z = false
	}
	if reg == nil {
		writeAddress(addr, rot)
	} else {
		*reg = rot
	}
	<-c
}

func sla(reg *byte) {
	c := make(chan int, 1)
	var val byte
	var addr uint16
	if reg == nil {
		go clockTicks(16, c)
		addr = bytesToUint16(regs.L, regs.H)
		val = readAddress(addr)
	} else {
		go clockTicks(8, c)
		val = *reg
	}
	regs.flag.N = false
	regs.flag.H = false
	if val>>7 != 0x00 {
		regs.flag.C = true
	} else {
		regs.flag.C = false
	}
	shift := val << 1
	if shift == 0x00 {
		regs.flag.Z = true
	} else {
		regs.flag.Z = false
	}
	if reg == nil {
		writeAddress(addr, shift)
	} else {
		*reg = shift
	}
	<-c
}

func sra(reg *byte) {
	c := make(chan int, 1)
	var val byte
	var addr uint16
	if reg == nil {
		go clockTicks(16, c)
		addr = bytesToUint16(regs.L, regs.H)
		val = readAddress(addr)
	} else {
		go clockTicks(8, c)
		val = *reg
	}
	regs.flag.N = false
	regs.flag.H = false
	if val&0x1 != 0x00 {
		regs.flag.C = true
	} else {
		regs.flag.C = false
	}
	shift := val>>1 | (val & 0x80)
	if shift == 0x00 {
		regs.flag.Z = true
	} else {
		regs.flag.Z = false
	}
	if reg == nil {
		writeAddress(addr, shift)
	} else {
		*reg = shift
	}
	<-c
}

func swap(reg *byte) {
	c := make(chan int, 1)
	var val byte
	var addr uint16
	if reg == nil {
		go clockTicks(16, c)
		addr = bytesToUint16(regs.L, regs.H)
		val = readAddress(addr)
	} else {
		go clockTicks(8, c)
		val = *reg
	}
	regs.flag.N = false
	regs.flag.H = false
	regs.flag.C = false
	swap := val>>4 | val<<4
	if swap == 0x00 {
		regs.flag.Z = true
	} else {
		regs.flag.Z = false
	}
	if reg == nil {
		writeAddress(addr, swap)
	} else {
		*reg = swap
	}
	<-c
}

func srl(reg *byte) {
	c := make(chan int, 1)
	var val byte
	var addr uint16
	if reg == nil {
		go clockTicks(16, c)
		addr = bytesToUint16(regs.L, regs.H)
		val = readAddress(addr)
	} else {
		go clockTicks(8, c)
		val = *reg
	}
	regs.flag.N = false
	regs.flag.H = false
	if val&0x1 != 0x00 {
		regs.flag.C = true
	} else {
		regs.flag.C = false
	}
	shift := val >> 1
	if shift == 0x00 {
		regs.flag.Z = true
	} else {
		regs.flag.Z = false
	}
	if reg == nil {
		writeAddress(addr, shift)
	} else {
		*reg = shift
	}
	<-c
}

func bit(reg *byte, nBit byte) {
	c := make(chan int, 1)
	var val byte
	var addr uint16
	if reg == nil {
		go clockTicks(16, c)
		addr = bytesToUint16(regs.L, regs.H)
		val = readAddress(addr)
	} else {
		go clockTicks(8, c)
		val = *reg
	}
	b := val >> nBit & 0x1
	if b == 0x0 {
		regs.flag.Z = true
	} else {
		regs.flag.Z = false
	}
	regs.flag.N = false
	regs.flag.H = true
	<-c
}

func res(reg *byte, nBit byte) {
	c := make(chan int, 1)
	var val byte
	var addr uint16
	if reg == nil {
		go clockTicks(16, c)
		addr = bytesToUint16(regs.L, regs.H)
		val = readAddress(addr)
	} else {
		go clockTicks(8, c)
		val = *reg
	}
	var b byte = 0x1 << nBit
	val &^= b
	if reg == nil {
		writeAddress(addr, val)
	} else {
		*reg = val
	}
	<-c
}

func set(reg *byte, nBit byte) {
	c := make(chan int, 1)
	var val byte
	var addr uint16
	if reg == nil {
		go clockTicks(16, c)
		addr = bytesToUint16(regs.L, regs.H)
		val = readAddress(addr)
	} else {
		go clockTicks(8, c)
		val = *reg
	}
	var b byte = 0x1 << nBit
	val |= b
	if reg == nil {
		writeAddress(addr, val)
	} else {
		*reg = val
	}
	<-c
}
