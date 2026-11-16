---
id: T-0247
title: An untargeted pool split as it resolves, and Hammerhand Bivouac (wave H shape, sonnet; directive #157)
kind: shape
tier: sonnet
wave: H
status: queued
branch: m2/autoloop-wave-h/T-0247-bivouac-pool
depends_on: [T-0245, T-0246]
claims: [shape:untargeted distributed pool (a step that splits a fixed amount among the entities matching its filter as it resolves, positive whole amounts summing to the pool, SVY-R51; timeout puts the whole pool on the first in fixed order, SVY-R52)]
cards: [Hammerhand Bivouac]
requirements: [FR-TGT-002, FR-TGT-003, FR-MOD-001, FR-TRN-001]
---

## Goal
*Hammerhand Bivouac* ("At the start of your turn, distribute 2 armor among allied units in the same zone.") with the pool directive #157 rules on: an untargeted pool, split by its controller as the step resolves among the allied units in the zone T-0246 locks on placement (`SVY-R44`, `SVY-R51`), with the timeout of `SVY-R52`.

Covers (from directive #157; write the rest against the code of the day, after T-0246 merges, before this task leaves `queued`):
- The pool decision: positive whole amounts summing to exactly the pool; one match takes the whole pool with nothing asked (`FR-MOD-001` §1); none does nothing; a timeout puts the whole pool on the first option in `SubjectPick`'s fixed order (§6). Built so a later targeted pool (*Macrocannon Barrage*, outside M2) can split on declaration through the same amounts check, without building that path now.
- *Hammerhand Bivouac*'s definition: `HookTurnStart` with `Turn: TurnYours` (T-0245), the pool over allied units with `ZoneRefSame` (T-0246), `OpAddArmor`; off `cards.NotYetExecutable`.
- #157's tests (a)–(f), and marking design note Open Question 24 answered.

## Inputs
- To write before `ready`: FRD 01 v3.28 (`FR-TGT-002` line 409, `FR-TGT-003` line 416, `FR-MOD-001` §6 line 437, `SVY-R51`, `SVY-R52`), directive #157, and the code after T-0246.

## Tests
- To write before `ready`, from #157 (a)–(f).

## Allowed Files
- To write before `ready`.

## Done When
- [ ] `go tool task check` passes
- [ ] `go tool task sim` runs clean on the starter decks (*Bivouac* is in neither starter deck, so the baseline shouldn't move)

## Log
- 2026-11-15 lead-architect (wave H run, session 7): queued after T-0246, which it depends on for the locked zone. Directive #157 is `open: before T-0247` and is applied here.
