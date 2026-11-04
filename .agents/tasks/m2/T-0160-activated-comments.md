---
id: T-0160
title: The docs say activated abilities are built, stack.go, doc.go, costs.go and effects.go comments (harness, flash, comments only)
kind: harness
tier: flash
wave: D
status: in-progress
branch: m2/autoloop-wave-d/T-0160-activated-comments
depends_on: [T-0158]
claims: [comments:pkg/engine/stack.go StackAbility Tracked Keyword, comments:pkg/engine/doc.go activated, comments:pkg/engine/costs.go, comments:pkg/engine/effects.go resolveAbility]
cards: []
requirements: []
---

## Goal
No comment still says that a `StackAbility` item is only a trigger, or that activated abilities are not built. No code changes: same declarations, same tests, same sim output.

## Inputs
- T-0158's review (the `costs.go` comments) and its D2.4 split. FR-ACT-001 and FR-ACT-002 in FRD 01.
- Lead decisions beyond the first list of what this task covers, all comments:
  - **`effects.go`:** `resolveAbility`'s comment also says a triggered ability's or a Maneuver's, so it joins this task. That makes four files, the D2.4 limit.
  - **`payCosts`:** it said "the Maneuver is pushed" and "the declared card", but it also runs for an activated ability (`declare`, `pkg/engine/placement.go:40`), so it now says "the item".
- Left alone on purpose, each with its reason:
  - `placeable`'s and `chooseModes`' comments (`pkg/engine/placement.go:105-107`, `:136-137`) still hold for every card and test: `check` reads only `costsPayable` for an activation (`pkg/engine/actions.go:195`), and no activated ability has modes. Whether an activation needs `placeable` is #104's first question, so a comment here would guess at the ruling.
  - `placement.costs`' comment (`pkg/engine/placement.go:21`) and `CostKind`'s (`pkg/engine/ability/ability.go:44`) say "additional cost", loosely. Neither says an item is only a trigger or that activation isn't built, so they are outside the Goal, and the two files would pass the size limit.
- Why it is comments only: the lead dry-ran these exact edits in a throwaway worktree of the run branch at dfce2be: `go tool task check` passes, `go tool task sim` runs clean, and `go tool task sim-compare -- autoloop/m2-wave-d` printed `[PASS]`.

## Edits (lead decisions; copy exactly)

### 1. `pkg/engine/stack.go`, three trailing comments
Only the trailing comments change, and the code and the alignment stay:
- Line 15: replace `// a triggered ability: runs Effect's steps for Controller` with `// a triggered or activated ability: runs Effect's steps for Controller`.
- Line 32: after `its owner's discard pile (FR-ZON-004)`, add `; false for an activated ability`.
- Line 33: replace `zero for a printed ability` with `zero for a printed trigger or an activated ability`.

### 2. `pkg/engine/effects.go`, `resolveAbility`
Replace lines 25-26 with these two lines. Line 27 stays:
```go
// resolveAbility runs the stack item's effect, a triggered or activated
// ability's or an enacted Maneuver's, step by step in order. It resolves even if
```

### 3. `pkg/engine/costs.go`, line 1
Replace the line with:
```go
// Costs: reading and paying a card's cost, and checking and paying a placement's costs, a Maneuver's additional costs or an activated ability's (FR-CST-001, FR-TRN-003a, FR-ACT-001).
```

### 4. `pkg/engine/costs.go`, `costsPayable`
Replace its comment (lines 67-69) with:
```go
// costsPayable reports whether every cost of pl can be paid now: a resources
// cost needs its Requisition and Tactics (CostResources), exhausting the source
// needs it Readied (CostExhaustSelf), and a kill or exhaust cost needs at least
// one option (FR-TRN-003a, FR-ACT-001, SVY-R29). A discard cost is always
// payable, because discarding none pays it, and so is killing the source
// (CostKillSelf).
```

### 5. `pkg/engine/costs.go`, `payCosts`
Replace its comment (lines 94-104) with:
```go
// payCosts pays each cost of pl in printed order: its controller picks the asset
// (costDecision, pickFrom) and it is killed through kill (CostKill) or exhausted
// through exhaust (CostExhaust), the source itself is killed or exhausted the
// same way (CostKillSelf, CostExhaustSelf), a resources cost takes its
// Requisition and Tactics (CostResources), and a discard cost moves the cards
// picked earlier (CostDiscard). A replaced kill (not the discard cost) still
// counts as paid (SVY-R15), and the triggers a kill or an exhaust fires wait in
// the pending buffer to go above the declared item (FR-TRN-003a). It runs after
// the item is pushed, so it is already on the stack when the cost's triggers are
// queued (FR-TRN-003), and after its targets were chosen, so they were judged
// against the state before the cost moved anything (SVY-R2). A discard cost's
// cards were picked before the modes and targets (chooseCostDiscards) and are
// moved here, through discardCards.
```

### 6. `pkg/engine/doc.go`, a new bullet under "What runs now"
After line 43 (`//     to kill, exhaust or discard (Def.ExtraCosts), paid on declaration (FR-TRN-003a).`), and before `//   - Ward on targets, ...`, insert:
```go
//   - Activated abilities (ActActivate, FR-ACT-001, FR-ACT-002): a printed
//     "[Cost]: [Effect]." ability of a battlefield asset (Def.Activated), activated
//     on an empty stack. On declaration its targets are chosen, it goes on the stack
//     as a StackAbility item and its costs are paid, including CostResources,
//     CostExhaustSelf and CostKillSelf (declare, SVY-R2).
```

### 7. `pkg/engine/doc.go`, "Not built yet"
Replace `//   - Activated and continuous abilities, continuous Cornered,` (line 55 today) with:
```go
//   - Continuous abilities, continuous Cornered,
```

### 8. `pkg/engine/doc.go`, the Files entry
In the line `//     additional costs. targeting.go: target choice and filters. effects.go:` (line 69 today), replace `additional costs.` with `a placement's costs.`

No other file changes.

## Acceptance Cases (harness tasks only; written by the lead)
No new tests: comments only. W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **Existing tests unchanged:** `git -C W diff autoloop/m2-wave-d -- '*_test.go'` prints nothing, and `go -C W test ./...` passes.
2. **Sims byte-identical:** `go -C W tool task sim-compare -- autoloop/m2-wave-d` prints `[PASS]`.
3. **Only the listed edits:** `git -C W diff autoloop/m2-wave-d --numstat -- pkg` prints exactly these four lines: `20	15	pkg/engine/costs.go`, `7	2	pkg/engine/doc.go`, `2	2	pkg/engine/effects.go` and `3	3	pkg/engine/stack.go`.
4. **An item may be activated:** `git -C W grep -n -F 'triggered or activated' -- pkg/engine/stack.go pkg/engine/effects.go` prints exactly two lines, at `effects.go:25` and `stack.go:15`.
5. **The field comments:** `git -C W grep -n -e 'false for an activated ability' -e 'zero for a printed trigger or an activated ability' -- pkg/engine/stack.go` prints exactly two lines, at `stack.go:32` and `:33`.
6. **`doc.go` names what T-0158 built:** `git -C W grep -n -e ActActivate -e Def.Activated -e CostKillSelf -- pkg/engine/doc.go` prints exactly three lines, at `doc.go:44`, `:45` and `:48`.
7. **`doc.go` no longer calls it unbuilt:** `git -C W grep -n -F 'Activated and continuous' -- pkg/engine/doc.go` prints nothing (git exits 1).
8. **`costs.go` no longer speaks only of additional costs or Maneuvers:** `git -C W grep -n -e 'additional cost of pl' -e 'the Maneuver is pushed' -e 'the declared card' -- pkg/engine/costs.go` prints nothing (git exits 1).
9. **Fences:** `go -C W run ./cmd/cardtown lint fences` exits 0.
10. **Nothing else changed:** `git -C W diff autoloop/m2-wave-d --stat` lists only the four files in case 3 and this task file.
11. `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/engine/stack.go`
- `pkg/engine/effects.go`
- `pkg/engine/costs.go`
- `pkg/engine/doc.go`

## Done When
- [ ] Cases 1–11 hold, with their commands and results in the Log
- [ ] `go tool task sim-compare -- autoloop/m2-wave-d` passes (starter and stress sims byte-identical to the run branch)
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-04 lead-architect: split off T-0158 under the D2 size rule (D2.4). Written `queued`.
- 2026-11-04 lead-architect (wave D run, session 3): wrote Inputs, Edits, Acceptance Cases and Allowed Files against the run branch at dfce2be; `main` has no commits to sync. Harness, so it skips `tests` and goes straight to `ready`; tier `flash`, because the edits are exact and were dry-run. Every command here is on the headless allow-list, one per line, with no pipes, `$` or unquoted parentheses. Checked each claim against the code: `ActActivate`, `Def.Activated` and the three cost kinds exist; `check` refuses an activation on a non-empty stack (`pkg/engine/actions.go:145`); `declare` chooses targets, pushes the item and pays the costs in that order (`pkg/engine/placement.go:40-45`); `pendingTrigger` items set `Tracked` only for `HookKilled` and the activation's item leaves it and `Keyword` zero (`pkg/engine/trigger.go:181-182`, `pkg/engine/actions.go:103`). Waiting directives checked: none has triggered (#6, #7: mode is `serial`; #8: no `deck.go`; #15: no turn-boundary hook; #71, #91 (2): no such card; #104: not due, this task has no `cards:`). Next: `kernel-engineer` on Flash.
