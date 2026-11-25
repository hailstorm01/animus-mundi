---
id: T-0300
title: Make Icewind Regent executable (exhaust to return a chosen unit, wave A, card, flash)
kind: card
tier: flash
wave: A
status: tests
branch: m3/autoloop-wave-a/T-0300-icewind-regent
depends_on: [T-0299]
claims: [card:Icewind Regent]
cards: [Icewind Regent]
requirements: [FR-ACT-001, FR-TGT-001, FR-ZON-004]
---

## Goal
*Icewind Regent* (a Concordia unit, R3 T3, 5/8, Reserves; 1 copy in the starter deck): *"Exhaust: Return chosen unit to its owner's hand."* It only combines shapes that are `done`: an activated ability whose cost is `CostExhaustSelf` (*Cliff of Portents*) and whose effect is one unit target and one `OpReturnToHand` step (*Ambush Tactics*' second mode).

Covers:
- The definition in `pkg/engine/cards/concordia.go`, as under Definition below.
- The name off `cards.NotYetExecutable`.

## Inputs
- Card (`data/cards/units-and-structures-concordia.csv` line 14).
- Survey row: `docs/design/survey/m3_full_catalog_ability_survey.md` §6, "Concordia: deck" (*Icewind Regent*: "May choose itself").
- FRD 01: `FR-ACT-001` and `FR-ACT-002` (lines 388-391), `FR-TGT-001` (lines 407-413), `FR-TGT-003` (line 436), `FR-ZON-004` (line 279), `FR-ZON-005` (line 290), `FR-KWD-002h` Ward (line 648), `FR-ZON-003` (ownership and control, line 250).
- Rulings: `SVY-R36` (an ability needs its required targets). "chosen unit" has no allegiance, zone or exclusion (`FR-TGT-001`), so any unit on the battlefield is legal, the source included.
- Design note: `docs/architecture/m3-full-catalog.md`, Boundaries (`pkg/engine/cards`), Deliverable 5.

### Definition (the engineer adds it to `concordiaDefs`)
- `"Icewind Regent"`, `Text`: `"Exhaust: Return chosen unit to its owner's hand."`.
- One `Activated`: `Costs` is `{Kind: ability.CostExhaustSelf}` only. `Effect`: `Targets` is one exactly-one slot with `Filter{Kinds: ability.KindUnit}`; `Steps` is one `{Op: ability.OpReturnToHand, Subject: ability.SubjectTarget, Target: 0}`.

A lead probe with the definition on the run branch at 762f2ded, since reverted, passed `go tool task check` and `go tool task sim-matrix`.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
In a new file `pkg/engine/cards_concordia_regent_test.go`, from the CSV text, loading the real card (`s.card`), in the style of `cards_animus_rhapsodys_hollow_test.go`. Player 0 controls a Readied *Icewind Regent* in its Reserves and is the active player, with an empty pool. Reuse the existing test helpers (`activate`, `offersActivate`, `actNotOffered`, `setPool`, `refs`, `sortedCopy`, `putScratch`).
- `TestCard_IcewindRegent_ReturnsChosenOpposingUnit` (`FR-ACT-001`, `FR-ZON-004`): a chosen damaged opposing unit in the Main Theater leaves the battlefield and its card is in its owner's hand, not the discard pile or the controller's hand. The Regent is Exhausted, and no Requisition or Tactics is spent. The other units are untouched.
- `TestCard_IcewindRegent_AnyUnitInAnyZoneOffered` (`FR-TGT-001`): the target options are exactly all units on the battlefield: allied and opposing, Main Theater and both players' Reserves, the Regent itself and a face-down Scratch Company. No structure or Headquarters is offered.
- `TestCard_IcewindRegent_MayReturnItself` (`FR-ACT-001`, `FR-ZON-004`): choosing itself (after its exhaust cost) returns it to its owner's hand. With no other unit on the battlefield the ability is still offered.
- `TestCard_IcewindRegent_ReturnsControlledUnitToItsOwner` (`FR-ZON-003`): a unit player 0 controls but player 1 owns returns to player 1's hand.
- `TestCard_IcewindRegent_ReturnsScratchCompanyToOwnersHand` (`FR-ZON-005`): an opposing face-down Scratch Company returned goes to its owner's hand as its card.
- `TestCard_IcewindRegent_WardedOpposingUnitNotOffered` (`FR-KWD-002h`): an opposing unit with Ward is not an option; an allied unit with Ward is.
- `TestCard_IcewindRegent_NotOfferedWhileExhausted` (`FR-ACT-002`): an Exhausted Regent isn't offered, and a rejected activation moves nothing. (`actNotOffered` asserts the asset is not Exhausted, so it doesn't fit here: use `offersActivate` and `s.reject`.)
- **Existing tests:** every other test passes unchanged, including `TestCards_Executable` with the name removed.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected.

## Allowed Files
- `pkg/engine/cards/concordia.go` (the definition)
- `pkg/engine/cards/cards.go` (remove the name from `NotYetExecutable`)

## Done When
- [ ] The card tests pass: `go test ./pkg/engine -run TestCard_IcewindRegent`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-matrix` exits 0 (just over two minutes: use a timeout above the two-minute default)
- [ ] The card loads with no unsupported lines (`TestCards_Executable` passes with the name removed)

## Log
- 2026-11-25 lead-architect (wave A run, session 2): written against 762f2ded, with the lead's probe above. Board read twice (D2.0): no directive is due here (#185 (3) and (4) are *Will Breaker*'s and *Tempest City Ray*'s), and no waiting trigger has fired. Every part of the Definition exists, and no other card prints an unrestricted exhaust-to-return. Set to `tests`.
