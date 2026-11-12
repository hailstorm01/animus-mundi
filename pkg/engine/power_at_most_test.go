package engine

import (
	"slices"
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0209: the "Dynamic Stats" criterion of FR-TGT-001 (line 375),
// first as "with power N or less". A card in a discard pile is read by its
// printed power (FR-ZON-004: off the battlefield it has only printed
// characteristics); an asset on the battlefield by its effective power
// (Layer 6). Cards are synthetic. Every scene deploys the unit "Source"
// (printed power 1) from player 0's hand (see iydDeploy). Wherever a filter
// matches battlefield assets it sets Other, so the Source (power 1, which
// would satisfy "power 3 or less") stays out of the offers.

// paMax3 is "with power 3 or less".
func paMax3(f ability.Filter) ability.Filter {
	f.Stat = ability.StatPower
	f.AtMost = 3
	return f
}

func paSorted(xs []int) []int {
	out := slices.Clone(xs)
	slices.Sort(out)
	return out
}

// paTargetAsk returns the one DecideTarget decision asked of seat 0.
func paTargetAsk(t *testing.T, s *scene) Decision {
	t.Helper()
	var out []Decision
	for _, d := range s.asked(0) {
		if d.Kind == DecideTarget {
			out = append(out, d)
		}
	}
	if len(out) != 1 {
		t.Fatalf("expected exactly one target question, got %d (%v)", len(out), kindsOf(s.asked(0)))
	}
	return out[0]
}

func TestFR_TGT_001_PowerAtMostOffersPileCardsByPrintedPower(t *testing.T) {
	// Covers: FR-ZON-004
	s := newScene(t)
	mine := s.inDiscard(0, unit("P2", 2, 1, Keywords{}), unit("P3", 3, 1, Keywords{}),
		unit("P4", 4, 1, Keywords{}), structure("Wall0", 0, 3, Keywords{}))
	s.inDiscard(1, unit("Theirs1", 1, 1, Keywords{}))
	pile := ability.Target{Filter: paMax3(ppAllyUnitCards)}
	s.script(0, Answer{Picks: []int{0}})
	s.iydDeploy(iydTrigger(false, []ability.Target{pile}, ptStep(ability.OpReturnToHand, 0)))
	d := paTargetAsk(t, s)
	want := ppInstIDs(mine[0], mine[1])
	if got := ppRefs(d); !slices.Equal(paSorted(got), paSorted(want)) {
		t.Errorf("'power 3 or less' offers exactly your unit cards of printed power 2 and 3 (inclusive bound; kind and allegiance still apply): got %v, want %v (FR-TGT-001, FR-ZON-004)", got, want)
	}
}

func TestFR_TGT_001_PowerAtMostReadsEffectivePowerOnTheBattlefield(t *testing.T) {
	// Covers: FR-TGT-001, FR-LAY-001
	s := newScene(t)
	p2 := s.put(1, ZoneMainTheater, unit("P2", 2, 5, Keywords{}))
	p3 := s.put(1, ZoneMainTheater, unit("P3", 3, 5, Keywords{}))
	s.put(1, ZoneMainTheater, unit("P4", 4, 5, Keywords{}))
	up := s.put(1, ZoneMainTheater, unit("P3plus1", 3, 5, Keywords{}))
	up.Modifiers = []Modifier{{Layer: 5, Power: 1}} // effective 4
	down := s.put(1, ZoneMainTheater, unit("P4minus2", 4, 5, Keywords{}))
	down.Modifiers = []Modifier{{Layer: 5, Power: -2}} // effective 2
	slot := ability.Target{Filter: paMax3(ability.Filter{Kinds: ability.KindUnit, Other: true})}
	s.script(0, Answer{Picks: []int{0}})
	s.iydDeploy(iydTrigger(false, []ability.Target{slot}, armorStep(1, 0)))
	d := paTargetAsk(t, s)
	want := ids(p2, p3, down)
	if got := refs(d); !slices.Equal(paSorted(got), paSorted(want)) {
		t.Errorf("'unit with power 3 or less' offers by effective power: printed 2, printed 3 and the unit lowered to 2; not the one raised to 4 nor printed 4: got %v, want %v (FR-TGT-001 line 375, Layer 6)", got, want)
	}
}

func TestFR_TGT_003_PowerAtMostTargetCheckedAgainOnResolution(t *testing.T) {
	// Covers: FR-TGT-001
	s := newScene(t)
	a := s.put(1, ZoneMainTheater, unit("A", 3, 5, Keywords{}))
	s.put(1, ZoneMainTheater, unit("B", 1, 5, Keywords{}))
	slot := ability.Target{Filter: paMax3(ability.Filter{Kinds: ability.KindUnit, Other: true})}
	// The chooser picks A, then A gains +1 power (effective 4) before the
	// effect resolves.
	s.asks = map[PlayerID][]Decision{}
	s.g.choosers[0] = func(_ *Game, d Decision) Answer {
		s.asks[0] = append(s.asks[0], d)
		for i, o := range d.Options {
			if o.Ref == a.ID {
				a.Modifiers = []Modifier{{Layer: 5, Power: 1}}
				return Answer{Picks: []int{i}}
			}
		}
		s.t.Errorf("A (power 3) is a legal target on placement: %+v", d.Options)
		return Answer{Timeout: true}
	}
	s.iydDeploy(iydTrigger(false, []ability.Target{slot}, armorStep(1, 0)))
	if len(s.asked(0)) == 0 {
		t.Fatalf("precondition: the target was chosen on placement")
	}
	if a.Armor != 0 {
		t.Errorf("a chosen unit that no longer has power 3 or less when the effect resolves is an illegal target and gains no armor: Armor %d, want 0 (FR-TGT-003 second check)", a.Armor)
	}
}

func TestFR_TGT_001_PowerAtMostAllStepReadsEffectivePower(t *testing.T) {
	// Covers: FR-TGT-002, FR-LAY-001
	s := newScene(t)
	allyLow := s.put(0, ZoneMainTheater, unit("AllyLow", 2, 5, Keywords{}))
	allyHigh := s.put(0, ZoneMainTheater, unit("AllyHigh", 5, 5, Keywords{}))
	allyDown := s.put(0, ZoneMainTheater, unit("AllyDown", 4, 5, Keywords{}))
	allyDown.Modifiers = []Modifier{{Layer: 5, Power: -1}} // effective 3
	foeEdge := s.put(1, ZoneMainTheater, unit("FoeEdge", 3, 5, Keywords{}))
	foeUp := s.put(1, ZoneMainTheater, unit("FoeUp", 3, 5, Keywords{}))
	foeUp.Modifiers = []Modifier{{Layer: 5, Power: 1}} // effective 4
	foeDown := s.put(1, ZoneMainTheater, unit("FoeDown", 5, 5, Keywords{}))
	foeDown.Modifiers = []Modifier{{Layer: 5, Power: -3}} // effective 2
	f := paMax3(ability.Filter{Kinds: ability.KindUnit, Other: true})
	s.script(0)
	s.iydDeploy(iydTrigger(false, nil, allStep(ability.OpAddArmor, 1, f)))
	for _, a := range []*Asset{allyLow, allyDown, foeEdge, foeDown} {
		if a.Armor != 1 {
			t.Errorf("%s has effective power 3 or less and gains 1 armor: Armor %d (FR-TGT-001 line 375)", a, a.Armor)
		}
	}
	for _, a := range []*Asset{allyHigh, foeUp} {
		if a.Armor != 0 {
			t.Errorf("%s has effective power above 3 and gains no armor: Armor %d (FR-TGT-001 line 375)", a, a.Armor)
		}
	}
}

func TestFR_TGT_001_PowerAtMostPilePickReadsPrintedPower(t *testing.T) {
	// Covers: SVY-R3, FR-ZON-004
	s := newScene(t)
	mine := s.inDiscard(0, unit("P2", 2, 1, Keywords{}), unit("P3", 3, 1, Keywords{}),
		unit("P4", 4, 1, Keywords{}), structure("Wall1", 1, 3, Keywords{}))
	s.inDiscard(1, unit("Theirs1", 1, 1, Keywords{}))
	s.ppScript(Answer{Picks: []int{0}})
	s.iydDeploy(iydTrigger(false, nil, ppPick(ability.OpReturnToHand, paMax3(ppAllyUnitCards), 0)))
	asks := ppPicks(s)
	if len(asks) != 1 {
		t.Fatalf("two cards match, so one pick question: got %d", len(asks))
	}
	want := ppInstIDs(mine[0], mine[1])
	if got := ppRefs(asks[0]); !slices.Equal(paSorted(got), paSorted(want)) {
		t.Errorf("the pick offers exactly your unit cards of printed power 3 or less: got %v, want %v (SVY-R3, FR-ZON-004)", got, want)
	}
}

func TestFR_TGT_001_NoStatCriterionIgnoresAtMost(t *testing.T) {
	// Covers: FR-TGT-001
	s := newScene(t)
	card := s.inDiscard(0, unit("Big Card", 5, 1, Keywords{}))[0]
	big := s.put(1, ZoneMainTheater, unit("Big", 5, 5, Keywords{}))
	pile := ability.Target{Filter: ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit, Zone: ability.ZoneDiscard, AtMost: 0}}
	field := ability.Target{Filter: ability.Filter{Allegiance: ability.AllegianceOpposing, Kinds: ability.KindUnit, AtMost: 0}}
	s.script(0)
	s.iydDeploy(iydTrigger(false, []ability.Target{pile, field},
		ptStep(ability.OpReturnToHand, 0), armorStep(1, 1)))
	if !inHand(s.g.Players[0], card) {
		t.Errorf("with no stat set, AtMost means nothing: the power 5 unit card in the pile is a legal target (FR-TGT-001)")
	}
	if big.Armor != 1 {
		t.Errorf("with no stat set, AtMost means nothing: the power 5 unit on the battlefield is a legal target: Armor %d, want 1 (FR-TGT-001)", big.Armor)
	}
}
