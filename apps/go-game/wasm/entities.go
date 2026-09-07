package main

import "math"

// Enemy is a simple patrol-and-fall enemy: it walks back and forth within a
// fixed range around its spawn point (reversing at that range's edges or at
// a wall), and falls under gravity, so it can walk off ledges into pits.
// Touching the player ends the game (no stomp mechanic).
type Enemy struct {
	x, y, vx, vy float64
	minX, maxX   float64
	alive        bool
	grounded     bool
	walkPhase    float64
}

// patrolRadius keeps enemies from wandering across a whole stage once
// nothing but the ground blocks them (e.g. floating item/brick blocks no
// longer sit low enough to act as walls).
const patrolRadius = 3 * tileSize

func newEnemy(x, y float64) *Enemy {
	return &Enemy{
		x: x, y: y,
		vx:    -enemySpeed,
		minX:  x - patrolRadius,
		maxX:  x + patrolRadius,
		alive: true,
	}
}

func (e *Enemy) Rect() rect {
	return rect{e.x, e.y, enemyW, enemyH}
}

func (e *Enemy) Update(st *Stage) {
	if !e.alive {
		return
	}

	e.grounded = false
	e.walkPhase += 0.3

	e.vy += gravity
	if e.vy > maxFallSpeed {
		e.vy = maxFallSpeed
	}

	nextX := e.x + e.vx
	var checkTx int
	if e.vx < 0 {
		checkTx = int(math.Floor(nextX / tileSize))
	} else {
		checkTx = int(math.Floor((nextX + enemyW - 1) / tileSize))
	}
	top := int(math.Floor(e.y / tileSize))
	bottom := int(math.Floor((e.y + enemyH - 1) / tileSize))
	blocked := nextX < e.minX || nextX > e.maxX
	if !blocked {
		for ty := top; ty <= bottom; ty++ {
			if st.isSolid(checkTx, ty) {
				blocked = true
				break
			}
		}
	}
	if blocked {
		e.vx = -e.vx
	} else {
		e.x = nextX
	}

	e.y += e.vy
	if e.vy >= 0 {
		left := int(math.Floor(e.x / tileSize))
		right := int(math.Floor((e.x + enemyW - 1) / tileSize))
		by := int(math.Floor((e.y + enemyH - 1) / tileSize))
		for tx := left; tx <= right; tx++ {
			if st.isSolid(tx, by) {
				e.y = float64(by*tileSize) - enemyH
				e.vy = 0
				e.grounded = true
				break
			}
		}
	}

	if e.y > float64(heightTiles*tileSize) {
		e.alive = false
	}
}

// CoinPop is a purely cosmetic effect shown when a coin block is hit; the
// score is credited the instant it appears, so it never needs to be
// touched. It pops up one tile, lingers briefly, then disappears.
type CoinPop struct {
	x, y  float64
	baseY float64
	vy    float64
	age   int
}

func newCoinPop(tileX, tileY int) *CoinPop {
	spawnY := float64(tileY * tileSize)
	return &CoinPop{
		x:     float64(tileX*tileSize) + (tileSize-coinSize)/2,
		y:     spawnY,
		baseY: spawnY - tileSize,
		vy:    -2.2,
	}
}

// Update reports whether the effect is still alive (should keep being drawn).
func (c *CoinPop) Update() bool {
	if c.y > c.baseY {
		c.y += c.vy
		if c.y < c.baseY {
			c.y = c.baseY
		}
		return true
	}
	c.age++
	return c.age < 20
}

// OneUp is a 1UP power-up: it emerges from its block, then walks along the
// ground/platforms like the enemy (bouncing off walls, falling under
// gravity, and off ledges into pits) until the player touches it.
type OneUp struct {
	x, y, vx, vy float64
	emergeTarget float64
	emerging     bool
	alive        bool
}

func newOneUp(tileX, tileY int) *OneUp {
	spawnY := float64(tileY * tileSize)
	return &OneUp{
		x:            float64(tileX*tileSize) + (tileSize-oneUpSize)/2,
		y:            spawnY,
		vx:           oneUpSpeed,
		emergeTarget: spawnY - tileSize,
		emerging:     true,
		alive:        true,
	}
}

func (o *OneUp) Rect() rect {
	return rect{o.x, o.y, oneUpSize, oneUpSize}
}

func (o *OneUp) Update(st *Stage) {
	if !o.alive {
		return
	}
	if o.emerging {
		o.y -= 1
		if o.y <= o.emergeTarget {
			o.y = o.emergeTarget
			o.emerging = false
		}
		return
	}

	o.vy += gravity
	if o.vy > maxFallSpeed {
		o.vy = maxFallSpeed
	}

	nextX := o.x + o.vx
	var checkTx int
	if o.vx < 0 {
		checkTx = int(math.Floor(nextX / tileSize))
	} else {
		checkTx = int(math.Floor((nextX + oneUpSize - 1) / tileSize))
	}
	top := int(math.Floor(o.y / tileSize))
	bottom := int(math.Floor((o.y + oneUpSize - 1) / tileSize))
	blocked := false
	for ty := top; ty <= bottom; ty++ {
		if st.isSolid(checkTx, ty) {
			blocked = true
			break
		}
	}
	if blocked {
		o.vx = -o.vx
	} else {
		o.x = nextX
	}

	o.y += o.vy
	if o.vy >= 0 {
		left := int(math.Floor(o.x / tileSize))
		right := int(math.Floor((o.x + oneUpSize - 1) / tileSize))
		by := int(math.Floor((o.y + oneUpSize - 1) / tileSize))
		for tx := left; tx <= right; tx++ {
			if st.isSolid(tx, by) {
				o.y = float64(by*tileSize) - oneUpSize
				o.vy = 0
				break
			}
		}
	}

	if o.y > float64(heightTiles*tileSize) {
		o.alive = false
	}
}

// FloatText is a short-lived cosmetic message (e.g. "1UP") that drifts
// upward and disappears, like the pickup feedback in classic platformers.
type FloatText struct {
	x, y float64
	text string
	age  int
}

func newFloatText(x, y float64, s string) *FloatText {
	return &FloatText{x: x, y: y, text: s}
}

func (f *FloatText) Update() bool {
	f.y -= 0.6
	f.age++
	return f.age < 40
}
