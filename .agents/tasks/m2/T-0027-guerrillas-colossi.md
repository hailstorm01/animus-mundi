---
id: T-0027
title: Make Devilhorn Guerrillas, Spoil Forest Colossus and Ironclad Colossus executable
kind: card
tier: flash
wave: A2 (task 4, card task 2)
status: done
branch: m2/T-0027-guerrillas-colossi
depends_on: [T-0019]
claims: [card:Devilhorn Guerrillas, card:Spoil Forest Colossus, card:Ironclad Colossus]
cards: [Devilhorn Guerrillas, Spoil Forest Colossus, Ironclad Colossus]
requirements: []
---

## Goal
Three units with targeted triggers now run in the kernel. Each one is plain `ability` data, built only from shapes that T-0017 and T-0019 already provide.
- *Devilhorn Guerrillas* (Path, 3 copies) and *Spoil Forest Colossus* (Animus, 2 copies) are starter-deck cards.
- *Ironclad Colossus* isn't in a deck. It joins because every shape it needs already exists (board directive #14c).

## Inputs
- Cards: `data/cards/units-and-structures-path.csv` and `units-and-structures-animus.csv`. Each has exactly one unsupported line today, quoted here:
  - *Devilhorn Guerrillas*: Path Unit, 4R 0T, deploys to Main Theater or Reserves, 3/5, yields 2R 0T. "When this unit deploys, if it was deployed to the Main Theater, deal 2 damage to chosen opposing unit."
  - *Spoil Forest Colossus*: Animus Unit, 6R 0T, Reserves, 6/7, yields 2R 2T. "Whenever this unit deploys or attacks, chosen asset heals 2."
  - *Ironclad Colossus*: Path Unit, 6R 0T, Reserves, 6/7, yields 2R 2T. "Whenever this unit deploys or attacks, chosen unit gains 2 armor."
- Survey rows: `docs/design/m2_starter_set_ability_survey.md` lines 96, 132 and 133.
- Shapes they use (all `done`):
  - `HookDeploy` (T-0007) and `HookAttacks` (T-0010);
  - `Effect.Targets`, `Filter{Allegiance, Kinds}`, `SubjectTarget`, `Step.Target` and `OpDealDamage` (T-0017);
  - `Trigger.ToZone`, `OpHeal` and `OpAddArmor` (T-0019).
- Requirements for context:
  - `FR-TRN-003`, the deploy-zone gate (see T-0019). A gate that fails never buffers the trigger, so no target is asked for.
  - `FR-KWD-002d`: Heal removes up to N damage, never below 0, and leaves Armor alone.
  - `FR-LAY-002` and `FR-KWD-002k`: armor grants stack in one pool with printed Armor and never expire.
  - `FR-CBT-002`: attack triggers resolve at Step 2, before Step 3 combat damage.
  - `FR-TGT-001` and `FR-TGT-003`: targets are chosen on placement and re-verified at resolution.
  - `FR-KWD-002h`: Ward blocks opponents only.
  - `FR-MOD-001` §1 and §6: the fast path and the random fallback.
- Kernel conventions from T-0017: targets are chosen in slot order when the trigger is placed. Options are listed oldest asset first. A filter without "other" includes the source itself. Tests choose by asset (`scriptRefs`), never by option index.
- Rulings that apply: none.

### Definitions (the engineer adds them)
- *Devilhorn Guerrillas*, in `pathDefs` (`pkg/engine/cards/path.go`): one `HookDeploy` trigger with `ToZone: ZoneMainTheater` and one target `{Allegiance: Opposing, Kinds: Unit}`. Its step is `DealDamage 2` to slot 0. A Scratch Company is a legal target, because the text has no "non-Scratch".
- *Spoil Forest Colossus*, in `animusDefs` (`pkg/engine/cards/animus.go`): two triggers with the same effect, one on `HookDeploy` and one on `HookAttacks`. The effect is one target `{}` (any allegiance, any asset) and the step `Heal 2` to slot 0.
- *Ironclad Colossus*, in `pathDefs`: like Spoil Forest Colossus, but the target is `{Kinds: Unit}` (any allegiance) and the step is `AddArmor 2` to slot 0.
- Each `Text` is the quoted line exactly. A card with two triggers still has one `Text` line.
- No new types, fields or logic. If a card seems to need one, set `blocked`.

### What the tests can observe (for rules-test-author)
- **Guerrillas:**
  - Deployed to the Main Theater, it offers only opposing units (Scratch Companies included, Structures and HQs excluded) and deals 2.
  - Deployed to Reserves, nothing is asked and nothing happens.
  - With no opposing unit, it still deploys.
  - An opposing Ward unit can't be chosen.
  - A sole target is fast-pathed, and a timeout picks a legal one.
  - Lethal damage destroys the unit.
- **Spoil Forest Colossus:**
  - It enters Reserves as a 6/7.
  - On deploy, it offers every asset on both sides, itself and both HQs included, minus opposing Ward assets. The chosen asset loses up to 2 damage.
  - It heals again each time it attacks.
  - The attack heal resolves at Step 2. So a Colossus with 6 damage that attacks a 2-power defender and heals itself ends with 6 damage, and survives a retaliation that would otherwise kill it.
  - Healing an undamaged asset changes nothing.
- **Ironclad Colossus:**
  - It enters Reserves as a 6/7.
  - On deploy and on each attack, the chosen unit (allied or opposing, itself included; Structures and HQs excluded; opposing Ward units excluded) gains 2 Armor.
  - The grants stack, add to printed Armor, and persist past end of turn.
  - The attack grant resolves at Step 2. So when it chooses itself, the Armor absorbs that attack's retaliation 1-for-1 before health.
- **All three:** the cards are real (`s.card`) and deployed with `ActDeploy`, never placed directly.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
### Devilhorn Guerrillas (`pkg/engine/cards_path_test.go`, 11 tests)
- `TestCard_DevilhornGuerrillas_EntersDeclaredZoneAsThreeFive`: 3/5, costs 4R, enters whichever zone was declared.
- `..._MainTheaterOffersOnlyOpposingUnitsAndDeals2`: options are opposing units only (Scratch and Reserves units included; no Structures, HQs or allied units); the chosen one takes 2.
- `..._ScratchCompanyIsALegalTargetAndDies`: a sole Scratch Company is auto-selected and destroyed.
- `..._ReservesAsksNothingAndDamagesNothing` and `..._ReservesWithSoleOpposingUnitDoesNothing`: the zone gate (FR-TRN-003); no ask, no damage, even with a sole target.
- `..._NoOpposingUnitStillDeploys`: only an opposing Structure and HQ exist; no ask, it deploys.
- `..._OpposingWardUnitCannotBeChosen` and `..._OnlyOpposingUnitHasWardSoNothingHappens` (FR-KWD-002h).
- `..._SoleOpposingTargetIsNotAsked` (FR-MOD-001 sec. 1) and `..._TimeoutDamagesExactlyOneLegalUnit` (sec. 6).
- `..._LethalDamageDestroysTheUnit`.

### Ironclad Colossus (`pkg/engine/cards_path_test.go`, 11 tests)
- `EntersReservesAsSixSeven`; `DeployOffersAnyUnitButNotStructuresHeadquartersOrOpposingWard` (itself and Scratch included); `AlliedWardUnitIsAnOption`; `CanArmorAnOpposingUnit`.
- `OnlyItselfOnTheBoardIsGivenArmorWithoutAsking` and `OpposingWardIsTheOnlyOtherUnitSoOnlyItselfIsChosen` (self is the sole option, fast path).
- `GrantAddsToPrintedArmor` (printed 1 + 2 = 3), `GrantPersistsPastEndOfTurn`, `EachAttackGrantsAgainAndGrantsStack` (deploy + attack = 4, two questions).
- `AttackGrantResolvesBeforeCombatDamage` (Step 2: choosing itself, 3 retaliation leaves 1 damage, not 3), `DefendingDoesNotTrigger`.

### Spoil Forest Colossus (`pkg/engine/cards_animus_test.go`, 12 tests)
- `EntersReservesAsSixSeven`; `DeployOffersEveryAssetOfBothSidesButOpposingWard` (both HQs, Structures, Scratch, itself).
- `DeployHealsChosenAssetOf2`, `CanHealAnOpposingAsset`, `CanHealAHeadquarters`, `HealNeverGoesBelowZero`, `HealingAnUndamagedAssetChangesNothing`, `HealLeavesArmorAlone` (FR-KWD-002d).
- `OpposingWardUnitCannotBeChosenButAlliedCan`.
- `HealsAgainEachTimeItAttacks` (deploy plus two attacks, three questions).
- `AttackHealResolvesBeforeCombatDamage` (6 damage, heals itself at Step 2, takes 2, ends at 6 and survives), `DefendingDoesNotHeal`.

Attack tests deploy through `ActDeploy` into Reserves, then use `ActMarch` and two end-turns (helper `colossusCycle` in `cards_path_test.go`, shared with the animus file) so the Colossus attacks from the Main Theater, Readied.

Deliberately left out:
- Spoil Forest Colossus fast path: never reachable, because both HQs and the Colossus itself are always options.
- Timeout fallback for the two Colossi: not in the task's observable list, and the Guerrillas and Saboteur tests already pin FR-MOD-001 sec. 6.
- A chosen target that dies or leaves before resolution: covered by the T-0017 and T-0019 shape tests.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None. The tests use only existing helpers (`scriptRefs`, `targetBoard`, `putScratch`, `assetOf`, `wantOptions`) plus local helpers defined in the test files.

## Allowed Files
- `pkg/engine/cards/path.go`
- `pkg/engine/cards/animus.go`
- `pkg/engine/cards/cards.go` (remove the three names from `NotYetExecutable`)

## Done When
- [x] The cards' tests pass: `go test ./pkg/engine -run 'TestCard_(DevilhornGuerrillas|SpoilForestColossus|IroncladColossus)'`
- [x] `TestDefs_MatchRulesText` and `TestStarterSet_Executable` pass unchanged
- [x] `go tool task check` passes
- [x] `go run ./cmd/cardtown sim -deck starter -p0 ANIMUS -p1 PATH` exits 0. Every invariant holds and the rule monitors report no violation. The executable counts rise by 3 for Path (3 Guerrillas) and 2 for Animus (2 Spoil Forest Colossus). Report the counts and the `targeting` and `damage-sequence` coverage rows, before and after, in the Log.
- [x] `go run ./cmd/cardtown sim` exits 0. Report the before and after summary lines in the Log.

## Log
- 2026-10-21 lead-architect: Wrote the task (A2 card task 2), with status `tests` for `rules-test-author`. Starter-deck cards not yet executable before this task: 34. Expected after: 32.
- 2026-10-21 rules-test-author: Wrote 34 tests (11 Guerrillas, 11 Ironclad, 12 Spoil Forest); see Tests for the list and omissions. All fail today with `Deploy rejected: <card> needs rules the kernel does not implement yet`; every other engine test passes. gofmt, vet and `cardtown lint rules` are clean. Helpers requested: none. No open rule questions. The march-and-cycle setup in the attack tests is unrun until the cards deploy. Status set to `tests-review`.
- 2026-10-21 lead-architect: Test review: **approved**, status `ready`.
  - Scope: 33557b6 adds 34 `TestCard_*` tests and changes no existing test.
    - Guerrillas: 11 tests, and Ironclad Colossus: 11, both in `cards_path_test.go`.
    - Spoil Forest Colossus: 12, in `cards_animus_test.go`. It shares `deployPath`, `colossusCycle` and `wantOptions` from the Path file, which is fine within one package.
    - Real cards, `ActDeploy`, targets chosen by asset.
  - Rules: every expected outcome checks against the cited text.
    - `FR-TRN-003`: the gate fails in Reserves, and nothing is asked.
    - `FR-KWD-002d`: heal to a floor of 0, Armor untouched, any asset including HQs.
    - `FR-KWD-002k` and `FR-LAY-002`: grants stack with printed Armor and persist.
    - `FR-CBT-002`: the Step 2 attack trigger comes before combat damage, and defending doesn't trigger.
    - `FR-KWD-002h`: Ward blocks opponents only.
    - `FR-MOD-001` §1 and §6: the fast path and the timeout.
  - The omissions are accepted. Spoil Forest Colossus can't take the fast path, because both HQs are always options. The Colossus timeout duplicates the Guerrillas case. Loss at resolution is pinned by the shape tests.
  - **Satisfiability check by the reviewer** (a throwaway worktree, not committed). These definitions make all 34 pass:
    - Guerrillas: `HookDeploy` with `ToZone: ZoneMainTheater` and `{Opposing, Unit}`, `DealDamage 2`.
    - Spoil Forest Colossus: `HookDeploy` and `HookAttacks`, each `{}` with `Heal 2`.
    - Ironclad Colossus: the same two hooks, each `{Kinds: Unit}` with `AddArmor 2`.
  - **Mutations each fail the tests:**
    - no zone gate: 2 failures;
    - no attack triggers: 4;
    - armor on any asset: 6;
    - heal on units only: 2.
  - Next: **the user runs `@kernel-engineer T-0027` in the Antigravity agent (Gemini Flash)**.
- 2026-10-21 kernel-engineer: Implemented. Created ability definitions for "Devilhorn Guerrillas" (HookDeploy, ToZone: ZoneMainTheater, Target 0: Opposing Unit, Step: DealDamage 2 to slot 0) and "Ironclad Colossus" (HookDeploy and HookAttacks, Target 0: Kinds Unit, Step: AddArmor 2 to slot 0) in pkg/engine/cards/path.go; added "Spoil Forest Colossus" (HookDeploy and HookAttacks, Target 0: empty filter matching any asset, Step: Heal 2 to slot 0) in pkg/engine/cards/animus.go; removed all three from NotYetExecutable in pkg/engine/cards/cards.go while maintaining sort order. Reused HookDeploy, HookAttacks, ToZone: ZoneMainTheater, Effect.Targets, Filter, AllegianceOpposing, KindUnit, OpDealDamage, OpHeal, OpAddArmor, SubjectTarget, and Step.Target; added no new shapes or engine logic. Gates: go test ./pkg/engine -run 'TestCard_(DevilhornGuerrillas|SpoilForestColossus|IroncladColossus)' passed (34/34), TestDefs_MatchRulesText and TestStarterSet_Executable passed, go tool task check passed (fmt-check, vet, all tests, full lint --all). Starter deck sim (ANIMUS vs PATH) exited 0 with all invariants holding: P0 (Animus starter) executable cards rose from 19/50 to 21/50 (+2 Spoil Forest Colossus), P1 (Path starter) rose from 9/50 to 12/50 (+3 Devilhorn Guerrillas); targeting coverage rose from 170 (FR-TGT-001: 170, SVY-R3: 170) to 436 (FR-TGT-001: 436, SVY-R3: 436); damage-sequence coverage changed from (FR-CBT-002: 2037, FR-KWD-002b: 55, FR-KWD-002c: 49, FR-KWD-002k: 18, FR-REP-004: 4244) to (FR-CBT-002: 1913, FR-KWD-002b: 53, FR-KWD-002c: 46, FR-KWD-002k: 25, FR-REP-004: 4005). Default sim before and after: "SOULLESS (Colossus Platform) vs CONCORDIA (Enclave), 200 games: P0 wins 18.0% P1 wins 82.0% draws 0.0% first player wins 55.0%, Avg match length: 19.5 turns" (unchanged). Reviewer: check ToZone: ZoneMainTheater on Guerrillas and filter scope on the two Colossi.
- 2026-10-21 lead-architect: Code review: **approved**, status `done`.
  - Scope: 8e7025e touches only `cards/path.go`, `cards/animus.go` and `cards/cards.go`. No test changed after `test(T-0027)` 33557b6.
  - Fences: plain data with no new shapes. The definitions are exactly the ones checked at test review, and each `Text` is the CSV line exactly.
  - Gates re-run by the reviewer:
    - The 34 card tests, `TestDefs_MatchRulesText` and `TestStarterSet_Executable` pass, and `go tool task check` exits 0.
    - The starter sim (progress) exits 0, byte-identical across runs, with no monitor violation. Counts: Animus 21/50 (+2), Path 12/50 (+3). Result: P0 85.5%, P1 14.5%, draws 0.0%.
    - `targeting` checks go from 170 to 436.
    - New deploys in 200 games: Guerrillas 54 and Spoil Forest Colossus 147 under progress. Random deploys Guerrillas once and the Colossus never, so the random starter summary is unchanged at 51.0/17.0/32.0.
    - The default stress sim is unchanged at 18.0/82.0.
  - Starter-deck cards not yet executable: 32 (from 34).
