---
id: T-0203
title: The moves-zone hook, for Chainbreaker Phalanx and Broken Horn Armorer (wave F shape, sonnet)
kind: shape
tier: sonnet
wave: F
status: done
branch: m2/autoloop-wave-f/T-0203-move-hook
depends_on: []
claims: [hook:moves (ability.HookMoves, Trigger.ToZone on it, Game.fireMoved)]
cards: []
requirements: [FR-ZON-002, FR-TRG-002, FR-TRG-004]
---

## Goal
A printed trigger can watch its own unit moving to a new battlefield zone, by March or by a step that moves it, and `Trigger.ToZone` narrows it to "moves to the Main Theater". Entering the battlefield is not moving (`SVY-R1`). Two wave F deck cards need this hook: *Chainbreaker Phalanx* (T-0204) and *Broken Horn Armorer* (T-0205). Both card definitions go in later tasks, under the D2 size rule: this shape touches three engine files, and each card adds `cards/path.go` and `cards/cards.go` (Armorer also needs a "same zone" filter). So this task's tests use synthetic cards, and no card prints the hook yet.

Covers:
- `ability.HookMoves` and `ToZone` on it, as under Shape below.
- One place the hook fires from: `moveOnBattlefield`, the one function both moves (March and `OpMoveZone`) go through.

## Inputs
- FRD 01: `FR-ZON-002` (lines 230–236), above all "What Counts as Moving (`SVY-R1`)" (line 233) and the ruling summary at line 99; March in `FR-CAP-001` (line 323): the unit exhausts on declaration and moves when the March resolves. `FR-TRG-002` (self hooks), `FR-TRG-003` and `FR-TRG-004` (placement after the item's sweep, cascades).
- Survey rows: `docs/design/m2_starter_set_ability_survey.md` line 198 (*Broken Horn Armorer*, "self moves zone") and line 199 (*Chainbreaker Phalanx*, "self moves to Main Theater (may)", "can chain into another Phalanx"); `R1` at line 59.
- Rulings: `SVY-R1`.
- Design note: `docs/architecture/m2-ability-model.md`, Key types (`Trigger`, `Hook`: "moves zone" is already listed) and Boundaries.
- Card text (`data/cards/units-and-structures-path.csv`), for the record only: *Broken Horn Armorer*: "Whenever this unit moves to a new zone, chosen asset in the same zone gains 1 armor." *Chainbreaker Phalanx*: "Whenever this unit moves to the Main Theater, you may move chosen allied unit from your Reserves to the Main Theater. That unit gains 2 armor."

### Shape (the engineer builds it)
- **`ability.HookMoves`**, appended after `HookCombatKill` in the `Hook` enum ([ability.go:115-121](../../../pkg/engine/ability/ability.go)), so no existing value changes. Its comment: this unit moved to a new battlefield zone, by March or by a step that moves it (`OpMoveZone`); entering the battlefield, by any deploy, is not moving (`SVY-R1`, `FR-ZON-002`).
- **`Trigger.ToZone`** ([ability.go:159](../../../pkg/engine/ability/ability.go)) applies to `HookMoves` too: "moves to the Main Theater" is `ToZone: ZoneMainTheater`, and `ZoneAny` ("moves to a new zone") fires on every move. Update its comment.
- **`fireMoved(a)`** in `trigger.go`, beside `fireDeploy` and the others: `g.fireHook(a, ability.HookMoves)`. In `fireHook` ([trigger.go:90](../../../pkg/engine/trigger.go)) the `ToZone` gate checks `HookMoves` as well as `HookDeploy`, with the same `deployedTo` read, whose comment ([zones.go:235-237](../../../pkg/engine/zones.go)) then says "the zone its deploy or move put it in". The Cornered, `Turn` and `MaxPerTurn` gates apply as for every hook, through the same loop.
- **`moveOnBattlefield`** ([zones.go:198-204](../../../pkg/engine/zones.go)) calls `g.fireMoved(a)` after placing the asset. Its two callers are `resolveMarch` ([stack.go:122](../../../pkg/engine/stack.go)) and `stepMoveZone` ([effects.go:137](../../../pkg/engine/effects.go)), and each already skips a unit that doesn't move (a March whose unit left or already moved, a unit already in the step's zone), so a move that doesn't happen fires nothing. Neither caller changes. Deploys reach the battlefield through `placeOnBattlefield`, not `moveOnBattlefield`, so they fire nothing (`SVY-R1`). A face-down asset has no printed abilities, as for every hook.
- The buffered trigger is placed in the flush after the item's sweep, as every trigger fired during resolution is (`FR-TRG-003`). A unit moved by a resolving move trigger fires its own, placed in the next flush (`FR-TRG-004`).

No card prints `HookMoves` in this task, so `go tool task sim-compare` must be byte-identical: a gate here.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
In a new file `pkg/engine/move_hook_test.go`, with synthetic cards (as `exhaust_hook_test.go` builds them, for example a unit printing "whenever this moves, Scratch 1", observed with `wantScratches` and the log order):
- `TestFR_ZON_002_MarchFiresMoveTrigger` (`FR-ZON-002`, `FR-CAP-001`): a unit with a `HookMoves` trigger (`ToZone` any) marches from Reserves to the Main Theater, and its trigger resolves after the March; marching back to Reserves fires it again.
- `TestFR_ZON_002_MoveStepFiresMoveTrigger` (`FR-ZON-002`): a synthetic Maneuver or trigger whose `OpMoveZone` step moves the watcher fires the watcher's trigger, placed after that item resolves.
- `TestFR_ZON_002_DeployIsNotMoving` (`SVY-R1`): deploying the watcher, to Reserves and (with a Main Theater deploy zone) to the Main Theater, fires no move trigger.
- `TestFR_ZON_002_MovesToZoneGate` (`FR-ZON-002`, `SVY-R1`): a trigger with `ToZone: ZoneMainTheater` fires when its unit marches into the Main Theater and not when it marches to Reserves.
- `TestFR_ZON_002_NoMoveNoTrigger` (`FR-ZON-002`): an `OpMoveZone` step to the zone the watcher is already in fires nothing.
- `TestFR_ZON_002_OnlyTheMovedUnitFires` (`FR-TRG-002`): a self hook; another watcher that doesn't move fires nothing when the first one marches.
- `TestFR_TRG_004_MoveTriggerCascades` (`FR-TRG-004`): watcher A's move trigger moves allied watcher B from Reserves to the Main Theater (an `OpMoveZone` step on a target); B's own move trigger fires and is placed after A's trigger resolves (*Chainbreaker Phalanx*'s chain, survey line 199).
- **Existing tests:** every other test passes unchanged.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected (`striker`, `scratchN`, `deployFromHand`, `wantScratches` and `logIndex` exist). The test author keeps any small builder in the test file.

## Allowed Files
- `pkg/engine/ability/ability.go` (`HookMoves`, the `ToZone` comment)
- `pkg/engine/trigger.go` (`fireMoved`, the `ToZone` gate)
- `pkg/engine/zones.go` (`moveOnBattlefield` fires it, the `deployedTo` comment)

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run "TestFR_ZON_002|TestFR_TRG_004_MoveTriggerCascades"`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-compare -- autoloop/m2-wave-f` passes (no card prints the hook yet)

## Log
- 2026-11-09 lead-architect (wave F run, session 2): written against 2002405. Board read (D2.0): no `open: before` directive applies to this task (#140's next part is `open: before` the *Battle Trance* task, which isn't written). Waiting triggers checked: #6, #7 (no mode change), #15 (no turn hook), #71 (no each-player effect), #91 (2) (no mode cost), #106 (no criteria search), #120 (no mandatory discard cost), #128 (no discard deploy), #136 (no off-battlefield trigger) and #137 (no keyword grant) have not fired. Searched `pkg/engine` and the task archive for a move hook or a same-zone filter: none exist (the `Hook` enum ends at `HookCombatKill`; `Filter` has no same-zone criterion), so nothing is duplicated, and both March and `OpMoveZone` already share `moveOnBattlefield`, the one place this fires from. Split under D2.4: *Chainbreaker Phalanx* is T-0204 (card, flash) and *Broken Horn Armorer* with the same-zone filter is T-0205 (shape, sonnet), both `queued` on this task. Set to `tests`.
- 2026-11-09 rules-test-author: wrote `pkg/engine/move_hook_test.go` (7 tests, all as named in the task): `TestFR_ZON_002_MarchFiresMoveTrigger`, `TestFR_ZON_002_MoveStepFiresMoveTrigger`, `TestFR_ZON_002_DeployIsNotMoving`, `TestFR_ZON_002_MovesToZoneGate`, `TestFR_ZON_002_NoMoveNoTrigger`, `TestFR_ZON_002_OnlyTheMovedUnitFires`, `TestFR_TRG_004_MoveTriggerCascades`. Local builders in the test file: `moveTrigger`, `moveWatcher`, `selfDamage`, `(*scene).moveTargetFree` (reuses `moveStep`, `choosesAssets`, `enactFree`). Helpers requested: none. `go vet` fails only on `ability.HookMoves`; `cardtown lint rules` passes. Left out: nothing from the task list. Cases not written: the face-down no-abilities edge (no FRD ruling needed beyond existing hook rules), and the Cornered/Turn/MaxPerTurn gates on this hook (shared loop, covered by other hooks' tests). Move-step and cascade ordering is observed through the `damage`/`scratch` log events, since draws and steps are not logged.
- 2026-11-09 lead-architect (wave F run, session 3): tests reviewed (D4) against 3fea3a0. Each of the 7 tests checks the outcome its cited rule gives (`FR-ZON-002`, `SVY-R1`, `FR-TRG-002`, `FR-TRG-003`, `FR-TRG-004`), with no assertion on implementation details. The Main Theater deploy in `TestFR_ZON_002_DeployIsNotMoving` is legal, because the `unit` builder prints both deploy zones. Corrected the test author's count in its Log entry ("8 tests, 7 as named" now reads "7 tests, all as named"), as it asked. Set to `ready`, dispatching `kernel-engineer` (sonnet).
- 2026-11-09 kernel-engineer: added `ability.HookMoves` (after `HookCombatKill`) and widened the `ToZone` comment; `fireMoved` in `trigger.go` beside `fireDeploy`; the `ToZone` gate in `fireHook` now covers `HookDeploy` and `HookMoves` through `deployedTo`; `moveOnBattlefield` calls `g.fireMoved(a)` after placing the asset; `deployedTo` comment updated. Reused `fireHook`/`deployedTo`; no callers changed. Ran: `go test ./pkg/engine -run "TestFR_ZON_002|TestFR_TRG_004_MoveTriggerCascades"` ok; `go tool task check` exit 0; `go tool task sim` clean; `go tool task sim-compare -- autoloop/m2-wave-f` PASS byte-identical.
- 2026-11-09 lead-architect (wave F run, session 3): code reviewed (D5) against b25a2b1. Scope: only the three Allowed Files changed, and `move_hook_test.go` is the same as in 3fea3a0. Fences: no card names, no `OPEN(…)`. Reuse: `fireMoved` goes through `fireHook`, and the `ToZone` gate reuses `deployedTo`, so nothing is duplicated. Fidelity: the hook fires from `moveOnBattlefield`, the one function both March and `OpMoveZone` use, and deploys still go through `placeOnBattlefield` alone (`SVY-R1`). Gates re-run in the review worktree: the listed tests pass, `go tool task check` passes, `go tool task sim` runs clean (200 games, P0 73.5% / P1 26.0% / draws 0.5%), and `go tool task sim-compare -- autoloop/m2-wave-f` is byte-identical. One drift item: the hook list in `pkg/engine/doc.go:18-19` doesn't name `HookMoves`, and `doc.go` isn't one of this task's Allowed Files. It goes into T-0204's Covers, with the done commit, because T-0205 would exceed four non-test files with it. Approved.
- 2026-11-09 lead-architect (wave F run, session 3): merged into `autoloop/m2-wave-f` as c5bc636. Set to `done`.
