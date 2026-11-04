package main

import (
	"bytes"
	"io"
	"os"
	"slices"
	"strings"
	"testing"
)

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

func TestUsage_ListsNoAuditCommand(t *testing.T) {
	out := captureOutput(printUsage)
	if strings.Contains(strings.ToLower(out), "audit") {
		t.Errorf("usage mentions audit:\n%s", out)
	}
	lines := strings.Split(out, "\n")
	for _, cmd := range []string{"play", "lint", "hook", "sim", "replay", "whats-up"} {
		prefix := "  " + cmd + " "
		n := 0
		for _, l := range lines {
			if strings.HasPrefix(l, prefix) {
				n++
			}
		}
		if n != 1 {
			t.Errorf("usage has %d lines starting %q, want 1", n, prefix)
		}
	}
}

func TestUsage_ListsOnlyLiveCommands(t *testing.T) {
	lines := strings.Split(captureOutput(printUsage), "\n")
	var got []string
	inCommands := false
	for _, l := range lines {
		switch {
		case l == "Available Commands:":
			inCommands = true
		case inCommands && strings.HasPrefix(l, "  "):
			got = append(got, strings.Fields(l)[0])
		default:
			inCommands = false
		}
	}
	want := []string{"play", "lint", "hook", "sim", "replay", "whats-up"}
	if !slices.Equal(got, want) {
		t.Errorf("usage commands = %q, want %q", got, want)
	}
	const lint = "  lint      Unified developer linter suite (FRD, executable rules, CSV sheets, rules syntax, fences)"
	if !slices.Contains(lines, lint) {
		t.Errorf("usage lacks the line %q", lint)
	}
}
