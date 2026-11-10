---
id: T-0209
title: A stat criterion in filters, "with power N or less" (wave F shape, sonnet)
kind: shape
tier: sonnet
wave: F
status: queued
branch: m2/autoloop-wave-f/T-0209-power-at-most
depends_on: []
claims: [shape:filter stat criterion (ability.Filter "stat N or less", power first; matchesCard reads a pile card's printed power, matchesFilterAt an asset's effective power)]
cards: []
requirements: [FR-TGT-001, FR-ZON-004]
---

## Goal
The "Dynamic Stats" criterion of `FR-TGT-001` (line 375), first as "with power N or less", which *Reaver's Rush* (T-0210) prints: "Deploy chosen unit card with power 3 or less from your discard pile for free." One criterion read in both filter checks: a card in a discard pile by its printed power (it has no other, `FR-ZON-004`), and an asset on the battlefield by its effective power (`g.Power`, Layer 6). A later stat (*Devilhorn Scouting Party*'s "Requisition cost 3 or less") is another value of the same criterion, not a second field. Split from T-0208 under the D2 size rule.

Covers:
- An `ability.Filter` stat criterion: an enum of stats whose zero value is "no criterion" and whose first value is power, and the bound N ("N or less").
- `matchesCard` and `matchesFilterAt` in `filter.go` check it, each in its own way (printed for a pile card, effective for an asset).
- Synthetic-card tests only; no card prints it until T-0210.

- Not this task: *Reaver's Rush* (T-0210, `queued`); the Requisition-cost stat (the *Devilhorn Scouting Party* task).

Inputs, Shape, Tests and Allowed Files are written against the run branch before this task leaves `queued`.

## Log
- 2026-11-10 lead-architect (wave F run, session 8): split off T-0208 under the D2 size rule; `queued`. It can go before or after T-0208, and T-0210 needs both.
