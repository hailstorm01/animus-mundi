---
id: T-0245
title: A start-of-turn trigger that resolves in the Ready Phase, before the Draw Phase (wave H shape, sonnet; Hammerhand Bivouac's hook, #15 (1))
kind: shape
tier: sonnet
wave: H
status: tests-review
branch: m2/autoloop-wave-h/T-0245-turn-start-hook
depends_on: []
claims: [hook:turn-start (ability.HookTurnStart; fireTurnStart fires it for every battlefield asset), shape:start-of-turn trigger resolution (beginTurn places and resolves the Ready Phase's triggers before the Draw Phase, directive #15 (1))]
cards: []
requirements: [FR-TRN-001, FR-TRG-003, FR-TRG-004]
---

## Goal
*Hammerhand Bivouac* ("At the start of your turn, distribute 2 armor among allied units in the same zone.") needs a trigger that fires when a turn begins, and the turn needs to resolve it before the Draw Phase. Today no hook watches a turn start ([ability.go:141-149](../../../pkg/engine/ability/ability.go)), and `beginTurn` readies, then draws, with no flush or resolution between ([turn.go:6-25](../../../pkg/engine/turn.go)), which is the gap waiting directive #15 names. This task builds the hook and the Ready Phase's resolution cycle. The Bivouac's pool and the card wait for the ruling question [`2026-11-15-wave-h-bivouac-untargeted-pool`](../../../docs/design/outcomes/2026-11-15-wave-h-bivouac-untargeted-pool.md) (design note Open Question 24, on `main` at f8b574d), which doesn't touch the hook.

Covers:
- `ability.HookTurnStart`, and `fireTurnStart`, which fires it.
- `beginTurn`: start-of-turn triggers are placed in APNAP order and resolve completely, cascades included, before the Draw Phase (`FR-TRN-001` 1, directive #15 (1)).
- Not this task: the distributed armor pool, the zone locked on placement (#141 (3)) and the card (they wait for the ruling); #15's draw-trigger and End Phase parts (4a/4b triggers, the sweep after 4c), which stay `waiting` on #15's row because no hook fires there and no modifier's expiry can make an asset lethal yet.

## Inputs
- FRD 01 v3.27: `FR-TRN-001` (lines 298–313; line 304: "All *"At the start of your turn"* triggered abilities (*Hammerhand Bivouac*) enqueue onto the Resolution Stack in APNAP order (`FR-TRG-003`) and resolve completely before the turn advances to the Draw Phase", after the active player's assets ready and once-per-turn limits reset, line 303); the change log at line 138 (they trigger and resolve during the Ready Phase, after controlled assets unexhaust); `FR-TRG-003` (APNAP: the non-active player's triggers are placed on top and resolve first); `FR-TRG-004` (cascading triggers, death sweep after each item).
- Survey: `docs/design/m2_starter_set_ability_survey.md` line 183 (*Hammerhand Bivouac*: "turn start → distribute 2 armor …", `FR-TGT-002`, `FR-TRN-001`) and the wave B row (line 30, "turn-start … hooks").
- Rulings that apply: `SVY-R7` (default order of one player's simultaneous triggers).
- Design note: `docs/architecture/m2-ability-model.md`, Deliverable 6m ("From wave B": "The turn-start hook fires waiting directive #15") and Open Question 24.
- Board directive #15 (this task applies its part (1); its trigger, "the first task that adds a turn-start, turn-end or draw hook", has happened with this task).
- Code: `beginTurn` and `endTurn` ([turn.go](../../../pkg/engine/turn.go)); `setup.go:100` (the first turn's `beginTurn`); `Apply`'s `ActEndTurn` case ([actions.go:93-96](../../../pkg/engine/actions.go): `endTurn`, then `settled`, no `resolveStack`); `resolveStack` ([stack.go:112-122](../../../pkg/engine/stack.go): flush, then resolve, sweep, flush and settle per item); `fireHook`, `fireHookFor` (the one place a trigger's gates are checked, including `Turn == TurnYours` against `g.Active`) and `flushTriggers` ([trigger.go](../../../pkg/engine/trigger.go)); `Hook` ([ability.go:137-149](../../../pkg/engine/ability/ability.go)); `doc.go:18-21` (lists the hooks).
- Existing equivalents searched (D2.3): no hook watches a turn boundary (`git grep -n "TurnStart" pkg` finds only `emitTurnStart`, the event-log snapshot, and `EventTurnStart`). `resolveStack` is the one flush-resolve-sweep cycle: reuse it at the turn boundary. A second loop of flush, resolve and sweep in `turn.go` would be a duplicate (D2.0).

### Shape (the engineer builds it)
- **`ability.HookTurnStart`**, added at the end of the `Hook` constants so no existing value changes: a turn began, fired in the Ready Phase after the active player's assets ready (`FR-TRN-001`). It fires on every turn start, so "At the start of your turn" is `HookTurnStart` with `Turn: TurnYours`, the existing gate.
- **`fireTurnStart()`** in `trigger.go`: fires `HookTurnStart` through `fireHook` for every battlefield asset, either player's (`assetsByID`). The gates stay in `fireHookFor` (a face-down asset prints nothing; `ActiveZone`, `Turn`, `Cornered` and `MaxPerTurn` as for any hook).
- **`beginTurn`**: after the active player's assets ready, `fireTurnStart()`, then `resolveStack()` (which places the buffer in APNAP order and resolves every item, with cascades and a death sweep after each). If the match ended, return before the Draw Phase. The phase stays `PhaseReady` while they resolve, then the Draw Phase runs as today. Remove the "No executable card has start-of-turn triggers yet" comment. `ActEndTurn` and setup keep calling `beginTurn` as they do; the first turn of a match fires it too.
- `doc.go`: add `HookTurnStart` to the hooks it lists, with one clause on when it resolves (`FR-TRN-001`).
- Behavior doesn't change for any printed card: no definition uses `HookTurnStart` yet, and an empty buffer places nothing. `go tool task sim-compare -- autoloop/m2-wave-h` must be byte-identical (a gate here).

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
In a new file `pkg/engine/turn_start_trigger_test.go`, with synthetic cards: a structure printing a `HookTurnStart` trigger with `Turn: TurnYours` whose effect is an `OpScratch` 1 for its controller (the top card of the deck goes to Reserves, which the draw that follows shows), and variants as each test needs. Turns pass with `s.endTurn()`.
- `TestFR_TRN_001_StartOfTurnTriggerResolvesBeforeDraw`: player 0's structure; player 1 ends their turn. At player 0's turn start the trigger has resolved before the draw: the card that was on top of player 0's deck is a Scratch Company in Reserves, and the card second from the top is the one drawn into the hand. Afterwards the phase is Main and the stack is empty.
- `TestFR_TRN_001_StartOfYourTurnOnly`: the same structure doesn't fire when the opponent's turn starts (player 0 ends their turn: player 0's deck and Reserves are unchanged).
- `TestFR_TRN_001_StartOfTurnTriggerAfterReady`: start-of-turn triggers fire after the active player's assets ready (`FR-TRN-001` 1, line 138). `ability.Filter` has no readiness criterion, so observe it at a decision: give the start trigger a target slot (for example one chosen allied unit, `OpAddArmor` 1) and two allied units so the target is asked, and when the chooser is asked, check that the active player's unit that was Exhausted at the end of their previous turn is already Readied.
- `TestFR_TRN_001_StartOfTurnCascadeResolvesBeforeDraw` (`FR-TRG-004`): the start trigger deals 1 damage to an allied 1-health unit whose death trigger Scratches 1 for its controller. Both resolve before the draw: the top two cards of the deck went to Reserves, and the third is the one drawn.
- `TestFR_TRG_003_StartOfTurnTriggersInAPNAPOrder`: the active player's `TurnYours` start trigger and the non-active player's start trigger without a `Turn` gate fire on the same turn start; the non-active player's resolves first (it is placed on top, `FR-TRG-003`). Observe the order through the event log or resolution side effects.
- `TestFR_TRG_003_TwoStartOfTurnTriggersAreOrdered` (`SVY-R7`): two of the active player's start triggers fire together; the player is asked a `DecideOrder` at turn start, as for any two simultaneous triggers, and the answer decides which resolves first.
- `TestFR_TRN_001_MatchEndsDuringStartOfTurn`: the start trigger deals enough damage to the opposing Headquarters to defeat it (`SubjectAll`, opposing Headquarters filter). The match is over and no draw happened (the active player's hand size is unchanged).
- **Existing tests:** every other test passes unchanged.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected. The test author keeps any small builder in the test file.

## Allowed Files
- `pkg/engine/ability/ability.go` (`HookTurnStart`)
- `pkg/engine/trigger.go` (`fireTurnStart`)
- `pkg/engine/turn.go` (`beginTurn` fires and resolves start-of-turn triggers before the Draw Phase)
- `pkg/engine/doc.go` (the hooks list)

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run "TestFR_(TRN_001_(StartOfTurn|StartOfYourTurn|MatchEndsDuringStartOfTurn)|TRG_003_(StartOfTurn|TwoStartOfTurn))"`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-compare -- autoloop/m2-wave-h` passes (no printed card has a start-of-turn trigger yet)

## Log
- 2026-11-15 lead-architect (wave H run, session 6): written for *Hammerhand Bivouac*'s hook (design note 6m, "From wave B"). FRD 01 doesn't settle when an untargeted pool is split or what a timeout does with one, so the pool and the card wait for the ruling question committed on `main` (f8b574d, design note Open Question 24). The hook doesn't depend on it. Directive #15's trigger fired with this task: its part (1), the Ready Phase, is applied here; its draw and End Phase parts stay `waiting` on its row. Set to `tests`.
- 2026-11-15 rules-test-author: wrote pkg/engine/turn_start_trigger_test.go: TestFR_TRN_001_StartOfTurnTriggerResolvesBeforeDraw, TestFR_TRN_001_StartOfYourTurnOnly, TestFR_TRN_001_StartOfTurnTriggerAfterReady, TestFR_TRN_001_StartOfTurnCascadeResolvesBeforeDraw, TestFR_TRG_003_StartOfTurnTriggersInAPNAPOrder, TestFR_TRG_003_TwoStartOfTurnTriggersAreOrdered, TestFR_TRN_001_MatchEndsDuringStartOfTurn. Helpers requested: none (small builders startTrigger, startSteps, startStructure stay in the test file). `go vet` fails only on the undefined `ability.HookTurnStart` the engineer adds. `cardtown lint rules` passes. Nothing left out. The AfterReady test installs its own chooser via `s.g.choosers[0]` to read the unit's state at the moment the target is asked.
