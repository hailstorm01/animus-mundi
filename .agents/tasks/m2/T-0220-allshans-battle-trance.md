---
id: T-0220
title: "Allshan's Battle Trance: deployed from the discard pile to the Main Theater, with Stalwart and +2 power until end of turn (wave F card, flash; #140 (2))"
kind: card
tier: flash
wave: F
status: queued
branch: m2/autoloop-wave-f/T-0220-allshans-battle-trance
depends_on: [T-0219]
claims: [card:Allshan's Battle Trance]
cards: [Allshan's Battle Trance]
requirements: [FR-TRN-003a, FR-LAY-001, FR-LAY-003, FR-MNV-001]
---

## Goal
*Allshan's Battle Trance* (Path Maneuver, R0 T4, mythic; in `data/decks/path-starter.txt`; FRD 02 benchmark #11): "Deploy chosen unit card from your discard pile to the Main Theater. Until end of turn, it gains Stalwart and gets +2 power." is executable from shapes that are `done`: a pile slot (as *Reaver's Rush*), `OpDeployFree` with `To` the Main Theater, and T-0219's `OpGrantKeyword` Stalwart and `OpModifyStat` +2 on the same slot. It leaves `cards.NotYetExecutable`.

## Covers (written by the lead at queue time; Inputs, Tests and Allowed Files are written against the code of the day before this task leaves `queued`)
- Its definition in `pkg/engine/cards/path.go` and its name removed from `cards.NotYetExecutable` in `pkg/engine/cards/cards.go`.
- `pkg/engine/doc.go`'s shape list names `OpGrantKeyword` (T-0219's Allowed Files had no room for it).
- Directive #140 (2)'s test: after Sub-Phase 4c the unit stays in the Main Theater without Stalwart and at printed power.
- Card tests from the CSV text: a unit printing only Reserves still enters the Main Theater (`FR-TRN-003a` 2); it has Stalwart and printed power + 2 that turn; its deploy triggers fire; *Veil Render* deployed this way pays no kill (`FR-TRN-003a`'s example); it can't be enacted without a unit card in its controller's own discard pile (`SVY-R29`); a target that left the pile before resolution deploys nothing and grants nothing.
- Check `TestView_Case8_ExecutableFlag` and the sim: if the card's removal from `NotYetExecutable` changes a pinned hand or a seed, name the edit under Tests (as T-0216 did).

## Log
- 2026-11-11, lead-architect (wave F run, session 21): queued, split off T-0219 under the D2 size rule (D2.4). Takes #140 (2) from T-0219, which now holds only the shape.
