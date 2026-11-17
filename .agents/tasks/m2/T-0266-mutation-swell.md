---
id: T-0266
title: "Mutation Swell: a counted kill cost {2, 4} on allied Scratch Company units, one different chosen unit card in the discard pile per 2 killed, each deployed for free (wave H card, flash)"
kind: card
tier: flash
wave: H
status: queued
branch: m2/autoloop-wave-h/T-0266-mutation-swell
depends_on: [T-0264, T-0265]
claims: [card:Mutation Swell]
cards: [Mutation Swell]
requirements: [FR-TRN-003a, FR-MNV-001, FR-TGT-001, FR-TGT-003, FR-MOD-001, FR-REP-001, SVY-R2, SVY-R15, SVY-R21, SVY-R53, SVY-R54]
---

## Goal
*Mutation Swell* (`data/cards/maneuvers-animus.csv:12`): "As an additional cost to enact this maneuver, kill 2 or 4 allied Scratch Company units. For every 2 units killed this way, deploy a different chosen unit card from your discard pile for free." Directive #158 (3) and (4), with T-0264's `Cost.Counts` and T-0265's `Target.PerKilled` and `Target.Different`. No new shape.

Covers (to be written in full against the code of the day before it leaves `queued`):
- `pkg/engine/cards/animus.go`: the definition: `ExtraCosts` one `CostKill` with `Counts: []int{2, 4}` on allied Scratch Company units; one target slot `PerKilled: 2`, `Different: true`, filtered like *Reaver's Rush*'s without the power limit (`pkg/engine/cards/path.go:643-649`: allied, unit, `ZoneDiscard`); one `OpDeployFree` step on that slot, deploying each target still in the pile at resolution, in printed zones (`FR-TRN-003a` 1), a gone target skipped (`FR-TGT-003`).
- `pkg/engine/cards/cards.go`: drop *Mutation Swell* from `NotYetExecutable`.
- Tests (#158 (4)): not enactable with 1 allied Scratch Company unit, or with 2 and no unit card in the pile; with 3 units, or with 4 and only 1 unit card, only 2 is offered; the killed units' cards aren't offered as targets; *Shifting Impossibility* replacing every cost kill still deploys every target (`SVY-R15`, `SVY-R21`); one target removed before resolution and the other still deploys; a timeout kills 2 (the first Scratch Companies in fixed order) and takes the first different unit cards from the top of the pile.
- Allowed Files (expected): `pkg/engine/cards/animus.go`, `pkg/engine/cards/cards.go`.

## Log
- 2026-11-17 lead-architect (wave H run, session 26): queued with T-0264, split off under the D2 size rule. Write its Inputs, Tests and Done When against the code of the day once T-0265 is `done`. Directive #158 (3), (4); when it merges, #158 is fully applied and moves to the archive.
