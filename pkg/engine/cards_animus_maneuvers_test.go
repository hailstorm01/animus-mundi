package engine

import "testing"

// Tests for T-0120 (FR-MNV-001, FR-TGT-002, FR-REP-004): the Animus Maneuvers
// Despoiler's Emergence and Deafening Aria, loaded from data/cards. Player 0
// enacts; player 1 is the opponent.
//
// Despoiler's Emergence: "Deal 4 damage to all opposing assets in the Main
// Theater." / "Deal 2 damage to all opposing non-Scratch assets in opponents'
// Reserves."
// Deafening Aria: "Deal 1 damage to all opposing units in the Main Theater."

// enactFor gives player 0 the named real Maneuver and enough resources, and
// returns the card instance. Req is 3 and Tac is 5 beforehand.
func enactFor(s *scene, name string) *Instance {
	inst := s.give(0, s.card(name))
	p := s.g.Players[0]
	p.Req, p.Tac = 3, 5
	return inst
}

// damageEventsOn returns how many damage events in the log hit asset a.
func damageEventsOn(s *scene, a *Asset) int {
	n := 0
	for _, e := range damageEvents(s.g.Log) {
		if e.Target == a.ID {
			n++
		}
	}
	return n
}

func TestCard_DespoilersEmergence_FourToEveryOpposingAssetInTheMainTheater(t *testing.T) {
	s := newScene(t)
	u1 := s.put(1, ZoneMainTheater, unit("E1", 1, 9, Keywords{}))
	u2 := s.put(1, ZoneMainTheater, unit("E2", 1, 9, Keywords{}))
	st := s.put(1, ZoneMainTheater, structure("EWall", 0, 9, Keywords{}))
	ally := s.put(0, ZoneMainTheater, unit("Ally", 1, 9, Keywords{}))
	allySt := s.put(0, ZoneMainTheater, structure("AWall", 0, 9, Keywords{}))
	s.do(enactAction(enactFor(s, "Despoiler's Emergence")))
	for _, a := range []*Asset{u1, u2, st} {
		s.wantDamage(a, 4)
	}
	s.wantDamage(ally, 0)
	s.wantDamage(allySt, 0)
	if n := damageEventsOn(s, ally) + damageEventsOn(s, allySt); n != 0 {
		t.Errorf("allied assets are not opposing: %d damage events on them, want 0 (FR-MNV-001)", n)
	}
}

func TestCard_DespoilersEmergence_TwoToOpposingNonScratchAssetsInReserves(t *testing.T) {
	s := newScene(t)
	u := s.put(1, ZoneReserves, unit("RUnit", 1, 9, Keywords{}))
	st := s.put(1, ZoneReserves, structure("RWall", 0, 9, Keywords{}))
	scratch := s.putScratch(1)
	allyU := s.put(0, ZoneReserves, unit("AUnit", 1, 9, Keywords{}))
	allyScratch := s.putScratch(0)
	s.do(enactAction(enactFor(s, "Despoiler's Emergence")))
	s.wantDamage(u, 2)
	s.wantDamage(st, 2)
	s.wantDamage(scratch, 0)
	if n := damageEventsOn(s, scratch); n != 0 {
		t.Errorf("a face-down Scratch Company is excluded from the Reserves step: %d damage events, want 0", n)
	}
	s.wantDamage(allyU, 0)
	s.wantDamage(allyScratch, 0)
}

func TestCard_DespoilersEmergence_NoHeadquartersIsDamaged(t *testing.T) {
	s := newScene(t)
	s.put(1, ZoneMainTheater, unit("E1", 1, 9, Keywords{}))
	s.do(enactAction(enactFor(s, "Despoiler's Emergence")))
	for i, p := range s.g.Players {
		s.wantDamage(p.HQ, 0)
		if n := damageEventsOn(s, p.HQ); n != 0 {
			t.Errorf("player %d's Headquarters is in neither zone: %d damage events, want 0", i, n)
		}
	}
}

func TestCard_DespoilersEmergence_ArmorAbsorbsEachInstance(t *testing.T) {
	// Covers: FR-REP-004
	s := newScene(t)
	armored := s.put(1, ZoneMainTheater, unit("Armored", 1, 9, Keywords{Armor: 3}))
	plain := s.put(1, ZoneMainTheater, unit("Plain", 1, 9, Keywords{}))
	s.do(enactAction(enactFor(s, "Despoiler's Emergence")))
	if armored.Armor != 0 {
		t.Errorf("Armor 3 absorbs 3 of the 4 damage first: Armor %d, want 0 (FR-REP-004)", armored.Armor)
	}
	s.wantDamage(armored, 1)
	s.wantDamage(plain, 4)
}

func TestCard_DespoilersEmergence_LethalAssetsDieAfterItResolves(t *testing.T) {
	s := newScene(t)
	mt := s.put(1, ZoneMainTheater, unit("Frail", 1, 4, Keywords{}))
	res := s.put(1, ZoneReserves, unit("Frailer", 1, 2, Keywords{}))
	survivor := s.put(1, ZoneReserves, unit("Sturdy", 1, 3, Keywords{}))
	inst := enactFor(s, "Despoiler's Emergence")
	s.do(enactAction(inst))
	if s.alive(mt) || s.alive(res) {
		t.Errorf("assets left lethal (4 in the Main Theater, 2 in Reserves) die in the sweep after it resolves")
	}
	if !s.alive(survivor) {
		t.Errorf("a 3-health Reserves unit survives 2 damage")
	}
	opp := s.g.Players[1]
	if !inDiscardPile(opp, mt.Inst) || !inDiscardPile(opp, res.Inst) {
		t.Errorf("dead assets go to their owner's discard pile")
	}
	if !inDiscardPile(s.g.Players[0], inst) {
		t.Errorf("the resolved Maneuver ends in its owner's discard pile (FR-MNV-001)")
	}
}

func TestCard_DespoilersEmergence_EnactableWithNoOpposingAsset(t *testing.T) {
	// Covers: SVY-R29
	s := newScene(t)
	ally := s.put(0, ZoneMainTheater, unit("Ally", 1, 9, Keywords{}))
	inst := enactFor(s, "Despoiler's Emergence")
	if !offersEnact(s.g, inst) {
		t.Fatalf("a Maneuver with no target slot is a legal Enact with no opposing asset (SVY-R29)")
	}
	s.do(enactAction(inst))
	s.wantDamage(ally, 0)
	if n := len(damageEvents(s.g.Log)); n != 0 {
		t.Errorf("nothing matched, so nothing is damaged: %d damage events, want 0", n)
	}
	if lastDiscard(s.g.Players[0]) != inst {
		t.Errorf("the Maneuver goes to its owner's discard pile (FR-MNV-001)")
	}
}

func TestCard_DespoilersEmergence_AsksNothing(t *testing.T) {
	// Covers: FR-TGT-002
	s := newScene(t)
	s.put(1, ZoneMainTheater, unit("E1", 1, 9, Keywords{Ward: true}))
	s.put(1, ZoneReserves, unit("E2", 1, 9, Keywords{}))
	s.script(0)
	s.script(1)
	s.do(enactAction(enactFor(s, "Despoiler's Emergence")))
	for p := PlayerID(0); p < 2; p++ {
		if got := s.asked(p); len(got) != 0 {
			t.Errorf("\"all\" is not targeted: player %d was asked %d decisions, want none (FR-TGT-002)", p, len(got))
		}
	}
}

func TestCard_DespoilersEmergence_PaysFourTactics(t *testing.T) {
	s := newScene(t)
	s.do(enactAction(enactFor(s, "Despoiler's Emergence")))
	p := s.g.Players[0]
	if p.Tac != 1 || p.Req != 3 {
		t.Errorf("enacting costs 4 Tactics and no Requisition: pool R%d T%d, want R3 T1 (FR-MNV-001)", p.Req, p.Tac)
	}
}

func TestCard_DeafeningAria_OneToEveryOpposingUnitInTheMainTheater(t *testing.T) {
	s := newScene(t)
	u1 := s.put(1, ZoneMainTheater, unit("E1", 1, 9, Keywords{}))
	u2 := s.put(1, ZoneMainTheater, unit("E2", 1, 9, Keywords{}))
	st := s.put(1, ZoneMainTheater, structure("EWall", 0, 9, Keywords{}))
	res := s.put(1, ZoneReserves, unit("ERes", 1, 9, Keywords{}))
	ally := s.put(0, ZoneMainTheater, unit("Ally", 1, 9, Keywords{}))
	s.do(enactAction(enactFor(s, "Deafening Aria")))
	s.wantDamage(u1, 1)
	s.wantDamage(u2, 1)
	s.wantDamage(st, 0)
	s.wantDamage(res, 0)
	s.wantDamage(ally, 0)
	for _, a := range []*Asset{st, res, ally, s.g.Players[1].HQ} {
		if n := damageEventsOn(s, a); n != 0 {
			t.Errorf("%s is not an opposing Main Theater unit: %d damage events, want 0", a, n)
		}
	}
	if p := s.g.Players[0]; p.Tac != 3 || p.Req != 3 {
		t.Errorf("enacting costs 2 Tactics: pool R%d T%d, want R3 T3", p.Req, p.Tac)
	}
}

func TestCard_DeafeningAria_KillsOneHealthUnits(t *testing.T) {
	s := newScene(t)
	one := s.put(1, ZoneMainTheater, unit("OneHealth", 1, 1, Keywords{}))
	two := s.put(1, ZoneMainTheater, unit("TwoHealth", 1, 2, Keywords{}))
	inst := enactFor(s, "Deafening Aria")
	s.do(enactAction(inst))
	if s.alive(one) {
		t.Errorf("a 1-health unit left lethal dies in the sweep after the Maneuver resolves")
	}
	if !inDiscardPile(s.g.Players[1], one.Inst) {
		t.Errorf("the dead unit goes to its owner's discard pile")
	}
	if !s.alive(two) {
		t.Errorf("a 2-health unit survives 1 damage")
	}
	if !inDiscardPile(s.g.Players[0], inst) {
		t.Errorf("the resolved Maneuver ends in its owner's discard pile (FR-MNV-001)")
	}
}
