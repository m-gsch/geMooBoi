package main

func twoBPP(tile []byte) (decoded [8][8]byte) {

	x := 0
	y := 0
	for i := 0; i < 16; i += 2 {
		for j := 7; j >= 0; j-- {
			lowBit := tile[i] >> uint(j) & 0x1
			highBit := tile[i+1] >> uint(j) << 1 & 0x2
			color := lowBit | highBit
			decoded[y][x] = color
			x++
		}
		x = 0
		y++
	}
	return
}
