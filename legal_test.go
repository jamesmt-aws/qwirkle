package main

import "testing"

func TestLegal_EmptyBoardOpening(t *testing.T) {
	b := NewBoard()
	hand := []Tile{
		mustTile(t, "Ro"),
		mustTile(t, "Rs"),
		mustTile(t, "Rd"),
		mustTile(t, "Ys"),
	}
	moves := LegalPlacements(b, hand)
	if len(moves) == 0 {
		t.Fatal("expected some legal opening moves")
	}
	// All moves should validate.
	for _, m := range moves {
		if _, err := ValidatePlace(b, m.Placements); err != nil {
			t.Errorf("enumerated move failed validation: %v\nmove: %v", err, m.Placements)
		}
	}
}

func TestLegal_FindsBridgeMove(t *testing.T) {
	// Existing: Ro at (0,0), Go at (2,0). Hand has Yo. Should find placing
	// Yo at (1,0) bridging into a Ro Yo Go line.
	b := boardWith(t, map[Coord]string{
		{0, 0}: "Ro",
		{2, 0}: "Go",
	})
	hand := []Tile{mustTile(t, "Yo"), mustTile(t, "Po"), mustTile(t, "Bs")}
	moves := LegalPlacements(b, hand)
	found := false
	for _, m := range moves {
		if len(m.Placements) == 1 &&
			m.Placements[0].Coord == (Coord{1, 0}) &&
			m.Placements[0].Tile == mustTile(t, "Yo") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected bridge move Yo@1,0 in legal moves; got %d moves", len(moves))
	}
}

func TestLegal_FindsBothSidesMove(t *testing.T) {
	// Existing: single Ro at (5,0). Hand has Rs, Rd, R*.
	// Should find moves placing tiles on both sides of (5,0), e.g.
	// {Rs@4,0, Rd@6,0}.
	b := boardWith(t, map[Coord]string{
		{5, 0}: "Ro",
	})
	hand := []Tile{mustTile(t, "Rs"), mustTile(t, "Rd"), mustTile(t, "R*")}
	moves := LegalPlacements(b, hand)
	for _, m := range moves {
		if _, err := ValidatePlace(b, m.Placements); err != nil {
			t.Errorf("invalid enumerated move: %v\n%v", err, m.Placements)
		}
	}
	found := false
	for _, m := range moves {
		if len(m.Placements) == 2 {
			coords := map[Coord]bool{}
			for _, p := range m.Placements {
				coords[p.Coord] = true
			}
			if coords[Coord{4, 0}] && coords[Coord{6, 0}] {
				found = true
				break
			}
		}
	}
	if !found {
		t.Error("expected to find a move placing tiles on both sides of (5,0)")
	}
}

func TestLegal_DedupeWorks(t *testing.T) {
	b := boardWith(t, map[Coord]string{
		{0, 0}: "Ro",
	})
	hand := []Tile{mustTile(t, "Rs"), mustTile(t, "Rd")}
	moves := LegalPlacements(b, hand)
	keys := map[string]bool{}
	for _, m := range moves {
		k := placementKey(m.Placements)
		if keys[k] {
			t.Errorf("duplicate move after dedupe: %s", k)
		}
		keys[k] = true
	}
}
