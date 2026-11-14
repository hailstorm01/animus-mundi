---
id: T-0235
title: "The play screen shows a decision's source card and its rules text, and the stack listing names each item's card (directive #149 (c), wave F.2 harness, flash)"
kind: harness
tier: flash
wave: F.2
status: ready
branch: m2/autoloop-wave-f.2/T-0235-play-screen-source
depends_on: [T-0234]
claims: [func:sourceLine, func:stackLine]
cards: []
requirements: [FR-ZON-005]
---

## Goal
On `cardtown play`, a decision with a `Source` (T-0233, T-0234) prints its source card's name and printed rules text under the prompt, before the options or the y/n question, and the stack listing names the card of an ability, a Deploy and a Maneuver (#149 (c)). When this merges, #149 is `applied`.

## Inputs
- Board directive #149 (user playtest feedback, 2026-11-13), `open: before the next Deliverables task`; parts (a) and (b) applied by T-0233 (merge f99b291) and T-0234 (merge 17297e6). This task is part (c), the last.
- Design (lead decisions):
  - **The source line.** `terminalChooser` (`cmd/cardtown/play_input.go`) prints, right after `d.Prompt` and only when `d.Source` is not nil, one line from a new `sourceLine(c *engine.Card) string`: `  [<Name>] <RulesText>`, each further line of a rules text of several lines indented four spaces under the first, and `  [<Name>]` for a card with no rules text. It reads `Decision.Source`, never the prompt (#130). A decision without a source (the trigger order) prints exactly as before (#149 (d)). The mulligan is a `Config.Mulligan` callback, not a decision, so it doesn't change.
  - **Hidden information.** `Decision.Source` is the card as both seats see it (FR-ZON-005: the Scratch face for a face-down asset; `choice.go:45-50`), and the human's chooser only prints the human seat's decisions, so the line shows nothing hidden. `TestPlay_Case4_NoHiddenInformation` replays the scripted session's whole output, now with its source lines, and still passes.
  - **The stack listing.** `printScreen` (`cmd/cardtown/play_screen.go`) prints each item through a new `stackLine(item view.StackView) string`: an ability with a card reads `Ability of <card>`, or `<keyword> trigger of <card>` when `item.Keyword` is set, as the engine's place event names it (`stack.go:72-80`), followed by `(Player N, Source S)` as now; a Deploy reads `Deploy of <card> to <zone> (Player N)`; a Maneuver, a March, an Attack, a Scratch, and an ability with no card read as now. `view.StackView` already holds `Card`, `Zone` and `Keyword` (`pkg/view/view.go:82-93`), so `pkg/view` doesn't change.
- Reuse search (D2.3): `cmd/cardtown` had no source or rules-text formatting (`actionLabel` prints an Activate's `as.Text` inline). The stack naming follows the place event's wording, but the place event writes a lowercase log sentence from the engine's `StackItem`, and the screen writes from the view's `StackView`; neither can call the other without moving presentation into `engine` or `view`, and the wording is display text, not a rule, so this is not a duplicate under D2.0.
- Size (D2.4): two non-test files.
- **Expected output difference:** none in `cardtown sim` (the play screen is not in sim output), so `go tool task sim-compare` passes. `cardtown play` gains the source lines and the new stack lines.
- The lead dry-ran these exact edits and this exact test file in a throwaway worktree of the run branch at 920cdcb. Before the edits, every new test fails except `TestPlay_ChooserPrintsTheSourceCard/(d)_the_trigger_order_prints_as_before`, which pins the unchanged case. With both edits: `gofmt -l cmd pkg` clean; `go test ./cmd/cardtown` ok; `go tool task check` exit 0; `go tool task sim-compare -- autoloop/m2-wave-f.2` `[PASS]`; `cardtown lint fences` exit 0; `go tool task sim` exit 0 (200 games, P0 63.5% / P1 35.0% / draws 1.5%).

## Commits
1. `test(T-0235): the play screen shows a decision's source card; the stack names each card` with only the new test file below. Before committing, run `go -C W test ./cmd/cardtown -run TestPlay_ChooserPrintsTheSourceCard` and record that it fails (every subtest but `(d)`).
2. `feat(cardtown): play prompts show their source card and text; the stack names each card (T-0235)` with Edits 1-2.

## Tests (lead decisions; copy exactly, tabs for indentation)
### New file `cmd/cardtown/play_source_test.go`
```go
package main

import (
	"bufio"
	"bytes"
	"slices"
	"strings"
	"testing"

	"animus-mundi/pkg/engine"
	"animus-mundi/pkg/view"
)

// Tests for T-0235 (directive #149 (c)): the play screen prints a decision's
// source card and its rules text under the prompt, and the stack listing names
// each item's card.

func TestPlay_ChooserPrintsTheSourceCard(t *testing.T) {
	captain := &engine.Card{Name: "Devilhorn Captain", RulesText: "Unbowed: Deal 1 damage to chosen asset in the Main Theater."}
	volley := &engine.Card{Name: "Disorienting Pulse", RulesText: "Return chosen opposing unit in the Main Theater to its owner's hand."}
	ambush := &engine.Card{Name: "Ambush Tactics", RulesText: "Choose two:\n- Kill chosen unit.\n- Return chosen unit to its owner's hand."}
	apostle := &engine.Card{Name: "Infested Apostle", RulesText: "When this unit deploys, you may kill an allied unit. If you do, deal 2 damage to chosen opposing unit."}
	remnant := &engine.Card{Name: "Eerie Remnant", RulesText: "Deploys Exhausted\nWhen this card is put into your discard pile from your hand, you may deploy it from your discard pile for free."}
	vanilla := &engine.Card{Name: "Plain Unit"}
	two := []engine.Option{{Label: "Foe#3"}, {Label: "Ally#4"}}
	for _, tc := range []struct {
		name  string
		d     engine.Decision
		input string
		picks []int
		want  string // the whole output
	}{
		{
			"(a) a trigger's target",
			engine.Decision{Kind: engine.DecideTarget, Prompt: "Devilhorn Captain: trigger 1: choose target 1", Options: two, Min: 1, Max: 1, Source: captain},
			"2\n", []int{1},
			"Devilhorn Captain: trigger 1: choose target 1\n  [Devilhorn Captain] Unbowed: Deal 1 damage to chosen asset in the Main Theater.\n  1) Foe#3\n  2) Ally#4\nChoose option (1-N): ",
		},
		{
			"(b) a Maneuver's target",
			engine.Decision{Kind: engine.DecideTarget, Prompt: "Disorienting Pulse: choose target 1", Options: two, Min: 1, Max: 1, Source: volley},
			"1\n", []int{0},
			"Disorienting Pulse: choose target 1\n  [Disorienting Pulse] Return chosen opposing unit in the Main Theater to its owner's hand.\n  1) Foe#3\n  2) Ally#4\nChoose option (1-N): ",
		},
		{
			"(b) a mode choice, text of several lines",
			engine.Decision{Kind: engine.DecideMode, Prompt: "Ambush Tactics: choose 2 modes", Options: []engine.Option{{Label: "mode 1"}, {Label: "mode 2"}}, Min: 2, Max: 2, Source: ambush},
			"1 2\n", []int{0, 1},
			"Ambush Tactics: choose 2 modes\n  [Ambush Tactics] Choose two:\n    - Kill chosen unit.\n    - Return chosen unit to its owner's hand.\n  1) mode 1\n  2) mode 2\nChoose 2-2 options (space-separated): ",
		},
		{
			"(b) an optional trigger's you may",
			engine.Decision{Kind: engine.DecideOptional, Prompt: "Infested Apostle: trigger 1: you may", Min: 0, Max: 1, Source: apostle},
			"y\n", []int{0},
			"Infested Apostle: trigger 1: you may\n  [Infested Apostle] When this unit deploys, you may kill an allied unit. If you do, deal 2 damage to chosen opposing unit.\nChoose [y/n]: ",
		},
		{
			"(c) a trigger from a discard pile",
			engine.Decision{Kind: engine.DecideOptional, Prompt: "Eerie Remnant: trigger 1: you may", Min: 0, Max: 1, Source: remnant},
			"n\n", []int{},
			"Eerie Remnant: trigger 1: you may\n  [Eerie Remnant] Deploys Exhausted\n    When this card is put into your discard pile from your hand, you may deploy it from your discard pile for free.\nChoose [y/n]: ",
		},
		{
			"a source without rules text",
			engine.Decision{Kind: engine.DecideTarget, Prompt: "Plain Unit: choose target 1", Options: two, Min: 1, Max: 1, Source: vanilla},
			"1\n", []int{0},
			"Plain Unit: choose target 1\n  [Plain Unit]\n  1) Foe#3\n  2) Ally#4\nChoose option (1-N): ",
		},
		{
			"(d) the trigger order prints as before",
			engine.Decision{Kind: engine.DecideOrder, Prompt: "order your triggers", Options: []engine.Option{{Label: "a"}, {Label: "b"}}, Min: 2, Max: 2},
			"2 1\n", []int{1, 0},
			"order your triggers\n  1) a\n  2) b\nOrder all 2 options (space-separated): ",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var out bytes.Buffer
			a := terminalChooser(bufio.NewScanner(strings.NewReader(tc.input)), &out)(nil, tc.d)
			if !slices.Equal(a.Picks, tc.picks) {
				t.Errorf("picks %v, want %v", a.Picks, tc.picks)
			}
			if got := out.String(); got != tc.want {
				t.Errorf("output:\n%q\nwant:\n%q", got, tc.want)
			}
		})
	}
}

// A scripted session prints, under the prompt of a real decision with a source,
// that card's line (#149 (a)-(c)): the prompt names the card, and the line
// under it starts with the same name in brackets.
func TestPlay_ScriptedSessionPrintsSourceLines(t *testing.T) {
	args := []string{"-deck", "starter", "-p0", "ANIMUS", "-p1", "PATH", "-seat", "0", "-seed", "3"}
	var out bytes.Buffer
	if code := runPlay(args, &scriptedReader{out: &out}, &out); code != 0 {
		t.Fatalf("runPlay exited with code %d", code)
	}
	lines := strings.Split(out.String(), "\n")
	sources := 0
	for i := 1; i < len(lines); i++ {
		rest, ok := strings.CutPrefix(lines[i], "  [")
		name, _, closed := strings.Cut(rest, "]")
		if ok && closed && strings.Contains(lines[i-1], name+": ") {
			sources++
		}
	}
	if sources == 0 {
		t.Errorf("no prompt is followed by its source card's line")
	}
}

func TestPlay_PrintScreenNamesEachStackItemsCard(t *testing.T) {
	v := view.View{Stack: []view.StackView{
		{Kind: "Ability", Controller: 1, Source: 12, Card: "Devilhorn Captain"},
		{Kind: "Ability", Controller: 0, Source: 7, Card: "Ripper Pack", Keyword: "Aggressive"},
		{Kind: "Ability", Controller: 0, Source: 5},
		{Kind: "Deploy", Controller: 0, Card: "Chainbreaker Irregular", Zone: "Reserves"},
		{Kind: "Maneuver", Controller: 0, Card: "Disorienting Pulse"},
		{Kind: "March", Controller: 1, Source: 9, Zone: "Reserves"},
		{Kind: "Attack", Controller: 1, Source: 7, Target: 3},
		{Kind: "Scratch", Controller: 0, Source: 1},
	}}
	var buf bytes.Buffer
	printScreen(&buf, v, nil)
	want := "Resolution Stack:\n" +
		"  Ability of Devilhorn Captain (Player 1, Source 12)\n" +
		"  Aggressive trigger of Ripper Pack (Player 0, Source 7)\n" +
		"  Ability (Player 0, Source 5)\n" +
		"  Deploy of Chainbreaker Irregular to Reserves (Player 0)\n" +
		"  Maneuver: Disorienting Pulse (Player 0)\n" +
		"  March (Player 1, Source 9)\n" +
		"  Attack (Player 1, Source 7)\n" +
		"  Scratch (Player 0, Source 1)\n"
	if !strings.Contains(buf.String(), want) {
		t.Errorf("stack section:\n%s\nwant:\n%s", buf.String(), want)
	}
}
```

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)
Line numbers are at 920cdcb.

### 1. `cmd/cardtown/play_input.go`: the source line
In `terminalChooser`, line 43 reads
```go
		fmt.Fprintln(out, d.Prompt)
```
Insert right after it (before the blank line and `if d.Kind == engine.DecideOptional {`):
```go
		if d.Source != nil {
			fmt.Fprintln(out, sourceLine(d.Source))
		}
```
At the end of the file, after `terminalChooser`'s closing `}`, add one blank line and:
```go
// sourceLine is what terminalChooser prints under the prompt of a decision with
// a source card (Decision.Source, #149): the card's name and its printed rules
// text, each further line of the text indented under the first.
func sourceLine(c *engine.Card) string {
	text := strings.ReplaceAll(strings.TrimSpace(c.RulesText), "\n", "\n    ")
	if text == "" {
		return "  [" + c.Name + "]"
	}
	return "  [" + c.Name + "] " + text
}
```

### 2. `cmd/cardtown/play_screen.go`: the stack line
Lines 82-88, which read
```go
		for _, item := range v.Stack {
			if item.Kind == "Maneuver" {
				fmt.Fprintf(out, "  Maneuver: %s (Player %d)\n", item.Card, item.Controller)
			} else {
				fmt.Fprintf(out, "  %s (Player %d, Source %d)\n", item.Kind, item.Controller, item.Source)
			}
		}
```
become
```go
		for _, item := range v.Stack {
			fmt.Fprintf(out, "  %s\n", stackLine(item))
		}
```
Right before `func formatAsset(a view.AssetView) string {`, insert this block followed by one blank line:
```go
// stackLine is the stack listing's line for item. An ability names its card as
// the engine's place event does ("ability of", or "<keyword> trigger of"), a
// Deploy its card and zone, and a Maneuver its card (#149); a March, an Attack,
// a Scratch, and an ability with no card name their source asset.
func stackLine(item view.StackView) string {
	switch {
	case item.Kind == "Maneuver":
		return fmt.Sprintf("Maneuver: %s (Player %d)", item.Card, item.Controller)
	case item.Kind == "Deploy" && item.Card != "":
		return fmt.Sprintf("Deploy of %s to %s (Player %d)", item.Card, item.Zone, item.Controller)
	case item.Kind == "Ability" && item.Card != "":
		what := "Ability of " + item.Card
		if item.Keyword != "" {
			what = item.Keyword + " trigger of " + item.Card
		}
		return fmt.Sprintf("%s (Player %d, Source %d)", what, item.Controller, item.Source)
	}
	return fmt.Sprintf("%s (Player %d, Source %d)", item.Kind, item.Controller, item.Source)
}
```

No other file changes.

## Acceptance Cases (harness tasks only; written by the lead)
W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **The tests failed first:** at the `test(T-0235)` commit, before Edit 1, `go -C W test ./cmd/cardtown -run TestPlay_ChooserPrintsTheSourceCard` fails: every subtest but `(d)_the_trigger_order_prints_as_before` fails, and `(d)` passes.
2. **A prompt shows its source card and text** (#149 (a)-(c)): `go -C W test ./cmd/cardtown -run TestPlay_ChooserPrintsTheSourceCard` passes: a trigger's target (*Devilhorn Captain*), a Maneuver's target, a mode choice whose text has several lines, an optional trigger's "you may" (*Infested Apostle*), a trigger from a discard pile (*Eerie Remnant*) and a card with no rules text each print the source line under the prompt, before the options or the y/n question.
3. **A decision without a source is unchanged** (#149 (d)): the `(d)` subtest of case 2's run passes: the trigger order prints exactly as before.
4. **Real decisions carry their source to the screen:** `go -C W test ./cmd/cardtown -run TestPlay_ScriptedSessionPrintsSourceLines` passes: the seed-3 scripted session prints a source line under a prompt that names the same card.
5. **The stack names each card** (#149 (e)): `go -C W test ./cmd/cardtown -run TestPlay_PrintScreen` passes: an ability, a keyword trigger, a Deploy and a Maneuver name their cards; an ability with no card, a March, an Attack and a Scratch read as before; `TestPlay_PrintScreenManeuverStackLine` still passes.
6. **No hidden information** (#149 (f)): `go -C W test ./cmd/cardtown -run TestPlay_Case4_NoHiddenInformation` passes, and `go -C W test ./pkg/view/...` passes (the `viewtest` leak tests).
7. **Sims byte-identical:** `go -C W tool task sim-compare -- autoloop/m2-wave-f.2` prints `[PASS] starter and stress sims match autoloop/m2-wave-f.2 byte for byte, exit codes included`.
8. **Only the listed edits:** `git -C W diff autoloop/m2-wave-f.2 --numstat -- pkg cmd` prints exactly three lines: `14	0	cmd/cardtown/play_input.go`, `21	5	cmd/cardtown/play_screen.go` and `137	0	cmd/cardtown/play_source_test.go`.
9. **Fences:** `go -C W run ./cmd/cardtown lint fences` exits 0.
10. **Nothing else changed:** `git -C W diff autoloop/m2-wave-f.2 --stat` lists only the three files of case 8 and this task file.
11. `gofmt -l W/cmd` prints nothing; `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `cmd/cardtown/play_input.go`
- `cmd/cardtown/play_screen.go`
- `cmd/cardtown/play_source_test.go` (new; the `test(T-0235)` commit only)

## Done When
- [ ] Cases 1-11 hold, with their commands and results in the Log
- [ ] `go tool task sim-compare -- autoloop/m2-wave-f.2` passes (starter and stress sims byte-identical to the run branch)
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-14, lead-architect (wave F.2 run, session 1): queued stub, split off T-0233 under D2.4. Follows T-0234.
- 2026-11-14, lead-architect (wave F.2 run, session 3): written out in full against 920cdcb for #149 (c). Nothing on `main` to sync (`git log autoloop/m2-wave-f.2..main` is empty). Harness, so it skips `tests` and goes straight to `ready`. Tier `flash`, because both edits and the test file are exact and were dry-run. Changes from the stub: the tests go in a new file, `cmd/cardtown/play_source_test.go`, rather than `play_test.go`, so the test commit adds one file and touches no earlier test; the claims are the two new helpers. The chooser tests build `engine.Card` literals with the printed texts of the day rather than loading `data/cards`, so #152's change to *Devilhorn Captain*'s text doesn't touch them. Every command here is on the headless allow-list: one per line, no pipes, no `$`, no unquoted parentheses. Waiting directives checked: #6, #7 (mode `serial`), #15 (no turn-boundary hook), #71, #91 (2), #106, #120, #141 (3), #142: not fired. Next: `kernel-engineer` on Flash.
