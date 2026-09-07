package main

import (
	"math"
	"math/rand"
)

// enemyKind selects an enemy's behavior; stages introduce these one at a
// time so the pattern variety grows as the player progresses.
type enemyKind int

const (
	enemyWalker  enemyKind = iota // stage 1-3: patrol left/right
	enemyJumper                   // stage 4-6: patrol + periodic hop
	enemyDropper                  // stage 7-9: falls from above, then patrols
	enemyThrower                  // stage 10: patrol + periodically throws a projectile
)

// enemyKindForStage returns the enemy behavior introduced by the given
// stage (0-based index), matching the requested progression.
func enemyKindForStage(idx int) enemyKind {
	switch {
	case idx <= 2:
		return enemyWalker
	case idx <= 5:
		return enemyJumper
	case idx <= 8:
		return enemyDropper
	default:
		return enemyThrower
	}
}

// Enemy is a patrol-and-fall enemy: it walks back and forth within a fixed
// range around its spawn point (reversing at that range's edges or at a
// wall), and falls under gravity, so it can walk off ledges into pits.
// Touching the player ends the game (no stomp mechanic). Depending on
// kind, it may also hop (enemyJumper), start by free-falling from above
// (enemyDropper), or periodically launch a projectile (enemyThrower).
type Enemy struct {
	kind         enemyKind
	x, y, vx, vy float64
	minX, maxX   float64
	alive        bool
	grounded     bool
	hasLanded    bool
	jumpTimer    int
	throwTimer   int
}

// patrolRadius keeps enemies from wandering across a whole stage once
// nothing but the ground blocks them (e.g. floating item/brick blocks no
// longer sit low enough to act as walls).
const patrolRadius = 3 * tileSize

func newEnemy(x, y float64, kind enemyKind) *Enemy {
	e := &Enemy{
		kind:  kind,
		x:     x,
		y:     y,
		vx:    -enemySpeed,
		minX:  x - patrolRadius,
		maxX:  x + patrolRadius,
		alive: true,

		jumpTimer:  enemyJumpInterval,
		throwTimer: enemyThrowInterval,
	}
	if kind == enemyDropper {
		e.vx = 0 // stays put horizontally until it lands
	}
	return e
}

func (e *Enemy) Rect() rect {
	return rect{e.x, e.y, enemyW, enemyH}
}

// Update advances the enemy by one tick and returns a newly-launched
// projectile, if this tick's behavior fired one (nil otherwise).
func (e *Enemy) Update(st *Stage) *Projectile {
	if !e.alive {
		return nil
	}

	e.grounded = false

	e.vy += gravity
	if e.vy > maxFallSpeed {
		e.vy = maxFallSpeed
	}

	// droppers don't move horizontally until they've hit the ground once
	if e.kind != enemyDropper || e.hasLanded {
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
		return nil
	}

	var spawned *Projectile
	switch e.kind {
	case enemyDropper:
		if e.grounded && !e.hasLanded {
			e.hasLanded = true
			e.vx = -enemySpeed
		}
	case enemyJumper:
		if e.grounded {
			e.jumpTimer--
			if e.jumpTimer <= 0 {
				e.vy = enemyJumpVelocity
				e.grounded = false
				e.jumpTimer = enemyJumpInterval
			}
		}
	case enemyThrower:
		e.throwTimer--
		if e.throwTimer <= 0 {
			e.throwTimer = enemyThrowInterval
			dir := 1.0
			if e.vx < 0 {
				dir = -1
			}
			spawned = newProjectile(e.x+enemyW/2, e.y+enemyH/2-projectileH/2, dir*projectileSpeed)
		}
	}
	return spawned
}

// Projectile is a hazard thrown by an enemyThrower: it flies in a straight
// line until it hits a wall or leaves the stage, and ends the game on
// contact with the player.
type Projectile struct {
	x, y, vx float64
	alive    bool
}

func newProjectile(x, y, vx float64) *Projectile {
	return &Projectile{x: x, y: y, vx: vx, alive: true}
}

func (pr *Projectile) Rect() rect {
	return rect{pr.x, pr.y, projectileW, projectileH}
}

func (pr *Projectile) Update(st *Stage) {
	if !pr.alive {
		return
	}
	nextX := pr.x + pr.vx
	var tx int
	if pr.vx < 0 {
		tx = int(math.Floor(nextX / tileSize))
	} else {
		tx = int(math.Floor((nextX + projectileW - 1) / tileSize))
	}
	ty := int(math.Floor(pr.y / tileSize))
	if st.isSolid(tx, ty) {
		pr.alive = false
		return
	}
	pr.x = nextX
	if pr.x < -tileSize*4 || pr.x > st.widthPx()+tileSize*4 {
		pr.alive = false
	}
}

// Particle is a short-lived cosmetic speck used for the death explosion
// effect; it has no gameplay effect.
type Particle struct {
	x, y, vx, vy float64
	age, life    int
}

func newParticle(x, y float64) *Particle {
	angle := rand.Float64() * 2 * math.Pi
	speed := 1.2 + rand.Float64()*2.4
	return &Particle{
		x: x, y: y,
		vx:   math.Cos(angle) * speed,
		vy:   math.Sin(angle)*speed - 1.5,
		life: 25 + rand.Intn(20),
	}
}

// Update reports whether the particle is still alive (should keep being drawn).
func (pt *Particle) Update() bool {
	pt.vy += gravity * 0.4
	pt.x += pt.vx
	pt.y += pt.vy
	pt.age++
	return pt.age < pt.life
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

// dropItemKind distinguishes what a block-dropped, walking item does when
// the player touches it.
type dropItemKind int

const (
	dropOneUp  dropItemKind = iota // extra life
	dropPoison                     // "tabasco" bottle: touching it ends the game
)

// DropItem is a walking power-up/hazard released from a block: it emerges
// from the block, then walks along the ground/platforms like the enemy
// (bouncing off walls, falling under gravity, and off ledges into pits)
// until the player touches it.
type DropItem struct {
	kind         dropItemKind
	x, y, vx, vy float64
	emergeTarget float64
	emerging     bool
	alive        bool
}

func newDropItem(tileX, tileY int, kind dropItemKind) *DropItem {
	spawnY := float64(tileY * tileSize)
	return &DropItem{
		kind:         kind,
		x:            float64(tileX*tileSize) + (tileSize-dropItemSize)/2,
		y:            spawnY,
		vx:           dropItemSpeed,
		emergeTarget: spawnY - tileSize,
		emerging:     true,
		alive:        true,
	}
}

func (d *DropItem) Rect() rect {
	return rect{d.x, d.y, dropItemSize, dropItemSize}
}

func (d *DropItem) Update(st *Stage) {
	if !d.alive {
		return
	}
	if d.emerging {
		d.y -= 1
		if d.y <= d.emergeTarget {
			d.y = d.emergeTarget
			d.emerging = false
		}
		return
	}

	d.vy += gravity
	if d.vy > maxFallSpeed {
		d.vy = maxFallSpeed
	}

	nextX := d.x + d.vx
	var checkTx int
	if d.vx < 0 {
		checkTx = int(math.Floor(nextX / tileSize))
	} else {
		checkTx = int(math.Floor((nextX + dropItemSize - 1) / tileSize))
	}
	top := int(math.Floor(d.y / tileSize))
	bottom := int(math.Floor((d.y + dropItemSize - 1) / tileSize))
	blocked := false
	for ty := top; ty <= bottom; ty++ {
		if st.isSolid(checkTx, ty) {
			blocked = true
			break
		}
	}
	if blocked {
		d.vx = -d.vx
	} else {
		d.x = nextX
	}

	d.y += d.vy
	if d.vy >= 0 {
		left := int(math.Floor(d.x / tileSize))
		right := int(math.Floor((d.x + dropItemSize - 1) / tileSize))
		by := int(math.Floor((d.y + dropItemSize - 1) / tileSize))
		for tx := left; tx <= right; tx++ {
			if st.isSolid(tx, by) {
				d.y = float64(by*tileSize) - dropItemSize
				d.vy = 0
				break
			}
		}
	}

	if d.y > float64(heightTiles*tileSize) {
		d.alive = false
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
