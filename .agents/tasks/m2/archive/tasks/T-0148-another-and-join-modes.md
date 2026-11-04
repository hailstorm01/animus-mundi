---
id: T-0148
title: One read of "another", anotherOf, and modes joined in one place, joinModes (directive #97, harness, flash, pure move)
kind: harness
tier: flash
wave: C.1 (board directive #97, structural review 7, finding 6)
status: done
branch: m2/autoloop-wave-c.1/T-0148-another-and-join-modes
depends_on: [T-0147]
claims: [helper:anotherOf, helper:joinModes, helper:shiftSlots]
cards: []
requirements: []
---

## Goal
`chooseTargets` and `slotsLegal` each work out which earlier slot an "another" slot differs from, and both carry a `k >= 0` that can never be false. That read moves into `anotherOf(ts []ability.Target, i int) (k int, ok bool)` in `pkg/engine/targeting.go`, and both call it. `chooseModes` both asks and renumbers every slot field by hand, so the joining moves into `joinModes`, with the renumbering in one helper, `shiftSlots`, both in `pkg/engine/placement.go` (#97). A pure move: same behavior, same test names, same sim output.

## Inputs
- Board directive #97 (structural review 7, approved by the user, 2026-11-03).
- Design note: `docs/architecture/m2-ability-model.md`, Deliverable 6f (wave C.1).
- Rules these serve, unchanged: `FR-MOD-001` §5, `FR-TGT-001`, `SVY-R34`.
- Not in this task: `ability.Target.AnotherSlot` stays, and `anotherOf` is now its only caller. `anotherOf` keeps the `k < i` check, because dropping it would change what a malformed definition does. Only the `k >= 0` goes: `AnotherSlot` returns `ok` only when `Another > 0`, so `k` is at least 0 then. `withoutRefs` and the rest of `chooseTargets` don't change.
- Placement: `anotherOf` goes in `targeting.go`, above `withoutRefs`, beside the choosing it serves. `joinModes` and `shiftSlots` go in `placement.go`, directly below `chooseModes`.
- Why it is a pure move: `anotherOf`'s `ok` is the old condition without its always-true part, and `slotsLegal` still tests `exactlyOne(ts[k])` only when that holds. The joined effect has the same targets and steps in the same order with the same shifts, and stays nil-sliced when a mode has none (`slices.Clone` and an empty `append` both keep nil). The card's own slices are still never written to, which `TestFR_MOD_001_AnotherSlotDoesNotChangeTheCardDefinition` pins. The lead dry-ran these exact edits in a throwaway worktree of the run branch at eb11d8e: `gofmt -l` clean, `go tool task check` passes, `go tool task sim` runs clean, and `go tool task sim-compare -- autoloop/m2-wave-c.1` printed `[PASS]`.

### Where it is today (at `autoloop/m2-wave-c.1`, eb11d8e)
| Where | Today |
| :--- | :--- |
| `pkg/engine/targeting.go:35` | `chooseTargets`' read of the slot an "another" slot differs from |
| `pkg/engine/placement.go:80-81` | `slotsLegal`'s read of the same |
| `pkg/engine/placement.go:147-169` | `chooseModes` joins the picked modes and renumbers the slot fields by hand |

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)

### 1. `pkg/engine/targeting.go`, in `chooseTargets`
Replace line 35, `if k, ok := t.AnotherSlot(); ok && k >= 0 && k < i {`, with this line:
```go
		if k, ok := anotherOf(targets, i); ok {
```

### 2. `pkg/engine/targeting.go`, new function
Insert this text directly above the comment `// withoutRefs returns d without the options whose Ref is in refs, with its`, with one blank line on each side:
```go
// anotherOf returns the index of the earlier slot that slot i of ts must differ
// from (Target.Another, SVY-R34), and whether it has one.
func anotherOf(ts []ability.Target, i int) (k int, ok bool) {
	k, ok = ts[i].AnotherSlot()
	return k, ok && k < i
}
```

### 3. `pkg/engine/placement.go`, in `slotsLegal`
Replace lines 80-81 with these two lines:
```go
		k, hasPrev := anotherOf(ts, i)
		hasPrev = hasPrev && exactlyOne(ts[k])
```

### 4. `pkg/engine/placement.go`, the end of `chooseModes`, and two new functions
Replace everything after the line `slices.Sort(picked)` to the end of the file (lines 147-170 before these edits: the `// New slices throughout` comment, the `joined` loop, `return joined` and `chooseModes`' closing brace) with this text:
```go
	modes := make([]ability.Effect, len(picked))
	for i, j := range picked {
		modes[i] = e.Modes[legal[j]]
	}
	return joinModes(modes)
}

// joinModes joins modes, in the order given, into the one effect the stack
// item resolves: each mode's targets and steps follow the last mode's, with
// their slot references renumbered for the joined Targets (shiftSlots).
func joinModes(modes []ability.Effect) ability.Effect {
	var joined ability.Effect
	for _, m := range modes {
		shifted := shiftSlots(m, len(joined.Targets))
		joined.Targets = append(joined.Targets, shifted.Targets...)
		joined.Steps = append(joined.Steps, shifted.Steps...)
	}
	return joined
}

// shiftSlots returns a copy of mode m with every reference to a target slot
// moved up by shift: a target's Another, and a step's Target and Dealer. New
// slices: a mode's own Targets and Steps belong to the card's definition and
// must never be written to.
func shiftSlots(m ability.Effect, shift int) ability.Effect {
	out := ability.Effect{Targets: slices.Clone(m.Targets), Steps: slices.Clone(m.Steps)}
	for i := range out.Targets {
		if out.Targets[i].Another > 0 {
			out.Targets[i].Another += shift
		}
	}
	for i := range out.Steps {
		st := &out.Steps[i]
		if st.Subject == ability.SubjectTarget {
			st.Target += shift
		}
		if st.Dealer > 0 {
			st.Dealer += shift
		}
	}
	return out
}
```

No other file changes.

## Acceptance Cases (harness tasks only; written by the lead)
No new tests: a pure move. W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **Existing tests unchanged** (#97): `git -C W diff autoloop/m2-wave-c.1 -- '*_test.go'` prints nothing, and `go -C W test ./...` passes.
2. **Sims byte-identical** (#97): `go -C W tool task sim-compare -- autoloop/m2-wave-c.1` prints `[PASS]`.
3. **One read of "another", two callers** (#97): `git -C W grep -n 'anotherOf(' -- pkg/engine` prints exactly three lines: `pkg/engine/placement.go:80` (the `slotsLegal` call), `pkg/engine/targeting.go:35` (the `chooseTargets` call) and `pkg/engine/targeting.go:53` (the declaration).
4. **`AnotherSlot` has one reader:** `git -C W grep -n 'AnotherSlot()' -- pkg/engine` prints exactly two lines: `pkg/engine/ability/ability.go:171` (the declaration) and `pkg/engine/targeting.go:54`.
5. **The dead check is gone** (#97): `git -C W grep -n 'k >= 0' -- pkg/engine/placement.go pkg/engine/targeting.go` prints nothing (git grep exits 1 when nothing matches; that is the expected result).
6. **One join** (#97): `git -C W grep -n 'joinModes(' -- pkg/engine` prints exactly two lines: `pkg/engine/placement.go:151` (the `chooseModes` call) and `pkg/engine/placement.go:157` (the declaration).
7. **One renumbering** (#97): `git -C W grep -n 'shiftSlots(' -- pkg/engine` prints exactly two lines: `pkg/engine/placement.go:160` (the `joinModes` call) and `pkg/engine/placement.go:171` (the declaration). `git -C W grep -n 'Another += ' -- pkg/engine` prints exactly one line, at `pkg/engine/placement.go:175`.
8. **Only the listed edits:** `git -C W diff autoloop/m2-wave-c.1 --numstat -- pkg` prints exactly these two lines: `39	21	pkg/engine/placement.go` and `8	1	pkg/engine/targeting.go`.
9. **Fences:** `go -C W run ./cmd/cardtown lint fences` exits 0.
10. **Nothing else changed:** `git -C W diff autoloop/m2-wave-c.1 --stat` lists only `pkg/engine/placement.go`, `pkg/engine/targeting.go` and this task file.
11. `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/engine/placement.go`
- `pkg/engine/targeting.go`

## Done When
- [x] Cases 1–11 hold, with their commands and results in the Log
- [x] `go tool task sim-compare -- autoloop/m2-wave-c.1` passes (starter and stress sims byte-identical to the run branch)
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-03, lead-architect (wave C.1 run, session 5): wrote this task from directive #97 against the run branch at eb11d8e; `main` had no commits to sync. Harness, so it skips `tests` and goes straight to `ready`; tier `flash`, because the edits are exact and were dry-run. Every command here is on the headless allow-list, one per line, with no pipes, `$` or unquoted parentheses. Searched for existing equivalents: `pkg/engine` has no `anotherOf`, `joinModes` or `shiftSlots`; `AnotherSlot` is read only at the two sites this task replaces, and `chooseModes` is the only place that renumbers slot fields. The directive named `anotherOf` and `joinModes` and left the renumbering helper's name open: it is `shiftSlots`, taking a whole mode, so one function holds all three slot fields. Waiting directives checked: none has triggered (#6, #7: mode is `serial`; #8: no `deck.go`; #15: no turn-boundary hook; #71, #91 (2): no such card). Next: `kernel-engineer` on Flash.
- 2026-11-03, kernel-engineer (session 1): implemented pure move of "another" read to anotherOf in pkg/engine/targeting.go, and mode joining and renumbering to joinModes and shiftSlots in pkg/engine/placement.go (#97).
  - Diff shows:
    - pkg/engine/targeting.go: replaced inline AnotherSlot read in chooseTargets with anotherOf(targets, i); declared anotherOf helper above withoutRefs.
    - pkg/engine/placement.go: replaced inline AnotherSlot read in slotsLegal with anotherOf(ts, i) and simplified hasPrev; refactored mode joining at end of chooseModes to return joinModes(modes); declared joinModes and shiftSlots helpers.
  - Reuse: anotherOf unifies reading an earlier slot reference between chooseTargets and slotsLegal. joinModes and shiftSlots unify mode concatenation and slot index renumbering.
  - Acceptance cases verified:
    - Case 1: git -C W diff autoloop/m2-wave-c.1 -- "*_test.go" printed nothing; go -C W test ./... passed (exit 0).
    - Case 2: go -C W tool task sim-compare -- autoloop/m2-wave-c.1 printed "[PASS] starter and stress sims match autoloop/m2-wave-c.1 byte for byte, exit codes included".
    - Case 3: git -C W grep -n 'anotherOf(' -- pkg/engine printed exactly 3 lines: pkg/engine/placement.go:80, pkg/engine/targeting.go:35, and pkg/engine/targeting.go:53.
    - Case 4: git -C W grep -n 'AnotherSlot()' -- pkg/engine printed exactly 2 lines: pkg/engine/ability/ability.go:171 and pkg/engine/targeting.go:54.
    - Case 5: git -C W grep -n 'k >= 0' -- pkg/engine/placement.go pkg/engine/targeting.go printed nothing (exit 1).
    - Case 6: git -C W grep -n 'joinModes(' -- pkg/engine printed exactly 2 lines: pkg/engine/placement.go:151 and pkg/engine/placement.go:157.
    - Case 7: git -C W grep -n 'shiftSlots(' -- pkg/engine printed exactly 2 lines: pkg/engine/placement.go:160 and pkg/engine/placement.go:171; git -C W grep -n 'Another += ' -- pkg/engine printed exactly 1 line at pkg/engine/placement.go:175.
    - Case 8: git -C W diff autoloop/m2-wave-c.1 --numstat -- pkg printed "39\t21\tpkg/engine/placement.go" and "8\t1\tpkg/engine/targeting.go".
    - Case 9: go -C W run ./cmd/cardtown lint fences exited 0 ("[PASS]").
    - Case 10: git -C W diff autoloop/m2-wave-c.1 --stat listed only pkg/engine/placement.go, pkg/engine/targeting.go, and this task file.
    - Case 11: go -C W tool task check and go -C W tool task sim passed all checks.
  - Handoff commit feat(engine): anotherOf, joinModes and shiftSlots helpers (T-0148). Task status set to review.
- 2026-11-03, lead-architect (wave C.1 run, session 5, D5): **Approved, `done`.** Merged into `autoloop/m2-wave-c.1` as cca2a9b. Flash run `bce46ea5-77a5-47ab-b7f7-4cfe7c83eb0c`: exit 0, `status: SUCCESS`, 369 s, no refusals (no `denied_actions` key; the transcript has no step with `"status":"ERROR"` and no `permission check failed` match; the `.err` file is empty). Review of `git diff autoloop/m2-wave-c.1...m2/autoloop-wave-c.1/T-0148-another-and-join-modes` (0333de6):
  - Scope: only `placement.go`, `targeting.go` and this file changed. No test file changed.
  - Pure move, checked against the diff: both code files match the task's Edits exactly. `anotherOf`'s `ok` is the old condition without `k >= 0`, `slotsLegal` still tests `exactlyOne(ts[k])` only when it holds, and `shiftSlots` shifts the same three fields under the same conditions on copies of the card's slices. Only this file's status, Log and Done When boxes changed besides.
  - Fences and placement: no card name, no new rule. `anotherOf` sits in `targeting.go`; `joinModes` and `shiftSlots` sit below `chooseModes` in `placement.go`.
  - Re-run by the reviewer in a scratch worktree of the task branch, all as expected: cases 1-10; `cardtown lint fences` `[PASS]`; `gofmt -l` clean; `go test ./...` passes; `go tool task check` passes; `go tool task sim` clean, 200 games, P0 82.5% / P1 17.5% / draws 0.0%, 42/50 and 32/50 main cards executable; `go tool task sim-compare -- autoloop/m2-wave-c.1` `[PASS]`.
  - The merge's `pkg/` matches the task tip's, so the gates hold on the run branch. Directive #97 is applied. Nothing is queued now; the next session takes #98 from the board.
