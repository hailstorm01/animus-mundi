package engine

import (
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for an aura's Cornered grant (T-0262, FR-KWD-002e, SVY-R18): "Other
// allied assets in the same zone have Cornered: Resistant 1." Every asset the
// aura's Filter matches has the aura's WhileCornered only while that asset's
// controller is Cornered; its keywords are a Layer 3 grant and its power Layer 4
// (FR-LAY-001). Cards are synthetic.
//
// This file does not compile until ability.Static has a Cornered field.

// corneredAura is a unit printing the aura "assets matching f have Cornered: c".
func corneredAura(name string, power, health int, f ability.Filter, c ability.WhileCornered) *Card {
	card := unit(name, power, health, Keywords{})
	card.Ability = &ability.Def{Statics: []ability.Static{{Kind: ability.StaticAura, Filter: f, Cornered: c}}}
	return card
}

// monument is "other allied assets in the same zone have Cornered: Resistant 1".
func monument(name string) *Card {
	return corneredAura(name, 0, 9, otherAlliedSame, ability.WhileCornered{Resistant: 1})
}

// foes puts n player 1 units in the Main Theater and returns them.
func (s *scene) foes(n int) []*Asset {
	var out []*Asset
	for i := 0; i < n; i++ {
		out = append(out, s.put(1, ZoneMainTheater, unit("Foe", 0, 5, Keywords{})))
	}
	return out
}

func TestFR_KWD_002e_AuraCorneredGrantOnlyWhileCornered(t *testing.T) {
	s := newScene(t)
	src := s.put(0, ZoneMainTheater, monument("Source"))
	a := s.put(0, ZoneMainTheater, unit("A", 2, 5, Keywords{}))
	s.foes(1)
	s.wantResistant(a, 0, "2 against 1 in the Main Theater: player 0 is not Cornered, so the grant is not active")
	more := s.foes(2)
	s.wantResistant(a, 1, "2 against 3: player 0 is Cornered, so the matched asset has the aura's Cornered Resistant 1")
	s.moveTo(more[0], ZoneReserves)
	s.wantResistant(a, 0, "2 against 2: no longer Cornered, so the grant ends")
	s.wantResistant(src, 0, "Other: the source is not matched by its own aura")
	s.wantPower(a, 2, "a Cornered keyword grant adds no power")
	s.wantPower(src, 0, "a Cornered keyword grant adds no power")
}

func TestFR_LAY_001_AuraCorneredGrantMatchesFilter(t *testing.T) {
	s := newScene(t)
	s.put(0, ZoneMainTheater, monument("Source"))
	unitMT := s.put(0, ZoneMainTheater, unit("Unit", 2, 5, Keywords{}))
	structMT := s.put(0, ZoneMainTheater, structure("Struct", 0, 5, Keywords{}))
	unitRes := s.put(0, ZoneReserves, unit("Reserve Unit", 2, 5, Keywords{}))
	foes := s.foes(4)
	s.wantResistant(unitMT, 1, "3 against 4: Cornered, allied unit in the source's zone matches")
	s.wantResistant(structMT, 1, "zero Kinds matches an allied structure too")
	s.wantResistant(unitRes, 0, "an allied unit in another zone does not match (SVY-R44)")
	s.wantResistant(foes[0], 0, "an opposing unit is not allied")

	// Player 1's own source: player 1 is not Cornered (4 against 3, then more).
	s.put(1, ZoneMainTheater, monument("Foe Source"))
	s.wantResistant(foes[0], 0, "player 1 has 5 against 3 and is not Cornered, even though player 0 is")
	s.wantResistant(unitMT, 1, "player 0's own grant is unaffected by player 1's source")
}

func TestFR_LAY_001_AuraCorneredGrantAddsToOwnCornered(t *testing.T) {
	s := newScene(t)
	s.put(0, ZoneMainTheater, monument("Source"))
	own := corneredCard("Own", 2, 9, ability.WhileCornered{Resistant: 1})
	own.Keywords.Resistant = 1
	u := s.put(0, ZoneMainTheater, own)
	s.foes(3)
	s.wantResistant(u, 3, "printed 1, its own Cornered 1 and the aura's Cornered 1 sum")
	s.put(0, ZoneMainTheater, monument("Second Source"))
	s.foes(1) // keep player 0 Cornered: 3 against 4
	s.wantResistant(u, 4, "two Cornered grants add to printed and own")

	t.Run("power", func(t *testing.T) {
		s := newScene(t)
		s.put(0, ZoneMainTheater, corneredAura("Source", 0, 9, otherAlliedSame, ability.WhileCornered{Power: 2}))
		u := s.put(0, ZoneMainTheater, unit("U", 2, 9, Keywords{}))
		foes := s.foes(3)
		s.wantPower(u, 4, "Cornered: printed 2 plus the aura's Cornered power 2 (Layer 4)")
		s.moveTo(foes[0], ZoneReserves)
		s.wantPower(u, 2, "2 against 2: not Cornered, so printed power only")
	})
}

func TestFR_ZON_005_AuraCorneredGrantFaceDown(t *testing.T) {
	s := newScene(t)
	s.put(1, ZoneMainTheater, unit("Foe", 0, 5, Keywords{}))
	s.put(0, ZoneReserves, monument("Source"))
	down := s.putScratch(0)
	s.wantResistant(down, 1, "player 0 is Cornered (0 against 1); a face-down asset receives the grant though its face has no Ability")

	t.Run("face-down source", func(t *testing.T) {
		s := newScene(t)
		s.put(1, ZoneMainTheater, unit("Foe", 0, 5, Keywords{}))
		hidden := s.putScratch(0)
		hidden.Inst.Card = monument("Hidden Aura")
		ally := s.put(0, ZoneReserves, unit("Ally", 2, 5, Keywords{}))
		s.wantResistant(ally, 0, "a face-down source has no abilities, so its printed aura grants nothing")
	})
}

func TestFR_LAY_003_AuraCorneredResistantLockedForCombat(t *testing.T) {
	s := newScene(t)
	s.put(1, ZoneMainTheater, monument("Source"))
	d := s.put(1, ZoneMainTheater, unit("D", 5, 5, Keywords{}))
	att := s.put(0, ZoneMainTheater, unit("Attacker", 3, 1, Keywords{}))
	s.put(0, ZoneMainTheater, unit("P1", 0, 5, Keywords{}))
	s.put(0, ZoneMainTheater, unit("P2", 0, 5, Keywords{}))
	s.wantResistant(d, 1, "3 against 2: player 1 is Cornered, so D has the aura's Resistant 1")
	s.do(s.attack(att, d))
	if s.alive(att) {
		t.Fatalf("precondition: the attacker takes 5 and dies")
	}
	s.wantDamage(d, 2)
	if e := damageEvent(t, s, d); e.Resistant != 1 || e.Amount != 2 {
		t.Errorf("damage event = %+v, want Resistant 1 and amount 2: Cornered is locked for the simultaneous damage step (FR-LAY-003)", e)
	}
	s.wantResistant(d, 0, "after the sweep it is 2 against 2, so the grant has ended")
}
