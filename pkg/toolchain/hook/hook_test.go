package hook

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func findSh(t *testing.T) string {
	t.Helper()
	if p, err := exec.LookPath("sh"); err == nil {
		return p
	}
	if runtime.GOOS == "windows" {
		for _, candidate := range []string{
			`C:\Program Files\Git\bin\sh.exe`,
			`C:\Program Files\Git\usr\bin\sh.exe`,
		} {
			if _, err := os.Stat(candidate); err == nil {
				return candidate
			}
		}
	}
	t.Skip("sh not found on PATH")
	return ""
}

func checkTools(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not found on PATH")
	}
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go not found on PATH")
	}
	return findSh(t)
}

func writeMainGo(t *testing.T, dir, marker string, exitCode int) {
	t.Helper()
	content := fmt.Sprintf(`package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println(%q)
	os.Exit(%d)
}
`, marker, exitCode)
	p := filepath.Join(dir, "cmd", "cardtown", "main.go")
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write main.go: %v", err)
	}
}

func setupFixture(t *testing.T) (string, string) {
	t.Helper()
	shPath := checkTools(t)
	dir := t.TempDir()

	gitCmd := exec.Command("git", "init")
	gitCmd.Dir = dir
	if out, err := gitCmd.CombinedOutput(); err != nil {
		t.Fatalf("git init failed: %v, output: %s", err, string(out))
	}

	goMod := "module fixture\n\ngo 1.24\n"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0644); err != nil {
		t.Fatalf("failed to write go.mod: %v", err)
	}

	cmdDir := filepath.Join(dir, "cmd", "cardtown")
	if err := os.MkdirAll(cmdDir, 0755); err != nil {
		t.Fatalf("failed to mkdir cmdDir: %v", err)
	}
	writeMainGo(t, dir, "fixture lint v1", 0)

	pkgDir := filepath.Join(dir, "pkg", "foo")
	if err := os.MkdirAll(pkgDir, 0755); err != nil {
		t.Fatalf("failed to mkdir pkgDir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pkgDir, "foo.go"), []byte("package foo\n"), 0644); err != nil {
		t.Fatalf("failed to write foo.go: %v", err)
	}

	if err := os.MkdirAll(filepath.Join(dir, "tools", "linter"), 0755); err != nil {
		t.Fatalf("failed to mkdir tools/linter: %v", err)
	}

	if err := Install(dir); err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	return dir, shPath
}

func binaryName() string {
	if runtime.GOOS == "windows" {
		return "cardtown.exe"
	}
	return "cardtown"
}

func buildFixtureBinary(t *testing.T, dir string) string {
	t.Helper()
	binDir := filepath.Join(dir, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		t.Fatalf("failed to create bin dir: %v", err)
	}
	binPath := filepath.Join(binDir, binaryName())
	cmd := exec.Command("go", "build", "-o", binPath, "./cmd/cardtown")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v, output: %s", err, string(out))
	}
	return binPath
}

func runHook(t *testing.T, shPath, dir string) (string, int) {
	t.Helper()
	cmd := exec.Command(shPath, ".git/hooks/pre-commit")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("failed to execute hook: %v", err)
		}
	}
	return string(out), exitCode
}

func TestHook_Case1_FreshBinary(t *testing.T) {
	dir, shPath := setupFixture(t)
	binPath := buildFixtureBinary(t, dir)

	baseTime := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	_ = os.Chtimes(filepath.Join(dir, "go.mod"), baseTime, baseTime)
	_ = os.Chtimes(filepath.Join(dir, "cmd", "cardtown", "main.go"), baseTime, baseTime)
	_ = os.Chtimes(filepath.Join(dir, "pkg", "foo", "foo.go"), baseTime, baseTime)

	newerBinTime := baseTime.Add(10 * time.Minute)
	if err := os.Chtimes(binPath, newerBinTime, newerBinTime); err != nil {
		t.Fatalf("failed to chtimes binPath: %v", err)
	}

	// Change source to print v2, but set its time to baseTime (older than binary)
	writeMainGo(t, dir, "fixture lint v2", 0)
	_ = os.Chtimes(filepath.Join(dir, "cmd", "cardtown", "main.go"), baseTime, baseTime)

	binInfoBefore, err := os.Stat(binPath)
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}

	out, code := runHook(t, shPath, dir)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d. Output: %s", code, out)
	}
	if !strings.Contains(out, "fixture lint v1") {
		t.Errorf("expected output to contain 'fixture lint v1', got: %s", out)
	}
	if strings.Contains(strings.ToLower(out), "rebuild") || strings.Contains(strings.ToLower(out), "stale") || strings.Contains(strings.ToLower(out), "missing") {
		t.Errorf("expected no rebuild message for fresh binary, got: %s", out)
	}

	binInfoAfter, err := os.Stat(binPath)
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}
	if !binInfoBefore.ModTime().Equal(binInfoAfter.ModTime()) {
		t.Errorf("binary modification time changed on fresh binary check")
	}
}

func TestHook_Case2_StaleFromCmd(t *testing.T) {
	dir, shPath := setupFixture(t)
	binPath := buildFixtureBinary(t, dir)

	baseTime := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	_ = os.Chtimes(filepath.Join(dir, "go.mod"), baseTime, baseTime)
	_ = os.Chtimes(filepath.Join(dir, "pkg", "foo", "foo.go"), baseTime, baseTime)
	_ = os.Chtimes(binPath, baseTime, baseTime)

	// Change source to v2 and set time newer than binary
	writeMainGo(t, dir, "fixture lint v2", 0)
	newerTime := baseTime.Add(10 * time.Minute)
	_ = os.Chtimes(filepath.Join(dir, "cmd", "cardtown", "main.go"), newerTime, newerTime)

	out, code := runHook(t, shPath, dir)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d. Output: %s", code, out)
	}
	if !strings.Contains(strings.ToLower(out), "stale") {
		t.Errorf("expected rebuild line naming 'stale', got: %s", out)
	}
	if !strings.Contains(out, binaryName()) {
		t.Errorf("expected rebuild line naming %q, got: %s", binaryName(), out)
	}
	if !strings.Contains(out, "fixture lint v2") {
		t.Errorf("expected output to contain updated 'fixture lint v2', got: %s", out)
	}
}

func TestHook_Case3_StaleFromPkg(t *testing.T) {
	dir, shPath := setupFixture(t)
	binPath := buildFixtureBinary(t, dir)

	baseTime := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	_ = os.Chtimes(filepath.Join(dir, "go.mod"), baseTime.Add(-10*time.Minute), baseTime.Add(-10*time.Minute))
	_ = os.Chtimes(filepath.Join(dir, "cmd", "cardtown", "main.go"), baseTime.Add(-10*time.Minute), baseTime.Add(-10*time.Minute))
	_ = os.Chtimes(binPath, baseTime, baseTime)

	// Only pkg .go file is newer
	newerTime := baseTime.Add(10 * time.Minute)
	_ = os.Chtimes(filepath.Join(dir, "pkg", "foo", "foo.go"), newerTime, newerTime)

	out, code := runHook(t, shPath, dir)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d. Output: %s", code, out)
	}
	if !strings.Contains(strings.ToLower(out), "stale") {
		t.Errorf("expected rebuild line naming 'stale', got: %s", out)
	}
}

func TestHook_Case4_StaleFromGoMod(t *testing.T) {
	dir, shPath := setupFixture(t)
	binPath := buildFixtureBinary(t, dir)

	baseTime := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	_ = os.Chtimes(filepath.Join(dir, "cmd", "cardtown", "main.go"), baseTime.Add(-10*time.Minute), baseTime.Add(-10*time.Minute))
	_ = os.Chtimes(filepath.Join(dir, "pkg", "foo", "foo.go"), baseTime.Add(-10*time.Minute), baseTime.Add(-10*time.Minute))
	_ = os.Chtimes(binPath, baseTime, baseTime)

	// Only go.mod is newer
	newerTime := baseTime.Add(10 * time.Minute)
	_ = os.Chtimes(filepath.Join(dir, "go.mod"), newerTime, newerTime)

	out, code := runHook(t, shPath, dir)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d. Output: %s", code, out)
	}
	if !strings.Contains(strings.ToLower(out), "stale") {
		t.Errorf("expected rebuild line naming 'stale', got: %s", out)
	}
}

func TestHook_Case5_NotStaleFromOtherFiles(t *testing.T) {
	dir, shPath := setupFixture(t)
	binPath := buildFixtureBinary(t, dir)

	baseTime := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	_ = os.Chtimes(filepath.Join(dir, "go.mod"), baseTime.Add(-10*time.Minute), baseTime.Add(-10*time.Minute))
	_ = os.Chtimes(filepath.Join(dir, "pkg", "foo", "foo.go"), baseTime.Add(-10*time.Minute), baseTime.Add(-10*time.Minute))
	_ = os.Chtimes(binPath, baseTime, baseTime)

	// Change cmd source but keep mtime older than binary
	writeMainGo(t, dir, "fixture lint v2", 0)
	_ = os.Chtimes(filepath.Join(dir, "cmd", "cardtown", "main.go"), baseTime.Add(-10*time.Minute), baseTime.Add(-10*time.Minute))

	// Write newer .md in pkg and newer .go in tools/
	newerTime := baseTime.Add(10 * time.Minute)
	mdPath := filepath.Join(dir, "pkg", "foo", "readme.md")
	if err := os.WriteFile(mdPath, []byte("# Note\n"), 0644); err != nil {
		t.Fatalf("failed to write md: %v", err)
	}
	_ = os.Chtimes(mdPath, newerTime, newerTime)

	toolsDir := filepath.Join(dir, "tools", "helper")
	_ = os.MkdirAll(toolsDir, 0755)
	toolPath := filepath.Join(toolsDir, "helper.go")
	if err := os.WriteFile(toolPath, []byte("package helper\n"), 0644); err != nil {
		t.Fatalf("failed to write helper.go: %v", err)
	}
	_ = os.Chtimes(toolPath, newerTime, newerTime)

	out, code := runHook(t, shPath, dir)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d. Output: %s", code, out)
	}
	if !strings.Contains(out, "fixture lint v1") {
		t.Errorf("expected output to contain 'fixture lint v1', got: %s", out)
	}
	if strings.Contains(strings.ToLower(out), "rebuild") || strings.Contains(strings.ToLower(out), "stale") || strings.Contains(strings.ToLower(out), "missing") {
		t.Errorf("expected no rebuild message when only non-matching files are newer, got: %s", out)
	}
}

func TestHook_Case6_MissingBinary(t *testing.T) {
	dir, shPath := setupFixture(t)
	// No bin directory created

	out, code := runHook(t, shPath, dir)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d. Output: %s", code, out)
	}
	if !strings.Contains(strings.ToLower(out), "missing") {
		t.Errorf("expected rebuild line naming 'missing', got: %s", out)
	}
	if !strings.Contains(out, binaryName()) {
		t.Errorf("expected rebuild line naming %q, got: %s", binaryName(), out)
	}

	binPath := filepath.Join(dir, "bin", binaryName())
	if _, err := os.Stat(binPath); err != nil {
		t.Errorf("expected binary %s to exist after rebuild, stat error: %v", binPath, err)
	}
	if !strings.Contains(out, "fixture lint v1") {
		t.Errorf("expected output to contain current marker 'fixture lint v1', got: %s", out)
	}
}

func TestHook_Case7_BrokenBuild(t *testing.T) {
	dir, shPath := setupFixture(t)
	binPath := buildFixtureBinary(t, dir)

	baseTime := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	_ = os.Chtimes(binPath, baseTime, baseTime)

	// Add broken.go with a syntax error and set newer time
	brokenPath := filepath.Join(dir, "cmd", "cardtown", "broken.go")
	if err := os.WriteFile(brokenPath, []byte("package main\n broken syntax {{\n"), 0644); err != nil {
		t.Fatalf("failed to write broken.go: %v", err)
	}
	newerTime := baseTime.Add(10 * time.Minute)
	_ = os.Chtimes(brokenPath, newerTime, newerTime)

	out, code := runHook(t, shPath, dir)
	if code == 0 {
		t.Fatalf("expected non-zero exit on broken build, got 0. Output: %s", out)
	}
	if !strings.Contains(out, "broken.go") {
		t.Errorf("expected compiler error naming 'broken.go', got: %s", out)
	}
	if !strings.Contains(out, "go run") {
		t.Errorf("expected fallback line naming 'go run', got: %s", out)
	}
	if strings.Contains(out, "fixture lint v1") {
		t.Errorf("old binary's marker should not appear after failed rebuild, got: %s", out)
	}
}

func TestHook_Case8_LintFails(t *testing.T) {
	dir, shPath := setupFixture(t)
	writeMainGo(t, dir, "fixture lint v1", 1)
	buildFixtureBinary(t, dir)

	out, code := runHook(t, shPath, dir)
	if code != 1 {
		t.Fatalf("expected exit 1 when lint fails, got %d. Output: %s", code, out)
	}
	if !strings.Contains(out, "Pre-commit linting failed!") {
		t.Errorf("expected output to contain 'Pre-commit linting failed!', got: %s", out)
	}
}

func TestHook_Case9_InstalledCopyMatches(t *testing.T) {
	shPath := checkTools(t)
	_ = shPath
	dir := t.TempDir()

	hooksDir := filepath.Join(dir, ".git", "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		t.Fatalf("mkdir hooksDir failed: %v", err)
	}
	toolsLinterDir := filepath.Join(dir, "tools", "linter")
	if err := os.MkdirAll(toolsLinterDir, 0755); err != nil {
		t.Fatalf("mkdir toolsLinterDir failed: %v", err)
	}

	if err := Install(dir); err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	gitHook, err := os.ReadFile(filepath.Join(hooksDir, "pre-commit"))
	if err != nil {
		t.Fatalf("read git hook failed: %v", err)
	}
	toolsHook, err := os.ReadFile(filepath.Join(toolsLinterDir, "pre-commit"))
	if err != nil {
		t.Fatalf("read tools hook failed: %v", err)
	}

	expected := []byte(preCommitHookScript)
	if !bytes.Equal(gitHook, expected) {
		t.Errorf(".git/hooks/pre-commit does not match preCommitHookScript")
	}
	if !bytes.Equal(toolsHook, expected) {
		t.Errorf("tools/linter/pre-commit does not match preCommitHookScript")
	}
}
