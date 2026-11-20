---
id: T-0289
title: "The play screen's stack listing shows each item's chosen targets (directive #175; wave H.2 harness, flash)"
kind: harness
tier: flash
wave: H.2 (board directive #175)
status: in-progress
branch: m2/autoloop-wave-h.2/T-0289-stack-line-shows-targets
depends_on: [T-0288]
claims: [func:main.stackLine, func:main.stackItemLine, func:main.chosenTargetsText, func:main.assetRef, method:view.View.Asset]
cards: []
requirements: [FR-TGT-001]
---

## Goal
The play screen's stack listing shows, after each stack item's line, the targets it chose that are still there (`FR-TGT-001`: chosen targets are public). It reads `view.StackView.Chosen` and `ChosenCards`, which nothing outside tests read until now, so the view fields stay the surface a later client reads. The asset lookup by ID that `actionLabel` wrote inside itself becomes one method of the view, `view.View.Asset`, which `actionLabel` and the stack listing both call, and the screen's "Name#ID" is written once (`assetRef`). When this merges, #175 is `applied`.

## Inputs
- Board directive #175 (structural review 14, finding 6 (b)), approved by the user 2026-11-19, `open`; wave H.2, after #174 (T-0288, `done`), which edited the same function.
- Where it is today (at `autoloop/m2-wave-h.2` b33c338):
  - `stackLine(item)` ([play_screen.go:98-110](../../../cmd/cardtown/play_screen.go)) prints what the item is, and nothing of its targets.
  - `actionLabel` ([play_screen.go:131-155](../../../cmd/cardtown/play_screen.go)) holds `lookupAsset`, a closure that finds an `AssetView` by ID in both HQs, both Reserves and the Main Theater; nothing else can call it. `findAsset` and `formatAsset` each write `"%s#%d"`.
  - `view.StackView.Chosen` and `ChosenCards` ([view.go:96-98](../../../pkg/view/view.go)) are filled by `chosenTargets` ([view.go:177-193](../../../pkg/view/view.go)): per slot, the chosen asset IDs (empty for a pile slot) and, for a pile slot, the names of its chosen cards still in a discard pile (empty for an asset slot). Only Ability and Maneuver items have them.
- Design (lead decisions, the directive's fix):
  - `view.View.Asset(id) *AssetView` (new, in `view.go`, beside `View`) is the lookup, a method of the view so a later client has it too. It searches both seats' HQ and Reserves, then the Main Theater; IDs are unique, so the order only fixes which field is read. `actionLabel` calls it in place of `lookupAsset`, which goes (ADR 0007: no forwarding closure).
  - `stackLine(v, item)` is `stackItemLine(item)` (the body #174 wrote, renamed, unchanged) followed by `" -> "` and `chosenTargetsText(v, item)` when that is not empty, and by nothing otherwise. The arrow is `->`, not the `→` the directive's examples use: every arrow the play screen already prints is ASCII `->` (`actionLabel`'s Generate and Attack lines), and nothing in `cmd/cardtown`, `pkg/view` or `pkg/eventlog` prints a non-ASCII character, which a legacy Windows console would garble. Every other separator is the directive's.
  - `chosenTargetsText` walks the slots in order (as many as the longer of `Chosen` and `ChosenCards`). An asset slot's picks still on the battlefield (`v.Asset` not nil) are named by `assetRef`, in pick order, joined by `", "`. A slot with no such asset but cards in `ChosenCards` prints those names joined by `", "`, then `" (discard pile)"` once. A slot with nothing left is left out. The slots are joined by `"; "`.
  - `assetRef(a)` is `"<a.Card>#<a.ID>"`, the screen's one way to name an asset. `findAsset` and `formatAsset` call it. A face-down asset's `Card` is already the Scratch face's name in the view (`assetView`, `Printed()`), so case (e) needs no code. The engine's `Asset.String` ([state.go:92-98](../../../pkg/engine/state.go)) is the event log's naming ("face-down Scratch Company#9"), a different wording on purpose, so it is not a copy.
  - No engine change. The `view` package doc lists what a seat sees; a lookup over what it already sees adds nothing to that list. #176 brings the docs up to date with the wave.
- Reuse search (D2.3): `git grep -n "Reserves\[i\].ID == id" -- pkg cmd` finds only `lookupAsset`; `git grep -n "Chosen" -- cmd` finds nothing; `git grep -n '"%s#%d' -- cmd pkg/view` finds `findAsset` and `formatAsset` only.
- Size (D2.4): two non-test files and one new test file.
- **Expected output difference:** none in the sims, which print no play screen. `go tool task sim-compare` passes.
- The lead dry-ran these exact edits in a throwaway worktree of the run branch at b33c338. With Edit 3 and without Edits 1-2, `go test ./cmd/cardtown -run TestPlay_StackLine -v` failed in exactly six tests and passed `TestPlay_StackLineWithoutTargetsIsUnchanged` (case (f) pins that nothing else changes, so it passes before and after). With Edits 1-2 all seven passed, and so did every `TestPlay_` test. `gofmt -l pkg cmd` was clean, `go vet ./cmd/... ./pkg/view/...` clean, `go tool task check` exited 0, `go tool task sim` exited 0, and `go tool task sim-compare -- autoloop/m2-wave-h.2` printed `[PASS]`. Blobs: `play_screen.go` ce8d6eb, `view.go` 6b823f3, `play_targets_test.go` 27bb89e.

## Commits
1. `test(T-0289): the stack listing shows chosen targets` with Edit 3 only. Run case 1 before committing it and record the result.
2. `feat(cardtown): the play screen's stack listing shows each item's chosen targets (T-0289)` with Edits 1-2.
No double quotes in either message: the allow-list refuses an escaped one.

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)
### 1. `pkg/view/view.go`: `View.Asset`
```diff
@@ -36,6 +36,28 @@ type View struct {
 	Stack               []StackView // bottom first
 }
 
+// Asset is the battlefield asset with ID id as v shows it, in either seat's
+// Headquarters Zone or Reserves or in the Main Theater, or nil when no asset
+// there has that ID.
+func (v View) Asset(id int) *AssetView {
+	for _, s := range []Side{v.You, v.Opponent} {
+		if s.HQ != nil && s.HQ.ID == id {
+			return s.HQ
+		}
+		for i := range s.Reserves {
+			if s.Reserves[i].ID == id {
+				return &s.Reserves[i]
+			}
+		}
+	}
+	for i := range v.Theater {
+		if v.Theater[i].ID == id {
+			return &v.Theater[i]
+		}
+	}
+	return nil
+}
+
 // Side is one seat's zones as the viewing seat sees them.
 type Side struct {
 	HQ        *AssetView // nil once destroyed
```

### 2. `cmd/cardtown/play_screen.go`: `stackLine` shows chosen targets; `lookupAsset` goes
```diff
@@ -80,7 +80,7 @@ func printScreen(out io.Writer, v view.View, newLogs []string) {
 	if len(v.Stack) > 0 {
 		fmt.Fprintln(out, "Resolution Stack:")
 		for _, item := range v.Stack {
-			fmt.Fprintf(out, "  %s\n", stackLine(item))
+			fmt.Fprintf(out, "  %s\n", stackLine(v, item))
 		}
 	} else {
 		fmt.Fprintln(out, "Resolution Stack: (empty)")
@@ -90,12 +90,22 @@ func printScreen(out io.Writer, v view.View, newLogs []string) {
 	fmt.Fprintf(out, "Discards: You %d | Opponent %d\n\n", len(v.You.Discard), len(v.Opponent.Discard))
 }
 
-// stackLine is the stack listing's line for item: its Label, the engine's one
-// wording of what it is (Game.StackLabel), with a capital first letter, then
-// the declared zone of a Deploy, then its controller, and the source asset of
-// anything but a Deploy or a Maneuver. An item with no Label, and an ability
-// with no card, print their Kind.
-func stackLine(item view.StackView) string {
+// stackLine is the stack listing's line for item: what it is (stackItemLine),
+// then, after " -> ", the targets it chose that are still there
+// (chosenTargetsText). Chosen targets are public (FR-TGT-001).
+func stackLine(v view.View, item view.StackView) string {
+	if targets := chosenTargetsText(v, item); targets != "" {
+		return stackItemLine(item) + " -> " + targets
+	}
+	return stackItemLine(item)
+}
+
+// stackItemLine is what item is: its Label, the engine's one wording of it
+// (Game.StackLabel), with a capital first letter, then the declared zone of a
+// Deploy, then its controller, and the source asset of anything but a Deploy
+// or a Maneuver. An item with no Label, and an ability with no card, print
+// their Kind.
+func stackItemLine(item view.StackView) string {
 	if item.Label == "" || item.Kind == "Ability" && item.Card == "" {
 		return fmt.Sprintf("%s (Player %d, Source %d)", item.Kind, item.Controller, item.Source)
 	}
@@ -109,6 +119,38 @@ func stackLine(item view.StackView) string {
 	return fmt.Sprintf("%s (Player %d, Source %d)", label, item.Controller, item.Source)
 }
 
+// chosenTargetsText names item's chosen targets that are still where they were
+// chosen: the slots in order, separated by "; ", and each slot's picks in pick
+// order, separated by ", ". An asset is named as the view shows it (assetRef;
+// a face-down asset by the Scratch face), and a pile slot's cards are followed
+// by " (discard pile)". A chosen asset no longer on the battlefield is left
+// out, and so is a slot with nothing left. It is "" when nothing is left.
+func chosenTargetsText(v view.View, item view.StackView) string {
+	var slots []string
+	for i := 0; i < max(len(item.Chosen), len(item.ChosenCards)); i++ {
+		var picks []string
+		if i < len(item.Chosen) {
+			for _, id := range item.Chosen[i] {
+				if a := v.Asset(id); a != nil {
+					picks = append(picks, assetRef(*a))
+				}
+			}
+		}
+		if len(picks) > 0 {
+			slots = append(slots, strings.Join(picks, ", "))
+		} else if i < len(item.ChosenCards) && len(item.ChosenCards[i]) > 0 {
+			slots = append(slots, strings.Join(item.ChosenCards[i], ", ")+" (discard pile)")
+		}
+	}
+	return strings.Join(slots, "; ")
+}
+
+// assetRef is how the screen names a battlefield asset: its card as the view
+// shows it, and its ID.
+func assetRef(a view.AssetView) string {
+	return fmt.Sprintf("%s#%d", a.Card, a.ID)
+}
+
 func formatAsset(a view.AssetView) string {
 	status := "Ready"
 	if a.Exhausted {
@@ -125,38 +167,13 @@ func formatAsset(a view.AssetView) string {
 	if len(extra) > 0 {
 		statStr = ", " + strings.Join(extra, ", ")
 	}
-	return fmt.Sprintf("%s#%d (%d/%d, %s, Yield R%d T%d%s)", a.Card, a.ID, a.Power, a.MaxHealth, status, a.ReqYield, a.TacYield, statStr)
+	return fmt.Sprintf("%s (%d/%d, %s, Yield R%d T%d%s)", assetRef(a), a.Power, a.MaxHealth, status, a.ReqYield, a.TacYield, statStr)
 }
 
 func actionLabel(v view.View, a engine.Action) string {
-	lookupAsset := func(id int) *view.AssetView {
-		if v.You.HQ != nil && v.You.HQ.ID == id {
-			return v.You.HQ
-		}
-		for i := range v.You.Reserves {
-			if v.You.Reserves[i].ID == id {
-				return &v.You.Reserves[i]
-			}
-		}
-		if v.Opponent.HQ != nil && v.Opponent.HQ.ID == id {
-			return v.Opponent.HQ
-		}
-		for i := range v.Opponent.Reserves {
-			if v.Opponent.Reserves[i].ID == id {
-				return &v.Opponent.Reserves[i]
-			}
-		}
-		for i := range v.Theater {
-			if v.Theater[i].ID == id {
-				return &v.Theater[i]
-			}
-		}
-		return nil
-	}
-
 	findAsset := func(id int) string {
-		if as := lookupAsset(id); as != nil {
-			return fmt.Sprintf("%s#%d", as.Card, id)
+		if as := v.Asset(id); as != nil {
+			return assetRef(*as)
 		}
 		return fmt.Sprintf("Asset#%d", id)
 	}
@@ -183,7 +200,7 @@ func actionLabel(v view.View, a engine.Action) string {
 	case engine.ActEnact:
 		return fmt.Sprintf("Enact: %s", findCard(v.You.Hand, a.Card))
 	case engine.ActActivate:
-		if as := lookupAsset(a.Asset); as != nil {
+		if as := v.Asset(a.Asset); as != nil {
 			return fmt.Sprintf("Activate: %s: %s", findAsset(a.Asset), as.Text)
 		}
 		return fmt.Sprintf("Activate: %s", findAsset(a.Asset))
```

### 3. `cmd/cardtown/play_targets_test.go` (new file, in the `test(T-0289)` commit)
```go
package main

import (
	"bytes"
	"strings"
	"testing"

	"animus-mundi/pkg/view"
)

// Tests for T-0289: the stack listing shows each item's chosen targets
// (directive #175, acceptance cases (a) to (g)).

// targetsView is a view with Ripper Pack#7 and Devilhorn Captain#12 in the Main
// Theater, a face-down asset #9 in the opponent's Reserves, and stack.
func targetsView(stack ...view.StackView) view.View {
	return view.View{
		Theater: []view.AssetView{
			{ID: 7, Card: "Ripper Pack", Controller: 0, Zone: "Main Theater"},
			{ID: 12, Card: "Devilhorn Captain", Controller: 1, Zone: "Main Theater"},
		},
		Opponent: view.Side{Reserves: []view.AssetView{
			{ID: 9, Card: "Scratch Company", Controller: 1, Zone: "Reserves", FaceDown: true},
		}},
		Stack: stack,
	}
}

// stackSection checks that printScreen's output for v holds want.
func stackSection(t *testing.T, v view.View, want string) {
	t.Helper()
	var buf bytes.Buffer
	printScreen(&buf, v, nil)
	if !strings.Contains(buf.String(), want) {
		t.Errorf("stack section:\n%s\nwant:\n%s", buf.String(), want)
	}
}

func TestPlay_StackLineShowsChosenAsset(t *testing.T) {
	v := targetsView(view.StackView{Kind: "Ability", Controller: 0, Source: 5, Card: "Devilhorn Captain",
		Label: "ability of Devilhorn Captain", Chosen: [][]int{{7}}, ChosenCards: [][]string{{}}})
	stackSection(t, v, "Resolution Stack:\n  Ability of Devilhorn Captain (Player 0, Source 5) -> Ripper Pack#7\n")
}

func TestPlay_StackLineSeparatesSlotsAndPicks(t *testing.T) {
	v := targetsView(
		view.StackView{Kind: "Ability", Controller: 0, Source: 5, Card: "Devilhorn Captain",
			Label: "ability of Devilhorn Captain", Chosen: [][]int{{7}, {12}}, ChosenCards: [][]string{{}, {}}},
		view.StackView{Kind: "Ability", Controller: 0, Source: 5, Card: "Devilhorn Captain",
			Label: "ability of Devilhorn Captain", Chosen: [][]int{{12, 7}}, ChosenCards: [][]string{{}}},
	)
	stackSection(t, v, "Resolution Stack:\n"+
		"  Ability of Devilhorn Captain (Player 0, Source 5) -> Ripper Pack#7; Devilhorn Captain#12\n"+
		"  Ability of Devilhorn Captain (Player 0, Source 5) -> Devilhorn Captain#12, Ripper Pack#7\n")
}

func TestPlay_StackLineShowsChosenPileCards(t *testing.T) {
	v := targetsView(
		view.StackView{Kind: "Ability", Controller: 0, Source: 5, Card: "Devilhorn Captain",
			Label: "ability of Devilhorn Captain", Chosen: [][]int{{}}, ChosenCards: [][]string{{"Eerie Remnant"}}},
		view.StackView{Kind: "Ability", Controller: 0, Source: 5, Card: "Devilhorn Captain",
			Label: "ability of Devilhorn Captain", Chosen: [][]int{{}, {7}},
			ChosenCards: [][]string{{"Eerie Remnant", "Ripper Pack"}, {}}},
	)
	stackSection(t, v, "Resolution Stack:\n"+
		"  Ability of Devilhorn Captain (Player 0, Source 5) -> Eerie Remnant (discard pile)\n"+
		"  Ability of Devilhorn Captain (Player 0, Source 5) -> Eerie Remnant, Ripper Pack (discard pile); Ripper Pack#7\n")
}

func TestPlay_StackLineLeavesOutGoneTargets(t *testing.T) {
	v := targetsView(
		view.StackView{Kind: "Ability", Controller: 0, Source: 5, Card: "Devilhorn Captain",
			Label: "ability of Devilhorn Captain", Chosen: [][]int{{40}, {41, 12}}, ChosenCards: [][]string{{}, {}}},
		view.StackView{Kind: "Ability", Controller: 0, Source: 5, Card: "Devilhorn Captain",
			Label: "ability of Devilhorn Captain", Chosen: [][]int{{40}, {}}, ChosenCards: [][]string{{}, {}}},
	)
	stackSection(t, v, "Resolution Stack:\n"+
		"  Ability of Devilhorn Captain (Player 0, Source 5) -> Devilhorn Captain#12\n"+
		"  Ability of Devilhorn Captain (Player 0, Source 5)\n")
}

func TestPlay_StackLineShowsFaceDownTargetAsViewed(t *testing.T) {
	v := targetsView(view.StackView{Kind: "Ability", Controller: 0, Source: 5, Card: "Devilhorn Captain",
		Label: "ability of Devilhorn Captain", Chosen: [][]int{{9}}, ChosenCards: [][]string{{}}})
	stackSection(t, v, "Resolution Stack:\n  Ability of Devilhorn Captain (Player 0, Source 5) -> Scratch Company#9\n")
}

func TestPlay_StackLineWithoutTargetsIsUnchanged(t *testing.T) {
	v := targetsView(
		view.StackView{Kind: "Ability", Controller: 0, Source: 5, Card: "Devilhorn Captain",
			Label: "ability of Devilhorn Captain", Chosen: [][]int{}, ChosenCards: [][]string{}},
		view.StackView{Kind: "Ability", Controller: 0, Source: 5, Card: "Devilhorn Captain",
			Label: "ability of Devilhorn Captain", Chosen: [][]int{{}}, ChosenCards: [][]string{{}}},
		view.StackView{Kind: "Deploy", Controller: 0, Card: "Ripper Pack", Zone: "Reserves", Label: "Deploy of Ripper Pack"},
		view.StackView{Kind: "March", Controller: 0, Source: 7, Zone: "Main Theater", Label: "March"},
		view.StackView{Kind: "Attack", Controller: 0, Source: 7, Target: 12, Label: "Attack"},
		view.StackView{Kind: "Scratch", Controller: 1, Source: 12, Label: "Scratch"},
	)
	stackSection(t, v, "Resolution Stack:\n"+
		"  Ability of Devilhorn Captain (Player 0, Source 5)\n"+
		"  Ability of Devilhorn Captain (Player 0, Source 5)\n"+
		"  Deploy of Ripper Pack to Reserves (Player 0)\n"+
		"  March (Player 0, Source 7)\n"+
		"  Attack (Player 0, Source 7)\n"+
		"  Scratch (Player 1, Source 12)\n")
}

func TestPlay_StackLineShowsManeuverTargets(t *testing.T) {
	v := targetsView(view.StackView{Kind: "Maneuver", Controller: 1, Card: "Alpha Volley",
		Label: "Maneuver Alpha Volley", Chosen: [][]int{{7}, {}}, ChosenCards: [][]string{{}, {"Eerie Remnant"}}})
	stackSection(t, v, "Resolution Stack:\n  Maneuver Alpha Volley (Player 1) -> Ripper Pack#7; Eerie Remnant (discard pile)\n")
}
```

## Acceptance Cases (harness tasks only; written by the lead)
Each of #175's cases (a)-(g) is one test in Edit 3: (a) `TestPlay_StackLineShowsChosenAsset`, (b) `TestPlay_StackLineSeparatesSlotsAndPicks`, (c) `TestPlay_StackLineShowsChosenPileCards`, (d) `TestPlay_StackLineLeavesOutGoneTargets`, (e) `TestPlay_StackLineShowsFaceDownTargetAsViewed`, (f) `TestPlay_StackLineWithoutTargetsIsUnchanged`, (g) `TestPlay_StackLineShowsManeuverTargets`, each line as the Design above words it.

W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **The tests fail first:** with Edit 3 applied and Edits 1-2 not yet, `go -C W test ./cmd/cardtown -run TestPlay_StackLine -v` fails in exactly six tests, every one but `TestPlay_StackLineWithoutTargetsIsUnchanged`, which passes (case (f) pins that lines without targets don't change, so it passes before and after).
2. **They pass after:** with Edits 1-2, the same command passes all seven, and `go -C W test ./cmd/cardtown -run TestPlay_ -v` passes.
3. **One asset lookup:** `git -C W grep -n "lookupAsset" -- pkg cmd` prints nothing, and `git -C W grep -n "Reserves.i..ID == id" -- pkg cmd ":!*_test.go"` prints exactly one line, `pkg/view/view.go:48` (the test helper `pkg/view/view_test.go:1037` is setup and stays).
4. **One way to name an asset on the screen:** `git -C W grep -n "#%d" -- cmd/cardtown/play_screen.go` prints exactly three lines: `cmd/cardtown/play_screen.go:151` (`assetRef`), `cmd/cardtown/play_screen.go:178` (`Asset#%d`, the fallback for an ID not on the battlefield) and `cmd/cardtown/play_screen.go:187` (`Card#%d`, `findCard`'s fallback).
5. **Only the new test file:** `git -C W diff autoloop/m2-wave-h.2 --numstat -- "*_test.go"` prints exactly one line, `112	0	cmd/cardtown/play_targets_test.go`, and it is the only file in the `test(T-0289)` commit.
6. **Only the listed edits:** `git -C W diff autoloop/m2-wave-h.2 --numstat -- pkg cmd docs` prints exactly three lines: `53	36	cmd/cardtown/play_screen.go`, `112	0	cmd/cardtown/play_targets_test.go` and `22	0	pkg/view/view.go`.
7. **Nothing else changed:** `git -C W diff autoloop/m2-wave-h.2 --stat` lists only the three files of case 6 and this task file.
8. **Sims byte-identical:** `go -C W tool task sim-compare -- autoloop/m2-wave-h.2` prints `[PASS]`.
9. `gofmt -l W/pkg W/cmd` prints nothing; `go -C W tool task check` (which runs the `pkg/view/viewtest` leak tests) and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/view/view.go` (Edit 1)
- `cmd/cardtown/play_screen.go` (Edit 2)
- `cmd/cardtown/play_targets_test.go` (Edit 3, new, in the `test(T-0289)` commit)

## Tests
- New, in the `test(T-0289)` commit: the seven `TestPlay_StackLine*` tests in `cmd/cardtown/play_targets_test.go` (Edit 3), one per #175 case.
- No existing test is edited. Every other test passes unedited.

## Done When
- [ ] Cases 1-9 hold, with their commands and results in the Log
- [ ] `go tool task sim-compare -- autoloop/m2-wave-h.2` passes (starter and stress sims byte-identical to the run branch)
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-19, lead-architect (wave H.2 run, session 9): written against b33c338 for #175; T-0288 is `done`, so #174 is `applied`.
  - `main` (1859fe4) has nothing the run branch lacks, so no sync. Not `WAVE-COMPLETE` before the task: #175 and #176 are `open` and this wave's (the run file's done-when names #169 to #176). No `open: before` directive is left.
  - This is a harness task, so it skips `tests` and goes straight to `ready`. Tier is `flash`, because the edits are exact and were dry-run.
  - Every command here is on the headless allow-list: one per line, no pipes, no `$`, and patterns with special characters are quoted. Case 3's pattern uses `.` for the brackets so it holds no `[` or `]`.
  - The arrow is `->` where the directive's examples write `→`, for the reason the Design gives; the user reads it in this row and in the run file.
  - #176 (the docs say what wave H.2 built) is next and last; it does not go into this task.
  - Waiting directives checked: #6, #7 (mode `serial`), #15 (2) (no draw or end-of-turn hook), #71, #91 (2), #106, #120, #142 (wave H.2 has no cards). None has fired.
  - Next: `kernel-engineer` on Flash.
