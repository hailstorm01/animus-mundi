---
id: T-0226
title: "Eerie Remnant: when put into your discard pile from your hand, you may deploy it from there for free (wave F deck card, flash; #136's card part)"
kind: card
tier: flash
wave: F
status: queued
branch: m2/autoloop-wave-f/T-0226-eerie-remnant
depends_on: [T-0225]
claims: [card:Eerie Remnant]
cards: [Eerie Remnant]
requirements: [FR-ZON-004, FR-TRN-003a]
---

## Goal
*Eerie Remnant* (Path unit, `data/cards/units-and-structures-path.csv:4`; 3 copies in `data/decks/path-starter.txt`): "Deploys Exhausted / When this card is put into your discard pile from your hand, you may deploy it from your discard pile for free." is executable from shapes that are `done`: T-0225's `HookDiscardedFromHand` with `ActiveZone` `ZoneDiscard`, `Optional`, and T-0208's `OpDeployFree` on `SubjectSelf`. It leaves `cards.NotYetExecutable`, the last wave F deck card to do so. With it, directive #136 is fully applied.

Covers (split from T-0225 on 2026-11-12 under the D2 size rule; Inputs, Tests and Allowed Files are written against the code of the day before this task leaves `queued`):
- The definition in `pkg/engine/cards/path.go`: one trigger, `Hook: HookDiscardedFromHand`, `ActiveZone: ZoneDiscard`, `Optional: true`, one step `OpDeployFree` with `SubjectSelf` and no `To` (the card prints Reserves only, `FR-TRN-003a` 1). Deploys Exhausted is its keyword (`TestFR_TRN_003a_DeployFreeKeepsDeploysExhausted` pins it for a free deploy).
- The removal of *Eerie Remnant* from `cards.NotYetExecutable`.
- `pkg/engine/doc.go`: the hook list names `HookDiscardedFromHand`, and the triggers line says a trigger may work from its card's discard pile (`Trigger.ActiveZone`, survey G1).
- Card tests: discarded from hand by *Chainbreaker Irregulars*' Unbowed draw-then-discard or another deck card's discard (cost or step), it may be deployed Exhausted into Reserves; declined or timed out, it stays in the pile; killed on the battlefield, it fires nothing.
- Before writing the Tests, try the definition locally (as T-0223 did) to find existing tests that need an authorized edit (`TestView_Case8_ExecutableFlag`, replay or sim-policy tests) and whether the sim's policy answers the optional question.

## Log
- 2026-11-12, lead-architect (wave F run, session 27): split from T-0225 under the D2 size rule, `queued` on it, Covers only.
