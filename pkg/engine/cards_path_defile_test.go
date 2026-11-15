package engine

import (
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Card tests for Blind Defile (T-0242), written from the CSV rules text in
// data/cards/units-and-structures-path.csv: "Whenever an opposing unit attacks,
// you may kill this structure. If you do, kill that attacking unit."
// Rulings: SVY-R49 (the attacker's kill depends on the self-kill), SVY-R9
// ("you may" asked on placement), SVY-R15 / FR-REP-001 (a replaced kill still
// counts as done), FR-TGT-001 ("You May X. If You Do, Y"), FR-CBT-002 Step 2
// (attack triggers resolve before combat damage), FR-TRG-003, FR-ZON-004.
// Player 1 controls the Defile; player 0 attacks player 1's Headquarters with
// a 3/9 vanilla unit, so only the kill can remove it.

func newDefileScene(t *testing.T) (*scene, *Asset, *Asset) {
	t.Helper()
	s := newScene(t)
	d := s.put(1, ZoneMainTheater, s.card("Blind Defile"))
	att := s.put(0, ZoneMainTheater, unit("Hitter", 3, 9, Keywords{}))
	return s, d, att
}

func TestCard_BlindDefile_AcceptedKillsItselfAndTheAttacker(t *testing.T) {
	s, d, att := newDefileScene(t)
	s.script(0)
	s.script(1, yes)
	s.do(s.attack(att, s.g.Players[1].HQ))
	if got := kindsOf(s.asked(1)); len(got) != 1 || got[0] != DecideOptional {
		t.Errorf("the Defile's controller is asked \"you may\" once (SVY-R9): asked %v", got)
	}
	if s.alive(d) || !inDiscard(s.g.Players[1], d.Inst) {
		t.Errorf("accepted, the Defile kills itself and is in its owner's discard pile (FR-TGT-001)")
	}
	if s.alive(att) || !inDiscard(s.g.Players[0], att.Inst) {
		t.Errorf("if you do, the attacking unit is killed and is in its owner's discard pile (SVY-R49)")
	}
	s.wantDamage(s.g.Players[1].HQ, 0)
}

func TestCard_BlindDefile_DeclinedOrTimedOutKillsNothing(t *testing.T) {
	for _, tc := range []struct {
		name string
		ans  Answer
	}{{"Declined", no}, {"TimedOut", Answer{Timeout: true}}} {
		t.Run(tc.name, func(t *testing.T) {
			s, d, att := newDefileScene(t)
			s.script(0)
			s.script(1, tc.ans)
			items := recordPlaced(s)
			s.do(s.attack(att, s.g.Players[1].HQ))
			if len(*items) != 1 || (*items)[0].Kind != StackAttack {
				t.Errorf("a declined or timed-out optional trigger is not placed (FR-MOD-001 sections 6, 7): placed %+v", *items)
			}
			if !s.alive(d) || !s.alive(att) {
				t.Errorf("nothing is killed when \"you may\" is not accepted: Defile alive %v, attacker alive %v", s.alive(d), s.alive(att))
			}
			s.wantDamage(s.g.Players[1].HQ, 3)
		})
	}
}

func TestCard_BlindDefile_TwoDefilesBothAccepted(t *testing.T) {
	s, d1, att := newDefileScene(t)
	d2 := s.put(1, ZoneMainTheater, s.card("Blind Defile"))
	s.script(0)
	// Answer by decision kind, so the test does not pin when the order is asked.
	var kinds []DecisionKind
	s.g.choosers[1] = func(_ *Game, d Decision) Answer {
		kinds = append(kinds, d.Kind)
		if d.Kind == DecideOrder {
			return Answer{Picks: []int{0, 1}}
		}
		return yes
	}
	s.do(s.attack(att, s.g.Players[1].HQ))
	order := 0
	for _, k := range kinds {
		if k == DecideOrder {
			order++
		}
	}
	if order != 1 {
		t.Errorf("two simultaneous triggers of one controller are ordered by it with one DecideOrder (FR-TRG-003): asked %v", kinds)
	}
	if s.alive(d1) || s.alive(d2) || s.alive(att) {
		t.Errorf("both Defiles and the attacker die (FR-TGT-001): d1 alive %v, d2 alive %v, attacker alive %v", s.alive(d1), s.alive(d2), s.alive(att))
	}
	s.wantDamage(s.g.Players[1].HQ, 0)
	n := 0
	for _, e := range eventsOfKind(s.g.Log, "destroy") {
		if e.Target == att.ID {
			n++
		}
	}
	if n != 1 {
		t.Errorf("the attacker is killed once: %d destroy events name it", n)
	}
}

func TestCard_BlindDefile_DefileGoneBeforeResolutionKillsNothing(t *testing.T) {
	s, d, att := newDefileScene(t)
	s.script(0)
	s.script(1, yes)
	s.declareThen(s.attack(att, s.g.Players[1].HQ), func() {
		s.g.flushTriggers()
		if n := len(s.g.Stack); n != 2 {
			t.Fatalf("the Attack and the Defile's trigger are on the stack: %d items", n)
		}
		s.g.leaveBattlefield(d, s.g.effectiveState(d))
	})
	if !s.alive(att) {
		t.Errorf("the Defile left before the trigger resolved, so the self-kill selects nothing and the attacker is not killed (FR-TGT-001, SVY-R49)")
	}
	s.wantDamage(s.g.Players[1].HQ, 3)
}

func TestCard_BlindDefile_ReplacedSelfKillStillKillsTheAttacker(t *testing.T) {
	// Covers: SVY-R15
	s, d, att := newDefileScene(t)
	guardian := structure("Guardian", 0, 9, Keywords{})
	guardian.Ability = &ability.Def{Replacements: []ability.Replacement{{
		Event:  ability.WouldBeDestroyed,
		Filter: ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindStructure},
		Steps:  []ability.Step{{Op: ability.OpReturnToHand, Subject: ability.SubjectReplaced}},
	}}}
	s.put(1, ZoneMainTheater, guardian)
	s.script(0)
	s.script(1, yes)
	s.do(s.attack(att, s.g.Players[1].HQ))
	if s.alive(d) || !inHand(s.g.Players[1], d.Inst) || inDiscardPile(s.g.Players[1], d.Inst) {
		t.Errorf("the replacement returns the Defile to its owner's hand instead of destroying it (FR-REP-001)")
	}
	if s.alive(att) || !inDiscard(s.g.Players[0], att.Inst) {
		t.Errorf("a replaced kill still counts as done, so the attacker is killed (FR-REP-001, SVY-R15)")
	}
}
