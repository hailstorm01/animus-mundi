package engine

import (
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for the deploy-zone gate, Heal and AddArmor (T-0019). Cards are
// synthetic. Targets are chosen by asset with scriptRefs, never by option
// index. Outcomes are observed through Damage, Armor, zones and the recorded
// Decisions.
//
// The task cites FR-TRN-003 at "line 268"; the rule text is in the
// "FR-TRN-003, FR-TRN-003a & FR-MNV-001" block (line 273): a Unit resolving
// from the stack enters its declared battlefield zone.

// healStep is "chosen <slot> heals n".
func healStep(n, slot int) ability.Step {
	return ability.Step{Op: ability.OpHeal, Amount: ability.Value{N: n}, Subject: ability.SubjectTarget, Target: slot}
}

// armorStep is "chosen <slot> gains n armor".
func armorStep(n, slot int) ability.Step {
	return ability.Step{Op: ability.OpAddArmor, Amount: ability.Value{N: n}, Subject: ability.SubjectTarget, Target: slot}
}

// gatedDeploy is "When this unit deploys, if it was deployed to z, <steps>".
func gatedDeploy(z ability.Zone, targets []ability.Target, steps ...ability.Step) ability.Trigger {
	return ability.Trigger{Hook: ability.HookDeploy, ToZone: z, Effect: ability.Effect{Targets: targets, Steps: steps}}
}

// zonedUnit is a free unit that can deploy to Reserves or the Main Theater.
func zonedUnit(name string, trigs ...ability.Trigger) *Card {
	c := trigUnit(name, trigs...)
	c.DeployZones = []Zone{ZoneReserves, ZoneMainTheater}
	return c
}

// deployTo deploys c from p's hand into the declared zone and returns its asset.
func (s *scene) deployTo(p PlayerID, c *Card, z Zone) *Asset {
	s.t.Helper()
	inst := s.give(p, c)
	s.do(Action{Kind: ActDeploy, Player: p, Card: inst.ID, Zone: z})
	return assetOf(s.g, c)
}

func orderAsks(ds []Decision) int {
	n := 0
	for _, d := range ds {
		if d.Kind == DecideOrder {
			n++
		}
	}
	return n
}

var (
	anyAsset  = tgt(ability.AllegianceAny, 0, false)
	alliedU   = tgt(ability.AllegianceAllied, ability.KindUnit, true)
	opposingU = tgt(ability.AllegianceOpposing, ability.KindUnit, true)
)

// --- FR-TRN-003 and the deploy-zone gate ---

func TestFR_TRN_003_GatedDeployTriggerFiresWhenDeployedToMainTheater(t *testing.T) {
	s := newScene(t)
	e1 := s.put(1, ZoneMainTheater, unit("Enemy A", 1, 9, Keywords{}))
	e2 := s.put(1, ZoneMainTheater, unit("Enemy B", 1, 9, Keywords{}))
	s.scriptRefs(0, nil, e2)
	src := s.deployTo(0, zonedUnit("Guerrilla", gatedDeploy(ability.ZoneMainTheater, []ability.Target{opposingU}, dmgTo(2, 0))), ZoneMainTheater)
	if src.Zone != ZoneMainTheater {
		t.Fatalf("a unit enters its declared battlefield zone: zone %s, want Main Theater (FR-TRN-003)", src.Zone)
	}
	if n := len(s.asked(0)); n != 1 {
		t.Errorf("deployed to the Main Theater, the gated trigger fires and asks for its target: asked %d times", n)
	}
	s.wantDamage(e2, 2)
	s.wantDamage(e1, 0)
}

func TestFR_TRN_003_GatedDeployTriggerFastPathsASingleTargetInMainTheater(t *testing.T) {
	s := newScene(t)
	e := s.put(1, ZoneMainTheater, unit("Enemy", 1, 9, Keywords{}))
	s.script(0)
	s.deployTo(0, zonedUnit("Guerrilla", gatedDeploy(ability.ZoneMainTheater, []ability.Target{opposingU}, dmgTo(2, 0))), ZoneMainTheater)
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("a single legal target is chosen without a question (FR-MOD-001): asked %d times", n)
	}
	s.wantDamage(e, 2)
}

func TestFR_TRN_003_GatedDeployTriggerDoesNothingWhenDeployedToReserves(t *testing.T) {
	s := newScene(t)
	e1 := s.put(1, ZoneMainTheater, unit("Enemy A", 1, 9, Keywords{}))
	e2 := s.put(1, ZoneMainTheater, unit("Enemy B", 1, 9, Keywords{}))
	s.script(0)
	src := s.deployTo(0, zonedUnit("Guerrilla", gatedDeploy(ability.ZoneMainTheater, []ability.Target{opposingU}, dmgTo(2, 0))), ZoneReserves)
	if src.Zone != ZoneReserves {
		t.Fatalf("a unit enters its declared battlefield zone: zone %s, want Reserves (FR-TRN-003)", src.Zone)
	}
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("a trigger whose zone condition fails never goes on the stack, so no target is asked: asked %d times", n)
	}
	s.wantDamage(e1, 0)
	s.wantDamage(e2, 0)
	if len(s.g.Stack) != 0 {
		t.Errorf("the stack is empty after the deploy: %d items", len(s.g.Stack))
	}
}

func TestFR_TRN_003_UngatedDeployTriggerFiresInEitherZone(t *testing.T) {
	for _, z := range []Zone{ZoneMainTheater, ZoneReserves} {
		t.Run(z.String(), func(t *testing.T) {
			s := newScene(t)
			s.put(1, ZoneMainTheater, unit("Enemy A", 1, 9, Keywords{}))
			e2 := s.put(1, ZoneMainTheater, unit("Enemy B", 1, 9, Keywords{}))
			s.scriptRefs(0, nil, e2)
			s.deployTo(0, zonedUnit("Twin", gatedDeploy(ability.ZoneAny, []ability.Target{opposingU}, dmgTo(2, 0))), z)
			s.wantDamage(e2, 2)
			if n := len(s.asked(0)); n != 1 {
				t.Errorf("an ungated deploy trigger fires when deployed to %s: asked %d times", z, n)
			}
		})
	}
}

func TestFR_TRN_003_ReservesGateFiresInReservesAndNotInMainTheater(t *testing.T) {
	cases := []struct {
		zone Zone
		fire bool
	}{{ZoneReserves, true}, {ZoneMainTheater, false}}
	for _, tc := range cases {
		t.Run(tc.zone.String(), func(t *testing.T) {
			s := newScene(t)
			s.put(1, ZoneMainTheater, unit("Enemy A", 1, 9, Keywords{}))
			e2 := s.put(1, ZoneMainTheater, unit("Enemy B", 1, 9, Keywords{}))
			s.scriptRefs(0, nil, e2)
			s.deployTo(0, zonedUnit("Mirror", gatedDeploy(ability.ZoneReserves, []ability.Target{opposingU}, dmgTo(2, 0))), tc.zone)
			wantDmg, wantAsks := 0, 0
			if tc.fire {
				wantDmg, wantAsks = 2, 1
			}
			s.wantDamage(e2, wantDmg)
			if got := len(s.asked(0)); got != wantAsks {
				t.Errorf("a Reserves-gated trigger deployed to %s: fires=%v, asked %d times", tc.zone, tc.fire, got)
			}
		})
	}
}

func TestFR_TRN_003_FailedGateTriggerIsNotOrderedWithItsSibling(t *testing.T) {
	// A gated-off trigger is never collected, so the one remaining trigger
	// needs no ordering question (FR-TRG-003).
	s := newScene(t)
	cards := s.onTop(0, vanilla("A", "B")...)
	e := s.put(1, ZoneMainTheater, unit("Enemy", 1, 9, Keywords{}))
	s.script(0)
	s.deployTo(0, zonedUnit("Guerrilla",
		gatedDeploy(ability.ZoneMainTheater, []ability.Target{opposingU}, dmgTo(2, 0)),
		deployTrigger(drawN(1))), ZoneReserves)
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("only the ungated trigger fires, so there is nothing to order or target: asked %d times (FR-TRG-003)", n)
	}
	if !inHand(s.g.Players[0], cards[0]) {
		t.Errorf("the ungated Draw trigger still resolves")
	}
	s.wantDamage(e, 0)
}

func TestFR_TRN_003_PassedGateTriggerIsOrderedWithItsSibling(t *testing.T) {
	s := newScene(t)
	cards := s.onTop(0, vanilla("A", "B")...)
	s.put(1, ZoneMainTheater, unit("Enemy A", 1, 9, Keywords{}))
	e2 := s.put(1, ZoneMainTheater, unit("Enemy B", 1, 9, Keywords{}))
	s.scriptRefs(0, []int{0, 1}, e2)
	s.deployTo(0, zonedUnit("Guerrilla",
		gatedDeploy(ability.ZoneMainTheater, []ability.Target{opposingU}, dmgTo(2, 0)),
		deployTrigger(drawN(1))), ZoneMainTheater)
	if n := orderAsks(s.asked(0)); n != 1 {
		t.Errorf("both triggers fire, so the controller orders them once: %d order questions (FR-TRG-003)", n)
	}
	s.wantDamage(e2, 2)
	if !inHand(s.g.Players[0], cards[0]) {
		t.Errorf("the Draw trigger resolves too")
	}
}

// --- FR-KWD-002d: Heal ---

func TestFR_KWD_002d_HealRemovesUpToNDamage(t *testing.T) {
	s := newScene(t)
	ally := s.put(0, ZoneMainTheater, unit("Ally", 1, 9, Keywords{}))
	ally.Damage = 5
	s.scriptRefs(0, nil, ally)
	s.deployTo(0, zonedUnit("Healer", targetTrigger([]ability.Target{alliedU}, healStep(2, 0))), ZoneReserves)
	s.wantDamage(ally, 3)
}

func TestFR_KWD_002d_HealNeverGoesBelowZeroDamage(t *testing.T) {
	s := newScene(t)
	ally := s.put(0, ZoneMainTheater, unit("Ally", 1, 9, Keywords{}))
	ally.Damage = 1
	s.scriptRefs(0, nil, ally)
	s.deployTo(0, zonedUnit("Healer", targetTrigger([]ability.Target{alliedU}, healStep(2, 0))), ZoneReserves)
	s.wantDamage(ally, 0)
	if h := s.g.CurrentHealth(ally); h != 9 {
		t.Errorf("healing cannot raise Current Health above Effective Maximum Health: health %d, want 9 (FR-KWD-002d)", h)
	}
}

func TestFR_KWD_002d_HealingAnUndamagedAssetChangesNothing(t *testing.T) {
	s := newScene(t)
	ally := s.put(0, ZoneMainTheater, unit("Ally", 1, 9, Keywords{Armor: 1}))
	s.scriptRefs(0, nil, ally)
	s.deployTo(0, zonedUnit("Healer", targetTrigger([]ability.Target{alliedU}, healStep(3, 0))), ZoneReserves)
	s.wantDamage(ally, 0)
	if ally.Armor != 1 {
		t.Errorf("Heal does nothing to an undamaged asset and never touches Armor: Armor %d, want 1 (FR-KWD-002d)", ally.Armor)
	}
}

func TestFR_KWD_002d_HealLeavesArmorAlone(t *testing.T) {
	s := newScene(t)
	ally := s.put(0, ZoneMainTheater, unit("Ally", 1, 9, Keywords{Armor: 2}))
	ally.Damage = 3
	s.scriptRefs(0, nil, ally)
	s.deployTo(0, zonedUnit("Healer", targetTrigger([]ability.Target{alliedU}, healStep(2, 0))), ZoneReserves)
	s.wantDamage(ally, 1)
	if ally.Armor != 2 {
		t.Errorf("Heal removes damage, not Armor: Armor %d, want 2 (FR-KWD-002d)", ally.Armor)
	}
}

func TestFR_KWD_002d_AnyKindOfAssetCanBeHealed(t *testing.T) {
	names := []string{"allied unit", "allied structure", "allied Scratch Company", "opposing unit",
		"opposing structure", "opposing Scratch Company", "allied Headquarters", "opposing Headquarters"}
	for i, name := range names {
		t.Run(name, func(t *testing.T) {
			s := newScene(t)
			b := s.targetBoard()
			victims := []*Asset{b.u0, b.s0, b.c0, b.u1, b.s1, b.c1, s.g.Players[0].HQ, s.g.Players[1].HQ}
			v := victims[i]
			// A face-down Scratch Company is 1/1 (FR-ZON-005), so no damage on it
			// survives a sweep. For those rows the check is that it is an option,
			// is chosen and resolves cleanly, and stays undamaged.
			scratch := v.FaceDown
			if !scratch {
				v.Damage = 4
			}
			s.scriptRefs(0, nil, v)
			s.deployTo(0, zonedUnit("Healer", targetTrigger([]ability.Target{anyAsset}, healStep(3, 0))), ZoneReserves)
			if scratch {
				if len(s.asked(0)) != 1 {
					t.Fatalf("the any-asset heal asks one question: asked %d times", len(s.asked(0)))
				}
				found := false
				for _, r := range refs(s.asked(0)[0]) {
					found = found || r == v.ID
				}
				if !found {
					t.Errorf("a Scratch Company is among the options of an any-asset heal (FR-KWD-002d)")
				}
				if !s.alive(v) {
					t.Errorf("healing an undamaged Scratch Company leaves it on the battlefield")
				}
				s.wantDamage(v, 0)
				if got := totalDamage(s.g); got != 0 {
					t.Errorf("nothing is damaged: total damage %d, want 0", got)
				}
				return
			}
			s.wantDamage(v, 1)
			if got := totalDamage(s.g); got != 1 {
				t.Errorf("only the chosen asset is healed: total damage %d, want 1 (FR-KWD-002d)", got)
			}
		})
	}
}

func TestFR_KWD_002d_OpposingWardUnitCannotBeChosenToHeal(t *testing.T) {
	s := newScene(t)
	ward := s.put(1, ZoneMainTheater, unit("Warded", 1, 9, Keywords{Ward: true}))
	ward.Damage = 4
	s.script(0)
	src := s.deployTo(0, zonedUnit("Healer", targetTrigger([]ability.Target{opposingU}, healStep(3, 0))), ZoneReserves)
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("an opposing Ward unit is not a legal target, so nothing is asked: asked %d times (FR-KWD-002h)", n)
	}
	s.wantDamage(ward, 4)
	if !s.alive(src) {
		t.Errorf("the healer still deploys when its trigger has no legal target (FR-TGT-003)")
	}
}

func TestFR_KWD_002d_HealWithNoLegalTargetDoesNothing(t *testing.T) {
	s := newScene(t)
	cards := s.onTop(0, vanilla("A")...)
	s.script(0)
	s.deployTo(0, zonedUnit("Healer", targetTrigger([]ability.Target{opposingU}, healStep(2, 0), drawN(1))), ZoneReserves)
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("no legal target means no question: asked %d times", n)
	}
	if !inHand(s.g.Players[0], cards[0]) {
		t.Errorf("the other clauses of the effect still resolve (FR-TGT-003)")
	}
}

// --- FR-LAY-002 / FR-KWD-002k: the Armor pool ---

func TestFR_LAY_002_ArmorGrantAddsToThePool(t *testing.T) {
	s := newScene(t)
	ally := s.put(0, ZoneMainTheater, unit("Ally", 1, 9, Keywords{}))
	s.scriptRefs(0, nil, ally)
	s.deployTo(0, zonedUnit("Plater", targetTrigger([]ability.Target{alliedU}, armorStep(2, 0))), ZoneReserves)
	if ally.Armor != 2 {
		t.Errorf("\"gains 2 armor\" puts 2 in the Armor pool: Armor %d, want 2 (FR-LAY-002)", ally.Armor)
	}
	s.wantDamage(ally, 0)
}

func TestFR_LAY_002_SeparateGrantsStackAdditively(t *testing.T) {
	s := newScene(t)
	ally := s.put(0, ZoneMainTheater, unit("Ally", 1, 9, Keywords{}))
	s.scriptRefs(0, []int{0, 1}, ally, ally)
	armor := func(n int) ability.Trigger { return targetTrigger([]ability.Target{alliedU}, armorStep(n, 0)) }
	s.deployTo(0, zonedUnit("Plater", armor(2), armor(3)), ZoneReserves)
	if ally.Armor != 5 {
		t.Errorf("two discrete grants of 2 and 3 stack to 5 (FR-LAY-002): Armor %d", ally.Armor)
	}
}

func TestFR_KWD_002k_GrantedArmorJoinsPrintedArmorInOnePool(t *testing.T) {
	s := newScene(t)
	ally := s.put(0, ZoneMainTheater, unit("Ally", 1, 9, Keywords{Armor: 1}))
	s.scriptRefs(0, nil, ally)
	s.deployTo(0, zonedUnit("Plater", targetTrigger([]ability.Target{alliedU}, armorStep(2, 0))), ZoneReserves)
	if ally.Armor != 3 {
		t.Errorf("printed Armor 1 and a grant of 2 share one pool of 3 (FR-KWD-002k): Armor %d", ally.Armor)
	}
}

func TestFR_LAY_002_GrantedArmorDoesNotExpireAtEndOfTurn(t *testing.T) {
	s := newScene(t)
	ally := s.put(0, ZoneMainTheater, unit("Ally", 1, 9, Keywords{}))
	s.scriptRefs(0, nil, ally)
	s.deployTo(0, zonedUnit("Plater", targetTrigger([]ability.Target{alliedU}, armorStep(2, 0))), ZoneReserves)
	s.endTurn()
	if ally.Armor != 2 {
		t.Errorf("Armor never expires at the end of the turn: Armor %d after one end of turn, want 2 (FR-LAY-002)", ally.Armor)
	}
	s.endTurn()
	if ally.Armor != 2 {
		t.Errorf("Armor still holds after a full round: Armor %d, want 2 (FR-LAY-002)", ally.Armor)
	}
}

func TestFR_LAY_002_GrantedArmorAbsorbsLaterDamageOneForOneBeforeHealth(t *testing.T) {
	s := newScene(t)
	ally := s.put(0, ZoneMainTheater, unit("Ally", 1, 9, Keywords{}))
	s.scriptRefs(0, nil, ally)
	s.deployTo(0, zonedUnit("Plater", targetTrigger([]ability.Target{alliedU}, armorStep(2, 0))), ZoneReserves)
	s.endTurn()
	hitter := s.put(1, ZoneMainTheater, unit("Hitter", 3, 9, Keywords{}))
	s.do(s.attack(hitter, ally))
	s.wantDamage(ally, 1)
	if ally.Armor != 0 {
		t.Errorf("3 combat damage: 2 absorbed 1-for-1 by Armor (consumed), 1 reaches health: Armor %d, want 0 (FR-REP-004 Stage 4)", ally.Armor)
	}
}

func TestFR_ZON_004_MarchKeepsGrantedArmor(t *testing.T) {
	s := newScene(t)
	walker := s.put(0, ZoneReserves, unit("Walker", 2, 9, Keywords{}))
	s.scriptRefs(0, nil, walker)
	s.deployTo(0, zonedUnit("Plater", targetTrigger([]ability.Target{alliedU}, armorStep(2, 0))), ZoneReserves)
	if walker.Armor != 2 {
		t.Fatalf("precondition: the grant landed: Armor %d, want 2", walker.Armor)
	}
	s.do(Action{Kind: ActMarch, Player: 0, Asset: walker.ID})
	if walker.Zone != ZoneMainTheater {
		t.Fatalf("the walker marched to the Main Theater: zone %s", walker.Zone)
	}
	if walker.Armor != 2 {
		t.Errorf("moving between Reserves and the Main Theater keeps dynamic Armor: Armor %d, want 2 (FR-ZON-004)", walker.Armor)
	}
}

// --- deploys or attacks (FR-CBT-002 Step 2) ---

// healBoth is a unit healing a chosen allied unit 2 on deploy and on attack.
func healBoth(health int) *Card {
	heal := []ability.Target{alliedU}
	c := zonedUnit("Colossus", targetTrigger(heal, healStep(2, 0)))
	c.Ability.Triggers = append(c.Ability.Triggers, ability.Trigger{
		Hook: ability.HookAttacks, Effect: ability.Effect{Targets: heal, Steps: []ability.Step{healStep(2, 0)}}})
	c.Health = health
	return c
}

func TestFR_CBT_002_DeployOrAttackHealsOnceOnDeployAndOnceEachAttack(t *testing.T) {
	s := newScene(t)
	ally := s.put(0, ZoneMainTheater, unit("Ally", 1, 10, Keywords{}))
	ally.Damage = 7
	dummy := s.put(1, ZoneMainTheater, unit("Dummy", 0, 20, Keywords{}))
	s.scriptRefs(0, nil, ally, ally, ally)
	src := s.deployTo(0, healBoth(9), ZoneMainTheater)
	s.wantDamage(ally, 5)
	s.do(s.attack(src, dummy))
	s.wantDamage(ally, 3)
	s.endTurn()
	s.endTurn()
	s.do(s.attack(src, dummy))
	s.wantDamage(ally, 1)
	if n := len(s.asked(0)); n != 3 {
		t.Errorf("one target question for the deploy and one per attack: asked %d times, want 3", n)
	}
}

func TestFR_CBT_002_AttackHealResolvesBeforeCombatDamage(t *testing.T) {
	// Without the heal the attacker would die: 4 damage + 1 retaliation = 5.
	// Healing 2 at Step 2 leaves it at 3 after combat.
	s := newScene(t)
	att := s.put(0, ZoneMainTheater, healBoth(5))
	att.Damage = 4
	def := s.put(1, ZoneMainTheater, unit("Defender", 1, 20, Keywords{}))
	s.scriptRefs(0, nil, att)
	s.do(s.attack(att, def))
	if !s.alive(att) {
		t.Fatalf("the Step 2 heal resolves before combat damage, so the attacker survives 1 retaliation (FR-CBT-002)")
	}
	s.wantDamage(att, 3)
}

func TestFR_CBT_002_AttackHealOfTheAttackerDoesNotSaveItFromLethalRetaliation(t *testing.T) {
	s := newScene(t)
	att := s.put(0, ZoneMainTheater, healBoth(5))
	att.Damage = 3
	inst := att.Inst
	def := s.put(1, ZoneMainTheater, unit("Defender", 4, 20, Keywords{}))
	s.scriptRefs(0, nil, att)
	s.do(s.attack(att, def))
	if s.alive(att) || !inZone(s.g.Players[0].Discard, inst) {
		t.Errorf("heal leaves 1 damage, and 4 retaliation is still lethal on 5 health: the attacker is destroyed (FR-CBT-002)")
	}
}

func TestFR_CBT_002_AttackHealOfADamagedAllyIsVisibleAfterCombat(t *testing.T) {
	s := newScene(t)
	att := s.put(0, ZoneMainTheater, healBoth(9))
	ally := s.put(0, ZoneMainTheater, unit("Ally", 1, 10, Keywords{}))
	ally.Damage = 4
	def := s.put(1, ZoneMainTheater, unit("Defender", 1, 20, Keywords{}))
	s.scriptRefs(0, nil, ally)
	s.do(s.attack(att, def))
	s.wantDamage(ally, 2)
	s.wantDamage(def, 1)
}
