---
id: T-0047
title: One match-setup scheme, sim.MatchDecks, sim.SeatSeed and sim.DefaultEndBias (directive #41, harness, sonnet)
kind: harness
tier: sonnet, high effort (user, 2026-10-22)
wave: none (board directive #41, structural review 2 cleanup 3 of 5, before the next shape task)
status: done
branch: m2/T-0047-match-setup
depends_on: [T-0046]
claims: [func:sim.MatchDecks, func:sim.SeatSeed, const:sim.DefaultEndBias]
cards: []
requirements: []
---

## Goal
`cardtown sim`, `cardtown play` and `sim.SimulateWith` set a match up the same way, because they call the same three things in `pkg/sim` instead of keeping copies. A `play` game is then reproducible in `sim` by construction. No output changes.

## Inputs
- Board directive #41 (structural review 2, finding 4).
- The copies, at the integration branch when the task was written:
  - **deck choice and naming:** [cmd sim.go:54-92](../../../cmd/cardtown/sim.go) and [play.go:56-85](../../../cmd/cardtown/play.go). With `-deck`, both use `SimDecks` and name the decks with `DecklistName`; without it, `FactionDeck` per seat and `StressName`.
  - **per-seat seeds** `seed*2+1` / `seed*2+2`: [pkg/sim/sim.go:104,129](../../../pkg/sim/sim.go) (the choosers and the policies) and [play.go:91-93](../../../cmd/cardtown/play.go) (`botSeed`).
  - **random-policy end bias** `0.1`: `pkg/sim/sim.go:127` (and the `SimOptions.Policy` comment at :70) and `play.go:130`.
- Existing helpers in [pkg/sim/decks.go](../../../pkg/sim/decks.go): `DecklistName`, `StressName`, `ResolveDeck` and `SimDecks`.
- Rulings that apply: none.

### Lead decisions
1. **In `pkg/sim/decks.go`:**
   ```go
   // MatchDecks returns the two seats' decks and their recorded names: with deck,
   // the validated decklists data/decks/<faction>-<deck>.txt named by DecklistName;
   // without it, each faction's stress deck named by StressName.
   func MatchDecks(repoRoot string, cards map[string]*engine.Card, deck string, factions [2]string) ([2]engine.DeckList, [2]string, error)
   ```
   - With `deck != ""`, it calls `SimDecks(cards, filepath.Join(repoRoot, "data", "decks"), deck, factions)`.
   - Otherwise it calls `FactionDeck(cards, f)` for each seat.
   - Errors are returned unwrapped, so the commands' `[ERROR] %v` lines stay byte-identical.
   - On error, the names are the zero value.
2. **In `pkg/sim/sim.go`:**
   ```go
   // SeatSeed is the seed of seat's chooser and policy in the game with seed:
   // seed*2+1 for seat 0 and seed*2+2 for seat 1.
   func SeatSeed(seed uint64, seat engine.PlayerID) uint64

   // DefaultEndBias is RandomPolicy's chance of ending the turn when other actions remain.
   const DefaultEndBias = 0.1
   ```
   - `SimulateWith` uses `SeatSeed(s, 0)` and `SeatSeed(s, 1)` for both the choosers and the policies.
   - The default policy is `RandomPolicy(seed, DefaultEndBias)`.
   - The `SimOptions.Policy` comment names `DefaultEndBias` instead of `0.1`.
3. **In `cmd/cardtown/sim.go` and `play.go`:**
   - Each still loads the cards with `engine.LoadCards`.
   - The deck-choice and naming blocks become one `sim.MatchDecks` call. `sim.go` puts the names in `opt.DeckNames`, and `play.go` in `deckNames`.
   - `play.go`'s `botSeed` becomes `sim.SeatSeed(*seed, botSeat)`, and its random bot is `sim.RandomPolicy(botSeed, sim.DefaultEndBias)`.
   - `sim.go`'s summary lines that print `sim.DecklistName(*p0, *deck)` (around :170) stay as they are.
4. **Not in this task:** the tests' own literal copies of the seed scheme. Existing tests don't change, and `play.go` doesn't change beyond decision 3. #42 splits it next.

## Acceptance Cases (harness tasks only; written by the lead)
Case 1 is new tests in a new `pkg/sim/match_test.go`, committed as `test(T-0047): …` before the implementation. Their failing to compile counts. Cases 2–4 are checks recorded in the Log against a baseline CLI built from `rework/executable-rules` before any edit; the reviewer re-runs them.
1. **Unit tests (repo-root data, as the other `sim` tests reach it):**
   - (a) `MatchDecks(root, cards, "starter", [2]string{"ANIMUS", "PATH"})` returns the same decks as `SimDecks` (compare the HQ names and main-deck card names in order), and the names `data/decks/animus-starter.txt` and `data/decks/path-starter.txt`.
   - (b) `MatchDecks(root, cards, "", [2]string{"ANIMUS", "PATH"})` returns the same decks as `FactionDeck` per seat, and the names `stress:ANIMUS` and `stress:PATH`.
   - (c) `MatchDecks(root, cards, "nonexistent", …)` returns an error.
   - (d) `SeatSeed(5, 0) == 11`, `SeatSeed(5, 1) == 12`, and `DefaultEndBias == 0.1`.
2. **Outputs unchanged:** against the baseline binary, these are byte-identical (`cmp`), including exit codes:
   - `cardtown sim -deck starter -p0 ANIMUS -p1 PATH` under `-policy random` and `-policy progress`;
   - bare `cardtown sim`;
   - `cardtown sim -deck starter -p0 ANIMUS -p1 PATH -games 1 -seed 7 -log`;
   - `cardtown sim -deck starter -p0 ANIMUS -p1 PATH -games 10 -events <file>`, comparing the `.jsonl` files with `"revision":…,"modified":…` removed;
   - `cardtown play -deck starter -p0 ANIMUS -p1 PATH -seat 0 -seed 3`, fed a file of `1` lines;
   - `cardtown play -deck starter -p0 ANIMUS -p1 PATH -seat 1 -seed 3 -policy random`, fed the same file (this covers the other bot seat and the random policy);
   - `cardtown sim -deck nonexistent -p0 ANIMUS -p1 PATH`, and the same for `play` (the error lines).
3. **No copies left:** outside tests, `git grep -n -E '\*2 ?\+ ?[12]\b' -- pkg/sim cmd` matches only inside `SeatSeed`. `git grep -n -E '[, (]0\.1\b' -- pkg/sim cmd ':!*_test.go'` matches only the `DefaultEndBias` declaration. `git grep -n -E 'sim\.(SimDecks|FactionDeck|StressName)\(' -- cmd` finds nothing.
4. `go tool task check` passes, and no existing test changed.

## Allowed Files
- `pkg/sim/decks.go` (`MatchDecks`)
- `pkg/sim/sim.go` (`SeatSeed`, `DefaultEndBias`, their use in `SimulateWith` and the `SimOptions` comment)
- `cmd/cardtown/sim.go` and `cmd/cardtown/play.go` (decision 3 only)
- Tests: new `pkg/sim/match_test.go`

## Done When
- [ ] Case 1 is committed before the implementation and failed before it
- [ ] Cases 2–4 hold, with their commands and results in the Log
- [ ] `go tool task check` passes

## Log
- 2026-10-22 lead-architect: Wrote the task from board directive #41. Harness, tier sonnet. **The user asked for Sonnet at high effort** for this task, so the lead dispatches `kernel-engineer` with effort high. T-0046 is done, so the task is `ready` on branch `m2/T-0047-match-setup`.
- 2026-10-22 kernel-engineer: Implemented in scratch worktree on `m2/T-0047-match-setup`. Commits: `0386429` status in-progress, `c2b26b5` test(T-0047) (case 1, failed to compile before the implementation: `undefined: MatchDecks`, `undefined: SeatSeed`), then the feat commit. Status is now `review`.
  - **Reuse search:** `SimDecks`, `FactionDeck`, `DecklistName`, `StressName` and `RandomPolicy` already existed in `pkg/sim`; `MatchDecks` only composes them. The test file reuses `deckNames` from `pkg/sim/decknames_test.go`. No engine shape added.
  - **Diff:** `pkg/sim/decks.go` adds `MatchDecks` (errors unwrapped; names zero on error). `pkg/sim/sim.go` adds `SeatSeed` and `DefaultEndBias`, and `SimulateWith` uses them for both choosers and policies; the `SimOptions.Policy` comment names `DefaultEndBias`. `cmd/cardtown/sim.go` and `play.go` replace the deck-choice and naming blocks with one `sim.MatchDecks` call (`opt.DeckNames` / `deckNames`), and `play.go` uses `sim.SeatSeed(*seed, botSeat)` and `sim.RandomPolicy(botSeed, sim.DefaultEndBias)`. The `sim.DecklistName(*p0, *deck)` summary lines are untouched. No existing test changed.
  - **Case 1:** `go test ./pkg/sim -run MatchSetup -v`: 1a, 1b, 1c, 1d pass.
  - **Case 2:** baseline binary built from `ac4ef56` (`go build -o ct47-base.exe ./cmd/cardtown`) before any edit; new binary built from the feat tree. Both run from the same worktree directory; `play` fed 5000 `1` lines. stdout+stderr and exit code compared with `cmp`; all identical: `sim -deck starter -p0 ANIMUS -p1 PATH -policy random` (rc 0, `[PASS]` line), `-policy progress` (rc 0, `[PASS]`), bare `sim` (rc 0, `[PASS]`), `-games 1 -seed 7 -log` (rc 0, 382 lines, `[PASS]`), `play ... -seat 0 -seed 3` (rc 0, ends `Result: Player 1 won (Headquarters destroyed)`, 10526 lines), `play ... -seat 1 -seed 3 -policy random` (rc 0, ends `Result: Player 0 won (Headquarters destroyed)`, 2035 lines), `sim -deck nonexistent ...` and `play -deck nonexistent ...` (rc 1, the one `[ERROR] reading decklist ...animus-nonexistent.txt` line). `sim -games 10 -events <file>`: stdout and rc identical, and the `.jsonl` files (4213 lines) are identical after removing `"revision":…,"modified":…`; raw files differ only in those two fields (the commit hash and `modified`).
  - **Case 3:** `git grep -n -E '\*2 ?\+ ?[12]\b' -- pkg/sim cmd` matches `SeatSeed` (sim.go:75-76) and the tests' own literal copies (`cmd/cardtown/play_test.go:145`, `pkg/sim/policy_test.go:18,29`), which decision 4 leaves alone. The `0.1` grep (non-test) matches only `sim.go:79`, the `DefaultEndBias` declaration. `git grep -E 'sim\.(SimDecks|FactionDeck|StressName)\(' -- cmd` finds only `cmd/cardtown/play_test.go:138` (a test); no non-test hits.
  - **Case 4:** `go tool task check` exits 0 (gofmt, vet, all tests, `cardtown lint --all`).
  - **For the reviewer:** `SeatSeed` is `seed*2 + 1 + uint64(seat)`, which equals the old `botSeed` branch for seats 0 and 1.
- 2026-10-22 lead-architect: Review: **approved**, merged as cb46eb0, status `done`.
  - Scope: 0386429 changes only the status line, and c2b26b5 adds only `match_test.go`. 9f98029 changes only the four Allowed Files and this file, and no test changed after c2b26b5.
  - Design: it matches the Lead decisions. `MatchDecks` replaces both commands' deck blocks, with unwrapped errors. `SeatSeed` and `DefaultEndBias` replace every non-test literal. The seed-scheme matches left in `play_test.go` and `policy_test.go` are tests' own copies, which decision 4 keeps.
  - Gates, re-run in a scratch worktree:
    - At c2b26b5, vet fails on the undefined names.
    - At the head, `go tool task check` exits 0.
    - Against a baseline built from c2b26b5 (same non-test code as the integration branch), stdout+stderr and exit code are byte-identical (`cmp`) for: the starter sim under both policies, bare sim, `-games 1 -seed 7 -log`, `play` with the bot at seat 1 (progress) and at seat 0 (random), a stress-deck `play` that exercises `MatchDecks`'s other branch, and `sim -deck nonexistent` (exit 1).
  - Board directive #41 is complete.
