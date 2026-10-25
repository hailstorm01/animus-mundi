---
id: T-0062
title: Fired triggers are placed when the stack starts resolving, not only after an item (directive #50 part 1, harness, flash, pure move)
kind: harness
tier: flash
wave: B (board directive #50, structural review 3 finding 2; before the exhaust hook, T-0063)
status: ready
branch: m2/T-0062-flush-before-resolve
depends_on: [T-0061]
claims: [convention:resolveStack flushes first]
cards: []
requirements: []
---

## Goal
`resolveStack` places any fired triggers before it resolves anything, and `declareAttack` no longer flushes the buffer itself. A trigger fired by an action that puts nothing on the stack, such as Generate, is then placed right after that action and resolves before the next one (`SVY-R17`). Today it would stay in the buffer and fail `CheckInvariants`. This is part 1 of directive #50. Part 2, the exhaust hook, is T-0063. The buffer is always empty when `resolveStack` starts today, so nothing else changes: same tests, same sim output.

## Inputs
- Board directive #50 (structural review 3, finding 2).
- Ruling: `SVY-R17` (FRD 01 v3.11, `FR-CAP-001`): a Generate trigger is placed after the generation and resolves before the next action. A March or Attack trigger is placed on top of the action.
- Why it is a pure move. `Apply` calls `resolveStack` after every action except End Turn (`actions.go:98`). `declareAttack` is called only from `Apply` (`actions.go:89`), and its last statement is the flush. So moving that flush to the start of `resolveStack` changes no order: no event is logged between the two places.

### Where it is today (at `main` adb055b)
| Where | Today |
| :--- | :--- |
| `pkg/engine/combat.go:14-16`, end of `declareAttack` | `// Step 2: attack triggers enqueue on top of the Attack action.`, then `g.fireAttacks(att)`, then `g.flushTriggers()` |
| `pkg/engine/stack.go:66-74`, `resolveStack` | flushes only inside the loop, after each item's death sweep |

## Edits (lead decisions; copy exactly)

### 1. `pkg/engine/stack.go`: `resolveStack`
The doc comment and the first line of the body become:
```go
// resolveStack places any triggers the action fired, then resolves top-down
// until the stack is empty, sweeping for deaths after each item (FR-QUE-001,
// FR-LAY-002, FR-TRG-004). Placing first puts a March's or an Attack's
// triggers on top of it, and resolves a Generate's before the next action
// (SVY-R17).
func (g *Game) resolveStack() {
	g.flushTriggers()
	for len(g.Stack) > 0 && !g.Outcome.Over {
```
The rest of the loop is unchanged, including its own `g.flushTriggers()` after each item.

### 2. `pkg/engine/combat.go`: `declareAttack`
Delete the final `g.flushTriggers()` line. The comment above `g.fireAttacks(att)` becomes:
```go
	// Step 2: attack triggers enqueue on top of the Attack action; Apply's
	// resolveStack places them before anything resolves.
```

## Acceptance Cases (harness tasks only; written by the lead)
No new tests: a pure move. Run each command on its own, pointed at your worktree. Record each command and its result in the Log; the reviewer re-runs them.
1. **Existing tests unchanged:** `git -C <worktree> diff main -- '*_test.go'` prints nothing, and `go -C <worktree> test ./...` passes.
2. **Sims byte-identical:** `go -C <worktree> tool task sim-compare` prints `[PASS]`.
3. **Two flushes, both in resolveStack:** `git -C <worktree> grep -n 'g.flushTriggers()' -- pkg/engine/stack.go pkg/engine/combat.go` prints exactly two lines, both in `stack.go`.
4. **Only these edits:** `git -C <worktree> diff main --stat` lists only `pkg/engine/stack.go`, `pkg/engine/combat.go` and this task file.
5. `go -C <worktree> tool task check` and `go -C <worktree> tool task sim` pass.

## Allowed Files
- `pkg/engine/stack.go` (`resolveStack`'s comment and first line)
- `pkg/engine/combat.go` (`declareAttack`'s flush and comment)

## Done When
- [ ] Cases 1–4 hold, with their commands and results in the Log
- [ ] `go tool task sim-compare` passes (starter and stress sims byte-identical to the integration branch)
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks

## Log
- 2026-10-25 lead-architect: Wrote the task from board directive #50, whose trigger fired this prompt: T-0063 claims the exhaust hook. The directive asks one task to move the flush, then add the hook. Together with the Shambler card that is six non-test files, so under the size rule (D2.4) the work is split into this pure move, the exhaust-hook shape (T-0063, `queued`) and the Shambler card (T-0064, `queued`), in the directive's order. Next in wave B: the exhaust hook was chosen ahead of the Cornered gate. Cornered's only deck card, Allshan, also needs a power step (`MODIFY_STAT`), so it would be two shapes for one deck card, against one shape for three Shambler copies. Checked the line references against `main` adb055b. Every command under Acceptance Cases and Done When fits the headless allowlist. Waiting triggers checked: #6 and #7 (mode is `serial`), #8 (no `deck.go` behavior change), #15 (no turn hook), #20 (no wave C shape task) and #51 (no "if you do" task). None has fired. #50 has, and is now `open`.
  - Next: dispatch `kernel-engineer` on Gemini Flash through `agy` (D3).
