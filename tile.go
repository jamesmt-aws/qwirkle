package main

import (
	"fmt"
	"strings"
)

type Color uint8
type Shape uint8

const (
	Red Color = iota
	Orange
	Yellow
	Green
	Blue
	Purple
	NumColors = 6
)

const (
	Circle Shape = iota
	Square
	Diamond
	Clover
	Star
	Cross
	NumShapes = 6
)

const TilesPerKind = 3

type Tile struct {
	Color Color
	Shape Shape
}

func AllTiles() []Tile {
	out := make([]Tile, 0, int(NumColors)*int(NumShapes)*TilesPerKind)
	for c := Color(0); c < NumColors; c++ {
		for s := Shape(0); s < NumShapes; s++ {
			for k := 0; k < TilesPerKind; k++ {
				out = append(out, Tile{c, s})
			}
		}
	}
	return out
}

var colorLetter = [NumColors]byte{Red: 'R', Orange: 'O', Yellow: 'Y', Green: 'G', Blue: 'B', Purple: 'P'}
var shapeGlyph = [NumShapes]byte{Circle: 'o', Square: 's', Diamond: 'd', Clover: 'c', Star: '*', Cross: '+'}

// shapeDisplay is the 2-cell-wide glyph used in the rendered board/hand.
// Each entry is a Unicode shape chosen to resemble the corresponding
// Qwirkle tile, doubled so the cell occupies 2 columns.
//
// "Cross" historically named the 8-point starburst tile in this code;
// "Star" is the 4-point sparkle; "Clover" is the 4-prong clover/club.
var shapeDisplay = [NumShapes]string{
	Circle:  "●●", // ●●
	Square:  "■■", // ■■
	Diamond: "◆◆", // ◆◆
	Clover:  "✤✤", // ✤✤  4-leaf clover / 4-prong club
	Star:    "✦✦", // ✦✦  4-pointed sparkle
	Cross:   "❋❋", // ❋❋  8-point starburst
}

// Code returns the canonical 2-char identifier for the tile (color
// letter + shape letter). Used for parsing, tests, and debug output.
func (t Tile) Code() string {
	return string([]byte{colorLetter[t.Color], shapeGlyph[t.Shape]})
}

// Glyph returns the 2-char display glyph for the tile's shape. Display
// code colors the glyph using the tile's color.
func (t Tile) Glyph() string { return shapeDisplay[t.Shape] }

func (t Tile) String() string { return t.Code() }

func ParseTile(s string) (Tile, error) {
	s = strings.TrimSpace(s)
	if len(s) != 2 {
		return Tile{}, fmt.Errorf("tile %q must be 2 chars (color+shape), e.g. Ro", s)
	}
	cb := s[0]
	if cb >= 'a' && cb <= 'z' {
		cb -= 'a' - 'A'
	}
	var c Color
	switch cb {
	case 'R':
		c = Red
	case 'O':
		c = Orange
	case 'Y':
		c = Yellow
	case 'G':
		c = Green
	case 'B':
		c = Blue
	case 'P':
		c = Purple
	default:
		return Tile{}, fmt.Errorf("unknown color %q (use R/O/Y/G/B/P)", string(cb))
	}
	sb := s[1]
	if sb >= 'A' && sb <= 'Z' {
		sb += 'a' - 'A'
	}
	var sh Shape
	switch sb {
	case 'o':
		sh = Circle
	case 's':
		sh = Square
	case 'd':
		sh = Diamond
	case 'c':
		sh = Clover
	case '*':
		sh = Star
	case '+':
		sh = Cross
	default:
		return Tile{}, fmt.Errorf("unknown shape %q (use o/s/d/c/*/+)", string(sb))
	}
	return Tile{c, sh}, nil
}
