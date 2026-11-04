---
id: T-0151
title: One check for a Scratch Company card, IsScratchCompany (directive #98, part 3 of 3, harness, flash, pure move)
kind: harness
tier: flash
wave: C.1 (board directive #98, structural review 7, finding 7)
status: queued
branch: m2/autoloop-wave-c.1/T-0151-is-scratch-company
depends_on: [T-0150]
claims: [helper:IsScratchCompany]
cards: []
requirements: []
---

## Goal
`HasSubtype("Scratch Company")` is repeated in three files. `(*Card).IsScratchCompany()` in `pkg/engine/card.go`, below `HasSubtype`, becomes the one check, and those sites call it (#98). A pure move: same behavior, same test names, same sim output.

## Covers
- Split from T-0149 under the D2 size rule: #98 touches nine non-test files.
- `pkg/engine/card.go` (the new method), `pkg/engine/targeting.go:263` and `:265`, `pkg/engine/deck.go:71` and `pkg/sim/sim.go:38`. The `scratchFace` literal stays.
- This task touches `deck.go` as a pure move, so it doesn't fire #8's trigger.
- When its last part merges, #98 moves to the directive archive.
- Inputs, Acceptance Cases, Allowed Files and Done When are written against the code of the day, once T-0150 is `done`, before this task leaves `queued`.

## Log
- 2026-11-03, lead-architect (wave C.1 run, session 6): queued with T-0149 (D2.4).
