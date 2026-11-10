---
id: T-0205
title: A zone locked on the stack item, first a move's destination (wave F shape, sonnet; #141 (1))
kind: shape
tier: sonnet
wave: F
status: review
branch: m2/autoloop-wave-f/T-0205-locked-zone
depends_on: [T-0203]
claims: [shape:locked zone on the stack item (engine.LockedZone, StackItem.Locked, placement.locked; fireMoved locks the move's destination)]
cards: []
requirements: [FR-TGT-001, FR-TGT-003, FR-ZON-002]
---

## Goal
Directive #141 (1): a zone criterion relative to the source ("in that zone", "in the same zone") compares an entity's zone with **one zone stored on the stack item**, set once and read at both target checks and by untargeted picks on resolution (`SVY-R44`, `SVY-R45`). This task builds that stored zone and its first setter: a trigger on a move locks the zone the move went to, fixed by the event (`SVY-R45`). T-0206 adds the filter criterion that reads it, and T-0207 builds *Broken Horn Armorer* ("Whenever this unit moves to a new zone, chosen asset in that zone gains 1 armor."). The three were one task, split under the D2 size rule: the whole shape touches seven non-test files (this task's three, and `ability.go`, `filter.go`, `targeting.go`, `effects.go` for the criterion).

Covers:
- `LockedZone`, `StackItem.Locked` and `placement.locked`, as under Shape below.
- One setter: `fireMoved` locks the move's destination on the triggers it buffers.
- Not this task: the criterion that reads the zone (T-0206, `queued`); the "same zone" setter, the source's zone on placement or its last-known zone (`SVY-R44`), which arrives with *Hammerhand Bivouac* or the first stack-item card that prints "in the same zone" (#141 (3), `waiting:`).

## Inputs
- FRD 01 v3.23: `FR-TGT-001` (lines 370–383), above all the Zone criterion (line 374), "**"In the Same Zone" Is Locked on Placement (`SVY-R44`)**" (line 382) and "**"That Zone" After a Move (`SVY-R45`)**" (line 383); `FR-TGT-003` Two-Stage Verification (line 395); `FR-ZON-001` (line 220: the Main Theater is shared, Reserves and the Headquarters Zone are each player's own); `FR-ZON-002` (moving, `SVY-R1`); `FR-ZON-004` (line 246: the zone is not a value read from the source).
- Rulings: `SVY-R44`, `SVY-R45` (survey lines 103–104); outcome `docs/design/outcomes/2026-11-09-wave-f-same-zone-source.md`.
- Board directive #141 (1).
- Design note: `docs/architecture/m2-ability-model.md`, Key types (`Filter`, line 39) and Open Question 20.
- Code: `fireMoved` and `fireCombatKill` ([trigger.go:47-73](../../../pkg/engine/trigger.go)), `bufferTrigger` and `flushTriggers` ([trigger.go:111-170](../../../pkg/engine/trigger.go)), `placement` and `stackItem` ([placement.go:16-29](../../../pkg/engine/placement.go)), `StackItem` ([stack.go:20-33](../../../pkg/engine/stack.go)), `moveOnBattlefield` ([zones.go:201-205](../../../pkg/engine/zones.go)).

### Shape (the engineer builds it)
- **`LockedZone`**, in `stack.go` beside `StackItem`: a battlefield zone, `Zone Zone` (`ZoneReserves`, `ZoneMainTheater` or `ZoneHeadquarters`), and `Player PlayerID`, whose Reserves or Headquarters Zone it is (unused for the shared Main Theater, `FR-ZON-001`). Its comment: the zone a stack item's source-relative filters read, locked once (`SVY-R44`, `SVY-R45`).
- **`StackItem.Locked *LockedZone`**: nil when the item locked no zone. Its comment names the kinds that can carry one (`StackAbility`, `StackManeuver`) and the rulings.
- **`placement.locked *LockedZone`**, and `stackItem` copies it onto the item, so every item built from a placement (a flushed trigger, a `declare`d Maneuver or activated ability) carries the placement's zone. Nothing sets it for Maneuvers or activated abilities yet.
- **`fireMoved(a)`** locks, on each trigger it buffers, `LockedZone{Zone: a.Zone, Player: a.Controller}`: `moveOnBattlefield` calls it after placing `a`, so that is the move's destination. Mark the triggers as `fireCombatKill` marks its own (the pending triggers buffered from index `n` on). Each trigger gets its own value: a later move of the same unit, before or after the flush, never changes a zone already locked (`SVY-R45`, "fixed by the event"). A trigger whose source has left the battlefield before the flush is still placed, as today, with its zone.
- Behavior doesn't change: nothing reads `Locked` yet, and no printed card's output does. `go tool task sim-compare` must be byte-identical (a gate here).

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
In a new file `pkg/engine/locked_zone_test.go`, with synthetic units printing a `HookMoves` trigger (`ToZone` any, an effect such as Scratch 1; `moveTrigger` and `moveWatcher` exist in `move_hook_test.go`). To see an item between placement and resolution, drive the steps directly: `s.g.moveOnBattlefield(w, z)`, then `s.g.flushTriggers()`, then read `s.g.Stack`; `s.g.resolveStack()` afterwards if the test needs a settled game.
- `TestFR_TGT_001_MoveLocksMainTheater` (`SVY-R45`): player 0's watcher moves from Reserves to the Main Theater; the placed item's `Locked` is the Main Theater.
- `TestFR_TGT_001_MoveLocksOwnReserves` (`SVY-R45`, `FR-ZON-001`): player 1's watcher moves from the Main Theater to Reserves; the item's `Locked` is Reserves with `Player` 1.
- `TestFR_TGT_001_EachMoveLocksItsDestination` (`SVY-R45`; the outcome's third rejected alternative): the watcher moves to the Main Theater and back to Reserves before one flush; the two placed items lock the Main Theater and Reserves, one each (whatever their order on the stack).
- `TestFR_TGT_001_LockedZoneOutlivesSource` (`SVY-R45`, `FR-ZON-004`): (a) the watcher moves to the Main Theater and then leaves the battlefield (for example `s.g.leaveBattlefield(w, s.g.effectiveState(w))`) before the flush: the item is placed and locks the Main Theater; (b) after the flush the watcher is moved again without firing (`s.moveTo`): the placed item still locks the Main Theater.
- `TestFR_TGT_001_NoLockWithoutMove`: a deploy trigger's item (a synthetic `HookDeploy` unit, deployed with `s.g.deployCard` and flushed) has a nil `Locked`.
- **Existing tests:** every other test passes unchanged.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected. The test author keeps any small builder in the test file.

## Allowed Files
- `pkg/engine/stack.go` (`LockedZone`, `StackItem.Locked`)
- `pkg/engine/placement.go` (`placement.locked`, `stackItem` copies it)
- `pkg/engine/trigger.go` (`fireMoved` locks the destination)

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run "TestFR_TGT_001_(MoveLocks|EachMoveLocks|LockedZoneOutlives|NoLockWithoutMove)"`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-compare -- autoloop/m2-wave-f` passes (nothing reads the zone yet)

## Log
- 2026-11-09 lead-architect (wave F run, session 2): split off T-0203 under the D2 size rule; `queued` on it.
- 2026-11-09 lead-architect (wave F run, session 4): not written yet, stays `queued`. Its Covers left the departed-source reading of "same zone" to FRD 01, and FRD 01 settles neither a source that moved to the other zone nor one that left the battlefield before the second target check (`FR-TGT-003`, `FR-ZON-004`). Ruling question queued on `main` in `docs/design/outcomes/2026-11-09-wave-f-same-zone-source.md` (111c8fb), design note Open Question 20. Write Inputs, Tests and Allowed Files once its `Status:` reads `applied`.
- 2026-11-10 lead-architect (wave F run, session 5): the outcome reads `Status: applied: FRD 01 v3.23`, so #141 is `open: before T-0205`. Rewritten for #141 (1): the Armorer's shape is one zone locked on the stack item plus a filter criterion that reads it, seven non-test files, so it splits under the D2 size rule into this task (the stored zone and the move-destination setter), T-0206 (the "in that zone" criterion, `queued`) and T-0207 (*Broken Horn Armorer*, card, `queued`). Renamed from `T-0205-broken-horn-armorer.md`. Set to `tests`.
- 2026-11-10 rules-test-author: tests written in pkg/engine/locked_zone_test.go: TestFR_TGT_001_MoveLocksMainTheater, TestFR_TGT_001_MoveLocksOwnReserves, TestFR_TGT_001_EachMoveLocksItsDestination, TestFR_TGT_001_LockedZoneOutlivesSource (subtests: source leaves before flush; source moves again after flush), TestFR_TGT_001_NoLockWithoutMove. Helpers requested: none (small helpers flushedItems and wantLocked live in the test file). `go vet` fails only on the missing StackItem.Locked field; `cardtown lint rules` passes. Left out: nothing. Status set to tests-review.
- 2026-11-10 lead-architect (wave F run, session 5): tests reviewed (60ae12a, 5 tests, one with two subtests): each matches the Tests list and `SVY-R45`, `FR-ZON-001`; set to `ready`.
- 2026-11-10 kernel-engineer: implemented. Added `LockedZone` and `StackItem.Locked` (stack.go), `placement.locked` copied by `stackItem` (placement.go), and `fireMoved` now locks `LockedZone{Zone: a.Zone, Player: a.Controller}` on each trigger it buffers from index n on, one fresh value each, as `fireCombatKill` marks its own (trigger.go). Reuse: `fireHook` and the pending-index marking idiom of `fireCombatKill`; no new shape. Ran: the T-0205 tests (ok), `go tool task check` (exit 0), `go tool task sim` (exit 0), `go tool task sim-compare -- autoloop/m2-wave-f` (byte-identical). Reviewer: nothing reads `Locked` yet.
