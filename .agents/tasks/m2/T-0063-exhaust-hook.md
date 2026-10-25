---
id: T-0063
title: An exhaust hook, fired from (g *Game) exhaust on every change from Readied to Exhausted (SVY-R17, directive #50 part 2, wave B shape, sonnet)
kind: shape
tier: sonnet
wave: B
status: tests-review
branch: m2/T-0063-exhaust-hook
depends_on: [T-0062]
claims: [hook:exhaust]
cards: []
requirements: [FR-CAP-001]
---

## Goal
A printed trigger can fire when its own asset exhausts: `ability.HookExhausted`, fired from `(g *Game) exhaust` (T-0058), the one place a Readied asset becomes Exhausted. That covers Generate, March, Attack and Scratch. Entering the battlefield Exhausted is not an exhaust. Placement follows `SVY-R17`, through the flush that T-0062 moves to the start of `resolveStack`. A March's or an Attack's trigger is placed on top of the action and resolves before it. A Generate's trigger is placed after the generation and resolves before the next action. The first card, Animoceps Shambler, is T-0064.

## Inputs
- FRD 01: `FR-CAP-001` (inherent actions; exhausting), `FR-TRG-002` (trigger hooks), and `SVY-R17` (v3.11).
- Survey row: Animoceps Shambler (wave B), line 107.
- Board directive #47 part 4 (the named settling test below) and #50 part 2.
- Design note: `docs/architecture/m2-ability-model.md`, Key types (`Hook`).
- Kernel today: `exhaust` at `pkg/engine/actions.go:104` has no hook. `fireHook` (`trigger.go`) buffers a self hook's triggers in printed order, and `fireDeploy`, `fireKilled` and `fireDamaged` wrap it.

### Shape (lead decisions)
- `ability.HookExhausted`, added after `HookDamaged` with the comment `// this asset exhausted: changed from Readied to Exhausted, for any reason (SVY-R17, FR-CAP-001)`.
- `func (g *Game) fireExhausted(a *Asset) { g.fireHook(a, ability.HookExhausted) }` in `trigger.go`, beside the other `fire*` wrappers, with a one-line doc comment.
- `exhaust` calls `g.fireExhausted(a)` after setting `Exhausted`. Entering Exhausted (`resolveDeploy`, `scratch`) still doesn't call `exhaust`.
- `doc.go`: add `HookExhausted` to the self hooks list in "Built".

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
Written in `pkg/engine/exhaust_hook_test.go`, with synthetic cards only (the Shambler is T-0064). The file does not compile until `ability.HookExhausted` exists. The watcher is a unit whose exhaust trigger Scratches 1, observed with `wantScratches` and by event order in the log.
- **Named (directive #47 part 4):** `TestFR_CAP_001_EnteringExhaustedIsNotAnExhaust`: a Deploys Exhausted unit with an exhaust trigger is deployed, and nothing fires.
- **Every exhaust fires:** `TestFR_CAP_001_ExhaustTriggerFiresWhenItsAssetGeneratesRequisition`, `...GeneratesTactics`, `...Marches`, `...Attacks`, `...WhenAHeadquartersDeploysScratchCompany`, `TestFR_CAP_001_HeadquartersExhaustTriggerFiresOnceWhenItGenerates`.
- **Placement (SVY-R17):** `TestFR_CAP_001_MarchExhaustTriggerResolvesBeforeTheMarch`, `TestFR_CAP_001_AttackExhaustTriggerResolvesBeforeCombatDamage`, `TestFR_CAP_001_GenerateExhaustTriggerResolvesAfterTheGenerationBeforeTheNextAction` (after `Apply` the stack is empty and the invariants hold), `TestFR_CAP_001_GenerateTriggerResolvesSoALaterMarchIsLegal`.
- **No other change fires it:** `TestFR_CAP_001_ReadyingAtTurnStartDoesNotFireTheExhaustTrigger`, `TestFR_TRG_002_AnotherAssetExhaustingDoesNotFireTheExhaustTrigger`, `TestFR_TRG_002_AnOpposingAssetExhaustingDoesNotFireTheExhaustTrigger`.
- **No limit:** `TestFR_TRG_002_ExhaustTriggerFiresOnEachExhaustAcrossTurns`.
- **Left out:** exhaust or additional costs and effect-caused exhausts (no engine support yet: `FR-ACT-001`, `FR-TRN-003a` are not executable); ordering between an exhaust trigger and an attack trigger on one attacker (FRD 01 does not fix it).

## Helpers Needed
None. The test file defines its own `exhaustTrigger`, `exhaustWatcher` and `logIndex`, and reuses existing helpers.

## Allowed Files
- `pkg/engine/ability/ability.go` (`HookExhausted`)
- `pkg/engine/actions.go` (`exhaust` fires the hook)
- `pkg/engine/trigger.go` (`fireExhausted`)
- `pkg/engine/doc.go` (the hooks list)

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run 'TestFR_CAP_001_(EnteringExhaustedIsNotAnExhaust|ExhaustTrigger|HeadquartersExhaustTrigger|MarchExhaustTrigger|AttackExhaustTrigger|GenerateExhaustTrigger|GenerateTrigger|ReadyingAtTurnStart)|TestFR_TRG_002_(AnotherAsset|AnOpposingAsset|ExhaustTriggerFiresOnEach)'`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks. No starter card has an exhaust trigger yet, so the sim should be unchanged; check with `go tool task sim-compare` and record the result.

## Log
- 2026-10-25 lead-architect: Wrote the task, `queued` behind T-0062 (directive #50 part 1). Split from directive #50 by the size rule; see T-0062's Log.
- 2026-10-25 lead-architect: T-0062 is `done` (7bfeb38), so `resolveStack` now flushes first. Re-checked against `main` 7bfeb38: `exhaust` (`actions.go:104-106`) only sets the field; `fireDeploy`, `fireKilled` and `fireDamaged` wrap `fireHook` (`trigger.go:23-36`); `HookDamaged` is the last hook constant (`ability.go:22`). No change to the plan. Waiting triggers checked: #6, #7, #8, #15, #20 and #51. None has fired.
  - Next: status `tests`; dispatch `rules-test-author`.
- 2026-10-25 rules-test-author: Wrote 15 tests in `pkg/engine/exhaust_hook_test.go` (listed under Tests); no helpers requested. The file does not compile until `ability.HookExhausted` exists. With a throwaway `HookExhausted` and `g.fireHook(a, HookExhausted)` in `exhaust` (reverted, not committed) all 15 pass. Left out: cost and effect exhausts (not executable yet), exhaust-versus-attack trigger ordering (FRD 01 silent). Status `tests-review`.
