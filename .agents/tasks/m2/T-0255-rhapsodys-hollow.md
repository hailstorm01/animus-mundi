---
id: T-0255
title: "Rhapsody's Hollow: T3, exhaust, kill this structure: gain control of chosen opposing unit in the Main Theater (wave H card, flash)"
kind: card
tier: flash
wave: H
status: queued
branch: m2/autoloop-wave-h/T-0255-rhapsodys-hollow
depends_on: [T-0254]
claims: [card:Rhapsody's Hollow]
cards: [Rhapsody's Hollow]
requirements: [FR-ACT-001, FR-ACT-002, FR-ZON-002, FR-ZON-003, FR-TRN-003a, FR-TGT-003]
---

## Goal
*Rhapsody's Hollow* (Animus structure, `data/cards/units-and-structures-animus.csv:9`; in neither starter deck): "T3, exhaust, kill this structure: Gain control of chosen opposing unit in the Main Theater." becomes executable from shapes that are `done`: an activated ability with costs `CostResources` T3, `CostExhaustSelf` and `CostKillSelf` (as *Sinuous Growth*, `pkg/engine/cards/animus.go:537`), one target `{Allegiance: Opposing, Kinds: Unit, Zone: Main Theater}`, and T-0254's `OpGainControl` on it. The card leaves `cards.NotYetExecutable`.

## Covers (to be written against the code of the day before it leaves `queued`)
- Inputs, Shape, Tests and Allowed Files (`pkg/engine/cards/animus.go`, `pkg/engine/cards/cards.go`), written once T-0254 is `done`, with the names it merged.
- Tests to pin on the real card: an opposing Main Theater unit changes controller and keeps its state; the Hollow is in its owner's discard pile after paying; an opposing unit in Reserves, an allied unit and a structure are not offered; with no opposing unit in the Main Theater the ability isn't offered and nothing is paid (`SVY-R36`); a target that has left the Main Theater before resolution stays with its controller and the costs stay paid (`FR-TGT-003`, `FR-ACT-001`).

## Log
- 2026-11-16 lead-architect (wave H run, session 15): queued, split off T-0254 under the D2 size rule. Waits for T-0254.
