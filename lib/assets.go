package lib

import (
	"bytes"
	"fmt"
	"image/png"
	"io/fs"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font/gofont/gobold"
)

type Assets struct {
	Bg      *ebiten.Image
	Scooter *ebiten.Image
	Rabbit  [4]*ebiten.Image
	HudFont *text.GoTextFace
}

func LoadAssets(assetsFS fs.FS) (*Assets, error) {
	src, err := text.NewGoTextFaceSource(bytes.NewReader(gobold.TTF))
	if err != nil {
		return nil, fmt.Errorf("load font: %w", err)
	}
	a := &Assets{
		HudFont: &text.GoTextFace{Source: src, Size: 8},
	}
	if a.Bg, err = loadImage(assetsFS, "assets/background.png"); err != nil {
		return nil, err
	}
	if a.Scooter, err = loadImage(assetsFS, "assets/scooter.png"); err != nil {
		return nil, err
	}
	for i := range a.Rabbit {
		name := fmt.Sprintf("assets/rabbit_f%d.png", i)
		if a.Rabbit[i], err = loadImage(assetsFS, name); err != nil {
			return nil, err
		}
	}
	return a, nil
}

func loadImage(assetsFS fs.FS, name string) (*ebiten.Image, error) {
	f, err := assetsFS.Open(name)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", name, err)
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("decode %s: %w", name, err)
	}
	return ebiten.NewImageFromImage(img), nil
}
