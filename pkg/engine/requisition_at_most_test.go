package engine

import (
	"slices"
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0221: the Requisition cost as the second stat of the
// FR-TGT-001 "Dynamic Stats" criterion ("with Requisition cost N or less").
// A card off the battlefield is read by its printed cost (FR-ZON-004), an
// asset by the cost printed on what it currently is (no layer changes a
// Requisition cost). Cards are synthetic. Every scene deploys the unit
// "Source" (Requisition cost 0, which would satisfy "cost 3 or less") from
// player 0's hand (see iydDeploy), so battlefield filters set Other to keep it
// out of the offers.

// raMax3 is "with Requisition cost 3 or less".
func raMax3(f ability.Filter) ability.Filter {
	f.Stat = ability.StatRequisition
	f.AtMost = 3
	return f
}

// raUnit is a synthetic unit with the given Requisition cost and power.
func raUnit(name string, cost, power int) *Card {
	c := unit(name, power, 5, Keywords{})
	c.ReqCost = cost
	return c
}

func TestFR_TGT_001_RequisitionAtMostOffersPileCardsByPrintedCost(t *testing.T) {
	// Covers: FR-ZON-004, FR-CRD-101
	s := newScene(t)
	// Power is varied against cost: the cost-4 card has power 1 and the
	// cost-2 card power 5, so reading power instead of cost fails.
	mine := s.inDiscard(0, raUnit("C2", 2, 5), raUnit("C3", 3, 3),
		raUnit("C4", 4, 1), structure("Wall", 0, 3, Keywords{}))
	s.inDiscard(1, raUnit("Theirs1", 1, 1))
	pile := ability.Target{Filter: raMax3(ppAllyUnitCards)}
	s.script(0, Answer{Picks: []int{0}})
	s.iydDeploy(iydTrigger(false, []ability.Target{pile}, ptStep(ability.OpReturnToHand, 0)))
	d := paTargetAsk(t, s)
	want := ppInstIDs(mine[0], mine[1])
	if got := ppRefs(d); !slices.Equal(paSorted(got), paSorted(want)) {
		t.Errorf("'Requisition cost 3 or less' offers exactly your unit cards of printed cost 2 and 3 (inclusive bound; kind and allegiance still apply; cost, not power): got %v, want %v (FR-TGT-001, FR-ZON-004, FR-CRD-101)", got, want)
	}
}

func TestFR_TGT_001_RequisitionAtMostPilePickReadsPrintedCost(t *testing.T) {
	// Covers: SVY-R3, FR-ZON-004
	s := newScene(t)
	mine := s.inDiscard(0, raUnit("C2", 2, 5), raUnit("C3", 3, 3),
		raUnit("C4", 4, 1), structure("Wall", 0, 3, Keywords{}))
	s.inDiscard(1, raUnit("Theirs1", 1, 1))
	s.ppScript(Answer{Picks: []int{0}})
	s.iydDeploy(iydTrigger(false, nil, ppPick(ability.OpReturnToHand, raMax3(ppAllyUnitCards), 0)))
	asks := ppPicks(s)
	if len(asks) != 1 {
		t.Fatalf("two cards match, so one pick question: got %d", len(asks))
	}
	want := ppInstIDs(mine[0], mine[1])
	if got := ppRefs(asks[0]); !slices.Equal(paSorted(got), paSorted(want)) {
		t.Errorf("the pick offers exactly your unit cards of printed Requisition cost 3 or less: got %v, want %v (SVY-R3, FR-ZON-004)", got, want)
	}
}

func TestFR_TGT_001_RequisitionAtMostReadsAnAssetsPrintedCost(t *testing.T) {
	// Covers: FR-TGT-001, FR-CRD-101, FR-LAY-001
	s := newScene(t)
	// Power is varied against cost: the cost-4 unit has power 1, the cost-2 unit power 5.
	c2 := s.put(1, ZoneMainTheater, raUnit("C2", 2, 5))
	c3 := s.put(1, ZoneMainTheater, raUnit("C3", 3, 3))
	c3.Modifiers = []Modifier{{Layer: 5, Power: 2}} // power 5; its cost stays 3
	s.put(1, ZoneMainTheater, raUnit("C4", 4, 1))
	slot := ability.Target{Filter: raMax3(ability.Filter{Kinds: ability.KindUnit, Other: true})}
	s.script(0, Answer{Picks: []int{0}})
	s.iydDeploy(iydTrigger(false, []ability.Target{slot}, armorStep(1, 0)))
	d := paTargetAsk(t, s)
	want := ids(c2, c3)
	if got := refs(d); !slices.Equal(paSorted(got), paSorted(want)) {
		t.Errorf("'unit with Requisition cost 3 or less' offers the cost-2 and cost-3 units (a power modifier does not change a cost), not the cost-4 unit; the Source is kept out by Other: got %v, want %v (FR-TGT-001, FR-CRD-101)", got, want)
	}
}
