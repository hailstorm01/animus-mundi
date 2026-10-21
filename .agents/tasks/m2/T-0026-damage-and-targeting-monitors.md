---
id: T-0026
title: Damage-stage and stack-placement event facts, with the damage-sequence and targeting monitors
kind: harness
tier: sonnet
wave: none (board directive #23, part 2 of 2)
status: ready
branch: m2/T-0026-damage-and-targeting-monitors
depends_on: [T-0025]
claims: [event:place, field:Event.Raw, field:Event.Combat, field:Event.Armor, field:Event.Absorbed, field:combat-event.Source-Target, monitor:damage-sequence, monitor:targeting]
cards: []
requirements: [FR-REP-004, FR-CBT-002, FR-KWD-002b, FR-KWD-002c, FR-KWD-002k, FR-TGT-001, FR-KWD-002h]
---

## Goal
Two more rule monitors run in every sim game:
- **`damage-sequence`:** every damage instance follows the 5-stage mitigation sequence, and combat damage uses each combatant's effective power.
- **`targeting`:** every target is chosen while its item is being placed on the stack, never while an item resolves, and an opponent's Ward asset is never offered.

Today's events can't support either check. A damage event carries only the committed amount, and nothing marks the moment an item goes on the stack. So this task first adds those facts to the events, additively under ADR-0005.

## Inputs
- Board directive #23, part 2 (`.agents/tasks/m2/BOARD.md`). T-0025 built the framework (`pkg/engine/sim/monitor.go`): `Monitors`, the `monitorRules` table, `Violation`, `Coverage`, pending checks resolved at the next settled snapshot, and lookup of printed characteristics by owner and card name, with the Scratch Company face for face-down assets.
- ADR-0005 (`docs/architecture/decisions/0005-structured-game-events-and-observer.md`): fields are added additively. Its Consequences already ask for an `ability` event "when the metrics work starts". The `place` event below covers placement for every item kind, triggered abilities included.
- **Engine today:**
  - `dealDamage` (`pkg/engine/damage.go`) computes a `DamageResult` with `Raw`, `AfterRes`, `AfterCap`, `Absorbed` and `Committed`. It logs a `damage` event with `Amount` = `Committed`, and the raw amount, armor and stages only in `Detail`.
  - `push` (`pkg/engine/stack.go`) is the single place an item goes on the stack. `flushTriggers` (`trigger.go`) calls `chooseTargets` (one `Target` decision per slot, recorded as a `decision` event) before it pushes the item.
  - `resolveAttack` logs the `combat` event with `logf`, so that event has no Source or Target.
- **Rules:**
  - `FR-REP-004`, Stages 2 to 5: Resistant subtracts and negative Resistant amplifies, Stalwart caps combat damage at 1 and skips non-combat damage, and Armor absorbs 1-for-1.
  - `FR-CBT-002` Step 3: both combatants deal combat damage equal to their effective power at the same moment. Berserk cleave is non-combat damage.
  - `FR-TGT-001` with `SVY-R3`: targets are locked when the item is placed, and no target is chosen while an item resolves.
  - `FR-KWD-002h`: Ward makes an asset an illegal target for its controller's opponents.
  - Keywords: `FR-KWD-002b` Resistant, `FR-KWD-002c` Stalwart, `FR-KWD-002k` Armor.
- Rulings that apply: none beyond those cited.

### Shapes
```go
// pkg/engine/state.go: Event gains (damage events only)
	Raw      int  `json:"raw,omitempty"`      // the incoming damage before Stage 2
	Combat   bool `json:"combat,omitempty"`   // combat damage (Stalwart applies)
	Armor    int  `json:"armor,omitempty"`    // the target's Armor pool before this damage
	Absorbed int  `json:"absorbed,omitempty"` // Stage 4: damage absorbed by Armor
// Amount stays the damage committed to Health (Stage 5).

// New event kind "place", emitted by push for every item, right after it is
// appended to the stack:
//   Player     the item's controller
//   Source     the item's Source (attacker, marching unit, scratching HQ, ability or Aggressive source)
//   Target, Secondary   an Attack's primary and Berserk secondary targets
//   Card, Inst the card, for a Deploy item and for an ability item
//   Amount     the stack depth after the push
//   Detail     "player P places <what> on the stack", where <what> names the item kind

// The existing "combat" event gains Source = attacker and Target = defender.
// Its Kind, Detail and Turn are unchanged.
```

### Lead decisions (technical)
- **Additive only.** Every existing event keeps its `Kind`, `Detail` and `Turn` (`TestEvents_Case9_ExistingEventsUnchanged` pins this). New fields use `omitempty`, so an event that doesn't use them serializes as before. No new pointer fields, so `Event.clone` doesn't change. The replay divergence check (`eventlog`) doesn't change either: `Detail` already carries the damage stages, and the `place` events are compared field by field, like any other event.
- **Game play doesn't change.** Adding events changes no state and touches no random source. The sim summary is byte-identical to `rework/executable-rules`, and `-events` files differ only by the new `place` lines (with Seq renumbered) and the new keys.
- **Monitor `damage-sequence`** (`FR-REP-004`, `FR-CBT-002`, `FR-KWD-002b`, `FR-KWD-002c`, `FR-KWD-002k`). On every `damage` event, using the target's printed characteristics, as in T-0025:
  - Let `r = max(0, Raw − Resistant)`. Then `c = min(r, 1)` if the event is `Combat` and the target has Stalwart, and `c = r` otherwise. The check is that `Absorbed == min(c, Armor)` and `Amount == c − Absorbed`.
  - Every damage event counts `FR-REP-004`. `FR-KWD-002b` is counted when Resistant ≠ 0, `FR-KWD-002c` when the target has Stalwart, and `FR-KWD-002k` when `Armor` > 0.
  - **The armor pool is real:** `Armor` equals the target's Armor in the most recent settled snapshot, minus the `Absorbed` of earlier damage events on that target since that snapshot. If the target isn't in that snapshot, skip this part.
  - **Combat pairs (`FR-CBT-002`):** for each `combat` event, the next two `damage` events are `Combat`. Both of them use the combat event's attacker and defender, in either order. One is attacker to defender with `Raw` equal to the attacker's `Power`, and the other is defender to attacker with the defender's `Power`, both from the most recent settled snapshot. Every other damage event is not `Combat`. `FR-CBT-002` is counted once per combat event.
  - The target's characteristics come from the most recent settled snapshot. If the target isn't there, look it up by the next settled snapshot instead, and skip the check if it's in neither.
- **Monitor `targeting`** (`FR-TGT-001`, `SVY-R3`, `FR-KWD-002h`). On every `decision` event whose decision kind is `Target`:
  - The next event that isn't a `decision` must be a `place` event with the same `Player`, and no event of any other kind may come between them. A target chosen while an item resolves would be followed by that item's effects instead. `FR-TGT-001` and `SVY-R3` are counted on every Target decision.
  - No offered option ref is an asset controlled by the decider's opponent that has Ward (printed, looked up in the most recent settled snapshot; skip refs that aren't there). `FR-KWD-002h` is counted when the opponent has a Ward asset on the battlefield in that snapshot.
  - If the game ends after a Target decision with no event of another kind following it, the check is dropped.
- **Registration:** both monitors join `monitorRules` and `NewMonitors`, so `SimulateWith`, `cardtown sim` and the keyword-deck playout test run them with no further change. The coverage table picks them up automatically.
- **ADR-0005 addendum** (one short paragraph under Decision, plus an edit to the Consequences bullet about the `ability` event):
  - It records the four damage fields, the `place` event and the combat event's Source and Target.
  - The bullet changes to say that placement is now recorded for every item, while resolving a triggered ability still emits no event of its own.
- **File-count exception (lead):** this task has five non-test files, one over the rule of four. The fifth is the ADR addendum, a few lines long. Splitting the event facts from their only readers would merge fields with nothing checking them, so the lead keeps them together.
- **Not in this task:**
  - the APNAP monitor (`FR-TRG-003`) and the illegal-target monitor (`FR-TGT-003`), which the `place` event makes possible, but which starter cards don't reach yet;
  - events for Heal and AddArmor (none of today's cards changes armor during the same resolution as damage; if one ever does, the armor-pool check flags it, and the lead routes an event for it);
  - an `ability` resolution event;
  - Berserk secondary-target legality.

## Acceptance Cases (harness tasks only; written by the lead)
Cases 1 to 4 go in new test files: `pkg/engine/event_facts_test.go` (package `engine`, using the existing scenario helpers; don't change `scenario_test.go`) and `pkg/engine/sim/monitor_damage_test.go`. The one authorized edit in case 5 goes in the same first `test(T-0026): …` commit, which must fail before the implementation. Cases 6 and 7 are CLI checks, run by the engineer and re-run by the reviewer. Record their output in the Log.
1. **Damage fields** (scenes, `event_facts_test.go`):
   - An attacker with power 3 attacks a Stalwart defender with Armor 2. The defender's damage event has `Raw` 3, `Combat` true, `Armor` 2, `Absorbed` 1 and `Amount` 0. The retaliation event has the defender's power as `Raw`.
   - A Resistant −1 target takes 0 raw combat damage: `Amount` is 1.
   - A non-combat damage instance (a `DealDamage` ability, as in the T-0017 tests) on a Stalwart target has `Combat` false and is not capped.
   - The `combat` event has the attacker as `Source` and the defender as `Target`, and its `Detail` is unchanged.
2. **`place` events** (scenes):
   - A Deploy gives `action`, then a `place` with the card, its instance and `Amount` 1, then `deploy`.
   - A deployed unit with a two-slot targeted deploy trigger gives its two `Target` decisions, then one `place` for the ability, with the controller as `Player` and the source asset as `Source`.
   - An attack with an Aggressive attacker gives a `place` for the Attack, then one for the Aggressive trigger with `Amount` 2.
   - Across 20 starter games (`SimulateWith` with a collecting observer), check every Deploy, March, Attack and Scratch `action` event. The next event that isn't a `decision` is a `place` with the same `Player`, and with the action's `Inst` (Deploy) or `Source` (the other three).
3. **Each new monitor can fail.** Hand-built sequences go to `NewMonitors`, as in T-0025 case 1. Each gives a `Violation` with the right monitor, all its rule IDs, the game, the seed and the breaking `Seq`:
   - **damage-sequence:** `Amount` that ignores Resistant; Stalwart applied to non-combat damage; Stalwart not applied to combat damage; Armor that absorbs nothing; an `Armor` that doesn't match the settled snapshot; a combat damage `Raw` different from the attacker's power; a third `Combat` damage event after one `combat`.
   - **targeting:** a Target decision followed by a `damage` event; a Target decision followed by a `place` from the other player; an opponent's Ward asset offered as an option.

   A valid sequence of each kind gives no violation, and counts above 0.
4. **Clean on real games:** `SimulateWith` on the starter decks, 500 games under each policy, and 200 default stress-deck games under each policy, return no error.
5. **Authorized edit to an existing test:** `TestMonitor_Case4_Coverage` in `pkg/engine/sim/monitor_test.go`. Its `want` map gains `"damage-sequence"` and `"targeting"` with their rule IDs (declared in the new test file), and its "above 0 on the starter decks" list gains `damage-sequence/FR-REP-004`, `damage-sequence/FR-CBT-002`, `targeting/FR-TGT-001` and `targeting/SVY-R3`. Nothing else in that test changes.
6. **CLI, game play unchanged:**
   - For each of: `sim -deck starter -p0 ANIMUS -p1 PATH` and bare `sim`, under each `-policy`, with `-monitors=false`, the output is byte-identical to `rework/executable-rules`.
   - With monitors on, the output is the baseline's plus the new coverage rows, and it is byte-identical across two runs.
7. **CLI, events additive:**
   - For `-games 20 -events`, take the sequence of `"kind"` values per line in the new file, with every `place` line removed. It is identical to the baseline file's sequence.
   - The new file is byte-identical across two runs.
   - `-log -seed 7` shows the `place` lines.
8. `go tool task check` passes.

If a monitor reports a violation in case 4, 6 or 7, it may be a real engine bug. Don't fix the engine and don't loosen the monitor. Log the violation and its `ReplayCommand`, check that the replay reaches it, and stop at `blocked`. If the monitor's check was wrong against the FRD 01 text cited here, fix the monitor and say so in the Log. If an existing test fails only because it counts or indexes events and now sees `place` events, stop at `blocked` too, naming the test.

## Allowed Files
- `pkg/engine/state.go` (the `Event` fields only)
- `pkg/engine/damage.go`
- `pkg/engine/stack.go`
- `pkg/engine/sim/monitor.go`
- `docs/architecture/decisions/0005-structured-game-events-and-observer.md` (the addendum; file-count exception above)
- `pkg/engine/event_facts_test.go` and `pkg/engine/sim/monitor_damage_test.go` (new, the `test(T-0026)` commit)
- `pkg/engine/sim/monitor_test.go` (only the case 5 edit, in the `test(T-0026)` commit)

## Done When
- [ ] Cases 1 to 5 have tests, committed before the implementation, and all pass
- [ ] Cases 6 and 7 hold, with their output and the new starter coverage rows in the Log
- [ ] `go tool task check` passes
- [ ] Existing tests pass, unchanged apart from the authorized edit

## Log
- 2026-10-21 lead-architect: Wrote the task from board directive #23 (part 2 of 2), with status `ready` (harness, sonnet). It has one recorded file-count exception, the ADR-0005 addendum.
