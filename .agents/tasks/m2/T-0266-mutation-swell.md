---
id: T-0266
title: "Mutation Swell: a counted kill cost {2, 4} on allied Scratch Company units, one different chosen unit card in the discard pile per 2 killed, each deployed for free (wave H card, flash)"
kind: card
tier: flash
wave: H
status: in-progress
branch: m2/autoloop-wave-h/T-0266-mutation-swell
depends_on: [T-0264, T-0265]
claims: [card:Mutation Swell]
cards: [Mutation Swell]
requirements: [FR-TRN-003a, FR-MNV-001, FR-TGT-001, FR-TGT-003, FR-MOD-001, FR-REP-001, SVY-R2, SVY-R15, SVY-R16, SVY-R21, SVY-R29, SVY-R43, SVY-R53, SVY-R54]
---

## Goal
*Mutation Swell* executes: *"As an additional cost to enact this maneuver, kill 2 or 4 allied Scratch Company units. For every 2 units killed this way, deploy a different chosen unit card from your discard pile for free."* It is T-0264's counted kill cost (`Cost.Counts`) on allied Scratch Company units, T-0265's per-kill slot with different picks (`Target.PerKilled: 2`, `Target.Different`) on allied unit cards in the discard pile, and one `OpDeployFree` step on that slot. Directive #158 (3) and (4). No new shape. The card leaves `cards.NotYetExecutable`.

## Inputs
- Card text (`data/cards/maneuvers-animus.csv:12`): Maneuver, Animus, Requisition 0, Tactics 3, rare. It is in no starter or stress deck.
- FRD 01 (`docs/frd/01_Game_Rules_FRD.md`):
  - the *Mutation Swell* sub-bullet of *Targets Are Locked on Placement* (`:401`): the count, the kills and one different target per 2 killed are chosen on declaration; the killed units' cards can't be chosen (`SVY-R2`); a target gone by resolution is skipped and the other still deploys (`SVY-R53`);
  - `FR-TRN-003a`: additional costs, and item 1, a free deploy into the card's printed zones;
  - `FR-MNV-001`: a Maneuver needs its required targets (`SVY-R29`);
  - `FR-TGT-001` *Repeated Targets* (`SVY-R54`); `FR-TGT-003`: second-stage verification;
  - `FR-MOD-001` §6 (choice of cost counts) and the fixed order (`SVY-R16`: piles from the top);
  - `FR-REP-001` (`SVY-R15`, `SVY-R21`): replaced kills still pay.
- Survey (`docs/design/m2_starter_set_ability_survey.md`): its row (`:169`), `SVY-R53` (`:112`), `SVY-R54` (`:113`), `SVY-R43` (`:102`, a zone timeout enters Reserves), `SVY-R21` (`:80`).
- Outcome: `docs/design/outcomes/2026-11-16-wave-h-mutation-swell-kill-count.md`; directive #158.
- Code of the day (run branch at `d8cc593`):
  - `ability.Cost.Counts` and `Cost.Counted` (`pkg/engine/ability/ability.go:134-165`); `Target.PerKilled`, `Target.Different` (`:308-318`); `OpDeployFree` (`:441`; with no `Step.To`, the card's printed zones, a `DecideZone` when it prints both, a timeout to Reserves).
  - `legalCounts`, `targetsLegal`, `slotsLegal`, `declare` (`pkg/engine/placement.go`); `repetitions`, `targetDecision` (`pkg/engine/targeting.go`); `payableCounts`, `chooseCostCount` (`pkg/engine/costs.go`).
  - *Reaver's Rush* (`pkg/engine/cards/path.go:734-756`) is the same slot and step, with a power limit and no repetition. *Shifting Impossibility* (`pkg/engine/cards/animus.go:610`, Health 4) replaces an allied Scratch Company unit's death.
  - `NotYetExecutable` (`pkg/engine/cards/cards.go:13-17`): *Mutation Swell* is its second entry.
  - T-0264 pinned the counted cost on synthetic Maneuvers in `pkg/engine/kill_count_cost_test.go`, and T-0265 the per-kill slot in `pkg/engine/per_kill_target_test.go` (counts narrowed to fillable targets, a repeated pick rejected, timeouts, a per-kill slot on a trigger). This task pins the real card.

## Shape (the lead's plan)
- `pkg/engine/cards/animus.go`: one definition, after *Melodic Pummeler* and before *Portent Scrivener*. `Text` is the CSV line, as one string.
  - `ExtraCosts`: one `{Kind: ability.CostKill, Filter: ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit, Scratch: ability.ScratchOnly}, Counts: []int{2, 4}}`.
  - `Effect.Targets`: one `{Filter: ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit, Zone: ability.ZoneDiscard}, PerKilled: 2, Different: true}`.
  - `Effect.Steps`: one `{Op: ability.OpDeployFree, Subject: ability.SubjectTarget, Target: 0}`, with no `To` (printed zones, `FR-TRN-003a` 1).
  - No new shape and no new field.
- `pkg/engine/cards/cards.go`: remove `"Mutation Swell"` from `NotYetExecutable`.
- The lead dry-ran this definition and the removal on the run branch at `d8cc593` (2026-11-17, wave H session 28):
  - `gofmt -l pkg cmd` clean.
  - `go tool task check` exit 0.
  - `go tool task sim` exit 0 (P0 58.5%, P1 39.5%, draws 2.0%).
  - `go tool task sim-compare -- autoloop/m2-wave-h` `[PASS]`.

## Tests (card tasks: written by rules-test-author; engineers never edit these)
New tests go in a new file `pkg/engine/cards_animus_mutation_swell_test.go`, with a file comment naming the card, its CSV line and T-0266, as `cards_path_monument_test.go` does.
- Use the real card (`s.card("Mutation Swell")`, given with `s.give`) and the real *Shifting Impossibility*. Player 0 is the active player and enacts.
- Allied Scratch Company units come from `s.putScratch(0)`. Their hidden cards are unit cards, so a killed one's card lands in the pile as a unit card. Pile cards come from `s.inDiscard(0, …)`, the last one on top.
- The helpers in `per_kill_target_test.go` may be reused (`pkCard`, `pkScript`, `pkPick`, `pkAsks`, `pkDeployed`, `pkTop`), as may `kcCountScript` and `kcRefs` from `kill_count_cost_test.go`. A free-deployed card that prints only Reserves asks no zone; `dfUnit` (`deploy_free_test.go`) sets printed zones.

Tests:
- `TestCard_MutationSwell_NotEnactableWithoutItsCostAndTargets` (`FR-MNV-001`, `SVY-R29`, `SVY-R53`):
  - 1 allied Scratch Company, an allied unit card in player 0's pile, and one Scratch Company of player 1's: not offered, and a rejected Enact leaves it in hand.
  - 2 allied Scratch Companies; player 0's pile holds only a non-unit card (a Maneuver), and player 1's pile holds a unit card: not offered.
  - Add a unit card to player 0's pile: offered.
- `TestCard_MutationSwell_OnlyTwoOfferedWhenFourCantBeFilled` (`SVY-R53`, `FR-MOD-001` §1, §6), two subtests:
  - `three units`: 3 allied Scratch Companies and 2 unit cards in the pile. No `DecideMode` is asked. The kill pick offers the 3 and takes exactly 2. One target is chosen. Exactly 2 Scratch Companies die, the chosen card is deployed, and the other stays in the pile.
  - `one unit card`: 4 allied Scratch Companies and 1 unit card. No `DecideMode` is asked. 2 Scratch Companies die and 2 live. The one card is deployed, with no target asked.
- `TestCard_MutationSwell_KillFourDeploysTwoDifferent` (`SVY-R53`, `SVY-R54`, `FR-TRN-003a` 1):
  - Setup: 4 allied Scratch Companies; pile cards A (prints only the Main Theater), B and C (each prints only Reserves), C on top.
  - The `DecideMode` offers "kill 2" then "kill 4"; answer 4. The target decision has `Min` = `Max` = 2, `Repeats` false, and options exactly A, B and C; answer A and B.
  - All 4 Scratch Companies are dead. A is in player 0's Main Theater and B in player 0's Reserves, both controlled by player 0. C is still in the pile. *Mutation Swell* is in player 0's discard pile.
- `TestCard_MutationSwell_KilledUnitsCardsAreNotTargets` (`SVY-R2`): 2 allied Scratch Companies and 2 unit cards X and Y in the pile. The target decision's options are exactly X and Y, not the killed Scratch Companies' cards. After resolution, both killed Scratch Companies' cards are in player 0's discard pile, not deployed.
- `TestCard_MutationSwell_ReplacedKillsStillDeploy` (`SVY-R15`, `SVY-R21`, `FR-REP-001`): player 0's *Shifting Impossibility* in Reserves, 4 allied Scratch Companies, 2 unit cards in the pile. Answer count 4. Every Scratch Company is returned to player 0's hand (none is in the discard pile), and both unit cards are deployed. Do not assert where *Shifting Impossibility* ends up: its 4 damage on Health 4 is lethal, and its fate isn't this card's rule.
- `TestCard_MutationSwell_GoneTargetSkippedOtherDeploys` (`FR-TGT-003`, `SVY-R53`): 4 allied Scratch Companies, unit cards P and Q in the pile, count 4, P and Q chosen. Once *Mutation Swell* is placed, P leaves the pile for player 0's hand (a `recorder` on the first `place` event, as `TestFR_TGT_003_DeployFreeTargetLeftPileDeploysNothing` does). P is not deployed and stays in hand; Q is deployed.
- `TestCard_MutationSwell_Timeout` (`SVY-R54`, `SVY-R16`, `FR-MOD-001` §6, `SVY-R43`), two subtests:
  - `everything timed out`: 2 allied Scratch Companies in Reserves and 2 in the Main Theater, 3 unit cards in the pile. The count is 2, the two Reserves Scratch Companies die, the two Main Theater ones live, and only the top card is deployed.
  - `count 4 answered, target timed out`: 4 allied Scratch Companies and 3 unit cards, the top one printing both Reserves and the Main Theater. The top two cards are deployed and the third stays. The both-zones card's `DecideZone`, timed out, puts it in Reserves.
- `TestDefs_MatchRulesText` already checks `Text` against the CSV.
- Not this task: a repeated answer rejected, a per-kill slot on a trigger, and a per-kill slot without `Different`. T-0265 pinned each on synthetic Maneuvers in `per_kill_target_test.go`, and the card adds nothing to them.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- none. The tests use existing helpers only; setup local to the new file stays in it.

## Allowed Files
- `pkg/engine/cards/animus.go`
- `pkg/engine/cards/cards.go`

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run "MutationSwell|TestDefs"`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-compare -- autoloop/m2-wave-h` passes (the card is in no starter or stress deck)

## Log
- 2026-11-17 lead-architect (wave H run, session 26): queued with T-0264, split off under the D2 size rule. Write its Inputs, Tests and Done When against the code of the day once T-0265 is `done`. Directive #158 (3), (4); when it merges, #158 is fully applied and moves to the archive.
- 2026-11-17 lead-architect (wave H run, session 28): T-0265 is `done` (merge 7cfde94) with the names it planned (`Target.PerKilled`, `Target.Different`, `legalCounts`).
  - Wrote Inputs, Shape, Tests and Allowed Files against the run branch at `d8cc593`, and dry-ran the definition (above). `requirements:` gains `SVY-R16`, `SVY-R29` and `SVY-R43`.
  - `main` has no commits the run branch lacks, so no sync.
  - Directives: #158 is `open: before T-0266`. This task applies parts (3) and (4), and #158 is fully applied when it merges. I checked each `waiting` one, and none has fired:
    - #6, #7: mode is `serial`.
    - #15 (2): no draw or end-of-turn hook here.
    - #71: no each-player effect.
    - #91 (2): no cost on a mode.
    - #106, #120, #142: no deck search with criteria, no mandatory discard cost, no tuck.
  - No new shape: the card is T-0264's `Cost.Counts`, T-0265's `PerKilled` and `Different`, and the `OpDeployFree` step *Reaver's Rush* uses.
  - Re-read `BOARD.md` before committing: no directive added since.
  - Status `tests`.
- 2026-11-17 rules-test-author: wrote `pkg/engine/cards_animus_mutation_swell_test.go` (8 tests, 11 with subtests), all as planned in Tests. `go vet ./pkg/engine` is clean, no helpers needed; they fail now because the card is still in `NotYetExecutable` ("needs rules the kernel does not implement yet"). No `TestFR_*` added, so `cardtown lint rules` has nothing new to trace (the binary was not on PATH here).
  - `TestCard_MutationSwell_NotEnactableWithoutItsCostAndTargets`, `_OnlyTwoOfferedWhenFourCantBeFilled` (`three units`, `one unit card`), `_KillFourDeploysTwoDifferent`, `_KilledUnitsCardsAreNotTargets`, `_ReplacedKillsStillDeploy`, `_GoneTargetSkippedOtherDeploys`, `_Timeout` (`everything timed out`, `count 4 answered, target timed out`).
  - Reuses `pkCard`, `pkScript`, `pkPick`, `pkAsks`, `pkDeployed`, `dfUnit`, `kcRefs`; local helpers `msScene` and `msSortedRefs` stay in the new file.
  - Left out as the task said: repeated answer, per-kill slot on a trigger, per-kill slot without `Different`; Shifting Impossibility's own fate in the replaced-kills test.
- 2026-11-17 lead-architect (D4, session 28): tests approved, round 1.
  - Each test checks what the cited rules give, on the real card:
    - Not enactable with one allied Scratch Company, or with two and no allied unit card in player 0's pile; a Maneuver in the pile and the opponent's unit card don't count; offered once a unit card is added (`FR-MNV-001`, `SVY-R29`, `SVY-R53`).
    - With 3 Scratch Companies, or with 4 and one unit card, only "kill 2" is legal and is taken unasked; the kill pick takes exactly 2 (`SVY-R53`, `FR-MOD-001` §1).
    - "kill 2" then "kill 4" offered; count 4 asks one target decision, `Min` = `Max` = 2, not `Repeats`, over exactly the pile's unit cards; each deploys into its printed zone (`SVY-R54`, `FR-TRN-003a` 1).
    - The killed Scratch Companies' cards are not options (`SVY-R2`); replaced kills still pay and both targets deploy (`SVY-R15`, `SVY-R21`); a target gone by resolution is skipped and the other deploys (`FR-TGT-003`, `SVY-R53`).
    - Timeouts take count 2, the Reserves Scratch Companies first, the top pile cards, and Reserves for a both-zones card (`SVY-R54`, `SVY-R16`, `SVY-R43`).
  - Only the new file; no existing test edited.
  - Sanity check in the review worktree: with the Shape's definition applied, `go test ./pkg/engine -run "MutationSwell|TestDefs"` passes; without it, the card's tests fail. The change was reverted.
  - Status `ready`, dispatched to `kernel-engineer` (flash).
