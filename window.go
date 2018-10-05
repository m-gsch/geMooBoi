package main

import (
	"log"

	"github.com/hajimehoshi/ebiten"
)

const (
	initScreenWidth  = 160
	initScreenHeight = 144
	initScreenScale  = 2
)

var newPixels = make([]byte, 4*initScreenHeight*initScreenWidth)
var pixels = make([]byte, 4*initScreenHeight*initScreenWidth)
var xTile = 0
var yTile = 0

// update is called every frame (1/60 [s]).
func update(screen *ebiten.Image) error {

	// Write your game's logical update.
	updateState()
	if ebiten.IsDrawingSkipped() {
		// When the game is running slowly, the rendering result
		// will not be adopted.
		return nil
	}

	// Write your game's rendering.
	screen.ReplacePixels(pixels)

	return nil
}

func showWindow() {

	if err := ebiten.Run(update, initScreenWidth, initScreenHeight, initScreenScale, "geMooBoi"); err != nil {
		log.Fatal(err)
	}
}

func drawTile(tile [8][8]byte) {

	for i, line := range tile {
		h := i*4*160 + 160*4*8*yTile
		for j, pixel := range line {
			pos := j*4 + 8*4*xTile
			switch pixel {
			case 0:
				pixels[pos+h] = 0xff
				pixels[pos+1+h] = 0xff
				pixels[pos+2+h] = 0xff
				pixels[pos+3+h] = 0xff
			case 1:
				pixels[pos+h] = 0xff
				pixels[pos+1+h] = 0xaa
				pixels[pos+2+h] = 0xaa
				pixels[pos+3+h] = 0xaa
			case 2:
				pixels[pos+h] = 0xff
				pixels[pos+1+h] = 0x55
				pixels[pos+2+h] = 0x55
				pixels[pos+3+h] = 0x55
			case 3:
				pixels[pos+h] = 0xff
				pixels[pos+1+h] = 0x00
				pixels[pos+2+h] = 0x00
				pixels[pos+3+h] = 0x00

			}
		}
	}
	xTile++
	if xTile == 20 {
		yTile++
		xTile = 0
	}
}
