package engine

import "testing"

// Chainbreaker Irregulars: "Unbowed: Draw a card, then discard a card."
// Unit, Infantry, Chainbreaker, 3 Requisition, Reserves, 2/5. Unbowed fires
// when the asset takes damage (FR-KWD-002a). The draw and the discard are the
// controller's. Choosing a card from your own Hand is not an inspection, so a
// one-card choice is automatic and a timeout discards the card most recently
// added (SVY-R16, FR-MOD-001).

// irregularsScene puts the real card in player 1's Main Theater, gives player 1
// a hand of the named vanilla cards (the last is most recently added) and puts
// the named vanilla cards on top of player 1's deck (the first on top). It
// returns the scene, the asset, the hand instances and the deck instances.
func irregularsScene(t *testing.T, hand, deck []string) (*scene, *Asset, []*Instance, []*Instance) {
	t.Helper()
	s := newScene(t)
	s.g.Players[1].Hand = nil
	hs := make([]*Instance, len(hand))
	for i, c := range vanilla(hand...) {
		hs[i] = s.give(1, c)
	}
	ds := s.onTop(1, vanilla(deck...)...)
	a := s.put(1, ZoneMainTheater, s.card("Chainbreaker Irregulars"))
	return s, a, hs, ds
}

func topOfDiscard(p *Player) *Instance {
	if len(p.Discard) == 0 {
		return nil
	}
	return p.Discard[len(p.Discard)-1]
}

// Control: a 2/5 that enters Reserves for 3 Requisition and does nothing on deploy.
func TestCard_ChainbreakerIrregulars_EntersReservesAsTwoFiveForThreeRequisitionAndDoesNothing(t *testing.T) {
	s := newScene(t)
	s.script(0)
	s.script(1)
	hand, deck := len(s.g.Players[0].Hand), len(s.g.Players[0].Deck)
	c := deployPath(s, "Chainbreaker Irregulars", 3)
	a := assetOf(s.g, c)
	if a == nil || a.Zone != ZoneReserves {
		t.Fatalf("Chainbreaker Irregulars is on the battlefield in Reserves after deploying")
	}
	if p := a.Printed(); p.Power != 2 || p.Health != 5 {
		t.Errorf("Chainbreaker Irregulars is a 2/5: got %d/%d", p.Power, p.Health)
	}
	if s.g.Players[0].Req != 0 {
		t.Errorf("Chainbreaker Irregulars costs 3 Requisition: %d left", s.g.Players[0].Req)
	}
	if got := len(s.g.Players[0].Deck); got != deck {
		t.Errorf("Unbowed is not a deploy trigger: nothing is drawn on deploy, deck %d -> %d", deck, got)
	}
	if got := len(s.g.Players[0].Hand); got != hand {
		t.Errorf("deploying removes nothing but the card itself from hand: hand is %d, want %d", got, hand)
	}
	if n := len(s.asked(0)) + len(s.asked(1)); n != 0 {
		t.Errorf("no question is asked on deploy: asked %d times", n)
	}
}

func TestCard_ChainbreakerIrregulars_DamageDrawsThenDiscardsTheCardTheControllerPicks(t *testing.T) {
	s, def, hand, deck := irregularsScene(t, []string{"A", "B"}, []string{"D"})
	att := s.put(0, ZoneMainTheater, unit("Hitter", 2, 9, Keywords{}))
	s.script(0)
	// The hand is A, B, then the drawn D. Pick B.
	s.script(1, Answer{Picks: []int{1}})
	s.do(s.attack(att, def))
	s.wantDamage(def, 2)
	asks := s.asked(1)
	if len(asks) != 1 {
		t.Fatalf("the controller is asked once, for the discard, after the draw: asked %d times (FR-KWD-002a)", len(asks))
	}
	if asks[0].Kind != DecideHand {
		t.Errorf("kind = %v, want Hand (SVY-R16)", asks[0].Kind)
	}
	if len(asks[0].Options) != 3 {
		t.Errorf("the discard is chosen from the 3 cards in hand, the drawn card included: got %d options", len(asks[0].Options))
	}
	if len(s.asked(0)) != 0 {
		t.Errorf("the damaging player is never asked: asked %d times", len(s.asked(0)))
	}
	p := s.g.Players[1]
	if topOfDiscard(p) != hand[1] {
		t.Errorf("the picked card B is discarded (FR-KWD-002a)")
	}
	if len(p.Hand) != 2 || p.Hand[0] != hand[0] || p.Hand[1] != deck[0] {
		t.Errorf("hand is A and the drawn D after drawing one card and discarding B, got %d cards", len(p.Hand))
	}
}

func TestCard_ChainbreakerIrregulars_MayDiscardTheCardJustDrawn(t *testing.T) {
	s, def, hand, deck := irregularsScene(t, []string{"A", "B"}, []string{"D"})
	att := s.put(0, ZoneMainTheater, unit("Hitter", 2, 9, Keywords{}))
	s.script(0)
	s.script(1, Answer{Picks: []int{2}})
	s.do(s.attack(att, def))
	p := s.g.Players[1]
	if topOfDiscard(p) != deck[0] {
		t.Errorf("the card just drawn is a legal discard (FR-KWD-002a)")
	}
	if len(p.Hand) != 2 || p.Hand[0] != hand[0] || p.Hand[1] != hand[1] {
		t.Errorf("hand is unchanged, A and B, got %d cards", len(p.Hand))
	}
}

// Covers: FR-MOD-001
func TestCard_ChainbreakerIrregulars_TimeoutDiscardsTheCardJustDrawn(t *testing.T) {
	s, def, hand, deck := irregularsScene(t, []string{"A", "B"}, []string{"D"})
	att := s.put(0, ZoneMainTheater, unit("Hitter", 2, 9, Keywords{}))
	s.script(0)
	s.script(1, Answer{Timeout: true})
	s.do(s.attack(att, def))
	p := s.g.Players[1]
	if topOfDiscard(p) != deck[0] {
		t.Errorf("a timeout discards the card just drawn (SVY-R16)")
	}
	if len(p.Hand) != 2 || p.Hand[0] != hand[0] || p.Hand[1] != hand[1] {
		t.Errorf("hand is A and B after the timeout discard, got %d cards", len(p.Hand))
	}
}

// Covers: FR-MOD-001
func TestCard_ChainbreakerIrregulars_EmptyHandDiscardsTheDrawnCardWithoutAsking(t *testing.T) {
	s, def, _, deck := irregularsScene(t, nil, []string{"D"})
	att := s.put(0, ZoneMainTheater, unit("Hitter", 2, 9, Keywords{}))
	s.script(0)
	s.script(1)
	s.do(s.attack(att, def))
	if n := len(s.asked(1)); n != 0 {
		t.Errorf("the drawn card is the only option, so nothing is asked: asked %d times (FR-MOD-001 §1, SVY-R16)", n)
	}
	p := s.g.Players[1]
	if topOfDiscard(p) != deck[0] || len(p.Hand) != 0 {
		t.Errorf("the drawn card is discarded and the hand ends empty, got hand %d (SVY-R16)", len(p.Hand))
	}
}

func TestCard_ChainbreakerIrregulars_LethalDamageStillDrawsAndDiscardsAfterTheDeathSweep(t *testing.T) {
	s, def, hand, deck := irregularsScene(t, []string{"A", "B"}, []string{"D"})
	att := s.put(0, ZoneMainTheater, unit("Hitter", 5, 9, Keywords{}))
	inst := def.Inst
	s.script(0)
	s.script(1, Answer{Picks: []int{0}})
	s.do(s.attack(att, def))
	if s.alive(def) || !inZone(s.g.Players[1].Discard, inst) {
		t.Fatalf("precondition: 5 damage destroys the 2/5 Irregulars into its owner's Discard")
	}
	p := s.g.Players[1]
	if len(s.asked(1)) != 1 {
		t.Fatalf("lethal damage still fires Unbowed: the controller is asked once for the discard, asked %d times (SVY-R13)", len(s.asked(1)))
	}
	if !inZone(p.Discard, hand[0]) {
		t.Errorf("the picked card A is discarded when the trigger resolves after the death sweep (SVY-R13)")
	}
	if len(p.Hand) != 2 || p.Hand[0] != hand[1] || p.Hand[1] != deck[0] {
		t.Errorf("hand is B and the drawn D, got %d cards (SVY-R13)", len(p.Hand))
	}
	if len(s.g.Stack) != 0 {
		t.Errorf("the stack is empty once the trigger resolves: %d items", len(s.g.Stack))
	}
	if err := s.g.CheckInvariants(); err != nil {
		t.Errorf("state after the trigger: %v", err)
	}
}

func TestCard_ChainbreakerIrregulars_EachSeparateDamageInstanceFiresAgain(t *testing.T) {
	s, def, _, _ := irregularsScene(t, []string{"A"}, []string{"D1", "D2"})
	att1 := s.put(0, ZoneMainTheater, unit("Hitter1", 1, 9, Keywords{}))
	att2 := s.put(0, ZoneMainTheater, unit("Hitter2", 1, 9, Keywords{}))
	s.script(0)
	s.script(1, Answer{Timeout: true}, Answer{Timeout: true})
	deckBefore := len(s.g.Players[1].Deck)
	s.do(s.attack(att1, def))
	s.do(s.attack(att2, def))
	if got := deckBefore - len(s.g.Players[1].Deck); got != 2 {
		t.Errorf("two separate damage instances draw 2 cards, drew %d (FR-KWD-002a)", got)
	}
	if n := len(s.asked(1)); n != 2 {
		t.Errorf("each damage instance fires its own Unbowed discard: asked %d times, want 2 (FR-KWD-002a)", n)
	}
	if got := len(s.g.Players[1].Hand); got != 1 {
		t.Errorf("each instance draws one and discards one, so the hand stays at 1: got %d", got)
	}
}

func TestCard_ChainbreakerIrregulars_OnlyTheControllersHandAndDeckAreUsed(t *testing.T) {
	s, def, _, _ := irregularsScene(t, []string{"A", "B"}, []string{"D"})
	att := s.put(0, ZoneMainTheater, unit("Hitter", 2, 9, Keywords{}))
	s.script(0)
	s.script(1, Answer{Timeout: true})
	p0 := s.g.Players[0]
	h0, d0, x0 := len(p0.Hand), len(p0.Deck), len(p0.Discard)
	s.do(s.attack(att, def))
	if len(p0.Hand) != h0 || len(p0.Deck) != d0 || len(p0.Discard) != x0 {
		t.Errorf("the attacker's hand, deck and discard are untouched: Unbowed belongs to the Irregulars' controller (FR-TRG-001)")
	}
}
