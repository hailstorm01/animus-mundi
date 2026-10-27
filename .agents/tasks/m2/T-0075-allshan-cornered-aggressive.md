---
id: T-0075
title: Allshan reads "Cornered: Aggressive 3." (directive #60 part 2, card, flash)
kind: card
tier: flash
wave: B (board directive #60 (2), blocking before T-0073)
status: ready
branch: m2/T-0075-allshan-cornered-aggressive
depends_on: [T-0074]
claims: [card:Allshan, Chainbreaker Revenant]
cards: [Allshan, Chainbreaker Revenant]
requirements: [FR-KWD-002e, FR-KWD-002f]
---

## Goal
*Allshan, Chainbreaker Revenant*'s second line is printed as "Cornered: Aggressive 3." (the style of "Cornered: Resistant 1.") and its definition uses T-0074's `ability.Aggressive(3)` trigger with `Cornered` set, instead of a hand-written attack trigger with an `OpModifyStat` step. Behavior is unchanged: the T-0070 card tests pass unchanged.

## Inputs
- Board directive #60 (2) (user, 2026-10-26).
- Card (`data/cards/units-and-structures-path.csv`, line 26): Path unit, Specialist, cost 5, Reserves. Its second rules-text line "Cornered: Whenever this unit attacks, it gets +3 power until end of turn." becomes "Cornered: Aggressive 3." Nothing else in the row changes.
- Survey row: `docs/design/m2_starter_set_ability_survey.md` (Allshan, reworded on `main` 4214eaa). Ruling `R18`: Allshan has Aggressive only while Cornered. Nothing reads "has Aggressive" yet, so this task builds nothing for it.
- `FR-KWD-002e`: Aggressive is a triggered ability, so Cornered is checked when the attack happens (`Trigger.Cornered`, T-0069), as today.
- Shapes it uses: `ability.Aggressive` and `Trigger.Keyword` (T-0074), `Trigger.Cornered` (T-0069), `HookDamaged` + `OpAddArmor` (Unbowed, unchanged).

### Definition (the engineer changes it in `pkg/engine/cards/path.go`)
- `Text`: `"Unbowed: This unit gains 2 armor."`, then `"Cornered: Aggressive 3."`.
- `Triggers`: the Unbowed trigger unchanged, then `ability.Aggressive(3)` with `Cornered: true`. Go cannot set a field on a call result inside a composite literal, so add one unexported helper beside `pathDefs`: `func cornered(t ability.Trigger) ability.Trigger { t.Cornered = true; return t }`, and write `cornered(ability.Aggressive(3))`.
- "Cornered: Aggressive 3." does not match the keyword-line pattern in `cardload.go` (it has a prefix and a period), so it stays a rules-text line that `Text` must equal. `cardload.go` does not change.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
Written by `rules-test-author`, in `pkg/engine/cards_path_allshan_test.go` (new test functions only; the T-0070 tests stay unchanged):
- `TestCard_Allshan_CorneredAggressiveIsTheKeywordTrigger`: while Cornered, Allshan's attack places a `StackAbility` item with `Keyword: ability.KeywordAggressive`; not Cornered, none is placed.

## Helpers Needed
- none expected

## Allowed Files
- `data/cards/units-and-structures-path.csv` (Allshan's second rules-text line only)
- `pkg/engine/cards/path.go`

## Done When
- [ ] The T-0070 tests and the new test pass: `go test ./pkg/engine -run 'Allshan'`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-compare -- autoloop/m2-wave-b` passes, or the Log names every changed line (only the place wording of Allshan's Cornered trigger may differ)
- [ ] the card loads with no unsupported lines

## Log
- 2026-10-26, lead-architect (wave B run, session 7): written with T-0074 from directive #60 (2). Stays `queued` until T-0074 is `done`. T-0073 now also depends on it, because #60 blocks T-0073.
- 2026-10-26, lead-architect (wave B run, session 8): T-0074 is `done` (125b9ab), so directive #60 (2) applies here. Branch created from `autoloop/m2-wave-b` at d5a7d39. Status `tests`; dispatched `rules-test-author`.
- 2026-10-26, rules-test-author: wrote `TestCard_Allshan_CorneredAggressiveIsTheKeywordTrigger` (subtests cornered / not cornered; also checks source and +3 TempPower) in `pkg/engine/cards_path_allshan_test.go`. Helpers requested: none (reuses `recordPlaced`). Left out: a printed-text check (no test-visible surface for card Text; the CSV/lint covers it). Verified red before implementation: the cornered subtest fails (no Aggressive-keyword item is placed, the old hand-written trigger is a plain ability item); not cornered passes. `cardtown lint rules` passes.
- 2026-10-26, lead-architect (wave B run, session 8): tests approved. `TestCard_Allshan_CorneredAggressiveIsTheKeywordTrigger` checks the Cornered case (one `StackAbility` item marked `KeywordAggressive`, sourced from Allshan, +3 power) and the not-Cornered case (none), per `FR-KWD-002e`/`FR-KWD-002f`; the T-0070 tests are unchanged. Status `ready`; dispatching the flash engineer.
- 2026-10-26, lead-architect (wave B run, session 8): the Flash run (`agy`, conversation `bea5448f-2608-45eb-8db1-0e328bb6cb5b`) exited 0 with `status: SUCCESS` but one denied action: at step 14 the engineer ran `git branch -a`, which is off the allow-list (its skill names `git rev-parse --verify <branch>` instead). Before that it had searched the main checkout (on `main`) for the task file with `Get-ChildItem`. This task file exists only on the run branch and this task branch, so it found nothing. It made no commits and left no worktree. Status stays `ready`. Not re-dispatched (D3); the run stops for the user.
