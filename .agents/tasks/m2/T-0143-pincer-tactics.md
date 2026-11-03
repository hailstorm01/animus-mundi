---
id: T-0143
title: Make Pincer Tactics executable (1 damage to each of up to 2 chosen allied units, then 2 damage to a chosen non-Scratch asset; wave C card, flash)
kind: card
tier: flash
wave: C
status: ready
branch: m2/autoloop-wave-c/T-0143-pincer-tactics
depends_on: [T-0017, T-0030, T-0056, T-0104, T-0107, T-0116, T-0117]
claims: [card:Pincer Tactics]
cards: [Pincer Tactics]
requirements: [FR-MNV-001, FR-TGT-001, FR-TGT-002, FR-TGT-003, FR-MOD-001, FR-REP-004, FR-KWD-002h]
---

## Goal
*Pincer Tactics* (Path Maneuver, not in a starter deck, kept in this run's `Cards` by #14 (c); 0 Requisition, 3 Tactics) is executable. It only combines shapes that are `done`: enacting a Maneuver (T-0116) with its required targets (T-0117), chosen targets and `OpDealDamage` (T-0017), an "up to N" slot of distinct entities (T-0030, `SVY-R11`), the `ScratchExcluded` filter (T-0104), presence checks between steps (T-0107) and timed-out targets in fixed order (T-0056). It also carries #91 (4)'s *Pincer Tactics* same-unit test (`SVY-R35`), the last open part of #91 (4).

## Inputs
- Card (`data/cards/maneuvers-path.csv` line 9): *Pincer Tactics*, Path Maneuver, cost 0 Requisition / 3 Tactics, one rules-text line: "Deal 1 damage to each of up to 2 chosen allied units. Deal 2 damage to chosen non-Scratch asset."
- Survey row (`docs/design/m2_starter_set_ability_survey.md` line 177): `DEAL_DAMAGE 1` to up to 2 chosen allied units; `DEAL_DAMAGE 2` to a chosen non-Scratch asset; ruling R35 (the non-Scratch asset may be one of the two units).
- FRD 01 (run branch, v3.20):
  - `FR-MNV-001` (line 340, `SVY-R29`): enacted from Hand, costs paid on declaration; enactable only if each exactly-one slot has a legal entity, and "an *up to N* slot never stops it". The only exactly-one slot is the non-Scratch asset, and the enacting player's own Headquarters is always a legal one (Ward guards only against opponents, `FR-KWD-002h`), so the card is enactable whenever its Tactics can be paid. The card goes to its owner's Discard pile after resolving.
  - `FR-TGT-002`, "Up To N" Targets Are Distinct (`SVY-R11`, line 369): the up-to slot picks 0, 1 or 2 different allied units, never one unit twice.
  - `FR-TGT-002`, Separate Target Clauses (`SVY-R35`, line 371): the two clauses are two independent slots. The non-Scratch asset may be one of the allied units, which then takes 1 + 2 damage; *Pincer Tactics* is the FRD's own example.
  - `FR-TGT-001` (line 353): "chosen non-Scratch asset" has no allegiance or kind: allied and opposing units, structures and Headquarters are all legal, in either zone, but no Scratch Company unit (face up or face down). "Chosen allied units" is allied units only, Scratch Company included; no structure or Headquarters.
  - `FR-TGT-003` (lines 375-377, `SVY-R25`, `SVY-R26`): between the stages a step checks only presence. Lethal damage is applied in the state-based death sweep, not mid-resolution, so a 1-health allied unit that takes the 1 is still there for the 2 when it is also the second slot's pick.
  - `FR-KWD-002h` (line 578, Ward): an opposing Ward unit is never a legal second-slot target; an allied Ward unit is legal in either slot.
  - `FR-REP-004`: each step is ordinary non-combat damage (Armor absorbs it first).
  - `FR-MOD-001` §6 (line 397, `SVY-R16`): on a timeout an "up to N" selection selects none, and the mandatory non-Scratch slot takes the first legal entity in fixed order: the choosing player's own assets first, Headquarters first. A fully timed-out *Pincer Tactics* deals 2 damage to its controller's own Headquarters.
  - `FR-KWD-002a` (line 569): Unbowed triggers whenever its asset takes actual damage, so an allied Unbowed unit chosen for the 1 triggers.
- Shapes it uses, all `done`: `ActEnact` and `StackManeuver` (T-0116); `SVY-R29` legality, which skips up-to slots (T-0117); `ability.Target.UpTo` (T-0030, used by *Craterfall Ancient*, `animus.go:74`, and *Tactical Withdrawal*, `path.go:544`); `ability.ScratchExcluded` with `Kinds` zero (T-0104; `matchesPrinted` in `targeting.go:254-258` lets structures and Headquarters through); `OpDealDamage` on `SubjectTarget` (T-0017); `stepAssets` presence checks (T-0107).

### Definition (the engineer adds it to `pathDefs`)
- *Pincer Tactics*: `Text` holds the one line. `Effect`:
  - `Targets`: slot 0 `ability.Target{Filter: ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit}, UpTo: 2}`; slot 1 `ability.Target{Filter: ability.Filter{Scratch: ability.ScratchExcluded}}`.
  - `Steps`: `{Op: ability.OpDealDamage, Amount: ability.Value{N: 1}, Subject: ability.SubjectTarget, Target: 0}`, then `{Op: ability.OpDealDamage, Amount: ability.Value{N: 2}, Subject: ability.SubjectTarget, Target: 1}`.
- No `Triggers`, `ExtraCosts` or `Modes`. Place it in the map's alphabetical order (between *Necropolis Hunt* and *Rogue Fleshshaper*). Model it on *Saboteur* (two slots, two damage steps, `path.go:466`) and *Tactical Withdrawal* (an up-to slot on a Maneuver, `path.go:535`).
- Remove the name from `cards.NotYetExecutable`.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
To be added to `pkg/engine/cards_path_maneuvers_test.go`, from the CSV text, loading the real card (`s.card("Pincer Tactics")`). Player 0 enacts; player 1 is the opponent. Give player 0 the Tactics to pay. Answer each `DecideTarget` decision separately (slot 0 is the up-to allied units, slot 1 the non-Scratch asset); a file-local chooser is fine. Use units with enough health to survive unless a test says otherwise.
- `TestCard_PincerTactics_DamagesTwoAlliesAndAnAsset` (`FR-MNV-001`, `FR-REP-004`): allied units A and B each take 1 damage; opposing unit C takes 2; *Pincer Tactics* is then the top of player 0's discard pile; 3 Tactics are spent.
- `TestCard_PincerTactics_SameUnitTakesOneAndTwo` (`SVY-R35`, `FR-TGT-003`, #91 (4)): slot 0 picks allied non-Scratch unit A alone and slot 1 picks A too: A has 3 damage. A second case: A has 1 health left, so it is still on the battlefield for the 2 and dies in the sweep after the effect.
- `TestCard_PincerTactics_UpToTwoMayBeNoneOrOne`: choosing no allied unit deals only the 2 to the second pick; choosing one deals 1 to it alone.
- `TestCard_PincerTactics_AlliedUnitsAreDistinct` (`SVY-R11`): the up-to slot can't name one unit twice (as `TestFR_TGT_002_UpToTwoAcceptsZeroOneOrTwoDistinctTargets` in `optional_upto_test.go` checks it); a unit picked there takes 1 damage, not 2.
- `TestCard_PincerTactics_FirstSlotOffersAlliedUnitsOnly` (`FR-TGT-001`): its options include allied units in either zone, an allied Ward unit and an allied face-down Scratch Company unit, but no opposing unit, no structure and no Headquarters.
- `TestCard_PincerTactics_SecondSlotIsAnyNonScratchAsset` (`FR-TGT-001`, `FR-KWD-002h`): its options include allied and opposing non-Scratch units in either zone, an allied Ward unit, allied and opposing structures and both Headquarters, but no Scratch Company unit (allied or opposing, face up or face down) and no opposing Ward unit.
- `TestCard_PincerTactics_EnactableWithNoUnits` (`SVY-R29`, `FR-MNV-001`): with no unit on the battlefield (only Headquarters and a structure) it is offered (`offersEnact`) and enacted: the up-to slot asks nothing or offers nothing, and the chosen opposing Headquarters takes 2 damage.
- `TestCard_PincerTactics_ArmorAbsorbsTheDamage` (`FR-REP-004`): an allied unit with 1 Armor chosen for slot 0 loses the Armor and takes no damage.
- `TestCard_PincerTactics_DamageTriggersUnbowed` (`FR-KWD-002a`): an allied executable Unbowed unit that survives 1 damage, chosen for slot 0 only, has its Unbowed trigger placed once. Pick a unit whose Unbowed effect is easy to observe (for example *Allshan, Chainbreaker Revenant*, whose Unbowed gains it 2 Armor) and say which in the Log.
- `TestCard_PincerTactics_TargetTimeoutsTakeFixedOrder` (`FR-MOD-001` §6, `SVY-R16`): both target decisions time out, with allied and opposing units on the battlefield: no allied unit takes damage, and player 0's own Headquarters takes 2.
- `TestCard_PincerTactics_PaysThreeTactics`: with 2 Tactics it is not offered and enacting it is rejected (`s.reject`), and it stays in hand; with 4 it is enacted, leaving 1 Tactics and spending no Requisition.
- **Existing tests:** every other test passes unchanged, including `TestStarterSet_Executable` with the name removed, `TestView_Case8_ExecutableFlag`, `TestProgress_Case4_Coverage`, `TestEvents_Case4_EveryActionKind` and the *Saboteur*, *Craterfall Ancient* and *Tactical Withdrawal* tests.
- **Left out on purpose:** whether one unit picked by both slots triggers its Unbowed once per step. Each step is a separate damage event, but no test needs the count, so don't pin it.

## Helpers Needed
- None expected. The tests may add their own setup helper in the test file.

## Allowed Files
- `pkg/engine/cards/path.go` (the definition)
- `pkg/engine/cards/cards.go` (remove the name from `NotYetExecutable`)

## Done When
- [ ] The card tests pass: `go test ./pkg/engine -run TestCard_PincerTactics`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] The card loads with no unsupported lines (`TestStarterSet_Executable` passes with the name removed)

## Log
- 2026-11-03 lead-architect (wave C session 36): written against the run branch at 49a0dd5 (no `main` commits to sync). Not `WAVE-COMPLETE`: *Pincer Tactics* is the last header card on `NotYetExecutable`. A non-deck card kept by #14 (c): every shape it needs is `done`. #91 (4)'s last part goes into this task (the same-unit test); #91 (2) is about a cost on a mode, and this card has no modes. Waiting triggers #6, #7 (mode is `serial`), #8 (this task doesn't touch `deck.go`), #15 (no turn-start, turn-end or draw hook; damage doesn't expire), #71 (no each-player effect) not fired. Status `tests`.
- 2026-11-03 rules-test-author: wrote 12 tests in `pkg/engine/cards_path_maneuvers_test.go` (with file-local helpers `pincerChoose`, `pincerOffers`): `TestCard_PincerTactics_` `DamagesTwoAlliesAndAnAsset`, `SameUnitTakesOneAndTwo`, `SameUnitDiesInTheSweepAfterBothSteps` (the 1-health case as its own test), `UpToTwoMayBeNoneOrOne`, `AlliedUnitsAreDistinct`, `FirstSlotOffersAlliedUnitsOnly`, `SecondSlotIsAnyNonScratchAsset`, `EnactableWithNoUnits`, `ArmorAbsorbsTheDamage`, `DamageTriggersUnbowed` (uses *Allshan, Chainbreaker Revenant*: survives 1 damage, Armor becomes 2), `TargetTimeoutsTakeFixedOrder`, `PaysThreeTactics`. `go vet ./pkg/engine` and `cardtown lint rules` are clean; all 12 fail now (card not yet executable). Helpers requested: none. Left out: face-up Scratch Company units in the option checks (the harness can only build face-down ones; face-down allied and opposing are covered); the Unbowed once-per-step count, as the task says. The repeated-pick case asserts only that the unit takes no more than the 1 (an invalid answer falls back to none, as in `optional_upto_test.go`). Slot 0 is told from slot 1 by `Decision.Min == 0`.
- 2026-11-03 lead-architect (wave C session 36): tests approved (8a89c94). Checked each test against FRD 01: `SVY-R29` (line 340: an up-to slot never stops it, and the controller's own Headquarters is always legal) for enacting with no unit; `SVY-R11` (line 369) for distinct allies (the weaker repeated-pick check matches `optional_upto_test.go`, where an invalid answer falls back to none); `SVY-R35` (line 371) and `FR-TGT-003` (lines 375-377) for one unit taking 1 + 2, including the 1-health unit still there for the 2 and dying in the sweep; `FR-TGT-001` and `FR-KWD-002h` for both slots' options (face-up Scratch Company left out: the harness builds only face-down ones, and the filter reads the subtype either way); `FR-REP-004` for Armor; `FR-KWD-002a` for *Allshan*'s Unbowed; `FR-MOD-001` §6 for the timeouts (none for the up-to slot, the controller's own Headquarters for the other). Run against a discarded trial definition (the Definition above): all 12 pass, and `go test ./pkg/engine/...` passes with it. Five mutations are each caught: both steps on slot 0 (7 tests fail), no Scratch exclusion (1), any allegiance for the up-to slot (1), an opposing-only second slot (4) and the wrong first amount (7). Status `ready`, dispatching `kernel-engineer` (flash).
