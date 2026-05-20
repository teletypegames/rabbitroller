package main

import (
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
	"golang.org/x/image/font/basicfont"
)

const (
	screenW = 240
	screenH = 136
	scale   = 4

	nLanes   = 5
	horizonY = 28.0
	horizonX = 120.0
	playerY  = 116.0
	maxMiss  = 3
)

var laneX = [nLanes]float32{40, 80, 120, 160, 200}

// Sweetie-16 palette
var (
	lcdBg  = color.RGBA{0xf4, 0xf4, 0xf4, 0xff} // #12
	lcdFg  = color.RGBA{0x1a, 0x1c, 0x2c, 0xff} // #0
	lcdDim = color.RGBA{0x94, 0xb0, 0xc2, 0xff} // #13
)

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
	font       *text.GoXFace
}

func newGame() *Game {
	return &Game{
		state:      stateTitle,
		playerLane: 2,
		font:       text.NewGoXFace(basicfont.Face7x13),
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

func drawText(dst *ebiten.Image, s string, x, y float64, face *text.GoXFace, c color.Color) {
	opts := &text.DrawOptions{}
	opts.GeoM.Translate(x, y)
	opts.ColorScale.ScaleWithColor(c)
	text.Draw(dst, s, face, opts)
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
	// left tree
	rectFill(dst, 14, 44, 3, 10, lcdFg)
	circFill(dst, 15, 42, 5, lcdFg)
	// right field hint
	for i := 0; i < 5; i++ {
		fi := float32(i)
		line(dst, 225, 50+fi*4, 235, 46+fi*4, lcdDim)
	}
}

func (g *Game) drawHud(dst *ebiten.Image) {
	drawText(dst, "SCORE", 4, 2, g.font, lcdFg)
	drawText(dst, fmt.Sprintf("%03d", g.score), 40, 2, g.font, lcdFg)
	drawText(dst, "MISS", 170, 2, g.font, lcdFg)
	for i := 0; i < g.miss; i++ {
		drawText(dst, "X", float64(206+i*8), 2, g.font, lcdFg)
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
	drawText(dst, "RABBIT  ROLLER", 60, 0, g.font, lcdFg)
	drawText(dst, "GAME & WATCH", 2, 0, g.font, lcdFg)
	drawText(dst, "WIDE SCREEN", 158, 0, g.font, lcdFg)
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
		drawText(dst, "PRESS Z TO START", 60, 58, g.font, lcdFg)
	}
	drawText(dst, "ARROWS MOVE  Z START", 48, 72, g.font, lcdFg)
}

func (g *Game) drawOver(dst *ebiten.Image) {
	g.drawPlay(dst)
	rectFill(dst, 60, 50, 120, 36, lcdBg)
	rectStroke(dst, 60, 50, 120, 36, lcdFg)
	drawText(dst, "GAME  OVER", 80, 54, g.font, lcdFg)
	drawText(dst, fmt.Sprintf("SCORE  %03d", g.score), 76, 66, g.font, lcdFg)
	if (g.frame/30)%2 == 0 {
		drawText(dst, "Z TO RESTART", 74, 76, g.font, lcdFg)
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
	switch g.state {
	case stateTitle:
		g.drawTitle(screen)
	case statePlay:
		g.drawPlay(screen)
	case stateOver:
		g.drawOver(screen)
	}
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
