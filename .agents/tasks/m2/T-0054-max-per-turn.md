---
id: T-0054
title: A once-per-turn limit on triggers (Trigger.MaxPerTurn), for Animoceps Rager (wave B shape 3, sonnet)
kind: shape
tier: sonnet
wave: B
status: tests-review
branch: m2/T-0054-max-per-turn
depends_on: [T-0050, T-0052]
claims: [trigger:MaxPerTurn]
cards: []
requirements: [FR-TRN-001, FR-REP-004]
---

## Goal
A printed trigger can be limited to firing a fixed number of times each turn (`Trigger.MaxPerTurn`). That is the last shape the starter Headquarters *Animoceps Rager* needs: "Whenever this Headquarters is dealt damage, Scratch 1. This ability triggers only once each turn." The card itself follows as the flash card task T-0055.

This task also refreshes the stale "Not built yet" list in `pkg/engine/doc.go`, owed since T-0050 (T-0052 decision 5).

## Inputs
- FRD 01:
  - `FR-TRN-001` Ready Phase (line 225): "all once-per-turn ability limits reset for the new turn";
  - `FR-REP-004` Stage 5, the damage-taken threshold. The survey's §4 note records the default that a Headquarters "is dealt damage" uses the same threshold as Unbowed (`FR-KWD-002a`);
  - `FR-TRG-003`/`FR-TRG-004`, placement.
- Survey: wave B (§2), and the Animoceps Rager row (§6, line 106: `maxPerTurn: 1`).
- Design note: `docs/architecture/m2-ability-model.md`, Key types (`Trigger` gate conditions already name `MaxPerTurn`) and "Changes in `engine`".
- Existing machinery to reuse: `fireHook` in `trigger.go` (the gates sit there), `Game.Turn` in `state.go`, `HookDamaged` (T-0050), which already fires on any asset, Headquarters included.

### Lead decisions (technical)
1. `ability.Trigger` gains `MaxPerTurn int`. 0 (the zero value) means no limit, so every existing definition is unchanged.
2. **The limit counts firings**, that is, the times the trigger goes into the pending buffer, not resolutions. A firing whose clauses later fizzle still used the turn's allowance. ("This ability triggers only once each turn.")
3. **"Each turn" is every turn, both players'.** The count is kept per asset and per trigger, keyed by `Game.Turn`, so it resets at the start of every turn (`FR-TRN-001`). A trigger limited to once each turn can fire once on the opponent's turn and once again on its controller's turn.
4. **Per asset:** two assets printing the same trigger each have their own allowance.
5. Where the count lives is the engineer's call: a field on `Asset` or a map on `Game` in `state.go`. Nothing copies a `Game`, so a map needs no clone support. The count is not part of snapshots or events.
6. **`doc.go`:** update the "Built" and "Not built yet" lists. Unbowed (T-0050), the zone criterion (T-0052) and this limit are built; Headquarters abilities stay "not built" until T-0055 lands.

### What the tests can observe (for rules-test-author, synthetic cards)
- A unit whose `HookDamaged` trigger has `MaxPerTurn: 1` (a Scratch 1, observed with `wantScratches`):
  - takes two separate damage instances in one turn and Scratches once;
  - takes damage again on the next turn, whichever player's it is, and Scratches again;
  - takes damage on the opponent's turn and on its controller's next turn: one Scratch each.
- The same trigger without a limit (`MaxPerTurn` 0) fires every time: existing behavior.
- Two assets printing the limited trigger, both damaged in one turn, Scratch once each.
- A firing whose clause fizzles still uses the allowance. For example, a limited trigger with a chosen target whose target is gone by resolution, then a second damage instance in the same turn: no second firing.
- A synthetic **Headquarters** with a limited `HookDamaged` trigger: attacked twice in one turn (by two attackers), it Scratches once. A Headquarters can also take damage from its own attack's retaliation.
- `MaxPerTurn: 2` allows two firings in a turn and not a third.

Not tested: an asset that leaves the battlefield and returns in the same turn. It is a new asset (`FR-ZON-004`), so it would start with a fresh allowance, but no starter card can show it.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
Tests are in `pkg/engine/max_per_turn_test.go`. Done When regex:
`go test ./pkg/engine -run 'TestFR_TRN_001_(LimitedTriggerFiresOnceForTwoDamageInstancesInOneTurn|UnlimitedTriggerStillFiresEveryTime|LimitOfTwoAllowsTwoFiringsAndNotAThird|EachAssetHasItsOwnAllowance|LimitResetsOnTheControllersNextTurn|LimitResetsOnTheOpponentsNextTurn|OneFiringOnTheOpponentsTurnAndOneOnTheControllersNext|AFizzledFiringStillUsesTheTurnsAllowance|HeadquartersLimitedTriggerFiresOnceWhenAttackedTwice|HeadquartersFiresAgainFromItsOwnAttacksRetaliationNextTurn)$'`

- `TestFR_TRN_001_LimitedTriggerFiresOnceForTwoDamageInstancesInOneTurn`: MaxPerTurn 1, two attackers, one Scratch.
- `TestFR_TRN_001_UnlimitedTriggerStillFiresEveryTime`: MaxPerTurn 0, two Scratches (existing behavior).
- `TestFR_TRN_001_LimitOfTwoAllowsTwoFiringsAndNotAThird`: MaxPerTurn 2, three attackers, two Scratches.
- `TestFR_TRN_001_EachAssetHasItsOwnAllowance`: two assets printing the limited trigger, each hit twice, one Scratch each (2 total).
- `TestFR_TRN_001_LimitResetsOnTheControllersNextTurn`: fires turn 1, two end-turns, fires again.
- `TestFR_TRN_001_LimitResetsOnTheOpponentsNextTurn`: fires on its controller's turn (retaliation), then again on the opponent's turn.
- `TestFR_TRN_001_OneFiringOnTheOpponentsTurnAndOneOnTheControllersNext`: damaged on the opponent's turn, then on its controller's next turn: one each.
- `TestFR_TRN_001_AFizzledFiringStillUsesTheTurnsAllowance`: chosen target removed before resolution; a second damage in the same turn does not fire (the spare target takes no damage).
- `TestFR_TRN_001_HeadquartersLimitedTriggerFiresOnceWhenAttackedTwice`: synthetic Headquarters (ability set on the HQ card), two attackers, one Scratch.
- `TestFR_TRN_001_HeadquartersFiresAgainFromItsOwnAttacksRetaliationNextTurn`: Headquarters hit on turn 1, then attacks on its own turn and takes retaliation: fires again.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
None. The test file defines its own `limited(n, trig)` and `(*scene).hitters(n)`; it reuses `bowed`, `unbowed`, `wantScratches`, `scratchN`, `tgt`, `dmgTo`.

## Allowed Files
- `pkg/engine/ability/ability.go`
- `pkg/engine/trigger.go`
- `pkg/engine/state.go`
- `pkg/engine/doc.go`

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run '<regex from the Tests section>'`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean, and `cardtown sim -deck starter -p0 ANIMUS -p1 PATH` output is byte-identical before and after (no card uses the limit yet). Record it in the Log.

## Log
- 2026-10-23 lead-architect: Wrote the task (wave B shape 3), with status `tests` for `rules-test-author`. Checked every waiting directive's trigger first: none has fired (#8: no `deck.go` change; #15: this is a limit, not a turn hook; #20: wave B). Animoceps Rager was picked over Infested Apostle, Chainbreaker Irregulars and Animoceps Shambler because those three each need a ruling (design note Open Question 8). The card task T-0055 follows. Starter-deck cards not yet executable: 28, plus both Headquarters.
- 2026-10-23 rules-test-author: Wrote 10 tests in `pkg/engine/max_per_turn_test.go` (listed under Tests). The package does not compile until `ability.Trigger` has `MaxPerTurn int` (`go vet` reports only `trig.MaxPerTurn undefined`); that is expected. `cardtown lint rules` passes. Once it compiles, the tests that fail against an engine ignoring the limit are: LimitedTriggerFiresOnce..., LimitOfTwo..., EachAssetHasItsOwnAllowance, AFizzledFiring..., HeadquartersLimitedTriggerFiresOnce...; the other five (unlimited, three reset tests, HQ retaliation) pin behavior a correct limit must keep. Unverified at runtime, since nothing compiles: that a Headquarters may attack (used in the HQ retaliation test, from the task text), and that the fizzle test's mid-resolution removal (`removeFromBattlefield` in the chooser, after the target is picked) fizzles the clause. Left out: Animoceps Rager card tests (T-0055); an asset leaving and returning in the same turn (task: not tested); the doc.go refresh (not testable). No helpers requested.
