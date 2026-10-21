---
id: T-0024
title: Move the sim and event-log code out of package engine
kind: harness
tier: sonnet
wave: none (board directive #22, before replay #19)
status: ready
branch: m2/T-0024-move-sim-out-of-engine
depends_on: [T-0022]
claims: [package:pkg/engine/sim, package:pkg/engine/eventlog, file:pkg/engine/invariants.go, boundary:engine-sim-eventlog]
cards: []
requirements: []
---

## Goal
`pkg/engine` holds the rules and nothing that drives or prints a match. The policies, playouts and simulation move to `pkg/engine/sim`, and event formatting and the JSON Lines writer move to `pkg/engine/eventlog`. It's a pure move: every output is byte-identical, and every test keeps its name and assertions. Imports run one way: `cmd/cardtown` → `sim` and `eventlog` → `engine`.

## Inputs
- Board directive #22 (`.agents/tasks/m2/BOARD.md`), whose text is the source, including the list of affected test files.
- Today:
  - `playout.go`: `Policy`, `RandomPolicy`, `ProgressPolicy`, `PlayoutResult`, `Playout` and `CheckInvariants`.
  - `sim.go`: `FactionDeck`, `SimStats`, `SimOptions`, `Simulate` and `SimulateWith`.
  - `deck.go`: `SimDecks`.
  - `events.go`: `FormatEvent` and `WriteEventsJSONL`.
  - The callers are `cmd/cardtown/main.go` and the test files the directive lists.
- Rulings that apply: none.

### The boundary (lead decision, as #22 asks)
| Package | Holds |
| :--- | :--- |
| `pkg/engine` | Rules, state, `Apply`/`LegalActions`, `Chooser` and `RandomChooser`, `ValidateDeck` and `LoadDeck`, the `Observer` and every event and record type, and `CheckInvariants`. It moves to a new `invariants.go`, and `playout.go` is deleted. |
| `pkg/engine/sim` | `Policy`, `RandomPolicy`, `ProgressPolicy`, `PlayoutResult`, `Playout`, `FactionDeck`, `SimDecks`, `SimStats`, `SimOptions`, `Simulate` and `SimulateWith`. Same names, same behavior. |
| `pkg/engine/eventlog` | `FormatEvent` and `WriteEventsJSONL`. Same names, same output. `ReadEventsJSONL` and replay's CLI glue land here in T-0023. |
- `sim` imports `engine` only. `eventlog` imports `engine` only. Neither imports the other.
- Move code as-is. Change only package qualifiers (`engine.Game` and so on). If a moved symbol uses an unexported engine helper, export the smallest thing that works, or keep a thin exported wrapper in `engine`, and list it in the Log. Don't change behavior to avoid an export.
- `doc.go` gets one paragraph naming the boundary: rules in `engine`, match driving in `sim`, event output in `eventlog`, and imports never pointing back into these from `engine`. The lead updates the design note at review.

### Tests (decisions the directive left to the lead)
- **Tests move with the code they test.** Each test keeps its name and assertions. The tests go to `pkg/engine/sim/*_test.go` or `pkg/engine/eventlog/*_test.go`, as their subject dictates:
  - `playout_test.go` → `sim`
  - `policy_test.go` → `sim`
  - `sim_events_test.go` → `sim` (it may import `eventlog`)
  - the `TestSimDecks_*` cases of `deck_load_test.go` → `sim`
- **Tests that use unexported scenario helpers:**
  - `TestProgress_Case2_Preferences` moves to `sim`. It builds its three positions with `engine.NewGame` and exported state: append to `Player.Hand`, `Player.Reserves` and `Game.Theater`, and set `Req`/`Tac`. Instance and asset IDs above the game's own range are fine. The assertions don't change.
  - `death_trigger_test.go` (`playoutWith`, `TestPlayout_*`) and `events_test.go` (`playStarter`) stay in `engine`. Their calls to `Playout`/`RandomPolicy` are replaced with a short unexported test helper in the same file. It loops on `LegalActions`/`Apply` with a seeded `math/rand/v2` pick, using the `RandomPolicy` rule (end-turn bias 0.1), and calls `CheckInvariants` after every action. It also fails on an `invalid-answer` event, exactly as `Playout` does.
  - Don't change `scenario_test.go`. No `TestFR_*` test leaves `pkg/engine` (`cardtown lint rules` scans only `pkg/engine/*_test.go`).
- **Names and counts:** `go test -list . ./pkg/engine/...` gives the same set of test names before and after, only spread across packages. Record both lists in the Log, or their diff, which should be empty once sorted.

## Acceptance Cases (harness tasks only; written by the lead)
This is a pure move, so the checks are commands. The engineer runs them and the reviewer re-runs them. Record the output in the Log. Build a "before" binary from `rework/executable-rules` into a temp dir first.
1. `go tool task check` passes.
2. The sorted test-name list across `./pkg/engine/...` is identical before and after.
3. For seed 1 and both `-policy random` and `-policy progress`, these are byte-identical between the before and after binaries (`cmp`):
   - `cardtown sim -deck starter -p0 ANIMUS -p1 PATH`
   - `cardtown sim`
   - `cardtown sim -deck starter -p0 ANIMUS -p1 PATH -games 10 -events <tmp>/e.jsonl`, comparing the `.jsonl` files
   - `cardtown sim -deck starter -p0 ANIMUS -p1 PATH -log -seed 7`
4. `grep` for `func (RandomPolicy|ProgressPolicy|Playout|Simulate|SimulateWith|FactionDeck|SimDecks|FormatEvent|WriteEventsJSONL)\b` and `type (Policy|PlayoutResult|SimStats|SimOptions)\b` across `pkg/engine/*.go` gives no match.
5. `go list -deps ./pkg/engine` includes neither `pkg/engine/sim` nor `pkg/engine/eventlog`. `go list -deps ./pkg/engine/sim` doesn't include `eventlog`, and the reverse holds too.
6. `go run ./cmd/cardtown lint --all` passes, including fences, which now also scans the two new packages.

## Allowed Files
- `pkg/engine/playout.go` (deleted), `pkg/engine/invariants.go` (new), `pkg/engine/sim.go` (deleted), `pkg/engine/deck.go` (`SimDecks` out), `pkg/engine/events.go` (`FormatEvent` and `WriteEventsJSONL` out), `pkg/engine/doc.go`, and any engine file needing a minimal export (list it)
- `pkg/engine/sim/*.go` and `pkg/engine/eventlog/*.go` (new)
- `cmd/cardtown/main.go` (imports and qualifiers only)
- The test files named under Tests, moved or edited only as described there. No other test changes.
- Size exception: a pure move across many files, as directive #22 specifies.

## Done When
- [ ] Acceptance cases 1 to 6 hold, with their output in the Log
- [ ] Each moved test is listed in the Log (old file → new file), with its assertions unchanged
- [ ] Never change a test's assertions. If a test can't keep them after the move, stop at `blocked` and say which and why.

## Log
- 2026-10-20 lead-architect: Wrote the task from board directive #22, with status `ready` (harness, sonnet). The directive says it runs before replay, so T-0023 is back at `queued` and depends on this task. Its Shapes get revised for the new packages when this is `done`.
