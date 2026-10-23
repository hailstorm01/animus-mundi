---
id: T-0050
title: Damage-taken hook for Unbowed (FR-KWD-002a), with step<Op> functions for effect ops (wave B shape 1, sonnet)
kind: shape
tier: sonnet
wave: B
status: done
branch: m2/T-0050-unbowed-hook
depends_on: [T-0045, T-0046, T-0047, T-0048, T-0049]
claims: [hook:damaged, func:fireDamaged, convention:step<Op>]
cards: []
requirements: [FR-KWD-002a, FR-REP-004]
---

## Goal
An asset's triggered ability can fire when the asset **takes damage**: `armorAbsorbed + healthCommitted > 0` at `FR-REP-004` Stage 5. That is the hook behind **Unbowed** (`FR-KWD-002a`) and, later, a Headquarters' "is dealt damage" (Animoceps Rager, which also needs a once-per-turn limit, a separate shape).

When this lands, the card task that follows can make **Hammerhand Brigands** (`Unbowed: Scratch 1.`) and **Devilhorn Captain** (`Unbowed: Deal 1 damage to chosen asset in the Main Theater.`) executable. Both are starter-deck cards.

This task also carries board directive #40: every effect op's body gets its own `step<Op>` function before wave B adds more ops.

## Inputs
- FRD 01:
  - `FR-KWD-002a` Unbowed (line 505 and the two bullets after it);
  - `FR-REP-004` Stage 5, damage-taken triggers (around line 454);
  - `FR-TRG-003`, APNAP, and `FR-TRG-004`, cascading placement after the resolving item and its sweeps;
  - `FR-ZON-004`, last-known information, and `FR-TGT-003`, fizzling.
- Keyword matrix rulings: `Unb × Res`, `Unb × Stw`, `Unb × Heal`, `Unb × Scr` and `Unb × Bsk` (`docs/frd/keyword_interaction_matrix.md` lines 74–81).
- Survey: wave B (§2), and the rows for Hammerhand Brigands, Devilhorn Captain, Chainbreaker Irregulars, Allshan and Macabre Dancers (§6).
- Design note: `docs/architecture/m2-ability-model.md`, Boundaries and "Changes in `engine`".
- Board directives:
  - #40: the `step<Op>` convention, folded in here;
  - #10: trigger-order timeout uses collection order;
  - #14: deck cards first. Brigands and Captain are deck cards.
- Existing machinery to reuse (search before adding):
  - `DamageResult.Taken()` in `damage.go`;
  - `fireHook` and the pending-trigger buffer in `trigger.go`;
  - death triggers that track the card in discard (`SubjectSelf`, T-0009);
  - `itemSource`, last-known effective state (T-0028);
  - the APNAP flush.

### Lead decisions (technical)
1. **The hook:**
   - `ability.HookDamaged`: "this asset took damage: Armor absorbed plus Health committed was more than 0 (`FR-REP-004` Stage 5)".
   - In the engine, `dealDamage` fires it on the **target** whenever `r.Taken()`. It fires once per damage instance, for combat and non-combat damage alike. It does **not** fire when the damage was reduced to 0 before Stage 4.
   - The trigger goes into the existing pending buffer like every other hook.
2. **A trigger whose source has been destroyed** resolves from discard, as death triggers already do:
   - its controller is the source's last-known controller;
   - steps that act on the destroyed asset itself fizzle (`FR-TGT-003`);
   - other steps resolve fully (`FR-KWD-002a`'s second and third bullets).
3. **Directive #40, done first as a pure move:**
   - `resolveAbility` keeps the second-stage verification and a `switch` that only dispatches.
   - Today's inline bodies become `(g *Game) step<Op>(it *StackItem, st ability.Step, chosen [][]*Asset)` methods: `OpDealDamage`; `OpHeal`/`OpAddArmor`, as one or two methods; and `OpScratch`/`OpDraw`. The existing `relocateSelf` call becomes `stepRelocate` or stays as the dispatch target.
   - Any op this task needs follows the same rule. It is expected to need none: Scratch and DealDamage exist.
   - Commit the move separately (`refactor(engine): step<Op> functions for effect ops (T-0050)`), before the hook. It must leave every test passing and `cardtown sim` output byte-identical.
4. **Placement timing is not decided in this task.** See Open Question below. Tests and code follow `FR-TRG-004` as written: a damage-taken trigger goes into the same pending buffer as the death triggers of the same window, and is placed with them in APNAP order. No test may pin the relative order of a dying asset's Unbowed and death triggers from the same sweep.
5. **Not in this task:**
   - any card definition (the card task follows);
   - the once-per-turn limit, Cornered and other wave B shapes;
   - a rule monitor for Unbowed (optional later; the damage monitor already checks the Stage 5 numbers).

### Open Question (for the user; the lead does not rule)
`FR-KWD-002a` says a lethally damaged asset's Unbowed "enqueues onto the Resolution Stack **before** the asset is destroyed". `FR-TRG-004` says buffered triggers are placed only **after** the resolving item and its state-based sweeps complete.

This matters when a dying asset also has a death trigger, or another asset dies in the same sweep: is the Unbowed trigger placed below those death triggers (so it resolves after them), or together with them in one APNAP window?

No starter card can show it yet. Hammerhand Brigands and Devilhorn Captain have no death trigger. So this task keeps the `FR-TRG-004` reading (decision 4), and the question goes to the user.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
All in `pkg/engine/unbowed_test.go`, with test-local ability definitions (`bowed`, `unbowed`, `healSelf`, `armorSelf`, `putControlled`, `wantScratches`). Outcomes are observed through face-down Scratch Companies, Damage, Armor, discard and recorded Decisions.

Fires or not (FR-KWD-002a, FR-REP-004 Stage 5):
- `TestFR_KWD_002a_UnbowedFiresWhenHealthDamageIsCommitted`
- `TestFR_REP_004_UnbowedFiresWhenArmorAbsorbsAllTheDamage` (Unb x Arm)
- `TestFR_REP_004_UnbowedDoesNotFireWhenResistantReducesDamageToZero` (Unb x Res)
- `TestFR_REP_004_UnbowedStillFiresWhenResistantLeavesSomeDamage`
- `TestFR_REP_004_UnbowedFiresWhenStalwartCapsDamageToOne` (Unb x Stw)
- `TestFR_REP_004_UnbowedDoesNotFireOnAZeroPowerAttack`
- `TestFR_REP_004_UnbowedFiresWhenNegativeResistantAmplifiesZeroToOne`
- `TestFR_KWD_002a_OnlyTheAssetThatTookDamageFires`
- `TestFR_KWD_002a_NonCombatDamageFiresUnbowedOnce`
- `TestFR_REP_004_NonCombatDamageReducedToZeroDoesNotFireUnbowed`
- `TestFR_KWD_002a_TwoSeparateDamageInstancesFireTwice`

Combat and Berserk (FR-CBT-002, FR-KWD-002o):
- `TestFR_KWD_002a_BothCombatantsFireTheirOwnUnbowed`
- `TestFR_KWD_002a_BerserkCleaveDamageFiresTheSecondarysUnbowed` (Unb x Bsk)
- `TestFR_KWD_002a_BerserkCleaveReducedToZeroDoesNotFireUnbowed`
- `TestFR_KWD_002a_RetaliationOnABerserkAttackerFiresItsUnbowed` (Unb x Bsk)

Survivor and lethal damage (FR-ZON-004, FR-TGT-003, FR-TGT-001):
- `TestFR_KWD_002a_SurvivorHealsItselfAfterTheDamageIsCommitted`
- `TestFR_KWD_002a_SurvivorGainsArmorFromItsOwnUnbowed`
- `TestFR_KWD_002a_LethalDamageStillFiresAndANonSelfStepResolves` (Unb x Scr)
- `TestFR_KWD_002a_LethalDamageSelfHealFizzlesButSiblingStepsResolve` (Unb x Heal)
- `TestFR_KWD_002a_LethalDamageSelfArmorFizzlesButSiblingStepsResolve`
- `TestFR_KWD_002a_LethalDamageChosenTargetStepResolvesFromDiscard`
- `TestFR_KWD_002a_LethalDamageTriggerBelongsToTheLastKnownController`

Placement (FR-TRG-003, FR-TRG-004):
- `TestFR_TRG_003_UnbowedTriggersFollowApnapNonActivePlayerResolvesFirst`
- `TestFR_TRG_004_UnbowedTriggerResolvesAfterTheDamageIsCommitted`
- `TestFR_TRG_004_UnbowedDamageCascadesIntoAnotherUnbowed`

Deliberately not tested: the order of a dying asset's Unbowed trigger relative to death triggers from the same sweep (Open Question). No test gives a dying asset both an Unbowed and a death trigger.

Note for the engineer: the self-heal and self-armor tests use `ability.SubjectSelf` with `OpHeal` and `OpAddArmor`. If `SubjectSelf` is not yet accepted for those ops, supporting it is part of making these tests pass (T-0009 already uses that subject for relocation).

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None. The tests define their own local helpers in `unbowed_test.go`. The file compiles only once `ability.HookDamaged` exists (the engineer adds it).

## Allowed Files
- `pkg/engine/ability/ability.go` (`HookDamaged`)
- `pkg/engine/effects.go` (directive #40's `step<Op>` methods)
- `pkg/engine/damage.go` (firing the hook)
- `pkg/engine/trigger.go` (if firing or last-known-controller handling needs it)

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run 'TestFR_KWD_002a|TestFR_REP_004'`
- [ ] The `refactor(engine)` commit for directive #40 comes first, with every test passing and `cardtown sim -deck starter -p0 ANIMUS -p1 PATH` byte-identical before and after it (recorded in the Log)
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks

## Log
- 2026-10-22 lead-architect: Wrote the first wave B shape task. Directive #38's priority rule is satisfied (T-0045 to T-0049 are done), and directive #40 is folded in (decision 3).
  - Scope: the hook only. The two Unbowed starter-deck cards it unlocks, Hammerhand Brigands and Devilhorn Captain, follow as a flash card task. Chainbreaker Irregulars also needs `DISCARD_CARD`, and Animoceps Rager also needs the once-per-turn limit.
  - The placement-timing question between `FR-KWD-002a` and `FR-TRG-004` is raised with the user, and kept out of the tests.
  - Status `tests` on branch `m2/T-0050-unbowed-hook`. Next: the lead dispatches `rules-test-author` (Sonnet).
- 2026-10-22 rules-test-author: Wrote 25 failing tests in `pkg/engine/unbowed_test.go` (listed under Tests).
  - Helpers requested: none (local helpers only). The file fails to compile until `ability.HookDamaged` exists, as expected.
  - Left out on purpose: the Open Question ordering (Unbowed versus death triggers in one sweep).
  - Also not covered: a Headquarters "is dealt damage" trigger (belongs to the Animoceps Rager shape).
  - `cardtown lint rules`: PASS, every `TestFR_*` traces to an active requirement.
  - Status `tests-review`. Next: the lead reviews the tests.
- 2026-10-22 lead-architect: Tests review (D4): **approved**, status `ready`.
  - Each of the 25 tests checked against its FRD 01 citation and the matrix rows. The expected outcomes match:
    - the Stage 5 threshold, including Armor-only absorption and negative Resistant amplifying 0 to 1;
    - Resistant to 0, and the Stalwart cap still firing;
    - Berserk cleave and retaliation;
    - once per damage instance, not per point;
    - a survivor's self-Heal and self-Armor applying after the damage;
    - lethal damage resolving from discard, where the self-step fizzles, the sibling step runs, a chosen target is picked by the last-known controller, and the controller is distinct from the owner;
    - APNAP, with the non-active player's trigger resolving first;
    - the cascade.
  - The 0-power attack case is legal under FRD 01 "Zero-Power Attacks" (line 469).
  - Nothing pins the Open Question: no test gives a dying asset both an Unbowed and a death trigger.
  - Local `scene` methods in the test file follow the precedent of six other engine test files.
  - Note for the engineer, as the test author flagged: `SubjectSelf` on `OpHeal`/`OpAddArmor` may need support. That falls under decision 2, inside Allowed Files.
- 2026-10-22 kernel-engineer: Implemented, status `review`. Two code commits after the `in-progress` commit.
  - Commit 1, `refactor(engine): step<Op> functions for effect ops (T-0050)`, only `effects.go`: `resolveAbility` keeps the second-stage verification and a dispatching `switch`. New methods `stepDealDamage`, `stepHeal`, `stepAddArmor`, `stepScratch`, `stepDraw`, `stepRelocate` (wrapping `relocateSelf`), plus a small `stepPlayers` helper for the `who` computation. Bodies moved unchanged. Baseline CLI built from the branch's starting code; `cardtown sim -deck starter -p0 ANIMUS -p1 PATH` output `cmp`-identical before and after. Every test except `unbowed_test.go` (needs `HookDamaged`) passed at that commit (`go test ./pkg/engine` with that file set aside, `go vet`).
  - Commit 2, the hook: `ability.HookDamaged` (`ability.go`); `fireDamaged` (`trigger.go`) calls the existing `fireHook`; `dealDamage` (`damage.go`) calls it on the target when `r.Taken()`, after logging; `stepAssets` (`effects.go`) gains `SubjectSelf`, returning the source if still on the battlefield, else nothing (fizzle).
  - Reused: `DamageResult.Taken()`, `fireHook` and the pending buffer (controller captured at fire time, so a destroyed source's trigger keeps its last-known controller), `chooseTargets`/`itemSource`/`LastKnown`, the APNAP flush. No new placement logic (decision 4).
  - Verification: `go test ./pkg/engine -run 'TestFR_KWD_002a|TestFR_REP_004'` passes (26 PASS lines incl. subtests/older tests); `go tool task check` exit 0; `go tool task sim` exit 0; starter sim output still byte-identical to the baseline (no card uses the hook yet).
  - Reviewer: look at `SubjectSelf` in `stepAssets` (it is not restricted by hook, unlike `relocateSelf`); Open Question left untouched.
- 2026-10-22 lead-architect: Review: **approved**, merged as 1ca11e9, status `done`.
  - Scope:
    - 91499ad changes only this file.
    - 7b15223 changes only `effects.go`.
    - 1300753 changes only `ability.go`, `damage.go`, `trigger.go`, `effects.go` and this file.
    - `unbowed_test.go` is unchanged since 0eef686.
  - The refactor (directive #40) is a pure move. The bodies are verbatim, and `stepPlayers` merges the identical player-subject code of Scratch and Draw. At 7b15223, with the new test file set aside, every `engine` and `sim` test passes.
  - Rules fidelity: `dealDamage` fires `HookDamaged` on any target, combat or not, once per instance, whenever `Taken()`. That is the general Stage 5 rule, not the tested cases only.
  - `SubjectSelf` in `stepAssets` is the source if it is still on the battlefield, else nothing, so self-steps fizzle (`FR-TGT-003`). It changes no existing card: the only current `SubjectSelf` steps are relocate ops, which go to `stepRelocate`.
  - Gates, re-run in a scratch worktree:
    - 29 `--- PASS` lines for the Unbowed and related runs;
    - `go tool task check` and `go tool task sim` exit 0;
    - against a baseline built from 92dc9a0, the starter sim is byte-identical under both policies.
  - Decision 4 and the Open Question are untouched.
  - The lead recorded the `step<Op>` convention in the design note's "Changes in `engine`" (directive #40), and corrected that section's stale Fences line.
  - Next: a flash card task for Hammerhand Brigands and Devilhorn Captain.
