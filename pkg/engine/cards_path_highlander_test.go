package engine

import (
	"reflect"
	"testing"
)

// Card tests for Brigand Highlander (T-0261), written from the CSV rules text in
// data/cards/units-and-structures-path.csv: Path Unit (Officer, Hammerhand),
// Requisition 1, Tactics 1, Main Theater, 2/3, "Cornered: This unit has +2 power."
// Rules: FR-KWD-002e (Cornered, locked for a simultaneous step), FR-LAY-001
// (a static power bonus is Layer 4), FR-LAY-003 (the lock), FR-CBT-002.

// highlanderCombat puts player 0's Highlander in the Main Theater facing a
// player 1 defender (power 0, Health 20). If cornered, player 1 controls a
// second Main Theater unit, so player 0 has 1 against 2 (FR-KWD-002e);
// otherwise 1 against 1.
func highlanderCombat(t *testing.T, cornered bool) (s *scene, hl, def *Asset) {
	t.Helper()
	s = newScene(t)
	hl = s.put(0, ZoneMainTheater, s.card("Brigand Highlander"))
	def = s.put(1, ZoneMainTheater, unit("Defender", 0, 20, Keywords{}))
	if cornered {
		s.put(1, ZoneMainTheater, unit("Second Foe", 0, 5, Keywords{}))
	}
	return s, hl, def
}

// Covers: FR-KWD-002e, FR-LAY-001
func TestCard_BrigandHighlander_PlusTwoPowerOnlyWhileCornered(t *testing.T) {
	s := newScene(t)
	hl := s.put(0, ZoneMainTheater, s.card("Brigand Highlander"))
	s.put(1, ZoneMainTheater, unit("Foe", 0, 5, Keywords{}))
	s.wantPower(hl, 2, "1 against 1 is not Cornered, so printed power only")
	second := s.put(1, ZoneMainTheater, unit("Second Foe", 0, 5, Keywords{}))
	s.wantPower(hl, 4, "Cornered (1 against 2): printed 2 plus the Cornered +2")
	if got, want := s.g.Keywords(hl), hl.Printed().Keywords; !reflect.DeepEqual(got, want) {
		t.Errorf("the Cornered line grants power only, no keywords, while Cornered: got %+v, want printed %+v", got, want)
	}
	s.moveTo(second, ZoneReserves)
	s.wantPower(hl, 2, "no longer Cornered once an enemy leaves the Main Theater")
	if got, want := s.g.Keywords(hl), hl.Printed().Keywords; !reflect.DeepEqual(got, want) {
		t.Errorf("the Cornered line grants power only, no keywords, while not Cornered: got %+v, want printed %+v", got, want)
	}
}

// Covers: FR-KWD-002e, FR-LAY-003, FR-CBT-002
func TestCard_BrigandHighlander_CorneredAttackDealsFour(t *testing.T) {
	s, hl, def := highlanderCombat(t, true)
	s.do(s.attack(hl, def))
	s.wantDamage(def, 4) // printed 2 plus Cornered +2
}

// Covers: FR-KWD-002e
func TestCard_BrigandHighlander_NotCorneredAttackDealsTwo(t *testing.T) {
	s, hl, def := highlanderCombat(t, false)
	s.do(s.attack(hl, def))
	s.wantDamage(def, 2) // printed power only at 1 against 1
}
