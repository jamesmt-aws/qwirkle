package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

func main() {
	seed := flag.Int64("seed", time.Now().UnixNano(), "RNG seed")
	mode := flag.String("mode", "vs", "game mode: vs (human vs bot) | bot (bot vs bot) | hotseat (human vs human)")
	you := flag.String("name", "You", "your display name")
	size := flag.Int("size", 52, "minimum board view in cells per axis (TUI modes)")
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

	if *mode == "bot" {
		runBotSpectator(g, bots)
		return
	}
	if err := RunTUI(g, bots, *size); err != nil {
		fmt.Fprintln(os.Stderr, "tui error:", err)
		os.Exit(1)
	}
}

func runBotSpectator(g *Game, bots []Bot) {
	consecutivePasses := 0
	for !g.Finished() {
		printState(g, "", false)
		bot := bots[g.Current()]
		m := bot.Choose(g)
		fmt.Printf("%s plays: %s\n", g.Names[g.Current()], RenderMove(m))
		if m.Type == MoveExchange && len(m.Exchange) == 0 {
			consecutivePasses++
			if consecutivePasses >= g.NumPlayers {
				fmt.Println("all players passed; ending")
				break
			}
			g.current = (g.current + 1) % g.NumPlayers
			continue
		}
		consecutivePasses = 0
		if err := g.Apply(m); err != nil {
			fmt.Println("error:", err)
			break
		}
	}
	fmt.Println()
	printState(g, "Final:", false)
	if w := g.Winner(); w < 0 {
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
