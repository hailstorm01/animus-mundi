package engine

import "testing"

// Tests for T-0215 (FR-TRN-003, FR-TRN-003a, directive #128): DiscardDeployCost,
// the reader a client plans a discard deploy with. It uses T-0214's ddUnit,
// whose permission costs killing an allied unit. Player 0 owns every card.

func TestDiscardDeployCost_PermissionAndPayableCosts(t *testing.T) {
	s := newScene(t)
	inst := s.inDiscard(0, ddUnit("Phoenix", 2, 1))[0]
	if req, tac, ok := s.g.DiscardDeployCost(inst); req != 2 || tac != 1 || ok {
		t.Errorf("with no allied unit to kill: got R%d T%d %v, want R2 T1 false (FR-TRN-003a)", req, tac, ok)
	}
	s.put(0, ZoneMainTheater, unit("Ally", 1, 5, Keywords{}))
	if req, tac, ok := s.g.DiscardDeployCost(inst); req != 2 || tac != 1 || !ok {
		t.Errorf("with an allied unit to kill: got R%d T%d %v, want R2 T1 true (FR-TRN-003)", req, tac, ok)
	}
	s.setPool(0, 0, 0)
	if _, _, ok := s.g.DiscardDeployCost(inst); !ok {
		t.Errorf("an empty pool doesn't change ok: it doesn't check the resources against the pool")
	}
}

func TestDiscardDeployCost_OnlyAPermittedCardInItsOwnersPile(t *testing.T) {
	s := newScene(t)
	s.put(0, ZoneMainTheater, unit("Ally", 1, 5, Keywords{}))
	inHand := s.give(0, ddUnit("Phoenix", 2, 0))
	if _, _, ok := s.g.DiscardDeployCost(inHand); ok {
		t.Errorf("a permitted card in hand: ok true, want false")
	}
	plain := unit("Plain", 1, 1, Keywords{})
	plain.ReqCost = 1
	if _, _, ok := s.g.DiscardDeployCost(s.inDiscard(0, plain)[0]); ok {
		t.Errorf("a discard card with no permission: ok true, want false (FR-TRN-003)")
	}
}
