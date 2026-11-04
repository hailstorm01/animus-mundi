---
id: T-0158
title: Activate a printed ability of a battlefield asset, its costs paid and its targets chosen on declaration (FR-ACT-001, FR-ACT-002, FR-TRN-003a, SVY-R2, wave D shape, sonnet)
kind: shape
tier: sonnet
wave: D
status: tests-review
branch: m2/autoloop-wave-d/T-0158-activated-abilities
depends_on: [T-0144, T-0145]
claims: [action:activate, def:activated, cost:resources, cost:exhaust-self, cost:kill-self]
cards: []
requirements: [FR-ACT-001, FR-ACT-002, FR-TRN-003a, FR-TGT-001, FR-QUE-001]
---

## Goal
A printed `[Cost]: [Effect].` ability of a battlefield asset can be activated in its controller's Main Phase on an empty stack: its costs are paid and its targets chosen on declaration, and it resolves as a stack item. This is the first wave D shape. All four of the wave's cards build on it. No card uses it yet.

## Inputs
- FRD 01:
  - `FR-ACT-001`, `FR-ACT-002` (lines 334-336): the empty-stack requirement, costs paid on declaration, targets chosen at the same time, and activating while Exhausted when the cost has no `exhaust`.
  - `FR-TRN-003a` (line 341): triggers from paying a cost go on top of the declared ability. A replaced kill still counts as paid (`SVY-R15`).
  - `FR-TGT-001` (`SVY-R2`, line 359): targets are judged before the cost moves anything. `FR-TGT-003` (line 376 on): two-stage verification, and "when an ability on a battlefield Asset fizzles, the Asset remains on the battlefield".
  - `FR-KWD-002h` (Ward): opponents can't choose the asset as a target of an activated ability.
  - `FR-TRN-002` (line 288): non-active players can't activate.
  - `FR-ZON-003`: the controller, not the owner, acts.
- Survey rows: wave D (`docs/design/m2_starter_set_ability_survey.md`, §2, gap G3, R2). No card in this task.
- Rulings that apply: `SVY-R2`, `SVY-R3`, `SVY-R15`, `SVY-R17`.
- Design note: `docs/architecture/m2-ability-model.md`, Key types (`Activated`: `Costs` plus an `Effect`; `Cost`: resources, exhaust self, kill self, survey G3) and Changes in `engine` (Actions: `ActActivate`, asset plus ability index).
- Kernel today:
  - `declare` (`placement.go:39`) is the one declare sequence: discard picks, modes and targets, push, costs. Structural review 7 (#92) made it for this use.
  - `costsPayable` and `payCosts` (`costs.go`) walk `placement.costs`, and every kind but `CostDiscard` picks an asset through `costDecision`.
  - `kill` (`sweep.go:31`) goes through replacements and `exhaust` (`actions.go:119`) fires the exhausted hook; both only buffer their triggers.
  - `StackAbility` already resolves any `Effect`, with `Source` and `Inst` (`stack.go`), and survives its source leaving (`resolveAbility`).
  - `Asset.Printed` gives the Scratch Company face for a face-down asset, whose `Ability` is nil.

### Shape (lead decisions, binding)
- **`ability.Activated`** (`ability/ability.go`): `Costs []Cost` in printed order, and `Effect Effect`. **`Def.Activated []Activated`**, the printed activated abilities in printed order.
- **Three new `CostKind`s, after `CostDiscard`,** so the existing kinds keep their numbers:
  - `CostResources`: pays `Cost.Req` Requisition and `Cost.Tac` Tactics. New fields on `Cost`, read by no other kind.
  - `CostExhaustSelf`: "exhaust". Exhausts the ability's source.
  - `CostKillSelf`: "kill this unit" (survey G3). Kills the ability's source through `kill`.
  Update `Cost`'s comment, and `CostKind`'s if it needs it, so it says a cost is paid when a card or an ability is declared.
- **`costsPayable` and `payCosts`** (`costs.go`) handle the new kinds, and `pl.source` is the asset. `CostResources` is payable when the pool covers it. `CostExhaustSelf` is payable when the source is Readied. `CostKillSelf` is always payable. Payment follows the kinds' printed order, as for the others. The triggers either self cost fires wait in the pending buffer to go above the item (`FR-TRN-003a`). `costDecision` is still never called for a kind that takes no chosen asset, so correct its comment. No second kill or exhaust path.
- **`activatedPlacement(pid PlayerID, a *Asset, act ability.Activated) placement`** (`placement.go`, beside `maneuverPlacement`): controller `pid`, source `a.ID`, inst `a.Inst`, label the printed name, costs `act.Costs`. Correct `placement`'s type comment. Legality and declaring both build it with this function.
- **`ActActivate`** (`actions.go`): a new `ActionKind` **between `ActEnact` and `ActEndTurn`**, with `String()` giving `"Activate"`. `ActEndTurn` stays last, for `pkg/eventlog`'s replay loop. New field **`Action.Ability int`**: the index into the asset's `Def.Activated`. `Action.Asset`'s comment names Activate too.
  - **`check`:** the asset is on the battlefield and controlled by the player, `Printed().Ability` is non-nil and the index is in range, and `costsPayable` holds. Readiness is not checked here, only through `CostExhaustSelf` (`FR-ACT-002`). The existing Main Phase, active-player and empty-stack checks apply.
  - **`Apply`:** `g.declare(activatedPlacement(...), act.Effect, &StackItem{Kind: StackAbility, Controller: a.Player, Source: as.ID, Inst: as.Inst})`, then resolve the stack as every action does. Costs are paid inside `declare`, after the push and after the targets, so `SVY-R2` holds with no other code. No new stack kind and no new decision kind.
  - **`LegalActions`:** one `ActActivate` for each of the player's assets and each printed ability that `check` accepts, after the Enact loop and before Scratch.
- **Not this task: an exactly-one target slot with no legal entity.** FRD 01 settles this for a Maneuver (`SVY-R29`) and a trigger or deploy (`FR-TGT-003`), not for an activated ability. Board directive #104 asks the user before the first card task with an activated ability. Here, `check` calls `costsPayable` only, not `placeable`, and **no test may pin either side of it:** every test's exactly-one slot has a legal entity on declaration.
- **Clients (T-0159, `queued`):** `emitAction`, replay, the sim's `ProgressPolicy`, the play screen's label and the monitors. **Comments (T-0160, `queued`):** `StackAbility`'s comment (`stack.go`) and `doc.go`'s not-built list. No executable card has an activated ability until the wave's first card task, and both come before it.
- **Sims:** no card gains a `Def.Activated`, so `LegalActions` doesn't change for any existing card. `go tool task sim-compare -- autoloop/m2-wave-d` should pass byte for byte. If it differs, don't change code to make it match: report the first differing game in the Log and hand off.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
Synthetic cards only (a `Unit` or `Structure` with an `ability.Def` whose `Activated` the test builds, put on the battlefield through the scene helpers), in a new file `pkg/engine/activated_test.go`. Every existing test passes unchanged, except the one edit named below.
- `TestFR_ACT_001_CostsPaidAndEffectResolves`: an asset with "R1, T1, exhaust: draw a card", and R1 T1 in the pool. After `ActActivate`: the pool is R0 T0, the asset is Exhausted, and the player drew one card.
- `TestFR_QUE_001_ActivatedAbilityIsPlacedOnTheStack`: a `place` event naming the card, at stack depth 1, comes before the events of the ability's steps. Use a step that logs one, such as damage to the opposing Headquarters. A successful draw logs none.
- `TestFR_ACT_001_NeedsAnEmptyStack`: `ActActivate` is rejected (`ErrIllegal`) with an item on the stack, whatever the costs.
- `TestFR_ACT_001_OnlyTheActiveControllerActivates`: rejected on the opponent's turn (`FR-TRN-002`), rejected for an asset the player doesn't control, and allowed for an asset the player controls but doesn't own (`FR-ZON-003`).
- `TestFR_ACT_001_RejectsWhatHasNoSuchAbility`: rejected for an index out of range, for an asset with no activated ability, and for a face-down Scratch Company asset whose card has one (`Printed`).
- `TestFR_ACT_001_UnpayableIsNotOffered`: with too little Requisition, too little Tactics, or an Exhausted asset with an `exhaust` cost: `LegalActions` has no such `ActActivate`, `Apply` returns `ErrIllegal`, and nothing was spent. Once it can be paid, it is offered.
- `TestFR_ACT_001_OfferedOncePerPrintedAbility`: an asset with two abilities offers both, and each index runs its own effect.
- `TestFR_ACT_002_NoExhaustCostWorksWhileExhausted`: an Exhausted asset with "R1: add 1 Armor to this asset" is offered and activates, and stays Exhausted.
- `TestFR_ACT_002_NoExhaustCostRepeats`: with R2, the same ability activates twice in one turn for 2 Armor, and a third time is rejected for lack of Requisition. An ability with an `exhaust` cost is rejected the second time.
- `TestFR_ACT_001_KillSelfCostPaidOnDeclaration`: "kill this unit: draw a card": the source's card is in its owner's discard pile, and its effect still resolves after it left.
- `TestFR_ACT_001_SacrificedCardIsNotAChoice` (`SVY-R2`): "kill this unit: return chosen unit card from your discard pile to your hand", with one other unit card in the pile. The `DecideTarget` lists only that card or, being a single option, is taken unasked. That card ends in hand, and the source's card in the pile. Add a case with two others in the pile: the options are those two, not the source.
- `TestFR_TRN_003a_ActivatedCostTriggersResolveFirst`: the source has a death trigger and "kill this unit: deal 1 damage to the opposing Headquarters". The ability is placed at depth 1 and the trigger at depth 2, and the trigger's event comes first. Add a case for `exhaust` with a `HookExhausted` trigger (`SVY-R17`).
- `TestFR_TRN_003a_ReplacedKillSelfCostStaysPaid` (`SVY-R15`): a synthetic destruction replacement keeps the source from being destroyed. The ability still resolves, and the source's death trigger does not fire.
- `TestFR_TGT_001_ActivatedTargetChosenOnDeclaration`: "T1: kill chosen opposing unit" with two opposing units. One `DecideTarget` lists both. The scripted answer's unit dies and the other survives.
- `TestFR_KWD_002h_WardBlocksActivatedTarget`: an opposing Ward unit is not offered as that target, while another opposing unit is.
- `TestFR_TGT_003_FizzledAbilityLeavesSourceOnBattlefield`: "exhaust: kill chosen opposing unit" on a source with a `HookExhausted` trigger that removes that unit first. The ability does nothing, and the source is still on the battlefield.
- **Edit to an existing test** (`pkg/engine/events_test.go`, `TestEvents_Case4_EveryActionKind`, in the `test(T-0158)` commit): the loop's skip reads `if k == ActEnact || k == ActActivate`. The comment adds that no starter card has an executable activated ability yet, and that T-0159 adds Activate's own action-event test. Nothing else in the test changes.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected. The test file may declare its own, for example `activatedUnit(name string, acts ...ability.Activated) *Card`.

## Allowed Files
- `pkg/engine/ability/ability.go` (`Activated`, `Def.Activated`, the cost kinds, `Cost.Req`, `Cost.Tac`)
- `pkg/engine/actions.go` (`ActActivate`, `String`, `Action.Ability`, `check`, `Apply`, `LegalActions`)
- `pkg/engine/costs.go` (`costDecision`'s comment, `costsPayable`, `payCosts`)
- `pkg/engine/placement.go` (`activatedPlacement`, `placement`'s comment)

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run 'TestFR_ACT_00|TestFR_QUE_001_Activated|TestFR_TRN_003a_(Activated|ReplacedKillSelf)|TestFR_TGT_001_ActivatedTarget|TestFR_KWD_002h_WardBlocksActivated|TestFR_TGT_003_FizzledAbility|TestEvents_Case4'`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-compare -- autoloop/m2-wave-d` passes (no card gains a `Def.Activated`; see Shape, Sims)
- [ ] `ActEndTurn` is still the last `ActionKind`.

## Log
- 2026-11-04 lead-architect: wave D run, session 1. No `main` commits to sync (`main` and the run branch are both at 58a2bc9). Not `WAVE-COMPLETE`: all four header cards are on `NotYetExecutable`. No `open: before` directive; #14 is `open` and orders the cards (deck cards first, *Rhapsody's Hollow* deferred by (b), *Ruin Partisans* kept by (c)). Waiting triggers #6, #7, #8, #15, #71, #91 (2) not fired. Wave D's cards need, in order: this shape, its clients (T-0159), comments (T-0160), a subtype criterion and a Maneuver kind for pile filters (*Sinuous Growth*, *Boneyard Scavenger*), and `SEARCH_DECK` (*Cliff of Portents*). Reuse search: no activated path exists (`grep Activated|ActActivate|KillSelf|ExhaustSelf`); `declare`, `costsPayable`, `payCosts`, `kill`, `exhaust` and `resolveAbility` are reused, not copied. Four non-test files, so the clients and comments are split off as `queued` tasks (D2.4). Two rules questions go on the board as #104, to ask before the first card task: activating with no legal target, and a deck search's choice and timeout. Neither is answered here. Status `tests`; dispatching `rules-test-author`.
- 2026-11-04 rules-test-author: tests written in `pkg/engine/activated_test.go` (15 test functions, two with subtests): TestFR_ACT_001_CostsPaidAndEffectResolves, TestFR_QUE_001_ActivatedAbilityIsPlacedOnTheStack, TestFR_ACT_001_NeedsAnEmptyStack, TestFR_ACT_001_OnlyTheActiveControllerActivates, TestFR_ACT_001_RejectsWhatHasNoSuchAbility, TestFR_ACT_001_UnpayableIsNotOffered, TestFR_ACT_001_OfferedOncePerPrintedAbility, TestFR_ACT_002_NoExhaustCostWorksWhileExhausted, TestFR_ACT_002_NoExhaustCostRepeats, TestFR_ACT_001_KillSelfCostPaidOnDeclaration, TestFR_ACT_001_SacrificedCardIsNotAChoice, TestFR_TRN_003a_ActivatedCostTriggersResolveFirst, TestFR_TRN_003a_ReplacedKillSelfCostStaysPaid, TestFR_TGT_001_ActivatedTargetChosenOnDeclaration, TestFR_KWD_002h_WardBlocksActivatedTarget, TestFR_TGT_003_FizzledAbilityLeavesSourceOnBattlefield. Edit: `TestEvents_Case4_EveryActionKind` skips `ActEnact || ActActivate` with the comment. Helpers requested: none (the file declares its own: actUnit, actAbility, activate, offersActivate, scene.setPool and others). Left out: nothing from the case list; no test pins an exactly-one slot with no legal entity (#104). `go vet` fails only on the missing `ability.CostExhaustSelf` surface; `cardtown lint rules` passes. Notes for the engineer: tests assume `place` events of an activated ability carry the source's card name with `Amount` = stack depth, as for triggers; the single-option target cases accept either asked or unasked.
