---
id: T-0221
title: A Requisition cost criterion in filters, "with Requisition cost N or less" (wave F shape, sonnet; for Devilhorn Scouting Party)
kind: shape
tier: sonnet
wave: F
status: tests-review
branch: m2/autoloop-wave-f/T-0221-requisition-at-most
depends_on: []
claims: [shape:filter Requisition cost criterion (ability.StatRequisition, read by cardStat and assetStat); the card-level filter check cardMeets (matchesCard without its pile guard)]
cards: []
requirements: [FR-TGT-001, FR-ZON-004, FR-CRD-101]
---

## Goal
*Devilhorn Scouting Party* prints "You may deploy a unit card with Requisition cost 3 or less from among them for free." The filter's stat criterion (T-0209) gets its second stat, the Requisition cost, as another value of `ability.Stat`, read the way T-0209's comment says: a card off the battlefield by its printed value (`FR-ZON-004`), an asset by the cost printed on what it currently is. The look at the top cards (T-0222) checks cards that are in a deck, not a discard pile, so this task also splits `matchesCard`'s card-level criteria out of its pile guard, so that the look calls the same check instead of copying it. First of three tasks for *Devilhorn Scouting Party*, split under the D2 size rule: the look itself is T-0222 and the card is T-0223, both `queued`.

Covers:
- `ability.StatRequisition`, the next `Stat` after `StatPower`.
- `cardStat` reads a card's printed `ReqCost` for it, `assetStat` the asset's `Printed().ReqCost`.
- `cardMeets`, the card-level part of `matchesCard` (Other, Stat, `matchesPrinted`) with no zone guard; `matchesCard` becomes its pile guard plus `cardMeets`.
- Synthetic-card tests only; no card prints it until T-0223.

- Not this task: the look at the top cards (T-0222, `queued`, #140 (3)); *Devilhorn Scouting Party* (T-0223, `queued`).

## Inputs
- FRD 01 v3.24: `FR-TGT-001` (line 373 on), above all "Readiness, Subtypes & Dynamic Stats" (line 378: a stat comparison); `FR-ZON-004` (line 244: a card off the battlefield has only its printed characteristics); `FR-CRD-101` (line 193: the printed Requisition cost).
- Card text: `data/cards/maneuvers-path.csv:6` (*Devilhorn Scouting Party*).
- Design note: `docs/architecture/m2-ability-model.md`, Key types (`Filter`, line 39: "stat comparisons").
- Code: `ability.Stat` and `Filter.Stat`, `Filter.AtMost` ([ability.go:248-268](../../../pkg/engine/ability/ability.go)); `matchesCard`, `withinStat`, `cardStat`, `assetStat` ([filter.go:22-58](../../../pkg/engine/filter.go)); `matchesFilterAt` ([filter.go:128](../../../pkg/engine/filter.go)); `Card.ReqCost` ([card.go:87](../../../pkg/engine/card.go)); `Asset.Printed` ([state.go:84](../../../pkg/engine/state.go)) and `scratchFace` ([card.go:134](../../../pkg/engine/card.go), which prints no cost). Every pile check goes through `matchesCard` and every asset check through `matchesFilterAt`, so no caller changes.

### Shape (the engineer builds it)
- **`ability.StatRequisition`** after `StatPower`, commented "Requisition cost (FR-CRD-101)". Update the `Stat` type's comment: Requisition cost is now a value; a later stat (health) is another, read the same way. Update `Filter.AtMost`'s comment example if it names only power (it may keep the power example).
- **`filter.go`:**
  - `cardStat` returns `c.ReqCost` for `StatRequisition` (printed, `FR-ZON-004`).
  - `assetStat` returns `a.Printed().ReqCost` for `StatRequisition`: the cost printed on what the asset is now (a face-down unit's is the Scratch Company face's, 0). No layer changes a Requisition cost, so printed is effective. Say so in its comment.
  - Write both as a `switch s` so a third stat is one more `case`.
  - **`(g *Game) cardMeets(ctl PlayerID, src *Instance, f ability.Filter, inst *Instance) bool`**: every card-level criterion of `f` for card `inst`, wherever it is: `Other` (not `src`), `Stat` by printed value, then `matchesPrinted` with the card's owner as its side. It reads no zone criterion. Its comment says the look at the top cards (T-0222) checks a deck card with it. `matchesCard` becomes `f.InPile() && g.cardMeets(ctl, src, f, inst)`, with its comment unchanged in meaning. This part is a pure move: no behavior changes.
- Behavior doesn't change for any printed card: none sets `StatRequisition`. `go tool task sim-compare -- autoloop/m2-wave-f` must be byte-identical (a gate here).

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
In a new file `pkg/engine/requisition_at_most_test.go`, using the scenes and helpers of `power_at_most_test.go` (`paTargetAsk`, the deploy-trigger scenes of `if_you_do_test.go`, `ppAllyUnitCards` and `ppPick` in `pile_pick_test.go`) the way T-0209's tests do. Synthetic unit cards with their `ReqCost` set per test. The filter under test is "Requisition cost 3 or less": `Stat: ability.StatRequisition, AtMost: 3`. Give every card a power that the test doesn't care about, and vary power against cost in at least one test so that reading the wrong stat fails it.
- `TestFR_TGT_001_RequisitionAtMostOffersPileCardsByPrintedCost` (`FR-TGT-001`, `FR-ZON-004`, `FR-CRD-101`): player 0's discard pile holds unit cards costing 2, 3 and 4 (the cost-4 one with power 1, the cost-2 one with power 5), and player 1's a unit card costing 1. A pile target slot with `ppAllyUnitCards` plus "Requisition cost 3 or less" offers exactly player 0's cost-2 and cost-3 unit cards (the bound is inclusive; allegiance and kind still apply).
- `TestFR_TGT_001_RequisitionAtMostPilePickReadsPrintedCost` (`SVY-R3`): a `SubjectPick` step on the same pile filter (`ppPick`) offers exactly player 0's unit cards costing 3 or less.
- `TestFR_TGT_001_RequisitionAtMostReadsAnAssetsPrintedCost` (`FR-TGT-001`): battlefield units costing 2, 3 and 4, the cost-3 one given +2 power until end of turn. An asset target slot with "unit, Requisition cost 3 or less" offers exactly the cost-2 and cost-3 units (a power modifier doesn't change a cost). Keep the Source out of the offer, or in it, and say which.
- **Existing tests:** every other test passes unchanged, `TestFR_TGT_001_PowerAtMost*` and `TestFR_TGT_001_NoStatCriterionIgnoresAtMost` included.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected. The test author keeps any small builder in the test file.

## Allowed Files
- `pkg/engine/ability/ability.go` (`StatRequisition`, the `Stat` and `AtMost` comments)
- `pkg/engine/filter.go` (`cardStat`, `assetStat`, `cardMeets`, `matchesCard`)

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run "RequisitionAtMost"`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-compare -- autoloop/m2-wave-f` passes (no printed card sets `StatRequisition`)

## Log
- 2026-11-11 lead-architect (wave F run, session 23): written against the run branch at e50acda, the first of three tasks for *Devilhorn Scouting Party* (#140 (3)). Searched `pkg/engine` and the task archive for an existing equivalent: T-0209's `ability.Stat` is the criterion, extended here by one value as its comment planned; no card-level check exists without `matchesCard`'s pile guard. The look would need five non-test files with this in it (`ability.go`, `filter.go`, `effects.go`, `choice.go`, `doc.go`), so under the D2 size rule the criterion and the `cardMeets` split come first here, the look is T-0222 (shape, sonnet, `queued`) and the card is T-0223 (card, flash, `queued`). Set to `tests`.
- 2026-11-11 rules-test-author: tests written in `pkg/engine/requisition_at_most_test.go`: `TestFR_TGT_001_RequisitionAtMostOffersPileCardsByPrintedCost`, `TestFR_TGT_001_RequisitionAtMostPilePickReadsPrintedCost`, `TestFR_TGT_001_RequisitionAtMostReadsAnAssetsPrintedCost`. Power is varied against cost in all three (cost-4 power 1, cost-2 power 5). The asset test keeps the Source out with `Other` (its cost 0 would match). Local builders `raMax3` and `raUnit` live in the test file. Helpers requested: none. Left out: none. `go vet` fails only on the missing `ability.StatRequisition`; `cardtown lint rules` passes.
