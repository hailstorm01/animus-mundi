package engine

import (
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0314 (FR-TRN-003, FR-SET-004, FR-ACT-001, FR-ACT-002, FR-TGT-001,
// FR-TGT-003, FR-MOD-001, FR-KWD-002h, FR-LAY-002, FR-REP-004, FR-ZON-004,
// FR-ZON-005, FR-ENT-103, FR-CBT-003): Dolus Macrocannon (Soulless structure,
// R4 T0, Main Theater, 2/6), loaded from data/cards: "R1, exhaust: Deal 2 damage
// to chosen non-Scratch asset." Player 0 controls it and is the active player.

// dmcScene is a scene with a Readied Macrocannon in player 0's Main Theater and
// pool R1 T0.
func dmcScene(t *testing.T) (*scene, *Asset) {
	t.Helper()
	s := newScene(t)
	m := s.put(0, ZoneMainTheater, s.card("Dolus Macrocannon"))
	s.setPool(0, 1, 0)
	return s, m
}

func TestCard_DolusMacrocannon_EntersMainTheaterAsTwoSix(t *testing.T) {
	// Covers: FR-TRN-003, FR-SET-004
	s := newScene(t)
	c := s.card("Dolus Macrocannon")
	inst := s.give(0, c)
	s.setPool(0, 4, 0)
	s.do(Action{Kind: ActDeploy, Player: 0, Card: inst.ID, Zone: ZoneMainTheater})
	a := assetOf(s.g, c)
	if a == nil || a.Zone != ZoneMainTheater || a.Controller != 0 {
		t.Fatalf("the Macrocannon is on the battlefield in player 0's Main Theater after deploying (FR-TRN-003)")
	}
	if p := a.Printed(); p.Power != 2 || p.Health != 6 {
		t.Errorf("the Macrocannon is a 2/6: got %d/%d", p.Power, p.Health)
	}
	if a.Exhausted {
		t.Errorf("a newly deployed structure enters Readied (FR-SET-004)")
	}
	if p := s.g.Players[0]; p.Req != 0 || p.Tac != 0 {
		t.Errorf("the Macrocannon costs R4: pool R%d T%d, want R0 T0 (FR-TRN-003)", p.Req, p.Tac)
	}

	s2 := newScene(t)
	c2 := s2.card("Dolus Macrocannon")
	inst2 := s2.give(0, c2)
	s2.setPool(0, 4, 0)
	s2.reject(Action{Kind: ActDeploy, Player: 0, Card: inst2.ID, Zone: ZoneReserves})
	if !inHand(s2.g.Players[0], inst2) || assetOf(s2.g, c2) != nil {
		t.Errorf("deploying the Macrocannon to Reserves is rejected and the card stays in hand (FR-TRN-003)")
	}

	s3 := newScene(t)
	c3 := s3.card("Dolus Macrocannon")
	inst3 := s3.give(0, c3)
	s3.setPool(0, 3, 0)
	s3.reject(Action{Kind: ActDeploy, Player: 0, Card: inst3.ID, Zone: ZoneMainTheater})
	if !inHand(s3.g.Players[0], inst3) || assetOf(s3.g, c3) != nil {
		t.Errorf("with R3 the Macrocannon (cost R4) cannot be deployed (FR-TRN-003)")
	}
}

func TestCard_DolusMacrocannon_DealsTwoToTheChosenAsset(t *testing.T) {
	// Covers: FR-ACT-001, FR-LAY-002
	s, m := dmcScene(t)
	hit := s.put(1, ZoneMainTheater, unit("Hit", 1, 5, ability.Keywords{}))
	other := s.put(1, ZoneMainTheater, unit("Other", 1, 5, ability.Keywords{}))
	ally := s.put(0, ZoneMainTheater, unit("Ally", 1, 5, ability.Keywords{}))
	activateChoosing(s, m, hit)
	s.wantDamage(hit, 2)
	evs := damageTo(s.g.Log, hit)
	if len(evs) != 1 {
		t.Fatalf("one non-combat damage event against the chosen unit, got %d (FR-LAY-002)", len(evs))
	}
	if e := evs[0]; e.Source != m.ID || e.Amount != 2 || e.Combat {
		t.Errorf("2 non-combat damage whose Source is the Macrocannon: %+v (Macrocannon asset %d)", e, m.ID)
	}
	if p := s.g.Players[0]; p.Req != 0 || p.Tac != 0 {
		t.Errorf("the R1 is paid: pool R%d T%d, want R0 T0 (FR-ACT-001)", p.Req, p.Tac)
	}
	if !s.alive(m) || !m.Exhausted {
		t.Errorf("the Macrocannon is Exhausted and on the battlefield: alive %v exhausted %v (FR-ACT-001)", s.alive(m), m.Exhausted)
	}
	s.wantDamage(other, 0)
	s.wantDamage(ally, 0)
	if got := totalDamage(s.g); got != 2 {
		t.Errorf("only the chosen unit is damaged: total %d, want 2", got)
	}
}

func TestCard_DolusMacrocannon_OffersEveryNonScratchAsset(t *testing.T) {
	// Covers: FR-TGT-001, SVY-R57, FR-ZON-005, FR-KWD-002h
	s, m := dmcScene(t)
	aRes := s.put(0, ZoneReserves, unit("AllyRes", 1, 5, ability.Keywords{}))
	aMain := s.put(0, ZoneMainTheater, unit("AllyMain", 1, 5, ability.Keywords{}))
	fRes := s.put(1, ZoneReserves, unit("FoeRes", 1, 5, ability.Keywords{}))
	fMain := s.put(1, ZoneMainTheater, unit("FoeMain", 1, 5, ability.Keywords{}))
	aWall := s.put(0, ZoneMainTheater, structure("AllyWall", 0, 5, ability.Keywords{}))
	fWall := s.put(1, ZoneMainTheater, structure("FoeWall", 0, 5, ability.Keywords{}))
	aWard := s.put(0, ZoneMainTheater, unit("AllyWard", 1, 5, ability.Keywords{Ward: true}))
	s.put(1, ZoneMainTheater, unit("FoeWard", 1, 5, ability.Keywords{Ward: true}))
	s.putScratch(0)
	s.putScratch(1)
	s.script(0, Answer{Picks: []int{0}})
	s.do(activate(m, 0))
	d := targetAsk(t, s)
	if d.Player != 0 || d.Min != 1 || d.Max != 1 {
		t.Errorf("the question is a mandatory single target decision for player 0: %+v", d)
	}
	wantOptions(t, d, "both Headquarters, the four units, both structures, the allied Warded unit and the Macrocannon; no Scratch Company and no opposing Warded unit",
		s.g.Players[0].HQ, s.g.Players[1].HQ, aRes, aMain, fRes, fMain, aWall, fWall, aWard, m)
}

func TestCard_DolusMacrocannon_UnitTurnedFaceDownNotOffered(t *testing.T) {
	// Covers: FR-ENT-103, SVY-R59
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Foe", 1, 5, ability.Keywords{}))
	morphEnact(s, morphManeuver())
	if a := s.g.Asset(foe.ID); a == nil || !a.FaceDown {
		t.Fatalf("setup: the opposing unit should be face down (FR-ENT-103)")
	}
	up := s.put(1, ZoneMainTheater, unit("FaceUp", 1, 5, ability.Keywords{}))
	m := s.put(0, ZoneMainTheater, s.card("Dolus Macrocannon"))
	s.setPool(0, 1, 0)
	s.script(0, Answer{Picks: []int{0}})
	s.do(activate(m, 0))
	wantOptions(t, targetAsk(t, s), "a face-down unit is a Scratch Company and not offered; the face-up unit, both Headquarters and the Macrocannon are",
		up, s.g.Players[0].HQ, s.g.Players[1].HQ, m)
}

func TestCard_DolusMacrocannon_MayChooseItself(t *testing.T) {
	// Covers: FR-TGT-001
	s, m := dmcScene(t)
	activateChoosing(s, m, m)
	s.wantDamage(m, 2)
	if !s.alive(m) || !m.Exhausted {
		t.Errorf("a 2/6 Macrocannon survives 2 damage and stays Exhausted: alive %v exhausted %v (FR-TGT-001)", s.alive(m), m.Exhausted)
	}
}

func TestCard_DolusMacrocannon_OfferedWithOnlyItselfAndHeadquarters(t *testing.T) {
	// Covers: SVY-R36, FR-ACT-001
	s, m := dmcScene(t)
	if !offersActivate(s.g, m, 0) {
		t.Fatalf("the Macrocannon is a legal target, so Activate is offered (SVY-R36)")
	}
	s.script(0, Answer{Picks: []int{0}})
	s.do(activate(m, 0))
	wantOptions(t, targetAsk(t, s), "the two Headquarters and the Macrocannon",
		s.g.Players[0].HQ, s.g.Players[1].HQ, m)
}

func TestCard_DolusMacrocannon_DamagesTheOpposingHeadquarters(t *testing.T) {
	// Covers: FR-CBT-003, FR-LAY-002
	s, m := dmcScene(t)
	hq1 := s.g.Players[1].HQ
	before := s.g.CurrentHealth(hq1)
	activateChoosing(s, m, hq1)
	if got := s.g.CurrentHealth(hq1); got != before-2 {
		t.Errorf("the opposing Headquarters remaining Health drops by 2: got %d, want %d (FR-CBT-003)", got, before-2)
	}
	if s.g.Outcome.Over {
		t.Errorf("the match is not over (FR-CBT-003)")
	}

	s2, m2 := dmcScene(t)
	hq := s2.g.Players[1].HQ
	hq.Damage += s2.g.CurrentHealth(hq) - 2
	if got := s2.g.CurrentHealth(hq); got != 2 {
		t.Fatalf("setup: the opposing Headquarters has 2 Health left, got %d", got)
	}
	activateChoosing(s2, m2, hq)
	if !s2.g.Outcome.Over || s2.g.Outcome.Draw || s2.g.Outcome.Winner != 0 {
		t.Errorf("an opposing Headquarters at 0 Health loses the match: outcome %+v (FR-CBT-003)", s2.g.Outcome)
	}
}

func TestCard_DolusMacrocannon_LethalDamageKills(t *testing.T) {
	// Covers: FR-LAY-002
	s, m := dmcScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Frail", 1, 2, ability.Keywords{}))
	activateChoosing(s, m, foe)
	if s.alive(foe) || !inDiscardPile(s.g.Players[1], foe.Inst) {
		t.Errorf("a Health 2 unit taking 2 dies and its card is in its owner's discard pile (FR-LAY-002)")
	}
}

func TestCard_DolusMacrocannon_ArmorAbsorbs(t *testing.T) {
	// Covers: FR-REP-004, FR-KWD-002k
	s, m := dmcScene(t)
	arm := s.put(1, ZoneMainTheater, unit("Armored", 1, 5, ability.Keywords{Armor: 1}))
	activateChoosing(s, m, arm)
	if arm.Armor != 0 {
		t.Errorf("the first point of damage removes the 1 Armor: Armor %d, want 0 (FR-KWD-002k)", arm.Armor)
	}
	s.wantDamage(arm, 1)
}

func TestCard_DolusMacrocannon_NotOfferedWithoutOneRequisition(t *testing.T) {
	// Covers: FR-ACT-001
	s, m := dmcScene(t)
	s.setPool(0, 0, 5)
	foe := s.put(1, ZoneMainTheater, unit("Foe", 1, 5, ability.Keywords{}))
	if offersActivate(s.g, m, 0) {
		t.Errorf("with R0 T5 the R1 cost cannot be paid, so Activate is not offered (FR-ACT-001)")
	}
	s.reject(activate(m, 0))
	p := s.g.Players[0]
	if m.Exhausted || p.Req != 0 || p.Tac != 5 || totalDamage(s.g) != 0 {
		t.Errorf("a rejected Activate moves nothing: exhausted %v, pool R%d T%d, damage %d (FR-ACT-001)", m.Exhausted, p.Req, p.Tac, totalDamage(s.g))
	}
	s.wantDamage(foe, 0)
}

func TestCard_DolusMacrocannon_NotOfferedWhileExhausted(t *testing.T) {
	// Covers: FR-ACT-002
	s, m := dmcScene(t)
	s.setPool(0, 5, 0)
	m.Exhausted = true
	if offersActivate(s.g, m, 0) {
		t.Errorf("an Exhausted Macrocannon cannot pay its exhaust cost, so Activate is not offered (FR-ACT-002)")
	}
	s.reject(activate(m, 0))
	if p := s.g.Players[0]; p.Req != 5 || totalDamage(s.g) != 0 {
		t.Errorf("a rejected Activate pays nothing and deals no damage: pool R%d, damage %d (FR-ACT-002)", p.Req, totalDamage(s.g))
	}
}

func TestCard_DolusMacrocannon_NotOfferedWhileTheStackIsNotEmpty(t *testing.T) {
	// Covers: FR-ACT-001
	s, m := dmcScene(t)
	s.g.Stack = append(s.g.Stack, &StackItem{Kind: StackScratch, Controller: 0})
	s.reject(activate(m, 0))
	if offersActivate(s.g, m, 0) {
		t.Errorf("no Activate is offered with an item on the stack (FR-ACT-001)")
	}
	if p := s.g.Players[0]; p.Req != 1 || m.Exhausted {
		t.Errorf("a rejected Activate spends nothing: R%d exhausted %v (FR-ACT-001)", p.Req, m.Exhausted)
	}
}

func TestCard_DolusMacrocannon_MayActivateTheTurnItDeploys(t *testing.T) {
	// Covers: FR-SET-004
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Foe", 1, 5, ability.Keywords{}))
	c := s.card("Dolus Macrocannon")
	inst := s.give(0, c)
	s.setPool(0, 5, 0)
	s.do(Action{Kind: ActDeploy, Player: 0, Card: inst.ID, Zone: ZoneMainTheater})
	m := assetOf(s.g, c)
	if m == nil {
		t.Fatalf("the Macrocannon is deployed")
	}
	if !offersActivate(s.g, m, 0) {
		t.Fatalf("a structure deployed this turn may activate right away (FR-SET-004)")
	}
	activateChoosing(s, m, foe)
	s.wantDamage(foe, 2)
}

func TestCard_DolusMacrocannon_TimeoutTakesTheFirstOption(t *testing.T) {
	// Covers: FR-MOD-001, SVY-R16
	s, m := dmcScene(t)
	s.targetBoard()
	s.script(0, Answer{Timeout: true})
	s.do(activate(m, 0))
	if n := len(s.asked(0)); n != 1 {
		t.Fatalf("asked %d times, want 1", n)
	}
	s.wantDamage(s.g.Players[0].HQ, 2)
	if got := totalDamage(s.g); got != 2 {
		t.Errorf("a timeout deals exactly 2 to exactly one target, the first in fixed order, the choosing player's own Headquarters (FR-MOD-001 sec. 6, SVY-R16): total %d", got)
	}
}

func TestCard_DolusMacrocannon_TargetGoneDealsNothing(t *testing.T) {
	// Covers: FR-TGT-003
	s := newScene(t)
	opposingUnits := ability.Filter{Allegiance: ability.AllegianceOpposing, Kinds: ability.KindUnit}
	m := s.put(0, ZoneMainTheater, s.cardWithTrigger("Dolus Macrocannon", exhaustTrigger(ability.Step{Op: ability.OpReturnToHand, Subject: ability.SubjectAll, Filter: opposingUnits})))
	s.setPool(0, 1, 0)
	foe := s.put(1, ZoneMainTheater, unit("Fleeting", 1, 5, ability.Keywords{}))
	activateChoosing(s, m, foe)
	if s.alive(foe) {
		t.Fatalf("setup: the synthetic trigger should have returned the target to hand before the ability resolves")
	}
	if !inHand(s.g.Players[1], foe.Inst) {
		t.Fatalf("setup: the target's card should be in player 1's hand")
	}
	if n := len(damageEvents(s.g.Log)); n != 0 {
		t.Errorf("a target that left before resolution is skipped and no damage event is logged: %d (FR-TGT-003)", n)
	}
	if got := totalDamage(s.g); got != 0 {
		t.Errorf("nothing takes damage: total %d (FR-TGT-003)", got)
	}
	if p := s.g.Players[0]; p.Req != 0 || !m.Exhausted || !s.alive(m) {
		t.Errorf("the costs stay paid and the Macrocannon stays: pool R%d, exhausted %v, alive %v (FR-TGT-003)", p.Req, m.Exhausted, s.alive(m))
	}
}

func TestCard_DolusMacrocannon_TargetGainsWardDealsNothing(t *testing.T) {
	// Covers: FR-TGT-003, FR-KWD-002h
	s := newScene(t)
	m := s.put(0, ZoneMainTheater, s.cardWithTrigger("Dolus Macrocannon", exhaustTrigger(ability.Step{Op: ability.OpAddArmor, Amount: ability.Value{N: 0}, Subject: ability.SubjectSelf})))
	s.setPool(0, 1, 0)
	foe := s.put(1, ZoneMainTheater, unit("Warding", 1, 5, ability.Keywords{}))
	chooseAsset(s, foe)
	fired, askedAtFire, stackAtFire := 0, 0, 0
	s.ptPlaceHook(func(asked int, _ StackItem) {
		// The synthetic trigger is placed after the exhaust cost is paid, above the
		// ability, which is already on the stack with its target chosen.
		if m.Exhausted && !foe.Inst.Card.Keywords.Ward {
			fired++
			askedAtFire, stackAtFire = asked, len(s.g.Stack)
			foe.Inst.Card.Keywords.Ward = true
		}
	})
	s.do(activate(m, 0))
	if fired != 1 {
		t.Fatalf("setup: Ward should be set once, after the costs were paid: fired %d", fired)
	}
	if askedAtFire < 1 || stackAtFire < 2 {
		t.Fatalf("setup: the target should have been chosen with the ability on the stack before Ward was set: asked %d, stack %d", askedAtFire, stackAtFire)
	}
	s.wantDamage(foe, 0)
	if got := totalDamage(s.g); got != 0 {
		t.Errorf("a target that gained Ward before resolution is no longer legal and takes no damage: total %d (FR-TGT-003, FR-KWD-002h)", got)
	}
	if p := s.g.Players[0]; p.Req != 0 || !m.Exhausted {
		t.Errorf("the costs stay paid: pool R%d, exhausted %v (FR-TGT-003)", p.Req, m.Exhausted)
	}
}

func TestCard_DolusMacrocannon_SourceGoneStillDeals(t *testing.T) {
	// Covers: FR-ZON-004, FR-TGT-003
	s := newScene(t)
	m := s.put(0, ZoneMainTheater, s.cardWithTrigger("Dolus Macrocannon", exhaustTrigger(ability.Step{Op: ability.OpKill, Subject: ability.SubjectSelf})))
	s.setPool(0, 1, 0)
	foe := s.put(1, ZoneMainTheater, unit("Foe", 1, 5, ability.Keywords{}))
	activateChoosing(s, m, foe)
	if s.alive(m) {
		t.Fatalf("setup: the Macrocannon should be dead before the ability resolves")
	}
	if !inDiscardPile(s.g.Players[0], m.Inst) {
		t.Errorf("the Macrocannon's card is in its owner's discard pile (FR-ZON-004)")
	}
	s.wantDamage(foe, 2)
}
