---
id: T-0189
title: Effective reads live in `layers.go`, `cornered`, `EffectiveState` and `effectiveState` move there (directive #126, harness, flash, pure move)
kind: harness
tier: flash
wave: D.2
status: done
branch: m2/autoloop-wave-d.2/T-0189-effective-reads-in-layers
depends_on: [T-0188]
claims: [file-move:cornered, file-move:EffectiveState, file-move:effectiveState]
cards: []
requirements: []
---

## Goal
Every read of an asset's effective characteristics sits in `pkg/engine/layers.go` before wave E adds more Layer 3 and Layer 4 reads (#126). `cornered` moves from `trigger.go`, and `EffectiveState` and `effectiveState` move from `state.go`. A pure move: same declarations, same test names, same sim output.

## Inputs
- Board directive #126 (structural review 9, approved by the user, 2026-11-06).
- Design note: `docs/architecture/m2-ability-model.md`, Deliverable 6h.
- Why it is a pure move: the three declarations keep their names, signatures, comments and bodies, in the same package. Neither source file loses its last use of an import.
- `doc.go` names no moved declaration in its file list, but its one line on `layers.go` would no longer cover the file, so Edit 3 extends it. That is the update #126 asks for.
- Not in this task: `lastKnownRecord`, `LastKnown` and `LastKnownEffective` stay in `state.go`, because #126 doesn't name them and `state.go`'s header claims last-known information.
- Reuse search (D2.3): no second declaration of any of the three exists in `pkg/engine`. This task adds no new code.
- The lead dry-ran these exact edits in a throwaway worktree of the run branch at 2d71351: `gofmt -l` clean, `go tool task check` and `go tool task sim` pass, and `go tool task sim-compare -- autoloop/m2-wave-d.2` printed `[PASS]`. Every expected output below is from that dry run.

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)
Line numbers are at 2d71351, before any edit.

### 1. `pkg/engine/state.go`: remove `EffectiveState` and `effectiveState`
Delete lines 158-170: from `// EffectiveState is an asset's effective characteristics at one moment, read` through the blank line after `effectiveState`'s closing brace. `// lastKnownRecord is an asset as it left the battlefield` then follows `firingKey`'s closing brace and one blank line.

### 2. `pkg/engine/trigger.go`: remove `cornered`
Delete lines 125-131: from `// cornered reports whether p is Cornered` through the blank line after its closing brace. `// fireAttacks buffers` then follows `bufferTrigger`'s closing brace and one blank line.

### 3. `pkg/engine/doc.go`: the file list
Line 88 today ends `layers.go: derived power, health and keywords.`. Change that ending to `layers.go: derived power, health and keywords,`, and add this new line right after it:
```go
//     the Cornered check and EffectiveState.
```

### 4. `pkg/engine/layers.go`: add the three
Replace the `IsLethal` declaration at the end of the file (lines 37-40) with:
```go
// cornered reports whether p is Cornered: the opponent controls strictly more
// battlefield assets (units and structures) in the Main Theater than p does.
// Headquarters are never in the Main Theater (FR-KWD-002e).
func (g *Game) cornered(p PlayerID) bool {
	return len(g.TheaterOf(p.Opponent())) > len(g.TheaterOf(p))
}

// IsLethal reports whether the next death sweep destroys a (FR-LAY-002).
func (g *Game) IsLethal(a *Asset) bool {
	return g.MaxHealth(a) <= 0 || g.CurrentHealth(a) <= 0
}

// EffectiveState is an asset's effective characteristics at one moment, read
// through the layer pipeline (FR-LAY-001).
type EffectiveState struct {
	Power     int
	MaxHealth int
	Keywords  Keywords
}

// effectiveState reads a's effective characteristics now.
func (g *Game) effectiveState(a *Asset) EffectiveState {
	return EffectiveState{Power: g.Power(a), MaxHealth: g.MaxHealth(a), Keywords: g.Keywords(a)}
}
```

No other file changes.

## Acceptance Cases (harness tasks only; written by the lead)
No new tests: a pure move, so no `test(T-0189)` commit. W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **Existing tests unchanged** (#126): `git -C W diff autoloop/m2-wave-d.2 -- '*_test.go'` prints nothing, and `go -C W test ./...` passes.
2. **Sims byte-identical** (#126): `go -C W tool task sim-compare -- autoloop/m2-wave-d.2` prints `[PASS]`.
3. **Each declaration once, in `layers.go`:** `git -C W grep -n -e 'func (g [*]Game) cornered(' -e 'func (g [*]Game) effectiveState(' -e 'type EffectiveState struct' -- pkg/engine` prints exactly three lines: `pkg/engine/layers.go` lines 40, 51 and 58.
4. **Only the listed edits:** `git -C W diff autoloop/m2-wave-d.2 --numstat -- pkg cmd` prints exactly four lines: `2	1	pkg/engine/doc.go`, `20	0	pkg/engine/layers.go`, `0	13	pkg/engine/state.go` and `0	7	pkg/engine/trigger.go`.
5. **Fences:** `go -C W run ./cmd/cardtown lint fences` exits 0.
6. **Nothing else changed:** `git -C W diff autoloop/m2-wave-d.2 --stat` lists only those four files and this task file.
7. `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/engine/doc.go`
- `pkg/engine/layers.go`
- `pkg/engine/state.go`
- `pkg/engine/trigger.go`

## Done When
- [x] Cases 1-7 hold, with their commands and results in the Log
- [x] `go tool task sim-compare -- autoloop/m2-wave-d.2` passes (starter and stress sims byte-identical to the run branch)
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-07, lead-architect (wave D.2 run, session 3): wrote this task from board directive #126, next in board order, against the run branch at 2d71351. No `main` commits to sync. Not `WAVE-COMPLETE`: #126, #127 and #129-#135 are `open: before the next Deliverables task`. Harness, so it skips `tests` and goes straight to `ready`; tier `flash`, because the edits are exact and were dry-run. Every command here is on the headless allow-list, one per line, with no pipes, `$`, double quotes or unquoted parentheses. Waiting directives checked: none has triggered (#6, #7: mode is `serial`; #15: no turn-boundary hook; #71, #91 (2), #106, #118 (b), #120, #128: no such card). Next: `kernel-engineer` on Flash.
- 2026-11-07, kernel-engineer: pure move implementation complete per directive #126.
  - What changed:
    - `pkg/engine/state.go`: removed `EffectiveState` and `effectiveState` (lines 158-170).
    - `pkg/engine/trigger.go`: removed `cornered` (lines 125-131).
    - `pkg/engine/doc.go`: updated `layers.go` description to mention the Cornered check and `EffectiveState`.
    - `pkg/engine/layers.go`: added `cornered`, `EffectiveState`, and `effectiveState` alongside `IsLethal`.
  - Reuse: verified no other declaration of `cornered`, `EffectiveState`, or `effectiveState` exists in `pkg/engine`.
  - Commands run and results:
    - Case 1: `git -C W diff autoloop/m2-wave-d.2 -- '*_test.go'` returned no output; `go -C W test ./...` passed (exit 0).
    - Case 2: `go -C W tool task sim-compare -- autoloop/m2-wave-d.2` printed `[PASS] starter and stress sims match autoloop/m2-wave-d.2 byte for byte, exit codes included`.
    - Case 3: `git -C W grep -n -e 'func (g [*]Game) cornered(' -e 'func (g [*]Game) effectiveState(' -e 'type EffectiveState struct' -- pkg/engine` returned lines 40, 51, and 58 in `pkg/engine/layers.go`.
    - Case 4: `git -C W diff autoloop/m2-wave-d.2 --numstat -- pkg cmd` printed `2 1 pkg/engine/doc.go`, `20 0 pkg/engine/layers.go`, `0 13 pkg/engine/state.go`, `0 7 pkg/engine/trigger.go`.
    - Case 5: `go -C W run ./cmd/cardtown lint fences` passed (exit 0).
    - Case 6: `git -C W diff autoloop/m2-wave-d.2 --stat` showed only the four allowed files and this task file.
    - Case 7: `go -C W tool task check` passed all stages (fmt-check, vet, tests, all linters); `go -C W tool task sim` passed (all 200 matches terminated with invariants holding).
  - Reviewer notes: Pure move, byte-identical sim behavior, no test files modified.
- 2026-11-07, lead-architect (wave D.2 run, session 3, D5): **Approved, `done`.** Merged into `autoloop/m2-wave-d.2` as 8ee37a4. Flash run `676985c2-eb94-483d-8ebb-4b288a83b6d5`: exit 0, `status: SUCCESS`, no refusals (no `denied_actions` value; the transcript has no step with `"status":"ERROR"` and no `permission check failed` match; `.err` holds only agy's "root agent idle" notice). Review of `git diff autoloop/m2-wave-d.2...m2/autoloop-wave-d.2/T-0189-effective-reads-in-layers` (480edf7):
  - Scope: only the four Allowed Files and this file changed. No test file changed.
  - Pure move, checked against the diff: the hunks are exactly Edits 1-4. The three declarations' comments and bodies are the same text, in the same package.
  - Fences and placement: no card name, no new rule, no new declaration. `layers.go` is where #126 places them.
  - Re-run by the reviewer in a scratch worktree of the task branch, all as expected: cases 1-6; `gofmt -l` clean; `cardtown lint fences` exit 0; `go tool task check` exit 0; `go tool task sim` exit 0, clean, 200 games, P0 78.5% / P1 19.0% / draws 2.5%, 47/50 and 34/50 main cards executable; `go tool task sim-compare -- autoloop/m2-wave-d.2` `[PASS]`.
  - The merge's `pkg/` and `cmd/` match the task tip's (`git diff` empty), so the gates hold on the run branch. Directive #126 is applied. #127 is next.
