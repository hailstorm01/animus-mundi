---
id: T-0257
title: "Animoceps Stalk, Animoceps Node, Melodic Pummeler: same-zone power auras (wave H cards, flash)"
kind: card
tier: flash
wave: H
status: review
branch: m2/autoloop-wave-h/T-0257-power-aura-cards
depends_on: [T-0256]
claims: [card:Animoceps Stalk, card:Animoceps Node, card:Melodic Pummeler]
cards: [Animoceps Stalk, Animoceps Node, Melodic Pummeler]
requirements: [FR-LAY-001, FR-TGT-001, FR-ZON-001]
---

## Goal
The three power auras of wave E execute, each one `ability.StaticAura` (T-0256) with a different `Filter`. The shapes are identical, so they form one card task under the D2 size rule:
- *Animoceps Stalk*: "Other allied assets in the same zone have +1 power." Filter: allied, `Other`, `ZoneRefSame`, Power 1. Survey `:162`: `excludeSelf`.
- *Animoceps Node*: "Allied units in the same zone have +1 power." Filter: allied units, `ZoneRefSame`, Power 1. Survey `:163`: a Structure, so it doesn't buff itself.
- *Melodic Pummeler*: "Allied Horror units in the same zone have +1 power." Filter: allied units with subtype Horror, `ZoneRefSame`, Power 1. Survey `:164`: includes itself.

Each card leaves `cards.NotYetExecutable`.

## Inputs
- Card text (`data/cards/units-and-structures-animus.csv`):
  - *Animoceps Stalk* `:7`: Unit, Fungus, deploys to the Main Theater, power 1, Health 4.
  - *Animoceps Node* `:23`: Structure, Fungus, Territory, deploys to the Main Theater, power 3, Health 8.
  - *Melodic Pummeler* `:26`: Unit, Horror, deploys to Reserves, power 4, Health 7.
  - None of the three is in a starter or stress deck.
- Survey (`docs/design/m2_starter_set_ability_survey.md:162-164`): the three rows.
- FRD 01 (`docs/frd/01_Game_Rules_FRD.md`):
  - `FR-LAY-001`: Layer 4 auras are evaluated continuously.
  - `FR-TGT-001` `SVY-R44`, last sentence: a continuous ability uses its source's current zone.
  - `FR-ZON-001`: Reserves are each player's own.
- Code of the day (run branch at 28166f6). T-0256 merged the names it planned:
  - `ability.StaticAura` and `Static.Filter` / `Static.Power` (`pkg/engine/ability/ability.go:56-75`).
  - `Game.Power` and `Game.auras` (`pkg/engine/layers.go:56-85`), which are the one reader.
  - `ActiveZone` stays `ZoneAny`, the zero value.
  - The only other `Statics` definition is *Veil Render* (`pkg/engine/cards/animus.go:708`).
  - `NotYetExecutable` is at `pkg/engine/cards/cards.go:13`.
  - T-0256 pinned the shape on synthetic sources in `pkg/engine/aura_test.go`: zone following, later arrivals, the source leaving, the face-down source, stacking and the clamp, controller change, combat damage and the sweep's last-known record. This task pins the three real cards' filters.

## Shape (the lead's plan)
- `pkg/engine/cards/animus.go`: three definitions, in the file's order. *Animoceps Node* goes before *Animoceps Shambler*, *Animoceps Stalk* after *Animoceps Spores* (before *Cliff of Portents*), and *Melodic Pummeler* after *Infested Apostle* (before *Portent Scrivener*). `Text` is the CSV line. Each has one `Statics` entry `{Kind: ability.StaticAura, Filter: …, Power: 1}`:
  - Stalk: `ability.Filter{Allegiance: ability.AllegianceAllied, Other: true, ZoneRef: ability.ZoneRefSame}`
  - Node: `ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit, ZoneRef: ability.ZoneRefSame}`
  - Pummeler: `ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit, Subtypes: []string{"Horror"}, ZoneRef: ability.ZoneRefSame}`
  - No new shape and no new field.
- `pkg/engine/cards/cards.go`: remove the three names from `NotYetExecutable`.
- The lead dry-ran these definitions and the removals on the run branch at 28166f6 (2026-11-16, wave H session 18):
  - `gofmt -l pkg cmd` clean.
  - `go tool task check` exit 0.
  - `go tool task sim` exit 0 (P0 58.5%, P1 39.5%, draws 2.0%).
  - `go tool task sim-compare -- autoloop/m2-wave-h` `[PASS]`.

## Tests (card tasks: written by rules-test-author; engineers never edit these)
New tests go in `pkg/engine/cards_animus_auras_test.go`.
- Use the real cards (`s.card("Animoceps Stalk")` and so on), placed with `s.put` and moved with `s.moveTo`.
- Read effective power with `s.g.Power(a)` (or a local helper; `wantPower` already exists in `aura_test.go` and can be reused).
- Other assets are synthetic (`unit`, `structure`, with `c.Subtypes` set where a subtype matters).
- Player 0 is the active player.
- Each expected value is printed power plus the aura.

Tests:
- `TestCard_AnimocepsStalk_OtherAlliedAssetsInItsZone` (`FR-LAY-001`, survey `excludeSelf`):
  - Setup: the Stalk in player 0's Main Theater.
  - Gain 1: player 0's unit (power 2) in the Main Theater has 3, and player 0's structure (power 0) there has 1 (any asset, not only units).
  - Unchanged: the Stalk itself has 1; player 0's unit in its Reserves keeps its printed power; player 1's unit (power 2) in the Main Theater has 2.
- `TestCard_AnimocepsStalk_TwoStalksBuffEachOther`: two Stalks in player 0's Main Theater each have power 2. Each is "other" to the other.
- `TestCard_AnimocepsStalk_FollowsItsZone` (`SVY-R44`, last sentence; `FR-ZON-001`):
  - Setup: the Stalk in the Main Theater with player 0's units A in the Main Theater and B in player 0's Reserves.
  - After `s.moveTo(stalk, ZoneReserves)`: A has its printed power, and B has +1.
  - In the same state, player 1's unit in player 1's Reserves has its printed power.
- `TestCard_AnimocepsNode_AlliedUnitsInItsZone` (`FR-LAY-001`, survey "does not buff itself"):
  - Setup: the Node in player 0's Main Theater.
  - Gain 1: player 0's unit (power 2) in the Main Theater has 3, and a face-down Scratch Company of player 0 moved into the Main Theater (`s.putScratch(0)`, then `s.moveTo`) has its printed power + 1, because a Scratch Company is a unit.
  - Unchanged: the Node itself has 3; player 0's structure (power 1) in the Main Theater has 1; player 0's unit in its Reserves and player 1's unit in the Main Theater keep their printed power.
- `TestCard_AnimocepsNode_FollowsItsZone` (`SVY-R44`): with the Node moved to player 0's Reserves, player 0's unit there has +1 and player 0's Main Theater unit has its printed power.
- `TestCard_MelodicPummeler_AlliedHorrorUnitsInItsZone` (`FR-LAY-001`, survey "includes itself"):
  - Setup: the Pummeler in player 0's Reserves (its deploy zone).
  - Gain 1: the Pummeler itself has 5; player 0's Horror unit (power 3, `Subtypes: ["Horror"]`) in that Reserves has 4; a unit with subtypes `["Treefolk", "Horror"]` there has +1 (any one subtype matches).
  - Unchanged: player 0's non-Horror unit in that Reserves; player 0's Horror structure there (not a unit); player 0's Horror unit in the Main Theater; player 1's Horror unit in player 1's Reserves.
- `TestCard_MelodicPummeler_FollowsItsZone` (`SVY-R44`):
  - After `s.moveTo(pummeler, ZoneMainTheater)`: player 0's Main Theater Horror unit has +1, and its Reserves Horror unit has its printed power.
  - The Pummeler still has 5.
- `TestDefs_MatchRulesText` already checks `Text` against the CSV.
- Not this task: combat damage, the source leaving or dying, a face-down source, stacking with Layer 5 and the clamp, a control change, and the sweep's last-known record. T-0256 pinned each on synthetic sources in `aura_test.go`, and the cards add nothing to them.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- none. The tests use existing helpers only. Helpers local to the new file live in it.

## Allowed Files
- `pkg/engine/cards/animus.go`
- `pkg/engine/cards/cards.go`

## Done When
- [x] The listed tests pass: `go test ./pkg/engine -run "AnimocepsStalk|AnimocepsNode|MelodicPummeler|TestDefs"`
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks
- [x] `go tool task sim-compare -- autoloop/m2-wave-h` passes (none of the three is in a starter or stress deck)

## Log
- 2026-11-16 lead-architect (wave H run, session 17): queued with T-0256, split off under the D2 size rule. Written in full once T-0256 is `done`.
- 2026-11-16 lead-architect (wave H run, session 18): T-0256 is `done` (merge 327c9ec) with the names it planned (`StaticAura`, `Static.Filter`, `Static.Power`, `Game.auras`).
  - Wrote Inputs, Shape, Tests and Allowed Files against the run branch at 28166f6 and dry-ran the definitions (above).
  - Sync: merged `main` (7d6a309, the `pkg/engine` struct diagram) into the run branch first, with no conflicts.
  - The previous session (`20261116-113258-6`) hit a usage limit before it wrote anything. Its clean run-branch worktree was removed here.
  - Directives: none `open`. I checked each `waiting` one, and none has fired:
    - #6, #7: mode is `serial`.
    - #15 (2): no draw or end-of-turn hook here.
    - #71: no each-player effect.
    - #91 (2), #106, #120, #142: they name cards or shapes this task doesn't touch.
  - No new shape: each card is T-0256's `StaticAura` with its own filter.
  - Re-read `BOARD.md` before committing: no directive added since.
  - Status `tests`.
- 2026-11-16 rules-test-author: wrote `pkg/engine/cards_animus_auras_test.go`, status `tests-review`.
  - Tests: `TestCard_AnimocepsStalk_OtherAlliedAssetsInItsZone`, `TestCard_AnimocepsStalk_TwoStalksBuffEachOther`, `TestCard_AnimocepsStalk_FollowsItsZone`, `TestCard_AnimocepsNode_AlliedUnitsInItsZone`, `TestCard_AnimocepsNode_FollowsItsZone`, `TestCard_MelodicPummeler_AlliedHorrorUnitsInItsZone`, `TestCard_MelodicPummeler_FollowsItsZone`.
  - Helpers requested: none (local `auraSub` lives in the new file; `wantPower` reused).
  - Left out: nothing beyond the task's own exclusions. `go vet` is clean, `cardtown lint rules` passes. Node, Pummeler tests fail now and the Stalk ones partly pass or fail until the cards are defined.
- 2026-11-16 lead-architect (D4): tests returned, round 1. Every other assertion checks what the cited rules and survey rows give. One note:
  1. `TestCard_AnimocepsNode_AlliedUnitsInItsZone` expects the face-down Scratch Company to have `scratch.Inst.Card.Power+1`, which is the hidden card's printed power. A face-down asset has the Scratch Company face's characteristics (`FR-ZON-005`): power 1. So the expected value is the literal `2`, with the reason "a face-down Scratch Company (power 1) is a unit". The test passes today only because `putScratch`'s hidden filler also has power 1. Change only that line, in a new `test(T-0257)` commit.
  - Status `tests`.
- 2026-11-16 rules-test-author: D4 round 1, note 1 fixed. In `TestCard_AnimocepsNode_AlliedUnitsInItsZone` the Scratch Company's expected power is now the literal 2 (face-down characteristics, power 1, plus the aura; FR-ZON-005). No other change. Status `tests-review`.
- 2026-11-16 lead-architect (D4): tests approved, round 2.
  - Note 1 is fixed in its own `test(T-0257)` commit (82124a9): the face-down Scratch Company's expected power is the literal 2 (`FR-ZON-005`). No other line changed.
  - Each test checks what the cited rules and survey rows give, on the real cards:
    - The Stalk buffs other allied assets of any kind in its current zone. It doesn't buff itself, its other-zone allies or opposing assets. Two Stalks buff each other.
    - The Node buffs allied units only (a Scratch Company too). It doesn't buff itself or an allied structure.
    - The Pummeler buffs allied Horror units in its zone, itself included, and any one matching subtype is enough. It doesn't buff non-Horror units, Horror structures, other-zone allies or opposing Horror units.
    - Each aura follows its source's current zone (`SVY-R44`, last sentence). Reserves are each player's own (`FR-ZON-001`).
  - No existing test edited.
  - Sanity check in a throwaway worktree: the seven tests fail without the cards and pass with the Shape's definitions, together with `TestDefs_MatchRulesText`. The test file is gofmt-clean.
  - Status `ready`, dispatched to `kernel-engineer` (flash).
- 2026-11-16 kernel-engineer (flash): implemented same-zone power auras Animoceps Stalk, Animoceps Node, Melodic Pummeler (commit 0ed4c870). Status `review`.
  - Added ability definitions in `pkg/engine/cards/animus.go` using existing `ability.StaticAura` (T-0256) and `ability.Filter` (`Allegiance: AllegianceAllied`, `ZoneRef: ZoneRefSame`, `Power: 1`):
    - `Animoceps Node`: `Kinds: KindUnit`
    - `Animoceps Stalk`: `Other: true`
    - `Melodic Pummeler`: `Kinds: KindUnit`, `Subtypes: []string{"Horror"}`
  - Removed all three cards from `NotYetExecutable` in `pkg/engine/cards/cards.go`.
  - Reused: `ability.StaticAura`, `ability.ZoneRefSame`, `ability.AllegianceAllied`, `ability.KindUnit`, and `Game.auras`/`Game.Power` layer 4 evaluation.
  - Verification results:
    - `go test ./pkg/engine -run "AnimocepsStalk|AnimocepsNode|MelodicPummeler|TestDefs"`: pass (1.12s)
    - `go tool task check`: pass (gofmt, vet, test across all packages, and cardtown lint --all)
    - `go tool task sim`: pass (200 games, P0 58.5%, P1 39.5%, draws 2.0%, 0 invariant violations)
    - `go tool task sim-compare -- autoloop/m2-wave-h`: [PASS] starter and stress sims match byte for byte, exit codes included
