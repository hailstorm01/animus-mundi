---
id: T-0302
title: Make Seismic Instability executable (two untargeted Main Theater damage steps, wave A, card, flash)
kind: card
tier: flash
wave: A
status: ready
branch: m3/autoloop-wave-a/T-0302-seismic-instability
depends_on: [T-0301]
claims: [card:Seismic Instability]
cards: [Seismic Instability]
requirements: [FR-MNV-001, FR-TGT-002, FR-REP-004, FR-LAY-002]
---

## Goal
*Seismic Instability* (a Concordia Maneuver, R0 T4; 2 copies in the starter deck): *"Deal 3 damage to all assets in the Main Theater."* / *"Deal 4 damage to all structures in the Main Theater."* It only combines shapes that are `done`: two `OpDealDamage` steps with `SubjectAll` and a zone filter (*Despoiler's Emergence*).

Covers:
- The definition in `pkg/engine/cards/concordia.go`, as under Definition below.
- The name off `cards.NotYetExecutable`.

## Inputs
- Card (`data/cards/maneuvers-concordia.csv` lines 12-13).
- Survey row: `docs/design/survey/m3_full_catalog_ability_survey.md` §6, "Concordia: deck" (*Seismic Instability*: "Structures take two instances").
- FRD 01: `FR-MNV-001` (lines 393-396), `FR-TGT-002` (lines 426-427), `FR-LAY-002` (line 536, State-Based Death Sweep Timing), `FR-REP-004` (lines 562-588), `FR-KWD-002a` (line 639), `FR-KWD-002b` (line 642), `FR-KWD-002h` (line 648), `FR-KWD-002k` (line 651).
- Rulings: `SVY-R29` (a Maneuver with no target slot is always a legal Enact), `SVY-R3` ("all" is not a target).
- Design note: `docs/architecture/m3-full-catalog.md`, Boundaries (`pkg/engine/cards`), Deliverable 5.

### Definition (the engineer adds it to `concordiaDefs`)
- `"Seismic Instability"`, `Text`: `"Deal 3 damage to all assets in the Main Theater."`, `"Deal 4 damage to all structures in the Main Theater."` (two lines).
- `Effect`: no targets. `Steps` are `{Op: ability.OpDealDamage, Amount: ability.Value{N: 3}, Subject: ability.SubjectAll, Filter: ability.Filter{Zone: ability.ZoneMainTheater}}`, then `{Op: ability.OpDealDamage, Amount: ability.Value{N: 4}, Subject: ability.SubjectAll, Filter: ability.Filter{Kinds: ability.KindStructure, Zone: ability.ZoneMainTheater}}`.

A lead probe with the definition on the run branch at cf2d76d1, since reverted, passed `go tool task check`.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
In a new file `pkg/engine/cards_concordia_seismic_test.go`, from the CSV text, loading the real card (`s.card`), in the style of the *Despoiler's Emergence* tests (`cards_animus_maneuvers_test.go`). Player 0 enacts; player 1 is the opponent. Reuse the existing test helpers (`enactAction`, `enactFor`, `offersEnact`, `damageEventsOn`, `damageEvents`, `inDiscardPile`, `lastDiscard`, `s.moveTo`, `s.putScratch`, `s.wantDamage`). The other cards may be synthetic.
- `TestCard_SeismicInstability_ThreeToEveryUnitInTheMainTheater` (`FR-MNV-001`): units of both players in the Main Theater take 3 each, in one damage event each.
- `TestCard_SeismicInstability_StructuresTakeBothInstances` (`FR-REP-004`): structures of both players in the Main Theater take 3 and then 4, two damage events, 7 in total.
- `TestCard_SeismicInstability_ReservesAndHeadquartersUntouched` (`FR-TGT-002`): units, structures and a Scratch Company in Reserves, and both Headquarters, take nothing and have no damage event.
- `TestCard_SeismicInstability_ScratchCompanyInTheMainTheaterIsAUnit` (`FR-ZON-005`): a face-down Scratch Company in the Main Theater whose hidden card is a structure takes one instance of damage only, and dies.
- `TestCard_SeismicInstability_ResistantAppliesToEachInstance` (`FR-KWD-002b`): a structure with Resistant 1 takes 2 and 3, 5 in total. A unit with Resistant -1 takes 4.
- `TestCard_SeismicInstability_ArmorAbsorbsAcrossBothInstances` (`FR-REP-004`, `FR-KWD-002k`): a structure with Armor 5 takes 2 damage in total and has no Armor left.
- `TestCard_SeismicInstability_LethalStructureStillTakesTheSecondInstance` (`FR-LAY-002`): a structure with 3 Health is not destroyed between the two lines: both damage events are logged on it, and it dies after the Maneuver finishes. A unit with 3 Health dies and one with 4 survives. Dead assets go to their owners' discard piles, and the Maneuver to its own.
- `TestCard_SeismicInstability_UnbowedStructureTriggersTwice` (`FR-KWD-002a`): a synthetic surviving structure with an Unbowed trigger triggers once for each of the two instances, and once only for a unit.
- `TestCard_SeismicInstability_AsksNothing` (`FR-TGT-002`, `FR-KWD-002h`): with an opposing Ward asset in the Main Theater and scripts with no answers for both players, nobody is asked anything, and the Ward asset takes its damage.
- `TestCard_SeismicInstability_EnactableWithAnEmptyMainTheater` (`SVY-R29`): it is offered with no asset in the Main Theater, damages nothing, and goes to the discard pile.
- `TestCard_SeismicInstability_PaysFourTactics` (`FR-MNV-001`): from R3 T5 the pool is R3 T1 afterwards. With R5 T3 it is not offered.
- **Existing tests:** every other test passes unchanged, including `TestCards_Executable` with the name removed.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected.

## Allowed Files
- `pkg/engine/cards/concordia.go` (the definition)
- `pkg/engine/cards/cards.go` (remove the name from `NotYetExecutable`)

## Done When
- [ ] The card tests pass: `go test ./pkg/engine -run TestCard_SeismicInstability`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-matrix` exits 0 (just over two minutes: use a timeout above the two-minute default)
- [ ] The card loads with no unsupported lines (`TestCards_Executable` passes with the name removed)

## Log
- 2026-11-25 lead-architect (wave A run, session 4): written against cf2d76d1, with the lead's probe above. No directive is due here (#185 (3) and (4) are *Will Breaker*'s and *Tempest City Ray*'s), and no waiting trigger has fired. Every part of the Definition exists. Set to `tests`.
- 2026-11-25 rules-test-author: wrote the eleven tests in `pkg/engine/cards_concordia_seismic_test.go` as listed under Tests: ThreeToEveryUnitInTheMainTheater, StructuresTakeBothInstances, ReservesAndHeadquartersUntouched, ScratchCompanyInTheMainTheaterIsAUnit, ResistantAppliesToEachInstance, ArmorAbsorbsAcrossBothInstances, LethalStructureStillTakesTheSecondInstance, UnbowedStructureTriggersTwice, AsksNothing, EnactableWithAnEmptyMainTheater, PaysFourTactics. `go vet ./pkg/engine` is clean and `cardtown lint rules` passes. All eleven fail now (the card is not yet executable), as intended. Helpers requested: none. Cases left out: none. The Scratch test replaces the face-down asset's hidden card with a structure and moves it to the Main Theater, and it asserts one damage event and death.
- 2026-11-25 lead-architect: tests round 1 returned (D4). Each test agrees with the FRD 01 text it cites, and the lead's probe passes all 11. Of 13 mutants of the Definition, 12 fail at least one test. Set to `tests` for one note:
  1. `TestCard_SeismicInstability_StructuresTakeBothInstances` doesn't check the order the task names ("take 3 and then 4"): swapping the two steps passes every test. Also check that each structure's two damage events, in log order, carry the raw damage 3 and then 4 (`Event.Raw`). Change nothing else.
- 2026-11-25 rules-test-author: applied note 1. `TestCard_SeismicInstability_StructuresTakeBothInstances` now also collects each structure's damage events in log order and checks their `Event.Raw` is 3 then 4, so swapping the two steps fails it. Nothing else changed. `go vet ./pkg/engine` is clean. Helpers requested: none. Set to `tests-review`.
- 2026-11-25 lead-architect: tests approved (D4, round 2). Note 1 is applied and nothing else changed. A lead probe with the Definition, since reverted, passes all 11, and all 13 mutants (both amounts, swapped amounts, step order, zones, kinds, allegiances, Scratch excluded, kill instead of damage) fail at least one. The committed test file is LF and gofmt-clean. Set to `ready`.
