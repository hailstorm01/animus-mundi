---
id: T-0153
title: The linters share one scan, pkg/toolchain/specdoc (directive #100, harness, flash, pure move)
kind: harness
tier: flash
wave: C.1 (board directive #100, structural review 7, finding 9)
status: done
branch: m2/autoloop-wave-c.1/T-0153-specdoc
depends_on: []
claims: [package:pkg/toolchain/specdoc, helper:ReqRegex, helper:DefLineRegex, helper:ReadLines]
cards: []
requirements: []
---

## Goal
`reqRegex`, `defLineRegex` and `readLines` are each written in `frd.go` and `rules.go`. A new package, `pkg/toolchain/specdoc`, holds them, and both linters call it (#100). The fences linter's identical `readLines` folds in too. A pure move: same behavior, same test names, same lint output.

## Inputs
- Board directive #100 (structural review 7, approved by the user, 2026-11-03).
- Design note: `docs/architecture/m2-ability-model.md`, Boundaries (`pkg/toolchain` row) and Deliverable 6f (wave C.1).
- **Beyond #100's text, on purpose:** `pkg/toolchain/linter/fences/fences.go:470-485` is a third copy of `readLines`, the same lines as `rules.go`'s. The reuse search found it, and keeping it would be a duplicate this task's plan keeps (D2.3), so this task removes it.
- **Left out, on purpose:**
  - `pkg/toolchain/linter/tdd/tdd.go` and `pkg/toolchain/tddslice` keep their copies, because #100 says so. Directive #103 deletes both.
  - Each linter keeps its own line-shape regexes (`subsumedRow`, `testNameRegex`, `coversRegex` in `rules.go`, `ownerRegex` in `ownership.go`, `markerRegex` in `fences.go`). #100 says each linter keeps its own subsumption rules, and these match different lines with different captures.
- No test names any of the moved declarations (searched `reqRegex`, `defLineRegex` and `readLines` in `pkg/` and `cmd/`), and no test file changes.
- `specdoc` gets no tests of its own: `frd_test.go`, `rules_test.go` and the fences tests already run it through each linter.
- `go tool task sim-compare` doesn't apply: nothing under `pkg/toolchain` runs in a sim. #100's check is the lint output. `cardtown lint frd` prints the repository path, and a byte-for-byte comparison needs a built binary and redirects, which aren't on the headless allow-list. So the engineer checks the counts and verdict lines below, and the reviewer compares the bytes.
- The lead dry-ran these exact edits in a throwaway worktree of the run branch at 4c2677b: `gofmt -l` clean, `go tool task check` passes, `go tool task sim` runs clean, and `cardtown lint frd`, `lint rules` and `lint fences` print the same bytes with the same exit codes as a build of the run branch, run from the same directory.

### Where it is today (at `autoloop/m2-wave-c.1`, 4c2677b)
| Where | Today |
| :--- | :--- |
| `pkg/toolchain/linter/frd/frd.go:4`, `:8` | the `bufio` and `regexp` imports, which only these declarations use |
| `pkg/toolchain/linter/frd/frd.go:14-16` | `reqRegex`, `defLineRegex` and a blank line, at the top of the `var` block |
| `pkg/toolchain/linter/frd/frd.go:101`, `:109`, `:110`, `:130`, `:138`, `:180` | the uses |
| `pkg/toolchain/linter/frd/frd.go:204-221` | a blank line and `readLines`, at the end of the file |
| `pkg/toolchain/linter/rules/rules.go:4`, `:6` | the `bufio` and `os` imports, which only `readLines` uses |
| `pkg/toolchain/linter/rules/rules.go:14-15` | `reqRegex` and `defLineRegex` |
| `pkg/toolchain/linter/rules/rules.go:53`, `:62`, `:127`, `:136`, `:137` | the uses |
| `pkg/toolchain/linter/rules/rules.go:144-160` | a blank line and `readLines`, at the end of the file |
| `pkg/toolchain/linter/fences/fences.go:4` | the `bufio` import, which only `readLines` uses (`os` stays) |
| `pkg/toolchain/linter/fences/fences.go:19` | the `cardsheet` import |
| `pkg/toolchain/linter/fences/fences.go:227` | the use |
| `pkg/toolchain/linter/fences/fences.go:469-485` | a blank line and `readLines`, at the end of the file |

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)

### 1. `pkg/toolchain/specdoc/specdoc.go` (new file)
```go
// Package specdoc holds what the toolchain linters share when they scan a file: requirement IDs, the lines that define them, and reading a file as lines.
package specdoc

import (
	"bufio"
	"os"
	"regexp"
)

var (
	// ReqRegex matches a requirement ID.
	ReqRegex = regexp.MustCompile(`FR-[A-Z]{3}-\d{3}[a-z]?`)
	// DefLineRegex matches a line that can define a requirement: a heading, or a line, bullet or table row that opens with a bold ID.
	DefLineRegex = regexp.MustCompile(`^(####|###|\*\*FR-|- \*\*FR-|\| \*\*FR-)`)
)

// ReadLines returns the lines of the file at path.
func ReadLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	// Accommodate very long lines if needed (up to 1MB)
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, 1024*1024)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}
```

### 2. `pkg/toolchain/linter/frd/frd.go`
- In the import block, delete the lines `"bufio"` and `"regexp"`. Below the line `"strings"`, insert one blank line and then this line:
  ```go
  	"animus-mundi/pkg/toolchain/specdoc"
  ```
- In the `var` block, delete the `reqRegex` line, the `defLineRegex` line and the blank line below them, so the block opens with `ExpectedFiles = []string{`.
- Replace all three `readLines(` with `specdoc.ReadLines(`, both `reqRegex.` with `specdoc.ReqRegex.`, and the one `defLineRegex.` with `specdoc.DefLineRegex.`.
- Delete the function `readLines` and the blank line above it, so the file ends with the closing brace of `Lint`.

### 3. `pkg/toolchain/linter/rules/rules.go`
- In the import block, delete the lines `"bufio"` and `"os"`. Below the line `"strings"`, insert one blank line and then this line:
  ```go
  	"animus-mundi/pkg/toolchain/specdoc"
  ```
- In the `var` block, delete the `reqRegex` and `defLineRegex` lines. The other lines stay as they are, alignment included.
- Replace both `readLines(` with `specdoc.ReadLines(`, both `reqRegex.` with `specdoc.ReqRegex.`, and the one `defLineRegex.` with `specdoc.DefLineRegex.`.
- Delete the function `readLines` and the blank line above it, so the file ends with the closing brace of `frdRequirements`.

### 4. `pkg/toolchain/linter/fences/fences.go`
- In the import block, delete the line `"bufio"`. Below the line `"animus-mundi/pkg/toolchain/linter/cardsheet"`, insert this line:
  ```go
  	"animus-mundi/pkg/toolchain/specdoc"
  ```
- Replace the one `readLines(` with `specdoc.ReadLines(`.
- Delete the function `readLines` and the blank line above it, so the file ends with the closing brace of the function above it.

No other file changes.

## Acceptance Cases (harness tasks only; written by the lead)
No new tests: a pure move. W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **Existing tests unchanged** (#100): `git -C W diff autoloop/m2-wave-c.1 -- '*_test.go'` prints nothing, and `go -C W test ./pkg/toolchain/... ./cmd/...` passes.
2. **`lint frd` unchanged** (#100): `go -C W run ./cmd/cardtown lint frd` exits 0 and prints `Discovered 203 defined requirement IDs across 7 FRD files.`, `Checking 2191 requirement references across 7 target FRD file(s)...` and `[PASS] Zero dangling references! All requirement cross-references resolve cleanly.`
3. **`lint rules` unchanged** (#100): `go -C W run ./cmd/cardtown lint rules` exits 0 and prints `51 of 62 active FRD 01 requirements have a TestFR_* scenario test.` and `[PASS] Every scenario test traces to an active FRD 01 requirement.`
4. **`lint fences` unchanged:** `go -C W run ./cmd/cardtown lint fences` exits 0 and prints `[PASS] Package code is free of card names, invalid markers and disallowed imports.`
5. **One definition of each** (#100): each of these prints exactly one line, in `pkg/toolchain/specdoc/specdoc.go`:
   - `git -C W grep -n 'ReqRegex = ' -- pkg` prints line 12;
   - `git -C W grep -n 'DefLineRegex = ' -- pkg` prints line 14;
   - `git -C W grep -n 'func ReadLines' -- pkg` prints line 18.
6. **No copy left in the three linters:** `git -C W grep -n -e 'reqRegex' -e 'defLineRegex' -e 'readLines' -- pkg/toolchain/linter/frd pkg/toolchain/linter/rules pkg/toolchain/linter/fences` prints nothing. `git grep` exits 1 when it finds nothing, which is the expected result here.
7. **The TDD copy stays** (#100): `git -C W grep -n 'func readLines' -- pkg cmd` prints exactly one line, `pkg/toolchain/linter/tdd/tdd.go:621`.
8. **Every use goes through `specdoc`:** `git -C W grep -c -F 'specdoc.' -- pkg/toolchain/linter` prints exactly three lines: `pkg/toolchain/linter/fences/fences.go:1`, `pkg/toolchain/linter/frd/frd.go:6` and `pkg/toolchain/linter/rules/rules.go:5`.
9. **Only the listed edits:** `git -C W diff autoloop/m2-wave-c.1 --numstat -- pkg cmd` prints exactly these four lines: `2	19	pkg/toolchain/linter/fences/fences.go`, `8	29	pkg/toolchain/linter/frd/frd.go`, `7	26	pkg/toolchain/linter/rules/rules.go` and `34	0	pkg/toolchain/specdoc/specdoc.go`.
10. **Nothing else changed:** `git -C W diff autoloop/m2-wave-c.1 --stat` lists only those four files and this task file.
11. `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/toolchain/specdoc/specdoc.go` (new)
- `pkg/toolchain/linter/frd/frd.go`
- `pkg/toolchain/linter/rules/rules.go`
- `pkg/toolchain/linter/fences/fences.go`

## Done When
- [x] Cases 1-11 hold, with their commands and results in the Log
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-03, lead-architect (wave C.1 run, session 10): wrote this task against the run branch at 4c2677b; `main` had no commits to sync. Harness, so it skips `tests` and goes straight to `ready`; tier `flash`, because the edits are exact and were dry-run. Every command here is on the headless allow-list, one per line, with no pipes, `$`, double quotes or unquoted parentheses. Placement: the design note's Boundaries row for `pkg/toolchain` names `specdoc`. Reuse search: no shared scan exists under `pkg/toolchain`; `linter/cardsheet` reads CSV sheets only. The search found a third `readLines` in `fences.go`, which this task removes (Inputs). Waiting directives checked: none has triggered (#6, #7: mode is `serial`; #8: no task touches `deck.go`; #15: no turn-boundary hook; #71, #91 (2): no such card). Next: `kernel-engineer` on Flash.
- 2026-11-03, kernel-engineer (wave C.1 run, session 10): implemented directive #100 pure move.
  - Diff: created `pkg/toolchain/specdoc/specdoc.go` holding `ReqRegex`, `DefLineRegex` and `ReadLines`. Updated `pkg/toolchain/linter/frd/frd.go`, `pkg/toolchain/linter/rules/rules.go` and `pkg/toolchain/linter/fences/fences.go` to import `animus-mundi/pkg/toolchain/specdoc` and call `specdoc.ReqRegex`, `specdoc.DefLineRegex` and `specdoc.ReadLines`. Removed local regexes, `readLines` implementations, and unused imports (`bufio`, `regexp` from `frd.go`, `bufio`, `os` from `rules.go`, `bufio` from `fences.go`).
  - Reuse: `specdoc` consolidates the three duplicate regex and line-reading implementations across the linters.
  - Commands and results:
    - Case 1: `git -C W diff autoloop/m2-wave-c.1 -- '*_test.go'` printed nothing; `go -C W test ./pkg/toolchain/... ./cmd/...` passed.
    - Case 2: `go -C W run ./cmd/cardtown lint frd` exited 0 and printed expected catalog counts (203 defined IDs across 7 FRDs, 2191 references) with zero dangling references PASS.
    - Case 3: `go -C W run ./cmd/cardtown lint rules` exited 0 and printed 51 of 62 active FRD 01 requirements have scenario tests with full traceability PASS.
    - Case 4: `go -C W run ./cmd/cardtown lint fences` exited 0 and printed clean PASS.
    - Case 5: `git -C W grep -n 'ReqRegex = ' -- pkg` printed `pkg/toolchain/linter/tdd/tdd.go:18: headingReqRegex = ...` (kept per #100) and `pkg/toolchain/specdoc/specdoc.go:12: ReqRegex = ...`; `git -C W grep -n 'DefLineRegex = ' -- pkg` printed `pkg/toolchain/specdoc/specdoc.go:14`; `git -C W grep -n 'func ReadLines' -- pkg` printed `pkg/toolchain/specdoc/specdoc.go:18`.
    - Case 6: `git -C W grep -n -e 'reqRegex' -e 'defLineRegex' -e 'readLines' -- pkg/toolchain/linter/frd pkg/toolchain/linter/rules pkg/toolchain/linter/fences` printed nothing (exit 1).
    - Case 7: `git -C W grep -n 'func readLines' -- pkg cmd` printed `pkg/toolchain/linter/tdd/tdd.go:621` (preserved per #100).
    - Case 8: `git -C W grep -c -F 'specdoc.' -- pkg/toolchain/linter` printed `fences.go:1`, `frd.go:6`, `rules.go:5`.
    - Case 9: `git -C W diff autoloop/m2-wave-c.1 --numstat -- pkg cmd` printed exactly:
      2 19 pkg/toolchain/linter/fences/fences.go
      8 29 pkg/toolchain/linter/frd/frd.go
      7 26 pkg/toolchain/linter/rules/rules.go
      34 0 pkg/toolchain/specdoc/specdoc.go
    - Case 10: `git -C W diff autoloop/m2-wave-c.1 --stat` listed only those four toolchain files and the task file.
    - Case 11: `go -C W tool task check` passed all stages (fmt-check, vet, test, build, lint --all); `go -C W tool task sim` ran clean on starter decks (200 matches, all invariants held).
  - Reviewer note: `git grep -n 'ReqRegex = '` matched `headingReqRegex = ` in `tdd.go` due to substring match; `tdd.go` was kept intact per directive #100.
- 2026-11-03, lead-architect (wave C.1 run, session 10, D5): **Approved, `done`.** Merged into `autoloop/m2-wave-c.1` as e2c4313. Flash run `f79ec8f5-4a3e-4ca6-969b-45ad447b6aa1`: exit 0, `status: SUCCESS`, 406 s, no refusals (no `denied_actions` key; no `permission check failed` match; the transcript's one step with `"status":"ERROR"` is an edit whose target text did not match, which the engineer redid; the `.err` file holds only agy's line about waiting on a background task). Review of `git diff autoloop/m2-wave-c.1...m2/autoloop-wave-c.1/T-0153-specdoc` (97f7031):
  - Scope: only the four Allowed Files and this file changed. No test file changed.
  - Pure move, checked against the diff: `specdoc.go` holds the two regexes and `readLines` as they were, and the linters' added lines are only the import and the renamed calls. The numstat matches the lead's dry run.
  - **A slip in this task, not in the code:** case 5 said `git grep -n 'ReqRegex = ' -- pkg` prints one line, but it also matches `headingReqRegex = ` at `pkg/toolchain/linter/tdd/tdd.go:18`, a different regex that #103 deletes. The engineer reported this rather than changing anything. The reviewer checked with `git grep -n -w 'ReqRegex' -- pkg/toolchain/specdoc`: one definition, at line 12.
  - Fences and placement: no card name, no new rule. `specdoc` sits where the design note's Boundaries row puts it, and imports only the standard library.
  - Re-run by the reviewer in a scratch worktree of the task branch, all as expected: cases 1-11 (case 5 as above); `gofmt -l` clean; `go tool task check` passes; `go tool task sim` clean, 200 games, P0 82.5% / P1 17.5% / draws 0.0%, 42/50 and 32/50 main cards executable; `go tool task sim-compare -- autoloop/m2-wave-c.1` `[PASS]`, though this task doesn't require it.
  - #100's check: a build of the task branch and a build of 4c2677b, run from the same directory, print the same bytes with the same exit codes for `lint frd`, `lint rules` and `lint fences` (all exit 0), and for the error path `lint frd nosuch.md` (exit 1).
  - The merge's `pkg/` and `cmd/` match the task tip's, so the gates hold on the run branch.
