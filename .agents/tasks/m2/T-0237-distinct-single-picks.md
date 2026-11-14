---
id: T-0237
title: "A single-pick prompt lists identical options once (directive #151, wave F.2 harness, flash)"
kind: harness
tier: flash
wave: F.2
status: ready
branch: m2/autoloop-wave-f.2/T-0237-distinct-single-picks
depends_on: []
claims: [func:shownOptions]
cards: []
requirements: [FR-MOD-001]
---

## Goal
On `cardtown play`, a decision that picks at most one (`Max` 1) lists each option text once, numbered in first-seen order, and a pick answers the first option with that text, so a look at *Chainbreaker Irregular*, *Boneyard Amalgam*, *Boneyard Amalgam* no longer offers the same outcome twice (#151). A decision with `Max` above 1 lists every option, duplicates included. The engine still lists every outcome, so rules, sims and event records don't change. When this merges, #151 is `applied`.

## Inputs
- Board directive #151 (user playtest feedback, 2026-11-13), `open: before the next Deliverables task`. The user chose the display-only fix.
- Design (lead decisions):
  - **One helper.** `shownOptions(d engine.Decision) []int` (`cmd/cardtown/play_input.go`) returns the indices of the options `terminalChooser` lists, in order; listed number k+1 answers option `shown[k]`. With `d.Max == 1` it keeps the first option of each `Label`; otherwise every option. `terminalChooser` prints from it, bounds a typed number by `len(shown)` (the "Invalid choice" message names that count), and maps each typed number through it before `d.Valid`. The `DecideOptional` branch returns before the listing, so a y/n question doesn't change.
  - **Why equal labels are the same choice in M2.** Name-only labels come from the hand (`zones.go:101`), the deck search (`effects.go:129`), the discard-pile decisions (`targeting.go:144`) and the look's outcomes (`freedeploy.go:126`). Battlefield assets and replacements carry their ID in the label (`state.go:97`, `replace.go:126-133`); modes and zones are distinct. Every M2 pile pick reads "your discard pile" (`pkg/engine/cards`), so two equal names in one pick are copies of one card in one zone of one owner. A later card that picks from either player's discard pile would list two owners' copies under one name; that is a label question for its own task (the label already can't tell them apart today), not this one.
  - **Scripted sessions.** `scriptedReader` answers `1`, and `shown[0]` is always option 0, so the scripted `play` tests answer the same options as before; they don't pin option lists, and they pass unchanged in the dry run.
- Reuse search (D2.3): `cmd/cardtown` had no option de-duplication; `d.Valid` (`pkg/engine/choice.go:63`) still checks the mapped picks. Nothing copied.
- Size (D2.4): one non-test file.
- **Expected output difference:** none in `cardtown sim` (the play screen is not in sim output), so `go tool task sim-compare` passes. `cardtown play` lists identical single-pick options once.
- The lead dry-ran this exact edit and this exact test file in a throwaway worktree of the run branch at af77383. Before the edit, every new subtest fails except `(c)_discard_2_lists_both_copies` and `(d)_a_decision_with_no_duplicates_is_numbered_as_before`, which pin the unchanged cases. With the edit: `gofmt -l cmd pkg` clean; `go test ./cmd/cardtown ./pkg/view/...` ok; `go tool task check` exit 0; `go tool task sim-compare -- autoloop/m2-wave-f.2` `[PASS]`; `cardtown lint fences` exit 0; `go tool task sim` exit 0 (200 games, P0 63.5% / P1 35.0% / draws 1.5%). Blobs: `play_input.go` d725f6f, `play_distinct_test.go` 34e22e2.

## Commits
1. `test(T-0237): a single-pick prompt lists identical options once` with only the new test file below. Before committing, run `go -C W test ./cmd/cardtown -run TestPlay_ChooserListsIdenticalOptionsOnce` and record that it fails (every subtest but `(c)` and `(d)`).
2. `feat(cardtown): a single-pick prompt lists identical options once (T-0237)` with Edit 1.

## Tests (lead decisions; copy exactly, tabs for indentation)
### New file `cmd/cardtown/play_distinct_test.go`
```go
package main

import (
	"bufio"
	"bytes"
	"slices"
	"strings"
	"testing"

	"animus-mundi/pkg/engine"
)

// Tests for T-0237 (directive #151): a single-pick prompt lists identical
// options once, and a pick answers the first option with that text; a prompt
// that may pick more than one lists every option.

func TestPlay_ChooserListsIdenticalOptionsOnce(t *testing.T) {
	keep := "Deploy Chainbreaker Irregular to Reserves; bottom: Boneyard Amalgam, Boneyard Amalgam"
	swap := "Deploy Boneyard Amalgam to Reserves; bottom: Chainbreaker Irregular, Boneyard Amalgam"
	look := []engine.Option{{Label: keep}, {Label: keep}, {Label: swap}, {Label: swap}}
	search := []engine.Option{{Label: "Boneyard Amalgam"}, {Label: "Chainbreaker Irregular"}, {Label: "Boneyard Amalgam"}}
	hand := []engine.Option{{Label: "Boneyard Amalgam"}, {Label: "Boneyard Amalgam"}, {Label: "Chainbreaker Irregular"}}
	for _, tc := range []struct {
		name  string
		d     engine.Decision
		input string
		picks []int
		want  string // the whole output
	}{
		{
			"(a) a look's identical outcomes are listed once",
			engine.Decision{Kind: engine.DecideLook, Prompt: "Look at the top 3 cards of your deck", Options: look, Min: 1, Max: 1},
			"2\n", []int{2},
			"Look at the top 3 cards of your deck\n  1) " + keep + "\n  2) " + swap + "\nChoose option (1-N): ",
		},
		{
			"(a) picking the first listed answers the first option",
			engine.Decision{Kind: engine.DecideLook, Prompt: "Look at the top 3 cards of your deck", Options: look, Min: 1, Max: 1},
			"1\n", []int{0},
			"Look at the top 3 cards of your deck\n  1) " + keep + "\n  2) " + swap + "\nChoose option (1-N): ",
		},
		{
			"(b) a single-pick deck search lists two copies once",
			engine.Decision{Kind: engine.DecideSearch, Prompt: "Search your deck for a card", Options: search, Min: 1, Max: 1},
			"2\n", []int{1},
			"Search your deck for a card\n  1) Boneyard Amalgam\n  2) Chainbreaker Irregular\nChoose option (1-N): ",
		},
		{
			"(b) a number past the listed options is refused",
			engine.Decision{Kind: engine.DecideSearch, Prompt: "Search your deck for a card", Options: search, Min: 1, Max: 1},
			"3\n1\n", []int{0},
			"Search your deck for a card\n  1) Boneyard Amalgam\n  2) Chainbreaker Irregular\nChoose option (1-N): " +
				"Invalid choice \"3\": must be a number between 1 and 2\nChoose option (1-N): ",
		},
		{
			"(b) an up-to-one discard lists two copies once",
			engine.Decision{Kind: engine.DecideHand, Prompt: "Discard up to 1 card", Options: hand, Min: 0, Max: 1},
			"2\n", []int{2},
			"Discard up to 1 card\n  1) Boneyard Amalgam\n  2) Chainbreaker Irregular\nChoose up to 1 options (space-separated, empty for none): ",
		},
		{
			"(c) discard 2 lists both copies",
			engine.Decision{Kind: engine.DecideHand, Prompt: "Discard 2 cards", Options: hand, Min: 2, Max: 2},
			"1 2\n", []int{0, 1},
			"Discard 2 cards\n  1) Boneyard Amalgam\n  2) Boneyard Amalgam\n  3) Chainbreaker Irregular\nChoose 2-2 options (space-separated): ",
		},
		{
			"(d) a decision with no duplicates is numbered as before",
			engine.Decision{Kind: engine.DecideTarget, Prompt: "Devilhorn Captain: trigger 1: choose target 1", Options: []engine.Option{{Label: "Foe#3"}, {Label: "Ally#4"}, {Label: "Foe#5"}}, Min: 1, Max: 1},
			"3\n", []int{2},
			"Devilhorn Captain: trigger 1: choose target 1\n  1) Foe#3\n  2) Ally#4\n  3) Foe#5\nChoose option (1-N): ",
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
```

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)
Line numbers are at af77383.

### 1. `cmd/cardtown/play_input.go`: list each single-pick option text once
(a) In `terminalChooser`, lines 68-70, which read
```go
		for i, opt := range d.Options {
			fmt.Fprintf(out, "  %d) %s\n", i+1, opt.Label)
		}
```
become
```go
		shown := shownOptions(d)
		for k, i := range shown {
			fmt.Fprintf(out, "  %d) %s\n", k+1, d.Options[i].Label)
		}
```
(b) Lines 94-99, which read
```go
				if err != nil || idx < 1 || idx > len(d.Options) {
					fmt.Fprintf(out, "Invalid choice %q: must be a number between 1 and %d\n", p, len(d.Options))
					parsed = false
					break
				}
				picks = append(picks, idx-1)
```
become
```go
				if err != nil || idx < 1 || idx > len(shown) {
					fmt.Fprintf(out, "Invalid choice %q: must be a number between 1 and %d\n", p, len(shown))
					parsed = false
					break
				}
				picks = append(picks, shown[idx-1])
```
(c) At the end of the file, after `sourceLine`'s closing `}`, add one blank line and:
```go
// shownOptions returns the indices of d's options that terminalChooser lists,
// in order: the listed number k+1 answers option shownOptions(d)[k]. A decision
// that picks at most one lists each option text once, at its first option, so
// identical outcomes (two copies of a card) aren't offered twice (#151); any
// other decision lists every option, because picking two copies there is a
// real choice.
func shownOptions(d engine.Decision) []int {
	var shown []int
	seen := make(map[string]bool, len(d.Options))
	for i, opt := range d.Options {
		if d.Max == 1 {
			if seen[opt.Label] {
				continue
			}
			seen[opt.Label] = true
		}
		shown = append(shown, i)
	}
	return shown
}
```

No other file changes.

## Acceptance Cases (harness tasks only; written by the lead)
W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **The tests failed first:** at the `test(T-0237)` commit, before Edit 1, `go -C W test ./cmd/cardtown -run TestPlay_ChooserListsIdenticalOptionsOnce` fails: every subtest fails but `(c)_discard_2_lists_both_copies` and `(d)_a_decision_with_no_duplicates_is_numbered_as_before`, which pass.
2. **Identical single-pick options are listed once** (#151 (a), (b)): `go -C W test ./cmd/cardtown -run TestPlay_ChooserListsIdenticalOptionsOnce` passes: a `DecideLook` with two pairs of identical outcomes lists two, and each listed number answers the first option with its text; a single-pick deck search and an up-to-one discard list two copies once; a number past the listed options is refused with the listed count.
3. **A multi-pick decision lists every copy** (#151 (c)): the `(c)` subtest of case 2's run passes: "Discard 2 cards" lists both copies, and picking both answers both.
4. **No duplicates, no change** (#151 (d)): the `(d)` subtest of case 2's run passes, and `go -C W test ./cmd/cardtown -run TestPlay_` passes (the earlier play tests, the scripted sessions included, unchanged).
5. **Sims byte-identical:** `go -C W tool task sim-compare -- autoloop/m2-wave-f.2` prints `[PASS] starter and stress sims match autoloop/m2-wave-f.2 byte for byte, exit codes included`.
6. **Only the listed edits:** `git -C W diff autoloop/m2-wave-f.2 --numstat -- pkg cmd` prints exactly two lines: `27	5	cmd/cardtown/play_input.go` and `85	0	cmd/cardtown/play_distinct_test.go`.
7. **Fences:** `go -C W run ./cmd/cardtown lint fences` exits 0.
8. **Nothing else changed:** `git -C W diff autoloop/m2-wave-f.2 --stat` lists only the two files of case 6 and this task file.
9. `gofmt -l W/cmd` prints nothing; `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `cmd/cardtown/play_input.go`
- `cmd/cardtown/play_distinct_test.go` (new; the `test(T-0237)` commit only)

## Done When
- [ ] Cases 1-9 hold, with their commands and results in the Log
- [ ] `go tool task sim-compare -- autoloop/m2-wave-f.2` passes (starter and stress sims byte-identical to the run branch)
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-14, lead-architect (wave F.2 run, session 5): written in full against af77383 for #151. Nothing on `main` to sync (`git log autoloop/m2-wave-f.2..main` is empty); the scratch folder was empty. Not `WAVE-COMPLETE` before the task (#151 `open: before the next Deliverables task`, #152 `open`). Harness, so it skips `tests` and goes straight to `ready`. Tier `flash`, because the edit and the test file are exact and were dry-run. The tests go in a new file, so the test commit touches no earlier test. Every command here is on the headless allow-list: one per line, no pipes, no `$`, no unquoted parentheses. Waiting directives checked: #6, #7 (mode `serial`), #15 (no turn-boundary hook), #71, #91 (2), #106, #120, #141 (3), #142: not fired. Next: `kernel-engineer` on Flash.
