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
