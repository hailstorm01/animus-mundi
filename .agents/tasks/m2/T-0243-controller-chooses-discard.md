---
id: T-0243
title: A discard chosen by the ability's controller from an opponent's revealed hand, with the one hand timeout rule (wave H shape, sonnet; Drifting Concept's shape; directive #156)
kind: shape
tier: sonnet
wave: H
status: ready
branch: m2/autoloop-wave-h/T-0243-controller-chooses-discard
depends_on: []
claims: [shape:controller-chosen discard (ability.Step.ControllerChooses on OpDiscard; chooseDiscards takes the chooser), shape:hand timeout ranks the hand that holds the options (handPicks, SVY-R50)]
cards: []
requirements: [FR-MOD-001]
---

## Goal
Wave H's next card, *Drifting Concept* ("When this unit deploys, chosen opponent reveals their hand. You choose a card from it. That player discards that card."), needs a discard the kernel can't express: today every `OpDiscard` has each step player discard cards **of their own choice** (`stepDiscard`, [effects.go:264-272](../../../pkg/engine/effects.go); `chooseDiscards` always asks the hand's owner, [zones.go:84](../../../pkg/engine/zones.go)). This task adds a step flag by which the **ability's controller** chooses the cards from each step player's hand, with directive #156's rules (`SVY-R50`): the choice is an inspection, so it is always asked, even with one card, and its timeout takes the card most recently added to **that** hand, through the one hand ranking. The card itself is T-0244 (`queued`): shape and card together touch seven non-test files, so under the D2 size rule they split, as T-0241 and T-0242 did.

Covers:
- `Step.ControllerChooses`, read by `stepDiscard` (Shape below).
- `chooseDiscards` asks a chooser who may differ from the hand's owner; `handPicks` ranks by the hand that holds the options.
- `DecideHand`'s comment (`choice.go:18`).
- Directive #156's tests (a)–(d), on a synthetic card (below).
- Not this task: the *Drifting Concept* definition, `NotYetExecutable`, and `doc.go:36` (T-0244, to keep this task at four files); a reveal event in the match log (FR-MOD-001 §3 asks only that the chooser can inspect the hand, which the decision's options give it; the seat views and the protocol are FRD 04's, outside M2); choosing among several opponents ("chosen opponent" is the one opponent in a two-player match, `SubjectOpponents`, as *Despoiler Worm* uses it).

## Inputs
- Board directive #156 (`SVY-R50`, `open: before T-0243`), from the [outcome](../../../docs/design/outcomes/2026-11-15-wave-h-drifting-concept-revealed-hand-timeout.md) (`Status: applied: FRD 01 v3.27`, commit 24e938d, merged into the run branch at b061faf).
- FRD 01 v3.27 `FR-MOD-001` (from line 421):
  - §1 (line 422): the single-option fast path doesn't apply to a choice that inspects hidden cards in a Hand; choosing from your own Hand is not an inspection (`SVY-R16`).
  - §3 (line 424): when an effect reveals an opponent's hand (*Drifting Concept*), the choosing player always gets an inspection window before the chosen card is discarded.
  - §6 *Choosing Cards from a Hand* (line 434, `SVY-R16`, `SVY-R50`): a timeout selects the card most recently added to that Hand, whoever is choosing; from an opponent's revealed hand, the eligible card most recently added to the opponent's Hand; if no card is eligible, none is selected.
- Survey: `docs/design/m2_starter_set_ability_survey.md` §4 row R50, and the *Drifting Concept* card row (line 143: "deploy → `REVEAL_HAND` opponent, `DISCARD_CARD` chosen"; "Always asked, even with one card (§1, §3); timeout picks the card most recently added to the opponent's hand").
- Design note: `docs/architecture/m2-ability-model.md`, Deliverable 6m ("From wave B": "`RevealHand` with a chosen discard (*Drifting Concept*)") and Open Question 23 (answered).
- Code: `stepDiscard` and `stepPlayers` ([effects.go:170-175, 264-272](../../../pkg/engine/effects.go)); `discardFromHand`, `chooseDiscards` and `discardCards` ([zones.go:67-122](../../../pkg/engine/zones.go)); `choose`'s fast path (`choice.go:125`, skipped when `Decision.Hidden`), `fallbackPicks` and `handPicks` ([choice.go:150-205](../../../pkg/engine/choice.go)); `Decision.Hidden` (`choice.go:43`: "the choice inspects hidden cards in a Hand or Deck (no fast path, and the play log keeps it from the other seat)"); `Step` ([ability.go:362-375](../../../pkg/engine/ability/ability.go)); `fireDiscarded` ([trigger.go:85](../../../pkg/engine/trigger.go)).
- Existing equivalents searched (D2.3): `chooseDiscards` is the one place every hand-discard decision is built, and `handPicks` the one hand ranking. This task extends both rather than adding a second discard or a second ranking. `DecideSearch` and `DecideLook` are other hidden choices, but of your own deck, with their own listed-order fallbacks (`SVY-R37`, `SVY-R41`), so they aren't a hand choice to reuse. No step today lets one player choose for another (`git grep -n "Chooses\|chooser" pkg/engine` finds only the `Chooser` type and its callers).

### Shape (the engineer builds it)
- **`ability.Step.ControllerChooses bool`**, read only by `OpDiscard` so far. False, the zero value, is every step today: each step player discards cards of their own choice. True: the **ability's controller** chooses the cards from each step player's hand (with `SubjectOpponents`, the opponent's hand, revealed to the controller for the choice, `FR-MOD-001` §3). Its comment says so and cites `FR-MOD-001` §3 and `SVY-R50`.
- **`chooseDiscards`** gains the deciding player beside the hand's owner, so it stays the one place a hand-discard decision is built. The `Decision` it builds has `Player` the chooser, `Kind: DecideHand`, the owner's hand as its options, and `Hidden` true exactly when the chooser isn't the owner: that is an inspection, so `choose` skips the fast path even with one card (`FR-MOD-001` §1, §3), while choosing from your own hand still fast-paths (`TestFR_MOD_001_DiscardFromAOneCardHandIsAutomatic`). The cards picked go from the owner's hand to the owner's discard pile through `discardCards`, which logs the discard under the owner and fires `fireDiscarded`, as today. An empty hand asks nothing and discards nothing, as today. `costs.go` stays as it is: its discard cost is always the payer's own choice, so keep a call form for that (for example a thin wrapper with the owner as chooser) rather than editing `costs.go`, which isn't an Allowed File.
- **`stepDiscard`**: with `st.ControllerChooses`, each step player's discard is chosen by `r.it.Controller`.
- **`handPicks`** ranks the options by their positions in **the hand that holds them** (the player whose `Hand` holds the options' cards; every `DecideHand` decision's options come from one hand), not by `d.Player`'s hand. Most recent first, as today; still no draw from `g.rng`. This is the one ranking function for every hand choice: no second copy for opponents' hands. Today, with the chooser not the owner, every option would rank "not in hand" and the timeout would take the **oldest** card, which `SVY-R50` contradicts.
- **`DecideHand`'s comment** (`choice.go:18`, "choose cards from your own hand: not an inspection (fast path applies)"): it becomes "choose cards from a hand: from your own, not an inspection (fast path applies); from an opponent's revealed hand, Hidden (no fast path, FR-MOD-001 §3); fallback picks the cards most recently added to that hand (SVY-R16, SVY-R50)".
- Behavior doesn't change for any printed card: no definition sets `ControllerChooses`, and for every existing decision owner and chooser are the same player. `go tool task sim-compare -- autoloop/m2-wave-h` must be byte-identical (a gate here).

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
In a new file `pkg/engine/revealed_hand_test.go`, with a synthetic card whose ability has one step `{Op: OpDiscard, Amount: {N: 1}, Subject: SubjectOpponents, ControllerChooses: true}` (a Maneuver, or a deploy trigger as in the other trigger tests), controlled by player 0. Give player 1 a hand whose cards were added in a known order (the last one appended is the most recent), and give player 0 a hand of its own too, so a ranking by the wrong hand shows.
- `TestFR_MOD_001_ControllerChoosesFromOpponentsHand`: player 1 holds 3 cards. Player 0 is asked one `DecideHand` decision, with `Hidden` true and player 1's 3 cards as its options; player 1 is asked nothing. Player 0 picks the second card: it leaves player 1's hand for player 1's discard pile, player 0's hand and discard pile are unchanged, and the discard event names player 1.
- `TestFR_MOD_001_RevealedHandTimeoutTakesMostRecent` (#156 (a), `SVY-R50`): player 0's chooser times out on player 1's 3-card hand: the card most recently added to player 1's hand is discarded, from player 1's hand to player 1's discard pile. (Ranking by player 0's hand would take the oldest.) Add the same check for an invalid answer, which takes the same fallback (`FR-MOD-001` §6).
- `TestFR_MOD_001_RevealedOneCardHandIsAsked` (#156 (b), `FR-MOD-001` §1, §3): player 1 holds 1 card: player 0's chooser is still asked (a decision with `Hidden` true and one option), and the card is discarded.
- `TestFR_MOD_001_RevealedEmptyHandAsksNothing` (#156 (c)): player 1's hand is empty: no one is asked, nothing is discarded, and the step does nothing (a following `IfDone` step, for example the controller draws a card, doesn't run).
- `TestFR_MOD_001_RevealedHandDiscardFiresOwnersTrigger` (#156 (d)): player 1's hand holds *Eerie Remnant* (real card: "When this card is put into your discard pile from your hand, you may deploy it from your discard pile for free.") and another card; player 0 picks *Eerie Remnant*: its trigger is its owner's, so **player 1** is asked its "you may" (`DecideOptional`); accepting deploys it for player 1.
- **Existing tests:** every other test passes unchanged, in particular `discard_test.go`'s own-hand cases (`TestFR_MOD_001_DiscardFromAOneCardHandIsAutomatic`, `TestFR_MOD_001_DiscardTimeoutDiscardsTheCardMostRecentlyAdded`).

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected. The test author keeps any small builder in the test file.

## Allowed Files
- `pkg/engine/ability/ability.go` (`Step.ControllerChooses`)
- `pkg/engine/zones.go` (`chooseDiscards` takes the chooser; `discardFromHand` passes it)
- `pkg/engine/effects.go` (`stepDiscard` reads `ControllerChooses`)
- `pkg/engine/choice.go` (`handPicks` ranks the hand that holds the options; `DecideHand`'s comment)

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run "TestFR_MOD_001_(ControllerChoosesFromOpponentsHand|Revealed)"`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-compare -- autoloop/m2-wave-h` passes (no printed card sets `ControllerChooses` yet)

## Log
- 2026-11-15 lead-architect (wave H run, session 4): synced `main` into the run branch (b061faf: directive #156 and FRD 01 v3.27, `SVY-R50`; `BOARD.md` header conflict resolved as `Last directive #156`, `Last task T-0242`, row 156 added). #156's trigger fired (the outcome reads `Status: applied: FRD 01 v3.27`), so it is `open: before T-0243`. Written for *Drifting Concept*'s shape (design note 6m, "From wave B"); the card is T-0244 (`queued`), split under the D2 size rule. Design note Open Question 23 marked answered. Set to `tests`.
- 2026-11-15 rules-test-author: wrote `pkg/engine/revealed_hand_test.go` (7 tests): `TestFR_MOD_001_ControllerChoosesFromOpponentsHand`, `TestFR_MOD_001_RevealedHandTimeoutTakesMostRecent`, `TestFR_MOD_001_RevealedHandInvalidAnswerTakesMostRecent`, `TestFR_MOD_001_RevealedOneCardHandIsAsked`, `TestFR_MOD_001_RevealedEmptyHandAsksNothing`, `TestFR_MOD_001_RevealedHandDiscardFiresOwnersTrigger` (the first two plus the invalid-answer case cover directive #156 (a); the others (b), (c), (d)). Helpers requested: none (small builders live in the test file). `go vet` fails only on the missing `ability.Step.ControllerChooses` field, the shape the engineer adds. `cardtown lint rules` passes. Cases left out: none.
- 2026-11-15 lead-architect (wave H run, session 4): tests reviewed (762efd9): six tests, the five listed plus the invalid-answer check split into its own test, match the Tests list, directive #156 (a)–(d) and `FR-MOD-001` §1, §3, §6 (`SVY-R16`, `SVY-R50`). The timeout and invalid-answer tests discriminate: ranking by the chooser's hand (today's `handPicks`) would take A, the oldest. Only the missing `Step.ControllerChooses` keeps them from compiling. Set to `ready`.
