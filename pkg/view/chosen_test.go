package view_test

import (
	"reflect"
	"testing"

	"animus-mundi/pkg/engine"
	"animus-mundi/pkg/engine/ability"
	"animus-mundi/pkg/view"
)

// T-0293 (directive #179): an asset slot's Chosen holds only the chosen assets
// still on the battlefield, as a pile slot's ChosenCards holds only the cards
// still in a discard pile. A slot with nothing left stays, empty.
func TestStackView_ChosenOmitsAssetsThatLeft(t *testing.T) {
	s := newScene(t)
	src := s.put(0, engine.ZoneMainTheater, testUnit("Chooser", 1, 1, ability.Keywords{}))
	stays := s.put(1, engine.ZoneMainTheater, testUnit("Still Here", 1, 1, ability.Keywords{}))
	gone := s.put(1, engine.ZoneReserves, testUnit("Gone Away", 1, 1, ability.Keywords{}))
	s.g.Stack = []*engine.StackItem{{
		Kind: engine.StackAbility, Controller: 0, Source: src.ID, Inst: src.Inst,
		Chosen: [][]int{{gone.ID, stays.ID}, {gone.ID}},
	}}
	s.g.Players[1].Reserves = nil // gone leaves the battlefield before the view is taken

	for seat := engine.PlayerID(0); seat < 2; seat++ {
		v := view.PlayerView(s.g, seat)
		if len(v.Stack) != 1 {
			t.Fatalf("seat %d: stack = %+v", seat, v.Stack)
		}
		if got, want := v.Stack[0].Chosen, [][]int{{stays.ID}, {}}; !reflect.DeepEqual(got, want) {
			t.Errorf("seat %d: Chosen = %v, want %v", seat, got, want)
		}
	}
}
