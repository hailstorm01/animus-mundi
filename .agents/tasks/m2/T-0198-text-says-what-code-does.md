---
id: T-0198
title: The text says what the code does (directive #135, harness, flash)
kind: harness
tier: flash
wave: D.2
status: review
branch: m2/autoloop-wave-d.2/T-0198-text-says-what-code-does
depends_on: [T-0190]
claims: []
cards: []
requirements: []
---

## Goal
`cardtown`'s help and its sim header stop calling every playout "random-policy", because the default seat policy is `progress` (`cmd/cardtown/sim.go:27`). They say playouts, naming neither policy (#135 (b)). Nothing else changes: the sims print the same bytes except their header line.

## Inputs
- Board directive #135 (structural review 9, approved by the user, 2026-11-06), the last directive of wave D.2.
- **(a) needs no edit.** `pkg/engine/doc.go:96-97` says `sim` and `eventlog` import only `engine` from this module, and `view` imports `engine` and `eventlog`. `pkg/sim/sim.go:7-9` says `sim` imports `engine` only from this module. Both are true after #127 (T-0190): at 8e0e402, `go list -f '{{.Imports}}'` gives `animus-mundi/pkg/engine` as the only module import of `pkg/sim` and of `pkg/eventlog`, and `animus-mundi/pkg/engine` and `animus-mundi/pkg/eventlog` for `pkg/view`. #135 says to correct only what is still wrong, so neither file changes.
- **(b) the four printed lines,** at 8e0e402: `cmd/cardtown/main.go:68` (the `sim` line of `printUsage`), `cmd/cardtown/sim.go:32` (`cardtown sim -h`), `cmd/cardtown/sim.go:137` (the header every sim prints) and `pkg/toolchain/docs/whatsup.go:99` (#135 gave `:98`; the line moved by one). A search for `random-policy` (any case) in `cmd` and `pkg` finds only these and the doc comment `pkg/sim/sim.go:96` on `Simulate`, which does play `engine.RandomChooser` on both seats (`:111`), so it is true and stays. The prose in `docs/` and `.agents/` that says random-policy describes `go test ./pkg/engine`'s fuzz playouts or is outside the lead's files, and #135 doesn't name it.
- The header keeps its width: the new string is as long as the old one (61 characters, quotes included), padded with spaces, so the `=` rules above and below still frame it.
- Placement and reuse: wording only, no declaration added or moved. Nothing in `pkg/engine` changes.
- The lead dry-ran these exact edits in a throwaway worktree of the run branch at 8e0e402: `gofmt -l pkg cmd` clean, `go tool task check` passes, `go tool task sim` clean (P0 78.5% / P1 19.0% / draws 2.5%, 47/50 and 34/50 main cards executable). `go tool task sim-compare -- autoloop/m2-wave-d.2` exits non-zero and prints two diffs, each a single `@@ -1,5 +1,5 @@` hunk that changes only the header line, as case 4 says. Every expected output below is from that dry run. No test reads these strings (searched `Random-policy`, `RANDOM-POLICY` and `Playouts` in `*_test.go`), and #135's acceptance is the commands, so this task adds no test.

## Acceptance Cases (harness tasks only; written by the lead)
W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **No random-policy wording left in what `cardtown` prints:** `git -C W grep -n -i -e 'random-policy' -- cmd pkg/toolchain` prints nothing (exit code 1).
2. **The four lines, reworded:** `git -C W grep -n -e 'Playouts on the rules kernel' -e 'RULES KERNEL: PLAYOUTS' -- cmd pkg` prints exactly four lines: `cmd/cardtown/main.go:68`, `cmd/cardtown/sim.go:32`, `cmd/cardtown/sim.go:137` and `pkg/toolchain/docs/whatsup.go:99`.
3. **Only the listed edits:** `git -C W diff autoloop/m2-wave-d.2 --numstat -- pkg cmd` prints exactly three lines: `1	1	cmd/cardtown/main.go`, `2	2	cmd/cardtown/sim.go` and `1	1	pkg/toolchain/docs/whatsup.go`. `git -C W diff autoloop/m2-wave-d.2 --stat` lists only those three files and this task file.
4. **The sims change only their header:** `go -C W tool task sim-compare -- autoloop/m2-wave-d.2` exits non-zero and prints no `[PASS]` line. It prints two diffs, `base-starter.txt` against `new-starter.txt` and `base-stress.txt` against `new-stress.txt`. Each has one hunk, `@@ -1,5 +1,5 @@`, whose only `-` and `+` lines are `-  CARD TOWN RULES KERNEL: RANDOM-POLICY PLAYOUTS` and `+  CARD TOWN RULES KERNEL: PLAYOUTS` (each followed by its trailing spaces). This is the expected result (#135), not a failure: copy both diffs into the Log.
5. `go -C W tool task check` and `go -C W tool task sim` pass.

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)
Line numbers are at 8e0e402. Change only the text inside the quotes; each line keeps its indentation and its `fmt.Println(` call.

### 1. `cmd/cardtown/main.go`
Line 68: replace `"  sim       Random-policy playouts on the rules kernel (termination, invariants, win rates)"` with `"  sim       Playouts on the rules kernel (termination, invariants, win rates)"`.

### 2. `cmd/cardtown/sim.go`
**2a.** Line 32: replace `"Random-policy playouts on the rules kernel (termination, invariants, win rates)"` with `"Playouts on the rules kernel (termination, invariants, win rates)"`.

**2b.** Line 137: replace the string `"  CARD TOWN RULES KERNEL: RANDOM-POLICY PLAYOUTS           "` with `"  CARD TOWN RULES KERNEL: PLAYOUTS                         "`. The new string has two leading spaces, then `CARD TOWN RULES KERNEL: PLAYOUTS`, then 25 trailing spaces, so it is 61 characters long with its quotes, the same as the old one.

### 3. `pkg/toolchain/docs/whatsup.go`
Line 99: replace `"- cardtown sim [flags] : Random-policy playouts on the rules kernel; fails on non-termination or broken invariants"` with `"- cardtown sim [flags] : Playouts on the rules kernel; fails on non-termination or broken invariants"`.

No other file changes. Run `gofmt -l pkg cmd` after the edits; it must print nothing.

## Allowed Files
- `cmd/cardtown/main.go`
- `cmd/cardtown/sim.go`
- `pkg/toolchain/docs/whatsup.go`

## Done When
- [x] Cases 1-5 hold, with their commands and results in the Log
- [x] `go tool task sim-compare -- autoloop/m2-wave-d.2` shows only the header line changed, in the starter and the stress output (case 4)
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-07, lead-architect (wave D.2 run, session 12): wrote this task from board directive #135 against the run branch at 8e0e402. No `main` commits to sync. Not `WAVE-COMPLETE`: #135 is `open: before the next Deliverables task`. Harness, so it skips `tests` and goes straight to `ready`; tier `flash`, because the edits are exact and were dry-run. Three non-test files, within the D2.4 limit. Part (a) needs no edit (Inputs). Every command here is on the headless allow-list, one per line, with no pipes, `$`, double quotes or unquoted parentheses. Waiting directives checked: none has triggered (#6, #7: mode is `serial`; #15: no turn-boundary hook; #71, #91 (2), #106, #118 (b), #120, #128: no such card). Next: `kernel-engineer` on Flash.
- 2026-11-07, kernel-engineer (T-0198):
  - Status moved to in-progress (41f44ae), then review.
  - Implemented lead directive #135 edits across three allowed files: `cmd/cardtown/main.go`, `cmd/cardtown/sim.go`, and `pkg/toolchain/docs/whatsup.go`.
  - Replaced "Random-policy playouts" with "Playouts" in usage and whatsup strings.
  - In `cmd/cardtown/sim.go:137`, updated header banner string to `"  CARD TOWN RULES KERNEL: PLAYOUTS                         "` keeping 61 character width intact.
  - Verified `gofmt -l pkg cmd` prints nothing.
  - Verified Acceptance Cases 1-5:
    - Case 1: `git -C W grep -n -i -e 'random-policy' -- cmd pkg/toolchain` returned exit status 1 with no output.
    - Case 2: `git -C W grep -n -e 'Playouts on the rules kernel' -e 'RULES KERNEL: PLAYOUTS' -- cmd pkg` printed exactly four lines:
      `cmd/cardtown/main.go:68:	fmt.Println("  sim       Playouts on the rules kernel (termination, invariants, win rates)")`
      `cmd/cardtown/sim.go:32:		fmt.Println("Playouts on the rules kernel (termination, invariants, win rates)")`
      `cmd/cardtown/sim.go:137:	fmt.Println("  CARD TOWN RULES KERNEL: PLAYOUTS                         ")`
      `pkg/toolchain/docs/whatsup.go:99:	fmt.Println("- cardtown sim [flags] : Playouts on the rules kernel; fails on non-termination or broken invariants")`
    - Case 3: `git -C W diff autoloop/m2-wave-d.2 --numstat -- pkg cmd` printed:
      1	1	cmd/cardtown/main.go
      2	2	cmd/cardtown/sim.go
      1	1	pkg/toolchain/docs/whatsup.go
      `git -C W diff autoloop/m2-wave-d.2 --stat` listed only those three files and the task file.
    - Case 4: `go -C W tool task sim-compare -- autoloop/m2-wave-d.2` exited non-zero with no `[PASS]` line, outputting two single-hunk diffs modifying only the header line:
      ```diff
      diff --git "a/bin/sim-compare/base-starter.txt" "b/bin/sim-compare/new-starter.txt"
      index 60b2717..6b4a199 100644
      --- "a/bin/sim-compare/base-starter.txt"
      +++ "b/bin/sim-compare/new-starter.txt"
      @@ -1,5 +1,5 @@
       ==========================================================
      -  CARD TOWN RULES KERNEL: RANDOM-POLICY PLAYOUTS           
      +  CARD TOWN RULES KERNEL: PLAYOUTS                         
       ==========================================================
      ```
      ```diff
      diff --git "a/bin/sim-compare/base-stress.txt" "b/bin/sim-compare/new-stress.txt"
      index 801c884..dce2bb0 100644
      --- "a/bin/sim-compare/base-stress.txt"
      +++ "b/bin/sim-compare/new-stress.txt"
      @@ -1,5 +1,5 @@
       ==========================================================
      -  CARD TOWN RULES KERNEL: RANDOM-POLICY PLAYOUTS           
      +  CARD TOWN RULES KERNEL: PLAYOUTS                         
       ==========================================================
      ```
    - Case 5: `go -C W tool task check` passed all stages (gofmt, vet, test, linter --all). `go -C W tool task sim` passed clean (P0 78.5%, P1 19.0%, draws 2.5%).
  - Committed implementation as `feat(toolchain): the text says what the code does (T-0198)` (731e20e).
