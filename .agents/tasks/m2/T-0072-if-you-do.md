---
id: T-0072
title: "You may X. If you do, Y": an untargeted pick on resolution, a kill step, and steps that report what they did (SVY-R14, directive #51 part 2, wave B shape, sonnet)
kind: shape
tier: sonnet
wave: B
status: tests-review
branch: m2/T-0072-if-you-do
depends_on: [T-0071]
claims: [primitive:KILL, subject:pick, gate:if-done, convention:step-counts]
cards: []
requirements: [FR-TGT-001, FR-TRG-005, FR-LAY-002, FR-MOD-001]
---

## Goal
A triggered ability can say "you may X. If you do, Y" (`SVY-R14`). X is a step whose entities are picked, untargeted, as it resolves (`ability.SubjectPick`). The first such X is a kill (`ability.OpKill`). Y is a step gated on X (`Step.IfDone`): it runs only if X did at least one thing. To know that, every `step<Op>` returns how many things it did, and the per-resolution struct records it (board directive #51 part 2). The first card, Infested Apostle ("When this unit deploys, you may kill an allied unit. If you do, deal 2 damage to chosen opposing unit."), is T-0073.

## Inputs
- FRD 01:
  - `FR-TGT-001` "Untargeted Selections (`SVY-R3`)" (line 326): no "chosen" means no target. The pick is made as the clause resolves, from everything that matches then. Ward and `FR-TGT-003` don't apply. With no match, only that clause does nothing.
  - `FR-TGT-001` "You May X. If You Do, Y (`SVY-R14`)" (line 327): "you may" and Y's targets are settled on placement. X is carried out on resolution, as an untargeted selection. Y applies only if X was carried out. No match or a timeout on X means no Y. The source may select itself. Assets X destroys have their death triggers placed after the trigger finishes, so Y resolves first.
  - `FR-MOD-001` §6 (line 356): mandatory targets and untargeted selections time out to the first legal options in fixed order. "The selection for X ... is optional and also selects none, so X is not carried out."
  - `FR-LAY-002` (line 425): direct destruction ("kill") happens inline, as the effect says, not at the death sweep. `FR-ZON-002` Destruction (line 186): a killed asset goes face-up to its owner's discard pile.
  - `FR-TRG-005` (line 387): the X action is not a cost.
  - `SVY-R15` (`FR-REP-001`, line 444): a replaced X still counts as done. No kill replacement exists yet (Shifting Impossibility is later). This task's count is "the step carried out its action on N entities", which is the number SVY-R15 needs.
- Survey rows: Infested Apostle (line 109), Arkha, Sun's Fury (line 150). Rulings `SVY-R3`, `SVY-R14`, `SVY-R15`, `SVY-R16` (survey lines 61, 73-75).
- Design note: `docs/architecture/m2-ability-model.md`, Key types (`Trigger`, untargeted pick, `Step`, `Value`), Effect ops, and Open Question 8(a).
- Board directives: #51 part 2 (this task); #20 (a) (an untargeted pick is a resolution-time `Chooser` choice, not a `Target`, and may take a count); #47 (1).
- Kernel today:
  - `resolveAbility` (`effects.go`) builds `r := &resolution{it, chosen}` and dispatches each step to `(g *Game) step<Op>(r, st)`, which returns nothing (T-0071).
  - `stepAssets` resolves `SubjectTarget`, `SubjectSelf` and `SubjectAll`.
  - `deathSweep` (`damage.go`) destroys each dying asset inline: `fireKilled`, `leaveBattlefield(a, known)`, `putCard(a.Inst, ZoneDiscard, false)` and a `destroy` event.
  - `chooseTargets` (`targeting.go`) lists options oldest first (`assetsByID`), and `DecideTarget`'s fallback is `fixedOrderPicks` (`choice.go`), which picks none when `Min` is 0.

### Shape (lead decisions, binding)
- **`ability.OpKill`**, appended after `OpModifyStat`, with the comment `// destroy the Subject at once, into its owner's discard pile (FR-LAY-002, FR-ZON-002); a Headquarters is skipped (no M2 card kills one)`. It works with any asset Subject (`SubjectTarget`, `SubjectSelf`, `SubjectAll`, `SubjectPick`), so the wave C kills (Devilhorn Sanction) reuse it.
- **`ability.SubjectPick`**, appended after `SubjectAll`, with the comment `// battlefield assets matching Step.Filter, picked by the ability's controller as the step resolves: one, or with Step.UpTo up to N, possibly none. Not a target (SVY-R3, SVY-R14)`.
- **`Step.UpTo int`**: `// SubjectPick: 0 picks exactly one; N > 0 picks up to N distinct assets, possibly none`. It mirrors `Target.UpTo`.
- **`Step.IfDone bool`**: `// runs only if the step just before it did at least one thing ("if you do", SVY-R14, SVY-R15)`.
- **Step counts (directive #51 part 2):** every `step<Op>` returns an `int`: how many things it did. `resolution` gains `done []int`, one entry per step, and `resolveAbility` records each result. A step with `IfDone` is skipped, and records 0, when it is the first step or the step before it recorded 0. What counts:
  - asset steps (`DealDamage`, `Heal`, `AddArmor`, `ModifyStat`, `Kill`): the assets it acted on;
  - `Draw` and `Scratch`: the cards that left the step players' decks. Count them in `effects.go`, for example from deck sizes before and after. Don't change `zones.go`;
  - `Discard`: the cards discarded;
  - `ReturnToHand` and `PutOnDeck`: 1 if the card moved, else 0. `relocateSelf` may return a `bool` or an `int`.
- **The pick:** `stepAssets` handles `SubjectPick`. It asks one `Decision` each time a step with that Subject resolves:
  - `Kind: DecideTarget` (`FR-MOD-001` §6 gives untargeted selections the same fallback as mandatory targets, so there is no new kind and `choice.go` is unchanged).
  - `Player`: the item's controller.
  - `Options`: every battlefield asset for which `matchesFilter` holds (not `legalTarget`: Ward doesn't apply), oldest first (`assetsByID`), with `Label` `a.String()` and `Ref` `a.ID`.
  - `Prompt`: `"<card name>: pick one"` or, for `UpTo` N, `"<card name>: pick up to N"`, with the name from `cardName(r.it.Inst)`.
  - `UpTo` 0: `Min = Max = 1`. One option takes the fast path; a timeout picks the first in fixed order. `UpTo` N: `Min = 0`, `Max = min(N, len(Options))`. It is always asked, and a timeout picks none.
  - With no options, nothing is asked, no decision event is emitted, and the step acts on nothing.
- **One destroy, not two:** move the per-asset body of `deathSweep`'s destroy loop into one `(g *Game) destroy(a *Asset, known EffectiveState)` in `damage.go`. The body is `fireKilled`, `leaveBattlefield`, `putCard` into discard, and the `destroy` event, in today's order. `deathSweep` calls it with each `known[i]`. `stepKill` calls it with `g.effectiveState(a)`, which is the lone-departure case `leaveBattlefield`'s comment describes. The Headquarters loss check stays in `deathSweep`, which is why `stepKill` skips a Headquarters. A killed face-down Scratch Company reveals into the discard pile exactly as one destroyed by damage does.
- **`doc.go`:** add `OpKill`, `SubjectPick` and the `IfDone` gate to "What runs now".
- **Design note:** the lead updates Key types and Effect ops in the commit that writes this task. Engineers don't edit it.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
To be written in `pkg/engine/if_you_do_test.go`, with synthetic cards only (the Apostle is T-0073). The file won't compile until `OpKill`, `SubjectPick`, `Step.UpTo` and `Step.IfDone` exist. A convenient source is a unit with an optional deploy trigger, `Optional: true`, whose steps are under test. Required cases:
- **Pick on resolution (`SVY-R3`):** `TestFR_TGT_001_UntargetedPickIsMadeOnResolutionFromWhatMatchesThen`. The "you may" question and Y's target are asked on placement and the pick is asked later, as X resolves. Its options are the assets matching then: an allied unit that has left by resolution isn't offered.
- **Ward doesn't apply:** `TestFR_TGT_001_UntargetedPickIgnoresWard`. A pick over opposing units offers an opposing Ward unit.
- **The source may pick itself (`SVY-R14`):** `TestFR_TGT_001_UntargetedPickMayPickTheSource`. With `Filter.Other` false, the source is an option. Killing it still resolves Y, and Y's damage event names the source.
- **Fast path and timeout (`FR-MOD-001` §1, §6; `SVY-R16`):** `TestFR_MOD_001_SinglePickWithOneMatchIsNotAsked`, and `TestFR_MOD_001_SinglePickTimeoutPicksFirstInFixedOrder`. Choose the setup so the oldest asset is not first in fixed order, for example the older unit is in the Main Theater and the newer one in Reserves.
- **Up-to timeout:** `TestFR_MOD_001_UpToPickTimeoutPicksNone`. A step with `UpTo` 1 is asked even with one option, and a timeout picks nothing.
- **No match:** `TestFR_TGT_001_UntargetedPickWithNoMatchDoesNothing`. Nothing is asked, and the ability still resolves.
- **Kill (`FR-LAY-002`):** `TestFR_LAY_002_KillDestroysAtOnceDuringResolution`. The killed unit is in its owner's discard pile before the next step of the same effect: an "all" step right after it doesn't reach it. One `destroy` event names it.
- **Kill a Scratch Company:** `TestFR_LAY_002_KilledScratchCompanyGoesFaceUpToItsOwnersDiscard`.
- **Kill skips a Headquarters:** `TestFR_LAY_002_KillSkipsAHeadquarters`. A kill step whose Subject includes a Headquarters (for example an "all" step over allied assets) leaves the Headquarters on the battlefield and kills the rest. This pins the documented M2 limit, not a rule.
- **If you do (`SVY-R14`):** `TestFR_TRG_005_IfYouDoRunsYWhenXWasDone` (X kills, so Y deals its damage to the target chosen on placement), `TestFR_TRG_005_IfYouDoSkipsYWhenXPickedNothing` (an `UpTo` pick with a timeout), and `TestFR_TRG_005_IfYouDoSkipsYWhenNothingMatchesX`.
- **Death triggers after Y (`FR-TRG-004`):** `TestFR_TRG_004_DeathTriggersOfAnXKillArePlacedAfterTheTriggerFinishes`. A unit with a death trigger is killed by X. Y's damage happens first, and the death trigger is placed and resolved after the item.
- **Counts beyond kill:** `TestFR_TRG_005_IfYouDoFollowsADiscardCount`. With X a discard (`OpDiscard`), Y runs with a card in hand and is skipped with an empty hand.
- **Existing tests:** every test in `pkg/engine` passes unchanged.
- **Left out unless FRD 01 fixes it:** an `IfDone` step that is not the step right after X, a replaced kill (`SVY-R15`, with Shifting Impossibility), and Arkha's pick from a pile (directive #59 (1)).

## Helpers Needed
None expected. `put`, `putScratch`, `give`, `script`, `asked`, `alive` and `endTurn` exist in `scenario_test.go`. If a helper is needed, request it here and the engineer adds it.

## Allowed Files
- `pkg/engine/ability/ability.go` (`OpKill`, `SubjectPick`, `Step.UpTo`, `Step.IfDone`)
- `pkg/engine/effects.go` (`resolution.done`, the gate, step counts, `stepKill`, the pick in `stepAssets`)
- `pkg/engine/damage.go` (`destroy`, called by `deathSweep`)
- `pkg/engine/doc.go` (the "What runs now" lines)

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run 'TestFR_(TGT_001_UntargetedPick|MOD_001_(SinglePick|UpToPick)|LAY_002_Kill|TRG_005_IfYouDo|TRG_004_DeathTriggersOfAnXKill)'`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] No card uses the new shapes yet, so the sims are unchanged: `go tool task sim-compare -- autoloop/m2-wave-b` passes

## Log
- 2026-10-26 lead-architect (wave run B, session 6): Wrote the task. Wave B deck cards still to build are Infested Apostle, Arkha, Sun's Fury and Eternal Watchman (directive #14 (a)). The Apostle and Arkha share this shape, and T-0071's plan puts it next. It carries directive #51 part 2. The Apostle card would make six non-test files, so it is split off as T-0073 (`queued`).
  - Untargeted picks reuse `DecideTarget`. `FR-MOD-001` §6 gives them the same fallback, and `fixedOrderPicks` already picks none at `Min` 0. That also keeps the task at four non-test files (D2.4).
  - The X pick is optional on resolution (`Min` 0 with `UpTo`), as §6 line 356 calls it: "The selection for X ... is optional and also selects none". The Apostle (T-0073) sets `UpTo: 1`.
  - Directive #20 (a) applies here: this is the first untargeted pick. Part (b), Steel Rain, still waits for wave C. #20's trigger was "the first wave C shape task", but `OpKill` and the pick are wave C vocabulary arriving early, so (a) is applied in this task.
  - Waiting triggers checked against `autoloop/m2-wave-b` ee50aaf: #6 and #7 (mode is `serial`), #8 (no `deck.go` change), #15 (no turn hook), #58 (no exhaust by a cost or an effect) and #59 (no pile pick, no repeated targets). None has fired.
  - No existing equivalent: `ability` has no kill op, no picked subject and no step gate. `deathSweep`'s destroy body is the one other copy of a destroy, and the task moves it into `destroy`.
  - Claims don't overlap any open task, since none is open. Status `tests`; dispatching `rules-test-author`.
- 2026-10-26 rules-test-author: Wrote `pkg/engine/if_you_do_test.go`, status `tests-review`. Helpers requested: none (all helpers are file-local `iyd*`/`scene.iyd*` in the test file).
  - Tests: `TestFR_TGT_001_UntargetedPickIsMadeOnResolutionFromWhatMatchesThen`, `..._UntargetedPickIgnoresWard`, `..._UntargetedPickMayPickTheSource`, `..._UntargetedPickWithNoMatchDoesNothing`; `TestFR_MOD_001_SinglePickWithOneMatchIsNotAsked`, `..._SinglePickTimeoutPicksFirstInFixedOrder`, `..._UpToPickTimeoutPicksNone`; `TestFR_LAY_002_KillDestroysAtOnceDuringResolution`, `..._KilledScratchCompanyGoesFaceUpToItsOwnersDiscard`, `..._KillSkipsAHeadquarters`; `TestFR_TRG_005_IfYouDoRunsYWhenXWasDone`, `..._SkipsYWhenXPickedNothing`, `..._SkipsYWhenNothingMatchesX`, `..._FollowsADiscardCount`; `TestFR_TRG_004_DeathTriggersOfAnXKillArePlacedAfterTheTriggerFinishes`.
  - `go vet` fails only on the missing `OpKill`, `SubjectPick`, `Step.UpTo`, `Step.IfDone`. `cardtown lint rules` passes.
  - The first test removes an ally with a `SubjectTarget` kill step before the pick, to show the pick sees only what matches at resolution.
  - Prompt text of the pick decision is not asserted (implementation detail). Left out per the task: a non-adjacent `IfDone`, a replaced kill (SVY-R15), Arkha's pile pick.
