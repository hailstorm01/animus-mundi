---
id: T-0261
title: "Brigand Highlander: \"Cornered: This unit has +2 power.\", on T-0260's `WhileCornered.Power` (wave H card, flash)"
kind: card
tier: flash
wave: H
status: queued
branch: m2/autoloop-wave-h/T-0261-brigand-highlander
depends_on: [T-0260]
claims: [card:Brigand Highlander]
cards: [Brigand Highlander]
requirements: [FR-KWD-002e, FR-LAY-001, FR-LAY-003]
---

## Goal
*Brigand Highlander* (Path unit, Officer, Hammerhand, R1 T1, Main Theater, 2/3, "Cornered: This unit has +2 power.") loads with no unsupported lines and has power 4 while its controller is Cornered, through T-0260's `ability.WhileCornered{Power: 2}` on `Def.Cornered`. It leaves `cards.NotYetExecutable`.

Covers: the card definition in `pkg/engine/cards/path.go` and its removal from `cards.NotYetExecutable` (`pkg/engine/cards/cards.go`). Inputs, Tests and Done When are written against the code of the day when T-0260 is `done`.

## Log
- 2026-11-16 lead-architect (wave H run, session 21): queued with T-0260, split off under the D2 size rule.
