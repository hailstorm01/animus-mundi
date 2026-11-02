---
id: T-0133
title: Make Disorienting Pulse and Tactical Withdrawal executable (return a chosen unit to its owner's hand, wave C card, flash)
kind: card
tier: flash
wave: C
status: review
branch: m2/autoloop-wave-c/T-0133-return-chosen-unit
depends_on: [T-0116, T-0117, T-0121]
claims: [card:Disorienting Pulse, card:Tactical Withdrawal]
cards: [Disorienting Pulse, Tactical Withdrawal]
requirements: [FR-MNV-001, FR-TGT-001, FR-TGT-002, FR-TGT-003, FR-KWD-002h, FR-ZON-004, FR-ZON-005]
---

## Goal
*Disorienting Pulse* (Animus) and *Tactical Withdrawal* (Path), both non-deck cards kept in this run (#14 (c)), run in the kernel. Each is plain `ability` data: a Maneuver `Effect` with one target slot and an `OpReturnToHand` step on it, and for *Tactical Withdrawal* an "up to one" slot followed by `OpDraw` 1. No new shape.

## Inputs
- Cards:
  - `data/cards/maneuvers-animus.csv` line 2: *Disorienting Pulse*, Animus Maneuver, cost 0 Requisition / 1 Tactics: "Return chosen opposing unit in the Main Theater to its owner's hand."
  - `data/cards/maneuvers-path.csv` line 4: *Tactical Withdrawal*, Path Maneuver, cost 0 Requisition / 1 Tactics: "Return up to one chosen allied unit to its owner's hand. Draw a card."
- Survey rows: `docs/design/m2_starter_set_ability_survey.md` line 125 (*Disorienting Pulse*: `RETURN_TO_HAND` chosen opposing unit in Main Theater) and line 169 (*Tactical Withdrawal*: `RETURN_TO_HAND` up to one chosen allied unit; `DRAW_CARD 1`; `minCount: 0`).
- FRD 01:
  - `FR-MNV-001` (lines 332-334): enacted from Hand, costs paid on declaration, the effect runs when it resolves, then the card goes to its owner's Discard pile. It can be enacted only if each exactly-one target slot has a legal entity on declaration; an "up to N" slot never stops it (`SVY-R29`).
  - `FR-TGT-001` (line 347 on): a target must meet every criterion. "Chosen opposing unit in the Main Theater" is a unit an opponent controls, in that player's Main Theater, never one in Reserves, a structure, a Headquarters or an allied unit. "Chosen allied unit" is a unit its controller controls in either battlefield zone, face-up or a face-down Scratch Company (a face-down Scratch Company is a unit, `FR-ZON-005`), never a structure or a Headquarters.
  - `FR-TGT-002` (line 361 on) and `SVY-R11`: "up to one" may choose none.
  - `FR-KWD-002h` (line 570): Ward stops only opponents. An opposing Ward unit can't be chosen for *Disorienting Pulse*; an allied Ward unit can be chosen for *Tactical Withdrawal*.
  - `FR-TGT-003` (line 367): a target illegal at resolution is skipped, and the rest of the effect still runs.
  - `FR-ZON-004` (line 219) and `FR-ZON-005` (line 228): a unit that leaves the battlefield loses its state; a face-down Scratch Company goes to its owner's hand as its printed card.
- Shapes it uses, all `done`: `Def.Effect`, `ActEnact` and `StackManeuver` (T-0116), the `SVY-R29` enact legality (T-0117), `Effect.Targets` with an allegiance, kind and zone `Filter` (`matches` reads `Filter.Zone` through `deployedTo`, `targeting.go`), `Target.UpTo`, `OpReturnToHand` on `SubjectTarget` (`stepRelocate`, `effects.go`), `OpDraw`, Ward in `legalTarget` (`targeting.go`).
- Rulings: none beyond the FRD text above. Each card has one target slot, so neither distinctness between slots (design note Open Question 16) nor repeated targets (#59 (2)) arises.

### Definitions (the engineer adds them)
- *Disorienting Pulse*, in `animusDefs` (`pkg/engine/cards/animus.go`): `Text` holds its line. `Effect`:
  - `Targets`: one, `{Filter: ability.Filter{Allegiance: ability.AllegianceOpposing, Kinds: ability.KindUnit, Zone: ability.ZoneMainTheater}}`.
  - `Steps`: one, `{Op: ability.OpReturnToHand, Subject: ability.SubjectTarget, Target: 0}`.
- *Tactical Withdrawal*, in `pathDefs` (`pkg/engine/cards/path.go`): `Text` holds its line. `Effect`:
  - `Targets`: one, `{Filter: ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit}, UpTo: 1}`.
  - `Steps`: `{Op: ability.OpReturnToHand, Subject: ability.SubjectTarget, Target: 0}`, then `{Op: ability.OpDraw, Amount: ability.Value{N: 1}}`.
- No `Triggers` or `Modes`. Remove both names from `cards.NotYetExecutable`.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
From the CSV text, loading the real cards (`s.card`). Player 0 enacts; player 1 is the opponent. Give player 0 enough Tactics to pay (`enactFor` in `cards_animus_maneuvers_test.go` does). Targets are answered with a scripted decision.

*Disorienting Pulse*, appended to `pkg/engine/cards_animus_maneuvers_test.go`:
- `TestCard_DisorientingPulse_ReturnsChosenOpposingUnitToItsOwnersHand`: a chosen opposing unit in the Main Theater leaves the battlefield and is in player 1's hand once the Maneuver resolves; the Maneuver ends in player 0's discard pile.
- `TestCard_DisorientingPulse_OnlyOpposingMainTheaterUnitsAreOffered`: with two opposing units in the Main Theater, an opposing unit in Reserves, an opposing structure, an allied unit in the Main Theater and both Headquarters, the target decision offers exactly the two opposing Main Theater units.
- `TestCard_DisorientingPulse_WardUnitCannotBeChosen`: an opposing Ward unit in the Main Theater is not offered while two plain opposing Main Theater units are (two, so a decision is asked; see T-0123's round 2).
- `TestCard_DisorientingPulse_NotEnactableWithoutLegalTarget`: with opposing units only in Reserves, or only an opposing Ward unit in the Main Theater, or only allied units in the Main Theater, enacting it is not a legal action and the card stays in hand (`SVY-R29`). With one plain opposing unit in the Main Theater, it is.
- `TestCard_DisorientingPulse_FaceDownScratchCompanyReturnsAsItsCard`: an opposing face-down Scratch Company unit in the Main Theater can be chosen, and its printed card instance goes to player 1's hand (`FR-ZON-004`, `FR-ZON-005`).
- `TestCard_DisorientingPulse_PaysOneTactics`: enacting it spends 1 Tactics and no Requisition.

*Tactical Withdrawal*, appended to `pkg/engine/cards_path_maneuvers_test.go`:
- `TestCard_TacticalWithdrawal_ReturnsChosenAlliedUnitThenDraws`: a chosen allied unit is in player 0's hand once the Maneuver resolves, player 0 has drawn the top card of their deck, and the Maneuver ends in player 0's discard pile. The unit is returned before the draw (the hand gains the unit, then the deck's top card; check the order through the log or by hand order, whichever the engine records).
- `TestCard_TacticalWithdrawal_ChoosingNoneStillDraws`: with an allied unit on the battlefield, answering the up-to-one decision with no pick returns nothing and still draws one card.
- `TestCard_TacticalWithdrawal_EnactableWithNoAlliedUnit`: with no allied unit (only opposing units and an allied structure), enacting it is a legal action (`SVY-R29`, an "up to" slot never stops it), no target decision is asked, nothing leaves the battlefield and player 0 draws one card.
- `TestCard_TacticalWithdrawal_OnlyAlliedUnitsAreOffered`: with two allied units (one in Reserves, one in the Main Theater), an allied structure, an opposing unit and both Headquarters, the decision offers exactly the two allied units.
- `TestCard_TacticalWithdrawal_AlliedWardUnitCanBeChosen`: an allied unit with Ward is offered and, when chosen, returned (`FR-KWD-002h`: Ward stops only opponents).
- `TestCard_TacticalWithdrawal_FaceDownScratchCompanyReturnsAsItsCard`: an allied face-down Scratch Company unit in Reserves can be chosen, and its printed card instance goes to player 0's hand.
- `TestCard_TacticalWithdrawal_PaysOneTactics`: enacting it spends 1 Tactics and no Requisition.
- Not a card test: a target that turns illegal between placement and resolution (`FR-TGT-003`), for the reason T-0123 gives (nothing in a scene runs between a Maneuver's placement and its resolution; the rule is pinned on synthetic cards by `TestFR_TGT_003_UpToResolvesForTheSurvivingChosenEntity` and T-0053).
- **Existing tests:** every other test passes unchanged, including `TestStarterSet_Executable` with both names removed, `TestFR_CRD_101_ManeuverCarriesCSVCharacteristics` (it reads *Disorienting Pulse*'s CSV columns only), `TestProgress_Case4_Coverage` and `TestEvents_Case4_EveryActionKind`.

## Helpers Needed
- None expected. The tests use existing helpers (`newScene`, `put`, `putScratch`, `give`, `card`, `do`, `script`, `asked`), those in `maneuver_test.go` (`enactAction`, `offersEnact`, `lastDiscard`), `enactFor` in `cards_animus_maneuvers_test.go`, and `chooseAsset`, `targetOffers` in `cards_path_maneuvers_test.go`; a test file may add its own setup helpers.

## Allowed Files
- `pkg/engine/cards/animus.go` (*Disorienting Pulse*'s definition)
- `pkg/engine/cards/path.go` (*Tactical Withdrawal*'s definition)
- `pkg/engine/cards/cards.go` (remove both names from `NotYetExecutable`)

## Done When
- [x] The card tests pass: `go test ./pkg/engine -run "TestCard_DisorientingPulse|TestCard_TacticalWithdrawal"`
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks
- [x] The cards load with no unsupported lines (`TestStarterSet_Executable` passes with both names removed)

## Log
- 2026-11-02 lead-architect: wave C run, session 21. No `main` commits to sync (`git rev-list autoloop/m2-wave-c..main` is empty). Not `WAVE-COMPLETE`: 15 deck cards are on `NotYetExecutable`, two of them in this wave's `Cards` (*Ambush Tactics*, *Steel Rain*), plus the kept *Disorienting Pulse*, *Conversion Spike*, *Ensnaring Mist*, *Chainbreaker's Tenacity*, *Fortify*, *Tactical Withdrawal*, *Pincer Tactics*. Directives read on the run branch: no `open: before` directive. #14 (a) puts deck cards first, but both remaining wave C deck cards wait on a ruling asked before their next task: *Steel Rain*'s per-discarded-card targets need #59 (2), and *Ambush Tactics* needs #61's fifth question and #88's distinctness ruling. As in run rows 4 and 5, work that needs no ruling goes first; under #14 (c) a kept non-deck card whose shapes all exist may be built. *Disorienting Pulse* is the first kept card in the survey's wave C order, and *Tactical Withdrawal* uses the same shapes (`OpReturnToHand` on a target slot, plus `OpDraw`), so the two share one card task (D2.4: a few cards with identical shapes). *Ensnaring Mist* and *Pincer Tactics* stay out (two slots that could name the same unit, Open Question 16). Waiting triggers checked: #6, #7 (mode is `serial`), #8 (no `deck.go` change), #15 (no turn hook), #59 (2) (no repeated targets), #61's fifth question (no modal card in `cards:`), #71 ("Draw a card" is the controller's, not each player's), #88's distinctness (one slot per card). None has fired. Duplicates searched: no new code; `OpReturnToHand` on `SubjectTarget` is served by `stepRelocate`, the zone criterion by `matches`, and "up to one" by `Target.UpTo`. Three non-test files. Status `tests`; dispatching `rules-test-author`.
- 2026-11-02 rules-test-author: wrote 13 tests, no helpers requested. `cards_animus_maneuvers_test.go`: `TestCard_DisorientingPulse_{ReturnsChosenOpposingUnitToItsOwnersHand, OnlyOpposingMainTheaterUnitsAreOffered, WardUnitCannotBeChosen, NotEnactableWithoutLegalTarget, FaceDownScratchCompanyReturnsAsItsCard, PaysOneTactics}`. `cards_path_maneuvers_test.go`: `TestCard_TacticalWithdrawal_{ReturnsChosenAlliedUnitThenDraws, ChoosingNoneStillDraws, EnactableWithNoAlliedUnit, OnlyAlliedUnitsAreOffered, AlliedWardUnitCanBeChosen, FaceDownScratchCompanyReturnsAsItsCard, PaysOneTactics}`. All 13 fail now (the cards are not executable yet); `go vet` and `cardtown lint rules` are clean. The opposing face-down Scratch Company is moved to the Main Theater with `moveTo`. Return-before-draw is checked by hand order (unit, then the drawn card last). Left out as the task says: the FR-TGT-003 resolution-time fizzle.
- 2026-11-02 lead-architect: tests approved (08efdf0). The thirteen check their outcomes against the card text and FRD 01. *Disorienting Pulse*: the chosen opposing Main Theater unit goes to its owner's hand and the unchosen one stays; the decision offers exactly the opposing Main Theater units, never one in Reserves, a structure, an allied unit or a Headquarters (`FR-TGT-001`); an opposing Ward unit isn't offered beside two plain ones (`FR-KWD-002h`); not enactable with opposing units only in Reserves, only a Ward one or only allies, and enactable with one plain opposing Main Theater unit (`SVY-R29`); an opposing face-down Scratch Company moved to the Main Theater returns as its card (`FR-ZON-004`, `FR-ZON-005`); 1 Tactics. *Tactical Withdrawal*: the chosen allied unit returns, then the top card is drawn (hand order: the unit, then the drawn card last); choosing none still draws (`FR-TGT-002`); enactable with no allied unit, asking nothing and still drawing (`SVY-R29`); the decision offers exactly the allied units in both zones; an allied Ward unit can be chosen (`FR-KWD-002h`); an allied face-down Scratch Company returns as its card; 1 Tactics. The asserts read zones, hands, pools, offered options and decisions, not internals. No existing test edited. All thirteen fail now on the missing definitions (eleven on "needs rules the kernel does not implement yet", two on Enact not offered). Checked against a discarded trial of the task's Definitions: with it, all of `go test ./pkg/engine` passes. Status `ready`; dispatching `kernel-engineer` (flash) through `agy`.
- 2026-11-02 kernel-engineer: added ability definitions for Disorienting Pulse in animusDefs (one target slot with AllegianceOpposing, KindUnit, ZoneMainTheater; OpReturnToHand step on SubjectTarget 0) and Tactical Withdrawal in pathDefs (one target slot with AllegianceAllied, KindUnit, UpTo: 1; OpReturnToHand step on SubjectTarget 0 followed by OpDraw 1). Reused existing ability shapes OpReturnToHand, stepRelocate, matches, and OpDraw. Removed both cards from cards.NotYetExecutable. Ran go test ./pkg/engine -run "TestCard_DisorientingPulse|TestCard_TacticalWithdrawal" (13 tests pass), go tool task check (fmt-check, vet, all tests, cardtown lint --all clean), go tool task sim (200 random playouts clean, all invariants hold), and TestStarterSet_Executable (clean). Status review.
