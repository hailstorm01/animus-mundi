---
id: T-0216
title: Veil Render, deployed from its owner's discard pile by killing an allied asset (wave F card, flash)
kind: card
tier: flash
wave: F
status: queued
branch: m2/autoloop-wave-f/T-0216-veil-render
depends_on: [T-0213, T-0214, T-0215, T-0217]
claims: [card:Veil Render]
cards: [Veil Render]
requirements: [FR-TRN-003, FR-TRN-003a, FR-CST-001]
---

## Goal
*Veil Render* ("You may deploy this card from your discard pile by killing an allied asset in addition to paying its deployment costs.", `data/cards/units-and-structures-animus.csv` line 17; FRD 02 benchmark #6) is executable: one discard deploy permission (T-0214) whose cost is one `CostKill` with `Filter{Allegiance: AllegianceAllied}` (an asset: any allied Unit, Structure or Headquarters, as the benchmark's `entityType: ASSET` says). It leaves `cards.NotYetExecutable`. An Animus starter deck card (1 copy).

Covers (Tests written against the code of the day before it leaves `queued`):
- From the discard pile it pays R3 T3 and kills the chosen allied asset after the push; the kill's death triggers resolve above it (`FR-TRN-003`'s example); a replaced kill still pays (`SVY-R15`).
- From hand it deploys as a plain Unit to Reserves, with no kill.
- *Allshan's Battle Trance* or *Reaver's Rush*-style free deploys waive the kill (`FR-TRN-003a`'s example) where a printed card can do it today (*Reaver's Rush* can't: power 5).
- Not legal from the opponent's pile, or with no allied asset to kill.
- `pkg/engine/doc.go`'s shape list gains a line for `Def.Statics`: a card's own discard-active deploy permission (`ability.StaticDeployPermission`), deployed as an `ActDeploy` with `From: ZoneDiscard` and its `Costs` paid on declaration (`FR-TRN-003`, `FR-TRN-003a`; from T-0214, whose Allowed Files had no room for it).

## Log
- 2026-11-10 lead-architect (wave F run, session 13): queued with T-0213 (D2.4). Directive #128 is `open: before T-0216`, applied by T-0213, T-0214 and T-0215.
- 2026-11-11 lead-architect (wave F run, session 16): T-0215 split under the D2 size rule; the play screen's label went to T-0217, now in `depends_on`.
