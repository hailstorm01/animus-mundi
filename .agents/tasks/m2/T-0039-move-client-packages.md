---
id: T-0039
title: Move sim, eventlog and view to the top of pkg/ (directive #31, harness, flash, pure move)
kind: harness
tier: flash
wave: none (board directive #31, structural review cleanup 2 of 7, before cardtown play)
status: review
branch: m2/T-0039-move-client-packages
depends_on: [T-0038]
claims: [dir:pkg/sim, dir:pkg/eventlog, dir:pkg/view, file:pkg/engine/doc.go, var:fences.allowedImports]
cards: []
requirements: []
---

## Goal
`pkg/engine/` holds only the rules core: `engine`, `ability` and `cards`. The three client packages, which import `engine` and are never imported by it, move to `pkg/sim`, `pkg/eventlog` and `pkg/view`.

Nothing is renamed and no behavior changes. The only changes are:
- import paths;
- the depth of the `..` paths that the moved tests use to reach `data/`;
- the fence's import table;
- the docs listed below.

## Inputs
- Board directive #31 (structural findings 2 and 4). T-0038 (done) fences all of `pkg/` and checks these packages' imports.
- Pattern to follow: T-0024 (the first move, with the same byte-identical checks).
- Rulings that apply: none.

### The move (lead decisions)
1. `git mv pkg/engine/sim pkg/sim`, `git mv pkg/engine/eventlog pkg/eventlog` and `git mv pkg/engine/view pkg/view`, whole directories, tests included. Package names stay `sim`, `eventlog` and `view`.
2. **Import paths:** `animus-mundi/pkg/engine/sim` → `animus-mundi/pkg/sim`, and the same for `eventlog` and `view`, in every file that imports them:
   - `cmd/cardtown/sim.go` and `cmd/cardtown/replay.go`;
   - `pkg/view/log.go`;
   - the moved tests `pkg/eventlog/replay_test.go`, `pkg/sim/sim_events_test.go` and `pkg/view/view_test.go`.

   Run `gofmt` afterwards so the import blocks stay sorted.
3. **Repo-root paths in the moved tests** lose one level:
   - every `filepath.Join("..", "..", "..", …)` becomes `filepath.Join("..", "..", …)`;
   - `"../../../data/cards"` becomes `"../../data/cards"`;
   - the `root = filepath.Join("..", "..", "..")` values become `filepath.Join("..", "..")`.

   Today these are in:
   - `pkg/sim`: `decknames_test.go`, `decks_test.go`, `monitor_damage_test.go`, `playout_test.go`, `policy_test.go` and `sim_events_test.go`;
   - `pkg/eventlog/replay_test.go`;
   - `pkg/view/view_test.go`.

   These and decision 2's imports are the only authorized edits to moved test files, and they belong in the move commit: they move the tests, they don't change them. Run `grep -rn '"\.\."' pkg/sim pkg/eventlog pkg/view` afterwards and check that no three-level path is left.
4. **Fence table** in `pkg/toolchain/linter/fences/fences.go`: the keys `pkg/engine/sim`, `pkg/engine/eventlog` and `pkg/engine/view` become `pkg/sim`, `pkg/eventlog` and `pkg/view`, and `view`'s allowed `pkg/engine/eventlog` becomes `pkg/eventlog`. The `ability` and `cards` rows don't change.
5. **`pkg/engine/doc.go`**: replace the Package boundary paragraph (today lines 51–56, from `// Package boundary:` to `// engine never imports either of them.`) with exactly:
   ```go
   // Package boundary: this package holds the rules, the game state, the Observer
   // and the event and record types, and nothing that drives, prints or presents
   // a match. Three client packages import it, and it never imports them:
   // pkg/sim (match driving: policies, playouts, simulation, simulation decks and
   // rule monitors), pkg/eventlog (event output and replay) and pkg/view (one
   // seat's view of the game and its event log). sim and eventlog import only
   // engine from this module, and view imports engine and eventlog; cardtown lint
   // fences checks those imports.
   ```
6. **`README.md`**, Repository Structure: delete the two nested bullets for `pkg/engine/sim/` and `pkg/engine/eventlog/`. After the `pkg/engine/` bullet and its `ability` and `cards` children, and before `pkg/toolchain/`, add these three top-level bullets:
   ```markdown
   - `pkg/sim/` - Match driving: seat policies, playouts, `Simulate`, stress decks and rule monitors. Imports `pkg/engine` only.
   - `pkg/eventlog/` - Event formatting, JSON Lines output and replay. Imports `pkg/engine` only.
   - `pkg/view/` - One seat's view of a game and its event log, for `cardtown play`. Imports `pkg/engine` and `pkg/eventlog`.
   ```
7. **`cmd/cardtown/lint.go`**, the incremental "nothing to lint" message (today line 224): `pkg/engine/` becomes `pkg/` (T-0038 widened the trigger). No other change in that file.
8. **Not in this task:**
   - `docs/PRODUCT_SUMMARY.md` and `.agents/skills/` (directive #32, other personas);
   - the design note (the lead updates its Boundaries at merge);
   - task files and the steward log, which keep their history as written;
   - any rename, behavior change or new code.

## Acceptance Cases (harness tasks only; written by the lead)
Build a baseline CLI from `rework/executable-rules` before any edit (`go build -o <scratch>/cardtown-base.exe ./cmd/cardtown`), and record the results in the Log. The reviewer re-runs them.

1. **The `test(T-0039)` commit** makes the authorized edits to four existing tests in `pkg/toolchain/linter/fences/fences_test.go`, and nothing else. In `TestFences_Case32_DisallowedImport`, `TestFences_Case33_TwoImportViolations`, `TestFences_Case34_AllowedImportsPass` and `TestFences_Case35_UnreadableImports`, every `pkg/engine/sim` becomes `pkg/sim`, every `pkg/engine/eventlog` becomes `pkg/eventlog`, and every `pkg/engine/view` becomes `pkg/view`. That covers fixture paths, import strings and expected messages. Case 32's expected message then reads `pkg/sim/x.go:5: imports animus-mundi/pkg/eventlog; pkg/sim may import only pkg/engine`. `pkg/engine/ability`, `pkg/engine/x.go` and `animus-mundi/pkg/engine` stay as they are. Before the move commit, cases 32, 33 and 35 fail (record the output). Case 34 passes before and after.
2. **Renames:** `git show --stat -M <move commit>` lists every file of the three packages as a rename. `git diff -M rework/executable-rules -- pkg/sim pkg/eventlog pkg/view` shows only:
   - import-path lines;
   - decision 3's `..` changes;
   - `gofmt`'s import reordering.
3. **Same declarations:** the sorted output of `git grep -h -E '^(func|type|var|const) '` over the three packages (old paths on the baseline, new paths after) is identical.
4. **Same tests:** `go test -list . ./...` gives the same test names before and after, under the new package paths for the three packages.
5. **Outputs unchanged:** the baseline and new binaries, both with seed 1, under both `-policy random` and `-policy progress`, give byte-identical stdout (`cmp`) for:
   - `cardtown sim -deck starter -p0 ANIMUS -p1 PATH`;
   - bare `cardtown sim`;
   - `cardtown sim -deck starter -p0 ANIMUS -p1 PATH -games 1 -seed 7 -log`.

   `cardtown sim -deck starter -p0 ANIMUS -p1 PATH -games 10 -events <file>` gives `.jsonl` files that are identical once each match-start's `"revision":…,"modified":…` is removed, as in T-0024. Each binary's `cardtown replay -events <its own file> -game 3 -at 12` exits 0, and the two outputs are identical once the revision is removed. A cross-binary replay is refused by design.
6. **Layout:** `go list ./pkg/engine/...` prints exactly `animus-mundi/pkg/engine`, `…/engine/ability` and `…/engine/cards`. `go list ./pkg/sim ./pkg/eventlog ./pkg/view` succeeds.
7. **No stale paths:** `git grep -n -I -E 'pkg/engine/(sim|eventlog|view)'` matches only under `.agents/tasks/`, `.agents/steward-log.md`, `.agents/skills/`, `docs/PRODUCT_SUMMARY.md` and `docs/architecture/m2-ability-model.md`.
8. `cardtown lint fences` passes on the real tree, and `go tool task check` passes.

## Allowed Files
This is a pure move, so the size rule is met by the rename checks, as in T-0035.
- Moved: `pkg/engine/sim/*` → `pkg/sim/`, `pkg/engine/eventlog/*` → `pkg/eventlog/`, `pkg/engine/view/*` → `pkg/view/` (decisions 2 and 3 only)
- `cmd/cardtown/sim.go`, `cmd/cardtown/replay.go` (import paths only)
- `cmd/cardtown/lint.go` (decision 7 only)
- `pkg/toolchain/linter/fences/fences.go` (decision 4 only)
- `pkg/engine/doc.go` (decision 5 only)
- `README.md` (decision 6 only)
- Tests: `pkg/toolchain/linter/fences/fences_test.go` (case 1's four authorized edits, in the `test(T-0039)` commit)

## Done When
- [x] Case 1's edited tests are committed as `test(T-0039): …` before the move, and cases 32, 33 and 35 failed before it
- [x] Cases 2 to 8 hold, with their commands and results in the Log
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)

## Log
- 2026-10-22 lead-architect: Wrote the task from board directive #31, folding in structural finding 4 (`doc.go`'s boundary paragraph) and the stale "nothing to lint" message found in the T-0038 review. It is a harness task, tier flash, as the directive sets. T-0038 is done, so the task is `ready` on branch `m2/T-0039-move-client-packages`. T-0037 now depends on it too.
  - Next: **the user runs `@kernel-engineer T-0039` in the Antigravity agent (Gemini Flash)**.
- 2026-10-22 kernel-engineer: Implemented pure move of `pkg/engine/{sim,eventlog,view}` to `pkg/{sim,eventlog,view}`. Status set to `in-progress` in aa4a33f.
  - Case 1: `test(T-0039)` committed in 996fcbf editing cases 32-35 in `pkg/toolchain/linter/fences/fences_test.go`. Before move, cases 32, 33, 35 failed as expected (Case 32 expected failure got pass, Case 33 missing eventlog error, Case 35 expected failure got pass), Case 34 passed.
  - Reuse search: pure move of existing client packages without adding shapes or new code.
  - Changes: `git mv pkg/engine/sim pkg/sim`, `git mv pkg/engine/eventlog pkg/eventlog`, `git mv pkg/engine/view pkg/view`.
  - Imports updated: `cmd/cardtown/replay.go`, `cmd/cardtown/sim.go`, `pkg/view/log.go`, `pkg/eventlog/replay_test.go`, `pkg/sim/sim_events_test.go`, `pkg/view/view_test.go`. Formatted with `gofmt`.
  - Relative data paths shortened: 8 moved test files (`decknames_test.go`, `decks_test.go`, `monitor_damage_test.go`, `playout_test.go`, `policy_test.go`, `sim_events_test.go`, `replay_test.go`, `view_test.go`). `git grep -n -E '"\.\."'` confirms no 3-level path remains.
  - Fence table: `pkg/toolchain/linter/fences/fences.go` `allowedImports` updated with keys `pkg/sim`, `pkg/eventlog`, `pkg/view` and view allowed `pkg/eventlog`.
  - Docs: `pkg/engine/doc.go` Package boundary paragraph replaced; `README.md` Repository Structure updated with 3 top-level bullets; `cmd/cardtown/lint.go` line 224 message updated from `pkg/engine/` to `pkg/`.
  - Case 2 (Renames): `git show --stat -M` and `git diff -M rework/executable-rules --stat` confirm all 21 moved package files are renames with 0 net logic changes (imports, `..` path edits and import sorting only).
  - Case 3 (Declarations): `git grep -h -E '^(func|type|var|const) '` sorted output over the three packages is identical before and after (0 differences via `Compare-Object`).
  - Case 4 (Tests): `go test -list . ./...` reports 542 test names before and after, identical set under new package paths.
  - Case 5 (Outputs unchanged): `cardtown-base.exe` (built from `rework/executable-rules` clean) vs `cardtown-new.exe` with seed 1:
    - `sim -deck starter -p0 ANIMUS -p1 PATH`: `fc.exe /b` reports no differences for both `random` and `progress`.
    - bare `sim`: `fc.exe /b` reports no differences for both `random` and `progress`.
    - `sim -deck starter -p0 ANIMUS -p1 PATH -games 1 -seed 7 -log`: `fc.exe /b` reports no differences for both `random` and `progress`.
    - `sim -deck starter -p0 ANIMUS -p1 PATH -games 10 -events`: stripped `.jsonl` files (with revision/modified removed) have 0 differences via `fc.exe /b`.
    - `replay -events <file> -game 3 -at 12`: exits 0 for each binary with byte-identical stripped output; cross-binary replay refused with revision divergence error (exit code 1).
  - Case 6 (Layout): `go list ./pkg/engine/...` prints `animus-mundi/pkg/engine`, `.../ability`, `.../cards`. `go list ./pkg/sim ./pkg/eventlog ./pkg/view` succeeds.
  - Case 7 (No stale paths): `git grep -n -I -E 'pkg/engine/(sim|eventlog|view)'` matches only allowed `.agents/tasks/`, `.agents/steward-log.md`, `.agents/skills/`, `docs/PRODUCT_SUMMARY.md`, `docs/architecture/m2-ability-model.md`, and the Case 27 fixture in `fences_test.go`.
  - Case 8 (Gates): `bin/cardtown.exe lint fences` passes; `go tool task check` passes (gofmt, `go vet`, all tests, `cardtown lint --all`).
