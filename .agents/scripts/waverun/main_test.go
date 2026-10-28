package main

import (
	"reflect"
	"testing"
)

const runFile = "# Wave Run: m2 wave B\n" +
	"\n" +
	"- **Run branch:** `autoloop/m2-wave-b`, from `main` at `4d0abf0`\n" +
	"- **Cards:** Hammerhand Brigands; Allshan, Chainbreaker Revenant; Rogue Fleshshaper\n" +
	"- **Deferred:** Blind Defile (#14 (b))\n" +
	"- **Kept:** Rogue Fleshshaper (non-deck; Ambush Tactics needs its modal shape, #14 (b), (c))\n" +
	"\n" +
	"| # | Date | Task | Status | Merge | Review rounds | Directives added | Sim (starter) | Signal |\n" +
	"| :---: | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |\n" +
	"| 1 | 2026-10-26 | T-0067 | done | `124cf3d` | tests 1 | none | clean | CONTINUE |\n" +
	"| 2 | 2026-10-26 | T-0068 | ready | none | tests 1 | none | not run | STOP: Flash run did not finish |\r\n" +
	"| 3 | 2026-10-28 | T-0088 | in-progress | none | code 0 | none | not run | STOP: denied `Get-Content a \\| Select-Object -First 1` |\n"

const cardsSrc = `package cards

// NotYetExecutable lists every starter-set card the kernel cannot execute yet.
var NotYetExecutable = []string{
	"Allshan, Chainbreaker Revenant",
	"Ambush Tactics",
	"Rogue Fleshshaper",
}

var other = []string{"Hammerhand Brigands"}
`

func TestRunRowsAndSignal(t *testing.T) {
	rows := runRows(lines(runFile))
	if len(rows) != 3 {
		t.Fatalf("got %d rows, want 3 (header and divider are not rows)", len(rows))
	}
	for i, want := range []string{"CONTINUE", "STOP: Flash run did not finish", "STOP: denied `Get-Content a | Select-Object -First 1`"} {
		if got := signal(rows[i]); got != want {
			t.Errorf("row %d signal = %q, want %q", i+1, got, want)
		}
	}
}

func TestRunRowsEmpty(t *testing.T) {
	if rows := runRows(nil); len(rows) != 0 {
		t.Errorf("no run file gave %d rows", len(rows))
	}
}

func TestCardsLine(t *testing.T) {
	names, ok := cardsLine(lines(runFile))
	want := []string{"Hammerhand Brigands", "Allshan, Chainbreaker Revenant", "Rogue Fleshshaper"}
	if !ok || !reflect.DeepEqual(names, want) {
		t.Errorf("cardsLine = %q, %t; want %q", names, ok, want)
	}
	if _, ok := cardsLine([]string{"- **Kept:** none"}); ok {
		t.Error("found a Cards line in a header without one")
	}
}

func TestPendingCards(t *testing.T) {
	listed, err := notYetExecutable(cardsSrc)
	if err != nil {
		t.Fatal(err)
	}
	names, _ := cardsLine(lines(runFile))
	got := pending(names, listed)
	want := []string{"Allshan, Chainbreaker Revenant", "Rogue Fleshshaper"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("pending = %q, want %q", got, want)
	}
}

func TestNotYetExecutableMissing(t *testing.T) {
	if _, err := notYetExecutable("package cards\n"); err == nil {
		t.Error("no error for cards.go without NotYetExecutable")
	}
}
