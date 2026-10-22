---
id: T-0037
title: cardtown play, a text-mode match against a sim policy, drawn only from PlayerView
kind: harness
tier: flash (board directive #27)
wave: none (design note Deliverable 5, part 2 of 2; board directive #24)
status: queued
branch: m2/T-0037-cardtown-play
depends_on: [T-0036, T-0038, T-0039, T-0040]
claims: [cli:play, func:runPlay, func:terminalChooser, task:play]
cards: []
requirements: [FR-MOD-001]
---

## Goal
A human plays a full match in the terminal against a sim policy. `cardtown play` shows the human's view (T-0036), numbers the legal actions and every decision's options, and reads a number. This is design note Deliverable 5: done when a human finishes a match with the cards executable so far. Once it lands, the user can playtest.

## Inputs
- Board directive #24, cases (a) and (b). T-0036 covers (b), and this task covers (a): the play code reads state only through `PlayerView`.
- T-0036: `view.PlayerView`, `view.LogLines`.
- Engine API:
  - `engine.NewGame` and `Config`, with a per-seat `Chooser`.
  - `Game.LegalActions` (EndTurn is always last) and `Game.Apply`.
  - `Decision`, which has `Kind`, `Prompt`, `Options`, `Min`, `Max` and `Repeats`, and the `Answer` type, which has `Picks` and `Timeout`.
- Sim API:
  - `sim.ProgressPolicy` and `sim.RandomPolicy`;
  - `engine.RandomChooser`;
  - `sim.SimDecks`, `sim.FactionDeck` and `sim.DecklistName`.
- CLI: since T-0035, every command has its own file in `cmd/cardtown/`, and `main.go` dispatches. Follow `sim.go`'s flag style.
- Rulings that apply: `FR-MOD-001`.
  - §1: the kernel's fast path never asks.
  - §6: on a timeout or an invalid answer, the kernel's fallback applies. The play loop doesn't need its own.

### Shapes
```go
// cmd/cardtown/play.go
// runPlay plays one match: the human at -seat, the other seat by -policy.
// It reads from in and writes to out, so tests can script a session.
func runPlay(args []string, in io.Reader, out io.Writer) int // exit code
```
`cardtown play [-deck starter] [-p0 FACTION] [-p1 FACTION] [-seat 0|1] [-seed S] [-policy progress|random]`.
- The defaults follow `sim` where flags overlap.
- `-seat` defaults to 0.
- Without `-deck`, the stress decks are used, as in `sim`.

### Lead decisions (technical)
- **The loop:**
  - On the human's turn, print the screen, then the numbered `LegalActions()`, and read one line.
  - On the other seat's turn, apply its policy's action, then print the new log lines.
  - It continues until the match is over, then prints the result.
- **The screen** comes only from `view.PlayerView(g, seat)` and `view.LogLines(g.Log[seen:], seat)`:
  - the new log lines since the last screen;
  - turn and phase;
  - both pools;
  - the opponent's HQ, Reserves and hand and deck counts;
  - the Main Theater;
  - the human's HQ, Reserves, hand with costs, and deck count;
  - the stack;
  - discard counts.
  Exact layout is the engineer's choice, kept short.
- **Actions** are listed one per line, with a label built only from the action's fields and the view, for example `3) Attack: Saboteur#12 -> Eternal Watchman#2`.
- **Decisions:**
  - The human seat's `Chooser` prints `Prompt` and the numbered options, and reads the picks.
  - It reads one number for `Min == Max == 1`, and space-separated numbers otherwise.
  - For `DecideOptional`, it reads `y`/`n`.
  - For `DecideOrder`, it reads the full order.
  - An empty line on an optional or "up to N" decision means none.
  - It returns `Answer{Picks}`. An answer the kernel rejects falls back per `FR-MOD-001` §6, and the loop prints a one-line note.
- **The bot seat** uses `RandomChooser` with a seed derived from `-seed`, as `sim` does.
- **Input:**
  - A non-number or an out-of-range number re-prompts, with no limit.
  - `q` or end of input quits cleanly, printing "quit" and returning exit code 0.
- **Directive #24 (a):** `play.go` must not read game state except through `view.PlayerView` and `view.LogLines`.
  - It may call `LegalActions`, `Apply` and `NewGame`, and use the `Chooser`, `Decision` and `Answer` plumbing.
  - Reading `g.Log[seen:]` only to pass it to `LogLines` is allowed.
  - Reading `g.Outcome.Over` for the loop condition is allowed. The view's `Over` is preferred.
  - No `g.Players`, `g.Theater`, `g.Stack`, `.Inst.Card` or `Snapshot()`.
- **Wiring:**
  - `main.go` dispatches `play` and lists it in the usage text.
  - `Taskfile.yml` gains a `play` task: `cardtown play -deck starter -p0 ANIMUS -p1 PATH {{.CLI_ARGS}}`.
- **Not in this task:**
  - match clocks;
  - a TUI;
  - an undo command;
  - saving a game;
  - a smarter bot;
  - networking.

## Acceptance Cases (harness tasks only; written by the lead)
Cases 1 to 4 go in a new `cmd/cardtown/play_test.go`, committed first as `test(T-0037): …` and failing before the implementation.
1. **A scripted session finishes:**
   - `runPlay` with `-deck starter -p0 ANIMUS -p1 PATH -seat 0 -seed 3`, and an input stream that answers `1` to every prompt and `y` to every yes/no question, finishes the match and returns 0.
   - Its output ends with the result line.
   - It is byte-identical across two runs.
2. **Re-prompting:** an input of `x`, then `999`, then `1` takes the first legal action. The output shows two re-prompts.
3. **Quit:** `q` at the first prompt prints "quit" and returns 0. An empty input stream (EOF) does the same.
4. **No hidden information on screen:**
   - In case 1's whole output, no name of a card in the bot seat's hand or deck appears unless it is publicly visible at that time.
   - No face-down asset's real name appears unless it is publicly visible.
   - The test may check this with the same rule as T-0036's case 4, by replaying the seed.
5. **Grep (directive #24 a):** `grep -nE "\.Players|\.Theater|\.Stack|Inst\.Card|Snapshot\(" cmd/cardtown/play.go` finds nothing. Record the output in the Log.
6. **CLI:** `go run ./cmd/cardtown play -deck starter -p0 ANIMUS -p1 PATH -seed 1 < <file of 1s>` finishes. `cardtown` prints `play` in its usage text. Other commands' outputs are unchanged.
7. `go tool task check` passes.

After merge, the lead asks the user to play one match by hand (Deliverable 5's done-when).

## Allowed Files
- `cmd/cardtown/play.go` (new)
- `cmd/cardtown/main.go` (dispatch and usage only)
- `Taskfile.yml` (the `play` task)
- `cmd/cardtown/play_test.go` (new, the `test(T-0037)` commit)

## Done When
- [ ] Cases 1 to 4 have tests, committed before the implementation, and pass
- [ ] Cases 5 and 6 hold, with output in the Log
- [ ] `go tool task check` passes

## Log
- 2026-10-21 lead-architect: Wrote the task (Deliverable 5 part 2, board directive #24), with status `queued` behind T-0036. Harness task, tier flash (board directive #27, the user's choice).
- 2026-10-21 lead-architect: Held behind the structural cleanup (board directives #30–#37; the user ruled that it comes before feature work). It now depends on T-0038 (directive #30). Each later cleanup task is added here as D2 writes it. After directive #31 moves `view`, this task imports `animus-mundi/pkg/view` instead of `pkg/engine/view`.
