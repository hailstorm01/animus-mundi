---
id: T-0018
title: Make Saboteur and Chainbreaker Sappers executable
kind: card
tier: flash
wave: A2 (task 2, card task 1)
status: tests-review
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
- [ ] The cards' tests pass: `go test ./pkg/engine -run 'TestCard_(Saboteur|ChainbreakerSappers)'`
- [ ] `TestDefs_MatchRulesText` and `TestStarterSet_Executable` pass unchanged
- [ ] `go tool task check` passes
- [ ] `go run ./cmd/cardtown sim -deck starter -p0 ANIMUS -p1 PATH` exits 0 with every invariant holding, and the Path executable count rises by 3 (3 Saboteur), from 6/50 to 9/50
- [ ] `go run ./cmd/cardtown sim` exits 0. Report the before and after summary lines in the Log.

## Log
- 2026-10-20 lead-architect: Wrote the task (A2 card task 1), with status `tests` for `rules-test-author`.
- 2026-10-20 rules-test-author: Wrote 23 card tests in cards_path_test.go (see Tests for omissions). No helpers requested. All 23 fail today because the kernel rejects both deploys ("needs rules the kernel does not implement yet"), so no control test passes in isolation; every other engine test still passes. Status tests-review.
