---
id: T-0151
title: One check for a Scratch Company card, IsScratchCompany (directive #98, part 3 of 3, harness, flash, pure move)
kind: harness
tier: flash
wave: C.1 (board directive #98, structural review 7, finding 7)
status: done
branch: m2/autoloop-wave-c.1/T-0151-is-scratch-company
depends_on: [T-0150]
claims: [helper:IsScratchCompany]
cards: []
requirements: []
---

## Goal
`HasSubtype("Scratch Company")` is repeated in three files. `(*Card).IsScratchCompany()` in `pkg/engine/card.go`, below `HasSubtype`, becomes the one check, and those sites call it (#98). A pure move: same behavior, same test names, same sim output.

## Inputs
- Board directive #98 (structural review 7, approved by the user, 2026-11-03).
- Design note: `docs/architecture/m2-ability-model.md`, Deliverable 6f (wave C.1).
- Split from T-0149 under the D2 size rule: #98 touches nine non-test files. T-0149 (`done`, merge 2d98782) added `Resource.String` and the sites that write a name, and T-0150 (`done`, merge 0a64407) the sites that read one. This task is the last part. When it merges, #98 moves to the directive archive.
- Not in this task: the `scratchFace` literal in `card.go` stays (#98), as do messages and comments that say "Scratch Company", and the test fixtures. `HasSubtype` stays, because the new method calls it and `pkg/engine/foundation_test.go` and `rules_test.go` use it.
- This task touches `deck.go` as a pure move, so it doesn't fire #8's trigger.
- The fences linter exempts "Scratch Company" as a rules object (`pkg/toolchain/linter/fences/fences.go:25`), so the method's literal passes.
- Why it is a pure move: the method only calls `HasSubtype` with the same string.
- The lead dry-ran these exact edits in a throwaway worktree of the run branch at 900933f: `gofmt -l` clean, `go tool task check` passes, `go tool task sim` runs clean, and `go tool task sim-compare -- autoloop/m2-wave-c.1` printed `[PASS]`.

### Where it is today (at `autoloop/m2-wave-c.1`, 900933f)
| Where | Today |
| :--- | :--- |
| `pkg/engine/card.go:105-113` | `HasSubtype`, with `// Supported reports ...` right below it at line 115 |
| `pkg/engine/targeting.go:263` and `:265` | `ScratchOnly` and `ScratchExcluded` type out the subtype |
| `pkg/engine/deck.go:71` | the deck check types out the subtype |
| `pkg/sim/sim.go:38` | the stress-deck pool types out the subtype |

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)

### 1. `pkg/engine/card.go`, below `HasSubtype`
Above the line `// Supported reports whether the kernel can fully execute this card.`, insert these lines, followed by one blank line:
```go
// IsScratchCompany reports whether the card is a Scratch Company (FR-ZON-005).
func (c *Card) IsScratchCompany() bool {
	return c.HasSubtype("Scratch Company")
}
```

### 2. `pkg/engine/targeting.go`, in the `f.Scratch` switch
Replace line 263 with this line:
```go
		return c.IsScratchCompany()
```
Replace line 265 with this line:
```go
		return !c.IsScratchCompany()
```

### 3. `pkg/engine/deck.go`
Replace line 71 with this line:
```go
		case c.IsScratchCompany():
```

### 4. `pkg/sim/sim.go`
Replace line 38 with this line:
```go
		case c.Supported() && !c.IsScratchCompany() &&
```

No other file changes.

## Acceptance Cases (harness tasks only; written by the lead)
No new tests: a pure move. W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **Existing tests unchanged** (#98): `git -C W diff autoloop/m2-wave-c.1 -- '*_test.go'` prints nothing, and `go -C W test ./...` passes.
2. **Sims byte-identical** (#98): `go -C W tool task sim-compare -- autoloop/m2-wave-c.1` prints `[PASS]`.
3. **One check** (#98): `git -C W grep -n 'HasSubtype(.Scratch Company.)' -- pkg/engine/card.go pkg/engine/targeting.go pkg/engine/deck.go pkg/sim/sim.go` prints exactly one line, at `pkg/engine/card.go:117`.
4. **The sites call it** (#98): `git -C W grep -n 'IsScratchCompany()' -- pkg/engine/targeting.go pkg/engine/deck.go pkg/sim/sim.go` prints exactly four lines: `pkg/engine/deck.go:71`, `pkg/engine/targeting.go:263`, `pkg/engine/targeting.go:265` and `pkg/sim/sim.go:38`.
5. **The literal stays** (#98): `git -C W grep -n 'Subtypes: ' -- pkg/engine/card.go` still prints the `scratchFace` line naming `Infantry` and `Scratch Company`.
6. **Only the listed edits:** `git -C W diff autoloop/m2-wave-c.1 --numstat -- pkg cmd` prints exactly these four lines: `5	0	pkg/engine/card.go`, `1	1	pkg/engine/deck.go`, `2	2	pkg/engine/targeting.go` and `1	1	pkg/sim/sim.go`.
7. **Fences:** `go -C W run ./cmd/cardtown lint fences` exits 0.
8. **Nothing else changed:** `git -C W diff autoloop/m2-wave-c.1 --stat` lists only those four files and this task file.
9. `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/engine/card.go`
- `pkg/engine/targeting.go`
- `pkg/engine/deck.go`
- `pkg/sim/sim.go`

## Done When
- [x] Cases 1-9 hold, with their commands and results in the Log
- [x] `go tool task sim-compare -- autoloop/m2-wave-c.1` passes (starter and stress sims byte-identical to the run branch)
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-03, lead-architect (wave C.1 run, session 6): queued with T-0149 (D2.4).
- 2026-11-03, lead-architect (wave C.1 run, session 8): wrote Inputs, Edits, Acceptance Cases and Allowed Files against the run branch at 900933f; `main` had no commits to sync. Harness, so it skips `tests` and goes straight to `ready`; tier `flash`, because the edits are exact and were dry-run. Every command here is on the headless allow-list, one per line, with no pipes, `$`, double quotes or unquoted parentheses. Reuse search: no helper for this subtype exists in `pkg/engine` or `pkg/sim`; `HasSubtype` is what it calls. Waiting directives checked: none has triggered (#6, #7: mode is `serial`; #8: this task is a pure move; #15: no turn-boundary hook; #71, #91 (2): no such card). Next: `kernel-engineer` on Flash.
- 2026-11-03, kernel-engineer (Gemini Flash): implemented pure move per lead edits.
  - Diff shows:
    - `pkg/engine/card.go`: added `func (c *Card) IsScratchCompany() bool { return c.HasSubtype("Scratch Company") }` below `HasSubtype`.
    - `pkg/engine/targeting.go`: replaced `c.HasSubtype("Scratch Company")` and `!c.HasSubtype(...)` with `c.IsScratchCompany()` and `!c.IsScratchCompany()`.
    - `pkg/engine/deck.go`: replaced `c.HasSubtype("Scratch Company")` in the deck check with `c.IsScratchCompany()`.
    - `pkg/sim/sim.go`: replaced `!c.HasSubtype("Scratch Company")` with `!c.IsScratchCompany()`.
  - Reuse: `IsScratchCompany()` calls `c.HasSubtype("Scratch Company")`.
  - Acceptance cases 1-9 verified:
    1. `git -C W diff autoloop/m2-wave-c.1 -- '*_test.go'` printed nothing; `go -C W test ./...` passed (all packages ok).
    2. `go -C W tool task sim-compare -- autoloop/m2-wave-c.1` printed `[PASS] starter and stress sims match autoloop/m2-wave-c.1 byte for byte, exit codes included`.
    3. `git -C W grep -n 'HasSubtype(.Scratch Company.)' -- pkg/engine/card.go pkg/engine/targeting.go pkg/engine/deck.go pkg/sim/sim.go` printed `pkg/engine/card.go:117:	return c.HasSubtype("Scratch Company")` (1 line).
    4. `git -C W grep -n 'IsScratchCompany()' -- pkg/engine/targeting.go pkg/engine/deck.go pkg/sim/sim.go` printed exactly 4 lines (`pkg/engine/deck.go:71`, `pkg/engine/targeting.go:263`, `pkg/engine/targeting.go:265`, `pkg/sim/sim.go:38`).
    5. `git -C W grep -n 'Subtypes: ' -- pkg/engine/card.go` printed `pkg/engine/card.go:137:	Subtypes: []string{"Infantry", "Scratch Company"},`.
    6. `git -C W diff autoloop/m2-wave-c.1 --numstat -- pkg cmd` printed `5 0 pkg/engine/card.go`, `1 1 pkg/engine/deck.go`, `2 2 pkg/engine/targeting.go`, `1 1 pkg/sim/sim.go`.
    7. `go -C W run ./cmd/cardtown lint fences` exited 0.
    8. `git -C W diff autoloop/m2-wave-c.1 --stat` listed only the four allowed files and this task file.
    9. `go -C W tool task check` (fmt, vet, test, lint --all) and `go -C W tool task sim` (200 games, invariants held) both passed.
  - Implementation committed as `feat(engine): IsScratchCompany replaces repeated HasSubtype checks (T-0151)` (9aedbdf). Handed off to review.
- 2026-11-03, lead-architect (wave C.1 run, session 8, D5): **Approved, `done`.** Merged into `autoloop/m2-wave-c.1` as d933973. Flash run `0bcd4331-eb1e-4132-aa51-0deecfe5248a`: exit 0, `status: SUCCESS`, 287 s, no refusals (no `denied_actions` key; the transcript has no step with `"status":"ERROR"` and no `permission check failed` match; the `.err` file holds only agy's line about waiting on a background task). Review of `git diff autoloop/m2-wave-c.1...m2/autoloop-wave-c.1/T-0151-is-scratch-company` (9aedbdf):
  - Scope: only `card.go`, `targeting.go`, `deck.go`, `sim.go` and this file changed. No test file changed.
  - Pure move, checked against the diff: the four code files match the task's Edits exactly. The method only calls `HasSubtype` with the same string, and the `scratchFace` literal stays.
  - Fences and placement: no card name, no new rule. The one new declaration is on `Card`, in `pkg/engine`, because the engine and `pkg/sim` both read it.
  - Re-run by the reviewer in a scratch worktree of the task branch, all as expected: cases 1-9; `gofmt -l` clean; `go tool task check` passes; `go tool task sim` clean, 200 games, P0 82.5% / P1 17.5% / draws 0.0%, 42/50 and 32/50 main cards executable; `go tool task sim-compare -- autoloop/m2-wave-c.1` `[PASS]`.
  - The merge's `pkg/` and `cmd/` match the task tip's, so the gates hold on the run branch. This was #98's last part, so #98 moves to the directive archive.
