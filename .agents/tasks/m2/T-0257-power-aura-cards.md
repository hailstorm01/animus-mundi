---
id: T-0257
title: "Animoceps Stalk, Animoceps Node, Melodic Pummeler: same-zone power auras (wave H cards, flash)"
kind: card
tier: flash
wave: H
status: queued
branch: m2/autoloop-wave-h/T-0257-power-aura-cards
depends_on: [T-0256]
claims: [card:Animoceps Stalk, card:Animoceps Node, card:Melodic Pummeler]
cards: [Animoceps Stalk, Animoceps Node, Melodic Pummeler]
requirements: [FR-LAY-001, FR-TGT-001]
---

## Goal
The three power auras of wave E execute, each one `ability.StaticAura` built by T-0256 with a different `Filter` (identical shapes, so one card task under the D2 size rule):
- *Animoceps Stalk*: "Other allied assets in the same zone have +1 power." (allied, `Other`, `ZoneRefSame`, Power 1; survey `:162`, `excludeSelf`).
- *Animoceps Node*: "Allied units in the same zone have +1 power." (allied units, `ZoneRefSame`, Power 1; survey `:163`: a Structure, so it doesn't buff itself).
- *Melodic Pummeler*: "Allied Horror units in the same zone have +1 power." (allied units with subtype Horror, `ZoneRefSame`, Power 1; survey `:164`: includes itself).

Each leaves `cards.NotYetExecutable`.

## Covers (to be written against the code of the day before this task leaves `queued`)
- Inputs, Tests and Acceptance: `TestCard_AnimocepsStalk_*`, `TestCard_AnimocepsNode_*`, `TestCard_MelodicPummeler_*`, from the CSV text: who gains the power and who doesn't (self, other zone, opposing, non-unit for Node and Pummeler, non-Horror for Pummeler), and that moving the source moves the aura.
- Allowed Files: `pkg/engine/cards/animus.go`, `pkg/engine/cards/cards.go`.
- Done When: the card tests, `go tool task check`, `go tool task sim`, and `go tool task sim-compare -- autoloop/m2-wave-h` (none of the three is in a starter or stress deck).

## Log
- 2026-11-16 lead-architect (wave H run, session 17): queued with T-0256, split off under the D2 size rule. Written in full once T-0256 is `done`.
