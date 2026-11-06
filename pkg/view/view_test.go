package view_test

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"testing"

	"animus-mundi/pkg/engine"
	"animus-mundi/pkg/engine/ability"
	"animus-mundi/pkg/eventlog"
	"animus-mundi/pkg/sim"
	"animus-mundi/pkg/view"
	"animus-mundi/pkg/view/viewtest"
)

// Tests for T-0036: PlayerView and LogLines (board directive #24, FR-ZON-005).
// Each TestView_Case<N> is Acceptance Case N.

// --- scene helpers -----------------------------------------------------------

func testHQ(name string) *engine.Card {
	return &engine.Card{Name: name, Type: engine.TypeHeadquarters, Power: 3, Health: 20, ReqYield: 1, TacYield: 2,
		DeployZones: []engine.Zone{engine.ZoneHeadquarters}}
}

func testUnit(name string, power, health int, kw engine.Keywords) *engine.Card {
	return &engine.Card{Name: name, Type: engine.TypeUnit, Power: power, Health: health, ReqCost: 2, TacCost: 1,
		ReqYield: 1, TacYield: 1, RulesText: name + " rules text", Keywords: kw,
		DeployZones: []engine.Zone{engine.ZoneReserves, engine.ZoneMainTheater}}
}

// scene is a game from NewGame whose exported state a test then sets by hand.
// Instance and asset IDs start above the range NewGame uses.
type scene struct {
	g      *engine.Game
	nextID int
}

// newScene starts a match where player 0 takes Turn 1, with HQs "Alpha HQ"
// and "Beta HQ", then empties both hands, decks and discard piles.
func newScene(t *testing.T) *scene {
	t.Helper()
	filler := func(name string) []*engine.Card {
		out := make([]*engine.Card, 10)
		for i := range out {
			out[i] = testUnit(name, 1, 1, engine.Keywords{})
		}
		return out
	}
	g, err := engine.NewGame(engine.Config{
		Seed: 1,
		Decks: [2]engine.DeckList{
			{HQ: testHQ("Alpha HQ"), Main: filler("Alpha Filler")},
			{HQ: testHQ("Beta HQ"), Main: filler("Beta Filler")},
		},
		PlayFirst: func(decider engine.PlayerID) bool { return decider == 0 },
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range g.Players {
		p.Hand, p.Deck, p.Discard = nil, nil, nil
	}
	return &scene{g: g, nextID: 100000}
}

func (s *scene) id() int {
	s.nextID++
	return s.nextID
}

func (s *scene) inst(p engine.PlayerID, c *engine.Card) *engine.Instance {
	return &engine.Instance{ID: s.id(), Card: c, Owner: p}
}

// put places a Readied, face-up asset onto the battlefield.
func (s *scene) put(p engine.PlayerID, zone engine.Zone, c *engine.Card) *engine.Asset {
	a := &engine.Asset{ID: s.id(), Inst: s.inst(p, c), Controller: p, Zone: zone, Armor: c.Keywords.Armor}
	switch zone {
	case engine.ZoneMainTheater:
		s.g.Theater = append(s.g.Theater, a)
	case engine.ZoneReserves:
		s.g.Players[p].Reserves = append(s.g.Players[p].Reserves, a)
	}
	return a
}

func mustJSON(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return string(b)
}

func cardNames(cs []view.CardView) []string {
	var out []string
	for _, c := range cs {
		out = append(out, c.Card)
	}
	return out
}

// contentScene is the Case 1 scene, also used by Case 3.
type contentScene struct {
	*scene
	p0Hand                 *engine.Instance
	p1Hand                 [2]*engine.Instance
	sentry, lancer, raider *engine.Asset
	incoming               *engine.Instance
}

func newContentScene(t *testing.T) *contentScene {
	t.Helper()
	s := &contentScene{scene: newScene(t)}
	g := s.g
	p0, p1 := g.Players[0], g.Players[1]

	s.p0Hand = s.inst(0, testUnit("Alpha Courier", 2, 3, engine.Keywords{Ward: true}))
	p0.Hand = []*engine.Instance{s.p0Hand}
	s.p1Hand = [2]*engine.Instance{
		s.inst(1, testUnit("Beta Secret Plan", 4, 4, engine.Keywords{})),
		s.inst(1, testUnit("Beta Hidden Edge", 1, 5, engine.Keywords{Stalwart: true})),
	}
	p1.Hand = s.p1Hand[:]
	for range 4 {
		p0.Deck = append(p0.Deck, s.inst(0, testUnit("Alpha Reserve Stock", 1, 1, engine.Keywords{})))
	}
	for range 3 {
		p1.Deck = append(p1.Deck, s.inst(1, testUnit("Beta Deck Card", 1, 1, engine.Keywords{})))
	}
	p0.Discard = []*engine.Instance{s.inst(0, testUnit("Alpha Fallen", 1, 1, engine.Keywords{}))}
	p1.Discard = []*engine.Instance{s.inst(1, testUnit("Beta Fallen", 1, 1, engine.Keywords{}))}

	p0.HQ.Damage = 4
	s.sentry = s.put(0, engine.ZoneReserves, testUnit("Alpha Sentry", 2, 4, engine.Keywords{}))
	s.sentry.Damage, s.sentry.Armor, s.sentry.Exhausted, s.sentry.TempPower = 1, 2, true, 3
	s.lancer = s.put(0, engine.ZoneMainTheater, testUnit("Alpha Lancer", 3, 2, engine.Keywords{Aggressive: 2}))
	s.raider = s.put(1, engine.ZoneMainTheater, testUnit("Beta Raider", 3, 3, engine.Keywords{Guardian: true}))

	s.incoming = s.inst(1, testUnit("Beta Incoming", 2, 2, engine.Keywords{}))
	g.Stack = []*engine.StackItem{
		{Kind: engine.StackDeploy, Controller: 1, Inst: s.incoming, Zone: engine.ZoneReserves},
		{Kind: engine.StackAttack, Controller: 0, Source: s.lancer.ID, Target: s.raider.ID},
		{Kind: engine.StackAbility, Controller: 0, Source: s.sentry.ID, Inst: s.sentry.Inst, Chosen: [][]int{{s.raider.ID}}},
	}

	p0.Req, p0.Tac, p1.Req, p1.Tac = 3, 1, 2, 4
	g.Turn, g.Phase, g.Active, g.FirstPlayer = 5, engine.PhaseMain, 1, 0
	return s
}

// --- Case 1: contents --------------------------------------------------------

func TestView_Case1_Contents(t *testing.T) {
	s := newContentScene(t)
	g := s.g

	v := view.PlayerView(g, 0)
	js := mustJSON(t, v)

	if v.Seat != 0 || v.Active != 1 || v.First != 0 || v.Turn != 5 || v.Phase != "main" {
		t.Errorf("seat/active/first/turn/phase = %d/%d/%d/%d/%q, want 0/1/0/5/main", v.Seat, v.Active, v.First, v.Turn, v.Phase)
	}
	if v.Over || v.Draw || v.Winner != -1 || v.Reason != "" {
		t.Errorf("outcome = over %t draw %t winner %d reason %q, want a match in progress, winner -1", v.Over, v.Draw, v.Winner, v.Reason)
	}

	// Own hand: names, printed characteristics and instance IDs.
	if len(v.You.Hand) != 1 || v.You.HandCount != 1 {
		t.Fatalf("own hand = %+v (count %d), want one card", v.You.Hand, v.You.HandCount)
	}
	h := v.You.Hand[0]
	if h.Card != "Alpha Courier" || h.Inst != s.p0Hand.ID || h.Type != "Unit" || h.ReqCost != 2 || h.TacCost != 1 ||
		h.Power != 2 || h.Health != 3 || !slices.Contains(h.Keywords, "Ward") {
		t.Errorf("own hand card = %+v, want Alpha Courier inst %d, Unit, R2 T1, 2/3, Ward", h, s.p0Hand.ID)
	}

	// Opponent's hand: a count, and no names anywhere.
	if len(v.Opponent.Hand) != 0 || v.Opponent.HandCount != 2 {
		t.Errorf("opponent hand = %+v count %d, want no cards and count 2", v.Opponent.Hand, v.Opponent.HandCount)
	}
	for _, name := range []string{"Beta Secret Plan", "Beta Hidden Edge", "Beta Deck Card", "Alpha Reserve Stock"} {
		if strings.Contains(js, name) {
			t.Errorf("seat 0's view names %q, a hidden card", name)
		}
	}
	for _, inst := range s.p1Hand {
		if strings.Contains(js, fmt.Sprint(inst.ID)) {
			t.Errorf("seat 0's view contains the opponent's hand instance ID %d", inst.ID)
		}
	}

	// Decks: counts only.
	if v.You.DeckCount != 4 || v.Opponent.DeckCount != 3 {
		t.Errorf("deck counts = %d/%d, want 4/3", v.You.DeckCount, v.Opponent.DeckCount)
	}

	// Discard piles: both, with names, no instance IDs.
	if got := cardNames(v.You.Discard); !slices.Equal(got, []string{"Alpha Fallen"}) {
		t.Errorf("own discard = %v, want [Alpha Fallen]", got)
	}
	if got := cardNames(v.Opponent.Discard); !slices.Equal(got, []string{"Beta Fallen"}) {
		t.Errorf("opponent discard = %v, want [Beta Fallen]", got)
	}
	for _, c := range append(slices.Clone(v.You.Discard), v.Opponent.Discard...) {
		if c.Inst != 0 {
			t.Errorf("discard card %q carries instance ID %d, want 0", c.Card, c.Inst)
		}
	}

	// HQs, Reserves and Main Theater.
	if v.You.HQ == nil || v.Opponent.HQ == nil {
		t.Fatalf("HQs = %v/%v, want both", v.You.HQ, v.Opponent.HQ)
	}
	if hq := *v.You.HQ; hq.Card != "Alpha HQ" || hq.ID != g.Players[0].HQ.ID || hq.Damage != 4 || hq.Power != 3 ||
		hq.MaxHealth != 20 || hq.Zone != "Headquarters Zone" {
		t.Errorf("own HQ = %+v, want Alpha HQ, damage 4, 3/20", hq)
	}
	if hq := *v.Opponent.HQ; hq.Card != "Beta HQ" || hq.ID != g.Players[1].HQ.ID || hq.Controller != 1 {
		t.Errorf("opponent HQ = %+v, want Beta HQ controlled by 1", hq)
	}
	if len(v.You.Reserves) != 1 || len(v.Opponent.Reserves) != 0 {
		t.Fatalf("reserves = %+v / %+v, want one own asset", v.You.Reserves, v.Opponent.Reserves)
	}
	if r := v.You.Reserves[0]; r.ID != s.sentry.ID || r.Card != "Alpha Sentry" || r.Damage != 1 || r.Armor != 2 ||
		r.Power != 5 || r.MaxHealth != 4 || !r.Exhausted || r.FaceDown || r.Zone != "Reserves" || r.Controller != 0 || r.Owner != 0 {
		t.Errorf("own reserve = %+v, want Alpha Sentry, damage 1, armor 2, effective power 5, max health 4, exhausted", r)
	}
	if len(v.Theater) != 2 {
		t.Fatalf("theater = %+v, want 2 assets", v.Theater)
	}
	if a := v.Theater[0]; a.ID != s.lancer.ID || a.Card != "Alpha Lancer" || a.Power != 3 || a.MaxHealth != 2 ||
		a.Exhausted || a.Zone != "Main Theater" || !slices.Equal(a.Keywords, []string{"Aggressive 2"}) {
		t.Errorf("theater[0] = %+v, want Alpha Lancer 3/2 Aggressive 2, ready", a)
	}
	if a := v.Theater[1]; a.ID != s.raider.ID || a.Card != "Beta Raider" || a.Controller != 1 || a.Owner != 1 ||
		!slices.Equal(a.Keywords, []string{"Guardian"}) {
		t.Errorf("theater[1] = %+v, want Beta Raider controlled by 1, Guardian", a)
	}
	for _, a := range append(append(slices.Clone(v.Theater), v.You.Reserves...), *v.You.HQ, *v.Opponent.HQ) {
		for _, inst := range []*engine.Instance{s.sentry.Inst, s.lancer.Inst, s.raider.Inst} {
			if strings.Contains(mustJSON(t, a), fmt.Sprint(inst.ID)) {
				t.Errorf("asset view %+v contains instance ID %d", a, inst.ID)
			}
		}
	}

	// Resource pools.
	if v.You.Req != 3 || v.You.Tac != 1 || v.Opponent.Req != 2 || v.Opponent.Tac != 4 {
		t.Errorf("pools = R%d T%d / R%d T%d, want R3 T1 / R2 T4", v.You.Req, v.You.Tac, v.Opponent.Req, v.Opponent.Tac)
	}

	// The stack, bottom first.
	if len(v.Stack) != 3 {
		t.Fatalf("stack = %+v, want 3 items", v.Stack)
	}
	if it := v.Stack[0]; it.Kind != "Deploy" || it.Controller != 1 || it.Card != "Beta Incoming" || it.Zone != "Reserves" {
		t.Errorf("stack[0] = %+v, want player 1's Deploy of Beta Incoming to Reserves", it)
	}
	if it := v.Stack[1]; it.Kind != "Attack" || it.Controller != 0 || it.Source != s.lancer.ID || it.Target != s.raider.ID {
		t.Errorf("stack[1] = %+v, want player 0's Attack by %d on %d", it, s.lancer.ID, s.raider.ID)
	}
	if it := v.Stack[2]; it.Kind != "Ability" || it.Controller != 0 || it.Source != s.sentry.ID || it.Card != "Alpha Sentry" ||
		!reflect.DeepEqual(it.Chosen, [][]int{{s.raider.ID}}) {
		t.Errorf("stack[2] = %+v, want player 0's ability of Alpha Sentry targeting %d", it, s.raider.ID)
	}

	// The other seat sees the mirror image.
	w := view.PlayerView(g, 1)
	wjs := mustJSON(t, w)
	if w.Seat != 1 || len(w.You.Hand) != 2 || w.You.Hand[0].Inst != s.p1Hand[0].ID || w.You.Hand[1].Card != "Beta Hidden Edge" ||
		!slices.Equal(w.You.Hand[1].Keywords, []string{"Stalwart"}) {
		t.Errorf("seat 1's own hand = %+v, want Beta Secret Plan and Beta Hidden Edge with instance IDs", w.You.Hand)
	}
	if len(w.Opponent.Hand) != 0 || w.Opponent.HandCount != 1 || w.You.DeckCount != 3 || w.Opponent.DeckCount != 4 {
		t.Errorf("seat 1 sees opponent hand %+v count %d, decks %d/%d, want no cards, count 1, decks 3/4",
			w.Opponent.Hand, w.Opponent.HandCount, w.You.DeckCount, w.Opponent.DeckCount)
	}
	for _, name := range []string{"Alpha Courier", "Alpha Reserve Stock", "Beta Deck Card"} {
		if strings.Contains(wjs, name) {
			t.Errorf("seat 1's view names %q, a hidden card", name)
		}
	}
	if strings.Contains(wjs, fmt.Sprint(s.p0Hand.ID)) {
		t.Errorf("seat 1's view contains the opponent's hand instance ID %d", s.p0Hand.ID)
	}
	if w.You.HQ == nil || w.You.HQ.Card != "Beta HQ" || w.Opponent.HQ == nil || w.Opponent.HQ.Card != "Alpha HQ" ||
		len(w.Opponent.Reserves) != 1 || w.You.Req != 2 || w.Opponent.Tac != 1 || len(w.Theater) != 2 || len(w.Stack) != 3 {
		t.Errorf("seat 1's view = %+v, want the same public state from the other side", w)
	}

	// Outcome: a win, then a draw.
	g.Phase, g.Outcome = engine.PhaseOver, engine.Outcome{Over: true, Winner: 1, Reason: "headquarters destroyed"}
	if v := view.PlayerView(g, 0); !v.Over || v.Draw || v.Winner != 1 || v.Reason != "headquarters destroyed" || v.Phase != "over" {
		t.Errorf("won outcome = over %t draw %t winner %d reason %q phase %q, want player 1's win", v.Over, v.Draw, v.Winner, v.Reason, v.Phase)
	}
	g.Outcome = engine.Outcome{Over: true, Draw: true, Reason: "both headquarters destroyed"}
	if v := view.PlayerView(g, 1); !v.Over || !v.Draw || v.Winner != -1 || v.Reason != "both headquarters destroyed" {
		t.Errorf("drawn outcome = over %t draw %t winner %d reason %q, want a draw with winner -1", v.Over, v.Draw, v.Winner, v.Reason)
	}
}

// --- Case 2: face-down -------------------------------------------------------

func TestView_Case2_FaceDown(t *testing.T) {
	s := newScene(t)
	g := s.g
	hidden := testUnit("Alpha Hidden Gem", 5, 7, engine.Keywords{Ward: true, Armor: 2})
	fd := &engine.Asset{ID: s.id(), Inst: &engine.Instance{ID: 424242, Card: hidden, Owner: 0}, Controller: 0,
		Zone: engine.ZoneMainTheater, FaceDown: true}
	g.Theater = append(g.Theater, fd)
	raider := s.put(1, engine.ZoneMainTheater, testUnit("Beta Raider", 3, 3, engine.Keywords{}))

	for seat := range engine.PlayerID(2) {
		v := view.PlayerView(g, seat)
		js := mustJSON(t, v)
		var got *view.AssetView
		for i := range v.Theater {
			if v.Theater[i].ID == fd.ID {
				got = &v.Theater[i]
			}
		}
		if got == nil {
			t.Fatalf("seat %d: face-down asset %d missing from the theater %+v", seat, fd.ID, v.Theater)
		}
		if got.Card != "Scratch Company" || !got.FaceDown || got.Power != 1 || got.MaxHealth != 1 || !slices.Equal(got.Keywords, []string{"Deploys Exhausted"}) ||
			got.Controller != 0 || got.Owner != 0 {
			t.Errorf("seat %d sees the face-down asset as %+v, want a 1/1 Scratch Company with exactly Deploys Exhausted", seat, *got)
		}
		if strings.Contains(js, "Alpha Hidden Gem") {
			t.Errorf("seat %d's view names the face-down card", seat)
		}
		if strings.Contains(js, "424242") {
			t.Errorf("seat %d's view contains the face-down card's instance ID", seat)
		}
	}

	// Destroy it: player 1's 3/3 attacks it.
	g.Active = 1
	var attack *engine.Action
	for _, a := range g.LegalActions() {
		if a.Kind == engine.ActAttack && a.Asset == raider.ID && a.Target == fd.ID {
			attack = &a
		}
	}
	if attack == nil {
		t.Fatalf("no legal attack by %d on the face-down asset %d", raider.ID, fd.ID)
	}
	if err := g.Apply(*attack); err != nil {
		t.Fatal(err)
	}
	if g.Asset(fd.ID) != nil {
		t.Fatalf("the face-down asset survived the attack")
	}
	for seat := range engine.PlayerID(2) {
		v := view.PlayerView(g, seat)
		owner := v.You
		if seat == 1 {
			owner = v.Opponent
		}
		if got := cardNames(owner.Discard); !slices.Equal(got, []string{"Alpha Hidden Gem"}) {
			t.Errorf("seat %d sees player 0's discard as %v, want [Alpha Hidden Gem]", seat, got)
		}
		if strings.Contains(mustJSON(t, v), "424242") {
			t.Errorf("seat %d's view contains the destroyed card's instance ID", seat)
		}
	}
}

// --- Case 3: deep copy -------------------------------------------------------

// scribble changes every value reachable from v: strings, numbers and bools,
// through pointers, slices and arrays, and appends to every slice.
func scribble(v reflect.Value) {
	switch v.Kind() {
	case reflect.Pointer:
		if !v.IsNil() {
			scribble(v.Elem())
		}
	case reflect.Struct:
		for i := range v.NumField() {
			scribble(v.Field(i))
		}
	case reflect.Array:
		for i := range v.Len() {
			scribble(v.Index(i))
		}
	case reflect.Slice:
		for i := range v.Len() {
			scribble(v.Index(i))
		}
		if v.CanSet() {
			v.Set(reflect.Append(v, reflect.Zero(v.Type().Elem())))
		}
	case reflect.String:
		v.SetString(v.String() + " scribbled")
	case reflect.Int:
		v.SetInt(v.Int() + 1000)
	case reflect.Bool:
		v.SetBool(!v.Bool())
	}
}

// stackCopy deep-copies the stack items' values.
func stackCopy(st []*engine.StackItem) []engine.StackItem {
	var out []engine.StackItem
	for _, it := range st {
		c := *it
		c.Chosen = nil
		for _, ids := range it.Chosen {
			c.Chosen = append(c.Chosen, slices.Clone(ids))
		}
		out = append(out, c)
	}
	return out
}

func TestView_Case3_DeepCopy(t *testing.T) {
	s := newContentScene(t)
	g := s.g
	for seat := range engine.PlayerID(2) {
		ref := view.PlayerView(g, seat)
		snap := g.Snapshot()
		stack := stackCopy(g.Stack)
		names := []string{s.p0Hand.Card.Name, s.p1Hand[0].Card.Name, s.sentry.Inst.Card.Name, s.incoming.Card.Name}

		v := view.PlayerView(g, seat)
		scribble(reflect.ValueOf(&v))
		if reflect.DeepEqual(v, ref) {
			t.Fatalf("seat %d: scribble changed nothing", seat)
		}

		if got := g.Snapshot(); !reflect.DeepEqual(got, snap) {
			t.Errorf("seat %d: changing the view changed the game state", seat)
		}
		if got := stackCopy(g.Stack); !reflect.DeepEqual(got, stack) {
			t.Errorf("seat %d: changing the view changed the stack: %+v, want %+v", seat, got, stack)
		}
		if got := []string{s.p0Hand.Card.Name, s.p1Hand[0].Card.Name, s.sentry.Inst.Card.Name, s.incoming.Card.Name}; !slices.Equal(got, names) {
			t.Errorf("seat %d: changing the view renamed cards: %v, want %v", seat, got, names)
		}
		if got := view.PlayerView(g, seat); !reflect.DeepEqual(got, ref) {
			t.Errorf("seat %d: a fresh view after changing the old one = %+v, want %+v", seat, got, ref)
		}
	}
}

// --- Case 4: no leak over seeded starter matches ------------------------------

func starterDecks(t *testing.T) [2]engine.DeckList {
	t.Helper()
	cards, err := engine.LoadCards(filepath.Join("..", "..", "data", "cards"))
	if err != nil {
		t.Fatalf("load cards: %v", err)
	}
	decks, err := sim.SimDecks(cards, filepath.Join("..", "..", "data", "decks"), "starter", [2]string{"ANIMUS", "PATH"})
	if err != nil {
		t.Fatalf("SimDecks: %v", err)
	}
	return decks
}

var starterDeckNames = [2]string{"data/decks/animus-starter.txt", "data/decks/path-starter.txt"}

// mulliganFirst sets aside the first n cards of the hand, n = seat + 1, so
// both seats have mulliganed names to hide.
func mulliganFirst(p engine.PlayerID, hand []*engine.Instance) []int {
	var ids []int
	for _, inst := range hand[:min(int(p)+1, len(hand))] {
		ids = append(ids, inst.ID)
	}
	return ids
}

// leakWatcher collects a game's events and, at every settled point, checks
// both seats' views, and each seat's log lines new since the last point, for
// hidden names (Case 4 as amended: a line is checked once, at the first
// settled point at or after its event). A log line never names a hidden card
// that was not public when its event happened or at the first settled point
// after it (T-0079: a pick from a public pile logs the card while it is
// public, and the same step may then hide it).
type leakWatcher struct {
	t       *testing.T
	seed    uint64
	g       *engine.Game // nil during NewGame; the setup point is checked after it returns
	events  []engine.Event
	pubAt   []map[string]bool // public names when each event was recorded; nil while g is nil
	lines   [2][]string       // each seat's LogLines so far, built one settled point at a time
	checked int               // events already turned into lines
	points  int
	failed  bool
}

func (w *leakWatcher) Event(e engine.Event) {
	w.events = append(w.events, e)
	var pub map[string]bool
	if w.g != nil {
		pub = viewtest.PublicNames(w.g)
	}
	w.pubAt = append(w.pubAt, pub)
}

func (w *leakWatcher) Settled(engine.Snapshot) {
	if w.g != nil {
		w.check()
	}
}

// jsonName is name as it appears inside a JSON string.
func jsonName(name string) string {
	b, _ := json.Marshal(name)
	return string(b[1 : len(b)-1])
}

func (w *leakWatcher) check() {
	w.points++
	g := w.g
	pub := viewtest.PublicNames(g)
	newEvents := w.events[w.checked:]
	newPub := w.pubAt[w.checked:]
	w.checked = len(w.events)
	for seat := range engine.PlayerID(2) {
		// The seat's LogLines so far are one line per event: the earlier lines
		// plus those of the new events. The full call is compared with this at
		// the end of the match.
		newLines := view.LogLines(newEvents, seat)
		w.lines[seat] = append(w.lines[seat], newLines...)
		if w.failed {
			continue // one report per match
		}
		js := mustJSON(w.t, view.PlayerView(g, seat))
		for name := range viewtest.HiddenNames(g, seat.Opponent()) {
			if pub[name] {
				continue
			}
			if strings.Contains(js, jsonName(name)) {
				w.t.Errorf("seed %d, settled point %d (turn %d, after event %d): seat %d's view names %q, hidden by player %d",
					w.seed, w.points, g.Turn, len(w.events), seat, name, seat.Opponent())
				w.failed = true
			}
			for i, line := range newLines {
				if newPub[i][name] {
					continue
				}
				if strings.Contains(line, name) {
					w.t.Errorf("seed %d, settled point %d (turn %d): seat %d's log names %q, hidden by player %d: %s",
						w.seed, w.points, g.Turn, seat, name, seat.Opponent(), line)
					w.failed = true
				}
			}
		}
	}
}

func TestView_Case4_NoLeak(t *testing.T) {
	decks := starterDecks(t)
	var universe []string
	seen := map[string]bool{}
	for _, d := range decks {
		for _, c := range append([]*engine.Card{d.HQ}, d.Main...) {
			if !seen[c.Name] {
				seen[c.Name] = true
				universe = append(universe, c.Name)
			}
		}
	}
	// The substring checks rely on no card name containing another one.
	for _, a := range universe {
		for _, b := range universe {
			if a != b && strings.Contains(a, b) {
				t.Fatalf("card name %q contains %q; the leak check would misfire", a, b)
			}
		}
	}

	const matches = 20
	points := 0
	for i := range uint64(matches) {
		seed := 1 + i
		w := &leakWatcher{t: t, seed: seed}
		g, err := engine.NewGame(engine.Config{Seed: seed, Decks: decks, DeckNames: starterDeckNames, Mulligan: mulliganFirst,
			Choosers: [2]engine.Chooser{engine.RandomChooser(seed*2 + 1), engine.RandomChooser(seed*2 + 2)}, Observer: w})
		if err != nil {
			t.Fatal(err)
		}
		w.g = g
		w.check() // the setup settled point: NewGame's last step, before it returned
		res, err := sim.Playout(g, [2]sim.Policy{sim.ProgressPolicy(seed*2 + 1), sim.ProgressPolicy(seed*2 + 2)}, 400)
		if err != nil {
			t.Fatalf("seed %d: %v", seed, err)
		}
		if !res.Outcome.Over {
			t.Fatalf("seed %d: the match did not finish", seed)
		}
		for seat := range engine.PlayerID(2) {
			if got := view.LogLines(w.events, seat); !slices.Equal(got, w.lines[seat]) {
				t.Errorf("seed %d: seat %d's LogLines over the whole match differ from the lines built point by point", seed, seat)
			}
		}
		if w.points < 2 {
			t.Errorf("seed %d: only %d settled points checked", seed, w.points)
		}
		points += w.points
	}
	t.Logf("%d matches, %d settled points checked for both seats", matches, points)
}

// --- Case 5: LogLines --------------------------------------------------------

func TestView_Case5_LogLines(t *testing.T) {
	const seed = 987654321
	decks := starterDecks(t)
	g, err := engine.NewGame(engine.Config{Seed: seed, Decks: decks, DeckNames: starterDeckNames, Mulligan: mulliganFirst,
		Choosers: [2]engine.Chooser{engine.RandomChooser(seed*2 + 1), engine.RandomChooser(seed*2 + 2)}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := sim.Playout(g, [2]sim.Policy{sim.ProgressPolicy(seed*2 + 1), sim.ProgressPolicy(seed*2 + 2)}, 400); err != nil {
		t.Fatal(err)
	}
	events := g.Log
	turnStarts := 0
	for seat := range engine.PlayerID(2) {
		opp := seat.Opponent()
		lines := view.LogLines(events, seat)
		if len(lines) != len(events) {
			t.Fatalf("seat %d: %d lines for %d events", seat, len(lines), len(events))
		}
		for i, e := range events {
			line := lines[i]
			switch e.Kind {
			case "match-start":
				st := e.Start
				if len(st.Mulligans[0]) != 1 || len(st.Mulligans[1]) != 2 {
					t.Fatalf("mulligans = %v, want 1 and 2 cards set aside", st.Mulligans)
				}
				head := fmt.Sprintf("%d T%d %s match-start: ", e.Seq, e.Turn, e.Phase)
				if !strings.HasPrefix(line, head) {
					t.Errorf("seat %d: match-start line %q, want it to start %q", seat, line, head)
				}
				if strings.Contains(line, "seed") || strings.Contains(line, fmt.Sprint(seed)) {
					t.Errorf("seat %d: match-start line shows the seed: %s", seat, line)
				}
				for _, want := range []string{
					fmt.Sprintf("%q vs %q", st.Decks[0], st.Decks[1]), st.HQs[0], st.HQs[1],
					fmt.Sprintf("player %d decides, play first=%t", st.Decider, st.PlayFirst),
					fmt.Sprintf("player %d set aside %d", opp, len(st.Mulligans[opp])),
				} {
					if !strings.Contains(line, want) {
						t.Errorf("seat %d: match-start line %q lacks %q", seat, line, want)
					}
				}
				for _, name := range st.Mulligans[seat] {
					if !strings.Contains(line, name) {
						t.Errorf("seat %d: match-start line %q lacks its own mulliganed card %q", seat, line, name)
					}
				}
				for _, name := range st.Mulligans[opp] {
					if !slices.Contains(st.Mulligans[seat], name) && strings.Contains(line, name) {
						t.Errorf("seat %d: match-start line %q names the opponent's mulliganed card %q", seat, line, name)
					}
				}
			case "turn-start":
				turnStarts++
				want := fmt.Sprintf("%d T%d %s turn-start: turn %d, player %d is active", e.Seq, e.Turn, e.Phase, e.Turn, e.Player)
				if line != want {
					t.Errorf("seat %d: turn-start line %q, want %q", seat, line, want)
				}
			default:
				if want := eventlog.FormatEvent(e); line != want {
					t.Errorf("seat %d: event %d line %q, want FormatEvent's %q", seat, e.Seq, line, want)
				}
			}
		}
	}
	if turnStarts < 4 {
		t.Errorf("only %d turn-start lines checked", turnStarts)
	}
}

// T-0074: an Aggressive trigger is an Ability item with a Keyword.
func TestView_AbilityItemShowsItsKeyword(t *testing.T) {
	s := newContentScene(t)
	s.g.Stack = []*engine.StackItem{
		{Kind: engine.StackAbility, Controller: 0, Source: s.lancer.ID, Inst: s.lancer.Inst, Keyword: ability.KeywordAggressive},
		{Kind: engine.StackAbility, Controller: 0, Source: s.sentry.ID, Inst: s.sentry.Inst},
	}
	v := view.PlayerView(s.g, 0)
	if len(v.Stack) != 2 {
		t.Fatalf("stack = %+v", v.Stack)
	}
	if it := v.Stack[0]; it.Kind != "Ability" || it.Keyword != "Aggressive" || it.Card != "Alpha Lancer" {
		t.Errorf("Aggressive item = %+v", it)
	}
	if it := v.Stack[1]; it.Kind != "Ability" || it.Keyword != "" {
		t.Errorf("printed item = %+v", it)
	}
}

// T-0081: a chosen card in a discard pile is shown by name, never by instance ID.
func TestView_PileTargetShowsCardNamesNotInstanceIDs(t *testing.T) {
	s := newScene(t)
	src := s.put(0, engine.ZoneMainTheater, testUnit("Pile Source", 1, 1, engine.Keywords{}))
	victim := s.put(1, engine.ZoneMainTheater, testUnit("Asset Target", 1, 1, engine.Keywords{}))
	inPile := s.inst(0, testUnit("Card In Pile", 1, 1, engine.Keywords{}))
	gone := s.inst(0, testUnit("Card Gone From Piles", 1, 1, engine.Keywords{}))
	s.g.Players[0].Discard = []*engine.Instance{inPile}
	s.g.Stack = []*engine.StackItem{{
		Kind: engine.StackAbility, Controller: 0, Source: src.ID, Inst: src.Inst,
		Chosen: [][]int{{victim.ID}, {inPile.ID, gone.ID}},
	}}
	s.g.Stack[0].Effect.Targets = []ability.Target{
		{Filter: ability.Filter{Allegiance: ability.AllegianceOpposing, Kinds: ability.KindUnit}},
		{Filter: ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit, Zone: ability.ZoneDiscard}},
	}
	for seat := engine.PlayerID(0); seat < 2; seat++ {
		v := view.PlayerView(s.g, seat)
		if len(v.Stack) != 1 {
			t.Fatalf("seat %d: stack = %+v", seat, v.Stack)
		}
		it := v.Stack[0]
		if len(it.Chosen) != 2 || !slices.Equal(it.Chosen[0], []int{victim.ID}) || len(it.Chosen[1]) != 0 {
			t.Errorf("seat %d: Chosen is the asset ID for the asset slot and empty for the pile slot: %v", seat, it.Chosen)
		}
		if len(it.ChosenCards) != 2 || len(it.ChosenCards[0]) != 0 ||
			!slices.Equal(it.ChosenCards[1], []string{"Card In Pile"}) {
			t.Errorf("seat %d: ChosenCards is empty for the asset slot and only the name of the card still in a pile for the pile slot: %v", seat, it.ChosenCards)
		}
		js := mustJSON(t, v)
		for _, id := range []int{inPile.ID, gone.ID} {
			if regexp.MustCompile(fmt.Sprintf(`\b%d\b`, id)).MatchString(js) {
				t.Errorf("seat %d: the view carries a card instance ID %d, which is never shown", seat, id)
			}
		}
		if strings.Contains(js, "Card Gone From Piles") {
			t.Errorf("seat %d: a chosen card that left the piles is not named", seat)
		}
	}
}

// T-0122 Case 1: A Maneuver stack item shows its kind, card and asset targets.
func TestView_ManeuverItemShowsItsCardAndTargets(t *testing.T) {
	s := newScene(t)
	opp := s.put(1, engine.ZoneMainTheater, testUnit("Opposing Unit", 1, 1, engine.Keywords{}))
	volley := s.inst(0, &engine.Card{Name: "Alpha Volley", Type: engine.TypeManeuver})
	s.g.Stack = []*engine.StackItem{{
		Kind:       engine.StackManeuver,
		Controller: 0,
		Inst:       volley,
		Source:     0,
		Effect: ability.Effect{
			Targets: []ability.Target{
				{Filter: ability.Filter{Allegiance: ability.AllegianceOpposing, Kinds: ability.KindUnit}},
			},
		},
		Chosen: [][]int{{opp.ID}},
	}}

	for seat := engine.PlayerID(0); seat < 2; seat++ {
		v := view.PlayerView(s.g, seat)
		if len(v.Stack) != 1 {
			t.Fatalf("seat %d: stack len = %d, want 1", seat, len(v.Stack))
		}
		it := v.Stack[0]
		if it.Kind != "Maneuver" {
			t.Errorf("seat %d: Kind = %q, want %q", seat, it.Kind, "Maneuver")
		}
		if it.Controller != 0 {
			t.Errorf("seat %d: Controller = %d, want 0", seat, it.Controller)
		}
		if it.Source != 0 {
			t.Errorf("seat %d: Source = %d, want 0", seat, it.Source)
		}
		if it.Card != "Alpha Volley" {
			t.Errorf("seat %d: Card = %q, want %q", seat, it.Card, "Alpha Volley")
		}
		if !reflect.DeepEqual(it.Chosen, [][]int{{opp.ID}}) {
			t.Errorf("seat %d: Chosen = %v, want [[%d]]", seat, it.Chosen, opp.ID)
		}
		if !reflect.DeepEqual(it.ChosenCards, [][]string{{}}) {
			t.Errorf("seat %d: ChosenCards = %v, want [[]]", seat, it.ChosenCards)
		}
		if it.Keyword != "" {
			t.Errorf("seat %d: Keyword = %q, want empty", seat, it.Keyword)
		}
		if it.Zone != "" {
			t.Errorf("seat %d: Zone = %q, want empty", seat, it.Zone)
		}
		if it.Target != 0 {
			t.Errorf("seat %d: Target = %d, want 0", seat, it.Target)
		}
	}
}

// T-0122 Case 2: A Maneuver's pile slot shows card names, as an Ability's does.
func TestView_ManeuverPileTargetShowsCardNamesNotInstanceIDs(t *testing.T) {
	s := newScene(t)
	volley := s.inst(0, &engine.Card{Name: "Maneuver In Question", Type: engine.TypeManeuver})
	victim := s.put(1, engine.ZoneMainTheater, testUnit("Asset Target", 1, 1, engine.Keywords{}))
	inPile := s.inst(0, testUnit("Card In Pile", 1, 1, engine.Keywords{}))
	gone := s.inst(0, testUnit("Card Gone From Piles", 1, 1, engine.Keywords{}))
	s.g.Players[0].Discard = []*engine.Instance{inPile}
	s.g.Stack = []*engine.StackItem{{
		Kind:       engine.StackManeuver,
		Controller: 0,
		Source:     0,
		Inst:       volley,
		Chosen:     [][]int{{victim.ID}, {inPile.ID, gone.ID}},
		Effect: ability.Effect{
			Targets: []ability.Target{
				{Filter: ability.Filter{Allegiance: ability.AllegianceOpposing, Kinds: ability.KindUnit}},
				{Filter: ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit, Zone: ability.ZoneDiscard}},
			},
		},
	}}

	for seat := engine.PlayerID(0); seat < 2; seat++ {
		v := view.PlayerView(s.g, seat)
		if len(v.Stack) != 1 {
			t.Fatalf("seat %d: stack = %+v", seat, v.Stack)
		}
		it := v.Stack[0]
		if len(it.Chosen) != 2 || !slices.Equal(it.Chosen[0], []int{victim.ID}) || len(it.Chosen[1]) != 0 {
			t.Errorf("seat %d: Chosen is the asset ID for the asset slot and empty for the pile slot: %v", seat, it.Chosen)
		}
		if len(it.ChosenCards) != 2 || len(it.ChosenCards[0]) != 0 ||
			!slices.Equal(it.ChosenCards[1], []string{"Card In Pile"}) {
			t.Errorf("seat %d: ChosenCards is empty for the asset slot and only the name of the card still in a pile for the pile slot: %v", seat, it.ChosenCards)
		}
		js := mustJSON(t, v)
		for _, id := range []int{inPile.ID, gone.ID} {
			if regexp.MustCompile(fmt.Sprintf(`\b%d\b`, id)).MatchString(js) {
				t.Errorf("seat %d: the view carries a card instance ID %d, which is never shown", seat, id)
			}
		}
		if strings.Contains(js, "Card Gone From Piles") {
			t.Errorf("seat %d: a chosen card that left the piles is not named", seat)
		}
	}
}

// T-0122 Case 5: After resolution, the card is in its owner's discard pile in both views, with no new event.
func TestView_EnactedManeuverEndsInItsOwnersDiscard(t *testing.T) {
	s := newScene(t)
	s.g.Players[0].Req, s.g.Players[0].Tac = 2, 2
	deckCard := s.inst(0, testUnit("Drawn Card", 1, 1, engine.Keywords{}))
	s.g.Players[0].Deck = []*engine.Instance{deckCard}
	maneuverCard := &engine.Card{
		Name: "Tactical Volley", Type: engine.TypeManeuver, ReqCost: 1, TacCost: 0,
		Ability: &ability.Def{Effect: ability.Effect{Steps: []ability.Step{
			{Op: ability.OpDraw, Amount: ability.Value{N: 1}},
		}}},
	}
	mInst := s.inst(0, maneuverCard)
	s.g.Players[0].Hand = []*engine.Instance{mInst}
	initialHandCount := len(s.g.Players[0].Hand)

	err := s.g.Apply(engine.Action{Kind: engine.ActEnact, Player: 0, Card: mInst.ID})
	if err != nil {
		t.Fatalf("apply ActEnact: %v", err)
	}

	v0 := view.PlayerView(s.g, 0)
	v1 := view.PlayerView(s.g, 1)

	if len(v0.You.Discard) == 0 || v0.You.Discard[len(v0.You.Discard)-1].Card != "Tactical Volley" {
		t.Errorf("seat 0 You.Discard does not end with Tactical Volley: %+v", v0.You.Discard)
	}
	if len(v1.Opponent.Discard) == 0 || v1.Opponent.Discard[len(v1.Opponent.Discard)-1].Card != "Tactical Volley" {
		t.Errorf("seat 1 Opponent.Discard does not end with Tactical Volley: %+v", v1.Opponent.Discard)
	}
	if len(v0.Stack) != 0 || len(v1.Stack) != 0 {
		t.Errorf("views have non-empty stack: seat 0 %+v, seat 1 %+v", v0.Stack, v1.Stack)
	}
	if v0.You.HandCount != initialHandCount {
		t.Errorf("seat 0 You.HandCount = %d, want %d", v0.You.HandCount, initialHandCount)
	}

	for seat := engine.PlayerID(0); seat < 2; seat++ {
		lines := view.LogLines(s.g.Log, seat)
		found := false
		for _, l := range lines {
			if strings.Contains(l, "Tactical Volley") {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("seat %d: LogLines has no line containing Tactical Volley", seat)
		}
	}
}

// T-0122 Case 6: viewtest counts a Maneuver on the stack as public.
func TestViewtest_ManeuverOnTheStackIsPublic(t *testing.T) {
	s := newScene(t)
	m := s.inst(0, &engine.Card{Name: "Alpha Volley", Type: engine.TypeManeuver})
	s.g.Stack = []*engine.StackItem{{
		Kind: engine.StackManeuver, Controller: 0, Inst: m,
	}}
	pub := viewtest.PublicNames(s.g)
	if !pub["Alpha Volley"] {
		t.Errorf("PublicNames does not include Maneuver on stack Alpha Volley")
	}
	hid := viewtest.HiddenNames(s.g, 0)
	if hid["Alpha Volley"] {
		t.Errorf("HiddenNames includes Maneuver on stack Alpha Volley")
	}
}

// T-0159 case 6: an activated ability on the stack, seen through the real
// action path. The ability's target is chosen on declaration; the source's
// "whenever this exhausts" trigger then asks for a target of its own while the
// activated item is on the stack.
func TestView_ActivatedAbilityItemShowsItsTargets(t *testing.T) {
	killOpposing := ability.Effect{
		Targets: []ability.Target{{Filter: ability.Filter{Allegiance: ability.AllegianceOpposing, Kinds: ability.KindUnit}}},
		Steps:   []ability.Step{{Op: ability.OpKill, Subject: ability.SubjectTarget, Target: 0}},
	}
	src := testUnit("Tapper", 1, 5, engine.Keywords{})
	src.RulesText = "exhaust: kill chosen opposing unit"
	src.Ability = &ability.Def{
		Triggers:  []ability.Trigger{{Hook: ability.HookExhausted, Effect: killOpposing}},
		Activated: []ability.Activated{{Costs: []ability.Cost{{Kind: ability.CostExhaustSelf}}, Effect: killOpposing}},
	}

	var picked int
	var views [2]view.View
	asked := 0
	chooser := func(g *engine.Game, d engine.Decision) engine.Answer {
		asked++
		if len(g.Stack) == 0 {
			picked = d.Options[0].Ref
			return engine.Answer{Picks: []int{0}}
		}
		for seat := range views {
			views[seat] = view.PlayerView(g, engine.PlayerID(seat))
		}
		return engine.Answer{Picks: []int{len(d.Options) - 1}}
	}
	filler := make([]*engine.Card, 10)
	for i := range filler {
		filler[i] = testUnit("Filler", 1, 1, engine.Keywords{})
	}
	g, err := engine.NewGame(engine.Config{
		Seed:      1,
		Decks:     [2]engine.DeckList{{HQ: testHQ("Alpha HQ"), Main: filler}, {HQ: testHQ("Beta HQ"), Main: filler}},
		PlayFirst: func(decider engine.PlayerID) bool { return decider == 0 },
		Choosers:  [2]engine.Chooser{chooser, chooser},
	})
	if err != nil {
		t.Fatal(err)
	}
	s := &scene{g: g, nextID: 100000}
	a := s.put(0, engine.ZoneMainTheater, src)
	s.put(1, engine.ZoneMainTheater, testUnit("Foe One", 1, 5, engine.Keywords{}))
	s.put(1, engine.ZoneMainTheater, testUnit("Foe Two", 1, 5, engine.Keywords{}))
	asked = 0
	if err := g.Apply(engine.Action{Kind: engine.ActActivate, Player: 0, Asset: a.ID, Ability: 0}); err != nil {
		t.Fatal(err)
	}
	if asked < 2 {
		t.Fatalf("the chooser was asked %d times, want the declaration and the trigger's target", asked)
	}
	for seat, v := range views {
		var item *view.StackView
		for i := range v.Stack {
			if v.Stack[i].Source == a.ID && v.Stack[i].Kind == "Ability" {
				item = &v.Stack[i]
			}
		}
		if item == nil {
			t.Fatalf("seat %d: no activated item on the stack view: %+v", seat, v.Stack)
		}
		if item.Card != "Tapper" || item.Keyword != "" || !reflect.DeepEqual(item.Chosen, [][]int{{picked}}) {
			t.Errorf("seat %d: item %+v, want Card Tapper, empty Keyword, Chosen [[%d]]", seat, *item, picked)
		}
	}
}

// T-0161 Case 1: An asset's printed rules text is visible in view.AssetView.
func TestView_AssetTextIsPrinted(t *testing.T) {
	s := newScene(t)
	g := s.g
	c0 := testUnit("Unit Zero", 2, 2, engine.Keywords{})
	c0.RulesText = "exhaust: test text"
	a0 := s.put(0, engine.ZoneReserves, c0)

	c1 := testUnit("Unit One", 3, 3, engine.Keywords{})
	c1.RulesText = "T1: other text"
	a1 := s.put(1, engine.ZoneMainTheater, c1)

	for seat := range engine.PlayerID(2) {
		v := view.PlayerView(g, seat)

		findInReserves := func(side view.Side, id int) *view.AssetView {
			for i := range side.Reserves {
				if side.Reserves[i].ID == id {
					return &side.Reserves[i]
				}
			}
			return nil
		}
		findInTheater := func(id int) *view.AssetView {
			for i := range v.Theater {
				if v.Theater[i].ID == id {
					return &v.Theater[i]
				}
			}
			return nil
		}

		var got0 *view.AssetView
		if seat == 0 {
			got0 = findInReserves(v.You, a0.ID)
		} else {
			got0 = findInReserves(v.Opponent, a0.ID)
		}
		if got0 == nil {
			t.Fatalf("seat %d: asset 0 not found in reserves", seat)
		}
		if got0.Text != "exhaust: test text" {
			t.Errorf("seat %d: asset 0 Text = %q, want %q", seat, got0.Text, "exhaust: test text")
		}

		got1 := findInTheater(a1.ID)
		if got1 == nil {
			t.Fatalf("seat %d: asset 1 not found in theater", seat)
		}
		if got1.Text != "T1: other text" {
			t.Errorf("seat %d: asset 1 Text = %q, want %q", seat, got1.Text, "T1: other text")
		}
	}
}

// T-0161 Case 2: A face-down asset hides its printed rules text.
func TestView_FaceDownAssetHidesItsText(t *testing.T) {
	s := newScene(t)
	g := s.g
	hidden := testUnit("Alpha Secret Agent", 2, 2, engine.Keywords{})
	hidden.RulesText = "Alpha secret text"
	fd := &engine.Asset{
		ID:         s.id(),
		Inst:       &engine.Instance{ID: s.id(), Card: hidden, Owner: 0},
		Controller: 0,
		Zone:       engine.ZoneMainTheater,
		FaceDown:   true,
	}
	g.Theater = append(g.Theater, fd)

	for seat := range engine.PlayerID(2) {
		v := view.PlayerView(g, seat)
		js := mustJSON(t, v)

		var got *view.AssetView
		for i := range v.Theater {
			if v.Theater[i].ID == fd.ID {
				got = &v.Theater[i]
				break
			}
		}
		if got == nil {
			t.Fatalf("seat %d: face-down asset %d missing from theater", seat, fd.ID)
		}
		if got.Text != "" {
			t.Errorf("seat %d: face-down asset Text = %q, want empty string", seat, got.Text)
		}
		if strings.Contains(js, "Alpha secret text") {
			t.Errorf("seat %d's JSON view reveals face-down card's rules text", seat)
		}
	}
}

// T-0181 (directive #117 (a)): an asset's view shows a keyword behind Cornered
// to both seats, and only while its controller is Cornered (SVY-R18).
func TestView_KeywordBehindCorneredShowsOnlyWhileCornered(t *testing.T) {
	s := newScene(t)
	g := s.g
	c := testUnit("Alpha Brawler", 2, 2, engine.Keywords{})
	c.Ability = &ability.Def{Cornered: ability.CorneredKeywords{Aggressive: 3}}
	brawler := s.put(0, engine.ZoneMainTheater, c)
	s.put(1, engine.ZoneMainTheater, testUnit("Beta One", 1, 1, engine.Keywords{}))

	keywords := func(seat engine.PlayerID) []string {
		t.Helper()
		v := view.PlayerView(g, seat)
		for _, a := range v.Theater {
			if a.ID == brawler.ID {
				return a.Keywords
			}
		}
		t.Fatalf("seat %d: asset %d missing from the theater %+v", seat, brawler.ID, v.Theater)
		return nil
	}
	want := func(when string, kw []string) {
		t.Helper()
		for seat := range engine.PlayerID(2) {
			if got := keywords(seat); !slices.Equal(got, kw) {
				t.Errorf("%s: seat %d sees keywords %v, want %v", when, seat, got, kw)
			}
		}
	}

	want("1 against 1, not Cornered", nil)
	second := s.put(1, engine.ZoneMainTheater, testUnit("Beta Two", 1, 1, engine.Keywords{}))
	want("1 against 2, Cornered", []string{"Aggressive 3"})
	g.Theater = slices.DeleteFunc(g.Theater, func(a *engine.Asset) bool { return a == second })
	want("back to 1 against 1, not Cornered", nil)
}

// T-0182 (directive #118 (a)): each seat's own hand cards carry what they cost
// now, beside the printed cost. A discard card carries the printed cost and no
// cost now. No card changes a cost yet, so now equals printed until #118 (b).
func TestView_HandCardCarriesItsCostNow(t *testing.T) {
	s := newContentScene(t)
	for seat := range engine.PlayerID(2) {
		v := view.PlayerView(s.g, seat)
		if len(v.You.Hand) == 0 || len(v.You.Discard) == 0 || len(v.Opponent.Discard) == 0 {
			t.Fatalf("seat %d: hand %+v, discards %+v and %+v, want a card in each", seat, v.You.Hand, v.You.Discard, v.Opponent.Discard)
		}
		for _, h := range v.You.Hand {
			if h.ReqNow != 2 || h.TacNow != 1 || h.ReqCost != 2 || h.TacCost != 1 {
				t.Errorf("seat %d: hand card %+v, want R2 T1 now and printed", seat, h)
			}
		}
		for _, c := range append(slices.Clone(v.You.Discard), v.Opponent.Discard...) {
			if c.ReqNow != 0 || c.TacNow != 0 || c.ReqCost != 2 || c.TacCost != 1 {
				t.Errorf("seat %d: discard card %+v, want printed R2 T1 and no cost now", seat, c)
			}
		}
	}
}

// T-0183 (directive #119): a hidden decision of any kind names no option and no
// pick to the other seat, and its own seat sees the whole line.
func TestView_HiddenDecisionLineNamesNoOptionAndNoPick(t *testing.T) {
	for p := range engine.PlayerID(2) {
		ev := engine.Event{
			Seq:    40,
			Turn:   3,
			Phase:  "main",
			Kind:   engine.EventDecision,
			Player: int(p),
			Decision: &engine.DecisionRecord{
				Kind:    engine.DecideTarget.String(),
				Player:  int(p),
				Prompt:  "Look at Alpha Secret",
				How:     "answer",
				Hidden:  true,
				Options: []string{"Alpha Secret", "Beta Secret"},
				Refs:    []int{7, 8},
				Picks:   []int{1},
			},
		}
		opp := p.Opponent()
		got := view.LogLines([]engine.Event{ev}, opp)
		want := fmt.Sprintf("40 T3 main decision: player %d makes a hidden choice", p)
		if len(got) != 1 || got[0] != want {
			t.Errorf("seat %d seeing player %d's hidden choice: got %v, want [%s]", opp, p, got, want)
		}
		own := view.LogLines([]engine.Event{ev}, p)
		wantOwn := eventlog.FormatEvent(ev)
		if len(own) != 1 || own[0] != wantOwn || !strings.Contains(own[0], "Beta Secret") {
			t.Errorf("seat %d seeing its own hidden choice: got %v, want [%s]", p, own, wantOwn)
		}
	}
}

// T-0183 (directive #119): a search record with the flag keeps its wording.
func TestView_FlaggedSearchKeepsItsWording(t *testing.T) {
	for _, tc := range []struct {
		picks []int
		want  string
	}{
		{[]int{1}, "40 T3 main decision: player 0 searches their deck and finds a card"},
		{[]int{}, "40 T3 main decision: player 0 searches their deck and finds no card"},
	} {
		ev := engine.Event{
			Seq:    40,
			Turn:   3,
			Phase:  "main",
			Kind:   engine.EventDecision,
			Player: 0,
			Decision: &engine.DecisionRecord{
				Kind:    engine.DecideSearch.String(),
				Player:  0,
				Prompt:  "Search your deck for a card",
				How:     "answer",
				Hidden:  true,
				Options: []string{"Alpha Secret", "Beta Secret"},
				Refs:    []int{7, 8},
				Picks:   tc.picks,
			},
		}
		got := view.LogLines([]engine.Event{ev}, 1)
		if len(got) != 1 || got[0] != tc.want {
			t.Errorf("picks %v: got %v, want [%s]", tc.picks, got, tc.want)
		}
	}
}

// T-0183 (directive #119): a decision that isn't hidden prints as before, to both
// seats.
func TestView_OpenDecisionLinePrintsToBothSeats(t *testing.T) {
	ev := engine.Event{
		Seq:    40,
		Turn:   3,
		Phase:  "main",
		Kind:   engine.EventDecision,
		Player: 0,
		Decision: &engine.DecisionRecord{
			Kind:    engine.DecideTarget.String(),
			Player:  0,
			Prompt:  "Choose a target",
			How:     "answer",
			Options: []string{"Alpha Open", "Beta Open"},
			Refs:    []int{7, 8},
			Picks:   []int{1},
		},
	}
	want := eventlog.FormatEvent(ev)
	for seat := range engine.PlayerID(2) {
		got := view.LogLines([]engine.Event{ev}, seat)
		if len(got) != 1 || got[0] != want || !strings.Contains(got[0], "Beta Open") {
			t.Errorf("seat %d: got %v, want [%s]", seat, got, want)
		}
	}
}

// T-0168 Cases 1-4: The play log hides an opponent's deck search pick.

func TestView_SearchLineHidesThePickFromTheOpponent(t *testing.T) {
	for p := range engine.PlayerID(2) {
		ev := engine.Event{
			Seq:    40,
			Turn:   3,
			Phase:  "main",
			Kind:   engine.EventDecision,
			Player: int(p),
			Decision: &engine.DecisionRecord{
				Kind:    engine.DecideSearch.String(),
				Player:  int(p),
				Prompt:  "Search your deck for a card",
				How:     "answer",
				Options: []string{"Alpha Secret", "Beta Secret"},
				Refs:    []int{7, 8},
				Picks:   []int{1},
				Hidden:  true,
			},
		}
		opp := p.Opponent()
		got := view.LogLines([]engine.Event{ev}, opp)
		want := fmt.Sprintf("40 T3 main decision: player %d searches their deck and finds a card", p)
		if len(got) != 1 || got[0] != want {
			t.Errorf("seat %d seeing player %d search: got %v, want [%s]", opp, p, got, want)
		}
	}
}

func TestView_SearchLineShowsThePickToTheSearcher(t *testing.T) {
	for p := range engine.PlayerID(2) {
		ev := engine.Event{
			Seq:    40,
			Turn:   3,
			Phase:  "main",
			Kind:   engine.EventDecision,
			Player: int(p),
			Decision: &engine.DecisionRecord{
				Kind:    engine.DecideSearch.String(),
				Player:  int(p),
				Prompt:  "Search your deck for a card",
				How:     "answer",
				Options: []string{"Alpha Secret", "Beta Secret"},
				Refs:    []int{7, 8},
				Picks:   []int{1},
			},
		}
		got := view.LogLines([]engine.Event{ev}, p)
		want := eventlog.FormatEvent(ev)
		if len(got) != 1 || got[0] != want {
			t.Errorf("searcher seat %d: got %v, want [%s]", p, got, want)
		}
		if len(got) > 0 && !strings.Contains(got[0], "Beta Secret") {
			t.Errorf("searcher seat %d: line %q does not contain Beta Secret", p, got[0])
		}
	}

	handEv := engine.Event{
		Seq:    40,
		Turn:   3,
		Phase:  "main",
		Kind:   engine.EventDecision,
		Player: 0,
		Decision: &engine.DecisionRecord{
			Kind:    engine.DecideHand.String(),
			Player:  0,
			Prompt:  "Search your deck for a card",
			How:     "answer",
			Options: []string{"Alpha Secret", "Beta Secret"},
			Refs:    []int{7, 8},
			Picks:   []int{1},
		},
	}
	wantHand := eventlog.FormatEvent(handEv)
	for seat := range engine.PlayerID(2) {
		got := view.LogLines([]engine.Event{handEv}, seat)
		if len(got) != 1 || got[0] != wantHand {
			t.Errorf("seat %d seeing DecideHand: got %v, want [%s]", seat, got, wantHand)
		}
	}
}

func TestView_SearchLineSaysNoCardWasFound(t *testing.T) {
	ev := engine.Event{
		Seq:    40,
		Turn:   3,
		Phase:  "main",
		Kind:   engine.EventDecision,
		Player: 0,
		Decision: &engine.DecisionRecord{
			Kind:    engine.DecideSearch.String(),
			Player:  0,
			Prompt:  "Search your deck for a card",
			How:     "answer",
			Options: []string{"Alpha Secret", "Beta Secret"},
			Refs:    []int{7, 8},
			Picks:   []int{},
			Hidden:  true,
		},
	}
	got := view.LogLines([]engine.Event{ev}, 1)
	want := "40 T3 main decision: player 0 searches their deck and finds no card"
	if len(got) != 1 || got[0] != want {
		t.Errorf("seat 1 seeing empty search: got %v, want [%s]", got, want)
	}
}

// T-0186 (directive #123): the flag alone hides a decision, so a search record
// without it prints whole, like any other record. Whether a search is secret is
// said once, where the engine sets the flag.
func TestView_SearchWithoutTheFlagPrintsWhole(t *testing.T) {
	ev := engine.Event{
		Seq:    40,
		Turn:   3,
		Phase:  "main",
		Kind:   engine.EventDecision,
		Player: 0,
		Decision: &engine.DecisionRecord{
			Kind:    engine.DecideSearch.String(),
			Player:  0,
			Prompt:  "Search your deck for a card",
			How:     "answer",
			Options: []string{"Alpha Secret", "Beta Secret"},
			Refs:    []int{7, 8},
			Picks:   []int{1},
		},
	}
	got := view.LogLines([]engine.Event{ev}, 1)
	want := eventlog.FormatEvent(ev)
	if len(got) != 1 || got[0] != want {
		t.Errorf("seat 1 seeing a search with no flag: got %v, want [%s]", got, want)
	}
}

func TestView_SearchInAGameStaysPrivate(t *testing.T) {
	chooser := func(g *engine.Game, d engine.Decision) engine.Answer {
		return engine.Answer{Picks: []int{1}}
	}
	filler := make([]*engine.Card, 10)
	for i := range filler {
		filler[i] = testUnit("Filler", 1, 1, engine.Keywords{})
	}
	g, err := engine.NewGame(engine.Config{
		Seed:      1,
		Decks:     [2]engine.DeckList{{HQ: testHQ("Alpha HQ"), Main: filler}, {HQ: testHQ("Beta HQ"), Main: filler}},
		PlayFirst: func(decider engine.PlayerID) bool { return decider == 0 },
		Choosers:  [2]engine.Chooser{chooser, chooser},
	})
	if err != nil {
		t.Fatal(err)
	}
	s := &scene{g: g, nextID: 100000}
	g.Players[0].Deck = []*engine.Instance{
		s.inst(0, testUnit("Alpha Search One", 1, 1, engine.Keywords{})),
		s.inst(0, testUnit("Alpha Search Two", 1, 1, engine.Keywords{})),
		s.inst(0, testUnit("Alpha Search Three", 1, 1, engine.Keywords{})),
	}
	src := testUnit("Search Unit", 1, 5, engine.Keywords{})
	src.Ability = &ability.Def{
		Activated: []ability.Activated{
			{
				Costs:  []ability.Cost{{Kind: ability.CostExhaustSelf}},
				Effect: ability.Effect{Steps: []ability.Step{{Op: ability.OpSearchDeck}}},
			},
		},
	}
	a := s.put(0, engine.ZoneMainTheater, src)

	if err := g.Apply(engine.Action{Kind: engine.ActActivate, Player: 0, Asset: a.ID, Ability: 0}); err != nil {
		t.Fatal(err)
	}

	foundInHand := false
	for _, inst := range g.Players[0].Hand {
		if inst.Card.Name == "Alpha Search Two" {
			foundInHand = true
			break
		}
	}
	if !foundInHand {
		t.Fatalf("Alpha Search Two not found in player 0's hand")
	}

	lines1 := view.LogLines(g.Log, 1)
	secretNames := []string{"Alpha Search One", "Alpha Search Two", "Alpha Search Three"}
	for _, line := range lines1 {
		for _, name := range secretNames {
			if strings.Contains(line, name) {
				t.Errorf("seat 1 log line contains leaked secret name %q: %s", name, line)
			}
		}
	}

	lines0 := view.LogLines(g.Log, 0)
	searchDecIdx := -1
	shuffleIdx := -1
	for i, e := range g.Log {
		if e.Decision != nil && e.Decision.Kind == engine.DecideSearch.String() {
			searchDecIdx = i
		}
		if e.Kind == engine.EventShuffle && e.Player == 0 {
			shuffleIdx = i
		}
	}
	if searchDecIdx == -1 {
		t.Fatal("no search decision event found in game log")
	}
	if shuffleIdx == -1 {
		t.Fatal("no shuffle event found in game log")
	}

	decEv := g.Log[searchDecIdx]
	if lines0[searchDecIdx] != eventlog.FormatEvent(decEv) {
		t.Errorf("seat 0 search decision line: got %q, want %q", lines0[searchDecIdx], eventlog.FormatEvent(decEv))
	}
	if !strings.Contains(lines0[searchDecIdx], "Alpha Search Two") {
		t.Errorf("seat 0 search decision line does not contain 'Alpha Search Two': %q", lines0[searchDecIdx])
	}
	wantSuffix := "player 0 searches their deck and finds a card"
	if !strings.HasSuffix(lines1[searchDecIdx], wantSuffix) {
		t.Errorf("seat 1 search decision line does not end with %q: got %q", wantSuffix, lines1[searchDecIdx])
	}

	shufEv := g.Log[shuffleIdx]
	wantShuf := eventlog.FormatEvent(shufEv)
	if lines0[shuffleIdx] != wantShuf || !strings.Contains(lines0[shuffleIdx], "player 0 shuffles their deck") {
		t.Errorf("seat 0 shuffle line: got %q, want %q", lines0[shuffleIdx], wantShuf)
	}
	if lines1[shuffleIdx] != wantShuf || !strings.Contains(lines1[shuffleIdx], "player 0 shuffles their deck") {
		t.Errorf("seat 1 shuffle line: got %q, want %q", lines1[shuffleIdx], wantShuf)
	}
}
