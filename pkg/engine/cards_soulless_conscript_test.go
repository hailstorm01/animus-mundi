package engine

import (
	"reflect"
	"sort"
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0311: Conscript Rifles (unit, R2 T0, 2/2, Reserves): "When this
// unit deploys, deal 1 damage to chosen non-Scratch Company unit." The real card
// is loaded from data/cards. Player 0 deploys it from hand to Reserves with a
// pool of R2 and is the active player, unless a case says otherwise. The card
// is always a legal target of its own trigger, so a case with no other legal
// unit asks nothing (the Single-Option Fast-Path, FR-TGT-003).

// crDeployCard gives player 0 the card c and R2, and deploys it to Reserves.
func crDeployCard(s *scene, c *Card) *Instance {
	s.t.Helper()
	inst := s.give(0, c)
	s.setPool(0, 2, 0)
	s.do(Action{Kind: ActDeploy, Player: 0, Card: inst.ID, Zone: ZoneReserves})
	return inst
}

// crWantOffered checks that d offers exactly the given assets, in any order.
func crWantOffered(t *testing.T, d Decision, what string, want ...*Asset) {
	t.Helper()
	got, w := append([]int(nil), refs(d)...), ids(want...)
	sort.Ints(got)
	sort.Ints(w)
	if !reflect.DeepEqual(got, w) {
		t.Errorf("%s: options got %v, want %v", what, got, w)
	}
}

// crOneQuestion returns the one DecideTarget question put to player 0.
func crOneQuestion(t *testing.T, s *scene) Decision {
	t.Helper()
	asked := s.asked(0)
	if len(asked) != 1 {
		t.Fatalf("one chosen target is one question: asked %d times (FR-TGT-001)", len(asked))
	}
	d := asked[0]
	if d.Kind != DecideTarget || d.Player != 0 || d.Min != 1 || d.Max != 1 {
		t.Errorf("the question is a mandatory single DecideTarget for the deploying player: %+v", d)
	}
	return d
}

func TestCard_ConscriptRifles_EntersReservesAsTwoTwo(t *testing.T) {
	// Covers: FR-TRN-003, FR-SET-004
	s := newScene(t)
	s.script(0)
	c := s.card("Conscript Rifles")
	crDeployCard(s, c)
	a := assetOf(s.g, c)
	if a == nil || a.Zone != ZoneReserves || a.Controller != 0 {
		t.Fatalf("Conscript Rifles is on the battlefield in player 0's Reserves after deploying (FR-TRN-003)")
	}
	if p := a.Printed(); p.Power != 2 || p.Health != 2 {
		t.Errorf("Conscript Rifles is a 2/2: got %d/%d", p.Power, p.Health)
	}
	if a.Exhausted {
		t.Errorf("a newly deployed unit enters Readied (FR-SET-004)")
	}
	if s.g.Players[0].Req != 0 {
		t.Errorf("Conscript Rifles costs 2 Requisition: %d left", s.g.Players[0].Req)
	}

	s2 := newScene(t)
	c2 := s2.card("Conscript Rifles")
	inst := s2.give(0, c2)
	s2.setPool(0, 2, 0)
	s2.reject(Action{Kind: ActDeploy, Player: 0, Card: inst.ID, Zone: ZoneMainTheater})
	if !inHand(s2.g.Players[0], inst) || assetOf(s2.g, c2) != nil {
		t.Errorf("deploying Conscript Rifles to the Main Theater is rejected and the card stays in hand (FR-TRN-003)")
	}

	s3 := newScene(t)
	c3 := s3.card("Conscript Rifles")
	inst3 := s3.give(0, c3)
	s3.setPool(0, 1, 0)
	s3.reject(Action{Kind: ActDeploy, Player: 0, Card: inst3.ID, Zone: ZoneReserves})
	if !inHand(s3.g.Players[0], inst3) || assetOf(s3.g, c3) != nil {
		t.Errorf("with 1 Requisition Conscript Rifles (cost 2) cannot be deployed (FR-TRN-003)")
	}
}

func TestCard_ConscriptRifles_AloneDealsOneToItself(t *testing.T) {
	// Covers: FR-TGT-001, FR-TGT-003
	s := newScene(t)
	s.script(0)
	c := s.card("Conscript Rifles")
	crDeployCard(s, c)
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("the only legal unit is Conscript Rifles itself, so nothing is asked: asked %d times (Single-Option Fast-Path)", n)
	}
	if len(s.g.Stack) != 0 {
		t.Errorf("the trigger has resolved and the stack is empty: %d items", len(s.g.Stack))
	}
	a := assetOf(s.g, c)
	if a == nil {
		t.Fatalf("Conscript Rifles survives 1 damage (Health 2)")
	}
	s.wantDamage(a, 1)
	if got := s.g.MaxHealth(a) - a.Damage; got != 1 {
		t.Errorf("Conscript Rifles has Current Health 1: got %d (FR-TGT-001)", got)
	}
	s.wantDamage(s.g.Players[0].HQ, 0)
	s.wantDamage(s.g.Players[1].HQ, 0)
}

func TestCard_ConscriptRifles_DealsOneToTheChosenUnit(t *testing.T) {
	// Covers: FR-TGT-001, FR-LAY-002
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Foe", 1, 3, ability.Keywords{}))
	stormChoose(s, func() *Asset { return foe })
	c := s.card("Conscript Rifles")
	crDeployCard(s, c)
	d := crOneQuestion(t, s)
	crWantOffered(t, d, "the opposing unit and Conscript Rifles itself (FR-TGT-001)", foe, assetOf(s.g, c))
	s.wantDamage(foe, 1)
	s.wantDamage(assetOf(s.g, c), 0)
	if got := totalDamage(s.g); got != 1 {
		t.Errorf("exactly 1 damage is dealt in total: %d (FR-LAY-002)", got)
	}
}

func TestCard_ConscriptRifles_OffersEveryNonScratchUnit(t *testing.T) {
	// Covers: FR-TGT-001, FR-ZON-005
	s := newScene(t)
	allyMain := s.put(0, ZoneMainTheater, unit("AllyMain", 1, 5, ability.Keywords{}))
	allyRes := s.put(0, ZoneReserves, unit("AllyRes", 1, 5, ability.Keywords{}))
	foeMain := s.put(1, ZoneMainTheater, unit("FoeMain", 1, 5, ability.Keywords{}))
	foeRes := s.put(1, ZoneReserves, unit("FoeRes", 1, 5, ability.Keywords{}))
	s.put(0, ZoneMainTheater, structure("AllyStructure", 0, 5, ability.Keywords{}))
	s.put(1, ZoneMainTheater, structure("FoeStructure", 0, 5, ability.Keywords{}))
	s.putScratch(0)
	s.putScratch(1)
	stormChoose(s, func() *Asset { return foeMain })
	c := s.card("Conscript Rifles")
	crDeployCard(s, c)
	d := crOneQuestion(t, s)
	crWantOffered(t, d, "the four units and Conscript Rifles; no structure, Headquarters or Scratch Company (FR-TGT-001, FR-ZON-005)",
		allyMain, allyRes, foeMain, foeRes, assetOf(s.g, c))
}

func TestCard_ConscriptRifles_UnitTurnedFaceDownNotOffered(t *testing.T) {
	// Covers: FR-ENT-103, FR-TGT-001
	s := newScene(t)
	morphed := s.put(1, ZoneMainTheater, unit("Morphed", 1, 5, ability.Keywords{}))
	morphEnact(s, morphManeuver())
	if !morphed.FaceDown {
		t.Fatalf("setup: the opposing unit should be face down")
	}
	foe := s.put(1, ZoneReserves, unit("FaceUp", 1, 5, ability.Keywords{}))
	stormChoose(s, func() *Asset { return foe })
	c := s.card("Conscript Rifles")
	crDeployCard(s, c)
	d := crOneQuestion(t, s)
	crWantOffered(t, d, "a unit turned face down is a Scratch Company and is not offered; the face-up unit and Conscript Rifles are (FR-ENT-103, FR-TGT-001)",
		foe, assetOf(s.g, c))
	s.wantDamage(morphed, 0)
	s.wantDamage(foe, 1)
}

func TestCard_ConscriptRifles_LethalDamageKills(t *testing.T) {
	// Covers: FR-LAY-002, FR-ZON-002
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Frail", 1, 1, ability.Keywords{}))
	inst := foe.Inst
	stormChoose(s, func() *Asset { return foe })
	crDeployCard(s, s.card("Conscript Rifles"))
	if s.alive(foe) || !inDiscardPile(s.g.Players[1], inst) {
		t.Errorf("a Health 1 unit taking 1 dies and its card is in its owner's discard pile (FR-LAY-002, FR-ZON-002)")
	}
}

func TestCard_ConscriptRifles_ResistantAndArmorApply(t *testing.T) {
	// Covers: FR-REP-004, FR-KWD-002k
	s := newScene(t)
	res := s.put(1, ZoneMainTheater, unit("Resistant", 1, 5, ability.Keywords{Resistant: 1}))
	stormChoose(s, func() *Asset { return res })
	crDeployCard(s, s.card("Conscript Rifles"))
	s.wantDamage(res, 0)

	s2 := newScene(t)
	arm := s2.put(1, ZoneMainTheater, unit("Armored", 1, 5, ability.Keywords{Armor: 1}))
	stormChoose(s2, func() *Asset { return arm })
	crDeployCard(s2, s2.card("Conscript Rifles"))
	s2.wantDamage(arm, 0)
	if arm.Armor != 0 {
		t.Errorf("the 1 Armor absorbs the 1 damage and is spent: Armor %d, want 0 (FR-KWD-002k)", arm.Armor)
	}
}

func TestCard_ConscriptRifles_WardedOpposingUnitNotOffered(t *testing.T) {
	// Covers: FR-KWD-002h
	s := newScene(t)
	foeWard := s.put(1, ZoneMainTheater, unit("FoeWard", 1, 5, ability.Keywords{Ward: true}))
	s.script(0)
	c := s.card("Conscript Rifles")
	crDeployCard(s, c)
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("an opposing Warded unit is not offered, so Conscript Rifles is the only option and nothing is asked: asked %d (FR-KWD-002h)", n)
	}
	s.wantDamage(foeWard, 0)
	s.wantDamage(assetOf(s.g, c), 1)

	s2 := newScene(t)
	allyWard := s2.put(0, ZoneMainTheater, unit("AllyWard", 1, 5, ability.Keywords{Ward: true}))
	s2.put(1, ZoneMainTheater, unit("FoeWard2", 1, 5, ability.Keywords{Ward: true}))
	stormChoose(s2, func() *Asset { return allyWard })
	c2 := s2.card("Conscript Rifles")
	crDeployCard(s2, c2)
	d := crOneQuestion(t, s2)
	crWantOffered(t, d, "an allied Warded unit is offered, the opposing one is not (FR-KWD-002h)", allyWard, assetOf(s2.g, c2))
	s2.wantDamage(allyWard, 1)
}

func TestCard_ConscriptRifles_TargetGoneDealsNothing(t *testing.T) {
	// Covers: FR-TGT-003, SVY-R25
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Fleeting", 1, 5, ability.Keywords{}))
	stormChoose(s, func() *Asset { return foe })
	s.ptPlaceHook(func(int, StackItem) {
		if s.alive(foe) {
			stormKill(s, foe)
		}
	})
	c := s.card("Conscript Rifles")
	crDeployCard(s, c)
	if n := len(damageEvents(s.g.Log)); n != 0 {
		t.Errorf("a target that left before resolution is skipped and no damage event is logged: %d (FR-TGT-003, SVY-R25)", n)
	}
	if got := totalDamage(s.g); got != 0 {
		t.Errorf("nothing takes damage, Conscript Rifles included: total %d (FR-TGT-003)", got)
	}
	if assetOf(s.g, c) == nil {
		t.Errorf("Conscript Rifles stays on the battlefield")
	}
}

func TestCard_ConscriptRifles_SourceGoneStillDeals(t *testing.T) {
	// Covers: FR-ZON-004, FR-TGT-003
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Foe", 1, 5, ability.Keywords{}))
	stormChoose(s, func() *Asset { return foe })
	c := s.card("Conscript Rifles")
	s.ptPlaceHook(func(int, StackItem) {
		if cr := assetOf(s.g, c); cr != nil {
			stormKill(s, cr)
		}
	})
	inst := crDeployCard(s, c)
	if assetOf(s.g, c) != nil {
		t.Fatalf("setup: Conscript Rifles should have left before resolution")
	}
	s.wantDamage(foe, 1)
	if !inDiscardPile(s.g.Players[0], inst) {
		t.Errorf("the card of Conscript Rifles is in its owner's discard pile (FR-ZON-004)")
	}
}

func TestCard_ConscriptRifles_FaceDownScratchDoesNothing(t *testing.T) {
	// Covers: FR-ZON-005
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Foe", 1, 5, ability.Keywords{}))
	s.script(0)
	s.onTop(0, s.card("Conscript Rifles"))
	inst := s.give(0, s.card("Growing Legion"))
	s.setPool(0, 0, 1)
	s.do(enactAction(inst))
	p := s.g.Players[0]
	if n := faceDownCount(p); n != 1 {
		t.Fatalf("Scratch 1 makes one Scratch Company in player 0's Reserves: %d face-down units (FR-ZON-005)", n)
	}
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("a face-down Scratch Company has no abilities, so no target is asked: asked %d (FR-ZON-005)", n)
	}
	s.wantDamage(foe, 0)
	if got := totalDamage(s.g); got != 0 {
		t.Errorf("no unit is dealt damage: total %d (FR-ZON-005)", got)
	}
}
