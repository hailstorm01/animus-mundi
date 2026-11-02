package engine

import (
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0132 (FR-MNV-001, FR-TRN-003a; SVY-R3, SVY-R14): a step that
// reads "Then, if you have no cards in hand, ...". The condition is read when
// the step resolves, after the steps before it, on the ability controller's
// hand. The enacted Maneuver is on the stack, never in hand, while its steps
// run. Maneuvers are synthetic; a gated step that must be seen in the event
// log is hqPing(), since a draw logs no event. Player 0 is the active player
// and enacts.

// handEmpty is st with the "if you have no cards in hand" gate set.
func handEmpty(st ability.Step) ability.Step {
	st.IfHandEmpty = true
	return st
}

// handEmptyScene gives player 0 an empty hand, then the Maneuver (steps ss)
// and n other vanilla cards. It returns the scene, the Maneuver and the others.
func handEmptyScene(t *testing.T, n int, ss ...ability.Step) (*scene, *Instance, []*Instance) {
	t.Helper()
	s := newScene(t)
	s.g.Players[0].Hand = nil
	inst := s.give(0, maneuver("Gated", 0, 0, steps(ss...)))
	var others []*Instance
	for i := 0; i < n; i++ {
		others = append(others, s.give(0, unit(string(rune('A'+i)), 1, 1, Keywords{})))
	}
	return s, inst, others
}

func TestFR_MNV_001_HandEmptyStepRunsWithAnEmptyHand(t *testing.T) {
	// Covers: FR-TRN-003a
	s, inst, _ := handEmptyScene(t, 0, handEmpty(drawN(2)))
	p := s.g.Players[0]
	top := []*Instance{p.Deck[0], p.Deck[1]}
	s.do(enactAction(inst))
	if len(p.Hand) != 2 || !inHand(p, top[0]) || !inHand(p, top[1]) {
		t.Errorf("the Maneuver is on the stack, so the hand is empty and the gated draw runs: hand holds %d cards, want the top 2 of the deck (FR-MNV-001)", len(p.Hand))
	}
	if lastDiscard(p) != inst {
		t.Errorf("the Maneuver goes to its owner's discard pile once resolved (FR-MNV-001)")
	}
}

func TestFR_MNV_001_HandEmptyStepSkippedWithACardInHand(t *testing.T) {
	s, inst, o := handEmptyScene(t, 1, handEmpty(drawN(2)))
	p := s.g.Players[0]
	deck := len(p.Deck)
	s.do(enactAction(inst))
	if len(p.Hand) != 1 || !inHand(p, o[0]) {
		t.Errorf("with a card in hand the gated draw does not run: hand holds %d cards, want only the other card (FR-MNV-001)", len(p.Hand))
	}
	if len(p.Deck) != deck {
		t.Errorf("no card was drawn: deck %d, want %d (FR-MNV-001)", len(p.Deck), deck)
	}
	if lastDiscard(p) != inst {
		t.Errorf("the Maneuver goes to its owner's discard pile even though its step was skipped (FR-MNV-001)")
	}
}

func TestFR_MNV_001_HandEmptyReadAfterEarlierSteps(t *testing.T) {
	// Covers: SVY-R3
	t.Run("an earlier discard empties the hand", func(t *testing.T) {
		s, inst, o := handEmptyScene(t, 1, discardN(1), handEmpty(hqPing()))
		s.script(0, Answer{Picks: []int{0}})
		s.do(enactAction(inst))
		if !inDiscard(s.g.Players[0], o[0]) {
			t.Fatalf("the first step discarded the other card")
		}
		if got := pingCount(s); got != 1 {
			t.Errorf("the condition is read after the discard, when the hand is empty: %d pings, want 1 (SVY-R3)", got)
		}
	})
	t.Run("an earlier draw fills the hand", func(t *testing.T) {
		s, inst, _ := handEmptyScene(t, 0, drawN(1), handEmpty(hqPing()))
		p := s.g.Players[0]
		top := p.Deck[0]
		s.do(enactAction(inst))
		if !inHand(p, top) {
			t.Fatalf("the first step drew the top card")
		}
		if got := pingCount(s); got != 0 {
			t.Errorf("the condition is read after the draw, when the hand is not empty: %d pings, want 0 (SVY-R3)", got)
		}
	})
}

func TestFR_MNV_001_HandEmptyReadsTheControllersHand(t *testing.T) {
	t.Run("opponent hand empty, controller holds a card", func(t *testing.T) {
		s, inst, _ := handEmptyScene(t, 1, handEmpty(hqPing()))
		s.g.Players[1].Hand = nil
		s.do(enactAction(inst))
		if got := pingCount(s); got != 0 {
			t.Errorf("\"you\" is the controller, who holds a card: %d pings, want 0 (FR-MNV-001)", got)
		}
	})
	t.Run("controller hand empty, opponent holds cards", func(t *testing.T) {
		s, inst, _ := handEmptyScene(t, 0, handEmpty(hqPing()))
		s.give(1, unit("Theirs", 1, 1, Keywords{}))
		s.do(enactAction(inst))
		if got := pingCount(s); got != 1 {
			t.Errorf("\"you\" is the controller, whose hand is empty: %d pings, want 1 (FR-MNV-001)", got)
		}
	})
}

func TestFR_MNV_001_SkippedHandEmptyStepCountsAsNotDone(t *testing.T) {
	// Covers: SVY-R14
	follow := hqPing()
	follow.Amount = ability.Value{N: 2}
	follow.IfDone = true
	s, inst, _ := handEmptyScene(t, 1, handEmpty(hqPing()), follow)
	s.do(enactAction(inst))
	n := 0
	for _, e := range s.g.Log {
		if e.Kind == "damage" && e.Target == s.g.Players[1].HQ.ID {
			n++
		}
	}
	if n != 0 {
		t.Errorf("a skipped step did nothing, so the \"if you do\" step after it does not run either: %d damage events, want 0 (SVY-R14)", n)
	}
}

func TestFR_TRN_003a_DiscardCostThenHandEmptyStep(t *testing.T) {
	setup := func(t *testing.T) (*scene, *Instance, []*Instance) {
		s := newScene(t)
		s.g.Players[0].Hand = nil
		inst := s.give(0, discardCostManeuver("Steel Rain", 1, 1, 3, steps(handEmpty(hqPing()))))
		o := []*Instance{s.give(0, unit("A", 1, 1, Keywords{})), s.give(0, unit("B", 1, 1, Keywords{}))}
		s.g.Players[0].Req, s.g.Players[0].Tac = 1, 1
		return s, inst, o
	}
	t.Run("discarding the whole hand", func(t *testing.T) {
		s, inst, _ := setup(t)
		s.script(0, Answer{Picks: []int{0, 1}})
		s.do(enactAction(inst))
		if got := pingCount(s); got != 1 {
			t.Errorf("the cost emptied the hand before the step ran: %d pings, want 1 (FR-TRN-003a)", got)
		}
	})
	t.Run("discarding one of two", func(t *testing.T) {
		s, inst, o := setup(t)
		s.script(0, Answer{Picks: []int{0}})
		s.do(enactAction(inst))
		if got := pingCount(s); got != 0 {
			t.Errorf("a card is still in hand: %d pings, want 0 (FR-TRN-003a)", got)
		}
		if !inHand(s.g.Players[0], o[1]) {
			t.Errorf("the card not discarded stays in hand (FR-TRN-003a)")
		}
	})
}
