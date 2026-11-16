---
id: T-0250
title: "Blindside: chosen opposing unit gets -3 power until the end of your opponent's turn (wave H card, flash)"
kind: card
tier: flash
wave: H
status: ready
branch: m2/autoloop-wave-h/T-0250-blindside
depends_on: [T-0249]
claims: [card:Blindside]
cards: [Blindside]
requirements: [FR-LAY-001, FR-LAY-003, FR-TRN-001, FR-MNV-001]
---

## Goal
*Blindside* (Path Maneuver, `data/cards/maneuvers-path.csv:5`; R0 T2; in neither starter deck): "Chosen opposing unit gets -3 power until the end of your opponent's turn." becomes executable from shapes that are `done`: one `AllegianceOpposing` unit target and `OpModifyStat` with T-0249's `Step.Until: ability.UntilEndOfOpponentTurn`. The card leaves `cards.NotYetExecutable`.

## Inputs
- Card text: `data/cards/maneuvers-path.csv:5`. Survey row: `docs/design/m2_starter_set_ability_survey.md:193`.
- FRD 01: `FR-LAY-003` *Full-Round Durations* (line 514, which names *Blindside*: active through the rest of the current turn and throughout the opponent's next turn, expiring in Sub-Phase 4c of that opponent's End Phase); `FR-TRN-001` (line 301, Sub-Phase 4c); `FR-LAY-001` (line 500, power is clamped at 0); `FR-TRN-003a`/`FR-MNV-001` (line 372, a Maneuver with no legal target can't be enacted).
- Code of the day (run branch at 451e613): `ability.Duration` and `UntilEndOfOpponentTurn` (`pkg/engine/ability/ability.go:183-187`), `Step.Until` (`:415`), `OpModifyStat` (`:376`); `NotYetExecutable` (`pkg/engine/cards/cards.go:12`). T-0249 (merge e1f7550) pinned the duration on synthetic Maneuvers in `pkg/engine/opponent_turn_duration_test.go`; this task pins it on the real card.

## Shape (the lead's plan)
- `pkg/engine/cards/path.go`: the definition, in alphabetical order (after *Blind Defile*, before *Boneyard Amalgam*). `Text` is the CSV line. `Effect` has one target `{Filter: ability.Filter{Allegiance: ability.AllegianceOpposing, Kinds: ability.KindUnit}}` and one step `{Op: ability.OpModifyStat, Amount: ability.Value{N: -3}, Subject: ability.SubjectTarget, Target: 0, Until: ability.UntilEndOfOpponentTurn}`. No new shape, no new field.
- `pkg/engine/cards/cards.go`: remove "Blindside" from `NotYetExecutable`.
- The lead dry-ran the definition and the removal on the run branch at 451e613 (2026-11-15, wave H session 11): `gofmt -l pkg cmd` clean, `go tool task check` exit 0, `go tool task sim` exit 0 (P0 58.5%, P1 39.5%, draws 2.0%), `go tool task sim-compare -- autoloop/m2-wave-h` `[PASS]` (byte-identical: the card is in neither starter deck).

## Tests (card tasks: written by rules-test-author; engineers never edit these)
New tests go in `pkg/engine/cards_path_blindside_test.go`. Use the real card (`s.card("Blindside")`), given to player 0 (`s.give`) with exactly R0 T2 (`s.setPool(0, 0, 2)`), enacted with `s.do(enactAction(inst))` in player 0's Main Phase. Read power with `s.g.Power(a)`. Each scene below has exactly one legal target unless it says otherwise, so the choice is forced.
- `TestCard_Blindside_LastsThroughOpponentsTurn` (`FR-LAY-003`, `FR-TRN-001`): an enemy unit with printed power 4 in the Main Theater. Right after *Blindside* resolves its power is 1. After `s.endTurn()` (player 1's Main Phase) it is still 1. After the next `s.endTurn()` (player 1's turn ended) it is 4, and it is still 4 after player 0's following turn ends.
- `TestCard_Blindside_ReducedPowerInOpponentsAttack` (`FR-LAY-003`, `FR-LAY-001`): an enemy unit with printed power 4 in the Main Theater, and an allied unit with enough health to survive in the Main Theater. After *Blindside* and `s.endTurn()`, player 1 attacks the allied unit with the enemy unit (`s.attack`). The allied unit takes 1 damage, not 4 (`s.wantDamage`).
- `TestCard_Blindside_ClampsAtZero` (`FR-LAY-001` Layer 6): an enemy unit with printed power 2: its power is 0 after *Blindside* (not negative), and 2 again after player 1's turn ends.
- `TestCard_Blindside_AlliedUnitIsNotATarget` (`FR-TGT-001`, `FR-MNV-001`): an allied unit and an enemy unit in the Main Theater: only the enemy unit's power changes, and the allied unit's power is unchanged. If a target decision is asked, its options hold only the enemy unit.
- `TestCard_Blindside_NotEnactableWithoutOpposingUnit` (`FR-MNV-001`), table-driven like `armorNotEnactable` (`pkg/engine/cards_path_maneuvers_test.go:558`) with the allegiances swapped (don't edit that helper): no unit at all → not enactable; only an opposing structure and allied units → not enactable; one opposing unit in player 1's Reserves → enactable. A refused enactment leaves *Blindside* in player 0's hand.
- `TestDefs_MatchRulesText` already checks `Text` against the CSV.
- Not this task: a target that leaves before resolution (`FR-TGT-003`). *Blindside* has no way to be responded to in a test without a synthetic responder, and the second target check is the same code every targeted Maneuver goes through.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- none. The tests use existing helpers only (`enactAction`, `offersEnact`, `inHand`, `chooseAsset`, `setPool`, `wantPower` from `opponent_turn_duration_test.go`).

## Allowed Files
- `pkg/engine/cards/path.go`
- `pkg/engine/cards/cards.go`

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run "Blindside|TestDefs"`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-compare -- autoloop/m2-wave-h` passes (*Blindside* is in neither starter deck)

## Log
- 2026-11-15 lead-architect (wave H run, session 10): queued, split off T-0249 under the D2 size rule (shape and card together touch six non-test files). Waits for T-0249.
- 2026-11-15 lead-architect (wave H run, session 11): T-0249 is `done` (merge e1f7550), and the names it merged (`Step.Until`, `ability.UntilEndOfOpponentTurn`) are the ones planned. Wrote Inputs, Shape, Tests and Allowed Files against the run branch at 451e613 and dry-ran the definition (above). No new shape: `OpModifyStat` with a negative `Amount` already exists (*Ensnaring Mist*, `pkg/engine/cards/animus.go:304`). Dropped the fizzle case from the queued Covers (see Tests). Re-read `BOARD.md` before committing: no directive added since, none applies. Status `tests`.
- 2026-11-15 rules-test-author: wrote `pkg/engine/cards_path_blindside_test.go`: `TestCard_Blindside_LastsThroughOpponentsTurn`, `_ReducedPowerInOpponentsAttack`, `_ClampsAtZero`, `_AlliedUnitIsNotATarget`, `_NotEnactableWithoutOpposingUnit`. No helpers requested; the file adds one local helper `blindside(s)`. Nothing left out beyond the task's own exclusion (FR-TGT-003). All five fail now for the right reason (card not yet executable). `cardtown lint rules` passes. Status `tests-review`.
- 2026-11-15 lead-architect (D4): tests approved. Each checks what `FR-LAY-003` (line 514, which names *Blindside*), `FR-TRN-001` 4c and `FR-LAY-001` give, on the real card: -3 lasts through player 1's turn and expires in its Sub-Phase 4c, not player 0's; player 1's attack deals the reduced power (1, not 4); power is clamped at 0 and returns to printed; only the opposing unit is offered and changed; with no opposing unit (none at all, or only an opposing structure) it isn't offered or enactable and stays in hand, and an opposing unit in Reserves makes it enactable. No existing test edited; `wantPower` reused from `opponent_turn_duration_test.go`. Status `ready`, dispatched to `kernel-engineer` (flash).
