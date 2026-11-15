---
id: T-0244
title: "Drifting Concept: when it deploys, you choose a card from the opponent's revealed hand and they discard it (wave H card, flash; directive #156)"
kind: card
tier: flash
wave: H
status: queued
branch: m2/autoloop-wave-h/T-0244-drifting-concept
depends_on: [T-0243]
claims: [card:Drifting Concept]
cards: [Drifting Concept]
requirements: [FR-MOD-001, FR-CBT-001]
---

## Goal
*Drifting Concept* (Animus unit, `data/cards/units-and-structures-animus.csv:19`; in neither starter deck): "Ethereal / When this unit deploys, chosen opponent reveals their hand. You choose a card from it. That player discards that card." becomes executable from T-0243's `Step.ControllerChooses` and shapes already `done` (`HookDeploy`, `OpDiscard`, `SubjectOpponents`; Ethereal is parsed from the CSV). It leaves `cards.NotYetExecutable`. With it, directive #156 is fully applied.

Covers (written now so the plan is on the board; the lead rewrites Inputs, Tests and Allowed Files against T-0243's merged code before this task leaves `queued`, D2.4):
- `pkg/engine/cards/animus.go`: the definition, in alphabetical order (after *Disorienting Pulse*). `Text` is the ability line of the CSV text: `"When this unit deploys, chosen opponent reveals their hand. You choose a card from it. That player discards that card."`. One trigger: `Hook: ability.HookDeploy`, an effect with no targets and one step `{Op: ability.OpDiscard, Amount: ability.Value{N: 1}, Subject: ability.SubjectOpponents, ControllerChooses: true}`.
- `pkg/engine/cards/cards.go`: remove "Drifting Concept" from `NotYetExecutable`.
- `pkg/engine/doc.go:36`, comment only (T-0243 left it to this task): "choosing from your own hand (DecideHand, SVY-R16)" becomes "choosing from a hand, your own or an opponent's revealed one (DecideHand, SVY-R16, SVY-R50)".
- Card tests in `pkg/engine/cards_animus_drifting_test.go` (real card): deployed, its controller is asked to choose from the opponent's hand (hidden, even with one card) and the chosen card goes to the opponent's discard pile; a timeout discards the card most recently added to the opponent's hand (`SVY-R50`); an empty opposing hand: nothing is asked or discarded; it has Ethereal from its CSV keyword line (`FR-CBT-001`, as `CanAttack` reads it).

## Done When
- [ ] The card tests pass
- [ ] `go tool task check` passes
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-compare -- autoloop/m2-wave-h` passes (*Drifting Concept* is in neither starter deck)

## Log
- 2026-11-15 lead-architect (wave H run, session 4): queued, split off T-0243 under the D2 size rule (shape and card together touch seven non-test files). Waits for T-0243.
