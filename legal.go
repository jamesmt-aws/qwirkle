package main

import (
	"sort"
	"strings"
)

// LegalPlacements returns all legal placement moves for `hand` on `board`.
// Returns nil if there are no legal placements (caller may then exchange).
func LegalPlacements(board *Board, hand []Tile) []Move {
	var out []Move
	anchors := anchorCells(board)
	for _, anchor := range anchors {
		for _, axis := range axes {
			seen := map[Tile]bool{}
			for i, t := range hand {
				if seen[t] {
					continue
				}
				seen[t] = true
				p := Placement{Coord: anchor, Tile: t}
				if _, err := ValidatePlace(board, []Placement{p}); err != nil {
					continue
				}
				placed := []Placement{p}
				newHand := dropAt(hand, i)
				out = append(out, Move{Type: MovePlace, Placements: append([]Placement{}, placed...)})
				extendBothWays(board, axis, placed, newHand, anchor, anchor, &out)
			}
		}
	}
	return dedupeMoves(out)
}

func extendBothWays(board *Board, axis Coord, placed []Placement, hand []Tile, leftmost, rightmost Coord, out *[]Move) {
	if len(hand) == 0 {
		return
	}
	// Find the next empty cell on each side, skipping over any existing tiles.
	leftCell := leftmost.Sub(axis)
	for {
		if _, ok := board.At(leftCell); !ok {
			break
		}
		leftCell = leftCell.Sub(axis)
	}
	rightCell := rightmost.Add(axis)
	for {
		if _, ok := board.At(rightCell); !ok {
			break
		}
		rightCell = rightCell.Add(axis)
	}

	seen := map[Tile]bool{}
	for i, t := range hand {
		if seen[t] {
			continue
		}
		seen[t] = true
		newHand := dropAt(hand, i)
		// Try left
		left := append(append([]Placement{}, placed...), Placement{leftCell, t})
		if _, err := ValidatePlace(board, left); err == nil {
			*out = append(*out, Move{Type: MovePlace, Placements: append([]Placement{}, left...)})
			extendBothWays(board, axis, left, newHand, leftCell, rightmost, out)
		}
		// Try right
		right := append(append([]Placement{}, placed...), Placement{rightCell, t})
		if _, err := ValidatePlace(board, right); err == nil {
			*out = append(*out, Move{Type: MovePlace, Placements: append([]Placement{}, right...)})
			extendBothWays(board, axis, right, newHand, leftmost, rightCell, out)
		}
	}
}

// anchorCells returns all empty cells adjacent to a placed tile, or [(0,0)]
// if the board is empty.
func anchorCells(b *Board) []Coord {
	if b.IsEmpty() {
		return []Coord{{0, 0}}
	}
	set := map[Coord]bool{}
	for c := range b.cells {
		for _, d := range neighbors4 {
			n := c.Add(d)
			if _, occ := b.At(n); occ {
				continue
			}
			set[n] = true
		}
	}
	out := make([]Coord, 0, len(set))
	for c := range set {
		out = append(out, c)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Y != out[j].Y {
			return out[i].Y < out[j].Y
		}
		return out[i].X < out[j].X
	})
	return out
}

func dropAt(hand []Tile, i int) []Tile {
	out := make([]Tile, 0, len(hand)-1)
	out = append(out, hand[:i]...)
	out = append(out, hand[i+1:]...)
	return out
}

func dedupeMoves(moves []Move) []Move {
	seen := map[string]bool{}
	out := make([]Move, 0, len(moves))
	var sb strings.Builder
	for _, m := range moves {
		ps := append([]Placement{}, m.Placements...)
		sort.Slice(ps, func(i, j int) bool {
			if ps[i].Coord.X != ps[j].Coord.X {
				return ps[i].Coord.X < ps[j].Coord.X
			}
			return ps[i].Coord.Y < ps[j].Coord.Y
		})
		sb.Reset()
		for _, p := range ps {
			sb.WriteString(p.Tile.Code())
			sb.WriteByte('@')
			writeInt(&sb, p.Coord.X)
			sb.WriteByte(',')
			writeInt(&sb, p.Coord.Y)
			sb.WriteByte('|')
		}
		k := sb.String()
		if !seen[k] {
			seen[k] = true
			out = append(out, Move{Type: MovePlace, Placements: ps})
		}
	}
	return out
}

func writeInt(sb *strings.Builder, n int) {
	if n < 0 {
		sb.WriteByte('-')
		n = -n
	}
	if n == 0 {
		sb.WriteByte('0')
		return
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	sb.Write(buf[i:])
}
