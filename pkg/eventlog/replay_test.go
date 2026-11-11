package eventlog_test

import (
	"bytes"
	"fmt"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime/debug"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"

	"animus-mundi/pkg/engine"
	"animus-mundi/pkg/engine/ability"
	"animus-mundi/pkg/eventlog"
	"animus-mundi/pkg/sim"
)

// Tests for T-0023: ReadEventsJSONL and Replay. Each TestReplay_Case<N> is
// Acceptance Case N. The records are built in-process with sim.SimulateWith and
// go through JSON (WriteEventsJSONL, ReadEventsJSONL) before they are replayed.

type collector struct{ events []engine.Event }

func (c *collector) Event(e engine.Event)      { c.events = append(c.events, e) }
func (c *collector) Settled(s engine.Snapshot) {}

var (
	starterNames = [2]string{"data/decks/animus-starter.txt", "data/decks/path-starter.txt"}
	stressNames  = [2]string{"stress:SOULLESS", "stress:CONCORDIA"}
	root         = filepath.Join("..", "..")

	cardsOnce sync.Once
	cardMap   map[string]*engine.Card
	cardErr   error

	gamesOnce sync.Once
	sample    [][]engine.Event // 20 starter games, then 5 stress games
	gamesErr  error
)

func loadCards(t *testing.T) map[string]*engine.Card {
	t.Helper()
	cardsOnce.Do(func() { cardMap, cardErr = engine.LoadCards(filepath.Join(root, "data", "cards")) })
	if cardErr != nil {
		t.Fatalf("load cards: %v", cardErr)
	}
	return cardMap
}

func resolveDeck(t *testing.T) func(string) (engine.DeckList, error) {
	cards := loadCards(t)
	return func(name string) (engine.DeckList, error) {
		if f, ok := strings.CutPrefix(name, "stress:"); ok {
			return sim.FactionDeck(cards, f)
		}
		return engine.LoadDeck(filepath.Join(root, filepath.FromSlash(name)), cards)
	}
}

// play simulates games and returns each game's events after a JSON round trip.
func play(decks [2]engine.DeckList, names [2]string, games int, seed uint64) ([][]engine.Event, error) {
	var cols []*collector
	opt := sim.SimOptions{DeckNames: names, Policy: sim.ProgressPolicy, Observer: func(int, uint64) engine.Observer {
		c := &collector{}
		cols = append(cols, c)
		return c
	}}
	if _, err := sim.SimulateWith(decks, games, seed, 400, opt); err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	for i, c := range cols {
		if err := eventlog.WriteEventsJSONL(&buf, i+1, c.events); err != nil {
			return nil, err
		}
	}
	byGame, err := eventlog.ReadEventsJSONL(&buf)
	if err != nil {
		return nil, err
	}
	out := make([][]engine.Event, games)
	for g := 1; g <= games; g++ {
		out[g-1] = byGame[g]
	}
	return out, nil
}

func allGames(t *testing.T) [][]engine.Event {
	t.Helper()
	cards := loadCards(t)
	gamesOnce.Do(func() {
		var sd, xd [2]engine.DeckList
		sd, gamesErr = sim.SimDecks(cards, filepath.Join(root, "data", "decks"), "starter", [2]string{"ANIMUS", "PATH"})
		if gamesErr != nil {
			return
		}
		for i, f := range []string{"SOULLESS", "CONCORDIA"} {
			if xd[i], gamesErr = sim.FactionDeck(cards, f); gamesErr != nil {
				return
			}
		}
		var starters, stress [][]engine.Event
		if starters, gamesErr = play(sd, starterNames, 20, 100); gamesErr != nil {
			return
		}
		if stress, gamesErr = play(xd, stressNames, 5, 500); gamesErr != nil {
			return
		}
		sample = append(starters, stress...)
	})
	if gamesErr != nil {
		t.Fatalf("building sample records: %v", gamesErr)
	}
	return sample
}

func TestReplay_ReadEventsJSONL(t *testing.T) {
	games := allGames(t)
	if len(games) != 25 {
		t.Fatalf("%d games, want 25", len(games))
	}
	for i, evs := range games {
		if len(evs) == 0 || evs[0].Kind != "match-start" || evs[len(evs)-1].Kind != "match-end" {
			t.Fatalf("game %d: bad record bounds (%d events)", i+1, len(evs))
		}
		for j, e := range evs {
			if e.Seq != j+1 {
				t.Fatalf("game %d: event %d has seq %d", i+1, j, e.Seq)
			}
		}
	}
	if st := games[0][0].Start; st == nil || st.MulliganInsts[0] == nil || st.MulliganInsts[1] == nil {
		t.Errorf("MulliganInsts must be non-nil on both seats: %+v", st)
	}
}

func TestReplay_Case1_TurnStartsRoundTrip(t *testing.T) {
	resolve := resolveDeck(t)
	for gi, evs := range allGames(t) {
		n := 0
		for _, e := range evs {
			if e.Kind != "turn-start" {
				continue
			}
			n++
			got, err := eventlog.Replay(evs, e.Seq, resolve)
			if err != nil {
				t.Fatalf("game %d turn-start seq %d: %v", gi+1, e.Seq, err)
			}
			if !reflect.DeepEqual(got, *e.Snapshot) {
				t.Fatalf("game %d turn-start seq %d: replayed snapshot differs from the stored one", gi+1, e.Seq)
			}
		}
		if n == 0 {
			t.Fatalf("game %d has no turn-start", gi+1)
		}
	}
}

func TestReplay_Case2_WholeGameReproduces(t *testing.T) {
	resolve := resolveDeck(t)
	for gi, evs := range allGames(t) {
		last := evs[len(evs)-1]
		got, err := eventlog.Replay(evs, last.Seq, resolve)
		if err != nil {
			t.Fatalf("game %d: %v", gi+1, err)
		}
		end := last.End
		if end == nil {
			t.Fatalf("game %d: last event is not match-end", gi+1)
		}
		if !got.Over || got.Draw != end.Draw || got.Winner != end.Winner || got.Reason != end.Reason {
			t.Errorf("game %d: replayed over=%t draw=%t winner=%d reason=%q, recorded %+v",
				gi+1, got.Over, got.Draw, got.Winner, got.Reason, *end)
		}
	}
}

func TestReplay_Case3_MidGamePoint(t *testing.T) {
	resolve := resolveDeck(t)
	found := false
	for gi, evs := range allGames(t) {
		for _, e := range evs {
			if e.Kind != "damage" {
				continue
			}
			found = true
			a, err := eventlog.Replay(evs, e.Seq, resolve)
			if err != nil {
				t.Fatalf("game %d damage seq %d: %v", gi+1, e.Seq, err)
			}
			if a.Turn != e.Turn {
				t.Errorf("game %d seq %d: snapshot turn %d, event turn %d", gi+1, e.Seq, a.Turn, e.Turn)
			}
			b, err := eventlog.Replay(evs, e.Seq, resolve)
			if err != nil || !reflect.DeepEqual(a, b) {
				t.Errorf("game %d seq %d: two replays differ (err %v)", gi+1, e.Seq, err)
			}
			break
		}
		if found {
			break
		}
	}
	if !found {
		t.Fatal("no damage event in the sample games")
	}

	// at 1 is the state right after match-start, before turn 1 begins.
	evs := allGames(t)[0]
	got, err := eventlog.Replay(evs, 1, resolve)
	if err != nil {
		t.Fatalf("at 1: %v", err)
	}
	want := *evs[1].Snapshot // only scalar fields change, so a shallow copy is enough
	want.Turn, want.Phase, want.Active = 0, "setup", 0
	if !reflect.DeepEqual(got, want) {
		t.Errorf("at 1: snapshot is not the turn-1 snapshot with Turn 0, Phase setup, Active 0")
	}
	if got2, err := eventlog.Replay(evs, 2, resolve); err != nil || !reflect.DeepEqual(got2, *evs[1].Snapshot) {
		t.Errorf("at 2: want the stored turn-start snapshot (err %v)", err)
	}
}

var seqInErr = regexp.MustCompile(`seq (\d+)`)

// errSeq is the Seq an error names, or -1.
func errSeq(err error) int {
	m := seqInErr.FindStringSubmatch(err.Error())
	if m == nil {
		return -1
	}
	n, _ := strconv.Atoi(m[1])
	return n
}

// withEvent returns events with the event at index i replaced by e.
func withEvent(events []engine.Event, i int, e engine.Event) []engine.Event {
	out := append([]engine.Event(nil), events...)
	out[i] = e
	return out
}

// assertDiverges checks that replaying mutated to its end names a Seq >= seq,
// and that replaying to the Seq before the change still works.
func assertDiverges(t *testing.T, mutated []engine.Event, seq int, resolve func(string) (engine.DeckList, error)) {
	t.Helper()
	_, err := eventlog.Replay(mutated, mutated[len(mutated)-1].Seq, resolve)
	if err == nil {
		t.Fatalf("change at seq %d: replay to the end succeeded", seq)
	}
	if got := errSeq(err); got < seq {
		t.Errorf("change at seq %d: error names seq %d: %v", seq, got, err)
	}
	if _, err := eventlog.Replay(mutated, seq-1, resolve); err != nil {
		t.Errorf("replay to seq %d, before the change, failed: %v", seq-1, err)
	}
}

func TestReplay_Case4_DivergenceIsCaught_AttackTarget(t *testing.T) {
	resolve := resolveDeck(t)
	for gi, evs := range allGames(t) {
		var assets []int // asset IDs of the latest turn-start
		for i, e := range evs {
			if e.Snapshot != nil {
				assets = assets[:0]
				for _, p := range e.Snapshot.Players {
					if p.HQ != nil {
						assets = append(assets, p.HQ.ID)
					}
					for _, r := range p.Reserves {
						assets = append(assets, r.ID)
					}
				}
				for _, a := range e.Snapshot.Theater {
					assets = append(assets, a.ID)
				}
			}
			if e.Kind != "action" || !strings.HasPrefix(e.Detail, "Attack") {
				continue
			}
			other := 0
			for _, id := range assets {
				if id != e.Target && id != e.Source {
					other = id
					break
				}
			}
			if other == 0 {
				continue
			}
			m := e
			m.Target = other
			t.Logf("game %d: attack at seq %d retargeted %d -> %d", gi+1, e.Seq, e.Target, other)
			assertDiverges(t, withEvent(evs, i, m), e.Seq, resolve)
			return
		}
	}
	t.Fatal("no Attack action in the sample games")
}

func TestReplay_Case4_DivergenceIsCaught_AnswerPick(t *testing.T) {
	resolve := resolveDeck(t)
	for gi, evs := range allGames(t) {
		for i, e := range evs {
			d := e.Decision
			if d == nil || d.How != "answer" || len(d.Options) < 2 || len(d.Picks) != 1 {
				continue
			}
			m := e
			dd := *d
			dd.Picks = []int{(d.Picks[0] + 1) % len(d.Options)}
			m.Decision = &dd
			t.Logf("game %d: decision at seq %d pick %v -> %v", gi+1, e.Seq, d.Picks, dd.Picks)
			assertDiverges(t, withEvent(evs, i, m), e.Seq, resolve)
			return
		}
	}
	t.Skip("no answered multi-option decision in the sample games")
}

func TestReplay_Case5_BuildMismatch(t *testing.T) {
	evs := allGames(t)[0]
	st := *evs[0].Start
	st.Revision = "not-this-build"
	m := evs[0]
	m.Start = &st
	_, err := eventlog.Replay(withEvent(evs, 0, m), evs[1].Seq, resolveDeck(t))
	if err == nil {
		t.Fatal("a record from another build must be refused")
	}
	rev := ""
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, s := range info.Settings {
			if s.Key == "vcs.revision" {
				rev = s.Value
			}
		}
	}
	for _, want := range []string{fmt.Sprintf("%q", "not-this-build"), fmt.Sprintf("%q", rev)} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error should name %s: %v", want, err)
		}
	}
}

func TestReplay_Case6_BadInput(t *testing.T) {
	resolve := resolveDeck(t)
	evs := allGames(t)[0]
	last := evs[len(evs)-1].Seq
	for _, at := range []int{0, -1, last + 1, last + 1000} {
		if _, err := eventlog.Replay(evs, at, resolve); err == nil {
			t.Errorf("at %d: want an error", at)
		}
	}
	if _, err := eventlog.Replay(evs[1:], 2, resolve); err == nil {
		t.Error("events without a leading match-start: want an error")
	}
	if _, err := eventlog.Replay(nil, 1, resolve); err == nil {
		t.Error("no events: want an error")
	}
	bad := func(name string) (engine.DeckList, error) {
		return engine.DeckList{}, fmt.Errorf("unknown deck %q", name)
	}
	if _, err := eventlog.Replay(evs, last, bad); err == nil || !strings.Contains(err.Error(), "unknown deck") {
		t.Errorf("a resolveDeck error must come back: %v", err)
	}
}

func TestReplay_Case6_InvalidAnswerRefused(t *testing.T) {
	resolve := resolveDeck(t)
	for _, evs := range allGames(t) {
		for i, e := range evs {
			if e.Decision == nil || e.Decision.How == "auto" {
				continue
			}
			m := e
			dd := *e.Decision
			dd.How = "invalid"
			m.Decision = &dd
			_, err := eventlog.Replay(withEvent(evs, i, m), evs[len(evs)-1].Seq, resolve)
			if err == nil || errSeq(err) != e.Seq {
				t.Fatalf("invalid decision at seq %d: want an error naming that seq, got %v", e.Seq, err)
			}
			return
		}
	}
	t.Skip("no non-auto decision in the sample games")
}

// activationDeck is a synthetic deck whose units print three activated
// abilities, one per cost kind (T-0159).
func activationDeck() engine.DeckList {
	hq := &engine.Card{Name: "Act HQ", Type: engine.TypeHeadquarters, Power: 3, Health: 20, ReqYield: 3, TacYield: 2,
		DeployZones: []engine.Zone{engine.ZoneHeadquarters}}
	u := &engine.Card{Name: "Act Unit", Type: engine.TypeUnit, Power: 1, Health: 3, ReqCost: 1, ReqYield: 1, TacYield: 1,
		DeployZones: []engine.Zone{engine.ZoneReserves, engine.ZoneMainTheater},
		Ability: &ability.Def{Activated: []ability.Activated{
			{Costs: []ability.Cost{{Kind: ability.CostExhaustSelf}},
				Effect: ability.Effect{Steps: []ability.Step{{Op: ability.OpDraw, Amount: ability.Value{N: 1}}}}},
			{Costs: []ability.Cost{{Kind: ability.CostKillSelf}},
				Effect: ability.Effect{
					Targets: []ability.Target{{Filter: ability.Filter{Allegiance: ability.AllegianceOpposing}}},
					Steps:   []ability.Step{{Op: ability.OpDealDamage, Amount: ability.Value{N: 1}, Subject: ability.SubjectTarget, Target: 0}}}},
			{Costs: []ability.Cost{{Kind: ability.CostResources, Req: 1}},
				Effect: ability.Effect{Steps: []ability.Step{{Op: ability.OpAddArmor, Amount: ability.Value{N: 1}, Subject: ability.SubjectSelf}}}},
		}}}
	main := make([]*engine.Card, 30)
	for i := range main {
		main[i] = u
	}
	return engine.DeckList{HQ: hq, Main: main}
}

var activationNames = [2]string{"act", "act"}

func activationResolve(name string) (engine.DeckList, error) {
	if name != "act" {
		return engine.DeckList{}, fmt.Errorf("unknown deck %q", name)
	}
	return activationDeck(), nil
}

func activationGames(t *testing.T) [][]engine.Event {
	t.Helper()
	d := activationDeck()
	games, err := play([2]engine.DeckList{d, d}, activationNames, 30, 7)
	if err != nil {
		t.Fatalf("activation games: %v", err)
	}
	return games
}

func TestReplay_ActivationsRoundTrip(t *testing.T) {
	games := activationGames(t)
	seen := map[int]int{}
	for gi, evs := range games {
		for _, e := range evs {
			if e.Kind == "action" && e.Detail == "Activate" {
				seen[e.Amount]++
			}
		}
		if _, err := eventlog.Replay(evs, len(evs), activationResolve); err != nil {
			t.Fatalf("game %d: %v", gi+1, err)
		}
	}
	for i := 0; i < 3; i++ {
		if seen[i] == 0 {
			t.Errorf("no Activate action event with Amount %d over 30 games: %v", i, seen)
		}
	}
}

func TestReplay_ActivationWithAnotherIndexDiverges(t *testing.T) {
	for _, evs := range activationGames(t) {
		for i, e := range evs {
			if e.Kind != "action" || e.Detail != "Activate" || e.Amount != 1 {
				continue
			}
			cp := append([]engine.Event{}, evs...)
			cp[i].Amount = 0
			_, err := eventlog.Replay(cp, len(cp), activationResolve)
			if err == nil || !strings.Contains(err.Error(), "replay diverged") {
				t.Fatalf("Activate index 1 recorded as 0: error %v, want replay diverged", err)
			}
			return
		}
	}
	t.Fatal("no game holds an Activate event with Amount 1")
}

// searchDeck is a synthetic deck whose units search the deck on exhaust (T-0168).
func searchDeck() engine.DeckList {
	hq := &engine.Card{Name: "Search HQ", Type: engine.TypeHeadquarters, Power: 3, Health: 20, ReqYield: 3, TacYield: 2,
		DeployZones: []engine.Zone{engine.ZoneHeadquarters}}
	u := &engine.Card{Name: "Search Unit", Type: engine.TypeUnit, Power: 1, Health: 3, ReqCost: 1, ReqYield: 1, TacYield: 1,
		DeployZones: []engine.Zone{engine.ZoneReserves, engine.ZoneMainTheater},
		Ability: &ability.Def{Activated: []ability.Activated{
			{Costs: []ability.Cost{{Kind: ability.CostExhaustSelf}},
				Effect: ability.Effect{Steps: []ability.Step{{Op: ability.OpSearchDeck}}}},
		}}}
	main := make([]*engine.Card, 30)
	for i := range main {
		main[i] = u
	}
	return engine.DeckList{HQ: hq, Main: main}
}

var searchNames = [2]string{"search", "search"}

func searchResolve(name string) (engine.DeckList, error) {
	if name != "search" {
		return engine.DeckList{}, fmt.Errorf("unknown deck %q", name)
	}
	return searchDeck(), nil
}

func searchGames(t *testing.T) [][]engine.Event {
	t.Helper()
	d := searchDeck()
	games, err := play([2]engine.DeckList{d, d}, searchNames, 30, 7)
	if err != nil {
		t.Fatalf("search games: %v", err)
	}
	return games
}

func TestReplay_SearchesRoundTrip(t *testing.T) {
	games := searchGames(t)
	searchCount := 0
	shuffledSearches := 0
	for gi, evs := range games {
		var searches [2][]*engine.DecisionRecord
		for _, e := range evs {
			if e.Decision != nil && e.Decision.Kind == engine.DecideSearch.String() {
				searchCount++
				searches[e.Decision.Player] = append(searches[e.Decision.Player], e.Decision)
			}
		}
		for seat := 0; seat < 2; seat++ {
			sList := searches[seat]
			if len(sList) >= 2 {
				d1 := sList[0]
				d2 := sList[1]
				pickedSet := map[int]bool{}
				for _, p := range d1.Picks {
					if p >= 0 && p < len(d1.Refs) {
						pickedSet[d1.Refs[p]] = true
					}
				}
				d2Set := map[int]bool{}
				for _, ref := range d2.Refs {
					d2Set[ref] = true
				}
				var filtered1 []int
				for _, ref := range d1.Refs {
					if !pickedSet[ref] && d2Set[ref] {
						filtered1 = append(filtered1, ref)
					}
				}
				if !slices.Equal(filtered1, d2.Refs) {
					shuffledSearches++
				}
			}
		}
		if _, err := eventlog.Replay(evs, len(evs), searchResolve); err != nil {
			t.Fatalf("game %d: %v", gi+1, err)
		}
	}
	if searchCount == 0 {
		t.Fatal("no decision has Kind DecideSearch over 30 games")
	}
	if shuffledSearches == 0 {
		t.Fatal("no seat searched twice with a shuffled order between over 30 games")
	}
}

func TestReplay_SearchWithSwappedOptionsDiverges(t *testing.T) {
	for _, evs := range searchGames(t) {
		for i, e := range evs {
			if e.Decision == nil || e.Decision.Kind != engine.DecideSearch.String() || len(e.Decision.Refs) < 2 {
				continue
			}
			cp := append([]engine.Event{}, evs...)
			decCopy := *e.Decision
			decCopy.Refs = append([]int{}, e.Decision.Refs...)
			decCopy.Refs[0], decCopy.Refs[1] = decCopy.Refs[1], decCopy.Refs[0]
			cp[i].Decision = &decCopy
			_, err := eventlog.Replay(cp, len(cp), searchResolve)
			if err == nil || !strings.Contains(err.Error(), "replay diverged") {
				t.Fatalf("search with swapped refs: error %v, want replay diverged", err)
			}
			return
		}
	}
	t.Fatal("no game holds a search decision with at least two refs")
}

// T-0183 (directive #119): in real games every search record is flagged hidden
// and no other record is, in the record and in its JSON, where an open decision
// has no hidden key.
func TestReplay_OnlySearchRecordsAreHidden(t *testing.T) {
	searches, open := 0, 0
	for gi, evs := range slices.Concat(searchGames(t), allGames(t)) {
		flagged := 0
		for _, e := range evs {
			if e.Decision == nil {
				continue
			}
			isSearch := e.Decision.Kind == engine.DecideSearch.String()
			if e.Decision.Hidden != isSearch {
				t.Fatalf("game %d: decision %+v, want Hidden only on a search", gi+1, *e.Decision)
			}
			if isSearch {
				flagged++
			} else {
				open++
			}
		}
		searches += flagged
		var buf bytes.Buffer
		if err := eventlog.WriteEventsJSONL(&buf, gi+1, evs); err != nil {
			t.Fatal(err)
		}
		js := buf.String()
		if got := strings.Count(js, `"hidden":`); got != flagged {
			t.Errorf("game %d: JSON has %d hidden keys, want %d, one per search", gi+1, got, flagged)
		}
		if got := strings.Count(js, `"hidden":true`); got != flagged {
			t.Errorf("game %d: JSON has %d hidden:true, want %d", gi+1, got, flagged)
		}
	}
	if searches == 0 || open == 0 {
		t.Fatalf("%d searches and %d other decisions, want some of each", searches, open)
	}
}

// T-0183 (directive #119): the flag is part of the record a replay compares.
func TestReplay_SearchWithoutItsHiddenFlagDiverges(t *testing.T) {
	for _, evs := range searchGames(t) {
		for i, e := range evs {
			if e.Decision == nil || e.Decision.Kind != engine.DecideSearch.String() {
				continue
			}
			cp := append([]engine.Event{}, evs...)
			decCopy := *e.Decision
			decCopy.Hidden = false
			cp[i].Decision = &decCopy
			_, err := eventlog.Replay(cp, len(cp), searchResolve)
			if err == nil || !strings.Contains(err.Error(), "replay diverged") {
				t.Fatalf("search with no flag: error %v, want replay diverged", err)
			}
			return
		}
	}
	t.Fatal("no game holds a search decision")
}

// T-0191 (directive #129): a recorded game with an effect's events replays with
// no divergence after the JSON round trip, and the sample holds every kind.
func TestReplay_EffectEventsRoundTrip(t *testing.T) {
	seen := map[string]int{}
	for i, evs := range allGames(t) {
		found := false
		for _, e := range evs {
			switch e.Kind {
			case engine.EventHeal, engine.EventArmor, engine.EventPower:
				if e.Target == 0 {
					t.Errorf("game %d, seq %d: a %s event with no target", i+1, e.Seq, e.Kind)
				}
			case engine.EventCardMove:
				if e.Card == "" || e.Inst == 0 || e.Zone == "" {
					t.Errorf("game %d, seq %d: a card-move event with no card, instance or zone", i+1, e.Seq)
				}
			default:
				continue
			}
			seen[e.Kind]++
			found = true
		}
		if !found {
			continue
		}
		if _, err := eventlog.Replay(evs, len(evs), resolveDeck(t)); err != nil {
			t.Fatalf("game %d: %v", i+1, err)
		}
	}
	for _, kind := range []string{engine.EventHeal, engine.EventArmor, engine.EventPower, engine.EventCardMove} {
		if seen[kind] == 0 {
			t.Errorf("no %s event in the sample games", kind)
		} else {
			t.Logf("%s: %d events", kind, seen[kind])
		}
	}
}

// Tests for T-0215 (directive #128, FR-TRN-003): a deploy from a discard pile
// replays from its action event's From.

// revenantDeck is a synthetic deck whose units may be deployed from their
// owner's discard pile by killing an allied unit.
func revenantDeck() engine.DeckList {
	hq := &engine.Card{Name: "Revenant HQ", Type: engine.TypeHeadquarters, Power: 3, Health: 20, ReqYield: 3, TacYield: 2,
		DeployZones: []engine.Zone{engine.ZoneHeadquarters}}
	u := &engine.Card{Name: "Revenant", Type: engine.TypeUnit, Power: 2, Health: 1, ReqCost: 1, ReqYield: 1, TacYield: 1,
		DeployZones: []engine.Zone{engine.ZoneReserves, engine.ZoneMainTheater},
		Ability: &ability.Def{Statics: []ability.Static{{
			Kind: ability.StaticDeployPermission, ActiveZone: ability.ZoneDiscard,
			Costs: []ability.Cost{{Kind: ability.CostKill,
				Filter: ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit}}},
		}}}}
	main := make([]*engine.Card, 30)
	for i := range main {
		main[i] = u
	}
	return engine.DeckList{HQ: hq, Main: main}
}

var revenantNames = [2]string{"revenant", "revenant"}

func revenantResolve(name string) (engine.DeckList, error) {
	if name != "revenant" {
		return engine.DeckList{}, fmt.Errorf("unknown deck %q", name)
	}
	return revenantDeck(), nil
}

func revenantGames(t *testing.T) [][]engine.Event {
	t.Helper()
	d := revenantDeck()
	games, err := play([2]engine.DeckList{d, d}, revenantNames, 30, 7)
	if err != nil {
		t.Fatalf("revenant games: %v", err)
	}
	return games
}

// isDeploy reports whether e is a Deploy action event from the pile from names
// ("" for the hand).
func isDeploy(e engine.Event, from string) bool {
	return e.Kind == "action" && e.Detail == "Deploy" && e.From == from
}

func TestReplay_DiscardDeploysRoundTrip(t *testing.T) {
	seen := 0
	for gi, evs := range revenantGames(t) {
		for _, e := range evs {
			if isDeploy(e, "Discard") {
				seen++
			}
		}
		if _, err := eventlog.Replay(evs, len(evs), revenantResolve); err != nil {
			t.Fatalf("game %d: %v", gi+1, err)
		}
	}
	if seen == 0 {
		t.Errorf("no Deploy action event with From %q over 30 games", "Discard")
	}
}

func TestReplay_DiscardDeployFromAnUnknownZoneIsAnError(t *testing.T) {
	for _, evs := range revenantGames(t) {
		for i, e := range evs {
			if !isDeploy(e, "Discard") {
				continue
			}
			cp := append([]engine.Event{}, evs...)
			cp[i].From = "Nowhere"
			_, err := eventlog.Replay(cp, len(cp), revenantResolve)
			want := fmt.Sprintf("replay diverged at seq %d: unknown zone %q", e.Seq, "Nowhere")
			if err == nil || !strings.Contains(err.Error(), want) {
				t.Fatalf("From recorded as Nowhere: error %v, want it to contain %q", err, want)
			}
			return
		}
	}
	t.Fatal("no game holds a Deploy event with From Discard")
}

func TestReplay_DeployWithAnotherFromDiverges(t *testing.T) {
	for _, evs := range revenantGames(t) {
		for i, e := range evs {
			if !isDeploy(e, "") {
				continue
			}
			cp := append([]engine.Event{}, evs...)
			cp[i].From = "Hand"
			_, err := eventlog.Replay(cp, len(cp), revenantResolve)
			want := fmt.Sprintf("replay diverged at seq %d: action event fields", e.Seq)
			if err == nil || !strings.Contains(err.Error(), want) {
				t.Fatalf("a hand Deploy recorded with From Hand: error %v, want it to contain %q", err, want)
			}
			return
		}
	}
	t.Fatal("no game holds a Deploy event from hand")
}
