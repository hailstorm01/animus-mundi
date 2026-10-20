---
id: T-0013
title: Make Portent Scrivener and Despoiler Worm executable
kind: card
tier: flash
wave: A1 (board directive #4, task 4: card task 3 of 4)
status: queued
branch: m2/T-0013-scrivener-worm
depends_on: [T-0012]
claims: [card:Portent Scrivener, card:Despoiler Worm]
cards: [Portent Scrivener, Despoiler Worm]
requirements: []
---

## Goal
Two Animus units with draw death triggers now run in the kernel. One is gated to its controller's turn, and the other makes the opponent draw. Each is plain `ability` data added to `animusDefs` in `pkg/engine/cards/animus.go`.

## Inputs
- Cards (`data/cards/units-and-structures-animus.csv`). Each has exactly one unsupported line today, quoted here:
  - *Portent Scrivener*: Unit, 1R 0T, Reserves, 2/1. "When this unit is killed on your turn, draw a card."
  - *Despoiler Worm*: Unit, 8R 0T, Reserves, 8/9, keyword Resistant 2 (already supported). "When this unit is killed, all opponents draw 2 cards."
- Survey rows: `docs/design/m2_starter_set_ability_survey.md` (*Portent Scrivener* `turnOwner: SELF`; *Despoiler Worm* "Can deck the opponent").
- Shapes they use (all `done`): `HookKilled` (T-0008), `Trigger.Turn = TurnYours` and `SubjectOpponents` (T-0010), `OpDraw` with `Value{N}`.
- Requirements for context: `FR-ZON-004` (death triggers use the last-known controller), `FR-WIN-001` (a required draw from an empty deck loses at the next sweep), `FR-TRG-003` (cross-player APNAP when both players' units die).
- Rulings that apply: none.

### Definitions (the engineer adds them to `animusDefs`)
- *Portent Scrivener*: one `HookKilled` trigger with `Turn: TurnYours`, `Draw 1`.
- *Despoiler Worm*: one `HookKilled` trigger, `Draw 2` with `Subject: SubjectOpponents`.
- Each `Text` is the quoted line exactly. No new types, fields or logic. If a card seems to need one, set `blocked`.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
- (rules-test-author fills this in.)

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected.

## Allowed Files
- `pkg/engine/cards/animus.go`
- `pkg/engine/cards/cards.go` (remove the two names from `NotYetExecutable`)

## Done When
- [ ] The cards' tests pass: `go test ./pkg/engine -run 'TestCard_(PortentScrivener|DespoilerWorm)'`
- [ ] `TestDefs_MatchRulesText` and `TestStarterSet_Executable` pass unchanged
- [ ] `go test ./...`, `go vet ./...` and `gofmt -l pkg cmd` are clean
- [ ] `go run ./cmd/cardtown lint --all` passes
- [ ] `go run ./cmd/cardtown sim -deck starter -p0 ANIMUS -p1 PATH` exits 0, and the Animus executable count rises by 5 (4 Scrivener, 1 Worm)
- [ ] `go run ./cmd/cardtown sim` exits 0. Report the before and after summary lines in the Log.

## Log
- 2026-10-20 lead-architect: Wrote the task (A1 card task 3 of 4). It stays `queued` until T-0012 is `done`, because they share `animus.go` and `cards.go`.
