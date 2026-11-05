package engine

import (
	"fmt"
	"reflect"
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0167 (FR-DEK-001, FR-DEK-002, FR-MOD-001; SVY-R14, SVY-R36,
// SVY-R37, SVY-R38): OpSearchDeck asks its controller to choose a card from
// their deck, shuffles the rest and puts the card into their hand. Cards are
// synthetic: an activated ability whose cost is exhausting its source.

func searchStep() ability.Step { return ability.Step{Op: ability.OpSearchDeck} }

// searcher puts a unit on p's Main Theater whose only ability exhausts it to run steps.
func (s *scene) searcher(p PlayerID, ss ...ability.Step) *Asset {
	return s.put(p, ZoneMainTheater, actUnit("Searcher", nil, actAbility([]ability.Cost{actExhaust}, ss...)))
}

// namedDeck replaces p's deck with n distinct cards, "Card 0" on top.
func (s *scene) namedDeck(p PlayerID, n int) []*Instance {
	s.g.Players[p].Deck = nil
	cards := make([]*Card, n)
	for i := range cards {
		cards[i] = unit(fmt.Sprintf("Card %d", i), 1, 1, Keywords{})
	}
	return s.onTop(p, cards...)
}

func instIDs(xs []*Instance) []int {
	out := []int{}
	for _, x := range xs {
		out = append(out, x.ID)
	}
	return out
}

func searchRecords(g *Game) []*DecisionRecord {
	var out []*DecisionRecord
	for _, e := range eventsOfKind(g.Log, EventDecision) {
		if e.Decision != nil && e.Decision.Kind == DecideSearch.String() {
			out = append(out, e.Decision)
		}
	}
	return out
}

func TestFR_DEK_001_SearchPutsTheChosenCardIntoHand(t *testing.T) {
	s := newScene(t)
	insts := s.namedDeck(0, 3)
	oppDeck, oppHand := instIDs(s.g.Players[1].Deck), instIDs(s.g.Players[1].Hand)
	a := s.searcher(0, searchStep())
	s.script(0, Answer{Picks: []int{1}})
	s.do(activate(a, 0))
	p := s.g.Players[0]
	if !inHand(p, insts[1]) {
		t.Errorf("the chosen (second) card goes into the searcher's hand (FR-DEK-001)")
	}
	if got := instIDs(p.Deck); len(got) != 2 || !reflect.DeepEqual(sortedCopy(got), sortedCopy([]int{insts[0].ID, insts[2].ID})) {
		t.Errorf("deck = %v, want only the two cards not chosen (FR-DEK-001)", got)
	}
	if !reflect.DeepEqual(instIDs(s.g.Players[1].Deck), oppDeck) || !reflect.DeepEqual(instIDs(s.g.Players[1].Hand), oppHand) {
		t.Errorf("the opponent's deck and hand are unchanged by a search (FR-DEK-001)")
	}
	s.wantSettled("a deck search")
}

func TestFR_DEK_001_SearchListsTheWholeDeckTopFirst(t *testing.T) {
	s := newScene(t)
	insts := s.namedDeck(0, 4)
	a := s.searcher(0, searchStep())
	s.script(0, Answer{Picks: []int{0}})
	s.do(activate(a, 0))
	asked := s.asked(0)
	if len(asked) != 1 {
		t.Fatalf("a search asks its controller once, asked %d times (FR-DEK-001)", len(asked))
	}
	d := asked[0]
	if d.Kind != DecideSearch || d.Min != 1 || d.Max != 1 {
		t.Errorf("decision = kind %v min %d max %d, want Search, 1, 1 (SVY-R38)", d.Kind, d.Min, d.Max)
	}
	if len(d.Options) != len(insts) {
		t.Fatalf("options = %d, want every card in the deck (%d) (SVY-R38)", len(d.Options), len(insts))
	}
	for i, o := range d.Options {
		if o.Ref != insts[i].ID || o.Label != insts[i].Card.Name {
			t.Errorf("option %d = %q ref %d, want %q ref %d: the deck in order, top first (SVY-R38)",
				i, o.Label, o.Ref, insts[i].Card.Name, insts[i].ID)
		}
	}
}

func TestFR_DEK_001_UnrestrictedSearchMustChoose(t *testing.T) {
	s := newScene(t)
	insts := s.namedDeck(0, 3)
	a := s.searcher(0, searchStep())
	s.script(0, Answer{})
	s.do(activate(a, 0))
	if !inHand(s.g.Players[0], insts[0]) {
		t.Errorf("an answer with no picks is invalid, so the top card goes to hand (SVY-R37, FR-MOD-001 sec. 6)")
	}
	if len(eventsOfKind(s.g.Log, EventInvalidAnswer)) != 1 {
		t.Errorf("an invalid-answer event is logged (SVY-R37, FR-MOD-001 sec. 6)")
	}
}

func TestFR_MOD_001_SearchAsksEvenWithOneCard(t *testing.T) {
	s := newScene(t)
	insts := s.namedDeck(0, 1)
	a := s.searcher(0, searchStep())
	s.script(0, Answer{Picks: []int{0}})
	s.do(activate(a, 0))
	if n := len(s.asked(0)); n != 1 {
		t.Errorf("a deck inspection has no fast path: asked %d times with one card in the deck, want 1 (FR-MOD-001 sec. 1)", n)
	}
	if !inHand(s.g.Players[0], insts[0]) {
		t.Errorf("the only card goes to hand (FR-DEK-001)")
	}
}

func TestFR_DEK_001_EmptyDeckStillShuffles(t *testing.T) {
	s := newScene(t)
	s.g.Players[0].Deck = nil
	a := s.searcher(0, searchStep())
	s.script(0)
	hand := instIDs(s.g.Players[0].Hand)
	s.do(activate(a, 0))
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("an empty deck asks nothing, asked %d times (SVY-R37)", n)
	}
	if recs := eventsOfKind(s.g.Log, EventDecision); len(recs) != 0 {
		t.Errorf("an empty deck logs no decision event, got %d (SVY-R37)", len(recs))
	}
	if !reflect.DeepEqual(instIDs(s.g.Players[0].Hand), hand) {
		t.Errorf("the hand is unchanged by a search of an empty deck (SVY-R37)")
	}
	sh := eventsOfKind(s.g.Log, EventShuffle)
	if len(sh) != 1 || sh[0].Player != 0 {
		t.Errorf("shuffle events = %v, want exactly one naming player 0: the deck is shuffled even when empty (FR-DEK-001, SVY-R37)", sh)
	}
	if s.g.Players[0].DrewFromEmpty || s.g.Outcome.Over {
		t.Errorf("a search is not a draw: an empty deck loses nothing (FR-WIN-001)")
	}
	s.wantSettled("a search of an empty deck")
}

func TestFR_MOD_001_SearchTimeoutTakesTopCard(t *testing.T) {
	for _, tc := range []struct {
		name    string
		chooser bool
	}{{"timeout", true}, {"no chooser", false}} {
		t.Run(tc.name, func(t *testing.T) {
			s := newScene(t)
			insts := s.namedDeck(0, 3)
			s.put(0, ZoneMainTheater, structure("Bystander", 0, 3, Keywords{}))
			s.inDiscard(0, unit("Discarded", 1, 1, Keywords{}))
			a := s.searcher(0, searchStep())
			if tc.chooser {
				s.script(0, Answer{Timeout: true})
			} else {
				s.g.choosers[0] = nil
			}
			s.do(activate(a, 0))
			if !inHand(s.g.Players[0], insts[0]) {
				t.Errorf("on a timeout the card that was on top goes to hand (SVY-R38)")
			}
			recs := searchRecords(s.g)
			if len(recs) != 1 || recs[0].How != "timeout" {
				t.Errorf("search decision records = %+v, want one with How %q (SVY-R38)", recs, "timeout")
			}
		})
	}
}

func TestFR_DEK_001_SearchShufflesTheRest(t *testing.T) {
	s := newScene(t)
	insts := s.namedDeck(0, 20)
	before := instIDs(insts)
	a := s.searcher(0, searchStep())
	s.script(0, Answer{Picks: []int{5}})
	s.do(activate(a, 0))
	p := s.g.Players[0]
	rest := append(append([]int{}, before[:5]...), before[6:]...)
	got := instIDs(p.Deck)
	if !reflect.DeepEqual(sortedCopy(got), sortedCopy(rest)) {
		t.Fatalf("deck = %v, want the other 19 cards (FR-DEK-002)", got)
	}
	if reflect.DeepEqual(got, rest) {
		t.Errorf("the remaining 19 cards keep their old order, want them shuffled (FR-DEK-001, FR-DEK-002)")
	}
	dec, sh := -1, []int{}
	for i, e := range s.g.Log {
		switch e.Kind {
		case EventDecision:
			dec = i
		case EventShuffle:
			sh = append(sh, i)
			if e.Card != "" || e.Inst != 0 {
				t.Errorf("a shuffle event names no card: card %q inst %d (FR-DEK-001)", e.Card, e.Inst)
			}
		}
	}
	if dec < 0 || len(sh) != 1 || sh[0] < dec {
		t.Errorf("want exactly one shuffle event after the decision event, got decision at %d, shuffles at %v (FR-DEK-001)", dec, sh)
	}
}

func TestFR_DEK_001_SearchCountsTheCardItFound(t *testing.T) {
	follow := actHQDamage(1)
	follow.IfDone = true
	for _, tc := range []struct {
		name   string
		cards  int
		wantHQ int
	}{{"found a card", 2, 1}, {"empty deck", 0, 0}} {
		t.Run(tc.name, func(t *testing.T) {
			s := newScene(t)
			s.namedDeck(0, tc.cards)
			a := s.searcher(0, searchStep(), follow)
			s.script(0, Answer{Picks: []int{0}})
			s.do(activate(a, 0))
			if got := s.g.Players[1].HQ.Damage; got != tc.wantHQ {
				t.Errorf("IfDone step after a search with %d cards in the deck: HQ damage %d, want %d: a search counts as done only if it found a card (SVY-R14)",
					tc.cards, got, tc.wantHQ)
			}
		})
	}
}

func TestFR_ACT_001_SearchAbilityActivatesWithAnEmptyDeck(t *testing.T) {
	s := newScene(t)
	s.g.Players[0].Deck = nil
	a := s.searcher(0, searchStep())
	s.script(0)
	if !offersActivate(s.g, a, 0) {
		t.Fatalf("an ability with no target slot needs only its costs, so it is offered with an empty deck (SVY-R36)")
	}
	s.do(activate(a, 0))
	if !a.Exhausted {
		t.Errorf("the cost is paid (SVY-R36)")
	}
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("nothing is asked of an empty deck, asked %d (SVY-R36)", n)
	}
	if len(eventsOfKind(s.g.Log, EventShuffle)) != 1 {
		t.Errorf("the shuffle event appears (SVY-R36, FR-DEK-001)")
	}
}
