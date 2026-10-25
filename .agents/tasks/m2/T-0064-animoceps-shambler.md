---
id: T-0064
title: Make Animoceps Shambler executable (wave B card, flash)
kind: card
tier: flash
wave: B
status: ready
branch: m2/T-0064-animoceps-shambler
depends_on: [T-0063]
claims: [card:Animoceps Shambler]
cards: [Animoceps Shambler]
requirements: []
---

## Goal
*Animoceps Shambler* (Animus unit, 2/2, three copies in the Animus starter deck) runs in the kernel: "Whenever this unit exhausts, chosen allied asset in the Main Theater heals 1." It is plain `ability` data in `pkg/engine/cards/animus.go`.

## Inputs
- Card (`data/cards/*.csv`): Animus unit, Fungus, cost 2, 2/2, yields 1R 2T, deploys to Reserves.
- Survey row: `docs/design/m2_starter_set_ability_survey.md` line 107.
- Shapes it uses: `ability.HookExhausted` (T-0063), a target with an allied filter and a Main Theater zone (T-0017, T-0052), and `OpHeal` (T-0019).
- Rulings: `SVY-R17` (placement), `SVY-R3` (targets locked on placement), `SVY-R16` (timeout picks the first legal option in fixed order).

### Definition (the engineer adds it to `animusDefs`)
- One trigger: `Hook: ability.HookExhausted`, one target slot with `Filter{Allegiance: ability.AllegianceAllied, Zone: ability.ZoneMainTheater}` (zero `Kinds`: any asset).
- One step: `Op: ability.OpHeal`, `Amount: ability.Value{N: 1}`, `Subject: ability.SubjectTarget`, slot 0.
- `Text` holds the printed rules text, one entry.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
Written in `pkg/engine/cards_animus_test.go` (8f4b916):
- `TestCard_AnimocepsShambler_DeploysToReservesReadiedAsTwoTwoAndAsksNothing`: deploys to Reserves for 2R as a 2/2 yielding 1R 2T, Readied, and deploying does not fire the exhaust trigger (`FR-CAP-001`, `SVY-R17`).
- **Named (directive #47 part 4):** `TestCard_AnimocepsShambler_MarchTriggerResolvesBeforeTheMarch`: a damaged Shambler marching from the Main Theater to Reserves heals itself while still in the Main Theater before the March moves it (`FR-CAP-001`, `FR-QUE-001`, `FR-TGT-003`, `SVY-R17`).
- **Named (directive #47 part 4):** `TestCard_AnimocepsShambler_GenerateTriggerResolvesBeforeTheNextAction`: generating Tactics in the Main Theater yields 2T, places the trigger and chooses its target after the generation event, and heals the chosen ally before the next action (`FR-CAP-001`, `FR-QUE-001`, `SVY-R17`, `SVY-R3`).
- `TestCard_AnimocepsShambler_MarchingFromReservesCannotChooseItself`: target is locked on placement while the Shambler is still in Reserves, so only Main Theater allies are offered (`FR-TGT-001`, `SVY-R3`).
- `TestCard_AnimocepsShambler_AttackingFromTheMainTheaterMayChooseItself`: an attacking Shambler in the Main Theater is offered alongside other Main Theater allies (`FR-TGT-001`, `FR-CBT-002`).
- `TestCard_AnimocepsShambler_OffersOnlyAlliedAssetsInTheMainTheater`: offers allied units, structures, Scratch Companies and Warded allies in the Main Theater; excludes Reserves, Headquarters and opposing assets (`FR-TGT-001`, `FR-KWD-002h`).
- `TestCard_AnimocepsShambler_NoLegalTargetAsksNothingAndHealsNothing`: with no allied asset in the Main Theater, nothing is asked, nothing heals, and the March still resolves (`FR-TGT-003`).
- `TestCard_AnimocepsShambler_HealsExactlyOne`: removes 1 damage from a chosen ally with 3 damage (`FR-KWD-002d`).
- `TestCard_AnimocepsShambler_HealingAnUndamagedAssetChangesNothing`: targeting an undamaged ally leaves 0 damage and both assets alive (`FR-KWD-002d`).
- `TestCard_AnimocepsShambler_AttackHealResolvesBeforeCombatDamage`: a 2/2 Shambler with 1 damage attacks a 1-power defender, heals to full at Step 2 before combat damage, and survives with 1 damage (`FR-CBT-002`, `FR-KWD-002d`, `SVY-R17`).
- `TestCard_AnimocepsShambler_TimeoutHealsTheFirstAlliedAssetToEnterTheMainTheater`: timeout selects the first allied asset in the Main Theater in entry order (`FR-MOD-001` §6, `SVY-R16`).
- `TestCard_AnimocepsShambler_HealsAgainOnEachExhaust`: fires on Generate in one turn and again on March in a later turn (`FR-CAP-001`).
- `TestCard_AnimocepsShambler_DefendingDoesNotHeal`: being attacked does not exhaust the defender, so nothing fires (`FR-CBT-002`).
- `TestCard_AnimocepsShambler_FaceDownCopyExhaustingHealsNothing`: a face-down copy exhausting does not fire the printed trigger (`FR-ZON-005`).

## Helpers Needed
None. The test file defines local `putShambler`, `putDamaged` and `firstTargetDecision` helpers and reuses existing scenario helpers.

## Allowed Files
- `pkg/engine/cards/animus.go`
- `pkg/engine/cards/cards.go` (remove the name from `NotYetExecutable`)

## Done When
- [ ] The card's tests pass: `go test ./pkg/engine -run 'TestCard_AnimocepsShambler'`
- [ ] `TestDefs_MatchRulesText` and `TestStarterSet_Executable` pass unchanged
- [ ] `go tool task check` passes
- [ ] `go tool task sim` runs clean; report the before and after summary lines in the Log

## Log
- 2026-10-25 lead-architect: Wrote the task, `queued` behind T-0063 (the exhaust hook).
- 2026-10-25 lead-architect: T-0063 is `done` (3a9e549). Re-checked against `main` 3a9e549: `ability.HookExhausted`, `Filter.Allegiance`/`Filter.Zone` (`ability.go:69-75`, zero `Kinds` means any asset) and `OpHeal` exist, so the card needs no new shape. Spelled out the target filter in the Definition. Waiting triggers checked: #6, #7, #8, #15, #20 and #51. None has fired. Directive #14: the Shambler is a deck card (3 copies, Animus starter).
  - Next: status `tests`; dispatch `rules-test-author`.
- 2026-10-25 rules-test-author (run on Opus in Antigravity at the user's choice): Wrote 14 `TestCard_AnimocepsShambler_*` tests in `pkg/engine/cards_animus_test.go` (8f4b916), listed under Tests, including both directive #47 part 4 names. Sources: FR-CAP-001 / SVY-R17 (placement), FR-TGT-001 / SVY-R3 (locked targets), FR-TGT-003 (re-verification), FR-MOD-001 §6 / SVY-R16 (timeout order), FR-KWD-002d (Heal), FR-CBT-002 Step 2, FR-ZON-005. Helpers requested: none. `go vet ./pkg/engine` clean; `cardtown lint rules` PASS; 10 of the 14 fail as expected until the card is executable (deploy is rejected, the trigger does not fire); 4 negative cases (no legal target, undamaged, defending, face-down) pass already.
  - Left out: a target leaving the Main Theater between placement and resolution (no starter-set way to move it mid-stack from this trigger; the generic re-verification is covered by the targeting shape tests); an opposing Shambler (nothing in M2 exhausts an asset on the opponent's turn); healing above maximum health (Heal 1 on a damaged asset cannot overshoot; covered generically in `heal_armor_test.go`).
  - Next: status `tests-review`; lead reviews the tests.
- 2026-10-25 lead-architect (run on Gemini Pro High at the user's choice after Opus 429 quota): Tests review (D4): **approved**, status `ready`. All 14 tests in 8f4b916 match FRD 01, `SVY-R17`, `SVY-R3`, `SVY-R16` and the printed card text (`data/cards/*.csv`).
  - **Coverage:** both named directive #47 part 4 tests (`_MarchTriggerResolvesBeforeTheMarch`, `_GenerateTriggerResolvesBeforeTheNextAction`), locked targets on placement (`_MarchingFromReservesCannotChooseItself`, `_AttackingFromTheMainTheaterMayChooseItself`), filter for allied assets in the Main Theater (`_OffersOnlyAlliedAssetsInTheMainTheater`, `_NoLegalTargetAsksNothingAndHealsNothing`), Heal 1 behavior and Step 2 timing (`_HealsExactlyOne`, `_HealingAnUndamagedAssetChangesNothing`, `_AttackHealResolvesBeforeCombatDamage`), fixed-order timeout (`_TimeoutHealsTheFirstAlliedAssetToEnterTheMainTheater`), repeated exhausts across turns (`_HealsAgainOnEachExhaust`), defending (`_DefendingDoesNotHeal`), face-down (`_FaceDownCopyExhaustingHealsNothing`), and printed characteristics/deployment (`_DeploysToReservesReadiedAsTwoTwoAndAsksNothing`).
  - **Independent oracle:** verified `rules-test-author` subagent transcript (`94551256-5fdc-488b-a863-c198c57e7eed`); it never read or edited any non-test file in `pkg/engine/`.
  - Next: dispatch `kernel-engineer` (tier `flash`).
