---
id: T-0204
title: Chainbreaker Phalanx (wave F card, flash)
kind: card
tier: flash
wave: F
status: queued
branch: m2/autoloop-wave-f/T-0204-chainbreaker-phalanx
depends_on: [T-0203]
claims: [card:Chainbreaker Phalanx]
cards: [Chainbreaker Phalanx]
requirements: [FR-ZON-002, FR-TRG-004]
---

## Goal
*Chainbreaker Phalanx* (a Path deck card, 2 copies) executes: "Whenever this unit moves to the Main Theater, you may move chosen allied unit from your Reserves to the Main Theater. That unit gains 2 armor." It only combines shapes that exist once T-0203 is `done`.

Covers:
- The definition in `pkg/engine/cards/path.go`: an optional `HookMoves` trigger with `ToZone: ZoneMainTheater`; one target, an allied unit in Reserves; an `OpMoveZone` step to the Main Theater and an `OpAddArmor` 2 step, both on that target. The name off `cards.NotYetExecutable`.
- Card tests (rules-test-author): marching it into the Main Theater offers the move and the armor; declining does nothing; marching to Reserves or deploying it (`SVY-R1`) fires nothing; no allied unit in Reserves leaves nothing to choose; a second Phalanx moved this way fires its own trigger (survey line 199, `FR-TRG-004`).

Inputs, Tests and Allowed Files are written against the code of the day before this task leaves `queued` (D2.4).

## Log
- 2026-11-09 lead-architect (wave F run, session 2): split off T-0203 under the D2 size rule; `queued` on it.
