---
id: T-0125
title: Make Improvised Orders executable (draw 3, then discard a card, wave C card, flash)
kind: card
tier: flash
wave: C
status: review
branch: m2/T-0125-improvised-orders
depends_on: [T-0116, T-0117, T-0121]
claims: [card:Improvised Orders]
cards: [Improvised Orders]
requirements: [FR-MNV-001, FR-MOD-001, FR-WIN-001]
---

## Goal
*Improvised Orders* (Path starter deck, 1 copy) runs in the kernel. It is plain `ability` data in `pkg/engine/cards/path.go`: a Maneuver `Effect` with no targets and two steps, `OpDraw` 3 then `OpDiscard` 1, the same steps *Chainbreaker Irregulars*' trigger uses. No new shape.

## Inputs
- Card (`data/cards/maneuvers-path.csv` line 7): *Improvised Orders*, Path Maneuver, cost 0 Requisition / 2 Tactics, one rules-text line:
  - "Draw 3 cards, then discard a card."
- Survey row: `docs/design/m2_starter_set_ability_survey.md` line 171: `DRAW_CARD 3`, `DISCARD_CARD 1`.
- FRD 01:
  - `FR-MNV-001` (lines 332-334): enacted from Hand, costs paid on declaration, the effect runs when it resolves, then the card goes to its owner's Discard pile. It has no target slot, so `SVY-R29` never stops it being enacted. While it resolves it is on the Resolution Stack, not in Hand, so it can't be the discarded card.
  - `FR-MOD-001` §1 (line 380): choosing a card from your own Hand is not an inspection. With exactly one card in Hand when the discard happens, it is discarded without asking (`SVY-R16`).
  - `FR-MOD-001` §6 (line 385 on), *Choosing Cards from Your Own Hand*: a timeout discards the card most recently added to the Hand, which here is the third card drawn.
  - `FR-WIN-001` (lines 158-159): a draw from an empty deck does nothing, and the player loses at the next state-based death sweep, the next time no item is resolving. So with fewer than 3 cards in the deck, the Maneuver draws what is there, still discards a card, and the loss comes at the sweep after it resolves.
- Shapes it uses, all `done`: `Def.Effect`, `ActEnact` and `StackManeuver` (T-0116), the `SVY-R29` enact legality (T-0117), `OpDraw` (`stepDraw`, `draw` in `zones.go`), `OpDiscard` (`stepDiscard`, `effects.go`; the one `discardFromHand`, `zones.go`, directive #64), both on `SubjectController`.
- Rulings: `SVY-R16` as above. No other.

### Definition (the engineer adds it to `pathDefs`)
- *Improvised Orders*: `Text` holds its line. `Effect`:
  - `Steps`: two, in order: `{Op: ability.OpDraw, Amount: ability.Value{N: 3}}`, then `{Op: ability.OpDiscard, Amount: ability.Value{N: 1}}` (both `SubjectController`, the zero value), as in *Chainbreaker Irregulars*.
- No `Targets`, `Triggers` or `Modes`.
- Remove the name from `cards.NotYetExecutable`.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
To be added to `pkg/engine/cards_path_maneuvers_test.go`, from the CSV text, loading the real card (`s.card`). Player 0 enacts; player 1 is the opponent. Give player 0 enough Tactics to pay (`enactFor` in `cards_animus_maneuvers_test.go` does). Set the decks so the drawn cards are known. The discard is answered with a scripted decision.
- `TestCard_ImprovisedOrders_DrawsThreeThenDiscardsTheCardPicked`: player 0 has one other card in hand and at least 3 in deck. After it resolves, the top 3 deck cards were drawn, a `DecideHand` decision was asked offering all 4 hand cards (the earlier card and the 3 drawn), the picked card is in player 0's discard pile, the other 3 are in hand, and *Improvised Orders* is in its owner's discard pile.
- `TestCard_ImprovisedOrders_MayDiscardACardJustDrawn`: picking one of the drawn cards discards it, and the earlier hand card stays.
- `TestCard_ImprovisedOrders_TimeoutDiscardsTheLastCardDrawn`: with a scripted timeout, the third card drawn is discarded (`FR-MOD-001` §6, `SVY-R16`).
- `TestCard_ImprovisedOrders_ItselfIsNeverOffered`: *Improvised Orders* is never among the discard decision's options, and it ends in the discard pile as its own resolution, not as the discarded card (the discard event names another card).
- `TestCard_ImprovisedOrders_ShortDeckDrawsWhatIsThereThenLoses`: with 1 card in player 0's deck and none in hand besides the Maneuver, it draws that card, discards it without asking (the only card in hand, `FR-MOD-001` §1), and player 0 loses at the death sweep after it resolves (`FR-WIN-001`), not before the discard.
- `TestCard_ImprovisedOrders_OnlyTheControllersHandAndDeckAreUsed`: player 1's hand and deck are unchanged.
- `TestCard_ImprovisedOrders_PaysTwoTactics`: enacting it spends 2 Tactics and no Requisition, and it is offered with an empty battlefield (no target slot, `SVY-R29`).
- **Existing tests:** every other test passes unchanged, including `TestStarterSet_Executable` with the name removed, `TestProgress_Case4_Coverage` and `TestEvents_Case4_EveryActionKind`.

## Helpers Needed
- None expected. The tests use existing helpers (`newScene`, `give`, `card`, `do`, `script`, `asked`), those in `maneuver_test.go` (`enactAction`, `offersEnact`, `lastDiscard`) and `enactFor` in `cards_animus_maneuvers_test.go`; the test file may add its own setup helpers.

## Allowed Files
- `pkg/engine/cards/path.go` (the definition)
- `pkg/engine/cards/cards.go` (remove the name from `NotYetExecutable`)

## Done When
- [x] The card tests pass: `go test ./pkg/engine -run TestCard_ImprovisedOrders`
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks
- [x] The card loads with no unsupported lines (`TestStarterSet_Executable` passes with the name removed)

## Log
- 2026-10-31 lead-architect: wave C run, session 11. No `main` commits to sync (`git log autoloop/m2-wave-c..main` is empty). Not `WAVE-COMPLETE`: 20 deck cards are on `NotYetExecutable`, 7 of them in this wave's `Cards` (*Rapturous Sermon*, *Animoceps Spores*, *Rhapsody's Embrace*, *Funeral Fungus*, *Ambush Tactics*, *Steel Rain*, *Improvised Orders*). Directives read on the run branch: no `open: before` directive; #90 applied by T-0124. #14 (a): deck cards first. *Improvised Orders* is a deck card whose every shape is `done` (the steps of *Chainbreaker Irregulars*' trigger, on a Maneuver), so it needs no shape task; *Animoceps Spores* is the other such deck card and stays for a later session. No non-deck card shares this card's shapes, so none joins under #14 (c). #20 (b) stays open for *Steel Rain*'s discard-cost task. Waiting triggers checked: #8 (no `deck.go` change), #15 (a draw step is not a draw hook), #58 (no exhaust in `claims:`), #59 (2) (no target slot), #61's fifth question (no modal card in `cards:`), #71 (acts on its controller only, not on each player), #88's distinctness (no slot, no *Ambush Tactics*). None has fired. Duplicates searched: no new code; `OpDraw` and `OpDiscard` go through the one `draw` and the one `discardFromHand`. Two non-test files. Status `tests`; dispatching `rules-test-author`.
- 2026-10-31 rules-test-author: wrote 7 tests in `pkg/engine/cards_path_maneuvers_test.go` (with setup helpers `improvisedScene`, `handDiscards` in that file): `TestCard_ImprovisedOrders_DrawsThreeThenDiscardsTheCardPicked`, `_MayDiscardACardJustDrawn`, `_TimeoutDiscardsTheLastCardDrawn`, `_ItselfIsNeverOffered`, `_ShortDeckDrawsWhatIsThereThenLoses`, `_OnlyTheControllersHandAndDeckAreUsed`, `_PaysTwoTactics`. All fail now because the card is on `NotYetExecutable` (Enact rejected: needs rules the kernel does not implement yet). Helpers requested: none. Cases left out: none. `go vet` clean, `cardtown lint rules` passes. Status `tests-review`.
- 2026-10-31 lead-architect: tests approved. The seven check their outcomes against the card text and FRD 01: the top 3 deck cards are drawn and one `DecideHand` decision offers the whole hand, the earlier card and the 3 drawn; the picked card goes to the discard pile and the rest stay; a card just drawn may be discarded; a timeout discards the third card drawn (`FR-MOD-001` §6, `SVY-R16`); the Maneuver, on the Resolution Stack, is never offered or discarded from hand and ends in its owner's discard pile (`FR-MNV-001`); with a 1-card deck the drawn card is discarded without asking (`FR-MOD-001` §1) and player 0 loses afterwards, the discard event before the match end (`FR-WIN-001`); the opponent's piles are untouched and the opponent is never asked; 2 Tactics and no Requisition, offered on an empty battlefield (`SVY-R29`). The asserts read zones, pools, decisions and the log, not internals. No existing test edited. All seven fail on the missing definition (re-run in a review worktree). Status `ready`; dispatching `kernel-engineer` (flash) through `agy`.
- 2026-10-31 kernel-engineer: recovered from earlier stopped run (branch had in-progress commit, worktree clean). Added Improvised Orders definition to pathDefs in pkg/engine/cards/path.go with two steps: OpDraw 3 then OpDiscard 1, matching Chainbreaker Irregulars; removed Improvised Orders from NotYetExecutable in pkg/engine/cards/cards.go. Reused existing OpDraw, OpDiscard, and Effect on Maneuver; no new shape or helpers. Results: all 7 card tests pass (TestCard_ImprovisedOrders_*); TestStarterSet_Executable passes; go tool task check passes (fmt-check, vet, all tests, cardtown lint --all); go tool task sim runs clean (Path starter deck now 28/50 executable). Status review.
