---
id: T-0303
title: Make Soul Whip executable (a chosen allied unit deals its power, wave A, card, flash)
kind: card
tier: flash
wave: A
status: done
branch: m3/autoloop-wave-a/T-0303-soul-whip
depends_on: [T-0302]
claims: [card:Soul Whip]
cards: [Soul Whip]
requirements: [FR-MNV-001, FR-TGT-001, FR-TGT-003, FR-REP-004]
---

## Goal
*Soul Whip* (a Concordia Maneuver, R0 T2; 2 copies in the starter deck): *"Chosen allied unit deals damage equal to its power to chosen opposing unit."* It only combines shapes that are `done`: one `OpDealDamage` step with a chosen `Dealer` (*Ambush Tactics*' third mode), on two slots with different allegiances. It is the second card to use `Dealer` (survey §3, "Pinned by one card in M2").

Covers:
- The definition in `pkg/engine/cards/concordia.go`, as under Definition below.
- The name off `cards.NotYetExecutable`.

## Inputs
- Card (`data/cards/maneuvers-concordia.csv` line 8).
- Survey row: `docs/design/survey/m3_full_catalog_ability_survey.md` §6, "Concordia: deck" (*Soul Whip*).
- FRD 01: `FR-MNV-001` (lines 393-396), `FR-TGT-001` (lines 407-413), `FR-TGT-003` (lines 435-441, "A Chosen Asset That Deals Damage"), `FR-ZON-004` (line 283), `FR-LAY-001`, `FR-LAY-002` (line 536), `FR-REP-004` (lines 562-588, Stages 2, 3 and 5), `FR-KWD-002a` (line 639), `FR-KWD-002h` (line 648).
- Rulings: `SVY-R27` (a chosen dealer that is gone deals nothing; otherwise its current power), `SVY-R29` (both slots need a legal entity), `SVY-R3` (targets are chosen on declaration).
- Design note: `docs/architecture/m3-full-catalog.md`, Boundaries (`pkg/engine/cards`), Deliverable 5.

### Definition (the engineer adds it to `concordiaDefs`)
- `"Soul Whip"`, `Text`: `"Chosen allied unit deals damage equal to its power to chosen opposing unit."`.
- `Effect`: `Targets` is `{Filter: ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit}}`, then `{Filter: ability.Filter{Allegiance: ability.AllegianceOpposing, Kinds: ability.KindUnit}}`. `Steps` is `{Op: ability.OpDealDamage, Amount: ability.Value{From: ability.ValueSourcePower}, Subject: ability.SubjectTarget, Target: 1, Dealer: 1}`.

A lead probe with the definition on the run branch at 0972bf78, since reverted, passed `go tool task check`.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
In a new file `pkg/engine/cards_concordia_soulwhip_test.go`, from the CSV text, loading the real card (`s.card`), in the style of the *Ambush Tactics* tests (`cards_path_maneuvers_test.go`) and `dealer_test.go`. Player 0 enacts; player 1 is the opponent. Reuse the existing test helpers (`enactAction`, `enactFor`, `offersEnact`, `refs`, `sortedCopy`, `pick`, `s.scriptModal`, `wantDamageFrom`, `nonCombatDamage`, `damageTo`, `inDiscardPile`, `lastDiscard`, `s.putScratch`, `s.setPool`, `s.onTop`). The other cards may be synthetic. Give every dealer a power above 0 that differs from the other unit's, unless the case says otherwise.
- `TestCard_SoulWhip_ChosenAllyDealsItsPower` (`FR-TGT-003`, `SVY-R27`): the opposing unit takes the dealer's power in exactly one damage event, which names the dealer as `Event.Source` and is not combat damage. The dealer takes nothing (no retaliation), and is not Exhausted by dealing. Other units are untouched. Soul Whip ends in its owner's discard pile.
- `TestCard_SoulWhip_TwoTargetQuestions` (`FR-TGT-001`): with units of both players in the Main Theater and in Reserves, a face-down Scratch Company of each, a structure of each and both Headquarters, two `DecideTarget` questions are asked, in this order: the first offers exactly player 0's units, the second exactly player 1's. Scratch Companies are units; structures and Headquarters are never offered. Each takes exactly one pick.
- `TestCard_SoulWhip_AnyZoneAndExhausted` (`FR-TGT-001`): an Exhausted dealer in Reserves deals to an opposing unit in Reserves.
- `TestCard_SoulWhip_DealerUsesItsCurrentPower` (`SVY-R27`, `FR-LAY-001`): a dealer with +1 power until end of turn (*Chainbreaker's Tenacity*, as `TestCard_AmbushTactics_DealerUsesItsCurrentPower`) deals the modified power.
- `TestCard_SoulWhip_PowerReadOnResolution` (`SVY-R27`, `SVY-R3`): the dealer's power changes after the Maneuver is placed and before it resolves (the observer pattern of `TestFR_TGT_003_DealerGoneBeforeResolutionDealsNothing`). It deals the power it has on resolution.
- `TestCard_SoulWhip_DealerGoneDealsNothing` (`FR-TGT-003`, `SVY-R27`, `FR-ZON-004`): the dealer leaves the battlefield between placement and resolution. No damage event is logged, the opposing unit is undamaged, and Soul Whip still goes to the discard pile.
- `TestCard_SoulWhip_TargetGoneDealsNothing` (`FR-TGT-003`): the opposing unit leaves between placement and resolution. No damage event is logged, and the dealer stays.
- `TestCard_SoulWhip_StalwartDoesNotCapIt` (`FR-REP-004` Stage 3): an opposing unit with Stalwart takes the full power.
- `TestCard_SoulWhip_ResistantAndArmorApply` (`FR-REP-004`, `FR-KWD-002b`, `FR-KWD-002k`): against Resistant 1 and Armor 1, a power 4 dealer marks 2 damage, and the Armor is gone.
- `TestCard_SoulWhip_LethalDamageKills` (`FR-LAY-002`): an opposing unit with Health equal to the dealer's power dies, and its card goes to its owner's discard pile.
- `TestCard_SoulWhip_UnbowedTriggers` (`FR-KWD-002a`): a synthetic surviving opposing unit with an Unbowed trigger triggers once.
- `TestCard_SoulWhip_WardedOpposingUnitNotOffered` (`FR-KWD-002h`): an opposing Ward unit is not offered beside one without Ward, and an allied Ward unit is offered as dealer. With the Ward unit the only opposing unit, Soul Whip is not offered.
- `TestCard_SoulWhip_NeedsBothUnits` (`SVY-R29`): not offered with no allied unit (an opposing unit present), and not with no opposing unit (allied units present). Offered with one of each.
- `TestCard_SoulWhip_PaysTwoTactics` (`FR-MNV-001`): from R3 T5 the pool is R3 T3 afterwards. With R5 T1 it is not offered.
- **Existing tests:** every other test passes unchanged, including `TestCards_Executable` with the name removed.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected.

## Allowed Files
- `pkg/engine/cards/concordia.go` (the definition)
- `pkg/engine/cards/cards.go` (remove the name from `NotYetExecutable`)

## Done When
- [x] The card tests pass: `go test ./pkg/engine -run TestCard_SoulWhip`
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks
- [x] `go tool task sim-matrix` exits 0 (just over two minutes: use a timeout above the two-minute default)
- [x] The card loads with no unsupported lines (`TestCards_Executable` passes with the name removed)

## Log
- 2026-11-25 lead-architect (wave A run, session 5): written against 0972bf78, with the lead's probe above. No directive is due here (#185 (3) and (4) are *Will Breaker*'s and *Tempest City Ray*'s), and no waiting trigger has fired. Every part of the Definition exists. Set to `tests`.
- 2026-11-25 rules-test-author: tests written in pkg/engine/cards_concordia_soulwhip_test.go (all fail until the card is executable): TestCard_SoulWhip_ChosenAllyDealsItsPower, _TwoTargetQuestions, _AnyZoneAndExhausted, _DealerUsesItsCurrentPower, _PowerReadOnResolution, _DealerGoneDealsNothing, _TargetGoneDealsNothing, _StalwartDoesNotCapIt, _ResistantAndArmorApply, _LethalDamageKills, _UnbowedTriggers, _WardedOpposingUnitNotOffered, _NeedsBothUnits, _PaysTwoTactics. Helpers requested: none. Cases left out: none. go vet is clean, gofmt clean, cardtown lint rules passes. Set to tests-review.
- 2026-11-25 lead-architect: tests round 1 returned (D4). Each test agrees with the FRD 01 text it cites. With the lead's probe, 13 of 14 pass. Set to `tests` for one note:
  1. `TestCard_SoulWhip_WardedOpposingUnitNotOffered` fails against a correct definition ("two target questions are asked, got 0"): each slot has one legal unit, so neither choice is asked (`FR-MOD-001` §6, the single-option fast path). Put a second allied unit without Ward and a second opposing unit without Ward in its first scene, so both questions are asked, and check that the second question offers exactly the two opposing units without Ward. Change nothing else.
- 2026-11-25 rules-test-author: round 1 note 1 fixed in TestCard_SoulWhip_WardedOpposingUnitNotOffered (second allied and second opposing unit without Ward; the second question must offer exactly the two non-Ward opposing units). Nothing else changed. Set to tests-review.
- 2026-11-25 lead-architect: tests approved (D4, round 2). Note 1 is applied and nothing else changed (the note's "§6" should read `FR-MOD-001` §1, the Single-Option Fast-Path, FRD 01 line 449). A lead probe with the Definition, since reverted, passes all 14, and all 21 mutants (dealer, target, fixed amounts, allegiances, kinds, zones, an optional slot, kill, all, a doubled step) fail at least one. `go vet`, `gofmt` and `cardtown lint rules` are clean. Set to `ready`.
- 2026-11-25 kernel-engineer: implemented Soul Whip in pkg/engine/cards/concordia.go as specified in Definition (two Target slots: allied unit, then opposing unit; one OpDealDamage step with Target: 1, Dealer: 1, Amount from ValueSourcePower, reusing the OpDealDamage Dealer shape from Ambush Tactics). Removed Soul Whip from NotYetExecutable in pkg/engine/cards/cards.go. Commands run: go test ./pkg/engine -run TestCard_SoulWhip (all 14 tests pass), go test ./pkg/cardcheck -run TestCards_Executable (passes), go tool task check (gofmt, vet, test, cardtown lint --all pass), go tool task sim (passes starter decks, 200 games, all invariants hold), go tool task sim-matrix (passes 24 starter and 24 stress runs, all invariants hold). Set to review.
- 2026-11-25 lead-architect (wave A run, session 5): approved (D5, round 1), merged as d9272bc3. The diff is the task's Definition and the one name off `NotYetExecutable`, inside `Allowed Files`. The test file changed only in 656394a5 and 311099f4, both before the implementation. No new shape, no card name in engine code. Re-run by the reviewer on the task branch, with exit codes: `go test ./pkg/engine -run TestCard_SoulWhip`, `go test ./pkg/cardcheck -run TestCards_Executable`, `go tool task check`, `go tool task sim`, `go tool task sim-matrix`, all 0. The merge has the same tree as the reviewed branch. The Animus vs Path sim is unchanged (P0 58.5% / P1 39.5% / draws 2.0%). The Flash run left no denied action.
