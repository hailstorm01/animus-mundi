package engine

import "testing"

// Directive #86: each event kind is a named constant whose value is the
// ADR-0005 schema string, unchanged.
func TestEventKinds_ValuesAreTheSchema(t *testing.T) {
	cases := []struct{ got, want string }{
		{EventMatchStart, "match-start"},
		{EventTurnStart, "turn-start"},
		{EventAction, "action"},
		{EventDecision, "decision"},
		{EventInvalidAnswer, "invalid-answer"},
		{EventGenerate, "generate"},
		{EventPlace, "place"},
		{EventDeploy, "deploy"},
		{EventMarch, "march"},
		{EventScratch, "scratch"},
		{EventAttack, "attack"},
		{EventFizzle, "fizzle"},
		{EventCombat, "combat"},
		{EventDamage, "damage"},
		{EventReplace, "replace"},
		{EventDestroy, "destroy"},
		{EventLeave, "leave"},
		{EventDraw, "draw"},
		{EventDiscard, "discard"},
		{EventResult, "result"},
		{EventMatchEnd, "match-end"},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("event kind %q, want %q", c.got, c.want)
		}
	}
}
