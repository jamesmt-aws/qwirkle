package main

import (
	"fmt"
	"strings"
)

const (
	ansiReset = "\x1b[0m"
	ansiBold  = "\x1b[1m"
	ansiDim   = "\x1b[2m"
)

var ansiColorNum = [NumColors]int{
	Red:    196,
	Orange: 208,
	Yellow: 226,
	Green:  46,
	Blue:   33,
	Purple: 165,
}

var ansiColor = [NumColors]string{
	Red:    "\x1b[38;5;196m",
	Orange: "\x1b[38;5;208m",
	Yellow: "\x1b[38;5;226m",
	Green:  "\x1b[38;5;46m",
	Blue:   "\x1b[38;5;33m",
	Purple: "\x1b[38;5;165m",
}

// RenderTile returns a 2-char colored representation of a tile.
func RenderTile(t Tile) string {
	return ansiColor[t.Color] + t.Code() + ansiReset
}

func renderEmpty() string {
	return ansiDim + ".." + ansiReset
}

// RenderBoard returns a multi-line string showing the board with axis labels.
// The view is padded by `pad` cells around the bounding box of occupied cells.
func RenderBoard(b *Board, pad int) string {
	if b.IsEmpty() {
		return "(empty board — place tiles anywhere; the first tile is the origin)\n"
	}
	min, max := b.Bounds()
	min.X -= pad
	min.Y -= pad
	max.X += pad
	max.Y += pad

	const labelW = 5
	var sb strings.Builder
	// Top column header.
	sb.WriteString(strings.Repeat(" ", labelW+1))
	for x := min.X; x <= max.X; x++ {
		sb.WriteString(fmt.Sprintf("%3d", x))
	}
	sb.WriteByte('\n')
	for y := min.Y; y <= max.Y; y++ {
		sb.WriteString(fmt.Sprintf("%*d: ", labelW, y))
		for x := min.X; x <= max.X; x++ {
			sb.WriteByte(' ')
			if t, ok := b.At(Coord{x, y}); ok {
				sb.WriteString(RenderTile(t))
			} else {
				sb.WriteString(renderEmpty())
			}
		}
		sb.WriteByte('\n')
	}
	return sb.String()
}

func RenderHand(hand []Tile) string {
	parts := make([]string, len(hand))
	for i, t := range hand {
		parts[i] = RenderTile(t)
	}
	return strings.Join(parts, " ")
}

func RenderScoreboard(g *Game) string {
	var sb strings.Builder
	for i, name := range g.Names {
		marker := "  "
		if i == g.Current() && !g.Finished() {
			marker = "->"
		}
		fmt.Fprintf(&sb, "%s %s: %d (hand: %d)\n", marker, name, g.Score(i), len(g.Hand(i)))
	}
	fmt.Fprintf(&sb, "   Bag: %d tiles left\n", g.BagRemaining())
	return sb.String()
}

// RenderBoardTUI renders the board with cursor and pending placement
// overlays. The cursor cell is reverse-video; pending placements are
// bold+underlined.
func RenderBoardTUI(b *Board, pad int, pending []Placement, cursor Coord) string {
	cells := make(map[Coord]Tile, len(b.cells)+len(pending))
	for c, t := range b.cells {
		cells[c] = t
	}
	pendingSet := make(map[Coord]bool, len(pending))
	for _, p := range pending {
		cells[p.Coord] = p.Tile
		pendingSet[p.Coord] = true
	}

	var min, max Coord
	if len(cells) == 0 {
		min, max = cursor, cursor
	} else {
		first := true
		for c := range cells {
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
		if cursor.X < min.X {
			min.X = cursor.X
		}
		if cursor.Y < min.Y {
			min.Y = cursor.Y
		}
		if cursor.X > max.X {
			max.X = cursor.X
		}
		if cursor.Y > max.Y {
			max.Y = cursor.Y
		}
	}
	min.X -= pad
	min.Y -= pad
	max.X += pad
	max.Y += pad

	const labelW = 5
	var sb strings.Builder
	sb.WriteString(strings.Repeat(" ", labelW+1))
	for x := min.X; x <= max.X; x++ {
		sb.WriteString(fmt.Sprintf("%3d", x))
	}
	sb.WriteByte('\n')
	for y := min.Y; y <= max.Y; y++ {
		sb.WriteString(fmt.Sprintf("%*d: ", labelW, y))
		for x := min.X; x <= max.X; x++ {
			sb.WriteByte(' ')
			c := Coord{x, y}
			t, has := cells[c]
			sb.WriteString(renderCellOverlay(t, has, pendingSet[c], cursor == c))
		}
		sb.WriteByte('\n')
	}
	return sb.String()
}

func renderCellOverlay(t Tile, has, pending, cursor bool) string {
	codes := make([]string, 0, 4)
	if cursor {
		codes = append(codes, "7")
	}
	if pending {
		codes = append(codes, "1", "4")
	}
	if has {
		codes = append(codes, fmt.Sprintf("38;5;%d", ansiColorNum[t.Color]))
		return "\x1b[" + strings.Join(codes, ";") + "m" + t.Code() + ansiReset
	}
	if !cursor && !pending {
		return ansiDim + ".." + ansiReset
	}
	if len(codes) == 0 {
		return ".."
	}
	return "\x1b[" + strings.Join(codes, ";") + "m" + ".." + ansiReset
}

func RenderMove(m Move) string {
	switch m.Type {
	case MovePlace:
		parts := make([]string, len(m.Placements))
		for i, p := range m.Placements {
			parts[i] = fmt.Sprintf("%s@%d,%d", RenderTile(p.Tile), p.Coord.X, p.Coord.Y)
		}
		return "place " + strings.Join(parts, " ")
	case MoveExchange:
		if len(m.Exchange) == 0 {
			return "pass (forced)"
		}
		parts := make([]string, len(m.Exchange))
		for i, t := range m.Exchange {
			parts[i] = RenderTile(t)
		}
		return "exchange " + strings.Join(parts, " ")
	}
	return "(unknown move)"
}
