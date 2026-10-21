package sim

import (
	"bytes"
	"errors"
	"reflect"
	"strings"
	"testing"

	"animus-mundi/pkg/engine"
)

// Tests for T-0025: rule monitors. Each TestMonitor_Case<N> is Acceptance Case N.

const (
	zRes = "Reserves"
	zMT  = "Main Theater"
	zHQ  = "Headquarters Zone"
)

var (
	attackRules   = []string{"FR-CAP-001", "FR-CBT-001", "FR-KWD-002g", "FR-KWD-002j"}
	entryRules    = []string{"FR-KWD-002k", "FR-KWD-002l", "FR-SET-004", "FR-ZON-005"}
	resourceRules = []string{"FR-CAP-001", "FR-KWD-002i", "FR-RES-002", "FR-RES-003", "KIM-S03"}
)

// monitorDecks gives both seats the same small pool of named cards.
func monitorDecks() [2]engine.DeckList {
	dud := unit("Dud", 1, 1, engine.Keywords{})
	dud.ReqYield, dud.TacYield = 0, 0
	pool := []*engine.Card{
		unit("Vanilla", 2, 2, engine.Keywords{}),
		unit("Ghost", 2, 1, engine.Keywords{Ethereal: true}),
		unit("Bodyguard", 0, 4, engine.Keywords{Guardian: true}),
		structure("Bastion", 0, 6, engine.Keywords{}),
		unit("Slowpoke", 3, 3, engine.Keywords{DeploysExhausted: true}),
		unit("Plated", 2, 1, engine.Keywords{Armor: 2}),
		unit("Sneak", 1, 2, engine.Keywords{Versatile: true}),
		dud,
	}
	return [2]engine.DeckList{{HQ: hq(3, 20), Main: pool}, {HQ: hq(3, 20), Main: pool}}
}

func snapAsset(id int, card string, owner int, zone string) engine.AssetSnap {
	return engine.AssetSnap{ID: id, Inst: id + 1000, Card: card, Owner: owner, Controller: owner, Zone: zone}
}

// baseSnap is a Main Phase of player 0's turn with both HQs Readied and the
// given extra assets: Reserves assets are filed under their controller,
// Main Theater assets in the shared theater.
func baseSnap(req0, tac0 int, assets ...engine.AssetSnap) engine.Snapshot {
	s := engine.Snapshot{Turn: 1, Phase: "main", Active: 0, Winner: -1}
	for p := range 2 {
		h := snapAsset(1+p, hq(3, 20).Name, p, zHQ)
		s.Players[p].HQ = &h
	}
	s.Players[0].Req, s.Players[0].Tac = req0, tac0
	for _, a := range assets {
		switch a.Zone {
		case zRes:
			s.Players[a.Controller].Reserves = append(s.Players[a.Controller].Reserves, a)
		case zMT:
			s.Theater = append(s.Theater, a)
		}
	}
	return s
}

// feed sends hand-built Event and Settled calls to a Monitors for game 3.
type feed struct {
	m   *Monitors
	seq int
}

const feedSeed = 77

func newFeed() *feed {
	f := &feed{m: NewMonitors(monitorDecks(), 3)}
	f.event(engine.Event{Kind: "match-start", Player: -1, Start: &engine.MatchStart{Seed: feedSeed}})
	return f
}

// event numbers e and sends it, returning its Seq.
func (f *feed) event(e engine.Event) int {
	f.seq++
	e.Seq = f.seq
	if e.Phase == "" {
		e.Phase = "main"
	}
	f.m.Event(e)
	return e.Seq
}

func (f *feed) settle(s engine.Snapshot) { f.m.Settled(s) }

func wantViolation(t *testing.T, name string, m *Monitors, monitor string, rules []string, seq int) {
	t.Helper()
	v := m.Violation()
	if v == nil {
		t.Errorf("%s: no violation, want one from %s at Seq %d", name, monitor, seq)
		return
	}
	if v.Monitor != monitor || !reflect.DeepEqual(v.Rules, rules) || v.Game != 3 || v.Seed != feedSeed || v.Seq != seq {
		t.Errorf("%s: got %+v, want monitor %s rules %v game 3 seed %d seq %d", name, *v, monitor, rules, feedSeed, seq)
	}
	if v.Detail == "" {
		t.Errorf("%s: violation has no Detail", name)
	}
}

func wantClean(t *testing.T, name string, m *Monitors) {
	t.Helper()
	if v := m.Violation(); v != nil {
		t.Errorf("%s: unexpected violation: %v", name, v)
	}
}

func wantCounts(t *testing.T, name string, m *Monitors, monitor string, ids ...string) {
	t.Helper()
	for _, id := range ids {
		if m.Coverage()[monitor][id] == 0 {
			t.Errorf("%s: coverage %s/%s is 0, want above 0", name, monitor, id)
		}
	}
}

func attackFeed(before engine.Snapshot, attacker, defender int) (*feed, int) {
	f := newFeed()
	f.settle(before)
	seq := f.event(engine.Event{Kind: "action", Detail: "Attack", Player: 0, Source: attacker, Target: defender})
	return f, seq
}

func TestMonitor_Case1_AttackDeclaration(t *testing.T) {
	vanilla := func(id int, owner int, zone string) engine.AssetSnap { return snapAsset(id, "Vanilla", owner, zone) }
	exhausted := vanilla(10, 0, zMT)
	exhausted.Exhausted = true
	cases := []struct {
		name               string
		snap               engine.Snapshot
		attacker, defender int
	}{
		{"exhausted attacker", baseSnap(0, 0, exhausted, vanilla(11, 1, zMT)), 10, 11},
		{"attacker in Reserves", baseSnap(0, 0, vanilla(12, 0, zRes), vanilla(11, 1, zMT)), 12, 11},
		{"Structure attacker", baseSnap(0, 0, snapAsset(13, "Bastion", 0, zMT), vanilla(11, 1, zMT)), 13, 11},
		{"defender in Reserves", baseSnap(0, 0, vanilla(10, 0, zMT), vanilla(14, 1, zRes)), 10, 14},
		{"Ethereal defender", baseSnap(0, 0, vanilla(10, 0, zMT), snapAsset(15, "Ghost", 1, zMT)), 10, 15},
		{"non-Guardian defender beside a Guardian", baseSnap(0, 0, vanilla(10, 0, zMT), vanilla(11, 1, zMT), snapAsset(16, "Bodyguard", 1, zMT)), 10, 11},
	}
	for _, c := range cases {
		f, seq := attackFeed(c.snap, c.attacker, c.defender)
		wantViolation(t, c.name, f.m, "attack-declaration", attackRules, seq)
	}

	// Valid sequences.
	f, _ := attackFeed(baseSnap(0, 0, vanilla(10, 0, zMT), vanilla(11, 1, zMT)), 10, 11)
	wantClean(t, "unit attacks unit", f.m)
	wantCounts(t, "unit attacks unit", f.m, "attack-declaration", "FR-CBT-001", "FR-CAP-001")

	f, _ = attackFeed(baseSnap(0, 0, vanilla(10, 0, zMT)), 10, 2)
	wantClean(t, "HQ defender", f.m)

	f, _ = attackFeed(baseSnap(0, 0, vanilla(11, 1, zMT)), 1, 11)
	wantClean(t, "HQ attacker", f.m)

	f, _ = attackFeed(baseSnap(0, 0, vanilla(10, 0, zMT), vanilla(11, 1, zMT), snapAsset(16, "Bodyguard", 1, zMT), snapAsset(15, "Ghost", 1, zMT)), 10, 16)
	wantClean(t, "Guardian defender beside an Ethereal", f.m)
	wantCounts(t, "Guardian defender", f.m, "attack-declaration", "FR-KWD-002g", "FR-KWD-002j")
}

func TestMonitor_Case1_EntryState(t *testing.T) {
	deployFeed := func(card string, zone string, mutate func(*engine.AssetSnap)) (*feed, int) {
		f := newFeed()
		f.settle(baseSnap(0, 0))
		seq := f.event(engine.Event{Kind: "deploy", Player: 0, Card: card, Source: 30, Zone: zone})
		a := snapAsset(30, card, 0, zone)
		mutate(&a)
		f.settle(baseSnap(0, 0, a))
		return f, seq
	}
	scratchFeed := func(mutate func(*engine.AssetSnap)) (*feed, int) {
		f := newFeed()
		f.settle(baseSnap(0, 0))
		seq := f.event(engine.Event{Kind: "scratch", Player: 0, Source: 30})
		a := snapAsset(30, "Vanilla", 0, zRes)
		a.FaceDown, a.Exhausted = true, true
		mutate(&a)
		f.settle(baseSnap(0, 0, a))
		return f, seq
	}

	f, seq := scratchFeed(func(a *engine.AssetSnap) { a.Exhausted = false })
	wantViolation(t, "Readied Scratch Company", f.m, "entry-state", entryRules, seq)
	f, _ = scratchFeed(func(a *engine.AssetSnap) {})
	wantClean(t, "Scratch Company", f.m)
	wantCounts(t, "Scratch Company", f.m, "entry-state", "FR-ZON-005")

	f, seq = deployFeed("Vanilla", zMT, func(a *engine.AssetSnap) { a.Exhausted = true })
	wantViolation(t, "Exhausted unit", f.m, "entry-state", entryRules, seq)
	f, _ = deployFeed("Vanilla", zMT, func(a *engine.AssetSnap) {})
	wantClean(t, "Readied unit", f.m)
	wantCounts(t, "Readied unit", f.m, "entry-state", "FR-SET-004")

	f, seq = deployFeed("Slowpoke", zMT, func(a *engine.AssetSnap) {})
	wantViolation(t, "Readied Deploys Exhausted unit", f.m, "entry-state", entryRules, seq)
	f, _ = deployFeed("Slowpoke", zMT, func(a *engine.AssetSnap) { a.Exhausted = true })
	wantClean(t, "Exhausted Slowpoke", f.m)
	wantCounts(t, "Exhausted Slowpoke", f.m, "entry-state", "FR-KWD-002l")

	f, seq = deployFeed("Plated", zMT, func(a *engine.AssetSnap) { a.Armor = 0 })
	wantViolation(t, "Armor 2 at Armor 0", f.m, "entry-state", entryRules, seq)
	f, _ = deployFeed("Plated", zMT, func(a *engine.AssetSnap) { a.Armor = 2 })
	wantClean(t, "Plated", f.m)
	wantCounts(t, "Plated", f.m, "entry-state", "FR-KWD-002k")
}

func TestMonitor_Case1_Resources(t *testing.T) {
	// generate: a Readied asset before, the given pools and the Exhausted asset after.
	generate := func(card, zone, resource string, afterReq, afterTac int) (*feed, int) {
		f := newFeed()
		a := snapAsset(20, card, 0, zone)
		f.settle(baseSnap(0, 0, a))
		seq := f.event(engine.Event{Kind: "action", Detail: "Generate " + resource, Player: 0, Source: 20, Resource: resource})
		a.Exhausted = true
		f.settle(baseSnap(afterReq, afterTac, a))
		return f, seq
	}
	f, seq := generate("Vanilla", zRes, "Tactics", 0, 1)
	wantViolation(t, "Tactics from a Unit in Reserves", f.m, "resources", resourceRules, seq)
	f, seq = generate("Vanilla", zRes, "Requisition", 2, 0)
	wantViolation(t, "pool rise differs from yield", f.m, "resources", resourceRules, seq)
	f, seq = generate("Vanilla", zRes, "Requisition", 1, 1)
	wantViolation(t, "other pool changed", f.m, "resources", resourceRules, seq)

	f, _ = generate("Vanilla", zRes, "Requisition", 1, 0)
	wantClean(t, "Requisition from Reserves", f.m)
	wantCounts(t, "Requisition from Reserves", f.m, "resources", "FR-CAP-001")
	f, _ = generate("Vanilla", zMT, "Tactics", 0, 1)
	wantClean(t, "Tactics from the Main Theater", f.m)
	f, _ = generate("Sneak", zRes, "Tactics", 0, 1)
	wantClean(t, "Versatile Tactics from Reserves", f.m)
	wantCounts(t, "Versatile", f.m, "resources", "FR-KWD-002i")
	f, _ = generate("Dud", zRes, "Requisition", 0, 0)
	wantClean(t, "zero-yield generation", f.m)
	wantCounts(t, "zero-yield generation", f.m, "resources", "KIM-S03")

	convert := func(afterReq, afterTac int) (*feed, int) {
		f := newFeed()
		f.settle(baseSnap(3, 0))
		seq := f.event(engine.Event{Kind: "action", Detail: "Convert", Player: 0})
		f.settle(baseSnap(afterReq, afterTac))
		return f, seq
	}
	f, seq = convert(3, 0)
	wantViolation(t, "Convert leaves Requisition unchanged", f.m, "resources", resourceRules, seq)
	f, _ = convert(1, 1)
	wantClean(t, "Convert", f.m)
	wantCounts(t, "Convert", f.m, "resources", "FR-RES-003")

	turnStart := func(req int) (*feed, int) {
		f := newFeed()
		s := baseSnap(req, 0)
		seq := f.event(engine.Event{Kind: "turn-start", Player: 0, Snapshot: &s})
		return f, seq
	}
	f, seq = turnStart(1)
	wantViolation(t, "turn-start with 1 Requisition", f.m, "resources", resourceRules, seq)
	f, _ = turnStart(0)
	wantClean(t, "turn-start", f.m)
	wantCounts(t, "turn-start", f.m, "resources", "FR-RES-002")
}

func TestMonitor_FirstViolationWins(t *testing.T) {
	f, seq := attackFeed(baseSnap(0, 0, snapAsset(12, "Vanilla", 0, zRes), snapAsset(11, "Vanilla", 1, zMT)), 12, 11)
	f.event(engine.Event{Kind: "action", Detail: "Attack", Player: 0, Source: 12, Target: 99})
	wantViolation(t, "first of two", f.m, "attack-declaration", attackRules, seq)
}

// Case 2: the starter decks hold every rule over many real games.
func TestMonitor_Case2_StarterGamesAreClean(t *testing.T) {
	games := 500
	if testing.Short() {
		games = 50
	}
	decks := simStarterDecks(t)
	if _, err := SimulateWith(decks, games, 1, 400, SimOptions{DeckNames: simDeckNames}); err != nil {
		t.Errorf("RandomPolicy: %v", err)
	}
	if _, err := SimulateWith(decks, games, 1, 400, SimOptions{DeckNames: simDeckNames, Policy: ProgressPolicy}); err != nil {
		t.Errorf("ProgressPolicy: %v", err)
	}
}

// Case 3: monitors observe and change nothing.
func TestMonitor_Case3_ObservingChangesNothing(t *testing.T) {
	decks := simStarterDecks(t)
	run := func(noMonitors bool) (SimStats, []byte) {
		var cols []*gameCollector
		opt := SimOptions{DeckNames: simDeckNames, NoMonitors: noMonitors, Observer: func(game int, s uint64) engine.Observer {
			c := &gameCollector{}
			cols = append(cols, c)
			return c
		}}
		st, err := SimulateWith(decks, 20, 9, 400, opt)
		if err != nil {
			t.Fatalf("SimulateWith(NoMonitors %t): %v", noMonitors, err)
		}
		games := make([][]engine.Event, len(cols))
		for i, c := range cols {
			games[i] = c.events
		}
		return st, jsonlBytes(t, games)
	}
	on, onBytes := run(false)
	off, offBytes := run(true)
	if on.Coverage == nil {
		t.Errorf("default run has nil Coverage")
	}
	if off.Coverage != nil {
		t.Errorf("NoMonitors run has Coverage %v, want nil", off.Coverage)
	}
	on.Coverage = nil
	if !reflect.DeepEqual(on, off) {
		t.Errorf("SimStats differ:\n on  %+v\n off %+v", on, off)
	}
	if !bytes.Equal(onBytes, offBytes) {
		t.Errorf("event JSONL differs with monitors on and off")
	}
}

// Case 4: coverage is deterministic and has every rule ID.
func TestMonitor_Case4_Coverage(t *testing.T) {
	decks := simStarterDecks(t)
	run := func() Coverage {
		st, err := SimulateWith(decks, 50, 3, 400, SimOptions{DeckNames: simDeckNames})
		if err != nil {
			t.Fatal(err)
		}
		return st.Coverage
	}
	a, b := run(), run()
	if !reflect.DeepEqual(a, b) {
		t.Errorf("coverage differs between identical runs:\n%v\n%v", a, b)
	}
	want := map[string][]string{"attack-declaration": attackRules, "entry-state": entryRules, "resources": resourceRules,
		"damage-sequence": damageRules, "targeting": targetingRules}
	if len(a) != len(want) {
		t.Errorf("coverage has %d monitors, want %d: %v", len(a), len(want), a)
	}
	for mon, ids := range want {
		if len(a[mon]) != len(ids) {
			t.Errorf("%s has %d rule IDs, want %d: %v", mon, len(a[mon]), len(ids), a[mon])
		}
		for _, id := range ids {
			if _, ok := a[mon][id]; !ok {
				t.Errorf("coverage lacks %s/%s", mon, id)
			}
		}
	}
	for _, c := range [][2]string{
		{"attack-declaration", "FR-CBT-001"}, {"attack-declaration", "FR-CAP-001"}, {"resources", "FR-CAP-001"},
		{"entry-state", "FR-SET-004"}, {"entry-state", "FR-ZON-005"},
		{"damage-sequence", "FR-REP-004"}, {"damage-sequence", "FR-CBT-002"}, {"targeting", "FR-TGT-001"}, {"targeting", "SVY-R3"},
	} {
		if a[c[0]][c[1]] == 0 {
			t.Errorf("coverage %s/%s is 0 on the starter decks", c[0], c[1])
		}
	}
	// A fresh Monitors lists every ID with zero counts.
	fresh := NewMonitors(monitorDecks(), 1).Coverage()
	for mon, ids := range want {
		for _, id := range ids {
			if n, ok := fresh[mon][id]; !ok || n != 0 {
				t.Errorf("fresh coverage %s/%s = %d, present %t; want 0 and present", mon, id, n, ok)
			}
		}
	}
}

// Case 5: the violation text and the replay hint.
func TestMonitor_Case5_ViolationText(t *testing.T) {
	v := &Violation{Monitor: "attack-declaration", Rules: attackRules, Game: 3, Seed: 17, Seq: 240, Detail: "attacker was Exhausted"}
	msg := v.Error()
	if strings.Contains(msg, "\n") {
		t.Errorf("Error() is not one line: %q", msg)
	}
	for _, part := range append([]string{"attack-declaration", "3", "17", "240", "attacker was Exhausted"}, attackRules...) {
		if !strings.Contains(msg, part) {
			t.Errorf("Error() %q lacks %q", msg, part)
		}
	}
	var asErr error = v
	var target *Violation
	if !errors.As(asErr, &target) || target != v {
		t.Errorf("Violation does not work with errors.As")
	}

	got := ReplayCommand([]string{"-deck", "starter", "-p0", "ANIMUS", "-p1", "PATH", "-policy", "random"}, v)
	want := "cardtown sim -deck starter -p0 ANIMUS -p1 PATH -policy random -seed 17 -games 1 -events monitor-17.jsonl\n" +
		"cardtown replay -events monitor-17.jsonl -game 1 -at 240"
	if got != want {
		t.Errorf("ReplayCommand:\n got %q\nwant %q", got, want)
	}
}
