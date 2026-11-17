---
id: T-0263
title: "Ancestral Monument: \"Other allied assets in the same zone have Cornered: Resistant 1.\", on T-0262's Static.Cornered (wave H card, flash)"
kind: card
tier: flash
wave: H
status: ready
branch: m2/autoloop-wave-h/T-0263-ancestral-monument
depends_on: [T-0262]
claims: [card:Ancestral Monument]
cards: [Ancestral Monument]
requirements: [FR-LAY-001, FR-LAY-003, FR-KWD-002b, FR-KWD-002e, FR-REP-004, SVY-R18, SVY-R44]
---

## Goal
*Ancestral Monument* executes: *"Other allied assets in the same zone have Cornered: Resistant 1."* It is one `ability.StaticAura` (T-0256) with the filter "other allied assets in the same zone" (`Filter{Allegiance: AllegianceAllied, Other: true, ZoneRef: ZoneRefSame}`, zero `Kinds`: any asset) and T-0262's `Static.Cornered: ability.WhileCornered{Resistant: 1}`. Each matched asset has Resistant 1 only while its controller is Cornered (`FR-KWD-002e`, `SVY-R18`). The card leaves `cards.NotYetExecutable`.

## Inputs
- Card text (`data/cards/units-and-structures-path.csv:22`): Structure, Path, subtype Tower, Requisition 4, Tactics 0, deploys to the Main Theater, power 0, Health 9. It is in no starter or stress deck.
- Survey (`docs/design/m2_starter_set_ability_survey.md`): its row (`:205`: Continuous, `FR-LAY-001`, `FR-LAY-003`, G2); gap G2 (`:48`: keyword grants go in Layer 3); R18 (`:77`: a keyword behind Cornered exists only while Cornered).
- FRD 01 (`docs/frd/01_Game_Rules_FRD.md`):
  - `FR-KWD-002e` (`:617`): Cornered is active for a player whose opponent controls strictly more Main Theater assets, units and structures combined; `SVY-R18`.
  - `FR-LAY-003` (`:515`): Cornered is live, and locked at the start of a simultaneous damage step.
  - `FR-LAY-001`: continuous auras are evaluated dynamically.
  - `FR-KWD-002b` (`:614`) and `FR-REP-004` Stage 2: total Resistant reduces each instance of damage.
  - `FR-TGT-001` `SVY-R44`, last sentence: a continuous ability uses its source's current zone.
- Code of the day (run branch at 20f37da):
  - `ability.StaticAura`, `Static.Filter` and `Static.Cornered` (`pkg/engine/ability/ability.go`); `WhileCornered`.
  - `Game.whileCornered` sums each `g.auras(a)` aura's `Cornered` under the one gate `g.cornered(a.Controller)` (`pkg/engine/layers.go`); `Game.Keywords` adds it, and damage reads `g.Keywords` in Stage 2.
  - `NotYetExecutable` is at `pkg/engine/cards/cards.go:13`; `"Ancestral Monument"` is its first entry.
  - `pkg/engine/cards/path.go`: entries are alphabetical; *Ancestral Monument* goes after *Ambush Tactics* (`:70`) and before *Arkha, Sun's Fury* (`:146`).
  - T-0262 pinned the shape on synthetic sources in `pkg/engine/aura_cornered_test.go`: the live gate, the filter, summing with an asset's own Cornered and a second source, face-down sources and receivers, the power half, and the combat lock. This task pins the real card's filter and gate and one damage case.

## Shape (the lead's plan)
- `pkg/engine/cards/path.go`: one definition, after *Ambush Tactics* and before *Arkha, Sun's Fury*. `Text` is the CSV line. It has one `Statics` entry:
  `{Kind: ability.StaticAura, Filter: ability.Filter{Allegiance: ability.AllegianceAllied, Other: true, ZoneRef: ability.ZoneRefSame}, Cornered: ability.WhileCornered{Resistant: 1}}`.
  No `Power`, no `Resistant` (the grant is behind Cornered). No new shape and no new field.
- `pkg/engine/cards/cards.go`: remove `"Ancestral Monument"` from `NotYetExecutable`.
- The lead dry-ran this definition and the removal on the run branch at 20f37da (2026-11-16, wave H session 24):
  - `gofmt -l pkg cmd` clean.
  - `go tool task check` exit 0.
  - `go tool task sim` exit 0 (P0 58.5%, P1 39.5%, draws 2.0%).
  - `go tool task sim-compare -- autoloop/m2-wave-h` `[PASS]`.

## Tests (card tasks: written by rules-test-author; engineers never edit these)
New tests go in a new file `pkg/engine/cards_path_monument_test.go`, with a file comment naming the card, its CSV line and T-0263, as `cards_path_highlander_test.go` does.
- Use the real card (`s.card("Ancestral Monument")`), placed with `s.put` and moved with `s.moveTo`.
- Read effective Resistant with `s.wantResistant` (`aura_keywords_test.go`) and power with `s.wantPower`. Other assets are synthetic (`unit`, `structure`); `s.foes(n)` (`aura_cornered_test.go`) puts n player 1 units in the Main Theater.
- Count the Monument itself among its controller's Main Theater assets when building Cornered (`FR-KWD-002e`: units and structures combined). Player 0 is the active player.

Tests:
- `TestCard_AncestralMonument_OtherAlliedAssetsInItsZoneWhileCornered` (`FR-KWD-002e`, `FR-LAY-001`, `SVY-R18`):
  - Setup: player 0's Monument, allied unit U and allied structure St in the Main Theater (3 assets); allied unit R in player 0's Reserves; `s.foes(3)` (3 against 3: not Cornered).
  - Not Cornered: U and St have Resistant 0.
  - Add one foe (3 against 4: Cornered): U and St have Resistant 1; R has 0 (another zone); the Monument has 0 (`Other`); each foe has 0 (not allied).
  - Move one foe to player 1's Reserves (3 against 3): U and St have 0 again.
  - Throughout, the Monument keeps power 0 and U its printed power (the grant adds no power).
- `TestCard_AncestralMonument_AddsToPrintedResistant` (`FR-REP-004` Stage 2's total): player 0 Cornered; an allied unit printing Resistant 2 in the Monument's zone has Resistant 3.
- `TestCard_AncestralMonument_FollowsItsZone` (`SVY-R44`, last sentence):
  - Setup: player 0's Monument and allied unit U in the Main Theater, allied unit R in Reserves, `s.foes(3)` (Cornered).
  - After `s.moveTo(monument, ZoneReserves)` (1 against 3, still Cornered): R has Resistant 1, and U has 0.
- `TestCard_AncestralMonument_CorneredUnitTakesOneLess` (`FR-KWD-002b`, `FR-REP-004`, `FR-LAY-003`), two subtests:
  - `cornered`: player 1's Monument and player 1's unit D (power 0, Health 5) in the Main Theater; player 0's Readied attacker (power 3) and two more player 0 units there (3 against 2: player 1 Cornered). The attacker attacks D: D takes 2.
  - `not cornered`: the same with one more player 0 unit only (2 against 2): D takes 3.
  - Use `s.do(s.attack(att, d))` and `s.wantDamage`, as `TestCard_BrigandHighlander_CorneredAttackDealsFour` does.
- `TestDefs_MatchRulesText` already checks `Text` against the CSV.
- Not this task: a second source, the asset's own Cornered abilities, face-down sources and receivers, the power half, and the lock against a death in the same step. T-0262 pinned each on synthetic sources in `aura_cornered_test.go`, and the card adds nothing to them.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- none. The tests use existing helpers only.

## Allowed Files
- `pkg/engine/cards/path.go`
- `pkg/engine/cards/cards.go`

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run "AncestralMonument|TestDefs"`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-compare -- autoloop/m2-wave-h` passes (the card is in no starter or stress deck)

## Log
- 2026-11-16 lead-architect (wave H run, session 23): queued, split off from T-0262 under the D2 size rule.
- 2026-11-16 lead-architect (wave H run, session 24): T-0262 is `done` (merge 1904eca) with the names it planned (`Static.Cornered`, summed by `Game.whileCornered` over `Game.auras`).
  - Wrote Inputs, Shape, Tests and Allowed Files against the run branch at 20f37da and dry-ran the definition (above). `claims:` now names the card.
  - `main` has no commits the run branch lacks, so no sync.
  - Directives: none `open`. I checked each `waiting` one, and none has fired:
    - #6, #7: mode is `serial`.
    - #15 (2): no draw or end-of-turn hook here.
    - #71: no each-player effect.
    - #91 (2), #106, #120, #142: they name cards or shapes this task doesn't touch.
  - No new shape: the card is T-0256's `StaticAura` with T-0262's `Static.Cornered` and the "other allied assets in the same zone" filter.
  - Re-read `BOARD.md` before committing: no directive added since.
  - Status `tests`.
- 2026-11-16 rules-test-author: wrote `pkg/engine/cards_path_monument_test.go`, status `tests-review`.
  - Tests: `TestCard_AncestralMonument_OtherAlliedAssetsInItsZoneWhileCornered`, `TestCard_AncestralMonument_AddsToPrintedResistant`, `TestCard_AncestralMonument_FollowsItsZone`, `TestCard_AncestralMonument_CorneredUnitTakesOneLess` (subtests `cornered`, `not cornered`).
  - Helpers requested: none. `gofmt` and `go vet` are clean, `cardtown lint rules` passes. The tests fail now as expected (the card is not yet executable; the `not cornered` subtest passes).
  - Cases left out: none beyond those the task lists as not this task.
- 2026-11-16 lead-architect (D4): tests approved, round 1.
  - Each test checks what the cited rules and the survey row give, on the real card:
    - The Monument grants Resistant 1 to an allied unit and an allied structure in its zone only while player 0 is Cornered, counting the Monument itself among player 0's Main Theater assets (3 against 3, 3 against 4, back to 3 against 3; `FR-KWD-002e`, `SVY-R18`).
    - It grants nothing to itself (`Other`), to an allied unit in another zone or to a foe, and it adds no power (`FR-LAY-001`).
    - The grant adds to printed Resistant (`FR-REP-004` Stage 2's total).
    - The grant follows the source's current zone (`SVY-R44`, last sentence).
    - A covered unit takes 3 − 1 = 2 combat damage while its controller is Cornered, and 3 while not (`FR-KWD-002b`, `FR-LAY-003`).
  - No existing test edited; one new file.
  - Sanity check in a throwaway worktree: the tests fail without the card (the `not cornered` subtest passes either way, as it should) and pass with the Shape's definition, together with `TestDefs_MatchRulesText`. The test file is gofmt-clean.
  - Status `ready`, dispatched to `kernel-engineer` (flash).
