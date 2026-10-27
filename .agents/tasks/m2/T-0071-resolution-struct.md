---
id: T-0071
title: Effect steps take one per-resolution struct (directive #51 part 1, harness, flash, pure move)
kind: harness
tier: flash
wave: B (board directive #51, structural review 3 finding 3; before the "if you do" shape for Infested Apostle and Arkha, SVY-R14)
status: review
branch: m2/T-0071-resolution-struct
depends_on: [T-0070]
claims: [type:resolution, convention:step<Op>(r, st)]
cards: []
requirements: []
---

## Goal
Every `step<Op>` function in `pkg/engine/effects.go` takes one `*resolution` (the stack item and the chosen entities still legal) instead of the separate `it *StackItem` and `chosen [][]*Asset` parameters. This is part 1 of directive #51, a pure move: same tests, same sim output. Part 2, steps that return how many things they did, goes with the "you may X. If you do, Y" shape task (`SVY-R14`: Infested Apostle, Arkha). That task adds a field to `resolution` without touching every signature again.

## Inputs
- Board directive #51 (structural review 3, finding 3).
- Design note: `docs/architecture/m2-ability-model.md`, Key types. The lead records the convention there in part 2, once steps report counts.
- Why it is a pure move: every caller of these functions is in `effects.go`. `git grep` over `pkg` and `cmd` finds only `resolveAbility`, called from `stack.go:94`, and no test calls a step function directly. The struct carries the same two values the parameters did, built the same way.

### Where it is today (at `autoloop/m2-wave-b` 1d9fded)
| Where | Today |
| :--- | :--- |
| `effects.go:14-45`, `resolveAbility` | builds `chosen := make([][]*Asset, len(it.Chosen))`, then calls `g.step<Op>(it, st, chosen)` for each step |
| `effects.go:59-137`, the eight `step<Op>` functions | `(it *StackItem, st ability.Step, chosen [][]*Asset)`; `stepScratch`, `stepDraw`, `stepDiscard` and `stepRelocate` name the last one `_` |
| `effects.go:171`, `stepAssets` | `(it *StackItem, st ability.Step, chosen [][]*Asset) []*Asset` |

## Edits (lead decisions; copy exactly)

### 1. A new type, right above `resolveAbility`
```go
// resolution is one resolving ability as its steps see it: the stack item and,
// for each target slot, the chosen entities still legal after the second
// verification stage, in pick order (board directive #51).
type resolution struct {
	it     *StackItem
	chosen [][]*Asset
}
```

### 2. `resolveAbility`
Its doc comment is unchanged. The body becomes:
```go
	// Second verification stage (FR-TGT-003), entity by entity: one that is
	// gone or no longer legal is dropped, and the slot keeps the rest (partial
	// invalidation).
	r := &resolution{it: it, chosen: make([][]*Asset, len(it.Chosen))}
	for i, ids := range it.Chosen {
		for _, id := range ids {
			if a := g.Asset(id); a != nil && g.legalTarget(it.Controller, it.Source, it.Effect.Targets[i].Filter, a) {
				r.chosen[i] = append(r.chosen[i], a)
			}
		}
	}
	for _, st := range it.Effect.Steps {
		switch st.Op {
		case ability.OpDealDamage:
			g.stepDealDamage(r, st)
		case ability.OpHeal:
			g.stepHeal(r, st)
		case ability.OpAddArmor:
			g.stepAddArmor(r, st)
		case ability.OpScratch:
			g.stepScratch(r, st)
		case ability.OpDraw:
			g.stepDraw(r, st)
		case ability.OpDiscard:
			g.stepDiscard(r, st)
		case ability.OpModifyStat:
			g.stepModifyStat(r, st)
		case ability.OpReturnToHand, ability.OpPutOnDeck:
			g.stepRelocate(r, st)
		}
	}
```

### 3. The eight `step<Op>` functions and `stepAssets`
Change these nine signatures: `stepDealDamage`, `stepHeal`, `stepAddArmor`, `stepScratch`, `stepDraw`, `stepModifyStat`, `stepDiscard`, `stepRelocate` and `stepAssets`. In each one:
- The parameters `it *StackItem, st ability.Step, chosen [][]*Asset` (or `_ [][]*Asset`) become `r *resolution, st ability.Step`. `stepAssets` keeps its `[]*Asset` result.
- Inside the body, `it` becomes `r.it` and `chosen` becomes `r.chosen`.
- Each `g.stepAssets(it, st, chosen)` becomes `g.stepAssets(r, st)`.
- `stepRelocate`'s body becomes `{ g.relocateSelf(r.it, st) }`.

Doc comments stay unchanged. `stepPlayers`, `value`, `itemSource` and `relocateSelf` read only the item, so they keep their signatures. Their callers pass `r.it`, for example `g.value(r.it, st.Amount)` and `stepPlayers(r.it, st)`.

No other file changes.

## Acceptance Cases (harness tasks only; written by the lead)
No new tests: a pure move. Run each command on its own, pointed at your worktree. Record each command and its result in the Log. The reviewer re-runs them.
1. **Existing tests unchanged:** `git -C <worktree> diff autoloop/m2-wave-b -- '*_test.go'` prints nothing, and `go -C <worktree> test ./...` passes.
2. **Sims byte-identical:** `go -C <worktree> tool task sim-compare -- autoloop/m2-wave-b` prints `[PASS]`.
3. **No step takes `chosen` any more:** `git -C <worktree> grep -n "chosen \[\]\[\]\*Asset" -- pkg/engine` prints exactly one line, the `resolution` field in `effects.go`.
4. **Every step takes the struct:** `git -C <worktree> grep -n "func (g \*Game) step" -- pkg/engine/effects.go` prints nine lines, and each contains `(r *resolution, st ability.Step)`.
5. **Only these edits:** `git -C <worktree> diff autoloop/m2-wave-b --stat` lists only `pkg/engine/effects.go` and this task file.
6. `go -C <worktree> tool task check` and `go -C <worktree> tool task sim` pass.

## Allowed Files
- `pkg/engine/effects.go`

## Done When
- [x] Cases 1–5 hold, with their commands and results in the Log
- [x] `go tool task sim-compare -- autoloop/m2-wave-b` passes (starter and stress sims byte-identical to the run branch)
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks

## Log
- 2026-10-26 lead-architect (wave run B, session 5): Wrote the task. Wave B's deck cards still to build are Infested Apostle (4 in the Animus list), Arkha, Sun's Fury (1 in Path) and Eternal Watchman (the Path HQ). Directive #14 (a) puts deck cards first. The Apostle and Arkha share one new shape, "you may X. If you do, Y" (`SVY-R14`), which fires directive #51. #51 asks for a pure move first, then steps that return counts. As with #50 (T-0062, then T-0063), the size rule (D2.4) splits it. This pure move goes first. The "if you do" shape follows: the step counts, an untargeted pick on resolution, `KILL`, and the gate on an earlier step's count. Then the Apostle card. Arkha's discard-pile pick fires #59 (1) and comes after that. Eternal Watchman needs a combat-kill hook and a target in the discard pile, so it waits until the pile-pick convention is in.
  - Line references checked against `autoloop/m2-wave-b` 1d9fded. Every command under Acceptance Cases and Done When fits the headless allowlist in `~/.gemini/antigravity-cli/settings.json`.
  - Waiting triggers checked: #6 and #7 (mode is `serial`), #8 (no `deck.go` behavior change), #15 (no turn hook), #20 (no wave C shape task), #58 (no exhaust by a cost or an effect) and #59 (no pile pick, no repeated targets). None has fired. #51 has, and is now `open`.
  - Harness task, so it skips `tests` and goes straight to `ready`. Next: dispatch `kernel-engineer` on Gemini Flash through `agy` (D3).
- 2026-10-26 kernel-engineer (Gemini Flash): Implemented T-0071.
  - What changed:
    - Added `type resolution struct { it *StackItem; chosen [][]*Asset }` right above `resolveAbility` in `pkg/engine/effects.go`.
    - `resolveAbility` constructs `r := &resolution{it: it, chosen: make([][]*Asset, len(it.Chosen))}` and passes `(r, st)` to each `step<Op>` call.
    - Updated signatures and bodies of `stepDealDamage`, `stepHeal`, `stepAddArmor`, `stepScratch`, `stepDraw`, `stepModifyStat`, `stepDiscard`, `stepRelocate`, and `stepAssets` to take `(r *resolution, st ability.Step)`.
    - Internal accesses in step functions changed from `it` / `chosen` to `r.it` / `r.chosen` and `g.stepAssets(r, st)`.
    - `stepRelocate`'s body became `{ g.relocateSelf(r.it, st) }`.
  - Reuse:
    - Bundled existing parameters `it` and `chosen` into the single per-resolution struct `resolution`. No new shape added.
  - Verification results:
    - Case 1: `git diff autoloop/m2-wave-b -- '*_test.go'` returned no output (tests unchanged). `go test ./...` passed across all packages.
    - Case 2: `go tool task sim-compare -- autoloop/m2-wave-b` printed `[PASS] starter and stress sims match autoloop/m2-wave-b byte for byte, exit codes included`.
    - Case 3: `git grep -n "chosen \[\]\[\]\*Asset" -- pkg/engine` printed exactly one line: `pkg/engine/effects.go:17:	chosen [][]*Asset`.
    - Case 4: `git grep -n "func (g \*Game) step" -- pkg/engine/effects.go` printed exactly nine lines, all containing `(r *resolution, st ability.Step)`.
    - Case 5: `git diff autoloop/m2-wave-b --stat` showed only `pkg/engine/effects.go` and `.agents/tasks/m2/T-0071-resolution-struct.md`.
    - `go tool task check` passed (gofmt, `go vet`, all tests, `cardtown lint --all`).
    - `go tool task sim` passed clean (200 matches, all invariants held).
  - Status moved to `review`.
