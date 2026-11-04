---
id: T-0149
title: A resource's name lives in one place, Resource.String (directive #98, part 1 of 3, harness, flash, pure move)
kind: harness
tier: flash
wave: C.1 (board directive #98, structural review 7, finding 7)
status: done
branch: m2/autoloop-wave-c.1/T-0149-resource-string
depends_on: [T-0148]
claims: [helper:Resource.String]
cards: []
requirements: []
---

## Goal
`Resource` has no `String()`, so `emitAction` and the play screen each type out "Requisition" and "Tactics". `Resource.String()` in `pkg/engine/actions.go` becomes the one place that names a resource, and both call it (#98). A pure move: same behavior, same test names, same sim output, same event JSON.

## Inputs
- Board directive #98 (structural review 7, approved by the user, 2026-11-03).
- Design note: `docs/architecture/m2-ability-model.md`, Deliverable 6f (wave C.1).
- Decision record: ADR-0005 (event JSON values don't change).
- Split under the D2 size rule: #98 touches nine non-test files. This task is the declaration and the two sites that write a name. T-0150 (`queued`) moves the two sites that read one, `replay.go` and `monitor_resources.go`. T-0151 (`queued`) is `IsScratchCompany`.
- Not in this task: `"Convert: 2 Requisition -> 1 Tactics"` at `cmd/cardtown/play_screen.go` and the `illegal("conversion needs 2 Requisition")` message are sentences, not a resource's name, and stay. So does `Event.Resource`, which stays a `string`.
- Placement: `String` sits directly below `Resource`'s `const` block, as `ActionKind.String` sits below `ActionKind`'s.
- Why it is a pure move:
  - `String` is the old `if` itself, not an array literal like `ActionKind.String`. `canYield` accepts any `Resource` value on a Structure, a Headquarters or a Versatile unit, and today anything but `Tactics` is named "Requisition". An array literal would panic there instead.
  - No non-test code prints an `Action` or a `Resource` with `%v`, so adding a `String` method changes no other output (searched `pkg/engine`, `pkg/sim`, `pkg/eventlog`, `pkg/view`, `cmd`; the `%+v` calls at `pkg/eventlog/replay.go:221` print `Event.Resource`, a `string`).
  - JSON doesn't use `String`, and `Event.Resource` is a `string` either way.
- The lead dry-ran these exact edits in a throwaway worktree of the run branch at 17c4d50: `gofmt -l` clean, `go tool task check` passes, `go tool task sim` runs clean, and `go tool task sim-compare -- autoloop/m2-wave-c.1` printed `[PASS]`. `sim-compare` doesn't read event JSON or the play screen, so the lead also compared them against the run branch: `cardtown sim -events` JSON Lines for 50 starter and 50 stress games, identical apart from the `modified` provenance field (54,218 lines, both resources present), and a scripted `cardtown play -deck starter -p0 ANIMUS -p1 PATH -seed 7` screen, byte-identical (208 Requisition and 19 Tactics Generate labels).

### Where it is today (at `autoloop/m2-wave-c.1`, 17c4d50)
| Where | Today |
| :--- | :--- |
| `pkg/engine/actions.go:36-42` | `Resource` and its constants, no `String` |
| `pkg/engine/events.go:147-150` | `emitAction` types out both names |
| `cmd/cardtown/play_screen.go:153-157` | `actionLabel` types out both names |

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)

### 1. `pkg/engine/actions.go`, new method
Insert this text directly above the comment `// Action is one manual declaration. Fields not used by Kind are ignored.`, with one blank line on each side:
```go
// String names the resource, as events and the play screen write it.
func (r Resource) String() string {
	if r == Tactics {
		return "Tactics"
	}
	return "Requisition"
}
```

### 2. `pkg/engine/events.go`, in `emitAction`
Replace lines 147-150 (the `e.Resource = "Requisition"` line through the `if` block's closing brace) with this line:
```go
		e.Resource = a.Resource.String()
```

### 3. `cmd/cardtown/play_screen.go`, in `actionLabel`
Replace lines 153-157 (the `res := "Requisition"` line through the `return` under `case engine.ActGenerate:`) with this line:
```go
		return fmt.Sprintf("Generate: %s -> %s", findAsset(a.Asset), a.Resource)
```

No other file changes.

## Acceptance Cases (harness tasks only; written by the lead)
No new tests: a pure move. W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **Existing tests unchanged** (#98): `git -C W diff autoloop/m2-wave-c.1 -- '*_test.go'` prints nothing, and `go -C W test ./...` passes.
2. **Sims byte-identical** (#98): `go -C W tool task sim-compare -- autoloop/m2-wave-c.1` prints `[PASS]`.
3. **One declaration:** `git -C W grep -n 'Resource) String()' -- pkg/engine` prints exactly one line, at `pkg/engine/actions.go:45`.
4. **Events use it** (#98): `git -C W grep -n 'a.Resource.String()' -- pkg cmd` prints exactly one line, at `pkg/engine/events.go:147`.
5. **The play screen uses it** (#98): `git -C W grep -n 'findAsset(a.Asset), a.Resource' -- cmd/cardtown/play_screen.go` prints exactly one line, at `cmd/cardtown/play_screen.go:153`.
6. **The typed-out choices are gone** (#98): `git -C W grep -n 'a.Resource == ' -- pkg/engine/events.go cmd/cardtown/play_screen.go` prints nothing (git grep exits 1 when nothing matches; that is the expected result).
7. **Only the listed edits:** `git -C W diff autoloop/m2-wave-c.1 --numstat -- pkg cmd` prints exactly these three lines: `1	5	cmd/cardtown/play_screen.go`, `8	0	pkg/engine/actions.go` and `1	4	pkg/engine/events.go`.
8. **Fences:** `go -C W run ./cmd/cardtown lint fences` exits 0.
9. **Nothing else changed:** `git -C W diff autoloop/m2-wave-c.1 --stat` lists only `cmd/cardtown/play_screen.go`, `pkg/engine/actions.go`, `pkg/engine/events.go` and this task file.
10. `go -C W tool task check` and `go -C W tool task sim` pass.

The reviewer also repeats the event JSON and play screen comparison from Inputs. That one isn't yours to run.

## Allowed Files
- `pkg/engine/actions.go`
- `pkg/engine/events.go`
- `cmd/cardtown/play_screen.go`

## Done When
- [x] Cases 1–10 hold, with their commands and results in the Log
- [x] `go tool task sim-compare -- autoloop/m2-wave-c.1` passes (starter and stress sims byte-identical to the run branch)
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-03, lead-architect (wave C.1 run, session 6): wrote this task from directive #98 against the run branch at 17c4d50; `main` had no commits to sync. Harness, so it skips `tests` and goes straight to `ready`; tier `flash`, because the edits are exact and were dry-run. Every command here is on the headless allow-list, one per line, with no pipes, `$`, double quotes or unquoted parentheses. Searched for existing equivalents: `pkg/engine` has no `Resource.String` and no other resource-name helper. #98 is nine non-test files, so it splits into this task, T-0150 and T-0151, both `queued` with board rows (D2.4). Waiting directives checked: none has triggered (#6, #7: mode is `serial`; #8: T-0151 touches `deck.go` but is a pure move; #15: no turn-boundary hook; #71, #91 (2): no such card). Next: `kernel-engineer` on Flash.
- 2026-11-03, kernel-engineer (T-0149): implemented pure move of resource names into `Resource.String()`.
  - Diff shows:
    - `pkg/engine/actions.go`: added `func (r Resource) String() string` method on `Resource`.
    - `pkg/engine/events.go`: replaced manual `if a.Resource == Tactics` block with `e.Resource = a.Resource.String()`.
    - `cmd/cardtown/play_screen.go`: replaced manual `res` string assignment with direct `a.Resource` formatting in `actionLabel`.
  - Searched `pkg/engine` for reuse; confirmed no existing `Resource.String()` or equivalent resource-naming helper.
  - Acceptance cases verified:
    - Case 1: `git -C W diff autoloop/m2-wave-c.1 -- "*_test.go"` printed nothing; `go -C W test ./...` passed (all packages ok).
    - Case 2: `go -C W tool task sim-compare -- autoloop/m2-wave-c.1` passed (`[PASS] starter and stress sims match autoloop/m2-wave-c.1 byte for byte, exit codes included`).
    - Case 3: `git -C W grep -n "Resource) String()" -- pkg/engine` printed `pkg/engine/actions.go:45:func (r Resource) String() string {`.
    - Case 4: `git -C W grep -n "a.Resource.String()" -- pkg cmd` printed `pkg/engine/events.go:147:		e.Resource = a.Resource.String()`.
    - Case 5: `git -C W grep -n "findAsset(a.Asset), a.Resource" -- cmd/cardtown/play_screen.go` printed `cmd/cardtown/play_screen.go:153:		return fmt.Sprintf("Generate: %s -> %s", findAsset(a.Asset), a.Resource)`.
    - Case 6: `git -C W grep -n "a.Resource == " -- pkg/engine/events.go cmd/cardtown/play_screen.go` returned exit code 1 with no output.
    - Case 7: `git -C W diff autoloop/m2-wave-c.1 --numstat -- pkg cmd` printed `1	5	cmd/cardtown/play_screen.go`, `8	0	pkg/engine/actions.go`, and `1	4	pkg/engine/events.go`.
    - Case 8: `go -C W run ./cmd/cardtown lint fences` passed (exit code 0).
    - Case 9: `git -C W diff autoloop/m2-wave-c.1 --stat` lists only `cmd/cardtown/play_screen.go`, `pkg/engine/actions.go`, `pkg/engine/events.go`, and this task file.
    - Case 10: `go -C W tool task check` passed all stages (gofmt, vet, tests, linters, fences); `go -C W tool task sim` passed 200 starter matches cleanly.
  - Ready for review.
- 2026-11-03, lead-architect (wave C.1 run, session 6, D5): **Approved, `done`.** Merged into `autoloop/m2-wave-c.1` as 2d98782. Flash run `d1df6897-ce7e-48f5-ac56-6bd64673633f`: exit 0, `status: SUCCESS`, 308 s, no refusals (no `denied_actions` key; the transcript has no step with `"status":"ERROR"` and no `permission check failed` match; the `.err` file holds only agy's line about waiting on a background task). Review of `git diff autoloop/m2-wave-c.1...m2/autoloop-wave-c.1/T-0149-resource-string` (a2d3ae3):
  - Scope: only `actions.go`, `events.go`, `play_screen.go` and this file changed. No test file changed.
  - Pure move, checked against the diff: the three code files match the task's Edits exactly. `String` is the old `if`, so a value other than `Tactics` still reads "Requisition" and cannot panic.
  - Fences and placement: no card name, no new rule. `String` sits directly below `Resource`'s constants.
  - Re-run by the reviewer in a scratch worktree of the task branch, all as expected: cases 1-9; `gofmt -l` clean; `go tool task check` passes; `go tool task sim` clean, 200 games, P0 82.5% / P1 17.5% / draws 0.0%, 42/50 and 32/50 main cards executable; `go tool task sim-compare -- autoloop/m2-wave-c.1` `[PASS]`.
  - The lead's comparison against the run branch at 9dee39d, which `sim-compare` doesn't cover: `cardtown sim -events` for 50 starter and 50 stress games, identical once the `revision` and `modified` provenance fields are set aside (32,150 and 22,068 lines; the starter stream holds 5,827 Requisition and 937 Tactics values), and a scripted `cardtown play -deck starter -p0 ANIMUS -p1 PATH -seed 7` screen, byte-identical (208 Requisition and 19 Tactics Generate labels).
  - The merge's `pkg/` and `cmd/` match the task tip's, so the gates hold on the run branch. #98 stays open for T-0150 and T-0151. A note for T-0150's writer: its reviewer should repeat the event JSON comparison only if `actionOf` or the monitors could change what is written; they only read events, so the replay tests and `sim-compare` are the gates that bear on it.
