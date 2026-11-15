---
id: T-0244
title: "Drifting Concept: when it deploys, you choose a card from the opponent's revealed hand and they discard it (wave H card, flash; directive #156)"
kind: card
tier: flash
wave: H
status: tests-review
branch: m2/autoloop-wave-h/T-0244-drifting-concept
depends_on: [T-0243]
claims: [card:Drifting Concept]
cards: [Drifting Concept]
requirements: [FR-MOD-001, FR-KWD-002g]
---

## Goal
*Drifting Concept* (Animus unit, `data/cards/units-and-structures-animus.csv:19`; in neither starter deck): "Ethereal / When this unit deploys, chosen opponent reveals their hand. You choose a card from it. That player discards that card." becomes executable from shapes that are `done`: T-0243's `Step.ControllerChooses` (with `chooseHandCards` and `handPicks` ranking the hand that holds the options), `HookDeploy`, `OpDiscard` and `SubjectOpponents`. Ethereal is parsed from the CSV keyword line. The card leaves `cards.NotYetExecutable`. With it, directive #156 is fully applied.

## Inputs
- Card text: `data/cards/units-and-structures-animus.csv:19` (Animus unit, Spirit, R4 T0, Reserves, 3/4, yields 1/1, Ethereal).
- Board directive #156 (`SVY-R50`, `open: before T-0244`), from the [outcome](../../../docs/design/outcomes/2026-11-15-wave-h-drifting-concept-revealed-hand-timeout.md) (`Status: applied: FRD 01 v3.27`). T-0243 (merge a0cd951) applied its kernel part, with tests (a)–(d) on a synthetic card in `pkg/engine/revealed_hand_test.go`. This task adds the real card, its card tests and the `doc.go:36` comment.
- FRD 01 v3.27, `FR-MOD-001` (from line 420):
  - §1 (line 421), §3 (line 424): choosing from an opponent's revealed hand is an inspection, so it is always asked, even with one card.
  - §6 *Choosing Cards from a Hand* (line 434, `SVY-R16`, `SVY-R50`): a timeout selects the card most recently added to that hand, whoever is choosing.
- `FR-KWD-002g` (line 615): Ethereal, "This asset cannot be attacked."
- "Chosen opponent" in a two-player game is the one opponent: `SubjectOpponents`, with no choice asked.
- Code of the day (run branch at 6512ae3): `Step.ControllerChooses` (`pkg/engine/ability/ability.go:374`), `stepDiscard` (`pkg/engine/effects.go:267`), `chooseHandCards` (`pkg/engine/zones.go:94`), `handPicks` (`pkg/engine/choice.go:193`).

## Shape (the lead's plan)
- `pkg/engine/cards/animus.go`: the definition, in alphabetical order (after *Disorienting Pulse*, before *Ensnaring Mist*). `Text` is the ability line of the CSV text: `"When this unit deploys, chosen opponent reveals their hand. You choose a card from it. That player discards that card."`. One trigger, `Hook: ability.HookDeploy`, with an effect with no targets and one step: `{Op: ability.OpDiscard, Amount: ability.Value{N: 1}, Subject: ability.SubjectOpponents, ControllerChooses: true}`. No new shape, no new field.
- `pkg/engine/cards/cards.go`: remove "Drifting Concept" from `NotYetExecutable`.
- `pkg/engine/doc.go:36`, comment only (T-0243 left it to this task). In that line, `choosing from your own hand (DecideHand, SVY-R16)` becomes `choosing from a hand, your own or an opponent's revealed one (DecideHand, SVY-R16, SVY-R50),`. The rest of the line and everything else in `doc.go` stay as they are.
- The lead dry-ran the definition and the `NotYetExecutable` removal on the run branch at 6512ae3 (2026-11-15, wave H session 5): `gofmt -l pkg cmd` clean, `go tool task check` exit 0 with no existing test edited, `go tool task sim` exit 0, and `go tool task sim-compare -- autoloop/m2-wave-h` `[PASS]` (byte-identical: the card is in neither starter deck).

## Tests (card tasks: written by rules-test-author; engineers never edit these)
New tests go in `pkg/engine/cards_animus_drifting_test.go`. Use the real card (`s.card("Drifting Concept")`), deployed by player 0 (the active player) from hand with `s.deployFromHand` (give player 0 the pool it needs). Give player 1 a hand of vanilla cards, added in a known order (`s.give`). Answer the hand decision through `s.script`, and check the asked decisions with `s.asked`. "Discarded" means the card has left player 1's hand and is on top of player 1's discard pile, and player 0's hand and discard pile are unchanged apart from the deployed card.
- `TestCard_DriftingConcept_ControllerChoosesFromOpponentsHand` (`FR-MOD-001` §3, #156): player 1 holds three cards. Player 0 is asked exactly one `DecideHand` decision, `Hidden`, whose options are player 1's three cards; player 1 is asked nothing. Player 0 picks the middle card, and that card is discarded.
- `TestCard_DriftingConcept_TimeoutDiscardsMostRecentlyAdded` (`FR-MOD-001` §6, `SVY-R50`, #156 (a)): player 1 holds three cards. Player 0 times out (`Answer{Timeout: true}`). The card most recently added to player 1's hand is discarded.
- `TestCard_DriftingConcept_OneCardIsStillAsked` (`FR-MOD-001` §1, §3, #156 (b)): player 1 holds one card. Player 0 is still asked one `DecideHand` decision with one option, and that card is discarded.
- `TestCard_DriftingConcept_EmptyHandAsksNothing` (#156 (c)): player 1's hand is empty. Neither player is asked a `DecideHand` decision, and no discard pile changes.
- `TestCard_DriftingConcept_IsEthereal` (`FR-KWD-002g`): with *Drifting Concept* under player 1 in the Main Theater and a readied attacker of player 0's in the Main Theater, an attack declared on it is illegal (as `TestFR_KWD_002g_EtherealCannotBeAttackedAndOverridesGuardian` at `rules_test.go:336` checks a synthetic Ethereal unit). This pins the CSV keyword line, not the kernel rule, so it passes before the implementation.
- `TestDefs_MatchRulesText` already checks `Text` against the CSV. #156 (d) (*Eerie Remnant*'s put-into-discard-from-hand trigger) is pinned by T-0243's `revealed_hand_test.go` with the same step, so nothing to add here.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- none expected. If the test author needs one, it names it here.

## Allowed Files
- `pkg/engine/cards/animus.go`
- `pkg/engine/cards/cards.go`
- `pkg/engine/doc.go`

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run "DriftingConcept|TestDefs"`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-compare -- autoloop/m2-wave-h` passes (*Drifting Concept* is in neither starter deck)

## Log
- 2026-11-15 lead-architect (wave H run, session 4): queued, split off T-0243 under the D2 size rule (shape and card together touch seven non-test files). Waits for T-0243.
- 2026-11-15 lead-architect (wave H run, session 5): T-0243 is `done` (merge a0cd951). Wrote Inputs, Shape, Tests and Allowed Files against the run branch at 6512ae3, and dry-ran the definition (gates above). Re-read `BOARD.md` before committing: no directive added since; #156 is the one this task applies. Status `tests`.
- 2026-11-15 rules-test-author: wrote `pkg/engine/cards_animus_drifting_test.go`: TestCard_DriftingConcept_ControllerChoosesFromOpponentsHand, _TimeoutDiscardsMostRecentlyAdded, _OneCardIsStillAsked, _EmptyHandAsksNothing, _IsEthereal. Helpers requested: none (file-local `dcScene`, `dcDeploy`, `dcWantDiscarded`). Nothing left out. Four tests fail now with "needs rules the kernel does not implement yet"; IsEthereal passes (pins the CSV keyword). `go vet` clean, `cardtown lint rules` PASS. Status `tests-review`.
