---
id: T-0183
title: A hidden decision stays hidden by its flag (directive #119, harness, flash)
kind: harness
tier: flash
wave: D.1
status: ready
branch: m2/autoloop-wave-d.1/T-0183-hidden-decision-flag
depends_on: [T-0182]
claims: [field:DecisionRecord.Hidden, view:LogLines hidden redaction, func:hiddenDecision, test:TestView_HiddenDecisionLineNamesNoOptionAndNoPick, test:TestView_FlaggedSearchKeepsItsWording, test:TestView_OpenDecisionLinePrintsToBothSeats, test:TestReplay_OnlySearchRecordsAreHidden, test:TestReplay_SearchWithoutItsHiddenFlagDiverges]
cards: []
requirements: [FR-DEK-001]
---

## Goal
`DecisionRecord` carries `Decision.Hidden`, and `view.LogLines` redacts any hidden decision of another seat. A search keeps its wording, and any other kind gets a generic line that names no option and no pick. The next card that looks at hidden cards (*Devilhorn Scouting Party*, wave F) needs no second case in `LogLines`.

## Inputs
- Board directive #119 (structural review 8, finding 12, approved by the user, 2026-11-05).
- `pkg/view/log.go:31` hides by kind. `Decision.Hidden` (`pkg/engine/choice.go:41`) already says the choice is secret, and `emitDecision` (`pkg/engine/events.go:125`) doesn't copy it.
- `pkg/engine/effects.go:113` is the only non-test decision that sets `Hidden`, so a search is the only hidden record in a real game today. `pileDecision` is public by its own comment, and `DecideHand` isn't flagged (T-0168's Shape: its picks go to a public discard pile).
- `eventlog.decisionEqual` (`pkg/eventlog/replay.go:250`) names each field it compares, so a new field goes unchecked unless it's added there. `Event.clone` copies the struct, so it needs nothing.
- **Directive #119's two lines that don't fit together.** Its title and fix read the flag, and its case (a) says T-0168's tests pass unedited. `TestView_SearchLineHidesThePickFromTheOpponent` and `TestView_SearchLineSaysNoCardWasFound` hand-build search records with no `Hidden` and expect them redacted, so a check of the flag alone fails both (measured on the dry run). The user approved "unedited", so this task keeps it: the check reads `Hidden || Kind == search`. Every search in a game sets the flag (`TestReplay_OnlySearchRecordsAreHidden`), so only those two records reach the search clause. That clause is a second statement of "a search is secret" (D2.3), so board directive #123, added in this task's commit, holds its removal. It waits on the user, because the removal edits those two records.
- **The sims don't read it:** `cardtown sim` prints no decision record, and `omitempty` leaves an open decision's JSON unchanged, so `sim-compare` is expected to pass.
- The lead dry-ran these exact edits in a throwaway worktree of the run branch at aae9b98. With only the test edits, both test packages exit 1 on the missing field (Case 1). With all the edits: `gofmt -l` clean, the tests pass, `go tool task check` and `go tool task sim` exit 0, and `go tool task sim-compare -- autoloop/m2-wave-d.1` prints `[PASS]`. Every expected output below is from that dry run.
- **Mutations of the dry run, each reverted, each failing a test:** `emitDecision` not copying the flag; the check ignoring the flag; the check redacting the deciding seat too; `decisionEqual` skipping the flag; the JSON tag without `omitempty`; the generic line printing the prompt; a search losing its wording; the search clause removed (the two T-0168 tests above, and no others).

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)

### 1. `pkg/view/view_test.go`, in the `test(T-0183)` commit
Add this right before the comment `// T-0168 Cases 1-4: The play log hides an opponent's deck search pick.`, with nothing else changed:
```go
// T-0183 (directive #119): a hidden decision of any kind names no option and no
// pick to the other seat, and its own seat sees the whole line.
func TestView_HiddenDecisionLineNamesNoOptionAndNoPick(t *testing.T) {
	for p := range engine.PlayerID(2) {
		ev := engine.Event{
			Seq:    40,
			Turn:   3,
			Phase:  "main",
			Kind:   engine.EventDecision,
			Player: int(p),
			Decision: &engine.DecisionRecord{
				Kind:    engine.DecideTarget.String(),
				Player:  int(p),
				Prompt:  "Look at Alpha Secret",
				How:     "answer",
				Hidden:  true,
				Options: []string{"Alpha Secret", "Beta Secret"},
				Refs:    []int{7, 8},
				Picks:   []int{1},
			},
		}
		opp := p.Opponent()
		got := view.LogLines([]engine.Event{ev}, opp)
		want := fmt.Sprintf("40 T3 main decision: player %d makes a hidden choice", p)
		if len(got) != 1 || got[0] != want {
			t.Errorf("seat %d seeing player %d's hidden choice: got %v, want [%s]", opp, p, got, want)
		}
		own := view.LogLines([]engine.Event{ev}, p)
		wantOwn := eventlog.FormatEvent(ev)
		if len(own) != 1 || own[0] != wantOwn || !strings.Contains(own[0], "Beta Secret") {
			t.Errorf("seat %d seeing its own hidden choice: got %v, want [%s]", p, own, wantOwn)
		}
	}
}

// T-0183 (directive #119): a search record with the flag keeps its wording.
func TestView_FlaggedSearchKeepsItsWording(t *testing.T) {
	for _, tc := range []struct {
		picks []int
		want  string
	}{
		{[]int{1}, "40 T3 main decision: player 0 searches their deck and finds a card"},
		{[]int{}, "40 T3 main decision: player 0 searches their deck and finds no card"},
	} {
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
				Hidden:  true,
				Options: []string{"Alpha Secret", "Beta Secret"},
				Refs:    []int{7, 8},
				Picks:   tc.picks,
			},
		}
		got := view.LogLines([]engine.Event{ev}, 1)
		if len(got) != 1 || got[0] != tc.want {
			t.Errorf("picks %v: got %v, want [%s]", tc.picks, got, tc.want)
		}
	}
}

// T-0183 (directive #119): a decision that isn't hidden prints as before, to both
// seats.
func TestView_OpenDecisionLinePrintsToBothSeats(t *testing.T) {
	ev := engine.Event{
		Seq:    40,
		Turn:   3,
		Phase:  "main",
		Kind:   engine.EventDecision,
		Player: 0,
		Decision: &engine.DecisionRecord{
			Kind:    engine.DecideTarget.String(),
			Player:  0,
			Prompt:  "Choose a target",
			How:     "answer",
			Options: []string{"Alpha Open", "Beta Open"},
			Refs:    []int{7, 8},
			Picks:   []int{1},
		},
	}
	want := eventlog.FormatEvent(ev)
	for seat := range engine.PlayerID(2) {
		got := view.LogLines([]engine.Event{ev}, seat)
		if len(got) != 1 || got[0] != want || !strings.Contains(got[0], "Beta Open") {
			t.Errorf("seat %d: got %v, want [%s]", seat, got, want)
		}
	}
}
```

### 2. `pkg/eventlog/replay_test.go`, in the `test(T-0183)` commit
Add this at the very end of the file, after `TestReplay_SearchWithSwappedOptionsDiverges`' closing brace, with nothing else changed:
```go
// T-0183 (directive #119): in real games every search record is flagged hidden
// and no other record is, in the record and in its JSON, where an open decision
// has no hidden key.
func TestReplay_OnlySearchRecordsAreHidden(t *testing.T) {
	searches, open := 0, 0
	for gi, evs := range slices.Concat(searchGames(t), allGames(t)) {
		flagged := 0
		for _, e := range evs {
			if e.Decision == nil {
				continue
			}
			isSearch := e.Decision.Kind == engine.DecideSearch.String()
			if e.Decision.Hidden != isSearch {
				t.Fatalf("game %d: decision %+v, want Hidden only on a search", gi+1, *e.Decision)
			}
			if isSearch {
				flagged++
			} else {
				open++
			}
		}
		searches += flagged
		var buf bytes.Buffer
		if err := eventlog.WriteEventsJSONL(&buf, gi+1, evs); err != nil {
			t.Fatal(err)
		}
		js := buf.String()
		if got := strings.Count(js, `"hidden":`); got != flagged {
			t.Errorf("game %d: JSON has %d hidden keys, want %d, one per search", gi+1, got, flagged)
		}
		if got := strings.Count(js, `"hidden":true`); got != flagged {
			t.Errorf("game %d: JSON has %d hidden:true, want %d", gi+1, got, flagged)
		}
	}
	if searches == 0 || open == 0 {
		t.Fatalf("%d searches and %d other decisions, want some of each", searches, open)
	}
}

// T-0183 (directive #119): the flag is part of the record a replay compares.
func TestReplay_SearchWithoutItsHiddenFlagDiverges(t *testing.T) {
	for _, evs := range searchGames(t) {
		for i, e := range evs {
			if e.Decision == nil || e.Decision.Kind != engine.DecideSearch.String() {
				continue
			}
			cp := append([]engine.Event{}, evs...)
			decCopy := *e.Decision
			decCopy.Hidden = false
			cp[i].Decision = &decCopy
			_, err := eventlog.Replay(cp, len(cp), searchResolve)
			if err == nil || !strings.Contains(err.Error(), "replay diverged") {
				t.Fatalf("search with no flag: error %v, want replay diverged", err)
			}
			return
		}
	}
	t.Fatal("no game holds a search decision")
}
```

### 3. `pkg/engine/events.go`
1. In `DecisionRecord`, replace the lines for `Options`, `Refs`, `Picks` and `How` with these five, which `gofmt` aligns:
   ```go
   	Options []string `json:"options"`          // the option labels offered, in order
   	Refs    []int    `json:"refs"`             // the option refs, in order
   	Picks   []int    `json:"picks"`            // indexes into Options, as applied
   	How     string   `json:"how"`              // "answer", "auto" (fast path), "timeout" (fallback), "invalid" (fallback after an invalid answer)
   	Hidden  bool     `json:"hidden,omitempty"` // Decision.Hidden: a seat's view keeps the options and picks from the other seat
   ```
2. In `emitDecision`, `Prompt: d.Prompt, How: how,` becomes `Prompt: d.Prompt, How: how, Hidden: d.Hidden,`.

### 4. `pkg/engine/choice.go`
On `Decision`'s `Hidden` line, the comment `// the choice inspects hidden cards in a Hand or Deck (no fast path)` becomes `// the choice inspects hidden cards in a Hand or Deck (no fast path, and the play log keeps it from the other seat)`.

### 5. `pkg/eventlog/replay.go`
In `decisionEqual`, `a.Prompt == b.Prompt && a.How == b.How &&` becomes `a.Prompt == b.Prompt && a.How == b.How && a.Hidden == b.Hidden &&`.

### 6. `pkg/view/log.go`
1. In `LogLines`' doc comment, replace the bullet
   ```go
   //   - an opponent's deck search decision names whether a card was found,
   //     never the card picked.
   ```
   with
   ```go
   //   - another seat's hidden decision names no option and no pick: a deck search
   //     says whether a card was found, and any other kind says only that a choice
   //     was made.
   ```
2. Replace the whole search `case`, from `case e.Decision != nil && e.Decision.Kind == engine.DecideSearch.String() && e.Decision.Player != int(seat):` through its `out = append(out, eventlog.FormatEvent(redact(e, detail)))`, with:
   ```go
   		// The search clause is only for T-0168's hand-built records, which don't set
   		// the flag: every search in a game does. Directive #123 drops it.
   		case e.Decision != nil && e.Decision.Player != int(seat) && (e.Decision.Hidden || e.Decision.Kind == engine.DecideSearch.String()):
   			out = append(out, eventlog.FormatEvent(redact(e, hiddenDecision(e.Decision))))
   ```
3. Right before `func matchStart(`, add this, with one blank line after it:
   ```go
   // hiddenDecision is what the other seat sees of d. It names no option and no pick.
   func hiddenDecision(d *engine.DecisionRecord) string {
   	switch {
   	case d.Kind != engine.DecideSearch.String():
   		return fmt.Sprintf("player %d makes a hidden choice", d.Player)
   	case len(d.Picks) > 0:
   		return fmt.Sprintf("player %d searches their deck and finds a card", d.Player)
   	}
   	return fmt.Sprintf("player %d searches their deck and finds no card", d.Player)
   }
   ```

## Acceptance Cases (harness tasks only; written by the lead)
Edits 1 and 2 are the whole of the test commit: commit them as `test(T-0183): a hidden decision stays hidden by its flag`, run case 1, then make edits 3 to 6 and commit them as `feat(view): a hidden decision stays hidden by its flag (T-0183)`. W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **Fails before the edits** (after the test commit, before edit 3):
   - `go -C W test ./pkg/view -run TestView_HiddenDecisionLineNamesNoOptionAndNoPick` exits 1, and its output has `unknown field Hidden`.
   - `go -C W test ./pkg/eventlog -run TestReplay_OnlySearchRecordsAreHidden` exits 1, and its output has `e.Decision.Hidden undefined`.
2. **A search record has `Hidden` set** (#119 (a)): `go -C W test ./pkg/eventlog -run TestReplay_OnlySearchRecordsAreHidden -v` exits 0, and its output has `--- PASS` for that test.
3. **A hidden record of another kind** (#119 (b)): `go -C W test ./pkg/view -run TestView_HiddenDecisionLineNamesNoOptionAndNoPick -v` exits 0, and its output has `--- PASS` for that test.
4. **A record that isn't hidden prints as before** (#119 (c)): `go -C W test ./pkg/view -run TestView_OpenDecisionLinePrintsToBothSeats -v` exits 0, and its output has `--- PASS` for that test.
5. **A search keeps its wording, and T-0168's tests pass unedited** (#119 (a)): `go -C W test ./pkg/view -run TestView_FlaggedSearchKeepsItsWording -v` exits 0 with `--- PASS` for that test, and `go -C W test ./pkg/view -run TestView_SearchLine -v` exits 0 with `--- PASS` for `TestView_SearchLineHidesThePickFromTheOpponent`, `TestView_SearchLineShowsThePickToTheSearcher` and `TestView_SearchLineSaysNoCardWasFound`.
6. **A search still replays** (#119 (d)): `go -C W test ./pkg/eventlog -run TestReplay_Search -v` exits 0, and its output has `--- PASS` for `TestReplay_SearchesRoundTrip`, `TestReplay_SearchWithSwappedOptionsDiverges` and `TestReplay_SearchWithoutItsHiddenFlagDiverges`.
7. **Only the listed edits:** `git -C W diff autoloop/m2-wave-d.1 --numstat -- pkg cmd` prints exactly these lines: `1	1	pkg/engine/choice.go`, `6	5	pkg/engine/events.go`, `1	1	pkg/eventlog/replay.go`, `60	0	pkg/eventlog/replay_test.go`, `18	8	pkg/view/log.go` and `96	0	pkg/view/view_test.go`. The 0 deletions in both test files are what show that no existing test changed.
8. **Nothing else changed:** `git -C W diff autoloop/m2-wave-d.1 --stat` lists only those six files and this task file.
9. **Sims byte-identical** (#119): `go -C W tool task sim-compare -- autoloop/m2-wave-d.1` prints `[PASS]`.
10. `gofmt -l W/pkg/view W/pkg/engine W/pkg/eventlog` prints nothing, and `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/engine/events.go`
- `pkg/engine/choice.go` (comment only)
- `pkg/eventlog/replay.go`
- `pkg/view/log.go`
- `pkg/view/view_test.go` (edit 1 only)
- `pkg/eventlog/replay_test.go` (edit 2 only)

## Done When
- [ ] Cases 1-10 hold, with their commands and results in the Log
- [ ] `go tool task sim-compare -- autoloop/m2-wave-d.1` passes (starter and stress sims byte-identical to the run branch)
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-05 lead-architect (wave D.1 run, session 14): no `main` commits to sync. Not `WAVE-COMPLETE`: #119 and #121 are `open: before the next Deliverables task`. #122 is `open: after Wave D.1 (#121), before Wave E`, so it doesn't apply yet. Wrote this task from #119, next in board order, and #123 for the clause it keeps (D2.3; see Inputs). Harness, so it skips `tests` and goes straight to `ready`; tier `flash`, because the edits are exact and were dry-run. Every command here is on the headless allow-list, one per line, with no pipes, `$`, double quotes or unquoted parentheses. Waiting directives checked: none has triggered (#6, #7: mode is `serial`; #8: no `deck.go`; #15: no turn-boundary hook; #71, #91 (2), #106, #118 (b), #120: no such card). Next: `kernel-engineer` on Flash.
