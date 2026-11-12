---
id: T-0222
title: OpLookTop, a look at the top cards of a deck with free deploys and a bottom order in one prompt (wave F shape, sonnet; #140 (3))
kind: shape
tier: sonnet
wave: F
status: in-progress
branch: m2/autoloop-wave-f/T-0222-look-top
depends_on: [T-0221]
claims: [shape:OpLookTop (ability.OpLookTop; stepLookTop; the one look decision DecideLook, its outcome options and fallback); the free-deploy zone list freeDeployZones shared with stepDeployFree]
cards: []
requirements: [FR-MOD-001, FR-TRN-003a, FR-CST-001]
---

## Goal
The `LOOK_TOP_CARDS` partition (`FR-MOD-001` §4, §6; `SVY-R41`–`SVY-R43`; directive #140 (3)): a resolving step has its controller look at the top N cards of their deck and, in one hidden prompt, choose which of them (up to a number, matching a card filter) are deployed for free, the zone of each, and the order the rest go to the bottom. The design note's Open Question 5 left "the answer encoding" to this task; it is fixed below. Second of three tasks for *Devilhorn Scouting Party* (T-0221 before it, `done`; T-0223 after, `queued`).

Covers:
- `ability.OpLookTop`, acting for the ability's controller: `Amount` is N, `Filter` the card-level criteria a deployed card must meet (read with T-0221's `cardMeets`; no zone criterion), `UpTo` how many may be deployed (`UpTo` N > 0: up to N, possibly none; 0: exactly one, the required pick of `SVY-R41`, when any card is eligible). The only destination besides the bottom is a free deploy; a look into a hand (*Hourglass Diviner*, *Light of the World Soul*, outside M2) is a later field.
- **The answer encoding:** one `Decision` of a new kind `DecideLook`, whose options are whole outcomes, the first listed being `SVY-R41`'s and `SVY-R42`'s timeout.
- One free-deploy zone list, `freeDeployZones`, shared with `stepDeployFree` (a pure move out of it), so the eligibility and zone rule of `FR-TRN-003a` 1 and 2 is written once.
- `doc.go`'s step list names `OpLookTop`, and its decisions line names `DecideLook`.
- Synthetic-card tests only; no card prints it until T-0223.
- Done by the lead with this task's commit: the design note's Open Question 5 names this encoding.

- Not this task: *Devilhorn Scouting Party* itself (T-0223, `queued`); a look into a hand; an event for the cards put on the bottom (the decision event records the outcome, and replay re-runs it).

## Inputs
- FRD 01 v3.24: `FR-MOD-001` (line 410 on): §1 (line 411, no fast path for a deck inspection), §4 (line 414, one prompt for the partition and each deploy's zone, `SVY-R43`), §6 *Top-of-Deck Inspections* (line 417, `SVY-R41`, `SVY-R42`: an optional pick times out to no card, a required one to the first eligible card from the top, and the rest go to the bottom in the order seen, top first) and *Free-Deploy Zone* (line 419, Reserves); `FR-TRN-003a` 1 and 2 (line 364 on: a free deploy goes to a zone the card prints, the player choosing between Reserves and the Main Theater, unless the effect names one); `FR-CST-001` (line 314: no cost is paid).
- Rulings: `SVY-R41`, `SVY-R42`, `SVY-R43` (FRD 01 lines 27–29; outcome `docs/design/outcomes/2026-11-09-wave-f-keyword-grant-duration-and-inspection-timeout.md`); directive #140 (3).
- Design note: `docs/architecture/m2-ability-model.md`, Open Question 5 (line 185).
- Code (run branch at 1b6e9d1):
  - `ability.Op` list and `Step` ([ability.go:320-363](../../../pkg/engine/ability/ability.go)): `Amount`, `Filter`, `UpTo` and `To` are read as below; their comments name `OpLookTop` where it reads them.
  - `runStep` and `stepDeployFree` ([effects.go:84-158](../../../pkg/engine/effects.go)); `stepSearchDeck` ([effects.go:160](../../../pkg/engine/effects.go), the other hidden deck decision); `g.value(r.it.Source, st.Amount, nil)` reads `Amount` ([effects.go:284](../../../pkg/engine/effects.go)).
  - `DecisionKind`, its `String`, `choose` and `fallbackPicks` ([choice.go:13-142](../../../pkg/engine/choice.go)). A `Hidden` decision takes no fast path ([choice.go:105](../../../pkg/engine/choice.go)), and `view.LogLines` already hides any other seat's hidden decision whatever its kind ([view/log.go:32](../../../pkg/view/log.go)).
  - `cardMeets` ([filter.go:36](../../../pkg/engine/filter.go)); `takeCard`, `putCard(inst, ZoneDeck, true)` (the bottom), `deployCard` ([zones.go:43](../../../pkg/engine/zones.go), [:126-157](../../../pkg/engine/zones.go)); `battlefieldZone` ([zones.go:247](../../../pkg/engine/zones.go)); `Card.Supported` ([card.go:122](../../../pkg/engine/card.go)); `Zone.String` ([state.go:30](../../../pkg/engine/state.go)). `p.Deck[0]` is the top card, and `scratch` takes it ([zones.go:53-66](../../../pkg/engine/zones.go)).
- Searched `pkg/engine` and the task archive for an existing equivalent: no look op exists (`LookTop`, `OpLook`, `DecideLook` absent). The one free-deploy rule is `stepDeployFree`'s inline zone code, moved out below so the look doesn't copy it.

### Shape (the engineer builds it)
- **`ability.OpLookTop`**, after `OpGrantKeyword`. Comment: its controller looks at the top `Amount` cards of their deck and, in one hidden prompt, chooses which are deployed for free (up to `UpTo`, or exactly one when `UpTo` is 0 and one is eligible), each card matching `Filter`'s card-level criteria, the zone of each, and the order the rest go to the bottom (`FR-MOD-001` §4, `SVY-R41`–`SVY-R43`). Add `OpLookTop` to the comments of `Step.Filter`, `Step.UpTo` (same meaning as for `SubjectPick`) and `Step.To` (overrides the printed zones, as for `OpDeployFree`). The subject is always the controller (`Subject` unread).
- **`freeDeployZones(c *Card, to ability.Zone) []Zone`** in `effects.go` (a pure move out of `stepDeployFree`): nil unless `c` is a supported Unit or Structure; `[to]` if `battlefieldZone(to)` names one (`FR-TRN-003a` 2); else the printed Reserves, then the Main Theater, whichever `c.DeployZones` contains (`FR-TRN-003a` 1; Reserves first, `SVY-R43`). `stepDeployFree` then reads: skip a card not in its discard pile or with no zones; ask its `DecideZone` only when the list has two zones (Reserves listed first, a Main Theater pick takes the second); else the one zone. Its behavior doesn't change: every `deploy_free_test.go` and card test passes unchanged, and `go tool task sim-compare -- autoloop/m2-wave-f` is byte-identical.
- **`DecideLook`**, after `DecideZone`, `String` "Look", commented: a look at the top cards of your own deck (`FR-MOD-001` §4): each option is a whole outcome; a deck inspection, so no fast path; fallback takes the first listed (`SVY-R41`, `SVY-R42`, `SVY-R43`). `fallbackPicks` adds it to the first-listed case.
- **`stepLookTop(r, st) int`** in `effects.go`, `runStep`'s `case ability.OpLookTop`:
  1. The looked cards are the top `min(N, len(deck))` of the controller's deck, top first (look order). With none, nothing is asked and it returns 0.
  2. A looked card is **eligible** if `cardMeets(controller, r.it.Inst, st.Filter, inst)` and `freeDeployZones(inst.Card, st.To)` is not empty.
  3. **Deploy sets:** with `UpTo` N > 0, every set of 0 to N eligible cards; with `UpTo` 0, every set of exactly one eligible card, or only the empty set when none is eligible. Listed smaller sets first, and sets of one size in look order (lexicographic by look position).
  4. **Options**, one per outcome, listed in this order: for each deploy set in step 3's order; for each assignment of a zone to each card of the set, from its `freeDeployZones` list, lexicographic with the set's first card (in look order) varying slowest; for each order of the remaining looked cards, every permutation, lexicographic by look position, so the order seen comes first. So the first option is "deploy nothing (or, with `UpTo` 0, the first eligible card into its first zone), the rest in the order seen" (`SVY-R41`, `SVY-R42`, `SVY-R43`). Three cards, all eligible and printing both zones, with `UpTo` 1, give 6 + 3 × 2 × 2 = 18 options.
  5. Each option's `Label`: `Deploy nothing` or `Deploy <name> to <zone>[, <name> to <zone>…]` in look order, then `; bottom: <name>, <name>…` top first, or `; bottom: nothing` when no card is left (`Zone.String` names the zone). `Ref` 0. The labels are for the play screen and the log; tests read outcomes, not labels.
  6. The decision: `Player` the controller, `Kind: DecideLook`, `Prompt` "Look at the top N cards of your deck" (N the count looked at), `Min` and `Max` 1, `Hidden: true`. An invalid answer or a timeout takes option 0 (`choose`'s fallback).
  7. **On the answer:** every looked card leaves the deck (`takeCard`); the remaining ones go to the bottom in the chosen order (`putCard(…, ZoneDeck, true)` one by one, so the first named ends highest); then each chosen card enters through `deployCard(inst, controller, zone)` in look order (deploy triggers fire; it is not a move, `SVY-R1`). It returns how many cards it deployed, so an `IfDone` step after it runs only after a deploy.
- **`doc.go`:** the step list names `OpLookTop` (a look at the top cards with free deploys and a bottom order, `FR-MOD-001` §4), and the decisions line names "a look at the top cards of your own deck (DecideLook, SVY-R41–SVY-R43)".

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
In a new file `pkg/engine/look_top_test.go`, synthetic cards only, in the style of `deploy_free_test.go` (`maneuver`, `unit`, `s.onTop`, `s.give`, `s.script`, `enactAction`, `s.asked`). Player 0 enacts a Maneuver with no targets whose steps are an `OpLookTop` (`Amount` 3 unless the test says otherwise) and then an `OpScratch` 1, so every test also checks that the Scratch takes the deck's new top card. Unless a test says otherwise, `UpTo` is 1 and the `Filter` is `Kinds: ability.KindUnit, Stat: ability.StatRequisition, AtMost: 3`, and "eligible" means a supported unit costing 3 or less. Each test computes the answer index from the option order the Shape fixes (step 4) and states it in a comment, and checks the deck by instance order (`Players[0].Deck`, top first), the battlefield (asset, zone, controller) and the decisions asked. Give every card a distinct name; make the deck exactly the cards the test lists.
- `TestFR_MOD_001_LookTopTimeoutDeploysNothingKeepsOrderSeen` (`SVY-R41`, `SVY-R42`): deck A, B, C, D, with A–C eligible units printing both zones; seat 0 times out. Player 0 is asked one `DecideLook`: `Hidden`, `Min` and `Max` 1, 18 options, recorded as answered with option 0. Nothing is deployed from the look, and after the Scratch the deck is A, B, C, with D the Scratch Company in Reserves.
- `TestFR_MOD_001_LookTopKeepsThePlayersBottomOrder` (`FR-MOD-001` §4): the same deck; the answer is "deploy nothing, bottom C, A, B" (option 4). After the Scratch the deck is C, A, B.
- `TestFR_MOD_001_LookTopZoneChosenInTheSamePrompt` (`SVY-R43`, `FR-TRN-003a` 1): deck A (a Maneuver card), B (a cost-3 unit printing both zones), C (a cost-4 unit), D. Options: 6 with no deploy, then B to Reserves (bottom A, C; then C, A), then B to the Main Theater (A, C; C, A): 10 in all. The answer "deploy B to the Main Theater, bottom C, A" (option 9) puts B under player 0 in the Main Theater, and after the Scratch the deck is C, A. Player 0 is asked no `DecideZone`, and the `DecideLook` is the only decision asked.
- `TestFR_MOD_001_LookTopOffersOnlyEligibleCards` (`FR-TRN-003a`, `FR-TGT-001`): the top three are a Maneuver card, a cost-4 unit and a cost-1 unit with an `Unsupported` line, then a fourth card D. The one `DecideLook` has exactly 6 options (the bottom orders), and answering the last of them (option 5) deploys nothing and leaves the deck, after the Scratch, as the three looked cards in reverse look order.
- `TestFR_MOD_001_LookTopRequiredPickTimesOutToTheFirstEligible` (`SVY-R41`): `UpTo` 0; deck A (a Maneuver card), B and C (eligible units printing both zones), D; seat 0 times out. No option deploys nothing: 8 options (B or C, two zones each, two bottom orders). Option 0 deploys B into Reserves, and after the Scratch the deck is A, C.
- `TestFR_MOD_001_LookTopShortDeckLooksAtWhatIsThere` (`FR-MOD-001` §4): a deck of two eligible units printing both zones, `Amount` 3: the prompt names 2 cards, with 6 options (2 + 2 + 2). In a second scene, an empty deck asks no `DecideLook`, and the step after it (`OpScratch` 1) deploys nothing.
- `TestFR_MOD_001_LookTopIsHiddenSoNoFastPath` (`FR-MOD-001` §1): a deck of one Maneuver card: its single option is still asked of seat 0 (the decision event's `How` is "answer", not "auto").
- `TestFR_TRN_003a_LookTopDeployCountsForIfDone` (`SVY-R14`): a step `OpDraw` 1 marked `IfDone` right after the look (before the Scratch): deploying a card from the look draws one card; deploying none draws none.
- `TestFR_TRN_003a_LookTopDeployFiresDeployTriggers` (`SVY-R1`): a looked unit with a deploy trigger (as in the deploy-trigger scenes of `if_you_do_test.go` or `deploy_free_test.go`), deployed from the look, has its trigger placed on the stack after the Maneuver resolves.
- **Existing tests:** every other test passes unchanged, every `TestFR_TRN_003a_DeployFree*` test and the *Reaver's Rush* and *Allshan's Battle Trance* card tests included.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected. The test author keeps any small builder in the test file.

## Allowed Files
- `pkg/engine/ability/ability.go` (`OpLookTop`; the `Step.Filter`, `Step.UpTo` and `Step.To` comments)
- `pkg/engine/effects.go` (`freeDeployZones`, `stepDeployFree`, `stepLookTop`, `runStep`)
- `pkg/engine/choice.go` (`DecideLook`, `String`, `fallbackPicks`)
- `pkg/engine/doc.go` (the step list and decisions line)

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run "LookTop"`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-compare -- autoloop/m2-wave-f` passes (no printed card has `OpLookTop`, and `stepDeployFree`'s zone code is a pure move)

## Log
- 2026-11-11 lead-architect (wave F run, session 23): split off T-0221 under the D2 size rule, with T-0223; `queued` on T-0221.
- 2026-11-12 lead-architect (wave F run, session 24): Inputs, Shape, Tests and Allowed Files written against the run branch at 1b6e9d1. Searched `pkg/engine` and the task archive: no look op or look decision exists, and the free-deploy zone rule lives only in `stepDeployFree`, which this task moves into `freeDeployZones` so the look reuses it. Four non-test files: the look goes in `effects.go` beside `stepDeployFree`, not a new file, so it stays inside the D2 size rule. The design note's Open Question 5 is updated by the lead in this commit. `FR-DEK-001` dropped from `requirements` (a look is not a search; #106 not fired). Set to `tests`.
- 2026-11-12 rules-test-author: wrote pkg/engine/look_top_test.go: TestFR_MOD_001_LookTopTimeoutDeploysNothingKeepsOrderSeen, TestFR_MOD_001_LookTopKeepsThePlayersBottomOrder, TestFR_MOD_001_LookTopZoneChosenInTheSamePrompt, TestFR_MOD_001_LookTopOffersOnlyEligibleCards, TestFR_MOD_001_LookTopRequiredPickTimesOutToTheFirstEligible, TestFR_MOD_001_LookTopShortDeckLooksAtWhatIsThere, TestFR_MOD_001_LookTopIsHiddenSoNoFastPath, TestFR_TRN_003a_LookTopDeployCountsForIfDone, TestFR_TRN_003a_LookTopDeployFiresDeployTriggers. Helpers requested: none (small lt* builders live in the test file). go vet fails only on the missing ability.OpLookTop and DecideLook; cardtown lint rules passes. Deliberately left out: nothing from the Tests section. Notes: the timeout test pins the recorded Picks [0] but not the decision event's How (the task says answered, the event log records timeouts as timeout); the draw-count checks for IfDone and deploy triggers use the hand-size change (-1 for the enacted card). Set to tests-review.
- 2026-11-12 lead-architect (wave F run, session 24): tests reviewed (3969d3b): the nine tests match the Tests list and `FR-MOD-001` §1, §4, §6, `FR-TRN-003a` 1, `SVY-R41`–`SVY-R43`. Each answer index and option count was checked by hand against the Shape's order (18, 10, 6, 8, 6, 1; options 4, 9, 5, 6); `unit()` prints both zones, so the counts hold. Leaving `How` out of the timeout test is fine: the picks `[0]` pin the fallback. Set to `ready`.
