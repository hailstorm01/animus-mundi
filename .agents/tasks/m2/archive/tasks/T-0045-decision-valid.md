---
id: T-0045
title: Export Decision.Valid and use it in cardtown play's terminal chooser (directive #38, harness, sonnet)
kind: harness
tier: sonnet
wave: none (board directive #38, structural review 2 cleanup 1 of 5, before the next shape task)
status: done
branch: m2/T-0045-decision-valid
depends_on: []
claims: [func:Decision.Valid]
cards: []
requirements: []
---

## Goal
A client checks an answer with the kernel's own rule, `Decision.Valid`, instead of a copy. `cardtown play`'s terminal chooser then accepts exactly what the kernel accepts. So a decision kind added in waves B–G can't get a human answer that the kernel silently replaces with its fallback.

## Inputs
- Board directive #38 (structural review 2, finding 1).
- `Decision.valid` at [choice.go](../../../pkg/engine/choice.go): the rules are count within `Min`..`Max`, each index in range, and no repeat unless `Repeats`. It is called once, in `choose` (choice.go:113).
- The copy: `terminalChooser` in [play.go](../../../cmd/cardtown/play.go), the duplicate check and the count check (around lines 443–462 at the integration branch when the task was written).
- ADR-0003 (the Chooser contract): the lead adds a line at merge.
- Rulings that apply: `FR-MOD-001` (unchanged behavior).

### Lead decisions
1. **Rename, don't wrap.** `func (d Decision) valid(picks []int) bool` becomes `func (d Decision) Valid(picks []int) bool`, with the same body. Update the one caller in `choose`. Doc comment: `// Valid reports whether picks is a legal answer to d: Min to Max picks, each an index into Options, with no repeats unless Repeats (FR-MOD-001). The kernel applies the same rule to every Chooser answer.`
2. **The terminal chooser:**
   - It keeps its input parsing and messages: `q` quits, `y`/`n` for `DecideOptional`, a non-number or out-of-range number re-prompts with its current message, and an empty line means no picks.
   - It removes its own duplicate and count checks, and calls `d.Valid(picks)` instead.
   - On false, it re-prompts with one line: `Invalid choice: pick <Min> to <Max> option(s)` plus `, no repeats` when `!d.Repeats`, for example `Invalid choice: pick 2 to 2 option(s), no repeats`.
   - The empty line goes through `Valid` too, so it is accepted exactly when `Min == 0`. That replaces the separate "selection required" message.
   - The `y`/`n` branch for `DecideOptional` is unchanged.
3. **Not in this task:** any other `play.go` change (#41 and #42 come later), any change to `Decision`'s fields, and any change to the fallback rules.

## Acceptance Cases (harness tasks only; written by the lead)
Cases 1 and 2 are new tests, committed as `test(T-0045): …` before the implementation. They fail to compile until the implementation exists, which counts. Cases 3 and 4 are checks recorded in the Log against a baseline CLI built from `rework/executable-rules` before any edit; the reviewer re-runs them.
1. **`Decision.Valid`**, a table test in a new `pkg/engine/choice_valid_test.go` (`package engine`), on a decision with 3 options:
   - `Min 1, Max 2`: `[]` → false; `[0]` → true; `[0 2]` → true; `[0 1 2]` → false; `[3]` → false; `[-1]` → false.
   - `Min 2, Max 2`, `Repeats` false: `[1 1]` → false.
   - `Min 2, Max 2`, `Repeats` true: `[1 1]` → true.
   - `Min 0, Max 1`: `[]` → true.
2. **The terminal chooser**, a new test in `cmd/cardtown/play_test.go` (a new function only; no existing test changes). Call `terminalChooser(scanner, &out)(nil, d)` on a `DecideTarget` decision with 3 options, `Min 2, Max 2`, `Repeats` false, fed the input lines `1`, `1 1` and `1 2`:
   - it returns `Answer{Picks: [0 1]}`;
   - the output contains `Invalid choice: pick 2 to 2 option(s), no repeats` exactly twice.
   A second sub-test uses `Min 0, Max 2` and an empty line, and returns no picks with no `Invalid` line.
3. **Outputs unchanged:** against the baseline binary, these are byte-identical:
   - `cardtown sim -deck starter -p0 ANIMUS -p1 PATH` under both policies;
   - the case 1 scripted `play` session (`-deck starter -p0 ANIMUS -p1 PATH -seat 0 -seed 3`, with `1` to every prompt and `y` to every yes/no). It never gives an invalid answer, so the message change doesn't show.
4. `git grep -n 'seen\[p\]\|hasDup\|len(picks) < d.Min' -- cmd/cardtown/play.go` finds nothing. `git grep -n '\.valid(' -- pkg cmd` finds nothing.
5. `go tool task check` passes, and every existing test is unchanged.

## Allowed Files
- `pkg/engine/choice.go` (the rename, its doc comment and the one call site)
- `cmd/cardtown/play.go` (`terminalChooser` only)
- Tests: new `pkg/engine/choice_valid_test.go`; `cmd/cardtown/play_test.go` (one new test function)

## Done When
- [ ] Cases 1 and 2 have tests, committed before the implementation, and they failed before it
- [ ] Cases 3 and 4 hold, with their commands and results in the Log
- [ ] `go tool task check` passes

## Log
- 2026-10-22 lead-architect: Wrote the task from board directive #38. No starter-deck card can be built from `done` shapes: all 30 remaining cards are in waves B–G. So, under the user's ordering (#38: the cleanup goes before the next shape task, and cards first where possible), the cleanup starts now. It is a harness task, tier sonnet, as the directive sets. The task is `ready` on branch `m2/T-0045-decision-valid`.
  - Next: the lead dispatches the `kernel-engineer` subagent (Sonnet).
- 2026-10-22 kernel-engineer: Implemented. Commits: 8f9bd5d (in-progress), 46c4856 (tests for cases 1 and 2), then the feat commit.
  - Reuse search: `Decision.valid` was the only existing rule; renamed to `Valid` (body unchanged, doc comment per Lead decision 1), one caller in `choose` updated. No new shape.
  - `terminalChooser` (play.go): the empty-line special case, the duplicate check and the count check are removed. Input parsing, `q`, the out-of-range message and the `y`/`n` branch are unchanged. Picks now start as `[]int{}`, so an empty line goes through `d.Valid`; false prints `Invalid choice: pick <Min> to <Max> option(s)` plus `, no repeats` when `!d.Repeats`.
  - Case 1/2 failed before implementation: `go vet` reported `d.Valid undefined`; `go test ./cmd/cardtown -run TerminalChooserUses` failed on the old messages.
  - Case 3: baseline built from 851b9e6 before any edit (`ct45-base.exe`). `sim -deck starter -p0 ANIMUS -p1 PATH -policy progress` and `-policy random`: exit 0 on both, `cmp` identical. `play -deck starter -p0 ANIMUS -p1 PATH -seat 0 -seed 3 < 2000 lines of "1"`: both exit 0, 441692 bytes, `cmp` identical, 0 "Invalid" lines, reaches `Result: Player 1 won`.
  - Case 4: both `git grep` commands print nothing.
  - Case 5: `go tool task check` exits 0. No existing test changed (play_test.go gained one function and the `bufio` import).
  - Reviewer note: the scripted play feeds `1` to all prompts including y/n; it still completes, so I did not need a separate y answer.
- 2026-10-22 lead-architect: Review: **approved**, merged as f823737, status `done`.
  - Scope: 8f9bd5d changes only the status line. 46c4856 adds only `choice_valid_test.go` and one new `play_test.go` function (plus its import). a110aae changes only the two Allowed Files and this file. No test changed after 46c4856.
  - Design: it matches the Lead decisions. `valid` is renamed to `Valid` with the given comment, and its one caller is updated. The chooser keeps its parse messages, and asks `d.Valid` for the count, range and repeat rules with the specified message. An empty line still returns `Picks: []int{}`.
  - Gates, re-run in a scratch worktree:
    - At 46c4856, vet reports `d.Valid undefined`.
    - At the head, `go tool task check` exits 0.
    - Against a baseline built from 46c4856 (same non-test code as the integration branch), these are byte-identical (`cmp`): the starter sim under both policies, and a full scripted `play -seat 0 -seed 3` game fed `1` lines (ending `Player 1 won (Headquarters destroyed)`, exit 0).
  - The lead added the `Decision.Valid` line to ADR-0003's Decision. Board directive #38 is applied, and its priority rule now waits on #39, #41, #42 and #44.
