---
id: T-0131
title: An additional cost can discard up to N cards from the payer's hand (FR-TRN-003a, SVY-R3, directive #20 (b), wave C shape, sonnet)
kind: shape
tier: sonnet
wave: C
status: done
branch: m2/autoloop-wave-c/T-0131-discard-cost
depends_on: [T-0128, T-0129]
claims: [const:CostDiscard, field:Cost.UpTo, func:costsPayable, func:payCosts, func:discardFromHand]
cards: []
requirements: [FR-TRN-003a, FR-TRN-003, FR-MNV-001, FR-MOD-001, FR-TGT-001]
---

## Goal
A Maneuver can print *"As an additional cost to enact this maneuver, discard up to 3 cards."* (*Steel Rain*, a deck card). Its controller chooses up to that many cards from their own hand as it is declared, possibly none, and they are discarded then, after the Maneuver is on the stack. The cost never stops the Maneuver from being enacted, because discarding none pays it. Split from *Steel Rain* under the D2 size rule: the hand-empty condition is T-0132 (queued). One target choice per discarded card, and then the card, stay on directive #20 (b): that shape's `claims:` include repeated targets, so it is written only after #59 (2)'s ruling.

## Inputs
- FRD 01 (run branch):
  - `FR-TRN-003a`, "Additional Sacrifice, Discard & Exhaust Costs" (line 335): a cost of discarding a card (*Steel Rain*) is paid immediately on declaration.
  - `FR-TRN-003`, `FR-MNV-001` (line 333): on declaration the card moves from Hand into the Resolution Stack and the player then pays all costs upfront, so the Maneuver itself is never a card it can discard.
  - `FR-TGT-001`, Targets Are Locked on Placement (`SVY-R3`, lines 354-356): *Steel Rain*'s discard is an additional cost so that its outcome is known on declaration. The targets per discarded card are not this task (#20 (b), #59 (2)).
  - `FR-MOD-001` §1 (line 380): choosing from your own hand is not an inspection, and a required choice with one legal option is taken unasked. §6 (line 389): "An *"up to N"* selection selects none." §6 (line 390): a choice from your own hand that times out takes the card most recently added.
- Rulings that apply: `SVY-R2`, `SVY-R3`, `SVY-R16`, `SVY-R29`.
- Directive #20 (b) (`BOARD.md`): *Steel Rain*'s discard is an additional cost, and the number of target choices comes from it. This task builds the cost; the count waits on #59 (2)'s ruling.
- Survey: wave C row (`docs/design/m2_starter_set_ability_survey.md` line 31) and the *Steel Rain* row (line 176): "discard up to 3".
- Design note (`docs/architecture/m2-ability-model.md`): Boundaries line 34, the `Cost` shape ("... discard N, ...").
- Kernel today (run branch at ca50083):
  - `ability.CostKind` with `CostKill` and `CostExhaust` (`ability/ability.go:33-39`), and `ability.Cost{Kind, Filter}` (lines 41-47), whose doc comment says `Filter` is a battlefield filter.
  - `placement.go`: `costDecision` (line 49) builds the asset pick for a kill or exhaust cost; `costsPayable` (line 65) needs at least one option for every cost; `payCosts` (line 82) pays each cost in printed order by kind and panics on an unknown kind.
  - `Apply`'s `ActEnact` case (`actions.go:83-90`): the card leaves hand (`takeCard`), its targets are chosen, it is pushed, then `g.payCosts(pl)` runs.
  - `discardFromHand` (`zones.go:52-82`): the one function that discards from hand. It discards exactly `n` cards (the whole hand if it holds fewer) as a `DecideHand` with `Min = Max = n`, logs an `EventDiscard` per card with `Zone: ZoneHand.String()` (`"Hand"`, pinned by `discard_test.go:179`), and returns the cards. Its only caller is `stepDiscard` (`effects.go:201-209`).
  - The `DecideHand` timeout (`handPicks`, `choice.go`) takes the `d.Min` most recently added cards, so with `Min` 0 it takes none.

### Shape (lead decisions, binding)
- **`ability.CostDiscard`**, the third `CostKind` member: "discard up to `Cost.UpTo` cards of the payer's choice from their own hand, possibly none (`FR-TRN-003a`)".
- **`ability.Cost.UpTo int`**, a new field, read only by `CostDiscard`: the most cards it discards (`UpTo > 0`). `Cost`'s doc comment says that `Filter` is read by the kill and exhaust costs and `UpTo` by the discard cost. No card discards an exact number as a cost, so there is no exactly-N form.
- **`costsPayable`**: a `CostDiscard` is always payable, because discarding none pays it, so it adds no condition. Kill and exhaust costs still need an option from `costDecision`. `costDecision` stays asset-only: it is never called for a `CostDiscard`.
- **`payCosts`**: a `CostDiscard` is paid through `discardFromHand`, in printed order with the other costs and at the same moment (after the push, so the Maneuver is not in hand). No second discard loop anywhere.
- **`discardFromHand`** gains the "up to" form and stays the one function that discards from hand: `discardFromHand(pid PlayerID, n int, upTo bool) []*Instance`. With `upTo`, the decision is a `DecideHand` with `Min` 0 and `Max` `min(n, len(hand))`, its prompt reads "Discard up to N cards" ("Discard up to 1 card" for one), it is asked whenever the hand isn't empty (two or more valid answers, so no fast path), and a timeout discards none (§6 line 389, through `handPicks` with `Min` 0). With an empty hand nothing is asked. Without `upTo`, behavior is exactly as today; `stepDiscard` passes `false`. Its doc comment names both forms.
- **Out of scope:** the target choice per discarded card and recording how many were discarded (#20 (b), #59 (2)); "if you have no cards in hand" (T-0132); *Steel Rain* (#20 (b)); discard costs on triggers or modes (#61's fifth question).
- **Sims:** no card uses `CostDiscard` yet and `stepDiscard` is unchanged, so `go tool task sim-compare -- autoloop/m2-wave-c` must pass (starter and stress sims byte-identical to the run branch).

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
In a new file `pkg/engine/discard_cost_test.go`, on synthetic Maneuvers built with `maneuver_test.go`'s `maneuver`, `steps`, `enactAction`, `offersEnact` and `lastDiscard`. A `discardCostManeuver(name, req, tac, upTo, e)` helper (as `exhaustCostManeuver`) goes in the new test file. Each Maneuver's step logs an event (1 damage to the opposing Headquarters), not a draw, because a draw logs no event. No existing test is edited.
- `TestFR_TRN_003a_DiscardCostPaidOnDeclaration`: up to 3; the controller holds the Maneuver and four other cards, and a scripted chooser picks two of them. After `Apply`: those two are in the controller's discard pile, the other two are still in hand, the opposing Headquarters took 1 damage, the Maneuver is in its owner's discard pile, and its resource cost was spent. In the event log, both `discard` events (zone `"Hand"`, `ZoneHand.String()`) come after the Maneuver's `place` and before its damage.
- `TestFR_TRN_003a_DiscardCostMayDiscardNone`: up to 3, with three other cards in hand; the scripted chooser answers no picks. Nothing is discarded, all three cards stay in hand, and the step ran.
- `TestFR_TRN_003a_DiscardCostOffersTheRestOfTheHand` (`FR-TRN-003`): the decision is asked of the controller as a `DecideHand`; its options are exactly the other cards in hand, never the Maneuver; `Min` is 0. With two other cards and up to 3, `Max` is 2. With five other cards, `Max` is 3.
- `TestFR_TRN_003a_DiscardCostWithAnEmptyHand` (`SVY-R29`): the Maneuver is the controller's only card in hand. It is offered (`offersEnact`), `Apply` asks the controller no decision, and the step runs.
- `TestFR_MOD_001_DiscardCostTimeoutDiscardsNone` (§6, line 389): up to 3, with two other cards in hand; the decision times out. Nothing is discarded and the step runs.
- `TestFR_MOD_001_DiscardCostAskedWithOneCard` (§1): up to 3, with one other card in hand. The decision is still asked (none or that card are both valid answers), with `Max` 1.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected.

## Allowed Files
- `pkg/engine/ability/ability.go` (`CostDiscard`, `Cost.UpTo`, `Cost`'s doc comment)
- `pkg/engine/placement.go` (`costsPayable`, `payCosts`, their doc comments)
- `pkg/engine/zones.go` (`discardFromHand` only)
- `pkg/engine/effects.go` (`stepDiscard`'s call only)

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run 'TestFR_TRN_003a_DiscardCost|TestFR_MOD_001_DiscardCost'`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim-compare -- autoloop/m2-wave-c` passes (no card uses the discard cost yet)

## Log
- 2026-11-02 lead-architect (wave C session 19): written against the run branch at ca50083, with T-0132 (the hand-empty condition) queued (D2.4). Directive #20 (b) split: the cost here; the per-discard target choices and the card stay on #20 (b), behind #59 (2)'s ruling, which must be asked before a task claiming repeated targets is written. Reading for the timeout: §6 line 389's "an up to N selection selects none" sets how many, so line 390's "most recently added" never applies with `Min` 0; `handPicks` already does this. Status `tests`.
- 2026-11-02 rules-test-author: wrote `pkg/engine/discard_cost_test.go` with the six listed tests (`TestFR_TRN_003a_DiscardCostPaidOnDeclaration`, `..._DiscardCostMayDiscardNone`, `..._DiscardCostOffersTheRestOfTheHand` (table of 2 and 5 other cards), `..._DiscardCostWithAnEmptyHand`, `TestFR_MOD_001_DiscardCostTimeoutDiscardsNone`, `..._DiscardCostAskedWithOneCard`) and local helpers `discardCostManeuver`, `hqPing`, `discardCostScene`, `pingCount`. Reuses existing `inDiscard`/`inHand`. No helpers requested for `scenario_test.go`. `go vet` fails only on `ability.CostDiscard` and `Cost.UpTo` (the task's own claims). `cardtown lint rules` passes. Nothing deliberately left out. Status `tests-review`.
- 2026-11-02 lead-architect: tests approved (0c6d949): the six listed tests check the cited FRD 01 text (FR-TRN-003a line 335, FR-TRN-003 line 333, FR-MOD-001 sec. 1 line 380 and sec. 6 line 389, SVY-R29); the order test watches the Maneuver's `place`, the hand `discard` events and logged damage, not a draw; the Maneuver is checked never to be an option. `go vet` fails only on the task's own claims. Status `ready`, kernel-engineer (sonnet).
- 2026-11-02 kernel-engineer: BLOCKED. Reuse search: `discardFromHand` (zones.go) extended with `upTo`; `payCosts`/`costsPayable` branch on `CostDiscard`; no second discard loop. Implemented and verified: all new tests pass except `TestFR_TRN_003a_DiscardCostPaidOnDeclaration`, and `sim-compare -- autoloop/m2-wave-c` passes. The failure is a test contradiction: `discard_cost_test.go:86` requires the discard event's `Zone` to be `"hand"` (the task text also says `Zone: "hand"`), but `discardFromHand` logs `ZoneHand.String()` = `"Hand"`, and the existing `discard_test.go:179` requires `"Hand"`. Both cannot pass without editing a protected test or changing `Zone` for `stepDiscard` too (outside the allowed edit and would change event logs). Needs the lead to fix `discard_cost_test.go:86` to `"Hand"` (or rule otherwise). Implementation is committed on the branch as-is; `go tool task check` fails only on that one assertion.
- 2026-11-02 lead-architect: wave C run, session 20. Session 19 hit a usage limit after the engineer's block, so it left no run-file row; its clean scratch worktrees (`20261102-120939-2/run`, `/task`) were removed. No `main` commits to sync (`main` d237625 is in the run branch, 2f3224d). Not `WAVE-COMPLETE`: 15 deck cards are on `NotYetExecutable`. Directives re-read on the run branch: no `open: before` directive; #20 (b) is this task; waiting triggers #6, #7, #8, #15, #59 (2), #61 (fifth question), #71 and #88 (distinctness) not fired by this task's `claims:`. Block reviewed; a lead spec defect, not a rules question. The task's Inputs and Tests said the discard event's zone is `hand`, but `discardFromHand` logs `ZoneHand.String()`, `"Hand"`, which `discard_test.go:179` pins; re-running `go test ./pkg/engine -run 'TestFR_.*Discard'` on b28b653 fails only on `discard_cost_test.go:87`. Inputs line and the first test's text corrected to `"Hand"`. Status `tests`, round 2 for `rules-test-author`. Notes:
  1. `TestFR_TRN_003a_DiscardCostPaidOnDeclaration` (`discard_cost_test.go:86`): compare the discard event's `Zone` to `"Hand"` (or `ZoneHand.String()`), not `"hand"`. Change nothing else in any test.
  2. For the engineer's next round (not the tests): b28b653 otherwise matches the Shape (`CostDiscard`, `Cost.UpTo`, `costsPayable` skipping it, `payCosts` paying it through the one `discardFromHand` after the push, `stepDiscard` passing `false`). Re-wrap `payCosts`' doc comment (`placement.go`), whose line ending `so a replaced kill` is now short, and make it read that a replaced kill still counts as paid, not the discard cost.
- 2026-11-02 rules-test-author: round 2, note 1 applied in a3c969d: `TestFR_TRN_003a_DiscardCostPaidOnDeclaration` now compares the discard event's `Zone` to `"Hand"`. Nothing else changed in any test; no helpers requested. Note 2 left for the engineer. Status `tests-review`.
- 2026-11-02 lead-architect: round-2 tests approved. a3c969d changes only line 86 of `TestFR_TRN_003a_DiscardCostPaidOnDeclaration`, `"hand"` to `"Hand"`, as note 1 asked; no other test changed. Against b28b653, `go test ./pkg/engine -run 'TestFR_TRN_003a_DiscardCost|TestFR_MOD_001_DiscardCost'` passes. Status `ready`; dispatching `kernel-engineer` (sonnet) for note 2 (the `payCosts` doc comment) and the Done When gates.
- 2026-11-02 kernel-engineer: round 2, status `review`. Answered lead note 2: re-wrapped the `payCosts` doc comment in `pkg/engine/placement.go` (comment only, lines filled evenly); it now reads that a replaced kill (not the discard cost) still counts as paid (SVY-R15). No test changed. Commands: `go test ./pkg/engine -run 'TestFR_TRN_003a_DiscardCost|TestFR_MOD_001_DiscardCost'` ok; `go tool task check` passed (all stages, lint PASS); `go tool task sim-compare -- autoloop/m2-wave-c` PASS, starter and stress byte for byte.
- 2026-11-02 lead-architect: approved, merged into `autoloop/m2-wave-c` as 1ae7a2e (`--no-ff`). Reviewed `git diff autoloop/m2-wave-c...m2/autoloop-wave-c/T-0131-discard-cost`, not the Log. Scope: only Allowed Files and the task file changed; `discard_cost_test.go` changes only in the two `test(T-0131)` commits (0c6d949, a3c969d); no existing test edited. Shape: `CostDiscard` and `Cost.UpTo` in `ability`; `costsPayable` skips a discard cost; `payCosts` pays it in printed order through the one `discardFromHand`, after the push; `discardFromHand(pid, n, upTo)` with `upTo` asks a `DecideHand` with `Min` 0 and `Max` `min(n, len(hand))` whenever the hand isn't empty, and `stepDiscard` passes `false`, so its behavior is unchanged. Fences: no card name in engine code, no new decision kind, no second discard loop. Note 2 met (b561f1d, comment only). Gates re-run by the lead: `go tool task check` exit 0 on the task branch and on the run branch after the merge; `go tool task sim-compare -- autoloop/m2-wave-c` byte-identical; `go tool task sim` clean, 200 games, P0 76.0% / P1 23.5% / draws 0.5%. Status `done`.
