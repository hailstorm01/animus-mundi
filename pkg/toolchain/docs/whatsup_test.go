package docs

import (
	"bytes"
	"io"
	"os"
	"slices"
	"strings"
	"testing"
)

const repoRoot = "../../.."

func captureOutput(f func()) string {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		outC <- buf.String()
	}()

	f()

	_ = w.Close()
	os.Stdout = oldStdout
	return <-outC
}

func TestShowTools_ListsNoAuditTool(t *testing.T) {
	out := captureOutput(func() { ShowTools(repoRoot) })
	if strings.Contains(out, "cardtown audit") {
		t.Errorf("tools section lists an audit tool:\n%s", out)
	}
	for _, want := range []string{"cardtown lint", "cardtown sim", "cardtown hook install", "cardtown whats-up"} {
		if !strings.Contains(out, want) {
			t.Errorf("tools section lacks %q", want)
		}
	}
}

func TestShowState_PrintsNoPendingAudits(t *testing.T) {
	out := captureOutput(func() { ShowState(repoRoot) })
	lines := strings.Split(out, "\n")
	var hasGit, hasHook bool
	for i, l := range lines {
		switch {
		case l == "Pending audits:":
			t.Errorf("state section prints %q", l)
		case strings.HasPrefix(l, "Git: branch "):
			hasGit = true
		case strings.HasPrefix(l, "Pre-commit hook: "):
			hasHook = true
		case strings.HasPrefix(l, "Portfolio roadmap (docs/PRODUCT_SUMMARY.md)"):
			if i > 0 && !strings.HasPrefix(lines[i-1], "Keyword interaction matrix: ") {
				t.Errorf("portfolio line follows %q, want the keyword-matrix line", lines[i-1])
			}
		}
	}
	if !hasGit {
		t.Error("no line starting \"Git: branch \"")
	}
	if !hasHook {
		t.Error("no line starting \"Pre-commit hook: \"")
	}
}

func TestShowTools_ListsOnlyLiveTools(t *testing.T) {
	lines := strings.Split(captureOutput(func() { ShowTools(repoRoot) }), "\n")
	var got []string
	for _, l := range lines {
		if cmd, _, ok := strings.Cut(l, " : "); ok && strings.HasPrefix(cmd, "- cardtown ") {
			got = append(got, strings.TrimPrefix(cmd, "- "))
		}
	}
	want := []string{
		"cardtown lint [flags] [files...]",
		"cardtown lint rules",
		"cardtown lint fences",
		"cardtown sim [flags]",
		"cardtown hook install",
		"cardtown whats-up [flags]",
	}
	if !slices.Equal(got, want) {
		t.Errorf("tools = %q, want %q", got, want)
	}
	for _, line := range []string{
		"- cardtown lint [flags] [files...] : Unified developer linter suite (FRD, CSV sheets & syntax)",
		"    cardtown lint [--all|--frd|--rules|--cards|--sheets|--syntax|--fences|--staged] [--files <list>]",
	} {
		if !slices.Contains(lines, line) {
			t.Errorf("tools section lacks the line %q", line)
		}
	}
}

func TestShowState_PrintsNoDesignDocumentsBlock(t *testing.T) {
	out := captureOutput(func() { ShowState(repoRoot) })
	lines := strings.Split(out, "\n")
	for _, l := range lines {
		if strings.HasPrefix(l, "Technical Design Documents") {
			t.Errorf("state section prints %q", l)
		}
	}
}
