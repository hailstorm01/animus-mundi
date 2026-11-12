---
id: T-0223
title: "Devilhorn Scouting Party: a look at the top 3 with one free deploy of a unit costing R3 or less, the rest to the bottom, then Scratch 1 (wave F deck card, flash; #140 (3))"
kind: card
tier: flash
wave: F
status: in-progress
branch: m2/autoloop-wave-f/T-0223-devilhorn-scouting-party
depends_on: [T-0221, T-0222]
claims: [card:Devilhorn Scouting Party]
cards: [Devilhorn Scouting Party]
requirements: [FR-MOD-001, FR-TRN-003a, FR-KWD-002m, FR-ZON-004]
---

## Goal
*Devilhorn Scouting Party* (Path Maneuver, R0 T2; in `data/decks/path-starter.txt`): "Look at the top 3 cards of your deck. You may deploy a unit card with Requisition cost 3 or less from among them for free. Put the rest on the bottom of your deck in any order. Then, Scratch 1." is executable from shapes that are `done`: T-0222's `OpLookTop` (`Amount` 3, `UpTo` 1, `Filter` unit cards with T-0221's `StatRequisition` at most 3), then `OpScratch` 1. It leaves `cards.NotYetExecutable`. With it, directive #140 (3) is fully applied.

## Inputs
- Card text: `data/cards/maneuvers-path.csv:6` (Path Maneuver, R0 T2).
- FRD 01 sections:
  - `FR-MOD-001` §4 (line 414): a partitioned deck inspection is one prompt holding the free deploys, their zones and the bottom order (`SVY-R43`). *Scouting Party* is its example.
  - `FR-MOD-001` §6 (line 417): a timed-out *Scouting Party* deploys nothing and still Scratches 1; the rest go to the bottom in the order seen, top first (`SVY-R41`, `SVY-R42`).
  - `FR-MOD-001` §1 (line 411): a deck inspection has no single-option fast path, so the look always asks.
  - `FR-TRN-003a` (lines 363-364): a free deploy pays no costs (`FR-CST-001`). Item 1: with no zone in the text, the card enters one of its printed Deployment Zones, chosen when it prints both.
  - `FR-SET-004` Immediate Readiness (line 169): a newly deployed asset enters Readied unless it has Deploys Exhausted.
  - `FR-ZON-004` (line 244): a card off the battlefield has only its printed characteristics, so "Requisition cost 3 or less" reads the printed cost.
  - `FR-KWD-002m` (line 609): Scratch 1 deploys the top card of the deck face-down into Reserves; safe on an empty deck.
  - `SVY-R1` (line 240): a deploy is not a move, and deploy triggers fire.
- Survey row: `docs/design/m2_starter_set_ability_survey.md` line 205.
- Rulings that apply: `SVY-R41`, `SVY-R42`, `SVY-R43`.
- Directive #140 (3) (on `BOARD.md`): its tests are `_TimeoutDeploysNothingAndStillScratches` and `_ChosenBottomOrderKept` below.
- Shapes it combines, all `done`: `OpLookTop` (T-0222), the `StatRequisition` filter criterion (T-0221), `OpScratch`. `TestFR_MOD_001_LookTop*` (`pkg/engine/look_top_test.go`) is the synthetic form of this card: its `ltScene` Maneuver has this card's two steps.

## Shape (the lead's plan)
- `pkg/engine/cards/path.go`: the definition, in alphabetical order (after *Devilhorn Sanction*, before *Fortify*). `Text` is the CSV text, one line. No targets. Two steps: `OpLookTop` with `Amount: ability.Value{N: 3}`, `UpTo: 1` and `Filter: ability.Filter{Kinds: ability.KindUnit, Stat: ability.StatRequisition, AtMost: 3}` (no `To`: the card names no zone), then `OpScratch` with `Amount: ability.Value{N: 1}`. No new shape, no new field.
- `pkg/engine/cards/cards.go`: remove "Devilhorn Scouting Party" from `NotYetExecutable`.
- `pkg/engine/doc.go`: T-0222's review nit, comment only. Line 28 reads "… and from your own deck (DecideSearch, SVY-R37, SVY-R38), the zone of a free deploy (DecideZone, SVY-R43) and a look …" (a comma after `SVY-R38)`), and line 30 reads "… OpMorphScratch (a face-down transformation, FR-ENT-103), OpGrantKeyword …" (a comma after `FR-ENT-103)`). Nothing else on either line changes.
- The lead tried this definition locally on 2026-11-12, on the run branch at 66f37b3, without committing it. `go tool task sim` ran clean (Path 47/50 executable, P0 60.5% / P1 38.5% / draws 1.0%, every invariant holding), and `TestView_Case8_ExecutableFlag` passed unedited. One existing test failed: `TestReplay_OnlySearchRecordsAreHidden` (`pkg/eventlog/replay_test.go:586`). Its real games now hold a look, which is a hidden decision by design (T-0222, `SVY-R43`), and the view already redacts any hidden decision of another seat by its flag (#119, T-0183). Its authorized edit is below. With it, `go tool task check` passed.

## Tests (card tasks: written by rules-test-author; engineers never edit these)
New tests go in `pkg/engine/cards_path_scouting_party_test.go`. Use the real card (`s.card("Devilhorn Scouting Party")`), with player 0 enacting it from exactly R0 T2 (`setPool`), and set player 0's deck exactly with `onTop` (top first). Pick a `DecideLook` option by its label (T-0222's wording: `Deploy <name> to <zone>; bottom: <names>`, or `Deploy nothing; bottom: …`), not by a hard-coded index, except where a test pins the option count or a timeout. T-0222's `ltLooks`, `ltDeck`, `ltWantDeck`, `ltAsset`, `ltEligible` and `ltOther` may be reused. Real cards used below: *Chainbreaker Irregulars* (unit, R3, Reserves), *Chainbreaker Sneak* (unit, R1, Main Theater), *Devilhorn Guerrillas* (unit, R4, both zones), *Saboteur* (unit, R3, Reserves, a deploy trigger), *Devilhorn Sanction* and *Steel Rain* (Maneuvers). X is any fourth card.
- `TestCard_DevilhornScoutingParty_DeploysAUnitCostingThreeForFree`: deck *Chainbreaker Irregulars*, *Devilhorn Guerrillas*, *Devilhorn Sanction*, X. Player 0 picks "deploy the Irregulars to Reserves; bottom: Guerrillas, Sanction". After resolution:
  - exactly one `DecideLook` was asked, to player 0, hidden; no `DecideZone` was asked;
  - the Irregulars are an asset in player 0's Reserves, Readied (line 169), and out of the deck;
  - the pool is R0 T0: the Maneuver's T2 paid, the deploy free (`FR-CST-001`);
  - X became a face-down Scratch Company in player 0's Reserves (Scratch 1 took the deck's new top);
  - the deck, top first, is the Guerrillas, then the Sanction;
  - *Scouting Party* is in its owner's discard pile.
- `TestCard_DevilhornScoutingParty_OnlyUnitsCostingThreeOrLessAreOffered` (`FR-ZON-004`): deck *Devilhorn Guerrillas*, *Devilhorn Sanction*, *Chainbreaker Sneak*, X. No option deploys the Guerrillas (R4) or the Sanction (a Maneuver). The options are the 6 bottom orders deploying nothing and the 2 that deploy the Sneak to the Main Theater (its only zone), 8 in all.
- `TestCard_DevilhornScoutingParty_NothingEligibleStillAsks` (`FR-MOD-001` §1, §4): deck *Devilhorn Guerrillas*, *Devilhorn Sanction*, *Steel Rain*, X. One `DecideLook` is still asked, with 6 options, all deploying nothing. Player 0 picks the order Steel Rain, Guerrillas, Sanction. After X is scratched, the deck is Steel Rain, Guerrillas, Sanction (top first).
- `TestCard_DevilhornScoutingParty_TimeoutDeploysNothingAndStillScratches` (#140 (3), `SVY-R41`, `SVY-R42`, `FR-MOD-001` §6): deck *Chainbreaker Irregulars*, *Chainbreaker Sneak*, *Devilhorn Guerrillas*, X; player 0 times out. Nothing is deployed (neither the Irregulars nor the Sneak is an asset), the deck is Irregulars, Sneak, Guerrillas (the order seen, top first), and X became a Scratch Company in Reserves.
- `TestCard_DevilhornScoutingParty_ChosenBottomOrderKept` (#140 (3)): deck exactly *Devilhorn Guerrillas*, *Devilhorn Sanction*, *Steel Rain* (no fourth card). Player 0 deploys nothing and picks the order Sanction, Steel Rain, Guerrillas. Scratch 1 then takes the Sanction (the highest of the three) face-down into Reserves, and the deck is Steel Rain, Guerrillas.
- `TestCard_DevilhornScoutingParty_ZoneIsPartOfTheLook` (`SVY-R43`, `FR-TRN-003a` 1): deck a synthetic R3 unit printing both zones (`ltEligible`), two `ltOther` Maneuvers, X. Options deploying it exist for each zone. Player 0 picks one deploying it to the Main Theater: it is an asset in player 0's Main Theater, and no `DecideZone` was asked.
- `TestCard_DevilhornScoutingParty_DeployTriggerFires` (`SVY-R1`): deck *Saboteur*, two `ltOther` Maneuvers, X. Player 1 has one non-Scratch unit in the Main Theater with health above 2, and player 0 one unit in Reserves. Player 0 deploys the Saboteur. Its deploy trigger resolves: the opposing unit takes 2 damage, and one allied unit (whichever the trigger chose) takes 1.
- `TestCard_DevilhornScoutingParty_EmptyDeck` (`FR-KWD-002m`): player 0's deck is empty. *Scouting Party* can be enacted and resolves: no `DecideLook` is asked, no Scratch Company enters, player 0 has not lost, and the card is in its owner's discard pile.
- `TestDefs_MatchRulesText` already checks `Text` against the CSV. Nothing to add there.

### Authorized edit to an existing test (in the `test(T-0223)` commit)
- `TestReplay_OnlySearchRecordsAreHidden` in `pkg/eventlog/replay_test.go` (line 586), from T-0183: rename it `TestReplay_OnlySearchAndLookRecordsAreHidden`, and count a look as hidden like a search. Exactly:
  - its comment becomes `// T-0183 (directive #119), T-0223: in real games every search record and every` / `// look record is flagged hidden and no other record is, in the record and in its` / `// JSON, where an open decision has no hidden key.`;
  - `isSearch := e.Decision.Kind == engine.DecideSearch.String()` becomes `isSearch := e.Decision.Kind == engine.DecideSearch.String() || e.Decision.Kind == engine.DecideLook.String()`;
  - the `t.Fatalf` message `want Hidden only on a search` becomes `want Hidden only on a search or a look`.

  Nothing else in the test changes. Before the implementation commit it still passes (no real game plays the card yet). After it, the unedited test fails (the lead's trial: game 31's look).

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- none expected. Existing helpers: newScene, card, give, setPool, onTop, put, script, do, asked, enactAction, inDiscardPile, alive, wantDamage, unit, maneuver, scratchN, and T-0222's lt* helpers in `look_top_test.go`.

## Allowed Files
- `pkg/engine/cards/path.go`
- `pkg/engine/cards/cards.go`
- `pkg/engine/doc.go`

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run "DevilhornScoutingParty|TestDefs"`
- [ ] The edited test passes: `go test ./pkg/eventlog -run TestReplay_OnlySearchAndLookRecordsAreHidden`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-11 lead-architect (wave F run, session 23): split off T-0221 under the D2 size rule, with T-0222; `queued` on T-0221 and T-0222.
- 2026-11-12 lead-architect (wave F run, session 24): T-0222 is `done` (merge 143be71); its `doc.go` separator nit added to the Covers.
- 2026-11-12 lead-architect (wave F run, session 25): written in full against the run branch at 66f37b3 (Inputs, Shape, Tests, Allowed Files). A card task with three non-test files and no new shape. The local trial found one existing test to edit (`TestReplay_OnlySearchRecordsAreHidden`), named above. Waiting triggers checked: #6, #7, #15, #71, #91 (2), #106 (a look is not a deck search), #120, #136, #141 (3) and #142 don't fire. Set to `tests`.
- 2026-11-12 rules-test-author: tests written in pkg/engine/cards_path_scouting_party_test.go: TestCard_DevilhornScoutingParty_{DeploysAUnitCostingThreeForFree, OnlyUnitsCostingThreeOrLessAreOffered, NothingEligibleStillAsks, TimeoutDeploysNothingAndStillScratches, ChosenBottomOrderKept, ZoneIsPartOfTheLook, DeployTriggerFires, EmptyDeck}. Authorized edit made: TestReplay_OnlySearchRecordsAreHidden renamed TestReplay_OnlySearchAndLookRecordsAreHidden (passes now). Helpers requested: none (test-local sct* helpers in the new file). Cases left out: none. Red now: Enact rejected, the card is still in NotYetExecutable.
- 2026-11-12 lead-architect (wave F run, session 25): tests reviewed (9b3a916). Seven of the eight, and the authorized edit to `TestReplay_OnlySearchAndLookRecordsAreHidden`, match the Tests list and the FRD 01 text they cite. All eight fail on the task branch and pass against the Shape's definition, tried locally and discarded. Returned for one note:
  1. `TestCard_DevilhornScoutingParty_DeployTriggerFires` asserts `ally.Damage + sab.Damage == 1`, but X's face-down Scratch Company (scratched by the same resolution, before the Saboteur's trigger is placed) is also an allied unit, so "chosen allied unit" may legally pick it, and the test would then fail on a legal outcome. It passes today only because of the kernel's fixed target order (an implementation detail). Sum the damage over every asset player 0 controls and want exactly 1 (and none on player 1's side other than the enemy's 2), as the Tests list says: "one allied unit (whichever the trigger chose) takes 1". Change nothing else.
- 2026-11-12 rules-test-author: note 1 fixed: DeployTriggerFires now sums damage over every asset player 0 controls and wants exactly 1 (the enemy still takes 2). Nothing else changed.
- 2026-11-12 lead-architect (wave F run, session 25): tests re-reviewed (f9120ce). Note 1 is fixed as asked: `_DeployTriggerFires` sums the damage over every asset player 0 controls, and nothing else changed. All eight tests match the Tests list and the FRD 01 text they cite (`FR-MOD-001` §1, §4, §6, `FR-TRN-003a` 1, `FR-CST-001`, `FR-SET-004`, `FR-ZON-004`, `FR-KWD-002m`, `SVY-R1`, `SVY-R41`–`SVY-R43`), and they assert outcomes, picking look options by label. The authorized edit in `pkg/eventlog/replay_test.go` is exactly the one named. Tests approved; set to `ready`.
