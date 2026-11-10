package engine

import (
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for the moves-zone hook (T-0203): a printed trigger that fires when its
// own unit moves to a new battlefield zone, by March or by a step that moves
// it (OpMoveZone). FR-ZON-002 and SVY-R1 (FRD 01, "What Counts as Moving"):
// entering the battlefield, by any deploy, is not moving. FR-TRG-002 makes it a
// self hook. A trigger fired while an item resolves is placed after that item's
// sweep (FR-TRG-003), and a unit moved by a resolving move trigger fires its
// own (FR-TRG-004).
//
// Cards are synthetic: a move trigger that Scratches 1, observed with
// wantScratches and by the order of events in the game log.
//
// This file does not compile until ability has a HookMoves constant; that is
// expected until T-0203 is implemented.

// moveTrigger is one "Whenever this moves [to zone], <steps>" ability.
func moveTrigger(to ability.Zone, steps ...ability.Step) ability.Trigger {
	return ability.Trigger{Hook: ability.HookMoves, ToZone: to, Effect: ability.Effect{Steps: steps}}
}

// moveWatcher is a unit printing a "whenever this moves, Scratch 1" trigger.
func moveWatcher(name string) *Card {
	return striker(name, 1, 5, Keywords{}, moveTrigger(ability.ZoneAny, scratchN(1)))
}

// selfDamage is "this deals 1 damage to itself", a step observed through the
// "damage" log event, so the order of a trigger's resolution can be told apart
// from a Scratch.
func selfDamage() ability.Step {
	return ability.Step{Op: ability.OpDealDamage, Subject: ability.SubjectSelf, Amount: ability.Value{N: 1}}
}

// moveTargetFree has player 0 enact a free Maneuver that moves the allied unit
// w to zone z, then runs the extra steps.
func (s *scene) moveTargetFree(w *Asset, z ability.Zone, extra ...ability.Step) {
	s.t.Helper()
	s.choosesAssets(0, w.ID)
	steps := append([]ability.Step{moveStep(ability.SubjectTarget, 0, ability.Filter{}, z)}, extra...)
	s.enactFree(ability.Effect{
		Targets: []ability.Target{{Filter: mzAlliedUnits}},
		Steps:   steps,
	})
}

func TestFR_ZON_002_MarchFiresMoveTrigger(t *testing.T) {
	// Covers: FR-CAP-001
	s := newScene(t)
	w := s.put(0, ZoneReserves, moveWatcher("Watcher"))
	s.do(Action{Kind: ActMarch, Player: 0, Asset: w.ID})
	if got := s.g.Asset(w.ID); got == nil || got.Zone != ZoneMainTheater {
		t.Fatalf("precondition: the March moves the watcher to the Main Theater, got %v", got)
	}
	s.wantScratches(0, 1, "a March moves the unit to a new zone, which fires its move trigger once (FR-ZON-002)")
	mv, sc := logIndex(s.g, "march"), logIndex(s.g, "scratch")
	if mv < 0 || sc < 0 {
		t.Fatalf("expected a march event and a scratch event in the log, got march at %d, scratch at %d", mv, sc)
	}
	if sc < mv {
		t.Errorf("the unit moves when the March resolves, so its move trigger resolves after the March (FR-ZON-002, FR-CAP-001): scratch event %d came before march event %d", sc, mv)
	}
	s.wantSettled("a March with a move trigger")

	s.endTurn()
	s.endTurn()
	s.do(Action{Kind: ActMarch, Player: 0, Asset: w.ID})
	if got := s.g.Asset(w.ID); got == nil || got.Zone != ZoneReserves {
		t.Fatalf("precondition: the second March moves the watcher back to Reserves, got %v", got)
	}
	s.wantScratches(0, 2, "marching back to Reserves is a move to a new zone and fires the trigger again (FR-ZON-002)")
	s.wantSettled("a March back with a move trigger")
}

func TestFR_ZON_002_MoveStepFiresMoveTrigger(t *testing.T) {
	s := newScene(t)
	w := s.put(0, ZoneReserves, striker("Watcher", 1, 5, Keywords{}, moveTrigger(ability.ZoneAny, selfDamage())))
	s.moveTargetFree(w, ability.ZoneMainTheater, scratchN(1))
	if got := s.g.Asset(w.ID); got == nil || got.Zone != ZoneMainTheater {
		t.Fatalf("precondition: the move step moves the watcher to the Main Theater, got %v", got)
	}
	s.wantDamage(w, 1)
	sc, dm := logIndex(s.g, "scratch"), logIndex(s.g, "damage")
	if sc < 0 || dm < 0 {
		t.Fatalf("expected a scratch event (the Maneuver's last step) and a damage event (the trigger) in the log, got scratch at %d, damage at %d", sc, dm)
	}
	if dm < sc {
		t.Errorf("a move trigger fired during resolution is placed after the item's sweep, so it resolves after the Maneuver's later steps (FR-ZON-002, FR-TRG-003): damage event %d came before scratch event %d", dm, sc)
	}
	s.wantSettled("a move step with a move trigger")
}

func TestFR_ZON_002_DeployIsNotMoving(t *testing.T) {
	// Covers: SVY-R1
	s := newScene(t)
	s.deployFromHand(0, moveWatcher("Reserve Recruit"))
	s.wantScratches(0, 0, "deploying to Reserves is entering the battlefield, not moving, so no move trigger fires (SVY-R1, FR-ZON-002)")

	inst := s.give(0, moveWatcher("Front Recruit"))
	s.do(Action{Kind: ActDeploy, Player: 0, Card: inst.ID, Zone: ZoneMainTheater})
	s.wantScratches(0, 0, "deploying to the Main Theater is entering the battlefield, not moving, so no move trigger fires (SVY-R1, FR-ZON-002)")
	s.wantSettled("two deploys of units with move triggers")
}

func TestFR_ZON_002_MovesToZoneGate(t *testing.T) {
	// Covers: SVY-R1
	s := newScene(t)
	w := s.put(0, ZoneReserves, striker("Watcher", 1, 5, Keywords{}, moveTrigger(ability.ZoneMainTheater, scratchN(1))))
	s.do(Action{Kind: ActMarch, Player: 0, Asset: w.ID})
	s.wantScratches(0, 1, "a trigger that watches for moving to the Main Theater fires when its unit marches into it (FR-ZON-002)")

	s.endTurn()
	s.endTurn()
	s.do(Action{Kind: ActMarch, Player: 0, Asset: w.ID})
	if got := s.g.Asset(w.ID); got == nil || got.Zone != ZoneReserves {
		t.Fatalf("precondition: the second March moves the watcher back to Reserves, got %v", got)
	}
	s.wantScratches(0, 1, "marching to Reserves is not moving to the Main Theater, so the gated trigger does not fire (FR-ZON-002)")
	s.wantSettled("a gated move trigger over two Marches")
}

func TestFR_ZON_002_NoMoveNoTrigger(t *testing.T) {
	s := newScene(t)
	front := s.put(0, ZoneMainTheater, moveWatcher("Front Watcher"))
	s.moveTargetFree(front, ability.ZoneMainTheater)
	s.wantScratches(0, 0, "a move step to the zone the unit is already in moves nothing, so it fires no move trigger (FR-ZON-002)")

	back := s.put(0, ZoneReserves, moveWatcher("Back Watcher"))
	s.moveTargetFree(back, ability.ZoneReserves)
	s.wantScratches(0, 0, "a move step to Reserves for a unit already in Reserves fires no move trigger (FR-ZON-002)")
	s.wantSettled("two move steps that move nothing")
}

func TestFR_ZON_002_OnlyTheMovedUnitFires(t *testing.T) {
	// Covers: FR-TRG-002
	s := newScene(t)
	mover := s.put(0, ZoneReserves, moveWatcher("Mover"))
	s.put(0, ZoneReserves, moveWatcher("Bystander"))
	s.put(1, ZoneReserves, moveWatcher("Opposing Bystander"))
	s.do(Action{Kind: ActMarch, Player: 0, Asset: mover.ID})
	s.wantScratches(0, 1, "the move hook is a self hook: only the unit that moved fires its trigger, not another watcher in the same zone (FR-TRG-002)")
	s.wantScratches(1, 0, "an opposing watcher does not fire when another player's unit moves (FR-TRG-002)")
}

func TestFR_TRG_004_MoveTriggerCascades(t *testing.T) {
	// Covers: FR-ZON-002, FR-TRG-003
	s := newScene(t)
	b := s.put(0, ZoneReserves, moveWatcher("Second"))
	reserveUnits := ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit, Zone: ability.ZoneReserves}
	a := s.put(0, ZoneReserves, striker("First", 1, 5, Keywords{}, ability.Trigger{
		Hook:   ability.HookMoves,
		ToZone: ability.ZoneMainTheater,
		Effect: ability.Effect{
			Targets: []ability.Target{{Filter: reserveUnits}},
			Steps:   []ability.Step{moveStep(ability.SubjectTarget, 0, ability.Filter{}, ability.ZoneMainTheater), selfDamage()},
		},
	}))
	s.choosesAssets(0, b.ID)
	s.do(Action{Kind: ActMarch, Player: 0, Asset: a.ID})
	if got := s.g.Asset(b.ID); got == nil || got.Zone != ZoneMainTheater {
		t.Fatalf("precondition: the first unit's move trigger moves the second to the Main Theater, got %v", got)
	}
	s.wantScratches(0, 1, "the unit moved by a resolving move trigger fires its own move trigger (FR-TRG-004)")
	dm, sc := logIndex(s.g, "damage"), logIndex(s.g, "scratch")
	if dm < 0 || sc < 0 {
		t.Fatalf("expected a damage event (first trigger's last step) and a scratch event (second trigger) in the log, got damage at %d, scratch at %d", dm, sc)
	}
	if sc < dm {
		t.Errorf("the second unit's trigger is placed after the first trigger's sweep, so it resolves after the first trigger finishes (FR-TRG-003, FR-TRG-004): scratch event %d came before damage event %d", sc, dm)
	}
	s.wantSettled("a cascade of two move triggers")
}
