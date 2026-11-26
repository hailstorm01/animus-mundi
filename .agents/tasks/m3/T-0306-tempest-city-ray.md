---
id: T-0306
title: Make Tempest City Ray executable (move an opposing unit to its controller's Reserves, wave A, card, flash, #185 (4))
kind: card
tier: flash
wave: A
status: tests
branch: m3/autoloop-wave-a/T-0306-tempest-city-ray
depends_on: [T-0305]
claims: [card:Tempest City Ray]
cards: [Tempest City Ray]
requirements: [FR-TRN-003, FR-TGT-001, FR-TGT-003, FR-ACT-001, FR-ACT-002, FR-ZON-002, FR-ZON-003, FR-ZON-004, FR-KWD-002h]
---

## Goal
*Tempest City Ray* (a Concordia structure, Organic, Tempest, R5 T0, 0/8, Main Theater, yields R1 T2; in the starter deck): *"T2, exhaust: Move chosen opposing unit to its controller's Reserves."* It only combines shapes that are `done`: an activated ability paid with Tactics and `CostExhaustSelf` (*Stormbolt Ray*, T-0304), one exactly-one opposing unit slot, and `OpMoveZone` to `ZoneReserves` on that target (*Ensnaring Mist*, `pkg/engine/cards/animus.go`). `stepMoveZone` already moves a unit to its controller's Reserves (`pkg/engine/effects.go`, `SVY-R65`), and skips a unit that is already there.

Covers:
- The definition in `pkg/engine/cards/concordia.go`, as under Definition below.
- The name off `cards.NotYetExecutable`.
- Directive #185 (4), its *Tempest City Ray* half: the definition uses the reworded text (the CSV already prints it), and a card test moves a unit controlled through *Rhapsody's Hollow* to its controller's Reserves. Its *Awaken Nightmare* half stays for that card's task (wave F).

## Inputs
- Card (`data/cards/units-and-structures-concordia.csv` line 25).
- Survey row: `docs/design/survey/m3_full_catalog_ability_survey.md` §6, "Concordia: deck" (*Tempest City Ray*).
- FRD 01: `FR-TRN-003` (line 394), `FR-TGT-001` (lines 407-413), `FR-TGT-003` (lines 435-442), `FR-ACT-001` and `FR-ACT-002` (lines 388-391), `FR-ZON-002` (lines 271-276), `FR-ZON-003` (line 250), `FR-ZON-004` (lines 278-283), `FR-SET-004` Immediate Readiness (line 203), `FR-KWD-002h` (line 648).
- Rulings: `SVY-R65` (a move to Reserves goes to the controller's Reserves, whoever owns the unit), `SVY-R57` (only the stated criteria count: "chosen opposing unit" has no zone, so an opposing unit already in Reserves is legal, and choosing it moves nothing), `SVY-R36` (an activated ability needs a legal entity for each exactly-one slot), `SVY-R1` (a move is not a deploy), `SVY-R7` (a control change keeps the unit's zone and state), `SVY-R2` (targets and costs are one declaration).
- Design note: `docs/architecture/m3-full-catalog.md`, Boundaries (`pkg/engine/cards`), Deliverable 5.
- Directive #185 (4) (`BOARD.md`).

### Definition (the engineer adds it to `concordiaDefs`)
- `"Tempest City Ray"`, `Text`: `"T2, exhaust: Move chosen opposing unit to its controller's Reserves."`.
- One `Activated`: `Costs` is `{Kind: ability.CostResources, Tac: 2}`, then `{Kind: ability.CostExhaustSelf}`. `Effect`: `Targets` is one exactly-one slot with `Filter: ability.Filter{Allegiance: ability.AllegianceOpposing, Kinds: ability.KindUnit}`; `Steps` is one `{Op: ability.OpMoveZone, To: ability.ZoneReserves, Subject: ability.SubjectTarget, Target: 0}`.

A lead probe with the definition on the run branch at aeb9f187, since reverted, passed `go tool task check` and `go tool task sim-matrix`.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
In a new file `pkg/engine/cards_concordia_tempest_ray_test.go`, from the CSV text, loading the real card (`s.card`), in the style of `cards_concordia_stormbolt_test.go` and `cards_concordia_regent_test.go`. Player 0 controls the Ray and is the active player, unless the case says otherwise. Reuse the existing test helpers (`deployTo`, `s.put`, `s.putControlled`, `s.script`, `wantOptions`, `activate`, `offersActivate`, `s.reject`, `s.setPool`, `s.alive`, `inAssets`, `effectEvents`, `s.endTurn`, `inHand`, `inDiscardPile`). The other cards may be synthetic.

Deploying:
- `TestCard_TempestCityRay_EntersMainTheaterAsZeroEight` (`FR-TRN-003`): deployed from hand for R5 it is a Readied 0/8 structure in the Main Theater; deploying it to Reserves is rejected, and with R4 it can't be deployed.

Activating (a Readied Ray put in player 0's Main Theater, pool R0 T2, unless the case says otherwise):
- `TestCard_TempestCityRay_MovesChosenOpposingUnitToReserves` (`FR-ZON-002`, `FR-ZON-004`, `SVY-R1`): an opposing Main Theater unit with 1 damage, 1 Armor and a +1 power modifier until end of turn, Exhausted, is the same asset (same ID) afterwards, in player 1's Reserves and not in the Main Theater, still Exhausted with its damage, Armor and modifier. One `EventMove` names it. Its synthetic deploy trigger does not fire, and its synthetic `HookMoves` trigger does. The pool is R0 T0, the Ray is Exhausted and stays on the battlefield, and no other asset moves. A table also covers a face-down Scratch Company of player 1's in the Main Theater: it moves and stays face down.
- `TestCard_TempestCityRay_OnlyOpposingUnitsOffered` (`FR-TGT-001`, `SVY-R57`, `FR-ZON-002`): one exactly-one `DecideTarget` for player 0 whose options are exactly player 1's units in the Main Theater and in Reserves, a face-down Scratch Company of each zone included. Not offered: player 0's units in either zone, a unit player 1 owns and player 0 controls (`s.putControlled(1, 0, …)`), any structure, either Headquarters, and the Ray itself.
- `TestCard_TempestCityRay_UnitAlreadyInReservesStaysPut` (`SVY-R57`, `FR-ZON-002`): choosing an opposing unit in player 1's Reserves is legal; afterwards it is still there, the same asset with its state, and no `EventMove` is logged. The costs are still paid (pool R0 T0, the Ray Exhausted).
- `TestCard_TempestCityRay_ControlledUnitGoesToItsControllersReserves` (`SVY-R65`, `FR-ZON-003`, `SVY-R7`; #185 (4)): player 1, the active player, activates a real *Rhapsody's Hollow* (`s.card("Rhapsody's Hollow")`, as in `cards_animus_rhapsodys_hollow_test.go`) to gain control of a unit player 0 owns in the Main Theater. Then the turn passes to player 0 (`s.endTurn()`, then set the pool and make sure the Ray is Readied), who activates the Ray on that unit. The unit ends in player 1's Reserves and not player 0's, still owned by player 0 and controlled by player 1. Use the Hollow itself; don't stand in `s.putControlled` for it.
- `TestCard_TempestCityRay_WardedOpposingUnitNotOffered` (`FR-KWD-002h`, `SVY-R36`): an opposing Ward unit is not among the options. When it is the only opposing unit, the ability is not offered, and a rejected activation moves nothing and pays nothing.
- `TestCard_TempestCityRay_NotOfferedWithoutAnOpposingUnit` (`SVY-R36`, `FR-ACT-001`): with only allied units, an opposing structure and both Headquarters on the battlefield, the ability is not offered, and a rejected activation moves nothing and pays nothing.
- `TestCard_TempestCityRay_NotOfferedWithoutTwoTactics` (`FR-ACT-001`): with R5 T1 it is not offered, and a rejected activation moves nothing.
- `TestCard_TempestCityRay_NotOfferedWhileExhausted` (`FR-ACT-002`): with R0 T4 an Exhausted Ray isn't offered, and a rejected activation moves nothing.
- `TestCard_TempestCityRay_MayActivateTheTurnItDeploys` (`FR-SET-004`): deployed from hand with R5 T2 and an opposing unit on the battlefield, its ability is offered right away.
- `TestCard_TempestCityRay_TargetGoneMovesNothing` (`FR-TGT-003`): the target leaves the battlefield between placement and resolution. Nothing moves, no `EventMove` is logged, and the Ray stays on the battlefield, Exhausted.
- `TestCard_TempestCityRay_TargetNowAlliedMovesNothing` (`FR-TGT-003`, `FR-ZON-002`): player 0 gains control of the target between placement and resolution (it stays in the Main Theater). It no longer matches "opposing", so it is illegal at the second check: it stays in the Main Theater under player 0's control, and no `EventMove` is logged.
- `TestCard_TempestCityRay_SourceGoneStillMoves` (`FR-TGT-003`, `FR-ZON-004`): the Ray is killed between placement and resolution. The chosen unit still moves to player 1's Reserves.
- **Between placement and resolution:** `s.ptPlaceHook` on the ability's own placement fires while it is being declared, before its costs are paid (`SVY-R2`; T-0304's tests round 1, note 2). Use T-0304's technique instead (`stormTapCard` in `cards_concordia_stormbolt_test.go`): a copy of the real card (copy the `Card` and its `Def`, so the shared definition isn't changed) whose `Triggers` also hold one synthetic `HookExhausted` trigger. It is placed on top of the ability and resolves first. Its step does the change (for example `{Op: ability.OpKill, Subject: ability.SubjectSelf}` for the Ray, an `OpReturnToHand` or `OpGainControl` step on `SubjectAll` with an opposing-unit filter for the target), or a hook at its placement does it.
- **Existing tests:** every other test passes unchanged, including `TestCards_Executable` with the name removed.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected.

## Allowed Files
- `pkg/engine/cards/concordia.go` (the definition)
- `pkg/engine/cards/cards.go` (remove the name from `NotYetExecutable`)

## Done When
- [ ] The card tests pass: `go test ./pkg/engine -run TestCard_TempestCityRay`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-matrix` exits 0 (just over two minutes: use a timeout above the two-minute default)
- [ ] The card loads with no unsupported lines (`TestCards_Executable` passes with the name removed)

## Log
- 2026-11-25 lead-architect (wave A run, session 8): written against aeb9f187 (the run branch after syncing `main`'s 94e9c254), with the lead's probe above. Board read twice (D2.0): #185 (4) is due here, for its *Tempest City Ray* half; #185 (3) stays for *Will Breaker* and the *Awaken Nightmare* half of (4) for wave F. No waiting trigger has fired. Every part of the Definition exists; `stepMoveZone` already uses the controller's Reserves, so #185 (4) needs no kernel change. Set to `tests`.
