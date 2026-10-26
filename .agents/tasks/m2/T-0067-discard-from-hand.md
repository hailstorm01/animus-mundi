---
id: T-0067
title: A discard step that picks from your own hand, with the most-recently-added timeout (SVY-R16, wave B shape, sonnet)
kind: shape
tier: sonnet
wave: B
status: ready
branch: m2/T-0067-discard-from-hand
depends_on: []
claims: [primitive:DISCARD_CARD, decision:hand]
cards: []
requirements: [FR-MOD-001]
---

## Goal
A triggered ability can make its controller discard cards from their own hand: `ability.OpDiscard`, which the controller answers through a new `DecideHand` decision. Under `SVY-R16`, choosing from your own hand is not an inspection. A one-card hand is picked automatically, a timeout picks the card most recently added to the hand, and an empty hand discards nothing. The first card, Chainbreaker Irregulars ("Unbowed: Draw a card, then discard a card."), is T-0068.

## Inputs
- FRD 01: `FR-MOD-001` §1 (single-option fast path, line 347: choosing from your own hand is not an inspection) and §6 (line 357: "Choosing Cards from Your Own Hand ... selects the card most recently added to the Hand. For *Draw a card, then discard a card* this is the card just drawn, including when the Hand was empty before the draw"). Ruling `SVY-R16` (FRD 01 v3.11, line 24; survey line 75).
- Survey row: Chainbreaker Irregulars (wave B), line 146.
- Design note: `docs/architecture/m2-ability-model.md`, Decisions (ADR-0002 paragraph: "Choosing from your own hand will fall back to the card most recently added") and Open Question 8(b).
- Board directive #14: Irregulars is a deck card (4 copies in the Path starter deck).
- Kernel today: steps run in `resolveAbility` (`effects.go`), each `step<Op>` taking `(it, st, chosen)`; player steps use `stepPlayers`. Cards move between piles only through `takeCard` and `putCard` (`zones.go`, T-0059). `putCard` appends to the hand, so the hand's last element is the card most recently added. Decision kinds and their fallbacks are in `choice.go`. `fixedOrderPicks` ranks only battlefield assets, and an `Option.Ref` holds an asset ID or a card instance ID, so the hand rule gets its own kind rather than another case in the fixed order.

### Shape (lead decisions)
- `ability.OpDiscard`, appended after `OpAddArmor`, with the comment `// its controller discards Amount cards of their choice from their own hand (FR-MOD-001 §1, §6; SVY-R16)`. Only `SubjectController` (the zero value) is defined. Making an opponent discard, or choosing from a revealed opposing hand (Drifting Concept), is later work.
- `DecideHand`, appended after `DecideOrder` (so existing kinds keep their values), with the comment `// choose cards from your own hand: not an inspection (fast path applies); fallback picks the cards most recently added (SVY-R16)`. `String()` returns `"Hand"`.
- `stepDiscard(it, st, chosen)` in `effects.go`, beside the other steps, and a `case ability.OpDiscard` in `resolveAbility`. For each step player with n = `g.value(it, st.Amount)` > 0 and a non-empty hand, it asks one `Decision`:
  - `Kind: DecideHand`, `Player`: the step player, `Prompt`: `"Discard a card"` (or `"Discard N cards"` for N > 1).
  - `Options`: one per hand card, in hand order (oldest first), `Label` the card name, `Ref` the instance ID.
  - `Min = Max = min(n, len(hand))`, and `Hidden: false`, so a hand of exactly that many cards takes the fast path.
  
  With an empty hand, or n = 0, nothing is asked and no decision event is emitted. Each picked card, in pick order, goes `takeCard` then `putCard(inst, ZoneDiscard, false)` into its owner's discard pile. It emits one event: `Kind: "discard"`, `Player`, `Card` (name), `Inst`, `Zone: "Hand"` (where it came from), and the detail `player %d discards %s`. The card is public once it is in the discard pile, so naming it leaks nothing (`view.LogLines` prints it as is).
- `fallbackPicks` for `DecideHand`: the `d.Min` options whose cards come latest in `g.Players[d.Player].Hand`, most recent first. Rank by the card's position in that hand (find each option's `Ref` among the hand's instance IDs), not by option order. It draws nothing from `g.rng`.
- `doc.go`: add `OpDiscard` to the steps list in "What runs now", and the hand decision to the `Chooser` line.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
To be written in `pkg/engine/discard_test.go`, with synthetic cards only (Irregulars is T-0068). The file won't compile until `ability.OpDiscard` and `DecideHand` exist. A convenient source is a unit whose deploy trigger has the steps under test. Required cases:
- **Asks the controller:** `TestFR_MOD_001_DiscardAsksTheControllerForACardFromTheirHand`. With three cards in hand, the controller (not the opponent) gets one decision. Its options are exactly those three cards, picking one is required, and the picked card ends on top of its owner's discard pile (last element) while the other two stay in hand.
- **Fast path (§1):** `TestFR_MOD_001_DiscardFromAOneCardHandIsAutomatic`. With one card in hand, the scripted chooser is not asked, and that card is discarded.
- **Empty hand:** `TestFR_MOD_001_DiscardFromAnEmptyHandDoesNothing`. Nothing is asked, the discard pile is unchanged, and the ability still resolves.
- **Timeout (§6):** `TestFR_MOD_001_DiscardTimeoutDiscardsTheCardMostRecentlyAdded`. Cards are given to the hand in a known order, and a timeout discards the last one given. Choose the setup so that the first option listed is not the right answer.
- **Invalid answer:** `TestFR_MOD_001_InvalidDiscardAnswerFallsBackToTheCardMostRecentlyAdded`. An out-of-range pick falls back the same way.
- **Draw, then discard (§6, line 357):** `TestFR_MOD_001_DrawThenDiscardTimeoutDiscardsTheCardJustDrawn`. With steps Draw 1 then Discard 1 and two cards already in hand, a timeout discards the card just drawn (put on top of the deck with `onTop`).
- **Draw into an empty hand:** `TestFR_MOD_001_DrawThenDiscardFromAnEmptyHandDiscardsTheDrawnCardAutomatically`. With an empty hand, the drawn card is the only option. It is discarded without asking.
- **Event:** `TestFR_MOD_001_DiscardEmitsADiscardEvent`. The log holds one `discard` event naming the discarding player, the card's name and its instance ID, after the step's decision event.
- **Left out unless FRD 01 fixes it:** discarding more cards than the hand holds, and an opponent discarding. No starter card needs either yet.

## Helpers Needed
None expected. `give`, `onTop`, `script` and `asked` already exist in `scenario_test.go`. If a helper is needed, request it here and the engineer adds it.

## Allowed Files
- `pkg/engine/ability/ability.go` (`OpDiscard`)
- `pkg/engine/choice.go` (`DecideHand`, its `String` name and its fallback)
- `pkg/engine/effects.go` (`stepDiscard` and its `resolveAbility` case)
- `pkg/engine/doc.go` (the steps and decisions lines)

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run 'TestFR_MOD_001_(Discard|InvalidDiscard|DrawThenDiscard)'`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] No card uses `OpDiscard` yet, so the sims are unchanged: `go tool task sim-compare -- autoloop/m2-wave-b` passes

## Log
- 2026-10-26 lead-architect (wave run B, session 1): wrote the task. It is the next wave B deck card's shape (directive #14 (a): Irregulars is in the Path starter deck), and it is split from its card task T-0068 because the two together would have five non-test files. Directives checked: none open applies. #59 (1) is about piles, and the ruling says your hand is not one. #8, #15, #51 and #58 are not triggered. Claims don't overlap any open task, since none is open. No existing equivalent: `pkg/engine` has no discard step and no decision about hand cards. Status `tests`; dispatching `rules-test-author`.
- 2026-10-26 rules-test-author: wrote `pkg/engine/discard_test.go` with the eight required tests: TestFR_MOD_001_DiscardAsksTheControllerForACardFromTheirHand, ..._DiscardFromAOneCardHandIsAutomatic, ..._DiscardFromAnEmptyHandDoesNothing, ..._DiscardTimeoutDiscardsTheCardMostRecentlyAdded, ..._InvalidDiscardAnswerFallsBackToTheCardMostRecentlyAdded, ..._DrawThenDiscardTimeoutDiscardsTheCardJustDrawn, ..._DrawThenDiscardFromAnEmptyHandDiscardsTheDrawnCardAutomatically, ..._DiscardEmitsADiscardEvent. Helpers requested: none (the file defines its own `discardScene`, `runDiscardSource`, `discardN`). Left out: discarding more than the hand holds and an opponent discarding, as the task said. Tests do not compile until `ability.OpDiscard` and `DecideHand` exist.
- 2026-10-26 lead-architect (tests-review, round 1): **Approved, `ready`.** Checked against FRD 01 `FR-MOD-001` §1 (line 347) and §6 (line 357) and `SVY-R16`. The expected outcomes are right: a one-card hand takes the fast path, an empty hand asks nothing, and both the timeout and an invalid answer pick the card most recently added. In the draw-then-discard cases, the drawn card is that card, including from an empty hand. In the timeout, invalid-answer and draw cases, the first option listed is not the answer, so option order can't pass them by accident. Every helper used exists (`trigUnit`, `deployTrigger`, `drawN`, `vanilla`, `deployFromHand`, `eventsOfKind`, `script`). The source card leaves the hand when it is deployed, before its trigger asks. The event assertions read the structured match log (ADR 0005), as T-0066's did. The file fails only because `ability.OpDiscard` and `DecideHand` don't exist yet, which is expected. No earlier test is modified. Next: dispatch `kernel-engineer` (tier `sonnet`). The Shape section's lead decisions are binding.
