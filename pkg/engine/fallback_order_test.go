package engine

import (
	"reflect"
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for the fixed-order fallback of a mandatory target choice (T-0056,
// FR-MOD-001 sec. 6, SVY-R16, board directive #47 part 3). A timeout, a
// missing chooser or an invalid answer selects the first Min legal options in
// fixed order: the choosing player's assets before the opponent's; within each
// player the Headquarters, then Reserves, then the Main Theater, each zone in
// the order its assets entered it. Nothing about the pick is random.
// Cards here are synthetic: a unit's deploy trigger deals 2 damage to one
// chosen target. The trigger's source unit enters its controller's Reserves
// last, so it is itself an option whenever the filter allows that unit.

// fallbackTrigger deploys a source for p whose trigger deals 2 damage to one
// chosen target matching the filter.
func (s *scene) fallbackTrigger(p PlayerID, a ability.Allegiance, k ability.Kinds) *Asset {
	s.t.Helper()
	return s.deployTargeter(p, []ability.Target{tgt(a, k, false)}, dmgTo(2, 0))
}

// wantOnlyDamaged fails unless exactly want took the 2 damage and no other
// asset on the battlefield took any.
func wantOnlyDamaged(t *testing.T, s *scene, want *Asset, rule string) {
	t.Helper()
	for _, a := range s.g.Assets() {
		w := 0
		if a == want {
			w = 2
		}
		if a.Damage != w {
			t.Errorf("%s: %s damage = %d, want %d (the fixed-order first legal target is %s)", rule, a, a.Damage, w, want)
		}
	}
}

func TestFR_MOD_001_TargetTimeoutPicksOwnAssetBeforeOpponents(t *testing.T) {
	// Covers: SVY-R16
	s := newScene(t)
	theirs := s.put(1, ZoneMainTheater, structure("Their Structure", 0, 5, Keywords{})) // older ID
	mine := s.put(0, ZoneMainTheater, structure("My Structure", 0, 5, Keywords{}))
	if theirs.ID >= mine.ID {
		t.Fatalf("setup: the opponent's structure must be older (lower ID)")
	}
	s.script(0, Answer{Timeout: true})
	s.fallbackTrigger(0, ability.AllegianceAny, ability.KindStructure)
	if n := len(s.asked(0)); n != 1 {
		t.Fatalf("two options are asked once: %d", n)
	}
	wantOnlyDamaged(t, s, mine, "the choosing player's own assets come before the opponent's, even when older (SVY-R16)")
}

func TestFR_MOD_001_TargetTimeoutPicksHeadquartersThenReservesThenMainTheater(t *testing.T) {
	// Covers: SVY-R16
	t.Run("Headquarters first", func(t *testing.T) {
		s := newScene(t)
		s.put(0, ZoneMainTheater, unit("Theater Unit", 1, 5, Keywords{})) // oldest
		s.put(0, ZoneReserves, unit("Reserves Unit", 1, 5, Keywords{}))
		s.script(0, Answer{Timeout: true})
		s.fallbackTrigger(0, ability.AllegianceAllied, ability.KindUnit|ability.KindHeadquarters)
		wantOnlyDamaged(t, s, s.g.Players[0].HQ, "the Headquarters comes before Reserves and the Main Theater (SVY-R16)")
	})
	t.Run("Reserves before Main Theater", func(t *testing.T) {
		s := newScene(t)
		s.put(0, ZoneMainTheater, unit("Theater Unit", 1, 5, Keywords{})) // oldest
		r := s.put(0, ZoneReserves, unit("Reserves Unit", 1, 5, Keywords{}))
		s.script(0, Answer{Timeout: true})
		s.fallbackTrigger(0, ability.AllegianceAllied, ability.KindUnit)
		wantOnlyDamaged(t, s, r, "Reserves come before the Main Theater, even for an older Main Theater unit (SVY-R16)")
	})
}

func TestFR_MOD_001_TargetTimeoutOrdersAZoneByArrivalNotByID(t *testing.T) {
	// Covers: SVY-R16, SVY-R7
	s := newScene(t)
	older := s.put(1, ZoneReserves, unit("Older Unit", 1, 5, Keywords{}))
	newer := s.put(1, ZoneMainTheater, unit("Newer Unit", 1, 5, Keywords{}))
	s.moveTo(older, ZoneMainTheater) // enters the Main Theater after the newer unit
	if older.ID >= newer.ID {
		t.Fatalf("setup: the moved unit must have the lower ID")
	}
	s.script(0, Answer{Timeout: true})
	s.fallbackTrigger(0, ability.AllegianceOpposing, ability.KindUnit)
	wantOnlyDamaged(t, s, newer, "within a zone, assets are ordered by arrival, not by ID (SVY-R16, SVY-R7)")
}

func TestFR_MOD_001_TargetTimeoutOrderIsRelativeToTheChoosingPlayer(t *testing.T) {
	// Covers: SVY-R16
	t.Run("trigger controlled by player 1", func(t *testing.T) {
		s := newScene(t)
		s.endTurn() // player 1 is active
		s.put(0, ZoneMainTheater, structure("P0 Structure", 0, 5, Keywords{}))
		mine := s.put(1, ZoneMainTheater, structure("P1 Structure", 0, 5, Keywords{}))
		s.script(1, Answer{Timeout: true})
		s.fallbackTrigger(1, ability.AllegianceAny, ability.KindStructure)
		wantOnlyDamaged(t, s, mine, "the choosing player's assets come first, whichever seat that is (SVY-R16)")
	})
	t.Run("chooser is not the active player", func(t *testing.T) {
		s := newScene(t) // player 0 is active
		hq0 := s.g.Players[0].HQ
		u0 := s.put(0, ZoneMainTheater, unit("P0 Unit", 1, 5, Keywords{}))
		r1 := s.put(1, ZoneReserves, unit("P1 Reserves", 1, 5, Keywords{}))
		u1 := s.put(1, ZoneMainTheater, unit("P1 Theater", 1, 5, Keywords{}))
		options := []Option{{Label: "P0 HQ", Ref: hq0.ID}, {Label: "P0 Unit", Ref: u0.ID},
			{Label: "P1 Reserves", Ref: r1.ID}, {Label: "P1 Theater", Ref: u1.ID}}
		d := Decision{Player: 1, Kind: DecideTarget, Prompt: "target", Options: options, Min: 1, Max: 1}
		if got := s.g.choose(d); !reflect.DeepEqual(got, []int{2}) {
			t.Errorf("player 1 choosing on player 0's turn picks player 1's own first asset, their Reserves unit: got %v, want [2] (SVY-R16)", got)
		}
		d.Player = 0
		if got := s.g.choose(d); !reflect.DeepEqual(got, []int{0}) {
			t.Errorf("player 0 choosing picks their own Headquarters: got %v, want [0] (SVY-R16)", got)
		}
	})
}

func TestFR_MOD_001_TargetTimeoutOrdersOpponentsAssetsTheSameWay(t *testing.T) {
	// Covers: SVY-R16
	t.Run("Headquarters before Reserves", func(t *testing.T) {
		s := newScene(t)
		s.put(1, ZoneMainTheater, unit("Theater Unit", 1, 5, Keywords{})) // oldest
		s.put(1, ZoneReserves, unit("Reserves Unit", 1, 5, Keywords{}))
		s.script(0, Answer{Timeout: true})
		s.fallbackTrigger(0, ability.AllegianceOpposing, ability.KindUnit|ability.KindHeadquarters)
		wantOnlyDamaged(t, s, s.g.Players[1].HQ, "the opponent's Headquarters comes before their Reserves and Main Theater (SVY-R16)")
	})
	t.Run("Reserves before Main Theater", func(t *testing.T) {
		s := newScene(t)
		s.put(1, ZoneMainTheater, unit("Theater Unit", 1, 5, Keywords{})) // oldest
		r := s.put(1, ZoneReserves, unit("Reserves Unit", 1, 5, Keywords{}))
		s.script(0, Answer{Timeout: true})
		s.fallbackTrigger(0, ability.AllegianceOpposing, ability.KindUnit)
		wantOnlyDamaged(t, s, r, "the opponent's Reserves come before their Main Theater (SVY-R16)")
	})
}

func TestFR_MOD_001_TargetNoChooserAndInvalidAnswerPickLikeATimeout(t *testing.T) {
	// Covers: SVY-R16
	for _, tc := range []struct {
		name    string
		install func(s *scene)
	}{
		{"timeout", func(s *scene) { s.script(0, Answer{Timeout: true}) }},
		{"no chooser", func(s *scene) { s.g.choosers[0] = nil }},
		{"out of range", func(s *scene) { s.script(0, Answer{Picks: []int{9}}) }},
		{"negative", func(s *scene) { s.script(0, Answer{Picks: []int{-1}}) }},
		{"too many", func(s *scene) { s.script(0, Answer{Picks: []int{0, 1}}) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := newScene(t)
			s.put(0, ZoneMainTheater, unit("Theater Unit", 1, 5, Keywords{})) // oldest
			r := s.put(0, ZoneReserves, unit("Reserves Unit", 1, 5, Keywords{}))
			tc.install(s)
			s.fallbackTrigger(0, ability.AllegianceAllied, ability.KindUnit)
			wantOnlyDamaged(t, s, r, tc.name+" picks the first legal target in fixed order (FR-MOD-001 sec. 6, SVY-R16)")
		})
	}
}
