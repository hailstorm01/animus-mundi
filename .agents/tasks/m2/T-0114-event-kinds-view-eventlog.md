---
id: T-0114
title: pkg/view and pkg/eventlog match events by the event kind constants, so no event kind is a literal outside events.go (directive #86 part 5 of 5, harness, flash, pure move)
kind: harness
tier: flash
wave: G.2 (board directive #86, structural review 6, finding 5)
status: in-progress
branch: m2/T-0114-event-kinds-view-eventlog
depends_on: [T-0113]
claims: []
cards: []
requirements: []
---

## Goal
The event matches in `pkg/view/log.go`, `pkg/eventlog/eventlog.go` and `pkg/eventlog/replay.go` name the event kind by the constants T-0110 declared in `pkg/engine/events.go`, not by a string literal. A pure move: same tests, same sim output. After this task no non-test Go file under `pkg/` or `cmd/cardtown` writes an event kind as a string literal, except the constants in `events.go` (#86 (c)), and directive #86 is applied.

## Inputs
- Board directive #86 (structural review 6, approved by the user, 2026-10-30), part 5 of the five-task split recorded in its status cell: T-0110 (done: the constants, and `events.go`, `choice.go`, `sweep.go`, `zones.go`), T-0111 (done: `actions.go`, `combat.go`, `damage.go`), T-0112 (done: `effects.go`, `replace.go`, `stack.go`), T-0113 (done: `pkg/sim`), **T-0114 (this task)**: `pkg/view/log.go`, `pkg/eventlog/eventlog.go`, `pkg/eventlog/replay.go`, and case (c) over all of `pkg/` and `cmd/cardtown`.
- ADR-0005 (`docs/architecture/decisions/0005-structured-game-events-and-observer.md`): the kind strings are its schema. `TestEventKinds_ValuesAreTheSchema` (T-0110) pins each constant to its value.
- Design note: `docs/architecture/m2-ability-model.md`, Deliverable 6e (wave G.2).
- Why it is a pure move: each replaced literal becomes the constant with the same value, so every comparison gives the same result. The lead dry-ran these exact edits on the run branch at 1fca5bf: `gofmt -l pkg cmd` printed nothing, `go vet ./pkg/view ./pkg/eventlog` passed, `cardtown lint fences` passed, `go tool task check` passed, and `go tool task sim-compare -- autoloop/m2-wave-g.2` printed `[PASS]`.
- What else the lead checked for #86 (c): `cmd/cardtown` matches no event kind (its `.Kind` reads are decision, action and stack kinds, already by constant). `pkg/eventlog`'s `actionOf` and decision check already compare with `ActionKind.String()` and `DecisionKind.String()`. The remaining `case "…":` lines under `pkg/` and `cmd/cardtown` switch on CLI words, card types, zones, audit severities and build settings, not event kinds. `damage.go`'s `kind = "combat"` is the damage type in the message text, not an event kind. The comments on `Event.Start`, `Snapshot` and `End` (`// Kind "match-start"` …) are comments: leave them.

### Where it is today (at `autoloop/m2-wave-g.2` 1fca5bf)
| Where | Today |
| :--- | :--- |
| `pkg/view/log.go:25` | `case e.Kind == "match-start" && e.Start != nil:` |
| `pkg/view/log.go:27` | `case e.Kind == "turn-start":` |
| `pkg/eventlog/eventlog.go:32` | `case e.Kind == "turn-start":` |
| `pkg/eventlog/replay.go:30` | `if len(events) == 0 \|\| events[0].Kind != "match-start" \|\| events[0].Start == nil {` |
| `pkg/eventlog/replay.go:98` | `if e.Kind != "action" {` |

## Edits (lead decisions; copy exactly)
On each line below, replace only the quoted kind; nothing else on the line changes. All three files already import `animus-mundi/pkg/engine`, so no import changes.

### 1. `pkg/view/log.go` (2 lines)
- Line 25: `e.Kind == "match-start"` becomes `e.Kind == engine.EventMatchStart`.
- Line 27: `e.Kind == "turn-start"` becomes `e.Kind == engine.EventTurnStart`.

### 2. `pkg/eventlog/eventlog.go` (1 line)
- Line 32: `e.Kind == "turn-start"` becomes `e.Kind == engine.EventTurnStart`.

### 3. `pkg/eventlog/replay.go` (2 lines)
- Line 30: `events[0].Kind != "match-start"` becomes `events[0].Kind != engine.EventMatchStart`. The error message on line 31 (`… must start with a match-start event`) is text: leave it.
- Line 98: `e.Kind != "action"` becomes `e.Kind != engine.EventAction`.

Nothing else changes in these three files, and no other file changes except this task file.

## Acceptance Cases (harness tasks only; written by the lead)
No new tests: a pure move, and T-0110's `TestEventKinds_ValuesAreTheSchema` already pins the event values. Run each command on its own, pointed at your worktree (W). Record each command and its result in the Log. The reviewer re-runs them.
1. **Existing tests unchanged** (#86 (a)): `git -C W diff autoloop/m2-wave-g.2 -- '*_test.go'` prints nothing, and `go -C W test ./...` passes.
2. **Sims byte-identical** (#86 (b)): `go -C W tool task sim-compare -- autoloop/m2-wave-g.2` prints `[PASS]`.
3. **No event kind literal left under `pkg/` or `cmd/cardtown`** (#86 (c)): `git -C W grep -n -E -e "Kind [!=]= [^A-Za-z]" -e "Kind: [^A-Za-z]" -e "logf[(][^A-Za-z]" -- pkg cmd/cardtown ':!*_test.go'` prints nothing (git grep exits 1 when nothing matches; that is the expected result). Before the edits it prints exactly the five lines in the table above. The three patterns cover a kind compared by literal, an `Event{Kind: …}` literal and a `logf` kind argument; the constants in `events.go` (`EventMatchStart    = "match-start"` …) match none of them.
4. **The five matches use the constants:** `git -C W grep -n -E -e "engine[.]Event[A-Z]" -- pkg/view/log.go pkg/eventlog/eventlog.go pkg/eventlog/replay.go` prints exactly five lines: `pkg/eventlog/eventlog.go:32`, `pkg/eventlog/replay.go:30` and `:98`, and `pkg/view/log.go:25` and `:27`. (Before the edits it prints nothing.)
5. **Moved, not changed:** `git -C W diff autoloop/m2-wave-g.2 --numstat -- pkg` prints exactly these three lines: `1	1	pkg/eventlog/eventlog.go`, `2	2	pkg/eventlog/replay.go` and `2	2	pkg/view/log.go`.
6. **Fences:** `go -C W run ./cmd/cardtown lint fences` exits 0.
7. **Only these edits:** `git -C W diff autoloop/m2-wave-g.2 --stat` lists only the three files in case 5 and this task file.
8. `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/view/log.go`
- `pkg/eventlog/eventlog.go`
- `pkg/eventlog/replay.go`

## Done When
- [ ] Cases 1–8 hold, with their commands and results in the Log
- [ ] `go tool task sim-compare -- autoloop/m2-wave-g.2` passes (starter and stress sims byte-identical to the run branch)
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks

## Log
- 2026-10-30, lead-architect (wave G.2 run, session 8): wrote the task from board directive #86, part 5 of the five-task split T-0110 recorded; T-0113 is `done` (merge f206224). Harness, so it skips `tests` and goes straight to `ready`; tier `flash` as #86 gives. No new test, for the same reason as T-0111 to T-0113: the unchanged tests and `sim-compare` pin a pure move, and T-0110's test pins each constant's value. A `git grep` over every non-test Go file under `pkg/` and `cmd/cardtown` for a kind compared by literal, an `Event{Kind: …}` literal or a `logf` kind argument found exactly the five lines above, all in the three files #86's split gives T-0114 (Inputs lists what else was checked). Every command here is on the headless allow-list, one per line, with no `|`, `$` or unquoted parentheses. The edits were dry-run on the run branch (Inputs). `main` (37d9c89) has no commits the run branch lacks, so no sync. Waiting directives checked: none has triggered (#6, #7: mode is `serial`; #8: no `deck.go`; #15, #20 (b), #58, #59 (2), #61, #71: no such hook, claim or card here). When this task is `done`, #86 is `applied`, and #87 (c) is the last `open: before` directive of the wave. Next: `kernel-engineer` on Flash.
