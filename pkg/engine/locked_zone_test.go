package engine

import (
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for the zone locked on a stack item (T-0205; SVY-R44, SVY-R45, FRD 01
// FR-TGT-001, "That Zone" After a Move). A trigger on a move locks the zone
// the move went to, fixed by the event. These tests drive the steps directly
// (move, flush) to see the item between placement and resolution.
//
// This file does not compile until StackItem has a Locked field of type
// *LockedZone (Zone, Player); that is expected until T-0205 is implemented.

// flushedItems flushes the trigger buffer and returns the items it placed.
func flushedItems(s *scene) []*StackItem {
	s.t.Helper()
	before := len(s.g.Stack)
	s.g.flushTriggers()
	return append([]*StackItem(nil), s.g.Stack[before:]...)
}

func wantLocked(t *testing.T, it *StackItem, zone Zone, p PlayerID, why string) {
	t.Helper()
	if it.Locked == nil {
		t.Fatalf("%s: the item locked no zone, want %v (player %d)", why, zone, p)
	}
	if it.Locked.Zone != zone {
		t.Errorf("%s: locked zone is %v, want %v", why, it.Locked.Zone, zone)
	}
	if zone != ZoneMainTheater && it.Locked.Player != p {
		t.Errorf("%s: locked zone belongs to player %d, want %d (FR-ZON-001)", why, it.Locked.Player, p)
	}
}

func TestFR_TGT_001_MoveLocksMainTheater(t *testing.T) {
	// Covers: SVY-R45
	s := newScene(t)
	w := s.put(0, ZoneReserves, moveWatcher("Watcher"))
	s.g.moveOnBattlefield(w, ZoneMainTheater)
	items := flushedItems(s)
	if len(items) != 1 {
		t.Fatalf("expected the move trigger placed once, got %d items", len(items))
	}
	wantLocked(t, items[0], ZoneMainTheater, 0, "a move to the Main Theater locks the Main Theater (SVY-R45)")
}

func TestFR_TGT_001_MoveLocksOwnReserves(t *testing.T) {
	// Covers: SVY-R45, FR-ZON-001
	s := newScene(t)
	w := s.put(1, ZoneMainTheater, moveWatcher("Watcher"))
	s.g.moveOnBattlefield(w, ZoneReserves)
	items := flushedItems(s)
	if len(items) != 1 {
		t.Fatalf("expected the move trigger placed once, got %d items", len(items))
	}
	wantLocked(t, items[0], ZoneReserves, 1, "a move to Reserves locks that player's own Reserves (SVY-R45, FR-ZON-001)")
}

func TestFR_TGT_001_EachMoveLocksItsDestination(t *testing.T) {
	// Covers: SVY-R45
	s := newScene(t)
	w := s.put(0, ZoneReserves, moveWatcher("Watcher"))
	s.g.moveOnBattlefield(w, ZoneMainTheater)
	s.g.moveOnBattlefield(w, ZoneReserves)
	items := flushedItems(s)
	if len(items) != 2 {
		t.Fatalf("expected one trigger per move, got %d items", len(items))
	}
	main, res := 0, 0
	for _, it := range items {
		if it.Locked == nil {
			t.Fatalf("a move trigger locked no zone (SVY-R45)")
		}
		switch it.Locked.Zone {
		case ZoneMainTheater:
			main++
		case ZoneReserves:
			res++
			if it.Locked.Player != 0 {
				t.Errorf("locked Reserves belongs to player %d, want 0", it.Locked.Player)
			}
		}
	}
	if main != 1 || res != 1 {
		t.Errorf("each move's trigger locks its own destination, fixed by the event (SVY-R45): got %d Main Theater and %d Reserves locks, want one each", main, res)
	}
}

func TestFR_TGT_001_LockedZoneOutlivesSource(t *testing.T) {
	// Covers: SVY-R45, FR-ZON-004
	t.Run("source leaves before the flush", func(t *testing.T) {
		s := newScene(t)
		w := s.put(0, ZoneReserves, moveWatcher("Watcher"))
		s.g.moveOnBattlefield(w, ZoneMainTheater)
		s.g.leaveBattlefield(w, s.g.effectiveState(w))
		items := flushedItems(s)
		if len(items) != 1 {
			t.Fatalf("a trigger whose source left before the flush is still placed, got %d items", len(items))
		}
		wantLocked(t, items[0], ZoneMainTheater, 0, "the zone is fixed by the move, not read from the departed source (SVY-R45, FR-ZON-004)")
	})
	t.Run("source moves again after the flush", func(t *testing.T) {
		s := newScene(t)
		w := s.put(0, ZoneReserves, moveWatcher("Watcher"))
		s.g.moveOnBattlefield(w, ZoneMainTheater)
		items := flushedItems(s)
		if len(items) != 1 {
			t.Fatalf("expected the move trigger placed once, got %d items", len(items))
		}
		s.moveTo(w, ZoneReserves)
		wantLocked(t, items[0], ZoneMainTheater, 0, "a later move of the source never changes a locked zone (SVY-R45)")
	})
}

func TestFR_TGT_001_NoLockWithoutMove(t *testing.T) {
	s := newScene(t)
	c := striker("Deployer", 1, 5, Keywords{}, ability.Trigger{Hook: ability.HookDeploy, Effect: ability.Effect{Steps: []ability.Step{scratchN(1)}}})
	inst := &Instance{ID: s.g.newID(), Card: c, Owner: 0}
	s.g.deployCard(inst, 0, ZoneMainTheater)
	items := flushedItems(s)
	if len(items) != 1 {
		t.Fatalf("expected the deploy trigger placed once, got %d items", len(items))
	}
	if items[0].Locked != nil {
		t.Errorf("a deploy trigger is not a move, so it locks no zone (SVY-R45), got %+v", *items[0].Locked)
	}
}
