package engine

import (
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for the combat-kill hook (T-0080): a printed trigger that fires when its
// own asset, as the declared attacker, deals combat damage that leaves the
// defender lethal, so the death sweep after the Attack destroys it (FR-CBT-002
// Step 3, "Post-Attack & Combat-Kill Triggers"). Retaliation kills do not count
// (is_attacker false), a Berserk secondary killed by cleave is not a combat kill
// (FR-KWD-002o: cleave is non-combat effect damage), and a fizzled Attack deals
// no combat damage (FR-CBT-002 Step 3, "Combatant Target Check"). Like every
// trigger it is placed after the sweep, in APNAP order with the defender's death
// triggers (FR-TRG-003, FR-TRG-004).
//
// Cards are synthetic: a combat-kill trigger that Scratches 1, observed with
// wantScratches. Eternal Watchman is a later card task.
//
// This file does not compile until ability has a HookCombatKill constant; that
// is expected until T-0080 is implemented.

// combatKillTrigger is one "Whenever this attacks and kills, <steps>" ability.
func combatKillTrigger(steps ...ability.Step) ability.Trigger {
	return ability.Trigger{Hook: ability.HookCombatKill, Effect: ability.Effect{Steps: steps}}
}

func TestFR_CBT_002_CombatKillTriggerFiresWhenTheAttackerKillsTheDefender(t *testing.T) {
	s := newScene(t)
	att := s.put(0, ZoneMainTheater, striker("Slayer", 3, 9, Keywords{}, combatKillTrigger(scratchN(1))))
	def := s.put(1, ZoneMainTheater, unit("Victim", 0, 3, Keywords{}))
	s.do(s.attack(att, def))
	if s.alive(def) {
		t.Fatalf("precondition: the defender is destroyed by the combat damage")
	}
	s.wantScratches(0, 1, "the attacker's combat damage destroyed the defender, so its combat-kill trigger fires once and resolves (FR-CBT-002 Step 3)")
	s.wantScratches(1, 0, "the trigger belongs to the attacker's controller (FR-TRG-001)")
}

func TestFR_CBT_002_CombatKillTriggerFiresWhenTheDefenderIsAStructure(t *testing.T) {
	s := newScene(t)
	att := s.put(0, ZoneMainTheater, striker("Wrecker", 3, 9, Keywords{}, combatKillTrigger(scratchN(1))))
	def := s.put(1, ZoneMainTheater, structure("Bunker", 0, 3, Keywords{}))
	s.do(s.attack(att, def))
	if s.alive(def) {
		t.Fatalf("precondition: the structure is destroyed by the combat damage")
	}
	s.wantScratches(0, 1, "destroying a structure defender is a combat kill (FR-CBT-002 Step 3)")
}

func TestFR_CBT_002_CombatKillTriggerFiresFromAHeadquartersAttacker(t *testing.T) {
	s := newScene(t)
	p := s.g.Players[0]
	p.HQ.Inst.Card.Ability = &ability.Def{Triggers: []ability.Trigger{combatKillTrigger(scratchN(1))}}
	def := s.put(1, ZoneMainTheater, unit("Victim", 0, 3, Keywords{}))
	s.do(s.attack(p.HQ, def))
	if s.alive(def) {
		t.Fatalf("precondition: the Headquarters' combat damage (3) destroys the unit")
	}
	s.wantScratches(0, 1, "a Headquarters that attacks and destroys its defender fires its combat-kill trigger (FR-CBT-002 Step 3)")
}

func TestFR_CBT_002_CombatKillTriggerDoesNotFireWhenTheDefenderSurvives(t *testing.T) {
	s := newScene(t)
	att := s.put(0, ZoneMainTheater, striker("Slayer", 2, 9, Keywords{}, combatKillTrigger(scratchN(1))))
	def := s.put(1, ZoneMainTheater, unit("Wall", 0, 5, Keywords{}))
	s.do(s.attack(att, def))
	if !s.alive(def) {
		t.Fatalf("precondition: the defender survives the combat damage")
	}
	s.wantScratches(0, 0, "the defender was not destroyed, so there is no combat kill (FR-CBT-002 Step 3)")
}

func TestFR_CBT_002_CombatKillTriggerDoesNotFireForTheDefender(t *testing.T) {
	// The defender's retaliation destroys the attacker, and the defender survives.
	s := newScene(t)
	att := s.put(0, ZoneMainTheater, unit("Rash Attacker", 1, 1, Keywords{}))
	def := s.put(1, ZoneMainTheater, striker("Avenger", 5, 9, Keywords{}, combatKillTrigger(scratchN(1))))
	s.do(s.attack(att, def))
	if s.alive(att) || !s.alive(def) {
		t.Fatalf("precondition: retaliation destroys the attacker and the defender survives (attacker alive=%v, defender alive=%v)", s.alive(att), s.alive(def))
	}
	s.wantScratches(1, 0, "a retaliation kill is not a combat kill, because the defender did not deal the damage as the declared attacker (FR-CBT-002 Step 3)")
	s.wantScratches(0, 0, "the attacker has no combat-kill trigger")
}

func TestFR_CBT_002_CombatKillTriggerDoesNotFireForABerserkSecondary(t *testing.T) {
	// Covers: FR-KWD-002o
	s := newScene(t)
	att := s.put(0, ZoneMainTheater, striker("Berserker", 1, 9, Keywords{Berserk: 2}, combatKillTrigger(scratchN(1))))
	def := s.put(1, ZoneMainTheater, unit("Wall", 0, 9, Keywords{}))
	sec := s.put(1, ZoneMainTheater, unit("Bystander", 0, 2, Keywords{}))
	s.do(s.cleave(att, def, sec))
	if s.alive(sec) || !s.alive(def) {
		t.Fatalf("precondition: cleave destroys the secondary and the defender survives (secondary alive=%v, defender alive=%v)", s.alive(sec), s.alive(def))
	}
	s.wantScratches(0, 0, "Berserk cleave is non-combat effect damage to a secondary, so killing it is not a combat kill (FR-KWD-002o, FR-CBT-002 Step 3)")
}

func TestFR_CBT_002_CombatKillTriggerDoesNotFireWhenTheAttackFizzles(t *testing.T) {
	// The attacker's attack trigger destroys the defender at Step 2. The Attack
	// then fizzles and deals no combat damage.
	s := newScene(t)
	opposing := allDamage(99, ability.Filter{Allegiance: ability.AllegianceOpposing, Kinds: ability.KindUnit})
	att := s.put(0, ZoneMainTheater, striker("Sniper", 3, 9, Keywords{}, attackTrigger(opposing), combatKillTrigger(scratchN(1))))
	def := s.put(1, ZoneMainTheater, unit("Victim", 0, 3, Keywords{}))
	s.do(s.attack(att, def))
	if s.alive(def) {
		t.Fatalf("precondition: the defender is destroyed by the attack trigger before the Attack resolves")
	}
	s.wantScratches(0, 0, "a fizzled Attack deals no combat damage, so it kills nothing in combat (FR-CBT-002 Step 3, Combatant Target Check)")
}

func TestFR_TRG_003_CombatKillTriggerAndTheDefendersDeathTriggerAreAPNAP(t *testing.T) {
	// Both decks are empty and both triggers are Draw 1, so whichever resolves
	// first loses at its own sweep. The active player's combat-kill trigger is
	// placed first, the non-active player's death trigger above it, so player 1
	// resolves first and loses.
	s := newScene(t)
	s.g.Players[0].Deck = nil
	s.g.Players[1].Deck = nil
	att := s.put(0, ZoneMainTheater, striker("Slayer", 3, 9, Keywords{}, combatKillTrigger(drawN(1))))
	def := s.put(1, ZoneMainTheater, mortal("Victim", 0, 3, killedTrigger(drawN(1))))
	s.do(s.attack(att, def))
	if s.alive(def) {
		t.Fatalf("precondition: the defender is destroyed by the combat damage")
	}
	wantPlaced(t, s, "the active player's combat-kill trigger is placed before the non-active player's death trigger (APNAP)",
		[]*Asset{att, def}, att, def)
	if o := s.g.Outcome; !o.Over || o.Draw || o.Winner != 0 {
		t.Errorf("the non-active player's death trigger resolves first (FR-TRG-003): player 1 draws from an empty deck and loses, outcome %+v", o)
	}
}
