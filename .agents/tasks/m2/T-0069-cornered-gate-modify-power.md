---
id: T-0069
title: A Cornered gate on triggers, and a step that gives power until end of turn (FR-KWD-002e, FR-LAY-003, wave B shape, sonnet)
kind: shape
tier: sonnet
wave: B
status: review
branch: m2/T-0069-cornered-gate-modify-power
depends_on: []
claims: [gate:Cornered (triggers), primitive:MODIFY_STAT (power, until end of turn)]
cards: []
requirements: [FR-KWD-002e, FR-LAY-003, FR-LAY-001]
---

## Goal
Two shapes. First, a printed trigger can be gated on Cornered: `ability.Trigger.Cornered`, checked when the trigger's event happens (`FR-KWD-002e`, `FR-LAY-003`). Second, a step can give the Subject +N power until end of turn: `ability.OpModifyStat`, the design note's `ModifyStat` op, for now only power and only until end of turn (Layer 5, `FR-LAY-001`; it expires in Sub-Phase 4c, `FR-TRN-001`). The first card that uses both is *Allshan, Chainbreaker Revenant*: "Cornered: Whenever this unit attacks, it gets +3 power until end of turn." That card is T-0070. It is split from this task because together they would have six non-test files.

## Inputs
- FRD 01: `FR-KWD-002e` (line 533): Cornered is active for a player whenever their opponent controls strictly more battlefield assets (units and structures combined) in the Main Theater than that player controls there. For triggered abilities it is checked strictly when the trigger event occurs. `FR-LAY-003` (lines 429-433): Cornered is an enqueue gate, and a Cornered triggered ability that was placed resolves even if the counts change before it resolves. `FR-KWD-002f` (line 534) and `FR-CBT-002` (line 506): attack triggers fire in combat Step 2 and resolve before Step 3 damage, which is when "+N power until end of turn" matters. `FR-TRN-001` Sub-Phase 4c (line 248): "until end of turn" modifiers expire.
- Survey row: Allshan, Chainbreaker Revenant (wave B), line 147.
- Design note: `docs/architecture/m2-ability-model.md`, Key types (`Trigger` gate conditions include `Cornered`; `Step` ops include `ModifyStat`, with an optional `Duration`).
- Board directive #14 (a): Allshan is a deck card (one copy in the Path starter deck).
- Kernel today:
  - Trigger gates (`Turn`, `ToZone`, `MaxPerTurn`) are checked in `fireHook` (`trigger.go`), which runs when the event happens and buffers the trigger. `deployedTo` sits beside it as a gate helper.
  - `Asset.TempPower` (`state.go`) is the Layer 5 "until end of turn" power modifier. `Power` (`layers.go`) adds it and clamps at 0, and `endTurn` (`turn.go`) clears it in 4c. Aggressive (`StackAggressive` in `stack.go`) already adds to it.
  - Steps dispatch in `resolveAbility` (`effects.go`), one `step<Op>` per op (directive #40). `stepAssets` gives `SubjectSelf`, `SubjectTarget` and `SubjectAll`, and `SubjectSelf` fizzles once the source has left the battlefield.

### Shape (lead decisions)
- **`ability.Trigger.Cornered bool`**, after `Optional`, with the comment `// "Cornered: ...": fires only if its controller is Cornered when the event happens (FR-KWD-002e, FR-LAY-003)`.
- **`(g *Game) cornered(p PlayerID) bool`** in `trigger.go`, beside `deployedTo`. It returns true when p's opponent controls strictly more battlefield assets in `ZoneMainTheater` than p does. Count units and structures, face-down Scratch Company units included (they are units). Headquarters are never in the Main Theater, so they never count. Continuous Cornered (wave E) may move it to `layers.go` when it needs it.
- **The gate goes in `fireHook`,** for every hook: a trigger with `Cornered` set is skipped unless `g.cornered(a.Controller)`. Check it before the `MaxPerTurn` counter, so a firing the gate stopped does not use up the limit. The gate is checked only there: placement and resolution never check it again (`FR-LAY-003`).
- **`ability.OpModifyStat`**, appended after `OpDiscard`, with the comment `// the Subject gets +Amount power until end of turn (Layer 5, FR-LAY-001; expires in Sub-Phase 4c, FR-LAY-003). Other stats and durations come later`. No new `Step` field: power and until end of turn are the only stat and duration so far. When wave C adds max health and "until the end of your opponent's turn", it adds `Stat` and `Duration` fields whose zero values mean power and until end of turn. Amount may be negative.
- **`(g *Game) stepModifyStat(it, st, chosen)`** in `effects.go`, beside the other steps, with a `case ability.OpModifyStat` in `resolveAbility`. For each asset from `g.stepAssets(it, st, chosen)` it adds `g.value(it, st.Amount)` to `TempPower`. It emits nothing new: the settled snapshot already shows `TempPower` and effective `Power`, as for Aggressive.
- **No `stack.go` change.** Aggressive's `a.TempPower += it.Amount` and the new step both write the one Layer 5 field. The field, its clamp in `Power` and its expiry in `endTurn` each stay in one place. The design note says `TempPower` becomes a list of timestamped modifiers with durations. The task that does that gives both writers one helper.
- **`doc.go`:** add `OpModifyStat` to the steps line and the Cornered gate to the gates in "What runs now". Remove "Cornered" from "Not built yet", and add "continuous Cornered" in its place.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
To be written in `pkg/engine/cornered_test.go`, with synthetic cards only (Allshan is T-0070). The file won't compile until `ability.Trigger.Cornered` and `ability.OpModifyStat` exist. A convenient source is a unit with an attack trigger (`HookAttacks`) whose steps are under test. Required cases:
- **Gate open:** `TestFR_KWD_002e_CorneredTriggerFiresWhenTheOpponentHasMoreInTheMainTheater`. The opponent controls 2 Main Theater assets and the source's controller 1 (the attacker itself). The trigger fires and resolves.
- **Strictly more:** `TestFR_KWD_002e_CorneredTriggerDoesNotFireOnEqualCounts`. With equal Main Theater counts, it does not fire. No decision is asked and nothing reaches the stack.
- **Only the Main Theater counts:** `TestFR_KWD_002e_CorneredCountsOnlyMainTheaterAssets`. The opponent's extra assets are in Reserves, and the Headquarters are ignored. It does not fire.
- **Structures and Scratch Company count:** `TestFR_KWD_002e_CorneredCountsStructuresAndScratchCompanyUnits`. The opponent's majority is made up of a structure and a face-down Scratch Company unit (`putScratch`), and it fires.
- **Checked when the event happens (`FR-LAY-003`):** `TestFR_LAY_003_PlacedCorneredTriggerResolvesAfterTheCountsChange`. The source has two attack triggers: a Cornered one with `OpModifyStat` +N on `SubjectSelf`, and a non-Cornered one that deals lethal damage to an opposing Main Theater unit. Order them, through the scripted `DecideOrder` answer, so that the damage resolves first. Once it does, the controller is no longer Cornered, yet the Cornered trigger still resolves and the power is gained.
- **Gate before the limit:** `TestFR_KWD_002e_ACorneredFiringTheGateStoppedDoesNotUseUpMaxPerTurn`. A Cornered damage-taken trigger (`HookDamaged`) with `MaxPerTurn: 1` fails the gate on the first damage. The counts then change in the same turn (an opposing unit moves into the Main Theater), and the second damage fires it.
- **Power until end of turn:** `TestFR_LAY_001_ModifyStatGivesPowerThatCountsInCombatDamage`. An attacker with an attack trigger "+3 power" on `SubjectSelf` deals its printed power + 3 to the defender in Step 3.
- **Expiry (`FR-TRN-001` 4c):** `TestFR_LAY_003_ModifyStatPowerExpiresAtEndOfTurn`. After `endTurn`, `Power` is back to printed.
- **Adds to Aggressive:** `TestFR_LAY_001_ModifyStatStacksWithAggressive`. An Aggressive 1 attacker with a "+3 power" attack trigger has printed power + 4 at Step 3.
- **Fizzles off the battlefield:** `TestFR_TGT_003_ModifyStatOnSelfFizzlesWhenTheSourceHasLeft`. If the source is gone when the step resolves, nothing changes and nothing panics.
- **Negative amount and the clamp:** `TestFR_LAY_001_NegativeModifyStatClampsPowerAtZero`. A "-5 power" step on a 2-power target leaves `Power` at 0. Use a chosen target, `SubjectTarget`.
- **Left out unless FRD 01 fixes it:** Cornered continuous effects and their lock during simultaneous steps (wave E), and durations other than until end of turn (wave C).

## Helpers Needed
None expected. `put`, `putScratch`, `moveTo`, `attack`, `endTurn`, `script` and `wantDamage` already exist in `scenario_test.go`. If a helper is needed, request it here and the engineer adds it.

## Allowed Files
- `pkg/engine/ability/ability.go` (`Trigger.Cornered`, `OpModifyStat`)
- `pkg/engine/trigger.go` (`cornered` and the gate in `fireHook`)
- `pkg/engine/effects.go` (`stepModifyStat` and its `resolveAbility` case)
- `pkg/engine/doc.go` (the gates, steps and "Not built yet" lines)

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run 'TestFR_KWD_002e_Cornered|TestFR_LAY_003_(PlacedCornered|ModifyStat)|TestFR_LAY_001_(ModifyStat|NegativeModifyStat)|TestFR_TGT_003_ModifyStat'`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] No card uses either shape yet, so the sims are unchanged: `go tool task sim-compare -- autoloop/m2-wave-b` passes

## Log
- 2026-10-26 lead-architect (wave run B, session 3): wrote the task. Run branch in sync with `main`; no `open: before` directive. The wave B deck cards still not executable are Allshan, Chainbreaker Revenant; Eternal Watchman; Infested Apostle; and Arkha, Sun's Fury (directive #14 (a)). Allshan is first in survey order. Directives checked: #15 is not triggered, because a power modifier can't make an asset lethal when it expires in 4c (only Max Health modifiers can). #51 (no "if you do"), #58 (no exhaust), #59 (no pile pick) and #8 (no `deck.go`) don't apply. #20 waits for a wave C shape task, and this is wave B. Claims don't overlap any open task, since none is open. There's no existing equivalent: `pkg/engine` has no Cornered check, and the only "+N power until end of turn" is the Aggressive keyword item, which is not an effect step. The card task, T-0070, is written and `queued`. Status `tests`; dispatching `rules-test-author`.
- 2026-10-26 rules-test-author: wrote `pkg/engine/cornered_test.go`, 11 tests, one per required case: `TestFR_KWD_002e_CorneredTriggerFiresWhenTheOpponentHasMoreInTheMainTheater`, `..._CorneredTriggerDoesNotFireOnEqualCounts`, `..._CorneredCountsOnlyMainTheaterAssets`, `..._CorneredCountsStructuresAndScratchCompanyUnits`, `..._ACorneredFiringTheGateStoppedDoesNotUseUpMaxPerTurn`, `TestFR_LAY_003_PlacedCorneredTriggerResolvesAfterTheCountsChange`, `TestFR_LAY_003_ModifyStatPowerExpiresAtEndOfTurn`, `TestFR_LAY_001_ModifyStatGivesPowerThatCountsInCombatDamage`, `..._ModifyStatStacksWithAggressive`, `..._NegativeModifyStatClampsPowerAtZero`, `TestFR_TGT_003_ModifyStatOnSelfFizzlesWhenTheSourceHasLeft`. Helpers requested: none (the file defines its own `cornered`, `modifyPower`, `modifyTarget`, `targetedAttack`, `wall`, and reuses `limited`, `striker`, `attackTrigger`, `placedOrder` from other test files). `go vet` fails only on the missing `Trigger.Cornered`; `cardtown lint rules` passes. Notes: (1) the Scratch Company case moves `putScratch`'s Reserves unit into the Main Theater with `moveTo`, because `putScratch` places it in Reserves. (2) The gate-before-limit case moves the opponent's ally out of the Main Theater (2 v 2 becomes 2 v 1) rather than moving an attacker in; same count change. (3) The FR-LAY-003 and fizzle cases order triggers with scripted `DecideOrder` picks `[0, 1]` (first pick is placed first and resolves last). Left out as the task says: continuous Cornered, other durations/stats.
- 2026-10-26 lead-architect (tests-review, round 1): **Approved, `ready`.** Checked against `FR-KWD-002e` (line 533), `FR-LAY-003` (lines 429-433), `FR-LAY-001` and `FR-TRN-001` 4c. The expected outcomes are right:
  - strictly more is required, and equal counts don't fire;
  - only the Main Theater counts, so Reserves, Headquarters and a Scratch Company unit in Reserves are ignored;
  - a structure and a face-down Scratch Company unit in the Main Theater both count;
  - the Cornered player is the trigger's controller, not the active player (the max-per-turn case gates player 1 during player 0's turn).

  Two tests would catch the plausible wrong builds:
  - `PlacedCorneredTriggerResolvesAfterTheCountsChange` fails if the gate is checked again at resolution.
  - `ACorneredFiringTheGateStoppedDoesNotUseUpMaxPerTurn` fails if the counter runs before the gate.

  The setup differences the author logged keep each case's meaning. `DecideOrder` picks `[0, 1]` put the first listed trigger on the stack first, so it resolves last (`SVY-R7`). Every helper used exists (`striker`, `attackTrigger`, `limited`, `unbowed`, `bowed`, `placedOrder`, `tgt`, `dmgTo`, `scratchN`, `wantScratches`). No earlier test is modified. The file fails only because `Trigger.Cornered` and `OpModifyStat` don't exist yet. Next: dispatch `kernel-engineer` (tier `sonnet`). The Shape section's lead decisions are binding.
- 2026-10-26 kernel-engineer (sonnet): implemented. Diff: `ability.Trigger.Cornered` and `ability.OpModifyStat` (ability.go); `(g *Game) cornered` (counts `g.Theater` assets by controller; HQs and Reserves are not in it) and the gate in `fireHook`, placed before the `MaxPerTurn` counter (trigger.go); `stepModifyStat` plus its `resolveAbility` case, adding `g.value(it, st.Amount)` to `TempPower` via `stepAssets` (effects.go); doc.go steps/gates/"Not built yet" lines. Reuse: no existing Cornered check or power-step found; `TempPower`, `Power` clamp and `endTurn` expiry reused unchanged, no stack.go change. Ran: the listed `go test -run` filter (ok), `go tool task check` (all pass), `go tool task sim` (clean), `go tool task sim-compare -- autoloop/m2-wave-b` (PASS, byte for byte). No tests or helpers edited.
