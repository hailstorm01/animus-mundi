package engine

import (
	"strings"
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0249: a temporary effect that lasts "until the end of your
// opponent's turn" (FR-LAY-003 Full-Round Durations, FR-TRN-001 Sub-Phase 4c,
// SVY-R39). Step.Until names how long a modifier made by OpModifyStat or
// OpGrantKeyword lasts; the zero value is until end of turn. Cards are
// synthetic. Player 0 is the active player.

// powerFor is "chosen unit gets n power" lasting until.
func powerFor(n int, until ability.Duration) ability.Step {
	return ability.Step{Op: ability.OpModifyStat, Amount: ability.Value{N: n}, Subject: ability.SubjectTarget, Target: 0, Until: until}
}

// enactOn has player 0 enact a free Maneuver whose one chosen unit of the given
// allegiance gets step. Each scene holds exactly one such unit, so the choice is forced.
func enactOn(s *scene, a ability.Allegiance, step ability.Step) {
	s.t.Helper()
	inst := s.give(0, maneuver("Temporary", 0, 0, ability.Effect{
		Targets: []ability.Target{tgt(a, ability.KindUnit, false)},
		Steps:   []ability.Step{step},
	}))
	s.do(enactAction(inst))
}

func wantPower(t *testing.T, s *scene, a *Asset, want int, when string) {
	t.Helper()
	if got := s.g.Power(a); got != want {
		t.Errorf("%s: power %d, want %d (FR-LAY-003)", when, got, want)
	}
}

func TestFR_LAY_003_OpponentTurnPowerLastsThroughOpponentsTurn(t *testing.T) {
	// Covers: FR-TRN-001, FR-LAY-001
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Foe", 4, 9, ability.Keywords{}))
	enactOn(s, ability.AllegianceOpposing, powerFor(-3, ability.UntilEndOfOpponentTurn))
	wantPower(t, s, foe, 1, "right after the Maneuver")
	s.endTurn()
	if s.g.Active != 1 || s.g.Phase != PhaseMain {
		t.Fatalf("precondition: player 1's Main Phase, got active %d phase %v", s.g.Active, s.g.Phase)
	}
	wantPower(t, s, foe, 1, "through the opponent's turn (it did not expire in player 0's Sub-Phase 4c)")
	s.endTurn()
	wantPower(t, s, foe, 4, "after the opponent's turn ended (expires in its Sub-Phase 4c)")
	s.endTurn()
	wantPower(t, s, foe, 4, "after player 0's next turn ended")
}

func TestFR_LAY_003_OpponentTurnPowerOnAnAlliedUnit(t *testing.T) {
	s := newScene(t)
	ally := s.put(0, ZoneMainTheater, unit("Ally", 1, 9, ability.Keywords{}))
	enactOn(s, ability.AllegianceAllied, powerFor(2, ability.UntilEndOfOpponentTurn))
	wantPower(t, s, ally, 3, "right after the Maneuver")
	s.endTurn()
	wantPower(t, s, ally, 3, "through player 1's turn")
	s.endTurn()
	wantPower(t, s, ally, 1, "after player 1's turn ended")
}

func TestFR_LAY_003_OpponentTurnIsTheAbilityControllersOpponent(t *testing.T) {
	// "Your opponent" is the ability controller's opponent, not the asset
	// controller's: the asset here is player 1's, the ability player 0's.
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Foe", 3, 9, ability.Keywords{}))
	enactOn(s, ability.AllegianceOpposing, powerFor(2, ability.UntilEndOfOpponentTurn))
	s.endTurn() // player 0's turn ends: not the end of the ability controller's opponent's turn
	wantPower(t, s, foe, 5, "after the ability controller's own turn ended")
	s.endTurn() // player 1's turn ends
	wantPower(t, s, foe, 3, "after the ability controller's opponent's turn ended")
	s.endTurn()
	wantPower(t, s, foe, 3, "after player 0's next turn ended")
}

func TestFR_TRN_001_OpponentTurnModifierMadeOffTurnExpiresThatTurn(t *testing.T) {
	// Covers: FR-LAY-003
	// 4c: "until the end of your opponent's turn" when the current turn is that
	// opponent's turn. A modifier made in the opponent's own turn ends with it.
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Foe", 4, 9, ability.Keywords{}))
	weaken := ability.Step{Op: ability.OpModifyStat, Amount: ability.Value{N: -1}, Subject: ability.SubjectAll,
		Filter: ability.Filter{Allegiance: ability.AllegianceOpposing, Kinds: ability.KindUnit},
		Until:  ability.UntilEndOfOpponentTurn}
	s.put(0, ZoneMainTheater, startStructure("Chill", startSteps(ability.TurnAny, weaken)))
	s.endTurn()
	if s.g.Active != 1 || s.g.Phase != PhaseMain {
		t.Fatalf("precondition: player 1's Main Phase, got active %d phase %v", s.g.Active, s.g.Phase)
	}
	wantPower(t, s, foe, 3, "during player 1's Main Phase, after the trigger resolved at their turn start")
	s.endTurn()
	wantPower(t, s, foe, 3, "in player 0's turn: the first modifier expired in player 1's 4c, the second was just made, not -2")
	if n := len(foe.Modifiers); n != 1 {
		t.Errorf("the enemy unit has %d modifiers, want exactly 1 (FR-TRN-001 4c)", n)
	}
}

func TestFR_LAY_003_OpponentTurnKeywordGrant(t *testing.T) {
	// Covers: SVY-R39
	s := newScene(t)
	ally := s.put(0, ZoneMainTheater, unit("Ally", 1, 9, ability.Keywords{}))
	enactOn(s, ability.AllegianceAllied, ability.Step{Op: ability.OpGrantKeyword, Subject: ability.SubjectTarget, Target: 0,
		Keywords: ability.Keywords{Stalwart: true}, Until: ability.UntilEndOfOpponentTurn})
	if !s.g.Keywords(ally).Stalwart {
		t.Errorf("the unit has Stalwart right after the grant")
	}
	s.endTurn()
	if !s.g.Keywords(ally).Stalwart {
		t.Errorf("the unit still has Stalwart during the opponent's turn (FR-LAY-003, SVY-R39)")
	}
	s.endTurn()
	if s.g.Keywords(ally).Stalwart {
		t.Errorf("the unit loses Stalwart when the opponent's turn ends (FR-LAY-003, SVY-R39)")
	}
}

func TestFR_LAY_003_UntilEndOfTurnUnchanged(t *testing.T) {
	if ability.Duration(0) != ability.UntilEndOfTurn {
		t.Errorf("the zero Duration is UntilEndOfTurn (FR-LAY-003)")
	}
	s := newScene(t)
	ally := s.put(0, ZoneMainTheater, unit("Ally", 1, 9, ability.Keywords{}))
	enactOn(s, ability.AllegianceAllied, ability.Step{Op: ability.OpModifyStat, Amount: ability.Value{N: 2}, Subject: ability.SubjectTarget, Target: 0})
	enactOn(s, ability.AllegianceAllied, ability.Step{Op: ability.OpGrantKeyword, Subject: ability.SubjectTarget, Target: 0, Keywords: ability.Keywords{Stalwart: true}})
	wantPower(t, s, ally, 3, "during the turn it was made")
	if !s.g.Keywords(ally).Stalwart {
		t.Errorf("Stalwart is granted during the turn it was made")
	}
	s.endTurn()
	wantPower(t, s, ally, 1, "after the turn it was made in ended (no Until: until end of turn)")
	if s.g.Keywords(ally).Stalwart {
		t.Errorf("a grant with no Until expires at the end of the turn it was made in (FR-TRN-001 4c)")
	}
}

func TestFR_LAY_003_MixedDurationsExpireSeparately(t *testing.T) {
	s := newScene(t)
	ally := s.put(0, ZoneMainTheater, unit("Ally", 1, 9, ability.Keywords{}))
	enactOn(s, ability.AllegianceAllied, powerFor(1, ability.UntilEndOfTurn))
	enactOn(s, ability.AllegianceAllied, powerFor(2, ability.UntilEndOfOpponentTurn))
	wantPower(t, s, ally, 4, "right after both Maneuvers")
	s.endTurn()
	wantPower(t, s, ally, 3, "after player 0's turn ended: only the until-end-of-turn +1 is gone")
	s.endTurn()
	wantPower(t, s, ally, 1, "after player 1's turn ended")
}

func TestFR_LAY_003_OpponentTurnModifierLeavesWithTheAsset(t *testing.T) {
	// Covers: FR-ZON-004
	s := newScene(t)
	card := unit("Foe", 4, 9, ability.Keywords{})
	foe := s.put(1, ZoneMainTheater, card)
	enactOn(s, ability.AllegianceOpposing, powerFor(-3, ability.UntilEndOfOpponentTurn))
	enactOn(s, ability.AllegianceOpposing, ability.Step{Op: ability.OpReturnToHand, Subject: ability.SubjectTarget, Target: 0})
	if s.alive(foe) {
		t.Fatalf("precondition: the unit returned to its owner's hand")
	}
	s.endTurn()
	var inst *Instance
	for _, i := range s.g.Players[1].Hand {
		if i.Card == card {
			inst = i
		}
	}
	if inst == nil {
		t.Fatalf("precondition: the returned card is in player 1's hand")
	}
	s.do(Action{Kind: ActDeploy, Player: 1, Card: inst.ID, Zone: ZoneMainTheater})
	back := assetOf(s.g, card)
	if back == nil {
		t.Fatalf("precondition: the card redeployed")
	}
	wantPower(t, s, back, 4, "the redeployed unit has printed power: leaving the battlefield cleared the modifier (FR-ZON-004)")
	if len(back.Modifiers) != 0 {
		t.Errorf("the redeployed unit has no modifiers, got %+v (FR-ZON-004)", back.Modifiers)
	}
}

func TestFR_LAY_003_OpponentTurnLogText(t *testing.T) {
	logLine := func(until ability.Duration) string {
		s := newScene(t)
		s.put(1, ZoneMainTheater, unit("Foe", 4, 9, ability.Keywords{}))
		from := len(s.g.Log)
		enactOn(s, ability.AllegianceOpposing, powerFor(-1, until))
		evs := effectEvents(s.g, from, EventPower)
		if len(evs) != 1 {
			t.Fatalf("%d power events, want 1", len(evs))
		}
		return evs[0].Detail
	}
	if got := logLine(ability.UntilEndOfOpponentTurn); !strings.HasSuffix(got, "until the end of its controller's opponent's turn") {
		t.Errorf("the power event line is %q, want it to end in \"until the end of its controller's opponent's turn\"", got)
	}
	if got := logLine(ability.UntilEndOfTurn); !strings.HasSuffix(got, "until end of turn") {
		t.Errorf("the power event line is %q, want it to end in \"until end of turn\"", got)
	}
}
