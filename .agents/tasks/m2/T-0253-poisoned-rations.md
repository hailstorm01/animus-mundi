---
id: T-0253
title: "Poisoned Rations: chosen unit gets -3 max health until end of turn (wave H card, flash)"
kind: card
tier: flash
wave: H
status: queued
branch: m2/autoloop-wave-h/T-0253-poisoned-rations
depends_on: [T-0251, T-0252]
claims: [card:Poisoned Rations]
cards: [Poisoned Rations]
requirements: [FR-LAY-001, FR-LAY-002, FR-LAY-003, FR-TRN-001, FR-MNV-001]
---

## Goal
*Poisoned Rations* (Path Maneuver, `data/cards/maneuvers-path.csv:10`; R0 T3; in neither starter deck): "Chosen unit gets -3 max health until end of turn." becomes executable from shapes that are `done`: one unit target of any allegiance and `OpModifyStat` with T-0252's `Stat: ability.StatMaxHealth`, `Amount` -3. The card leaves `cards.NotYetExecutable`.

## Covers (write Inputs, Shape, Tests and Allowed Files against the code of the day before this leaves `queued`)
- `pkg/engine/cards/path.go`: the definition, in alphabetical order. `pkg/engine/cards/cards.go`: remove "Poisoned Rations" from `NotYetExecutable`.
- Tests to pin, on the real card: an enemy unit with Health 3 or less is destroyed at the sweep after it resolves, with no damage dealt (`FR-LAY-002`, which names the card); a unit with Health 5 and 2 damage dies (Current Health 0); a unit with Health 5 and no damage survives, and has Maximum Health 5 again after the turn ends, with its damage unchanged; an allied unit is a legal target too ("chosen unit"); a structure is not, and with no unit on the battlefield it isn't enactable and stays in hand (`FR-MNV-001`).

## Log
- 2026-11-15 lead-architect (wave H run, session 12): queued, split off T-0251 under the D2 size rule. Waits for T-0251 and T-0252.
