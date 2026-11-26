---
id: T-0313
title: Make Culverin Brigade executable (deploy trigger, 1 damage to each of up to 2 chosen opposing assets; wave A, card, flash)
kind: card
tier: flash
wave: A
status: in-progress
branch: m3/autoloop-wave-a/T-0313-culverin-brigade
depends_on: [T-0312]
claims: [card:Culverin Brigade]
cards: [Culverin Brigade]
requirements: [FR-TRN-003, FR-SET-004, FR-TGT-001, FR-TGT-002, FR-TGT-003, FR-MOD-001, FR-KWD-002h, FR-LAY-002, FR-REP-004, FR-ZON-004, FR-ZON-005, FR-CBT-003]
---

## Goal
*Culverin Brigade* (Soulless unit, Infantry, R4 T0, Reserves, power 5, health 4; two copies in `data/decks/soulless-starter.txt`): *"When this unit deploys, deal 1 damage to each of up to 2 chosen opposing assets."* It only combines shapes that are `done`: a `HookDeploy` trigger (*Conscript Rifles*) with one `UpTo` 2 target slot (*Craterfall Ancient*'s slot, `SVY-R11`) whose filter is `Allegiance: ability.AllegianceOpposing`, and one `OpDealDamage` 1 step on that slot (`Subject: ability.SubjectTarget`), which deals one damage instance to each chosen asset. The slot has no `Kinds` and no `Scratch` criterion, so the opposing Headquarters, structures and face-down Scratch Companies may be chosen. An "up to N" slot is always asked, its picks are distinct, and a timeout picks none (`FR-MOD-001` §6). The trigger never needs a target (`FR-TGT-003`, Deployment Triggers Without Legal Targets).

Covers:
- The definition in `soullessDefs` (`pkg/engine/cards/soulless.go`), as under Definition below.
- The name off `cards.NotYetExecutable`.

## Inputs
- Card (`data/cards/units-and-structures-soulless.csv` line 20): Soulless unit, Infantry, Requisition cost 4, Tactics cost 0, Reserves, power 5, health 4.
- Survey row: `docs/design/survey/m3_full_catalog_ability_survey.md` §6, "Soulless: deck" (*Culverin Brigade*, line 170).
- FRD 01 (`docs/frd/01_Game_Rules_FRD.md`): `FR-TRN-003` (line 393), `FR-SET-004` (line 201), `FR-TGT-001` (line 407), `FR-TGT-002` (lines 426-428: "Up To N" Targets Are Distinct, `SVY-R11`), `FR-TGT-003` (lines 435-442: two-stage verification, Partial Invalidation, Deployment Triggers Without Legal Targets), `FR-MOD-001` §1 (line 449, the Single-Option Fast-Path, which an "up to" choice never takes because Min 0 differs from Max) and §6 (line 463: an *"up to N"* selection times out to none), `FR-KWD-002h` Ward (line 648), `FR-LAY-002` (line 534), `FR-REP-004` (line 562), `FR-ZON-004` (line 278, last-known state), `FR-ZON-005` (line 287), `FR-CBT-003` (line 630).
- Rulings that apply: `SVY-R3` (targets chosen on placement), `SVY-R11` (up to N, distinct), `SVY-R25` (a removed target is no longer that target).
- Design note: `docs/architecture/m3-full-catalog.md`, Boundaries (`pkg/engine/cards`), Deliverable 5.

### Definition (the engineer adds it to `soullessDefs`)
- `"Culverin Brigade"`, `Text`: `"When this unit deploys, deal 1 damage to each of up to 2 chosen opposing assets."` (one string).
- One `Trigger`: `Hook: ability.HookDeploy`. `Effect`:
  - `Targets`: one slot, `Filter: ability.Filter{Allegiance: ability.AllegianceOpposing}`, `UpTo: 2`.
  - `Steps`: `{Op: ability.OpDealDamage, Amount: ability.Value{N: 1}, Subject: ability.SubjectTarget, Target: 0}`.

A lead probe with the definition on the run branch at 080ed9ac, since reverted, passed `go tool task check`, `go tool task sim` (P0 58.5% / P1 39.5% / draws 2.0%, unchanged) and `go tool task sim-matrix`.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
In a new file `pkg/engine/cards_soulless_culverin_test.go`, from the CSV text, loading the real card (`s.card("Culverin Brigade")`), in the style of `cards_soulless_conscript_test.go` (a deploy trigger on a Soulless unit) and the *Craterfall Ancient* tests in `cards_animus_test.go` (an "up to 2" slot). Player 0 deploys it from hand to Reserves with exactly R4 T0 and is the active player, unless a case says otherwise. Answer the target decision with `s.scriptPlan` and `pick(...)` or `timeoutPick` (`optional_upto_test.go`). Reuse the existing test helpers (`s.give`, `s.setPool`, `s.put`, `s.putScratch`, `s.onTop`, `s.do`, `s.reject`, `s.asked`, `s.alive`, `s.wantDamage`, `s.ptPlaceHook`, `stormKill`, `wantOptions`, `wantDealt`, `nonCombatDamage`, `damageEvents`, `totalDamage`, `inDiscardPile`, `enactAction`, `faceDownCount`, `unit`, `structure`). The other cards may be synthetic. Every opposing asset that a case doesn't mean to kill has enough health to survive 1 damage.

- `TestCard_CulverinBrigade_EntersReservesAsFiveFour` (`FR-TRN-003`, `FR-SET-004`): after deploying with R4 it is a 5/4 in player 0's Reserves, Readied, and the pool is empty. Deploying it to the Main Theater is rejected and the card stays in hand.
- `TestCard_CulverinBrigade_DealsOneToEachOfTwoChosen` (`FR-TGT-002`, `FR-LAY-002`, `SVY-R11`): with two opposing units and a third beside them, deploying asks player 0 one `DecideTarget` with Min 0 and Max 2. Choosing the two leaves each with 1 damage (two non-combat damage instances, each from *Culverin Brigade*), and the third undamaged.
- `TestCard_CulverinBrigade_OffersEveryOpposingAsset` (`FR-TGT-001`, `FR-KWD-002h`, `FR-ZON-005`): with an allied and an opposing unit in each of the Reserves and the Main Theater, an allied and an opposing structure, a face-down Scratch Company on each side (`s.putScratch`) and an opposing Warded unit, the decision offers exactly player 1's Headquarters, the two opposing units, the opposing structure and the opposing Scratch Company. No allied asset is offered, *Culverin Brigade* and player 0's Headquarters included, and neither is the Warded unit.
- `TestCard_CulverinBrigade_MayChooseOne` (`SVY-R11`): choosing one opposing unit of two deals it 1 damage, one instance in all, and the other none.
- `TestCard_CulverinBrigade_MayChooseNone` (`FR-TGT-002`): answering with no pick deals no damage to anything, and *Culverin Brigade* stays on the battlefield.
- `TestCard_CulverinBrigade_TimeoutChoosesNone` (`FR-MOD-001` §6): a timed-out target decision (`timeoutPick`) deals no damage to anything.
- `TestCard_CulverinBrigade_OnlyTheHeadquartersStillAsks` (`FR-MOD-001` §1, `FR-CBT-003`): with no opposing asset but player 1's Headquarters, deploying still asks (Min 0, Max 1, the one option). Choosing it lowers player 1's Headquarters' remaining Health by 1, and the match isn't over.
- `TestCard_CulverinBrigade_LethalDamageKills` (`FR-LAY-002`): a chosen opposing unit with health 1 is destroyed and its card is in player 1's discard pile. The other chosen asset takes 1 damage.
- `TestCard_CulverinBrigade_ArmorAbsorbs` (`FR-REP-004` Stage 4, `FR-KWD-002k`): a chosen opposing unit with printed Armor 1 loses its Armor, and its Health is undamaged.
- `TestCard_CulverinBrigade_OneTargetGoneTheOtherStillTakesOne` (`FR-TGT-003` Partial Invalidation, `SVY-R25`): with two opposing units chosen, one is removed between the trigger's placement and its resolution (`s.ptPlaceHook` with `stormKill`, as in `TestCard_ConscriptRifles_TargetGoneDealsNothing`). The other takes 1 damage, and exactly one non-combat damage instance is logged.
- `TestCard_CulverinBrigade_TargetGainsWardDealsNothingToIt` (`FR-TGT-003`, `FR-KWD-002h`): with two opposing units chosen, one gains Ward between placement and resolution (`s.ptPlaceHook`, setting `Inst.Card.Keywords.Ward = true` on a synthetic card, as in `attack_reverify_test.go`). It takes no damage, and the other takes 1.
- `TestCard_CulverinBrigade_SourceGoneStillDeals` (`FR-ZON-004`, `FR-TGT-003`): *Culverin Brigade* is removed between its trigger's placement and resolution (`s.ptPlaceHook` with `stormKill` on it). The chosen opposing unit still takes 1 damage, and the card is in player 0's discard pile.
- `TestCard_CulverinBrigade_OnlyItsOwnDeployTriggers` (`FR-TRN-003`, `FR-ZON-005`): an attack by *Culverin Brigade* (`s.put` in the Main Theater, Readied) asks player 0 no target. A face-down Scratch Company made from a *Culverin Brigade* card on top of player 0's deck (`s.onTop`, then *Growing Legion*, as in `TestCard_ConscriptRifles_FaceDownScratchDoesNothing`) asks nothing and deals no damage.
- **Existing tests:** every other test passes unchanged, including `TestCards_Executable` with the name removed.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None requested. File-local helpers (cbDeploy, cbWantOffered, cbOneQuestion, cbFoe) are in the new test file.

## Allowed Files
- `pkg/engine/cards/soulless.go` (the definition)
- `pkg/engine/cards/cards.go` (remove the name from `NotYetExecutable`)

## Done When
- [ ] The card tests pass: `go test ./pkg/engine -run TestCard_CulverinBrigade`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-matrix` exits 0 (just over two minutes: use a timeout above the two-minute default)
- [ ] The card loads with no unsupported lines (`TestCards_Executable` passes with the name removed)

## Log
- 2026-11-25 lead-architect (wave A run, session 15): written against 080ed9ac, with the lead's probe above. `main` has no commits the run branch lacks. Board read (D2.0): no directive is due here (#185 has only (4)'s *Awaken Nightmare* half left, wave F), and no waiting trigger has fired: the card has no each-player effect (#71), no cost on a mode (#91 (2)), no deck search (#106), no mandatory discard (#120), no tuck (#142), no pile round trip (#183 (2)) and no draw or end-of-turn hook (#15 (2)), and no `needsRework` row exists (#184). Taken next in the survey's Soulless deck order, after *Conscript Rifles*. Searched `pkg/engine` and the task archives: every part of the Definition exists (*Conscript Rifles*' `HookDeploy` trigger, *Craterfall Ancient*'s `UpTo` 2 slot, the `AllegianceOpposing` filter, `OpDealDamage` on `SubjectTarget`), so no kernel change. Set to `tests`.
- 2026-11-25 rules-test-author: wrote 12 tests in pkg/engine/cards_soulless_culverin_test.go: TestCard_CulverinBrigade_{EntersReservesAsFiveFour, DealsOneToEachOfTwoChosen, OffersEveryOpposingAsset, MayChooseOne, MayChooseNone, TimeoutChoosesNone, OnlyTheHeadquartersStillAsks, LethalDamageKills, ArmorAbsorbs, OneTargetGoneTheOtherStillTakesOne, TargetGainsWardDealsNothingToIt, SourceGoneStillDeals, OnlyItsOwnDeployTriggers}. Helpers requested: none. Cases left out: none. go vet clean, cardtown lint rules passes; the tests fail now because the card is not yet executable, except OnlyItsOwnDeployTriggers, which passes until the definition exists (it pins that nothing else triggers). Set to tests-review.
- 2026-11-25 lead-architect: tests approved (D4, round 1). cafcbd23 adds only `pkg/engine/cards_soulless_culverin_test.go` (13 tests, one per case; the entry above says 12 by mistake) and the task file. Each expected outcome agrees with the FRD 01 text it cites: `FR-TGT-002` and `SVY-R11` for up to 2 distinct picks (one, two or none), `FR-MOD-001` §6 for the timeout taking none, §1 for the one-option case still asking (Min 0, Max 1), `FR-TGT-001`, `FR-KWD-002h` and `FR-ZON-005` for the opposing-asset filter (the opposing Headquarters, structure and Scratch Company offered, Ward and every allied asset not), `FR-TGT-003` for Partial Invalidation and for Ward gained before resolution, `FR-ZON-004` for the source gone, and `FR-LAY-002`, `FR-REP-004` Stage 4 and `FR-CBT-003` for the damage. A lead probe with the Definition, since reverted, passes all 13, and each of 10 mutants (any allegiance, allied, units only, Scratch excluded, `UpTo` 1, exactly one, `UpTo` 3, 2 damage, attack hook, `Optional`) fails at least one. Set to `ready`.
