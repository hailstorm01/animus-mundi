---
id: T-0232
title: "Free deploys and the look at the top cards move to freedeploy.go (directive #148, wave F.1 harness, flash, pure move)"
kind: harness
tier: flash
wave: F.1
status: done
branch: m2/autoloop-wave-f.1/T-0232-freedeploy-file
depends_on: [T-0231]
claims: [file:pkg/engine/freedeploy.go]
cards: []
requirements: []
---

## Goal
`effects.go` holds one job again: resolving an ability's steps. The free-deploy and look code (`stepDeployFree`, `freeDeployZones`, `lookOutcome`, `stepLookTop`, `combinations`, `permutations`, `zoneAssignments`, `lookLabel`) moves, unchanged, to a new `pkg/engine/freedeploy.go` (#148, structural review 11, finding 5: a mixed file). A pure move: the same declarations, the same test names, the same sim output.

## Inputs
- Board directive #148 (approved by the user, 2026-11-13), `open: before the next Deliverables task`; design note Deliverable 6j (wave F.1), last. It doesn't share T-0231 (the size rule; run file row 5).
- At 0f3ba6d (the run branch after T-0231), `pkg/engine/effects.go` is 666 lines. Lines 117-314 are the eight declarations above with their comments, from `// stepDeployFree puts each card of stepCards, each still in its owner's` to `lookLabel`'s closing `}`; line 315 is the blank line after it, and line 316 starts `// stepSearchDeck`.
- Imports: the moved code uses `fmt`, `slices`, `strings` and `ability`. What stays in `effects.go` uses only `slices` and `ability`, so `fmt` and `strings` leave its import block (the build fails on an unused import otherwise).
- Callers don't change: `runStep` (`effects.go:93-114`) still calls `g.stepDeployFree` and `g.stepLookTop`, and nothing outside `effects.go` names the other six (searched `pkg`).
- The lead dry-ran these exact edits in a throwaway worktree of the run branch at 0f3ba6d: `gofmt -l pkg` clean, `go tool task check` exit 0, `go tool task sim-compare -- autoloop/m2-wave-f.1` `[PASS]`, `cardtown lint fences` exit 0, `go tool task sim` exit 0 (200 games, P0 63.5% / P1 35.0% / draws 1.5%). Every expected output below is from that dry run.

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)
Line numbers are at 0f3ba6d. There is no test commit: no test changes, and the existing tests are the pin. Commit the three files as one implementation commit. Use the file tools to copy: read `effects.go` lines 117-314 and paste them byte for byte, comments included. Change nothing inside the moved block.

### 1. New file `pkg/engine/freedeploy.go`
These 12 lines (line 12 is empty), then `effects.go` lines 117-314 at 0f3ba6d, unchanged, so the file ends with `lookLabel`'s closing `}` and one newline (210 lines):
```go
// Free deploys: deploying a card from a discard pile for free, and looking at the top cards of a deck to deploy some for free and put the rest on the bottom (FR-TRN-003a, FR-MOD-001 §4).

package engine

import (
	"fmt"
	"slices"
	"strings"

	"animus-mundi/pkg/engine/ability"
)

```

### 2. `pkg/engine/effects.go`: remove the moved code
Delete lines 117-315 (the moved block and the blank line after it), so the blank line after `runStep`'s closing `}` (line 116) is followed directly by `// stepSearchDeck has each step player choose a card from their deck, shuffle`.

### 3. `pkg/engine/effects.go`: header and imports
Line 1,
```go
// Effects: resolving a stack item's effect, its steps in order and their values (FR-TGT-003). Untargeted picks are targeting.go's.
```
becomes
```go
// Effects: resolving a stack item's effect, its steps in order and their values (FR-TGT-003). Untargeted picks are targeting.go's; free deploys and the look at the top cards are freedeploy.go's.
```
Lines 5-11,
```go
import (
	"fmt"
	"slices"
	"strings"

	"animus-mundi/pkg/engine/ability"
)
```
become
```go
import (
	"slices"

	"animus-mundi/pkg/engine/ability"
)
```
`effects.go` is then 465 lines.

### 4. `pkg/engine/doc.go`: the Files list
After line 93,
```go
//     an ability's filter. effects.go: running a resolved ability's steps.
```
insert one line:
```go
//     freedeploy.go: free deploys and the look at the top cards.
```

No other file changes.

## Tests
No new tests and no test edits. Every existing test keeps its name and passes unchanged; the free-deploy tests (*Eerie Remnant*, the `FR-TRN-003a` cases), the look tests (*Devilhorn Scouting Party*, SVY-R41-SVY-R43) and the byte-identical sims are the pin.

## Acceptance Cases (harness tasks only; written by the lead)
W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them. Case 4 runs after your implementation commit.
1. **No test changes** (#148): `git -C W diff autoloop/m2-wave-f.1 --stat -- '*_test.go'` prints nothing.
2. **Sims byte-identical** (#148): `go -C W tool task sim-compare -- autoloop/m2-wave-f.1` prints `[PASS] starter and stress sims match autoloop/m2-wave-f.1 byte for byte, exit codes included`.
3. **Only the listed edits:** `git -C W diff autoloop/m2-wave-f.1 --numstat -- pkg cmd` prints exactly three lines: `1	0	pkg/engine/doc.go`, `1	202	pkg/engine/effects.go` and `210	0	pkg/engine/freedeploy.go`.
4. **Byte for byte the dry run** (after the commit): `git -C W rev-parse HEAD:pkg/engine/freedeploy.go HEAD:pkg/engine/effects.go HEAD:pkg/engine/doc.go` prints exactly `591a164190b702d7bb9b4e3dc03cdcdbab6d71ac`, `fc7ace94af7c1b303cd48b77840dae69f2c87699` and `6629e663271fc0b75040196191da0f67d7f97725`, in that order. If one differs, compare that file with Edits 1-4 and fix it in a second commit, so the case holds on the branch tip.
5. **The same declarations, in `freedeploy.go`** (#148): `git -C W grep -n -E '^(func|type) (.g .Game. )?(stepDeployFree|freeDeployZones|lookOutcome|stepLookTop|combinations|permutations|zoneAssignments|lookLabel)[ (]' -- pkg` prints exactly eight lines, all in `pkg/engine/freedeploy.go`: `:21:` `stepDeployFree`, `:55:` `freeDeployZones`, `:74:` `lookOutcome`, `:88:` `stepLookTop`, `:149:` `combinations`, `:163:` `permutations`, `:179:` `zoneAssignments`, `:194:` `lookLabel`.
6. **The headers point there:** `git -C W grep -n -F 'freedeploy.go' -- pkg` prints exactly two lines: `pkg/engine/doc.go:94:` and `pkg/engine/effects.go:1:`.
7. **Fences:** `go -C W run ./cmd/cardtown lint fences` exits 0.
8. **Nothing else changed:** `git -C W diff autoloop/m2-wave-f.1 --stat` lists only `pkg/engine/doc.go`, `pkg/engine/effects.go`, `pkg/engine/freedeploy.go` and this task file.
9. `gofmt -l W/pkg` prints nothing; `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/engine/freedeploy.go` (new)
- `pkg/engine/effects.go`
- `pkg/engine/doc.go`

## Done When
- [x] Cases 1–9 hold, with their commands and results in the Log
- [x] `go tool task sim-compare -- autoloop/m2-wave-f.1` passes (starter and stress sims byte-identical to the run branch)
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-13, lead-architect (wave F.1 run, session 6): written for #148 against 0f3ba6d (T-0231 merged; nothing new on `main` to sync). Harness, so it skips `tests` and goes straight to `ready`. Tier `flash`, because every edit is exact and was dry-run; case 4 pins the three files to the dry run's blobs. Every command here is on the headless allow-list: one per line, no pipes, no `$`, no double quotes, no unquoted parentheses. Waiting directives checked: #6, #7 (mode `serial`), #15 (no turn-boundary hook), #71, #91 (2), #106, #120, #141 (3), #142: not fired. #149–#151 (wave F.2) stay `open: before the next Deliverables task`, after this one. Next: `kernel-engineer` on Flash.
- 2026-11-13, kernel-engineer (T-0232): implemented pure move of free-deploy and look code to pkg/engine/freedeploy.go.
  - Changes: created pkg/engine/freedeploy.go (210 lines) containing stepDeployFree, freeDeployZones, lookOutcome, stepLookTop, combinations, permutations, zoneAssignments, lookLabel unchanged; removed lines 117-315 from pkg/engine/effects.go and removed unused fmt and strings imports; updated headers in pkg/engine/effects.go and pkg/engine/doc.go.
  - Reuse: pure move of existing engine declarations; no new shapes or tests.
  - Acceptance Cases:
    1. git -C W diff autoloop/m2-wave-f.1 --stat -- '*_test.go': clean (no output).
    2. go -C W tool task sim-compare -- autoloop/m2-wave-f.1: [PASS] starter and stress sims match autoloop/m2-wave-f.1 byte for byte, exit codes included.
    3. git -C W diff autoloop/m2-wave-f.1 --numstat -- pkg cmd:
       1	0	pkg/engine/doc.go
       1	202	pkg/engine/effects.go
       210	0	pkg/engine/freedeploy.go
    4. git -C W rev-parse HEAD:pkg/engine/freedeploy.go HEAD:pkg/engine/effects.go HEAD:pkg/engine/doc.go:
       591a164190b702d7bb9b4e3dc03cdcdbab6d71ac
       fc7ace94af7c1b303cd48b77840dae69f2c87699
       6629e663271fc0b75040196191da0f67d7f97725 (byte for byte matches the dry run).
    5. git -C W grep -n -E '^(func|type) (.g .Game. )?(stepDeployFree|freeDeployZones|lookOutcome|stepLookTop|combinations|permutations|zoneAssignments|lookLabel)[ (]' -- pkg: exactly 8 lines in freedeploy.go (lines 21, 55, 74, 88, 149, 163, 179, 194).
    6. git -C W grep -n -F 'freedeploy.go' -- pkg: exactly 2 lines (doc.go:94, effects.go:1).
    7. go -C W run ./cmd/cardtown lint fences: exits 0 ([PASS] Package code is free of card names, invalid markers and disallowed imports).
    8. git -C W diff autoloop/m2-wave-f.1 --stat: lists only doc.go, effects.go, freedeploy.go and task file.
    9. gofmt -l W/pkg: clean; go -C W tool task check: pass (fmt-check, vet, tests, full linter pass); go -C W tool task sim: pass (200 games, P0 63.5% / P1 35.0% / draws 1.5%).
- 2026-11-13, lead-architect (wave F.1 run, session 6, D5): **Approved, `done`.** Merged into `autoloop/m2-wave-f.1` as 8e2624d. Flash run `fc83cd90-0b02-4857-bd38-0afcacc18379`: exit 0, `status: SUCCESS`, no `denied_actions` key, no `permission check failed` in the transcript; `.err` holds only agy's "root agent idle" notice. Review of `git diff autoloop/m2-wave-f.1...m2/autoloop-wave-f.1/T-0232-freedeploy-file` (5307474):
  - Scope: the three Allowed Files and this file. No test file changed (case 1).
  - Pure move, checked against the diff: the three committed blobs are the dry run's (case 4: `591a164`, `fc7ace9`, `6629e66`), so `freedeploy.go` holds `effects.go` lines 117-314 unchanged under its header, and `effects.go` lost exactly those lines, `fmt` and `strings`, with its header pointing to the new file. Numstat `1 0`, `1 202`, `210 0` (case 3).
  - Fences and placement: no card name; the code stays in package `engine`, which owns step resolution, and `doc.go`'s Files list names the new file (case 6). Nothing duplicated: the eight declarations exist once, in `freedeploy.go` (case 5).
  - Rules fidelity: no behavior change; `runStep` still dispatches `OpDeployFree` and `OpLookTop` to the same functions.
  - Re-run by the reviewer in a scratch worktree of the task branch: cases 1, 3, 4 and 6 as written; `gofmt -l pkg` clean; `cardtown lint fences` exit 0; `go tool task check` exit 0; `go tool task sim-compare -- autoloop/m2-wave-f.1` `[PASS]`; `go tool task sim` exit 0, clean, 200 games, P0 63.5% / P1 35.0% / draws 1.5%, P0 50/50 and P1 50/50 main cards executable.
  - The merge's tree matches the task tip's (`git diff` between them is empty), so the gates hold on the run branch. #148 is now `applied` and archived.
