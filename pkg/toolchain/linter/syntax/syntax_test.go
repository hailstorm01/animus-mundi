package syntax_test

import (
	"path/filepath"
	"testing"

	"animus-mundi/pkg/toolchain/linter/syntax"
)

func TestSyntaxLint(t *testing.T) {
	repoRoot, err := filepath.Abs("../../../../")
	if err != nil {
		t.Fatalf("failed to resolve repo root: %v", err)
	}

	ok, err := syntax.Lint(repoRoot, nil, true)
	if err != nil {
		t.Fatalf("unexpected error during syntax lint: %v", err)
	}
	if !ok {
		t.Errorf("expected card syntax lint to pass, but failed")
	}
}
