package lib

import (
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
)

type Rabbit struct {
	Lane  int
	T     float64
	Speed float64
}

// newRabbit spawns a rabbit at the horizon with a speed that ramps up with score.
func newRabbit(score int) *Rabbit {
	return &Rabbit{
		Lane:  rand.Intn(LaneCount),
		Speed: 0.010 + rand.Float64()*0.006 + math.Min(float64(score), 60)*0.0002,
	}
}

// Position returns the on-screen pixel position along the perspective curve.
func (r *Rabbit) Position() (float32, float32) {
	endX := float64(LaneX[r.Lane])
	endY := PlayerY - 6.0
	x := HorizonX + (endX-HorizonX)*(r.T*r.T)
	y := HorizonY + (endY-HorizonY)*r.T
	hop := math.Abs(math.Sin(r.T*14)) * (2 + r.T*4)
	return float32(x), float32(y - hop)
}

// Draw renders the rabbit, growing from the horizon and cycling hop frames.
func (r *Rabbit) Draw(dst *ebiten.Image, frames [4]*ebiten.Image) {
	x, y := r.Position()
	h := float32(6 + r.T*22)
	frame := frames[int(r.T*16)%len(frames)]
	blitH(dst, frame, x, y+h/2, h)
}
