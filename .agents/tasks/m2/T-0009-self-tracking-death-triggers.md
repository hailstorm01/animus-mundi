---
id: T-0009
title: Let death triggers move the card itself, tracked in its owner's discard pile
kind: shape
tier: sonnet
wave: A1 (board directive #4, task 3)
status: tests
branch: m2/T-0009-self-tracking
depends_on: [T-0008]
claims: [type:ability.Subject, subject:self, op:ReturnToHand, op:PutOnDeck, field:Step.Bottom, func:self-tracking-fizzle]
cards: []
requirements: [FR-ZON-004]
---

## Goal
A death trigger can say "return it to its owner's hand" or "put it on the bottom of its owner's deck", where "it" is the card that died. The trigger tracks that exact card instance in its owner's discard pile. When the trigger resolves, it moves the card only if the card is still there, and otherwise does nothing (`FR-ZON-004`, Self-Relocating Trigger Tracking).

## Inputs
- FRD 01:
  - `FR-ZON-004` (lines 156–162 of `docs/frd/01_Game_Rules_FRD.md`): the Self-Relocating Trigger Tracking bullet (line 162) and the leaves-battlefield bullet (line 160).
  - `FR-TRG-006` (line 347): its last sentence says self-relocating death triggers follow the same invariant as delayed triggers. They track the exact card instance, not any card with that name.
  - Destruction puts the card in its owner's discard pile (line 154).
- Card text for context (the A1 card tasks use these; this task tests with synthetic cards): *Unceasing Tide*, "When this unit is killed, return it to its owner's hand."; *Stormfront of Ideals*, "When this unit is killed, put it on the bottom of its owner's deck."
- Built on: T-0007 and T-0008 (`trigger.go`: buffer, flush, `StackAbility`, `resolveAbility`; `HookKilled` fired from `deathSweep`).
- Rulings that apply: none.

### Vocabulary (`pkg/engine/ability/ability.go`)
```go
// Subject is what a step acts on.
type Subject int

const (
	SubjectController Subject = iota // the ability's controller (the zero value; T-0007's Scratch and Draw)
	SubjectSelf                      // the ability's own source card, tracked by card instance (FR-ZON-004)
)

const (
	// after OpScratch, OpDraw:
	OpReturnToHand // put the Subject card into its owner's hand
	OpPutOnDeck    // put the Subject card into its owner's deck: on top, or on the bottom if Bottom
)

// Step gains:
	Subject Subject
	Bottom  bool // OpPutOnDeck: the bottom of the deck instead of the top
```

### Lead decisions (technical, not rules)
- **Tracking:** a buffered trigger, and the `StackAbility` it becomes, remember the source's card instance (`*Instance`; the existing `StackItem.Inst` field may carry it). For a death trigger, `SubjectSelf` means that instance in its **owner's** discard pile. At resolution the step looks for that exact instance in the owner's `Discard`. If it is there, the step moves it. If not, the step does nothing, and the ability's other steps still run.
- **Owner, not controller:** the destination is the owner's hand or deck (`Instance.Owner`), as printed. The trigger's controller doesn't change (T-0008).
- **Deck position:** `Deck[0]` is the top (as `scratch` and `draw` use it), so the bottom is the end of `Deck`.
- **Face-up and printed:** the moved card is the plain card instance. It keeps no battlefield state, because it already reverted when it left (`FR-ZON-004`).
- **Scope:** `SubjectSelf` is defined only for `HookKilled` triggers in this task. On any other hook it does nothing; a later shape task defines it when a card needs it. `OpReturnToHand` and `OpPutOnDeck` with `SubjectController` do nothing. No card or test should build that combination.
- **Not in this task:** a last-known *effective* state snapshot (power, keywords, granted abilities at departure). No A1 card reads one, and nothing can grant abilities yet. Board directive #13 builds it with its first reader.

### What the tests can observe
- **Basic moves:** a unit printing a death trigger with `OpReturnToHand`/`SubjectSelf` dies in combat and ends in its owner's hand, not in discard. The `OpPutOnDeck`/`Bottom` version ends as the last card of its owner's deck. `Bottom: false` puts it on top, so the next draw takes it.
- **Fizzle:** a synthetic unit prints two death triggers, one returning itself to hand and one putting itself on the bottom of the deck. Whichever resolves first moves it, and the second does nothing. The chooser's order decides which. Each trigger's other steps (for example a `Draw` in the same effect) still happen.
- **Identity:** a second copy of the same card (a separate instance with the same `*Card`) already in the discard pile is never moved by the dead card's fizzled trigger.
- **Owner:** both players' units dying together each go back to their own owner's zone.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
- (rules-test-author fills this in.)

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected.

## Allowed Files
- `pkg/engine/ability/ability.go` (the vocabulary above)
- `pkg/engine/trigger.go` (carrying the source instance; the two new steps and the discard check in `resolveAbility`)
- `pkg/engine/stack.go` (only if `StackItem` needs a field or comment change for the source instance)

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run 'TestFR_ZON_004'`
- [ ] T-0006, T-0007 and T-0008 tests still pass unchanged
- [ ] `go test ./...`, `go vet ./...` and `gofmt -l pkg cmd` are clean
- [ ] `go run ./cmd/cardtown lint --all` passes
- [ ] `go run ./cmd/cardtown sim -deck starter -p0 ANIMUS -p1 PATH` exits 0, and plain `sim` output, **sorted**, is unchanged from the integration branch (directive #12: the unsorted line order varies between runs)

## Log
- 2026-10-20 lead-architect: Wrote the task (board directive #4, task 3). Reshaped from "last-known effective state" to FR-ZON-004's self-relocating tracking, which is what the A1 cards *Unceasing Tide* and *Stormfront of Ideals* need. No A1 card reads last-known power or keywords, so that snapshot waits for its first reader (directive #13). `ReturnToHand` and `PutOnDeck` move here from task 4. Set `status: tests` for `rules-test-author`.
