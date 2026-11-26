---
id: T-0307
title: Make Hurricane Outrigger executable (a deploy trigger gives its controller's Headquarters +1 power until end of turn, wave A, card, flash)
kind: card
tier: flash
wave: A
status: tests-review
branch: m3/autoloop-wave-a/T-0307-hurricane-outrigger
depends_on: [T-0306]
claims: [card:Hurricane Outrigger]
cards: [Hurricane Outrigger]
requirements: [FR-TRN-003, FR-LAY-001, FR-LAY-003, FR-TRN-001, FR-CBT-002, FR-ZON-004, FR-ZON-005]
---

## Goal
*Hurricane Outrigger* (a Concordia unit, Cavalry, Tempest, R2 T0, 2/1, Reserves, yields R1 T1; in the starter deck): *"When this unit deploys, your Headquarters gets +1 power until the end of your turn."* It only combines shapes that are `done`: a `HookDeploy` trigger with no target (*Stormbolt Ray*, T-0304, has one with a target), and an `OpModifyStat` step on `SubjectAll` with a filter (Layer 5 power, as *Rapturous Sermon* does in `pkg/engine/cards/animus.go`), here with `Kinds: ability.KindHeadquarters` (a filter kind *Arkha, Sun's Fury* already uses on a target, `pkg/engine/cards/path.go`). The modifier's default duration, `ability.UntilEndOfTurn`, expires in Sub-Phase 4c. The survey models "until the end of your turn" that way (§4, "Noted but unreachable": no card deploys a Concordia unit on an opponent's turn).

Covers:
- The definition in `pkg/engine/cards/concordia.go`, as under Definition below.
- The name off `cards.NotYetExecutable`.

## Inputs
- Card (`data/cards/units-and-structures-concordia.csv` line 2).
- Survey row: `docs/design/survey/m3_full_catalog_ability_survey.md` §6, "Concordia: deck" (*Hurricane Outrigger*, line 132), and §4's "Noted but unreachable" (line 104).
- FRD 01: `FR-TRN-003` (line 393), `FR-LAY-001` (lines 528-532), `FR-LAY-003` (lines 540-546), `FR-TRN-001` Sub-Phase 4c (line 338), `FR-CBT-002` (lines 613-632: a Headquarters attacks and retaliates with its effective Combat Power), `FR-ZON-004` (lines 278-283), `FR-ZON-005` Double-Blind Deployment (line 288: a face-down Scratch Company's only printed ability is Deploys Exhausted), `FR-SET-004` Immediate Readiness (line 203).
- Design note: `docs/architecture/m3-full-catalog.md`, Boundaries (`pkg/engine/cards`), Deliverable 5.

### Definition (the engineer adds it to `concordiaDefs`)
- `"Hurricane Outrigger"`, `Text`: `"When this unit deploys, your Headquarters gets +1 power until the end of your turn."`.
- One `Trigger`: `Hook: ability.HookDeploy`. `Effect`: no `Targets`; `Steps` is one `{Op: ability.OpModifyStat, Amount: ability.Value{N: 1}, Subject: ability.SubjectAll, Filter: ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindHeadquarters}}` (no `Stat`, no `Until`: power, until end of turn).

A lead probe with the definition on the run branch at 4800bb61, since reverted, passed `go tool task check`, `go tool task sim` (P0 58.5% / P1 39.5% / draws 2.0%, unchanged) and `go tool task sim-matrix`. Deployed from hand, it took player 0's Headquarters from 3 to 4 power and left player 1's at 3; after `s.endTurn()` both were 3.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
In a new file `pkg/engine/cards_concordia_outrigger_test.go`, from the CSV text, loading the real card (`s.card`), in the style of `cards_concordia_stormbolt_test.go`. Player 0 deploys the Outrigger and is the active player, unless the case says otherwise. Both Headquarters are `newScene`'s 3/20. Read effective power with `s.g.Power`, and find a player's Headquarters with `vrHQ`. Reuse the existing test helpers (`s.deployTo`, `s.setPool`, `s.put`, `s.give`, `s.onTop`, `s.attack`, `s.endTurn`, `s.alive`, `s.wantDamage`, `s.ptPlaceHook`, `effectEvents`, `inDiscardPile`). The other cards may be synthetic.

- `TestCard_HurricaneOutrigger_EntersReservesAsTwoOne` (`FR-TRN-003`, `FR-SET-004`): deployed from hand for R2 it is a Readied 2/1 unit in player 0's Reserves; deploying it to the Main Theater is rejected, and with R1 it can't be deployed.
- `TestCard_HurricaneOutrigger_HeadquartersGetsOnePower` (`FR-LAY-001`): after the deploy and its trigger resolve, player 0's Headquarters has power 4, player 1's has 3, and the Outrigger has 2. One `EventPower` of +1 names player 0's Headquarters, and no other power or max-health event is logged. Nothing is asked. Player 0's Headquarters' Health and damage are unchanged, and no other asset changes (a synthetic allied unit in the Main Theater keeps its power).
- `TestCard_HurricaneOutrigger_ExpiresAtEndOfTurn` (`FR-LAY-003`, `FR-TRN-001`): after `s.endTurn()`, player 0's Headquarters is back to power 3, and it is still 3 after player 1's turn ends too (no "until the end of your opponent's turn").
- `TestCard_HurricaneOutrigger_HeadquartersAttacksForFour` (`FR-CBT-002`): after the deploy resolves, player 0's Readied Headquarters attacks player 1's Headquarters. Player 1's Headquarters takes 4 combat damage, and player 0's takes 3 in retaliation.
- `TestCard_HurricaneOutrigger_TwoDeploysStack` (`FR-LAY-001`): two Outriggers deployed the same turn (pool R4) give player 0's Headquarters power 5.
- `TestCard_HurricaneOutrigger_OpponentsOutriggerBoostsTheirHeadquarters` (`FR-LAY-001`): after `s.endTurn()`, player 1 deploys an Outrigger on their own turn. Player 1's Headquarters has power 4, and player 0's has 3.
- `TestCard_HurricaneOutrigger_OtherDeploysDoNothing` (`FR-TRN-003`): a synthetic vanilla allied unit deployed from hand, with an Outrigger already on player 0's battlefield (`s.put`), leaves both Headquarters at power 3.
- `TestCard_HurricaneOutrigger_SourceGoneStillBoosts` (`FR-ZON-004`): the Outrigger is killed between its trigger's placement and resolution (`s.ptPlaceHook` on the trigger's placement, as in `TestCard_StormboltRay_SourceGoneStillDeals`). Player 0's Headquarters still gets power 4, and the Outrigger's card is in player 0's discard pile.
- `TestCard_HurricaneOutrigger_FaceDownScratchDoesNothing` (`FR-ZON-005`): with an Outrigger on top of player 0's deck (`s.onTop`), player 0 enacts the real *Growing Legion* (`s.card("Growing Legion")`, "Scratch 1."). A face-down Scratch Company enters player 0's Reserves, and both Headquarters stay at power 3.
- **Existing tests:** every other test passes unchanged, including `TestCards_Executable` with the name removed.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected.

## Allowed Files
- `pkg/engine/cards/concordia.go` (the definition)
- `pkg/engine/cards/cards.go` (remove the name from `NotYetExecutable`)

## Done When
- [ ] The card tests pass: `go test ./pkg/engine -run TestCard_HurricaneOutrigger`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-matrix` exits 0 (just over two minutes: use a timeout above the two-minute default)
- [ ] The card loads with no unsupported lines (`TestCards_Executable` passes with the name removed)

## Log
- 2026-11-25 lead-architect (wave A run, session 9): written against 4800bb61, with the lead's probe above. `main` has no commits the run branch lacks. Board read twice (D2.0): no directive is due here (#185 stays for (3) *Will Breaker* and (4)'s *Awaken Nightmare* half, wave F), and no waiting trigger has fired. Every part of the Definition exists, so no kernel change. Set to `tests`.
- 2026-11-25 rules-test-author: wrote the nine tests in pkg/engine/cards_concordia_outrigger_test.go: EntersReservesAsTwoOne, HeadquartersGetsOnePower, ExpiresAtEndOfTurn, HeadquartersAttacksForFour, TwoDeploysStack, OpponentsOutriggerBoostsTheirHeadquarters, OtherDeploysDoNothing, SourceGoneStillBoosts, FaceDownScratchDoesNothing. Helpers requested: none (reuses deployTo, vrHQ, effectEvents, ptPlaceHook, glEnact, enactAction, faceDownCount). Cases left out: none. `go vet` is clean and `cardtown lint rules` passes. Seven tests fail because the card is not yet executable; OtherDeploysDoNothing and FaceDownScratchDoesNothing pass already (they pin that nothing happens), and must still pass after the engineer's change.
