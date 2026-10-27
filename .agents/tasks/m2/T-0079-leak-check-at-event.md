---
id: T-0079
title: The view leak check accepts a name that was public when its event happened (harness, sonnet)
kind: harness
tier: sonnet
wave: B
status: queued
branch: m2/T-0079-leak-check-at-event
depends_on: [T-0077]
claims: [test:TestView_Case4_NoLeak (leakWatcher)]
cards: []
requirements: []
---

## Goal
`TestView_Case4_NoLeak` stops reporting a log line as a leak when the card it names was public at the moment its event happened, even if the same resolution then hid the card. Real leaks are still caught. This unblocks T-0078 (Arkha, Sun's Fury).

## Inputs
- The block: T-0078's Log on `m2/T-0078-arkha` (6d2cde9..4123001). With Arkha executable, `go test ./pkg/view` fails: `view_test.go:537: seed 13, settled point 61 (turn 11): seat 0's log names "Hammerhand Brigands", hidden by player 1: 122 T11 main decision: player 1, Pick, "Arkha, Sun's Fury: pick up to 1": picked [Hammerhand Brigands] (answer)`.
- Why it is not a leak: the discard pile is public (FRD 01 line 176), and the pick is made from it (`SVY-R14`), so both players saw the card when it was picked. The same step then puts it on the bottom of player 1's deck, so by the next settled point it is hidden and not in `viewtest.PublicNames`.
- The rule Case 4 checks, as amended in T-0036 (lead decision, T-0036 Log, 2026-10-21): "a log line, when it is first shown, never names a hidden card that is not publicly visible at that moment". The watcher reads "that moment" as the first settled point at or after the event. Unceasing Tide made that amendment necessary, because its move to hand emits no event. A pick emits its `decision` event before its step moves the card, so the line is written while the card is public. T-0036 rejected the weaker exception "public earlier in the match", and it stays rejected: what counts is the moment of the line's own event.
- Code: `pkg/view/view_test.go` lines 481-544 (`leakWatcher`, `Event`, `check`), `pkg/view/viewtest/viewtest.go` (`PublicNames`, `HiddenNames`). `view.LogLines` gives one line per event, in order.

## Acceptance Cases (harness tasks only; written by the lead)
1. **The named edit** (an edit to a test from T-0036/T-0046, authorized here; it lands in the `test(T-0079)` commit): in `leakWatcher`, `Event` also records, when `w.g` is set, `viewtest.PublicNames(w.g)` for that event. `check` treats a name in a new log line as public if it is public at this settled point **or** was public when that line's event was recorded. Events recorded while `w.g` is nil (inside `NewGame`) keep only the settled-point check. The view JSON check is unchanged. Update the `leakWatcher` comment to state the rule: a log line never names a hidden card that was not public when its event happened or at the first settled point after it.
2. On `m2/T-0079-leak-check-at-event` as branched: `go test -count=1 -v ./pkg/view -run TestView_Case4_NoLeak` passes, and its log line reports the same "20 matches, N settled points" count as before the edit.
3. With Arkha executable, in a throwaway branch or worktree only (cherry-pick 4123001 from `m2/T-0078-arkha`, never committed to this task's branch): Case 4 fails without the edit (seed 13, settled point 61, "Hammerhand Brigands") and passes with it.
4. Real leaks are still caught, each checked by a temporary mutation that is reverted before any commit:
   - `LogLines` prints the `turn-start` record unredacted (`eventlog.FormatEvent(e)`, snapshot included) → Case 4 fails. This is the path the edit changes, because turn-start events happen after `w.g` is set.
   - `matchStart` prints the opponent's set-aside names → Case 4 fails (the setup path).
   - Record both outcomes in the Log.
5. No non-test file changes. No other test changes.

## Allowed Files
- none (the only change is the named edit to `pkg/view/view_test.go`, Acceptance Case 1)

## Done When
- [ ] `go test -count=1 ./pkg/view` passes
- [ ] Acceptance Cases 2-4 checked and recorded in the Log
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)

## Log
- 2026-10-27 lead-architect (wave run B, session 14): Wrote the task, D6 for T-0078's block. T-0078 goes back to `queued` and depends on this task.
  - Lead decision: the T-0036 rule, read at the line's own event as well as at the next settled point. It is not a rules question. FRD 01 makes the discard pile public, and the picked card was in it when the line was written.
  - Harness, tier `sonnet` rather than `flash`: the edit changes a protected leak oracle, and Cases 3 and 4 need a cherry-pick and source mutations that the headless allow-list does not cover.
  - No existing equivalent: `viewtest` has no per-event view of public names, and `PublicNames` is reused as is. Claims don't overlap any open task.
