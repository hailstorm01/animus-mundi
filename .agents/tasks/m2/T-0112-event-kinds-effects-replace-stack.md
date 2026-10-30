---
id: T-0112
title: effects.go, replace.go and stack.go emit with the event kind constants (directive #86 part 3 of 5, harness, flash, pure move)
kind: harness
tier: flash
wave: G.2 (board directive #86, structural review 6, finding 5)
status: ready
branch: m2/T-0112-event-kinds-effects-replace-stack
depends_on: [T-0111]
claims: []
cards: []
requirements: []
---

## Goal
The event emits in `pkg/engine/effects.go`, `replace.go` and `stack.go` name their kind by the constants T-0110 declared in `events.go` (`EventLeave`, `EventReplace`, `EventPlace`, `EventDeploy`, `EventMarch`), not by a string literal. A pure move: same tests, same sim output. After this task no non-test file in `pkg/engine` writes an event kind as a literal.

## Inputs
- Board directive #86 (structural review 6, approved by the user, 2026-10-30), part 3 of the five-task split recorded in its status cell: T-0110 (done: the constants, and `events.go`, `choice.go`, `sweep.go`, `zones.go`), T-0111 (done: `actions.go`, `combat.go`, `damage.go`), **T-0112 (this task)**: `effects.go`, `replace.go`, `stack.go`; T-0113: `pkg/sim`; T-0114: `pkg/view`, `pkg/eventlog`, and #86 (c) over all of `pkg/` and `cmd/cardtown`.
- ADR-0005 (`docs/architecture/decisions/0005-structured-game-events-and-observer.md`): the kind strings are its schema. `TestEventKinds_ValuesAreTheSchema` (T-0110) pins each constant to its value.
- Design note: `docs/architecture/m2-ability-model.md`, Deliverable 6e (wave G.2).
- Why it is a pure move: each replaced literal becomes the constant with the same value, so every emitted `Event` is unchanged. The lead dry-ran these exact edits on the run branch at 7da8c79: `gofmt -l pkg` printed nothing, `go tool task check` passed, `cardtown lint fences` passed, and `go tool task sim-compare -- autoloop/m2-wave-g.2` printed `[PASS]`.

### Where it is today (at `autoloop/m2-wave-g.2` 7da8c79)
| Where | Today |
| :--- | :--- |
| `pkg/engine/effects.go:196` | `g.logEvent(Event{Kind: "leave", …` |
| `pkg/engine/replace.go:110` | `g.logEvent(Event{Kind: "replace", …` |
| `pkg/engine/stack.go:46` | `e := Event{Kind: "place", …` |
| `pkg/engine/stack.go:108` | `g.logEvent(Event{Kind: "deploy", …` |
| `pkg/engine/stack.go:123` | `g.logEvent(Event{Kind: "march", …` |

## Edits (lead decisions; copy exactly)
On each line below, replace only the quoted kind with the constant; nothing else on the line changes.

### 1. `pkg/engine/effects.go`
- Line 196: `Kind: "leave"` becomes `Kind: EventLeave`.

### 2. `pkg/engine/replace.go`
- Line 110: `Kind: "replace"` becomes `Kind: EventReplace`.
- **Leave line 135 alone.** `Kind: DecideReplacement` is a `Decision`'s kind, not an event kind.

### 3. `pkg/engine/stack.go`
- Line 46: `Kind: "place"` becomes `Kind: EventPlace`.
- Line 108: `Kind: "deploy"` becomes `Kind: EventDeploy`.
- Line 123: `Kind: "march"` becomes `Kind: EventMarch`.

Nothing else changes in these three files, and no other file changes except this task file. `pkg/sim`, `pkg/view` and `pkg/eventlog` are T-0113's and T-0114's: leave them alone.

## Acceptance Cases (harness tasks only; written by the lead)
No new tests: a pure move, and T-0110's `TestEventKinds_ValuesAreTheSchema` already pins the values. Run each command on its own, pointed at your worktree (W). Record each command and its result in the Log. The reviewer re-runs them.
1. **Existing tests unchanged** (#86 (a)): `git -C W diff autoloop/m2-wave-g.2 -- '*_test.go'` prints nothing, and `go -C W test ./...` passes.
2. **Sims byte-identical** (#86 (b)): `go -C W tool task sim-compare -- autoloop/m2-wave-g.2` prints `[PASS]`.
3. **No kind literal left in the engine** (#86 (c), for `pkg/engine`): `git -C W grep -n -E -e "Kind: [^A-Za-z]" -e "logf[(][^A-Za-z]" -- pkg/engine/*.go ':!*_test.go'` prints nothing (git grep exits 1 when nothing matches; that is the expected result).
4. **The five emits use the constants:** `git -C W grep -n -E "Kind: Event" -- pkg/engine/effects.go pkg/engine/replace.go pkg/engine/stack.go` prints exactly five lines: `pkg/engine/effects.go:196` (`EventLeave`), `pkg/engine/replace.go:110` (`EventReplace`), `pkg/engine/stack.go:46` (`EventPlace`), `:108` (`EventDeploy`) and `:123` (`EventMarch`).
5. **Moved, not changed:** `git -C W diff autoloop/m2-wave-g.2 --numstat -- pkg` prints exactly these three lines: `1	1	pkg/engine/effects.go`, `1	1	pkg/engine/replace.go` and `3	3	pkg/engine/stack.go`.
6. **Fences:** `go -C W run ./cmd/cardtown lint fences` exits 0.
7. **Only these edits:** `git -C W diff autoloop/m2-wave-g.2 --stat` lists only the three files in case 5 and this task file.
8. `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/engine/effects.go`
- `pkg/engine/replace.go`
- `pkg/engine/stack.go`

## Done When
- [ ] Cases 1–8 hold, with their commands and results in the Log
- [ ] `go tool task sim-compare -- autoloop/m2-wave-g.2` passes (starter and stress sims byte-identical to the run branch)
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks

## Log
- 2026-10-30, lead-architect (wave G.2 run, session 6): wrote the task from board directive #86, part 3 of the five-task split T-0110 recorded; T-0111 is `done` (merge 10d99e8). Harness, so it skips `tests` and goes straight to `ready`; tier `flash` as #86 gives. No new test, for the same reason as T-0111: the unchanged tests and `sim-compare` pin a pure move, and T-0110's test pins each constant's value. A `git grep` over every non-test engine file for a quoted kind (and a `grep` for every `Kind` in these three files) found exactly the five lines above; `replace.go:135`'s `DecideReplacement` is a decision kind, so it stays. Case 3 now covers all of `pkg/engine`, since this is the last engine part; it uses `[^A-Za-z]` so it matches a quote but not a `Stack…`, `Decide…` or `Event…` constant. Every command here is on the headless allow-list, one per line, with no `|`, `$` or unquoted parentheses. The edits were dry-run on the run branch (Inputs). `main` (281e7d8) has no commits the run branch lacks, so no sync. Waiting directives checked: none has triggered (#6, #7: mode is `serial`; #8: no `deck.go`; #15, #20 (b), #58, #59 (2), #61, #71: no such hook, claim or card here). #87 (c) stays `open: before the next Deliverables task`, after #86. Next: `kernel-engineer` on Flash.
