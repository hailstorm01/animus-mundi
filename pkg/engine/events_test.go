package engine

import (
	"fmt"
	"math/rand/v2"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"sync"
	"testing"
)

// Tests for T-0020: structured events, the read-only Observer, settled
// notifications and match records. Each TestEvents_Case<N> is Acceptance Case N.

// recorder is an Observer that keeps everything it is told. The optional hooks
// let a test act on each notification as it arrives.
type recorder struct {
	events    []Event
	settled   []Snapshot
	onEvent   func(e Event)
	onSettled func(s Snapshot)
}

func (r *recorder) Event(e Event) {
	r.events = append(r.events, e)
	if r.onEvent != nil {
		r.onEvent(e)
	}
}

func (r *recorder) Settled(s Snapshot) {
	r.settled = append(r.settled, s)
	if r.onSettled != nil {
		r.onSettled(s)
	}
}

var (
	starterOnce  sync.Once
	starterDecks [2]DeckList
	starterErr   error
)

func loadStarters(t *testing.T) [2]DeckList {
	t.Helper()
	starterOnce.Do(func() {
		cards, err := LoadCards(filepath.Join("..", "..", "data", "cards"))
		if err != nil {
			starterErr = err
			return
		}
		for i, name := range []string{"animus-starter.txt", "path-starter.txt"} {
			starterDecks[i], err = LoadDeck(filepath.Join("..", "..", "data", "decks", name), cards)
			if err != nil {
				starterErr = err
				return
			}
		}
	})
	if starterErr != nil {
		t.Fatal(starterErr)
	}
	return starterDecks
}

// starterGame builds a starter-deck match for seed. Each player sets aside the
// first two cards of the opening hand so the mulligan record has content.
func starterGame(t *testing.T, seed uint64, obs Observer) *Game {
	t.Helper()
	g, err := NewGame(Config{
		Seed:      seed,
		Decks:     loadStarters(t),
		DeckNames: [2]string{"ANIMUS", "PATH"},
		Observer:  obs,
		Mulligan: func(_ PlayerID, hand []*Instance) []int {
			return []int{hand[0].ID, hand[1].ID}
		},
		Choosers: [2]Chooser{RandomChooser(seed*2 + 1), RandomChooser(seed*2 + 2)},
	})
	if err != nil {
		t.Fatal(err)
	}
	return g
}

// playResult is what playStarter reports about a finished match.
type playResult struct {
	Outcome Outcome
	Turns   int
	Actions int
}

// playStarter plays g to completion with a seeded random pick per seat, using
// sim.RandomPolicy's rule (end the turn with probability 0.1, else a uniform
// non-EndTurn action). It cannot call sim, which imports engine. Like
// sim.Playout, it checks invariants after every action and fails on an
// invalid-answer event.
func playStarter(t *testing.T, g *Game, seed uint64) playResult {
	t.Helper()
	var rs [2]*rand.Rand
	for i, s := range [2]uint64{seed*2 + 1, seed*2 + 2} {
		rs[i] = rand.New(rand.NewPCG(s, ^s))
	}
	var res playResult
	for !g.Outcome.Over {
		if g.Turn > 400 {
			t.Fatalf("seed %d: match did not end within 400 turns", seed)
		}
		legal := g.LegalActions()
		r := rs[g.Active]
		a := legal[len(legal)-1] // LegalActions always ends with EndTurn
		if len(legal) > 1 && r.Float64() >= 0.1 {
			a = legal[r.IntN(len(legal)-1)]
		}
		logged := len(g.Log)
		if err := g.Apply(a); err != nil {
			t.Fatalf("seed %d: turn %d: Apply rejected an action LegalActions offered: %v", seed, g.Turn, err)
		}
		for _, e := range g.Log[logged:] {
			if e.Kind == "invalid-answer" {
				t.Fatalf("seed %d: turn %d after %s: a chooser answer was rejected: %s", seed, g.Turn, a.Kind, e.Detail)
			}
		}
		res.Actions++
		if err := g.CheckInvariants(); err != nil {
			t.Fatalf("seed %d: turn %d after %s: %v", seed, g.Turn, a.Kind, err)
		}
	}
	res.Outcome = g.Outcome
	res.Turns = g.Turn
	return res
}

func eventsOfKind(log []Event, kind string) []Event {
	var out []Event
	for _, e := range log {
		if e.Kind == kind {
			out = append(out, e)
		}
	}
	return out
}

// Case 1: every event is numbered 1, 2, 3 ... with no gaps.
func TestEvents_Case1_SeqHasNoGaps(t *testing.T) {
	g := starterGame(t, 7, nil)
	playStarter(t, g, 7)
	if len(g.Log) == 0 {
		t.Fatal("empty log")
	}
	for i, e := range g.Log {
		if e.Seq != i+1 {
			t.Fatalf("Log[%d].Seq = %d, want %d (kind %s)", i, e.Seq, i+1, e.Kind)
		}
	}
}

// Case 2: the match opens with match-start, closes with match-end and has one
// turn-start per turn.
func TestEvents_Case2_MatchRecords(t *testing.T) {
	const seed = 7
	g := starterGame(t, seed, nil)
	playStarter(t, g, seed)

	first := g.Log[0]
	if first.Kind != "match-start" || first.Start == nil {
		t.Fatalf("Log[0] = %q with Start %v, want match-start with Start set", first.Kind, first.Start)
	}
	st := first.Start
	if st.Seed != seed {
		t.Errorf("Start.Seed = %d, want %d", st.Seed, seed)
	}
	if st.Decider != 0 && st.Decider != 1 {
		t.Errorf("Start.Decider = %d, want 0 or 1", st.Decider)
	}
	if st.Decks != [2]string{"ANIMUS", "PATH"} {
		t.Errorf("Start.Decks = %v", st.Decks)
	}
	decks := loadStarters(t)
	for i := range 2 {
		if st.HQs[i] == "" || st.HQs[i] != decks[i].HQ.Name {
			t.Errorf("Start.HQs[%d] = %q, want %q", i, st.HQs[i], decks[i].HQ.Name)
		}
		if len(st.Mulligans[i]) != 2 {
			t.Errorf("Start.Mulligans[%d] = %v, want the two cards set aside", i, st.Mulligans[i])
		}
	}

	last := g.Log[len(g.Log)-1]
	if last.Kind != "match-end" || last.End == nil {
		t.Fatalf("last event = %q with End %v, want match-end with End set", last.Kind, last.End)
	}
	if len(eventsOfKind(g.Log, "match-end")) != 1 {
		t.Errorf("match-end must be emitted exactly once")
	}
	end := last.End
	if end.Draw != g.Outcome.Draw || end.Reason != g.Outcome.Reason || end.Turns != g.Turn {
		t.Errorf("End = %+v does not match Outcome %+v / Turn %d", *end, g.Outcome, g.Turn)
	}
	if end.Draw && end.Winner != -1 {
		t.Errorf("a draw has Winner -1, got %d", end.Winner)
	}
	if !end.Draw && end.Winner != int(g.Outcome.Winner) {
		t.Errorf("End.Winner = %d, want %d", end.Winner, g.Outcome.Winner)
	}
	if i := slices.IndexFunc(g.Log, func(e Event) bool { return e.Kind == "result" }); i < 0 || i >= len(g.Log)-1 {
		t.Errorf("match-end must follow the result event (result at %d of %d)", i, len(g.Log))
	}

	perTurn := map[int]int{}
	for _, e := range eventsOfKind(g.Log, "turn-start") {
		if e.Snapshot == nil || e.Snapshot.Turn != e.Turn {
			t.Fatalf("turn-start at turn %d has snapshot %+v", e.Turn, e.Snapshot)
		}
		perTurn[e.Turn]++
	}
	for turn := 1; turn <= g.Turn; turn++ {
		if perTurn[turn] != 1 {
			t.Errorf("turn %d has %d turn-start events, want 1", turn, perTurn[turn])
		}
	}
	if len(perTurn) != g.Turn {
		t.Errorf("turn-start events cover %d turns, game lasted %d", len(perTurn), g.Turn)
	}
}

// Case 3: a turn-start snapshot holds hidden information too.
func TestEvents_Case3_SnapshotIsComplete(t *testing.T) {
	const seed = 11
	var g *Game
	checked, faceDown := 0, 0
	check := func(s Snapshot) {
		t.Helper()
		for i, p := range g.Players {
			ps := s.Players[i]
			if len(ps.Deck) != len(p.Deck) || len(ps.Hand) != len(p.Hand) || len(ps.Discard) != len(p.Discard) {
				t.Fatalf("turn %d player %d: snapshot zones %d/%d/%d, live %d/%d/%d", s.Turn, i,
					len(ps.Deck), len(ps.Hand), len(ps.Discard), len(p.Deck), len(p.Hand), len(p.Discard))
			}
			for j, inst := range p.Deck {
				if ps.Deck[j].Inst != inst.ID || ps.Deck[j].Card != inst.Card.Name {
					t.Fatalf("turn %d player %d: deck[%d] = %+v, live %d %s", s.Turn, i, j, ps.Deck[j], inst.ID, inst.Card.Name)
				}
			}
			for j, inst := range p.Hand {
				if ps.Hand[j].Inst != inst.ID || ps.Hand[j].Card != inst.Card.Name {
					t.Fatalf("turn %d player %d: hand[%d] = %+v, live %d %s", s.Turn, i, j, ps.Hand[j], inst.ID, inst.Card.Name)
				}
			}
		}
		for _, as := range g.Assets() {
			var snap *AssetSnap
			for _, ps := range s.Players {
				if ps.HQ != nil && ps.HQ.ID == as.ID {
					snap = ps.HQ
				}
				for k := range ps.Reserves {
					if ps.Reserves[k].ID == as.ID {
						snap = &ps.Reserves[k]
					}
				}
			}
			for k := range s.Theater {
				if s.Theater[k].ID == as.ID {
					snap = &s.Theater[k]
				}
			}
			if snap == nil {
				t.Fatalf("turn %d: asset %s is missing from the snapshot", s.Turn, as)
			}
			if snap.Card != as.Inst.Card.Name || snap.FaceDown != as.FaceDown || snap.Inst != as.Inst.ID {
				t.Fatalf("turn %d: asset snapshot %+v does not match %s (real card %s)", s.Turn, *snap, as, as.Inst.Card.Name)
			}
			if as.FaceDown {
				faceDown++
			}
		}
		checked++
	}
	obs := &recorder{}
	obs.onEvent = func(e Event) {
		if e.Kind == "turn-start" && g != nil {
			check(*e.Snapshot)
		}
	}
	g = starterGame(t, seed, obs)
	// Turn 1 was emitted inside NewGame, before g existed; the first player does
	// not draw, so the live state still equals that snapshot.
	for _, e := range obs.events {
		if e.Kind == "turn-start" {
			check(*e.Snapshot)
			break
		}
	}
	playStarter(t, g, seed)
	if checked < 3 {
		t.Fatalf("only %d snapshots checked", checked)
	}
	if faceDown == 0 {
		t.Errorf("seed %d never had a face-down asset at a turn start", seed)
	}
}

// Case 4: action events cover every ActionKind.
func TestEvents_Case4_EveryActionKind(t *testing.T) {
	seen := map[string]bool{}
	for seed := uint64(1); seed <= 10; seed++ {
		g := starterGame(t, seed, nil)
		playStarter(t, g, seed)
		for _, e := range eventsOfKind(g.Log, "action") {
			seen[strings.Fields(e.Detail)[0]] = true
		}
	}
	for k := ActGenerate; k <= ActEndTurn; k++ {
		if !seen[k.String()] {
			t.Errorf("no action event for %s", k)
		}
	}
}

func TestEvents_Case4_ActionEventCarriesItsFields(t *testing.T) {
	s := newScene(t)
	enemy := s.put(1, ZoneMainTheater, unit("Enemy", 1, 5, Keywords{}))
	att := s.put(0, ZoneMainTheater, unit("Att", 2, 5, Keywords{}))
	s.do(s.attack(att, enemy))
	acts := eventsOfKind(s.g.Log, "action")
	if len(acts) != 1 || acts[0].Detail != "Attack" {
		t.Fatalf("want one Attack action event, got %+v", acts)
	}
	a := acts[0]
	if a.Player != 0 || a.Source != att.ID || a.Target != enemy.ID {
		t.Errorf("action event fields: %+v", a)
	}
	// The action event comes before the effects.
	if i, j := slices.IndexFunc(s.g.Log, func(e Event) bool { return e.Kind == "action" }),
		slices.IndexFunc(s.g.Log, func(e Event) bool { return e.Kind == "attack" }); i < 0 || j < 0 || i > j {
		t.Errorf("action at %d must precede attack at %d", i, j)
	}
}

// Case 5: one decision event per choose that returns picks.
func TestEvents_Case5_DecisionAnswer(t *testing.T) {
	s := newScene(t)
	e1 := s.put(1, ZoneMainTheater, unit("Enemy A", 1, 5, Keywords{}))
	e2 := s.put(1, ZoneMainTheater, unit("Enemy B", 1, 5, Keywords{}))
	a1 := s.put(0, ZoneMainTheater, unit("Ally A", 1, 5, Keywords{}))
	s.scriptRefs(0, nil, e2, a1)
	deploySaboteur(s)
	decs := eventsOfKind(s.g.Log, "decision")
	if len(decs) != 2 {
		t.Fatalf("want 2 decision events (one per clause), got %d", len(decs))
	}
	e := decs[0]
	d := e.Decision
	if d == nil {
		t.Fatal("decision event has no Decision")
	}
	if d.How != "answer" || d.Kind != DecideTarget.String() || d.Player != 0 || e.Player != 0 || d.Prompt == "" {
		t.Errorf("decision record: %+v (event player %d)", *d, e.Player)
	}
	if !reflect.DeepEqual(d.Refs, []int{e1.ID, e2.ID}) || len(d.Options) != 2 {
		t.Errorf("offered refs %v options %v, want refs [%d %d]", d.Refs, d.Options, e1.ID, e2.ID)
	}
	if len(d.Picks) != 1 || d.Refs[d.Picks[0]] != e2.ID {
		t.Errorf("picks %v do not select asset %d", d.Picks, e2.ID)
	}
}

func TestEvents_Case5_DecisionAuto(t *testing.T) {
	s := newScene(t)
	e1 := s.put(1, ZoneMainTheater, unit("Enemy A", 1, 5, Keywords{}))
	s.script(0)
	deploySaboteur(s)
	decs := eventsOfKind(s.g.Log, "decision")
	if len(decs) != 2 {
		t.Fatalf("want 2 decision events, got %d", len(decs))
	}
	d := decs[0].Decision
	if d.How != "auto" || !reflect.DeepEqual(d.Refs, []int{e1.ID}) || !reflect.DeepEqual(d.Picks, []int{0}) {
		t.Errorf("single-option decision: %+v", *d)
	}
	if decs[1].Decision.How != "auto" {
		t.Errorf("allied clause with only the Saboteur is also auto: %+v", *decs[1].Decision)
	}
}

func TestEvents_Case5_DecisionTimeoutAndInvalid(t *testing.T) {
	setup := func() *scene {
		s := newScene(t)
		s.put(1, ZoneMainTheater, unit("Enemy A", 1, 5, Keywords{}))
		s.put(1, ZoneMainTheater, unit("Enemy B", 1, 5, Keywords{}))
		s.put(0, ZoneMainTheater, unit("Ally A", 1, 5, Keywords{}))
		return s
	}
	s := setup()
	s.script(0, Answer{Timeout: true}, Answer{Timeout: true})
	deploySaboteur(s)
	for i, e := range eventsOfKind(s.g.Log, "decision") {
		if e.Decision.How != "timeout" || len(e.Decision.Picks) != 1 {
			t.Errorf("decision %d after a timeout: %+v", i, *e.Decision)
		}
	}

	s = setup()
	s.script(0, Answer{Picks: []int{9}}, Answer{Timeout: true})
	deploySaboteur(s)
	decs := eventsOfKind(s.g.Log, "decision")
	if len(decs) == 0 || decs[0].Decision.How != "invalid" {
		t.Fatalf("an invalid answer is recorded as How invalid: %+v", decs)
	}
	if len(eventsOfKind(s.g.Log, "invalid-answer")) != 1 {
		t.Errorf("the invalid-answer event stays")
	}
}

func TestEvents_Case5_NoSatisfiableOptionsEmitsNothing(t *testing.T) {
	s := newScene(t)
	before := len(s.g.Log)
	s.g.choose(Decision{Player: 0, Kind: DecideTarget, Prompt: "impossible", Options: opts(1), Min: 2, Max: 2})
	if n := len(eventsOfKind(s.g.Log[before:], "decision")); n != 0 {
		t.Errorf("an unsatisfiable decision emitted %d decision events", n)
	}
}

// Case 6: existing call sites fill their structured fields.
func TestEvents_Case6_StructuredFields(t *testing.T) {
	s := newScene(t)
	e1 := s.put(1, ZoneMainTheater, unit("Enemy A", 1, 5, Keywords{}))
	e2 := s.put(1, ZoneMainTheater, unit("Enemy B", 1, 5, Keywords{}))
	a1 := s.put(0, ZoneMainTheater, unit("Ally A", 1, 5, Keywords{}))
	_ = e1
	s.scriptRefs(0, nil, e2, a1)
	c := deploySaboteur(s)
	sab := assetOf(s.g, c)

	deps := eventsOfKind(s.g.Log, "deploy")
	if len(deps) != 1 {
		t.Fatalf("want 1 deploy event, got %d", len(deps))
	}
	d := deps[0]
	if d.Player != 0 || d.Card != "Saboteur" || d.Inst != sab.Inst.ID || d.Source != sab.ID || d.Zone != "Reserves" {
		t.Errorf("deploy event fields: %+v (asset %d inst %d)", d, sab.ID, sab.Inst.ID)
	}

	var dmg *Event
	for _, e := range eventsOfKind(s.g.Log, "damage") {
		if e.Target == e2.ID {
			dmg = &e
		}
	}
	if dmg == nil {
		t.Fatal("no damage event against the chosen enemy")
	}
	if dmg.Source != sab.ID || dmg.Amount != 2 {
		t.Errorf("damage event fields: %+v", *dmg)
	}
}

// Case 7: Settled is only called when the state is settled.
func TestEvents_Case7_SettledStateHasNoLethalAsset(t *testing.T) {
	for seed := uint64(1); seed <= 5; seed++ {
		obs := &recorder{}
		obs.onSettled = func(s Snapshot) {
			var all []AssetSnap
			for _, p := range s.Players {
				if p.HQ != nil {
					all = append(all, *p.HQ)
				}
				all = append(all, p.Reserves...)
			}
			all = append(all, s.Theater...)
			for _, a := range all {
				if a.Damage >= a.MaxHealth {
					t.Fatalf("seed %d: settled snapshot has lethal asset %+v", seed, a)
				}
			}
		}
		g := starterGame(t, seed, obs)
		res := playStarter(t, g, seed)
		if len(obs.settled) < res.Actions {
			t.Errorf("seed %d: %d settled calls for %d actions", seed, len(obs.settled), res.Actions)
		}
	}
}

// Case 8: an observer changes nothing, and cannot change anything.
func TestEvents_Case8_ObserverIsReadOnlyAndDeterministic(t *testing.T) {
	for seed := uint64(1); seed <= 4; seed++ {
		plain := starterGame(t, seed, nil)
		pr := playStarter(t, plain, seed)

		obs := &recorder{}
		obs.onEvent = func(e Event) {
			if e.Snapshot != nil {
				for i := range e.Snapshot.Players {
					e.Snapshot.Players[i].Hand = nil
					e.Snapshot.Players[i].Deck = e.Snapshot.Players[i].Deck[:0]
				}
			}
			if e.Decision != nil && len(e.Decision.Picks) > 0 {
				e.Decision.Picks[0] = 99
			}
			if e.Start != nil {
				e.Start.Mulligans[0] = nil
			}
		}
		obs.onSettled = func(s Snapshot) {
			for i := range s.Players {
				for j := range s.Players[i].Hand {
					s.Players[i].Hand[j] = CardSnap{Card: "tampered"}
				}
				s.Players[i].Hand = s.Players[i].Hand[:0]
				s.Players[i].Deck = nil
			}
			s.Theater = nil
		}
		watched := starterGame(t, seed, obs)
		wr := playStarter(t, watched, seed)

		if pr.Outcome != wr.Outcome || pr.Turns != wr.Turns || pr.Actions != wr.Actions {
			t.Errorf("seed %d: result %+v with observer, %+v without", seed, wr, pr)
		}
		if !reflect.DeepEqual(plain.Log, watched.Log) {
			t.Errorf("seed %d: Log differs with an observer set", seed)
		}
		if len(obs.events) != len(watched.Log) {
			t.Errorf("seed %d: observer saw %d events, Log has %d", seed, len(obs.events), len(watched.Log))
		}
	}
}

// Case 9: existing event kinds keep their Turn and Detail.
func TestEvents_Case9_ExistingEventsUnchanged(t *testing.T) {
	s := newScene(t)
	enemy := s.put(1, ZoneMainTheater, unit("Enemy", 1, 5, Keywords{}))
	att := s.put(0, ZoneMainTheater, unit("Att", 2, 5, Keywords{}))
	hqAsset := s.g.Players[0].HQ
	s.do(Action{Kind: ActGenerate, Player: 0, Asset: hqAsset.ID, Resource: Requisition})
	s.do(s.attack(att, enemy))

	old := map[string]bool{"generate": true, "attack": true, "combat": true, "damage": true}
	var got []string
	for _, e := range s.g.Log {
		if old[e.Kind] {
			if e.Turn != 1 {
				t.Errorf("%s event Turn = %d, want 1", e.Kind, e.Turn)
			}
			got = append(got, e.Kind+": "+e.Detail)
		}
	}
	want := []string{
		fmt.Sprintf("generate: %s exhausts for resources (R1 T0)", hqAsset),
		fmt.Sprintf("attack: %s attacks %s", att, enemy),
		fmt.Sprintf("combat: %s (2) and %s (1) exchange combat damage", att, enemy),
		fmt.Sprintf("damage: %s deals 2 combat damage to %s (resist->2, cap->2, armor 0, health 2)", att, enemy),
		fmt.Sprintf("damage: %s deals 1 combat damage to %s (resist->1, cap->1, armor 0, health 1)", enemy, att),
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("existing events changed:\n got %q\nwant %q", got, want)
	}
}
