---
id: T-0009
title: Let death triggers move the card itself, tracked in its owner's discard pile
kind: shape
tier: sonnet
wave: A1 (board directive #4, task 3)
status: review
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
- `pkg/engine/self_tracking_test.go` (all `TestFR_ZON_004_*`; the copy tests also cover FR-TRG-006):
  - `DeathTriggerReturnsItselfToOwnersHand`, `DeathTriggerPutsItselfOnBottomOfOwnersDeck`, `DeathTriggerPutsItselfOnTopOfOwnersDeck`
  - `OtherStepsOfTheSameEffectStillRun` (Draw plus return-self in one effect)
  - `ReturnResolvingFirstFizzlesThePutOnBottom`, `PutOnBottomResolvingFirstFizzlesTheReturn` (two triggers on one unit, DecideOrder picks the order; each effect's Draw still runs)
  - `AnotherCopyInDiscardIsNeverMoved` (both orders), `SingleTriggerMovesOnlyTheDeadInstanceNotASameNamedCopy`
  - `EachPlayersCardGoesToItsOwnersZone` (both units die together)
  - Not tested (out of scope per the task): SubjectSelf on non-death hooks; ReturnToHand/PutOnDeck with SubjectController.

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
- 2026-10-20 rules-test-author: Wrote 9 TestFR_ZON_004_* tests in self_tracking_test.go (listed under Tests). No helpers requested. `go vet` fails only on the missing ability vocabulary (OpReturnToHand, Subject, SubjectSelf, Bottom). Left out: the out-of-scope combinations. Fizzle tests assume the existing DecideOrder convention (first pick placed first, resolves last). Set `status: tests-review`.
- 2026-10-20 lead-architect: Test review, round 1: **changes requested**, status `tests`. The tests are correct against FR-ZON-004 and the lead decisions. I checked both fizzle orders, the draws in fizzled triggers, owner destinations, and the identity cases, where the copy is placed first in discard so a search by name would pick the wrong card. One note:
  1. Remove the two `// Covers: FR-TRG-006` comments, on `AnotherCopyInDiscardIsNeverMoved` and `SingleTriggerMovesOnlyTheDeadInstanceNotASameNamedCopy`. Keep the tests and their FR-TRG-006 wording in the failure messages. FR-TRG-006 is the delayed-trigger requirement: creation, firing once, binding to the instance, and end-of-turn expiry. These tests pin only its last sentence, so `lint rules` would count FR-TRG-006 as executable while delayed triggers don't exist, and overstate M2 progress. Check afterwards that `lint rules` lists FR-TRG-006 under "Not yet executable" again.
- 2026-10-20 rules-test-author: Round 1 fix: removed the two '// Covers: FR-TRG-006' comments (tests and failure messages unchanged). Set status: tests-review.
- 2026-10-20 lead-architect: Test review, round 2: **approved**, status `ready`. Note 1 is done: `84b9ae5` removes exactly the two `// Covers: FR-TRG-006` lines, and `lint rules` lists FR-TRG-006 under "Not yet executable" again. Next: `kernel-engineer` (tier sonnet).
- 2026-10-20 kernel-engineer: Implemented. Diff: `ability.go` gains Subject (SubjectController = 0, SubjectSelf), OpReturnToHand and OpPutOnDeck after OpDraw, Step.Subject and Step.Bottom. `trigger.go`: pendingTrigger carries `inst` (the asset's `*Instance`), copied into the existing `StackItem.Inst` when flushed (stack.go unchanged); new `relocateSelf` called from `resolveAbility` finds that exact pointer in the owner's Discard and moves it to the owner's Hand, or Deck (top = Deck[0], Bottom appends); otherwise does nothing and the remaining steps run. Does nothing unless Subject is SubjectSelf and the hook is HookKilled. Reuse: searched trigger.go/stack.go, reused StackItem.Inst, no other helper existed. Commands: `go test ./pkg/engine -run TestFR_ZON_004` 11 PASS lines (9 tests plus subtests); `go test ./...` ok; `go vet ./...` clean; `gofmt -l pkg cmd` empty; `go run ./cmd/cardtown lint --all` passed; `sim -deck starter -p0 ANIMUS -p1 PATH` exit 0, sorted output identical to a worktree of rework/executable-rules. Reviewer: check the discard removal slice expression (`[:i:i]` forces a copy).
- 2026-10-20 lead-architect: Code review, round 1: **changes requested**, status `changes`. Otherwise the code is right: identity tracking by `*Instance`, the owner's zones, top/bottom placement, the remaining steps still run, and `stack.go` is untouched. Gates re-run by the reviewer: `go test ./...`, `go vet`, `gofmt` and `lint --all` pass. `lint rules` shows 37 of 62. Sorted plain `sim` output is identical to the integration branch, and the starter sim exits 0. One note:
  1. **Reuse (Runbook D5.3):** the discard search and splice in `relocateSelf` re-implements `remove[T comparable]` in `pkg/engine/state.go`, down to the `[:i:i]` copy. Use the existing helper instead. For example, return early unless `slices.Contains(owner.Discard, it.Inst)`, then `owner.Discard = remove(owner.Discard, it.Inst)` and place `it.Inst`. Behavior must stay the same, so the same tests pass. Don't edit `state.go`.
- 2026-10-20 kernel-engineer: Round 1 fix. Note 1: `relocateSelf` now returns early unless `slices.Contains(owner.Discard, it.Inst)`, then uses the existing `remove` from state.go (added the `slices` import to trigger.go); placement into hand/deck is unchanged. Only trigger.go changed. Gates: `go test ./...`, `go vet ./...`, `gofmt -l pkg cmd` (empty), `lint --all` pass; TestFR_ZON_004 passes; sorted plain `sim` identical to rework/executable-rules; starter sim exit 0.
