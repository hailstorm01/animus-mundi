---
id: T-0326
title: "Shared test setup: `cardWithTrigger`, `deployPaid` and `activateChoosing` replace twelve per-card helpers (directive #193; wave A.1 harness, flash)"
kind: harness
tier: flash
wave: A.1 (board directive #193, design note Deliverable 5a)
status: in-progress
branch: m3/autoloop-wave-a.1/T-0326-shared-test-setup
depends_on: [T-0325]
claims: [func:scene.cardWithTrigger, func:scene.deployPaid, func:activateChoosing]
cards: []
requirements: []
---

## Goal
Wave A's card tests wrote the same setup several times: five fixtures that copy a card and append a `HookExhausted` trigger, five wrappers that give a card, set the pool and deploy it to Reserves, and two identical choose-then-activate helpers. After this task, three shared helpers do these jobs: `(s *scene) cardWithTrigger(name string, trig ability.Trigger) *Card` and `(s *scene) deployPaid(p PlayerID, c *Card, req, tac int) *Instance` in `scenario_test.go`, and `activateChoosing(s *scene, a, target *Asset)` beside `chooseAsset` in `cards_path_maneuvers_test.go`. The twelve per-card helpers are deleted, and every call site uses the shared ones. Only test files change: the same 1726 test names pass, and the sims are byte-identical.

## Inputs
- Board directive #193 (structural review 16, finding 4), approved by the user 2026-11-26, `open: before the next Deliverables task`, the last of wave A.1's four directives. Design note `docs/architecture/m3-full-catalog.md` Deliverable 5a. Directives #190 to #192 (T-0323 to T-0325) are `applied`.
- The helpers to delete, at `autoloop/m3-wave-a.1` 1de939c1, each with its doc comment (all in `pkg/engine`):
  - **Tap-card fixtures.** Each copies the real card and appends `ability.Trigger{Hook: ability.HookExhausted, Effect: ability.Effect{Steps: []ability.Step{step}}}`, which is exactly `exhaustTrigger(step)` (`exhaust_hook_test.go:26`): `stormTapCard` (`cards_concordia_stormbolt_test.go:65`, *Stormbolt Ray*), `trayTapCard` (`cards_concordia_tempest_ray_test.go:30`, *Tempest City Ray*), `maTapCard` (`cards_soulless_armory_test.go:27`, *Mobile Armory*), `dmcTapCard` (`cards_soulless_macrocannon_test.go:28`, *Dolus Macrocannon*), `drfTapCard` (`cards_soulless_refiner_test.go:39`, *Dolus Refiner*, with the fixed step `{Op: ability.OpKill, Subject: ability.SubjectSelf}`).
  - **Paid-deploy wrappers.** Each gives the card to player 0, calls `s.setPool(0, R, 0)`, deploys to Reserves and returns the hand instance: `ecnDeploy` (`cards_concordia_nightmare_test.go:17`, R5), `wbDeployCard` (`cards_concordia_willbreaker_test.go:30`, R4), `crDeployCard` (`cards_soulless_conscript_test.go:17`, R2), `cbDeploy` (`cards_soulless_culverin_test.go:17`, R4), `ecDeploy` (`cards_soulless_earthshaker_test.go:18`, R6). `crDeployCard` is also called from `cards_soulless_hellfire_test.go`, and `ecDeploy` from `cards_soulless_max8_test.go` (R6 is MAX-8's cost too).
  - **Choose-then-activate:** `maActivate` (`cards_soulless_armory_test.go:41`) and `dmcActivate` (`cards_soulless_macrocannon_test.go:42`), both `chooseAsset(s, target); s.do(activate(m, 0))`.
- `s.card` already returns a shallow copy of the real card (`realCard` in `scenario_test.go`), so `cardWithTrigger` copies only the definition, as the fixtures did. `s.setPool` is at `activated_test.go:63`, `activate` at `activated_test.go:41`, `chooseAsset` at `cards_path_maneuvers_test.go:18`. `s.deployFromHand` (`trigger_test.go:40`) stays as it is: it sets no pool, and its callers rely on that.
- Old names under `.agents/` (the board's #193 row, archived task files, the wave A structural review) are history and stay as written.
- Rulings that apply: none. No behavior changes, and no non-test file changes.
- **Expected output difference:** none. `go tool task sim-compare -- autoloop/m3-wave-a.1` passes, and `go test -list . ./pkg/engine/...` still gives 1726 names.
- The lead dry-ran these exact edits in a throwaway worktree of the run branch at 1de939c1: `gofmt -l` clean, `go vet` clean, `go tool task check` exit 0, `go tool task sim-compare -- autoloop/m3-wave-a.1` `[PASS]`, 1726 tests. Two lead mutants: `cardWithTrigger` without the append fails 13 tests, and `deployPaid` setting `req-1` fails 75. Resulting blobs (`git hash-object`, first seven characters): `scenario_test.go` 5e394c4, `cards_path_maneuvers_test.go` 2251863, `cards_concordia_nightmare_test.go` d0cd884, `cards_concordia_stormbolt_test.go` 26fd7ac, `cards_concordia_tempest_ray_test.go` 616a0b0, `cards_concordia_willbreaker_test.go` 69b3497, `cards_soulless_armory_test.go` f247abe, `cards_soulless_conscript_test.go` 94d19be, `cards_soulless_culverin_test.go` 5c45d9f, `cards_soulless_earthshaker_test.go` 85eca79, `cards_soulless_hellfire_test.go` 5decd11, `cards_soulless_macrocannon_test.go` d42f431, `cards_soulless_max8_test.go` 755754f, `cards_soulless_refiner_test.go` d604df1.

## Tests
No test is added or removed. These 86 existing tests are authorized edits (D2.4). Each changes only by Edit 4's call-site rewrites: a call to a deleted helper becomes a call to a shared one, with the same card and pool.
- `pkg/engine/cards_concordia_nightmare_test.go` (11): `TestCard_EchoingNightmare_ControlledUnitGoesToItsOwnersHand`, `TestCard_EchoingNightmare_EntersReservesAsSixTwo`, `TestCard_EchoingNightmare_MoveFiresMovesHooks`, `TestCard_EchoingNightmare_MovedNightmareMayAttack`, `TestCard_EchoingNightmare_NoTargetStaysInReserves`, `TestCard_EchoingNightmare_OnlyOpposingMainTheaterUnitsOffered`, `TestCard_EchoingNightmare_ReturnsTheUnitAndMoves`, `TestCard_EchoingNightmare_SourceGoneStillReturns`, `TestCard_EchoingNightmare_TargetGoneNoMove`, `TestCard_EchoingNightmare_TargetLeftTheMainTheaterNoMove`, `TestCard_EchoingNightmare_WardedUnitNotOffered`
- `pkg/engine/cards_concordia_stormbolt_test.go` (2): `TestCard_StormboltRay_CostsPaidBeforeItResolves`, `TestCard_StormboltRay_GoneBeforeResolutionStaysGone`
- `pkg/engine/cards_concordia_tempest_ray_test.go` (3): `TestCard_TempestCityRay_SourceGoneStillMoves`, `TestCard_TempestCityRay_TargetGoneMovesNothing`, `TestCard_TempestCityRay_TargetNowAlliedMovesNothing`
- `pkg/engine/cards_concordia_willbreaker_test.go` (1): `TestCard_WillBreaker_SourceGoneStillKills`
- `pkg/engine/cards_soulless_armory_test.go` (10): `TestCard_MobileArmory_BuffStaysWhenTheAllyMovesAfterwards`, `TestCard_MobileArmory_ChosenAllyGetsPlusOnePower`, `TestCard_MobileArmory_FromReservesOffersOwnReservesOnly`, `TestCard_MobileArmory_MayActivateTheTurnItDeploys`, `TestCard_MobileArmory_MayChooseItself`, `TestCard_MobileArmory_PowerExpiresAtEndOfTurn`, `TestCard_MobileArmory_SourceGoneStillBuffs`, `TestCard_MobileArmory_SourceMovedStillBuffsInTheLockedZone`, `TestCard_MobileArmory_TargetGoneGetsNothing`, `TestCard_MobileArmory_TargetLeftTheZoneGetsNothing`
- `pkg/engine/cards_soulless_conscript_test.go` (10): `TestCard_ConscriptRifles_AloneDealsOneToItself`, `TestCard_ConscriptRifles_DealsOneToTheChosenUnit`, `TestCard_ConscriptRifles_EntersReservesAsTwoTwo`, `TestCard_ConscriptRifles_LethalDamageKills`, `TestCard_ConscriptRifles_OffersEveryNonScratchUnit`, `TestCard_ConscriptRifles_ResistantAndArmorApply`, `TestCard_ConscriptRifles_SourceGoneStillDeals`, `TestCard_ConscriptRifles_TargetGoneDealsNothing`, `TestCard_ConscriptRifles_UnitTurnedFaceDownNotOffered`, `TestCard_ConscriptRifles_WardedOpposingUnitNotOffered`
- `pkg/engine/cards_soulless_culverin_test.go` (12): `TestCard_CulverinBrigade_ArmorAbsorbs`, `TestCard_CulverinBrigade_DealsOneToEachOfTwoChosen`, `TestCard_CulverinBrigade_EntersReservesAsFiveFour`, `TestCard_CulverinBrigade_LethalDamageKills`, `TestCard_CulverinBrigade_MayChooseNone`, `TestCard_CulverinBrigade_MayChooseOne`, `TestCard_CulverinBrigade_OffersEveryOpposingAsset`, `TestCard_CulverinBrigade_OneTargetGoneTheOtherStillTakesOne`, `TestCard_CulverinBrigade_OnlyTheHeadquartersStillAsks`, `TestCard_CulverinBrigade_SourceGoneStillDeals`, `TestCard_CulverinBrigade_TargetGainsWardDealsNothingToIt`, `TestCard_CulverinBrigade_TimeoutChoosesNone`
- `pkg/engine/cards_soulless_earthshaker_test.go` (12): `TestCard_EarthshakerColossus_AloneStillAsks`, `TestCard_EarthshakerColossus_ArmorAbsorbs`, `TestCard_EarthshakerColossus_DefendingDoesNotTrigger`, `TestCard_EarthshakerColossus_DeployDamagesTheOpposingHeadquarters`, `TestCard_EarthshakerColossus_DeployDealsTwoToTheChosenAsset`, `TestCard_EarthshakerColossus_DeployOffersEveryNonScratchAsset`, `TestCard_EarthshakerColossus_EntersReservesAsSixSeven`, `TestCard_EarthshakerColossus_LethalDamageKills`, `TestCard_EarthshakerColossus_SourceGoneStillDeals`, `TestCard_EarthshakerColossus_TargetGoneDealsNothing`, `TestCard_EarthshakerColossus_TimeoutTakesTheFirstOption`, `TestCard_EarthshakerColossus_UnitTurnedFaceDownNotOffered`
- `pkg/engine/cards_soulless_hellfire_test.go` (1): `TestCard_HellfireArmorers_HasWard`
- `pkg/engine/cards_soulless_macrocannon_test.go` (9): `TestCard_DolusMacrocannon_ArmorAbsorbs`, `TestCard_DolusMacrocannon_DamagesTheOpposingHeadquarters`, `TestCard_DolusMacrocannon_DealsTwoToTheChosenAsset`, `TestCard_DolusMacrocannon_LethalDamageKills`, `TestCard_DolusMacrocannon_MayActivateTheTurnItDeploys`, `TestCard_DolusMacrocannon_MayChooseItself`, `TestCard_DolusMacrocannon_SourceGoneStillDeals`, `TestCard_DolusMacrocannon_TargetGainsWardDealsNothing`, `TestCard_DolusMacrocannon_TargetGoneDealsNothing`
- `pkg/engine/cards_soulless_max8_test.go` (14): `TestCard_MAX8AutonomousWarsuit_AlliedWardedUnitIsOffered`, `TestCard_MAX8AutonomousWarsuit_ArmorAbsorbsAndHealLeavesArmor`, `TestCard_MAX8AutonomousWarsuit_AsksTwoSlotsDamageFirst`, `TestCard_MAX8AutonomousWarsuit_CanDamageItself`, `TestCard_MAX8AutonomousWarsuit_CanHealAHeadquarters`, `TestCard_MAX8AutonomousWarsuit_DamageTargetGoneStillHeals`, `TestCard_MAX8AutonomousWarsuit_EntersReservesAsFourSeven`, `TestCard_MAX8AutonomousWarsuit_HealNeverGoesBelowZero`, `TestCard_MAX8AutonomousWarsuit_HealTargetGoneStillDamages`, `TestCard_MAX8AutonomousWarsuit_HealsAnotherAssetLethallyDamagedOneDies`, `TestCard_MAX8AutonomousWarsuit_SameAssetForBothTakesTwoThenHealsTwo`, `TestCard_MAX8AutonomousWarsuit_ScratchCompanyCanBeKilled`, `TestCard_MAX8AutonomousWarsuit_SourceGoneStillResolves`, `TestCard_MAX8AutonomousWarsuit_TimeoutTakesTheFirstOptionForEach`
- `pkg/engine/cards_soulless_refiner_test.go` (1): `TestCard_DolusRefiner_SourceGoneStillDrawsAndDiscards`

Three setup functions that are not tests change the same way: `wbDeploy` (`cards_concordia_willbreaker_test.go`), `maHookScene` (`cards_soulless_armory_test.go`) and `ecAttackScene` (`cards_soulless_earthshaker_test.go`). Every other test must pass unchanged.

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)
1. **`pkg/engine/scenario_test.go`**, two insertions. Right after the closing `}` of `func (s *scene) give(`, a blank line and then:
   ```go
   // deployPaid gives p the card c, sets p's pool to req Requisition and tac
   // Tactics, deploys it from hand to Reserves, and returns its hand instance.
   func (s *scene) deployPaid(p PlayerID, c *Card, req, tac int) *Instance {
   	s.t.Helper()
   	inst := s.give(p, c)
   	s.setPool(p, req, tac)
   	s.do(Action{Kind: ActDeploy, Player: p, Card: inst.ID, Zone: ZoneReserves})
   	return inst
   }
   ```
   Right after the closing `}` of `func (s *scene) card(name string) *Card`, a blank line and then:
   ```go
   // cardWithTrigger returns a copy of the real named card whose definition's
   // triggers also hold trig. The shared definition is not changed. With an
   // exhaustTrigger, paying the card's exhaust cost fires trig, so it is placed on
   // top of the ability and resolves first (FR-TRN-003a, SVY-R17).
   func (s *scene) cardWithTrigger(name string, trig ability.Trigger) *Card {
   	s.t.Helper()
   	c := s.card(name)
   	if c.Ability == nil {
   		s.t.Fatalf("%s has no executable definition", name)
   	}
   	d := *c.Ability
   	d.Triggers = append(append([]ability.Trigger(nil), d.Triggers...), trig)
   	c.Ability = &d
   	return c
   }
   ```
2. **`pkg/engine/cards_path_maneuvers_test.go`:** right after the closing `}` of `func chooseAsset(`, a blank line and then:
   ```go
   // activateChoosing activates a's first activated ability, player 0 choosing
   // target for every decision it asks (chooseAsset).
   func activateChoosing(s *scene, a, target *Asset) {
   	s.t.Helper()
   	chooseAsset(s, target)
   	s.do(activate(a, 0))
   }
   ```
3. **Delete the twelve helpers** listed under Inputs, each with its doc comment (the `// <name> …` lines right above it) and the blank line after its closing `}`. Nothing around them changes.
4. **Rewrite every call.** Keep each call's receiver and card argument as written (`s` or `s2`; `c`, `c2`, `s.card("Culverin Brigade")`, `s2.card(mxName)` and so on):

   | Old call | New call |
   | :--- | :--- |
   | `ecnDeploy(X, C)` | `X.deployPaid(0, C, 5, 0)` |
   | `wbDeployCard(X, C)` | `X.deployPaid(0, C, 4, 0)` |
   | `crDeployCard(X, C)` | `X.deployPaid(0, C, 2, 0)` |
   | `cbDeploy(X, C)` | `X.deployPaid(0, C, 4, 0)` |
   | `ecDeploy(X, C)` | `X.deployPaid(0, C, 6, 0)` |
   | `stormTapCard(s, STEP)` | `s.cardWithTrigger("Stormbolt Ray", exhaustTrigger(STEP))` |
   | `trayTapCard(s, STEP)` | `s.cardWithTrigger("Tempest City Ray", exhaustTrigger(STEP))` |
   | `maTapCard(s, STEP)` | `s.cardWithTrigger("Mobile Armory", exhaustTrigger(STEP))` |
   | `dmcTapCard(s, STEP)` | `s.cardWithTrigger("Dolus Macrocannon", exhaustTrigger(STEP))` |
   | `drfTapCard(s)` | `s.cardWithTrigger("Dolus Refiner", exhaustTrigger(ability.Step{Op: ability.OpKill, Subject: ability.SubjectSelf}))` |
   | `maActivate(X, M, T)`, `dmcActivate(X, M, T)` | `activateChoosing(X, M, T)` |

   A call assigned to a variable keeps the assignment: `inst := ecnDeploy(s, c)` becomes `inst := s.deployPaid(0, c, 5, 0)`, and `rifles := crDeployCard(s, s.card("Conscript Rifles"))` becomes `rifles := s.deployPaid(0, s.card("Conscript Rifles"), 2, 0)`. The 12 card test files hold 98 calls: 69 deploys (nightmare 13, willbreaker 2, conscript 12, hellfire 1, culverin 12, earthshaker 13, max8 16), 10 tap cards (stormbolt 2, tempest ray 3, armory 1, macrocannon 3, refiner 1) and 19 activates (armory 11, macrocannon 8).
5. **`pkg/engine/cards_soulless_max8_test.go` header comment:** replace `// and is the active player. Uses ecDeploy from cards_soulless_earthshaker_test.go.` with `// and is the active player, deploying it from hand to Reserves with exactly R6 T0.`

Edit with the file tool, file by file. Then run `go -C <worktree> tool task fmt`, and check every blob against the Inputs with `git -C <worktree> hash-object pkg/engine/<file>`. A blob that differs means an edit was missed or doubled; `git -C <worktree> diff -- pkg/engine/<file>` shows where.

## Commits
1. `test(T-0326): shared cardWithTrigger, deployPaid and activateChoosing setup, for directive 193`, with Edits 1-5. It is the only commit besides the status commits, as in T-0323: every change is test code, so there is no implementation commit. No double quotes and no `#` in the message.

## Acceptance Cases (harness tasks only; written by the lead)
1. **No old helper is left in the code:** `git -C <worktree> grep -n -w -e stormTapCard -e trayTapCard -e maTapCard -e dmcTapCard -e drfTapCard -e ecnDeploy -e wbDeployCard -e crDeployCard -e cbDeploy -e ecDeploy -e maActivate -e dmcActivate -- pkg cmd docs` finds nothing. Exit code 1 with no output is the pass.
2. **The shared helpers are used:** `git -C <worktree> grep -n -w -e deployPaid -- pkg/engine` finds 71 lines (the doc comment, the `func` line and 69 calls); with `-e cardWithTrigger`, 12 lines (doc comment, `func` line, 10 calls); with `-e activateChoosing`, 21 lines (doc comment, `func` line, 19 calls).
3. **The same tests pass:** `go -C <worktree> test -list . ./pkg/engine/...` gives 1726 test names, as on the run branch, and `go -C <worktree> test ./...` passes.
4. `git -C <worktree> diff --stat autoloop/m3-wave-a.1 HEAD` lists only the 14 files under Allowed Files and this task file.
5. `go -C <worktree> tool task sim-compare -- autoloop/m3-wave-a.1` passes.
6. `go -C <worktree> tool task check` passes.

## Allowed Files
- Tests, all in `pkg/engine` (no non-test file changes): `scenario_test.go`, `cards_path_maneuvers_test.go`, `cards_concordia_nightmare_test.go`, `cards_concordia_stormbolt_test.go`, `cards_concordia_tempest_ray_test.go`, `cards_concordia_willbreaker_test.go`, `cards_soulless_armory_test.go`, `cards_soulless_conscript_test.go`, `cards_soulless_culverin_test.go`, `cards_soulless_earthshaker_test.go`, `cards_soulless_hellfire_test.go`, `cards_soulless_macrocannon_test.go`, `cards_soulless_max8_test.go`, `cards_soulless_refiner_test.go`. The twelve card test files change only to delete a helper and update the references to it, so under the D2.4 move exception they don't count toward the four-file limit.

## Done When
- [ ] Edits 1-5 are committed as the one `test(T-0326): …` commit
- [ ] Cases 1-5 hold, with their commands and results in the Log
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim-compare -- autoloop/m3-wave-a.1` passes (starter and stress sims byte-identical to the run branch)

## Log
- 2026-11-26 lead-architect: Wrote the task from board directive #193, the last of wave A.1's four directives (design note Deliverable 5a). Harness, tier flash: three shared setup helpers and the call-site updates in 14 test files, with no non-test change. The lead dry-ran the edits (Inputs). Status `ready` on `m3/autoloop-wave-a.1/T-0326-shared-test-setup`, from `autoloop/m3-wave-a.1`. Once this task is `done`, all of wave A.1's header work is applied.
