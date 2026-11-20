---
id: T-0293
title: "The view's `StackView.Chosen` holds only the chosen assets still on the battlefield (directive #179; wave H.3 harness, flash)"
kind: harness
tier: flash
wave: H.3 (board directive #179)
status: done
branch: m2/autoloop-wave-h.3/T-0293-view-chosen-still-there
depends_on: [T-0292]
claims: [func:view.chosenTargets, field:view.StackView.Chosen]
cards: []
requirements: []
---

## Goal
The view's `chosenTargets` keeps an asset slot's chosen ID only while `g.Asset(id)` finds it on the battlefield, so `StackView.Chosen` means "still there", as `ChosenCards` already does for a pile slot. The rule "a chosen target is shown only while it is still where it was chosen" then lives in `pkg/view` alone, and a second client only formats what the view gives it. When this merges, #179 is `applied`, and wave H.3's directives (#177 to #179) are all `applied`.

## Inputs
- Board directive #179 (structural review 15, finding 2), approved by the user 2026-11-20, `open`; wave H.3, after #178 (T-0292, `done`).
- Where it is today (at `autoloop/m2-wave-h.3` bc0fea2):
  - `chosenTargets` ([view.go:200-216](../../../pkg/view/view.go)) drops a pile slot's cards that have left the discard pile (`g.DiscardCard`), but copies every asset slot's IDs through (`append(chosen[i], ids...)`).
  - `chosenTargetsText` ([play_screen.go:122-146](../../../cmd/cardtown/play_screen.go), #175) drops the IDs `v.Asset` doesn't find, which is the other half of the rule.
- `g.Asset` ([state.go:290-298](../../../pkg/engine/state.go)) searches `g.Assets()`: both HQs, both Reserves and the Main Theater. `view.View.Asset` searches the same three places, so the view now drops exactly the IDs the play screen dropped, and the screen's output can't change.
- Design (lead decisions):
  - `chosenTargets` gets a doc comment, and `StackView.Chosen`'s field comment says "still on the battlefield".
  - `chosenTargetsText` keeps its `v.Asset` lookup, now to name each asset as the view shows it (directive #179). It keeps the `a != nil` check: it can't name an asset it can't find, and `TestPlay_StackLineLeavesOutGoneTargets` ([play_targets_test.go:70](../../../cmd/cardtown/play_targets_test.go)) hands it a view holding IDs that view doesn't show, and must pass unedited. Its comment now says the view gives only what is still there, and the screen leaves out an asset ID the view doesn't show.
  - The new test goes in a new file, `pkg/view/chosen_test.go`, using `view_test.go`'s `newScene`, `put` and `testUnit`. The second slot (one pick, which left) pins that a slot with nothing left stays, as an empty list, so slot indexes still line up with `ChosenCards`.
  - Docs: `docs/architecture/m2-ability-model.md` already describes #179's end state (line 241); nothing else names `StackView.Chosen`.
- Every existing test that puts an asset ID in `Chosen` (`pkg/view/view_test.go`: `TestView_Case*` contentScene, `TestView_PileTargetShowsCardNamesNotInstanceIDs`, `TestView_ManeuverPileTargetShowsCardNamesNotInstanceIDs`, the line-788 and line-1014 tests) uses an asset still on the battlefield, so each passes unedited (dry run below).
- Size (D2.4): two non-test files.
- **Expected output difference:** none in the sims (`pkg/view` doesn't run there), and none on the play screen. `go tool task sim-compare` passes.
- The lead dry-ran these exact edits and this exact test in a throwaway worktree of the run branch at bc0fea2. The test failed before Edits 1-2 (`seat 0: Chosen = [[100005 100003] [100005]], want [[100003] []]`, the same for seat 1) and passed after. `gofmt -l pkg cmd` was clean, `go vet ./pkg/... ./cmd/...` clean, `go tool task check` exited 0, `go tool task sim-compare -- autoloop/m2-wave-h.3` printed `[PASS]`, and `go tool task sim` exited 0 (P0 58.5% / P1 39.5% / draws 2.0%). Blobs: `view.go` c657779, `play_screen.go` 5eac7bb, `chosen_test.go` c26c74a.

## Commits
1. `test(T-0293): the view's Chosen omits chosen assets that left (directive #179)` with the test file below only. Run case 1 on this commit, before Edits 1-2.
2. `fix(view): StackView.Chosen holds only chosen assets still on the battlefield (T-0293)` with Edits 1-2.

No double quotes in either message: the allow-list refuses an escaped one.

## Test file (copy exactly; a new file, tabs as shown)
`pkg/view/chosen_test.go`:
```go
package view_test

import (
	"reflect"
	"testing"

	"animus-mundi/pkg/engine"
	"animus-mundi/pkg/engine/ability"
	"animus-mundi/pkg/view"
)

// T-0293 (directive #179): an asset slot's Chosen holds only the chosen assets
// still on the battlefield, as a pile slot's ChosenCards holds only the cards
// still in a discard pile. A slot with nothing left stays, empty.
func TestStackView_ChosenOmitsAssetsThatLeft(t *testing.T) {
	s := newScene(t)
	src := s.put(0, engine.ZoneMainTheater, testUnit("Chooser", 1, 1, ability.Keywords{}))
	stays := s.put(1, engine.ZoneMainTheater, testUnit("Still Here", 1, 1, ability.Keywords{}))
	gone := s.put(1, engine.ZoneReserves, testUnit("Gone Away", 1, 1, ability.Keywords{}))
	s.g.Stack = []*engine.StackItem{{
		Kind: engine.StackAbility, Controller: 0, Source: src.ID, Inst: src.Inst,
		Chosen: [][]int{{gone.ID, stays.ID}, {gone.ID}},
	}}
	s.g.Players[1].Reserves = nil // gone leaves the battlefield before the view is taken

	for seat := engine.PlayerID(0); seat < 2; seat++ {
		v := view.PlayerView(s.g, seat)
		if len(v.Stack) != 1 {
			t.Fatalf("seat %d: stack = %+v", seat, v.Stack)
		}
		if got, want := v.Stack[0].Chosen, [][]int{{stays.ID}, {}}; !reflect.DeepEqual(got, want) {
			t.Errorf("seat %d: Chosen = %v, want %v", seat, got, want)
		}
	}
}
```

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)
### 1. `pkg/view/view.go`: `chosenTargets` keeps only the assets still on the battlefield
```diff
@@ -115,7 +115,7 @@ type StackView struct {
 	Zone       string  // Deploy: the declared zone; March: the origin zone
 	Target     int     // Attack: the primary defender (asset ID)
 	Label      string  // what the item is, as the engine's place event words it (Game.StackLabel): "ability of X", "Aggressive trigger of X", "Deploy of X", "March", ...
-	Chosen     [][]int // Ability, Maneuver: per target slot, the chosen asset IDs; empty for a pile slot (see ChosenCards)
+	Chosen     [][]int // Ability, Maneuver: per target slot, for an asset slot, the IDs of its chosen assets that are still on the battlefield, in pick order; empty for a pile slot (see ChosenCards)
 	// ChosenCards: Ability, Maneuver: per target slot, for a pile slot, the names of its chosen cards that are still in a discard pile, in pick order; empty for an asset slot
 	ChosenCards [][]string
 }
@@ -197,13 +197,21 @@ func assetViews(g *engine.Game, as []*engine.Asset) []AssetView {
 	return out
 }
 
+// chosenTargets is what it chose that is still where it was chosen, per
+// target slot and in pick order: an asset slot's assets still on the
+// battlefield (Chosen) and a pile slot's cards still in a discard pile
+// (ChosenCards). A slot with nothing left stays, empty.
 func chosenTargets(g *engine.Game, it *engine.StackItem) ([][]int, [][]string) {
 	chosen := make([][]int, len(it.Chosen))
 	chosenCards := make([][]string, len(it.Chosen))
 	for i, ids := range it.Chosen {
 		chosen[i], chosenCards[i] = []int{}, []string{}
 		if !it.PileSlot(i) {
-			chosen[i] = append(chosen[i], ids...)
+			for _, id := range ids {
+				if g.Asset(id) != nil {
+					chosen[i] = append(chosen[i], id)
+				}
+			}
 			continue
 		}
 		for _, id := range ids {
```

### 2. `cmd/cardtown/play_screen.go`: `chosenTargetsText`'s comment says the view does the dropping
```diff
@@ -119,11 +119,12 @@ func stackItemLine(item view.StackView) string {
 	return fmt.Sprintf("%s (Player %d, Source %d)", label, item.Controller, item.Source)
 }
 
-// chosenTargetsText names item's chosen targets that are still where they were
-// chosen: the slots in order, separated by "; ", and each slot's picks in pick
-// order, separated by ", ". An asset is named as the view shows it (assetRef;
-// a face-down asset by the Scratch face), and a pile slot's cards are followed
-// by " (discard pile)". A chosen asset no longer on the battlefield is left
+// chosenTargetsText names item's chosen targets, which the view gives only
+// while they are still where they were chosen (view.StackView.Chosen and
+// ChosenCards): the slots in order, separated by "; ", and each slot's picks
+// in pick order, separated by ", ". An asset is named as the view shows it
+// (assetRef; a face-down asset by the Scratch face), and a pile slot's cards
+// are followed by " (discard pile)". An asset ID the view doesn't show is left
 // out, and so is a slot with nothing left. It is "" when nothing is left.
 func chosenTargetsText(v view.View, item view.StackView) string {
 	var slots []string
```

## Acceptance Cases (harness tasks only; written by the lead)
W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **The test fails first:** on the `test(T-0293)` commit, before Edits 1-2, `go -C W test ./pkg/view/ -run TestStackView_ChosenOmitsAssetsThatLeft` fails, with `Chosen = [[` followed by two IDs, for seat 0 and for seat 1.
2. **Then it passes:** after Edits 1-2, the same command passes.
3. **The view checks the battlefield:** `git -C W grep -n "g.Asset(id)" -- pkg/view/view.go` prints exactly one line, `pkg/view/view.go:211`.
4. **The play screen keeps its lookup:** `git -C W grep -n "v.Asset(id)" -- cmd/cardtown/play_screen.go` prints exactly two lines, `cmd/cardtown/play_screen.go:135` and `:176`.
5. **Existing tests pass, unedited:** `go -C W test ./pkg/... ./cmd/...` passes (`TestPlay_*` and the `pkg/view/viewtest` leak tests included), and `git -C W diff autoloop/m2-wave-h.3 --numstat -- "*_test.go"` prints exactly one line, `35	0	pkg/view/chosen_test.go`.
6. **Sims byte-identical:** `go -C W tool task sim-compare -- autoloop/m2-wave-h.3` prints `[PASS]`.
7. **Only the listed edits:** `git -C W diff autoloop/m2-wave-h.3 --numstat -- pkg cmd docs` prints exactly three lines: `6	5	cmd/cardtown/play_screen.go`, `35	0	pkg/view/chosen_test.go` and `10	2	pkg/view/view.go`.
8. **Nothing else changed:** `git -C W diff autoloop/m2-wave-h.3 --stat` lists only the three files of case 7 and this task file.
9. `gofmt -l W/pkg W/cmd` prints nothing; `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/view/view.go` (Edit 1)
- `cmd/cardtown/play_screen.go` (Edit 2)
- `pkg/view/chosen_test.go` (new, the test file above; `test(T-0293)` commit only)

## Tests
- New: `TestStackView_ChosenOmitsAssetsThatLeft` in `pkg/view/chosen_test.go` (the file above, copied exactly).
- No existing test edited.

## Done When
- [x] Cases 1-9 hold, with their commands and results in the Log
- [x] `go tool task sim-compare -- autoloop/m2-wave-h.3` passes (starter and stress sims byte-identical to the run branch)
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-20, lead-architect (wave H.3 run, session 3): written against bc0fea2 for #179; T-0292 is `done`, so #178 is `applied`.
  - `main` (54c8db4) has nothing the run branch lacks, so no sync. Not `WAVE-COMPLETE` before the task: the run file's header puts #179 in the wave, and it is `open`.
  - This is a harness task, so it skips `tests` and goes straight to `ready`. Tier is `flash`, because the test and the edits are exact and were dry-run.
  - Every command here is on the headless allow-list: one per line, no pipes, no `$`, and patterns with special characters are quoted.
  - Waiting directives checked: #6, #7 (mode `serial`), #15 (2) (no draw or end-of-turn hook in `ability.Hook`), #71, #91 (2), #106, #120, #142 (wave H.3 has no cards). None has fired.
  - Next: `kernel-engineer` on Flash.
- 2026-11-20, kernel-engineer (T-0293): implemented test and Edits 1-2 in 122c2c0d and baf6764b; verified Cases 1-9.
  - Diff shows:
    - `pkg/view/chosen_test.go`: added `TestStackView_ChosenOmitsAssetsThatLeft` verifying that `StackView.Chosen` retains only chosen asset IDs still on the battlefield (and keeps empty slice for empty slots).
    - `pkg/view/view.go`: `chosenTargets` verifies `g.Asset(id) != nil` before adding `id` to `chosen[i]`; doc comments on `chosenTargets` and `StackView.Chosen` updated.
    - `cmd/cardtown/play_screen.go`: doc comment on `chosenTargetsText` updated to note the view filters out assets no longer on the battlefield.
  - Reused: `g.Asset(id)` from `pkg/engine/state.go`.
  - Acceptance Cases & verification results:
    - Case 1: `go -C W test ./pkg/view/ -run TestStackView_ChosenOmitsAssetsThatLeft` on commit 122c2c0d failed before Edits 1-2: `seat 0: Chosen = [[100005 100003] [100005]], want [[100003] []]` and `seat 1: Chosen = [[100005 100003] [100005]], want [[100003] []]`.
    - Case 2: `go -C W test ./pkg/view/ -run TestStackView_ChosenOmitsAssetsThatLeft` after Edits 1-2 passed (`ok animus-mundi/pkg/view 0.637s`).
    - Case 3: `git -C W grep -n "g.Asset(id)" -- pkg/view/view.go` -> `pkg/view/view.go:211:				if g.Asset(id) != nil {` (exactly 1 line).
    - Case 4: `git -C W grep -n "v.Asset(id)" -- cmd/cardtown/play_screen.go` -> `cmd/cardtown/play_screen.go:135` and `:176` (exactly 2 lines).
    - Case 5: `go -C W test ./pkg/... ./cmd/...` -> ok (exit 0); `git -C W diff autoloop/m2-wave-h.3 --numstat -- "*_test.go"` -> `35 0 pkg/view/chosen_test.go` (exactly 1 line).
    - Case 6: `go -C W tool task sim-compare -- autoloop/m2-wave-h.3` -> `[PASS] starter and stress sims match autoloop/m2-wave-h.3 byte for byte, exit codes included`.
    - Case 7: `git -C W diff autoloop/m2-wave-h.3 --numstat -- pkg cmd docs` -> `6 5 cmd/cardtown/play_screen.go`, `35 0 pkg/view/chosen_test.go`, `10 2 pkg/view/view.go` (exactly 3 lines).
    - Case 8: `git -C W diff autoloop/m2-wave-h.3 --stat` -> only the 3 code files and this task file.
    - Case 9: `gofmt -l W/pkg W/cmd` -> clean (no output); `go -C W tool task check` -> passed; `go -C W tool task sim` -> passed (P0 58.5% / P1 39.5% / draws 2.0%).
  - Handed off for review.
- 2026-11-20, lead-architect (wave H.3 run, session 3, D5): **Approved, `done`.** Merged into `autoloop/m2-wave-h.3` as 96b25b9. Flash run `5441530e-89f8-4a26-9cfe-22704dda1eee`: exit 0, `status: SUCCESS`, no `denied_actions` key, and no step with `"status":"ERROR"` in its transcript. `.err` holds only agy's "root agent idle" notice. Review of `git diff autoloop/m2-wave-h.3...m2/autoloop-wave-h.3/T-0293-view-chosen-still-there` (f169fc2):
  - Scope: the three Allowed Files and this file. The test file is alone in the `test(T-0293)` commit (122c2c0), and Edits 1-2 are in the implementation commit (baf6764); no existing test touched. All three blobs match the lead's dry run (`view.go` c657779, `play_screen.go` 5eac7bb, `chosen_test.go` c26c74a).
  - Fences and reuse: the rule "a chosen target is shown only while it is still where it was chosen" now lives in `chosenTargets` for both slot kinds; the play screen only names what the view gives it. No new declaration; `g.Asset` reused.
  - Re-run by the reviewer in a review worktree: case 1 on 122c2c0 (fails, `Chosen = [[100005 100003] [100005]], want [[100003] []]` for both seats), cases 2-5 and 7 on the task branch as listed; `gofmt -l pkg cmd` clean; `go tool task check` exit 0; `go tool task sim-compare -- autoloop/m2-wave-h.3` `[PASS]`; `go tool task sim` exit 0, P0 58.5% / P1 39.5% / draws 2.0%, P0 50/50 and P1 50/50 main cards executable, no monitor violation.
  - #179 is `applied` and archived. Wave H.3's directives #177 to #179 are all `applied`.
