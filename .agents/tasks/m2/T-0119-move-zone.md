---
id: T-0119
title: An effect can move a unit to its controller's Reserves or the Main Theater (`MOVE_ZONE`), keeping its state (FR-ZON-002, FR-ZON-004, SVY-R26, directive #88 part 4, wave C shape, sonnet)
kind: shape
tier: sonnet
wave: C
status: review
branch: m2/T-0119-move-zone
depends_on: [T-0118]
claims: [op:OpMoveZone, field:Step.To, event:move, func:battlefieldZone]
cards: []
requirements: [FR-ZON-002, FR-ZON-004, FR-TGT-003, FR-MOD-001]
---

## Goal
A step can move the units it acts on to a battlefield zone: *"Move chosen unit to its controller's Reserves"* (*Ensnaring Mist*) and *"Move all allied Scratch Company units to the Main Theater"* (*Rapturous Sermon*, a deck card). A moved unit stays the same asset: its damage, Armor, modifiers and readiness are kept, and a later step of the same effect still acts on it even though it no longer matches its slot's zone (`SVY-R26`). This is directive #88's test (e). The two cards are the next task, a card task: with them this one would have six non-test files (D2 size rule).

## Inputs
- FRD 01:
  - `FR-ZON-002` (line 211): Headquarters and Structures are immune to, and ineligible targets for, any effect that moves an asset between battlefield zones. Units move between their controller's Reserves and the Main Theater by March or by card effects (*Ensnaring Mist*). "What Counts as Moving (`SVY-R1`)": only March and an effect that says to move a unit to a zone; deploying is not moving.
  - `FR-ZON-004`, "Intra-Battlefield Movement Preservation" (line 220): moving a unit between Reserves and the Main Theater does not reset its state. Damage, Armor and modifiers are kept.
  - `FR-TGT-003`, "Targets Removed During Resolution (`SVY-R25`, `SVY-R26`)" (line 369): between the two verification stages a step checks only that each chosen asset is still on the battlefield, so an earlier step that moves a target to another battlefield zone doesn't stop a later step from acting on it.
  - `FR-MOD-001` §6 (line 386): fixed order goes zone by zone, each zone in the order its assets entered it (`SVY-R7`, `SVY-R16`). A unit that moves enters its new zone when it moves.
  - `FR-LAY-001`, "One-Shot Modifier Snapshot Locking" (line 455): *Rapturous Sermon*'s +2 goes to the units that match when that step resolves. That is the card task's; nothing here.
- Rulings that apply: `SVY-R1`, `SVY-R7`, `SVY-R16`, `SVY-R26`.
- Survey rows: *Ensnaring Mist* (`docs/design/m2_starter_set_ability_survey.md` line 130) and *Rapturous Sermon* (line 133).
- Board directive #84 (applied in T-0108): one `moveOnBattlefield`, which `MOVE_ZONE` reuses. Directive #88, test (e).
- Kernel today:
  - `moveOnBattlefield` (`zones.go:162`): moves an asset between battlefield zones under its controller, keeps its state, and gives it a new arrival number (`SVY-R7`). Its only caller is `resolveMarch` (`stack.go:127`), which then logs `EventMarch`.
  - `runStep` (`effects.go:63`): one `case` per `ability.Op`. `stepAssets` (`effects.go:301`) gives the assets of any asset subject. `stepRelocate` (`effects.go:200`) logs `EventLeave` with `Player: -1`, `Target` the asset and `Zone` where it went.
  - `ability.Op` (`ability/ability.go:177`), last value `OpKill`; `ability.Step` (`ability.go:206`); `ability.Zone` (`ability.go:108`): `ZoneAny`, `ZoneReserves`, `ZoneMainTheater`, `ZoneDiscard`.
  - `deployedTo` (`trigger.go:128`): the one place that compares an `ability.Zone` with an asset's engine `Zone` today, as a two-case switch.
  - Event kinds (`events.go:233`), every one listed in `TestEventKinds_ValuesAreTheSchema` (`eventkinds_test.go`). ADR-0005: a new kind is additive.

### Shape (lead decisions, binding)
- **`ability.OpMoveZone`**, appended after `OpKill` so no existing value changes. Doc comment: move each Subject unit to `Step.To`, in its controller's battlefield (FR-ZON-002, FR-ZON-004). A unit already there, a Structure and a Headquarters don't move.
- **`ability.Step.To ability.Zone`**: OpMoveZone only, `ZoneReserves` or `ZoneMainTheater`. The Reserves are the unit's controller's, as March's are.
- **One zone mapping.** Add `func battlefieldZone(z ability.Zone) (Zone, bool)` in `trigger.go`, beside `deployedTo`: `ZoneReserves` and `ZoneMainTheater` map to the engine's zones (true); anything else is false. `deployedTo` is rewritten on it with no change in behavior (no battlefield zone → true; otherwise `a.Zone` equals it). The move step uses it too, so no second switch maps ability zones to engine zones.
- **`stepMoveZone(r *resolution, st ability.Step) int`** in `effects.go`, with a `runStep` case. With `battlefieldZone(st.To)` false it does nothing and returns 0. Otherwise, for each asset of `stepAssets(r, st)`, in order: skip it unless its printed type is a unit (a face-down Scratch Company is one), and skip it if it is already in the destination. Else `g.moveOnBattlefield(a, to)`, then log one `EventMove` event: `Player: -1`, `Target: a.ID`, `Zone: to.String()`, text `"%s moves to %s"`. Like `stepRelocate`'s `leave`, it carries no `Card`. It returns how many units moved, so a unit that was already there doesn't count for "if you do".
- **`EventMove = "move"`** in `events.go`, after `EventMarch`. Its comment says an effect moved an asset between battlefield zones; March keeps `EventMarch`.
- **No hook.** Moving fires no trigger yet. Wave F's "moves to a new zone" hook (*Broken Horn Armorer*, *Chainbreaker Phalanx*) will fire for March and this step alike. Moving is not deploying (`SVY-R1`), so it fires no deploy trigger.
- **Out of scope:**
  - The *Ensnaring Mist* and *Rapturous Sermon* definitions, and their removal from `cards.NotYetExecutable`: the next task.
  - Distinctness between two slots (*Ambush Tactics*' *"another chosen unit"*): it waits for a ruling (design note Open Question 16).
  - *Ensnaring Mist*'s -1 is today's `OpModifyStat` with `Amount: Value{N: -1}`. Power clamps at 0.
- **Sims:** no card moves anything yet, so `go tool task sim-compare -- autoloop/m2-wave-c` should pass byte for byte. If it differs, don't change code to make it match: report the first differing game in the Log and hand off.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
Synthetic Maneuvers only, in a new file `pkg/engine/move_zone_test.go`, reusing `maneuver_test.go`'s `maneuver`, `steps`, `enactAction` and `lastDiscard` helpers. "Move S to Z" below is one `OpMoveZone` step with `To: Z`. Every existing test passes unchanged, except the one named edit at the end.
- `TestFR_TGT_003_MovedTargetStillTakesLaterSteps` (#88 (e)): one slot, "chosen unit in the Main Theater" (`Kinds: KindUnit`, `Zone: ZoneMainTheater`). Step 1 moves the slot to `ZoneReserves`, step 2 gives the slot -1 power until end of turn. The chosen unit X is the opponent's, with power 2 or more. X ends in its controller's (the opponent's) Reserves with power one less than printed: the -1 applies though X no longer matches the slot's zone (`SVY-R26`).
- `TestFR_ZON_004_MovedUnitKeepsItsState`: an allied unit X in the Reserves, with damage (not lethal), Armor, a +1 power modifier and Exhausted, is moved to the Main Theater by "move chosen unit to the Main Theater". `g.Asset` of X's ID is still X, in the Main Theater, with the same damage, Armor and power, still Exhausted.
- `TestFR_ZON_002_MoveAllAlliedUnits`: "move all allied units to the Main Theater" (`SubjectAll`, filter allied units). Two allied units in the Reserves move. An allied unit already in the Main Theater and an opposing unit in its Reserves stay where they are. There is exactly one `move` event per moved unit, with `Target` its ID and `Zone` the Main Theater, and none for the others.
- `TestFR_ZON_002_StructuresAndHeadquartersAreNotMoved`: "move all assets to the Main Theater" (`SubjectAll`, empty filter). An allied Structure in the Reserves and both Headquarters stay in their zones, with no `move` event. A unit in the Reserves does move.
- `TestFR_ZON_002_MovingIsNotDeploying` (`SVY-R1`): a unit with a deploy trigger (synthetic, for example "When this unit deploys, draw a card") is moved by the step. No `deploy` event follows the move, no trigger is placed, and the controller's hand shows no draw.
- `TestFR_ZON_002_AUnitAlreadyThereDoesNotMove`: "move chosen unit to the Main Theater", then "if you do: draw a card" (`IfDone`). The chosen unit is already in the Main Theater. No `move` event is logged, and the hand shows no draw.
- `TestFR_MOD_001_MovedUnitIsLastInItsNewZone` (`SVY-R7`, §6): allied units Y, then X, then Z were deployed in that order, Y and Z to the Main Theater and X to the Reserves. The step moves X to the Main Theater. A following mandatory target choice among allied Main Theater units, answered by timeout, picks Y (not X, though X is older than Z), and with Y gone first it would pick Z before X. Pin at least the first part; reuse `fallback_order_test.go`'s way of timing a choice out.
- **Authorized edit to an existing test:** add `{EventMove, "move"}` after `{EventMarch, "march"}` in `TestEventKinds_ValuesAreTheSchema` (`pkg/engine/eventkinds_test.go`), in the `test(T-0119)` commit.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected.

## Allowed Files
- `pkg/engine/ability/ability.go` (`OpMoveZone`, `Step.To`)
- `pkg/engine/effects.go` (`stepMoveZone`, its `runStep` case)
- `pkg/engine/events.go` (`EventMove`)
- `pkg/engine/trigger.go` (`battlefieldZone`, `deployedTo` on it)

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run 'TestFR_TGT_003|TestFR_ZON_002|TestFR_ZON_004|TestFR_MOD_001|TestEventKinds'`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-compare -- autoloop/m2-wave-c` passes (no card moves anything yet; see Shape, Sims)
- [ ] `moveOnBattlefield` is still the one function that moves an asset between battlefield zones, and `battlefieldZone` the one mapping from `ability.Zone` to `Zone`.

## Log
- 2026-10-30 lead-architect: wave C run, session 5. No `main` commits to sync (`git log autoloop/m2-wave-c..main` is empty). Not `WAVE-COMPLETE`: 22 deck cards are still on `NotYetExecutable`, and #88 is `open`. Directives read on the run branch: #88's next pieces were distinctness between two slots, then `MOVE_ZONE` (test (e)). Distinctness needs a ruling FRD 01 doesn't give: whether a mode whose two exactly-one slots are *"chosen unit"* and *"another chosen unit"* is legal with only one legal unit (`SVY-R28` counts slots one at a time), and what a timeout picks for the first slot when only one choice leaves the second slot fillable. Neither the survey nor FRD 01 (`SVY-R11` covers one "up to N" clause, not two slots) settles it, so it is recorded as design note Open Question 16, to be asked in the survey's question format before the distinctness task is written, in the same round as #61's fifth question (both block only *Ambush Tactics*; in a wave run that is a `STOP`). This task needs no ruling and goes first: distinctness only has to come before the *Ambush Tactics* card task. #14 and #20 (b) stay open for the card tasks and *Steel Rain*. Waiting triggers checked: #58 (no exhaust in `claims:`), #59 (2) (no repeated targets), #61's fifth question (`cards:` is empty), #71 (no each-player effect), #15 and #8 (no hook or `deck.go` change). None has fired. Duplicates searched: `moveOnBattlefield` is the one battlefield move (#84) and is reused. No step moves an asset today, and `EventMarch` stays the March action's. The ability-to-engine zone switch in `deployedTo` becomes `battlefieldZone`, shared by the move step. The two cards are left out to keep four non-test files. `doc.go`'s Steps list (and its "Not built yet: Maneuvers", stale since T-0116) is not in this task, so it stays within the size rule; reported to the user. Status `tests`; dispatching `rules-test-author`.
- 2026-10-30 rules-test-author: tests written in `pkg/engine/move_zone_test.go`: `TestFR_TGT_003_MovedTargetStillTakesLaterSteps`, `TestFR_ZON_004_MovedUnitKeepsItsState`, `TestFR_ZON_002_MoveAllAlliedUnits`, `TestFR_ZON_002_StructuresAndHeadquartersAreNotMoved`, `TestFR_ZON_002_MovingIsNotDeploying`, `TestFR_ZON_002_AUnitAlreadyThereDoesNotMove`, `TestFR_MOD_001_MovedUnitIsLastInItsNewZone`. Authorized edit: `{EventMove, "move"}` added to `TestEventKinds_ValuesAreTheSchema`. Test-local helpers (in the new file, not scenario_test.go): `moveStep`, `(*scene).choosesAssets`, `(*scene).enactFree`. Helpers requested: none. `go vet` fails only on `EventMove`, `OpMoveZone` and `Step.To`; `cardtown lint rules` passes. Move events are matched by the literal kind "move". The ordering test's mandatory choices are two `SubjectPick` kill steps answered by timeout (Y, then Z, X survives), since chosen slots are picked before any step runs. Left out: nothing.
- 2026-10-30 lead-architect: tests approved. Each of the seven checks its outcome against FRD 01: `SVY-R26` (the -1 lands on a target moved out of its slot's zone, in its controller's Reserves), `FR-ZON-004` (same asset, damage, Armor, power and Exhausted kept), `FR-ZON-002` (allied units move, one `move` event each; a unit already there, an opposing unit, a Structure and both Headquarters don't), `SVY-R1` (no deploy event or trigger), the already-there case with its "if you do" draw, and `SVY-R7`/§6 (Y, X, Z by ID but Y, Z, X by arrival after the move, so two timed-out kill picks take Y then Z). The asserts read zones, stats, the log and deck counts, not internals. The `TestEventKinds` edit is the named one, in the test commit (d33573c). `go vet ./pkg/engine` fails only on `EventMove`, `OpMoveZone` and `Step.To`. Status `ready`; dispatching `kernel-engineer`.
- 2026-10-30 kernel-engineer: implemented. Added `ability.OpMoveZone` (after `OpKill`) and `Step.To`; `EventMove = "move"` in `events.go`; `battlefieldZone` in `trigger.go`, with `deployedTo` rewritten on it (same behavior); `stepMoveZone` and its `runStep` case in `effects.go`. Reused: `moveOnBattlefield` (the only mover), `stepAssets`, `logEvent`, the `a.Printed().Type != TypeUnit` check as in `actions.go`. No hook, no deploy trigger. Commands: the listed `go test ./pkg/engine -run ...` passes; `go tool task check` passes; `go tool task sim` runs clean; `go tool task sim-compare -- autoloop/m2-wave-c` passes byte for byte. Look closely at: the event text `"%s moves to %s"` formats the asset and the engine `Zone` with %s.
