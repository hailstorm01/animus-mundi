---
id: T-0176
title: The resource economy gets its own file, resources.go, and exhaust moves to state.go (directive #114, harness, flash, pure move)
kind: harness
tier: flash
wave: D.1 (board directive #114, structural review 8, finding 11)
status: ready
branch: m2/autoloop-wave-d.1/T-0176-resources-file
depends_on: [T-0175]
claims: [file:pkg/engine/resources.go]
cards: []
requirements: []
---

## Goal
`pkg/engine/actions.go` stops holding the resource economy. A new `pkg/engine/resources.go` holds `Resource`, `generate` and `convert` (extracted from `Apply`), `canYield`, and `canPay` and `pay` from `costs.go` (#114). `exhaust` moves to `state.go`, because `costs.go` and `combat.go` call it too. `EffectiveCost` and `payCost` stay in `costs.go`. A pure move: same behavior, same test names, same sim output.

## Inputs
- Board directive #114 (structural review 8, approved by the user, 2026-11-05).
- Design note: `docs/architecture/m2-ability-model.md`, Deliverable 6g (wave D.1).
- Reuse search (D2.3): no `resources.go`, `generate` or `convert` exists in `pkg/engine`. It found one more copy beyond #114's list, and this task folds it in: the conversion's rate `2` is written in `check` (`p.canPay(2, 0)` and its message) and in `Apply` (`p.pay(2, 0)`), which this move would put in two files. The new const `convertReq` is the one spelling in the engine, and the message's text doesn't change.
- Kept on purpose, not duplicates of the rule: `checkConvert` in `pkg/sim` restates the rate because a rule monitor is an independent check of the engine, and `cmd/cardtown/play_screen.go:162` is prompt text for a person.
- Why it is a pure move: the moved declarations are the same text in the same package. The bodies of `generate` and `convert` are `Apply`'s two cases, in the same order, and `Apply` still calls `emitAction` before them. `actions.go` still uses `errors`, `fmt` and `slices`, `costs.go` still uses `fmt` and `ability`, and `state.go`'s imports don't change. `resources.go` imports nothing.
- Not in this task: `doc.go`'s file map doesn't name `filter.go` or `resources.go` yet. #121 (c) fixes it last, so `doc.go` is not an Allowed File here.
- The lead dry-ran these exact edits in a throwaway worktree of the run branch at 14e8def: `gofmt -l` clean, `go tool task check` exits 0, `go tool task sim` exits 0, and `go tool task sim-compare -- autoloop/m2-wave-d.1` printed `[PASS]`. Every expected output in the cases below is from that dry run.

### Where it is today (at `autoloop/m2-wave-d.1`, 14e8def)
| Where | Today |
| :--- | :--- |
| `pkg/engine/actions.go:37-51` | `Resource`, its constants and `String` |
| `pkg/engine/actions.go:73-87` | the `ActGenerate` and `ActConvert` cases of `Apply` |
| `pkg/engine/actions.go:117-126` | `exhaust` |
| `pkg/engine/actions.go:154-155` | `check`'s conversion rate |
| `pkg/engine/actions.go:265-275` | `canYield` |
| `pkg/engine/costs.go:18-28` | `canPay` and `pay` |

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)
Line numbers are from the file as it stands at 14e8def. In `actions.go`, make the edits from the bottom up (edit 2 lists them in that order), so that earlier numbers stay right.

### 1. `pkg/engine/resources.go`, a new file
Its whole content, ending in one line break after the last `}`:
```go
// Resources: Requisition and Tactics, generating and converting them, and paying with them (FR-RES-001, FR-RES-003, FR-CAP-001).

package engine

// Resource selects Requisition or Tactics.
type Resource int

const (
	Requisition Resource = iota
	Tactics
)

// String names the resource, as events and the play screen write it.
func (r Resource) String() string {
	if r == Tactics {
		return "Tactics"
	}
	return "Requisition"
}

// convertReq is the Requisition one conversion to Tactics takes (FR-RES-003).
const convertReq = 2

// generate carries out a checked Generate: the asset exhausts and its controller
// gains the asset's printed yield of the chosen resource (FR-RES-001).
func (g *Game) generate(a Action) {
	p := g.Players[a.Player]
	as := g.Asset(a.Asset)
	g.exhaust(as)
	yield := as.Printed().TacYield
	if a.Resource == Requisition {
		yield = as.Printed().ReqYield
		p.Req += yield
	} else {
		p.Tac += yield
	}
	g.logEvent(Event{Kind: EventGenerate, Player: int(a.Player), Source: as.ID, Amount: yield},
		"%s exhausts for resources (R%d T%d)", as, p.Req, p.Tac)
}

// convert carries out a checked conversion (FR-RES-003).
func (p *Player) convert() {
	p.pay(convertReq, 0)
	p.Tac++
}

// canYield implements the Generate Resources rows of FR-CAP-001. A zero yield
// is still a legal choice (Design Rationale KIM-S03).
func (g *Game) canYield(a *Asset, r Resource) bool {
	if a.Printed().Type != TypeUnit || g.Keywords(a).Versatile {
		return true
	}
	if a.Zone == ZoneReserves {
		return r == Requisition
	}
	return r == Tactics
}

// canPay reports whether p holds req Requisition and tac Tactics, the one read
// of "can pay".
func (p *Player) canPay(req, tac int) bool {
	return p.Req >= req && p.Tac >= tac
}

// pay takes req Requisition and tac Tactics from p, the one payment.
func (p *Player) pay(req, tac int) {
	p.Req -= req
	p.Tac -= tac
}
```

### 2. `pkg/engine/actions.go`, from the bottom up
1. Delete lines 265-276: `canYield`, its comment and the blank line after it.
2. Replace lines 154-155 with:
   ```go
   		if !p.canPay(convertReq, 0) {
   			return illegal("conversion needs %d Requisition", convertReq)
   ```
3. Delete lines 117-127: `exhaust`, its comment and the blank line after it.
4. Replace lines 86-87 (`p.pay(2, 0)` and `p.Tac++`) with the one line:
   ```go
   		p.convert()
   ```
5. Replace lines 74-84 (the body of `case ActGenerate:`) with the one line:
   ```go
   		g.generate(a)
   ```
6. Delete lines 37-52: `Resource`, its constants, `String` and the blank line after it.

`Apply` keeps its `p := g.Players[a.Player]`, because other cases use it. The import block doesn't change.

### 3. `pkg/engine/costs.go`
Delete lines 18-29: `canPay`, `pay`, their comments and the blank line after each. The import block doesn't change.

### 4. `pkg/engine/state.go`
Replace line 1 with:
```go
// Core state: the game, player and asset types and their lookups, with last-known information (FR-ZON-004), and exhausting an asset.
```
Add to the end of the file, after `TheaterOf`'s closing brace and one blank line, ending in one line break:
```go
// exhaust is the one place a Readied asset becomes Exhausted, for any reason
// (SVY-R17). Entering the battlefield Exhausted is not an exhaust. Exhausting
// an asset that is already Exhausted changes nothing and fires nothing.
func (g *Game) exhaust(a *Asset) {
	if a.Exhausted {
		return
	}
	a.Exhausted = true
	g.fireExhausted(a)
}
```

No other file changes.

## Acceptance Cases (harness tasks only; written by the lead)
No new tests: a pure move, so no `test(T-0176)` commit. W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them. Stage the new file (`git -C W add pkg/engine/resources.go`) before running cases 3 and 4, because `git grep` and `git diff` don't see an untracked file.
1. **Existing tests unchanged** (#114): `git -C W diff autoloop/m2-wave-d.1 -- '*_test.go'` prints nothing, and `go -C W test ./...` passes.
2. **Sims byte-identical** (#114): `go -C W tool task sim-compare -- autoloop/m2-wave-d.1` prints `[PASS]`.
3. **The declarations, in their places** (#114): `git -C W grep -n -e 'type Resource int' -e 'convertReq = 2' -e ') generate(' -e ') convert(' -e ') canYield(' -e ') canPay(' -e ') pay(' -e ') exhaust(' -e ') payCost(' -e ') EffectiveCost(' -- pkg/engine` prints exactly ten lines: `pkg/engine/costs.go:14` and `:19`; `pkg/engine/resources.go:6`, `:22`, `:26`, `:42`, `:49`, `:61` and `:66`; and `pkg/engine/state.go:324`.
4. **Only the listed edits:** `git -C W diff autoloop/m2-wave-d.1 --numstat -- pkg cmd` prints exactly these lines: `4	54	pkg/engine/actions.go`, `0	12	pkg/engine/costs.go`, `69	0	pkg/engine/resources.go` and `12	1	pkg/engine/state.go`.
5. **Fences:** `go -C W run ./cmd/cardtown lint fences` exits 0.
6. **Nothing else changed:** `git -C W diff autoloop/m2-wave-d.1 --stat` lists only those four files and this task file.
7. `gofmt -l W/pkg/engine` prints nothing, and `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/engine/resources.go`
- `pkg/engine/actions.go`
- `pkg/engine/costs.go`
- `pkg/engine/state.go`

## Done When
- [ ] Cases 1–7 hold, with their commands and results in the Log
- [ ] `go tool task sim-compare -- autoloop/m2-wave-d.1` passes (starter and stress sims byte-identical to the run branch)
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-05, lead-architect (wave D.1 run, session 7): no `main` commits to sync. Wrote this task from board directive #114, next in board order. Not `WAVE-COMPLETE`: #114–#119 and #121 are `open: before the next Deliverables task`. #122 is `open: after Wave D.1 (#121), before Wave E`, so it doesn't apply yet. Harness, so it skips `tests` and goes straight to `ready`; tier `flash`, because the edits are exact and were dry-run. Every command here is on the headless allow-list, one per line, with no pipes, `$`, double quotes or unquoted parentheses. Searched for existing equivalents: the finds are in Inputs. Waiting directives checked: none has triggered (#6, #7: mode is `serial`; #8: no `deck.go`; #15: no turn-boundary hook; #71, #91 (2), #106, #118 (b), #120: no such card). Next: `kernel-engineer` on Flash.
