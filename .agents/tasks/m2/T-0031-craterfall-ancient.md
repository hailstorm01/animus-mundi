---
id: T-0031
title: Make Craterfall Ancient executable
kind: card
tier: flash
wave: A2 (task 8, card task 4; last of the wave)
status: in-progress
branch: m2/T-0031-craterfall-ancient
depends_on: [T-0028, T-0030]
claims: [card:Craterfall Ancient]
cards: [Craterfall Ancient]
requirements: []
---

## Goal
*Craterfall Ancient* (Animus, 2 copies in the starter deck) now runs in the kernel. It is plain `ability` data, using the shapes from T-0028 and T-0030. This is the last wave A2 card.

## Inputs
- Card (`data/cards/units-and-structures-animus.csv`): Animus Unit (Elemental), cost 10R 0T, deploys to Reserves, 6/13, yields 3R 3T, rarity Rare. It prints two lines, and both are unsupported today:
  - "Whenever this unit attacks, you may deal damage equal to its power to each of up to 2 other chosen assets."
  - "When this unit is killed, deal 4 damage to all assets."
- Survey row: `docs/design/m2_starter_set_ability_survey.md` line 97. Its note says "Death trigger can destroy both HQs (Match Draw)".
- Shapes it uses (all `done`):
  - `HookAttacks` (T-0010) and `HookKilled` (T-0008);
  - `Effect.Targets` and `OpDealDamage` (T-0017);
  - `ValueSourcePower` and `SubjectAll` (T-0028);
  - `Trigger.Optional`, `Target.UpTo` and `Filter.Other` (T-0030).
- Rulings and requirements for context:
  - `SVY-R9`: "you may" is asked on placement, before the targets. A declined trigger is not placed, and a timeout declines.
  - `SVY-R10`: "its power" is the effective power at resolution, or last-known if the source is gone.
  - `SVY-R11`: "up to N" targets are distinct.
  - `FR-TGT-002`: "all assets" is not a target, so Ward doesn't protect.
  - `FR-KWD-002h`: Ward.
  - `FR-CBT-002`:
    - Step 2 attack triggers resolve before Step 3 combat damage.
    - If the defender is destroyed in Step 2, the attack fizzles with no combat damage either way.
  - `FR-CBT-003`: both Headquarters destroyed in one sweep is a Match Draw.
  - `FR-ZON-004`: a face-down Scratch Company fires no printed trigger.
- Kernel conventions:
  - Tests choose targets by asset, and options are listed oldest first.
  - Since T-0028, damage from a dead source carries its asset ID.

### Definition (the engineer adds it to `animusDefs` in `pkg/engine/cards/animus.go`)
- Trigger 1:
  - `HookAttacks` with `Optional: true`.
  - One target `{Filter: {Other: true}, UpTo: 2}`: any allegiance, any asset kind, excluding itself.
  - Step: `DealDamage` with `Amount: Value{From: ValueSourcePower}` to `SubjectTarget` slot 0.
- Trigger 2:
  - `HookKilled`, with no targets.
  - Step: `DealDamage` with `Amount: Value{N: 4}` to `SubjectAll`, with an empty `Filter` (every asset).
- `Text` holds both quoted lines exactly, in printed order.
- No new types, fields or logic. If the card seems to need one, set `blocked`.
- **Reviewer pre-check** (a throwaway prototype of exactly this definition, not committed):
  - `go test ./...` passes, `TestProgress_Case4_Coverage` included: Craterfall Ancient gets 748 deploys in 1000 games.
  - The 1000-game starter sim passes with no rule-monitor violation, and its death trigger fired 45 times.

### What the tests can observe (for rules-test-author)
- It enters Reserves as a 6/13 for 10 Requisition.
- **Attack trigger:**
  - Each attack asks its controller "you may" first, before any target question. No question is asked on deploy.
  - **Yes:** up to 2 distinct assets, chosen among every asset of both sides except itself and opposing Ward assets. Both Headquarters are included, its controller's own HQ too. Each chosen asset takes one non-combat instance of damage equal to its power: 6.
  - Choosing 1 or 0 assets is allowed. With 0, nothing is damaged.
  - **No, or a timeout:** nothing happens and no target is asked.
  - The power is read at resolution: with +2 temporary power set in the scene, the instances are 8.
  - **Step 2 timing:** choosing the defender with enough damage to destroy it in Step 2 makes the attack fizzle. The Ancient takes no retaliation.
- **Death trigger:**
  - When it is destroyed, every asset on the battlefield takes one 4-damage instance: both sides, both Headquarters, opposing Ward assets, Resistant and Armor applied per instance.
  - No question is asked.
  - Assets with 4 or less health die in the same sweep.
  - If both Headquarters are at 4 or less current health, the match is a Draw. With only one, that player loses.
  - The damage events' `Source` is the dead Ancient's asset ID.
- **Both triggers in one combat:** it attacks, the player says yes and picks 2 assets, then lethal retaliation kills it. Its death trigger then deals 4 to every remaining asset as well.
- **Face-down:** a Craterfall Ancient card that dies as a face-down Scratch Company fires nothing.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
Written in `pkg/engine/cards_animus_test.go`. Scenes use the real card deployed through `ActDeploy` with 10 Requisition (`deployPath`), marched and readied by `colossusCycle`, around a board of ally, opposing units (20+ health, so nothing dies unless a test says so) and an opposing Ward unit. Targets are chosen by asset (`scriptPlan`, `pick`). 20 tests:
- `EntersReservesAsSixThirteenForTenRequisition`: 6/13, Reserves, 0 Requisition left, no question and no damage on deploy (kills "attack trigger also fires on deploy").
- `AttackAsksMayThenOffersEveryOtherAssetButOpposingWard`: decisions are Optional then Target, Min 0 Max 2 (kills exactly 1 target, not optional), options are both HQs and every asset except the Ancient and the opposing Ward unit (kills no `Other`).
- `YesDealsPowerToEachOfTwoAssetsOwnHeadquartersIncluded`: 6 to own HQ and an opposing unit, one instance each, source is the Ancient, nothing else damaged.
- `YesCanHitTheOpposingHeadquartersAndAnAlliedUnit`: opposing HQ and allied unit take 6.
- `YesWithOneAssetDealsOneInstance`, `YesWithZeroAssetsDamagesNothing` (SVY-R11).
- `NoDealsNothingAndAsksNoTarget`, `TimeoutOnTheQuestionDeclines` (SVY-R9).
- `DamageEqualsEffectivePowerAtResolution`: +2 `TempPower` gives 8 (SVY-R10, kills fixed 6).
- `AttackDamageResolvesBeforeCombatSoAKilledDefenderFizzlesTheAttack`: choosing a 5-power 6-health defender destroys it at Step 2 and the Ancient takes 0 (FR-CBT-002).
- `AsksAgainOnEachAttack`, `DefendingDoesNotTrigger`.
- `FaceDownCopyAttackingAsksNothing` (FR-ZON-004). Passes today, since a face-down card fires no printed trigger; it guards against regressions.
- `DeathDealsFourToEveryAssetOfBothSidesWithoutAQuestion`: both HQs, allied, opposing, the opposing Ward unit and the killer each take exactly one 4 instance, source is the dead Ancient, no question (kills "opposing only", "skips HQ", Ward protection).
- `DeathKillsAssetsWithFourOrLessHealthAndSparesTheRest`.
- `DeathAppliesResistantAndArmorPerInstance` (FR-REP-004): Resistant 2 takes 2; Armor 3 takes 1 and ends with 0 Armor.
- `DeathWithBothHeadquartersAtFourOrLessIsAMatchDraw`, `DeathWithOneHeadquartersAtFourOrLessLosesThatPlayer` (both players tried) (FR-CBT-003).
- `FaceDownCopyFiresNothingWhenKilled` (FR-ZON-004). Passes today, same reason.
- `AttackThenLethalRetaliationFiresBothTriggers`: yes, picks 2 assets, 13 retaliation kills it; they take 6 then 4, unchosen assets and the HQ take 4.

Deliberately left out:
- Repeated pick of one asset and timeout on the target question: the fallback is T-0030's engine-level behavior (`optional_upto_test.go`), not specific to this card.
- Losing the attack trigger when the Ancient dies mid-resolution, and the "last-known power" branch of SVY-R10 on the death trigger (the death trigger has a fixed 4).
- Text, yield and rarity checks: `TestDefs_MatchRulesText` covers the text.

Expected today: 18 tests fail on `Deploy rejected: ... needs rules the kernel does not implement yet`; the two face-down tests pass.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None. Every helper used already exists: `deployPath`, `colossusCycle`, `wantOptions`, `killOnOpponentsTurn`, `scriptPlan`, `pick`, `wantDealt`, `kindsOf`. The tests add `ancientScene` and `ancientKilled` in `cards_animus_test.go`.

## Allowed Files
- `pkg/engine/cards/animus.go`
- `pkg/engine/cards/cards.go` (remove the name from `NotYetExecutable`)

## Done When
- [ ] The card's tests pass: `go test ./pkg/engine -run 'TestCard_CraterfallAncient'`
- [ ] `TestDefs_MatchRulesText` and `TestStarterSet_Executable` pass unchanged
- [ ] `go tool task check` passes, `TestProgress_Case4_Coverage` included
- [ ] `go run ./cmd/cardtown sim -deck starter -p0 ANIMUS -p1 PATH` exits 0. Every invariant holds and the rule monitors report no violation. The Animus executable count rises by 2, from 21/50 to 23/50. Report the before and after summary lines in the Log.
- [ ] `go run ./cmd/cardtown sim` exits 0. Report the before and after summary lines in the Log.

## Log
- 2026-10-21 lead-architect: Wrote the task (A2 card task 4, last of the wave), with status `tests` for `rules-test-author`. Starter-deck cards not yet executable before this task: 31. Expected after: 30.
- 2026-10-21 rules-test-author: Wrote 20 `TestCard_CraterfallAncient_*` tests in `pkg/engine/cards_animus_test.go` (list and omissions in Tests). No helpers requested. gofmt, vet and `cardtown lint rules` are clean. 18 tests fail today with the kernel rejecting the deploy of a not-yet-executable card; the two face-down tests pass already. Every other engine test passes. No open questions. Status: `tests-review`.
- 2026-10-21 lead-architect: Test review: **approved**, status `ready`.
  - 29c1d48 adds 20 `TestCard_CraterfallAncient_*` tests plus the `ancientScene` and `ancientKilled` helpers, and changes no existing test.
  - Rules content checks out against FRD 01 (`FR-CBT-002` Step 2 timing and fizzle, `FR-CBT-003` draw, `FR-TGT-002`, `FR-KWD-002h`), `SVY-R9`/`R10`/`R11`, and the task's conventions. The omissions are accepted.
  - **Reviewer check** (the throwaway definition from the pre-check, not committed): all 20 pass, and so does `go test ./...`.
  - **Mutations each fail tests:**
    - not optional: 5;
    - exactly one target: 6;
    - no `Other`: 1;
    - a fixed 6: 1;
    - the death trigger on opposing assets only: 6;
    - the death trigger skipping HQs: 4;
    - the attack trigger on deploy instead: 18.
  - Next: **the user runs `@kernel-engineer T-0031` in the Antigravity agent (Gemini Flash)**.
