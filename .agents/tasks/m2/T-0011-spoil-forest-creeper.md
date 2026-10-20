---
id: T-0011
title: Make Spoil Forest Creeper executable
kind: card
tier: flash
wave: A1 (board directive #4, task 4: first card task)
status: tests
branch: m2/T-0011-spoil-forest-creeper
depends_on: [T-0010]
claims: [card:Spoil Forest Creeper, file:pkg/engine/cards/animus.go]
cards: [Spoil Forest Creeper]
requirements: []
---

## Goal
*Spoil Forest Creeper* runs in the kernel: "When this unit deploys, Scratch 1." It is the first card task, so it also proves the card pipeline end to end. That covers the `Def` in `pkg/engine/cards`, its exact `Text` line attached by `LoadCards`, the card leaving `cards.NotYetExecutable`, and a `TestCard_*` test played through the real card from `data/cards`.

## Inputs
- Card: `data/cards/units-and-structures-animus.csv`, *Spoil Forest Creeper*: Unit, Animus, Treefolk/Horror, cost 3R 0T, Reserves, rules text "When this unit deploys, Scratch 1.", 3 power, 2 health.
- Survey row: `docs/design/m2_starter_set_ability_survey.md` (*Spoil Forest Creeper*: deploy → `SCRATCH 1`, "Simplest first card").
- Shapes it uses (all `done`): `ability.HookDeploy` (T-0007), `ability.OpScratch` with `Value{N: 1}` and the zero `Subject` (controller) (T-0007, T-0009).
- Requirements for context: `FR-KWD-002m` (Scratch N), `FR-ZON-005` (a Scratch Company enters face-down and Exhausted), `FR-TRG-004` (the trigger resolves after the deploy).
- Design note: `docs/architecture/m2-ability-model.md`, Boundaries (`pkg/engine/cards`: one `ability.Def` per card in `animus.go`; `cards.go` holds `Defs` and `NotYetExecutable`) and Test Plan (card tests are written from the CSV text, never from `cards/*.go`).
- Rulings that apply: none.

### Definition (the engineer writes it in `pkg/engine/cards/animus.go`)
- `Text: []string{"When this unit deploys, Scratch 1."}`. The line must equal the CSV line exactly; `LoadCards` rejects anything else.
- One `Trigger{Hook: HookDeploy, Effect: Effect{Steps: []Step{{Op: OpScratch, Amount: Value{N: 1}}}}}`.
- `animus.go` adds its entries to `Defs` (an `init` func, or a map literal that `cards.go` merges). Choose the form later Animus card tasks can append to. No logic, only `ability` values.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
- (rules-test-author fills this in.)

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected. `s.card(name)` loads the real card; `s.give`, `s.do`, `s.onTop`, `s.script` and `s.asked` exist.

## Allowed Files
- `pkg/engine/cards/animus.go` (new: the *Spoil Forest Creeper* definition)
- `pkg/engine/cards/cards.go` (remove "Spoil Forest Creeper" from `NotYetExecutable`; merge `animus.go`'s entries into `Defs` if the chosen form needs it)

## Done When
- [ ] The card's tests pass: `go test ./pkg/engine -run 'TestCard_SpoilForestCreeper'`
- [ ] `TestDefs_MatchRulesText` and `TestStarterSet_Executable` pass unchanged (the card loads with no unsupported lines)
- [ ] `go test ./...`, `go vet ./...` and `gofmt -l pkg cmd` are clean
- [ ] `go run ./cmd/cardtown lint --all` passes (fences: the card name appears only under `pkg/engine/cards`)
- [ ] `go run ./cmd/cardtown sim -deck starter -p0 ANIMUS -p1 PATH` exits 0, and its header reports more executable Animus main-deck cards than before (the deck holds 4 copies)
- [ ] `go run ./cmd/cardtown sim` exits 0. Its output may change now that the card is playable; report the before and after summary lines in the Log.

## Log
- 2026-10-20 lead-architect: Wrote the first A1 card task. It has one card so the card pipeline is proven before any batching. Set `status: tests` for `rules-test-author`. After the tests are reviewed, the task goes to `ready` for `@kernel-engineer T-0011` (tier flash, Antigravity). Note for the engineer: the definition is plain `ability` data, with no new types, fields or logic. If something seems to need one, set `blocked`.
