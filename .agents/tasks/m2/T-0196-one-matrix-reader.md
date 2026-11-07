---
id: T-0196
title: One reader of the keyword matrix, and Analyze split by job (directive #133, harness, flash, pure move)
kind: harness
tier: flash
wave: D.2
status: ready
branch: m2/autoloop-wave-d.2/T-0196-one-matrix-reader
depends_on: [T-0195]
claims: [specdoc:ReadMatrix, specdoc:MatrixItem, specdoc:MatrixFile, fences:Analyze-split]
cards: []
requirements: []
---

## Goal
The keyword matrix's Open Items and Rulings tables are parsed in one place, `specdoc.ReadMatrix`, which the fences linter and `whats-up` both call. `fences.Analyze` splits into its four jobs: loading the card names, reading the matrix and survey IDs, scanning the files and checking the imports (#133). A pure move: same test names, unchanged `cardtown lint fences` and `whats-up` output, same sim output.

## Inputs
- Board directive #133 (structural review 9, approved by the user, 2026-11-06).
- Design note: `docs/architecture/m2-ability-model.md`, Deliverable 6h and the `pkg/toolchain` Boundaries row, which this task's commit updates to name the matrix reader.
- The two copies at d83bdcc: `openMatrixItems` in `pkg/toolchain/docs/whatsup.go:194-217` and the matrix loop in `fences.Analyze` (`pkg/toolchain/linter/fences/fences.go:127-157`).
- How the two copies differed, and why the shared reader changes no output: fences trims each line before it matches `## ` and `|`, and needs 3 cells; `openMatrixItems` matches `## ` on the raw line, accepts a row that doesn't start with `|`, and needs 4 cells. The reader takes the fences rules. They give different results only for an indented heading or a row of fewer than 4 cells. The matrix has neither: every heading and row starts in column 0, and every KIM- row has 5 cells. `ReadLines` drops a trailing `\r`, which both copies' `TrimSpace` or prefix checks already ignored.
- The survey's Ruling Questions table is parsed only in `fences` (searched `Ruling Questions` in `pkg` and `cmd`), so its loop moves into `readMarkerIDs` unchanged and stays there.
- Not this task: `frd.KeywordMatrix` (`pkg/toolchain/linter/frd/ownership.go:14`) is the file name the ownership linter exempts, not a parse of the file. The paths in `cmd/cardtown/lint.go:288-289` belong to #134.
- The missing-file error keeps naming the file (`TestFences_Case24_MissingDocsError`): `readMarkerIDs` wraps `ReadMatrix`'s error as `reading keyword_interaction_matrix.md: …`, as before.
- Reuse search (D2.3): no matrix reader in `pkg/toolchain` besides the two copies (searched `KIM-`, `Open Items` and `keyword_interaction` in `pkg` and `cmd`).
- The lead dry-ran these exact edits in a throwaway worktree of the run branch at d83bdcc: `gofmt -l pkg cmd` clean, `go tool task check` and `go tool task sim` pass, and `go tool task sim-compare -- autoloop/m2-wave-d.2` printed `[PASS]`. A throwaway test (not committed) gave `openMatrixItems` the same results before and after on a fixture with three open items (two reachable), a ruled `R` row, KIM- rows outside both tables and CRLF line endings (3 open, 2 reachable), and on a missing file (`ok=false`). Every expected output below is from that dry run.

## Acceptance Cases (harness tasks only; written by the lead)
No new tests: a pure move, so no `test(T-0196)` commit. W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **Existing tests unchanged** (#133): `git -C W diff autoloop/m2-wave-d.2 -- '*_test.go'` prints nothing, and `go -C W test ./...` passes.
2. **Sims byte-identical:** `go -C W tool task sim-compare -- autoloop/m2-wave-d.2` prints `[PASS]`.
3. **`cardtown lint fences` prints what it printed before** (#133): `go -C W run ./cmd/cardtown lint fences` exits 0 and prints exactly these five lines (the second has trailing spaces, as before):
   ```
   ==========================================================
     CARD TOWN LINTER: ENGINE ARCHITECTURE FENCES            
   ==========================================================
     [PASS] Package code is free of card names, invalid markers and disallowed imports.
   ==========================================================
   ```
4. **`whats-up -section state` prints the same matrix line** (#133): `go -C W run ./cmd/cardtown whats-up -section state` exits 0 and prints the line `Keyword interaction matrix: 1 open item(s), 0 reachable with current cards`. Its `Git:` and `Uncommitted changes:` lines depend on the checkout, so compare only this line.
5. **One reader** (#133): `git -C W grep -n -e 'Open Items' -e 'keyword_interaction_matrix' -e '## Rulings' -- pkg/toolchain/docs/whatsup.go pkg/toolchain/linter/fences/fences.go pkg/toolchain/specdoc` prints exactly six lines, all in `pkg/toolchain/specdoc/matrix.go`: `:9`, `:11`, `:14`, `:18`, `:32` and `:33`.
6. **Both callers use it:** `git -C W grep -n -e 'ReadMatrix' -- pkg cmd` prints exactly four lines: `pkg/toolchain/docs/whatsup.go:197`, `pkg/toolchain/linter/fences/fences.go:181`, `pkg/toolchain/specdoc/matrix.go:18` (the comment) and `:19` (the function).
7. **`Analyze` split by job** (#133): `git -C W grep -n -e '^func ' -e '^type ' -- pkg/toolchain/linter/fences/fences.go` prints exactly these lines, in this order: `:47` `type Report`, `:51` `type violationItem`, `:57` `type cardInfo`, `:64` `type markerIDs`, `:71` `func Analyze`, `:112` `func loadCardNames`, `:178` `func readMarkerIDs`, `:237` `func scannedFiles`, `:267` `func scanFiles`, `:346` `func checkAllImports`, `:362` `func Lint`, `:386` `func goFilesUnder`, `:413` `func checkImports`, `:462` `func splitWords`, `:485` `func containsWordSequence`.
8. **Only the listed edits:** `git -C W diff autoloop/m2-wave-d.2 --numstat -- pkg cmd` prints exactly four lines: `6	12	pkg/toolchain/docs/whatsup.go`, `80	47	pkg/toolchain/linter/fences/fences.go`, `51	0	pkg/toolchain/specdoc/matrix.go` and `1	1	pkg/toolchain/specdoc/specdoc.go`.
9. **Nothing else changed:** `git -C W diff autoloop/m2-wave-d.2 --stat` lists only those four files and this task file.
10. `go -C W tool task check` and `go -C W tool task sim` pass.

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)
Line numbers are at d83bdcc, before any edit. In each file, apply the edits from the bottom up (the order below), so the line numbers still hold.

### 1. `pkg/toolchain/specdoc/matrix.go` (new file)
```go
package specdoc

import (
	"path/filepath"
	"strings"
)

// MatrixFile is the keyword interaction matrix's file name, in docs/frd.
const MatrixFile = "keyword_interaction_matrix.md"

// MatrixItem is a row of the keyword matrix's Open Items or Rulings table.
type MatrixItem struct {
	ID    string // "KIM-C01"
	Open  bool   // in "## Open Items"; false for a row of "## Rulings"
	Reach string // an open item's Reach cell: "R" if reachable with current cards, "U" if not
}

// ReadMatrix returns the KIM- rows of the keyword matrix's Open Items and Rulings tables, in file order.
func ReadMatrix(repoRoot string) ([]MatrixItem, error) {
	lines, err := ReadLines(filepath.Join(repoRoot, "docs", "frd", MatrixFile))
	if err != nil {
		return nil, err
	}
	var items []MatrixItem
	section := ""
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "## ") {
			section = trimmed
			continue
		}
		open := strings.HasPrefix(section, "## Open Items")
		if !open && !strings.HasPrefix(section, "## Rulings") || !strings.HasPrefix(trimmed, "|") {
			continue
		}
		cells := strings.Split(trimmed, "|")
		if len(cells) < 3 {
			continue
		}
		id := strings.TrimSpace(cells[1])
		if !strings.HasPrefix(id, "KIM-") {
			continue
		}
		item := MatrixItem{ID: id, Open: open}
		if open && len(cells) > 3 {
			item.Reach = strings.TrimSpace(cells[3])
		}
		items = append(items, item)
	}
	return items, nil
}
```

### 2. `pkg/toolchain/specdoc/specdoc.go`
Replace line 1 with:
```go
// Package specdoc holds what the toolchain shares when it scans a file: requirement IDs, the lines that define them, reading a file as lines, and the keyword matrix's items.
```

### 3. `pkg/toolchain/docs/whatsup.go`
**3a.** Replace lines 194-217 (`openMatrixItems`, its comment included) with:
```go
// openMatrixItems counts the keyword matrix's open items, and how many are reachable ("R").
func openMatrixItems(repoRoot string) (open, reachable int, ok bool) {
	items, err := specdoc.ReadMatrix(repoRoot)
	if err != nil {
		return 0, 0, false
	}
	for _, it := range items {
		if !it.Open {
			continue
		}
		open++
		if it.Reach == "R" {
			reachable++
		}
	}
	return open, reachable, true
}
```
**3b.** After line 13 (`"animus-mundi/pkg/toolchain/linter/rules"`), add:
```go
	"animus-mundi/pkg/toolchain/specdoc"
```

### 4. `pkg/toolchain/linter/fences/fences.go`
**4a.** Replace lines 299-326 (the blank line after the file loop's closing brace, through the closing brace of `Analyze`) with:
```go
	return items, nil
}

// checkAllImports checks the imports of the scanned files and of pkg/engine/cards.
func checkAllImports(repoRoot string, files []string) ([]violationItem, error) {
	// Import direction also applies to the cards package, which the scan skips.
	importFiles := append([]string(nil), files...)
	cardsFiles, err := goFilesUnder(repoRoot, "pkg/engine/cards")
	if err != nil {
		return nil, err
	}
	importFiles = append(importFiles, cardsFiles...)
	var items []violationItem
	for _, name := range importFiles {
		items = append(items, checkImports(repoRoot, name)...)
	}
	return items, nil
}
```
**4b.** On lines 260 and 262, put `ids.` in front of `openSurvey[markerID]` and `ruledSurvey[markerID]`. On lines 244 and 246, put `ids.` in front of `openMatrix[markerID]` and `ruledMatrix[markerID]`. Nothing else on those lines changes.

**4c.** Replace line 224 (the blank line before `var items []violationItem`) with:
```go
	return files, nil
}

// scanFiles reports the card names and invalid OPEN(...) markers in files.
func scanFiles(repoRoot string, files []string, cards []cardInfo, ids *markerIDs) ([]violationItem, error) {
```
**4d.** Replace lines 198-201 (the blank line after the survey loop, through `err = filepath.WalkDir(pkgDir, func(path string, d fs.DirEntry, err error) error {`) with:
```go
	return &markerIDs{openMatrix: openMatrix, ruledMatrix: ruledMatrix, openSurvey: openSurvey, ruledSurvey: ruledSurvey}, nil
}

// scannedFiles lists the non-test Go files under pkg/ that the card-name and marker
// scan reads: all but pkg/engine/cards and pkg/toolchain, slash-separated.
func scannedFiles(repoRoot string) ([]string, error) {
	pkgDir := filepath.Join(repoRoot, "pkg")
	var files []string
	err := filepath.WalkDir(pkgDir, func(path string, d fs.DirEntry, err error) error {
```
**4e.** Replace lines 126-157 (the blank line after the card sort, through the closing brace of the matrix loop) with:
```go
	return cards, nil
}

// readMarkerIDs reads the open and ruled items of the keyword matrix and of the survey's
// Ruling Questions table.
func readMarkerIDs(repoRoot string) (*markerIDs, error) {
	openMatrix := make(map[string]bool)
	ruledMatrix := make(map[string]bool)
	matrix, err := specdoc.ReadMatrix(repoRoot)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", specdoc.MatrixFile, err)
	}
	for _, it := range matrix {
		if it.Open {
			openMatrix[it.ID] = true
		} else {
			ruledMatrix[it.ID] = true
		}
	}
```
Line 158 (blank) and the survey block from line 159 stay as they are.

**4f.** Replace lines 62-65 (the comment on `Analyze` and its `func` line) with:
```go
// markerIDs holds the open and ruled items of the keyword matrix and the survey, which
// OPEN(...) markers are checked against.
type markerIDs struct {
	openMatrix, ruledMatrix, openSurvey, ruledSurvey map[string]bool
}

// Analyze scans non-test Go files under pkg/ (except pkg/engine/cards and pkg/toolchain)
// for card names and invalid markers, and checks the imports of the packages in
// allowedImports (pkg/engine/cards included).
func Analyze(repoRoot string) (*Report, error) {
	cards, err := loadCardNames(repoRoot)
	if err != nil {
		return nil, err
	}
	ids, err := readMarkerIDs(repoRoot)
	if err != nil {
		return nil, err
	}
	files, err := scannedFiles(repoRoot)
	if err != nil {
		return nil, err
	}
	items, err := scanFiles(repoRoot, files, cards, ids)
	if err != nil {
		return nil, err
	}
	importItems, err := checkAllImports(repoRoot, files)
	if err != nil {
		return nil, err
	}
	items = append(items, importItems...)

	sort.Slice(items, func(i, j int) bool {
		if items[i].file != items[j].file {
			return items[i].file < items[j].file
		}
		if items[i].line != items[j].line {
			return items[i].line < items[j].line
		}
		return items[i].msg < items[j].msg
	})

	rep := &Report{}
	for _, it := range items {
		rep.Violations = append(rep.Violations, fmt.Sprintf("%s:%d: %s", it.file, it.line, it.msg))
	}
	return rep, nil
}

// loadCardNames reads every card name in data/cards except the exempt ones, sorted by name.
func loadCardNames(repoRoot string) ([]cardInfo, error) {
```
Line 66 (`cardsDir := filepath.Join(repoRoot, "data", "cards")`) stays as it is. The imports don't change: `os`, `io` and `strings` are still used.

No other file changes.

## Allowed Files
- `pkg/toolchain/specdoc/matrix.go` (new)
- `pkg/toolchain/specdoc/specdoc.go`
- `pkg/toolchain/docs/whatsup.go`
- `pkg/toolchain/linter/fences/fences.go`

## Done When
- [ ] Cases 1-10 hold, with their commands and results in the Log
- [ ] `go tool task sim-compare -- autoloop/m2-wave-d.2` passes (starter and stress sims byte-identical to the run branch)
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-07, lead-architect (wave D.2 run, session 10): wrote this task from board directive #133 against the run branch at d83bdcc. No `main` commits to sync. Not `WAVE-COMPLETE`: #133-#135 are `open: before the next Deliverables task`. Harness, so it skips `tests` and goes straight to `ready`; tier `flash`, because the edits are exact and were dry-run. Four non-test files, within the D2.4 limit. The design note's `pkg/toolchain` Boundaries row now names the matrix reader and its two users (this commit). Every command here is on the headless allow-list, one per line, with no pipes, `$`, double quotes or unquoted parentheses. Waiting directives checked: none has triggered (#6, #7: mode is `serial`; #15: no turn-boundary hook; #71, #91 (2), #106, #118 (b), #120, #128: no such card). Next: `kernel-engineer` on Flash.
