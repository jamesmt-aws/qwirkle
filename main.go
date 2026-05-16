package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	seed := flag.Int64("seed", time.Now().UnixNano(), "RNG seed")
	mode := flag.String("mode", "vs", "game mode: vs (human vs bot) | bot (bot vs bot) | hotseat (human vs human)")
	you := flag.String("name", "You", "your display name")
	flag.Parse()

	var names []string
	switch *mode {
	case "vs":
		names = []string{*you, "Bot"}
	case "bot":
		names = []string{"Bot-1", "Bot-2"}
	case "hotseat":
		names = []string{"P1", "P2"}
	default:
		fmt.Fprintf(os.Stderr, "unknown mode %q\n", *mode)
		os.Exit(2)
	}

	g := NewGame(*seed, names)
	bots := make([]Bot, len(names))
	switch *mode {
	case "vs":
		bots[1] = NewGreedyBot(*seed + 1)
	case "bot":
		bots[0] = NewGreedyBot(*seed + 1)
		bots[1] = NewGreedyBot(*seed + 2)
	}

	fmt.Printf("Qwirkle — seed %d, mode %s\n", *seed, *mode)
	in := bufio.NewScanner(os.Stdin)
	in.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	consecutivePasses := 0

	for !g.Finished() {
		bot := bots[g.Current()]
		printState(g, "", bot == nil)
		var m Move
		if bot != nil {
			m = bot.Choose(g)
			fmt.Printf("%s plays: %s\n", g.Names[g.Current()], RenderMove(m))
		} else {
			var ok bool
			m, ok = readHumanMove(in, g, bots)
			if !ok {
				fmt.Println("(quitting)")
				return
			}
		}
		if m.Type == MoveExchange && len(m.Exchange) == 0 {
			consecutivePasses++
			fmt.Printf("%s passes.\n", g.Names[g.Current()])
			if consecutivePasses >= g.NumPlayers {
				fmt.Println("All players passed in a row — ending game.")
				break
			}
			g.current = (g.current + 1) % g.NumPlayers
			continue
		}
		consecutivePasses = 0
		if err := g.Apply(m); err != nil {
			fmt.Printf("illegal move: %v\n", err)
			continue
		}
	}

	fmt.Println()
	printState(g, "Final:", false)
	w := g.Winner()
	if w < 0 {
		fmt.Println("Tied!")
	} else {
		fmt.Printf("%s wins!\n", g.Names[w])
	}
}

func printState(g *Game, banner string, showHand bool) {
	fmt.Println()
	if banner != "" {
		fmt.Println(ansiBold + banner + ansiReset)
	}
	fmt.Print(RenderBoard(g.Board(), 1))
	fmt.Println()
	fmt.Print(RenderScoreboard(g))
	if showHand {
		fmt.Printf("Hand (%s): %s\n", g.Names[g.Current()], RenderHand(g.Hand(g.Current())))
	}
}

func readHumanMove(in *bufio.Scanner, g *Game, bots []Bot) (Move, bool) {
	for {
		fmt.Print("\n> ")
		if !in.Scan() {
			return Move{}, false
		}
		line := strings.TrimSpace(in.Text())
		if line == "" {
			printState(g, "", bots[g.Current()] == nil)
			continue
		}
		fields := strings.Fields(line)
		cmd := strings.ToLower(fields[0])
		switch cmd {
		case "q", "quit", "exit":
			return Move{}, false
		case "h", "help", "?":
			printHelp()
		case "r", "render":
			printState(g, "", bots[g.Current()] == nil)
		case "hint":
			showHint(g)
		case "moves":
			showAllMoves(g)
		case "p", "place":
			m, err := parsePlace(fields[1:])
			if err != nil {
				fmt.Println("error:", err)
				continue
			}
			if _, err := ValidatePlace(g.Board(), m.Placements); err != nil {
				fmt.Println("illegal:", err)
				continue
			}
			if err := checkSubset(g.Hand(g.Current()), placementTiles(m.Placements)); err != nil {
				fmt.Println("you don't have:", err)
				continue
			}
			return m, true
		case "x", "exchange":
			m, err := parseExchange(fields[1:])
			if err != nil {
				fmt.Println("error:", err)
				continue
			}
			if g.BagRemaining() == 0 {
				fmt.Println("bag is empty — cannot exchange")
				continue
			}
			if err := checkSubset(g.Hand(g.Current()), m.Exchange); err != nil {
				fmt.Println("you don't have:", err)
				continue
			}
			return m, true
		case "pass":
			return Move{Type: MoveExchange, Exchange: nil}, true
		default:
			fmt.Println("unknown command — type ? for help")
		}
	}
}

func parsePlace(args []string) (Move, error) {
	if len(args) == 0 {
		return Move{}, fmt.Errorf("usage: p <tile>@<x>,<y> [<tile>@<x>,<y> ...]")
	}
	ps := make([]Placement, 0, len(args))
	for _, a := range args {
		at := strings.Index(a, "@")
		if at < 0 {
			return Move{}, fmt.Errorf("placement %q missing @", a)
		}
		tile, err := ParseTile(a[:at])
		if err != nil {
			return Move{}, err
		}
		comma := strings.Index(a[at+1:], ",")
		if comma < 0 {
			return Move{}, fmt.Errorf("placement %q missing comma in coord", a)
		}
		xs := a[at+1 : at+1+comma]
		ys := a[at+2+comma:]
		x, err := strconv.Atoi(xs)
		if err != nil {
			return Move{}, fmt.Errorf("bad X in %q: %v", a, err)
		}
		y, err := strconv.Atoi(ys)
		if err != nil {
			return Move{}, fmt.Errorf("bad Y in %q: %v", a, err)
		}
		ps = append(ps, Placement{Coord{x, y}, tile})
	}
	return Move{Type: MovePlace, Placements: ps}, nil
}

func parseExchange(args []string) (Move, error) {
	if len(args) == 0 {
		return Move{}, fmt.Errorf("usage: x <tile> [<tile> ...]")
	}
	ts := make([]Tile, 0, len(args))
	for _, a := range args {
		t, err := ParseTile(a)
		if err != nil {
			return Move{}, err
		}
		ts = append(ts, t)
	}
	return Move{Type: MoveExchange, Exchange: ts}, nil
}

func showHint(g *Game) {
	bot := NewGreedyBot(0)
	m := bot.Choose(g)
	fmt.Printf("greedy suggests: %s\n", RenderMove(m))
}

func showAllMoves(g *Game) {
	moves := LegalPlacements(g.Board(), g.Hand(g.Current()))
	if len(moves) == 0 {
		fmt.Println("no legal placements")
		return
	}
	type scored struct {
		m Move
		s int
	}
	all := make([]scored, 0, len(moves))
	for _, m := range moves {
		r, err := ValidatePlace(g.Board(), m.Placements)
		if err != nil {
			continue
		}
		all = append(all, scored{m, r.Score})
	}
	// print top 10 by score
	for i := 0; i < len(all); i++ {
		for j := i + 1; j < len(all); j++ {
			if all[j].s > all[i].s {
				all[i], all[j] = all[j], all[i]
			}
		}
	}
	n := len(all)
	if n > 10 {
		n = 10
	}
	for i := 0; i < n; i++ {
		fmt.Printf("  [%d] %s\n", all[i].s, RenderMove(all[i].m))
	}
	if len(all) > n {
		fmt.Printf("  ... %d more\n", len(all)-n)
	}
}

func printHelp() {
	fmt.Println(`commands:
  p <tile>@<x>,<y> ...   place tiles. example: p Ro@0,0 Rs@1,0
  x <tile> ...           exchange tiles. example: x Ro Bs
  pass                   skip turn (only legal if bag is empty)
  moves                  show top legal placements
  hint                   show what the greedy bot would play
  r                      re-render the board
  ? / help               this message
  q                      quit

tile codes: 2 chars, color + shape.
  colors: R O Y G B P    (red orange yellow green blue purple)
  shapes: o s d c * +    (circle square diamond clover star cross)
example tiles: Ro = red circle, Y* = yellow star, B+ = blue cross`)
}
