package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// drawStickFigure renders a minimal, static stick figure (round head,
// torso, arms, two legs - no walk animation) inside the given on-screen
// box (x, y, w, h). Proportions are based on h (not w), since a humanoid
// figure is much taller than it is wide; arms are allowed to reach out
// past the box width since real arms are wider than a body's collision box.
func drawStickFigure(screen *ebiten.Image, x, y, w, h float32, facingRight bool, clr color.Color) {
	cx := x + w/2
	headR := h * 0.16
	headCY := y + headR + 1
	neckY := headCY + headR + 1
	hipY := y + h*0.62
	footY := y + h
	shoulderY := neckY + (hipY-neckY)*0.25

	vector.FillCircle(screen, cx, headCY, headR, clr, true)

	noseDir := float32(1)
	if !facingRight {
		noseDir = -1
	}
	vector.StrokeLine(screen, cx, headCY, cx+noseDir*headR*1.3, headCY, 1.2, clr, true)
	vector.StrokeLine(screen, cx, neckY, cx, hipY, 1.8, clr, true)

	// arms: reach out to the sides and slightly down, well above the hip
	// so they read as arms rather than a second pair of legs
	const armReach = 0.9
	armEndY := shoulderY + h*0.16
	vector.StrokeLine(screen, cx, shoulderY, cx-w*armReach, armEndY, 1.6, clr, true)
	vector.StrokeLine(screen, cx, shoulderY, cx+w*armReach, armEndY, 1.6, clr, true)

	const stance = 0.20 // leg spacing, as a fraction of w
	vector.StrokeLine(screen, cx, hipY, cx-w*stance, footY, 1.8, clr, true)
	vector.StrokeLine(screen, cx, hipY, cx+w*stance, footY, 1.8, clr, true)
}
