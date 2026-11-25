---
id: T-0305
title: Make Swirling Dustdevil executable (a hand-active cost reduction, wave A, card, flash)
kind: card
tier: flash
wave: A
status: in-progress
branch: m3/autoloop-wave-a/T-0305-swirling-dustdevil
depends_on: [T-0304]
claims: [card:Swirling Dustdevil]
cards: [Swirling Dustdevil]
requirements: [FR-CST-001, FR-TRN-003, FR-ZON-005]
---

## Goal
*Swirling Dustdevil* (a Concordia unit, R8 T0, 6/5, Reserves; 4 copies in the starter deck): *"This unit costs R1 less to deploy for each maneuver card in your discard pile."* It only combines shapes that are `done`: `Def.CostReduction`, which *Boneyard Amalgam* (T-0200) pins for unit cards, with `KindManeuver` in its pile filter (*Boneyard Scavenger*'s target filter). It is the second card to use `CostReduction` (survey §3, "Pinned by one card in M2").

Covers:
- The definition in `pkg/engine/cards/concordia.go`, as under Definition below.
- The name off `cards.NotYetExecutable`.

## Inputs
- Card (`data/cards/units-and-structures-concordia.csv` line 31).
- Survey row: `docs/design/survey/m3_full_catalog_ability_survey.md` §6, "Concordia: deck" (*Swirling Dustdevil*), and the "Pinned by one card in M2" note (line 81).
- FRD 01: `FR-CST-001` (lines 348-354), `FR-TRN-003` and `FR-MNV-001` (line 394), `FR-ZON-004` (line 279), `FR-ZON-005` (lines 287-290), `FR-SET-004` Immediate Readiness (line 203).
- Rulings: none. Only `SVY-R59` names the cost of a battlefield face, and it doesn't reach a card in hand.
- Design note: `docs/architecture/m3-full-catalog.md`, Boundaries (`pkg/engine/cards`), Deliverable 5.

### Definition (the engineer adds it to `concordiaDefs`)
- `"Swirling Dustdevil"`, `Text`: `"This unit costs R1 less to deploy for each maneuver card in your discard pile."`.
- `CostReduction: ability.CostReduction{Req: 1, Per: ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindManeuver, Zone: ability.ZoneDiscard}}`. Nothing else: no triggers, effect, statics or activated abilities.

A lead probe with the definition on the run branch at ca30f9ec, since reverted, passed `go tool task check` and `go tool task sim-matrix`.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
In a new file `pkg/engine/cards_concordia_dustdevil_test.go`, from the CSV text, loading the real card (`s.card`), in the style of `cards_path_amalgam_test.go`. Player 0 holds the Dustdevil and is the active player. Reuse the existing helpers, `amalgamCost`, `amalgamDeployLegal` and `amalgamDeploy` (none of them is specific to the Amalgam) included, rather than copying them. The other cards may be synthetic.

- `TestCard_SwirlingDustdevil_PrintedCostWithNoManeuverCards` (`FR-CST-001`): R8 T0 with an empty discard pile, and with only unit and structure cards in it (a second *Swirling Dustdevil* included).
- `TestCard_SwirlingDustdevil_OneLessPerManeuverCard` (`FR-CST-001`): with 3 Maneuver cards plus a unit and a structure card in the pile it costs R5, and a fourth put there afterwards makes it R4. Maneuvers of any faction count, and copies of one name count each.
- `TestCard_SwirlingDustdevil_OpponentsPileNotCounted` (`FR-CST-001`): Maneuver cards only in player 1's pile leave R8. With 2 in each pile it costs R6.
- `TestCard_SwirlingDustdevil_ManeuverCardsElsewhereNotCounted` (`FR-CST-001`): Maneuver cards in player 0's hand and deck, and a face-down Scratch Company whose card is a Maneuver, leave R8.
- `TestCard_SwirlingDustdevil_KilledScratchCompanyManeuverCounts` (`FR-ZON-005`): a Scratch Company of player 0 whose card is a Maneuver is destroyed: its card is in the pile with its printed characteristics, and the cost is R7.
- `TestCard_SwirlingDustdevil_EnactedManeuverCounts` (`FR-MNV-001`): player 0 enacts a real Maneuver (*Growing Legion*), and once it has resolved the cost is R7.
- `TestCard_SwirlingDustdevil_CostRisesWhenAManeuverLeavesThePile` (`FR-CST-001`): a real effect takes a Maneuver card out of the pile (*Boneyard Scavenger*, which returns one to hand), and the cost rises by R1.
- `TestCard_SwirlingDustdevil_CostClampsAtZero` (`FR-CST-001`): with 8 Maneuver cards, and with 10, it costs R0 T0. It may be deployed from an empty pool, which stays R0 T0.
- `TestCard_SwirlingDustdevil_DeployPaysTheLowerCost` (`FR-TRN-003`): with 3 Maneuver cards: with R5 the Deploy is offered and leaves R0. With R8 it leaves R3. With R4 it isn't offered, a Deploy is rejected, and the card and the pool stay as they were. Tactics can't stand in for the missing Requisition (R4 T4).
- `TestCard_SwirlingDustdevil_EntersReservesAsSixFive` (`FR-TRN-003`, `FR-SET-004`): deployed at a reduced cost it is a Readied 6/5 unit in Reserves with yields 2/2, and the pile still holds its Maneuver cards. Deploying it to the Main Theater is rejected.
- `TestCard_SwirlingDustdevil_ReducesOnlyItself` (`FR-CST-001`): with 3 Maneuver cards in the pile, a second Dustdevil in hand costs R5 by its own text, and another unit card in hand and a Maneuver card in hand cost what they print, with a Dustdevil in hand and with one on the battlefield.
- `TestCard_SwirlingDustdevil_PrintedCostInTheDiscardPile` (`FR-CST-001`): in the discard pile with 3 Maneuver cards it reads R8 (the reduction is hand-active).
- **Existing tests:** every other test passes unchanged, including `TestCards_Executable` with the name removed.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected.

## Allowed Files
- `pkg/engine/cards/concordia.go` (the definition)
- `pkg/engine/cards/cards.go` (remove the name from `NotYetExecutable`)

## Done When
- [ ] The card tests pass: `go test ./pkg/engine -run TestCard_SwirlingDustdevil`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-matrix` exits 0 (just over two minutes: use a timeout above the two-minute default)
- [ ] The card loads with no unsupported lines (`TestCards_Executable` passes with the name removed)

## Log
- 2026-11-25 lead-architect (wave A run, session 7): written against ca30f9ec, with the lead's probe above. Board read twice (D2.0): no directive is due here (#185 (3) and (4) are *Will Breaker*'s and *Tempest City Ray*'s), and no waiting trigger has fired. Every part of the Definition exists. Set to `tests`.
- 2026-11-25 rules-test-author: tests written in `pkg/engine/cards_concordia_dustdevil_test.go` (12 tests, all `TestCard_SwirlingDustdevil_*`: PrintedCostWithNoManeuverCards, OneLessPerManeuverCard, OpponentsPileNotCounted, ManeuverCardsElsewhereNotCounted, KilledScratchCompanyManeuverCounts, EnactedManeuverCounts, CostRisesWhenAManeuverLeavesThePile, CostClampsAtZero, DeployPaysTheLowerCost, EntersReservesAsSixFive, ReducesOnlyItself, PrintedCostInTheDiscardPile). They reuse `amalgamCost`, `amalgamDeployLegal`, `amalgamDeploy`. Helpers requested: none. Nothing left out. `go vet` is clean; the tests fail now with "needs rules the kernel does not implement yet" or the printed cost, as expected. The destroyed-Scratch test has player 1 attack the face-down unit on its turn (the spec did not name the killer). `cardtown lint rules` ran. Set to `tests-review`.
- 2026-11-25 lead-architect: tests approved (D4, round 1). Each expected outcome agrees with the FRD 01 text it cites. A lead probe with the Definition, since reverted, passes all 12, and 13 of 15 mutants fail at least one (amount 0 and 2, kinds, no kind, opposing, any allegiance, no zone, a stat criterion, Scratch only, a subtype). The two survivors add a deploy trigger the card does not print (scratch, draw), which the code review checks in the diff. `go vet` and `gofmt` are clean. Set to `ready`.
