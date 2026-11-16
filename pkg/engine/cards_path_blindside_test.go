package engine

import "testing"

// Tests for T-0250: the Path Maneuver Blindside, loaded from data/cards.
// Player 0 enacts in their Main Phase; player 1 is the opponent.
//
// Blindside: "Chosen opposing unit gets -3 power until the end of your
// opponent's turn." Cost 2 Tactics (R0 T2).

// blindside gives player 0 the real card with exactly R0 T2.
func blindside(s *scene) *Instance {
	inst := s.give(0, s.card("Blindside"))
	s.setPool(0, 0, 2)
	return inst
}

func TestCard_Blindside_LastsThroughOpponentsTurn(t *testing.T) {
	// Covers: FR-LAY-003, FR-TRN-001
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Foe", 4, 9, Keywords{}))
	s.do(enactAction(blindside(s)))
	wantPower(t, s, foe, 1, "right after Blindside resolves (printed 4 -3)")
	s.endTurn()
	if s.g.Active != 1 || s.g.Phase != PhaseMain {
		t.Fatalf("precondition: player 1's Main Phase, got active %d phase %v", s.g.Active, s.g.Phase)
	}
	wantPower(t, s, foe, 1, "through the opponent's turn (it does not expire in player 0's Sub-Phase 4c)")
	s.endTurn()
	wantPower(t, s, foe, 4, "after the opponent's turn ended (expires in its Sub-Phase 4c)")
	s.endTurn()
	wantPower(t, s, foe, 4, "after player 0's following turn ended")
}

func TestCard_Blindside_ReducedPowerInOpponentsAttack(t *testing.T) {
	// Covers: FR-LAY-003, FR-LAY-001
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Foe", 4, 9, Keywords{}))
	ally := s.put(0, ZoneMainTheater, unit("Ally", 1, 9, Keywords{}))
	s.do(enactAction(blindside(s)))
	s.endTurn()
	s.do(s.attack(foe, ally))
	s.wantDamage(ally, 1)
}

func TestCard_Blindside_ClampsAtZero(t *testing.T) {
	// Covers: FR-LAY-001, FR-LAY-003
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Foe", 2, 9, Keywords{}))
	s.do(enactAction(blindside(s)))
	wantPower(t, s, foe, 0, "printed 2 -3 is clamped at 0 (FR-LAY-001 Layer 6), not negative")
	s.endTurn()
	wantPower(t, s, foe, 0, "through the opponent's turn")
	s.endTurn()
	wantPower(t, s, foe, 2, "after the opponent's turn ended")
}

func TestCard_Blindside_AlliedUnitIsNotATarget(t *testing.T) {
	// Covers: FR-TGT-001, FR-MNV-001
	s := newScene(t)
	ally := s.put(0, ZoneMainTheater, unit("Ally", 3, 9, Keywords{}))
	foe := s.put(1, ZoneMainTheater, unit("Foe", 4, 9, Keywords{}))
	chooseAsset(s, foe)
	s.do(enactAction(blindside(s)))
	for _, d := range s.asked(0) {
		if d.Kind != DecideTarget {
			continue
		}
		for _, o := range d.Options {
			if o.Ref == ally.ID {
				t.Errorf("an allied unit must not be offered as a target (FR-TGT-001)")
			}
		}
	}
	wantPower(t, s, foe, 1, "the opposing unit gets -3")
	if got := s.g.Power(ally); got != 3 {
		t.Errorf("the allied unit is unchanged: power %d, want 3 (FR-TGT-001)", got)
	}
}

func TestCard_Blindside_NotEnactableWithoutOpposingUnit(t *testing.T) {
	// Covers: FR-MNV-001
	cases := []struct {
		name  string
		setup func(s *scene)
		want  bool
	}{
		{"no unit at all", func(s *scene) {}, false},
		{"only an opposing structure and allied units", func(s *scene) {
			s.put(1, ZoneMainTheater, structure("FoeWall", 0, 3, Keywords{}))
			s.put(0, ZoneMainTheater, unit("Ally", 1, 3, Keywords{}))
		}, false},
		{"one opposing unit in Reserves", func(s *scene) {
			s.put(1, ZoneReserves, unit("Foe", 1, 3, Keywords{}))
		}, true},
	}
	for _, c := range cases {
		s := newScene(t)
		c.setup(s)
		inst := blindside(s)
		if got := offersEnact(s.g, inst); got != c.want {
			t.Errorf("%s: Enact offered = %v, want %v (FR-MNV-001)", c.name, got, c.want)
		}
		if !c.want {
			s.reject(enactAction(inst))
			if !inHand(s.g.Players[0], inst) {
				t.Errorf("%s: the refused Maneuver stays in hand (FR-MNV-001)", c.name)
			}
		}
	}
}
