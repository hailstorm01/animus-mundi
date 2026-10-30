---
id: T-0113
title: pkg/sim matches events by the event kind constants, and actions by engine.ActAttack.String() and its siblings (directive #86 part 4 of 5, harness, flash, pure move)
kind: harness
tier: flash
wave: G.2 (board directive #86, structural review 6, finding 5)
status: in-progress
branch: m2/T-0113-event-kinds-sim
depends_on: [T-0112]
claims: []
cards: []
requirements: []
---

## Goal
The event matches in `pkg/sim/monitor.go`, `monitor_entry.go`, `monitor_targeting.go` and `playout.go` name the event kind by the constants T-0110 declared in `pkg/engine/events.go`, not by a string literal. The action monitor names the action kind by `engine.ActAttack.String()`, `engine.ActGenerate.String()` and `engine.ActConvert.String()`, and the targeting monitor names the decision kind by `engine.DecideTarget.String()`. A pure move: same tests, same sim output. After this task no non-test file in `pkg/sim` writes an event, action or decision kind as a literal.

## Inputs
- Board directive #86 (structural review 6, approved by the user, 2026-10-30), part 4 of the five-task split recorded in its status cell: T-0110 (done: the constants, and `events.go`, `choice.go`, `sweep.go`, `zones.go`), T-0111 (done: `actions.go`, `combat.go`, `damage.go`), T-0112 (done: `effects.go`, `replace.go`, `stack.go`), **T-0113 (this task)**: `pkg/sim`; T-0114: `pkg/view`, `pkg/eventlog`, and #86 (c) over all of `pkg/` and `cmd/cardtown`.
- #86's third bullet: "the action monitor compares with `engine.ActAttack.String()` and its siblings."
- ADR-0005 (`docs/architecture/decisions/0005-structured-game-events-and-observer.md`): the kind strings are its schema. `TestEventKinds_ValuesAreTheSchema` (T-0110) pins each event constant to its value. `ActionKind.String` (`pkg/engine/actions.go:31`) and `DecisionKind.String` (`pkg/engine/choice.go:21`) are what `emitAction` and `emitDecision` write into `Detail` and `DecisionRecord.Kind`, so comparing with them is comparing with the emitted value.
- Design note: `docs/architecture/m2-ability-model.md`, Deliverable 6e (wave G.2).
- Why it is a pure move: each replaced literal becomes a constant or `String()` call with the same value, so every comparison gives the same result. The lead dry-ran these exact edits on the run branch at e0be977: `gofmt -l pkg` printed nothing, `go vet ./pkg/sim` passed, `go tool task check` passed, `cardtown lint fences` passed, and `go tool task sim-compare -- autoloop/m2-wave-g.2` printed `[PASS]`.

### Where it is today (at `autoloop/m2-wave-g.2` e0be977)
| Where | Today |
| :--- | :--- |
| `pkg/sim/monitor.go:145` | `if e.Kind != "decision" {` |
| `pkg/sim/monitor.go:149–163` | `case "decision":`, `case "combat":`, `case "damage":`, `case "match-start":`, `case "turn-start":`, `case "deploy", "scratch":`, `case "action":` |
| `pkg/sim/monitor.go:165–169` | `case "Attack":`, `case "Generate":`, `case "Convert":` (the action kind cut out of `Detail`) |
| `pkg/sim/monitor_entry.go:13` | `if e.Kind == "scratch" {` |
| `pkg/sim/monitor_targeting.go:22` | `if e.Decision == nil \|\| e.Decision.Kind != "Target" {` |
| `pkg/sim/monitor_targeting.go:56` | `if e.Kind != "place" \|\| e.Player != t.player {` |
| `pkg/sim/playout.go:32` | `if e.Kind == "invalid-answer" {` |

## Edits (lead decisions; copy exactly)
On each line below, replace only the quoted kind; nothing else on the line changes. All four files already import `animus-mundi/pkg/engine`, so no import changes.

### 1. `pkg/sim/monitor.go` (11 lines)
- Line 145: `e.Kind != "decision"` becomes `e.Kind != engine.EventDecision`.
- Line 149: `case "decision":` becomes `case engine.EventDecision:`.
- Line 151: `case "combat":` becomes `case engine.EventCombat:`.
- Line 153: `case "damage":` becomes `case engine.EventDamage:`.
- Line 155: `case "match-start":` becomes `case engine.EventMatchStart:`.
- Line 159: `case "turn-start":` becomes `case engine.EventTurnStart:`.
- Line 161: `case "deploy", "scratch":` becomes `case engine.EventDeploy, engine.EventScratch:`.
- Line 163: `case "action":` becomes `case engine.EventAction:`.
- Line 165: `case "Attack":` becomes `case engine.ActAttack.String():`.
- Line 167: `case "Generate":` becomes `case engine.ActGenerate.String():`.
- Line 169: `case "Convert":` becomes `case engine.ActConvert.String():`.

### 2. `pkg/sim/monitor_entry.go` (1 line)
- Line 13: `e.Kind == "scratch"` becomes `e.Kind == engine.EventScratch`.

### 3. `pkg/sim/monitor_targeting.go` (2 lines)
- Line 22: `e.Decision.Kind != "Target"` becomes `e.Decision.Kind != engine.DecideTarget.String()`.
- Line 56: `e.Kind != "place"` becomes `e.Kind != engine.EventPlace`.

### 4. `pkg/sim/playout.go` (1 line)
- Line 32: `e.Kind == "invalid-answer"` becomes `e.Kind == engine.EventInvalidAnswer`.

Nothing else changes in these four files, and no other file changes except this task file. The monitor names (`monAttack` … `monTargeting`), rule IDs and message strings are not event kinds: leave them alone. `pkg/view` and `pkg/eventlog` are T-0114's: leave them alone.

## Acceptance Cases (harness tasks only; written by the lead)
No new tests: a pure move, and T-0110's `TestEventKinds_ValuesAreTheSchema` already pins the event values. Run each command on its own, pointed at your worktree (W). Record each command and its result in the Log. The reviewer re-runs them.
1. **Existing tests unchanged** (#86 (a)): `git -C W diff autoloop/m2-wave-g.2 -- '*_test.go'` prints nothing, and `go -C W test ./...` passes.
2. **Sims byte-identical** (#86 (b)): `go -C W tool task sim-compare -- autoloop/m2-wave-g.2` prints `[PASS]`.
3. **No kind literal left in `pkg/sim`** (#86 (c), for `pkg/sim`): `git -C W grep -n -E -e "Kind [!=]= [^A-Za-z]" -e "case [^A-Za-z0-9]" -- pkg/sim ':!*_test.go'` prints nothing (git grep exits 1 when nothing matches; that is the expected result). Before the edits it prints exactly the 15 lines in the table above; `sim.go`'s `case 0:` and `case 1:` are not matched.
4. **The fifteen matches use the constants:** `git -C W grep -n -E -e "engine[.]Event[A-Z]" -e "engine[.]Act[A-Z]" -e "engine[.]Decide[A-Z]" -- pkg/sim/monitor.go pkg/sim/monitor_entry.go pkg/sim/monitor_targeting.go pkg/sim/playout.go` prints exactly fifteen lines: `pkg/sim/monitor.go` lines 145, 149, 151, 153, 155, 159, 161, 163, 165, 167 and 169, `pkg/sim/monitor_entry.go:13`, `pkg/sim/monitor_targeting.go:22` and `:56`, and `pkg/sim/playout.go:32`. (Before the edits it prints nothing.)
5. **Moved, not changed:** `git -C W diff autoloop/m2-wave-g.2 --numstat -- pkg` prints exactly these four lines: `11	11	pkg/sim/monitor.go`, `1	1	pkg/sim/monitor_entry.go`, `2	2	pkg/sim/monitor_targeting.go` and `1	1	pkg/sim/playout.go`.
6. **Fences:** `go -C W run ./cmd/cardtown lint fences` exits 0.
7. **Only these edits:** `git -C W diff autoloop/m2-wave-g.2 --stat` lists only the four files in case 5 and this task file.
8. `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/sim/monitor.go`
- `pkg/sim/monitor_entry.go`
- `pkg/sim/monitor_targeting.go`
- `pkg/sim/playout.go`

## Done When
- [ ] Cases 1–8 hold, with their commands and results in the Log
- [ ] `go tool task sim-compare -- autoloop/m2-wave-g.2` passes (starter and stress sims byte-identical to the run branch)
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks

## Log
- 2026-10-30, lead-architect (wave G.2 run, session 7): wrote the task from board directive #86, part 4 of the five-task split T-0110 recorded; T-0112 is `done` (merge c7f9c81). Harness, so it skips `tests` and goes straight to `ready`; tier `flash` as #86 gives. No new test, for the same reason as T-0111 and T-0112: the unchanged tests and `sim-compare` pin a pure move, and T-0110's test pins each event constant's value. A `git grep` over every non-test `pkg/sim` file for a kind compared or switched on by literal found exactly the fifteen lines above, all in the four files #86 named. One lead decision beyond #86's wording: `monitor_targeting.go:22` matches `DecisionRecord.Kind` against `"Target"`, the same scheme as the action kinds in `Detail` (a kind string the engine writes with a `String()` method, matched by literal), so it takes `engine.DecideTarget.String()` like #86's third bullet does for actions; it is in an allowed file and keeps case 3 clean. Case 3 uses `[^A-Za-z0-9]` after `case` so `sim.go`'s numeric `case 0:` / `case 1:` don't match. Every command here is on the headless allow-list, one per line, with no `|`, `$` or unquoted parentheses. The edits were dry-run on the run branch (Inputs). `main` (37d9c89) has no commits the run branch lacks, so no sync. Waiting directives checked: none has triggered (#6, #7: mode is `serial`; #8: no `deck.go`; #15, #20 (b), #58, #59 (2), #61, #71: no such hook, claim or card here). #87 (c) stays `open: before the next Deliverables task`, after #86. Next: `kernel-engineer` on Flash.
