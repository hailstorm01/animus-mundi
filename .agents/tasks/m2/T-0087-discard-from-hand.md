---
id: T-0087
title: One place discards from hand, (g *Game) discardFromHand (directive #64, harness, flash, pure move)
kind: harness
tier: flash
wave: B.1 (board directive #64, structural review 4 finding 3)
status: ready
branch: m2/T-0087-discard-from-hand
depends_on: [T-0086]
claims: [func:discardFromHand]
cards: []
requirements: []
---

## Goal
"A player discards n cards of their choice from their own hand" is one function, `(g *Game) discardFromHand(pid PlayerID, n int) []*Instance`, in `zones.go` beside `draw` and `scratch`. It holds the `DecideHand` decision, the rule that the options are fixed before any card moves, the move and the `discard` event. `stepDiscard` calls it for each step player and counts the cards. Wave C's discard cost (*Steel Rain*) reuses it. This is a pure move: same tests, same sim output.

## Inputs
- Board directive #64 (structural review 4, finding 3).
- Design note: `docs/architecture/m2-ability-model.md`, Deliverable 6a (wave B.1).
- Rules the moved code implements, unchanged: `FR-MOD-001` §1, §6; `SVY-R16`.
- Why it is a pure move: `discardFromHand` runs the statements of `stepDiscard`'s loop body, in the same order (the clamp to the hand size, the empty check, one `g.choose(d)`, take every picked card before any moves, then move and log each). `stepDiscard` still computes `g.value` once, before the loop, so every decision and every random draw is the same. `git grep` over `pkg` and `cmd` finds no test that names `stepDiscard`.

### Where it is today (at `autoloop/m2-wave-b.1` e0811d7)
| Where | Today |
| :--- | :--- |
| `effects.go:167-199` | `stepDiscard`: for each step player, build the `DecideHand` decision, choose, take the picked cards, move each to the discard pile and log `discard` |
| `effects.go:5-10` | imports `fmt` (only `stepDiscard`'s `fmt.Sprintf` uses it) and `slices` (`stepCards`) |
| `zones.go:22-50` | `draw` and `scratch`, the other per-player card steps the effects call; `zones.go` already imports `fmt` |

## Edits (lead decisions; copy exactly)

### 1. `pkg/engine/zones.go`
Right after `scratch` (before `takeCard`'s doc comment), add:
```go

// discardFromHand has player pid discard n cards of their choice from their
// own hand, or the whole hand if it holds fewer, and returns the cards
// discarded, in pick order. Choosing from your own hand is not an inspection,
// so a hand of exactly that many cards is discarded without asking
// (FR-MOD-001 §1, §6; SVY-R16). Every discard from hand goes through it.
func (g *Game) discardFromHand(pid PlayerID, n int) []*Instance {
	hand := g.Players[pid].Hand
	n = min(n, len(hand))
	if n <= 0 {
		return nil
	}
	d := Decision{Player: pid, Kind: DecideHand, Prompt: "Discard a card", Min: n, Max: n}
	if n > 1 {
		d.Prompt = fmt.Sprintf("Discard %d cards", n)
	}
	for _, inst := range hand {
		d.Options = append(d.Options, Option{Label: inst.Card.Name, Ref: inst.ID})
	}
	// The options are fixed before any card moves, so take the cards first.
	var picked []*Instance
	for _, i := range g.choose(d) {
		picked = append(picked, hand[i])
	}
	for _, inst := range picked {
		g.takeCard(inst)
		g.putCard(inst, ZoneDiscard, false)
		g.logEvent(Event{Kind: "discard", Player: int(pid), Card: inst.Card.Name, Inst: inst.ID, Zone: ZoneHand.String()},
			"player %d discards %s", pid, inst.Card.Name)
	}
	return picked
}
```

### 2. `pkg/engine/effects.go`
- Replace `stepDiscard`, doc comment included, with:
  ```go
  // stepDiscard has each step player discard Amount cards of their choice from
  // their own hand (discardFromHand). It returns how many cards they discarded.
  func (g *Game) stepDiscard(r *resolution, st ability.Step) int {
  	n, total := g.value(r.it, st.Amount), 0
  	for _, pid := range stepPlayers(r.it, st) {
  		total += len(g.discardFromHand(pid, n))
  	}
  	return total
  }
  ```
- Remove `"fmt"` from the import block (nothing else in `effects.go` uses it). Keep `"slices"` and the `ability` import.

No other file changes.

## Acceptance Cases (harness tasks only; written by the lead)
No new tests: a pure move. Run each command on its own, pointed at your worktree (W). `git grep` ignores untracked files, so stage your edits with `git -C W add` before cases 3–5. Record each command and its result in the Log. The reviewer re-runs them.
1. **Existing tests unchanged:** `git -C W diff autoloop/m2-wave-b.1 -- '*_test.go'` prints nothing, and `go -C W test ./...` passes.
2. **Sims byte-identical:** `go -C W tool task sim-compare -- autoloop/m2-wave-b.1` prints `[PASS]`.
3. **One `discard` event** (directive #64 case (c)): `git -C W grep -n "Kind: \"discard\"" -- pkg/engine ":!*_test.go"` prints exactly one line, in `zones.go`'s `discardFromHand`.
4. **One hand decision for discards:** `git -C W grep -n "Kind: DecideHand" -- pkg/engine ":!*_test.go"` prints exactly one line, in `zones.go`'s `discardFromHand`.
5. **The step only calls it:** `git -C W grep -n "discardFromHand" -- pkg/engine` prints the `func discardFromHand` line, its doc comment's first line, `stepDiscard`'s doc comment line, and exactly one call, in `stepDiscard`.
6. **Only these edits:** `git -C W diff autoloop/m2-wave-b.1 --stat` lists only `pkg/engine/zones.go`, `pkg/engine/effects.go` and this task file.
7. `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/engine/zones.go`
- `pkg/engine/effects.go`

## Done When
- [ ] Cases 1–6 hold, with their commands and results in the Log
- [ ] `go tool task sim-compare -- autoloop/m2-wave-b.1` passes (starter and stress sims byte-identical to the run branch)
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks

## Log
- 2026-10-28, lead-architect (wave B.1 run, session 3): wrote the task from board directive #64. Harness, so it skips `tests` and goes straight to `ready`. Tier `flash` as #64 gives. The edits are spelled out so the engineer copies them. `discardFromHand` goes in `zones.go` beside `draw` and `scratch`, the other per-player card steps, so the cost task in wave C finds it there. It returns the cards, not a count, as #64 asks; the cost task widens it to "up to N". `main` has no commits the run branch lacks, so no sync. Directive #65 (with #67) stays open for the next session; #61 stays `waiting`, because no task's `cards:` has a modal card. Next: `kernel-engineer` (Flash, headless).
