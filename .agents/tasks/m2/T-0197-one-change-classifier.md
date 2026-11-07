---
id: T-0197
title: One classifier of changed files, and "all linters" means all (directive #134, harness, flash)
kind: harness
tier: flash
wave: D.2
status: in-progress
branch: m2/autoloop-wave-d.2/T-0197-one-change-classifier
depends_on: [T-0196]
claims: [cardtown:lintSelection, cardtown:allLinters, cardtown:changedFiles, cardtown:selectChanged, fences:SurveyFile]
cards: []
requirements: []
---

## Goal
`cardtown lint`'s smart mode classifies its changed files in one function, `selectChanged`, which the staged mode and the working-tree mode both call, and one function, `allLinters`, says what "all linters" is, for `--all` and for the git-failure path. That path now runs the fences linter too (#134). Apart from that path, a pure move: the same linters run on the same files with the same output. The matrix and survey file names that `isFencesPath` spells are read from the constants of the code that reads those files.

## Inputs
- Board directive #134 (structural review 9, approved by the user, 2026-11-06).
- The two copies at f97bea1: the staged loop at `cmd/cardtown/lint.go:137-161` and the status loop at `:162-186`. They differ only in the git call, the messages they print, and `path` vs `sf.Path`. Both failure paths (`:141-145`, `:166-170`) set four of the five linters and leave out fences, while `--all` (`:113-118`) sets all five: three spellings of "all", two of them wrong.
- **File names spelled once (D2.3).** T-0196's Inputs left the paths at `lint.go:288-289` to #134. They spell `keyword_interaction_matrix.md`, which `specdoc.MatrixFile` (`pkg/toolchain/specdoc/matrix.go:9`) and `frd.KeywordMatrix` (`pkg/toolchain/linter/frd/ownership.go:14`) also spell, and `m2_starter_set_ability_survey.md`, which `fences.go:195` and `:198` spell. Renaming either file would mean editing three spots. This task removes the copies instead of keeping them: `isFencesPath` and `frd.KeywordMatrix` read `specdoc.MatrixFile`, and a new `fences.SurveyFile` constant is read by `readMarkerIDs` and `isFencesPath`. The survey is read only by `fences` (searched `m2_starter_set_ability_survey` in `pkg` and `cmd`), so its name lives there. The missing-file error still names the file (`TestFences_Case24_MissingDocsError`). `pkg/engine/doc.go:108-110` names both files in prose and stays.
- Reuse search (D2.3): no other classifier of changed paths in `pkg` or `cmd` (searched `GetStagedFiles`, `GetStatus` and `isFencesPath`).
- Placement: `cmd/cardtown` already owns lint dispatch; the new type and functions stay in `lint.go`. Nothing in `pkg/engine` changes.
- The lead dry-ran these exact edits in a throwaway worktree of the run branch at f97bea1: `gofmt -l pkg cmd` clean, `go vet` clean, `go tool task check` passes and `go tool task sim-compare -- autoloop/m2-wave-d.2` printed `[PASS]`. Built before and after, `cardtown lint` printed byte-identical output and exit codes for `--all`, `-fences`, `-cards`, `-frd -rules`, no flags and `--staged`, on a clean checkout and with a staged FRD file and an edited engine file. With `-repo-root` set to a missing folder, the only difference was the added fences block (`>>> Running Engine Architecture Fences Linter...` and its output). Throwaway tests of cases 1-3 passed. Every expected output below is from that dry run.

## Acceptance Cases (harness tasks only; written by the lead)
Write cases 1-3 as tests appended to `cmd/cardtown/lint_test.go` (after `TestIsFencesPath`, which stays unchanged), with exactly these names, in the `test(T-0197)` commit. Compare selections with `reflect.DeepEqual`. Each must fail before the implementation commit (they don't compile yet; record the `go test` error in the Log). W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **`TestSelectChanged_Case1_Paths`** (#134): `selectChanged(paths, nil)` returns exactly the given `lintSelection`, table-driven:
   - `nil` → `lintSelection{}` (nothing selected, both lists nil);
   - `pkg/engine/x.go` → `rules`, `fences`;
   - `docs/frd/01_Game_Rules_FRD.md` → `frd`, `rules`, `changedFrds` = that path, with its case kept;
   - `data/cards/units.csv` → `sheets`, `syntax`, `fences`, `changedCsvs` = that path;
   - `docs/frd/keyword_interaction_matrix.md` → `frd`, `fences`, `changedFrds` = that path;
   - `docs/design/m2_starter_set_ability_survey.md` → `fences` only;
   - `pkg/toolchain/x.go` and `README.md` together → `lintSelection{}`.
2. **`TestSelectChanged_Case2_BothModes`** (#134: a list of changed paths selects the same linters in both modes): in a throwaway repo (`t.TempDir()`, then `git init`, run with `exec.Command("git", "-C", root, ...)` as `pkg/toolchain/git/git_test.go`'s `runGit` does; no commit needed), write `docs/frd/02_x.md`, `data/cards/units.csv` and `pkg/toolchain/x.go`, then `git add .`. For `staged` true and then false, `changedFiles(root, staged)` returns no error, and `selectChanged` of its result equals `lintSelection{frd: true, sheets: true, syntax: true, fences: true, changedFrds: []string{"docs/frd/02_x.md"}, changedCsvs: []string{"data/cards/units.csv"}}` (`rules` not selected).
3. **`TestSelectChanged_Case3_GitFailure`** (#134: the git-failure path selects all five): for `staged` true and then false, `changedFiles(filepath.Join(t.TempDir(), "missing"), staged)` returns an error, and `selectChanged` of its result and that error equals `allLinters()`. Also `selectChanged([]string{"README.md"}, errors.New("git failed"))` has all five of `frd`, `rules`, `sheets`, `syntax` and `fences` true, both lists nil, and `empty()` false.
4. **New tests pass, old tests unchanged:** `go -C W test ./cmd/cardtown -run TestSelectChanged -v` prints `--- PASS` for the three tests. `git -C W diff autoloop/m2-wave-d.2 -- pkg/toolchain/linter/fences/fences_test.go pkg/toolchain/linter/frd/ownership_test.go pkg/toolchain/git/git_test.go` prints nothing, and `git -C W diff autoloop/m2-wave-d.2 --numstat -- cmd/cardtown/lint_test.go` shows `0` deleted lines.
5. **Sims byte-identical:** `go -C W tool task sim-compare -- autoloop/m2-wave-d.2` prints `[PASS]`.
6. **One classifier, one "all":** `git -C W grep -n -e 'GetStagedFiles' -e 'GetStatus' -e 'isRulesPath(' -e 'isFencesPath(' -e 'allLinters()' -e 'Running all linters' -- cmd/cardtown/lint.go` prints exactly these eleven lines: `:107` (`sel = allLinters()`), `:123` and `:125` (the two failure messages), `:217` (`func isRulesPath`), `:229` (`func allLinters`), `:242` (`git.GetStagedFiles`), `:244` (`git.GetStatus`), `:259` (`return allLinters()`), `:264` (`isRulesPath(pLower)`), `:267` (`isFencesPath(pLower)`) and `:283` (`func isFencesPath`).
7. **Declarations:** `git -C W grep -n -e '^func ' -e '^type ' -- cmd/cardtown/lint.go` prints exactly these lines, in this order: `:21` `func runLint`, `:217` `func isRulesPath`, `:223` `type lintSelection`, `:229` `func allLinters`, `:234` `func (s lintSelection) empty`, `:240` `func changedFiles`, `:257` `func selectChanged`, `:283` `func isFencesPath`.
8. **Each file name spelled once in code:** `git -C W grep -n -e 'keyword_interaction_matrix' -e 'm2_starter_set_ability_survey' -- cmd/cardtown/lint.go pkg/toolchain/specdoc/matrix.go pkg/toolchain/linter/fences/fences.go pkg/toolchain/linter/frd/ownership.go` prints exactly two lines: `pkg/toolchain/linter/fences/fences.go:36` and `pkg/toolchain/specdoc/matrix.go:9`.
9. **Only the listed edits:** `git -C W diff autoloop/m2-wave-d.2 --numstat -- pkg cmd` prints exactly four lines: `94	96	cmd/cardtown/lint.go`, a `<N>	0	cmd/cardtown/lint_test.go` line, `6	2	pkg/toolchain/linter/fences/fences.go` and `3	1	pkg/toolchain/linter/frd/ownership.go`. `git -C W diff autoloop/m2-wave-d.2 --stat` lists only those four files and this task file.
10. `go -C W tool task check` and `go -C W tool task sim` pass.

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)
Line numbers are at f97bea1, before any edit. In each file, apply the edits from the bottom up (the order below), so the line numbers still hold.

### 1. `cmd/cardtown/lint.go`
**1a.** Replace lines 283-290 (the blank line after `isRulesPath`'s closing brace, through the closing brace of `isFencesPath`) with this block, which starts with one blank line:
```go

// lintSelection is which linters a lint run runs, and the changed FRD documents and card
// sheets that the FRD, sheet and syntax linters then read.
type lintSelection struct {
	frd, rules, sheets, syntax, fences bool
	changedFrds, changedCsvs           []string
}

// allLinters selects every linter.
func allLinters() lintSelection {
	return lintSelection{frd: true, rules: true, sheets: true, syntax: true, fences: true}
}

// empty reports whether s selects no linter.
func (s lintSelection) empty() bool {
	return !s.frd && !s.rules && !s.sheets && !s.syntax && !s.fences
}

// changedFiles lists the changed paths git reports: the staged files when staged is set,
// otherwise every path git status shows.
func changedFiles(repoRoot string, staged bool) ([]string, error) {
	if staged {
		return git.GetStagedFiles(repoRoot)
	}
	statusFiles, err := git.GetStatus(repoRoot)
	if err != nil {
		return nil, err
	}
	var paths []string
	for _, sf := range statusFiles {
		paths = append(paths, sf.Path)
	}
	return paths, nil
}

// selectChanged selects the linters that changes to paths can affect, or every linter when
// git couldn't list the changes (gitErr is not nil).
func selectChanged(paths []string, gitErr error) lintSelection {
	if gitErr != nil {
		return allLinters()
	}
	var s lintSelection
	for _, path := range paths {
		pLower := strings.ToLower(filepath.ToSlash(path))
		if isRulesPath(pLower) {
			s.rules = true
		}
		if isFencesPath(pLower) {
			s.fences = true
		}
		if strings.HasPrefix(pLower, "docs/frd/") && strings.HasSuffix(pLower, ".md") {
			s.changedFrds = append(s.changedFrds, path)
		} else if strings.HasPrefix(pLower, "data/cards/") && strings.HasSuffix(pLower, ".csv") {
			s.changedCsvs = append(s.changedCsvs, path)
		}
	}
	s.frd = len(s.changedFrds) > 0
	s.sheets = len(s.changedCsvs) > 0
	s.syntax = s.sheets
	return s
}

// isFencesPath reports whether a change can affect the package fences.
func isFencesPath(pLower string) bool {
	return (strings.HasPrefix(pLower, "pkg/") && !strings.HasPrefix(pLower, "pkg/toolchain/") && strings.HasSuffix(pLower, ".go")) ||
		(strings.HasPrefix(pLower, "data/cards/") && strings.HasSuffix(pLower, ".csv")) ||
		pLower == "docs/frd/"+specdoc.MatrixFile ||
		pLower == "docs/design/"+fences.SurveyFile
}
```

**1b.** Replace lines 104-265 (`runFrd := false` through the closing brace of the `if runFences {` block) with:
```go
	var sel lintSelection
	if *allFlag {
		sel = allLinters()
	} else if *frdFlag || *rulesFlag || *cardsFlag || *sheetsFlag || *syntaxFlag || *fencesFlag {
		sel.frd = *frdFlag
		sel.rules = *rulesFlag
		sel.sheets = *cardsFlag || *sheetsFlag
		sel.syntax = *cardsFlag || *syntaxFlag
		sel.fences = *fencesFlag
	} else {
		// Smart change detection
		if *stagedFlag {
			fmt.Println("[Detecting Staged Files via Git...]")
		} else {
			fmt.Println("[Detecting Changed Files via Git...]")
		}
		paths, err := changedFiles(repoRoot, *stagedFlag)
		if err != nil && *stagedFlag {
			fmt.Printf("Git staged diff unavailable: %v. Running all linters...\n", err)
		} else if err != nil {
			fmt.Println("  Git change detection unavailable. Running all linters...")
		}
		sel = selectChanged(paths, err)

		if len(sel.changedFrds) > 0 {
			fmt.Printf("  Detected %d modified/added FRD document(s).\n", len(sel.changedFrds))
		}
		if len(sel.changedCsvs) > 0 {
			fmt.Printf("  Detected %d modified/added Card CSV sheet(s).\n", len(sel.changedCsvs))
		}

		if sel.empty() {
			fmt.Println("  No changes detected in docs/frd/, pkg/, or data/cards/.")
			fmt.Println("  Nothing to lint! (Use --all to force full audit across all files)")
			fmt.Println()
			fmt.Println("==========================================================")
			os.Exit(0)
		}
	}

	anyFailed := false

	// 1. FRD Linter
	if sel.frd {
		fmt.Println("\n>>> Running FRD Requirements Linter...")
		files := parsedFiles
		if len(files) == 0 && len(sel.changedFrds) > 0 && !*allFlag {
			files = sel.changedFrds
		}
		ok, err := frd.Lint(repoRoot, files)
		if err != nil || !ok {
			anyFailed = true
		}
	}

	// 2. Executable Rules Traceability Linter
	if sel.rules {
		fmt.Println("\n>>> Running Executable Rules Traceability Linter...")
		ok, err := rules.Lint(repoRoot, nil)
		if err != nil || !ok {
			anyFailed = true
		}
	}

	// 3. Card Sheet CSV Integrity Linter
	if sel.sheets {
		fmt.Println("\n>>> Running Card Sheet CSV Integrity Linter...")
		files := parsedFiles
		if len(files) == 0 && len(sel.changedCsvs) > 0 && !*allFlag {
			files = sel.changedCsvs
		}
		ok, err := cards.Lint(repoRoot, files)
		if err != nil || !ok {
			anyFailed = true
		}
	}

	// 4. Card Rules Syntax & Terminology Linter
	if sel.syntax {
		fmt.Println("\n>>> Running Card Rules Syntax & Terminology Linter...")
		files := parsedFiles
		if len(files) == 0 && len(sel.changedCsvs) > 0 && !*allFlag {
			files = sel.changedCsvs
		}
		ok, err := syntax.Lint(repoRoot, files, true)
		if err != nil || !ok {
			anyFailed = true
		}
	}

	// 5. Engine Architecture Fences Linter
	if sel.fences {
		fmt.Println("\n>>> Running Engine Architecture Fences Linter...")
		ok, err := fences.Lint(repoRoot, nil)
		if err != nil || !ok {
			anyFailed = true
		}
	}
```
Lines 266-282 (the blank line, the summary and `isRulesPath`) stay as they are.

**1c.** After line 17 (`"animus-mundi/pkg/toolchain/linter/syntax"`), add:
```go
	"animus-mundi/pkg/toolchain/specdoc"
```

### 2. `pkg/toolchain/linter/fences/fences.go`
**2a.** Line 198: replace `fmt.Errorf("reading m2_starter_set_ability_survey.md: %w", err)` with `fmt.Errorf("reading %s: %w", SurveyFile, err)`.

**2b.** Line 195: replace `filepath.Join(repoRoot, "docs", "design", "m2_starter_set_ability_survey.md")` with `filepath.Join(repoRoot, "docs", "design", SurveyFile)`.

**2c.** After line 32 (`const modulePath = "animus-mundi"`), add this block, which starts with one blank line (line 33, blank, stays, so one blank line sits on each side):
```go

// SurveyFile is the ability survey's file name, in docs/design. OPEN(...) markers may cite
// the items of its Ruling Questions table.
const SurveyFile = "m2_starter_set_ability_survey.md"
```

### 3. `pkg/toolchain/linter/frd/ownership.go`
**3a.** Line 14: replace `KeywordMatrix = "keyword_interaction_matrix.md"` with `KeywordMatrix = specdoc.MatrixFile`.

**3b.** After line 7 (`"strings"`), add a blank line and then:
```go
	"animus-mundi/pkg/toolchain/specdoc"
```

No other file changes. Run `gofmt -l pkg cmd` after the edits; it must print nothing.

## Allowed Files
- `cmd/cardtown/lint.go`
- `cmd/cardtown/lint_test.go` (tests only, in the `test(T-0197)` commit)
- `pkg/toolchain/linter/fences/fences.go`
- `pkg/toolchain/linter/frd/ownership.go`

## Done When
- [ ] Cases 1-10 hold, with their commands and results in the Log
- [ ] `go tool task sim-compare -- autoloop/m2-wave-d.2` passes (starter and stress sims byte-identical to the run branch)
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-07, lead-architect (wave D.2 run, session 11): wrote this task from board directive #134 against the run branch at f97bea1. No `main` commits to sync. Not `WAVE-COMPLETE`: #134 and #135 are `open: before the next Deliverables task`. Harness, so it skips `tests` and goes straight to `ready`; tier `flash`, because the edits are exact and were dry-run. Three non-test files, within the D2.4 limit. The file-name constants remove the copies T-0196 left to #134 (Inputs), so no new directive is needed. Every command here is on the headless allow-list, one per line, with no pipes, `$`, double quotes or unquoted parentheses. Waiting directives checked: none has triggered (#6, #7: mode is `serial`; #15: no turn-boundary hook; #71, #91 (2), #106, #118 (b), #120, #128: no such card). Next: `kernel-engineer` on Flash.
