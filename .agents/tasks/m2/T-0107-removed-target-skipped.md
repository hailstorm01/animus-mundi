---
id: T-0107
title: A step acts only on chosen assets still on the battlefield, so a later step does nothing to a target an earlier one removed (FR-ZON-004, FR-TGT-003, directive #82, wave G.2 shape, sonnet)
kind: shape
tier: sonnet
wave: G.2
status: review
branch: m2/T-0107-removed-target-skipped
depends_on: [T-0102]
claims: [rule:removed-target-skipped, helper:step-assets-target]
cards: []
requirements: [FR-ZON-004, FR-TGT-003]
---

## Goal
Every asset step reads a target slot through one check: a chosen asset that an earlier step (or mode) of the same effect removed from the battlefield is no longer that target, so the later step does nothing to it and still acts on the rest of its slot. Today only `stepRelocate` skips such an asset; damage, heal, Armor and stat steps act on it, and damage to it fires its damage-taken triggers. This applies board directive #82 (design note Deliverable 6e, first task), so wave C's *Ambush Tactics* and `MOVE_ZONE` only combine shapes.

## Inputs
- FRD 01: `FR-ZON-004` (lines 209-221: an asset that leaves the battlefield is a new object in its new zone); `FR-TGT-003` (lines 356-361: two-stage verification, Partial Invalidation at line 358).
- Survey rows: none (wave G.2 has no survey row and no cards).
- Rulings that apply: **user ruling, 2026-10-30 (board directive #82):** a chosen asset that an earlier step or mode of the same effect removed from the battlefield is no longer that target. The later step does nothing to it, and still acts on the rest of its slot (`FR-ZON-004`, `FR-TGT-003` partial invalidation).
- Design note: `docs/architecture/m2-ability-model.md`, Deliverable 6e; Key types (`Step`, `resolution`).
- Board directive #82.
- Kernel today:
  - `resolveAbility` (`effects.go`) re-verifies each target once, when the item starts to resolve, and stores the survivors in `r.chosen`.
  - `stepAssets`' `SubjectTarget` case returns `r.chosen[st.Target]` as it is, so an asset an earlier step returned, put on a deck or killed is still in it.
  - `stepRelocate` skips such an asset itself (`g.Asset(a.ID) != a`); `stepDealDamage`, `stepHeal`, `stepAddArmor`, `stepModifyStat` and `stepKill` do not.
  - The other subjects already read the battlefield as it is: `SubjectSelf` and `SubjectReplaced` check `g.Asset`, and `SubjectAll` and `SubjectPick` look at the battlefield when the step runs.

### Shape (lead decisions, binding)
- **One check, in `stepAssets` (`effects.go`):** the `SubjectTarget` case returns only the slot's chosen assets still on the battlefield (`g.Asset(a.ID) == a`), in pick order. It does not change `r.chosen`. Every asset step reads targets through `stepAssets`, so nothing else checks.
- **`stepRelocate` loses its own check** (`if g.Asset(a.ID) != a { continue }`) and its doc comment's last sentence; the check is now `stepAssets`'.
- **Doc comments:** `stepAssets`' says the chosen entities are the ones still legal and still on the battlefield, so a target an earlier step or mode removed is skipped (`FR-ZON-004`, `FR-TGT-003` partial invalidation).
- **`replaceDeaths` keeps its checks** (`replace.go:79`, `replace.go:88`). They guard its own loop, where one replacement can remove another dying asset, and the death sweep calls it too. `replace.go` is not touched.
- **Counting:** a step counts only what it acted on, so a step whose whole slot was removed counts 0 and an `IfDone` step after it does not run (`SVY-R14`). This falls out of the check.
- **Pile slots:** out of scope. `stepCards` already re-reads each card's zone through `takeCard`.
- **Sims:** no executable card reaches the case. *Infested Apostle* kills an allied pick and damages an opposing target, which are never the same asset. So `go tool task sim-compare -- autoloop/m2-wave-g.2` should pass byte for byte. If it differs, don't change code to make it match: report the first differing game in the Log and hand off.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
Synthetic cards only, in a new file `pkg/engine/removed_target_test.go`. Reuse the scenario helpers in `scenario_test.go`, and the helpers in `relocate_test.go`, `replacement_test.go` or `if_you_do_test.go` where they fit. Every existing test passes unchanged.
- `TestFR_ZON_004_ReturnedTargetTakesNoLaterDamage` (directive #82 (a)): a triggered ability with one target slot for an enemy unit; step 1 is `OpReturnToHand` on that slot, step 2 deals 2 damage to the same slot. The chosen unit has a damage-taken trigger. After resolution its card is in its owner's hand, no `damage` event names it, and its damage-taken trigger did not fire.
- `TestFR_ZON_004_KilledTargetGetsNoLaterModifier` (directive #82 (b)): step 1 kills the chosen unit, step 2 gives the same slot +2 power, step 3 is an `IfDone` step with a visible result (for example the controller draws a card). Step 2 counts 0, so step 3 does not run.
- `TestFR_TGT_003_RemovedTargetLeavesTheRestOfItsSlot` (directive #82 (c)): slot 0 is "up to 2" units, slot 1 exactly one unit. The same unit X is chosen for slot 1 and, with another unit Y, for slot 0. Step 1 returns slot 1's unit to its owner's hand, step 2 deals 1 damage to slot 0. Y takes 1 damage, and no `damage` event names X.
- `TestFR_ZON_004_ReplacedKillLeavesNoTarget` (directive #82 (d)): a test-only destruction replacement returns the dying unit to its owner's hand (`SubjectReplaced`). Step 1 kills the chosen unit, whose death is replaced; step 2 deals 1 damage to the same slot. Nothing is damaged, and the unit's card is in its owner's hand.

Directive #82's acceptance (every existing test passes unchanged, `sim-compare`) is checked by the gates and the review, not by new tests.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected. The test file may declare its own helpers.

## Allowed Files
- `pkg/engine/effects.go` (`stepAssets`' `SubjectTarget` case and doc comment; `stepRelocate`'s check and doc comment)

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run 'TestFR_ZON_004|TestFR_TGT_003|TestFR_REP_001|TestCard_'`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-compare -- autoloop/m2-wave-g.2` passes (no executable card reaches the case; see Shape, Sims)
- [ ] `stepRelocate` has no `g.Asset(a.ID) != a` check of its own, and the only `g.Asset(...)` checks on a step's assets in `effects.go` are in `stepAssets`.

## Log
- 2026-10-30 lead-architect (wave G.2 run, session 1): Wrote the task for board directive #82 (`open: before the next Deliverables task`, design note Deliverable 6e's first task). Run file created; `main` (281e7d8) has nothing the run branch lacks, so no sync.
  - **Directives:** read `BOARD.md` on `autoloop/m2-wave-g.2` fresh. #82 is applied by this task; #83, #84, #85, #86 and #87 (c) follow it in board order, all `open: before the next Deliverables task`. #14 is a standing order. Waiting triggers checked: #6 and #7 (mode is `serial`); #8 (no `deck.go` change); #15 (no turn hook); #20 (b) (no wave C shape task); #58 (no exhaust by a cost or an effect); #59 (2) (no "for each" repeated targets: test (c) chooses one unit in two different slots, which `chooseTargets` already allows, slot by slot); #61 (no modal card); #71 (no effect on each player). None has fired.
  - **Reuse:** this task removes a duplicate rather than adding one: the "still on the battlefield" check for targets moves from `stepRelocate` into `stepAssets`, where every asset step reads it.
- 2026-10-30 rules-test-author: Wrote `pkg/engine/removed_target_test.go` (4 tests, synthetic cards, own helper `rtStep`; reuses `iydDeploy`, `rlStep`, `damageTo`, `mortal`). Tests: `TestFR_ZON_004_ReturnedTargetTakesNoLaterDamage` (a), `TestFR_ZON_004_KilledTargetGetsNoLaterModifier` (b), `TestFR_TGT_003_RemovedTargetLeavesTheRestOfItsSlot` (c; scripted chooser picks X and Y for the up-to-2 slot and X for the exact slot, told apart by `Decision.Max`), `TestFR_ZON_004_ReplacedKillLeavesNoTarget` (d; a structure with a test-only destruction replacement returning the dying ally unit to hand).
  - **Red before implementation:** all four fail now (later step still damages or modifies the removed unit; (a) also fires its damage-taken trigger). None already passes.
  - **Helpers requested:** none. Cases left out: none.
- 2026-10-30 lead-architect (tests review, round 1): Approved, `ready` for `kernel-engineer` (sonnet). The four tests match directive #82 (a)–(d) and the user's 2026-10-30 ruling: each later step skips the removed unit and still acts on the rest of its slot ((c): Y takes 1), a step whose whole slot left counts 0 so its `IfDone` follower does not run ((b), `SVY-R14`), and a replaced kill leaves no target ((d)). They assert outcomes only (events, zones, deck counts). Re-run in a review worktree at 6786c54: all four fail before the implementation. Engineer: follow the Shape exactly; `replace.go` is not touched.
- 2026-10-30 kernel-engineer: Implemented. Only `pkg/engine/effects.go` changed: `stepAssets`' `SubjectTarget` case now returns only chosen assets with `g.Asset(a.ID) == a` (pick order, `r.chosen` untouched) and its doc comment says so; `stepRelocate` lost its own `g.Asset(a.ID) != a` check and the doc sentence. `replace.go` untouched.
  - **Reuse:** no new shape; the check moved from `stepRelocate` into `stepAssets`.
  - **Gates:** `go test ./pkg/engine -run 'TestFR_ZON_004|TestFR_TGT_003|TestFR_REP_001|TestCard_'` ok; `go tool task check` passed; `go tool task sim` clean; `go tool task sim-compare -- autoloop/m2-wave-g.2` PASS byte for byte. `grep g.Asset( effects.go`: remaining checks are `resolveAbility` target re-verify, `value`, `itemSource`, and `stepAssets` (target, replaced, self); none in `stepRelocate`.
