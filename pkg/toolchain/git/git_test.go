package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// runGit runs git in dir and fails the test on error.
func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return string(out)
}

// newRepo builds a throwaway repo whose identity config is local to it.
func newRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	runGit(t, root, "init")
	runGit(t, root, "config", "user.name", "test")
	runGit(t, root, "config", "user.email", "test@example.com")
	runGit(t, root, "config", "commit.gpgsign", "false")
	return root
}

func commitEmpty(t *testing.T, root, msg string) {
	t.Helper()
	runGit(t, root, "commit", "--allow-empty", "-m", msg)
}

func TestLines(t *testing.T) {
	root := newRepo(t)
	commitEmpty(t, root, "first")
	commitEmpty(t, root, "second")

	got := Lines(root, "log", "--format=%s")
	want := []string{"second", "first"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Lines log = %q, want %q", got, want)
	}
	if got := Lines(root, "no-such-subcommand"); got != nil {
		t.Errorf("Lines on failing command = %q, want nil", got)
	}
}

func TestOutput(t *testing.T) {
	root := newRepo(t)
	commitEmpty(t, root, "first")

	out, err := Output(root, "rev-parse", "--short", "HEAD")
	if err != nil {
		t.Fatalf("Output rev-parse: %v", err)
	}
	want := strings.TrimSpace(runGit(t, root, "rev-parse", "--short", "HEAD"))
	if got := strings.TrimSpace(string(out)); got != want {
		t.Errorf("Output rev-parse = %q, want %q", got, want)
	}
	if _, err := Output(root, "remote", "get-url", "origin"); err == nil {
		t.Error("Output remote get-url origin with no remote: want error, got nil")
	}
}

func TestCombinedFailure(t *testing.T) {
	root := newRepo(t)
	commitEmpty(t, root, "first")

	out, err := Combined(root, "rm", "nope.txt")
	if err == nil {
		t.Fatal("Combined rm nope.txt: want error, got nil")
	}
	if !strings.Contains(string(out), "nope.txt") {
		t.Errorf("Combined output = %q, want it to mention nope.txt", out)
	}
}

func TestCombinedSuccess(t *testing.T) {
	root := newRepo(t)
	if err := os.WriteFile(filepath.Join(root, "a.txt"), []byte("a\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if out, err := Combined(root, "add", "a.txt"); err != nil {
		t.Fatalf("Combined add: %v\n%s", err, out)
	}
	if out, err := Combined(root, "commit", "-m", "add a"); err != nil {
		t.Fatalf("Combined commit: %v\n%s", err, out)
	}
	got := Lines(root, "log", "-1", "--format=%s")
	if want := []string{"add a"}; !reflect.DeepEqual(got, want) {
		t.Errorf("Lines log -1 = %q, want %q", got, want)
	}
}
