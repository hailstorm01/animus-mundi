---
id: T-0319
title: Make Inspired Salvage executable (up to 2 Mechanized cards from the pile to hand, wave A, card, flash)
kind: card
tier: flash
wave: A
status: tests
branch: m3/autoloop-wave-a/T-0319-inspired-salvage
depends_on: [T-0318]
claims: [card:Inspired Salvage]
cards: [Inspired Salvage]
requirements: [FR-MNV-001, FR-TGT-001, FR-TGT-002, FR-TGT-003, FR-MOD-001]
---

## Goal
*Inspired Salvage* (a Soulless Maneuver, R2 T2; 2 copies in the starter deck): *"Return up to 2 chosen Mechanized cards from your discard pile to your hand."* It only combines shapes that are `done`: *Leyline Regrowth*'s allied pile slot and `OpReturnToHand`, with *Sinuous Growth*'s pile `Subtypes` criterion and *Culverin Brigade*'s `UpTo: 2`.

Covers:
- The definition in `pkg/engine/cards/soulless.go`, as under Definition below.
- The name off `cards.NotYetExecutable`.

## Inputs
- Card (`data/cards/maneuvers-soulless.csv` line 7). Mechanized cards in the Soulless set: `data/cards/units-and-structures-soulless.csv` (for example *Brass Adjutant*, "Mechanized, Cybernetic", and *MAX-8 Autonomous Warsuit*, "Cybernetic, Mechanized").
- Survey row: `docs/design/survey/m3_full_catalog_ability_survey.md` §6, "Soulless: deck" (*Inspired Salvage*, `SVY-R11`).
- FRD 01: `FR-TRN-003` and `FR-MNV-001` (line 393 on, `SVY-R29` at line 395), `FR-TGT-001` (line 407 on; subtypes at line 412), `FR-TGT-002` (line 426 on; `SVY-R11` at line 428), `FR-TGT-003` (line 435 on), `FR-MOD-001` §6 (line 456 on), `FR-ZON-004` (line 279: a card in a pile has its printed characteristics).
- Rulings: `SVY-R11` (up to N chosen entities are distinct), `SVY-R29` (an "up to" slot never stops an Enact), `SVY-R3` (the targets are chosen on declaration), `SVY-R16` (a timed-out "up to N" selects none).
- Design note: `docs/architecture/m3-full-catalog.md`, Boundaries (`pkg/engine/cards`), Deliverable 5.

### Definition (the engineer adds it to `soullessDefs`, after *Hellfire Armorers*)
- `"Inspired Salvage"`, `Text`: `"Return up to 2 chosen Mechanized cards from your discard pile to your hand."`.
- `Effect`: `Targets` is one slot with `Filter{Allegiance: ability.AllegianceAllied, Subtypes: []string{"Mechanized"}, Zone: ability.ZoneDiscard}` and `UpTo: 2`. `Steps` is `{Op: ability.OpReturnToHand, Subject: ability.SubjectTarget, Target: 0}`.

A lead probe with the definition on the run branch at 009107e7, since reverted, passed `go tool task check`.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
In a new file `pkg/engine/cards_soulless_salvage_test.go`, from the CSV text, loading the real card (`s.card`), in the style of the *Leyline Regrowth* tests (`cards_concordia_regrowth_test.go`) and *Bring It Down*'s (`cards_soulless_maneuvers_test.go`). Player 0 enacts; player 1 is the opponent. Reuse the existing test helpers (`enactAction`, `enactFor`, `offersEnact`, `refs`, `sortedCopy`, `inHand`, `inDiscardPile`, `lastDiscard`, `s.inDiscard`, `s.onTop`, `s.setPool`). Pile cards should be real Soulless cards (`s.card`), so their printed subtypes are the CSV's.
- `TestCard_InspiredSalvage_ReturnsTwoChosenMechanizedCards` (`FR-MNV-001`, `FR-TGT-001`): with three Mechanized unit cards in player 0's pile, the two chosen leave the pile for the hand and the third stays. Nothing else changes zone, and *Inspired Salvage* ends in its owner's discard pile.
- `TestCard_InspiredSalvage_ChoosingOneReturnsOne` (`FR-TGT-002`): with two Mechanized cards in the pile, one pick returns that one, and the other stays.
- `TestCard_InspiredSalvage_ChoosingNoneReturnsNothing` (`FR-TGT-002`): an empty answer with Mechanized cards in the pile returns nothing, and the Maneuver still goes to the discard pile. A subtest times the choice out with `Answer{Timeout: true}`, with the same outcome (`FR-MOD-001` §6, `SVY-R16`).
- `TestCard_InspiredSalvage_OnlyOwnMechanizedCardsOffered` (`FR-TGT-001`, `FR-TGT-002`): with, in player 0's pile, *Brass Adjutant* (Mechanized listed first), *MAX-8 Autonomous Warsuit* (Mechanized listed second), a non-Mechanized Soulless unit and a Soulless Maneuver; a Mechanized card in player 1's pile; and a Mechanized unit of each player on the battlefield: the one `DecideTarget` question offers exactly the two Mechanized cards in player 0's pile, with `Min` 0 and `Max` 2.
- `TestCard_InspiredSalvage_SameCardCannotBeChosenTwice` (`SVY-R11`): the target question is not `Repeats`. An answer naming the same pile card twice is invalid: one `EventInvalidAnswer` is logged (as `cards_animus_cliff_test.go` counts it), and the `FR-MOD-001` §6 fallback for an "up to" slot selects none (`SVY-R16`, `choose` in `choice.go`), so both cards stay in the pile and nothing moves to the hand.
- `TestCard_InspiredSalvage_EnactableWithNoMechanizedCard` (`SVY-R29`): with an empty pile, and with only a non-Mechanized unit, a Maneuver and an opposing pile's Mechanized card, it is offered, nothing is asked, and it resolves into the discard pile with nothing returned.
- `TestCard_InspiredSalvage_OneTargetLeftThePileTheOtherStillReturns` (`FR-TGT-003`): two Mechanized cards are chosen, and one leaves the pile after the Maneuver is placed and before it resolves (the observer pattern of `TestCard_LeylineRegrowth_TargetLeftThePileStillDraws`, moving it to the bottom of the deck). That one is not returned and stays at the bottom of the deck; the other is returned.
- `TestCard_InspiredSalvage_PaysTwoAndTwo` (`FR-MNV-001`): from R3 T3 the pool is R1 T1 afterwards. With R1 T5, and with R5 T1, it is not offered.
- **Existing tests:** every other test passes unchanged, including `TestCards_Executable` with the name removed.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected.

## Allowed Files
- `pkg/engine/cards/soulless.go` (the definition)
- `pkg/engine/cards/cards.go` (remove the name from `NotYetExecutable`)

## Done When
- [ ] The card tests pass: `go test ./pkg/engine -run TestCard_InspiredSalvage`
- [ ] `go test ./pkg/cardcheck -run TestCards_Executable` passes with the name removed
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-matrix` exits 0 (just over two minutes: use a timeout above the two-minute default)

## Log
- 2026-11-26 lead-architect (wave A run, session 22): written against 009107e7, with the lead's probe above. `main` has no commits the run branch lacks. Board read (D2.0): no directive is due here (#185 has only (4)'s *Awaken Nightmare* half left, wave F), and no waiting trigger has fired: the card has no each-player effect (#71), no cost on a mode (#91 (2)), no deck search (#106), no discard cost (#120), no tuck (#142), no turn-boundary or draw hook (#15 (2)), and no pile round trip (#183 (2): it takes cards out of a pile and never puts one back while an item below targets or tracks it), the mode is still `serial` (#6, #7), and no `needsRework` row exists (#184). Taken next in the survey's Soulless deck order, after *Hellfire Armorers*. Searched `pkg/engine` and the task archives: every part of the Definition exists (`stepRelocate` returns each card of a pile slot through `stepCards`, `Filter.Subtypes` reads a pile card's printed subtypes, and `UpTo` gives one distinct `Min` 0 decision), so no kernel change. Set to `tests`.
- 2026-11-26 rules-test-author: wrote `pkg/engine/cards_soulless_salvage_test.go` (helpers `isEnact`, `isChoose` are local to it): TestCard_InspiredSalvage_ReturnsTwoChosenMechanizedCards, _ChoosingOneReturnsOne, _ChoosingNoneReturnsNothing (empty answer and timeout subtests), _OnlyOwnMechanizedCardsOffered, _SameCardCannotBeChosenTwice, _EnactableWithNoMechanizedCard (empty pile and no own Mechanized card subtests), _OneTargetLeftThePileTheOtherStillReturns, _PaysTwoAndTwo (pays R2 T2 from R3 T3; not offered with R1 T5 or R5 T1). Helpers requested: none. Cases left out: none. `go vet ./pkg/engine` is clean and `cardtown lint rules` passes; the tests fail today because the card is still in `NotYetExecutable`. Set to `tests-review`.
- 2026-11-26 lead-architect: tests returned (D4, round 1). With the lead's probe Definition 5 of the 8 pass, and 3 fail for these reasons:
  1. `ReturnsTwoChosenMechanizedCards` (line 66), `SameCardCannotBeChosenTwice` (line 169) and both subtests of `EnactableWithNoMechanizedCard` (line 195) count the whole hand, but `newScene` deals an opening hand (the probe saw 8 and 6 cards). Compare against the hand before the enact (less the Maneuver itself), or check exactly which cards entered the hand.
  2. `SameCardCannotBeChosenTwice` expects the twice-named card in the hand once. The answer `[i, i]` is not a valid one, so `choose` logs `EventInvalidAnswer` and takes the `FR-MOD-001` §6 fallback, which for an "up to" slot selects none (`SVY-R16`): the card stays in the pile. That expectation was the lead's task text, now corrected under Tests. Keep the `Repeats` check, count one `EventInvalidAnswer`, and assert that both cards stay in the pile and nothing enters the hand.
  The rest checks out against `FR-MNV-001`, `FR-TGT-001` to `003` and `FR-MOD-001` §6: the fixtures' subtypes match the CSV (*Conscript Rifles* is Infantry, *MAX-8* lists Mechanized second), and the offered set, `Min` 0 / `Max` 2, the timeout, the left-the-pile and the cost cases are right. Set to `tests`.
