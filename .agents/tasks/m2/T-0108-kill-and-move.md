---
id: T-0108
title: Killing is one function, (g *Game) kill, and moving between battlefield zones is one, (g *Game) moveOnBattlefield (directives #83 and #84, harness, flash, pure move)
kind: harness
tier: flash
wave: G.2 (board directives #83 and #84, structural review 6, findings 2 and 3)
status: ready
branch: m2/T-0108-kill-and-move
depends_on: [T-0107]
claims: [func:kill, func:moveOnBattlefield]
cards: []
requirements: []
---

## Goal
"Kill these assets now as one moment" lives in one function, `(g *Game) kill(as []*Asset) int` in `sweep.go`, so wave C's kill costs call it instead of copying `stepKill` (#83). "Move an asset from one battlefield zone to another" lives in one function, `(g *Game) moveOnBattlefield(a *Asset, to Zone)` in `zones.go`, so wave C's `MOVE_ZONE` calls it instead of copying `resolveMarch` (#84). Both are pure moves: same tests, same sim output.

## Inputs
- Board directives #83 and #84 (structural review 6, approved by the user, 2026-10-30). #83 says the two may share one task; together they touch four non-test files.
- Design note: `docs/architecture/m2-ability-model.md`, Deliverable 6e (wave G.2).
- Rules the moved code implements, unchanged: `FR-REP-001`, `SVY-R15`, `SVY-R21` (a kill and its replacements as one moment); `FR-ZON-004` (damage, Armor and modifiers persist across battlefield zones), `SVY-R7` (a moved asset gets a new arrival number, which `placeOnBattlefield` already gives).
- Why it is a pure move: the kill body and the two move calls keep their order and arguments. In `resolveMarch`, `removeFromBattlefield` now runs after `dest` is computed rather than before; computing `dest` reads only `it.Zone`, so nothing observable changes. The lead dry-ran these exact edits on the run branch at 99ccec1: `go vet ./pkg/engine` clean, and `go tool task sim-compare -- autoloop/m2-wave-g.2` printed `[PASS]`.

### Where it is today (at `autoloop/m2-wave-g.2` 99ccec1)
| Where | Today |
| :--- | :--- |
| `pkg/engine/effects.go:160-175` | `stepKill`, its doc comment (160-165) and body (166-175) |
| `pkg/engine/sweep.go:5-23` | `destroyAll`, its doc comment and body; line 23 is its closing `}` |
| `pkg/engine/stack.go:113-127` | `resolveMarch`; line 118 is `g.removeFromBattlefield(a)`, 123-124 the `FR-ZON-004` comment and `g.placeOnBattlefield(a, dest)` |
| `pkg/engine/zones.go:146-157` | `removeFromBattlefield`; line 157 is its closing `}` |

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)

### 1. `pkg/engine/effects.go` (#83)
Replace lines 160-175 (from `// stepKill kills each step asset except a Headquarters, as one simultaneous event.` through `stepKill`'s closing `}`) with:
```go
// stepKill kills the step assets as one moment (kill). It returns how many it killed.
func (g *Game) stepKill(r *resolution, st ability.Step) int {
	return g.kill(g.stepAssets(r, st))
}
```

### 2. `pkg/engine/sweep.go` (#83)
After line 23 (`destroyAll`'s closing `}`), insert one blank line and then:
```go
// kill kills each of as except a Headquarters, as one simultaneous event.
// A death a replacement applies to is replaced instead, in a fixed order,
// before the rest are destroyed (replaceDeaths), with no sweep between: a source a
// byproduct leaves lethal applies to every death of this kill and is destroyed at
// the sweep after the item (SVY-R21). A replaced kill still
// counts as done (SVY-R15). It returns how many it killed.
func (g *Game) kill(as []*Asset) int {
	var dying []*Asset
	for _, a := range as {
		if a.Printed().Type != TypeHeadquarters {
			dying = append(dying, a)
		}
	}
	g.destroyAll(g.replaceDeaths(dying, func(*Asset) bool { return true }, usedReplacements{}))
	return len(dying)
}
```
The blank line that followed line 23 now follows `kill`'s closing `}`, before `// deathSweep handles …`. Nothing else in `sweep.go` changes.

### 3. `pkg/engine/stack.go` (#84)
In `resolveMarch`, replace these lines (118-124):
```go
	g.removeFromBattlefield(a)
	dest := ZoneMainTheater
	if it.Zone == ZoneMainTheater {
		dest = ZoneReserves
	}
	// Damage, Armor and modifiers persist across battlefield zones (FR-ZON-004).
	g.placeOnBattlefield(a, dest)
```
with:
```go
	dest := ZoneMainTheater
	if it.Zone == ZoneMainTheater {
		dest = ZoneReserves
	}
	g.moveOnBattlefield(a, dest)
```
`resolveMarch` keeps its `nil`/zone check and its `march` event unchanged.

### 4. `pkg/engine/zones.go` (#84)
After line 157 (`removeFromBattlefield`'s closing `}`), insert one blank line and then:
```go
// moveOnBattlefield moves a from its battlefield zone to the battlefield zone to.
// Damage, Armor and modifiers persist across battlefield zones (FR-ZON-004), and
// a gets a new arrival number on placement (SVY-R7).
func (g *Game) moveOnBattlefield(a *Asset, to Zone) {
	g.removeFromBattlefield(a)
	g.placeOnBattlefield(a, to)
}
```
The blank line that followed line 157 now follows `moveOnBattlefield`'s closing `}`, before `// newAsset builds …`. Nothing else in `zones.go` changes.

No other file changes. `doc.go` needs nothing: no file is added.

## Acceptance Cases (harness tasks only; written by the lead)
No new tests: a pure move. Run each command on its own, pointed at your worktree (W). Record each command and its result in the Log. The reviewer re-runs them.
1. **Existing tests unchanged** (#83 (a), #84 (a)): `git -C W diff autoloop/m2-wave-g.2 -- '*_test.go'` prints nothing, and `go -C W test ./...` passes.
2. **Sims byte-identical** (#83 (b), #84 (b)): `go -C W tool task sim-compare -- autoloop/m2-wave-g.2` prints `[PASS]`.
3. **`effects.go` no longer kills by itself** (#83 (c)): `git -C W grep -n -e replaceDeaths -e destroyAll -- pkg/engine/effects.go` prints nothing (git grep exits 1 when nothing matches; that is the expected result).
4. **Only the two moves detach an asset** (#84 (c)): `git -C W grep -n "removeFromBattlefield(" -- pkg` prints exactly five lines: `pkg/engine/max_per_turn_test.go` (one), `pkg/engine/scenario_test.go` (one) and `pkg/engine/zones.go` (three: the call in `leaveBattlefield`, the declaration, the call in `moveOnBattlefield`). No line in `stack.go` or any other non-test file.
5. **The two functions exist where the board puts them:** `git -C W grep -n -e "func (g \*Game) kill(" -e "func (g \*Game) moveOnBattlefield(" -- pkg/engine` prints exactly two lines, one in `pkg/engine/sweep.go` and one in `pkg/engine/zones.go`.
6. **Moved, not changed:** `git -C W diff autoloop/m2-wave-g.2 --numstat -- pkg/engine` prints exactly these four lines: `2	14	pkg/engine/effects.go`, `1	3	pkg/engine/stack.go`, `17	0	pkg/engine/sweep.go` and `8	0	pkg/engine/zones.go`.
7. **Fences** (#83 (d)): `go -C W run ./cmd/cardtown lint fences` exits 0.
8. **Only these edits:** `git -C W diff autoloop/m2-wave-g.2 --stat` lists only the four files in case 6 and this task file.
9. `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/engine/effects.go`
- `pkg/engine/stack.go`
- `pkg/engine/sweep.go`
- `pkg/engine/zones.go`

## Done When
- [ ] Cases 1–9 hold, with their commands and results in the Log
- [ ] `go tool task sim-compare -- autoloop/m2-wave-g.2` passes (starter and stress sims byte-identical to the run branch)
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks

## Log
- 2026-10-30, lead-architect (wave G.2 run, session 2): wrote the task from board directives #83 and #84, next in board order now that #82 is applied (T-0107). One task for both, as #83 allows: four non-test files, within the size rule. Harness, so it skips `tests` and goes straight to `ready`; tier `flash` as both directives give. Every command here is on the headless allow-list, one per line, with no `|`, `$` or unquoted parentheses. #84 (c) is read as "no non-test file but `zones.go` calls `removeFromBattlefield`": two existing tests call it directly for setup, and case 4 lists them. The `FR-ZON-004` comment moves from `resolveMarch` into `moveOnBattlefield`'s doc comment, which also names `SVY-R7`, as #84 asks. The edits were dry-run on the run branch (Inputs). `main` (281e7d8) has no commits the run branch lacks, so no sync. Waiting directives checked: none has triggered (#6, #7: mode is `serial`; #8: no `deck.go`; #15, #20 (b), #58, #59 (2), #61, #71: no such hook, claim or card here). #85, #86 and #87 (c) stay `open: before the next Deliverables task`; #85 is next. Next: `kernel-engineer` on Flash.
