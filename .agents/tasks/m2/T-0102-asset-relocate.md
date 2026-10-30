---
id: T-0102
title: A step returns or puts any asset it acts on, not only a replaced one (FR-ZON-004, FR-TGT-003, directive #76, wave G.1 shape, sonnet)
kind: shape
tier: sonnet
wave: G.1
status: tests-review
branch: m2/T-0102-asset-relocate
depends_on: [T-0098]
claims: [primitive:RETURN_TO_HAND(asset), primitive:PUT_ON_DECK(asset), helper:relocate-destination]
cards: []
requirements: [FR-ZON-004, FR-TGT-003]
---

## Goal
`OpReturnToHand` and `OpPutOnDeck` move every battlefield asset their step acts on, whatever the subject: a target slot, an untargeted pick, the source itself, every matching asset, or a replaced asset. Each leaves the battlefield with its last-known state and goes to its owner's hand or deck. Moving an asset to a hand or deck is written once, in `effects.go`, so wave C's *Disorienting Pulse*, *Tactical Withdrawal*, *Animoceps Spores* and *Ambush Tactics* only combine shapes. This applies board directive #76 (design note Deliverable 6d).

## Inputs
- FRD 01 v3.18: `FR-ZON-004` (lines 209-221: state reset on leaving the battlefield, last-known state, Departure to a Private Zone at line 221); `FR-TGT-003` (line 356 on: two-stage verification, fizzle); untargeted selections (line 347, `SVY-R3`).
- Survey rows: none (wave G.1 has no survey row and no cards).
- Rulings that apply: `SVY-R3` (an untargeted pick is not a target, so `FR-TGT-003` does not apply to it).
- Design note: `docs/architecture/m2-ability-model.md`, Deliverable 6d; Key types (`Step`, untargeted pick).
- Board directive #76.
- Kernel today:
  - `relocateReplaced` (`replace.go`) moves a battlefield asset to its owner's hand or deck, with its last-known state, and `stepRelocate` (`effects.go`) reaches it only for `SubjectReplaced`.
  - For any other subject, `stepRelocate` moves only the cards `stepCards` returns (a tracked death trigger's own card, pile picks, pile targets). An asset in a target slot, an untargeted asset pick, a source still on the battlefield, or `SubjectAll` is not moved.
  - "Hand or deck, top or bottom" is decided twice: in `stepRelocate` and in `relocateReplaced`.

### Shape (lead decisions, binding)
- **One destination helper (`effects.go`):** a function that maps a relocate step to where its cards go, for example `relocateTo(st ability.Step) (to Zone, bottom bool)`: `ZoneHand` for `OpReturnToHand`, `ZoneDeck` and `st.Bottom` for `OpPutOnDeck`. Both halves below call it, and no other non-test file in `pkg/engine` maps either op to a zone. Name it as reads best.
- **`stepRelocate` (`effects.go`) has two halves, and no `SubjectReplaced` branch:**
  1. **Assets:** every asset `g.stepAssets(r, st)` returns leaves the battlefield and goes to its owner's hand or deck, as `relocateReplaced` does today: the `leave` event (`Player: -1`, `Target`, `Zone`, and no `Card`, because a face-down asset's card is private in a hand or deck), the same log text ("%s returns to its owner's hand", "%s is put into its owner's deck"), `g.leaveBattlefield(a, g.effectiveState(a))`, then `g.putCard(a.Inst, to, bottom)`.
  2. **Cards:** then every card `g.stepCards(r, st)` returns moves as today (`takeCard`, then `putCard` through the helper).
  It returns how many assets and cards it moved. `stepAssets` already covers every asset subject, `SubjectReplaced` included, so nothing else changes there.
- **`replace.go` keeps only replacement itself:** finding, ordering, choosing and applying (`replacementsFor`, `replacedQueue`, `replaceDeaths`, `replaceDeath`, `chooseReplacement`). `relocateReplaced` is deleted; its body is the asset half above.
- **Doc comments:** `stepRelocate`'s says it moves each asset of `stepAssets`, then each card of `stepCards`, to its owner's hand or deck. In `ability`, the comments on `OpReturnToHand` and `OpPutOnDeck` say "the Subject card or asset" in place of "the Subject card", and `Bottom`'s stays.
- **A Headquarters gets no special case.** A step's filter decides what it can act on, and every printed return or put-on-deck clause in M2 names units. No test covers a Headquarters.
- **Sims:** no current card reaches the new paths. *Unceasing Tide*'s and *Stormfront of Ideals*' `SelfDeath` steps run after their sources left the battlefield, so `stepAssets` returns nothing for them, and *Shifting Impossibility*'s `SubjectReplaced` step takes the same path as before. So `go tool task sim-compare -- autoloop/m2-wave-g.1` should pass byte for byte. If it differs, don't change code to make it match: report the first differing game in the Log and hand off.
- **Left for later:** no hook fires when an asset leaves the battlefield other than by destruction ("When this asset leaves the battlefield", the moves-zone hook). Neither is built, and no M2 card prints a trigger that a return or put-on-deck would fire.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
Synthetic cards only, in a new file `pkg/engine/relocate_test.go`. Reuse the scenario helpers in `scenario_test.go`, and the helpers in `pile_pick_test.go`, `targeting_test.go` or `replacement_test.go` where they fit. Every existing test passes unchanged.
- `TestFR_ZON_004_ReturnedTargetLeavesWithItsLastKnownState` (directive #76 (c)): a triggered ability with one target slot for an enemy unit and one `OpReturnToHand` step on `SubjectTarget`. The chosen enemy unit has taken damage and has a temporary power modifier when the ability resolves. It is no longer on the battlefield, its card is in its owner's hand, the `leave` event names it with zone `hand`, and `LastKnownEffective` returns its damage-free effective state at departure (its modified power). Back in hand, its card has printed characteristics: deploying it again gives a new asset with no damage and printed power.
- `TestFR_ZON_004_PickedUnitGoesOnTheBottomOfItsOwnersDeck` (directive #76 (d), `SVY-R3`): an `OpPutOnDeck` step with `Bottom` on `SubjectPick`, filter allied unit, `UpTo` 1. The picked allied unit leaves the battlefield and its card is the last card of its owner's deck; the top card is unchanged.
- `TestFR_ZON_004_PutOnDeckGoesOnTopWithoutBottom`: the same as the previous test with `Bottom` false, on a target slot: the card is the first card of its owner's deck.
- `TestFR_TGT_003_TargetKilledByAnEarlierStepIsNotMoved` (directive #76 (e)): a chosen unit that leaves the battlefield before the ability resolves (killed, so its card is in the discard pile) is not moved by the `OpReturnToHand` step: its card stays in the discard pile, no `leave` event to hand is logged, and the step counts 0 (a following `IfDone` step does not run).
- `TestFR_ZON_004_FaceDownAssetReturnsAsItsCard` (directive #76 (f), `FR-ZON-004` Departure to a Private Zone): a face-down Scratch Company unit chosen for an `OpReturnToHand` step goes to its owner's hand as its own printed card instance, and the `leave` event's `Card` is empty.
- `TestFR_ZON_004_SourceStillOnTheBattlefieldReturnsItself`: an ability whose step is `OpReturnToHand` on `SubjectSelf` resolves while its source is on the battlefield (for example a deploy trigger): the source leaves the battlefield and its card is in its owner's hand.
- `TestFR_ZON_004_ReturnAllMatchingUnits`: an `OpReturnToHand` step on `SubjectAll` with a filter for enemy units returns every enemy unit and no allied one; the step counts how many it moved (a following `IfDone` step runs).

Directive #76's (a) (every existing test passes unchanged, including `TestFR_REP_001_*` and *Shifting Impossibility*'s), (b) (`sim-compare`) and (g) (one mapping) are checked by the gates and the review, not by new tests.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected. The test file may declare its own helpers.

## Allowed Files
- `pkg/engine/effects.go` (`stepRelocate` with its asset half, the destination helper)
- `pkg/engine/replace.go` (`relocateReplaced` deleted)
- `pkg/engine/ability/ability.go` (the `OpReturnToHand` and `OpPutOnDeck` comments only)

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run 'TestFR_ZON_004|TestFR_TGT_003|TestFR_REP_001|TestCard_ShiftingImpossibility'`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-compare -- autoloop/m2-wave-g.1` passes (no current card reaches the new paths; see Shape, Sims)
- [ ] Outside the destination helper, no non-test file in `pkg/engine` maps `OpReturnToHand` or `OpPutOnDeck` to a zone (directive #76 (g)), and `replace.go` declares no relocation.

## Log
- 2026-10-29 lead-architect (wave G.1 run, session 1): Wrote the task for board directive #76 (`open: before the next Deliverables task`, design note Deliverable 6d's first task). Run file created; `main` (556811f) has nothing the run branch lacks, so no sync.
  - **Directives:** read `BOARD.md` on `autoloop/m2-wave-g.1` fresh. #76 is applied by this task; #77, #78 and #79 follow it in board order. #81's trigger has fired (`docs/design/outcomes/2026-10-29-replacement-once-per-sweep.md` is `applied: FRD 01 v3.18`), so it is `open`, and by its own Order line it goes after #79. #14 is a standing order. Waiting triggers checked: #6 and #7 (mode is `serial`); #8 (no `deck.go` change); #15 (no turn hook); #20 (b) (no wave C shape task); #58 (no exhaust by a cost or an effect); #59 (2) (no repeated targets); #61 (no modal card); #71 (no effect on each player). None has fired.
  - **Reuse:** this task removes a duplicate rather than adding one: `relocateReplaced`'s move becomes the asset half of `stepRelocate`, and the op-to-zone mapping is written once.
  - **Tests beyond the directive:** put on top without `Bottom`, the source returning itself, and `SubjectAll`, so every asset subject `stepAssets` returns is pinned once.
- 2026-10-29 rules-test-author: Wrote `pkg/engine/relocate_test.go` (the 7 tests listed under Tests): `TestFR_ZON_004_ReturnedTargetLeavesWithItsLastKnownState`, `TestFR_ZON_004_PickedUnitGoesOnTheBottomOfItsOwnersDeck`, `TestFR_ZON_004_PutOnDeckGoesOnTopWithoutBottom`, `TestFR_TGT_003_TargetKilledByAnEarlierStepIsNotMoved`, `TestFR_ZON_004_FaceDownAssetReturnsAsItsCard`, `TestFR_ZON_004_SourceStillOnTheBattlefieldReturnsItself`, `TestFR_ZON_004_ReturnAllMatchingUnits`. No helpers requested (the file declares its own `rl*` helpers).
  - **Fail before the implementation:** all but the FR-TGT-003 test, because `stepRelocate` moves no asset for `SubjectTarget`, `SubjectPick`, `SubjectSelf` or `SubjectAll`. Several stop at their first "leaves the battlefield" check, so their later checks (last-known power 4, new asset on redeploy, deck position, leave event) are not yet exercised; the redeploy mechanics were checked separately.
  - **Passes already:** `TestFR_TGT_003_TargetKilledByAnEarlierStepIsNotMoved` is a regression guard (the current code also moves nothing); it kills the target with a `OpKill` step earlier in the same ability, since a fizzle needs the unit gone before the return step.
  - **Left out:** nothing; no Headquarters test, as the task says.
- 2026-10-29 lead-architect (tests review, round 1): The outcomes match `FR-ZON-004`, `FR-TGT-003` and `SVY-R3`. The FR-TGT-003 test's mid-resolution departure is the stricter case and is kept: `r.chosen` still holds the killed asset when the return step runs, so the asset half must skip an asset no longer on the battlefield. A target gone before resolution is dropped by `resolveAbility`'s second verification, which every op shares and earlier tests pin. Back to `tests` for two edits, both in a new `test(T-0102)` commit:
  1. `TestFR_ZON_004_FaceDownAssetReturnsAsItsCard`: drop `NonScratch: false` from the target filter. It is the zero value, and directive #78 removes the field, which would make this an unnamed seventh test site for #78.
  2. Rename `TestFR_TGT_003_TargetThatLeftBeforeResolutionIsNotMoved` to `TestFR_TGT_003_TargetKilledByAnEarlierStepIsNotMoved`, so the name says what it sets up, and update the Tests list and your Log line to match.
- 2026-10-29 rules-test-author: Round 1 notes applied: dropped `NonScratch: false` from the face-down test's filter; renamed the fizzle test to `TestFR_TGT_003_TargetKilledByAnEarlierStepIsNotMoved` (setup and assertions unchanged).
