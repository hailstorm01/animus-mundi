package cards_test

import (
	"path/filepath"
	"testing"

	"animus-mundi/pkg/toolchain/linter/cards"
)

func TestSheetsLint(t *testing.T) {
	repoRoot, err := filepath.Abs("../../../../")
	if err != nil {
		t.Fatalf("failed to resolve repo root: %v", err)
	}

	ok, err := cards.Lint(repoRoot, nil)
	if err != nil {
		t.Fatalf("unexpected error during cards lint: %v", err)
	}
	if !ok {
		t.Errorf("expected card sheets lint to pass, but failed")
	}
}
