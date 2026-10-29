---
id: T-0099
title: A combat-kill trigger fires only when the sweep destroys the defender, so a replaced death is no kill (FR-CBT-002, SVY-R19, directive #73, wave G shape, sonnet)
kind: shape
tier: sonnet
wave: G
status: ready
branch: m2/T-0099-replaced-defender-no-kill
depends_on: [T-0098]
claims: [trigger:awaits-destruction, hook:combat-kill gate]
cards: []
requirements: [FR-CBT-002, FR-REP-001, FR-TRG-003, FR-TRG-004]
---

## Goal
"Attacks and kills" fires only when the death sweep after the Attack destroys the defender. If a destruction replacement saves it (*Shifting Impossibility* returning a Scratch Company unit to hand), the attacker's combat-kill trigger is never placed (`SVY-R19`). Everything T-0080 pinned stays: the trigger fires at Step 3 from the declared attacker only, still fires when the attacker dies in the same sweep, and is placed in APNAP order with the defender's death triggers. This applies board directive #73 and completes design note Deliverable 6c.

## Inputs
- FRD 01, `FR-CBT-002` "Post-Attack & Combat-Kill Triggers" (line 531): the victim is "the destroyed asset", and "If a replacement effect keeps the defender from being destroyed, it was not killed, and combat-kill abilities do not trigger (`SVY-R19`, `SVY-R15`)." `FR-REP-001` (Replaced Actions Still Count as Done, `SVY-R15`: a trigger watching for the replaced event does not fire). `FR-TRG-003`, `FR-TRG-004` (APNAP placement after the sweep).
- Design Outcome `docs/design/outcomes/2026-10-28-combat-kills-and-replaced-deaths.md` (applied in FRD 01 v3.15). Board directive #73. The Overkill half of #73 is M3 and already recorded (design note Open Question 12, corrected in T-0096's task commit); nothing here builds Overkill.
- Design note: `docs/architecture/m2-ability-model.md`, Deliverable 6c (T-0099).
- Kernel today:
  - `resolveAttack` (`combat.go`) calls `g.fireCombatKill(att)` when `g.IsLethal(def)` after the combat damage, before the sweep. `fireCombatKill` (`trigger.go`) is `fireHook(att, HookCombatKill)`, which appends to `g.pending`.
  - `resolveStack` (`stack.go`) runs `resolve`, then `deathSweep`, then `flushTriggers` for every item, so a combat-kill trigger always meets exactly one sweep before its flush.
  - `deathSweep` → `sweepDeaths` → `replaceDeaths` replaces what a replacement applies to; `destroyAll` (`damage.go`) destroys the rest, calling `fireKilled` for each. A replaced asset never reaches `destroyAll`.
  - Asset IDs start at 1 (`newID`), so 0 means "no asset".

### Shape (lead decisions, binding)
How the kernel knows the death was replaced is the lead's call (#73). The trigger still fires at Step 3, where the attacker is on the battlefield and its controller and position are known, but it waits in the buffer on the defender's destruction. `destroyAll` confirms it; the flush drops any trigger still waiting. A replaced or surviving defender never reaches `destroyAll`, so its kill is never confirmed.
- **`pendingTrigger` (`trigger.go`):** add `awaits int // an asset this trigger waits on: placed only if destroyAll destroys it before the flush (SVY-R19); 0 for none`.
- **`fireCombatKill(att, def *Asset)` (`trigger.go`):** remember `n := len(g.pending)`, call `g.fireHook(att, ability.HookCombatKill)`, then set `awaits = def.ID` on `g.pending[n:]`. Doc comment: resolveAttack calls it when the attacker's combat damage left the defender lethal; the triggers wait on the defender's destruction, because a replaced death is not a kill (`SVY-R19`).
- **`confirmDestroyed(id int)` (`trigger.go`):** clears `awaits` on every pending trigger whose `awaits == id`. One line of doc comment.
- **`flushTriggers` (`trigger.go`):** right after the `Outcome.Over` check, drop every buffered trigger with `awaits != 0` (`slices.DeleteFunc`), before the per-player split, so a dropped trigger is never offered in a `DecideOrder` and never asked about as optional. Add a sentence to the doc comment: a trigger still waiting on an asset's destruction is dropped, because the sweep did not destroy it (`SVY-R19`).
- **`destroyAll` (`damage.go`):** call `g.confirmDestroyed(a.ID)` for each destroyed asset, in the existing loop beside `fireKilled`.
- **`resolveAttack` (`combat.go`):** `g.fireCombatKill(att, def)`. Extend the comment above it: the trigger waits on the sweep destroying the defender (`SVY-R19`).
- **`HookCombatKill` comment (`ability/ability.go`):** "this asset attacked, its combat damage left the defender lethal, and the death sweep after the Attack destroyed it; a replaced death is not a kill (FR-CBT-002 Step 3, SVY-R19)".
- **Unchanged:** `stack.go`, `replace.go`, `fireHook`, `posOf` and the trigger's placement order (the `pos` recorded at Step 3), `doc.go`, the cards, the clients.
- **Sims:** not a pure move. The Path starter Headquarters is *Eternal Watchman* and the Animus deck plays *Shifting Impossibility*, so the starter sim can change. `go tool task sim` must run clean; record its result line.
- **Reuse:** searched `pkg/engine` and the `done` tasks for an existing "only if destroyed" gate on a buffered trigger and found none. `HookKilled` triggers are fired from inside `destroyAll`, so they need no gate; the combat-kill trigger cannot move there, because its source (the attacker) may already be gone and its placement is fixed at Step 3 (T-0080). `awaits` is generic (an asset ID), not a combat-only flag, so a later "dies this way" gate can reuse it.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
New file `pkg/engine/combat_kill_replaced_test.go`. Use the real *Eternal Watchman* through `wmScene` and `wmAttack` (`cards_headquarters_watchman_test.go`), the real *Shifting Impossibility* (`s.card`), face-down Scratch Companies (`s.putScratch`), and synthetic combat-kill triggers through `combatKillTrigger` and `wantScratches` (`combat_kill_test.go`), plus the replacement helpers in `replacement_test.go` (`guardianStructure`, `guardianUnit`, `wantReplaced`). No existing test is edited; every T-0080 test in `combat_kill_test.go` and every *Eternal Watchman* test must still pass unchanged.

- `TestFR_CBT_002_CombatKillTriggerDoesNotFireWhenTheDeathIsReplaced` (named by #73): *Eternal Watchman* attacks and deals lethal damage to an opposing face-down Scratch Company unit while its controller has *Shifting Impossibility*. The unit returns to its owner's hand, *Shifting Impossibility* takes 1 damage, Watchman's trigger is never placed, and no card returns from player 0's discard pile (player 0's pile holds a unit card so the trigger would have a target).
- `TestFR_CBT_002_CombatKillTriggerFiresWhenAnotherDeathInTheSweepIsReplaced`: a synthetic attacker with a combat-kill Scratch 1 kills an opposing structure, which `guardianReplacement`'s unit filter does not cover, while a different opposing unit, lethal at the same sweep (for example through damage set before the Attack), is replaced by player 1's guardian. The defender is destroyed and the trigger fires once: the gate is per defender, not "any replacement in the sweep".
- `TestFR_CBT_002_ReplacedDefenderGivesNoTriggerWhenTheAttackerAlsoDies`: the synthetic attacker kills a guardian-covered defender and dies to its retaliation in the same sweep. The attacker is destroyed, the defender is replaced, and the combat-kill trigger does not fire (contrast T-0080's same-sweep case, where the defender is destroyed and it does).
- `TestFR_CBT_002_DroppedCombatKillTriggerIsNotOfferedForOrdering`: the attacker's controller would have two triggers from the Attack, the attacker's combat-kill trigger and a damage-taken trigger (`HookDamaged`) on the attacker from retaliation that it survives; the defender's death is replaced. Player 0 is asked no `DecideOrder`, and only the damage-taken trigger resolves (`FR-TRG-003`: a single trigger is not asked about).

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected. The new test file may declare its own helpers.

## Allowed Files
- `pkg/engine/trigger.go` (`pendingTrigger.awaits`, `fireCombatKill`, `confirmDestroyed`, `flushTriggers`)
- `pkg/engine/damage.go` (`destroyAll` confirms)
- `pkg/engine/combat.go` (`resolveAttack` passes the defender, comment)
- `pkg/engine/ability/ability.go` (`HookCombatKill` comment only)

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run 'TestFR_CBT_002|TestFR_TRG_003_CombatKill|TestCard_EternalWatchman|TestCard_ShiftingImpossibility'`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks

## Log
- 2026-10-29 lead-architect (wave G run, session 7): Wrote the task, the last of design note Deliverable 6c after T-0098 (done, 3215a5b). It applies directive #73.
  - **Sync:** `main` (f8f9dbb) has nothing the run branch lacks; no merge.
  - **Directives:** read `BOARD.md` on `autoloop/m2-wave-g` fresh. #73 is `open: before the first wave C task`, and this task applies it. #74 and #75 are `applied`. #14 is a standing order; this task has no card. Waiting triggers checked: #6 and #7 (mode is `serial`); #8 (no `deck.go` change); #15 (no turn hook); #20 (b) (no wave C shape task); #58 (no exhaust by a cost or an effect); #59 (2) (no repeated targets); #61 (no modal card); #71 (no effect on each player). None has fired.
  - **Reuse and placement:** see Shape, Reuse. Every file sits where the Boundaries table puts it (`engine` for triggers, combat and the sweep; `ability` for the hook's comment).
  - **Deck cards still on `NotYetExecutable` (#14 (e)):** 22 (9 Animus, 13 Path); this task changes none.
- 2026-10-29 rules-test-author: Wrote `pkg/engine/combat_kill_replaced_test.go` (4 tests, all compile; they fail now for the intended reasons, and the T-0080 and Watchman tests are untouched).
  - `TestFR_CBT_002_CombatKillTriggerDoesNotFireWhenTheDeathIsReplaced`, `..._CombatKillTriggerFiresWhenAnotherDeathInTheSweepIsReplaced`, `..._ReplacedDefenderGivesNoTriggerWhenTheAttackerAlsoDies`, `..._DroppedCombatKillTriggerIsNotOfferedForOrdering`.
  - **Helpers requested:** none.
  - **Left out:** nothing. Note for case 2: the bystander's lethal damage is set before the Attack (Damage = Health), as the task suggested.
  - `cardtown lint rules`: every TestFR_* traces to an active requirement.
- 2026-10-29 lead-architect (wave G run, session 7): Tests approved (D4). Each matches `FR-CBT-002` line 531 and `SVY-R19`: the real *Eternal Watchman* killing a Scratch Company that *Shifting Impossibility* saves places no trigger and returns no card (an unscripted chooser would also catch a target question); a destroyed structure defender still fires when a bystander's death in the same sweep is replaced, so the gate is per defender; a replaced defender fires nothing even when the attacker dies to retaliation in the same sweep (T-0080's same-sweep case stays as it is); and the dropped trigger is never offered in a `DecideOrder`, with Scratch 3 against Scratch 1 showing which trigger resolved. The bystander's damage set before the Attack is lethal at the Attack's sweep, since `resolveStack` sweeps only after the item resolves. Status `ready`; dispatching `kernel-engineer` (sonnet).
