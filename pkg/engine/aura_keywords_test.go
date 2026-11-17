package engine

import (
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// alliedHorrorSame is "allied Horror units in the same zone".
var alliedHorrorSame = ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit, Subtypes: []string{"Horror"}, ZoneRef: ability.ZoneRefSame}

// horrorUnit is a Horror unit printing Resistant r.
func horrorUnit(name string, power, health, r int) *Card {
	c := unit(name, power, health, Keywords{Resistant: r})
	c.Subtypes = []string{"Horror"}
	return c
}

// resistAuraHorror is a Horror unit printing "allied Horror units in the same
// zone have Resistant n" (the filter includes the source).
func resistAuraHorror(name string, power, health, n int) *Card {
	c := horrorUnit(name, power, health, 0)
	c.Ability = &ability.Def{Statics: []ability.Static{{Kind: ability.StaticAura, Filter: alliedHorrorSame, Resistant: n}}}
	return c
}

func (s *scene) wantResistant(a *Asset, want int, why string) {
	s.t.Helper()
	if got := s.g.Keywords(a).Resistant; got != want {
		s.t.Errorf("%s Resistant = %d, want %d (%s)", a, got, want, why)
	}
}

// Covers: SVY-R44
func TestFR_LAY_001_AuraGrantsResistantToMatchesInSameZone(t *testing.T) {
	s := newScene(t)
	src := s.put(0, ZoneReserves, resistAuraHorror("Source", 0, 5, 1))
	other := s.put(0, ZoneReserves, horrorUnit("Other Horror", 2, 5, 0))
	plain := s.put(0, ZoneReserves, unit("Plain", 2, 5, Keywords{}))
	main := s.put(0, ZoneMainTheater, horrorUnit("Main Horror", 2, 5, 0))
	foe := s.put(1, ZoneReserves, horrorUnit("Foe Horror", 2, 5, 0))
	s.wantResistant(src, 1, "without other the source is in its own filter")
	s.wantResistant(other, 1, "another allied Horror unit in the zone has the grant")
	s.wantResistant(plain, 0, "a non-Horror unit does not match")
	s.wantResistant(main, 0, "a Horror unit in another zone does not match")
	s.wantResistant(foe, 0, "an opposing Horror unit does not match an allied aura")
	s.wantPower(src, 0, "a keyword grant adds no power")
	s.wantPower(other, 2, "a keyword grant adds no power")
	s.wantPower(plain, 2, "a keyword grant adds no power")
	s.wantPower(main, 2, "a keyword grant adds no power")
	s.wantPower(foe, 2, "a keyword grant adds no power")
}

func TestFR_LAY_001_AuraResistantAddsToPrintedAndStacks(t *testing.T) {
	s := newScene(t)
	src := s.put(0, ZoneReserves, resistAuraHorror("Source", 0, 5, 1))
	tough := s.put(0, ZoneReserves, horrorUnit("Tough Horror", 2, 5, 2))
	s.wantResistant(tough, 3, "the grant adds to printed Resistant 2")
	src2 := s.put(0, ZoneReserves, resistAuraHorror("Source 2", 0, 5, 1))
	s.wantResistant(src, 2, "each source includes itself and the other")
	s.wantResistant(src2, 2, "each source includes itself and the other")
	s.wantResistant(tough, 4, "two grants and printed Resistant 2 sum")
}

func TestFR_LAY_001_AuraResistantFollowsSourcesZone(t *testing.T) {
	// SVY-R44, last sentence: a continuous ability uses its source's current zone.
	s := newScene(t)
	src := s.put(0, ZoneReserves, resistAuraHorror("Source", 0, 5, 1))
	a := s.put(0, ZoneReserves, horrorUnit("A", 2, 5, 0))
	b := s.put(0, ZoneMainTheater, horrorUnit("B", 2, 5, 0))
	s.wantResistant(a, 1, "source in Reserves grants A")
	s.wantResistant(b, 0, "source in Reserves does not grant the Main Theater")
	s.moveTo(src, ZoneMainTheater)
	s.wantResistant(a, 0, "source in the Main Theater no longer grants Reserves")
	s.wantResistant(b, 1, "source in the Main Theater grants B")
	src.Damage = 5
	s.g.deathSweep()
	if s.alive(src) {
		t.Fatalf("setup: the lethally damaged source should be dead")
	}
	s.wantResistant(b, 0, "the grant ends when its source leaves the battlefield")
}

func TestFR_LAY_001_FaceDownSourceGrantsNoKeyword(t *testing.T) {
	s := newScene(t)
	down := s.putScratch(0)
	down.Inst.Card = resistAuraHorror("Hidden Aura", 0, 5, 1)
	s.moveTo(down, ZoneReserves)
	ally := s.put(0, ZoneReserves, horrorUnit("Ally", 2, 5, 0))
	s.wantResistant(ally, 0, "a face-down asset has no abilities, so its printed aura grants nothing")
}

// Covers: FR-KWD-002b
func TestFR_REP_004_AuraResistantReducesDamage(t *testing.T) {
	s := newScene(t)
	s.put(1, ZoneMainTheater, resistAuraHorror("Source", 0, 5, 1))
	d := s.put(1, ZoneMainTheater, horrorUnit("D", 0, 5, 0))
	att := s.put(0, ZoneMainTheater, unit("Attacker", 3, 5, Keywords{}))
	s.do(s.attack(att, d))
	s.wantDamage(d, 2)
	if e := damageEvent(t, s, d); e.Resistant != 1 || e.Amount != 2 {
		t.Errorf("damage event = %+v, want Resistant 1 and amount 2 (Stage 2 subtracts total Resistant)", e)
	}
}

func TestFR_ZON_004_SweepRecordsAuraResistantBeforeAnyLeaves(t *testing.T) {
	s := newScene(t)
	src := s.put(0, ZoneReserves, resistAuraHorror("Source", 0, 2, 1))
	a := s.put(0, ZoneReserves, horrorUnit("A", 2, 2, 0))
	src.Damage, a.Damage = 2, 2
	s.g.deathSweep()
	if s.alive(src) || s.alive(a) {
		t.Fatalf("setup: both lethally damaged assets should be dead")
	}
	lk := s.g.LastKnownEffective(a.ID)
	if lk == nil || lk.Keywords.Resistant != 1 {
		t.Errorf("last-known Resistant of A = %v, want 1: the sweep records every dying asset before any leaves (FR-ZON-004)", lk)
	}
}
