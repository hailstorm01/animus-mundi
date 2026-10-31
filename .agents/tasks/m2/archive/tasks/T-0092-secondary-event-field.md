---
id: T-0092
title: Events carry no secondary target, and doc.go stops listing Berserk (directive #68 part (d1), harness, flash)
kind: harness
tier: flash
wave: B.2 (board directive #68 part (d), first half, with #72; the user's code review, before wave C)
status: done
branch: m2/T-0092-secondary-event-field
depends_on: [T-0091]
claims: [field:Event.Secondary-removed, file:pkg/engine/events.go, file:pkg/engine/doc.go]
cards: []
requirements: []
---

## Goal
`Event` has no `Secondary` field and no `secondary` JSON key, so an Attack's `action` and `place` events carry only the attacker (`Source`) and the defender (`Target`). `engine`'s package comment no longer lists Berserk among the keywords that run. This is the first half of directive #68 part (d). The second half, T-0093, removes `Action.Secondary` and `StackItem.Secondary`.

## Inputs
- Board directive #68 (user code review, 2026-10-28), part (d) and acceptance cases (a)–(c); #72 (3): no test is rewritten into an Overkill test.
- Decision record: [`docs/architecture/decisions/0006-attacks-carry-no-secondary-target.md`](../../../docs/architecture/decisions/0006-attacks-carry-no-secondary-target.md), which supersedes ADR-0005 for the `secondary` field. #68 has the lead write it before part (d) merges; it is in this task's commit on the run branch.
- Design note: `docs/architecture/m2-ability-model.md`, Deliverable 6b (wave B.2).
- Why part (d) is two tasks: it edits five non-test files (`actions.go`, `combat.go`, `doc.go`, `events.go`, `stack.go`), over the D2 size rule's four. This half takes the event field: its two writers (`emitAction`, `push`) stop reading the other two fields, so T-0093 can drop them without touching `events.go`.
- Why nothing observable changes: since T-0089 the engine rejects an Attack with a non-zero `Secondary`, so every `Event.Secondary` is 0, and `omitempty` leaves the key out of every event file. No test and no client reads `Event.Secondary` (T-0091 removed the last reader, `pkg/eventlog/replay.go`).

### Where it is today (at `autoloop/m2-wave-b.2` 78b42bc)
| Where | Today |
| :--- | :--- |
| `pkg/engine/events.go:155` | `emitAction`'s Attack case: `e.Source, e.Target, e.Secondary = a.Asset, a.Target, a.Secondary` |
| `pkg/engine/events.go:218-219` | `Event`'s `Secondary int` line (commented "action events: an Attack's Berserk secondary target (asset ID)"), aligned with the `Resource string` line under it |
| `pkg/engine/stack.go:56` | `push`'s `StackAttack` case: `e.Target, e.Secondary = item.Target, item.Secondary` |
| `pkg/engine/doc.go:14` | the keyword list ends `Ward, Versatile, Guardian, Armor N, Deploys Exhausted, Berserk N.` |

## Edits (lead decisions; copy exactly)

### 1. `pkg/engine/events.go`
- In `emitAction`'s `case ActAttack:`, the line becomes:
  ```go
  		e.Source, e.Target = a.Asset, a.Target
  ```
- In `Event`, delete the `Secondary` line. Run `gofmt`; the `Resource` line left on its own becomes:
  ```go
  	Resource string `json:"resource,omitempty"` // action events: a Generate's "Requisition" or "Tactics"
  ```

### 2. `pkg/engine/stack.go`
- In `push`'s `case StackAttack:`, the line becomes:
  ```go
  		e.Target = item.Target
  ```
  Leave `StackItem`'s `Secondary` field alone: it is T-0093's.

### 3. `pkg/engine/doc.go`
- Line 14 becomes:
  ```go
  //     Ward, Versatile, Guardian, Armor N, Deploys Exhausted.
  ```

No other non-test file changes.

## Acceptance Cases (harness tasks only; written by the lead)
No new tests and no test commit: this removes a field that is always 0, and the cases below show it is gone. Run each command on its own, pointed at your worktree (W). Record each command and its result in the Log. The reviewer re-runs them.
1. **Existing tests unchanged** (#68 case (a)): `git -C W diff autoloop/m2-wave-b.2 -- '*_test.go'` prints nothing, and `go -C W test ./...` passes.
2. **Sims byte-identical** (#68 case (b)): `go -C W tool task sim-compare -- autoloop/m2-wave-b.2` prints `[PASS]`.
3. **No event field:** `git -C W grep -n -i secondary -- pkg/engine/events.go pkg/engine/doc.go` prints nothing (exit code 1).
4. **What is left is T-0093's:** `git -C W grep -l -i -E "berserk|secondary|cleave" -- pkg cmd ":!pkg/toolchain/linter"` lists exactly `pkg/engine/actions.go`, `pkg/engine/combat.go` and `pkg/engine/stack.go`.
5. **Only these edits:** `git -C W diff autoloop/m2-wave-b.2 --stat` lists only `pkg/engine/events.go`, `pkg/engine/stack.go`, `pkg/engine/doc.go` and this task file.
6. `go -C W tool task check` and `go -C W tool task sim` pass.

The reviewer also compares a starter `-events` file from both builds (#68 case (b)); it must differ only in each game's `match-start` `revision` stamp. That runs built binaries, so it is the reviewer's check, not the engineer's.

## Allowed Files
- `pkg/engine/events.go`
- `pkg/engine/stack.go`, for the one line in `push`
- `pkg/engine/doc.go`

## Done When
- [x] Cases 1–6 hold, with their commands and results in the Log
- [x] `go tool task sim-compare -- autoloop/m2-wave-b.2` passes (starter and stress sims byte-identical to the run branch)
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks

## Log
- 2026-10-28, kernel-engineer (T-0092): implemented first half of directive #68 part (d).
  - Changes:
    - `pkg/engine/events.go`: in `emitAction`'s `ActAttack` case, dropped assignment to `e.Secondary`; in `Event` struct, deleted `Secondary int` field and kept `Resource string`.
    - `pkg/engine/stack.go`: in `push`'s `StackAttack` case, dropped assignment to `e.Secondary`. Left `StackItem.Secondary` untouched for T-0093.
    - `pkg/engine/doc.go`: removed `Berserk N.` from line 14 keywords list.
  - Reused: no new shapes or fields needed; purely removed `Event.Secondary` field and doc reference.
  - Acceptance Cases & verification results:
    1. `git -C W diff autoloop/m2-wave-b.2 -- "*_test.go"` printed nothing (clean), and `go -C W test ./...` passed.
    2. `go -C W tool task sim-compare -- autoloop/m2-wave-b.2` printed `[PASS] starter and stress sims match autoloop/m2-wave-b.2 byte for byte, exit codes included`.
    3. `git -C W grep -n -i secondary -- pkg/engine/events.go pkg/engine/doc.go` exited 1 with no output.
    4. `git -C W grep -l -i -E "berserk|secondary|cleave" -- pkg cmd ":!pkg/toolchain/linter"` printed exactly `pkg/engine/actions.go`, `pkg/engine/combat.go`, `pkg/engine/stack.go`.
    5. `git -C W diff autoloop/m2-wave-b.2 --stat` listed only `pkg/engine/doc.go`, `pkg/engine/events.go`, `pkg/engine/stack.go`, and `.agents/tasks/m2/T-0092-secondary-event-field.md`.
    6. `go -C W tool task check` passed all checks (gofmt, vet, test, linters). `go -C W tool task sim` passed clean (200 matches, all invariants held).
  - Commit: `feat(engine): events carry no secondary target and doc.go stops listing Berserk (T-0092)` (00f0376). Handing off to `lead-architect` for review.
- 2026-10-28, lead-architect (wave B.2 run, session 4): synced `main`'s 90789c4 (lead D7: removals name literal paths) into the run branch as 78b42bc. Wrote this task from board directive #68 part (d), the next of wave B.2's blocking directives in board order. Part (d) edits five non-test files, so under the D2 size rule it is two tasks: this one (the `secondary` event field and `doc.go`) and T-0093 (`Action.Secondary`, `StackItem.Secondary`), written in the same commit and `queued`. Each half compiles and passes on its own. Wrote decision record 0006, which supersedes ADR-0005 for the `secondary` field, in this commit, as #68 asks before part (d) merges; ADR-0005's status line and the decisions index point to it. Harness, so it goes straight to `ready`; tier `flash`, the harness default, and every command here is on the headless allow-list. Waiting directives checked: none has triggered (#8: this task doesn't touch `deck.go`; #15, #20 (b), #58, #59 (2), #61, #71: no such hook, shape, claim or card here; #6, #7: mode is `serial`). #69 and #70 stay open for the sessions after T-0093. Next: `kernel-engineer` on Flash.
- 2026-10-28, lead-architect (wave B.2 run, session 4, D5): **Approved, `done`.** Merged into `autoloop/m2-wave-b.2` as 2c78e88. Flash run `b8a36176-70b7-4ce2-8aca-319afcc53f3c`: exit 0, no refusals (the JSON has no `denied_actions` entry, `cli.log` has no `soft-denying` line for it and the transcript no `permission check failed`). Review of `git diff autoloop/m2-wave-b.2...m2/T-0092-secondary-event-field` (00f0376):
  - Scope: only `events.go`, `stack.go`, `doc.go` and this file changed, and the three code edits match the task's text line for line. No test file changed, as the task asked (no test commit).
  - Fences and reuse: no card name, no new shape, nothing added; `StackItem.Secondary` and `Action.Secondary` are left for T-0093.
  - Gates re-run in a scratch worktree of the task branch, each with exit 0: case 1 (no `_test.go` diff); case 3 prints nothing; case 4 lists exactly `actions.go`, `combat.go`, `stack.go`; `go tool task sim-compare -- autoloop/m2-wave-b.2` `[PASS]` (starter and stress byte-identical); `go tool task check` passes; `go tool task sim` clean, 200 games, P0 63.5% / P1 34.5% / draws 2.0% (unchanged).
  - Reviewer's events check: starter `-events` files from both builds have 104197 lines each and differ in exactly 200 lines, each game's `match-start` `revision`. The base file holds no `secondary` key, as ADR 0006 says.
  - Nits, not worth a round: the engineer's status commit is `chore(task): start T-0092` rather than `chore(tasks): T-0092 in-progress`, and its Log entry sits above the lead's.
  - The merge's tree equals the task tip's, so the gates hold on the run branch.
