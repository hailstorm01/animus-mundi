package engine

import (
	"sort"
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for the "in that zone" filter criterion (T-0206; SVY-R45, FRD 01
// FR-TGT-001 "That Zone" After a Move, FR-TGT-003). A move trigger whose
// effect uses ability.Filter{ZoneRef: ability.ZoneRefThat} reads the zone its
// stack item locked (the move's destination), at both target checks and in
// untargeted steps.
//
// This file does not compile until ability has ZoneRef, ZoneRefThat and
// Filter.ZoneRef; that is expected until T-0206 is implemented.

var thatZone = ability.Filter{ZoneRef: ability.ZoneRefThat}

// tzField is the cast of one scene: a watcher (player 0) and one other unit in
// each of the four battlefield zones that matter.
type tzField struct {
	s            *scene
	w            *Asset
	main0, main1 *Asset
	res0, res1   *Asset
}

func tzScene(t *testing.T, watcher *Card, watcherZone Zone) *tzField {
	t.Helper()
	s := newScene(t)
	f := &tzField{s: s}
	f.w = s.put(0, watcherZone, watcher)
	f.main0 = s.put(0, ZoneMainTheater, unit("Main0", 1, 5, Keywords{}))
	f.main1 = s.put(1, ZoneMainTheater, unit("Main1", 1, 5, Keywords{}))
	f.res0 = s.put(0, ZoneReserves, unit("Res0", 1, 5, Keywords{}))
	f.res1 = s.put(1, ZoneReserves, unit("Res1", 1, 5, Keywords{}))
	return f
}

// tzScript installs a chooser for seat 0 that records every Decision, picks
// the first option whose Ref is in ids on a target question (a timeout when
// none matches), and times out on anything else.
func tzScript(s *scene, ids ...int) {
	if s.asks == nil {
		s.asks = map[PlayerID][]Decision{}
	}
	s.asks[0] = nil
	s.g.choosers[0] = func(_ *Game, d Decision) Answer {
		s.asks[0] = append(s.asks[0], d)
		if d.Kind == DecideTarget {
			for i, o := range d.Options {
				for _, id := range ids {
					if o.Ref == id {
						return Answer{Picks: []int{i}}
					}
				}
			}
		}
		return Answer{Timeout: true}
	}
}

// tzOffered returns the sorted asset IDs offered by the first decision of the
// given kind seat 0 was asked, and whether there was one.
func tzOffered(s *scene, kind DecisionKind) ([]int, bool) {
	for _, d := range s.asked(0) {
		if d.Kind == kind {
			var ids []int
			for _, o := range d.Options {
				ids = append(ids, o.Ref)
			}
			sort.Ints(ids)
			return ids, true
		}
	}
	return nil, false
}

func tzWantOffered(t *testing.T, s *scene, kind DecisionKind, why string, want ...*Asset) {
	t.Helper()
	got, ok := tzOffered(s, kind)
	if !ok {
		t.Fatalf("%s: no %v decision was put to player 0", why, kind)
	}
	var ids []int
	for _, a := range want {
		ids = append(ids, a.ID)
	}
	sort.Ints(ids)
	if len(got) != len(ids) {
		t.Errorf("%s: offered asset IDs %v, want exactly %v", why, got, ids)
		return
	}
	for i := range ids {
		if got[i] != ids[i] {
			t.Errorf("%s: offered asset IDs %v, want exactly %v", why, got, ids)
			return
		}
	}
}

func tzWantArmor(t *testing.T, a *Asset, want int, why string) {
	t.Helper()
	if a.Armor != want {
		t.Errorf("%s: %s armor = %d, want %d", why, a, a.Armor, want)
	}
}

// tzTargetWatcher prints "whenever this moves, target an asset in that zone;
// it gains 1 armor".
func tzTargetWatcher() *Card {
	return striker("Watcher", 1, 5, Keywords{}, ability.Trigger{
		Hook:   ability.HookMoves,
		ToZone: ability.ZoneAny,
		Effect: ability.Effect{
			Targets: []ability.Target{{Filter: thatZone}},
			Steps:   []ability.Step{armorStep(1, 0)},
		},
	})
}

func tzAllWatcher() *Card {
	return striker("Watcher", 1, 5, Keywords{}, moveTrigger(ability.ZoneAny, allStep(ability.OpAddArmor, 1, thatZone)))
}

func tzPickWatcher() *Card {
	return striker("Watcher", 1, 5, Keywords{}, moveTrigger(ability.ZoneAny,
		ability.Step{Op: ability.OpKill, Subject: ability.SubjectPick, Filter: thatZone, UpTo: 1}))
}

func TestFR_TGT_001_ThatZoneOffersLockedZoneOnly(t *testing.T) {
	// Covers: SVY-R45, FR-ZON-001
	t.Run("Main Theater lock", func(t *testing.T) {
		f := tzScene(t, tzTargetWatcher(), ZoneReserves)
		tzScript(f.s, f.main0.ID)
		f.s.g.moveOnBattlefield(f.w, ZoneMainTheater)
		f.s.g.flushTriggers()
		tzWantOffered(t, f.s, DecideTarget, "a move to the Main Theater locks it; \"in that zone\" offers the Main Theater units, both players', and no Reserves unit (SVY-R45, FR-ZON-001)",
			f.w, f.main0, f.main1)
	})
	t.Run("Reserves lock", func(t *testing.T) {
		f := tzScene(t, tzTargetWatcher(), ZoneMainTheater)
		tzScript(f.s, f.res0.ID)
		f.s.g.moveOnBattlefield(f.w, ZoneReserves)
		f.s.g.flushTriggers()
		tzWantOffered(t, f.s, DecideTarget, "a move to Reserves locks that player's own Reserves; \"in that zone\" offers it, not the other player's Reserves nor the Main Theater (SVY-R45, FR-ZON-001)",
			f.w, f.res0)
	})
}

func TestFR_TGT_003_ThatZoneTargetLeftZoneIsIllegal(t *testing.T) {
	// Covers: SVY-R45
	f := tzScene(t, tzTargetWatcher(), ZoneReserves)
	tzScript(f.s, f.main1.ID)
	f.s.g.moveOnBattlefield(f.w, ZoneMainTheater)
	f.s.g.flushTriggers()
	f.s.moveTo(f.main1, ZoneReserves)
	f.s.g.resolveStack()
	tzWantArmor(t, f.main1, 0, "a target that left the locked zone is illegal at the second check and takes no effect (FR-TGT-003, SVY-R45)")
}

func TestFR_TGT_003_ThatZoneTargetOutlivesSource(t *testing.T) {
	// Covers: SVY-R45, FR-ZON-004
	t.Run("source moves back to Reserves", func(t *testing.T) {
		f := tzScene(t, tzTargetWatcher(), ZoneReserves)
		tzScript(f.s, f.main1.ID)
		f.s.g.moveOnBattlefield(f.w, ZoneMainTheater)
		f.s.g.flushTriggers()
		f.s.moveTo(f.w, ZoneReserves)
		f.s.g.resolveStack()
		tzWantArmor(t, f.main1, 1, "the second check reads the locked zone, not the source's current zone (FR-TGT-003, SVY-R45)")
	})
	t.Run("source leaves the battlefield", func(t *testing.T) {
		f := tzScene(t, tzTargetWatcher(), ZoneReserves)
		tzScript(f.s, f.main1.ID)
		f.s.g.moveOnBattlefield(f.w, ZoneMainTheater)
		f.s.g.flushTriggers()
		f.s.g.leaveBattlefield(f.w, f.s.g.effectiveState(f.w))
		f.s.g.resolveStack()
		tzWantArmor(t, f.main1, 1, "the locked zone outlives the departed source (FR-ZON-004, SVY-R45)")
	})
}

func TestFR_TGT_001_ThatZoneAllStepActsOnLockedZone(t *testing.T) {
	// Covers: SVY-R45, FR-ZON-001, SVY-R3
	t.Run("Main Theater lock", func(t *testing.T) {
		f := tzScene(t, tzAllWatcher(), ZoneReserves)
		tzScript(f.s)
		f.s.g.moveOnBattlefield(f.w, ZoneMainTheater)
		f.s.g.flushTriggers()
		f.s.g.resolveStack()
		why := "an untargeted step with \"in that zone\" acts on every asset in the locked Main Theater and nothing else (SVY-R45)"
		tzWantArmor(t, f.w, 1, why)
		tzWantArmor(t, f.main0, 1, why)
		tzWantArmor(t, f.main1, 1, why)
		tzWantArmor(t, f.res0, 0, why)
		tzWantArmor(t, f.res1, 0, why)
	})
	t.Run("Reserves lock", func(t *testing.T) {
		f := tzScene(t, tzAllWatcher(), ZoneMainTheater)
		tzScript(f.s)
		f.s.g.moveOnBattlefield(f.w, ZoneReserves)
		f.s.g.flushTriggers()
		f.s.g.resolveStack()
		why := "an untargeted step with \"in that zone\" acts on the locked player's own Reserves only (SVY-R45, FR-ZON-001)"
		tzWantArmor(t, f.w, 1, why)
		tzWantArmor(t, f.res0, 1, why)
		tzWantArmor(t, f.res1, 0, why)
		tzWantArmor(t, f.main0, 0, why)
		tzWantArmor(t, f.main1, 0, why)
	})
}

func TestFR_TGT_001_ThatZonePickOffersLockedZone(t *testing.T) {
	// Covers: SVY-R45, SVY-R14
	f := tzScene(t, tzPickWatcher(), ZoneReserves)
	tzScript(f.s)
	f.s.g.moveOnBattlefield(f.w, ZoneMainTheater)
	f.s.g.flushTriggers()
	f.s.moveTo(f.res1, ZoneMainTheater) // enters the locked zone after the flush
	f.s.moveTo(f.main0, ZoneReserves)   // leaves the locked zone after the flush
	f.s.g.resolveStack()
	tzWantOffered(t, f.s, DecidePick, "an untargeted pick reads the locked zone when it resolves: it offers the assets then in the Main Theater and not those in Reserves (SVY-R45, SVY-R14)",
		f.w, f.main1, f.res1)
}

func TestFR_TGT_001_ThatZoneWithoutLockMatchesNothing(t *testing.T) {
	// Covers: SVY-R45
	s := newScene(t)
	main0 := s.put(0, ZoneMainTheater, unit("Main0", 1, 5, Keywords{}))
	res0 := s.put(0, ZoneReserves, unit("Res0", 1, 5, Keywords{}))
	main1 := s.put(1, ZoneMainTheater, unit("Main1", 1, 5, Keywords{}))
	tzScript(s, main0.ID, res0.ID, main1.ID)
	c := striker("Deployer", 1, 5, Keywords{}, ability.Trigger{
		Hook: ability.HookDeploy,
		Effect: ability.Effect{
			Targets: []ability.Target{{Filter: thatZone}},
			Steps:   []ability.Step{armorStep(1, 0), allStep(ability.OpAddArmor, 1, thatZone)},
		},
	})
	inst := &Instance{ID: s.g.newID(), Card: c, Owner: 0}
	s.g.deployCard(inst, 0, ZoneMainTheater)
	s.g.flushTriggers()
	s.g.resolveStack()
	if ids, ok := tzOffered(s, DecideTarget); ok && len(ids) > 0 {
		t.Errorf("a filter with \"in that zone\" and no locked zone matches nothing, but the target question offered %v (#141 (1))", ids)
	}
	why := "an item with no locked zone: \"in that zone\" matches nothing (#141 (1))"
	tzWantArmor(t, main0, 0, why)
	tzWantArmor(t, res0, 0, why)
	tzWantArmor(t, main1, 0, why)
	if a := assetOf(s.g, c); a != nil {
		tzWantArmor(t, a, 0, why)
	}
}
