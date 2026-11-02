---
id: T-0129
title: An additional cost can exhaust a readied allied unit, and exhausting an Exhausted asset does nothing (FR-TRN-003a, SVY-R17, directive #58, wave C shape, sonnet)
kind: shape
tier: sonnet
wave: C
status: queued
branch: m2/T-0129-exhaust-cost
depends_on: [T-0128]
claims: [const:CostExhaust, func:exhaust]
cards: []
requirements: [FR-TRN-003a, FR-CAP-001]
---

## Goal
A Maneuver can print *"As an additional cost to enact this maneuver, exhaust a readied allied unit."* (*Funeral Fungus*, a deck card). Split from T-0128 under the D2 size rule: with the kill cost and all three cards, one task would touch five non-test files. *Funeral Fungus*'s definition is T-0130.

Covers:
- `ability.CostExhaust`, a second `CostKind` member after `CostKill`: only a Readied asset matching `Cost.Filter` can pay it. It reuses T-0128's `costDecision`, `costsPayable` and `payCosts`, and pays through the one `exhaust` (`actions.go:110`), so the exhaust hook fires and its trigger goes on top of the Maneuver (`SVY-R17`, "Costs: as `FR-TRN-003a`, on top of the card or ability paid for", FRD 01 line 302).
- Directive #58: `exhaust` gains `if a.Exhausted { return }` at its top, with a test that exhausting an already-Exhausted asset fires no exhaust trigger.
- An exhaust cost that can't be paid (no Readied allied unit) blocks the enact, as T-0128's kill cost does.
- Since T-0128's round 3, `payCosts` runs in `Apply`'s `ActEnact` case right after the Maneuver is pushed (`FR-TRN-003`), so an exhaust cost's trigger also sees the card on the stack.
- Carried from T-0128's review (its note 2): re-wrap the `placeable` doc comment in `placement.go` (one 110-column line) to the file's usual width.

Inputs, Tests and Allowed Files are written against the code of the day before this task leaves `queued`. Expected Allowed Files: `pkg/engine/ability/ability.go`, `pkg/engine/placement.go`, `pkg/engine/actions.go`.

## Log
- 2026-10-31 lead-architect: queued in wave C session 15, with T-0128's task commit (D2.4). #58 fires with this task, whose `claims:` include exhausting an asset by a cost.
- 2026-11-02 lead-architect: T-0128 merged (d67487b). Covers updated: `payCosts` now runs after the push, and the `placeable` comment re-wrap carried from T-0128's review.
