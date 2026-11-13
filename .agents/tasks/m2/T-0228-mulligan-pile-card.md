---
id: T-0228
title: "The mulligan finds its cards with pileCard, handCard goes, and pileCard moves to zones.go (directive #145 (a), wave F.1 harness, flash, pure move)"
kind: harness
tier: flash
wave: F.1
status: ready
branch: m2/autoloop-wave-f.1/T-0228-mulligan-pile-card
depends_on: []
claims: [func:handCard (delete), func:pileCard (move to zones.go)]
cards: []
requirements: []
---

## Goal
One lookup of a card in a player's pile (#145 (a)). The mulligan calls `pileCard(p, id, ZoneHand)`, `handCard` is deleted, and `pileCard` moves unchanged from `actions.go` to `zones.go`, right after `DiscardCard`. A pure move: same behavior, same test names, same sim output. #145 (b), `Action.pile()`, is T-0229, `queued` after this task (D2.4: five non-test files in one task would break the size rule).

## Inputs
- Board directive #145 (structural review 11, finding 2, approved by the user, 2026-11-13), `open: before the next Deliverables task`; design note Deliverable 6j (wave F.1), second.
- At abe2ea2 (the run branch after syncing `main`): `handCard` (`pkg/engine/zones.go:267-274`) is called only by the mulligan (`pkg/engine/setup.go:80`). `pileCard` (`pkg/engine/actions.go:188-205`) reads `ZoneHand` as `p.Hand`, so `pileCard(p, id, ZoneHand)` returns exactly what `handCard(p, id)` did.
- Reuse search (D2.3): `pileCard` is the existing lookup this task keeps; `DiscardCard` (`zones.go:163`) looks in every player's discard pile and stays. The task removes one copy and adds none.
- The lead dry-ran these exact edits in a throwaway worktree of the run branch at abe2ea2: `gofmt -l pkg` clean, `go tool task check` exit 0, `go tool task sim-compare -- autoloop/m2-wave-f.1` printed `[PASS]`. Every expected output below is from that dry run.

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)
Line numbers are at abe2ea2. There is no test commit: no test changes, and the existing tests are the pin. Commit the three files as one implementation commit.

### 1. `pkg/engine/actions.go`: remove `pileCard`
Delete lines 188-206: the three comment lines starting `// pileCard returns the card with instance ID id in the pile of p that from`, the whole `func pileCard` through its closing `}` (line 205), and the blank line after it (206). Line 187 (blank) is then followed directly by `// canPlay returns the card with instance ID id if p may play it from the pile`. Nothing else in the file changes.

### 2. `pkg/engine/zones.go`: delete `handCard`
Delete lines 266-274 (the blank line before `func handCard` and the whole function through its closing `}`). The file then ends with `remove`'s closing `}` on line 265.

### 3. `pkg/engine/zones.go`: add `pileCard` after `DiscardCard`
After line 172 (the closing `}` of `DiscardCard`), and before the blank line that precedes `// leaveBattlefield removes a from the battlefield`, insert a blank line and then the function exactly as it was in `actions.go`:
```go

// pileCard returns the card with instance ID id in the pile of p that from
// names: the discard pile, or the hand for ZoneHand and the zero value. It is
// nil for any other zone, and for a card in another player's pile.
func pileCard(p *Player, id int, from Zone) *Instance {
	var pile []*Instance
	switch from {
	case ZoneDeck, ZoneHand:
		pile = p.Hand
	case ZoneDiscard:
		pile = p.Discard
	}
	for _, inst := range pile {
		if inst.ID == id {
			return inst
		}
	}
	return nil
}
```
So `DiscardCard`'s `}` (172), a blank line (173), the comment (174-176), `func pileCard` (177) through its `}` (191), a blank line (192), then `// leaveBattlefield removes ...` (193).

### 4. `pkg/engine/setup.go`: the mulligan calls `pileCard`
Line 80: change `if inst := handCard(p, id); inst != nil {` to `if inst := pileCard(p, id, ZoneHand); inst != nil {`, and nothing else on the line.

No other file changes.

## Tests
No new tests and no test edits. Every existing test keeps its name and passes unchanged; the mulligan test (`TestFR_SET_002_PartialMulliganKeepsSixCards`), the deploy and enact tests, and the byte-identical sims are the pin.

## Acceptance Cases (harness tasks only; written by the lead)
W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **No test changes** (#145): `git -C W diff autoloop/m2-wave-f.1 --stat -- '*_test.go'` prints nothing.
2. **Sims byte-identical** (#145): `go -C W tool task sim-compare -- autoloop/m2-wave-f.1` prints `[PASS] starter and stress sims match autoloop/m2-wave-f.1 byte for byte, exit codes included`.
3. **Only the listed edits:** `git -C W diff autoloop/m2-wave-f.1 --numstat -- pkg cmd` prints exactly three lines: `0	19	pkg/engine/actions.go`, `1	1	pkg/engine/setup.go` and `19	9	pkg/engine/zones.go`.
4. **One lookup, in zones.go** (#145 (a)): `git -C W grep -n -e 'func handCard' -e 'func pileCard' -- pkg` prints exactly one line, `pkg/engine/zones.go:177:func pileCard(p *Player, id int, from Zone) *Instance {`.
5. **No caller of handCard:** `git -C W grep -n -w handCard -- pkg cmd` prints nothing.
6. **Fences:** `go -C W run ./cmd/cardtown lint fences` exits 0.
7. **Nothing else changed:** `git -C W diff autoloop/m2-wave-f.1 --stat` lists only the three files of case 3 and this task file.
8. `gofmt -l W/pkg` prints nothing; `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/engine/actions.go`
- `pkg/engine/zones.go`
- `pkg/engine/setup.go`

## Done When
- [ ] Cases 1–8 hold, with their commands and results in the Log
- [ ] `go tool task sim-compare -- autoloop/m2-wave-f.1` passes (starter and stress sims byte-identical to the run branch)
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-13, lead-architect (wave F.1 run, session 2): wrote this task from board directive #145 (a), after syncing `main`'s directives #149–#151 into the run branch. #145's whole fix touches five non-test files (`actions.go`, `zones.go`, `setup.go`, `events.go`, `doc.go`), over the D2.4 limit of four, so it is split: (a) here, (b) in T-0229, `queued`, with its own board row. Harness, so it skips `tests` and goes straight to `ready`; tier `flash`, because every edit is exact and was dry-run. Every command here is on the headless allow-list, one per line, with no pipes, `$`, double quotes or unquoted parentheses. Waiting directives checked: #6, #7 (mode `serial`), #15 (no turn-boundary hook), #71, #91 (2), #106, #120, #141 (3), #142: not fired. Next: `kernel-engineer` on Flash.
