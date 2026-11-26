---
id: T-0320
title: Make MAX-8 Autonomous Warsuit executable (deploy: 2 damage to one chosen asset, heal 2 on another chosen asset, wave A, card, flash)
kind: card
tier: flash
wave: A
status: ready
branch: m3/autoloop-wave-a/T-0320-max-8-autonomous-warsuit
depends_on: [T-0319]
claims: [card:MAX-8 Autonomous Warsuit]
cards: [MAX-8 Autonomous Warsuit]
requirements: [FR-TRN-003, FR-TGT-001, FR-TGT-003, FR-MOD-001, FR-LAY-002, FR-KWD-002d]
---

## Goal
*MAX-8 Autonomous Warsuit* (Soulless unit, Cybernetic and Mechanized, R6 T0, Reserves, power 4, health 7; one copy in `data/decks/soulless-starter.txt`): *"When this unit deploys, deal 2 damage to chosen asset. Chosen asset heals 2."* It only combines shapes that are `done`: one `HookDeploy` self trigger (*Spoil Forest Colossus*, `pkg/engine/cards/animus.go`) whose effect has two slots, as *Ensnaring Mist*'s two clauses do (`animus.go`): slot 0 takes *Earthshaker Colossus*' `OpDealDamage` 2, and slot 1 takes *Spoil Forest Colossus*' `OpHeal` 2. Each *"chosen asset"* is its own slot with its own choice, and the two may pick the same asset (`SVY-R35`, `FR-TGT-001` line 432). Neither slot has a filter (*Spoil Forest Colossus*' `Filter{}`), so each offers both Headquarters, every structure, every unit (Scratch Companies and the Warsuit itself included) except an opposing Warded unit. Both Headquarters are always legal, so both slots are always asked: the Single-Option Fast-Path never applies.

Covers:
- The definition in `soullessDefs` (`pkg/engine/cards/soulless.go`), as under Definition below.
- The name off `cards.NotYetExecutable`.

## Inputs
- Card (`data/cards/units-and-structures-soulless.csv` line 28): Soulless unit, Cybernetic and Mechanized, Requisition cost 6, Tactics cost 0, Reserves, power 4, health 7.
- Survey row: `docs/design/survey/m3_full_catalog_ability_survey.md` §6, "Soulless: deck" (*MAX-8 Autonomous Warsuit*, line 177: `SVY-R35`, "Two slots").
- FRD 01 (`docs/frd/01_Game_Rules_FRD.md`): `FR-TRN-003` (line 393), `FR-SET-004` (line 201), `FR-TGT-001` (line 407; `SVY-R35` at line 432), `FR-TGT-003` (line 435 on: Partial Invalidation, Deployment Triggers Without Legal Targets), the Single-Option Fast-Path and `FR-MOD-001` §6 timeouts (line 449 on), `FR-LAY-002` (line 534: the death sweep runs only after the whole trigger resolves, never between its two steps), `FR-REP-004` (line 562), `FR-KWD-002d` Heal (line 644), `FR-KWD-002h` Ward (line 648), `FR-KWD-002k` Armor, `FR-ZON-004` (line 278), `FR-ZON-005` (line 287: a Scratch Company is a 1/1 unit).
- Rulings that apply: `SVY-R35` (each "chosen" clause is its own slot, and two slots may pick the same asset), `SVY-R3` (targets are chosen on placement), `SVY-R16` (timeouts take the first option in fixed order), `SVY-R25` (a removed target is no longer that target).
- Design note: `docs/architecture/m3-full-catalog.md`, Boundaries (`pkg/engine/cards`), Deliverable 5.

### Definition (the engineer adds it to `soullessDefs`, after *Inspired Salvage*)
- `"MAX-8 Autonomous Warsuit"`, `Text`: `"When this unit deploys, deal 2 damage to chosen asset. Chosen asset heals 2."` (one string: the card checks read the CSV text as one line, and two strings fail `TestDefs_MatchRulesText`).
- One `Trigger`, `Hook: ability.HookDeploy`, with this `Effect`:
  - `Targets`: two slots, each `Filter: ability.Filter{}`.
  - `Steps`, in this order: `{Op: ability.OpDealDamage, Amount: ability.Value{N: 2}, Subject: ability.SubjectTarget, Target: 0}`, then `{Op: ability.OpHeal, Amount: ability.Value{N: 2}, Subject: ability.SubjectTarget, Target: 1}`.

A lead probe with the definition on the run branch at e6573b07, since reverted, passed `go test ./pkg/cardcheck`, `go tool task check`, `go tool task sim` (P0 58.5% / P1 39.5% / draws 2.0%, unchanged) and `go tool task sim-matrix` (exit 0).

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
In a new file `pkg/engine/cards_soulless_max8_test.go`, from the CSV text, loading the real card (`s.card("MAX-8 Autonomous Warsuit")`), in the style of `cards_soulless_earthshaker_test.go` (a deploy trigger with an unfiltered-but-for-Scratch slot: `ecDeploy`, `ecBoard`) and the *Spoil Forest Colossus* heal tests in `cards_animus_test.go` (`TestCard_SpoilForestColossus_*`). Player 0 deploys the Warsuit from hand to Reserves with exactly R6 and is the active player, unless a case says otherwise. Reuse the existing test helpers (`deployPath`, `wantOptions`, `s.targetBoard`, `s.put`, `s.putScratch`, `s.give`, `s.onTop`, `s.setPool`, `s.script`, `s.scriptRefs`, `s.asked`, `s.alive`, `s.wantDamage`, `s.ptPlaceHook`, `stormKill`, `chooseAsset`, `wantDealt`, `nonCombatDamage`, `totalDamage`, `inDiscardPile`). The other cards may be synthetic. Every asset that a case doesn't mean to kill has enough health to survive. "Pre-damaged" means the test sets the asset's damage before the deploy.

- `TestCard_MAX8AutonomousWarsuit_EntersReservesAsFourSeven` (`FR-TRN-003`, `FR-SET-004`): it is a Readied 4/7 unit in player 0's Reserves, and the pool is empty. Deploying it to the Main Theater is rejected, and with R5 it can't be deployed; in both cases the card stays in hand.
- `TestCard_MAX8AutonomousWarsuit_AsksTwoSlotsDamageFirst` (`FR-TGT-001`, `SVY-R35`, `SVY-R3`): with `s.targetBoard()` and an opposing Warded unit, exactly two `DecideTarget` questions are asked of player 0, both before the trigger resolves. Each offers exactly both Headquarters, the six `targetBoard` assets (both Scratch Companies included) and the Warsuit itself, and not the opposing Warded unit. The answer to the first question is the one dealt 2 damage, and the answer to the second the one healed: choosing an opposing unit first and a pre-damaged (2) allied unit second gives the opposing unit 2 damage, one non-combat damage instance from the Warsuit, and leaves the allied unit with 0 damage. No other asset's damage changes.
- `TestCard_MAX8AutonomousWarsuit_AlliedWardedUnitIsOffered` (`FR-KWD-002h`): an allied Warded unit is offered by both questions, and an opposing Warded unit by neither.
- `TestCard_MAX8AutonomousWarsuit_SameAssetForBothTakesTwoThenHealsTwo` (`SVY-R35`, `FR-LAY-002`): choosing the same opposing unit with health 2 for both slots leaves it on the battlefield with 0 damage after the trigger resolves. It took 2 (an `EventDamage` of 2 is logged before the `EventHeal` of 2), and the death sweep, which runs only after the whole trigger, finds it alive.
- `TestCard_MAX8AutonomousWarsuit_HealsAnotherAssetLethallyDamagedOneDies` (`FR-LAY-002`, `FR-KWD-002d`): choosing an opposing unit with health 2 for the damage and a pre-damaged (3) allied unit for the heal destroys the opposing unit at the sweep (its card is in player 1's discard pile) and leaves the allied unit with 1 damage.
- `TestCard_MAX8AutonomousWarsuit_HealNeverGoesBelowZero` (`FR-KWD-002d`): a pre-damaged (1) asset chosen for the heal ends with 0 damage, and an undamaged one changes nothing (`EventHeal` with amount 0, as `TestCard_SpoilForestColossus_HealingAnUndamagedAssetChangesNothing` checks).
- `TestCard_MAX8AutonomousWarsuit_CanHealAHeadquarters` (`FR-KWD-002d`): player 0's Headquarters, pre-damaged by 3, chosen for the heal ends with 1 damage. The damage slot here takes player 1's Headquarters, whose remaining Health drops by 2, and the match isn't over.
- `TestCard_MAX8AutonomousWarsuit_CanDamageItself` (`FR-TGT-001`): choosing the Warsuit for the damage and another asset for the heal leaves the Warsuit on the battlefield with 2 damage.
- `TestCard_MAX8AutonomousWarsuit_ScratchCompanyCanBeKilled` (`FR-ZON-005`, `FR-LAY-002`): choosing an opposing Scratch Company (a 1/1) for the damage destroys it at the sweep.
- `TestCard_MAX8AutonomousWarsuit_ArmorAbsorbsAndHealLeavesArmor` (`FR-REP-004` Stage 4, `FR-KWD-002k`): a chosen opposing unit with printed Armor 1 loses its Armor and takes 1 damage. Then, choosing the same unit for both slots, a unit with Armor 1 and pre-damage 1 loses its Armor, takes 1 and heals 2, ending with 0 damage and 0 Armor (heal leaves Armor alone and never restores it).
- `TestCard_MAX8AutonomousWarsuit_TimeoutTakesTheFirstOptionForEach` (`FR-MOD-001` §6, `SVY-R16`): both target decisions timed out (`Answer{Timeout: true}` twice) each take their first option in fixed order, as in `TestCard_EarthshakerColossus_TimeoutTakesTheFirstOption` (player 0's own Headquarters). With player 0's Headquarters pre-damaged by 1, it ends with 1 damage: it took 2 and healed 2.
- `TestCard_MAX8AutonomousWarsuit_DamageTargetGoneStillHeals` (`FR-TGT-003` Partial Invalidation, `SVY-R25`): the damage target is removed between the trigger's placement and its resolution (`s.ptPlaceHook` with `stormKill`, as in `TestCard_EarthshakerColossus_TargetGoneDealsNothing`). No asset is dealt damage, and the pre-damaged (2) heal target ends with 0 damage.
- `TestCard_MAX8AutonomousWarsuit_HealTargetGoneStillDamages` (`FR-TGT-003` Partial Invalidation): the heal target is removed in the same window. The damage target still takes 2.
- `TestCard_MAX8AutonomousWarsuit_SourceGoneStillResolves` (`FR-ZON-004`): the Warsuit is removed between its trigger's placement and resolution (as in `TestCard_EarthshakerColossus_SourceGoneStillDeals`). The chosen opposing unit still takes 2 damage from the Warsuit, the pre-damaged (2) heal target still ends with 0 damage, and the Warsuit's card is in player 0's discard pile.
- `TestCard_MAX8AutonomousWarsuit_FaceDownScratchDoesNothing` (`FR-ZON-005`): with a *MAX-8 Autonomous Warsuit* on top of player 0's deck (`s.onTop`) and an opposing unit, player 0 enacts the real *Growing Legion* (as `TestCard_EarthshakerColossus_FaceDownScratchDoesNothing` does). A face-down Scratch Company enters player 0's Reserves, no target is asked, and no asset's damage changes.
- **Existing tests:** every other test passes unchanged, including `TestCards_Executable` and `TestDefs_MatchRulesText` with the name removed from `NotYetExecutable`.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected.

## Allowed Files
- `pkg/engine/cards/soulless.go` (the definition)
- `pkg/engine/cards/cards.go` (remove the name from `NotYetExecutable`)

## Done When
- [ ] The card tests pass: `go test ./pkg/engine -run TestCard_MAX8AutonomousWarsuit`
- [ ] `go test ./pkg/cardcheck` passes with the name removed
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-matrix` exits 0 (just over two minutes: use a timeout above the two-minute default)

## Log
- 2026-11-26 lead-architect (wave A run, session 23): written against e6573b07, with the lead's probe above. `main` has no commits the run branch lacks. Board read (D2.0): no directive is due here (#185 has only (4)'s *Awaken Nightmare* half left, wave F), and no waiting trigger has fired: the card has no each-player effect (#71), no mode or "another" slot (#91: its two slots may pick the same asset, `SVY-R35`, so no `Another`), no deck search (#106), no discard cost (#120), no tuck (#142), no turn-boundary or draw hook (#15 (2)), and no pile (#183 (2)); the mode is still `serial` (#6, #7), and no `needsRework` row exists (#184). Taken next in the survey's Soulless deck order, after *Inspired Salvage*. Searched `pkg/engine` and the task archives: every part of the Definition exists (`chooseTargets` asks one decision per slot in slot order, `stepDealDamage` and `stepHeal` act on `SubjectTarget` by slot, and the sweep runs only after the item resolves), so no kernel change. Set to `tests`.
- 2026-11-26 rules-test-author: wrote `pkg/engine/cards_soulless_max8_test.go`, 15 `TestCard_MAX8AutonomousWarsuit_*` tests, one per case under Tests (EntersReservesAsFourSeven, AsksTwoSlotsDamageFirst, AlliedWardedUnitIsOffered, SameAssetForBothTakesTwoThenHealsTwo, HealsAnotherAssetLethallyDamagedOneDies, HealNeverGoesBelowZero, CanHealAHeadquarters, CanDamageItself, ScratchCompanyCanBeKilled, ArmorAbsorbsAndHealLeavesArmor, TimeoutTakesTheFirstOptionForEach, DamageTargetGoneStillHeals, HealTargetGoneStillDamages, SourceGoneStillResolves, FaceDownScratchDoesNothing). Helpers requested: none (a file-local `mxChoose` picks per-slot answers). `go vet` is clean and `cardtown lint rules` passes. 14 tests fail until the card is defined; FaceDownScratchDoesNothing passes already (a Scratch Company has no ability, so nothing happens either way). No cases left out. Set to `tests-review`.
- 2026-11-26 lead-architect: tests approved (D4, round 1). 8c48b370 adds the 15 tests under Tests and changes nothing else. With the lead's probe Definition all 15 pass, and each of the 11 lead mutants fails 1 to 13 of them (the two slots' targets swapped, one slot for both steps, heal 1, damage 1, the heal step before the damage step, `ScratchExcluded` on both slots, units only on the heal slot, opposing only on the damage slot, `Another: 1` on the heal slot, `UpTo: 1` on the heal slot, `HookAttacks` in place of `HookDeploy`). The expected outcomes match the text they cite: `FR-TGT-001` line 432 (`SVY-R35`: two slots, the same asset allowed), `FR-LAY-002` line 535 (no sweep between the steps, so the same 2-health unit survives), `FR-KWD-002d` (heal never below 0, Armor untouched), `FR-KWD-002h`, `FR-TGT-003` Partial Invalidation, `FR-MOD-001` §6 (`SVY-R16`: each timeout takes player 0's Headquarters), `FR-ZON-004` and `FR-ZON-005`. `FaceDownScratchDoesNothing` passes before the implementation too, as a Scratch Company has no ability. Set to `ready`.
