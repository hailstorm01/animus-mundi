package engine

import (
	"reflect"
	"testing"
)

// Card tests for Inspired Salvage (T-0319), written from the CSV text:
// Soulless Maneuver, R2 T2, "Return up to 2 chosen Mechanized cards from your
// discard pile to your hand." Player 0 enacts; player 1 is the opponent.

// isEnact gives player 0 Inspired Salvage with exactly R2 T2.
func isEnact(s *scene) *Instance {
	s.t.Helper()
	inst := s.give(0, s.card("Inspired Salvage"))
	s.setPool(0, 2, 2)
	return inst
}

// isChoose installs a chooser for seat 0 that picks, in order, the options whose
// Ref is one of the given pile instances, and records the decisions asked.
func isChoose(s *scene, targets ...*Instance) {
	if s.asks == nil {
		s.asks = map[PlayerID][]Decision{}
	}
	s.asks[0] = nil
	s.g.choosers[0] = func(_ *Game, d Decision) Answer {
		s.asks[0] = append(s.asks[0], d)
		picks := []int{}
		for _, target := range targets {
			found := false
			for i, o := range d.Options {
				if o.Ref == target.ID {
					picks = append(picks, i)
					found = true
					break
				}
			}
			if !found {
				s.t.Errorf("pile card %d is not offered in %q: %v", target.ID, d.Prompt, d.Options)
				return Answer{Timeout: true}
			}
		}
		return Answer{Picks: picks}
	}
}

func TestCard_InspiredSalvage_ReturnsTwoChosenMechanizedCards(t *testing.T) {
	// Covers: FR-MNV-001, FR-TGT-001
	s := newScene(t)
	piles := s.inDiscard(0, s.card("Brass Adjutant"), s.card("Shrike Scout Suits"), s.card("Sentinel Warsuit Core"))
	a, b, rest := piles[0], piles[2], piles[1]
	inst := isEnact(s)
	isChoose(s, a, b)
	s.do(enactAction(inst))
	p := s.g.Players[0]
	for _, c := range []*Instance{a, b} {
		if inDiscardPile(p, c) || !inHand(p, c) {
			t.Errorf("a chosen Mechanized card leaves the discard pile for the hand (FR-MNV-001, FR-TGT-001)")
		}
	}
	if !inDiscardPile(p, rest) || inHand(p, rest) {
		t.Errorf("the unchosen Mechanized card stays in the discard pile (FR-TGT-001)")
	}
	if n := len(p.Hand); n != 2 {
		t.Errorf("only the two returned cards are in hand once the Maneuver is enacted: hand has %d (FR-MNV-001)", n)
	}
	if lastDiscard(p) != inst {
		t.Errorf("the Maneuver ends in its owner's discard pile (FR-MNV-001)")
	}
}

func TestCard_InspiredSalvage_ChoosingOneReturnsOne(t *testing.T) {
	// Covers: FR-TGT-002
	s := newScene(t)
	piles := s.inDiscard(0, s.card("Brass Adjutant"), s.card("Shrike Scout Suits"))
	inst := isEnact(s)
	isChoose(s, piles[1])
	s.do(enactAction(inst))
	p := s.g.Players[0]
	if !inHand(p, piles[1]) || inDiscardPile(p, piles[1]) {
		t.Errorf("the one chosen card is returned (FR-TGT-002)")
	}
	if !inDiscardPile(p, piles[0]) || inHand(p, piles[0]) {
		t.Errorf("the unchosen card stays in the discard pile (FR-TGT-002)")
	}
}

func TestCard_InspiredSalvage_ChoosingNoneReturnsNothing(t *testing.T) {
	// Covers: FR-TGT-002, FR-MOD-001, SVY-R16
	cases := []struct {
		name string
		ans  Answer
	}{
		{"empty answer", Answer{Picks: []int{}}},
		{"timed out", Answer{Timeout: true}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := newScene(t)
			piles := s.inDiscard(0, s.card("Brass Adjutant"), s.card("Shrike Scout Suits"))
			inst := isEnact(s)
			s.script(0, c.ans)
			s.do(enactAction(inst))
			p := s.g.Players[0]
			for _, pc := range piles {
				if !inDiscardPile(p, pc) || inHand(p, pc) {
					t.Errorf("choosing no card returns nothing (FR-TGT-002, SVY-R16)")
				}
			}
			if lastDiscard(p) != inst {
				t.Errorf("the Maneuver still goes to its owner's discard pile (FR-MNV-001)")
			}
		})
	}
}

func TestCard_InspiredSalvage_OnlyOwnMechanizedCardsOffered(t *testing.T) {
	// Covers: FR-TGT-001, FR-TGT-002
	s := newScene(t)
	mech := s.inDiscard(0, s.card("Brass Adjutant"), s.card("MAX-8 Autonomous Warsuit"))
	s.inDiscard(0, s.card("Conscript Rifles"), s.card("Bring It Down"))
	s.inDiscard(1, s.card("Shrike Scout Suits"))
	s.put(0, ZoneMainTheater, s.card("Sentinel Warsuit Core"))
	s.put(1, ZoneMainTheater, s.card("Sentinel Warsuit Core"))
	inst := isEnact(s)
	s.script(0, Answer{Picks: []int{}})
	s.do(enactAction(inst))
	asks := s.asked(0)
	if len(asks) != 1 || asks[0].Kind != DecideTarget {
		t.Fatalf("one target question is asked, got %d decisions (FR-TGT-001)", len(asks))
	}
	got := sortedCopy(refs(asks[0]))
	want := sortedCopy([]int{mech[0].ID, mech[1].ID})
	if !reflect.DeepEqual(got, want) {
		t.Errorf("options are exactly the two Mechanized cards in player 0's pile: got %v, want %v (FR-TGT-001)", got, want)
	}
	if asks[0].Min != 0 || asks[0].Max != 2 {
		t.Errorf("the question takes up to two picks and may take none: min %d max %d (FR-TGT-002)", asks[0].Min, asks[0].Max)
	}
}

func TestCard_InspiredSalvage_SameCardCannotBeChosenTwice(t *testing.T) {
	// Covers: SVY-R11, FR-TGT-002
	s := newScene(t)
	piles := s.inDiscard(0, s.card("Brass Adjutant"), s.card("Shrike Scout Suits"))
	inst := isEnact(s)
	isChoose(s, piles[0], piles[0])
	s.do(enactAction(inst))
	for _, d := range s.asked(0) {
		if d.Repeats {
			t.Errorf("the up-to slot may not pick one pile card twice (SVY-R11)")
		}
	}
	p := s.g.Players[0]
	once := 0
	for _, c := range p.Hand {
		if c == piles[0] {
			once++
		}
	}
	if once != 1 || inDiscardPile(p, piles[0]) {
		t.Errorf("a card named twice moves once: it is in hand %d times (SVY-R11)", once)
	}
	if !inDiscardPile(p, piles[1]) || inHand(p, piles[1]) {
		t.Errorf("the other Mechanized card does not move (SVY-R11)")
	}
	if n := len(p.Hand); n != 1 {
		t.Errorf("nothing else moves to the hand: hand has %d cards (SVY-R11)", n)
	}
}

func TestCard_InspiredSalvage_EnactableWithNoMechanizedCard(t *testing.T) {
	// Covers: SVY-R29, FR-TGT-002
	check := func(t *testing.T, s *scene) {
		t.Helper()
		inst := isEnact(s)
		if !offersEnact(s.g, inst) {
			t.Fatalf("an up-to slot never stops an Enact, even with no Mechanized card to choose (SVY-R29)")
		}
		s.script(0)
		s.do(enactAction(inst))
		if n := len(s.asked(0)); n != 0 {
			t.Errorf("no legal target, so no decision is asked: asked %d times (FR-TGT-002)", n)
		}
		p := s.g.Players[0]
		if len(p.Hand) != 0 {
			t.Errorf("nothing is returned to the hand: hand has %d cards (FR-TGT-002)", len(p.Hand))
		}
		if lastDiscard(p) != inst {
			t.Errorf("the Maneuver resolves into its owner's discard pile (FR-MNV-001)")
		}
	}
	t.Run("empty pile", func(t *testing.T) {
		check(t, newScene(t))
	})
	t.Run("no own Mechanized card", func(t *testing.T) {
		s := newScene(t)
		s.inDiscard(0, s.card("Conscript Rifles"), s.card("Bring It Down"))
		s.inDiscard(1, s.card("Brass Adjutant"))
		check(t, s)
	})
}

func TestCard_InspiredSalvage_OneTargetLeftThePileTheOtherStillReturns(t *testing.T) {
	// Covers: FR-TGT-003, FR-MNV-001
	s := newScene(t)
	piles := s.inDiscard(0, s.card("Brass Adjutant"), s.card("Shrike Scout Suits"))
	leaves, stays := piles[0], piles[1]
	inst := isEnact(s)
	isChoose(s, leaves, stays)
	moved := false
	s.g.obs = &recorder{onEvent: func(e Event) {
		if e.Kind == "place" && !moved {
			moved = true
			if s.g.takeCard(leaves) {
				s.g.putCard(leaves, ZoneDeck, true)
			}
		}
	}}
	s.do(enactAction(inst))
	if !moved {
		t.Fatalf("the Maneuver was placed on the stack")
	}
	p := s.g.Players[0]
	if inHand(p, leaves) {
		t.Errorf("a card that left the pile is not returned (FR-TGT-003)")
	}
	if n := len(p.Deck); n == 0 || p.Deck[n-1] != leaves {
		t.Errorf("the card that left the pile stays at the bottom of the deck (FR-TGT-003)")
	}
	if !inHand(p, stays) {
		t.Errorf("the other chosen card is still returned (FR-TGT-003)")
	}
	if !inDiscardPile(p, inst) {
		t.Errorf("the Maneuver still goes to its owner's discard pile (FR-TGT-003)")
	}
}

func TestCard_InspiredSalvage_PaysTwoAndTwo(t *testing.T) {
	// Covers: FR-MNV-001
	t.Run("pays R2 T2", func(t *testing.T) {
		s := newScene(t)
		inst := s.give(0, s.card("Inspired Salvage"))
		s.setPool(0, 3, 3)
		if !offersEnact(s.g, inst) {
			t.Fatalf("Enact is offered with R3 T3 (FR-MNV-001)")
		}
		s.script(0)
		s.do(enactAction(inst))
		if p := s.g.Players[0]; p.Req != 1 || p.Tac != 1 {
			t.Errorf("enacting costs 2 Requisition and 2 Tactics: pool R%d T%d, want R1 T1 (FR-MNV-001)", p.Req, p.Tac)
		}
	})
	for _, pool := range []struct {
		name     string
		req, tac int
	}{{"not offered with one Requisition", 1, 5}, {"not offered with one Tactics", 5, 1}} {
		t.Run(pool.name, func(t *testing.T) {
			s := newScene(t)
			inst := s.give(0, s.card("Inspired Salvage"))
			s.setPool(0, pool.req, pool.tac)
			if offersEnact(s.g, inst) {
				t.Errorf("Enact is not offered with R%d T%d: it costs R2 T2 (FR-MNV-001)", pool.req, pool.tac)
			}
		})
	}
}
