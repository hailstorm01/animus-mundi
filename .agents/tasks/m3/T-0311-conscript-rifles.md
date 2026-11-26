---
id: T-0311
title: Make Conscript Rifles executable (deploy trigger, 1 damage to a chosen non-Scratch unit; wave A, card, flash)
kind: card
tier: flash
wave: A
status: in-progress
branch: m3/autoloop-wave-a/T-0311-conscript-rifles
depends_on: [T-0310]
claims: [card:Conscript Rifles]
cards: [Conscript Rifles]
requirements: [FR-TRN-003, FR-SET-004, FR-TGT-001, FR-TGT-003, FR-REP-004, FR-LAY-002, FR-ZON-002, FR-ZON-004, FR-ZON-005, FR-ENT-103, FR-KWD-002h, FR-KWD-002k]
---

## Goal
*Conscript Rifles* (a Soulless Infantry unit, R2 T0, 2/2, Reserves; in the Soulless starter deck, 4 copies): *"When this unit deploys, deal 1 damage to chosen non-Scratch Company unit."* It only combines shapes that are `done`: a `HookDeploy` trigger (*Stormbolt Ray*, *Will Breaker*) with one target slot whose filter is `Kinds: ability.KindUnit` and `Scratch: ability.ScratchExcluded`, and one `OpDealDamage` step on that target (*Bring It Down*'s step with 1 in place of 2). The slot has no `Other`, so *Conscript Rifles* may choose itself (survey row: "May choose itself"), and when it is the only non-Scratch unit on the battlefield it must.

Covers:
- The definition in `soullessDefs` (`pkg/engine/cards/soulless.go`), as under Definition below.
- The name off `cards.NotYetExecutable`.

## Inputs
- Card (`data/cards/units-and-structures-soulless.csv` line 6): Soulless Infantry unit, R2 T0, power 2, health 2, Reserves.
- Survey row: `docs/design/survey/m3_full_catalog_ability_survey.md` §6, "Soulless: deck" (*Conscript Rifles*, line 169).
- FRD 01 (`docs/frd/01_Game_Rules_FRD.md`): `FR-TRN-003` (line 394), `FR-SET-004` Immediate Readiness (line 203), `FR-TGT-001` (line 407), `FR-TGT-003` (lines 435-442, including Deployment Triggers Without Legal Targets at line 442), the Single-Option Fast-Path (line 449), `FR-LAY-002` (line 534), `FR-REP-004` (line 562; Stage 2 Resistant, Stage 4 Armor at line 587), `FR-KWD-002h` Ward (line 648), `FR-KWD-002k` Armor (line 651), `FR-ZON-002` (line 271, Destruction at line 276), `FR-ZON-004` (line 279), `FR-ZON-005` (line 288), `FR-ENT-103` (line 298).
- Rulings that apply: `SVY-R25` (a removed target is no longer that target), `SVY-R26` (presence only between the two verifications).
- Design note: `docs/architecture/m3-full-catalog.md`, Boundaries (`pkg/engine/cards`), Deliverable 5.

### Definition (the engineer adds it to `soullessDefs`)
- `"Conscript Rifles"`, `Text`: `"When this unit deploys, deal 1 damage to chosen non-Scratch Company unit."` (one string).
- One `Trigger`: `Hook: ability.HookDeploy`. `Effect`:
  - `Targets`: one slot, `Filter: ability.Filter{Kinds: ability.KindUnit, Scratch: ability.ScratchExcluded}`.
  - `Steps`: `{Op: ability.OpDealDamage, Amount: ability.Value{N: 1}, Subject: ability.SubjectTarget, Target: 0}`.

A lead probe with the definition on the run branch at e298f225, since reverted, passed `go tool task check`, `go tool task sim` (P0 58.5% / P1 39.5% / draws 2.0%, unchanged) and `go tool task sim-matrix`.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
In a new file `pkg/engine/cards_soulless_conscript_test.go`, from the CSV text, loading the real card (`s.card`), in the style of `cards_concordia_willbreaker_test.go` and `cards_concordia_stormbolt_test.go`. Player 0 deploys *Conscript Rifles* from hand to Reserves (pool R2) and is the active player, unless the case says otherwise. Reuse the existing test helpers (`s.deployTo`, `s.setPool`, `s.put`, `s.putScratch`, `s.give`, `s.onTop`, `s.script`, `s.asked`, `s.alive`, `s.ptPlaceHook`, `stormChoose`, `stormKill`, `effectEvents`, `targetAsk`, `wantOptions`, `unit`, `structure`, `morphManeuver`/`morphEnact` from `morph_scratch_test.go`). The other cards may be synthetic.

Remember the Single-Option Fast-Path (line 449): when exactly one unit is legal, nothing is asked. Every case where *Conscript Rifles* is the only non-Scratch unit on the battlefield is such a case, because it is always a legal target of its own trigger.

- `TestCard_ConscriptRifles_EntersReservesAsTwoTwo` (`FR-TRN-003`, `FR-SET-004`): deployed from hand for R2 it is a Readied unit in player 0's Reserves with power 2 and Maximum Health 2. Deploying it to the Main Theater is rejected, and with R1 it can't be deployed.
- `TestCard_ConscriptRifles_AloneDealsOneToItself` (`FR-TGT-001`, Single-Option Fast-Path): with no other unit on the battlefield (Headquarters only), nothing is asked, the stack is empty once play returns, and *Conscript Rifles* is alive with 1 damage (Current Health 1). Neither Headquarters is dealt damage.
- `TestCard_ConscriptRifles_DealsOneToTheChosenUnit` (`FR-TGT-001`, `FR-LAY-002`): with one opposing unit of health 3 in player 1's Main Theater, the target decision offers exactly it and *Conscript Rifles*. Choosing the opposing unit leaves it with 1 damage, and *Conscript Rifles* is undamaged.
- `TestCard_ConscriptRifles_OffersEveryNonScratchUnit` (`FR-TGT-001`, `FR-ZON-005`): with an allied and an opposing unit in each of the Reserves and the Main Theater, an allied and an opposing structure, and a face-down Scratch Company on each side (`s.putScratch`), the target decision offers exactly the four units and *Conscript Rifles*. No structure, Headquarters or Scratch Company is offered.
- `TestCard_ConscriptRifles_UnitTurnedFaceDownNotOffered` (`FR-ENT-103`, `FR-TGT-001`): an opposing unit turned face down (`morphEnact` with `morphManeuver()`) before *Conscript Rifles* deploys is a Scratch Company, so it is not offered. A face-up opposing unit beside it is, with *Conscript Rifles*.
- `TestCard_ConscriptRifles_LethalDamageKills` (`FR-LAY-002`, `FR-ZON-002`): a chosen opposing unit with health 1 is destroyed, and its card is in player 1's discard pile.
- `TestCard_ConscriptRifles_ResistantAndArmorApply` (`FR-REP-004`, `FR-KWD-002k`): a chosen unit with Resistant 1 takes no damage. A chosen unit with Armor 1 loses its Armor and takes no damage to its Health.
- `TestCard_ConscriptRifles_WardedOpposingUnitNotOffered` (`FR-KWD-002h`): an opposing Warded unit is not offered (with only it and *Conscript Rifles*, nothing is asked and *Conscript Rifles* takes the damage; the Warded unit is undamaged). An allied Warded unit is offered (Ward stops only opponents).
- `TestCard_ConscriptRifles_TargetGoneDealsNothing` (`FR-TGT-003`, `SVY-R25`): the chosen opposing unit is removed between the trigger's placement and its resolution (`s.ptPlaceHook`, which fires for an ability on the stack, with `stormKill`). Nothing else is dealt damage, *Conscript Rifles* included.
- `TestCard_ConscriptRifles_SourceGoneStillDeals` (`FR-ZON-004`): *Conscript Rifles* is removed between its trigger's placement and resolution (as in `TestCard_StormboltRay_SourceGoneStillDeals`). The chosen opposing unit is still dealt 1, and *Conscript Rifles*' card is in player 0's discard pile.
- `TestCard_ConscriptRifles_FaceDownScratchDoesNothing` (`FR-ZON-005`): with a *Conscript Rifles* on top of player 0's deck (`s.onTop`) and an opposing unit, player 0 enacts the real *Growing Legion*. A face-down Scratch Company enters player 0's Reserves, nothing is asked about a target, and no unit is dealt damage.
- **Existing tests:** every other test passes unchanged, including `TestCards_Executable` with the name removed.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected.

## Allowed Files
- `pkg/engine/cards/soulless.go` (the definition)
- `pkg/engine/cards/cards.go` (remove the name from `NotYetExecutable`)

## Done When
- [ ] The card tests pass: `go test ./pkg/engine -run TestCard_ConscriptRifles`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-matrix` exits 0 (just over two minutes: use a timeout above the two-minute default)
- [ ] The card loads with no unsupported lines (`TestCards_Executable` passes with the name removed)

## Log
- 2026-11-25 lead-architect (wave A run, session 13): written against e298f225, with the lead's probe above. `main` has no commits the run branch lacks. Board read (D2.0): no directive is due here (#185 has only (4)'s *Awaken Nightmare* half left, wave F), and no waiting trigger has fired. Searched `pkg/engine` and the task archives: every part of the Definition exists (the `HookDeploy` trigger of *Will Breaker*, the `KindUnit` filter of *Icewind Regent*, `ScratchExcluded` and the step of *Bring It Down*), so no kernel change. Set to `tests`.
- 2026-11-25 rules-test-author: wrote `pkg/engine/cards_soulless_conscript_test.go` with the 11 tests listed under Tests (EntersReservesAsTwoTwo, AloneDealsOneToItself, DealsOneToTheChosenUnit, OffersEveryNonScratchUnit, UnitTurnedFaceDownNotOffered, LethalDamageKills, ResistantAndArmorApply, WardedOpposingUnitNotOffered, TargetGoneDealsNothing, SourceGoneStillDeals, FaceDownScratchDoesNothing). No helpers requested (file-local `cr*` helpers only). Option lists are compared as sets, since FRD 01 fixes no order. `go vet` and `cardtown lint rules` are clean; 10 tests fail as expected until the definition exists, and FaceDownScratchDoesNothing already passes as a guard (the card has no ability yet, and a Scratch Company never gets one). No case left out.
- 2026-11-25 lead-architect: tests approved (D4, round 1). a91bf531 adds only `pkg/engine/cards_soulless_conscript_test.go` (11 tests, one per case) and the Log line. Each expected outcome agrees with the FRD 01 text it cites (the Single-Option Fast-Path for the self-target cases, `FR-TGT-001` and `FR-ENT-103` for the Scratch exclusion, `FR-REP-004` Stages 2 and 4 for Resistant and Armor, `FR-KWD-002h` for Ward, `FR-TGT-003` and `SVY-R25` for the target gone, `FR-ZON-004` for the source gone). A lead probe with the Definition, since reverted, passes all 11, and each of 7 mutants (no Scratch filter, any asset kind, `Other`, 2 damage, opposing only, Main Theater only, kill in place of damage) fails at least one. Set to `ready`.
