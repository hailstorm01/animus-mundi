---
id: T-0174
title: One "this asset is yours" check, ownAsset, and "Aggressive" spelled once (directives #111 and #112, harness, flash, pure move)
kind: harness
tier: flash
wave: D.1 (board directives #111 and #112, structural review 8, findings 8 and 9)
status: ready
branch: m2/autoloop-wave-d.1/T-0174-own-asset-and-aggressive-name
depends_on: [T-0173]
claims: [helper:ownAsset]
cards: []
requirements: []
---

## Goal
The check "this asset is yours on the battlefield" and its message are written once: `ownAsset(a Action) (*Asset, error)` in `pkg/engine/actions.go`, used by `readyAsset` and by `ActActivate` (#111). The name "Aggressive" is written once, in `ability.Keyword.String`, and `keywordDefs` and `fireAttacks`' label read it (#112). A pure move: same behavior, same test names, same sim output.

## Inputs
- Board directives #111 and #112 (structural review 8, approved by the user, 2026-11-05). #111 says they may share a task.
- Design note: `docs/architecture/m2-ability-model.md`, Deliverable 6g (wave D.1).
- Reuse search (D2.3): `is not yours` appears only at the two sites below. Nothing else in `pkg/` or `cmd/` writes the check. `readyAsset`'s callers (`check`'s Generate and March, `checkAttack`) don't change.
- The search for `Aggressive` found no other place that writes the name for code to use. These stay: the `Keywords.Aggressive` field and the `ability.Aggressive` constructor are identifiers; `"Cornered: Aggressive 3."` in `cards/path.go` is a card's printed text, which the card checks compare with the CSV; the mentions in comments are prose.
- Why it is a pure move:
  - **`ownAsset`:** the same two conditions in the same order, the same message. `ActActivate` returned that error from `check`, and still does. `readyAsset` still checks ownership before `Exhausted`.
  - **`keywordDefs`:** `ability.KeywordAggressive.String()` returns the literal it replaces, and `card.go` already imports `ability`, whose package initializes first.
  - **`fireAttacks`' label:** the same string from the same format, so the option labels that `aggressive_item_test.go` pins don't change.
- Not in this task: #113 onward, and #117's Cornered keywords, which build on #112.
- The lead dry-ran these exact edits in a throwaway worktree of the run branch at a26fcd5: `gofmt -l` clean, `go tool task check` passes, `go tool task sim` runs clean, and `go tool task sim-compare -- autoloop/m2-wave-d.1` printed `[PASS]`. Every expected output in the cases below is from that dry run.

### Where it is today (at `autoloop/m2-wave-d.1`, engine code as at a26fcd5)
| Where | Today |
| :--- | :--- |
| `pkg/engine/actions.go:181-184` | `check`'s `ActActivate` writes the check and message |
| `pkg/engine/actions.go:245-249` | `readyAsset` writes them again |
| `pkg/engine/ability/ability.go:120` | `Keyword.String` writes "Aggressive" (it stays) |
| `pkg/engine/card.go:55` | `keywordDefs` writes it |
| `pkg/engine/trigger.go:137` | `fireAttacks`' label writes it |

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)
Line numbers are each file before any edit, so work from the bottom up or match on the text.

### 1. `pkg/engine/actions.go`, `readyAsset`
Replace lines 245-249, from `func (g *Game) readyAsset(a Action) (*Asset, error) {` through the closing brace of its first `if`, with:
```go
// ownAsset returns the battlefield asset a names, if a.Player controls it.
func (g *Game) ownAsset(a Action) (*Asset, error) {
	as := g.Asset(a.Asset)
	if as == nil || as.Controller != a.Player {
		return nil, illegal("asset %d is not yours on the battlefield", a.Asset)
	}
	return as, nil
}

func (g *Game) readyAsset(a Action) (*Asset, error) {
	as, err := g.ownAsset(a)
	if err != nil {
		return nil, err
	}
```

### 2. `pkg/engine/actions.go`, `ActActivate`
Replace lines 181-184, from `as := g.Asset(a.Asset)` through the closing brace of its `if`, with:
```go
		as, err := g.ownAsset(a)
		if err != nil {
			return err
		}
```

### 3. `pkg/engine/card.go`, `keywordDefs`
Replace line 55 with:
```go
	{name: ability.KeywordAggressive.String(), num: func(k *Keywords) *int { return &k.Aggressive }},
```

### 4. `pkg/engine/card.go`, its comment
Replace line 51, `// and Keywords.Lines both walk it, so each keyword's name is written only here.`, with:
```go
// and Keywords.Lines both walk it, so each keyword's name is written only here,
// or, for a keyword that is a triggered ability, in ability.Keyword.
```

### 5. `pkg/engine/trigger.go`, `fireAttacks`
Replace line 137 with:
```go
		g.bufferTrigger(a, fmt.Sprintf("%s: %s %d", a.Printed().Name, ability.KeywordAggressive.String(), n), ability.Aggressive(n))
```

No other file changes.

## Acceptance Cases (harness tasks only; written by the lead)
No new tests: a pure move, so no `test(T-0174)` commit. W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **Existing tests unchanged** (#111, #112): `git -C W diff autoloop/m2-wave-d.1 -- '*_test.go'` prints nothing, and `go -C W test ./...` passes.
2. **Sims byte-identical** (#111, #112): `go -C W tool task sim-compare -- autoloop/m2-wave-d.1` prints `[PASS]`.
3. **One check** (#111): `git -C W grep -n -e 'is not yours' -- pkg/engine` prints exactly one line, at `pkg/engine/actions.go:249`.
4. **`ownAsset` and its callers** (#111): `git -C W grep -n -e 'ownAsset(' -- pkg` prints exactly three lines of `pkg/engine/actions.go`: `:181`, `:246` (the declaration) and `:255`.
5. **`readyAsset`'s callers unchanged** (#111): `git -C W grep -n -e 'readyAsset(' -- pkg/engine` prints exactly four lines: `pkg/engine/actions.go:146`, `pkg/engine/actions.go:194`, `pkg/engine/actions.go:254` (the declaration) and `pkg/engine/combat.go:12`.
6. **The two spellings are gone** (#112): `git -C W grep -n -e 'name: .Aggressive' -e ': Aggressive %d' -- pkg/engine` prints nothing and exits 1.
7. **Both read the one name** (#112): `git -C W grep -n -e 'KeywordAggressive.String' -- pkg/engine/card.go pkg/engine/trigger.go` prints exactly two lines: `pkg/engine/card.go:56` and `pkg/engine/trigger.go:137`.
8. **Only the listed edits:** `git -C W diff autoloop/m2-wave-d.1 --numstat -- pkg cmd` prints exactly these lines: `13	4	pkg/engine/actions.go`, `3	2	pkg/engine/card.go` and `1	1	pkg/engine/trigger.go`.
9. **Fences:** `go -C W run ./cmd/cardtown lint fences` exits 0.
10. **Nothing else changed:** `git -C W diff autoloop/m2-wave-d.1 --stat` lists only those three files and this task file.
11. `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/engine/actions.go`
- `pkg/engine/card.go`
- `pkg/engine/trigger.go`

## Done When
- [ ] Cases 1–11 hold, with their commands and results in the Log
- [ ] `go tool task sim-compare -- autoloop/m2-wave-d.1` passes (starter and stress sims byte-identical to the run branch)
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-05, lead-architect (wave D.1 run, session 5): wrote this task from board directives #111 and #112, next in board order, against the run branch at a26fcd5. No `main` commits to sync. Not `WAVE-COMPLETE`: #111–#119 and #121 are `open: before the next Deliverables task`. Harness, so it skips `tests` and goes straight to `ready`; tier `flash`, because the edits are exact and were dry-run. Every command here is on the headless allow-list, one per line, with no pipes, `$`, double quotes or unquoted parentheses. Searched for existing equivalents: the finds are in Inputs. Waiting directives checked: none has triggered (#6, #7: mode is `serial`; #8: no `deck.go`; #15: no turn-boundary hook; #71, #91 (2), #106, #118 (b), #120: no such card). Next: `kernel-engineer` on Flash.
