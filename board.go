package main

type Board struct {
	cells map[Coord]Tile
}

func NewBoard() *Board { return &Board{cells: map[Coord]Tile{}} }

func (b *Board) At(c Coord) (Tile, bool) {
	t, ok := b.cells[c]
	return t, ok
}

func (b *Board) Place(c Coord, t Tile) { b.cells[c] = t }

func (b *Board) IsEmpty() bool { return len(b.cells) == 0 }

func (b *Board) Count() int { return len(b.cells) }

func (b *Board) Clone() *Board {
	nb := &Board{cells: make(map[Coord]Tile, len(b.cells))}
	for k, v := range b.cells {
		nb.cells[k] = v
	}
	return nb
}

// Bounds returns the inclusive (min, max) of occupied cells. Caller must check
// IsEmpty before calling.
func (b *Board) Bounds() (Coord, Coord) {
	var min, max Coord
	first := true
	for c := range b.cells {
		if first {
			min, max = c, c
			first = false
			continue
		}
		if c.X < min.X {
			min.X = c.X
		}
		if c.Y < min.Y {
			min.Y = c.Y
		}
		if c.X > max.X {
			max.X = c.X
		}
		if c.Y > max.Y {
			max.Y = c.Y
		}
	}
	return min, max
}

// Line returns the coordinates of the maximal contiguous run of occupied
// cells through `c` in direction `d` (extended both ways). Empty if `c`
// itself is unoccupied.
func (b *Board) Line(c Coord, d Coord) []Coord {
	if _, ok := b.cells[c]; !ok {
		return nil
	}
	start := c
	for {
		prev := start.Sub(d)
		if _, ok := b.cells[prev]; !ok {
			break
		}
		start = prev
	}
	var line []Coord
	for p := start; ; p = p.Add(d) {
		if _, ok := b.cells[p]; !ok {
			break
		}
		line = append(line, p)
	}
	return line
}
