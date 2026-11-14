---
id: T-0238
title: A "you may" that could do nothing isn't asked; Devilhorn Captain and Necropolis Hunt say "you may" (#152, wave F.2 shape, sonnet)
kind: shape
tier: sonnet
wave: F.2
status: review
branch: m2/autoloop-wave-f.2/T-0238-optional-could-do-nothing
depends_on: []
claims: [shape:couldDoNothing (an optional trigger with nothing to act on is not asked or placed, SVY-R48), card-text:Devilhorn Captain, card-text:Necropolis Hunt]
cards: [Devilhorn Captain, Necropolis Hunt]
requirements: [FR-MOD-001, FR-TGT-002]
---

## Goal
Directive #152 (`SVY-R47`, `SVY-R48`). An optional ("you may") trigger is not asked, and not placed, when accepting could do nothing: it has target slots, every step acts only on them, and no slot has a legal entity at placement. *Devilhorn Captain* and *Necropolis Hunt* become optional triggers whose text says "you may", so a lone damaged Captain can decline instead of hitting itself until it dies, and a timeout declines.

Covers:
- (1) The skip, in the trigger placement path, judged by the same options `chooseTargets` would offer (no second legality copy).
- (2) The two cards: their `rulesText` in `data/cards/units-and-structures-path.csv`, their `Text` in `pkg/engine/cards/path.go` (`TestDefs_MatchRulesText` pins the pair), and `Optional: true` on both triggers, in one commit.
- The merged tests the change contradicts, edited as listed under Tests, and the new tests.

Not this task: *Stardust Colossus* (Concordia, outside M2, no definition; `pm-frd-drafter` already reworded its CSV row).

## Inputs
- FRD 01 v3.25: `FR-MOD-001` §1 (line 416, Single-Option Fast-Path), §7 (line 432, "Not Asked When It Could Do Nothing", `SVY-R48`), §5 (line 420: a mandatory trigger is still placed with nothing legal, `SVY-R30`); `FR-TGT-002` (line 397, "You May" Makes a Trigger Optional, `SVY-R47`); revision summary v3.24 → v3.25 (lines 18–19).
- Survey: `docs/design/m2_starter_set_ability_survey.md` rows R47 and R48 (lines 106–107); card rows *Necropolis Hunt* (line 176) and *Devilhorn Captain* (line 178).
- Rulings that apply: `SVY-R47`, `SVY-R48`, `SVY-R9` (optional triggers), `SVY-R14` (an untargeted pick is not a target), `SVY-R30`, `SVY-R16`.
- Design Outcome: `docs/design/outcomes/2026-11-14-optional-damage-triggers.md` (card texts under "Card text").
- Design note: `docs/architecture/m2-ability-model.md`, Deliverable 6k (wave F.2).
- Code: `placeTrigger` and `accepts` ([trigger.go:230-253](../../../pkg/engine/trigger.go)); `targetDecision` ([targeting.go:78-104](../../../pkg/engine/targeting.go)), whose `Options` `chooseTargets` offers and `slotsLegal` reads; `placeable` and `legalModes` ([placement.go:167-195](../../../pkg/engine/placement.go)); `ability.Subject` ([ability.go:340-350](../../../pkg/engine/ability/ability.go)); the two definitions ([cards/path.go:349-373](../../../pkg/engine/cards/path.go), [:578-602](../../../pkg/engine/cards/path.go)); the CSV rows (`units-and-structures-path.csv` lines 10 and 19).

### Shape (the engineer builds it)
- **`(*Game).couldDoNothing(pl placement, e ability.Effect) bool`** in `placement.go`, next to `slotsLegal` and `placeable`: true when `e` has no `Modes`, has at least one target slot, every step's `Subject` is `ability.SubjectTarget`, and every slot's `g.targetDecision(pl, i, t).Options` is empty. An effect with modes, or with no target slots, is never judged here (false). Its comment cites `SVY-R48` and `FR-MOD-001` §7 and says only an optional trigger reads it; a mandatory trigger is placed anyway (`SVY-R30`).
- **`placeTrigger`** ([trigger.go:236](../../../pkg/engine/trigger.go)): `if p.trig.Optional && (g.couldDoNothing(p.placement, p.trig.Effect) || !g.accepts(p)) { return }`. Update its comment, `flushTriggers`' comment ("An optional trigger is asked about…") and `placeable`'s last sentence ("A trigger is still placed whatever it finds") to name the `SVY-R48` exception.
- **The cards**, one commit: CSV `rulesText` → `Unbowed: You may deal 1 damage to chosen asset in the Main Theater.` and `When this unit is killed, you may deal damage equal to its power to chosen non-Headquarters asset.`; the same strings as each definition's `Text`; `Optional: true` on each trigger. Keep the CSV's line endings and quoting as they are (edit the two cells only).
- Nothing else changes: no new `Decision` kind, no change to `accepts`, `chooseTargets` or the choosers.

## Tests (written by rules-test-author; engineers never edit these)
Authorized edits to merged tests, all in the `test(T-0238)` commit:
- `pkg/engine/cards_path_captain_test.go`:
  - Every test that scripts a target answer for the Captain's controller now answers "yes" (`yes`, from `optional_upto_test.go`) to the Captain's `DecideOptional` first, and its asked-count and `asked[k]` index checks shift to match: `TestCard_DevilhornCaptain_DamagedOffersEveryMainTheaterAssetOfEitherSideButNotOpposingWard`, `TestCard_DevilhornCaptain_CanDamageAnAlliedStructureInTheMainTheater`, `TestCard_DevilhornCaptain_AlliedWardUnitInTheMainTheaterIsAnOption`, `TestCard_DevilhornCaptain_FiresInReservesAndStillOffersOnlyMainTheaterAssets` (Saboteur's allied-clause answer stays first), `TestCard_DevilhornCaptain_CanChooseItselfAndItsOwnDamageFiresUnbowedAgain` (a "yes" before each of the two triggers), `TestCard_DevilhornCaptain_LethalDamageResolvesFromDiscardAndTheDeadCaptainIsNotAnOption`, `TestCard_DevilhornCaptain_LethalDamageIsChosenByTheLastKnownController`.
  - `TestCard_DevilhornCaptain_SoleLegalTargetIsNotAsked`: player 1 answers "yes"; it is asked exactly one `DecideOptional` and no `DecideTarget`, and the hitter still takes 3.
  - `TestCard_DevilhornCaptain_TimeoutDamagesExactlyOneLegalTarget` is renamed `TestCard_DevilhornCaptain_TimeoutDeclinesAndDealsNothing`: a timeout on the "you may" declines (`FR-MOD-001` §6), so none of the Main Theater assets takes the Captain's damage and no `DecideTarget` is asked.
- `pkg/engine/cards_path_test.go`:
  - Every Hunt test that scripts a target answer, or relies on the trigger resolving, answers "yes" first: `TestCard_NecropolisHunt_KilledAttackingOffersEveryUnitAndStructureButNotHeadquartersItselfOrOpposingWard`, `TestCard_NecropolisHunt_DamageEventNamesTheDeadHuntAsSourceAndIsNonCombat`, `TestCard_NecropolisHunt_KilledOnOpponentsTurnStillTriggersAndItsControllerChooses`, `TestCard_NecropolisHunt_StalwartTargetTakesTheFullFive`, `TestCard_NecropolisHunt_ResistantTargetReducesTheFive`, `TestCard_NecropolisHunt_ArmorAbsorbsPartOfTheFive`, `TestCard_NecropolisHunt_FiveDamageDestroysAFiveHealthUnit`, `TestCard_NecropolisHunt_NoEventDetailHasAFormatError`, `TestCard_NecropolisHunt_DealsItsLastKnownEffectivePowerNotAFixedFive`. If one of these scripts its answers through a helper shared with other tests, edit the call sites, not the helper, unless every caller needs the "yes".
  - `TestCard_NecropolisHunt_SoleLegalTargetIsNotAsked`: "yes", then one `DecideOptional` and no `DecideTarget`; the sole target takes the damage.
  - `TestCard_NecropolisHunt_TimeoutDamagesExactlyOneLegalTarget` is renamed `TestCard_NecropolisHunt_TimeoutDeclinesAndDealsNothing`: a timeout declines and no asset takes the Hunt's damage.
  - `TestCard_NecropolisHunt_NoOtherUnitOrStructureAsksNothingAndDoesNothing` stays as it is (it now pins `SVY-R48` too).
- `cmd/cardtown/play_source_test.go`, `TestPlay_ScriptedSessionPrintsSourceLines`: seed 3's session no longer reaches a prompt with a source line once Path's random answers shift (the lead's dry run: 0 source lines after the change, 2 before; seeds 5–8 unchanged). Make it play seeds 1 to 8 in turn with the same args and `scriptedReader`, stop at the first session whose output has a prompt followed by its source card's line (the same line check as now), and fail only if none does. Each session must still exit 0.

New tests:
- `pkg/engine/optional_nothing_test.go` (new file; synthetic cards, reusing `arena`, `attackTrig`, `upTo`, `yes` and `no` from `optional_upto_test.go`; `// Covers: SVY-R48` plus the `FR-MOD-001` name):
  - `TestFR_MOD_001_OptionalTriggerWithNoLegalTargetIsNotAskedOrPlaced`: an optional attack trigger with one exactly-one slot that nothing matches and a `SubjectTarget` damage step: no `DecideOptional`, no `DecideTarget`, nothing placed (check the placed items as `TestCard_ChainbreakerPhalanx_DeclineIsNotPlaced` does), no damage.
  - `TestFR_MOD_001_OptionalUpToTriggerWithNoLegalTargetIsNotAsked`: the same with an "up to 2" slot.
  - `TestFR_MOD_001_OptionalTriggerWithOneFillableSlotOfTwoIsStillAsked`: two slots, only the second has a legal entity: the "you may" is asked.
  - `TestFR_MOD_001_OptionalTriggerWithAnUntargetedStepIsStillAsked`: one empty slot plus a step whose subject is not `SubjectTarget` (for example a `SubjectController` draw): it is asked.
  - `TestFR_MOD_001_MandatoryTriggerWithNoLegalTargetIsStillPlaced`: the non-optional version of the first test is placed and resolves doing nothing (`SVY-R30`, `FR-TGT-003`); nothing is asked.
- `pkg/engine/cards_path_captain_test.go`:
  - `TestCard_DevilhornCaptain_DecliningDealsNoDamage`: damaged with Main Theater assets present, player 1 answers "no": one `DecideOptional`, no `DecideTarget`, no asset takes the Captain's damage.
  - `TestCard_DevilhornCaptain_LoneDamagedCaptainCanDeclineAndSurvives`: a Captain alone in the Main Theater (no other asset there on either side) takes non-lethal damage from a non-combat source; it is asked once, declines, takes no more damage and is still on the battlefield.
  - `TestCard_DevilhornCaptain_EmptyMainTheaterAsksNothing`: a Captain in Reserves is damaged while the Main Theater is empty: nothing is asked of its controller and no Captain trigger is placed (`SVY-R48`).
  - `TestCard_DevilhornCaptain_ChoosingAnAlliedUnbowedUnitFiresItsUnbowed`: player 1 has two Captains in the Main Theater; the damaged one's controller says "yes" and picks the other Captain, which takes 1 and then its own Unbowed "you may" is asked.
- `pkg/engine/cards_path_test.go`:
  - `TestCard_NecropolisHunt_DecliningDealsNoDamage`: killed with legal targets present, "no": one `DecideOptional`, no `DecideTarget`, no damage dealt by the Hunt.
- `pkg/engine/cards_path_phalanx_test.go`:
  - `TestCard_ChainbreakerPhalanx_NoAlliedUnitInReservesAsksNothing`: the Phalanx Marches to the Main Theater with no other allied unit in Reserves: no `DecideOptional`, nothing placed.
- `pkg/engine/cards_animus_apostle_test.go`:
  - `TestCard_InfestedApostle_NoOpposingUnitIsStillAsked`: deployed with no opposing unit on the battlefield: the "you may" is still asked, because its kill is a pick, not a target (`SVY-R14`, `SVY-R48`).
- `pkg/engine/cards_path_remnant_test.go`:
  - `TestCard_EerieRemnant_IsStillAskedWithNothingToTarget`: discarded from hand, its "you may" is asked (its step acts on itself, not a target).

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected. If the test author needs one, it is named here before handoff.

## Allowed Files
- `pkg/engine/placement.go`
- `pkg/engine/trigger.go`
- `pkg/engine/cards/path.go`
- `data/cards/units-and-structures-path.csv`

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run "DevilhornCaptain|NecropolisHunt|ChainbreakerPhalanx|InfestedApostle|EerieRemnant|TestFR_MOD_001|TestDefs_MatchRulesText"` and `go test ./cmd/cardtown -run TestPlay_ScriptedSessionPrintsSourceLines`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks. Its win rates shift from the run branch's (P0 63.5% / P1 35.0% / draws 1.5%; the lead's dry run gave P0 58.5% / P1 39.5% / draws 2.0%), which the user accepted (#152), so `go tool task sim-compare` is expected to differ and is not a gate here.
- [ ] Both cards still load with no unsupported lines (`cards.NotYetExecutable` unchanged; the sim header reads 50/50 for both decks).

## Log
- 2026-11-14, lead-architect (wave F.2 run session 6): written for directive #152 from the run branch at `8f215c9`. The lead dry-ran the Shape's exact edits in a scratch worktree: the build passes, and the only failures are the 20 merged Captain and Hunt tests listed under Tests (9 Captain, 11 Hunt) and `TestPlay_ScriptedSessionPrintsSourceLines` (seed 3); `pkg/sim`, `pkg/eventlog` and `pkg/view` pass; `go tool task sim` is clean at P0 58.5% / P1 39.5% / draws 2.0%. Four non-test files, within the D2.4 limit. Next: `rules-test-author`.
- 2026-11-14, rules-test-author: tests written (commit `test(T-0238)`). New: `optional_nothing_test.go` (five `TestFR_MOD_001_*`, plus helpers `optionalAsks`, `targetAsks`, `picksOf`, `abilityItemsFrom` in that file), `TestCard_DevilhornCaptain_{DecliningDealsNoDamage,LoneDamagedCaptainCanDeclineAndSurvives,EmptyMainTheaterAsksNothing,ChoosingAnAlliedUnbowedUnitFiresItsUnbowed}` (with a `pingCaptain` scene helper using `deployTargeter` for the non-combat damage), `TestCard_NecropolisHunt_DecliningDealsNoDamage`, `TestCard_ChainbreakerPhalanx_NoAlliedUnitInReservesAsksNothing`, `TestCard_InfestedApostle_NoOpposingUnitIsStillAsked`, `TestCard_EerieRemnant_IsStillAskedWithNothingToTarget`. Edited as authorized: the listed Captain tests (scriptPlan with a "yes", shifted ask indexes, timeout test renamed), the Hunt tests (the `huntAttacks` and `huntDefends` helpers now answer "yes" via scriptPlan; every caller needs it or is unaffected, `NoOtherUnitOrStructure...` untouched), `TestPlay_ScriptedSessionPrintsSourceLines` (seeds 1 to 8). Before the engine change 20 tests fail: the two no-legal-target FR tests, the Phalanx, 9 Captain (incl. 3 new) and 5 Hunt tests; the Mandatory, FillableSlotOfTwo, UntargetedStep, Apostle and Remnant tests pass already as regression guards, as do Hunt/Captain edits that only script a "yes" without asserting on it. `gofmt -l` clean, `go vet` clean, no helpers requested for `scenario_test.go`. Not covered: nothing left out. Next: lead review of tests.
- 2026-11-14, lead-architect (wave F.2 run session 7, tests review round 1): back to `tests`. The lead applied its dry-run Shape over `test(T-0238)` (7287366) in a scratch worktree: every listed test passes but one, `go tool task check` fails only there, and `cardtown lint --all` passes. Note:
  1. `TestCard_NecropolisHunt_DecliningDealsNoDamage` fails after the Shape (`cards_path_test.go:1096`: "the Hunt dealt 5 to asset 59"): its loop over `damageEvents` counts the Hunt's own combat damage to the Defender, whose `Source` is also the Hunt. Only the trigger's damage is non-combat (`Event.Combat`, as `TestCard_NecropolisHunt_DamageEventNamesTheDeadHuntAsSourceAndIsNonCombat` checks), so skip events with `e.Combat` there. Nothing else changes; the rest of the commit is approved as is. Make the fix in a new `test(T-0238)` commit.
- 2026-11-14, rules-test-author (round 2): note 1 fixed in a new `test(T-0238)` commit: `TestCard_NecropolisHunt_DecliningDealsNoDamage` skips `e.Combat` damage events, so only the trigger's damage counts. Nothing else changed. `gofmt -l` and `go vet` clean. Next: lead review of tests.
- 2026-11-14, lead-architect (wave F.2 run session 7, tests review round 2): tests approved, `ready`. Note 1 is fixed in af34281 (the decline check skips combat damage). With the lead's dry-run Shape applied over the branch, every listed test passes and `go tool task check` passes; without it the new and edited tests fail as the author listed. Next: `kernel-engineer` (sonnet).
- 2026-11-14, kernel-engineer: built the Shape. `couldDoNothing` added to `placement.go` after `placeable` (reuses `targetDecision` Options; no second legality copy); `placeTrigger` condition changed as specified, with the `placeTrigger`, `flushTriggers` and `placeable` comments naming `SVY-R48`. Both cards: CSV `rulesText` cells (only those two cells, LF endings kept), `Text` in `cards/path.go` and `Optional: true` on each trigger. Nothing else touched; no tests edited. Ran: the listed `go test ./pkg/engine -run ...` ok; `go test ./cmd/cardtown -run TestPlay_ScriptedSessionPrintsSourceLines` ok; `go tool task check` passes; `go tool task sim` clean, 50/50 both decks, P0 58.5% / P1 39.5% / draws 2.0% (as the lead's dry run).
