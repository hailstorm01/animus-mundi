---
id: T-0223
title: Devilhorn Scouting Party (wave F deck card, flash)
kind: card
tier: flash
wave: F
status: queued
branch: m2/autoloop-wave-f/T-0223-devilhorn-scouting-party
depends_on: [T-0221, T-0222]
claims: [card:Devilhorn Scouting Party]
cards: [Devilhorn Scouting Party]
requirements: [FR-MOD-001, FR-TRN-003a, FR-ZON-005]
---

## Goal
*Devilhorn Scouting Party* (Path Maneuver; in `data/decks/path-starter.txt`): "Look at the top 3 cards of your deck. You may deploy a unit card with Requisition cost 3 or less from among them for free. Put the rest on the bottom of your deck in any order. Then, Scratch 1." is executable from shapes that are `done`: T-0222's `OpLookTop` (`Amount` 3, `UpTo` 1, `Filter` unit cards with T-0221's `StatRequisition` at most 3), then `OpScratch` 1. It leaves `cards.NotYetExecutable`.

Covers (Inputs, Tests and Allowed Files are written against the code of the day before this leaves `queued`):
- Its `Def` in `pkg/engine/cards/path.go` and its removal from `cards.NotYetExecutable`. T-0222 already named `OpLookTop` and `DecideLook` in `doc.go`.
- From T-0222's review (a comment-only fix in `pkg/engine/doc.go`): put back the separators its two lists lost, so the decisions line reads "… from your own deck (DecideSearch, SVY-R37, SVY-R38), the zone of a free deploy (DecideZone, SVY-R43) and a look at the top cards of your own deck (DecideLook, SVY-R41–SVY-R43)" and the step list reads "… OpMorphScratch (a face-down transformation, FR-ENT-103), OpGrantKeyword (a keyword granted until end of turn, FR-LAY-001) and OpLookTop (…)".
- Card tests: a cost-3 unit among the top 3 can be deployed for free into a zone it prints, a cost-4 unit and a Maneuver are never offered, the rest go to the bottom in the chosen order, Scratch 1 follows either way; a timeout deploys nothing, keeps the order seen, and still Scratches 1 (`SVY-R41`, `SVY-R42`).
- Whether `TestView_Case8_ExecutableFlag` needs an authorized edit (as T-0216's did) is checked by a local trial when the task is written.

## Log
- 2026-11-11 lead-architect (wave F run, session 23): split off T-0221 under the D2 size rule, with T-0222; `queued` on T-0221 and T-0222.
- 2026-11-12 lead-architect (wave F run, session 24): T-0222 is `done` (merge 143be71); its `doc.go` separator nit added to the Covers.
