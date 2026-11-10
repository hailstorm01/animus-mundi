---
id: T-0204
title: Chainbreaker Phalanx (wave F card, flash)
kind: card
tier: flash
wave: F
status: review
branch: m2/autoloop-wave-f/T-0204-chainbreaker-phalanx
depends_on: [T-0203]
claims: [card:Chainbreaker Phalanx]
cards: [Chainbreaker Phalanx]
requirements: [FR-ZON-002, FR-TRG-004]
---

## Goal
*Chainbreaker Phalanx* (a Path deck card, 2 copies) executes: "Whenever this unit moves to the Main Theater, you may move chosen allied unit from your Reserves to the Main Theater. That unit gains 2 armor." It only combines shapes that are `done`: T-0203's `HookMoves` with `ToZone`, an `Optional` trigger (as *Arkha, Sun's Fury*), a target filtered by zone, `OpMoveZone` on a target and `OpAddArmor` (as *Chainbreaker's Tenacity*).

Covers:
- The definition in `pkg/engine/cards/path.go`, as under Definition below.
- The name off `cards.NotYetExecutable`.
- The self-hook list in the `pkg/engine/doc.go` package comment (lines 18-19) names `HookMoves`, after `HookCombatKill`. T-0203 added the hook, and `doc.go` was not one of its Allowed Files (T-0203 review).

## Inputs
- Card (`data/cards/units-and-structures-path.csv` line 24): Path Unit, Infantry, Chainbreaker, R4 T0, deploys to Reserves, power 4, health 5.
- Survey row: `docs/design/m2_starter_set_ability_survey.md` line 199 ("self moves to Main Theater (may)", "can chain into another Phalanx"); `R1` at line 59.
- FRD 01: `FR-ZON-002` (line 230), above all "What Counts as Moving (`SVY-R1`)" (line 233), which names this card; `FR-TGT-003` (line 388): the target is verified on placement and again before resolving, and between the two a step checks only that the chosen asset is still on the battlefield, not the slot's zone, so the armor step still applies to the unit the move step just moved (`SVY-R25`, `SVY-R26`); "Complete Target Invalidation" (same section): with no legal target the clause fizzles. `FR-MOD-001` §7 (line 417) and the ruling summary at line 81: a "you may" trigger is asked as it is placed, before its targets, and a decline or timeout means it is not placed (`SVY-R9`). `FR-TRG-003` (line 435) and `FR-TRG-004` (line 441): a trigger fired while an item resolves is placed after it; a unit moved by a resolving trigger fires its own. `FR-LAY-002` (line 481): armor.
- Rulings: `SVY-R1`, `SVY-R9`, `SVY-R25`, `SVY-R26`.
- Design note: `docs/architecture/m2-ability-model.md`, Key types (`Trigger`, `Hook`).

### Definition (the engineer adds it to `pathDefs`, between *Chainbreaker Irregulars* and *Chainbreaker Sappers*)
- `Text`: the CSV text above, as one line.
- `Triggers`: one, `Hook: ability.HookMoves`, `ToZone: ability.ZoneMainTheater`, `Optional: true`.
  - `Effect.Targets`: one exactly-one slot, `Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit, Zone: ability.ZoneReserves}`.
  - `Effect.Steps`: `{Op: ability.OpMoveZone, To: ability.ZoneMainTheater, Subject: ability.SubjectTarget, Target: 0}`, then `{Op: ability.OpAddArmor, Amount: ability.Value{N: 2}, Subject: ability.SubjectTarget, Target: 0}`.
- Remove the name from `cards.NotYetExecutable`.
- `doc.go`: in the list "Triggered abilities on self hooks (HookDeploy, HookKilled, HookAttacks, HookDamaged, HookExhausted, HookCombatKill)", add `HookMoves` after `HookCombatKill`. No other change to the comment.

A lead probe with this definition on the run branch at 03d19c7, since reverted, passes `go tool task check`, and `go tool task sim` stays clean: Path 40/50 main cards executable, P0 70.0% / P1 28.5% / draws 1.5% over 200 games. So `sim-compare` will differ here and isn't a gate.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
In a new file `pkg/engine/cards_path_phalanx_test.go`, from the CSV text, loading the real card (`s.card`), in the style of `cards_path_dancers_test.go`. Move the Phalanx with a March (`ActMarch`), or with an `OpMoveZone` step as `move_hook_test.go` does. Answer the "you may" and the target with `s.script`.
- `TestCard_ChainbreakerPhalanx_MarchMovesChosenUnitAndArmorsIt` (`FR-ZON-002`, `FR-TGT-003`, `SVY-R25`): the Phalanx and another allied unit X in Reserves. The Phalanx marches, its controller accepts and chooses X. After the trigger resolves, X is in the Main Theater with 2 armor, and the Phalanx gained no armor.
- `TestCard_ChainbreakerPhalanx_DeclineIsNotPlaced` (`SVY-R9`, `FR-MOD-001` §7): as above, but the controller declines; in a second case, nobody answers (a timeout). Either way no target is asked for, X stays in Reserves with no armor.
- `TestCard_ChainbreakerPhalanx_MarchToReservesFiresNothing` (`FR-ZON-002`): the Phalanx in the Main Theater marches to Reserves, with X in Reserves. No "you may" is asked, and X doesn't move.
- `TestCard_ChainbreakerPhalanx_DeployFiresNothing` (`SVY-R1`): deploying the Phalanx from hand asks nothing and moves nothing, with X in Reserves.
- `TestCard_ChainbreakerPhalanx_OnlyAlliedUnitsInReserves` (`FR-TGT-001`, `FR-TGT-003`): the Phalanx marches with no allied unit in Reserves, but an opposing unit in its controller's Reserves and another allied unit already in the Main Theater. Neither is offered as the target, nothing moves and nothing gains armor. Don't assert whether the "you may" is asked.
- `TestCard_ChainbreakerPhalanx_ChainsIntoAnotherPhalanx` (`FR-TRG-004`, `FR-TRG-003`, survey line 199): Phalanx A marches, and its trigger moves Phalanx B from Reserves. B gains 2 armor, and its own trigger fires and is placed after A's has resolved. Accepted, it moves a third allied unit C from Reserves to the Main Theater with 2 armor.
- **Existing tests:** every other test passes unchanged, including `TestStarterSet_Executable` with the name removed, `TestView_Case8_ExecutableFlag`, `TestProgress_Case4_Coverage`, `TestEvents_Case4_EveryActionKind` and the T-0203 tests in `move_hook_test.go`.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected. The test author keeps any small builder in the test file.

## Allowed Files
- `pkg/engine/cards/path.go` (the definition)
- `pkg/engine/cards/cards.go` (remove the name from `NotYetExecutable`)
- `pkg/engine/doc.go` (`HookMoves` in the self-hook list)

## Done When
- [x] The card tests pass: `go test ./pkg/engine -run TestCard_ChainbreakerPhalanx`
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks
- [x] The card loads with no unsupported lines (`TestStarterSet_Executable` passes with the name removed)

## Log
- 2026-11-09 lead-architect (wave F run, session 2): split off T-0203 under the D2 size rule; `queued` on it.
- 2026-11-09 lead-architect (wave F run, session 3): T-0203 is `done` (merge c5bc636). Added the `doc.go` hook-list line to Covers, from T-0203's review. It goes here, not in T-0205, so T-0205 stays within four non-test files.
- 2026-11-09 lead-architect (wave F run, session 4): written against 03d19c7, with the lead's probe above. No `main` commits to sync. Board read twice (D2.0): #140 is `open: before` the *Battle Trance* task only, and no other `open: before` directive applies. Waiting triggers checked, none fired: #6, #7 (no mode change), #15 (no turn hook), #71 (no each-player effect), #91 (2) (no mode cost), #106 (no criteria search), #120 (no mandatory discard cost), #128 (no discard deploy), #136 (no off-battlefield trigger), #137 (no keyword grant). Searched `pkg/engine` and the task archive: no new shape; every part of the Definition exists. Set to `tests`.
- 2026-11-09 rules-test-author: wrote `pkg/engine/cards_path_phalanx_test.go`: TestCard_ChainbreakerPhalanx_MarchMovesChosenUnitAndArmorsIt, _DeclineIsNotPlaced (subtests decline, timeout), _MarchToReservesFiresNothing, _DeployFiresNothing, _OnlyAlliedUnitsInReserves, _ChainsIntoAnotherPhalanx. Helpers requested: none (`phalanxChooser`, `countKind`, `phalanxScene` live in the test file). Left out: no case asserts whether a target question is asked when exactly one legal target exists (SVY-R3 auto-choice), so the chooser answers it if asked. `go vet` clean; the tests fail now because the card is in `NotYetExecutable` (Deploy and March are rejected or fire nothing). Set to `tests-review`.
- 2026-11-09 lead-architect (wave F run, session 4): tests approved (D4, round 1) against b080d72. Each test checks the outcome its cited rule gives: `FR-ZON-002` and `SVY-R1` (only a move into the Main Theater fires, a deploy doesn't), `SVY-R9` (asked once on placement; a decline or timeout places nothing and asks no target), `FR-TGT-001` (only allied units in your Reserves are offered), `FR-TGT-003` with `SVY-R25`/`SVY-R26` (the armor lands on the unit the move step just moved), and `FR-TRG-003`/`FR-TRG-004` (B's trigger is placed alone, after A's resolves). No assertion on implementation details. A lead probe with the Definition, since reverted, passes all six and `go test ./pkg/...`. Set to `ready`, dispatching `kernel-engineer` (flash).
- 2026-11-09 kernel-engineer (flash): added Chainbreaker Phalanx definition to pathDefs in pkg/engine/cards/path.go between Chainbreaker Irregulars and Chainbreaker Sappers with HookMoves, ToZone: ZoneMainTheater, Optional: true, target filter allied unit in Reserves, and steps OpMoveZone to ZoneMainTheater and OpAddArmor (2); removed Chainbreaker Phalanx from NotYetExecutable in pkg/engine/cards/cards.go; added HookMoves after HookCombatKill to the comment in pkg/engine/doc.go. Reused HookMoves from T-0203, existing Optional trigger shape, OpMoveZone, and OpAddArmor. Ran `go test ./pkg/engine -run TestCard_ChainbreakerPhalanx` (pass), `go tool task check` (all vet, tests, and linter checks pass clean), and `go tool task sim` (Path 40/50 main cards executable, P0 70.0% / P1 28.5% / draws 1.5% over 200 games, all invariants held). Set to `review`.
