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
