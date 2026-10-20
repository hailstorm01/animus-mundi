---
id: T-0012
title: Make Despoiler Larva, Fungal Evangelist and Titanic Spore Carrier executable
kind: card
tier: flash
wave: A1 (board directive #4, task 4: card task 2 of 4)
status: in-progress
branch: m2/T-0012-scratch-draw-cards
depends_on: [T-0011]
claims: [card:Despoiler Larva, card:Fungal Evangelist, card:Titanic Spore Carrier]
cards: [Despoiler Larva, Fungal Evangelist, Titanic Spore Carrier]
requirements: []
---

## Goal
Three Animus units whose printed triggers make their controller Scratch or draw now run in the kernel. Each is plain `ability` data in `pkg/engine/cards/animus.go`, added to the `animusDefs` map that T-0011 created.

## Inputs
- Cards (`data/cards/units-and-structures-animus.csv`). Each has exactly one unsupported line today, quoted here:
  - *Despoiler Larva*: Unit, 1R 0T, Reserves, 1/1. "Whenever this unit attacks, Scratch 1."
  - *Fungal Evangelist*: Unit, 6R 0T, Reserves, 2/3. "When this unit deploys or is killed, Scratch 2."
  - *Titanic Spore Carrier*: Unit, 8R 0T, Reserves, 7/7. "When this unit is killed, Scratch 1 and draw a card."
- Survey rows: `docs/design/m2_starter_set_ability_survey.md` (*Despoiler Larva*, *Fungal Evangelist* "Compound hook", *Titanic Spore Carrier* "Composite sequence").
- Shapes they use (all `done`): `HookAttacks` (T-0010, Step 2 with Aggressive), `HookDeploy` (T-0007), `HookKilled` (T-0008), `OpScratch` and `OpDraw` with the zero `Subject` (controller), and `Value{N}`.
- Requirements for context: `FR-CBT-002` Step 2 (attack triggers before damage), `FR-KWD-002m` (Scratch N), `FR-ZON-005`, `FR-ZON-004` (death triggers; a face-down copy never triggers), `FR-TRG-003` (one player's simultaneous triggers are ordered), `FR-WIN-001`.
- Rulings that apply: none.

### Definitions (the engineer adds them to `animusDefs` in `pkg/engine/cards/animus.go`)
- *Despoiler Larva*: one `HookAttacks` trigger, `Scratch 1`.
- *Fungal Evangelist*: one `Text` line and **two** triggers, `HookDeploy` and `HookKilled`, each `Scratch 2`. "deploys or is killed" is two events.
- *Titanic Spore Carrier*: one `HookKilled` trigger with two steps in printed order, `Scratch 1` then `Draw 1`.
- Each `Text` is the quoted line exactly. No new types, fields or logic. If a card seems to need one, set `blocked`.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
All in `pkg/engine/cards_animus_test.go`, loaded with `s.card`. Controls (marked) pass before the definitions exist because nothing fires; the rest fail now.
- Despoiler Larva: `AttackScratchesOne`, `ScratchResolvesBeforeCombatDamage` (lethal on a 19-damage HQ; the Scratch must still happen, FR-CBT-002 Step 2), `ScratchOnEmptyDeckDoesNothing` (control), `FaceDownCopyAttackingDoesNotScratch` (control, FR-ZON-005), `DefendingDoesNotScratch` (control).
- Fungal Evangelist: `DeployScratchesTwo`, `DeployWithOneCardLeftScratchesItAndDoesNotLose`, `DeployOnEmptyDeckDoesNothing`, `DeathScratchesTwoOnItsControllersTurn`, `DeathOnOpponentsTurnScratchesForItsController`, `DeathOnEmptyDeckDoesNothing` (control), `FaceDownCopyDoesNotScratchWhenKilled` (control, FR-ZON-004).
- Titanic Spore Carrier: `DeathScratchesFirstCardThenDrawsSecond`, `DrawFromEmptyDeckLosesAtThatSweep`, `ScratchTakesTheLastCardSoOnlyTheDrawLoses`, `FaceDownCopyDoesNothingWhenKilled` (control).

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected.

## Allowed Files
- `pkg/engine/cards/animus.go` (three entries in `animusDefs`)
- `pkg/engine/cards/cards.go` (remove the three names from `NotYetExecutable`)

## Done When
- [ ] The cards' tests pass: `go test ./pkg/engine -run 'TestCard_(DespoilerLarva|FungalEvangelist|TitanicSporeCarrier)'`
- [ ] `TestDefs_MatchRulesText` and `TestStarterSet_Executable` pass unchanged
- [ ] `go test ./...`, `go vet ./...` and `gofmt -l pkg cmd` are clean
- [ ] `go run ./cmd/cardtown lint --all` passes
- [ ] `go run ./cmd/cardtown sim -deck starter -p0 ANIMUS -p1 PATH` exits 0, and the Animus executable count rises by the copies in the deck. Of these three only *Despoiler Larva* is in it (4 copies), so expect 5/50 → 9/50.
- [ ] `go run ./cmd/cardtown sim` exits 0. Report the before and after summary lines in the Log.

## Log
- 2026-10-20 lead-architect: Wrote the task (A1 card task 2 of 4, grouped as controller Scratch/Draw triggers). Set `status: tests` for `rules-test-author`. After test review it goes to `ready` for `@kernel-engineer T-0012` (tier flash, Antigravity).
- 2026-10-20 rules-test-author: Wrote 16 TestCard_* tests (listed under Tests) from the CSV text. 10 fail now; 6 are controls that pass until a definition fires wrongly. No helpers requested. Left out: a two-attack Larva test (the faceDownCount==1 check pins once per attack) and Evangelist trigger-ordering tests (its two triggers are never simultaneous).
- 2026-10-20 lead-architect: Test review: **approved**, status `ready`. 16 tests were appended to `cards_animus_test.go`, with no line of an existing test changed. They are written from the CSV text with the real cards.
  - The Larva-versus-HQ case can only see the Scratch if Step 2 resolves before damage (FR-CBT-002).
  - `killOnOpponentsTurn` arranges the deck before the turn passes, so the opponent's draw doesn't disturb it.
  - The Carrier's Scratch-then-Draw order shows on named cards, and the one-card case separates a safe Scratch from a losing Draw (FR-KWD-002m, FR-WIN-001).
  - The six controls pass today, and they guard against definitions that fire on face-down copies, on defense or on an empty deck.
  - Next: **the user runs `@kernel-engineer T-0012` in the Antigravity agent (Gemini Flash)**.
