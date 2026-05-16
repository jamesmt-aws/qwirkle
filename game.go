package main

import (
	"errors"
	"fmt"
)

const HandSize = 6

type Game struct {
	Seed       int64
	NumPlayers int
	Names      []string

	board   *Board
	bag     *Bag
	hands   [][]Tile
	scores  []int
	current int   // index of player whose turn it is
	log     []Move

	// finished is true when the game has ended (someone emptied their hand
	// after the bag ran out, or no progress is possible).
	finished bool
}

func NewGame(seed int64, names []string) *Game {
	g := &Game{
		Seed:       seed,
		NumPlayers: len(names),
		Names:      append([]string{}, names...),
		board:      NewBoard(),
		bag:        NewBag(seed),
		hands:      make([][]Tile, len(names)),
		scores:     make([]int, len(names)),
	}
	for i := range g.hands {
		g.hands[i] = g.bag.Draw(HandSize)
	}
	return g
}

func (g *Game) Board() *Board     { return g.board }
func (g *Game) Hand(i int) []Tile { return g.hands[i] }
func (g *Game) Score(i int) int   { return g.scores[i] }
func (g *Game) Current() int      { return g.current }
func (g *Game) BagRemaining() int { return g.bag.Remaining() }
func (g *Game) Finished() bool    { return g.finished }
func (g *Game) Log() []Move       { return g.log }

// Apply executes the move for the current player and advances the turn.
func (g *Game) Apply(m Move) error {
	if g.finished {
		return errors.New("game is over")
	}
	hand := g.hands[g.current]

	switch m.Type {
	case MovePlace:
		// Check hand contains the tiles being placed (multiset).
		if err := checkSubset(hand, placementTiles(m.Placements)); err != nil {
			return fmt.Errorf("hand: %w", err)
		}
		res, err := ValidatePlace(g.board, m.Placements)
		if err != nil {
			return err
		}
		for _, p := range m.Placements {
			g.board.Place(p.Coord, p.Tile)
		}
		g.hands[g.current] = removeTiles(hand, placementTiles(m.Placements))
		g.scores[g.current] += res.Score

		// End-of-game bonus: if bag is empty and player empties their hand.
		usedAll := len(g.hands[g.current]) == 0 && g.bag.Remaining() == 0
		// Refill (no-op if bag is empty).
		refill := g.bag.Draw(HandSize - len(g.hands[g.current]))
		g.hands[g.current] = append(g.hands[g.current], refill...)

		if usedAll {
			g.scores[g.current] += QwirkleBonus
			g.finished = true
		}

	case MoveExchange:
		if g.bag.Remaining() == 0 {
			return errors.New("cannot exchange: bag is empty")
		}
		if len(m.Exchange) == 0 {
			return errors.New("exchange requires at least one tile (or play instead)")
		}
		if err := checkSubset(hand, m.Exchange); err != nil {
			return fmt.Errorf("hand: %w", err)
		}
		drawn := g.bag.Exchange(m.Exchange)
		g.hands[g.current] = removeTiles(hand, m.Exchange)
		g.hands[g.current] = append(g.hands[g.current], drawn...)

	default:
		return fmt.Errorf("unknown move type %d", m.Type)
	}

	g.log = append(g.log, m)
	if !g.finished {
		g.current = (g.current + 1) % g.NumPlayers
	}
	return nil
}

func placementTiles(ps []Placement) []Tile {
	out := make([]Tile, len(ps))
	for i, p := range ps {
		out[i] = p.Tile
	}
	return out
}

// checkSubset confirms that `want` is contained (as a multiset) in `have`.
func checkSubset(have, want []Tile) error {
	counts := map[Tile]int{}
	for _, t := range have {
		counts[t]++
	}
	for _, t := range want {
		if counts[t] == 0 {
			return fmt.Errorf("missing tile %s", t)
		}
		counts[t]--
	}
	return nil
}

// removeTiles returns `have` with one copy of each tile in `want` removed.
// Assumes checkSubset has succeeded.
func removeTiles(have, want []Tile) []Tile {
	remove := map[Tile]int{}
	for _, t := range want {
		remove[t]++
	}
	out := have[:0]
	for _, t := range have {
		if remove[t] > 0 {
			remove[t]--
			continue
		}
		out = append(out, t)
	}
	return out
}

// Winner returns the index of the winning player, or -1 if tied.
func (g *Game) Winner() int {
	best, bestI, tied := -1, -1, false
	for i, s := range g.scores {
		if s > best {
			best, bestI, tied = s, i, false
		} else if s == best {
			tied = true
		}
	}
	if tied {
		return -1
	}
	return bestI
}
