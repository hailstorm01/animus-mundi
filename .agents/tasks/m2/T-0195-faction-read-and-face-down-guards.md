---
id: T-0195
title: One read of "this faction or Neutral", and no face-down guard that can't fire (directive #132 (c), (d), harness, flash, pure move)
kind: harness
tier: flash
wave: D.2
status: review
branch: m2/autoloop-wave-d.2/T-0195-faction-read-and-face-down-guards
depends_on: [T-0194]
claims: [engine:Card-InFactionOrNeutral, engine:fireHook-face-down-guard, engine:replacementsFor-face-down-guard]
cards: []
requirements: []
---

## Goal
"This faction or Neutral" is one `Card` method, read by `ValidateDeck` and `FactionDeck`, and the two face-down guards that can never fire are gone (#132 (c), (d)). A pure move: same test names, same sim output. This is the last of #132; T-0194 took (a) and (b).

## Inputs
- Board directive #132 (structural review 9, approved by the user, 2026-11-06), split under D2.4 in T-0194.
- FRD 01 `FR-SET-001` (lines 124-134): a main deck holds only cards of its Headquarters' faction and Neutral cards.
- Design note: `docs/architecture/m2-ability-model.md`, Deliverable 6h.
- Why the method sits in `deck.go`, not beside the other `Card` methods in `card.go`: it is `FR-SET-001`'s rule, and it keeps this task at four non-test files (D2.4).
- Why `deck.go:52` stays: "a Neutral Headquarters is illegal" is another rule, and it doesn't ask this question. `sim.go:37` (a faction's Headquarters) stays for the same reason: it must not match Neutral.
- Why the guards can't fire (#132 (d)): `Printed()` (`pkg/engine/state.go:84`) returns `scratchFace` for a face-down asset. `scratchFace` (`pkg/engine/card.go:134`) sets no `Ability`, and the only non-test assignment to a card's `Ability` is `cardload.go:75`, on cards loaded from the sheets. So `a.Printed().Ability` is already nil there. `TestFR_ZON_004_FaceDownScratchCompanyFiresNoDeathTrigger` and the `TestCard_*_FaceDownCopy*` tests keep pinning the trigger side.
- The comments stay as written: `fireHook`'s "A face-down asset has no printed abilities" and `replacementsFor`'s "printed by a face-up asset" still say what the code does, and now say why no guard is needed.
- Reuse search (D2.3): no `Card` method reads a faction (searched `func (c *Card)` and `.Faction` in `pkg` and `cmd`).
- The lead dry-ran these exact edits in a throwaway worktree of the run branch at a0393e7: `gofmt -l pkg cmd` clean, `go tool task check` and `go tool task sim` pass, and `go tool task sim-compare -- autoloop/m2-wave-d.2` printed `[PASS]`. Every expected output below is from that dry run.

## Acceptance Cases (harness tasks only; written by the lead)
No new tests: a pure move, so no `test(T-0195)` commit. W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **Existing tests unchanged** (#132): `git -C W diff autoloop/m2-wave-d.2 -- '*_test.go'` prints nothing, and `go -C W test ./...` passes.
2. **Sims byte-identical** (#132): `go -C W tool task sim-compare -- autoloop/m2-wave-d.2` prints `[PASS]`.
3. **One read of the rule** (#132 (c)): `git -C W grep -n -e 'InFactionOrNeutral' -- pkg cmd` prints exactly four lines: `pkg/engine/deck.go:36` (the comment), `:38` (the method), `:81` (`ValidateDeck`) and `pkg/sim/sim.go:39` (`FactionDeck`).
4. **Neutral is spelled only where #132 keeps it** (#132 (c)): `git -C W grep -n -e 'NEUTRAL' -- pkg/engine/deck.go pkg/sim/sim.go` prints exactly two lines: `pkg/engine/deck.go:39` (the method) and `pkg/engine/deck.go:58` (a Neutral Headquarters is illegal).
5. **No guard is left** (#132 (d)): `git -C W grep -n -e 'FaceDown' -- pkg/engine/trigger.go pkg/engine/replace.go` prints nothing. At the run branch it prints two lines, so an exit code of 1 here is the pass.
6. **Only the listed edits:** `git -C W diff autoloop/m2-wave-d.2 --numstat -- pkg cmd` prints exactly four lines: `7	1	pkg/engine/deck.go`, `1	1	pkg/engine/replace.go`, `0	3	pkg/engine/trigger.go` and `1	2	pkg/sim/sim.go`.
7. **Nothing else changed:** `git -C W diff autoloop/m2-wave-d.2 --stat` lists only those four files and this task file.
8. `go -C W tool task check` and `go -C W tool task sim` pass.

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)
Line numbers are at a0393e7, before any edit.

### 1. `pkg/engine/deck.go`
After `illegalDeck` (after line 34), add a blank line and:
```go
// InFactionOrNeutral reports whether the card belongs to faction or is Neutral:
// the cards a main deck of that faction may hold (FR-SET-001).
func (c *Card) InFactionOrNeutral(faction string) bool {
	return strings.EqualFold(c.Faction, faction) || strings.EqualFold(c.Faction, "NEUTRAL")
}
```
Replace line 75 with:
```go
		if hqFaction != "" && !c.InFactionOrNeutral(hqFaction) {
```

### 2. `pkg/sim/sim.go`
Replace lines 39-40 with:
```go
		case c.Supported() && !c.IsScratchCompany() && c.InFactionOrNeutral(faction):
```

### 3. `pkg/engine/trigger.go`
In `fireHook`, delete lines 82-84 (`if a.FaceDown {`, `return`, `}`). Leave the comment above it as it is.

### 4. `pkg/engine/replace.go`
Replace line 39 with:
```go
		if src.Printed().Ability == nil {
```
Leave the comment above `replacementsFor` as it is.

No other file changes.

## Allowed Files
- `pkg/engine/deck.go`
- `pkg/sim/sim.go`
- `pkg/engine/trigger.go`
- `pkg/engine/replace.go`

## Done When
- [x] Cases 1-8 hold, with their commands and results in the Log
- [x] `go tool task sim-compare -- autoloop/m2-wave-d.2` passes (starter and stress sims byte-identical to the run branch)
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-07, lead-architect (wave D.2 run, session 8): queued with T-0194, which split #132 under D2.4. The next session takes this task from the board.
- 2026-11-07, lead-architect (wave D.2 run, session 9): wrote this task out from board directive #132 (c), (d) against the run branch at a0393e7. No `main` commits to sync. Not `WAVE-COMPLETE`: #132-#135 are `open: before the next Deliverables task`. Harness, so it skips `tests` and goes straight to `ready`; tier `flash`, because the edits are exact and were dry-run. `claims:` now names the method. Every command here is on the headless allow-list, one per line, with no pipes, `$`, double quotes or unquoted parentheses. Waiting directives checked: none has triggered (#6, #7: mode is `serial`; #15: no turn-boundary hook; #71, #91 (2), #106, #118 (b), #120, #128: no such card). Next: `kernel-engineer` on Flash.
- 2026-11-07, kernel-engineer (wave D.2 run, session 9): implemented T-0195 pure move (#132 (c), (d)).
  - What changed:
    - `pkg/engine/deck.go`: added `(*Card).InFactionOrNeutral(faction string) bool` helper; updated `ValidateDeck` to use it instead of open-coded string comparisons.
    - `pkg/sim/sim.go`: updated `FactionDeck` to use `c.InFactionOrNeutral(faction)`.
    - `pkg/engine/trigger.go`: removed unused `if a.FaceDown { return }` guard in `fireHook` (unreachable because face-down cards return `scratchFace` whose `Printed().Ability` is nil).
    - `pkg/engine/replace.go`: simplified guard in `replacementsFor` to `if src.Printed().Ability == nil { continue }`, removing redundant `src.FaceDown` check.
  - Reuse search: confirmed no prior `Card` method reading faction exists in `pkg` or `cmd`.
  - Commands run and results:
    1. `git -C W diff autoloop/m2-wave-d.2 -- '*_test.go'`: no diff; `go -C W test ./...`: pass (all packages passed).
    2. `go -C W tool task sim-compare -- autoloop/m2-wave-d.2`: `[PASS] starter and stress sims match autoloop/m2-wave-d.2 byte for byte, exit codes included`.
    3. `git -C W grep -n -e 'InFactionOrNeutral' -- pkg cmd`: printed exactly 4 lines (`pkg/engine/deck.go:36`, `:38`, `:81`, and `pkg/sim/sim.go:39`).
    4. `git -C W grep -n -e 'NEUTRAL' -- pkg/engine/deck.go pkg/sim/sim.go`: printed exactly 2 lines (`pkg/engine/deck.go:39`, `:58`).
    5. `git -C W grep -n -e 'FaceDown' -- pkg/engine/trigger.go pkg/engine/replace.go`: printed nothing (exit code 1).
    6. `git -C W diff autoloop/m2-wave-d.2 --numstat -- pkg cmd`: printed exactly 4 lines (7 1 pkg/engine/deck.go, 1 1 pkg/engine/replace.go, 0 3 pkg/engine/trigger.go, 1 2 pkg/sim/sim.go).
    7. `git -C W diff autoloop/m2-wave-d.2 --stat`: only those 4 files plus task file.
    8. `go -C W tool task check`: all formatting, vet, unit tests, and linters passed; `go -C W tool task sim`: passed cleanly (200 games, P0 78.5% / P1 19.0% / draws 2.5%, all invariants held).
  - Reviewer notice: pure move; sim outputs and tests byte-identical to run branch.
