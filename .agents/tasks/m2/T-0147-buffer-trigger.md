---
id: T-0147
title: One function buffers a fired trigger, bufferTrigger (directive #96, harness, flash, pure move)
kind: harness
tier: flash
wave: C.1 (board directive #96, structural review 7, finding 5)
status: review
branch: m2/autoloop-wave-c.1/T-0147-buffer-trigger
depends_on: [T-0146]
claims: [helper:bufferTrigger]
cards: []
requirements: []
---

## Goal
`fireHook` and `fireAttacks` build the same `pendingTrigger`. That literal moves into `(g *Game) bufferTrigger(a *Asset, label string, t ability.Trigger)` in `pkg/engine/trigger.go`, and both call it (#96). Wave F's off-battlefield triggers would otherwise add a third copy. A pure move: same behavior, same test names, same sim output.

## Inputs
- Board directive #96 (structural review 7, approved by the user, 2026-11-03).
- Design note: `docs/architecture/m2-ability-model.md`, Deliverable 6f (wave C.1).
- Rules the buffer serves, unchanged: `FR-TRG-002` to `FR-TRG-004`.
- Not in this task: `fireHook`'s gates (`FaceDown`, `Turn`, `ToZone`, `Cornered`, `MaxPerTurn`) stay where they are, because `fireAttacks`' Aggressive trigger passes none of them. `fireCombatKill` keeps setting `awaits` afterwards, on the rows `fireHook` added. Each caller still builds its own label, because the two labels differ.
- Placement: `bufferTrigger` goes in `trigger.go`, directly below `fireHook`, because that file owns the pending-trigger buffer.
- Why it is a pure move: the buffer gets the same fields in the same order of calls. `label` is computed by the caller from the same `Sprintf`, and `t` is the same trigger. The lead dry-ran these exact edits in a throwaway worktree of the run branch at 6e4106a: `gofmt -l` clean, `go tool task check` passes, `go tool task sim` runs clean, and `go tool task sim-compare -- autoloop/m2-wave-c.1` printed `[PASS]`.

### Where it is today (at `autoloop/m2-wave-c.1`, 6e4106a)
| Where | Today |
| :--- | :--- |
| `pkg/engine/trigger.go:111-115` | `fireHook`'s `pendingTrigger` literal |
| `pkg/engine/trigger.go:152-156` | `fireAttacks`' `pendingTrigger` literal |

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)

### 1. `pkg/engine/trigger.go`, in `fireHook`
Replace lines 111-115 (`g.pending = append(g.pending, pendingTrigger{` to its closing `})`) with this one line:
```go
		g.bufferTrigger(a, fmt.Sprintf("%s: trigger %d", a.Printed().Name, i+1), t)
```

### 2. `pkg/engine/trigger.go`, new function
Insert this text directly above the comment `// cornered reports whether p is Cornered: the opponent controls strictly more`, with one blank line on each side:
```go
// bufferTrigger adds the fired trigger t to the buffer: a is its source, a's
// controller now is its controller, and label names it in the flush's prompts.
func (g *Game) bufferTrigger(a *Asset, label string, t ability.Trigger) {
	g.pending = append(g.pending, pendingTrigger{
		placement: placement{source: a.ID, inst: a.Inst, controller: a.Controller, label: label},
		trig:      t,
		pos:       posOf(a),
	})
}
```

### 3. `pkg/engine/trigger.go`, in `fireAttacks`
Replace `fireAttacks`' five-line `g.pending = append(g.pending, pendingTrigger{` statement (lines 152-156 before edit 1) with this one line:
```go
		g.bufferTrigger(a, fmt.Sprintf("%s: Aggressive %d", a.Printed().Name, n), ability.Aggressive(n))
```

No other file changes.

## Acceptance Cases (harness tasks only; written by the lead)
No new tests: a pure move. W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **Existing tests unchanged** (#96): `git -C W diff autoloop/m2-wave-c.1 -- '*_test.go'` prints nothing, and `go -C W test ./...` passes.
2. **Sims byte-identical** (#96): `go -C W tool task sim-compare -- autoloop/m2-wave-c.1` prints `[PASS]`.
3. **One declaration, two callers** (#96): `git -C W grep -n 'bufferTrigger(' -- pkg/engine` prints exactly three lines: `pkg/engine/trigger.go:111` (the `fireHook` call), `pkg/engine/trigger.go:117` (the declaration) and `pkg/engine/trigger.go:158` (the `fireAttacks` call).
4. **One `pendingTrigger` built:** `git -C W grep -n 'pendingTrigger{' -- pkg/engine` prints exactly one line, at `pkg/engine/trigger.go:118`.
5. **One append to the buffer:** `git -C W grep -n 'g.pending = append' -- pkg/engine` prints exactly one line, at `pkg/engine/trigger.go:118`.
6. **Only the listed edits:** `git -C W diff autoloop/m2-wave-c.1 --numstat -- pkg` prints exactly this line: `12	10	pkg/engine/trigger.go`.
7. **Fences:** `go -C W run ./cmd/cardtown lint fences` exits 0.
8. **Nothing else changed:** `git -C W diff autoloop/m2-wave-c.1 --stat` lists only `pkg/engine/trigger.go` and this task file.
9. `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/engine/trigger.go`

## Done When
- [x] Cases 1–9 hold, with their commands and results in the Log
- [x] `go tool task sim-compare -- autoloop/m2-wave-c.1` passes (starter and stress sims byte-identical to the run branch)
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-03, lead-architect (wave C.1 run, session 4): wrote this task from directive #96 against the run branch at 6e4106a, after syncing `main` (merge 6e4106a: `BOARD.md` and `docs/PRODUCT_SUMMARY.md`, no conflicts). Harness, so it skips `tests` and goes straight to `ready`; tier `flash`, because the edits are exact and were dry-run. Every command here is on the headless allow-list, one per line, with no pipes, `$` or unquoted parentheses. Searched for existing equivalents: `pkg/engine` has no `bufferTrigger`; these two literals are the only places that build a `pendingTrigger` or append to `g.pending`. Waiting directives checked: none has triggered (#6, #7: mode is `serial`; #8: no `deck.go`; #15: no turn-boundary hook; #71, #91 (2): no such card). Next: `kernel-engineer` on Flash.
- 2026-11-03, kernel-engineer (session 1): implemented pure move of pendingTrigger literal to bufferTrigger helper in pkg/engine/trigger.go (#96).
  - Diff shows:
    - pkg/engine/trigger.go: replaced pendingTrigger literal in fireHook with g.bufferTrigger(a, fmt.Sprintf("%s: trigger %d", a.Printed().Name, i+1), t); added bufferTrigger method; replaced pendingTrigger literal in fireAttacks with g.bufferTrigger(a, fmt.Sprintf("%s: Aggressive %d", a.Printed().Name, n), ability.Aggressive(n)).
  - Reuse: bufferTrigger unifies pendingTrigger construction between fireHook and fireAttacks. Searched pkg/engine for bufferTrigger; no prior declaration existed.
  - Acceptance cases verified:
    - Case 1: git -C W diff autoloop/m2-wave-c.1 -- "*_test.go" printed nothing; go -C W test ./... passed (exit 0).
    - Case 2: go -C W tool task sim-compare -- autoloop/m2-wave-c.1 printed "[PASS] starter and stress sims match autoloop/m2-wave-c.1 byte for byte, exit codes included".
    - Case 3: git -C W grep -n "bufferTrigger(" -- pkg/engine printed exactly 3 lines: pkg/engine/trigger.go:111, pkg/engine/trigger.go:117, and pkg/engine/trigger.go:158.
    - Case 4: git -C W grep -n "pendingTrigger{" -- pkg/engine printed exactly 1 line: pkg/engine/trigger.go:118.
    - Case 5: git -C W grep -n "g.pending = append" -- pkg/engine printed exactly 1 line: pkg/engine/trigger.go:118.
    - Case 6: git -C W diff autoloop/m2-wave-c.1 --numstat -- pkg printed "12\t10\tpkg/engine/trigger.go".
    - Case 7: go -C W run ./cmd/cardtown lint fences exited 0 ("[PASS]").
    - Case 8: git -C W diff autoloop/m2-wave-c.1 --stat lists only pkg/engine/trigger.go and this task file.
    - Case 9: go -C W tool task check and go -C W tool task sim passed all checks.
  - Handoff commit feat(engine): bufferTrigger helper buffers pending triggers (T-0147). Task status set to review.
