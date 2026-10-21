---
id: T-0033
title: One deck-naming scheme and one build-revision reader
kind: harness
tier: sonnet
wave: none (board directive #26, cleanup 2 of 4, before cardtown play)
status: review
branch: m2/T-0033-deck-names-and-build-info
depends_on: [T-0032]
claims: [func:DecklistName, func:StressName, func:ResolveDeck, func:BuildInfo]
cards: []
requirements: []
---

## Goal
Two pieces of logic that replay depends on exist once instead of several times. Today the deck-naming scheme is written in four places:
- `sim/decks.go` builds decklist paths;
- `cmd/cardtown/main.go` builds the recorded names twice;
- `main.go`'s replay resolver turns a recorded name back into a deck.

The build revision is read twice: once in `events.go`, which stamps match-start, and once in `eventlog/replay.go`, which compares against it. If one copy of either drifts, recorded games stop replaying.

## Inputs
- Board directive #26 (the user's audit), duplicates 1 and 2.
- Today's code:
  - `sim.SimDecks` (`pkg/engine/sim/decks.go`);
  - `runSim`'s `opt.DeckNames` and the executable-count lines, and `runReplay`'s `resolve` (`cmd/cardtown/main.go`);
  - `emitMatchStart` (`pkg/engine/events.go`);
  - `buildInfo` (`pkg/engine/eventlog/replay.go`).
- T-0021 fixed the naming scheme: a decklist name is the path `data/decks/<faction lowercased>-<deck>.txt` with forward slashes, and a stress deck is `stress:<FACTION>`.
- Rulings that apply: none.

### Shapes
```go
// pkg/engine/sim/decks.go
func DecklistName(faction, deck string) string // "data/decks/<lower(faction)>-<deck>.txt", slash-separated, relative to the repo root
func StressName(faction string) string         // "stress:" + faction
// ResolveDeck turns a recorded deck name back into a deck: a stress name
// builds FactionDeck, any other name is a decklist path under repoRoot,
// loaded with LoadDeck (not validated, as replay does today).
func ResolveDeck(repoRoot string, cards map[string]*engine.Card, name string) (engine.DeckList, error)

// pkg/engine (events.go)
// BuildInfo is this binary's VCS revision and dirty flag from
// runtime/debug.ReadBuildInfo, "" and false when absent.
func BuildInfo() (revision string, modified bool)
```

### Lead decisions (technical)
- **`SimDecks`** builds its paths with `DecklistName`. Its error messages stay byte-identical.
- **`runSim`:** the recorded deck names and the printed executable-count lines use `DecklistName`/`StressName`.
- **`runReplay`:** its resolver is `sim.ResolveDeck`.
- **`emitMatchStart` and `eventlog`:** both use `engine.BuildInfo`, and `eventlog.buildInfo` is removed.
- **No output change.** Every CLI output, error message, `-events` file and replay result stays byte-identical.
- **Not in this task:**
  - the rule-code duplicates (T-0034);
  - splitting `main.go` (T-0035);
  - validating decks on replay.

## Acceptance Cases (harness tasks only; written by the lead)
Cases 1 and 2 go in new test files, committed first as `test(T-0033): …` and failing before the implementation:
- `pkg/engine/sim/decknames_test.go`
- `pkg/engine/eventlog/buildinfo_test.go`, or an engine test file of the engineer's choice

Cases 3 to 5 are CLI checks, recorded in the Log.
1. **Round trip:**
   - For each of `ANIMUS` and `PATH`, `ResolveDeck(repoRoot, cards, DecklistName(f, "starter"))` returns the same `DeckList` as `SimDecks` for that faction, compared by card names in order.
   - `ResolveDeck(repoRoot, cards, StressName(f))` equals `FactionDeck(cards, f)` for `SOULLESS` and `CONCORDIA`.
   - `DecklistName("ANIMUS", "starter")` is exactly `data/decks/animus-starter.txt`.
   - An unknown decklist name returns an error naming the path.
2. **One revision source:** a new game's match-start `Revision` and `Modified` equal `engine.BuildInfo()`.
3. **Replay still works:**
   - Write `cardtown sim -deck starter -p0 ANIMUS -p1 PATH -games 5 -events <tmp>/e.jsonl`.
   - Then `cardtown replay -events <tmp>/e.jsonl -game 3 -at <seq of game 3's turn-2 turn-start>` exits 0, and its output equals that turn-start's stored snapshot.
   - A stress-deck file (bare `sim -games 3 -events`) replays the same way.
4. **Outputs unchanged:** `sim` (starter and stress, both policies) and `-log -seed 7` are byte-identical to a baseline CLI built from `rework/executable-rules` before any edit. `-events` files are identical once the `revision`/`modified` stamp is normalized.
5. `go tool task check` passes.

## Allowed Files
- `pkg/engine/sim/decks.go`
- `cmd/cardtown/main.go`
- `pkg/engine/events.go`
- `pkg/engine/eventlog/replay.go`
- the new test files (the `test(T-0033)` commit)

## Done When
- [ ] Cases 1 and 2 have tests, committed before the implementation, and pass
- [ ] Cases 3 and 4 hold, with output in the Log
- [ ] `go tool task check` passes

## Log
- 2026-10-21 lead-architect: Wrote the task from board directive #26, with status `queued` behind T-0032 (harness, sonnet).
- 2026-10-21 lead-architect: T-0032 is merged (e9ddbc5). Set to `ready`. File paths in Inputs are unchanged by T-0032, apart from `emitMatchStart`, which is still in `events.go`.
- 2026-10-21 kernel-engineer: Implemented. Tests first in 04f97c5 (`sim/decknames_test.go` Case 1, `eventlog/buildinfo_test.go` Case 2); they failed to compile before the change (undefined `DecklistName`, `engine.BuildInfo`) and pass now. Not edited since.
  - Reuse: searched; no existing naming helper or BuildInfo reader beyond the two copies being merged. The old copies are removed (`eventlog.buildInfo`, `runtime/debug` import in replay.go; the inline reader in `emitMatchStart`; `runSim`'s two path builders and `file0`/`file1`; `runReplay`'s inline resolver).
  - Diff: `sim/decks.go` adds `DecklistName`, `StressName`, `ResolveDeck`; `SimDecks` takes `filename` as `DecklistName` minus the `data/decks/` prefix, so its errors are byte-identical (the join with `decksDir` is unchanged). `events.go` adds `BuildInfo`; `emitMatchStart` calls it. `replay.go` calls `engine.BuildInfo`. `main.go`: `runSim` DeckNames and the two executable-count lines use `sim.DecklistName`/`StressName`; `resolve` is `sim.ResolveDeck(repoRoot, cards, name)`.
  - Case 3: `sim -deck starter -p0 ANIMUS -p1 PATH -games 5 -events` and bare `sim -games 3 -events` (new binary). `replay -game 3 -at 12` (seq of game 3's turn-2 turn-start) exits 0 for both; output is JSON-equal to the stored snapshot (checked with a throwaway Go program). Only the expected `[WARN]` on stderr (dirty tree).
  - Case 4: baseline CLI built from `rework/executable-rules` in a separate worktree (`../baseline-T0033`, now removed). `sim` starter and stress, `-policy random` and `progress`, and `-log -seed 7` (starter and stress): stdout byte-identical, `-events` files byte-identical after normalizing `revision`/`modified`. Error outputs for a missing decklist differ only in the repo-root prefix (the two checkouts live in different directories).
  - Gates: `go test ./...` all ok; `go tool task check` rc=0.
  - Reviewer: `SimDecks` derives `filename` with `strings.TrimPrefix(DecklistName(...), "data/decks/")`, which couples it to that prefix. `ResolveDeck` is not validated, as asked.
