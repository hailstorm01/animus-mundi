---
id: T-0132
title: A step can run only if its controller has no cards in hand (wave C shape, sonnet)
kind: shape
tier: sonnet
wave: C
status: review
branch: m2/autoloop-wave-c/T-0132-hand-empty-condition
depends_on: [T-0131]
claims: [field:Step.IfHandEmpty, func:runSteps]
cards: []
requirements: [FR-MNV-001, FR-TRN-003a]
---

## Goal
A step can print *"Then, if you have no cards in hand, ..."* (*Steel Rain*'s last clause: *"Then, if you have no cards in hand, draw 2 cards."*). The step runs only if the ability's controller has no cards in hand when that step resolves, after the steps before it. Split from *Steel Rain* with T-0131 under the D2 size rule; the per-discard target choices and the card stay on directive #20 (b).

## Inputs
- Card text (`data/cards/maneuvers-path.csv` lines 18-20, *Steel Rain*): *"As an additional cost to enact this maneuver, discard up to 3 cards. For each card discarded this way, deal 2 damage to chosen asset. Then, if you have no cards in hand, draw 2 cards."* FRD 01's v3.7 change note 3 (line 74) gives the same text.
- FRD 01 (run branch):
  - `FR-MNV-001` (line 333): on declaration the Maneuver moves from Hand to the Resolution Stack; when it resolves, its effects execute, then it goes to its owner's Discard pile. So the Maneuver itself is never in its controller's hand while its steps run.
  - `FR-TRN-003a` (line 335): the discard cost is paid on declaration, so it has emptied the hand (or not) before any step runs.
  - `FR-TGT-001`, Untargeted Selections (`SVY-R3`, line 357): a later clause resolves after the earlier clauses of the same effect have finished, with nothing resolving in between. So "Then, if ..." is read after the steps before it.
  - `FR-TGT-003`, Composite Clauses (line 372): a clause that is not targeted resolves in full whatever happened to the targeted ones. The condition is not a target; it is read, not chosen.
  - `SVY-R14` (line 358) and `FR-TGT-003` (line 372): a step that does not run did nothing, so an "if you do" step right after it does not run either.
- Rulings that apply: `SVY-R3`, `SVY-R14`. No new ruling is needed: "you" is the ability's controller, and the clause is read when it resolves, as every untargeted clause is.
- Survey: the *Steel Rain* row (`docs/design/m2_starter_set_ability_survey.md` line 176): "if hand empty, `DRAW_CARD 2`".
- Design note (`docs/architecture/m2-ability-model.md`): line 39, `Step` and its `IfDone` gate; line 55, `resolution` records what each step did, for `IfDone`.
- Kernel today (run branch at dd3774c):
  - `ability.Step` (`ability/ability.go:230-241`): `IfDone bool` (line 240) is the one step gate, "runs only if the step just before it did at least one thing".
  - `runSteps` (`effects.go:49-60`): runs each step in order, skips an `IfDone` step whose predecessor did nothing, and records 0 for a skipped step in `r.done`. `resolveAbility` (`effects.go:46`) and inline replacement steps (`replace.go:114`) both go through it.
  - `Player.Hand []*Instance` (`state.go:104`). The Maneuver leaves hand on declaration (`Apply`'s `ActEnact` case, `takeCard`), so it is never counted.
  - `doc.go:29-34` lists the steps and gates the kernel supports.

### Shape (lead decisions, binding)
- **`ability.Step.IfHandEmpty bool`**, a new field after `IfDone`: runs only if the ability's controller has no cards in hand when the step resolves ("if you have no cards in hand"). No card names in `ability`.
- **One gate check, in `runSteps`.** A step runs only if every gate it sets holds: `IfDone` as today, and `IfHandEmpty` when `len(g.Players[r.it.Controller].Hand) == 0` at that moment. It is read when the step is reached, after the steps before it have run, never on placement. A step skipped by either gate counts 0 in `r.done`, as an `IfDone` skip does today, so a following `IfDone` step does not run. Put the gate test in one small function beside `runSteps` (for example `(g *Game) stepRuns(r *resolution, i int, st ability.Step) bool`), so the next gate is one more line there and not a second loop. No new `Op` or `Subject`, and no change to any `step<Op>`.
- **Whose hand:** the ability's controller's (`r.it.Controller`), never an opponent's, for a Maneuver, a trigger or a replacement alike.
- **Out of scope:** the target choice per discarded card and *Steel Rain* itself (#20 (b), behind #59 (2)'s ruling); any other condition.
- **Sims:** no card uses `IfHandEmpty` yet, so `go tool task sim-compare -- autoloop/m2-wave-c` must pass (starter and stress sims byte-identical to the run branch).

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
In a new file `pkg/engine/hand_empty_test.go`, on synthetic Maneuvers built with `maneuver_test.go`'s `maneuver`, `steps`, `enactAction` and `lastDiscard`, and `discard_cost_test.go`'s `discardCostManeuver`, `hqPing` and `pingCount`. Player 0 is the active player and enacts. Set its hand explicitly (`s.g.Players[0].Hand = nil`, then `s.give`), and give it at least two deck cards where a step draws. Where the gated step must be seen in the event log it is `hqPing()` with `IfHandEmpty` set, because a draw logs no event. No existing test is edited.
- `TestFR_MNV_001_HandEmptyStepRunsWithAnEmptyHand`: the Maneuver is the controller's only card in hand, and its one step is "draw 2 cards" (`OpDraw`, controller) with `IfHandEmpty`. After `Apply`, the controller's hand holds exactly the two cards that were on top of its deck, and the Maneuver is in its owner's discard pile.
- `TestFR_MNV_001_HandEmptyStepSkippedWithACardInHand`: as above, with one other card in hand. After `Apply`, the hand holds only that card, the deck is unchanged, and the Maneuver is in its owner's discard pile.
- `TestFR_MNV_001_HandEmptyReadAfterEarlierSteps`: (a) steps "discard 1 card" (`OpDiscard`, controller) then `hqPing` with `IfHandEmpty`, with one other card in hand: that card is discarded and the ping is logged once. (b) steps "draw 1 card" then `hqPing` with `IfHandEmpty`, with no other card in hand: the drawn card is in hand and no ping is logged.
- `TestFR_MNV_001_HandEmptyReadsTheControllersHand`: (a) the controller holds one other card and the opponent's hand is empty: no ping. (b) the controller holds no other card and the opponent holds cards: the ping is logged once.
- `TestFR_MNV_001_SkippedHandEmptyStepCountsAsNotDone` (`SVY-R14`): steps `hqPing` with `IfHandEmpty`, then 2 damage to the opposing Headquarters with `IfDone`, with one other card in hand: neither damage event is logged.
- `TestFR_TRN_003a_DiscardCostThenHandEmptyStep`: `discardCostManeuver` with up to 3 and the one step `hqPing` with `IfHandEmpty`, with two other cards in hand. (a) A scripted chooser discards both: the ping is logged once. (b) It discards one: no ping, and the other card is still in hand.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected.

## Allowed Files
- `pkg/engine/ability/ability.go` (`Step.IfHandEmpty` only)
- `pkg/engine/effects.go` (`runSteps`, the gate function beside it, and their doc comments)
- `pkg/engine/doc.go` (the supported-steps list: name the hand-empty gate beside `IfDone`)

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run 'TestFR_MNV_001_HandEmpty|TestFR_MNV_001_SkippedHandEmpty|TestFR_TRN_003a_DiscardCostThenHandEmpty'`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim-compare -- autoloop/m2-wave-c` passes (no card uses the condition yet)

## Log
- 2026-11-02 lead-architect (wave C session 19): queued with T-0131's task commit (D2.4).
- 2026-11-02 lead-architect (wave C session 20): Inputs, Shape, Tests, Allowed Files and Done When written against the run branch at dd3774c. `claims:` narrowed from `condition:hand-empty` to the new field and `runSteps`, the one place step gates are checked. Reuse search: `IfDone` is the only step gate (`effects.go:54`); trigger gates (`TurnOwner`, `ToZone`, `Cornered`) are read on the event, not per step, and nothing in `pkg/engine` or `archive/tasks/task_archive.md` reads a hand at resolution. No ruling needed (Inputs). Status `tests`.
- 2026-11-02 rules-test-author: wrote pkg/engine/hand_empty_test.go: TestFR_MNV_001_HandEmptyStepRunsWithAnEmptyHand, _HandEmptyStepSkippedWithACardInHand, _HandEmptyReadAfterEarlierSteps, _HandEmptyReadsTheControllersHand, _SkippedHandEmptyStepCountsAsNotDone, TestFR_TRN_003a_DiscardCostThenHandEmptyStep. Helpers requested: none (local test helpers handEmpty and handEmptyScene live in the new file). Cases left out: none. go vet fails only on the missing field ability.Step.IfHandEmpty (the shape under test). cardtown lint rules passes.
- 2026-11-02 lead-architect: tests approved (8645967). The six tests check the cited text: the gate is read on the controller's hand as the step resolves, after earlier steps (a discard empties it, a draw fills it; `SVY-R3`, line 357), with the Maneuver on the stack and so never counted (`FR-MNV-001`, line 333); the discard cost empties the hand before any step runs (`FR-TRN-003a`, line 335); a skipped gated step counts as not done, so the next `IfDone` step does not run (`SVY-R14`). The gated step is watched through logged Headquarters damage or the hand and deck, not a draw event. `go vet` fails only on the task's own claim, `Step.IfHandEmpty`. Status `ready`, kernel-engineer (sonnet).
- 2026-11-02 kernel-engineer: added `Step.IfHandEmpty` (ability.go, gofmt realigned the neighbouring struct fields' comments) and `(g *Game) stepRuns` beside `runSteps` (effects.go): one function holding the `IfDone` test (moved unchanged in meaning) and the new hand test on `g.Players[r.it.Controller].Hand`, read as the step is reached; a skipped step still counts 0 in `r.done`. doc.go names the gate beside `IfDone`. Reuse: extended the existing `IfDone` gate site, no new Op/Subject, no step<Op> change. Ran: the listed tests pass; `go tool task check` passes; `go tool task sim-compare -- autoloop/m2-wave-c` passes (byte-identical). Reviewer: check the gofmt-only realignment in the Step struct.
