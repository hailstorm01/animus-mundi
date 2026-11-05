---
id: T-0171
title: One function plays a card from hand, playFromHand, and one emitAction case (directive #108, harness, flash, pure move)
kind: harness
tier: flash
wave: D.1 (board directive #108, structural review 8, finding 5)
status: in-progress
branch: m2/autoloop-wave-d.1/T-0171-play-from-hand
depends_on: [T-0170]
claims: [helper:Game.playFromHand]
cards: []
requirements: []
---

## Goal
Finding a card in hand, paying its cost and taking it lives in one place: `(*Game).playFromHand(p *Player, id int) *Instance` in `pkg/engine/actions.go`, called by both `Apply` cases. `emitAction` has one case for Deploy and Enact (#108). A pure move: same behavior, same test names, same sim output.

## Inputs
- Board directive #108 (structural review 8, approved by the user, 2026-11-05).
- Design note: `docs/architecture/m2-ability-model.md`, Deliverable 6g (wave D.1).
- Placement: `playFromHand` goes right after `handPlay`, the check that approves the same card. `handPlay` stays as it is: it reports why a card can't be played, and `playFromHand` runs only after `check` has passed.
- Not in this task: `setup.go:80-81` also looks up a hand card and takes it, but a mulligan pays nothing and isn't a play. #109 changes the `g.push` and `g.declare` lines these cases keep.
- Why it is a pure move:
  - `playFromHand` runs the same three calls in the same order, with the same arguments, and `Apply` does what it did before with the returned card.
  - In `emitAction`, Deploy still sets `Card`, `Inst` and `Zone`, and Enact still sets `Card` and `Inst` and leaves `Zone` empty. The new `if` keeps it so: `Zone`'s zero value is `ZoneDeck`, so setting it for an Enact would write a zone into the event.
- No test pins an Enact event's empty `Zone` (`TestEvents_Case4_EnactActionEventNamesItsCard` reads only `Player`, `Card` and `Inst`), so the reviewer checks that line against the diff.
- The lead dry-ran these exact edits in a throwaway worktree of the run branch at 35cd48c: `gofmt -l` clean, `go tool task check` passes, `go tool task sim` runs clean, and `go tool task sim-compare -- autoloop/m2-wave-d.1` printed `[PASS]`. Every expected output in the cases below is from that dry run.

### Where it is today (at `autoloop/m2-wave-d.1`, engine code as at 35cd48c)
| Where | Today |
| :--- | :--- |
| `pkg/engine/actions.go:89-91` | Deploy finds, pays and takes in `Apply` |
| `pkg/engine/actions.go:94-96` | Enact finds, pays and takes in `Apply` |
| `pkg/engine/events.go:155-163` | `emitAction`'s two cases |

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)

### 1. `pkg/engine/actions.go`, `Apply`, `case ActDeploy:`
Replace lines 89-91, `inst := handCard(p, a.Card)`, `g.payCost(p, inst)` and `g.takeCard(inst)`, with this line:
```go
		inst := g.playFromHand(p, a.Card)
```
The `g.push` line stays.

### 2. `pkg/engine/actions.go`, `Apply`, `case ActEnact:`
Replace lines 94-96, the same three lines, with this line:
```go
		inst := g.playFromHand(p, a.Card)
```
The `g.declare` lines stay.

### 3. `pkg/engine/actions.go`, the new function
Between the closing brace of `handPlay` and `func (g *Game) readyAsset`, add this, with one blank line on each side:
```go
// playFromHand finds the card handPlay approved, pays its cost and takes it from
// hand: what Deploy and Enact do before each goes on in its own way.
func (g *Game) playFromHand(p *Player, id int) *Instance {
	inst := handCard(p, id)
	g.payCost(p, inst)
	g.takeCard(inst)
	return inst
}
```

### 4. `pkg/engine/events.go`, `emitAction`
Replace lines 155-163, from `case ActDeploy:` through the closing brace of the `if` in `case ActEnact:`, with:
```go
	case ActDeploy, ActEnact:
		if inst := handCard(g.Players[a.Player], a.Card); inst != nil {
			e.Card, e.Inst = inst.Card.Name, inst.ID
		}
		if a.Kind == ActDeploy { // Zone's zero value is ZoneDeck: an Enact leaves it empty
			e.Zone = a.Zone.String()
		}
```
The closing brace of the `switch` and `g.emit(e)` stay.

No other file changes.

## Acceptance Cases (harness tasks only; written by the lead)
No new tests: a pure move, so no `test(T-0171)` commit. W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **Existing tests unchanged** (#108): `git -C W diff autoloop/m2-wave-d.1 -- '*_test.go'` prints nothing, and `go -C W test ./...` passes.
2. **Sims byte-identical** (#108): `go -C W tool task sim-compare -- autoloop/m2-wave-d.1` prints `[PASS]`.
3. **One function, called twice:** `git -C W grep -n -e 'playFromHand(' -- pkg/engine` prints exactly three lines: `pkg/engine/actions.go:89`, `pkg/engine/actions.go:92` and `pkg/engine/actions.go:240` (the declaration).
4. **The copies are gone:** `git -C W grep -n -e 'g.payCost(' -e 'g.takeCard(' -- pkg/engine/actions.go` prints exactly two lines, both inside `playFromHand`: `pkg/engine/actions.go:242` and `pkg/engine/actions.go:243`.
5. **One lookup in `emitAction`:** `git -C W grep -n -e 'handCard(' -- pkg/engine/events.go` prints exactly one line, `pkg/engine/events.go:156`.
6. **One `emitAction` case, and `Zone` only for a Deploy:** `git -C W grep -n -e 'case ActDeploy' -e 'case ActEnact' -e 'a.Kind == ActDeploy' -e 'e.Zone = a.Zone' -- pkg/engine/events.go` prints exactly three lines: `pkg/engine/events.go:155`, `pkg/engine/events.go:159` and `pkg/engine/events.go:160`.
7. **Only the listed edits:** `git -C W diff autoloop/m2-wave-d.1 --numstat -- pkg cmd` prints exactly these two lines: `11	6	pkg/engine/actions.go` and `3	5	pkg/engine/events.go`.
8. **Fences:** `go -C W run ./cmd/cardtown lint fences` exits 0.
9. **Nothing else changed:** `git -C W diff autoloop/m2-wave-d.1 --stat` lists only `pkg/engine/actions.go`, `pkg/engine/events.go` and this task file.
10. `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/engine/actions.go`
- `pkg/engine/events.go`

## Done When
- [ ] Cases 1–10 hold, with their commands and results in the Log
- [ ] `go tool task sim-compare -- autoloop/m2-wave-d.1` passes (starter and stress sims byte-identical to the run branch)
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-05, lead-architect (wave D.1 run, session 2): wrote this task from board directive #108, next in board order, against the run branch at 35cd48c. No `main` commits to sync. Not `WAVE-COMPLETE`: #108–#119 and #121 are `open: before the next Deliverables task`. Harness, so it skips `tests` and goes straight to `ready`; tier `flash`, because the edits are exact and were dry-run. Every command here is on the headless allow-list, one per line, with no pipes, `$`, double quotes or unquoted parentheses. Searched for existing equivalents: `pkg/engine` has no function that finds, pays and takes; `handPlay` checks and changes nothing; `setup.go`'s mulligan takes without paying (see Inputs). Waiting directives checked: none has triggered (#6, #7: mode is `serial`; #8: no `deck.go`; #15: no turn-boundary hook; #71, #91 (2), #106, #118 (b), #120: no such card). Next: `kernel-engineer` on Flash.
