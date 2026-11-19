package engine

import (
	"animus-mundi/pkg/engine/ability"
	"reflect"
	"testing"
)

// Tests for T-0255: Rhapsody's Hollow, "T3, exhaust, kill this structure: Gain
// control of chosen opposing unit in the Main Theater." (FR-ACT-001, FR-ACT-002,
// FR-ZON-002, FR-ZON-003, FR-TGT-003, SVY-R36). The real card is loaded from
// data/cards. Player 0 is the active player and activates.

// hollowScene is a scene with Rhapsody's Hollow Readied in player 0's Main Theater
// and exactly R0 T3 in player 0's pool.
func hollowScene(t *testing.T) (*scene, *Asset) {
	t.Helper()
	s := newScene(t)
	h := s.put(0, ZoneMainTheater, s.card("Rhapsody's Hollow"))
	s.setPool(0, 0, 3)
	return s, h
}

func TestCard_RhapsodysHollow_GainsControlOfOpposingUnit(t *testing.T) {
	// Covers: FR-ZON-002, FR-ZON-003, FR-ACT-001
	s, h := hollowScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Foe", 3, 5, ability.Keywords{}))
	foe.Damage = 1
	s.script(0, Answer{Picks: []int{0}})
	from := len(s.g.Log)
	s.do(activate(h, 0))
	got := s.g.Asset(foe.ID)
	if got != foe || !inAssets(s.g.Theater, got) {
		t.Fatalf("the unit is the same asset still in the Main Theater (FR-ZON-002: a control change keeps its zone)")
	}
	if got.Controller != 0 {
		t.Errorf("controller %d, want 0 (the activating player gains control)", got.Controller)
	}
	if got.Inst.Owner != 1 {
		t.Errorf("owner %d, want 1 (ownership never changes, FR-ZON-003)", got.Inst.Owner)
	}
	if got.Exhausted || got.Damage != 1 {
		t.Errorf("exhausted %v, damage %d; want Readied with 1 damage (state is kept)", got.Exhausted, got.Damage)
	}
	evs := effectEvents(s.g, from, EventControl)
	if len(evs) != 1 || evs[0].Target != foe.ID {
		t.Errorf("want one control event naming the unit, got %+v", evs)
	}
	if s.alive(h) {
		t.Errorf("the Hollow is killed as a cost and is off the battlefield (FR-ACT-001)")
	}
	p := s.g.Players[0]
	if !inDiscardPile(p, h.Inst) {
		t.Errorf("the Hollow's card is in its owner's discard pile (FR-ACT-001)")
	}
	if p.Req != 0 || p.Tac != 0 {
		t.Errorf("exactly 3 Tactics paid: pool R%d T%d, want R0 T0 (FR-ACT-001)", p.Req, p.Tac)
	}
}

func TestCard_RhapsodysHollow_OnlyOpposingMainTheaterUnitsOffered(t *testing.T) {
	// Covers: FR-TGT-001
	s, h := hollowScene(t)
	s.put(0, ZoneMainTheater, unit("Ally", 2, 5, ability.Keywords{}))
	s.put(1, ZoneReserves, unit("Reserve", 2, 5, ability.Keywords{}))
	s.put(1, ZoneMainTheater, structure("Wall", 0, 5, ability.Keywords{}))
	a := s.put(1, ZoneMainTheater, unit("FoeA", 2, 5, ability.Keywords{}))
	b := s.put(1, ZoneMainTheater, unit("FoeB", 3, 5, ability.Keywords{}))
	s.script(0, Answer{Picks: []int{0}})
	s.do(activate(h, 0))
	asks := s.asked(0)
	if len(asks) != 1 || asks[0].Kind != DecideTarget {
		t.Fatalf("one target question is asked, got %d decisions (FR-TGT-001)", len(asks))
	}
	if got, want := refs(asks[0]), sortedCopy([]int{a.ID, b.ID}); !reflect.DeepEqual(got, want) {
		t.Errorf("options are player 1's two Main Theater units only: got %v, want %v (FR-TGT-001)", got, want)
	}
	chosen := asks[0].Options[0].Ref
	for _, x := range []*Asset{a, b} {
		if wantCtl := PlayerID(1); x.ID == chosen {
			wantCtl = 0
			if x.Controller != wantCtl {
				t.Errorf("the chosen unit %s has controller %d, want 0", x, x.Controller)
			}
		} else if x.Controller != wantCtl {
			t.Errorf("the unit not chosen %s has controller %d, want 1", x, x.Controller)
		}
	}
}

func TestCard_RhapsodysHollow_NotOfferedWithoutTarget(t *testing.T) {
	// Covers: SVY-R36, FR-ACT-001
	t.Run("no other asset", func(t *testing.T) {
		s, h := hollowScene(t)
		actNotOffered(t, s, h, 0, "no other asset on the battlefield")
	})
	t.Run("only an allied unit", func(t *testing.T) {
		s, h := hollowScene(t)
		s.put(0, ZoneMainTheater, unit("Ally", 2, 5, ability.Keywords{}))
		actNotOffered(t, s, h, 0, "only an allied unit in the Main Theater")
	})
	t.Run("only an opposing unit in Reserves", func(t *testing.T) {
		s, h := hollowScene(t)
		s.put(1, ZoneReserves, unit("Reserve", 2, 5, ability.Keywords{}))
		actNotOffered(t, s, h, 0, "only an opposing unit in its Reserves")
	})
	t.Run("only an opposing structure", func(t *testing.T) {
		s, h := hollowScene(t)
		s.put(1, ZoneMainTheater, structure("Wall", 0, 5, ability.Keywords{}))
		actNotOffered(t, s, h, 0, "only an opposing structure in the Main Theater")
	})
}

func TestCard_RhapsodysHollow_CostsNeedTacticsAndReadied(t *testing.T) {
	// Covers: FR-ACT-001, FR-ACT-002
	t.Run("Requisition cannot pay Tactics", func(t *testing.T) {
		s, h := hollowScene(t)
		s.put(1, ZoneMainTheater, unit("Foe", 2, 5, ability.Keywords{}))
		s.setPool(0, 5, 2)
		actNotOffered(t, s, h, 0, "R5 T2 cannot pay T3")
	})
	t.Run("Exhausted", func(t *testing.T) {
		s, h := hollowScene(t)
		s.put(1, ZoneMainTheater, unit("Foe", 2, 5, ability.Keywords{}))
		h.Exhausted = true
		if offersActivate(s.g, h, 0) {
			t.Errorf("an Exhausted Hollow cannot pay its exhaust cost, so Activate is not offered (FR-ACT-002)")
		}
		s.reject(activate(h, 0))
		if p := s.g.Players[0]; !s.alive(h) || p.Tac != 3 || len(p.Discard) != 0 {
			t.Errorf("a rejected Activate pays nothing and the Hollow stays: alive %v T%d (FR-ACT-001)", s.alive(h), p.Tac)
		}
	})
}

func TestCard_RhapsodysHollow_TargetLeftMainTheaterBeforeResolution(t *testing.T) {
	// Covers: FR-TGT-003, FR-ACT-001
	s, h := hollowScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Foe", 3, 5, ability.Keywords{}))
	s.script(0, Answer{Picks: []int{0}})
	moved := false
	s.g.obs = &recorder{onEvent: func(e Event) {
		if e.Kind == "place" && !moved {
			moved = true
			s.moveTo(foe, ZoneReserves)
		}
	}}
	from := len(s.g.Log)
	s.do(activate(h, 0))
	if !moved {
		t.Fatalf("the ability was placed on the stack")
	}
	if foe.Controller != 1 || !inAssets(s.g.Players[1].Reserves, foe) {
		t.Errorf("the unit stays in player 1's Reserves under player 1's control, controller %d (FR-TGT-003)", foe.Controller)
	}
	if ev := effectEvents(s.g, from, EventControl); len(ev) != 0 {
		t.Errorf("no control event is logged, got %+v (FR-TGT-003)", ev)
	}
	p := s.g.Players[0]
	if !inDiscardPile(p, h.Inst) || p.Tac != 0 {
		t.Errorf("the costs stay paid: Hollow in discard %v, T%d, want true and T0 (FR-ACT-001)", inDiscardPile(p, h.Inst), p.Tac)
	}
}
