---
id: T-0168
title: The play log hides an opponent's deck search pick, and a search replays (FR-DEK-001, harness, flash)
kind: harness
tier: flash
wave: D
status: review
branch: m2/autoloop-wave-d/T-0168-search-stays-private
depends_on: [T-0167]
claims: [view:LogLines search redaction, comment:doc.go OpSearchDeck]
cards: []
requirements: [FR-DEK-001]
---

## Goal
`FR-DEK-001` gives the searching player, and only that player, sight of the deck. `view.LogLines` prints a decision's picked card names to either seat, so `cardtown play` (`play.go:111`) would tell a human which card the opponent found. Close that before any card can search, and pin that a game with a search replays.

## Inputs
- `FR-DEK-001`, `FR-DEK-002` in FRD 01.
- `pkg/view/log.go`: `LogLines` and `redact`. The leak is `eventlog.FormatEvent`'s `e.Decision` case, which prints `d.Options[p]` for each pick.
- The decision record is the only line that names the card: `stepSearchDeck` (`pkg/engine/effects.go:112`) logs a decision and a shuffle event, and `takeCard` and `putCard` log nothing. The JSON event log and `Replay` keep the full record, because they aren't a seat's view.
- Design note: `docs/architecture/m2-ability-model.md`, Boundaries. `pkg/view` holds what a seat sees.
- Lead probe on the run branch at `29391a7`, since reverted: Case 5's deck under `play(..., 30, 7)` gives 762 searches, all `How` `answer`; 60 seats search twice in one game, and in all 60 the second search lists the cards in a different order from the first's; all 30 games replay; swapping two `Refs` of a game's first search diverges in all 30.

## Shape (lead decisions)
- `LogLines` gains one `case`, ahead of `default`: `e.Decision != nil && e.Decision.Kind == engine.DecideSearch.String() && e.Decision.Player != int(seat)`. It goes through the existing `redact`, with a detail that says only whether a card was found. No second redaction helper.
- The doc comment's "except two records" becomes three, with a bullet for this case.
- Only this kind changes: a `Hand` decision's picks go to a public discard pile, so its line stays FormatEvent's.

## Acceptance Cases (harness tasks only; written by the lead)
W is your worktree. Add the tests first and commit them as `test(T-0168): …`. Cases 1, 3 and 4 fail until `log.go` changes, which the commit records. Cases 2, 5 and 6 pin behavior that already holds, so they pass from their first run: say so in the Log. No existing test changes.

The search record in cases 1 to 3 is an `engine.Event{Seq: 40, Turn: 3, Phase: "main", Kind: engine.EventDecision, Player: P, Decision: &engine.DecisionRecord{Kind: engine.DecideSearch.String(), Player: P, Prompt: "Search your deck for a card", How: "answer", Options: []string{"Alpha Secret", "Beta Secret"}, Refs: []int{7, 8}, Picks: []int{1}}}`.

1. **`TestView_SearchLineHidesThePickFromTheOpponent`, in `pkg/view/view_test.go`:** for P 0 and P 1, `view.LogLines` for the other seat returns exactly `40 T3 main decision: player P searches their deck and finds a card`, with P's number. `go -C W test ./pkg/view -run TestView_SearchLineHidesThePickFromTheOpponent` passes.
2. **`TestView_SearchLineShowsThePickToTheSearcher`, in `pkg/view/view_test.go`:** for P 0 and P 1, `view.LogLines` for seat P returns exactly `eventlog.FormatEvent` of the event, and the line contains `Beta Secret`. Also, a record with the same fields but `Kind: engine.DecideHand.String()` and `Player` 0 returns `eventlog.FormatEvent`'s line for both seats.
3. **`TestView_SearchLineSaysNoCardWasFound`, in `pkg/view/view_test.go`:** with `Picks: []int{}` and P 0, seat 1's line is exactly `40 T3 main decision: player 0 searches their deck and finds no card`. Nothing reaches this today, because a search for *"a card"* must choose one (`SVY-R37`). Directive #106's search with criteria will, and the line must not then claim a card.
4. **`TestView_SearchInAGameStaysPrivate`, in `pkg/view/view_test.go`, through a real game:** set up like `TestView_ActivatedAbilityItemShowsItsTargets`: `engine.NewGame` with a chooser for both seats and `PlayFirst` giving player 0 Turn 1. The chooser answers `Picks: []int{1}`. Using `s.inst`, replace player 0's deck with three instances of `testUnit` cards named `Alpha Search One`, `Alpha Search Two` and `Alpha Search Three`, in that order. Put a Main Theater unit under player 0 whose `Ability` has one `Activated`: `Costs` `{Kind: ability.CostExhaustSelf}`, `Effect.Steps` one `{Op: ability.OpSearchDeck}`. Apply `engine.Action{Kind: engine.ActActivate, Player: 0, Asset: a.ID, Ability: 0}`. Then:
   - `Alpha Search Two` is in player 0's hand;
   - no line of `view.LogLines(g.Log, 1)` contains any of the three names;
   - seat 0's line for the search's decision event is `eventlog.FormatEvent`'s and contains `Alpha Search Two`, and seat 1's line for it ends `player 0 searches their deck and finds a card`;
   - for both seats, the shuffle event's line (kind `engine.EventShuffle`) is `eventlog.FormatEvent`'s and contains `player 0 shuffles their deck`.
5. **`TestReplay_SearchesRoundTrip`, in `pkg/eventlog/replay_test.go`:** add `searchDeck()`, modeled on `activationDeck`: the same HQ and unit fields under the names `Search HQ` and `Search Unit`, with one `Activated`, `Costs` `{Kind: ability.CostExhaustSelf}`, `Effect.Steps` one `{Op: ability.OpSearchDeck}`, and 30 copies of the unit. Add `searchNames`, `searchResolve` and `searchGames(t)` beside it, as `activationNames`, `activationResolve` and `activationGames(t)` are, with `play(..., 30, 7)`. For every game, `eventlog.Replay(evs, len(evs), searchResolve)` returns no error. Over the 30 games:
   - at least one decision has `Kind` `engine.DecideSearch.String()`;
   - at least one seat searches twice in one game with a shuffled order between: take the two decisions' `Refs`, remove from the first its picked ref and every ref absent from the second, and the result differs from the second's `Refs`.

   That second search's `Refs` are the shuffle's result, and `Replay` compares them, so this is what pins the shuffle's draws from the game PRNG. Fail with a message if either count is zero.
6. **`TestReplay_SearchWithSwappedOptionsDiverges`, in `pkg/eventlog/replay_test.go`:** take the first decision in `searchGames(t)` of that kind with at least two `Refs`. Copy the event slice as `TestReplay_ActivationWithAnotherIndexDiverges` does, copy that event's `DecisionRecord` and its `Refs` so the shared record is untouched, and swap the first two `Refs` in the copy. `eventlog.Replay(cp, len(cp), searchResolve)` returns an error containing `replay diverged`.
7. **The sims don't read it:** `go -C W tool task sim-compare -- autoloop/m2-wave-d` prints `[PASS]`.
8. **`pkg/engine/doc.go` names the search, comments only:** in the Chooser bullet, `choosing from your own hand (DecideHand, SVY-R16)` becomes `choosing from your own hand (DecideHand, SVY-R16) and from your own deck (DecideSearch, SVY-R37, SVY-R38)`. In the Steps bullet, `OpKill (inline destruction) and OpMoveZone` becomes `OpKill (inline destruction), OpMoveZone and OpSearchDeck`.
9. **Nothing else changed:** `git -C W diff autoloop/m2-wave-d --stat` lists only `pkg/view/log.go`, `pkg/view/view_test.go`, `pkg/eventlog/replay_test.go`, `pkg/engine/doc.go` and this task file.

## Allowed Files
- `pkg/view/log.go`
- `pkg/engine/doc.go` (comments only)
- `pkg/view/view_test.go` (the four new tests only; no existing test changes)
- `pkg/eventlog/replay_test.go` (the two new tests and their helpers only; no existing test changes)

## Done When
- [x] Cases 1-9 hold, with their commands and results in the Log
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks
- [x] `go tool task sim-compare -- autoloop/m2-wave-d` passes

## Log
- 2026-11-04 lead-architect (wave D run, session 12): split off T-0167 (D2.4). Queued.
- 2026-11-04 lead-architect (wave D run, session 13): wrote Inputs, Shape, Acceptance Cases, Allowed Files and Done When against the run branch at `29391a7`; `main` has no commits to sync. Harness, so it skips `tests` and goes straight to `ready`; tier `flash`, because the shape is one `case` and the tests are named line by line. Every command here is on the headless allow-list, one per line. Probed the replay cases on a synthetic deck (see Inputs). Waiting directives checked: none has triggered (#6, #7: mode is `serial`; #8: no `deck.go`; #15: no turn-boundary hook; #71, #91 (2), #106: no such card). #105's parts (2), (3) stay with T-0169. Next: `kernel-engineer` on Flash.
- 2026-11-05 lead-architect (wave D run, session 14): session 13 hit the usage limit before committing this file, so its text was still uncommitted in its worktree. Checked it against the run branch at `29391a7` and committed it; the one change is Case 8's quote, whose old form spanned a line break in `doc.go`. `main` has no commits to sync, and the directives check above still holds.
- 2026-11-05 kernel-engineer: added acceptance tests in `pkg/view/view_test.go` (Cases 1-4) and `pkg/eventlog/replay_test.go` (Cases 5-6). Initial run verified Cases 1, 3, and 4 failed while Cases 2, 5, and 6 passed, committed under `test(T-0168)`. Implemented Shape: `pkg/view/log.go` gains a redaction case in `LogLines` ahead of `default` for `e.Decision != nil && e.Decision.Kind == engine.DecideSearch.String() && e.Decision.Player != int(seat)` using `redact(e, detail)` announcing whether a card was found (`"player %d searches their deck and finds a card"` vs `"player %d searches their deck and finds no card"`), and updated the doc comment. Updated `pkg/engine/doc.go` comment bullets for Chooser and Steps to document `DecideSearch` and `OpSearchDeck`. Reused existing `redact` helper and `eventlog.FormatEvent`. Ran: `go test ./pkg/view -run TestView_SearchLineHidesThePickFromTheOpponent` (exit 0), `go test ./pkg/view -run TestView_SearchLineShowsThePickToTheSearcher` (exit 0), `go test ./pkg/view -run TestView_SearchLineSaysNoCardWasFound` (exit 0), `go test ./pkg/view -run TestView_SearchInAGameStaysPrivate` (exit 0), `go test ./pkg/eventlog -run TestReplay_SearchesRoundTrip` (exit 0), `go test ./pkg/eventlog -run TestReplay_SearchWithSwappedOptionsDiverges` (exit 0), `go tool task sim` (exit 0), `go tool task sim-compare -- autoloop/m2-wave-d` (exit 0, `[PASS] starter and stress sims match autoloop/m2-wave-d byte for byte, exit codes included`), `go tool task check` (exit 0; gofmt, vet, all package tests, cardtown lint --all passed), `git diff autoloop/m2-wave-d --stat` confirmed only the 5 allowed files changed. Status `review`.
