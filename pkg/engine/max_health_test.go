package engine

import "testing"

// Tests for Maximum Health modifiers in Layer 5 (FR-LAY-001, FR-LAY-002) and
// the state-based death sweep after Sub-Phase 4c (FR-TRN-001, board directive
// #15 (3)). No step makes a max-health modifier yet, so the tests give it
// directly on the asset.

func addMaxHealth(a *Asset, n int) {
	a.Modifiers = append(a.Modifiers, Modifier{Layer: 5, MaxHealth: n})
}

// sweepNow makes a stack item resolve, and so a death sweep run, by deploying
// a free unit from player 0's hand.
func sweepNow(s *scene) {
	s.t.Helper()
	c := unit("Sweep Trigger", 1, 1, Keywords{})
	c.ReqYield, c.TacYield = 0, 0
	c.DeployZones = []Zone{ZoneReserves}
	inst := s.give(0, c)
	s.do(Action{Kind: ActDeploy, Player: 0, Card: inst.ID, Zone: ZoneReserves})
}

// firstEventAfter returns the index in log of the first event of kind at or
// after from, or -1.
func firstEventAfter(log []Event, from int, kind string) int {
	for i := from; i < len(log); i++ {
		if log[i].Kind == kind {
			return i
		}
	}
	return -1
}

func destroyIndexFor(log []Event, from int, a *Asset) int {
	for i := from; i < len(log); i++ {
		e := log[i]
		if e.Kind == EventDestroy && (e.Source == a.ID || e.Target == a.ID || e.Card == a.Inst.Card.Name) {
			return i
		}
	}
	return -1
}

func TestFR_LAY_002_MaxHealthModifierChangesEffectiveMaxHealth(t *testing.T) {
	s := newScene(t)
	a := s.put(0, ZoneMainTheater, unit("Subject", 4, 5, Keywords{}))
	addMaxHealth(a, -3)
	if got := s.g.MaxHealth(a); got != 2 {
		t.Errorf("a -3 Maximum Health modifier on printed Health 5 gives Maximum Health 2 (FR-LAY-002), got %d", got)
	}
	if got := s.g.CurrentHealth(a); got != 2 {
		t.Errorf("Current Health is Maximum Health minus damage: 2 (FR-LAY-002), got %d", got)
	}
	s.wantDamage(a, 0)
	a.Damage = 1
	if got := s.g.CurrentHealth(a); got != 1 {
		t.Errorf("with 1 damage Current Health is 1 (FR-LAY-002), got %d", got)
	}
	a.Damage = 0
	addMaxHealth(a, 1)
	if got := s.g.MaxHealth(a); got != 3 {
		t.Errorf("Maximum Health modifiers add: 5 - 3 + 1 = 3 (FR-LAY-002), got %d", got)
	}
	if got := s.g.Power(a); got != 4 {
		t.Errorf("a Maximum Health modifier changes no other stat: power 4 (FR-LAY-002), got %d", got)
	}
	var found bool
	for _, as := range s.g.Snapshot().Theater {
		if as.ID == a.ID {
			found = true
			if as.MaxHealth != 3 {
				t.Errorf("the snapshot reports effective Maximum Health 3 (FR-LAY-002), got %d", as.MaxHealth)
			}
		}
	}
	if !found {
		t.Fatal("precondition: the asset is in the snapshot's Main Theater")
	}
}

func TestFR_LAY_001_MaxHealthDoesNotClamp(t *testing.T) {
	s := newScene(t)
	a := s.put(0, ZoneMainTheater, unit("Subject", 1, 2, Keywords{}))
	addMaxHealth(a, -3)
	if got := s.g.MaxHealth(a); got != -1 {
		t.Errorf("Maximum Health does not clamp (FR-LAY-001): want -1, got %d", got)
	}
	if got := s.g.CurrentHealth(a); got != -1 {
		t.Errorf("Current Health does not clamp (FR-LAY-001): want -1, got %d", got)
	}
}

func TestFR_LAY_002_MaxHealthAtZeroIsLethalWithoutDamage(t *testing.T) {
	s := newScene(t)
	dead := s.put(0, ZoneMainTheater, unit("Doomed", 1, 3, Keywords{}))
	addMaxHealth(dead, -3)
	spared := s.put(1, ZoneMainTheater, unit("Spared", 1, 3, Keywords{}))
	addMaxHealth(spared, -2)
	if !s.g.IsLethal(dead) {
		t.Errorf("Maximum Health 0 is lethal with no damage (FR-LAY-002)")
	}
	if s.g.IsLethal(spared) {
		t.Errorf("Maximum Health 1 with no damage is not lethal (FR-LAY-002)")
	}
	sweepNow(s)
	if s.alive(dead) {
		t.Errorf("a lethal asset is destroyed in the next death sweep (FR-LAY-002)")
	}
	if !s.alive(spared) {
		t.Errorf("an asset at Maximum Health 1 with no damage survives the sweep (FR-LAY-002)")
	}
}

func TestFR_LAY_002_MaxHealthReductionExpiresAndRestores(t *testing.T) {
	s := newScene(t)
	a := s.put(0, ZoneMainTheater, unit("Subject", 1, 5, Keywords{}))
	a.Damage = 2
	addMaxHealth(a, -2)
	if got := s.g.CurrentHealth(a); got != 1 {
		t.Fatalf("precondition: Current Health 5 - 2 - 2 = 1, got %d", got)
	}
	sweepNow(s)
	if !s.alive(a) {
		t.Fatal("an asset with Current Health 1 survives a sweep (FR-LAY-002)")
	}
	s.endTurn()
	if !s.alive(a) {
		t.Fatal("the reduction expires in Sub-Phase 4c and the asset survives (FR-LAY-002)")
	}
	if got := s.g.MaxHealth(a); got != 5 {
		t.Errorf("expiry restores Maximum Health to 5 (FR-LAY-002), got %d", got)
	}
	if got := s.g.CurrentHealth(a); got != 3 {
		t.Errorf("expiry restores Current Health to 3 (FR-LAY-002), got %d", got)
	}
	s.wantDamage(a, 2)
}

func testExpiryAtFourC(t *testing.T, p PlayerID) {
	s := newScene(t)
	a := s.put(p, ZoneMainTheater, unit("Bonus Holder", 1, 2, Keywords{}))
	addMaxHealth(a, 2)
	a.Damage = 3
	if !s.alive(a) || s.g.CurrentHealth(a) != 1 {
		t.Fatalf("precondition: alive with Current Health 1, got alive=%v health=%d", s.alive(a), s.g.CurrentHealth(a))
	}
	from := len(s.g.Log)
	s.endTurn()
	if s.alive(a) {
		t.Fatalf("the bonus expires in Sub-Phase 4c, leaving 3 damage on Health 2: destroyed in the End Phase sweep (FR-LAY-002, FR-TRN-001)")
	}
	d := destroyIndexFor(s.g.Log, from, a)
	ts := firstEventAfter(s.g.Log, from, EventTurnStart)
	if d < 0 {
		t.Fatalf("the destruction is logged as a destroy event (FR-LAY-002)")
	}
	if ts < 0 {
		t.Fatalf("the next turn begins after the End Phase, so a turn-start event is logged (FR-TRN-001)")
	}
	if d > ts {
		t.Errorf("the destroy event comes before the next turn's turn-start (FR-TRN-001 4c before 4d)")
	}
}

func TestFR_TRN_001_ExpiryAtFourCDestroysInTheEndPhase(t *testing.T) {
	// Covers: FR-LAY-002, FR-LAY-003
	t.Run("active player's unit", func(t *testing.T) { testExpiryAtFourC(t, 0) })
	t.Run("opponent's unit", func(t *testing.T) { testExpiryAtFourC(t, 1) })
}

func TestFR_TRN_001_FourCDeathTriggersResolveBeforeHandover(t *testing.T) {
	// Covers: FR-LAY-002, FR-TRG-004
	// The trigger is "Scratch 1": a successful draw is secret and logs no
	// event, but a Scratch logs an event and puts a face-down Scratch Company
	// on the battlefield.
	s := newScene(t)
	s.onTop(0, vanilla("Top A", "Top B")...)
	a := s.put(0, ZoneMainTheater, mortal("Mortal Holder", 1, 2, killedTrigger(scratchN(1))))
	addMaxHealth(a, 2)
	a.Damage = 3
	faceDown := faceDownCount(s.g.Players[0])
	from := len(s.g.Log)
	s.endTurn()
	if s.alive(a) {
		t.Fatalf("precondition: the unit is destroyed at 4c")
	}
	if got := faceDownCount(s.g.Players[0]); got != faceDown+1 {
		t.Errorf("its death trigger (Scratch 1) resolves in the End Phase: face-down Scratch Companies %d -> %d, want %d (FR-LAY-002, FR-TRN-001)", faceDown, got, faceDown+1)
	}
	if len(s.g.Stack) != 0 {
		t.Errorf("the stack is empty after the End Phase (FR-TRN-001), has %d", len(s.g.Stack))
	}
	ts := firstEventAfter(s.g.Log, from, EventTurnStart)
	if ts < 0 {
		t.Fatalf("the next turn begins after the End Phase, so a turn-start event is logged (FR-TRN-001)")
	}
	scratch := -1
	for i := from; i < len(s.g.Log); i++ {
		if s.g.Log[i].Kind == EventScratch && s.g.Log[i].Player == 0 {
			scratch = i
			break
		}
	}
	if scratch < 0 {
		t.Fatalf("the death trigger's Scratch is logged for player 0")
	}
	if scratch > ts {
		t.Errorf("the trigger resolves before the next turn's turn-start (FR-TRN-001 4c before 4d)")
	}
}

func TestFR_TRN_001_FourCSweepLossEndsTheMatch(t *testing.T) {
	// Covers: FR-LAY-002, FR-CBT-003
	s := newScene(t)
	h := s.g.Players[1].HQ
	addMaxHealth(h, 2)
	h.Damage = h.Printed().Health
	if s.g.Outcome.Over || !s.alive(h) {
		t.Fatalf("precondition: the match is on and the HQ is alive")
	}
	turn := s.g.Turn
	from := len(s.g.Log)
	s.endTurn()
	if !s.g.Outcome.Over || s.g.Outcome.Draw || s.g.Outcome.Winner != 0 {
		t.Errorf("player 1's HQ is lethal after the bonus expires at 4c: player 0 wins (FR-LAY-002), outcome %+v", s.g.Outcome)
	}
	if firstEventAfter(s.g.Log, from, EventTurnStart) >= 0 || s.g.Turn != turn {
		t.Errorf("a loss found by the 4c sweep ends the match before Sub-Phase 4d: no next turn begins (FR-TRN-001)")
	}
}

func TestFR_TRN_001_NoFourCDeathWithoutExpiry(t *testing.T) {
	s := newScene(t)
	a := s.put(0, ZoneMainTheater, unit("Scarred", 1, 4, Keywords{}))
	a.Damage = 3
	s.endTurn()
	if !s.alive(a) {
		t.Errorf("a damaged, non-lethal asset with no modifier survives the End Phase (FR-LAY-002)")
	}
	s.wantDamage(a, 3)
}
