---
id: T-0068
title: Make Chainbreaker Irregulars executable (wave B card, flash)
kind: card
tier: flash
wave: B
status: queued
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
To be written once T-0067 is `done`.

## Helpers Needed
None expected.

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
