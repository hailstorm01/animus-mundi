---
id: T-0216
title: Veil Render, deployed from its owner's discard pile by killing an allied unit (wave F card, flash)
kind: card
tier: flash
wave: F
status: done
branch: m2/autoloop-wave-f/T-0216-veil-render
depends_on: [T-0213, T-0214, T-0215, T-0217]
claims: [card:Veil Render]
cards: [Veil Render]
requirements: [FR-TRN-003, FR-TRN-003a, FR-CST-001, FR-MOD-001]
---

## Goal
*Veil Render* (Animus Unit, R3 T3, Reserves, power 5, health 4; one copy in `data/decks/animus-starter.txt`; FRD 02 benchmark #6, its cost narrowed to a unit by `SVY-R46`) is executable: one discard deploy permission (T-0214) whose cost is one `CostKill` of an allied **unit**, with the same filter as *Feed the Despoiler*. It leaves `cards.NotYetExecutable`. It combines shapes only. Directive #143 (1), (2) and (3) apply here.

Covers:
- Its definition in `pkg/engine/cards/animus.go`, after *Unceasing Tide*, and its name removed from `cards.NotYetExecutable`.
- Card tests from the CSV text: from the discard pile it pays R3 T3 and kills the chosen allied unit after the push; a Scratch Company unit can pay; with only allied Structures and the Headquarters it can't be deployed from the pile (`SVY-R46`); a timed-out cost pick kills the first allied unit in fixed order (Reserves before the Main Theater, oldest first) and never the Headquarters or a Structure (`FR-MOD-001` §6, `SVY-R46`); the kill's death trigger resolves above it (`FR-TRN-003`'s example); a replaced kill still pays (`SVY-R15`); from hand it deploys as a plain Unit with no kill; a free deploy waives the kill (`FR-TRN-003a`'s example; *Reaver's Rush* can't take it at power 5, so a synthetic `OpDeployFree` Maneuver does); it is never deployed from the opponent's pile.
- `pkg/engine/doc.go`'s shape list gains a line for `Def.Statics` (from T-0214, whose Allowed Files had no room for it).
- `TestView_Case8_ExecutableFlag` needed *Veil Render* as the last unsupported Animus card in seat 0's seed-1 hand; it is edited as named under Tests.

## Inputs
- Card (`data/cards/units-and-structures-animus.csv` line 17): Animus Unit (Spirit, Horror), R3 T3, `RESERVES`, power 5, health 4: "You may deploy this card from your discard pile by killing an allied unit in addition to paying its deployment costs."
- Survey row: `docs/design/m2_starter_set_ability_survey.md` line 160 (deploy from discard + kill allied unit; `FR-TRN-003`, `FR-TRN-003a`, `FR-CST-001`); ruling row `SVY-R46` (line 105).
- FRD 01 v3.24: `FR-TRN-003` (line 360: a deploy from the Discard pile when a static ability on the card permits it), `FR-TRN-003a` (line 362: the kill is paid on declaration and its triggers go on top; line 363: a free deploy pays no additional cost), "Non-Battlefield Cost Evaluation" (line 320, `FR-CST-001`: `activeZone: DISCARD` on *Veil Render*), `FR-REP-001` with `SVY-R15` (line 512), `FR-MOD-001` §6 (a timed-out choice takes the first legal option in fixed order), revision summary v3.24 (line 18, `SVY-R46`).
- Rulings: `SVY-R46` (the cost kills an allied unit, Scratch Company units included, never a Headquarters or Structure), `SVY-R15`, `SVY-R16`, `SVY-R29`.
- Directive #143 (`BOARD.md`), outcome `docs/design/outcomes/2026-11-11-wave-f-veil-render-headquarters-kill.md`, design note Open Question 21 (answered).
- Code: `ability.Static`, `StaticDeployPermission`, `ActiveZone`, `Costs` ([ability.go:46-67](../../../pkg/engine/ability/ability.go)); `deployPermission` and `deployPlacement` ([placement.go:70-95](../../../pkg/engine/placement.go)); *Feed the Despoiler*'s `CostKill` filter ([cards/animus.go:291-310](../../../pkg/engine/cards/animus.go)). T-0214's synthetic tests in `pkg/engine/deploy_discard_test.go` cover the same permission with a synthetic card; this task's tests use the real card.
- `kill` still skips every Headquarters ([sweep.go:31](../../../pkg/engine/sweep.go)): a documented M2 limit, left as it is (#143 (4)). With the unit filter it never comes up here.

### Definition (the engineer adds it to `animusDefs`, after *Unceasing Tide*, the last entry)
- `Text`: the CSV text above, as one line.
- `Statics`: one `ability.Static{Kind: ability.StaticDeployPermission, ActiveZone: ability.ZoneDiscard, Costs: []ability.Cost{{Kind: ability.CostKill, Filter: ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit}}}}`.
- No `Triggers`, `Effect` or `ExtraCosts`: a hand deploy has no additional cost.
- Remove `"Veil Render"` from `cards.NotYetExecutable` in `pkg/engine/cards/cards.go`.

### doc.go (the engineer adds this bullet to the shape list in `pkg/engine/doc.go`, right after the bullet that begins `A card's own hand-active cost reduction`)
```go
//   - A card's own discard-active deploy permission ("You may deploy this card
//     from your discard pile by ...", Def.Statics with ability.StaticDeployPermission):
//     its owner deploys it as an ActDeploy with From: ZoneDiscard, and the
//     permission's Costs are paid on declaration (FR-TRN-003, FR-TRN-003a).
```

A lead probe with this definition on the run branch at 40df6d9, since reverted, passed probe versions of the tests below, and with the `TestView_Case8_ExecutableFlag` edit below `go tool task check` passes. `go tool task sim` stays clean: Animus 49 → 50/50 main cards executable, P0 71.5% / P1 27.5% / draws 1.0% over 200 games. So `sim-compare` will differ here and isn't a gate.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
Add them to a new file `pkg/engine/cards_animus_veil_test.go`, in the style of the T-0214 tests in `pkg/engine/deploy_discard_test.go`, reusing their helpers: `s.card("Veil Render")` put into a discard pile with `s.inDiscard`, `ddDeploy`, `ddOffers`, `dcAssetOf`, `kcScript`, `s.scriptRefs`, `dfManeuver`, `enactAction`, `inDiscard`, `inHand`, `s.setPool`, `s.reject`. Synthetic allied cards (`unit`, `structure`) and the real *Unceasing Tide* and *Shifting Impossibility*. Player 0 is the active player and deploys. Script only the decisions actually asked: a single legal option is taken without a question (`FR-MOD-001` §1).
- `TestCard_VeilRender_DiscardDeployKillsTheChosenAlliedUnit` (`FR-TRN-003`, `FR-TRN-003a`): player 0 has a unit in the Main Theater and one in Reserves, *Veil Render* in their discard pile, and R2 T3: no discard deploy is offered. With R3 T3 it is offered exactly once (its only deploy zone is Reserves). Choosing the Main Theater unit for the cost: that unit is dead and its card is in player 0's discard pile, the Reserves unit is alive, *Veil Render* is an asset in player 0's Reserves, controlled by player 0, with power 5 and Maximum Health 4 (`g.Power`, `g.MaxHealth`), and no longer in the discard pile, and the pool is R0 T0.
- `TestCard_VeilRender_ScratchCompanyCanPay` (`SVY-R46`): player 0's only unit is a Scratch Company (`s.putScratch`). The discard deploy is offered with R3 T3, and deploying it kills the Scratch Company (its card in player 0's discard pile) and puts *Veil Render* into Reserves.
- `TestCard_VeilRender_NotDeployableWithoutAnAlliedUnit` (`SVY-R46`, `SVY-R29`): player 0 has an allied Structure in the Main Theater (and their Headquarters), player 1 has a unit, and the pool is R5 T5. The discard deploy isn't offered, and a rejected one changes nothing: *Veil Render* stays in the discard pile, the Structure and both Headquarters are alive and undamaged, player 1's unit is alive, and the pool is still R5 T5.
- `TestCard_VeilRender_TimedOutCostKillsTheFirstAlliedUnit` (`FR-MOD-001` §6, `SVY-R46`, `SVY-R16`): player 0 has, put in this order, a unit in the Main Theater, unit A in Reserves, unit B in Reserves, and a Structure in the Main Theater; R3 T3. Seat 0 answers one decision with `Answer{Timeout: true}`. Exactly one decision is asked, and its options refer to exactly the three units (not the Structure, nor player 0's Headquarters). Unit A is killed; the Main Theater unit, unit B, the Structure and player 0's Headquarters are alive, and *Veil Render* is in Reserves.
- `TestCard_VeilRender_CostDeathTriggerResolvesFirst` (`FR-TRN-003a`, `FR-TRN-003`): player 0's only unit is the real *Unceasing Tide*; R3 T3. After the discard deploy, *Unceasing Tide*'s card is in player 0's hand (not the discard pile) and *Veil Render* is an asset in Reserves. In the log, the `place` event of *Unceasing Tide*'s trigger comes after *Veil Render*'s `place` event and before its `deploy` event (the same check as `TestFR_TRN_003_DiscardDeployPaysItsKillCostAfterThePush`).
- `TestCard_VeilRender_ReplacedKillStillPays` (`SVY-R15`, `FR-REP-001`): player 0 has the real *Shifting Impossibility* in Reserves and a Scratch Company; R3 T3. The cost pick chooses the Scratch Company (`s.scriptRefs`). It is returned to player 0's hand, not the discard pile, *Shifting Impossibility* has 1 damage, and *Veil Render* is an asset in Reserves.
- `TestCard_VeilRender_HandDeployHasNoKill` (`FR-TRN-003`): *Veil Render* in player 0's hand, an allied unit in the Main Theater, R3 T3, and no decision scripted (`s.script(0)`). A Deploy from hand to Reserves (no `From`) asks no decision, kills nothing, puts *Veil Render* into Reserves and leaves the pool R0 T0.
- `TestCard_VeilRender_FreeDeployWaivesTheKill` (`FR-TRN-003a`, `FR-CST-001`): *Veil Render* in player 0's discard pile, an allied unit in the Main Theater, R3 T3, and a synthetic `dfManeuver(0, 0, ability.ZoneAny)` in hand; no decision scripted. Enacting it puts *Veil Render* into Reserves with no decision asked, the allied unit is alive, and the pool is still R3 T3.
- `TestCard_VeilRender_NotFromTheOpponentsPile` (`FR-TRN-003`): *Veil Render* in player 1's discard pile, player 0 has an allied unit and R5 T5. No discard deploy is offered, and a rejected one leaves it in player 1's discard pile.
- **Existing test, authorized edit (in the `test(T-0216)` commit):** `TestView_Case8_ExecutableFlag` in `pkg/view/executable_test.go`. Seat 0's seed-1 hand has no unsupported card once *Veil Render* is executable. Replace the last two lines of its header comment (`// Seed 1 is chosen: ... as well as Veil Render (unsupported).`) with:
  ```go
  // Seed 1 is chosen: seat 0's initial hand contains Despoiler Larva (supported).
  // Every Animus card is executable since T-0216 (Veil Render), so the test swaps
  // seat 0's last hand card for a copy with an unsupported line.
  ```
  and insert right after the `NewGame` error check:
  ```go
  	hand := g.Players[0].Hand
  	if len(hand) == 0 {
  		t.Fatalf("seat 0 has empty hand")
  	}
  	unsupported := *hand[len(hand)-1].Card
  	unsupported.Unsupported = []string{"An unsupported line."}
  	hand[len(hand)-1].Card = &unsupported
  ```
  Nothing else in the test changes. It passes before and after the definition.
- **Other existing tests:** every other test passes unchanged, including `TestStarterSet_Executable` with the name removed, `TestProgress_Case4_Coverage`, `TestEvents_Case4_EveryActionKind`, the T-0214 tests in `deploy_discard_test.go`, and the deck count in `deck_load_test.go`.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected. The test author keeps any small builder in the test file.

## Allowed Files
- `pkg/engine/cards/animus.go` (the definition)
- `pkg/engine/cards/cards.go` (remove the name from `NotYetExecutable`)
- `pkg/engine/doc.go` (the shape-list bullet above)

## Done When
- [x] The card tests pass: `go test ./pkg/engine -run TestCard_VeilRender`
- [x] `go test ./pkg/view -run TestView_Case8_ExecutableFlag` passes
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks
- [x] The card loads with no unsupported lines (`TestStarterSet_Executable` passes with the name removed)

## Log
- 2026-11-10 lead-architect (wave F run, session 13): queued with T-0213 (D2.4). Directive #128 is `open: before T-0216`, applied by T-0213, T-0214 and T-0215.
- 2026-11-11 lead-architect (wave F run, session 16): T-0215 split under the D2 size rule; the play screen's label went to T-0217, now in `depends_on`.
- 2026-11-11 lead-architect (wave F run, session 17): not written, stays `queued`. A probe of the benchmark's definition (one `StaticDeployPermission` with `CostKill`, `Filter{Allegiance: AllegianceAllied}`, since reverted) passes everything except `TestView_Case8_ExecutableFlag`, which needs an unsupported card in seat 0's hand at seed 1 (*Veil Render* was the last Animus one). The task names that test as an authorized edit: swap one hand card's `Card` for a copy with an `Unsupported` line. Writing it found a rules gap. "An allied asset" includes the Headquarters (FRD 01 line 211), but `kill` skips a Headquarters ([sweep.go:31](../../../pkg/engine/sweep.go)), so the cost would be free with it, and a timed-out pick takes it first (`FR-MOD-001` §6). FRD 01 doesn't say what a killed Headquarters does either. Ruling question on `main` (5b35925): `docs/design/outcomes/2026-11-11-wave-f-veil-render-headquarters-kill.md`, design note Open Question 21. Of the Covers, "no allied asset to kill" can't happen with the real card (there is always a Headquarters unless the ruling excludes it), and no printed executable card can free-deploy it (*Reaver's Rush* allows power 3 or less), so that case uses a synthetic `OpDeployFree` Maneuver on the real card. Model cards for the tests: *Unceasing Tide* (a death trigger above the deploy) and *Shifting Impossibility* with a Scratch Company (a replaced kill).
- 2026-11-11 lead-architect (wave F run, session 19): `main` synced into the run branch (40df6d9: `SVY-R46`, FRD 01 v3.24, *Veil Render*'s rewording, directive #143). #143's trigger fired (the outcome reads `Status: applied: FRD 01 v3.24`), so it is `open: before T-0216` and applied here: (1) the Definition uses the unit filter, (2) the Covers are rewritten for a unit, (3) the title and Goal say "an allied unit" and the design note's Open Question 21 is marked answered in this commit, (4) `kill` is left as it is. Inputs, Definition, Tests and Allowed Files written against the run branch at 40df6d9, with T-0213, T-0214, T-0215 and T-0217 `done`. Searched `pkg/engine` and the task archive for an existing equivalent: `StaticDeployPermission` (T-0214) is the shape, reused as is, with *Feed the Despoiler*'s filter; no new shape. No other waiting directive fires with this card (#136 waits for a trigger that fires off the battlefield; a static permission isn't one). Probed the definition, the tests and the Case 8 edit (see above), then reverted them. Set to `tests`.
- 2026-11-11 rules-test-author: wrote pkg/engine/cards_animus_veil_test.go (commit b7f2cda): TestCard_VeilRender_DiscardDeployKillsTheChosenAlliedUnit, _ScratchCompanyCanPay, _NotDeployableWithoutAnAlliedUnit, _TimedOutCostKillsTheFirstAlliedUnit, _CostDeathTriggerResolvesFirst, _ReplacedKillStillPays, _HandDeployHasNoKill, _FreeDeployWaivesTheKill, _NotFromTheOpponentsPile. Made the authorized TestView_Case8_ExecutableFlag edit (passes now). Helpers requested: none (small vrHQ builder kept in the test file). Cases left out: none. Expected red until the definition lands: 6 fail (card unsupported); NotDeployable and NotFromTheOpponentsPile pass trivially while unsupported. cardtown lint rules passes. Set to tests-review.
- 2026-11-11 lead-architect (wave F run, session 19): tests reviewed (b7f2cda, 9 tests). Each matches the Tests list and `FR-TRN-003`, `FR-TRN-003a`, `FR-CST-001`, `FR-MOD-001` §6, `FR-REP-001` (`SVY-R15`, `SVY-R16`, `SVY-R29`, `SVY-R46`); only asked decisions are scripted. The `TestView_Case8_ExecutableFlag` edit is exactly the one named. Probed them in a review worktree with the Definition added, since reverted: all nine pass, and `TestStarterSet_Executable` and `pkg/view` pass with the name removed; without it, seven fail (the two "not offered" tests pass either way and guard the rule once it is defined). Set to `ready`.
- 2026-11-11 kernel-engineer: implemented Veil Render definition (commit 47ff082).
  - pkg/engine/cards/animus.go: added "Veil Render" definition to animusDefs with StaticDeployPermission on ZoneDiscard and CostKill with Filter{Allegiance: AllegianceAllied, Kinds: KindUnit} (reusing StaticDeployPermission from T-0214 and the filter from Feed the Despoiler).
  - pkg/engine/cards/cards.go: removed "Veil Render" from NotYetExecutable.
  - pkg/engine/doc.go: documented the discard-active deploy permission shape under Def.Statics with StaticDeployPermission.
  - Verification: `go test ./pkg/engine -run TestCard_VeilRender` (9/9 passed); `go test ./pkg/view -run TestView_Case8_ExecutableFlag` (passed); `go test ./pkg/engine -run 'TestStarterSet_Executable|TestDefs_MatchRulesText'` (passed); `go tool task check` passed clean (fmt-check, vet, all package tests, all linters); `go tool task sim` clean (200 games, P0 50/50 executable, 71.5% win rate, all invariants held). Set to review.
- 2026-11-11 lead-architect (wave F run, session 19): code reviewed (47ff082): only the three Allowed Files changed, with the Definition and the doc.go bullet as written; the test commit is untouched; no card name in engine code and no new shape. Re-ran `go test ./pkg/engine -run TestCard_VeilRender`, `go test ./pkg/view -run TestView_Case8_ExecutableFlag`, `go tool task check` and `go tool task sim` (clean, Animus 50/50, P0 71.5% / P1 27.5% / draws 1.0%) in a review worktree. `agy` conversation `844198b6-0958-4008-9a01-4d09354588e0`, no denied actions. Approved.
- 2026-11-11 lead-architect (wave F run, session 19): merged into `autoloop/m2-wave-f` (8630b2a); `done`. Directive #143 fully applied.
