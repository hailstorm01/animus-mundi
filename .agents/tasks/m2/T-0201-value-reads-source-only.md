---
id: T-0201
title: "`value` takes only what it reads, and the docs say what wave E built (directives #138 and #139, harness, flash, pure move)"
kind: harness
tier: flash
wave: E.1
status: review
branch: m2/autoloop-wave-e.1/T-0201-value-reads-source-only
depends_on: [T-0200]
claims: [signature:Game.value, comments:pkg/engine/doc.go]
cards: []
requirements: []
---

## Goal
`value` takes only the source asset ID, the `Value` and the dealer, and its six callers in `effects.go` stop passing the controller and the instance (#138). `doc.go`'s "What runs now" lists a card's own hand-active cost reduction, and "Not built yet" stops implying that it is missing (#139). A pure move plus comments: same behavior, same test names, same sim output.

## Inputs
- Board directives #138 and #139 (structural review 10, findings 3 and 4, approved by the user, 2026-11-09). #139 may share #138's task, so this task covers both.
- Design note: `docs/architecture/m2-ability-model.md`, Deliverable 6i (wave E.1). The lead rewrote its `Value` line (Key types) in this task's commit, as #138 asks: no count source is built, and the first one reads `countPileCards`.
- Why it is a pure move: `value`'s body reads only `source`, `v` and `dealer` (`pkg/engine/effects.go:346-360` at 799f1a8). `controller` and `inst` have been unused since T-0188 added them, because T-0200 built *Boneyard Amalgam*'s count as `Def.CostReduction` and `countPileCards` (`pkg/engine/costs.go:17-41`). `value` has no callers outside these six and none in a test file (`git grep`, at 799f1a8).
- Reuse search (D2.3): no other reader of `ability.Value` exists in `pkg/engine`, and no count `Value` exists in `pkg/engine/ability`.
- The `doc.go` edits are comments only. *Boneyard Amalgam*'s text is "This unit costs R1 less to deploy for each unit card in your discard pile." (`data/cards`), and `Def.CostReduction`'s comment says what it is (`pkg/engine/ability/ability.go:38-49`). The Files entry for `costs.go` already says "a card's cost", so it stays.
- The lead dry-ran these exact edits in a throwaway worktree of the run branch at 799f1a8: `gofmt -l pkg` clean, `go tool task check` and `go tool task sim` exit 0, and `go tool task sim-compare -- autoloop/m2-wave-e.1` printed `[PASS]`. Every expected output below is from that dry run.

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)
Line numbers are at 799f1a8.

### 1. `pkg/engine/effects.go`, the six callers
In each of these lines, change `g.value(r.it.Controller, r.it.Source, r.it.Inst, st.Amount` to `g.value(r.it.Source, st.Amount`, and change nothing else on the line:

| Line | In | Today |
| :--- | :--- | :--- |
| 178 | `stepDealDamage` | `n := g.value(r.it.Controller, r.it.Source, r.it.Inst, st.Amount, dealer)` |
| 189 | `stepHeal` | `n := g.value(r.it.Controller, r.it.Source, r.it.Inst, st.Amount, nil)` |
| 201 | `stepAddArmor` | `n := g.value(r.it.Controller, r.it.Source, r.it.Inst, st.Amount, nil)` |
| 227 | `deckDrop` | `take(pid, g.value(r.it.Controller, r.it.Source, r.it.Inst, st.Amount, nil))` |
| 238 | `stepModifyStat` | `n := g.value(r.it.Controller, r.it.Source, r.it.Inst, st.Amount, nil)` |
| 253 | `stepDiscard` | `n, total := g.value(r.it.Controller, r.it.Source, r.it.Inst, st.Amount, nil), 0` |

### 2. `pkg/engine/effects.go`, `value` itself
Replace lines 339-346, from `// value reads v as a step resolves (SVY-R10), or wherever no stack item exists,` through the `func` line, with these six lines. The body doesn't change:
```go
// value reads v as a step resolves (SVY-R10), for the ability whose source asset
// has the ID source. With a nil dealer, source power is the source's effective
// power if it is on the battlefield, else the effective power recorded as it left
// (FR-ZON-004), else 0. With a chosen dealer it is that asset's current power,
// never a last-known value (SVY-R27).
func (g *Game) value(source int, v ability.Value, dealer *Asset) int {
```

### 3. `pkg/engine/doc.go`, What runs now
Add these three lines right after the line `//     is Cornered (FR-LAY-001, FR-KWD-002e, SVY-R18).` (line 52), before `//   - Ward on targets, ...`:
```go
//   - A card's own hand-active cost reduction ("costs R1 less to deploy for each
//     unit card in your discard pile", Def.CostReduction): EffectiveCost lowers its
//     Requisition cost while it is in its owner's hand (FR-CST-001).
```

### 4. `pkg/engine/doc.go`, "Not built yet"
Replace these two lines (64-65):
```go
//   - Continuous abilities other than a card's own Cornered keywords (auras, and
//     "Cornered: This unit has +2 power."),
```
with these two. The next line, `//     Regroup N, ...`, stays:
```go
//   - Continuous abilities other than a card's own Cornered keywords and its own
//     hand-active cost reduction (auras, and "Cornered: This unit has +2 power."),
```

No other file changes.

## Acceptance Cases (harness tasks only; written by the lead)
No new tests: a pure move plus comments, so no `test(T-0201)` commit. W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **Existing tests unchanged** (#138): `git -C W diff autoloop/m2-wave-e.1 -- '*_test.go'` prints nothing, and `go -C W test ./...` passes.
2. **Sims byte-identical** (#138, #139): `go -C W tool task sim-compare -- autoloop/m2-wave-e.1` prints `[PASS]`.
3. **Every read passes only the source:** `git -C W grep -n -e 'g.value(' -e 'func (g [*]Game) value(' -- pkg/engine` prints exactly seven lines: `pkg/engine/effects.go` lines 178, 189, 201, 227, 238 and 253, each holding `g.value(r.it.Source, st.Amount`, and line 344, `func (g *Game) value(source int, v ability.Value, dealer *Asset) int {`.
4. **The docs name the reduction** (#139): `git -C W grep -n -e 'hand-active cost reduction' -- pkg/engine/doc.go` prints exactly two lines, 53 and 68.
5. **Only the listed edits:** `git -C W diff autoloop/m2-wave-e.1 --numstat -- pkg cmd` prints exactly two lines: `5	2	pkg/engine/doc.go` and `12	14	pkg/engine/effects.go`.
6. **Fences:** `go -C W run ./cmd/cardtown lint fences` exits 0.
7. **Nothing else changed:** `git -C W diff autoloop/m2-wave-e.1 --stat` lists only `pkg/engine/doc.go`, `pkg/engine/effects.go` and this task file.
8. `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/engine/effects.go`
- `pkg/engine/doc.go`

## Done When
- [x] Cases 1–8 hold, with their commands and results in the Log
- [x] `go tool task sim-compare -- autoloop/m2-wave-e.1` passes (starter and stress sims byte-identical to the run branch)
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-09, lead-architect (wave E.1 run, session 1): wrote this task from board directives #138 and #139, the only `open: before` rows, against the run branch at 799f1a8. Run file created; no `main` commits to sync. Not `WAVE-COMPLETE`: #138 and #139 are `open: before the next Deliverables task`. Rewrote the design note's `Value` line in this commit (#138). Harness, so it skips `tests` and goes straight to `ready`; tier `flash`, because the edits are exact and were dry-run. Two non-test files, under the size rule. Every command here is on the headless allow-list, one per line, with no pipes, `$`, double quotes or unquoted parentheses. Waiting directives checked: none has triggered (#6, #7: mode is `serial`; #15: no turn-boundary hook; #71, #91 (2), #106, #120, #128, #136, #137: no such card in a task). Next: `kernel-engineer` on Flash.
- 2026-11-09, kernel-engineer (session 1): implemented directives #138 and #139.
  - Changes:
    - `pkg/engine/effects.go`: simplified `g.value(source int, v ability.Value, dealer *Asset)` signature and documentation comment to drop unused `controller` and `inst` arguments (#138). Updated all 6 caller sites in `stepDealDamage`, `stepHeal`, `stepAddArmor`, `deckDrop`, `stepModifyStat`, and `stepDiscard` to pass only `r.it.Source, st.Amount`.
    - `pkg/engine/doc.go`: documented a card's own hand-active cost reduction (`Def.CostReduction`) under "What runs now" (lines 53-55) and updated "Not built yet" continuous abilities clause (lines 68-69) (#139).
  - Reuse:
    - Checked `pkg/engine` for callers or alternate implementations of `ability.Value`; verified only the 6 caller sites in `pkg/engine/effects.go` resolve `ability.Value` at runtime, matching Deliverable 6i and directives #138/#139.
  - Acceptance Cases & verification commands:
    - Case 1: `git -C W diff autoloop/m2-wave-e.1 -- '*_test.go'` (clean, no test diff); `go -C W test ./...` passed (exit 0).
    - Case 2: `go -C W tool task sim-compare -- autoloop/m2-wave-e.1` printed `[PASS] starter and stress sims match autoloop/m2-wave-e.1 byte for byte, exit codes included`.
    - Case 3: `git -C W grep -n -e 'g.value(' -e 'func (g [*]Game) value(' -- pkg/engine` printed exactly the 7 expected lines in `pkg/engine/effects.go` (178, 189, 201, 227, 238, 253, and 344).
    - Case 4: `git -C W grep -n -e 'hand-active cost reduction' -- pkg/engine/doc.go` printed lines 53 and 68.
    - Case 5: `git -C W diff autoloop/m2-wave-e.1 --numstat -- pkg cmd` printed `5	2	pkg/engine/doc.go` and `12	14	pkg/engine/effects.go`.
    - Case 6: `go -C W run ./cmd/cardtown lint fences` passed with `[PASS] Package code is free of card names, invalid markers and disallowed imports.`.
    - Case 7: `git -C W diff autoloop/m2-wave-e.1 --stat` listed only `pkg/engine/doc.go`, `pkg/engine/effects.go`, and this task file.
    - Case 8: `go -C W tool task check` passed (fmt, vet, tests, full cardtown linter suite all clean); `go -C W tool task sim` passed (200 games, all invariants holding).
