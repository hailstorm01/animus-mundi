package engine

import (
	"reflect"
	"testing"
)

// Tests for T-0165: Sinuous Growth, "T1, kill this unit: Return chosen Fungus or
// Plant card from your discard pile to your hand." (FR-ACT-001, FR-ACT-002,
// FR-TGT-001, SVY-R2, SVY-R36). The real card is loaded from data/cards. Player 0
// is the active player and activates.

// sgScene is a scene with Sinuous Growth on player 0's Main Theater and 1 Tactics
// (and no Requisition) in player 0's pool.
func sgScene(t *testing.T) (*scene, *Asset) {
	t.Helper()
	s := newScene(t)
	a := s.put(0, ZoneMainTheater, s.card("Sinuous Growth"))
	s.setPool(0, 0, 1)
	return s, a
}

func TestCard_SinuousGrowth_ReturnsTheChosenCard(t *testing.T) {
	s, a := sgScene(t)
	s.setPool(0, 3, 2)
	older := s.card("Animoceps Shambler")
	newer := s.card("Spoil Forest Colossus")
	piles := s.inDiscard(0, older, newer)
	s.script(0, Answer{Picks: []int{0}})
	s.do(activate(a, 0))
	p := s.g.Players[0]
	if !inHand(p, piles[0]) || inDiscardPile(p, piles[0]) {
		t.Errorf("the chosen card is returned to hand (FR-ACT-001)")
	}
	if !inDiscardPile(p, piles[1]) || inHand(p, piles[1]) {
		t.Errorf("the card not chosen stays in the discard pile")
	}
	if !inDiscardPile(p, a.Inst) || s.alive(a) {
		t.Errorf("Sinuous Growth is killed and sits in its owner's discard pile (FR-ACT-001)")
	}
	if p.Req != 3 || p.Tac != 1 {
		t.Errorf("the cost is exactly 1 Tactics and no Requisition: pool R%d T%d, want R3 T1 (FR-ACT-001)", p.Req, p.Tac)
	}
}

func TestCard_SinuousGrowth_ListsFungusOrPlantCardsOfAnyType(t *testing.T) {
	s, a := sgScene(t)
	other := unit("Other Subtype", 1, 1, Keywords{})
	other.Subtypes = []string{"Mutant"}
	piles := s.inDiscard(0,
		s.card("Animoceps Shambler"),
		s.card("Spoil Forest Colossus"),
		s.card("Animoceps Spore Cloud"),
		s.card("Funeral Fungus"),
		other)
	s.script(0, Answer{Picks: []int{0}})
	s.do(activate(a, 0))
	asks := s.asked(0)
	if len(asks) != 1 || asks[0].Kind != DecideTarget {
		t.Fatalf("one target question is asked, got %d decisions (FR-TGT-001)", len(asks))
	}
	got := sortedCopy(refs(asks[0]))
	want := sortedCopy([]int{piles[0].ID, piles[1].ID, piles[2].ID})
	if !reflect.DeepEqual(got, want) {
		t.Errorf("options are the Fungus unit, the Plant unit and the Fungus structure only: got %v, want %v (FR-TGT-001)", got, want)
	}
}

func TestCard_SinuousGrowth_OwnCardCannotBeChosen(t *testing.T) {
	// Covers: SVY-R2, SVY-R36
	other := unit("Other Subtype", 1, 1, Keywords{})
	other.Subtypes = []string{"Mutant"}
	t.Run("empty pile", func(t *testing.T) {
		s, a := sgScene(t)
		actNotOffered(t, s, a, 0, "an empty discard pile")
	})
	t.Run("only non-matching cards", func(t *testing.T) {
		s, a := sgScene(t)
		s.inDiscard(0, other, s.card("Funeral Fungus"))
		actNotOffered(t, s, a, 0, "a pile with no Fungus or Plant card")
	})
}

func TestCard_SinuousGrowth_AnotherCopyCanReturnTheFirst(t *testing.T) {
	// Covers: SVY-R2
	s, first := sgScene(t)
	s.setPool(0, 0, 2)
	second := s.put(0, ZoneMainTheater, s.card("Sinuous Growth"))
	target := s.inDiscard(0, s.card("Animoceps Shambler"))[0]
	s.script(0, Answer{Picks: []int{0}}, Answer{Picks: []int{0}})
	s.do(activate(first, 0))
	p := s.g.Players[0]
	if !inHand(p, target) {
		t.Fatalf("the first copy returned the only other matching card (SVY-R2)")
	}
	if !offersActivate(s.g, second, 0) {
		t.Fatalf("the first copy's card is now a legal choice, so the second copy is offered (SVY-R2)")
	}
	before := len(s.asked(0))
	s.do(activate(second, 0))
	asks := s.asked(0)[before:]
	if len(asks) > 1 || (len(asks) == 1 && !reflect.DeepEqual(refs(asks[0]), []int{first.Inst.ID})) {
		t.Errorf("the second copy's only option is the first copy's card (SVY-R2)")
	}
	if !inHand(p, first.Inst) || inDiscardPile(p, first.Inst) {
		t.Errorf("the first copy's card ends in hand (SVY-R2)")
	}
	if !inDiscardPile(p, second.Inst) {
		t.Errorf("the second copy's card is in the discard pile (SVY-R2)")
	}
}

func TestCard_SinuousGrowth_OpponentsPileDoesNotCount(t *testing.T) {
	// Covers: FR-TGT-001
	s, a := sgScene(t)
	s.inDiscard(1, s.card("Animoceps Shambler"))
	actNotOffered(t, s, a, 0, "only the opponent's discard pile holds a match")
}

func TestCard_SinuousGrowth_NeedsOneTactics(t *testing.T) {
	// Covers: FR-ACT-001
	s, a := sgScene(t)
	s.setPool(0, 5, 0)
	s.inDiscard(0, s.card("Animoceps Shambler"))
	actNotOffered(t, s, a, 0, "no Tactics in the pool")
	if s.g.Players[0].Req != 5 {
		t.Errorf("Requisition cannot pay a Tactics cost: pool R%d, want R5 (FR-ACT-001)", s.g.Players[0].Req)
	}
}

func TestCard_SinuousGrowth_WorksWhileExhausted(t *testing.T) {
	// Covers: FR-ACT-002
	s, a := sgScene(t)
	a.Exhausted = true
	target := s.inDiscard(0, s.card("Animoceps Shambler"))[0]
	s.script(0, Answer{Picks: []int{0}})
	if !offersActivate(s.g, a, 0) {
		t.Fatalf("an Exhausted Sinuous Growth is offered, its cost has no exhaust (FR-ACT-002)")
	}
	s.do(activate(a, 0))
	if !inHand(s.g.Players[0], target) {
		t.Errorf("the ability resolves and returns the card (FR-ACT-002)")
	}
}
