package engine

import (
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0320 (FR-TRN-003, FR-SET-004, FR-TGT-001, FR-TGT-003, FR-MOD-001,
// FR-LAY-002, FR-KWD-002d, FR-KWD-002h, FR-KWD-002k, FR-ZON-004, FR-ZON-005,
// FR-REP-004): MAX-8 Autonomous Warsuit (Soulless unit, Cybernetic and
// Mechanized, R6 T0, Reserves, 4/7), loaded from data/cards: "When this unit
// deploys, deal 2 damage to chosen asset. Chosen asset heals 2." Each chosen
// asset is its own slot (SVY-R35): the first is dealt 2 damage, the second is
// healed 2, and the two may be the same asset. Player 0 controls the Warsuit
// and is the active player. Uses ecDeploy from cards_soulless_earthshaker_test.go.

const mxName = "MAX-8 Autonomous Warsuit"

// mxChoose makes seat 0 answer its successive target questions with the asset
// each pick function returns when asked (so the Warsuit itself can be picked).
// It records every decision in s.asked(0).
func mxChoose(s *scene, picks ...func() *Asset) {
	if s.asks == nil {
		s.asks = map[PlayerID][]Decision{}
	}
	s.asks[0] = nil
	next := 0
	s.g.choosers[0] = func(_ *Game, d Decision) Answer {
		s.asks[0] = append(s.asks[0], d)
		if next >= len(picks) {
			s.t.Errorf("seat 0 asked past its script: %q", d.Prompt)
			return Answer{Timeout: true}
		}
		want := picks[next]()
		next++
		for i, o := range d.Options {
			if want != nil && o.Ref == want.ID {
				return Answer{Picks: []int{i}}
			}
		}
		s.t.Errorf("the chosen asset is not among the options %v", refs(d))
		return Answer{Timeout: true}
	}
}

func mxFixed(a *Asset) func() *Asset { return func() *Asset { return a } }

func mxSelf(s *scene, c *Card) func() *Asset { return func() *Asset { return assetOf(s.g, c) } }

func mxHeals(log []Event, a *Asset) []Event {
	var out []Event
	for _, e := range eventsOfKind(log, EventHeal) {
		if e.Target == a.ID {
			out = append(out, e)
		}
	}
	return out
}

func TestCard_MAX8AutonomousWarsuit_EntersReservesAsFourSeven(t *testing.T) {
	// Covers: FR-TRN-003, FR-SET-004
	s := newScene(t)
	mxChoose(s, mxFixed(s.g.Players[0].HQ), mxFixed(s.g.Players[0].HQ))
	c := s.card(mxName)
	ecDeploy(s, c)
	a := assetOf(s.g, c)
	if a == nil || a.Zone != ZoneReserves || a.Controller != 0 {
		t.Fatalf("the Warsuit is on the battlefield in player 0's Reserves after deploying (FR-TRN-003)")
	}
	if p := a.Printed(); p.Power != 4 || p.Health != 7 {
		t.Errorf("the Warsuit is a 4/7: got %d/%d", p.Power, p.Health)
	}
	if a.Exhausted {
		t.Errorf("a newly deployed unit enters Readied (FR-SET-004)")
	}
	if p := s.g.Players[0]; p.Req != 0 || p.Tac != 0 {
		t.Errorf("the Warsuit costs R6: pool R%d T%d, want R0 T0 (FR-TRN-003)", p.Req, p.Tac)
	}

	s2 := newScene(t)
	c2 := s2.card(mxName)
	inst2 := s2.give(0, c2)
	s2.setPool(0, 6, 0)
	s2.reject(Action{Kind: ActDeploy, Player: 0, Card: inst2.ID, Zone: ZoneMainTheater})
	if !inHand(s2.g.Players[0], inst2) || assetOf(s2.g, c2) != nil {
		t.Errorf("deploying the Warsuit to the Main Theater is rejected and the card stays in hand (FR-TRN-003)")
	}

	s3 := newScene(t)
	c3 := s3.card(mxName)
	inst3 := s3.give(0, c3)
	s3.setPool(0, 5, 0)
	s3.reject(Action{Kind: ActDeploy, Player: 0, Card: inst3.ID, Zone: ZoneReserves})
	if !inHand(s3.g.Players[0], inst3) || assetOf(s3.g, c3) != nil {
		t.Errorf("with R5 the Warsuit (cost R6) cannot be deployed (FR-TRN-003)")
	}
}

func TestCard_MAX8AutonomousWarsuit_AsksTwoSlotsDamageFirst(t *testing.T) {
	// Covers: FR-TGT-001, SVY-R35, SVY-R3
	s := newScene(t)
	b := s.targetBoard()
	s.put(1, ZoneMainTheater, unit("Warded", 1, 5, ability.Keywords{Ward: true}))
	b.u0.Damage = 2
	mxChoose(s, mxFixed(b.u1), mxFixed(b.u0))
	askedAtPlace := -1
	s.ptPlaceHook(func(n int, _ StackItem) {
		if askedAtPlace < 0 {
			askedAtPlace = n
		}
	})
	c := s.card(mxName)
	ecDeploy(s, c)
	war := assetOf(s.g, c)
	if war == nil {
		t.Fatalf("the Warsuit is on the battlefield")
	}
	asks := targetAsks(s.asked(0))
	if len(asks) != 2 {
		t.Fatalf("two target questions, one per chosen-asset slot: asked %d (FR-TGT-001, SVY-R35)", len(asks))
	}
	if askedAtPlace != 2 {
		t.Errorf("both targets are chosen on placement, before the trigger resolves: %d asked at placement (SVY-R3)", askedAtPlace)
	}
	for i, d := range asks {
		if d.Player != 0 {
			t.Errorf("question %d is put to player 0: %+v", i, d)
		}
		wantOptions(t, d, "both Headquarters, the six board assets and the Warsuit; no opposing Warded unit",
			s.g.Players[0].HQ, s.g.Players[1].HQ, b.u0, b.s0, b.c0, b.u1, b.s1, b.c1, war)
	}
	s.wantDamage(b.u1, 2)
	s.wantDamage(b.u0, 0)
	if evs := damageTo(s.g.Log, b.u1); len(evs) != 1 || evs[0].Amount != 2 || evs[0].Source != war.ID {
		t.Errorf("the first answer takes one non-combat 2 damage instance from the Warsuit: %+v", evs)
	}
	if got := totalDamage(s.g); got != 2 {
		t.Errorf("no other asset's damage changes: total %d, want 2", got)
	}
}

func TestCard_MAX8AutonomousWarsuit_AlliedWardedUnitIsOffered(t *testing.T) {
	// Covers: FR-KWD-002h
	s := newScene(t)
	allyWard := s.put(0, ZoneMainTheater, unit("AllyWard", 1, 5, ability.Keywords{Ward: true}))
	foeWard := s.put(1, ZoneMainTheater, unit("FoeWard", 1, 5, ability.Keywords{Ward: true}))
	hq := s.g.Players[0].HQ
	mxChoose(s, mxFixed(hq), mxFixed(hq))
	ecDeploy(s, s.card(mxName))
	asks := targetAsks(s.asked(0))
	if len(asks) != 2 {
		t.Fatalf("two target questions: asked %d", len(asks))
	}
	for i, d := range asks {
		var gotAlly, gotFoe bool
		for _, o := range d.Options {
			gotAlly = gotAlly || o.Ref == allyWard.ID
			gotFoe = gotFoe || o.Ref == foeWard.ID
		}
		if !gotAlly {
			t.Errorf("question %d offers the allied Warded unit (FR-KWD-002h)", i)
		}
		if gotFoe {
			t.Errorf("question %d does not offer the opposing Warded unit (FR-KWD-002h)", i)
		}
	}
}

func TestCard_MAX8AutonomousWarsuit_SameAssetForBothTakesTwoThenHealsTwo(t *testing.T) {
	// Covers: SVY-R35, FR-LAY-002
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Frail", 1, 2, ability.Keywords{}))
	mxChoose(s, mxFixed(foe), mxFixed(foe))
	ecDeploy(s, s.card(mxName))
	if !s.alive(foe) {
		t.Fatalf("the death sweep runs only after the whole trigger, so the unit healed back to 0 damage survives (FR-LAY-002, SVY-R35)")
	}
	s.wantDamage(foe, 0)
	dmg, heal := damageTo(s.g.Log, foe), mxHeals(s.g.Log, foe)
	if len(dmg) != 1 || dmg[0].Amount != 2 || len(heal) != 1 || heal[0].Amount != 2 {
		t.Fatalf("it took 2 damage and then healed 2: damage %+v, heal %+v", dmg, heal)
	}
	if dmg[0].Seq >= heal[0].Seq {
		t.Errorf("the damage step precedes the heal step (SVY-R35)")
	}
}

func TestCard_MAX8AutonomousWarsuit_HealsAnotherAssetLethallyDamagedOneDies(t *testing.T) {
	// Covers: FR-LAY-002, FR-KWD-002d
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Frail", 1, 2, ability.Keywords{}))
	ally := s.put(0, ZoneMainTheater, unit("Hurt", 1, 9, ability.Keywords{}))
	ally.Damage = 3
	mxChoose(s, mxFixed(foe), mxFixed(ally))
	ecDeploy(s, s.card(mxName))
	if s.alive(foe) || !inDiscardPile(s.g.Players[1], foe.Inst) {
		t.Errorf("a Health 2 unit taking 2 dies at the sweep and its card is in its owner's discard pile (FR-LAY-002)")
	}
	s.wantDamage(ally, 1)
}

func TestCard_MAX8AutonomousWarsuit_HealNeverGoesBelowZero(t *testing.T) {
	// Covers: FR-KWD-002d
	s := newScene(t)
	ally := s.put(0, ZoneMainTheater, unit("Hurt", 1, 9, ability.Keywords{}))
	ally.Damage = 1
	mxChoose(s, mxFixed(s.g.Players[1].HQ), mxFixed(ally))
	ecDeploy(s, s.card(mxName))
	s.wantDamage(ally, 0)

	s2 := newScene(t)
	whole := s2.put(0, ZoneMainTheater, unit("Whole", 1, 9, ability.Keywords{}))
	mxChoose(s2, mxFixed(s2.g.Players[1].HQ), mxFixed(whole))
	ecDeploy(s2, s2.card(mxName))
	s2.wantDamage(whole, 0)
	if h := mxHeals(s2.g.Log, whole); len(h) != 1 || h[0].Amount != 0 {
		t.Errorf("healing an undamaged asset logs a heal of 0 and changes nothing: %+v", h)
	}
}

func TestCard_MAX8AutonomousWarsuit_CanHealAHeadquarters(t *testing.T) {
	// Covers: FR-KWD-002d, FR-CBT-003
	s := newScene(t)
	hq0, hq1 := s.g.Players[0].HQ, s.g.Players[1].HQ
	hq0.Damage = 3
	before := s.g.CurrentHealth(hq1)
	mxChoose(s, mxFixed(hq1), mxFixed(hq0))
	ecDeploy(s, s.card(mxName))
	s.wantDamage(hq0, 1)
	if got := s.g.CurrentHealth(hq1); got != before-2 {
		t.Errorf("the opposing Headquarters remaining Health drops by 2: got %d, want %d", got, before-2)
	}
	if s.g.Outcome.Over {
		t.Errorf("the match is not over")
	}
}

func TestCard_MAX8AutonomousWarsuit_CanDamageItself(t *testing.T) {
	// Covers: FR-TGT-001
	s := newScene(t)
	ally := s.put(0, ZoneMainTheater, unit("Ally", 1, 9, ability.Keywords{}))
	c := s.card(mxName)
	mxChoose(s, mxSelf(s, c), mxFixed(ally))
	ecDeploy(s, c)
	war := assetOf(s.g, c)
	if war == nil {
		t.Fatalf("a 4/7 Warsuit survives 2 damage")
	}
	s.wantDamage(war, 2)
}

func TestCard_MAX8AutonomousWarsuit_ScratchCompanyCanBeKilled(t *testing.T) {
	// Covers: FR-ZON-005, FR-LAY-002
	s := newScene(t)
	sc := s.putScratch(1)
	hq := s.g.Players[0].HQ
	mxChoose(s, mxFixed(sc), mxFixed(hq))
	ecDeploy(s, s.card(mxName))
	if s.alive(sc) {
		t.Errorf("a Scratch Company (1/1) taking 2 damage is destroyed at the sweep (FR-ZON-005, FR-LAY-002)")
	}
}

func TestCard_MAX8AutonomousWarsuit_ArmorAbsorbsAndHealLeavesArmor(t *testing.T) {
	// Covers: FR-REP-004, FR-KWD-002k
	s := newScene(t)
	arm := s.put(1, ZoneMainTheater, unit("Armored", 1, 5, ability.Keywords{Armor: 1}))
	mxChoose(s, mxFixed(arm), mxFixed(s.g.Players[0].HQ))
	ecDeploy(s, s.card(mxName))
	if arm.Armor != 0 {
		t.Errorf("the first point of damage removes the 1 Armor: Armor %d, want 0 (FR-KWD-002k)", arm.Armor)
	}
	s.wantDamage(arm, 1)

	s2 := newScene(t)
	arm2 := s2.put(1, ZoneMainTheater, unit("Armored", 1, 5, ability.Keywords{Armor: 1}))
	arm2.Damage = 1
	mxChoose(s2, mxFixed(arm2), mxFixed(arm2))
	ecDeploy(s2, s2.card(mxName))
	s2.wantDamage(arm2, 0)
	if arm2.Armor != 0 {
		t.Errorf("heal leaves Armor alone and never restores it: Armor %d, want 0 (FR-KWD-002d)", arm2.Armor)
	}
}

func TestCard_MAX8AutonomousWarsuit_TimeoutTakesTheFirstOptionForEach(t *testing.T) {
	// Covers: FR-MOD-001, SVY-R16
	s := newScene(t)
	s.targetBoard()
	hq0 := s.g.Players[0].HQ
	hq0.Damage = 1
	s.script(0, Answer{Timeout: true}, Answer{Timeout: true})
	ecDeploy(s, s.card(mxName))
	if n := len(s.asked(0)); n != 2 {
		t.Fatalf("asked %d times, want 2", n)
	}
	s.wantDamage(hq0, 1)
	if got := totalDamage(s.g); got != 1 {
		t.Errorf("each timeout takes the first option in fixed order, the choosing player's own Headquarters: it took 2 and healed 2 (FR-MOD-001 sec. 6, SVY-R16): total %d, want 1", got)
	}
}

func TestCard_MAX8AutonomousWarsuit_DamageTargetGoneStillHeals(t *testing.T) {
	// Covers: FR-TGT-003, SVY-R25
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Fleeting", 1, 5, ability.Keywords{}))
	ally := s.put(0, ZoneMainTheater, unit("Hurt", 1, 9, ability.Keywords{}))
	ally.Damage = 2
	mxChoose(s, mxFixed(foe), mxFixed(ally))
	s.ptPlaceHook(func(int, StackItem) {
		if s.alive(foe) {
			stormKill(s, foe)
		}
	})
	ecDeploy(s, s.card(mxName))
	if s.alive(foe) {
		t.Fatalf("setup: the damage target should have left before the trigger resolves")
	}
	if n := len(nonCombatDamage(s.g.Log)); n != 0 {
		t.Errorf("no damage is dealt to a target that left: %d events (FR-TGT-003, SVY-R25)", n)
	}
	s.wantDamage(ally, 0)
}

func TestCard_MAX8AutonomousWarsuit_HealTargetGoneStillDamages(t *testing.T) {
	// Covers: FR-TGT-003
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Foe", 1, 5, ability.Keywords{}))
	ally := s.put(0, ZoneMainTheater, unit("Fleeting", 1, 9, ability.Keywords{}))
	ally.Damage = 2
	mxChoose(s, mxFixed(foe), mxFixed(ally))
	s.ptPlaceHook(func(int, StackItem) {
		if s.alive(ally) {
			stormKill(s, ally)
		}
	})
	ecDeploy(s, s.card(mxName))
	if s.alive(ally) {
		t.Fatalf("setup: the heal target should have left before the trigger resolves")
	}
	s.wantDamage(foe, 2)
}

func TestCard_MAX8AutonomousWarsuit_SourceGoneStillResolves(t *testing.T) {
	// Covers: FR-ZON-004, FR-TGT-003
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Foe", 1, 5, ability.Keywords{}))
	ally := s.put(0, ZoneMainTheater, unit("Hurt", 1, 9, ability.Keywords{}))
	ally.Damage = 2
	mxChoose(s, mxFixed(foe), mxFixed(ally))
	c := s.card(mxName)
	s.ptPlaceHook(func(int, StackItem) {
		if war := assetOf(s.g, c); war != nil {
			stormKill(s, war)
		}
	})
	inst := ecDeploy(s, c)
	if assetOf(s.g, c) != nil {
		t.Fatalf("setup: the Warsuit should have left before the trigger resolves")
	}
	s.wantDamage(foe, 2)
	s.wantDamage(ally, 0)
	if !inDiscardPile(s.g.Players[0], inst) {
		t.Errorf("the Warsuit card is in its owner's discard pile (FR-ZON-004)")
	}
}

func TestCard_MAX8AutonomousWarsuit_FaceDownScratchDoesNothing(t *testing.T) {
	// Covers: FR-ZON-005
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Foe", 1, 5, ability.Keywords{}))
	s.script(0)
	s.onTop(0, s.card(mxName))
	inst := s.give(0, s.card("Growing Legion"))
	s.setPool(0, 0, 1)
	s.do(enactAction(inst))
	if n := faceDownCount(s.g.Players[0]); n != 1 {
		t.Fatalf("Scratch 1 makes one Scratch Company in player 0's Reserves: %d face-down units (FR-ZON-005)", n)
	}
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("a face-down Scratch Company has no abilities, so no target is asked: asked %d (FR-ZON-005)", n)
	}
	s.wantDamage(foe, 0)
	if got := totalDamage(s.g); got != 0 {
		t.Errorf("no asset's damage changes: total %d (FR-ZON-005)", got)
	}
}
