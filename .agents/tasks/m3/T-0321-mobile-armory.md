---
id: T-0321
title: Make Mobile Armory executable (exhaust: chosen allied unit in the same zone +1 power until end of turn, wave A, card, flash)
kind: card
tier: flash
wave: A
status: review
branch: m3/autoloop-wave-a/T-0321-mobile-armory
depends_on: [T-0320]
claims: [card:Mobile Armory]
cards: [Mobile Armory]
requirements: [FR-TRN-003, FR-SET-004, FR-ACT-001, FR-ACT-002, FR-TGT-001, FR-TGT-003, FR-ZON-001, FR-ZON-004, FR-LAY-003, FR-MOD-001, FR-KWD-002h]
---

## Goal
*Mobile Armory* (Soulless unit, Mechanized, R3 T0, Main Theater or Reserves, power 1, health 3; three copies in `data/decks/soulless-starter.txt`): *"Exhaust: Chosen allied unit in the same zone gets +1 power until end of turn."* It only combines shapes that are `done`: *Icewind Regent*'s `Activated` ability with a lone `CostExhaustSelf` cost (`pkg/engine/cards/concordia.go`), one exactly-one target slot whose filter is allied units with `ZoneRef: ability.ZoneRefSame` (the activated same-zone slot that `TestFR_ACT_001_SameZoneActivatedAbility` pins, `pkg/engine/same_zone_test.go:161`; T-0246), and *Chainbreaker's Tenacity*'s `OpModifyStat` +1 power on `SubjectTarget` with the zero-value `UntilEndOfTurn` duration (`pkg/engine/cards/path.go`). The zone is the Armory's zone when the ability is placed, locked with the target (`SVY-R44`): for Reserves, its controller's own (`FR-ZON-001`). There is no "other", so the Armory may choose itself (survey row: "May choose itself"), and the ability can always be activated while the Armory is Readied and the stack is empty (`SVY-R36`).

Covers:
- The definition in `soullessDefs` (`pkg/engine/cards/soulless.go`), as under Definition below.
- The name off `cards.NotYetExecutable`.

## Inputs
- Card (`data/cards/units-and-structures-soulless.csv` line 15): Soulless unit, Mechanized, Requisition cost 3, Tactics cost 0, deploys to the Main Theater or Reserves, power 1, health 3.
- Survey row: `docs/design/survey/m3_full_catalog_ability_survey.md` §6, "Soulless: deck" (*Mobile Armory*, line 178: `SVY-R44`, "Built. May choose itself").
- FRD 01 (`docs/frd/01_Game_Rules_FRD.md`): `FR-TRN-003`, `FR-SET-004` (line 201), `FR-ZON-001` (line 257: each player's own Reserves), `FR-ZON-004` (line 278), `FR-ACT-001` and `FR-ACT-002` (line 388: clean stack, exhaust paid on declaration, an Exhausted asset can't pay an exhaust cost), `FR-TGT-001` (line 407; "In the Same Zone" Is Locked on Placement, `SVY-R44`, line 421), `FR-TGT-003` (line 435: two-stage verification against the locked zone), `FR-MOD-001` (line 448: the Single-Option Fast-Path, §1, and §6 timeouts), `FR-LAY-003` (line 540: until end of turn), `FR-KWD-002h` Ward.
- Rulings that apply: `SVY-R44` (the same zone is locked on placement, last-known if the source has left, and a target still in it at the second check stays legal), `SVY-R36` (an activated ability needs a legal entity for each exactly-one slot), `SVY-R3` (targets are chosen on placement), `SVY-R16` (timeouts take the first option in fixed order), `SVY-R25` (a removed target is no longer that target).
- Design note: `docs/architecture/m3-full-catalog.md`, Boundaries (`pkg/engine/cards`), Deliverable 5.

### Definition (the engineer adds it to `soullessDefs`, after *MAX-8 Autonomous Warsuit*)
- `"Mobile Armory"`, `Text`: `"Exhaust: Chosen allied unit in the same zone gets +1 power until end of turn."` (one string, as the CSV has it).
- One `Activated` entry:
  - `Costs`: `{Kind: ability.CostExhaustSelf}` only.
  - `Effect.Targets`: one slot, `Filter: ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit, ZoneRef: ability.ZoneRefSame}`.
  - `Effect.Steps`: `{Op: ability.OpModifyStat, Amount: ability.Value{N: 1}, Subject: ability.SubjectTarget, Target: 0}` (no `Stat`: power; no `Until`: end of turn).

A lead probe with the definition on the run branch at e0d72c38, since reverted, passed `go test ./pkg/cardcheck`, `go tool task check`, `go tool task sim` (P0 58.5% / P1 39.5% / draws 2.0%, unchanged) and `go tool task sim-matrix` (exit 0).

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
In a new file `pkg/engine/cards_soulless_armory_test.go`, from the CSV text, loading the real card (`s.card("Mobile Armory")`), in the style of `cards_soulless_macrocannon_test.go` (an activated `CostExhaustSelf` ability on a Soulless card, with `dmcTapCard`'s between-placement-and-resolution technique), `same_zone_test.go` (`TestFR_ACT_001_SameZoneActivatedAbility`) and the *Chainbreaker's Tenacity* tests in `cards_path_maneuvers_test.go` (+1 power, expiring at end of turn). Player 0 controls the Armory and is the active player, unless a case says otherwise. Reuse the existing test helpers (`s.put`, `s.putScratch`, `s.setPool`, `s.script`, `s.reject`, `s.alive`, `s.moveTo`, `s.endTurn`, `s.onTop`, `activate`, `offersActivate`, `wantOptions`, `chooseAsset`, `inDiscardPile`, `inAssets`). The other cards may be synthetic. Read power through `s.g.Power`.

Deploying:
- `TestCard_MobileArmory_EntersEitherZoneAsOneThree` (`FR-TRN-003`): deployed from hand with exactly R3, to the Main Theater in one case and to Reserves in another, it is a Readied 1/3 unit in that zone of player 0's, and the pool is empty. With R2 it can't be deployed, and the card stays in hand.

Activating (a Readied Armory put in player 0's Main Theater, pool R0 T0, unless the case says otherwise):
- `TestCard_MobileArmory_ChosenAllyGetsPlusOnePower` (`FR-ACT-001`, `FR-LAY-001`): with an allied power-2 unit in the Main Theater, activating and choosing it gives it power 3. The Armory is Exhausted, its own power is still 1, the pool is still R0 T0 (the ability costs no resources), and no other asset's power changed.
- `TestCard_MobileArmory_OffersAlliedUnitsInItsZoneOnly` (`FR-TGT-001`, `SVY-R44`, `FR-KWD-002h`): with an allied unit and an allied Warded unit in the Main Theater, an allied unit in player 0's Reserves, an allied structure in the Main Theater, an opposing unit and an opposing Warded unit in the Main Theater, and an opposing unit in player 1's Reserves, one exactly-one `DecideTarget` for player 0 offers exactly the two allied Main Theater units and the Armory itself. No Headquarters, structure, Reserves unit or opposing unit is offered.
- `TestCard_MobileArmory_FromReservesOffersOwnReservesOnly` (`SVY-R44`, `FR-ZON-001`): an Armory in player 0's Reserves, with an allied unit and an allied face-down Scratch Company (`s.putScratch`) in player 0's Reserves, an allied unit in the Main Theater and an opposing unit in player 1's Reserves, offers exactly the allied Reserves unit, the Scratch Company and the Armory. Choosing the Scratch Company gives it power 2.
- `TestCard_MobileArmory_MayChooseItself` (`FR-TGT-001`, survey row): with another allied unit in its zone, choosing the Armory gives it power 2, and it is Exhausted.
- `TestCard_MobileArmory_AloneInItsZoneBuffsItselfWithoutAsking` (`SVY-R36`, `FR-MOD-001` §1): with no other allied unit in its zone (an allied unit in player 0's Reserves and an opposing unit in the Main Theater), the ability is offered, no `DecideTarget` is asked (the Single-Option Fast-Path), and the Armory gets power 2.
- `TestCard_MobileArmory_PowerExpiresAtEndOfTurn` (`FR-LAY-003`): after the chosen ally's buff resolves, `s.endTurn()` returns it to its printed power.
- `TestCard_MobileArmory_BuffStaysWhenTheAllyMovesAfterwards` (`SVY-R44`, `FR-ZON-004`): once the ability has resolved, moving the buffed ally from the Main Theater to player 0's Reserves (`s.moveTo`) leaves it with +1 power: the zone is checked only on placement and at resolution, and temporary modifiers end only on leaving the battlefield or at their duration.
- `TestCard_MobileArmory_NotOfferedWhileExhausted` (`FR-ACT-002`): an Exhausted Armory with an allied unit beside it isn't offered, and a rejected activation (`s.reject`) changes no power and leaves it Exhausted.
- `TestCard_MobileArmory_NotOfferedWhileTheStackIsNotEmpty` (`FR-ACT-001` clean-stack requirement): while an item is on the stack, the ability isn't offered to the player who holds priority, as in `TestCard_DolusMacrocannon_NotOfferedWhileTheStackIsNotEmpty` (reuse its setup).
- `TestCard_MobileArmory_MayActivateTheTurnItDeploys` (`FR-SET-004`): deployed from hand to the Main Theater with R3 beside an allied unit, its ability is offered right away, and activating it on that unit gives it +1 power.
- `TestCard_MobileArmory_TimeoutTakesTheFirstOption` (`FR-MOD-001` §6, `SVY-R16`): with two other allied units in the Main Theater, a timed-out target decision (`s.script(0, Answer{Timeout: true})`) takes the decision's first option in fixed order, as in `TestCard_DolusMacrocannon_TimeoutTakesTheFirstOption`: that asset gets +1 power, and no other asset's power changes.

Between placement and resolution (`dmcTapCard`'s technique: a copy of the real card, its `Def` copied so the shared definition isn't changed, whose `Triggers` also hold one synthetic `HookExhausted` trigger placed on top of the ability; its step, or a hook at its placement, does the change):
- `TestCard_MobileArmory_TargetLeftTheZoneGetsNothing` (`FR-TGT-003`, `SVY-R44`): the chosen allied Main Theater unit moves to player 0's Reserves. It is illegal at the second check: its power is unchanged, and the Armory stays Exhausted.
- `TestCard_MobileArmory_SourceMovedStillBuffsInTheLockedZone` (`SVY-R44`, `FR-TGT-003`): the Armory moves itself to player 0's Reserves (`{Op: ability.OpMoveZone, Subject: ability.SubjectSelf, To: ability.ZoneReserves}`). The chosen ally, still in the Main Theater, gets +1 power.
- `TestCard_MobileArmory_SourceGoneStillBuffs` (`FR-ZON-004`, `SVY-R44`): the Armory is killed (`{Op: ability.OpKill, Subject: ability.SubjectSelf}`). The chosen ally still gets +1 power, and the Armory's card is in player 0's discard pile.
- `TestCard_MobileArmory_TargetGoneGetsNothing` (`FR-TGT-003`, `SVY-R25`): the chosen ally returns to its owner's hand. No asset's power changes.

- **Existing tests:** every other test passes unchanged, including `TestCards_Executable` and `TestDefs_MatchRulesText` with the name removed from `NotYetExecutable`.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected.

## Allowed Files
- `pkg/engine/cards/soulless.go` (the definition)
- `pkg/engine/cards/cards.go` (remove the name from `NotYetExecutable`)

## Done When
- [x] The card tests pass: `go test ./pkg/engine -run TestCard_MobileArmory`
- [x] `go test ./pkg/cardcheck` passes with the name removed
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks
- [x] `go tool task sim-matrix` exits 0 (just over two minutes: use a timeout above the two-minute default)

## Log
- 2026-11-26 lead-architect (wave A run, session 24): written against e0d72c38, with the lead's probe above. `main` has no commits the run branch lacks. Board read (D2.0): no directive is due here (#185 has only (4)'s *Awaken Nightmare* half left, wave F), and no waiting trigger has fired: the card has no each-player effect (#71), no mode, and its slot is not an "another" slot (#91: no "other", so it may choose itself), no deck search (#106), no discard cost (#120: its only cost is an exhaust), no tuck (#142), no turn-boundary or draw hook (#15 (2)), and no pile (#183 (2)); the mode is still `serial` (#6, #7), and no `needsRework` row exists (#184). Taken next in the survey's Soulless deck order, after *MAX-8 Autonomous Warsuit*. Searched `pkg/engine` and the task archives: every part of the Definition exists (`activatedPlacement` locks the zone through `sameZoneLock`, `legalTarget` reads the lock at both checks, and `stepModifyStat` acts on `SubjectTarget`), so no kernel change. Set to `tests`.
- 2026-11-26 rules-test-author: wrote `pkg/engine/cards_soulless_armory_test.go`, 16 tests: TestCard_MobileArmory_ EntersEitherZoneAsOneThree, ChosenAllyGetsPlusOnePower, OffersAlliedUnitsInItsZoneOnly, FromReservesOffersOwnReservesOnly, MayChooseItself, AloneInItsZoneBuffsItselfWithoutAsking, PowerExpiresAtEndOfTurn, BuffStaysWhenTheAllyMovesAfterwards, NotOfferedWhileExhausted, NotOfferedWhileTheStackIsNotEmpty, MayActivateTheTurnItDeploys, TimeoutTakesTheFirstOption, TargetLeftTheZoneGetsNothing, SourceMovedStillBuffsInTheLockedZone, SourceGoneStillBuffs, TargetGoneGetsNothing (16 listed cases; the deploy test covers both zones plus the R2 rejection). Helpers requested: none (file-local helpers maScene, maTapCard, maActivate, maHookScene). `go vet` clean, `cardtown lint rules` PASS; all fail now (card not yet executable). Notes: the three synthetic-trigger cases that move or return the other allied units use `Filter{Other: true, Zone: ability.ZoneMainTheater}` on a SubjectAll step so the Armory itself is untouched; the timeout test reads the first option from the decision rather than assuming an order. No case left out.
- 2026-11-26 lead-architect: tests approved (D4, round 1). 869350e8 adds the 16 tests under Tests in `pkg/engine/cards_soulless_armory_test.go` and changes nothing else but this file's status and Log. With the lead's probe Definition all 16 pass, and each of the 11 lead mutants fails 1 to 13 of them (no `ZoneRef`, `Zone: ZoneMainTheater` in place of `ZoneRefSame`, `Other: true`, no allegiance, no kind, +2 power, `UntilEndOfOpponentTurn`, `StatMaxHealth`, `SubjectSelf`, an added `CostResources` R1, `UpTo: 1`). The expected outcomes match the text they cite: `FR-TGT-001` line 421 (`SVY-R44`: the zone locked on placement, its controller's own Reserves under `FR-ZON-001`, last-known when the source is gone), `FR-TGT-003` (the second check against the locked zone), `FR-MOD-001` §1 (the Armory alone is taken without asking) and §6 (`SVY-R16`), `FR-ACT-001`, `FR-ACT-002`, `FR-SET-004`, `FR-LAY-003` and `FR-KWD-002h`. Set to `ready`.
- 2026-11-26 kernel-engineer: implemented Mobile Armory. Diff shows:
  - `pkg/engine/cards/soulless.go`: added "Mobile Armory" entry to `soullessDefs` with Text matching CSV rules text, Activated ability with CostExhaustSelf, single target slot for allied units with ZoneRef: ability.ZoneRefSame, and OpModifyStat +1 power on SubjectTarget.
  - `pkg/engine/cards/cards.go`: removed "Mobile Armory" from NotYetExecutable.
  - Reused: CostExhaustSelf (Icewind Regent / Dolus Refiner), ZoneRefSame target filtering, OpModifyStat (Chainbreaker's Tenacity).
  - Commands run and results:
    - `go test ./pkg/engine -run TestCard_MobileArmory`: PASS (16 tests)
    - `go test ./pkg/cardcheck`: PASS
    - `go tool task check`: PASS (gofmt, go vet, all unit tests, cardtown lint --all)
    - `go tool task sim`: PASS (P0 58.5% / P1 39.5% / draws 2.0%)
    - `go tool task sim-matrix`: PASS (all 24 starter and 24 stress runs terminated clean with exit 0)
  Set to `review`.
