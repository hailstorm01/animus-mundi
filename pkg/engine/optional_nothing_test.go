package engine

import (
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0238 (directive #152): an optional ("you may") trigger is not
// asked, and not placed, when accepting could do nothing (SVY-R48, FR-MOD-001
// section 7): it has target slots, every step acts only on them, and no slot
// has a legal entity at placement. A mandatory trigger is still placed with
// nothing legal (SVY-R30, FR-TGT-003). Cards are synthetic; scenes are the
// arena of optional_upto_test.go.

// noOpposingStructure matches nothing in the arena: it has no structure.
var noOpposingStructure = ability.Filter{Allegiance: ability.AllegianceOpposing, Kinds: ability.KindStructure}

// optionalAsks counts the "you may" questions in ds.
func optionalAsks(ds []Decision) int { return countKind(ds, DecideOptional) }

// targetAsks returns the DecideTarget questions in ds, in order.
func targetAsks(ds []Decision) []Decision {
	var out []Decision
	for _, d := range ds {
		if d.Kind == DecideTarget {
			out = append(out, d)
		}
	}
	return out
}

// picksOf turns assets into scripted picks, one answer each.
func picksOf(as []*Asset) []tpick {
	out := make([]tpick, len(as))
	for i, a := range as {
		out[i] = pick(a)
	}
	return out
}

// abilityItemsFrom counts the placed ability items whose Source is a.
func abilityItemsFrom(items []StackItem, a *Asset) int {
	n := 0
	for _, it := range items {
		if it.Kind == StackAbility && it.Source == a.ID {
			n++
		}
	}
	return n
}

// Covers: SVY-R48
func TestFR_MOD_001_OptionalTriggerWithNoLegalTargetIsNotAskedOrPlaced(t *testing.T) {
	a := newArena(t, 1, attackTrig(true, []ability.Target{{Filter: noOpposingStructure}}, dmgTo(3, 0)))
	a.s.scriptPlan(0, nil, []Answer{yes})
	a.s.scriptPlan(1, nil, nil)
	items := recordPlaced(a.s)
	a.attack()
	if n := len(a.s.asked(0)); n != 0 {
		t.Errorf("an optional trigger whose only slot has no legal entity could do nothing and is not asked: asked %d times (SVY-R48, FR-MOD-001 sec. 7)", n)
	}
	if n := abilityItemsFrom(*items, a.hero); n != 0 {
		t.Errorf("it is not placed either: %d ability items placed (SVY-R48)", n)
	}
	wantNoNonCombatDamage(t, a.s.g, "nothing was placed, so nothing dealt damage")
}

// Covers: SVY-R48, SVY-R11
func TestFR_MOD_001_OptionalUpToTriggerWithNoLegalTargetIsNotAsked(t *testing.T) {
	a := newArena(t, 1, attackTrig(true, []ability.Target{upTo(2, noOpposingStructure)}, dmgTo(3, 0)))
	a.s.scriptPlan(0, nil, []Answer{yes})
	a.s.scriptPlan(1, nil, nil)
	items := recordPlaced(a.s)
	a.attack()
	if n := len(a.s.asked(0)); n != 0 {
		t.Errorf("an 'up to 2' slot with no legal entity could do nothing, so the 'you may' is not asked: asked %d times (SVY-R48)", n)
	}
	if n := abilityItemsFrom(*items, a.hero); n != 0 {
		t.Errorf("it is not placed either: %d ability items placed (SVY-R48)", n)
	}
	wantNoNonCombatDamage(t, a.s.g, "nothing was placed, so nothing dealt damage")
}

// Covers: SVY-R48
func TestFR_MOD_001_OptionalTriggerWithOneFillableSlotOfTwoIsStillAsked(t *testing.T) {
	a := newArena(t, 1, attackTrig(true,
		[]ability.Target{{Filter: noOpposingStructure}, {Filter: opposingUnit}}, dmgTo(1, 0), dmgTo(1, 1)))
	a.s.scriptPlan(0, nil, []Answer{yes}, pick(a.e1))
	a.s.scriptPlan(1, nil, nil)
	a.attack()
	if n := optionalAsks(a.s.asked(0)); n != 1 {
		t.Errorf("one slot has a legal entity, so accepting could do something and the 'you may' is asked once: asked %d times (SVY-R48)", n)
	}
}

// Covers: SVY-R48, SVY-R14
func TestFR_MOD_001_OptionalTriggerWithAnUntargetedStepIsStillAsked(t *testing.T) {
	draw := ability.Step{Op: ability.OpDraw, Subject: ability.SubjectController, Amount: ability.Value{N: 1}}
	a := newArena(t, 1, attackTrig(true, []ability.Target{{Filter: noOpposingStructure}}, dmgTo(1, 0), draw))
	a.s.scriptPlan(0, nil, []Answer{no})
	a.s.scriptPlan(1, nil, nil)
	a.attack()
	if n := optionalAsks(a.s.asked(0)); n != 1 {
		t.Errorf("a step that does not act on the empty slot can still do something, so the 'you may' is asked once: asked %d times (SVY-R48)", n)
	}
}

// Covers: SVY-R30, FR-TGT-003
func TestFR_MOD_001_MandatoryTriggerWithNoLegalTargetIsStillPlaced(t *testing.T) {
	a := newArena(t, 1, attackTrig(false, []ability.Target{{Filter: noOpposingStructure}}, dmgTo(3, 0)))
	a.s.scriptPlan(0, nil, nil)
	a.s.scriptPlan(1, nil, nil)
	items := recordPlaced(a.s)
	a.attack()
	if n := abilityItemsFrom(*items, a.hero); n != 1 {
		t.Errorf("a mandatory trigger is placed even with nothing legal to target: %d ability items placed, want 1 (SVY-R30)", n)
	}
	if n := len(a.s.asked(0)); n != 0 {
		t.Errorf("nothing is asked of a mandatory trigger with nothing legal: asked %d times", n)
	}
	wantNoNonCombatDamage(t, a.s.g, "the placed trigger resolves doing nothing (FR-TGT-003)")
	if len(a.s.g.Stack) != 0 {
		t.Errorf("the stack is empty once it has resolved: %d items", len(a.s.g.Stack))
	}
}
