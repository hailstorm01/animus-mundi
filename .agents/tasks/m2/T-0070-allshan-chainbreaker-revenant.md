---
id: T-0070
title: Make Allshan, Chainbreaker Revenant executable (wave B card, flash)
kind: card
tier: flash
wave: B
status: queued
branch: m2/T-0070-allshan-chainbreaker-revenant
depends_on: [T-0069]
claims: [card:Allshan, Chainbreaker Revenant]
cards: [Allshan, Chainbreaker Revenant]
requirements: []
---

## Goal
*Allshan, Chainbreaker Revenant* (Path unit, one copy in the Path starter deck) runs in the kernel. It is plain `ability` data in `pkg/engine/cards/path.go`. Its two rules-text lines are:
- "Unbowed: This unit gains 2 armor."
- "Cornered: Whenever this unit attacks, it gets +3 power until end of turn."

## Inputs
- Card (`data/cards/units-and-structures-path.csv`): Path unit, Specialist, Mythic, cost 5, deploys to Reserves, 5/6, with the two lines above.
- Survey row: `docs/design/m2_starter_set_ability_survey.md` line 147.
- Shapes it uses:
  - `ability.HookDamaged` (Unbowed, T-0050).
  - `OpAddArmor` (A2).
  - `HookAttacks` (A1).
  - `Trigger.Cornered` and `OpModifyStat` (T-0069).
  - `SubjectSelf`.
- Rulings: `SVY-R13` (Unbowed still fires on lethal damage, after the death sweep). `FR-KWD-002e` and `FR-LAY-003` (Cornered is checked when the attack is declared).

### Definition (the engineer adds it to `pathDefs`)
- Trigger 1: `Hook: ability.HookDamaged`, one step `{Op: ability.OpAddArmor, Subject: ability.SubjectSelf, Amount: ability.Value{N: 2}}`.
- Trigger 2: `Hook: ability.HookAttacks`, `Cornered: true`, one step `{Op: ability.OpModifyStat, Subject: ability.SubjectSelf, Amount: ability.Value{N: 3}}`.
- `Text` holds the two printed lines, in printed order.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
To be written by `rules-test-author` in `pkg/engine/cards_path_allshan_test.go`, from the CSV text, when this task reaches `tests`. At least:
- deploys to Reserves as a 5/6 for its cost, and nothing fires on deploy;
- Unbowed: surviving damage gives 2 Armor, which absorbs later damage;
- attacking while Cornered deals 8 combat damage, and power is back to 5 after `endTurn`;
- attacking while not Cornered (equal Main Theater counts) gains nothing.

## Helpers Needed
None expected.

## Allowed Files
- `pkg/engine/cards/path.go`
- `pkg/engine/cards/cards.go` (remove the name from `NotYetExecutable`)

## Done When
- [ ] The card's tests pass: `go test ./pkg/engine -run 'TestCard_AllshanChainbreakerRevenant'`
- [ ] `TestDefs_MatchRulesText` and `TestStarterSet_Executable` pass unchanged
- [ ] `go tool task check` passes
- [ ] `go tool task sim` runs clean; report the before and after summary lines in the Log

## Log
- 2026-10-26 lead-architect (wave run B, session 3): written with T-0069 and left `queued` until T-0069 is `done`.
