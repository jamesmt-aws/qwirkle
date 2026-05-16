package main

// UnseenTileTypes returns the set of tile types that, from `byPlayer`'s
// perspective, could still be in the bag or another player's hand. This is
// the universe of tiles an opponent might place next turn.
func (g *Game) UnseenTileTypes(byPlayer int) map[Tile]bool {
	counts := make(map[Tile]int, int(NumColors)*int(NumShapes))
	for c := Color(0); c < NumColors; c++ {
		for s := Shape(0); s < NumShapes; s++ {
			counts[Tile{c, s}] = TilesPerKind
		}
	}
	for _, t := range g.board.cells {
		counts[t]--
	}
	for _, t := range g.hands[byPlayer] {
		counts[t]--
	}
	out := make(map[Tile]bool, len(counts))
	for t, c := range counts {
		if c > 0 {
			out[t] = true
		}
	}
	return out
}

// OpponentCeiling returns an upper bound on the score the next player could
// earn with a single-tile placement, given they might hold any tile in
// `available`. If `available` is nil, all 36 tile types are considered.
//
// This is an UPPER bound: it assumes the opponent has the perfect tile for
// the best single-tile spot. Multi-tile replies could score more, but in
// practice the best 1-tile reply is a good local-greedy proxy for "how
// much did I expose."
func OpponentCeiling(b *Board, available map[Tile]bool) int {
	if b.IsEmpty() {
		return 0
	}
	tiles := make([]Tile, 0, int(NumColors)*int(NumShapes))
	if available == nil {
		for c := Color(0); c < NumColors; c++ {
			for s := Shape(0); s < NumShapes; s++ {
				tiles = append(tiles, Tile{c, s})
			}
		}
	} else {
		for t := range available {
			tiles = append(tiles, t)
		}
	}
	anchors := anchorCells(b)
	best := 0
	for _, a := range anchors {
		for _, t := range tiles {
			res, err := ValidatePlace(b, []Placement{{a, t}})
			if err != nil {
				continue
			}
			if res.Score > best {
				best = res.Score
			}
		}
	}
	return best
}

// EvaluateMove returns the move's own score plus the opponent's ceiling
// after the move is applied. Net = score - ceiling.
type MoveEval struct {
	Score    int
	Ceiling  int
	Net      int
}

func EvaluateMove(g *Game, m Move) (MoveEval, error) {
	if m.Type != MovePlace {
		// Exchanges score 0 and don't directly change the board, so the
		// ceiling is the current board's ceiling.
		c := OpponentCeiling(g.Board(), g.UnseenTileTypes(g.Current()))
		return MoveEval{Score: 0, Ceiling: c, Net: -c}, nil
	}
	res, err := ValidatePlace(g.Board(), m.Placements)
	if err != nil {
		return MoveEval{}, err
	}
	tmp := g.Board().Clone()
	for _, p := range m.Placements {
		tmp.Place(p.Coord, p.Tile)
	}
	// Available tiles to the opponent: from `g.Current()`'s perspective,
	// unseen tiles minus the tiles we just placed (now visible).
	avail := g.UnseenTileTypes(g.Current())
	c := OpponentCeiling(tmp, avail)
	return MoveEval{Score: res.Score, Ceiling: c, Net: res.Score - c}, nil
}
