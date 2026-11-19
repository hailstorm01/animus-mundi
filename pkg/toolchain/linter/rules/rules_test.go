package rules_test

import (
	"path/filepath"
	"testing"

	"animus-mundi/pkg/toolchain/linter/rules"
)

func TestRulesLint(t *testing.T) {
	repoRoot, err := filepath.Abs("../../../../")
	if err != nil {
		t.Fatalf("failed to resolve repo root: %v", err)
	}

	ok, err := rules.Lint(repoRoot, nil)
	if err != nil {
		t.Fatalf("unexpected error during rules lint: %v", err)
	}
	if !ok {
		t.Errorf("expected rules lint to pass, but failed")
	}
}

// Reads covers FRD 01 and the Go files under pkg/engine (directive #172).
func TestReads(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		{"docs/frd/01_game_rules_frd.md", true},
		{"pkg/engine/x.go", true},
		{"pkg/engine/x_test.go", true},
		{"pkg/sim/x.go", false},
	}
	for _, c := range cases {
		if got := rules.Reads(c.path); got != c.want {
			t.Errorf("rules.Reads(%q) = %v, want %v", c.path, got, c.want)
		}
	}
}
