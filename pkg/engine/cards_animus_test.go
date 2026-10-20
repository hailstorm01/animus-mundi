package engine

import "testing"

// Card tests for Animus cards. Each is written from the card's CSV rules text
// in data/cards/units-and-structures-animus.csv and loads the real card with
// s.card. Later Animus card tasks append their tests here.

// deployCreeper gives the real Spoil Forest Creeper (cost 3 Requisition) to
// player 0 with enough Requisition and declares its deploy into Reserves.
func deployCreeper(s *scene) *Instance {
	s.t.Helper()
	inst := s.give(0, s.card("Spoil Forest Creeper"))
	s.g.Players[0].Req = 3
	s.do(Action{Kind: ActDeploy, Player: 0, Card: inst.ID, Zone: ZoneReserves})
	return inst
}

// "When this unit deploys, Scratch 1."
// Covers: FR-KWD-002m, FR-ZON-005, FR-TRG-004
func TestCard_SpoilForestCreeper_DeployScratchesOne(t *testing.T) {
	s := newScene(t)
	p := s.g.Players[0]
	top := s.onTop(0, vanilla("Top", "Next")...)
	deckBefore := len(p.Deck)
	deployCreeper(s)
	if !scratched(p, top[0]) {
		t.Errorf("deploying the Creeper scratches the top card of its controller's deck into Reserves as a face-down Scratch Company (FR-KWD-002m, FR-ZON-005)")
	}
	if n := faceDownCount(p); n != 1 {
		t.Errorf("Scratch 1 creates exactly one Scratch Company: got %d (FR-KWD-002m)", n)
	}
	if len(p.Deck) != deckBefore-1 || p.Deck[0] != top[1] {
		t.Errorf("Scratch 1 takes exactly one card, leaving the second card on top of the deck (FR-KWD-002m)")
	}
}

// Covers: FR-ZON-005
func TestCard_SpoilForestCreeper_ScratchCompanyEntersExhausted(t *testing.T) {
	s := newScene(t)
	p := s.g.Players[0]
	top := s.onTop(0, vanilla("Top")...)
	deployCreeper(s)
	for _, a := range p.Reserves {
		if a.Inst == top[0] {
			if !a.FaceDown || !a.Exhausted || a.Zone != ZoneReserves {
				t.Errorf("a Scratch Company enters face-down and Exhausted in Reserves: faceDown=%v exhausted=%v zone=%v (FR-ZON-005)", a.FaceDown, a.Exhausted, a.Zone)
			}
			return
		}
	}
	t.Errorf("the scratched card is a Scratch Company in Reserves (FR-ZON-005)")
}

// Covers: FR-TRG-004
func TestCard_SpoilForestCreeper_ItselfIsFaceUpThreeTwoInReserves(t *testing.T) {
	s := newScene(t)
	p := s.g.Players[0]
	inst := deployCreeper(s)
	for _, a := range p.Reserves {
		if a.Inst == inst {
			if a.FaceDown {
				t.Errorf("the deployed Creeper is face-up; only the scratched card is face-down")
			}
			if c := a.Printed(); c.Power != 3 || c.Health != 2 {
				t.Errorf("the Creeper is a 3/2 unit: got %d/%d", c.Power, c.Health)
			}
			if inHand(p, inst) {
				t.Errorf("the Creeper left the hand when deployed")
			}
			return
		}
	}
	t.Errorf("the Creeper is on the battlefield in Reserves after deploying (FR-TRG-004)")
}

// Covers: FR-ZON-005
func TestCard_SpoilForestCreeper_DeployPaysThreeRequisition(t *testing.T) {
	s := newScene(t)
	p := s.g.Players[0]
	deployCreeper(s)
	if p.Req != 0 {
		t.Errorf("deploying with exactly 3 Requisition spends all of it: %d left", p.Req)
	}
}

// Covers: FR-ZON-005
func TestCard_SpoilForestCreeper_ScratchOnEmptyDeckDoesNothing(t *testing.T) {
	s := newScene(t)
	p := s.g.Players[0]
	p.Deck = nil
	inst := deployCreeper(s)
	if n := faceDownCount(p); n != 0 {
		t.Errorf("Scratch 1 on an empty deck creates no Scratch Company: got %d (FR-KWD-002m)", n)
	}
	if len(p.Reserves) != 1 || p.Reserves[0].Inst != inst {
		t.Errorf("only the Creeper itself is in Reserves after a Scratch on an empty deck: %d assets", len(p.Reserves))
	}
	if s.g.Outcome.Over {
		t.Errorf("Scratch on an empty deck is not a loss: %+v (FR-KWD-002m)", s.g.Outcome)
	}
}

// Covers: FR-ZON-005, FR-TRG-003
func TestCard_SpoilForestCreeper_FaceDownCopyDoesNotTrigger(t *testing.T) {
	s := newScene(t)
	p := s.g.Players[0]
	cards := s.onTop(0, s.card("Spoil Forest Creeper"), vanilla("B")[0], vanilla("C")[0])
	s.script(0)
	s.do(Action{Kind: ActScratch, Player: 0, Asset: p.HQ.ID})
	if !scratched(p, cards[0]) {
		t.Fatalf("the Creeper on top of the deck became a face-down Scratch Company")
	}
	if n := faceDownCount(p); n != 1 {
		t.Errorf("a face-down Scratch Company has no printed abilities, so no second Scratch happens: %d face-down (FR-ZON-005)", n)
	}
	if len(p.Deck) == 0 || p.Deck[0] != cards[1] {
		t.Errorf("the face-down Creeper does not take another card from the deck (FR-ZON-005)")
	}
}
