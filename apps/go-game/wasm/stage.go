package main

import "math/rand"

// tile represents the content of one grid cell in a stage.
type tile byte

const (
	tileEmpty tile = iota
	tileHard       // indestructible ground/block
	tileBrick      // destroyed when hit from below
	tileItem       // spawns a coin when hit from below, then becomes tileUsed
	tileUsed       // an item block that has already been used (still solid)
	tilePipe       // solid pipe; standing on its top row + pressing Down enters another stage
)

// point is a tile-grid coordinate (not pixels).
type point struct{ x, y int }

// pipeSpec places a pipe of the given size on top of the ground, entered by
// standing on its top row and pressing Down.
type pipeSpec struct {
	x, width, height int
	nextStage        int
}

// stageSpec is the human-authored definition of a stage, expressed in tile
// coordinates rather than raw ASCII art so it's easy to validate and extend.
type stageSpec struct {
	widthTiles  int
	groundY     int // topmost ground row; ground occupies rows [groundY, heightTiles)
	pits        [][2]int
	bricks      []point
	items       []point
	oneUpItems  []point // subset of items that drop a 1UP instead of a coin
	walls       []point
	enemies     []point
	pipes       []pipeSpec
	playerStart point
}

type pipe struct {
	tileX, tileY, widthTiles, heightTiles, nextStage int
}

// Stage is the built, mutable runtime representation of a stageSpec.
type Stage struct {
	tiles       [][]tile
	widthTiles  int
	playerStart struct{ x, y float64 }
	enemySpawns []struct{ x, y float64 }
	pipes       []pipe
	oneUps      []point
}

func tileToPixel(p point) struct{ x, y float64 } {
	return struct{ x, y float64 }{float64(p.x * tileSize), float64(p.y * tileSize)}
}

func buildStage(spec stageSpec) *Stage {
	tiles := make([][]tile, heightTiles)
	for y := range tiles {
		tiles[y] = make([]tile, spec.widthTiles)
	}

	inPit := func(x int) bool {
		for _, r := range spec.pits {
			if x >= r[0] && x <= r[1] {
				return true
			}
		}
		return false
	}
	for x := 0; x < spec.widthTiles; x++ {
		if inPit(x) {
			continue
		}
		for y := spec.groundY; y < heightTiles; y++ {
			tiles[y][x] = tileHard
		}
	}
	for _, p := range spec.bricks {
		tiles[p.y][p.x] = tileBrick
	}
	for _, p := range spec.items {
		tiles[p.y][p.x] = tileItem
	}
	for _, p := range spec.walls {
		tiles[p.y][p.x] = tileHard
	}
	for _, ps := range spec.pipes {
		for dx := 0; dx < ps.width; dx++ {
			for dy := 0; dy < ps.height; dy++ {
				y := spec.groundY - 1 - dy
				if y >= 0 {
					tiles[y][ps.x+dx] = tilePipe
				}
			}
		}
	}

	st := &Stage{
		tiles:      tiles,
		widthTiles: spec.widthTiles,
		oneUps:     spec.oneUpItems,
	}
	st.playerStart = tileToPixel(spec.playerStart)
	for _, ps := range spec.pipes {
		st.pipes = append(st.pipes, pipe{
			tileX:       ps.x,
			tileY:       spec.groundY - ps.height,
			widthTiles:  ps.width,
			heightTiles: ps.height,
			nextStage:   ps.nextStage,
		})
	}
	for _, p := range spec.enemies {
		st.enemySpawns = append(st.enemySpawns, tileToPixel(p))
	}
	return st
}

func (s *Stage) tileAt(tx, ty int) tile {
	if tx < 0 || tx >= s.widthTiles || ty < 0 || ty >= heightTiles {
		return tileEmpty
	}
	return s.tiles[ty][tx]
}

func (s *Stage) setTile(tx, ty int, t tile) {
	if tx < 0 || tx >= s.widthTiles || ty < 0 || ty >= heightTiles {
		return
	}
	s.tiles[ty][tx] = t
}

func (s *Stage) isSolid(tx, ty int) bool {
	return s.tileAt(tx, ty) != tileEmpty
}

func (s *Stage) widthPx() float64 {
	return float64(s.widthTiles * tileSize)
}

func (s *Stage) pipeEntranceAt(tx, ty int) *pipe {
	for i := range s.pipes {
		p := &s.pipes[i]
		if ty == p.tileY && tx >= p.tileX && tx < p.tileX+p.widthTiles {
			return p
		}
	}
	return nil
}

func (s *Stage) isOneUp(tx, ty int) bool {
	for _, p := range s.oneUps {
		if p.x == tx && p.y == ty {
			return true
		}
	}
	return false
}

// stageBuilder assembles a stageSpec out of fixed-width chunks so features
// never overlap: each chunk owns an exclusive 10-tile slice of the stage,
// and any pit/enemy/block within it is centered with ground to spare on
// both sides, guaranteeing the stage stays completable.
const chunkWidth = 10

type stageBuilder struct {
	pits       [][2]int
	bricks     []point
	items      []point
	oneUpItems []point
	walls      []point
	enemies    []point
	col        int
}

func (b *stageBuilder) plain() {
	b.col += chunkWidth
}

func (b *stageBuilder) pit(width int) {
	start := b.col + (chunkWidth-width)/2
	b.pits = append(b.pits, [2]int{start, start + width - 1})
	b.col += chunkWidth
}

func (b *stageBuilder) bricksAndItem(oneUp bool) {
	base := b.col + 2
	b.bricks = append(b.bricks, point{base, 6}, point{base + 1, 6}, point{base + 2, 6})
	item := point{base + 3, 6}
	b.items = append(b.items, item)
	if oneUp {
		b.oneUpItems = append(b.oneUpItems, item)
	}
	b.col += chunkWidth
}

func (b *stageBuilder) platform() {
	base := b.col + 3
	b.walls = append(b.walls, point{base, 8}, point{base + 1, 8}, point{base + 2, 8})
	b.col += chunkWidth
}

func (b *stageBuilder) enemy() {
	b.enemies = append(b.enemies, point{b.col + 5, 9})
	b.col += chunkWidth
}

// generateStage deterministically builds stage idx (0-based) out of random
// but bounded chunks, growing longer and busier for later stages. Its pipe
// leads to the next stage, wrapping from the last stage back to the first.
func generateStage(idx int) stageSpec {
	rng := rand.New(rand.NewSource(int64(idx)*97 + 13))
	b := &stageBuilder{col: 6} // flat safe zone at the spawn point

	numChunks := 6 + idx
	oneUpChunk := numChunks / 2

	kinds := []string{"plain", "plain", "bricks", "platform"}
	for e := 0; e <= idx/3; e++ {
		kinds = append(kinds, "enemy")
	}
	for p := 0; p <= idx/4; p++ {
		kinds = append(kinds, "pit")
	}

	for i := 0; i < numChunks; i++ {
		if i == oneUpChunk {
			b.bricksAndItem(true)
			continue
		}
		switch kinds[rng.Intn(len(kinds))] {
		case "plain":
			b.plain()
		case "bricks":
			b.bricksAndItem(false)
		case "enemy":
			b.enemy()
		case "platform":
			b.platform()
		case "pit":
			width := 2
			if idx > 3 && rng.Intn(2) == 0 {
				width = 3
			}
			b.pit(width)
		}
	}

	b.col += 2
	pipeX := b.col
	b.col += 2 + 3

	return stageSpec{
		widthTiles:  b.col,
		groundY:     10,
		pits:        b.pits,
		bricks:      b.bricks,
		items:       b.items,
		oneUpItems:  b.oneUpItems,
		walls:       b.walls,
		enemies:     b.enemies,
		pipes:       []pipeSpec{{x: pipeX, width: 2, height: 2, nextStage: (idx + 1) % numStages}},
		playerStart: point{2, 9},
	}
}

// allStageSpecs builds the full numStages-stage progression; the last
// stage's pipe loops back to the first.
func allStageSpecs() []stageSpec {
	specs := make([]stageSpec, numStages)
	for i := range specs {
		specs[i] = generateStage(i)
	}
	return specs
}
