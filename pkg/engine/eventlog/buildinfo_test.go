package eventlog_test

import (
	"testing"

	"animus-mundi/pkg/engine"
)

// T-0033 Acceptance Case 2: match-start carries the revision engine.BuildInfo reports.
func TestBuildInfo_Case2_MatchStartUsesBuildInfo(t *testing.T) {
	games := allGames(t)
	st := games[0][0].Start
	if st == nil {
		t.Fatal("first event has no match-start payload")
	}
	rev, mod := engine.BuildInfo()
	if st.Revision != rev || st.Modified != mod {
		t.Errorf("match-start = (%q, %t), BuildInfo = (%q, %t)", st.Revision, st.Modified, rev, mod)
	}
}
