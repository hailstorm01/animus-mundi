---
id: T-0250
title: "Blindside: chosen opposing unit gets -3 power until the end of your opponent's turn (wave H card, flash)"
kind: card
tier: flash
wave: H
status: queued
branch: m2/autoloop-wave-h/T-0250-blindside
depends_on: [T-0249]
claims: [card:Blindside]
cards: [Blindside]
requirements: [FR-LAY-001, FR-LAY-003, FR-TRN-001, FR-MNV-001]
---

## Goal
*Blindside* (Path Maneuver, `data/cards/maneuvers-path.csv:5`; R0 T2; in neither starter deck): "Chosen opposing unit gets -3 power until the end of your opponent's turn." becomes executable from shapes that are `done`: one `AllegianceOpposing` unit target and `OpModifyStat` with T-0249's `Step.Until: ability.UntilEndOfOpponentTurn`. The card leaves `cards.NotYetExecutable`.

Covers (write Inputs, Tests and Allowed Files against the code after T-0249 merges, before this task leaves `queued`; check that the field and constant names below are the ones T-0249 merged):
- The definition in `pkg/engine/cards/path.go`, in alphabetical order (after *Blind Defile*): `Text` the CSV line; one target `{Filter: {Allegiance: AllegianceOpposing, Kinds: KindUnit}}` and one step `{Op: OpModifyStat, Amount: Value{N: -3}, Subject: SubjectTarget, Target: 0, Until: UntilEndOfOpponentTurn}`. No new shape, no new field.
- `pkg/engine/cards/cards.go`: remove "Blindside" from `NotYetExecutable`.
- Card tests on the real card (`s.card("Blindside")`), player 0 enacting: the opposing unit's power drops by 3 (clamped at 0 for a unit with power 2, `FR-LAY-001` Layer 6), stays down through player 1's turn (an attack by it in player 1's turn deals the reduced power), and is back to printed after player 1's turn ends; an allied unit isn't a legal target, and with no opposing unit *Blindside* can't be enacted (`FR-MNV-001`); a target that left the battlefield before resolution makes it fizzle (`FR-TGT-003`).
- Done When: `go tool task check`, `go tool task sim`, and `go tool task sim-compare -- autoloop/m2-wave-h` (in neither starter deck).

## Log
- 2026-11-15 lead-architect (wave H run, session 10): queued, split off T-0249 under the D2 size rule (shape and card together touch six non-test files). Waits for T-0249.
