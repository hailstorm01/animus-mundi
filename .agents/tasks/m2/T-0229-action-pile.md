---
id: T-0229
title: "One read of which pile an action names: Action.pile() (directive #145 (b), wave F.1 harness, flash, pure move)"
kind: harness
tier: flash
wave: F.1
status: in-progress
branch: m2/autoloop-wave-f.1/T-0229-action-pile
depends_on: [T-0228]
claims: [func:Action.pile, func:pileCard (ZoneDeck case), func:emitAction (pile read)]
cards: []
requirements: []
---

## Goal
"A Deploy takes its card from `a.From`, an Enact from the hand" is decided in one place, `func (a Action) pile() Zone`, and `pileCard` no longer reads `ZoneDeck` (`Action.From`'s zero value) as the hand (#145 (b)). A pure move: same behavior, same test names, same sim output. Split from T-0228 (#145 (a), `done`, merge 0b66395) by the D2.4 size rule.

## Inputs
- Board directive #145 (structural review 11, finding 2, approved by the user, 2026-11-13), `open: before the next Deliverables task`; design note Deliverable 6j (wave F.1), second.
- At b73a27f (the run branch after T-0228): the pile an action names is worked out in `Apply` (`pkg/engine/actions.go:65`, `:68`), `check` (`:123`, `:128`, `:137`, `:141`) and `emitAction` (`pkg/engine/events.go:158-168`). `pileCard` (`pkg/engine/zones.go:174-191`) reads `ZoneDeck` as the hand.
- **Lead correction to the directive's wording, to keep it a pure move** (recorded in #145's status cell): #145 says `pile()` returns `ZoneDiscard` for a discard Deploy "else `ZoneHand`". That would make a Deploy whose `From` names no pile (for example `ZoneMainTheater`) take its card from the hand, which `TestFR_TRN_003_DeployFromAnotherZoneIsIllegal` rejects today. So `pile()` returns `ZoneHand` for an Enact and for a Deploy whose `From` is the zero value, and a Deploy's `From` as it is otherwise; `pileCard` finds no card in a zone that is not a pile, as now.
- Outside this package, `cmd/cardtown/play_screen.go:164` reads `a.From == engine.ZoneDiscard` for a label and `pkg/eventlog/replay.go:275` writes `From`. Neither reads the zero value as the hand, `pile()` is unexported, and neither file is in scope; `doc.go` therefore says "inside this package".
- Reuse search (D2.3): `pileCard` and `pileName` stay as they are; `pile()` replaces three inline copies and adds none.
- The lead dry-ran these exact edits in a throwaway worktree of the run branch at b73a27f: `gofmt -l pkg` clean, `go tool task check` exit 0, `go tool task sim-compare -- autoloop/m2-wave-f.1` `[PASS]`, `cardtown lint fences` exit 0, `go tool task sim` exit 0 (200 games, P0 63.5% / P1 35.0% / draws 1.5%). Every expected output below is from that dry run.

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)
Line numbers are at b73a27f. There is no test commit: no test changes, and the existing tests are the pin. Commit the four files as one implementation commit.

### 1. `pkg/engine/actions.go`: `From`'s comment
Line 47, the `From` field. At the end of its comment, after `(a deploy permission, FR-TRN-003)`, append `; this package reads it only in pile`. The line becomes:
```go
	From     Zone     // Deploy: the pile the card comes from: the hand (the zero value, or ZoneHand) or ZoneDiscard (a deploy permission, FR-TRN-003); this package reads it only in pile
```

### 2. `pkg/engine/actions.go`: add `pile()` after the `Action` struct
After line 50 (the struct's closing `}`) and before the blank line that precedes `// Apply validates and executes a`, insert a blank line and then:
```go
// pile returns the pile a takes its card from: ZoneHand for an Enact and for a
// Deploy whose From is the zero value, else the Deploy's From, in which pileCard
// finds no card unless it is ZoneHand or ZoneDiscard (FR-TRN-003). It is the one
// reader of From in this package.
func (a Action) pile() Zone {
	if a.Kind == ActDeploy && a.From != ZoneDeck {
		return a.From
	}
	return ZoneHand
}
```
So the struct's `}` (50), a blank line (51), the comment (52-55), `func (a Action) pile() Zone {` (56) through its `}` (61), a blank line (62), then `// Apply validates and executes a, ...` (63).

### 3. `pkg/engine/actions.go`: `Apply` calls `a.pile()`
Old lines 65, 66 and 68 (76, 77 and 79 after Edit 2): replace each `a.From`, and the `ZoneHand` argument, with `a.pile()`:
```go
		inst := g.payAndTake(p, a.Card, a.pile())
		g.declare(deployPlacement(a.Player, inst, a.Zone, a.pile()), ability.Effect{}, StackDeploy)
	case ActEnact:
		inst := g.payAndTake(p, a.Card, a.pile())
```

### 4. `pkg/engine/actions.go`: `check` calls `a.pile()`
Four lines; nothing else on them changes (old line numbers; add 11 after Edit 2):
- 123: `inst, err := g.canPlay(p, a.Card, a.From)` → `inst, err := g.canPlay(p, a.Card, a.pile())`
- 128: `if a.From == ZoneDiscard && deployPermission(c) == nil {` → `if a.pile() == ZoneDiscard && deployPermission(c) == nil {`
- 137: `if !g.placeable(deployPlacement(a.Player, inst, a.Zone, a.From), ability.Effect{}) {` → `if !g.placeable(deployPlacement(a.Player, inst, a.Zone, a.pile()), ability.Effect{}) {`
- 141 (the `ActEnact` case): `inst, err := g.canPlay(p, a.Card, ZoneHand)` → `inst, err := g.canPlay(p, a.Card, a.pile())`

### 5. `pkg/engine/events.go`: `emitAction` calls `a.pile()`
Lines 158-161: delete the four lines
```go
		from := ZoneHand
		if a.Kind == ActDeploy {
			from = a.From
		}
```
and change the next line (162), `if inst := pileCard(g.Players[a.Player], a.Card, from); inst != nil {`, to
```go
		if inst := pileCard(g.Players[a.Player], a.Card, a.pile()); inst != nil {
```
Lines 167-168: change
```go
			if a.From == ZoneDiscard {
				e.From = a.From.String()
```
to
```go
			if a.pile() == ZoneDiscard {
				e.From = ZoneDiscard.String()
```
Nothing else in the function changes (the `if a.Kind == ActDeploy {` line and its comment stay).

### 6. `pkg/engine/zones.go`: `pileCard` drops its `ZoneDeck` case
Lines 175-176, the second and third lines of `pileCard`'s comment, become:
```go
// names: the hand or the discard pile (Action.pile says which one an action
// names). It is nil for any other zone, and for a card in another player's pile.
```
Line 180: `case ZoneDeck, ZoneHand:` → `case ZoneHand:`.

### 7. `pkg/engine/doc.go`: the Enums paragraph
Lines 115-116,
```go
// an unset marker, and ActionKind.String indexes an array by value. The shape
// enums in pkg/engine/ability follow that package's three rules instead.
```
become
```go
// an unset marker, and ActionKind.String indexes an array by value. One field
// is the exception: Action.From's zero value reads as the hand, and inside
// this package only Action.pile reads it. The shape enums in pkg/engine/ability
// follow that package's three rules instead.
```

No other file changes.

## Tests
No new tests and no test edits. Every existing test keeps its name and passes unchanged; the deploy, discard-deploy and enact tests (among them `TestFR_TRN_003_DeployFromAnotherZoneIsIllegal`), the event tests and the byte-identical sims are the pin.

## Acceptance Cases (harness tasks only; written by the lead)
W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **No test changes** (#145): `git -C W diff autoloop/m2-wave-f.1 --stat -- '*_test.go'` prints nothing.
2. **Sims byte-identical** (#145; events unchanged): `go -C W tool task sim-compare -- autoloop/m2-wave-f.1` prints `[PASS] starter and stress sims match autoloop/m2-wave-f.1 byte for byte, exit codes included`.
3. **Only the listed edits:** `git -C W diff autoloop/m2-wave-f.1 --numstat -- pkg cmd` prints exactly four lines: `19	8	pkg/engine/actions.go`, `4	2	pkg/engine/doc.go`, `3	7	pkg/engine/events.go` and `3	3	pkg/engine/zones.go`.
4. **The pile is decided only in `pile()`** (#145 (b)): `git -C W grep -n -F a.From -- pkg/engine` prints exactly four lines: `pkg/engine/actions.go:57:` and `pkg/engine/actions.go:58:` (inside `pile()`), and `pkg/engine/deploy_discard_test.go:47:` and `:297:` (unchanged tests).
5. **Every caller uses it:** `git -C W grep -n -F 'a.pile()' -- pkg` prints exactly nine lines: `pkg/engine/actions.go` lines 76, 77, 79, 134, 139, 148 and 152, and `pkg/engine/events.go` lines 158 and 163.
6. **`pileCard` reads no zero value:** `git -C W grep -n -F 'case ZoneDeck, ZoneHand' -- pkg` prints nothing.
7. **Fences:** `go -C W run ./cmd/cardtown lint fences` exits 0.
8. **Nothing else changed:** `git -C W diff autoloop/m2-wave-f.1 --stat` lists only the four files of case 3 and this task file.
9. `gofmt -l W/pkg` prints nothing; `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/engine/actions.go`
- `pkg/engine/events.go`
- `pkg/engine/zones.go`
- `pkg/engine/doc.go`

## Done When
- [ ] Cases 1–9 hold, with their commands and results in the Log
- [ ] `go tool task sim-compare -- autoloop/m2-wave-f.1` passes (starter and stress sims byte-identical to the run branch)
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-13, lead-architect (wave F.1 run, session 3): written out in full from #145 (b) against b73a27f (T-0228 merged; nothing new on `main` to sync). Harness, so it skips `tests` and goes straight to `ready`; tier `flash`, because every edit is exact and was dry-run. Every command here is on the headless allow-list, one per line, with no pipes, `$`, double quotes or unquoted parentheses. Waiting directives checked: #6, #7 (mode `serial`), #15 (no turn-boundary hook), #71, #91 (2), #106, #120, #141 (3), #142: not fired. Next: `kernel-engineer` on Flash.
- 2026-11-13, lead-architect (wave F.1 run, session 2): queued with T-0228, which takes #145 (a). Written out in full (Inputs, exact Edits, Acceptance Cases) when T-0228 is `done`.
