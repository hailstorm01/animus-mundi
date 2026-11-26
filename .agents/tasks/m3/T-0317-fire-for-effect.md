---
id: T-0317
title: Make Fire for Effect executable (opposing Main Theater assets gain Resistant -1 until end of turn; 1 to a chosen Main Theater asset; wave A, card, flash)
kind: card
tier: flash
wave: A
status: tests-review
branch: m3/autoloop-wave-a/T-0317-fire-for-effect
depends_on: [T-0316]
claims: [card:Fire for Effect]
cards: [Fire for Effect]
requirements: [FR-MNV-001, FR-TGT-001, FR-TGT-003, FR-MOD-001, FR-LAY-001, FR-LAY-003, FR-REP-004, FR-KWD-002b, FR-KWD-002h, FR-KWD-002k, FR-LAY-002, FR-CBT-002, FR-ZON-005]
---

## Goal
*Fire for Effect* (Soulless Maneuver, R0 T4, rare; one copy in `data/decks/soulless-starter.txt`): *"Opposing assets in the Main Theater gain Resistant -1 until end of turn. Deal 1 damage to chosen asset in the Main Theater."* It only combines shapes that are `done`: an untargeted `OpGrantKeyword` step on `SubjectAll` (the grant of *Allshan's Battle Trance* in `pkg/engine/cards/path.go`, which `FR-LAY-001`'s One-Shot Modifier Snapshot Locking names *Fire for Effect* for), with *Seismic Instability*'s Main Theater zone filter (`pkg/engine/cards/concordia.go`) plus `AllegianceOpposing`, then *Bring It Down*'s `OpDealDamage` step on one exactly-one slot, here filtered to the Main Theater only (any allegiance, Scratch Companies included). The grant comes first, as printed, so a chosen opposing asset takes 2 (`FR-REP-004` Stage 2: negative Resistant amplifies), and a chosen allied asset takes 1.

Covers:
- The definition in `soullessDefs` (`pkg/engine/cards/soulless.go`), as under Definition below.
- The name off `cards.NotYetExecutable`.

## Inputs
- Card (`data/cards/maneuvers-soulless.csv` line 16): Soulless Maneuver, Requisition cost 0, Tactics cost 4.
- Survey row: `docs/design/survey/m3_full_catalog_ability_survey.md` §6, "Soulless: deck" (*Fire for Effect*, line 174), and §3's "Pinned by one card in M2" note (`OpGrantKeyword`).
- FRD 01 (`docs/frd/01_Game_Rules_FRD.md`): `FR-MNV-001` and `SVY-R29` (line 395: a Maneuver needs a legal entity for each exactly-one slot), `FR-TGT-001` (line 407: "in the Main Theater" is a zone criterion), `FR-TGT-003` (line 435; line 441: a composite effect whose targets are all gone fizzles only its targeted clause, and the Maneuver still goes to the discard pile), the Single-Option Fast-Path and `FR-MOD-001` §6 timeouts (line 448 on), `FR-LAY-001` (line 528; line 532: One-Shot Modifier Snapshot Locking, which names this card), `FR-LAY-003` (line 540: durations), `FR-REP-004` (line 562; line 585: Stage 2, negative Resistant amplifies, even 0 raw damage), `FR-KWD-002b` Resistant (line 642), `FR-KWD-002h` Ward, `FR-KWD-002k` Armor, `FR-LAY-002` (line 534), `FR-ZON-005`.
- Rulings that apply: `SVY-R16` (timeouts take the first option in fixed order), `SVY-R25` (a removed target is no longer that target), `SVY-R29`, `SVY-R59` (a face-down Scratch Company is an asset; "chosen asset" has no non-Scratch exclusion here).
- Design note: `docs/architecture/m3-full-catalog.md`, Boundaries (`pkg/engine/cards`), Deliverable 5.

### Definition (the engineer adds it to `soullessDefs`, after *Earthshaker Colossus*)
- `"Fire for Effect"`, `Text`: two strings, as the CSV's two lines: `"Opposing assets in the Main Theater gain Resistant -1 until end of turn."`, `"Deal 1 damage to chosen asset in the Main Theater."`.
- `Effect`:
  - `Targets`: one slot, `Filter: ability.Filter{Zone: ability.ZoneMainTheater}`.
  - `Steps`, in this order:
    1. `{Op: ability.OpGrantKeyword, Subject: ability.SubjectAll, Filter: ability.Filter{Allegiance: ability.AllegianceOpposing, Zone: ability.ZoneMainTheater}, Keywords: ability.Keywords{Resistant: -1}}` (`Until` left at its zero value, until end of turn).
    2. `{Op: ability.OpDealDamage, Amount: ability.Value{N: 1}, Subject: ability.SubjectTarget, Target: 0}`.

A lead probe with the definition on the run branch at 8b323f83, since reverted, passed `go tool task check`, `go tool task sim` and `go tool task sim-matrix` (exit 0). A probe test, also reverted, showed the expected outcomes below: a chosen opposing Main Theater unit has Resistant -1 and 2 damage, Reserves and Headquarters are untouched, the grant is gone after `s.endTurn()`, a timeout damaged player 0's Main Theater unit (player 0's assets come first in fixed order), and with an empty Main Theater the Enact isn't offered.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
In a new file `pkg/engine/cards_soulless_fire_for_effect_test.go`, from the CSV text, loading the real card (`s.card("Fire for Effect")`), in the style of `cards_soulless_maneuvers_test.go` (*Bring It Down*: `bidEnact`, `offersEnact`, `enactAction`, `chooseAsset`, `targetAsk`) and `cards_concordia_seismic_test.go` (*Seismic Instability*: all assets in the Main Theater). Player 0 enacts with exactly R0 T4 and is the active player; player 1 is the opponent. Read keywords with `s.g.Keywords(a)`. Reuse the existing test helpers (`s.put`, `s.putScratch`, `s.give`, `s.setPool`, `s.script`, `s.asked`, `s.alive`, `s.attack` (with `s.do`), `s.endTurn`, `s.moveTo`, `s.wantDamage`, `wantOptions`, `stormKill`, `totalDamage`, `inDiscardPile`, `damageEvent`, `recorder`). The other cards may be synthetic. Every asset that a case doesn't mean to kill has enough health to survive the damage it takes.

- `TestCard_FireForEffect_ChosenOpposingAssetTakesTwo` (`FR-REP-004` Stage 2, `FR-KWD-002b`, `FR-LAY-001`): with two opposing Main Theater units, choosing one deals it 2 damage in one non-combat damage instance from *Fire for Effect* (the grant resolves first); the other takes none. Both have Resistant -1 afterwards.
- `TestCard_FireForEffect_ChosenAlliedAssetTakesOne` (`FR-TGT-001`, `FR-LAY-001`): with an allied and an opposing Main Theater unit, choosing the allied one deals it 1 damage. It has Resistant 0; the opposing unit has Resistant -1 and no damage.
- `TestCard_FireForEffect_OffersEveryMainTheaterAsset` (`FR-TGT-001`, `FR-KWD-002h`, `FR-ZON-005`, `SVY-R59`): with an allied and an opposing unit in each of the Reserves and the Main Theater, an allied and an opposing Main Theater structure, an allied and an opposing face-down Scratch Company in the Main Theater (`s.putScratch` and `s.moveTo`, or the helper that puts one there), an allied Warded Main Theater unit and an opposing Warded Main Theater unit, exactly one `DecideTarget` is asked of player 0, and it offers exactly the two Main Theater units, both Main Theater structures, both Scratch Companies and the allied Warded unit. Neither Headquarters, no Reserves asset and not the opposing Warded unit is offered.
- `TestCard_FireForEffect_GrantReachesOnlyOpposingMainTheaterAssets` (`FR-LAY-001`, `FR-ZON-005`): on the same board, after it resolves (any target), every opposing Main Theater asset (the unit, the structure, the Scratch Company and the Warded unit) has Resistant -1; the opposing Reserves unit, player 1's Headquarters and every allied asset have Resistant 0. The grant is not a target, so the opposing Warded unit gets it.
- `TestCard_FireForEffect_AddsToPrintedResistant` (`FR-KWD-002b`, `FR-REP-004` Stage 2): an opposing Main Theater unit with printed Resistant 1 has Resistant 0 afterwards, and choosing it deals it 1 damage. One with printed Resistant 2 has Resistant 1, and choosing it deals 0 (no damage is committed).
- `TestCard_FireForEffect_ArmorAbsorbs` (`FR-REP-004` Stage 4, `FR-KWD-002k`): a chosen opposing Main Theater unit with printed Armor 1 takes the amplified 2: it loses its Armor and takes 1 damage.
- `TestCard_FireForEffect_LethalDamageKills` (`FR-LAY-002`): a chosen opposing Main Theater unit with health 2 is destroyed, and its card is in player 1's discard pile.
- `TestCard_FireForEffect_LaterDamageThisTurnIsAmplified` (`FR-KWD-002b`, `FR-REP-004`, `FR-CBT-002`): after it resolves, an allied Main Theater unit with power 3 attacks an opposing Main Theater unit that was not chosen (health 20): the defender takes 4 combat damage. A second synthetic non-combat 1-damage step on it this turn deals 2.
- `TestCard_FireForEffect_GrantEndsAtEndOfTurn` (`FR-LAY-003`): after it resolves and `s.endTurn()`, every opposing asset that had Resistant -1 has its printed Resistant again (0, and 1 for one with printed Resistant 1).
- `TestCard_FireForEffect_SnapshotLocksTheGrant` (`FR-LAY-001` One-Shot Modifier Snapshot Locking): after it resolves, an opposing Reserves unit moved into the Main Theater (`s.moveTo`) has Resistant 0, and an opposing unit that had the grant and is moved to its Reserves keeps Resistant -1 until end of turn.
- `TestCard_FireForEffect_OneLegalTargetAsksNothing` (Single-Option Fast-Path, `FR-MOD-001`): with exactly one asset in the Main Theater, an opposing unit, player 0 is asked nothing and the unit takes 2.
- `TestCard_FireForEffect_TimeoutTakesTheFirstOption` (`FR-MOD-001` §6, `SVY-R16`): with an opposing and an allied Main Theater unit, a timed-out target decision (`s.script(0, Answer{Timeout: true})`) deals damage only to the decision's first option in fixed order, the allied unit, which takes 1; the total damage dealt is 1.
- `TestCard_FireForEffect_NotEnactableWithAnEmptyMainTheater` (`SVY-R29`, `FR-MNV-001`): with assets only in the Reserves and the Headquarters Zone, the Enact is not offered. With a single allied Main Theater unit, it is offered (an allied asset is a legal target).
- `TestCard_FireForEffect_TargetGoneStillGrants` (`FR-TGT-003` Complete Target Invalidation & Composite Clauses, `SVY-R25`): the chosen opposing unit is removed between placement and resolution (as in `TestCard_BringItDown_TargetGoneDealsNothing`, with `stormKill` on the `place` event). No asset is dealt damage, the other opposing Main Theater unit still has Resistant -1, and *Fire for Effect* is in player 0's discard pile.
- `TestCard_FireForEffect_PaysFourTactics` (`FR-MNV-001`): with R0 T3 the Enact is not offered; with R0 T4 it is, and the pool is empty after it resolves.
- **Existing tests:** every other test passes unchanged, including `TestCards_Executable` with the name removed.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected.

## Allowed Files
- `pkg/engine/cards/soulless.go` (the definition)
- `pkg/engine/cards/cards.go` (remove the name from `NotYetExecutable`)

## Done When
- [ ] The card tests pass: `go test ./pkg/engine -run TestCard_FireForEffect`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-matrix` exits 0 (just over two minutes: use a timeout above the two-minute default)
- [ ] The card loads with no unsupported lines (`TestCards_Executable` passes with the name removed)

## Log
- 2026-11-26 lead-architect (wave A run, session 19): written against 8b323f83, with the lead's probe above. `main` has no commits the run branch lacks. Board read (D2.0): no directive is due here (#185 has only (4)'s *Awaken Nightmare* half left, wave F), and no waiting trigger has fired: the card has no each-player effect (#71), no cost on a mode (#91 (2)), no deck search (#106), no discard cost (#120), no tuck (#142), no pile round trip (#183 (2)) and no turn-boundary or draw hook (#15 (2): "until end of turn" expires at the existing end-of-turn expiry, which T-0251 already sweeps after), the mode is still `serial` (#6, #7), and no `needsRework` row exists (#184). Taken next in the survey's Soulless deck order, after *Earthshaker Colossus*. Searched `pkg/engine` and the task archives: every part of the Definition exists (*Allshan's Battle Trance*' `OpGrantKeyword`, `SubjectAll` with a Main Theater zone filter as in *Seismic Instability*, and *Bring It Down*'s `OpDealDamage` step), so no kernel change. Set to `tests`.
- 2026-11-26 rules-test-author: wrote `pkg/engine/cards_soulless_fire_for_effect_test.go`, 14 tests: TestCard_FireForEffect_{ChosenOpposingAssetTakesTwo, ChosenAlliedAssetTakesOne, OffersEveryMainTheaterAsset, GrantReachesOnlyOpposingMainTheaterAssets, AddsToPrintedResistant, ArmorAbsorbs, LethalDamageKills, LaterDamageThisTurnIsAmplified, GrantEndsAtEndOfTurn, SnapshotLocksTheGrant, OneLegalTargetAsksNothing, TimeoutTakesTheFirstOption, NotEnactableWithAnEmptyMainTheater, TargetGoneStillGrants, PaysFourTactics}. Helpers requested: none (local helpers ffe* live in the test file). Cases left out: none. Notes: the target-offer set is compared order-insensitively (fixed order is pinned only by the timeout test); the second non-combat 1 damage is an activated "Pinger" unit hitting opposing Main Theater assets. `go vet` and `cardtown lint rules` pass; the tests fail until the card is defined.
