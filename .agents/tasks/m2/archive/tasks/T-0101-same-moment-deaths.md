---
id: T-0101
title: A replacement sees every death at the same moment, in a fixed order, with no destruction partway through a sweep (FR-REP-001, SVY-R20, SVY-R21, SVY-R22, directive #74, wave G shape, sonnet)
kind: shape
tier: sonnet
wave: G
status: done
branch: m2/T-0101-same-moment-deaths
depends_on: [T-0097]
claims: [replacement:death-order, sweep:passes]
cards: []
requirements: [FR-REP-001, FR-LAY-002]
---

## Goal
The death sweep works in passes. In each pass, every death a replacement applies to is replaced, one at a time in a fixed order, and only then are the pass's other lethal assets destroyed together. Nothing is destroyed partway through a pass, so a replacement source that is lethal, or that a byproduct leaves lethal, keeps applying to every death of the pass. What byproducts newly left lethal dies in the next pass. No player is asked to order replaced deaths: the active player's go first, then the non-active player's, each player's oldest asset first. Kill steps use the same fixed order. This applies board directive #74 and changes merged T-0096 and T-0097 code.

## Inputs
- FRD 01 v3.16, `FR-REP-001` (lines 454-462): Per-Entity Death Replacement (line 460), Same-Moment Deaths (line 461, `SVY-R21`), Replacement Within a Death Sweep (line 462, `SVY-R20`, `SVY-R22`). `FR-LAY-002` (line 440, the sweep runs only when nothing is resolving).
- Survey §4 rows R5 (revised), R20, R21, R22 (lines 64, 79-81).
- Design Outcome: `docs/design/outcomes/2026-10-29-replacement-sees-same-moment-deaths.md`.
- Board directive #74.
- Kernel today:
  - `sweepDeaths` (`damage.go`) calls `replaceDeaths` with a `between` callback that runs an interim sweep after each replaced death, so a source a byproduct leaves lethal is destroyed before the next death.
  - `replacedQueue` (`replace.go`) asks each player with two or more queued deaths a `DecideOrder`.
  - `stepKill` (`effects.go`) calls `replaceDeaths` with an always-true `dies` and an empty `between`.

### Shape (lead decisions, binding)
- **The fixed order (`replace.go`):** `replacedQueue` keeps its signature and asks nothing. It returns the assets of `dying` that a replacement applies to, the active player's first, then the non-active player's, each player's sorted by asset ID (oldest first). Doc comment: "replacedQueue returns the assets of dying that a replacement applies to, in the order their deaths are replaced: the active player's first (SVY-R22), then the non-active player's, each player's oldest asset first (SVY-R20). No player orders them."
- **One loop, no callback (`replace.go`):** `replaceDeaths(dying []*Asset, dies func(*Asset) bool) []*Asset` loses the `between` parameter. Its body is otherwise unchanged: skip a queued asset that has left the battlefield or for which `dies` no longer holds, replace it if `replaceDeath` still finds a replacement, and return the rest in `dying`'s order. Doc comment: "replaceDeaths replaces the deaths of dying that a replacement applies to, one at a time in replacedQueue's order, with nothing destroyed between them: a source that is lethal, or that a byproduct leaves lethal, still applies (SVY-R20, SVY-R21). It returns the rest, to be destroyed together: the assets of dying still on the battlefield, for which dies still holds, whose death it did not replace."
- **Passes (`damage.go`):** `sweepDeaths` runs passes until a pass finds nothing: each pass takes `g.newlyLethal(seen)`, and if it is empty, stops; otherwise it calls `g.destroyDying(g.replaceDeaths(dying, g.IsLethal), loses, reasons)`. A destroyed Headquarters doesn't end the loop: the losses are applied after the sweep, as today. It takes `seen`, `loses` and `reasons` and no `dying`; `deathSweep` calls `g.sweepDeaths(seen, &loses, &reasons)`, or inlines the loop if that reads better. No recursion. Doc comment: "sweepDeaths takes up the lethal assets in passes. In each pass, the deaths a replacement applies to are replaced first (replaceDeaths), then the rest die together; what a byproduct newly left lethal dies in the next pass, a new moment (SVY-R20, FR-LAY-002). seen holds every asset the sweep has taken up, so a pass takes up only new ones." `deathSweep`'s doc comment drops "in their controller's order (SVY-R5)" and says "in a fixed order (SVY-R20, SVY-R22)".
- **The kill step (`effects.go`):** `stepKill` calls `g.replaceDeaths(dying, func(*Asset) bool { return true })`. In its doc comment, "in its controller's order" becomes "in a fixed order", and "(design note Open Question 13)" becomes "(SVY-R21)".
- **Comments:** `replace.go`'s file comment cites `FR-REP-001, SVY-R20 to SVY-R22` in place of `SVY-R5`. `findReplacement`'s comment stays.
- **Unchanged:** `findReplacement`, `replaceDeath`, `relocateReplaced`, `destroyDying`, `destroyAll`, `newlyLethal`. `DecideOrder` stays: triggers still use it.
- **Sims:** no executable card prints a replacement, and with no replacement a pass destroys exactly what one sweep destroys today. A second pass finds nothing, since nothing executable changes an asset's Health when another leaves. So the sims should stay byte-identical. If `sim-compare` differs, don't change code to make it match: report the first differing game in the Log and hand off.
- **Left for later:** the choice between several replacements for one death is T-0100 (planned against this task: no `DecideOrder` for replaced deaths). *Mutation Swell*'s scenario (four kills replaced, and two units still deployed from the discard pile, `SVY-R15`, `SVY-R21`) goes in wave F's *Mutation Swell* task, once T-0098 builds the Scratch Company filter.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
Authorized edits to protected tests from T-0096 and T-0097, all in the `test(T-0101)` commit. Synthetic cards only, reusing the helpers in `replacement_test.go` (`guardianStructure`, `guardianUnit`, `threeVictims`, `eventSeq`, `wantReplaced`, `sweep`) and `replacement_kill_test.go` (`killStep`, `killer`).

**Delete** (their outcomes are contradicted by `SVY-R20`; each is replaced by a new test below):
- `TestFR_REP_001_ControllerOrdersTheReplacedDeaths` and `TestFR_REP_001_ReplacedDeathOrderFallsBackToOldestFirst` (`replacement_test.go`) → `TestFR_REP_001_ReplacedDeathsAreNotOrderedByThePlayer`.
- `TestFR_REP_001_SourceLeftLethalByAByproductStopsApplying` (`replacement_test.go`) → `TestFR_REP_001_ByproductLethalSourceStillReplacesEveryDeathInThePass`.
- `TestFR_REP_001_TheAffectedAssetsControllerOrdersTheirDeaths` (`replacement_test.go`) → `TestFR_REP_001_ActivePlayersReplacedDeathsGoFirst`.
- `TestFR_REP_001_ControllerOrdersTheReplacedKills` (`replacement_kill_test.go`) → `TestFR_REP_001_ReplacedKillsAreNotOrderedByThePlayer`.

**Comment and message edits only** (no assertion changes): in both files, the header comment and the `wantReplaced` failure messages cite `SVY-R20`/`SVY-R21`/`SVY-R22` in place of `SVY-R5` and in place of "user ruling 2026-10-29 (not yet in FRD 01)"; the same in the comments and messages of `TestFR_REP_001_SourceLethalInTheSameSweepReplacesEveryDeath` (`SVY-R20`), `TestFR_REP_001_OneKillStepsDeathsAreAllReplaced` and `TestFR_REP_001_SourceKilledByTheSameStepReplacesTheOthersFirst` (`SVY-R21`).

**New tests:**
- `TestFR_REP_001_ReplacedDeathsAreNotOrderedByThePlayer` in `replacement_test.go` (`SVY-R20`): `threeVictims(s, 2)`, and player 0 is scripted with an order answer (`Picks: []int{2, 0, 1}`) that would reverse the order if asked. No `DecideOrder` is asked of player 0. All three deaths are replaced, oldest first (`wantReplaced` in ID order), and all three units are in their owner's hand. The guardian (2 Health, 3 damage taken) is destroyed, and its `destroy` event comes after the third unit's `leave` event.
- `TestFR_REP_001_ByproductLethalSourceStillReplacesEveryDeathInThePass` in `replacement_test.go` (`SVY-R20`, FRD 01's second example): a guardian structure with 4 Health and five allied 1-Health units, all lethal at one sweep. All five deaths are replaced, oldest first, and no unit is in the discard pile. The guardian (−1 Current Health) is destroyed in the next pass: its `destroy` event comes after the fifth unit's `leave` event.
- `TestFR_REP_001_ActivePlayersReplacedDeathsGoFirst` in `replacement_test.go` (`SVY-R22`): each player has a guardian structure with 3 Health and two allied 1-Health units, and player 1's units are put first, so they are older than player 0's. Player 0, the active player, deploys a sweeper that deals 1 damage to every other unit. All four deaths are replaced: player 0's two first, oldest first, then player 1's two, oldest first. Neither player is asked a `DecideOrder`.
- `TestFR_REP_001_ReplacedKillsAreNotOrderedByThePlayer` in `replacement_kill_test.go` (`SVY-R20`, `SVY-R21`): a guardian structure with 3 Health and two allied units, player 0 scripted with `Picks: []int{1, 0}`, one kill step kills both. No `DecideOrder` is asked. Both kills are replaced, oldest first.
- **Directive #74's `TestFR_REP_001_OneKillStepReplacesEveryKill`:** not written. `TestFR_REP_001_OneKillStepsDeathsAreAllReplaced` (T-0097) is the same scenario: a 1-Health source, one kill step killing two allied units, both replaced, the source destroyed at the sweep after the item.
- **Existing tests:** every other test in `pkg/engine` passes unchanged, including `TestFR_REP_001_SourceLethalInTheSameSweepReplacesEveryDeath` and `TestFR_REP_001_ByproductDamageFiresTheSourcesDamageTrigger`.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected. The test files may declare their own helpers.

## Allowed Files
- `pkg/engine/replace.go` (`replacedQueue`, `replaceDeaths`, the file comment)
- `pkg/engine/damage.go` (`sweepDeaths` in passes, `deathSweep`'s call and doc comment)
- `pkg/engine/effects.go` (`stepKill`'s call and doc comment)

## Done When
- [x] The listed tests pass: `go test ./pkg/engine -run 'TestFR_REP_001'`
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks
- [x] `go tool task sim-compare -- autoloop/m2-wave-g` passes (no executable card prints a replacement, so the sims are unchanged; see Shape, Sims)

## Log
- 2026-10-29 lead-architect (wave G run, session 3): Wrote the task for board directive #74 (`open`), the lead's call being a task of its own rather than folded into T-0097, which merged before #74 reached the run branch. It goes before T-0100, which now plans against it. Design note: Deliverable 6c lists this task and drops "in its controller's order" from T-0096's line, and Open Question 13 is rewritten as answered (this commit).
  - **Sync:** merged `main` (b4c26f8, the D7.2 conflict rule) into the run branch with no conflict (221da73).
  - **Directives:** read `BOARD.md` on `autoloop/m2-wave-g` fresh. #74 is applied by this task. #73 stays `open: before the first wave C task` (T-0098, T-0099 and T-0100 still to come). #14 is a standing order. Waiting triggers checked: #6 and #7 (mode is `serial`); #8 (no `deck.go` change); #15 (no turn hook); #20 (b) (no wave C shape task); #58 (no exhaust by a cost or an effect); #59 (2) (no repeated targets); #61 (no modal card); #71 (no effect on each player). None has fired.
  - **Contradicted tests:** directive #74 names four of T-0096's tests. T-0097's `TestFR_REP_001_ControllerOrdersTheReplacedKills` is contradicted the same way (it was written before #74), so it is named here too. Two pairs of the directive's rewrites would have produced the same scenario as a settling test, so the old tests are deleted and replaced, not rewritten in place.
  - **Reuse:** no duplicate. The replacement loop stays the one shared `replaceDeaths` (T-0097); this task removes its `between` callback, and the sweep's recursion becomes a loop of passes.
- 2026-10-29 rules-test-author: Tests written (commit on the task branch, `test(T-0101)`).
  - **New (fail until the engineer implements, as intended):** `TestFR_REP_001_ReplacedDeathsAreNotOrderedByThePlayer`, `TestFR_REP_001_ByproductLethalSourceStillReplacesEveryDeathInThePass`, `TestFR_REP_001_ActivePlayersReplacedDeathsGoFirst` (`replacement_test.go`); `TestFR_REP_001_ReplacedKillsAreNotOrderedByThePlayer` (`replacement_kill_test.go`). A small shared helper `noOrderAsked` (asserts no `DecideOrder` was asked of a player) is declared in `replacement_test.go`.
  - **Deleted (as authorized):** `ControllerOrdersTheReplacedDeaths`, `ReplacedDeathOrderFallsBackToOldestFirst`, `SourceLeftLethalByAByproductStopsApplying`, `TheAffectedAssetsControllerOrdersTheirDeaths`, `ControllerOrdersTheReplacedKills`.
  - **Comment/message edits only:** headers and `wantReplaced` messages now cite `SVY-R20`/`SVY-R21`/`SVY-R22`; the three named tests' ruling citations updated. No assertion changes.
  - **Not written:** `TestFR_REP_001_OneKillStepReplacesEveryKill` (same scenario as `OneKillStepsDeathsAreAllReplaced`, per the task).
  - **Helpers requested:** none. `go vet` is clean; `cardtown lint rules` passes; `go test -run TestFR_REP_001` fails only on the four new tests.
- 2026-10-29 lead-architect (D4): **Tests approved, `ready`.** Checked 7ff925d against FRD 01 v3.16 `FR-REP-001` (lines 460-462), `SVY-R20` to `SVY-R22` and the task's Tests list:
  - the five named tests are deleted, each replaced by its named new test, and the other edits are comments and messages only (no assertion changed);
  - `ReplacedDeathsAreNotOrderedByThePlayer`: all three replaced oldest first though the 2-Health guardian is lethal after the second byproduct, no `DecideOrder`, and the guardian is destroyed after the third `leave` (`SVY-R20`);
  - `ByproductLethalSourceStillReplacesEveryDeathInThePass`: FRD 01's second example, five replaced, then the guardian at -1 is destroyed in the next pass;
  - `ActivePlayersReplacedDeathsGoFirst`: player 0's (newer) units are replaced before player 1's (older), and neither player is asked (`SVY-R22`);
  - `ReplacedKillsAreNotOrderedByThePlayer`: the kill step uses the same fixed order (`SVY-R21`).
- 2026-10-29 kernel-engineer: Implemented; `status: review`.
  - **Diff:** `replace.go`: `replacedQueue` asks nothing (active player's first, then the other's, each oldest first); `replaceDeaths` loses `between`; file comment cites SVY-R20 to SVY-R22. `damage.go`: `sweepDeaths(seen, loses, reasons)` is a loop of passes (`newlyLethal`, then `destroyDying(replaceDeaths(dying, IsLethal))`); `deathSweep` calls it, doc comments updated. `effects.go`: `stepKill` calls the two-argument `replaceDeaths`; doc comment cites SVY-R21.
  - **Reuse:** no new shape; removed the `between` callback and recursion. `DecideOrder` untouched (triggers use it).
  - **Commands:** `go test ./pkg/engine -run TestFR_REP_001` ok; `go tool task check` passed; `go tool task sim` clean; `go tool task sim-compare -- autoloop/m2-wave-g` PASS (byte-identical).
- 2026-10-29 lead-architect (D5): **Approved, merged into `autoloop/m2-wave-g` as 5885f1c.**
  - **Scope:** c2c11f0 changes only `replace.go`, `damage.go` and `effects.go`; no test changed after 7ff925d.
  - **Fences and reuse:** no card name, no new shape; the one shared `replaceDeaths` loses its callback, and the sweep's recursion is a loop of passes that ends because `seen` only grows.
  - **Rules fidelity:** `replacedQueue` gives the active player's deaths first, each player's oldest first, with no prompt (`SVY-R20`, `SVY-R22`); nothing is destroyed within a pass, and newly lethal assets die in the next pass (`SVY-R20`, `FR-LAY-002`); kill steps use the same order (`SVY-R21`).
  - **Gates, re-run in a review worktree:** `go tool task check` passes; `go tool task sim` clean, 200 games, P0 63.5% / P1 34.5% / draws 2.0% (unchanged); `go tool task sim-compare -- autoloop/m2-wave-g` byte-identical.
