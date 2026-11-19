---
id: T-0285
title: "`cardtown lint` asks the linters what they read: `fences.Reads`, `rules.Reads` (directive #172 (b); wave H.2 harness, flash)"
kind: harness
tier: flash
wave: H.2 (board directive #172 (b))
status: ready
branch: m2/autoloop-wave-h.2/T-0285-linters-say-what-they-read
depends_on: [T-0284]
claims: [func:fences.Reads, func:rules.Reads]
cards: []
requirements: []
---

## Goal
`cmd/cardtown`'s `isFencesPath` and `isRulesPath` restate which files the fences and rules linters read ([lint.go:217-220](../../../cmd/cardtown/lint.go), [:283-290](../../../cmd/cardtown/lint.go)), so T-0274 had to widen both the fences linter and `isFencesPath`. After this task, `fences.Reads(path)` and `rules.Reads(path)` report whether a change to a repo-relative, slash-separated, lower-cased path can affect that linter. They are built from `aliasRoots`, `cardsheet.Dir`, `specdoc.FRDDir`, `specdoc.MatrixFile`, `specdoc.RulesFRD`, `rules.EngineDir` and the survey's location. `selectChanged` calls them, and `isFencesPath` and `isRulesPath` go, with no alias or wrapper ([ADR 0007](../../../docs/architecture/decisions/0007-moves-update-every-reference.md)). `scannedFiles` lists `goFilesUnder(repoRoot, "pkg")` minus its two skipped trees (`pkg/engine/cards`, `pkg/toolchain`), so the walk for non-test Go files is written once. When this merges after T-0284, #172 is `applied`.

## Inputs
- Board directive #172 (structural review 14, finding 4), approved by the user 2026-11-19, `open: before the next Deliverables task`. Part (a) is T-0284 (`done`, merge 7cca4ea). This task is (b).
- Where it is today (at `autoloop/m2-wave-h.2` d64b6fa):
  - `isRulesPath` ([lint.go:217-220](../../../cmd/cardtown/lint.go)) hard-codes `pkg/engine/` beside `rules.EngineDir`.
  - `isFencesPath` ([lint.go:283-290](../../../cmd/cardtown/lint.go)) hard-codes `pkg/`, `cmd/` and `tools/` beside `aliasRoots` ([aliases.go:13](../../../pkg/toolchain/linter/fences/aliases.go)). It also hard-codes `docs/design/` beside the fences linter's own `filepath.Join(repoRoot, "docs", "design", SurveyFile)` ([fences.go:205](../../../pkg/toolchain/linter/fences/fences.go)).
  - `scannedFiles` ([fences.go:245-273](../../../pkg/toolchain/linter/fences/fences.go)) and `goFilesUnder` ([fences.go:395-420](../../../pkg/toolchain/linter/fences/fences.go)) each walk for non-test Go files.
- Design (lead decisions):
  - Each `Reads` sits in its linter's main file (`fences.go`, `rules.go`), beside `Lint`. Each lower-cases the constants it compares with, because `selectChanged` passes a lower-cased path and `specdoc.RulesFRD` is mixed-case.
  - `rules.Reads` keeps what `isRulesPath` selected: FRD 01, or any Go file under `EngineDir`. `TestSelectChanged_Case1_Paths` expects `pkg/engine/x.go` to select the rules linter, and the directive's new cases include an engine Go file. `EngineDir` is an OS path (`filepath.Join`), so `Reads` compares `filepath.ToSlash(EngineDir)`.
  - `fences.Reads` keeps what `isFencesPath` selected: any Go file under an `aliasRoots` root (test files included; the moved cases keep `pkg/sim/x_test.go` true), a CSV under `cardsheet.Dir`, the keyword matrix and the survey.
  - The survey's directory becomes an unexported `surveyDir` const (`docs/design`) beside `SurveyFile`, read by `readMarkerIDs` and `Reads`. So `lint.go`'s `"docs/design/"` copy goes, and no Go file under `pkg/toolchain` or `cmd` writes the directory twice.
  - `scannedFiles` filters `goFilesUnder(repoRoot, "pkg")` by the prefixes `pkg/engine/cards/` and `pkg/toolchain/`. Both walks are `filepath.WalkDir`, which visits in lexical order, so the list is the same, in the same order. The lead compared the old and new lists on the repo: 45 files, `reflect.DeepEqual`. The only other difference is a repo with no `pkg` directory at all: `scannedFiles` now returns no files instead of an error, as `goFilesUnder` already does for `pkg/engine/cards`. The repo always has `pkg`.
- **Authorized test edits** (#172), in the `test(T-0285)` commit:
  - `TestIsFencesPath` ([lint_test.go:13-39](../../../cmd/cardtown/lint_test.go)) moves to `pkg/toolchain/linter/fences/fences_test.go` as `TestReads`, with the same fourteen cases, calling `fences.Reads` (Edits 1 and 2).
  - A new `TestReads` in `pkg/toolchain/linter/rules/rules_test.go` covers FRD 01, an engine Go file, an engine test file and a `pkg/sim` file (Edit 3).
  - `TestSelectChanged_*` stay as they are, and pass.
  At that commit the fences and rules test packages don't build (case 1).
- Reuse search (D2.3): nothing outside `cmd/cardtown/lint.go` reports which paths a linter reads (`git grep -n -e isFencesPath -e isRulesPath -e "func Reads"` finds only these two functions). `goFilesUnder` is the fences package's only other walk for Go files.
- Size (D2.4): three non-test files (`lint.go`, `fences.go`, `rules.go`), within the limit of four.
- Docs: no doc names `isFencesPath` or `isRulesPath` outside the board, this file and the run files. #176 (b) updates the docs for wave H.2 as a whole.
- **Expected output difference:** none. `cardtown lint --all` prints the same lines as on the run branch apart from the worktree path (the lead compared them byte for byte with the root path masked), and `go tool task sim-compare` passes.
- The lead dry-ran these exact edits in a throwaway worktree of the run branch at d64b6fa:
  - `gofmt -l pkg cmd` clean, `go vet ./pkg/toolchain/... ./cmd/...` clean, `go tool task check` exit 0, `go tool task sim-compare -- autoloop/m2-wave-h.2` `[PASS]`.
  - With only Edits 1-3 applied, `go vet ./pkg/toolchain/linter/fences/ ./pkg/toolchain/linter/rules/` fails with `undefined: fences.Reads` and `undefined: rules.Reads`.
  - Blobs: `lint_test.go` 85f479a, `fences_test.go` ee42b4c, `rules_test.go` 96c8495, `lint.go` de4b4df, `fences.go` d67cfe3, `rules.go` bade881.

## Commits
1. `test(T-0285): TestIsFencesPath moves to fences as TestReads, and rules gains TestReads` with Edits 1-3 only. At this commit the fences and rules test packages do not build (case 1).
2. `refactor(toolchain): cardtown lint asks the linters what they read, fences.Reads and rules.Reads (T-0285)` with Edits 4-6.

No double quotes in either message: the allow-list refuses an escaped one.

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)
### 1. `cmd/cardtown/lint_test.go` (commit 1: `TestIsFencesPath` leaves)
```diff
@@ -10,34 +10,6 @@ import (
 	"reflect"
 )
 
-// 37. isFencesPath covers Go files under pkg/, cmd/ and tools/, plus the data inputs.
-func TestIsFencesPath(t *testing.T) {
-	cases := []struct {
-		path string
-		want bool
-	}{
-		{"pkg/sim/x.go", true},
-		{"pkg/engine/x.go", true},
-		{"pkg/engine/cards/animus.go", true},
-		{"pkg/foo/bar/x.go", true},
-		{"pkg/sim/x_test.go", true},
-		{"data/cards/units.csv", true},
-		{"docs/frd/keyword_interaction_matrix.md", true},
-		{"docs/design/m2_starter_set_ability_survey.md", true},
-		{"pkg/toolchain/linter/fences/fences.go", true},
-		{"cmd/cardtown/main.go", true},
-		{"tools/importer/main.go", true},
-		{"scripts/x.go", false},
-		{"pkg/sim/readme.md", false},
-		{"data/decks/x.txt", false},
-	}
-	for _, c := range cases {
-		if got := isFencesPath(c.path); got != c.want {
-			t.Errorf("isFencesPath(%q) = %v, want %v", c.path, got, c.want)
-		}
-	}
-}
-
 func TestSelectChanged_Case1_Paths(t *testing.T) {
 	cases := []struct {
 		name  string
```

### 2. `pkg/toolchain/linter/fences/fences_test.go` (commit 1: `TestReads` added at the end of the file)
```diff
@@ -1007,3 +1007,31 @@ func TestFences_Case40_SimImportsOnlyEngine(t *testing.T) {
 		t.Fatalf("expected %q, got:\n%s", want, out)
 	}
 }
+
+// Reads covers Go files under pkg/, cmd/ and tools/, plus the data inputs (directive #172).
+func TestReads(t *testing.T) {
+	cases := []struct {
+		path string
+		want bool
+	}{
+		{"pkg/sim/x.go", true},
+		{"pkg/engine/x.go", true},
+		{"pkg/engine/cards/animus.go", true},
+		{"pkg/foo/bar/x.go", true},
+		{"pkg/sim/x_test.go", true},
+		{"data/cards/units.csv", true},
+		{"docs/frd/keyword_interaction_matrix.md", true},
+		{"docs/design/m2_starter_set_ability_survey.md", true},
+		{"pkg/toolchain/linter/fences/fences.go", true},
+		{"cmd/cardtown/main.go", true},
+		{"tools/importer/main.go", true},
+		{"scripts/x.go", false},
+		{"pkg/sim/readme.md", false},
+		{"data/decks/x.txt", false},
+	}
+	for _, c := range cases {
+		if got := fences.Reads(c.path); got != c.want {
+			t.Errorf("fences.Reads(%q) = %v, want %v", c.path, got, c.want)
+		}
+	}
+}
```

### 3. `pkg/toolchain/linter/rules/rules_test.go` (commit 1: `TestReads` added at the end of the file)
```diff
@@ -21,3 +21,21 @@ func TestRulesLint(t *testing.T) {
 		t.Errorf("expected rules lint to pass, but failed")
 	}
 }
+
+// Reads covers FRD 01 and the Go files under pkg/engine (directive #172).
+func TestReads(t *testing.T) {
+	cases := []struct {
+		path string
+		want bool
+	}{
+		{"docs/frd/01_game_rules_frd.md", true},
+		{"pkg/engine/x.go", true},
+		{"pkg/engine/x_test.go", true},
+		{"pkg/sim/x.go", false},
+	}
+	for _, c := range cases {
+		if got := rules.Reads(c.path); got != c.want {
+			t.Errorf("rules.Reads(%q) = %v, want %v", c.path, got, c.want)
+		}
+	}
+}
```

### 4. `cmd/cardtown/lint.go` (commit 2: `isRulesPath` and `isFencesPath` go; the last hunk removes the end of the file)
```diff
@@ -214,11 +214,6 @@ func runLint(args []string) {
 	}
 }
 
-// isRulesPath reports whether a change can affect executable rules traceability.
-func isRulesPath(pLower string) bool {
-	return pLower == strings.ToLower(specdoc.FRDDir+"/"+specdoc.RulesFRD) || (strings.HasPrefix(pLower, "pkg/engine/") && strings.HasSuffix(pLower, ".go"))
-}
-
 // lintSelection is which linters a lint run runs, and the changed FRD documents and card
 // sheets that the FRD, sheet and syntax linters then read.
 type lintSelection struct {
@@ -262,10 +257,10 @@ func selectChanged(paths []string, gitErr error) lintSelection {
 	var s lintSelection
 	for _, path := range paths {
 		pLower := strings.ToLower(filepath.ToSlash(path))
-		if isRulesPath(pLower) {
+		if rules.Reads(pLower) {
 			s.rules = true
 		}
-		if isFencesPath(pLower) {
+		if fences.Reads(pLower) {
 			s.fences = true
 		}
 		if strings.HasPrefix(pLower, specdoc.FRDDir+"/") && strings.HasSuffix(pLower, ".md") {
@@ -279,12 +274,3 @@ func selectChanged(paths []string, gitErr error) lintSelection {
 	s.syntax = s.sheets
 	return s
 }
-
-// isFencesPath reports whether a change can affect the package fences: any Go file under
-// pkg/, cmd/ or tools/ (the alias check reads them all), plus the data inputs.
-func isFencesPath(pLower string) bool {
-	return ((strings.HasPrefix(pLower, "pkg/") || strings.HasPrefix(pLower, "cmd/") || strings.HasPrefix(pLower, "tools/")) && strings.HasSuffix(pLower, ".go")) ||
-		(strings.HasPrefix(pLower, cardsheet.Dir+"/") && strings.HasSuffix(pLower, ".csv")) ||
-		pLower == specdoc.FRDDir+"/"+specdoc.MatrixFile ||
-		pLower == "docs/design/"+fences.SurveyFile
-}
```

### 5. `pkg/toolchain/linter/fences/fences.go` (commit 2: `surveyDir`, `scannedFiles` over `goFilesUnder`, `Reads`)
```diff
@@ -31,10 +31,13 @@ var (
 // modulePath is the Go module path that import rules are written against.
 const modulePath = "animus-mundi"
 
-// SurveyFile is the ability survey's file name, in docs/design. OPEN(...) markers may cite
+// SurveyFile is the ability survey's file name, in surveyDir. OPEN(...) markers may cite
 // the items of its Ruling Questions table.
 const SurveyFile = "m2_starter_set_ability_survey.md"
 
+// surveyDir is the repo-relative directory that holds SurveyFile, slash-separated.
+const surveyDir = "docs/design"
+
 // allowedImports maps a package directory to the in-module import paths (relative to
 // the module root) its non-test files may use. Packages not listed are not checked.
 var allowedImports = map[string][]string{
@@ -202,7 +205,7 @@ func readMarkerIDs(repoRoot string) (*markerIDs, error) {
 
 	openSurvey := make(map[string]bool)
 	ruledSurvey := make(map[string]bool)
-	surveyPath := filepath.Join(repoRoot, "docs", "design", SurveyFile)
+	surveyPath := filepath.Join(repoRoot, surveyDir, SurveyFile)
 	sContent, err := os.ReadFile(surveyPath)
 	if err != nil {
 		return nil, fmt.Errorf("reading %s: %w", SurveyFile, err)
@@ -245,30 +248,15 @@ func readMarkerIDs(repoRoot string) (*markerIDs, error) {
 // scannedFiles lists the non-test Go files under pkg/ that the card-name and marker
 // scan reads: all but pkg/engine/cards and pkg/toolchain, slash-separated.
 func scannedFiles(repoRoot string) ([]string, error) {
-	pkgDir := filepath.Join(repoRoot, "pkg")
+	all, err := goFilesUnder(repoRoot, "pkg")
+	if err != nil {
+		return nil, err
+	}
 	var files []string
-	err := filepath.WalkDir(pkgDir, func(path string, d fs.DirEntry, err error) error {
-		if err != nil {
-			return err
-		}
-		rel, err := filepath.Rel(repoRoot, path)
-		if err != nil {
-			return err
-		}
-		rel = filepath.ToSlash(rel)
-		if d.IsDir() {
-			if rel == "pkg/engine/cards" || rel == "pkg/toolchain" {
-				return filepath.SkipDir
-			}
-			return nil
-		}
-		if strings.HasSuffix(rel, ".go") && !strings.HasSuffix(rel, "_test.go") {
+	for _, rel := range all {
+		if !strings.HasPrefix(rel, "pkg/engine/cards/") && !strings.HasPrefix(rel, "pkg/toolchain/") {
 			files = append(files, rel)
 		}
-		return nil
-	})
-	if err != nil {
-		return nil, fmt.Errorf("reading pkg directory: %w", err)
 	}
 	return files, nil
 }
@@ -392,6 +380,23 @@ func Lint(repoRoot string, _ []string) (bool, error) {
 	return true, nil
 }
 
+// Reads reports whether a change to path (repo-relative, slash-separated and lower-cased)
+// can affect the fences: any Go file under aliasRoots (the alias check reads them all),
+// a card sheet, the keyword matrix or the ability survey.
+func Reads(path string) bool {
+	if strings.HasSuffix(path, ".go") {
+		for _, root := range aliasRoots {
+			if strings.HasPrefix(path, root+"/") {
+				return true
+			}
+		}
+		return false
+	}
+	return (strings.HasPrefix(path, strings.ToLower(cardsheet.Dir)+"/") && strings.HasSuffix(path, ".csv")) ||
+		path == strings.ToLower(specdoc.FRDDir+"/"+specdoc.MatrixFile) ||
+		path == strings.ToLower(surveyDir+"/"+SurveyFile)
+}
+
 // goFilesUnder lists non-test Go files under a repo-relative directory, slash-separated.
 func goFilesUnder(repoRoot, relDir string) ([]string, error) {
 	root := filepath.Join(repoRoot, filepath.FromSlash(relDir))
```

### 6. `pkg/toolchain/linter/rules/rules.go` (commit 2: `Reads`)
```diff
@@ -119,6 +119,13 @@ func Lint(repoRoot string, _ []string) (bool, error) {
 	return true, nil
 }
 
+// Reads reports whether a change to path (repo-relative, slash-separated and lower-cased)
+// can affect the rules linter: FRD 01, or a Go file under EngineDir.
+func Reads(path string) bool {
+	return path == strings.ToLower(specdoc.FRDDir+"/"+specdoc.RulesFRD) ||
+		(strings.HasPrefix(path, strings.ToLower(filepath.ToSlash(EngineDir))+"/") && strings.HasSuffix(path, ".go"))
+}
+
 func frdRequirements(path string) (defined, subsumed map[string]bool, err error) {
 	lines, err := specdoc.ReadLines(path)
 	if err != nil {
```

## Acceptance Cases (harness tasks only; written by the lead)
W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **The tests fail before the implementation:** at the `test(T-0285)` commit, `go -C W vet ./pkg/toolchain/linter/fences/` fails, reporting `undefined: fences.Reads` in `fences_test.go`, and `go -C W vet ./pkg/toolchain/linter/rules/` fails, reporting `undefined: rules.Reads` in `rules_test.go`.
2. **The old names are gone:** `git -C W grep -n -e isFencesPath -e isRulesPath -- "*.go"` prints nothing.
3. **Each input path is written once:** `git -C W grep -n -i -e "data/cards" -e "docs/frd" -e "data., .cards" -e "docs., .frd" -e "01_game_rules_frd" -e "docs/design" -e "docs., .design" -- pkg/toolchain cmd ":!*_test.go"` prints exactly eleven lines:
   - `cmd/cardtown/lint.go:138` (a printed message);
   - `pkg/toolchain/linter/cardsheet/cardsheet.go:1`, `:15` and `:19`;
   - `pkg/toolchain/linter/fences/fences.go:39`, `:124`, `:129` and `:132` (the `surveyDir` declaration, a comment and two error messages);
   - `pkg/toolchain/specdoc/matrix.go:8`;
   - `pkg/toolchain/specdoc/specdoc.go:13` and `:16`.
4. **One walk for Go files in the fences linter:** `git -C W grep -n "filepath.WalkDir" -- pkg/toolchain/linter/fences ":!*_test.go"` prints exactly one line, `pkg/toolchain/linter/fences/fences.go:407` (in `goFilesUnder`).
5. **The tests pass, with only the authorized edits:**
   - `go -C W test ./pkg/toolchain/... ./cmd/...` passes, including `TestReads` in both packages and every `TestSelectChanged_*`.
   - `git -C W diff autoloop/m2-wave-h.2 --numstat -- "*_test.go"` prints exactly three lines: `0	28	cmd/cardtown/lint_test.go`, `28	0	pkg/toolchain/linter/fences/fences_test.go` and `18	0	pkg/toolchain/linter/rules/rules_test.go`.
6. **The linters' output is unchanged:** `go -C W run ./cmd/cardtown lint --all` exits 0 and ends with `LINTER SUMMARY: All executed checks PASSED!`. Its fences section ends with `[PASS] Package code is free of card names, invalid markers, disallowed imports and type aliases.`
7. **Sims byte-identical:** `go -C W tool task sim-compare -- autoloop/m2-wave-h.2` prints `[PASS]`.
8. **Only the listed edits:**
   - `git -C W diff autoloop/m2-wave-h.2 --numstat -- pkg cmd docs` prints exactly six lines: `2	16	cmd/cardtown/lint.go`, `0	28	cmd/cardtown/lint_test.go`, `28	23	pkg/toolchain/linter/fences/fences.go`, `28	0	pkg/toolchain/linter/fences/fences_test.go`, `7	0	pkg/toolchain/linter/rules/rules.go` and `18	0	pkg/toolchain/linter/rules/rules_test.go`.
   - `git -C W diff autoloop/m2-wave-h.2 --stat` lists only those six files and this task file.
9. `gofmt -l W/pkg W/cmd` prints nothing; `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `cmd/cardtown/lint.go` (Edit 4: `isRulesPath` and `isFencesPath` go; `selectChanged` calls `rules.Reads` and `fences.Reads`)
- `pkg/toolchain/linter/fences/fences.go` (Edit 5: `surveyDir`, `scannedFiles`, `Reads`)
- `pkg/toolchain/linter/rules/rules.go` (Edit 6: `Reads`)

## Tests
Authorized edits (#172), in the `test(T-0285)` commit:
- `cmd/cardtown/lint_test.go` (Edit 1): `TestIsFencesPath` is removed (it moves, Edit 2). `TestSelectChanged_Case1_Paths`, `TestSelectChanged_Case2_BothModes` and `TestSelectChanged_Case3_GitFailure` stay unedited.
- `pkg/toolchain/linter/fences/fences_test.go` (Edit 2): `TestReads`, with `TestIsFencesPath`'s fourteen cases, calling `fences.Reads`.
- `pkg/toolchain/linter/rules/rules_test.go` (Edit 3): `TestReads`, four cases.

## Done When
- [ ] Cases 1-9 hold, with their commands and results in the Log
- [ ] `go tool task sim-compare -- autoloop/m2-wave-h.2` passes (starter and stress sims byte-identical to the run branch)
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-19, lead-architect (wave H.2 run, session 4): split off from #172 as the directive allows ((a) first, in T-0284). Queued behind T-0284; #173 waits for this task, because #172 blocks it ("Wave H.2, after #172").
- 2026-11-19, lead-architect (wave H.2 run, session 5): written in full against d64b6fa for #172 (b), next in wave H.2's order (T-0284 is `done`).
  - `main` (1859fe4) has nothing the run branch lacks, so no sync. Not `WAVE-COMPLETE` before the task: #172, #173 and #174 are `open: before the next Deliverables task`.
  - This is a harness task, so it skips `tests` and goes straight to `ready`. Tier is `flash`, because the edits are exact and were dry-run.
  - Every command here is on the headless allow-list: one per line, no pipes, no `$`, and patterns with special characters are quoted.
  - Waiting directives checked: #6, #7 (mode `serial`), #15 (2) (no draw or end-of-turn hook), #71, #91 (2), #106, #120, #142 (wave H.2 has no cards). None has fired.
  - Next: `kernel-engineer` on Flash.
