package engine

import (
	"animus-mundi/pkg/engine/ability"
	"slices"
	"testing"
)

// Tests for T-0253: the Path Maneuver Poisoned Rations, loaded from
// data/cards: "Chosen unit gets -3 max health until end of turn." Player 0
// enacts in their Main Phase with R0 T3; player 1 is the opponent.

// poisonedRations gives player 0 the real card and exactly R0 T3.
func poisonedRations(s *scene) *Instance {
	inst := s.give(0, s.card("Poisoned Rations"))
	s.setPool(0, 0, 3)
	return inst
}

func TestCard_PoisonedRations_KillsHealthThreeWithoutDamage(t *testing.T) {
	// Covers: FR-LAY-002, FR-MNV-001
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Foe", 2, 3, ability.Keywords{}))
	inst := poisonedRations(s)
	from := len(s.g.Log)
	chooseAsset(s, foe)
	s.do(enactAction(inst))
	if s.alive(foe) {
		t.Errorf("a Health 3 unit at -3 Maximum Health (0) is destroyed at the sweep (FR-LAY-002)")
	}
	if !slices.Contains(s.g.Players[1].Discard, foe.Inst) {
		t.Errorf("the destroyed unit's card is in its owner's discard pile (FR-LAY-002)")
	}
	if n := mhEventsNaming(s, from, EventDamage, foe); n != 0 {
		t.Errorf("%d damage events name the unit, want none: Maximum Health 0 is lethal with no damage (FR-LAY-002)", n)
	}
}

func TestCard_PoisonedRations_KillsDamagedUnit(t *testing.T) {
	// Covers: FR-LAY-002
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Foe", 2, 5, ability.Keywords{}))
	foe.Damage = 2
	inst := poisonedRations(s)
	chooseAsset(s, foe)
	s.do(enactAction(inst))
	if s.alive(foe) {
		t.Errorf("a Health 5 unit with 2 damage at Maximum Health 2 has Current Health 0 and is destroyed (FR-LAY-002)")
	}
}

func TestCard_PoisonedRations_SurvivorRestoredAtEndOfTurn(t *testing.T) {
	// Covers: FR-LAY-002, FR-LAY-003, FR-TRN-001
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Foe", 2, 5, ability.Keywords{}))
	inst := poisonedRations(s)
	chooseAsset(s, foe)
	s.do(enactAction(inst))
	if !s.alive(foe) {
		t.Fatalf("a Health 5 unit survives -3 Maximum Health (FR-LAY-002)")
	}
	mhWant(t, s, foe, 2, 2, 0, "after Poisoned Rations")
	s.endTurn()
	mhWant(t, s, foe, 5, 5, 0, "after the end of the turn (Sub-Phase 4c)")

	s2 := newScene(t)
	hurt := s2.put(1, ZoneMainTheater, unit("Hurt", 2, 6, ability.Keywords{}))
	hurt.Damage = 1
	inst2 := poisonedRations(s2)
	chooseAsset(s2, hurt)
	s2.do(enactAction(inst2))
	if !s2.alive(hurt) {
		t.Fatalf("a Health 6 unit with 1 damage survives -3 Maximum Health (FR-LAY-002)")
	}
	mhWant(t, s2, hurt, 3, 2, 1, "damaged unit after Poisoned Rations")
	s2.endTurn()
	mhWant(t, s2, hurt, 6, 5, 1, "damaged unit after the end of the turn (damage unchanged by expiry)")
}

func TestCard_PoisonedRations_AlliedUnitIsATarget(t *testing.T) {
	// Covers: FR-TGT-001
	s := newScene(t)
	ally := s.put(0, ZoneMainTheater, unit("Ally", 2, 5, ability.Keywords{}))
	foe := s.put(1, ZoneMainTheater, unit("Foe", 2, 5, ability.Keywords{}))
	inst := poisonedRations(s)
	chooseAsset(s, ally)
	s.do(enactAction(inst))
	got := targetOffers(s)
	if !got[ally.ID] || !got[foe.ID] {
		t.Errorf("\"chosen unit\" offers both the allied and the enemy unit, got refs %v (FR-TGT-001)", got)
	}
	mhWant(t, s, ally, 2, 2, 0, "the chosen allied unit")
	mhWant(t, s, foe, 5, 5, 0, "the unchosen enemy unit")
}

func TestCard_PoisonedRations_NotEnactableWithoutUnit(t *testing.T) {
	// Covers: FR-MNV-001
	cases := []struct {
		name  string
		setup func(s *scene)
		want  bool
	}{
		{"no unit at all", func(s *scene) {}, false},
		{"only structures", func(s *scene) {
			s.put(0, ZoneMainTheater, structure("AWall", 0, 3, ability.Keywords{}))
			s.put(1, ZoneMainTheater, structure("FWall", 0, 3, ability.Keywords{}))
		}, false},
		{"one opposing unit in Reserves", func(s *scene) {
			s.put(1, ZoneReserves, unit("Foe", 1, 3, ability.Keywords{}))
		}, true},
		{"one allied unit in Reserves", func(s *scene) {
			s.put(0, ZoneReserves, unit("Ally", 1, 3, ability.Keywords{}))
		}, true},
	}
	for _, c := range cases {
		s := newScene(t)
		c.setup(s)
		inst := poisonedRations(s)
		if got := offersEnact(s.g, inst); got != c.want {
			t.Errorf("%s: Enact offered = %v, want %v (FR-MNV-001)", c.name, got, c.want)
		}
		if !c.want {
			s.reject(enactAction(inst))
			if !inHand(s.g.Players[0], inst) {
				t.Errorf("%s: the card stays in hand (FR-MNV-001)", c.name)
			}
		}
	}
}
