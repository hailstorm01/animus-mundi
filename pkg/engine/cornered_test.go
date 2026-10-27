package engine

import (
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for the Cornered gate on triggers and the power-until-end-of-turn step
// (T-0069). Cornered (FR-KWD-002e) is active for a player whenever their
// opponent controls strictly more battlefield assets (units and structures,
// face-down Scratch Company units included) in the Main Theater than they do.
// A triggered ability is gated at the moment its event happens; once placed it
// resolves regardless (FR-LAY-003). OpModifyStat gives the Subject +Amount power
// until end of turn (Layer 5, FR-LAY-001), expiring in Sub-Phase 4c
// (FR-TRN-001). Cards are synthetic; the observation is combat damage (Step 3),
// effective Power, Damage and the log's placed triggers.
//
// This file does not compile until ability.Trigger has a Cornered bool field
// and ability has OpModifyStat; that is expected until T-0069 is implemented.

// cornered marks trig as a "Cornered: ..." ability.
func cornered(trig ability.Trigger) ability.Trigger {
	trig.Cornered = true
	return trig
}

// modifyPower is "<subject> gets +n power until end of turn".
func modifyPower(n int, subj ability.Subject) ability.Step {
	return ability.Step{Op: ability.OpModifyStat, Amount: ability.Value{N: n}, Subject: subj}
}

// modifyTarget is "chosen <slot> gets +n power until end of turn".
func modifyTarget(n, slot int) ability.Step {
	return ability.Step{Op: ability.OpModifyStat, Amount: ability.Value{N: n}, Subject: ability.SubjectTarget, Target: slot}
}

// targetedAttack is one "Whenever this unit attacks" ability with chosen targets.
func targetedAttack(targets []ability.Target, steps ...ability.Step) ability.Trigger {
	return ability.Trigger{Hook: ability.HookAttacks, Effect: ability.Effect{Targets: targets, Steps: steps}}
}

// wall is a harmless 0-power, 10-Health defender.
func wall(name string) *Card { return unit(name, 0, 10, Keywords{}) }

// --- FR-KWD-002e: the gate ---

func TestFR_KWD_002e_CorneredTriggerFiresWhenTheOpponentHasMoreInTheMainTheater(t *testing.T) {
	s := newScene(t)
	att := s.put(0, ZoneMainTheater, striker("Raider", 1, 9, Keywords{}, cornered(attackTrigger(modifyPower(3, ability.SubjectSelf)))))
	def := s.put(1, ZoneMainTheater, wall("Wall"))
	s.put(1, ZoneMainTheater, wall("Second"))
	s.do(s.attack(att, def))
	s.wantDamage(def, 4)
}

func TestFR_KWD_002e_CorneredTriggerDoesNotFireOnEqualCounts(t *testing.T) {
	s := newScene(t)
	s.script(0) // any decision is a failure
	att := s.put(0, ZoneMainTheater, striker("Raider", 1, 9, Keywords{}, cornered(attackTrigger(modifyPower(3, ability.SubjectSelf)))))
	def := s.put(1, ZoneMainTheater, wall("Wall"))
	s.do(s.attack(att, def))
	s.wantDamage(def, 1)
	if got := placedOrder(s, att); len(got) != 0 {
		t.Errorf("Cornered needs strictly more: at 1 against 1 nothing reaches the stack, got %v placed (FR-KWD-002e)", got)
	}
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("a gated trigger asks no decision, asked %d (FR-KWD-002e)", n)
	}
}

func TestFR_KWD_002e_CorneredCountsOnlyMainTheaterAssets(t *testing.T) {
	// The opponent has one Main Theater asset, and more in Reserves plus a
	// Headquarters. Only the Main Theater counts, so the counts are 1 and 1.
	s := newScene(t)
	att := s.put(0, ZoneMainTheater, striker("Raider", 1, 9, Keywords{}, cornered(attackTrigger(modifyPower(3, ability.SubjectSelf)))))
	def := s.put(1, ZoneMainTheater, wall("Wall"))
	s.put(1, ZoneReserves, wall("Reserve A"))
	s.put(1, ZoneReserves, wall("Reserve B"))
	s.putScratch(1)
	s.do(s.attack(att, def))
	s.wantDamage(def, 1)
}

func TestFR_KWD_002e_CorneredCountsStructuresAndScratchCompanyUnits(t *testing.T) {
	// The opponent's two Main Theater assets are a structure and a face-down
	// Scratch Company unit. Both must count for 2 against 1.
	s := newScene(t)
	att := s.put(0, ZoneMainTheater, striker("Raider", 1, 9, Keywords{}, cornered(attackTrigger(modifyPower(3, ability.SubjectSelf)))))
	def := s.put(1, ZoneMainTheater, structure("Bunker", 0, 10, Keywords{}))
	scratch := s.putScratch(1)
	s.moveTo(scratch, ZoneMainTheater)
	s.do(s.attack(att, def))
	s.wantDamage(def, 4)
}

// --- FR-LAY-003: checked when the event happens, not when it resolves ---

func TestFR_LAY_003_PlacedCorneredTriggerResolvesAfterTheCountsChange(t *testing.T) {
	// Covers: FR-KWD-002e, FR-TRG-003
	// The attacker has a Cornered +3 trigger and a trigger killing a chosen
	// opposing unit. The opponent has 2 Main Theater assets against 1 when the
	// attack happens (combat hits the Wall). The kill resolves first (the first listed trigger is placed
	// first and so resolves last), leaving 1 against 1; the Cornered trigger
	// was already placed and still resolves.
	s := newScene(t)
	att := s.put(0, ZoneMainTheater, striker("Raider", 1, 9, Keywords{},
		cornered(attackTrigger(modifyPower(3, ability.SubjectSelf))),
		targetedAttack([]ability.Target{tgt(ability.AllegianceOpposing, ability.KindUnit, true)}, dmgTo(99, 0))))
	def := s.put(1, ZoneMainTheater, wall("Wall"))
	victim := s.put(1, ZoneMainTheater, unit("Victim", 0, 1, Keywords{}))
	s.g.choosers[0] = func(_ *Game, d Decision) Answer {
		switch d.Kind {
		case DecideOrder:
			return Answer{Picks: []int{0, 1}}
		case DecideTarget:
			for i, o := range d.Options {
				if o.Ref == victim.ID {
					return Answer{Picks: []int{i}}
				}
			}
		}
		t.Errorf("unexpected decision %q", d.Prompt)
		return Answer{Timeout: true}
	}
	s.do(s.attack(att, def))
	if s.alive(victim) {
		t.Fatalf("precondition: the kill trigger resolves first and removes the victim")
	}
	s.wantDamage(def, 4)
}

func TestFR_KWD_002e_ACorneredFiringTheGateStoppedDoesNotUseUpMaxPerTurn(t *testing.T) {
	// Covers: FR-TRN-001
	// Player 1's unit has a Cornered damage-taken trigger limited to once each
	// turn. The first damage happens at 2 against 2, so the gate stops it.
	// An ally then leaves the Main Theater (2 against 1) and the second damage
	// in the same turn fires it: the stopped firing did not use the allowance.
	s := newScene(t)
	h0 := s.put(0, ZoneMainTheater, unit("Hitter A", 1, 9, Keywords{}))
	h1 := s.put(0, ZoneMainTheater, unit("Hitter B", 1, 9, Keywords{}))
	trig := cornered(limited(1, unbowed(nil, scratchN(1))))
	def := s.put(1, ZoneMainTheater, bowed("Rager", 0, 9, Keywords{}, trig))
	ally := s.put(1, ZoneMainTheater, wall("Ally"))
	s.do(s.attack(h0, def))
	s.wantScratches(1, 0, "not Cornered at 2 against 2, so the gate stops the first firing")
	s.moveTo(ally, ZoneReserves)
	s.do(s.attack(h1, def))
	s.wantDamage(def, 2)
	s.wantScratches(1, 1, "Cornered at 2 against 1: the gate-stopped firing did not use up the once-each-turn allowance (FR-TRN-001)")
}

// --- FR-LAY-001: power until end of turn ---

func TestFR_LAY_001_ModifyStatGivesPowerThatCountsInCombatDamage(t *testing.T) {
	// Covers: FR-CBT-002
	s := newScene(t)
	att := s.put(0, ZoneMainTheater, striker("Raider", 2, 9, Keywords{}, attackTrigger(modifyPower(3, ability.SubjectSelf))))
	def := s.put(1, ZoneMainTheater, wall("Wall"))
	s.do(s.attack(att, def))
	s.wantDamage(def, 5)
	if p := s.g.Power(att); p != 5 {
		t.Errorf("effective Power after +3 until end of turn = %d, want 5 (FR-LAY-001)", p)
	}
}

func TestFR_LAY_003_ModifyStatPowerExpiresAtEndOfTurn(t *testing.T) {
	// Covers: FR-TRN-001
	s := newScene(t)
	att := s.put(0, ZoneMainTheater, striker("Raider", 2, 9, Keywords{}, attackTrigger(modifyPower(3, ability.SubjectSelf))))
	def := s.put(1, ZoneMainTheater, wall("Wall"))
	s.do(s.attack(att, def))
	s.endTurn()
	if p := s.g.Power(att); p != 2 {
		t.Errorf("power until end of turn expires in Sub-Phase 4c: Power = %d, want printed 2 (FR-TRN-001)", p)
	}
}

func TestFR_LAY_001_ModifyStatStacksWithAggressive(t *testing.T) {
	// Covers: FR-KWD-002f
	s := newScene(t)
	s.script(0, Answer{Timeout: true}) // order Aggressive and the attack trigger
	att := s.put(0, ZoneMainTheater, striker("Raider", 2, 9, Keywords{Aggressive: 1}, attackTrigger(modifyPower(3, ability.SubjectSelf))))
	def := s.put(1, ZoneMainTheater, wall("Wall"))
	s.do(s.attack(att, def))
	s.wantDamage(def, 6)
}

func TestFR_TGT_003_ModifyStatOnSelfFizzlesWhenTheSourceHasLeft(t *testing.T) {
	// Covers: FR-LAY-001
	// The attacker has a +3 trigger and a trigger killing itself. The kill
	// resolves first, so the source is gone when +3 resolves: nothing changes.
	s := newScene(t)
	selfKill := ability.Step{Op: ability.OpDealDamage, Amount: ability.Value{N: 99}, Subject: ability.SubjectSelf}
	att := s.put(0, ZoneMainTheater, striker("Raider", 2, 9, Keywords{},
		attackTrigger(modifyPower(3, ability.SubjectSelf)), attackTrigger(selfKill)))
	bystander := s.put(0, ZoneMainTheater, unit("Bystander", 2, 9, Keywords{}))
	def := s.put(1, ZoneMainTheater, wall("Wall"))
	s.script(0, Answer{Picks: []int{0, 1}}) // +3 placed first, so the kill resolves first
	s.do(s.attack(att, def))
	if s.alive(att) {
		t.Fatalf("precondition: the attacker killed itself")
	}
	if p := s.g.Power(bystander); p != 2 {
		t.Errorf("a fizzled +3 on a departed source changes nothing else: bystander Power = %d, want 2 (FR-TGT-003)", p)
	}
	s.wantDamage(def, 0)
}

func TestFR_LAY_001_NegativeModifyStatClampsPowerAtZero(t *testing.T) {
	s := newScene(t)
	att := s.put(0, ZoneMainTheater, striker("Raider", 1, 9, Keywords{},
		targetedAttack([]ability.Target{tgt(ability.AllegianceOpposing, ability.KindUnit, true)}, modifyTarget(-5, 0))))
	def := s.put(1, ZoneMainTheater, unit("Guard", 2, 10, Keywords{}))
	s.g.choosers[0] = func(_ *Game, d Decision) Answer {
		if d.Kind == DecideTarget {
			return Answer{Picks: []int{0}}
		}
		return Answer{Timeout: true}
	}
	s.do(s.attack(att, def))
	if p := s.g.Power(def); p != 0 {
		t.Errorf("2 power with -5 until end of turn: Power = %d, want 0 (Power clamps at 0, FR-LAY-001)", p)
	}
	s.wantDamage(att, 0)
}
