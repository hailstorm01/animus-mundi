package engine

import (
	"reflect"
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for a card's own Cornered power (T-0260, SVY-R18): "Cornered: This unit
// has +2 power." While its controller is Cornered (FR-KWD-002e) a face-up
// battlefield asset has +Power Combat Power as a Layer 4 static (FR-LAY-001),
// read by Game.Power. The gate is live and locked for a simultaneous step
// (FR-LAY-003). Synthetic cards come from corneredCard with
// ability.WhileCornered{Power: 2}.
//
// This file does not compile until ability.WhileCornered has a Power field.

func TestFR_KWD_002e_CorneredPowerOnlyWhileCornered(t *testing.T) {
	s := newScene(t)
	u := s.put(0, ZoneMainTheater, corneredCard("Highlander", 2, 9, ability.WhileCornered{Power: 2}))
	s.put(1, ZoneMainTheater, unit("Foe", 0, 5, Keywords{}))
	s.wantPower(u, 2, "1 against 1 is not Cornered, so printed power only")
	second := s.put(1, ZoneMainTheater, unit("Second Foe", 0, 5, Keywords{}))
	s.wantPower(u, 4, "Cornered (1 against 2): printed 2 plus the Cornered +2")
	s.moveTo(second, ZoneReserves)
	s.wantPower(u, 2, "no longer Cornered once an enemy leaves the Main Theater")
	if got, want := s.g.Keywords(u), u.Printed().Keywords; !reflect.DeepEqual(got, want) {
		t.Errorf("a Cornered power grants no keywords: got %+v, want printed %+v", got, want)
	}
	s.put(1, ZoneMainTheater, unit("Third Foe", 0, 5, Keywords{}))
	if got, want := s.g.Keywords(u), u.Printed().Keywords; !reflect.DeepEqual(got, want) {
		t.Errorf("a Cornered power grants no keywords while Cornered either: got %+v, want printed %+v", got, want)
	}
}

func TestFR_KWD_002e_CorneredPowerInReservesFollowsTheController(t *testing.T) {
	s := newScene(t)
	u := s.put(0, ZoneReserves, corneredCard("Highlander", 2, 9, ability.WhileCornered{Power: 2}))
	s.put(1, ZoneMainTheater, unit("Foe", 0, 5, Keywords{}))
	s.wantPower(u, 4, "Cornered belongs to the controller (0 against 1 in the Main Theater), wherever the unit is")
	s.put(0, ZoneMainTheater, unit("Ally", 0, 5, Keywords{}))
	s.wantPower(u, 2, "1 against 1 is not Cornered; Headquarters count for neither side")
}

func TestFR_LAY_001_CorneredPowerAddsBeforeTheClamp(t *testing.T) {
	t.Run("clamp", func(t *testing.T) {
		s := newScene(t)
		u := s.put(0, ZoneMainTheater, corneredCard("Highlander", 1, 9, ability.WhileCornered{Power: 2}))
		s.put(1, ZoneMainTheater, unit("Foe", 0, 5, Keywords{}))
		s.put(1, ZoneMainTheater, unit("Second Foe", 0, 5, Keywords{}))
		u.Modifiers = []Modifier{{Layer: 5, Power: -4}}
		s.wantPower(u, 0, "1 + 2 - 4 clamps at 0 in Layer 6")
		u.Modifiers = []Modifier{{Layer: 5, Power: -2}}
		s.wantPower(u, 1, "1 + 2 - 2 = 1: the Cornered power is added before the clamp")
	})
	t.Run("with an aura", func(t *testing.T) {
		s := newScene(t)
		s.put(0, ZoneReserves, auraUnit("Aura", 0, 5, otherAlliedSame, 1))
		u := s.put(0, ZoneReserves, corneredCard("Highlander", 1, 9, ability.WhileCornered{Power: 2}))
		s.put(1, ZoneMainTheater, unit("Foe", 0, 5, Keywords{}))
		s.wantPower(u, 4, "printed 1 + aura 1 + Cornered 2: both are Layer 4 and add")
	})
}

func TestFR_LAY_003_SimultaneousDamageLocksCorneredPower(t *testing.T) {
	// Covers: FR-KWD-002e, FR-CBT-002
	s := newScene(t)
	att := s.put(0, ZoneMainTheater, corneredCard("Highlander", 1, 20, ability.WhileCornered{Power: 2}))
	def := s.put(1, ZoneMainTheater, unit("Defender", 0, 3, Keywords{}))
	s.put(1, ZoneMainTheater, unit("Wall", 0, 20, Keywords{}))
	s.do(s.attack(att, def))
	if s.alive(def) {
		t.Fatalf("precondition: the defender takes 3 and is destroyed, which evens the counts")
	}
	s.wantDamage(def, 3) // power 1 + Cornered 2, locked for the combat step (FR-LAY-003)
	s.wantPower(att, 1, "after the sweep the counts are 1 against 1: printed power only")
}

func TestFR_ZON_004_SweepRecordsCorneredPowerBeforeAnyLeaves(t *testing.T) {
	s := newScene(t)
	u := s.put(0, ZoneMainTheater, corneredCard("Highlander", 1, 2, ability.WhileCornered{Power: 2}))
	a := s.put(1, ZoneMainTheater, unit("A", 0, 2, Keywords{}))
	s.put(1, ZoneMainTheater, unit("B", 0, 20, Keywords{}))
	u.Damage, a.Damage = 2, 2
	s.g.deathSweep()
	if s.alive(u) || s.alive(a) {
		t.Fatalf("setup: both lethally damaged assets should be dead")
	}
	lk := s.g.LastKnownEffective(u.ID)
	if lk == nil || lk.Power != 3 {
		t.Errorf("last-known power of the unit = %v, want 3 (printed 1 + Cornered 2): the sweep records every dying asset before any leaves (FR-ZON-004)", lk)
	}
}

func TestFR_ZON_005_FaceDownAssetHasNoCorneredPower(t *testing.T) {
	s := newScene(t)
	c := corneredCard("Hidden", 5, 5, ability.WhileCornered{Power: 2})
	f := &Asset{ID: s.g.newID(), Inst: &Instance{ID: s.g.newID(), Card: c, Owner: 0}, Controller: 0, FaceDown: true, Exhausted: true}
	s.g.placeOnBattlefield(f, ZoneReserves)
	s.put(1, ZoneMainTheater, unit("Foe", 0, 5, Keywords{}))
	if got, want := s.g.Power(f), f.Printed().Power; got != want {
		t.Errorf("a face-down asset has no printed abilities, so no Cornered power: power %d, want the face-down printed %d (FR-ZON-005)", got, want)
	}
}
