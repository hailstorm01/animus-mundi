---
id: T-0024
title: Move the sim and event-log code out of package engine
kind: harness
tier: sonnet
wave: none (board directive #22, before replay #19)
status: review
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
- 2026-10-20 kernel-engineer: Moved the code and tests as one `refactor(T-0024)` commit (nothing new is tested, so no separate test commit). Status `review`. The "before" binary was built from `rework/executable-rules` at 9892b44 into a temp dir outside the repo.
  - **Code moved (qualifiers only):** `playout.go` → deleted. `CheckInvariants` → `pkg/engine/invariants.go` (unchanged, still reads `g.pending`). `Policy`, `RandomPolicy`, `ProgressPolicy` → `sim/policy.go`. `PlayoutResult`, `Playout` → `sim/playout.go`. `sim.go` → `sim/sim.go` (`FactionDeck`, `SimStats`, `SimOptions`, `Simulate`, `SimulateWith`). `SimDecks` (out of `deck.go`) → `sim/decks.go`. `FormatEvent`, `WriteEventsJSONL` (out of `events.go`) → `eventlog/eventlog.go`. `cmd/cardtown/main.go`: imports and qualifiers only. `doc.go`: one "Package boundary" paragraph.
  - **Unexported engine helpers exported:** none. Everything the moved code touches was already exported (`Game.Asset`, `Asset.Printed`, `Card.Supported`, `LoadDeck`, `ValidateDeck` and so on).
  - **Reuse search:** nothing new written, so nothing to reuse.
  - **Tests moved (old file → new file, name and assertions unchanged):**
    - `playout_test.go` → `sim/playout_test.go`: `TestRandomPlayoutsTerminateAndHoldInvariants`, `TestPlayoutsAreDeterministicPerSeed`, `TestLoadCardsFromRepositoryData`, plus the `keywordDeck` helper.
    - `policy_test.go` → `sim/policy_test.go`: `TestProgress_Case1_LegalOnly`, `TestProgress_Case2_Preferences`, `TestProgress_Case3_Deterministic`, `TestProgress_Case4_Coverage`, `TestProgress_Case5_TerminatesAndHolds`.
    - `sim_events_test.go` → `sim/sim_events_test.go`: `TestSimEvents_Case1_JSONLShape` to `TestSimEvents_Case5_FormatEvent`. It imports `eventlog` for `WriteEventsJSONL` and `FormatEvent`.
    - `deck_load_test.go` → `sim/decks_test.go`: `TestSimDecks_Case13_BuildsValidDecks`, `TestSimDecks_Case14_MissingFileNamesPath`, `TestSimDecks_Case15_InvalidDeckFails`, `TestSimDecks_Case16_FactionDeckStressDecksUnchanged`.
    - Changes inside the moved tests are qualifiers (`engine.`), `eventlog.` for the two event functions, and the data path gaining one more `..` because the package is one directory deeper. No assertion changed.
  - **New `sim/helpers_test.go`:** copies of `hq`, `unit`, `structure`, `deckHQ`, `deckCard` and `writeDeckFile` (engine's are unexported test code, so `sim` can't import them), and a small `scene` with `newScene`/`put`/`give`. It is built from `engine.NewGame` and exported state (append to `Player.Hand`, `Player.Reserves` and `Game.Theater`; IDs start at 100001, above the game's own), and it is what `TestProgress_Case2_Preferences` uses. Case 2's body is unchanged. Reviewer: this file is the one addition beyond the Tests section's literal list.
  - **Stay in engine:** `events_test.go`: `playStarter` now loops on `LegalActions`/`Apply` with a seeded `math/rand/v2` pick (`rand.New(rand.NewPCG(s, ^s))` per seat with the seeds `Playout` used, end bias 0.1), calls `CheckInvariants` after every action and fails on `invalid-answer` events. It returns a local `playResult` (`Outcome`, `Turns`, `Actions`). `death_trigger_test.go`: `playoutWith` calls a new `playDeploy` with the same loop and checks. `TestPlayout_*` and `deployPolicy` are unchanged. `scenario_test.go` untouched; no `TestFR_*` moved.
  - **Case 1:** `go tool task check` exit 0 (gofmt, vet, all tests, `cardtown lint --all`).
  - **Case 2:** `go test -list . ./pkg/engine/...` before and after: 346 `Test*` names each, and `diff` of the two sorted lists is empty.
  - **Case 3:** seed 1 (default), both `-policy random` and `-policy progress`, before vs after binaries, `cmp`: `sim -deck starter -p0 ANIMUS -p1 PATH` (807/805 bytes) identical; `sim` (692/629) identical; `-games 10 -events` stdout (801/802) identical; `-log -seed 7` (15676/16367) identical. The `-events` `.jsonl` files differ only in `"modified"` on the match-start `build` record: the after binary was built from the uncommitted tree (`"modified":true`) and the before binary from a clean tree. Re-run below after the commit.
  - **Case 4:** the grep over `pkg/engine/*.go` has no match (rg exit 1).
  - **Case 5:** `go list -deps ./pkg/engine` includes neither `sim` nor `eventlog` (count 0). `sim`'s deps hold `ability`, `cards` and `sim` under `engine/`, and `eventlog`'s hold `ability`, `cards` and `eventlog`, so neither imports the other.
  - **Case 6:** `go run ./cmd/cardtown lint --all` passes. The fences check walks `pkg/engine` recursively, so it now also scans both new packages.
  - **Case 3 re-run, `-events` files:** the match-start `build` record holds the build's VCS revision and dirty flag, which differ between any two commits by design. With that one field (`"revision":...,"modified":...`) removed from both files, `cmp` of the 10-game `-events` files is identical for `random` (1576329 bytes) and `progress` (2243925 bytes). Every other byte, in all 10 games, matches.
