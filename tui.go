package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type tuiMode int

const (
	modePlay tuiMode = iota
	modeExchange
)

type tuiModel struct {
	g    *Game
	bots []Bot

	cursor   Coord
	selected int // -1 = none
	pending  []Placement

	mode     tuiMode
	exchange map[int]bool

	boardMin int // minimum board view dimension (cells per axis)
	info     string
	botMsg   string
	showHelp bool
}

type botActMsg struct{ move Move }

func RunTUI(g *Game, bots []Bot, boardMin int) error {
	m := tuiModel{
		g:        g,
		bots:     bots,
		selected: -1,
		exchange: map[int]bool{},
		boardMin: boardMin,
	}
	if !g.Board().IsEmpty() {
		min, _ := g.Board().Bounds()
		m.cursor = min
	}
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func (m tuiModel) Init() tea.Cmd { return m.maybeBotMove() }

func (m tuiModel) maybeBotMove() tea.Cmd {
	if m.g.Finished() {
		return nil
	}
	bot := m.bots[m.g.Current()]
	if bot == nil {
		return nil
	}
	g := m.g
	return func() tea.Msg {
		return botActMsg{move: bot.Choose(g)}
	}
}

func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case botActMsg:
		return m.handleBotMove(msg)
	case tea.KeyMsg:
		// Quit always honored.
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
		if m.g.Finished() {
			switch msg.String() {
			case "q", "enter", "esc":
				return m, tea.Quit
			}
			return m, nil
		}
		if m.bots[m.g.Current()] != nil {
			// Bot's turn; only quit honored.
			if msg.String() == "q" {
				return m, tea.Quit
			}
			return m, nil
		}
		return m.handleKey(msg)
	}
	return m, nil
}

func (m tuiModel) handleBotMove(msg botActMsg) (tea.Model, tea.Cmd) {
	mover := m.g.Current()
	move := msg.move
	if move.Type == MoveExchange && len(move.Exchange) == 0 {
		m.botMsg = fmt.Sprintf("%s passed (no legal play, bag empty)", m.g.Names[mover])
		m.g.current = (m.g.current + 1) % m.g.NumPlayers
		return m, m.maybeBotMove()
	}
	eval, _ := EvaluateMove(m.g, move)
	if err := m.g.Apply(move); err != nil {
		m.info = "bot error: " + err.Error()
		return m, nil
	}
	if move.Type == MovePlace {
		m.botMsg = fmt.Sprintf("%s played %s  (score %d, exposes ≤%d, net %+d)",
			m.g.Names[mover], RenderMove(move), eval.Score, eval.Ceiling, eval.Net)
	} else {
		m.botMsg = fmt.Sprintf("%s exchanged %d tile(s)", m.g.Names[mover], len(move.Exchange))
	}
	return m, m.maybeBotMove()
}

func (m tuiModel) handleKey(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.info = ""
	s := key.String()
	switch s {
	case "q":
		return m, tea.Quit
	case "?":
		m.showHelp = !m.showHelp
	case "up":
		m.cursor.Y--
	case "down":
		m.cursor.Y++
	case "left":
		m.cursor.X--
	case "right":
		m.cursor.X++
	case "tab":
		hand := m.g.Hand(m.g.Current())
		if len(hand) > 0 {
			m.selected = (m.selected + 1) % len(hand)
		}
	case "shift+tab":
		hand := m.g.Hand(m.g.Current())
		if len(hand) > 0 {
			m.selected = (m.selected - 1 + len(hand)) % len(hand)
		}
	case "1", "2", "3", "4", "5", "6":
		idx := int(s[0] - '1')
		hand := m.g.Hand(m.g.Current())
		if idx >= len(hand) {
			return m, nil
		}
		if m.mode == modeExchange {
			if m.exchange[idx] {
				delete(m.exchange, idx)
			} else {
				m.exchange[idx] = true
			}
		} else {
			m.selected = idx
		}
	case "enter":
		if m.mode == modeExchange {
			return m.commitExchange()
		}
		return m.placeAtCursor()
	case "backspace":
		if m.mode == modeExchange {
			m.exchange = map[int]bool{}
		} else if len(m.pending) > 0 {
			m.pending = m.pending[:len(m.pending)-1]
		}
	case "esc":
		m.pending = nil
		m.exchange = map[int]bool{}
		m.mode = modePlay
		m.selected = -1
	case " ":
		if m.mode == modeExchange {
			return m.commitExchange()
		}
		return m.commitPending()
	case "x":
		m.mode = modeExchange
		m.pending = nil
		m.selected = -1
	case "p":
		if m.g.BagRemaining() > 0 || len(LegalPlacements(m.g.Board(), m.g.Hand(m.g.Current()))) > 0 {
			m.info = "pass is only legal when the bag is empty and you have no legal play"
			return m, nil
		}
		m.botMsg = "you passed"
		m.g.current = (m.g.current + 1) % m.g.NumPlayers
		return m, m.maybeBotMove()
	case "h":
		return m.showHint()
	}
	return m, nil
}

func (m tuiModel) placeAtCursor() (tea.Model, tea.Cmd) {
	if m.selected < 0 {
		m.info = "select a tile first (1–6 or Tab)"
		return m, nil
	}
	hand := m.g.Hand(m.g.Current())
	tile := hand[m.selected]
	countInHand := 0
	for _, t := range hand {
		if t == tile {
			countInHand++
		}
	}
	usedInPending := 0
	for _, p := range m.pending {
		if p.Tile == tile {
			usedInPending++
		}
	}
	if usedInPending >= countInHand {
		m.info = "no more copies of that tile in hand"
		return m, nil
	}
	if _, occ := m.g.Board().At(m.cursor); occ {
		m.info = "cell is already occupied"
		return m, nil
	}
	for _, p := range m.pending {
		if p.Coord == m.cursor {
			m.info = "you already placed at that cell"
			return m, nil
		}
	}
	candidate := append(append([]Placement{}, m.pending...), Placement{m.cursor, tile})
	if _, err := ValidatePlace(m.g.Board(), candidate); err != nil {
		m.info = "illegal: " + err.Error()
		return m, nil
	}
	m.pending = candidate
	return m, nil
}

func (m tuiModel) commitPending() (tea.Model, tea.Cmd) {
	if len(m.pending) == 0 {
		m.info = "nothing pending — pick a tile and place it"
		return m, nil
	}
	move := Move{Type: MovePlace, Placements: append([]Placement{}, m.pending...)}
	eval, _ := EvaluateMove(m.g, move)
	if err := m.g.Apply(move); err != nil {
		m.info = "error: " + err.Error()
		return m, nil
	}
	m.botMsg = fmt.Sprintf("you played %s  (score %d, exposes ≤%d, net %+d)",
		RenderMove(move), eval.Score, eval.Ceiling, eval.Net)
	m.pending = nil
	m.selected = -1
	return m, m.maybeBotMove()
}

func (m tuiModel) commitExchange() (tea.Model, tea.Cmd) {
	if m.g.BagRemaining() == 0 {
		m.info = "bag is empty — cannot exchange"
		return m, nil
	}
	if len(m.exchange) == 0 {
		m.info = "select tiles to exchange first (1–6)"
		return m, nil
	}
	hand := m.g.Hand(m.g.Current())
	var tiles []Tile
	for i := range hand {
		if m.exchange[i] {
			tiles = append(tiles, hand[i])
		}
	}
	move := Move{Type: MoveExchange, Exchange: tiles}
	if err := m.g.Apply(move); err != nil {
		m.info = "exchange failed: " + err.Error()
		return m, nil
	}
	m.botMsg = fmt.Sprintf("you exchanged %d tile(s)", len(tiles))
	m.exchange = map[int]bool{}
	m.mode = modePlay
	m.selected = -1
	return m, m.maybeBotMove()
}

func (m tuiModel) showHint() (tea.Model, tea.Cmd) {
	bot := NewGreedyBot(0)
	move := bot.Choose(m.g)
	eval, err := EvaluateMove(m.g, move)
	if err != nil {
		m.info = "hint failed: " + err.Error()
		return m, nil
	}
	m.info = fmt.Sprintf("greedy: %s  (score %d, exposes ≤%d, net %+d)",
		RenderMove(move), eval.Score, eval.Ceiling, eval.Net)
	return m, nil
}

func (m tuiModel) View() string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "%sQwirkle%s  —  seed %d\n\n", ansiBold, ansiReset, m.g.Seed)

	sb.WriteString(RenderBoardTUI(m.g.Board(), 2, m.boardMin, m.pending, m.cursor))
	sb.WriteByte('\n')
	sb.WriteString(RenderScoreboard(m.g))

	if !m.g.Finished() && m.bots[m.g.Current()] == nil {
		sb.WriteByte('\n')
		sb.WriteString(renderHand(m.g.Hand(m.g.Current()), m.selected, m.mode, m.exchange, m.pending))
		sb.WriteByte('\n')
	}

	if m.mode == modeExchange {
		fmt.Fprintf(&sb, "\n%s[exchange]%s 1–6 toggles tiles, Enter/Space commits, Esc cancels\n", ansiBold, ansiReset)
	} else if len(m.pending) > 0 {
		move := Move{Type: MovePlace, Placements: m.pending}
		eval, err := EvaluateMove(m.g, move)
		if err == nil {
			fmt.Fprintf(&sb, "\nPending: score %d  exposes ≤%d  net %+d   (Space/Enter to commit, Bksp to undo, Esc to cancel)\n",
				eval.Score, eval.Ceiling, eval.Net)
		} else {
			fmt.Fprintf(&sb, "\nPending: %v\n", err)
		}
	}

	if m.botMsg != "" {
		fmt.Fprintf(&sb, "\n%s\n", m.botMsg)
	}
	if m.info != "" {
		fmt.Fprintf(&sb, "\n%s%s%s\n", ansiBold, m.info, ansiReset)
	}

	switch {
	case m.g.Finished():
		w := m.g.Winner()
		result := "Tied!"
		if w >= 0 {
			result = m.g.Names[w] + " wins!"
		}
		fmt.Fprintf(&sb, "\n%sGame over.%s %s  (q/Enter to quit)\n", ansiBold, ansiReset, result)
	case m.bots[m.g.Current()] != nil:
		sb.WriteString("\nbot is thinking... (q to quit)\n")
	case m.showHelp:
		sb.WriteString("\n" + helpText + "\n")
	default:
		sb.WriteString("\narrows: move  |  1-6: pick tile  |  Enter: place  |  Bksp: undo  |  Space: commit  |  x: exchange  |  h: hint  |  ?: help  |  q: quit\n")
	}

	return sb.String()
}

const helpText = `controls:
  arrows         move cursor
  1..6           select tile from hand (or, in exchange mode, toggle for exchange)
  Tab/Shift+Tab  cycle hand selection
  Enter          place selected tile at cursor (adds to pending)
                 in exchange mode: commit selected tiles
  Space          commit your move
  Backspace      undo last placement (or clear exchange selection)
  x              enter exchange mode
  Esc            cancel pending / leave exchange mode
  h              show what greedy would play
  p              pass (only when bag is empty and you have no legal play)
  ?              toggle help
  q / Ctrl+C     quit`

func renderHand(hand []Tile, selected int, mode tuiMode, exchange map[int]bool, pending []Placement) string {
	// Count pending uses per tile to dim "used up" entries.
	pendingUses := map[Tile]int{}
	for _, p := range pending {
		pendingUses[p.Tile]++
	}
	available := map[Tile]int{}
	for _, t := range hand {
		available[t]++
	}
	var sb strings.Builder
	sb.WriteString("Hand: ")
	for i, t := range hand {
		sb.WriteByte(' ')
		isSel := selected == i && mode == modePlay
		isEx := mode == modeExchange && exchange[i]
		isDim := pendingUses[t] >= available[t] && mode == modePlay
		sb.WriteString(renderHandTile(t, i+1, isSel, isEx, isDim))
	}
	return sb.String()
}

func renderHandTile(t Tile, num int, selected, exchangeMark, dim bool) string {
	codes := []string{}
	if selected {
		codes = append(codes, "7")
	}
	if exchangeMark {
		codes = append(codes, "4", "1")
	}
	if dim {
		codes = append(codes, "2")
	}
	tileStr := RenderTile(t)
	label := fmt.Sprintf("[%d]", num)
	if len(codes) == 0 {
		return label + tileStr
	}
	wrap := "\x1b[" + strings.Join(codes, ";") + "m"
	return wrap + label + tileStr + ansiReset
}
