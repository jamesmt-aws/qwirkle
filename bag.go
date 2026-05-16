package main

import "math/rand"

type Bag struct {
	tiles []Tile
	rng   *rand.Rand
}

func NewBag(seed int64) *Bag {
	b := &Bag{tiles: AllTiles(), rng: rand.New(rand.NewSource(seed))}
	b.shuffle()
	return b
}

func (b *Bag) shuffle() {
	b.rng.Shuffle(len(b.tiles), func(i, j int) { b.tiles[i], b.tiles[j] = b.tiles[j], b.tiles[i] })
}

func (b *Bag) Remaining() int { return len(b.tiles) }

func (b *Bag) Draw(n int) []Tile {
	if n > len(b.tiles) {
		n = len(b.tiles)
	}
	out := make([]Tile, n)
	copy(out, b.tiles[:n])
	b.tiles = b.tiles[n:]
	return out
}

// Exchange takes `returned` tiles back into the bag and returns the same
// number of freshly drawn tiles. The exchanged tiles can't come back in the
// same call: we draw replacements first, then add the returned tiles and
// reshuffle.
func (b *Bag) Exchange(returned []Tile) []Tile {
	n := len(returned)
	if n > len(b.tiles) {
		n = len(b.tiles)
	}
	drawn := make([]Tile, n)
	copy(drawn, b.tiles[:n])
	b.tiles = b.tiles[n:]
	b.tiles = append(b.tiles, returned...)
	b.shuffle()
	return drawn
}
