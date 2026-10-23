---
id: T-0050
title: Damage-taken hook for Unbowed (FR-KWD-002a), with step<Op> functions for effect ops (wave B shape 1, sonnet)
kind: shape
tier: sonnet
wave: B
status: tests
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
<!-- rules-test-author fills this in: TestFR_KWD_002a_* and any TestFR_REP_004_* cases, from FRD 01 and the matrix rows above. Use test-local definitions, not real card names, since no Unbowed card is executable yet. Cover at least: fires on Health damage; fires on Armor-only absorption; not when Resistant reduces it to 0; fires when Stalwart caps it to 1; fires for both combatants in combat and on a Berserk secondary; fires on lethal damage and then resolves from discard (a non-self step resolves; a self Heal/Armor step fizzles; the controller is the last-known controller); two separate damage instances fire twice. Do not pin the Open Question's ordering. -->

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
<!-- rules-test-author lists any setup helpers here. -->

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
