---
id: T-0193
title: One read of "the source, or its last-known record", `resolveAttack` calls `itemSource` (directive #131, harness, flash, pure move)
kind: harness
tier: flash
wave: D.2
status: in-progress
branch: m2/autoloop-wave-d.2/T-0193-one-read-of-source-or-last-known
depends_on: [T-0192]
claims: [engine:resolveAttack-source-lookup, engine:itemSource-comment]
cards: []
requirements: []
---

## Goal
`FR-ZON-004`'s lookup, "the source, or its last-known record", is written in both `resolveAttack` and `itemSource`. `resolveAttack` now calls `itemSource` and keeps `it.Controller` as the fallback for a nil result (#131). A pure move: same test names, same sim output.

## Inputs
- Board directive #131 (structural review 9, approved by the user, 2026-11-06).
- Design note: `docs/architecture/m2-ability-model.md`, Deliverable 6h.
- Why it is a pure move: `itemSource(it)` is `g.Asset(it.Source)`, or else `g.LastKnown(it.Source)`, which is what `resolveAttack` wrote out. A nil result leaves `attController` at `it.Controller`, as both of the old branches failing did.
- Not in this task: `value` in `effects.go` also tries `g.Asset`, then `g.LastKnownEffective`. It reads the effective record's power, not the `Asset` copy, and takes an ID rather than a stack item. It isn't the lookup #131 names, so it stays.
- Reuse search (D2.3): the only non-test callers of `g.LastKnown` are `resolveAttack` and `itemSource` (searched `LastKnown(` in `pkg` and `cmd`). This task adds no new declaration.
- The lead dry-ran these exact edits in a throwaway worktree of the run branch at a8bad00: `gofmt -l pkg cmd` clean, `go tool task check` and `go tool task sim` pass, and `go tool task sim-compare -- autoloop/m2-wave-d.2` printed `[PASS]`. Every expected output below is from that dry run.

## Acceptance Cases (harness tasks only; written by the lead)
No new tests: a pure move, so no `test(T-0193)` commit. W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **Existing tests unchanged** (#131): `git -C W diff autoloop/m2-wave-d.2 -- '*_test.go'` prints nothing, and `go -C W test ./...` passes.
2. **Sims byte-identical** (#131): `go -C W tool task sim-compare -- autoloop/m2-wave-d.2` prints `[PASS]`.
3. **One lookup is left:** `git -C W grep -n -e 'g.LastKnown(' -- pkg/engine/combat.go pkg/engine/effects.go` prints exactly one line, at `pkg/engine/effects.go:370`.
4. **Only the listed edits:** `git -C W diff autoloop/m2-wave-d.2 --numstat -- pkg cmd` prints exactly two lines: `1	3	pkg/engine/combat.go` and `3	2	pkg/engine/effects.go`.
5. **Nothing else changed:** `git -C W diff autoloop/m2-wave-d.2 --stat` lists only those two files and this task file.
6. `go -C W tool task check` and `go -C W tool task sim` pass.

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)
Line numbers are at a8bad00, before any edit.

### 1. `pkg/engine/combat.go`
Replace lines 89-93, from `if a := g.Asset(it.Source); a != nil {` through the closing `}` of the `else if`, with:
```go
	if a := g.itemSource(it); a != nil {
		attController = a.Controller
	}
```
The comment above and `attController := it.Controller` stay.

### 2. `pkg/engine/effects.go`
Replace lines 362-364, the comment above `itemSource`, with:
```go
// itemSource returns the item's source asset, or its last-known record once it has
// left the battlefield (FR-ZON-004), so a damage event still names and attributes
// it (FR-REP-004 Stage 5) and an attack still knows its attacker's controller
// (FR-CBT-002 Step 3). It is nil only for a source that was never an asset.
```
The function itself doesn't change.

No other file changes.

## Allowed Files
- `pkg/engine/combat.go`
- `pkg/engine/effects.go`

## Done When
- [ ] Cases 1-6 hold, with their commands and results in the Log
- [ ] `go tool task sim-compare -- autoloop/m2-wave-d.2` passes (starter and stress sims byte-identical to the run branch)
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-07, lead-architect (wave D.2 run, session 7): wrote this task from board directive #131, next in board order, against the run branch at a8bad00. No `main` commits to sync. Not `WAVE-COMPLETE`: #131-#135 are `open: before the next Deliverables task`. Harness, so it skips `tests` and goes straight to `ready`; tier `flash`, because the edits are exact and were dry-run. Every command here is on the headless allow-list, one per line, with no pipes, `$`, double quotes or unquoted parentheses. Waiting directives checked: none has triggered (#6, #7: mode is `serial`; #15: no turn-boundary hook; #71, #91 (2), #106, #118 (b), #120, #128: no such card). Next: `kernel-engineer` on Flash.
