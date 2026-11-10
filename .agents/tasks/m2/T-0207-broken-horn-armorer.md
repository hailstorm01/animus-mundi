---
id: T-0207
title: Broken Horn Armorer (wave F deck card, flash; #141 (2))
kind: card
tier: flash
wave: F
status: review
branch: m2/autoloop-wave-f/T-0207-broken-horn-armorer
depends_on: [T-0206]
claims: [card:Broken Horn Armorer]
cards: [Broken Horn Armorer]
requirements: [FR-ZON-002, FR-TGT-001, FR-TGT-003, FR-ZON-004, FR-LAY-002]
---

## Goal
*Broken Horn Armorer* (a Path deck card, 2 copies) executes: "Whenever this unit moves to a new zone, chosen asset in that zone gains 1 armor." (reworded by `SVY-R45`). Directive #141 (2). It only combines shapes that are `done`: T-0203's `HookMoves` (with `ToZone` left at any zone), T-0205's zone locked from the move's destination, T-0206's `Filter.ZoneRef` with `ZoneRefThat`, and `OpAddArmor` on a target (as *Chainbreaker Phalanx*).

Covers:
- The definition in `pkg/engine/cards/path.go`, as under Definition below.
- The name off `cards.NotYetExecutable`.

## Inputs
- Card (`data/cards/units-and-structures-path.csv` line 8): Path Unit, Devilhorn Specialist, R2 T0, deploys to Reserves, power 2, health 2.
- Survey row: `docs/design/m2_starter_set_ability_survey.md` line 200 ("self moves zone → `ADD_ARMOR 1` chosen asset in that zone (the move's destination)"); `R1` at line 59, `R45` at line 104.
- FRD 01: `FR-ZON-002` (line 234), above all "What Counts as Moving (`SVY-R1`)" (line 237), which names this card: only a March or an effect that says "move" fires it, a deploy doesn't. `FR-TGT-001` (line 370), "'That Zone' After a Move (`SVY-R45`)" (line 383): "that zone" is the triggering move's destination, fixed by the event, whatever happens to the source later. The text has no "other" and no allegiance, so any asset in that zone is legal, the Armorer itself and enemy assets included. `FR-ZON-001` (line 220): Reserves are each player's own, so a move to Reserves locks its controller's Reserves. `FR-TGT-003` (line 394): the target is checked on placement and again before resolving; a target that left that zone is illegal and the clause fizzles (line 400). `FR-ZON-004` (line 242): the source leaving doesn't change the locked zone. `FR-TRG-004` (line 447): a unit moved by a resolving effect fires its own trigger. `FR-LAY-002` (line 487): armor.
- Rulings: `SVY-R1`, `SVY-R45`, and directive #141 (2).
- Design note: `docs/architecture/m2-ability-model.md`, Key types (`Trigger`, `Hook`, `Filter`).

### Definition (the engineer adds it to `pathDefs`, between *Boneyard Scavenger* and *Chainbreaker Irregulars*)
- `Text`: the CSV text above, as one line.
- `Triggers`: one, `Hook: ability.HookMoves` (no `ToZone`, no `Optional`).
  - `Effect.Targets`: one exactly-one slot, `Filter{ZoneRef: ability.ZoneRefThat}` (no other criterion).
  - `Effect.Steps`: `{Op: ability.OpAddArmor, Amount: ability.Value{N: 1}, Subject: ability.SubjectTarget, Target: 0}`.
- Remove the name from `cards.NotYetExecutable`.

A lead probe with this definition on the run branch at a5f66ec, since reverted, passes `go tool task check`, and `go tool task sim` stays clean: Path 42/50 main cards executable, P0 68.0% / P1 30.5% / draws 1.5% over 200 games. A probe test marched the Armorer to the Main Theater and armored a chosen enemy unit there, then moved it back to Reserves and armored a chosen allied unit in its Reserves. So `sim-compare` will differ here and isn't a gate.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
In a new file `pkg/engine/cards_path_armorer_test.go`, from the CSV text, loading the real card (`s.card`), in the style of `cards_path_phalanx_test.go`. Move the Armorer with a March (`ActMarch`), or with `s.g.moveOnBattlefield` followed by `s.g.flushTriggers()` and `s.g.resolveStack()` when something must happen between placement and resolution, as `that_zone_test.go` does. Answer the target with a chooser (a small builder in the test file; `phalanxChooser` is in another test file and may be reused).
- `TestCard_BrokenHornArmorer_MarchToMainTheaterArmorsChosenAsset` (`FR-ZON-002`, `SVY-R45`, `FR-TGT-001`, `FR-LAY-002`): the Armorer and an allied unit R in player 0's Reserves, an enemy unit X in the Main Theater. The Armorer marches. The target offered is exactly the Main Theater's assets (the Armorer and X), not R. Choosing X gives X 1 armor, and the Armorer and R gain none.
- `TestCard_BrokenHornArmorer_MarchToReservesOffersOwnReserves` (`SVY-R45`, `FR-ZON-001`): the Armorer in the Main Theater, an allied unit R in player 0's Reserves, an enemy unit E in player 1's Reserves and an enemy unit X in the Main Theater. The Armorer marches to Reserves. The target offered is exactly the Armorer and R, not E or X. Choosing the Armorer itself gives it 1 armor (the text has no "other").
- `TestCard_BrokenHornArmorer_DeployFiresNothing` (`SVY-R1`): deploying the Armorer from hand asks nothing, and no asset gains armor.
- `TestCard_BrokenHornArmorer_MovedByEffectFires` (`FR-ZON-002`, `FR-TRG-004`): the real *Chainbreaker Phalanx* and the Armorer in player 0's Reserves. The Phalanx marches, its controller accepts and chooses the Armorer, which moves to the Main Theater and gains 2 armor. The Armorer's own trigger fires; choosing the Phalanx gives the Phalanx 1 armor.
- `TestCard_BrokenHornArmorer_TargetOutlivesSource` (`SVY-R45`, `FR-TGT-003`, `FR-ZON-004`): with X chosen in the Main Theater after the Armorer moved there, and before the trigger resolves, (a) the Armorer moves back to Reserves, (b) the Armorer leaves the battlefield. Either way X still gains 1 armor.
- `TestCard_BrokenHornArmorer_TargetLeftZoneIsIllegal` (`FR-TGT-003`): with X chosen in the Main Theater after the Armorer moved there, X moves to its controller's Reserves before the trigger resolves. X gains no armor, and no other asset does.
- **Existing tests:** every other test passes unchanged, including `TestStarterSet_Executable` with the name removed, `TestView_Case8_ExecutableFlag`, `TestProgress_Case4_Coverage`, `TestEvents_Case4_EveryActionKind`, the *Chainbreaker Phalanx* tests and the T-0203, T-0205 and T-0206 tests.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected. The test author keeps any small builder in the test file.

## Allowed Files
- `pkg/engine/cards/path.go` (the definition)
- `pkg/engine/cards/cards.go` (remove the name from `NotYetExecutable`)

## Done When
- [x] The card tests pass: `go test ./pkg/engine -run TestCard_BrokenHornArmorer`
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks
- [x] The card loads with no unsupported lines (`TestStarterSet_Executable` passes with the name removed)

## Log
- 2026-11-10 lead-architect (wave F run, session 5): split off T-0205 under the D2 size rule (#141 (2)); `queued` on T-0206.
- 2026-11-10 lead-architect (wave F run, session 7): written against a5f66ec, with the lead's probe above. No `main` commits to sync, and the scratch folder was empty. Board read twice (D2.0): #140 is `open: before` the *Battle Trance* task only, #141 is `open` with (2) in this task, and no other `open: before` directive applies. Waiting triggers checked, none fired: #6, #7 (no mode change), #15 (no turn hook), #71 (no each-player effect), #91 (2) (no mode cost), #106 (no criteria search), #120 (no mandatory discard cost), #128 (no discard deploy), #136 (no off-battlefield trigger), #137 (no keyword grant), #141 (3) (no "in the same zone" card, no aura). Searched `pkg/engine` and the task archive: no new shape; every part of the Definition exists. Set to `tests`.
- 2026-11-10 rules-test-author: wrote `pkg/engine/cards_path_armorer_test.go` (6 tests, all fail until the card is defined; the file compiles): `TestCard_BrokenHornArmorer_MarchToMainTheaterArmorsChosenAsset`, `_MarchToReservesOffersOwnReserves`, `_DeployFiresNothing`, `_MovedByEffectFires`, `_TargetOutlivesSource` (two subtests), `_TargetLeftZoneIsIllegal`. Reuses `tzScript`, `tzWantOffered`, `tzWantArmor` (that_zone_test.go) and `phalanxChooser` (cards_path_phalanx_test.go). Helpers requested: none. Cases left out: none. Set to `tests-review`.
- 2026-11-10 lead-architect (wave F run, session 7): tests approved (D4, round 1) against f8ec408. Each test checks the outcome its cited rule gives: `SVY-R1` (a deploy asks nothing), `SVY-R45` with `FR-TGT-001` (the offered targets are exactly the move destination's assets, the Armorer itself and enemy assets included) and `FR-ZON-001` (a move to Reserves offers only its controller's own Reserves), `FR-ZON-002`/`FR-TRG-004` (a move by the Phalanx's effect fires the Armorer), `FR-TGT-003`/`FR-ZON-004` (the locked zone outlives the source; a target that left that zone fizzles). The offered-options checks read the decision put to the player, not engine internals. A lead probe with the Definition, since reverted, passes all six (seven with the subtests), and `go test ./pkg/...` passes. Set to `ready`, dispatching `kernel-engineer` (flash).
- 2026-11-10 kernel-engineer: added card definition for Broken Horn Armorer in pkg/engine/cards/path.go using HookMoves, Filter.ZoneRefThat, and OpAddArmor (SubjectTarget, amount 1); removed Broken Horn Armorer from NotYetExecutable in pkg/engine/cards/cards.go. Reused shapes: HookMoves (T-0203), ZoneRefThat (T-0206), OpAddArmor (FR-LAY-002). Ran `go test ./pkg/engine -run TestCard_BrokenHornArmorer` (passed), `go tool task check` (all format, vet, test, linter suites passed), `go tool task sim` (clean: Path 42/50 main cards executable, 200 games P0 68.0% / P1 30.5% / draws 1.5%), and `go test ./pkg/engine -run TestStarterSet_Executable` (passed). Set to `review`.
- 2026-11-10 lead-architect (wave F run, session 7): code reviewed (D5, round 1) against aa0d4cb. Scope: only the two Allowed Files changed, and `cards_path_armorer_test.go` is the same as in f8ec408. The definition is the task's Definition, field for field. Fences: no card name in engine code, no new shape, no `OPEN(…)`. Reuse: `HookMoves`, `Filter.ZoneRef` with `ZoneRefThat`, an exactly-one target and `OpAddArmor`, all existing. Gates re-run in the review worktree, all exit 0: `go test ./pkg/engine -run TestCard_BrokenHornArmorer`, `go tool task check`, `go tool task sim` (200 games, Path 42/50 main cards executable, P0 68.0% / P1 30.5% / draws 1.5%, the same as the lead's probe). The Flash run (`agy` conversation `e01c5df5-aa7e-41f3-948e-60d1cd880fa8`) left no denied action. Approved.
