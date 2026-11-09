---
id: T-0200
title: Hand-active cost reduction, first card Boneyard Amalgam (wave E shape, sonnet)
kind: shape
tier: sonnet
wave: E
status: tests
branch: m2/autoloop-wave-e/T-0200-boneyard-amalgam
depends_on: [T-0182]
claims: [shape:CostReduction (ability.CostReduction, Def.CostReduction), card:Boneyard Amalgam]
cards: [Boneyard Amalgam]
requirements: [FR-CST-001]
---

## Goal
*Boneyard Amalgam* (a Path deck card, 2 copies): *"This unit costs R1 less to deploy for each unit card in your discard pile."* It needs the first cost-changing shape: a card's own hand-active cost reduction, read on demand by `EffectiveCost` (`FR-CST-001`). Directive #118 (b) fires with it: the view shows the lower cost.

Covers:
- The shape in `ability` and its read in `EffectiveCost`, as under Shape below.
- The definition in `pkg/engine/cards/path.go`, and the name off `cards.NotYetExecutable`.
- #118 (b): a view test that a hand card shows its lower cost.

## Inputs
- Card (`data/cards/units-and-structures-path.csv` line 33): Path Unit, Skeleton, Giant, R10 T0, deploys to Reserves, power 8, health 6, yields R3 T3.
- Survey row: `docs/design/m2_starter_set_ability_survey.md` line 190 ("Continuous (hand-active)", "costs R1 less per unit card in your discard", `QUERY` count).
- FRD 01: `FR-CST-001` (lines 300–306): Effective Cost = max(0, printed + increases − reductions), clamped at 0; a hand cost reduction (*Boneyard Amalgam* is named) evaluates on demand on the `CardInstance` in its current zone (`activeZone: HAND`), during cost calculation and deployment legality checks. Also line 469 (the non-battlefield boundary).
- Rulings: none.
- Design note: `docs/architecture/m2-ability-model.md`, Key types (`Static`, `Value`) and Boundaries.
- Directive #118 (b) on `BOARD.md`.

### Shape (the engineer builds it)
- **`ability.CostReduction`**, a new struct: `Req int` (Requisition off for each counted card) and `Per Filter` (the cards counted: a pile filter, `Per.InPile()`, its Allegiance relative to the card's owner). `Def` gains `CostReduction CostReduction`, zero for a card without one. No Tactics field: no card in `data/cards` lowers a Tactics cost. The design note's planned `Static` type isn't built: this is the card's own cost, not an aura, and the lead updates the note's Key types line when this merges.
- **`Game.EffectiveCost(inst)`** ([costs.go:14](../../../pkg/engine/costs.go)) stays the one read of a card's cost. While `inst` is in its owner's hand and its `Def.CostReduction.Req` is above 0, it subtracts `Req` times the number of cards in discard piles that match `Per` (each checked with `matchesCard`, controller `inst.Owner`, source `inst`), and clamps Requisition at 0. Anywhere else, or with no reduction, it returns the printed cost as today. Put the count in a small helper in `costs.go` (a later "the number of matching cards" `Value` reuses it). Its callers (`handPlay`'s legality check, `payCost`, `sim`'s `ProgressPolicy`, `view`'s `ReqNow`) need no change, because they already read it.
- `payCost` runs before `takeCard` ([actions.go:200-204](../../../pkg/engine/actions.go)), so the card is still in hand when its cost is paid.

### Definition (the engineer adds it to `pathDefs`, before *Boneyard Scavenger*)
- `Text`: `"This unit costs R1 less to deploy for each unit card in your discard pile."`
- `CostReduction: ability.CostReduction{Req: 1, Per: ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit, Zone: ability.ZoneDiscard}}`.
- Remove the name from `cards.NotYetExecutable`.

A lead probe of exactly this shape and definition on the run branch at e74170a, since reverted, passes `go tool task check`, and `go tool task sim` runs clean: Path 38/50 main cards executable, P0 73.5% / P1 26.0% / draws 0.5% over 200 games. So `sim-compare` will differ here and isn't a gate.

## Acceptance Cases (#118 (b); the engineer writes this test in `pkg/view/view_test.go` and commits it as `test(T-0200): …` before the implementation, like a harness task's)
- `TestView_HandCardShowsItsLowerCost`: a game with the real card set loaded (as at `view_test.go:459`). Seat 0 has *Boneyard Amalgam* in hand and 3 unit cards in its own discard pile. In seat 0's `PlayerView`, that hand card has `ReqCost` 10, `TacCost` 0, `ReqNow` 7 and `TacNow` 0. It fails before the implementation (`ReqNow` is 10). Leave `TestView_HandCardCarriesItsCostNow` as it is.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
In a new file `pkg/engine/cards_path_amalgam_test.go`, from the CSV text and `FR-CST-001`, loading the real card (`s.card`) and putting cards in discard piles with `s.inDiscard` and in hand with `s.give`. Read the cost with `g.EffectiveCost` and check deploy legality with `LegalActions` or `Apply`.
- `TestCard_BoneyardAmalgam_PrintedCostWithNoUnitCards` (`FR-CST-001`): in hand, with only a Maneuver card and a Structure card in its owner's discard pile, it costs R10 T0.
- `TestCard_BoneyardAmalgam_OneLessPerUnitCard` (`FR-CST-001`): 3 unit cards (plus a Maneuver and a Structure) in its owner's discard pile: R7 T0. A fourth unit card put there afterwards makes it R6, read on demand.
- `TestCard_BoneyardAmalgam_OpponentsPileNotCounted` (`FR-CST-001`): 3 unit cards in the opponent's discard pile and none in its owner's: R10.
- `TestCard_BoneyardAmalgam_CountsAnotherAmalgamInThePile` (`FR-CST-001`): a second *Boneyard Amalgam* in its owner's discard pile is a unit card: the one in hand costs R9.
- `TestCard_BoneyardAmalgam_CostClampsAtZero` (`FR-CST-001`): 12 unit cards in its owner's discard pile: R0 T0, never negative. With an empty pool, a Deploy of it is legal, and applying it leaves the pool at R0 T0.
- `TestCard_BoneyardAmalgam_DeployPaysTheLowerCost` (`FR-CST-001`): 3 unit cards in its owner's discard pile and a pool of R7: the Deploy is legal and leaves R0. With a pool of R6 instead, the Deploy is not in `LegalActions` and `Apply` rejects it.
- `TestFR_CST_001_ReductionOnlyFromHand` (`FR-CST-001`, `activeZone: HAND`): a *Boneyard Amalgam* in its owner's discard pile, with 3 other unit cards there, reads its printed R10 from `EffectiveCost`.
- **Existing tests:** every other test passes unchanged, including `TestStarterSet_Executable` with the name removed, `TestDefs_MatchRulesText`, `TestView_Case8_ExecutableFlag`, `TestView_HandCardCarriesItsCostNow`, `TestProgress_Case4_Coverage` and `TestEvents_Case4_EveryActionKind`.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected (`card`, `give`, `inDiscard` exist).

## Allowed Files
- `pkg/engine/ability/ability.go` (`CostReduction`, `Def.CostReduction`)
- `pkg/engine/costs.go` (`EffectiveCost` and its count helper)
- `pkg/engine/cards/path.go` (the definition)
- `pkg/engine/cards/cards.go` (remove the name from `NotYetExecutable`)
- Test: a new test function in `pkg/view/view_test.go` (the acceptance case above, in the `test(T-0200)` commit)

## Done When
- [ ] The card and rules tests pass: `go test ./pkg/engine -run "TestCard_BoneyardAmalgam|TestFR_CST_001"`
- [ ] The view test passes: `go test ./pkg/view -run TestView_HandCardShowsItsLowerCost`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] The card loads with no unsupported lines (`TestStarterSet_Executable` passes with the name removed)

## Log
- 2026-11-09 lead-architect (wave E run, session 2): written against e74170a, with the lead's probe above. Board read twice (D2.0), no `open: before` directive. Waiting triggers checked: #118 (b) fires (this task's `cards:` hold *Boneyard Amalgam*), set to `open` and applied here; #15, #71, #91 (2), #106, #120 and #128 have not fired (no turn hook, each-player effect, mode cost, criteria search, mandatory discard cost or discard deploy among this task's cards). Searched `pkg/engine` and the task archive for an existing cost change: none (`EffectiveCost` returns the printed cost, `costs.go:11-16`), and `matchesCard` already checks pile cards against a filter, so nothing is duplicated. Set to `tests`.
