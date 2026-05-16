package main

import (
	"errors"
	"fmt"
)

type MoveType uint8

const (
	MovePlace MoveType = iota
	MoveExchange
)

type Placement struct {
	Coord Coord
	Tile  Tile
}

type Move struct {
	Type       MoveType
	Placements []Placement // for MovePlace
	Exchange   []Tile      // for MoveExchange (may be empty = forced pass)
}

const QwirkleLen = 6
const QwirkleBonus = 6

type MoveResult struct {
	Score int
}

// ValidatePlace validates a placement move against the board and returns its
// score. It does not mutate the board.
func ValidatePlace(b *Board, placements []Placement) (MoveResult, error) {
	if len(placements) == 0 {
		return MoveResult{}, errors.New("no placements")
	}

	// 1. Unique coords; no overlap with existing.
	placedAt := make(map[Coord]bool, len(placements))
	for _, p := range placements {
		if placedAt[p.Coord] {
			return MoveResult{}, fmt.Errorf("duplicate placement at %v", p.Coord)
		}
		placedAt[p.Coord] = true
		if _, occ := b.At(p.Coord); occ {
			return MoveResult{}, fmt.Errorf("cell %v already occupied", p.Coord)
		}
	}

	// 2. Determine placement axis (collinear).
	var axis Coord
	if len(placements) == 1 {
		axis = Right // arbitrary; only used for single-tile placement line dedup
	} else {
		x0 := placements[0].Coord.X
		y0 := placements[0].Coord.Y
		sameRow, sameCol := true, true
		for _, p := range placements[1:] {
			if p.Coord.Y != y0 {
				sameRow = false
			}
			if p.Coord.X != x0 {
				sameCol = false
			}
		}
		switch {
		case sameRow && !sameCol:
			axis = Right
		case sameCol && !sameRow:
			axis = Down
		case sameRow && sameCol:
			return MoveResult{}, errors.New("duplicate placements")
		default:
			return MoveResult{}, errors.New("placements not in a single row or column")
		}
	}

	// 3. Build tentative board.
	tmp := b.Clone()
	for _, p := range placements {
		tmp.Place(p.Coord, p.Tile)
	}

	// 4. Contiguity along axis: walk from min to max placed; every cell on
	//    the line must be occupied in tmp.
	if len(placements) > 1 {
		minP, maxP := placements[0].Coord, placements[0].Coord
		for _, p := range placements[1:] {
			if axis == Right {
				if p.Coord.X < minP.X {
					minP = p.Coord
				}
				if p.Coord.X > maxP.X {
					maxP = p.Coord
				}
			} else {
				if p.Coord.Y < minP.Y {
					minP = p.Coord
				}
				if p.Coord.Y > maxP.Y {
					maxP = p.Coord
				}
			}
		}
		for c := minP; ; c = c.Add(axis) {
			if _, ok := tmp.At(c); !ok {
				return MoveResult{}, fmt.Errorf("gap in placement line at %v", c)
			}
			if c == maxP {
				break
			}
		}
	}

	// 5. Connectivity: unless the board was empty, at least one placed tile
	//    must be edge-adjacent to a pre-existing tile.
	if !b.IsEmpty() {
		touches := false
		for _, p := range placements {
			for _, d := range neighbors4 {
				n := p.Coord.Add(d)
				if placedAt[n] {
					continue
				}
				if _, ok := b.At(n); ok {
					touches = true
					break
				}
			}
			if touches {
				break
			}
		}
		if !touches {
			return MoveResult{}, errors.New("placement is not connected to existing tiles")
		}
	}

	// 6. Validate every line through a placed tile and accumulate score.
	type lineKey struct{ Start, Dir Coord }
	scored := map[lineKey]bool{}
	score := 0
	for _, p := range placements {
		for _, d := range axes {
			line := tmp.Line(p.Coord, d)
			if len(line) == 0 {
				continue
			}
			k := lineKey{line[0], d}
			if scored[k] {
				continue
			}
			scored[k] = true
			if !validQwirkleLine(tmp, line) {
				return MoveResult{}, fmt.Errorf("invalid line through %v in direction %v", p.Coord, d)
			}
			if len(line) >= 2 {
				score += len(line)
				if len(line) == QwirkleLen {
					score += QwirkleBonus
				}
			}
		}
	}

	// Opening single-tile placement scores 1 per official rules.
	if score == 0 {
		score = len(placements)
	}

	return MoveResult{Score: score}, nil
}

func validQwirkleLine(b *Board, line []Coord) bool {
	if len(line) <= 1 {
		return true
	}
	if len(line) > QwirkleLen {
		return false
	}
	tiles := make([]Tile, len(line))
	for i, c := range line {
		tiles[i], _ = b.At(c)
	}
	sameColor, sameShape := true, true
	for _, t := range tiles[1:] {
		if t.Color != tiles[0].Color {
			sameColor = false
		}
		if t.Shape != tiles[0].Shape {
			sameShape = false
		}
	}
	if sameColor == sameShape {
		// Either no shared attribute or all duplicates.
		return false
	}
	if sameColor {
		seen := make(map[Shape]bool, len(tiles))
		for _, t := range tiles {
			if seen[t.Shape] {
				return false
			}
			seen[t.Shape] = true
		}
	} else {
		seen := make(map[Color]bool, len(tiles))
		for _, t := range tiles {
			if seen[t.Color] {
				return false
			}
			seen[t.Color] = true
		}
	}
	return true
}
