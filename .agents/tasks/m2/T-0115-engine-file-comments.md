---
id: T-0115
title: File comments on damage.go, choice.go, invariants.go and layers.go, so every engine file says its job above the package line (directive #87 (c), harness, flash, comments only)
kind: harness
tier: flash
wave: G.2 (board directive #87 (c), structural review 6, finding 6)
status: ready
branch: m2/T-0115-engine-file-comments
depends_on: [T-0114]
claims: []
cards: []
requirements: []
---

## Goal
Four non-test files in `pkg/engine` have no file comment above their `package` line: `damage.go` (it lost its job description when #79 moved the sweep to `sweep.go`), `choice.go`, `invariants.go`, and `layers.go`, whose comment sits below the `package` line. Each gets a one-line file comment above `package engine` in the style of the other engine files (`// <Topic>: <its job> (<FR IDs>).`, a blank line, then `package engine`). Comments only: same code, same tests, same sim output. After this task directive #87 is `applied`, the last `open: before` directive of wave G.2.

## Inputs
- Board directive #87 (structural review 6, approved by the user, 2026-10-30). Parts (a) and (b) were applied in the design note when #82–#87 were recorded; this task is part (c) only.
- The style: every other non-test file in `pkg/engine` starts with a `//` comment line, a blank line, then `package engine`. For example `sweep.go`: `// Death sweep: destroying lethal assets, replaced deaths first, and the match-loss checks (FR-LAY-002, FR-WIN-001).` `doc.go` is the package doc and is not touched.
- The blank line between the file comment and `package engine` matters: without it the comment would become a second package doc comment and show in `go doc ./pkg/engine`.
- Why the FR IDs: each one is already cited in its file (`FR-REP-004` in `dealDamage`'s comment, `FR-MOD-001 §6` in `DecisionKind`'s and `choose`'s, `FR-LAY-001` in `layers.go`'s existing comment). `invariants.go` cites none, and its comment cites none.
- The lead dry-ran these exact edits on the run branch at 122b2c6: `gofmt -l pkg/engine` printed nothing, case 1's `git grep` printed 24 lines with no `:1:`, `go tool task check` passed, and `go tool task sim-compare -- autoloop/m2-wave-g.2` printed `[PASS]`.

### Where it is today (at `autoloop/m2-wave-g.2` 122b2c6)
| File | Lines 1–5 today |
| :--- | :--- |
| `pkg/engine/damage.go` | `package engine`, blank, `// DamageResult records what each mitigation stage did to one damage instance.` |
| `pkg/engine/choice.go` | `package engine`, blank, `import (` |
| `pkg/engine/invariants.go` | `package engine`, blank, `import "fmt"` |
| `pkg/engine/layers.go` | `package engine`, blank, then a three-line comment starting `// Effective characteristics through the six-layer pipeline (FR-LAY-001).`, blank, `// Power returns …` |

## Edits (lead decisions; copy exactly)
Nothing else changes in these four files, and no other file changes except this task file.

### 1. `pkg/engine/damage.go`: add two lines at the top
```go
// Damage: the five-stage mitigation pipeline for one damage instance (FR-REP-004).

package engine
```

### 2. `pkg/engine/choice.go`: add two lines at the top
```go
// Choices: the decisions put to a player, the Chooser that answers them, and the fallback for a missing, timed-out or invalid answer (FR-MOD-001 §6).

package engine
```
(`§` is the UTF-8 section sign, as in `DecisionKind`'s comment on line 8 of the file today.)

### 3. `pkg/engine/invariants.go`: add two lines at the top
```go
// Invariants: the state properties CheckInvariants verifies between actions.

package engine
```

### 4. `pkg/engine/layers.go`: move the comment above the `package` line
The three-line comment moves above `package engine`, and its first line gains the `Layers:` topic. Lines 2 and 3 of the comment are unchanged. The file's first eight lines become exactly:
```go
// Layers: effective characteristics through the six-layer pipeline (FR-LAY-001).
// Nothing grants types (Layer 2) or keywords (Layer 3), and there are no auras
// (Layer 4) yet, so the pipeline reduces to Layer 1 + Layer 5 + Layer 6 clamping.

package engine

// Power returns effective Combat Power, the only stat that clamps at 0.
func (g *Game) Power(a *Asset) int {
```

## Acceptance Cases (harness tasks only; written by the lead)
No new tests: comments only, and the unchanged tests and `sim-compare` pin that no code changed. Run each command on its own, pointed at your worktree (W). Record each command and its result in the Log. The reviewer re-runs them.
1. **Every engine file has a file comment** (#87 (a)): `git -C W grep -n -E -e "^package " -- ':(glob)pkg/engine/*.go' ':!*_test.go'` prints 24 lines, and none of them has line number 1 (no line reads `pkg/engine/<file>.go:1:package engine`). `damage.go`, `choice.go` and `invariants.go` show `:3:`, `layers.go` shows `:5:`, and `doc.go` shows `:89:`. Before the edits, those four files show `:1:`.
2. **The four comments are there:** `git -C W grep -n -e "^// Damage: " -e "^// Choices: " -e "^// Invariants: " -e "^// Layers: " -- pkg/engine` prints exactly four lines, each at line 1: `pkg/engine/choice.go:1:`, `pkg/engine/damage.go:1:`, `pkg/engine/invariants.go:1:` and `pkg/engine/layers.go:1:`.
3. **Comments only:** `git -C W diff autoloop/m2-wave-g.2 --numstat -- pkg` prints exactly these four lines: `2	0	pkg/engine/choice.go`, `2	0	pkg/engine/damage.go`, `2	0	pkg/engine/invariants.go` and `3	3	pkg/engine/layers.go`.
4. **The package doc is unchanged:** `go -C W doc ./pkg/engine` begins `package engine // import "animus-mundi/pkg/engine"`, a blank line, then `Package engine is the executable form of the Game Rules FRD`. None of the four new comments appears in its output.
5. **Existing tests unchanged:** `git -C W diff autoloop/m2-wave-g.2 -- '*_test.go'` prints nothing, and `go -C W test ./...` passes.
6. **Sims byte-identical** (#87 (b)): `go -C W tool task sim-compare -- autoloop/m2-wave-g.2` prints `[PASS]`.
7. **Only these edits:** `git -C W diff autoloop/m2-wave-g.2 --stat` lists only the four files in case 3 and this task file.
8. `go -C W tool task check` (#87 (c)) and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/engine/damage.go`
- `pkg/engine/choice.go`
- `pkg/engine/invariants.go`
- `pkg/engine/layers.go`

## Done When
- [ ] Cases 1–8 hold, with their commands and results in the Log
- [ ] `go tool task sim-compare -- autoloop/m2-wave-g.2` passes (starter and stress sims byte-identical to the run branch)
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks

## Log
- 2026-10-30, lead-architect (wave G.2 run, session 9): wrote the task from board directive #87 (c), the last `open: before` directive of wave G.2; T-0114 is `done` (merge 8375cdf) and #86 is `applied`. Harness, so it skips `tests` and goes straight to `ready`; tier `flash` as #87 gives. No new test: comments only, pinned by the unchanged tests and `sim-compare`. Every command here is on the headless allow-list, one per line, with no `|`, `$` or unquoted parentheses. The edits were dry-run on the run branch (Inputs). `main` (37d9c89) has no commits the run branch lacks, so no sync. Waiting directives checked: none has triggered (#6, #7: mode is `serial`; #8: no `deck.go`; #15, #20 (b), #58, #59 (2), #61, #71: no such hook, claim or card here). When this task is `done`, #87 is `applied` and wave G.2's directives #82 to #87 are all applied. Next: `kernel-engineer` on Flash.
