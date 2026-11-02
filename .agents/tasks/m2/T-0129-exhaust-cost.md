---
id: T-0129
title: An additional cost can exhaust a readied allied unit, and exhausting an Exhausted asset does nothing (FR-TRN-003a, SVY-R17, directive #58, wave C shape, sonnet)
kind: shape
tier: sonnet
wave: C
status: ready
branch: m2/autoloop-wave-c/T-0129-exhaust-cost
depends_on: [T-0063, T-0128]
claims: [const:CostExhaust, func:exhaust, func:costDecision, func:payCosts]
cards: []
requirements: [FR-TRN-003a, FR-TRN-003, FR-MNV-001, FR-CAP-001, FR-MOD-001]
---

## Goal
A Maneuver can print *"As an additional cost to enact this maneuver, exhaust a readied allied unit."* (*Funeral Fungus*, a deck card). Its controller chooses one Readied allied unit as it is declared, and the unit exhausts then, through the one `exhaust`, so its exhaust triggers go on the stack above the Maneuver and resolve first. A Maneuver whose exhaust cost can't be paid can't be enacted. Directive #58 lands here too: exhausting an asset that is already Exhausted does nothing. Split from T-0128 under the D2 size rule; *Funeral Fungus*'s definition is T-0130.

## Inputs
- FRD 01:
  - `FR-TRN-003a`, "Additional Sacrifice, Discard & Exhaust Costs" (line 335): a cost of exhausting an allied unit (*Funeral Fungus*) is paid immediately on declaration. Abilities triggered by paying it go onto the Resolution Stack **on top of** the declared card and resolve before it.
  - `FR-TRN-003`, `FR-MNV-001` (line 333): the declared card is already on the Resolution Stack before cost-triggered abilities enqueue. `SVY-R29` (line 334): a Maneuver that can't be enacted stays in hand.
  - `FR-CAP-001`, "What Counts as Exhausting" (`SVY-R17`, lines 299-303): an asset *exhausts* when it changes from **Readied** to **Exhausted** for any reason, an additional cost included. Triggers caused by exhausting for a cost are placed "as `FR-TRN-003a`, on top of the card or ability paid for" (line 302).
  - `FR-MOD-001` §1 (line 380), the single-option fast path, and §6 (line 389, "Mandatory Targets and Untargeted Selections"): a timed-out untargeted selection takes the first legal option in fixed order (own Headquarters, then Reserves, then the Main Theater, each zone in the order its assets entered).
- Rulings that apply: `SVY-R2`, `SVY-R16`, `SVY-R17`, `SVY-R29`.
- Directive #58 (`BOARD.md`): `exhaust` changes only a Readied asset. Add `if a.Exhausted { return }` at its top, plus a test that exhausting an already-Exhausted asset fires no exhaust trigger. Every caller today checks readiness first (`check` for Generate, March, Attack and Scratch), so no current play changes.
- Survey: wave C row (`docs/design/m2_starter_set_ability_survey.md` line 31, *Funeral Fungus* ★); its FRD 02 benchmark blueprint #8 is T-0130's test source.
- Design note (`docs/architecture/m2-ability-model.md`): Boundaries line 34, the `Cost` shape ("kill/exhaust a chosen allied asset, discard N, ...").
- Kernel today (run branch at 5ed2ca1):
  - `ability.CostKind` with its one member `CostKill` (`ability/ability.go:33-38`), and `ability.Cost{Kind, Filter}` (line 40-46).
  - `placement.go`: `costDecision` (line 44-54) lists the assets matching `c.Filter` as a `DecidePick`, with a kill prompt; `costsPayable` (line 56-65); `payCosts` (line 67-81) kills each picked asset; `placeable` (line 94-106), whose doc comment has one 110-column line.
  - `Apply`'s `ActEnact` case (`actions.go:83-90`): `g.payCosts(pl)` runs right after `g.push` of the Maneuver (T-0128 round 3).
  - `exhaust` (`actions.go:110-115`): sets `Exhausted` and calls `fireExhausted` (`trigger.go:57`), which buffers the asset's `HookExhausted` triggers; `resolveStack` flushes the buffer before resolving anything, so a trigger buffered after the push is placed on top of the Maneuver.

### Shape (lead decisions, binding)
- **`ability.CostExhaust`**, the second `CostKind` member, after `CostKill`: "exhaust one Readied allied asset matching `Cost.Filter`, chosen by the payer (`FR-TRN-003a`, `SVY-R17`)". Readiness belongs to the kind, not to `Filter`: no field is added to `ability.Filter`.
- **`costDecision`**: for `CostExhaust`, an asset is an option only if it matches `c.Filter` and is not `Exhausted`; the prompt says "exhaust" where the kill cost's says "kill". `CostKill`'s options and prompt are unchanged. `costsPayable` and `placeable` need no change of logic: they already read `costDecision`'s options.
- **`payCosts`**: for each cost in printed order, pick through `costDecision` and `pickFrom` as today, then by kind: `CostKill` through `g.kill`, `CostExhaust` through `g.exhaust`. A `CostKind` it doesn't know panics (no silent no-op). Its doc comment names both kinds.
- **`exhaust`**: `if a.Exhausted { return }` at its top (directive #58), and its doc comment says that exhausting an Exhausted asset changes nothing and fires nothing.
- **`placeable`'s doc comment**: re-wrapped to the file's usual width (T-0128 review note 2). No code change.
- **Out of scope:** *Funeral Fungus* (T-0130); the discard cost and *Steel Rain* (#20 (b), #59 (2)); exhaust costs on abilities or modes (#61's fifth question); a `Filter` field for readiness.
- **Sims:** no card uses `CostExhaust` yet, and the `exhaust` guard changes no current path, so `go tool task sim-compare -- autoloop/m2-wave-c` must pass (starter and stress sims byte-identical to the run branch).

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
In a new file `pkg/engine/exhaust_cost_test.go`, on synthetic Maneuvers built with `maneuver_test.go`'s `maneuver`, `steps`, `enactAction`, `offersEnact` and `lastDiscard`, `kill_cost_test.go`'s `alliedUnitFilter` and `kcScript`, and `exhaust_hook_test.go`'s `exhaustTrigger`, `exhaustWatcher` and `logIndex`. An `exhaustCostManeuver` helper (as `killCostManeuver`) goes in the new test file. No existing test is edited.
- `TestFR_TRN_003a_ExhaustCostPaidOnDeclaration`: a synthetic Maneuver with the exhaust cost (allied unit) and an untargeted step with a visible result. The controller has two Readied allied units, and a scripted chooser picks the second. After `Apply`: that unit is Exhausted and still on the battlefield, the other is still Readied, the step ran, the Maneuver is in its owner's discard pile, and its resource cost was spent.
- `TestFR_TRN_003a_ExhaustCostNeedsAReadiedUnit` (`SVY-R29`): the controller's allied units are all Exhausted, and its Headquarters and an allied Structure are Readied. The Maneuver is not in `LegalActions` (`offersEnact` false), `Apply` of it returns an error wrapping `ErrIllegal`, and the card stays in hand with resources unspent. With one Readied allied unit added, it is offered.
- `TestFR_TRN_003a_ExhaustCostTriggerResolvesBeforeTheManeuver` (`SVY-R17`, `FR-CAP-001` line 302): the paying unit has a synthetic exhaust trigger with a step that logs an event (2 damage to the opposing Headquarters), and the Maneuver's step logs one too (1 damage to it), as T-0128's `TestFR_TRN_003a_CostTriggersResolveBeforeTheManeuver` does. A draw logs no event, so neither step is a draw. In the event log: the Maneuver's `place` at stack depth 1, then the trigger's `place` at depth 2, then the trigger's damage, then the Maneuver's damage, then the Maneuver to the discard pile.
- `TestFR_MOD_001_ExhaustCostTimeoutPicksFixedOrder` (§6, `SVY-R16`): two Readied allied units, one in the Main Theater deployed first and one in the Reserves deployed later. The cost's decision times out: the Reserves unit is exhausted and the Main Theater unit stays Readied.
- `TestFR_MOD_001_ExhaustCostSingleOptionNotAsked` (§1): one Readied and one Exhausted allied unit. The Maneuver resolves with no decision asked of the controller, and the Readied unit is now Exhausted: an Exhausted unit is not an option.
- `TestFR_CAP_001_ExhaustingAnExhaustedAssetDoesNothing` (directive #58, `SVY-R17`): an `exhaustWatcher` on the battlefield that is already Exhausted. Calling `exhaust` on it, then letting the game settle, fires no exhaust trigger (no Scratch happens, nothing is placed on the stack) and leaves it Exhausted. The same unit Readied and exhausted once fires its trigger once, as a control.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected.

## Allowed Files
- `pkg/engine/ability/ability.go` (`CostExhaust`)
- `pkg/engine/placement.go` (`costDecision`, `payCosts`, their doc comments, `placeable`'s doc comment)
- `pkg/engine/actions.go` (`exhaust` only)

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run 'TestFR_TRN_003a_ExhaustCost|TestFR_MOD_001_ExhaustCost|TestFR_CAP_001_ExhaustingAnExhausted'`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim-compare -- autoloop/m2-wave-c` passes (no card uses the exhaust cost yet)

## Log
- 2026-10-31 lead-architect: queued in wave C session 15, with T-0128's task commit (D2.4). #58 fires with this task, whose `claims:` include exhausting an asset by a cost.
- 2026-11-02 lead-architect: T-0128 merged (d67487b). Covers updated: `payCosts` now runs after the push, and the `placeable` comment re-wrap carried from T-0128's review.
- 2026-11-02 lead-architect (wave C session 16): Inputs, Shape, Tests, Allowed Files and Done When written against the run branch at 5ed2ca1; branch renamed for the wave run (D7.4); directive #58 applied here. Status `tests`.
- 2026-11-02 rules-test-author: wrote the six listed tests in pkg/engine/exhaust_cost_test.go (TestFR_TRN_003a_ExhaustCostPaidOnDeclaration, ..._ExhaustCostNeedsAReadiedUnit, ..._ExhaustCostTriggerResolvesBeforeTheManeuver, TestFR_MOD_001_ExhaustCostTimeoutPicksFixedOrder, ..._ExhaustCostSingleOptionNotAsked, TestFR_CAP_001_ExhaustingAnExhaustedAssetDoesNothing with a control subtest) plus the exhaustCostManeuver helper in that file. Helpers requested: none. Cases left out: none. go vet fails only on undefined ability.CostExhaust (expected); cardtown lint rules passes. Status tests-review.
- 2026-11-02 lead-architect: tests approved (78d9582): the six listed tests check the cited FRD 01 text (FR-TRN-003a line 335, FR-CAP-001 line 302, SVY-R29, FR-MOD-001 sec. 1 and 6), the order test watches logged damage rather than a draw, and #58's test has a Readied control. Status `ready`, kernel-engineer (sonnet).
