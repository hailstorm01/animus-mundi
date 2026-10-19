package frd_test

import (
	"path/filepath"
	"testing"

	"animus-mundi/pkg/toolchain/linter/frd"
)

func TestFrdLint(t *testing.T) {
	// Locate repository root relative to test
	repoRoot, err := filepath.Abs("../../../../")
	if err != nil {
		t.Fatalf("failed to resolve repo root: %v", err)
	}

	ok, err := frd.Lint(repoRoot, nil)
	if err != nil {
		t.Fatalf("unexpected error during FRD lint: %v", err)
	}
	if !ok {
		t.Errorf("expected FRD lint to pass, but failed")
	}
}
