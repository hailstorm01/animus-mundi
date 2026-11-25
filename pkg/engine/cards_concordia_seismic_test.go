package engine

import (
	"animus-mundi/pkg/engine/ability"
	"testing"
)

// Tests for T-0302 (FR-MNV-001, FR-TGT-002, FR-REP-004, FR-LAY-002): the
// Concordia Maneuver Seismic Instability, loaded from data/cards. Player 0
// enacts; player 1 is the opponent.
//
// Seismic Instability: "Deal 3 damage to all assets in the Main Theater." /
// "Deal 4 damage to all structures in the Main Theater."

func TestCard_SeismicInstability_ThreeToEveryUnitInTheMainTheater(t *testing.T) {
	// Covers: FR-MNV-001
	s := newScene(t)
	enemy := s.put(1, ZoneMainTheater, unit("E1", 1, 9, ability.Keywords{}))
	ally := s.put(0, ZoneMainTheater, unit("Ally", 1, 9, ability.Keywords{}))
	s.do(enactAction(enactFor(s, "Seismic Instability")))
	for _, a := range []*Asset{enemy, ally} {
		s.wantDamage(a, 3)
		if n := damageEventsOn(s, a); n != 1 {
			t.Errorf("a unit is hit by the first line only: %d damage events, want 1 (FR-MNV-001)", n)
		}
	}
}

func TestCard_SeismicInstability_StructuresTakeBothInstances(t *testing.T) {
	// Covers: FR-REP-004
	s := newScene(t)
	enemy := s.put(1, ZoneMainTheater, structure("EWall", 0, 20, ability.Keywords{}))
	ally := s.put(0, ZoneMainTheater, structure("AWall", 0, 20, ability.Keywords{}))
	s.do(enactAction(enactFor(s, "Seismic Instability")))
	for _, a := range []*Asset{enemy, ally} {
		s.wantDamage(a, 7)
		if n := damageEventsOn(s, a); n != 2 {
			t.Errorf("a structure takes 3 and then 4 as two instances: %d damage events, want 2 (FR-REP-004)", n)
		}
	}
}

func TestCard_SeismicInstability_ReservesAndHeadquartersUntouched(t *testing.T) {
	// Covers: FR-TGT-002
	s := newScene(t)
	var untouched []*Asset
	for p := PlayerID(0); p < 2; p++ {
		untouched = append(untouched,
			s.put(p, ZoneReserves, unit("RUnit", 1, 9, ability.Keywords{})),
			s.put(p, ZoneReserves, structure("RWall", 0, 9, ability.Keywords{})),
			s.putScratch(p),
			s.g.Players[p].HQ)
	}
	s.put(1, ZoneMainTheater, unit("E1", 1, 9, ability.Keywords{}))
	s.do(enactAction(enactFor(s, "Seismic Instability")))
	for _, a := range untouched {
		s.wantDamage(a, 0)
		if n := damageEventsOn(s, a); n != 0 {
			t.Errorf("%s is in Reserves or a Headquarters, not the Main Theater: %d damage events, want 0 (FR-MNV-001)", a, n)
		}
	}
}

func TestCard_SeismicInstability_ScratchCompanyInTheMainTheaterIsAUnit(t *testing.T) {
	// Covers: FR-ZON-005
	s := newScene(t)
	sc := s.putScratch(1)
	sc.Inst.Card = structure("Hidden Wall", 0, 9, ability.Keywords{})
	s.moveTo(sc, ZoneMainTheater)
	s.do(enactAction(enactFor(s, "Seismic Instability")))
	if n := damageEventsOn(s, sc); n != 1 {
		t.Errorf("a face-down Scratch Company is a unit whatever its hidden card: %d damage events, want 1 (FR-ZON-005)", n)
	}
	if s.alive(sc) {
		t.Errorf("a 1-health Scratch Company takes 3 and dies in the sweep")
	}
}

func TestCard_SeismicInstability_ResistantAppliesToEachInstance(t *testing.T) {
	// Covers: FR-KWD-002b
	s := newScene(t)
	wall := s.put(1, ZoneMainTheater, structure("Resistant Wall", 0, 20, ability.Keywords{Resistant: 1}))
	soft := s.put(1, ZoneMainTheater, unit("Soft", 1, 20, ability.Keywords{Resistant: -1}))
	s.do(enactAction(enactFor(s, "Seismic Instability")))
	s.wantDamage(wall, 5) // 3-1 and 4-1
	s.wantDamage(soft, 4) // 3+1
}

func TestCard_SeismicInstability_ArmorAbsorbsAcrossBothInstances(t *testing.T) {
	// Covers: FR-REP-004, FR-KWD-002k
	s := newScene(t)
	wall := s.put(1, ZoneMainTheater, structure("Armored Wall", 0, 20, ability.Keywords{Armor: 5}))
	s.do(enactAction(enactFor(s, "Seismic Instability")))
	if wall.Armor != 0 {
		t.Errorf("Armor 5 absorbs the first 3 and 2 of the second 4: Armor %d, want 0 (FR-KWD-002k)", wall.Armor)
	}
	s.wantDamage(wall, 2)
}

func TestCard_SeismicInstability_LethalStructureStillTakesTheSecondInstance(t *testing.T) {
	// Covers: FR-LAY-002
	s := newScene(t)
	wall := s.put(1, ZoneMainTheater, structure("Frail Wall", 0, 3, ability.Keywords{}))
	dies := s.put(1, ZoneMainTheater, unit("Frail", 1, 3, ability.Keywords{}))
	lives := s.put(1, ZoneMainTheater, unit("Sturdy", 1, 4, ability.Keywords{}))
	inst := enactFor(s, "Seismic Instability")
	s.do(enactAction(inst))
	if n := damageEventsOn(s, wall); n != 2 {
		t.Errorf("death is swept only after the Maneuver finishes, so the structure takes both instances: %d damage events, want 2 (FR-LAY-002)", n)
	}
	if s.alive(wall) || s.alive(dies) {
		t.Errorf("a 3-health structure and a 3-health unit die in the sweep after it resolves")
	}
	if !s.alive(lives) {
		t.Errorf("a 4-health unit survives 3 damage")
	}
	opp := s.g.Players[1]
	if !inDiscardPile(opp, wall.Inst) || !inDiscardPile(opp, dies.Inst) {
		t.Errorf("dead assets go to their owner's discard pile")
	}
	if !inDiscardPile(s.g.Players[0], inst) {
		t.Errorf("the resolved Maneuver ends in its owner's discard pile (FR-MNV-001)")
	}
}

func TestCard_SeismicInstability_UnbowedStructureTriggersTwice(t *testing.T) {
	// Covers: FR-KWD-002a
	s := newScene(t)
	wallCard := structure("Bowed Wall", 0, 20, ability.Keywords{})
	wallCard.Ability = &ability.Def{Triggers: []ability.Trigger{unbowed(nil, scratchN(1))}}
	s.put(1, ZoneMainTheater, wallCard)
	s.put(0, ZoneMainTheater, bowed("Bowed Unit", 1, 20, ability.Keywords{}, unbowed(nil, scratchN(1))))
	s.do(enactAction(enactFor(s, "Seismic Instability")))
	s.wantScratches(1, 2, "the structure takes two damage instances, so its Unbowed triggers twice (FR-KWD-002a)")
	s.wantScratches(0, 1, "the unit takes one damage instance, so its Unbowed triggers once (FR-KWD-002a)")
}

func TestCard_SeismicInstability_AsksNothing(t *testing.T) {
	// Covers: FR-TGT-002, FR-KWD-002h
	s := newScene(t)
	ward := s.put(1, ZoneMainTheater, unit("Warded", 1, 9, ability.Keywords{Ward: true}))
	s.script(0)
	s.script(1)
	s.do(enactAction(enactFor(s, "Seismic Instability")))
	for p := PlayerID(0); p < 2; p++ {
		if got := s.asked(p); len(got) != 0 {
			t.Errorf("\"all\" is not targeted: player %d was asked %d decisions, want none (FR-TGT-002)", p, len(got))
		}
	}
	s.wantDamage(ward, 3)
}

func TestCard_SeismicInstability_EnactableWithAnEmptyMainTheater(t *testing.T) {
	// Covers: SVY-R29
	s := newScene(t)
	res := s.put(1, ZoneReserves, unit("RUnit", 1, 9, ability.Keywords{}))
	inst := enactFor(s, "Seismic Instability")
	if !offersEnact(s.g, inst) {
		t.Fatalf("a Maneuver with no target slot is a legal Enact with no asset in the Main Theater (SVY-R29)")
	}
	s.do(enactAction(inst))
	s.wantDamage(res, 0)
	if n := len(damageEvents(s.g.Log)); n != 0 {
		t.Errorf("nothing matched, so nothing is damaged: %d damage events, want 0", n)
	}
	if lastDiscard(s.g.Players[0]) != inst {
		t.Errorf("the Maneuver goes to its owner's discard pile (FR-MNV-001)")
	}
}

func TestCard_SeismicInstability_PaysFourTactics(t *testing.T) {
	// Covers: FR-MNV-001
	s := newScene(t)
	s.do(enactAction(enactFor(s, "Seismic Instability")))
	p := s.g.Players[0]
	if p.Tac != 1 || p.Req != 3 {
		t.Errorf("enacting costs 4 Tactics and no Requisition: pool R%d T%d, want R3 T1 (FR-MNV-001)", p.Req, p.Tac)
	}
	s2 := newScene(t)
	inst := enactFor(s2, "Seismic Instability")
	s2.g.Players[0].Req, s2.g.Players[0].Tac = 5, 3
	if offersEnact(s2.g, inst) {
		t.Errorf("with 3 Tactics the Maneuver costing 4 is not offered (FR-MNV-001)")
	}
}
