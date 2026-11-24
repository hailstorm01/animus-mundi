---
id: T-0296
title: "LoadCards reports duplicate names, non-numeric or empty numeric cells, and Units or Structures with no zone (M3 Deliverable 1b; harness)"
kind: harness
tier: flash
wave: none (M3 design note Deliverable 1, part b)
status: queued
branch: m3/T-0296-loader-sheet-checks
depends_on: [T-0295]
claims: [func:engine.loadSheets, func:engine.readCards]
cards: []
requirements: [FR-DAT-002, FR-DAT-006]
---

## Goal
`LoadCards` refuses card sheets with the three data errors FRD 02 v2.0 gives, and returns one error that lists every problem it found, in file then row order. Today the loader silently keeps the later of two same-named rows ([cardload.go:146-148](../../../pkg/engine/cardload.go)) and reads a non-number as 0 ([cardload.go:177-180](../../../pkg/engine/cardload.go)).

## Covers (from M3 design note Decisions and Test Plan, Sheet checks)
- A card name on two rows, in one file or across two: `FR-DAT-002`. The problem names both rows.
- A non-numeric or empty cell in an asset sheet's numeric column: `FR-DAT-006`. An empty cell counts, because a card is never loaded with a guessed value. The lead settles which columns are numeric for which card types (Maneuvers have no power or health) against FRD 02 and today's sheets before this task leaves `queued`.
- A Unit or a Structure with no deployment zone: `FR-DAT-006`.
- Each problem names the file, row, card and column. Several problems in one load are reported together.
- Today's sheets have none of the three errors (checked 2026-11-23), so `go tool task check` and the sims don't change.

Inputs, Acceptance Cases and Allowed Files are written against the code of the day before this leaves `queued` (D2.4). `cardload.go` is expected to be the only non-test file, plus fixture sheets under a test folder.

## Done When
- [ ] The sheet-check tests pass
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim-compare` passes

## Log
- 2026-11-23 lead-architect: queued with T-0295 as M3 Deliverable 1 (b). It doesn't depend on T-0295's code, and follows it in serial mode.
