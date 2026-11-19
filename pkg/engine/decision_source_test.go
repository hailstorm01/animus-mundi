package engine

import (
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0233 (directive #149 (a), FR-MOD-001, FR-ZON-005): every
// decision asked while a Deploy, a Maneuver or an activated ability is declared
// names the card it comes from (Decision.Source), so a client can show that
// card's rules text, and SourceCard names a stack item's card as both seats see
// it. Cards are synthetic; player 0 declares.

// wantSources reports every decision in ds whose Source is not want, and an
// empty ds.
func wantSources(t *testing.T, ds []Decision, want *Card) {
	t.Helper()
	if len(ds) == 0 {
		t.Fatalf("no decision was asked")
	}
	for i, d := range ds {
		if d.Source != want {
			t.Errorf("decision %d (%s, %q): Source %v, want %s", i, d.Kind, d.Prompt, d.Source, want.Name)
		}
	}
}

func TestDecisionSource_ManeuverTargets(t *testing.T) {
	s := newScene(t)
	s.put(1, ZoneMainTheater, unit("X", 1, 9, ability.Keywords{}))
	s.put(1, ZoneMainTheater, unit("Y", 1, 9, ability.Keywords{}))
	c := maneuver("Ambush", 0, 0, anotherEffect())
	c.RulesText = "Chosen unit deals damage equal to its power to another chosen unit."
	inst := s.give(0, c)
	s.script(0, Answer{Picks: []int{0}})
	s.do(enactAction(inst))
	wantSources(t, s.asked(0), c)
	if got := s.asked(0)[0].Source.RulesText; got != c.RulesText {
		t.Errorf("Source.RulesText %q, want %q", got, c.RulesText)
	}
	if s.g.asking != nil {
		t.Errorf("after the declaration the scope is still open: asking %s, want nil", s.g.asking.Name)
	}
}

func TestDecisionSource_ManeuverModes(t *testing.T) {
	s := newScene(t)
	s.put(1, ZoneMainTheater, unit("X", 1, 9, ability.Keywords{}))
	s.put(1, ZoneMainTheater, unit("Y", 1, 9, ability.Keywords{}))
	c := maneuver("Pick One", 0, 0, ability.Effect{Modes: []ability.Effect{returnMode(), dealerMode()}, ChooseN: 1})
	inst := s.give(0, c)
	s.script(0, Answer{Picks: []int{0}}, Answer{Picks: []int{0}})
	s.do(enactAction(inst))
	wantSources(t, s.asked(0), c)
	if s.asked(0)[0].Kind != DecideMode {
		t.Errorf("first decision %s, want Mode", s.asked(0)[0].Kind)
	}
}

func TestDecisionSource_ActivatedCostPicks(t *testing.T) {
	s := newScene(t)
	costs := []ability.Cost{{Kind: ability.CostDiscard, UpTo: 1}, {Kind: ability.CostKill, Filter: alliedUnitFilter}}
	a := s.put(0, ZoneMainTheater, actUnit("Engine", nil, actAbility(costs, drawN(1))))
	s.put(0, ZoneMainTheater, unit("Fodder", 1, 1, ability.Keywords{}))
	s.give(0, unit("Spare", 1, 1, ability.Keywords{}))
	s.script(0, Answer{Picks: []int{}}, Answer{Picks: []int{1}})
	s.do(activate(a, 0))
	wantSources(t, s.asked(0), a.Inst.Card)
	if len(s.asked(0)) != 2 || s.asked(0)[0].Kind != DecideHand || s.asked(0)[1].Kind != DecidePick {
		t.Errorf("asked %d decisions, want the discard cost's Hand then the kill cost's Pick", len(s.asked(0)))
	}
}

func TestDecisionSource_DeployCostPick(t *testing.T) {
	s := newScene(t)
	s.put(0, ZoneMainTheater, unit("A", 1, 1, ability.Keywords{}))
	s.put(0, ZoneMainTheater, unit("B", 1, 1, ability.Keywords{}))
	c := killCostUnit("Sacrificer", 0, 0)
	inst := s.give(0, c)
	s.script(0, Answer{Picks: []int{0}})
	s.do(dcDeployToReserves(inst))
	wantSources(t, s.asked(0), c)
}

func TestDecisionSource_ScopesNestAndAnExplicitSourceStays(t *testing.T) {
	s := newScene(t)
	outer, inner, own := unit("Outer", 1, 1, ability.Keywords{}), unit("Inner", 1, 1, ability.Keywords{}), unit("Own", 1, 1, ability.Keywords{})
	ask := func(src *Card) *Card {
		s.script(0, Answer{Picks: []int{0}})
		s.g.choose(Decision{Player: 0, Kind: DecideTarget, Prompt: "pick", Min: 1, Max: 1, Source: src,
			Options: []Option{{Label: "a", Ref: 1}, {Label: "b", Ref: 2}}})
		return s.asked(0)[0].Source
	}
	if got := ask(nil); got != nil {
		t.Errorf("outside any scope: Source %v, want nil", got)
	}
	closeOuter := s.g.askFor(outer)
	if got := ask(nil); got != outer {
		t.Errorf("in the outer scope: Source %v, want Outer", got)
	}
	closeInner := s.g.askFor(inner)
	if got := ask(nil); got != inner {
		t.Errorf("in the inner scope: Source %v, want Inner", got)
	}
	if got := ask(own); got != own {
		t.Errorf("an explicit Source in a scope: Source %v, want Own", got)
	}
	closeInner()
	if got := ask(nil); got != outer {
		t.Errorf("after the inner scope closed: Source %v, want Outer", got)
	}
	closeOuter()
	if got := ask(nil); got != nil {
		t.Errorf("after both closed: Source %v, want nil", got)
	}
}

func TestSourceCard(t *testing.T) {
	s := newScene(t)
	a := s.put(0, ZoneMainTheater, unit("Hidden", 1, 1, ability.Keywords{}))
	it := &StackItem{Kind: StackAbility, Controller: 0, Source: a.ID, Inst: a.Inst}
	if got := s.g.SourceCard(it); got != a.Inst.Card {
		t.Errorf("a face-up source asset: %v, want its card", got)
	}
	a.FaceDown = true
	if got := s.g.SourceCard(it); got != scratchFace {
		t.Errorf("a face-down source asset: %v, want the Scratch face (FR-ZON-005)", got)
	}
	gone := s.inDiscard(0, unit("Gone", 1, 1, ability.Keywords{}))[0]
	if got := s.g.SourceCard(&StackItem{Kind: StackAbility, Source: 999, Inst: gone}); got != gone.Card {
		t.Errorf("a source asset that has left the battlefield: %v, want its card instance's card", got)
	}
	m := s.give(0, maneuver("Played", 0, 0, anotherEffect()))
	if got := s.g.SourceCard(&StackItem{Kind: StackManeuver, Inst: m}); got != m.Card {
		t.Errorf("a Maneuver: %v, want its own card", got)
	}
	if got := s.g.SourceCard(&StackItem{Kind: StackAbility}); got != nil {
		t.Errorf("no source asset and no instance: %v, want nil", got)
	}
}
