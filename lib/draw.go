package lib

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

func px(dst *ebiten.Image, x, y float32, c color.Color) {
	vector.DrawFilledRect(dst, x, y, 1, 1, c, false)
}

func line(dst *ebiten.Image, x1, y1, x2, y2 float32, c color.Color) {
	vector.StrokeLine(dst, x1, y1, x2, y2, 1, c, false)
}

func rectFill(dst *ebiten.Image, x, y, w, h float32, c color.Color) {
	vector.DrawFilledRect(dst, x, y, w, h, c, false)
}

func rectStroke(dst *ebiten.Image, x, y, w, h float32, c color.Color) {
	vector.StrokeRect(dst, x, y, w, h, 1, c, false)
}

func circFill(dst *ebiten.Image, cx, cy, r float32, c color.Color) {
	vector.DrawFilledCircle(dst, cx, cy, r, c, false)
}

// blitH draws img scaled to targetH pixels tall, anchored so (cx,cy) is the
// horizontal center / vertical bottom of the sprite.
func blitH(dst, img *ebiten.Image, cx, cy, targetH float32) {
	b := img.Bounds()
	s := float64(targetH) / float64(b.Dy())
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(s, s)
	op.GeoM.Translate(float64(cx)-float64(b.Dx())*s/2, float64(cy)-float64(b.Dy())*s)
	dst.DrawImage(img, op)
}

func drawText(dst *ebiten.Image, s string, x, y float64, face text.Face, c color.Color) {
	opts := &text.DrawOptions{}
	opts.GeoM.Translate(x, y)
	opts.ColorScale.ScaleWithColor(c)
	text.Draw(dst, s, face, opts)
}

func drawTextCentered(dst *ebiten.Image, s string, cx, y float64, face text.Face, c color.Color) {
	w, _ := text.Measure(s, face, 0)
	drawText(dst, s, cx-w/2, y, face, c)
}

func drawTextRight(dst *ebiten.Image, s string, rx, y float64, face text.Face, c color.Color) {
	w, _ := text.Measure(s, face, 0)
	drawText(dst, s, rx-w, y, face, c)
}
