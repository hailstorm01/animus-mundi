---
id: T-0309
title: Make Will Breaker executable (a deploy trigger kills a chosen opposing unit with Requisition cost 3 or less, wave A, card, flash)
kind: card
tier: flash
wave: A
status: tests
branch: m3/autoloop-wave-a/T-0309-will-breaker
depends_on: [T-0308]
claims: [card:Will Breaker]
cards: [Will Breaker]
requirements: [FR-TRN-003, FR-TGT-001, FR-TGT-003, FR-LAY-002, FR-ZON-002, FR-ZON-005, FR-ENT-103, FR-KWD-002h, FR-SET-004, FR-ZON-004]
---

## Goal
*Will Breaker* (a Concordia unit, Spirit, Ego, R4 T0, 5/4, Reserves, yields R1 T2; in the starter deck, 2 copies): *"When this unit deploys, kill chosen opposing unit with Requisition cost 3 or less."* It only combines shapes that are `done`: a `HookDeploy` trigger with one target slot (*Stormbolt Ray*, T-0304), a `Filter.Stat` bound (`StatRequisition`, `AtMost` 3, as *Devilhorn Scouting Party* in `pkg/engine/cards/path.go`), which reads a battlefield asset through `assetStat` (`pkg/engine/filter.go`: the cost printed on what the asset is now), and `OpKill` on that target (as in `pkg/engine/cards/animus.go`). A face-down Scratch Company, including a unit turned face down, costs 0 (`SVY-R59`), so it is a legal target. This task applies directive #185 (3).

Covers:
- The definition in `pkg/engine/cards/concordia.go`, as under Definition below.
- The name off `cards.NotYetExecutable`.
- Directive #185 (3): the three *Will Breaker* card tests it names (marked below).

## Inputs
- Card (`data/cards/units-and-structures-concordia.csv` line 22).
- Survey row: `docs/design/survey/m3_full_catalog_ability_survey.md` §6, "Concordia: deck" (*Will Breaker*, line 134), and `SVY-R59` in §4 (line 93).
- FRD 01: `SVY-R59` (line 20: a face-down Scratch Company costs 0 and 0, "so *Will Breaker* may choose it. This covers a unit turned face down"), `FR-TRN-003` (line 394), `FR-TGT-001` (line 407, Dynamic Stats at line 412), `FR-TGT-003` (lines 435-442, including Deployment Triggers Without Legal Targets at line 442), `FR-LAY-002` (line 534, destruction), `FR-ZON-002` (line 271: a destroyed asset goes to its owner's discard pile), `FR-ZON-005` (line 288), `FR-ENT-103` (line 298), `FR-SET-004` Immediate Readiness, `FR-KWD-002h` Ward, `FR-ZON-004` (line 279).
- Rulings that apply: `SVY-R59`, `SVY-R25`.
- Directive #185 (3) on `BOARD.md`.
- Design note: `docs/architecture/m3-full-catalog.md`, Boundaries (`pkg/engine/cards`), Deliverable 5.

### Definition (the engineer adds it to `concordiaDefs`)
- `"Will Breaker"`, `Text`: `"When this unit deploys, kill chosen opposing unit with Requisition cost 3 or less."` (one string).
- One `Trigger`: `Hook: ability.HookDeploy`. `Effect`:
  - `Targets`: one slot, `Filter: ability.Filter{Allegiance: ability.AllegianceOpposing, Kinds: ability.KindUnit, Stat: ability.StatRequisition, AtMost: 3}`.
  - `Steps`: `{Op: ability.OpKill, Subject: ability.SubjectTarget, Target: 0}`.

A lead probe with the definition on the run branch at 370c9464, since reverted, passed `go tool task check`, `go tool task sim` (P0 58.5% / P1 39.5% / draws 2.0%, unchanged) and `go tool task sim-matrix`.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
In a new file `pkg/engine/cards_concordia_willbreaker_test.go`, from the CSV text, loading the real card (`s.card`), in the style of `cards_concordia_stormbolt_test.go` and `cards_concordia_nightmare_test.go`. Player 0 deploys *Will Breaker* from hand to Reserves (`s.deployTo`, pool R4) and is the active player, unless the case says otherwise. Reuse the existing test helpers (`s.deployTo`, `s.setPool`, `s.put`, `s.putScratch`, `s.give`, `s.onTop`, `s.script`, `s.asked`, `s.alive`, `s.ptPlaceHook`, `stormChoose`, `stormKill`, `inHand`, `inAssets`, `effectEvents`, `unit`, `morphManeuver`/`morphEnact` from `morph_scratch_test.go`). The other cards may be synthetic; give each synthetic unit the printed Requisition cost the case needs.

- `TestCard_WillBreaker_EntersReservesAsFiveFour` (`FR-TRN-003`, `FR-SET-004`): with no opposing unit on the battlefield, deployed from hand for R4 it is a Readied 5/4 unit in player 0's Reserves, and nothing is placed or asked; deploying it to the Main Theater is rejected, and with R3 it can't be deployed.
- `TestCard_WillBreaker_KillsTheChosenUnit` (`FR-LAY-002`, `FR-ZON-002`): with one opposing unit of cost 3 in player 1's Main Theater, nothing is asked (the Single-Option Fast-Path), the unit is off the battlefield and its card is in player 1's discard pile. One destroy event names it.
- `TestCard_WillBreaker_OnlyCostThreeOrLessOffered` (`FR-TGT-001` Dynamic Stats): with opposing units of cost 2, 3 and 4 (one in each of player 1's zones where that fits), an allied unit of cost 1 and an opposing structure of cost 1, the target decision offers exactly the opposing units of cost 2 and 3. Choosing the cost 2 one kills only it.
- `TestCard_WillBreaker_MayChooseAScratchCompany` (`SVY-R59`, `FR-ZON-005`; #185 (3)): with an opposing face-down Scratch Company (`s.putScratch(1)`, whose hidden card has a printed cost over 3) and an opposing unit of cost 3, both are offered, and choosing the Scratch Company kills it.
- `TestCard_WillBreaker_MustChooseTheOnlyScratchCompany` (`SVY-R59`, `FR-TGT-003`; #185 (3)): with an opposing face-down Scratch Company the only opposing unit, nothing is asked and it is killed.
- `TestCard_WillBreaker_MayChooseAUnitTurnedFaceDown` (`SVY-R59`, `FR-ENT-103`; #185 (3)): an opposing unit with printed cost 5 is turned face down (`morphEnact` with `morphManeuver()`, or the real *Forced Mutation*) before *Will Breaker* deploys. Beside an opposing face-up unit of cost 5, it is the only target offered (or nothing is asked), and it is killed; the face-up one stays.
- `TestCard_WillBreaker_NoTargetDoesNothing` (`FR-TGT-003` Deployment Triggers Without Legal Targets): with only opposing units of cost 4 or more and an allied unit of cost 1, *Will Breaker* enters Reserves normally, nothing is asked, the stack is empty once play returns, no destroy event is logged, and every unit stays. Whether the trigger is placed (its targeted clause then fizzles) is not pinned: line 442 allows either ("when that trigger is placed on the stack or resolves").
- `TestCard_WillBreaker_WardedUnitNotOffered` (`FR-KWD-002h`): with only an opposing Warded unit of cost 2, it stays. With a Warded and a plain one of cost 2, only the plain one is offered.
- `TestCard_WillBreaker_TargetGoneDoesNothing` (`FR-TGT-003`): the chosen unit is removed between the trigger's placement and its resolution (`s.ptPlaceHook`, `stormKill`). Nothing else is killed.
- `TestCard_WillBreaker_TargetChangedControllerDoesNothing` (`FR-TGT-003`, `FR-ZON-002`): between placement and resolution, the chosen unit (player 1's) comes under player 0's control (set its `Controller` in the hook, or use the real *Rhapsody's Hollow* as in `TestCard_TempestCityRay_ControlledUnitGoesToItsControllersReserves`). It is no longer an opposing unit, so it stays on the battlefield.
- `TestCard_WillBreaker_SourceGoneStillKills` (`FR-ZON-004`): *Will Breaker* is removed between its trigger's placement and resolution (as in `TestCard_StormboltRay_SourceGoneStillDeals`). The chosen unit is still killed, and *Will Breaker*'s card is in player 0's discard pile.
- `TestCard_WillBreaker_FaceDownScratchDoesNothing` (`FR-ZON-005`): with a *Will Breaker* on top of player 0's deck (`s.onTop`) and an opposing unit of cost 1, player 0 enacts the real *Growing Legion*. A face-down Scratch Company enters player 0's Reserves, and the opposing unit stays.
- **Existing tests:** every other test passes unchanged, including `TestCards_Executable` with the name removed.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected.

## Allowed Files
- `pkg/engine/cards/concordia.go` (the definition)
- `pkg/engine/cards/cards.go` (remove the name from `NotYetExecutable`)

## Done When
- [ ] The card tests pass: `go test ./pkg/engine -run TestCard_WillBreaker`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-matrix` exits 0 (just over two minutes: use a timeout above the two-minute default)
- [ ] The card loads with no unsupported lines (`TestCards_Executable` passes with the name removed)

## Log
- 2026-11-25 lead-architect (wave A run, session 11): written against 370c9464, with the lead's probe above. `main` has no commits the run branch lacks. Board read twice (D2.0): #185 (3) is due here and is the three tests marked #185 (3); #185 (4)'s *Awaken Nightmare* half stays for wave F, and no waiting trigger has fired. Every part of the Definition exists, so no kernel change. Set to `tests`.
- 2026-11-25 rules-test-author: wrote pkg/engine/cards_concordia_willbreaker_test.go, 12 tests: EntersReservesAsFiveFour, KillsTheChosenUnit, OnlyCostThreeOrLessOffered, MayChooseAScratchCompany, MustChooseTheOnlyScratchCompany, MayChooseAUnitTurnedFaceDown (the three #185 (3) tests are the Scratch and face-down ones), NoTargetDoesNothing, WardedUnitNotOffered, TargetGoneDoesNothing, TargetChangedControllerDoesNothing, SourceGoneStillKills, FaceDownScratchDoesNothing. Helpers requested: none (test-local wbUnit, wbStructure, wbDeployCard, wbDeploy, wbDestroys are in the new file). Cases left out: none. go vet clean, cardtown lint rules passes. 11 fail until the definition exists; FaceDownScratchDoesNothing passes now and guards the rule. Set to tests-review.
- 2026-11-25 lead-architect: tests returned (D4, round 1). A lead probe with the Definition, since reverted, passes 11 of the 12, and each of 7 mutants (no stat bound, `AtMost` 2, any allegiance, any kind, Main Theater only, return in place of kill, a power bound) fails at least one of those 11. Note:
  1. `TestCard_WillBreaker_NoTargetDoesNothing` asserts that no trigger is placed (`placed` through `s.ptPlaceHook`). FRD 01 line 442 doesn't say that: it lets the trigger be placed with no legal target, and then only its targeted clause fizzles ("when that trigger is placed on the stack or resolves"). The kernel places it and it fizzles. The "no trigger is placed" wording was the lead's, in the task text, and is now corrected under Tests. Drop the placement assertion. Assert instead what the rule fixes: *Will Breaker* is in player 0's Reserves, nothing is asked, `s.g.Stack` is empty after the deploy, no destroy event is logged (`effectEvents(s.g, 0, "destroy")` is empty), and every unit stays. Change nothing else. Set to `tests`.
