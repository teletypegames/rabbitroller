package main

import (
	"embed"
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	"rabbitroller/lib"
)

//go:embed assets/background.png assets/scooter.png assets/rabbit_f0.png assets/rabbit_f1.png assets/rabbit_f2.png assets/rabbit_f3.png
var assetsFS embed.FS

func main() {
	assets, err := lib.LoadAssets(assetsFS)
	if err != nil {
		log.Fatal(err)
	}

	ebiten.SetWindowSize(450, 336)
	ebiten.SetWindowTitle("Rabbit Scooter")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(lib.NewGame(assets)); err != nil {
		log.Fatal(err)
	}
}
