package engine

import (
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0305: Swirling Dustdevil, "This unit costs R1 less to deploy for each
// maneuver card in your discard pile." (FR-CST-001). The real cards are loaded from
// data/cards. Printed cost is R8 T0. Player 0 is the active player. It reuses
// amalgamCost, amalgamDeployLegal and amalgamDeploy from cards_path_amalgam_test.go.

// dustScene is a scene with a Swirling Dustdevil in player 0's hand.
func dustScene(t *testing.T) (*scene, *Instance) {
	t.Helper()
	s := newScene(t)
	return s, s.give(0, s.card("Swirling Dustdevil"))
}

// dustManeuvers returns n Maneuver cards of mixed factions for a discard pile.
func dustManeuvers(s *scene, n int) []*Card {
	names := []string{"Steel Rain", "Fortify", "Growing Legion", "Forced Mutation"}
	cs := make([]*Card, n)
	for i := range cs {
		cs[i] = s.card(names[i%len(names)])
	}
	return cs
}

// dustAssetOf returns the asset in player 0's Reserves whose card is inst.
func dustAssetOf(s *scene, inst *Instance) *Asset {
	for _, a := range s.g.Players[0].Reserves {
		if a.Inst == inst {
			return a
		}
	}
	return nil
}

func TestCard_SwirlingDustdevil_PrintedCostWithNoManeuverCards(t *testing.T) {
	s, inst := dustScene(t)
	amalgamCost(t, s, inst, 8, "an empty discard pile")
	s.inDiscard(0, s.card("Saboteur"), s.card("Hammerhand Bivouac"), s.card("Swirling Dustdevil"))
	amalgamCost(t, s, inst, 8, "only unit and structure cards in the pile, a second Swirling Dustdevil included")
}

func TestCard_SwirlingDustdevil_OneLessPerManeuverCard(t *testing.T) {
	s, inst := dustScene(t)
	s.inDiscard(0, dustManeuvers(s, 3)...)
	s.inDiscard(0, s.card("Saboteur"), s.card("Hammerhand Bivouac"))
	amalgamCost(t, s, inst, 5, "3 Maneuver cards of different factions in the pile")
	s.inDiscard(0, s.card("Fortify"))
	amalgamCost(t, s, inst, 4, "a fourth Maneuver card put in the pile afterwards")
	s.inDiscard(0, s.card("Fortify"), s.card("Fortify"))
	amalgamCost(t, s, inst, 2, "copies of one name count each")
}

func TestCard_SwirlingDustdevil_OpponentsPileNotCounted(t *testing.T) {
	s, inst := dustScene(t)
	s.inDiscard(1, dustManeuvers(s, 3)...)
	amalgamCost(t, s, inst, 8, "Maneuver cards only in the opposing pile")
	s.inDiscard(0, dustManeuvers(s, 2)...)
	s.inDiscard(1, dustManeuvers(s, 2)...)
	amalgamCost(t, s, inst, 6, "2 Maneuver cards in each pile count only player 0's")
}

func TestCard_SwirlingDustdevil_ManeuverCardsElsewhereNotCounted(t *testing.T) {
	s, inst := dustScene(t)
	for _, c := range dustManeuvers(s, 2) {
		s.give(0, c)
	}
	s.onTop(0, dustManeuvers(s, 2)...)
	fd := s.putScratch(0)
	fd.Inst.Card = s.card("Steel Rain")
	amalgamCost(t, s, inst, 8, "Maneuver cards in hand, in the deck and as a face-down Scratch Company are not in the discard pile")
}

func TestCard_SwirlingDustdevil_KilledScratchCompanyManeuverCounts(t *testing.T) {
	// Covers: FR-ZON-005
	s, inst := dustScene(t)
	down := s.putScratch(0)
	down.Inst.Card = s.card("Steel Rain")
	s.moveTo(down, ZoneMainTheater)
	s.endTurn()
	hitter := s.put(1, ZoneMainTheater, unit("Hitter", 5, 9, ability.Keywords{}))
	s.do(s.attack(hitter, down))
	if s.alive(down) {
		t.Fatalf("precondition: the face-down Scratch Company is destroyed by the attack")
	}
	p := s.g.Players[0]
	if !inDiscardPile(p, down.Inst) || down.Inst.Card.Type != TypeManeuver || down.Inst.Card.Name != "Steel Rain" {
		t.Fatalf("a destroyed Scratch Company's card goes to the pile with its printed characteristics, a Maneuver (FR-ZON-005)")
	}
	amalgamCost(t, s, inst, 7, "the destroyed Scratch Company's Maneuver card is in the pile")
}

func TestCard_SwirlingDustdevil_EnactedManeuverCounts(t *testing.T) {
	// Covers: FR-MNV-001
	s, inst := dustScene(t)
	amalgamCost(t, s, inst, 8, "before enacting")
	m := s.give(0, s.card("Growing Legion"))
	s.setPool(0, 0, 1)
	s.do(enactAction(m))
	if len(s.g.Stack) != 0 || !inDiscardPile(s.g.Players[0], m) {
		t.Fatalf("the resolved Maneuver is in its owner's discard pile (FR-MNV-001)")
	}
	amalgamCost(t, s, inst, 7, "an enacted Growing Legion, once resolved, is a Maneuver card in the pile")
}

func TestCard_SwirlingDustdevil_CostRisesWhenAManeuverLeavesThePile(t *testing.T) {
	s, inst := dustScene(t)
	scav := s.put(0, ZoneMainTheater, s.card("Boneyard Scavenger"))
	s.setPool(0, 0, 1)
	s.inDiscard(0, dustManeuvers(s, 3)...)
	amalgamCost(t, s, inst, 5, "3 Maneuver cards in the pile")
	s.script(0, Answer{Picks: []int{0}})
	s.do(activate(scav, 0))
	amalgamCost(t, s, inst, 6, "one Maneuver card returned to hand leaves 2 in the pile")
}

func TestCard_SwirlingDustdevil_CostClampsAtZero(t *testing.T) {
	for _, n := range []int{8, 10} {
		s, inst := dustScene(t)
		s.inDiscard(0, dustManeuvers(s, n)...)
		amalgamCost(t, s, inst, 0, "Maneuver cards in the pile, never negative")
		s.setPool(0, 0, 0)
		if !amalgamDeployLegal(s, inst) {
			t.Fatalf("%d Maneuver cards: a Deploy costing R0 is legal with an empty pool (FR-CST-001)", n)
		}
		s.do(amalgamDeploy(inst))
		if p := s.g.Players[0]; p.Req != 0 || p.Tac != 0 {
			t.Errorf("%d Maneuver cards: the pool stays R0 T0, got R%d T%d (FR-CST-001)", n, p.Req, p.Tac)
		}
	}
}

func TestCard_SwirlingDustdevil_DeployPaysTheLowerCost(t *testing.T) {
	// Covers: FR-TRN-003
	t.Run("pool of R5 pays R5", func(t *testing.T) {
		s, inst := dustScene(t)
		s.inDiscard(0, dustManeuvers(s, 3)...)
		s.setPool(0, 5, 0)
		if !amalgamDeployLegal(s, inst) {
			t.Fatalf("a Deploy costing R5 is offered with R5 (FR-CST-001)")
		}
		s.do(amalgamDeploy(inst))
		if p := s.g.Players[0]; p.Req != 0 || p.Tac != 0 {
			t.Errorf("the lower cost R5 is paid: pool R%d T%d, want R0 T0 (FR-TRN-003)", p.Req, p.Tac)
		}
	})
	t.Run("pool of R8 leaves R3", func(t *testing.T) {
		s, inst := dustScene(t)
		s.inDiscard(0, dustManeuvers(s, 3)...)
		s.setPool(0, 8, 0)
		s.do(amalgamDeploy(inst))
		if p := s.g.Players[0]; p.Req != 3 || p.Tac != 0 {
			t.Errorf("only R5 is paid from R8: pool R%d T%d, want R3 T0 (FR-TRN-003)", p.Req, p.Tac)
		}
	})
	t.Run("pool of R4 T4 is short", func(t *testing.T) {
		s, inst := dustScene(t)
		s.inDiscard(0, dustManeuvers(s, 3)...)
		s.setPool(0, 4, 4)
		if amalgamDeployLegal(s, inst) {
			t.Errorf("a Deploy costing R5 is not offered with R4 T4: Tactics cannot pay Requisition (FR-TRN-003)")
		}
		pile := len(s.g.Players[0].Discard)
		s.reject(amalgamDeploy(inst))
		p := s.g.Players[0]
		if !inHand(p, inst) || p.Req != 4 || p.Tac != 4 || len(p.Discard) != pile || dustAssetOf(s, inst) != nil {
			t.Errorf("a rejected Deploy leaves the card in hand and the pool at R4 T4: R%d T%d (FR-TRN-003)", p.Req, p.Tac)
		}
	})
}

func TestCard_SwirlingDustdevil_EntersReservesAsSixFive(t *testing.T) {
	// Covers: FR-TRN-003, FR-SET-004
	s, inst := dustScene(t)
	s.inDiscard(0, dustManeuvers(s, 3)...)
	s.setPool(0, 5, 0)
	s.reject(Action{Kind: ActDeploy, Player: 0, Card: inst.ID, Zone: ZoneMainTheater})
	s.do(amalgamDeploy(inst))
	a := dustAssetOf(s, inst)
	if a == nil || a.Zone != ZoneReserves {
		t.Fatalf("the deployed Dustdevil is an asset in player 0's Reserves (FR-TRN-003)")
	}
	if a.Exhausted {
		t.Errorf("it enters Readied (FR-SET-004)")
	}
	if pw, h := s.g.Power(a), s.g.MaxHealth(a); pw != 6 || h != 5 {
		t.Errorf("it is a %d/%d unit, want 6/5", pw, h)
	}
	if c := a.Printed(); c.ReqYield != 2 || c.TacYield != 2 {
		t.Errorf("yields R%d T%d, want R2 T2", c.ReqYield, c.TacYield)
	}
	if n := len(s.g.Players[0].Discard); n != 3 {
		t.Errorf("the pile still holds its 3 Maneuver cards, has %d (FR-CST-001)", n)
	}
}

func TestCard_SwirlingDustdevil_ReducesOnlyItself(t *testing.T) {
	check := func(t *testing.T, s *scene, why string) {
		t.Helper()
		other := s.give(0, s.card("Saboteur"))
		man := s.give(0, s.card("Steel Rain"))
		for _, i := range []*Instance{other, man} {
			req, tac := s.g.EffectiveCost(i)
			if req != i.Card.ReqCost || tac != i.Card.TacCost {
				t.Errorf("%s: %s costs R%d T%d, want its printed R%d T%d (FR-CST-001)", why, i.Card.Name, req, tac, i.Card.ReqCost, i.Card.TacCost)
			}
		}
	}
	s, _ := dustScene(t)
	s.inDiscard(0, dustManeuvers(s, 3)...)
	second := s.give(0, s.card("Swirling Dustdevil"))
	amalgamCost(t, s, second, 5, "a second Dustdevil in hand reads its own text")
	check(t, s, "a Dustdevil in hand")
	s2 := newScene(t)
	s2.put(0, ZoneMainTheater, s2.card("Swirling Dustdevil"))
	s2.inDiscard(0, dustManeuvers(s2, 3)...)
	check(t, s2, "a Dustdevil on the battlefield")
}

func TestCard_SwirlingDustdevil_PrintedCostInTheDiscardPile(t *testing.T) {
	// Covers: FR-CST-001
	s := newScene(t)
	insts := s.inDiscard(0, s.card("Swirling Dustdevil"))
	s.inDiscard(0, dustManeuvers(s, 3)...)
	amalgamCost(t, s, insts[0], 8, "a Dustdevil in the discard pile reads its printed cost, the reduction is hand-active")
}
