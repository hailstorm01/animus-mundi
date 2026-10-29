package engine

import (
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0099 (board directive #73): a combat-kill trigger fires only when
// the death sweep after the Attack destroys the defender. A defender whose
// destruction is replaced was not killed, so the attacker's combat-kill trigger
// is never placed (FR-CBT-002 Step 3, SVY-R19, SVY-R15, FR-REP-001). Everything
// T-0080 pinned in combat_kill_test.go stays true.
//
// Replacements come from the guardian cards in replacement_test.go (player 1's
// guardian returns an allied unit to hand instead of destroying it) and from the
// real Shifting Impossibility. Combat-kill triggers are synthetic.

// Covers: FR-REP-001, FR-TRG-004
func TestFR_CBT_002_CombatKillTriggerDoesNotFireWhenTheDeathIsReplaced(t *testing.T) {
	s := wmScene(t)
	p0 := s.g.Players[0]
	s.inDiscard(0, unit("Fallen", 1, 1, Keywords{}))
	pile, hand := len(p0.Discard), len(p0.Hand)
	si := s.put(1, ZoneReserves, s.card("Shifting Impossibility"))
	sc := s.putScratch(1)
	s.moveTo(sc, ZoneMainTheater)
	hq := p0.HQ
	s.script(0)
	s.do(s.attack(hq, sc))
	if s.alive(sc) || !inHand(s.g.Players[1], sc.Inst) {
		t.Fatalf("precondition: the Scratch Company's death is replaced and it returns to its owner's hand")
	}
	s.wantDamage(si, 1)
	if got := placedOrder(s, hq); len(got) != 0 {
		t.Errorf("a replaced death is not a kill, so the Watchman's combat-kill trigger is never placed, got placements %v (FR-CBT-002, SVY-R19)", got)
	}
	if len(p0.Discard) != pile || len(p0.Hand) != hand {
		t.Errorf("no card returns from player 0's discard pile: pile %d -> %d, hand %d -> %d (SVY-R19)", pile, len(p0.Discard), hand, len(p0.Hand))
	}
	s.wantSettled("the replaced Attack finished")
}

// Covers: FR-REP-001, FR-TRG-004
func TestFR_CBT_002_CombatKillTriggerFiresWhenAnotherDeathInTheSweepIsReplaced(t *testing.T) {
	s := newScene(t)
	s.put(1, ZoneMainTheater, guardianStructure(5))
	att := s.put(0, ZoneMainTheater, striker("Slayer", 3, 9, Keywords{}, combatKillTrigger(scratchN(1))))
	def := s.put(1, ZoneMainTheater, structure("Bunker", 0, 3, Keywords{}))
	other := s.put(1, ZoneReserves, unit("Bystander", 1, 1, Keywords{}))
	other.Damage = 1 // lethal at the same sweep as the defender
	s.do(s.attack(att, def))
	if s.alive(def) {
		t.Fatalf("precondition: the structure defender, not covered by the guardian, is destroyed")
	}
	wantReplaced(t, s, other)
	s.wantScratches(0, 1, "the defender was destroyed, so the trigger fires once even though another death in the sweep was replaced: the gate is per defender (FR-CBT-002 Step 3, SVY-R19)")
}

// Covers: FR-REP-001, FR-TRG-004
func TestFR_CBT_002_ReplacedDefenderGivesNoTriggerWhenTheAttackerAlsoDies(t *testing.T) {
	s := newScene(t)
	s.put(1, ZoneMainTheater, guardianStructure(5))
	att := s.put(0, ZoneMainTheater, striker("Slayer", 3, 1, Keywords{}, combatKillTrigger(scratchN(1))))
	def := s.put(1, ZoneMainTheater, unit("Victim", 1, 3, Keywords{}))
	s.do(s.attack(att, def))
	if s.alive(att) {
		t.Fatalf("precondition: the attacker is destroyed by retaliation in the same sweep")
	}
	if s.alive(def) || !inHand(s.g.Players[1], def.Inst) {
		t.Fatalf("precondition: the defender's death is replaced and it returns to its owner's hand")
	}
	wantReplaced(t, s, def)
	s.wantScratches(0, 0, "the defender was replaced, not destroyed, so the combat-kill trigger does not fire although the attacker died in the same sweep (FR-CBT-002 Step 3, SVY-R19)")
}

// Covers: FR-REP-001, FR-TRG-003
func TestFR_CBT_002_DroppedCombatKillTriggerIsNotOfferedForOrdering(t *testing.T) {
	// The combat-kill trigger would Scratch 3 and the damage-taken trigger
	// Scratches 1, so a total of 1 shows only the damage-taken trigger resolved.
	s := newScene(t)
	s.put(1, ZoneMainTheater, guardianStructure(5))
	damaged := ability.Trigger{Hook: ability.HookDamaged, Effect: ability.Effect{Steps: []ability.Step{scratchN(1)}}}
	att := s.put(0, ZoneMainTheater, striker("Slayer", 3, 5, Keywords{}, combatKillTrigger(scratchN(3)), damaged))
	def := s.put(1, ZoneMainTheater, unit("Victim", 1, 3, Keywords{}))
	s.script(0, Answer{Picks: []int{1, 0}})
	s.do(s.attack(att, def))
	if !s.alive(att) || s.alive(def) || !inHand(s.g.Players[1], def.Inst) {
		t.Fatalf("precondition: the attacker survives retaliation and the defender's death is replaced")
	}
	noOrderAsked(t, s, 0)
	s.wantScratches(0, 1, "only the damage-taken trigger resolves; the dropped combat-kill trigger is neither ordered nor resolved, and a single trigger is not asked about (FR-TRG-003, SVY-R19)")
}
