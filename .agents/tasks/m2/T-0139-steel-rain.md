---
id: T-0139
title: Make Steel Rain executable (discard up to 3 cards as an additional cost, 2 damage to a chosen asset per card discarded, then draw 2 if your hand is empty, wave C card, flash)
kind: card
tier: flash
wave: C
status: ready
branch: m2/autoloop-wave-c/T-0139-steel-rain
depends_on: [T-0116, T-0131, T-0132, T-0136, T-0137, T-0138]
claims: [card:Steel Rain]
cards: [Steel Rain]
requirements: [FR-TRN-003a, FR-MNV-001, FR-TGT-001, FR-MOD-001]
---

## Goal
*Steel Rain* (a deck card, two copies in `data/decks/path-starter.txt`; 0 Requisition, 4 Tactics) is executable. The last part of directive #20 (b): it only combines shapes that are `done`, T-0131's discard cost, T-0138's per-discard target slot and T-0132's hand-empty gate.

## Inputs
- Card (`data/cards/maneuvers-path.csv` line 18): *Steel Rain*, Path Maneuver, cost 0 Requisition / 4 Tactics, three rules-text lines:
  - "As an additional cost to enact this maneuver, discard up to 3 cards."
  - "For each card discarded this way, deal 2 damage to chosen asset."
  - "Then, if you have no cards in hand, draw 2 cards."
- Survey row (`docs/design/m2_starter_set_ability_survey.md` line 180): discard up to 3; `DEAL_DAMAGE 2` chosen asset per discarded card; if hand empty, `DRAW_CARD 2`; "Targets chosen with the discard; repeats allowed; timeout concentrates on one asset (R32)". Rulings R3, R4, R4a (lines 62-63), R32 (line 91).
- FRD 01 (run branch, v3.20):
  - `FR-TRN-003a` (line 341, which names this card): the additional cost is paid on declaration.
  - `FR-MNV-001` (line 338): enacted from Hand, costs paid on declaration, the effect runs when it resolves, then the card goes to its owner's Discard pile.
  - `FR-TGT-001`, Targets Are Locked on Placement (`SVY-R3`, line 360): the targets are chosen with the discard, on declaration.
  - `FR-TGT-001`, Repeated Targets (`SVY-R4`, line 365): one independent choice per card discarded; the same asset may be chosen more than once, and each repetition is its own damage instance (`FR-REP-004`). "*Steel Rain* with 3 cards discarded may put all three 2-damage instances on one asset." On a timeout every repetition picks the first legal entity in fixed order (`SVY-R32`, `FR-MOD-001` §6: the choosing player's assets first, Headquarters first).
  - `FR-KWD-002h` (Ward): an opponent's Ward asset is never a legal target. "Chosen asset" has no allegiance or kind, so allied assets and both Headquarters are legal.
  - `SVY-R16` (`FR-MOD-001`): choosing from your own hand is not an inspection; an "up to" hand discard times out to none (`chooseDiscards`, `zones.go:61-93`).
- Shapes it uses, all `done`: `Def.Effect`, `ActEnact`, `StackManeuver` (T-0116); `ability.CostDiscard` with `Cost.UpTo` (T-0131); `Step.IfHandEmpty` (T-0132: read when the step resolves, after the steps before it); the `Repeats` timeout (T-0136); the discard picked before the targets (T-0137: `ActEnact`, `actions.go:83-91`, takes the Maneuver out of hand first, then `chooseCostDiscards`); `Target.PerDiscard` (T-0138: `Min = Max =` the cards discarded with `Repeats`, asks nothing with none discarded, never blocks enacting); `OpDealDamage` on `SubjectTarget`, one damage instance per entry; `OpDraw`; the progress policy's enacts (T-0121).
- Rulings: `SVY-R3`, `SVY-R4` (with R4a), `SVY-R16`, `SVY-R29`, `SVY-R32`. Directive #91's only *Steel Rain* part, (1), is applied (T-0136).
- The synthetic Maneuvers in `discard_cost_test.go`, `discard_order_test.go`, `hand_empty_test.go` and `per_discard_test.go` are named "Steel Rain" but built by `discardCostManeuver` and handed out with `s.give`, so the real definition doesn't change them.

### Definition (the engineer adds it to `pathDefs`)
- *Steel Rain*: `Text` holds the three lines. `ExtraCosts`: one `{Kind: ability.CostDiscard, UpTo: 3}`. `Effect`:
  - `Targets`: one slot, `{PerDiscard: true}` with an empty `Filter` (any asset).
  - `Steps`: `{Op: ability.OpDealDamage, Amount: ability.Value{N: 2}, Subject: ability.SubjectTarget, Target: 0}`, then `{Op: ability.OpDraw, Amount: ability.Value{N: 2}, IfHandEmpty: true}`.
- No `Triggers` or `Modes`. Model the cost on `discardCostManeuver` (`discard_cost_test.go:19-23`), the slot and damage step on `perDiscardManeuver` (`per_discard_test.go:17-23`), and the draw on *Improvised Orders*' `OpDraw` step.
- Remove the name from `cards.NotYetExecutable`.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
To be added to `pkg/engine/cards_path_maneuvers_test.go`, from the CSV text, loading the real card (`s.card("Steel Rain")`). Player 0 enacts; player 1 is the opponent. Give player 0 the Tactics to pay. Use units with enough health that nothing dies mid-test. A file-local chooser like `pdScript` (`per_discard_test.go`) is fine: answer the `DecideHand` discard and the `DecideTarget` picks separately.
- `TestCard_SteelRain_DealsTwoToAChosenAssetPerDiscard`: player 0's hand is *Steel Rain* and four other cards. Discarding two and choosing opposing units A then B: each takes 2, as two separate 2-damage events in pick order; the two discarded cards and then *Steel Rain* are in player 0's discard pile (`lastDiscard` is *Steel Rain*); two cards stay in hand and nothing is drawn.
- `TestCard_SteelRain_SameAssetForEveryDiscard` (`SVY-R4`, R4a, `FR-REP-004`): three discarded, all three picks on one opposing unit: it takes 6 damage as three separate 2-damage events.
- `TestCard_SteelRain_ChosenAssetIsAnyAsset` (`FR-TGT-001`, `FR-KWD-002h`): the target decision offers player 0's Headquarters, an allied unit and an allied structure, player 1's Headquarters and an opposing unit, but not an opposing Ward unit; an allied Ward unit is offered.
- `TestCard_SteelRain_DiscardingNoneDealsNothing` (`SVY-R4`, T-0138): with two other cards in hand, discarding none asks no target, deals no damage and draws nothing; both cards stay in hand.
- `TestCard_SteelRain_EmptyHandAfterTheDiscardDrawsTwo` (`FR-MNV-001`, T-0132), table: hand of *Steel Rain* and three others, discarding all three draws the top two cards of player 0's deck; discarding two of the three draws nothing.
- `TestCard_SteelRain_AloneInHandDrawsTwo`: *Steel Rain* is player 0's only card in hand. It is offered (`offersEnact`); enacting it asks no discard and no target and deals no damage, and player 0 draws 2, because the Maneuver left the hand when it was enacted and the hand is read as the draw step resolves.
- `TestCard_SteelRain_TimeoutConcentratesOnTheFirstLegalAsset` (`SVY-R32`, `FR-MOD-001` §6): three discarded, then the target decision times out. Player 0's own Headquarters takes 6 damage and no other asset takes any. Put two or more opposing units on the battlefield, so the fallback runs and not the single-option fast path.
- `TestCard_SteelRain_PaysFourTactics`: with 3 Tactics it is not offered and enacting it is rejected (`s.reject`); with 4 it is enacted, spending 4 Tactics and no Requisition.
- **Existing tests:** every other test passes unchanged, including `TestStarterSet_Executable` with the name removed, `TestView_Case8_ExecutableFlag` (seat 0's seed-1 hand still holds unsupported cards), `TestProgress_Case4_Coverage`, `TestEvents_Case4_EveryActionKind` and the synthetic "Steel Rain" tests above.

## Helpers Needed
- None expected. The tests may add their own setup helper in the test file.

## Allowed Files
- `pkg/engine/cards/path.go` (the definition)
- `pkg/engine/cards/cards.go` (remove the name from `NotYetExecutable`)

## Done When
- [ ] The card tests pass: `go test ./pkg/engine -run TestCard_SteelRain`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] The card loads with no unsupported lines (`TestStarterSet_Executable` passes with the name removed)

## Log
- 2026-11-02 lead-architect (wave C session 30): written against the run branch at ff872eb (no `main` commits to sync). #20 (b) is the oldest open directive and its shape parts are all done (T-0131, T-0132, T-0136, T-0137, T-0138), leaving this card; a deck card (#14 (a)). #91's only *Steel Rain* part, (1), is applied; its other parts go with the distinctness, *Ensnaring Mist* and *Pincer Tactics* tasks. Waiting triggers #6, #7, #8 (this task doesn't touch `deck.go`), #15, #71 (no each-player effect) not fired. Re-read `BOARD.md` before committing: no new or changed directive. Status `tests`.
- 2026-11-02 rules-test-author: wrote 8 tests in `pkg/engine/cards_path_maneuvers_test.go` (with a file-local `steelRainScene`, reusing `pdScript`/`maneuverDamage`): `TestCard_SteelRain_DealsTwoToAChosenAssetPerDiscard`, `_SameAssetForEveryDiscard`, `_ChosenAssetIsAnyAsset`, `_DiscardingNoneDealsNothing`, `_EmptyHandAfterTheDiscardDrawsTwo`, `_AloneInHandDrawsTwo`, `_TimeoutConcentratesOnTheFirstLegalAsset`, `_PaysFourTactics`. They fail now because the card is not yet executable. Helpers requested: none. Cases left out: none. Status `tests-review`.
- 2026-11-02 lead-architect (wave C session 30): tests approved (f281b03). Checked each against FRD 01: `SVY-R4`/R4a (line 365) for one independent choice per discard, repeats and separate 2-damage events; `FR-TGT-001` and `FR-KWD-002h` for "chosen asset" (any asset, both Headquarters, an allied Ward unit, never an opposing Ward unit); `SVY-R32` and `FR-MOD-001` §6 for the timeout on player 0's own Headquarters, with two opposing units so the fallback runs; T-0132's hand-empty gate, read as the draw resolves after the Maneuver left the hand. `PaysFourTactics` uses 5 Tactics and expects 1 left, which pins the cost at 4 as well. Run against a discarded trial definition (the Definition above): all 8 pass, and `go test ./...` passes with it. Status `ready`, dispatching `kernel-engineer` (flash).
