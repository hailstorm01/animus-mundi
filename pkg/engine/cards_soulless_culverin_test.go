package engine

import (
	"reflect"
	"sort"
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0313: Culverin Brigade (unit, R4 T0, 5/4, Reserves): "When this
// unit deploys, deal 1 damage to each of up to 2 chosen opposing assets." The
// real card is loaded from data/cards. Player 0 deploys it from hand to Reserves
// with a pool of exactly R4 T0 and is the active player, unless a case says
// otherwise. The other cards are synthetic. Every opposing asset a case does not
// mean to kill has enough health to survive 1 damage.

// cbDeploy gives player 0 the card c and R4 T0, and deploys it to Reserves.
func cbDeploy(s *scene, c *Card) *Instance {
	s.t.Helper()
	inst := s.give(0, c)
	s.setPool(0, 4, 0)
	s.do(Action{Kind: ActDeploy, Player: 0, Card: inst.ID, Zone: ZoneReserves})
	return inst
}

// cbWantOffered checks that d offers exactly the given assets, in any order.
func cbWantOffered(t *testing.T, d Decision, what string, want ...*Asset) {
	t.Helper()
	got, w := append([]int(nil), refs(d)...), ids(want...)
	sort.Ints(got)
	sort.Ints(w)
	if !reflect.DeepEqual(got, w) {
		t.Errorf("%s: options got %v, want %v", what, got, w)
	}
}

// cbOneQuestion returns the one "up to" DecideTarget question put to player 0.
func cbOneQuestion(t *testing.T, s *scene, max int) Decision {
	t.Helper()
	asked := s.asked(0)
	if len(asked) != 1 {
		t.Fatalf("one up-to-2 slot is one question: asked %d times (FR-TGT-002)", len(asked))
	}
	d := asked[0]
	if d.Kind != DecideTarget || d.Player != 0 || d.Min != 0 || d.Max != max {
		t.Errorf("the question is a DecideTarget for the deploying player with Min 0, Max %d: %+v", max, d)
	}
	return d
}

func cbFoe(s *scene, name string, zone Zone, health int) *Asset {
	return s.put(1, zone, unit(name, 1, health, ability.Keywords{}))
}

func TestCard_CulverinBrigade_EntersReservesAsFiveFour(t *testing.T) {
	// Covers: FR-TRN-003, FR-SET-004
	s := newScene(t)
	s.scriptPlan(0, nil, nil, pick())
	c := s.card("Culverin Brigade")
	cbDeploy(s, c)
	a := assetOf(s.g, c)
	if a == nil || a.Zone != ZoneReserves || a.Controller != 0 {
		t.Fatalf("Culverin Brigade is on the battlefield in player 0's Reserves after deploying (FR-TRN-003)")
	}
	if p := a.Printed(); p.Power != 5 || p.Health != 4 {
		t.Errorf("Culverin Brigade is a 5/4: got %d/%d", p.Power, p.Health)
	}
	if a.Exhausted {
		t.Errorf("a newly deployed unit enters Readied (FR-SET-004)")
	}
	if p := s.g.Players[0]; p.Req != 0 || p.Tac != 0 {
		t.Errorf("Culverin Brigade costs R4 T0, leaving an empty pool: R%d T%d", p.Req, p.Tac)
	}

	s2 := newScene(t)
	c2 := s2.card("Culverin Brigade")
	inst := s2.give(0, c2)
	s2.setPool(0, 4, 0)
	s2.reject(Action{Kind: ActDeploy, Player: 0, Card: inst.ID, Zone: ZoneMainTheater})
	if !inHand(s2.g.Players[0], inst) || assetOf(s2.g, c2) != nil {
		t.Errorf("deploying Culverin Brigade to the Main Theater is rejected and the card stays in hand (FR-TRN-003)")
	}
}

func TestCard_CulverinBrigade_DealsOneToEachOfTwoChosen(t *testing.T) {
	// Covers: FR-TGT-002, FR-LAY-002, SVY-R11
	s := newScene(t)
	f1 := cbFoe(s, "Foe1", ZoneMainTheater, 5)
	f2 := cbFoe(s, "Foe2", ZoneReserves, 5)
	f3 := cbFoe(s, "Foe3", ZoneMainTheater, 5)
	s.scriptPlan(0, nil, nil, pick(f1, f2))
	c := s.card("Culverin Brigade")
	cbDeploy(s, c)
	cbOneQuestion(t, s, 2)
	s.wantDamage(f1, 1)
	s.wantDamage(f2, 1)
	s.wantDamage(f3, 0)
	ev := nonCombatDamage(s.g.Log)
	if len(ev) != 2 {
		t.Errorf("two chosen assets are two non-combat damage instances: %d (FR-LAY-002)", len(ev))
	}
	for _, e := range ev {
		if e.Source != assetOf(s.g, c).ID {
			t.Errorf("the damage comes from Culverin Brigade: source %d, want %d", e.Source, assetOf(s.g, c).ID)
		}
	}
}

func TestCard_CulverinBrigade_OffersEveryOpposingAsset(t *testing.T) {
	// Covers: FR-TGT-001, FR-KWD-002h, FR-ZON-005
	s := newScene(t)
	s.put(0, ZoneMainTheater, unit("AllyMain", 1, 5, ability.Keywords{}))
	s.put(0, ZoneReserves, unit("AllyRes", 1, 5, ability.Keywords{}))
	foeMain := cbFoe(s, "FoeMain", ZoneMainTheater, 5)
	foeRes := cbFoe(s, "FoeRes", ZoneReserves, 5)
	s.put(0, ZoneMainTheater, structure("AllyStructure", 0, 5, ability.Keywords{}))
	foeStruct := s.put(1, ZoneMainTheater, structure("FoeStructure", 0, 5, ability.Keywords{}))
	s.putScratch(0)
	foeScratch := s.putScratch(1)
	s.put(1, ZoneMainTheater, unit("FoeWard", 1, 5, ability.Keywords{Ward: true}))
	s.scriptPlan(0, nil, nil, pick())
	cbDeploy(s, s.card("Culverin Brigade"))
	d := cbOneQuestion(t, s, 2)
	cbWantOffered(t, d, "player 1's Headquarters, the two opposing units, the opposing structure and the opposing Scratch Company; no allied asset, Culverin Brigade or Warded unit (FR-TGT-001, FR-KWD-002h, FR-ZON-005)",
		s.g.Players[1].HQ, foeMain, foeRes, foeStruct, foeScratch)
}

func TestCard_CulverinBrigade_MayChooseOne(t *testing.T) {
	// Covers: SVY-R11
	s := newScene(t)
	f1 := cbFoe(s, "Foe1", ZoneMainTheater, 5)
	f2 := cbFoe(s, "Foe2", ZoneMainTheater, 5)
	s.scriptPlan(0, nil, nil, pick(f1))
	cbDeploy(s, s.card("Culverin Brigade"))
	wantDealt(t, s.g, f1, 1, 1, "the one chosen unit takes 1 damage")
	wantDealt(t, s.g, f2, 0, 0, "the unchosen unit takes none")
	if n := len(nonCombatDamage(s.g.Log)); n != 1 {
		t.Errorf("one chosen asset is one damage instance in all: %d", n)
	}
}

func TestCard_CulverinBrigade_MayChooseNone(t *testing.T) {
	// Covers: FR-TGT-002
	s := newScene(t)
	cbFoe(s, "Foe1", ZoneMainTheater, 5)
	s.scriptPlan(0, nil, nil, pick())
	c := s.card("Culverin Brigade")
	cbDeploy(s, c)
	cbOneQuestion(t, s, 2)
	wantNoNonCombatDamage(t, s.g, "up to 2 allows choosing none")
	if got := totalDamage(s.g); got != 0 {
		t.Errorf("nothing takes damage: total %d (FR-TGT-002)", got)
	}
	if assetOf(s.g, c) == nil {
		t.Errorf("Culverin Brigade stays on the battlefield")
	}
}

func TestCard_CulverinBrigade_TimeoutChoosesNone(t *testing.T) {
	// Covers: FR-MOD-001
	s := newScene(t)
	cbFoe(s, "Foe1", ZoneMainTheater, 5)
	cbFoe(s, "Foe2", ZoneReserves, 5)
	s.scriptPlan(0, nil, nil, timeoutPick)
	cbDeploy(s, s.card("Culverin Brigade"))
	cbOneQuestion(t, s, 2)
	wantNoNonCombatDamage(t, s.g, "an up-to-N selection that times out chooses none (FR-MOD-001 section 6)")
	if got := totalDamage(s.g); got != 0 {
		t.Errorf("nothing takes damage: total %d (FR-MOD-001 section 6)", got)
	}
}

func TestCard_CulverinBrigade_OnlyTheHeadquartersStillAsks(t *testing.T) {
	// Covers: FR-MOD-001, FR-CBT-003
	s := newScene(t)
	hq1 := s.g.Players[1].HQ
	s.scriptPlan(0, nil, nil, pick(hq1))
	cbDeploy(s, s.card("Culverin Brigade"))
	d := cbOneQuestion(t, s, 1)
	wantOptions(t, d, "only player 1's Headquarters is legal", hq1)
	s.wantDamage(hq1, 1)
	if s.g.Outcome.Over {
		t.Errorf("1 damage to a Headquarters does not end the match (FR-CBT-003): %+v", s.g.Outcome)
	}
}

func TestCard_CulverinBrigade_LethalDamageKills(t *testing.T) {
	// Covers: FR-LAY-002
	s := newScene(t)
	frail := cbFoe(s, "Frail", ZoneMainTheater, 1)
	inst := frail.Inst
	other := cbFoe(s, "Sturdy", ZoneReserves, 5)
	s.scriptPlan(0, nil, nil, pick(frail, other))
	cbDeploy(s, s.card("Culverin Brigade"))
	if s.alive(frail) || !inDiscardPile(s.g.Players[1], inst) {
		t.Errorf("a Health 1 unit taking 1 dies and its card is in its owner's discard pile (FR-LAY-002)")
	}
	s.wantDamage(other, 1)
}

func TestCard_CulverinBrigade_ArmorAbsorbs(t *testing.T) {
	// Covers: FR-REP-004, FR-KWD-002k
	s := newScene(t)
	arm := s.put(1, ZoneMainTheater, unit("Armored", 1, 5, ability.Keywords{Armor: 1}))
	s.scriptPlan(0, nil, nil, pick(arm))
	cbDeploy(s, s.card("Culverin Brigade"))
	s.wantDamage(arm, 0)
	if arm.Armor != 0 {
		t.Errorf("the 1 Armor absorbs the 1 damage and is spent: Armor %d, want 0 (FR-KWD-002k)", arm.Armor)
	}
}

func TestCard_CulverinBrigade_OneTargetGoneTheOtherStillTakesOne(t *testing.T) {
	// Covers: FR-TGT-003, SVY-R25
	s := newScene(t)
	gone := cbFoe(s, "Fleeting", ZoneMainTheater, 5)
	stay := cbFoe(s, "Staying", ZoneReserves, 5)
	s.scriptPlan(0, nil, nil, pick(gone, stay))
	done := false
	s.ptPlaceHook(func(int, StackItem) {
		if !done && s.alive(gone) {
			done = true
			stormKill(s, gone)
		}
	})
	cbDeploy(s, s.card("Culverin Brigade"))
	if !done {
		t.Fatalf("setup: the trigger should have been placed")
	}
	s.wantDamage(stay, 1)
	if n := len(nonCombatDamage(s.g.Log)); n != 1 {
		t.Errorf("only the remaining target takes damage, one instance: %d (FR-TGT-003, SVY-R25)", n)
	}
}

func TestCard_CulverinBrigade_TargetGainsWardDealsNothingToIt(t *testing.T) {
	// Covers: FR-TGT-003, FR-KWD-002h
	s := newScene(t)
	warded := cbFoe(s, "Soon Warded", ZoneMainTheater, 5)
	other := cbFoe(s, "Other", ZoneReserves, 5)
	s.scriptPlan(0, nil, nil, pick(warded, other))
	done := false
	s.ptPlaceHook(func(int, StackItem) {
		if !done {
			done = true
			warded.Inst.Card.Keywords.Ward = true
		}
	})
	cbDeploy(s, s.card("Culverin Brigade"))
	if !done {
		t.Fatalf("setup: the trigger should have been placed")
	}
	s.wantDamage(warded, 0)
	s.wantDamage(other, 1)
}

func TestCard_CulverinBrigade_SourceGoneStillDeals(t *testing.T) {
	// Covers: FR-ZON-004, FR-TGT-003
	s := newScene(t)
	foe := cbFoe(s, "Foe", ZoneMainTheater, 5)
	s.scriptPlan(0, nil, nil, pick(foe))
	c := s.card("Culverin Brigade")
	s.ptPlaceHook(func(int, StackItem) {
		if cb := assetOf(s.g, c); cb != nil {
			stormKill(s, cb)
		}
	})
	inst := cbDeploy(s, c)
	if assetOf(s.g, c) != nil {
		t.Fatalf("setup: Culverin Brigade should have left before resolution")
	}
	s.wantDamage(foe, 1)
	if !inDiscardPile(s.g.Players[0], inst) {
		t.Errorf("the card of Culverin Brigade is in its owner's discard pile (FR-ZON-004)")
	}
}

func TestCard_CulverinBrigade_OnlyItsOwnDeployTriggers(t *testing.T) {
	// Covers: FR-TRN-003, FR-ZON-005
	s := newScene(t)
	dummy := s.put(1, ZoneMainTheater, unit("Dummy", 0, 30, ability.Keywords{}))
	brigade := s.put(0, ZoneMainTheater, s.card("Culverin Brigade"))
	s.scriptPlan(0, nil, nil)
	s.do(s.attack(brigade, dummy))
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("attacking is not deploying, so no target is asked: asked %d (FR-TRN-003)", n)
	}
	wantNoNonCombatDamage(t, s.g, "Culverin Brigade has only a deploy trigger")

	s2 := newScene(t)
	foe := cbFoe(s2, "Foe", ZoneMainTheater, 5)
	s2.scriptPlan(0, nil, nil)
	s2.onTop(0, s2.card("Culverin Brigade"))
	inst := s2.give(0, s2.card("Growing Legion"))
	s2.setPool(0, 0, 1)
	s2.do(enactAction(inst))
	if n := faceDownCount(s2.g.Players[0]); n != 1 {
		t.Fatalf("Scratch 1 makes one Scratch Company in player 0's Reserves: %d face-down units (FR-ZON-005)", n)
	}
	if n := len(s2.asked(0)); n != 0 {
		t.Errorf("a face-down Scratch Company has no abilities, so nothing is asked: asked %d (FR-ZON-005)", n)
	}
	s2.wantDamage(foe, 0)
	if got := totalDamage(s2.g); got != 0 {
		t.Errorf("no asset is dealt damage: total %d (FR-ZON-005)", got)
	}
}
