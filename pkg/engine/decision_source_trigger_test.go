package engine

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0234 (directive #149 (b), FR-MOD-001, FR-TRG-003, FR-ZON-005):
// every decision asked while a trigger is placed, or while an ability or a
// Maneuver resolves, names its source card (Decision.Source), the trigger
// order names none, and the decision record carries the source's name.

// wantSourceNamed reports every decision in ds whose Source is not a card
// named name, and an empty ds.
func wantSourceNamed(t *testing.T, ds []Decision, name string) {
	t.Helper()
	if len(ds) == 0 {
		t.Fatalf("no decision was asked")
	}
	for i, d := range ds {
		if d.Source == nil || d.Source.Name != name {
			t.Errorf("decision %d (%s, %q): Source %v, want %s", i, d.Kind, d.Prompt, d.Source, name)
		}
	}
}

// answerFirst answers every decision with its first option, a "you may" with
// yes, and a decision that may pick none with a timeout.
func answerFirst(d Decision) Answer {
	if d.Min == 0 && d.Kind != DecideOptional {
		return Answer{Timeout: true}
	}
	picks := make([]int, max(d.Min, 1))
	for i := range picks {
		picks[i] = i
	}
	return Answer{Picks: picks}
}

func TestDecisionSource_TriggerTarget_DevilhornCaptain(t *testing.T) {
	s := newScene(t)
	b := s.captainBoard()
	hitter := s.put(0, ZoneMainTheater, unit("Hitter", 1, 9, ability.Keywords{}))
	s.byKind(1, func(d Decision) Answer {
		for i, o := range d.Options {
			if d.Kind == DecideTarget && o.Ref == b.u0.ID {
				return Answer{Picks: []int{i}}
			}
		}
		return answerFirst(d)
	})
	s.do(s.attack(hitter, b.cap))
	asked := s.asked(1)
	wantSources(t, asked, b.cap.Inst.Card)
	if len(askedKind(s, 1, DecideTarget)) != 1 {
		t.Errorf("Unbowed asks one target decision: kinds %v", kindsOf(asked))
	}
	if src := asked[len(asked)-1].Source; src != nil && !strings.Contains(src.RulesText, "Unbowed: ") {
		t.Errorf("Source.RulesText %q, want the Captain's printed text", src.RulesText)
	}
}

func TestDecisionSource_OptionalTargetAndPick_InfestedApostle(t *testing.T) {
	s := newScene(t)
	ally := s.put(0, ZoneMainTheater, unit("Ally", 0, 5, ability.Keywords{}))
	v := s.put(1, ZoneMainTheater, unit("Victim One", 0, 5, ability.Keywords{}))
	s.put(1, ZoneMainTheater, unit("Victim Two", 0, 5, ability.Keywords{}))
	s.scriptPlan(0, nil, []Answer{yes}, pick(v), pick(ally))
	inst := deployApostle(s)
	asks := s.asked(0)
	if got, want := kindsOf(asks), []DecisionKind{DecideOptional, DecideTarget, DecidePick}; !slices.Equal(got, want) {
		t.Fatalf("kinds %v, want %v", got, want)
	}
	wantSources(t, asks, inst.Card)
}

func TestDecisionSource_ModalTrigger(t *testing.T) {
	b := newModalBoard(t)
	c := trigUnit("Modal", hitArmorTrigger())
	b.s.byKind(0, answerFirst)
	b.s.deployFromHand(0, c)
	asks := b.s.asked(0)
	wantSources(t, asks, c)
	if len(asks) == 0 || asks[0].Kind != DecideMode {
		t.Errorf("first decision %v, want the mode choice", kindsOf(asks))
	}
}

func TestDecisionSource_PileTriggerAndResolvingDiscard_EerieRemnant(t *testing.T) {
	s, orders, rem, _ := remnantOrdersScene(t, yes)
	hand := askedKind(s, 0, DecideHand)
	opt := askedKind(s, 0, DecideOptional)
	if len(hand) != 1 || len(opt) != 1 {
		t.Fatalf("asked %v, want one hand discard and one optional question", kindsOf(s.asked(0)))
	}
	wantSources(t, hand, orders.Card)
	wantSources(t, opt, rem.Card)
}

func TestDecisionSource_ResolvingDeckSearch(t *testing.T) {
	s, a, _ := cliffScene(t, 3)
	s.script(0, Answer{Picks: []int{1}})
	s.do(activate(a, 0))
	wantSources(t, s.asked(0), a.Inst.Card)
	if len(askedKind(s, 0, DecideSearch)) != 1 {
		t.Errorf("asked %v, want one deck search", kindsOf(s.asked(0)))
	}
}

func TestDecisionSource_ResolvingLook(t *testing.T) {
	s, _, man := sctScene(t, sctReal("X", "X", "X"), "", nil)
	if len(askedKind(s, 0, DecideLook)) != 1 {
		t.Fatalf("asked %v, want one look", kindsOf(s.asked(0)))
	}
	wantSources(t, s.asked(0), man.Card)
}

func TestDecisionSource_FreeDeployZone(t *testing.T) {
	s, _, _ := rrGuerrillas(t, Answer{Timeout: true})
	wantSourceNamed(t, rrZoneAsks(s), "Reaver's Rush")
}

func TestDecisionSource_TriggerOrderNamesNone(t *testing.T) {
	s := newScene(t)
	s.onTop(0, vanilla("A", "B")...)
	s.script(0, Answer{Picks: []int{0, 1}})
	s.deployFromHand(0, trigUnit("Source", deployTrigger(drawN(1)), deployTrigger(scratchN(1))))
	asks := s.asked(0)
	if len(asks) != 1 || asks[0].Kind != DecideOrder {
		t.Fatalf("asked %v, want one trigger order", kindsOf(asks))
	}
	if asks[0].Source != nil {
		t.Errorf("the trigger order: Source %v, want nil", asks[0].Source)
	}
}

func TestDecisionSource_FlushInsideAScopeOrdersWithNone(t *testing.T) {
	s := newScene(t)
	src := s.put(0, ZoneReserves, trigUnit("Source", deployTrigger(drawN(1)), deployTrigger(scratchN(1))))
	outer := unit("Outer", 1, 1, ability.Keywords{})
	s.g.fireHook(assetSource(src), ability.HookDeploy, nil)
	s.script(0, Answer{Picks: []int{0, 1}})
	closeOuter := s.g.askFor(outer)
	s.g.flushTriggers()
	if s.g.asking != outer {
		t.Errorf("after the flush the outer scope is not restored: asking %v, want Outer", s.g.asking)
	}
	closeOuter()
	asks := s.asked(0)
	if len(asks) != 1 || asks[0].Kind != DecideOrder || asks[0].Source != nil {
		t.Errorf("asked %v, want one trigger order with no Source", kindsOf(asks))
	}
}

func TestDecisionSource_Record(t *testing.T) {
	s := newScene(t)
	s.put(1, ZoneMainTheater, unit("Foe", 0, 5, ability.Keywords{}))
	s.onTop(0, vanilla("A", "B")...)
	s.script(0, Answer{Picks: []int{0, 1}})
	hit := ability.Trigger{Hook: ability.HookDeploy, Effect: hitMode()}
	s.deployFromHand(0, trigUnit("Hitter", deployTrigger(drawN(1)), hit))
	var order, target *DecisionRecord
	for _, e := range eventsOfKind(s.g.Log, EventDecision) {
		switch e.Decision.Kind {
		case DecideOrder.String():
			order = e.Decision
		case DecideTarget.String():
			target = e.Decision
		}
	}
	if order == nil || target == nil {
		t.Fatalf("want a trigger-order record and a target record")
	}
	if order.Source != "" {
		t.Errorf("the trigger order's record: source %q, want none", order.Source)
	}
	if target.Source != "Hitter" || target.How != "auto" {
		t.Errorf("the sole target's record: source %q how %q, want Hitter and auto", target.Source, target.How)
	}
	for _, tc := range []struct {
		rec  DecisionRecord
		want bool
	}{{*order, false}, {*target, true}} {
		js, err := json.Marshal(tc.rec)
		if err != nil {
			t.Fatal(err)
		}
		if got := strings.Contains(string(js), `"source":`); got != tc.want {
			t.Errorf("%s: has \"source\" %t, want %t: %s", tc.rec.Kind, got, tc.want, js)
		}
	}
}
