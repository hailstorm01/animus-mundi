---
id: T-0224
title: "A trigger fires from a trigger source, so one fireHook checks every trigger's gates (directive #136, wave F harness, flash, pure move)"
kind: harness
tier: flash
wave: F
status: ready
branch: m2/autoloop-wave-f/T-0224-trigger-source
depends_on: []
claims: [type:triggerSource, func:assetSource, func:fireHook (signature), func:bufferTrigger (signature), func:deployedTo (signature)]
cards: []
requirements: []
---

## Goal
`fireHook` and `bufferTrigger` take a `triggerSource` instead of an `*Asset` (#136). The source holds what firing reads: the printed face, the source ID, the card instance, the controller, the zone and the rank in the flush's source order. `assetSource` builds one from a battlefield asset, and every caller uses it. So one function still checks a trigger's gates (hook, `Turn`, `ToZone`, `Cornered`, `MaxPerTurn`) and buffers it, and T-0225 (*Eerie Remnant*) adds a source built from a card in a pile without copying them. `deployedTo` takes the zone, not the asset, because a source has no asset. A pure move: same behavior, same test names, same sim output.

## Inputs
- Board directive #136 (structural review 10, finding 1, approved by the user, 2026-11-09). Its trigger fired in this task's commit: T-0225, whose `cards:` include *Eerie Remnant* ("When this card is put into your discard pile from your hand, you may deploy it from your discard pile for free."), is now written and `queued`, so #136 is `open: before T-0225`.
- Why it is a pure move: `fireHook` and `bufferTrigger` have no callers outside `trigger.go`, and no test calls either of them, `posOf` or `deployedTo` (`git grep -n -e deployedTo -e fireHook -e bufferTrigger -e posOf -- '*_test.go'` prints nothing at d67fff4). Every source is built at the moment the old code read the asset (`fireAttacks` builds one before its Aggressive trigger and uses it for both, and buffering changes nothing on the asset), so every field holds the same value the old reads did. `deployedTo` has one other caller, `matchesFilterAt` (`filter.go:152`), which passes `a.Zone`.
- Reuse search (D2.3): no trigger source type exists (`git grep -n -i -e 'triggerSource' -e 'trigSource' -- pkg`: nothing). This task removes the asset assumption from the only copy of the gates, so it adds no duplicate.
- The lead dry-ran these exact edits in a throwaway worktree of the run branch at d67fff4: `gofmt -l pkg` clean, `go tool task check` exit 0, `go tool task sim-compare -- autoloop/m2-wave-f` printed `[PASS]`, `cardtown lint fences` exit 0, and `go tool task sim` exit 0 (200 games, P0 60.5% / P1 38.5% / draws 1.0%, P0 50/50 and P1 47/50 main cards executable, every invariant holding). Every expected output below is from that dry run.

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)
Line numbers are at d67fff4. There is no test commit: no test changes, and the existing tests are the pin. Commit the three files as one implementation commit.

### 1. `pkg/engine/trigger.go`, the source type
After line 41 (the closing `}` of `sourcePos.compare`), insert one blank line and then:
```go
// triggerSource is what fires a trigger, read when it fires: the face it
// prints, its controller, its zone and its rank in the flush's source order,
// and the IDs its placement and MaxPerTurn counter use. Every source so far
// is a battlefield asset (assetSource).
type triggerSource struct {
	card       *Card     // its printed face (the Scratch face if face-down)
	id         int       // source asset ID: the placement's source and the MaxPerTurn key
	inst       *Instance // source card instance (FR-ZON-004)
	controller PlayerID
	zone       Zone
	pos        sourcePos
}

// assetSource is a as a trigger source, as it stands now.
func assetSource(a *Asset) triggerSource {
	return triggerSource{card: a.Printed(), id: a.ID, inst: a.Inst, controller: a.Controller, zone: a.Zone, pos: posOf(a)}
}
```
so that the existing blank line 42 separates it from the `fireDeploy` comment.

### 2. `pkg/engine/trigger.go`, the callers
In each of these lines, change `g.fireHook(a, ` to `g.fireHook(assetSource(a), ` (line 75: `g.fireHook(att, ` to `g.fireHook(assetSource(att), `), and nothing else on the line:
- line 45 (`fireDeploy`), line 52 (`fireMoved`), line 61 (`fireKilled`), line 65 (`fireDamaged`), line 68 (`fireExhausted`), line 75 (`fireCombatKill`).

### 3. `pkg/engine/trigger.go`, `fireHook`
Replace lines 90-93:
```go
// fireHook buffers a's own triggers for hook in printed order. A face-down
// asset has no printed abilities.
func (g *Game) fireHook(a *Asset, hook ability.Hook) {
	def := a.Printed().Ability
```
with:
```go
// fireHook buffers src's own triggers for hook in printed order: the one place
// a trigger's gates are checked. A face-down asset has no printed abilities.
func (g *Game) fireHook(src triggerSource, hook ability.Hook) {
	def := src.card.Ability
```
Then, in the same function:
- line 98: `g.Active != a.Controller)` → `g.Active != src.controller)`
- line 101: `!deployedTo(a, t.ToZone)` → `!deployedTo(src.zone, t.ToZone)`
- line 106: `!g.cornered(a.Controller)` → `!g.cornered(src.controller)`
- line 110: `firingKey{asset: a.ID,` → `firingKey{asset: src.id,`
- line 119: `g.bufferTrigger(a, fmt.Sprintf("%s: trigger %d", a.Printed().Name, i+1), t)` → `g.bufferTrigger(src, fmt.Sprintf("%s: trigger %d", src.card.Name, i+1), t)`

### 4. `pkg/engine/trigger.go`, `bufferTrigger`
Replace lines 123-129:
```go
// bufferTrigger adds the fired trigger t to the buffer: a is its source, a's
// controller now is its controller, and label names it in the flush's prompts.
func (g *Game) bufferTrigger(a *Asset, label string, t ability.Trigger) {
	g.pending = append(g.pending, pendingTrigger{
		placement: placement{source: a.ID, inst: a.Inst, controller: a.Controller, label: label},
		trig:      t,
		pos:       posOf(a),
```
with:
```go
// bufferTrigger adds the fired trigger t to the buffer: src is its source, src's
// controller is its controller, and label names it in the flush's prompts.
func (g *Game) bufferTrigger(src triggerSource, label string, t ability.Trigger) {
	g.pending = append(g.pending, pendingTrigger{
		placement: placement{source: src.id, inst: src.inst, controller: src.controller, label: label},
		trig:      t,
		pos:       src.pos,
```

### 5. `pkg/engine/trigger.go`, `fireAttacks`
Replace lines 137-140:
```go
	if n := g.Keywords(a).Aggressive; n > 0 {
		g.bufferTrigger(a, fmt.Sprintf("%s: %s %d", a.Printed().Name, ability.KeywordAggressive.String(), n), ability.Aggressive(n))
	}
	g.fireHook(a, ability.HookAttacks)
```
with:
```go
	src := assetSource(a)
	if n := g.Keywords(a).Aggressive; n > 0 {
		g.bufferTrigger(src, fmt.Sprintf("%s: %s %d", src.card.Name, ability.KeywordAggressive.String(), n), ability.Aggressive(n))
	}
	g.fireHook(src, ability.HookAttacks)
```

### 6. `pkg/engine/zones.go`, `deployedTo`
Replace lines 236-240:
```go
// deployedTo reports whether a is in zone z now, which for a deploy trigger is
// the zone its deploy or move put it in. ZoneAny always holds.
func deployedTo(a *Asset, z ability.Zone) bool {
	if to, ok := battlefieldZone(z); ok {
		return a.Zone == to
```
with:
```go
// deployedTo reports whether an asset in zone at is in zone z, which for a deploy
// trigger is the zone its deploy or move put it in. ZoneAny always holds.
func deployedTo(at Zone, z ability.Zone) bool {
	if to, ok := battlefieldZone(z); ok {
		return at == to
```

### 7. `pkg/engine/filter.go`, `matchesFilterAt`
Line 152: change `!deployedTo(a, f.Zone)` to `!deployedTo(a.Zone, f.Zone)`, and nothing else on the line.

No other file changes.

## Tests
No new tests and no test edits. Every existing test keeps its name and passes unchanged; the trigger tests (`TestFR_TRG_*`, the card tests of every triggered card) and the byte-identical sims are the pin.

## Acceptance Cases (harness tasks only; written by the lead)
W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **No test changes** (#136): `git -C W diff autoloop/m2-wave-f --stat -- '*_test.go'` prints nothing.
2. **Sims byte-identical** (#136): `go -C W tool task sim-compare -- autoloop/m2-wave-f` prints `[PASS] starter and stress sims match autoloop/m2-wave-f byte for byte, exit codes included`.
3. **Only the listed edits:** `git -C W diff autoloop/m2-wave-f --numstat -- pkg cmd` prints exactly three lines: `1	1	pkg/engine/filter.go`, `41	22	pkg/engine/trigger.go` and `4	4	pkg/engine/zones.go`.
4. **One place reads the asset:** `git -C W grep -n -e 'Printed()' -- pkg/engine/trigger.go` prints exactly one line, `pkg/engine/trigger.go:58` (the `return triggerSource{card: a.Printed(), ...}` line of `assetSource`).
5. **Every firing goes through a source:** `git -C W grep -n -e 'fireHook(' -e 'bufferTrigger(' -- pkg cmd ':!*_test.go'` prints exactly eleven lines, all in `pkg/engine/trigger.go`: `:63`, `:70`, `:79`, `:83`, `:86` and `:93` (each `g.fireHook(assetSource(a), ...` or, at `:93`, `assetSource(att)`), `:110` (`func (g *Game) fireHook(src triggerSource, hook ability.Hook) {`), `:137` (`g.bufferTrigger(src, ...`), `:143` (`func (g *Game) bufferTrigger(src triggerSource, label string, t ability.Trigger) {`), `:157` (`g.bufferTrigger(src, ...` in `fireAttacks`) and `:159` (`g.fireHook(src, ability.HookAttacks)`).
6. **Fences:** `go -C W run ./cmd/cardtown lint fences` exits 0.
7. **Nothing else changed:** `git -C W diff autoloop/m2-wave-f --stat` lists only the three files of case 3 and this task file.
8. `gofmt -l W/pkg` prints nothing; `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/engine/trigger.go`
- `pkg/engine/zones.go`
- `pkg/engine/filter.go`

## Done When
- [ ] Cases 1–8 hold, with their commands and results in the Log
- [ ] `go tool task sim-compare -- autoloop/m2-wave-f` passes (starter and stress sims byte-identical to the run branch)
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-12, lead-architect (wave F run, session 26): wrote this task from board directive #136, whose trigger fired when T-0225 (*Eerie Remnant*, the last header card on `NotYetExecutable`) was written `queued` in the same commit. #136 is `open: before T-0225`, so this task goes first. The card task then adds a source built from a card in its owner's discard pile, controlled by its owner and ranked after every battlefield asset (ADR-0002's fixed order), and fires it through this `fireHook`. Harness, so it skips `tests` and goes straight to `ready`; tier `flash`, because every edit is exact and was dry-run. Every command here is on the headless allow-list, one per line, with no pipes, `$`, double quotes or unquoted parentheses. Waiting directives checked: #6, #7 (mode `serial`), #15 (no turn-boundary hook), #71, #91 (2), #106, #120, #141 (3), #142: not fired. Next: `kernel-engineer` on Flash.
