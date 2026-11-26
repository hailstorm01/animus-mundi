---
id: T-0308
title: Make Echoing Nightmare executable (a deploy trigger returns a chosen opposing Main Theater unit, and if it does the Nightmare moves to the Main Theater, wave A, card, flash)
kind: card
tier: flash
wave: A
status: tests-review
branch: m3/autoloop-wave-a/T-0308-echoing-nightmare
depends_on: [T-0307]
claims: [card:Echoing Nightmare]
cards: [Echoing Nightmare]
requirements: [FR-TRN-003, FR-TGT-001, FR-TGT-003, FR-ZON-002, FR-ZON-003, FR-ZON-004, FR-SET-004, FR-KWD-002h, FR-ZON-005]
---

## Goal
*Echoing Nightmare* (a Concordia unit, Spirit, Dread, R5 T0, 6/2, Reserves, yields R2 T2; in the starter deck): *"When this unit deploys, return chosen opposing unit in the Main Theater to its owner's hand. If a unit is returned this way, move this unit to the Main Theater."* It only combines shapes that are `done`: a `HookDeploy` trigger with one target slot (*Stormbolt Ray*, T-0304), `OpReturnToHand` on that target (*Icewind Regent*, T-0300), and an `IfDone` step (`SVY-R14`, as *Infested Apostle* in `pkg/engine/cards/animus.go`) that moves `SubjectSelf` with `OpMoveZone` (as *Tempest City Ray*, T-0306, moves its target). "If a unit is returned this way" is read as `IfDone`: the move applies only if the return did something (`SVY-R14`, `FR-TGT-003`'s Complete Target Invalidation). The move is a move (`SVY-R1`), so it fires the Nightmare's own `HookMoves` triggers; the deploy does not.

Covers:
- The definition in `pkg/engine/cards/concordia.go`, as under Definition below.
- The name off `cards.NotYetExecutable`.

## Inputs
- Card (`data/cards/units-and-structures-concordia.csv` line 23).
- Survey row: `docs/design/survey/m3_full_catalog_ability_survey.md` §6, "Concordia: deck" (*Echoing Nightmare*, line 133).
- FRD 01: `FR-TRN-003` (line 394), `FR-TGT-001` (line 407, and `SVY-R14` at line 423), `FR-TGT-003` (lines 435-442, including Deployment Triggers Without Legal Targets at line 442), `FR-ZON-002` (lines 271-275: what counts as moving, `SVY-R1`, and "returned to a non-battlefield zone, an asset always goes to its owner's ... Hand"), `FR-ZON-003` (line 250), `FR-ZON-004` (line 279), `FR-SET-004` Immediate Readiness (line 203), `FR-KWD-002h` Ward, `FR-ZON-005` Double-Blind Deployment (line 288: a face-down Scratch Company's only printed ability is Deploys Exhausted).
- Rulings that apply: `SVY-R14`, `SVY-R1`, `SVY-R25`.
- Design note: `docs/architecture/m3-full-catalog.md`, Boundaries (`pkg/engine/cards`), Deliverable 5.

### Definition (the engineer adds it to `concordiaDefs`)
- `"Echoing Nightmare"`, `Text`: `"When this unit deploys, return chosen opposing unit in the Main Theater to its owner's hand. If a unit is returned this way, move this unit to the Main Theater."` (one string).
- One `Trigger`: `Hook: ability.HookDeploy`. `Effect`:
  - `Targets`: one slot, `Filter: ability.Filter{Allegiance: ability.AllegianceOpposing, Kinds: ability.KindUnit, Zone: ability.ZoneMainTheater}`.
  - `Steps`: `{Op: ability.OpReturnToHand, Subject: ability.SubjectTarget, Target: 0}`, then `{Op: ability.OpMoveZone, To: ability.ZoneMainTheater, Subject: ability.SubjectSelf, IfDone: true}`.

A lead probe with the definition on the run branch at 146dc507, since reverted, passed `go tool task check`, `go tool task sim` (P0 58.5% / P1 39.5% / draws 2.0%, unchanged) and `go tool task sim-matrix`. Deployed from hand with one opposing unit in the Main Theater, nothing was asked (the Single-Option Fast-Path), the unit went to player 1's hand, and the Nightmare moved to the Main Theater, Readied, with power 6. With only an opposing unit in Reserves, no trigger was placed and the Nightmare stayed in Reserves.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
In a new file `pkg/engine/cards_concordia_nightmare_test.go`, from the CSV text, loading the real card (`s.card`), in the style of `cards_concordia_stormbolt_test.go` and `cards_concordia_tempest_ray_test.go`. Player 0 deploys the Nightmare from hand to Reserves (`s.deployTo`, pool R5) and is the active player, unless the case says otherwise. Reuse the existing test helpers (`s.deployTo`, `s.setPool`, `s.put`, `s.give`, `s.onTop`, `s.script`, `s.asked`, `s.attack`, `s.endTurn`, `s.alive`, `s.ptPlaceHook`, `inHand`, `inAssets`, `vrHQ`, `glEnact`, `effectEvents`, `unit`). The other cards may be synthetic.

- `TestCard_EchoingNightmare_EntersReservesAsSixTwo` (`FR-TRN-003`, `FR-SET-004`): with no opposing unit on the battlefield, deployed from hand for R5 it is a Readied 6/2 unit in player 0's Reserves, and nothing is placed or asked; deploying it to the Main Theater is rejected, and with R4 it can't be deployed.
- `TestCard_EchoingNightmare_ReturnsTheUnitAndMoves` (`FR-ZON-002`, `SVY-R14`): with one opposing unit in the Main Theater (damaged, so the reset shows), the unit's card is in player 1's hand and off the battlefield, and the Nightmare is in player 0's Main Theater, still Readied. One leave event names the unit and one move event names the Nightmare, the leave first.
- `TestCard_EchoingNightmare_MovedNightmareMayAttack` (`FR-SET-004`): after the trigger resolves, the Nightmare attacks player 1's Headquarters, which takes 6 combat damage.
- `TestCard_EchoingNightmare_OnlyOpposingMainTheaterUnitsOffered` (`FR-TGT-001`): with two opposing units in the Main Theater, one opposing unit in player 1's Reserves, an opposing structure in the Main Theater and an allied unit in the Main Theater, the target decision offers exactly the two opposing Main Theater units. Choosing the second returns only it.
- `TestCard_EchoingNightmare_NoTargetStaysInReserves` (`FR-TGT-003` Deployment Triggers Without Legal Targets): with opposing units only in player 1's Reserves (and an allied unit in the Main Theater), the Nightmare enters Reserves and stays there, and nothing returns.
- `TestCard_EchoingNightmare_WardedUnitNotOffered` (`FR-KWD-002h`): with only an opposing Warded unit in the Main Theater, it stays on the battlefield and the Nightmare stays in Reserves. With a Warded and a plain one, only the plain one is offered.
- `TestCard_EchoingNightmare_TargetGoneNoMove` (`FR-TGT-003`, `SVY-R14`): the chosen unit is killed between the trigger's placement and its resolution (`s.ptPlaceHook`). Nothing returns, and the Nightmare stays in Reserves.
- `TestCard_EchoingNightmare_TargetLeftTheMainTheaterNoMove` (`FR-TGT-003`): the chosen unit is moved to player 1's Reserves between placement and resolution. It stays in player 1's Reserves, and the Nightmare stays in player 0's Reserves.
- `TestCard_EchoingNightmare_SourceGoneStillReturns` (`FR-ZON-004`): the Nightmare is killed between its trigger's placement and resolution (as in `TestCard_StormboltRay_SourceGoneStillDeals`). The chosen unit still returns to player 1's hand, nothing moves, and the Nightmare's card is in player 0's discard pile.
- `TestCard_EchoingNightmare_ControlledUnitGoesToItsOwnersHand` (`FR-ZON-002`, `FR-ZON-003`): a unit player 0 owns, which player 1 controls in the Main Theater through the real *Rhapsody's Hollow* (as in `TestCard_TempestCityRay_ControlledUnitGoesToItsControllersReserves`), is a legal target. Chosen, it goes to player 0's hand, not player 1's, and the Nightmare moves.
- `TestCard_EchoingNightmare_MoveFiresMovesHooks` (`SVY-R1`, `FR-ZON-002`): a copy of the real card (the shared definition is not changed, as `trayTapCard` does) with one more synthetic `HookMoves` trigger on itself (+1 Armor on `SubjectSelf`). With a target, it ends with 1 Armor (the move fired it once, the deploy didn't). With no target, it has 0 Armor.
- `TestCard_EchoingNightmare_FaceDownScratchDoesNothing` (`FR-ZON-005`): with a Nightmare on top of player 0's deck (`s.onTop`) and an opposing unit in the Main Theater, player 0 enacts the real *Growing Legion*. A face-down Scratch Company enters player 0's Reserves, the opposing unit stays, and nothing else moves.
- **Existing tests:** every other test passes unchanged, including `TestCards_Executable` with the name removed.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected.

## Allowed Files
- `pkg/engine/cards/concordia.go` (the definition)
- `pkg/engine/cards/cards.go` (remove the name from `NotYetExecutable`)

## Done When
- [ ] The card tests pass: `go test ./pkg/engine -run TestCard_EchoingNightmare`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-matrix` exits 0 (just over two minutes: use a timeout above the two-minute default)
- [ ] The card loads with no unsupported lines (`TestCards_Executable` passes with the name removed)

## Log
- 2026-11-25 lead-architect (wave A run, session 10): written against 146dc507, with the lead's probe above. `main` has no commits the run branch lacks. Board read twice (D2.0): no directive is due here (#185 stays for (3) *Will Breaker* and (4)'s *Awaken Nightmare* half, wave F), and no waiting trigger has fired. Every part of the Definition exists, so no kernel change. Set to `tests`.
- 2026-11-25 rules-test-author: wrote `pkg/engine/cards_concordia_nightmare_test.go`, 11 tests as listed under Tests: EntersReservesAsSixTwo, ReturnsTheUnitAndMoves, MovedNightmareMayAttack, OnlyOpposingMainTheaterUnitsOffered, NoTargetStaysInReserves, WardedUnitNotOffered, TargetGoneNoMove, TargetLeftTheMainTheaterNoMove, SourceGoneStillReturns, ControlledUnitGoesToItsOwnersHand, MoveFiresMovesHooks, FaceDownScratchDoesNothing (12 with the last). `go vet` and `cardtown lint rules` pass; 11 fail until the definition is added, and FaceDownScratchDoesNothing passes already (a regression guard). Helpers requested: none (reuses stormChoose, stormKill from the Stormbolt test file). Cases left out: none. Set to `tests-review`.
