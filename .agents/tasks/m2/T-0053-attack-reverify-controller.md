---
id: T-0053
title: Re-verify an attack's defender against the attacker's current controller (SVY-R8, directive #29, shape, sonnet)
kind: shape
tier: sonnet
wave: B
status: ready
branch: m2/T-0053-attack-reverify
depends_on: [T-0037]
claims: [func:resolveAttack defender check]
cards: []
requirements: [FR-CBT-002, FR-KWD-002o, FR-TGT-003]
---

## Goal
When an Attack resolves, the defender is re-verified as opposing to the **attacker's current controller** (its last-known controller if the attacker has left), not to the player who declared the attack. If either combatant has changed sides, the attack fizzles with no combat or retaliation damage and no Berserk cleave. This is board directive #29. Its trigger, FRD 01 recording `SVY-R8`, happened on 2026-10-21 (0cadda1, FRD 01 v3.8), but the board still said `waiting` until the lead noticed on 2026-10-23.

## Inputs
- FRD 01:
  - `FR-CBT-002` Step 3, "Combatant Target Check (`SVY-R8`)" (around line 491);
  - the v3.8 changelog entries at lines 27 (`SVY-R8`) and 28 (Ward on placement and at resolution);
  - `KIM-B06` as extended in v3.8 (line 50): no cleave when the defender fails its check, including no longer opposing;
  - `FR-KWD-002o` Berserk, `FR-TGT-003` fizzling, `FR-KWD-002h` Ward.
- Board directive #29 (the full text is the source for this task).
- Code: `resolveAttack` in `pkg/engine/combat.go` (line 94). Its defender check compares `def.Controller` with `it.Controller`, the declaring player. The comment at lines 96–97 points at directive #29; remove it with the fix. `cleave` and `legalSecondary` already check the secondary against the attacker's current controller.

### Lead decisions (technical)
1. **The fix:** in `resolveAttack`, the defender check uses the attacker's current controller if the attacker is on the battlefield, otherwise `g.LastKnown(it.Source).Controller`. The existing attacker-side fizzle and cleave path is unchanged.
2. **A control change on either side** leaves both combatants under one player, so the defender check fails. The attack fizzles, there is no combat or retaliation damage, and there is no cleave, whatever happened to the attacker.
3. **The Berserk secondary changing sides** does not stop combat. The cleave dissipates, which `legalSecondary` already does; the test pins it.
4. **Ward** is checked on placement and at resolution for every chosen target, and never applies to the defender. The kernel already re-runs `legalTarget` at resolution (`effects.go`); only the Berserk secondary has a test. This task pins the chosen-target case; that test should pass when written.
5. **How the tests change control mid-stack:** no starter card changes control while an Attack is on the stack. The tests give the attacker a synthetic optional `HookAttacks` trigger and install a test `Chooser` for its controller (`Chooser` is `func(*Game, Decision) Answer`) that flips `Controller` (or grants Ward) when the "you may" question is asked. That runs at Step 2, after declaration and before the Attack resolves. If a reusable form is wanted, it is a setup-only scene helper in `scenario_test.go`.
6. **Rule monitors:** the targeting and attack-declaration monitors are **not** extended to attack re-verification in this task. No starter card changes control, so a sim check could never fire. A monitor waits for the first `GainControl` card (Rhapsody's Hollow, deferred by directive #14).
7. **Berserk is under review** (survey `R12`, open). Don't widen Berserk work beyond these tests.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
Named by directive #29; rules-test-author may add closely related cases:
- `TestFR_CBT_002_DefenderChangesSidesFizzles`: the defender's control passes to the attacker's controller before resolution. No combat or retaliation damage, no cleave.
- `TestFR_CBT_002_AttackerChangesSidesFizzles`: a Berserk attacker passes to the defender's controller. Same outcome, no cleave.
- `TestFR_KWD_002o_SecondaryChangesSidesDissipates`: combat happens, the cleave dissipates.
- `TestFR_TGT_003_TargetThatGainsWardBeforeResolutionIsIllegal`: an opposing target of a chosen-target trigger gains Ward before it resolves; that clause does nothing to it.

Expected today: the first two fail (the defender check compares with the declaring player); the last two pass.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None. The tests use `pkg/engine/attack_reverify_test.go`'s own `(*scene).midStack` (a test-file chooser wrapper), plus existing helpers. Decision 5's reusable form is not needed.

## Allowed Files
- `pkg/engine/combat.go`

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run 'TestFR_CBT_002|TestFR_KWD_002o|TestFR_TGT_003'`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean, and `cardtown sim -deck starter -p0 ANIMUS -p1 PATH` output is byte-identical before and after (no starter card changes control). Record it in the Log.
- [ ] The `directive #29` comment in `resolveAttack` is gone, and `grep -n "SVY-R8" pkg/engine/combat.go` shows only a requirement citation, if any.

## Log
- 2026-10-23 lead-architect: Wrote the task for directive #29, whose trigger fired on 2026-10-21 and was missed. It runs before the queued T-0052 because directives come first (Runbook D2.0). Status `tests` for `rules-test-author`. Starter-deck cards not yet executable: 29, plus both Headquarters (unchanged by this task).

### Tests written (rules-test-author, `pkg/engine/attack_reverify_test.go`)
- `TestFR_CBT_002_DefenderChangesSidesFizzles`: Berserk attacker; the defender's control passes to the attacker's controller. No combat, retaliation or cleave damage. **Passes today** (the old check compares with the declaring player, so a flipped defender also fizzles).
- `TestFR_CBT_002_AttackerChangesSidesFizzles`: Berserk attacker passes to the defender's controller. Same outcome. **Fails today** (attacker damage 2, defender damage 3).
- `TestFR_CBT_002_CombatantsSwapSidesStillResolves` (added): both combatants swap controllers, so they are still opposing and combat happens (attacker power 3, retaliation 2). **Fails today** (defender damage 0, want 3).
- `TestFR_KWD_002o_SecondaryChangesSidesDissipates`: combat happens, cleave dissipates. Passes today.
- `TestFR_TGT_003_TargetThatGainsWardBeforeResolutionIsIllegal`: chosen-target trigger plus an optional trigger whose answer grants Ward after the target is chosen. Passes today.
- 2026-10-23 rules-test-author: Wrote 5 tests in `pkg/engine/attack_reverify_test.go` (the four named plus the swap case, which FR-CBT-002 Step 3 determines: "opponent of the attacker's current controller"). Mid-stack changes follow lead decision 5 (optional `HookAttacks` draw trigger, test chooser mutates when the "you may" is asked); the helper is local to the test file. Today: AttackerChangesSides and CombatantsSwapSides fail; DefenderChangesSides passes (contrary to the Tests note, a flipped defender equals the declaring player, so the old check fizzles it); the other two pass. Helpers requested: none. Left out: attacker changes sides then leaves the battlefield (last-known controller path), since forcing a death inside the Step 2 chooser would pin setup details, not the rule. `go vet` and `cardtown lint rules` pass.
- 2026-10-23 lead-architect: Test review (D4): **approved**, status `ready`.
  - 574f441 adds `attack_reverify_test.go` (five tests, the local `midStack` chooser and the `sides` board) and changes no existing test.
  - Rules content checks out against FRD 01 v3.8 `FR-CBT-002` Step 3 (`SVY-R8`), `KIM-B06` as extended, `FR-KWD-002o`, `FR-TGT-003` and `FR-KWD-002h`. The added `CombatantsSwapSidesStillResolves` is right: after a swap the defender is still opposing to the attacker's current controller.
  - The test author's correction stands: `DefenderChangesSidesFizzles` passes today, because the flipped defender equals the declaring player. It stays as a guard. Failing today: `AttackerChangesSidesFizzles` and `CombatantsSwapSidesStillResolves`.
  - Accepted omission: no test pins the last-known-controller branch (attacker changes sides, then leaves). Lead decision 1 still requires it, and the code review checks it in the diff.
  - **Reviewer check** (a throwaway fix following decision 1, not committed): all five pass, `go test ./...` passes, and the starter sim is byte-identical to the baseline.
  - Next: the lead dispatches `kernel-engineer` (Sonnet).
