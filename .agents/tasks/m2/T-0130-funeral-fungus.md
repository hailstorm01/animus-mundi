---
id: T-0130
title: Make Funeral Fungus executable (exhaust a readied allied unit as an additional cost, then return chosen unit card from your discard pile to your hand, wave C card, flash)
kind: card
tier: flash
wave: C
status: queued
branch: m2/T-0130-funeral-fungus
depends_on: [T-0081, T-0128, T-0129]
claims: [card:Funeral Fungus]
cards: [Funeral Fungus]
requirements: [FR-TRN-003a, FR-MNV-001, FR-TGT-001]
---

## Goal
*Funeral Fungus* (a deck card, 0 Requisition, 1 Tactics): *"As an additional cost to enact this maneuver, exhaust a readied allied unit."* / *"Return chosen unit card from your discard pile to your hand."* Split from T-0128 under the D2 size rule; it only combines T-0129's exhaust cost with T-0081's discard-pile target and `OpReturnToHand`.

Covers:
- The definition in `pkg/engine/cards/animus.go`: the exhaust cost on a Readied allied unit, and one exactly-one target slot "unit card in your discard pile" with an `OpReturnToHand` step on it.
- The name off `cards.NotYetExecutable`.
- Benchmark blueprint #8 (`docs/frd/reference/02_Benchmark_Card_Blueprints.md` line 307) as a test source: no Readied allied unit, or no unit card in its controller's discard pile, means it can't be enacted (`SVY-R29`).

Inputs, Tests and Allowed Files are written against the code of the day before this task leaves `queued`. Expected Allowed Files: `pkg/engine/cards/animus.go`, `pkg/engine/cards/cards.go`.

## Log
- 2026-10-31 lead-architect: queued in wave C session 15, with T-0128's task commit (D2.4).
