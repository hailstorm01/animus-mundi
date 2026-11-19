---
id: T-0274
title: "`cardtown lint fences` rejects type aliases (directive #165, ADR 0007; wave H.1 harness, flash)"
kind: harness
tier: flash
wave: H.1 (board directive #165)
status: done
branch: m2/autoloop-wave-h.1/T-0274-fences-reject-type-aliases
depends_on: [T-0273]
claims: [func:fences.checkAliases, var:fences.aliasRoots, func:isFencesPath]
cards: []
requirements: []
---

## Goal
`cardtown lint fences` reports every type alias declaration (`type X = Y`: a `TypeSpec` with `Assign` set, at any depth, generic or not, alone or in a group) in the non-test Go files under `pkg/`, `cmd/` and `tools/`, naming the file, the line and the alias, and citing ADR 0007. A move or rename updates every reference, so no old name survives as an alias. The repository has no alias left (#159, T-0272 and T-0273), so it passes. `isFencesPath` in `cmd/cardtown` widens to match, so a staged Go file anywhere under `pkg/`, `cmd/` or `tools/` runs the fences linter. No game code changes, so match output is unchanged. When this merges, #165 is `applied`.

## Inputs
- Board directive #165 (user, 2026-11-18, ADR 0007), `open: before the next Deliverables task`; wave H.1, right after #159 (applied, T-0273).
- [ADR 0007](../../../docs/architecture/decisions/0007-moves-update-every-reference.md): "`cardtown lint fences` rejects type alias declarations in non-test Go code (board directive #165)". Forwarding wrappers stay a plan and review check, not this task.
- Where it is today (at `autoloop/m2-wave-h.1` ee15b7f):
  - `fences.Analyze` ([fences.go:75-113](../../../pkg/toolchain/linter/fences/fences.go)) gathers the card-name and marker scan (`scanFiles`, over `scannedFiles`: `pkg/` minus `pkg/engine/cards` and `pkg/toolchain`) and the import check (`checkAllImports`), then sorts every `violationItem` by file, line and message. `goFilesUnder` ([fences.go:390](../../../pkg/toolchain/linter/fences/fences.go)) lists the non-test Go files under a directory and returns nothing for a missing one.
  - `Lint`'s PASS line ([fences.go:384](../../../pkg/toolchain/linter/fences/fences.go)), pinned by `TestFences_Case36_PassLine`.
  - `isFencesPath` ([lint.go:282-289](../../../cmd/cardtown/lint.go)) leaves out `pkg/toolchain/` and every file outside `pkg/`, pinned by `TestIsFencesPath` and the "toolchain and readme" row of `TestSelectChanged_Case1_Paths` ([lint_test.go](../../../cmd/cardtown/lint_test.go)).
- Design (lead decisions):
  - **A file of its own.** The check is one job, so it goes in a new `pkg/toolchain/linter/fences/aliases.go`: `aliasRoots` (`pkg`, `cmd`, `tools`) and `checkAliases(repoRoot) ([]violationItem, error)`, which parses each file from `goFilesUnder` with `go/parser` and walks it with `ast.Inspect`, so a function-local alias is caught too. `Analyze` appends its items before the sort, so they interleave with the other findings by file and line.
  - **Scope wider than the card-name scan.** Aliases are banned everywhere ADR 0007 applies, so `pkg/engine/cards` and `pkg/toolchain` are in scope, and so are `cmd/` and `tools/`. Test files are out (`goFilesUnder` skips `_test.go`), as the directive says.
  - **A file that doesn't parse is skipped.** The import check already reports a broken import block for its packages, and the compiler reports the rest, so the alias check adds no second error for the same file (`TestFences_Case35_UnreadableImports` counts exactly one error for `pkg/sim/bad.go`, unedited).
  - **Message:** `declares type alias <Name>; a move or rename updates every reference instead (ADR 0007)`.
  - **PASS line:** `  [PASS] Package code is free of card names, invalid markers, disallowed imports and type aliases.` (Test edit 1).
  - **`isFencesPath`:** any `.go` file under `pkg/`, `cmd/` or `tools/`, plus the data inputs as today (Test edits 2 and 3).
- Reuse search (D2.3): `git grep -n "Assign" -- pkg cmd tools` at ee15b7f finds no alias check; `goFilesUnder` is reused, not copied. `git grep -nE "type \w+ = " -- "*.go"` finds no alias in the repository.
- Size (D2.4): three non-test files.
- **Expected output difference:** none for the game: `go tool task sim-compare` passes. The fences linter's PASS line gains "type aliases".
- The lead dry-ran these exact edits and test files in a throwaway worktree of the run branch at ee15b7f. Before Edits 1-3, the new tests and Test edit 1 fail (`Case36` and `Case41` to `Case47`), and so do Test edits 2 and 3 (`TestIsFencesPath`, `TestSelectChanged_Case1_Paths`). With them: `gofmt -l cmd pkg` clean; `go test ./pkg/toolchain/linter/fences/ ./cmd/cardtown/` ok; `cardtown lint fences` exit 0; `go tool task check` exit 0. Blobs: `aliases.go` ac7c9f6, `fences.go` a2b14c4, `lint.go` 6e2a894, `aliases_test.go` 3593f07, `fences_test.go` 715d195, `lint_test.go` fd6b126.

## Commits
1. `test(T-0274): cardtown lint fences rejects type aliases` with only the new test file below and Test edits 1-3. Before committing, run `go -C W test ./pkg/toolchain/linter/fences/ ./cmd/cardtown/` and record that it fails.
2. `feat(toolchain): cardtown lint fences rejects type aliases (T-0274)` with Edits 1-3.

## Tests (lead decisions; copy exactly, tabs for indentation)
### New file `pkg/toolchain/linter/fences/aliases_test.go`
```go
package fences_test

import (
	"strings"
	"testing"
)

// Tests for T-0274 (directive #165, ADR 0007): cardtown lint fences rejects every type
// alias declaration in non-test Go files under pkg/, cmd/ and tools/. A move or rename
// updates every reference, so no old name survives as an alias.

const aliasRule = "a move or rename updates every reference instead (ADR 0007)"

// 41. An alias in a non-test file fails, naming its file, line and name.
func TestFences_Case41_TypeAliasFails(t *testing.T) {
	repo := setupFixtureRepo(t)
	writePkgFile(t, repo, "pkg/engine/alias.go", "package engine\n\nimport \"animus-mundi/pkg/engine/ability\"\n\ntype Keywords = ability.Keywords\n")
	out := expectFail(t, repo)
	want := "[ERROR] pkg/engine/alias.go:5: declares type alias Keywords; " + aliasRule
	if strings.Count(out, "[ERROR]") != 1 || !strings.Contains(out, want) {
		t.Fatalf("expected one error %q, got:\n%s", want, out)
	}
}

// 42. Each alias in a grouped type declaration fails on its own line; a defined type
// in the same group passes.
func TestFences_Case42_GroupedAliasesFail(t *testing.T) {
	repo := setupFixtureRepo(t)
	writePkgFile(t, repo, "pkg/sim/group.go", "package sim\n\ntype (\n\tA = int\n\tB string\n\tC = []A\n)\n")
	out := expectFail(t, repo)
	if strings.Count(out, "[ERROR]") != 2 {
		t.Fatalf("expected two errors, got:\n%s", out)
	}
	for _, want := range []string{
		"[ERROR] pkg/sim/group.go:4: declares type alias A; ",
		"[ERROR] pkg/sim/group.go:6: declares type alias C; ",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected %q, got:\n%s", want, out)
		}
	}
}

// 43. An alias declared inside a function body fails too.
func TestFences_Case43_AliasInFunctionBodyFails(t *testing.T) {
	repo := setupFixtureRepo(t)
	writePkgFile(t, repo, "pkg/view/local.go", "package view\n\nfunc f() {\n\ttype n = int\n\tvar _ n\n}\n")
	out := expectFail(t, repo)
	if !strings.Contains(out, "[ERROR] pkg/view/local.go:4: declares type alias n; ") {
		t.Fatalf("expected local.go:4, got:\n%s", out)
	}
}

// 44. A generic alias fails.
func TestFences_Case44_GenericAliasFails(t *testing.T) {
	repo := setupFixtureRepo(t)
	writePkgFile(t, repo, "pkg/sim/generic.go", "package sim\n\ntype List[T any] = []T\n")
	out := expectFail(t, repo)
	if !strings.Contains(out, "[ERROR] pkg/sim/generic.go:3: declares type alias List; ") {
		t.Fatalf("expected generic.go:3, got:\n%s", out)
	}
}

// 45. The alias check covers every non-test Go file under pkg/, cmd/ and tools/,
// including pkg/engine/cards and pkg/toolchain, which the card-name scan skips.
func TestFences_Case45_AliasScope(t *testing.T) {
	for _, rel := range []string{
		"pkg/engine/cards/a.go",
		"pkg/toolchain/x/a.go",
		"pkg/foo/bar/a.go",
		"cmd/cardtown/a.go",
		"tools/importer/a.go",
	} {
		repo := setupFixtureRepo(t)
		writePkgFile(t, repo, rel, "package a\n\ntype X = int\n")
		out := expectFail(t, repo)
		if !strings.Contains(out, "[ERROR] "+rel+":3: declares type alias X; ") {
			t.Fatalf("%s: expected an alias error, got:\n%s", rel, out)
		}
	}
}

// 46. Test files, files outside pkg/, cmd/ and tools/, and defined types pass, and the
// PASS line names the alias check.
func TestFences_Case46_AliasExemptions(t *testing.T) {
	repo := setupFixtureRepo(t)
	writePkgFile(t, repo, "pkg/engine/alias_test.go", "package engine\n\ntype X = int\n")
	writePkgFile(t, repo, "cmd/cardtown/a_test.go", "package main\n\ntype X = int\n")
	writePkgFile(t, repo, "scripts/a.go", "package scripts\n\ntype X = int\n")
	writePkgFile(t, repo, "pkg/engine/defined.go", "package engine\n\ntype Y int\n\ntype (\n\tZ struct{}\n\tW Y\n)\n")
	out := expectPass(t, repo)
	if !strings.Contains(out, "disallowed imports and type aliases.") {
		t.Fatalf("expected the PASS line to name type aliases, got:\n%s", out)
	}
}

// 47. A file that doesn't parse gets no alias error, only the import check's; an alias
// in another file is still reported.
func TestFences_Case47_UnparseableFileHasNoAliasError(t *testing.T) {
	repo := setupFixtureRepo(t)
	writePkgFile(t, repo, "pkg/sim/bad.go", "package sim\n\nimport (\n\t\"fmt\n)\n\ntype X = int\n")
	writePkgFile(t, repo, "pkg/sim/good.go", "package sim\n\ntype Y = int\n")
	out := expectFail(t, repo)
	if strings.Count(out, "[ERROR]") != 2 ||
		!strings.Contains(out, "[ERROR] pkg/sim/bad.go:4: cannot read imports") ||
		!strings.Contains(out, "[ERROR] pkg/sim/good.go:3: declares type alias Y; ") {
		t.Fatalf("expected bad.go's import error and good.go's alias error only, got:\n%s", out)
	}
}
```

### Test edit 1. `pkg/toolchain/linter/fences/fences_test.go`, `TestFences_Case36_PassLine`: the PASS line names type aliases
Replace the one line
```go
	want := "  [PASS] Package code is free of card names, invalid markers and disallowed imports."
```
with
```go
	want := "  [PASS] Package code is free of card names, invalid markers, disallowed imports and type aliases."
```

### Test edit 2. `cmd/cardtown/lint_test.go`, `TestIsFencesPath`: every Go file under pkg/, cmd/ and tools/
Replace the comment line
```go
// 37. isFencesPath covers pkg/ Go files except pkg/toolchain/, plus the data inputs.
```
with
```go
// 37. isFencesPath covers Go files under pkg/, cmd/ and tools/, plus the data inputs.
```
and replace the two rows
```go
		{"pkg/toolchain/linter/fences/fences.go", false},
		{"cmd/cardtown/main.go", false},
```
with
```go
		{"pkg/toolchain/linter/fences/fences.go", true},
		{"cmd/cardtown/main.go", true},
		{"tools/importer/main.go", true},
		{"scripts/x.go", false},
```

### Test edit 3. `cmd/cardtown/lint_test.go`, `TestSelectChanged_Case1_Paths`: a toolchain Go file selects the fences linter
In the `"toolchain and readme"` row, replace
```go
			want:  lintSelection{},
```
with
```go
			want:  lintSelection{fences: true},
```
Its `name` and `paths` stay as they are.

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)
### 1. New file `pkg/toolchain/linter/fences/aliases.go`
```go
package fences

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
)

// aliasRoots are the repo-relative directories whose non-test Go files may declare no
// type alias: a move or rename updates every reference instead (ADR 0007).
var aliasRoots = []string{"pkg", "cmd", "tools"}

// checkAliases reports every type alias declaration (type X = Y), at any depth, in the
// non-test Go files under aliasRoots. A file that doesn't parse is skipped: the import
// check and the compiler report it.
func checkAliases(repoRoot string) ([]violationItem, error) {
	var items []violationItem
	for _, root := range aliasRoots {
		files, err := goFilesUnder(repoRoot, root)
		if err != nil {
			return nil, err
		}
		for _, rel := range files {
			fset := token.NewFileSet()
			f, err := parser.ParseFile(fset, filepath.Join(repoRoot, filepath.FromSlash(rel)), nil, parser.SkipObjectResolution)
			if err != nil {
				continue
			}
			ast.Inspect(f, func(n ast.Node) bool {
				if ts, ok := n.(*ast.TypeSpec); ok && ts.Assign.IsValid() {
					items = append(items, violationItem{
						file: rel,
						line: fset.Position(ts.Pos()).Line,
						msg:  fmt.Sprintf("declares type alias %s; a move or rename updates every reference instead (ADR 0007)", ts.Name.Name),
					})
				}
				return true
			})
		}
	}
	return items, nil
}
```

### 2. `pkg/toolchain/linter/fences/fences.go`: `Analyze` runs the alias check, and the PASS line names it
```diff
@@ -70,8 +70,9 @@ type markerIDs struct {
 }
 
 // Analyze scans non-test Go files under pkg/ (except pkg/engine/cards and pkg/toolchain)
-// for card names and invalid markers, and checks the imports of the packages in
-// allowedImports (pkg/engine/cards included).
+// for card names and invalid markers, checks the imports of the packages in
+// allowedImports (pkg/engine/cards included), and reports every type alias in the
+// non-test Go files under pkg/, cmd/ and tools/ (checkAliases).
 func Analyze(repoRoot string) (*Report, error) {
 	cards, err := loadCardNames(repoRoot)
 	if err != nil {
@@ -94,6 +95,11 @@ func Analyze(repoRoot string) (*Report, error) {
 		return nil, err
 	}
 	items = append(items, importItems...)
+	aliasItems, err := checkAliases(repoRoot)
+	if err != nil {
+		return nil, err
+	}
+	items = append(items, aliasItems...)
 
 	sort.Slice(items, func(i, j int) bool {
 		if items[i].file != items[j].file {
@@ -381,7 +387,7 @@ func Lint(repoRoot string, _ []string) (bool, error) {
 		return false, nil
 	}
 
-	fmt.Println("  [PASS] Package code is free of card names, invalid markers and disallowed imports.")
+	fmt.Println("  [PASS] Package code is free of card names, invalid markers, disallowed imports and type aliases.")
 	fmt.Println("==========================================================")
 	return true, nil
 }
```

### 3. `cmd/cardtown/lint.go`: `isFencesPath` takes every Go file under pkg/, cmd/ and tools/
```diff
@@ -279,9 +279,10 @@ func selectChanged(paths []string, gitErr error) lintSelection {
 	return s
 }
 
-// isFencesPath reports whether a change can affect the package fences.
+// isFencesPath reports whether a change can affect the package fences: any Go file under
+// pkg/, cmd/ or tools/ (the alias check reads them all), plus the data inputs.
 func isFencesPath(pLower string) bool {
-	return (strings.HasPrefix(pLower, "pkg/") && !strings.HasPrefix(pLower, "pkg/toolchain/") && strings.HasSuffix(pLower, ".go")) ||
+	return ((strings.HasPrefix(pLower, "pkg/") || strings.HasPrefix(pLower, "cmd/") || strings.HasPrefix(pLower, "tools/")) && strings.HasSuffix(pLower, ".go")) ||
 		(strings.HasPrefix(pLower, "data/cards/") && strings.HasSuffix(pLower, ".csv")) ||
 		pLower == "docs/frd/"+specdoc.MatrixFile ||
 		pLower == "docs/design/"+fences.SurveyFile
```

## Acceptance Cases (harness tasks only; written by the lead)
W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **The tests failed first:** at the `test(T-0274)` commit, before Edits 1-3, `go -C W test ./pkg/toolchain/linter/fences/ ./cmd/cardtown/` fails: `TestFences_Case36_PassLine`, `TestFences_Case41` to `TestFences_Case47`, `TestIsFencesPath` and `TestSelectChanged_Case1_Paths` fail.
2. **An alias fails with its file, line and name** (`type X = other.Y`, grouped, function-local, generic): `go -C W test ./pkg/toolchain/linter/fences/ -run "TestFences_Case4[1-4]"` passes.
3. **Scope:** an alias under `pkg/engine/cards`, `pkg/toolchain`, a new `pkg/` package, `cmd/` and `tools/` fails; the same declaration in a `_test.go` file, or outside those three directories, passes; a defined type (`type X Y`) passes: `go -C W test ./pkg/toolchain/linter/fences/ -run "TestFences_Case4[5-6]"` passes.
4. **An unparseable file gets one error, not two:** `go -C W test ./pkg/toolchain/linter/fences/ -run "TestFences_Case3[5-6]|TestFences_Case47"` passes, `TestFences_Case35_UnreadableImports` unedited.
5. **The staged-file selection matches:** `go -C W test ./cmd/cardtown/ -run "TestIsFencesPath|TestSelectChanged"` passes.
6. **The repository passes:** `go -C W run ./cmd/cardtown lint fences` exits 0 and prints the new PASS line.
7. **Sims byte-identical:** `go -C W tool task sim-compare -- autoloop/m2-wave-h.1` prints `[PASS]`.
8. **Only the listed edits:** `git -C W diff autoloop/m2-wave-h.1 --numstat -- pkg cmd` prints exactly six lines: `3	2	cmd/cardtown/lint.go`, `6	4	cmd/cardtown/lint_test.go`, `44	0	pkg/toolchain/linter/fences/aliases.go`, `109	0	pkg/toolchain/linter/fences/aliases_test.go`, `9	3	pkg/toolchain/linter/fences/fences.go` and `1	1	pkg/toolchain/linter/fences/fences_test.go`.
9. **Nothing else changed:** `git -C W diff autoloop/m2-wave-h.1 --stat` lists only the six files of case 8 and this task file.
10. `gofmt -l W/cmd W/pkg` prints nothing; `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/toolchain/linter/fences/aliases.go` (new)
- `pkg/toolchain/linter/fences/fences.go`
- `cmd/cardtown/lint.go`
- `pkg/toolchain/linter/fences/aliases_test.go` (new; the `test(T-0274)` commit only)
- `pkg/toolchain/linter/fences/fences_test.go` (Test edit 1 only; the `test(T-0274)` commit only)
- `cmd/cardtown/lint_test.go` (Test edits 2 and 3 only; the `test(T-0274)` commit only)

## Done When
- [x] Cases 1-10 hold, with their commands and results in the Log
- [x] `go tool task sim-compare -- autoloop/m2-wave-h.1` passes (starter and stress sims byte-identical to the run branch)
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-18, lead-architect (wave H.1 run, session 3): written in full against ee15b7f for #165. `main` (7a8e4f5) has nothing the run branch lacks, so no sync. Not `WAVE-COMPLETE` before the task (#165, #160, #161, #166, #167, #168, #162, #163 and #164 are `open: before the next Deliverables task`). Harness, so it skips `tests` and goes straight to `ready`. Tier `flash`, because the edits and the test files are exact and were dry-run. Test edits 1-3 are the only changes to existing tests: the PASS line names the new check, and `isFencesPath` must cover the files the check reads. Every command here is on the headless allow-list: one per line, no pipes, no `$`, and the `-run` patterns with `|` or brackets are quoted. Waiting directives checked: #6, #7 (mode `serial`), #15 (2) (no draw or end-of-turn hook), #71, #91 (2), #106, #120, #142: not fired. Next: `kernel-engineer` on Flash.
- 2026-11-18, kernel-engineer: implemented T-0274 (directive #165, ADR 0007).
  - Changes:
    - Added `pkg/toolchain/linter/fences/aliases.go` defining `aliasRoots` (`pkg`, `cmd`, `tools`) and `checkAliases(repoRoot)` parsing non-test Go files with `ast.Inspect` to reject type aliases (`*ast.TypeSpec` with `Assign.IsValid()`), citing ADR 0007.
    - Updated `pkg/toolchain/linter/fences/fences.go`: `Analyze` calls `checkAliases` and aggregates alias violations; `Lint` prints updated PASS line including `and type aliases.`.
    - Updated `cmd/cardtown/lint.go`: `isFencesPath` includes any Go file under `pkg/`, `cmd/`, or `tools/`.
    - Added tests in `pkg/toolchain/linter/fences/aliases_test.go` (`TestFences_Case41` to `TestFences_Case47`).
    - Updated `pkg/toolchain/linter/fences/fences_test.go` (`TestFences_Case36_PassLine`) and `cmd/cardtown/lint_test.go` (`TestIsFencesPath`, `TestSelectChanged_Case1_Paths`).
  - Reused:
    - `goFilesUnder` in `pkg/toolchain/linter/fences/fences.go` for non-test Go file discovery.
  - Acceptance cases:
    - Case 1: `go -C W test ./pkg/toolchain/linter/fences/ ./cmd/cardtown/` failed at `test(T-0274)` before Edits 1-3 (Case36, Case41-47, TestIsFencesPath, TestSelectChanged_Case1_Paths failed).
    - Case 2: `go -C W test ./pkg/toolchain/linter/fences/ -run "TestFences_Case4[1-4]"` passed (`ok animus-mundi/pkg/toolchain/linter/fences 0.600s`).
    - Case 3: `go -C W test ./pkg/toolchain/linter/fences/ -run "TestFences_Case4[5-6]"` passed (`ok animus-mundi/pkg/toolchain/linter/fences 0.637s`).
    - Case 4: `go -C W test ./pkg/toolchain/linter/fences/ -run "TestFences_Case3[5-6]|TestFences_Case47"` passed (`ok animus-mundi/pkg/toolchain/linter/fences 0.597s`).
    - Case 5: `go -C W test ./cmd/cardtown/ -run "TestIsFencesPath|TestSelectChanged"` passed (`ok animus-mundi/cmd/cardtown 0.748s`).
    - Case 6: `go -C W run ./cmd/cardtown lint fences` passed (exit 0, `[PASS] Package code is free of card names, invalid markers, disallowed imports and type aliases.`).
    - Case 7: `go -C W tool task sim-compare -- autoloop/m2-wave-h.1` passed (`[PASS] starter and stress sims match autoloop/m2-wave-h.1 byte for byte, exit codes included`).
    - Case 8: `git -C W diff autoloop/m2-wave-h.1 --numstat -- pkg cmd` printed exactly the 6 expected files with matching stat numbers.
    - Case 9: `git -C W diff autoloop/m2-wave-h.1 --stat` confirmed only the 6 files and the task file changed.
    - Case 10: `gofmt -l W/cmd W/pkg` clean; `go -C W tool task sim` passed (200/200 games, all invariants hold); `go -C W tool task check` passed cleanly (fmt-check, vet, all tests, cardtown lint --all).
  - Status moved to `review`. Ready for `lead-architect`.
- 2026-11-18, lead-architect (wave H.1 run, session 3, D5): **Approved, `done`.** Merged into `autoloop/m2-wave-h.1` as df482ac. Flash run `c8be70d1-fd77-4007-ac3b-159e4b012d19`: exit 0, `status: SUCCESS`, no `denied_actions` key, and no step with `"status":"ERROR"` in its transcript; `.err` holds only agy's "root agent idle" notice. Review of `git diff autoloop/m2-wave-h.1...m2/autoloop-wave-h.1/T-0274-fences-reject-type-aliases` (c2da904):
  - Scope: the three Allowed non-test files, the new test file, Test edits 1-3 and this file. The `test(T-0274)` commit (b7ddece) holds only `aliases_test.go` 3593f07, `fences_test.go` 715d195 and `lint_test.go` fd6b126; the implementation commit (cf32497) holds only `aliases.go` ac7c9f6, `fences.go` a2b14c4 and `lint.go` 6e2a894: every blob is the lead's dry run's. The tests failed first: re-run at b7ddece by the reviewer, `Case36`, `Case41` to `Case47`, `TestIsFencesPath` and `TestSelectChanged_Case1_Paths` fail.
  - Fences and placement: the check lives in `pkg/toolchain/linter/fences` beside the other fences, in a file of its own; nothing in a game package changes. Reuse: `goFilesUnder` is reused; no second file walker. ADR 0007: `git grep` finds no type alias in non-test Go code, and `cardtown lint fences` now enforces it.
  - Re-run by the reviewer in a scratch worktree of the task branch: `gofmt -l cmd pkg` clean; `cardtown lint fences` exit 0 with the new PASS line; `go tool task check` exit 0; `go tool task sim-compare -- autoloop/m2-wave-h.1` `[PASS]`; `go tool task sim` exit 0, 200 games, P0 58.5% / P1 39.5% / draws 2.0%, P0 50/50 and P1 50/50 main cards executable.
  - The merge's tree matches the task tip's (`git diff` between them is empty), so the gates hold on the run branch. #165 is `applied` (T-0274) and moves to the directive archive. The design note's Fences bullet now names the alias check (this commit).
