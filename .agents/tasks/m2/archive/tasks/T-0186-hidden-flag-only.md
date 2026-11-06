---
id: T-0186
title: Drop the search clause from the hidden-decision check (directive #123, harness, flash)
kind: harness
tier: flash
wave: D.1
status: done
branch: m2/T-0186-hidden-flag-only
depends_on: [T-0183]
claims: [view:LogLines hidden check, test:TestView_SearchWithoutTheFlagPrintsWhole]
cards: []
requirements: [FR-DEK-001]
---

## Goal
`LogLines` hides another seat's decision by its `Hidden` flag alone. "A search is secret" is then said once, where the engine sets the flag ([effects.go](../../../../../pkg/engine/effects.go)). No game output changes, because every search in a game already sets it.

## Inputs
- Board directive #123. The user lifted #119's "T-0168's tests pass unedited" line on 2026-11-06, so the two edits below are authorized (D2.4).
- `pkg/view/log.go:32-35`: the clause and its comment.
- `TestReplay_OnlySearchRecordsAreHidden` pins that every search in a game sets the flag.
- `hiddenDecision` keeps its `DecideSearch` case: that is a search's wording, not whether it is hidden.
- The lead dry-ran these exact edits in a throwaway worktree of `main` at d709c4c. With the clause removed and no test edited, exactly `TestView_SearchLineHidesThePickFromTheOpponent` and `TestView_SearchLineSaysNoCardWasFound` fail. With both edits they pass. On the unchanged `log.go`, `TestView_SearchWithoutTheFlagPrintsWhole` fails and nothing else does. With all the edits: `gofmt -l` is clean, and `go tool task check`, `go tool task sim` and `go tool task sim-compare` exit 0.

## Acceptance Cases
1. `pkg/view/log.go`'s check reads only the flag: the clause and its comment are gone.
2. `TestView_SearchLineHidesThePickFromTheOpponent` and `TestView_SearchLineSaysNoCardWasFound` pass with one added field each and no other change.
3. A search record with no `Hidden` prints whole to the other seat (`TestView_SearchWithoutTheFlagPrintsWhole`). This test wasn't in #123. It is added so that nothing can restore the clause unnoticed. It fails before the implementation commit.
4. `go tool task sim-compare` passes: the view is not in the sims.

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)

### 1. `pkg/view/view_test.go`, in the `test(T-0186)` commit
1. In `TestView_SearchLineHidesThePickFromTheOpponent`, right after the line `Picks:   []int{1},` of its record, add:
   ```go
   				Hidden:  true,
   ```
2. In `TestView_SearchLineSaysNoCardWasFound`, right after the line `Picks:   []int{},` of its record, add:
   ```go
   			Hidden:  true,
   ```
3. Right after the closing brace of `TestView_SearchLineSaysNoCardWasFound`, add this, with one blank line before it:
   ```go
   // T-0186 (directive #123): the flag alone hides a decision, so a search record
   // without it prints whole, like any other record. Whether a search is secret is
   // said once, where the engine sets the flag.
   func TestView_SearchWithoutTheFlagPrintsWhole(t *testing.T) {
   	ev := engine.Event{
   		Seq:    40,
   		Turn:   3,
   		Phase:  "main",
   		Kind:   engine.EventDecision,
   		Player: 0,
   		Decision: &engine.DecisionRecord{
   			Kind:    engine.DecideSearch.String(),
   			Player:  0,
   			Prompt:  "Search your deck for a card",
   			How:     "answer",
   			Options: []string{"Alpha Secret", "Beta Secret"},
   			Refs:    []int{7, 8},
   			Picks:   []int{1},
   		},
   	}
   	got := view.LogLines([]engine.Event{ev}, 1)
   	want := eventlog.FormatEvent(ev)
   	if len(got) != 1 || got[0] != want {
   		t.Errorf("seat 1 seeing a search with no flag: got %v, want [%s]", got, want)
   	}
   }
   ```
Remove the three leading spaces of the code blocks' indentation when you copy them. Commit only this file as `test(T-0186): ...`, after confirming that `TestView_SearchWithoutTheFlagPrintsWhole` fails.

### 2. `pkg/view/log.go`
Delete the two comment lines that begin `// The search clause is only for T-0168's hand-built records`, and make the case read:
```go
		case e.Decision != nil && e.Decision.Player != int(seat) && e.Decision.Hidden:
```

## Allowed Files
- `pkg/view/log.go`
- `pkg/view/view_test.go` (Edit 1 only, in the test commit)

## Done When
- [x] `go test ./pkg/view ./pkg/eventlog` passes
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks
- [x] `go tool task sim-compare` passes (starter and stress sims byte-identical to the integration branch)

## Log
- 2026-11-06 (lead-architect): Written for #123, `ready`. The user opened the directive on 2026-11-06.
- 2026-11-06 (kernel-engineer): Implemented for #123. Edit 1 added Hidden: true to TestView_SearchLineHidesThePickFromTheOpponent and TestView_SearchLineSaysNoCardWasFound in pkg/view/view_test.go, plus new test TestView_SearchWithoutTheFlagPrintsWhole (confirmed failing on unchanged log.go, committed as test(T-0186)). Edit 2 removed the search clause and comment in pkg/view/log.go so LogLines checks e.Decision.Hidden alone. Verification: go test ./pkg/view ./pkg/eventlog, go tool task check, go tool task sim, and go tool task sim-compare all passed cleanly. Set status to review.
- 2026-11-06 (lead-architect): Approved on the first round, with no refused command. The diff matches the task's edits. `TestView_SearchWithoutTheFlagPrintsWhole` fails at 1b7119a. Re-ran `go test ./pkg/view ./pkg/eventlog`, `go tool task check`, `go tool task sim` and `go tool task sim-compare`: exit 0. Merged as 8143a01.
