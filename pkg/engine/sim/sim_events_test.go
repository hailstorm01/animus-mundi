package sim

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"animus-mundi/pkg/engine"
	"animus-mundi/pkg/engine/eventlog"
)

// Tests for T-0021: SimulateWith, FormatEvent and WriteEventsJSONL. Each
// TestSimEvents_Case<N> is Acceptance Case N.

// gameCollector keeps one game's events.
type gameCollector struct{ events []engine.Event }

func (c *gameCollector) Event(e engine.Event)      { c.events = append(c.events, e) }
func (c *gameCollector) Settled(s engine.Snapshot) {}

var simDeckNames = [2]string{"data/decks/animus-starter.txt", "data/decks/path-starter.txt"}

func simStarterDecks(t *testing.T) [2]engine.DeckList {
	t.Helper()
	cards, err := engine.LoadCards(filepath.Join("..", "..", "..", "data", "cards"))
	if err != nil {
		t.Fatalf("load cards: %v", err)
	}
	decks, err := SimDecks(cards, filepath.Join("..", "..", "..", "data", "decks"), "starter", [2]string{"ANIMUS", "PATH"})
	if err != nil {
		t.Fatalf("SimDecks: %v", err)
	}
	return decks
}

// collectGames plays games and returns each game's events, game 1 first.
func collectGames(t *testing.T, games int, seed uint64) ([][]engine.Event, SimStats) {
	t.Helper()
	var cols []*gameCollector
	opt := SimOptions{DeckNames: simDeckNames, Observer: func(game int, s uint64) engine.Observer {
		if game != len(cols)+1 {
			t.Errorf("observer asked for game %d, want %d", game, len(cols)+1)
		}
		if s != seed+uint64(game-1) {
			t.Errorf("game %d seed %d, want %d", game, s, seed+uint64(game-1))
		}
		c := &gameCollector{}
		cols = append(cols, c)
		return c
	}}
	st, err := SimulateWith(simStarterDecks(t), games, seed, 400, opt)
	if err != nil {
		t.Fatalf("SimulateWith: %v", err)
	}
	out := make([][]engine.Event, len(cols))
	for i, c := range cols {
		out[i] = c.events
	}
	return out, st
}

func jsonlBytes(t *testing.T, games [][]engine.Event) []byte {
	t.Helper()
	var buf bytes.Buffer
	for i, evs := range games {
		if err := eventlog.WriteEventsJSONL(&buf, i+1, evs); err != nil {
			t.Fatalf("WriteEventsJSONL: %v", err)
		}
	}
	return buf.Bytes()
}

func parseJSONL(t *testing.T, b []byte) []map[string]any {
	t.Helper()
	var out []map[string]any
	for _, line := range strings.Split(strings.TrimSuffix(string(b), "\n"), "\n") {
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Fatalf("line is not a JSON object: %v: %.80s", err, line)
		}
		out = append(out, m)
	}
	return out
}

func TestSimEvents_Case1_JSONLShape(t *testing.T) {
	games, _ := collectGames(t, 3, 11)
	if len(games) != 3 {
		t.Fatalf("observers created for %d games, want 3", len(games))
	}
	lines := parseJSONL(t, jsonlBytes(t, games))
	prevGame, prevSeq := 0, 0
	var kinds []string
	for i, m := range lines {
		g, gok := m["game"].(float64)
		s, sok := m["seq"].(float64)
		if !gok || !sok || g != float64(int(g)) || s != float64(int(s)) {
			t.Fatalf("line %d: game and seq must be integers: %v", i, m)
		}
		game, seq := int(g), int(s)
		switch {
		case game == prevGame:
			if seq != prevSeq+1 {
				t.Fatalf("line %d: seq %d after %d", i, seq, prevSeq)
			}
		case game == prevGame+1:
			if i > 0 && kinds[len(kinds)-1] != "match-end" {
				t.Errorf("game %d did not end with match-end", prevGame)
			}
			if seq != 1 || m["kind"] != "match-start" {
				t.Fatalf("game %d must open with seq 1 match-start, got seq %d kind %v", game, seq, m["kind"])
			}
			start := m["start"].(map[string]any)
			if want := float64(11 + game - 1); start["seed"] != want {
				t.Errorf("game %d seed %v, want %v", game, start["seed"], want)
			}
			decks := start["decks"].([]any)
			if decks[0] != simDeckNames[0] || decks[1] != simDeckNames[1] {
				t.Errorf("game %d deck names %v", game, decks)
			}
		default:
			t.Fatalf("line %d: game %d after %d", i, game, prevGame)
		}
		prevGame, prevSeq = game, seq
		kinds = append(kinds, m["kind"].(string))
	}
	if prevGame != 3 || kinds[len(kinds)-1] != "match-end" {
		t.Errorf("last game %d, last kind %s", prevGame, kinds[len(kinds)-1])
	}
	if !strings.HasPrefix(string(jsonlBytes(t, games)), `{"game":1,"seq":1,`) {
		t.Errorf("the game key must lead each object")
	}
}

func TestSimEvents_Case2_OneTurnStartPerTurn(t *testing.T) {
	games, _ := collectGames(t, 3, 11)
	lines := parseJSONL(t, jsonlBytes(t, games))
	turns := map[int][]int{}
	endTurns := map[int]int{}
	for _, m := range lines {
		g := int(m["game"].(float64))
		switch m["kind"] {
		case "turn-start":
			turns[g] = append(turns[g], int(m["turn"].(float64)))
		case "match-end":
			endTurns[g] = int(m["end"].(map[string]any)["turns"].(float64))
		}
	}
	for g := 1; g <= 3; g++ {
		if len(turns[g]) != endTurns[g] {
			t.Errorf("game %d: %d turn-start lines, match-end turns %d", g, len(turns[g]), endTurns[g])
		}
		for i, turn := range turns[g] {
			if turn != i+1 {
				t.Errorf("game %d: turn-start %d has turn %d", g, i, turn)
			}
		}
	}
}

func TestSimEvents_Case3_ByteIdentical(t *testing.T) {
	a, _ := collectGames(t, 3, 11)
	b, _ := collectGames(t, 3, 11)
	c, _ := collectGames(t, 3, 12)
	if !bytes.Equal(jsonlBytes(t, a), jsonlBytes(t, b)) {
		t.Errorf("same seed gave different bytes")
	}
	if bytes.Equal(jsonlBytes(t, a), jsonlBytes(t, c)) {
		t.Errorf("different seed gave the same bytes")
	}
}

func TestSimEvents_Case4_ObservingChangesNothing(t *testing.T) {
	_, observed := collectGames(t, 5, 21)
	plain, err := Simulate(simStarterDecks(t), 5, 21, 400)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(observed, plain) {
		t.Errorf("observed %+v != plain %+v", observed, plain)
	}
}

func TestSimEvents_Case5_FormatEvent(t *testing.T) {
	games, _ := collectGames(t, 1, 7)
	sawDecision, sawStart := false, false
	for _, e := range games[0] {
		line := eventlog.FormatEvent(e)
		if strings.Contains(line, "\n") {
			t.Errorf("seq %d: line has a newline: %q", e.Seq, line)
		}
		if !strings.HasPrefix(line, strconv.Itoa(e.Seq)) {
			t.Errorf("seq %d: line does not start with Seq: %q", e.Seq, line)
		}
		if !strings.Contains(line, e.Kind) {
			t.Errorf("seq %d: line lacks kind %q: %q", e.Seq, e.Kind, line)
		}
		switch {
		case e.Decision != nil && len(e.Decision.Picks) > 0:
			sawDecision = true
			label := e.Decision.Options[e.Decision.Picks[0]]
			if !strings.Contains(line, label) || !strings.Contains(line, e.Decision.How) {
				t.Errorf("seq %d: decision line lacks label %q or how %q: %q", e.Seq, label, e.Decision.How, line)
			}
		case e.Start != nil:
			sawStart = true
			for _, want := range []string{e.Start.HQs[0], e.Start.HQs[1], strconv.FormatUint(e.Start.Seed, 10)} {
				if !strings.Contains(line, want) {
					t.Errorf("match-start line lacks %q: %q", want, line)
				}
			}
		}
	}
	if !sawStart {
		t.Errorf("game had no match-start")
	}
	// The starter decks make no decisions under random play (sawDecision stays
	// false for them), so check the decision line on a constructed event too.
	_ = sawDecision
	d := engine.Event{Seq: 12, Turn: 2, Phase: "main", Kind: "decision", Player: 1, Decision: &engine.DecisionRecord{
		Kind: "Target", Player: 1, Prompt: "pick a target", Options: []string{"Alpha", "Beta"},
		Refs: []int{4, 5}, Picks: []int{1}, How: "answer"}}
	line := eventlog.FormatEvent(d)
	if !strings.HasPrefix(line, "12") || !strings.Contains(line, "decision") ||
		!strings.Contains(line, "Beta") || !strings.Contains(line, "answer") || strings.Contains(line, "Alpha") {
		t.Errorf("decision line wrong: %q", line)
	}
}
