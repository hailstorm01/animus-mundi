---
id: T-0242
title: Blind Defile (wave H card, flash; survey wave B)
kind: card
tier: flash
wave: H
status: queued
branch: m2/autoloop-wave-h/T-0242-blind-defile
depends_on: [T-0241]
claims: [card:Blind Defile]
cards: [Blind Defile]
requirements: [FR-CBT-002, FR-TRG-003, FR-TRG-005, FR-TGT-001, FR-MOD-001]
---

## Goal
*Blind Defile* (Path structure, `data/cards/units-and-structures-path.csv:7`): "Whenever an opposing unit attacks, you may kill this structure and that attacking unit." An optional `HookAttacks` trigger with `Watches` on opposing units (T-0241), whose steps kill the Defile (`SubjectSelf`) and the attacker (`SubjectWatched`). It leaves `cards.NotYetExecutable`.

Covers:
- The card definition in `pkg/engine/cards/path.go`, and its name off `NotYetExecutable` in `cards/cards.go`.
- `pkg/engine/doc.go`: the triggered-abilities line names watching triggers beside the self hooks (T-0241 left it to this task).
- Whether the attacker's kill is `IfDone` on the Defile's, and the card text if it changes: as the ruling question [`2026-11-15-wave-h-blind-defile-kill-both`](../../../docs/design/outcomes/2026-11-15-wave-h-blind-defile-kill-both.md) is answered (design note Open Question 22).

Inputs, Tests and Allowed Files are written against the code of the day once T-0241 is `done` and that outcome's `Status:` reads `applied`, before this task leaves `queued`.

## Log
- 2026-11-15 lead-architect (wave H run, session 1): split off T-0241 under the D2 size rule; `queued` on it and on the ruling question committed on `main` (07bd2da).
