---
id: T-0205
title: A "same zone" filter, first card Broken Horn Armorer (wave F shape, sonnet)
kind: shape
tier: sonnet
wave: F
status: queued
branch: m2/autoloop-wave-f/T-0205-broken-horn-armorer
depends_on: [T-0203]
claims: [shape:same-zone filter (ability.Filter same-zone criterion), card:Broken Horn Armorer]
cards: [Broken Horn Armorer]
requirements: [FR-TGT-001, FR-ZON-002]
---

## Goal
*Broken Horn Armorer* (a Path deck card, 2 copies) executes: "Whenever this unit moves to a new zone, chosen asset in the same zone gains 1 armor." It needs a filter criterion `FR-TGT-001` names ("in the same zone", line 370) and the design note's Key types list for `Filter` ("same zone as source"), which no card has used yet.

Covers:
- A same-zone criterion on `ability.Filter`, read in `matchesFilter` ([filter.go:85](../../../pkg/engine/filter.go)): the entity is in its source's battlefield zone. The Main Theater is shared, and Reserves and the Headquarters Zone are each player's own (`FR-ZON-001`), so in those the entity's controller is the source's too. Whether a source that has left the battlefield reads its last-known zone (`itemSource`) is settled when this task is written, against FRD 01 then. The deferred same-zone auras (*Animoceps Stalk*, *Animoceps Node*, *Melodic Pummeler*, *Hulking Incongruity*, *Ancestral Monument*, #14 (b)) and *Hammerhand Bivouac* reuse it.
- The definition in `pkg/engine/cards/path.go`: a `HookMoves` trigger (`ToZone` any), one target, any asset in the same zone (itself included: the text says no "other"), and an `OpAddArmor` 1 step on it. The name off `cards.NotYetExecutable`.

Inputs, Tests and Allowed Files are written against the code of the day before this task leaves `queued` (D2.4).

## Log
- 2026-11-09 lead-architect (wave F run, session 2): split off T-0203 under the D2 size rule; `queued` on it.
- 2026-11-09 lead-architect (wave F run, session 4): not written yet, stays `queued`. Its Covers left the departed-source reading of "same zone" to FRD 01, and FRD 01 settles neither a source that moved to the other zone nor one that left the battlefield before the second target check (`FR-TGT-003`, `FR-ZON-004`). Ruling question queued on `main` in `docs/design/outcomes/2026-11-09-wave-f-same-zone-source.md` (111c8fb), design note Open Question 20. Write Inputs, Tests and Allowed Files once its `Status:` reads `applied`.
