package engine

import (
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0307: Hurricane Outrigger (Concordia unit, R2 T0, 2/1, Reserves):
// "When this unit deploys, your Headquarters gets +1 power until the end of your
// turn." The real card is loaded from data/cards. Player 0 deploys it and is the
// active player unless a case says otherwise.

// hoDeploy gives player 0 the real Outrigger and the stated pool, deploys it into
// Reserves and returns the card.
func hoDeploy(s *scene, req int) *Card {
	s.t.Helper()
	c := s.card("Hurricane Outrigger")
	inst := s.give(0, c)
	s.setPool(0, req, 0)
	s.do(Action{Kind: ActDeploy, Player: 0, Card: inst.ID, Zone: ZoneReserves})
	return c
}

func TestCard_HurricaneOutrigger_EntersReservesAsTwoOne(t *testing.T) {
	// Covers: FR-TRN-003, FR-SET-004
	s := newScene(t)
	c := hoDeploy(s, 2)
	a := assetOf(s.g, c)
	if a == nil || a.Zone != ZoneReserves || a.Controller != 0 {
		t.Fatalf("the Outrigger is on the battlefield in player 0's Reserves after deploying (FR-TRN-003)")
	}
	if p := a.Printed(); p.Power != 2 || p.Health != 1 {
		t.Errorf("the Outrigger is a 2/1: got %d/%d", p.Power, p.Health)
	}
	if a.Exhausted {
		t.Errorf("a newly deployed unit enters Readied (FR-SET-004)")
	}
	if s.g.Players[0].Req != 0 {
		t.Errorf("the Outrigger costs 2 Requisition: %d left", s.g.Players[0].Req)
	}

	s2 := newScene(t)
	c2 := s2.card("Hurricane Outrigger")
	inst := s2.give(0, c2)
	s2.setPool(0, 2, 0)
	s2.reject(Action{Kind: ActDeploy, Player: 0, Card: inst.ID, Zone: ZoneMainTheater})
	if !inHand(s2.g.Players[0], inst) || assetOf(s2.g, c2) != nil {
		t.Errorf("deploying the Outrigger to the Main Theater is rejected and the card stays in hand (FR-TRN-003)")
	}

	s3 := newScene(t)
	c3 := s3.card("Hurricane Outrigger")
	inst3 := s3.give(0, c3)
	s3.setPool(0, 1, 0)
	s3.reject(Action{Kind: ActDeploy, Player: 0, Card: inst3.ID, Zone: ZoneReserves})
	if !inHand(s3.g.Players[0], inst3) || assetOf(s3.g, c3) != nil {
		t.Errorf("with 1 Requisition the Outrigger (cost 2) cannot be deployed (FR-TRN-003)")
	}
}

func TestCard_HurricaneOutrigger_HeadquartersGetsOnePower(t *testing.T) {
	// Covers: FR-LAY-001
	s := newScene(t)
	ally := s.put(0, ZoneMainTheater, unit("Ally", 2, 5, ability.Keywords{}))
	hq0, hq1 := vrHQ(s, 0), vrHQ(s, 1)
	from := len(s.g.Log)
	c := hoDeploy(s, 2)
	if got := s.g.Power(hq0); got != 4 {
		t.Errorf("player 0's Headquarters has power 4 (3 + 1): got %d (FR-LAY-001)", got)
	}
	if got := s.g.Power(hq1); got != 3 {
		t.Errorf("player 1's Headquarters stays at power 3: got %d (FR-LAY-001)", got)
	}
	if got := s.g.Power(assetOf(s.g, c)); got != 2 {
		t.Errorf("the Outrigger itself stays at power 2: got %d", got)
	}
	if got := s.g.Power(ally); got != 2 {
		t.Errorf("another allied unit keeps its power 2: got %d", got)
	}
	evs := effectEvents(s.g, from, EventPower)
	if len(evs) != 1 || evs[0].Target != hq0.ID || evs[0].Amount != 1 {
		t.Errorf("exactly one power event of +1 naming player 0's Headquarters: %+v", evs)
	}
	if n := len(effectEvents(s.g, from, EventMaxHealth)); n != 0 {
		t.Errorf("no max-health event is logged: %d", n)
	}
	if n := len(s.asked(0)) + len(s.asked(1)); n != 0 {
		t.Errorf("nothing is asked: %d decisions", n)
	}
	s.wantDamage(hq0, 0)
	s.wantDamage(hq1, 0)
	s.wantDamage(ally, 0)
	if got := totalDamage(s.g); got != 0 {
		t.Errorf("no damage is dealt: total %d", got)
	}
}

func TestCard_HurricaneOutrigger_ExpiresAtEndOfTurn(t *testing.T) {
	// Covers: FR-LAY-003, FR-TRN-001
	s := newScene(t)
	hoDeploy(s, 2)
	hq0 := vrHQ(s, 0)
	if got := s.g.Power(hq0); got != 4 {
		t.Fatalf("setup: the Headquarters has power 4 this turn, got %d", got)
	}
	s.endTurn()
	if got := s.g.Power(hq0); got != 3 {
		t.Errorf("the +1 ends with player 0's turn: power %d, want 3 (FR-LAY-003, FR-TRN-001)", got)
	}
	s.endTurn()
	if got := s.g.Power(hq0); got != 3 {
		t.Errorf("the +1 does not return after player 1's turn: power %d, want 3 (FR-LAY-003)", got)
	}
}

func TestCard_HurricaneOutrigger_HeadquartersAttacksForFour(t *testing.T) {
	// Covers: FR-CBT-002
	s := newScene(t)
	hoDeploy(s, 2)
	hq0, hq1 := vrHQ(s, 0), vrHQ(s, 1)
	s.do(s.attack(hq0, hq1))
	s.wantDamage(hq1, 4)
	s.wantDamage(hq0, 3)
}

func TestCard_HurricaneOutrigger_TwoDeploysStack(t *testing.T) {
	// Covers: FR-LAY-001
	s := newScene(t)
	c1, c2 := s.card("Hurricane Outrigger"), s.card("Hurricane Outrigger")
	i1, i2 := s.give(0, c1), s.give(0, c2)
	s.setPool(0, 4, 0)
	s.do(Action{Kind: ActDeploy, Player: 0, Card: i1.ID, Zone: ZoneReserves})
	s.do(Action{Kind: ActDeploy, Player: 0, Card: i2.ID, Zone: ZoneReserves})
	if got := s.g.Power(vrHQ(s, 0)); got != 5 {
		t.Errorf("two Outriggers give the Headquarters +2: power %d, want 5 (FR-LAY-001)", got)
	}
}

func TestCard_HurricaneOutrigger_OpponentsOutriggerBoostsTheirHeadquarters(t *testing.T) {
	// Covers: FR-LAY-001
	s := newScene(t)
	s.endTurn()
	c := s.card("Hurricane Outrigger")
	inst := s.give(1, c)
	s.setPool(1, 2, 0)
	s.do(Action{Kind: ActDeploy, Player: 1, Card: inst.ID, Zone: ZoneReserves})
	if got := s.g.Power(vrHQ(s, 1)); got != 4 {
		t.Errorf("player 1's Headquarters has power 4: got %d (FR-LAY-001)", got)
	}
	if got := s.g.Power(vrHQ(s, 0)); got != 3 {
		t.Errorf("player 0's Headquarters stays at power 3: got %d (FR-LAY-001)", got)
	}
}

func TestCard_HurricaneOutrigger_OtherDeploysDoNothing(t *testing.T) {
	// Covers: FR-TRN-003
	s := newScene(t)
	s.put(0, ZoneReserves, s.card("Hurricane Outrigger"))
	s.deployTo(0, unit("Vanilla", 1, 1, ability.Keywords{}), ZoneReserves)
	if got := s.g.Power(vrHQ(s, 0)); got != 3 {
		t.Errorf("only the Outrigger's own deploy triggers it: player 0's Headquarters power %d, want 3 (FR-TRN-003)", got)
	}
	if got := s.g.Power(vrHQ(s, 1)); got != 3 {
		t.Errorf("player 1's Headquarters power %d, want 3", got)
	}
}

func TestCard_HurricaneOutrigger_SourceGoneStillBoosts(t *testing.T) {
	// Covers: FR-ZON-004
	s := newScene(t)
	c := s.card("Hurricane Outrigger")
	s.ptPlaceHook(func(int, StackItem) {
		if a := assetOf(s.g, c); a != nil {
			s.g.removeFromBattlefield(a)
			s.g.putCard(a.Inst, ZoneDiscard, false)
		}
	})
	inst := s.give(0, c)
	s.setPool(0, 2, 0)
	s.do(Action{Kind: ActDeploy, Player: 0, Card: inst.ID, Zone: ZoneReserves})
	if assetOf(s.g, c) != nil {
		t.Fatalf("setup: the Outrigger should have left before its trigger resolved")
	}
	if got := s.g.Power(vrHQ(s, 0)); got != 4 {
		t.Errorf("the trigger still resolves from last-known state: power %d, want 4 (FR-ZON-004)", got)
	}
	if !inDiscardPile(s.g.Players[0], inst) {
		t.Errorf("the Outrigger's card is in its owner's discard pile (FR-ZON-004)")
	}
}

func TestCard_HurricaneOutrigger_FaceDownScratchDoesNothing(t *testing.T) {
	// Covers: FR-ZON-005
	s := newScene(t)
	s.onTop(0, s.card("Hurricane Outrigger"))
	s.do(enactAction(glEnact(s)))
	p := s.g.Players[0]
	if n := faceDownCount(p); n != 1 {
		t.Fatalf("setup: Scratch 1 makes one face-down Scratch Company, got %d", n)
	}
	if got := s.g.Power(vrHQ(s, 0)); got != 3 {
		t.Errorf("a face-down Scratch Company has no printed ability but Deploys Exhausted: player 0's Headquarters power %d, want 3 (FR-ZON-005)", got)
	}
	if got := s.g.Power(vrHQ(s, 1)); got != 3 {
		t.Errorf("player 1's Headquarters power %d, want 3", got)
	}
}
