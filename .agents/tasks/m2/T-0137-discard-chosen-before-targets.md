---
id: T-0137
title: A discard cost's cards are chosen before the targets and discarded after the push (FR-TRN-003a, SVY-R2, SVY-R4, directive #20 (b), wave C shape, sonnet)
kind: shape
tier: sonnet
wave: C
status: tests
branch: m2/autoloop-wave-c/T-0137-discard-chosen-before-targets
depends_on: [T-0136]
claims: [field:placement.discards, func:chooseDiscards, func:discardCards, func:discardFromHand, func:chooseCostDiscards, func:payCosts]
cards: []
requirements: [FR-TRN-003a, FR-TRN-003, FR-TGT-001, FR-MNV-001]
---

## Goal
When a Maneuver with a discard cost is enacted, its controller picks the cards to discard first, then its modes and targets are chosen, then it is pushed and the picked cards are discarded. So the number of cards discarded is known when the targets are chosen (*Steel Rain*: one target per card discarded, `SVY-R4`), while targets are still judged against the state before the cost moved anything (`SVY-R2`). The second of three shape tasks split from #20 (b) under the D2 size rule: T-0136 (done) made a repeated-choice timeout concentrate, and T-0138 (queued) adds the per-discard target slot that reads the cards this task records. Then the *Steel Rain* card task.

## Inputs
- FRD 01 (run branch, v3.20):
  - `FR-TRN-003a`, Additional Sacrifice, Discard & Exhaust Costs (line 341): the cost "is paid immediately upon declaration. Any abilities triggered by paying that additional cost enqueue onto the Resolution Stack **on top of** the declared card or ability, resolving before it."
  - `FR-TGT-001`, Targets and Costs Are Committed Together (`SVY-R2`, line 359): "Choosing targets and paying costs happen at the same moment of declaration, and target eligibility is judged against the game state at that moment. A card that moves to a new zone by paying a cost … is not yet in that zone when targets are chosen."
  - `FR-TGT-001`, Targets Are Locked on Placement (`SVY-R3`, lines 360 and 362): "Where the number or identity of targets is meant to depend on an earlier clause, that clause is an additional cost so that its outcome is known on declaration (`FR-TRN-003a`)." *Steel Rain*: "the discard is paid on declaration, so the number of targets is known then, and one asset is chosen for each card discarded at that same moment."
  - `FR-TGT-001`, Repeated Targets (`SVY-R4`, line 365).
  - `FR-MOD-001` §6 (line 397): an *"up to N"* selection times out to none.
- Survey: *Steel Rain* row (`docs/design/m2_starter_set_ability_survey.md` line 180): "Targets chosen with the discard".
- Rulings that apply: `SVY-R2`, `SVY-R3`, `SVY-R4`, `SVY-R16`.
- Directive #20 (b) (`BOARD.md`).
- Kernel today (run branch at 446b753):
  - `Apply`'s `ActEnact` (`actions.go:83-90`): `payCost`, `takeCard`, `maneuverPlacement`, then `chooseOnPlacement` (modes, then targets), `push`, then `payCosts`.
  - `payCosts` (`placement.go:88-106`): in printed cost order; a `CostDiscard` calls `discardFromHand(pl.controller, c.UpTo, true)`, which both asks and moves, so the discard decision comes **after** the target decisions today. Kill and exhaust costs pick their asset there too (`costDecision`, `pickFrom`).
  - `discardFromHand` (`zones.go:60-92`): builds the `DecideHand` decision over the whole hand (prompt, `Min`/`Max`, `upTo`), asks it, then takes each picked card, puts it in the discard pile and logs `EventDiscard` with zone `"Hand"`. Its other caller is `stepDiscard` (`effects.go:216-222`).
  - `placement` (`placement.go:15-21`) is passed by value to `chooseOnPlacement`, `chooseTargets` and `payCosts`. A trigger's placement (`trigger.go:199`) has no costs.
  - No card uses `CostDiscard` yet (*Steel Rain* is the first, after T-0138); only the T-0131 and T-0132 tests' synthetic Maneuvers do.

### Shape (lead decisions, binding)
- **Split `discardFromHand`** (`zones.go`), so one function still builds the hand-discard decision:
  - `chooseDiscards(pid PlayerID, n int, upTo bool) []*Instance` builds and asks exactly the decision `discardFromHand` asks today (same prompts, bounds, options and fast path) and returns the picked cards in pick order. It moves nothing.
  - `discardCards(pid PlayerID, cards []*Instance)` takes each card from its pile, puts it in its owner's discard pile and logs `EventDiscard` (zone `"Hand"`), in the given order, as the second half of `discardFromHand` does today.
  - `discardFromHand` stays, as `chooseDiscards` then `discardCards`, returning the cards. `stepDiscard` is unchanged. Doc comments: every hand-discard decision is built by `chooseDiscards`, and every discard from hand is moved by `discardCards`.
- **`placement.discards []*Instance`** (`placement.go`): the hand cards its discard cost picked on declaration, in pick order. Nil for a trigger and for a Maneuver without a discard cost.
- **`chooseCostDiscards(pl placement) []*Instance`** (`placement.go`): for pl's `CostDiscard` (`chooseDiscards(pl.controller, c.UpTo, true)`), the cards picked. A placement has at most one discard cost (no card prints two); a second one panics, like `payCosts`' unknown cost kind. Without one it returns nil and asks nothing.
- **`ActEnact`** (`actions.go`, that case only): after `takeCard(inst)` and `maneuverPlacement`, set `pl.discards = g.chooseCostDiscards(pl)`, then `chooseOnPlacement`, `push` and `payCosts` as today. The Maneuver has left the hand, so it is never a discard option (`FR-TRN-003`, pinned by `TestFR_TRN_003a_DiscardCostOffersTheRestOfTheHand`).
- **`payCosts`**: a `CostDiscard` now calls `discardCards(pl.controller, pl.discards)` (it asks nothing). Kill and exhaust costs are unchanged: they still pick after the targets, after the push. Its doc comment says the discard's cards were picked before the modes and targets (`chooseCostDiscards`) and are moved here, so targets are still judged before any card moves (`SVY-R2`).
- Nothing else changes: the picked cards stay in hand while modes and targets are chosen (nothing reads the hand on placement), the event order (Maneuver `place`, then the cost's `discard` events, then resolution) is unchanged, and `costsPayable` still treats a discard cost as always payable.
- **Sims:** no card uses `CostDiscard`, so `go tool task sim-compare -- autoloop/m2-wave-c` must pass (starter and stress sims byte-identical to the run branch).

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
In a new file `pkg/engine/discard_order_test.go`, with synthetic Maneuvers built by `discardCostManeuver` (`discard_cost_test.go`). No existing test is edited; the T-0131 and T-0132 tests (`TestFR_TRN_003a_DiscardCost*`, `TestFR_MOD_001_DiscardCostTimeoutDiscardsNone`, `TestFR_TRN_003a_DiscardCostThenHandEmptyStep`) must keep passing as they are.
- `TestFR_TRN_003a_DiscardChosenBeforeTargets` (`SVY-R3`, `SVY-R4`): a Maneuver "As an additional cost, discard up to 3 cards. Deal 1 damage to chosen opposing unit", two opposing units on the battlefield (so the target is asked), two other cards in player 0's hand. Player 0 discards both and targets the second unit. The decisions asked of player 0 are, in order, the `DecideHand` discard and then the `DecideTarget`, and nothing else. Both cards end in the discard pile and the chosen unit takes 1 damage. Fails today: the target is asked first.
- `TestFR_TRN_003a_DiscardChosenBeforeTargetsOnTimeout` (`FR-MOD-001` §6): the same Maneuver with a timed-out discard decision: nothing is discarded, the target decision is still asked after it, and its pick takes effect.
- `TestFR_TGT_001_DiscardCostCardsNotTargetableInPile` (`SVY-R2`): a Maneuver "As an additional cost, discard up to 2 cards. Return chosen unit card in your discard pile to your hand" (an allied unit-card pile slot as in `TestFR_TRN_003a_CostPaidAfterTargetsAreChosen`, `kill_cost_test.go`), player 0's discard pile already holding two unit cards C and D, and two unit cards A and B in hand. Player 0 discards A and B. The target decision's options are exactly C and D, never A or B; the chosen pile card returns to hand, and A and B end in the discard pile. A guard: it passes today too, and must still pass once the cards are picked before the targets.
- `TestFR_TRN_003a_DiscardCostEventsStillAfterPlace`: in the first test's scene, the two `discard` events come after the Maneuver's `place` event and before its `damage` event, and are logged in pick order (pick B before A and check the order). Guard for the push-then-pay order (`FR-TRN-003a`).

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected. `discardCostManeuver`, `hqPing`, `s.script`, `s.asked`, `kindsOf`, `inDiscard`, `inHand` and `s.inDiscard` exist.

## Allowed Files
- `pkg/engine/zones.go` (`discardFromHand`, `chooseDiscards`, `discardCards`)
- `pkg/engine/placement.go` (`placement`, `chooseCostDiscards`, `payCosts` and their doc comments)
- `pkg/engine/actions.go` (`Apply`'s `ActEnact` case only)

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run 'TestFR_TRN_003a_Discard|TestFR_TGT_001_DiscardCost|TestFR_MOD_001_DiscardCost'`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim-compare -- autoloop/m2-wave-c` passes (no card uses `CostDiscard` yet)

## Log
- 2026-11-02 lead-architect (wave C session 27): queued with T-0136 when #20 (b) was split (D2.4). Write its Inputs, Tests and Done When against the code of the day before it leaves `queued`.
- 2026-11-02 lead-architect (wave C session 28): written against the run branch at 446b753 (no `main` commits to sync). #20 (b) is the oldest open directive and its next part; #91 (1) is applied, and #91's other parts sit with the tasks whose `cards:` include *Steel Rain*, *Ambush Tactics*, *Ensnaring Mist* or *Pincer Tactics*, which this task's empty `cards:` doesn't. Added `chooseCostDiscards` to the claims, so `ActEnact` gains one line. Re-read `BOARD.md` before committing: no new or changed directive. Status `tests`.
