---
id: T-0271
title: "Animoceps Spore Cloud: one WouldBeDealtDamage replacement over allied units in the same zone (wave H card, flash)"
kind: card
tier: flash
wave: H
status: queued
branch: m2/autoloop-wave-h/T-0271-animoceps-spore-cloud
depends_on: [T-0269, T-0270]
claims: [card:Animoceps Spore Cloud]
cards: [Animoceps Spore Cloud]
requirements: [FR-REP-001, FR-REP-004, FR-CBT-002, SVY-R23, SVY-R44]
---

## Goal
*Animoceps Spore Cloud* (Animus Structure, "Redirect all damage that would be dealt to allied units in the same zone to this structure.") is executable: one T-0269 `ability.Replacement{Event: ability.WouldBeDealtDamage, Filter: {Allegiance: AllegianceAllied, Kinds: KindUnit, ZoneRef: ZoneRefSame}}`, written as the same-zone aura cards write their filter (`cards/animus.go:13-30`, *Animoceps Node*). It leaves `cards.NotYetExecutable`, which is then empty: M2's card done-when (design note `:211`, wave H). `pkg/engine/doc.go` moves damage redirection from "Not built yet" to the built list and drops the now-empty `NotYetExecutable` sentence's "still in that state" wording as fits. No new shape.

Covers: the definition in `pkg/engine/cards/animus.go`; the removal from `pkg/engine/cards/cards.go` (the slice stays, empty, since `defs_test.go` and `.agents/scripts/waverun` read it); `pkg/engine/doc.go`. Card tests in a new `pkg/engine/cards_animus_spore_cloud_test.go`: printed stats; effect damage to an allied unit in the Main Theater lands on the Cloud; combat damage to an allied defender lands on the Cloud with retaliation unchanged; a unit in Reserves is not covered; enemy units are not covered; two Clouds ask the unit's controller (`SVY-R23`).

## Inputs
To be written against the code of the day when this task leaves `queued` (after T-0269 and T-0270 merge): the card's CSV line (`data/cards/units-and-structures-animus.csv:18`), the `WouldBeDealtDamage` T-0269 adds, the aura filters in `cards/animus.go`, `NotYetExecutable` (`cards/cards.go:13-15`), and `doc.go`'s built and not-built lists (`:89-104` at e008353c).

## Tests (card tasks: written by rules-test-author; engineers never edit these)
To be written when the task leaves `queued`.

## Allowed Files
- `pkg/engine/cards/animus.go`
- `pkg/engine/cards/cards.go`
- `pkg/engine/doc.go`

## Done When
- [ ] The card's tests pass: `go test ./pkg/engine -run SporeCloud`
- [ ] `go tool task check` passes
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-compare -- autoloop/m2-wave-h` passes (the card is in no starter deck, and the default stress sim plays Soulless and Concordia)

## Log
- 2026-11-18 lead-architect (wave H run, session 31): queued with T-0269, split off under the D2 size rule. Depends on T-0269 (the shape) and T-0270 (the sim monitor reads its redirected combat damage).
