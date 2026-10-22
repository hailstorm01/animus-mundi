---
id: T-0038
title: Fence every package under pkg/, and check the client packages' imports (directive #30, harness, sonnet)
kind: harness
tier: sonnet
wave: none (board directive #30, structural review cleanup 1 of 7, before cardtown play)
status: done
branch: m2/T-0038-fence-all-packages
depends_on: []
claims: [lint:fences-scope, lint:fences-imports, func:isFencesPath]
cards: []
requirements: []
---

## Goal
`cardtown lint fences` checks every non-test Go file under `pkg/` for card names and `OPEN(…)` markers, not only the files under `pkg/engine/`. Two parts of `pkg/` stay exempt: `pkg/engine/cards/` (card definitions) and `pkg/toolchain/` (tooling).

The fence also checks the imports of the rules-core and client packages against a table, so a package that imports outside its allowed direction fails.

The pre-commit trigger runs the fence on a staged change to any of these files. This lets directive #31 move `sim`, `eventlog` and `view` out of `pkg/engine/` without leaving them unfenced.

## Inputs
- Board directive #30 (structural findings 1 and 10).
- Today's scope:
  - [fences.go:181-205](../../../pkg/toolchain/linter/fences/fences.go) walks `pkg/engine/` and skips `cards/`.
  - [lint.go:322-326](../../../cmd/cardtown/lint.go) `isFencesPath` fires only on `pkg/engine/` paths.
- Design note Boundaries: `docs/architecture/m2-ability-model.md` (the import direction table below comes from it).
- Rulings that apply: none.

### Lead decisions
1. **Scope:** walk `<repoRoot>/pkg/`. Skip the directories `pkg/engine/cards` and `pkg/toolchain` with everything under them. Skip `_test.go` files at any depth. Everything else under `pkg/` is fenced, including any new package.
2. **Paths in messages:** violations name the file relative to the repo root, slash-separated (`pkg/engine/x.go:4: …`, `pkg/sim/y.go:3: …`), and stay sorted by path, then line, then message.
3. **Import table**, keyed by package directory, with the in-module import paths each may use. It covers non-test files only. Packages not in the table are not import-checked. Standard-library and third-party imports are never checked. The module path is the constant `animus-mundi`; directive #31 updates the table when it moves the packages.

   | Package directory | May import from `animus-mundi/…` |
   | :--- | :--- |
   | `pkg/engine/ability` | nothing |
   | `pkg/engine/cards` | `pkg/engine/ability` |
   | `pkg/engine/sim` | `pkg/engine` |
   | `pkg/engine/eventlog` | `pkg/engine` |
   | `pkg/engine/view` | `pkg/engine`, `pkg/engine/eventlog` |

   Read imports with `go/parser` in `ImportsOnly` mode, so files whose bodies don't type-check still work. A disallowed import is reported at the import spec's line, using this message format:

   `pkg/engine/sim/x.go:5: imports animus-mundi/pkg/engine/view; pkg/engine/sim may import only pkg/engine`

   A package that may import nothing uses `may import nothing from this module`. A table-package file that fails to parse is reported as `<path>:<line>: cannot read imports: <error>`.

   `pkg/engine/cards` is in the table but outside the card-name scan. Its imports are still checked.
4. **Messages:** the `[PASS]` line becomes `[PASS] Package code is free of card names, invalid markers and disallowed imports.` Update the `Report` and `Analyze` doc comments to the new scope. The header line stays the same.
5. **Trigger:** `isFencesPath` returns true for any `.go` file under `pkg/` except `pkg/toolchain/`. It is also true for `data/cards/*.csv`, the matrix and the survey, as today. Callers pass lowercase slash paths, as today.

## Acceptance Cases (harness tasks only; written by the lead)
Each case is a new test in `pkg/toolchain/linter/fences/fences_test.go` (cases 26–36), built with `setupFixtureRepo`. Add a sibling helper of `writeEngineFile` for files outside `pkg/engine/`. Case 37 goes in a new `cmd/cardtown/lint_test.go`.

26. `pkg/sim/x.go` contains `var _ = "Steel Rain"` on line 3 → fails, with exactly one `[ERROR] pkg/sim/x.go:3: contains card name "Steel Rain"`.
27. Same card name in `pkg/engine/sim/x.go` → fails, naming `pkg/engine/sim/x.go:3`.
28. Same card name in a new `pkg/foo/bar/w.go` → fails, naming `pkg/foo/bar/w.go:3` (new packages are fenced by default).
29. The same card name in each of `pkg/toolchain/lint/t.go`, `pkg/engine/cards/y.go`, `pkg/engine/cards/sub/w.go` and `pkg/sim/x_test.go` → passes.
30. `// OPEN(KIM-Z99)` in `pkg/view/v.go` → fails, saying the item is not in the matrix, naming `pkg/view/v.go`.
31. Violations in `pkg/sim/y.go:3` and `pkg/engine/x.go:3` → both reported, `pkg/engine/x.go:3` first, and the summary counts 2.
32. `pkg/engine/sim/x.go` imports `animus-mundi/pkg/engine/eventlog` → fails, with the exact message from lead decision 3 at the import's line.
33. `pkg/engine/eventlog/x.go` imports `animus-mundi/pkg/engine/view`, and `pkg/engine/ability/x.go` imports `animus-mundi/pkg/engine` → fails with two import errors. The ability one says `may import nothing from this module`.
34. Each of these passes:
    - `pkg/engine/view/x.go` imports `fmt`, `animus-mundi/pkg/engine` and `animus-mundi/pkg/engine/eventlog`;
    - `pkg/engine/cards/x.go` imports `animus-mundi/pkg/engine/ability`;
    - `pkg/engine/sim/x_test.go` imports `animus-mundi/pkg/engine/view` (test files are not import-checked);
    - `pkg/engine/x.go` and `pkg/foo/x.go` import `animus-mundi/pkg/engine/sim` (packages outside the table are not import-checked).
35. `pkg/engine/sim/bad.go` has a malformed import block (`import (`, then `"fmt` with no closing quote) → fails, with one `[ERROR] pkg/engine/sim/bad.go:` line containing `cannot read imports`. `Lint` returns `(false, nil)`, not an error.
36. The `[PASS]` line on a clean fixture is exactly the line in lead decision 4.
37. `isFencesPath` table test:
    - true: `pkg/sim/x.go`, `pkg/engine/x.go`, `pkg/engine/cards/animus.go`, `pkg/foo/bar/x.go`, `pkg/sim/x_test.go`, `data/cards/units.csv`, `docs/frd/keyword_interaction_matrix.md`, `docs/design/m2_starter_set_ability_survey.md`;
    - false: `pkg/toolchain/linter/fences/fences.go`, `cmd/cardtown/main.go`, `pkg/sim/readme.md`, `data/decks/x.txt`.

**Authorized edit to an existing test (in the `test(T-0038)` commit):** `TestFences_Case21_AbilityPackageInScope` expects `[ERROR] pkg/engine/ability/z.go:3` in place of `[ERROR] ability/z.go:3`, in both the check and its failure message. No other existing test changes. Cases 2, 18, 22, 25 and the rest must pass unchanged with the new path format.

Recorded in the Log, not as tests:
- `cardtown lint fences` on the real tree passes, and case 20 still passes.
- `cardtown lint --all` passes.

## Allowed Files
- `pkg/toolchain/linter/fences/fences.go`
- `cmd/cardtown/lint.go` (`isFencesPath` only)
- Tests: `pkg/toolchain/linter/fences/fences_test.go` (new cases, a helper, the case 21 edit), new `cmd/cardtown/lint_test.go`

## Done When
- [x] Cases 26–37 have tests, committed as `test(T-0038): …` before the implementation, and each failed before it
- [x] `go test ./pkg/toolchain/linter/fences ./cmd/cardtown` passes
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)

## Log
- 2026-10-21 lead-architect: Wrote the task from board directive #30, folding in structural finding 10 (the import table). It changes two non-test files. The `ability` and `cards` rows of the import table go beyond the directive's client packages. They cost two rows, and they pin the design note's one-way `engine` → `cards` → `ability` direction, which is otherwise enforced only by Go's cycle rule, and only for `engine`. T-0037 now depends on this task. It has no dependencies of its own, so it is set to `ready`. Branch `m2/T-0038-fence-all-packages` is created from the integration branch. Next: the lead dispatches the `kernel-engineer` subagent (Sonnet).
- 2026-10-21 kernel-engineer: Implemented. Reuse search: the fences walk, `violationItem` sort and `readLines` are kept; `go/parser` ImportsOnly is the only new mechanism (no existing import checker in pkg/). Tests commit c6490ce (cases 26-37 plus the case 21 edit); before the implementation the failing set was cases 21, 26, 27, 28, 30, 31, 32, 33, 35, 36 and `TestIsFencesPath`. Cases 29 and 34 are negative (must pass) cases, so they passed before the change too. Diff: `fences.go` walks `pkg/` skipping `pkg/engine/cards` and `pkg/toolchain`, reports repo-relative paths, adds `allowedImports` table, `checkImports` and `goFilesUnder` (so `pkg/engine/cards` imports are checked though its names are not), new PASS line and doc comments; `lint.go` `isFencesPath` only. The header line was left unchanged per decision 4. A parse failure reports the first scanner error's line. Commands: `go test ./pkg/toolchain/linter/fences ./cmd/cardtown` ok; `go tool task check` exit 0 (includes `cardtown lint --all` and case 20, real tree clean); `go run ./cmd/cardtown lint fences` PASS with the new line. Reviewer: check the cannot-read-imports line number choice.
- 2026-10-21 lead-architect: Review: **approved**, merged as d474cd9, status `done`.
  - Scope: the test commit c6490ce touches only the two test files. The feat commit 26668e8 touches only `fences.go`, `isFencesPath` in `lint.go` and this file, and changes no test. The only edit to an existing test is the authorized case 21 string.
  - Tests: every acceptance case has a test. Re-run at c6490ce, the failing set matches the Log; cases 29 and 34 are must-pass cases. Case 27 failed only on the new path format, which is expected.
  - Gates re-run in a scratch worktree at 26668e8: `go tool task check` exits 0, and `cardtown lint fences` on the real tree prints the new `[PASS]` line. No gameplay changed, so `go tool task sim` was not needed.
  - The line choice for a file that fails to parse (first scanner error, else line 1) is fine.
  - Follow-up, non-blocking: that message repeats the file as an absolute path, because it prints the whole `scanner.Error`. Example: `[ERROR] pkg/engine/sim/zz_bad.go:4: cannot read imports: C:\…\rev38\pkg\engine\sim\zz_bad.go:4:2: string literal not terminated`. It should print `el[0].Msg`. This is folded into directive #36, whose task edits `fences.go`.
  - Process slip: step 1's `in-progress` status commit was skipped. This is harmless in serial mode, but it is the step that shows the task is taken.
