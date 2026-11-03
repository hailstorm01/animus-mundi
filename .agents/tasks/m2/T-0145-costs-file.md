---
id: T-0145
title: Additional costs and a card's cost get their own file, costs.go (directive #94, harness, flash, pure move)
kind: harness
tier: flash
wave: C.1 (board directive #94, structural review 7, finding 3)
status: queued
branch: m2/autoloop-wave-c.1/T-0145-costs-file
depends_on: [T-0144]
claims: [file:pkg/engine/costs.go]
cards: []
requirements: []
---

## Goal
A new `pkg/engine/costs.go` holds `costDecision`, `chooseCostDiscards`, `costsPayable`, `payCosts`, `payCost` and `EffectiveCost`, moved unchanged from `placement.go` and `actions.go`, and `doc.go`'s Files list names it (#94). A pure move: same declarations, same test names, same sim output.

## Covers
- Split from T-0144 under the D2 size rule: #92, #93 and #94 together touch five non-test files. This task touches `actions.go`, `placement.go`, `costs.go` and `doc.go`.
- Inputs, Acceptance Cases, Allowed Files and Done When are written against the code of the day, once T-0144 is `done`, before this task leaves `queued`.

## Log
- 2026-11-03, lead-architect (wave C.1 run, session 1): queued with T-0144 (D2.4).
