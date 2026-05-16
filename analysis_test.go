package main

import "testing"

func TestOpponentCeiling_EmptyBoard(t *testing.T) {
	b := NewBoard()
	if c := OpponentCeiling(b, nil); c != 0 {
		t.Errorf("empty board ceiling should be 0, got %d", c)
	}
}

func TestOpponentCeiling_OpenLine(t *testing.T) {
	// 5-tile red line; ceiling is 12 (Qwirkle by completing with R+).
	b := boardWith(t, map[Coord]string{
		{0, 0}: "Ro",
		{1, 0}: "Rs",
		{2, 0}: "Rd",
		{3, 0}: "Rc",
		{4, 0}: "R*",
	})
	c := OpponentCeiling(b, nil)
	if c != 12 {
		t.Errorf("expected ceiling 12 (completed qwirkle), got %d", c)
	}
}

func TestOpponentCeiling_RespectsAvailable(t *testing.T) {
	b := boardWith(t, map[Coord]string{
		{0, 0}: "Ro",
		{1, 0}: "Rs",
		{2, 0}: "Rd",
		{3, 0}: "Rc",
		{4, 0}: "R*",
	})
	// If R+ is not available, the qwirkle can't be completed.
	avail := map[Tile]bool{}
	for c := Color(0); c < NumColors; c++ {
		for s := Shape(0); s < NumShapes; s++ {
			avail[Tile{c, s}] = true
		}
	}
	delete(avail, mustTile(t, "R+"))
	got := OpponentCeiling(b, avail)
	if got >= 12 {
		t.Errorf("without R+, ceiling should be < 12; got %d", got)
	}
}
