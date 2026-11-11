---
id: T-0215
title: Replay, the play screen and ProgressPolicy handle a deploy from a discard pile (wave F harness)
kind: harness
tier: flash
wave: F
status: queued
branch: m2/autoloop-wave-f/T-0215-discard-deploy-clients
depends_on: [T-0214]
claims: [file:pkg/eventlog/replay.go (actionOf reads Event.From), file:cmd/cardtown/play_screen.go (a discard deploy's label), file:pkg/sim/policy.go (handCosts counts discard deploys)]
cards: []
requirements: [FR-TRN-003]
---

## Goal
The clients follow T-0214's discard deploy, so *Veil Render* (T-0216) can be played, simulated and replayed. Directive #128's last part.

Covers (Acceptance Cases written against the code of the day before it leaves `queued`):
- `actionOf` ([replay.go](../../../pkg/eventlog/replay.go)) rebuilds `Action.From` from the action event's `From`, so a recorded discard deploy replays with no divergence (#128's Acceptance); an unknown `From` is an error, as an unknown zone is.
- The play screen's action label ([play_screen.go](../../../cmd/cardtown/play_screen.go)) names a discard deploy's card from the seat's discard pile and says it comes from the discard pile, instead of `Card#N`.
- `ProgressPolicy` counts a discard deploy's cost with the hand cards' when it budgets Generate and Convert (`handCosts`, #122 (a): "and discard-pile deploy costs when Wave F adds them"; #128: "`ProgressPolicy` counts the deploy's cost").
- `FormatEvent` and the view's log lines, if they print an action event's card, say a discard deploy's origin (checked against `pkg/eventlog` and `pkg/view` when written).
- `go tool task sim-compare -- autoloop/m2-wave-f` byte-identical: no printed card has the permission until T-0216.

## Allowed Files
- `pkg/eventlog/replay.go`
- `cmd/cardtown/play_screen.go`
- `pkg/sim/policy.go`
- (one more if `FormatEvent` or the view needs it, named when written)

## Log
- 2026-11-10 lead-architect (wave F run, session 13): queued with T-0213, split off *Veil Render* under the D2 size rule (D2.4). Depends on T-0214's `Action.From` and `Event.From`.
