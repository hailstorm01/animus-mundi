---
id: T-0146
title: One function deploys a card onto the battlefield, deployCard (directive #95, harness, flash, pure move)
kind: harness
tier: flash
wave: C.1 (board directive #95, structural review 7, finding 4)
status: done
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
- [x] Cases 1–9 hold, with their commands and results in the Log
- [x] `go tool task sim-compare -- autoloop/m2-wave-c.1` passes (starter and stress sims byte-identical to the run branch)
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-03, lead-architect (wave C.1 run, session 3): wrote this task from directive #95 against the run branch at f03e129, after syncing `main` (merge f03e129, `docs/PRODUCT_SUMMARY.md` only). Harness, so it skips `tests` and goes straight to `ready`; tier `flash`, because the edits are exact and were dry-run. Every command here is on the headless allow-list, one per line, with no pipes, `$` or unquoted parentheses. Searched for existing equivalents: `pkg/engine` has no `deployCard`; `resolveDeploy` is the only code that builds an `EventDeploy`. Waiting directives checked: none has triggered (#6, #7: mode is `serial`; #8: no `deck.go`; #15: no turn-boundary hook; #71, #91 (2): no such card). Next: `kernel-engineer` on Flash.
- 2026-11-03, kernel-engineer (session 1): implemented pure move of `resolveDeploy` body to `deployCard` helper in `pkg/engine/zones.go`.
  - Diff shows:
    - `pkg/engine/stack.go`: `resolveDeploy` delegates to `g.deployCard(it.Inst, it.Controller, it.Zone)` (-6 lines, +1 line).
    - `pkg/engine/zones.go`: added `(g *Game) deployCard(inst *Instance, controller PlayerID, zone Zone) *Asset` with doc comment directly above `scratch` (+14 lines).
  - Reuse: `newAsset`, `placeOnBattlefield`, `logEvent` with `EventDeploy`, `fireDeploy`. Searched `pkg/engine` for `deployCard`; no prior declaration existed. `scratch` preserved its face-down path unchanged.
  - Acceptance cases verified:
    - Case 1: `git -C W diff autoloop/m2-wave-c.1 -- "*_test.go"` returned empty; `go -C W test ./...` passed (exit 0).
    - Case 2: `go -C W tool task sim-compare -- autoloop/m2-wave-c.1` printed `[PASS] starter and stress sims match autoloop/m2-wave-c.1 byte for byte, exit codes included`.
    - Case 3: `git -C W grep -n "deployCard(" -- pkg/engine` returned 2 lines (`stack.go:119` and `zones.go:41`).
    - Case 4: `git -C W grep -n "Kind: EventDeploy" -- pkg/engine/stack.go pkg/engine/zones.go` returned 1 line (`zones.go:44`).
    - Case 5: `git -C W grep -n "g.newAsset(inst, pid, true)" -- pkg/engine/zones.go` returned 1 line (`zones.go:59`).
    - Case 6: `git -C W diff autoloop/m2-wave-c.1 --numstat -- pkg` returned `1	6	pkg/engine/stack.go` and `14	0	pkg/engine/zones.go`.
    - Case 7: `go -C W run ./cmd/cardtown lint fences` exited 0 (`[PASS]`).
    - Case 8: `git -C W diff autoloop/m2-wave-c.1 --stat` lists only `pkg/engine/stack.go`, `pkg/engine/zones.go` and `.agents/tasks/m2/T-0146-deploy-card.md`.
    - Case 9: `go -C W tool task check` passed all checks; `go -C W tool task sim` passed 200 starter deck games with all invariants holding.
  - Handoff commit `feat(engine): deployCard helper extracts resolveDeploy body (T-0146)`. Task status set to `review`.
- 2026-11-03, lead-architect (wave C.1 run, session 3, D5): **Approved, `done`.** Merged into `autoloop/m2-wave-c.1` as bac48cf. Flash run `c04af5b4-4af3-413a-8f00-a9324f2188d3`: exit 0, `status: SUCCESS`, 304 s, no refusals (no `denied_actions` key; the transcript has no step with `"status":"ERROR"`; its one `permission check failed` match is step 70, a `DONE` step quoting the runbook; the `.err` file holds only agy's "root agent idle" notice). Review of `git diff autoloop/m2-wave-c.1...m2/autoloop-wave-c.1/T-0146-deploy-card` (f378fe5):
  - Scope: only `stack.go`, `zones.go` and this file changed. No test file changed.
  - Pure move, checked against the diff: both code files match the task's Edits exactly. The four statements keep their order, with `it.Inst`, `it.Controller` and `it.Zone` passed in. Only this file's status, Log and Done When boxes changed besides.
  - Fences and placement: no card name, no new rule. `scratch` and `setup.go`'s Headquarters keep their own paths, as the Inputs say.
  - Re-run by the reviewer in a scratch worktree of the task branch, all as expected: cases 1-8; `cardtown lint fences` `[PASS]`; `go tool task check` passes (it runs every test); `go tool task sim` clean, 200 games, P0 82.5% / P1 17.5% / draws 0.0%, 42/50 and 32/50 main cards executable; `go tool task sim-compare -- autoloop/m2-wave-c.1` `[PASS]`.
  - The merge's `pkg/` matches the task tip's, so the gates hold on the run branch. Directive #95 is applied. Nothing is queued now; the next session takes #96 from the board.
