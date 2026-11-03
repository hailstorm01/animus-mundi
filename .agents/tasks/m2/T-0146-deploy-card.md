---
id: T-0146
title: One function deploys a card onto the battlefield, deployCard (directive #95, harness, flash, pure move)
kind: harness
tier: flash
wave: C.1 (board directive #95, structural review 7, finding 4)
status: ready
branch: m2/autoloop-wave-c.1/T-0146-deploy-card
depends_on: [T-0145]
claims: [helper:deployCard]
cards: []
requirements: []
---

## Goal
`resolveDeploy`'s body becomes `(g *Game) deployCard(inst *Instance, controller PlayerID, zone Zone) *Asset` in `pkg/engine/zones.go`, and `resolveDeploy` calls it (#95). Wave F's free deploys reuse it. A pure move: same behavior, same test names, same sim output.

## Inputs
- Board directive #95 (structural review 7, approved by the user, 2026-11-03).
- Design note: `docs/architecture/m2-ability-model.md`, Deliverable 6f (wave C.1).
- Rule the function serves, unchanged: `FR-TRN-003a` (in-resolution free deploys enter during resolution, with no stack item and no costs). No free-deploy card is in this task: *Reaver's Rush*, *Allshan's Battle Trance*, *Devilhorn Scouting Party* and *Eerie Remnant* are wave F.
- Not in this task: `scratch` keeps its own path, because it deploys face-down and logs `EventScratch` (#95). `setup.go`'s Headquarters keeps its own path too: it enters during setup with no deploy event or trigger. Both still build their asset with `newAsset`.
- Placement: `deployCard` goes in `zones.go` beside `scratch`, `newAsset` and `placeOnBattlefield`, because that file moves cards and assets. Wave F's effect step will call it from outside `stack.go`.
- Its `*Asset` result has no reader yet. #95 fixes the signature, for wave F's step.
- Why it is a pure move: the four statements keep their order, with `it.Inst`, `it.Controller` and `it.Zone` passed in as `inst`, `controller` and `zone`, and `c.Name` read as `inst.Card.Name`. The lead dry-ran these exact edits in a throwaway worktree of the run branch at f03e129: `gofmt -l` clean, `go tool task check` passes, `go tool task sim` runs clean, and `go tool task sim-compare -- autoloop/m2-wave-c.1` printed `[PASS]`.

### Where it is today (at `autoloop/m2-wave-c.1`, f03e129)
| Where | Today |
| :--- | :--- |
| `pkg/engine/stack.go:118-125` | `resolveDeploy`, the only place that builds an `EventDeploy` |
| `pkg/engine/zones.go:36` | the doc comment of `scratch`; `deployCard` goes directly above it |

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)

### 1. `pkg/engine/stack.go`
Replace the body of `resolveDeploy` (lines 119-124), so that the function reads:
```go
func (g *Game) resolveDeploy(it *StackItem) {
	g.deployCard(it.Inst, it.Controller, it.Zone)
}
```

### 2. `pkg/engine/zones.go`
Insert this text directly above line 36 (`// scratch performs Deploy Scratch Company n times (FR-ZON-005, FR-KWD-002m).`), with one blank line between the closing `}` and that comment:
```go
// deployCard puts inst, already out of its pile, onto the battlefield face-up in
// zone under controller, with its own deploy event and deploy triggers, and
// returns the new asset. A free deploy enters this way during resolution, with
// no stack item and no costs (FR-TRN-003a). scratch has its own path, because it
// deploys face-down and logs its own event.
func (g *Game) deployCard(inst *Instance, controller PlayerID, zone Zone) *Asset {
	a := g.newAsset(inst, controller, false)
	g.placeOnBattlefield(a, zone)
	g.logEvent(Event{Kind: EventDeploy, Player: int(controller), Card: inst.Card.Name, Inst: inst.ID, Source: a.ID, Zone: zone.String()},
		"player %d deploys %s to %s", controller, a, zone)
	g.fireDeploy(a)
	return a
}
```

No other file changes.

## Acceptance Cases (harness tasks only; written by the lead)
No new tests: a pure move. W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **Existing tests unchanged** (#95): `git -C W diff autoloop/m2-wave-c.1 -- '*_test.go'` prints nothing, and `go -C W test ./...` passes.
2. **Sims byte-identical** (#95): `go -C W tool task sim-compare -- autoloop/m2-wave-c.1` prints `[PASS]`.
3. **One declaration, one caller** (#95): `git -C W grep -n 'deployCard(' -- pkg/engine` prints exactly two lines: `pkg/engine/stack.go:119` (the call) and `pkg/engine/zones.go:41` (the declaration).
4. **One `EventDeploy` built:** `git -C W grep -n 'Kind: EventDeploy' -- pkg/engine/stack.go pkg/engine/zones.go` prints exactly one line, at `pkg/engine/zones.go:44`.
5. **`scratch` keeps its own path** (#95): `git -C W grep -n 'g.newAsset(inst, pid, true)' -- pkg/engine/zones.go` prints exactly one line, at `pkg/engine/zones.go:59`.
6. **Only the listed edits:** `git -C W diff autoloop/m2-wave-c.1 --numstat -- pkg` prints exactly these two lines: `1	6	pkg/engine/stack.go` and `14	0	pkg/engine/zones.go`.
7. **Fences:** `go -C W run ./cmd/cardtown lint fences` exits 0.
8. **Nothing else changed:** `git -C W diff autoloop/m2-wave-c.1 --stat` lists only the two files in case 6 and this task file.
9. `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/engine/stack.go`
- `pkg/engine/zones.go`

## Done When
- [ ] Cases 1–9 hold, with their commands and results in the Log
- [ ] `go tool task sim-compare -- autoloop/m2-wave-c.1` passes (starter and stress sims byte-identical to the run branch)
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-03, lead-architect (wave C.1 run, session 3): wrote this task from directive #95 against the run branch at f03e129, after syncing `main` (merge f03e129, `docs/PRODUCT_SUMMARY.md` only). Harness, so it skips `tests` and goes straight to `ready`; tier `flash`, because the edits are exact and were dry-run. Every command here is on the headless allow-list, one per line, with no pipes, `$` or unquoted parentheses. Searched for existing equivalents: `pkg/engine` has no `deployCard`; `resolveDeploy` is the only code that builds an `EventDeploy`. Waiting directives checked: none has triggered (#6, #7: mode is `serial`; #8: no `deck.go`; #15: no turn-boundary hook; #71, #91 (2): no such card). Next: `kernel-engineer` on Flash.
