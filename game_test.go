package main

import "testing"

func TestGame_Deterministic(t *testing.T) {
	g1 := NewGame(42, []string{"A", "B"})
	g2 := NewGame(42, []string{"A", "B"})
	for i := 0; i < g1.NumPlayers; i++ {
		h1, h2 := g1.Hand(i), g2.Hand(i)
		if len(h1) != len(h2) {
			t.Fatalf("hand %d length mismatch", i)
		}
		for j := range h1 {
			if h1[j] != h2[j] {
				t.Errorf("hand %d tile %d mismatch: %v vs %v", i, j, h1[j], h2[j])
			}
		}
	}
	if g1.BagRemaining() != g2.BagRemaining() {
		t.Errorf("bag remaining mismatch: %d vs %d", g1.BagRemaining(), g2.BagRemaining())
	}
}

func TestGame_BagDistribution(t *testing.T) {
	g := NewGame(1, []string{"A"})
	counts := map[Tile]int{}
	for _, t := range g.Hand(0) {
		counts[t]++
	}
	// drain bag
	for g.BagRemaining() > 0 {
		drawn := g.bag.Draw(1)
		counts[drawn[0]]++
	}
	if len(counts) != int(NumColors)*int(NumShapes) {
		t.Errorf("expected %d unique tiles, got %d", int(NumColors)*int(NumShapes), len(counts))
	}
	for tile, c := range counts {
		if c != TilesPerKind {
			t.Errorf("tile %v: count %d, want %d", tile, c, TilesPerKind)
		}
	}
}

func TestGame_PlayAndScore(t *testing.T) {
	g := NewGame(1, []string{"A", "B"})
	// Force a known state: replace player 0's hand with three reds.
	g.hands[0] = []Tile{
		mustTile(t, "Ro"),
		mustTile(t, "Rs"),
		mustTile(t, "Rd"),
		mustTile(t, "Yc"),
		mustTile(t, "G*"),
		mustTile(t, "B+"),
	}
	m := Move{Type: MovePlace, Placements: []Placement{
		{Coord{0, 0}, mustTile(t, "Ro")},
		{Coord{1, 0}, mustTile(t, "Rs")},
		{Coord{2, 0}, mustTile(t, "Rd")},
	}}
	if err := g.Apply(m); err != nil {
		t.Fatalf("apply: %v", err)
	}
	if g.Score(0) != 3 {
		t.Errorf("score: got %d, want 3", g.Score(0))
	}
	if g.Current() != 1 {
		t.Errorf("turn did not advance; current=%d", g.Current())
	}
	if len(g.Hand(0)) != HandSize {
		t.Errorf("hand size after refill: got %d, want %d", len(g.Hand(0)), HandSize)
	}
}

func TestGame_RejectsTilesNotInHand(t *testing.T) {
	g := NewGame(1, []string{"A", "B"})
	g.hands[0] = []Tile{mustTile(t, "Ro")}
	m := Move{Type: MovePlace, Placements: []Placement{
		{Coord{0, 0}, mustTile(t, "Ys")},
	}}
	if err := g.Apply(m); err == nil {
		t.Error("expected error when placing tile not in hand")
	}
}

func TestGame_BotPlaysSomeGames(t *testing.T) {
	// Smoke test: two greedy bots play to completion without panic.
	g := NewGame(7, []string{"A", "B"})
	a := NewGreedyBot(1)
	b := NewGreedyBot(2)
	consecutivePasses := 0
	maxTurns := 500
	for turn := 0; turn < maxTurns && !g.Finished(); turn++ {
		bot := a
		if g.Current() == 1 {
			bot = b
		}
		m := bot.Choose(g)
		if m.Type == MoveExchange && len(m.Exchange) == 0 {
			consecutivePasses++
			if consecutivePasses >= g.NumPlayers {
				break // deadlock
			}
			g.current = (g.current + 1) % g.NumPlayers
			continue
		}
		consecutivePasses = 0
		if err := g.Apply(m); err != nil {
			t.Fatalf("turn %d: bot produced illegal move: %v\n%+v", turn, err, m)
		}
	}
	if g.Score(0) == 0 && g.Score(1) == 0 {
		t.Errorf("bots scored nothing across the game")
	}
}
