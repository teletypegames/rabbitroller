package main

import (
	"bytes"
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand"
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/gofont/gobold"
)

const (
	screenW = 320
	screenH = 240
	scale   = 4

	lcdW    = 240
	lcdH    = 136
	lcdOffX = 40
	lcdOffY = 60

	nLanes   = 5
	horizonY = 28.0
	horizonX = 120.0
	playerY  = 116.0
	maxMiss  = 3
)

var laneX = [nLanes]float32{40, 80, 120, 160, 200}

// LCD palette (sweetie-16 inspired)
var (
	lcdBg  = color.RGBA{0xc8, 0xcc, 0xb8, 0xff} // pale olive — classic G&W LCD
	lcdFg  = color.RGBA{0x1a, 0x1c, 0x2c, 0xff}
	lcdDim = color.RGBA{0x6a, 0x76, 0x6e, 0xff}
)

// Bezel palette
var (
	bezelRed   = color.RGBA{0x8a, 0x1c, 0x1c, 0xff}
	bezelTan   = color.RGBA{0xd8, 0xcf, 0xb6, 0xff}
	bezelDark  = color.RGBA{0x3a, 0x2e, 0x24, 0xff}
	wheatGold  = color.RGBA{0xb8, 0x9a, 0x5f, 0xff}
	grassGreen = color.RGBA{0x55, 0x95, 0x42, 0xff}
)

var bezelFontSource *text.GoTextFaceSource

func init() {
	s, err := text.NewGoTextFaceSource(bytes.NewReader(gobold.TTF))
	if err != nil {
		log.Fatal(err)
	}
	bezelFontSource = s
}

type stateKind int

const (
	stateTitle stateKind = iota
	statePlay
	stateOver
)

type rabbit struct {
	lane  int
	t     float64
	speed float64
}

type Game struct {
	state      stateKind
	playerLane int
	score      int
	miss       int
	rabbits    []*rabbit
	spawnTimer int
	frame      int
	flash      int

	lcdImg *ebiten.Image

	hudFont   *text.GoTextFace
	titleFont *text.GoTextFace
	smallFont *text.GoTextFace
}

func newGame() *Game {
	return &Game{
		state:      stateTitle,
		playerLane: 2,
		lcdImg:     ebiten.NewImage(lcdW, lcdH),
		hudFont:    &text.GoTextFace{Source: bezelFontSource, Size: 8},
		titleFont:  &text.GoTextFace{Source: bezelFontSource, Size: 14},
		smallFont:  &text.GoTextFace{Source: bezelFontSource, Size: 8},
	}
}

func (g *Game) reset() {
	g.playerLane = 2
	g.score = 0
	g.miss = 0
	g.rabbits = g.rabbits[:0]
	g.spawnTimer = 0
	g.flash = 0
}

// btnp(key, hold, period): fires on press, then every `period` frames after held for `hold` frames.
func pressedRepeat(key ebiten.Key, hold, period int) bool {
	if inpututil.IsKeyJustPressed(key) {
		return true
	}
	d := inpututil.KeyPressDuration(key)
	if d > hold && (d-hold)%period == 0 {
		return true
	}
	return false
}

func startPressed() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyZ) ||
		inpututil.IsKeyJustPressed(ebiten.KeySpace) ||
		inpututil.IsKeyJustPressed(ebiten.KeyEnter)
}

func (g *Game) spawnRabbit() {
	r := &rabbit{
		lane:  rand.Intn(nLanes),
		t:     0,
		speed: 0.010 + rand.Float64()*0.006 + math.Min(float64(g.score), 60)*0.0002,
	}
	g.rabbits = append(g.rabbits, r)
}

func rabbitPos(r *rabbit) (float32, float32) {
	endX := float64(laneX[r.lane])
	endY := playerY - 6.0
	x := horizonX + (endX-horizonX)*(r.t*r.t)
	y := horizonY + (endY-horizonY)*r.t
	hop := math.Abs(math.Sin(r.t*14)) * (2 + r.t*4)
	return float32(x), float32(y - hop)
}

func (g *Game) updatePlay() {
	if pressedRepeat(ebiten.KeyLeft, 12, 6) && g.playerLane > 0 {
		g.playerLane--
	}
	if pressedRepeat(ebiten.KeyRight, 12, 6) && g.playerLane < nLanes-1 {
		g.playerLane++
	}

	g.spawnTimer++
	interval := 70 - g.score/2
	if interval < 22 {
		interval = 22
	}
	if g.spawnTimer >= interval {
		g.spawnTimer = 0
		g.spawnRabbit()
	}

	for i := len(g.rabbits) - 1; i >= 0; i-- {
		r := g.rabbits[i]
		r.t += r.speed
		if r.t >= 1 {
			if r.lane == g.playerLane {
				g.miss++
				g.flash = 12
				if g.miss >= maxMiss {
					g.state = stateOver
				}
			} else {
				g.score++
			}
			g.rabbits = append(g.rabbits[:i], g.rabbits[i+1:]...)
		}
	}

	if g.flash > 0 {
		g.flash--
	}
}

// --- drawing primitives wrapping ebiten/vector ---

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

func circStroke(dst *ebiten.Image, cx, cy, r float32, c color.Color) {
	vector.StrokeCircle(dst, cx, cy, r, 1, c, false)
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

// --- scene drawing ---

func (g *Game) drawRoad(dst *ebiten.Image) {
	// trapezoidal road outline
	line(dst, horizonX-20, horizonY, 0, 136, lcdFg)
	line(dst, horizonX+20, horizonY, 239, 136, lcdFg)
	// shoulder hash marks (perspective)
	for i := 1; i <= 6; i++ {
		t := float32(i) / 7.0
		y := float32(horizonY) + (136-float32(horizonY))*t
		lx := (float32(horizonX) - 20) + (0-(float32(horizonX)-20))*t
		rx := (float32(horizonX) + 20) + (239-(float32(horizonX)+20))*t
		px(dst, lx-1, y, lcdDim)
		px(dst, rx+1, y, lcdDim)
	}
	// LCD tree silhouette on the left shoulder
	drawLcdTree(dst, 14, 70)
}

func drawLcdTree(dst *ebiten.Image, x, y float32) {
	// trunk
	rectFill(dst, x+3, y, 3, 10, lcdFg)
	// blobby canopy
	circFill(dst, x+4, y-3, 5, lcdFg)
	circFill(dst, x+8, y-1, 4, lcdFg)
	circFill(dst, x, y-1, 4, lcdFg)
	circFill(dst, x+5, y-7, 3, lcdFg)
}

func (g *Game) drawHud(dst *ebiten.Image) {
	drawText(dst, "PONT", 4, 2, g.hudFont, lcdFg)
	drawText(dst, fmt.Sprintf("%03d", g.score), 32, 2, g.hudFont, lcdFg)
	// Reserve fixed slot for miss markers on the right
	missSlotRight := float64(lcdW - 4)
	missSlotX := missSlotRight - float64(maxMiss*8)
	drawTextRight(dst, "HIBA", missSlotX-2, 2, g.hudFont, lcdFg)
	for i := 0; i < g.miss; i++ {
		drawText(dst, "X", missSlotX+float64(i*8), 2, g.hudFont, lcdFg)
	}
}

func drawRabbit(dst *ebiten.Image, x, y float32, t float64) {
	s := float32(1 + int(t*3))
	// body
	rectFill(dst, x-s, y-s, s*2, s*2, lcdFg)
	// head
	rectFill(dst, x+s-1, y-s-1, 2, 2, lcdFg)
	// ears
	px(dst, x+s, y-s-2, lcdFg)
	px(dst, x+s-2, y-s-2, lcdFg)
	// feet stretch on hop
	stretch := float32(int(math.Abs(math.Sin(t*14)) * 2))
	px(dst, x-s, y+s, lcdFg)
	px(dst, x-s-stretch, y+s, lcdFg)
}

func (g *Game) drawScooter(dst *ebiten.Image, x, y float32, blink bool) {
	if blink && (g.frame/4)%2 == 0 {
		return
	}
	// wheels
	circStroke(dst, x-6, y+4, 3, lcdFg)
	circStroke(dst, x+6, y+4, 3, lcdFg)
	px(dst, x-6, y+4, lcdFg)
	px(dst, x+6, y+4, lcdFg)
	// deck
	rectFill(dst, x-8, y, 16, 3, lcdFg)
	// seat post + seat
	line(dst, x+4, y, x+4, y-5, lcdFg)
	rectFill(dst, x+2, y-6, 5, 2, lcdFg)
	// handlebar column
	line(dst, x-4, y, x-5, y-8, lcdFg)
	// handlebars
	line(dst, x-8, y-8, x-2, y-8, lcdFg)
	// rider head
	circStroke(dst, x-5, y-12, 2, lcdFg)
}

func (g *Game) drawPlay(dst *ebiten.Image) {
	dst.Fill(lcdBg)
	g.drawRoad(dst)
	g.drawHud(dst)
	// draw far rabbits first
	sort.Slice(g.rabbits, func(i, j int) bool { return g.rabbits[i].t < g.rabbits[j].t })
	for _, r := range g.rabbits {
		x, y := rabbitPos(r)
		drawRabbit(dst, x, y, r.t)
	}
	g.drawScooter(dst, laneX[g.playerLane], playerY, g.flash > 0)
}

func (g *Game) drawTitle(dst *ebiten.Image) {
	dst.Fill(lcdBg)
	g.drawRoad(dst)
	// preview scooters
	for i := 0; i < nLanes; i++ {
		g.drawScooter(dst, laneX[i], playerY, false)
	}
	// demo rabbits
	for i := 1; i <= 3; i++ {
		fake := &rabbit{
			lane: (g.frame/30 + i) % nLanes,
			t:    float64((g.frame+i*25)%90) / 90.0,
		}
		x, y := rabbitPos(fake)
		drawRabbit(dst, x, y, fake.t)
	}
	if (g.frame/30)%2 == 0 {
		drawTextCentered(dst, "NYOMJ Z-T A KEZDÉSHEZ", 120, 56, g.hudFont, lcdFg)
	}
	drawTextCentered(dst, "NYILAK MOZGÁS   Z START", 120, 72, g.hudFont, lcdFg)
}

func (g *Game) drawOver(dst *ebiten.Image) {
	g.drawPlay(dst)
	rectFill(dst, 60, 50, 120, 36, lcdBg)
	rectStroke(dst, 60, 50, 120, 36, lcdFg)
	drawTextCentered(dst, "VÉGE", 120, 54, g.hudFont, lcdFg)
	drawTextCentered(dst, fmt.Sprintf("PONT  %03d", g.score), 120, 66, g.hudFont, lcdFg)
	if (g.frame/30)%2 == 0 {
		drawTextCentered(dst, "Z = ÚJRA", 120, 76, g.hudFont, lcdFg)
	}
}

// --- bezel (Game & Watch case art) ---

func (g *Game) drawBezel(dst *ebiten.Image) {
	// Outer red frame
	dst.Fill(bezelRed)
	// Tan inner panel
	rectFill(dst, 4, 4, screenW-8, screenH-8, bezelTan)

	// "JÁTSZ & NÉZZ" boxed logo (top-left)
	logoX, logoY, logoW, logoH := float32(8), float32(6), float32(56), float32(36)
	rectStroke(dst, logoX, logoY, logoW, logoH, bezelDark)
	rectStroke(dst, logoX+1, logoY+1, logoW-2, logoH-2, bezelDark)
	drawTextCentered(dst, "JÁTSZ", float64(logoX+logoW/2), float64(logoY+4), g.smallFont, bezelDark)
	drawTextCentered(dst, "&", float64(logoX+logoW/2), float64(logoY+15), g.smallFont, bezelDark)
	drawTextCentered(dst, "NÉZZ", float64(logoX+logoW/2), float64(logoY+25), g.smallFont, bezelDark)

	// Center title "NYÚLÓS RUGÓS"
	drawTextCentered(dst, "NYÚLÓS RUGÓS", float64(screenW)/2, 18, g.titleFont, bezelDark)

	// Wheat field on right bezel (printed art)
	g.drawWheatField(dst, 286, 78)

	// LCD inner frame (dark groove around the active screen)
	rectFill(dst, lcdOffX-4, lcdOffY-4, lcdW+8, lcdH+8, bezelDark)
	rectFill(dst, lcdOffX-2, lcdOffY-2, lcdW+4, lcdH+4, color.RGBA{0x15, 0x15, 0x18, 0xff})
}

func (g *Game) drawWheatField(dst *ebiten.Image, vx, vy float32) {
	// diagonal hatching emanating from a vanishing point — wheat rows
	for i := 0; i < 14; i++ {
		angle := math.Pi*0.55 + float64(i)*math.Pi*0.025
		ex := vx + float32(math.Cos(angle)*46)
		ey := vy + float32(math.Sin(angle)*46)
		line(dst, vx, vy, ex, ey, wheatGold)
	}
	// grass tufts along the road edge
	for i := 0; i < 7; i++ {
		gy := vy + 6 + float32(i)*8
		gx := vx - 6 - float32(i)
		px(dst, gx, gy, grassGreen)
		px(dst, gx+1, gy-1, grassGreen)
		px(dst, gx-1, gy-1, grassGreen)
		px(dst, gx, gy-2, grassGreen)
	}
}

// --- ebiten Game interface ---

func (g *Game) Update() error {
	g.frame++
	switch g.state {
	case stateTitle:
		if startPressed() {
			g.reset()
			g.state = statePlay
		}
	case statePlay:
		g.updatePlay()
	case stateOver:
		if startPressed() {
			g.reset()
			g.state = statePlay
		}
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.drawBezel(screen)

	switch g.state {
	case stateTitle:
		g.drawTitle(g.lcdImg)
	case statePlay:
		g.drawPlay(g.lcdImg)
	case stateOver:
		g.drawOver(g.lcdImg)
	}

	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(lcdOffX, lcdOffY)
	screen.DrawImage(g.lcdImg, opts)
}

func (g *Game) Layout(_, _ int) (int, int) {
	return screenW, screenH
}

func main() {
	ebiten.SetWindowSize(screenW*scale, screenH*scale)
	ebiten.SetWindowTitle("Rabbit Roller")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	if err := ebiten.RunGame(newGame()); err != nil {
		log.Fatal(err)
	}
}
