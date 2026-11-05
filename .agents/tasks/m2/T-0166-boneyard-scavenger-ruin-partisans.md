---
id: T-0166
title: Make Boneyard Scavenger and Ruin Partisans executable (Tactics, exhaust, kill this unit, return a chosen card of one type from your discard pile to your hand, wave D card, flash)
kind: card
tier: flash
wave: D
status: tests
branch: m2/autoloop-wave-d/T-0166-boneyard-scavenger-ruin-partisans
depends_on: [T-0158, T-0162, T-0163, T-0165]
claims: [card:Boneyard Scavenger, card:Ruin Partisans]
cards: [Boneyard Scavenger, Ruin Partisans]
requirements: [FR-ACT-001, FR-ACT-002, FR-TGT-001]
---

## Goal
*Boneyard Scavenger* (a deck card, 2 copies): *"T1, exhaust, kill this unit: Return chosen maneuver card from your discard pile to your hand."* *Ruin Partisans* (not in a deck; #14 (c) and the run file's `Kept` line let it join, because it needs nothing new): *"T2, exhaust, kill this unit: Return chosen unit card from your discard pile to your hand."* Both only combine T-0158's activated ability and costs, T-0162's `KindManeuver`, the existing `KindUnit`, and T-0081's discard-pile target with `OpReturnToHand`.

Covers:
- The two definitions in `pkg/engine/cards/path.go`, as under Definition below.
- The two names off `cards.NotYetExecutable`.

## Inputs
- Cards: `data/cards/units-and-structures-path.csv` line 23 (*Boneyard Scavenger*: Path Unit, Specialist, Reaver, deploys to Reserves) and line 29 (*Ruin Partisans*: Path Unit, Infantry, Reaver, deploys to Reserves).
- Survey rows: `docs/design/m2_starter_set_ability_survey.md` lines 184 and 185.
- FRD 01 v3.21: `FR-ACT-001` (line 341, which names *Ruin Partisans*), `SVY-R36` (line 342), `FR-ACT-002` (line 343), `FR-TGT-001`, `SVY-R2`.
- Rulings: `SVY-R2`, `SVY-R36`.

### Definition (the engineer adds both to `pathDefs`, in the map's alphabetical order)
*Boneyard Scavenger*:
- `Text`: the printed line, as above.
- `Activated`: one ability.
  - `Costs`, in printed order: `{Kind: ability.CostResources, Tac: 1}`, `{Kind: ability.CostExhaustSelf}`, `{Kind: ability.CostKillSelf}`.
  - `Effect.Targets`: one exactly-one slot, `Filter: ability.Filter{Allegiance: ability.AllegianceAllied, Zone: ability.ZoneDiscard, Kinds: ability.KindManeuver}`.
  - `Effect.Steps`: one, `{Op: ability.OpReturnToHand, Subject: ability.SubjectTarget, Target: 0}`.

*Ruin Partisans*: the same with its own `Text`, `Tac: 2` and `Kinds: ability.KindUnit`.

Remove both names from `cards.NotYetExecutable`.

A lead probe with these definitions, since reverted, passes `go tool task check`, and the starter sim stays clean: 34/50 Path main cards executable, P0 83.0% / P1 15.5% / draws 1.5% over 200 games, no monitor violation. So `sim-compare` will differ here and isn't a gate.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
In a new file `pkg/engine/cards_path_reclaim_test.go`, from the CSV text, loading the real cards (`s.card`). Player 0 activates. `pkg/engine/cards_animus_sinuous_test.go` is the model: give player 0 Tactics with `setPool`, use `activate`, and "not offered" means `actNotOffered`, which checks both `offersActivate` and `s.reject`. A pile with one legal card may not ask a question, so a test that reads the options holds at least two.

*Boneyard Scavenger* (Maneuvers: for example *Steel Rain*, *Fortify*; unit: *Saboteur*; structure: *Hammerhand Bivouac*):
- `TestCard_BoneyardScavenger_ReturnsTheChosenManeuver`: two Maneuvers in player 0's pile, player 0 choosing the older one. The chosen card is in hand, the other stays in the pile, *Boneyard Scavenger* is in its owner's discard pile, and the pool lost exactly 1 Tactics and no Requisition.
- `TestCard_BoneyardScavenger_ListsOnlyManeuverCards` (`FR-TGT-001`): the pile holds two Maneuvers, a unit and a structure. The target question lists exactly the two Maneuvers.
- `TestCard_BoneyardScavenger_NeedsAManeuverInThePile` (`SVY-R36`): with an empty pile, or a pile holding only a unit and a structure, it is not offered. Nothing is paid and it stays on the battlefield, Readied.
- `TestCard_BoneyardScavenger_OpponentsPileDoesNotCount` (`FR-TGT-001`): the only Maneuver is in player 1's discard pile. Not offered.
- `TestCard_BoneyardScavenger_NeedsOneTactics` (`FR-ACT-001`): a Maneuver in the pile, 0 Tactics and plenty of Requisition. Not offered, and nothing changes.
- `TestCard_BoneyardScavenger_NeedsToBeReadied` (`FR-ACT-002`): a Maneuver in the pile and 1 Tactics, but the unit is Exhausted. Not offered, because its cost includes exhausting it.
- `TestCard_BoneyardScavenger_WorksFromReserves` (`FR-ACT-001`): a Readied one in Reserves, where it deploys, can be activated and returns the card.

*Ruin Partisans* (units: for example *Saboteur*, *Devilhorn Captain*; Maneuver: *Steel Rain*; structure: *Hammerhand Bivouac*):
- `TestCard_RuinPartisans_ReturnsTheChosenUnit`: two units in player 0's pile, player 0 choosing the older one. The chosen card is in hand, the other stays in the pile, *Ruin Partisans* is in its owner's discard pile, and the pool lost exactly 2 Tactics and no Requisition.
- `TestCard_RuinPartisans_ListsOnlyUnitCards` (`FR-TGT-001`): the pile holds two units, a Maneuver and a structure. The target question lists exactly the two units.
- `TestCard_RuinPartisans_OwnCardCannotBeChosen` (`SVY-R2`, `SVY-R36`): with an empty pile, or a pile holding only a Maneuver and a structure, it is not offered although its own card is a unit. Nothing is paid and it stays on the battlefield, Readied.
- `TestCard_RuinPartisans_AnotherCopyCanReturnTheFirst` (`SVY-R2`): two copies on the battlefield and one other unit card in the pile. After the first copy resolves, the second is offered, and its question lists the first copy's card.
- `TestCard_RuinPartisans_NeedsTwoTactics` (`FR-ACT-001`): a unit in the pile, 1 Tactics and plenty of Requisition. Not offered, and nothing changes.
- `TestCard_RuinPartisans_NeedsToBeReadied` (`FR-ACT-002`): a unit in the pile and 2 Tactics, but it is Exhausted. Not offered.
- **Existing tests:** every other test passes unchanged, including `TestStarterSet_Executable` with the names removed, `TestProgress_Case4_Coverage` and `TestEvents_Case4_EveryActionKind`.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected.

## Allowed Files
- `pkg/engine/cards/path.go` (the definitions)
- `pkg/engine/cards/cards.go` (remove the names from `NotYetExecutable`)

## Done When
- [ ] The card tests pass: `go test ./pkg/engine -run TestCard_BoneyardScavenger`
- [ ] The card tests pass: `go test ./pkg/engine -run TestCard_RuinPartisans`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] The cards load with no unsupported lines (`TestStarterSet_Executable` passes with the names removed)

## Log
- 2026-11-04 lead-architect (wave D run, session 11): written against 69de466, with the lead's probe above. #105's part (1) is applied, and parts (2), (3) go with *Cliff of Portents*, which also needs `SEARCH_DECK`, so this card task goes first. Searched `pkg/engine` and the task archive: `KindManeuver`, `KindUnit`, the discard-pile target and the three cost kinds all exist, so nothing here is a new shape. Board re-read, no new directive. Set to `tests`.
