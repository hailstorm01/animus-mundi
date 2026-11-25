---
id: T-0299
title: Make Growing Legion executable (Scratch 1; first Concordia card, wave A, card, flash)
kind: card
tier: flash
wave: A
status: in-progress
branch: m3/autoloop-wave-a/T-0299-growing-legion
depends_on: [T-0295]
claims: [card:Growing Legion, file:pkg/engine/cards/concordia.go]
cards: [Growing Legion]
requirements: [FR-KWD-002m, FR-ZON-005, FR-MNV-001]
---

## Goal
*Growing Legion* (a Concordia Maneuver, 3 copies in the starter deck): *"Scratch 1."* It only combines shapes that are `done`: a Maneuver's `Effect` with one `OpScratch` step, as *Animoceps Spores*' first step.

Covers:
- `pkg/engine/cards/concordia.go`, a new file with the definition, as under Definition below.
- The name off `cards.NotYetExecutable`.
- Directive #185 (2): `scratchFace`'s doc comment cites `SVY-R59`. Comment only, no behavior change.

## Inputs
- Card (`data/cards/maneuvers-concordia.csv` line 2): Concordia Maneuver, R0 T1.
- Survey row: `docs/design/survey/m3_full_catalog_ability_survey.md` §6, "Concordia: deck" (*Growing Legion*).
- FRD 01: `FR-KWD-002m` (line 653), `FR-ZON-005` (line 288), `FR-MNV-001`.
- Rulings: `KIM-S01` and `KIM-S05` (Scratch Companies print Deploys Exhausted; a scratch is a unit deploying). `SVY-R59` (a face-down Scratch Company costs 0 and 0) for the comment only: its test is *Will Breaker*'s, #185 (3).
- Design note: `docs/architecture/m3-full-catalog.md`, Boundaries (`pkg/engine/cards`), Deliverable 5.

### Definition (the engineer adds it)
`concordia.go` has the same form as `animus.go`: `package cards`, the `ability` import, an `init` that copies `concordiaDefs` into `Defs`, and `concordiaDefs` with the comment `// concordiaDefs holds ability definitions for Concordia cards.` Its one entry:
- `"Growing Legion"`, `Text`: `"Scratch 1."`.
- `Effect`: one step, `{Op: ability.OpScratch, Amount: ability.Value{N: 1}}`.

`scratchFace`'s doc comment in `pkg/engine/card.go` becomes:
```go
// scratchFace is the face-down Scratch Company characteristic set (FR-ZON-005).
// Its Requisition and Tactics costs are 0 (SVY-R59).
```

A lead probe with the definition on the run branch at 6414ae6d, since reverted, passed `go tool task check` and `go tool task sim-matrix`.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
In a new file `pkg/engine/cards_concordia_maneuvers_test.go`, from the CSV text, loading the real card (`s.card`), in the style of `cards_animus_maneuvers_test.go`. Player 0 enacts with exactly R0 T1. Reuse the existing test helpers (`scratched`, `faceDownCount`, `enactAction`, `inDiscardPile`, `setPool`).
- `TestCard_GrowingLegion_ScratchesOne` (`FR-KWD-002m`, `FR-ZON-005`): the top card of player 0's deck is the only Scratch Company, face-down in player 0's Reserves and Exhausted. The second card is now on top, and the opponent has no Scratch Company. The T1 is paid and *Growing Legion* is in its controller's discard pile.
- `TestCard_GrowingLegion_DoesNotUseTheHeadquarters` (`FR-KWD-002m`): player 0's Headquarters, Readied before, is still Readied afterwards.
- `TestCard_GrowingLegion_EnactableWithAnExhaustedHeadquarters` (`FR-KWD-002m`): with player 0's Headquarters Exhausted, it still scratches one.
- `TestCard_GrowingLegion_EmptyDeckScratchesNothingAndDoesNotLose` (`FR-KWD-002m`): it can be enacted with a 0-card deck, makes no unit, and the match isn't over.
- **Existing tests:** every other test passes unchanged, including `TestCards_Executable` with the name removed.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected.

## Allowed Files
- `pkg/engine/cards/concordia.go` (new: the definition)
- `pkg/engine/cards/cards.go` (remove the name from `NotYetExecutable`)
- `pkg/engine/card.go` (`scratchFace`'s doc comment only)

## Done When
- [ ] The card tests pass: `go test ./pkg/engine -run TestCard_GrowingLegion`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-matrix` exits 0 (just over two minutes: use a timeout above the two-minute default)
- [ ] The card loads with no unsupported lines (`TestCards_Executable` passes with the name removed)

## Log
- 2026-11-25 lead-architect (wave A run, session 1): written against 6414ae6d, with the lead's probe above. Board read twice (D2.0). #185 is `open: before the first wave A card task`: its (2) is in this task. Triggers of #186 to #189 have fired (the outcome reads `Status: applied: FRD 01 v3.32`), so they are now `open: before` the wave B to E tasks they name. Every other waiting trigger was checked and none has fired. Searched `pkg/engine` and the task archives: every part of the Definition exists, and no other card prints just "Scratch 1." Set to `tests`.
- 2026-11-25 rules-test-author (wave A run): wrote `pkg/engine/cards_concordia_maneuvers_test.go` with `TestCard_GrowingLegion_ScratchesOne`, `_DoesNotUseTheHeadquarters`, `_EnactableWithAnExhaustedHeadquarters` and `_EmptyDeckScratchesNothingAndDoesNotLose`. No helpers requested (uses `onTop`, `setPool`, `scratched`, `faceDownCount`, `inDiscardPile`, `offersEnact`). `go vet` and `cardtown lint rules` pass; all four fail now with "needs rules the kernel does not implement yet", as expected. Nothing left out. Set to `tests-review`.
- 2026-11-25 lead-architect: tests approved (D4, round 1), each checked against `FR-KWD-002m`, `FR-ZON-005` and `FR-MNV-001`. A lead probe with the Definition above, since reverted, passes all four, and a Scratch 2 mutant fails `ScratchesOne`. Set to `ready`.
