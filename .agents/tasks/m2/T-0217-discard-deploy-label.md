---
id: T-0217
title: The play screen names a deploy from the discard pile (wave F harness)
kind: harness
tier: flash
wave: F
status: queued
branch: m2/autoloop-wave-f/T-0217-discard-deploy-label
depends_on: [T-0214]
claims: [file:pkg/view/view.go (Inst on the seat's own discard cards), file:cmd/cardtown/play_screen.go (a discard deploy's label)]
cards: []
requirements: [FR-TRN-003]
---

## Goal
`cardtown play` labels a deploy from the discard pile with its card's name and its origin, so a human can play *Veil Render* (T-0216) from the discard pile. The rest of directive #128's last part, split off T-0215 under the D2 size rule.

Covers (Inputs, Acceptance Cases and Edits written against the code of the day before it leaves `queued`):
- `view.side` ([view.go](../../../pkg/view/view.go)) sets `CardView.Inst` on the seat's own discard pile cards, as it does on its hand cards, because an Action now names them (`Action.From`, T-0214). `CardView.Inst`'s comment says so. `ReqNow` and `TacNow` stay 0 there. The opponent's discard cards keep `Inst` 0.
- `actionLabel` ([play_screen.go](../../../cmd/cardtown/play_screen.go)) labels an `ActDeploy` with `From: ZoneDiscard` from `v.You.Discard`, saying it comes from the discard pile (for example `Deploy: Veil Render from discard pile to Reserves`), with the `Card#N` fallback when the card isn't there. A hand deploy's label doesn't change.
- Check first that no golden output or test pins a discard `CardView`'s `Inst` at 0.
- `go tool task sim-compare -- autoloop/m2-wave-f` byte-identical.

## Log
- 2026-11-11 lead-architect (wave F run, session 16): queued with T-0215, whose original Covers needed five non-test files (D2.4). T-0216 depends on it. Directive #128 is `open: before T-0216`, and this task applies its last piece.
