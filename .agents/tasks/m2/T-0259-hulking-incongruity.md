---
id: T-0259
title: "Hulking Incongruity: allied Horror units in the same zone have Resistant 1 (wave H card, flash)"
kind: card
tier: flash
wave: H
status: in-progress
branch: m2/autoloop-wave-h/T-0259-hulking-incongruity
depends_on: [T-0258]
claims: [card:Hulking Incongruity]
cards: [Hulking Incongruity]
requirements: [FR-LAY-001, FR-KWD-002b, FR-REP-004, SVY-R44]
---

## Goal
*Hulking Incongruity* executes: "Allied Horror units in the same zone have Resistant 1." It is one `ability.StaticAura` (T-0256) with T-0258's `Static.Resistant` 1 and the filter *Melodic Pummeler* uses (allied units with subtype Horror, `ZoneRefSame`), Power 0. Survey `:165`: it is a Plant, Horror unit, so it includes itself. The card leaves `cards.NotYetExecutable`.

## Inputs
- Card text (`data/cards/units-and-structures-animus.csv:21`): Unit, Animus, subtypes Plant, Horror, Requisition 4, deploys to Reserves, power 2, Health 5. It is in no starter or stress deck.
- Survey (`docs/design/m2_starter_set_ability_survey.md`): its row (`:165`: Continuous, `FR-LAY-001`, G2, "includes itself"); gap G2 (`:48`: keyword grants go in Layer 3, per FRD 01).
- FRD 01 (`docs/frd/01_Game_Rules_FRD.md`):
  - `FR-LAY-001`: continuous auras are evaluated dynamically.
  - `FR-KWD-002b` (`:614`): Resistant N reduces each discrete instance of incoming damage by N.
  - `FR-REP-004` Stage 2: the target's total Resistant is subtracted.
  - `FR-TGT-001` `SVY-R44`, last sentence: a continuous ability uses its source's current zone.
- Code of the day (run branch at b18ceba):
  - `ability.StaticAura`, `Static.Filter`, `Static.Power` and `Static.Resistant` (`pkg/engine/ability/ability.go:56-84`).
  - `Game.Keywords` sums `Static.Resistant` over `Game.auras`, the one aura reader (`pkg/engine/layers.go`), and damage reads `g.Keywords` in Stage 2.
  - *Melodic Pummeler*'s definition (`pkg/engine/cards/animus.go:473-487`) has the same filter.
  - `NotYetExecutable` is at `pkg/engine/cards/cards.go:13`.
  - T-0258 pinned the shape on synthetic sources in `pkg/engine/aura_keywords_test.go`: stacking with printed Resistant and a second source, zone following, the source dying, a face-down source, combat damage and the sweep's last-known record. This task pins the real card's filter and one damage case.

## Shape (the lead's plan)
- `pkg/engine/cards/animus.go`: one definition, after *Fungal Evangelist* and before *Infested Apostle*. `Text` is the CSV line. It has one `Statics` entry:
  `{Kind: ability.StaticAura, Filter: ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit, Subtypes: []string{"Horror"}, ZoneRef: ability.ZoneRefSame}, Resistant: 1}`.
  No `Power` (0). No new shape and no new field.
- `pkg/engine/cards/cards.go`: remove `"Hulking Incongruity"` from `NotYetExecutable`.
- The lead dry-ran this definition and the removal on the run branch at b18ceba (2026-11-16, wave H session 20):
  - `gofmt -l pkg cmd` clean.
  - `go tool task check` exit 0.
  - `go tool task sim` exit 0 (P0 58.5%, P1 39.5%, draws 2.0%).
  - `go tool task sim-compare -- autoloop/m2-wave-h` `[PASS]`.

## Tests (card tasks: written by rules-test-author; engineers never edit these)
Add the new tests to `pkg/engine/cards_animus_auras_test.go`, after the *Melodic Pummeler* tests, and add T-0259 to the file comment.
- Use the real card (`s.card("Hulking Incongruity")`), placed with `s.put` and moved with `s.moveTo`.
- Read effective Resistant with `s.wantResistant` (`aura_keywords_test.go`) and power with `s.wantPower`.
- Other assets are synthetic. `auraSub` (this file) gives a unit subtypes; `horrorUnit` (`aura_keywords_test.go`) gives a Horror unit printed Resistant.
- Player 0 is the active player.

Tests:
- `TestCard_HulkingIncongruity_AlliedHorrorUnitsInItsZone` (`FR-LAY-001`, survey "includes itself"):
  - Setup: the Incongruity in player 0's Reserves (its deploy zone).
  - Resistant 1: the Incongruity itself; player 0's Horror unit in that Reserves; a unit with subtypes `["Treefolk", "Horror"]` there (any one subtype matches).
  - Resistant 0: player 0's non-Horror unit in that Reserves; player 0's Horror structure there (not a unit); player 0's Horror unit in the Main Theater; player 1's Horror unit in player 1's Reserves.
  - The Incongruity keeps power 2, and the allied Horror unit keeps its printed power (the grant adds no power).
- `TestCard_HulkingIncongruity_AddsToPrintedResistant` (`FR-REP-004` Stage 2's total): player 0's Horror unit printing Resistant 2 in the Incongruity's zone has Resistant 3.
- `TestCard_HulkingIncongruity_FollowsItsZone` (`SVY-R44`, last sentence):
  - After `s.moveTo(incongruity, ZoneMainTheater)`: player 0's Main Theater Horror unit has Resistant 1, and its Reserves Horror unit has 0.
  - The Incongruity itself still has 1.
- `TestCard_HulkingIncongruity_CoveredUnitTakesOneLess` (`FR-KWD-002b`, `FR-REP-004`):
  - Setup: player 1's Incongruity and player 1's Horror unit D (no printed Resistant, Health 5) in the Main Theater. Player 0's Readied unit (power 3) is there too.
  - Player 0's unit attacks D. D takes 2 damage.
  - Use `s.do(s.attack(att, d))` and `s.wantDamage`, as `TestFR_REP_004_AuraResistantReducesDamage` does.
- `TestDefs_MatchRulesText` already checks `Text` against the CSV.
- Not this task: a second source, the source dying, a face-down source and the sweep's last-known record. T-0258 pinned each on synthetic sources in `aura_keywords_test.go`, and the card adds nothing to them.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- none. The tests use existing helpers only.

## Allowed Files
- `pkg/engine/cards/animus.go`
- `pkg/engine/cards/cards.go`

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run "HulkingIncongruity|TestDefs"`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-compare -- autoloop/m2-wave-h` passes (the card is in no starter or stress deck)

## Log
- 2026-11-16 lead-architect (wave H run, session 19): queued with T-0258, split off under the D2 size rule.
- 2026-11-16 lead-architect (wave H run, session 20): T-0258 is `done` (merge 5c519e8) with the names it planned (`Static.Resistant`, summed by `Game.Keywords` over `Game.auras`).
  - Wrote Inputs, Shape, Tests and Allowed Files against the run branch at b18ceba and dry-ran the definition (above).
  - `main` has no commits the run branch lacks, so no sync.
  - Directives: none `open`. I checked each `waiting` one, and none has fired:
    - #6, #7: mode is `serial`.
    - #15 (2): no draw or end-of-turn hook here.
    - #71: no each-player effect.
    - #91 (2), #106, #120, #142: they name cards or shapes this task doesn't touch.
  - No new shape: the card is T-0256's `StaticAura` with T-0258's `Static.Resistant` and *Melodic Pummeler*'s filter.
  - Re-read `BOARD.md` before committing: no directive added since.
  - Status `tests`.
- 2026-11-16 rules-test-author: wrote four tests in `pkg/engine/cards_animus_auras_test.go` (after the Pummeler tests; file comment updated): `TestCard_HulkingIncongruity_AlliedHorrorUnitsInItsZone`, `_AddsToPrintedResistant`, `_FollowsItsZone`, `_CoveredUnitTakesOneLess`.
  - Helpers requested: none. They vet clean and fail as expected (card has no aura yet). `cardtown lint rules`: no new `TestFR_*`, so nothing to trace.
  - Cases left out: none beyond the task's "Not this task" list.
  - Status `tests-review`.
- 2026-11-16 lead-architect (D4): tests approved, round 1.
  - Each test checks what the cited rules and the survey row give, on the real card:
    - The Incongruity grants Resistant 1 to allied Horror units in its zone, itself included, and any one matching subtype is enough (`FR-LAY-001`, survey `:165`).
    - It grants nothing to a non-Horror unit, a Horror structure, a Horror unit in another zone or an opposing Horror unit, and it adds no power.
    - The grant adds to printed Resistant (`FR-REP-004` Stage 2's total).
    - The grant follows the source's current zone (`SVY-R44`, last sentence).
    - A covered unit takes 3 − 1 = 2 combat damage (`FR-KWD-002b`).
  - No existing test edited. The file-comment change is the one the task asked for.
  - Sanity check in a throwaway worktree: the four tests fail without the card and pass with the Shape's definition, together with `TestDefs_MatchRulesText`. The test file is gofmt-clean.
  - Status `ready`, dispatched to `kernel-engineer` (flash).
