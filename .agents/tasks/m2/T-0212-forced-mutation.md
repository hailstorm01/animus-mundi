---
id: T-0212
title: Forced Mutation, turning a chosen opposing unit face down (wave F deck card, flash)
kind: card
tier: flash
wave: F
status: tests
branch: m2/autoloop-wave-f/T-0212-forced-mutation
depends_on: [T-0211]
claims: [card:Forced Mutation]
cards: [Forced Mutation]
requirements: [FR-ENT-103, FR-MNV-001, FR-TGT-001]
---

## Goal
*Forced Mutation* (Animus Maneuver, R0 T4, two copies in `data/decks/animus-starter.txt`): "Turn chosen opposing unit face down. It becomes a Scratch Company unit." One exactly-one target (opposing unit, either zone) and one `OpMorphScratch` step on it (T-0211). It combines shapes only.

Covers:
- Its definition in `pkg/engine/cards/animus.go`, between *Feed the Despoiler* and *Funeral Fungus*, and its name removed from `cards.NotYetExecutable`.
- Card tests from the CSV text: the target is turned face down in place and keeps its state (`FR-ENT-103`); a damaged target dies in the sweep after the item (`FR-LAY-002`); only opposing units are offered, in either zone, and not a Ward unit (`FR-TGT-001`, `FR-KWD-002h`); it can't be enacted without a legal opposing unit (`SVY-R29`, `FR-MNV-001`).

## Inputs
- Card (`data/cards/maneuvers-animus.csv` line 17): Animus Maneuver, R0 T4, "Turn chosen opposing unit face down. It becomes a Scratch Company unit." The second sentence states what `FR-ENT-103` does; it is not a second step.
- Survey row: `docs/design/m2_starter_set_ability_survey.md` line 159 ("`MORPH_SCRATCH` chosen opposing unit", `FR-ENT-103`).
- FRD 01 v3.23: `FR-ENT-103` (line 260; its first item: in place, a 1/1 Infantry Scratch Company with only Deploys Exhausted, keeping readiness, damage, Armor and temporary modifiers; lethal damage destroys it, `FR-LAY-002`). Its second item, the Attached purge, has nothing to act on in M2 and is directive #142, not this task. `FR-MNV-001` with "A Maneuver Needs Its Required Targets" (line 358, `SVY-R29`). `FR-TGT-001` (line 370). `FR-LAY-001` (line 481). `FR-LAY-002` (line 487). `FR-KWD-002h` Ward (line 601).
- Rulings: `SVY-R29`.
- Design note: `docs/architecture/m2-ability-model.md`, Key types (`Filter`, `Step`).
- Code: `ability.OpMorphScratch` and `stepMorphScratch` ([effects.go:455-472](../../../pkg/engine/effects.go), T-0211); an opposing-unit target in the same form as *Disorienting Pulse*'s ([cards/animus.go:236-256](../../../pkg/engine/cards/animus.go)), without its zone.

### Definition (the engineer adds it to `animusDefs`, between *Feed the Despoiler* and *Funeral Fungus*)
- `Text`: the CSV text above, as one line.
- `Effect.Targets`: one exactly-one slot, `Filter{Allegiance: ability.AllegianceOpposing, Kinds: ability.KindUnit}` (no `Zone`: either zone).
- `Effect.Steps`: `{Op: ability.OpMorphScratch, Subject: ability.SubjectTarget, Target: 0}`.
- Remove the name from `cards.NotYetExecutable`.

A lead probe with this definition on the run branch at 7259ec3, since reverted, passes `go tool task check`, and `go tool task sim` stays clean: Animus 47 → 49/50 main cards executable, P0 74.0% / P1 25.0% / draws 1.0% over 200 games. So `sim-compare` will differ here and isn't a gate.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
Add them to `pkg/engine/cards_animus_maneuvers_test.go`, after the *Disorienting Pulse* tests, in their style: load the real Maneuver with `enactFor` (Req 3, Tac 5), use `chooseAsset`, `targetOffers`, `offersEnact` and `s.reject`, and synthetic opposing cards (`unit`, `structure`). Player 0 enacts. Script only the decisions actually asked: a single legal target is chosen without a question (`FR-MOD-001` §1).
- `TestCard_ForcedMutation_TurnsChosenOpposingUnitFaceDownInPlace` (`FR-ENT-103`, `FR-MNV-001`, `FR-LAY-001`): two opposing units in player 1's Main Theater. Choose one, a power-4 health-5 Stalwart unit, Exhausted, with Armor 2. After resolution it is the same asset (same ID), still in player 1's Main Theater and controlled by player 1, face down, still Exhausted, with Armor 2; its effective power and Maximum Health are 1 and its effective keywords are Deploys Exhausted only. The other opposing unit is not face down. *Forced Mutation* is in player 0's discard pile, and player 0's Tactics went 5 → 1.
- `TestCard_ForcedMutation_DamagedTargetIsDestroyed` (`FR-ENT-103`, `FR-LAY-002`): the only opposing unit has 1 damage on 4 health. After resolution it is no longer on the battlefield, and its own card is in player 1's discard pile.
- `TestCard_ForcedMutation_OnlyOpposingUnitsAreOffered` (`FR-TGT-001`, `FR-KWD-002h`): player 1 has a unit in the Main Theater, a unit in Reserves, a Structure in the Main Theater and a Ward unit in the Main Theater; player 0 has a unit in the Main Theater. The target decision offers exactly player 1's two non-Ward units (both zones); the Structure, the Ward unit, player 0's unit and both Headquarters are not offered. Choosing the Reserves unit turns it face down in Reserves, and the Main Theater unit stays face up.
- `TestCard_ForcedMutation_NotEnactableWithoutLegalTarget` (`SVY-R29`, `FR-MNV-001`), table-driven like *Disorienting Pulse*'s: Enact is not offered, and a rejected Enact leaves the card in hand, with (a) no units on the battlefield, (b) only an opposing Structure, (c) only an allied unit, (d) only an opposing Ward unit. It is offered with (e) one plain opposing unit in player 1's Reserves.
- **Existing tests:** every other test passes unchanged, including `TestStarterSet_Executable` with the name removed, `TestView_Case8_ExecutableFlag`, `TestProgress_Case4_Coverage`, `TestEvents_Case4_EveryActionKind`, and the T-0211 tests in `morph_scratch_test.go`.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected. The test author keeps any small builder in the test file.

## Allowed Files
- `pkg/engine/cards/animus.go` (the definition)
- `pkg/engine/cards/cards.go` (remove the name from `NotYetExecutable`)

## Done When
- [ ] The card tests pass: `go test ./pkg/engine -run TestCard_ForcedMutation`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] The card loads with no unsupported lines (`TestStarterSet_Executable` passes with the name removed)

## Log
- 2026-11-10 lead-architect (wave F run, session 11): queued with T-0211 (the `OpMorphScratch` shape), split under the D2 size rule.
- 2026-11-10 lead-architect (wave F run, session 12): Inputs, Definition, Tests and Allowed Files written against the run branch at 7259ec3, with T-0211 `done`. Searched `pkg/engine` and the task archive for an existing equivalent: `OpMorphScratch` (T-0211) is the shape, reused as is, with *Disorienting Pulse*'s target form; no new shape. No waiting directive fires with this card (#142 waits for the first card that tucks). Probed the definition (see above), then reverted it. Set to `tests`.
