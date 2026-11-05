package engine

import (
	"reflect"
	"testing"
)

// Tests for T-0169: Cliff of Portents, "R1, T1, exhaust: Search your deck for a
// card, put it into your hand, then shuffle your deck." (FR-ACT-001, FR-ACT-002,
// FR-DEK-001, FR-DEK-002, SVY-R36, SVY-R37, SVY-R38). The real card is loaded
// from data/cards. Player 0 is the active player and activates from the Main
// Theater.

// cliffScene is a scene with Cliff of Portents on player 0's Main Theater, an
// n-card named deck for player 0 and R3 T2 in the pool.
func cliffScene(t *testing.T, n int) (*scene, *Asset, []*Instance) {
	t.Helper()
	s := newScene(t)
	a := s.put(0, ZoneMainTheater, s.card("Cliff of Portents"))
	insts := s.namedDeck(0, n)
	s.setPool(0, 3, 2)
	return s, a, insts
}

func TestCard_CliffOfPortents_PutsTheChosenCardIntoHand(t *testing.T) {
	// Covers: FR-ACT-001, FR-DEK-001
	s, a, insts := cliffScene(t, 3)
	oppDeck, oppHand := instIDs(s.g.Players[1].Deck), instIDs(s.g.Players[1].Hand)
	s.script(0, Answer{Picks: []int{1}})
	s.do(activate(a, 0))
	p := s.g.Players[0]
	if !inHand(p, insts[1]) {
		t.Errorf("the chosen (second) card goes into hand (FR-DEK-001)")
	}
	if got := instIDs(p.Deck); !reflect.DeepEqual(sortedCopy(got), sortedCopy([]int{insts[0].ID, insts[2].ID})) {
		t.Errorf("deck = %v, want only the two cards not chosen (FR-DEK-001)", got)
	}
	if p.Req != 2 || p.Tac != 1 {
		t.Errorf("the cost is 1 Requisition and 1 Tactics: pool R%d T%d, want R2 T1 (FR-ACT-001)", p.Req, p.Tac)
	}
	if !s.alive(a) || a.Zone != ZoneMainTheater || !a.Exhausted {
		t.Errorf("Cliff of Portents stays in the Main Theater, Exhausted (FR-ACT-001)")
	}
	if sh := eventsOfKind(s.g.Log, EventShuffle); len(sh) != 1 || sh[0].Player != 0 {
		t.Errorf("shuffle events = %v, want exactly one naming player 0 (FR-DEK-001)", sh)
	}
	if !reflect.DeepEqual(instIDs(s.g.Players[1].Deck), oppDeck) || !reflect.DeepEqual(instIDs(s.g.Players[1].Hand), oppHand) {
		t.Errorf("the opponent's deck and hand are unchanged (FR-DEK-001)")
	}
}

func TestCard_CliffOfPortents_ListsTheWholeDeckTopFirst(t *testing.T) {
	// Covers: SVY-R38
	s, a, insts := cliffScene(t, 4)
	s.script(0, Answer{Picks: []int{0}})
	s.do(activate(a, 0))
	asked := s.asked(0)
	if len(asked) != 1 {
		t.Fatalf("one search decision is asked, got %d (SVY-R38)", len(asked))
	}
	d := asked[0]
	if d.Kind != DecideSearch || d.Min != 1 || d.Max != 1 {
		t.Errorf("decision = kind %v min %d max %d, want Search, 1, 1 (SVY-R38)", d.Kind, d.Min, d.Max)
	}
	if len(d.Options) != len(insts) {
		t.Fatalf("options = %d, want every card in the deck (%d) (SVY-R38)", len(d.Options), len(insts))
	}
	for i, o := range d.Options {
		if o.Ref != insts[i].ID {
			t.Errorf("option %d = ref %d, want ref %d: deck order, top first (SVY-R38)", i, o.Ref, insts[i].ID)
		}
	}
}

func TestCard_CliffOfPortents_MustChooseACard(t *testing.T) {
	// Covers: SVY-R37, FR-MOD-001
	s, a, insts := cliffScene(t, 3)
	s.script(0, Answer{})
	s.do(activate(a, 0))
	if !inHand(s.g.Players[0], insts[0]) {
		t.Errorf("an answer with no picks is invalid, so the top card goes to hand (SVY-R37, FR-MOD-001 sec. 6)")
	}
	if len(eventsOfKind(s.g.Log, EventInvalidAnswer)) != 1 {
		t.Errorf("an invalid-answer event is logged (SVY-R37, FR-MOD-001 sec. 6)")
	}
}

func TestCard_CliffOfPortents_AsksEvenWithOneCardInDeck(t *testing.T) {
	// Covers: FR-MOD-001
	s, a, insts := cliffScene(t, 1)
	s.script(0, Answer{Picks: []int{0}})
	s.do(activate(a, 0))
	if n := len(s.asked(0)); n != 1 {
		t.Errorf("a deck inspection has no fast path: asked %d times, want 1 (FR-MOD-001 sec. 1)", n)
	}
	if !inHand(s.g.Players[0], insts[0]) {
		t.Errorf("the only card goes to hand (FR-DEK-001)")
	}
}

func TestCard_CliffOfPortents_TimeoutTakesTheTopCard(t *testing.T) {
	// Covers: SVY-R38
	s, a, insts := cliffScene(t, 3)
	s.script(0, Answer{Timeout: true})
	s.do(activate(a, 0))
	p := s.g.Players[0]
	if !inHand(p, insts[0]) {
		t.Errorf("on a timeout the card that was on top before the shuffle goes to hand (SVY-R38)")
	}
	if p.Req != 2 || p.Tac != 1 || !a.Exhausted {
		t.Errorf("costs are paid as usual: pool R%d T%d exhausted %v, want R2 T1 true (FR-ACT-001)", p.Req, p.Tac, a.Exhausted)
	}
	if sh := eventsOfKind(s.g.Log, EventShuffle); len(sh) != 1 {
		t.Errorf("exactly one shuffle event is logged, got %d (FR-DEK-001)", len(sh))
	}
}

func TestCard_CliffOfPortents_EmptyDeckStillActivatesAndShuffles(t *testing.T) {
	// Covers: SVY-R36, SVY-R37, FR-DEK-001
	s, a, _ := cliffScene(t, 0)
	s.script(0)
	hand := instIDs(s.g.Players[0].Hand)
	if !offersActivate(s.g, a, 0) {
		t.Fatalf("with an empty deck the ability is offered: it has no target slot (SVY-R36)")
	}
	s.do(activate(a, 0))
	p := s.g.Players[0]
	if p.Req != 2 || p.Tac != 1 || !a.Exhausted {
		t.Errorf("both resources are paid and the Structure exhausted: pool R%d T%d exhausted %v (FR-ACT-001)", p.Req, p.Tac, a.Exhausted)
	}
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("an empty deck asks nothing, asked %d times (SVY-R37)", n)
	}
	if !reflect.DeepEqual(instIDs(p.Hand), hand) {
		t.Errorf("the hand is unchanged (SVY-R37)")
	}
	if sh := eventsOfKind(s.g.Log, EventShuffle); len(sh) != 1 || sh[0].Player != 0 {
		t.Errorf("shuffle events = %v, want exactly one naming player 0 (FR-DEK-001, SVY-R37)", sh)
	}
	if p.DrewFromEmpty || s.g.Outcome.Over {
		t.Errorf("a search is not a draw: an empty deck loses nothing (FR-WIN-001)")
	}
}

func TestCard_CliffOfPortents_NeedsOneRequisitionAndOneTactics(t *testing.T) {
	// Covers: FR-ACT-001
	for _, tc := range []struct {
		name     string
		req, tac int
	}{{"no Requisition", 0, 5}, {"no Tactics", 5, 0}} {
		t.Run(tc.name, func(t *testing.T) {
			s, a, _ := cliffScene(t, 3)
			s.setPool(0, tc.req, tc.tac)
			actNotOffered(t, s, a, 0, tc.name)
			p := s.g.Players[0]
			if p.Req != tc.req || p.Tac != tc.tac {
				t.Errorf("the pool is unchanged: R%d T%d, want R%d T%d (FR-ACT-001)", p.Req, p.Tac, tc.req, tc.tac)
			}
		})
	}
}

func TestCard_CliffOfPortents_NeedsToBeReadied(t *testing.T) {
	// Covers: FR-ACT-002
	t.Run("exhausted", func(t *testing.T) {
		s, a, _ := cliffScene(t, 3)
		s.setPool(0, 9, 9)
		a.Exhausted = true
		if offersActivate(s.g, a, 0) {
			t.Errorf("an Exhausted Cliff of Portents is not offered (FR-ACT-002)")
		}
		s.reject(activate(a, 0))
	})
	t.Run("once per turn", func(t *testing.T) {
		s, a, _ := cliffScene(t, 3)
		s.setPool(0, 9, 9)
		s.script(0, Answer{Picks: []int{0}})
		s.do(activate(a, 0))
		if offersActivate(s.g, a, 0) {
			t.Errorf("after activating it is Exhausted and not offered again this turn, although the pool could pay (FR-ACT-002)")
		}
		s.reject(activate(a, 0))
	})
}
