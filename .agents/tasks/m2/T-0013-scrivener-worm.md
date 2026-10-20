---
id: T-0013
title: Make Portent Scrivener and Despoiler Worm executable
kind: card
tier: flash
wave: A1 (board directive #4, task 4: card task 3 of 4)
status: done
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
All in `pkg/engine/cards_animus_test.go`, loading the real cards with `s.card(name)`:
- `TestCard_PortentScrivener_KilledOnYourTurnDrawsOne`: attacks and dies on its controller's turn; controller draws exactly the top card.
- `TestCard_PortentScrivener_KilledOnOpponentsTurnDrawsNothing`: no draw on the opponent's turn.
- `TestCard_PortentScrivener_DrawFromEmptyDeckLosesAtThatSweep` (FR-WIN-001).
- `TestCard_PortentScrivener_FaceDownCopyDoesNotDraw` (FR-ZON-004).
- `TestCard_DespoilerWorm_KilledMakesOpponentDrawTwo`: exactly the top two named cards.
- `TestCard_DespoilerWorm_ControllerDrawsNothing`.
- `TestCard_DespoilerWorm_TriggersOnItsControllersTurnToo`.
- `TestCard_DespoilerWorm_OpponentWithOneCardInDeckLosesAtThatSweep` and `..._OpponentWithEmptyDeckLosesAtThatSweep` (FR-WIN-001).
- `TestCard_DespoilerWorm_ResistantTwoReducesDamage`: 5 damage becomes 3.
- `TestCard_DespoilerWorm_FaceDownCopyDoesNotMakeOpponentDraw` (FR-ZON-004).
- Local helpers added in the same file: `scrivenerKilledOnOwnTurn`, `wormKilled`.
- Left out: the APNAP case (FR-TRG-003) with a Worm and a Scrivener dying in one combat. Resistant 2 means the Worm cannot be killed by the Scrivener's 2 power, and the Scrivener only triggers on its own controller's turn, so no clean two-sided setup exists through state.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected.

## Allowed Files
- `pkg/engine/cards/animus.go`
- `pkg/engine/cards/cards.go` (remove the two names from `NotYetExecutable`)

## Done When
- [x] The cards' tests pass: `go test ./pkg/engine -run 'TestCard_(PortentScrivener|DespoilerWorm)'`
- [x] `TestDefs_MatchRulesText` and `TestStarterSet_Executable` pass unchanged
- [x] `go test ./...`, `go vet ./...` and `gofmt -l pkg cmd` are clean
- [x] `go run ./cmd/cardtown lint --all` passes
- [x] `go run ./cmd/cardtown sim -deck starter -p0 ANIMUS -p1 PATH` exits 0, and the Animus executable count rises by 5 (4 Scrivener, 1 Worm)
- [x] `go run ./cmd/cardtown sim` exits 0. Report the before and after summary lines in the Log.

## Log
- 2026-10-20 lead-architect: Wrote the task (A1 card task 3 of 4). It stays `queued` until T-0012 is `done`, because they share `animus.go` and `cards.go`.
- 2026-10-20 lead-architect: T-0012 is `done`, so status `tests` for `rules-test-author`.
- 2026-10-20 rules-test-author: Wrote 11 tests (4 Scrivener, 7 Worm) in `cards_animus_test.go`; see Tests. No helpers requested (two local helpers added in the test file). Left out the APNAP case: the Worm's Resistant 2 stops a Scrivener from killing it, and the Scrivener triggers only on its controller's turn, so no clean state-observable setup exists. Status `tests-review`.
- 2026-10-20 lead-architect: Test review: **approved**, status `ready`. The 11 tests landed in `test(T-0013)` 4a78d82, which `rules-test-author` committed itself after a classifier outage blocked its Bash last session. They are appended to `cards_animus_test.go`, with no line of an existing test changed, and they use the real cards.
  - The Scrivener's "on your turn" condition is checked both ways: it draws when killed attacking on its controller's turn, and draws nothing through `killOnOpponentsTurn`.
  - `wormKilled` replaces the opponent's deck after that player's turn-start draw, so the two-card count is exact. The 20-power hitter still kills the 8/9 Worm through Resistant 2.
  - The one-card and empty-deck cases separate a successful first draw from the losing second draw (FR-WIN-001). The face-down cases follow FR-ZON-004's last-known characteristics.
  - Leaving out the APNAP case (FR-TRG-003) is accepted, for the reason given above.
  - Re-run: 6 fail (cards undefined) and 5 controls pass.
  - Next: **the user runs `@kernel-engineer T-0013` in the Antigravity agent (Gemini Flash)**.
- 2026-10-20 kernel-engineer: Implemented. animus.go: added definitions to animusDefs for "Portent Scrivener" (HookKilled, Turn: TurnYours, OpDraw 1) and "Despoiler Worm" (HookKilled, OpDraw 2 with Subject: SubjectOpponents) in alphabetical order. cards.go: removed "Despoiler Worm" and "Portent Scrivener" from NotYetExecutable allowlist while preserving sort order. Reused HookKilled, TurnYours, OpDraw, SubjectOpponents, Value{N}; no new shapes or engine logic added. Gates: go test ./pkg/engine -run 'TestCard_(PortentScrivener|DespoilerWorm)' passed (11/11), TestDefs_MatchRulesText and TestStarterSet_Executable passed, go test ./... ok, go vet ./... clean, gofmt -l pkg cmd clean, cardtown lint --all passed. Starter deck sim reported P0 (Animus starter) main executable cards rose from 9/50 to 14/50 (4 Scrivener, 1 Worm). Default sim before and after: "SOULLESS (Colossus Platform) vs CONCORDIA (Enclave), P0 wins 66.0% P1 wins 30.0% draws 4.0% first player wins 47.0%, Avg match length: 30.2 turns" (unchanged as default sim runs Soulless vs Concordia). Reviewer: check that Despoiler Worm uses SubjectOpponents for OpDraw 2, and Portent Scrivener gates trigger execution with Turn: TurnYours.
- 2026-10-20 lead-architect: Code review: **approved**, status `done`.
  - Scope: f3209ef changes only `animus.go` and `cards.go`. No test changed after `test(T-0013)` 4a78d82.
  - Fences: plain `ability` data with no new shapes. Both use `HookKilled` and `OpDraw`. The Scrivener adds `Turn: TurnYours`, and the Worm adds `Subject: SubjectOpponents`, matching how T-0010 tests use that shape. Each `Text` is the CSV line exactly. The fences lint passes.
  - Gates re-run by the reviewer: 11/11 card tests, `TestDefs_MatchRulesText` and `TestStarterSet_Executable` pass. `go test ./...`, `go vet ./...` and `gofmt -l pkg cmd` are clean, and `cardtown lint --all` passes.
  - Starter sim exits 0. Animus executable count goes from 9/50 to 14/50 (+5). Result: P0 49.5%, P1 16.5%, draws 34.0%, all invariants hold. The default sim also exits 0.
