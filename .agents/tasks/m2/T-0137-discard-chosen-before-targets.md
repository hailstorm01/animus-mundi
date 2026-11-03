---
id: T-0137
title: A discard cost's cards are chosen before the targets and discarded after the push (FR-TRN-003a, SVY-R2, SVY-R4, directive #20 (b), wave C shape, sonnet)
kind: shape
tier: sonnet
wave: C
status: queued
branch: m2/autoloop-wave-c/T-0137-discard-chosen-before-targets
depends_on: [T-0136]
claims: [field:placement.discards, func:chooseDiscards, func:discardCards, func:discardFromHand, func:payCosts]
cards: []
requirements: [FR-TRN-003a, FR-TRN-003, FR-TGT-001, FR-MNV-001]
---

## Goal
When a Maneuver with a discard cost is enacted, its controller picks the cards to discard first, then its modes and targets are chosen, then it is pushed and the picked cards are discarded. So the number of cards discarded is known when the targets are chosen (*Steel Rain*: one target per card discarded, `SVY-R4`, FRD 01 line 362), while targets are still judged against the state before the cost moved anything (`SVY-R2`). Split from #20 (b) under the D2 size rule; T-0138 (queued) reads the recorded cards.

## Covers (to be written out against the code of the day before this task leaves `queued`, D2.4)
- `discardFromHand` (`zones.go`) splits into choosing and moving, and stays the one place that builds the hand-discard decision: a `chooseDiscards(pid, n, upTo) []*Instance` that only asks, and a `discardCards(pid, cards)` that moves and logs (`EventDiscard`, zone `"Hand"`). `discardFromHand` is the two in a row, so `stepDiscard` is unchanged.
- `placement` gains `discards []*Instance`, the cards its discard costs picked. In `Apply`'s `ActEnact` (`actions.go`), after `takeCard` (so the Maneuver is never an option, `FR-TRN-003`) and before `chooseOnPlacement`, the discard costs are chosen into `pl.discards`; `payCosts` then discards exactly those cards after the push, in printed cost order. Kill and exhaust costs keep their picks after the targets, as today (`SVY-R2`).
- Tests to name then: the discard decision is asked before the target decision; targets are still judged before the cards move (a "chosen card in your discard pile" slot never offers a card picked for the cost, `SVY-R2`); T-0131's `TestFR_TRN_003a_DiscardCost*` tests keep passing unedited (discard events after the Maneuver's `place`, before its damage).
- `go tool task sim-compare -- autoloop/m2-wave-c` passes: no card uses `CostDiscard` yet.

## Allowed Files (planned)
- `pkg/engine/zones.go` (`discardFromHand` and the two functions split from it)
- `pkg/engine/placement.go` (`placement`, `payCosts`)
- `pkg/engine/actions.go` (`Apply`'s `ActEnact` case only)

## Log
- 2026-11-02 lead-architect (wave C session 27): queued with T-0136 when #20 (b) was split (D2.4). Write its Inputs, Tests and Done When against the code of the day before it leaves `queued`.
