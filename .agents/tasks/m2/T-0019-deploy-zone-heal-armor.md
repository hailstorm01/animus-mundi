---
id: T-0019
title: Gate deploy triggers on the deployed zone, and add Heal and AddArmor steps
kind: shape
tier: sonnet
wave: A2 (task 3)
status: done
branch: m2/T-0019-deploy-zone-heal-armor
depends_on: [T-0017]
claims: [type:ability.Zone, field:Trigger.ToZone, gate:deploy-zone, op:Heal, op:AddArmor]
cards: []
requirements: [FR-KWD-002d, FR-LAY-002, FR-TRN-003]
---

## Goal
Three shapes:
- **Deploy-zone gate:** "When this unit deploys, if it was deployed to the Main Theater, …" fires only when the unit entered that zone.
- **Heal:** "chosen asset heals N" removes up to N accumulated damage.
- **AddArmor:** "chosen unit gains N armor" adds N to the asset's Armor pool.

After this task, *Devilhorn Guerrillas*, *Spoil Forest Colossus* and *Ironclad Colossus* can be written as card data. The first two are in the starter decks.

## Inputs
- FRD 01 (`docs/frd/01_Game_Rules_FRD.md`):
  - `FR-KWD-002d` Heal N (line 491): removes up to N accumulated damage. It can't go below 0 or raise Current Health above Effective Maximum Health.
  - `FR-LAY-002` (lines 389–393), the Dynamic Armor Pool: a grant adds to Armor, grants stack additively, Armor never expires at end of turn, and it absorbs damage 1-for-1 in Stage 4 (`FR-REP-004`).
  - `FR-KWD-002k` Armor N: printed Armor and granted Armor share one pool.
  - `FR-ZON-004` (line 165): dynamic Armor and damage are removed when the asset leaves the battlefield. Line 166: moving between Reserves and the Main Theater keeps them.
  - `FR-TRN-003` (line 273): a Unit enters its declared battlefield zone.
- Card text for context. The card tasks that follow use these, and this task's tests use synthetic cards:
  - *Devilhorn Guerrillas* (deploy zones Main Theater, Reserves): "When this unit deploys, if it was deployed to the Main Theater, deal 2 damage to chosen opposing unit."
  - *Spoil Forest Colossus*: "Whenever this unit deploys or attacks, chosen asset heals 2."
  - *Ironclad Colossus*: "Whenever this unit deploys or attacks, chosen unit gains 2 armor."
  - "Deploys or attacks" is two triggers with the same effect, written as `HookDeploy` and `HookAttacks` under one `Text` line (precedent: *Fungal Evangelist*, "deploys or is killed").
- Built on: T-0017 (`Effect.Targets`, `Filter`, `SubjectTarget`, `chooseTargets`, `legalTarget`, re-verification in `resolveAbility`), T-0007 (`fireDeploy`, called from `stack.go` once the deploy resolves) and T-0010 (`HookAttacks`).
- Rulings that apply: none. FRD 01 has no rule for when a "was deployed to" condition is checked. The zone a unit entered can't change, and nothing can move it before its trigger resolves, because March needs an empty stack (`FR-QUE-001`). So checking it when the deploy fires is complete.

### Vocabulary (`pkg/engine/ability/ability.go`)
```go
// Zone names a battlefield zone a shape refers to.
type Zone int

const (
	ZoneAny         Zone = iota // the zero value: no zone condition
	ZoneReserves
	ZoneMainTheater
)

// Trigger gains:
	ToZone Zone // HookDeploy only: fires only if the asset was deployed into this zone

const ( // Op gains:
	OpHeal     // remove up to Amount accumulated damage from the Subject (FR-KWD-002d)
	OpAddArmor // add Amount to the Subject's Armor pool (FR-LAY-002)
)
```
Names may be adjusted for clarity if needed. Record any rename in the Log. Later tasks reuse `ability.Zone` for zone filters, so keep it free of engine types.

### Lead decisions (technical, not rules)
- **Deploy-zone gate:** checked in `fireHook` when `HookDeploy` fires, against the zone the asset is in at that moment, which is the zone its deploy put it in. A trigger whose gate fails is never buffered, so its targets are never asked for. `ToZone` on any other hook is ignored. No card builds that.
- **Heal:** `Damage = max(0, Damage - N)` on the chosen asset. Damage never exceeds what the Effective Maximum Health allows, so "not above Effective Maximum Health" follows. Healing an undamaged asset does nothing. Heal doesn't touch Armor.
- **AddArmor:** `Armor += N` on the chosen asset. No end-of-turn expiry. Leaving the battlefield already discards the `Asset`, so the pool goes with it.
- Both ops use `SubjectTarget` with T-0017's choose-at-placement and re-verification, unchanged. A Heal or AddArmor slot with no legal target does nothing.
- **Not in this task:** "up to N", "other", zone filters on targets, "Heal N" as a keyword on cards (Survey wave B), distributed armor pools (`FR-TGT-002`), and Regroup.

### What the tests can observe
- **Gate:** a synthetic unit with deploy zones Main Theater and Reserves and a gated deploy trigger. Deployed to the Main Theater, it fires (a target is asked, or fast-pathed, and damage lands). Deployed to Reserves, nothing is asked and nothing happens. An ungated twin fires in both zones. A gate for Reserves behaves the mirror way.
- **Heal:** a damaged asset heals by N. If its damage is below N, it goes to 0, not negative. An undamaged asset is unchanged. Any kind of asset can be healed, including a Headquarters and an allied or opposing one, as the filter allows. Heal leaves Armor alone. Ward still blocks an opponent choosing it (T-0017).
- **AddArmor:** a unit gains N Armor. A second grant stacks. A unit with printed Armor adds to its pool. The pool still holds after the turn ends. Later damage is absorbed 1-for-1 before health (`FR-REP-004` Stage 4). After the unit marches between Reserves and the Main Theater it keeps the Armor (FR-ZON-004 line 166).
- **Deploys-or-attacks:** a unit with the same Heal effect on `HookDeploy` and `HookAttacks` heals once when deployed and once each time it attacks. On attack, the heal resolves at Step 2, before combat damage (`FR-CBT-002`). So healing the attacker itself doesn't save it from lethal retaliation, but healing an already damaged ally is visible after combat.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
All in `pkg/engine/heal_armor_test.go` (synthetic cards; targets chosen by asset with `scriptRefs`).
- Gate (FR-TRN-003): `TestFR_TRN_003_GatedDeployTriggerFiresWhenDeployedToMainTheater`, `..._GatedDeployTriggerFastPathsASingleTargetInMainTheater`, `..._GatedDeployTriggerDoesNothingWhenDeployedToReserves`, `..._UngatedDeployTriggerFiresInEitherZone`, `..._ReservesGateFiresInReservesAndNotInMainTheater`, `..._FailedGateTriggerIsNotOrderedWithItsSibling`, `..._PassedGateTriggerIsOrderedWithItsSibling`.
- Heal (FR-KWD-002d): `TestFR_KWD_002d_HealRemovesUpToNDamage`, `..._HealNeverGoesBelowZeroDamage`, `..._HealingAnUndamagedAssetChangesNothing`, `..._HealLeavesArmorAlone`, `..._AnyKindOfAssetCanBeHealed` (8 subtests: unit, structure, Scratch Company, HQ, both sides), `..._OpposingWardUnitCannotBeChosenToHeal`, `..._HealWithNoLegalTargetDoesNothing`.
- Armor: `TestFR_LAY_002_ArmorGrantAddsToThePool`, `..._SeparateGrantsStackAdditively`, `..._GrantedArmorDoesNotExpireAtEndOfTurn`, `..._GrantedArmorAbsorbsLaterDamageOneForOneBeforeHealth`, `TestFR_KWD_002k_GrantedArmorJoinsPrintedArmorInOnePool`, `TestFR_ZON_004_MarchKeepsGrantedArmor`.
- Deploys-or-attacks (FR-CBT-002): `TestFR_CBT_002_DeployOrAttackHealsOnceOnDeployAndOnceEachAttack`, `..._AttackHealResolvesBeforeCombatDamage`, `..._AttackHealOfTheAttackerDoesNotSaveItFromLethalRetaliation`, `..._AttackHealOfADamagedAllyIsVisibleAfterCombat`.

Run: `go test ./pkg/engine -run 'TestFR_(TRN_003_(Gated|Ungated|Reserves|Failed|Passed)|KWD_002d|LAY_002|KWD_002k_Granted|ZON_004_MarchKeepsGranted|CBT_002_(DeployOrAttack|AttackHeal))'`

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None. The tests define their own local helpers (`deployTo`, `healStep`, `armorStep`, `gatedDeploy`, `zonedUnit`, `orderAsks`) in `heal_armor_test.go`.

## Allowed Files
- `pkg/engine/ability/ability.go`
- `pkg/engine/trigger.go`
- `pkg/engine/stack.go` (only if the deploy zone must be passed to `fireDeploy`)

## Done When
- [x] The listed tests pass: `go test ./pkg/engine -run '<regex from the Tests section>'`
- [x] `go tool task check` passes
- [x] `go run ./cmd/cardtown sim -deck starter -p0 ANIMUS -p1 PATH` and `go run ./cmd/cardtown sim` exit 0 with every invariant holding. No executable count changes.

## Log
- 2026-10-20 lead-architect: Wrote the task (A2 task 3), with status `tests` for `rules-test-author`.
- 2026-10-20 rules-test-author: Wrote the tests above (status tests-review). The package will not compile until the Vocabulary block exists (`ability.Zone`, `ZoneAny/ZoneReserves/ZoneMainTheater`, `Trigger.ToZone`, `OpHeal`, `OpAddArmor`); `go vet` fails only on those. Left out: (1) FR-ZON-004 line 165 (Armor and damage discarded on leaving the battlefield): not observable without redeploying the same asset, and the observation list does not call for it. (2) `ToZone` on a non-deploy hook: the task says it is ignored and no card builds it, so FRD 01 has no rule to pin. (3) "Heal can't exceed Effective Maximum Health" is pinned only through Damage clamping at 0 plus CurrentHealth == max. The task cites FR-TRN-003 at line 268, but the text is at line 273 (268 is the FR-ACT-001 heading). The attack-heal ordering test relies on FR-CBT-002 Step 2: an attacker with 4 damage on 5 health survives a 1-power retaliation only if the heal resolves first.
- 2026-10-20 lead-architect: Test review: **approved**, status `ready`. 71cfdc1 adds `heal_armor_test.go` (synthetic cards) and changes no existing test.
  - The tests cover:
    - The gate: it fires in its zone, does nothing in the other, the ungated twin fires in both, and the Reserves gate is the mirror. A gated-off trigger is never collected, so no order question is asked.
    - Heal: removes up to N, never goes below 0, leaves an undamaged asset and Armor alone, and reaches every asset kind. Ward and no-target are covered too.
    - The Armor pool: grants are additive and join printed Armor. It survives end of turn, absorbs 1-for-1 before health, and is kept through March.
    - The Step 2 attack-trigger timing, through the two retaliation cases.
  - Targets are chosen by asset.
  - Accepted omissions: line 165 (state reset on leaving the battlefield) isn't observable on a discarded `Asset`, and `ToZone` on non-deploy hooks has no rule to pin.
  - Lead fix: the task's FR-TRN-003 citation is line 273, not 268. FRD 01 v3.7 (09b9bdb) shifted the lines.
  - Next: dispatch `kernel-engineer` (tier sonnet).
- 2026-10-20 kernel-engineer: **blocked**. Implemented `ability.Zone`/`ToZone`/`OpHeal`/`OpAddArmor` (ability.go) and the `HookDeploy` gate (`deployedTo`) plus the Heal/AddArmor step in `resolveAbility` (trigger.go); no stack.go change needed (the asset's zone at fire time is the deploy zone). Reused T-0017 `chooseTargets`/`legalTarget`/re-verification and the existing `OpDealDamage` step guard shape. All heal_armor tests pass except `TestFR_KWD_002d_AnyKindOfAssetCanBeHealed/allied_Scratch_Company` and `/opposing_Scratch_Company`. Cause is the test, not the engine: `targetBoard().c0/c1` are face-down Scratch Company with Health 1 (`scratchFace`), and the subtest sets `Damage = 4`. The first death sweep after the deploy destroys the 1-health unit, so it is gone before the trigger is placed, and it is not among the target options. Healing it to 1 damage would leave it above lethal anyway, so these two subtests can't pass under FR-REP/death rules. Needs rules-test-author to use damage below the Scratch's health (e.g. Damage 0 is uninformative; raise its health or assert differently). `go tool task check` is red only for these two subtests. Both sims exit 0 with invariants PASS; executable counts unchanged (P0 19/50, P1 9/50).
- 2026-10-20 lead-architect: The block is a **test bug** that the lead's review missed, not a missing shape or a ruling. Status `tests`, to `rules-test-author` for one fix:
  1. `TestFR_KWD_002d_AnyKindOfAssetCanBeHealed`, rows `allied Scratch Company` and `opposing Scratch Company`. A face-down Scratch Company is 1/1 (FR-ZON-005), so `Damage = 4` is lethal, and the sweep removes it before the heal trigger is placed. No damage on a Scratch Company survives to be healed, so these rows can't observe a heal amount. Replace them with a check that the Scratch Company is among the options of the "any asset" heal and is chosen and resolved without error, its damage staying 0. Or drop them and say why. Leave the other six rows unchanged.
  - The fix lands as a second `test(T-0019)` commit after the engineer's ed47d09. This is an approved exception to "tests are protected after the test commit", made by the lead for this one subtest pair. The engineer's code stays as committed. When the test commit lands, the lead reviews the test change and the code together.
- 2026-10-20 rules-test-author: Fixed the two Scratch Company rows of `TestFR_KWD_002d_AnyKindOfAssetCanBeHealed` (approved exception). They no longer set Damage; they check the Scratch Company is among the options of the any-asset heal, is chosen and resolves cleanly, stays on the battlefield with Damage 0 and total damage 0. The other six rows and all other tests are unchanged. Status tests-review.
- 2026-10-20 lead-architect: Test fix and code review: **approved**, status `done`.
  - The test fix 61a0c52 touches only the two Scratch Company rows. They now check that the Scratch Company is an option, is chosen, resolves cleanly and stays at 0 damage, which is what FR-ZON-005's 1/1 allows. The other rows and tests are unchanged.
  - Code (ed47d09) changes only `ability.go` and `trigger.go`.
    - `fireHook` skips a `HookDeploy` trigger whose `ToZone` doesn't match the asset's current zone (`deployedTo`).
    - `OpHeal` is `Damage = max(0, Damage-N)` and `OpAddArmor` is `Armor += N`, both on T-0017's re-verified slots.
    - There are no card names and no new types beyond the Vocabulary.
  - Minor, no change requested: the slot guard (`Subject == SubjectTarget`, bounds, nil) is now repeated in two `case` arms. Fold it into a helper when the next targeted op lands (A2 task 4).
  - Gates re-run by the reviewer:
    - 27 heal/armor/gate results pass, and `go tool task check` exits 0.
    - The sorted starter-sim and default-sim summaries are identical to `rework/executable-rules`, and every invariant holds.
  - Process note: the lead's first test review missed the lethal-damage setup, so the engineer correctly stopped at `blocked` instead of editing a protected test.
