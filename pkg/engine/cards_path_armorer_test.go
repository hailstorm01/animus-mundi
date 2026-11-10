package engine

import "testing"

// Card tests for Broken Horn Armorer (T-0207), written from the CSV rules text
// in data/cards/units-and-structures-path.csv: Path Unit, Devilhorn,
// Specialist, R2 T0, deploys to Reserves, 2/2, "Whenever this unit moves to a
// new zone, chosen asset in that zone gains 1 armor."
// Rules: FR-ZON-002 and SVY-R1 (what counts as moving: a deploy does not),
// SVY-R45 and FR-TGT-001 ("that zone" is the move's destination; no "other"
// and no allegiance), FR-ZON-001 (Reserves are each player's own), FR-TGT-003
// (a target is checked on placement and before resolving), FR-ZON-004 (the
// source leaving does not change the locked zone), FR-TRG-004, FR-LAY-002.
// The helpers tzScript, tzWantOffered, tzWantArmor (that_zone_test.go) and
// phalanxChooser (cards_path_phalanx_test.go) are reused.

// Covers: FR-ZON-002, SVY-R45, FR-TGT-001, FR-LAY-002
func TestCard_BrokenHornArmorer_MarchToMainTheaterArmorsChosenAsset(t *testing.T) {
	s := newScene(t)
	a := s.put(0, ZoneReserves, s.card("Broken Horn Armorer"))
	r := s.put(0, ZoneReserves, unit("R", 2, 4, Keywords{}))
	x := s.put(1, ZoneMainTheater, unit("X", 2, 4, Keywords{}))
	tzScript(s, x.ID)
	s.do(Action{Kind: ActMarch, Player: 0, Asset: a.ID})
	if got := s.g.Asset(a.ID); got == nil || got.Zone != ZoneMainTheater {
		t.Fatalf("precondition: the March moves the Armorer to the Main Theater, got %v", got)
	}
	tzWantOffered(t, s, DecideTarget, "the move's destination is the Main Theater, so the legal targets are its assets (the Armorer and X) and not Reserves unit R (SVY-R45, FR-TGT-001)", a, x)
	tzWantArmor(t, x, 1, "the chosen asset gains 1 armor (FR-LAY-002)")
	tzWantArmor(t, a, 0, "the Armorer was not chosen")
	tzWantArmor(t, r, 0, "R was not chosen and is not in that zone")
	s.wantSettled("an Armorer March to the Main Theater")
}

// Covers: SVY-R45, FR-ZON-001
func TestCard_BrokenHornArmorer_MarchToReservesOffersOwnReserves(t *testing.T) {
	s := newScene(t)
	a := s.put(0, ZoneMainTheater, s.card("Broken Horn Armorer"))
	r := s.put(0, ZoneReserves, unit("R", 2, 4, Keywords{}))
	e := s.put(1, ZoneReserves, unit("E", 2, 4, Keywords{}))
	x := s.put(1, ZoneMainTheater, unit("X", 2, 4, Keywords{}))
	tzScript(s, a.ID)
	s.do(Action{Kind: ActMarch, Player: 0, Asset: a.ID})
	if got := s.g.Asset(a.ID); got == nil || got.Zone != ZoneReserves {
		t.Fatalf("precondition: the March moves the Armorer to Reserves, got %v", got)
	}
	tzWantOffered(t, s, DecideTarget, "a move to Reserves locks its controller's own Reserves: the Armorer and R are offered, not E nor X (SVY-R45, FR-ZON-001)", a, r)
	tzWantArmor(t, a, 1, "the text has no \"other\", so the Armorer may choose itself (SVY-R45)")
	tzWantArmor(t, r, 0, "R was not chosen")
	tzWantArmor(t, e, 0, "E is in the other player's Reserves")
	tzWantArmor(t, x, 0, "X is in the Main Theater")
	s.wantSettled("an Armorer March to Reserves")
}

// Covers: SVY-R1
func TestCard_BrokenHornArmorer_DeployFiresNothing(t *testing.T) {
	s := newScene(t)
	x := s.put(0, ZoneMainTheater, unit("X", 2, 4, Keywords{}))
	s.g.Players[0].Req, s.g.Players[0].Tac = 2, 0
	tzScript(s, x.ID)
	s.deployFromHand(0, s.card("Broken Horn Armorer"))
	if len(s.asked(0)) != 0 {
		t.Errorf("deploying is entering the battlefield, not moving, so nothing is asked: %v (SVY-R1)", kindsOf(s.asked(0)))
	}
	tzWantArmor(t, x, 0, "a deploy fires nothing (SVY-R1)")
	s.wantSettled("an Armorer deploy")
}

// Covers: FR-ZON-002, FR-TRG-004
func TestCard_BrokenHornArmorer_MovedByEffectFires(t *testing.T) {
	s := newScene(t)
	p := s.put(0, ZoneReserves, s.card("Chainbreaker Phalanx"))
	a := s.put(0, ZoneReserves, s.card("Broken Horn Armorer"))
	// First question (Phalanx): only the Armorer is an allied Reserves unit.
	// Second question (Armorer, Main Theater): prefer the Phalanx.
	s.phalanxChooser(yes, p, a)
	s.do(Action{Kind: ActMarch, Player: 0, Asset: p.ID})
	if got := s.g.Asset(a.ID); got == nil || got.Zone != ZoneMainTheater {
		t.Fatalf("precondition: the Phalanx's trigger moves the Armorer to the Main Theater, got %v", got)
	}
	tzWantArmor(t, a, 2, "the Phalanx's armor step still applies to the moved Armorer")
	tzWantArmor(t, p, 1, "a unit moved by an effect fires its own trigger; the chosen Phalanx gains 1 armor (FR-ZON-002, FR-TRG-004)")
	s.wantSettled("a Phalanx move of an Armorer")
}

// Covers: SVY-R45, FR-TGT-003, FR-ZON-004
func TestCard_BrokenHornArmorer_TargetOutlivesSource(t *testing.T) {
	setup := func(t *testing.T) (*scene, *Asset, *Asset) {
		s := newScene(t)
		a := s.put(0, ZoneReserves, s.card("Broken Horn Armorer"))
		x := s.put(1, ZoneMainTheater, unit("X", 2, 4, Keywords{}))
		tzScript(s, x.ID)
		s.g.moveOnBattlefield(a, ZoneMainTheater)
		s.g.flushTriggers()
		return s, a, x
	}
	t.Run("source moves back to Reserves", func(t *testing.T) {
		s, a, x := setup(t)
		s.moveTo(a, ZoneReserves)
		s.g.resolveStack()
		tzWantArmor(t, x, 1, "the second check reads the locked zone, not the source's current zone (FR-TGT-003, SVY-R45)")
	})
	t.Run("source leaves the battlefield", func(t *testing.T) {
		s, a, x := setup(t)
		s.g.leaveBattlefield(a, s.g.effectiveState(a))
		s.g.resolveStack()
		tzWantArmor(t, x, 1, "the locked zone outlives the departed source (FR-ZON-004, SVY-R45)")
	})
}

// Covers: FR-TGT-003
func TestCard_BrokenHornArmorer_TargetLeftZoneIsIllegal(t *testing.T) {
	s := newScene(t)
	a := s.put(0, ZoneReserves, s.card("Broken Horn Armorer"))
	x := s.put(0, ZoneMainTheater, unit("X", 2, 4, Keywords{}))
	o := s.put(1, ZoneMainTheater, unit("O", 2, 4, Keywords{}))
	tzScript(s, x.ID)
	s.g.moveOnBattlefield(a, ZoneMainTheater)
	s.g.flushTriggers()
	s.moveTo(x, ZoneReserves)
	s.g.resolveStack()
	why := "a target that left the locked zone is illegal at the second check and the clause fizzles (FR-TGT-003)"
	tzWantArmor(t, x, 0, why)
	tzWantArmor(t, a, 0, why)
	tzWantArmor(t, o, 0, why)
}
