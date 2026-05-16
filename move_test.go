package main

import "testing"

func mustTile(t *testing.T, s string) Tile {
	t.Helper()
	tile, err := ParseTile(s)
	if err != nil {
		t.Fatalf("ParseTile(%q): %v", s, err)
	}
	return tile
}

func boardWith(t *testing.T, tiles map[Coord]string) *Board {
	t.Helper()
	b := NewBoard()
	for c, s := range tiles {
		b.Place(c, mustTile(t, s))
	}
	return b
}

func TestValidate_OpeningSingleTile(t *testing.T) {
	b := NewBoard()
	res, err := ValidatePlace(b, []Placement{{Coord{0, 0}, mustTile(t, "Ro")}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Score != 1 {
		t.Errorf("opening single tile should score 1, got %d", res.Score)
	}
}

func TestValidate_OpeningLine(t *testing.T) {
	b := NewBoard()
	placements := []Placement{
		{Coord{0, 0}, mustTile(t, "Ro")},
		{Coord{1, 0}, mustTile(t, "Yo")},
		{Coord{2, 0}, mustTile(t, "Go")},
	}
	res, err := ValidatePlace(b, placements)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Score != 3 {
		t.Errorf("3-tile line should score 3, got %d", res.Score)
	}
}

func TestValidate_Qwirkle(t *testing.T) {
	b := NewBoard()
	placements := []Placement{
		{Coord{0, 0}, mustTile(t, "Ro")},
		{Coord{1, 0}, mustTile(t, "Rs")},
		{Coord{2, 0}, mustTile(t, "Rd")},
		{Coord{3, 0}, mustTile(t, "Rc")},
		{Coord{4, 0}, mustTile(t, "R*")},
		{Coord{5, 0}, mustTile(t, "R+")},
	}
	res, err := ValidatePlace(b, placements)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Score != 12 {
		t.Errorf("qwirkle should score 6+6=12, got %d", res.Score)
	}
}

func TestValidate_DuplicateInLine(t *testing.T) {
	b := NewBoard()
	placements := []Placement{
		{Coord{0, 0}, mustTile(t, "Ro")},
		{Coord{1, 0}, mustTile(t, "Ro")},
	}
	if _, err := ValidatePlace(b, placements); err == nil {
		t.Error("expected error for duplicate tile in line")
	}
}

func TestValidate_NoSharedAttribute(t *testing.T) {
	b := NewBoard()
	placements := []Placement{
		{Coord{0, 0}, mustTile(t, "Ro")},
		{Coord{1, 0}, mustTile(t, "Ys")},
	}
	if _, err := ValidatePlace(b, placements); err == nil {
		t.Error("expected error: tiles share neither color nor shape")
	}
}

func TestValidate_NotCollinear(t *testing.T) {
	b := NewBoard()
	placements := []Placement{
		{Coord{0, 0}, mustTile(t, "Ro")},
		{Coord{1, 1}, mustTile(t, "Yo")},
	}
	if _, err := ValidatePlace(b, placements); err == nil {
		t.Error("expected error for non-collinear placement")
	}
}

func TestValidate_GapWithoutBridge(t *testing.T) {
	b := NewBoard()
	placements := []Placement{
		{Coord{0, 0}, mustTile(t, "Ro")},
		{Coord{2, 0}, mustTile(t, "Yo")},
	}
	if _, err := ValidatePlace(b, placements); err == nil {
		t.Error("expected error for gap in line")
	}
}

func TestValidate_GapBridgedByExisting(t *testing.T) {
	b := boardWith(t, map[Coord]string{
		{1, 0}: "Po",
	})
	placements := []Placement{
		{Coord{0, 0}, mustTile(t, "Ro")},
		{Coord{2, 0}, mustTile(t, "Yo")},
	}
	res, err := ValidatePlace(b, placements)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Score != 3 {
		t.Errorf("line of 3 should score 3, got %d", res.Score)
	}
}

func TestValidate_Disconnected(t *testing.T) {
	b := boardWith(t, map[Coord]string{
		{0, 0}: "Ro",
	})
	placements := []Placement{
		{Coord{5, 5}, mustTile(t, "Yo")},
	}
	if _, err := ValidatePlace(b, placements); err == nil {
		t.Error("expected error for disconnected placement")
	}
}

func TestValidate_PerpendicularScoring(t *testing.T) {
	// Existing horizontal line Ro Yo Go at y=0, x=0..2.
	b := boardWith(t, map[Coord]string{
		{0, 0}: "Ro",
		{1, 0}: "Yo",
		{2, 0}: "Go",
	})
	// Place Rs below Ro (vertical line of 2, share Red).
	res, err := ValidatePlace(b, []Placement{{Coord{0, 1}, mustTile(t, "Rs")}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Vertical line Ro Rs (len 2) = 2; horizontal line at y=1 is just Rs alone, len 1, no score.
	if res.Score != 2 {
		t.Errorf("expected score 2, got %d", res.Score)
	}
}

func TestValidate_TwoLinesScored(t *testing.T) {
	// Existing: horizontal Ro Yo Go (y=0), vertical Ro Rs (x=0).
	b := boardWith(t, map[Coord]string{
		{0, 0}: "Ro",
		{1, 0}: "Yo",
		{2, 0}: "Go",
		{0, 1}: "Rs",
	})
	// Place Ys at (1,1): horizontal line y=1 is Rs Ys (len 2, both squares),
	// vertical line x=1 is Yo Ys (len 2, both yellow).
	res, err := ValidatePlace(b, []Placement{{Coord{1, 1}, mustTile(t, "Ys")}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Score != 4 {
		t.Errorf("expected score 4 (2+2), got %d", res.Score)
	}
}

func TestValidate_QwirkleBonusViaExtension(t *testing.T) {
	// Existing 5-tile line, complete to qwirkle.
	b := boardWith(t, map[Coord]string{
		{0, 0}: "Ro",
		{1, 0}: "Rs",
		{2, 0}: "Rd",
		{3, 0}: "Rc",
		{4, 0}: "R*",
	})
	res, err := ValidatePlace(b, []Placement{{Coord{5, 0}, mustTile(t, "R+")}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Score != 12 {
		t.Errorf("qwirkle should score 6+6=12, got %d", res.Score)
	}
}

func TestValidate_LineTooLong(t *testing.T) {
	b := boardWith(t, map[Coord]string{
		{0, 0}: "Ro",
		{1, 0}: "Rs",
		{2, 0}: "Rd",
		{3, 0}: "Rc",
		{4, 0}: "R*",
		{5, 0}: "R+",
	})
	// Place a 7th red tile — impossible since only 6 shapes exist, but
	// try placing any red tile at (6,0); it would duplicate something.
	if _, err := ValidatePlace(b, []Placement{{Coord{6, 0}, mustTile(t, "Ro")}}); err == nil {
		t.Error("expected error extending a complete line")
	}
}
