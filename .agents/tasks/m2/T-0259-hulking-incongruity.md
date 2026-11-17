---
id: T-0259
title: "Hulking Incongruity: allied Horror units in the same zone have Resistant 1 (wave H card, flash)"
kind: card
tier: flash
wave: H
status: queued
branch: m2/autoloop-wave-h/T-0259-hulking-incongruity
depends_on: [T-0258]
claims: [card:Hulking Incongruity]
cards: [Hulking Incongruity]
requirements: [FR-LAY-001, FR-KWD-002b, SVY-R44]
---

## Goal
*Hulking Incongruity* executes: "Allied Horror units in the same zone have Resistant 1." One `ability.StaticAura` (T-0256) with T-0258's `Static.Resistant` 1 and the filter *Melodic Pummeler* uses (allied units with subtype Horror, `ZoneRefSame`), Power 0. Survey `:165`: it is a Plant, Horror unit, so it includes itself. The task removes it from `cards.NotYetExecutable`.

Covers: its `Def` in `pkg/engine/cards/animus.go` (its `Text` line exactly as in `data/cards/units-and-structures-animus.csv`), its removal from `NotYetExecutable` in `pkg/engine/cards/cards.go`, and card tests (it grants itself and an allied Horror unit in its zone Resistant 1, not an allied non-Horror unit or one in another zone; damage to a covered unit is reduced by 1). Inputs, Tests and Allowed Files are written against the code of the day once T-0258 is `done`, before this task leaves `queued`.

## Log
- 2026-11-16 lead-architect (wave H run, session 19): queued with T-0258, split off under the D2 size rule.
