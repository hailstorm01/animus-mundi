---
id: T-0287
title: "One walker over an `Effect` and its `Modes`, and `checkStepReads` asks `InPile` (directive #173 (c), (d); wave H.2 harness, flash)"
kind: harness
tier: flash
wave: H.2 (board directive #173 (c), (d))
status: queued
branch: m2/autoloop-wave-h.2/T-0287-effect-walker
depends_on: [T-0286]
claims: [func:ability.Effect walker]
cards: []
requirements: []
---

## Goal
Split off from #173 by T-0286 (the four parts together touch six non-test files). This task covers #173 (c) and (d), as pure moves:
- **(c)** Walking an effect's steps into its `Modes`, recursively, is written in `Effect.ReadsZoneRef` ([ability.go:372-388](../../../pkg/engine/ability/ability.go)) and `checkEffectReads` ([cardload.go:99-109](../../../pkg/engine/cardload.go)). After this task, one walker in `ability` (lead's plan: `Effect.Parts()`, returning `e` and each of its `Modes`' parts, recursively, depth first, in the order both walks visit today) holds the recursion, and both read it. `for … range e.Modes` then appears only in the walker and in `legalModes` ([placement.go:285](../../../pkg/engine/placement.go)), which picks modes, not steps.
- **(d)** `checkStepReads` tests `s.Filter.Zone == ability.ZoneDiscard` ([cardload.go:113](../../../pkg/engine/cardload.go)), though `Filter.InPile`'s doc says every check that tells piles from the battlefield asks it. After this task it calls `s.Filter.InPile()` (the directive writes `!s.Filter.InPile()` for the requirement; inside that condition's list of failures the call is not negated, so behavior is unchanged), and `cardload.go` has no `Zone == ability.ZoneDiscard`.

Files: `pkg/engine/ability/ability.go` and `pkg/engine/cardload.go`, two non-test files. No test added or edited. When this merges after T-0286, #173 is `applied`.

## Inputs
To be written against the run branch before this task leaves `queued` (D2.4): exact edits, dry-run blobs, and acceptance cases (no `range e.Modes` outside the walker and `legalModes`, no `Zone == ability.ZoneDiscard` in `cardload.go`, the same tests pass unedited, `go tool task check`, `go tool task sim-compare -- autoloop/m2-wave-h.2` byte for byte).

## Allowed Files
- `pkg/engine/ability/ability.go`
- `pkg/engine/cardload.go`

## Tests
- None added or edited: a pure move. Every existing test passes unedited.

## Done When
- [ ] Acceptance cases hold, with their commands and results in the Log
- [ ] `go tool task sim-compare -- autoloop/m2-wave-h.2` passes (starter and stress sims byte-identical to the run branch)
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-19, lead-architect (wave H.2 run, session 6): split off from #173 by T-0286 under the D2 size rule, as #173 allows. Queued behind T-0286; #174 waits for both, because #173 blocks it ("Wave H.2, after #173").
