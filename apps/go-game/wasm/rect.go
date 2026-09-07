package main

// rect is an axis-aligned bounding box in pixel space, used for collision checks
// between entities (player/enemy/coin) that don't align to the tile grid.
type rect struct {
	x, y, w, h float64
}

func (r rect) overlaps(o rect) bool {
	return r.x < o.x+o.w && r.x+r.w > o.x && r.y < o.y+o.h && r.y+r.h > o.y
}
