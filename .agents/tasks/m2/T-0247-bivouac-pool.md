---
id: T-0247
title: An untargeted pool split as it resolves (wave H shape, sonnet; directive #157)
kind: shape
tier: sonnet
wave: H
status: tests-review
branch: m2/autoloop-wave-h/T-0247-bivouac-pool
depends_on: [T-0245, T-0246]
claims: [shape:untargeted distributed pool (Step.Split: a SubjectPick step whose Amount is a pool its controller splits among the assets matching its filter as it resolves, one point per pick, positive whole shares summing to the pool, SVY-R51; timeout puts the whole pool on the first in fixed order, SVY-R52)]
cards: []
requirements: [FR-TGT-002, FR-TGT-003, FR-MOD-001]
---

## Goal
The pool directive #157 rules on (`SVY-R51`, `SVY-R52`), as a shape: a `SubjectPick` step with `Split: true` treats its `Amount` as a pool. As the step resolves, its controller splits the pool among the battlefield assets that match the step's filter then (the zone T-0246 locks on placement included). Each asset that receives part of it gets a positive whole amount, and the amounts add up to exactly the pool. It is not a target: no `FR-TGT-003` re-check, nothing evaporates. One match takes the whole pool with nothing asked (`FR-MOD-001` §1), none does nothing, and a timeout puts the whole pool on the first option in the fixed order `SubjectPick` uses (§6, `SVY-R52`). `OpAddArmor` is the only op that reads `Split` in this task.

*Hammerhand Bivouac*, the card, is T-0248 (split off under the D2 size rule: shape and card together touch six non-test files).

## Inputs
- FRD 01 v3.28: `FR-TGT-002` Distributed Pool Allocations (line 409, `SVY-R51`), `FR-TGT-003` Distributed Pool Evaporation (line 416), `FR-MOD-001` §1 (line 421) and §6 *Splitting a Pool* (line 437, `SVY-R52`), and §6's *Mandatory Targets and Untargeted Selections* fixed order just above it.
- Board directive #157 (`open: before T-0247`), from the [outcome](../../../docs/design/outcomes/2026-11-15-wave-h-bivouac-untargeted-pool.md) (`Status: applied: FRD 01 v3.28`). Its tests (a)–(e) are pinned here on a synthetic card; T-0248 pins them on the real card, with (f).
- Code of the day (run branch at 08219ba):
  - `Step` (`pkg/engine/ability/ability.go:385`); `SubjectPick` (`:379`).
  - `pickAssets`, `pickFrom` (`pkg/engine/targeting.go:158`, `:185`); `assetDecision` (`:108`, oldest asset first); the `PerDiscard` repeated bounds (`:97`, `d.Min, d.Max, d.Repeats = n, n, true`).
  - `stepAddArmor` (`pkg/engine/effects.go:214`); `stepAssets`' `SubjectPick` case (`:432`).
  - `Decision.Valid`, `soleAnswer`, `choose`, `fallbackPicks`, `firstPicks`, `fixedOrderPicks` (`pkg/engine/choice.go:63`, `:90`, `:117`, `:153`, `:177`, `:219`): a `Repeats` decision with `Min == Max == n` already has exactly the bounds a pool needs. A valid answer is `n` picks; with one option the sole answer is `n` copies of it, taken unasked; the fallback puts every pick on the first option in fixed order (`SVY-R32`), which is `SVY-R52`.
  - `fixedOrderPositions` (`pkg/engine/state.go:255`): the controller's Headquarters, then Reserves, then their Main Theater in `g.Theater` order (the order units entered it).
  - `Effect.ReadsZoneRef` (`ability.go:295`) already reads Steps' filters, so a `ZoneRefSame` pick step locks its source's zone on placement (T-0246).

## Shape (the lead's plan)
- `pkg/engine/ability/ability.go`: a new `Step` field, `Split bool`, commented: `SubjectPick over assets with OpAddArmor only: Amount is a pool the ability's controller splits among the assets matching Filter as the step resolves, not a target (SVY-R51): each picked asset gets a positive whole share and the shares add up to Amount; one match takes it all unasked, none does nothing, and a timeout puts it all on the first in fixed order (SVY-R52). UpTo must be 0`. Add a pointer to it in `SubjectPick`'s comment.
- `pkg/engine/targeting.go`:
  - **One repeated-picks bound.** A helper `repeatedPicks(d Decision, n int) Decision` sets `Min` and `Max` to `n` and `Repeats` to true. The `PerDiscard` branch of `targetDecision` (`:97`) uses it instead of its own line, as a pure move, and the pool uses it too, so "n picks, repeats allowed" is written once. A later targeted pool (*Macrocannon Barrage*, outside M2) splits on declaration through the same helper and the same share tally; don't build that path now.
  - **One pick decision.** Extract the decision `pickAssets` builds (its `assetDecision` call with the `matchesFilterAt` predicate) into one function that both `pickAssets` and the pool call, so the pool's options are exactly `SubjectPick`'s.
  - `splitAssets(r *resolution, st ability.Step, pool int) ([]*Asset, []int)`: with `pool <= 0` or no matching asset, nothing is asked and nothing returned. Otherwise it asks the pick decision bounded by `repeatedPicks(d, pool)`, through `pickFrom`, so `choose`'s fast path, validity check and fallback apply unchanged. It tallies the picks into recipients, in first-pick order, each with its share. The prompt names the pool: `"<card>: distribute <pool>"`.
- `pkg/engine/effects.go`: `stepAddArmor`, when `st.Split`, gives each recipient of `splitAssets` its share (`a.Armor += share`), logs one `EventArmor` per recipient with `Amount` the share, and returns the number of recipients. Without `Split` it is unchanged.
- `pkg/engine/doc.go`: one clause beside the `SubjectPick` sentence (`:46`): `Step.Split makes a pick step's Amount a pool split among the matches as it resolves (SVY-R51, SVY-R52).`
- No new `DecisionKind`: the pool is a `DecidePick` with `Repeats`. No change to `choice.go`.

## Tests (shape tasks: written by rules-test-author; engineers never edit these)
New tests go in `pkg/engine/pool_test.go`. Use a synthetic structure (`startStructure`, `turn_start_trigger_test.go:24`) for player 0 in the Main Theater with one start-of-your-turn trigger (`startSteps(ability.TurnYours, …)`) whose one step is `{Op: ability.OpAddArmor, Amount: ability.Value{N: 2}, Subject: ability.SubjectPick, Split: true, Filter: ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit, ZoneRef: ability.ZoneRefSame}}`, and run to the start of player 0's next turn (`s.endTurn()` twice). Answer through `s.script`; check asked decisions with `s.asked`. Units have no printed Armor. "Nothing is asked" means no decision of kind `DecidePick` was asked of either seat.

**About "splits like 0+2 are rejected" (#157 (a)).** The answer is one pick per point (`Picks` indexes into `Options`, with `Repeats`), so a share of 0 can't be written: an asset that gets nothing simply isn't picked, and "0 on A, 2 on B" is the answer `[1, 1]`, which is the legal "2 on one". What the per-point form can get wrong is the total or an index, so those are the rejected answers.

- `TestFR_TGT_002_UntargetedPoolSplitAsItResolves` (`SVY-R51`, #157 (a)): two allied units A (older) and B in the Main Theater. Player 0 is asked exactly one decision: `Kind` `DecidePick` (not `DecideTarget`), `Repeats`, `Min == Max == 2`, options A then B. Three subtests: answer `[0, 0]`: A gains 2, B 0; `[0, 1]`: each gains 1; `[1, 1]`: B gains 2, A 0. In each, one `EventArmor` per receiving unit, whose `Amount` is its share (2, or 1 and 1), never two events of 1 for one unit.
- `TestFR_TGT_002_UntargetedPoolRejectsWrongTotals` (`SVY-R51`, `FR-MOD-001` §6, #157 (a)): the same two units. Three subtests, answers `[0]` (a total of 1), `[0, 0, 1]` (a total of 3) and `[0, 2]` (an index out of range): each is invalid, an `EventInvalidAnswer` is logged, and the fallback applies as for a timeout: A, the first in fixed order, gains 2 and B 0.
- `TestFR_MOD_001_UntargetedPoolOneMatchTakesAll` (`FR-MOD-001` §1, #157 (b)): one allied unit in the Main Theater: it gains 2, and nothing is asked.
- `TestFR_TGT_002_UntargetedPoolNoMatchDoesNothing` (`SVY-R51`, #157 (c)): no allied unit in the Main Theater (one enemy unit there, one allied unit in Reserves): no asset's Armor changes, no `EventArmor` is logged, and nothing is asked.
- `TestFR_MOD_001_UntargetedPoolTimeoutPutsAllOnFirstInFixedOrder` (`SVY-R52`, #157 (d)): three allied units in the Main Theater, where the one with the lowest ID (so listed first in `Options`) entered the Main Theater last: put it in Reserves first, put the other two (X, then Y) in the Main Theater, then move it there (`s.moveTo`). Player 0 times out (`Answer{Timeout: true}`). X, the unit that entered the Main Theater earliest, gains 2, and the other two gain 0: the whole pool on one unit, not one point each, and by fixed order, not option order.
- `TestFR_TGT_002_UntargetedPoolOffersOnlyLockedZone` (`SVY-R44`, `SVY-R51`, #157 (e)): two allied units in the Main Theater, one allied unit in Reserves, one enemy unit in the Main Theater. The one decision's options are exactly the two allied Main Theater units: not the Reserves unit, not the enemy, not either Headquarters.
- Existing `PerDiscard` tests stay as they are and must still pass (the `repeatedPicks` move is pure).

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- none expected. If the test author needs one, it names it here.

## Allowed Files
- `pkg/engine/ability/ability.go`
- `pkg/engine/targeting.go`
- `pkg/engine/effects.go`
- `pkg/engine/doc.go`

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run "UntargetedPool|PerDiscard"`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-compare -- autoloop/m2-wave-h` passes (no card uses `Split` yet, and the `PerDiscard` change is a pure move)

## Log
- 2026-11-15 lead-architect (wave H run, session 7): queued after T-0246, which it depends on for the locked zone. Directive #157 is `open: before T-0247` and is applied here.
- 2026-11-15 lead-architect (wave H run, session 8): T-0245 and T-0246 are `done`. Wrote Inputs, Shape, Tests and Allowed Files against the run branch at 08219ba, and split the card off as T-0248 (`queued`) under the D2 size rule. The pool reuses `choose`'s `Repeats` bounds, fast path and fallback, which already give `SVY-R52`; the new code is the tally into shares, and `repeatedPicks` keeps the "n picks with repeats" bound in one place (`PerDiscard` moves to it). Searched `pkg/engine` and the task archive for an existing pool or distribute: none (as #157 found). No Ward test: Ward only stops an opposing controller's targets (`filter.go:81`), and the pool's filter is allied. Marked design note Open Question 24 answered. Re-read `BOARD.md` before committing: no directive added since; #157 is the one this task and T-0248 apply. Status `tests`.
- 2026-11-15 rules-test-author (wave H run): wrote `pkg/engine/pool_test.go`: `TestFR_TGT_002_UntargetedPoolSplitAsItResolves` (3 subtests), `TestFR_TGT_002_UntargetedPoolRejectsWrongTotals` (3 subtests), `TestFR_MOD_001_UntargetedPoolOneMatchTakesAll`, `TestFR_TGT_002_UntargetedPoolNoMatchDoesNothing`, `TestFR_MOD_001_UntargetedPoolTimeoutPutsAllOnFirstInFixedOrder`, `TestFR_TGT_002_UntargetedPoolOffersOnlyLockedZone`. Local test-file helpers only (`poolStructure`, `poolScene`, `(*scene).runPool`, `picksAsked`); none requested for `scenario_test.go`. `go vet` fails only on the missing `ability.Step.Split` field (the shape itself). `cardtown lint rules` passes. No case left out. Status `tests-review`.
