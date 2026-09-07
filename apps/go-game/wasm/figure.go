package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// drawCharacter renders a chunky-limbed toon figure (round head, thick
// tube-like arms/legs ending in round hands/feet) inside the given
// on-screen box (x, y, w, h), static (no walk animation). angry selects
// the enemy palette (red body, slit angry eyes); otherwise it's the player
// palette (white body, red cap, blue feet, dot eyes).
func drawCharacter(screen *ebiten.Image, x, y, w, h float32, facingRight, angry bool) {
	cx := x + w/2
	headR := h * 0.26
	headCY := y + headR + 1
	hipY := y + h*0.60
	footY := y + h
	shoulderY := headCY + headR*0.85

	bodyClr := color.RGBA{240, 240, 240, 255}
	handClr := color.RGBA{255, 255, 255, 255}
	footClr := color.RGBA{40, 110, 220, 255}
	outline := color.RGBA{25, 25, 25, 255}
	if angry {
		bodyClr = color.RGBA{215, 30, 30, 255}
		handClr = color.RGBA{190, 20, 20, 255}
		footClr = color.RGBA{150, 15, 15, 255}
	}

	const limbW = 2.0
	const handR = 1.8
	const footR = 2.0

	dir := float32(1)
	if !facingRight {
		dir = -1
	}
	lx, rx := cx-dir*w*0.22, cx+dir*w*0.42 // back/front foot, leaning into facing direction
	feetY := footY - footR*0.5
	vector.StrokeLine(screen, cx, hipY, lx, feetY, limbW, bodyClr, true)
	vector.StrokeLine(screen, cx, hipY, rx, feetY, limbW, bodyClr, true)
	vector.FillCircle(screen, lx, feetY, footR, footClr, true)
	vector.FillCircle(screen, rx, feetY, footR, footClr, true)

	vector.StrokeLine(screen, cx, shoulderY, cx, hipY, limbW+0.6, bodyClr, true)

	handY := shoulderY + h*0.22
	ax := w * 0.58
	vector.StrokeLine(screen, cx, shoulderY, cx-ax, handY, limbW, bodyClr, true)
	vector.StrokeLine(screen, cx, shoulderY, cx+ax, handY, limbW, bodyClr, true)
	vector.FillCircle(screen, cx-ax, handY, handR, handClr, true)
	vector.FillCircle(screen, cx+ax, handY, handR, handClr, true)

	vector.FillCircle(screen, cx, headCY, headR, bodyClr, true)
	vector.StrokeCircle(screen, cx, headCY, headR, 1, outline, true)

	if !angry {
		vector.FillRect(screen, cx-headR*0.9, headCY-headR*1.1, headR*1.8, headR*0.55, color.RGBA{210, 30, 30, 255}, true)
	}

	eyeOff := headR * 0.42
	if angry {
		vector.StrokeLine(screen, cx-eyeOff-1.2, headCY-1.2, cx-eyeOff+1.2, headCY+1.2, 1.1, outline, true)
		vector.StrokeLine(screen, cx+eyeOff-1.2, headCY+1.2, cx+eyeOff+1.2, headCY-1.2, 1.1, outline, true)
	} else {
		vector.FillCircle(screen, cx-eyeOff, headCY, 1, outline, true)
		vector.FillCircle(screen, cx+eyeOff, headCY, 1, outline, true)
	}
}
