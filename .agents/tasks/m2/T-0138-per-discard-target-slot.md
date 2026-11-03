---
id: T-0138
title: A target slot can choose once per card the discard cost discarded (FR-TGT-001, SVY-R4, SVY-R32, directive #20 (b), wave C shape, sonnet)
kind: shape
tier: sonnet
wave: C
status: queued
branch: m2/autoloop-wave-c/T-0138-per-discard-target-slot
depends_on: [T-0136, T-0137]
claims: [field:Target.PerDiscard, func:chooseTargets, func:targetDecision, func:slotsLegal]
cards: []
requirements: [FR-TGT-001, FR-TRN-003a, FR-MNV-001, FR-MOD-001]
---

## Goal
A Maneuver can print *"For each card discarded this way, deal 2 damage to chosen asset."* (*Steel Rain*, a deck card). On declaration its controller makes one target choice for each card its discard cost discarded, independently, so the same asset may be chosen more than once, and each repetition is its own damage instance (`SVY-R4`, `FR-REP-004`). A timeout puts them all on the first legal entity in fixed order (`SVY-R32`, T-0136). Split from #20 (b) under the D2 size rule; the *Steel Rain* card task follows it.

## Covers (to be written out against the code of the day before this task leaves `queued`, D2.4)
- `ability.Target.PerDiscard bool`: the slot is chosen once per card in `placement.discards` (T-0137), each an exactly-one choice, repeats allowed; `UpTo` stays 0 with it.
- `chooseTargets` / `targetDecision` (`targeting.go`): a `PerDiscard` slot asks one `DecideTarget` with `Min = Max =` the number of cards discarded and `Repeats`. With none discarded, or no legal entity, it asks nothing and the slot is empty. Its Refs may repeat, and the second verification stage and `stepAssets` already keep repeated entries (`effects.go` `resolveAbility`, `slotAssets`), so each repetition deals its damage separately; the tests pin that.
- `slotsLegal` (`placement.go`): a `PerDiscard` slot never makes a Maneuver illegal to enact, because discarding none pays the cost and then needs no target (lead's reading of `SVY-R29`, which names exactly-one slots; record it in the task's Inputs).
- Tests to name then: 0, 1 and 3 discards give 0, 1 and 3 choices; all three on one asset deal three 2-damage instances; a timeout concentrates on the first legal asset in fixed order (`SVY-R32`); Ward assets are never offered; a trigger (no costs) with a `PerDiscard` slot chooses nothing.
- `go tool task sim-compare -- autoloop/m2-wave-c` passes: no card uses `PerDiscard` yet.

## Allowed Files (planned)
- `pkg/engine/ability/ability.go` (`Target.PerDiscard`)
- `pkg/engine/targeting.go` (`chooseTargets`, `targetDecision`)
- `pkg/engine/placement.go` (`slotsLegal`)

## Log
- 2026-11-02 lead-architect (wave C session 27): queued with T-0136 when #20 (b) was split (D2.4). Write its Inputs, Tests and Done When against the code of the day before it leaves `queued`.
