package main

import (
	"reflect"
	"strings"
	"testing"
)

func reportText(status, findings string) string {
	return "# Structural Review: m3 wave A\n\n- **Status:** " + status + "\r\n- **Findings:** " + findings + "\n"
}

func TestParseReport(t *testing.T) {
	got := parseReport("main", "p", reportText("open", "5 (2 duplicates)"))
	want := report{path: "p", branch: "main", status: "open", findings: "5 (2 duplicates)"}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
	if rep := parseReport("main", "p", "# nothing here\n"); rep.decided() {
		t.Error("a report without a Status line counts as decided")
	}
}

func TestReportProblem(t *testing.T) {
	for _, c := range []struct {
		status, findings string
		ok               bool
	}{
		{"open", "5 (2 duplicates)", true},
		{"decided (no findings)", "0 (0 duplicates)", true},
		{"decided (2026-11-24)", "5 (2 duplicates)", false}, // only the user decides findings
		{"open", "0", false},
		{"", "5", false},
		{"open", "", false},
		{"open", "several", false},
	} {
		got := parseReport("b", "p", reportText(c.status, c.findings)).problem()
		if (got == "") != c.ok {
			t.Errorf("status %q, findings %q: problem %q, want ok %t", c.status, c.findings, got, c.ok)
		}
	}
}

func TestUndecided(t *testing.T) {
	reps := []report{
		{path: "a", branch: "main", status: "decided (2026-11-24)"},
		{path: "b", branch: "main", status: "open"},
		{path: "a", branch: "autoloop/m3-wave-a", status: "open"}, // stale copy of a decided report
		{path: "b", branch: "autoloop/m3-wave-b", status: "open"},
		{path: "c", branch: "autoloop/m3-wave-c", status: ""},
		{path: "own", branch: "autoloop/m3-wave-d", status: "open"},
	}
	var got []string
	for _, rep := range undecided(reps, "own") {
		got = append(got, rep.branch+":"+rep.path)
	}
	if want := []string{"main:b", "autoloop/m3-wave-c:c"}; !reflect.DeepEqual(got, want) {
		t.Errorf("undecided = %q, want %q", got, want)
	}
}

func TestBlockedPrintsTheDecidePrompt(t *testing.T) {
	out := strings.Join(blocked([]report{{path: "p", branch: "main", status: "open", findings: "1 (0 duplicates)"}}), "\n")
	if want := "@lead-architect decide structural review (branch main, report p)"; !strings.Contains(out, want) {
		t.Errorf("%q does not contain %q", out, want)
	}
}

func TestNote(t *testing.T) {
	r := &run{}
	r.note(report{status: "decided (no findings)", findings: "0"})
	if r.undecided || len(r.notes) != 1 {
		t.Errorf("a decided report gave undecided %t, notes %q", r.undecided, r.notes)
	}
	r.note(report{status: "open", findings: "1"})
	if !r.undecided || len(r.notes) != 3 {
		t.Errorf("an open report gave undecided %t, notes %q", r.undecided, r.notes)
	}
}
