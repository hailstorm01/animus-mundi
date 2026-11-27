---
id: T-0304
title: Make Stormbolt Ray executable (deploy damage and return itself, wave A, card, flash)
kind: card
tier: flash
wave: A
status: done
branch: m3/autoloop-wave-a/T-0304-stormbolt-ray
depends_on: [T-0303]
claims: [card:Stormbolt Ray]
cards: [Stormbolt Ray]
requirements: [FR-TRN-003, FR-TGT-001, FR-TGT-003, FR-ACT-001, FR-ACT-002, FR-ZON-004, FR-REP-004]
---

## Goal
*Stormbolt Ray* (a Concordia structure, R3 T0, 0/7, Main Theater; 2 copies in the starter deck): *"When this structure deploys, deal 3 damage to chosen asset."* and *"T2, exhaust: Return this structure to its owner's hand."* It only combines shapes that are `done`: a deploy trigger that damages one chosen target (*Chainbreaker Sappers*), and an activated ability paid with Tactics and `CostExhaustSelf` (*Cliff of Portents*) whose one step is `OpReturnToHand` on `SubjectSelf`. *Unceasing Tide* already returns itself that way from the discard pile; here the source is a battlefield asset, which `stepAssets` handles.

Covers:
- The definition in `pkg/engine/cards/concordia.go`, as under Definition below.
- The name off `cards.NotYetExecutable`.

## Inputs
- Card (`data/cards/units-and-structures-concordia.csv` lines 16-17).
- Survey row: `docs/design/survey/m3_full_catalog_ability_survey.md` §6, "Concordia: deck" (*Stormbolt Ray*).
- FRD 01: `FR-TRN-003` (line 394), `FR-TGT-001` (lines 407-413), `FR-TGT-003` (lines 435-442), `FR-ACT-001` and `FR-ACT-002` (lines 388-391), `FR-ZON-003` (line 250), `FR-ZON-004` (lines 278-283), `FR-SET-004` Immediate Readiness (line 203), `FR-REP-004` (lines 562-588), `FR-KWD-002h` (line 648), `FR-MOD-001` §1 (the single-option fast path) and §6 (timeouts).
- Rulings: `SVY-R3` (targets are chosen on placement), `SVY-R36` (an ability with no target slot needs only its costs), `SVY-R57` (only the stated criteria count). "chosen asset" has no allegiance, zone or exclusion, so any unit, structure or Headquarters is legal, the Ray itself and its controller's Headquarters included, and the trigger is mandatory.
- Design note: `docs/architecture/m3-full-catalog.md`, Boundaries (`pkg/engine/cards`), Deliverable 5.

### Definition (the engineer adds it to `concordiaDefs`)
- `"Stormbolt Ray"`, `Text`: `"When this structure deploys, deal 3 damage to chosen asset."`, then `"T2, exhaust: Return this structure to its owner's hand."`.
- One `Trigger`: `Hook: ability.HookDeploy`, not `Optional`. `Effect`: `Targets` is one exactly-one slot with `Filter: ability.Filter{}`; `Steps` is one `{Op: ability.OpDealDamage, Amount: ability.Value{N: 3}, Subject: ability.SubjectTarget, Target: 0}`.
- One `Activated`: `Costs` is `{Kind: ability.CostResources, Tac: 2}`, then `{Kind: ability.CostExhaustSelf}`. `Effect`: no targets; `Steps` is one `{Op: ability.OpReturnToHand, Subject: ability.SubjectSelf}`.

A lead probe with the definition on the run branch at a6dc8a21, since reverted, passed `go tool task check` and `go tool task sim-matrix`.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
In a new file `pkg/engine/cards_concordia_stormbolt_test.go`, from the CSV text, loading the real card (`s.card`), in the style of the *Chainbreaker Sappers* tests (`cards_path_test.go`) and `cards_concordia_regent_test.go`. Player 0 controls the Ray and is the active player. Reuse the existing test helpers (`deployTo`, `s.targetBoard`, `s.scriptRefs`, `wantOptions`, `s.wantDamage`, `totalDamage`, `damageTo`, `nonCombatDamage`, `activate`, `offersActivate`, `s.reject`, `s.setPool`, `inHand`, `inDiscardPile`, `s.putControlled`, `s.ptPlaceHook`). The other cards may be synthetic.

Deploying:
- `TestCard_StormboltRay_EntersMainTheaterAsZeroSeven` (`FR-TRN-003`): deployed from hand for R3 it is a Readied 0/7 structure in the Main Theater; deploying it to Reserves is rejected, and with R2 it can't be deployed.
- `TestCard_StormboltRay_AnyAssetOffered` (`FR-TGT-001`, `SVY-R57`): one mandatory single `DecideTarget` for player 0, whose options are exactly every battlefield asset: both players' units in the Main Theater and Reserves, a face-down Scratch Company of each, a structure of each, both Headquarters and the Ray itself.
- `TestCard_StormboltRay_DealsThreeToEachKind` (`FR-TGT-001`): a table over an allied unit, an opposing unit in Reserves, an allied and an opposing structure, both Headquarters and the Ray itself: the chosen asset has exactly 3 damage, and nothing else has any.
- `TestCard_StormboltRay_DamageIsNotCombat` (`FR-REP-004`): one damage event of 3, naming the Ray as `Event.Source`, not combat damage. A target with Stalwart takes the full 3. Against Resistant 1 and Armor 1, 1 damage is marked and the Armor is gone.
- `TestCard_StormboltRay_LethalDamageKills` (`FR-LAY-002`): an opposing unit with Health 3 dies and its card is in its owner's discard pile. With Health 4 it survives.
- `TestCard_StormboltRay_OpposingHeadquartersToZeroWins` (`FR-CBT-003`): an opposing Headquarters with 3 Current Health loses the match.
- `TestCard_StormboltRay_WardedOpposingAssetNotOffered` (`FR-KWD-002h`): an opposing Ward asset is not offered, and an allied one is.
- `TestCard_StormboltRay_TimeoutTakesTheFirstOption` (`FR-MOD-001`): a timeout still deals exactly 3 to exactly one legal target. FRD 01 `FR-MOD-001` §6's fixed order says which: assert that one.
- `TestCard_StormboltRay_TargetGoneDealsNothing` (`FR-TGT-003`): the target leaves between placement and resolution (`s.ptPlaceHook`). No damage event is logged, and the Ray stays.
- `TestCard_StormboltRay_SourceGoneStillDeals` (`FR-TGT-003`): the Ray leaves between placement and resolution, and another asset was chosen. The chosen asset still takes 3.

Activating (a Readied Ray put in player 0's Main Theater, R0 T2, unless the case says otherwise):
- `TestCard_StormboltRay_ReturnsToOwnersHand` (`FR-ACT-001`, `FR-ZON-004`): the Ray leaves the battlefield, its card is in player 0's hand and not in a discard pile, and the pool is R0 T0. Nothing is asked, and no other asset changes.
- `TestCard_StormboltRay_CostsPaidBeforeItResolves` (`FR-ACT-001`): when the ability is placed on the stack the Ray is still on the battlefield and Exhausted, and the 2 Tactics are gone.
- `TestCard_StormboltRay_GoesToItsOwnersHand` (`FR-ZON-003`): a Ray that player 1 owns and player 0 controls returns to player 1's hand, and not player 0's.
- `TestCard_StormboltRay_RedeployedIsNew` (`FR-ZON-004`, `FR-SET-004`): a Ray with damage and Armor, returned and deployed again the same turn, is a new Readied 0/7 with no damage and no Armor, and its deploy trigger deals 3 again.
- `TestCard_StormboltRay_MayActivateTheTurnItDeploys` (`FR-SET-004`): deployed from hand with R3 T2, its ability is offered right away.
- `TestCard_StormboltRay_NotOfferedWithoutTwoTactics` (`FR-ACT-001`): with R5 T1 it is not offered, and a rejected activation moves nothing.
- `TestCard_StormboltRay_NotOfferedWhileExhausted` (`FR-ACT-002`): with R0 T4 an Exhausted Ray isn't offered, and a rejected activation moves nothing.
- `TestCard_StormboltRay_GoneBeforeResolutionStaysGone` (`FR-ZON-004`, `FR-TGT-003`): the Ray is killed between placement and resolution. Its card stays in its owner's discard pile and is not in hand.
- **Existing tests:** every other test passes unchanged, including `TestCards_Executable` with the name removed.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected.

## Allowed Files
- `pkg/engine/cards/concordia.go` (the definition)
- `pkg/engine/cards/cards.go` (remove the name from `NotYetExecutable`)

## Done When
- [x] The card tests pass: `go test ./pkg/engine -run TestCard_StormboltRay`
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks
- [x] `go tool task sim-matrix` exits 0 (just over two minutes: use a timeout above the two-minute default)
- [x] The card loads with no unsupported lines (`TestCards_Executable` passes with the name removed)

## Log
- 2026-11-25 lead-architect (wave A run, session 6): written against a6dc8a21, with the lead's probe above. Board read twice (D2.0): no directive is due here (#185 (3) and (4) are *Will Breaker*'s and *Tempest City Ray*'s), and no waiting trigger has fired. Every part of the Definition exists; no other card returns itself from the battlefield. Set to `tests`.
- 2026-11-25 rules-test-author: wrote `pkg/engine/cards_concordia_stormbolt_test.go`, 21 tests, all in the task's list: Deploying (`EntersMainTheaterAsZeroSeven`, `AnyAssetOffered`, `DealsThreeToEachKind`, `DamageIsNotCombat`, `LethalDamageKills`, `OpposingHeadquartersToZeroWins`, `WardedOpposingAssetNotOffered`, `TimeoutTakesTheFirstOption`, `TargetGoneDealsNothing`, `SourceGoneStillDeals`) and Activating (`ReturnsToOwnersHand`, `CostsPaidBeforeItResolves`, `GoesToItsOwnersHand`, `RedeployedIsNew`, `MayActivateTheTurnItDeploys`, `NotOfferedWithoutTwoTactics`, `NotOfferedWhileExhausted`, `GoneBeforeResolutionStaysGone`). Helpers requested: none (file-local `stormChoose`, `stormDeploy`, `stormRay`, `stormKill`). `go vet ./pkg/engine` and `cardtown lint rules` are clean; 16 tests fail as expected until the card is executable (the two `NotOffered` tests pass trivially). The timeout test asserts the FR-MOD-001 sec. 6 fixed order: the choosing player's own Headquarters takes the 3. Nothing left out.
- 2026-11-25 lead-architect: tests round 1 returned (D4). Each expected outcome agrees with the FRD 01 text it cites (the file holds 18 tests, as listed, not 21). With the lead's probe of the Definition, 15 of 18 pass. Set to `tests` for two notes on setup; change nothing else:
  1. `TestCard_StormboltRay_DealsThreeToEachKind/the_Ray_itself` fails against a correct definition: `c` is assigned only when `stormDeploy` returns, so it is still nil when the target question is asked inside it. Make the Ray findable while the question is asked (for example, create the card first, or look the asset up by its card's name).
  2. `TestCard_StormboltRay_CostsPaidBeforeItResolves` and `TestCard_StormboltRay_GoneBeforeResolutionStaysGone` use `s.ptPlaceHook` on the ability's own placement. That event comes while the ability is being declared, before its costs are paid (targets and costs are one declaration, `SVY-R2`), so the first sees unpaid costs and the second removes the Ray before its exhaust is paid (a nil-pointer panic). Nothing is between declaration and resolution but the triggers the costs fire, so use one (`FR-TRN-003a`, `SVY-R17`): put a copy of the real card (copy the `Card` and its `Def`, so the shared definition isn't changed) whose `Triggers` also hold one synthetic `HookExhausted` trigger. It is placed on top of the ability and resolves first. In `CostsPaidBeforeItResolves` give it a harmless targetless step, and check at its placement (the second one) that the Ray is on the battlefield and Exhausted, the 2 Tactics are gone, and the return is still on the stack below; then the Ray ends in hand. In `GoneBeforeResolutionStaysGone` its step is `{Op: ability.OpKill, Subject: ability.SubjectSelf}`; the Ray ends in the discard pile and not in hand.
- 2026-11-25 rules-test-author: tests round 2, both notes fixed in `pkg/engine/cards_concordia_stormbolt_test.go` only. (1) `DealsThreeToEachKind/the_Ray_itself`: the card is now created before the chooser is set, and a new file-local `stormDeployCard` deploys that card, so `assetOf(s.g, c)` finds the Ray while the question is asked. (2) `CostsPaidBeforeItResolves` and `GoneBeforeResolutionStaysGone` no longer hook the ability's own placement. A new file-local `stormTapCard` copies the real `Card` and its `Def` (the shared definition is untouched) and adds one synthetic `HookExhausted` trigger. In `CostsPaid...` its step is `actHQDamage(1)` (targetless) and, at the second placement, the test checks the Ray is on the battlefield and Exhausted, Tactics are 0 and the return is the stack item below; then the Ray ends in hand. In `GoneBefore...` its step is `OpKill` on `SubjectSelf`; the Ray ends in the discard pile and not in hand. `go vet ./pkg/engine` and gofmt are clean. The card is not executable here, so the card tests still fail on the missing definition; not run against it. Helpers requested: none.
- 2026-11-25 lead-architect: tests approved (D4, round 2). Both notes are applied and nothing else changed. A lead probe with the Definition, since reverted, passes all 18, and all 28 mutants (amounts, optional, hook, an optional slot, allegiances, kinds, zone, other, non-Scratch, all, self, dealer, a doubled step, kill, each cost, the return's op and subject, an extra draw) fail at least one. `go vet` and `gofmt` are clean. Set to `ready`.
- 2026-11-25 kernel-engineer: implemented Stormbolt Ray in concordiaDefs (pkg/engine/cards/concordia.go) with a deploy trigger dealing 3 damage to one chosen asset and an activated ability (cost T2 + exhaust self) returning the structure to owner's hand; removed name from NotYetExecutable in pkg/engine/cards/cards.go. Reused shapes: HookDeploy, OpDealDamage, CostResources, CostExhaustSelf, OpReturnToHand with SubjectSelf. Verified: go test ./pkg/engine -run TestCard_StormboltRay (ok, 18/18 passed), go test ./pkg/cardcheck -run TestCards_Executable (ok), go tool task check (fmt, vet, tests, lint all passed), go tool task sim (200 games clean, all invariants hold), go tool task sim-matrix (24 starter + 24 stress runs clean, all invariants hold). Set to review.
- 2026-11-25 lead-architect (wave A run, session 6): approved (D5, round 1), merged as 726db3f4. The diff is the task's Definition and the one name off `NotYetExecutable`, inside `Allowed Files`. The test file changed only in 585701de and 663fe093, both before the implementation. No new shape, no card name in engine code. Re-run by the reviewer on the task branch, with exit codes: `go test ./pkg/engine -run TestCard_StormboltRay`, `go test ./pkg/cardcheck -run TestCards_Executable`, `go tool task check`, `go tool task sim`, `go tool task sim-matrix`, all 0. The merge has the same tree as the reviewed branch. The Animus vs Path sim is unchanged (P0 58.5% / P1 39.5% / draws 2.0%). The Flash run left no denied action.
