package engine

import (
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0131 (FR-TRN-003a, FR-TRN-003, FR-MNV-001, FR-MOD-001,
// FR-TGT-001; SVY-R3, SVY-R29): a Maneuver whose additional cost is "discard
// up to N cards". The cost is paid on declaration, after the Maneuver is on
// the stack, from the payer's own hand, possibly none. Maneuvers are
// synthetic; their step deals 1 damage to the opposing Headquarters so the
// event log shows when it ran (a draw logs no event). Player 0 is the active
// player and enacts.

// discardCostManeuver is maneuver(...) with the printed cost "As an
// additional cost to enact this maneuver, discard up to upTo cards."
func discardCostManeuver(name string, req, tac, upTo int, e ability.Effect) *Card {
	c := maneuver(name, req, tac, e)
	c.Ability.ExtraCosts = []ability.Cost{{Kind: ability.CostDiscard, UpTo: upTo}}
	return c
}

// hqPing is a step dealing 1 damage to the opposing Headquarters.
func hqPing() ability.Step {
	return ability.Step{Op: ability.OpDealDamage, Amount: ability.Value{N: 1}, Subject: ability.SubjectAll,
		Filter: ability.Filter{Allegiance: ability.AllegianceOpposing, Kinds: ability.KindHeadquarters}}
}

// discardCostScene gives player 0 an empty hand, then the cost Maneuver and
// n other cards. It returns the Maneuver and the other cards in hand order.
func discardCostScene(t *testing.T, upTo, n int) (*scene, *Instance, []*Instance) {
	s := newScene(t)
	s.g.Players[0].Hand = nil
	inst := s.give(0, discardCostManeuver("Steel Rain", 1, 1, upTo, steps(hqPing())))
	var others []*Instance
	for i := 0; i < n; i++ {
		others = append(others, s.give(0, unit(string(rune('A'+i)), 1, 1, Keywords{})))
	}
	p := s.g.Players[0]
	p.Req, p.Tac = 1, 1
	return s, inst, others
}

// pingCount counts damage events of 1 to player 1's Headquarters.
func pingCount(s *scene) int {
	n := 0
	for _, e := range s.g.Log {
		if e.Kind == "damage" && e.Target == s.g.Players[1].HQ.ID && e.Amount == 1 {
			n++
		}
	}
	return n
}

func TestFR_TRN_003a_DiscardCostPaidOnDeclaration(t *testing.T) {
	// Covers: FR-TRN-003, FR-MNV-001
	s, inst, o := discardCostScene(t, 3, 4)
	p := s.g.Players[0]
	s.script(0, Answer{Picks: []int{0, 2}})
	s.do(enactAction(inst))
	if !inDiscard(p, o[0]) || !inDiscard(p, o[2]) {
		t.Errorf("the two cards the controller chose are discarded as the cost (FR-TRN-003a)")
	}
	if !inHand(p, o[1]) || !inHand(p, o[3]) {
		t.Errorf("the two cards not chosen stay in hand (FR-TRN-003a)")
	}
	if got := pingCount(s); got != 1 {
		t.Errorf("the Maneuver's step still ran: %d damage events, want 1 (FR-TRN-003a)", got)
	}
	if lastDiscard(p) != inst {
		t.Errorf("the Maneuver goes to its owner's discard pile once resolved (FR-MNV-001)")
	}
	if p.Req != 0 || p.Tac != 0 {
		t.Errorf("enacting pays the Maneuver's resource cost R1 T1: pool R%d T%d, want R0 T0 (FR-TRN-003)", p.Req, p.Tac)
	}
	var place, dmg int
	var discards []int
	for _, e := range s.g.Log {
		switch {
		case e.Kind == "place" && e.Card == "Steel Rain":
			place = e.Seq
		case e.Kind == "discard" && (e.Inst == o[0].ID || e.Inst == o[2].ID):
			discards = append(discards, e.Seq)
			if e.Zone != "Hand" {
				t.Errorf("a cost discard is from the hand, got zone %q (FR-TRN-003a)", e.Zone)
			}
		case e.Kind == "damage" && e.Target == s.g.Players[1].HQ.ID && e.Amount == 1:
			dmg = e.Seq
		}
	}
	if place == 0 || dmg == 0 || len(discards) != 2 {
		t.Fatalf("missing events: Maneuver place %d, damage %d, hand discards %d (want 2)", place, dmg, len(discards))
	}
	for _, d := range discards {
		if d < place || d > dmg {
			t.Errorf("the cost is paid after the Maneuver is declared (placed, seq %d) and before it resolves (damage, seq %d), got discard at %d (FR-TRN-003a)", place, dmg, d)
		}
	}
}

func TestFR_TRN_003a_DiscardCostMayDiscardNone(t *testing.T) {
	s, inst, o := discardCostScene(t, 3, 3)
	p := s.g.Players[0]
	s.script(0, Answer{Picks: []int{}})
	s.do(enactAction(inst))
	for i, c := range o {
		if !inHand(p, c) {
			t.Errorf("card %d stays in hand: discarding none pays an up-to cost (FR-TRN-003a, FR-MOD-001 sec. 6)", i)
		}
	}
	if got := pingCount(s); got != 1 {
		t.Errorf("the Maneuver's step ran: %d damage events, want 1 (FR-TRN-003a)", got)
	}
}

func TestFR_TRN_003a_DiscardCostOffersTheRestOfTheHand(t *testing.T) {
	// Covers: FR-TRN-003
	for _, tc := range []struct{ others, wantMax int }{{2, 2}, {5, 3}} {
		s, inst, o := discardCostScene(t, 3, tc.others)
		s.script(0, Answer{Picks: []int{}})
		s.do(enactAction(inst))
		asks := s.asked(0)
		if len(asks) != 1 {
			t.Errorf("%d other cards: the controller is asked once, got %d decisions (FR-TRN-003a)", tc.others, len(asks))
			continue
		}
		d := asks[0]
		if d.Player != 0 || d.Kind != DecideHand {
			t.Errorf("%d other cards: asked of the controller as a Hand decision, got player %d kind %v (FR-TRN-003a)", tc.others, d.Player, d.Kind)
		}
		if d.Min != 0 || d.Max != tc.wantMax {
			t.Errorf("%d other cards, up to 3: Min/Max = %d/%d, want 0/%d (FR-MOD-001 sec. 6)", tc.others, d.Min, d.Max, tc.wantMax)
		}
		if len(d.Options) != tc.others {
			t.Fatalf("%d other cards: options are exactly the other cards in hand, got %d", tc.others, len(d.Options))
		}
		for i, opt := range d.Options {
			if opt.Ref == inst.ID {
				t.Errorf("the Maneuver itself is on the stack and is never a discard option (FR-TRN-003, FR-MNV-001)")
			}
			if opt.Ref != o[i].ID {
				t.Errorf("option %d refers to instance %d, want hand card %d", i, opt.Ref, o[i].ID)
			}
		}
	}
}

func TestFR_TRN_003a_DiscardCostWithAnEmptyHand(t *testing.T) {
	// Covers: SVY-R29
	s, inst, _ := discardCostScene(t, 3, 0)
	if !offersEnact(s.g, inst) {
		t.Errorf("an up-to discard cost is paid by discarding none, so the Maneuver is offered with no other cards in hand (SVY-R29, FR-TRN-003a)")
	}
	s.script(0)
	s.do(enactAction(inst))
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("with no other card to discard nothing is asked, got %d decisions (FR-MOD-001 sec. 1)", n)
	}
	if got := pingCount(s); got != 1 {
		t.Errorf("the Maneuver's step ran: %d damage events, want 1 (FR-TRN-003a)", got)
	}
}

func TestFR_MOD_001_DiscardCostTimeoutDiscardsNone(t *testing.T) {
	// Covers: FR-TRN-003a
	s, inst, o := discardCostScene(t, 3, 2)
	p := s.g.Players[0]
	s.script(0, Answer{Timeout: true})
	s.do(enactAction(inst))
	if !inHand(p, o[0]) || !inHand(p, o[1]) {
		t.Errorf("a timed-out up-to selection selects none, so nothing is discarded (FR-MOD-001 sec. 6)")
	}
	if got := pingCount(s); got != 1 {
		t.Errorf("the Maneuver's step ran: %d damage events, want 1 (FR-TRN-003a)", got)
	}
}

func TestFR_MOD_001_DiscardCostAskedWithOneCard(t *testing.T) {
	// Covers: FR-TRN-003a
	s, inst, o := discardCostScene(t, 3, 1)
	p := s.g.Players[0]
	s.script(0, Answer{Picks: []int{0}})
	s.do(enactAction(inst))
	asks := s.asked(0)
	if len(asks) != 1 {
		t.Fatalf("none or the one card are both valid answers, so the decision is still asked, got %d decisions (FR-MOD-001 sec. 1)", len(asks))
	}
	if asks[0].Min != 0 || asks[0].Max != 1 {
		t.Errorf("Min/Max = %d/%d, want 0/1 (FR-MOD-001 sec. 6)", asks[0].Min, asks[0].Max)
	}
	if !inDiscard(p, o[0]) {
		t.Errorf("the chosen card is discarded (FR-TRN-003a)")
	}
}
