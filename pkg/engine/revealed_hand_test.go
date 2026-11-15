package engine

import (
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0243 (FR-MOD-001 §1, §3, §6; SVY-R16, SVY-R50; board directive
// #156): a discard chosen by the ability's controller from an opponent's
// revealed hand (Drifting Concept's shape). Synthetic cards, plus Eerie
// Remnant for the trigger check.

// rhControllerDiscard is a step in which the controller chooses one card from
// each opponent's hand to be discarded.
func rhControllerDiscard() ability.Step {
	return ability.Step{
		Op:                ability.OpDiscard,
		Amount:            ability.Value{N: 1},
		Subject:           ability.SubjectOpponents,
		ControllerChooses: true,
	}
}

// rhScene gives player 0 a hand of two own cards (L0, L1) and player 1 a hand
// of the named cards (the last is the most recently added). It returns the
// scene and player 1's instances.
func rhScene(t *testing.T, names ...string) (*scene, []*Instance) {
	t.Helper()
	s := newScene(t)
	s.g.Players[0].Hand = nil
	s.g.Players[1].Hand = nil
	s.give(0, vanilla("L0")[0])
	s.give(0, vanilla("L1")[0])
	var insts []*Instance
	for _, c := range vanilla(names...) {
		insts = append(insts, s.give(1, c))
	}
	return s, insts
}

func (s *scene) rhRun(steps ...ability.Step) {
	s.t.Helper()
	s.deployFromHand(0, trigUnit("Source", deployTrigger(steps...)))
}

// rhDiscardedFromOpponent checks that exactly want left player 1's hand for
// player 1's discard pile, as its top card.
func rhDiscardedFromOpponent(t *testing.T, s *scene, all []*Instance, want *Instance, why string) {
	t.Helper()
	p1 := s.g.Players[1]
	if n := len(p1.Discard); n == 0 || p1.Discard[n-1] != want {
		t.Errorf("%s: the card goes onto player 1's discard pile, got %v", why, discardNames(p1))
	}
	if inHand(p1, want) {
		t.Errorf("%s: the card has left player 1's hand", why)
	}
	if len(p1.Hand) != len(all)-1 {
		t.Errorf("%s: player 1's hand has %d cards, want %d", why, len(p1.Hand), len(all)-1)
	}
	if len(s.g.Players[0].Discard) != 0 {
		t.Errorf("%s: player 0's discard pile is unchanged, got %v", why, discardNames(s.g.Players[0]))
	}
}

func TestFR_MOD_001_ControllerChoosesFromOpponentsHand(t *testing.T) {
	s, insts := rhScene(t, "A", "B", "C")
	s.script(0, Answer{Picks: []int{1}})
	s.script(1)
	hand0 := len(s.g.Players[0].Hand)
	s.rhRun(rhControllerDiscard())
	if len(s.asked(1)) != 0 {
		t.Errorf("the hand's owner is asked nothing: the controller chooses, got %d decisions (FR-MOD-001 §3)", len(s.asked(1)))
	}
	asks := s.asked(0)
	if len(asks) != 1 {
		t.Fatalf("the controller is asked exactly one decision, got %d (FR-MOD-001 §3)", len(asks))
	}
	d := asks[0]
	if d.Kind != DecideHand {
		t.Errorf("kind = %v, want Hand (SVY-R16)", d.Kind)
	}
	if !d.Hidden {
		t.Errorf("choosing from an opponent's hand is an inspection: Hidden must be true (FR-MOD-001 §1, §3)")
	}
	if len(d.Options) != 3 {
		t.Fatalf("options are exactly player 1's 3 cards, got %d", len(d.Options))
	}
	for i, o := range d.Options {
		if o.Ref != insts[i].ID {
			t.Errorf("option %d refers to instance %d, want player 1's card %d", i, o.Ref, insts[i].ID)
		}
	}
	rhDiscardedFromOpponent(t, s, insts, insts[1], "the picked card B")
	if got := len(s.g.Players[0].Hand); got != hand0 {
		t.Errorf("player 0's hand is unchanged, %d cards, want %d", got, hand0)
	}
	evs := eventsOfKind(s.g.Log, "discard")
	if len(evs) != 1 || evs[0].Player != 1 {
		t.Errorf("one discard event naming player 1 (the card's owner), got %+v", evs)
	}
}

func TestFR_MOD_001_RevealedHandTimeoutTakesMostRecent(t *testing.T) {
	// Covers: FR-MOD-001
	s, insts := rhScene(t, "A", "B", "C")
	s.script(0, Answer{Timeout: true})
	s.script(1)
	s.rhRun(rhControllerDiscard())
	rhDiscardedFromOpponent(t, s, insts, insts[2], "a timeout takes the card most recently added to the opponent's hand, C (FR-MOD-001 §6, SVY-R50)")
}

func TestFR_MOD_001_RevealedHandInvalidAnswerTakesMostRecent(t *testing.T) {
	// Covers: FR-MOD-001
	s, insts := rhScene(t, "A", "B", "C")
	s.script(0, Answer{Picks: []int{7}})
	s.script(1)
	s.rhRun(rhControllerDiscard())
	rhDiscardedFromOpponent(t, s, insts, insts[2], "an out-of-range pick falls back to the card most recently added to the opponent's hand, C (FR-MOD-001 §6, SVY-R50)")
}

func TestFR_MOD_001_RevealedOneCardHandIsAsked(t *testing.T) {
	s, insts := rhScene(t, "A")
	s.script(0, Answer{Picks: []int{0}})
	s.script(1)
	s.rhRun(rhControllerDiscard())
	asks := s.asked(0)
	if len(asks) != 1 {
		t.Fatalf("an inspection of a one-card opponent hand is still asked, got %d decisions (FR-MOD-001 §1, §3)", len(asks))
	}
	if !asks[0].Hidden || len(asks[0].Options) != 1 {
		t.Errorf("decision is Hidden with one option, got Hidden=%v options=%d (FR-MOD-001 §1)", asks[0].Hidden, len(asks[0].Options))
	}
	rhDiscardedFromOpponent(t, s, insts, insts[0], "the only card")
}

func TestFR_MOD_001_RevealedEmptyHandAsksNothing(t *testing.T) {
	s, _ := rhScene(t)
	s.script(0)
	s.script(1)
	drawStep := ability.Step{Op: ability.OpDraw, Subject: ability.SubjectController, Amount: ability.Value{N: 1}, IfDone: true}
	hand0 := len(s.g.Players[0].Hand)
	s.rhRun(rhControllerDiscard(), drawStep)
	if len(s.asked(0))+len(s.asked(1)) != 0 {
		t.Errorf("an empty hand asks no one, asked %d and %d (FR-MOD-001)", len(s.asked(0)), len(s.asked(1)))
	}
	if len(s.g.Players[1].Discard) != 0 || len(eventsOfKind(s.g.Log, "discard")) != 0 {
		t.Errorf("nothing is discarded from an empty hand")
	}
	if got := len(s.g.Players[0].Hand); got != hand0 {
		t.Errorf("the IfDone draw does not run when nothing was discarded: player 0 hand %d, want %d (FR-TRN-003a)", got, hand0)
	}
}

func TestFR_MOD_001_RevealedHandDiscardFiresOwnersTrigger(t *testing.T) {
	// Covers: FR-ZON-004, FR-TRG-002
	s := newScene(t)
	s.g.Players[0].Hand = nil
	s.g.Players[1].Hand = nil
	s.give(0, vanilla("L0")[0])
	other := s.give(1, vanilla("Other")[0])
	rem := s.give(1, s.card("Eerie Remnant"))
	s.script(0, Answer{Picks: []int{1}})
	s.byKind(1, func(d Decision) Answer {
		if d.Kind == DecideOptional {
			return Answer{Picks: []int{0}}
		}
		return Answer{Timeout: true}
	})
	s.rhRun(rhControllerDiscard())
	if got := askedKind(s, 1, DecideOptional); len(got) != 1 {
		t.Fatalf("Remnant's trigger is its owner's: player 1 is asked its \"you may\" once, got %d (FR-TRG-002)", len(got))
	}
	if got := askedKind(s, 0, DecideOptional); len(got) != 0 {
		t.Errorf("the controller is not asked Remnant's optional, got %d", len(got))
	}
	if !remnantExhaustedInReserves(s, 1, rem) {
		t.Errorf("accepting deploys Remnant for player 1, Exhausted, in player 1's Reserves (FR-ZON-004)")
	}
	if !inHand(s.g.Players[1], other) {
		t.Errorf("the other card stays in player 1's hand")
	}
}
