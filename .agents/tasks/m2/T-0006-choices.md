---
id: T-0006
title: Ask players through a per-seat Chooser, with the FR-MOD-001 fast path and fallbacks
kind: shape
tier: sonnet
wave: A1 (board directive #4, task 1 of 4)
status: in-progress
branch: m2/T-0006-choices
depends_on: [T-0005]
claims: [type:Decision, type:DecisionKind, type:Option, type:Answer, type:Chooser, config:Choosers, func:Game.choose, func:RandomChooser]
cards: []
requirements: [FR-MOD-001, FR-TRG-002]
---

## Goal
The kernel can ask a player a question. `g.choose(d)` takes the single-option fast path itself, otherwise asks the deciding seat's `Chooser`, and applies the `FR-MOD-001` §6 fallback when there is no chooser, the answer is a timeout, or the answer is invalid. Nothing in play asks a question yet. The next A1 task (events and APNAP placement) is the first caller.

## Inputs
- FRD 01: `FR-MOD-001` (lines 305–315 of `docs/frd/01_Game_Rules_FRD.md`), `FR-TRG-002` (lines 330–331). Ruling `SVY-R6` (distinct modes) is folded into `FR-MOD-001` §5.
- Design note: `docs/architecture/m2-ability-model.md`, Boundaries ("Choices"), Test Plan ("Choices") and Open Questions 3 (no clocks in M2: a timeout is an answer).
- Decision record: `docs/architecture/decisions/0003-synchronous-chooser-for-player-choices.md`. ADR-0002: the random fallback draws from the game PRNG.
- Rulings that apply: `SVY-R6`.

### API the tests code against (the engineer creates it in `pkg/engine/choice.go`)
```go
// DecisionKind says what a Decision asks, and so which FR-MOD-001 §6 fallback applies.
type DecisionKind int

const (
	DecideTarget   DecisionKind = iota // mandatory targets: fallback picks random legal options
	DecideMode                         // "Choose N" modes: fallback picks the first listed options
	DecideOptional                     // "you may": fallback declines (no picks)
)

// Option is one legal answer element, in the order the card or rules list it.
type Option struct {
	Label string // for logs and cardtown play
	Ref   int    // the asset or card instance ID it refers to, or 0
}

// Decision is one question to one player. Options holds only legal options.
type Decision struct {
	Player   PlayerID
	Kind     DecisionKind
	Prompt   string
	Options  []Option
	Min, Max int  // how many picks a valid answer has
	Hidden   bool // the choice inspects hidden cards in a Hand or Deck (no fast path)
	Repeats  bool // the same option may be picked more than once (FR-MOD-001 §5)
}

// Answer is a Chooser's reply: indexes into Decision.Options, or a timeout.
type Answer struct {
	Picks   []int
	Timeout bool
}

// Chooser answers decisions for one seat.
type Chooser func(g *Game, d Decision) Answer

// RandomChooser returns a Chooser that gives a uniformly random valid answer to
// every Decision, deterministic for a seed. Playouts and cardtown sim use it.
func RandomChooser(seed uint64) Chooser

// choose returns the picks for d (indexes into d.Options).
func (g *Game) choose(d Decision) []int
```
`Config` gains `Choosers [2]Chooser`, indexed by `PlayerID`. A nil entry means that seat always times out.

### Lead decisions (technical, not rules)
- **Valid answer:** not a timeout, `Min ≤ len(Picks) ≤ Max`, every pick is an index into `Options`, and no index repeats unless `Repeats`.
- **Fast path (§1):** when `Hidden` is false and exactly one valid answer exists, `choose` returns it without calling the `Chooser`. Exactly one exists when `Min == Max == len(Options)` (take everything), or when `Repeats` and there is one option and `Min == Max`. A `DecideOptional` decision always has two valid answers (take it or decline), so it is always asked. When `Hidden` is true the `Chooser` is asked even if only one answer exists.
- **No valid answer** (for example `Min > len(Options)` without `Repeats`): `choose` returns no picks and asks no one. Callers handle it (a fizzle, or `LegalActions` not offering the declaration).
- **Who is asked:** `Config.Choosers[d.Player]`, whichever player is active (`FR-TRG-002`).
- **Fallbacks (§6):** apply on a nil chooser, `Timeout`, or an invalid answer. `DecideTarget` picks `Min` distinct options uniformly at random from the game PRNG. `DecideMode` picks options `0 … Min-1`. `DecideOptional` picks nothing.
- **An invalid answer** also appends a `Game.Log` event of kind `invalid-answer` whose detail names the player and the prompt. Timeouts and nil choosers are not logged as invalid. (The design note's playout invariant "a `Chooser` answer is never rejected" is wired by the next A1 task, when play first asks a question.)
- **Not in this task:** trigger ordering (`FR-TRG-003`, next task), top-of-deck inspections (§4, wave F fixes the encoding), and the random-target fallback when `Min < Max` (only `Min == Max` is pinned here; wave A2 raises it if a card needs it).

### Testing the shape
No card asks a question yet, so rule tests call `s.g.choose(d)` directly on a scene, with a scripted chooser installed through the helper below. This seam is allowed for this task only. Later tasks test choices through play.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
All in `pkg/engine/choice_test.go`, calling `s.g.choose(d)` directly.
- FR-MOD-001 §1 fast path: `TestFR_MOD_001_FastPathSoleTargetNotAsked`, `_FastPathTakeEverythingNotAsked`, `_FastPathRepeatsSingleOptionNotAsked`, `_MultipleOptionsAskedAndAnswerUsed`, `_FastPathNotTakenWhenHidden`, `_OptionalAlwaysAsked`, `_OptionalDeclineAnswerHonoured`, `_NoLegalAnswerReturnsNothingAndAsksNoOne`
- FR-MOD-001 §5 distinct modes: `_DistinctModesRepeatedPickIsInvalid`, `_RepeatedModeAllowedWhenCardSaysSo`
- FR-MOD-001 §6 fallbacks: `_ModeFallbackOnTimeoutPicksFirstListed`, `_ModeFallbackWithNoChooserPicksFirstListed`, `_OptionalFallbackOnTimeoutDeclines`, `_OptionalFallbackWithNoChooserDeclines`, `_TargetFallbackOnTimeoutPicksLegalDistinctTargets`, `_TargetFallbackWithNoChooserPicksLegalTarget`, `_TargetFallbackDeterministicForSameSeed`, `_TargetFallbackIsRandomNotFixed`, `_InvalidAnswerFallsBackAndIsLogged` (5 subcases), `_InvalidOptionalAnswerDeclines`
- FR-TRG-002: `TestFR_TRG_002_InactivePlayerIsAskedOwnChooser`, `_InactivePlayerFastPathsSoleOption`, `_InactivePlayerHiddenInspectionStillAsked`
- `TestRandomChooser_AnswersAreAlwaysValid`, `_SameSeedSameAnswers`, `_NotConstant`, `_DifferentSeedsDiffer`

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- `func (s *scene) script(p PlayerID, answers ...Answer)`: installs a chooser for seat p that returns `answers` in order and records every `Decision` it is asked. Asked past the end of the script, it calls `s.t.Errorf` naming the decision's prompt and returns a timeout.
- `func (s *scene) asked(p PlayerID) []Decision`: the decisions seat p's scripted chooser was asked, in order.

## Allowed Files
- `pkg/engine/choice.go` (new: the API above)
- `pkg/engine/turn.go` (`Config.Choosers`, copied onto the game in `NewGame`)
- `pkg/engine/state.go` (the unexported field on `Game` that holds the choosers)
- `pkg/engine/sim.go` (`Simulate` sets both seats to `RandomChooser`, seeded from the game seed as the policies are)

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run 'TestFR_MOD_001|TestFR_TRG_002|TestRandomChooser'`
- [ ] `go test ./...`, `go vet ./...` and `gofmt -l pkg cmd` are clean
- [ ] `cardtown lint rules` shows `FR-MOD-001` pinned
- [ ] `go run ./cmd/cardtown lint --all` passes
- [ ] `go run ./cmd/cardtown sim -deck starter -p0 ANIMUS -p1 PATH` exits 0, and plain `sim` output is unchanged from the integration branch (nothing asks yet)

## Log
- 2026-10-19 lead-architect: Wrote the task (board directive #4, task 1 of 4: choices first). Set `status: tests` for `rules-test-author`. Besides the rule tests, I'd like `TestRandomChooser_*` tests for: every answer valid across many decisions of each kind, including `Repeats`; the same seed gives the same answers; and it covers more than one answer (it isn't constant).
- 2026-10-19 rules-test-author: Wrote the tests above on m2/T-0006-choices and set `status: tests-review`. Helpers requested: `s.script`, `s.asked` (as already listed). Tests do not compile until the engineer adds the API. Left out: trigger ordering, top-of-deck inspections, and the random-target fallback when Min < Max (out of scope per the task). FRD 01 does not say which index the random fallback picks, so tests assert validity, distinctness, determinism per game seed and non-constancy only. The invalid-answer log test checks that the detail contains the prompt and the player's digit, since the format is not specified. Engineer note: the RandomChooser mix covers Min=0 target/mode decisions and Repeats with Max above the option count.
- 2026-10-19 lead-architect: Test review: **approved**, status `ready`. The tests match FR-MOD-001 §1, §5 and §6 (target, mode and optional), FR-TRG-002 and the task's lead decisions. The random fallback is pinned by validity, determinism and spread, not by index, which is right because FRD 01 only says "a random legal target". Non-blocking: the `invalid-answer` player check matches only the digit, and the invalid-answer fallback isn't exercised on a target decision. The engineer implements the fallback per kind anyway. Next: `kernel-engineer` (tier sonnet) adds the API and the two helpers.
