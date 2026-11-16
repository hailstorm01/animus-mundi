---
id: T-0251
title: "Maximum Health modifiers in Layer 5, and the death sweep after Sub-Phase 4c (#15 (3); wave H shape, sonnet)"
kind: shape
tier: sonnet
wave: H
status: blocked
branch: m2/autoloop-wave-h/T-0251-max-health-modifier-and-4c-sweep
depends_on: []
claims: [shape:Layer 5 Maximum Health modifier (Modifier.MaxHealth, summed by Game.MaxHealth, unclamped), shape:End Phase 4c death sweep with its flush and resolve cycle (#15 (3))]
cards: []
requirements: [FR-LAY-001, FR-LAY-002, FR-LAY-003, FR-TRN-001]
---

## Goal
Effective Maximum Health goes through the layer pipeline: a Layer 5 `Modifier` can change Maximum Health as well as power (`FR-LAY-002` Temporary Maximum Health Modifiers: it lowers Effective Maximum Health without adding damage, and its expiry in Sub-Phase 4c restores it). Maximum Health never clamps (`FR-LAY-001` Exclusive Combat Power Clamping), so a reduction to 0 or below makes the asset lethal with no damage (`FR-LAY-002` Current Health & Lethal Destruction).

A max-health modifier is the first modifier whose expiry in 4c can leave an asset lethal (a damaged asset loses a bonus), so this task also applies directive #15 (3): End Phase Sub-Phase 4c is followed by a state-based death sweep, and the triggers that sweep fires are flushed and resolved, cascades included, before the Sub-Phase 4d handover (`FR-LAY-002` State-Based Death Sweep Timing: nothing is resolving then; #15: "give each boundary the same sweep, flush and resolve cycle that `resolveStack` uses"). A loss found by that sweep ends the match before 4d.

No step makes a max-health modifier yet: `OpModifyStat` on Maximum Health is T-0252, and *Poisoned Rations*, the card, is T-0253 (both split off under the D2 size rule: the shape and the card together touch eight non-test files).

## Inputs
- FRD 01 (`docs/frd/01_Game_Rules_FRD.md`): `FR-LAY-002` (line 506): Current Health & Lethal Destruction (`:507`, which names *Poisoned Rations*), State-Based Death Sweep Timing (`:508`), Temporary Maximum Health Modifiers (`:509`); `FR-LAY-001` (line 500), Exclusive Combat Power Clamping (`:503`: Maximum Health and Current Health do not clamp); `FR-LAY-003` (line 512, durations); `FR-TRN-001` (line 301), Sub-Phase 4c and 4d (`:317`, `:318`).
- Board directive #15 (3) (`.agents/tasks/m2/BOARD.md`): "the sweep after 4c: waiting: the first task that adds a modifier whose expiry in 4c can leave an asset lethal (a max-health modifier, as *Poisoned Rations*' task adds)". This is that task. #15 (2) (draw and end-of-turn hooks) stays `waiting`.
- Design note Deliverable 6m, "From wave C" (`docs/architecture/m2-ability-model.md:207`): max-health modifiers go through `layers.go`'s `Modifier`.
- Code of the day (run branch at 0647f4f):
  - `Modifier` (`pkg/engine/layers.go:17-23`), `TempPower` (`:26`), `expireAtEndOf` (`:40`), `MaxHealth` (`:56-59`, printed Health only), `CurrentHealth` (`:62`), `IsLethal` (`:104`), `effectiveState` (`:117`). `MaxHealth` is the only reader of printed Health in `pkg/engine` (searched `.Health` in non-test files), so `CurrentHealth`, `IsLethal`, `EffectiveState` and the snapshot (`pkg/engine/snapshot.go:64`) all follow it.
  - `endTurn` (`pkg/engine/turn.go:35-50`): 4c resets pools and calls `expireAtEndOf`, then goes straight to 4d. `beginTurn` (`:6-32`) is the pattern for a boundary: `fireTurnStart`, `resolveStack`, return if `g.Outcome.Over`; and the Draw Phase's `g.deathSweep()` then the `Outcome.Over` check.
  - `resolveStack` (`pkg/engine/stack.go:114-124`) flushes buffered triggers first, then resolves, sweeping and flushing after each item. `deathSweep` (`pkg/engine/sweep.go:49`) buffers death triggers (`FR-TRG-004`) and applies the match-loss checks.

## Shape (the lead's plan)
- `pkg/engine/layers.go`:
  - `Modifier` gains `MaxHealth int`: a Layer 5 modifier adds it to Maximum Health (`FR-LAY-002`). Its comment says so beside `Power`'s.
  - `MaxHealth(a)` returns printed Health plus the `MaxHealth` of every Layer 5 modifier on `a`, unclamped (`FR-LAY-001`, `FR-LAY-002`). Damage is untouched.
  - Update the file's header comment so it no longer implies only power lives in Layer 5.
  - `expireAtEndOf` is unchanged: a max-health modifier expires by its `Until` like any other.
- `pkg/engine/turn.go` (`endTurn`): after 4c's pool reset and expiry, `g.deathSweep()`, then `g.resolveStack()` (it flushes the sweep's death triggers and resolves them and their cascades, sweeping after each), and return before 4d if `g.Outcome.Over` after either. Replace the comment so 4c names the sweep and #15 (3). The Phase stays `PhaseEnd` throughout. Don't add a 4a/4b trigger step (#15 (2) waits for an end-of-turn hook).
- No new `Op`, no event kind, no change in package `ability`.

## Tests (shape tasks: written by rules-test-author; engineers never edit these)
New tests go in `pkg/engine/max_health_test.go`. Cards are synthetic (`unit`, `hq` in `scenario_test.go`); place them with `s.put`. Since no step makes a max-health modifier yet, give it directly: `a.Modifiers = append(a.Modifiers, Modifier{Layer: 5, MaxHealth: n})` (the zero `Until` is until end of turn), and set damage with `a.Damage = n`. Player 0 is the active player; `s.endTurn()` ends its turn. Read stats with `s.g.MaxHealth`, `s.g.CurrentHealth`, `s.g.IsLethal`; use `s.alive`.

- `TestFR_LAY_002_MaxHealthModifierChangesEffectiveMaxHealth`: a unit with printed Health 5 and a `MaxHealth: -3` modifier has Maximum Health 2 and Current Health 2; with 1 damage, Current Health 1. Its `Damage` is unchanged by the modifier (no damage counters, `FR-LAY-002`). A second `MaxHealth: +1` modifier on it makes 3 (modifiers add). Its power is its printed power (the modifier changes no other stat), and `s.g.Snapshot()`'s entry for it (or `effectiveState`, whichever the test reaches) reports the effective Maximum Health.
- `TestFR_LAY_001_MaxHealthDoesNotClamp`: printed Health 2 with a `MaxHealth: -3` modifier: Maximum Health is -1, not 0, and Current Health is -1.
- `TestFR_LAY_002_MaxHealthAtZeroIsLethalWithoutDamage`: printed Health 3, no damage, a `MaxHealth: -3` modifier: `IsLethal` is true. A death sweep (any stack item resolving; for example player 0 deploys a free synthetic unit, or a free Maneuver with an unrelated step) destroys it. The same unit with a `MaxHealth: -2` modifier and no damage survives the sweep.
- `TestFR_LAY_002_MaxHealthReductionExpiresAndRestores` (Temporary Maximum Health Modifiers): printed Health 5, 2 damage, a `MaxHealth: -2` modifier: Current Health 1, it survives a sweep. After `s.endTurn()`, it is alive with Maximum Health 5, Current Health 3 and `Damage` still 2.
- `TestFR_TRN_001_ExpiryAtFourCDestroysInTheEndPhase` (#15 (3), `FR-LAY-002` sweep timing): player 0's unit with printed Health 2, a `MaxHealth: +2` modifier and 3 damage is alive (Current Health 1). After `s.endTurn()` it is destroyed, and the log's `EventDestroy` event for it comes before the `EventTurnStart` event of the next turn (read `s.g.Log` from an index taken before `s.endTurn()`). Do the same for a player 1 unit (the expiry and sweep don't depend on whose asset it is).
- `TestFR_TRN_001_FourCDeathTriggersResolveBeforeHandover` (#15 (3): the sweep's triggers flush and resolve before 4d): player 0's unit as in the previous test also prints a `HookKilled` trigger that draws 1 (`mortal`, `killedTrigger`, `drawN` in `death_trigger_test.go:17-26`, `trigger_test.go:17`). After `s.endTurn()` player 0's hand has one more card, and that draw's event comes before the next turn's `EventTurnStart`; the stack is empty.
- `TestFR_TRN_001_FourCSweepLossEndsTheMatch` (`FR-LAY-002`, `FR-CBT-003`): player 1's HQ (`hq`) with a `MaxHealth: +2` modifier and damage equal to its printed Health is alive; after `s.endTurn()` the match is over with player 0 the winner, and no next turn began (no `EventTurnStart` after the end, and `s.g.Turn` didn't advance, or whichever of the two the engine shows; check what `Outcome` and the log give and assert the observable one).
- `TestFR_TRN_001_NoFourCDeathWithoutExpiry`: a guard that the new sweep doesn't change ordinary turns: a damaged unit that is not lethal and has no modifier is still alive after `s.endTurn()`, with its damage unchanged.

Existing tests stay as they are and must still pass.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- none expected. If the test author needs one, it names it here.

## Allowed Files
- `pkg/engine/layers.go`
- `pkg/engine/turn.go`

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run "MaxHealth|FourC"`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-compare -- autoloop/m2-wave-h` passes (no card makes a max-health modifier yet, so nothing can be lethal at 4c that wasn't before, and the 4c sweep finds nothing in a sim)

## Log
- 2026-11-15 lead-architect (wave H run, session 12): *Poisoned Rations* is the next wave H card (wave C's, after *Blindside*). Its survey row (`docs/design/m2_starter_set_ability_survey.md:197`) needs a max-health modifier, which the kernel lacks: `MaxHealth` reads printed Health only (`layers.go:57`), `Modifier` has no health field and `OpModifyStat` changes power only. Searched `pkg/engine` and the task archive for an existing health modifier: none (`git grep -n "MaxHealth"` finds only the reader, `IsLethal`, `EffectiveState` and the snapshot). That fires #15 (3), set `open` on the board in this commit and applied here. The whole shape (layer, 4c sweep, the step with `ability.StatMaxHealth` read by filters too, its event kind) touches six non-test files, and the card two more, so split under the D2 size rule: this task (the layer and the sweep), T-0252 (`OpModifyStat` on Maximum Health, `queued`), T-0253 (*Poisoned Rations*, `queued`). No ruling needed: `FR-LAY-002` settles the layer, the unclamped value and the restore on expiry; #15 (user-approved audit) settles where the 4c sweep's triggers resolve. Re-read `BOARD.md` before committing: no directive added since. Status `tests`.
- 2026-11-15 rules-test-author: tests written in `pkg/engine/max_health_test.go` (local helpers `addMaxHealth`, `sweepNow`, `firstEventAfter`, `destroyIndexFor` live in that file): TestFR_LAY_002_MaxHealthModifierChangesEffectiveMaxHealth, TestFR_LAY_001_MaxHealthDoesNotClamp, TestFR_LAY_002_MaxHealthAtZeroIsLethalWithoutDamage, TestFR_LAY_002_MaxHealthReductionExpiresAndRestores, TestFR_TRN_001_ExpiryAtFourCDestroysInTheEndPhase (both players, as subtests), TestFR_TRN_001_FourCDeathTriggersResolveBeforeHandover, TestFR_TRN_001_FourCSweepLossEndsTheMatch, TestFR_TRN_001_NoFourCDeathWithoutExpiry. Helpers requested for scenario_test.go: none. `go vet` fails only on the missing `Modifier.MaxHealth` field (the shape itself). Loss test asserts no turn-start event after the end and `Turn` unchanged. Destroy events are matched by asset ID or card name. `cardtown lint rules` passes. No cases left out.
- 2026-11-15 lead-architect (D4): tests approved. Each checks what `FR-LAY-002` and `FR-LAY-001` give: a Layer 5 max-health modifier lowers or raises Maximum Health without damage counters, modifiers add, power is untouched, the snapshot reports the effective value; Maximum and Current Health go negative unclamped; Maximum Health 0 is lethal with no damage and the next sweep destroys it, while 1 survives; a reduction's expiry in 4c restores Maximum and Current Health with damage unchanged. For #15 (3): a damaged asset whose bonus expires in 4c is destroyed before the next `turn-start` (both players' assets), its death trigger resolves before the handover with the stack empty, an HQ made lethal there ends the match with no next turn, and an ordinary damaged asset is untouched. Each 4c case fails on today's code for the right reason: the asset now dies only at the next turn's Draw Phase sweep, after `turn-start`, and its trigger stays buffered. The file fails to compile only on `Modifier.MaxHealth`. Status `ready`, dispatched to `kernel-engineer` (sonnet).
- 2026-11-15 kernel-engineer (sonnet): status `blocked`. Implemented `Modifier.MaxHealth` and the summed, unclamped `Game.MaxHealth` in `layers.go` (header and Modifier comments updated), and in `turn.go` `endTurn` a `deathSweep`, then `resolveStack`, after 4c with an `Outcome.Over` return after each, before 4d. Reuse: `deathSweep` and `resolveStack` as in `beginTurn`; no new shape. `go test ./pkg/engine` passes except `TestFR_TRN_001_FourCDeathTriggersResolveBeforeHandover` (fails only at `max_health_test.go:197`, "player 0's draw is logged"); `go vet ./pkg/engine` is clean. Why blocked (test seems to contradict the engine): the test looks for an `EventDraw` for player 0 between the end and the next `turn-start`. `g.draw` (`zones.go:26`) logs `EventDraw` only for a draw from an empty deck; a successful draw (`moveTopToHand`) logs no event at all. A debug dump of the log after `endTurn` shows `action, destroy, place, turn-start`: the trigger did resolve (the hand grew by 1 and the stack is empty, both asserted earlier in the test, which passed), but its draw leaves no event. Making it pass needs a draw event in `zones.go` (outside `Allowed Files`, and it would change every sim log and the event-kind set), or the test should check the hand count and the `destroy`/`place` ordering instead. Not run: `task check`, sim, sim-compare (the check cannot pass until this is settled). Lead: decide between adding a successful-draw event (a design call) and having rules-test-author relax that assertion.
