---
id: T-0169
title: Make Cliff of Portents executable (R1, T1, exhaust, search your deck for a card; wave D card, flash)
kind: card
tier: flash
wave: D
status: in-progress
branch: m2/autoloop-wave-d/T-0169-cliff-of-portents
depends_on: [T-0158, T-0163, T-0167, T-0168]
claims: [card:Cliff of Portents]
cards: [Cliff of Portents]
requirements: [FR-ACT-001, FR-ACT-002, FR-DEK-001, FR-DEK-002]
---

## Goal
*Cliff of Portents* (a deck card, 1 copy): *"R1, T1, exhaust: Search your deck for a card, put it into your hand, then shuffle your deck."* It only combines T-0158's activated ability and costs with T-0167's `OpSearchDeck`. The last card in the wave run's header, and the last of #105: its parts (2) and (3) are tested here on the real card.

Covers:
- The definition in `pkg/engine/cards/animus.go`, as under Definition below.
- The name off `cards.NotYetExecutable`.

## Inputs
- Card (`data/cards/units-and-structures-animus.csv` line 12): Animus Structure, subtype Terrain, deploys to the Main Theater.
- Survey row: `docs/design/m2_starter_set_ability_survey.md` line 144.
- FRD 01 v3.21: `FR-DEK-001` (lines 161-165), `FR-DEK-002` (line 167), `FR-ACT-001` (line 341), `SVY-R36` (line 342), `FR-ACT-002` (line 343), `FR-MOD-001` §1 (line 395) and §6's *Deck Searches* (line 402).
- Rulings: `SVY-R36`, `SVY-R37`, `SVY-R38`. Directive #105's parts (2) and (3); this task closes that row.

### Definition (the engineer adds it to `animusDefs`, before *Craterfall Ancient*)
- `Text`: the printed line, as above.
- `Activated`: one ability. `Costs`, in printed order: `{Kind: ability.CostResources, Req: 1, Tac: 1}`, `{Kind: ability.CostExhaustSelf}`. `Effect.Steps`: one, `{Op: ability.OpSearchDeck}`. No target slot.
- Remove the name from `cards.NotYetExecutable`.

A lead probe with this Definition on T-0167's Shape, since reverted, passes `go tool task check`, and the starter sim stays clean: 47/50 Animus main cards executable, P0 84.0% / P1 15.0% / draws 1.0% over 200 games, no monitor violation. Only one of those games activates it, so a clean sim says little about the card, and its tests have to carry it. `sim-compare` will differ here and isn't a gate.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
In a new file `pkg/engine/cards_animus_cliff_test.go`, from the CSV text, loading the real card (`s.card`), on `pkg/engine/cards_animus_sinuous_test.go`'s model. Player 0 is active and activates, from the Main Theater. `setPool` gives the resources, `search_deck_test.go`'s `namedDeck` the deck, and `activated_test.go` has `activate`, `offersActivate` and `actNotOffered` ("not offered" means offered false, rejected, nothing paid, nothing asked). A timeout is `Answer{Timeout: true}`.
- `TestCard_CliffOfPortents_PutsTheChosenCardIntoHand` (`FR-ACT-001`, `FR-DEK-001`): pool R3 T2, three cards in the deck, player 0 choosing the second. It is in hand, and the deck holds exactly the other two. The pool ends R2 T1, and the Structure stays in the Main Theater, Exhausted. Exactly one shuffle event names player 0. The opponent's deck and hand are unchanged.
- `TestCard_CliffOfPortents_ListsTheWholeDeckTopFirst` (`SVY-R38`): one search decision, min 1 and max 1, whose options are every card in the deck, in deck order, top first.
- `TestCard_CliffOfPortents_MustChooseACard` (`SVY-R37`, `FR-MOD-001` §6): with cards in the deck, an answer with no picks is invalid: the top card goes to hand and an invalid-answer event is logged.
- `TestCard_CliffOfPortents_AsksEvenWithOneCardInDeck` (`FR-MOD-001` §1): one card in the deck, one decision asked, and that card ends in hand.
- `TestCard_CliffOfPortents_TimeoutTakesTheTopCard` (`SVY-R38`): a timeout takes the card that was on top before the shuffle. Costs are paid as usual, and exactly one shuffle event is logged.
- `TestCard_CliffOfPortents_EmptyDeckStillActivatesAndShuffles` (`SVY-R36`, `SVY-R37`, `FR-DEK-001`): with an empty deck the ability is offered, because it has no target slot. Activating it pays both resources and exhausts the Structure, asks nothing, leaves the hand as it was, logs exactly one shuffle event, and doesn't lose the game: a search is not a draw.
- `TestCard_CliffOfPortents_NeedsOneRequisitionAndOneTactics` (`FR-ACT-001`): two cases, pool R0 T5 and pool R5 T0. Not offered in either, and the pool is unchanged.
- `TestCard_CliffOfPortents_NeedsToBeReadied` (`FR-ACT-002`): an Exhausted *Cliff of Portents* with a full pool is not offered. A Readied one, once activated, is not offered again that turn, although the pool could pay again. `actNotOffered` expects a Readied asset, so these two use `offersActivate` and `s.reject`.
- **Existing tests:** every other test passes unchanged, including `TestStarterSet_Executable` with the name removed, `TestView_Case8_ExecutableFlag` (seat 0's seed-1 hand still holds *Veil Render*), `TestProgress_Case4_Coverage` and `TestEvents_Case4_EveryActionKind`.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected.

## Allowed Files
- `pkg/engine/cards/animus.go` (the definition)
- `pkg/engine/cards/cards.go` (remove the name from `NotYetExecutable`)

## Done When
- [ ] The card tests pass: `go test ./pkg/engine -run TestCard_CliffOfPortents`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] The card loads with no unsupported lines (`TestStarterSet_Executable` passes with the name removed)

## Log
- 2026-11-04 lead-architect (wave D run, session 12): split off T-0167 (D2.4). Queued.
- 2026-11-05 lead-architect (wave D run, session 15): written against 0b83a30. The card's CSV text, `FR-DEK-001`, `FR-DEK-002`, `FR-ACT-001`, `SVY-R36`, `FR-ACT-002` and `FR-MOD-001` (FRD 01 v3.21), and the shapes in the Definition all read as written. Searched `pkg/engine` and the task archive: `OpSearchDeck`, `CostResources` and `CostExhaustSelf` exist, so nothing here is a new shape. Board re-read, no new directive. Set to `tests`.
- 2026-11-05 rules-test-author: wrote pkg/engine/cards_animus_cliff_test.go: TestCard_CliffOfPortents_{PutsTheChosenCardIntoHand,ListsTheWholeDeckTopFirst,MustChooseACard,AsksEvenWithOneCardInDeck,TimeoutTakesTheTopCard,EmptyDeckStillActivatesAndShuffles,NeedsOneRequisitionAndOneTactics,NeedsToBeReadied}. Helpers requested: none. Nothing left out. go vet clean; lint rules passes.
- 2026-11-05 lead-architect: tests returned (D4, round 1), one note. The other seven tests read right against `FR-ACT-001`, `FR-ACT-002`, `FR-DEK-001`, `SVY-R36`, `SVY-R37` and `SVY-R38`. A lead probe with the Definition above, since reverted, passes all eight and `go tool task check`; dropping `CostExhaustSelf`, the Requisition or the Tactics from it fails `NeedsToBeReadied` or `NeedsOneRequisitionAndOneTactics`. Note:
  1. `TestCard_CliffOfPortents_EmptyDeckStillActivatesAndShuffles` never calls `s.script(0)`, so `s.asked(0)` is empty whatever the engine does: `asks` is filled only by a scripted chooser (`scenario_test.go:175`), and "an empty deck asks nothing" (`SVY-R37`) can't fail. Add `s.script(0)` before the activation, as `TestFR_DEK_001_EmptyDeckStillShuffles` does. Also drop the line `s.g.Players[0].Deck = nil`, since `cliffScene(t, 0)` already leaves the deck empty.
- 2026-11-05 rules-test-author: note 1 fixed: EmptyDeckStillActivatesAndShuffles now calls s.script(0) before activating, so "asks nothing" can fail; dropped the redundant Deck = nil. Nothing else changed.
- 2026-11-05 lead-architect: tests approved (D4, round 2). `26c415d` is the note's change and nothing else, and the test file changed only in `test(T-0169)` commits. Before the definition, seven of the eight fail; `NeedsOneRequisitionAndOneTactics` passes until the card has its ability, and guards against over-offering after. Board re-read, no new directive. Set to `ready`.
