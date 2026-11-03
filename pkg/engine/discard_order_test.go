package engine

import (
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0137 (FR-TRN-003a, FR-TGT-001, FR-MOD-001, FR-MNV-001; SVY-R2,
// SVY-R3, SVY-R4): a Maneuver's discard cost is picked first, then its modes
// and targets are chosen, then it is pushed and the picked cards are
// discarded. Targets are still judged against the state before any card moves.
// Maneuvers are synthetic, built by discardCostManeuver (discard_cost_test.go).

// dcScript installs a chooser for seat 0 that records every Decision. A
// DecideHand is answered with the options referring to hand, in that order
// (a timeout when timeoutHand); a DecideTarget with the option referring to
// target (an *Asset or an *Instance, by ID). A missing option is a test error.
func dcScript(s *scene, hand []*Instance, timeoutHand bool, target any) {
	if s.asks == nil {
		s.asks = map[PlayerID][]Decision{}
	}
	s.asks[0] = nil
	find := func(d Decision, ref int) int {
		for i, o := range d.Options {
			if o.Ref == ref {
				return i
			}
		}
		s.t.Errorf("no option referring to %d among the options of %q", ref, d.Prompt)
		return -1
	}
	s.g.choosers[0] = func(_ *Game, d Decision) Answer {
		s.asks[0] = append(s.asks[0], d)
		switch d.Kind {
		case DecideHand:
			if timeoutHand {
				return Answer{Timeout: true}
			}
			picks := []int{}
			for _, c := range hand {
				picks = append(picks, find(d, c.ID))
			}
			return Answer{Picks: picks}
		case DecideTarget:
			ref := 0
			switch v := target.(type) {
			case *Asset:
				ref = v.ID
			case *Instance:
				ref = v.ID
			}
			if i := find(d, ref); i >= 0 {
				return Answer{Picks: []int{i}}
			}
		}
		return Answer{Timeout: true}
	}
}

// discardTargetScene gives player 0 the Maneuver "As an additional cost,
// discard up to 3 cards. Deal 1 damage to chosen opposing unit" and two other
// cards, with two opposing units on the battlefield.
func discardTargetScene(t *testing.T) (s *scene, inst *Instance, hand []*Instance, u1, u2 *Asset) {
	s = newScene(t)
	s.g.Players[0].Hand = nil
	opp := tgt(ability.AllegianceOpposing, ability.KindUnit, false)
	inst = s.give(0, discardCostManeuver("Steel Rain", 0, 0, 3, ability.Effect{
		Targets: []ability.Target{opp},
		Steps:   []ability.Step{dmgTo(1, 0)},
	}))
	hand = []*Instance{
		s.give(0, unit("A", 1, 1, Keywords{})),
		s.give(0, unit("B", 1, 1, Keywords{})),
	}
	u1 = s.put(1, ZoneMainTheater, unit("Victim One", 1, 5, Keywords{}))
	u2 = s.put(1, ZoneMainTheater, unit("Victim Two", 1, 5, Keywords{}))
	return
}

func TestFR_TRN_003a_DiscardChosenBeforeTargets(t *testing.T) {
	// Covers: SVY-R3, SVY-R4, FR-TGT-001
	s, inst, o, u1, u2 := discardTargetScene(t)
	p := s.g.Players[0]
	dcScript(s, o, false, u2)
	s.do(enactAction(inst))
	asks := s.asked(0)
	if len(asks) != 2 || asks[0].Kind != DecideHand || asks[1].Kind != DecideTarget {
		t.Fatalf("the discard is picked first, then the target: got %d decisions %v, want [Hand Target] (FR-TRN-003a, SVY-R3)", len(asks), kindsOf(asks))
	}
	if !inDiscard(p, o[0]) || !inDiscard(p, o[1]) {
		t.Errorf("both picked cards end in the discard pile (FR-TRN-003a)")
	}
	if u2.Damage != 1 || u1.Damage != 0 {
		t.Errorf("the chosen unit takes 1 damage and the other none: got %d and %d (FR-TGT-001)", u2.Damage, u1.Damage)
	}
}

func TestFR_TRN_003a_DiscardChosenBeforeTargetsOnTimeout(t *testing.T) {
	// Covers: FR-MOD-001
	s, inst, o, u1, u2 := discardTargetScene(t)
	p := s.g.Players[0]
	dcScript(s, nil, true, u2)
	s.do(enactAction(inst))
	asks := s.asked(0)
	if len(asks) != 2 || asks[0].Kind != DecideHand || asks[1].Kind != DecideTarget {
		t.Fatalf("a timed-out discard is still asked first, and the target is asked after it: got %d decisions %v, want [Hand Target] (FR-MOD-001 sec. 6)", len(asks), kindsOf(asks))
	}
	if !inHand(p, o[0]) || !inHand(p, o[1]) {
		t.Errorf("a timed-out up-to selection selects none, so nothing is discarded (FR-MOD-001 sec. 6)")
	}
	if u2.Damage != 1 || u1.Damage != 0 {
		t.Errorf("the target pick takes effect: got damage %d and %d, want 1 and 0 (FR-TGT-001)", u2.Damage, u1.Damage)
	}
}

func TestFR_TGT_001_DiscardCostCardsNotTargetableInPile(t *testing.T) {
	// Covers: SVY-R2, FR-TRN-003a
	s := newScene(t)
	s.g.Players[0].Hand = nil
	pileSlot := ability.Target{Filter: ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit, Zone: ability.ZoneDiscard}}
	inst := s.give(0, discardCostManeuver("Recall", 0, 0, 2, ability.Effect{
		Targets: []ability.Target{pileSlot},
		Steps:   []ability.Step{ptStep(ability.OpReturnToHand, 0)},
	}))
	piled := s.inDiscard(0, unit("C", 1, 1, Keywords{}), unit("D", 1, 1, Keywords{}))
	hand := []*Instance{s.give(0, unit("A", 1, 1, Keywords{})), s.give(0, unit("B", 1, 1, Keywords{}))}
	dcScript(s, hand, false, piled[0])
	s.do(enactAction(inst))
	var tdec *Decision
	for _, d := range s.asked(0) {
		if d.Kind == DecideTarget {
			d := d
			tdec = &d
		}
	}
	if tdec == nil {
		t.Fatalf("the pile slot is asked as a target decision (SVY-R2)")
	}
	got, want := ppRefs(*tdec), ppInstIDs(piled[0], piled[1])
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("the target options are exactly the cards already in the pile, never the cards picked to be discarded: got %v, want %v (SVY-R2)", got, want)
	}
	p := s.g.Players[0]
	if !inHand(p, piled[0]) {
		t.Errorf("the chosen pile card C returns to hand (SVY-R2)")
	}
	if !inDiscard(p, hand[0]) || !inDiscard(p, hand[1]) {
		t.Errorf("A and B end in the discard pile (FR-TRN-003a)")
	}
}

func TestFR_TRN_003a_DiscardCostEventsStillAfterPlace(t *testing.T) {
	// Covers: FR-MNV-001
	s, inst, o, _, u2 := discardTargetScene(t)
	dcScript(s, []*Instance{o[1], o[0]}, false, u2) // pick B before A
	s.do(enactAction(inst))
	var place, dmg, dA, dB int
	for _, e := range s.g.Log {
		switch {
		case e.Kind == "place" && e.Card == "Steel Rain":
			place = e.Seq
		case e.Kind == "discard" && e.Inst == o[0].ID:
			dA = e.Seq
		case e.Kind == "discard" && e.Inst == o[1].ID:
			dB = e.Seq
		case e.Kind == "damage" && e.Target == u2.ID && e.Amount == 1:
			dmg = e.Seq
		}
	}
	if place == 0 || dmg == 0 || dA == 0 || dB == 0 {
		t.Fatalf("missing events: place %d, damage %d, discard A %d, discard B %d", place, dmg, dA, dB)
	}
	if dB < place || dA > dmg {
		t.Errorf("the cost is paid after the Maneuver is placed (%d) and before it resolves (%d): discards at %d and %d (FR-TRN-003a)", place, dmg, dB, dA)
	}
	if dB > dA {
		t.Errorf("the discards are logged in pick order, B (seq %d) before A (seq %d) (FR-TRN-003a)", dB, dA)
	}
}
