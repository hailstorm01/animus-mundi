package engine

import "testing"

// Card tests for Ancestral Monument (T-0263), written from the CSV rules text in
// data/cards/units-and-structures-path.csv:22: Path Structure (Tower),
// Requisition 4, Tactics 0, Main Theater, power 0, Health 9,
// "Other allied assets in the same zone have Cornered: Resistant 1."
// Rules: FR-KWD-002e (Cornered, SVY-R18), FR-LAY-001 (continuous, dynamic),
// FR-LAY-003 (the lock), FR-KWD-002b and FR-REP-004 Stage 2 (Resistant),
// SVY-R44 (a continuous ability uses its source's current zone).

// Covers: FR-KWD-002e, FR-LAY-001, SVY-R18
func TestCard_AncestralMonument_OtherAlliedAssetsInItsZoneWhileCornered(t *testing.T) {
	s := newScene(t)
	mon := s.put(0, ZoneMainTheater, s.card("Ancestral Monument"))
	u := s.put(0, ZoneMainTheater, unit("U", 2, 5, Keywords{}))
	st := s.put(0, ZoneMainTheater, structure("St", 0, 5, Keywords{}))
	r := s.put(0, ZoneReserves, unit("R", 2, 5, Keywords{}))
	foes := s.foes(3)

	s.wantResistant(u, 0, "3 against 3 is not Cornered, so the grant is not active")
	s.wantResistant(st, 0, "3 against 3 is not Cornered, so the grant is not active")

	extra := s.put(1, ZoneMainTheater, unit("Extra Foe", 0, 5, Keywords{}))
	s.wantResistant(u, 1, "3 against 4: Cornered, an other allied unit in the zone has Resistant 1")
	s.wantResistant(st, 1, "3 against 4: Cornered, an other allied structure in the zone has Resistant 1")
	s.wantResistant(r, 0, "an allied unit in another zone is not matched")
	s.wantResistant(mon, 0, "Other: the Monument does not grant itself")
	for _, f := range append(foes, extra) {
		s.wantResistant(f, 0, "a foe is not allied")
	}
	s.wantPower(mon, 0, "the grant adds no power to the Monument")
	s.wantPower(u, 2, "the grant adds no power")

	s.moveTo(extra, ZoneReserves)
	s.wantResistant(u, 0, "3 against 3: no longer Cornered, so the grant ends")
	s.wantResistant(st, 0, "3 against 3: no longer Cornered, so the grant ends")
	s.wantPower(mon, 0, "the grant adds no power")
	s.wantPower(u, 2, "the grant adds no power")
}

// Covers: FR-REP-004, FR-KWD-002b, FR-KWD-002e
func TestCard_AncestralMonument_AddsToPrintedResistant(t *testing.T) {
	s := newScene(t)
	s.put(0, ZoneMainTheater, s.card("Ancestral Monument"))
	a := s.put(0, ZoneMainTheater, unit("Armored", 2, 5, Keywords{Resistant: 2}))
	s.foes(3)
	s.wantResistant(a, 3, "Cornered (2 against 3): printed Resistant 2 plus the granted 1")
}

// Covers: SVY-R44
func TestCard_AncestralMonument_FollowsItsZone(t *testing.T) {
	s := newScene(t)
	mon := s.put(0, ZoneMainTheater, s.card("Ancestral Monument"))
	u := s.put(0, ZoneMainTheater, unit("U", 2, 5, Keywords{}))
	r := s.put(0, ZoneReserves, unit("R", 2, 5, Keywords{}))
	s.foes(3)
	s.wantResistant(u, 1, "Cornered (2 against 3), same zone as the Monument")
	s.wantResistant(r, 0, "Reserves is not the Monument's zone")
	s.moveTo(mon, ZoneReserves)
	s.wantResistant(r, 1, "still Cornered (1 against 3); the Monument is now in Reserves, so R matches")
	s.wantResistant(u, 0, "the Monument left the Main Theater, so U no longer matches")
}

// Covers: FR-KWD-002b, FR-REP-004, FR-LAY-003, FR-KWD-002e
func TestCard_AncestralMonument_CorneredUnitTakesOneLess(t *testing.T) {
	run := func(t *testing.T, extraAttackers int, want int) {
		s := newScene(t)
		s.put(1, ZoneMainTheater, s.card("Ancestral Monument"))
		d := s.put(1, ZoneMainTheater, unit("D", 0, 5, Keywords{}))
		att := s.put(0, ZoneMainTheater, unit("Attacker", 3, 5, Keywords{}))
		for i := 0; i < extraAttackers; i++ {
			s.put(0, ZoneMainTheater, unit("Ally", 0, 5, Keywords{}))
		}
		s.do(s.attack(att, d))
		s.wantDamage(d, want)
	}
	t.Run("cornered", func(t *testing.T) {
		run(t, 2, 2) // 3 against 2: player 1 Cornered, Resistant 1 reduces 3 to 2
	})
	t.Run("not cornered", func(t *testing.T) {
		run(t, 1, 3) // 2 against 2: not Cornered, full 3
	})
}
