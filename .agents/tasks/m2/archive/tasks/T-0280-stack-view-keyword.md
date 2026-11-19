---
id: T-0280
title: "`view.StackView.Keyword` goes (directive #164; wave H.1 harness, flash, pure removal)"
kind: harness
tier: flash
wave: H.1 (board directive #164)
status: done
branch: m2/autoloop-wave-h.1/T-0280-stack-view-keyword
depends_on: [T-0279]
claims: [field:view.StackView.Keyword]
cards: []
requirements: []
---

## Goal
`view.StackView.Keyword` ([view.go:95](../../../pkg/view/view.go)) has no reader. Since #153 (T-0239) the play screen prints `StackView.Label`, the engine's own wording (`Game.StackLabel`), which already names the keyword ("Aggressive trigger of X"). So `Keyword` is only set ([view.go:207](../../../pkg/view/view.go)) and read by tests. The field and its assignment go, and every test that reads or sets it is updated in the same task (ADR 0007). Match output is byte for byte the same. When this merges, #164 is `applied` and wave H.1's directives are all applied.

## Inputs
- Board directive #164 (structural review 13, finding 6), approved by the user 2026-11-18, `open: before the next Deliverables task`; design note Deliverable 6n ([m2-ability-model.md:224](../../../docs/architecture/m2-ability-model.md)).
- Where it is today (at `autoloop/m2-wave-h.1` 717f70c): `git grep -n Keyword -- pkg/view cmd` finds the field and its assignment, and these readers or setters of `StackView.Keyword`, all in tests:
  - `TestView_AbilityItemShowsItsKeyword` ([view_test.go:714, 717](../../../pkg/view/view_test.go)) reads it.
  - `TestView_ManeuverItemShowsItsCardAndTargets` ([view_test.go:815-817](../../../pkg/view/view_test.go)) reads it.
  - `TestView_ActivatedAbilityItemShowsItsTargets` ([view_test.go:1017-1018](../../../pkg/view/view_test.go)) reads it.
  - `TestPlay_PrintScreenNamesEachStackItemsCard` ([play_source_test.go:121](../../../cmd/cardtown/play_source_test.go)) sets it in a `view.StackView` literal.

  The other `Keyword:` hits (`view_test.go:707`, `stack_label_test.go:23`) set `engine.StackItem.Keyword`, which stays. Nothing outside Go reads the field: `StackView` has no JSON tags, and no golden file, doc or client names a `"Keyword"` stack field (`git grep -n "\"Keyword\"" -- . ":!docs/tdd"` finds nothing).
- **The directive names one test edit; ADR 0007 needs four.** #164 authorizes only the `view_test.go:1017` edit, but deleting the field breaks the build of the other three tests above. Under ADR 0007 (a move updates every reference) and D2.4 (an edit to an existing test is allowed when the task names the test and the exact change), this task names all four. Lead decisions:
  - `TestView_AbilityItemShowsItsKeyword` keeps its name and its point, checked through `Label`, which is where the keyword now shows: the Aggressive item's `Label` is `Aggressive trigger of Alpha Lancer`, the printed item's is `ability of Alpha Sentry`. Its comment says so.
  - `TestView_ManeuverItemShowsItsCardAndTargets` drops its `Keyword` check; it checks the other empty fields still.
  - `TestView_ActivatedAbilityItemShowsItsTargets` drops `item.Keyword != ""` and the words `empty Keyword,` from its message, as #164 says.
  - `TestPlay_PrintScreenNamesEachStackItemsCard` drops `Keyword: "Aggressive", ` from its literal; its expected screen doesn't change, because `printScreen` reads `Label`.
- Docs: `engine-structs.md` covers `pkg/engine` only, and `pkg/view/doc.go` never named the field, so neither changes. The design note's Deliverable 6n line is the record of #164 and stays.
- Reuse search (D2.3): nothing is added.
- Size (D2.4): one non-test file.
- **Expected output difference:** none: `go tool task sim-compare` passes.
- The lead dry-ran these exact edits in a throwaway worktree of the run branch at 717f70c: `gofmt -l pkg cmd` clean, `go vet ./pkg/view/ ./cmd/cardtown/` clean, `go tool task check` exit 0, `go tool task sim-compare -- autoloop/m2-wave-h.1` `[PASS]`. Blobs: `cmd/cardtown/play_source_test.go` 44c59e7, `pkg/view/view.go` af64dc3, `pkg/view/view_test.go` 372d40a.

## Commits
1. `test(T-0280): stack view tests stop reading StackView.Keyword` with Edits 1 and 2 only. The tests still build and pass at this commit, because the field is still there (case 1).
2. `refactor(view): StackView.Keyword goes (T-0280)` with Edit 3.

No double quotes in either message: the allow-list refuses an escaped one.

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)
### 1. `pkg/view/view_test.go`: the authorized test edits (commit 1)
```diff
@@ -700,7 +700,7 @@ func TestView_Case5_LogLines(t *testing.T) {
 	}
 }
 
-// T-0074: an Aggressive trigger is an Ability item with a Keyword.
+// T-0074: an Aggressive trigger is an Ability item whose Label names its keyword.
 func TestView_AbilityItemShowsItsKeyword(t *testing.T) {
 	s := newContentScene(t)
 	s.g.Stack = []*engine.StackItem{
@@ -711,10 +711,10 @@ func TestView_AbilityItemShowsItsKeyword(t *testing.T) {
 	if len(v.Stack) != 2 {
 		t.Fatalf("stack = %+v", v.Stack)
 	}
-	if it := v.Stack[0]; it.Kind != "Ability" || it.Keyword != "Aggressive" || it.Card != "Alpha Lancer" {
+	if it := v.Stack[0]; it.Kind != "Ability" || it.Label != "Aggressive trigger of Alpha Lancer" || it.Card != "Alpha Lancer" {
 		t.Errorf("Aggressive item = %+v", it)
 	}
-	if it := v.Stack[1]; it.Kind != "Ability" || it.Keyword != "" {
+	if it := v.Stack[1]; it.Kind != "Ability" || it.Label != "ability of Alpha Sentry" {
 		t.Errorf("printed item = %+v", it)
 	}
 }
@@ -812,9 +812,6 @@ func TestView_ManeuverItemShowsItsCardAndTargets(t *testing.T) {
 		if !reflect.DeepEqual(it.ChosenCards, [][]string{{}}) {
 			t.Errorf("seat %d: ChosenCards = %v, want [[]]", seat, it.ChosenCards)
 		}
-		if it.Keyword != "" {
-			t.Errorf("seat %d: Keyword = %q, want empty", seat, it.Keyword)
-		}
 		if it.Zone != "" {
 			t.Errorf("seat %d: Zone = %q, want empty", seat, it.Zone)
 		}
@@ -1014,8 +1011,8 @@ func TestView_ActivatedAbilityItemShowsItsTargets(t *testing.T) {
 		if item == nil {
 			t.Fatalf("seat %d: no activated item on the stack view: %+v", seat, v.Stack)
 		}
-		if item.Card != "Tapper" || item.Keyword != "" || !reflect.DeepEqual(item.Chosen, [][]int{{picked}}) {
-			t.Errorf("seat %d: item %+v, want Card Tapper, empty Keyword, Chosen [[%d]]", seat, *item, picked)
+		if item.Card != "Tapper" || !reflect.DeepEqual(item.Chosen, [][]int{{picked}}) {
+			t.Errorf("seat %d: item %+v, want Card Tapper, Chosen [[%d]]", seat, *item, picked)
 		}
 	}
 }
```

### 2. `cmd/cardtown/play_source_test.go`: the authorized test edit (commit 1)
```diff
@@ -118,7 +118,7 @@ func TestPlay_ScriptedSessionPrintsSourceLines(t *testing.T) {
 func TestPlay_PrintScreenNamesEachStackItemsCard(t *testing.T) {
 	v := view.View{Stack: []view.StackView{
 		{Kind: "Ability", Controller: 1, Source: 12, Card: "Devilhorn Captain", Label: "ability of Devilhorn Captain"},
-		{Kind: "Ability", Controller: 0, Source: 7, Card: "Ripper Pack", Keyword: "Aggressive", Label: "Aggressive trigger of Ripper Pack"},
+		{Kind: "Ability", Controller: 0, Source: 7, Card: "Ripper Pack", Label: "Aggressive trigger of Ripper Pack"},
 		{Kind: "Ability", Controller: 0, Source: 5},
 		{Kind: "Deploy", Controller: 0, Card: "Chainbreaker Irregular", Zone: "Reserves"},
 		{Kind: "Maneuver", Controller: 0, Card: "Disorienting Pulse"},
```

### 3. `pkg/view/view.go`: the field and its assignment go (commit 2)
```diff
@@ -92,7 +92,6 @@ type StackView struct {
 	Card       string  // Deploy: the card on the stack; Ability: its source card; Maneuver: the card on the stack
 	Zone       string  // Deploy: the declared zone; March: the origin zone
 	Target     int     // Attack: the primary defender (asset ID)
-	Keyword    string  // Ability: the keyword its trigger came from ("Aggressive"); empty for a printed ability
 	Label      string  // what the item is, as the engine's place event words it (Game.StackLabel): "ability of X", "Aggressive trigger of X", "Deploy of X", "March", ...
 	Chosen     [][]int // Ability, Maneuver: per target slot, the chosen asset IDs; empty for a pile slot (see ChosenCards)
 	// ChosenCards: Ability, Maneuver: per target slot, for a pile slot, the names of its chosen cards that are still in a discard pile, in pick order; empty for an asset slot
@@ -204,7 +203,6 @@ func stackView(g *engine.Game, it *engine.StackItem) StackView {
 	case engine.StackAttack:
 		s.Target = it.Target
 	case engine.StackAbility:
-		s.Keyword = it.Keyword.String()
 		// A face-down source names only the Scratch face (Game.SourceCard).
 		if c := g.SourceCard(it); c != nil {
 			s.Card = c.Name
```

## Acceptance Cases (harness tasks only; written by the lead)
W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **The edited tests pass before the removal:** at the `test(T-0280)` commit, `go -C W test ./pkg/view/ ./cmd/cardtown/` passes. No test reads or sets `StackView.Keyword` any more, so the field can go without a test change after this commit.
2. **The field is gone:** `git -C W grep -n "s.Keyword" -- pkg/view` prints nothing, and `git -C W grep -n "Keyword    string" -- pkg/view` prints nothing.
3. **No test reads it:** `git -C W grep -n "it.Keyword" -- pkg/view cmd` prints nothing, and `git -C W grep -n "item.Keyword" -- pkg/view cmd` prints nothing.
4. **The same tests pass, with only the authorized edits:** `go -C W test ./pkg/... ./cmd/...` passes, and `git -C W diff autoloop/m2-wave-h.1 --numstat -- "*_test.go"` prints exactly two lines, `1	1	cmd/cardtown/play_source_test.go` and `5	8	pkg/view/view_test.go`.
5. **Sims byte-identical:** `go -C W tool task sim-compare -- autoloop/m2-wave-h.1` prints `[PASS]`.
6. **Only the listed edits:** `git -C W diff autoloop/m2-wave-h.1 --numstat -- pkg cmd docs` prints exactly three lines: `1	1	cmd/cardtown/play_source_test.go`, `0	2	pkg/view/view.go` and `5	8	pkg/view/view_test.go`.
7. **Nothing else changed:** `git -C W diff autoloop/m2-wave-h.1 --stat` lists only the three files of case 6 and this task file.
8. `gofmt -l W/pkg W/cmd` prints nothing; `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/view/view.go` (Edit 3)

## Tests
- `pkg/view/view_test.go`: the authorized edits to `TestView_AbilityItemShowsItsKeyword`, `TestView_ManeuverItemShowsItsCardAndTargets` and `TestView_ActivatedAbilityItemShowsItsTargets` (Edit 1), in the `test(T-0280)` commit.
- `cmd/cardtown/play_source_test.go`: the authorized edit to `TestPlay_PrintScreenNamesEachStackItemsCard` (Edit 2), in the `test(T-0280)` commit.
- No other test file changes.

## Done When
- [x] Cases 1-8 hold, with their commands and results in the Log
- [x] `go tool task sim-compare -- autoloop/m2-wave-h.1` passes (starter and stress sims byte-identical to the run branch)
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-19, lead-architect (wave H.1 run, session 9): written in full against 717f70c for #164, the last wave H.1 directive. `main` (7a8e4f5) has nothing the run branch lacks, so no sync. Not `WAVE-COMPLETE` before the task: #164 is `open: before the next Deliverables task`. #164 authorizes one test edit, but three more tests read or set the field, so this task names all four (ADR 0007, D2.4; see Inputs). This is a harness task, so it skips `tests` and goes straight to `ready`. Tier is `flash`, because the edits are exact and were dry-run. Every command here is on the headless allow-list: one per line, no pipes, no `$`, and patterns with special characters are quoted. Waiting directives checked: #6, #7 (mode `serial`), #15 (2) (no draw or end-of-turn hook), #71, #91 (2), #106, #120, #142 (wave H.1 has no cards): none has fired. Next: `kernel-engineer` on Flash.
- 2026-11-19, kernel-engineer: implemented T-0280 (directive #164).
  - Status moved to in-progress.
  - Applied Edits 1 & 2 to pkg/view/view_test.go (TestView_AbilityItemShowsItsKeyword, TestView_ManeuverItemShowsItsCardAndTargets, TestView_ActivatedAbilityItemShowsItsTargets) and cmd/cardtown/play_source_test.go (TestPlay_PrintScreenNamesEachStackItemsCard). Committed as `test(T-0280): stack view tests stop reading StackView.Keyword` (e10c41ac).
  - Applied Edit 3 removing `StackView.Keyword` and its assignment in `stackView()` in pkg/view/view.go. Committed as `refactor(view): StackView.Keyword goes (T-0280)` (813b456f).
  - Verification & Acceptance Cases:
    1. Pre-removal test check: `go -C W test ./pkg/view/ ./cmd/cardtown/` passed.
    2. Field removal grep: `git -C W grep -n "s.Keyword" -- pkg/view` and `git -C W grep -n "Keyword    string" -- pkg/view` returned empty (exit 1).
    3. Test readers grep: `git -C W grep -n "it.Keyword" -- pkg/view cmd` and `git -C W grep -n "item.Keyword" -- pkg/view cmd` returned empty (exit 1).
    4. Test diff check: `go -C W test ./pkg/... ./cmd/...` passed; `git -C W diff autoloop/m2-wave-h.1 --numstat -- "*_test.go"` output exactly 2 lines:
       1	1	cmd/cardtown/play_source_test.go
       5	8	pkg/view/view_test.go
    5. Sim compare: `go -C W tool task sim-compare -- autoloop/m2-wave-h.1` returned `[PASS] starter and stress sims match autoloop/m2-wave-h.1 byte for byte, exit codes included`.
    6. Code diff: `git -C W diff autoloop/m2-wave-h.1 --numstat -- pkg cmd docs` output exactly 3 lines:
       1	1	cmd/cardtown/play_source_test.go
       0	2	pkg/view/view.go
       5	8	pkg/view/view_test.go
    7. Overall stat: `git -C W diff autoloop/m2-wave-h.1 --stat` listed only the 3 files above plus this task file.
    8. Code formatting and gates: `gofmt -l W/pkg W/cmd` produced no output; `go -C W tool task check` passed all linters and tests; `go -C W tool task sim` passed on starter decks (200 games, seeds 1..200).
  - Moved status to review.
- 2026-11-19, lead-architect (wave H.1 run, session 9, D5): **Approved, `done`.** Merged into `autoloop/m2-wave-h.1` as 26f9748. Flash run `aff3ed13-96ba-4fd0-993e-529f3d7b50ee`: exit 0, `status: SUCCESS`, no `denied_actions` key, and no step with `"status":"ERROR"` in its transcript. `.err` holds only agy's "root agent idle" notice. Review of `git diff autoloop/m2-wave-h.1...m2/autoloop-wave-h.1/T-0280-stack-view-keyword` (77e8724):
  - Scope: the one Allowed File, the four authorized test edits and this file. The `test(T-0280)` commit (e10c41a) holds only `cmd/cardtown/play_source_test.go` 44c59e7 and `pkg/view/view_test.go` 372d40a; the implementation commit (813b456) holds only `pkg/view/view.go` af64dc3. Every blob matches the lead's dry run. No test changed after the test commit.
  - Fences and placement: nothing outside `pkg/view` and its tests in `cmd/cardtown` changes; no declaration moves. Reuse and ADR 0007: `git grep -n it.Keyword -- pkg/view cmd`, `git grep -n item.Keyword -- pkg/view cmd` and `git grep -n "Keyword    string" -- pkg/view` find nothing; no alias or wrapper is added. `engine.StackItem.Keyword` stays, read by `Game.StackLabel`.
  - Re-run by the reviewer on the merge: `gofmt -l pkg cmd` clean; `go tool task check` exit 0; `go tool task sim-compare -- 4d51ad17` `[PASS]`; `go tool task sim` exit 0, 200 games, P0 58.5% / P1 39.5% / draws 2.0%, P0 50/50 and P1 50/50 main cards executable, no monitor violation.
  - #164 is `applied` (T-0280) and moves to the directive archive.
