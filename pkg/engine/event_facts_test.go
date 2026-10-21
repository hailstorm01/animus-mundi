package engine

import (
	"fmt"
	"strings"
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0026: damage-stage and stack-placement event facts. Each
// TestEventFacts_Case<N> is Acceptance Case N (the 20-game sweep of case 2 is
// in package sim, which can run SimulateWith).

// damageEvents returns the damage events of log, in order.
func damageEvents(log []Event) []Event { return eventsOfKind(log, "damage") }

// Case 1: damage events carry the mitigation facts.
func TestEventFacts_Case1_DamageFields(t *testing.T) {
	// Power 3 into a Stalwart defender with Armor 2.
	s := newScene(t)
	att := s.put(0, ZoneMainTheater, unit("Att", 3, 9, Keywords{}))
	def := s.put(1, ZoneMainTheater, unit("Wall", 2, 9, Keywords{Stalwart: true, Armor: 2}))
	s.do(s.attack(att, def))
	dmg := damageEvents(s.g.Log)
	if len(dmg) != 2 {
		t.Fatalf("got %d damage events, want 2", len(dmg))
	}
	d := dmg[0]
	if d.Source != att.ID || d.Target != def.ID || d.Raw != 3 || !d.Combat || d.Armor != 2 || d.Absorbed != 1 || d.Amount != 0 {
		t.Errorf("attacker's damage = %+v, want raw 3, combat, armor 2, absorbed 1, amount 0", d)
	}
	if r := dmg[1]; r.Source != def.ID || r.Target != att.ID || r.Raw != 2 || !r.Combat || r.Armor != 0 || r.Absorbed != 0 || r.Amount != 2 {
		t.Errorf("retaliation = %+v, want raw 2 (the defender's power), combat, amount 2", r)
	}

	// The combat event names both combatants and keeps its Detail.
	combat := eventsOfKind(s.g.Log, "combat")
	if len(combat) != 1 {
		t.Fatalf("got %d combat events, want 1", len(combat))
	}
	if c := combat[0]; c.Source != att.ID || c.Target != def.ID {
		t.Errorf("combat event Source/Target = %d/%d, want %d/%d", c.Source, c.Target, att.ID, def.ID)
	} else if want := fmt.Sprintf("%s (3) and %s (2) exchange combat damage", att, def); c.Detail != want {
		t.Errorf("combat Detail = %q, want %q", c.Detail, want)
	}

	// Resistant -1 amplifies 0 raw damage to 1.
	s = newScene(t)
	zero := s.put(0, ZoneMainTheater, unit("Zero", 0, 9, Keywords{}))
	soft := s.put(1, ZoneMainTheater, unit("Soft", 0, 9, Keywords{Resistant: -1}))
	s.do(s.attack(zero, soft))
	if d := damageEvents(s.g.Log)[0]; d.Target != soft.ID || d.Raw != 0 || !d.Combat || d.Amount != 1 {
		t.Errorf("amplified damage = %+v, want raw 0, combat, amount 1", d)
	}

	// Non-combat damage on a Stalwart target is not combat and not capped.
	s = newScene(t)
	wall := s.put(1, ZoneMainTheater, unit("Wall", 1, 9, Keywords{Stalwart: true}))
	src := s.deployTargeter(0, []ability.Target{tgt(ability.AllegianceOpposing, ability.KindUnit, false)}, dmgTo(3, 0))
	dmg = damageEvents(s.g.Log)
	if len(dmg) != 1 {
		t.Fatalf("got %d damage events, want 1", len(dmg))
	}
	if d := dmg[0]; d.Source != src.ID || d.Target != wall.ID || d.Raw != 3 || d.Combat || d.Amount != 3 {
		t.Errorf("ability damage = %+v, want raw 3, not combat, amount 3", d)
	}
}

// indexOfKind is the index of the first event of kind in log, -1 if none.
func indexOfKind(log []Event, kind string) int {
	for i, e := range log {
		if e.Kind == kind {
			return i
		}
	}
	return -1
}

func checkPlaceDetail(t *testing.T, e Event) {
	t.Helper()
	if want := fmt.Sprintf("player %d places ", e.Player); !strings.HasPrefix(e.Detail, want) || !strings.HasSuffix(e.Detail, " on the stack") {
		t.Errorf("place Detail = %q, want %q... on the stack", e.Detail, want)
	}
}

// Case 2: place events.
func TestEventFacts_Case2_PlaceEvents(t *testing.T) {
	// A Deploy: action, place, deploy.
	s := newScene(t)
	c := unit("Newcomer", 1, 1, Keywords{})
	inst := s.give(0, c)
	s.do(Action{Kind: ActDeploy, Player: 0, Card: inst.ID, Zone: ZoneReserves})
	i := indexOfKind(s.g.Log, "action")
	if i < 0 || len(s.g.Log) < i+3 {
		t.Fatalf("log too short: %d events", len(s.g.Log))
	}
	a, p, d := s.g.Log[i], s.g.Log[i+1], s.g.Log[i+2]
	if a.Kind != "action" || p.Kind != "place" || d.Kind != "deploy" {
		t.Fatalf("kinds = %s, %s, %s; want action, place, deploy", a.Kind, p.Kind, d.Kind)
	}
	if p.Player != 0 || p.Card != "Newcomer" || p.Inst != inst.ID || p.Amount != 1 {
		t.Errorf("deploy place = %+v, want player 0, card Newcomer, inst %d, amount 1", p, inst.ID)
	}
	checkPlaceDetail(t, p)

	// A two-slot targeted deploy trigger: two Target decisions, then one place.
	s = newScene(t)
	s.put(1, ZoneMainTheater, unit("Enemy 1", 1, 9, Keywords{}))
	s.put(1, ZoneMainTheater, unit("Enemy 2", 1, 9, Keywords{}))
	slot := tgt(ability.AllegianceOpposing, ability.KindUnit, false)
	s.script(0, Answer{Picks: []int{0}}, Answer{Picks: []int{0}})
	src := s.deployTargeter(0, []ability.Target{slot, slot}, dmgTo(1, 0), dmgTo(1, 1))
	j := indexOfKind(s.g.Log, "deploy")
	if j < 0 || len(s.g.Log) < j+4 {
		t.Fatalf("log too short after the deploy: %d events", len(s.g.Log))
	}
	for k := 1; k <= 2; k++ {
		if e := s.g.Log[j+k]; e.Kind != "decision" || e.Decision == nil || e.Decision.Kind != "Target" {
			t.Fatalf("event after deploy +%d = %s, want a Target decision", k, e.Kind)
		}
	}
	ab := s.g.Log[j+3]
	if ab.Kind != "place" || ab.Player != 0 || ab.Source != src.ID || ab.Card != "Source" || ab.Inst != src.Inst.ID {
		t.Errorf("ability place = %+v, want player 0, source %d, card Source, inst %d", ab, src.ID, src.Inst.ID)
	}
	checkPlaceDetail(t, ab)

	// An attack with an Aggressive attacker: Attack place, then Aggressive place at depth 2.
	s = newScene(t)
	raider := s.put(0, ZoneMainTheater, unit("Raider", 1, 9, Keywords{Aggressive: 2}))
	def := s.put(1, ZoneMainTheater, unit("Target", 1, 9, Keywords{}))
	s.do(s.attack(raider, def))
	places := eventsOfKind(s.g.Log, "place")
	if len(places) != 2 {
		t.Fatalf("got %d place events, want 2", len(places))
	}
	if p := places[0]; p.Player != 0 || p.Source != raider.ID || p.Target != def.ID || p.Amount != 1 {
		t.Errorf("attack place = %+v, want player 0, source %d, target %d, amount 1", p, raider.ID, def.ID)
	}
	if p := places[1]; p.Player != 0 || p.Source != raider.ID || p.Amount != 2 {
		t.Errorf("Aggressive place = %+v, want player 0, source %d, amount 2", p, raider.ID)
	}
	for _, p := range places {
		checkPlaceDetail(t, p)
	}
}
