---
id: T-0085
title: Placing an effect chooses its modes and targets from a placement, not a pending trigger (directive #62, harness, sonnet, pure move)
kind: harness
tier: sonnet
wave: B.1 (board directive #62, structural review 4 finding 1; before wave C's Maneuvers, which place effects with no trigger)
status: review
branch: m2/T-0085-placement
depends_on: [T-0084]
claims: [type:placement, func:legalModes, func:chooseOnPlacement, file:pkg/engine/placement.go]
cards: []
requirements: []
---

## Goal
Choosing an effect's modes and targets on placement no longer needs a fired trigger. `chooseModes`, `chooseTargets` and `targetDecision` take a `placement` (controller, source asset, source card instance, label) and the `ability.Effect` or `ability.Target` they choose for. Mode legality is its own function, `legalModes`. "Modes, then targets" is one function, `chooseOnPlacement`, which `flushTriggers` calls and a wave C Maneuver task will call too. All of it moves to a new `pkg/engine/placement.go`. This is directive #62, a pure move: same tests, same sim output. #61's rulings later edit `legalModes` alone.

## Inputs
- Board directive #62 (structural review 4, finding 1).
- Design note: `docs/architecture/m2-ability-model.md`, Deliverable 6a (wave B.1) and Open Question 11 (the legality reading `legalModes` keeps unchanged).
- Rules the moved code implements, unchanged: `FR-MOD-001` §1, §5, §6, `FR-TGT-001`, `FR-TGT-003`, `SVY-R3`, `SVY-R11`, `SVY-R16`.
- Why it is a pure move: `pendingTrigger`, `chooseModes`, `chooseTargets` and `targetDecision` are used only in `trigger.go`, `targeting.go` and `state.go` (the `pending` field). `git grep` over `pkg` and `cmd` finds no test that names any of them. The functions read the same four values they read today, now through `placement`, and are called in the same order.

### Where it is today (at `autoloop/m2-wave-b.1` c9c27ee)
| Where | Today |
| :--- | :--- |
| `trigger.go:13-21`, `pendingTrigger` | fields `source`, `inst`, `controller`, `label`, `pos`, `trig` |
| `trigger.go:95-98` (`fireHook`) and `trigger.go:126-130` (`fireAttacks`) | build `pendingTrigger` literals with all six fields |
| `trigger.go:171-176`, in `flushTriggers` | `if len(p.trig.Effect.Modes) > 0 { p.trig.Effect = g.chooseModes(p) }`, then `g.push(&StackItem{… Effect: p.trig.Effect, … Chosen: g.chooseTargets(p)})` |
| `trigger.go:191-240`, `chooseModes(p pendingTrigger) ability.Effect` | legality loop (lines 199-212), then the `DecideMode` decision and the join |
| `targeting.go:11-36`, `chooseTargets(p pendingTrigger) [][]int` | reads `p.trig.Effect.Targets` |
| `targeting.go:38-59`, `targetDecision(p pendingTrigger, i int, t ability.Target) Decision` | reads `p.label`, `p.controller`, `p.inst`, `p.source` |

## Edits (lead decisions; copy exactly)

### 1. New file `pkg/engine/placement.go`
Its first line is the file header, then the package clause and imports (`fmt`, `slices`, `animus-mundi/pkg/engine/ability`), then, in this order:

```go
// Placement: placing an effect on the stack, with its modes and then its targets chosen on placement (FR-MOD-001 §5, FR-TGT-001, SVY-R3).
```

```go
// placement is an effect being placed on the stack, as choosing its modes and
// targets sees it: who controls it, its source and the label its prompts use.
// A fired trigger carries one (pendingTrigger); a played Maneuver will too.
type placement struct {
	controller PlayerID
	source     int       // source asset ID
	inst       *Instance // source card instance, tracked by identity (FR-ZON-004)
	label      string
}
```

```go
// chooseOnPlacement chooses what e needs on placement, in rules order: its
// modes first, if it has any, then the targets of the effect that leaves
// (FR-MOD-001 §5, SVY-R3). It returns the effect the stack item resolves and
// the chosen targets, one slice of Refs per target slot.
func (g *Game) chooseOnPlacement(pl placement, e ability.Effect) (ability.Effect, [][]int) {
	if len(e.Modes) > 0 {
		e = g.chooseModes(pl, e)
	}
	return e, g.chooseTargets(pl, e)
}
```

```go
// legalModes returns the indices of e's legal modes, in printed order. A mode
// is legal if each of its exactly-one target slots has a legal entity now; an
// "up to N" slot never makes it illegal (design note Open Question 11).
func (g *Game) legalModes(pl placement, e ability.Effect) []int {
	var legal []int
	for k, m := range e.Modes {
		ok := true
		for i, t := range m.Targets {
			if t.UpTo == 0 && len(g.targetDecision(pl, i, t).Options) == 0 {
				ok = false
				break
			}
		}
		if ok {
			legal = append(legal, k)
		}
	}
	return legal
}
```

Then `chooseModes`, moved from `trigger.go` with these changes only:
- The signature becomes `func (g *Game) chooseModes(pl placement, e ability.Effect) ability.Effect`.
- The legality loop (today's `var legal []int` through the end of its `for`) becomes `legal := g.legalModes(pl, e)`.
- In the rest of the body, `p.trig.Effect` becomes `e`, `p.label` becomes `pl.label`, `p.controller` becomes `pl.controller` and `p.source` becomes `pl.source`.
- Its doc comment becomes:
  ```go
  // chooseModes asks pl's controller which modes of e to use, when it is placed
  // and before its targets (FR-MOD-001 §5, SVY-R3), and returns the effect the
  // item resolves: the chosen modes joined in printed order. Only legal modes
  // (legalModes) are offered, so a timeout takes the first listed legal ones
  // (§6) and a single legal mode is answered automatically (§1). With no legal
  // mode nothing is asked and the effect is empty. With fewer legal modes than
  // ChooseN all of them are chosen.
  ```
Everything else in its body, including the `// New slices throughout` comment and the `slices.Sort(picked)` line, stays as it is.

### 2. `pkg/engine/trigger.go`
- `pendingTrigger` embeds `placement` and keeps only its own two fields:
  ```go
  // pendingTrigger is a fired trigger waiting in the buffer for the flush.
  type pendingTrigger struct {
  	placement
  	pos  sourcePos // where the source stood when it fired
  	trig ability.Trigger
  }
  ```
- The two literals become `pendingTrigger{placement: placement{source: a.ID, inst: a.Inst, controller: a.Controller, label: <the same fmt.Sprintf as today>}, trig: <as today>, pos: posOf(a)}`. Keep each `label` expression exactly as it is.
- In `flushTriggers`, the `if len(p.trig.Effect.Modes) > 0 {…}` block and the `g.push` call become:
  ```go
  			eff, chosen := g.chooseOnPlacement(p.placement, p.trig.Effect)
  			g.push(&StackItem{Kind: StackAbility, Controller: p.controller, Source: p.source, Inst: p.inst,
  				Effect: eff, Tracked: p.trig.Hook == ability.HookKilled, Keyword: p.trig.Keyword,
  				Chosen: chosen})
  ```
- `chooseModes` and its doc comment are removed (they are in `placement.go` now).
- `accepts` keeps taking a `pendingTrigger`: "you may" is a trigger's question (SVY-R9), asked before placement.
- The file header becomes `// Triggers: the pending-trigger buffer, hook firing and the flush in APNAP order (FR-TRG-002 to FR-TRG-004). Placing each one is placement.go's.`
- Remove an import only if the compiler says it is unused (`cmp` and `slices` stay in use today).

### 3. `pkg/engine/targeting.go`
- `chooseTargets` becomes `func (g *Game) chooseTargets(pl placement, e ability.Effect) [][]int`. In its body, `p.trig.Effect.Targets` becomes `e.Targets` and `g.targetDecision(p, i, t)` becomes `g.targetDecision(pl, i, t)`. In its doc comment, `asks p's controller to choose each of its Effect.Targets` becomes `asks pl's controller to choose each of e's Targets`, and `when the trigger is placed on the stack` becomes `when the effect is placed on the stack`. The rest of the comment stays.
- `targetDecision` becomes `func (g *Game) targetDecision(pl placement, i int, t ability.Target) Decision`. In its body, `p.` becomes `pl.` (four fields). In its doc comment, `of p's trigger` becomes `of the effect pl places`, and `the legality check of a modal trigger's modes reads its Options` becomes `legalModes reads its Options`.

No other file changes. `pkg/engine/state.go`'s `pending []pendingTrigger` field stays as it is.

## Acceptance Cases (harness tasks only; written by the lead)
No new tests: a pure move. Run each command on its own, pointed at your worktree. Record each command and its result in the Log. The reviewer re-runs them.
1. **Existing tests unchanged:** `git -C <worktree> diff autoloop/m2-wave-b.1 -- '*_test.go'` prints nothing, and `go -C <worktree> test ./...` passes.
2. **Sims byte-identical:** `go -C <worktree> tool task sim-compare -- autoloop/m2-wave-b.1` prints `[PASS]`.
3. **No placing function takes a pending trigger** (directive #62 case (c)): `git -C <worktree> grep -n "pendingTrigger" -- pkg/engine/placement.go pkg/engine/targeting.go` prints exactly one line, the doc comment of `placement` in `placement.go`, and no function signature.
4. **One legality rule:** `git -C <worktree> grep -n "UpTo == 0" -- pkg/engine` prints exactly one line, in `legalModes` in `placement.go`.
5. **One placing order:** `git -C <worktree> grep -n "chooseModes(\|chooseTargets(" -- pkg/engine` lists, apart from the two `func` lines and doc-comment mentions, calls only inside `chooseOnPlacement`; `flushTriggers` calls `chooseOnPlacement` and neither of the others.
6. **Only these edits:** `git -C <worktree> diff autoloop/m2-wave-b.1 --stat` lists only `pkg/engine/placement.go`, `pkg/engine/trigger.go`, `pkg/engine/targeting.go` and this task file.
7. `go -C <worktree> tool task check` and `go -C <worktree> tool task sim` pass.

## Allowed Files
- `pkg/engine/placement.go` (new)
- `pkg/engine/trigger.go`
- `pkg/engine/targeting.go`

## Done When
- [ ] Cases 1–6 hold, with their commands and results in the Log
- [ ] `go tool task sim-compare -- autoloop/m2-wave-b.1` passes (starter and stress sims byte-identical to the run branch)
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks

## Log
- 2026-10-28, lead-architect (wave B.1 run, session 1): wrote the task from board directive #62, the first task of wave B.1. Harness, so it skips `tests` and goes straight to `ready`. Tier `sonnet` as the directive gives. Directives #63–#67 stay open for the next sessions, in board order (#63 with #66, #64, #65 with #67). Directive #61 stays `waiting`: no task's `cards:` has a modal card. Next: `kernel-engineer`.
- 2026-10-28, kernel-engineer (wave B.1 run, session 1): implemented the edits as written. Added `pkg/engine/placement.go` (`placement`, `chooseOnPlacement`, `legalModes`, `chooseModes` moved with the listed changes). `trigger.go`: `pendingTrigger` embeds `placement`, both literals rebuilt with unchanged label expressions, `flushTriggers` calls `chooseOnPlacement`, `chooseModes` removed, header updated. `targeting.go`: `chooseTargets(pl, e)`, `targetDecision(pl, i, t)`, doc comments updated. Reuse: no new shape; the legality loop and mode join are the old code moved. Note: `git grep` ignores untracked files, so `placement.go` was `git add`ed before cases 3-5. Acceptance commands (W = worktree):
  1. `git -C W diff autoloop/m2-wave-b.1 -- '*_test.go'` printed nothing; `go -C W test ./...` all ok.
  2. `go -C W tool task sim-compare -- autoloop/m2-wave-b.1` printed `[PASS] starter and stress sims match ... byte for byte`.
  3. `git -C W grep -n "pendingTrigger" -- pkg/engine/placement.go pkg/engine/targeting.go` printed only `placement.go:14` (the `placement` doc comment).
  4. `git -C W grep -n "UpTo == 0" -- pkg/engine` printed only `placement.go:41`, in `legalModes`.
  5. `git -C W grep -n "chooseModes(\|chooseTargets(" -- pkg/engine` printed: calls at `placement.go:28` and `:30` (inside `chooseOnPlacement`), `func` lines `placement.go:60` and `targeting.go:20`; `flushTriggers` (`trigger.go:169`) calls only `chooseOnPlacement`.
  6. `git -C W diff autoloop/m2-wave-b.1 --stat` (committed state) lists only placement.go, trigger.go, targeting.go and this task file.
  7. `go -C W tool task check` ended `All executed checks PASSED`; `go -C W tool task sim` ran clean.
