package engine

import "testing"

// Card tests for Allshan, Chainbreaker Revenant (T-0070), written from the CSV
// rules text in data/cards/units-and-structures-path.csv: Unit, 5 Requisition,
// Reserves, 5/6, "Unbowed: This unit gains 2 armor." and "Cornered: Whenever
// this unit attacks, it gets +3 power until end of turn."
// Rules: FR-KWD-002a (Unbowed), FR-KWD-002k (Armor), FR-KWD-002e (Cornered),
// FR-LAY-003 (gate checked when the event happens), SVY-R13.

// Control: a 5/6 that enters Reserves for 5 Requisition and does nothing on deploy.
func TestCard_AllshanChainbreakerRevenant_EntersReservesAsFiveSixForFiveRequisitionAndDoesNothing(t *testing.T) {
	s := newScene(t)
	s.put(1, ZoneMainTheater, unit("Enemy", 1, 5, Keywords{}))
	s.script(0)
	s.script(1)
	c := deployPath(s, "Allshan, Chainbreaker Revenant", 5)
	a := assetOf(s.g, c)
	if a == nil || a.Zone != ZoneReserves {
		t.Fatalf("Allshan is on the battlefield in Reserves after deploying")
	}
	if p := a.Printed(); p.Power != 5 || p.Health != 6 {
		t.Errorf("Allshan is a 5/6: got %d/%d", p.Power, p.Health)
	}
	if s.g.Players[0].Req != 0 {
		t.Errorf("Allshan costs 5 Requisition: %d left", s.g.Players[0].Req)
	}
	if a.Armor != 0 || a.TempPower != 0 {
		t.Errorf("Unbowed and Cornered are not deploy triggers: Armor %d, TempPower %d", a.Armor, a.TempPower)
	}
	if n := len(s.asked(0)) + len(s.asked(1)); n != 0 {
		t.Errorf("nothing is asked on deploy: asked %d times", n)
	}
	if got := totalDamage(s.g); got != 0 {
		t.Errorf("nothing happens on deploy: total damage %d", got)
	}
}

// Covers: FR-KWD-002a
func TestCard_AllshanChainbreakerRevenant_UnbowedGainsTwoArmorWhenItSurvivesDamage(t *testing.T) {
	s := newScene(t)
	al := s.put(1, ZoneMainTheater, s.card("Allshan, Chainbreaker Revenant"))
	hitter := s.put(0, ZoneMainTheater, unit("Hitter", 2, 9, Keywords{}))
	s.do(s.attack(hitter, al))
	s.wantDamage(al, 2)
	if al.Armor != 2 {
		t.Errorf("Unbowed: Allshan gains 2 Armor after taking damage and surviving: got %d", al.Armor)
	}
}

// Covers: FR-KWD-002k, FR-REP-004
func TestCard_AllshanChainbreakerRevenant_UnbowedArmorAbsorbsLaterDamage(t *testing.T) {
	s := newScene(t)
	al := s.put(1, ZoneMainTheater, s.card("Allshan, Chainbreaker Revenant"))
	h1 := s.put(0, ZoneMainTheater, unit("First Hitter", 2, 9, Keywords{}))
	h2 := s.put(0, ZoneMainTheater, unit("Second Hitter", 3, 9, Keywords{}))
	s.do(s.attack(h1, al))
	s.do(s.attack(h2, al))
	// 3 damage: 2 absorbed by the Armor from the first hit, 1 marked on Health;
	// the second hit also triggers Unbowed for another 2 Armor.
	s.wantDamage(al, 3)
	if al.Armor != 2 {
		t.Errorf("the second hit consumes 2 Armor and Unbowed grants 2 more: Armor %d, want 2", al.Armor)
	}
}

// Covers: FR-KWD-002a
func TestCard_AllshanChainbreakerRevenant_UnbowedDoesNotFireWhenNoDamageIsDealt(t *testing.T) {
	s := newScene(t)
	al := s.put(1, ZoneMainTheater, s.card("Allshan, Chainbreaker Revenant"))
	hitter := s.put(0, ZoneMainTheater, unit("Harmless", 0, 9, Keywords{}))
	s.do(s.attack(hitter, al))
	if al.Armor != 0 {
		t.Errorf("Unbowed needs at least 1 damage taken: Armor %d, want 0", al.Armor)
	}
}

// Covers: FR-KWD-002e, FR-LAY-003, FR-LAY-001
func TestCard_AllshanChainbreakerRevenant_CorneredAttackDealsEightAndPowerResetsAtEndOfTurn(t *testing.T) {
	s := newScene(t)
	al := s.put(0, ZoneMainTheater, s.card("Allshan, Chainbreaker Revenant"))
	def := s.put(1, ZoneMainTheater, unit("Wall", 0, 20, Keywords{}))
	s.put(1, ZoneMainTheater, unit("Second Wall", 0, 20, Keywords{}))
	s.do(s.attack(al, def))
	s.wantDamage(def, 8)
	if al.TempPower != 3 {
		t.Errorf("Cornered: +3 power until end of turn: TempPower %d, want 3", al.TempPower)
	}
	s.endTurn()
	if al.TempPower != 0 {
		t.Errorf("the +3 power expires at end of turn: TempPower %d, want 0", al.TempPower)
	}
	if p := al.Printed(); p.Power != 5 {
		t.Errorf("Allshan is a 5 power unit again: %d", p.Power)
	}
}

// Covers: FR-KWD-002e
func TestCard_AllshanChainbreakerRevenant_NotCorneredAttackGainsNothing(t *testing.T) {
	s := newScene(t)
	al := s.put(0, ZoneMainTheater, s.card("Allshan, Chainbreaker Revenant"))
	def := s.put(1, ZoneMainTheater, unit("Wall", 0, 20, Keywords{}))
	s.do(s.attack(al, def))
	s.wantDamage(def, 5)
	if al.TempPower != 0 {
		t.Errorf("Cornered needs the opponent to control strictly more Main Theater assets: TempPower %d, want 0", al.TempPower)
	}
}
