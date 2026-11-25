---
id: T-0300
title: Make Icewind Regent executable (exhaust to return a chosen unit, wave A, card, flash)
kind: card
tier: flash
wave: A
status: done
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
- [x] The card tests pass: `go test ./pkg/engine -run TestCard_IcewindRegent`
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks
- [x] `go tool task sim-matrix` exits 0 (just over two minutes: use a timeout above the two-minute default)
- [x] The card loads with no unsupported lines (`TestCards_Executable` passes with the name removed)

## Log
- 2026-11-25 lead-architect (wave A run, session 2): written against 762f2ded, with the lead's probe above. Board read twice (D2.0): no directive is due here (#185 (3) and (4) are *Will Breaker*'s and *Tempest City Ray*'s), and no waiting trigger has fired. Every part of the Definition exists, and no other card prints an unrestricted exhaust-to-return. Set to `tests`.
- 2026-11-25 rules-test-author: wrote `pkg/engine/cards_concordia_regent_test.go` with the seven listed tests: `TestCard_IcewindRegent_ReturnsChosenOpposingUnit`, `_AnyUnitInAnyZoneOffered`, `_MayReturnItself`, `_ReturnsControlledUnitToItsOwner`, `_ReturnsScratchCompanyToOwnersHand`, `_WardedOpposingUnitNotOffered`, `_NotOfferedWhileExhausted`. A file-local helper `regentChoose` picks an option by asset ID (option order is not specified). Helpers requested: none. Nothing left out. `go vet` is clean and `cardtown lint rules` passes. Six tests fail because the card has no activated ability yet; `NotOfferedWhileExhausted` passes now, which is expected before implementation. Set to `tests-review`.
- 2026-11-25 lead-architect: tests approved (D4, round 1), each checked against `FR-ACT-001`, `FR-ACT-002`, `FR-TGT-001`, `FR-ZON-003` to `005` and `FR-KWD-002h`. A lead probe with the Definition, since reverted, passes all seven. Mutants without the exhaust cost, with an opposing-only, asset-wide or Main Theater-only filter each fail at least two. Set to `ready`.
- 2026-11-25 kernel-engineer: added Icewind Regent definition to concordiaDefs in pkg/engine/cards/concordia.go with Activated ability (CostExhaustSelf cost, Effect with one unit target filter and OpReturnToHand step targeting target 0). Removed Icewind Regent from NotYetExecutable in pkg/engine/cards/cards.go. Reused CostExhaustSelf (as in Cliff of Portents) and OpReturnToHand with SubjectTarget (as in Ambush Tactics mode 2). Commands ran: 'go test ./pkg/engine -run TestCard_IcewindRegent' passed (1.104s), 'go tool task check' passed (gofmt, go vet, all unit/package tests, all cardtown linters clean), 'go tool task sim' passed (200 matches with all invariants holding), and 'go tool task sim-matrix' passed (all 24 starter deck matches and 24 stress deck matches clean, exit 0). Set to review.
- 2026-11-25 lead-architect (wave A run, session 2): approved (D5, round 1), merged as e72d89d5. The diff is the task's Definition and the one name off `NotYetExecutable`, inside `Allowed Files`, and the test file changed only in 08c3f1e7. No new shape, no card name in engine code. Re-run by the reviewer on the task branch, with exit codes: `go test ./pkg/engine -run TestCard_IcewindRegent`, `go tool task check`, `go tool task sim`, `go tool task sim-matrix`, all 0. The merge has the same tree as the reviewed branch. The Animus vs Path sim is unchanged (P0 58.5% / P1 39.5% / draws 2.0%). The Flash run left no denied action.
