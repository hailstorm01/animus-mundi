---
id: T-0169
title: Make Cliff of Portents executable (R1, T1, exhaust, search your deck for a card; wave D card, flash)
kind: card
tier: flash
wave: D
status: queued
branch: m2/autoloop-wave-d/T-0169-cliff-of-portents
depends_on: [T-0158, T-0163, T-0167, T-0168]
claims: [card:Cliff of Portents]
cards: [Cliff of Portents]
requirements: [FR-ACT-001, FR-ACT-002, FR-DEK-001, FR-DEK-002]
---

## Goal
*Cliff of Portents* (a deck card, 1 copy): *"R1, T1, exhaust: Search your deck for a card, put it into your hand, then shuffle your deck."* It only combines T-0158's activated ability and costs with T-0167's `OpSearchDeck`. The last card in the wave run's header, and the last of #105: its parts (2) and (3) are tested here on the real card.

Covers (write the Inputs and Tests against the code of the day, once T-0168 has merged):
- The Definition below in `pkg/engine/cards/animus.go`, and the name off `cards.NotYetExecutable` in `pkg/engine/cards/cards.go`.
- Card tests from the CSV text, in a new `pkg/engine/cards_animus_cliff_test.go`, on `pkg/engine/cards_animus_sinuous_test.go`'s model: the chosen card goes to hand and the pool loses exactly 1 Requisition and 1 Tactics; the Structure ends Exhausted; a timeout takes the top card (`SVY-R38`); an empty deck still activates, pays, asks nothing and shuffles (`SVY-R36`, `SVY-R37`); not offered without either resource or while Exhausted (`FR-ACT-002`).

### Definition (the engineer adds it to `animusDefs`, before *Craterfall Ancient*)
- `Text`: the printed line, as above.
- `Activated`: one ability. `Costs`, in printed order: `{Kind: ability.CostResources, Req: 1, Tac: 1}`, `{Kind: ability.CostExhaustSelf}`. `Effect.Steps`: one, `{Op: ability.OpSearchDeck}`. No target slot.

A lead probe with this Definition on T-0167's Shape, since reverted, passes `go tool task check`, and the starter sim stays clean: 47/50 Animus main cards executable, P0 84.0% / P1 15.0% / draws 1.0% over 200 games, no monitor violation. Only one of those games activates it, so a clean sim says little about the card, and its tests have to carry it. `sim-compare` will differ here and isn't a gate.

## Log
- 2026-11-04 lead-architect (wave D run, session 12): split off T-0167 (D2.4). Queued.
