---
id: T-0051
title: Make Hammerhand Brigands executable (wave B card, flash)
kind: card
tier: flash
wave: B
status: review
branch: m2/T-0051-hammerhand-brigands
depends_on: [T-0050]
claims: [card:Hammerhand Brigands]
cards: [Hammerhand Brigands]
requirements: []
---

## Goal
*Hammerhand Brigands* (Path, 3 copies in the starter deck) now runs in the kernel. It is plain `ability` data: the damage-taken hook from T-0050 plus the existing `Scratch` step. It is the first wave B card.

## Inputs
- Card (`data/cards/units-and-structures-path.csv`): Path Unit (Infantry, Hammerhand), cost 1R 0T, deploys to Reserves, 2/1, yields 1R 2T, rarity Common. It prints one line, unsupported today:
  - "Unbowed: Scratch 1."
- Survey row: `docs/design/m2_starter_set_ability_survey.md` line 139. FRD 01 names this card in `FR-KWD-002a`'s second bullet (line 506): its `Scratch 1` does not act on the destroyed asset, so it resolves fully from the discard pile.
- Shapes it uses (all `done`):
  - `HookDamaged` (T-0050);
  - `OpScratch` with the default subject, `SubjectController` (T-0007).
- Rulings and requirements for context:
  - `FR-KWD-002a` Unbowed: fires whenever the asset takes damage, including lethal damage; the trigger belongs to the last-known controller.
  - `FR-REP-004` Stage 5: "takes damage" means Armor absorbed plus Health committed is more than 0. Armor absorbing everything still counts; damage reduced to 0 does not.
  - `FR-TRG-003`/`FR-TRG-004`: the trigger is buffered and placed after the resolving item and its sweeps, in APNAP order.
  - `FR-ZON-004`: a face-down Scratch Company fires no printed trigger.
  - Keyword matrix: `Unb × Scr` (`docs/frd/keyword_interaction_matrix.md` lines 74–81).
- Board directive #14: Brigands is a deck card. Devilhorn Captain, the other Unbowed deck card T-0050 named, needs a zone criterion on `Filter` ("in the Main Theater", `FR-TGT-001`) that does not exist yet, so it moved to the shape task T-0052.

### Definition (the engineer adds it to `pathDefs` in `pkg/engine/cards/path.go`)
- One trigger: `Hook: ability.HookDamaged`, no targets.
- One step: `Op: ability.OpScratch`, `Amount: ability.Value{N: 1}`, default subject (the trigger's controller).
- `Text` holds the quoted line exactly.
- No new types, fields or logic. If the card seems to need one, set `blocked`.
- **Reviewer pre-check** (a throwaway prototype of exactly this definition, not committed, 2026-10-23):
  - `go test ./...` passes, `TestProgress_Case4_Coverage` included.
  - The 200-game starter sim exits 0 and the monitors find no violation. Path rises from 13/50 to 16/50 executable. Summary line before: `P0 wins 88.0%  P1 wins 12.0%  draws 0.0%  first player wins 50.0%`, 35.5 turns. After: `P0 wins 63.0%  P1 wins 36.5%  draws 0.5%  first player wins 56.0%`, 34.2 turns.

### What the tests can observe (for rules-test-author)
- It enters Reserves as a 2/1 for 1 Requisition. Nothing happens on deploy.
- **Fires (its controller Scratches 1: the top card of their deck becomes a face-down Scratch Company):**
  - Lethal combat damage, as attacker (retaliation) or as defender. It dies, and the Scratch still happens, from the discard pile.
  - Non-combat damage from a real card: for example *Saboteur*'s "1 damage to chosen allied unit" aimed at an allied Brigands, or *Devilhorn Guerrillas*'s "2 damage to chosen opposing unit" deployed by the opponent.
  - Armor that absorbs all the damage: it survives with 1 health and still Scratches 1.
  - In Reserves as well as in the Main Theater.
  - Once per damage instance: two Brigands hit by one "all assets" effect (for example *Craterfall Ancient*'s death trigger) give two Scratches; one Brigands hit twice by two separate instances (while Armor keeps it alive) gives two.
  - Under another player's control: the Scratch goes to its controller at the time of the damage, not its owner (`putControlled` in `unbowed_test.go` sets this up).
- **Does not fire:**
  - Damage reduced to 0 before Stage 4 (for example a 0-power attacker, or Resistant from a scene keyword).
  - Damage dealt to another asset.
  - A Hammerhand Brigands card that is a face-down Scratch Company (`FR-ZON-004`).
- No question is ever asked.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
Written in `pkg/engine/cards_path_test.go`. Run with `go test ./pkg/engine -run TestCard_HammerhandBrigands`:
- `TestCard_HammerhandBrigands_EntersReservesAsTwoOneForOneRequisitionAndDoesNothing`
- `..._LethalCombatDamageAsDefenderScratchesFromDiscard` (also: no question asked, empty stack, invariants)
- `..._LethalRetaliationAsAttackerScratchesFromDiscard`
- `..._ArmorAbsorbingAllDamageStillScratches`
- `..._SaboteurDamageToAlliedBrigandsInReservesScratches` (real Saboteur, Brigands in Reserves)
- `..._OpposingGuerrillasDamageScratchesForTheBrigandsController` (real Guerrillas deployed by the opponent)
- `..._TwoBrigandsHitByOneAllAssetsEffectScratchTwice` (synthetic all-assets 1-damage deploy trigger)
- `..._OneBrigandsHitByTwoSeparateInstancesScratchesTwice` (Armor keeps it alive)
- `..._ScratchGoesToTheControllerNotTheOwner` (`putControlled`, survives on Armor)
- `..._ZeroPowerAttackerDoesNotScratch`
- `..._DamageReducedToZeroByResistantDoesNotScratch`
- `..._DamageToAnotherAssetDoesNotScratch`
- `..._FaceDownCopyFiresNothingWhenDamaged` (FR-ZON-004)

Failing today: the enters-Reserves test (Deploy is rejected for a not-yet-executable card) and the eight firing tests (no Scratch happens). The four negative tests (zero power, Resistant, other asset, face-down copy) pass today and guard against over-firing.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None requested. `putBrigands(p, zone, armor)` is a local helper defined in `cards_path_test.go`. Existing helpers used include `deployPath`, `colossusCycle`, `wantOptions`, `scriptPlan`, `pick`, `wantDealt` and, in `unbowed_test.go`, `putControlled` and `wantScratches`.

## Allowed Files
- `pkg/engine/cards/path.go`
- `pkg/engine/cards/cards.go` (remove the name from `NotYetExecutable`)

## Done When
- [x] The card's tests pass: `go test ./pkg/engine -run 'TestCard_HammerhandBrigands'`
- [x] `TestDefs_MatchRulesText` and `TestStarterSet_Executable` pass unchanged
- [x] `go tool task check` passes, `TestProgress_Case4_Coverage` included
- [x] `go tool task sim` runs clean: every invariant holds and the rule monitors report no violation. Path's executable count rises by 3, from 13/50 to 16/50. Report the before and after summary lines in the Log.

## Log
- 2026-10-23 lead-architect: Wrote the task (wave B card task 1), with status `tests` for `rules-test-author`. Devilhorn Captain was split out to T-0052 because it needs a new `Filter` zone criterion. Starter-deck cards not yet executable before this task: 30, plus both Headquarters. Expected after: 29.
- 2026-10-23 rules-test-author: Wrote 13 `TestCard_HammerhandBrigands_*` tests in `cards_path_test.go` (list under Tests). No helpers requested. Left out: no real Craterfall Ancient (not executable; a synthetic all-assets damage source stands in); no test pins the order of the Unbowed trigger against death triggers in one sweep (open question noted in `unbowed_test.go`); no `TestFR_*` tests since `requirements` is empty. Status set to `tests-review`.
- 2026-10-23 lead-architect: Test review (D4): **approved**, status `ready`.
  - a1aedf6 adds 13 `TestCard_HammerhandBrigands_*` tests and the local `putBrigands` helper, and changes no existing test.
  - Rules content checks out against `FR-KWD-002a` (lethal damage resolves from discard, controller at the time of the damage), `FR-REP-004` Stage 5 (Armor-only absorption counts; 0 power and Resistant to 0 do not; once per instance) and `FR-ZON-004` (face-down copy).
  - Correction to the test author's note: Craterfall Ancient has been executable since T-0031. The synthetic all-assets source is still acceptable, because it isolates the one-trigger-per-asset behavior.
  - **Reviewer check** (the throwaway definition from the pre-check, not committed): all 13 pass, and so does `go test ./...`.
  - **Mutations each fail tests:** `HookKilled` instead of `HookDamaged`: 3; `SubjectOpponents`: 8; Scratch 2: 8.
  - Next: the lead dispatches the flash engineer through `agy` (D3).
- 2026-10-23 kernel-engineer: Implementation handoff (status `review`):
  - Diff shows:
    - `pkg/engine/cards/path.go`: Added ability definition for "Hammerhand Brigands" using existing `ability.HookDamaged` and `ability.OpScratch` with amount 1.
    - `pkg/engine/cards/cards.go`: Removed "Hammerhand Brigands" from `NotYetExecutable`.
  - Reused shapes:
    - `HookDamaged` from T-0050 for Unbowed damage trigger hook.
    - `OpScratch` with default subject `SubjectController` (0) and `ability.Value{N: 1}`.
  - Commands and results:
    - `go test ./pkg/engine -run TestCard_HammerhandBrigands`: all 13 tests PASS.
    - `go test ./pkg/engine/... -run 'TestDefs_MatchRulesText|TestStarterSet_Executable'`: PASS.
    - `go tool task check`: PASS (fmt-check, vet, all tests, linter dispatch clean).
    - `go tool task sim`: clean run, 200 games, all invariants hold, no monitor violations. Path executable count rose from 13/50 to 16/50.
      - Before: `P0 wins 88.0%  P1 wins 12.0%  draws 0.0%  first player wins 50.0%`, 35.5 turns (13/50 executable)
      - After: `P0 wins 63.0%  P1 wins 36.5%  draws 0.5%  first player wins 56.0%`, 34.2 turns (16/50 executable)
