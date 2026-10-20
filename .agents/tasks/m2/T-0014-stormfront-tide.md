---
id: T-0014
title: Make Stormfront of Ideals and Unceasing Tide executable
kind: card
tier: flash
wave: A1 (board directive #4, task 4: card task 4 of 4)
status: tests
branch: m2/T-0014-stormfront-tide
depends_on: [T-0013]
claims: [card:Stormfront of Ideals, card:Unceasing Tide]
cards: [Stormfront of Ideals, Unceasing Tide]
requirements: []
---

## Goal
Two Animus units whose death triggers move the card itself now run in the kernel. The trigger tracks that exact card in its owner's discard pile and does nothing if the card has left. Each is plain `ability` data added to `animusDefs` in `pkg/engine/cards/animus.go`. With this task, every A1 card is executable.

## Inputs
- Cards (`data/cards/units-and-structures-animus.csv`). Each has exactly one unsupported line today, quoted here:
  - *Stormfront of Ideals*: Unit, 3R 0T, Reserves, 6/5. "When this unit is killed, put it on the bottom of its owner's deck."
  - *Unceasing Tide*: Unit, 3R 0T, Reserves, 3/2. "When this unit is killed, return it to its owner's hand."
- Survey rows: `docs/design/m2_starter_set_ability_survey.md` (both "Self-tracking fizzle").
- Shapes they use (all `done`): `HookKilled` (T-0008), `SubjectSelf`, `OpPutOnDeck` with `Bottom: true`, and `OpReturnToHand` (T-0009).
- Requirements for context: `FR-ZON-004`, the Self-Relocating Trigger Tracking bullet. FRD 01 names both cards.
- Rulings that apply: none.

### Definitions (the engineer adds them to `animusDefs`)
- *Stormfront of Ideals*: one `HookKilled` trigger, `{Op: OpPutOnDeck, Subject: SubjectSelf, Bottom: true}`.
- *Unceasing Tide*: one `HookKilled` trigger, `{Op: OpReturnToHand, Subject: SubjectSelf}`.
- Each `Text` is the quoted line exactly. No new types, fields or logic. If a card seems to need one, set `blocked`.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
- (rules-test-author fills this in.)

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected.

## Allowed Files
- `pkg/engine/cards/animus.go`
- `pkg/engine/cards/cards.go` (remove the two names from `NotYetExecutable`)

## Done When
- [ ] The cards' tests pass: `go test ./pkg/engine -run 'TestCard_(StormfrontOfIdeals|UnceasingTide)'`
- [ ] `TestDefs_MatchRulesText` and `TestStarterSet_Executable` pass unchanged
- [ ] `go test ./...`, `go vet ./...` and `gofmt -l pkg cmd` are clean
- [ ] `go run ./cmd/cardtown lint --all` passes
- [ ] `go run ./cmd/cardtown sim -deck starter -p0 ANIMUS -p1 PATH` exits 0, and the Animus executable count rises by 5 (2 Stormfront, 3 Tide)
- [ ] `go run ./cmd/cardtown sim` exits 0. Report the before and after summary lines in the Log.

## Log
- 2026-10-20 lead-architect: Wrote the task (A1 card task 4 of 4). It stays `queued` until T-0013 is `done`, because they share `animus.go` and `cards.go`.
- 2026-10-20 lead-architect: T-0013 is `done`, so status `tests` for `rules-test-author`.
