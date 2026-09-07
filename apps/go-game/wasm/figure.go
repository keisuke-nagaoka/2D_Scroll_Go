package main

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// drawCharacter draws the reference player/enemy artwork (embedded in
// assets.go), scaled to height h and mirrored horizontally when facing
// left, so the same source image works for both directions.
func drawCharacter(screen *ebiten.Image, x, y, w, h float32, facingRight, angry bool) {
	img := playerSprite
	if angry {
		img = enemySprite
	}
	b := img.Bounds()
	iw, ih := float64(b.Dx()), float64(b.Dy())
	scale := float64(h) / ih

	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(-iw/2, -ih/2) // pivot at the sprite's center
	if !facingRight {
		opts.GeoM.Scale(-1, 1)
	}
	opts.GeoM.Scale(scale, scale)
	opts.GeoM.Translate(float64(x)+float64(w)/2, float64(y)+float64(h)/2)
	screen.DrawImage(img, opts)
}
