package engine

import "testing"

// Tests for T-0231 (directive #147, FR-CST-001, FR-TRN-003): PlayCost, the one
// reader a client plans a card play with, from its owner's hand or discard
// pile. It uses T-0214's ddUnit, whose permission costs killing an allied unit,
// and T-0140's anotherEffect for a supported Maneuver. Player 0 owns every card.

func TestPlayCost_HandCards(t *testing.T) {
	s := newScene(t)
	u := unit("Grunt", 1, 1, Keywords{})
	u.ReqCost, u.TacCost = 2, 1
	if req, tac, ok := s.g.PlayCost(s.give(0, u)); req != 2 || tac != 1 || !ok {
		t.Errorf("a supported unit in hand: got R%d T%d %v, want R2 T1 true (EffectiveCost)", req, tac, ok)
	}
	if req, tac, ok := s.g.PlayCost(s.give(0, maneuver("Ambush", 1, 2, anotherEffect()))); req != 1 || tac != 2 || !ok {
		t.Errorf("a supported Maneuver in hand: got R%d T%d %v, want R1 T2 true", req, tac, ok)
	}
	odd := unit("Odd", 1, 1, Keywords{})
	odd.Unsupported = []string{"an unexecutable line"}
	if _, _, ok := s.g.PlayCost(s.give(0, odd)); ok {
		t.Errorf("an unsupported card in hand: ok true, want false")
	}
	if _, _, ok := s.g.PlayCost(s.give(0, &Card{Name: "Dud", Type: TypeManeuver, TacCost: 2})); ok {
		t.Errorf("a Maneuver with no effect in hand: ok true, want false (not Supported)")
	}
	stuck := unit("Stuck", 1, 1, Keywords{})
	stuck.DeployZones = nil
	if _, _, ok := s.g.PlayCost(s.give(0, stuck)); ok {
		t.Errorf("a hand card with no deploy zone that is not a Maneuver: ok true, want false")
	}
}

func TestPlayCost_DiscardCards(t *testing.T) {
	s := newScene(t)
	s.put(0, ZoneMainTheater, unit("Ally", 1, 5, Keywords{}))
	if req, tac, ok := s.g.PlayCost(s.inDiscard(0, ddUnit("Phoenix", 2, 1))[0]); req != 2 || tac != 1 || !ok {
		t.Errorf("a permitted discard card with payable costs: got R%d T%d %v, want R2 T1 true (DiscardDeployCost)", req, tac, ok)
	}
	zoneless := ddUnit("Zoneless", 2, 1)
	zoneless.DeployZones = nil
	if _, _, ok := s.g.PlayCost(s.inDiscard(0, zoneless)[0]); ok {
		t.Errorf("a discard card with no deploy zone: ok true, want false")
	}
	if _, _, ok := s.g.PlayCost(s.inDiscard(0, unit("Plain", 1, 1, Keywords{}))[0]); ok {
		t.Errorf("a discard card with no permission: ok true, want false (FR-TRN-003)")
	}
}

func TestPlayCost_OtherZones(t *testing.T) {
	s := newScene(t)
	if _, _, ok := s.g.PlayCost(s.put(0, ZoneMainTheater, unit("Fielded", 1, 1, Keywords{})).Inst); ok {
		t.Errorf("a card on the battlefield: ok true, want false")
	}
	if _, _, ok := s.g.PlayCost(s.onTop(0, ddUnit("Buried", 2, 1))[0]); ok {
		t.Errorf("a card in a deck: ok true, want false")
	}
}
