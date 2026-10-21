package engine

import (
	"strings"
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0028: damage equal to the source's power (SVY-R10), last-known
// effective state (FR-ZON-004, board directive #13) and steps on "all assets"
// matching a filter (FR-TGT-002). Cards are synthetic. SVY-R10 is ruled but not
// yet in FRD 01, so tests that pin only it carry a `// Covers: SVY-R10` comment
// and a TestFR_ name for the closest FRD 01 requirement.
//
// Trigger option order for a Step 2 window is: Aggressive, then the printed
// triggers in printed order (FR-TRG-003, SVY-R7). The first pick of an Order
// answer is placed first and resolves last.

// powerStep is "deal damage equal to its power to chosen <slot>".
func powerStep(slot int) ability.Step {
	return ability.Step{Op: ability.OpDealDamage, Amount: ability.Value{From: ability.ValueSourcePower},
		Subject: ability.SubjectTarget, Target: slot}
}

// allStep is "<op> n to all assets matching f".
func allStep(op ability.Op, n int, f ability.Filter) ability.Step {
	return ability.Step{Op: op, Amount: ability.Value{N: n}, Subject: ability.SubjectAll, Filter: f}
}

func allDamage(n int, f ability.Filter) ability.Step { return allStep(ability.OpDealDamage, n, f) }

// scriptAnyOf installs a chooser for seat p that answers every DecideOrder with
// order and every DecideTarget by picking the first option referring to any of
// want. It records every Decision like scene.script, and reports a target
// decision none of whose options is wanted.
func (s *scene) scriptAnyOf(p PlayerID, order []int, want ...*Asset) {
	if s.asks == nil {
		s.asks = map[PlayerID][]Decision{}
	}
	s.asks[p] = nil
	s.g.choosers[p] = func(_ *Game, d Decision) Answer {
		s.asks[p] = append(s.asks[p], d)
		if d.Kind == DecideOrder {
			return Answer{Picks: order}
		}
		for i, o := range d.Options {
			for _, w := range want {
				if o.Ref == w.ID {
					return Answer{Picks: []int{i}}
				}
			}
		}
		s.t.Errorf("none of the wanted assets is among the options of %q", d.Prompt)
		return Answer{Timeout: true}
	}
}

// noFormatErrors fails if any event Detail shows a fmt failure such as "%!s(<nil>)".
func noFormatErrors(t *testing.T, g *Game) {
	t.Helper()
	for _, e := range g.Log {
		if strings.Contains(e.Detail, "%!") {
			t.Errorf("event %d (%s) Detail contains a formatting failure: %q", e.Seq, e.Kind, e.Detail)
		}
	}
}

// nonCombatDamage returns the non-combat damage events of the log.
func nonCombatDamage(log []Event) []Event {
	var out []Event
	for _, e := range damageEvents(log) {
		if !e.Combat {
			out = append(out, e)
		}
	}
	return out
}

func damageTo(log []Event, target *Asset) []Event {
	var out []Event
	for _, e := range nonCombatDamage(log) {
		if e.Target == target.ID {
			out = append(out, e)
		}
	}
	return out
}

// --- SVY-R10: the power is read when the step resolves ---

// raiderScene has Raider (power 3, Aggressive 2) attack Dummy with the given
// printed attack triggers. Victim is a second opposing unit.
func raiderScene(t *testing.T, health int, trigs ...ability.Trigger) (s *scene, raider, dummy, victim *Asset) {
	t.Helper()
	s = newScene(t)
	raider = s.put(0, ZoneMainTheater, striker("Raider", 3, health, Keywords{Aggressive: 2}, trigs...))
	dummy = s.put(1, ZoneMainTheater, unit("Dummy", 0, 20, Keywords{}))
	victim = s.put(1, ZoneMainTheater, unit("Victim", 0, 20, Keywords{}))
	return s, raider, dummy, victim
}

func TestFR_TRG_004_SourcePowerIsReadWhenTheStepResolves(t *testing.T) {
	// Covers: SVY-R10
	// Options: 0 = Aggressive, 1 = the power trigger. Picks [1 0] place Aggressive
	// last, so it resolves first and the power is 5; picks [0 1] resolve the power
	// trigger first, while the power is still 3.
	for _, tc := range []struct {
		name  string
		picks []int
		want  int
	}{{"Aggressive resolves first", []int{1, 0}, 5}, {"power trigger resolves first", []int{0, 1}, 3}} {
		trig := targetTrigger([]ability.Target{tgt(ability.AllegianceOpposing, ability.KindUnit, false)}, powerStep(0))
		trig.Hook = ability.HookAttacks
		s, raider, dummy, victim := raiderScene(t, 9, trig)
		s.scriptAnyOf(0, tc.picks, victim)
		s.scriptAnyOf(1, nil)
		s.do(s.attack(raider, dummy))
		if victim.Damage != tc.want {
			t.Errorf("%s: damage equal to its power = %d, want %d (power is read when the step resolves, SVY-R10)", tc.name, victim.Damage, tc.want)
		}
		noFormatErrors(t, s.g)
	}
}

func TestFR_ZON_004_SourceGoneBeforeResolutionDealsItsLastKnownEffectivePower(t *testing.T) {
	// Covers: SVY-R10
	// Raider (3, Aggressive 2, health 4) has two attack triggers: T1 deals 9 to a
	// chosen allied unit (itself), T2 deals damage equal to its power to a chosen
	// opposing unit. Options: 0 = Aggressive, 1 = T1, 2 = T2. Picks [2 1 0] resolve
	// Aggressive, then T1 (the sweep destroys Raider), then T2.
	t1 := targetTrigger([]ability.Target{tgt(ability.AllegianceAllied, ability.KindUnit, false)}, dmgTo(9, 0))
	t1.Hook = ability.HookAttacks
	t2 := targetTrigger([]ability.Target{tgt(ability.AllegianceOpposing, ability.KindUnit, false)}, powerStep(0))
	t2.Hook = ability.HookAttacks
	s, raider, dummy, victim := raiderScene(t, 4, t1, t2)
	s.scriptAnyOf(0, []int{2, 1, 0}, raider, victim)
	s.scriptAnyOf(1, nil)
	s.do(s.attack(raider, dummy))
	if s.alive(raider) {
		t.Fatalf("precondition: T1's 9 damage destroys Raider before T2 resolves")
	}
	s.wantDamage(victim, 5)
	ev := damageTo(s.g.Log, victim)
	if len(ev) != 1 {
		t.Fatalf("T2 still resolves and deals exactly one damage instance to the victim: %d events (SVY-R10)", len(ev))
	}
	if ev[0].Source != raider.ID {
		t.Errorf("the damage event's Source is the dead unit's asset ID %d, got %d (FR-REP-004 Stage 5)", raider.ID, ev[0].Source)
	}
	if ev[0].Raw != 5 || ev[0].Amount != 5 {
		t.Errorf("the dead source deals its last-known effective power 5 (3 + Aggressive 2): raw %d, amount %d (SVY-R10, FR-ZON-004)", ev[0].Raw, ev[0].Amount)
	}
	noFormatErrors(t, s.g)
}

// --- FR-ZON-004: death triggers use last-known effective power ---

type deathBoard struct {
	s         *scene
	dead      *Asset
	bystander *Asset
	others    []*Asset // every non-Headquarters asset that may be chosen
}

// powerDeath has Zealot (the given power, health 3, Aggressive aggr) print
// "When this is killed, deal damage equal to its power to chosen non-Headquarters
// asset", then dies. If attacking it attacks a 4-power Wall on its controller's
// turn and the retaliation kills it. Otherwise it is attacked on the opponent's
// turn by a 5-power Hitter. Player 0 chooses the Bystander.
func powerDeath(t *testing.T, power, aggr int, attacking bool) deathBoard {
	t.Helper()
	s := newScene(t)
	c := unit("Zealot", power, 3, Keywords{Aggressive: aggr})
	c.Ability = &ability.Def{Triggers: []ability.Trigger{{
		Hook: ability.HookKilled,
		Effect: ability.Effect{
			Targets: []ability.Target{tgt(ability.AllegianceAny, ability.KindUnit|ability.KindStructure, false)},
			Steps:   []ability.Step{powerStep(0)},
		},
	}}}
	var b deathBoard
	b.s = s
	if attacking {
		b.dead = s.put(0, ZoneMainTheater, c)
		wall := s.put(1, ZoneMainTheater, unit("Wall", 4, 20, Keywords{}))
		b.bystander = s.put(1, ZoneMainTheater, unit("Bystander", 0, 20, Keywords{}))
		keep := s.put(0, ZoneMainTheater, structure("Ally Keep", 0, 20, Keywords{}))
		b.others = []*Asset{wall, b.bystander, keep}
		s.scriptAnyOf(0, nil, b.bystander)
		s.scriptAnyOf(1, nil)
		s.do(s.attack(b.dead, wall))
	} else {
		s.endTurn()
		b.dead = s.put(0, ZoneMainTheater, c)
		hitter := s.put(1, ZoneMainTheater, unit("Hitter", 5, 20, Keywords{}))
		b.bystander = s.put(1, ZoneMainTheater, unit("Bystander", 0, 20, Keywords{}))
		keep := s.put(0, ZoneMainTheater, structure("Ally Keep", 0, 20, Keywords{}))
		b.others = []*Asset{hitter, b.bystander, keep}
		s.scriptAnyOf(0, nil, b.bystander)
		s.scriptAnyOf(1, nil)
		s.do(s.attack(hitter, b.dead))
	}
	if s.alive(b.dead) {
		t.Fatalf("precondition: Zealot is destroyed")
	}
	return b
}

func TestFR_ZON_004_DeathTriggerUsesLastKnownPowerWithTheAggressiveBonus(t *testing.T) {
	// Covers: SVY-R10
	b := powerDeath(t, 3, 2, true)
	b.s.wantDamage(b.bystander, 5)
	ev := damageTo(b.s.g.Log, b.bystander)
	if len(ev) != 1 || ev[0].Source != b.dead.ID || ev[0].Amount != 5 {
		t.Errorf("one damage event of 5 from the dead unit (asset %d): %+v (FR-ZON-004, SVY-R10)", b.dead.ID, ev)
	}
}

func TestFR_ZON_004_DeathTriggerOutsideCombatHasNoAggressiveBonus(t *testing.T) {
	// Covers: SVY-R10
	b := powerDeath(t, 3, 2, false)
	b.s.wantDamage(b.bystander, 3)
	ev := damageTo(b.s.g.Log, b.bystander)
	if len(ev) != 1 || ev[0].Source != b.dead.ID || ev[0].Amount != 3 {
		t.Errorf("one damage event of 3 from the dead unit (asset %d): %+v (FR-ZON-004, SVY-R10)", b.dead.ID, ev)
	}
}

func TestFR_ZON_004_DeathTriggerOfAZeroPowerUnitDealsZero(t *testing.T) {
	// Covers: SVY-R10
	b := powerDeath(t, 0, 0, false)
	b.s.wantDamage(b.bystander, 0)
	for _, e := range nonCombatDamage(b.s.g.Log) {
		if e.Source == b.dead.ID && e.Amount != 0 {
			t.Errorf("a 0-power source commits no damage: %+v (SVY-R10)", e)
		}
	}
	noFormatErrors(t, b.s.g)
}

func TestFR_TGT_001_DeathTriggerPowerTargetExcludesHeadquartersAndTheDeadUnit(t *testing.T) {
	for _, attacking := range []bool{true, false} {
		b := powerDeath(t, 3, 2, attacking)
		var d *Decision
		for _, x := range b.s.asked(0) {
			if x.Kind == DecideTarget {
				x := x
				d = &x
			}
		}
		if d == nil {
			t.Fatalf("attacking=%v: with several legal targets, the controller is asked to choose", attacking)
		}
		got, want := refs(*d), ids(b.others...)
		if len(got) != len(want) {
			t.Fatalf("attacking=%v: options = %v, want %v: both Headquarters and the dead unit are excluded", attacking, got, want)
		}
		for i := range got {
			if got[i] != want[i] {
				t.Errorf("attacking=%v: options = %v, want %v", attacking, got, want)
			}
		}
	}
}

// --- FR-TGT-002: "all assets" ---

// allBoard is a board for "deal 4 damage to all assets". Ancient (power 0, health
// 1, printing that death trigger) attacks Wall and is killed by the retaliation.
type allBoard struct {
	s                                         *scene
	ancient, wall                             *Asset
	allyFodder, allyResist, allyArmor, allySc *Asset
	ward, stalwart, enemyFodder, enemyKeep    *Asset
	enemySc                                   *Asset
	before                                    []*Asset // every asset but Ancient, before the attack
}

func newAllBoard(t *testing.T, hqDamage [2]int, emptyDecks bool) *allBoard {
	t.Helper()
	s := newScene(t)
	if emptyDecks {
		s.g.Players[0].Deck = nil
		s.g.Players[1].Deck = nil
	}
	s.g.Players[0].HQ.Damage = hqDamage[0]
	s.g.Players[1].HQ.Damage = hqDamage[1]
	b := &allBoard{s: s}
	b.ancient = s.put(0, ZoneMainTheater, mortal("Ancient", 0, 1, killedTrigger(allDamage(4, ability.Filter{}))))
	b.wall = s.put(1, ZoneMainTheater, unit("Wall", 5, 30, Keywords{}))
	b.allyFodder = s.put(0, ZoneMainTheater, mortal("Ally Fodder", 0, 4, killedTrigger(drawN(1))))
	b.allyResist = s.put(0, ZoneMainTheater, unit("Ally Resistant", 0, 10, Keywords{Resistant: 1}))
	b.allyArmor = s.put(0, ZoneMainTheater, structure("Ally Armored Keep", 0, 10, Keywords{Armor: 2}))
	b.allySc = s.putScratch(0)
	b.ward = s.put(1, ZoneMainTheater, unit("Warded", 0, 10, Keywords{Ward: true}))
	b.stalwart = s.put(1, ZoneMainTheater, unit("Stalwart", 0, 10, Keywords{Stalwart: true}))
	b.enemyFodder = s.put(1, ZoneMainTheater, mortal("Enemy Fodder", 0, 4, killedTrigger(drawN(1))))
	b.enemyKeep = s.put(1, ZoneMainTheater, structure("Enemy Keep", 0, 4, Keywords{}))
	b.enemySc = s.putScratch(1)
	for _, a := range s.g.Assets() {
		if a != b.ancient {
			b.before = append(b.before, a)
		}
	}
	s.script(0)
	s.script(1)
	s.do(s.attack(b.ancient, b.wall))
	if s.alive(b.ancient) {
		s.t.Fatalf("precondition: the Wall's retaliation destroys Ancient")
	}
	return b
}

func TestFR_TGT_002_AllAssetsDealsOneInstanceToEveryAssetOnBothSidesWithoutAsking(t *testing.T) {
	b := newAllBoard(t, [2]int{}, false)
	for _, a := range b.before {
		ev := damageTo(b.s.g.Log, a)
		if len(ev) != 1 {
			t.Errorf("%s takes exactly one non-combat damage instance from 'all assets' (FR-REP-004): %d events", a, len(ev))
			continue
		}
		if ev[0].Source != b.ancient.ID {
			t.Errorf("%s: the damage event's Source is the dead Ancient %d, got %d", a, b.ancient.ID, ev[0].Source)
		}
	}
	for p := PlayerID(0); p < 2; p++ {
		if hq := b.s.g.Players[p].HQ; hq.Damage != 4 {
			t.Errorf("player %d's Headquarters is an asset and takes 4: damage %d", p, hq.Damage)
		}
		if n := len(b.s.asked(p)); n != 0 {
			t.Errorf("'all assets' targets nothing, so player %d is asked nothing: %d decisions (FR-TGT-002)", p, n)
		}
	}
	b.s.wantDamage(b.wall, 4)
	noFormatErrors(t, b.s.g)
}

func TestFR_KWD_002h_AllAssetsIgnoresWard(t *testing.T) {
	b := newAllBoard(t, [2]int{}, false)
	if !b.s.alive(b.ward) {
		t.Fatalf("precondition: the 10-health Warded unit survives 4 damage")
	}
	b.s.wantDamage(b.ward, 4)
}

func TestFR_REP_004_AllAssetsAppliesResistantStalwartAndArmorPerInstance(t *testing.T) {
	b := newAllBoard(t, [2]int{}, false)
	b.s.wantDamage(b.allyResist, 3) // Resistant 1 reduces the instance
	b.s.wantDamage(b.stalwart, 4)   // Stalwart only reduces combat damage
	b.s.wantDamage(b.allyArmor, 2)  // Armor 2 absorbs 2 of the 4
	if b.allyArmor.Armor != 0 {
		t.Errorf("Armor 2 absorbs 2 damage and is consumed: Armor = %d, want 0 (FR-KWD-002k)", b.allyArmor.Armor)
	}
}

func TestFR_TGT_002_AllAssetsKillsEveryAssetWithFourOrLessHealthInOneSweep(t *testing.T) {
	b := newAllBoard(t, [2]int{}, false)
	for _, a := range []*Asset{b.allyFodder, b.enemyFodder, b.enemyKeep, b.allySc, b.enemySc} {
		if b.s.alive(a) {
			t.Errorf("%s has 4 or less health and is destroyed by the 4-damage instance", a)
		}
	}
	for _, a := range []*Asset{b.allyResist, b.allyArmor, b.ward, b.stalwart, b.wall} {
		if !b.s.alive(a) {
			t.Errorf("%s has more than 4 effective health and survives", a)
		}
	}
	if b.s.g.Outcome.Over {
		t.Errorf("both Headquarters have 20 health and survive 4: outcome %+v", b.s.g.Outcome)
	}
}

func TestFR_TRG_003_AllAssetsDeathTriggersResolveInApnapOrder(t *testing.T) {
	// Both Fodders die in the one sweep and both draw from empty decks. The
	// non-active player's trigger resolves first (FR-TRG-003), so player 1 loses.
	b := newAllBoard(t, [2]int{}, true)
	if b.s.alive(b.allyFodder) || b.s.alive(b.enemyFodder) {
		t.Fatalf("precondition: both Fodders die to the all-assets damage")
	}
	if o := b.s.g.Outcome; !o.Over || o.Draw || o.Winner != 0 {
		t.Errorf("the units killed by the sweep fire their death triggers afterwards in APNAP order: the non-active player 1 draws first from an empty deck and loses, outcome %+v (FR-TRG-003, FR-TRG-004)", o)
	}
}

func TestFR_TRG_004_AllAssetsDeathTriggersCascade(t *testing.T) {
	// Ancient's 4 damage destroys Echo (p1, health 4), whose death trigger deals 1
	// damage to all opposing units. That finishes Survivor (p0, health 5, already
	// at 4), and Survivor's own death trigger draws a card.
	s := newScene(t)
	cards := s.onTop(0, vanilla("A", "B")...)
	ancient := s.put(0, ZoneMainTheater, mortal("Ancient", 0, 1, killedTrigger(allDamage(4, ability.Filter{}))))
	wall := s.put(1, ZoneMainTheater, unit("Wall", 5, 30, Keywords{}))
	survivor := s.put(0, ZoneMainTheater, mortal("Survivor", 0, 5, killedTrigger(drawN(1))))
	echo := s.put(1, ZoneMainTheater, mortal("Echo", 0, 4,
		killedTrigger(allDamage(1, ability.Filter{Allegiance: ability.AllegianceOpposing, Kinds: ability.KindUnit}))))
	s.script(0)
	s.script(1)
	s.do(s.attack(ancient, wall))
	if s.alive(echo) {
		t.Fatalf("precondition: Echo dies to the first 4-damage instance")
	}
	if s.alive(survivor) {
		t.Errorf("Echo's death trigger deals 1 more to Survivor, who had taken 4 of 5 health, so it dies (FR-TRG-004)")
	}
	if !inHand(s.g.Players[0], cards[0]) {
		t.Errorf("Survivor's death trigger, caused during the cascade, resolves and draws the top card (FR-TRG-004)")
	}
	if d := s.g.Players[0].HQ.Damage; d != 4 {
		t.Errorf("Echo's filtered damage skips Headquarters: HQ 0 damage %d, want 4", d)
	}
	if len(s.g.Stack) != 0 {
		t.Errorf("the stack is empty after the cascade: %d items", len(s.g.Stack))
	}
	noFormatErrors(t, s.g)
}

func TestFR_CBT_003_AllAssetsBothHeadquartersDestroyedIsADraw(t *testing.T) {
	b := newAllBoard(t, [2]int{16, 16}, false)
	o := b.s.g.Outcome
	if !o.Over || !o.Draw {
		t.Fatalf("both Headquarters reach 0 health at the same sweep: a Match Draw (FR-CBT-003), outcome %+v", o)
	}
	if !strings.Contains(strings.ToLower(o.Reason), "both players lost") {
		t.Errorf("the draw reason is the 'both players lost' reason, got %q", o.Reason)
	}
	noFormatErrors(t, b.s.g)
}

func TestFR_CBT_003_AllAssetsOneHeadquartersDestroyedLosesThatPlayer(t *testing.T) {
	for _, tc := range []struct {
		hq     [2]int
		winner PlayerID
	}{{[2]int{0, 16}, 0}, {[2]int{16, 0}, 1}} {
		b := newAllBoard(t, tc.hq, false)
		o := b.s.g.Outcome
		if !o.Over || o.Draw || o.Winner != tc.winner {
			t.Errorf("Headquarters damage %v: only the damaged player's Headquarters is destroyed, so player %d wins (FR-CBT-003): outcome %+v", tc.hq, tc.winner, o)
		}
	}
}

// --- FR-TGT-002: filtered "all" ---

// sweeper deploys a free Reserves unit for player p whose deploy trigger is the
// given step, on a board with one entity of each kind per side. It returns the
// source's asset (nil if the step destroyed it) and its card.
func sweeper(t *testing.T, p PlayerID, step ability.Step) (*scene, tboard, *Asset, *Card) {
	t.Helper()
	s := newScene(t)
	b := s.targetBoard()
	s.script(0)
	s.script(1)
	c := trigUnit("Sweeper", targetTrigger(nil, step))
	s.deployFromHand(p, c)
	return s, b, assetOf(s.g, c), c
}

func TestFR_TGT_002_FilteredAllOpposingUnitsHitsOnlyOpposingUnits(t *testing.T) {
	s, b, src, _ := sweeper(t, 0, allDamage(1, ability.Filter{Allegiance: ability.AllegianceOpposing, Kinds: ability.KindUnit}))
	s.wantDamage(b.u1, 1)
	if s.alive(b.c1) {
		t.Errorf("a Scratch Company is a unit and an opposing 1/1 one is destroyed by 1 damage")
	}
	s.wantDamage(b.s1, 0)
	s.wantDamage(s.g.Players[1].HQ, 0)
	s.wantDamage(b.u0, 0)
	s.wantDamage(b.s0, 0)
	s.wantDamage(b.c0, 0)
	s.wantDamage(src, 0)
	s.wantDamage(s.g.Players[0].HQ, 0)
	if n := len(s.asked(0)) + len(s.asked(1)); n != 0 {
		t.Errorf("'all' is not a target, nothing is asked: %d decisions", n)
	}
}

func TestFR_TGT_002_FilteredAllNonScratchSkipsScratchCompanies(t *testing.T) {
	s, b, _, _ := sweeper(t, 0, allDamage(1, ability.Filter{Allegiance: ability.AllegianceOpposing, Kinds: ability.KindUnit, NonScratch: true}))
	s.wantDamage(b.u1, 1)
	s.wantDamage(b.c1, 0)
}

func TestFR_TGT_002_FilteredAllStructuresAndHeadquartersHitsExactlyThose(t *testing.T) {
	s, b, _, _ := sweeper(t, 0, allDamage(1, ability.Filter{Allegiance: ability.AllegianceOpposing, Kinds: ability.KindStructure | ability.KindHeadquarters}))
	s.wantDamage(b.s1, 1)
	s.wantDamage(s.g.Players[1].HQ, 1)
	s.wantDamage(b.u1, 0)
	s.wantDamage(b.c1, 0)
	s.wantDamage(b.s0, 0)
	s.wantDamage(s.g.Players[0].HQ, 0)
}

func TestFR_TGT_002_FilteredAllAlliedUnitsIncludesTheSourceItself(t *testing.T) {
	// The filter has no "other", so the 1/1 Sweeper on the battlefield is hit.
	s, b, src, card := sweeper(t, 0, allDamage(1, ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit}))
	inDiscard := false
	for _, inst := range s.g.Players[0].Discard {
		if inst.Card == card {
			inDiscard = true
		}
	}
	if src != nil || !inDiscard {
		t.Errorf("the source is an allied unit still on the battlefield when the step resolves, so it takes the 1 damage and is destroyed: on battlefield %v, in Discard %v", src != nil, inDiscard)
	}
	s.wantDamage(b.u0, 1)
	if s.alive(b.c0) {
		t.Errorf("the allied Scratch Company 1/1 is destroyed")
	}
	s.wantDamage(b.u1, 0)
}

func TestFR_TGT_002_FilteredAllWithNoMatchDoesNothingAndAsksNothing(t *testing.T) {
	s := newScene(t)
	s.put(1, ZoneMainTheater, unit("Enemy Unit", 1, 5, Keywords{}))
	s.script(0)
	s.script(1)
	s.deployFromHand(0, trigUnit("Sweeper", targetTrigger(nil,
		allDamage(3, ability.Filter{Allegiance: ability.AllegianceOpposing, Kinds: ability.KindStructure}))))
	if d := totalDamage(s.g); d != 0 {
		t.Errorf("no opposing structure exists, so the step does nothing: %d damage on the battlefield", d)
	}
	if n := len(s.asked(0)) + len(s.asked(1)); n != 0 {
		t.Errorf("nothing is asked: %d decisions", n)
	}
	if len(s.g.Stack) != 0 {
		t.Errorf("the stack is empty: %d items", len(s.g.Stack))
	}
}

func TestFR_KWD_002h_FilteredAllHitsOpposingWardUnits(t *testing.T) {
	s := newScene(t)
	w := s.put(1, ZoneMainTheater, unit("Warded", 0, 5, Keywords{Ward: true}))
	s.script(0)
	s.deployFromHand(0, trigUnit("Sweeper", targetTrigger(nil,
		allDamage(2, ability.Filter{Allegiance: ability.AllegianceOpposing, Kinds: ability.KindUnit}))))
	s.wantDamage(w, 2)
}

func TestFR_KWD_002d_AllHealRemovesDamageFromEveryMatchingAsset(t *testing.T) {
	s := newScene(t)
	b := s.targetBoard()
	b.u0.Damage, b.s0.Damage, b.u1.Damage = 3, 2, 3
	s.g.Players[0].HQ.Damage = 2
	s.script(0)
	s.deployFromHand(0, trigUnit("Medic", targetTrigger(nil,
		allStep(ability.OpHeal, 2, ability.Filter{Allegiance: ability.AllegianceAllied}))))
	s.wantDamage(b.u0, 1)
	s.wantDamage(b.s0, 0)
	s.wantDamage(s.g.Players[0].HQ, 0)
	s.wantDamage(b.u1, 3)
}

func TestFR_LAY_002_AllAddArmorGrantsEveryMatchingAsset(t *testing.T) {
	s := newScene(t)
	b := s.targetBoard()
	s.script(0)
	c := trigUnit("Smith", targetTrigger(nil,
		allStep(ability.OpAddArmor, 2, ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit})))
	s.deployFromHand(0, c)
	src := assetOf(s.g, c)
	if b.u0.Armor != 2 || b.c0.Armor != 2 || src.Armor != 2 {
		t.Errorf("every allied unit gains 2 Armor (including the source and the Scratch Company): u0 %d, c0 %d, source %d", b.u0.Armor, b.c0.Armor, src.Armor)
	}
	if b.s0.Armor != 0 || b.u1.Armor != 0 {
		t.Errorf("structures and opposing units gain none: s0 %d, u1 %d", b.s0.Armor, b.u1.Armor)
	}
}

// --- no formatting failure in any event Detail ---

func TestFR_REP_004_NoEventDetailContainsAFormattingFailure(t *testing.T) {
	// A source that has left the battlefield must not reach the damage event's
	// Detail as a nil asset (follow-up from T-0017).
	for _, tc := range []struct {
		power, aggr int
		attacking   bool
	}{{3, 2, true}, {3, 2, false}, {0, 0, false}} {
		b := powerDeath(t, tc.power, tc.aggr, tc.attacking)
		noFormatErrors(t, b.s.g)
	}
	noFormatErrors(t, newAllBoard(t, [2]int{16, 16}, false).s.g)
	noFormatErrors(t, newAllBoard(t, [2]int{}, true).s.g)
}
