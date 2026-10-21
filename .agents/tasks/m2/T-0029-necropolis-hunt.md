---
id: T-0029
title: Make Necropolis Hunt executable
kind: card
tier: flash
wave: A2 (task 6, card task 3)
status: done
branch: m2/T-0029-necropolis-hunt
depends_on: [T-0028]
claims: [card:Necropolis Hunt]
cards: [Necropolis Hunt]
requirements: []
---

## Goal
*Necropolis Hunt* (Path, 1 copy in the starter deck) now runs in the kernel. It is plain `ability` data, using the source-power value that T-0028 added.

## Inputs
- Card (`data/cards/units-and-structures-path.csv`):
  - *Necropolis Hunt*: Path Unit, subtypes Reaver and Cavalry, costs 3R 3T, deploys to the Main Theater only, is 5/2, yields 1R 2T, rarity Mythic.
  - Its one unsupported line: "When this unit is killed, deal damage equal to its power to chosen non-Headquarters asset."
- Survey row: `docs/design/m2_starter_set_ability_survey.md` line 134.
- Shapes it uses (all `done`):
  - `HookKilled` (T-0008);
  - `Effect.Targets`, `Filter{Kinds}`, `SubjectTarget` and `OpDealDamage` (T-0017);
  - `Value{From: ValueSourcePower}` (T-0028).
- Requirements and rulings for context:
  - `FR-ZON-004`: a death trigger reads the last-known characteristics. A face-down Scratch Company fires no printed death trigger.
  - `SVY-R10`: "its power" of a source that has left is its last-known effective power.
  - `FR-TGT-001`, `FR-TGT-003` and `FR-KWD-002h`: targeting, verification and Ward.
  - `FR-MOD-001` §1 and §6: the fast path and the timeout.
  - `FR-TRG-003`: death triggers are placed after the sweep, APNAP.
  - `FR-REP-004`: non-combat damage skips Stalwart.
- Kernel conventions:
  - Targets are chosen when the death trigger is placed, after the sweep, so the dead unit is never an option.
  - Options are listed oldest first.
  - Tests choose by asset (`scriptRefs`).
  - Since T-0028, a damage event from a dead source carries that source's asset ID.

### Definition (the engineer adds it to `pathDefs` in `pkg/engine/cards/path.go`)
- *Necropolis Hunt*: one `HookKilled` trigger with no turn gate. It has one target, `{Kinds: KindUnit | KindStructure}` (any allegiance, so "non-Headquarters asset" means a unit or a structure of either side), and the step `DealDamage` with `Amount: Value{From: ValueSourcePower}` to slot 0.
- The `Text` is the quoted line exactly. No new types, fields or logic. If the card seems to need one, set `blocked`.

### What the tests can observe (for rules-test-author)
- It deploys to the Main Theater as a 5/2 for 3R 3T.
- **Killed while attacking** (lethal retaliation):
  - Its controller chooses among every unit and structure of both sides. Both Headquarters, the dead Hunt and opposing Ward assets are excluded, and Scratch Companies are included.
  - The chosen asset takes 5 non-combat damage. The damage event's `Source` is the Hunt's asset ID.
- **Killed on the opponent's turn** (it defends, or takes Saboteur-style damage): it still triggers, because there is no turn gate, and still deals 5. Its controller chooses.
- **Damage rules:** the damage is non-combat, so a Stalwart target takes the full 5, while Resistant and Armor still apply. 5 damage destroys a unit with 5 health or less.
- **Edge cases:**
  - With no other unit or structure on the battlefield, nothing is asked, and nothing happens.
  - A sole legal target is not asked about.
  - A timeout damages exactly one legal target.
- **Face-down:** a Necropolis Hunt card that is a face-down Scratch Company when it dies fires nothing (`FR-ZON-004`).
- **No Detail text:** no event `Detail` contains `%!`.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
In `pkg/engine/cards_path_test.go` (local helpers `deployHunt`, `huntAttacks`, `huntDefends`):
- `TestCard_NecropolisHunt_EntersMainTheaterAsFiveTwoForThreeRequisitionThreeTactics`
- `TestCard_NecropolisHunt_KilledAttackingOffersEveryUnitAndStructureButNotHeadquartersItselfOrOpposingWard` (options, Scratch included, both HQs, dead Hunt and opposing Ward assets excluded, chosen asset takes 5)
- `TestCard_NecropolisHunt_DamageEventNamesTheDeadHuntAsSourceAndIsNonCombat`
- `TestCard_NecropolisHunt_KilledOnOpponentsTurnStillTriggersAndItsControllerChooses`
- `TestCard_NecropolisHunt_StalwartTargetTakesTheFullFive`
- `TestCard_NecropolisHunt_ResistantTargetReducesTheFive` (Resistant 2: 3 damage)
- `TestCard_NecropolisHunt_ArmorAbsorbsPartOfTheFive` (Armor 2: 3 damage, Armor spent)
- `TestCard_NecropolisHunt_FiveDamageDestroysAFiveHealthUnit`
- `TestCard_NecropolisHunt_NoOtherUnitOrStructureAsksNothingAndDoesNothing`
- `TestCard_NecropolisHunt_SoleLegalTargetIsNotAsked`
- `TestCard_NecropolisHunt_TimeoutDamagesExactlyOneLegalTarget`
- `TestCard_NecropolisHunt_FaceDownCopyFiresNothingWhenKilled` (scratched from the deck top, readied, marched, killed; passes today)
- `TestCard_NecropolisHunt_NoEventDetailHasAFormatError`
- `TestCard_NecropolisHunt_DealsItsLastKnownEffectivePowerNotAFixedFive` (TempPower 2 on the Hunt, dies the same turn, deals 7)

Omitted: option ordering ("oldest first") is not pinned, since FRD 01 does not require it. No FR-* requirement is listed, so there are no TestFR tests.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None. Tactics cost is covered by a local `deployHunt` helper in `cards_path_test.go`.

## Allowed Files
- `pkg/engine/cards/path.go`
- `pkg/engine/cards/cards.go` (remove the name from `NotYetExecutable`)
- `pkg/engine/sim/policy_test.go` (only the authorized edit below, by `rules-test-author` in a `test(T-0029)` commit)

## Done When
- [x] The card's tests pass: `go test ./pkg/engine -run 'TestCard_NecropolisHunt'`
- [x] `TestDefs_MatchRulesText` and `TestStarterSet_Executable` pass unchanged
- [x] `go tool task check` passes
- [x] `go run ./cmd/cardtown sim -deck starter -p0 ANIMUS -p1 PATH` exits 0. Every invariant holds and the rule monitors report no violation. The Path executable count rises by 1, from 12/50 to 13/50. Report the number of `Necropolis Hunt` deploys in 200 games (from `-events`) in the Log.
- [x] `go run ./cmd/cardtown sim` exits 0. Report the before and after summary lines in the Log.

## Log
- 2026-10-21 lead-architect: Wrote the task (A2 card task 3), with status `tests` for `rules-test-author`. Starter-deck cards not yet executable before this task: 32. Expected after: 31.
- 2026-10-21 rules-test-author: Wrote 13 `TestCard_NecropolisHunt_*` tests in `cards_path_test.go` (list under Tests). No shared helpers requested. Expected red: 12 fail at Deploy ("needs rules the kernel does not implement yet"); the face-down test passes today because it needs no definition. All other engine tests pass; gofmt, vet and `cardtown lint rules` are clean. Omitted: option-order assertions (not an FRD 01 rule). No open questions.
- 2026-10-21 lead-architect: Test review: **changes**, status `tests`. The reviewer ran a throwaway definition (`HookKilled`, `{Kinds: Unit|Structure}`, `DealDamage` with `ValueSourcePower`), not committed: 12 of 13 pass. Mutations "units only" and "any asset including HQs" are each caught by 3 to 4 tests.
  1. `TestCard_NecropolisHunt_KilledAttackingOffersEveryUnitAndStructureButNotHeadquartersItselfOrOpposingWard` expects the Defender as an option. But the Hunt's 5 combat damage kills the 2/5 Defender in the same sweep, so it isn't on the battlefield when the death trigger is placed. The kernel offers 7 options and the test wants 8. Drop `def` from the expected options, and assert that it's excluded because it died with the Hunt (`FR-TGT-001`: only battlefield assets). Other tests rely on the Defender dying, so keep `huntAttacks` as it is.
  2. No test tells "its power" apart from a fixed 5: a definition with `Value{N: 5}` passes all 13. Add one test where the Hunt dies with 2 temporary power (set `TempPower = 2` on the Hunt in the scene, standing in for an "until end of turn" modifier, as other tests set `Damage` directly). Its death trigger then deals 7: last-known *effective* power (`FR-ZON-004`, `SVY-R10`).
- 2026-10-21 rules-test-author: Answered review notes. Note 1: the options test no longer expects the Defender; it now asserts the Defender is not an option because it died with the Hunt; `huntAttacks` is unchanged. Note 2: added `TestCard_NecropolisHunt_DealsItsLastKnownEffectivePowerNotAFixedFive` (Hunt has `TempPower = 2` when an opposing Hitter kills it the same turn; the chosen unit takes 7). Status back to `tests-review`.
- 2026-10-21 lead-architect: Test review, round 2: **approved**, status `ready`.
  - e9b8959 answers both notes:
    - The options test now expects 7 options and asserts that the Defender, which died in the same sweep, is excluded.
    - New `TestCard_NecropolisHunt_DealsItsLastKnownEffectivePowerNotAFixedFive`: the Hunt dies with +2 temporary power and deals 7.
  - Reviewer re-check with the throwaway definition (not committed): all 14 pass, and so does `go test ./pkg/engine`. A fixed `Value{N: 5}` now fails the new test.
  - Next: **the user runs `@kernel-engineer T-0029` in the Antigravity agent (Gemini Flash)**.
- 2026-10-21 kernel-engineer: BLOCKED on `go tool task check` (TestProgress_Case4_Coverage). Implemented Necropolis Hunt in `pkg/engine/cards/path.go` (HookKilled, Target: Kinds Unit|Structure, Step: OpDealDamage with Amount: Value{From: ValueSourcePower} to slot 0) and removed from NotYetExecutable in `pkg/engine/cards/cards.go`. Reused HookKilled, Effect.Targets, Filter, Kinds (Unit|Structure), OpDealDamage, ValueSourcePower, SubjectTarget, and Step.Target; added no new shapes or engine logic. Tests: `go test ./pkg/engine -run 'TestCard_NecropolisHunt'` passes (14/14); `TestDefs_MatchRulesText` and `TestStarterSet_Executable` pass; `pkg/engine` passes all tests. Simulations: `go run ./cmd/cardtown sim -deck starter -p0 ANIMUS -p1 PATH` exits 0 with all invariants holding; Path executable cards rose from 12/50 to 13/50 (+1 Necropolis Hunt); Necropolis Hunt deployed 3 times in 200 games (verified via `-events`). Default sim exits 0 (before and after: P0 18.0%, P1 82.0%, draws 0.0%, first player 55.0%, avg 19.5 turns). Block: `go tool task check` runs `go test ./...` which fails at `pkg/engine/sim/policy_test.go:141` (`TestProgress_Case4_Coverage`): "Necropolis Hunt deployed 3 times in 200 games, want at least 10". Necropolis Hunt is a 1-of card in path-starter.txt (1 copy in 50 cards) with cost 3R 3T, unlike previous 2- or 3-copy cards. Fixing the test threshold or policy requires modifying `pkg/engine/sim/policy_test.go` or `policy.go`, which are outside Allowed Files.
- 2026-10-21 lead-architect: The block is **not a policy or engine fault**. Necropolis Hunt is a 1-copy card costing 3R and 3T, so it is just rare.
  - Measured with the definition in place: 45 deploys in 1000 starter games (about 4.5% of games), so about 9 are expected in 200. Seeds 1 to 200 happen to give 3, and the same seeds with the seats swapped give 11. A 200-game minimum of 10 is a coin flip for this card.
  - In the same 1000 games, the next-rarest card is Hammerhand Brigade with 148.
  - **Lead decision (harness parameter, recorded on directive #21):** the coverage check runs **1000 games** instead of 200, and the minimum stays at 10. A 1000-game run takes about 5 s with `-events` written.
  - **Authorized test edit:** in `pkg/engine/sim/policy_test.go`, `TestProgress_Case4_Coverage`, change the `SimulateWith` game count from 200 to 1000 and the failure message's "in 200 games" to "in 1000 games". Nothing else in that test or file changes. `rules-test-author` makes the edit in a new `test(T-0029): …` commit on top of a0ece3a. The engineer's definition (a0ece3a) is already reviewed against the test-review prototype and needs no change.
  - Then the lead re-runs the gates and reviews the task.
- 2026-10-21 rules-test-author: Made the lead-authorized edit (directive #21a) to `TestProgress_Case4_Coverage` in `pkg/engine/sim/policy_test.go`: `SimulateWith` game count 200 to 1000, and the failure message "in 200 games" to "in 1000 games". The minimum stays at 10; nothing else changed. The test passes. Deploys in 1000 games: Chainbreaker Sneak 1730, Despoiler Larva 1898, Despoiler Worm 344, Devilhorn Guerrillas 319, Hammerhand Brigade 148, Necropolis Hunt 45 (lowest), Portent Scrivener 1903, Saboteur 424, Spoil Forest Amalgam 410, Spoil Forest Colossus 774, Spoil Forest Creeper 1662, Stormfront of Ideals 822, Unceasing Tide 1781. `go tool task check` passes (exit 0). Status set to `review`.
- 2026-10-21 lead-architect: Code review: **approved**, status `done`.
  - Scope:
    - a0ece3a touches only `cards/path.go` and `cards/cards.go`. Its definition is exactly the one verified at test review: `HookKilled`, `{Kinds: Unit|Structure}`, `DealDamage` with `ValueSourcePower`, and `Text` equal to the CSV line.
    - No test changed after e9b8959 except the authorized 63acd4f: two lines in `TestProgress_Case4_Coverage`, 200 to 1000 games, in the call and the message.
  - Gates re-run by the reviewer:
    - The 14 card tests, `TestDefs_MatchRulesText` and `TestStarterSet_Executable` pass, and `go tool task check` exits 0.
    - The starter sim exits 0, byte-identical across runs, with no monitor violation. Path is at 13/50 (+1). Result: P0 88.0%, P1 12.0%.
    - The stress sim is unchanged at 18.0/82.0.
    - Coverage over 1000 games: Necropolis Hunt 45, which is the lowest.
  - Starter-deck cards not yet executable: 31.
