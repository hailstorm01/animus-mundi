---
id: T-0021
title: cardtown sim -log and -events (JSON Lines), with sorted end reasons
kind: harness
tier: sonnet
wave: none (board directive #18, part 2 of 2: CLI; also directive #12)
status: ready
branch: m2/T-0021-sim-log-and-events
depends_on: [T-0020]
claims: [cli:sim-log, cli:sim-events, func:SimulateWith, func:FormatEvent, func:WriteEventsJSONL, format:events-jsonl, task:sim-log, task:sim-events]
cards: []
requirements: []
---

## Goal
T-0020's events reach the user. `cardtown sim -log` prints one game's events as readable text, and `cardtown sim -events <file>` writes every game's events as JSON Lines. For the same seed, both outputs are byte-identical, and so is the normal `sim` summary, because end reasons now print in sorted order (directive #12).

## Inputs
- Board directive #18 (d) and (e), and directive #12 (`.agents/tasks/m2/BOARD.md`).
- ADR-0005 (`docs/architecture/decisions/0005-structured-game-events-and-observer.md`): the event schema, the observer contract and the storage model.
- T-0020: `Event` (with json tags), `Observer`, `Config.Observer`, `Config.DeckNames` and the `match-start`, `turn-start` and `match-end` records (`pkg/engine/events.go`).
- Today: `runSim` in `cmd/cardtown/main.go` prints the summary and ranges over the `SimStats.Reasons` map, so the line order varies. `Simulate` is in `pkg/engine/sim.go`, and the `sim` task is in `Taskfile.yml`.
- Rulings that apply: none.

### Shapes
```go
// pkg/engine/sim.go
type SimOptions struct {
	DeckNames [2]string                       // recorded in each match-start
	Observer  func(game int, seed uint64) Observer // optional; one observer per game, game counts from 1
}
func SimulateWith(decks [2]DeckList, games int, seed uint64, maxTurns int, opt SimOptions) (SimStats, error)
// Simulate keeps its signature and calls SimulateWith with zero options.

// pkg/engine/events.go
func FormatEvent(e Event) string                       // one line, no newline, starting with the Seq
func WriteEventsJSONL(w io.Writer, game int, events []Event) error
// One JSON object per event and per line, in order. Each object is the event's
// own JSON with a leading "game" key: {"game":1,"seq":1,...}.
```

### Lead decisions (technical)
- **`-log`:** plays exactly one game, the one at `-seed`. It forces `-games 1`, and an explicit `-games` other than 1 together with `-log` is an error. It prints that game's events with `FormatEvent`, one per line in Seq order, then the normal summary for that one game. `FormatEvent` shows Seq, Turn, Phase, Kind and Detail. Records get a readable line:
  - `match-start`: seed, both decks, both HQs, the decider and their choice;
  - `turn-start`: turn and active player;
  - `decision`: player, kind, prompt, the picked option labels and `how`;
  - `match-end`: outcome, reason, turns.
  The full snapshot is never printed.
- **`-events <file>`:** creates or truncates the file and writes every game's events, games in seed order, with `WriteEventsJSONL`. `game` counts from 1, and each game's `match-start` carries its seed. Settled notifications aren't written. It works with or without `-log`. A file that can't be written prints `[ERROR]` and exits 1, and the file is closed before the summary prints.
- **Deck names:** the decklist file path when `-deck` is set (for example `data/decks/animus-starter.txt`), else `stress:<FACTION>`.
- **Determinism:** the observer only collects. The summary, `-log` and `-events` outputs are byte-identical across runs with the same flags. End reasons print sorted by reason text (#12).
- **Taskfile:** add `sim-log` (`sim -deck starter -p0 ANIMUS -p1 PATH -log {{.CLI_ARGS}}`) and `sim-events` (the same, with `-events {{.CLI_ARGS}}`, where the file path is passed after `--`). Follow the existing `sim` task's form.
- **Not in this task:** the replay command (#19), metrics, reports, redaction for players, and an `ability` event (ADR-0005, Consequences).

## Acceptance Cases (harness tasks only; written by the lead)
Tests 1 to 5 go in a new `pkg/engine/sim_events_test.go`, committed first as `test(T-0021): …` and failing before the implementation. Cases 6 to 8 are CLI checks, run by the engineer and re-run by the reviewer. Record their output in the Log.
1. **JSON Lines shape:** `SimulateWith` on the starter decks for 3 games, with observers collecting events, then `WriteEventsJSONL` for each game into one buffer. Every line parses as a JSON object with integer `game` and `seq`. Games appear as 1, 2, 3 in order. Within a game, `seq` runs 1..n without gaps. The first line of each game is `match-start`, carrying that game's seed and the deck names, and the last is `match-end`.
2. **One turn-start per turn:** in that output, each game has exactly one `turn-start` line per turn number from 1 to the game's `match-end` turn count.
3. **Byte-identical:** the case 1 run, done twice with the same seed, produces identical bytes. A different seed produces different bytes.
4. **Observing changes nothing:** `SimulateWith` with collecting observers and `Simulate` with no options return equal `SimStats` for the same decks, games and seed.
5. **FormatEvent:** for every event of one starter game, the line is one line, starts with its Seq, and contains its Kind. A `decision` line contains the picked option's label and its `how`. A `match-start` line contains both HQ names and the seed.
6. **CLI `-log`:** `go run ./cmd/cardtown sim -deck starter -p0 ANIMUS -p1 PATH -log -seed 7` exits 0. It prints the game's event lines and then a summary of 1 game. Running it twice gives identical output, and `-log -games 5` is rejected.
7. **CLI `-events`:** `go run ./cmd/cardtown sim -deck starter -p0 ANIMUS -p1 PATH -games 20 -events <tmp>/e.jsonl` exits 0. Its summary equals the same run without `-events`. The file has 20 `match-start` lines, and running it twice gives byte-identical files. An unwritable path prints `[ERROR]` and exits 1.
8. **#12:** two runs of `go run ./cmd/cardtown sim` and of `go run ./cmd/cardtown sim -deck starter -p0 ANIMUS -p1 PATH` each give byte-identical stdout. The summary numbers are unchanged from `rework/executable-rules` (compare with sorted lines).

## Allowed Files
- `cmd/cardtown/main.go`
- `pkg/engine/sim.go`
- `pkg/engine/events.go` (`FormatEvent`, `WriteEventsJSONL`)
- `Taskfile.yml`
- `pkg/engine/sim_events_test.go` (new, the `test(T-0021)` commit)

## Done When
- [ ] Cases 1 to 5 have tests in `sim_events_test.go`, committed before the implementation, and all pass
- [ ] Cases 6 to 8 hold, with their output in the Log
- [ ] `go tool task check` passes
- [ ] `go tool task sim-log` and `go tool task sim-events -- <file>` work

## Log
- 2026-10-20 lead-architect: Wrote the task (directive #18 part 2, and #12), with status `ready` (harness, sonnet).
