---
id: T-0248
title: "Hammerhand Bivouac: at the start of your turn, distribute 2 armor among allied units in the same zone (wave H card, flash; directive #157)"
kind: card
tier: flash
wave: H
status: queued
branch: m2/autoloop-wave-h/T-0248-hammerhand-bivouac
depends_on: [T-0245, T-0246, T-0247]
claims: [card:Hammerhand Bivouac]
cards: [Hammerhand Bivouac]
requirements: [FR-TRN-001, FR-TGT-001, FR-TGT-002, FR-MOD-001]
---

## Goal
*Hammerhand Bivouac* (Path structure, `data/cards/units-and-structures-path.csv:17`; Territory, Hammerhand; R3 T0, Main Theater, 0/7, yields 2/2; in neither starter deck): "At the start of your turn, distribute 2 armor among allied units in the same zone." becomes executable from shapes that are `done`: T-0245's `HookTurnStart` with `Turn: TurnYours`, T-0246's `ZoneRefSame` and T-0247's `Step.Split` on `OpAddArmor`. The card leaves `cards.NotYetExecutable`. With it, directive #157 is fully applied.

Covers (write Inputs, Tests and Allowed Files against the code after T-0247 merges, before this task leaves `queued`; check that the field names below are the ones T-0247 merged):
- The definition in `pkg/engine/cards/path.go`, in alphabetical order (before *Hammerhand Brigands*): `Text` the CSV ability line; one trigger `{Hook: ability.HookTurnStart, Turn: ability.TurnYours}` whose effect has no targets and one step `{Op: ability.OpAddArmor, Amount: ability.Value{N: 2}, Subject: ability.SubjectPick, Split: true, Filter: ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit, ZoneRef: ability.ZoneRefSame}}`. No new shape, no new field.
- `pkg/engine/cards/cards.go`: remove "Hammerhand Bivouac" from `NotYetExecutable`.
- Card tests on the real card (`s.card("Hammerhand Bivouac")`) for player 0 in the Main Theater, at the start of player 0's turn: #157 (a) a split between two allied Main Theater units; (b) one unit takes both, unasked; (c) no allied unit in the Main Theater: nothing; (d) a timeout with three allied units puts both on the one that entered the Main Theater earliest; (e) allied units in Reserves and the Headquarters aren't offered; (f) the armor persists: still there after player 0's turn ends and through player 1's turn, and the next start of player 0's turn adds to it; plus it does nothing at the start of player 1's turn.
- Done When: `go tool task check`, `go tool task sim`, and `go tool task sim-compare -- autoloop/m2-wave-h` (in neither starter deck).

## Log
- 2026-11-15 lead-architect (wave H run, session 8): queued, split off T-0247 under the D2 size rule (shape and card together touch six non-test files). Waits for T-0247.
