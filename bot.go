package main

import (
	"math/rand"
	"sort"
)

type Bot interface {
	Choose(g *Game) Move
}

type GreedyBot struct {
	rng *rand.Rand
}

func NewGreedyBot(seed int64) *GreedyBot {
	return &GreedyBot{rng: rand.New(rand.NewSource(seed))}
}

func (b *GreedyBot) Choose(g *Game) Move {
	hand := g.Hand(g.Current())
	moves := LegalPlacements(g.Board(), hand)
	if len(moves) == 0 {
		return b.exchangeAll(hand, g.BagRemaining())
	}

	type scored struct {
		move  Move
		score int
	}
	all := make([]scored, 0, len(moves))
	best := -1
	for _, m := range moves {
		res, err := ValidatePlace(g.Board(), m.Placements)
		if err != nil {
			continue
		}
		all = append(all, scored{m, res.Score})
		if res.Score > best {
			best = res.Score
		}
	}
	if best < 0 {
		return b.exchangeAll(hand, g.BagRemaining())
	}
	var top []Move
	for _, s := range all {
		if s.score == best {
			top = append(top, s.move)
		}
	}
	// Stable tiebreak: lexicographic by sorted placements, then random pick among ties.
	sort.Slice(top, func(i, j int) bool {
		return placementKey(top[i].Placements) < placementKey(top[j].Placements)
	})
	return top[b.rng.Intn(len(top))]
}

func (b *GreedyBot) exchangeAll(hand []Tile, bagRemaining int) Move {
	if bagRemaining == 0 {
		// Forced pass: exchange of zero. Game.Apply will reject empty
		// exchange; caller treats this as a skipped turn. For now, return
		// an empty exchange and let the loop detect it.
		return Move{Type: MoveExchange, Exchange: nil}
	}
	n := len(hand)
	if n > bagRemaining {
		n = bagRemaining
	}
	out := make([]Tile, n)
	copy(out, hand[:n])
	return Move{Type: MoveExchange, Exchange: out}
}

func placementKey(ps []Placement) string {
	sorted := append([]Placement{}, ps...)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Coord.X != sorted[j].Coord.X {
			return sorted[i].Coord.X < sorted[j].Coord.X
		}
		return sorted[i].Coord.Y < sorted[j].Coord.Y
	})
	out := make([]byte, 0, len(sorted)*16)
	for _, p := range sorted {
		out = append(out, p.Tile.Code()...)
		out = append(out, '@')
		out = append(out, []byte(itoa(p.Coord.X))...)
		out = append(out, ',')
		out = append(out, []byte(itoa(p.Coord.Y))...)
		out = append(out, '|')
	}
	return string(out)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
