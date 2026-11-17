---
id: T-0261
title: "Brigand Highlander: \"Cornered: This unit has +2 power.\", on T-0260's `WhileCornered.Power` (wave H card, flash)"
kind: card
tier: flash
wave: H
status: review
branch: m2/autoloop-wave-h/T-0261-brigand-highlander
depends_on: [T-0260]
claims: [card:Brigand Highlander]
cards: [Brigand Highlander]
requirements: [FR-KWD-002e, FR-LAY-001, FR-LAY-003]
---

## Goal
*Brigand Highlander* (Path unit, Officer, Hammerhand, R1 T1, Main Theater, 2/3, "Cornered: This unit has +2 power.") loads with no unsupported lines and has power 4 while its controller is Cornered, through T-0260's `ability.WhileCornered{Power: 2}` on `Def.Cornered`. It leaves `cards.NotYetExecutable`.

## Inputs
- Card text (`data/cards/units-and-structures-path.csv:2`): Unit, Path, subtypes Officer, Hammerhand, Requisition 1, Tactics 1, deploys to the Main Theater, power 2, Health 3. It is in no starter or stress deck.
- Survey (`docs/design/m2_starter_set_ability_survey.md:203`): wave E, Continuous (Cornered), "self +2 power while Cornered", `FR-KWD-002e`, `FR-LAY-003`.
- FRD 01 (`docs/frd/01_Game_Rules_FRD.md`):
  - `FR-KWD-002e` (`:617`): a player is Cornered while the opponent controls strictly more battlefield assets in the Main Theater than they do there; for continuous effects in a simultaneous combat or damage step, Cornered is locked at the start of the step.
  - `FR-LAY-001`: a static power bonus is Layer 4; the clamp at 0 is Layer 6.
  - `FR-LAY-003`: the lock for a simultaneous step.
- Code of the day (run branch at 78b4e69):
  - `ability.WhileCornered{Aggressive, Resistant, Power}` and `Def.Cornered` (`pkg/engine/ability/ability.go:34-37`, `:95-104`).
  - `Game.Power` adds `WhileCornered.Power` in Layer 4 through `Game.whileCornered`, the one reader of the gate (`pkg/engine/layers.go:143`).
  - The other `Def.Cornered` cards: *Allshan, Chainbreaker Revenant* (`pkg/engine/cards/path.go:68`) and *Macabre Dancers* (`:640`).
  - `NotYetExecutable` is at `pkg/engine/cards/cards.go:13`.
  - T-0260 pinned the shape on synthetic cards (`corneredCard`) in `pkg/engine/cornered_power_test.go`: the gate turning on and off, a unit in Reserves following its controller, the clamp, adding with an aura, the combat lock, the sweep's last-known power and a face-down asset. This task pins the real card and one combat case.

## Shape (the lead's plan)
- `pkg/engine/cards/path.go`: one definition, after *Boneyard Scavenger* and before *Broken Horn Armorer*:
  ```go
  "Brigand Highlander": {
  	Text:     []string{"Cornered: This unit has +2 power."},
  	Cornered: ability.WhileCornered{Power: 2},
  },
  ```
  No new shape and no new field.
- `pkg/engine/cards/cards.go`: remove `"Brigand Highlander"` from `NotYetExecutable`.
- The lead dry-ran this definition and the removal on the run branch at 78b4e69 (2026-11-16, wave H session 22):
  - `gofmt -l pkg cmd` clean.
  - `go tool task check` exit 0.
  - `go tool task sim` exit 0 (P0 58.5%, P1 39.5%, draws 2.0%).
  - `go tool task sim-compare -- autoloop/m2-wave-h` `[PASS]`.

## Tests (card tasks: written by rules-test-author; engineers never edit these)
Add a new file `pkg/engine/cards_path_highlander_test.go`, with a file comment naming the card, its CSV text, T-0261 and the rules it cites (as `cards_path_dancers_test.go` does).
- Use the real card (`s.card("Brigand Highlander")`), placed with `s.put` and moved with `s.moveTo`.
- Read power with `s.wantPower` (`aura_test.go`). Other assets are synthetic `unit(...)` cards.
- Player 0 is the active player.

Tests:
- `TestCard_BrigandHighlander_PlusTwoPowerOnlyWhileCornered` (`FR-KWD-002e`, `FR-LAY-001`):
  - Setup: player 0's Highlander in the Main Theater, and one player 1 unit there.
  - Power 2: 1 against 1 is not Cornered.
  - Power 4: after a second player 1 unit enters the Main Theater (1 against 2).
  - Power 2 again: after that second unit moves to player 1's Reserves.
  - `s.g.Keywords` of the Highlander equals its printed keywords, both while Cornered and not (the line grants power only).
- `TestCard_BrigandHighlander_CorneredAttackDealsFour` (`FR-KWD-002e`, `FR-LAY-003`, `FR-CBT-002`):
  - Setup: player 0's Readied Highlander in the Main Theater; player 1's defender (power 0, Health 20) and a second player 1 unit (power 0) there, so player 0 is Cornered.
  - The Highlander attacks the defender (`s.do(s.attack(hl, def))`). The defender has 4 damage (`s.wantDamage`).
- `TestCard_BrigandHighlander_NotCorneredAttackDealsTwo` (`FR-KWD-002e`):
  - The same setup without the second player 1 unit (1 against 1). The defender has 2 damage.
- `TestDefs_MatchRulesText` already checks `Text` against the CSV.
- Not this task: a unit in Reserves, the clamp, an aura, the gate changing mid-combat, the sweep's last-known power and a face-down asset. T-0260 pinned each on synthetic cards in `cornered_power_test.go`, and the card adds nothing to them.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- none. The tests use existing helpers only.

## Allowed Files
- `pkg/engine/cards/path.go`
- `pkg/engine/cards/cards.go`

## Done When
- [x] The listed tests pass: `go test ./pkg/engine -run "BrigandHighlander|TestDefs"`
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks
- [x] `go tool task sim-compare -- autoloop/m2-wave-h` passes (the card is in no starter or stress deck)

## Log
- 2026-11-16 lead-architect (wave H run, session 21): queued with T-0260, split off under the D2 size rule.
- 2026-11-16 lead-architect (wave H run, session 22): T-0260 is `done` (merge 05c6b53) with the names it planned (`WhileCornered.Power`, read by `Game.Power` through `Game.whileCornered`).
  - Wrote Inputs, Shape, Tests and Allowed Files against the run branch at 78b4e69 and dry-ran the definition (above).
  - `main` has no commits the run branch lacks, so no sync.
  - Directives: none `open`. I checked each `waiting` one, and none has fired:
    - #6, #7: mode is `serial`.
    - #15 (2): no draw or end-of-turn hook here.
    - #71: no each-player effect.
    - #91 (2), #106, #120, #142: they name cards or shapes this task doesn't touch.
  - No new shape: the card is T-0260's `WhileCornered.Power`, the third `Def.Cornered` card after *Allshan* and *Macabre Dancers*.
  - Re-read `BOARD.md` before committing: no directive added since.
  - Status `tests`.
- 2026-11-16 rules-test-author: wrote `pkg/engine/cards_path_highlander_test.go`: `TestCard_BrigandHighlander_PlusTwoPowerOnlyWhileCornered`, `TestCard_BrigandHighlander_CorneredAttackDealsFour`, `TestCard_BrigandHighlander_NotCorneredAttackDealsTwo`.
  - Helpers requested: none. Cases left out: those the task assigns to T-0260 (Reserves, clamp, aura, mid-combat gate, sweep, face-down).
  - Compiles; the Cornered power and attack-four tests fail until the card definition exists (Highlander is NotYetExecutable); `cardtown lint rules` passes. Status `tests-review`.
- 2026-11-16 lead-architect (D4): tests approved, round 1.
  - Each test checks what the cited rules and the survey row give, on the real card:
    - Power 2 at 1 against 1, 4 at 1 against 2, and 2 again once the second enemy leaves the Main Theater (`FR-KWD-002e`, `FR-LAY-001`).
    - The line grants no keywords, Cornered or not.
    - A Cornered attack deals 2 + 2 = 4, and an attack at 1 against 1 deals 2 (`FR-LAY-003`, `FR-CBT-002`).
  - `highlanderCombat` is setup only, local to the test file. No existing test edited.
  - Sanity check in the review worktree: with the Shape's definition and removal applied (then reverted), the three tests and `TestDefs_MatchRulesText` pass. The test file is gofmt-clean.
  - Status `ready`, dispatched to `kernel-engineer` (flash).
- 2026-11-16 kernel-engineer (flash): implemented Brigand Highlander (commit 6678e7d6).
  - Reused existing shape: `ability.WhileCornered{Power: 2}` from T-0260 on `Def.Cornered`. No new shapes or fields added.
  - Diff:
    - `pkg/engine/cards/path.go`: added `"Brigand Highlander"` with `Text: []string{"Cornered: This unit has +2 power."}` and `Cornered: ability.WhileCornered{Power: 2}` between `Boneyard Scavenger` and `Broken Horn Armorer`.
    - `pkg/engine/cards/cards.go`: removed `"Brigand Highlander"` from `NotYetExecutable`.
  - Commands and results:
    - `go test ./pkg/engine -run "BrigandHighlander|TestDefs"`: pass (ok animus-mundi/pkg/engine 0.946s).
    - `go tool task check`: pass (gofmt, go vet, all unit/integration tests, all cardtown linters pass).
    - `go tool task sim`: pass (P0 wins 58.5%, P1 wins 39.5%, draws 2.0%).
    - `go tool task sim-compare -- autoloop/m2-wave-h`: pass (`[PASS] starter and stress sims match autoloop/m2-wave-h byte for byte, exit codes included`).
  - Status `review`.
