package main

import (
	"bytes"
	"embed"
	"fmt"
	"image/color"
	"image/png"
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
	// Logical screen matches the bezel artwork (assets/background.png) 1:1.
	screenW = 1200
	screenH = 896

	// Game world is rendered to this offscreen, then scaled into the LCD cutout.
	lcdW = 240
	lcdH = 136

	// Olive LCD screen rectangle within the bezel image (measured from the PNG).
	lcdScreenX = 331
	lcdScreenY = 203
	lcdScreenW = 537
	lcdScreenH = 399

	nLanes   = 5
	horizonY = 28.0
	horizonX = 120.0
	playerY  = 116.0
	maxMiss  = 3
)

var laneX = [nLanes]float32{40, 80, 120, 160, 200}

// How the lcdW×lcdH world maps onto the bezel's LCD cutout (uniform fit, centered).
var (
	lcdScale = float64(lcdScreenW) / float64(lcdW)
	lcdDrawX = float64(lcdScreenX)
	lcdDrawY = float64(lcdScreenY) + (float64(lcdScreenH)-float64(lcdH)*lcdScale)/2
)

// LCD palette — lcdBg matches the bezel's printed screen olive so the
// letterboxed margins blend into the case art.
var (
	lcdBg  = color.RGBA{0xbf, 0xc2, 0xa3, 0xff} // bezel screen olive
	lcdFg  = color.RGBA{0x1a, 0x1c, 0x2c, 0xff}
	lcdDim = color.RGBA{0x6a, 0x76, 0x6e, 0xff}
)

//go:embed assets/background.png assets/scooter.png assets/rabbit_f0.png assets/rabbit_f1.png assets/rabbit_f2.png assets/rabbit_f3.png
var assetsFS embed.FS

var (
	bgImg      *ebiten.Image
	scooterImg *ebiten.Image
	rabbitImgs [4]*ebiten.Image
)

var bezelFontSource *text.GoTextFaceSource

func loadImg(name string) *ebiten.Image {
	f, err := assetsFS.Open(name)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		log.Fatal(err)
	}
	return ebiten.NewImageFromImage(img)
}

func init() {
	s, err := text.NewGoTextFaceSource(bytes.NewReader(gobold.TTF))
	if err != nil {
		log.Fatal(err)
	}
	bezelFontSource = s

	bgImg = loadImg("assets/background.png")
	scooterImg = loadImg("assets/scooter.png")
	for i := range rabbitImgs {
		rabbitImgs[i] = loadImg(fmt.Sprintf("assets/rabbit_f%d.png", i))
	}
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

	hudFont *text.GoTextFace
}

func newGame() *Game {
	return &Game{
		state:      stateTitle,
		playerLane: 2,
		lcdImg:     ebiten.NewImage(lcdW, lcdH),
		hudFont:    &text.GoTextFace{Source: bezelFontSource, Size: 8},
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

// blitH draws img scaled to targetH pixels tall, anchored so (cx,cy) is the
// horizontal center / vertical bottom of the sprite. Nearest filtering keeps
// the LCD pixels crisp.
func blitH(dst, img *ebiten.Image, cx, cy, targetH float32) {
	b := img.Bounds()
	s := float64(targetH) / float64(b.Dy())
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(s, s)
	op.GeoM.Translate(float64(cx)-float64(b.Dx())*s/2, float64(cy)-float64(b.Dy())*s)
	dst.DrawImage(img, op)
}

func drawRabbit(dst *ebiten.Image, x, y float32, t float64) {
	// Grows from far (small) to near (large); the hop cycle picks one of 4 frames.
	h := float32(6 + t*22)
	frame := rabbitImgs[int(t*16)%len(rabbitImgs)]
	blitH(dst, frame, x, y+h/2, h)
}

func (g *Game) drawScooter(dst *ebiten.Image, x, y float32, blink bool) {
	if blink && (g.frame/4)%2 == 0 {
		return
	}
	blitH(dst, scooterImg, x, y+8, 24)
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
	// Bezel artwork fills the whole logical screen 1:1.
	screen.DrawImage(bgImg, nil)

	switch g.state {
	case stateTitle:
		g.drawTitle(g.lcdImg)
	case statePlay:
		g.drawPlay(g.lcdImg)
	case stateOver:
		g.drawOver(g.lcdImg)
	}

	// Scale the game world into the bezel's LCD cutout.
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Scale(lcdScale, lcdScale)
	opts.GeoM.Translate(lcdDrawX, lcdDrawY)
	screen.DrawImage(g.lcdImg, opts)
}

func (g *Game) Layout(_, _ int) (int, int) {
	return screenW, screenH
}

func main() {
	ebiten.SetWindowSize(900, 672)
	ebiten.SetWindowTitle("Rabbit Roller")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	if err := ebiten.RunGame(newGame()); err != nil {
		log.Fatal(err)
	}
}
