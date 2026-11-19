package engine

import (
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0282 (directive #170, FR-ZON-005): an untargeted pick's prompt,
// and a pool's, name the card the stack item shows as its source (SourceCard),
// so a source turned face down after its ability was placed shows only the
// Scratch face, in the decision asked and in its match record.

func TestFR_ZON_005_PickPromptNamesTheShownFace(t *testing.T) {
	// Covers: FR-MOD-001
	enemy := ability.Filter{Allegiance: ability.AllegianceOpposing, Kinds: ability.KindUnit}
	for _, tc := range []struct {
		name string
		step ability.Step
		want string
	}{
		{"pick", ability.Step{Op: ability.OpKill, Subject: ability.SubjectPick, Filter: enemy}, "Scratch Company: pick one"},
		{"pick up to", ability.Step{Op: ability.OpKill, Subject: ability.SubjectPick, Filter: enemy, UpTo: 2}, "Scratch Company: pick up to 2"},
		{"pool", ability.Step{Op: ability.OpAddArmor, Amount: ability.Value{N: 2}, Subject: ability.SubjectPick, Split: true, Filter: enemy}, "Scratch Company: distribute 2"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := newScene(t)
			a := s.put(0, ZoneMainTheater, unit("Hidden Source", 1, 1, ability.Keywords{}))
			s.put(1, ZoneMainTheater, unit("Foe A", 1, 5, ability.Keywords{}))
			s.put(1, ZoneMainTheater, unit("Foe B", 1, 5, ability.Keywords{}))
			s.g.push(&StackItem{Kind: StackAbility, Controller: 0, Source: a.ID, Inst: a.Inst,
				Effect: ability.Effect{Steps: []ability.Step{tc.step}}})
			a.FaceDown = true
			s.script(0, Answer{Timeout: true})
			s.g.resolveStack()
			var asked []Decision
			for _, d := range s.asked(0) {
				if d.Kind == DecidePick {
					asked = append(asked, d)
				}
			}
			if len(asked) != 1 {
				t.Fatalf("asked %d pick decisions, want 1", len(asked))
			}
			if asked[0].Prompt != tc.want {
				t.Errorf("prompt %q, want %q", asked[0].Prompt, tc.want)
			}
			var recs []*DecisionRecord
			for _, e := range eventsOfKind(s.g.Log, EventDecision) {
				if e.Decision.Kind == DecidePick.String() {
					recs = append(recs, e.Decision)
				}
			}
			if len(recs) != 1 {
				t.Fatalf("logged %d pick records, want 1", len(recs))
			}
			if recs[0].Prompt != tc.want || recs[0].Source != "Scratch Company" {
				t.Errorf("record prompt %q, source %q; want %q and the Scratch face", recs[0].Prompt, recs[0].Source, tc.want)
			}
		})
	}
}
