---
id: T-0213
title: Deploy goes through declare, so a deployed card's additional costs are paid (wave F shape, sonnet)
kind: shape
tier: sonnet
wave: F
status: tests
branch: m2/autoloop-wave-f/T-0213-deploy-declare
depends_on: []
claims: [shape:deployPlacement (a deploy's placement; Deploy through declare; placement.zone)]
cards: []
requirements: [FR-TRN-003, FR-TRN-003a]
---

## Goal
A Deploy action goes through the same declare sequence as an enacted Maneuver and an activated ability: the card leaves its pile and is paid for, a placement for the deployed card is filled into its `StackDeploy` item, the item is pushed, and then the card's additional costs (`Def.ExtraCosts`) are paid, with the triggers they fire going on top of it (`FR-TRN-003a`). A Deploy whose additional costs can't be paid is not legal (`SVY-R29`). This is directive #128's first part: *Veil Render*'s discard deploy (T-0214, T-0216) adds its kill cost to this placement instead of copying the sequence.

Covers:
- `deployPlacement` in `placement.go`, and a `zone` field on `placement` that `stackItem` copies into `StackItem.Zone`.
- `Apply`'s `ActDeploy` case calls `declare` with it; `check`'s `ActDeploy` case asks `placeable` of it.
- Synthetic-card tests only. No printed Unit or Structure has `ExtraCosts`, so behavior doesn't change for any card.

- Not this task: deploying from a discard pile (`Action.From`, the discard-active permission and its costs, the action event's `From`: T-0214, `queued`); replay, the play screen and `ProgressPolicy` for a discard deploy (T-0215, `queued`); *Veil Render* (T-0216, `queued`).

## Inputs
- FRD 01 v3.23: `FR-TRN-003` and `FR-TRN-003a` (lines 356–359: the card moves to the stack on declaration and every cost is paid upfront; an additional cost's triggers go on top of the declared card; a replaced kill cost still counts as paid, `SVY-R15`); `FR-MNV-001`'s required-target rule (line 358, `SVY-R29`) as `placeable` already applies it to costs.
- Board directive #128 (`BOARD.md`): Fix and Acceptance, first two items.
- Design note: `docs/architecture/m2-ability-model.md`, Key types (`Def.ExtraCosts`, line 31) and Changes in `engine` (Actions, line 54).
- Code: `Apply` and `check` ([actions.go:50-89](../../../pkg/engine/actions.go), [:120-131](../../../pkg/engine/actions.go)); `handPlay` and `playFromHand` ([actions.go:182-205](../../../pkg/engine/actions.go)); `placement`, `stackItem`, `declare`, `maneuverPlacement`, `activatedPlacement` and `placeable` ([placement.go:13-70](../../../pkg/engine/placement.go), [:116-126](../../../pkg/engine/placement.go)); `costsPayable`, `otherCostsPayable` and `payCosts` ([costs.go](../../../pkg/engine/costs.go)); `resolveDeploy` ([stack.go:127](../../../pkg/engine/stack.go)).
- Model tests: `kill_cost_test.go` (a Maneuver's kill cost: paid after the push, its triggers on top, `TestFR_TRN_003a_ReplacedKillCostStaysPaid`).

### Shape (the engineer builds it)
- **`placement.zone Zone`**: a deploy's declared battlefield zone; zero for every other placement. `stackItem` sets `StackItem.Zone` from it.
- **`deployPlacement(pid PlayerID, inst *Instance, zone Zone) placement`**, beside `maneuverPlacement`: controller `pid`, `inst`, label `inst.Card.Name`, `zone`, and costs `inst.Card.Ability.ExtraCosts` when the card has an `Ability`. Used for both the legality check and declaring.
- **`Apply`, `ActDeploy`:** `inst := g.playFromHand(p, a.Card)`, then `g.declare(deployPlacement(a.Player, inst, a.Zone), ability.Effect{}, StackDeploy)`. The direct `g.push(&StackItem{Kind: StackDeploy, …})` goes.
- **`check`, `ActDeploy`:** after the zone check, the action is illegal if `!g.placeable(deployPlacement(a.Player, inst, a.Zone), ability.Effect{})`, with a message naming the card and `FR-TRN-003a`.
- Nothing else changes: `resolveDeploy`, `deployCard`, the events and free deploys (`stepDeployFree`, `FR-TRN-003a`'s in-resolution deploys, which have no placement and no costs) stay as they are. A deploy with no additional costs asks no new decision and logs the same events in the same order, so `go tool task sim-compare -- autoloop/m2-wave-f` must be byte-identical (a gate here).

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
In a new file `pkg/engine/deploy_cost_test.go`, with a synthetic Unit card for player 0 (`unit`, with an `Ability` whose `ExtraCosts` is one `{Kind: ability.CostKill, Filter: ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit}}` and nothing else), given to player 0's hand with enough resources, deployed with `Action{Kind: ActDeploy, Player: 0, Card: inst.ID, Zone: ZoneReserves}`. Model the tests on `kill_cost_test.go`. Script only the decisions actually asked: a single option is taken unasked (`FR-MOD-001` §1).
- `TestFR_TRN_003a_DeployPaysItsKillCostAfterThePush` (`FR-TRN-003`, `FR-TRN-003a`): two allied units, one of them printing a death trigger (`OpDraw` 1 for its controller). Player 0 picks that one for the cost. After the action, it is destroyed and in player 0's discard pile, the other unit is untouched, the deployed card is an asset in player 0's Reserves, and player 0 paid the card's Requisition and Tactics. In the event log, the deploy's `place` event comes before the cost's `destroy` event, and the death trigger's draw comes before the card's `deploy` event (the trigger went on top of the deploy and resolved first). The kill cost was asked as one `DecidePick` listing both allied units.
- `TestFR_TRN_003a_DeployKillCostNeedsAnAlliedUnit` (`SVY-R29`): with no allied unit on the battlefield (only player 0's Headquarters), the Deploy is not in `LegalActions` and `Apply` returns `ErrIllegal`; the card stays in hand and player 0's pool is unchanged. With one allied unit it is legal, and that unit is killed without a question.
- `TestFR_TRN_003a_DeployReplacedKillCostStaysPaid` (`SVY-R15`, `FR-REP-001`): the only allied unit has a destruction replacement that applies to it (as in `TestFR_TRN_003a_ReplacedKillCostStaysPaid`). The cost kills it, the replacement replaces the death, and the deploy still resolves: the card is an asset in player 0's Reserves.
- `TestFR_TRN_003_DeployWithoutAdditionalCostsAsksNothing` (`FR-TRN-003`): a synthetic Unit with no `Ability` deploys to the Main Theater as before: no decision is asked, the stack item's zone is the declared zone, and the asset enters there.
- **Existing tests:** every other test passes unchanged.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected. The test author keeps any small builder in the test file.

## Allowed Files
- `pkg/engine/actions.go` (`Apply` and `check`, `ActDeploy`)
- `pkg/engine/placement.go` (`placement.zone`, `stackItem`, `deployPlacement`)

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run "TestFR_TRN_003"`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-compare -- autoloop/m2-wave-f` passes (no printed Unit or Structure has `ExtraCosts`)

## Log
- 2026-11-10 lead-architect (wave F run, session 13): written against the run branch at 89c6c18, for *Veil Render*, the first card of the survey's row F still to build. Its task fires directive #128 (deploying from a discard pile goes through `declare`), which is now `open: before T-0216`. Searched `pkg/engine` and the task archive for an existing equivalent: `declare` and `placement` already serve Maneuvers and activated abilities (T-0163), and only Deploy pushes its item directly ([actions.go:62-63](../../../pkg/engine/actions.go)), so this task routes it through them and adds no second sequence. *Veil Render* with #128 needs about nine non-test files across `engine`, `eventlog`, `sim` and `cmd/cardtown`, so under the D2 size rule it splits into this task, T-0214 (the discard-active deploy permission and `Action.From`, shape, sonnet, `queued`), T-0215 (replay, play screen and `ProgressPolicy` for a discard deploy, harness, `queued`) and T-0216 (*Veil Render*, card, flash, `queued`). Set to `tests`.
- 2026-11-10 rules-test-author: tests written in `pkg/engine/deploy_cost_test.go`: `TestFR_TRN_003a_DeployPaysItsKillCostAfterThePush`, `TestFR_TRN_003a_DeployKillCostNeedsAnAlliedUnit`, `TestFR_TRN_003a_DeployReplacedKillCostStaysPaid`, `TestFR_TRN_003_DeployWithoutAdditionalCostsAsksNothing`. Helpers requested: none (file-local `killCostUnit`, `dcDeployToReserves`, `dcOffersDeploy`, `dcAssetOf`; reuses `kcScript`, `alliedUnitFilter`, `recordPlaced`). Vet clean; the first three fail now as expected, the fourth passes already (it pins unchanged behavior). `cardtown lint rules` passes. Left out: nothing. Note: in the replaced-kill test the Shifting Impossibility is itself an allied unit, so the cost asks a pick and the test scripts the Scratch Company (same as `TestFR_TRN_003a_ReplacedKillCostStaysPaid`).
- 2026-11-10 lead-architect (wave F run, session 13): tests reviewed (b0fef2d), returned to `tests`. A lead probe (the two Allowed Files edited as the Shape says, since discarded) passes three of the four tests. Notes:
  1. `TestFR_TRN_003a_DeployPaysItsKillCostAfterThePush` fails on that correct build with `missing events: … draw 0`: a successful draw logs no event (`moveTopToHand`, [zones.go:16-22](../../../pkg/engine/zones.go); only a draw from an empty deck is logged), so the `draw` event never appears. The probe's log reads `place` (Deploy of Sacrificer), the cost's `decision`, `destroy` (Martyr), `place` (ability of Martyr), then `deploy` (Sacrificer). Pin the order with what is logged: the Martyr ability's `place` event comes after the deploy's `place` and before its `deploy` event (it went on top and resolved first), and player 0's hand grew by one card (the trigger's draw resolved). Keep the other checks.
  The other three tests match the Tests list and `FR-TRN-003`, `FR-TRN-003a`, `SVY-R15` and `SVY-R29`, and stand as they are.
