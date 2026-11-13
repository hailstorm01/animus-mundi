---
id: T-0230
title: "fireHook returns the entries it buffered (directive #146, wave F.1 harness, flash, pure move)"
kind: harness
tier: flash
wave: F.1
status: in-progress
branch: m2/autoloop-wave-f.1/T-0230-firehook-returns-added
depends_on: [T-0229]
claims: [func:fireHook (return value), func:fireMoved (field set), func:fireCombatKill (field set)]
cards: []
requirements: []
---

## Goal
"Fire a hook, then set a field on each trigger it buffered" is written once: `fireHook` returns the buffer entries it added (`g.pending[n:]`), and `fireMoved` and `fireCombatKill` set `locked` and `awaits` on what it returns, instead of each recording `len(g.pending)` and looping over the new tail themselves (#146). A pure move: same behavior, same test names, same sim output. #141 (3)'s locked placement zone and #15's turn-boundary hooks then reuse the return value instead of writing the loop a third time.

## Inputs
- Board directive #146 (structural review 11, finding 3, approved by the user, 2026-11-13), `open: before the next Deliverables task`; design note Deliverable 6j (wave F.1), third.
- At c4bd35b (the run branch after T-0229): `fireMoved` (`pkg/engine/trigger.go:94-100`) and `fireCombatKill` (`pkg/engine/trigger.go:117-123`) each read `len(g.pending)`, call `fireHook`, and loop from that index to the buffer's end. `fireHook` (`pkg/engine/trigger.go:134-165`) returns nothing. Its other callers (`fireDiscarded` :84, `fireDeploy` :89, `fireKilled` :105, `fireDamaged` :109, `fireExhausted` :112, `fireAttacks` :185) don't change: Go lets a call statement drop a result.
- The returned slice aliases the buffer, so setting a field through it sets it on the buffered entry. It is read right away, before anything else appends to `g.pending`. `fireHook` takes `n` before it can return early, and returns `nil` when the source prints no ability.
- `pkg/engine/invariants.go:64-65` reads `len(g.pending)` for the between-actions invariant. That is not a `fireHook` caller and stays.
- Reuse search (D2.3): no helper returns new buffer entries today; `bufferTrigger` appends one entry and stays as it is.
- The lead dry-ran these exact edits in a throwaway worktree of the run branch at c4bd35b: `gofmt -l pkg` clean, `go tool task check` exit 0, `go tool task sim-compare -- autoloop/m2-wave-f.1` `[PASS]`, `cardtown lint fences` exit 0, `go tool task sim` exit 0 (200 games, P0 63.5% / P1 35.0% / draws 1.5%). Every expected output below is from that dry run.

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)
Line numbers are at c4bd35b. There is no test commit: no test changes, and the existing tests are the pin. Commit the one file as one implementation commit.

### 1. `pkg/engine/trigger.go`: `fireMoved`
Lines 95-98,
```go
	n := len(g.pending)
	g.fireHook(assetSource(a), ability.HookMoves)
	for i := n; i < len(g.pending); i++ {
		g.pending[i].locked = &LockedZone{Zone: a.Zone, Player: a.Controller}
```
become
```go
	added := g.fireHook(assetSource(a), ability.HookMoves)
	for i := range added {
		added[i].locked = &LockedZone{Zone: a.Zone, Player: a.Controller}
```
The comment above and the closing `}` lines stay.

### 2. `pkg/engine/trigger.go`: `fireCombatKill`
Lines 118-121,
```go
	n := len(g.pending)
	g.fireHook(assetSource(att), ability.HookCombatKill)
	for i := n; i < len(g.pending); i++ {
		g.pending[i].awaits = def.ID
```
become
```go
	added := g.fireHook(assetSource(att), ability.HookCombatKill)
	for i := range added {
		added[i].awaits = def.ID
```
The comment above and the closing `}` lines stay.

### 3. `pkg/engine/trigger.go`: `fireHook`'s comment, signature and early return
Lines 134-140,
```go
// fireHook buffers src's own triggers for hook in printed order: the one place
// a trigger's gates are checked. A face-down asset has no printed abilities.
func (g *Game) fireHook(src triggerSource, hook ability.Hook) {
	def := src.card.Ability
	if def == nil {
		return
	}
```
become
```go
// fireHook buffers src's own triggers for hook in printed order: the one place
// a trigger's gates are checked. A face-down asset has no printed abilities.
// It returns the entries it added, a slice of the buffer, so a caller can set a
// field on each before the flush; the other callers ignore it.
func (g *Game) fireHook(src triggerSource, hook ability.Hook) []pendingTrigger {
	n := len(g.pending)
	def := src.card.Ability
	if def == nil {
		return nil
	}
```

### 4. `pkg/engine/trigger.go`: `fireHook` returns the new entries
Lines 163-165,
```go
		g.bufferTrigger(src, fmt.Sprintf("%s: trigger %d", src.card.Name, i+1), t)
	}
}
```
become
```go
		g.bufferTrigger(src, fmt.Sprintf("%s: trigger %d", src.card.Name, i+1), t)
	}
	return g.pending[n:]
}
```

No other file changes.

## Tests
No new tests and no test edits. Every existing test keeps its name and passes unchanged; the move-trigger and locked-zone tests (*Broken Horn Armorer*), the combat-kill tests that wait on a destruction (SVY-R19), and the byte-identical sims are the pin.

## Acceptance Cases (harness tasks only; written by the lead)
W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **No test changes** (#146): `git -C W diff autoloop/m2-wave-f.1 --stat -- '*_test.go'` prints nothing.
2. **Sims byte-identical** (#146): `go -C W tool task sim-compare -- autoloop/m2-wave-f.1` prints `[PASS] starter and stress sims match autoloop/m2-wave-f.1 byte for byte, exit codes included`.
3. **Only the listed edits:** `git -C W diff autoloop/m2-wave-f.1 --numstat -- pkg cmd` prints exactly one line: `12	10	pkg/engine/trigger.go`.
4. **No caller of `fireHook` reads the buffer's length** (#146): `git -C W grep -n -F 'len(g.pending)' -- pkg` prints exactly three lines: `pkg/engine/invariants.go:64:`, `pkg/engine/invariants.go:65:` (the between-actions invariant) and `pkg/engine/trigger.go:137:` (inside `fireHook`).
5. **Both callers use the return value:** `git -C W grep -n -F 'added := g.fireHook' -- pkg` prints exactly two lines: `pkg/engine/trigger.go:95:` and `pkg/engine/trigger.go:117:`.
6. **The signature:** `git -C W grep -n -F 'hook ability.Hook) []pendingTrigger {' -- pkg/engine` prints exactly one line, `pkg/engine/trigger.go:136:`.
7. **Fences:** `go -C W run ./cmd/cardtown lint fences` exits 0.
8. **Nothing else changed:** `git -C W diff autoloop/m2-wave-f.1 --stat` lists only `pkg/engine/trigger.go` and this task file.
9. `gofmt -l W/pkg` prints nothing; `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/engine/trigger.go`

## Done When
- [ ] Cases 1–9 hold, with their commands and results in the Log
- [ ] `go tool task sim-compare -- autoloop/m2-wave-f.1` passes (starter and stress sims byte-identical to the run branch)
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-13, lead-architect (wave F.1 run, session 4): written for #146 against c4bd35b (T-0229 merged; nothing new on `main` to sync). Harness, so it skips `tests` and goes straight to `ready`. Tier `flash`, because every edit is exact and was dry-run. Every command here is on the headless allow-list: one per line, no pipes, no `$`, no double quotes, no unquoted parentheses. Waiting directives checked: #6, #7 (mode `serial`), #15 (no turn-boundary hook), #71, #91 (2), #106, #120, #141 (3), #142: not fired. Next: `kernel-engineer` on Flash.
