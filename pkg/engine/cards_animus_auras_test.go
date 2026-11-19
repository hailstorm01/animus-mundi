package engine

import (
	"animus-mundi/pkg/engine/ability"
	"testing"
)

// Tests for T-0259 (Hulking Incongruity) and T-0257, the same-zone power auras Animoceps Stalk, Animoceps Node
// and Melodic Pummeler (FR-LAY-001, FR-TGT-001 SVY-R44, FR-ZON-001). Player 0
// is the active player. Expected power is printed power plus the aura.

// auraSub is a unit with the given subtypes.
func auraSub(name string, power int, subs ...string) *Card {
	c := unit(name, power, 5, ability.Keywords{})
	c.Subtypes = subs
	return c
}

func TestCard_AnimocepsStalk_OtherAlliedAssetsInItsZone(t *testing.T) {
	// Covers: FR-LAY-001
	s := newScene(t)
	stalk := s.put(0, ZoneMainTheater, s.card("Animoceps Stalk"))
	u := s.put(0, ZoneMainTheater, unit("Ally", 2, 5, ability.Keywords{}))
	st := s.put(0, ZoneMainTheater, structure("Wall", 0, 5, ability.Keywords{}))
	res := s.put(0, ZoneReserves, unit("Reserve Ally", 2, 5, ability.Keywords{}))
	foe := s.put(1, ZoneMainTheater, unit("Foe", 2, 5, ability.Keywords{}))
	s.wantPower(u, 3, "an allied unit in the Stalk's zone has +1 power")
	s.wantPower(st, 1, "an allied structure in the Stalk's zone has +1 power (any asset)")
	s.wantPower(stalk, 1, "the Stalk says other, so it does not buff itself")
	s.wantPower(res, 2, "an allied unit in Reserves is not in the Stalk's zone")
	s.wantPower(foe, 2, "an opposing unit is not allied")
}

func TestCard_AnimocepsStalk_TwoStalksBuffEachOther(t *testing.T) {
	// Covers: FR-LAY-001
	s := newScene(t)
	a := s.put(0, ZoneMainTheater, s.card("Animoceps Stalk"))
	b := s.put(0, ZoneMainTheater, s.card("Animoceps Stalk"))
	s.wantPower(a, 2, "the other Stalk buffs this one")
	s.wantPower(b, 2, "the other Stalk buffs this one")
}

func TestCard_AnimocepsStalk_FollowsItsZone(t *testing.T) {
	// Covers: FR-TGT-001, FR-ZON-001 (SVY-R44, last sentence)
	s := newScene(t)
	stalk := s.put(0, ZoneMainTheater, s.card("Animoceps Stalk"))
	a := s.put(0, ZoneMainTheater, unit("A", 2, 5, ability.Keywords{}))
	b := s.put(0, ZoneReserves, unit("B", 2, 5, ability.Keywords{}))
	s.moveTo(stalk, ZoneReserves)
	foe := s.put(1, ZoneReserves, unit("Foe", 2, 5, ability.Keywords{}))
	s.wantPower(a, 2, "the Stalk left the Main Theater")
	s.wantPower(b, 3, "the Stalk is now in player 0's Reserves")
	s.wantPower(foe, 2, "player 1's Reserves are not player 0's Reserves")
}

func TestCard_AnimocepsNode_AlliedUnitsInItsZone(t *testing.T) {
	// Covers: FR-LAY-001
	s := newScene(t)
	node := s.put(0, ZoneMainTheater, s.card("Animoceps Node"))
	u := s.put(0, ZoneMainTheater, unit("Ally", 2, 5, ability.Keywords{}))
	scratch := s.putScratch(0)
	s.moveTo(scratch, ZoneMainTheater)
	st := s.put(0, ZoneMainTheater, structure("Wall", 1, 5, ability.Keywords{}))
	res := s.put(0, ZoneReserves, unit("Reserve Ally", 2, 5, ability.Keywords{}))
	foe := s.put(1, ZoneMainTheater, unit("Foe", 2, 5, ability.Keywords{}))
	s.wantPower(u, 3, "an allied unit in the Node's zone has +1 power")
	s.wantPower(scratch, 2, "a face-down Scratch Company (power 1) is a unit")
	s.wantPower(node, 3, "the Node is a structure, so it does not buff itself")
	s.wantPower(st, 1, "an allied structure is not a unit")
	s.wantPower(res, 2, "an allied unit in Reserves is not in the Node's zone")
	s.wantPower(foe, 2, "an opposing unit is not allied")
}

func TestCard_AnimocepsNode_FollowsItsZone(t *testing.T) {
	// Covers: FR-TGT-001 (SVY-R44)
	s := newScene(t)
	node := s.put(0, ZoneMainTheater, s.card("Animoceps Node"))
	main := s.put(0, ZoneMainTheater, unit("Main", 2, 5, ability.Keywords{}))
	res := s.put(0, ZoneReserves, unit("Res", 2, 5, ability.Keywords{}))
	s.moveTo(node, ZoneReserves)
	s.wantPower(res, 3, "the Node is now in Reserves")
	s.wantPower(main, 2, "the Node left the Main Theater")
}

func TestCard_MelodicPummeler_AlliedHorrorUnitsInItsZone(t *testing.T) {
	// Covers: FR-LAY-001
	s := newScene(t)
	pum := s.put(0, ZoneReserves, s.card("Melodic Pummeler"))
	h := s.put(0, ZoneReserves, auraSub("Horror", 3, "Horror"))
	multi := s.put(0, ZoneReserves, auraSub("Treefolk Horror", 3, "Treefolk", "Horror"))
	plain := s.put(0, ZoneReserves, auraSub("Plain", 3, "Treefolk"))
	hs := structure("Horror Wall", 3, 5, ability.Keywords{})
	hs.Subtypes = []string{"Horror"}
	hstruct := s.put(0, ZoneReserves, hs)
	mainH := s.put(0, ZoneMainTheater, auraSub("Main Horror", 3, "Horror"))
	foeH := s.put(1, ZoneReserves, auraSub("Foe Horror", 3, "Horror"))
	s.wantPower(pum, 5, "the Pummeler is a Horror unit in its own zone, so it buffs itself")
	s.wantPower(h, 4, "an allied Horror unit in the zone has +1 power")
	s.wantPower(multi, 4, "any one matching subtype is enough")
	s.wantPower(plain, 3, "a non-Horror unit gains nothing")
	s.wantPower(hstruct, 3, "a Horror structure is not a unit")
	s.wantPower(mainH, 3, "a Horror unit in another zone gains nothing")
	s.wantPower(foeH, 3, "an opposing Horror unit gains nothing")
}

func TestCard_MelodicPummeler_FollowsItsZone(t *testing.T) {
	// Covers: FR-TGT-001 (SVY-R44)
	s := newScene(t)
	pum := s.put(0, ZoneReserves, s.card("Melodic Pummeler"))
	mainH := s.put(0, ZoneMainTheater, auraSub("Main Horror", 3, "Horror"))
	resH := s.put(0, ZoneReserves, auraSub("Res Horror", 3, "Horror"))
	s.moveTo(pum, ZoneMainTheater)
	s.wantPower(mainH, 4, "the Pummeler is now in the Main Theater")
	s.wantPower(resH, 3, "the Pummeler left Reserves")
	s.wantPower(pum, 5, "the Pummeler still buffs itself")
}

func TestCard_HulkingIncongruity_AlliedHorrorUnitsInItsZone(t *testing.T) {
	// Covers: FR-LAY-001
	s := newScene(t)
	inc := s.put(0, ZoneReserves, s.card("Hulking Incongruity"))
	h := s.put(0, ZoneReserves, auraSub("Horror", 3, "Horror"))
	multi := s.put(0, ZoneReserves, auraSub("Treefolk Horror", 3, "Treefolk", "Horror"))
	plain := s.put(0, ZoneReserves, auraSub("Plain", 3, "Treefolk"))
	hs := structure("Horror Wall", 3, 5, ability.Keywords{})
	hs.Subtypes = []string{"Horror"}
	hstruct := s.put(0, ZoneReserves, hs)
	mainH := s.put(0, ZoneMainTheater, auraSub("Main Horror", 3, "Horror"))
	foeH := s.put(1, ZoneReserves, auraSub("Foe Horror", 3, "Horror"))
	s.wantResistant(inc, 1, "the Incongruity is a Horror unit in its own zone, so it covers itself")
	s.wantResistant(h, 1, "an allied Horror unit in the zone has Resistant 1")
	s.wantResistant(multi, 1, "any one matching subtype is enough")
	s.wantResistant(plain, 0, "a non-Horror unit gains nothing")
	s.wantResistant(hstruct, 0, "a Horror structure is not a unit")
	s.wantResistant(mainH, 0, "a Horror unit in another zone gains nothing")
	s.wantResistant(foeH, 0, "an opposing Horror unit gains nothing")
	s.wantPower(inc, 2, "the Incongruity keeps its printed power 2")
	s.wantPower(h, 3, "the grant adds Resistant, not power")
}

func TestCard_HulkingIncongruity_AddsToPrintedResistant(t *testing.T) {
	// Covers: FR-REP-004
	s := newScene(t)
	s.put(0, ZoneReserves, s.card("Hulking Incongruity"))
	h := s.put(0, ZoneReserves, horrorUnit("Armoured Horror", 3, 5, 2))
	s.wantResistant(h, 3, "printed Resistant 2 plus the aura's 1 is the total subtracted in Stage 2")
}

func TestCard_HulkingIncongruity_FollowsItsZone(t *testing.T) {
	// Covers: FR-TGT-001 (SVY-R44)
	s := newScene(t)
	inc := s.put(0, ZoneReserves, s.card("Hulking Incongruity"))
	mainH := s.put(0, ZoneMainTheater, auraSub("Main Horror", 3, "Horror"))
	resH := s.put(0, ZoneReserves, auraSub("Res Horror", 3, "Horror"))
	s.moveTo(inc, ZoneMainTheater)
	s.wantResistant(mainH, 1, "the Incongruity is now in the Main Theater")
	s.wantResistant(resH, 0, "the Incongruity left Reserves")
	s.wantResistant(inc, 1, "the Incongruity still covers itself")
}

func TestCard_HulkingIncongruity_CoveredUnitTakesOneLess(t *testing.T) {
	// Covers: FR-KWD-002b, FR-REP-004
	s := newScene(t)
	s.put(1, ZoneMainTheater, s.card("Hulking Incongruity"))
	d := s.put(1, ZoneMainTheater, horrorUnit("D", 0, 5, 0))
	att := s.put(0, ZoneMainTheater, unit("Attacker", 3, 5, ability.Keywords{}))
	s.do(s.attack(att, d))
	s.wantDamage(d, 2)
}
