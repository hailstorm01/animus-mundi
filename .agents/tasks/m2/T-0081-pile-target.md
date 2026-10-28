---
id: T-0081
title: A chosen target in a discard pile, chosen on placement and verified at resolution (FR-TGT-001, FR-TGT-003, wave B shape, sonnet)
kind: shape
tier: sonnet
wave: B
status: tests-review
branch: m2/T-0081-pile-target
depends_on: [T-0077]
claims: [target:pile, view:pile-targets]
cards: []
requirements: [FR-TGT-001, FR-TGT-003, FR-MOD-001]
---

## Goal
A `Target` whose `Filter.Zone` is `ability.ZoneDiscard` chooses cards in discard piles instead of battlefield assets. It is chosen when the trigger is placed (`SVY-R3`), verified again when the trigger resolves (`FR-TGT-003`), and the card steps `OpReturnToHand` and `OpPutOnDeck` with `SubjectTarget` move the chosen cards that are still legal. The view shows a pile target by its card's name, never its instance ID. The first card, *Eternal Watchman* ("Whenever this Headquarters attacks and kills an opposing unit or structure, return chosen unit card from your discard pile to your hand."), is the next card task, on top of `HookCombatKill` (T-0080).

## Inputs
- FRD 01:
  - `FR-TGT-001` Conjunction Eligibility (lines 319-323): a *Card* in a specified non-battlefield zone is a legal target class, and *"from your discard pile"* is a zone criterion.
  - `FR-TGT-001` "Targets Are Locked on Placement (`SVY-R3`)" (line 326): every target is chosen when the item is placed.
  - `FR-TGT-003` Two-Stage Verification and Partial Invalidation (lines 339-340): a target that left its zone, or no longer meets the filter, is illegal at resolution, and the rest of the effect resolves.
  - `FR-MOD-001` §1 (fast path) and §6 (line 359): mandatory targets time out to the first legal options in fixed order, "Cards in a pile such as the Discard pile: from the top down"; an "up to N" selection selects none.
  - `FR-ZON-004` (line 202): a card in a discard pile is face-up with its printed characteristics, which is what the filter reads.
  - `FR-KWD-002h` (line 539): Ward is a battlefield asset's keyword. A card in a pile has none, so `legalTarget`'s Ward check does not apply to it.
- Survey row: Eternal Watchman (line 153): `RETURN_TO_HAND` chosen unit card from discard.
- Rulings: `SVY-R3` (targets on placement), `SVY-R11` ("up to N" distinct), `SVY-R16` (fixed-order fallback).
- Design note: `docs/architecture/m2-ability-model.md`, Key types (`Target`, Stack) and Open Question 10 (added with this task).
- Kernel today (T-0077 built the pile half of this):
  - `chooseTargets` (`targeting.go`) builds every slot's decision with `assetDecision` and `legalTarget`. `pileDecision` and `matchesCard` (`targeting.go`) already build a pile decision and check a pile filter, for the untargeted pick.
  - `resolveAbility` (`effects.go`) re-verifies each slot's IDs with `g.Asset(id)` and `legalTarget` into `resolution.chosen [][]*Asset`. A pile slot's IDs are card instance IDs, so today they would all be dropped.
  - `stepCards` (`effects.go`) handles `SubjectSelf` and a `ZoneDiscard` `SubjectPick`, not `SubjectTarget`.
  - `fixedOrderPositions` (`state.go`) already ranks discard-pile cards top-down after every battlefield asset, so the timeout needs no change.
  - `pkg/view` promises that "no asset and no discard card carries a card instance ID" (`view.go` package comment). `stackView` copies `StackItem.Chosen` as is, so a pile slot would break that promise. `view.go` imports `engine` and `eventlog` only (the import fence), so it cannot read `ability.ZoneDiscard` itself.
  - Instance and asset IDs come from one counter (`newID`), so an ID never means both.

### Shape (lead decisions, binding)
- **Which slots are pile slots:** add `func (it *StackItem) PileSlot(i int) bool` in `stack.go`: whether `it.Effect.Targets[i]` chooses cards in discard piles (`Filter.Zone` is `ability.ZoneDiscard`), so `Chosen[i]` holds card instance IDs, not asset IDs. It returns false when `i` is out of range of `Effect.Targets` (view tests build items with `Chosen` and no `Targets`). Update the `Chosen` field comment: `// StackAbility: per Effect.Targets slot, in pick order, the asset IDs chosen, or for a PileSlot the card instance IDs; empty if none`.
- **Choosing on placement:** in `chooseTargets`, a slot whose `Filter.Zone` is `ZoneDiscard` builds its decision with `pileDecision(p.controller, DecideTarget, prompt, t.UpTo, …)` and `matchesCard(p.controller, p.inst, t.Filter, inst)`. Everything else stays as it is for both kinds of slot: the same prompts, the fast path, the fixed-order timeout, "up to N" always asked and distinct, and an empty slot that asks nothing. The loop that turns picks into `Ref`s is shared, not copied. Update the function comment to say a slot may be in a discard pile.
- **Verifying at resolution:** add `(g *Game) pileTarget(ctl PlayerID, src *Instance, f ability.Filter, id int) *Instance` in `targeting.go`, beside `legalTarget`. It returns the card with instance ID `id` if that card is still in a discard pile and `matchesCard` still holds, else nil. A card goes only to its owner's pile, so "still in a discard pile" plus the allegiance check is "still in the zone the filter names" (`FR-TGT-003`).
- **`resolution`:** add `cards [][]*Instance`, one entry per target slot: for a `PileSlot`, its chosen cards still legal, in pick order; empty for an asset slot. `chosen` stays the asset slots' assets (empty for a pile slot). `resolveAbility` fills each slot through `pileTarget` or through today's `g.Asset` and `legalTarget`, chosen by `it.PileSlot(i)`. Update the `resolution` comment.
- **Steps:** `stepCards` gains `SubjectTarget`: `r.cards[st.Target]`, bounds-checked as `stepAssets` checks `r.chosen`. Its comment names the new case. `stepRelocate` is unchanged: it already moves whatever `stepCards` returns, and counts it for `IfDone`. `stepAssets` is unchanged: a pile slot's `chosen` entry is empty, so an asset step on it does nothing.
- **The view (`pkg/view/view.go`):** `StackView` gains `ChosenCards [][]string // Ability: per target slot, for a pile slot, the names of its chosen cards that are still in a discard pile, in pick order; empty for an asset slot`. `stackView` fills `Chosen[i]` only for asset slots (a pile slot's entry is empty) and `ChosenCards[i]` only for pile slots, looking each instance ID up in both players' `Discard`. A chosen card that has left the piles is left out, so every name shown is public. The `Chosen` field comment becomes `// Ability: per target slot, the chosen asset IDs; empty for a pile slot (see ChosenCards)`. The package comment stays true and is not changed.
- **Ward:** `pileTarget` and the placement check use `matchesCard` only. A card in a pile has no Ward.
- **Not built:** a card that leaves the pile and comes back between placement and resolution is the same `*Instance`, so it would stay legal, though `FR-ZON-004` makes it a new object. No M2 card can do this inside that window. The design note's Open Question 10 records it (it applies to `Tracked` self-relocation too). Not tested.
- **`doc.go`** is not changed: its line "its effects may choose targets (… a zone) that are verified when it is placed and again when it resolves" stays true.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
Engine tests in `pkg/engine/pile_target_test.go`, with synthetic cards only (Eternal Watchman is the next task). A convenient source is a unit with a deploy trigger (as `iydDeploy` in `if_you_do_test.go`), whose `Effect.Targets` has a pile slot, for example `Filter{Allegiance: Allied, Kinds: KindUnit, Zone: ZoneDiscard}`, and whose step is `{Op: OpReturnToHand, Subject: SubjectTarget, Target: 0}`. `s.inDiscard` (T-0077) fills a pile. Required cases:
- **Options:** `TestFR_TGT_001_PileTargetOffersMatchingCardsInYourDiscardPile`. The trigger asks a `DecideTarget` whose options are exactly the unit cards in the controller's discard pile, bottom first, labelled with their card names. A structure card in that pile, a unit card in the opponent's pile and an allied unit on the battlefield are not offered.
- **On placement (`SVY-R3`):** `TestFR_TGT_001_PileTargetIsChosenOnPlacement`. The target decision is asked before the ability's `place` event, and the item's `Chosen` holds the chosen card's instance ID while it waits on the stack.
- **Return to hand:** `TestFR_TGT_001_PileTargetReturnsTheChosenCardToHand`. The chosen card leaves the discard pile and is the last card of its owner's hand.
- **Bottom of the deck:** `TestFR_TGT_001_PileTargetCanGoOnTheBottomOfTheDeck`. With `OpPutOnDeck` and `Bottom`, the chosen card is the last card of its owner's deck.
- **Left the pile (`FR-TGT-003`):** `TestFR_TGT_003_PileTargetThatLeftThePileIsIllegal`. After placement and before resolution the chosen card leaves the discard pile (for example a trigger placed above it moves it with a `ZoneDiscard` pick, T-0077, or the test moves it with `s.g.takeCard` and `s.g.putCard` between placement and resolution). The ability still resolves, the step moves nothing, and the card stays where it went. A later `IfDone` step does not run.
- **Partial invalidation (`FR-TGT-003`):** `TestFR_TGT_003_PileTargetUpToNKeepsTheCardsStillInThePile`. An `UpTo` 2 pile slot chooses two cards; one leaves the pile before resolution; the other is still returned to hand.
- **No match:** `TestFR_TGT_003_PileTargetWithNoMatchAsksNothing`. With no matching card, no target decision is asked, the trigger is still placed and resolves, and nothing moves.
- **Fast path (`FR-MOD-001` §1):** `TestFR_MOD_001_SinglePileTargetIsNotAsked`.
- **Top-down timeout (`FR-MOD-001` §6):** `TestFR_MOD_001_PileTargetTimeoutChoosesTheTopCard`. An exactly-one slot over three matching cards A, B and C, created in that order, lying in the discard pile bottom to top as A, C, B. A timeout chooses B, the top card, which is neither the oldest card nor the first option.
- **An asset step on a pile slot:** `TestFR_TGT_001_AssetStepOnAPileSlotDoesNothing`. An `OpDealDamage` step on a pile slot damages nothing, and the card step beside it still moves the card.
- **Existing tests:** every test in `pkg/engine` and `pkg/view` passes unchanged, including the asset-target tests from wave A2 and the pick tests from T-0072, T-0076 and T-0077.

View test in `pkg/view/view_test.go` (a new test; no existing test changes):
- `TestView_PileTargetShowsCardNamesNotInstanceIDs`. A game with a `StackAbility` item whose `Effect.Targets` has one asset slot and one pile slot, and whose `Chosen` holds an asset ID and two discard cards' instance IDs, one of which is no longer in any discard pile. In both seats' views: `Chosen` is the asset ID for the asset slot and empty for the pile slot, `ChosenCards` is empty for the asset slot and holds only the name of the card still in a pile for the pile slot, and no field of the view carries either card's instance ID.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected. If a test needs one, `rules-test-author` names it here.

## Allowed Files
- `pkg/engine/stack.go` (`PileSlot`, the `Chosen` comment)
- `pkg/engine/targeting.go` (`chooseTargets` pile slots, `pileTarget`)
- `pkg/engine/effects.go` (`resolution.cards`, `resolveAbility`'s verification, `stepCards`' `SubjectTarget`)
- `pkg/view/view.go` (`StackView.ChosenCards`, `stackView`)

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run 'TestFR_(TGT_00[13]_(PileTarget|AssetStepOnAPileSlot)|MOD_001_(SinglePileTarget|PileTargetTimeout))'` and `go test ./pkg/view -run TestView_PileTarget`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] No card uses a pile target yet, so the sims are unchanged: `go tool task sim-compare -- autoloop/m2-wave-b` passes

## Log
- 2026-10-27 lead-architect (wave run B, session 18): Wrote the task. The run branch has everything `main` has (merge base is `main` at 24a9ec9), and no `open: before` directive is on the board. Waiting triggers checked against `autoloop/m2-wave-b` 32e4966: #6 and #7 (mode is `serial`), #8 (no `deck.go` change), #15 (no turn hook), #20 (b) (no wave C shape task), #58 (no exhaust by a cost or an effect) and #59 (2) (no repeated targets: "up to N" targets stay distinct). None has fired. Directive #14 is open and followed: Eternal Watchman is a deck card (the Path Headquarters), so it goes before Rogue Fleshshaper.
  - Second of three tasks for the Watchman (T-0080 split): this one claims `target:pile` and `view:pile-targets`; the card task (`headquarters.go`, `cards.go`) comes next. No claims overlap: no other task is open.
  - Reuse: `pileDecision`, `matchesCard`, `pickBounds` and the top-down fixed order (T-0077) are reused, not copied. Searched `pkg/engine` and the `done` tasks for an existing chosen target outside the battlefield: none.
  - Placement: the view change is forced by the view's own promise that no discard card carries an instance ID, and the slot test sits on `StackItem` so that `pkg/view` doesn't import `ability`. Four non-test files.
- 2026-10-27 rules-test-author: Wrote the tests; status is now `tests-review`. No helpers requested.
  - `pkg/engine/pile_target_test.go`: `TestFR_TGT_001_PileTargetOffersMatchingCardsInYourDiscardPile`, `TestFR_TGT_001_PileTargetIsChosenOnPlacement`, `TestFR_TGT_001_PileTargetReturnsTheChosenCardToHand`, `TestFR_TGT_001_PileTargetCanGoOnTheBottomOfTheDeck`, `TestFR_TGT_003_PileTargetThatLeftThePileIsIllegal`, `TestFR_TGT_003_PileTargetUpToNKeepsTheCardsStillInThePile`, `TestFR_TGT_003_PileTargetWithNoMatchAsksNothing`, `TestFR_MOD_001_SinglePileTargetIsNotAsked`, `TestFR_MOD_001_PileTargetTimeoutChoosesTheTopCard`, `TestFR_TGT_001_AssetStepOnAPileSlotDoesNothing`.
  - `pkg/view/view_test.go`: `TestView_PileTargetShowsCardNamesNotInstanceIDs` (new test only; fails to compile until `StackView.ChosenCards` exists, which also stops the whole view test package from building).
  - The engine package compiles today (the tests use no new API). Eight fail now. `...ThatLeftThePileIsIllegal` and `...WithNoMatchAsksNothing` pass before the build because nothing moves today, so they guard against regressions. "Left the pile" moves the card from an observer hook on the ability's `place` event, using `takeCard` and `putCard`.
  - Left out: the leave-and-return case (Open Question 10), as the task says. `PileSlot` is covered through the view test (items with no `Targets`) and the engine tests, not by a direct unit test.
