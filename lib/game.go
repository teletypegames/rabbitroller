package lib

import (
	"fmt"
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
)

type State int

const (
	StateTitle State = iota
	StatePlay
	StateOver
)

type Game struct {
	assets *Assets

	state      State
	playerLane int
	score      int
	miss       int
	rabbits    []*Rabbit
	spawnTimer int
	frame      int
	flash      int

	lcd *ebiten.Image
}

func NewGame(a *Assets) *Game {
	return &Game{
		assets:     a,
		state:      StateTitle,
		playerLane: 2,
		lcd:        ebiten.NewImage(LcdW, LcdH),
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

// --- ebiten.Game ---

func (g *Game) Update() error {
	g.frame++
	switch g.state {
	case StateTitle, StateOver:
		if startPressed() {
			g.reset()
			g.state = StatePlay
		}
	case StatePlay:
		g.updatePlay()
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.DrawImage(g.assets.Bg, nil)

	switch g.state {
	case StateTitle:
		g.drawTitle(g.lcd)
	case StatePlay:
		g.drawPlay(g.lcd)
	case StateOver:
		g.drawOver(g.lcd)
	}

	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Scale(LcdScale, LcdScale)
	opts.GeoM.Translate(LcdDrawX, LcdDrawY)
	screen.DrawImage(g.lcd, opts)
}

func (g *Game) Layout(_, _ int) (int, int) {
	return ScreenW, ScreenH
}

// --- play loop ---

func (g *Game) updatePlay() {
	if pressedRepeat(ebiten.KeyLeft, 12, 6) && g.playerLane > 0 {
		g.playerLane--
	}
	if pressedRepeat(ebiten.KeyRight, 12, 6) && g.playerLane < LaneCount-1 {
		g.playerLane++
	}

	g.spawnTimer++
	interval := 70 - g.score/2
	if interval < 22 {
		interval = 22
	}
	if g.spawnTimer >= interval {
		g.spawnTimer = 0
		g.rabbits = append(g.rabbits, newRabbit(g.score))
	}

	for i := len(g.rabbits) - 1; i >= 0; i-- {
		r := g.rabbits[i]
		r.T += r.Speed
		if r.T < 1 {
			continue
		}
		if r.Lane == g.playerLane {
			g.miss++
			g.flash = 12
			if g.miss >= MaxMiss {
				g.state = StateOver
			}
		} else {
			g.score++
		}
		g.rabbits = append(g.rabbits[:i], g.rabbits[i+1:]...)
	}

	if g.flash > 0 {
		g.flash--
	}
}

// --- scene composition ---

func (g *Game) drawPlay(dst *ebiten.Image) {
	dst.Fill(ColorBg)
	drawRoad(dst)
	drawHud(dst, g.assets.HudFont, g.score, g.miss)
	// far rabbits first so closer ones overlap on top
	sort.Slice(g.rabbits, func(i, j int) bool { return g.rabbits[i].T < g.rabbits[j].T })
	for _, r := range g.rabbits {
		r.Draw(dst, g.assets.Rabbit)
	}
	drawScooter(dst, g.assets.Scooter, LaneX[g.playerLane], PlayerY, g.flash > 0, g.frame)
}

func (g *Game) drawTitle(dst *ebiten.Image) {
	dst.Fill(ColorBg)
	drawRoad(dst)
	for i := 0; i < LaneCount; i++ {
		drawScooter(dst, g.assets.Scooter, LaneX[i], PlayerY, false, g.frame)
	}
	for i := 1; i <= 3; i++ {
		demo := &Rabbit{
			Lane: (g.frame/30 + i) % LaneCount,
			T:    float64((g.frame+i*25)%90) / 90.0,
		}
		demo.Draw(dst, g.assets.Rabbit)
	}
	if (g.frame/30)%2 == 0 {
		drawTextCentered(dst, "PRESS SPACE TO START", 120, 56, g.assets.HudFont, ColorFg)
	}
	drawTextCentered(dst, "ARROWS MOVE   SPACE START", 120, 72, g.assets.HudFont, ColorFg)
}

func (g *Game) drawOver(dst *ebiten.Image) {
	g.drawPlay(dst)
	rectFill(dst, 60, 50, 120, 36, ColorBg)
	rectStroke(dst, 60, 50, 120, 36, ColorFg)
	drawTextCentered(dst, "GAME OVER", 120, 54, g.assets.HudFont, ColorFg)
	drawTextCentered(dst, fmt.Sprintf("SCORE  %03d", g.score), 120, 66, g.assets.HudFont, ColorFg)
	if (g.frame/30)%2 == 0 {
		drawTextCentered(dst, "SPACE = RETRY", 120, 76, g.assets.HudFont, ColorFg)
	}
}
