package engine

import (
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for the damage-taken hook behind Unbowed (T-0050): an asset's trigger
// fires when it takes damage, that is when armorAbsorbed + healthCommitted > 0
// at FR-REP-004 Stage 5. Cards are synthetic: no Unbowed card is executable
// yet. Outcomes are observed through face-down Scratch Companies, Damage,
// Armor, zones and the recorded Decisions.
//
// Deliberately not pinned: the order of a dying asset's Unbowed trigger
// relative to death triggers from the same sweep (the FR-KWD-002a "before the
// asset is destroyed" versus FR-TRG-004 placement question is open).

// unbowed is one "Unbowed: <steps>" ability, with chosen targets if any.
func unbowed(targets []ability.Target, steps ...ability.Step) ability.Trigger {
	return ability.Trigger{Hook: ability.HookDamaged, Effect: ability.Effect{Targets: targets, Steps: steps}}
}

// bowed is a unit printing the given Unbowed triggers.
func bowed(name string, power, health int, kw Keywords, trigs ...ability.Trigger) *Card {
	c := unit(name, power, health, kw)
	c.Ability = &ability.Def{Triggers: trigs}
	return c
}

// healSelf is "this unit heals n". armorSelf is "this unit gains n armor".
func healSelf(n int) ability.Step {
	return ability.Step{Op: ability.OpHeal, Amount: ability.Value{N: n}, Subject: ability.SubjectSelf}
}

func armorSelf(n int) ability.Step {
	return ability.Step{Op: ability.OpAddArmor, Amount: ability.Value{N: n}, Subject: ability.SubjectSelf}
}

// putControlled places a Main Theater asset owned by owner but controlled by ctrl.
func (s *scene) putControlled(owner, ctrl PlayerID, c *Card) *Asset {
	a := s.put(owner, ZoneMainTheater, c)
	a.Controller = ctrl
	return a
}

// wantScratches fails unless p has exactly want face-down Scratch Companies.
func (s *scene) wantScratches(p PlayerID, want int, why string) {
	s.t.Helper()
	if got := faceDownCount(s.g.Players[p]); got != want {
		s.t.Errorf("player %d Scratch Companies = %d, want %d: %s", p, got, want, why)
	}
}

// --- FR-KWD-002a / FR-REP-004 Stage 5: when the hook fires ---

func TestFR_KWD_002a_UnbowedFiresWhenHealthDamageIsCommitted(t *testing.T) {
	s := newScene(t)
	att := s.put(0, ZoneMainTheater, unit("Hitter", 2, 9, Keywords{}))
	def := s.put(1, ZoneMainTheater, bowed("Bowed", 1, 5, Keywords{}, unbowed(nil, scratchN(1))))
	s.do(s.attack(att, def))
	s.wantDamage(def, 2)
	s.wantScratches(1, 1, "damage committed to Health fires Unbowed once for the instance, not once per point (FR-KWD-002a)")
	s.wantScratches(0, 0, "the Unbowed trigger belongs to its asset's controller (FR-TRG-001)")
}

// Covers: FR-KWD-002a, FR-KWD-002k
func TestFR_REP_004_UnbowedFiresWhenArmorAbsorbsAllTheDamage(t *testing.T) {
	s := newScene(t)
	att := s.put(0, ZoneMainTheater, unit("Hitter", 2, 9, Keywords{}))
	def := s.put(1, ZoneMainTheater, bowed("Armored", 1, 5, Keywords{Armor: 3}, unbowed(nil, scratchN(1))))
	s.do(s.attack(att, def))
	s.wantDamage(def, 0)
	if def.Armor != 1 {
		t.Fatalf("precondition: Armor absorbs 2 of 3, leaving %d, want 1", def.Armor)
	}
	s.wantScratches(1, 1, "damage absorbed by Armor counts as taken even with 0 committed to Health (FR-REP-004 Stage 5)")
}

// Covers: FR-KWD-002a, FR-KWD-002b
func TestFR_REP_004_UnbowedDoesNotFireWhenResistantReducesDamageToZero(t *testing.T) {
	s := newScene(t)
	att := s.put(0, ZoneMainTheater, unit("Hitter", 2, 9, Keywords{}))
	def := s.put(1, ZoneMainTheater, bowed("Resister", 1, 5, Keywords{Resistant: 2}, unbowed(nil, scratchN(1))))
	s.do(s.attack(att, def))
	s.wantDamage(def, 0)
	s.wantScratches(1, 0, "damage reduced to 0 at Stage 2 means no damage-taken trigger (FR-REP-004 Stage 5, FR-KWD-002a)")
}

// Covers: FR-KWD-002a, FR-KWD-002b
func TestFR_REP_004_UnbowedStillFiresWhenResistantLeavesSomeDamage(t *testing.T) {
	s := newScene(t)
	att := s.put(0, ZoneMainTheater, unit("Hitter", 3, 9, Keywords{}))
	def := s.put(1, ZoneMainTheater, bowed("Resister", 1, 5, Keywords{Resistant: 2}, unbowed(nil, scratchN(1))))
	s.do(s.attack(att, def))
	s.wantDamage(def, 1)
	s.wantScratches(1, 1, "3 damage less Resistant 2 leaves 1, which is taken (FR-REP-004 Stage 5)")
}

// Covers: FR-KWD-002a, FR-KWD-002c
func TestFR_REP_004_UnbowedFiresWhenStalwartCapsDamageToOne(t *testing.T) {
	s := newScene(t)
	att := s.put(0, ZoneMainTheater, unit("Hitter", 5, 9, Keywords{}))
	def := s.put(1, ZoneMainTheater, bowed("Stalwart", 1, 5, Keywords{Stalwart: true}, unbowed(nil, scratchN(1))))
	s.do(s.attack(att, def))
	s.wantDamage(def, 1)
	s.wantScratches(1, 1, "damage capped to 1 is still taken, so Unbowed fires once (Unb x Stw, FR-REP-004 Stage 5)")
}

func TestFR_REP_004_UnbowedDoesNotFireOnAZeroPowerAttack(t *testing.T) {
	s := newScene(t)
	att := s.put(0, ZoneMainTheater, unit("Pacifist", 0, 9, Keywords{}))
	def := s.put(1, ZoneMainTheater, bowed("Bowed", 0, 5, Keywords{}, unbowed(nil, scratchN(1))))
	s.do(s.attack(att, def))
	s.wantDamage(def, 0)
	s.wantScratches(1, 0, "0 raw damage commits and absorbs nothing, so nothing is taken (FR-REP-004 Stage 5)")
}

// Covers: FR-KWD-002a, FR-KWD-002b
func TestFR_REP_004_UnbowedFiresWhenNegativeResistantAmplifiesZeroToOne(t *testing.T) {
	s := newScene(t)
	att := s.put(0, ZoneMainTheater, unit("Pacifist", 0, 9, Keywords{}))
	def := s.put(1, ZoneMainTheater, bowed("Brittle", 0, 5, Keywords{Resistant: -1}, unbowed(nil, scratchN(1))))
	s.do(s.attack(att, def))
	s.wantDamage(def, 1)
	s.wantScratches(1, 1, "Resistant -1 amplifies 0 to 1 at Stage 2, and that 1 is taken (FR-REP-004 Stage 2 and 5)")
}

func TestFR_KWD_002a_OnlyTheAssetThatTookDamageFires(t *testing.T) {
	s := newScene(t)
	att := s.put(0, ZoneMainTheater, unit("Hitter", 2, 9, Keywords{}))
	def := s.put(1, ZoneMainTheater, unit("Plain", 1, 5, Keywords{}))
	s.put(1, ZoneMainTheater, bowed("Bystander", 1, 5, Keywords{}, unbowed(nil, scratchN(1))))
	s.do(s.attack(att, def))
	s.wantScratches(1, 0, "Unbowed fires for damage to its own asset only (FR-KWD-002a)")
}

// --- combat: both combatants, Berserk ---

// Covers: FR-REP-004, FR-CBT-002
func TestFR_KWD_002a_BothCombatantsFireTheirOwnUnbowed(t *testing.T) {
	s := newScene(t)
	att := s.put(0, ZoneMainTheater, bowed("Attacker", 2, 9, Keywords{}, unbowed(nil, scratchN(1))))
	def := s.put(1, ZoneMainTheater, bowed("Defender", 2, 9, Keywords{}, unbowed(nil, scratchN(2))))
	s.do(s.attack(att, def))
	s.wantScratches(0, 1, "the attacker took retaliation damage, so its Unbowed fires (FR-REP-004 Stage 5)")
	s.wantScratches(1, 2, "the defender took combat damage, so its Unbowed fires (FR-REP-004 Stage 5)")
}

// Covers: FR-KWD-002o
func TestFR_KWD_002a_BerserkCleaveDamageFiresTheSecondarysUnbowed(t *testing.T) {
	s := newScene(t)
	att := s.put(0, ZoneMainTheater, unit("Berserker", 2, 9, Keywords{Berserk: 2}))
	def := s.put(1, ZoneMainTheater, unit("Wall", 0, 9, Keywords{}))
	sec := s.put(1, ZoneMainTheater, bowed("Bystander", 0, 5, Keywords{}, unbowed(nil, scratchN(1))))
	s.do(s.cleave(att, def, sec))
	s.wantDamage(sec, 2)
	s.wantScratches(1, 1, "cleave damage that reaches Health fires the secondary's Unbowed (Unb x Bsk, FR-REP-004)")
}

// Covers: FR-KWD-002b, FR-KWD-002o
func TestFR_KWD_002a_BerserkCleaveReducedToZeroDoesNotFireUnbowed(t *testing.T) {
	s := newScene(t)
	att := s.put(0, ZoneMainTheater, unit("Berserker", 2, 9, Keywords{Berserk: 2}))
	def := s.put(1, ZoneMainTheater, unit("Wall", 0, 9, Keywords{}))
	sec := s.put(1, ZoneMainTheater, bowed("Resister", 0, 5, Keywords{Resistant: 2}, unbowed(nil, scratchN(1))))
	s.do(s.cleave(att, def, sec))
	s.wantDamage(sec, 0)
	s.wantScratches(1, 0, "cleave reduced to 0 by Resistant is not taken (Unb x Res, FR-REP-004 Stage 5)")
}

// Covers: FR-KWD-002o
func TestFR_KWD_002a_RetaliationOnABerserkAttackerFiresItsUnbowed(t *testing.T) {
	s := newScene(t)
	att := s.put(0, ZoneMainTheater, bowed("Berserker", 2, 9, Keywords{Berserk: 2}, unbowed(nil, scratchN(1))))
	def := s.put(1, ZoneMainTheater, unit("Wall", 2, 9, Keywords{}))
	sec := s.put(1, ZoneMainTheater, unit("Bystander", 0, 5, Keywords{}))
	s.do(s.cleave(att, def, sec))
	s.wantDamage(att, 2)
	s.wantScratches(0, 1, "retaliation on a Berserk attacker fires the attacker's Unbowed (Unb x Bsk, FR-REP-004)")
}

// --- non-combat damage and separate instances ---

func TestFR_KWD_002a_NonCombatDamageFiresUnbowedOnce(t *testing.T) {
	s := newScene(t)
	def := s.put(1, ZoneMainTheater, bowed("Bowed", 1, 9, Keywords{}, unbowed(nil, scratchN(1))))
	src := trigUnit("Source", targetTrigger([]ability.Target{tgt(ability.AllegianceOpposing, ability.KindUnit, true)}, dmgTo(3, 0)))
	s.deployFromHand(0, src)
	s.wantDamage(def, 3)
	s.wantScratches(1, 1, "damage from an ability fires Unbowed once per instance, not once per point (FR-REP-004 Stage 5)")
}

// Covers: FR-KWD-002b
func TestFR_REP_004_NonCombatDamageReducedToZeroDoesNotFireUnbowed(t *testing.T) {
	s := newScene(t)
	def := s.put(1, ZoneMainTheater, bowed("Resister", 1, 9, Keywords{Resistant: 5}, unbowed(nil, scratchN(1))))
	src := trigUnit("Source", targetTrigger([]ability.Target{tgt(ability.AllegianceOpposing, ability.KindUnit, true)}, dmgTo(3, 0)))
	s.deployFromHand(0, src)
	s.wantDamage(def, 0)
	s.wantScratches(1, 0, "non-combat damage reduced to 0 at Stage 2 is not taken (FR-REP-004 Stage 5)")
}

func TestFR_KWD_002a_TwoSeparateDamageInstancesFireTwice(t *testing.T) {
	s := newScene(t)
	a1 := s.put(0, ZoneMainTheater, unit("Hitter One", 1, 9, Keywords{}))
	a2 := s.put(0, ZoneMainTheater, unit("Hitter Two", 1, 9, Keywords{}))
	def := s.put(1, ZoneMainTheater, bowed("Bowed", 0, 9, Keywords{}, unbowed(nil, scratchN(1))))
	s.do(s.attack(a1, def))
	s.wantScratches(1, 1, "the first damage instance fires Unbowed once (FR-KWD-002a)")
	s.do(s.attack(a2, def))
	s.wantScratches(1, 2, "a second, separate damage instance fires Unbowed again (FR-KWD-002a)")
}

// --- survivor resolution ---

func TestFR_KWD_002a_SurvivorHealsItselfAfterTheDamageIsCommitted(t *testing.T) {
	s := newScene(t)
	att := s.put(0, ZoneMainTheater, unit("Hitter", 3, 9, Keywords{}))
	def := s.put(1, ZoneMainTheater, bowed("Dancer", 0, 5, Keywords{}, unbowed(nil, healSelf(2))))
	s.do(s.attack(att, def))
	if !s.alive(def) {
		t.Fatalf("precondition: the defender survives 3 damage on 5 Health")
	}
	s.wantDamage(def, 1) // 3 taken, then Unbowed: Heal 2
}

func TestFR_KWD_002a_SurvivorGainsArmorFromItsOwnUnbowed(t *testing.T) {
	s := newScene(t)
	att := s.put(0, ZoneMainTheater, unit("Hitter", 3, 9, Keywords{}))
	def := s.put(1, ZoneMainTheater, bowed("Revenant", 0, 5, Keywords{}, unbowed(nil, armorSelf(2))))
	s.do(s.attack(att, def))
	if def.Armor != 2 {
		t.Errorf("a surviving asset's Unbowed: gains 2 armor leaves Armor %d, want 2 (FR-KWD-002a)", def.Armor)
	}
}

// --- lethal damage: the trigger resolves from discard ---

// Covers: FR-ZON-004
func TestFR_KWD_002a_LethalDamageStillFiresAndANonSelfStepResolves(t *testing.T) {
	s := newScene(t)
	att := s.put(0, ZoneMainTheater, unit("Hitter", 3, 9, Keywords{}))
	def := s.put(1, ZoneMainTheater, bowed("Martyr", 1, 1, Keywords{}, unbowed(nil, scratchN(1))))
	s.do(s.attack(att, def))
	if s.alive(def) {
		t.Fatalf("precondition: the defender is destroyed by 3 damage on 1 Health")
	}
	s.wantScratches(1, 1, "Unbowed: Scratch 1 does not touch the dead asset, so it resolves in full (Unb x Scr, FR-KWD-002a)")
}

// Covers: FR-ZON-004, FR-TGT-003
func TestFR_KWD_002a_LethalDamageSelfHealFizzlesButSiblingStepsResolve(t *testing.T) {
	s := newScene(t)
	att := s.put(0, ZoneMainTheater, unit("Hitter", 3, 9, Keywords{}))
	def := s.put(1, ZoneMainTheater, bowed("Dancer", 1, 1, Keywords{}, unbowed(nil, healSelf(2), scratchN(1))))
	inst := def.Inst
	s.do(s.attack(att, def))
	if s.alive(def) {
		t.Fatalf("precondition: the defender is destroyed")
	}
	if !inZone(s.g.Players[1].Discard, inst) {
		t.Errorf("a heal aimed at the destroyed asset fizzles: the card stays in its owner's discard pile (Unb x Heal, FR-TGT-003)")
	}
	s.wantScratches(1, 1, "only the clause aimed at the dead asset fizzles; the next step resolves (FR-KWD-002a)")
	if err := s.g.CheckInvariants(); err != nil {
		t.Errorf("state after the fizzled step: %v", err)
	}
}

// Covers: FR-ZON-004, FR-TGT-003
func TestFR_KWD_002a_LethalDamageSelfArmorFizzlesButSiblingStepsResolve(t *testing.T) {
	s := newScene(t)
	att := s.put(0, ZoneMainTheater, unit("Hitter", 3, 9, Keywords{}))
	def := s.put(1, ZoneMainTheater, bowed("Revenant", 1, 1, Keywords{}, unbowed(nil, armorSelf(2), scratchN(1))))
	inst := def.Inst
	s.do(s.attack(att, def))
	if s.alive(def) {
		t.Fatalf("precondition: the defender is destroyed")
	}
	if !inZone(s.g.Players[1].Discard, inst) {
		t.Errorf("a destroyed asset gains no armor and does not return to the battlefield (Unb x Arm, FR-TGT-003)")
	}
	s.wantScratches(1, 1, "only the armor clause fizzles; the next step resolves (FR-KWD-002a)")
}

// Covers: FR-ZON-004, FR-TGT-001
func TestFR_KWD_002a_LethalDamageChosenTargetStepResolvesFromDiscard(t *testing.T) {
	s := newScene(t)
	att := s.put(0, ZoneMainTheater, unit("Hitter", 3, 9, Keywords{}))
	other := s.put(0, ZoneMainTheater, unit("Other Ally", 1, 5, Keywords{}))
	captain := s.put(1, ZoneMainTheater, bowed("Captain", 1, 1, Keywords{},
		unbowed([]ability.Target{tgt(ability.AllegianceAny, ability.KindUnit|ability.KindStructure, false)}, dmgTo(1, 0))))
	s.scriptRefs(1, nil, other)
	s.script(0)
	s.do(s.attack(att, captain))
	if s.alive(captain) {
		t.Fatalf("precondition: the captain is destroyed")
	}
	s.wantDamage(other, 1)
	asked := s.asked(1)
	if len(asked) == 0 || asked[0].Kind != DecideTarget || asked[0].Player != 1 {
		t.Errorf("the dead asset's controller chooses the target (FR-ZON-004 last-known controller): %+v", asked)
	}
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("the opponent is not asked about the trigger's target: asked %d times", n)
	}
}

// Covers: FR-ZON-004
func TestFR_KWD_002a_LethalDamageTriggerBelongsToTheLastKnownController(t *testing.T) {
	s := newScene(t)
	s.endTurn() // player 1 is active
	att := s.put(1, ZoneMainTheater, unit("Hitter", 3, 9, Keywords{}))
	// Owned by player 1 and controlled by player 0.
	def := s.putControlled(1, 0, bowed("Captured", 1, 1, Keywords{}, unbowed(nil, scratchN(1))))
	s.do(s.attack(att, def))
	if s.alive(def) {
		t.Fatalf("precondition: the captured asset is destroyed")
	}
	s.wantScratches(0, 1, "the trigger resolves for the asset's last-known controller, player 0 (FR-ZON-004)")
	s.wantScratches(1, 0, "the card's owner is not the trigger's controller (FR-ZON-004)")
}

// --- FR-TRG-003 / FR-TRG-004: placement ---

// Covers: FR-REP-004
func TestFR_TRG_003_UnbowedTriggersFollowApnapNonActivePlayerResolvesFirst(t *testing.T) {
	// Both combatants take damage and survive. Both decks are empty and both
	// triggers are Draw 1, so whichever resolves first loses. The non-active
	// player's trigger resolves first, so player 1 loses and 0 wins.
	s := newScene(t)
	s.g.Players[0].Deck = nil
	s.g.Players[1].Deck = nil
	att := s.put(0, ZoneMainTheater, bowed("Attacker", 1, 9, Keywords{}, unbowed(nil, drawN(1))))
	def := s.put(1, ZoneMainTheater, bowed("Defender", 1, 9, Keywords{}, unbowed(nil, drawN(1))))
	s.do(s.attack(att, def))
	if !s.g.Outcome.Over || s.g.Outcome.Draw || s.g.Outcome.Winner != 0 {
		t.Errorf("the non-active player's damage-taken trigger resolves first (FR-TRG-003): player 1 draws from an empty deck and loses, outcome %+v", s.g.Outcome)
	}
}

func TestFR_TRG_004_UnbowedTriggerResolvesAfterTheDamageIsCommitted(t *testing.T) {
	// Heal 5 resolving after Stage 5 removes the 3 committed damage. If the
	// trigger resolved before the damage was committed, the heal would find
	// nothing to remove and 3 damage would remain.
	s := newScene(t)
	att := s.put(0, ZoneMainTheater, unit("Hitter", 3, 9, Keywords{}))
	def := s.put(1, ZoneMainTheater, bowed("Dancer", 0, 5, Keywords{}, unbowed(nil, healSelf(5))))
	s.do(s.attack(att, def))
	s.wantDamage(def, 0)
	if len(s.g.Stack) != 0 {
		t.Errorf("the stack is empty once the trigger resolves: %d items (FR-TRG-004)", len(s.g.Stack))
	}
}

// Covers: FR-TGT-001
func TestFR_TRG_004_UnbowedDamageCascadesIntoAnotherUnbowed(t *testing.T) {
	s := newScene(t)
	att := s.put(0, ZoneMainTheater, unit("Hitter", 2, 9, Keywords{}))
	b := s.put(0, ZoneMainTheater, bowed("Other Bowed", 1, 5, Keywords{}, unbowed(nil, scratchN(1))))
	def := s.put(1, ZoneMainTheater, bowed("Captain", 1, 5, Keywords{},
		unbowed([]ability.Target{tgt(ability.AllegianceOpposing, ability.KindUnit, true)}, dmgTo(1, 0))))
	s.scriptRefs(1, nil, b)
	s.script(0)
	s.do(s.attack(att, def))
	s.wantDamage(b, 1)
	s.wantScratches(0, 1, "damage dealt by an Unbowed trigger fires the damaged asset's own Unbowed (FR-TRG-004 cascade)")
	if len(s.g.Stack) != 0 {
		t.Errorf("the stack is empty once the cascade resolves: %d items", len(s.g.Stack))
	}
}
