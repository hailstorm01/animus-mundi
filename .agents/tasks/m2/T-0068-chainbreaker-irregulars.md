---
id: T-0068
title: Make Chainbreaker Irregulars executable (wave B card, flash)
kind: card
tier: flash
wave: B
status: in-progress
branch: m2/T-0068-chainbreaker-irregulars
depends_on: [T-0067]
claims: [card:Chainbreaker Irregulars]
cards: [Chainbreaker Irregulars]
requirements: []
---

## Goal
*Chainbreaker Irregulars* (Path unit, four copies in the Path starter deck) runs in the kernel: "Unbowed: Draw a card, then discard a card." It is plain `ability` data in `pkg/engine/cards/path.go`.

## Inputs
- Card (`data/cards/*.csv`): Path unit, Infantry, Chainbreaker, cost 3, deploys to Reserves; rules text `Unbowed: Draw a card, then discard a card.`
- Survey row: `docs/design/m2_starter_set_ability_survey.md` line 146.
- Shapes it uses: `ability.HookDamaged` (Unbowed, T-0050), `OpDraw` (A1), `OpDiscard` with the `DecideHand` decision (T-0067).
- Rulings: `SVY-R16` (one card fast-paths, a timeout discards the card just drawn), `SVY-R13` (lethal Unbowed is placed after the death sweep and still resolves).

### Definition (the engineer adds it to `pathDefs`)
- One trigger: `Hook: ability.HookDamaged`, two steps in order: `{Op: ability.OpDraw, Amount: ability.Value{N: 1}}`, then `{Op: ability.OpDiscard, Amount: ability.Value{N: 1}}`.
- `Text` holds the printed rules text, one entry.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
All in `pkg/engine/cards_path_irregulars_test.go`; Irregulars sits in player 1's Main Theater and player 0 attacks it.
- `TestCard_ChainbreakerIrregulars_EntersReservesAsTwoFiveForThreeRequisitionAndDoesNothing`: 2/5, 3 Requisition, Reserves, nothing drawn or asked on deploy.
- `..._DamageDrawsThenDiscardsTheCardTheControllerPicks`: one `DecideHand` to the controller over 3 options (drawn card included); the pick is discarded; the opponent is not asked.
- `..._MayDiscardTheCardJustDrawn`: the drawn card is a legal pick.
- `..._TimeoutDiscardsTheCardJustDrawn` (SVY-R16, FR-MOD-001).
- `..._EmptyHandDiscardsTheDrawnCardWithoutAsking` (SVY-R16 fast path).
- `..._LethalDamageStillDrawsAndDiscardsAfterTheDeathSweep` (SVY-R13).
- `..._EachSeparateDamageInstanceFiresAgain`: two attackers, two draws and two discards.
- `..._OnlyTheControllersHandAndDeckAreUsed`.

## Helpers Needed
None. The tests use only existing helpers (`newScene`, `give`, `onTop`, `put`, `script`, `vanilla`, `deployPath`, `assetOf`, `inZone`).

## Allowed Files
- `pkg/engine/cards/path.go`
- `pkg/engine/cards/cards.go` (remove the name from `NotYetExecutable`)

## Done When
- [ ] The card's tests pass: `go test ./pkg/engine -run 'TestCard_ChainbreakerIrregulars'`
- [ ] `TestDefs_MatchRulesText` and `TestStarterSet_Executable` pass unchanged
- [ ] `go tool task check` passes
- [ ] `go tool task sim` runs clean; report the before and after summary lines in the Log

## Log
- 2026-10-26 lead-architect (wave run B, session 1): written with T-0067 and left `queued` until T-0067 is `done`.
- 2026-10-26 lead-architect (wave run B, session 2): synced `main` into the run branch (1a3ef25). No `open: before` directive; #51, #58 and #59 don't apply (no "if you do", exhaust or pile claims). Status `tests`; branch created from `autoloop/m2-wave-b`; dispatched `rules-test-author`.
- 2026-10-26 rules-test-author: wrote 8 `TestCard_ChainbreakerIrregulars_*` tests (listed under Tests). No helpers requested. `go vet` is clean, the tests fail as expected (card not yet executable), and `cardtown lint rules` passes. Left out: an empty-deck draw (FRD 01 reviewed scope does not settle it here) and an Armor-absorbs-all case (already pinned for Unbowed generally by `TestFR_REP_004_UnbowedFiresWhenArmorAbsorbsAllTheDamage`).
- 2026-10-26 lead-architect (tests-review): approved. The eight tests match `FR-KWD-002a` (which names this card's lethal case), `SVY-R13` and `SVY-R16`. Seven fail before code for the right reason (not executable, no Unbowed); `OnlyTheControllersHandAndDeckAreUsed` is a control and passes already. The empty-deck draw is `OpDraw`'s behavior, not this card's, so leaving it out is right. Status `ready`; dispatching `kernel-engineer` on Flash.
