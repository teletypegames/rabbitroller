package lib

import "image/color"

const (
	// Logical screen matches the bezel artwork (assets/background.png) 1:1.
	ScreenW = 1200
	ScreenH = 896

	// Game world is rendered to this offscreen, then scaled into the LCD cutout.
	LcdW = 240
	LcdH = 136

	// Olive LCD screen rectangle within the bezel image (measured from the PNG).
	LcdScreenX = 331
	LcdScreenY = 203
	LcdScreenW = 537
	LcdScreenH = 399

	LaneCount = 5
	HorizonY  = 28.0
	HorizonX  = 120.0
	PlayerY   = 116.0
	MaxMiss   = 3
)

var LaneX = [LaneCount]float32{40, 80, 120, 160, 200}

// How the LcdW×LcdH world maps onto the bezel's LCD cutout (uniform fit, centered).
var (
	LcdScale = float64(LcdScreenW) / float64(LcdW)
	LcdDrawX = float64(LcdScreenX)
	LcdDrawY = float64(LcdScreenY) + (float64(LcdScreenH)-float64(LcdH)*LcdScale)/2
)

// LCD palette — ColorBg matches the bezel's printed screen olive so the
// letterboxed margins blend into the case art.
var (
	ColorBg  = color.RGBA{0xbf, 0xc2, 0xa3, 0xff}
	ColorFg  = color.RGBA{0x1a, 0x1c, 0x2c, 0xff}
	ColorDim = color.RGBA{0x6a, 0x76, 0x6e, 0xff}
)
