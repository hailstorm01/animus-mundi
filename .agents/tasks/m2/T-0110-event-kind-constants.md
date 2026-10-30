---
id: T-0110
title: Event kinds are named constants in events.go, and events.go, choice.go, sweep.go and zones.go emit with them (directive #86 part 1 of 5, harness, flash, pure move)
kind: harness
tier: flash
wave: G.2 (board directive #86, structural review 6, finding 5)
status: ready
branch: m2/T-0110-event-kind-constants
depends_on: [T-0109]
claims: [const:EventMatchStart, const:EventTurnStart, const:EventAction, const:EventDecision, const:EventInvalidAnswer, const:EventGenerate, const:EventPlace, const:EventDeploy, const:EventMarch, const:EventScratch, const:EventAttack, const:EventFizzle, const:EventCombat, const:EventDamage, const:EventReplace, const:EventDestroy, const:EventLeave, const:EventDraw, const:EventDiscard, const:EventResult, const:EventMatchEnd]
cards: []
requirements: []
---

## Goal
Every event kind is one exported untyped string constant in `pkg/engine/events.go` (`EventAction = "action"` … `EventMatchEnd = "match-end"`), whose value is the ADR-0005 event schema and never changes. This task adds the constants and makes the emitters in four engine files use them. Today `Event.Kind` is written as a bare literal at each emit and each client match, and T-0102 added `leave` with nothing listing it (#86). A pure move: same tests, same sim output.

## Inputs
- Board directive #86 (structural review 6, approved by the user, 2026-10-30). The directive splits the work under the D2 size rule: the constants and the engine's emitters first, then the clients. 17 non-test files use a kind literal, so it is five tasks of at most four non-test files: **T-0110 (this task)**: the constants, and the emits in `events.go`, `choice.go`, `sweep.go`, `zones.go`; T-0111: `actions.go`, `combat.go`, `damage.go`; T-0112: `effects.go`, `replace.go`, `stack.go`; T-0113: `pkg/sim` (`monitor.go`, `monitor_entry.go`, `monitor_targeting.go`, `playout.go`, with the action monitor's `engine.ActAttack.String()` and siblings); T-0114: `pkg/view/log.go`, `pkg/eventlog/eventlog.go`, `pkg/eventlog/replay.go`, and #86 (c) over all of `pkg/` and `cmd/cardtown`. `cmd/cardtown` matches no event kind today.
- ADR-0005 (`docs/architecture/decisions/0005-structured-game-events-and-observer.md`): the kind strings are its schema. Not one value changes.
- Design note: `docs/architecture/m2-ability-model.md`, Deliverable 6e (wave G.2).
- Why it is a pure move: each replaced literal becomes a constant with the same value, so every emitted `Event` is unchanged. The lead dry-ran these exact edits on the run branch at 0ca1048: `gofmt -l pkg` printed nothing, `go tool task check` passed, `cardtown lint fences` passed, and `go tool task sim-compare -- autoloop/m2-wave-g.2` printed `[PASS]`.

### Where it is today (at `autoloop/m2-wave-g.2` 0ca1048)
| Where | Today |
| :--- | :--- |
| `pkg/engine/events.go:137` | `g.emit(Event{Kind: "decision", …` |
| `pkg/engine/events.go:143` | `e := Event{Kind: "action", …` |
| `pkg/engine/events.go:181` | `g.emit(Event{Kind: "match-start", …` |
| `pkg/engine/events.go:188` | `g.emit(Event{Kind: "turn-start", …` |
| `pkg/engine/events.go:198` | `g.emit(Event{Kind: "match-end", …` |
| `pkg/engine/events.go:230` | `Event`'s closing `}`, the file's last line |
| `pkg/engine/choice.go:123` | `g.logf("invalid-answer", …` |
| `pkg/engine/sweep.go:21` | `g.logEvent(Event{Kind: "destroy", …` |
| `pkg/engine/sweep.go:124`, `:126` | `g.logf("result", …` (twice) |
| `pkg/engine/zones.go:29` | `g.logEvent(Event{Kind: "draw", …` |
| `pkg/engine/zones.go:47` | `g.logEvent(Event{Kind: "scratch", …` |
| `pkg/engine/zones.go:78` | `g.logEvent(Event{Kind: "discard", …` |

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)

### 1. `pkg/engine/events.go`
On each line below, replace only the quoted kind with the constant; nothing else on the line changes:
- line 137: `Kind: "decision"` becomes `Kind: EventDecision`;
- line 143: `Kind: "action"` becomes `Kind: EventAction`;
- line 181: `Kind: "match-start"` becomes `Kind: EventMatchStart`;
- line 188: `Kind: "turn-start"` becomes `Kind: EventTurnStart`;
- line 198: `Kind: "match-end"` becomes `Kind: EventMatchEnd`.

After line 230 (`Event`'s closing `}`), append one blank line and then:
```go
// Event kinds, the values of Event.Kind. They are the event schema of ADR-0005:
// written logs and the clients match on them, so a value never changes. Every
// emit and every match names a kind by its constant.
const (
	EventMatchStart    = "match-start"
	EventTurnStart     = "turn-start"
	EventAction        = "action"
	EventDecision      = "decision"
	EventInvalidAnswer = "invalid-answer"
	EventGenerate      = "generate"
	EventPlace         = "place"
	EventDeploy        = "deploy"
	EventMarch         = "march"
	EventScratch       = "scratch"
	EventAttack        = "attack"
	EventFizzle        = "fizzle"
	EventCombat        = "combat"
	EventDamage        = "damage"
	EventReplace       = "replace"
	EventDestroy       = "destroy"
	EventLeave         = "leave"
	EventDraw          = "draw"
	EventDiscard       = "discard"
	EventResult        = "result"
	EventMatchEnd      = "match-end"
)
```
The file ends with the `)` and a newline. The comments `// Kind "match-start"` and the like on lines 227-229 stay as they are. `logf` keeps its signature.

### 2. `pkg/engine/choice.go`
Line 123: `g.logf("invalid-answer", ` becomes `g.logf(EventInvalidAnswer, `.

### 3. `pkg/engine/sweep.go`
- Line 21: `Kind: "destroy"` becomes `Kind: EventDestroy`.
- Lines 124 and 126: `g.logf("result", ` becomes `g.logf(EventResult, `.

### 4. `pkg/engine/zones.go`
- Line 29: `Kind: "draw"` becomes `Kind: EventDraw`.
- Line 47: `Kind: "scratch"` becomes `Kind: EventScratch`.
- Line 78: `Kind: "discard"` becomes `Kind: EventDiscard`.

Nothing else changes in these four files, and no other non-test file changes. The kinds that `actions.go`, `combat.go`, `damage.go`, `effects.go`, `replace.go` and `stack.go` emit (`generate`, `attack`, `fizzle`, `combat`, `damage`, `leave`, `replace`, `place`, `deploy`, `march`) get their constants here but are switched over in T-0111 and T-0112: leave those files alone.

### 5. `pkg/engine/eventkinds_test.go` (new, the `test(T-0110)` commit)
Commit this first, on its own, as `test(T-0110): each event kind constant holds its ADR-0005 value`. It must fail before Edits 1–4 (it does not compile: `EventMatchStart` and the rest are undefined). Record that failure in the Log.
```go
package engine

import "testing"

// Directive #86: each event kind is a named constant whose value is the
// ADR-0005 schema string, unchanged.
func TestEventKinds_ValuesAreTheSchema(t *testing.T) {
	cases := []struct{ got, want string }{
		{EventMatchStart, "match-start"},
		{EventTurnStart, "turn-start"},
		{EventAction, "action"},
		{EventDecision, "decision"},
		{EventInvalidAnswer, "invalid-answer"},
		{EventGenerate, "generate"},
		{EventPlace, "place"},
		{EventDeploy, "deploy"},
		{EventMarch, "march"},
		{EventScratch, "scratch"},
		{EventAttack, "attack"},
		{EventFizzle, "fizzle"},
		{EventCombat, "combat"},
		{EventDamage, "damage"},
		{EventReplace, "replace"},
		{EventDestroy, "destroy"},
		{EventLeave, "leave"},
		{EventDraw, "draw"},
		{EventDiscard, "discard"},
		{EventResult, "result"},
		{EventMatchEnd, "match-end"},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("event kind %q, want %q", c.got, c.want)
		}
	}
}
```

## Acceptance Cases (harness tasks only; written by the lead)
Run each command on its own, pointed at your worktree (W). Record each command and its result in the Log. The reviewer re-runs them.
1. **Existing tests unchanged** (#86 (a)): `git -C W diff autoloop/m2-wave-g.2 --stat -- '*_test.go'` lists only `pkg/engine/eventkinds_test.go`, and `go -C W test ./...` passes.
2. **Sims byte-identical** (#86 (b)): `go -C W tool task sim-compare -- autoloop/m2-wave-g.2` prints `[PASS]`.
3. **Values pinned:** `go -C W test ./pkg/engine -run TestEventKinds_ -v` passes `TestEventKinds_ValuesAreTheSchema`, and the `test(T-0110)` commit alone failed to build (`undefined: EventMatchStart`).
4. **No kind literal left in these four files** (#86 (c), for this task's files): `git -C W grep -n -E -e "Kind: [^E]" -e "logf[(][^E]" -- pkg/engine/events.go pkg/engine/choice.go pkg/engine/sweep.go pkg/engine/zones.go` prints exactly these 4 lines and no others: `pkg/engine/events.go:70` (`logf`'s signature), `pkg/engine/events.go:71` (`Kind: kind` inside `logf`), `pkg/engine/events.go:127` (the `DecisionRecord`'s `Kind: d.Kind.String()`) and `pkg/engine/zones.go:63` (a `Decision`'s `Kind: DecideHand`). None of them is an event kind literal.
5. **Moved, not changed:** `git -C W diff autoloop/m2-wave-g.2 --numstat -- pkg` prints exactly these five lines: `1	1	pkg/engine/choice.go`, `36	0	pkg/engine/eventkinds_test.go`, `32	5	pkg/engine/events.go`, `3	3	pkg/engine/sweep.go` and `3	3	pkg/engine/zones.go`.
6. **Fences:** `go -C W run ./cmd/cardtown lint fences` exits 0.
7. **Only these edits:** `git -C W diff autoloop/m2-wave-g.2 --stat` lists only the five files in case 5 and this task file.
8. `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/engine/events.go`
- `pkg/engine/choice.go`
- `pkg/engine/sweep.go`
- `pkg/engine/zones.go`
- `pkg/engine/eventkinds_test.go` (new)

## Done When
- [ ] Cases 1–8 hold, with their commands and results in the Log
- [ ] `go tool task sim-compare -- autoloop/m2-wave-g.2` passes (starter and stress sims byte-identical to the run branch)
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks

## Log
- 2026-10-30, lead-architect (wave G.2 run, session 4): wrote the task from board directive #86, next in board order now that #85 is applied (T-0109). #86 says to split under the D2 size rule; 17 non-test files carry a kind literal (10 in `pkg/engine`, 4 in `pkg/sim`, 1 in `pkg/view`, 2 in `pkg/eventlog`; none in `cmd/cardtown`), so it is five tasks, T-0110 to T-0114 (Inputs), each at most four non-test files. Only this one is written; the other four are written in later sessions, in that order, and #86 stays open until T-0114 is done. Harness, so it skips `tests` and goes straight to `ready`; tier `flash` as #86 gives. The constants keep the order events happen in a match, not alphabetical. The pin test is the task's own test: sim-compare would not notice a misspelled value of a kind no monitor reads. Every command here is on the headless allow-list, one per line, with no `|`, `$` or unquoted parentheses; case 4 uses `[^E]` and `[(]` so it needs neither. The edits were dry-run on the run branch (Inputs). `main` (281e7d8) has no commits the run branch lacks, so no sync. Waiting directives checked: none has triggered (#6, #7: mode is `serial`; #8: no `deck.go`; #15, #20 (b), #58, #59 (2), #61, #71: no such hook, claim or card here). #87 (c) stays `open: before the next Deliverables task`, after #86. Next: `kernel-engineer` on Flash.
