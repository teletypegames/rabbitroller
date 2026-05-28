package lib

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// pressedRepeat fires on press, then every `period` frames after held for `hold` frames.
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
