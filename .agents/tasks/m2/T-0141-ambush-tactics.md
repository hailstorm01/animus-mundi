---
id: T-0141
title: Make Ambush Tactics executable (choose two of kill, return to hand, a chosen unit deals its power to another chosen unit, put on top of its owner's deck; wave C card, flash)
kind: card
tier: flash
wave: C
status: tests-review
branch: m2/autoloop-wave-c/T-0141-ambush-tactics
depends_on: [T-0083, T-0102, T-0107, T-0116, T-0117, T-0118, T-0123, T-0140]
claims: [card:Ambush Tactics]
cards: [Ambush Tactics]
requirements: [FR-MNV-001, FR-MOD-001, FR-TGT-002, FR-TGT-003, FR-ZON-004]
---

## Goal
*Ambush Tactics* (a deck card, two copies in `data/decks/path-starter.txt`; 0 Requisition, 4 Tactics) is executable. It only combines shapes that are `done`: "Choose N" modes chosen on placement (T-0083), enacting with N legal modes (T-0117), `OpKill` (T-0123), relocating a chosen asset to its owner's hand or deck (T-0102), the delegated dealer (T-0118), the "another chosen" slot (T-0140), and later steps skipping a removed target (T-0107).

## Inputs
- Card (`data/cards/maneuvers-path.csv` lines 13-17): *Ambush Tactics*, Path Maneuver, cost 0 Requisition / 4 Tactics, five rules-text lines:
  - "Choose two:"
  - "- Kill chosen unit."
  - "- Return chosen unit to its owner's hand."
  - "- Chosen unit deals damage equal to its power to another chosen unit."
  - "- Put chosen unit on top of its owner's deck."
- Survey row (`docs/design/m2_starter_set_ability_survey.md` line 179): choose two of `KILL`, `RETURN_TO_HAND`, delegated `DEAL_DAMAGE` = power, `PUT_CARD_ON_DECK` top; rulings R6, R25, R27–R29, R31, R33, R34.
- FRD 01 (run branch, v3.20):
  - `FR-MNV-001` (line 338): enacted from Hand, costs paid on declaration, the effect runs when it resolves, then the card goes to its owner's Discard pile. A Maneuver needs its required targets (`SVY-R29`).
  - `FR-MOD-001` §5 (line 392): two **different** modes (`SVY-R6`); a mode is legal only if each of its exactly-one slots has a legal entity (`SVY-R28`; no mode here has a cost, so `SVY-R33` adds nothing); enactable only with at least two legal modes (`SVY-R29`); the chosen modes resolve in printed order, whatever order they were picked in (`SVY-R31`); all chosen modes are one stack item.
  - `FR-MOD-001` §6 (line 393): a timed-out mode choice takes the first listed legal modes; timed-out targets take the first legal entity in fixed order (the choosing player's assets first; Headquarters, Reserves, Main Theater, each zone in entry order).
  - `FR-TGT-002`, "Another" Targets (`SVY-R34`, line 370): mode 3's second slot must pick a unit different from its first; the mode is legal only with two different legal units; on a timeout the second slot takes the first legal unit after the first slot's pick.
  - `FR-TGT-003`, Targets Removed During Resolution (`SVY-R25`, line 377) and A Chosen Asset That Deals Damage (`SVY-R27`, line 378): a unit an earlier mode removed is no longer that target; a removed dealer deals nothing and no damage event is emitted; the dealer's current power is used, never last-known.
  - Separate clauses (`SVY-R35`): the slots of different modes are independent, so two modes may name the same unit.
  - `FR-KWD-002h` (Ward): an opposing Ward unit is never a legal target, in any slot, including the dealer slot. "Chosen unit" has no allegiance, so allied units are legal, and no structure or Headquarters is.
  - `FR-ZON-004` / `FR-ZON-005` (lines 225, 236): a face-down Scratch Company unit returned to its owner's hand or put into its deck reverts to its printed card there.
- Shapes it uses, all `done`: `Effect.Modes` with `ChooseN` and the `DecideMode` decision (T-0083); `ActEnact` and `StackManeuver` (T-0116); `SVY-R29` enact legality with `legalModes` (T-0117); `OpKill` (T-0123, *Devilhorn Sanction*); `OpReturnToHand` and `OpPutOnDeck` on `SubjectTarget` (T-0102); `stepAssets` skipping a removed target (T-0107); `Step.Dealer` with `ValueSourcePower` (T-0118); `Target.Another` and the joint `slotsLegal` (T-0140); the progress policy's enacts (T-0121).
- The synthetic Maneuvers in `dealer_test.go` and `another_target_test.go` are named "Ambush" (not "Ambush Tactics") and handed out with `s.give`, so the real definition doesn't change them.

### Definition (the engineer adds it to `pathDefs`)
- *Ambush Tactics*: `Text` holds the five lines. `Effect`: `ChooseN: 2`, and `Modes`, in printed order, each with its own `Targets` and `Steps`. "Chosen unit" is `ability.Target{Filter: ability.Filter{Kinds: ability.KindUnit}}`:
  1. one unit slot; `{Op: ability.OpKill, Subject: ability.SubjectTarget, Target: 0}`.
  2. one unit slot; `{Op: ability.OpReturnToHand, Subject: ability.SubjectTarget, Target: 0}`.
  3. two unit slots, the second with `Another: 1`; `{Op: ability.OpDealDamage, Amount: ability.Value{From: ability.ValueSourcePower}, Subject: ability.SubjectTarget, Target: 1, Dealer: 1}`.
  4. one unit slot; `{Op: ability.OpPutOnDeck, Subject: ability.SubjectTarget, Target: 0}` (`Bottom` false: the top).
- No `Triggers`, `ExtraCosts` or top-level `Targets`/`Steps`. Model the modes on *Rogue Fleshshaper*'s `Modes` (`path.go`), mode 3 on `anotherEffect` (`another_target_test.go:21-26`, which builds on `dealerMode`, `dealer_test.go:47-50`), the kill on *Devilhorn Sanction*, and the return on *Tactical Withdrawal*.
- Remove the name from `cards.NotYetExecutable`.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
To be added to `pkg/engine/cards_path_maneuvers_test.go`, from the CSV text, loading the real card (`s.card("Ambush Tactics")`). Player 0 enacts; player 1 is the opponent. Give player 0 the Tactics to pay. Units have distinct power and enough health that nothing dies unless a test says so. Answer the `DecideMode` decision (pick indices into its offered options) and each `DecideTarget` decision separately; a file-local chooser is fine. For each test, check the mode options the decision offers by their order, not by an index assumed from the full printed list, because only legal modes are offered.
- `TestCard_AmbushTactics_KillsOneUnitAndReturnsAnother` (`FR-MOD-001` §5, `FR-MNV-001`): modes kill and return; opposing unit A is killed (its card in player 1's discard pile), opposing unit B returns to player 1's hand; *Ambush Tactics* is then the top of player 0's discard pile; 4 Tactics are spent.
- `TestCard_AmbushTactics_ChosenUnitDealsItsPowerToAnother` (`SVY-R27`, `FR-TGT-003`): modes 3 and 4; unit X (power 3) is the dealer and Y is dealt to: Y takes exactly one damage event of 3, attributed to X (`Source` is X); unit Z is put on top of its owner's deck (its card is the top card of that player's deck, and Z is off the battlefield). X takes no damage.
- `TestCard_AmbushTactics_DealerUsesItsCurrentPower` (`SVY-R27`): X is player 0's unit, and player 0 first enacts *Chainbreaker's Tenacity* on it (+1 power until end of turn, `s.card("Chainbreaker's Tenacity")`); then the damage *Ambush Tactics*' mode 3 makes X deal to Y is X's printed power plus 1.
- `TestCard_AmbushTactics_ModesResolveInPrintedOrder` (`SVY-R31`): pick mode 4 first, then mode 3, with X both the deck target and the dealer: X still deals its damage to Y, and then X goes on top of its owner's deck.
- `TestCard_AmbushTactics_ReturnedDealerDealsNothing` (`SVY-R25`, `SVY-R27`): modes 2 and 3, with X returned and X the dealer: X is in its owner's hand, Y takes no damage and no damage event is emitted.
- `TestCard_AmbushTactics_SameUnitInTwoModes` (`SVY-R35`, `SVY-R25`): modes kill and put on deck, both on opposing unit A: A is killed (in player 1's discard pile), and the deck step does nothing to it (player 1's deck is unchanged).
- `TestCard_AmbushTactics_TwoDifferentModes` (`SVY-R6`): with two or more units on the battlefield, the mode decision offers four options and asks for exactly two (`Min` = `Max` = 2), without repeats.
- `TestCard_AmbushTactics_AnotherUnitIsADifferentUnit` (`SVY-R34`, `FR-TGT-002`): in mode 3, the dealt-to slot's options don't include the unit chosen as the dealer; they include every other unit.
- `TestCard_AmbushTactics_ChosenUnitIsAnyUnit` (`FR-TGT-001`, `FR-KWD-002h`): a slot's options include allied and opposing units, in either zone, an allied Ward unit and a face-down Scratch Company unit, but no opposing Ward unit, no structure and no Headquarters.
- `TestCard_AmbushTactics_LegalityNeedsUnits` (`SVY-R29`, `SVY-R34`, `FR-MNV-001`), table: with no unit on the battlefield it is not offered (`offersEnact`) and enacting it is rejected (`s.reject`), and it stays in hand; with exactly one unit it is offered, and the mode decision offers three options, without mode 3; with two units it offers all four.
- `TestCard_AmbushTactics_ModeTimeoutTakesTheFirstTwoListed` (`FR-MOD-001` §6): with two or more units, the mode decision times out: kill and return are chosen (the first two listed legal modes), as the target decisions that follow show.
- `TestCard_AmbushTactics_TargetTimeoutsTakeFixedOrder` (`FR-MOD-001` §6, `SVY-R34`): modes 3 and 4, and every target decision times out. Player 0 has allied unit P (entered first) and player 1 has units Q and R, so the fixed order is P, Q, R. P is the dealer and deals its power to Q (the first legal unit after P); then P is put on top of player 0's deck.
- `TestCard_AmbushTactics_FaceDownScratchCompanyGoesBackAsItsCard` (`FR-ZON-005`): mode 4 on a face-down Scratch Company unit puts its printed card on top of its owner's deck.
- `TestCard_AmbushTactics_PaysFourTactics`: with 3 Tactics it is not offered and enacting it is rejected; with 5 it is enacted, leaving 1 Tactics and spending no Requisition.
- **Existing tests:** every other test passes unchanged, including `TestStarterSet_Executable` with the name removed, `TestView_Case8_ExecutableFlag` (seat 0's seed-1 hand still holds unsupported cards), `TestProgress_Case4_Coverage` (now also counts *Ambush Tactics*' enacts, at least 10 in 1000 games), `TestEvents_Case4_EveryActionKind`, and the synthetic "Ambush" tests in `dealer_test.go` and `another_target_test.go`.

## Helpers Needed
- None expected. The tests may add their own setup helper in the test file.

## Allowed Files
- `pkg/engine/cards/path.go` (the definition)
- `pkg/engine/cards/cards.go` (remove the name from `NotYetExecutable`)

## Done When
- [ ] The card tests pass: `go test ./pkg/engine -run TestCard_AmbushTactics`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] The card loads with no unsupported lines (`TestStarterSet_Executable` passes with the name removed)

## Log
- 2026-11-02 lead-architect (wave C session 32): written against the run branch at 2e88e84 (no `main` commits to sync). A deck card (#14 (a)), next on the board after T-0140 (#88's and #91 (3)'s distinctness shape). #91: (1) and (3) are applied; (2) is about a cost on a mode, and no *Ambush Tactics* mode has one; (4) goes with the *Ensnaring Mist* and *Pincer Tactics* tasks. Every shape is `done`, so no ruling is needed. Waiting triggers #6, #7, #8 (this task doesn't touch `deck.go`), #15, #71 (no each-player effect) not fired. Status `tests`.
- 2026-11-03 rules-test-author: wrote 14 tests in `pkg/engine/cards_path_maneuvers_test.go`: TestCard_AmbushTactics_ KillsOneUnitAndReturnsAnother, ChosenUnitDealsItsPowerToAnother, DealerUsesItsCurrentPower, ModesResolveInPrintedOrder, ReturnedDealerDealsNothing, SameUnitInTwoModes, TwoDifferentModes, AnotherUnitIsADifferentUnit, ChosenUnitIsAnyUnit, LegalityNeedsUnits, ModeTimeoutTakesTheFirstTwoListed, TargetTimeoutsTakeFixedOrder, FaceDownScratchCompanyGoesBackAsItsCard, PaysFourTactics. File-local helpers only (atBoard, atModeAsks, atTargetAsks, atOffers, atDeckTop); no helpers requested. All fail now with "needs rules the kernel does not implement yet"; vet and `cardtown lint rules` clean. Left out: nothing. Note: the ChosenUnitIsAnyUnit test also checks that no Headquarters option appears; the mode-3 dealt-to options are checked in AnotherUnitIsADifferentUnit.
