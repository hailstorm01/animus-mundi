package sim

import (
	"bytes"
	"reflect"
	"slices"
	"sort"
	"testing"

	"animus-mundi/pkg/engine"
	"animus-mundi/pkg/engine/ability"
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

// drawManeuver is a synthetic Maneuver costing req Requisition and tac Tactics
// with the untargeted step "draw 1".
func drawManeuver(name string, req, tac int) *engine.Card {
	return &engine.Card{Name: name, Type: engine.TypeManeuver, ReqCost: req, TacCost: tac,
		Ability: &ability.Def{Effect: ability.Effect{Steps: []ability.Step{
			{Op: ability.OpDraw, Amount: ability.Value{N: 1}}}}}}
}

func TestProgress_Case6_EnactsManeuvers(t *testing.T) {
	deployed, enacted := 0, 0
	for seed := uint64(1); seed <= 16; seed++ {
		// Case 1: only an affordable Maneuver in hand: Enact it.
		s := newScene(t)
		s.g.Players[0].Hand = nil
		m := s.give(0, drawManeuver("Study", 0, 0))
		s.put(0, engine.ZoneReserves, unit("Worker", 1, 1, engine.Keywords{}))
		if seed <= 8 {
			got := ProgressPolicy(seed)(s.g, s.g.LegalActions())
			if got.Kind != engine.ActEnact || got.Card != m.ID {
				t.Errorf("seed %d: affordable Maneuver, got %+v, want Enact of it", seed, got)
			}
		}

		// Case 2: a unit and a Maneuver: both kinds are offered, nothing else.
		s = newScene(t)
		s.g.Players[0].Hand = nil
		s.give(0, unit("Free", 1, 1, engine.Keywords{}))
		s.give(0, drawManeuver("Study", 0, 0))
		s.put(0, engine.ZoneReserves, unit("Worker", 1, 1, engine.Keywords{}))
		got := ProgressPolicy(seed)(s.g, s.g.LegalActions())
		switch got.Kind {
		case engine.ActDeploy:
			deployed++
		case engine.ActEnact:
			enacted++
		default:
			t.Errorf("seed %d: unit and Maneuver in hand, got %s, want Deploy or Enact", seed, got.Kind)
		}

		// Case 3: a 2 Tactics Maneuver, pool 0/0: Generate Tactics. An
		// unsupported one closes nothing.
		if seed <= 8 {
			s = newScene(t)
			s.g.Players[0].Hand = nil
			s.give(0, drawManeuver("Costly", 0, 2))
			s.g.Players[0].Req, s.g.Players[0].Tac = 0, 0
			worker := unit("Worker", 1, 1, engine.Keywords{})
			worker.TacYield = 1
			s.put(0, engine.ZoneReserves, worker)
			got = ProgressPolicy(seed)(s.g, s.g.LegalActions())
			if got.Kind != engine.ActGenerate || got.Resource != engine.Tactics {
				t.Errorf("seed %d: short on Tactics for a Maneuver, got %+v, want Generate Tactics", seed, got)
			}
			s.g.Players[0].Hand[0].Card = &engine.Card{Name: "Dud", Type: engine.TypeManeuver, TacCost: 2}
			got = ProgressPolicy(seed)(s.g, s.g.LegalActions())
			if got.Kind == engine.ActGenerate || got.Kind == engine.ActEnact {
				t.Errorf("seed %d: unsupported Maneuver closes nothing, got %+v", seed, got)
			}
		}
	}
	if deployed == 0 || enacted == 0 {
		t.Errorf("over 16 seeds: %d Deploys, %d Enacts, want both at least once", deployed, enacted)
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
	if _, err := SimulateWith(decks, 1000, 1, 400, opt); err != nil {
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
		t.Logf("%-40s %d plays", n, counts[n])
		if counts[n] < 10 {
			t.Errorf("%s played %d times in 1000 games, want at least 10", n, counts[n])
		}
	}
}

// deployCounter counts plays: a deploy event, or an Enact action event, by card.
type deployCounter struct{ counts map[string]int }

func (c *deployCounter) Event(e engine.Event) {
	if e.Kind == "deploy" || (e.Kind == "action" && e.Detail == "Enact") {
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
	cards, err := engine.LoadCards("../../data/cards")
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

func TestProgress_ActivationIsAnAct(t *testing.T) {
	activator := func() *engine.Card {
		c := unit("Engine", 1, 3, engine.Keywords{})
		c.Ability = &ability.Def{Activated: []ability.Activated{{
			Costs:  []ability.Cost{{Kind: ability.CostResources, Req: 1}},
			Effect: ability.Effect{Steps: []ability.Step{{Op: ability.OpDraw, Amount: ability.Value{N: 1}}}},
		}}}
		return c
	}
	scene := func(withHand bool) *scene {
		s := newScene(t)
		s.g.Players[0].Hand = nil
		s.put(0, engine.ZoneReserves, activator())
		if withHand {
			s.give(0, unit("Free", 0, 1, engine.Keywords{}))
		}
		for _, a := range s.g.Assets() {
			a.Exhausted = true
		}
		s.g.Players[0].Req, s.g.Players[0].Tac = 1, 0
		return s
	}
	activations := 0
	for seed := uint64(1); seed <= 16; seed++ {
		s := scene(false)
		legal := s.g.LegalActions()
		got := ProgressPolicy(seed)(s.g, legal)
		if got.Kind != engine.ActActivate && got.Kind != engine.ActEndTurn {
			t.Errorf("seed %d: got %s, want Activate or EndTurn", seed, got.Kind)
		}
		if got.Kind == engine.ActActivate {
			activations++
		}
		if !slices.Contains(legal, got) {
			t.Errorf("seed %d: %+v is not in LegalActions", seed, got)
		}
		s = scene(true)
		if got := ProgressPolicy(seed)(s.g, s.g.LegalActions()); got.Kind != engine.ActDeploy {
			t.Errorf("seed %d: a play from hand comes first, got %s", seed, got.Kind)
		}
	}
	if activations == 0 {
		t.Errorf("the policy never activated over 16 seeds")
	}
}
