---
id: T-0265
title: "Target slots that repeat per kill and pick different entities: Target.PerKilled and Target.Different, and the counts offered narrowed to those whose targets can be filled (wave H shape, sonnet)"
kind: shape
tier: sonnet
wave: H
status: queued
branch: m2/autoloop-wave-h/T-0265-per-kill-different-targets
depends_on: [T-0264]
claims: [shape:per-kill target slots (Target.PerKilled: a slot chosen once for every k assets the placement's counted kill cost kills), shape:different repeated targets (Target.Different: a repeated slot's picks are different entities, SVY-R54), shape:legal cost counts (the counts offered and the enact check narrowed to those whose per-kill slots can be filled)]
cards: []
requirements: [FR-TGT-001, FR-TGT-003, FR-TRN-003a, FR-MNV-001, FR-MOD-001, SVY-R2, SVY-R4, SVY-R11, SVY-R29, SVY-R53, SVY-R54]
---

## Goal
*"For every 2 units killed this way, deploy a different chosen unit card from your discard pile for free."* (*Mutation Swell*, directive #158 (2)). A target slot can be chosen once for every k assets the placement's counted kill cost (T-0264's `Cost.Counts`) kills, on placement, after the count is chosen. A slot marked *"different"* picks a different entity for each repetition and is legal only with that many different legal entities (`SVY-R54`); without it, the repetitions stay independent (`SVY-R4`, *Steel Rain*). The counts T-0264 offers, and the enact check, are narrowed to the counts whose per-kill slots can be filled (`SVY-R53`, `SVY-R29`). A timeout takes the first different legal entities in fixed order, top of the pile first for cards in a discard pile (`SVY-R54`, `FR-MOD-001` §6).

Covers (to be written in full against the code of the day before it leaves `queued`):
- `ability.Target` gains `PerKilled int` (k > 0: `pl.count / k` picks) and `Different bool` (with `PerKilled` or `PerDiscard`: the picks are distinct, `Repeats` false). The one rule "a slot repeated per cost unit" stays in `targetDecision` / `repeatedPicks`: `PerDiscard` and `PerKilled` give the number of repetitions, not two copies of the repeat logic (D2.0).
- `chooseTargets` and `slotsLegal` handle a per-kill slot; the legal counts (a `legalCounts(pl, e)` in `placement.go`, built on T-0264's `payableCounts`) feed `declare`'s `chooseCostCount` and `placeable`.
- `doc.go`: the targeting bullet (`:52-53`) names `PerKilled` and `Different`.
- Tests (synthetic Maneuvers, a counted kill cost {2, 4} and one per-kill slot of k = 2 on allied unit cards in the discard pile): not legal with 2 kills payable and no unit card in the pile; with 4 payable and only 1 unit card, only 2 is offered (asked nothing); with 4 and 2 cards, the slot asks for 2 different cards (`Min` = `Max` = 2, no repeats) and an answer repeating one card is invalid; the timeout takes the top two different cards of the pile; a `Different` `PerDiscard` slot picks distinct entities; `PerDiscard` without `Different` still repeats (`per_discard_test.go` unchanged).
- Allowed Files (expected): `pkg/engine/ability/ability.go`, `pkg/engine/targeting.go`, `pkg/engine/placement.go`, `pkg/engine/doc.go`.

## Log
- 2026-11-17 lead-architect (wave H run, session 26): queued with T-0264, split off under the D2 size rule. Write its Inputs, Tests and Allowed Files against the code of the day once T-0264 is `done`. Directive #158 (2).
