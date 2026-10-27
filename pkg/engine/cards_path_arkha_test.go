package engine

import (
	"slices"
	"testing"
)

// Card tests for Arkha, Sun's Fury (T-0078), written from the CSV rules text in
// data/cards/units-and-structures-path.csv: "Whenever this unit attacks, you may
// put a unit card from your discard pile on the bottom of your deck. If you do,
// deal 3 damage to chosen opposing Headquarters."
// Rulings: SVY-R9 ("you may" asked on placement), SVY-R3 (the Headquarters is
// chosen on placement; with one legal target it is chosen without a question,
// FR-MOD-001), SVY-R14 (the put is an untargeted pick on resolution; a timeout
// or no unit card means no damage), SVY-R16 (a timeout selects nothing).
// Player 0 controls Arkha and attacks an opposing 0/20 dummy, so any damage to
// player 1's Headquarters comes from the trigger.

type arkhaScene struct {
	s     *scene
	arkha *Asset
	dummy *Asset
}

func newArkhaScene(t *testing.T) arkhaScene {
	t.Helper()
	s := newScene(t)
	return arkhaScene{
		s:     s,
		arkha: s.put(0, ZoneMainTheater, s.card("Arkha, Sun's Fury")),
		dummy: s.put(1, ZoneMainTheater, unit("Dummy", 0, 20, Keywords{})),
	}
}

// arkhaChooser installs a chooser for seat 0 that answers "you may" with opt,
// each DecidePick with the option whose ref is want (a timeout if want is nil)
// and any target question with its first option. It records every Decision.
func (s *scene) arkhaChooser(opt Answer, want *Instance) {
	if s.asks == nil {
		s.asks = map[PlayerID][]Decision{}
	}
	s.asks[0] = nil
	s.g.choosers[0] = func(_ *Game, d Decision) Answer {
		s.asks[0] = append(s.asks[0], d)
		switch d.Kind {
		case DecideOptional:
			return opt
		case DecidePick:
			if want == nil {
				return Answer{Timeout: true}
			}
			for i, o := range d.Options {
				if o.Ref == want.ID {
					return Answer{Picks: []int{i}}
				}
			}
			s.t.Errorf("wanted card %d is not among the pick options", want.ID)
			return Answer{Timeout: true}
		}
		return Answer{Picks: []int{0}}
	}
}

func TestCard_ArkhasSunsFury_PutsAUnitCardOnTheBottomAndDamagesTheOpposingHeadquarters(t *testing.T) {
	// Covers: FR-TRG-005, FR-TGT-001, FR-MOD-001
	a := newArkhaScene(t)
	s := a.s
	cards := s.inDiscard(0, unit("A", 1, 1, Keywords{}), unit("B", 1, 1, Keywords{}))
	s.arkhaChooser(yes, cards[0])
	s.do(s.attack(a.arkha, a.dummy))
	if got := kindsOf(s.asked(0)); !slices.Equal(got, []DecisionKind{DecideOptional, DecidePick}) {
		t.Fatalf("'you may' is asked on placement, the single opposing Headquarters is chosen without a question, and the pick is asked on resolution: got %v (SVY-R9, SVY-R3, SVY-R14)", got)
	}
	p := s.g.Players[0]
	if n := len(p.Deck); n == 0 || p.Deck[n-1] != cards[0] {
		t.Errorf("the picked unit card is the bottom card of its controller's deck (SVY-R14)")
	}
	if inDiscardPile(p, cards[0]) {
		t.Errorf("the picked card leaves the discard pile")
	}
	if !inDiscardPile(p, cards[1]) {
		t.Errorf("the unit card that was not picked stays in the discard pile")
	}
	s.wantDamage(s.g.Players[1].HQ, 3)
	s.wantDamage(s.g.Players[0].HQ, 0)
}

func TestCard_ArkhasSunsFury_OnlyUnitCardsInYourDiscardPileAreOffered(t *testing.T) {
	// Covers: FR-TGT-001, FR-MOD-001
	a := newArkhaScene(t)
	s := a.s
	mine := s.inDiscard(0, unit("A", 1, 1, Keywords{}), structure("Wall", 0, 3, Keywords{}), unit("B", 1, 1, Keywords{}))
	s.inDiscard(1, unit("Theirs", 1, 1, Keywords{}))
	s.arkhaChooser(yes, mine[2])
	s.do(s.attack(a.arkha, a.dummy))
	var picks []Decision
	for _, d := range s.asked(0) {
		if d.Kind == DecidePick {
			picks = append(picks, d)
		}
	}
	if len(picks) != 1 {
		t.Fatalf("one pick question, got %d", len(picks))
	}
	got := make([]int, len(picks[0].Options))
	for i, o := range picks[0].Options {
		got[i] = o.Ref
	}
	slices.Sort(got)
	if want := []int{mine[0].ID, mine[2].ID}; !slices.Equal(got, want) {
		t.Errorf("only the unit cards in your own discard pile are offered, not the structure card or the opponent's unit card: got %v, want %v", got, want)
	}
}

func TestCard_ArkhasSunsFury_NoUnitCardMeansNoDamage(t *testing.T) {
	// Covers: SVY-R14
	a := newArkhaScene(t)
	s := a.s
	s.inDiscard(0, structure("Wall", 0, 3, Keywords{}))
	s.inDiscard(1, unit("Theirs", 1, 1, Keywords{}))
	deck := len(s.g.Players[0].Deck)
	s.arkhaChooser(yes, nil)
	s.do(s.attack(a.arkha, a.dummy))
	for _, d := range s.asked(0) {
		if d.Kind == DecidePick {
			t.Errorf("nothing is picked when no unit card is in the pile, so no pick is asked (SVY-R14)")
		}
	}
	if n := len(s.g.Players[0].Deck); n != deck {
		t.Errorf("no card is put on the deck: %d cards, want %d", n, deck)
	}
	s.wantDamage(s.g.Players[1].HQ, 0)
	if d := damageTo(s.g.Log, s.g.Players[1].HQ); len(d) != 0 {
		t.Errorf("with no unit card X was not carried out, so Y does nothing: %+v (SVY-R14)", d)
	}
}

func TestCard_ArkhasSunsFury_TimeoutAtThePickDealsNoDamage(t *testing.T) {
	// Covers: SVY-R14, SVY-R16
	a := newArkhaScene(t)
	s := a.s
	cards := s.inDiscard(0, unit("A", 1, 1, Keywords{}), unit("B", 1, 1, Keywords{}))
	deck := len(s.g.Players[0].Deck)
	s.arkhaChooser(yes, nil)
	s.do(s.attack(a.arkha, a.dummy))
	if got := kindsOf(s.asked(0)); !slices.Equal(got, []DecisionKind{DecideOptional, DecidePick}) {
		t.Fatalf("'you may' then the pick are asked: got %v", got)
	}
	p := s.g.Players[0]
	if !inDiscardPile(p, cards[0]) || !inDiscardPile(p, cards[1]) {
		t.Errorf("a timeout at an 'up to' pick selects none, so no card leaves the pile (SVY-R16)")
	}
	if n := len(p.Deck); n != deck {
		t.Errorf("no card is put on the deck: %d cards, want %d", n, deck)
	}
	s.wantDamage(s.g.Players[1].HQ, 0)
	if d := damageTo(s.g.Log, s.g.Players[1].HQ); len(d) != 0 {
		t.Errorf("X was not carried out, so Y does nothing: %+v (SVY-R14)", d)
	}
}

func TestCard_ArkhasSunsFury_DecliningDoesNothing(t *testing.T) {
	// Covers: SVY-R9
	a := newArkhaScene(t)
	s := a.s
	cards := s.inDiscard(0, unit("A", 1, 1, Keywords{}))
	deck := len(s.g.Players[0].Deck)
	s.arkhaChooser(no, cards[0])
	s.do(s.attack(a.arkha, a.dummy))
	if got := kindsOf(s.asked(0)); !slices.Equal(got, []DecisionKind{DecideOptional}) {
		t.Fatalf("declining 'you may' asks nothing more: got %v (SVY-R9)", got)
	}
	if !inDiscardPile(s.g.Players[0], cards[0]) || len(s.g.Players[0].Deck) != deck {
		t.Errorf("declining moves no card")
	}
	s.wantDamage(s.g.Players[1].HQ, 0)
	if s.g.Asset(a.arkha.ID) == nil || s.g.Asset(a.dummy.ID) == nil {
		t.Fatalf("both assets are still on the battlefield")
	}
	s.wantDamage(a.dummy, 6)
	if !a.arkha.Exhausted {
		t.Errorf("the attack went on: Arkha is Exhausted")
	}
}
