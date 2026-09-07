package main

import (
	"bytes"
	_ "embed"
	"image"
	_ "image/png"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed assets/player.png
var playerPNGBytes []byte

//go:embed assets/enemy.png
var enemyPNGBytes []byte

var playerSprite = mustDecodeSprite(playerPNGBytes)
var enemySprite = mustDecodeSprite(enemyPNGBytes)

func mustDecodeSprite(data []byte) *ebiten.Image {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		panic(err)
	}
	return ebiten.NewImageFromImage(img)
}
