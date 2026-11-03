---
id: T-0138
title: A target slot can choose once per card the discard cost discarded (FR-TGT-001, SVY-R4, SVY-R32, directive #20 (b), wave C shape, sonnet)
kind: shape
tier: sonnet
wave: C
status: tests-review
branch: m2/autoloop-wave-c/T-0138-per-discard-target-slot
depends_on: [T-0136, T-0137]
claims: [field:Target.PerDiscard, func:chooseTargets, func:targetDecision, func:slotsLegal]
cards: []
requirements: [FR-TGT-001, FR-TRN-003a, FR-MNV-001, FR-MOD-001]
---

## Goal
A Maneuver can print *"For each card discarded this way, deal 2 damage to chosen asset."* (*Steel Rain*, a deck card). On declaration its controller makes one target choice for each card its discard cost discarded, independently, so the same asset may be chosen more than once, and each repetition is its own damage instance (`SVY-R4`, `FR-REP-004`). A timeout puts them all on the first legal entity in fixed order (`SVY-R32`, T-0136). The last of three shape tasks split from #20 (b) under the D2 size rule (T-0136 and T-0137 are done); the *Steel Rain* card task follows it.

## Inputs
- FRD 01 (run branch, v3.20):
  - `FR-TGT-001`, Repeated Targets (`SVY-R4`, line 365): "A targeted clause repeated *"for each"* outcome makes a separate target choice for each repetition, all on placement. The choices are independent, so the same entity may be chosen more than once, and each repetition is its own damage instance (`FR-REP-004`). … On a timeout every repetition picks the same entity, the first legal one in fixed order (`FR-MOD-001` §6, `SVY-R32`)."
  - `FR-TGT-001`, Targets Are Locked on Placement (`SVY-R3`, lines 360 and 362): *Steel Rain*'s discard is paid on declaration, "so the number of targets is known then, and one asset is chosen for each card discarded at that same moment."
  - `FR-TGT-001`, "Up To N" Targets Are Distinct (`SVY-R11`, line 369): "Only a clause repeated *"for each"* outcome may choose the same entity more than once."
  - `FR-MNV-001`, A Maneuver Needs Its Required Targets (`SVY-R29`, line 340): "only if each of its exactly-one target slots has a legal entity when it is declared … An *"up to N"* slot never stops it."
  - `FR-MOD-001` §6 (line 397): a per-repetition choice times out to the first legal entity for every repetition (`SVY-R32`); fixed order is the choosing player's assets first, Headquarters first, then Reserves, then the Main Theater.
  - `FR-KWD-002h` (Ward): an opponent's Ward asset is never a legal target.
  - `FR-TGT-003`: a slot with no legal entity stays empty; the second verification stage drops each entry that is gone or no longer legal.
- Survey: *Steel Rain* row (`docs/design/m2_starter_set_ability_survey.md` line 180): "Targets chosen with the discard; repeats allowed; timeout concentrates on one asset (R32)". Rulings R4, R4a (lines 62-63), R32 (line 91).
- Rulings that apply: `SVY-R3`, `SVY-R4`, `SVY-R11`, `SVY-R16`, `SVY-R29`, `SVY-R32`.
- Directive #20 (b) (`BOARD.md`): "one target choice per card the cost discarded, fixed on declaration".
- **Lead's reading of `SVY-R29` (binding for this task):** a per-discard slot is not an exactly-one slot. Its controller can always discard none, which pays the cost (`costsPayable`, `placement.go:76-87`) and then needs no target, so a per-discard slot never makes a Maneuver illegal to enact. Legality is judged before the discard is picked (`placeable` runs on the legal-action list), so the count isn't known then. If cards are discarded while the slot has no legal entity, the slot stays empty (`FR-TGT-003`). For *Steel Rain*'s "chosen asset" this can't happen in play: the controller's own Headquarters is always a legal entity.
- Kernel today (run branch at 03f8983):
  - `ability.Target` (`ability/ability.go:152-155`): `Filter`, `UpTo` (0: exactly one; N: up to N distinct).
  - `ActEnact` (`actions.go:83-91`): `pl.discards = g.chooseCostDiscards(pl)` (T-0137), then `chooseOnPlacement` (modes, then `chooseTargets`), `push`, `payCosts`. A trigger's placement (`trigger.go:199`) has no costs and nil `discards`.
  - `chooseTargets` (`targeting.go:20-35`) asks `targetDecision` per slot; it skips an "up to N" slot with no options, and an exactly-one slot with no options asks nothing because `choose` finds it unsatisfiable.
  - `targetDecision` (`targeting.go:41-56`) builds bounds through `assetDecision`/`pileDecision` and `pickBounds`. `legalModes` and `slotsLegal` (`placement.go:125-133`) read its `Options`.
  - `Decision.Repeats` (`choice.go:41`): `Valid`, `satisfiable`, `soleAnswer` (one option and `Repeats`: `Min` copies of it, answered unasked), `fixedOrderPicks` (every pick `idx[0]`, `SVY-R32`, T-0136) and `RandomChooser` already handle it. With `Min = Max = 0` and two or more options, `choose` would still ask the Chooser, so `chooseTargets` must skip the slot itself.
  - Resolution: `resolveAbility` (`effects.go:28-48`) verifies each entry of `it.Chosen[i]` separately and keeps repeats; `stepDealDamage` (`effects.go:138-154`) calls `dealDamage` once per entry of `stepAssets`, so each repetition is already its own damage instance.
  - No card prints a per-discard slot yet; *Steel Rain* is the first (the next task).

### Shape (lead decisions, binding)
- **`ability.Target.PerDiscard bool`** (`ability/ability.go`): the slot is chosen once for each card in its placement's `discards`, each choice an exactly-one pick, independent of the others, so repeats are allowed (`SVY-R4`). `UpTo` stays 0 with it. Doc comment on the field, and the `UpTo` comment keeps saying what it says.
- **`targetDecision`** (`targeting.go`): for a `PerDiscard` slot, the same options as today, with `Min = Max = len(pl.discards)` and `Repeats: true`. Prompt: `"<label>: choose target <i+1> (<n> times)"`. Other slots unchanged.
- **`chooseTargets`** (`targeting.go`): a `PerDiscard` slot with no discards or no options asks nothing and stays empty. Otherwise it asks once and appends each pick's `Ref`, in pick order, repeats included. Its doc comment states this.
- **`slotsLegal`** (`placement.go`): skips `PerDiscard` slots, as it skips "up to N" ones; doc comment cites the reading above. `legalModes` and `placeable` are unchanged and get it through `slotsLegal`.
- Nothing else changes: `resolveAbility`, `slotAssets`, `stepAssets` and `stepDealDamage` already keep repeated entries. `chooseCostDiscards` already runs before `chooseOnPlacement`.
- **Sims:** no card uses `PerDiscard`, so `go tool task sim-compare -- autoloop/m2-wave-c` must pass.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
In a new file `pkg/engine/per_discard_test.go`, with synthetic Maneuvers built by `discardCostManeuver` (`discard_cost_test.go`): "As an additional cost, discard up to 3 cards. For each card discarded this way, deal 2 damage to chosen asset" (an `OpDealDamage` 2 step on slot 0, `ability.Target{PerDiscard: true}` with an empty filter unless a test says otherwise). Use units with enough health that nothing dies mid-test unless the test says so. No existing test is edited.
- `TestFR_TGT_001_PerDiscardSlotChoosesOncePerDiscard` (`SVY-R4`, `SVY-R3`), table over 0, 1 and 3 cards discarded, with at least two legal assets: the `DecideTarget` decision has `Min = Max =` the number discarded and `Repeats`, and is asked after the `DecideHand` discard. With 0 discarded, no `DecideTarget` is asked and no damage is dealt.
- `TestFR_TGT_001_PerDiscardSameAssetEachTime` (`SVY-R4`, R4a, `FR-REP-004`): 3 discarded, all three picks on one opposing unit. It takes 6 damage as three separate 2-damage events from the Maneuver.
- `TestFR_TGT_001_PerDiscardSplitAcrossAssets` (`SVY-R4`): 3 discarded, picks A, B, A. A takes 4 and B takes 2, and the damage events follow pick order.
- `TestFR_MOD_001_PerDiscardTimeoutConcentrates` (`SVY-R32`, §6): 3 discarded, the target decision times out. All three picks land on the first legal asset in §6 fixed order (6 damage on it, none elsewhere). Use the empty filter, and give the test a second legal asset so the fallback is exercised, not the single-option fast path.
- `TestFR_KWD_002h_PerDiscardSlotNeverOffersOpposingWard` (`FR-KWD-002h`): an opposing Ward unit is not among the target decision's options; an allied Ward unit is.
- `TestFR_MNV_001_PerDiscardSlotNeverBlocksEnact` (lead's reading of `SVY-R29` above): with the slot filtered to opposing units and no opposing unit on the battlefield, the Maneuver is among player 0's legal actions. Enacted with 2 cards discarded, no `DecideTarget` is asked, both cards end in the discard pile and no damage is dealt.
- `TestFR_TGT_001_PerDiscardSlotOnTriggerChoosesNothing` (`FR-TRN-003a`: only an enacted card pays additional costs): a triggered ability whose effect has a `PerDiscard` slot (for example a deploy trigger, as in the existing trigger tests) asks no `DecideTarget` and deals no damage.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected. `discardCostManeuver`, `discardCostScene`, `dcScript` (`discard_order_test.go`), `s.asked`, `inDiscard` and the trigger-test helpers exist. A file-local helper in `per_discard_test.go` is fine.

## Allowed Files
- `pkg/engine/ability/ability.go` (`Target.PerDiscard` only)
- `pkg/engine/targeting.go` (`chooseTargets`, `targetDecision` and their doc comments)
- `pkg/engine/placement.go` (`slotsLegal` and its doc comment)

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run 'PerDiscard'`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim-compare -- autoloop/m2-wave-c` passes (no card uses `PerDiscard` yet)

## Log
- 2026-11-02 lead-architect (wave C session 27): queued with T-0136 when #20 (b) was split (D2.4). Write its Inputs, Tests and Done When against the code of the day before it leaves `queued`.
- 2026-11-02 lead-architect (wave C session 29): written against the run branch at 03f8983 (no `main` commits to sync). #20 (b) is the oldest open directive and this is its last shape part; #91's open parts sit with the tasks whose `cards:` include *Steel Rain*, *Ambush Tactics*, *Ensnaring Mist* or *Pincer Tactics*, and this task's `cards:` is empty. Re-read `BOARD.md` before committing: no new or changed directive. Status `tests`.
- 2026-11-02 rules-test-author: wrote pkg/engine/per_discard_test.go: TestFR_TGT_001_PerDiscardSlotChoosesOncePerDiscard, TestFR_TGT_001_PerDiscardSameAssetEachTime, TestFR_TGT_001_PerDiscardSplitAcrossAssets, TestFR_MOD_001_PerDiscardTimeoutConcentrates, TestFR_KWD_002h_PerDiscardSlotNeverOffersOpposingWard, TestFR_MNV_001_PerDiscardSlotNeverBlocksEnact, TestFR_TGT_001_PerDiscardSlotOnTriggerChoosesNothing. All fail to compile today (unknown field ability.Target.PerDiscard). Helpers requested: none (file-local pdScript, perDiscardScene, maneuverDamage). Left out: nothing. Notes: the timeout test expects the choosing player's own Headquarters (first in fixed order with an empty filter); the 3-discard split test uses distinct picks via the first three hand options. go vet fails only on PerDiscard; cardtown lint rules passes. Status tests-review.
