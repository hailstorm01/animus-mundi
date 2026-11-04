---
id: T-0150
title: Event readers compare with Resource.String, not typed-out names (directive #98, part 2 of 3, harness, flash, pure move)
kind: harness
tier: flash
wave: C.1 (board directive #98, structural review 7, finding 7)
status: done
branch: m2/autoloop-wave-c.1/T-0150-resource-name-readers
depends_on: [T-0149]
claims: [file:pkg/eventlog/replay.go, file:pkg/sim/monitor_resources.go]
cards: []
requirements: []
---

## Goal
`actionOf` in `pkg/eventlog/replay.go` and the Generate monitor in `pkg/sim/monitor_resources.go` compare an event's resource against typed-out names. Both use `engine.Requisition.String()` and `engine.Tactics.String()` instead (#98). A pure move: same behavior, same test names, same sim output.

## Inputs
- Board directive #98 (structural review 7, approved by the user, 2026-11-03).
- Design note: `docs/architecture/m2-ability-model.md`, Deliverable 6f (wave C.1).
- Split from T-0149 under the D2 size rule: #98 touches nine non-test files. T-0149 (`done`, merge 2d98782) added `Resource.String` and the sites that write a name. This task is the two files that read one. T-0151 (`queued`) is `IsScratchCompany`.
- Not in this task: monitor messages written as sentences (`"...generated %s, not Requisition"`, `"...%s, not Tactics"`, `"player %d converted with %d Requisition, ..."`) stay, as do `pool`'s comment and `Event.Resource`'s comment in `events.go`. So do the event values `pkg/sim/monitor_test.go` passes in, which are test setup.
- Why it is a pure move: `Resource.String()` returns exactly `"Tactics"` for `engine.Tactics` and `"Requisition"` for `engine.Requisition`, and both files only compare strings against those values. Neither writes an event.
- `requisition, tactics := ...` follows the line `reserves := engine.ZoneReserves.String()` in the same function.
- The lead dry-ran these exact edits in a throwaway worktree of the run branch at 307153d: `gofmt -l` clean, `go tool task check` passes, `go tool task sim` runs clean, and `go tool task sim-compare -- autoloop/m2-wave-c.1` printed `[PASS]`.

### Where it is today (at `autoloop/m2-wave-c.1`, 307153d)
| Where | Today |
| :--- | :--- |
| `pkg/eventlog/replay.go:269` | `actionOf` compares against a typed-out `"Tactics"` |
| `pkg/sim/monitor_resources.go:19` | `pool` compares against a typed-out `"Requisition"` |
| `pkg/sim/monitor_resources.go:49-58` | `checkGenerate` picks `res`, `other` and `yield` by typed-out names |
| `pkg/sim/monitor_resources.go:64` and `:66` | `checkGenerate`'s zone checks compare against typed-out names |

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)

### 1. `pkg/eventlog/replay.go`, in `actionOf`
Replace line 269 with this line:
```go
	if a.Kind == engine.ActGenerate && e.Resource == engine.Tactics.String() {
```

### 2. `pkg/sim/monitor_resources.go`, in `pool`
Replace line 19 with this line:
```go
	if resource == engine.Requisition.String() {
```

### 3. `pkg/sim/monitor_resources.go`, in `checkGenerate`
Replace lines 49-54 (the `res, other := e.Resource, "Tactics"` line through the `other = "Requisition"` line) with these lines:
```go
	requisition, tactics := engine.Requisition.String(), engine.Tactics.String()
	res, other := e.Resource, tactics
	yield := c.TacYield
	if res == requisition {
		other, yield = tactics, c.ReqYield
	} else if res == tactics {
		other = requisition
```

### 4. `pkg/sim/monitor_resources.go`, in `checkGenerate`'s zone checks
Replace the line `case a.Zone == reserves && res != "Requisition":` with this line:
```go
		case a.Zone == reserves && res != requisition:
```
Replace the line `case a.Zone != reserves && res != "Tactics":` with this line:
```go
		case a.Zone != reserves && res != tactics:
```
Leave the `m.fail` line under each of them as it is.

No other file changes.

## Acceptance Cases (harness tasks only; written by the lead)
No new tests: a pure move. W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **Existing tests unchanged** (#98): `git -C W diff autoloop/m2-wave-c.1 -- '*_test.go'` prints nothing, and `go -C W test ./...` passes.
2. **Sims byte-identical** (#98): `go -C W tool task sim-compare -- autoloop/m2-wave-c.1` prints `[PASS]`.
3. **Replay uses it** (#98): `git -C W grep -n 'e.Resource == engine.Tactics.String()' -- pkg/eventlog/replay.go` prints exactly one line, at `pkg/eventlog/replay.go:269`.
4. **`pool` uses it** (#98): `git -C W grep -n 'resource == engine.Requisition.String()' -- pkg/sim/monitor_resources.go` prints exactly one line, at `pkg/sim/monitor_resources.go:19`.
5. **`checkGenerate` uses it** (#98): `git -C W grep -n 'requisition, tactics := engine.Requisition.String(), engine.Tactics.String()' -- pkg/sim/monitor_resources.go` prints exactly one line, at `pkg/sim/monitor_resources.go:49`.
6. **No typed-out Tactics left** (#98): `git -C W grep -n -E '[^ .]Tactics' -- pkg/eventlog/replay.go pkg/sim/monitor_resources.go` prints nothing (git grep exits 1 when nothing matches; that is the expected result). The pattern skips `engine.Tactics` and the sentences, where a space comes before the word.
7. **No typed-out Requisition left** (#98): `git -C W grep -n -E '[^ .]Requisition' -- pkg/eventlog/replay.go pkg/sim/monitor_resources.go` prints nothing (exit 1, as in case 6).
8. **Only the listed edits:** `git -C W diff autoloop/m2-wave-c.1 --numstat -- pkg cmd` prints exactly these two lines: `1	1	pkg/eventlog/replay.go` and `9	8	pkg/sim/monitor_resources.go`.
9. **Fences:** `go -C W run ./cmd/cardtown lint fences` exits 0.
10. **Nothing else changed:** `git -C W diff autoloop/m2-wave-c.1 --stat` lists only `pkg/eventlog/replay.go`, `pkg/sim/monitor_resources.go` and this task file.
11. `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/eventlog/replay.go`
- `pkg/sim/monitor_resources.go`

## Done When
- [x] Cases 1–11 hold, with their commands and results in the Log
- [x] `go tool task sim-compare -- autoloop/m2-wave-c.1` passes (starter and stress sims byte-identical to the run branch)
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-03, lead-architect (wave C.1 run, session 6): queued with T-0149 (D2.4).
- 2026-11-03, lead-architect (wave C.1 run, session 7): wrote Inputs, Edits, Acceptance Cases and Allowed Files against the run branch at 307153d; `main` had no commits to sync. Harness, so it skips `tests` and goes straight to `ready`; tier `flash`, because the edits are exact and were dry-run. Every command here is on the headless allow-list, one per line, with no pipes, `$`, double quotes or unquoted parentheses. The reuse is `Resource.String` itself (T-0149); no other resource-name helper exists. Waiting directives checked: none has triggered (#6, #7: mode is `serial`; #8: T-0151 touches `deck.go` but is a pure move; #15: no turn-boundary hook; #71, #91 (2): no such card). Next: `kernel-engineer` on Flash.
- 2026-11-03, kernel-engineer (wave C.1 run, Flash): implemented exact edits for T-0150.
  - Diff: `pkg/eventlog/replay.go:269` now compares `e.Resource == engine.Tactics.String()`. In `pkg/sim/monitor_resources.go:19`, `pool` compares `resource == engine.Requisition.String()`. In `checkGenerate` (lines 49-58), sets `requisition, tactics := engine.Requisition.String(), engine.Tactics.String()`, compares `res == requisition` and `res == tactics`, and uses `requisition` and `tactics` in zone checks `a.Zone == reserves && res != requisition` and `a.Zone != reserves && res != tactics`.
  - Reused `engine.Requisition.String()` and `engine.Tactics.String()` from T-0149 (`pkg/engine/actions.go`).
  - Acceptance cases verified:
    - Case 1: `git -C W diff autoloop/m2-wave-c.1 -- "*_test.go"` output empty; `go -C W test ./...` passed (exit 0).
    - Case 2: `go -C W tool task sim-compare -- autoloop/m2-wave-c.1` output `[PASS] starter and stress sims match autoloop/m2-wave-c.1 byte for byte, exit codes included`.
    - Case 3: `git -C W grep -n "e.Resource == engine.Tactics.String()" -- pkg/eventlog/replay.go` -> line 269.
    - Case 4: `git -C W grep -n "resource == engine.Requisition.String()" -- pkg/sim/monitor_resources.go` -> line 19.
    - Case 5: `git -C W grep -n "requisition, tactics := engine.Requisition.String(), engine.Tactics.String()" -- pkg/sim/monitor_resources.go` -> line 49.
    - Case 6: `git -C W grep -n -E "[^ .]Tactics" -- pkg/eventlog/replay.go pkg/sim/monitor_resources.go` -> exit code 1 (no matches).
    - Case 7: `git -C W grep -n -E "[^ .]Requisition" -- pkg/eventlog/replay.go pkg/sim/monitor_resources.go` -> exit code 1 (no matches).
    - Case 8: `git -C W diff autoloop/m2-wave-c.1 --numstat -- pkg cmd` -> `1 1 pkg/eventlog/replay.go` and `9 8 pkg/sim/monitor_resources.go`.
    - Case 9: `go -C W run ./cmd/cardtown lint fences` -> exit code 0 ([PASS]).
    - Case 10: `git -C W diff autoloop/m2-wave-c.1 --stat` -> only `pkg/eventlog/replay.go`, `pkg/sim/monitor_resources.go`, and task file.
    - Case 11: `go -C W tool task check` passed all checks; `go -C W tool task sim` passed (200 games clean, P0 82.5% / P1 17.5% / draws 0.0%).
- 2026-11-03, lead-architect (wave C.1 run, session 7, D5): **Approved, `done`.** Merged into `autoloop/m2-wave-c.1` as 0a64407. Flash run `0e62721b-6499-42e9-a090-c4dbeb953f14`: exit 0, `status: SUCCESS`, 338 s, no refusals (no `denied_actions` key; the transcript has no step with `"status":"ERROR"`; its one `permission check failed` match is step 22, a `DONE` step quoting the run file; the `.err` file holds only agy's line about waiting on a background task). Review of `git diff autoloop/m2-wave-c.1...m2/autoloop-wave-c.1/T-0150-resource-name-readers` (257e813):
  - Scope: only `replay.go`, `monitor_resources.go` and this file changed. No test file changed.
  - Pure move, checked against the diff: the two code files match the task's Edits exactly. Both only compare strings against what `Resource.String` returns, and neither writes an event, so as T-0149's review noted, the event JSON comparison isn't repeated: the replay tests and `sim-compare` are the gates that bear on it.
  - Fences and placement: no card name, no new rule, no new declaration.
  - Re-run by the reviewer in a scratch worktree of the task branch, all as expected: cases 1-10; `gofmt -l` clean; `go tool task check` passes; `go tool task sim` clean, 200 games, P0 82.5% / P1 17.5% / draws 0.0%, 42/50 and 32/50 main cards executable; `go tool task sim-compare -- autoloop/m2-wave-c.1` `[PASS]`.
  - The merge's `pkg/` and `cmd/` match the task tip's, so the gates hold on the run branch. #98 stays open for T-0151.
