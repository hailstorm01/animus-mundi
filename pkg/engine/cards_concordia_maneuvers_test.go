package engine

import (
	"animus-mundi/pkg/engine/ability"
	"testing"
)

// Tests for T-0299 (FR-KWD-002m, FR-ZON-005, FR-MNV-001): the Concordia
// Maneuver Growing Legion, loaded from data/cards. Player 0 enacts with exactly
// R0 T1; player 1 is the opponent.
//
// Growing Legion: "Scratch 1."

// glEnact gives player 0 the real Growing Legion and exactly R0 T1.
func glEnact(s *scene) *Instance {
	inst := s.give(0, s.card("Growing Legion"))
	s.setPool(0, 0, 1)
	return inst
}

func TestCard_GrowingLegion_ScratchesOne(t *testing.T) {
	// Covers: FR-ZON-005
	s := newScene(t)
	top := s.onTop(0, unit("TopA", 1, 1, ability.Keywords{}), unit("TopB", 1, 1, ability.Keywords{}))
	inst := glEnact(s)
	s.do(enactAction(inst))
	p := s.g.Players[0]
	if n := faceDownCount(p); n != 1 {
		t.Fatalf("Scratch 1 makes exactly one Scratch Company: %d face-down units, want 1 (FR-KWD-002m)", n)
	}
	if !scratched(p, top[0]) {
		t.Errorf("the top card of the deck becomes the Scratch Company (FR-KWD-002m)")
	}
	for _, a := range p.Reserves {
		if a.FaceDown && !a.Exhausted {
			t.Errorf("a Scratch Company deploys Exhausted in Reserves (FR-ZON-005, KIM-S01)")
		}
	}
	if len(p.Deck) == 0 || p.Deck[0] != top[1] {
		t.Errorf("the second card is now on top of the deck (FR-KWD-002m)")
	}
	if n := faceDownCount(s.g.Players[1]); n != 0 {
		t.Errorf("the opponent has no Scratch Company, got %d (FR-KWD-002m)", n)
	}
	if p.Req != 0 || p.Tac != 0 {
		t.Errorf("R0 T1 is paid: pool R%d T%d, want R0 T0 (FR-MNV-001)", p.Req, p.Tac)
	}
	if !inDiscardPile(p, inst) {
		t.Errorf("the resolved Maneuver ends in its owner's discard pile (FR-MNV-001)")
	}
}

func TestCard_GrowingLegion_DoesNotUseTheHeadquarters(t *testing.T) {
	// Covers: FR-KWD-002m
	s := newScene(t)
	p := s.g.Players[0]
	if p.HQ.Exhausted {
		t.Fatalf("setup: the Headquarters starts Readied")
	}
	s.do(enactAction(glEnact(s)))
	if p.HQ.Exhausted {
		t.Errorf("Scratch 1 does not Exhaust the Headquarters: it is still Readied (FR-KWD-002m)")
	}
}

func TestCard_GrowingLegion_EnactableWithAnExhaustedHeadquarters(t *testing.T) {
	// Covers: FR-KWD-002m
	s := newScene(t)
	p := s.g.Players[0]
	p.HQ.Exhausted = true
	inst := glEnact(s)
	if !offersEnact(s.g, inst) {
		t.Fatalf("an Exhausted Headquarters does not stop the Enact (FR-KWD-002m)")
	}
	s.do(enactAction(inst))
	if n := faceDownCount(p); n != 1 {
		t.Errorf("it still scratches one: %d face-down units, want 1 (FR-KWD-002m)", n)
	}
}

func TestCard_GrowingLegion_EmptyDeckScratchesNothingAndDoesNotLose(t *testing.T) {
	// Covers: FR-KWD-002m
	s := newScene(t)
	p := s.g.Players[0]
	p.Deck = nil
	inst := glEnact(s)
	if !offersEnact(s.g, inst) {
		t.Fatalf("it can be enacted with a 0-card deck (FR-KWD-002m)")
	}
	s.do(enactAction(inst))
	if n := faceDownCount(p); n != 0 {
		t.Errorf("an empty deck makes no unit: %d face-down units, want 0 (FR-KWD-002m)", n)
	}
	if s.g.Outcome.Over {
		t.Errorf("Scratching from an empty deck does not end the match (FR-KWD-002m)")
	}
}
