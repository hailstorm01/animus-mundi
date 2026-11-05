---
id: T-0173
title: One "first picks in this order", firstPicks, and one listedOrder (directive #110, harness, flash, pure move)
kind: harness
tier: flash
wave: D.1 (board directive #110, structural review 8, finding 7)
status: review
branch: m2/autoloop-wave-d.1/T-0173-first-picks
depends_on: [T-0172]
claims: [helper:firstPicks, helper:listedOrder]
cards: []
requirements: []
---

## Goal
"Take the first `d.Min` options in this order, and a repeated pick always lands on the first one" (`SVY-R32`) is written once: `firstPicks(d Decision, order []int) []int` in `pkg/engine/choice.go`, used by every fallback that takes first picks (#110). `listedOrder(d Decision) []int` is the one identity order the callers start from. A pure move: same behavior, same test names, same sim output.

## Inputs
- Board directive #110 (structural review 8, approved by the user, 2026-11-05).
- Design note: `docs/architecture/m2-ability-model.md`, Deliverable 6g (wave D.1).
- Beyond #110's three sites, the reuse search (D2.3) found two things this task folds in, as T-0170 did with Convert:
  - `handPicks` ends in the same "first `d.Min` of this order" (`idx[:min(d.Min, len(idx))]`).
  - `firstPicks`' listed-order callers need the identity order, whose loop `soleAnswer`, `handPicks`, `fixedOrderPicks` and `randomPicks` each write themselves. A helper beside those four would be a fifth copy, so they use it. `trigger.go:166` stays: its order indexes a trigger list, not a `Decision`.
- Why it is a pure move:
  - **Bounds:** `fallbackPicks` runs only after `choose` has checked `d.satisfiable()`, and nothing else calls it, `fixedOrderPicks` or `handPicks`. So a `Repeats` decision with `Min` above 0 has at least one option, and any other has at least `Min`. `order[0]` and `order[i]` are in range, and `handPicks`' `min` never changed anything.
  - **`DecideOrder`:** its only decision (`trigger.go:169`) doesn't set `Repeats`, so joining the mode and search case gives the same `0, 1, 2…`.
  - **Mode and search:** the old `picks[i] = 0` or `i` is `order[0]` or `order[i]` of the identity order.
  - **`Min` 0:** `fixedOrderPicks` keeps its early `return nil`. Every other caller already returned an empty, non-nil slice, and `firstPicks` still does. `handPicks` now returns a new slice rather than a slice of `idx`, which nothing else holds.
  - **`soleAnswer`:** `listedOrder` builds the same slice its loop did.
  - **`randomPicks`:** the same `idx`, and the same draws from `r`.
- Not in this task: #106's own `DecideSearch` case for a search with criteria, which edits this code next.
- The lead dry-ran these exact edits in a throwaway worktree of the run branch at 8e7edc4: `gofmt -l` clean, `go tool task check` passes, `go tool task sim` runs clean, and `go tool task sim-compare -- autoloop/m2-wave-d.1` printed `[PASS]`. Every expected output in the cases below is from that dry run.

### Where it is today (at `autoloop/m2-wave-d.1`, engine code as at 8e7edc4)
| Where | Today |
| :--- | :--- |
| `pkg/engine/choice.go:88-92` | `soleAnswer` writes the identity order |
| `pkg/engine/choice.go:140-145` | `fallbackPicks`' `DecideOrder` case has its own loop |
| `pkg/engine/choice.go:148-157` | `fallbackPicks`' mode and search case writes `SVY-R32` |
| `pkg/engine/choice.go:171-174`, `:182` | `handPicks` writes the identity order and slices the first `d.Min` |
| `pkg/engine/choice.go:194-197`, `:205-213` | `fixedOrderPicks` writes the identity order and `SVY-R32` |
| `pkg/engine/choice.go:228-231` | `randomPicks` writes the identity order |

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)
All in `pkg/engine/choice.go`. Line numbers are the file before any edit, so work from the bottom up or match on the text.

### 1. `randomPicks`
Replace lines 228-231, from `idx := make([]int, len(d.Options))` through the loop's closing brace, with:
```go
	idx := listedOrder(d)
```

### 2. `fixedOrderPicks`, the end
Replace lines 205-213, from `picks := make([]int, d.Min)` through `return picks`, with:
```go
	return firstPicks(d, idx)
```

### 3. `fixedOrderPicks`, the identity order
Replace lines 194-197, from `idx := make([]int, len(d.Options))` through the loop's closing brace, with:
```go
	idx := listedOrder(d)
```

### 4. `fixedOrderPicks`, the comment
Replace lines 185-188 with:
```go
// fixedOrderPicks returns the indexes of the first d.Min options in the FR-MOD-001
// §6 fixed order (SVY-R16), in that order. It draws nothing from g.rng.
```

### 5. `handPicks`, the end
Replace line 182 with:
```go
	return firstPicks(d, idx)
```

### 6. `handPicks`, the identity order
Replace lines 171-174, from `idx := make([]int, len(d.Options))` through the loop's closing brace, with:
```go
	idx := listedOrder(d)
```

### 7. `fallbackPicks`, and the two new functions
Replace lines 140-160, from `case DecideOrder:` through the closing brace of `fallbackPicks`, with:
```go
	case DecideOrder, DecideMode, DecideSearch:
		return firstPicks(d, listedOrder(d))
	case DecideHand:
		return g.handPicks(d)
	}
	return nil
}

// listedOrder returns d's option indexes in the order d lists them.
func listedOrder(d Decision) []int {
	order := make([]int, len(d.Options))
	for i := range order {
		order[i] = i
	}
	return order
}

// firstPicks returns the first d.Min options in order, which holds d's option
// indexes. For a Repeats decision every one of the d.Min picks is the first
// option in that order, never a cycle (SVY-R32).
func firstPicks(d Decision, order []int) []int {
	picks := make([]int, d.Min)
	for i := range picks {
		if d.Repeats {
			picks[i] = order[0]
		} else {
			picks[i] = order[i]
		}
	}
	return picks
}
```

### 8. `soleAnswer`
Replace lines 88-92, from `picks := make([]int, len(d.Options))` through `return picks, true`, with:
```go
		return listedOrder(d), true
```

No other file changes.

## Acceptance Cases (harness tasks only; written by the lead)
No new tests: a pure move, so no `test(T-0173)` commit. W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **Existing tests unchanged** (#110): `git -C W diff autoloop/m2-wave-d.1 -- '*_test.go'` prints nothing, and `go -C W test ./...` passes.
2. **Sims byte-identical** (#110): `go -C W tool task sim-compare -- autoloop/m2-wave-d.1` prints `[PASS]`.
3. **One of each, and every caller:** `git -C W grep -n -e 'firstPicks(' -e 'listedOrder(' -- pkg/engine` prints exactly these nine lines of `pkg/engine/choice.go`: `:88`, `:137`, `:145` (the `listedOrder` declaration), `:156` (the `firstPicks` declaration), `:177`, `:185`, `:195`, `:203` and `:218`.
4. **The copies are gone:** `git -C W grep -n -e 'idx.i. = i' -e 'picks.i. = i' -e 'picks.i. = idx' -e 'idx.:min' -- pkg/engine/choice.go` prints nothing and exits 1.
5. **`fixedOrderPicks` keeps its guard:** `git -C W grep -n -e 'd.Min == 0' -- pkg/engine/choice.go` prints exactly two lines: `pkg/engine/choice.go:75`, in `satisfiable`, and `pkg/engine/choice.go:191`, in `fixedOrderPicks`.
6. **Only the listed edits:** `git -C W diff autoloop/m2-wave-d.1 --numstat -- pkg cmd` prints exactly this line: `33	46	pkg/engine/choice.go`.
7. **Fences:** `go -C W run ./cmd/cardtown lint fences` exits 0.
8. **Nothing else changed:** `git -C W diff autoloop/m2-wave-d.1 --stat` lists only `pkg/engine/choice.go` and this task file.
9. `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/engine/choice.go`

## Done When
- [x] Cases 1–9 hold, with their commands and results in the Log
- [x] `go tool task sim-compare -- autoloop/m2-wave-d.1` passes (starter and stress sims byte-identical to the run branch)
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-05, lead-architect (wave D.1 run, session 4): wrote this task from board directive #110, next in board order, against the run branch at 8e7edc4. No `main` commits to sync. Not `WAVE-COMPLETE`: #110–#119 and #121 are `open: before the next Deliverables task`. Harness, so it skips `tests` and goes straight to `ready`; tier `flash`, because the edits are exact and were dry-run. Every command here is on the headless allow-list, one per line, with no pipes, `$`, double quotes or unquoted parentheses. Searched for existing equivalents: nothing in `pkg/` or `cmd/` takes first picks outside `choice.go`, and the two finds there are in Inputs. Waiting directives checked: none has triggered (#6, #7: mode is `serial`; #8: no `deck.go`; #15: no turn-boundary hook; #71, #91 (2), #106, #118 (b), #120: no such card). Next: `kernel-engineer` on Flash.
- 2026-11-05, kernel-engineer (wave D.1 run, session 4): implemented T-0173 in `pkg/engine/choice.go`. Replaced inline identity loops with `listedOrder(d)` across `soleAnswer`, `fallbackPicks`, `handPicks`, `fixedOrderPicks`, and `randomPicks`. Replaced fallback first-picks and repetition logic across `fallbackPicks`, `handPicks`, and `fixedOrderPicks` with `firstPicks(d, order)`. Searched `pkg/` and `cmd/` for existing reuse: no other occurrences outside `choice.go`. Verified Cases 1–9:
  - Case 1: `git -C W diff autoloop/m2-wave-d.1 -- '*_test.go'` printed nothing; `go -C W test ./...` passed.
  - Case 2: `go -C W tool task sim-compare -- autoloop/m2-wave-d.1` printed `[PASS] starter and stress sims match autoloop/m2-wave-d.1 byte for byte, exit codes included`.
  - Case 3: `git -C W grep -n -e 'firstPicks(' -e 'listedOrder(' -- pkg/engine` returned exactly 9 matches (`choice.go:88`, `:137`, `:145`, `:156`, `:177`, `:185`, `:195`, `:203`, `:218`).
  - Case 4: `git -C W grep -n -e 'idx.i. = i' -e 'picks.i. = i' -e 'picks.i. = idx' -e 'idx.:min' -- pkg/engine/choice.go` returned nothing (exit 1).
  - Case 5: `git -C W grep -n -e 'd.Min == 0' -- pkg/engine/choice.go` returned lines :75 and :191.
  - Case 6: `git -C W diff autoloop/m2-wave-d.1 --numstat -- pkg cmd` returned `33 46 pkg/engine/choice.go`.
  - Case 7: `go -C W run ./cmd/cardtown lint fences` exited 0.
  - Case 8: `git -C W diff autoloop/m2-wave-d.1 --stat` showed only `pkg/engine/choice.go` and task file.
  - Case 9: `go -C W tool task check` and `go -C W tool task sim` passed.
  Status moved to review.
