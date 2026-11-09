package engine

import "testing"

// Tests for T-0200: Boneyard Amalgam, "This unit costs R1 less to deploy for each
// unit card in your discard pile." (FR-CST-001). The real cards are loaded from
// data/cards. Printed cost is R10 T0. Player 0 is the active player.

// amalgamScene is a scene with a Boneyard Amalgam in player 0's hand.
func amalgamScene(t *testing.T) (*scene, *Instance) {
	t.Helper()
	s := newScene(t)
	return s, s.give(0, s.card("Boneyard Amalgam"))
}

// amalgamUnits returns n unit cards (Saboteur) for a discard pile.
func amalgamUnits(s *scene, n int) []*Card {
	cs := make([]*Card, n)
	for i := range cs {
		cs[i] = s.card("Saboteur")
	}
	return cs
}

func amalgamCost(t *testing.T, s *scene, inst *Instance, wantReq int, why string) {
	t.Helper()
	req, tac := s.g.EffectiveCost(inst)
	if req != wantReq || tac != 0 {
		t.Errorf("%s: effective cost R%d T%d, want R%d T0 (FR-CST-001)", why, req, tac, wantReq)
	}
}

func amalgamDeployLegal(s *scene, inst *Instance) bool {
	for _, a := range s.g.LegalActions() {
		if a.Kind == ActDeploy && a.Card == inst.ID && a.Zone == ZoneReserves {
			return true
		}
	}
	return false
}

func amalgamDeploy(inst *Instance) Action {
	return Action{Kind: ActDeploy, Player: 0, Card: inst.ID, Zone: ZoneReserves}
}

func TestCard_BoneyardAmalgam_PrintedCostWithNoUnitCards(t *testing.T) {
	s, inst := amalgamScene(t)
	s.inDiscard(0, s.card("Steel Rain"), s.card("Hammerhand Bivouac"))
	amalgamCost(t, s, inst, 10, "only a Maneuver and a Structure in the pile")
}

func TestCard_BoneyardAmalgam_OneLessPerUnitCard(t *testing.T) {
	s, inst := amalgamScene(t)
	s.inDiscard(0, amalgamUnits(s, 3)...)
	s.inDiscard(0, s.card("Steel Rain"), s.card("Hammerhand Bivouac"))
	amalgamCost(t, s, inst, 7, "3 unit cards in the pile")
	s.inDiscard(0, s.card("Saboteur"))
	amalgamCost(t, s, inst, 6, "a fourth unit card put in the pile afterwards")
}

func TestCard_BoneyardAmalgam_OpponentsPileNotCounted(t *testing.T) {
	s, inst := amalgamScene(t)
	s.inDiscard(1, amalgamUnits(s, 3)...)
	amalgamCost(t, s, inst, 10, "unit cards only in the opponent's pile")
}

func TestCard_BoneyardAmalgam_CountsAnotherAmalgamInThePile(t *testing.T) {
	s, inst := amalgamScene(t)
	s.inDiscard(0, s.card("Boneyard Amalgam"))
	amalgamCost(t, s, inst, 9, "a second Boneyard Amalgam (a unit card) in the pile")
}

func TestCard_BoneyardAmalgam_CostClampsAtZero(t *testing.T) {
	s, inst := amalgamScene(t)
	s.inDiscard(0, amalgamUnits(s, 12)...)
	amalgamCost(t, s, inst, 0, "12 unit cards in the pile, never negative")
	s.setPool(0, 0, 0)
	if !amalgamDeployLegal(s, inst) {
		t.Fatalf("a Deploy costing R0 is legal with an empty pool (FR-CST-001)")
	}
	s.do(amalgamDeploy(inst))
	if p := s.g.Players[0]; p.Req != 0 || p.Tac != 0 {
		t.Errorf("the pool stays R0 T0, got R%d T%d (FR-CST-001)", p.Req, p.Tac)
	}
}

func TestCard_BoneyardAmalgam_DeployPaysTheLowerCost(t *testing.T) {
	t.Run("pool of R7 pays R7", func(t *testing.T) {
		s, inst := amalgamScene(t)
		s.inDiscard(0, amalgamUnits(s, 3)...)
		s.setPool(0, 7, 0)
		if !amalgamDeployLegal(s, inst) {
			t.Fatalf("a Deploy costing R7 is legal with R7 (FR-CST-001)")
		}
		s.do(amalgamDeploy(inst))
		if p := s.g.Players[0]; p.Req != 0 || p.Tac != 0 {
			t.Errorf("the lower cost R7 is paid: pool R%d T%d, want R0 T0 (FR-CST-001)", p.Req, p.Tac)
		}
	})
	t.Run("pool of R6 is short", func(t *testing.T) {
		s, inst := amalgamScene(t)
		s.inDiscard(0, amalgamUnits(s, 3)...)
		s.setPool(0, 6, 0)
		if amalgamDeployLegal(s, inst) {
			t.Errorf("a Deploy costing R7 is not offered with R6 (FR-CST-001)")
		}
		s.reject(amalgamDeploy(inst))
	})
}

func TestFR_CST_001_ReductionOnlyFromHand(t *testing.T) {
	// Covers: FR-CST-001
	s := newScene(t)
	insts := s.inDiscard(0, s.card("Boneyard Amalgam"))
	s.inDiscard(0, amalgamUnits(s, 3)...)
	amalgamCost(t, s, insts[0], 10, "a Boneyard Amalgam in the discard pile reads its printed cost (activeZone HAND)")
}
