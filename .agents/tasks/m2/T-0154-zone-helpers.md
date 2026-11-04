---
id: T-0154
title: Zone helpers live with the zones, deployedTo and battlefieldZone move to zones.go (directive #101, harness, flash, pure move)
kind: harness
tier: flash
wave: C.1 (board directive #101, structural review 7, finding 11)
status: in-progress
branch: m2/autoloop-wave-c.1/T-0154-zone-helpers
depends_on: [T-0153]
claims: [helper:deployedTo, helper:battlefieldZone]
cards: []
requirements: []
---

## Goal
`deployedTo` and `battlefieldZone` move from `pkg/engine/trigger.go` to `pkg/engine/zones.go`, unchanged (#101). They read zones, and `targeting.go` and `effects.go` call them as well as `trigger.go`. A pure move: same declarations, same test names, same sim output.

## Inputs
- Board directive #101 (structural review 7, approved by the user, 2026-11-03). Its line numbers predate T-0147 and T-0148; the table below has today's.
- Design note: `docs/architecture/m2-ability-model.md`, Deliverable 6f (wave C.1).
- Not in this task: the callers don't change, because the package stays the same. `zones.go`'s header line and `doc.go`'s `zones.go: moving cards and assets` stay as they are; #102 does the docs.
- Placement: `zones.go`, directly above `remove` and below `placeOnBattlefield`, which puts an asset into the zone these helpers read. `zones.go` gains the `ability` import. `trigger.go` keeps its own, because it still uses `ability.Trigger` and the hooks.
- Why it is a pure move: the two functions and their comments move as they are, with no edit to a line of them. The lead dry-ran these exact edits in a throwaway worktree of the run branch at 01dcbf3: `gofmt -l` clean, `go tool task check` passes, `go tool task sim` runs clean, and `go tool task sim-compare -- autoloop/m2-wave-c.1` printed `[PASS]`.

### Where it is today (at `autoloop/m2-wave-c.1`, 01dcbf3)
| Where | Today |
| :--- | :--- |
| `pkg/engine/trigger.go:132-151` | `deployedTo` and `battlefieldZone`, with their comments, followed by one blank line |
| `pkg/engine/trigger.go:93` | a `deployedTo` call, unchanged |
| `pkg/engine/targeting.go:278` | a `deployedTo` call, unchanged |
| `pkg/engine/effects.go:108` | a `battlefieldZone` call, unchanged |

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)

### 1. `pkg/engine/trigger.go`
Delete lines 132-152: from the comment `// deployedTo reports whether a is in zone z now, which for a deploy trigger is` through the blank line after `battlefieldZone`'s closing `}`. The comment `// fireAttacks buffers a's Step 2 triggers` then follows `cornered`'s closing `}` and one blank line.

### 2. `pkg/engine/zones.go`, imports
Replace the import block with:
```go
import (
	"fmt"
	"slices"

	"animus-mundi/pkg/engine/ability"
)
```

### 3. `pkg/engine/zones.go`, the two functions
Insert this text directly above the line `func remove[T comparable](s []T, v T) []T {`, with one blank line between its last `}` and that line:
```go
// deployedTo reports whether a is in zone z now, which for a deploy trigger is
// the zone its deploy put it in. ZoneAny always holds.
func deployedTo(a *Asset, z ability.Zone) bool {
	if to, ok := battlefieldZone(z); ok {
		return a.Zone == to
	}
	return true
}

// battlefieldZone maps a battlefield zone a shape names to the engine's zone.
// It is false for ZoneAny and ZoneDiscard, which are not battlefield zones.
func battlefieldZone(z ability.Zone) (Zone, bool) {
	switch z {
	case ability.ZoneReserves:
		return ZoneReserves, true
	case ability.ZoneMainTheater:
		return ZoneMainTheater, true
	}
	return 0, false
}
```

No other file changes.

## Acceptance Cases (harness tasks only; written by the lead)
No new tests: a pure move. W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **Existing tests unchanged** (#101): `git -C W diff autoloop/m2-wave-c.1 -- '*_test.go'` prints nothing, and `go -C W test ./...` passes.
2. **Sims byte-identical** (#101): `go -C W tool task sim-compare -- autoloop/m2-wave-c.1` prints `[PASS]`.
3. **`deployedTo` moved, callers unchanged** (#101): `git -C W grep -n 'deployedTo(' -- pkg/engine` prints exactly three lines: `pkg/engine/targeting.go:278`, `pkg/engine/trigger.go:93` and `pkg/engine/zones.go:237` (the declaration).
4. **`battlefieldZone` moved, callers unchanged** (#101): `git -C W grep -n 'battlefieldZone(' -- pkg/engine` prints exactly three lines: `pkg/engine/effects.go:108`, `pkg/engine/zones.go:238` (`deployedTo`'s call) and `pkg/engine/zones.go:246` (the declaration).
5. **Only the listed edits:** `git -C W diff autoloop/m2-wave-c.1 --numstat -- pkg` prints exactly these two lines: `0	21	pkg/engine/trigger.go` and `23	0	pkg/engine/zones.go`.
6. **Fences:** `go -C W run ./cmd/cardtown lint fences` exits 0.
7. **Nothing else changed:** `git -C W diff autoloop/m2-wave-c.1 --stat` lists only `pkg/engine/trigger.go`, `pkg/engine/zones.go` and this task file.
8. `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/engine/trigger.go`
- `pkg/engine/zones.go`

## Done When
- [ ] Cases 1–8 hold, with their commands and results in the Log
- [ ] `go tool task sim-compare -- autoloop/m2-wave-c.1` passes (starter and stress sims byte-identical to the run branch)
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-03, lead-architect (wave C.1 run, session 11): wrote this task from directive #101 against the run branch at 01dcbf3; `main` has no commits to sync. Harness, so it skips `tests` and goes straight to `ready`; tier `flash`, because the edits are exact and were dry-run. Every command here is on the headless allow-list, one per line, with no pipes, `$` or unquoted parentheses. Searched for existing equivalents: `pkg/engine` declares each of the two once, and nothing else maps an `ability.Zone` to an engine `Zone`. Waiting directives checked: none has triggered (#6, #7: mode is `serial`; #8: no `deck.go`; #15: no turn-boundary hook; #71, #91 (2): no such card). Next: `kernel-engineer` on Flash.
