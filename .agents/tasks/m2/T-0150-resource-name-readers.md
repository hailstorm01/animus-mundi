---
id: T-0150
title: Event readers compare with Resource.String, not typed-out names (directive #98, part 2 of 3, harness, flash, pure move)
kind: harness
tier: flash
wave: C.1 (board directive #98, structural review 7, finding 7)
status: queued
branch: m2/autoloop-wave-c.1/T-0150-resource-name-readers
depends_on: [T-0149]
claims: [file:pkg/eventlog/replay.go, file:pkg/sim/monitor_resources.go]
cards: []
requirements: []
---

## Goal
`actionOf` in `pkg/eventlog/replay.go` and the Generate monitor in `pkg/sim/monitor_resources.go` compare an event's resource against typed-out names. Both use `engine.Requisition.String()` and `engine.Tactics.String()` instead (#98). A pure move: same behavior, same test names, same sim output.

## Covers
- Split from T-0149 under the D2 size rule: #98 touches nine non-test files.
- `pkg/eventlog/replay.go:269` and `pkg/sim/monitor_resources.go` (`pool`, and `checkGenerate`'s `res`, `other` and zone checks). Monitor messages written as sentences stay. So do the event values `pkg/sim/monitor_test.go` passes in, which are test setup.
- Inputs, Acceptance Cases, Allowed Files and Done When are written against the code of the day, once T-0149 is `done`, before this task leaves `queued`.

## Log
- 2026-11-03, lead-architect (wave C.1 run, session 6): queued with T-0149 (D2.4).
