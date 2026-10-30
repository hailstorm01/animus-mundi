package engine

import (
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0107 (board directive #82): a chosen asset that an earlier step
// of the same effect removed from the battlefield is no longer that target, so
// a later step does nothing to it and still acts on the rest of its slot
// (FR-ZON-004: a unit that left is a new object; FR-TGT-003 partial
// invalidation). Cards are synthetic. Every scene deploys a unit "Source" from
// player 0's hand (see iydDeploy); its deploy trigger runs the steps under
// test. Player 1 is the opponent.

func rtStep(op ability.Op, n int, slot int) ability.Step {
	return ability.Step{Op: op, Subject: ability.SubjectTarget, Target: slot, Amount: ability.Value{N: n}}
}

func TestFR_ZON_004_ReturnedTargetTakesNoLaterDamage(t *testing.T) {
	// Covers: FR-TGT-003
	s := newScene(t)
	e := s.put(1, ZoneMainTheater, mortal("Enemy", 1, 6,
		ability.Trigger{Hook: ability.HookDamaged, Effect: ability.Effect{Steps: []ability.Step{drawN(1)}}}))
	inst := e.Inst
	deck := len(s.g.Players[1].Deck)
	s.script(0, Answer{Picks: []int{0}})
	s.iydDeploy(iydTrigger(false, rlEnemyTarget,
		rlStep(ability.OpReturnToHand, ability.SubjectTarget, false),
		rtStep(ability.OpDealDamage, 2, 0)))
	if s.alive(e) || !inHand(s.g.Players[1], inst) {
		t.Fatalf("precondition: the returned unit's card is in its owner's hand")
	}
	if ds := damageTo(s.g.Log, e); len(ds) != 0 {
		t.Errorf("a unit that left the battlefield is no longer the target, so it takes no later damage (FR-ZON-004): %d damage events", len(ds))
	}
	if got := len(s.g.Players[1].Deck); got != deck {
		t.Errorf("its damage-taken trigger did not fire (no draw): deck %d -> %d (FR-ZON-004)", deck, got)
	}
}

func TestFR_ZON_004_KilledTargetGetsNoLaterModifier(t *testing.T) {
	// Covers: FR-TGT-003, SVY-R14
	s := newScene(t)
	e := s.put(1, ZoneMainTheater, unit("Enemy", 2, 6, Keywords{}))
	deck := len(s.g.Players[0].Deck)
	s.script(0, Answer{Picks: []int{0}})
	s.iydDeploy(iydTrigger(false, rlEnemyTarget,
		ability.Step{Op: ability.OpKill, Subject: ability.SubjectTarget},
		rtStep(ability.OpModifyStat, 2, 0),
		iydIfDone(drawN(1))))
	if s.alive(e) {
		t.Fatalf("precondition: the kill step destroyed the chosen unit")
	}
	if got := len(s.g.Players[0].Deck); got != deck {
		t.Errorf("the modifier step acted on nothing and counts 0, so the IfDone draw does not run: deck %d -> %d (FR-ZON-004, SVY-R14)", deck, got)
	}
}

func TestFR_TGT_003_RemovedTargetLeavesTheRestOfItsSlot(t *testing.T) {
	// Covers: FR-ZON-004
	s := newScene(t)
	x := s.put(1, ZoneMainTheater, unit("X", 1, 6, Keywords{}))
	y := s.put(1, ZoneMainTheater, unit("Y", 1, 6, Keywords{}))
	idx := func(d Decision, a *Asset) int {
		for i, o := range d.Options {
			if o.Ref == a.ID {
				return i
			}
		}
		s.t.Fatalf("asset %s is not among the options of %q", a, d.Prompt)
		return 0
	}
	// Each slot is asked on its own: slot 0 (up to 2) gets X and Y, slot 1 gets X.
	s.g.choosers[0] = func(_ *Game, d Decision) Answer {
		if d.Max == 2 {
			return Answer{Picks: []int{idx(d, x), idx(d, y)}}
		}
		return Answer{Picks: []int{idx(d, x)}}
	}
	s.iydDeploy(iydTrigger(false,
		[]ability.Target{{Filter: iydEnemy, UpTo: 2}, {Filter: iydEnemy}},
		ability.Step{Op: ability.OpReturnToHand, Subject: ability.SubjectTarget, Target: 1},
		rtStep(ability.OpDealDamage, 1, 0)))
	if s.alive(x) || !inHand(s.g.Players[1], x.Inst) {
		t.Fatalf("precondition: slot 1's unit X was returned to its owner's hand")
	}
	s.wantDamage(y, 1)
	if ds := damageTo(s.g.Log, x); len(ds) != 0 {
		t.Errorf("X left the battlefield, so it is skipped in slot 0: %d damage events (FR-TGT-003 partial invalidation)", len(ds))
	}
}

func TestFR_ZON_004_ReplacedKillLeavesNoTarget(t *testing.T) {
	// Covers: FR-REP-001
	s := newScene(t)
	g := structure("Rescuer", 1, 3, Keywords{})
	g.Ability = &ability.Def{Replacements: []ability.Replacement{{
		Event:  ability.WouldBeDestroyed,
		Filter: ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit, Other: true},
		Steps:  []ability.Step{{Op: ability.OpReturnToHand, Subject: ability.SubjectReplaced}},
	}}}
	s.put(1, ZoneMainTheater, g)
	e := s.put(1, ZoneMainTheater, unit("Enemy", 1, 6, Keywords{}))
	s.script(0, Answer{Picks: []int{0}})
	s.iydDeploy(iydTrigger(false, rlEnemyTarget,
		ability.Step{Op: ability.OpKill, Subject: ability.SubjectTarget},
		rtStep(ability.OpDealDamage, 1, 0)))
	if s.alive(e) || !inHand(s.g.Players[1], e.Inst) {
		t.Fatalf("precondition: the replaced kill returned the unit to its owner's hand")
	}
	if ds := damageTo(s.g.Log, e); len(ds) != 0 {
		t.Errorf("the replaced unit left the battlefield, so the later step has no target: %d damage events (FR-ZON-004)", len(ds))
	}
}
