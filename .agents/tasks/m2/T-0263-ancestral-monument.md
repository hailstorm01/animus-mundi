---
id: T-0263
title: "Ancestral Monument: \"Other allied assets in the same zone have Cornered: Resistant 1.\", on T-0262's Static.Cornered (wave H card, flash)"
kind: card
tier: flash
wave: H
status: queued
branch: m2/autoloop-wave-h/T-0263-ancestral-monument
depends_on: [T-0262]
claims: []
cards: [Ancestral Monument]
requirements: [FR-LAY-001, FR-KWD-002e, SVY-R18, SVY-R44]
---

## Goal
*Ancestral Monument* (Path structure, `data/cards/units-and-structures-path.csv:22`): *"Other allied assets in the same zone have Cornered: Resistant 1."* One `StaticAura` with `Filter{Allegiance: AllegianceAllied, Other: true, ZoneRef: ZoneRefSame}` (any asset) and T-0262's `Cornered: ability.WhileCornered{Resistant: 1}`. The card leaves `cards.NotYetExecutable`.

Covers: the card's definition in `pkg/engine/cards/path.go`, its removal from `NotYetExecutable` in `pkg/engine/cards/cards.go`, and `TestCard_AncestralMonument_*` tests against the real card. Inputs, Tests and Allowed Files are written against the code of the day before it leaves `queued`.

## Log
- 2026-11-16 lead-architect (wave H run, session 23): queued, split off from T-0262 under the D2 size rule.
