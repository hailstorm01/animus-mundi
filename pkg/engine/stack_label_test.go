package engine

import (
	"strings"
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0239 (directive #153, FR-ZON-005): one function, StackLabel,
// words what a stack item is, and the place event reads it. An ability names
// its card as both seats see it, so a face-down source shows only the Scratch
// face, and its place event carries no card instance ID. Cards are synthetic;
// player 0 controls.

func TestStackLabel_EachKind(t *testing.T) {
	s := newScene(t)
	a := s.put(0, ZoneMainTheater, unit("Open Source", 1, 1, Keywords{}))
	d := s.give(0, unit("Deployed", 1, 1, Keywords{}))
	m := s.give(0, maneuver("Played", 0, 0, anotherEffect()))
	gone := s.inDiscard(0, unit("Gone", 1, 1, Keywords{}))[0]
	for _, tc := range []struct {
		it   *StackItem
		want string
	}{
		{&StackItem{Kind: StackDeploy, Inst: d, Zone: ZoneReserves}, "Deploy of Deployed"},
		{&StackItem{Kind: StackMarch, Source: a.ID, Zone: ZoneMainTheater}, "March"},
		{&StackItem{Kind: StackAttack, Source: a.ID}, "Attack"},
		{&StackItem{Kind: StackScratch}, "Scratch"},
		{&StackItem{Kind: StackManeuver, Inst: m}, "Maneuver Played"},
		{&StackItem{Kind: StackAbility, Source: a.ID, Inst: a.Inst}, "ability of Open Source"},
		{&StackItem{Kind: StackAbility, Source: a.ID, Inst: a.Inst, Keyword: ability.KeywordAggressive}, "Aggressive trigger of Open Source"},
		{&StackItem{Kind: StackAbility, Source: 999, Inst: gone, Tracked: true}, "ability of Gone"},
	} {
		if got := s.g.StackLabel(tc.it); got != tc.want {
			t.Errorf("StackLabel(%v item) = %q, want %q", tc.it.Kind, got, tc.want)
		}
	}
}

func TestStackLabel_FaceDownSourceShowsOnlyTheScratchFace(t *testing.T) {
	s := newScene(t)
	a := s.put(0, ZoneMainTheater, unit("Hidden Source", 1, 1, Keywords{}))
	a.FaceDown = true
	it := &StackItem{Kind: StackAbility, Controller: 0, Source: a.ID, Inst: a.Inst}
	if got, want := s.g.StackLabel(it), "ability of Scratch Company"; got != want {
		t.Errorf("StackLabel = %q, want %q (FR-ZON-005)", got, want)
	}
	s.g.push(it)
	pl := eventsOfKind(s.g.Log, "place")
	if len(pl) == 0 {
		t.Fatalf("push logged no place event")
	}
	e := pl[len(pl)-1]
	if e.Card != "Scratch Company" {
		t.Errorf("place event Card = %q, want the Scratch face", e.Card)
	}
	if e.Inst != 0 {
		t.Errorf("place event Inst = %d, want 0: an instance ID could identify the face-down card", e.Inst)
	}
	if want := "player 0 places ability of Scratch Company on the stack"; e.Detail != want {
		t.Errorf("place event Detail = %q, want %q", e.Detail, want)
	}
	if strings.Contains(e.Detail, "Hidden Source") || e.Card == "Hidden Source" {
		t.Errorf("the place event names the face-down card: %+v", e)
	}
}

func TestStackLabel_FaceUpSourcePlaceEventNamesItsInstance(t *testing.T) {
	s := newScene(t)
	a := s.put(0, ZoneMainTheater, unit("Open Source", 1, 1, Keywords{}))
	s.g.push(&StackItem{Kind: StackAbility, Controller: 0, Source: a.ID, Inst: a.Inst, Keyword: ability.KeywordAggressive})
	pl := eventsOfKind(s.g.Log, "place")
	if len(pl) == 0 {
		t.Fatalf("push logged no place event")
	}
	e := pl[len(pl)-1]
	if e.Card != "Open Source" || e.Inst != a.Inst.ID {
		t.Errorf("place event Card %q, Inst %d; want %q, %d", e.Card, e.Inst, "Open Source", a.Inst.ID)
	}
	if want := "player 0 places Aggressive trigger of Open Source on the stack"; e.Detail != want {
		t.Errorf("place event Detail = %q, want %q", e.Detail, want)
	}
}
