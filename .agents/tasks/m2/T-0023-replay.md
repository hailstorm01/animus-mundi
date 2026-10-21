---
id: T-0023
title: Replay a recorded game to any event, as a function and a cardtown replay command
kind: harness
tier: sonnet
wave: none (board directive #19)
status: queued
branch: m2/T-0023-replay
depends_on: [T-0021, T-0024]
claims: [func:Replay, func:ReadEventsJSONL, cli:replay, field:Event.Secondary, field:Event.Resource, field:MatchStart.MulliganInsts, hook:replay-emit]
cards: []
requirements: []
---

## Goal
Given one recorded game (its `match-start` and the events after it), the kernel rebuilds the game deterministically and returns the full state at any event number. `cardtown replay -events <file> -game N -at <seq>` exposes this from a `cardtown sim -events` file. Replay refuses a record from another build, and it reports the first event where it diverges instead of returning a wrong state.

## Inputs
- Board directive #19 (`.agents/tasks/m2/BOARD.md`).
- ADR-0005: the storage model. Events are permanent, full state is stored only at turn starts, and any other point comes from deterministic replay on the same build.
- T-0020 and T-0021 (`pkg/engine/events.go`, `sim.go`): `Event`, `MatchStart` (the seed, deck names, decider, play-first choice, mulligans, revision, modified), `DecisionRecord` (`how`: answer, auto, timeout, invalid), `action` events, `Snapshot`, `WriteEventsJSONL`. Deck names are a decklist path such as `data/decks/animus-starter.txt`, or `stress:<FACTION>`.
- A game is fully determined by its seed, its decks, the play-first choice, the mulligan, every `Chooser` answer and every applied `Action` (ADR-0002, ADR-0003).
- Rulings that apply: none.

### What the record is missing today (add, additively, per ADR-0005)
- `action` events don't record an Attack's Berserk `Secondary`, and they hold the Generate resource only in the Detail text. Add `Secondary int json:"secondary,omitempty"` and `Resource string json:"resource,omitempty"` (`"Requisition"` or `"Tactics"`) to `Event`, set on `action` events.
- `MatchStart.Mulligans` holds names, which can't tell two copies apart. Add `MulliganInsts [2][]int json:"mulliganInsts"`, the set-aside instance IDs, always non-nil.
- Existing fields and their values don't change.

### Shapes
```go
// pkg/engine/events.go
// ReadEventsJSONL reads a -events file back: game number -> that game's events in Seq order.
func ReadEventsJSONL(r io.Reader) (map[int][]Event, error)

// pkg/engine/replay.go (new)
// Replay rebuilds the game recorded in events (match-start first, Seq order)
// and returns the state immediately after event `at` was emitted. That point
// may be mid-resolution. resolveDeck turns a recorded deck name into a DeckList.
func Replay(events []Event, at int, resolveDeck func(name string) (DeckList, error)) (Snapshot, error)
```

### Lead decisions (technical)
- **Rebuild:** `NewGame` takes the recorded seed, the decks from `resolveDeck`, a `PlayFirst` that returns the recorded choice, and a `Mulligan` that returns the recorded `MulliganInsts`. Each seat's `Chooser` replays that seat's recorded `decision` events in Seq order, skipping `auto` ones, which are never asked:
  - `answer`: the recorded picks.
  - `timeout` and `invalid`: `Answer{Timeout: true}`. Both fall back through the same seeded draw, so the picks match.
  Then each recorded `action` event is rebuilt into its `Action` and passed to `Apply`, in order.
- **Divergence check:** compare every replayed event with the recorded one at the same Seq: Kind, Detail, and the structured fields (Player, Card, Inst, Source, Target, Zone, Amount, Secondary, Resource). The first mismatch, a chooser asked something that doesn't match the next recorded decision (kind, player, refs), or a rejected `Apply` returns an error naming the Seq. It never panics.
- **State at `at`:** an unexported hook on `Game` (for example `onEmit func(Event)`), set only by `Replay`, takes `g.snapshot()` right after event `at` is emitted. Then replay stops: return as soon as `at` is reached, without replaying further. The hook isn't on `Observer` and isn't exported. The observer contract (copies only) doesn't change.
- **Build check:** if the record's `Revision` differs from this build's (`runtime/debug.ReadBuildInfo`), refuse with an error that names both. If either side is `modified`, proceed, because the divergence check still guards correctness. The CLI prints one warning line.
- **CLI:** `cardtown replay -events <file> -game N [-at SEQ]`. `-at` defaults to the game's last Seq. It prints the snapshot as indented JSON to stdout. Its `resolveDeck` loads `data/decks/*.txt` paths with `LoadDeck` from the repo root, and `stress:<FACTION>` with `FactionDeck`. A missing file, a missing game, an `-at` out of range, divergence or a build mismatch prints `[ERROR] …` and exits 1. Add it to `cardtown`'s usage list.
- **Not in this task:** stepping or interactive replay, replaying `cardtown play` games (no such files yet), metrics, and redaction.

## Acceptance Cases (harness tasks only; written by the lead)
Cases 1 to 6 go in a new `pkg/engine/replay_test.go`, committed first as `test(T-0023): …` and failing before the implementation. Case 7 is a CLI check, run by the engineer and re-run by the reviewer. Record its output in the Log. Build the sample records in-process: `SimulateWith` with collecting observers, then `WriteEventsJSONL` and `ReadEventsJSONL`, so the round trip goes through JSON.
1. **Turn starts round-trip:** for 20 starter games (progress policy) and 5 stress-deck games, `Replay` to each `turn-start` event's Seq returns a `Snapshot` deep-equal to that event's stored snapshot, for every turn of every game.
2. **Whole game reproduces:** replaying each of those games to its last Seq raises no divergence error, and the result reports the recorded outcome (`Over`, `Draw`, `Winner`, `Reason`).
3. **Mid-game point:** for a recorded `damage` event, `Replay` to its Seq returns a snapshot whose `Turn` equals the event's `Turn`. Two calls with the same `at` are deep-equal.
4. **Divergence is caught:**
   - Change one recorded `answer` decision's pick to another valid index, or one Attack action's `Target` to another asset.
   - `Replay` past that point returns an error naming a Seq at or after the change.
   - `Replay` to a Seq before the change still succeeds.
5. **Build mismatch:** a record whose `MatchStart.Revision` is set to `"not-this-build"` makes `Replay` return an error containing both revisions.
6. **Bad input:** these each return an error, not a panic:
   - `at` 0 or past the last Seq;
   - events without a `match-start` first;
   - a `resolveDeck` error for an unknown deck name.
7. **CLI:**
   - `go run ./cmd/cardtown sim -deck starter -p0 ANIMUS -p1 PATH -games 5 -events <tmp>/e.jsonl`.
   - `go run ./cmd/cardtown replay -events <tmp>/e.jsonl -game 3 -at <seq of game 3's turn 2 turn-start>` exits 0 and prints JSON equal to that line's `snapshot`.
   - `-game 9` and `-at 999999` print `[ERROR]` and exit 1.
   - Running the replay twice gives identical output.

## Allowed Files
- `pkg/engine/replay.go` (new)
- `pkg/engine/events.go` (the additive fields, `ReadEventsJSONL`, the unexported emit hook)
- `pkg/engine/turn.go` (`MulliganInsts`)
- `cmd/cardtown/main.go` (the `replay` command)
- `pkg/engine/replay_test.go` (new, the `test(T-0023)` commit)

## Done When
- [ ] Cases 1 to 6 have tests in `replay_test.go`, committed before the implementation, and all pass
- [ ] Case 7 holds, with its output in the Log
- [ ] `go tool task check` passes, and existing tests pass unchanged
- [ ] The `sim` summaries are unchanged (the `-events` files gain the new fields, which is expected)
- [ ] Never edit a test after its `test(T-0023)` commit. If a case can't pass as written, stop at `blocked` and say why.

## Log
- 2026-10-20 lead-architect: Wrote the task from board directive #19, with status `ready` (harness, sonnet).
- 2026-10-20 lead-architect: Back to `queued`, depending on T-0024. Board directive #22 (added by the user the same day) runs before replay and moves sim and event output out of `pkg/engine`. When T-0024 is `done`, revise Shapes and Allowed Files before dispatch:
  - `ReadEventsJSONL` and the CLI glue go to `pkg/engine/eventlog`.
  - `Replay` stays in `engine` only if it needs the unexported emit hook and `snapshot()`. Otherwise it goes to `eventlog`, behind a minimal exported hook.
