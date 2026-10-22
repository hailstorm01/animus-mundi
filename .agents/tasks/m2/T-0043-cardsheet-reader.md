---
id: T-0043
title: One card-sheet reader for the sheets, syntax and fences linters (directive #36, harness, sonnet)
kind: harness
tier: sonnet
wave: none (board directive #36, structural review cleanup 6 of 7, before cardtown play)
status: in-progress
branch: m2/T-0043-cardsheet-reader
depends_on: [T-0042]
claims: [pkg:toolchain/linter/cardsheet, func:cardsheet.List, func:cardsheet.Open, func:cardsheet.Column]
cards: []
requirements: []
---

## Goal
Three linters each find, open and header-parse the `data/cards` CSV sheets with their own copy of the code: `linter/cards` (`cardtown lint sheets`), `linter/syntax` and `linter/fences`. After this task, they share one small package, `pkg/toolchain/linter/cardsheet`. Every linter's output stays byte-identical.

The task also folds in the T-0038 review's follow-up. A fenced Go file whose imports cannot be parsed is reported without repeating its absolute path.

## Inputs
- Board directive #36 (structural finding 8), including its acceptance case (d).
- The copies:
  - [sheets.go:68-125](../../../pkg/toolchain/linter/cards/sheets.go) and [syntax.go:34-94](../../../pkg/toolchain/linter/syntax/syntax.go): the file listing is the same code, line for line;
  - [fences.go:71-100](../../../pkg/toolchain/linter/fences/fences.go): a third reader.
- `pkg/engine/cardload.go` keeps its own reader, so that engine and toolchain stay independent of each other. It is out of scope.
- Rulings that apply: none.

### Lead decisions
The linters differ in ways their output shows:
- `sheets` checks the header before reading any rows, and skips the rows when the header is wrong.
- `sheets` and `syntax` require a fixed field count. `fences` accepts ragged rows.
- Each linter prints its own message for an open, empty-file, header or rows failure.

So the package is thin: it shares the listing and the column lookup, and hands back the stages separately. Callers keep every message and every decision.

1. **`List(repoRoot string, paths []string) (files, missing []string, err error)`**, which replaces the shared listing block.
   - **With `paths`:** for each one, trimmed, skipping blanks: a relative path is joined to `repoRoot`. If `os.Stat` succeeds, the path goes in `files`. Otherwise the trimmed input goes in `missing`, in input order.
   - **Without `paths`:** it lists the non-directory `*.csv` entries of `<repoRoot>/data/cards` with `os.ReadDir`. A read error is returned as `fmt.Errorf("reading cards directory: %w", err)`.
   - `files` is sorted.
   - **Callers:**
     - `sheets` prints `[ERROR] Specified file not found: <name>` for each missing path and sets its error flag.
     - `syntax` prints the same line but sets no flag, as today.
     - `fences` keeps its own `os.Stat` checks on `data/cards` with their current messages, then calls `List(repoRoot, nil)` in place of `filepath.Glob`.
2. **`Open(path string, ragged bool) (*Sheet, error)`**: opens the file. Its error is the `os.Open` error, unwrapped. With `ragged`, the reader sets `FieldsPerRecord = -1`. `Sheet` has these methods:
   - `Header() ([]string, error)`: the first record, as `csv.Reader.Read` returns it, so an empty file gives `io.EOF`. Header cells are not trimmed or changed.
   - `Rows() ([][]string, error)`: the remaining records, via `ReadAll`.
   - `Close() error`.

   Callers keep their current order of calls and their messages:
   - `sheets`: Open, Header (with the `io.EOF` "File is empty!" branch), header checks, Rows;
   - `syntax`: Open, Header, Rows;
   - `fences`: Open with `ragged` true, Header and Rows. A file whose Header is `io.EOF` is skipped, as today's `len(rows) == 0` branch does.
3. **`Column(header []string, names ...string) int`**: the index of the first header cell that matches any of `names`. Before comparing, the cell is trimmed of spaces and of a leading `﻿` byte-order mark, and the comparison ignores case. It returns -1 when nothing matches.
   - `fences` uses `Column(h, "name")`.
   - `syntax` uses `Column(h, "name")` and `Column(h, "rulestext", "text")`.
   - Today `syntax` takes the *last* match when a header repeats. A repeated header is already a `sheets` error, and the real sheets have none, so first-match is accepted.
4. **Package doc** at the top of `cardsheet.go`: `// Package cardsheet finds and reads the data/cards CSV sheets for the toolchain linters. pkg/engine has its own loader.`
5. **Fence follow-up (directive #36 (d)):** in `checkImports` in `fences.go`, a parse failure reports `cannot read imports: <msg>`, where `<msg>` is the first `scanner.Error`'s `Msg`, not its `Error()`, which repeats the absolute path. A non-`ErrorList` error keeps `err.Error()`.

## Acceptance Cases (harness tasks only; written by the lead)
Cases 1–7 are unit tests in a new `pkg/toolchain/linter/cardsheet/cardsheet_test.go`, built on `t.TempDir()`. Case 8 is a new test in `fences_test.go`. Cases 9–11 are checks recorded in the Log against a baseline CLI built from `rework/executable-rules` before any edit; the reviewer re-runs them.
1. `List(root, nil)`, where `data/cards` holds `b.csv`, `a.csv`, `notes.txt` and a directory `sub.csv/` → `files` is `[<root>/data/cards/a.csv, <root>/data/cards/b.csv]`, `missing` is empty, `err` is nil.
2. `List(root, []string{"data/cards/b.csv", "  ", "data/cards/none.csv", <absolute path of a.csv>})` → `files` is the sorted absolute paths of `a.csv` and `b.csv`, and `missing` is `["data/cards/none.csv"]`.
3. `List(root, nil)` with no `data/cards` directory → an error whose text contains `reading cards directory`.
4. `Open` on a missing path → `errors.Is(err, fs.ErrNotExist)`.
5. `Open(emptyFile, false)` then `Header()` → `io.EOF`.
6. On `name,power\nA,1\nB\n`: with `ragged` false, `Header()` succeeds and `Rows()` returns an error. With `ragged` true, `Rows()` returns `[[A 1] [B]]`.
7. `Column([]string{"﻿Name", " rulesText ", "power"}, ...)`:
   - `"name"` → 0;
   - `"text", "rulestext"` → 1;
   - `"POWER"` → 2;
   - `"health"` → -1.
8. `TestFences_Case38_UnreadableImportsMessage`: the fixture of case 35 (`pkg/sim/bad.go` with `import (`, then `"fmt` unterminated) → the output contains exactly the line `  [ERROR] pkg/sim/bad.go:4: cannot read imports: string literal not terminated`.
9. **Outputs unchanged on the real tree:** stdout and exit code are byte-identical to the baseline for `cardtown lint sheets`, `lint syntax`, `lint fences`, `lint cards` and `lint --all`.
10. **Outputs unchanged on edge inputs:** stdout and exit code are byte-identical to the baseline for `cardtown lint sheets` and `cardtown lint syntax`, each given these four paths:
    - one real sheet (`data/cards/` + any one file there);
    - `data/cards/nope.csv` (missing);
    - an absolute path to an empty file in a scratch folder;
    - an absolute path to a header-only copy of a real sheet in a scratch folder.
11. Outside tests, `csv.NewReader` appears in `pkg/toolchain` only in `cardsheet` (`git grep -n 'csv.NewReader' -- pkg/toolchain ':!*_test.go'`). `filepath.Glob` no longer appears in `fences.go`.

## Allowed Files
- new: `pkg/toolchain/linter/cardsheet/cardsheet.go`
- `pkg/toolchain/linter/cards/sheets.go`
- `pkg/toolchain/linter/syntax/syntax.go`
- `pkg/toolchain/linter/fences/fences.go`
- Tests: new `pkg/toolchain/linter/cardsheet/cardsheet_test.go`; `pkg/toolchain/linter/fences/fences_test.go` (case 8 only, added)

## Done When
- [ ] Cases 1–8 have tests, committed as `test(T-0043): …` before the implementation, and each failed before it (the `cardsheet` tests fail to compile until the package exists, which counts)
- [ ] Cases 9–11 hold, with their commands and results in the Log
- [ ] `go test ./pkg/toolchain/...` passes
- [ ] `go tool task check` passes

## Log
- 2026-10-22 lead-architect: Wrote the task from board directive #36, folding in its case (d), the T-0038 follow-up. The API is deliberately thin (Lead decisions), because the three linters differ in field-count rules, call order and messages, and their output must not change. The shared parts are the listing, which is copied line for line between `sheets` and `syntax`, and the column lookup. It is a harness task, tier sonnet, as the directive sets. T-0042 is done, so the task is `ready` on branch `m2/T-0043-cardsheet-reader`. T-0037 now depends on it too.
  - Next: the lead dispatches the `kernel-engineer` subagent (Sonnet).
