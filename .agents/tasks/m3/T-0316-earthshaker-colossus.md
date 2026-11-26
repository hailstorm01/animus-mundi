---
id: T-0316
title: Make Earthshaker Colossus executable (deploys or attacks: 2 to a chosen non-Scratch asset; wave A, card, flash)
kind: card
tier: flash
wave: A
status: review
branch: m3/autoloop-wave-a/T-0316-earthshaker-colossus
depends_on: [T-0315]
claims: [card:Earthshaker Colossus]
cards: [Earthshaker Colossus]
requirements: [FR-TRN-003, FR-SET-004, FR-TGT-001, FR-TGT-003, FR-MOD-001, FR-CBT-001, FR-CBT-002, FR-CBT-003, FR-LAY-002, FR-REP-004, FR-KWD-002h, FR-KWD-002k, FR-ZON-004, FR-ZON-005, FR-ENT-103]
---

## Goal
*Earthshaker Colossus* (Soulless unit, Mechanized and Giant, R6 T0, Reserves, power 6, health 7; two copies in `data/decks/soulless-starter.txt`): *"Whenever this unit deploys or attacks, deal 2 damage to chosen non-Scratch asset."* It only combines shapes that are `done`: the two triggers of the M2 Colossi (*Spoil Forest Colossus* in `pkg/engine/cards/animus.go`, *Ironclad Colossus* in `pkg/engine/cards/path.go`), one `HookDeploy` and one `HookAttacks` self trigger with the same effect, and that effect is *Bring It Down*'s (`pkg/engine/cards/soulless.go`): one exactly-one slot with `Scratch: ability.ScratchExcluded`, and `OpDealDamage` 2 on it. The slot has no `Kinds`, `Allegiance` or `Other`, so it offers both Headquarters, every structure and every face-up unit, the Colossus itself included (survey G5 and `SVY-R59`: a Scratch Company is never offered). Both Headquarters are always legal, so the target is always asked: the Single-Option Fast-Path never applies.

Covers:
- The definition in `soullessDefs` (`pkg/engine/cards/soulless.go`), as under Definition below.
- The name off `cards.NotYetExecutable`.

## Inputs
- Card (`data/cards/units-and-structures-soulless.csv` line 27): Soulless unit, Mechanized and Giant, Requisition cost 6, Tactics cost 0, Reserves, power 6, health 7.
- Survey row: `docs/design/survey/m3_full_catalog_ability_survey.md` §6, "Soulless: deck" (*Earthshaker Colossus*, line 173).
- FRD 01 (`docs/frd/01_Game_Rules_FRD.md`): `FR-TRN-003` (line 393), `FR-SET-004` (line 201), `FR-TGT-001` (line 407), `FR-TGT-003` (line 435), the Single-Option Fast-Path and `FR-MOD-001` §6 timeouts (line 449 on), `FR-CBT-001` (line 598: only a Main Theater unit attacks), `FR-CBT-002` (line 613: Step 2 attack triggers resolve on top of the Attack, and Step 3's Combatant Target Check fizzles an attack whose defender is gone), `FR-CBT-003` (line 630), `FR-LAY-002` (line 534: the death sweep runs after each resolving item, so a defender the trigger kills is gone before the Attack resolves), `FR-REP-004` (line 562), `FR-KWD-002h` Ward and `FR-KWD-002k` Armor, `FR-ZON-004` (line 278), `FR-ZON-005` (line 287), `FR-ENT-103` (line 297).
- Rulings that apply: `SVY-R3` (an attack trigger's target is chosen on placement), `SVY-R16` (timeouts take the first option in fixed order, never random), `SVY-R25` (a removed target is no longer that target), `SVY-R59` (a face-down Scratch Company is 0/0 and not a non-Scratch asset).
- Design note: `docs/architecture/m3-full-catalog.md`, Boundaries (`pkg/engine/cards`), Deliverable 5.

### Definition (the engineer adds it to `soullessDefs`, after *Dolus Refiner*)
- `"Earthshaker Colossus"`, `Text`: `"Whenever this unit deploys or attacks, deal 2 damage to chosen non-Scratch asset."` (one string).
- Two `Triggers`, in this order, as *Spoil Forest Colossus* has them: `Hook: ability.HookDeploy`, then `Hook: ability.HookAttacks`. Each has the same `Effect`:
  - `Targets`: one slot, `Filter: ability.Filter{Scratch: ability.ScratchExcluded}`.
  - `Steps`: `{Op: ability.OpDealDamage, Amount: ability.Value{N: 2}, Subject: ability.SubjectTarget, Target: 0}`.

A lead probe with the definition on the run branch at 8333f9e2, since reverted, passed `go tool task check`, `go tool task sim` (P0 58.5% / P1 39.5% / draws 2.0%, unchanged) and `go tool task sim-matrix` (exit 0).

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
In a new file `pkg/engine/cards_soulless_earthshaker_test.go`, from the CSV text, loading the real card (`s.card("Earthshaker Colossus")`), in the style of the *Ironclad Colossus* tests in `cards_path_test.go` (a deploy and an attack trigger on one unit; `deployPath`, `colossusCycle`) and of `cards_soulless_macrocannon_test.go` and `cards_soulless_maneuvers_test.go` (*Bring It Down*'s non-Scratch slot). Player 0 controls the Colossus and is the active player, unless a case says otherwise. Reuse the existing test helpers (`deployPath`, `colossusCycle`, `wantOptions`, `s.put`, `s.putScratch`, `s.give`, `s.onTop`, `s.setPool`, `s.script`, `s.scriptRefs`, `s.asked`, `s.alive`, `s.attack`, `s.endTurn`, `s.wantDamage`, `s.ptPlaceHook`, `stormKill`, `chooseAsset`, `wantDealt`, `nonCombatDamage`, `totalDamage`, `inDiscardPile`, `morphEnact` and `morphManeuver` from `morph_scratch_test.go`). The other cards may be synthetic. Every asset that a case doesn't mean to kill has enough health to survive the damage it takes.

Deploying (player 0 deploys the Colossus from hand to Reserves with exactly R6, unless the case says otherwise):
- `TestCard_EarthshakerColossus_EntersReservesAsSixSeven` (`FR-TRN-003`): it is a Readied 6/7 unit in player 0's Reserves, and the pool is empty. Deploying it to the Main Theater is rejected, and with R5 it can't be deployed; in both cases the card stays in hand.
- `TestCard_EarthshakerColossus_DeployDealsTwoToTheChosenAsset` (`FR-TGT-001`, `FR-LAY-002`): choosing an opposing Main Theater unit gives it 2 damage, one non-combat damage instance from the Colossus, and no other asset is damaged.
- `TestCard_EarthshakerColossus_DeployOffersEveryNonScratchAsset` (`FR-TGT-001`, `FR-ZON-005`, `FR-KWD-002h`, `SVY-R59`): with an allied and an opposing unit in each of the Reserves and the Main Theater, an allied and an opposing structure, a face-down Scratch Company on each side (`s.putScratch`), an allied Warded unit and an opposing Warded unit, exactly one `DecideTarget` is asked of player 0, and it offers exactly both Headquarters, the four units, both structures, the allied Warded unit and the Colossus itself. Neither Scratch Company is offered, nor the opposing Warded unit.
- `TestCard_EarthshakerColossus_UnitTurnedFaceDownNotOffered` (`FR-ENT-103`, `SVY-R59`): an opposing unit turned face down by `morphEnact(s, morphManeuver())` before the Colossus deploys is a Scratch Company and not offered. A face-up opposing unit beside it is.
- `TestCard_EarthshakerColossus_AloneStillAsks` (`FR-TGT-001`, Single-Option Fast-Path): with no other asset on the battlefield, the deploy trigger still asks, offering exactly the two Headquarters and the Colossus. Choosing the Colossus leaves it on the battlefield with 2 damage.
- `TestCard_EarthshakerColossus_DeployDamagesTheOpposingHeadquarters` (`FR-CBT-003`): choosing player 1's Headquarters lowers its remaining Health by 2, and the match isn't over.
- `TestCard_EarthshakerColossus_TimeoutTakesTheFirstOption` (`FR-MOD-001` §6, `SVY-R16`): a timed-out target decision (`s.script(0, Answer{Timeout: true})`) takes its first option in fixed order, as in `TestCard_DolusMacrocannon_TimeoutTakesTheFirstOption`, so player 0's own Headquarters takes 2 damage, and the total damage dealt is 2.
- `TestCard_EarthshakerColossus_LethalDamageKills` (`FR-LAY-002`): a chosen opposing unit with health 2 is destroyed, and its card is in player 1's discard pile.
- `TestCard_EarthshakerColossus_ArmorAbsorbs` (`FR-REP-004` Stage 4, `FR-KWD-002k`): a chosen opposing unit with printed Armor 1 loses its Armor and takes 1 damage.
- `TestCard_EarthshakerColossus_TargetGoneDealsNothing` (`FR-TGT-003`, `SVY-R25`): the chosen opposing unit is removed between the deploy trigger's placement and its resolution (`s.ptPlaceHook` with `stormKill`, as in `TestCard_ConscriptRifles_TargetGoneDealsNothing`). No asset is dealt damage.
- `TestCard_EarthshakerColossus_SourceGoneStillDeals` (`FR-ZON-004`): the Colossus is removed between its deploy trigger's placement and resolution (as in `TestCard_ConscriptRifles_SourceGoneStillDeals`). The chosen opposing unit is still dealt 2, and the Colossus' card is in player 0's discard pile.
- `TestCard_EarthshakerColossus_FaceDownScratchDoesNothing` (`FR-ZON-005`): with an *Earthshaker Colossus* on top of player 0's deck (`s.onTop`) and an opposing unit, player 0 enacts the real *Growing Legion*. A face-down Scratch Company enters player 0's Reserves, no target is asked, and no asset is dealt damage.

Attacking (the Colossus deployed as above, its deploy trigger answered with player 1's Headquarters, which so has 2 damage before any attack, then marched to the Main Theater and readied with `colossusCycle`; the cases put player 1's assets in after that, so the deploy trigger's options don't depend on them):
- `TestCard_EarthshakerColossus_AttackDealsTwoToTheChosenAsset` (`FR-CBT-002`, `SVY-R3`): attacking an opposing Main Theater unit (power 1, health 20) and choosing player 1's Headquarters on the attack trigger: the Headquarters takes 2 more (4 in all), and the defender takes the Colossus' 6 combat damage and deals 1 back. One target question for the deploy and one for the attack: `s.asked(0)` has exactly two decisions.
- `TestCard_EarthshakerColossus_AttackOffersEveryNonScratchAsset` (`FR-TGT-001`, `FR-ZON-005`, `FR-KWD-002h`): the attack trigger's decision offers the same assets as `DeployOffersEveryNonScratchAsset` with the same board (the defender among them, and units in Reserves, which can't be attacked but can be chosen), and not the Scratch Companies or the opposing Warded unit.
- `TestCard_EarthshakerColossus_AttackTriggerResolvesBeforeCombat` (`FR-CBT-002` Steps 2 and 3, `FR-LAY-002`): attacking an opposing unit with power 5 and health 2, and choosing that defender on the attack trigger, destroys it at the death sweep after the trigger resolves. The Attack then fizzles: the Colossus is dealt no retaliation damage (it has no damage), the defender's card is in player 1's discard pile, and player 1's Headquarters has only the deploy trigger's 2 damage.
- `TestCard_EarthshakerColossus_EachAttackTriggersAgain` (`FR-CBT-002`): attacking on two of player 0's turns (with `colossusCycle`-style turn passes in between, as in `TestCard_IroncladColossus_EachAttackGrantsAgainAndGrantsStack`) asks the attack trigger each time, and each chosen asset takes 2.
- `TestCard_EarthshakerColossus_DefendingDoesNotTrigger` (`FR-CBT-002`): the Colossus in player 0's Main Theater is attacked on player 1's turn, as in `TestCard_IroncladColossus_DefendingDoesNotTrigger`. Player 0 is asked nothing while defending, and only combat damage is dealt.
- **Existing tests:** every other test passes unchanged, including `TestCards_Executable` with the name removed.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected.

## Allowed Files
- `pkg/engine/cards/soulless.go` (the definition)
- `pkg/engine/cards/cards.go` (remove the name from `NotYetExecutable`)

## Done When
- [x] The card tests pass: `go test ./pkg/engine -run TestCard_EarthshakerColossus`
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks
- [x] `go tool task sim-matrix` exits 0 (just over two minutes: use a timeout above the two-minute default)
- [x] The card loads with no unsupported lines (`TestCards_Executable` passes with the name removed)

## Log
- 2026-11-26 lead-architect (wave A run, session 18): written against 8333f9e2, with the lead's probe above. `main` has no commits the run branch lacks. Board read (D2.0): no directive is due here (#185 has only (4)'s *Awaken Nightmare* half left, wave F), and no waiting trigger has fired: the card has no each-player effect (#71), no cost on a mode (#91 (2)), no deck search (#106), no discard cost (#120), no tuck (#142), no pile round trip (#183 (2)) and no turn-boundary or draw hook (#15 (2): its hooks are deploy and attack), the mode is still `serial` (#6, #7), and no `needsRework` row exists (#184). #182 is about combat-kill triggers, which this card doesn't have. Taken next in the survey's Soulless deck order, after *Dolus Refiner*. Searched `pkg/engine` and the task archives: every part of the Definition exists (the M2 Colossi's `HookDeploy` and `HookAttacks` pair, *Bring It Down*'s `ScratchExcluded` slot and `OpDealDamage` step), so no kernel change. The repeated effect in the two triggers is card data, as in *Spoil Forest Colossus* and *Ironclad Colossus*, not a rule written twice. Set to `tests`.
- 2026-11-26 rules-test-author: wrote `pkg/engine/cards_soulless_earthshaker_test.go` (17 `TestCard_EarthshakerColossus_*` tests, one per bullet under Tests, deploying and attacking, plus helpers local to the file `ecDeploy`, `ecBoard`, `ecAttackScene`). `go vet` is clean; every test fails except `FaceDownScratchDoesNothing` (passes trivially until the card is executable) until the definition lands. Helpers requested: none. Cases left out: none. In the offers cases the opposing Warded unit sits in Reserves, so no attack rule can interfere with the defender.
- 2026-11-26 lead-architect: tests approved (D4, round 1). 4f12eb7e adds only `pkg/engine/cards_soulless_earthshaker_test.go` (17 tests, one per case) and the task file. With the lead's probe Definition all 17 pass, and each of 9 mutants (damage 1, damage 3, no `ScratchExcluded`, opposing only, units only, no Headquarters, no deploy trigger, no attack trigger, `HookDamaged` in place of `HookAttacks`) fails 3 to 15 of them. Expected outcomes agree with the FRD 01 text they cite: `AttackTriggerResolvesBeforeCombat` pins `FR-LAY-002`'s death sweep after the trigger and `FR-CBT-002`'s fizzle (no retaliation). The opposing Warded unit sits in Reserves in the two offers cases, which keeps it out of the attack's defender rules and still tests Ward. Set to `ready`.
- 2026-11-26 kernel-engineer: implemented Earthshaker Colossus in `pkg/engine/cards/soulless.go` and removed the name from `pkg/engine/cards/cards.go` `NotYetExecutable`. The diff shows:
  - Added "Earthshaker Colossus" to `soullessDefs` with two triggers: `HookDeploy` and `HookAttacks`. Each trigger specifies an effect targeting one slot with `ability.Filter{Scratch: ability.ScratchExcluded}` and an `OpDealDamage` step dealing 2 to `SubjectTarget` (reusing the trigger pair structure from `Spoil Forest Colossus` / `Ironclad Colossus` and the non-Scratch targeting shape from `Bring It Down` / `Dolus Macrocannon`).
  - Removed "Earthshaker Colossus" from `NotYetExecutable` in `cards.go`.
  - Commands run:
    - `go test ./pkg/engine -run TestCard_EarthshakerColossus`: pass (17/17 tests pass).
    - `go tool task check`: pass (gofmt, `go vet`, all tests, `cardtown lint --all`).
    - `go tool task sim`: pass (200 games, P0 58.5% / P1 39.5% / draws 2.0%, clean termination).
    - `go tool task sim-matrix`: pass (24 starter and 24 stress matrix runs all pass with code 0).
    - `go test -count=1 ./pkg/cardcheck`: pass (executable check).
  - Ready for review. Set to `review`.
- 2026-11-26 lead-architect (wave A run, session 18): approved (D5, round 1). The diff (ad32fa9e) is the task's Definition in `pkg/engine/cards/soulless.go` and the one name off `NotYetExecutable`, inside `Allowed Files`. The test file changed only in 4f12eb7e, before the implementation. No new shape, no card name in engine code, and no duplicate: the two triggers are the M2 Colossi's `HookDeploy` and `HookAttacks` pair, and the effect is *Bring It Down*'s `ScratchExcluded` slot and `OpDealDamage` step. Re-run by the reviewer on the task branch, with exit codes: `go test ./pkg/engine -run TestCard_EarthshakerColossus`, `go test ./pkg/cardcheck -run TestCards_Executable`, `go tool task check`, `go tool task sim`, `go tool task sim-matrix`, all 0. The Animus vs Path sim is unchanged (P0 58.5% / P1 39.5% / draws 2.0%). The Flash run (`agy` conversation `dcb90f22-9508-4ab6-b55c-d750a188a912`) left no denied action.
