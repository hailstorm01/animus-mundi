package engine

import (
	"animus-mundi/pkg/engine/ability"
	"testing"
)

// Card tests for Colossus Platform (T-0312), from the card text in
// data/cards/units-and-structures-HQ.csv.
//
// Colossus Platform: "Whenever this Headquarters attacks, chosen allied unit
// gains 1 armor." Headquarters, 3/20.
//
// The target is chosen on placement (SVY-R3). A Headquarters is not a unit, so
// the Platform cannot choose itself. Armor never expires and absorbs damage
// 1-for-1 (FR-KWD-002k).

const colossus = "Colossus Platform"

// cpScene starts a match with the Platform in seat 0 and the Watchman in seat 1.
func cpScene(t *testing.T) *scene { return hqScene(t, colossus, watchman) }

// cpAttack has seat 0's Platform attack seat 1's Headquarters.
func (s *scene) cpAttack() {
	s.do(s.attack(s.g.Players[0].HQ, s.g.Players[1].HQ))
}

// cpWantResolved checks that the attack went through: 3 damage each way.
func (s *scene) cpWantResolved() {
	s.t.Helper()
	s.wantDamage(s.g.Players[1].HQ, 3)
	s.wantDamage(s.g.Players[0].HQ, 3)
}

// cpWantNoArmor fails if any listed asset or either Headquarters has armor.
func (s *scene) cpWantNoArmor(as ...*Asset) {
	s.t.Helper()
	as = append(as, s.g.Players[0].HQ, s.g.Players[1].HQ)
	for _, a := range as {
		if a.Armor != 0 {
			s.t.Errorf("%s has Armor %d, want 0", a, a.Armor)
		}
	}
}

// Covers: FR-CAP-001
func TestCard_ColossusPlatform_IsAThreeTwentyHeadquarters(t *testing.T) {
	s := cpScene(t)
	a := s.g.Players[0].HQ
	if a == nil || a.Inst.Card.Name != colossus || a.Controller != 0 || a.Zone != ZoneHeadquarters {
		t.Fatalf("Colossus Platform is player 0's Headquarters in the Headquarters Zone, got %v", a)
	}
	if p, h := s.g.Power(a), s.g.MaxHealth(a); p != 3 || h != 20 {
		t.Errorf("Colossus Platform has power 3 and Maximum Health 20: got %d/%d", p, h)
	}
}

// Covers: FR-CBT-002, FR-TGT-001, FR-KWD-002k
func TestCard_ColossusPlatform_AttackGivesTheChosenUnitOneArmor(t *testing.T) {
	s := cpScene(t)
	res := s.put(0, ZoneReserves, unit("InReserves", 1, 5, ability.Keywords{}))
	main := s.put(0, ZoneMainTheater, unit("InMain", 1, 5, ability.Keywords{}))
	s.scriptRefs(0, nil, res)
	s.cpAttack()
	asks := s.asked(0)
	if len(asks) != 1 || asks[0].Kind != DecideTarget {
		t.Fatalf("the controller is asked one Target decision, got %+v", asks)
	}
	wantOptions(t, asks[0], "both allied units are offered", res, main)
	if res.Armor != 1 {
		t.Errorf("the chosen Reserves unit gains 1 armor: Armor = %d", res.Armor)
	}
	if main.Armor != 0 {
		t.Errorf("the unchosen unit gains no armor: Armor = %d", main.Armor)
	}
	s.cpWantResolved()
	s.wantSettled("the Platform's trigger resolved")
}

// Covers: FR-TGT-001, FR-ZON-005, FR-KWD-002h
func TestCard_ColossusPlatform_OffersOnlyAlliedUnits(t *testing.T) {
	s := cpScene(t)
	res := s.put(0, ZoneReserves, unit("AllyReserves", 1, 5, ability.Keywords{}))
	main := s.put(0, ZoneMainTheater, unit("AllyMain", 1, 5, ability.Keywords{}))
	ward := s.put(0, ZoneMainTheater, unit("AllyWard", 1, 5, ability.Keywords{Ward: true}))
	s.put(0, ZoneMainTheater, structure("AllyWall", 0, 5, ability.Keywords{}))
	scr := s.putScratch(0)
	s.put(1, ZoneReserves, unit("FoeReserves", 1, 5, ability.Keywords{}))
	s.put(1, ZoneMainTheater, unit("FoeMain", 1, 5, ability.Keywords{}))
	s.put(1, ZoneMainTheater, structure("FoeWall", 0, 5, ability.Keywords{}))
	s.putScratch(1)
	s.scriptRefs(0, nil, main)
	s.cpAttack()
	asks := s.asked(0)
	if len(asks) != 1 || asks[0].Kind != DecideTarget {
		t.Fatalf("the controller is asked one Target decision, got %+v", asks)
	}
	wantOptions(t, asks[0], "exactly the allied units, the allied Scratch Company included; no structure, Headquarters or opposing asset", res, main, ward, scr)
}

// Covers: FR-TGT-001
func TestCard_ColossusPlatform_OneAlliedUnitTakesItWithoutAsking(t *testing.T) {
	s := cpScene(t)
	ally := s.put(0, ZoneMainTheater, unit("Ally", 1, 5, ability.Keywords{}))
	foe := s.put(1, ZoneMainTheater, unit("Foe", 1, 5, ability.Keywords{}))
	s.script(0)
	s.cpAttack()
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("a single legal target is taken without asking (Single-Option Fast-Path): %d asked", n)
	}
	if ally.Armor != 1 {
		t.Errorf("the only allied unit gains 1 armor: Armor = %d", ally.Armor)
	}
	if foe.Armor != 0 {
		t.Errorf("the opposing unit gains no armor: Armor = %d", foe.Armor)
	}
}

// Covers: FR-TGT-003
func TestCard_ColossusPlatform_NoAlliedUnitStillAttacks(t *testing.T) {
	s := cpScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Foe", 1, 5, ability.Keywords{}))
	s.script(0)
	s.cpAttack()
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("with no legal target nothing is asked: %d asked", n)
	}
	s.cpWantNoArmor(foe)
	s.wantDamage(s.g.Players[1].HQ, 3)
	s.wantSettled("the clause with no legal target fizzled and the rest resolved")
}

// Covers: FR-KWD-002k, FR-LAY-002
func TestCard_ColossusPlatform_ArmorStacksAndLasts(t *testing.T) {
	s := cpScene(t)
	ally := s.put(0, ZoneMainTheater, unit("Armored", 1, 5, ability.Keywords{Armor: 1}))
	s.script(0)
	s.cpAttack()
	if ally.Armor != 2 {
		t.Fatalf("1 printed armor plus 1 gained is Armor 2, got %d", ally.Armor)
	}
	s.endTurn()
	s.endTurn()
	if ally.Armor != 2 {
		t.Errorf("armor never expires (FR-KWD-002k): Armor = %d after both turns end", ally.Armor)
	}
}

// Covers: FR-REP-004, FR-KWD-002k
func TestCard_ColossusPlatform_ArmorAbsorbsLaterDamage(t *testing.T) {
	s := cpScene(t)
	ally := s.put(0, ZoneMainTheater, unit("Ally", 0, 3, ability.Keywords{}))
	foe := s.put(1, ZoneMainTheater, unit("Hitter", 1, 5, ability.Keywords{}))
	s.script(0)
	s.cpAttack()
	if ally.Armor != 1 {
		t.Fatalf("setup: the ally should have Armor 1, got %d", ally.Armor)
	}
	s.endTurn()
	s.do(s.attack(foe, ally))
	if ally.Armor != 0 {
		t.Errorf("1 damage uses up the 1 armor (Stage 4): Armor = %d", ally.Armor)
	}
	s.wantDamage(ally, 0)
}

// Covers: FR-TGT-003, SVY-R25
func TestCard_ColossusPlatform_TargetGoneGivesNothing(t *testing.T) {
	s := cpScene(t)
	chosen := s.put(0, ZoneMainTheater, unit("Chosen", 1, 5, ability.Keywords{}))
	other := s.put(0, ZoneMainTheater, unit("Other", 1, 5, ability.Keywords{}))
	s.scriptRefs(0, nil, chosen)
	s.ptPlaceHook(func(int, StackItem) {
		if s.alive(chosen) {
			stormKill(s, chosen)
		}
	})
	s.cpAttack()
	if s.alive(chosen) {
		t.Fatalf("setup: the chosen unit should have been removed before resolution")
	}
	s.cpWantNoArmor(other)
	s.wantDamage(s.g.Players[1].HQ, 3)
	s.wantSettled("the trigger resolved without its removed target")
}

// Covers: FR-CBT-002
func TestCard_ColossusPlatform_OnlyItsOwnAttackTriggers(t *testing.T) {
	s := cpScene(t)
	ally := s.put(0, ZoneMainTheater, unit("Ally", 1, 9, ability.Keywords{}))
	foe := s.put(1, ZoneMainTheater, unit("Foe", 1, 9, ability.Keywords{}))
	s.script(0)
	s.do(s.attack(ally, s.g.Players[1].HQ))
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("an allied unit's attack does not fire the Platform: %d asked", n)
	}
	s.cpWantNoArmor(ally, foe)
	s.endTurn()
	s.do(s.attack(foe, s.g.Players[0].HQ))
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("an attack on the Platform does not fire it: %d asked", n)
	}
	s.cpWantNoArmor(ally, foe)
}
