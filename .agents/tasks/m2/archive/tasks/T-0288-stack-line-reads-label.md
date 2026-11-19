---
id: T-0288
title: "The play screen's stack listing prints `StackLabel` for every item (directive #174; wave H.2 harness, flash)"
kind: harness
tier: flash
wave: H.2 (board directive #174)
status: done
branch: m2/autoloop-wave-h.2/T-0288-stack-line-reads-label
depends_on: [T-0287]
claims: [func:main.stackLine]
cards: []
requirements: []
---

## Goal
The play screen's stack listing words every stack item the way the engine does, through `view.StackView.Label` (`Game.StackLabel`, "the one wording of what a stack item is"), instead of writing its own "Deploy of X" and "Maneuver: X". After this task, the words "Deploy of" and "Maneuver X" are written only in `StackLabel`, and the play screen's Maneuver line reads "Maneuver X (Player N)". Every other line keeps its text. When this merges, #174 is `applied`.

## Inputs
- Board directive #174 (structural review 14, finding 6 (a)), approved by the user 2026-11-19, `open: before the next Deliverables task`; wave H.2, after #173 (T-0286 and T-0287, both `done`).
- Where it is today (at `autoloop/m2-wave-h.2` 16dccc8): `stackLine` ([play_screen.go:97-107](../../../cmd/cardtown/play_screen.go)) reads `Label` only for an ability with a card, and words a Deploy (`"Deploy of %s to %s (Player %d)"`) and a Maneuver (`"Maneuver: %s (Player %d)"`) itself. `Game.StackLabel` ([stack.go:84-106](../../../pkg/engine/stack.go)) already says "Deploy of X" and "Maneuver X", and `view` copies it into every `StackView.Label` ([view.go:197](../../../pkg/view/view.go)).
- Design (lead decisions, the directive's fix):
  - An item with an empty `Label`, or an ability with no card, prints `"<Kind> (Player N, Source S)"`, exactly today's fallback line. That check comes first, so the label is never sliced when empty.
  - Otherwise the line is the `Label` with its first letter upper-cased, then `" to <Zone>"` for a Deploy, then `" (Player N)"` for a Deploy or a Maneuver and `" (Player N, Source S)"` for anything else.
  - March, Attack and Scratch lines don't change: their labels equal their kinds. A Deploy's line doesn't change, because its label is "Deploy of X". An ability's line doesn't change.
  - `stackLine` still compares `Kind` against the view's kind names ("Deploy", "Maneuver", "Ability"), which the directive allows (`stackKinds`' kind names).
  - No engine or view change. The comment on `StackLabel` already names the play screen's stack listing as a reader.
- Reuse search (D2.3): `git grep -n "Deploy of" -- pkg cmd ":!*_test.go"` finds `stackLine`, `StackLabel` and `view.go`'s field comment; nothing else words a stack item.
- Size (D2.4): one non-test file, plus the two test edits the directive authorizes.
- **Expected output difference:** none in the sims, which print no play screen. `go tool task sim-compare` passes.
- The lead dry-ran these exact edits in a throwaway worktree of the run branch at 16dccc8. With Edits 2-3 and without Edit 1, `go test ./cmd/cardtown -run TestPlay_PrintScreen` failed in exactly `TestPlay_PrintScreenNamesEachStackItemsCard` and `TestPlay_PrintScreenManeuverStackLine`; with Edit 1 it passed. `gofmt -l pkg cmd` was clean, `go vet ./cmd/...` clean, `go tool task check` exited 0, and `go tool task sim-compare -- autoloop/m2-wave-h.2` printed `[PASS]`. Blobs: `play_screen.go` 7f78b0b, `play_test.go` f65b76d, `play_source_test.go` 8feb636.

## Commits
1. `test(T-0288): the stack listing's Maneuver line reads its Label` with Edits 2-3 only. Run case 1 before committing it and record the failure.
2. `refactor(cardtown): the play screen's stack listing prints StackLabel for every item (T-0288)` with Edit 1.
No double quotes in either message: the allow-list refuses an escaped one.

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)
### 1. `cmd/cardtown/play_screen.go`: `stackLine` prints every item's `Label`
```diff
@@ -90,20 +90,23 @@ func printScreen(out io.Writer, v view.View, newLogs []string) {
 	fmt.Fprintf(out, "Discards: You %d | Opponent %d\n\n", len(v.You.Discard), len(v.Opponent.Discard))
 }
 
-// stackLine is the stack listing's line for item. An ability prints its Label,
-// the engine's own wording (Game.StackLabel), with a capital first letter, a
-// Deploy its card and zone, and a Maneuver its card (#149); a March, an Attack,
-// a Scratch, and an ability with no card name their source asset.
+// stackLine is the stack listing's line for item: its Label, the engine's one
+// wording of what it is (Game.StackLabel), with a capital first letter, then
+// the declared zone of a Deploy, then its controller, and the source asset of
+// anything but a Deploy or a Maneuver. An item with no Label, and an ability
+// with no card, print their Kind.
 func stackLine(item view.StackView) string {
-	switch {
-	case item.Kind == "Maneuver":
-		return fmt.Sprintf("Maneuver: %s (Player %d)", item.Card, item.Controller)
-	case item.Kind == "Deploy" && item.Card != "":
-		return fmt.Sprintf("Deploy of %s to %s (Player %d)", item.Card, item.Zone, item.Controller)
-	case item.Kind == "Ability" && item.Card != "" && item.Label != "":
-		return fmt.Sprintf("%s%s (Player %d, Source %d)", strings.ToUpper(item.Label[:1]), item.Label[1:], item.Controller, item.Source)
-	}
-	return fmt.Sprintf("%s (Player %d, Source %d)", item.Kind, item.Controller, item.Source)
+	if item.Label == "" || item.Kind == "Ability" && item.Card == "" {
+		return fmt.Sprintf("%s (Player %d, Source %d)", item.Kind, item.Controller, item.Source)
+	}
+	label := strings.ToUpper(item.Label[:1]) + item.Label[1:]
+	switch item.Kind {
+	case "Deploy":
+		return fmt.Sprintf("%s to %s (Player %d)", label, item.Zone, item.Controller)
+	case "Maneuver":
+		return fmt.Sprintf("%s (Player %d)", label, item.Controller)
+	}
+	return fmt.Sprintf("%s (Player %d, Source %d)", label, item.Controller, item.Source)
 }
 
 func formatAsset(a view.AssetView) string {
```

### 2. `cmd/cardtown/play_test.go`: `TestPlay_PrintScreenManeuverStackLine` (authorized by #174)
```diff
@@ -283,7 +283,7 @@ func TestPlay_ActionLabelNamesTheEnactedCard(t *testing.T) {
 func TestPlay_PrintScreenManeuverStackLine(t *testing.T) {
 	v := view.View{
 		Stack: []view.StackView{
-			{Kind: "Maneuver", Controller: 0, Card: "Alpha Volley"},
+			{Kind: "Maneuver", Controller: 0, Card: "Alpha Volley", Label: "Maneuver Alpha Volley"},
 			{Kind: "Attack", Controller: 1, Source: 7, Target: 3},
 		},
 	}
@@ -291,7 +291,7 @@ func TestPlay_PrintScreenManeuverStackLine(t *testing.T) {
 	printScreen(&buf, v, nil)
 	out := buf.String()
 
-	expectedSection := "Resolution Stack:\n  Maneuver: Alpha Volley (Player 0)\n  Attack (Player 1, Source 7)\n"
+	expectedSection := "Resolution Stack:\n  Maneuver Alpha Volley (Player 0)\n  Attack (Player 1, Source 7)\n"
 	if !strings.Contains(out, expectedSection) {
 		t.Errorf("printScreen output did not contain expected stack section:\n%s\nGot output:\n%s", expectedSection, out)
 	}
```

### 3. `cmd/cardtown/play_source_test.go`: `TestPlay_PrintScreenNamesEachStackItemsCard` (authorized by #174)
```diff
@@ -120,8 +120,8 @@ func TestPlay_PrintScreenNamesEachStackItemsCard(t *testing.T) {
 		{Kind: "Ability", Controller: 1, Source: 12, Card: "Devilhorn Captain", Label: "ability of Devilhorn Captain"},
 		{Kind: "Ability", Controller: 0, Source: 7, Card: "Ripper Pack", Label: "Aggressive trigger of Ripper Pack"},
 		{Kind: "Ability", Controller: 0, Source: 5},
-		{Kind: "Deploy", Controller: 0, Card: "Chainbreaker Irregular", Zone: "Reserves"},
-		{Kind: "Maneuver", Controller: 0, Card: "Disorienting Pulse"},
+		{Kind: "Deploy", Controller: 0, Card: "Chainbreaker Irregular", Zone: "Reserves", Label: "Deploy of Chainbreaker Irregular"},
+		{Kind: "Maneuver", Controller: 0, Card: "Disorienting Pulse", Label: "Maneuver Disorienting Pulse"},
 		{Kind: "March", Controller: 1, Source: 9, Zone: "Reserves"},
 		{Kind: "Attack", Controller: 1, Source: 7, Target: 3},
 		{Kind: "Scratch", Controller: 0, Source: 1},
@@ -133,7 +133,7 @@ func TestPlay_PrintScreenNamesEachStackItemsCard(t *testing.T) {
 		"  Aggressive trigger of Ripper Pack (Player 0, Source 7)\n" +
 		"  Ability (Player 0, Source 5)\n" +
 		"  Deploy of Chainbreaker Irregular to Reserves (Player 0)\n" +
-		"  Maneuver: Disorienting Pulse (Player 0)\n" +
+		"  Maneuver Disorienting Pulse (Player 0)\n" +
 		"  March (Player 1, Source 9)\n" +
 		"  Attack (Player 1, Source 7)\n" +
 		"  Scratch (Player 0, Source 1)\n"
```

## Acceptance Cases (harness tasks only; written by the lead)
W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **The tests fail first:** with Edits 2-3 applied and Edit 1 not yet, `go -C W test ./cmd/cardtown -run TestPlay_PrintScreen -v` fails in exactly two tests, `TestPlay_PrintScreenNamesEachStackItemsCard` and `TestPlay_PrintScreenManeuverStackLine`; the other `TestPlay_PrintScreen*` tests pass.
2. **They pass after:** with Edit 1, the same command passes.
3. **One wording of a Deploy:** `git -C W grep -n "Deploy of" -- pkg cmd ":!*_test.go"` prints exactly three lines: `pkg/engine/stack.go:79` and `pkg/engine/stack.go:87` (in `StackLabel`) and `pkg/view/view.go:95` (the `Label` field's comment).
4. **One wording of a Maneuver:** `git -C W grep -n "Maneuver:" -- cmd/cardtown/play_screen.go` prints nothing, and `git -C W grep -n "Maneuver" -- cmd/cardtown/play_screen.go` prints exactly two lines, `cmd/cardtown/play_screen.go:96` (the doc comment) and `cmd/cardtown/play_screen.go:106` (`case "Maneuver":`, a kind name).
5. **Only the authorized test edits:** `git -C W diff autoloop/m2-wave-h.2 --numstat -- "*_test.go"` prints exactly two lines: `3	3	cmd/cardtown/play_source_test.go` and `2	2	cmd/cardtown/play_test.go`, and both are in the `test(T-0288)` commit.
6. **Only the listed edits:** `git -C W diff autoloop/m2-wave-h.2 --numstat -- pkg cmd docs` prints exactly three lines: `16	13	cmd/cardtown/play_screen.go`, `3	3	cmd/cardtown/play_source_test.go` and `2	2	cmd/cardtown/play_test.go`.
7. **Nothing else changed:** `git -C W diff autoloop/m2-wave-h.2 --stat` lists only the three files of case 6 and this task file.
8. **Sims byte-identical:** `go -C W tool task sim-compare -- autoloop/m2-wave-h.2` prints `[PASS]`.
9. `gofmt -l W/pkg W/cmd` prints nothing; `go -C W tool task check` (which runs the `pkg/view/viewtest` leak tests) and `go -C W tool task sim` pass.

## Allowed Files
- `cmd/cardtown/play_screen.go` (Edit 1)
- `cmd/cardtown/play_test.go` (Edit 2, in the `test(T-0288)` commit)
- `cmd/cardtown/play_source_test.go` (Edit 3, in the `test(T-0288)` commit)

## Tests
- Edited, as #174 authorizes, in the `test(T-0288)` commit:
  - `TestPlay_PrintScreenManeuverStackLine` ([play_test.go:283](../../../cmd/cardtown/play_test.go)): its Maneuver fixture gains `Label: "Maneuver Alpha Volley"`, and it expects `"  Maneuver Alpha Volley (Player 0)"` (Edit 2).
  - `TestPlay_PrintScreenNamesEachStackItemsCard` ([play_source_test.go:118](../../../cmd/cardtown/play_source_test.go)): its Deploy and Maneuver fixtures gain `Label` (`"Deploy of Chainbreaker Irregular"`, `"Maneuver Disorienting Pulse"`), and its Maneuver line expects `"  Maneuver Disorienting Pulse (Player 0)"` (Edit 3).
- No test is added. Every other test passes unedited.

## Done When
- [x] Cases 1-9 hold, with their commands and results in the Log
- [x] `go tool task sim-compare -- autoloop/m2-wave-h.2` passes (starter and stress sims byte-identical to the run branch)
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-19, lead-architect (wave H.2 run, session 8): written against 16dccc8 for #174; T-0287 is `done`, so #173 is `applied`.
  - `main` (1859fe4) has nothing the run branch lacks, so no sync. Not `WAVE-COMPLETE` before the task: #174 is `open: before the next Deliverables task`, and #175 and #176 are `open`.
  - This is a harness task, so it skips `tests` and goes straight to `ready`. Tier is `flash`, because the edits are exact and were dry-run.
  - Every command here is on the headless allow-list: one per line, no pipes, no `$`, and patterns with special characters are quoted.
  - #175 edits the same function next, and #176 then documents the wave; neither goes into this task.
  - Waiting directives checked: #6, #7 (mode `serial`), #15 (2) (no draw or end-of-turn hook), #71, #91 (2), #106, #120, #142 (wave H.2 has no cards). None has fired.
  - Next: `kernel-engineer` on Flash.
- 2026-11-19, kernel-engineer: implemented T-0288 per instructions.
  - Applied Edits 2-3 to cmd/cardtown/play_test.go and cmd/cardtown/play_source_test.go.
  - Ran Acceptance Case 1: `go test ./cmd/cardtown -run TestPlay_PrintScreen -v` failed in exactly `TestPlay_PrintScreenNamesEachStackItemsCard` and `TestPlay_PrintScreenManeuverStackLine` (the other three passed). Committed as `test(T-0288): the stack listing's Maneuver line reads its Label`.
  - Applied Edit 1 to cmd/cardtown/play_screen.go (stackLine reads item.Label).
  - Ran Acceptance Case 2: `go test ./cmd/cardtown -run TestPlay_PrintScreen -v` passed all 5 tests. Committed as `refactor(cardtown): the play screen's stack listing prints StackLabel for every item (T-0288)`.
  - Ran Acceptance Case 3: `git grep -n "Deploy of" -- pkg cmd ":!*_test.go"` printed exactly 3 lines in stack.go:79, stack.go:87, and view.go:95.
  - Ran Acceptance Case 4: `git grep -n "Maneuver:" -- cmd/cardtown/play_screen.go` printed nothing; `git grep -n "Maneuver" -- cmd/cardtown/play_screen.go` printed exactly 2 lines (play_screen.go:96 and play_screen.go:106).
  - Ran Acceptance Case 5: `git diff autoloop/m2-wave-h.2 --numstat -- "*_test.go"` printed 3 3 cmd/cardtown/play_source_test.go and 2 2 cmd/cardtown/play_test.go in the test commit.
  - Ran Acceptance Case 6: `git diff autoloop/m2-wave-h.2 --numstat -- pkg cmd docs` printed 16 13 cmd/cardtown/play_screen.go, 3 3 cmd/cardtown/play_source_test.go, and 2 2 cmd/cardtown/play_test.go.
  - Ran Acceptance Case 7: `git diff autoloop/m2-wave-h.2 --stat` lists only play_screen.go, play_source_test.go, play_test.go, and the task file.
  - Ran Acceptance Case 8: `go tool task sim-compare -- autoloop/m2-wave-h.2` passed with `[PASS] starter and stress sims match autoloop/m2-wave-h.2 byte for byte, exit codes included`.
  - Ran Acceptance Case 9: `gofmt -l pkg cmd` printed nothing; `go tool task check` passed (all packages ok, linter checks PASSED); `go tool task sim` passed.
- 2026-11-19, lead-architect (wave H.2 run, session 8, D5): **Approved, `done`.** Merged into `autoloop/m2-wave-h.2` as ef7ef19. Flash run `f2b75451-9c3f-44aa-a61b-726cfbdbfebe`: exit 0, `status: SUCCESS`, no `denied_actions` key, and no step with `"status":"ERROR"` in its transcript. `.err` holds only agy's "root agent idle" notice. Review of `git diff autoloop/m2-wave-h.2...m2/autoloop-wave-h.2/T-0288-stack-line-reads-label` (90c77db):
  - Scope: the three Allowed Files and this file. The `test(T-0288)` commit (238cb2b) holds only the two authorized test edits, and the implementation commit (f68037c) only `play_screen.go`. All three blobs match the lead's dry run (`play_screen.go` 7f78b0b, `play_test.go` f65b76d, `play_source_test.go` 8feb636).
  - Fences and reuse: `stackLine` prints `StackView.Label` for every item, so "Deploy of" and "Maneuver X" are written only in `StackLabel`; the play screen compares only kind names. No engine or view change, and nothing old survives (ADR 0007).
  - Re-run by the reviewer in a review worktree of the task branch: case 1 at 238cb2b (exactly the two named tests fail), cases 2-4 as listed; `gofmt -l pkg cmd` clean; `go tool task check` exit 0; `go tool task sim-compare -- autoloop/m2-wave-h.2` `[PASS]`; `go tool task sim` exit 0, P0 58.5% / P1 39.5% / draws 2.0%, P0 50/50 and P1 50/50 main cards executable, no monitor violation.
  - #174 is `applied` and archived. #175 (the stack listing shows chosen targets, the same function) is next.
