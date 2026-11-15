package engine

import (
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for a trigger that watches another asset's attack and for the
// "that attacking unit" step subject (T-0241; the shape Blind Defile needs).
// FR-CBT-002 Step 2, FR-TRG-003, FR-TGT-001 (Untargeted Selections),
// FR-ZON-004, FR-MOD-001 section 7.
//
// This file does not compile until ability.Trigger has Watches (bool) and
// Watch (ability.Filter), ability.SubjectWatched exists and StackItem has a
// Watched (int) field; that is expected until T-0241 is implemented.

// opposingUnits is the watch filter "an opposing unit".
var opposingUnits = ability.Filter{Allegiance: ability.AllegianceOpposing, Kinds: ability.KindUnit}

// watchTrigger is one "Whenever an opposing unit attacks, <steps>" ability.
func watchTrigger(optional bool, steps ...ability.Step) ability.Trigger {
	return ability.Trigger{Hook: ability.HookAttacks, Watches: true, Watch: opposingUnits, Optional: optional,
		Effect: ability.Effect{Steps: steps}}
}

func killWatched() ability.Step {
	return ability.Step{Op: ability.OpKill, Subject: ability.SubjectWatched}
}

// watcherStructure is a structure printing the given triggers.
func watcherStructure(name string, trigs ...ability.Trigger) *Card {
	c := structure(name, 0, 5, Keywords{})
	c.Ability = &ability.Def{Triggers: trigs}
	return c
}

// --- FR-CBT-002 Step 2: a watcher fires on another asset's attack ---

func TestFR_CBT_002_WatcherFiresOnOpposingUnitAttack(t *testing.T) {
	s := newScene(t)
	w := s.put(1, ZoneMainTheater, watcherStructure("Watcher", watchTrigger(false, killWatched())))
	att := s.put(0, ZoneMainTheater, unit("Hitter", 3, 9, Keywords{}))
	items := recordPlaced(s)
	s.do(s.attack(att, s.g.Players[1].HQ))
	if len(*items) != 2 || (*items)[0].Kind != StackAttack {
		t.Fatalf("an opposing unit's attack places the Attack and the watcher's trigger: placed %+v (FR-CBT-002 Step 2)", *items)
	}
	it := (*items)[1]
	if it.Kind != StackAbility || it.Controller != 1 || it.Source != w.ID || it.Watched != att.ID {
		t.Errorf("the watcher's trigger is controlled by the watcher's controller (1), has the watcher as its source and the attacker as its watched asset: %+v", it)
	}
}

func TestFR_CBT_002_WatcherIgnoresUnmatchedAttacks(t *testing.T) {
	t.Run("AlliedUnit", func(t *testing.T) {
		s := newScene(t)
		s.put(1, ZoneMainTheater, watcherStructure("Watcher", watchTrigger(false, killWatched())))
		s.endTurn()
		att := s.put(1, ZoneMainTheater, unit("Ally", 3, 9, Keywords{}))
		items := recordPlaced(s)
		s.do(s.attack(att, s.g.Players[0].HQ))
		if len(*items) != 1 {
			t.Errorf("a unit allied to the watcher attacking does not fire an opposing-unit watcher (FR-CBT-002): placed %+v", *items)
		}
		if !s.alive(att) {
			t.Errorf("the allied attacker was not killed")
		}
	})
	t.Run("OpposingHeadquarters", func(t *testing.T) {
		s := newScene(t)
		s.put(1, ZoneMainTheater, watcherStructure("Watcher", watchTrigger(false, killWatched())))
		def := s.put(1, ZoneMainTheater, unit("Dummy", 0, 20, Keywords{}))
		items := recordPlaced(s)
		s.do(s.attack(s.g.Players[0].HQ, def))
		if len(*items) != 1 {
			t.Errorf("a Headquarters attacking does not fire a watcher of opposing units only (FR-CBT-002): placed %+v", *items)
		}
	})
}

func TestFR_CBT_002_WatcherIsNotASelfHook(t *testing.T) {
	// The reverse of TestFR_CBT_002_OnlyTheAttackerFiresItsAttackTrigger: a
	// watching trigger does not fire as its own source's attack trigger.
	s := newScene(t)
	cards := s.onTop(0, vanilla("A")...)
	att := s.put(0, ZoneMainTheater, striker("Hitter", 3, 9, Keywords{}, watchTrigger(false, drawN(1))))
	items := recordPlaced(s)
	s.do(s.attack(att, s.g.Players[1].HQ))
	if len(*items) != 1 || inHand(s.g.Players[0], cards[0]) {
		t.Errorf("a unit's watching trigger for opposing units does not fire on its own attack (FR-CBT-002): placed %+v", *items)
	}
}

// --- FR-TRG-003: ordering with the attacker's own triggers ---

func TestFR_TRG_003_OpposingWatcherResolvesBeforeAttackerTriggers(t *testing.T) {
	s := newScene(t)
	w := s.put(1, ZoneMainTheater, watcherStructure("Watcher", watchTrigger(false, drawN(1))))
	att := s.put(0, ZoneMainTheater, unit("Raider", 3, 9, Keywords{Aggressive: 1}))
	items := recordPlaced(s)
	s.do(s.attack(att, s.g.Players[1].HQ))
	if len(*items) != 3 || (*items)[0].Kind != StackAttack || (*items)[1].Keyword != ability.KeywordAggressive || (*items)[2].Source != w.ID {
		t.Fatalf("the non-active player's watcher is placed after, so above, the attacker's triggers and resolves first (FR-TRG-003): placed %+v", *items)
	}
}

func TestFR_TRG_003_TwoWatchersOneAttack(t *testing.T) {
	s := newScene(t)
	s.script(0)
	s.script(1, Answer{Timeout: true})
	s.put(1, ZoneMainTheater, watcherStructure("Watcher A", watchTrigger(false, drawN(1))))
	s.put(1, ZoneMainTheater, watcherStructure("Watcher B", watchTrigger(false, drawN(1))))
	att := s.put(0, ZoneMainTheater, unit("Hitter", 3, 9, Keywords{}))
	items := recordPlaced(s)
	s.do(s.attack(att, s.g.Players[1].HQ))
	if len(*items) != 3 {
		t.Errorf("one attack fires both watchers: placed %d items, want 3 (Attack and two triggers)", len(*items))
	}
	asked := s.asked(1)
	if len(asked) != 1 || asked[0].Kind != DecideOrder || asked[0].Player != 1 || len(asked[0].Options) != 2 {
		t.Errorf("two simultaneous triggers are ordered by their controller with one DecideOrder of 2 options (FR-TRG-003): %+v", asked)
	}
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("the attacker's controller has no triggers and is not asked: asked %d times", n)
	}
}

// --- FR-TGT-001: "that attacking unit" is an untargeted step subject ---

func TestFR_TGT_001_WatchedSubjectKillsTheAttacker(t *testing.T) {
	s := newScene(t)
	s.put(1, ZoneMainTheater, watcherStructure("Watcher", watchTrigger(false, killWatched())))
	att := s.put(0, ZoneMainTheater, unit("Hitter", 3, 9, Keywords{}))
	s.do(s.attack(att, s.g.Players[1].HQ))
	if s.alive(att) || !inDiscard(s.g.Players[0], att.Inst) {
		t.Errorf("killing the watched asset puts the attacker into its owner's discard pile")
	}
	s.wantDamage(s.g.Players[1].HQ, 0)
}

func TestFR_TGT_001_WatchedSubjectIsNotATarget(t *testing.T) {
	// Covers: FR-TGT-003
	s := newScene(t)
	s.script(0)
	s.script(1)
	s.put(1, ZoneMainTheater, watcherStructure("Watcher", watchTrigger(false, killWatched())))
	att := s.put(0, ZoneMainTheater, unit("Warded", 3, 9, Keywords{Ward: true}))
	s.do(s.attack(att, s.g.Players[1].HQ))
	if s.alive(att) {
		t.Errorf("a clause without \"chosen\" does not target, so Ward does not stop it (FR-TGT-001 Untargeted Selections)")
	}
	if n := len(s.asked(0)) + len(s.asked(1)); n != 0 {
		t.Errorf("an untargeted watched subject asks nothing of either player: asked %d times", n)
	}
}

func TestFR_TGT_001_WatchedSubjectGoneDoesNothing(t *testing.T) {
	// Covers: FR-ZON-004
	s := newScene(t)
	cards := s.onTop(1, vanilla("A")...)
	ifDoneDraw := drawN(1)
	ifDoneDraw.IfDone = true
	s.put(1, ZoneMainTheater, watcherStructure("Watcher", watchTrigger(false, killWatched(), ifDoneDraw)))
	att := s.put(0, ZoneMainTheater, unit("Hitter", 3, 9, Keywords{}))
	s.declareThen(s.attack(att, s.g.Players[1].HQ), func() {
		s.g.flushTriggers()
		if n := len(s.g.Stack); n != 2 {
			t.Fatalf("the Attack and the watcher's trigger are on the stack: %d items", n)
		}
		s.g.leaveBattlefield(att, s.g.effectiveState(att))
	})
	if inHand(s.g.Players[1], cards[0]) {
		t.Errorf("the watched asset left the battlefield, so the kill did nothing and the IfDone step did not run (FR-ZON-004)")
	}
}

// --- FR-MOD-001 section 7: "you may" ---

func TestFR_MOD_001_OptionalWatcherIsAsked(t *testing.T) {
	// Covers: SVY-R9, SVY-R48
	s := newScene(t)
	s.script(0)
	s.script(1, no)
	s.put(1, ZoneMainTheater, watcherStructure("Watcher", watchTrigger(true, killWatched())))
	att := s.put(0, ZoneMainTheater, unit("Hitter", 3, 9, Keywords{}))
	items := recordPlaced(s)
	s.do(s.attack(att, s.g.Players[1].HQ))
	if got := kindsOf(s.asked(1)); len(got) != 1 || got[0] != DecideOptional {
		t.Errorf("the watcher's controller is asked \"you may\" once when it is placed, even though its step is not a target (SVY-R9, SVY-R48): asked %v", got)
	}
	if len(*items) != 1 {
		t.Errorf("a declined optional trigger is not placed: placed %+v", *items)
	}
	if !s.alive(att) {
		t.Errorf("declined, the attacker is not killed")
	}
	s.wantDamage(s.g.Players[1].HQ, 3)
}
