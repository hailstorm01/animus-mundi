// Tests for T-0310 (FR-MNV-001, FR-TGT-001, FR-TGT-003, FR-REP-004, FR-LAY-002,
// FR-KWD-002h, FR-KWD-002k, FR-ZON-005, FR-ENT-103, FR-CBT-003): the Soulless
// Maneuver Bring It Down, loaded from data/cards. Player 0 enacts; player 1 is
// the opponent.
//
// Bring It Down: "Deal 2 damage to chosen non-Scratch asset." Cost 1 Tactics.
package engine

import (
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// bidEnact gives player 0 Bring It Down with exactly R0 T1.
func bidEnact(s *scene) *Instance {
	s.t.Helper()
	inst := s.give(0, s.card("Bring It Down"))
	s.setPool(0, 0, 1)
	return inst
}

// bidAsk enacts Bring It Down, picking the first option, and returns the one
// target question put to seat 0.
func bidAsk(t *testing.T, s *scene, inst *Instance) Decision {
	t.Helper()
	s.script(0, Answer{Picks: []int{0}})
	s.do(enactAction(inst))
	return targetAsk(t, s)
}

func TestCard_BringItDown_DealsTwoToTheChosenAsset(t *testing.T) {
	// Covers: FR-MNV-001, FR-LAY-002
	s := newScene(t)
	hit := s.put(1, ZoneMainTheater, unit("Hit", 1, 5, ability.Keywords{}))
	other := s.put(1, ZoneMainTheater, unit("Other", 1, 5, ability.Keywords{}))
	inst := bidEnact(s)
	chooseAsset(s, hit)
	s.do(enactAction(inst))
	s.wantDamage(hit, 2)
	if got := s.g.CurrentHealth(hit); got != 3 {
		t.Errorf("a health 5 unit with 2 damage has health 3, got %d (FR-LAY-002)", got)
	}
	s.wantDamage(other, 0)
	if p := s.g.Players[0]; p.Req != 0 || p.Tac != 0 {
		t.Errorf("the 1 Tactics is paid: pool R%d T%d, want R0 T0 (FR-MNV-001)", p.Req, p.Tac)
	}
	if !inDiscardPile(s.g.Players[0], inst) {
		t.Errorf("Bring It Down ends in its owner's discard pile (FR-MNV-001)")
	}
}

func TestCard_BringItDown_OffersEveryNonScratchAsset(t *testing.T) {
	// Covers: FR-TGT-001, SVY-R29
	s := newScene(t)
	aRes := s.put(0, ZoneReserves, unit("AllyRes", 1, 5, ability.Keywords{}))
	aMain := s.put(0, ZoneMainTheater, unit("AllyMain", 1, 5, ability.Keywords{}))
	fRes := s.put(1, ZoneReserves, unit("FoeRes", 1, 5, ability.Keywords{}))
	fMain := s.put(1, ZoneMainTheater, unit("FoeMain", 1, 5, ability.Keywords{}))
	wall := s.put(1, ZoneMainTheater, structure("FWall", 0, 5, ability.Keywords{}))
	s.putScratch(0)
	s.putScratch(1)
	inst := bidEnact(s)
	d := bidAsk(t, s, inst)
	wantOptions(t, d, "the four units, the structure and both Headquarters (Scratch Companies excluded)",
		aRes, aMain, fRes, fMain, wall, s.g.Players[0].HQ, s.g.Players[1].HQ)
}

func TestCard_BringItDown_UnitTurnedFaceDownNotOffered(t *testing.T) {
	// Covers: FR-ENT-103, FR-TGT-001
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Foe", 1, 5, ability.Keywords{}))
	morphEnact(s, morphManeuver())
	if a := s.g.Asset(foe.ID); a == nil || !a.FaceDown {
		t.Fatalf("setup: the opposing unit should be face down (FR-ENT-103)")
	}
	up := s.put(1, ZoneMainTheater, unit("FaceUp", 1, 5, ability.Keywords{}))
	inst := bidEnact(s)
	d := bidAsk(t, s, inst)
	wantOptions(t, d, "a face-down unit is a Scratch Company and not offered; the face-up unit and both Headquarters are",
		up, s.g.Players[0].HQ, s.g.Players[1].HQ)
}

func TestCard_BringItDown_EnactableWithOnlyHeadquarters(t *testing.T) {
	// Covers: SVY-R29, FR-MNV-001
	s := newScene(t)
	s.putScratch(0)
	s.putScratch(1)
	inst := bidEnact(s)
	if !offersEnact(s.g, inst) {
		t.Fatalf("a Headquarters is a legal non-Scratch asset, so the Enact is offered (SVY-R29)")
	}
	own := s.g.Players[0].HQ
	chooseAsset(s, own)
	s.do(enactAction(inst))
	wantOptions(t, targetAsk(t, s), "only the two Headquarters", s.g.Players[0].HQ, s.g.Players[1].HQ)
	s.wantDamage(own, 2)
}

func TestCard_BringItDown_DamagesTheOpposingHeadquarters(t *testing.T) {
	// Covers: FR-CBT-003
	s := newScene(t)
	hq1 := s.g.Players[1].HQ
	before := s.g.CurrentHealth(hq1)
	inst := bidEnact(s)
	chooseAsset(s, hq1)
	s.do(enactAction(inst))
	if got := s.g.CurrentHealth(hq1); got != before-2 {
		t.Errorf("the opposing Headquarters remaining Health drops by 2: got %d, want %d (FR-CBT-003)", got, before-2)
	}
	if s.g.Outcome.Over {
		t.Errorf("the match is not over (FR-CBT-003)")
	}
}

func TestCard_BringItDown_LethalDamageKills(t *testing.T) {
	// Covers: FR-LAY-002
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Frail", 1, 2, ability.Keywords{}))
	inst := bidEnact(s)
	chooseAsset(s, foe)
	s.do(enactAction(inst))
	if s.alive(foe) {
		t.Errorf("a unit with health 2 takes 2 damage and is destroyed (FR-LAY-002)")
	}
	if !inDiscardPile(s.g.Players[1], foe.Inst) {
		t.Errorf("the destroyed unit's card is in its owner's discard pile (FR-LAY-002)")
	}
}

func TestCard_BringItDown_ResistantAndArmorApply(t *testing.T) {
	// Covers: FR-REP-004, FR-KWD-002k
	s := newScene(t)
	res := s.put(1, ZoneMainTheater, unit("Resistant", 1, 5, ability.Keywords{Resistant: 1}))
	inst := bidEnact(s)
	chooseAsset(s, res)
	s.do(enactAction(inst))
	s.wantDamage(res, 1)

	s2 := newScene(t)
	arm := s2.put(1, ZoneMainTheater, unit("Armored", 1, 5, ability.Keywords{Armor: 1}))
	inst2 := bidEnact(s2)
	chooseAsset(s2, arm)
	s2.do(enactAction(inst2))
	if arm.Armor != 0 {
		t.Errorf("the first point of damage removes the 1 Armor: Armor %d, want 0 (FR-KWD-002k)", arm.Armor)
	}
	s2.wantDamage(arm, 1)
}

func TestCard_BringItDown_WardedOpposingAssetNotOffered(t *testing.T) {
	// Covers: FR-KWD-002h
	s := newScene(t)
	foeWard := s.put(1, ZoneMainTheater, unit("FoeWard", 1, 5, ability.Keywords{Ward: true}))
	allyWard := s.put(0, ZoneMainTheater, unit("AllyWard", 1, 5, ability.Keywords{Ward: true}))
	inst := bidEnact(s)
	d := bidAsk(t, s, inst)
	var foeOffered, allyOffered bool
	for _, o := range d.Options {
		foeOffered = foeOffered || o.Ref == foeWard.ID
		allyOffered = allyOffered || o.Ref == allyWard.ID
	}
	if foeOffered {
		t.Errorf("an opposing Warded unit cannot be chosen (FR-KWD-002h)")
	}
	if !allyOffered {
		t.Errorf("an allied Warded unit can be chosen: Ward stops only opponents (FR-KWD-002h)")
	}
}

func TestCard_BringItDown_TargetGoneDealsNothing(t *testing.T) {
	// Covers: FR-TGT-003
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Fleeting", 1, 5, ability.Keywords{}))
	other := s.put(1, ZoneMainTheater, unit("Other", 1, 5, ability.Keywords{}))
	inst := bidEnact(s)
	chooseAsset(s, foe)
	s.ptPlaceHook(func(int, StackItem) {
		if s.alive(foe) {
			stormKill(s, foe)
		}
	})
	s.do(enactAction(inst))
	if s.alive(foe) {
		t.Fatalf("setup: the chosen unit should have left before resolution")
	}
	if got := totalDamage(s.g); got != 0 {
		t.Errorf("the target is gone, so nothing is dealt damage: total %d (FR-TGT-003)", got)
	}
	s.wantDamage(other, 0)
	if !inDiscardPile(s.g.Players[0], inst) {
		t.Errorf("Bring It Down still ends in its owner's discard pile (FR-TGT-003)")
	}
}

func TestCard_BringItDown_PaysOneTactics(t *testing.T) {
	// Covers: FR-MNV-001
	s := newScene(t)
	inst := s.give(0, s.card("Bring It Down"))
	s.setPool(0, 0, 0)
	if offersEnact(s.g, inst) {
		t.Errorf("with R0 T0 the Enact is not offered (FR-MNV-001)")
	}
	s.setPool(0, 0, 1)
	if !offersEnact(s.g, inst) {
		t.Fatalf("with R0 T1 the Enact is offered (FR-MNV-001)")
	}
	chooseAsset(s, s.g.Players[1].HQ)
	s.do(enactAction(inst))
	if p := s.g.Players[0]; p.Req != 0 || p.Tac != 0 {
		t.Errorf("the pool is empty after it resolves: R%d T%d (FR-MNV-001)", p.Req, p.Tac)
	}
}
