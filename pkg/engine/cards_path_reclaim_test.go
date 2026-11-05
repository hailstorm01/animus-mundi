package engine

import (
	"reflect"
	"testing"
)

// Tests for T-0166: Boneyard Scavenger, "T1, exhaust, kill this unit: Return chosen
// maneuver card from your discard pile to your hand.", and Ruin Partisans, "T2,
// exhaust, kill this unit: Return chosen unit card from your discard pile to your
// hand." (FR-ACT-001, FR-ACT-002, FR-TGT-001, SVY-R2, SVY-R36). The real cards are
// loaded from data/cards. Player 0 is the active player and activates.

// reclaimScene is a scene with the named card on player 0's Main Theater and the
// given Tactics (and no Requisition) in player 0's pool.
func reclaimScene(t *testing.T, name string, tac int) (*scene, *Asset) {
	t.Helper()
	s := newScene(t)
	a := s.put(0, ZoneMainTheater, s.card(name))
	s.setPool(0, 0, tac)
	return s, a
}

// reclaimExhaustedNotOffered checks an Exhausted activator is not offered and that a
// rejected Activate pays and moves nothing.
func reclaimExhaustedNotOffered(t *testing.T, s *scene, a *Asset) {
	t.Helper()
	p := s.g.Players[0]
	req, tac, pile, hand := p.Req, p.Tac, len(p.Discard), len(p.Hand)
	if offersActivate(s.g, a, 0) {
		t.Errorf("an Exhausted unit is not offered, its cost includes exhausting it (FR-ACT-002)")
	}
	s.reject(activate(a, 0))
	if !s.alive(a) || !a.Exhausted || p.Req != req || p.Tac != tac || len(p.Discard) != pile || len(p.Hand) != hand {
		t.Errorf("a rejected Activate pays nothing and moves nothing: alive %v exhausted %v R%d T%d pile %d hand %d (FR-ACT-002)",
			s.alive(a), a.Exhausted, p.Req, p.Tac, len(p.Discard), len(p.Hand))
	}
}

func TestCard_BoneyardScavenger_ReturnsTheChosenManeuver(t *testing.T) {
	s, a := reclaimScene(t, "Boneyard Scavenger", 1)
	s.setPool(0, 3, 2)
	piles := s.inDiscard(0, s.card("Steel Rain"), s.card("Fortify"))
	s.script(0, Answer{Picks: []int{0}})
	s.do(activate(a, 0))
	p := s.g.Players[0]
	if !inHand(p, piles[0]) || inDiscardPile(p, piles[0]) {
		t.Errorf("the chosen Maneuver is returned to hand (FR-ACT-001)")
	}
	if !inDiscardPile(p, piles[1]) || inHand(p, piles[1]) {
		t.Errorf("the Maneuver not chosen stays in the discard pile")
	}
	if !inDiscardPile(p, a.Inst) || s.alive(a) {
		t.Errorf("Boneyard Scavenger is killed and sits in its owner's discard pile (FR-ACT-001)")
	}
	if p.Req != 3 || p.Tac != 1 {
		t.Errorf("the cost is exactly 1 Tactics and no Requisition: pool R%d T%d, want R3 T1 (FR-ACT-001)", p.Req, p.Tac)
	}
}

func TestCard_BoneyardScavenger_ListsOnlyManeuverCards(t *testing.T) {
	// Covers: FR-TGT-001
	s, a := reclaimScene(t, "Boneyard Scavenger", 1)
	piles := s.inDiscard(0,
		s.card("Steel Rain"),
		s.card("Fortify"),
		s.card("Saboteur"),
		s.card("Hammerhand Bivouac"))
	s.script(0, Answer{Picks: []int{0}})
	s.do(activate(a, 0))
	asks := s.asked(0)
	if len(asks) != 1 || asks[0].Kind != DecideTarget {
		t.Fatalf("one target question is asked, got %d decisions (FR-TGT-001)", len(asks))
	}
	got := sortedCopy(refs(asks[0]))
	want := sortedCopy([]int{piles[0].ID, piles[1].ID})
	if !reflect.DeepEqual(got, want) {
		t.Errorf("options are the two Maneuvers only: got %v, want %v (FR-TGT-001)", got, want)
	}
}

func TestCard_BoneyardScavenger_NeedsAManeuverInThePile(t *testing.T) {
	// Covers: SVY-R36
	t.Run("empty pile", func(t *testing.T) {
		s, a := reclaimScene(t, "Boneyard Scavenger", 1)
		actNotOffered(t, s, a, 0, "an empty discard pile")
	})
	t.Run("only a unit and a structure", func(t *testing.T) {
		s, a := reclaimScene(t, "Boneyard Scavenger", 1)
		s.inDiscard(0, s.card("Saboteur"), s.card("Hammerhand Bivouac"))
		actNotOffered(t, s, a, 0, "a pile with no Maneuver")
	})
}

func TestCard_BoneyardScavenger_OpponentsPileDoesNotCount(t *testing.T) {
	// Covers: FR-TGT-001
	s, a := reclaimScene(t, "Boneyard Scavenger", 1)
	s.inDiscard(1, s.card("Steel Rain"))
	actNotOffered(t, s, a, 0, "only the opposing discard pile holds a Maneuver")
}

func TestCard_BoneyardScavenger_NeedsOneTactics(t *testing.T) {
	// Covers: FR-ACT-001
	s, a := reclaimScene(t, "Boneyard Scavenger", 0)
	s.setPool(0, 5, 0)
	s.inDiscard(0, s.card("Steel Rain"))
	actNotOffered(t, s, a, 0, "no Tactics in the pool")
	if s.g.Players[0].Req != 5 {
		t.Errorf("Requisition cannot pay a Tactics cost: pool R%d, want R5 (FR-ACT-001)", s.g.Players[0].Req)
	}
}

func TestCard_BoneyardScavenger_NeedsToBeReadied(t *testing.T) {
	// Covers: FR-ACT-002
	s, a := reclaimScene(t, "Boneyard Scavenger", 1)
	a.Exhausted = true
	s.inDiscard(0, s.card("Steel Rain"))
	reclaimExhaustedNotOffered(t, s, a)
}

func TestCard_BoneyardScavenger_WorksFromReserves(t *testing.T) {
	// Covers: FR-ACT-001
	s := newScene(t)
	a := s.put(0, ZoneReserves, s.card("Boneyard Scavenger"))
	s.setPool(0, 0, 1)
	target := s.inDiscard(0, s.card("Steel Rain"))[0]
	s.script(0, Answer{Picks: []int{0}})
	if !offersActivate(s.g, a, 0) {
		t.Fatalf("a Readied Boneyard Scavenger in Reserves is offered (FR-ACT-001)")
	}
	s.do(activate(a, 0))
	p := s.g.Players[0]
	if !inHand(p, target) {
		t.Errorf("the ability resolves from Reserves and returns the Maneuver (FR-ACT-001)")
	}
	if !inDiscardPile(p, a.Inst) {
		t.Errorf("Boneyard Scavenger ends in its owner's discard pile (FR-ACT-001)")
	}
}

func TestCard_RuinPartisans_ReturnsTheChosenUnit(t *testing.T) {
	s, a := reclaimScene(t, "Ruin Partisans", 2)
	s.setPool(0, 3, 4)
	piles := s.inDiscard(0, s.card("Saboteur"), s.card("Devilhorn Captain"))
	s.script(0, Answer{Picks: []int{0}})
	s.do(activate(a, 0))
	p := s.g.Players[0]
	if !inHand(p, piles[0]) || inDiscardPile(p, piles[0]) {
		t.Errorf("the chosen unit card is returned to hand (FR-ACT-001)")
	}
	if !inDiscardPile(p, piles[1]) || inHand(p, piles[1]) {
		t.Errorf("the unit card not chosen stays in the discard pile")
	}
	if !inDiscardPile(p, a.Inst) || s.alive(a) {
		t.Errorf("Ruin Partisans is killed and sits in its owner's discard pile (FR-ACT-001)")
	}
	if p.Req != 3 || p.Tac != 2 {
		t.Errorf("the cost is exactly 2 Tactics and no Requisition: pool R%d T%d, want R3 T2 (FR-ACT-001)", p.Req, p.Tac)
	}
}

func TestCard_RuinPartisans_ListsOnlyUnitCards(t *testing.T) {
	// Covers: FR-TGT-001
	s, a := reclaimScene(t, "Ruin Partisans", 2)
	piles := s.inDiscard(0,
		s.card("Saboteur"),
		s.card("Devilhorn Captain"),
		s.card("Steel Rain"),
		s.card("Hammerhand Bivouac"))
	s.script(0, Answer{Picks: []int{0}})
	s.do(activate(a, 0))
	asks := s.asked(0)
	if len(asks) != 1 || asks[0].Kind != DecideTarget {
		t.Fatalf("one target question is asked, got %d decisions (FR-TGT-001)", len(asks))
	}
	got := sortedCopy(refs(asks[0]))
	want := sortedCopy([]int{piles[0].ID, piles[1].ID})
	if !reflect.DeepEqual(got, want) {
		t.Errorf("options are the two unit cards only: got %v, want %v (FR-TGT-001)", got, want)
	}
}

func TestCard_RuinPartisans_OwnCardCannotBeChosen(t *testing.T) {
	// Covers: SVY-R2, SVY-R36
	t.Run("empty pile", func(t *testing.T) {
		s, a := reclaimScene(t, "Ruin Partisans", 2)
		actNotOffered(t, s, a, 0, "an empty discard pile")
	})
	t.Run("only a Maneuver and a structure", func(t *testing.T) {
		s, a := reclaimScene(t, "Ruin Partisans", 2)
		s.inDiscard(0, s.card("Steel Rain"), s.card("Hammerhand Bivouac"))
		actNotOffered(t, s, a, 0, "a pile with no unit card")
	})
}

func TestCard_RuinPartisans_AnotherCopyCanReturnTheFirst(t *testing.T) {
	// Covers: SVY-R2
	s, first := reclaimScene(t, "Ruin Partisans", 4)
	second := s.put(0, ZoneMainTheater, s.card("Ruin Partisans"))
	target := s.inDiscard(0, s.card("Saboteur"))[0]
	s.script(0, Answer{Picks: []int{0}}, Answer{Picks: []int{0}})
	s.do(activate(first, 0))
	p := s.g.Players[0]
	if !inHand(p, target) {
		t.Fatalf("the first copy returned the only other unit card (SVY-R2)")
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

func TestCard_RuinPartisans_NeedsTwoTactics(t *testing.T) {
	// Covers: FR-ACT-001
	s, a := reclaimScene(t, "Ruin Partisans", 0)
	s.setPool(0, 5, 1)
	s.inDiscard(0, s.card("Saboteur"))
	actNotOffered(t, s, a, 0, "only 1 Tactics in the pool")
	if s.g.Players[0].Req != 5 || s.g.Players[0].Tac != 1 {
		t.Errorf("nothing is paid: pool R%d T%d, want R5 T1 (FR-ACT-001)", s.g.Players[0].Req, s.g.Players[0].Tac)
	}
}

func TestCard_RuinPartisans_NeedsToBeReadied(t *testing.T) {
	// Covers: FR-ACT-002
	s, a := reclaimScene(t, "Ruin Partisans", 2)
	a.Exhausted = true
	s.inDiscard(0, s.card("Saboteur"))
	reclaimExhaustedNotOffered(t, s, a)
}
