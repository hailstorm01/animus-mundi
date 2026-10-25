---
id: T-0063
title: An exhaust hook, fired from (g *Game) exhaust on every change from Readied to Exhausted (SVY-R17, directive #50 part 2, wave B shape, sonnet)
kind: shape
tier: sonnet
wave: B
status: queued
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
To be written by rules-test-author, with synthetic cards only (the Shambler is T-0064), in a new `pkg/engine/exhaust_hook_test.go`. The lead's guidance on what to cover follows; the author names the tests and lists them here.
- **Named (directive #47 part 4):** `TestFR_CAP_001_EnteringExhaustedIsNotAnExhaust`. A synthetic Deploys Exhausted unit with an exhaust trigger is deployed, and nothing fires.
- **Every exhaust fires:** an exhaust-trigger watcher fires once when it Generates, once when it Marches, and once when it Attacks. A synthetic Headquarters with an exhaust trigger fires when it Scratches.
- **Placement (`SVY-R17`):** a March's trigger resolves before the March does: the watcher is still in its origin zone when its trigger resolves. An Attack's trigger resolves before combat damage. A Generate's trigger resolves before the next action, and after `Apply` returns the stack and buffer are empty and the invariants hold.
- **No other change fires it:** readying at turn start doesn't. Another asset exhausting doesn't (self hook only).
- **No limit:** the watcher fires on each exhaust across turns.
- Use an observable step, as the T-0054 tests did with Scratch 1 and `wantScratches`.

## Allowed Files
- `pkg/engine/ability/ability.go` (`HookExhausted`)
- `pkg/engine/actions.go` (`exhaust` fires the hook)
- `pkg/engine/trigger.go` (`fireExhausted`)
- `pkg/engine/doc.go` (the hooks list)

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run '<the author's pattern>'`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks. No starter card has an exhaust trigger yet, so the sim should be unchanged; check with `go tool task sim-compare` and record the result.

## Log
- 2026-10-25 lead-architect: Wrote the task, `queued` behind T-0062 (directive #50 part 1). Split from directive #50 by the size rule; see T-0062's Log.
