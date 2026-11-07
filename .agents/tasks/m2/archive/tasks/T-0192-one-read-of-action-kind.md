---
id: T-0192
title: One read of an action event's kind (directive #130, harness, flash)
kind: harness
tier: flash
wave: D.2
status: done
branch: m2/autoloop-wave-d.2/T-0192-one-read-of-action-kind
depends_on: [T-0191]
claims: [engine:Event.ActionKind, eventlog:actionOf, sim:Monitors.Event, test:TestEvent_ActionKind]
cards: []
requirements: []
---

## Goal
`emitAction` writes an action's kind as the first word of `Detail`, and `actionOf` and `Monitors.Event` each parse it back out of that text. A new method, `(Event).ActionKind() (ActionKind, bool)`, sits beside `emitAction` and is the one read, used by both (#130). `Detail` and the JSON don't change, so the sims stay byte-identical.

## Inputs
- Board directive #130 (structural review 9, approved by the user, 2026-11-06).
- Design note: `docs/architecture/m2-ability-model.md`, Deliverable 6h.
- `ActGenerate` is `ActionKind`'s zero value, so `Monitors.Event` must not switch on the kind alone: a false result would read as Generate. Edit 4 tests `ok` in every case.
- `actionOf` only sees events of kind `EventAction` (`replay.go:98` skips the rest), so the method's new kind check changes nothing for it. Its `unknown action %q` error stays.
- `pkg/sim/monitor_damage_test.go:181` and `TestEvents_Case4_EveryActionKind` also split `Detail` in tests. They are test setup, which D2.0 says doesn't count, and neither is edited. No other non-test code reads `Detail` for a kind (searched `.Detail` in `pkg` and `cmd`).
- Reuse search (D2.3): `pkg/engine` has no other read of an event's action kind.
- The lead dry-ran these exact edits in a throwaway worktree of the run branch at 9e27f6c. With only Edit 1, `go vet ./pkg/engine` prints `e.ActionKind undefined`. With all the edits, `gofmt -l pkg cmd` is clean, `go tool task check` and `go tool task sim` pass, and `go tool task sim-compare -- autoloop/m2-wave-d.2` prints `[PASS]`. Every expected output below is from that dry run.

## Acceptance Cases (harness tasks only; written by the lead)
W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **The test fails first.** With only Edit 1 in place: `go -C W vet ./pkg/engine` prints `e.ActionKind undefined`. Then commit Edit 1.
2. **The method reads the kind:** `go -C W test ./pkg/engine -run TestEvent_ActionKind` passes. It checks that every `ActionKind` comes back from the event `emitAction` wrote (including `Generate`, whose `Detail` has a second word), and that it is false for another event kind, an unknown word and an empty `Detail`.
3. **One parser is left:** `git -C W grep -n strings.Cut -- pkg/engine/events.go pkg/eventlog/replay.go pkg/sim/monitor.go` prints exactly one line, at `pkg/engine/events.go:175`.
4. **Replay and monitors are unchanged:** `go -C W test ./pkg/eventlog ./pkg/sim` passes.
5. **Sims byte-identical:** `go -C W tool task sim-compare -- autoloop/m2-wave-d.2` prints `[PASS]`.
6. **Only the listed edits:** `git -C W diff autoloop/m2-wave-d.2 --numstat -- pkg cmd` prints exactly these four lines:
   ```text
   17	0	pkg/engine/events.go
   22	0	pkg/engine/events_test.go
   4	11	pkg/eventlog/replay.go
   5	4	pkg/sim/monitor.go
   ```
7. `go -C W tool task check` and `go -C W tool task sim` pass.

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)
Line numbers are at 9e27f6c.

### 1. The test, in the `test(T-0192)` commit
**`pkg/engine/events_test.go`:** add this at the end of the file, with one blank line before it:
```go
// T-0192 (directive #130): ActionKind reads back the kind emitAction wrote, and
// nothing else.
func TestEvent_ActionKind(t *testing.T) {
	for k := ActGenerate; k <= ActEndTurn; k++ {
		s := newScene(t)
		s.g.emitAction(Action{Kind: k})
		e := s.g.Log[len(s.g.Log)-1]
		if got, ok := e.ActionKind(); !ok || got != k {
			t.Errorf("ActionKind of %q = %v, %v, want %v, true", e.Detail, got, ok, k)
		}
	}
	for _, e := range []Event{
		{Kind: EventDeploy, Detail: "Deploy"},
		{Kind: EventAction, Detail: "Dance"},
		{Kind: EventAction, Detail: ""},
	} {
		if got, ok := e.ActionKind(); ok {
			t.Errorf("ActionKind of %+v = %v, true, want false", e, got)
		}
	}
}
```
Commit only this file as `test(T-0192): ...`, after Acceptance Case 1. The `pkg/engine` tests don't compile until Edit 2, because the method doesn't exist yet. That is this commit's expected failure.

### 2. `pkg/engine/events.go`
1. In the import block, add `"strings"` as a new line after `"slices"` (line 7).
2. Add this after the closing `}` of `emitAction` (line 165), with one blank line on each side, so that it sits above the comment `// emitMatchStart is the first event of the game` (line 167):
```go
// ActionKind returns the kind an action event recorded, which emitAction wrote
// as the first word of Detail. It is false for another event kind or an unknown
// word.
func (e Event) ActionKind() (ActionKind, bool) {
	if e.Kind != EventAction {
		return 0, false
	}
	word, _, _ := strings.Cut(e.Detail, " ")
	for k := ActGenerate; k <= ActEndTurn; k++ {
		if k.String() == word {
			return k, true
		}
	}
	return 0, false
}
```

### 3. `pkg/eventlog/replay.go`
1. Delete line 7, `"strings"`. Nothing else in the file uses it.
2. Replace lines 255-268, from `// actionOf rebuilds the Action` through the closing `}` of `if !found {`, with:
```go
// actionOf rebuilds the Action an action event recorded.
func actionOf(e engine.Event) (engine.Action, error) {
	a := engine.Action{Player: engine.PlayerID(e.Player), Asset: e.Source, Target: e.Target, Card: e.Inst}
	kind, ok := e.ActionKind()
	if !ok {
		return a, fmt.Errorf("unknown action %q", e.Detail)
	}
	a.Kind = kind
```
The next line, `if a.Kind == engine.ActActivate {`, and the rest of the function stay.

### 4. `pkg/sim/monitor.go`
Replace lines 164-171, from `switch kind, _, _ := strings.Cut(e.Detail, " "); kind {` through the closing `}` of that switch, with the block below. `strings` stays imported, because lines 45 and 69 use it.
```go
		kind, ok := e.ActionKind()
		switch {
		case ok && kind == engine.ActAttack:
			m.checkAttack(e)
		case ok && kind == engine.ActGenerate:
			m.checkGenerate(e)
		case ok && kind == engine.ActConvert:
			m.checkConvert(e)
		}
```

No other file changes.

## Allowed Files
- `pkg/engine/events.go`
- `pkg/eventlog/replay.go`
- `pkg/sim/monitor.go`
- In the `test(T-0192)` commit only: `pkg/engine/events_test.go`, for the addition in Edit 1. No existing test changes.

## Done When
- [x] Cases 1–7 hold, with their commands and results in the Log
- [x] `go tool task sim-compare -- autoloop/m2-wave-d.2` passes (starter and stress sims byte-identical to the run branch)
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-07, lead-architect (wave D.2 run, session 6): wrote this task from board directive #130, next in board order, against the run branch at 9e27f6c. No `main` commits to sync. Not `WAVE-COMPLETE`: #130-#135 are `open: before the next Deliverables task`. Harness, so it skips `tests` and goes straight to `ready`; tier `flash`, because the edits are exact and were dry-run. Every command here is on the headless allow-list, one per line, with no pipes, `$`, double quotes or unquoted parentheses. Waiting directives checked: none has triggered (#6, #7: mode is `serial`; #15: no turn-boundary hook; #71, #91 (2), #106, #118 (b), #120, #128: no such card). Next: `kernel-engineer` on Flash.
- 2026-11-07, kernel-engineer (flash): implemented directive #130 (one read of an action event's kind).
  - Edit 1: added TestEvent_ActionKind to pkg/engine/events_test.go. Confirmed failure under Case 1 (`go vet ./pkg/engine` -> `e.ActionKind undefined`) and committed as test(T-0192).
  - Edit 2: added (Event).ActionKind() (ActionKind, bool) to pkg/engine/events.go, importing strings.
  - Edit 3: updated actionOf in pkg/eventlog/replay.go to use e.ActionKind(), removed unused strings import.
  - Edit 4: updated event handler in pkg/sim/monitor.go to switch on e.ActionKind().
  - Reused strings.Cut and ActionKind.String() matching surrounding engine patterns.
  - Verification:
    - Case 1: `go -C W vet ./pkg/engine` exited 1 with `e.ActionKind undefined` on Edit 1 alone.
    - Case 2: `go -C W test ./pkg/engine -run TestEvent_ActionKind` passed (`ok animus-mundi/pkg/engine 1.725s`).
    - Case 3: `git -C W grep -n strings.Cut -- pkg/engine/events.go pkg/eventlog/replay.go pkg/sim/monitor.go` printed single line `pkg/engine/events.go:175:	word, _, _ := strings.Cut(e.Detail, " ")`.
    - Case 4: `go -C W test ./pkg/eventlog ./pkg/sim` passed (`ok animus-mundi/pkg/eventlog 4.148s`, `ok animus-mundi/pkg/sim 52.322s`).
    - Case 5: `go -C W tool task sim-compare -- autoloop/m2-wave-d.2` passed with `[PASS] starter and stress sims match autoloop/m2-wave-d.2 byte for byte, exit codes included`.
    - Case 6: `git -C W diff autoloop/m2-wave-d.2 --numstat -- pkg cmd` printed exactly:
      17	0	pkg/engine/events.go
      22	0	pkg/engine/events_test.go
      4	11	pkg/eventlog/replay.go
      5	4	pkg/sim/monitor.go
    - Case 7: `go -C W tool task check` passed clean (fmt-check, vet, all tests, cardtown lint --all); `go -C W tool task sim` passed clean (200 games, 0 invariant violations).
  - Status moved to `review`. Ready for lead review.
- 2026-11-07, lead-architect (wave D.2 run, session 6, D5): **Approved, `done`.** Merged into `autoloop/m2-wave-d.2` as d1e1458. Flash run `72ee795a-404e-4906-b986-7b64b4c41ade`: exit 0, `status: SUCCESS`, no refusals (no `denied_actions` key; no `permission check failed` match in the transcript; its one `"status":"ERROR"` step, step 3, is a `file does not exist` from a read, not a refused command; `.err` holds only agy's "root agent idle" notice). Review of `git diff autoloop/m2-wave-d.2...m2/autoloop-wave-d.2/T-0192-one-read-of-action-kind` (9756e28):
  - Scope: only the Allowed Files and this file changed. `ef2a675` holds only `pkg/engine/events_test.go`, and no test file changed after it. No existing test was edited: the file gains only the addition in Edit 1.
  - Checked against the diff: the hunks are exactly Edits 2-4. `actionOf` and `Monitors.Event` both read the kind through `ActionKind`, and `Detail` and the JSON are untouched. Every monitor case tests `ok`, so a false result can't read as `ActGenerate`, the zero value.
  - Fences and placement: no card name, and the method sits beside `emitAction` as #130 asks. Rules fidelity: nothing rule-related changes.
  - Re-run by the reviewer in a scratch worktree of the task branch, all as expected: at `ef2a675` `go vet ./pkg/engine` exits 1 and prints `e.ActionKind undefined`; at the tip cases 2-6 hold, with one `strings.Cut` line at `pkg/engine/events.go:175` and the numstat exact; `gofmt -l pkg cmd` clean; `go tool task check` exit 0; `go tool task sim` exit 0, clean, 200 games, P0 78.5% / P1 19.0% / draws 2.5%, 47/50 and 34/50 main cards executable; `go tool task sim-compare -- autoloop/m2-wave-d.2` `[PASS]`.
  - The merge's `pkg/` and `cmd/` match the task tip's, because the run branch gained nothing after the branch point, so the gates hold on the run branch. Directive #130 is applied. #131 is next.
