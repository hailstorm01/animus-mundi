package eventlog_test

import (
	"strings"
	"testing"

	"animus-mundi/pkg/engine"
	"animus-mundi/pkg/eventlog"
)

// T-0234 (directive #149 (b)): a decision's source card is part of the record
// a replay compares.
func TestReplay_DecisionWithAnotherSourceDiverges(t *testing.T) {
	for _, evs := range searchGames(t) {
		for i, e := range evs {
			if e.Decision == nil || e.Decision.Source == "" {
				continue
			}
			cp := append([]engine.Event{}, evs...)
			decCopy := *e.Decision
			decCopy.Source = "Another Card"
			cp[i].Decision = &decCopy
			_, err := eventlog.Replay(cp, len(cp), searchResolve)
			if err == nil || !strings.Contains(err.Error(), "replay diverged") {
				t.Fatalf("decision with another source: error %v, want replay diverged", err)
			}
			return
		}
	}
	t.Fatal("no game holds a decision with a source")
}
