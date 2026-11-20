---
id: T-0291
title: "FR-DEK-002 and FR-REP-002 get their own scenario tests (directive #177; wave H.3, sonnet, tests only)"
kind: shape
tier: sonnet
wave: H.3 (board directive #177)
status: tests-review
branch: m2/autoloop-wave-h.3/T-0291-dek-002-rep-002-tests
depends_on: [T-0290]
claims: [test:TestFR_DEK_002_*, test:TestFR_REP_002_*]
cards: []
requirements: [FR-DEK-002, FR-REP-002]
---

## Goal
`FR-DEK-002` (deck shuffling) and `FR-REP-002` (inline substitution and trigger suppression) each get `TestFR_*` scenario tests under their own ID, so `cardtown lint rules` reports the M2 target of 58 of 62 active FRD 01 requirements. Both rules are already built and exercised under other IDs' tests, so no engine change is expected. `kind: shape` only because the tests go through `rules-test-author` and the lead's tests review (D3, D4); the task adds no machinery.

## Inputs
- FRD 01: `FR-DEK-002` (lines 198-199: a shuffle randomizes every card currently in that deck, as an immediate game state action that uses no Resolution Stack item, `FR-QUE-001`); `FR-DEK-001`'s "Sequence & Mandatory Shuffle" (line 195) for when a search shuffles. `FR-REP-002` is cited by §6.3's "Inline Event Substitution" (lines 529-530): replacement effects (*"If X would happen, do Y instead"*, damage redirection, **Resistant N**, **Stalwart**, **Deploys Exhausted**) intercept an event immediately before it occurs and replace or modify it inline, without the Resolution Stack; a completely replaced event never occurs, and triggers watching for it don't fire. The "Replaced Actions Still Count as Done" bullet (line 534) and §6.3's other bullets (lines 531-537) say what else holds; cite them with `// Covers:` where a test pins them too.
- Rulings that apply: `SVY-R15` (a replaced action still counts as done, but the replaced event's watchers don't fire), `SVY-R19` (a replaced death is not a combat kill), `SVY-R37`, `SVY-R38` (search order and shuffle).
- Design note: `docs/architecture/m2-ability-model.md`, Deliverable 6p.
- Board directive #177 (in `BOARD.md` on the run branch): the source of this task, with its acceptance list.
- Card text (`data/cards/*.csv`): *Shifting Impossibility* (the destruction replacement); *Cliff of Portents* (a deck search, if a real card is wanted).

### What exists (tests to read, not to rename or edit)
- `pkg/engine/search_deck_test.go`: `TestFR_DEK_001_SearchShufflesTheRest` (line 178) checks that a search shuffles the rest, with `s.namedDeck`, `s.searcher`, `searchStep`, `instIDs` and `EventShuffle`; `TestFR_DEK_001_EmptyDeckStillShuffles` (line 124).
- `pkg/engine/replacement_test.go`: `TestFR_REP_001_ReplacedDeathFiresNoDeathTrigger` (line 111), built with `guardianStructure`, `mortal`, `killedTrigger` and `s.sweep`.
- `pkg/engine/redirect_test.go`: the `TestFR_REP_004_Redirect*` tests (lines 48-262).
- `pkg/engine/cards_animus_impossibility_test.go`: *Shifting Impossibility*'s card tests.

### Tests to write (directive #177's acceptance; at least these, each through `newScene` or a real card)
- `TestFR_DEK_002_ShuffleKeepsEveryCardInTheDeck`: after a search shuffles, the deck holds the same cards as before, minus the chosen one.
- `TestFR_DEK_002_ShuffleUsesNoStackItem`: the shuffle happens while the search item resolves, and no stack item is placed for it.
- `TestFR_REP_002_RedirectionUsesNoStackItem`: redirected damage moves while its item resolves, and the stack gains nothing.
- `TestFR_REP_002_ReplacedDeathFiresNoEntersDiscardTrigger`: a death that *Shifting Impossibility* (or the scene's equivalent replacement) replaces puts the unit into its owner's hand, and neither its `HookKilled` nor its `HookEntersDiscard` trigger fires.

Add others for what `FR-DEK-002` and `FR-REP-002` say that no test pins under their own ID (for example, a shuffle touches only the searched player's deck; a Resistant or Stalwart modification uses no stack item), as long as FRD 01 settles the expected outcome. Don't assert a particular random order: assert that the cards are the same set, or compare against a fixed-seed run only where an existing test already does.

### Lead decisions (technical)
1. Existing tests are not renamed or edited. New tests go in new files or at the end of the files above, one behavior each.
2. If every new test passes on the `test(T-0291)` commit, `kernel-engineer` has nothing to do: the lead reviews the tests alone, as for #46 (T-0052), and merges. A new test that fails is a kernel bug. The lead then names the files and the rule under `Allowed Files` at tests review, and `kernel-engineer` fixes it in this task.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
All in `pkg/engine/dek_rep_inline_test.go`:
- `TestFR_DEK_002_ShuffleKeepsEveryCardInTheDeck`
- `TestFR_DEK_002_ShuffleUsesNoStackItem`
- `TestFR_DEK_002_ShuffleTouchesOnlyTheSearchedDeck`
- `TestFR_REP_002_RedirectionUsesNoStackItem`
- `TestFR_REP_002_ResistantUsesNoStackItem`
- `TestFR_REP_002_ReplacedDeathFiresNoEntersDiscardTrigger`

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- (none)

## Allowed Files
- None expected (tests only). The lead names engine files here at tests review only if a new test fails.

## Done When
- [ ] The new tests pass: `go test ./pkg/engine -run 'TestFR_DEK_002_|TestFR_REP_002_'`
- [ ] `go run ./cmd/cardtown lint rules` reports 58 of 62, its not-yet list is exactly `FR-ENT-102`, `FR-KWD-002n`, `FR-KWD-002p` and `FR-TRG-006`, and its trace check passes
- [ ] Existing tests pass unedited
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim-compare -- autoloop/m2-wave-h.3` passes (starter and stress sims byte-identical to the run branch: tests only)

## Log
- 2026-11-20 (lead-architect, wave H.3 session 1): task written from directive #177. Baseline on the run branch at 54c8db4: `cardtown lint rules` reports 56 of 62, not yet executable `FR-DEK-002`, `FR-ENT-102`, `FR-KWD-002n`, `FR-KWD-002p`, `FR-REP-002`, `FR-TRG-006`. #178 (`open: before the next Deliverables task`) and #179 (`open`) are this wave's next directives, after this task. Set `status: tests`; dispatching `rules-test-author` on the task branch, created from the run branch.
- 2026-11-20 (rules-test-author): wrote 6 tests in `pkg/engine/dek_rep_inline_test.go` (commit `test(T-0291)`). All 6 PASS on `go test ./pkg/engine -run 'TestFR_DEK_002_|TestFR_REP_002_'`, as expected for built rules; none fail. `cardtown lint rules`: "58 of 62 active FRD 01 requirements have a TestFR_* scenario test. Not yet executable: FR-ENT-102, FR-KWD-002n, FR-KWD-002p, FR-TRG-006"; trace check PASS. Stack-item tests compare "place" event counts against an identical scene without the replacement (a Maneuver or activation is itself placed), and never assert shuffle order. Helpers requested: none. Left out: Stalwart and Deploys Exhausted stack tests (Resistant covers the modification case; one behavior each). The replaced-death test uses the guardian synthetic scene, not the card.
