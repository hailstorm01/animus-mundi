package engine

import (
	"slices"
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0081: a chosen target in a discard pile, chosen when the trigger
// is placed (SVY-R3) and verified again when it resolves (FR-TGT-003). Cards are
// synthetic. Scenes deploy a unit "Source" from player 0's hand (see iydDeploy);
// its deploy trigger runs the steps under test. Player 1 is the opponent.

// ptSlot is one "chosen unit card in your discard pile" target slot.
func ptSlot(upTo int) ability.Target {
	return ability.Target{Filter: ppAllyUnitCards, UpTo: upTo}
}

// ptStep is "[op] chosen <slot>".
func ptStep(op ability.Op, slot int) ability.Step {
	return ability.Step{Op: op, Subject: ability.SubjectTarget, Target: slot, Bottom: op == ability.OpPutOnDeck}
}

// ptPlaceHook runs fn each time an ability item is placed on the stack, with the
// number of decisions seat 0 had been asked by then and a copy of the item.
func (s *scene) ptPlaceHook(fn func(asked int, it StackItem)) {
	r := &recorder{}
	r.onEvent = func(e Event) {
		if e.Kind != "place" || len(s.g.Stack) == 0 {
			return
		}
		if top := s.g.Stack[len(s.g.Stack)-1]; top.Kind == StackAbility {
			fn(len(s.asked(0)), *top)
		}
	}
	s.g.obs = r
}

func ptLast(is []*Instance) *Instance {
	if len(is) == 0 {
		return nil
	}
	return is[len(is)-1]
}

func TestFR_TGT_001_PileTargetOffersMatchingCardsInYourDiscardPile(t *testing.T) {
	// Covers: SVY-R3
	s := newScene(t)
	mine := s.inDiscard(0, unit("A", 1, 1, Keywords{}), structure("Wall", 0, 3, Keywords{}), unit("B", 1, 1, Keywords{}))
	s.inDiscard(1, unit("Theirs", 1, 1, Keywords{}))
	s.put(0, ZoneMainTheater, unit("On Field", 1, 1, Keywords{}))
	s.script(0, Answer{Picks: []int{0}})
	s.iydDeploy(iydTrigger(false, []ability.Target{ptSlot(0)}, ptStep(ability.OpReturnToHand, 0)))
	asks := s.asked(0)
	if len(asks) != 1 {
		t.Fatalf("one target question, got %d decisions", len(asks))
	}
	d := asks[0]
	if d.Kind != DecideTarget || d.Player != 0 {
		t.Errorf("the controller is asked a DecideTarget: %+v", d)
	}
	if got, want := ppRefs(d), ppInstIDs(mine[0], mine[2]); !slices.Equal(got, want) {
		t.Errorf("options are exactly the unit cards in your discard pile, bottom first: got %v, want %v (FR-TGT-001 zone criterion)", got, want)
	}
	var labels []string
	for _, o := range d.Options {
		labels = append(labels, o.Label)
	}
	if !slices.Equal(labels, []string{"A", "B"}) {
		t.Errorf("options are labelled with their card names: %v", labels)
	}
}

func TestFR_TGT_001_PileTargetIsChosenOnPlacement(t *testing.T) {
	// Covers: SVY-R3
	s := newScene(t)
	cards := s.inDiscard(0, unit("A", 1, 1, Keywords{}), unit("B", 1, 1, Keywords{}))
	s.script(0, Answer{Picks: []int{1}})
	var placed []StackItem
	asksAtPlace := -1
	s.ptPlaceHook(func(asked int, it StackItem) {
		placed = append(placed, it)
		asksAtPlace = asked
	})
	s.iydDeploy(iydTrigger(false, []ability.Target{ptSlot(0)}, ptStep(ability.OpReturnToHand, 0)))
	if len(placed) != 1 {
		t.Fatalf("one ability item is placed, got %d", len(placed))
	}
	if asksAtPlace != 1 {
		t.Errorf("the target is chosen before the ability is placed (SVY-R3): %d questions asked by then, want 1", asksAtPlace)
	}
	if len(placed[0].Chosen) != 1 || !slices.Equal(placed[0].Chosen[0], []int{cards[1].ID}) {
		t.Errorf("the item waits on the stack holding the chosen card's instance ID %d: %v", cards[1].ID, placed[0].Chosen)
	}
}

func TestFR_TGT_001_PileTargetReturnsTheChosenCardToHand(t *testing.T) {
	s := newScene(t)
	cards := s.inDiscard(0, unit("A", 1, 1, Keywords{}), unit("B", 1, 1, Keywords{}))
	s.script(0, Answer{Picks: []int{0}})
	s.iydDeploy(iydTrigger(false, []ability.Target{ptSlot(0)}, ptStep(ability.OpReturnToHand, 0)))
	p := s.g.Players[0]
	if inDiscardPile(p, cards[0]) {
		t.Errorf("the chosen card leaves the discard pile")
	}
	if ptLast(p.Hand) != cards[0] {
		t.Errorf("the chosen card is the last card of its owner's hand")
	}
	if !inDiscardPile(p, cards[1]) {
		t.Errorf("the card that was not chosen stays in the discard pile")
	}
}

func TestFR_TGT_001_PileTargetCanGoOnTheBottomOfTheDeck(t *testing.T) {
	s := newScene(t)
	cards := s.inDiscard(0, unit("A", 1, 1, Keywords{}), unit("B", 1, 1, Keywords{}))
	s.script(0, Answer{Picks: []int{1}})
	s.iydDeploy(iydTrigger(false, []ability.Target{ptSlot(0)}, ptStep(ability.OpPutOnDeck, 0)))
	p := s.g.Players[0]
	if inDiscardPile(p, cards[1]) {
		t.Errorf("the chosen card leaves the discard pile")
	}
	if ptLast(p.Deck) != cards[1] {
		t.Errorf("the chosen card is the last card of its owner's deck (bottom)")
	}
	if !inDiscardPile(p, cards[0]) {
		t.Errorf("the card that was not chosen stays in the discard pile")
	}
}

func TestFR_TGT_003_PileTargetThatLeftThePileIsIllegal(t *testing.T) {
	// Covers: FR-TGT-001, SVY-R3
	s := newScene(t)
	cards := s.inDiscard(0, unit("A", 1, 1, Keywords{}), unit("B", 1, 1, Keywords{}))
	a := cards[0]
	s.script(0, Answer{Picks: []int{0}})
	s.ptPlaceHook(func(int, StackItem) {
		// After placement and before resolution, the chosen card leaves the pile.
		if inDiscardPile(s.g.Players[0], a) {
			s.g.takeCard(a)
			s.g.putCard(a, ZoneDeck, true)
		}
	})
	slots := []ability.Target{ptSlot(0), ppOppHQ[0]}
	s.iydDeploy(iydTrigger(false, slots, dmgTo(2, 1), ptStep(ability.OpReturnToHand, 0), iydIfDone(dmgTo(5, 1))))
	p := s.g.Players[0]
	if inHand(p, a) {
		t.Errorf("a chosen card that left the pile is illegal at resolution, so it is not returned to hand (FR-TGT-003)")
	}
	if ptLast(p.Deck) != a {
		t.Errorf("the card stays where it went (the bottom of the deck)")
	}
	s.wantDamage(s.g.Players[1].HQ, 2)
}

func TestFR_TGT_003_PileTargetUpToNKeepsTheCardsStillInThePile(t *testing.T) {
	// Covers: FR-TGT-001, SVY-R11
	s := newScene(t)
	cards := s.inDiscard(0, unit("A", 1, 1, Keywords{}), unit("B", 1, 1, Keywords{}), unit("C", 1, 1, Keywords{}))
	a, b, c := cards[0], cards[1], cards[2]
	s.script(0, Answer{Picks: []int{0, 1}})
	s.ptPlaceHook(func(int, StackItem) {
		if inDiscardPile(s.g.Players[0], a) {
			s.g.takeCard(a)
			s.g.putCard(a, ZoneDeck, true)
		}
	})
	s.iydDeploy(iydTrigger(false, []ability.Target{ptSlot(2)}, ptStep(ability.OpReturnToHand, 0)))
	p := s.g.Players[0]
	if inHand(p, a) {
		t.Errorf("the card that left the pile is not returned (FR-TGT-003)")
	}
	if !inHand(p, b) || inDiscardPile(p, b) {
		t.Errorf("the other chosen card, still in the pile, is returned to hand (partial invalidation, FR-TGT-003)")
	}
	if !inDiscardPile(p, c) {
		t.Errorf("a card that was not chosen stays in the discard pile")
	}
}

func TestFR_TGT_003_PileTargetWithNoMatchAsksNothing(t *testing.T) {
	// Covers: FR-TGT-001, SVY-R16
	s := newScene(t)
	s.inDiscard(0, structure("Wall", 0, 3, Keywords{}))
	theirs := s.inDiscard(1, unit("Theirs", 1, 1, Keywords{}))
	s.script(0)
	places := 0
	s.ptPlaceHook(func(int, StackItem) { places++ })
	handBefore := len(s.g.Players[0].Hand)
	s.iydDeploy(iydTrigger(false, []ability.Target{ptSlot(0)}, scratchN(1), ptStep(ability.OpReturnToHand, 0)))
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("nothing is asked when no card matches: %d decisions", n)
	}
	if n := len(eventsOfKind(s.g.Log, "decision")); n != 0 {
		t.Errorf("no decision event is emitted when no card matches: %d", n)
	}
	if places != 1 {
		t.Errorf("the trigger is still placed: %d ability items placed", places)
	}
	if n := faceDownCount(s.g.Players[0]); n != 1 {
		t.Errorf("the ability still resolves, so Scratch happens: %d Scratch Companies, want 1", n)
	}
	if len(s.g.Players[0].Hand) != handBefore || len(s.g.Players[0].Discard) != 1 || !inDiscardPile(s.g.Players[1], theirs[0]) {
		t.Errorf("nothing moves when nothing was chosen")
	}
}

func TestFR_MOD_001_SinglePileTargetIsNotAsked(t *testing.T) {
	// Covers: SVY-R16
	s := newScene(t)
	cards := s.inDiscard(0, unit("Only", 1, 1, Keywords{}), structure("Wall", 0, 3, Keywords{}))
	s.script(0)
	s.iydDeploy(iydTrigger(false, []ability.Target{ptSlot(0)}, ptStep(ability.OpReturnToHand, 0)))
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("a single target with one legal option is chosen without a question (FR-MOD-001 sec. 1): %d decisions", n)
	}
	if !inHand(s.g.Players[0], cards[0]) {
		t.Errorf("the only legal card was chosen and returned to hand")
	}
}

func TestFR_MOD_001_PileTargetTimeoutChoosesTheTopCard(t *testing.T) {
	// Covers: SVY-R16
	s := newScene(t)
	// Created A, B, C (A has the lowest ID), then reordered to lie bottom to top as A, C, B.
	insts := s.inDiscard(0, unit("A", 1, 1, Keywords{}), unit("B", 1, 1, Keywords{}), unit("C", 1, 1, Keywords{}))
	ia, ib, ic := insts[0], insts[1], insts[2]
	p := s.g.Players[0]
	s.g.takeCard(ib)
	s.g.takeCard(ic)
	s.g.putCard(ic, ZoneDiscard, false)
	s.g.putCard(ib, ZoneDiscard, false)
	if got := p.Discard; len(got) != 3 || got[0] != ia || got[1] != ic || got[2] != ib {
		t.Fatalf("setup: the discard pile should be A, C, B from the bottom")
	}
	s.script(0, Answer{Timeout: true})
	s.iydDeploy(iydTrigger(false, []ability.Target{ptSlot(0)}, ptStep(ability.OpReturnToHand, 0)))
	if n := len(s.asked(0)); n != 1 {
		t.Fatalf("three matches are asked once: %d", n)
	}
	if !inHand(p, ib) {
		t.Errorf("a timeout chooses cards in a pile from the top down: B, the top card, is chosen (FR-MOD-001 sec. 6)")
	}
	if inHand(p, ia) || inHand(p, ic) {
		t.Errorf("only the top card is chosen, not the oldest (A) or the first option (A)")
	}
}

func TestFR_TGT_001_AssetStepOnAPileSlotDoesNothing(t *testing.T) {
	s := newScene(t)
	cards := s.inDiscard(0, unit("A", 1, 1, Keywords{}), unit("B", 1, 1, Keywords{}))
	s.put(1, ZoneMainTheater, unit("Bystander", 0, 5, Keywords{}))
	s.script(0, Answer{Picks: []int{0}})
	s.iydDeploy(iydTrigger(false, []ability.Target{ptSlot(0)}, dmgTo(3, 0), ptStep(ability.OpReturnToHand, 0)))
	if dmg := nonCombatDamage(s.g.Log); len(dmg) != 0 {
		t.Errorf("a damage step on a pile slot damages nothing, there being no asset chosen: %+v", dmg)
	}
	if !inHand(s.g.Players[0], cards[0]) {
		t.Errorf("the card step beside it still moves the chosen card")
	}
}
