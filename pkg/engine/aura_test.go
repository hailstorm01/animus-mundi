package engine

import (
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// auraStatic is a Layer 4 power aura with filter f and power n.
func auraStatic(f ability.Filter, n int) ability.Static {
	return ability.Static{Kind: ability.StaticAura, Filter: f, Power: n}
}

// auraUnit is a unit printing one power aura.
func auraUnit(name string, power, health int, f ability.Filter, n int) *Card {
	c := unit(name, power, health, ability.Keywords{})
	c.Ability = &ability.Def{Statics: []ability.Static{auraStatic(f, n)}}
	return c
}

// otherAlliedSame is "other allied assets in the same zone".
var otherAlliedSame = ability.Filter{Allegiance: ability.AllegianceAllied, Other: true, ZoneRef: ability.ZoneRefSame}

func (s *scene) wantPower(a *Asset, want int, why string) {
	s.t.Helper()
	if got := s.g.Power(a); got != want {
		s.t.Errorf("%s power = %d, want %d (%s)", a, got, want, why)
	}
}

func TestFR_LAY_001_AuraBuffsOtherAlliedInSameZone(t *testing.T) {
	s := newScene(t)
	src := s.put(0, ZoneMainTheater, auraUnit("Aura", 1, 5, otherAlliedSame, 1))
	u := s.put(0, ZoneMainTheater, unit("Ally", 2, 5, ability.Keywords{}))
	st := s.put(0, ZoneMainTheater, structure("Wall", 0, 5, ability.Keywords{}))
	res := s.put(0, ZoneReserves, unit("Reserve Ally", 2, 5, ability.Keywords{}))
	foe := s.put(1, ZoneMainTheater, unit("Foe", 2, 5, ability.Keywords{}))
	s.wantPower(u, 3, "an allied unit in the source zone gains the aura")
	s.wantPower(st, 1, "an allied structure in the source zone gains the aura")
	s.wantPower(res, 2, "an allied unit in another zone gains nothing")
	s.wantPower(foe, 2, "an opposing asset gains nothing from an allied aura")
	s.wantPower(src, 1, "an aura that says other excludes its source")
	s.wantPower(s.g.Players[0].HQ, 3, "the Headquarters Zone is not the source zone")
}

func TestFR_LAY_001_AuraIncludesSourceWithoutOther(t *testing.T) {
	s := newScene(t)
	horror := func(name string, power int) *Card {
		c := unit(name, power, 5, ability.Keywords{})
		c.Subtypes = []string{"Horror"}
		return c
	}
	f := ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit, Subtypes: []string{"Horror"}, ZoneRef: ability.ZoneRefSame}
	srcCard := horror("Horror Source", 4)
	srcCard.Ability = &ability.Def{Statics: []ability.Static{auraStatic(f, 1)}}
	src := s.put(0, ZoneReserves, srcCard)
	other := s.put(0, ZoneReserves, horror("Other Horror", 2))
	plain := s.put(0, ZoneReserves, unit("Plain", 2, 5, ability.Keywords{}))
	wall := s.put(0, ZoneReserves, structure("Wall", 0, 5, ability.Keywords{}))
	s.wantPower(src, 5, "without other the source is in its own filter")
	s.wantPower(other, 3, "another allied Horror unit in the zone gains the aura")
	s.wantPower(plain, 2, "a non-Horror unit does not match the subtype")
	s.wantPower(wall, 0, "a structure does not match the unit kind")
}

func TestFR_LAY_001_AuraFollowsSourcesCurrentZone(t *testing.T) {
	// SVY-R44, last sentence (FR-TGT-001): a continuous ability uses its source current zone.
	s := newScene(t)
	src := s.put(0, ZoneMainTheater, auraUnit("Aura", 1, 5, otherAlliedSame, 1))
	a := s.put(0, ZoneMainTheater, unit("A", 2, 5, ability.Keywords{}))
	b := s.put(0, ZoneReserves, unit("B", 2, 5, ability.Keywords{}))
	s.wantPower(a, 3, "source in Main Theater buffs A")
	s.wantPower(b, 2, "source in Main Theater does not buff Reserves")
	s.moveTo(src, ZoneReserves)
	s.wantPower(a, 2, "source in Reserves no longer buffs the Main Theater")
	s.wantPower(b, 3, "source in Reserves buffs B")
	s.moveTo(src, ZoneMainTheater)
	s.wantPower(a, 3, "source back in Main Theater buffs A again")
	s.wantPower(b, 2, "source back in Main Theater no longer buffs B")
}

func TestFR_LAY_001_AuraAppliesToLaterArrivals(t *testing.T) {
	// FR-LAY-001: auras evaluate dynamically across board state changes.
	s := newScene(t)
	s.put(0, ZoneMainTheater, auraUnit("Aura", 1, 5, otherAlliedSame, 1))
	late := s.put(0, ZoneMainTheater, unit("Late", 2, 5, ability.Keywords{}))
	s.wantPower(late, 3, "an asset entering the zone afterwards has the aura at once")
	mover := s.put(0, ZoneReserves, unit("Mover", 2, 5, ability.Keywords{}))
	s.wantPower(mover, 2, "in Reserves the mover has no aura")
	s.moveTo(mover, ZoneMainTheater)
	s.wantPower(mover, 3, "moving into the zone gains the aura on arrival")
	s.moveTo(mover, ZoneReserves)
	s.wantPower(mover, 2, "moving out of the zone loses the aura")
}

func TestFR_LAY_001_AuraEndsWhenSourceLeaves(t *testing.T) {
	s := newScene(t)
	src := s.put(0, ZoneMainTheater, auraUnit("Aura", 1, 2, otherAlliedSame, 1))
	a := s.put(0, ZoneMainTheater, unit("A", 2, 5, ability.Keywords{}))
	s.wantPower(a, 3, "the aura applies while its source is on the battlefield")
	src.Damage = 2
	s.g.deathSweep()
	if s.alive(src) {
		t.Fatalf("setup: the lethally damaged source should be dead")
	}
	s.wantPower(a, 2, "the aura stops when its source leaves the battlefield")
}

func TestFR_ZON_001_AuraSameZoneReservesAreOwn(t *testing.T) {
	s := newScene(t)
	f := ability.Filter{Other: true, ZoneRef: ability.ZoneRefSame}
	src := s.put(0, ZoneReserves, auraUnit("Aura", 1, 5, f, 1))
	mine := s.put(0, ZoneReserves, unit("Mine", 2, 5, ability.Keywords{}))
	theirs := s.put(1, ZoneReserves, unit("Theirs", 2, 5, ability.Keywords{}))
	s.wantPower(mine, 3, "the source own Reserves is its zone")
	s.wantPower(theirs, 2, "the opponent Reserves is a different zone (FR-ZON-001)")
	s.moveTo(src, ZoneMainTheater)
	shared := s.put(1, ZoneMainTheater, unit("Shared", 2, 5, ability.Keywords{}))
	s.wantPower(shared, 3, "the Main Theater is shared, so an opposing unit there is in the same zone")
}

func TestFR_LAY_001_FaceDownSourceHasNoAura(t *testing.T) {
	s := newScene(t)
	down := s.putScratch(0)
	down.Inst.Card = auraUnit("Hidden Aura", 1, 5, otherAlliedSame, 1)
	s.moveTo(down, ZoneMainTheater)
	ally := s.put(0, ZoneMainTheater, unit("Ally", 2, 5, ability.Keywords{}))
	s.wantPower(ally, 2, "a face-down asset has no abilities, so its printed aura does nothing")

	s2 := newScene(t)
	s2.put(0, ZoneMainTheater, auraUnit("Aura", 1, 5, otherAlliedSame, 1))
	sc := s2.putScratch(0)
	s2.moveTo(sc, ZoneMainTheater)
	s2.wantPower(sc, 2, "a face-down Scratch Company still receives a face-up source aura")
}

func TestFR_LAY_001_AurasStackAndClampAfterLayer5(t *testing.T) {
	s := newScene(t)
	s1 := s.put(0, ZoneMainTheater, auraUnit("Aura 1", 1, 5, otherAlliedSame, 1))
	s2 := s.put(0, ZoneMainTheater, auraUnit("Aura 2", 1, 5, otherAlliedSame, 1))
	a := s.put(0, ZoneMainTheater, unit("A", 1, 5, ability.Keywords{}))
	s.wantPower(a, 3, "two auras add")
	s.wantPower(s1, 2, "each source gains the other aura")
	s.wantPower(s2, 2, "each source gains the other aura")
	a.Modifiers = []Modifier{{Layer: 5, Power: -3}}
	s.wantPower(a, 0, "Layer 1 + Layer 4 + Layer 5 = 0")
	a.Modifiers = []Modifier{{Layer: 5, Power: -5}}
	s.wantPower(a, 0, "Layer 6 clamps the sum of Layers 1, 4 and 5 at 0, not each layer")
}

func TestFR_LAY_001_AuraFollowsSourcesController(t *testing.T) {
	s := newScene(t)
	src := s.put(0, ZoneMainTheater, auraUnit("Aura", 1, 5, otherAlliedSame, 1))
	mine := s.put(0, ZoneMainTheater, unit("Mine", 2, 5, ability.Keywords{}))
	u := s.put(1, ZoneMainTheater, unit("U", 2, 5, ability.Keywords{}))
	s.wantPower(u, 2, "an opposing unit gets no allied aura")
	s.wantPower(mine, 3, "setup: the controller unit has the aura")
	src.Controller = 1
	s.wantPower(u, 3, "allied is relative to the source current controller")
	s.wantPower(mine, 2, "the former controller unit loses the aura")
}

func TestFR_CBT_002_AuraPowerDealsCombatDamage(t *testing.T) {
	s := newScene(t)
	s.put(0, ZoneMainTheater, auraUnit("Aura", 1, 5, otherAlliedSame, 1))
	att := s.put(0, ZoneMainTheater, unit("Attacker", 2, 5, ability.Keywords{}))
	def := s.put(1, ZoneMainTheater, unit("Defender", 0, 10, ability.Keywords{}))
	s.do(s.attack(att, def))
	s.wantDamage(def, 3)
}

func TestFR_ZON_004_SweepRecordsAuraPowerBeforeAnyLeaves(t *testing.T) {
	s := newScene(t)
	src := s.put(0, ZoneMainTheater, auraUnit("Aura", 1, 2, otherAlliedSame, 1))
	a := s.put(0, ZoneMainTheater, unit("A", 2, 2, ability.Keywords{}))
	src.Damage, a.Damage = 2, 2
	s.g.deathSweep()
	if s.alive(src) || s.alive(a) {
		t.Fatalf("setup: both lethally damaged assets should be dead")
	}
	lk := s.g.LastKnownEffective(a.ID)
	if lk == nil || lk.Power != 3 {
		t.Errorf("last-known power of A = %v, want 3: the sweep records every dying asset before any leaves (FR-ZON-004)", lk)
	}
}
