---
id: T-0055
title: Make Animoceps Rager executable, the first Headquarters ability (wave B card, flash)
kind: card
tier: flash
wave: B
status: queued
branch: m2/T-0055-animoceps-rager
depends_on: [T-0054]
claims: [card:Animoceps Rager, file:pkg/engine/cards/headquarters.go]
cards: [Animoceps Rager]
requirements: []
---

## Goal
The Animus starter Headquarters *Animoceps Rager* runs in the kernel: "Whenever this Headquarters is dealt damage, Scratch 1. This ability triggers only once each turn." It is plain `ability` data in a new `pkg/engine/cards/headquarters.go` (design note Boundaries, board directive #45).

## Inputs
- Card (`data/cards/units-and-structures-HQ.csv`): Animus Headquarters, 3/20, yields 1R 2T.
- Survey row: `docs/design/m2_starter_set_ability_survey.md` line 106, and the §4 note that "is dealt damage" uses the Unbowed threshold.
- Shapes it uses: `HookDamaged` (T-0050), `OpScratch`, and `Trigger.MaxPerTurn` (T-0054).

### Definition (the engineer adds it to a new `headquartersDefs` map in `pkg/engine/cards/headquarters.go`, registered in `init` like `pathDefs`)
- One trigger: `Hook: ability.HookDamaged`, `MaxPerTurn: 1`, no targets.
- One step: `Op: ability.OpScratch`, `Amount: ability.Value{N: 1}`.
- `Text` holds the printed rules text. The CSV cell is one line, so `Text` has one entry with both sentences.

### Also in this task
- `cmd/cardtown/sim.go` prints "HQ abilities are ignored." That stops being true. Replace both lines with a line that reports, for each seat, whether its Headquarters is executable (`Supported()`), in the style of the "main cards executable now" lines.

To be completed when the task moves to `tests` (test guidance, reviewer pre-check).

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
To be written by rules-test-author.

## Allowed Files
- `pkg/engine/cards/headquarters.go` (new)
- `pkg/engine/cards/cards.go` (remove the name from `NotYetExecutable`)
- `cmd/cardtown/sim.go`
- `pkg/engine/doc.go` (Headquarters abilities move to the "Built" list; remove "(until T-0055)" and rewrap the paragraph to the surrounding width)

## Done When
- [ ] The card's tests pass: `go test ./pkg/engine -run 'TestCard_AnimocepsRager'`
- [ ] `TestDefs_MatchRulesText` and `TestStarterSet_Executable` pass unchanged
- [ ] `go tool task check` passes
- [ ] `go tool task sim` runs clean; report the before and after summary lines in the Log

## Log
- 2026-10-23 lead-architect: Wrote the task, `queued` behind T-0054.
