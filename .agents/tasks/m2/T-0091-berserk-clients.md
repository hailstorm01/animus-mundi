---
id: T-0091
title: The clients stop reading an Attack's secondary target (directive #68 part (c), harness, flash)
kind: harness
tier: flash
wave: B.2 (board directive #68 part (c), with #72; the user's code review, before wave C)
status: done
branch: m2/T-0091-berserk-clients
depends_on: [T-0090]
claims: [reader:secondary-removed, file:pkg/eventlog/replay.go, file:cmd/cardtown/play_screen.go]
cards: []
requirements: []
---

## Goal
Nothing outside `pkg/engine` reads an Attack's secondary target. The replay checker no longer compares or rebuilds it, the view's stack item no longer has a `Secondary` field, and `cardtown play` no longer labels an Attack with one. This is part (c) of directive #68. Part (a) (T-0089) made a non-zero `Secondary` illegal, so every `Secondary` the engine writes is 0 and these readers only ever see 0. Part (d) removes the fields themselves (`Action.Secondary`, `StackItem.Secondary`, `Event.Secondary`) and `doc.go`'s Berserk, so this task leaves every file in `pkg/engine` alone.

## Inputs
- Board directive #68 (user code review, 2026-10-28), part (c), its Tests list (the `view_test.go` item) and acceptance cases (a) and (b). Directive #72 (3): no test is rewritten into an Overkill test.
- Design note: `docs/architecture/m2-ability-model.md`, Deliverable 6b (wave B.2).
- Why nothing observable changes: since T-0089 the engine rejects an Attack with a non-zero `Secondary`, so every `Event.Secondary` and `StackItem.Secondary` is 0. The event log still writes `secondary` (`omitempty`, so never), because `Event` is part (d)'s. The view is not stored anywhere: `view_test.go` marshals it only to check for instance IDs, and no golden file holds it.

### Where it is today (at `autoloop/m2-wave-b.2` 8a06d07)
| Where | Today |
| :--- | :--- |
| `pkg/eventlog/replay.go:219-221` | `diff`'s field comparison includes `rec.Secondary != got.Secondary` |
| `pkg/eventlog/replay.go:247-248` | `structured` returns a `[9]any` that includes `e.Secondary` |
| `pkg/eventlog/replay.go:260` | `actionOf` sets `Secondary: e.Secondary` on the rebuilt Action |
| `pkg/view/view.go:87` | `StackView.Secondary`, commented "the Berserk secondary target" |
| `pkg/view/view.go:172` | `stackView`'s Attack case sets `s.Target, s.Secondary` |
| `cmd/cardtown/play_screen.go:161-163` | `actionLabel`'s Attack case labels a secondary when `a.Secondary > 0` |
| `pkg/view/view_test.go:267` | `TestPlayerView`'s `stack[1]` check ends with `|| it.Secondary != 0` |

## Edits (lead decisions; copy exactly)

### 1. `pkg/eventlog/replay.go`
- In `diff`, the field-comparison case (lines 219-221) becomes two lines:
  ```go
  	case rec.Player != got.Player || rec.Card != got.Card || rec.Inst != got.Inst || rec.Source != got.Source ||
  		rec.Target != got.Target || rec.Zone != got.Zone || rec.Amount != got.Amount || rec.Resource != got.Resource:
  ```
  The `return` line under it does not change.
- `structured` becomes:
  ```go
  func structured(e engine.Event) [8]any {
  	return [8]any{e.Player, e.Card, e.Inst, e.Source, e.Target, e.Zone, e.Amount, e.Resource}
  }
  ```
- In `actionOf`, the first line becomes:
  ```go
  	a := engine.Action{Player: engine.PlayerID(e.Player), Asset: e.Source, Target: e.Target, Card: e.Inst}
  ```

### 2. `pkg/view/view.go`
- In `StackView`, remove the line `Secondary  int     // Attack: the Berserk secondary target, 0 if none`. Run `gofmt`; the other fields' alignment does not change, because `Controller` is the longest name and `[][]int` the longest type in that block.
- In `stackView`'s `case engine.StackAttack:`, the line becomes `s.Target = it.Target`.

### 3. `cmd/cardtown/play_screen.go`
- In `actionLabel`'s `case engine.ActAttack:`, remove the three lines from `if a.Secondary > 0 {` through its closing `}`, so the case is only:
  ```go
  	case engine.ActAttack:
  		return fmt.Sprintf("Attack: %s -> %s", findAsset(a.Asset), findAsset(a.Target))
  ```

No other non-test file changes.

## Tests (harness: the engineer makes this edit, in the `test(T-0091)` commit)
No new tests: this task removes readers of a value that is always 0, and cases 2–3 show they are gone. One existing test changes, named here as D2.4 requires. It passes both before and after the implementation commit: commit it first, run `go test ./...` on that commit (the old code), and record the result.

- `pkg/view/view_test.go`, `TestPlayerView`: in the `stack[1]` check (line 267), delete ` || it.Secondary != 0` so the condition ends with `it.Target != s.raider.ID {`. The `t.Errorf` line under it does not change.

## Acceptance Cases (harness tasks only; written by the lead)
Run each command on its own, pointed at your worktree (W). Record each command and its result in the Log. The reviewer re-runs them.
1. **Tests first, and only the named one:** the branch has a `test(T-0091): …` commit before the implementation commit, holding the test edit above and nothing else. `go -C W test ./...` passes on the test commit (old code) and on the branch tip. `git -C W diff autoloop/m2-wave-b.2 --stat -- '*_test.go'` lists exactly `pkg/view/view_test.go`.
2. **No client reads a secondary:** `git -C W grep -n -i secondary -- pkg/eventlog pkg/view pkg/sim cmd` prints nothing (exit code 1).
3. **What is left is part (d):** `git -C W grep -l -i -E "berserk|secondary|cleave" -- pkg cmd ":!pkg/toolchain/linter"` lists exactly `pkg/engine/actions.go`, `pkg/engine/combat.go`, `pkg/engine/doc.go`, `pkg/engine/events.go` and `pkg/engine/stack.go`.
4. **Sims unchanged (#68 case (b)):** `go -C W tool task sim-compare -- autoloop/m2-wave-b.2` passes: the starter and stress sims match byte for byte, exit codes included.
5. **Only these files:** `git -C W diff autoloop/m2-wave-b.2 --stat` lists only `pkg/eventlog/replay.go`, `pkg/view/view.go`, `cmd/cardtown/play_screen.go`, `pkg/view/view_test.go` and this task file.
6. `go -C W tool task check` and `go -C W tool task sim` pass.

The reviewer also compares a starter `-events` file from both builds (#68 case (b)); it must differ only in each game's `match-start` `revision` stamp. That runs built binaries, so it is the reviewer's check, not the engineer's.

## Allowed Files
- `pkg/eventlog/replay.go`
- `pkg/view/view.go`
- `cmd/cardtown/play_screen.go`
- `pkg/view/view_test.go`, for the named test edit only

## Done When
- [x] Cases 1–6 hold, with their commands and results in the Log
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks

## Log
- 2026-10-28, lead-architect (wave B.2 run, session 3): wrote the task from board directive #68 part (c), the next of wave B.2's blocking directives in board order. Harness, so it goes straight to `ready`; tier `flash`, the harness default (#68 names `sonnet` only for part (a)), and every command here is on the headless allow-list. Lead decisions: (1) no new test (see Tests); (2) `actionLabel` has no test today and gets none, since its remaining Attack label is unchanged; (3) the `-events` comparison is the reviewer's. Applied #72's design-note part in the same commit: the Overkill settling tests are recorded as an M3 item (design note Open Question 12). Waiting directives checked: none has triggered (#8: this task doesn't touch `deck.go`; #15, #20 (b), #58, #59 (2), #61, #71: no such hook, shape, claim or card here; #6, #7: mode is `serial`). Part (d), #69 and #70 stay open for the next sessions. Next: `kernel-engineer` on Flash.
- 2026-10-28, kernel-engineer (T-0091): implemented part (c) of directive #68.
  - Test commit 08d410c: deleted ` || it.Secondary != 0` in `pkg/view/view_test.go` (`TestPlayerView`). Verified with `go test ./...` on the old code (passed, 0 exit).
  - Implementation commit ecb4bfd:
    - `pkg/eventlog/replay.go`: removed `rec.Secondary != got.Secondary` from `diff` field comparisons; updated `structured` to return `[8]any` omitting `Secondary`; removed `Secondary: e.Secondary` in `actionOf`.
    - `pkg/view/view.go`: removed `Secondary int` from `StackView`; in `stackView`'s `case engine.StackAttack:`, set `s.Target = it.Target`. Formatted with `gofmt`.
    - `cmd/cardtown/play_screen.go`: removed `if a.Secondary > 0 { ... }` branch from `actionLabel`'s `case engine.ActAttack:`.
  - Reuse: removed defunct readers of Attack secondary targets in clients and tools; engine fields remain untouched for part (d).
  - Acceptance Cases verified:
    1. Tests first: commit 08d410c holds test edit only. `go -C W test ./...` passed on test commit and branch tip. `git -C W diff autoloop/m2-wave-b.2 --stat -- '*_test.go'` output: `pkg/view/view_test.go | 2 +-` (1 file changed).
    2. No client reads secondary: `git -C W grep -n -i secondary -- pkg/eventlog pkg/view pkg/sim cmd` printed nothing (exit code 1).
    3. Part (d) remainder: `git -C W grep -l -i -E "berserk|secondary|cleave" -- pkg cmd ":!pkg/toolchain/linter"` listed exactly:
       - `pkg/engine/actions.go`
       - `pkg/engine/combat.go`
       - `pkg/engine/doc.go`
       - `pkg/engine/events.go`
       - `pkg/engine/stack.go`
    4. Sims match: `go -C W tool task sim-compare -- autoloop/m2-wave-b.2` passed: "[PASS] starter and stress sims match autoloop/m2-wave-b.2 byte for byte, exit codes included".
    5. Only allowed files touched: `git -C W diff autoloop/m2-wave-b.2 --stat` listed only `pkg/eventlog/replay.go`, `pkg/view/view.go`, `cmd/cardtown/play_screen.go`, `pkg/view/view_test.go`, and `.agents/tasks/m2/T-0091-berserk-clients.md`.
    6. Quality gates: `go -C W tool task check` passed (gofmt, `go vet`, all package tests, `cardtown lint --all`). `go -C W tool task sim` passed clean (200 games, all invariants held).
  - Handed off to reviewer.
- 2026-10-28, lead-architect (wave B.2 run, session 3, D5): **Approved, `done`.** Merged into `autoloop/m2-wave-b.2` as 43a709d. Flash run `1d08ba24-0d55-4d73-911d-336819e485d2`: exit 0, no refusals (the JSON has no `denied_actions` entry, `cli.log` has no `soft-denying` line and the transcript no `permission check failed`). Review of `git diff autoloop/m2-wave-b.2...m2/T-0091-berserk-clients` (08d410c test, ecb4bfd code):
  - Scope: only `replay.go`, `view.go`, `play_screen.go`, `view_test.go` and this file changed. The `test(T-0091)` commit holds the one test edit and no code; `go test ./...` passes on it (re-run on the old code) and on the tip.
  - The code and test edits match the task's text line for line. The task named the test `TestPlayerView`; its real name is `TestView_Case1_Contents`. The line number and the edit were right, and the engineer edited that line. The implementation commit's `feat(engine)` prefix is wrong for client code; a nit, not worth a round.
  - Fences and reuse: no card name, no new shape, nothing added; the engine is untouched.
  - Gates re-run in a scratch worktree of the task branch, each with exit 0: case 2 prints nothing; case 3 lists exactly the five `pkg/engine` files left for part (d); `go tool task sim-compare -- autoloop/m2-wave-b.2` `[PASS]` (starter and stress byte-identical); `go tool task check` passes; `go tool task sim` clean, 200 games, P0 63.5% / P1 34.5% / draws 2.0% (unchanged).
  - Reviewer's events check: starter `-events` files from both builds have 104197 lines each and differ in exactly 200 lines, each game's `match-start` `revision`; with it set aside they are identical.
  - The merge's tree equals the task tip's, so the gates hold on the run branch.
