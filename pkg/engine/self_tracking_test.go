package engine

import (
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for death triggers that move the dying card itself (FR-ZON-004,
// Self-Relocating Trigger Tracking; FR-TRG-006 identity): the trigger tracks
// the exact card instance in its owner's discard pile and does nothing if that
// instance has already left it.

func returnSelf() ability.Step {
	return ability.Step{Op: ability.OpReturnToHand, Subject: ability.SubjectSelf}
}

func putSelfOnDeck(bottom bool) ability.Step {
	return ability.Step{Op: ability.OpPutOnDeck, Subject: ability.SubjectSelf, Bottom: bottom}
}

func inZone(list []*Instance, inst *Instance) bool {
	for _, x := range list {
		if x == inst {
			return true
		}
	}
	return false
}

// killOnP1 has a 5/9 hitter of player 0 kill the given card, played by
// player 1. It returns the dead card's instance.
func killOnP1(s *scene, c *Card) *Instance {
	s.t.Helper()
	att := s.put(0, ZoneMainTheater, unit("Hitter", 5, 9, Keywords{}))
	def := s.put(1, ZoneMainTheater, c)
	inst := def.Inst
	s.do(s.attack(att, def))
	if s.alive(def) {
		s.t.Fatalf("precondition: the unit is destroyed by the combat")
	}
	return inst
}

func TestFR_ZON_004_DeathTriggerReturnsItselfToOwnersHand(t *testing.T) {
	s := newScene(t)
	inst := killOnP1(s, mortal("Tide", 1, 1, killedTrigger(returnSelf())))
	p := s.g.Players[1]
	if !inHand(p, inst) {
		t.Errorf("'return it to its owner's hand' puts the killed card in its owner's hand (FR-ZON-004)")
	}
	if inZone(p.Discard, inst) {
		t.Errorf("the returned card has left the discard pile (FR-ZON-004)")
	}
	if len(s.g.Stack) != 0 {
		t.Errorf("the stack is empty once the trigger resolves: %d items", len(s.g.Stack))
	}
}

func TestFR_ZON_004_DeathTriggerPutsItselfOnBottomOfOwnersDeck(t *testing.T) {
	s := newScene(t)
	inst := killOnP1(s, mortal("Front", 1, 1, killedTrigger(putSelfOnDeck(true))))
	p := s.g.Players[1]
	if n := len(p.Deck); n == 0 || p.Deck[n-1] != inst {
		t.Errorf("'put it on the bottom of its owner's deck' makes the card the last card of the deck (FR-ZON-004)")
	}
	if inZone(p.Discard, inst) || inHand(p, inst) {
		t.Errorf("the card is only in the deck, not in discard or hand (FR-ZON-004)")
	}
}

func TestFR_ZON_004_DeathTriggerPutsItselfOnTopOfOwnersDeck(t *testing.T) {
	s := newScene(t)
	inst := killOnP1(s, mortal("Topper", 1, 1, killedTrigger(putSelfOnDeck(false))))
	p := s.g.Players[1]
	if len(p.Deck) == 0 || p.Deck[0] != inst {
		t.Errorf("putting the card on its owner's deck without 'bottom' makes it the top card (FR-ZON-004)")
	}
	if inZone(p.Discard, inst) {
		t.Errorf("the card has left the discard pile (FR-ZON-004)")
	}
}

func TestFR_ZON_004_OtherStepsOfTheSameEffectStillRun(t *testing.T) {
	s := newScene(t)
	cards := s.onTop(1, vanilla("A", "B")...)
	inst := killOnP1(s, mortal("Tide", 1, 1, killedTrigger(drawN(1), returnSelf())))
	p := s.g.Players[1]
	if !inHand(p, cards[0]) || !inHand(p, inst) || inHand(p, cards[1]) {
		t.Errorf("a Draw 1 and a return-self in one effect both happen (FR-TRG-004, FR-ZON-004)")
	}
}

// twoSelfMovers prints a return-to-hand and a put-on-bottom death trigger,
// each also drawing a card. Options are in printed order: 0 = return, 1 = bottom.
func twoSelfMovers() *Card {
	return mortal("Twin", 1, 1,
		killedTrigger(returnSelf(), drawN(1)),
		killedTrigger(putSelfOnDeck(true), drawN(1)))
}

func TestFR_ZON_004_ReturnResolvingFirstFizzlesThePutOnBottom(t *testing.T) {
	s := newScene(t)
	cards := s.onTop(1, vanilla("A", "B")...)
	s.script(0)
	s.script(1, Answer{Picks: []int{1, 0}}) // the first pick is placed first and resolves last
	inst := killOnP1(s, twoSelfMovers())
	p := s.g.Players[1]
	if !inHand(p, inst) {
		t.Errorf("the return trigger resolves first and moves the card to hand (FR-ZON-004)")
	}
	if inZone(p.Deck, inst) || inZone(p.Discard, inst) {
		t.Errorf("the put-on-bottom trigger fizzles because the card already left the discard pile (FR-ZON-004)")
	}
	if !inHand(p, cards[0]) || !inHand(p, cards[1]) {
		t.Errorf("the fizzled trigger's other step (Draw) still happens (FR-ZON-004)")
	}
}

func TestFR_ZON_004_PutOnBottomResolvingFirstFizzlesTheReturn(t *testing.T) {
	s := newScene(t)
	cards := s.onTop(1, vanilla("A", "B")...)
	s.script(0)
	s.script(1, Answer{Picks: []int{0, 1}})
	inst := killOnP1(s, twoSelfMovers())
	p := s.g.Players[1]
	if n := len(p.Deck); n == 0 || p.Deck[n-1] != inst {
		t.Errorf("the put-on-bottom trigger resolves first and puts the card last in the deck (FR-ZON-004)")
	}
	if inHand(p, inst) || inZone(p.Discard, inst) {
		t.Errorf("the return trigger fizzles because the card already left the discard pile (FR-ZON-004)")
	}
	if !inHand(p, cards[0]) || !inHand(p, cards[1]) {
		t.Errorf("the fizzled trigger's other step (Draw) still happens (FR-ZON-004)")
	}
}

func TestFR_ZON_004_AnotherCopyInDiscardIsNeverMoved(t *testing.T) {
	// FR-TRG-006: tracking is by card instance, not by name. Covers both the
	// moving trigger and the fizzled one.
	for _, picks := range [][]int{{1, 0}, {0, 1}} {
		s := newScene(t)
		s.script(0)
		s.script(1, Answer{Picks: picks})
		card := twoSelfMovers()
		other := &Instance{ID: s.g.newID(), Card: card, Owner: 1}
		s.g.Players[1].Discard = append(s.g.Players[1].Discard, other)
		inst := killOnP1(s, card)
		p := s.g.Players[1]
		if !inZone(p.Discard, other) || inHand(p, other) || inZone(p.Deck, other) {
			t.Errorf("picks %v: another copy of the same card in discard is never moved (FR-TRG-006)", picks)
		}
		if !inHand(p, inst) && !inZone(p.Deck, inst) {
			t.Errorf("picks %v: the dead card itself is moved (FR-ZON-004)", picks)
		}
	}
}

func TestFR_ZON_004_SingleTriggerMovesOnlyTheDeadInstanceNotASameNamedCopy(t *testing.T) {
	// The copy was in the discard pile first; the trigger must not pick it.
	s := newScene(t)
	card := mortal("Tide", 1, 1, killedTrigger(returnSelf()))
	other := &Instance{ID: s.g.newID(), Card: card, Owner: 1}
	s.g.Players[1].Discard = append(s.g.Players[1].Discard, other)
	inst := killOnP1(s, card)
	p := s.g.Players[1]
	if !inHand(p, inst) || inHand(p, other) || !inZone(p.Discard, other) {
		t.Errorf("only the exact dead instance returns to hand; the same-named copy stays in discard (FR-TRG-006)")
	}
}

func TestFR_ZON_004_EachPlayersCardGoesToItsOwnersZone(t *testing.T) {
	s := newScene(t)
	att := s.put(0, ZoneMainTheater, mortal("Tide", 1, 1, killedTrigger(returnSelf())))
	def := s.put(1, ZoneMainTheater, mortal("Front", 1, 1, killedTrigger(putSelfOnDeck(true))))
	ai, di := att.Inst, def.Inst
	s.do(s.attack(att, def))
	if s.alive(att) || s.alive(def) {
		t.Fatalf("precondition: both units die together")
	}
	p0, p1 := s.g.Players[0], s.g.Players[1]
	if !inHand(p0, ai) || inHand(p1, ai) || inZone(p1.Deck, ai) {
		t.Errorf("player 0's card returns to player 0's hand (FR-ZON-004)")
	}
	if n := len(p1.Deck); n == 0 || p1.Deck[n-1] != di || inZone(p0.Deck, di) || inHand(p0, di) {
		t.Errorf("player 1's card goes to the bottom of player 1's own deck (FR-ZON-004)")
	}
	if inZone(p0.Discard, ai) || inZone(p1.Discard, di) {
		t.Errorf("neither card is left in the discard pile (FR-ZON-004)")
	}
}
