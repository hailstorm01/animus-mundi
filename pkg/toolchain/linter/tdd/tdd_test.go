package tdd_test

import (
	"path/filepath"
	"testing"

	"animus-mundi/pkg/toolchain/linter/tdd"
)

func TestTddLint(t *testing.T) {
	repoRoot, err := filepath.Abs("../../../../")
	if err != nil {
		t.Fatalf("failed to resolve repo root: %v", err)
	}

	ok, err := tdd.Lint(repoRoot, nil)
	if err != nil {
		t.Fatalf("unexpected error during TDD lint: %v", err)
	}
	if !ok {
		t.Errorf("expected TDD lint to pass, but failed")
	}
}
