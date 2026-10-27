package engine

import (
	"slices"
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0077: an untargeted pick from your discard pile (SVY-R3, SVY-R14,
// SVY-R16), the card steps that move the picked card, and the top-down timeout
// for cards in a pile (FR-MOD-001 sec. 6, board directive #59 (1)). Cards are
// synthetic. Scenes deploy a unit "Source" from player 0's hand (see iydDeploy);
// its deploy trigger runs the steps under test. Player 1 is the opponent.

var (
	ppAllyUnitCards = ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit, Zone: ability.ZoneDiscard}
	ppOppHQ         = []ability.Target{{Filter: ability.Filter{Allegiance: ability.AllegianceOpposing, Kinds: ability.KindHeadquarters}}}
)

// ppPick is "[op] a card picked from a discard pile", picked as the step resolves.
func ppPick(op ability.Op, f ability.Filter, upTo int) ability.Step {
	return ability.Step{Op: op, Subject: ability.SubjectPick, Filter: f, UpTo: upTo, Bottom: op == ability.OpPutOnDeck}
}

// ppScript installs a chooser for seat 0. It says yes to "you may", takes the
// first option of any target question, and answers each pick question from
// picks in order (a timeout once they run out).
func (s *scene) ppScript(picks ...Answer) {
	if s.asks == nil {
		s.asks = map[PlayerID][]Decision{}
	}
	s.asks[0] = nil
	next := 0
	s.g.choosers[0] = func(_ *Game, d Decision) Answer {
		s.asks[0] = append(s.asks[0], d)
		switch d.Kind {
		case DecideOptional:
			return yes
		case DecidePick:
			if next >= len(picks) {
				return Answer{Timeout: true}
			}
			next++
			return picks[next-1]
		}
		return Answer{Picks: []int{0}}
	}
}

// ppRefs returns the option refs in option order (unlike refs, which sorts).
func ppRefs(d Decision) []int {
	out := make([]int, len(d.Options))
	for i, o := range d.Options {
		out[i] = o.Ref
	}
	return out
}

func ppInstIDs(is ...*Instance) []int {
	out := make([]int, len(is))
	for i, in := range is {
		out[i] = in.ID
	}
	return out
}

func ppPicks(s *scene) []Decision {
	var out []Decision
	for _, d := range s.asked(0) {
		if d.Kind == DecidePick {
			out = append(out, d)
		}
	}
	return out
}

func TestFR_TGT_001_PilePickOffersMatchingCardsInYourDiscardPile(t *testing.T) {
	// Covers: SVY-R3
	s := newScene(t)
	mine := s.inDiscard(0, unit("A", 1, 1, Keywords{}), structure("Wall", 0, 3, Keywords{}), unit("B", 1, 1, Keywords{}))
	s.inDiscard(1, unit("Theirs", 1, 1, Keywords{}))
	s.put(0, ZoneMainTheater, unit("On Field", 1, 1, Keywords{}))
	s.ppScript(Answer{Picks: []int{0}})
	s.iydDeploy(iydTrigger(false, nil, ppPick(ability.OpReturnToHand, ppAllyUnitCards, 0)))
	asks := ppPicks(s)
	if len(asks) != 1 {
		t.Fatalf("one pick question, got %d", len(asks))
	}
	want := ppInstIDs(mine[0], mine[2])
	if got := ppRefs(asks[0]); !slices.Equal(got, want) {
		t.Errorf("options are exactly the unit cards in your discard pile, bottom first: got %v, want %v (FR-TGT-001 Zone)", got, want)
	}
	if asks[0].Kind != DecidePick || asks[0].Player != 0 {
		t.Errorf("the controller is asked a DecidePick: %+v", asks[0])
	}
}

func TestFR_TGT_001_PilePickIsMadeOnResolutionFromWhatMatchesThen(t *testing.T) {
	// Covers: SVY-R3
	s := newScene(t)
	old := s.inDiscard(0, unit("Old", 1, 1, Keywords{}))
	ally := s.put(0, ZoneMainTheater, unit("Doomed", 1, 5, Keywords{}))
	kill := ability.Step{Op: ability.OpKill, Subject: ability.SubjectTarget}
	targets := []ability.Target{{Filter: iydAlly}}
	s.ppScript(Answer{Picks: []int{0}})
	s.iydDeploy(iydTrigger(false, targets, kill, ppPick(ability.OpReturnToHand, ppAllyUnitCards, 0)))
	if s.alive(ally) {
		t.Fatalf("the kill step destroyed the chosen ally")
	}
	asks := ppPicks(s)
	if len(asks) != 1 {
		t.Fatalf("two cards match when the pick resolves, so one question: got %d", len(asks))
	}
	want := ppInstIDs(old[0], ally.Inst)
	if got := ppRefs(asks[0]); !slices.Equal(got, want) {
		t.Errorf("a unit card that reached the pile earlier in the same resolution is offered: got %v, want %v (SVY-R3)", got, want)
	}
}

func TestFR_TGT_001_PilePickPutsTheCardOnTheBottomOfItsOwnersDeck(t *testing.T) {
	// Covers: SVY-R14
	s := newScene(t)
	cards := s.inDiscard(0, unit("A", 1, 1, Keywords{}), unit("B", 1, 1, Keywords{}))
	s.ppScript(Answer{Picks: []int{1}})
	s.iydDeploy(iydTrigger(false, nil, ppPick(ability.OpPutOnDeck, ppAllyUnitCards, 0)))
	p := s.g.Players[0]
	if inDiscardPile(p, cards[1]) {
		t.Errorf("the picked card leaves the discard pile")
	}
	if !inDiscardPile(p, cards[0]) {
		t.Errorf("the card that was not picked stays in the discard pile")
	}
	if n := len(p.Deck); n == 0 || p.Deck[n-1] != cards[1] {
		t.Errorf("the picked card is the last card of its owner's deck (bottom)")
	}
}

func TestFR_TGT_001_PilePickCanReturnTheCardToHand(t *testing.T) {
	s := newScene(t)
	cards := s.inDiscard(0, unit("A", 1, 1, Keywords{}), unit("B", 1, 1, Keywords{}))
	s.ppScript(Answer{Picks: []int{0}})
	s.iydDeploy(iydTrigger(false, nil, ppPick(ability.OpReturnToHand, ppAllyUnitCards, 0)))
	p := s.g.Players[0]
	if !inHand(p, cards[0]) {
		t.Errorf("the picked card goes to its owner's hand")
	}
	if inDiscardPile(p, cards[0]) {
		t.Errorf("the picked card leaves the discard pile")
	}
	if !inDiscardPile(p, cards[1]) {
		t.Errorf("the card that was not picked stays in the discard pile")
	}
}

func TestFR_TGT_001_PilePickWithNoMatchDoesNothing(t *testing.T) {
	// Covers: SVY-R3
	s := newScene(t)
	s.inDiscard(0, structure("Wall", 0, 3, Keywords{}))
	s.inDiscard(1, unit("Theirs", 1, 1, Keywords{}))
	s.script(0)
	s.iydDeploy(iydTrigger(false, nil, ppPick(ability.OpReturnToHand, ppAllyUnitCards, 0), scratchN(1)))
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("nothing is asked when nothing matches: %d decisions", n)
	}
	if n := len(eventsOfKind(s.g.Log, "decision")); n != 0 {
		t.Errorf("no decision event is emitted when nothing matches: %d", n)
	}
	if n := faceDownCount(s.g.Players[0]); n != 1 {
		t.Errorf("only the pick clause does nothing; the ability still resolves, so Scratch happens: %d Scratch Companies, want 1", n)
	}
}

func TestFR_MOD_001_SinglePilePickWithOneMatchIsNotAsked(t *testing.T) {
	s := newScene(t)
	cards := s.inDiscard(0, unit("Only", 1, 1, Keywords{}), structure("Wall", 0, 3, Keywords{}))
	s.script(0)
	s.iydDeploy(iydTrigger(false, nil, ppPick(ability.OpReturnToHand, ppAllyUnitCards, 0)))
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("a single pick with one match is made without a question (FR-MOD-001 sec. 1): %d decisions", n)
	}
	if !inHand(s.g.Players[0], cards[0]) {
		t.Errorf("the only match was picked and moved")
	}
}

func TestFR_MOD_001_PilePickTimeoutPicksTheTopCard(t *testing.T) {
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
	s.iydDeploy(iydTrigger(false, nil, ppPick(ability.OpReturnToHand, ppAllyUnitCards, 0)))
	if n := len(s.asked(0)); n != 1 {
		t.Fatalf("three matches are asked once: %d", n)
	}
	if !inHand(p, ib) {
		t.Errorf("a timeout picks cards in a pile from the top down: B, the top card, is picked (FR-MOD-001 sec. 6)")
	}
	if inHand(p, ia) || inHand(p, ic) {
		t.Errorf("only the top card is picked, not the oldest (A) or the first option (A)")
	}
}

func TestFR_TRG_005_IfYouDoRunsYAfterAPilePick(t *testing.T) {
	// Covers: SVY-R14
	s := newScene(t)
	cards := s.inDiscard(0, unit("A", 1, 1, Keywords{}), unit("B", 1, 1, Keywords{}))
	s.ppScript(Answer{Picks: []int{0}})
	s.iydDeploy(iydTrigger(true, ppOppHQ, ppPick(ability.OpPutOnDeck, ppAllyUnitCards, 0), iydIfDone(dmgTo(3, 0))))
	p := s.g.Players[0]
	if n := len(p.Deck); n == 0 || p.Deck[n-1] != cards[0] {
		t.Errorf("X put the picked card on the bottom of the deck")
	}
	s.wantDamage(s.g.Players[1].HQ, 3)
}

func TestFR_TRG_005_IfYouDoSkipsYWhenThePilePickPicksNothing(t *testing.T) {
	// Covers: SVY-R14, SVY-R16
	s := newScene(t)
	cards := s.inDiscard(0, unit("A", 1, 1, Keywords{}))
	s.ppScript(Answer{Timeout: true})
	s.iydDeploy(iydTrigger(true, ppOppHQ, ppPick(ability.OpPutOnDeck, ppAllyUnitCards, 1), iydIfDone(dmgTo(3, 0))))
	if n := len(ppPicks(s)); n != 1 {
		t.Fatalf("an 'up to' pick is asked even with one option: %d", n)
	}
	if !inDiscardPile(s.g.Players[0], cards[0]) {
		t.Errorf("an 'up to' selection selects none on a timeout, so no card moves (FR-MOD-001 sec. 6)")
	}
	s.wantDamage(s.g.Players[1].HQ, 0)
}
