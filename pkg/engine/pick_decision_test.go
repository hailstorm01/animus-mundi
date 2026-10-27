package engine

import (
	"reflect"
	"slices"
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0076: an untargeted pick is its own decision kind, DecidePick,
// not a Target (SVY-R3, SVY-R14).

func TestFR_TGT_001_UntargetedPickAsksAPickDecision(t *testing.T) {
	// Covers: SVY-R3, SVY-R14
	s := newScene(t)
	a := s.put(0, ZoneMainTheater, unit("A", 0, 5, Keywords{}))
	b := s.put(0, ZoneMainTheater, unit("B", 0, 5, Keywords{}))
	s.put(0, ZoneMainTheater, unit("C", 0, 5, Keywords{}))
	v1 := s.put(1, ZoneMainTheater, unit("Victim", 0, 5, Keywords{}))
	s.put(1, ZoneMainTheater, unit("Bystander", 0, 5, Keywords{}))
	targets := []ability.Target{{Filter: iydEnemy}, {Filter: iydAlly}}
	kill1 := ability.Step{Op: ability.OpKill, Subject: ability.SubjectTarget, Target: 1}
	s.scriptPlan(0, nil, nil, pick(v1), pick(a), pick(b))
	s.iydDeploy(iydTrigger(false, targets, kill1, iydPick(iydAlly, 0)))
	want := []DecisionKind{DecideTarget, DecideTarget, DecidePick}
	if got := kindsOf(s.asked(0)); !slices.Equal(got, want) {
		t.Fatalf("the chosen targets are asked as Targets on placement and the untargeted pick as a Pick on resolution: got %v, want %v (SVY-R3)", got, want)
	}
	var logged []string
	for _, e := range eventsOfKind(s.g.Log, "decision") {
		if e.Decision != nil && e.Player == 0 {
			logged = append(logged, e.Decision.Kind)
		}
	}
	if wantLog := []string{"Target", "Target", "Pick"}; !slices.Equal(logged, wantLog) {
		t.Errorf("the decision events record the target as \"Target\" and the pick as \"Pick\": got %v, want %v (SVY-R3)", logged, wantLog)
	}
}

func TestFR_MOD_001_PickDecisionKindIsNamedPick(t *testing.T) {
	for _, tc := range []struct {
		kind DecisionKind
		name string
	}{
		{DecidePick, "Pick"},
		{DecideTarget, "Target"},
		{DecideHand, "Hand"},
	} {
		if got := tc.kind.String(); got != tc.name {
			t.Errorf("decision kind %d is named %q, got %q", int(tc.kind), tc.name, got)
		}
	}
	if DecidePick == DecideTarget {
		t.Errorf("a pick is a different decision kind from a target (SVY-R3)")
	}
}

func TestFR_MOD_001_PickDecisionTimeoutMinOneTakesFirstOption(t *testing.T) {
	s := newScene(t)
	s.script(0, Answer{Timeout: true})
	d := Decision{Player: 0, Kind: DecidePick, Prompt: "pick", Options: opts(3), Min: 1, Max: 1}
	if got := s.g.choose(d); !reflect.DeepEqual(got, []int{0}) {
		t.Errorf("a timed-out mandatory pick takes the first legal option in fixed order: got %v, want [0] (FR-MOD-001 §6)", got)
	}
}

func TestFR_MOD_001_PickDecisionTimeoutMinZeroTakesNone(t *testing.T) {
	s := newScene(t)
	s.script(0, Answer{Timeout: true})
	d := Decision{Player: 0, Kind: DecidePick, Prompt: "pick", Options: opts(3), Min: 0, Max: 2}
	if got := s.g.choose(d); len(got) != 0 {
		t.Errorf("a timed-out optional pick (Min 0) selects none: got %v (FR-MOD-001 §6)", got)
	}
}
