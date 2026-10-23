---
id: T-0046
title: One hidden-information oracle for the leak tests, in pkg/view/viewtest (directive #39, harness, flash)
kind: harness
tier: flash
wave: none (board directive #39, structural review 2 cleanup 2 of 5, before the next shape task)
status: done
branch: m2/T-0046-viewtest-oracle
depends_on: [T-0045]
claims: [pkg:view/viewtest, func:viewtest.PublicNames, func:viewtest.HiddenNames]
cards: []
requirements: []
---

## Goal
Both hidden-information leak tests use one oracle, a small test-support package `pkg/view/viewtest` (like `net/http/httptest`):
- T-0036's leak test in `pkg/view/view_test.go`;
- T-0037's case 4 in `cmd/cardtown/play_test.go`.

When waves B–G change what is public (`RevealHand` for Drifting Concept, `GainControl`), the oracle changes in one place. Test behavior is unchanged.

## Inputs
- Board directive #39 (structural review 2, finding 3).
- The two copies:
  - [view_test.go:502-538](../../../pkg/view/view_test.go): `publicNames` and `hiddenNames`, with doc comments; callers at lines 548 and 561;
  - [play_test.go:185-216](../../../cmd/cardtown/play_test.go): the same bodies as closures inside `TestPlay_Case4_NoHiddenInformation`; callers at lines 223–224.
- Rulings that apply: none.

### Lead decisions
1. **The package:** `pkg/view/viewtest/viewtest.go`, `package viewtest`, importing only `animus-mundi/pkg/engine` and `slices`.
   - Package doc: `// Package viewtest is the hidden-information oracle for leak tests: which card names are public, and which a seat hides, in a game's current state. It is test support, like net/http/httptest, and imports only engine.`
   - `func PublicNames(g *engine.Game) map[string]bool` and `func HiddenNames(g *engine.Game, seat engine.PlayerID) map[string]bool`: the bodies and doc comments of `view_test.go`'s `publicNames` and `hiddenNames`, verbatim apart from the names. Rename `hiddenNames`'s parameter `q` to `seat` in the signature, body and comment.
2. **Authorized edits to existing tests, in the `test(T-0046)` commit:**
   - **`pkg/view/view_test.go`:** delete `publicNames` and `hiddenNames` (with their comments). Call `viewtest.PublicNames(g)` and `viewtest.HiddenNames(g, seat.Opponent())` at the two call sites. Import `animus-mundi/pkg/view/viewtest`, and drop `slices` only if nothing else in the file uses it.
   - **`cmd/cardtown/play_test.go`:** in `TestPlay_Case4_NoHiddenInformation`, delete the `publicNames` and `hiddenNames` closures. Call `viewtest.PublicNames(g)` and `viewtest.HiddenNames(g, botSeat)` in `checkText`. Import the package, and drop `slices` only if nothing else uses it.
   - No other test changes. The tests fail to compile until the package exists, which counts as failing first.
3. **Fence table:** add `"pkg/view/viewtest": {"pkg/engine"}` to `allowedImports` in `pkg/toolchain/linter/fences/fences.go`, so the new package's imports are checked.

## Acceptance Cases (harness tasks only; written by the lead)
These are checks recorded in the Log, and the reviewer re-runs them. No new tests.
1. `go test -list . ./pkg/view ./cmd/cardtown` gives the same test names before and after, and all of them pass.
2. `git grep -n -E 'func publicNames|publicNames :=|func hiddenNames|hiddenNames :=' -- pkg cmd` finds nothing. `go doc ./pkg/view/viewtest` prints the package doc and both functions.
3. **The oracle still catches leaks.** As a temporary, uncommitted change, print the bot's hand after each screen in `play.go`: `TestPlay_Case4_NoHiddenInformation` fails with `head leaks hidden card`. Revert it, and record the first failing lines in the Log.
4. `cardtown lint fences` passes on the real tree, and the new import-table row is present.
5. `go tool task check` passes.

## Allowed Files
- new: `pkg/view/viewtest/viewtest.go`
- `pkg/toolchain/linter/fences/fences.go` (the one table row)
- Tests: `pkg/view/view_test.go` and `cmd/cardtown/play_test.go` (lead decision 2's edits only, in the `test(T-0046)` commit)

## Done When
- [x] Lead decision 2's edits are committed as `test(T-0046): …` before the package exists
- [x] Cases 1–4 hold, with their commands and results in the Log
- [x] `go tool task check` passes

## Log
- 2026-10-22 lead-architect: Wrote the task from board directive #39. It is a harness task, tier flash, as the directive sets. T-0045 is done, so the task is `ready` on branch `m2/T-0046-viewtest-oracle`. Board directive #38's priority rule puts it before the next shape task.
  - Next: **the user runs `@kernel-engineer T-0046` in the Antigravity agent (Gemini Flash)**.
- 2026-10-22 kernel-engineer: Implemented directive #39 (T-0046).
  - Committed lead decision 2 test changes in `pkg/view/view_test.go` and `cmd/cardtown/play_test.go` as `test(T-0046): view_test and play_test use viewtest oracle` (fa06b95), removing the duplicate helper functions and closures, and calling `viewtest.PublicNames` and `viewtest.HiddenNames`. Verified that tests failed before implementation (`package animus-mundi/pkg/view/viewtest is not in std`).
  - Added `pkg/view/viewtest/viewtest.go` exporting `PublicNames(g *engine.Game) map[string]bool` and `HiddenNames(g *engine.Game, seat engine.PlayerID) map[string]bool` with package doc comment, verbatim bodies from `view_test.go`, and parameter `q` renamed to `seat`.
  - Added `"pkg/view/viewtest": {"pkg/engine"}` to `allowedImports` in `pkg/toolchain/linter/fences/fences.go`.
  - Reused: bodies of `publicNames` and `hiddenNames` from `pkg/view/view_test.go`.
  - Committed implementation as `feat(view): one hidden-information oracle in pkg/view/viewtest (T-0046)` (af6a842).
  - Verified acceptance cases:
    1. `go test -list . ./pkg/view ./cmd/cardtown` lists the identical 6 tests in `pkg/view` and 6 tests in `cmd/cardtown` before and after; all tests pass (`go test ./pkg/view ./cmd/cardtown`: ok in both packages).
    2. `git grep -n -E 'func publicNames|publicNames :=|func hiddenNames|hiddenNames :=' -- pkg cmd` found 0 matches. `go doc ./pkg/view/viewtest` prints package doc, `func HiddenNames(g *engine.Game, seat engine.PlayerID) map[string]bool`, and `func PublicNames(g *engine.Game) map[string]bool`.
    3. Oracle leak detection verified: temporarily printed bot's hand in `cmd/cardtown/play.go` after `printScreen`: `TestPlay_Case4_NoHiddenInformation` failed with:
       ```
       --- FAIL: TestPlay_Case4_NoHiddenInformation (0.04s)
           play_test.go:197: turn 1: head leaks hidden card "Macabre Dancers" of bot seat
           play_test.go:197: turn 1: head leaks hidden card "Chainbreaker Phalanx" of bot seat
           play_test.go:197: turn 1: head leaks hidden card "Devilhorn Sanction" of bot seat
       ```
       Reverted temporary leak via `git checkout cmd/cardtown/play.go`.
    4. `go run ./cmd/cardtown lint fences` passed (`[PASS] Package code is free of card names, invalid markers and disallowed imports.`) with table entry present.
    5. `go tool task check` passed cleanly (fmt-check, vet, test, build, lint all passed).
  - Status moved to `review`.
  - Next: lead-architect review.
- 2026-10-22 lead-architect: Review: **approved**, merged as 445e615, status `done`.
  - Scope:
    - c0e53f8 and 18cabfd change only this file.
    - fa06b95 changes only the two test files: the helpers and closures are deleted, the calls swapped and the import added, exactly as lead decision 2 specified.
    - af6a842 adds `viewtest.go` and the one `fences.go` row.
  - `PublicNames` and `HiddenNames` are verbatim against `view_test.go`'s originals (diffed), with `q` renamed to `seat`.
  - Gates, re-run in a scratch worktree:
    - At fa06b95, vet fails on the missing package.
    - At the head, `go tool task check` exits 0, and the 12 tests in `pkg/view` and `cmd/cardtown` pass.
    - The lead's leak mutation (the bot's hand printed after each screen) gives 2927 `head leaks` failures, the same count as before the change, so the oracle catches what it caught before.
  - Board directive #39 is complete.
