package hook

import (
	"fmt"
	"os"
	"path/filepath"
)

const preCommitHookScript = `#!/bin/sh
# Card Town Git Pre-Commit Hook (Native Go)
# Fast-paths linting through compiled cardtown binary or fallback to go run.

REPO_ROOT="$(git rev-parse --show-toplevel 2>/dev/null)"
if [ -z "$REPO_ROOT" ]; then
    REPO_ROOT="."
fi

CARDTOWN_BIN="$REPO_ROOT/bin/cardtown.exe"
if [ ! -f "$CARDTOWN_BIN" ]; then
    CARDTOWN_BIN="$REPO_ROOT/bin/cardtown"
fi

if [ -f "$CARDTOWN_BIN" ]; then
    "$CARDTOWN_BIN" lint --staged --repo-root "$REPO_ROOT"
    EXIT_CODE=$?
else
    go run "$REPO_ROOT/cmd/cardtown" lint --staged --repo-root "$REPO_ROOT"
    EXIT_CODE=$?
fi

if [ $EXIT_CODE -ne 0 ]; then
    echo ""
    echo "Pre-commit linting failed! Please resolve errors before committing."
    exit 1
fi

exit 0
`

// Install installs the native Git pre-commit hook into .git/hooks/pre-commit.
func Install(repoRoot string) error {
	hooksDir := filepath.Join(repoRoot, ".git", "hooks")
	if _, err := os.Stat(hooksDir); os.IsNotExist(err) {
		return fmt.Errorf("git hooks directory not found at %s", hooksDir)
	}

	hookDest := filepath.Join(hooksDir, "pre-commit")
	if err := os.WriteFile(hookDest, []byte(preCommitHookScript), 0755); err != nil {
		return fmt.Errorf("failed to write hook to %s: %w", hookDest, err)
	}

	// Also sync tools/linter/pre-commit
	linterHookSrc := filepath.Join(repoRoot, "tools", "linter", "pre-commit")
	_ = os.WriteFile(linterHookSrc, []byte(preCommitHookScript), 0755)

	fmt.Printf("[SUCCESS] Installed Card Town native Go pre-commit hook into %s\n", hookDest)
	return nil
}
