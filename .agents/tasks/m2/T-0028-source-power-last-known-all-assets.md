---
id: T-0028
title: Damage equal to the source's power, last-known effective state, and steps on all matching assets
kind: shape
tier: opus
wave: A2 (task 5, shape)
status: review
branch: m2/T-0028-source-power-last-known-all-assets
depends_on: [T-0019, T-0027]
claims: [type:ValueSource, const:ValueFixed, const:ValueSourcePower, field:Value.From, const:SubjectAll, field:Step.Filter, func:lastKnownEffective]
cards: []
requirements: [FR-ZON-004, FR-TGT-002, FR-KWD-002h, FR-REP-004, FR-CBT-003, FR-TRG-004]
---

## Goal
Abilities can deal damage equal to their source's power, and can hit every asset that matches a filter. The kernel also keeps an asset's last-known *effective* characteristics when it leaves the battlefield (board directive #13). These shapes make two deck cards possible:
- *Necropolis Hunt*: "When this unit is killed, deal damage equal to its power to chosen non-Headquarters asset." Its card task comes next (T-0029, flash).
- *Craterfall Ancient*'s death trigger: "deal 4 damage to all assets". Its attack trigger also needs "you may", "up to 2" and "other", which come in a later shape task.

This task has no cards of its own, and the tests use synthetic cards.

## Inputs
- Board directive #13 (`.agents/tasks/m2/BOARD.md`). This is the first shape task with a `Value` that reads last-known stats.
- Survey ruling `SVY-R10` (`docs/design/m2_starter_set_ability_survey.md`, ruled by the user on 2026-10-21; not yet in FRD 01): "its power" is the source's current effective power when the step resolves. If the source has left the battlefield, it is the source's last-known effective power.
- FRD 01:
  - `FR-ZON-004`: leaves-battlefield triggers use last-known characteristics.
  - `FR-TGT-002`: "all assets" applies globally without targeting.
  - `FR-KWD-002h`: Ward doesn't stop non-targeted effects.
  - `FR-REP-004`: each damage instance goes through the five stages.
  - `FR-CBT-003`: both Headquarters destroyed in one sweep is a draw.
  - `FR-TRG-004`: triggers caused while an item resolves go on the stack before the next deeper item resolves.
  - Layer 6: effective power clamps at 0.
- Design note: `Value` ("fixed N, source power, last-known power") and `Subject` ("all matching a filter"). Changes in `engine`, Layers: a snapshot of effective values at departure.
- Kernel today:
  - `Value{N}` is fixed only, and `Step.Subject` has controller, self, opponents and target.
  - `leaveBattlefield` (`state.go`) stores a raw copy of the `Asset` in `lastKnown`.
  - `deathSweep` (`damage.go`) calls `fireKilled` then `leaveBattlefield` for each dying asset, one at a time.
  - Effective values come from `Power`, `MaxHealth` and `Keywords` (`layers.go`). With no auras or grants, they are computed from the asset alone.
  - `resolveAbility` (`trigger.go`) passes `g.Asset(it.Source)` as the damage source, which is nil once the source has left.
- Rulings that apply: `SVY-R10`, plus `SVY-R3` (an "all" step doesn't target, so Ward and `FR-TGT-003` don't apply).

### Shapes
```go
// pkg/engine/ability
// ValueSource is where a Value's number comes from.
type ValueSource int
const (
	ValueFixed       ValueSource = iota // the zero value: N
	ValueSourcePower                    // the source's effective power (SVY-R10); N is ignored
)
type Value struct {
	N    int
	From ValueSource
}

// Subject gains:
	SubjectAll // every battlefield asset matching Step.Filter when the step resolves; not a target

// Step gains:
	Filter Filter // SubjectAll: which assets
```

### Lead decisions (technical)
- **`ValueSourcePower`** is evaluated when the step resolves.
  - If the source asset (`StackItem.Source`) is on the battlefield, the value is `Power(source)`.
  - Otherwise, it is the effective power recorded when the source left the battlefield.
  - If neither exists, it is 0.
  - A death trigger always takes the second path.
  - It applies to any op that reads `Amount`. Only `DealDamage` uses it in this task's tests.
- **Last-known effective state (directive #13):**
  - When assets leave the battlefield, the kernel records each one's effective Power, MaxHealth and Keywords along with the raw copy it keeps today.
  - The record is taken "immediately prior to departure" (`FR-ZON-004`). When the death sweep destroys several assets at once, compute every dying asset's record before any of them leaves. That way a future aura on a dying asset still counts for the others. Today this is equivalent, so no test can tell the difference. The reviewer checks the order in the code.
  - Other departures (relocation effects, later bounce effects) record at their own departure.
  - The existing `LastKnown(id) *Asset` keeps working for Berserk cleave. The effective record is a new accessor, and the engineer chooses its name.
- **`SubjectAll`:**
  - It resolves to every battlefield asset that matches `Step.Filter` (allegiance relative to the ability's controller, `Kinds`, `NonScratch`) when the step resolves, in oldest-first (ID) order.
  - Each asset gets its own damage instance (`FR-REP-004`). All instances of one step are dealt before the death sweep, so the deaths are simultaneous.
  - It is not a target. Ward doesn't exclude anything, and no choice is asked. The source itself is included if it's still on the battlefield, because the filter has no "other".
  - If nothing matches, the step does nothing.
  - It works for `DealDamage`, `Heal` and `AddArmor`.
- **A damage source that has left the battlefield:**
  - The damage event's `Source` is still the source asset's ID, so "deals damage" attribution survives (`FR-REP-004` Stage 5).
  - Its `Detail` names the source by its last-known record.
  - No event `Detail` may contain `%!` (a formatting failure from a nil asset). `dealDamage` gets an explicit guard for a nil source (follow-up from T-0017).
- **Fences:** no card names. `OPEN(…)` markers only for real survey or matrix items.
- **Not in this task:**
  - "you may", "up to N", "other" and distinct targets (`SVY-R9`, `SVY-R11`): the next shape task;
  - a last-known *value* for anything but power;
  - keyword or ability grants;
  - switching the rule monitors to effective keywords (directive #13 ties that to the first keyword grant, and there still isn't one);
  - the Necropolis Hunt and Craterfall Ancient definitions.

### What the tests can observe
All with synthetic cards in a new `pkg/engine/source_power_test.go`.
- **Power at resolution (`SVY-R10`):**
  - A unit with power 3 and Aggressive 2 also has an attack trigger: "deal damage equal to its power to chosen opposing unit".
  - If the player orders the Aggressive trigger to resolve first, the damage is 5. If the power trigger resolves first, the damage is 3.
  - The value is read when the step resolves, not when the trigger is placed.
- **Source gone before resolution (`SVY-R10`):**
  - A unit with power 3 and Aggressive 2 has two attack triggers: one deals 9 damage to a chosen allied unit (it chooses itself), and the other deals damage equal to its power to a chosen opposing unit.
  - Ordered so that Aggressive resolves first, then the self-kill (the death sweep destroys it), then the power trigger: the power trigger still resolves and deals 5, the last-known effective power.
  - The damage event's `Source` is the dead unit's asset ID.
- **Death trigger (`FR-ZON-004`):**
  - A unit with power 3 and Aggressive 2 attacks, and lethal retaliation kills it. Its death trigger "deal damage equal to its power to chosen non-Headquarters asset" deals 5.
  - The same unit killed outside combat (no Aggressive bonus) deals 3.
  - A 0-power unit deals 0.
  - The options exclude both Headquarters and the dead unit.
- **All assets:** a death trigger deals 4 damage to every asset.
  - Every unit, structure and Headquarters on both sides takes one 4-damage instance, opponents' Ward assets included.
  - Resistant, Stalwart (non-combat, so no cap) and Armor apply per instance.
  - Assets with 4 or less health die in the one sweep.
  - Units it kills that have death triggers fire them afterwards, ordered APNAP (`FR-TRG-003`, `FR-TRG-004`).
  - With both Headquarters at 4 or less current health, the match is a draw with the "both players lost" reason (`FR-CBT-003`).
  - With only one Headquarters at 4 or less, that player loses.
- **Filtered "all":**
  - "deal 1 damage to all opposing units" hits only opposing units: Scratch Companies yes, structures and HQs no. Allied units are untouched.
  - With no match, nothing happens and nothing is asked.
- **Formatting:** across the death-trigger scenes, no event `Detail` contains `%!`.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
All in `pkg/engine/source_power_test.go`. Run regex for Done When: `TestFR_(TRG_004_SourcePower|ZON_004_(SourceGone|DeathTrigger)|TGT_001_DeathTriggerPower|TGT_002_|KWD_002h_(AllAssets|FilteredAll)|REP_004_(AllAssets|NoEventDetail)|TRG_003_AllAssets|TRG_004_AllAssets|CBT_003_AllAssets|KWD_002d_AllHeal|LAY_002_AllAddArmor)`

Source power (`SVY-R10`, `// Covers:` comments):
- `TestFR_TRG_004_SourcePowerIsReadWhenTheStepResolves`: Aggressive resolves first gives 5, power trigger first gives 3.
- `TestFR_ZON_004_SourceGoneBeforeResolutionDealsItsLastKnownEffectivePower`: Aggressive, then self-kill, then power trigger: 5 dealt, event `Source` = the dead unit's asset ID, Raw and Amount 5.
- `TestFR_ZON_004_DeathTriggerUsesLastKnownPowerWithTheAggressiveBonus`: 5, `Source` = dead unit.
- `TestFR_ZON_004_DeathTriggerOutsideCombatHasNoAggressiveBonus`: 3.
- `TestFR_ZON_004_DeathTriggerOfAZeroPowerUnitDealsZero`: 0.
- `TestFR_TGT_001_DeathTriggerPowerTargetExcludesHeadquartersAndTheDeadUnit`: options are exactly the non-HQ assets other than the dead unit.

All assets (`FR-TGT-002`):
- `TestFR_TGT_002_AllAssetsDealsOneInstanceToEveryAssetOnBothSidesWithoutAsking`
- `TestFR_KWD_002h_AllAssetsIgnoresWard`
- `TestFR_REP_004_AllAssetsAppliesResistantStalwartAndArmorPerInstance`
- `TestFR_TGT_002_AllAssetsKillsEveryAssetWithFourOrLessHealthInOneSweep`
- `TestFR_TRG_003_AllAssetsDeathTriggersResolveInApnapOrder`
- `TestFR_TRG_004_AllAssetsDeathTriggersCascade`
- `TestFR_CBT_003_AllAssetsBothHeadquartersDestroyedIsADraw` (reason contains "both players lost", case-insensitive)
- `TestFR_CBT_003_AllAssetsOneHeadquartersDestroyedLosesThatPlayer`

Filtered all and the other ops:
- `TestFR_TGT_002_FilteredAllOpposingUnitsHitsOnlyOpposingUnits`
- `TestFR_TGT_002_FilteredAllNonScratchSkipsScratchCompanies`
- `TestFR_TGT_002_FilteredAllStructuresAndHeadquartersHitsExactlyThose`
- `TestFR_TGT_002_FilteredAllAlliedUnitsIncludesTheSourceItself`
- `TestFR_TGT_002_FilteredAllWithNoMatchDoesNothingAndAsksNothing`
- `TestFR_KWD_002h_FilteredAllHitsOpposingWardUnits`
- `TestFR_KWD_002d_AllHealRemovesDamageFromEveryMatchingAsset`
- `TestFR_LAY_002_AllAddArmorGrantsEveryMatchingAsset`

Formatting: `TestFR_REP_004_NoEventDetailContainsAFormattingFailure` (no `%!` in any Detail across the death-trigger and all-assets scenes; other tests also call the same check).

Deliberately left out:
- Negative-power clamp to 0 (Layer 6): no way to give a unit negative power without grants.
- Whether the "simultaneous" damage events all precede the death events in the log: the event kind names are an implementation detail. Simultaneity is observed through the one-sweep deaths and the both-HQ draw.
- The exact `Detail` text for a dead source ("names the source by its last-known record"): not fixed by FRD 01 or SVY-R10; only the absence of `%!` is pinned.
- The effective-record accessor and its snapshot-before-any-departure order: no test can observe it today; the reviewer checks the code.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None. The tests define their own local helpers in `source_power_test.go` (`scene.scriptAnyOf`, `powerStep`, `allStep`, `noFormatErrors` and others), and use existing ones (`killedTrigger`, `mortal`, `striker`, `targetBoard`, `deployFromHand`, `putScratch`).

## Allowed Files
- `pkg/engine/ability/ability.go`
- `pkg/engine/trigger.go`
- `pkg/engine/state.go`
- `pkg/engine/damage.go`

## Done When
- [ ] The tests pass: `go test ./pkg/engine -run '<regex from rules-test-author>'`
- [ ] Every existing test passes unchanged
- [ ] `go tool task check` passes
- [ ] `go run ./cmd/cardtown sim -deck starter -p0 ANIMUS -p1 PATH` and `go run ./cmd/cardtown sim` exit 0, with every invariant holding and no rule-monitor violation, and with output byte-identical to `rework/executable-rules` (no card uses the new shapes yet)

## Log
- 2026-10-21 lead-architect: Wrote the task (A2 shape), with status `tests` for `rules-test-author`.
  - Tier `opus`, per the user ("use Opus for the shapes this time"; board directive #25).
  - It applies directive #13 and cites `SVY-R10`, which the user ruled on 2026-10-21.
  - It is split from its first card, Necropolis Hunt, to keep within four non-test files.
- 2026-10-21 rules-test-author: Wrote the failing tests in `pkg/engine/source_power_test.go` (list under Tests). Status `tests-review`.
  - Expected failure: the package's tests don't compile, because `ability.ValueSourcePower`, `Value.From`, `ability.SubjectAll` and `Step.Filter` don't exist yet. With the file moved aside, the engine package passes unchanged.
  - Helpers requested: none.
  - Assumptions: Order options are Aggressive first, then printed triggers (FRD 01 `SVY-R7`), and the "both players lost" draw reason is taken from the task text, not from FRD 01.
  - Omissions: see Tests.
- 2026-10-21 lead-architect: Test review: **changes**, status `tests`.
  - The reviewer prototyped the shapes in a throwaway worktree (not committed). All 27 tests compiled, and 26 pass. Every existing engine test passes.
  - The rules content of all 27 checks out against FRD 01, `SVY-R10` and the lead decisions.
  1. `TestFR_TGT_002_FilteredAllAlliedUnitsIncludesTheSourceItself` fails at `sweeper`'s precondition "Sweeper is on the battlefield". `deployFromHand` resolves the whole stack, so the deploy trigger has already killed the 1/1 Sweeper, which is exactly what the test means to show. `sweeper` must not require the source to survive. For example, take the source's asset ID from the `deploy` event, or check that its instance is in its owner's Discard. Then assert the death through that, and keep the other filtered-"all" tests' checks unchanged.
- 2026-10-21 rules-test-author: Answered note 1. `sweeper` no longer requires the source to survive: it returns the source's asset (nil if destroyed) and its card.
  - `TestFR_TGT_002_FilteredAllAlliedUnitsIncludesTheSourceItself` now asserts the death by the source's card instance being in its owner's Discard and no longer on the battlefield.
  - Every other test's assertions are unchanged. Status `tests-review`.
- 2026-10-21 lead-architect: Test review, round 2: **approved**, status `ready`.
  - a1a3705 answers note 1. `sweeper` returns the source's asset (nil when destroyed) and its card, and the allied-"all" test asserts that the source is off the battlefield and in its owner's Discard. No other assertion changed.
  - Reviewer re-check with the throwaway prototype (not committed): `go test ./pkg/engine` passes, all 27 new tests included.
  - Next: `kernel-engineer` on **Opus** (tier `opus`, directive #25).
- 2026-10-21 kernel-engineer (Opus): Implemented in 4de1413. Status `review`. No test file was edited, and only the four Allowed Files changed.
  - `ability.go`: adds `ValueSource` (`ValueFixed`, `ValueSourcePower`), `Value.From`, `SubjectAll` and `Step.Filter`, exactly as in Shapes.
  - `state.go`: `lastKnown` now maps to a `lastKnownRecord`, which holds the raw `Asset` copy and an `EffectiveState` (`Power`, `MaxHealth`, `Keywords`). `effectiveState(a)` reads those through the existing `layers.go` functions.
    - `LastKnown(id) *Asset` keeps its signature and behavior.
    - The new accessor is `LastKnownEffective(id) *EffectiveState`, which is nil if the asset never left. This is the claimed `func:lastKnownEffective`, exported.
    - `leaveBattlefield(a, eff)` now takes the state from its caller. A lone departure passes `g.effectiveState(a)`.
  - `damage.go`:
    - `deathSweep` computes `known[i] = g.effectiveState(a)` for every dying asset in one loop. Only then does the existing loop run `fireKilled`, `leaveBattlefield(a, known[i])`, discard and the log. The event order is unchanged.
    - `dealDamage` formats the source through an explicit nil guard (`"an unknown source"`), so `%!` cannot appear.
  - `trigger.go`:
    - `resolveAbility` reads every `Amount` through `value(it, v)` when the step resolves. For `ValueSourcePower` the order is: `Power(source)` if the source is on the battlefield, else `LastKnownEffective(source).Power`, else 0.
    - DealDamage, Heal and AddArmor act on `stepAssets(it, st, chosen)`: the chosen target if it is still legal, or for `SubjectAll` every asset matching the filter, oldest ID first. There is one `dealDamage` per asset, and all of them come before the item's sweep.
    - The damage source is `itemSource(it)`: the live asset, else its `LastKnown` copy. So `Source` stays the asset ID, and `Detail` names it as `Name#ID`.
  - Reuse:
    - `legalTarget` is now the Ward check plus the new `matchesFilter` (allegiance, kinds and NonScratch, moved out of `legalTarget` without change), and `SubjectAll` calls `matchesFilter` alone.
    - The oldest-first sort that was inline in `chooseTargets` became `assetsByID()`, which `chooseTargets` and `SubjectAll` share.
    - `itemSource` follows the existing live-or-`LastKnown` fallback in `resolveAttack`.
    - I searched for an existing value evaluator or "all assets" selector and found none.
  - Verification:
    - `go test ./pkg/engine -run '<Done When regex>' -v`: PASS (all T-0028 tests, plus the three older `TestFR_ZON_004_DeathTrigger*` tests the regex also matches).
    - `go test ./...`: exit 0.
    - `go tool task check`: exit 0 (fmt-check, vet, test, `lint --all`, fences PASS).
    - `go run ./cmd/cardtown sim -deck starter -p0 ANIMUS -p1 PATH` and `go run ./cmd/cardtown sim`: exit 0, `[PASS] Every match terminated with all invariants holding.`, and the monitors found no violation.
    - Sim byte-identity: the baseline CLI was built from `rework/executable-rules` (b862750) in a separate `git worktree` before any code edit. For each of the two policies (`progress` and `random`), the stdout of three runs (starter ANIMUS vs PATH, starter PATH vs ANIMUS, and stress) is byte-identical (`cmp`), and so are both `go run` outputs.
    - `-events` JSONL for starter ANIMUS vs PATH is also identical, except for the `match-start` build stamp (`revision`/`modified`). Neither log contains `%!`.
  - For the reviewer:
    1. The snapshot order is in `deathSweep`: the `known` loop runs before the `fireKilled`/`leaveBattlefield` loop.
    2. Intended event change: a damage event from a departed source used to carry `Source: 0` and `%!s(<nil>)`. It now carries the dead asset's ID and `Name#ID` in `Detail`. No starter or stress card reaches this path, which is why the sim output is unchanged.
    3. `dealDamage` may now get a source that is a last-known copy and is not on the battlefield. Today it reads only the source's ID and name. A future "when this deals damage" hook must not treat that copy as a live asset.
    4. `SubjectAll` evaluates the value once per step, before any instance is dealt.
    5. Directive #13's monitor switch to effective keywords stays out of scope, per this task's "Not in this task" list.
