---
id: T-0239
title: "One function names a stack item: `Game.StackLabel` (directive #153, wave F.3 harness, flash)"
kind: harness
tier: flash
wave: F.3
status: in-progress
branch: m2/autoloop-wave-f.3/T-0239-stack-label
depends_on: []
claims: [func:StackLabel, field:StackView.Label]
cards: []
requirements: [FR-ZON-005]
---

## Goal
Which card a stack item names, and its wording, are written once: `Game.StackLabel` in `pkg/engine/stack.go` returns a stack item's label ("Deploy of X", "March", "Attack", "Scratch", "Maneuver X", "ability of X", "<keyword> trigger of X"), reading an ability's card through `SourceCard` (FR-ZON-005). `push` writes its place event's text from it, and for an ability its `Event.Card` from `SourceCard`, with `Event.Inst` 0 when the shown card isn't the instance's own (a face-down source). `view.StackView` gains `Label`, filled from it, and the play screen's `stackLine` prints an ability's `Label` instead of rebuilding the wording. A pure move: sims are byte-identical, because no M2 ability source is face down while its item waits. When this merges, #153 is `applied`.

## Inputs
- Board directive #153 (structural review 12, finding 1, approved by user 2026-11-14), `open: before the next Deliverables task`. Wave F.3, first (design note Deliverable 6l).
- The three copies today (at d4b38f2): `push`'s `what` and `cardName(item.Inst)` ([stack.go:56-82](../../../pkg/engine/stack.go)), which names the real card; `stackLine`'s ability case ([play_screen.go:103-108](../../../cmd/cardtown/play_screen.go)), which rebuilds "Ability of" / "trigger of" from `StackView.Card` and `Keyword`; and the face-aware `shownCard`/`SourceCard` ([placement.go:62-78](../../../pkg/engine/placement.go)) that `view` reads for `StackView.Card`.
- Design (lead decisions):
  - **One function, exported.** `func (g *Game) StackLabel(it *StackItem) string`, beside `push` in `stack.go`. `view` calls it, so it is exported. Deploy and Maneuver read `cardName(it.Inst)` (nil-safe; `push` itself still requires an `Inst` for them, as today).
  - **Event.Inst for an ability** is `item.Inst.ID` only when `SourceCard(item) == item.Inst.Card`. A face-up source on the battlefield shows `a.Inst.Card` (`Asset.Printed`, [state.go:85-90](../../../pkg/engine/state.go)), and a source that has left shows `Inst.Card`, so both keep their ID as today; a face-down source shows `scratchFace`, so its ID is dropped. Every ability item's `Inst` is its source asset's instance (`activatedPlacement`, `placeTrigger`'s pending trigger, `replace.go:112`), so nothing else changes.
  - **The play screen** keeps `stackLine`'s Deploy, Maneuver and other lines. Its ability case prints `Label` with its first letter upper-cased, so "ability of X" prints "Ability of X" and "Aggressive trigger of X" is unchanged.
- **One existing test's fixture changes, its expected output doesn't.** Directive #153 asks for both "ability of" and "trigger of" written only in `StackLabel`, and `TestPlay_PrintScreenNamesEachStackItemsCard` passing unedited. That test builds its `view.StackView` values by hand, with no `Label`, and `stackLine` can no longer build the wording itself, so both can't hold. The duplication is the directive's point, so this task keeps the first and edits only that test's two ability fixture lines to carry the `Label` the view now fills (Test edit 1). Its `want` text is not touched: the printed screen is unchanged. No other existing test changes.
- Reuse search (D2.3): `git grep -n '"ability of \|trigger of "' -- pkg cmd ':!*_test.go'` at d4b38f2 finds only `stack.go:77-79` and `play_screen.go:104-106`; no other package parses the place text (`pkg/sim`, `pkg/eventlog`, `cmd` searched for "places"). After the edits it finds only `stack.go` (the code and its doc comment) and `view.go`'s `Label` field comment.
- Size (D2.4): three non-test files.
- **Expected output difference:** none. `go tool task sim-compare` passes, and the play screen prints the same stack lines.
- `doc.go`'s Files list gains `StackLabel` under #154 (d), not here.
- The lead dry-ran these exact edits and test files in a throwaway worktree of the run branch at d4b38f2. Before the edits, the new tests don't build (`StackLabel` and `StackView.Label` are undefined). With them: `gofmt -l cmd pkg` clean; `go test ./cmd/cardtown ./pkg/view/...` ok; `go tool task check` exit 0; `go tool task sim-compare -- autoloop/m2-wave-f.3` `[PASS]`; `cardtown lint fences` exit 0; `go tool task sim` exit 0 (200 games, P0 58.5% / P1 39.5% / draws 2.0%). Blobs: `stack.go` f41cdf5, `view.go` 454c6f2, `play_screen.go` ed269a0, `play_source_test.go` 28e4de2, `pkg/engine/stack_label_test.go` 16224ea, `pkg/view/stack_label_test.go` 1d41c27.

## Commits
1. `test(T-0239): one function names a stack item` with only the two new test files below and Test edit 1. Before committing, run `go -C W vet ./pkg/engine ./pkg/view` and record that it fails (`StackLabel` and `Label` undefined).
2. `refactor(engine): one function names a stack item, StackLabel (T-0239)` with Edits 1-3.

## Tests (lead decisions; copy exactly, tabs for indentation)
### New file `pkg/engine/stack_label_test.go`
```go
package engine

import (
	"strings"
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0239 (directive #153, FR-ZON-005): one function, StackLabel,
// words what a stack item is, and the place event reads it. An ability names
// its card as both seats see it, so a face-down source shows only the Scratch
// face, and its place event carries no card instance ID. Cards are synthetic;
// player 0 controls.

func TestStackLabel_EachKind(t *testing.T) {
	s := newScene(t)
	a := s.put(0, ZoneMainTheater, unit("Open Source", 1, 1, Keywords{}))
	d := s.give(0, unit("Deployed", 1, 1, Keywords{}))
	m := s.give(0, maneuver("Played", 0, 0, anotherEffect()))
	gone := s.inDiscard(0, unit("Gone", 1, 1, Keywords{}))[0]
	for _, tc := range []struct {
		it   *StackItem
		want string
	}{
		{&StackItem{Kind: StackDeploy, Inst: d, Zone: ZoneReserves}, "Deploy of Deployed"},
		{&StackItem{Kind: StackMarch, Source: a.ID, Zone: ZoneMainTheater}, "March"},
		{&StackItem{Kind: StackAttack, Source: a.ID}, "Attack"},
		{&StackItem{Kind: StackScratch}, "Scratch"},
		{&StackItem{Kind: StackManeuver, Inst: m}, "Maneuver Played"},
		{&StackItem{Kind: StackAbility, Source: a.ID, Inst: a.Inst}, "ability of Open Source"},
		{&StackItem{Kind: StackAbility, Source: a.ID, Inst: a.Inst, Keyword: ability.KeywordAggressive}, "Aggressive trigger of Open Source"},
		{&StackItem{Kind: StackAbility, Source: 999, Inst: gone, Tracked: true}, "ability of Gone"},
	} {
		if got := s.g.StackLabel(tc.it); got != tc.want {
			t.Errorf("StackLabel(%v item) = %q, want %q", tc.it.Kind, got, tc.want)
		}
	}
}

func TestStackLabel_FaceDownSourceShowsOnlyTheScratchFace(t *testing.T) {
	s := newScene(t)
	a := s.put(0, ZoneMainTheater, unit("Hidden Source", 1, 1, Keywords{}))
	a.FaceDown = true
	it := &StackItem{Kind: StackAbility, Controller: 0, Source: a.ID, Inst: a.Inst}
	if got, want := s.g.StackLabel(it), "ability of Scratch Company"; got != want {
		t.Errorf("StackLabel = %q, want %q (FR-ZON-005)", got, want)
	}
	s.g.push(it)
	pl := eventsOfKind(s.g.Log, "place")
	if len(pl) == 0 {
		t.Fatalf("push logged no place event")
	}
	e := pl[len(pl)-1]
	if e.Card != "Scratch Company" {
		t.Errorf("place event Card = %q, want the Scratch face", e.Card)
	}
	if e.Inst != 0 {
		t.Errorf("place event Inst = %d, want 0: an instance ID could identify the face-down card", e.Inst)
	}
	if want := "player 0 places ability of Scratch Company on the stack"; e.Detail != want {
		t.Errorf("place event Detail = %q, want %q", e.Detail, want)
	}
	if strings.Contains(e.Detail, "Hidden Source") || e.Card == "Hidden Source" {
		t.Errorf("the place event names the face-down card: %+v", e)
	}
}

func TestStackLabel_FaceUpSourcePlaceEventNamesItsInstance(t *testing.T) {
	s := newScene(t)
	a := s.put(0, ZoneMainTheater, unit("Open Source", 1, 1, Keywords{}))
	s.g.push(&StackItem{Kind: StackAbility, Controller: 0, Source: a.ID, Inst: a.Inst, Keyword: ability.KeywordAggressive})
	pl := eventsOfKind(s.g.Log, "place")
	if len(pl) == 0 {
		t.Fatalf("push logged no place event")
	}
	e := pl[len(pl)-1]
	if e.Card != "Open Source" || e.Inst != a.Inst.ID {
		t.Errorf("place event Card %q, Inst %d; want %q, %d", e.Card, e.Inst, "Open Source", a.Inst.ID)
	}
	if want := "player 0 places Aggressive trigger of Open Source on the stack"; e.Detail != want {
		t.Errorf("place event Detail = %q, want %q", e.Detail, want)
	}
}
```

### New file `pkg/view/stack_label_test.go`
```go
package view_test

import (
	"slices"
	"strings"
	"testing"

	"animus-mundi/pkg/engine"
	"animus-mundi/pkg/engine/ability"
	"animus-mundi/pkg/view"
)

// Tests for T-0239 (directive #153, FR-ZON-005): a stack item's Label is the
// engine's own wording (Game.StackLabel), and an ability whose source is face
// down shows only the Scratch face to both seats.

func TestView_StackItemLabel(t *testing.T) {
	s := newScene(t)
	up := s.put(0, engine.ZoneMainTheater, testUnit("Open Source", 1, 1, engine.Keywords{}))
	down := s.put(1, engine.ZoneMainTheater, testUnit("Secret Source", 1, 1, engine.Keywords{}))
	down.FaceDown = true
	s.g.Stack = []*engine.StackItem{
		{Kind: engine.StackAbility, Controller: 0, Source: up.ID, Inst: up.Inst, Keyword: ability.KeywordAggressive},
		{Kind: engine.StackAbility, Controller: 1, Source: down.ID, Inst: down.Inst},
		{Kind: engine.StackMarch, Controller: 0, Source: up.ID, Zone: engine.ZoneMainTheater},
	}
	want := []string{"Aggressive trigger of Open Source", "ability of Scratch Company", "March"}
	for seat := engine.PlayerID(0); seat < 2; seat++ {
		v := view.PlayerView(s.g, seat)
		var got []string
		for _, it := range v.Stack {
			got = append(got, it.Label)
		}
		if !slices.Equal(got, want) {
			t.Errorf("seat %d: labels %q, want %q", seat, got, want)
		}
		if js := mustJSON(t, v.Stack); strings.Contains(js, "Secret Source") {
			t.Errorf("seat %d: the stack names the face-down card: %s", seat, js)
		}
	}
}
```

### Test edit 1. `cmd/cardtown/play_source_test.go`, `TestPlay_PrintScreenNamesEachStackItemsCard`: the two ability fixtures carry their Label
Lines 120-121, which read
```go
		{Kind: "Ability", Controller: 1, Source: 12, Card: "Devilhorn Captain"},
		{Kind: "Ability", Controller: 0, Source: 7, Card: "Ripper Pack", Keyword: "Aggressive"},
```
become
```go
		{Kind: "Ability", Controller: 1, Source: 12, Card: "Devilhorn Captain", Label: "ability of Devilhorn Captain"},
		{Kind: "Ability", Controller: 0, Source: 7, Card: "Ripper Pack", Keyword: "Aggressive", Label: "Aggressive trigger of Ripper Pack"},
```
Nothing else in that file changes; its `want` text stays as it is.

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)
Line numbers are at d4b38f2.

### 1. `pkg/engine/stack.go`: `push` reads `StackLabel`, and `StackLabel` is added
Lines 56-82 of `push`, which read
```go
	e := Event{Kind: EventPlace, Player: int(item.Controller), Source: item.Source, Amount: len(g.Stack)}
	var what string
	switch item.Kind {
	case StackDeploy:
		e.Card, e.Inst = item.Inst.Card.Name, item.Inst.ID
		what = "Deploy of " + e.Card
	case StackMarch:
		what = "March"
	case StackAttack:
		e.Target = item.Target
		what = "Attack"
	case StackScratch:
		what = "Scratch"
	case StackManeuver:
		e.Card, e.Inst = item.Inst.Card.Name, item.Inst.ID
		what = "Maneuver " + e.Card
	case StackAbility:
		e.Card, e.Inst = cardName(item.Inst), 0
		if item.Inst != nil {
			e.Inst = item.Inst.ID
		}
		what = "ability of " + e.Card
		if item.Keyword != ability.KeywordNone {
			what = item.Keyword.String() + " trigger of " + e.Card
		}
	}
	g.logEvent(e, "player %d places %s on the stack", item.Controller, what)
}
```
become
```go
	e := Event{Kind: EventPlace, Player: int(item.Controller), Source: item.Source, Amount: len(g.Stack)}
	switch item.Kind {
	case StackDeploy, StackManeuver:
		e.Card, e.Inst = item.Inst.Card.Name, item.Inst.ID
	case StackAttack:
		e.Target = item.Target
	case StackAbility:
		// The card both seats see (SourceCard), and its instance ID only when
		// that is the instance's own card: a face-down source's ID could
		// identify the real card (FR-ZON-005).
		if c := g.SourceCard(item); c != nil {
			e.Card = c.Name
			if item.Inst != nil && c == item.Inst.Card {
				e.Inst = item.Inst.ID
			}
		}
	}
	g.logEvent(e, "player %d places %s on the stack", item.Controller, g.StackLabel(item))
}

// StackLabel names stack item it, in the one wording of what it is: "Deploy of
// X", "March", "Attack", "Scratch", "Maneuver X", "ability of X" or "<keyword>
// trigger of X". An ability names its card as both seats see it (SourceCard,
// FR-ZON-005), or no card if it has none. The place event and the play
// screen's stack listing (view.StackView.Label) both read it.
func (g *Game) StackLabel(it *StackItem) string {
	switch it.Kind {
	case StackDeploy:
		return "Deploy of " + cardName(it.Inst)
	case StackMarch:
		return "March"
	case StackAttack:
		return "Attack"
	case StackScratch:
		return "Scratch"
	case StackManeuver:
		return "Maneuver " + cardName(it.Inst)
	case StackAbility:
		name := ""
		if c := g.SourceCard(it); c != nil {
			name = c.Name
		}
		if it.Keyword != ability.KeywordNone {
			return it.Keyword.String() + " trigger of " + name
		}
		return "ability of " + name
	}
	return ""
}
```

### 2. `pkg/view/view.go`: `StackView.Label`, filled from `StackLabel`
(a) After line 95 (`Keyword    string  // Ability: the keyword its trigger came from ...`) add this line:
```go
	Label      string  // what the item is, as the engine's place event words it (Game.StackLabel): "ability of X", "Aggressive trigger of X", "Deploy of X", "March", ...
```
(b) Line 197, which reads
```go
	s := StackView{Kind: stackKinds[it.Kind], Controller: int(it.Controller), Source: it.Source}
```
becomes
```go
	s := StackView{Kind: stackKinds[it.Kind], Controller: int(it.Controller), Source: it.Source, Label: g.StackLabel(it)}
```

### 3. `cmd/cardtown/play_screen.go`: `stackLine` prints an ability's Label
Lines 93-108, which read
```go
// stackLine is the stack listing's line for item. An ability names its card as
// the engine's place event does ("ability of", or "<keyword> trigger of"), a
// Deploy its card and zone, and a Maneuver its card (#149); a March, an Attack,
// a Scratch, and an ability with no card name their source asset.
func stackLine(item view.StackView) string {
	switch {
	case item.Kind == "Maneuver":
		return fmt.Sprintf("Maneuver: %s (Player %d)", item.Card, item.Controller)
	case item.Kind == "Deploy" && item.Card != "":
		return fmt.Sprintf("Deploy of %s to %s (Player %d)", item.Card, item.Zone, item.Controller)
	case item.Kind == "Ability" && item.Card != "":
		what := "Ability of " + item.Card
		if item.Keyword != "" {
			what = item.Keyword + " trigger of " + item.Card
		}
		return fmt.Sprintf("%s (Player %d, Source %d)", what, item.Controller, item.Source)
```
become
```go
// stackLine is the stack listing's line for item. An ability prints its Label,
// the engine's own wording (Game.StackLabel), with a capital first letter, a
// Deploy its card and zone, and a Maneuver its card (#149); a March, an Attack,
// a Scratch, and an ability with no card name their source asset.
func stackLine(item view.StackView) string {
	switch {
	case item.Kind == "Maneuver":
		return fmt.Sprintf("Maneuver: %s (Player %d)", item.Card, item.Controller)
	case item.Kind == "Deploy" && item.Card != "":
		return fmt.Sprintf("Deploy of %s to %s (Player %d)", item.Card, item.Zone, item.Controller)
	case item.Kind == "Ability" && item.Card != "" && item.Label != "":
		return fmt.Sprintf("%s%s (Player %d, Source %d)", strings.ToUpper(item.Label[:1]), item.Label[1:], item.Controller, item.Source)
```
(`strings` is already imported.) No other file changes.

## Acceptance Cases (harness tasks only; written by the lead)
W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **The tests failed first:** at the `test(T-0239)` commit, before Edits 1-3, `go -C W vet ./pkg/engine ./pkg/view` fails: `StackLabel` and `Label` are undefined.
2. **One label per kind, an ability's through SourceCard:** `go -C W test ./pkg/engine -run TestStackLabel_` passes: each stack item kind's label as Goal lists it; a face-down source's label and its place event name only the Scratch face, with `Inst` 0; a face-up source's place event names its card and its instance ID, as before.
3. **The view carries the label** (FR-ZON-005): `go -C W test ./pkg/view -run TestView_StackItemLabel` passes for both seats, and the face-down card's name is nowhere in the stack view.
4. **The play screen is unchanged:** `go -C W test ./cmd/cardtown -run TestPlay_` passes, `TestPlay_PrintScreenNamesEachStackItemsCard` with only Test edit 1.
5. **One wording:** `git -C W grep -n -e "ability of " -e "trigger of " -- pkg cmd ":!*_test.go"` prints exactly nine lines: `ability.go:85`, `actions.go:31`, `doc.go:49` and `state.go:156` (comments about other things, unchanged), `stack.go:77`, `:78`, `:99` and `:101` (`StackLabel`'s comment and code), and `view.go:96` (`Label`'s comment). Nothing in `cmd/cardtown/play_screen.go`.
6. **Sims byte-identical:** `go -C W tool task sim-compare -- autoloop/m2-wave-f.3` prints `[PASS] starter and stress sims match autoloop/m2-wave-f.3 byte for byte, exit codes included`.
7. **Only the listed edits:** `git -C W diff autoloop/m2-wave-f.3 --numstat -- pkg cmd` prints exactly six lines: `4	8	cmd/cardtown/play_screen.go`, `2	2	cmd/cardtown/play_source_test.go`, `37	16	pkg/engine/stack.go`, `84	0	pkg/engine/stack_label_test.go`, `2	1	pkg/view/view.go` and `41	0	pkg/view/stack_label_test.go`.
8. **Fences:** `go -C W run ./cmd/cardtown lint fences` exits 0.
9. **Nothing else changed:** `git -C W diff autoloop/m2-wave-f.3 --stat` lists only the six files of case 7 and this task file.
10. `gofmt -l W/cmd W/pkg` prints nothing; `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/engine/stack.go`
- `pkg/view/view.go`
- `cmd/cardtown/play_screen.go`
- `pkg/engine/stack_label_test.go` (new; the `test(T-0239)` commit only)
- `pkg/view/stack_label_test.go` (new; the `test(T-0239)` commit only)
- `cmd/cardtown/play_source_test.go` (Test edit 1 only; the `test(T-0239)` commit only)

## Done When
- [ ] Cases 1-10 hold, with their commands and results in the Log
- [ ] `go tool task sim-compare -- autoloop/m2-wave-f.3` passes (starter and stress sims byte-identical to the run branch)
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-14, lead-architect (wave F.3 run, session 1): written in full against d4b38f2 for #153. Run file created; nothing on `main` to sync (`main` is the run branch's base, d4b38f2); the scratch folder was empty. Not `WAVE-COMPLETE` before the task (#153 and #154 `open: before the next Deliverables task`). Harness, so it skips `tests` and goes straight to `ready`. Tier `flash`, because the edits and the test files are exact and were dry-run. Test edit 1 is the one change to an existing test (see Inputs: the directive's two acceptance lines can't both hold, and the fixture edit keeps the screen's expected text as it is). Every command here is on the headless allow-list: one per line, no pipes, no `$`, no unquoted parentheses. Waiting directives checked: #6, #7 (mode `serial`), #15 (no turn-boundary hook), #71, #91 (2), #106, #120, #141 (3), #142: not fired. Next: `kernel-engineer` on Flash.
