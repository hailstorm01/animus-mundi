package engine

import (
	"reflect"
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0291 (board directive #177): FR-DEK-002 (a shuffle is an immediate
// game state action that randomizes the deck and uses no Resolution Stack item)
// and FR-REP-002 (inline substitution: replacement effects intercept an event
// without the Resolution Stack, and a completely replaced event fires none of
// its watchers). Cards are synthetic. Random order is never asserted.

func placedItems(s *scene) int { return len(eventsOfKind(s.g.Log, "place")) }

func TestFR_DEK_002_ShuffleKeepsEveryCardInTheDeck(t *testing.T) {
	s := newScene(t)
	insts := s.namedDeck(0, 12)
	before := instIDs(insts)
	a := s.searcher(0, searchStep())
	s.script(0, Answer{Picks: []int{3}})
	s.do(activate(a, 0))
	rest := append(append([]int{}, before[:3]...), before[4:]...)
	got := instIDs(s.g.Players[0].Deck)
	if !reflect.DeepEqual(sortedCopy(got), sortedCopy(rest)) {
		t.Errorf("deck after the search = %v, want the same cards as before minus the chosen one, in any order (FR-DEK-002)", got)
	}
	if len(got) != 11 {
		t.Errorf("deck size = %d, want 11: a shuffle neither adds nor removes cards (FR-DEK-002)", len(got))
	}
}

func TestFR_DEK_002_ShuffleUsesNoStackItem(t *testing.T) {
	s := newScene(t)
	s.namedDeck(0, 12)
	a := s.searcher(0, searchStep())
	s.script(0, Answer{Picks: []int{3}})
	stackAtShuffle := -1
	r := &recorder{}
	r.onEvent = func(e Event) {
		if e.Kind == EventShuffle {
			stackAtShuffle = len(s.g.Stack)
		}
	}
	s.g.obs = r
	s.do(activate(a, 0))
	if stackAtShuffle < 0 {
		t.Fatalf("no shuffle event: the search shuffles the rest of the deck (FR-DEK-001, FR-DEK-002)")
	}
	if n := placedItems(s); n != 1 {
		t.Errorf("place events = %d, want exactly 1 (the search ability itself): the shuffle places no stack item (FR-DEK-002, FR-QUE-001)", n)
	}
	if stackAtShuffle > 1 {
		t.Errorf("stack held %d items at the shuffle, want at most the search item: the shuffle is immediate (FR-DEK-002)", stackAtShuffle)
	}
	if len(s.g.Stack) != 0 {
		t.Errorf("stack not empty after the search: %d items (FR-DEK-002)", len(s.g.Stack))
	}
}

func TestFR_DEK_002_ShuffleTouchesOnlyTheSearchedDeck(t *testing.T) {
	s := newScene(t)
	s.namedDeck(0, 10)
	foe := s.namedDeck(1, 10)
	foeBefore := instIDs(foe)
	a := s.searcher(0, searchStep())
	s.script(0, Answer{Picks: []int{0}})
	s.do(activate(a, 0))
	if got := instIDs(s.g.Players[1].Deck); !reflect.DeepEqual(got, foeBefore) {
		t.Errorf("opposing deck order changed: %v -> %v; a shuffle randomizes only the searched deck (FR-DEK-002)", foeBefore, got)
	}
	for _, e := range eventsOfKind(s.g.Log, EventShuffle) {
		if e.Player != 0 {
			t.Errorf("shuffle event names player %d, want only player 0 (FR-DEK-002)", e.Player)
		}
	}
}

func TestFR_REP_002_RedirectionUsesNoStackItem(t *testing.T) {
	run := func(withRedirector bool) (*scene, *Asset, *Asset) {
		s := newScene(t)
		var red *Asset
		if withRedirector {
			red = s.put(0, ZoneMainTheater, redirector("Cloud", 9, ability.Keywords{}))
		}
		u := s.put(0, ZoneMainTheater, unit("Ward", 1, 5, ability.Keywords{}))
		s.dealToAlliedUnits(3)
		return s, u, red
	}
	base, _, _ := run(false)
	s, u, red := run(true)
	s.wantDamage(red, 3)
	s.wantDamage(u, 0)
	if got, want := placedItems(s), placedItems(base); got != want {
		t.Errorf("place events = %d, want %d (the same scene without the redirection): redirecting damage places no stack item (FR-REP-002)", got, want)
	}
	if len(s.g.Stack) != 0 {
		t.Errorf("stack not empty after the redirection: %d items (FR-REP-002)", len(s.g.Stack))
	}
}

func TestFR_REP_002_ResistantUsesNoStackItem(t *testing.T) {
	run := func(resistant int) (*scene, *Asset) {
		s := newScene(t)
		u := s.put(0, ZoneMainTheater, unit("Ward", 1, 9, ability.Keywords{Resistant: resistant}))
		s.dealToAlliedUnits(3)
		return s, u
	}
	base, _ := run(0)
	s, u := run(1)
	s.wantDamage(u, 2)
	if got, want := placedItems(s), placedItems(base); got != want {
		t.Errorf("place events = %d, want %d (the same scene without Resistant): Resistant modifies damage inline (FR-REP-002)", got, want)
	}
}

func TestFR_REP_002_ReplacedDeathFiresNoEntersDiscardTrigger(t *testing.T) {
	s := newScene(t)
	s.put(0, ZoneMainTheater, guardianStructure(3))
	u := s.put(0, ZoneReserves, mortal("Martyr", 1, 1,
		killedTrigger(drawN(1)), entersTrigger(drawN(1))))
	deck := len(s.g.Players[0].Deck)
	s.sweep(0, alliedUnits)
	if s.alive(u) || !inHand(s.g.Players[0], u.Inst) {
		t.Fatalf("precondition: the unit's death is replaced and it is in its owner's hand (FR-REP-002)")
	}
	if inDiscard(s.g.Players[0], u.Inst) {
		t.Errorf("the replaced death never occurs, so the unit is not in the discard pile (FR-REP-002)")
	}
	if got := len(s.g.Players[0].Deck); got != deck {
		t.Errorf("neither the killed nor the enters-discard trigger fires for a completely replaced death (no draw): deck %d -> %d (FR-REP-002)", deck, got)
	}
}
