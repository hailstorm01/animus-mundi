---
id: T-0188
title: A value doesn't need a stack item, `value` takes the controller, source and instance (directive #125, harness, flash, pure move)
kind: harness
tier: flash
wave: D.2
status: in-progress
branch: m2/autoloop-wave-d.2/T-0188-value-without-stack-item
depends_on: [T-0187]
claims: [signature:Game.value]
cards: []
requirements: []
---

## Goal
`value` reads a `Value` without a stack item: it takes the controller, the source asset ID and the source instance, and its six callers in `effects.go` pass them (#125). *Boneyard Amalgam*'s count is read while checking a cost, when no item exists, so this removes the fake item or second reader that task would otherwise need. A pure move: same behavior, same test names, same sim output. No count source is added here.

## Inputs
- Board directive #125 (structural review 9, approved by the user, 2026-11-06).
- Design note: `docs/architecture/m2-ability-model.md`, Key types (`Value`), Deliverable 6h.
- `value` reads only `it.Source` today, so `controller` and `inst` are unused until a count `Value` arrives. That is the signature #125 approves: don't drop them, and don't rename them `_`.
- Why it is a pure move: every caller passes `r.it`'s own `Source`, the one field the old body read. `value` has no callers outside these six and none in a test file (`git grep`, at 9cd61d9).
- Not in this task: `replaceDeath`'s item at `pkg/engine/replace.go:112` stays, because it runs steps, and `itemSource` still takes an item.
- Reuse search (D2.3): no other reader of `ability.Value` exists in `pkg/engine`.
- The lead dry-ran these exact edits in a throwaway worktree of the run branch at 9cd61d9: `gofmt -l` clean, `go tool task check` and `go tool task sim` pass, and `go tool task sim-compare -- autoloop/m2-wave-d.2` printed `[PASS]`. Every expected output below is from that dry run.

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)
Everything is in `pkg/engine/effects.go`. Line numbers are at 9cd61d9.

### 1. The six callers
In each of these lines, change `g.value(r.it, st.Amount` to `g.value(r.it.Controller, r.it.Source, r.it.Inst, st.Amount`, and change nothing else on the line:

| Line | In | Today |
| :--- | :--- | :--- |
| 178 | `stepDealDamage` | `n := g.value(r.it, st.Amount, dealer)` |
| 188 | `stepHeal` | `n := g.value(r.it, st.Amount, nil)` |
| 198 | `stepAddArmor` | `n := g.value(r.it, st.Amount, nil)` |
| 223 | `deckDrop` | `take(pid, g.value(r.it, st.Amount, nil))` |
| 234 | `stepModifyStat` | `a.TempPower += g.value(r.it, st.Amount, nil)` |
| 247 | `stepDiscard` | `n, total := g.value(r.it, st.Amount, nil), 0` |

### 2. `value` itself
Replace lines 329-334, from `// value reads v for it` through the `func` line, with:
```go
// value reads v as a step resolves (SVY-R10), or wherever no stack item exists,
// such as a cost check. It takes the three things a read may need: the controller
// and source instance are for a count, which no Value reads yet. With a nil
// dealer, source power is the source's effective power if it is on the
// battlefield, else the effective power recorded as it left (FR-ZON-004), else 0.
// With a chosen dealer it is that asset's current power, never a last-known value
// (SVY-R27).
func (g *Game) value(controller PlayerID, source int, inst *Instance, v ability.Value, dealer *Asset) int {
```
Then, in that function's body only, change `g.Asset(it.Source)` to `g.Asset(source)` and `g.LastKnownEffective(it.Source)` to `g.LastKnownEffective(source)`. `itemSource`, right after it, has its own `g.Asset(it.Source)`, which stays.

No other file changes.

## Acceptance Cases (harness tasks only; written by the lead)
No new tests: a pure move, so no `test(T-0188)` commit. W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **Existing tests unchanged** (#125): `git -C W diff autoloop/m2-wave-d.2 -- '*_test.go'` prints nothing, and `go -C W test ./...` passes.
2. **Sims byte-identical** (#125): `go -C W tool task sim-compare -- autoloop/m2-wave-d.2` prints `[PASS]`.
3. **Every read passes the three:** `git -C W grep -n -e 'g.value(' -e 'func (g [*]Game) value(' -- pkg/engine` prints exactly seven lines: `pkg/engine/effects.go` lines 178, 188, 198, 223, 234 and 247, each holding `g.value(r.it.Controller, r.it.Source, r.it.Inst, st.Amount`, and line 336, the declaration in Edit 2.
4. **Only the listed edits:** `git -C W diff autoloop/m2-wave-d.2 --numstat -- pkg cmd` prints exactly one line: `15	13	pkg/engine/effects.go`.
5. **Fences:** `go -C W run ./cmd/cardtown lint fences` exits 0.
6. **Nothing else changed:** `git -C W diff autoloop/m2-wave-d.2 --stat` lists only `pkg/engine/effects.go` and this task file.
7. `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/engine/effects.go`

## Done When
- [ ] Cases 1–7 hold, with their commands and results in the Log
- [ ] `go tool task sim-compare -- autoloop/m2-wave-d.2` passes (starter and stress sims byte-identical to the run branch)
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-06, lead-architect (wave D.2 run, session 2): wrote this task from board directive #125, next in board order, against the run branch at 9cd61d9. No `main` commits to sync. Not `WAVE-COMPLETE`: #125-#127 and #129-#135 are `open: before the next Deliverables task`. Harness, so it skips `tests` and goes straight to `ready`; tier `flash`, because the edits are exact and were dry-run. Every command here is on the headless allow-list, one per line, with no pipes, `$`, double quotes or unquoted parentheses. Waiting directives checked: none has triggered (#6, #7: mode is `serial`; #15: no turn-boundary hook; #71, #91 (2), #106, #118 (b), #120, #128: no such card). Next: `kernel-engineer` on Flash.
