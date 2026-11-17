---
id: T-0268
title: "Solarheart Monk: an enter-discard trigger, chosen asset heals 2 (wave H card, flash)"
kind: card
tier: flash
wave: H
status: queued
branch: m2/autoloop-wave-h/T-0268-solarheart-monk
depends_on: [T-0267]
claims: [card:Solarheart Monk]
cards: [Solarheart Monk]
requirements: [FR-ZON-004, FR-ZON-005, FR-KWD-002d, FR-TGT-001]
---

## Goal
*Solarheart Monk* (Path unit, "When this card is put into your discard pile, chosen asset heals 2.") is executable: one `HookEntersDiscard` trigger (T-0267) with `ActiveZone: ZoneDiscard`, one target slot over any asset and one `OpHeal` 2, written as *Spoil Forest Colossus* writes its heal (`cards/animus.go:690-707`). It leaves `cards.NotYetExecutable`.

Covers, to be written against the code of the day before it leaves `queued` (D2.4): the card definition in `pkg/engine/cards/path.go`, its removal from `NotYetExecutable` in `pkg/engine/cards/cards.go`, and card tests in `pkg/engine/cards_path_solarheart_monk_test.go`: its printed stats load; discarded from hand it heals a chosen asset 2; destroyed on the battlefield it heals; as a destroyed face-down Scratch Company it heals (`FR-ZON-004` `:267`); a gained-control Monk destroyed is its owner's trigger; returned to hand it does nothing.

## Inputs
- To be written with the task.

## Tests
- To be written with the task.

## Allowed Files
- `pkg/engine/cards/path.go`
- `pkg/engine/cards/cards.go`

## Done When
- [ ] The listed tests pass
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] The card loads with no unsupported lines

## Log
- 2026-11-17 lead-architect (wave H run, session 29): queued, split off T-0267 under the D2 size rule. Write its Inputs, Tests and Allowed Files against the code of the day once T-0267 is `done`.
