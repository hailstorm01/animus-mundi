package engine

import (
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for a destruction replacement applying at most once per sweep (T-0106,
// board directive #81): FR-REP-003 and SVY-R24 (a replacement applies at most
// once to a given asset per kill step or per death sweep), FR-REP-001 (death
// sweep passes, SVY-R23 controller choice) and FR-LAY-002 (the death sweep).
//
// Cards are synthetic. A "damage-instead" source is a structure printing: if
// another allied unit would be destroyed, deal 1 damage to this instead. The
// unit then stays on the battlefield, so it stays lethal if it was lethal.

func damageInsteadReplacement() ability.Replacement {
	return ability.Replacement{
		Event:  ability.WouldBeDestroyed,
		Filter: alliedUnits,
		Steps: []ability.Step{
			{Op: ability.OpDealDamage, Amount: ability.Value{N: 1}, Subject: ability.SubjectSelf},
		},
	}
}

// damageInsteadStructure has enough Health to take every point the tests deal it.
func damageInsteadStructure() *Card {
	c := structure("Damage Instead", 1, 10, Keywords{})
	c.Ability = &ability.Def{Replacements: []ability.Replacement{damageInsteadReplacement()}}
	return c
}

func wantInvariants(t *testing.T, s *scene) {
	t.Helper()
	if err := s.g.CheckInvariants(); err != nil {
		t.Errorf("invariants after the sweep: %v, want none: no lethal asset survives the death sweep (FR-LAY-002, SVY-R24)", err)
	}
}

func TestFR_REP_003_ReplacementAppliesOncePerSweep(t *testing.T) {
	s := newScene(t)
	src := s.put(0, ZoneMainTheater, damageInsteadStructure())
	u := s.put(0, ZoneReserves, unit("Victim", 1, 1, Keywords{}))
	s.sweep(0, alliedUnits)
	wantReplaced(t, s, u)
	if s.alive(u) || !inDiscard(s.g.Players[0], u.Inst) {
		t.Errorf("the replacement does not apply to the unit again, so the sweep's next pass destroys it (FR-REP-003, SVY-R24)")
	}
	s.wantDamage(src, 1)
	wantInvariants(t, s)
}

func TestFR_REP_003_NextPassAppliesAnotherReplacement(t *testing.T) {
	// SVY-R24, SVY-R23: the controller is asked in the first pass, where both
	// apply. In the second pass only the other one applies, so nothing is asked.
	s := newScene(t)
	a := s.put(0, ZoneMainTheater, damageInsteadStructure())
	b := s.put(0, ZoneMainTheater, damageInsteadStructure())
	u := s.put(0, ZoneReserves, unit("Victim", 1, 1, Keywords{}))
	s.script(0, Answer{Picks: []int{1}})
	s.sweep(0, alliedUnits)
	asks := replacementAsks(s, 0)
	if len(asks) != 1 {
		t.Fatalf("player 0 was asked %d replacement questions, want 1: only the first pass has two applicable replacements (SVY-R23, SVY-R24)", len(asks))
	}
	wantOffered(t, asks[0], a.ID, b.ID)
	wantReplaced(t, s, u, u)
	if got := replaceSources(s); len(got) != 2 || got[0] != b.ID || got[1] != a.ID {
		t.Errorf("replace events' sources = %v, want [%d %d]: the chosen B first, then A in the next pass (SVY-R23, SVY-R24)", got, b.ID, a.ID)
	}
	if s.alive(u) || !inDiscard(s.g.Players[0], u.Inst) {
		t.Errorf("both replacements have applied, so the third pass destroys the unit (FR-REP-003, SVY-R24)")
	}
	s.wantDamage(a, 1)
	s.wantDamage(b, 1)
	wantInvariants(t, s)
}

func TestFR_REP_003_ReplacementThatHealsSaves(t *testing.T) {
	s := newScene(t)
	src := structure("Healing Guardian", 1, 10, Keywords{})
	src.Ability = &ability.Def{Replacements: []ability.Replacement{{
		Event:  ability.WouldBeDestroyed,
		Filter: alliedUnits,
		Steps: []ability.Step{
			{Op: ability.OpHeal, Amount: ability.Value{N: 5}, Subject: ability.SubjectReplaced},
		},
	}}}
	s.put(0, ZoneMainTheater, src)
	u := s.put(0, ZoneReserves, unit("Victim", 1, 1, Keywords{}))
	s.sweep(0, alliedUnits)
	wantReplaced(t, s, u)
	if !s.alive(u) {
		t.Fatalf("a replacement that clears the unit's damage saves it: it stays on the battlefield (FR-REP-001, FR-REP-003)")
	}
	s.wantDamage(u, 0)
	if n := eventSeq(s, "destroy", u.ID); n != 0 {
		t.Errorf("the unit is not lethal afterwards, so it is not destroyed, got destroy at seq %d (FR-REP-001)", n)
	}
	wantInvariants(t, s)
}

func TestFR_REP_003_KillThenSweepAppliesAgain(t *testing.T) {
	// SVY-R24: each kill step is its own window. One item damages the unit to
	// 0 Current Health and kills it; the replacement applies to the kill, then
	// once more at the sweep after the item, then the next pass destroys it.
	s := newScene(t)
	src := s.put(0, ZoneMainTheater, damageInsteadStructure())
	u := s.put(0, ZoneReserves, unit("Victim", 1, 1, Keywords{}))
	s.killer(0, allDamage(1, alliedUnits), killStep(alliedUnits))
	wantReplaced(t, s, u, u)
	if s.alive(u) || !inDiscard(s.g.Players[0], u.Inst) {
		t.Errorf("the unit is destroyed in the sweep's next pass after both applications (FR-REP-003, SVY-R24)")
	}
	s.wantDamage(src, 2)
	wantInvariants(t, s)
}

func TestFR_REP_003_ReplacementSavesFromAKill(t *testing.T) {
	// SVY-R24 / FR-REP-001: a killed unit with Health left is not lethal after
	// the replacement, so it is saved.
	s := newScene(t)
	src := s.put(0, ZoneMainTheater, damageInsteadStructure())
	u := s.put(0, ZoneReserves, unit("Victim", 1, 3, Keywords{}))
	s.killer(0, killStep(alliedUnits))
	wantReplaced(t, s, u)
	if !s.alive(u) {
		t.Fatalf("a kill replaced by damage to the source leaves a unit with Health left on the battlefield (FR-REP-001)")
	}
	s.wantDamage(u, 0)
	s.wantDamage(src, 1)
	wantInvariants(t, s)
}
