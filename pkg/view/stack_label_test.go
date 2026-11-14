package view_test

import (
	"slices"
	"strings"
	"testing"

	"animus-mundi/pkg/engine"
	"animus-mundi/pkg/engine/ability"
	"animus-mundi/pkg/view"
)

// Tests for T-0239 (directive #153, FR-ZON-005): a stack item's Label is the
// engine's own wording (Game.StackLabel), and an ability whose source is face
// down shows only the Scratch face to both seats.

func TestView_StackItemLabel(t *testing.T) {
	s := newScene(t)
	up := s.put(0, engine.ZoneMainTheater, testUnit("Open Source", 1, 1, engine.Keywords{}))
	down := s.put(1, engine.ZoneMainTheater, testUnit("Secret Source", 1, 1, engine.Keywords{}))
	down.FaceDown = true
	s.g.Stack = []*engine.StackItem{
		{Kind: engine.StackAbility, Controller: 0, Source: up.ID, Inst: up.Inst, Keyword: ability.KeywordAggressive},
		{Kind: engine.StackAbility, Controller: 1, Source: down.ID, Inst: down.Inst},
		{Kind: engine.StackMarch, Controller: 0, Source: up.ID, Zone: engine.ZoneMainTheater},
	}
	want := []string{"Aggressive trigger of Open Source", "ability of Scratch Company", "March"}
	for seat := engine.PlayerID(0); seat < 2; seat++ {
		v := view.PlayerView(s.g, seat)
		var got []string
		for _, it := range v.Stack {
			got = append(got, it.Label)
		}
		if !slices.Equal(got, want) {
			t.Errorf("seat %d: labels %q, want %q", seat, got, want)
		}
		if js := mustJSON(t, v.Stack); strings.Contains(js, "Secret Source") {
			t.Errorf("seat %d: the stack names the face-down card: %s", seat, js)
		}
	}
}
