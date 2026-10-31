---
id: T-0116
title: Enact a Maneuver from hand, resolve its effect and put it into its owner's discard pile (FR-MNV-001, FR-TRN-003, directive #88 part 1, wave C shape, sonnet)
kind: shape
tier: sonnet
wave: C
status: queued
branch: m2/T-0116-enact-maneuver
depends_on: [T-0083, T-0107]
claims: [action:enact, stack:maneuver, def:effect]
cards: []
requirements: [FR-MNV-001, FR-TRN-003, FR-ZON-001, FR-QUE-001]
---

## Goal
A Maneuver whose definition has an effect can be enacted from hand: its resource costs are paid, it goes on the Resolution Stack with its modes and targets chosen on declaration, its steps resolve under two-stage verification, and the card then goes to its owner's discard pile. This is the first wave C shape. Every wave C card is a Maneuver, so all of them build on it.

## Inputs
- FRD 01:
  - `FR-TRN-003`/`FR-MNV-001` (line 333): enacting from Hand needs an empty stack. On declaration the card moves to the Resolution Stack zone and costs are paid. When it resolves, its effects execute and the card goes to its owner's Discard pile.
  - `FR-MNV-001` (line 334, `SVY-R29`): legality on declaration. **Not this task** (see Shape, Split).
  - `FR-TGT-001`, `FR-TGT-003` (lines 354-373): targets are locked on placement, then re-verified at resolution.
  - `FR-MOD-001` §5 (line 384): modes chosen on placement, before targets, and resolved in printed order.
  - `FR-QUE-001`: the stack.
- Survey rows: wave C (`docs/design/m2_starter_set_ability_survey.md`, §2 and lines 125-176). No card in this task.
- Rulings that apply: `SVY-R3` (targets locked on placement), `SVY-R25`, `SVY-R31`. `SVY-R29` is the next task's.
- Design note: `docs/architecture/m2-ability-model.md`, Key types (`Def`: "`Effect` (Maneuvers)"), Changes in `engine` ("`Supported()` accepts Maneuvers that have an `Effect`"; Actions: `ActEnact`; Stack: "Maneuvers resolve an `Effect` the same way").
- Board directive #88, first bullet (enacting a Maneuver), split under the D2 size rule (see Shape, Split).
- Kernel today:
  - `Card.Supported` (`card.go:116`) rejects every Maneuver.
  - `Action` has no enact kind (`actions.go`).
  - Placing an effect is `chooseOnPlacement(pl placement, e)` (`placement.go`): modes, then targets. A fired trigger builds its `placement` in `fireHook`, and `flushTriggers` pushes a `StackAbility` item (`trigger.go:189-193`).
  - `resolveAbility` (`effects.go`) re-verifies the chosen targets and runs the steps.
  - The view, the play screen, the replay and the sim policy don't know the new kind yet. That is a later harness task (see Shape, Clients).

### Shape (lead decisions, binding)
- **`ability.Def.Effect Effect`** (`ability/ability.go`): a Maneuver's effect, resolved when the Maneuver resolves (`FR-MNV-001`). It is zero for every other card type. It is a value, not a pointer, like `Trigger.Effect`.
- **`Card.Supported`** (`card.go`): a Maneuver is supported when it has no unsupported lines, and its `Ability` is non-nil with an `Effect` that has steps or modes. Every other type keeps today's rule. `TestLoadCards_ManeuverStaysUnsupported` (a Def with `Text` and no `Effect`) keeps passing unchanged.
- **`ActEnact`** (`actions.go`): a new `ActionKind`, declared **between `ActScratch` and `ActEndTurn`**, with `String()` giving `"Enact"`. That keeps `ActEndTurn` last, so `pkg/eventlog`'s `k <= engine.ActEndTurn` replay loop still covers every kind. Action kinds are recorded by name (`Detail`), never by number. Its field is `Action.Card`: the `Instance.ID` in hand, as for Deploy. Update the field's comment to say so.
  - **`check`:** the same preconditions as Deploy, minus the zone. The card must be in the player's hand and `Supported()`, its `Type` must be `TypeManeuver`, and the player must be able to pay `ReqCost` and `TacCost`. The empty-stack, active-player and Main Phase checks apply as they do now. `ActDeploy` keeps rejecting a Maneuver ("cannot be deployed").
  - **`Apply`:** pay `ReqCost` and `TacCost`, then take the card out of hand (`takeCard`), so the card is on the stack before anything is chosen (`FR-TRN-003`). Then choose on placement with `chooseOnPlacement`, using a `placement` whose controller is the player, whose `inst` is the card, whose `source` is 0 and whose label is the card's name. Push a `StackManeuver` item, and resolve the stack as every action does.
  - **`LegalActions`:** offer `ActEnact` for each card in hand that `check` accepts, after the deploys and before Scratch.
- **No source asset.** A Maneuver isn't an asset, so its item's `Source` is 0 (asset IDs start at 1). Steps that read a source asset are out of scope: `SubjectSelf`, `ValueSourcePower`, `Filter.Other`, and damage attribution (a `DealDamage` step from a Maneuver). Tests use draw, kill, return-to-hand and put-on-deck steps only. No M2 Maneuver reads its own power, and the damage event's attribution for a Maneuver comes with the first damage card (*Despoiler's Emergence*).
- **`StackManeuver`** (`stack.go`): a new `StackKind`, declared **after `StackAbility`** so that the existing kinds keep their numbers. It carries `Inst` (the card), `Effect` and `Chosen`, with `Source` 0. Update the field comments that say "StackAbility:" to name both kinds where both use the field.
  - **`push`:** its `place` event names the card (`Card`, `Inst`), as a deploy's does, and the log line says `Maneuver <name>`.
  - **`resolve`:** `resolveManeuver(it)` calls `resolveAbility(it)` (two-stage verification and the steps, unchanged), then puts the card into its owner's discard pile (`putCard(it.Inst, ZoneDiscard, false)`). The card goes there even when every target fizzled (`FR-MNV-001`, line 334). `effects.go` is not touched: `resolveManeuver`'s doc comment says it reuses `resolveAbility`, whose own comment ("a triggered ability") is corrected by the next task that edits `effects.go`.
  - No new event kind. The card's move from the stack to the discard pile emits nothing in this task (see Clients).
- **Split (D2 size rule).** Directive #88's first bullet also makes a Maneuver legal only if each exactly-one target slot has a legal entity on declaration, and a `ChooseN` Maneuver only if `legalModes` gives at least N (`SVY-R29`). That needs `placement.go`, a fifth non-test file, so it is the next task (#88 tests (a) and (b)), and it shares `legalModes`' slot check rather than copying it. **In this task, no test may pin either side of `SVY-R29`.** Every test Maneuver has a legal entity for each exactly-one slot, and a modal test Maneuver has at least `ChooseN` legal modes.
- **Clients (a later harness task, not this one):** `emitAction` recording the enacted card's name and instance, `pkg/view`'s stack-kind label and chosen targets for `StackManeuver`, `viewtest`'s public names, the play screen's `Enact:` label, the sim's `ProgressPolicy` choosing enacts, and a stack-to-discard event if the view needs one. No executable card is a Maneuver until wave C's first card task, so nothing reaches them before then.
- **Sims:** no card gains a `Def.Effect` here, so no Maneuver becomes executable. `go tool task sim-compare -- autoloop/m2-wave-c` should pass byte for byte. If it differs, don't change code to make it match: report the first differing game in the Log and hand off.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
Synthetic Maneuver cards only (`TypeManeuver`, an `ability.Def` with an `Effect`, a hand card through `scene.give`), in a new file `pkg/engine/maneuver_test.go`. Reuse the helpers in `scenario_test.go` and the synthetic-effect helpers of `modal_test.go`, `removed_target_test.go` or `relocate_test.go` where they fit. Every existing test passes unchanged.
- `TestFR_MNV_001_EnactedManeuverResolvesThenGoesToDiscard`: a Maneuver costing R1 T1 whose effect is "draw 2 cards", in the active player's hand with R1 T1 in the pool. After `ActEnact`: the pool is R0 T0, the player drew 2 cards, the card is no longer in hand, and it is the last card of its owner's discard pile.
- `TestFR_MNV_001_EnactIsALegalActionOnlyWhenPayable`: a supported Maneuver in hand appears in `LegalActions` as `ActEnact` with its instance ID while the player can pay its cost, and doesn't once they can't. Applying it then is rejected with `ErrIllegal`, and the card stays in hand.
- `TestFR_MNV_001_UnsupportedOrNonManeuverCannotBeEnacted`: `ActEnact` is rejected (`ErrIllegal`) and absent from `LegalActions` for a Maneuver with no `Def.Effect`, for a supported Unit card in hand, and for a card not in the player's hand. `ActDeploy` of a supported Maneuver is rejected.
- `TestFR_MNV_001_OnlyTheActivePlayerEnacts`: on the opponent's turn, the non-active player's `ActEnact` is rejected (`FR-TRN-002`).
- `TestFR_TGT_001_ManeuverTargetChosenOnDeclaration`: "kill chosen opposing unit" with two opposing units in play. The enacting player is asked one `DecideTarget` decision listing both. The scripted answer picks the second, which dies, and the first survives. The card ends in its owner's discard pile.
- `TestFR_QUE_001_ManeuverIsPlacedOnTheStack`: enacting emits a `place` event naming the card (`Card`, `Inst`) with stack depth 1, before the events of its steps.
- `TestFR_MOD_001_ModalManeuverResolvesOnlyItsChosenMode`: "Choose one" with modes "draw a card" and "return chosen unit to its owner's hand", one unit in play. The scripted answer picks mode 2 and the unit. The unit's card is in its owner's hand, no card was drawn, and the Maneuver is in the discard pile.
- `TestFR_MNV_001_FizzledManeuverStillGoesToDiscard`: "return chosen unit to its owner's hand, then kill the same slot" (`SVY-R25`: the kill step finds nothing). The unit's card is in hand, no `destroy` event is emitted, and the Maneuver is in its owner's discard pile.
- `TestCard_SupportedManeuverNeedsAnEffect` (in `card_test.go` or the new file): `Card.Supported()` is true for a Maneuver with no unsupported lines and a `Def.Effect` with steps (or with modes), and false with a nil `Ability`, with an empty `Effect`, or with an unsupported line.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected. The test file may declare its own helpers, for example `maneuver(name string, req, tac int, e ability.Effect) *Card`.

## Allowed Files
- `pkg/engine/ability/ability.go` (`Def.Effect`)
- `pkg/engine/card.go` (`Supported`)
- `pkg/engine/actions.go` (`ActEnact`: kind, `String`, `Action.Card` comment, `check`, `Apply`, `LegalActions`)
- `pkg/engine/stack.go` (`StackManeuver`, field comments, `push`, `resolve`, `resolveManeuver`)

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run 'TestFR_MNV_001|TestFR_TGT_001_ManeuverTarget|TestFR_QUE_001_Maneuver|TestFR_MOD_001_ModalManeuver|TestCard_SupportedManeuver|TestLoadCards_'`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-compare -- autoloop/m2-wave-c` passes (no card gains a `Def.Effect`; see Shape, Sims)
- [ ] `ActEndTurn` is still the last `ActionKind`, and `StackAbility` keeps its value.

## Log
- 2026-10-30 lead-architect: wave C run, session 1. Synced `main` (FRD 01 v3.19, `17ca104`) into `autoloop/m2-wave-c` (`5c155a0`). Directive #88's trigger has fired (the outcome's `Status:` is `applied: FRD 01 v3.19`), so it is `open`. Its first bullet is split under the D2 size rule: this task builds the enact action and its resolution (four non-test files). The next task builds `SVY-R29` legality (#88 tests (a), (b)) in `placement.go` and `actions.go`, sharing `legalModes`' exactly-one-slot check. After it come the delegated dealer (#88 tests (c), (d); it edits `effects.go`, so its task also corrects `resolveAbility`'s "triggered ability" doc comment), `MOVE_ZONE` (#88 test (e)), the client harness task (see Shape, Clients) and the card tasks, deck cards first (#14 (a)). Directive #20 (b) was triggered by this task, the first wave C shape task. It applies to the task for *Steel Rain*'s discard cost, so it stays `open` until then. Duplicates searched: there is no other enact path, and `chooseOnPlacement` and `resolveAbility` are reused, not copied. Status `tests`; dispatching `rules-test-author`.
