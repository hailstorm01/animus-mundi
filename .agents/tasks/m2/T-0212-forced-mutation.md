---
id: T-0212
title: Forced Mutation, turning a chosen opposing unit face down (wave F deck card, flash)
kind: card
tier: flash
wave: F
status: queued
branch: m2/autoloop-wave-f/T-0212-forced-mutation
depends_on: [T-0211]
claims: [card:Forced Mutation]
cards: [Forced Mutation]
requirements: [FR-ENT-103, FR-MNV-001, FR-TGT-001]
---

## Goal
*Forced Mutation* (Animus Maneuver, R0 T4, two copies in `data/decks/animus-starter.txt`): "Turn chosen opposing unit face down. It becomes a Scratch Company unit." One exactly-one target (opposing unit) and one `OpMorphScratch` step on it (T-0211). It combines shapes only.

Covers:
- Its definition in `pkg/engine/cards/animus.go`, between *Feed the Despoiler* and *Funeral Fungus*, and its name removed from `cards.NotYetExecutable`.
- Card tests from the CSV text: the target is turned face down in place and keeps its state (`FR-ENT-103`); a damaged target dies in the sweep after the item; only opposing units are offered; it can't be enacted without an opposing unit (`SVY-R29`, `FR-MNV-001`).

Inputs, Tests and the definition are written against the code of the day, after T-0211 is `done` and before this task leaves `queued`, with a lead probe of the definition first.

## Allowed Files
- `pkg/engine/cards/animus.go` (the definition)
- `pkg/engine/cards/cards.go` (remove the name from `NotYetExecutable`)

## Log
- 2026-11-10 lead-architect (wave F run, session 11): queued with T-0211 (the `OpMorphScratch` shape), split under the D2 size rule.
