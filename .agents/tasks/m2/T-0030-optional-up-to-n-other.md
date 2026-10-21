---
id: T-0030
title: Optional ("you may") triggers, "up to N" chosen targets, and "other"
kind: shape
tier: opus
wave: A2 (task 7, shape)
status: done
branch: m2/T-0030-optional-up-to-n-other
depends_on: [T-0028]
claims: [field:Trigger.Optional, field:Target.UpTo, field:Filter.Other, field:StackItem.Chosen-multi]
cards: []
requirements: [FR-TGT-001, FR-TGT-002, FR-TGT-003, FR-MOD-001]
---

## Goal
Triggers can now be optional ("you may"), a targeted clause can choose "up to N" distinct entities, and a filter can exclude the ability's own source ("other"). These are the last wave A2 shapes. With them and T-0028's shapes, *Craterfall Ancient* (Animus, 2 copies in the starter deck) becomes buildable:
- "Whenever this unit attacks, you may deal damage equal to its power to each of up to 2 other chosen assets."
- "When this unit is killed, deal 4 damage to all assets."

Its card task comes next (flash). This task has no cards of its own, and the tests use synthetic cards.

## Inputs
- Survey rulings (`docs/design/m2_starter_set_ability_survey.md`, ruled by the user on 2026-10-21; not yet in FRD 01):
  - `SVY-R9`: "you may" is answered when the trigger is placed on the stack, before its targets. A declined trigger chooses no targets and does nothing, and a timeout declines.
  - `SVY-R11`: "up to N chosen" means distinct entities.
- FRD 01:
  - `FR-TGT-001`: conjunction filters; targets are locked on placement (`SVY-R3`).
  - `FR-TGT-002`: quantifiers, including "up to 2 chosen units".
  - `FR-TGT-003`: two-stage verification, and partial invalidation. If some targets are illegal at resolution, the effect still resolves for the legal ones.
  - `FR-MOD-001` §1 (fast path) and §6 (an optional trigger declines; mandatory targets pick at random).
- **Choice machinery, already built and tested (T-0006, `choice.go`, `TestFR_MOD_001_*`):**
  - `DecideOptional` is a decision with one option, `Min 0`, `Max 1`. It is always asked, and its fallback declines.
  - A `DecideTarget` with `Min < Max` is never fast-pathed, and its fallback picks `Min` random legal options. So an "up to N" timeout picks none.
  - Picks are distinct unless `Repeats` is set.
  - Reuse all of this. Don't add a new decision kind.
- **Kernel today:**
  - `chooseTargets` (`trigger.go`) asks one `DecideTarget` (`Min 1, Max 1`) per `Effect.Targets` slot when a trigger is placed.
  - `StackItem.Chosen []int` holds one asset ID per slot.
  - `resolveAbility` re-verifies each slot, and `stepAssets` returns the chosen asset or, for `SubjectAll`, the matches.
  - `matchesFilter` and `legalTarget` take the controller, the filter and the asset, but not the source.
- Rulings that apply: `SVY-R3`, `SVY-R9` and `SVY-R11`.

### Shapes
```go
// pkg/engine/ability
// Trigger gains:
	Optional bool // "you may": asked when the trigger is placed, before its targets (SVY-R9)

// Target gains:
	UpTo int // 0: exactly one chosen entity (today's behavior). N > 0: up to N distinct entities, possibly none (SVY-R11)

// Filter gains:
	Other bool // excludes the ability's own source asset ("other")
```
`StackItem.Chosen` changes so that a slot can hold several asset IDs, in pick order. The engineer picks the representation (for example `[][]int`).

### Lead decisions (technical)
- **Optional triggers:**
  - When the flush is about to place an optional trigger, in its turn in the placement order (after any Order decision), its controller is asked one `DecideOptional`. Its prompt names the trigger, and its single option is "yes".
  - **Yes:** its targets are chosen as usual, and it is placed.
  - **No, a timeout, no chooser or an invalid answer** (T-0006's fallback): it isn't placed at all. No targets are asked, and no `place` event is emitted. With no responses in this game, that is the same as placing a trigger that does nothing.
  - The optional question comes even if the trigger has no targets.
- **"Up to N" (`UpTo > 0`):**
  - One `DecideTarget` per slot, with `Min 0`, `Max min(N, number of legal options)`, offering the same legal options as today, oldest first.
  - With no legal option, nothing is asked and no decision is recorded. The slot stays empty.
  - With one or more options, it is always asked (`Min != Max`, so no fast path, per T-0006).
  - A timeout picks none (T-0006's `Min` fallback).
  - Picks are distinct (`SVY-R11`). A repeated pick is an invalid answer and falls back.
- **Steps on a multi-entity slot:**
  - A `SubjectTarget` step acts on each entity chosen for its slot, in pick order.
  - Each entity is re-verified at resolution (`FR-TGT-003`). The step skips illegal ones and still acts on the legal ones (partial invalidation).
  - Each entity gets its own damage instance, all dealt before the item's death sweep.
  - The step's `Value` is read once per step, as in T-0028.
- **`Filter.Other`:** excludes the source asset (`pendingTrigger.source`, then `StackItem.Source`) in `matchesFilter`, which takes the source as an extra argument. That way it applies to chosen targets at placement, to re-verification and to `SubjectAll`.
- **Unchanged:** a slot with `UpTo == 0` behaves exactly as today: one mandatory target, with the fast path and the random fallback. Every existing test passes unchanged.
- **The `targeting` rule monitor** needs no change. An "up to N" `Target` decision is still followed by its item's `place` event, and a declined optional trigger has no `Target` decision.
- **Fences:** no card names.
- **Not in this task:**
  - the Craterfall Ancient definition (its card task is next);
  - optional non-resource *costs* on triggers (`FR-TRG-005`, Arkha, wave B);
  - "up to N" with distributed values (`FR-TGT-002` pools);
  - untargeted picks (directive #20, wave C).

### What the tests can observe
All with synthetic cards in a new `pkg/engine/optional_upto_test.go`.
- **Optional:**
  - The controller is asked a `DecideOptional` (always, even when only one answer would matter). Yes leads to the trigger's target question and the effect.
  - No leads to no target question, no effect, no `place` event for it, and an empty stack.
  - A timeout and a nil chooser both decline.
  - The optional question comes before the target question.
  - With two triggers of the same player, one of them optional, the Order decision comes first. The optional question is asked when that trigger's turn to be placed comes.
  - An optional trigger with no targets ("you may draw a card") works both ways.
- **Up to N:**
  - Picking 0, 1 or 2 distinct legal entities is valid. Each chosen entity gets its own damage instance from one step, and with 0 picked nothing happens.
  - A repeated pick is invalid. It is logged and falls back to none.
  - With exactly one legal option, the player is still asked (not fast-pathed).
  - With no legal option, nothing is asked and nothing happens.
  - A timeout picks none.
  - Options still exclude opposing Ward assets.
- **Partial invalidation (`FR-TGT-003`):** two entities are chosen, and one is destroyed by an item resolving above this one before it resolves. The surviving one still takes the damage, and the destroyed one takes none.
- **Other:**
  - "other chosen asset" never offers the source itself, while the same filter without `Other` does (today's convention).
  - `SubjectAll` with `Other` skips the source.
- **Composite (the Craterfall shape):** a unit with power 4, with an attack trigger that is optional, has `UpTo 2` targets `{Other: true}`, and deals `ValueSourcePower` damage.
  - Yes, picking two assets, deals 4 to each.
  - Yes, picking one asset, deals 4 to it.
  - No deals nothing.
  - Its own HQ is a legal option (any asset), and the unit itself is not.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
All in `pkg/engine/optional_upto_test.go`, synthetic cards only. Every scene has "Hero" (player 0) attack "Dummy" with attack triggers, and counts non-combat damage events and `place` events with Hero as Source (the attack itself is one). Targets are chosen by asset.
- Optional (SVY-R9, FR-MOD-001):
  - `TestFR_MOD_001_OptionalTriggerYesAsksOptionalThenTargetsThenResolves`
  - `TestFR_MOD_001_OptionalTriggerNoIsNotPlacedAndAsksNoTargets`
  - `TestFR_MOD_001_OptionalTriggerTimeoutNoChooserAndInvalidAnswerDecline`
  - `TestFR_MOD_001_OptionalTriggerWithNoTargetsIsStillAskedBothWays`
  - `TestFR_MOD_001_OptionalQuestionComesAfterTheOrderDecisionInPlacementOrder`
  - `TestFR_MOD_001_OptionalAndMandatoryTriggersOfOnePlayerShareAnOrderDecision`
- Up to N (SVY-R11, FR-TGT-002, FR-MOD-001, FR-KWD-002h):
  - `TestFR_TGT_002_UpToTwoAcceptsZeroOneOrTwoDistinctTargets`
  - `TestFR_TGT_002_UpToWithOneLegalOptionIsStillAsked`
  - `TestFR_TGT_002_UpToOptionsAreOldestFirst`
  - `TestFR_TGT_002_UpToRepeatedPickIsInvalidAndFallsBackToNone`
  - `TestFR_TGT_002_UpToNoLegalOptionAsksNothingAndDoesNothing`
  - `TestFR_MOD_001_UpToTimeoutAndNoChooserPickNone`
  - `TestFR_KWD_002h_UpToOptionsExcludeOpposingWardAssets`
- Partial invalidation (FR-TGT-003): `TestFR_TGT_003_UpToResolvesForTheSurvivingChosenEntity`
- Other (FR-TGT-001, FR-TGT-002):
  - `TestFR_TGT_001_OtherExcludesTheSourceFromChosenOptionsAndOnlyThat`
  - `TestFR_TGT_002_AllWithOtherSkipsTheSourceAndHitsTheRest`
- Composite (the Craterfall shape):
  - `TestFR_TGT_002_CompositeYesWithTwoAssetsDealsPowerToEach`
  - `TestFR_TGT_002_CompositeYesWithOneAssetDealsPowerToIt`
  - `TestFR_TGT_002_CompositeNoDealsNothing`

Run regex for Done When: `go test ./pkg/engine -run 'TestFR_(MOD_001_(Optional|UpTo)|TGT_002_(UpTo|AllWithOther|Composite)|TGT_003_UpTo|TGT_001_Other|KWD_002h_UpTo)'`

Omissions:
- The exact prompt wording of the optional question: only "non-empty, one option, Min 0, Max 1" is asserted.
- Re-verifying `Other` at resolution: the source cannot change, so it is not observable.
- `Other` on a death trigger's `SubjectAll`: the dead source is already off the battlefield, so the filter is the same as without it.
- Optional triggers of the non-active player (APNAP): T-0030 does not list it, and the same flush path covers it.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None. The tests define `arena`, `scriptPlan`, `pick` and the other helpers locally, and reuse existing test helpers (`striker`, `tgt`, `dmgTo`, `powerStep`, `allDamage`, `damageTo`, `nonCombatDamage`, `invalidLogs`, `eventsOfKind`, `noFormatErrors`).

## Allowed Files
- `pkg/engine/ability/ability.go`
- `pkg/engine/trigger.go`
- `pkg/engine/stack.go` (the `StackItem.Chosen` field and its `place` event handling, if needed)

## Done When
- [x] The tests pass: `go test ./pkg/engine -run '<regex from rules-test-author>'`
- [x] Every existing test passes unchanged
- [x] `go tool task check` passes
- [x] `go run ./cmd/cardtown sim -deck starter -p0 ANIMUS -p1 PATH` and `go run ./cmd/cardtown sim` exit 0, with every invariant holding and no rule-monitor violation, and with output byte-identical to `rework/executable-rules` (no card uses the new shapes yet)

## Log
- 2026-10-21 lead-architect: Wrote the task (A2 shape, last of the wave), with status `tests` for `rules-test-author`.
  - Tier `opus` (board directive #25).
  - It cites `SVY-R9` and `SVY-R11`.
  - No new rules question: the timeout and fast-path behavior for "up to N" and "you may" follows T-0006's tested choice rules.
- 2026-10-21 rules-test-author: Wrote `pkg/engine/optional_upto_test.go` (16 tests, listed under Tests), status `tests-review`.
  - Expected failure: the file does not compile, because `Trigger.Optional`, `Target.UpTo` and `Filter.Other` do not exist. With the file moved aside, every other engine test passes.
  - Helpers requested: none. `cardtown lint rules` passes.
  - Omissions: see the Tests section.
  - No open questions.
- 2026-10-21 lead-architect: Test review: **approved**, status `ready`.
  - 672bbf1 adds `optional_upto_test.go` with 19 tests and changes no existing test.
  - Rules content checks out against FRD 01, `SVY-R9`, `SVY-R11`, T-0006's choice rules and the lead decisions. The omissions are accepted.
  - **Reviewer prototype** (a throwaway worktree, not committed): `Optional` asked via `DecideOptional` before targets, with a decline skipping placement; `UpTo` as `Min 0`, `Max min(N, options)`, with no question when there are no options; `Chosen [][]int` re-verified per entity; `Other` in `matchesFilter`. With it, all of `go test ./pkg/engine` passes.
  - **Mutations each fail tests:**
    - optional ignored: 8;
    - a declined trigger still placed: 5;
    - `UpTo` ignored: 7;
    - `Other` ignored: 3;
    - asking with no options: 1;
    - a whole slot dropped on one invalid target: 1;
    - repeats allowed: 1.
  - Next: `kernel-engineer` on **Opus** (directive #25).
- 2026-10-21 kernel-engineer (Opus): Implemented, status `review`. The diff touches only the three Allowed Files. No test file and no `scenario_test.go` helper changed.
  - **`ability.go`:** adds `Trigger.Optional`, `Target.UpTo` and `Filter.Other`, as in the Shapes block. The `Target` and `SubjectTarget` comments now describe "up to N" and pick order.
  - **`stack.go`:** `StackItem.Chosen` is now `[][]int`: the asset IDs for each slot, in pick order, empty if none. `push` needed no change.
  - **`trigger.go`, optional triggers:**
    - In `flushTriggers`, inside the placement loop (after any Order decision), an optional trigger first calls the new `accepts(p)`.
    - `accepts` asks one `DecideOptional` through `g.choose`: prompt `"<label>: you may use it"`, one option "yes" (`Ref` is the source), `Min 0`, `Max 1`.
    - Unless exactly one pick comes back, the loop `continue`s, so no targets are asked and there is no `place` event. Keyword items (`p.item`) are never optional.
  - **`trigger.go`, "up to N":**
    - `chooseTargets` returns `[][]int`. It builds the same oldest-first option list as before.
    - When `UpTo > 0`, it skips the slot if there are no options (no `choose` call, so no decision is recorded). Otherwise it sets `Min 0`, `Max min(UpTo, options)` and the prompt `"<label>: choose up to N for target i"`.
    - `UpTo == 0` keeps today's `Min 1`, `Max 1` decision. Its prompt is unchanged, and so are the fast path, the fallback and the RNG use.
  - **`trigger.go`, resolution:**
    - `resolveAbility` re-verifies every chosen ID on its own with `legalTarget`, and keeps the legal ones in pick order (`[][]*Asset`, FR-TGT-003 partial invalidation).
    - `stepAssets` returns the slot's surviving list for `SubjectTarget`.
    - The deal-damage loop is unchanged: `Value` is read once per step, there is one `dealDamage` per entity, and every instance is dealt before the item's sweep.
  - **`trigger.go`, "other":** `matchesFilter(ctl, src, f, a)` and `legalTarget(ctl, src, f, a)` take the source asset ID, and `f.Other && a.ID == src` excludes it. Callers pass `p.source` at placement, and `it.Source` at re-verification and for `SubjectAll`.
  - **Reused:** T-0006's `choose`, `DecideOptional` and `DecideTarget`. Its validation already rejects repeats and out-of-range picks, and its fallback gives none for `Min 0` and declines for Optional. No new decision kind or helper. `matchesFilter` and `legalTarget` were extended, not duplicated.
  - **Done When regex** `go test ./pkg/engine -run 'TestFR_(MOD_001_(Optional|UpTo)|TGT_002_(UpTo|AllWithOther|Composite)|TGT_003_UpTo|TGT_001_Other|KWD_002h_UpTo)' -v`: ok. All 19 T-0030 tests pass, plus 4 existing T-0006 `Optional*` tests that the regex also matches.
  - `go test ./...`: ok.
  - `go tool task check`: exit 0 (fmt-check, vet, test, build, `lint --all` all PASS, and fences PASS).
  - **Sim byte identity:**
    - **Baseline:** a CLI built in a separate `git worktree` at `rework/executable-rules` (01512ab) before any edit. The worktree was removed afterwards.
    - **Stdout:** `go run ./cmd/cardtown sim -deck starter -p0 ANIMUS -p1 PATH` and `go run ./cmd/cardtown sim` both exit 0. Their stdout is byte-identical to the baseline (`cmp`), with every invariant holding and the monitors finding no violation. `-policy random` stdout is also byte-identical, for starter and for stress.
    - **Events:** every game's events were also compared through `-events`: starter and stress, progress and random, 200 games each. The JSONL is identical once the match-start `revision` and `modified` build stamp is normalized. That stamp is the only difference.
  - **Reviewer, look closely at:**
    1. **"Up to N" with zero options:** this case returns before `choose` on purpose. `choose` would otherwise take the sole-answer fast path (`Min == Max == 0 == len(options)`) and record an `auto` decision, and the tests forbid any record.
    2. **The Optional prompt and the "up to" prompt text:** both are new wording. Tests assert only that a prompt is non-empty.
    3. **`RandomChooser` on an optional trigger** says yes or no at random, and both sim policies answer every decision with it (`sim.go`). No card uses `Optional` yet, so that sim behavior first shows up in the Craterfall Ancient card task.
- 2026-10-21 lead-architect: Code review: **approved**, status `done`.
  - Scope: e6e1d55 touches only the three Allowed Files. No `_test.go` change after 672bbf1.
  - Code:
    - It matches the reviewer's test-review prototype: `accepts` for `DecideOptional` before targets, with a decline skipping placement; `UpTo` with `Min 0`, `Max min(N, options)`, and no question with zero options; `Chosen [][]int` re-verified per entity; `Other` in `matchesFilter`, with the source passed at placement, re-verification and `SubjectAll`.
    - `UpTo == 0` keeps the old decision exactly (prompt, fast path, fallback, RNG).
    - No new decision kind, and no card names.
  - Engineer notes accepted:
    - Note 1: returning before `choose` with zero options is correct. Otherwise the Min == Max == 0 fast path would record an `auto` decision.
    - Note 3: the sim answers optional triggers at random. That is fine for coverage, and the Craterfall card task will show it.
  - Gates re-run by the reviewer:
    - `go test ./...` passes, and `go tool task check` exits 0.
    - For starter and stress games under both policies, `cardtown sim` is byte-identical to a baseline CLI built from `rework/executable-rules`.
  - Wave A2's shapes are complete.
