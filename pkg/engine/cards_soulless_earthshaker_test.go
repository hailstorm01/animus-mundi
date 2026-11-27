package engine

import (
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0316 (FR-TRN-003, FR-SET-004, FR-TGT-001, FR-TGT-003, FR-MOD-001,
// FR-CBT-001, FR-CBT-002, FR-CBT-003, FR-LAY-002, FR-REP-004, FR-KWD-002h,
// FR-KWD-002k, FR-ZON-004, FR-ZON-005, FR-ENT-103): Earthshaker Colossus
// (Soulless unit, Mechanized and Giant, R6 T0, Reserves, 6/7), loaded from
// data/cards: "Whenever this unit deploys or attacks, deal 2 damage to chosen
// non-Scratch asset." Player 0 controls it and is the active player.

// ecDeploy gives player 0 the card with exactly R6 and deploys it to Reserves.
// It returns the card's instance.
func ecDeploy(s *scene, c *Card) *Instance {
	s.t.Helper()
	inst := s.give(0, c)
	s.setPool(0, 6, 0)
	s.do(Action{Kind: ActDeploy, Player: 0, Card: inst.ID, Zone: ZoneReserves})
	return inst
}

// ecBoard puts the board used by the offers cases (the Colossus excluded):
// units in Reserves and the Main Theater for both sides, a structure each, a
// Scratch Company each, and a Warded unit each. It returns the assets that must
// be offered, with FoeMain (the only opposing unit in the Main Theater) at
// index 5; the opposing Warded unit and the Scratch Companies are not offered.
func ecBoard(s *scene) []*Asset {
	aRes := s.put(0, ZoneReserves, unit("AllyRes", 1, 5, ability.Keywords{}))
	aMain := s.put(0, ZoneMainTheater, unit("AllyMain", 1, 5, ability.Keywords{}))
	fRes := s.put(1, ZoneReserves, unit("FoeRes", 1, 5, ability.Keywords{}))
	fMain := s.put(1, ZoneMainTheater, unit("FoeMain", 1, 20, ability.Keywords{}))
	aWall := s.put(0, ZoneMainTheater, structure("AllyWall", 0, 5, ability.Keywords{}))
	fWall := s.put(1, ZoneMainTheater, structure("FoeWall", 0, 5, ability.Keywords{}))
	aWard := s.put(0, ZoneMainTheater, unit("AllyWard", 1, 5, ability.Keywords{Ward: true}))
	s.put(1, ZoneReserves, unit("FoeWard", 1, 5, ability.Keywords{Ward: true}))
	s.putScratch(0)
	s.putScratch(1)
	return []*Asset{s.g.Players[0].HQ, s.g.Players[1].HQ, aRes, aMain, fRes, fMain, aWall, fWall, aWard}
}

// ecAttacker is a deployed Colossus in player 0's Main Theater, Readied, whose
// deploy trigger was answered with player 1's Headquarters (so it has 2 damage).
// pick is the asset every later target question is answered with.
type ecAttacker struct {
	s    *scene
	col  *Asset
	hq1  *Asset
	pick *Asset
}

func ecAttackScene(t *testing.T) *ecAttacker {
	t.Helper()
	s := newScene(t)
	e := &ecAttacker{s: s, hq1: s.g.Players[1].HQ}
	e.pick = e.hq1
	stormChoose(s, func() *Asset { return e.pick })
	c := s.card("Earthshaker Colossus")
	ecDeploy(s, c)
	e.col = assetOf(s.g, c)
	if e.col == nil {
		t.Fatalf("precondition: the Colossus is on the battlefield")
	}
	s.wantDamage(e.hq1, 2)
	colossusCycle(s, e.col)
	return e
}

func TestCard_EarthshakerColossus_EntersReservesAsSixSeven(t *testing.T) {
	// Covers: FR-TRN-003, FR-SET-004
	s := newScene(t)
	s.script(0, Answer{Picks: []int{0}})
	c := s.card("Earthshaker Colossus")
	ecDeploy(s, c)
	a := assetOf(s.g, c)
	if a == nil || a.Zone != ZoneReserves || a.Controller != 0 {
		t.Fatalf("the Colossus is on the battlefield in player 0's Reserves after deploying (FR-TRN-003)")
	}
	if p := a.Printed(); p.Power != 6 || p.Health != 7 {
		t.Errorf("the Colossus is a 6/7: got %d/%d", p.Power, p.Health)
	}
	if a.Exhausted {
		t.Errorf("a newly deployed unit enters Readied (FR-SET-004)")
	}
	if p := s.g.Players[0]; p.Req != 0 || p.Tac != 0 {
		t.Errorf("the Colossus costs R6: pool R%d T%d, want R0 T0 (FR-TRN-003)", p.Req, p.Tac)
	}

	s2 := newScene(t)
	c2 := s2.card("Earthshaker Colossus")
	inst2 := s2.give(0, c2)
	s2.setPool(0, 6, 0)
	s2.reject(Action{Kind: ActDeploy, Player: 0, Card: inst2.ID, Zone: ZoneMainTheater})
	if !inHand(s2.g.Players[0], inst2) || assetOf(s2.g, c2) != nil {
		t.Errorf("deploying the Colossus to the Main Theater is rejected and the card stays in hand (FR-TRN-003)")
	}

	s3 := newScene(t)
	c3 := s3.card("Earthshaker Colossus")
	inst3 := s3.give(0, c3)
	s3.setPool(0, 5, 0)
	s3.reject(Action{Kind: ActDeploy, Player: 0, Card: inst3.ID, Zone: ZoneReserves})
	if !inHand(s3.g.Players[0], inst3) || assetOf(s3.g, c3) != nil {
		t.Errorf("with R5 the Colossus (cost R6) cannot be deployed (FR-TRN-003)")
	}
}

func TestCard_EarthshakerColossus_DeployDealsTwoToTheChosenAsset(t *testing.T) {
	// Covers: FR-TGT-001, FR-LAY-002
	s := newScene(t)
	hit := s.put(1, ZoneMainTheater, unit("Hit", 1, 5, ability.Keywords{}))
	other := s.put(1, ZoneMainTheater, unit("Other", 1, 5, ability.Keywords{}))
	ally := s.put(0, ZoneMainTheater, unit("Ally", 1, 5, ability.Keywords{}))
	chooseAsset(s, hit)
	c := s.card("Earthshaker Colossus")
	ecDeploy(s, c)
	col := assetOf(s.g, c)
	s.wantDamage(hit, 2)
	evs := damageTo(s.g.Log, hit)
	if len(evs) != 1 {
		t.Fatalf("one non-combat damage event against the chosen unit, got %d (FR-LAY-002)", len(evs))
	}
	if e := evs[0]; e.Source != col.ID || e.Amount != 2 || e.Combat {
		t.Errorf("2 non-combat damage whose Source is the Colossus: %+v (Colossus asset %d)", e, col.ID)
	}
	s.wantDamage(other, 0)
	s.wantDamage(ally, 0)
	s.wantDamage(col, 0)
	if got := totalDamage(s.g); got != 2 {
		t.Errorf("only the chosen unit is damaged: total %d, want 2", got)
	}
}

func TestCard_EarthshakerColossus_DeployOffersEveryNonScratchAsset(t *testing.T) {
	// Covers: FR-TGT-001, FR-ZON-005, FR-KWD-002h, SVY-R59
	s := newScene(t)
	offered := ecBoard(s)
	chooseAsset(s, s.g.Players[1].HQ)
	c := s.card("Earthshaker Colossus")
	ecDeploy(s, c)
	d := targetAsk(t, s)
	if d.Player != 0 || d.Min != 1 || d.Max != 1 {
		t.Errorf("the question is a mandatory single target decision for player 0: %+v", d)
	}
	wantOptions(t, d, "both Headquarters, the four units, both structures, the allied Warded unit and the Colossus; no Scratch Company and no opposing Warded unit",
		append(offered, assetOf(s.g, c))...)
}

func TestCard_EarthshakerColossus_UnitTurnedFaceDownNotOffered(t *testing.T) {
	// Covers: FR-ENT-103, SVY-R59
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Foe", 1, 5, ability.Keywords{}))
	morphEnact(s, morphManeuver())
	if a := s.g.Asset(foe.ID); a == nil || !a.FaceDown {
		t.Fatalf("setup: the opposing unit should be face down (FR-ENT-103)")
	}
	up := s.put(1, ZoneMainTheater, unit("FaceUp", 1, 5, ability.Keywords{}))
	chooseAsset(s, up)
	c := s.card("Earthshaker Colossus")
	ecDeploy(s, c)
	wantOptions(t, targetAsk(t, s), "a face-down unit is a Scratch Company and not offered; the face-up unit, both Headquarters and the Colossus are",
		up, s.g.Players[0].HQ, s.g.Players[1].HQ, assetOf(s.g, c))
}

func TestCard_EarthshakerColossus_AloneStillAsks(t *testing.T) {
	// Covers: FR-TGT-001, FR-MOD-001
	s := newScene(t)
	c := s.card("Earthshaker Colossus")
	stormChoose(s, func() *Asset { return assetOf(s.g, c) })
	ecDeploy(s, c)
	col := assetOf(s.g, c)
	wantOptions(t, targetAsk(t, s), "both Headquarters are always legal, so the target is asked: the two Headquarters and the Colossus",
		s.g.Players[0].HQ, s.g.Players[1].HQ, col)
	if !s.alive(col) {
		t.Fatalf("a 6/7 Colossus survives 2 damage")
	}
	s.wantDamage(col, 2)
}

func TestCard_EarthshakerColossus_DeployDamagesTheOpposingHeadquarters(t *testing.T) {
	// Covers: FR-CBT-003
	s := newScene(t)
	hq1 := s.g.Players[1].HQ
	before := s.g.CurrentHealth(hq1)
	chooseAsset(s, hq1)
	ecDeploy(s, s.card("Earthshaker Colossus"))
	if got := s.g.CurrentHealth(hq1); got != before-2 {
		t.Errorf("the opposing Headquarters remaining Health drops by 2: got %d, want %d (FR-CBT-003)", got, before-2)
	}
	if s.g.Outcome.Over {
		t.Errorf("the match is not over (FR-CBT-003)")
	}
}

func TestCard_EarthshakerColossus_TimeoutTakesTheFirstOption(t *testing.T) {
	// Covers: FR-MOD-001, SVY-R16
	s := newScene(t)
	s.targetBoard()
	s.script(0, Answer{Timeout: true})
	ecDeploy(s, s.card("Earthshaker Colossus"))
	if n := len(s.asked(0)); n != 1 {
		t.Fatalf("asked %d times, want 1", n)
	}
	s.wantDamage(s.g.Players[0].HQ, 2)
	if got := totalDamage(s.g); got != 2 {
		t.Errorf("a timeout deals exactly 2 to exactly one target, the first in fixed order, the choosing player's own Headquarters (FR-MOD-001 sec. 6, SVY-R16): total %d", got)
	}
}

func TestCard_EarthshakerColossus_LethalDamageKills(t *testing.T) {
	// Covers: FR-LAY-002
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Frail", 1, 2, ability.Keywords{}))
	chooseAsset(s, foe)
	ecDeploy(s, s.card("Earthshaker Colossus"))
	if s.alive(foe) || !inDiscardPile(s.g.Players[1], foe.Inst) {
		t.Errorf("a Health 2 unit taking 2 dies and its card is in its owner's discard pile (FR-LAY-002)")
	}
}

func TestCard_EarthshakerColossus_ArmorAbsorbs(t *testing.T) {
	// Covers: FR-REP-004, FR-KWD-002k
	s := newScene(t)
	arm := s.put(1, ZoneMainTheater, unit("Armored", 1, 5, ability.Keywords{Armor: 1}))
	chooseAsset(s, arm)
	ecDeploy(s, s.card("Earthshaker Colossus"))
	if arm.Armor != 0 {
		t.Errorf("the first point of damage removes the 1 Armor: Armor %d, want 0 (FR-KWD-002k)", arm.Armor)
	}
	s.wantDamage(arm, 1)
}

func TestCard_EarthshakerColossus_TargetGoneDealsNothing(t *testing.T) {
	// Covers: FR-TGT-003, SVY-R25
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Fleeting", 1, 5, ability.Keywords{}))
	stormChoose(s, func() *Asset { return foe })
	s.ptPlaceHook(func(int, StackItem) {
		if s.alive(foe) {
			stormKill(s, foe)
		}
	})
	c := s.card("Earthshaker Colossus")
	ecDeploy(s, c)
	if s.alive(foe) {
		t.Fatalf("setup: the target should have left before the trigger resolves")
	}
	if n := len(damageEvents(s.g.Log)); n != 0 {
		t.Errorf("a target that left before resolution is skipped and no damage event is logged: %d (FR-TGT-003, SVY-R25)", n)
	}
	if got := totalDamage(s.g); got != 0 {
		t.Errorf("no asset is dealt damage: total %d (FR-TGT-003)", got)
	}
	if assetOf(s.g, c) == nil {
		t.Errorf("the Colossus stays on the battlefield")
	}
}

func TestCard_EarthshakerColossus_SourceGoneStillDeals(t *testing.T) {
	// Covers: FR-ZON-004, FR-TGT-003
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Foe", 1, 5, ability.Keywords{}))
	stormChoose(s, func() *Asset { return foe })
	c := s.card("Earthshaker Colossus")
	s.ptPlaceHook(func(int, StackItem) {
		if col := assetOf(s.g, c); col != nil {
			stormKill(s, col)
		}
	})
	inst := ecDeploy(s, c)
	if assetOf(s.g, c) != nil {
		t.Fatalf("setup: the Colossus should have left before the trigger resolves")
	}
	s.wantDamage(foe, 2)
	if !inDiscardPile(s.g.Players[0], inst) {
		t.Errorf("the Colossus card is in its owner's discard pile (FR-ZON-004)")
	}
}

func TestCard_EarthshakerColossus_FaceDownScratchDoesNothing(t *testing.T) {
	// Covers: FR-ZON-005
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Foe", 1, 5, ability.Keywords{}))
	s.script(0)
	s.onTop(0, s.card("Earthshaker Colossus"))
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
		t.Errorf("no asset is dealt damage: total %d (FR-ZON-005)", got)
	}
}

// --- Attacking ---

func TestCard_EarthshakerColossus_AttackDealsTwoToTheChosenAsset(t *testing.T) {
	// Covers: FR-CBT-001, FR-CBT-002, SVY-R3
	e := ecAttackScene(t)
	s := e.s
	def := s.put(1, ZoneMainTheater, unit("Defender", 1, 20, ability.Keywords{}))
	s.do(s.attack(e.col, def))
	s.wantDamage(e.hq1, 4)
	s.wantDamage(def, 6)
	s.wantDamage(e.col, 1)
	if n := len(s.asked(0)); n != 2 {
		t.Errorf("one target question for the deploy and one for the attack: asked %d times, want 2", n)
	}
}

func TestCard_EarthshakerColossus_AttackOffersEveryNonScratchAsset(t *testing.T) {
	// Covers: FR-TGT-001, FR-ZON-005, FR-KWD-002h
	e := ecAttackScene(t)
	s := e.s
	offered := ecBoard(s)
	s.do(s.attack(e.col, offered[5]))
	asks := s.asked(0)
	if len(asks) != 2 {
		t.Fatalf("one question for the deploy and one for the attack: asked %d times", len(asks))
	}
	wantOptions(t, asks[1], "both Headquarters, the four units (the defender and the units in Reserves included), both structures, the allied Warded unit and the Colossus; no Scratch Company and no opposing Warded unit",
		append(offered, e.col)...)
}

func TestCard_EarthshakerColossus_AttackTriggerResolvesBeforeCombat(t *testing.T) {
	// Covers: FR-CBT-002, FR-LAY-002
	e := ecAttackScene(t)
	s := e.s
	def := s.put(1, ZoneMainTheater, unit("Brittle", 5, 2, ability.Keywords{}))
	e.pick = def
	s.do(s.attack(e.col, def))
	if s.alive(def) || !inDiscardPile(s.g.Players[1], def.Inst) {
		t.Errorf("the defender takes 2 from the attack trigger and is destroyed at the death sweep, its card in its owner's discard pile (FR-LAY-002)")
	}
	s.wantDamage(e.col, 0)
	s.wantDamage(e.hq1, 2)
}

func TestCard_EarthshakerColossus_EachAttackTriggersAgain(t *testing.T) {
	// Covers: FR-CBT-002
	e := ecAttackScene(t)
	s := e.s
	dummy := s.put(1, ZoneMainTheater, unit("Dummy", 0, 30, ability.Keywords{}))
	chosen := s.put(1, ZoneReserves, unit("Chosen", 1, 5, ability.Keywords{}))
	s.do(s.attack(e.col, dummy))
	s.wantDamage(e.hq1, 4)
	s.endTurn()
	s.endTurn()
	e.pick = chosen
	s.do(s.attack(e.col, dummy))
	s.wantDamage(chosen, 2)
	s.wantDamage(e.hq1, 4)
	if n := len(s.asked(0)); n != 3 {
		t.Errorf("one target question for the deploy and one for each attack: asked %d times, want 3", n)
	}
}

func TestCard_EarthshakerColossus_DefendingDoesNotTrigger(t *testing.T) {
	// Covers: FR-CBT-002
	s := newScene(t)
	hq1 := s.g.Players[1].HQ
	chooseAsset(s, hq1)
	c := s.card("Earthshaker Colossus")
	ecDeploy(s, c)
	col := assetOf(s.g, c)
	s.do(Action{Kind: ActMarch, Player: 0, Asset: col.ID})
	s.endTurn()
	hitter := s.put(1, ZoneMainTheater, unit("Hitter", 1, 30, ability.Keywords{}))
	before := len(s.asked(0))
	s.do(s.attack(hitter, col))
	if n := len(s.asked(0)); n != before {
		t.Errorf("only attacking triggers the Colossus: asked %d more times while defending", n-before)
	}
	if n := len(nonCombatDamage(s.g.Log)); n != 1 {
		t.Errorf("only the deploy trigger dealt non-combat damage: %d events", n)
	}
	s.wantDamage(col, 1)
	s.wantDamage(hq1, 2)
}
