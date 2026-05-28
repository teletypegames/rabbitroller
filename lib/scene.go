package lib

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

func drawRoad(dst *ebiten.Image) {
	// trapezoidal road outline
	line(dst, HorizonX-20, HorizonY, 0, 136, ColorFg)
	line(dst, HorizonX+20, HorizonY, 239, 136, ColorFg)
	// shoulder hash marks (perspective)
	for i := 1; i <= 6; i++ {
		t := float32(i) / 7.0
		y := float32(HorizonY) + (136-float32(HorizonY))*t
		lx := (float32(HorizonX) - 20) + (0-(float32(HorizonX)-20))*t
		rx := (float32(HorizonX) + 20) + (239-(float32(HorizonX)+20))*t
		px(dst, lx-1, y, ColorDim)
		px(dst, rx+1, y, ColorDim)
	}
	drawTree(dst, 14, 70)
}

func drawTree(dst *ebiten.Image, x, y float32) {
	rectFill(dst, x+3, y, 3, 10, ColorFg)
	circFill(dst, x+4, y-3, 5, ColorFg)
	circFill(dst, x+8, y-1, 4, ColorFg)
	circFill(dst, x, y-1, 4, ColorFg)
	circFill(dst, x+5, y-7, 3, ColorFg)
}

// drawScooter blinks the sprite when `blink` is true, using the global frame counter.
func drawScooter(dst, sprite *ebiten.Image, x, y float32, blink bool, frame int) {
	if blink && (frame/4)%2 == 0 {
		return
	}
	blitH(dst, sprite, x, y+8, 24)
}

func drawHud(dst *ebiten.Image, font text.Face, score, miss int) {
	drawText(dst, "SCORE", 4, 2, font, ColorFg)
	drawText(dst, fmt.Sprintf("%03d", score), 40, 2, font, ColorFg)
	missSlotRight := float64(LcdW - 4)
	missSlotX := missSlotRight - float64(MaxMiss*8)
	drawTextRight(dst, "MISS", missSlotX-2, 2, font, ColorFg)
	for i := 0; i < miss; i++ {
		drawText(dst, "X", missSlotX+float64(i*8), 2, font, ColorFg)
	}
}
