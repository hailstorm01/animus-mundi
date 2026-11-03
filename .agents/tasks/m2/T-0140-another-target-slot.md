---
id: T-0140
title: An "another chosen" target slot must pick a different entity from the earlier slot of its clause (FR-TGT-002, SVY-R34, directives #88 and #91 (3), wave C shape, sonnet)
kind: shape
tier: sonnet
wave: C
status: in-progress
branch: m2/autoloop-wave-c/T-0140-another-target-slot
depends_on: [T-0116, T-0117, T-0118]
claims: [field:Target.Another, func:Target.AnotherSlot, func:chooseTargets, func:slotsLegal, func:chooseModes]
cards: []
requirements: [FR-TGT-002, FR-MOD-001, FR-MNV-001]
---

## Goal
A mode or Maneuver can print *"Chosen unit deals damage equal to its power to another chosen unit"* (*Ambush Tactics*' third mode, a deck card). The second slot must pick a different entity from the first, so the mode is legal (and a Maneuver with it enactable) only when both slots can be filled at once with different entities, the second slot is never offered the first slot's pick, and on a timeout it takes the first different legal entity in fixed order (`SVY-R34`). This is the distinctness shape task #88 and #91 (3) put before the *Ambush Tactics* card task, which follows it.

## Inputs
- FRD 01 (run branch, v3.20):
  - `FR-TGT-002`, "Another" Targets (`SVY-R34`, line 370): "A slot that says *"another chosen X"* (or *"other chosen"*) must pick an entity different from the slot it follows in the same clause (*Ambush Tactics*: *"Chosen unit deals damage equal to its power to another chosen unit"*). With no earlier slot in its clause, it must differ from the effect's source (*Tempest Runner*: *"Regroup another chosen allied unit"*). A mode or Maneuver with an *"another"* slot is legal only if enough different legal entities exist to fill every exactly-one slot at once, so *Ambush Tactics*' third mode is illegal with one unit on the battlefield (`SVY-R28`, `SVY-R29`). The *"another"* slot is offered every legal entity except the earlier pick. On a timeout, the earlier slot takes the first legal entity in fixed order and the *"another"* slot takes the first one after it that is different (`FR-MOD-001` §6)."
  - `FR-TGT-002`, Separate Target Clauses (`SVY-R35`, line 371): two slots may pick the same entity unless one says *"another"*. Different modes of one effect are separate slots (`SVY-R25`), so a pick in one chosen mode never restricts a slot in another.
  - `FR-MNV-001`, A Maneuver Needs Its Required Targets (`SVY-R29`, line 340): "Slots that must differ (*"another"*, `SVY-R34`) need different entities. … An *"up to N"* slot never stops it."
  - `FR-MOD-001` §5 (line 392): a mode is legal only if its exactly-one slots can be filled; §6 (line 393): timeouts take the first legal options in fixed order (`SVY-R16`).
  - `FR-KWD-002h` (Ward): an opponent's Ward asset is never a legal target, so it never counts toward "enough different entities".
- Survey: *Ambush Tactics* row (`docs/design/m2_starter_set_ability_survey.md`); ruling R34 (line 93).
- Outcome: [`docs/design/outcomes/2026-11-02-wave-c-repeated-distinct-targets-and-mode-costs.md`](../../../docs/design/outcomes/2026-11-02-wave-c-repeated-distinct-targets-and-mode-costs.md), question 3.
- Rulings that apply: `SVY-R11`, `SVY-R16`, `SVY-R25`, `SVY-R28`, `SVY-R29`, `SVY-R34`, `SVY-R35`.
- Directives: #88 (its last part: "distinctness between two slots …, its own shape task before the *Ambush Tactics* card task") and #91 (3) (`BOARD.md`).
- Kernel today (run branch at 7c1bd4b):
  - `ability.Target` (`ability/ability.go:152-161`): `Filter`, `UpTo`, `PerDiscard`. `Filter.Other` (`ability/ability.go:168`) already excludes the effect's source asset, which is `SVY-R34`'s no-earlier-slot case (*Tempest Runner*, M3).
  - `Step.Dealer` (`ability/ability.go:244`) and `Step.DealerSlot` (`ability/ability.go:250-254`) are the 1-based slot-reference idiom: 0 means none, k > 0 means `Effect.Targets[k-1]`.
  - `chooseTargets` (`targeting.go:23-41`) asks each slot in order through `targetDecision` (`targeting.go:47-68`), which `slotsLegal` (`placement.go:136-143`) also reads, slot by slot. `legalModes` (`placement.go:162-170`) and `placeable` (`placement.go:149-158`) get legality only through `slotsLegal`.
  - `chooseModes` (`placement.go:181-217`) joins the chosen modes' `Targets` and shifts each step's `Target` and `Dealer` by the slots before it.
  - `choose` (`choice.go:102`): a single legal option is taken unasked, a timeout takes `fixedOrderPicks` (`choice.go:188`), which follows the `FR-MOD-001` §6 order, not option order.
  - No card prints an "another" slot yet; *Ambush Tactics* is the first (the next task). `dealer_test.go`'s `dealerMode` has two plain `anyUnit()` slots; its file comment says distinctness is not pinned there. This task doesn't edit it.
- **Not reached by any M2 card:** an earlier slot whose pick leaves its "another" slot with no legal entity (the two filters differ, for example "chosen unit" then "another chosen opposing unit" with one opposing unit). FRD 01 restricts only the "another" slot's options, so the earlier slot is offered every legal entity, and an "another" slot left with none stays empty and asks nothing (`FR-TGT-003`), as any exactly-one slot with no legal entity does today. *Ambush Tactics*' two slots have the same filter, so once the mode is legal every earlier pick leaves a different entity. No test pins this case.

### Shape (lead decisions, binding)
- **`ability.Target.Another int`** (`ability/ability.go`): 0 means none. k > 0 means this slot must pick an entity different from the one chosen for `Effect.Targets[k-1]`, the earlier slot of the same clause (`SVY-R34`); k-1 is less than the slot's own index. The doc comment says that "another" with no earlier slot in its clause is `Filter.Other`, not this field. Add **`func (t Target) AnotherSlot() (slot int, ok bool)`**, mirroring `Step.DealerSlot`.
- **`chooseTargets`** (`targeting.go`): an `Another` slot's decision leaves out every Ref already chosen for slot k-1, so it is never offered, a single remaining option is taken unasked, and a timeout takes the first remaining one in fixed order. If no option remains it stays empty and asks nothing. Other slots unchanged. Doc comment states this.
- **`slotsLegal`** (`placement.go`): true only if the exactly-one slots (`UpTo == 0`, not `PerDiscard`) can all be filled **at once**: one legal entity each (from `targetDecision`'s options, as today), with every `Another` slot's entity different from the entity assigned to its slot k-1 when slot k-1 is exactly-one too. An "up to N" or `PerDiscard` slot still never makes it false, and an `Another` slot that follows one of them needs only a legal entity of its own (the earlier slot may pick none). Check a joint assignment, not slot by slot (#91 (3)); a small search over the exactly-one slots in slot order is enough. `legalModes`, `placeable` and the enact check get it through `slotsLegal`, unchanged. Doc comment cites `SVY-R34`.
- **`chooseModes`** (`placement.go`): when it joins modes, shift each joined target's `Another` by the slots before it, as it shifts `Target` and `Dealer`. The mode's own `Targets` slice belongs to the card definition and must not be modified in place (copy the slot before changing it).
- Nothing else changes. Resolution needs no second check: two different entities stay different, and `FR-TGT-003`'s verification is unchanged.
- **Sims:** no card uses `Another`, so `go tool task sim-compare -- autoloop/m2-wave-c` must pass.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
In a new file `pkg/engine/another_target_test.go`, with synthetic Maneuvers enacted by player 0 (`maneuver`, `steps`, `enactAction`, `offersEnact` in `maneuver_test.go`; `mode`, `scriptModal`, `labels` in `modal_test.go`; `dealerStep`, `anyUnit`, `returnMode`, `wantDamageFrom` in `dealer_test.go`). The "another" effect is *"Chosen unit deals damage equal to its power to another chosen unit"*: slots `anyUnit()` and `anyUnit()` with `Another: 1` on the second, and `dealerStep(1, 1)`. Give units enough health that nothing dies unless a test says so. No existing test is edited.
- `TestFR_MNV_001_AnotherSlotNeedsTwoDifferentUnitsToEnact` (`SVY-R34`, `SVY-R29`): a non-modal Maneuver with only the "another" effect is not among player 0's legal actions with one unit on the battlefield, and is with two. Also not with two units when one is an opposing Ward unit (`FR-KWD-002h`), since only one legal entity is left.
- `TestFR_MOD_001_AnotherSlotNeedsTwoUnits` (`SVY-R34`, `SVY-R28`, #91 (3) (a)): a "Choose one" Maneuver with modes "Kill chosen unit" then the "another" effect. With one unit on the battlefield, the dealer mode is not legal: the Maneuver is enactable and the mode decision offers (or answers unasked) only the kill mode. With two units both modes are offered.
- `TestFR_TGT_002_AnotherSlotExcludesEarlierPick` (`SVY-R34`, #91 (3) (b)): units X, Y, Z. The player picks Y for the first slot. The second slot's `DecideTarget` options are X and Z only, and Y deals its power to the unit picked there.
- `TestFR_MOD_001_AnotherSlotTimeoutTakesNextDifferent` (`SVY-R34`, §6, #91 (3) (c)): units X, Y, Z, both target decisions time out. The first slot takes the first unit in §6 fixed order and the second slot the next one in that order, never the same unit (check the dealt damage and its dealer). Use three units so the second slot is a real fallback, not the single-option fast path.
- `TestFR_TGT_002_AnotherSlotSingleRemainingOptionIsTakenUnasked` (`SVY-R34`, `FR-MOD-001` §1): with exactly two units, after the first slot picks one, the second slot is not asked (its pick is recorded as automatic) and takes the other.
- `TestFR_TGT_002_AnotherSlotInLaterModeFollowsItsOwnClause` (`SVY-R34`, `SVY-R35`, `SVY-R25`): a "Choose two" Maneuver with modes "Kill chosen unit" then the "another" effect, both chosen. Units X, Y, Z. The kill slot picks X and the dealer slot picks Y. The "another" slot is offered X and Z (not Y): it differs from its own clause's slot, not from the kill mode's pick. Picking X there is legal.
- `TestFR_MOD_001_AnotherSlotDoesNotChangeTheCardDefinition`: after enacting the "Choose two" Maneuver above with both modes, the definition's mode `Targets` still hold `Another: 1` on the dealer mode's second slot (the join copies, never edits, the definition), as `TestFR_MOD_001_ModalTriggerDoesNotChangeTheCardDefinition` does for steps.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected. A file-local helper in `another_target_test.go` (for example the "another" effect builder) is fine.

## Allowed Files
- `pkg/engine/ability/ability.go` (`Target.Another` and `Target.AnotherSlot` only)
- `pkg/engine/targeting.go` (`chooseTargets` and its doc comment)
- `pkg/engine/placement.go` (`slotsLegal`, `chooseModes` and their doc comments)

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run 'AnotherSlot'`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim-compare -- autoloop/m2-wave-c` passes (no card uses `Another` yet)

## Log
- 2026-11-02 lead-architect (wave C session 31): written against the run branch at 7c1bd4b (no `main` commits to sync). #88's last part and #91 (3) are the next work on the board: the distinctness shape task before the *Ambush Tactics* card task, a deck card (#14 (a)). Searched `pkg/engine` and `archive/tasks/task_archive.md` for an equivalent: `SVY-R11` makes one "up to N" slot's picks distinct and `Filter.Other` excludes the source, but nothing relates two slots, so `Target.Another` adds no duplicate, and its no-earlier-slot case reuses `Filter.Other`. Waiting triggers #6, #7 (parallel), #8 (this task doesn't touch `deck.go`), #15 (no turn hook) and #71 (no each-player card) not fired. Re-read `BOARD.md` before committing: no new or changed directive. Status `tests`.
- 2026-11-02 rules-test-author: wrote `pkg/engine/another_target_test.go`: TestFR_MNV_001_AnotherSlotNeedsTwoDifferentUnitsToEnact, TestFR_MOD_001_AnotherSlotNeedsTwoUnits, TestFR_TGT_002_AnotherSlotExcludesEarlierPick, TestFR_MOD_001_AnotherSlotTimeoutTakesNextDifferent, TestFR_TGT_002_AnotherSlotSingleRemainingOptionIsTakenUnasked, TestFR_TGT_002_AnotherSlotInLaterModeFollowsItsOwnClause, TestFR_MOD_001_AnotherSlotDoesNotChangeTheCardDefinition. Helpers requested: none (file-local helpers only: anotherEffect, anotherKillMode, anotherTargetAsks, anotherOffers, anotherTargetRecords). `go vet` fails only on the missing `ability.Target.Another` field, the contract this task adds. `cardtown lint rules` passes. Cases left out: none. Note: in the later-mode test the "another" slot picks X, which the kill mode has already killed, so that item fizzles at resolution; the test pins only the offered options, as the task describes.
- 2026-11-02 lead-architect (wave C session 31): tests approved (9ecab8e). Checked each against FRD 01: `SVY-R34` (line 370) for the joint legality with one unit, with an opposing Ward unit (`FR-KWD-002h`) and in a "Choose one" Maneuver (`SVY-R28`, `SVY-R29`), for the "another" slot never being offered the earlier pick, and for the timeout taking X then the next different unit Y in §6 fixed order; `FR-MOD-001` §1 for the single remaining option taken unasked; `SVY-R35` and `SVY-R25` for a later mode's "another" slot following its own clause, which fails if `chooseModes` doesn't shift `Another`; the definition test fails if the join edits a mode's `Targets` in place. `go vet` fails only on `ability.Target.Another`. The later-mode test's third pick (X, killed by the first mode) fizzles at resolution, as `SVY-R25` says; only the options are pinned, which is what the case needs. Status `ready`, dispatching `kernel-engineer` (sonnet).
