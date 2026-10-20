---
id: T-0018
title: Make Saboteur and Chainbreaker Sappers executable
kind: card
tier: flash
wave: A2 (task 2, card task 1)
status: done
branch: m2/T-0018-saboteur-sappers
depends_on: [T-0017]
claims: [card:Saboteur, card:Chainbreaker Sappers, file:pkg/engine/cards/path.go]
cards: [Saboteur, Chainbreaker Sappers]
requirements: []
---

## Goal
Two Path units with targeted deploy damage now run in the kernel. Each is plain `ability` data in a new `pathDefs` map in `pkg/engine/cards/path.go`, built the same way as `animusDefs` in `animus.go`. *Saboteur* is a starter-deck card (3 copies). *Chainbreaker Sappers* isn't in a deck, but every shape it needs exists, so it joins this task (board directive #14c).

## Inputs
- Cards (`data/cards/units-and-structures-path.csv`). Each has exactly one unsupported line today, quoted here:
  - *Saboteur*: Unit, 3R 0T, Reserves, 3/3. "When this unit deploys, deal 2 damage to chosen opposing non-Scratch unit and 1 damage to chosen allied unit."
  - *Chainbreaker Sappers*: Unit, 2R 0T, Reserves, 3/2. "When this unit deploys, deal 1 damage to chosen structure or Headquarters."
- Survey rows: `docs/design/m2_starter_set_ability_survey.md` (*Saboteur*, named in FRD 01 under FR-TGT-003; *Chainbreaker Sappers*, FR-TGT-001).
- Shapes they use (all `done`): `HookDeploy` (T-0007), and `Effect.Targets`, `Filter{Allegiance, Kinds, NonScratch}`, `SubjectTarget`, `Step.Target` and `OpDealDamage` (T-0017).
- Requirements for context:
  - `FR-TGT-001`: conjunction filtering.
  - `FR-TGT-003`: two-stage verification. "Deployment Triggers Without Legal Targets" names *Saboteur*: it still enters the battlefield, and only the clause with no target fizzles.
  - `FR-KWD-002h`: Ward.
  - `FR-MOD-001` §1 and §6: the fast path and the random fallback.
- Kernel conventions from T-0017: targets are chosen when the trigger is placed, in slot order. Options are listed oldest asset first, and a filter without "other" includes the source itself. Tests choose by asset (`scriptRefs` in `targeting_test.go`), never by option index.
- Rulings that apply: none.

### Definitions (the engineer adds them to `pathDefs`)
- *Saboteur*: one `HookDeploy` trigger with two targets, slot 0 `{Allegiance: Opposing, Kinds: Unit, NonScratch: true}` and slot 1 `{Allegiance: Allied, Kinds: Unit}`. Its steps are `DealDamage 2` to slot 0, then `DealDamage 1` to slot 1.
- *Chainbreaker Sappers*: one `HookDeploy` trigger with one target `{Kinds: Structure|Headquarters}` (any allegiance), and step `DealDamage 1` to slot 0.
- Each `Text` is the quoted line exactly. No new types, fields or logic. If a card seems to need one, set `blocked`.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
All in `pkg/engine/cards_path_test.go`: real cards via `s.card`, deployed with `ActDeploy`, targets chosen by asset (`scriptRefs`), options read from recorded Decisions. Test-file helpers: `deployPath`, `deploySaboteur`, `deploySappers`, `wantOptions`.
- Saboteur: `EntersReservesAsThreeThree`, `BothClausesHitTheirOwnChosenTarget`, `AsksOpposingClauseFirstThenAlliedClause`, `ScratchCompanyIsNotAnOpposingOptionButIsAnAlliedOne`, `StructuresAndHeadquartersAreNotOptions`, `ItselfIsALegalAlliedTarget`, `TestFR_TGT_003_SaboteurWithNoOpposingNonScratchUnitStillDeploysAndAlliedClauseResolves`, `NoOpposingUnitStillResolvesAlliedClauseOnChosenUnit`, `OpposingWardUnitCannotBeChosen`, `OnlyOpposingUnitHasWardSoOpposingClauseDoesNothing`, `AlliedWardUnitIsAnAlliedOption`, `LethalDamageDestroysBothChosenUnits`, `SoleOpposingTargetIsNotAsked`, `TimeoutOnOpposingChoiceDamagesExactlyOneLegalUnit`.
- Chainbreaker Sappers: `EntersReservesAsThreeTwo`, `OptionsAreStructuresAndHeadquartersOfEitherSideOnly`, `CanDamageEachKindOfLegalTarget` (4 subtests), `OpposingWardStructureCannotBeChosenButAlliedCan`, `LethalDamageDestroysAOneHealthStructure`, `DamagesHeadquartersWithoutEndingTheMatch`, `DamageToOpposingHeadquartersToZeroWinsTheMatch`, `TimeoutDamagesExactlyOneLegalTarget`, `SoleLegalTargetIsNotAsked` (gives the opposing HQ card Ward).
- Left out: Saboteur's allied clause having no legal target (the Saboteur is always an allied unit, so it cannot occur at placement); a target lost before resolution (needs another effect; resolution-time verification is already pinned by `TestFR_TGT_003_*` in targeting_test.go); Armor/Resistant/Stalwart mitigation (pinned generically in targeting_test.go).

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None. The tests use only existing helpers.

## Allowed Files
- `pkg/engine/cards/path.go` (new)
- `pkg/engine/cards/cards.go` (remove the two names from `NotYetExecutable`)

## Done When
- [x] The cards' tests pass: `go test ./pkg/engine -run 'TestCard_(Saboteur|ChainbreakerSappers)|TestFR_TGT_003_Saboteur'`
- [x] `TestDefs_MatchRulesText` and `TestStarterSet_Executable` pass unchanged
- [x] `go tool task check` passes
- [x] `go run ./cmd/cardtown sim -deck starter -p0 ANIMUS -p1 PATH` exits 0 with every invariant holding, and the Path executable count rises by 3 (3 Saboteur), from 6/50 to 9/50
- [x] `go run ./cmd/cardtown sim` exits 0. Report the before and after summary lines in the Log.

## Log
- 2026-10-20 lead-architect: Wrote the task (A2 card task 1), with status `tests` for `rules-test-author`.
- 2026-10-20 rules-test-author: Wrote 23 card tests in cards_path_test.go (see Tests for omissions). No helpers requested. All 23 fail today because the kernel rejects both deploys ("needs rules the kernel does not implement yet"), so no control test passes in isolation; every other engine test still passes. Status tests-review.
- 2026-10-20 lead-architect: Test review: **approved**, status `ready`. 9fa2050 adds `cards_path_test.go`, with 23 tests (one has 4 subtests), and changes no existing test.
  - Written from the CSV text with the real cards, deployed through `ActDeploy`, with targets chosen by asset.
  - Saboteur: both clauses, slot order, Scratch Companies excluded only from the opposing clause, units only, and itself as a legal allied target. FR-TGT-003's named case (no opposing target: it still deploys and the allied clause resolves) is covered, as are Ward on both sides, lethal damage, the fast path and the timeout.
  - Sappers: structures and Headquarters of either side, and each kind damaged. Ward on both sides, lethal damage, the HQ to 0 winning the match, the timeout and the fast path are covered. Ward on an HQ follows FR-KWD-002h's "asset", and each scene builds its own HQ card, so the mutation doesn't leak.
  - Accepted omissions: the allied clause can't lack a target, because the Saboteur is always one. Loss at resolution and mitigation are pinned generically in `targeting_test.go`.
  - Lead fix: the Done When regex now also matches `TestFR_TGT_003_Saboteur…`.
  - Next: **the user runs `@kernel-engineer T-0018` in the Antigravity agent (Gemini Flash)**.
- 2026-10-20 kernel-engineer: Implemented. Created pkg/engine/cards/path.go with pathDefs registering "Chainbreaker Sappers" (HookDeploy, Target: Kinds Structure|Headquarters, Step: OpDealDamage 1 to SubjectTarget 0) and "Saboteur" (HookDeploy, Target 0: AllegianceOpposing, KindUnit, NonScratch; Target 1: AllegianceAllied, KindUnit; Steps: OpDealDamage 2 to SubjectTarget 0, OpDealDamage 1 to SubjectTarget 1). pkg/engine/cards/cards.go: removed "Chainbreaker Sappers" and "Saboteur" from NotYetExecutable while maintaining sort order. Reused HookDeploy, Effect.Targets, Filter, AllegianceOpposing/Allied, KindUnit/Structure/Headquarters, OpDealDamage, SubjectTarget, and Step.Target; added no new shapes, types, fields, or logic. Gates: go test ./pkg/engine -run 'TestCard_(Saboteur|ChainbreakerSappers)|TestFR_TGT_003_Saboteur' passed (23/23 tests, 4 subtests), TestDefs_MatchRulesText and TestStarterSet_Executable passed, go tool task check passed (gofmt, vet, all tests, cardtown lint --all). Starter deck sim (ANIMUS vs PATH) exited 0 with all invariants holding; P1 (Path starter) main cards executable rose by 3 from 6/50 to 9/50 (3 Saboteur). Default sim (SOULLESS vs CONCORDIA) exited 0 with all invariants holding; before: "P0 wins 66.0% P1 wins 30.0% draws 4.0% first player wins 47.0%, Avg match length: 30.2 turns", after: "P0 wins 66.0% P1 wins 30.0% draws 4.0% first player wins 47.0%, Avg match length: 30.2 turns" (unchanged). Reviewer: check target definitions in path.go for slot order and conjunction filter flags.
- 2026-10-20 lead-architect: Code review: **approved**, status `done`.
  - Scope: e398f57 creates `cards/path.go` (`pathDefs`, registered like `animusDefs`) and removes the two names from `NotYetExecutable`. No test changed after `test(T-0018)` 9fa2050.
  - Fences: plain data with no new shapes. Saboteur's slot 0 is `{Opposing, Unit, NonScratch}` with `DealDamage 2`, and slot 1 is `{Allied, Unit}` with `DealDamage 1`. Sappers' target is `{Structure|Headquarters}` with `DealDamage 1`. Each `Text` is the CSV line exactly.
  - Gates re-run by the reviewer: 23 card tests, `TestDefs_MatchRulesText` and `TestStarterSet_Executable` pass, and `go tool task check` exits 0.
  - Starter sim exits 0. Path executable count goes from 6/50 to 9/50 (+3). Result: P0 51.0%, P1 17.0%, draws 32.0%, all invariants hold, so Saboteur's targeting runs clean under the random chooser. The default sim also exits 0.
