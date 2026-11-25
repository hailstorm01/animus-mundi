package engine

import (
	"animus-mundi/pkg/engine/ability"
	"reflect"
	"testing"
)

// Card tests for Leyline Regrowth (T-0301), written from the CSV text:
// Concordia Maneuver, R0 T2, "Return up to one chosen unit card from your
// discard pile to your hand. Draw a card." Player 0 enacts; player 1 is the
// opponent.

// lrChoosePile installs a chooser for seat 0 that picks the option whose Ref is
// the given pile instance, and records the decisions asked.
func lrChoosePile(s *scene, target *Instance) {
	if s.asks == nil {
		s.asks = map[PlayerID][]Decision{}
	}
	s.asks[0] = nil
	s.g.choosers[0] = func(_ *Game, d Decision) Answer {
		s.asks[0] = append(s.asks[0], d)
		for i, o := range d.Options {
			if o.Ref == target.ID {
				return Answer{Picks: []int{i}}
			}
		}
		s.t.Errorf("pile card %d is not offered in %q: %v", target.ID, d.Prompt, d.Options)
		return Answer{Timeout: true}
	}
}

func TestCard_LeylineRegrowth_ReturnsChosenUnitCardThenDraws(t *testing.T) {
	// Covers: FR-MNV-001, FR-TGT-001
	s := newScene(t)
	piles := s.inDiscard(0, rrUnit("PileA", 1), rrUnit("PileB", 2))
	chosen, other := piles[1], piles[0]
	top := s.onTop(0, unit("Top", 1, 1, ability.Keywords{}))[0]
	inst := enactFor(s, "Leyline Regrowth")
	lrChoosePile(s, chosen)
	s.do(enactAction(inst))
	p := s.g.Players[0]
	if inDiscardPile(p, chosen) || !inHand(p, chosen) {
		t.Errorf("the chosen unit card leaves the discard pile for the hand (FR-MNV-001, FR-TGT-001)")
	}
	if !inDiscardPile(p, other) || inHand(p, other) {
		t.Errorf("the unchosen unit card stays in the discard pile (FR-TGT-001)")
	}
	if !inHand(p, top) {
		t.Errorf("the top card of the deck is drawn (FR-MNV-001)")
	}
	if n := len(p.Hand); n < 2 || p.Hand[n-2] != chosen || p.Hand[n-1] != top {
		t.Errorf("the card is returned before the draw: hand %v, want the returned card then the drawn card last (FR-MNV-001)", p.Hand)
	}
	if lastDiscard(p) != inst {
		t.Errorf("the Maneuver ends in its owner's discard pile (FR-MNV-001)")
	}
}

func TestCard_LeylineRegrowth_ChoosingNoneStillDraws(t *testing.T) {
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
			pile := s.inDiscard(0, rrUnit("Pile", 1))[0]
			top := s.onTop(0, unit("Top", 1, 1, ability.Keywords{}))[0]
			s.script(0, c.ans)
			s.do(enactAction(enactFor(s, "Leyline Regrowth")))
			p := s.g.Players[0]
			if !inDiscardPile(p, pile) || inHand(p, pile) {
				t.Errorf("choosing no card returns nothing (FR-TGT-002, SVY-R16)")
			}
			if !inHand(p, top) {
				t.Errorf("the card is still drawn when none is chosen (FR-MNV-001)")
			}
		})
	}
}

func TestCard_LeylineRegrowth_OnlyOwnUnitCardsOffered(t *testing.T) {
	// Covers: FR-TGT-001, FR-TGT-002
	s := newScene(t)
	units := s.inDiscard(0, rrUnit("U1", 1), rrUnit("U2", 2))
	s.inDiscard(0, structure("Wall", 0, 3, ability.Keywords{}), maneuver("Spell", 0, 0, steps()))
	s.inDiscard(1, rrUnit("Theirs", 1))
	s.put(0, ZoneMainTheater, unit("MyUnit", 1, 3, ability.Keywords{}))
	s.put(1, ZoneMainTheater, unit("FoeUnit", 1, 3, ability.Keywords{}))
	s.script(0, Answer{Picks: []int{}})
	s.do(enactAction(enactFor(s, "Leyline Regrowth")))
	asks := s.asked(0)
	if len(asks) != 1 || asks[0].Kind != DecideTarget {
		t.Fatalf("one target question is asked, got %d decisions (FR-TGT-001)", len(asks))
	}
	got := sortedCopy(refs(asks[0]))
	want := sortedCopy([]int{units[0].ID, units[1].ID})
	if !reflect.DeepEqual(got, want) {
		t.Errorf("options are exactly the two unit cards in player 0's pile: got %v, want %v (FR-TGT-001)", got, want)
	}
	if asks[0].Min != 0 || asks[0].Max != 1 {
		t.Errorf("the question takes at most one pick and may take none: min %d max %d (FR-TGT-002)", asks[0].Min, asks[0].Max)
	}
}

func TestCard_LeylineRegrowth_EnactableWithNoUnitCard(t *testing.T) {
	// Covers: SVY-R29, FR-TGT-002
	check := func(t *testing.T, s *scene) {
		t.Helper()
		top := s.onTop(0, unit("Top", 1, 1, ability.Keywords{}))[0]
		inst := enactFor(s, "Leyline Regrowth")
		if !offersEnact(s.g, inst) {
			t.Fatalf("an \"up to\" slot never stops an Enact, even with no unit card to choose (SVY-R29)")
		}
		s.script(0)
		s.do(enactAction(inst))
		if n := len(s.asked(0)); n != 0 {
			t.Errorf("no legal target, so no decision is asked: asked %d times (FR-TGT-002)", n)
		}
		if !inHand(s.g.Players[0], top) {
			t.Errorf("the card is still drawn (FR-MNV-001)")
		}
	}
	t.Run("empty pile", func(t *testing.T) {
		check(t, newScene(t))
	})
	t.Run("no own unit card", func(t *testing.T) {
		s := newScene(t)
		s.inDiscard(0, structure("Wall", 0, 3, ability.Keywords{}), maneuver("Spell", 0, 0, steps()))
		s.inDiscard(1, rrUnit("Theirs", 1))
		check(t, s)
	})
}

func TestCard_LeylineRegrowth_TargetLeftThePileStillDraws(t *testing.T) {
	// Covers: FR-TGT-003, FR-MNV-001
	s := newScene(t)
	chosen := s.inDiscard(0, rrUnit("Leaves", 2))[0]
	top := s.onTop(0, unit("Top", 1, 1, ability.Keywords{}))[0]
	inst := enactFor(s, "Leyline Regrowth")
	lrChoosePile(s, chosen)
	moved := false
	s.g.obs = &recorder{onEvent: func(e Event) {
		if e.Kind == "place" && !moved {
			moved = true
			if s.g.takeCard(chosen) {
				s.g.putCard(chosen, ZoneDeck, true)
			}
		}
	}}
	s.do(enactAction(inst))
	if !moved {
		t.Fatalf("the Maneuver was placed on the stack")
	}
	p := s.g.Players[0]
	if inHand(p, chosen) {
		t.Errorf("a card that left the pile is not returned (FR-TGT-003)")
	}
	if !inHand(p, top) {
		t.Errorf("the card is still drawn when the target is gone (FR-TGT-003)")
	}
	if !inDiscardPile(p, inst) {
		t.Errorf("the Maneuver still goes to its owner's discard pile (FR-TGT-003)")
	}
}

func TestCard_LeylineRegrowth_EmptyDeckReturnsThenLoses(t *testing.T) {
	// Covers: FR-WIN-001
	s := newScene(t)
	chosen := s.inDiscard(0, rrUnit("Pile", 1))[0]
	s.g.Players[0].Deck = nil
	inst := enactFor(s, "Leyline Regrowth")
	lrChoosePile(s, chosen)
	s.do(enactAction(inst))
	if !inHand(s.g.Players[0], chosen) {
		t.Errorf("the chosen card is returned before the draw (FR-MNV-001)")
	}
	if o := s.g.Outcome; !o.Over || o.Draw || o.Winner != 1 {
		t.Errorf("drawing from an empty deck loses the match for player 0: %+v (FR-WIN-001)", o)
	}
}

func TestCard_LeylineRegrowth_PaysTwoTactics(t *testing.T) {
	// Covers: FR-MNV-001
	t.Run("pays R0 T2", func(t *testing.T) {
		s := newScene(t)
		s.script(0)
		s.do(enactAction(enactFor(s, "Leyline Regrowth")))
		if p := s.g.Players[0]; p.Tac != 3 || p.Req != 3 {
			t.Errorf("enacting costs 2 Tactics and no Requisition: pool R%d T%d, want R3 T3 (FR-MNV-001)", p.Req, p.Tac)
		}
	})
	t.Run("not offered with one Tactics", func(t *testing.T) {
		s := newScene(t)
		inst := s.give(0, s.card("Leyline Regrowth"))
		s.setPool(0, 5, 1)
		if offersEnact(s.g, inst) {
			t.Errorf("Enact is not offered with R5 T1: it costs 2 Tactics (FR-MNV-001)")
		}
	})
}
