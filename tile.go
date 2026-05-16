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

func (t Tile) Code() string {
	return string([]byte{colorLetter[t.Color], shapeGlyph[t.Shape]})
}

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
