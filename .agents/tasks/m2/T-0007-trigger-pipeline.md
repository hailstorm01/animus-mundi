---
id: T-0007
title: Run self-deploy triggers through a pending buffer, with player-ordered placement
kind: shape
tier: sonnet
wave: A1 (board directive #4, task 2a)
status: tests-review
branch: m2/T-0007-trigger-pipeline
depends_on: [T-0006]
claims: [type:ability.Trigger, type:ability.Hook, hook:deploy, type:ability.Effect, type:ability.Step, type:ability.Op, type:ability.Value, op:Scratch, op:Draw, kind:StackAbility, func:pending-trigger-buffer, decision:DecideOrder]
cards: []
requirements: [FR-TRG-003, FR-TRG-004]
---

## Goal
A card can print "When this deploys, …" as data. When the deploy resolves, the card's deploy triggers go into a pending buffer. After the deploy item and its death sweep finish, the buffer is flushed onto the Resolution Stack in APNAP order, and each player orders their own triggers through the `Chooser`. A trigger resolves by running its steps. This task adds two steps, `Scratch` and `Draw`, which the kernel can already perform.

## Inputs
- FRD 01: `FR-TRG-003` (lines 333–337 of `docs/frd/01_Game_Rules_FRD.md`), `FR-TRG-004` (line 340), `FR-QUE-001` (lines 323–328, for context), `FR-ZON-005` (Deploy Scratch Company) and `FR-WIN-001` (drawing from an empty deck), which the two steps reuse.
- Design note: `docs/architecture/m2-ability-model.md`, Boundaries (the `ability` key types and "Events and triggers") and Open Question 6.
- Board directive #10: the ordering fallback is collection order (provisional, user 2026-10-19).
- Choices: `Decision`, `Chooser` and `g.choose` from T-0006 (`pkg/engine/choice.go`).
- Rulings that apply: none beyond directive #10.

### Vocabulary the tests build synthetic cards from (`pkg/engine/ability/ability.go`)
```go
// Def gains:
	Triggers []Trigger // printed triggered abilities, in printed order

// Hook is the event a trigger watches.
type Hook int

const (
	HookDeploy Hook = iota + 1 // this asset was deployed (its deploy resolved onto the battlefield)
)

// Trigger is one printed triggered ability.
type Trigger struct {
	Hook   Hook
	Effect Effect
}

// Effect is what a resolving ability does: its steps, in order.
type Effect struct {
	Steps []Step
}

// Op is one primitive step.
type Op int

const (
	OpScratch Op = iota + 1 // its controller performs Deploy Scratch Company Amount times (FR-ZON-005)
	OpDraw                  // its controller draws Amount cards (FR-WIN-001 applies to an empty deck)
)

// Step applies one primitive.
type Step struct {
	Op     Op
	Amount Value
}

// Value is a number a step uses. Later shape tasks add other sources; the
// zero source always means the fixed N.
type Value struct {
	N int
}
```
A synthetic test card is an ordinary `*Card` with `Ability: &ability.Def{Triggers: …}`. No `Text` is needed, because only `LoadCards` checks it.

### Lead decisions (technical, not rules)
- **Self hook only:** `HookDeploy` fires for the deployed asset's own triggers, read from `a.Printed().Ability`. A face-down Scratch Company has no abilities, so it never triggers. HQ placement during match setup is not a deploy.
- **Controller:** a trigger's controller is the source asset's controller when the event happens. `Scratch` and `Draw` act on that controller. When a later task adds a `Subject` field, its zero value must keep meaning "the controller".
- **Buffer and flush (`FR-TRG-004`):** triggers that fire while an item resolves wait in a pending buffer on `Game`. After that item's death sweep (`resolveStack`), the buffer is flushed: the active player's triggers are placed first, then the non-active player's on top (`FR-TRG-003`). If the match ended at the sweep, the buffer is discarded.
- **Ordering decision:** a player with two or more triggers in one flush is asked one `Decision` of a new kind `DecideOrder`. `Options` holds one option per trigger in collection order, with `Min == Max == len(Options)`. `Label` names the source and the trigger, and `Ref` is the source asset ID. The answer lists triggers in **placement order**: the first pick goes on the stack first, so it resolves last of that player's triggers. A player with exactly one trigger is not asked.
- **`DecideOrder` in `choice.go`:** order matters, so the "take every option" fast path does not apply. A `DecideOrder` decision is asked whenever it has two or more options. Its fallback (nil chooser, timeout or invalid answer) is collection order (`0 … n-1`, directive #10). `valid` and `RandomChooser` already handle a permutation (`Min == Max == n`, no repeats). Check that `RandomChooser` gives a random permutation, not always the identity.
- **Collection order:** battlefield order (`g.Assets()`), then the card's printed trigger order. For one deploy event that is just printed order.
- **Stack item:** a new `StackAbility` kind carries the source asset ID, the controller and the `ability.Trigger` it resolves. It resolves whether or not its source is still on the battlefield, because untargeted steps have nothing to verify. Its own death sweep follows as for any item. A `Draw` from an empty deck therefore loses at that sweep, before the next trigger resolves.
- **Not in this task:**
  - death triggers and cross-player APNAP (T-0008, next);
  - last-known effective state (`FR-ZON-004`);
  - attack hooks, `turnOwner`, `ReturnToHand` and `PutOnDeck` (a later A1 shape task);
  - the deep-cascade case of `FR-TRG-004`, where a trigger's resolution causes a new trigger that must resolve before a deeper item. No A1 step causes an event, so that test arrives with the first step that kills or deploys (board directive #11).

### What the tests can observe
- **Order:** stack distinct named cards on top of a deck. A `Draw` trigger and a `Scratch` trigger on one card then show their order by where the top card ends up: in hand, or face-down in Reserves.
- **Questions:** `s.asked(p)` shows which ordering decisions were asked, and of whom.
- **Empty deck:** an empty deck under a `Draw` shows that a trigger's sweep happens before the next trigger.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
- `pkg/engine/trigger_test.go`:
  - `TestFR_TRG_004_DeployTriggerResolvesAfterCardIsOnBattlefield`, `_SingleTriggerResolvesWithoutQuestion`, `_StepsOfOneEffectRunInPrintedOrder`, `_ScratchAmountPerformsScratchCompanyThatManyTimes`, `_TriggerControllerIsTheDeployingPlayer`
  - `TestFR_TRG_003_TwoTriggersAskOneOrderDecision`, `_FirstPickIsPlacedFirstAndResolvesLast`, `_ReversedAnswerReversesResolution`, `_ThreeTriggersFollowThePlayersPermutation`, `_FallbackIsCollectionOrder` (timeout, no chooser, repeat, out of range, too few), `_InvalidOrderAnswerIsLoggedAsInvalid`, `_FaceDownScratchCompanyNeverTriggers`
  - `TestFR_TRG_004_EmptyDeckDrawLosesAtItsOwnSweepBeforeNextTrigger`, `_MatchEndedAtDeploySweepDiscardsBufferedTriggers`
  - DecideOrder vs T-0006: `TestFR_TRG_003_OrderDecisionWithManyOptionsIsNeverFastPathed`, `_OrderDecisionFallbackIsCollectionOrder`, `TestRandomChooser_OrderDecisionIsAValidPermutation`, `_OrderDecisionIsNotAlwaysTheIdentity`
  - Left out: cross-player APNAP (T-0008); deep cascade of FR-TRG-004 (directive #11).
  - Weak spot: in the empty-deck test, "before the next trigger" is observable only as "the later Scratch never resolved", because an empty deck leaves a Scratch nothing to take. A stronger check needs a step that acts on the other player.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- `func (s *scene) onTop(p PlayerID, cards ...*Card) []*Instance`: puts new instances of `cards` on top of p's deck, `cards[0]` on top, and returns them. (rules-test-author may request others.)

## Allowed Files
- `pkg/engine/ability/ability.go` (the vocabulary above)
- `pkg/engine/trigger.go` (new: the pending buffer, collection, APNAP flush, ordering decision, and resolving a `StackAbility`'s steps)
- `pkg/engine/stack.go` (`StackAbility`, the deploy event in `resolveDeploy`, the flush in `resolveStack`)
- `pkg/engine/choice.go` (`DecideOrder`: no take-everything fast path, collection-order fallback)

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run 'TestFR_TRG_003|TestFR_TRG_004'`
- [ ] T-0006's tests still pass unchanged: `go test ./pkg/engine -run 'TestFR_MOD_001|TestFR_TRG_002|TestRandomChooser'`
- [ ] `go test ./...`, `go vet ./...` and `gofmt -l pkg cmd` are clean
- [ ] `cardtown lint rules` shows `FR-TRG-003` and `FR-TRG-004` pinned
- [ ] `go run ./cmd/cardtown lint --all` passes (fences: no card names in `pkg/engine` or `pkg/engine/ability`)
- [ ] `go run ./cmd/cardtown sim -deck starter -p0 ANIMUS -p1 PATH` exits 0, and plain `sim` output is unchanged from the integration branch (no real card has a `Def` yet)

## Log
- 2026-10-19 lead-architect: Wrote the task (board directive #4, task 2, split under the D2 size rule into 2a, this one, and 2b, T-0008: death triggers, cross-player APNAP and the playout invariants). Applies directive #10 (collection-order fallback). Set `status: tests` for `rules-test-author`. Note for tests: FR-TRG-003's cross-player half can't be built yet, because one deploy event only triggers its own card. Pin same-player ordering here; T-0008 pins cross-player APNAP.
- 2026-10-19 rules-test-author: Wrote pkg/engine/trigger_test.go (see Tests). `go vet` fails only on the missing ability vocabulary, DecideOrder and s.onTop (already requested). `cardtown lint rules` passes with FR-TRG-003/004 traced. No new helpers requested; local helpers (trigUnit, deployFromHand, inHand, scratched) live in trigger_test.go. Left out: cross-player APNAP, deep cascade. Status set to tests-review.
