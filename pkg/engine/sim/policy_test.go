package sim

import (
	"bytes"
	"reflect"
	"sort"
	"testing"

	"animus-mundi/pkg/engine"
)

// Tests for T-0022: ProgressPolicy and SimOptions.Policy. Each
// TestProgress_Case<N> is Acceptance Case N (cases 6 and 7 are CLI checks).

func progressGame(t *testing.T, decks [2]engine.DeckList, seed uint64) *engine.Game {
	t.Helper()
	g, err := engine.NewGame(engine.Config{Seed: seed, Decks: decks,
		Choosers: [2]engine.Chooser{engine.RandomChooser(seed*2 + 1), engine.RandomChooser(seed*2 + 2)}})
	if err != nil {
		t.Fatal(err)
	}
	return g
}

func TestProgress_Case1_LegalOnly(t *testing.T) {
	decks := simStarterDecks(t)
	for i := range uint64(50) {
		s := 100 + i
		inner := [2]Policy{ProgressPolicy(s*2 + 1), ProgressPolicy(s*2 + 2)}
		var wrapped [2]Policy
		for seat := range wrapped {
			wrapped[seat] = func(g *engine.Game, legal []engine.Action) engine.Action {
				a := inner[seat](g, legal)
				for _, l := range legal {
					if l == a {
						return a
					}
				}
				t.Fatalf("seed %d: policy returned %+v, not in the legal slice", s, a)
				return a
			}
		}
		if _, err := Playout(progressGame(t, decks, s), wrapped, 400); err != nil {
			t.Fatalf("seed %d: %v", s, err)
		}
	}
}

func TestProgress_Case2_Preferences(t *testing.T) {
	for seed := uint64(1); seed <= 8; seed++ {
		// An affordable card and a readied asset: Deploy.
		s := newScene(t)
		s.g.Players[0].Hand = nil
		s.give(0, unit("Free", 1, 1, engine.Keywords{}))
		s.put(0, engine.ZoneReserves, unit("Worker", 1, 1, engine.Keywords{}))
		if got := ProgressPolicy(seed)(s.g, s.g.LegalActions()); got.Kind != engine.ActDeploy {
			t.Errorf("seed %d: affordable card in hand, got %s, want Deploy", seed, got.Kind)
		}

		// Nothing affordable, a 2R card in hand, pool 0: Generate Requisition.
		s = newScene(t)
		s.g.Players[0].Hand = nil
		costly := unit("Costly", 1, 1, engine.Keywords{})
		costly.ReqCost = 2
		s.give(0, costly)
		s.g.Players[0].Req, s.g.Players[0].Tac = 0, 0
		s.put(0, engine.ZoneReserves, unit("Worker", 1, 1, engine.Keywords{}))
		got := ProgressPolicy(seed)(s.g, s.g.LegalActions())
		if got.Kind != engine.ActGenerate || got.Resource != engine.Requisition {
			t.Errorf("seed %d: short on engine.Requisition, got %+v, want Generate Requisition", seed, got)
		}

		// Empty hand, every asset exhausted: EndTurn.
		s = newScene(t)
		s.g.Players[0].Hand = nil
		s.g.Players[0].Req, s.g.Players[0].Tac = 0, 0
		for _, a := range s.g.Assets() {
			a.Exhausted = true
		}
		if got := ProgressPolicy(seed)(s.g, s.g.LegalActions()); got.Kind != engine.ActEndTurn {
			t.Errorf("seed %d: nothing to do, got %s, want EndTurn", seed, got.Kind)
		}
	}
}

func progressRun(t *testing.T) (SimStats, []byte) {
	t.Helper()
	var cols []*gameCollector
	opt := SimOptions{DeckNames: simDeckNames, Policy: ProgressPolicy,
		Observer: func(game int, s uint64) engine.Observer {
			c := &gameCollector{}
			cols = append(cols, c)
			return c
		}}
	st, err := SimulateWith(simStarterDecks(t), 20, 7, 400, opt)
	if err != nil {
		t.Fatal(err)
	}
	games := make([][]engine.Event, len(cols))
	for i, c := range cols {
		games[i] = c.events
	}
	return st, jsonlBytes(t, games)
}

func TestProgress_Case3_Deterministic(t *testing.T) {
	st1, b1 := progressRun(t)
	st2, b2 := progressRun(t)
	if !reflect.DeepEqual(st1, st2) {
		t.Errorf("stats differ between runs: %+v vs %+v", st1, st2)
	}
	if !bytes.Equal(b1, b2) {
		t.Error("event output differs between runs")
	}
}

func TestProgress_Case4_Coverage(t *testing.T) {
	decks := simStarterDecks(t)
	counts := map[string]int{}
	opt := SimOptions{DeckNames: simDeckNames, Policy: ProgressPolicy,
		Observer: func(game int, s uint64) engine.Observer { return &deployCounter{counts} }}
	if _, err := SimulateWith(decks, 200, 1, 400, opt); err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{}
	for _, d := range decks {
		for _, c := range d.Main {
			if c.Supported() {
				want[c.Name] = true
			}
		}
	}
	var names []string
	for n := range want {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		t.Logf("%-40s %d deploys", n, counts[n])
		if counts[n] < 10 {
			t.Errorf("%s deployed %d times in 200 games, want at least 10", n, counts[n])
		}
	}
}

type deployCounter struct{ counts map[string]int }

func (c *deployCounter) Event(e engine.Event) {
	if e.Kind == "deploy" {
		c.counts[e.Card]++
	}
}
func (c *deployCounter) Settled(s engine.Snapshot) {}

func TestProgress_Case5_TerminatesAndHolds(t *testing.T) {
	opt := SimOptions{Policy: ProgressPolicy}
	st, err := SimulateWith(simStarterDecks(t), 200, 1, 400, opt)
	if err != nil {
		t.Fatalf("starter decks: %v", err)
	}
	if st.Games != 200 {
		t.Errorf("starter: played %d games, want 200", st.Games)
	}
	cards, err := engine.LoadCards("../../../data/cards")
	if err != nil {
		t.Fatal(err)
	}
	var stress [2]engine.DeckList
	for i, f := range []string{"SOULLESS", "CONCORDIA"} {
		if stress[i], err = FactionDeck(cards, f); err != nil {
			t.Fatal(err)
		}
	}
	st, err = SimulateWith(stress, 200, 1, 400, opt)
	if err != nil {
		t.Fatalf("stress decks: %v", err)
	}
	if st.Games != 200 {
		t.Errorf("stress: played %d games, want 200", st.Games)
	}
}
