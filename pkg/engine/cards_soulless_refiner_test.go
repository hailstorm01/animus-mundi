package engine

import (
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0315 (FR-TRN-003, FR-SET-004, FR-ACT-001, FR-ACT-002, FR-MOD-001,
// FR-WIN-001, FR-LAY-002, FR-ZON-004, FR-TRN-003a): Dolus Refiner (Soulless
// structure, R4 T0, Main Theater, 0/9), loaded from data/cards: "Exhaust: Draw 2
// cards, then discard 2 cards." Player 0 controls it and is the active player.

// drfHand empties player 0's hand and gives it the vanilla cards H1 and H2.
func drfHand(s *scene) []*Instance {
	s.g.Players[0].Hand = nil
	var hand []*Instance
	for _, c := range vanilla("H1", "H2") {
		hand = append(hand, s.give(0, c))
	}
	return hand
}

// drfScene is a scene with a Readied Refiner in player 0's Main Theater, pool
// R0 T0, a hand of H1 and H2 and D1, D2, D3 on top of player 0's deck.
func drfScene(t *testing.T) (*scene, *Asset, []*Instance, []*Instance) {
	t.Helper()
	s := newScene(t)
	hand := drfHand(s)
	deck := s.onTop(0, vanilla("D1", "D2", "D3")...)
	r := s.put(0, ZoneMainTheater, s.card("Dolus Refiner"))
	s.setPool(0, 0, 0)
	return s, r, hand, deck
}

// drfTapCard is a copy of the real card (the shared definition is not changed)
// whose triggers also hold one synthetic HookExhausted trigger killing the
// Refiner. It is placed on top of the ability and resolves first.
func drfTapCard(s *scene) *Card {
	s.t.Helper()
	c := *s.card("Dolus Refiner")
	if c.Ability == nil {
		s.t.Fatalf("the Refiner has no executable definition")
	}
	d := *c.Ability
	d.Triggers = append(append([]ability.Trigger(nil), d.Triggers...),
		ability.Trigger{Hook: ability.HookExhausted, Effect: ability.Effect{Steps: []ability.Step{{Op: ability.OpKill, Subject: ability.SubjectSelf}}}})
	c.Ability = &d
	return &c
}

// drfHandIs reports whether player 0's hand is exactly want, in order.
func drfHandIs(s *scene, want ...*Instance) bool {
	h := s.g.Players[0].Hand
	if len(h) != len(want) {
		return false
	}
	for i := range h {
		if h[i] != want[i] {
			return false
		}
	}
	return true
}

func TestCard_DolusRefiner_EntersMainTheaterAsZeroNine(t *testing.T) {
	// Covers: FR-TRN-003, FR-SET-004
	s := newScene(t)
	c := s.card("Dolus Refiner")
	inst := s.give(0, c)
	s.setPool(0, 4, 0)
	s.do(Action{Kind: ActDeploy, Player: 0, Card: inst.ID, Zone: ZoneMainTheater})
	a := assetOf(s.g, c)
	if a == nil || a.Zone != ZoneMainTheater || a.Controller != 0 {
		t.Fatalf("the Refiner is on the battlefield in player 0's Main Theater after deploying (FR-TRN-003)")
	}
	if p := a.Printed(); p.Power != 0 || p.Health != 9 {
		t.Errorf("the Refiner is a 0/9: got %d/%d", p.Power, p.Health)
	}
	if a.Exhausted {
		t.Errorf("a newly deployed structure enters Readied (FR-SET-004)")
	}
	if p := s.g.Players[0]; p.Req != 0 || p.Tac != 0 {
		t.Errorf("the Refiner costs R4: pool R%d T%d, want R0 T0 (FR-TRN-003)", p.Req, p.Tac)
	}

	s2 := newScene(t)
	c2 := s2.card("Dolus Refiner")
	inst2 := s2.give(0, c2)
	s2.setPool(0, 4, 0)
	s2.reject(Action{Kind: ActDeploy, Player: 0, Card: inst2.ID, Zone: ZoneReserves})
	if !inHand(s2.g.Players[0], inst2) || assetOf(s2.g, c2) != nil {
		t.Errorf("deploying the Refiner to Reserves is rejected and the card stays in hand (FR-TRN-003)")
	}

	s3 := newScene(t)
	c3 := s3.card("Dolus Refiner")
	inst3 := s3.give(0, c3)
	s3.setPool(0, 3, 0)
	s3.reject(Action{Kind: ActDeploy, Player: 0, Card: inst3.ID, Zone: ZoneMainTheater})
	if !inHand(s3.g.Players[0], inst3) || assetOf(s3.g, c3) != nil {
		t.Errorf("with R3 the Refiner (cost R4) cannot be deployed (FR-TRN-003)")
	}
}

func TestCard_DolusRefiner_DrawsTwoThenDiscardsTheTwoPicked(t *testing.T) {
	// Covers: FR-ACT-001, FR-MOD-001
	s, r, hand, deck := drfScene(t)
	if !offersActivate(s.g, r, 0) {
		t.Fatalf("the Refiner's ability is offered with pool R0 T0 (FR-ACT-001)")
	}
	s.script(0, Answer{Picks: []int{0, 1}})
	s.do(activate(r, 0))
	p := s.g.Players[0]
	asks := askedKind(s, 0, DecideHand)
	if len(asks) != 1 || len(s.asked(0)) != 1 {
		t.Fatalf("player 0 is asked exactly one Hand decision for the discard: %d hand, %d total (FR-MOD-001)", len(asks), len(s.asked(0)))
	}
	d := asks[0]
	if d.Min != 2 || d.Max != 2 {
		t.Errorf("the discard is mandatory, exactly 2 cards: Min %d Max %d (FR-MOD-001)", d.Min, d.Max)
	}
	want := []*Instance{hand[0], hand[1], deck[0], deck[1]}
	if len(d.Options) != len(want) {
		t.Fatalf("the discard is chosen from H1, H2 and the two drawn cards: got %d options (FR-MOD-001)", len(d.Options))
	}
	for i, w := range want {
		if d.Options[i].Ref != w.ID {
			t.Errorf("option %d should be instance %d, got ref %d", i, w.ID, d.Options[i].Ref)
		}
	}
	got := handDiscards(s)
	if len(got) != 2 || got[0] != hand[0].ID || got[1] != hand[1].ID {
		t.Errorf("H1 then H2 are discarded from hand: %v (FR-ACT-001)", got)
	}
	if !inDiscardPile(p, hand[0]) || !inDiscardPile(p, hand[1]) {
		t.Errorf("the picked cards are in their owner's discard pile")
	}
	if !drfHandIs(s, deck[0], deck[1]) {
		t.Errorf("the hand is exactly the two drawn cards D1 and D2, got %d cards", len(p.Hand))
	}
	if len(p.Deck) == 0 || p.Deck[0] != deck[2] {
		t.Errorf("exactly the top 2 cards were drawn: D3 stays on top of the deck (FR-ACT-001)")
	}
	if p.Req != 0 || p.Tac != 0 {
		t.Errorf("the ability costs no resources: pool R%d T%d, want R0 T0 (FR-ACT-001)", p.Req, p.Tac)
	}
	if !s.alive(r) || !r.Exhausted {
		t.Errorf("the Refiner is Exhausted and on the battlefield: alive %v exhausted %v (FR-ACT-001)", s.alive(r), r.Exhausted)
	}
}

func TestCard_DolusRefiner_MayDiscardTheCardsJustDrawn(t *testing.T) {
	// Covers: FR-MOD-001
	s, r, hand, deck := drfScene(t)
	s.script(0, Answer{Picks: []int{2, 3}})
	s.do(activate(r, 0))
	p := s.g.Players[0]
	if !inDiscardPile(p, deck[0]) || !inDiscardPile(p, deck[1]) {
		t.Errorf("the cards just drawn are legal discards: D1 and D2 are discarded (FR-MOD-001)")
	}
	if !drfHandIs(s, hand[0], hand[1]) {
		t.Errorf("the hand is exactly H1 and H2, got %d cards", len(p.Hand))
	}
}

func TestCard_DolusRefiner_TimeoutDiscardsTheTwoCardsJustDrawn(t *testing.T) {
	// Covers: FR-MOD-001, SVY-R16, SVY-R50
	s, r, hand, deck := drfScene(t)
	s.script(0, Answer{Timeout: true})
	s.do(activate(r, 0))
	p := s.g.Players[0]
	if !inDiscardPile(p, deck[0]) || !inDiscardPile(p, deck[1]) {
		t.Errorf("a timeout discards the two cards most recently added to the hand, D1 and D2 (FR-MOD-001 §6, SVY-R50)")
	}
	if !drfHandIs(s, hand[0], hand[1]) {
		t.Errorf("the hand is exactly H1 and H2 after the timeout, got %d cards", len(p.Hand))
	}
}

func TestCard_DolusRefiner_EmptyHandDiscardsBothDrawnWithoutAsking(t *testing.T) {
	// Covers: FR-MOD-001, SVY-R16
	s, r, _, deck := drfScene(t)
	s.g.Players[0].Hand = nil
	s.script(0)
	s.do(activate(r, 0))
	p := s.g.Players[0]
	if n := len(askedKind(s, 0, DecideHand)); n != 0 {
		t.Errorf("exactly two cards in hand are discarded without asking: %d Hand decisions (FR-MOD-001 §1, SVY-R16)", n)
	}
	if !inDiscardPile(p, deck[0]) || !inDiscardPile(p, deck[1]) {
		t.Errorf("D1 and D2 are drawn and discarded (FR-MOD-001)")
	}
	if len(p.Hand) != 0 {
		t.Errorf("the hand is empty, got %d cards", len(p.Hand))
	}
}

func TestCard_DolusRefiner_ShortDeckDrawsWhatIsThereThenLoses(t *testing.T) {
	// Covers: FR-WIN-001, FR-LAY-002, KIM-S04
	s, r, _, _ := drfScene(t)
	p := s.g.Players[0]
	p.Hand = nil
	p.Deck = nil
	deck := s.onTop(0, vanilla("D1")...)
	s.script(0)
	s.script(1)
	s.do(activate(r, 0))
	if n := len(askedKind(s, 0, DecideHand)); n != 0 {
		t.Errorf("the one drawn card is the whole hand, fewer than 2, so it is discarded without asking: %d decisions (FR-MOD-001 §1, SVY-R16)", n)
	}
	if !inDiscardPile(p, deck[0]) || len(p.Hand) != 0 || len(p.Deck) != 0 {
		t.Errorf("the one card in the deck is drawn and discarded, leaving hand and deck empty (FR-WIN-001)")
	}
	if o := s.g.Outcome; !o.Over || o.Draw || o.Winner != 1 {
		t.Errorf("player 0 drew from an empty deck and loses at the death sweep after resolving, outcome %+v (FR-WIN-001, KIM-S04)", o)
	}
}

func TestCard_DolusRefiner_OnlyTheControllersHandAndDeckAreUsed(t *testing.T) {
	// Covers: FR-ACT-001
	s, r, _, _ := drfScene(t)
	foe := s.g.Players[1]
	hand, deck, disc := len(foe.Hand), len(foe.Deck), len(foe.Discard)
	s.script(0, Answer{Picks: []int{0, 1}})
	s.script(1)
	s.do(activate(r, 0))
	if len(foe.Hand) != hand || len(foe.Deck) != deck || len(foe.Discard) != disc {
		t.Errorf("the opponent's hand, deck and discard are unchanged: hand %d->%d deck %d->%d discard %d->%d (FR-ACT-001)",
			hand, len(foe.Hand), deck, len(foe.Deck), disc, len(foe.Discard))
	}
	if n := len(askedKind(s, 1, DecideHand)); n != 0 {
		t.Errorf("the opponent is asked no Hand decision: asked %d times", n)
	}
}

func TestCard_DolusRefiner_DiscardingEerieRemnantFiresItsTrigger(t *testing.T) {
	// Covers: FR-ZON-004, FR-TRN-003a
	s, r, _, _ := drfScene(t)
	p := s.g.Players[0]
	p.Hand = nil
	rem := s.give(0, s.card("Eerie Remnant"))
	h1 := s.give(0, vanilla("H1")[0])
	s.byKind(0, func(d Decision) Answer {
		switch d.Kind {
		case DecideHand:
			var picks []int
			for i, o := range d.Options {
				if o.Ref == rem.ID || o.Ref == h1.ID {
					picks = append(picks, i)
				}
			}
			return Answer{Picks: picks}
		case DecideOptional:
			return yes
		}
		return Answer{Timeout: true}
	})
	s.script(1)
	s.do(activate(r, 0))
	if got := askedKind(s, 0, DecideOptional); len(got) != 1 {
		t.Fatalf("player 0 is asked one optional question for the Remnant's trigger, got %d (FR-ZON-004)", len(got))
	}
	if !inDiscardPile(p, h1) {
		t.Errorf("H1 is discarded with the Remnant")
	}
	if !remnantExhaustedInReserves(s, 0, rem) {
		t.Errorf("the accepted trigger deploys the Remnant under player 0, Exhausted (FR-ZON-004, FR-TRN-003a)")
	}
	if len(s.g.Stack) != 0 {
		t.Errorf("the stack is empty once everything resolves: %d items", len(s.g.Stack))
	}
}

func TestCard_DolusRefiner_NotOfferedWhileExhausted(t *testing.T) {
	// Covers: FR-ACT-002
	s, r, hand, deck := drfScene(t)
	r.Exhausted = true
	if offersActivate(s.g, r, 0) {
		t.Errorf("an Exhausted Refiner is not offered (FR-ACT-002)")
	}
	s.reject(activate(r, 0))
	p := s.g.Players[0]
	if !drfHandIs(s, hand[0], hand[1]) || len(p.Deck) == 0 || p.Deck[0] != deck[0] || len(handDiscards(s)) != 0 {
		t.Errorf("a rejected activation draws nothing and discards nothing (FR-ACT-002)")
	}
}

func TestCard_DolusRefiner_NotOfferedWhileTheStackIsNotEmpty(t *testing.T) {
	// Covers: FR-ACT-001
	s, r, hand, deck := drfScene(t)
	s.g.Stack = append(s.g.Stack, &StackItem{Kind: StackScratch, Controller: 0})
	s.reject(activate(r, 0))
	if offersActivate(s.g, r, 0) {
		t.Errorf("no Activate is offered with an item on the stack (FR-ACT-001)")
	}
	p := s.g.Players[0]
	if r.Exhausted || !drfHandIs(s, hand[0], hand[1]) || p.Deck[0] != deck[0] {
		t.Errorf("a rejected Activate spends and draws nothing (FR-ACT-001)")
	}
}

func TestCard_DolusRefiner_MayActivateTheTurnItDeploys(t *testing.T) {
	// Covers: FR-SET-004
	s := newScene(t)
	p := s.g.Players[0]
	hand := drfHand(s)
	deck := s.onTop(0, vanilla("D1", "D2", "D3")...)
	c := s.card("Dolus Refiner")
	inst := s.give(0, c)
	s.setPool(0, 4, 0)
	s.do(Action{Kind: ActDeploy, Player: 0, Card: inst.ID, Zone: ZoneMainTheater})
	a := assetOf(s.g, c)
	if a == nil {
		t.Fatalf("the Refiner is deployed")
	}
	if !offersActivate(s.g, a, 0) {
		t.Fatalf("the ability is offered the turn the Refiner deploys (FR-SET-004)")
	}
	// The hand is H1, H2, then the three-card-deep draw puts D1, D2 after them.
	s.script(0, Answer{Picks: []int{0, 1}})
	s.do(activate(a, 0))
	if !drfHandIs(s, deck[0], deck[1]) || !inDiscardPile(p, hand[0]) || !inDiscardPile(p, hand[1]) {
		t.Errorf("activating draws 2 cards and discards 2 (FR-SET-004)")
	}
}

func TestCard_DolusRefiner_SourceGoneStillDrawsAndDiscards(t *testing.T) {
	// Covers: FR-ZON-004
	s := newScene(t)
	p := s.g.Players[0]
	hand := drfHand(s)
	deck := s.onTop(0, vanilla("D1", "D2", "D3")...)
	r := s.put(0, ZoneMainTheater, drfTapCard(s))
	s.setPool(0, 0, 0)
	// When the discard is asked, the ability's steps are already under way,
	// so the Refiner must already be off the battlefield and in the pile.
	goneWhenAsked, pileWhenAsked := false, false
	s.byKind(0, func(d Decision) Answer {
		if d.Kind == DecideHand {
			goneWhenAsked = !s.alive(r)
			pileWhenAsked = inDiscardPile(p, r.Inst)
		}
		return Answer{Picks: []int{0, 1}}
	})
	s.do(activate(r, 0))
	if s.alive(r) {
		t.Fatalf("setup: the Refiner should have been killed before the ability resolved")
	}
	if !goneWhenAsked || !pileWhenAsked {
		t.Errorf("the Refiner is off the battlefield and in the discard pile before the discard is asked (FR-ZON-004): gone %v, in pile %v", goneWhenAsked, pileWhenAsked)
	}
	if !inDiscardPile(p, r.Inst) {
		t.Errorf("the Refiner's card is in player 0's discard pile")
	}
	for _, d := range askedKind(s, 0, DecideHand) {
		for _, o := range d.Options {
			if o.Ref == r.Inst.ID {
				t.Errorf("the Refiner's card is never offered in the discard decision")
			}
		}
	}
	if !inDiscardPile(p, hand[0]) || !inDiscardPile(p, hand[1]) || !drfHandIs(s, deck[0], deck[1]) {
		t.Errorf("the ability still draws D1 and D2 and discards the two picked (FR-ZON-004)")
	}
}
