---
id: T-0136
title: A timed-out repeated choice puts every repetition on the first option (SVY-R32, directive #91 (1), #20 (b), wave C shape, sonnet)
kind: shape
tier: sonnet
wave: C
status: tests
branch: m2/autoloop-wave-c/T-0136-repeat-timeout-concentrates
depends_on: []
claims: [func:fixedOrderPicks, func:fallbackPicks]
cards: []
requirements: [FR-MOD-001, FR-TGT-001]
---

## Goal
When a `Repeats` decision (the same option may be picked more than once) times out, gets an invalid answer or has no chooser, every pick is the first option in the §6 order, never a cycle through the options. For a target choice made once per repetition (*Steel Rain*'s "for each card discarded this way, deal 2 damage to chosen asset", a deck card), all the repetitions land on the first legal entity in fixed order (`SVY-R32`). Directive #91 (1). The first of three shape tasks split from #20 (b) under the D2 size rule: T-0137 (queued) chooses a discard cost's cards before the targets, and T-0138 (queued) adds the per-discard target slot. Then the *Steel Rain* card task.

## Inputs
- FRD 01 (run branch, v3.20):
  - `FR-MOD-001` §6, *Mandatory Targets and Untargeted Selections* (`SVY-R16`, line 397): "A target choice made separately for each repetition (`SVY-R4`) times out to the first legal entity for every repetition, so all of them land on one entity: a timed-out *Steel Rain* that discarded 3 cards deals all 6 damage to the first legal asset (`SVY-R32`)." The same line gives the fixed order: the choosing player's own assets first, then the opponent's; within each player the Headquarters, then Reserves, then the Main Theater, each zone in arrival order.
  - `FR-TGT-001`, Repeated Targets (`SVY-R4`, line 365): "On a timeout every repetition picks the same entity, the first legal one in fixed order (`FR-MOD-001` §6, `SVY-R32`)."
  - Version notes, line 18: `SVY-R32`.
- Rulings that apply: `SVY-R4`, `SVY-R16`, `SVY-R32`.
- Directive #91 (1) (`BOARD.md`): `fixedOrderPicks` returns `idx[0]` for every pick of a `Repeats` decision instead of cycling. "This applies to every `Repeats` decision, so a mode that may repeat (none exists yet) also times out to the first listed legal mode each time." Its test name: `TestFR_MOD_001_RepeatedTargetTimeoutConcentrates`.
- Outcome: `docs/design/outcomes/2026-11-02-wave-c-repeated-distinct-targets-and-mode-costs.md`, question 1.
- Survey: *Steel Rain* row (`docs/design/m2_starter_set_ability_survey.md` line 180): "timeout concentrates on one asset (R32)".
- Kernel today (run branch at the merge of `main`'s 2dbd536):
  - `fixedOrderPicks` (`choice.go:187-211`): sorts option indexes into fixed order, then for a `Repeats` decision returns `idx[i%len(idx)]`, cycling.
  - `fallbackPicks`' `DecideMode` case (`choice.go:146-156`): for a `Repeats` decision, `picks[i] = i % len(d.Options)`, cycling.
  - The fast path (`soleAnswer`, `choice.go:92`) already gives `Min` copies of option 0 for a `Repeats` decision with one option (`TestFR_MOD_001_FastPathRepeatsSingleOptionNotAsked`), and is unchanged.
  - No engine code builds a `Repeats` decision yet (T-0138 adds the first), so no test exercises the cycling except through `choose` directly, and none pins it (searched `Repeats: true` in `pkg/engine/*_test.go`).

### Shape (lead decisions, binding)
- **`fixedOrderPicks`**: for a `Repeats` decision, every one of the `d.Min` picks is the first option in fixed order (`idx[0]`). Without `Repeats`, unchanged: the first `d.Min` in fixed order. Its doc comment says so and cites `SVY-R32`.
- **`fallbackPicks`, `DecideMode`**: for a `Repeats` decision, every pick is option 0, the first listed. Without `Repeats`, unchanged.
- `DecideTarget`'s doc comment (`choice.go:14`) adds that a repeated target choice times out to the first legal entity for every repetition (`SVY-R32`).
- Nothing else changes: `DecideHand`, `DecideOrder` and `DecideOptional` never repeat; `randomPicks` and `Valid` are unchanged.
- **Sims:** no engine code builds a `Repeats` decision, so `go tool task sim-compare -- autoloop/m2-wave-c` must pass (starter and stress sims byte-identical to the run branch).

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
In `pkg/engine/fallback_order_test.go` (beside the other fixed-order timeout tests), calling `s.g.choose` directly on a hand-built `Decision`, as `choice_test.go` does. No existing test is edited.
- `TestFR_MOD_001_RepeatedTargetTimeoutConcentrates` (`SVY-R32`): a `DecideTarget` with `Repeats`, `Min = Max = 3`, and two legal assets on the battlefield listed so that option order differs from fixed order (for example option 0 an opposing unit, option 1 an allied unit, with player 0 deciding). On a timeout all three picks are the allied unit's index. Run it with a scripted timeout, with an invalid answer, and with no chooser; all three give the same picks.
- `TestFR_MOD_001_RepeatedTargetTimeoutTakesOwnHeadquartersFirst` (`SVY-R16` order): a `DecideTarget` with `Repeats`, `Min = Max = 2`, and three legal assets (an opposing unit, an allied Main Theater unit, the deciding player's Headquarters, listed in that order). On a timeout both picks are the Headquarters' index. (One option is the fast path, already pinned, so it is not tested here.)
- `TestFR_MOD_001_RepeatedModeTimeoutTakesFirstListed` (directive #91 (1)): a `DecideMode` with `Repeats`, two options, `Min = Max = 3`; a scripted timeout gives `[0 0 0]`.
- `TestFR_MOD_001_NonRepeatedFallbackUnchanged`: guard for the unchanged path. A `DecideTarget` without `Repeats`, `Min = Max = 2`, three legal assets in an option order that differs from fixed order: a timeout gives the first two in fixed order, distinct. A `DecideMode` without `Repeats`, `Min = Max = 2`, three options: a timeout gives `[0 1]`.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected.

## Allowed Files
- `pkg/engine/choice.go` (`fixedOrderPicks`, `fallbackPicks`, the `DecideTarget` doc comment)

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run 'TestFR_MOD_001_Repeated|TestFR_MOD_001_NonRepeatedFallbackUnchanged'`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim-compare -- autoloop/m2-wave-c` passes (no engine code builds a `Repeats` decision yet)

## Log
- 2026-11-02 lead-architect (wave C session 27): written against the run branch after syncing `main` (51c50df, 2dbd536: D7.1 marks a directive as asked on `main`). Directive #91's trigger fired (its outcome reads `Status: applied: FRD 01 v3.20`), so #91 is open; its part (1) is this task. #20 (b) split under the D2 size rule, because one task would touch six non-test files (`ability.go`, `targeting.go`, `placement.go`, `zones.go`, `actions.go`, `choice.go`): this task (`choice.go`), T-0137 (queued: a discard cost's cards are chosen before the targets and moved after the push, `zones.go`, `placement.go`, `actions.go`) and T-0138 (queued: `ability.Target.PerDiscard`, `ability.go`, `targeting.go`, `placement.go`), then the *Steel Rain* card task, which #20 (b) keeps. Status `tests`.
