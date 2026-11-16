---
id: T-0246
title: '"In the same zone" on a stack item reads its source''s zone, locked on placement (wave H shape, sonnet; Hammerhand Bivouac''s zone, #141 (3) for stack items)'
kind: shape
tier: sonnet
wave: H
status: ready
branch: m2/autoloop-wave-h/T-0246-same-zone-lock
depends_on: [T-0205, T-0206, T-0245]
claims: [shape:zone criterion "in the same zone" (ability.ZoneRefSame; a trigger or activated ability that reads it locks its source's zone on placement, last-known if gone, SVY-R44)]
cards: []
requirements: [FR-TGT-001, FR-TGT-003, FR-ZON-004, FR-ACT-001]
---

## Goal
*Hammerhand Bivouac* ("At the start of your turn, distribute 2 armor among allied units in the same zone.") picks among allied units in the zone its source was in when the trigger was placed (`SVY-R44`). Today the only source-relative zone criterion is `ZoneRefThat`, a move's destination ([ability.go:282-290](../../../pkg/engine/ability/ability.go)), and only `fireMoved` sets a stack item's `Locked` ([trigger.go:94-101](../../../pkg/engine/trigger.go)). This task adds "in the same zone" as a second `ZoneRef`, read through the same `Locked` and `inLocked` ([filter.go:125-133](../../../pkg/engine/filter.go)), which the `ZoneRef` comment already plans for. The distributed pool (directive #157, `SVY-R51`, `SVY-R52`) builds on this lock in the next task, T-0247.

Covers:
- `ability.ZoneRefSame`, and the one reader that tells which `ZoneRef` an effect reads.
- A trigger whose effect reads `ZoneRefSame` locks its source's zone when it is placed (`placeTrigger`), or the source's last-known zone if it has already left the battlefield (`FR-ZON-004`). An activated ability whose effect reads it locks its source's zone when activated.
- Not this task: the distributed armor pool and the *Bivouac* card (T-0247, `queued`, which applies directive #157); continuous same-zone auras, which read the source's current zone (#141 (3)'s other half, the wave E auras of design note 6m; they stay `waiting` on #141's row).

## Inputs
- FRD 01 v3.28: line 398 (**"In the Same Zone" Is Locked on Placement (`SVY-R44`)**: the zone the source is in when the item is placed, that player's own for Reserves and the Headquarters Zone, or the source's last-known zone if it has already left; locked with the targets; what happens to the source afterwards doesn't change it; untargeted selections pick among the entities in that zone as they resolve); line 412 (`FR-TGT-003`: a zone restriction relative to the source is checked against the zone locked on placement, at both stages); line 390 (`FR-TGT-001` Zone criteria); line 34 (change log for `SVY-R44`); `FR-ACT-001` and `SVY-R36` (an activated ability needs a legal entity for each required target when activated).
- Survey: `docs/design/m2_starter_set_ability_survey.md` line 103 (`R44`) and line 183 (*Hammerhand Bivouac*: "zone locked on placement (`SVY-R44`)").
- Design note: `docs/architecture/m2-ability-model.md`, Deliverable 6m ("From wave B": "*Bivouac* fires #141 (3): its zone is locked from the source's zone on placement (`SVY-R44`), through #141 (1)'s `LockedZone`").
- Board directive #141 (3): "*Hammerhand Bivouac*'s task uses (1)'s placement zone for its untargeted pick." This task builds that half for every stack item; directive #157 then uses it for the pool.
- Code: `LockedZone` and `StackItem.Locked` ([stack.go:19-43](../../../pkg/engine/stack.go)); `placement.locked`, `stackItem`, `activatedPlacement` ([placement.go](../../../pkg/engine/placement.go)); `fireMoved` and `placeTrigger` ([trigger.go](../../../pkg/engine/trigger.go)); `inLocked`, `matchesFilterAt`, `legalTarget` ([filter.go](../../../pkg/engine/filter.go)); the second-stage check ([effects.go:44](../../../pkg/engine/effects.go)) and the untargeted readers ([effects.go:427](../../../pkg/engine/effects.go), [targeting.go:160](../../../pkg/engine/targeting.go)), which already pass `Locked`; `LastKnown` ([state.go:169-176](../../../pkg/engine/state.go)).
- Existing equivalents searched (D2.3): `git grep -n "LockedZone{" pkg` finds one construction, in `fireMoved`. The new lock builds the same value from an asset, so both go through one helper; a second `&LockedZone{Zone: a.Zone, Player: a.Controller}` would be a copy. `inLocked` and `matchesFilterAt` already read any `ZoneRef` against `Locked` and stay one reader. No effect walker exists for zone criteria.

### Shape (the engineer builds it)
- **`ability.ZoneRefSame`**, after `ZoneRefThat` so no existing value changes: "in the same zone": the zone the stack item's source was in when the item was placed (`SVY-R44`). Update the `ZoneRef` comment: one locked zone per stack item, set by the `ZoneRef` its effect reads.
- **`(Effect) ReadsZoneRef(r ZoneRef) bool`** in `ability.go`: whether any filter of e reads r: its Targets' filters, its Steps' filters, and each of its Modes, recursively. It is the one place that walks an effect for a zone criterion.
- **One lock helper** in `placement.go`, for example `zoneLock(a *Asset) *LockedZone`: a's zone and controller. `fireMoved` uses it for the destination (behavior unchanged).
- **Triggers:** at the start of `placeTrigger`, before the optional check and before modes and targets are chosen, if the effect reads `ZoneRefSame`, set `p.locked` to its source's zone: the source asset if it is on the battlefield, else its last-known asset (`LastKnown`), else nil (a trigger from a card in a pile, source 0, locks nothing, so "in the same zone" matches nothing). This holds for a move trigger too: its "in that zone" lock is its destination, and an effect reads one or the other.
- **Activated abilities:** `activatedPlacement` locks `a`'s zone when `act.Effect` reads `ZoneRefSame`, so the legality check (`placeable`, `SVY-R36`) and the declared item read the same zone.
- An effect that doesn't read `ZoneRefSame` locks nothing new: a deploy trigger still has `Locked == nil`, and "in that zone" without a move still matches nothing.
- `stack.go`: update the `LockedZone` and `StackItem.Locked` comments to say which items lock which zone.
- Behavior doesn't change for any printed card: no definition uses `ZoneRefSame` yet. `go tool task sim-compare -- autoloop/m2-wave-h` must be byte-identical (a gate here).

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
In a new file `pkg/engine/same_zone_test.go`, with synthetic cards. `that_zone_test.go` and `locked_zone_test.go` show the setup (deploy a card with `deployCard`, then `flushTriggers` and `resolveStack`, or `flushedItems` to see an item before it resolves). A filter `sameZone` is `ability.Filter{ZoneRef: ability.ZoneRefSame}` plus whatever kinds or allegiance each test needs.
- `TestFR_TGT_001_SameZoneAllStepReadsSourceZone`: a structure of player 0 with a deploy trigger, "each unit in the same zone gains 1 armor" (`SubjectAll`, units, no allegiance criterion), deployed into the Main Theater. Every unit in the Main Theater, either player's, gains 1 armor; units in either player's Reserves don't.
- `TestFR_TGT_001_SameZoneLocksOwnReserves`: the same card deployed into player 0's Reserves. Player 0's Reserves units gain 1 armor; player 1's Reserves units and every Main Theater unit don't (`FR-ZON-001`). The placed item's `Locked` is player 0's Reserves.
- `TestFR_TGT_001_SameZoneLockedOnPlacement` (`SVY-R44`): the trigger is placed while its source unit is in the Main Theater, and the source then moves to Reserves before the armor step resolves (either a first step `OpMoveZone` on `SubjectSelf` to Reserves, or a move between placement and resolution). Main Theater units still gain the armor; Reserves units don't.
- `TestFR_ZON_004_SameZoneLastKnownZone`: a unit of player 0 in Reserves with a death trigger, "each allied unit in the same zone gains 1 armor". It is killed: the trigger is placed after it has left the battlefield and locks its last-known zone, player 0's Reserves. Player 0's other Reserves units gain 1 armor; player 0's Main Theater units don't.
- `TestFR_TGT_001_SameZoneTargetOffersSourceZoneOnly`: a deploy trigger with one target slot, "chosen unit in the same zone gains 1 armor", deployed into the Main Theater: the target question offers only Main Theater units (both players'), none in Reserves.
- `TestFR_TGT_003_SameZoneTargetOutlivesSource`: the same targeted trigger; after the target is chosen in the Main Theater, the source leaves the battlefield before resolution. The target is still in the locked zone, so it stays legal and gains 1 armor.
- `TestFR_TGT_003_SameZoneTargetLeftZoneIsIllegal`: the chosen target moves to Reserves before resolution: it is illegal at the second check and gains nothing.
- `TestFR_ACT_001_SameZoneActivatedAbility` (`SVY-R36`): a unit of player 0 in Reserves with an activated ability, "Exhaust: chosen allied unit in the same zone gains 1 armor" (no "another", so the source itself qualifies). Activating it offers only player 0's Reserves units, and a Main Theater ally isn't offered. A second case: the source in the Main Theater as player 0's only Main Theater unit, with no "other" filter excluding it, is offered itself; with `Other: true` and no other ally in its zone, the ability can't be activated (`placeable` is false) even though allies stand in Reserves.
- `TestFR_TGT_001_SameZoneWithoutSourceMatchesNothing`: a trigger with source 0 (a card in a discard pile, as `pileSource` builds) whose step reads `sameZone` locks nothing and its step acts on nothing.
- **Existing tests:** every other test passes unchanged, in particular `TestFR_TGT_001_NoLockWithoutMove` and `TestFR_TGT_001_ThatZoneWithoutLockMatchesNothing` (an effect that doesn't read "in the same zone" locks nothing).

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected. The test author keeps any small builder in the test file, and may reuse `flushedItems`, `wantLocked`, `armorStep` and `allStep` from the existing locked-zone and that-zone tests.

## Allowed Files
- `pkg/engine/ability/ability.go` (`ZoneRefSame`, `Effect.ReadsZoneRef`)
- `pkg/engine/placement.go` (the lock helper; `activatedPlacement`)
- `pkg/engine/trigger.go` (`placeTrigger` locks the source's zone; `fireMoved` uses the helper)
- `pkg/engine/stack.go` (the `LockedZone` and `Locked` comments)

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run "TestFR_(TGT_001_SameZone|TGT_003_SameZone|ZON_004_SameZone|ACT_001_SameZone)"`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-compare -- autoloop/m2-wave-h` passes (no printed card reads "in the same zone" yet)

## Log
- 2026-11-15 lead-architect (wave H run, session 7): written for *Hammerhand Bivouac*'s zone (design note 6m, "From wave B"; #141 (3), stack-item half). Synced `main` twice: 2bfc7d6 added directive #157 (the Bivouac pool ruling), and a3a4d84 applied its outcome (FRD 01 v3.28, `SVY-R51`, `SVY-R52`), so #157 is now `open: before T-0247`. T-0247 (the pool and the card, `queued`) reads `StackItem.Locked` from this task, so it goes second; the D2 size rule keeps the two apart (four non-test files here, the pool's decision and step there). The lock is set only for an effect that reads `ZoneRefSame`, because `TestFR_TGT_001_NoLockWithoutMove` and `TestFR_TGT_001_ThatZoneWithoutLockMatchesNothing` require a deploy trigger to lock nothing. Set to `tests`.
- 2026-11-15 rules-test-author: tests written in pkg/engine/same_zone_test.go (not compiling until ability.ZoneRefSame exists; `go vet` reports only that). Tests: TestFR_TGT_001_SameZoneAllStepReadsSourceZone, TestFR_TGT_001_SameZoneLocksOwnReserves, TestFR_TGT_001_SameZoneLockedOnPlacement, TestFR_ZON_004_SameZoneLastKnownZone, TestFR_TGT_001_SameZoneTargetOffersSourceZoneOnly, TestFR_TGT_003_SameZoneTargetOutlivesSource, TestFR_TGT_003_SameZoneTargetLeftZoneIsIllegal, TestFR_ACT_001_SameZoneActivatedAbility (three subtests), TestFR_TGT_001_SameZoneWithoutSourceMatchesNothing. Helpers requested: none (test-local `szDeploy`, `szTargetDeployer`; reuses tz*, flushedItems, wantLocked, armorStep, allStep, pileTrigger, enactWith). Notes: the source-0 case uses a discard-pile trigger (HookDiscardedFromHand, ActiveZone discard) enacted via a Maneuver, since no `pileSource` helper exists; the last-known test kills its source with a deploy-trigger OpKill on a chosen ally; the "locked on placement" test uses a unit that moves itself (a structure cannot move). No cases left out.
- 2026-11-15 lead-architect (wave H run, session 7): tests reviewed (58b1244); set back to `tests`. Notes:
  1. `TestFR_TGT_001_SameZoneWithoutSourceMatchesNothing` passes whatever the engine does, even if the pile trigger never fires: every assertion is "0 armor", which also holds when nothing resolved. Give the trigger a first step whose result the test can see (for example `scratchN(1)`, then check that player 0 gained a Scratch Company in Reserves, or a draw and the hand size), and assert it, so the test shows the trigger resolved and its "in the same zone" step acted on nothing. Everything else stays as it is.
  The other eight tests match the Tests list, line 398 (`SVY-R44`), `FR-TGT-003` (line 412), `FR-ZON-004` and `SVY-R36`. The deviations the Log names (a pile trigger through a Maneuver's discard, a deploy trigger that kills the source, a unit that moves itself) are fine.
- 2026-11-15 rules-test-author: review note 1 fixed. TestFR_TGT_001_SameZoneWithoutSourceMatchesNothing's pile trigger now has a first step drawN(1); the test fails unless player 0's hand is 1 (the trigger resolved), then asserts the same-zone armor step reached nothing. Nothing else changed.
- 2026-11-15 lead-architect (wave H run, session 7): tests reviewed again (c414532). Note 1 is applied: the pile trigger draws 1 first, and the test checks player 0's hand is 1 (the hand was emptied by the Maneuver and its discard) before it checks that no unit gained armor. Nothing else changed. The nine tests match the Tests list. Set to `ready`.
