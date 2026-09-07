package main

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// figureVisualBox enlarges an entity's small collision box into a taller,
// slightly wider box for drawing a stick figure. Gameplay (jump height, gap
// clearance, etc.) keeps using the original tight collision box; only the
// drawn figure is bigger, anchored to the same feet position, so the head/
// torso/legs have enough room to read as a person instead of a blob.
func figureVisualBox(x, y, w, h float32) (vx, vy, vw, vh float32) {
	vw = w * 1.4
	vh = h * 1.8
	vx = x - (vw-w)/2
	vy = y - (vh - h)
	return
}

// drawStickFigure renders a minimal stick figure (small round head, short
// torso, long legs - no arms) inside the given on-screen box (x, y, w, h).
// Proportions are based on h (not w), since a humanoid figure is much
// taller than it is wide; the caller should pass a box taller than the
// entity's actual collision box so the head/torso/legs have room to read
// clearly instead of collapsing into a blob. Standing still or airborne
// draws a fixed, un-animated pose; only walking on the ground swings the
// legs through phase (advanced by the caller each frame).
func drawStickFigure(screen *ebiten.Image, x, y, w, h float32, facingRight, moving, onGround bool, phase float64, clr color.Color) {
	cx := x + w/2
	headR := h * 0.15
	headCY := y + headR + 1
	neckY := headCY + headR + 1
	hipY := y + h*0.55
	footY := y + h

	vector.FillCircle(screen, cx, headCY, headR, clr, true)

	noseDir := float32(1)
	if !facingRight {
		noseDir = -1
	}
	vector.StrokeLine(screen, cx, headCY, cx+noseDir*headR*1.3, headCY, 1.2, clr, true)
	vector.StrokeLine(screen, cx, neckY, cx, hipY, 1.8, clr, true)

	const stance = 0.20 // idle leg spacing, as a fraction of w
	var swingL, swingR float32
	switch {
	case !onGround:
		swingL, swingR = w*0.32, w*0.32 // jump pose: legs spread apart
	case moving:
		s := float32(math.Sin(phase)) * w * 0.32
		swingL, swingR = s, -s // alternating stride
	}

	vector.StrokeLine(screen, cx, hipY, cx-w*stance+swingL, footY, 1.8, clr, true)
	vector.StrokeLine(screen, cx, hipY, cx+w*stance+swingR, footY, 1.8, clr, true)
}
