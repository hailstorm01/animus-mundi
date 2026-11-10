---
id: T-0206
title: The "in that zone" filter criterion, read against the locked zone (wave F shape, sonnet; #141 (1))
kind: shape
tier: sonnet
wave: F
status: tests-review
branch: m2/autoloop-wave-f/T-0206-that-zone-criterion
depends_on: [T-0205]
claims: [shape:"in that zone" filter criterion (ability.Filter relative-zone criterion; matchesFilter and legalTarget read the item's locked zone)]
cards: []
requirements: [FR-TGT-001, FR-TGT-003]
---

## Goal
Directive #141 (1), second half: a filter criterion "in that zone" (`SVY-R45`) that holds when the entity is in the zone its stack item locked (T-0205's `StackItem.Locked`, or `placement.locked` at the first target check). It is read at both target checks (`FR-TGT-003`) and by untargeted steps on resolution (`SubjectAll`, `pickAssets`), through the one filter check, so a later "in the same zone" (`SVY-R44`, #141 (3)) only adds another setter of the same locked zone, and no second criterion. Split from T-0205 under the D2 size rule.

Covers:
- An `ability.Filter` criterion relative to the source, as an enum whose zero value is "none" and whose first value is "in that zone, the triggering move's destination" (`SVY-R45`), with room for "in the same zone" (`SVY-R44`) later.
- `matchesFilter` ([filter.go:85](../../../pkg/engine/filter.go)) checks it against the locked zone: the entity's zone is the locked `Zone`, and for Reserves or the Headquarters Zone its controller is the locked `Player` (`FR-ZON-001`). A filter with the criterion and no locked zone matches nothing. `matchesFilter` never reads the source's current zone for it (#141 (1)).
- The callers that have a placement or an item pass its zone: `targetDecision` and `pickAssets` ([targeting.go:93](../../../pkg/engine/targeting.go), [:160](../../../pkg/engine/targeting.go)), the second target check and `SubjectAll` ([effects.go:41](../../../pkg/engine/effects.go), [:396](../../../pkg/engine/effects.go)). `costs.go` and `replace.go` keep their calls unchanged (keep a form of `matchesFilter` without a locked zone, so the task stays within four files).
- Synthetic-card tests only; no card prints it until T-0207.

- Not this task: *Broken Horn Armorer* (T-0207, `queued`); the "same zone" setter (`SVY-R44`, #141 (3), `waiting:`), which adds a second value of the same enum and sets the item's locked zone from the source on placement, with no second matcher.

## Inputs
- FRD 01 v3.23: `FR-TGT-001` (lines 370–383), above all the Zone criterion (line 374) and "**"That Zone" After a Move (`SVY-R45`)**" (line 383); `FR-TGT-003` Two-Stage Verification (line 395: a source-relative zone restriction is checked against the locked zone, not the source's zone at resolution); `FR-ZON-001` (the Main Theater is shared, Reserves and the Headquarters Zone are each player's own); `SVY-R3`, `SVY-R14` (an untargeted pick is not a target: no Ward, no second check).
- Rulings: `SVY-R44`, `SVY-R45` (survey lines 103–104); outcome `docs/design/outcomes/2026-11-09-wave-f-same-zone-source.md`.
- Board directive #141 (1).
- Design note: `docs/architecture/m2-ability-model.md`, Key types (`Filter`, line 39).
- Code: `LockedZone`, `StackItem.Locked` ([stack.go:19-42](../../../pkg/engine/stack.go)), `placement.locked` ([placement.go:22](../../../pkg/engine/placement.go)), `fireMoved` ([trigger.go:47-56](../../../pkg/engine/trigger.go)), all from T-0205; `ability.Filter` ([ability.go:217-229](../../../pkg/engine/ability/ability.go)); `legalTarget`, `matchesFilter` ([filter.go:33-93](../../../pkg/engine/filter.go)); `targetDecision`, `pickAssets` ([targeting.go:76-161](../../../pkg/engine/targeting.go)); the second check and `SubjectAll` ([effects.go:28-46](../../../pkg/engine/effects.go), [:393-399](../../../pkg/engine/effects.go)).

### Shape (the engineer builds it)
- **`ability.ZoneRef`**, an `int` enum in `ability.go` beside `Filter`, with `ZoneRefNone` (the zero value: no source-relative zone criterion) and `ZoneRefThat` ("in that zone": the zone the stack item locked, a move's destination, `SVY-R45`). Its comment says a later "in the same zone" (`SVY-R44`) is another value read the same way. **`Filter.ZoneRef ZoneRef`**: the entity must be in its item's locked zone. It is a conjunction with every other criterion, `Filter.Zone` included.
- **`filter.go`:** the one check of the locked zone, `inLocked(a *Asset, lock *LockedZone) bool`: `lock` is non-nil, `a.Zone == lock.Zone`, and, unless that zone is the shared Main Theater, `a.Controller == lock.Player` (`FR-ZON-001`). `matchesFilterAt(ctl, src, lock, f, a)` holds today's `matchesFilter` body and, when `f.ZoneRef != ZoneRefNone`, also requires `inLocked(a, lock)`; so a filter with the criterion and a nil lock matches nothing. It never reads the source's current zone (#141 (1)). `matchesFilter(ctl, src, f, a)` stays, as `matchesFilterAt` with a nil lock, for `costs.go` and `replace.go`, which have no item (they keep their calls unchanged). `legalTarget` gains a `lock *LockedZone` parameter after `src` and calls `matchesFilterAt`.
- **Callers pass the zone:** `targetDecision` passes `pl.locked` (so `chooseTargets` and `legalModes` both see it); `pickAssets` and `SubjectAll` pass `r.it.Locked`; the second check in `resolveAbility` passes `it.Locked`.
- Behavior doesn't change for any printed card: none sets `ZoneRef`. `go tool task sim-compare -- autoloop/m2-wave-f` must be byte-identical (a gate here).

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
In a new file `pkg/engine/that_zone_test.go`, with synthetic units printing a `HookMoves` trigger (`ToZone` any) whose effect uses `ability.Filter{ZoneRef: ability.ZoneRefThat}`: for targets, one `Target` and an `OpAddArmor` 1 step on it; for untargeted steps, a `SubjectAll` or `SubjectPick` step with that filter. Use `moveTrigger` (`move_hook_test.go`) and `flushedItems` (`locked_zone_test.go`) where they fit. Move with `s.g.moveOnBattlefield(w, z)` to fire, then `s.g.flushTriggers()` and `s.g.resolveStack()`; move without firing with `s.moveTo`. Read the offered options with `s.asked(p)` (script answers with `s.script`). Use one other unit in each of the four battlefield zones that matter (player 0's and player 1's in the Main Theater, player 0's and player 1's in Reserves).
- `TestFR_TGT_001_ThatZoneOffersLockedZoneOnly` (`SVY-R45`, `FR-ZON-001`): (a) player 0's watcher moves from Reserves to the Main Theater: the target decision offers exactly the watcher itself and both players' Main Theater units, none in Reserves; (b) player 0's watcher moves from the Main Theater to Reserves: it offers exactly the watcher and player 0's other Reserves unit, not player 1's.
- `TestFR_TGT_003_ThatZoneTargetLeftZoneIsIllegal` (`FR-TGT-003`, `SVY-R45`): after the flush, the chosen Main Theater unit is moved to Reserves (`s.moveTo`) before resolution: it gains no armor.
- `TestFR_TGT_003_ThatZoneTargetOutlivesSource` (`SVY-R45`, `FR-ZON-004`): after the flush and before resolution, (a) the watcher is moved back to Reserves (`s.moveTo`), (b) the watcher leaves the battlefield: in both, the chosen Main Theater unit still gains 1 armor.
- `TestFR_TGT_001_ThatZoneAllStepActsOnLockedZone` (`SVY-R45`, `FR-ZON-001`, `SVY-R3`): a `SubjectAll` step (for example `OpAddArmor` 1) acts on every asset in the locked zone and nothing else: for a Main Theater lock, both players' Main Theater assets; for a Reserves lock, player 0's Reserves assets only.
- `TestFR_TGT_001_ThatZonePickOffersLockedZone` (`SVY-R45`, `SVY-R14`): a `SubjectPick` step's `DecidePick` offers only assets in the locked zone, read when it resolves: an asset moved into the locked zone after the flush is offered, one moved out is not.
- `TestFR_TGT_001_ThatZoneWithoutLockMatchesNothing` (#141 (1)): on an item with no locked zone (a synthetic `HookDeploy` trigger with the same target and `SubjectAll` step, deployed with `s.g.deployCard` and flushed), the target slot offers nothing and stays empty, and the `SubjectAll` step acts on nothing.
- **Existing tests:** every other test passes unchanged.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected. The test author keeps any small builder in the test file.

## Allowed Files
- `pkg/engine/ability/ability.go` (`ZoneRef`, `Filter.ZoneRef`)
- `pkg/engine/filter.go` (`inLocked`, `matchesFilterAt`, `legalTarget`'s lock)
- `pkg/engine/targeting.go` (`targetDecision`, `pickAssets` pass the lock)
- `pkg/engine/effects.go` (the second check and `SubjectAll` pass the lock)

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run "TestFR_TGT_00[13]_ThatZone"`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-compare -- autoloop/m2-wave-f` passes (no printed card sets `ZoneRef`)

## Log
- 2026-11-10 lead-architect (wave F run, session 5): split off T-0205 under the D2 size rule (#141 (1)); `queued` on it.
- 2026-11-10 lead-architect (wave F run, session 6): Inputs, Shape, Tests and Allowed Files written against the run branch at dd65d01 (#141 (1), second half). Set to `tests`.
- 2026-11-10 rules-test-author: wrote pkg/engine/that_zone_test.go: TestFR_TGT_001_ThatZoneOffersLockedZoneOnly (Main Theater and Reserves locks), TestFR_TGT_003_ThatZoneTargetLeftZoneIsIllegal, TestFR_TGT_003_ThatZoneTargetOutlivesSource (source moved back; source left), TestFR_TGT_001_ThatZoneAllStepActsOnLockedZone (both locks), TestFR_TGT_001_ThatZonePickOffersLockedZone, TestFR_TGT_001_ThatZoneWithoutLockMatchesNothing. Helpers requested: none (local builders prefixed tz in the test file). Vet fails only on the missing ability.ZoneRefThat / Filter.ZoneRef. Left out: nothing. Note: the pick test uses an OpKill SubjectPick step (UpTo 1) answered with a timeout, so no unit dies; it asserts only the offered options.
