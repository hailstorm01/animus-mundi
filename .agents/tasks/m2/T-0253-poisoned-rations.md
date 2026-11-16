---
id: T-0253
title: "Poisoned Rations: chosen unit gets -3 max health until end of turn (wave H card, flash)"
kind: card
tier: flash
wave: H
status: review
branch: m2/autoloop-wave-h/T-0253-poisoned-rations
depends_on: [T-0251, T-0252]
claims: [card:Poisoned Rations]
cards: [Poisoned Rations]
requirements: [FR-LAY-001, FR-LAY-002, FR-LAY-003, FR-TRN-001, FR-MNV-001]
---

## Goal
*Poisoned Rations* (Path Maneuver, `data/cards/maneuvers-path.csv:10`; R0 T3; in neither starter deck): "Chosen unit gets -3 max health until end of turn." becomes executable from shapes that are `done`: one unit target of any allegiance and `OpModifyStat` with T-0252's `Stat: ability.StatMaxHealth`, `Amount` -3. The card leaves `cards.NotYetExecutable`.

## Inputs
- Card text: `data/cards/maneuvers-path.csv:10`. Survey row: `docs/design/m2_starter_set_ability_survey.md:197`.
- FRD 01 (`docs/frd/01_Game_Rules_FRD.md`): `FR-LAY-002` (line 506): Current Health & Lethal Destruction (`:507`, which names *Poisoned Rations*: Effective Maximum Health 0 or less is lethal at the sweep), State-Based Death Sweep Timing (`:508`), Temporary Maximum Health Modifiers (`:509`: no damage counters; the expiry in Sub-Phase 4c restores Maximum Health and so Current Health); `FR-LAY-001` (line 500); `FR-LAY-003` (line 512); `FR-TRN-001` (line 301, Sub-Phase 4c); `FR-TRN-003a`/`FR-MNV-001` (line 372: a Maneuver with no legal target can't be enacted).
- Code of the day (run branch at fa1edd3): `ability.StatMaxHealth` (`pkg/engine/ability/ability.go:296`), `Step.Stat` (`:416`), `OpModifyStat`; `stepModifyStat`'s max-health case (`pkg/engine/effects.go:276`); `Modifier.MaxHealth`, `Game.MaxHealth`, `Game.CurrentHealth` (`pkg/engine/layers.go`); `NotYetExecutable` (`pkg/engine/cards/cards.go:13`). T-0251 (merge 6676322) and T-0252 (merge 9cde374) pinned the modifier, its sweep, its 4c expiry and the step on synthetic Maneuvers in their test files (`pkg/engine/modify_max_health_test.go` for the step); this task pins them on the real card.

## Shape (the lead's plan)
- `pkg/engine/cards/path.go`: the definition, in alphabetical order (after *Pincer Tactics*, before *Reaver's Rush*). `Text` is the CSV line. `Effect` has one target `{Filter: ability.Filter{Kinds: ability.KindUnit}}` (any allegiance: "chosen unit") and one step `{Op: ability.OpModifyStat, Stat: ability.StatMaxHealth, Amount: ability.Value{N: -3}, Subject: ability.SubjectTarget, Target: 0}` (the zero `Until` is end of turn). No new shape, no new field.
- `pkg/engine/cards/cards.go`: remove "Poisoned Rations" from `NotYetExecutable`.
- The lead dry-ran the definition and the removal on the run branch at fa1edd3 (2026-11-16, wave H session 15): `gofmt -l pkg cmd` clean, `go tool task check` exit 0, `go tool task sim` exit 0, `go tool task sim-compare -- autoloop/m2-wave-h` `[PASS]` (byte-identical: the card is in neither starter deck).

## Tests (card tasks: written by rules-test-author; engineers never edit these)
New tests go in `pkg/engine/cards_path_poisoned_rations_test.go`. Use the real card (`s.card("Poisoned Rations")`), given to player 0 (`s.give`) with exactly R0 T3 (`s.setPool(0, 0, 3)`), enacted with `s.do(enactAction(inst))` in player 0's Main Phase. Read health with `s.g.MaxHealth`, `s.g.CurrentHealth` and `a.Damage` (`mhWant` in `modify_max_health_test.go:25` fits). Each scene has exactly one unit on the battlefield unless it says otherwise, so the choice is forced; where there are two, choose with `chooseAsset`.
- `TestCard_PoisonedRations_KillsHealthThreeWithoutDamage` (`FR-LAY-002` Current Health & Lethal Destruction): an enemy unit with Health 3 and no damage in the Main Theater. After *Poisoned Rations* resolves the unit is off the battlefield and its card is in player 1's discard pile, and no `damage` event names it (Maximum Health 0 is lethal with no damage).
- `TestCard_PoisonedRations_KillsDamagedUnit` (`FR-LAY-002`): an enemy unit with Health 5 and `Damage` 2: destroyed after resolution (Maximum Health 2, Current Health 0).
- `TestCard_PoisonedRations_SurvivorRestoredAtEndOfTurn` (`FR-LAY-002` Temporary Maximum Health Modifiers, `FR-LAY-003`, `FR-TRN-001` 4c): an enemy unit with Health 5 and no damage: alive with Maximum Health 2, Current Health 2, `Damage` 0 after resolution. After `s.endTurn()` it is alive with Maximum Health 5, Current Health 5, `Damage` 0. In a second scene, a Health 6 unit with `Damage` 1: alive with Maximum Health 3, Current Health 2 after resolution, and Maximum Health 6, Current Health 5, `Damage` 1 after `s.endTurn()` (the damage is unchanged by the expiry).
- `TestCard_PoisonedRations_AlliedUnitIsATarget` (`FR-TGT-001`, "chosen unit"): an allied unit with Health 5 and an enemy unit with Health 5 in the Main Theater. A target decision is asked and its options hold both units; choosing the allied unit (`chooseAsset`) gives it Maximum Health 2 and leaves the enemy unit at 5.
- `TestCard_PoisonedRations_NotEnactableWithoutUnit` (`FR-MNV-001`), table-driven like `armorNotEnactable` (`pkg/engine/cards_path_maneuvers_test.go:558`), written locally (don't edit that helper): no unit at all → not enactable; only structures (one allied, one opposing) → not enactable; one opposing unit in player 1's Reserves → enactable; one allied unit in player 0's Reserves → enactable. A refused enactment leaves *Poisoned Rations* in player 0's hand.
- `TestDefs_MatchRulesText` already checks `Text` against the CSV.
- Not this task: a target that leaves before resolution (`FR-TGT-003`; the second target check is the code every targeted Maneuver goes through), and the stat filter on Maximum Health (T-0252 pinned it; no M2 card reads it yet).

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- none. The tests use existing helpers only (`enactAction`, `offersEnact`, `inHand`, `chooseAsset`, `setPool`, `mhWant`, `effectEvents`).

## Allowed Files
- `pkg/engine/cards/path.go`
- `pkg/engine/cards/cards.go`

## Done When
- [x] The listed tests pass: `go test ./pkg/engine -run "PoisonedRations|TestDefs"`
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks
- [x] `go tool task sim-compare -- autoloop/m2-wave-h` passes (*Poisoned Rations* is in neither starter deck)

## Log
- 2026-11-15 lead-architect (wave H run, session 12): queued, split off T-0251 under the D2 size rule. Waits for T-0251 and T-0252.
- 2026-11-16 lead-architect (wave H run, session 15): T-0251 and T-0252 are `done` (merges 6676322, 9cde374), and the names T-0252 merged (`Step.Stat`, `ability.StatMaxHealth`) are the ones planned. Wrote Inputs, Shape, Tests and Allowed Files against the run branch at fa1edd3 and dry-ran the definition (above). No sync needed (`main` has nothing the run branch lacks). Directives: none `open`; the `waiting` ones were checked and none has fired (#15 (2): no draw or end-of-turn hook here; #71: "chosen unit" is not an each-player effect; #91 (2), #106, #120, #141 (3) aura half and #142 name cards or shapes this task doesn't touch). No new shape: the step is T-0252's. Re-read `BOARD.md` before committing: no directive added since. Status `tests`.
- 2026-11-16 rules-test-author: wrote `pkg/engine/cards_path_poisoned_rations_test.go` (commit test(T-0253)): TestCard_PoisonedRations_KillsHealthThreeWithoutDamage, _KillsDamagedUnit, _SurvivorRestoredAtEndOfTurn (both scenes), _AlliedUnitIsATarget, _NotEnactableWithoutUnit (four cases). Local helper `poisonedRations(s)` (gives the card, sets R0 T3). Helpers requested: none. Cases left out: none beyond the task's "Not this task" list. Tests compile, `cardtown lint rules` passes, and they fail on "needs rules the kernel does not implement yet" until the definition lands. Status `tests-review`.
- 2026-11-16 lead-architect (D4): tests approved, round 1. Each checks what `FR-LAY-002` (`:507`, which names *Poisoned Rations*: Maximum Health 0 is lethal at the sweep with no damage; `:509`: no damage counters, restored in 4c) and `FR-MNV-001` give, on the real card: Health 3 dies with no `damage` event and goes to its owner's discard; Health 5 with 2 damage dies; Health 5 survives at 2/2/0 and is 5/5/0 after the turn ends; Health 6 with 1 damage is 3/2/1, then 6/5/1; both an allied and an enemy unit are offered and only the chosen one changes; with no unit (none at all, or only structures) it isn't offered or enactable and stays in hand, and one unit of either side in Reserves makes it enactable. `chooseAsset` in the one-unit scenes is harmless when no question is asked. No existing test edited; `mhWant`, `mhEventsNaming` and `targetOffers` reused. Status `ready`, dispatched to `kernel-engineer` (flash).
- 2026-11-16 kernel-engineer (flash): defined Poisoned Rations in pkg/engine/cards/path.go with Target {Filter: {Kinds: KindUnit}} and Step {Op: OpModifyStat, Stat: StatMaxHealth, Amount: Value{N: -3}, Subject: SubjectTarget, Target: 0} in alphabetical order between Pincer Tactics and Reaver's Rush. Removed "Poisoned Rations" from NotYetExecutable in pkg/engine/cards/cards.go. Reused existing shapes (OpModifyStat with StatMaxHealth, KindUnit filter). Ran: go test ./pkg/engine -run "PoisonedRations|TestDefs" (pass); go tool task check (pass); go tool task sim (pass, 200 games); go tool task sim-compare -- autoloop/m2-wave-h (pass, starter and stress sims match autoloop/m2-wave-h byte for byte). Commit 9a31c2a5 feat(engine): define Poisoned Rations maneuver (T-0253). Status review.
