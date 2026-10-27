---
id: T-0076
title: An untargeted pick is its own decision kind, DecidePick, not a Target (SVY-R3, SVY-R14, directive #20 (a), wave B shape, sonnet)
kind: shape
tier: sonnet
wave: B
status: review
branch: m2/T-0076-pick-decision
depends_on: [T-0072]
claims: [decision:pick]
cards: []
requirements: [FR-TGT-001, FR-MOD-001]
---

## Goal
The decision an untargeted pick asks (`ability.SubjectPick`, T-0072) says it is a pick, not a target: `Decision.Kind` is a new `DecidePick`, and its `decision` event records `"Pick"`. Today `pickAssets` asks a `DecideTarget`, so the sim's targeting monitor (`pkg/sim/monitor_targeting.go`) holds the pick to the target rules: a Target decision must be followed by the placement of its item, and it must not offer an opponent's Ward asset. A pick is made while its item resolves and ignores Ward (`SVY-R3`, `SVY-R14`), so the first card with a pick, Infested Apostle (T-0073), fails the starter sim (wave B run row 10: game 3, seed 3, event 377).

## Inputs
- FRD 01:
  - `FR-TGT-001` "Untargeted Selections (`SVY-R3`)" (line 329): no "chosen" means no target. The pick is made as the clause resolves; Ward and `FR-TGT-003` don't apply.
  - `FR-TGT-001` "You May X. If You Do, Y (`SVY-R14`)" (line 330): X is an untargeted selection on resolution.
  - `FR-MOD-001` §6 (line 359): mandatory targets and untargeted selections time out to the same fallback, the first legal options in fixed order; the selection for X is optional and selects none.
- Rulings: `SVY-R3`, `SVY-R14`, `SVY-R16`.
- Board directive #20 (a): "an untargeted pick ... is a new resolution-time choice through the `Chooser` ... It is not a `Target`, so Ward and two-stage verification skip it." T-0072 applied the pick but reused `DecideTarget` for it (its Log, lead entry), which made the decision look like a target to every reader of the event stream.
- Design note: `docs/architecture/m2-ability-model.md`, Key types (untargeted pick).
- Kernel today:
  - `DecisionKind` (`choice.go`) says what a decision asks and so which §6 fallback applies. `fallbackPicks` gives `DecideTarget` `fixedOrderPicks`, which picks none at `Min` 0.
  - `assetDecision` (`targeting.go`) builds a `DecideTarget` over battlefield assets, oldest first. Both `chooseTargets` and `pickAssets` (`effects.go`) use it.
  - `emitDecision` (`events.go`) records `d.Kind.String()`; `pkg/sim/monitor_targeting.go` checks only decisions whose kind is `"Target"`, and `pkg/eventlog/replay.go` compares kinds as strings. Neither needs a change once picks have their own kind.

### Shape (lead decisions, binding)
- **`DecidePick`**, appended after `DecideHand` in `choice.go`, with the comment `// untargeted pick from the battlefield as a step resolves (SVY-R3, SVY-R14): not a target, so Ward and the targeting checks skip it; same fallback as DecideTarget`. Its `String()` is `"Pick"`.
- `fallbackPicks` gives `DecidePick` the same `fixedOrderPicks` as `DecideTarget`, in one `case` (no second copy of the fallback).
- `assetDecision` takes the kind from its caller: `chooseTargets` passes `DecideTarget`, `pickAssets` passes `DecidePick`. Update the comments on `assetDecision` and `pickAssets` ("It uses DecideTarget" no longer holds) and the `doc.go` line on `SubjectPick`.
- No other behavior changes: options, order, `Min`/`Max`, the single-option fast path and the timeouts stay as T-0072 made them. Targets stay `DecideTarget`.
- Nothing changes in `pkg/sim`, `pkg/eventlog`, `pkg/view` or `cmd/cardtown`. `cardtown play`'s terminal chooser handles any kind other than Optional and Order generically, and the targeting monitor skips a kind it doesn't check.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
To be written by `rules-test-author` when this task reaches `tests`. New tests go in `pkg/engine/pick_decision_test.go`:
- `TestFR_TGT_001_UntargetedPickAsksAPickDecision`: an effect with a chosen target and a `SubjectPick` step asks `DecideTarget` on placement and `DecidePick` on resolution; the pick's `decision` event in `g.Log` records kind `"Pick"`, the target's records `"Target"`.
- `TestFR_MOD_001_PickDecisionKindIsNamedPick`: `DecidePick.String()` is `"Pick"`, and the existing kinds keep their names.
- A timeout at a `DecidePick` with `Min` 1 takes the first option in fixed order, and with `Min` 0 takes none, through `g.choose` directly (the same fallback as `DecideTarget`, `FR-MOD-001` §6). Name it in the `TestFR_MOD_001_` family.

Authorized edits to existing tests (they land in the `test(T-0076)` commit):
- `pkg/engine/if_you_do_test.go`, `TestFR_TGT_001_UntargetedPickIsMadeOnResolutionFromWhatMatchesThen`: `want := []DecisionKind{DecideOptional, DecideTarget, DecideTarget}` becomes `{DecideOptional, DecideTarget, DecidePick}`, and its failure message keeps saying the pick is asked later.
- `pkg/engine/optional_upto_test.go`, `scriptPlan`: `case DecideTarget:` becomes `case DecideTarget, DecidePick:`, so targets and picks share the scripted `tgts` queue in the order they are asked. Update the comments on `tpick` and `scriptPlan` to say so. Every T-0072 test must keep passing with no other change.

## Helpers Needed
None expected.

## Allowed Files
- `pkg/engine/choice.go` (`DecidePick`, its name, its fallback)
- `pkg/engine/targeting.go` (`assetDecision` takes the kind)
- `pkg/engine/effects.go` (`pickAssets` passes `DecidePick`; its comment)
- `pkg/engine/doc.go` (the `SubjectPick` line)

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run 'TestFR_TGT_001_Untargeted|TestFR_MOD_001_|TestFR_TRG_005_IfYouDo'`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-compare -- autoloop/m2-wave-b` passes: no starter or stress card has a pick yet, so sim output is byte-identical

## Log
- 2026-10-27 lead-architect (wave run B, session 11): Wrote the task (D6, missing shape for T-0073). Session 10's Flash run for T-0073 found that the Apostle's pick trips the targeting monitor, because T-0072 reused `DecideTarget` for picks.
  - The fix is in the decision, not the monitor: a pick is a different question from a target (directive #20 (a)), and anything reading the event stream (the monitor, replay, a future client) should be able to tell them apart without knowing where in a resolution it is. Teaching the monitor to guess (for example "a Target decision during a resolution is a pick") would write the SVY-R3 rule into a second place.
  - Reuse: `fixedOrderPicks` and `assetDecision` are shared, not copied. The new kind adds one name and one `case` label.
  - Directives checked against `autoloop/m2-wave-b` 8d7ac9a (after syncing `main`, e76423d): no `open: before` row. Waiting triggers not fired: #6 and #7 (mode `serial`), #8 (no `deck.go` change), #15 (no turn hook), #20 (b) (no wave C shape), #58 (no exhaust by a cost or an effect), #59 (this pick is from the battlefield, not a pile; no repeated targets).
  - Claims overlap no open task (T-0073, the only one not `done`, claims only its card and goes back to `queued`).
- 2026-10-27 rules-test-author: Wrote the tests (commit 653df34), status tests-review. New in `pkg/engine/pick_decision_test.go`: `TestFR_TGT_001_UntargetedPickAsksAPickDecision`, `TestFR_MOD_001_PickDecisionKindIsNamedPick`, `TestFR_MOD_001_PickDecisionTimeoutMinOneTakesFirstOption`, `TestFR_MOD_001_PickDecisionTimeoutMinZeroTakesNone`. Authorized edits: `if_you_do_test.go` want is `{DecideOptional, DecideTarget, DecidePick}`; `scriptPlan` handles `case DecideTarget, DecidePick:` and its comments say so. Helpers requested: none. Left out: none. The package fails to compile only on `DecidePick`. `cardtown lint rules` passes.
- 2026-10-27 lead-architect: Tests approved (653df34), `ready`. The new tests pin the kind on the decision and on its event, the name, and both §6 timeouts at a pick; the two authorized edits are exactly as named. `TestFR_MOD_001_PickDecisionTimeoutMinOneTakesFirstOption` uses option refs that aren't assets, so it pins the shared fallback, not the battlefield order; `TestFR_MOD_001_SinglePickTimeoutPicksFirstInFixedOrder` (T-0072) still pins that order for a real pick. Dispatching `kernel-engineer` (sonnet).
- 2026-10-27 kernel-engineer: BLOCKED (a test contradicts its own setup). Implemented the Shape as written: `DecidePick` after `DecideHand` (name "Pick"), `DecideTarget, DecidePick` share one `case` in `fallbackPicks`, `assetDecision` takes a `kind` (chooseTargets passes `DecideTarget`, pickAssets `DecidePick`), comments and the `doc.go` line updated. Reuse: `fixedOrderPicks` and `assetDecision` shared, nothing copied. Results: `go test ./pkg/engine` fails only `TestFR_TGT_001_UntargetedPickAsksAPickDecision` (line 36): the decision events are `[Target Target Pick]`, the test wants `[Target Pick]`. Cause: in that test the enemy target (`iydEnemy`) has one option (Victim), so `choose` takes the single-option fast path (FR-MOD-001 §1) and `emitDecision` still logs it as a `decision` event with how `auto`. The test counts only asked decisions in `s.asked(0)` (which is `[Target Pick]`, passing) but all `decision` events for player 0 in the log. The task says the single-option fast path stays as T-0072 made it, so I did not change it. Fix needed in the test (for example give the enemy two options, or skip `How == "auto"` events, or want `[Target Target Pick]`); the kernel cannot satisfy it without dropping the auto decision event. `go tool task check` stops at that test; `go tool task sim` clean; `go tool task sim-compare -- autoloop/m2-wave-b` PASS (byte-identical). The other listed tests pass.
- 2026-10-27 lead-architect: Not a block; the test is wrong, and my test review missed it. Back to `tests`, with the implementation (3a350a4) left on the branch. Notes for `rules-test-author`:
  1. `TestFR_TGT_001_UntargetedPickAsksAPickDecision`: the scene has one opposing unit, so the chosen target is answered by the single-option fast path (`FR-MOD-001` §1). The kernel asks no one but still logs it as a `decision` event with how `auto`, so the log reads `[Target Target Pick]` while `s.asked(0)` reads `[Target Pick]`. Put a second opposing unit in the scene and answer its target question with a scripted `pick(...)` at the front of the `tgts` queue, so all three decisions (enemy target, ally target, pick) are asked and logged. Then want `[DecideTarget, DecideTarget, DecidePick]` asked and `["Target", "Target", "Pick"]` logged. Don't filter on `How`: the point is that every decision event names its kind.
  2. Change nothing else. The fix goes in a new `test(T-0076)` commit (a rework of this task's own test).
- 2026-10-27 rules-test-author: Applied review note 1 (commit 2bc7003). `TestFR_TGT_001_UntargetedPickAsksAPickDecision` now has a second opposing unit, scripts its enemy-target pick first in the tgts queue, and wants asked `[DecideTarget, DecideTarget, DecidePick]` and logged `["Target", "Target", "Pick"]` with no filter on How. Nothing else changed. `go test ./pkg/engine -run 'TestFR_TGT_001_Untargeted|TestFR_MOD_001_|TestFR_TRG_005_IfYouDo'` passes on the branch with 3a350a4. Status tests-review.
- 2026-10-27 lead-architect: Test rework approved (2bc7003): two opposing units, the enemy target scripted first, three asked and three logged decisions, as note 1 asked. `ready`. The implementation (3a350a4) needs no change for it; `kernel-engineer` re-runs Done When and hands off to `review`.
- 2026-10-27 kernel-engineer: Re-run after the test rework (2bc7003). Implementation (3a350a4) unchanged. Results: the listed `go test ./pkg/engine -run ...` ok; `go tool task check` exit 0; `go tool task sim` clean; `go tool task sim-compare -- autoloop/m2-wave-b` PASS (byte-identical). Status review.
