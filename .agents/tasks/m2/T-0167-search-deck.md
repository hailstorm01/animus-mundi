---
id: T-0167
title: A step can search its controller's deck for a card (OpSearchDeck, DecideSearch, a shuffle event; FR-DEK-001, wave D shape, sonnet)
kind: shape
tier: sonnet
wave: D
status: tests
branch: m2/autoloop-wave-d/T-0167-search-deck
depends_on: [T-0158, T-0163]
claims: [primitive:SEARCH_DECK, decision:DecideSearch, event:shuffle]
cards: []
requirements: [FR-DEK-001, FR-DEK-002, FR-MOD-001]
---

## Goal
`OpSearchDeck` asks its controller to choose a card from their deck, shuffles the rest and puts the card into their hand. *Cliff of Portents* (T-0169) needs it, and this task builds #105's parts (2) and (3) on synthetic cards. No card uses it yet.

## Inputs
- FRD 01 v3.21: `FR-DEK-001` (lines 161-165: must choose, sequence, mandatory shuffle, order and timeout), `FR-DEK-002` (line 168: immediate, off the stack), `FR-MOD-001` §1 (line 395: a deck inspection has no fast path) and §6 (line 402, *Deck Searches*), `FR-WIN-001` (line 171: only a draw from an empty deck loses), `FR-ACT-001`'s `SVY-R36` bullet (line 342: an ability with no target slot needs only its costs, and its example is a search of an empty deck).
- Survey rows: `docs/design/m2_starter_set_ability_survey.md` wave D. No card in this task.
- Rulings that apply: `SVY-R36`, `SVY-R37`, `SVY-R38`. Directive #105 parts (2) and (3).
- Design note: `docs/architecture/m2-ability-model.md`, Key types (`Step`, which lists `SearchDeck`) and Effect ops.
- Why it's its own task: the shape and the card together are six non-test files, over the D2.4 limit, so *Cliff of Portents* follows as T-0169 with its own tests of the two rulings.
- Kernel today:
  - The top of a deck is `Player.Deck[0]` (`moveTopToHand`). `g.shuffle`, `g.takeCard` and `g.putCard` are in `zones.go`, and `chooseDiscards` is the model for a decision built from a pile.
  - `Decision.Hidden` already turns off the fast path, and no builder sets it yet.
  - `choose` returns no picks and asks no one for an unsatisfiable decision, so a `Min: 1` decision over an empty deck asks nothing and logs no decision.
  - `pkg/view/log.go` prints a decision's picked card names to both seats. T-0168 hides a search's pick, before any card can search.

### Shape (lead decisions, binding)
- **`OpSearchDeck`** (`ability/ability.go`): the new last `Op`, so the existing values stay. No new `Step` field: the card always goes to its owner's hand, and `Step.Filter` isn't read. A search with criteria isn't built (see #106).
- **`DecideSearch`** (`choice.go`): the new last `DecisionKind`, so the existing values stay, named `"Search"` in `String()` (it is the logged `DecisionRecord.Kind`). Its comment cites `FR-DEK-001`, `SVY-R37` and `SVY-R38`. The fallback takes the first listed card: add it to `DecideMode`'s case in `fallbackPicks`, whose loop already picks the first `d.Min` listed. No new fallback code.
- **`stepSearchDeck`** (`effects.go`), reached from `runStep`'s switch by one case, like every op. It walks `stepPlayers`, as `deckDrop` does, and for each player, in this order (`FR-DEK-001`, *Sequence*):
  1. Builds a `Decision{Player, Kind: DecideSearch, Prompt: "Search your deck for a card", Min: 1, Max: 1, Hidden: true}` with one option per card in `Player.Deck`, in deck order, top first (`Label` the card's name, `Ref` its instance ID), and calls `g.choose`.
  2. If it picked a card, takes it out of the deck with `takeCard`.
  3. Calls `g.shuffle` on what remains, then logs the shuffle event.
  4. Puts the picked card into the hand with `putCard`.
  It returns how many cards moved to a hand. The empty deck isn't a special case: the decision is unsatisfiable, so the steps above already do what `SVY-R37` asks.
- **`EventShuffle = "shuffle"`** (`events.go`, after `EventDiscard`): `Event{Kind: EventShuffle, Player: <pid>, Zone: ZoneDeck.String()}`, Detail `player %d shuffles their deck`. It has no `Card` or `Inst`. Without it, the shuffle of an empty deck that `FR-DEK-001` requires can't be seen by anything. Only a search emits it. `g.shuffle` stays silent, because setup's shuffles come before `match-start`, which must stay the first event.
- **A search is not a draw:** no `g.draw`, no `DrewFromEmpty`, so an empty deck loses nothing (`FR-WIN-001`). No event names the card found. The decision record holds it, for the full log and replay.
- **Sims:** no card definition sets the op, so play doesn't change. `go tool task sim-compare -- autoloop/m2-wave-d` should pass byte for byte. If it differs, don't change code to make it match: report the first differing game in the Log and hand off.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
Synthetic cards only, in a new file `pkg/engine/search_deck_test.go`. Resolve the step through a synthetic activated ability (`activate` in `activated_test.go`) or a deploy trigger, whichever is shorter. `onTop` puts cards on a deck with `cards[0]` on top. The names `ability.OpSearchDeck`, `engine.DecideSearch` and `engine.EventShuffle` come from the Shape above, so the tests won't compile until the code exists. Every existing test passes unchanged.
- `TestFR_DEK_001_SearchPutsTheChosenCardIntoHand`: three cards in the deck, the player answering with the second option. That card is in hand and not in the deck, the other two are still in the deck, and the opponent's deck and hand are unchanged.
- `TestFR_DEK_001_SearchListsTheWholeDeckTopFirst` (`SVY-R38`): the one decision asked has kind `DecideSearch`, `Min` and `Max` 1, and options naming every card in the deck in deck order, top first, each `Ref` the card's instance ID.
- `TestFR_DEK_001_UnrestrictedSearchMustChoose` (`SVY-R37`): an answer with no picks is invalid, so the top card goes to hand, as the invalid-answer fallback requires (`FR-MOD-001` §6), and an `EventInvalidAnswer` is logged.
- `TestFR_MOD_001_SearchAsksEvenWithOneCard` (`FR-MOD-001` §1): with one card in the deck, the player is still asked.
- `TestFR_DEK_001_EmptyDeckStillShuffles` (`SVY-R37`): with an empty deck, nothing is asked, no decision event is logged, the hand is unchanged, and one `EventShuffle` names the player. The player doesn't lose at the next sweep (`FR-WIN-001`).
- `TestFR_MOD_001_SearchTimeoutTakesTopCard` (`SVY-R38`): a table of two cases, a timeout and no chooser at all. The player has three cards in the deck, an asset on the battlefield and a card in the discard pile. The card that was on top goes to hand, and the decision event's `How` is `timeout`.
- `TestFR_DEK_001_SearchShufflesTheRest` (`FR-DEK-002`): with 20 distinct cards in the deck, after the search the deck holds the other 19, in an order different from before (the scene's seed is fixed, so this can't flake), and exactly one `EventShuffle` follows the decision event, with no `Card` and no `Inst`.
- `TestFR_DEK_001_SearchCountsTheCardItFound` (`SVY-R14`): a later `IfDone` step in the same effect runs after a search that found a card, and doesn't run after a search of an empty deck.
- `TestFR_ACT_001_SearchAbilityActivatesWithAnEmptyDeck` (`SVY-R36`): an activated ability whose effect is only `OpSearchDeck`, with an empty deck. The action is offered, its cost is paid, nothing is asked, and the event appears.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected. The test file may declare its own.

## Allowed Files
- `pkg/engine/ability/ability.go` (`OpSearchDeck`)
- `pkg/engine/choice.go` (`DecideSearch`, `String`, `fallbackPicks`)
- `pkg/engine/effects.go` (`stepSearchDeck`, its case in `runStep`)
- `pkg/engine/events.go` (`EventShuffle`)

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run 'TestFR_DEK_001_|TestFR_MOD_001_Search|TestFR_ACT_001_SearchAbility'`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-compare -- autoloop/m2-wave-d` passes (no card uses the op; see Shape, Sims)
- [ ] `cards.NotYetExecutable` is unchanged

## Log
- 2026-11-04 lead-architect (wave D run, session 12): no `main` commits to sync. Not `WAVE-COMPLETE`: *Cliff of Portents* is on `NotYetExecutable`. No `open: before` directive blocks a task with `cards: []`; #105's parts (2), (3) are built here, and the row stays on the board until T-0169 tests them on the card. Waiting directives checked, none fired: #6, #7 (mode is `serial`); #8 (this task doesn't touch `deck.go`); #15 (no turn-boundary hook); #71, #91 (2) (no such card). Reuse search: `grep -i shuffle|Hidden|SearchDeck` finds `g.shuffle` and `Decision.Hidden` (both reused) and no search. A lead probe of this Shape plus T-0169's Definition, since reverted, passes `go tool task check` and the starter sim: 47/50 Animus main cards executable, P0 84.0% / P1 15.0% / draws 1.0% over 200 games, no monitor violation, no invalid answer. T-0168 and T-0169 are split off, `queued` (D2.4). A search with criteria is #106. Four non-test files. Status `tests`; dispatching `rules-test-author`.
