package main

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/basicfont"
)

type gameState int

const (
	statePlaying gameState = iota
	stateGameOver
	stateGameClear
)

// Game implements ebiten.Game and owns all mutable state for the running
// session: the current stage, player, enemies/items, score, lives, and input.
type Game struct {
	stageSpecs  []stageSpec
	stageIdx    int
	stage       *Stage
	player      *Player
	enemies     []*Enemy
	projectiles []*Projectile
	particles   []*Particle
	coinPops    []*CoinPop
	dropItems   []*DropItem
	floatTexts  []*FloatText
	input       *Input
	cameraX     float64
	score       int
	lives       int
	state       gameState
	face        text.Face
	stageBanner int
}

func NewGame() *Game {
	g := &Game{
		stageSpecs: allStageSpecs(),
		input:      NewInput(),
		face:       text.NewGoXFace(basicfont.Face7x13),
		lives:      startingLives,
	}
	g.loadStage(0)
	return g
}

// loadStage (re)builds the given stage from scratch and places the player at
// its start position. Score and lives are intentionally left untouched so
// they persist across pipe travel and respawns.
func (g *Game) loadStage(idx int) {
	g.stageIdx = idx
	g.stage = buildStage(g.stageSpecs[idx])
	g.player = &Player{x: g.stage.playerStart.x, y: g.stage.playerStart.y, facingRight: true}
	g.enemies = nil
	for _, sp := range g.stage.enemySpawns {
		g.enemies = append(g.enemies, newEnemy(sp.x, sp.y, sp.kind))
	}
	g.projectiles = nil
	g.particles = nil
	g.coinPops = nil
	g.dropItems = nil
	g.floatTexts = nil
	g.state = statePlaying
	g.stageBanner = stageBannerFrames
}

// die consumes one life, spawns a death explosion, and shows the game-over
// prompt. Continuing respawns in the stage the player died in (see
// Update), not back at stage1.
func (g *Game) die() {
	g.lives--
	g.state = stateGameOver
	p := g.player
	g.particles = nil
	for i := 0; i < particleCount; i++ {
		g.particles = append(g.particles, newParticle(p.x+playerW/2, p.y+playerH/2, color.RGBA{255, 140, 20, 255}))
	}
}

func (g *Game) Update() error {
	g.input.Update()

	if g.stageBanner > 0 {
		g.stageBanner--
	}

	for i := len(g.particles) - 1; i >= 0; i-- {
		if !g.particles[i].Update() {
			g.particles = append(g.particles[:i], g.particles[i+1:]...)
		}
	}

	switch g.state {
	case statePlaying:
		g.updatePlaying()
	case stateGameOver:
		if g.input.JumpPressed {
			if g.lives > 0 {
				g.loadStage(g.stageIdx)
			} else {
				g.lives = startingLives
				g.score = 0
				g.loadStage(0)
			}
		}
	case stateGameClear:
		if g.input.JumpPressed {
			g.lives = startingLives
			g.score = 0
			g.loadStage(0)
		}
	}
	return nil
}

func (g *Game) updatePlaying() {
	p := g.player

	p.vx = 0
	if g.input.Left {
		p.vx = -moveSpeed
		p.facingRight = false
	}
	if g.input.Right {
		p.vx = moveSpeed
		p.facingRight = true
	}
	p.vy += gravity
	if p.vy > maxFallSpeed {
		p.vy = maxFallSpeed
	}

	if g.input.JumpPressed {
		p.jumpBufferTimer = jumpBufferFrames
	} else if p.jumpBufferTimer > 0 {
		p.jumpBufferTimer--
	}
	if p.jumpBufferTimer > 0 && p.coyoteTimer > 0 {
		p.vy = jumpVelocity
		p.onGround = false
		p.jumpBufferTimer = 0
		p.coyoteTimer = 0
	}

	p.moveAndCollideX(g.stage)
	p.onGround = false
	p.moveAndCollideY(g.stage, g.onBlockHit)

	if p.onGround {
		p.coyoteTimer = coyoteFrames
	} else if p.coyoteTimer > 0 {
		p.coyoteTimer--
	}

	if p.y > float64(heightTiles*tileSize) {
		g.die()
		return
	}

	for _, e := range g.enemies {
		if proj := e.Update(g.stage, p.x); proj != nil {
			g.projectiles = append(g.projectiles, proj)
		}
		if e.alive && e.Rect().overlaps(p.Rect()) {
			g.die()
			return
		}
	}

	for i := len(g.projectiles) - 1; i >= 0; i-- {
		pr := g.projectiles[i]
		pr.Update(g.stage)
		if !pr.alive {
			g.projectiles = append(g.projectiles[:i], g.projectiles[i+1:]...)
			continue
		}
		if pr.Rect().overlaps(p.Rect()) {
			g.die()
			return
		}
	}

	for i := len(g.coinPops) - 1; i >= 0; i-- {
		if !g.coinPops[i].Update() {
			g.coinPops = append(g.coinPops[:i], g.coinPops[i+1:]...)
		}
	}

	for i := len(g.dropItems) - 1; i >= 0; i-- {
		d := g.dropItems[i]
		d.Update(g.stage)
		if !d.alive {
			g.dropItems = append(g.dropItems[:i], g.dropItems[i+1:]...)
			continue
		}
		if !d.emerging && d.Rect().overlaps(p.Rect()) {
			switch d.kind {
			case dropOneUp:
				g.lives++
				g.floatTexts = append(g.floatTexts, newFloatText(d.x, d.y, "1UP"))
				g.dropItems = append(g.dropItems[:i], g.dropItems[i+1:]...)
			case dropPoison:
				g.die()
				return
			}
		}
	}

	for i := len(g.floatTexts) - 1; i >= 0; i-- {
		if !g.floatTexts[i].Update() {
			g.floatTexts = append(g.floatTexts[:i], g.floatTexts[i+1:]...)
		}
	}

	if g.input.Down && p.onGround {
		footTx := int(math.Floor((p.x + playerW/2) / tileSize))
		footTy := int(math.Floor((p.y + playerH) / tileSize))
		if pp := g.stage.pipeEntranceAt(footTx, footTy); pp != nil {
			if g.stageIdx == numStages-1 {
				g.state = stateGameClear
				return
			}
			g.loadStage(pp.nextStage)
			return
		}
	}

	g.cameraX = p.x + playerW/2 - screenWidth/2
	maxCam := g.stage.widthPx() - screenWidth
	if maxCam < 0 {
		maxCam = 0
	}
	if g.cameraX < 0 {
		g.cameraX = 0
	}
	if g.cameraX > maxCam {
		g.cameraX = maxCam
	}
}

// onBlockHit is called when the player's head bumps into a tile from below.
// A coin block credits the score immediately (per the "hit = get" design)
// and just shows a cosmetic pop; a 1UP or poison ("tabasco") block releases
// a moving item that must actually be touched to take effect.
func (g *Game) onBlockHit(tx, ty int) {
	switch g.stage.tileAt(tx, ty) {
	case tileBrick:
		g.stage.setTile(tx, ty, tileEmpty)
		g.spawnBrickDebris(tx, ty)
	case tileItem:
		g.stage.setTile(tx, ty, tileUsed)
		switch {
		case g.stage.isOneUp(tx, ty):
			g.dropItems = append(g.dropItems, newDropItem(tx, ty, dropOneUp))
		case g.stage.isPoison(tx, ty):
			g.dropItems = append(g.dropItems, newDropItem(tx, ty, dropPoison))
		case g.stage.isEnemyItem(tx, ty):
			g.enemies = append(g.enemies, newEnemy(float64(tx*tileSize), float64(ty*tileSize), enemyKindForStage(g.stageIdx)))
		default:
			g.score += 100
			g.coinPops = append(g.coinPops, newCoinPop(tx, ty))
		}
	}
}

// spawnBrickDebris adds a small burst of brick-colored fragments at the
// given tile, giving visual feedback that the block was destroyed.
func (g *Game) spawnBrickDebris(tx, ty int) {
	cx := float64(tx*tileSize) + tileSize/2
	cy := float64(ty*tileSize) + tileSize/2
	for i := 0; i < 8; i++ {
		g.particles = append(g.particles, newParticle(cx, cy, color.RGBA{165, 92, 52, 255}))
	}
}

// drawTile renders one tile with a bit of texture instead of a flat color:
// grass-capped dirt for the ground, a brick grid, a "?" for item blocks, and
// a pipe with a darker rim so only its topmost segment looks like an opening.
func (g *Game) drawTile(screen *ebiten.Image, tx, ty int, x, y float32) {
	switch g.stage.tileAt(tx, ty) {
	case tileHard:
		vector.FillRect(screen, x, y, tileSize, tileSize, color.RGBA{101, 67, 33, 255}, false)
		if g.stage.tileAt(tx, ty-1) == tileEmpty {
			vector.FillRect(screen, x, y, tileSize, 4, color.RGBA{86, 172, 60, 255}, false)
		}
	case tileBrick:
		vector.FillRect(screen, x, y, tileSize, tileSize, color.RGBA{165, 92, 52, 255}, false)
		lineClr := color.RGBA{110, 60, 32, 255}
		vector.StrokeLine(screen, x, y+tileSize/2, x+tileSize, y+tileSize/2, 1, lineClr, false)
		vector.StrokeLine(screen, x+tileSize/2, y, x+tileSize/2, y+tileSize/2, 1, lineClr, false)
		vector.StrokeLine(screen, x+tileSize/4, y+tileSize/2, x+tileSize/4, y+tileSize, 1, lineClr, false)
		vector.StrokeLine(screen, x+tileSize*3/4, y+tileSize/2, x+tileSize*3/4, y+tileSize, 1, lineClr, false)
	case tileItem:
		vector.FillRect(screen, x, y, tileSize, tileSize, color.RGBA{230, 180, 40, 255}, false)
		vector.StrokeRect(screen, x+1, y+1, tileSize-2, tileSize-2, 1, color.RGBA{150, 110, 20, 255}, false)
		opts := &text.DrawOptions{}
		opts.GeoM.Translate(float64(x)+5, float64(y)+2)
		text.Draw(screen, "?", g.face, opts)
	case tileUsed:
		vector.FillRect(screen, x, y, tileSize, tileSize, color.RGBA{110, 100, 80, 255}, false)
		vector.StrokeRect(screen, x+1, y+1, tileSize-2, tileSize-2, 1, color.RGBA{80, 74, 60, 255}, false)
	case tilePipe:
		vector.FillRect(screen, x, y, tileSize, tileSize, color.RGBA{34, 139, 60, 255}, false)
		vector.FillRect(screen, x+2, y, 2, tileSize, color.RGBA{90, 200, 110, 255}, false)
		if g.stage.tileAt(tx, ty-1) != tilePipe {
			vector.FillRect(screen, x-1, y, tileSize+2, 4, color.RGBA{20, 100, 44, 255}, false)
		}
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{135, 206, 235, 255})

	startCol := int(g.cameraX) / tileSize
	endCol := startCol + screenWidth/tileSize + 2
	if endCol > g.stage.widthTiles {
		endCol = g.stage.widthTiles
	}
	for ty := 0; ty < heightTiles; ty++ {
		for tx := startCol; tx < endCol; tx++ {
			if g.stage.tileAt(tx, ty) == tileEmpty {
				continue
			}
			x := float32(tx*tileSize) - float32(g.cameraX)
			y := float32(ty * tileSize)
			g.drawTile(screen, tx, ty, x, y)
		}
	}

	for _, c := range g.coinPops {
		x := float32(c.x) - float32(g.cameraX) + coinSize/2
		y := float32(c.y) + coinSize/2
		vector.FillCircle(screen, x, y, coinSize/2, color.RGBA{255, 215, 0, 255}, true)
		vector.StrokeCircle(screen, x, y, coinSize/2, 1, color.RGBA{200, 150, 0, 255}, true)
		vector.FillCircle(screen, x, y, coinSize*0.28, color.RGBA{255, 240, 160, 255}, true)
	}

	for _, d := range g.dropItems {
		x := float32(d.x) - float32(g.cameraX)
		y := float32(d.y)
		drawDropItem(screen, x, y, d.kind)
	}

	for _, f := range g.floatTexts {
		x := float64(f.x) - g.cameraX
		opts := &text.DrawOptions{}
		opts.ColorScale.ScaleAlpha(1 - float32(f.age)/40)
		opts.ColorScale.ScaleWithColor(color.RGBA{40, 200, 90, 255})
		opts.GeoM.Translate(x, f.y)
		text.Draw(screen, f.text, g.face, opts)
	}

	for _, e := range g.enemies {
		if !e.alive {
			continue
		}
		x := float32(e.x) - float32(g.cameraX)
		y := float32(e.y)
		drawCharacter(screen, x, y, enemyW, enemyH, e.vx > 0, true)
	}

	for _, pr := range g.projectiles {
		x := float32(pr.x) - float32(g.cameraX) + projectileW/2
		y := float32(pr.y) + projectileH/2
		vector.FillCircle(screen, x, y, projectileW/2, color.RGBA{230, 40, 30, 255}, true)
		vector.StrokeCircle(screen, x, y, projectileW/2, 1, color.RGBA{120, 10, 10, 255}, true)
	}

	p := g.player
	if g.state != stateGameOver {
		px := float32(p.x) - float32(g.cameraX)
		py := float32(p.y)
		drawCharacter(screen, px, py, playerW, playerH, p.facingRight, false)
	}

	g.drawHUD(screen)
	g.drawTouchControls(screen)
	g.drawStageBanner(screen)

	if g.state == stateGameOver {
		g.drawGameOver(screen)
	}
	if g.state == stateGameClear {
		g.drawGameClear(screen)
	}

	// draw explosion/debris particles above the overlay so they stay visible
	for _, pt := range g.particles {
		x := float32(pt.x) - float32(g.cameraX)
		y := float32(pt.y)
		alpha := float32(1 - float32(pt.age)/float32(pt.life))
		r, gg, b, _ := pt.clr.RGBA()
		clr := color.RGBA{uint8(r >> 8), uint8(gg >> 8), uint8(b >> 8), uint8(255 * alpha)}
		vector.FillCircle(screen, x, y, 2, clr, true)
	}
}

// drawDropItem renders a block-dropped item: the 1UP as a small energy-
// drink can (body, metallic cap, stripe, lightning bolt), the poison as a
// red "tabasco" bottle (body, dark cap, label band).
func drawDropItem(screen *ebiten.Image, x, y float32, kind dropItemKind) {
	switch kind {
	case dropOneUp:
		vector.FillRect(screen, x+1, y+3, dropItemSize-2, dropItemSize-3, color.RGBA{20, 130, 120, 255}, false)
		vector.StrokeRect(screen, x+1, y+3, dropItemSize-2, dropItemSize-3, 1, color.RGBA{10, 80, 74, 255}, false)
		vector.FillRect(screen, x+2, y, dropItemSize-4, 4, color.RGBA{210, 212, 220, 255}, false)
		vector.FillRect(screen, x+1, y+7, dropItemSize-2, 2, color.RGBA{250, 180, 30, 255}, false)

		boltClr := color.RGBA{255, 255, 255, 255}
		mx := x + dropItemSize*0.5
		vector.StrokeLine(screen, mx+2, y+4, mx-2, y+9, 1.2, boltClr, true)
		vector.StrokeLine(screen, mx-2, y+9, mx+2, y+11, 1.2, boltClr, true)
	case dropPoison:
		vector.FillRect(screen, x+2, y+3, dropItemSize-4, dropItemSize-3, color.RGBA{190, 20, 20, 255}, false)
		vector.StrokeRect(screen, x+2, y+3, dropItemSize-4, dropItemSize-3, 1, color.RGBA{110, 10, 10, 255}, false)
		vector.FillRect(screen, x+3, y, dropItemSize-6, 4, color.RGBA{40, 40, 40, 255}, false)
		vector.FillRect(screen, x+2, y+7, dropItemSize-4, 2, color.RGBA{250, 250, 240, 255}, false)
	}
}

func (g *Game) drawStageBanner(screen *ebiten.Image) {
	if g.stageBanner <= 0 {
		return
	}
	msg := fmt.Sprintf("STAGE %d/%d", g.stageIdx+1, len(g.stageSpecs))
	opts := &text.DrawOptions{}
	opts.GeoM.Scale(2, 2)
	opts.GeoM.Translate(screenWidth/2-float64(len(msg))*7, 40)
	text.Draw(screen, msg, g.face, opts)
}

func (g *Game) drawHUD(screen *ebiten.Image) {
	msg := fmt.Sprintf("SCORE %04d  LIVES %d  STAGE %d/%d", g.score, g.lives, g.stageIdx+1, len(g.stageSpecs))
	opts := &text.DrawOptions{}
	opts.GeoM.Translate(8, 8)
	text.Draw(screen, msg, g.face, opts)
}

// drawArrowButton draws a small square button with a triangular arrow icon
// pointing in the given direction ("left", "right", or "down").
func drawArrowButton(screen *ebiten.Image, r image.Rectangle, dir string) {
	x, y, w, h := float32(r.Min.X), float32(r.Min.Y), float32(r.Dx()), float32(r.Dy())
	vector.FillRect(screen, x, y, w, h, color.RGBA{20, 20, 20, 130}, true)
	vector.StrokeRect(screen, x, y, w, h, 1.5, color.RGBA{255, 255, 255, 210}, true)

	cx, cy := x+w/2, y+h/2
	s := w * 0.28
	var path vector.Path
	switch dir {
	case "left":
		path.MoveTo(cx+s*0.6, cy-s)
		path.LineTo(cx-s*0.8, cy)
		path.LineTo(cx+s*0.6, cy+s)
	case "right":
		path.MoveTo(cx-s*0.6, cy-s)
		path.LineTo(cx+s*0.8, cy)
		path.LineTo(cx-s*0.6, cy+s)
	case "down":
		path.MoveTo(cx-s, cy-s*0.6)
		path.LineTo(cx, cy+s*0.8)
		path.LineTo(cx+s, cy-s*0.6)
	}
	path.Close()
	vector.FillPath(screen, &path, &vector.FillOptions{}, &vector.DrawPathOptions{AntiAlias: true})
}

// drawJumpButton draws a larger circular button with an upward-pointing
// triangle, visually distinct from the directional buttons.
func drawJumpButton(screen *ebiten.Image, r image.Rectangle) {
	cx := float32(r.Min.X) + float32(r.Dx())/2
	cy := float32(r.Min.Y) + float32(r.Dy())/2
	radius := float32(r.Dx()) / 2

	vector.FillCircle(screen, cx, cy, radius, color.RGBA{170, 30, 30, 160}, true)
	vector.StrokeCircle(screen, cx, cy, radius, 1.5, color.RGBA{255, 255, 255, 210}, true)

	s := radius * 0.5
	var path vector.Path
	path.MoveTo(cx, cy-s)
	path.LineTo(cx-s*0.9, cy+s*0.6)
	path.LineTo(cx+s*0.9, cy+s*0.6)
	path.Close()
	vector.FillPath(screen, &path, &vector.FillOptions{}, &vector.DrawPathOptions{AntiAlias: true})
}

func (g *Game) drawTouchControls(screen *ebiten.Image) {
	drawArrowButton(screen, g.input.leftBtn, "left")
	drawArrowButton(screen, g.input.rightBtn, "right")
	drawArrowButton(screen, g.input.downBtn, "down")
	drawJumpButton(screen, g.input.jumpBtn)
}

// drawGameOver shows one of two screens after a death: while lives remain,
// a Mario-style "lives remaining" interstitial (character icon + "x N"); once
// lives run out, a full GAME OVER screen that restarts the whole run.
func (g *Game) drawGameOver(screen *ebiten.Image) {
	vector.FillRect(screen, 0, 0, screenWidth, screenHeight, color.RGBA{10, 10, 20, 235}, false)

	if g.lives > 0 {
		drawCharacter(screen, screenWidth/2-playerW*1.5/2, screenHeight/2-40, playerW*1.5, playerH*1.5, true, false)

		opts := &text.DrawOptions{}
		opts.GeoM.Scale(2.5, 2.5)
		opts.GeoM.Translate(screenWidth/2+14, screenHeight/2-26)
		text.Draw(screen, fmt.Sprintf("x%d", g.lives), g.face, opts)

		hint := "PRESS JUMP TO CONTINUE"
		opts2 := &text.DrawOptions{}
		opts2.GeoM.Translate(screenWidth/2-float64(len(hint))*3.5, screenHeight-36)
		text.Draw(screen, hint, g.face, opts2)
		return
	}

	opts := &text.DrawOptions{}
	opts.GeoM.Scale(2, 2)
	opts.GeoM.Translate(screenWidth/2-72, screenHeight/2-24)
	text.Draw(screen, "GAME OVER", g.face, opts)

	hint := "PRESS JUMP TO RESTART"
	opts2 := &text.DrawOptions{}
	opts2.GeoM.Translate(screenWidth/2-float64(len(hint))*3.5, screenHeight/2+20)
	text.Draw(screen, hint, g.face, opts2)
}

// drawGameClear shows a congratulations screen after the final stage's pipe
// is entered, with the score reached, then restarts a fresh run on input.
func (g *Game) drawGameClear(screen *ebiten.Image) {
	vector.FillRect(screen, 0, 0, screenWidth, screenHeight, color.RGBA{20, 20, 60, 235}, false)

	title := "GAME CLEAR!"
	opts := &text.DrawOptions{}
	opts.GeoM.Scale(2, 2)
	opts.GeoM.Translate(screenWidth/2-float64(len(title))*7, screenHeight/2-30)
	text.Draw(screen, title, g.face, opts)

	score := fmt.Sprintf("FINAL SCORE %04d", g.score)
	opts2 := &text.DrawOptions{}
	opts2.GeoM.Translate(screenWidth/2-float64(len(score))*3.5, screenHeight/2+6)
	text.Draw(screen, score, g.face, opts2)

	hint := "PRESS JUMP TO PLAY AGAIN"
	opts3 := &text.DrawOptions{}
	opts3.GeoM.Translate(screenWidth/2-float64(len(hint))*3.5, screenHeight/2+24)
	text.Draw(screen, hint, g.face, opts3)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}
