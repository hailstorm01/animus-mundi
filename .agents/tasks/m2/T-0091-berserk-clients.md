---
id: T-0091
title: The clients stop reading an Attack's secondary target (directive #68 part (c), harness, flash)
kind: harness
tier: flash
wave: B.2 (board directive #68 part (c), with #72; the user's code review, before wave C)
status: in-progress
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
- [ ] Cases 1–6 hold, with their commands and results in the Log
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks

## Log
- 2026-10-28, lead-architect (wave B.2 run, session 3): wrote the task from board directive #68 part (c), the next of wave B.2's blocking directives in board order. Harness, so it goes straight to `ready`; tier `flash`, the harness default (#68 names `sonnet` only for part (a)), and every command here is on the headless allow-list. Lead decisions: (1) no new test (see Tests); (2) `actionLabel` has no test today and gets none, since its remaining Attack label is unchanged; (3) the `-events` comparison is the reviewer's. Applied #72's design-note part in the same commit: the Overkill settling tests are recorded as an M3 item (design note Open Question 12). Waiting directives checked: none has triggered (#8: this task doesn't touch `deck.go`; #15, #20 (b), #58, #59 (2), #61, #71: no such hook, shape, claim or card here; #6, #7: mode is `serial`). Part (d), #69 and #70 stay open for the next sessions. Next: `kernel-engineer` on Flash.
