---
id: T-0202
title: "*Chainbreaker's Tenacity*'s armor grant is its own sentence (directive #140 (1), harness, flash, text only)"
kind: harness
tier: flash
wave: F
status: ready
branch: m2/autoloop-wave-f/T-0202-tenacity-armor-sentence
depends_on: [T-0201]
claims: [text:Chainbreaker's Tenacity]
cards: []
requirements: []
---

## Goal
*Chainbreaker's Tenacity* reads `Chosen allied unit gets +1 power until end of turn. It gains 2 armor.` in its CSV row, in its definition's `Text`, and in the comment that quotes it above its tests (#140 (1), `SVY-R40`). Text only: same behavior, same test names, same sim output.

## Inputs
- Board directive #140 (1) (user via game-designer, 2026-11-09, [outcome](../../../docs/design/outcomes/2026-11-09-wave-f-keyword-grant-duration-and-inspection-timeout.md)): the three edits, in one commit. pm-frd-drafter left the CSV row alone for this task.
- FRD 01 `FR-LAY-003`, "Duration Reach & Wording" (`SVY-R39`, `SVY-R40`), already quotes the new text.
- Why nothing else changes: `TestDefs_MatchRulesText` (`pkg/engine/defs_test.go:145`) compares each definition's `Text` with its CSV `rulesText`, so the CSV and `cards/path.go` change together. *Tenacity*'s tests (`pkg/engine/cards_path_maneuvers_test.go:590-628`) pin its behavior, not its text. No other copy of the old text exists outside `.agents/` and `docs/design/outcomes/` (`git grep`, at 721f371).
- The lead dry-ran these exact edits in a throwaway worktree of the run branch at 721f371: `go tool task check` exit 0, `cardtown lint fences` exit 0, and `go tool task sim-compare -- autoloop/m2-wave-f` printed `[PASS]`. Every expected output below is from that dry run.

## Edits (lead decisions; copy exactly)
Line numbers are at 721f371. Keep each file's line endings as they are (LF).

### 1. `data/cards/maneuvers-path.csv`, line 2
In the `rulesText` field, change `Chosen allied unit gets +1 power until end of turn and gains 2 armor.` to `Chosen allied unit gets +1 power until end of turn. It gains 2 armor.` Nothing else on the line changes.

### 2. `pkg/engine/cards/path.go`, line 229
Change the same string the same way, so the line reads (two tabs first, as today):
```go
		Text: []string{"Chosen allied unit gets +1 power until end of turn. It gains 2 armor."},
```

### 3. `pkg/engine/cards_path_maneuvers_test.go`, lines 523-524 (comment only; authorized by #140 (1))
Replace these two comment lines:
```go
// Chainbreaker's Tenacity: "Chosen allied unit gets +1 power until end of turn
// and gains 2 armor." Cost 1 Tactics.
```
with these two:
```go
// Chainbreaker's Tenacity: "Chosen allied unit gets +1 power until end of turn.
// It gains 2 armor." Cost 1 Tactics.
```
No test function in that file changes.

All three edits go in one commit, `chore(cards): Chainbreaker's Tenacity's armor grant is its own sentence (T-0202)`. There are no new tests, so no `test(T-0202)` commit.

## Acceptance Cases (harness tasks only; written by the lead)
No new tests: a text change only. W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **The old text is gone:** `git -C W grep -n -F -e 'until end of turn and gains 2 armor' -- data pkg` prints nothing (exit 1 is expected: no match).
2. **The new text is in both places:** `git -C W grep -n -F -e 'until end of turn. It gains 2 armor.' -- data pkg` prints exactly two lines: `data/cards/maneuvers-path.csv:2:` and `pkg/engine/cards/path.go:229:`.
3. **The test comment quotes it:** `git -C W grep -n -F -e 'It gains 2 armor' -- pkg/engine/cards_path_maneuvers_test.go` prints exactly one line, line 524.
4. **Only the listed edits:** `git -C W diff autoloop/m2-wave-f --numstat -- data pkg cmd` prints exactly three lines: `1	1	data/cards/maneuvers-path.csv`, `1	1	pkg/engine/cards/path.go` and `2	2	pkg/engine/cards_path_maneuvers_test.go`.
5. **Defs match the data:** `go -C W test ./pkg/engine -run TestDefs_MatchRulesText` passes.
6. **Sims byte-identical:** `go -C W tool task sim-compare -- autoloop/m2-wave-f` prints `[PASS]`.
7. **Fences:** `go -C W run ./cmd/cardtown lint fences` exits 0.
8. **Nothing else changed:** `git -C W diff autoloop/m2-wave-f --stat` lists only the three files in case 4 and this task file.
9. `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `data/cards/maneuvers-path.csv`
- `pkg/engine/cards/path.go`
- `pkg/engine/cards_path_maneuvers_test.go` (the comment in Edit 3 only)

## Done When
- [ ] Cases 1–9 hold, with their commands and results in the Log
- [ ] `go tool task sim-compare -- autoloop/m2-wave-f` passes (starter and stress sims byte-identical to the run branch)
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-09, lead-architect (wave F run, session 1): wrote this task from board directive #140 (1), which goes ahead of wave F's first card task (D2.0), against the run branch at 721f371. Run file created; no `main` commits to sync. Not `WAVE-COMPLETE`: all eight of the header's cards are on `NotYetExecutable`. Harness, so it skips `tests` and goes straight to `ready`; tier `flash`, because the edits are exact and were dry-run. Two non-test files, under the size rule. Every command here is on the headless allow-list, one per line, with no pipes, `$`, double quotes or unquoted parentheses. Waiting directives checked: none has triggered (#6, #7: mode is `serial`; #15: no turn-boundary hook; #71, #91 (2), #106, #120, #128, #136, #137: no such card in a task). #140 (2)–(4) wait for their tasks. Deck cards still on `NotYetExecutable` (#14 (e)): 8. Next: `kernel-engineer` on Flash.
