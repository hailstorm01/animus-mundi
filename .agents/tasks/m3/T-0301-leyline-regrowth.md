---
id: T-0301
title: Make Leyline Regrowth executable (up to one unit card back, then draw, wave A, card, flash)
kind: card
tier: flash
wave: A
status: tests-review
branch: m3/autoloop-wave-a/T-0301-leyline-regrowth
depends_on: [T-0300]
claims: [card:Leyline Regrowth]
cards: [Leyline Regrowth]
requirements: [FR-MNV-001, FR-TGT-001, FR-TGT-002, FR-TGT-003]
---

## Goal
*Leyline Regrowth* (a Concordia Maneuver, R0 T2; 2 copies in the starter deck): *"Return up to one chosen unit card from your discard pile to your hand. Draw a card."* It only combines shapes that are `done`: an `UpTo: 1` slot followed by `OpReturnToHand` and `OpDraw` 1 (*Tactical Withdrawal*), with a pile filter (*Allshan's Battle Trance*, *Ruin Partisans*).

Covers:
- The definition in `pkg/engine/cards/concordia.go`, as under Definition below.
- The name off `cards.NotYetExecutable`.

## Inputs
- Card (`data/cards/maneuvers-concordia.csv` line 7).
- Survey row: `docs/design/survey/m3_full_catalog_ability_survey.md` §6, "Concordia: deck" (*Leyline Regrowth*).
- FRD 01: `FR-TRN-003`, `FR-MNV-001` (lines 393-396), `FR-TGT-001` (lines 407-413), `FR-TGT-002` (lines 426-429), `FR-TGT-003` (lines 435-441), `FR-MOD-001` §6 (line 456 on, "Mandatory Targets and Untargeted Selections"), `FR-WIN-001` (lines 218-219).
- Rulings: `SVY-R29` (an "up to" slot never stops an Enact), `SVY-R3` (the target is chosen on declaration), `SVY-R16` (a timed-out "up to N" selects none).
- Design note: `docs/architecture/m3-full-catalog.md`, Boundaries (`pkg/engine/cards`), Deliverable 5.

### Definition (the engineer adds it to `concordiaDefs`)
- `"Leyline Regrowth"`, `Text`: `"Return up to one chosen unit card from your discard pile to your hand. Draw a card."`.
- `Effect`: `Targets` is one slot with `Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit, Zone: ability.ZoneDiscard}` and `UpTo: 1`. `Steps` are `{Op: ability.OpReturnToHand, Subject: ability.SubjectTarget, Target: 0}`, then `{Op: ability.OpDraw, Amount: ability.Value{N: 1}}`.

A lead probe with the definition on the run branch at b99a9e55, since reverted, passed `go tool task check`.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
In a new file `pkg/engine/cards_concordia_regrowth_test.go`, from the CSV text, loading the real card (`s.card`), in the style of the *Tactical Withdrawal* tests (`cards_path_maneuvers_test.go`). Player 0 enacts; player 1 is the opponent. Reuse the existing test helpers (`enactAction`, `enactFor`, `offersEnact`, `refs`, `sortedCopy`, `inHand`, `inDiscardPile`, `lastDiscard`, `s.inDiscard`, `s.onTop`, `s.setPool`). Units in the pile may be synthetic.
- `TestCard_LeylineRegrowth_ReturnsChosenUnitCardThenDraws` (`FR-MNV-001`, `FR-TGT-001`): with two unit cards in the pile, the chosen one leaves the pile for the hand and the other stays. Then the top card of the deck is drawn: the returned card comes before the drawn card in the hand. Leyline Regrowth ends in its owner's discard pile.
- `TestCard_LeylineRegrowth_ChoosingNoneStillDraws` (`FR-TGT-002`): an empty answer with a unit card in the pile returns nothing, and the card is still drawn. A subtest times the choice out with `Answer{Timeout: true}`, with the same outcome (`FR-MOD-001` §6, `SVY-R16`).
- `TestCard_LeylineRegrowth_OnlyOwnUnitCardsOffered` (`FR-TGT-001`, `FR-TGT-002`): with two unit cards, a structure and a Maneuver in player 0's pile, a unit card in player 1's pile and units of both players on the battlefield, the one `DecideTarget` question offers exactly the two unit cards. It takes at most one pick and may take none.
- `TestCard_LeylineRegrowth_EnactableWithNoUnitCard` (`SVY-R29`): with an empty pile, and with only a structure, a Maneuver and an opposing pile's unit card, it is offered, nothing is asked, and a card is drawn.
- `TestCard_LeylineRegrowth_TargetLeftThePileStillDraws` (`FR-TGT-003`): the chosen card leaves the pile after the Maneuver is placed and before it resolves (the observer pattern of `TestCard_AllshansBattleTrance_TargetLeftThePile`, moving it to the bottom of the deck). It is not returned. The card is still drawn, and the Maneuver still goes to the discard pile. Put a known card on top of the deck, so the draw doesn't take the moved card.
- `TestCard_LeylineRegrowth_EmptyDeckReturnsThenLoses` (`FR-WIN-001`): with an empty deck, the chosen card is returned, and player 0 loses the match.
- `TestCard_LeylineRegrowth_PaysTwoTactics` (`FR-MNV-001`): from R3 T5 the pool is R3 T3 afterwards. With R5 T1 it is not offered.
- **Existing tests:** every other test passes unchanged, including `TestCards_Executable` with the name removed.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected.

## Allowed Files
- `pkg/engine/cards/concordia.go` (the definition)
- `pkg/engine/cards/cards.go` (remove the name from `NotYetExecutable`)

## Done When
- [ ] The card tests pass: `go test ./pkg/engine -run TestCard_LeylineRegrowth`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-matrix` exits 0 (just over two minutes: use a timeout above the two-minute default)
- [ ] The card loads with no unsupported lines (`TestCards_Executable` passes with the name removed)

## Log
- 2026-11-25 lead-architect (wave A run, session 3): written against b99a9e55, with the lead's probe above. Board read twice (D2.0): no directive is due here (#185 (3) and (4) are *Will Breaker*'s and *Tempest City Ray*'s), and no waiting trigger has fired. Every part of the Definition exists. *Inspired Salvage* is the nearest sibling, a Soulless card with a different slot, and stays its own task. Set to `tests`.
- 2026-11-25 rules-test-author: wrote `pkg/engine/cards_concordia_regrowth_test.go`: TestCard_LeylineRegrowth_ReturnsChosenUnitCardThenDraws, _ChoosingNoneStillDraws (empty answer and timeout subtests), _OnlyOwnUnitCardsOffered, _EnactableWithNoUnitCard (empty pile, no own unit card), _TargetLeftThePileStillDraws, _EmptyDeckReturnsThenLoses, _PaysTwoTactics. Helpers requested: none (a local `lrChoosePile` lives in the test file). Cases left out: none. `go vet` clean, `cardtown lint rules` passes; the tests fail now with "needs rules the kernel does not implement yet".
