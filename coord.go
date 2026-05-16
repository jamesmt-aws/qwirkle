package main

type Coord struct{ X, Y int }

func (c Coord) Add(d Coord) Coord { return Coord{c.X + d.X, c.Y + d.Y} }
func (c Coord) Sub(d Coord) Coord { return Coord{c.X - d.X, c.Y - d.Y} }

// Direction unit vectors. Y is graph-oriented: +Y is up on screen, -Y is down.
var (
	Right = Coord{1, 0}
	Left  = Coord{-1, 0}
	Up    = Coord{0, 1}
	Down  = Coord{0, -1}
)

var axes = []Coord{Right, Down}
var neighbors4 = []Coord{Right, Left, Down, Up}
