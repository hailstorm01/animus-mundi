---
id: T-0163
title: An activated ability can be activated only if its required targets exist (FR-ACT-001, SVY-R36, wave D shape, sonnet)
kind: shape
tier: sonnet
wave: D
status: done
branch: m2/autoloop-wave-d/T-0163-activation-needs-targets
depends_on: [T-0158, T-0162]
claims: [check:activate-placeable]
cards: []
requirements: [FR-ACT-001]
---

## Goal
`LegalActions` offers an activated ability, and `check` accepts it, only if the ability could be placed with everything it requires: its costs, a legal entity for each exactly-one target slot, and N legal modes for a *"Choose N"*. This is #105 (1). No card uses it yet; *Sinuous Growth*, *Boneyard Scavenger* and *Ruin Partisans* come later, in card tasks.

## Inputs
- FRD 01 v3.21: `FR-ACT-001`, the bullet *An Activated Ability Needs Its Required Targets* (`SVY-R36`, line 342). `FR-MNV-001` and `SVY-R29` for the conditions it shares with a Maneuver. `SVY-R34` for slots that must differ, `SVY-R2` for a sacrificed card, `FR-TGT-003` for what happens after declaration.
- Outcome: `docs/design/outcomes/2026-11-04-wave-d-activation-without-target-and-deck-search.md`, its directive row #105 (1).
- Survey rows: wave D (`docs/design/m2_starter_set_ability_survey.md`). No card in this task.
- Design note: `docs/architecture/m2-ability-model.md`, Actions paragraph (line 53), which already plans this.
- Kernel today:
  - The `ActActivate` case of `check` (`pkg/engine/actions.go:195`) asks only `costsPayable`. The `ActEnact` case (`:183`) asks `placeable`.
  - `placeable` (`pkg/engine/placement.go:108`) checks `costsPayable` first, then `legalModes` or `slotsLegal`, so it already covers every condition `SVY-R36` names, including `SVY-R34`'s joint assignment and the *"up to N"* exemption.
  - `declare` chooses targets before `payCosts`, so a card sacrificed as a cost is still on the battlefield when `placeable` looks, and can't satisfy its own pile slot (`SVY-R2`).

### Shape (lead decisions, binding)
- **`check`'s `ActActivate` case calls `placeable`** with `activatedPlacement(...)` and the ability's `Effect`, in place of `costsPayable`. Its `illegal` message keeps `FR-ACT-001` and adds `SVY-R36`, and says that a cost can't be paid or a required target or mode has no legal choice. No new helper, no second legality check, and no change to `declare`, `slotsLegal`, `legalModes` or `costsPayable`.
- **Triggers don't change.** They keep placing whatever they find (`SVY-R30`, `FR-TGT-003`).
- **Comments that stop being true:**
  - `placeable` (`placement.go:103-107`): it applies to enacting a Maneuver and to activating an ability (`SVY-R29`, `SVY-R36`). A trigger is still placed whatever it finds.
  - `chooseModes` (`placement.go:135-137`): an enacted Maneuver or an activated ability never gets there with fewer than `ChooseN`. Only a trigger does.
  - `doc.go`'s activated-abilities bullet (`pkg/engine/doc.go:44-48`): it gains that the ability is offered only if `placeable` holds (`SVY-R36`).
- **Sims:** no card definition has an activated ability yet, so play doesn't change. `go tool task sim-compare -- autoloop/m2-wave-d` should pass byte for byte. If it differs, don't change code to make it match: report the first differing game in the Log and hand off.
- **Reuse check:** `grep placeable|costsPayable` finds one legality path, and this task sends activation down it. A lead probe in the sync worktree, since reverted, made that one-line change and ran `go test ./...`: every package passes, so no existing test pins the old behavior.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
Synthetic cards only, added to `pkg/engine/activated_test.go`, using its `actUnit`, `activate`, `offersActivate`, `actKill`, `actExhaust` and `actRes` helpers and `s.reject`. "Not offered" below always means both: `offersActivate` is false and `s.reject(activate(...))` holds.

**Authorized edit of an existing comment:** the file header of `activated_test.go` says every exactly-one slot below has a legal entity and that board directive #104 leaves the other case open. Rewrite those two sentences to say the file's tests fill every exactly-one slot except the `SVY-R36` cases, which name it. This lands in the `test(T-0163)` commit. No existing test function changes.

- `TestFR_ACT_001_NoLegalTargetCannotActivate`, subtests:
  - *Empty pile:* a kill-self ability whose exactly-one slot is an allied pile filter (`ZoneDiscard`), with the discard pile empty. The source itself carries the filter's subtype, so its own card would match if it counted (`SVY-R2`). Not offered. Nothing is paid and nothing moves: the unit stays on the battlefield, Readied, the pile stays empty, no decision is asked.
  - *Only non-matching cards:* the same ability with a pile holding only a card that fails the filter. Not offered, and nothing changes.
  - *A cost that would otherwise be paid:* an ability costing resources and exhaust, with an exactly-one slot naming an opposing unit, none on the battlefield. Not offered. The pool stays as set and the source stays Readied.
  - *A legal entity appears:* from the empty-pile scene, put a matching card in the pile. Now it is offered, and activating it returns that card to hand with the source in the discard pile.
- `TestFR_ACT_001_UpToSlotStillActivates`: an *"up to 2"* slot with no legal entity. Offered, its costs are paid, nothing is asked, and it resolves with no target affected.
- `TestFR_ACT_001_WardedOnlyTargetCannotActivate`: the only entity matching an exactly-one slot naming an opposing unit is a Warded opposing unit (`FR-KWD-002h`, `FR-TGT-001`). Not offered. Read the keyword's rule in FRD 01 before writing the setup. If that rule makes the entity legal in this setup, leave the case out and say so in the Log.
- `TestFR_ACT_001_AnotherSlotNeedsTwoEntities` (`SVY-R34`): an ability with an exactly-one slot and an "another" slot that must differ from it. With one legal entity, not offered. With two, offered, and the two slots take different entities.
- `TestFR_ACT_001_ChooseNNeedsEnoughLegalModes`, table: a *"Choose one"* ability with no legal mode is not offered; with one legal mode it is, and activating it resolves that mode. A *"Choose two"* ability with one legal mode is not offered; with two it is.
- `TestFR_ACT_001_LegalTargetDoesNotExcuseACost`: a legal entity exists and the resources cost can't be paid. Not offered, and nothing changes.

Existing tests already cover an ability with no target slot needing only its costs (`TestFR_ACT_001_CostsPaidAndEffectResolves`, `TestFR_ACT_001_UnpayableIsNotOffered`), and costs staying paid when a target becomes illegal (`TestFR_TGT_003_FizzledAbilityLeavesSourceOnBattlefield`). Don't repeat them.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected. The test file may declare its own.

## Allowed Files
- `pkg/engine/actions.go` (the `ActActivate` case of `check`)
- `pkg/engine/placement.go` (the comments of `placeable` and `chooseModes`)
- `pkg/engine/doc.go` (the activated-abilities bullet)

## Done When
- [x] The listed tests pass: `go test ./pkg/engine -run 'TestFR_ACT_001_(NoLegalTargetCannotActivate|UpToSlotStillActivates|WardedOnlyTargetCannotActivate|AnotherSlotNeedsTwoEntities|ChooseNNeedsEnoughLegalModes|LegalTargetDoesNotExcuseACost)'`
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks
- [x] `go tool task sim-compare -- autoloop/m2-wave-d` passes (no card has an activated ability; see Shape, Sims)
- [x] `cards.NotYetExecutable` is unchanged.

## Log
- 2026-11-04 lead-architect (wave D run, session 7): synced `main` into the run branch (958db4a, two commits: outcome and FRD 01 v3.21). `BOARD.md` conflicted only in the counters and next to the new row; kept #104, added #105, `Last directive` #105, `Last task` T-0162, and no number is duplicated or missing. #105's trigger fired: the outcome reads `Status: applied: FRD 01 v3.21, 2026-11-04`. Set to `open`, and part (1) goes first as this task. #104 is `applied: answered by #105` and archived, as #105 asks. Design note Open Question 17 closed. Other waiting directives checked, none fired: #6, #7 (mode is `serial`); #8 (this task doesn't touch `deck.go`); #15 (no turn-boundary hook); #71, #91 (2) (no such card). Not `WAVE-COMPLETE`: all four header cards are on `NotYetExecutable`. Three non-test files. Status `tests`; dispatching `rules-test-author`.
- 2026-11-04 rules-test-author: tests written in `pkg/engine/activated_test.go` (header comment rewritten as authorized; no existing test changed). `go vet` clean, `cardtown lint rules` passes. Against the current kernel the following fail as expected: `NoLegalTargetCannotActivate` (all four subtests), `WardedOnlyTargetCannotActivate`, `AnotherSlotNeedsTwoEntities/one_legal_entity`, `ChooseNNeedsEnoughLegalModes` ("choose one, no legal mode", "choose two, one legal mode"). These pass already and guard the other side: `UpToSlotStillActivates`, `LegalTargetDoesNotExcuseACost`, `AnotherSlotNeedsTwoEntities/two_legal_entities`, the "one legal mode" and "two legal modes" ChooseN rows. Tests: TestFR_ACT_001_NoLegalTargetCannotActivate (empty pile, only non-matching cards, a cost that would otherwise be paid, a legal entity appears), TestFR_ACT_001_UpToSlotStillActivates, TestFR_ACT_001_WardedOnlyTargetCannotActivate, TestFR_ACT_001_AnotherSlotNeedsTwoEntities, TestFR_ACT_001_ChooseNNeedsEnoughLegalModes (4 rows), TestFR_ACT_001_LegalTargetDoesNotExcuseACost. Ward: FR-KWD-002h makes an opposing Warded unit not choosable by activated abilities, so the case is included. Helpers requested: none (file-local `actRecaller`, `actNotOffered`, `actHasMode`, `actPileFilter`). Cases left out: none.
- 2026-11-04 lead-architect (D4): tests reviewed, approved. Each case checks `SVY-R36`'s conditions against FRD 01 line 342 through `LegalActions`, `check` and the state after, not through `placeable`. The empty-pile source carries the filter's subtype, so its own card can't satisfy the slot (`SVY-R2`). `killChosenOpposing` names units only, so the opposing Headquarters doesn't fill its slot. `FR-KWD-002h` makes a Warded opposing unit unchoosable here, so that case stays. The failures on the current kernel, re-run by the reviewer, are exactly the not-offered cases plus the subtest that starts with one; `UpToSlotStillActivates`, `LegalTargetDoesNotExcuseACost` and the cases with enough entities or modes pass already and guard the other side. The only existing change is the authorized header comment. Two nits, no rework: `UpToSlotStillActivates` also cites `SVY-R11`, which it doesn't exercise, and `actHasMode` could be `slices.Contains`. Status `ready`; dispatching `kernel-engineer` (sonnet).
- 2026-11-04 kernel-engineer: `check`'s `ActActivate` case now calls `placeable(activatedPlacement(...), act.Effect)` in place of `costsPayable`; message cites FR-ACT-001, SVY-R36. Reused `placeable`/`activatedPlacement`; no new helper, no change to declare, slotsLegal, legalModes, costsPayable. Updated comments of `placeable`, `chooseModes` and the doc.go bullet. Ran: `go test ./pkg/engine -run TestFR_ACT_001_` ok; `go tool task check` passed (lint all PASS); `go tool task sim` ran; `go tool task sim-compare -- autoloop/m2-wave-d` passed byte for byte. NotYetExecutable untouched.
- 2026-11-04 lead-architect (wave D run, session 7, D5): **Approved, `done`.** Merged into `autoloop/m2-wave-d` as c112a32. Review of `git diff autoloop/m2-wave-d...m2/autoloop-wave-d/T-0163-activation-needs-targets` (07280c3), not the Log. Scope: only the three Allowed Files, the test file and this file changed, and `git diff ca51e4f` shows no `_test.go` change after `test(T-0163)`. Fences: `go run ./cmd/cardtown lint fences` exits 0. Reuse: `check` calls `placeable` with `activatedPlacement`, no second legality path, and nothing else in the kernel changed. Rules fidelity: `placeable` holds every condition of `FR-ACT-001`'s `SVY-R36` bullet in general, and triggers still don't read it. The engineer's Log rests on the tail of `check` and `sim`, so the reviewer re-ran everything with exit codes in a scratch worktree of the task branch: the six listed tests, `go tool task check`, `go tool task sim`, `go run ./cmd/cardtown lint fences` and `go tool task sim-compare -- autoloop/m2-wave-d` `[PASS]`, all exit 0. `go tool task check` and `go tool task sim` also exit 0 on the merged run branch: 200 games, P0 82.5% / P1 17.5% / draws 0.0%, 42/50 and 32/50 main cards executable, no monitor violation. `cards.NotYetExecutable` is unchanged. The engineer removed its worktree with `--force` and couldn't say why; the branch holds the implementation commit and the reviewer's own worktrees removed without it, so nothing was lost. One nit, no rework: the new `doc.go` line runs longer than its neighbours. Status `done`.
