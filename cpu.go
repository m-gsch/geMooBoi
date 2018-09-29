package main

import (
	"encoding/binary"
)

const (
	clockSpeed = 4194304
)

type flags struct {
	Z  bool
	NZ bool //Shit workaround
	N  bool
	H  bool
	C  bool
	NC bool //Shit workaround
}

type registers struct {
	A    byte
	flag flags
	B    byte
	C    byte
	D    byte
	E    byte
	H    byte
	L    byte
	PC   uint16
	SP   uint16
}

var regs registers

func uint16ToBytes(value uint16) (byte, byte) {
	b := make([]byte, 2)
	binary.LittleEndian.PutUint16(b, value)
	return b[0], b[1]
}

func bytesToUint16(b0, b1 byte) uint16 {
	b := make([]byte, 2)
	b[0] = b0
	b[1] = b1
	value := binary.LittleEndian.Uint16(b)
	return value
}

func clockTicks(n int, c chan int) {
	cyclesPassed += n

	c <- 0
}

func getRegF() (b byte) {
	if regs.flag.Z {
		b |= 0x80
	}
	if regs.flag.N {
		b |= 0x40
	}
	if regs.flag.H {
		b |= 0x20
	}
	if regs.flag.C {
		b |= 0x10
	}
	return
}

func setRegF(b byte) {

	if b&0x80 == 0x80 {
		regs.flag.Z = true
	} else {
		regs.flag.Z = false
	}

	if b&0x40 == 0x40 {
		regs.flag.N = true
	} else {
		regs.flag.N = false
	}

	if b&0x20 == 0x20 {
		regs.flag.H = true
	} else {
		regs.flag.H = false
	}

	if b&0x10 == 0x10 {
		regs.flag.C = true
	} else {
		regs.flag.C = false
	}
}
