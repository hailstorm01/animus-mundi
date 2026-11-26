---
id: T-0315
title: Make Dolus Refiner executable (exhaust: draw 2, then discard 2; wave A, card, flash)
kind: card
tier: flash
wave: A
status: in-progress
branch: m3/autoloop-wave-a/T-0315-dolus-refiner
depends_on: [T-0314]
claims: [card:Dolus Refiner]
cards: [Dolus Refiner]
requirements: [FR-TRN-003, FR-SET-004, FR-ACT-001, FR-ACT-002, FR-MOD-001, FR-WIN-001, FR-LAY-002, FR-ZON-004, FR-TRN-003a]
---

## Goal
*Dolus Refiner* (Soulless structure, Industrial, R4 T0, Main Theater, power 0, health 9; two copies in `data/decks/soulless-starter.txt`): *"Exhaust: Draw 2 cards, then discard 2 cards."* It only combines shapes that are `done`: one `Activated` ability whose only cost is `CostExhaustSelf` (*Icewind Regent*, *Stormbolt Ray*), with no target, and the two steps of *Improvised Orders* (`pkg/engine/cards/path.go`): `OpDraw` then `OpDiscard`, both on the controller, here with `Amount` 2 each. The discard is the existing `stepDiscard` → `discardFromHand` path: exactly 2 cards of the controller's choice from their own hand, or the whole hand if it holds fewer; no "if you do" follows, so #187 (2)'s all-or-nothing discard rule (wave C) doesn't apply.

Covers:
- The definition in `soullessDefs` (`pkg/engine/cards/soulless.go`), as under Definition below.
- The name off `cards.NotYetExecutable`.

## Inputs
- Card (`data/cards/units-and-structures-soulless.csv` line 22): Soulless structure, Industrial, Requisition cost 4, Tactics cost 0, Main Theater, power 0, health 9.
- Survey row: `docs/design/survey/m3_full_catalog_ability_survey.md` §6, "Soulless: deck" (*Dolus Refiner*, line 172).
- FRD 01 (`docs/frd/01_Game_Rules_FRD.md`): `FR-TRN-003` (line 394), `FR-SET-004` Immediate Readiness (line 203), `FR-ACT-001` and `FR-ACT-002` (lines 388-391, with the clean-stack requirement), `FR-MOD-001` §1 Single-Option Fast-Path and §6 timeouts (lines 449-465), `FR-WIN-001` (line 219, empty-deck loss at the next death sweep), `FR-LAY-002` (lines 534-538), `FR-ZON-004` (line 279, last-known state; discard-from-hand triggers), `FR-TRN-003a` (a free deploy, for *Eerie Remnant*).
- Rulings that apply: `SVY-R16` (choosing from your own hand is not an inspection, so a hand of exactly 2 is picked without asking; timeouts are never random), `SVY-R50` (a hand timeout takes the cards most recently added to that hand), `KIM-S04` (empty-deck loss at the death sweep).
- Design note: `docs/architecture/m3-full-catalog.md`, Boundaries (`pkg/engine/cards`), Deliverable 5.

### Definition (the engineer adds it to `soullessDefs`, after *Dolus Macrocannon*)
- `"Dolus Refiner"`, `Text`: `"Exhaust: Draw 2 cards, then discard 2 cards."` (one string).
- One `Activated`: `Costs` is `{Kind: ability.CostExhaustSelf}`. `Effect` has no `Targets`, and `Steps`:
  - `{Op: ability.OpDraw, Amount: ability.Value{N: 2}}`
  - `{Op: ability.OpDiscard, Amount: ability.Value{N: 2}}`

A lead probe with the definition on the run branch at 607c9752, since reverted, passed `go tool task check`, `go tool task sim` (P0 58.5% / P1 39.5% / draws 2.0%, unchanged) and `go tool task sim-matrix` (exit 0).

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
In a new file `pkg/engine/cards_soulless_refiner_test.go`, from the CSV text, loading the real card (`s.card("Dolus Refiner")`), in the style of `cards_soulless_macrocannon_test.go` and `cards_concordia_stormbolt_test.go` (an activated `CostExhaustSelf` ability on a structure) and of the *Improvised Orders* tests in `cards_path_maneuvers_test.go` (draw, then discard from hand; `handDiscards`). Player 0 controls the Refiner and is the active player, unless a case says otherwise. Reuse the existing test helpers (`s.put`, `s.give`, `s.onTop`, `s.setPool`, `s.script`, `s.reject`, `s.asked`, `askedKind`, `activate`, `offersActivate`, `vanilla`, `handDiscards`, `inDiscardPile`, `remnantChooser`). The other cards may be synthetic. Unless a case says otherwise, player 0's hand holds two known synthetic cards (H1, H2) and the top of their deck three known synthetic cards (D1, D2, D3).

Deploying:
- `TestCard_DolusRefiner_EntersMainTheaterAsZeroNine` (`FR-TRN-003`): deployed from hand with exactly R4 it is a Readied 0/9 structure in player 0's Main Theater, and the pool is empty. Deploying it to Reserves is rejected, and with R3 it can't be deployed; in both cases the card stays in hand.

Activating (a Readied Refiner put in player 0's Main Theater, pool R0 T0, unless the case says otherwise):
- `TestCard_DolusRefiner_DrawsTwoThenDiscardsTheTwoPicked` (`FR-ACT-001`, `FR-MOD-001`): with pool R0 T0 the ability is offered. Activating it draws D1 and D2 (D3 stays on top of the deck), then asks player 0 exactly one `DecideHand` decision with `Min` 2 and `Max` 2 offering exactly H1, H2, D1 and D2. Picking H1 and H2 discards them (`handDiscards` gives H1 then H2, both in player 0's discard pile), and the hand is exactly D1 and D2. Afterwards the pool is still R0 T0 and the Refiner is Exhausted and on the battlefield.
- `TestCard_DolusRefiner_MayDiscardTheCardsJustDrawn` (`FR-MOD-001`): picking D1 and D2 discards them, and the hand is exactly H1 and H2.
- `TestCard_DolusRefiner_TimeoutDiscardsTheTwoCardsJustDrawn` (`FR-MOD-001` §6, `SVY-R16`, `SVY-R50`): a timed-out discard decision (`s.script(0, Answer{Timeout: true})`) discards the two cards most recently added to the hand, D1 and D2, and the hand is exactly H1 and H2.
- `TestCard_DolusRefiner_EmptyHandDiscardsBothDrawnWithoutAsking` (`FR-MOD-001` §1, `SVY-R16`): with an empty hand, it draws D1 and D2, and with exactly two cards in hand the discard is made without asking: no `DecideHand` decision is asked, D1 and D2 are in player 0's discard pile, and the hand is empty.
- `TestCard_DolusRefiner_ShortDeckDrawsWhatIsThereThenLoses` (`FR-WIN-001`, `FR-LAY-002`, `KIM-S04`): with an empty hand and one card (D1) in the deck, it draws D1, discards it without asking (the whole hand is fewer than 2 cards), and player 0 loses at the death sweep after the ability resolves: the match is over with player 1 the winner. As in `TestCard_ImprovisedOrders_ShortDeckDrawsWhatIsThereThenLoses`.
- `TestCard_DolusRefiner_OnlyTheControllersHandAndDeckAreUsed` (`FR-ACT-001`): player 1's hand and deck are unchanged, and player 1 is asked no `DecideHand` decision.
- `TestCard_DolusRefiner_DiscardingEerieRemnantFiresItsTrigger` (`FR-ZON-004`, `FR-TRN-003a`): with the real *Eerie Remnant* in player 0's hand beside H1, picking the Remnant and H1 discards both, and the Remnant's discarded-from-hand trigger is offered to player 0 (answered yes through `remnantChooser`, as in `TestCard_EerieRemnant_ImprovisedOrdersDiscardDeploysItExhausted`, `cards_path_remnant_test.go`), so the Remnant ends on the battlefield under player 0, Exhausted.
- `TestCard_DolusRefiner_NotOfferedWhileExhausted` (`FR-ACT-002`): an Exhausted Refiner isn't offered, and a rejected activation draws nothing and discards nothing.
- `TestCard_DolusRefiner_NotOfferedWhileTheStackIsNotEmpty` (`FR-ACT-001` clean-stack requirement): while an item is on the stack, the ability isn't offered to the player who holds priority, as in `TestFR_ACT_001_NeedsAnEmptyStack` (`pkg/engine/activated_test.go:117`; reuse its setup).
- `TestCard_DolusRefiner_MayActivateTheTurnItDeploys` (`FR-SET-004`): deployed from hand with R4, its ability is offered right away, and activating it draws 2 cards and discards 2.
- `TestCard_DolusRefiner_SourceGoneStillDrawsAndDiscards` (`FR-ZON-004`): the Refiner is killed between placement and resolution. Player 0 still draws D1 and D2 and discards the two cards picked, and the Refiner's card is in player 0's discard pile and is never offered in the discard decision.
- **Between placement and resolution:** `s.ptPlaceHook` on the ability's own placement fires while it is being declared, before its costs are paid (`SVY-R2`; T-0304's tests round 1, note 2; T-0314's tests round 1). Use T-0304's technique (`stormTapCard` in `cards_concordia_stormbolt_test.go`): a copy of the real card (copy the `Card` and its `Def`, so the shared definition isn't changed) whose `Triggers` also hold one synthetic `HookExhausted` trigger with `{Op: ability.OpKill, Subject: ability.SubjectSelf}`. It is placed on top of the ability and resolves first. Assert that it did (the Refiner is off the battlefield before the draw).
- **Existing tests:** every other test passes unchanged, including `TestCards_Executable` with the name removed.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected.

## Allowed Files
- `pkg/engine/cards/soulless.go` (the definition)
- `pkg/engine/cards/cards.go` (remove the name from `NotYetExecutable`)

## Done When
- [ ] The card tests pass: `go test ./pkg/engine -run TestCard_DolusRefiner`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-matrix` exits 0 (just over two minutes: use a timeout above the two-minute default)
- [ ] The card loads with no unsupported lines (`TestCards_Executable` passes with the name removed)

## Log
- 2026-11-26 lead-architect (wave A run, session 17): written against 607c9752, with the lead's probe above. `main` has no commits the run branch lacks. Board read (D2.0): no directive is due here (#185 has only (4)'s *Awaken Nightmare* half left, wave F), and no waiting trigger has fired: the card has no each-player effect (#71, both steps act on the controller), no cost on a mode (#91 (2)), no deck search (#106), no mandatory discard in a cost (#120: the discard is an effect step), no tuck (#142), no pile round trip (#183 (2)) and no draw or end-of-turn hook (#15 (2): it draws, but adds no hook), the mode is still `serial` (#6, #7), and no `needsRework` row exists (#184). Taken next in the survey's Soulless deck order, after *Dolus Macrocannon*. Searched `pkg/engine` and the task archives: every part of the Definition exists (`CostExhaustSelf`, *Improvised Orders*' `OpDraw` and `OpDiscard` steps through `stepDiscard`), so no kernel change. Set to `tests`.
- 2026-11-26 rules-test-author: wrote `pkg/engine/cards_soulless_refiner_test.go` (12 tests, all in the Tests section: EntersMainTheaterAsZeroNine, DrawsTwoThenDiscardsTheTwoPicked, MayDiscardTheCardsJustDrawn, TimeoutDiscardsTheTwoCardsJustDrawn, EmptyHandDiscardsBothDrawnWithoutAsking, ShortDeckDrawsWhatIsThereThenLoses, OnlyTheControllersHandAndDeckAreUsed, DiscardingEerieRemnantFiresItsTrigger, NotOfferedWhileExhausted, NotOfferedWhileTheStackIsNotEmpty, MayActivateTheTurnItDeploys, SourceGoneStillDrawsAndDiscards). Helpers requested: none (test-local helpers drfScene, drfHand, drfTapCard, drfHandIs). `go vet ./pkg/engine` clean, `cardtown lint rules` PASS. Ten fail until the definition exists; NotOfferedWhileExhausted and NotOfferedWhileTheStackIsNotEmpty already pass (the ability is absent). Source-gone uses the kill trigger from `drfTapCard` and checks via the log that the kill precedes the first draw, rather than a `ptPlaceHook` (the exhaust trigger is only placed after costs are paid). Nothing left out.
- 2026-11-26 lead-architect: tests need changes (D4, round 1). e02ac47b adds only `pkg/engine/cards_soulless_refiner_test.go` (12 tests, one per case) and the task file. With the lead's probe Definition 11 pass, and each of 8 mutants (draw 1, draw 3, discard 1, discard 3, discard before draw, steps on the opponents, an added R1 cost, no exhaust cost) fails at least 3. Expected outcomes agree with the FRD 01 text they cite. Note:
  1. `TestCard_DolusRefiner_SourceGoneStillDrawsAndDiscards` fails against the correct Definition: "kill seq 6, first draw seq -1". It looks for an `EventDraw` in the log, but a draw that takes a card logs no event: `EventDraw` is logged only for a draw from an empty deck (`pkg/engine/zones.go:26-33`), so the check rests on an implementation detail. Drop the log scan. Show the kill came first from the game state instead: in the `DecideHand` answer (a `s.byKind` responder that also picks the first two options), record whether the Refiner is still on the battlefield (`s.alive(r)`) and whether its card is in player 0's discard pile, and assert it was gone and in the pile when the discard was asked. The discard comes after the draw in the same resolution, so this pins the kill before the ability's steps. Keep the other assertions (D1 and D2 drawn, the two picked discarded, the Refiner's card never offered).
Set to `tests`.
- 2026-11-26 rules-test-author: round 1 note 1 addressed. `TestCard_DolusRefiner_SourceGoneStillDrawsAndDiscards` no longer scans the log for `EventDraw`. A `s.byKind` responder picks the first two options and, when the `DecideHand` discard is asked, records `!s.alive(r)` and `inDiscardPile(p, r.Inst)`; the test asserts both were true then (the kill came before the ability's steps). The other assertions are kept (Refiner's card never offered, D1 and D2 drawn, the two picked discarded). No other test changed; no helpers requested. `go vet` clean, `cardtown lint rules` PASS; the card tests still fail only because the definition doesn't exist yet. Set to `tests-review`.
- 2026-11-26 lead-architect: tests approved (D4, round 2). 18c07d7b changes only `TestCard_DolusRefiner_SourceGoneStillDrawsAndDiscards`, the test the round 1 note named: the log scan is gone, and a `s.byKind` responder records, when the `DecideHand` discard is asked, that the Refiner is off the battlefield and its card in player 0's discard pile. With the lead's probe Definition all 12 tests pass, and the 8 mutants still fail (3 to 9 tests each). Set to `ready`.
