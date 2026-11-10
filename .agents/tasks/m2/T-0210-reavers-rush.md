---
id: T-0210
title: Reaver's Rush, a free deploy of a chosen unit card with power 3 or less from your discard pile (wave F deck card, flash; #140 (4))
kind: card
tier: flash
wave: F
status: queued
branch: m2/autoloop-wave-f/T-0210-reavers-rush
depends_on: [T-0208, T-0209]
claims: [card:Reaver's Rush]
cards: [Reaver's Rush]
requirements: [FR-TRN-003a, FR-MNV-001, FR-TGT-001, FR-MOD-001]
---

## Goal
*Reaver's Rush* (Path Maneuver, T3, two copies in `data/decks/path-starter.txt`): "Deploy chosen unit card with power 3 or less from your discard pile for free." One exactly-one pile target (allied, unit, power 3 or less, T-0209's criterion) and one `OpDeployFree` step on it (T-0208). It combines shapes only.

Covers:
- Its definition in `pkg/engine/cards/path.go`.
- `FR-MNV-001`'s own example: it can't be enacted without a unit card of power 3 or less in its controller's discard pile.
- Directive #140 (4) (`SVY-R43`): with *Devilhorn Guerrillas* (power 3, printing both zones) as the target, a timed-out zone choice puts it in Reserves, so its Main Theater deploy trigger doesn't fire; answering the Main Theater fires it.

Inputs, Tests and Allowed Files are written against the run branch before this task leaves `queued`.

## Log
- 2026-11-10 lead-architect (wave F run, session 8): split off T-0208 under the D2 size rule; `queued` on T-0208 and T-0209.
