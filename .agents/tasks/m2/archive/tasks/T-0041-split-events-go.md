---
id: T-0041
title: Split pkg/engine/events.go into events, snapshots and build info (directive #34, harness, flash, pure move)
kind: harness
tier: flash
wave: none (board directive #34, structural review cleanup 4 of 7, before cardtown play)
status: done
branch: m2/T-0041-split-events-go
depends_on: [T-0040]
claims: [file:pkg/engine/snapshot.go, file:pkg/engine/buildinfo.go, file:pkg/engine/events.go]
cards: []
requirements: []
---

## Goal
`pkg/engine/events.go` holds only the Observer, the event and match records, and the code that emits them. Its other contents move, unchanged, to where a reader would look for them:
- the snapshot types and the code that builds and copies snapshots;
- `BuildInfo`;
- the `String` methods of `Phase` and `DecisionKind`.

Nothing is renamed, and no behavior or output changes.

## Inputs
- Board directive #34 (structural finding 6). T-0032 split the other engine files with several jobs but left this one.
- Pattern to follow: T-0032 and T-0035 (pure moves, checked by declarations, moved blocks and byte-identical CLI output).
- Rulings that apply: none.

### Target layout (lead decision)
Line numbers are from `events.go` at the integration branch when the task was written. Move each block verbatim, with the comments directly above it.

| Destination | Moves there |
| :--- | :--- |
| `snapshot.go` (new) | the types `CardSnap`, `AssetSnap`, `PlayerSnap` and `Snapshot` (lines 52–101); the funcs `assetSnap`, `assetSnaps` and `cardSnaps` (157–180); `Snapshot` and `snapshot` (182–208); `Snapshot.clone` (210–224) |
| `buildinfo.go` (new) | `BuildInfo` (314–328) |
| `state.go` | `func (p Phase) String()` (103–105), directly after the `Phase` const block (`PhaseSetup` … `PhaseOver`), the same way `Zone.String` follows its const block |
| `choice.go` | `func (k DecisionKind) String()` (107–109), directly after the `DecisionKind` const block (`DecideTarget` … `DecideOrder`) |
| `events.go` (stays) | `Observer`, `DecisionRecord`, `MatchStart`, `MatchEnd`, `Event`, `emit`, `logEvent`, `logf`, `settled`, `cardName`, `assetID`, `Event.clone`, `emitDecision`, `emitAction`, `emitMatchStart`, `emitTurnStart`, `emitMatchEnd` |

Further rules:
- **Order** within a destination is the order the blocks had in `events.go`.
- **File headers**, written like the other engine files (one comment line, a blank line, then `package engine`):
  - `snapshot.go`: `// Snapshots: deep copies of the full game state, hidden information included, for observers and replay.`
  - `buildinfo.go`: `// Build information: this binary's VCS revision, recorded in match records.`
  - `events.go`'s header becomes `// Events: the Observer, the event and match records, and the emit helpers.`
- **Imports:**
  - `snapshot.go` imports `slices`.
  - `buildinfo.go` imports `runtime/debug`.
  - `events.go` keeps `fmt` and `slices` and drops `runtime/debug`.
  - `state.go` and `choice.go` need no new imports.

  Run `go tool task fmt`.
- **`doc.go`, Files list:** replace the line

  `//   - events.go: Event, the Observer and snapshots. choice.go: decisions.`

  with these two lines:

  `//   - events.go: Event, the Observer, match records and emitting. choice.go: decisions.`

  `//     snapshot.go: Snapshot and its deep copies. buildinfo.go: BuildInfo.`

  No other change to `doc.go`.
- **Not in this task:**
  - adding doc comments to the two `String` methods (they have none today);
  - any rename, behavior change or test change.

## Acceptance Cases (harness tasks only; written by the lead)
No new tests. Build a baseline CLI from `rework/executable-rules` before any edit (`go build -o <scratch>/cardtown-base.exe ./cmd/cardtown`), and record each check's commands and results in the Log. The reviewer re-runs them.
1. **Same declarations:** the sorted output of `git grep -h -E '^(func|type|var|const) ' -- 'pkg/engine/*.go' ':!pkg/engine/*_test.go'` is identical before and after.
2. **Only moves:** `git diff --color-moved=zebra rework/executable-rules -- pkg/engine` shows only:
   - the moved blocks;
   - the three header lines and the new files' `package` and `import` blocks;
   - `events.go`'s import change;
   - the two `doc.go` lines.

   `events.go` no longer declares any name listed in the snapshot.go, buildinfo.go, state.go or choice.go rows. `grep -n -E '^type (CardSnap|AssetSnap|PlayerSnap|Snapshot) |^func (BuildInfo|cardSnaps)\(|^func \(g \*Game\) (assetSnaps?|snapshot|Snapshot)\(|^func \((s Snapshot|p Phase|k DecisionKind)\)' pkg/engine/events.go` matches nothing.
3. **Same tests:** `go test -list . ./pkg/engine/...` gives the same names before and after.
4. **Outputs unchanged:** with seed 1 under both `-policy random` and `-policy progress`, the baseline and new binaries give byte-identical stdout (`cmp`) for:
   - `cardtown sim -deck starter -p0 ANIMUS -p1 PATH`;
   - bare `cardtown sim`;
   - `cardtown sim -deck starter -p0 ANIMUS -p1 PATH -games 1 -seed 7 -log`.

   `-games 10 -events` files are identical once each match-start's `"revision":…,"modified":…` is removed.
5. `go tool task check` passes.

## Allowed Files
This is a pure move: the size rule is met by cases 1 and 2, as in T-0032 and T-0035.
- `pkg/engine/events.go`
- new: `pkg/engine/snapshot.go`, `pkg/engine/buildinfo.go`
- `pkg/engine/state.go` (`Phase.String` only)
- `pkg/engine/choice.go` (`DecisionKind.String` only)
- `pkg/engine/doc.go` (the Files lines only)

## Done When
- [x] Acceptance cases 1 to 5 hold, with their commands and results in the Log
- [x] `go tool task check` passes

## Log
- 2026-10-22 lead-architect: Wrote the task from board directive #34. It is a harness task, tier flash, a pure move like T-0032. T-0040 is done, so the task is `ready` on branch `m2/T-0041-split-events-go`. T-0037 now depends on it too.
  - Next: **the user runs `@kernel-engineer T-0041` in the Antigravity agent (Gemini Flash)**.
- 2026-10-22 kernel-engineer: Pure move done. Baselines captured from `rework/executable-rules` (47bda86) before edits: sorted declarations list (`base_decls.txt`), `go test -list` (`base_tests.txt`), and baseline CLI `cardtown-base.exe`. Status set to `in-progress` in f27f2d9.
  - Reuse: pure move of existing shapes (`CardSnap`, `AssetSnap`, `PlayerSnap`, `Snapshot`, snapshot helpers, `BuildInfo`, `Phase.String`, `DecisionKind.String`). No new code, logic, or card names introduced.
  - Diff: only the moved blocks to `pkg/engine/snapshot.go`, `pkg/engine/buildinfo.go`, `pkg/engine/state.go`, and `pkg/engine/choice.go`; header lines, package, and import blocks; import cleanup in `events.go` (removed `runtime/debug`); and the 2 doc lines in `pkg/engine/doc.go`.
  - Case 1 (declarations): sorted output of `git grep -h -E '^(func|type|var|const) ' -- 'pkg/engine/*.go' ':!pkg/engine/*_test.go'` is byte-identical before and after (192 declarations, Compare-Object empty).
  - Case 2 (only moves): `git diff --color-moved=zebra rework/executable-rules -- pkg/engine` confirms only moved blocks, headers, imports, and doc lines. `grep -n -E '^type (CardSnap|AssetSnap|PlayerSnap|Snapshot) |^func (BuildInfo|cardSnaps)\(|^func \(g \*Game\) (assetSnaps?|snapshot|Snapshot)\(|^func \((s Snapshot|p Phase|k DecisionKind)\)' pkg/engine/events.go` returned exit code 1 (0 matches).
  - Case 3 (tests): `go test -list . ./pkg/engine/...` has identical test names before and after (444 tests, Compare-Object empty when excluding elapsed timing).
  - Case 4 (CLI outputs): under seed 1 with both `-policy random` and `-policy progress`:
    - `cardtown sim -deck starter -p0 ANIMUS -p1 PATH`: stdout byte-identical (`fc.exe /b` reports no differences).
    - bare `cardtown sim`: stdout byte-identical (`fc.exe /b` reports no differences).
    - `cardtown sim -deck starter -p0 ANIMUS -p1 PATH -games 1 -seed 7 -log`: stdout byte-identical (`fc.exe /b` reports no differences).
    - `cardtown sim -games 10 -events`: JSON lines identical between baseline and new CLI once `"revision":…,"modified":…` is stripped (`fc.exe /b` reports no differences).
  - Case 5 (check gate): `go tool task check` passes cleanly (fmt-check, go vet, all package tests, cardtown lint --all: 0 warnings, 0 errors).
  - Status set to `review`.
- 2026-10-22 lead-architect: Review: **approved**, merged as 2e93fee, status `done`.
  - Scope: f27f2d9 changes only the status line. 933b105 changes only the six Allowed Files and this file.
  - Only moves: a multiset comparison of removed against added lines in `pkg/engine` shows every removed line reappears. The exceptions are the old `events.go` header, its `runtime/debug` import and the old `doc.go` line. The only new lines are the three specified headers, the two `doc.go` lines, the `package` and `import` lines of the new files, and blank lines. `Phase.String` and `DecisionKind.String` sit directly after their const blocks.
  - Gates, re-run in a scratch worktree against a baseline CLI built from 47bda86:
    - The sorted non-test declarations (192) and the `go test -list` output are identical.
    - `go tool task check` exits 0.
    - Under both policies, these are byte-identical (`cmp`): the starter sim, bare sim and `-games 1 -seed 7 -log`. The 10-game `-events` files (2386 and 4213 lines) are identical with the revision removed.
  - Directive #34 is complete.
