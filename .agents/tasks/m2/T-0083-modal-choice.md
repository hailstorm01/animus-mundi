---
id: T-0083
title: Modal triggers, "choose one" of their modes on placement, before targets (FR-MOD-001 §5, §6, wave B shape, sonnet)
kind: shape
tier: sonnet
wave: B
status: in-progress
branch: m2/T-0083-modal-choice
depends_on: [T-0006, T-0030, T-0081]
claims: [effect:modes, decision:mode-on-placement]
cards: []
requirements: [FR-MOD-001, FR-TGT-001, FR-TGT-003]
---

## Goal
A triggered ability can be modal: its `Effect` lists `Modes`, each an `Effect` of its own, and `ChooseN` says how many different modes are chosen. The controller chooses the modes as the trigger is placed on the stack, before any target, and then chooses only the chosen modes' targets. The item resolves only the chosen modes. The first card, *Rogue Fleshshaper* ("When this unit deploys, choose one: - Chosen asset heals 1. - Deal 1 damage to chosen non-Scratch asset."), is the next card task. *Ambush Tactics* ("Choose two:", a Maneuver, wave C) reuses the shape later.

## Inputs
- FRD 01:
  - `FR-MOD-001` §5 (line 354): "Choose N" takes N **different** modes unless the card says otherwise (`SVY-R6`).
  - `FR-MOD-001` §6 (line 357): a timed-out modal choice "selects the first listed legal mode(s)".
  - `FR-MOD-001` §1 (line 350): a choice with only one legal option is made automatically.
  - `FR-TGT-001`, "Targets Are Locked on Placement" (line 326, `SVY-R3`), and `FR-MOD-001` §7 (line 362, `SVY-R9`): every commitment is made on placement. A mode decides which targeted clauses exist, so the mode is chosen on placement, before targets.
  - `FR-TGT-003` (lines 339-343): the chosen mode's targets are verified again at resolution, with partial invalidation.
- Survey rows: Rogue Fleshshaper (line 152, wave B, "Triggered, modal"), Ambush Tactics (line 162, wave C). Survey ruling R6.
- Rulings: `SVY-R6`, `SVY-R3`, `SVY-R9`, `SVY-R16`. None new.
- Design note: `docs/architecture/m2-ability-model.md`, Key types (`Effect`, updated with this task) and Open Question 11 (added with this task).
- Kernel today:
  - `DecideMode` already exists (`choice.go`): its fallback takes the first listed options, `Valid` rejects a repeated mode unless `Repeats`, and `soleAnswer` fast-paths a decision whose `Min == Max == len(Options)`. Its tests are in `choice_test.go` (T-0006).
  - `flushTriggers` (`trigger.go`) places each trigger: `accepts` asks the "you may" question (`SVY-R9`), then `chooseTargets` (`targeting.go`) asks each target slot of `p.trig.Effect`, then `push` puts a `StackAbility` item with `Effect: p.trig.Effect`.
  - `chooseTargets` builds each slot's decision with `pileDecision` (a `ZoneDiscard` slot) or `assetDecision`, over the entities `matchesCard` or `legalTarget` accept.

### Shape (lead decisions, binding)
- **`ability.Effect`** gains two fields, after `Steps`:
  - `Modes []Effect // "choose N": the modes in printed order; an effect with Modes has no Targets or Steps of its own (FR-MOD-001 §5)`
  - `ChooseN int // with Modes: how many different modes are chosen ("choose one" is 1, SVY-R6)`

  Update the `Effect` doc comment to say so. A mode's first step never has `IfDone`: "if you do" refers to a step inside the same mode.
- **When:** in `flushTriggers`, after `accepts` and before `chooseTargets`. A modal trigger that is declined asks nothing.
- **Which modes are offered:** a mode is legal if each of its exactly-one target slots (`UpTo` 0) has at least one legal entity at that moment. An "up to N" slot never makes a mode illegal. The decision lists only legal modes, in printed order. That makes §6's "first listed legal mode(s)" the existing `DecideMode` fallback, and §1's fast path the existing `soleAnswer`.
  - **One place builds a target slot's decision.** Factor the per-slot decision building out of `chooseTargets` into one function in `targeting.go` that both the mode legality check and `chooseTargets` call. The legality check must not repeat the `pileDecision` / `assetDecision` choice or the filter predicates.
  - **No legal mode:** ask nothing, and place the trigger with an empty effect, so it resolves and does nothing, as a deploy trigger without legal targets still lets the asset enter (`FR-TGT-003`). Fewer legal modes than `ChooseN`: choose all the legal ones. No M2 card reaches either case (design note Open Question 11). They are not tested.
- **The decision:** `Decision{Player: controller, Kind: DecideMode, Min: n, Max: n}`, where `n` is `min(ChooseN, legal modes)`. `Prompt` is `"<label>: choose one"` for `ChooseN` 1, otherwise `"<label>: choose <N>"`, where `<label>` is the trigger's label (`"Rogue Fleshshaper: trigger 1"`). Each option's `Label` is `"mode <k>"`, with `k` the mode's 1-based printed position, and `Ref` is the trigger's source asset ID. `Repeats` stays false (`SVY-R6`). It goes through `g.choose`, like every decision.
- **The item's effect:** the chosen modes, joined into one `Effect` in printed order (not pick order): their `Targets` appended, their `Steps` appended, and each `SubjectTarget` step's `Target` shifted by the number of target slots before its mode. `flushTriggers` sets the pending trigger's local copy of `Effect` to it before `chooseTargets` and `push`, so the stack item carries the effect it resolves, and `resolveAbility`, `PileSlot`, the view and the monitors need no change. Build new slices: never append onto a mode's own `Targets` or `Steps`, so the card's `Def` is never changed.
- Put the mode choice and the joining beside `accepts` in `trigger.go`, the other question asked on placement.
- **`doc.go`:** under "What runs now", add that a trigger may be modal ("choose N" modes, chosen on placement before its targets, `FR-MOD-001` §5).

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
To be written in `pkg/engine/modal_test.go`, with synthetic cards only (Rogue Fleshshaper is the next task). The file won't compile until `Effect.Modes` and `Effect.ChooseN` exist. Observable steps: `OpDraw` (hand size), `OpHeal` or `OpDealDamage` on a target, `OpAddArmor` on `SubjectSelf`. Required cases:
- **Asked on placement, before targets:** `TestFR_MOD_001_ModalTriggerAsksForItsModeBeforeItsTargets`. A deploy trigger, `ChooseN` 1, two modes, each with one exactly-one target that has two or more legal entities. The controller is asked a `DecideMode` with two options, `"mode 1"` and `"mode 2"`, then the chosen mode's `DecideTarget`, and nothing for the other mode. Both questions come before the item's `place` event.
- **Only the chosen mode resolves:** `TestFR_MOD_001_ModalTriggerResolvesOnlyTheChosenMode`. Answer mode 2: mode 2's step happens and mode 1's doesn't.
- **Timeout:** `TestFR_MOD_001_TimedOutModeIsTheFirstListed`. A timeout, or no chooser, chooses mode 1.
- **Illegal first mode:** `TestFR_MOD_001_ModeWithoutALegalTargetIsNotOffered`. Mode 1's only target slot (exactly one) has no legal entity, and mode 2 has one. The mode decision has the single option `"mode 2"` and is answered automatically (`FR-MOD-001` §1), and mode 2 resolves. With a timeout chooser, mode 2 still resolves (§6: the first listed **legal** mode).
- **Up to N:** `TestFR_MOD_001_UpToTargetDoesNotMakeAModeIllegal`. Mode 1's only target is "up to 1" and has no legal entity. Both modes are offered.
- **Choose two:** `TestFR_MOD_001_ChooseTwoResolvesTwoDifferentModes`. Three modes, `ChooseN` 2, answered with modes 3 and 1. Both of those modes' steps happen, mode 2's does not, and each chosen mode's target step acts on the entity chosen for that mode. An answer repeating a mode is already covered by `TestFR_MOD_001_DistinctModesRepeatedPickIsInvalid`, so it is not repeated here.
- **Second verification:** `TestFR_TGT_003_ChosenModesTargetIsVerifiedAgainAtResolution`. The chosen mode's target leaves the battlefield before the item resolves. The mode's targeted step does nothing, and no step of the other mode runs.
- **Optional and modal:** `TestFR_MOD_001_DeclinedModalTriggerAsksNoMode`. An optional modal trigger that is declined asks no `DecideMode` and is not placed.
- **Definition unchanged:** `TestFR_MOD_001_ModalTriggerDoesNotChangeTheCardDefinition`. After a modal trigger resolves twice with different modes, the synthetic card's `Def` (its modes' `Targets` and `Steps`) equals what it was before.
- **Existing tests:** every test in `pkg/engine` passes unchanged.
- **Left out:** no legal mode, and fewer legal modes than `ChooseN` (see Shape; design note Open Question 11).

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected. If a test needs one, `rules-test-author` names it here.

## Allowed Files
- `pkg/engine/ability/ability.go` (`Effect.Modes`, `Effect.ChooseN`)
- `pkg/engine/trigger.go` (the mode choice in `flushTriggers`, and joining the chosen modes)
- `pkg/engine/targeting.go` (one function builds a target slot's decision)
- `pkg/engine/doc.go` (the "What runs now" line)

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run 'TestFR_(MOD_001|TGT_003)'`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] No card uses modes yet, so the sims are unchanged: `go tool task sim-compare -- autoloop/m2-wave-b` passes

## Log
- 2026-10-27 lead-architect (wave run B, session 20): Wrote the task. The run branch has everything `main` has (`git log autoloop/m2-wave-b..main` is empty), and no `open: before` directive is on the board. Waiting triggers checked against `autoloop/m2-wave-b` a996ecc: #6 and #7 (mode is `serial`), #8 (no `deck.go` change), #15 (no turn hook), #20 (b) (no wave C shape task; modes are a wave B shape), #58 (no exhaust by a cost or an effect) and #59 (2) (no repeated targets). None has fired.
  - Wave B cards still on `NotYetExecutable`: only Rogue Fleshshaper (kept by the run header: Ambush Tactics, a deck card, needs the same modal shape, #14 (b), (c)). It needs this shape plus its definition. Together that is six non-test files, so it is split in two: this task, then the card task (`cards/path.go`, `cards/cards.go`), not written yet. Their claims don't overlap.
  - Reuse search: `DecideMode`, its fallback, its distinctness check and its fast path already exist and are reused unchanged. No `Modes` field or mode composition exists in `pkg/engine` or the `done` tasks. The design note already planned `Modes` with `ChooseN` on `Effect`.
  - Placement: the item carries the joined effect, so no stack, resolution, view or monitor code changes. The targeting monitor still sees each Target decision followed by its `place` event, because the mode is asked before the targets.
- 2026-10-27 rules-test-author: Wrote `pkg/engine/modal_test.go` (commit 06f213d), 9 tests: `TestFR_MOD_001_ModalTriggerAsksForItsModeBeforeItsTargets`, `..._ModalTriggerResolvesOnlyTheChosenMode`, `..._TimedOutModeIsTheFirstListed`, `..._ModeWithoutALegalTargetIsNotOffered`, `..._UpToTargetDoesNotMakeAModeIllegal`, `..._ChooseTwoResolvesTwoDifferentModes`, `TestFR_TGT_003_ChosenModesTargetIsVerifiedAgainAtResolution`, `..._DeclinedModalTriggerAsksNoMode`, `..._ModalTriggerDoesNotChangeTheCardDefinition`. `go vet` fails only on the missing `Effect.Modes` / `ChooseN`; `cardtown lint rules` passes.
  - Helpers requested in `scenario_test.go`: none. The test file defines its own local helpers (`scriptModal`, `newModalBoard`, `mode`, `modalTrigger`, `modeRecords`, `placeCount`, `totalArmor`, `labels`, `modeAnswer`).
  - Left out, as the task says: no legal mode, and fewer legal modes than `ChooseN` (Open Question 11).
  - Pinned beyond the prose, from the Shape section: prompt text `"<card name>: trigger 1: choose one"` / `"...: choose 2"`; an auto-answered single legal mode shows in the log as a Mode decision record with options `["mode 2"]` and How `"auto"`; mode targets are asked in printed order, not pick order; the Def test gives each mode's slices spare capacity and checks nothing was appended past their length.
- 2026-10-27 lead-architect (wave run B, session 20, D4): **Tests approved, `ready`.** I checked the 9 tests in 06f213d against `FR-MOD-001` §1, §5, §6 and §7, `FR-TGT-001` (`SVY-R3`) and `FR-TGT-003`. Mode before targets and before the `place` event, only the chosen modes' targets and steps, the first listed mode on a timeout or with no chooser, the single legal mode answered automatically and chosen on a timeout, an empty "up to" slot leaving its mode legal, "choose two" with each mode acting on its own target, the second verification of a mode's target, a declined optional modal trigger asking no mode, and the `Def` left unchanged. Each follows from the cited text or the task's Shape (the legality reading is design note Open Question 11). The prompts, labels and printed-order target questions they pin are the Shape's. The test file's own helpers are setup only; every shared helper they call already exists.
