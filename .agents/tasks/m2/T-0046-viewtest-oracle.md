---
id: T-0046
title: One hidden-information oracle for the leak tests, in pkg/view/viewtest (directive #39, harness, flash)
kind: harness
tier: flash
wave: none (board directive #39, structural review 2 cleanup 2 of 5, before the next shape task)
status: ready
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
- [ ] Lead decision 2's edits are committed as `test(T-0046): …` before the package exists
- [ ] Cases 1–4 hold, with their commands and results in the Log
- [ ] `go tool task check` passes

## Log
- 2026-10-22 lead-architect: Wrote the task from board directive #39. It is a harness task, tier flash, as the directive sets. T-0045 is done, so the task is `ready` on branch `m2/T-0046-viewtest-oracle`. Board directive #38's priority rule puts it before the next shape task.
  - Next: **the user runs `@kernel-engineer T-0046` in the Antigravity agent (Gemini Flash)**.
