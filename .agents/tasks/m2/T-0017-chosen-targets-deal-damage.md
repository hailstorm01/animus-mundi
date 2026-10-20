---
id: T-0017
title: Chosen targets with conjunction filters, two-stage verification, and DealDamage
kind: shape
tier: sonnet
wave: A2 (task 1, the targeting core)
status: review
branch: m2/T-0017-chosen-targets-deal-damage
depends_on: [T-0010]
claims: [field:Effect.Targets, type:ability.Target, type:ability.Filter, type:ability.Allegiance, type:ability.Kinds, subject:target, field:Step.Target, op:DealDamage, func:choose-targets-at-placement, func:target-reverification]
cards: []
requirements: [FR-TGT-001, FR-TGT-003, FR-KWD-002h, FR-TRG-004]
---

## Goal
A triggered ability can say "deal N damage to chosen <filter>". When the trigger is placed on the stack, its controller chooses each target. The choice goes through the existing `Chooser` (`DecideTarget`), so the single-option fast path and the random timeout fallback come from T-0006. Each target is checked again at resolution, and a target that became illegal is skipped. A clause with no legal target does nothing, while the rest of the effect still resolves. After this task, *Saboteur* and *Chainbreaker Sappers* can be written as card data.

## Inputs
- FRD 01 (`docs/frd/01_Game_Rules_FRD.md`):
  - `FR-TGT-001` (lines 280–291): conjunction filtering. Only the allegiance, entity-type and "non-Scratch" criteria are in scope here.
  - `FR-TGT-003` (lines 297–302): two-stage verification, partial invalidation, composite clauses, and deployment triggers without legal targets (it names *Saboteur*).
  - `FR-KWD-002h` (line 495): Ward. Opponents can't choose it for triggered abilities.
  - `FR-KWD-002b`, `FR-KWD-002c` and `FR-KWD-002k`: Resistant, Stalwart (combat only) and Armor, as they apply to non-combat damage (`FR-REP-004`).
  - `FR-MOD-001` §1 and §6: the fast path and the random fallback. These already exist at the `Decision` level (T-0006).
- Card text for context. The card task that follows uses these, and this task's tests use synthetic cards:
  - *Saboteur*: "When this unit deploys, deal 2 damage to chosen opposing non-Scratch unit and 1 damage to chosen allied unit."
  - *Chainbreaker Sappers*: "When this unit deploys, deal 1 damage to chosen structure or Headquarters."
- Built on: `choice.go` (`Decision`, `DecideTarget`, `Game.choose`), `trigger.go` (`pendingTrigger`, `flushTriggers`, `resolveAbility`) and `damage.go` (`dealDamage(source, target, raw, combat)`).
- Rulings that apply: none.

### Vocabulary (`pkg/engine/ability/ability.go`)
```go
// Effect gains:
	Targets []Target // "chosen" clauses, chosen when the ability is placed on the stack (FR-TGT-001)

// Target is one "chosen" entity: exactly one, for now ("up to N" comes later).
type Target struct {
	Filter Filter
}

// Filter is a conjunction: an entity must meet every set criterion (FR-TGT-001).
type Filter struct {
	Allegiance Allegiance // relative to the ability's controller
	Kinds      Kinds      // zero means any asset
	NonScratch bool       // excludes Scratch Company units
}

type Allegiance int

const (
	AllegianceAny      Allegiance = iota // the zero value
	AllegianceAllied                     // controlled by the ability's controller
	AllegianceOpposing                   // controlled by an opponent
)

// Kinds is a set of asset types; "structure or Headquarters" sets two bits.
type Kinds int

const (
	KindUnit Kinds = 1 << iota
	KindStructure
	KindHeadquarters
)

const ( // Subject gains:
	SubjectTarget // the entity chosen for Effect.Targets[Step.Target]
)

const ( // Op gains:
	OpDealDamage // Amount non-combat damage to the Subject (FR-REP-004)
)

// Step gains:
	Target int // SubjectTarget: index into Effect.Targets
```
Names may be adjusted for clarity if needed. Record any rename in the Log.

### Lead decisions (technical, not rules)
- **Legal entity:** an Asset on the battlefield that meets the filter. It is not Warded against the chooser: an asset with Ward is illegal only when its controller is not the ability's controller (`FR-KWD-002h`). A face-down Scratch Company is a Unit with the subtype Scratch Company, so `KindUnit` includes it and `NonScratch` excludes it. Without an exclusion, the source itself is legal ("chosen allied unit" may pick the source). "Other" comes in a later task.
- **When and who:** targets are chosen as each trigger is pushed in `flushTriggers`, after that player's `DecideOrder`, in push order, one `DecideTarget` per `Effect.Targets` entry, in slot order. The trigger's controller chooses. Options are only the legal entities, so a single option is fast-pathed and a timeout picks a random legal one.
- **No legal target at placement:** the trigger is still placed and still resolves. Steps on that slot do nothing, and other steps run. Deploying never requires a target (`FR-TGT-003`).
- **At resolution:** before the steps run, check each chosen entity again: still on the battlefield, still meeting the filter, and not Warded against the controller. An illegal slot is treated like an empty one. Store the chosen asset IDs on the `StackItem`.
- **DealDamage:** `dealDamage(source, target, Amount.N, false)`, so Resistant applies, Stalwart does not, and Armor absorbs. Deaths and HQ defeat are settled by the sweep after the item, as today. `source` is the ability's source asset if it is still on the battlefield, otherwise nil. Today `dealDamage` uses the source only for logging, so make sure nil is safe.
- **Not in this task:** "up to N", "other", zone filters, stat filters, optional triggers, `Value` sources other than a fixed N, distributed pools, resolution-time targets (`SVY-R3`/`SVY-R4`), Maneuvers and activated abilities.

### What the tests can observe
- **Filter:** on a board with one entity of each kind for each player, the `DecideTarget` options are exactly the legal entities, read from the scripted chooser's recorded `Decision`. Check each criterion and their conjunction: allied, opposing, unit, structure, Headquarters, "structure or Headquarters", and non-Scratch.
- **Ward:** an opposing Ward unit is not an option. An allied Ward unit is.
- **Fast path and fallback:** one legal option means no question is asked, and the damage lands. A timeout on two options damages exactly one of them.
- **Damage:** non-combat damage goes through Resistant and Armor. Stalwart doesn't reduce it. A unit with lethal damage dies at that item's sweep. An HQ taken to 0 loses the match at that sweep.
- **Two targets in one effect:** two slots with different filters ask two questions in slot order, and each step damages its own slot.
- **No legal target:** a "deal damage to chosen opposing unit" trigger with no opposing unit asks nothing and damages nothing. A second, untargeted step in the same effect (for example `Draw 1`) still happens, and the deployed unit stays on the battlefield.
- **Re-verification:** an entity chosen at placement that is gone at resolution takes no damage, and a composite effect's other steps still run. For example, a source with two deploy triggers both choose the same 1-health opposing unit. The first to resolve kills it at its sweep, so the second's damage step does nothing but its `Draw 1` still happens. A slot whose target is still legal resolves normally next to an illegal one (partial invalidation).
- **Deep cascade (board directive #11, `FR-TRG-004`), a required named case:** a trigger's damage kills a unit that has its own death trigger. That death trigger must resolve completely before the next item deeper in the stack. For example, the source has two deploy triggers, X above Y. X deals lethal damage to a chosen allied 1-health unit whose death trigger draws a card, and Y draws a card. The death trigger's draw takes the top card, and Y's draw takes the second. Name it `TestFR_TRG_004_*`.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
All in `pkg/engine/targeting_test.go`. Run: `go test ./pkg/engine -run 'TestFR_TGT_001|TestFR_TGT_003|TestFR_KWD_002[bchk]_|TestFR_TRG_004_DeathTriggerFromTargetedDamage'`
- `TestFR_TGT_001_OptionsAreExactlyTheEntitiesMeetingEveryCriterion` (15 subtests: allied, opposing, any, unit, structure, HQ, "structure or HQ", no kind, non-Scratch, conjunctions; single-option rows check that no question is asked and that the damage lands)
- `TestFR_TGT_001_AllegianceIsRelativeToTheTriggersController`
- `TestFR_TGT_001_SingleLegalTargetIsNotAskedAndTakesTheDamage`
- `TestFR_TGT_001_TimeoutOnTwoOptionsDamagesExactlyOneOfThem`
- `TestFR_TGT_001_LethalDamageKillsAtThatItemsSweep`
- `TestFR_TGT_001_DamageToHeadquartersToZeroLosesTheMatchAtThatSweep`
- `TestFR_TGT_001_TwoSlotsAskInSlotOrderAndEachStepHitsItsOwnSlot`
- `TestFR_KWD_002h_OpposingWardUnitIsNotAnOption`, `..._AlliedWardUnitIsAnOption`, `..._OnlyWardedTargetMeansNoLegalTarget`
- `TestFR_KWD_002b_ResistantReducesTargetedDamage`, `TestFR_KWD_002k_ArmorAbsorbsTargetedDamageOneForOne`, `TestFR_KWD_002c_StalwartDoesNotReduceNonCombatDamage`
- `TestFR_TGT_003_NoLegalTargetAsksNothingDamagesNothingAndDeploys`
- `TestFR_TGT_003_TargetGoneAtResolutionTakesNoDamageButOtherStepsRun`
- `TestFR_TGT_003_PartialInvalidationResolvesForTheRemainingLegalTarget`
- `TestFR_TRG_004_DeathTriggerFromTargetedDamageResolvesBeforeTheNextStackItem` (the named deep-cascade case)

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- `func (s *scene) putScratch(p PlayerID) *Asset`: places a face-down, Exhausted Scratch Company in `p`'s Reserves (a 1/1 Unit with subtypes Infantry and Scratch Company, no abilities, per `FR-ZON-005`) and returns it. It sets up state and asserts nothing.
- The synthetic-card builders (`tgt`, `dmgTo`, `targetTrigger`, `scriptRefs`) live in the test file.

## Allowed Files
- `pkg/engine/ability/ability.go`
- `pkg/engine/trigger.go`
- `pkg/engine/stack.go`
- `pkg/engine/damage.go` (only if needed for a nil source)

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run '<regex from the Tests section>'`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go run ./cmd/cardtown sim -deck starter -p0 ANIMUS -p1 PATH` and `go run ./cmd/cardtown sim` exit 0 with every invariant holding. No executable count changes, because no card is defined.

## Log
- 2026-10-20 lead-architect: Wrote the task (wave A2 task 1), with status `tests` for `rules-test-author`. Planned A2 order, with deck cards first per directive #14. Only this task is written, and the others stay unwritten until their dependencies are done:
  1. This shape.
  2. Card task (flash): *Saboteur* (deck) and *Chainbreaker Sappers* (non-deck, allowed by #14c because its shapes exist).
  3. Shape: the `toZone` deploy gate (*Devilhorn Guerrillas*, deck), plus `Heal` and `AddArmor` (*Spoil Forest Colossus*, deck; *Ironclad Colossus*).
  4. Shape: "up to N", "other", optional "you may", `Value` from the source's power, and "all assets" (*Craterfall Ancient*, deck).
  5. Shape: the last-known effective power snapshot (*Necropolis Hunt*, deck; triggers directive #13).
  The card tasks follow each shape.
- 2026-10-20 rules-test-author: Wrote the tests above and set status `tests-review`. Helper requested: `putScratch`. `go vet` fails only on the missing vocabulary (first error `ability.Allegiance`) and `putScratch`. `cardtown lint rules` passes. Left out or adapted:
  - The deep-cascade example has Y "Scratch 1" instead of "Draw 1". Both draws go to the hand, so draws could not show which one took the top card. The ordering rule (`FR-TRG-004`) is the same.
  - The re-verification case cannot show "no damage to the gone target", since a dead asset cannot be inspected. The test checks that nothing else takes damage and that the second `Draw 1` still happens.
  - Damage through Resistant, Armor and Stalwart is named `TestFR_KWD_002b/002k/002c_*`, because those are the FRD lines that fix the numbers. Those IDs are not in this task's `requirements:`. The lead may want to add them.
  - The HQ-loss test cites `FR-LAY-002`/`FR-CBT-003`, not a separate win-condition ID, so no test pins an FR-WIN-* line.
  - The allied-unit option sets include the source itself, as the task's Lead decisions say. FRD 01 does not address it, and "other" is a later task.
- 2026-10-20 lead-architect: Test review: **approved**, status `ready`. The tests are in `targeting_test.go` (1d3b49f, with the task file in 76e60dc), and no existing test changed.
  - The 15-row filter table checks each FR-TGT-001 criterion and their conjunctions through the recorded `DecideTarget` options. Single-option rows check the fast path instead.
  - Ward is checked both ways, and an only-Warded board counts as no legal target.
  - The Resistant, Armor and Stalwart checks fix non-combat mitigation. The HQ-to-0 loss stops the deeper trigger.
  - FR-TGT-003 covers no target, a target gone at resolution, and partial invalidation, each with its untargeted Draw still running.
  - The deep cascade (directive #11) shows order on named deck cards: the death-trigger Draw takes the top card, and the deeper trigger's Scratch takes the second. Swapping Y to Scratch is the right call.
  - The tests reuse the existing helpers (`trigUnit`, `deployFromHand`, `mortal`, `killedTrigger`, `inZone`, `scratched`).
  - **For the engineer:** build `putScratch` exactly as the kernel's Deploy Scratch Company does in `scratch` (`stack.go`). Use a real card `Instance` owned by `p` and an `Asset` with `FaceDown: true` and `Exhausted: true`, placed in Reserves. Don't make a synthetic card that merely carries a "Scratch Company" subtype. `NonScratch` must recognize real Scratch Companies.
  - Next: dispatch `kernel-engineer` (tier sonnet).
- 2026-10-20 kernel-engineer: Implemented; status `review`.
  - Changed: `ability/ability.go` (vocabulary exactly as in the task, no renames); `stack.go` (`StackItem.Chosen []int`); `trigger.go` (`chooseTargets` called as each ability item is pushed in `flushTriggers`, after DecideOrder, one `DecideTarget` per slot via `g.choose`; `legalTarget` filter + Ward check; `resolveAbility` re-verifies each chosen asset, then `OpDealDamage` calls `dealDamage(source-or-nil, target, N, false)`); `scenario_test.go` (`putScratch`, built like `scratch` in stack.go: real Instance, FaceDown, Exhausted, in Reserves).
  - Reused: `Decision`/`DecideTarget`/`g.choose` (fast path and fallback), `Printed().HasSubtype("Scratch Company")`, `g.Keywords(a).Ward`, `dealDamage`. `damage.go` unchanged: a nil source is safe because dealDamage only logs it with %s, and fmt prints a nil *Asset as <nil> (it recovers the String panic). Reviewer may prefer an explicit guard.
  - Review points: option order is oldest asset first (by ID), because the tests pick index 0 and expect the source to survive. A slot with no legal target stores 0 and its steps skip. The Ward legality test is also applied at resolution.
  - Gates: targeting tests pass; `go tool task check` green; `cardtown sim -deck starter -p0 ANIMUS -p1 PATH` and `cardtown sim` exit 0, "Every match terminated with all invariants holding". Default sim before and after: P0 66.0% / P1 30.0% / draws 4.0%, avg 30.2 turns (unchanged, no card defined).
