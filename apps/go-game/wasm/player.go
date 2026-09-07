package main

import "math"

// Player is the user-controlled character with simple tile-based physics.
type Player struct {
	x, y        float64
	vx, vy      float64
	onGround    bool
	facingRight bool

	// coyoteTimer lets a jump still register for a few frames after walking
	// off a ledge; jumpBufferTimer lets a jump pressed a few frames before
	// landing still fire the instant the player touches ground. Both exist
	// because requiring a frame-perfect press feels unresponsive.
	coyoteTimer     int
	jumpBufferTimer int
}

func (p *Player) Rect() rect {
	return rect{p.x, p.y, playerW, playerH}
}

// moveAndCollideX applies the current horizontal velocity and pushes the
// player out of any solid tile it now overlaps.
func (p *Player) moveAndCollideX(st *Stage) {
	p.x += p.vx
	top := int(math.Floor(p.y / tileSize))
	bottom := int(math.Floor((p.y + playerH - 1) / tileSize))

	if p.vx > 0 {
		right := int(math.Floor((p.x + playerW - 1) / tileSize))
		for ty := top; ty <= bottom; ty++ {
			if st.isSolid(right, ty) {
				p.x = float64(right*tileSize) - playerW
				break
			}
		}
	} else if p.vx < 0 {
		left := int(math.Floor(p.x / tileSize))
		for ty := top; ty <= bottom; ty++ {
			if st.isSolid(left, ty) {
				p.x = float64((left + 1) * tileSize)
				break
			}
		}
	}
}

// moveAndCollideY applies the current vertical velocity and resolves
// collisions against the tile grid. onHitCeiling is called with the tile
// coordinates of a block the player's head bumps into while moving upward.
func (p *Player) moveAndCollideY(st *Stage, onHitCeiling func(tx, ty int)) {
	p.y += p.vy
	left := int(math.Floor(p.x / tileSize))
	right := int(math.Floor((p.x + playerW - 1) / tileSize))

	if p.vy < 0 {
		top := int(math.Floor(p.y / tileSize))
		for tx := left; tx <= right; tx++ {
			if st.isSolid(tx, top) {
				p.y = float64((top + 1) * tileSize)
				p.vy = 0
				if onHitCeiling != nil {
					onHitCeiling(tx, top)
				}
				break
			}
		}
	} else if p.vy > 0 {
		bottom := int(math.Floor((p.y + playerH - 1) / tileSize))
		for tx := left; tx <= right; tx++ {
			if st.isSolid(tx, bottom) {
				p.y = float64(bottom*tileSize) - playerH
				p.vy = 0
				p.onGround = true
				break
			}
		}
	}
}
