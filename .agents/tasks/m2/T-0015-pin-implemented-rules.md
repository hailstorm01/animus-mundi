---
id: T-0015
title: Pin four implemented but untested requirements with scenario tests
kind: shape
tier: sonnet
wave: A1 (board directive #16)
status: tests-review
branch: m2/T-0015-pin-implemented-rules
depends_on: []
claims: [tests:FR-SET-003, tests:FR-ZON-001, tests:FR-ZON-003, tests:FR-CRD-101]
cards: []
requirements: [FR-SET-003, FR-ZON-001, FR-ZON-003, FR-CRD-101]
---

## Goal
The kernel already implements these four requirements, but no `TestFR_*` test traces to them. This task adds those tests. Because the behavior exists, the tests are expected to **pass when written**, and "fails first" does not apply. A test that fails is a real bug, found by testing what FRD 01 says. If one fails, the lead adds the fix's `Allowed Files` here and dispatches `kernel-engineer` (tier sonnet). If none fails, no engineer step happens: the lead reviews the tests and merges them.

## Inputs
- FRD 01 (`docs/frd/01_Game_Rules_FRD.md`):
  - `FR-SET-003` First Player Determination (lines 70–72): a random decider, chosen after the Headquarters are placed and before opening hands are drawn. The decider chooses Play First or Draw First, and a timeout means Play First.
  - `FR-CRD-101` Printed Card Characteristics (lines 108–126): the characteristics table.
  - `FR-ZON-003` Ownership, Control & Readiness (lines 128–132): Assets vs Cards, owner vs controller, and Readied vs Exhausted.
  - `FR-ZON-001` Zone Definitions & Uncapped Capacity (lines 138–147), especially the Universal Uncapped Capacity Invariant (no hand-size limit and no end-of-turn discard, also stated in sub-phase 4d, line 219).
- Card data for `FR-CRD-101`: `data/cards/*.csv`. Compare real cards loaded with `s.card` against their CSV rows.
- Match setup is reached through the public `Config` and `NewGame` (see `go doc ./pkg/engine Config`). The PlayFirst callback models the decider's answer, and a nil callback models the timeout.
- Rulings that apply: none.

## Tests (written by rules-test-author)
Put them in a new file, `pkg/engine/foundation_test.go`. Name them `TestFR_SET_003_*`, `TestFR_CRD_101_*`, `TestFR_ZON_003_*` and `TestFR_ZON_001_*`. Each tests one statement from the FRD text through state or the public API, never internal fields that the FRD does not name. Points to cover, if they can be observed without guessing:
- **FR-SET-003:** across seeds, each player is sometimes the decider. The decider's Play First makes it the Turn 1 player, and Draw First makes its opponent the Turn 1 player. A timeout (nil callback) means Play First. The decision comes before opening hands are drawn. Test this only if the order is observable through the public API; otherwise leave it out and say why.
- **FR-CRD-101:** a few real cards across types (a Unit, a Structure, a Maneuver and a Headquarters) carry their CSV name, type, faction, subtypes, both costs, power, health, both yields, deployment zones and rarity. Pick cards whose values differ, so a swapped column would fail.
- **FR-ZON-003:** ownership never changes, while control is who commands the asset. A Readied asset can pay an exhaust cost and an Exhausted one cannot. Don't repeat cases already pinned under other IDs (for example, the owner's-discard rule in FR-ZON-002, or T-0014's owner-versus-controller card tests). Cite those tests instead.
- **FR-ZON-001:** a hand far larger than 6 survives the end of turn with no discard, and Reserves and the Main Theater accept many assets.
If a test fails, don't weaken it. Leave it failing, say so in the Log, and set `tests-review`.

Written in `pkg/engine/foundation_test.go` (17 tests):
- FR-SET-003: `TestFR_SET_003_DeciderIsRandomAcrossSeeds`, `_DeciderPlayFirstTakesTurn1`, `_DeciderDrawFirstGivesOpponentTurn1`, `_TimeoutDefaultsToPlayFirst`.
- FR-CRD-101 (compared with the CSV row read at test time): `TestFR_CRD_101_UnitCarriesCSVCharacteristics` (Shifting Impossibility, Devilhorn Guerrillas), `_StructureCarriesCSVCharacteristics` (Cathedral of Self), `_ManeuverCarriesCSVCharacteristics` (Disorienting Pulse), `_HeadquartersCarriesCSVCharacteristics` (Command Barge).
- FR-ZON-003: `TestFR_ZON_003_BattlefieldZonesHoldAssetsOthersHoldCards`, `_DeployerControlsAndDeckOwnerOwns`, `_OwnerUnchangedByMovingBetweenZones`, `_ReadiedAssetCanPayExhaustCostOnce`, `_ExhaustedAssetCannotPayExhaustCost`.
- FR-ZON-001: `TestFR_ZON_001_HandOverSixSurvivesEndOfTurnWithoutDiscard`, `_ReservesAndMainTheaterAcceptManyAssets`, `_ReservesGenerateRequisitionTheaterGeneratesTactics`, `_ReservesAssetsAreProtectedFromAttacks`.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected. Local helpers go in `foundation_test.go`.

## Allowed Files
- None unless a test fails. The lead then fills this in before any engineer step.

## Done When
- [ ] `go test ./pkg/engine -run 'TestFR_(SET_003|CRD_101|ZON_003|ZON_001)_'` passes
- [ ] `go run ./cmd/cardtown lint rules` passes and traces all four IDs
- [ ] `go test ./...`, `go vet ./...` and `gofmt -l pkg cmd` are clean
- [ ] `go run ./cmd/cardtown lint --all` passes

## Log
- 2026-10-20 lead-architect: Wrote the task from board directive #16 (audit finding 3), with status `tests` for `rules-test-author`. The kernel already has the behavior, so the tests are expected to pass when written. No engineer step happens unless one fails.
- 2026-10-20 rules-test-author: Wrote 17 tests in `pkg/engine/foundation_test.go` (see Tests). 16 pass. 1 FAILS and is left failing: `TestFR_CRD_101_HeadquartersCarriesCSVCharacteristics`. Command Barge loads with rarity "HQ" (matching its CSV row), but FR-CRD-101 lists the legal rarity values as Common, Uncommon, Rare, Mythic or Headquarters. This is a mismatch between FRD 01 and the card data/loader vocabulary (`data/cards/units-and-structures-HQ.csv` says "HQ"), not clearly a kernel bug. Lead to decide: change the data/loader to Headquarters, or amend the FRD wording. All other HQ columns (name, type, faction, no subtypes, costs 0/0, power, health, yields, zone) match. Left out: FR-SET-003 "decider chosen before opening hands are drawn" (the callback receives only the decider, so the order is not observable through Config or the public API); FR-ZON-003 control differing from ownership (needs a control-changing effect or internal fields; T-0014 and the FR-ZON-002 tests cover the related behavior); the Attached zone (no public representation); uncapped Discard (nothing observable to drive it without a card effect). The timeout test cannot see the decider, so it asserts that both seats take Turn 1 across seeds. Helpers requested: none. `cardtown lint rules` passes and all four IDs trace.
