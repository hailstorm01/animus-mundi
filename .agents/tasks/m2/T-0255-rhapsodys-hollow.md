---
id: T-0255
title: "Rhapsody's Hollow: T3, exhaust, kill this structure: gain control of chosen opposing unit in the Main Theater (wave H card, flash)"
kind: card
tier: flash
wave: H
status: tests
branch: m2/autoloop-wave-h/T-0255-rhapsodys-hollow
depends_on: [T-0254]
claims: [card:Rhapsody's Hollow]
cards: [Rhapsody's Hollow]
requirements: [FR-ACT-001, FR-ACT-002, FR-ZON-002, FR-ZON-003, FR-TRN-003a, FR-TGT-003]
---

## Goal
*Rhapsody's Hollow* (Animus structure, `data/cards/units-and-structures-animus.csv:9`; in neither starter deck): "T3, exhaust, kill this structure: Gain control of chosen opposing unit in the Main Theater." becomes executable from shapes that are `done`: an activated ability with costs `CostResources` T3, `CostExhaustSelf` and `CostKillSelf` (as *Sinuous Growth*, `pkg/engine/cards/animus.go:537`), one target `{Allegiance: Opposing, Kinds: Unit, Zone: Main Theater}`, and T-0254's `OpGainControl` on it. The card leaves `cards.NotYetExecutable`.

## Inputs
- Card text: `data/cards/units-and-structures-animus.csv:9` (Structure, R2 T0, Main Theater, Health 5). Survey: wave D row (`docs/design/m2_starter_set_ability_survey.md:32`), card row (`:158`), `SVY-R36` (`:95`), `SVY-R8` (`:67`: an activated ability can't be used while an Attack is on the stack, so no pending-attack case here).
- FRD 01 (`docs/frd/01_Game_Rules_FRD.md`): `FR-ACT-001` & `FR-ACT-002` (line 367: costs paid on activation, `SVY-R36` required targets, `SVY-R2` targets and costs committed together, line 120); `FR-ZON-002` (line 250: Control Transfer Scope & Spatial Invariance `:254`, Destruction `:255`); `FR-ZON-003` (line 226, Ownership vs. Control `:229`); `FR-TGT-003` (line 411); `FR-TRN-003a` (line 372).
- Code of the day (run branch at 49a6e82): T-0254 merged the names it planned: `ability.OpGainControl`, `stepGainControl` (`pkg/engine/effects.go`, skips non-units, assets outside the Main Theater and assets the controller already controls), `EventControl`. The costs are *Sinuous Growth*'s and *Cliff of Portents*' (`pkg/engine/cards/animus.go:537`, `:61`): `CostResources`, `CostExhaustSelf`, `CostKillSelf`, paid in printed order by `payCosts` (`pkg/engine/costs.go:197`) after the targets are chosen. `NotYetExecutable` (`pkg/engine/cards/cards.go:13`). T-0254 pinned the step on synthetic sources in `pkg/engine/gain_control_test.go`; this task pins it on the real card.

## Shape (the lead's plan)
- `pkg/engine/cards/animus.go`: the definition, in alphabetical order (after *Rhapsody's Embrace*, before *Shifting Impossibility*). `Text` is the CSV line. One `Activated` with `Costs` `{CostResources, Tac: 3}`, `{CostExhaustSelf}`, `{CostKillSelf}` in printed order, and an `Effect` with one target `{Filter: {Allegiance: ability.AllegianceOpposing, Kinds: ability.KindUnit, Zone: ability.ZoneMainTheater}}` and one step `{Op: ability.OpGainControl, Subject: ability.SubjectTarget, Target: 0}`. No new shape, no new field.
- `pkg/engine/cards/cards.go`: remove "Rhapsody's Hollow" from `NotYetExecutable`.
- The lead dry-ran this definition and the removal on the run branch at 49a6e82 (2026-11-16, wave H session 16): `gofmt -l pkg cmd` clean, `go tool task check` exit 0, `go tool task sim` exit 0, `go tool task sim-compare -- autoloop/m2-wave-h` `[PASS]` (byte-identical: the card is in neither starter deck).

## Tests (card tasks: written by rules-test-author; engineers never edit these)
New tests go in `pkg/engine/cards_animus_rhapsodys_hollow_test.go`. Use the real card (`s.card("Rhapsody's Hollow")`), put Readied in player 0's Main Theater (`s.put(0, ZoneMainTheater, ...)`) with exactly R0 T3 (`s.setPool(0, 0, 3)`), activated with `s.do(activate(h, 0))` in player 0's Main Phase. Use `offersActivate` and `actNotOffered` (`pkg/engine/activated_test.go:45`, `:515`), `inDiscardPile`, `s.alive`, `effectEvents`. Each scene has exactly one legal target unless it says otherwise.
- `TestCard_RhapsodysHollow_GainsControlOfOpposingUnit` (`FR-ZON-002` `:254`, `FR-ZON-003`, `FR-ACT-001`): player 1's Readied unit (power 3, Health 5) in the Main Theater with `Damage` 1. After activation: the same `*Asset` is in `g.Theater` with `Controller` 0, `Inst.Owner` 1, Readied, `Damage` 1; one `control` event names it; the Hollow is off the battlefield and its card is in player 0's discard pile; player 0's pool is R0 T0 (exactly T3 paid).
- `TestCard_RhapsodysHollow_OnlyOpposingMainTheaterUnitsOffered` (`FR-TGT-001`): player 0 has an allied unit in the Main Theater; player 1 has a unit in its Reserves, a structure in the Main Theater, and two units in the Main Theater. One target question is asked, and its options are exactly player 1's two Main Theater units; the one chosen changes controller and the other doesn't.
- `TestCard_RhapsodysHollow_NotOfferedWithoutTarget` (`SVY-R36`, `FR-ACT-001`), table-driven with `actNotOffered` (which checks nothing was paid and the Hollow stays): no other asset; only an allied unit in the Main Theater; only player 1's unit in its Reserves; only player 1's structure in the Main Theater.
- `TestCard_RhapsodysHollow_CostsNeedTacticsAndReadied` (`FR-ACT-001`, `FR-ACT-002`): with one opposing Main Theater unit, two scenes, each with `actNotOffered`: the pool is R5 T2 (Requisition can't pay Tactics); the Hollow is Exhausted with T3.
- `TestCard_RhapsodysHollow_TargetLeftMainTheaterBeforeResolution` (`FR-TGT-003`, `FR-ACT-001`): the target is moved to player 1's Reserves after the item is placed and before it resolves (the `recorder` on the `place` event, as `TestCard_AllshansBattleTrance_TargetLeftThePile` does, `pkg/engine/cards_path_battle_trance_test.go:188`). The unit stays in `Players[1].Reserves` with `Controller` 1, no `control` event is logged, and the costs stay paid: the Hollow is in player 0's discard pile and the pool is T0.
- `TestDefs_MatchRulesText` already checks `Text` against the CSV.
- Not this task: what a gained unit may do afterwards (March, attack, readying, Cornered, destruction to its owner's discard): T-0254 pinned each on the step, and the card adds nothing to them.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- none. The tests use existing helpers only.

## Allowed Files
- `pkg/engine/cards/animus.go`
- `pkg/engine/cards/cards.go`

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run "RhapsodysHollow|TestDefs"`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-compare -- autoloop/m2-wave-h` passes (*Rhapsody's Hollow* is in neither starter deck)

## Log
- 2026-11-16 lead-architect (wave H run, session 15): queued, split off T-0254 under the D2 size rule. Waits for T-0254.
- 2026-11-16 lead-architect (wave H run, session 16): T-0254 is `done` (merge 9178485) with the names planned (`OpGainControl`, `EventControl`). Wrote Inputs, Shape, Tests and Allowed Files against the run branch at 49a6e82 and dry-ran the definition (above). No sync needed (`main` is already in the run branch). Directives: none `open`; the `waiting` ones were checked and none has fired (#6, #7: mode is `serial`; #15 (2): no draw or end-of-turn hook here; #71: "chosen opposing unit" is not an each-player effect; #91 (2), #106, #120, #141 (3) aura half and #142 name cards or shapes this task doesn't touch). No new shape: the costs are *Sinuous Growth*'s and *Cliff of Portents*', the step is T-0254's. Re-read `BOARD.md` before committing: no directive added since. Status `tests`.
