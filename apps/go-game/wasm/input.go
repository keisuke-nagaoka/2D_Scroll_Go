package main

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// Input unifies keyboard input (for PC) and on-screen button taps/clicks
// (for touch devices, but also usable with a mouse) into one state, so the
// rest of the game doesn't care which input method is being used.
type Input struct {
	Left, Right, Down bool
	JumpPressed       bool // edge-triggered: true only on the frame jump starts

	leftBtn, rightBtn, downBtn, jumpBtn image.Rectangle
	prevPointerJump                     bool
}

func NewInput() *Input {
	const (
		margin   = 6
		dpadSize = 24
		gap      = 4
		jumpSize = 32
		downSize = 20
	)
	return &Input{
		leftBtn:  image.Rect(margin, screenHeight-margin-dpadSize, margin+dpadSize, screenHeight-margin),
		rightBtn: image.Rect(margin+dpadSize+gap, screenHeight-margin-dpadSize, margin+dpadSize+gap+dpadSize, screenHeight-margin),
		jumpBtn:  image.Rect(screenWidth-margin-jumpSize, screenHeight-margin-jumpSize, screenWidth-margin, screenHeight-margin),
		downBtn:  image.Rect(screenWidth-margin-jumpSize-gap-downSize, screenHeight-margin-downSize, screenWidth-margin-jumpSize-gap, screenHeight-margin),
	}
}

func (in *Input) Update() {
	keyLeft := ebiten.IsKeyPressed(ebiten.KeyArrowLeft) || ebiten.IsKeyPressed(ebiten.KeyA)
	keyRight := ebiten.IsKeyPressed(ebiten.KeyArrowRight) || ebiten.IsKeyPressed(ebiten.KeyD)
	keyDown := ebiten.IsKeyPressed(ebiten.KeyArrowDown) || ebiten.IsKeyPressed(ebiten.KeyS)
	keyJump := inpututil.IsKeyJustPressed(ebiten.KeySpace) ||
		inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) ||
		inpututil.IsKeyJustPressed(ebiten.KeyW)

	var points []image.Point
	for _, id := range ebiten.TouchIDs() {
		x, y := ebiten.TouchPosition(id)
		points = append(points, image.Pt(x, y))
	}
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		points = append(points, image.Pt(x, y))
	}

	pointerLeft, pointerRight, pointerDown, pointerJump := false, false, false, false
	for _, pt := range points {
		if pt.In(in.leftBtn) {
			pointerLeft = true
		}
		if pt.In(in.rightBtn) {
			pointerRight = true
		}
		if pt.In(in.downBtn) {
			pointerDown = true
		}
		if pt.In(in.jumpBtn) {
			pointerJump = true
		}
	}
	pointerJumpEdge := pointerJump && !in.prevPointerJump
	in.prevPointerJump = pointerJump

	in.Left = keyLeft || pointerLeft
	in.Right = keyRight || pointerRight
	in.Down = keyDown || pointerDown
	in.JumpPressed = keyJump || pointerJumpEdge
}
