---
id: T-0210
title: Reaver's Rush, a free deploy of a chosen unit card with power 3 or less from your discard pile (wave F deck card, flash; #140 (4))
kind: card
tier: flash
wave: F
status: review
branch: m2/autoloop-wave-f/T-0210-reavers-rush
depends_on: [T-0208, T-0209]
claims: [card:Reaver's Rush]
cards: [Reaver's Rush]
requirements: [FR-TRN-003a, FR-MNV-001, FR-TGT-001, FR-MOD-001]
---

## Goal
*Reaver's Rush* (Path Maneuver, T3, two copies in `data/decks/path-starter.txt`): "Deploy chosen unit card with power 3 or less from your discard pile for free." One exactly-one pile target (allied, unit, power 3 or less, T-0209's criterion) and one `OpDeployFree` step on it (T-0208). It combines shapes only.

Covers:
- Its definition in `pkg/engine/cards/path.go`.
- `FR-MNV-001`'s own example: it can't be enacted without a unit card of power 3 or less in its controller's discard pile.
- Directive #140 (4) (`SVY-R43`): with *Devilhorn Guerrillas* (power 3, printing both zones) as the target, a timed-out zone choice puts it in Reserves, so its Main Theater deploy trigger doesn't fire; answering the Main Theater fires it.

## Inputs
- Card (`data/cards/maneuvers-path.csv` line 11): Path Maneuver, R0 T3, "Deploy chosen unit card with power 3 or less from your discard pile for free." *Devilhorn Guerrillas* (`data/cards/units-and-structures-path.csv` line 25): Path Unit, R4 T0, deploys to the Main Theater or Reserves, power 3, health 5, "When this unit deploys, if it was deployed to the Main Theater, deal 2 damage to chosen opposing unit." Its definition, a `HookDeploy` trigger with `ToZone: ZoneMainTheater`, is at [cards/path.go:338](../../../pkg/engine/cards/path.go).
- Survey row: `docs/design/m2_starter_set_ability_survey.md` line 203 ("`DEPLOY_FREE` chosen unit card with power ≤ 3 from discard"; "Zone timeout: Reserves (R43)"); `R43` at line 102.
- FRD 01 v3.23: `FR-TRN-003, FR-TRN-003a & FR-MNV-001` (line 356), above all "A Maneuver Needs Its Required Targets" (line 358, which names this card) and "In-Resolution Free Deployments" (line 360) with its item 1 (line 361: no zone named, so the printed zones; with both, the deploying player chooses, and a timeout enters Reserves). `FR-CST-001` (line 316: "for free" waives every cost). `FR-TGT-001` (line 370), "Readiness, Subtypes & Dynamic Stats" (line 375): a stat comparison. A card in a pile has no effective stats, so it is read by its printed power (`FR-ZON-004`, as T-0209 built it). `FR-MOD-001` §6, "Free-Deploy Zone (`SVY-R43`)" (line 416). `FR-TGT-003` (line 394).
- Rulings: `SVY-R29`, `SVY-R43`; directive #140 (4).
- Design note: `docs/architecture/m2-ability-model.md`, Key types (`Filter`, `Step`).
- Code: `ability.OpDeployFree` and `stepDeployFree` (T-0208); `Filter.Stat` and `Filter.AtMost` with `ability.StatPower` ([ability.go:225-237](../../../pkg/engine/ability/ability.go), T-0209); a pile target in the same form as *Boneyard Scavenger*'s ([cards/path.go:162-171](../../../pkg/engine/cards/path.go)).

### Definition (the engineer adds it to `pathDefs`, between *Pincer Tactics* and *Rogue Fleshshaper*)
- `Text`: the CSV text above, as one line.
- `Effect.Targets`: one exactly-one slot, `Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit, Zone: ability.ZoneDiscard, Stat: ability.StatPower, AtMost: 3}`.
- `Effect.Steps`: `{Op: ability.OpDeployFree, Subject: ability.SubjectTarget, Target: 0}` (no `To`: the card's printed zones, `FR-TRN-003a` 1).
- Remove the name from `cards.NotYetExecutable`.

A lead probe with this definition on the run branch at 8647386, since reverted, passes `go tool task check`, and `go tool task sim` stays clean: Path 42 → 44/50 main cards executable, P0 73.5% / P1 25.5% / draws 1.0% over 200 games. A probe test found Enact not offered with an empty pile, with only *Chainbreaker Phalanx* (power 4) or only *Devilhorn Sanction* (a Maneuver card) in it, and offered with *Devilhorn Guerrillas*. With the Guerrillas and one opposing unit in the Main Theater, a timed-out zone choice put the Guerrillas in Reserves and dealt no damage; answering the Main Theater dealt the opposing unit 2 damage. Each asked player 0 one decision. So `sim-compare` will differ here and isn't a gate.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
In a new file `pkg/engine/cards_path_reavers_rush_test.go`, from the CSV text, loading the real Maneuver (`s.card`, `enactFor`), in the style of `cards_path_maneuvers_test.go` and `deploy_free_test.go`. Player 0 enacts; put cards in a discard pile with `s.inDiscard`. Pile cards other than the Guerrillas and the Phalanx may be synthetic (`unit`, `structure`, `maneuver`, with power and `DeployZones` set per test), so no deploy trigger asks anything. Script only the decisions actually asked: a single legal pile target is chosen without a question (`FR-MOD-001` §1).
- `TestCard_ReaversRush_DeploysChosenUnitCardForFree` (`FR-TRN-003a`, `FR-CST-001`): a Reserves-only power-3 unit card costing R2 T1 in player 0's pile, player 0 with exactly R0 T3. After resolution it is an asset controlled by player 0 in player 0's Reserves and no longer in the pile; player 0's pool is R0 T0 (only the Maneuver's own T3 was paid); no `DecideZone` was asked; *Reaver's Rush* is in player 0's discard pile.
- `TestCard_ReaversRush_OffersOnlyUnitCardsPowerThreeOrLess` (`FR-TGT-001`, `FR-ZON-004`): player 0's pile holds a power-3 unit card, a power-2 unit card, a power-4 unit card, a Structure card with power 1 and a Maneuver card; player 1's pile holds a power-1 unit card. The target decision offers exactly the power-3 and power-2 unit cards (3 is included, 4 is not; only unit cards; only its controller's pile). Choosing the power-2 one deploys it, and the power-3 one stays in the pile.
- `TestCard_ReaversRush_NotEnactableWithoutLegalCard` (`SVY-R29`, `FR-MNV-001`): Enact is not offered, and a rejected Enact leaves the card in hand, with (a) an empty pile, (b) only the real *Chainbreaker Phalanx* (power 4) in player 0's pile, (c) only a Maneuver card in it, (d) only a power-1 unit card in player 1's pile. It is offered with (e) one power-3 unit card in player 0's pile.
- `TestCard_ReaversRush_GuerrillasZoneTimeoutEntersReserves` (#140 (4), `SVY-R43`, `FR-MOD-001` §6): the real *Devilhorn Guerrillas* in player 0's pile and an opposing unit in the Main Theater. The one decision player 0 is asked is a `DecideZone`, answered with a timeout. The Guerrillas are in player 0's Reserves, and the opposing unit has no damage (the Main Theater deploy trigger doesn't fire).
- `TestCard_ReaversRush_GuerrillasMainTheaterFiresDeployTrigger` (`FR-TRN-003a` 1): the same setup; answering the `DecideZone` with its second option (the Main Theater) puts the Guerrillas in the Main Theater, and the opposing unit, the trigger's only legal target, takes 2 damage.
- **Existing tests:** every other test passes unchanged, including `TestStarterSet_Executable` with the name removed, `TestView_Case8_ExecutableFlag`, `TestProgress_Case4_Coverage`, `TestEvents_Case4_EveryActionKind`, the *Devilhorn Guerrillas* tests and the T-0208 and T-0209 tests.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected. The test author keeps any small builder in the test file.

## Allowed Files
- `pkg/engine/cards/path.go` (the definition)
- `pkg/engine/cards/cards.go` (remove the name from `NotYetExecutable`)

## Done When
- [x] The card tests pass: `go test ./pkg/engine -run TestCard_ReaversRush`
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks
- [x] The card loads with no unsupported lines (`TestStarterSet_Executable` passes with the name removed)

## Log
- 2026-11-10 lead-architect (wave F run, session 8): split off T-0208 under the D2 size rule; `queued` on T-0208 and T-0209.
- 2026-11-10 lead-architect (wave F run, session 10): Inputs, Definition, Tests and Allowed Files written against the run branch at 8647386, with T-0208 and T-0209 `done`. Searched `pkg/engine` and the task archive for an existing equivalent: `OpDeployFree` (T-0208) and `Filter.Stat` (T-0209) are the shapes, reused as is; no new shape. Directive #140 (4) applies here. Probed the definition (see above), then reverted it. Set to `tests`.
- 2026-11-10 rules-test-author: wrote pkg/engine/cards_path_reavers_rush_test.go: TestCard_ReaversRush_DeploysChosenUnitCardForFree, _OffersOnlyUnitCardsPowerThreeOrLess, _NotEnactableWithoutLegalCard (subtests a-e), _GuerrillasZoneTimeoutEntersReserves, _GuerrillasMainTheaterFiresDeployTrigger. Helpers requested: none (small builders kept in the test file). No cases left out. All fail now as expected (card not yet executable). Set to tests-review.
- 2026-11-10 lead-architect (wave F run, session 10): tests reviewed (6978550, 5 tests, one with five subtests). Each matches the Tests list and `FR-TRN-003a`, `FR-MNV-001` (`SVY-R29`), `FR-TGT-001` and `SVY-R43`; only asked decisions are scripted, and the Guerrillas tests check the `DecideZone` is the only question. Probed them in a review worktree with the Definition added, since reverted: all five pass, and `TestStarterSet_Executable` passes with the name removed; without it, all fail. Set to `ready`.
- 2026-11-10 kernel-engineer: implemented Reaver's Rush in pkg/engine/cards/path.go between Pincer Tactics and Rogue Fleshshaper using ability.OpDeployFree (reused from T-0208) with one target filtered by AllegianceAllied, KindUnit, ZoneDiscard, StatPower AtMost 3 (reused from T-0209); removed Reaver's Rush from NotYetExecutable in pkg/engine/cards/cards.go. Verified: go test ./pkg/engine -run "TestCard_ReaversRush" passed; TestStarterSet_Executable passed; go tool task sim passed (starter decks clean: Path 44/50 executable, P0 73.5% / P1 25.5% / draws 1.0% over 200 games); go tool task check passed all formatting, vet, tests, and cardtown linters. Set to review.
