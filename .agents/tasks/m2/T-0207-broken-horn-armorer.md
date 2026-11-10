---
id: T-0207
title: Broken Horn Armorer (wave F deck card, flash; #141 (2))
kind: card
tier: flash
wave: F
status: queued
branch: m2/autoloop-wave-f/T-0207-broken-horn-armorer
depends_on: [T-0206]
claims: [card:Broken Horn Armorer]
cards: [Broken Horn Armorer]
requirements: [FR-TGT-001, FR-TGT-003, FR-ZON-002]
---

## Goal
*Broken Horn Armorer* (a Path deck card, 2 copies) executes: "Whenever this unit moves to a new zone, chosen asset in that zone gains 1 armor." (reworded by `SVY-R45`; `data/cards/units-and-structures-path.csv:8`). Directive #141 (2).

Covers:
- The definition in `pkg/engine/cards/path.go`: a `HookMoves` trigger (`ToZone` any), one target, any asset "in that zone" (T-0206's criterion; the Armorer itself may be chosen, as the text says no "other"), and an `OpAddArmor` 1 step on it. The name off `cards.NotYetExecutable`.
- Tests to pin (#141 (2)): after a March into the Main Theater, a chosen asset there gains 1 armor; if the Armorer moves again or leaves the battlefield before resolution, the target still gains the armor (in principle only in M2: set the stack up directly, as T-0205's tests do); a target that has left that zone is illegal (`FR-TGT-003`).

Inputs, Tests and Allowed Files are written against the code of the day before this task leaves `queued` (D2.4). Planned Allowed Files: `pkg/engine/cards/path.go`, `pkg/engine/cards/cards.go`.

## Log
- 2026-11-10 lead-architect (wave F run, session 5): split off T-0205 under the D2 size rule (#141 (2)); `queued` on T-0206.
