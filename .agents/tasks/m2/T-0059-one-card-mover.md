---
id: T-0059
title: One place moves a card between deck, hand and discard, takeCard and putCard (directive #52, harness, sonnet, pure move)
kind: harness
tier: sonnet
wave: none (board directive #52, structural review 3 finding 4, before T-0055)
status: review
branch: m2/T-0059-one-card-mover
depends_on: [T-0058]
claims: [helper:engine.takeCard, helper:engine.putCard]
cards: []
requirements: []
---

## Goal
Two helpers in `pkg/engine/zones.go` are the only code that changes a player's `Deck`, `Hand` or `Discard`: `takeCard` lifts an instance out of whichever pile holds it, and `putCard` puts an instance into a named pile. The six hand-rolled pile edits call them. Waves B and C add discarding from hand, discard to the bottom of the deck, return to hand, and the "enters discard" and "discarded from hand" hooks; those hooks will fire from these helpers. Nothing else changes: same tests, same sim output.

## Inputs
- Board directive #52 (structural review 3, finding 4).
- Pattern to follow: T-0058 (`exhaust`, a pure move checked by byte-identical sims).
- Rulings that apply: none. `FR-SET-002` (the mulligan sets cards aside, draws, then shuffles them in) and `FR-ZON-004` (`relocateSelf` tracks the exact instance) must still hold; the existing tests pin both.

### The helpers (lead decision)
Two helpers rather than one, because two of the six moves have no destination pile (hand to stack, deck to battlefield) and two have no source pile (battlefield to discard, set-aside to deck).

```go
// takeCard removes inst from whichever of its owner's piles (deck, hand or
// discard) holds it, and reports whether one did. A card on the battlefield,
// on the stack or set aside is in no pile.
func (g *Game) takeCard(inst *Instance) bool

// putCard puts inst into its owner's pile to, which is ZoneDeck, ZoneHand or
// ZoneDiscard. On the deck, bottom chooses the bottom over the top (index 0).
// Hand and discard always add at the end.
func (g *Game) putCard(inst *Instance, to Zone, bottom bool)
```
- The pile is always the **owner's** (`inst.Owner`). Every site below already uses the owner's piles.
- `takeCard` uses the existing `remove` helper on the pile it finds. Instances are unique, so the search order across the three piles doesn't matter.
- `putCard` with any other `to` panics with a message naming the zone. It's an engine bug, not a game state.
- Place both directly after `scratch` in `zones.go`, with the doc comments above. Update the file's header line only if it no longer describes the file.

### The six sites (at `main` c4240cd)
| Move | Where | Becomes |
| :--- | :--- | :--- |
| Deck to hand | `zones.go:9-14`, `moveTopToHand` | `takeCard(p.Deck[0])`, then `putCard(…, ZoneHand, false)`, inside the existing `len(p.Deck) > 0` guard |
| Deck to battlefield (scratch) | `zones.go:37-38`, `scratch` | `inst := p.Deck[0]`, then `takeCard(inst)` |
| Battlefield to discard | `damage.go:95-96`, the death sweep | `putCard(a.Inst, ZoneDiscard, false)` (the `owner` variable goes if unused) |
| Hand to stack | `actions.go:82`, `ActDeploy` | `takeCard(inst)` |
| Discard to hand or deck | `effects.go:158-170`, `relocateSelf` | Keep the `slices.Contains(owner.Discard, …)` guard: only a card **in the discard pile** relocates (`FR-ZON-004`). Then `takeCard`, then `putCard` to `ZoneHand`, or to `ZoneDeck` with `bottom: st.Bottom`. |
| The mulligan | `setup.go:76-87` | `takeCard(inst)` in place of `p.Hand = remove(p.Hand, inst)`. The set-aside cards stay out of every pile while the replacements are drawn (`FR-SET-002`). After the draws, `putCard(inst, ZoneDeck, true)` for each, in `aside` order, in place of `p.Deck = append(p.Deck, aside...)`. |

Each new call replaces the old edit at the same position, so nothing around it moves. If a site can't use the helpers without changing a test-visible order, leave that site as it is, and name it and why in the Log. The lead expects none.

### Left alone (lead decision)
- Setup's initial deck build, `setup.go:47` (`p.Deck = append(p.Deck, &Instance{…})`). The instance is being created, not moved.
- `shuffle`, `zones.go:6`. It reorders the deck in place and moves no card between piles.
- The snapshot's deep copy, `snapshot.go:118-120`.
- `pkg/view` (it builds view structs, not game state).

## Acceptance Cases (harness tasks only; written by the lead)
Before any edit, build a baseline CLI from `main` and save its sim outputs. Commit the new tests first as `test(T-0059): …`; they fail before the implementation (a compile failure counts). Record each check's commands and results in the Log; the reviewer re-runs them.

New unit tests for the helpers, in a new file `pkg/engine/zones_test.go`, on a `Game` built the way other in-package tests build one (one test per case):
1. `putCard(x, ZoneDeck, false)` → `x` is the owner's `Deck[0]`; the other deck cards keep their order behind it.
2. `putCard(x, ZoneDeck, true)` → `x` is the owner's last deck card; the others keep their order.
3. `putCard(x, ZoneHand, …)` and `putCard(x, ZoneDiscard, …)` → `x` is the last card of that pile, for both values of `bottom`.
4. `putCard` puts the card into the **owner's** pile, even when the other player is active.
5. `takeCard` on a card in the deck, the hand, or the discard pile → returns `true`; the card is in none of the owner's piles; the rest of that pile keeps its order.
6. `takeCard` on a card in no pile (for example a battlefield asset's instance) → returns `false`; every pile of both players is unchanged.
7. `putCard(x, ZoneStack, false)` panics (`ZoneReserves` too).

Repository checks:

8. **Existing tests unchanged:** `git diff main -- '*_test.go'` shows only the new `zones_test.go`, and `go test ./...` passes.
9. **Starter sim byte-identical:** `cardtown sim -deck starter -p0 ANIMUS -p1 PATH` from the baseline and the new binary: stdout+stderr and exit code match (`cmp`).
10. **Stress sim byte-identical:** the same for `cardtown sim` with no `-deck`.
11. **Only the helpers write piles:** `git grep -n -E '\.(Hand|Discard|Deck)(\[[^]]*\])? *(=|\+=)[^=]' -- 'pkg/engine/*.go' ':!pkg/engine/*_test.go'` prints only lines inside `takeCard` and `putCard`, plus `setup.go`'s initial deck build, `shuffle`'s swap and `snapshot.go`'s three clones. Record the output.
12. **Only these edits:** `git diff main --stat` lists only the Allowed Files, `pkg/engine/zones_test.go` and this task file.
13. `go tool task check` and `go tool task sim` pass.

## Allowed Files
Five non-test files, one over the size rule. As with T-0035 and T-0048, a pure move meets the rule through its byte-identical checks (cases 9–12). Each file loses an edit, and only `zones.go` gains code.
- `pkg/engine/zones.go` (the helpers; `moveTopToHand` and `scratch`)
- `pkg/engine/actions.go` (`ActDeploy`)
- `pkg/engine/damage.go` (the death sweep)
- `pkg/engine/effects.go` (`relocateSelf`)
- `pkg/engine/setup.go` (the mulligan)
- new test file: `pkg/engine/zones_test.go`

## Done When
- [ ] Cases 1–7 are tests in `zones_test.go`, committed before the implementation, and pass
- [ ] Cases 8–12 hold, with their commands and results in the Log
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks

## Log
- 2026-10-24 lead-architect: Wrote the task from board directive #52, the second of the four directives before T-0055 (#49 done in T-0058; #53 and #54 follow). Harness, tier sonnet, as the directive says. It is a pure move, so it skips `tests` and is `ready` on branch `m2/T-0059-one-card-mover`. The directive let the engineer pick the signature; the lead fixed it as two helpers instead, so the review has one shape to check. Checked the six sites against `main` c4240cd; the directive's line references still hold. Waiting triggers checked: #6 and #7 (mode is `serial`), #8 (no `deck.go` change), #15 (no turn hook), #20 (no wave C shape task), #50 (no task claims the exhaust hook), #51 (no "if you do" task) and #55 (#53's task not written). None has fired. Starter-deck cards not yet executable: 28, plus both Headquarters.
  - Next: the lead dispatches `kernel-engineer` (Sonnet).
- 2026-10-24 kernel-engineer (sonnet): Implemented. Status `review`.
  - Reuse search: `remove` (zones.go) is reused by `takeCard`; no existing pile-move helper existed.
  - Baseline: built `cardtown` from `main` (8373028) before any edit; saved starter sim and stress sim stdout+stderr and exit codes.
  - Commits: `41607be` test(T-0059) first (compile failure: `putCard` undefined), then the implementation commit.
  - Changed: `zones.go` gains `takeCard` and `putCard` after `scratch` (imports `fmt`, `slices`); `moveTopToHand`, `scratch`, the death sweep, `ActDeploy`, `relocateSelf` (guard `slices.Contains(owner.Discard, …)` kept) and the mulligan call them. Mulligan: `takeCard` per set-aside card, draws, then `putCard(inst, ZoneDeck, true)` per card in `aside` order. All six sites converted; none left as is. Header line of zones.go unchanged (still accurate).
  - Case 1-7: `TestPutCard_DeckTop`, `_DeckBottom`, `_HandAndDiscardAddAtEnd` (both `bottom` values), `_OwnersPileWhenOtherPlayerActive`, `TestTakeCard_FromEachPile`, `_InNoPile`, `TestPutCard_PanicsOnOtherZones` (Stack and Reserves). All pass.
  - Case 8: `git diff main -- '*_test.go' --stat` lists only `pkg/engine/zones_test.go`; `go test ./...` passes (via `go tool task check`).
  - Case 9: `cardtown sim -deck starter -p0 ANIMUS -p1 PATH` baseline vs new: `cmp` identical on output and exit code (0).
  - Case 10: `cardtown sim` (no `-deck`): `cmp` identical, exit code 0.
  - Case 11: `git grep -n -E '\.(Hand|Discard|Deck)(\[[^]]*\])? *(=|\+=)[^=]' -- 'pkg/engine/*.go' ':!pkg/engine/*_test.go'` prints: setup.go:47 (initial deck build); snapshot.go:118-120 (three clones); zones.go:11 (shuffle swap); zones.go:76,78,81,83 (inside `putCard`). `takeCard` writes via `*pile = remove(…)`, which the regex does not match.
  - Case 12: `git diff main --stat` lists actions.go, damage.go, effects.go, setup.go, zones.go, zones_test.go and this task file only.
  - Case 13: `go tool task check` exit 0; `go tool task sim` exit 0.
  - Review: `takeCard` searches deck, hand, discard in that order with `slices.Contains` then `remove`; the death sweep's unused `owner` variable is gone.
