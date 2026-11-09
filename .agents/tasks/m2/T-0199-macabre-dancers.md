---
id: T-0199
title: Make Macabre Dancers executable (Unbowed heals 2, Cornered Resistant 1, wave E card, flash)
kind: card
tier: flash
wave: E
status: in-progress
branch: m2/autoloop-wave-e/T-0199-macabre-dancers
depends_on: [T-0050, T-0180]
claims: [card:Macabre Dancers]
cards: [Macabre Dancers]
requirements: [FR-KWD-002a, FR-KWD-002d, FR-KWD-002e, FR-KWD-002b]
---

## Goal
*Macabre Dancers* (a Path deck card, 2 copies): *"Unbowed: Heal 2."* and *"Cornered: Resistant 1."* It only combines shapes that are `done`: an Unbowed trigger on `HookDamaged` (as *Allshan, Chainbreaker Revenant*), `OpHeal` on `SubjectSelf`, and T-0180's `CorneredKeywords.Resistant`.

Covers:
- The definition in `pkg/engine/cards/path.go`, as under Definition below.
- The name off `cards.NotYetExecutable`.

## Inputs
- Card (`data/cards/units-and-structures-path.csv` lines 11–12): Path Unit, Specialist, Reaver, R2 T2, deploys to Reserves, power 3, health 6.
- Survey row: `docs/design/m2_starter_set_ability_survey.md` line 188 ("Heal fizzles if lethal").
- FRD 01: `FR-KWD-002a` (line 577, which names this card: a lethal Unbowed still triggers, and its `Heal 2` fizzles once the asset is destroyed), `FR-KWD-002d` (line 582, Heal N never removes more than the accumulated damage), `FR-KWD-002e` (line 583, a keyword behind Cornered), `FR-KWD-002b` (line 579), `FR-REP-004` (line 500; Unbowed needs damage actually taken, line 527).
- Rulings: `SVY-R13` (lethal Unbowed placed after the death sweep), `SVY-R18` (a keyword behind Cornered counts only while Cornered).

### Definition (the engineer adds it to `pathDefs`, between *Ironclad Colossus* and *Necropolis Hunt*)
- `Text`: `"Unbowed: Heal 2."`, then `"Cornered: Resistant 1."`.
- `Triggers`: one, `Hook: ability.HookDamaged`, with one step `{Op: ability.OpHeal, Subject: ability.SubjectSelf, Amount: ability.Value{N: 2}}`.
- `Cornered: ability.CorneredKeywords{Resistant: 1}`.
- Remove the name from `cards.NotYetExecutable`.

A lead probe with this definition on the run branch at 54bcada, since reverted, passes `go tool task check`, and `go tool task sim` stays clean: Path 36/50 main cards executable, P0 79.5% / P1 18.5% / draws 2.0% over 200 games. So `sim-compare` will differ here and isn't a gate.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
In a new file `pkg/engine/cards_path_dancers_test.go`, from the CSV text, loading the real card (`s.card`), in the style of `cards_path_allshan_test.go`. Make player 1 Cornered, or not, by the Main Theater counts (`FR-KWD-002e`), as `cornered_keywords_test.go` does.
- `TestCard_MacabreDancers_UnbowedHealsTwoAfterSurvivingDamage` (`FR-KWD-002a`, `FR-KWD-002d`): 3 non-lethal damage, not Cornered. After the trigger resolves, it has 1 damage.
- `TestCard_MacabreDancers_HealNeverRemovesMoreThanItsDamage` (`FR-KWD-002d`): 1 damage, not Cornered. After the trigger resolves, it has 0 damage and its Current Health is its Effective Maximum Health.
- `TestCard_MacabreDancers_LethalUnbowedTriggersAndFizzles` (`FR-KWD-002a`, `SVY-R13`): 6 or more damage. It is in its owner's discard pile, the Unbowed trigger was placed and resolved after the sweep, and it healed nothing (it is not back on the battlefield, and nothing else changed).
- `TestCard_MacabreDancers_CorneredResistantReducesDamage` (`FR-KWD-002e`, `FR-KWD-002b`, `SVY-R18`): player 1 Cornered, 3 damage dealt. It takes 2, then heals to 0.
- `TestCard_MacabreDancers_CorneredResistantStopsOneDamageAndUnbowed` (`FR-KWD-002a`, `FR-REP-004`): player 1 Cornered, 1 damage dealt. It takes none, and no Unbowed trigger is placed.
- `TestCard_MacabreDancers_NotCorneredHasNoResistant` (`SVY-R18`): not Cornered, 1 damage dealt. It takes 1, its effective keywords (`g.Keywords`) have Resistant 0, and the trigger heals it to 0.
- **Existing tests:** every other test passes unchanged, including `TestStarterSet_Executable` with the name removed, `TestView_Case8_ExecutableFlag`, `TestProgress_Case4_Coverage` and `TestEvents_Case4_EveryActionKind`.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected.

## Allowed Files
- `pkg/engine/cards/path.go` (the definition)
- `pkg/engine/cards/cards.go` (remove the name from `NotYetExecutable`)

## Done When
- [ ] The card tests pass: `go test ./pkg/engine -run TestCard_MacabreDancers`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] The card loads with no unsupported lines (`TestStarterSet_Executable` passes with the name removed)

## Log
- 2026-11-09 lead-architect (wave E run, session 1): written against 54bcada, with the lead's probe above. Board read twice (D2.0), no `open: before` directive; every `waiting:` trigger checked and none fired (#118 (b) waits for *Boneyard Amalgam*, the next wave E task). Searched `pkg/engine` and the task archive: no new shape, every part of the Definition exists. Set to `tests`.
- 2026-11-09 rules-test-author: wrote `pkg/engine/cards_path_dancers_test.go` (6 tests, all `TestCard_MacabreDancers_*` as listed under Tests: UnbowedHealsTwoAfterSurvivingDamage, HealNeverRemovesMoreThanItsDamage, LethalUnbowedTriggersAndFizzles, CorneredResistantReducesDamage, CorneredResistantStopsOneDamageAndUnbowed, NotCorneredHasNoResistant). Helpers requested: none (a local `dancersScene` lives in the test file). All six fail now (card not executable: no Unbowed trigger, no Cornered Resistant); `go vet` is clean and `cardtown lint rules` passes. No cases left out. Set to `tests-review`.
- 2026-11-09 lead-architect: tests approved (D4, round 1), each checked against `FR-KWD-002a` (a lethal Unbowed triggers, is placed after the sweep, and its self-heal fizzles; no damage taken, no trigger), `FR-KWD-002d`, `FR-KWD-002e`/`FR-KWD-002b` with `SVY-R18`, and `SVY-R13`. A lead probe with the Definition above, since reverted, passes all six and `go test ./...`. Set to `ready`.
