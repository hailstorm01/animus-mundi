---
id: T-0127
title: Make Rapturous Sermon executable (move all allied Scratch Company units to the Main Theater, then +2 power until end of turn, wave C card, flash)
kind: card
tier: flash
wave: C
status: tests-review
branch: m2/T-0127-rapturous-sermon
depends_on: [T-0069, T-0098, T-0104, T-0116, T-0117, T-0119, T-0121]
claims: [card:Rapturous Sermon]
cards: [Rapturous Sermon]
requirements: [FR-MNV-001, FR-ZON-002, FR-ZON-004, FR-LAY-001, FR-LAY-003]
---

## Goal
*Rapturous Sermon* (Animus starter deck, 2 copies) runs in the kernel. It is plain `ability` data in `pkg/engine/cards/animus.go`: a Maneuver `Effect` with no targets and two `SubjectAll` steps over allied Scratch Company units, `OpMoveZone` to the Main Theater (T-0119), then `OpModifyStat` +2 (T-0069). No new shape.

## Inputs
- Card (`data/cards/maneuvers-animus.csv` line 18): *Rapturous Sermon*, Animus Maneuver, cost 0 Requisition / 4 Tactics, one rules-text line:
  - "Move all allied Scratch Company units to the Main Theater. Allied Scratch Company units get +2 power until end of turn."
- Survey row: `docs/design/m2_starter_set_ability_survey.md` line 133: `MOVE_ZONE` all allied Scratch Company → Main Theater; `MODIFY_STAT` +2 until end of turn; `FR-LAY-001` snapshot locking; named in FRD 01.
- FRD 01:
  - `FR-MNV-001` (lines 332-334): enacted from Hand, costs paid on declaration, the effect runs when it resolves, then the card goes to its owner's Discard pile. It has no target slot, so `SVY-R29` never stops it being enacted.
  - `FR-ZON-002` (line 211) and "What Counts as Moving" (`SVY-R1`): an effect that says to move a unit to a zone moves it; moving is not deploying. Only units move.
  - `FR-ZON-004` (line 218): a unit moved between Reserves and the Main Theater keeps its damage, Armor, modifiers and readiness.
  - `FR-ZON-001` (line 206): zones are uncapped, so every allied Scratch Company unit moves.
  - `FR-LAY-001`, *One-Shot Modifier Snapshot Locking* (line 455, which names this card): the +2 goes to the allied Scratch Company units that match when that step resolves. A unit that becomes one later in the turn doesn't get it. A unit that got it keeps it if it moves zones on the battlefield, and loses it if it leaves the battlefield (`FR-ZON-004`).
  - `FR-LAY-003` (line 463) and Sub-Phase 4c (line 278): the +2 expires at the end of the turn.
- Shapes it uses, all `done`: `Def.Effect`, `ActEnact` and `StackManeuver` (T-0116), the `SVY-R29` enact legality (T-0117), `OpMoveZone` and `Step.To` (T-0119, `stepMoveZone`, `effects.go`), `OpModifyStat` (T-0069, `stepModifyStat`, `effects.go`: `TempPower`, cleared in `turn.go`), `SubjectAll` with an allegiance, kind and Scratch Company filter (`stepAssets`; `ability.ScratchOnly`, T-0098, T-0104), and the progress policy's enacts (T-0121).
- Rulings: `SVY-R1`, `SVY-R26` (a later step still acts on a moved unit; here the second step re-selects anyway). No other.

### Definition (the engineer adds it to `animusDefs`)
- *Rapturous Sermon*: `Text` holds its line. `Effect`:
  - `Steps`: two, in order, both with `Subject: ability.SubjectAll` and `Filter: ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit, Scratch: ability.ScratchOnly}` (no `Zone`):
    1. `{Op: ability.OpMoveZone, To: ability.ZoneMainTheater, ...}`
    2. `{Op: ability.OpModifyStat, Amount: ability.Value{N: 2}, ...}`
- No `Targets`, `Triggers` or `Modes`.
- Remove the name from `cards.NotYetExecutable`.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
To be added to `pkg/engine/cards_animus_maneuvers_test.go`, from the CSV text, loading the real card (`s.card`). Player 0 enacts; player 1 is the opponent. Give player 0 enough Tactics to pay (`enactFor`). Scratch Company units are face-down 1/1 Infantry units (`FR-KWD-002m`).
- `TestCard_RapturousSermon_MovesAlliedScratchCompanyToTheMainTheater`: player 0 has two Scratch Company units in its Reserves and one in the Main Theater. After it resolves, all three are in player 0's Main Theater, still the same assets (`g.Asset` of each ID). There is exactly one `move` event for each of the two that moved, and none for the one already there.
- `TestCard_RapturousSermon_AlliedScratchCompanyGetPlusTwo`: after it resolves, each of those three has power 3 (printed 1, +2).
- `TestCard_RapturousSermon_OthersAreUntouched`: an allied non-Scratch unit in the Reserves and an opposing Scratch Company unit in its controller's Reserves stay where they are, with their printed power, and no `move` event names them.
- `TestCard_RapturousSermon_MovedUnitsKeepTheirState` (`FR-ZON-004`): an allied Scratch Company unit in the Reserves, Exhausted, with 1 Armor (a 1-health unit can't hold non-lethal damage). After it resolves the unit is in the Main Theater, still Exhausted, with its 1 Armor and power 3.
- `TestCard_RapturousSermon_LaterScratchCompanyDoesNotGetPlusTwo` (`FR-LAY-001` snapshot locking): after it resolves, a new allied Scratch Company unit enters the battlefield the same turn (for example by enacting *Animoceps Spores*, or a scene helper that scratches). The new unit has power 1, and the earlier ones still have 3.
- `TestCard_RapturousSermon_PlusTwoExpiresAtEndOfTurn` (`FR-LAY-003`): after it resolves and the turn ends, the units are back to power 1 and stay in the Main Theater.
- `TestCard_RapturousSermon_EnactableWithNoScratchCompany`: with no allied Scratch Company unit, it is offered (no target slot, `SVY-R29`). Enacted, it moves nothing, logs no `move` event, and goes to its owner's discard pile.
- `TestCard_RapturousSermon_AsksNothingAndPaysFourTactics`: enacting it spends 4 Tactics and no Requisition, and the Chooser is asked no decision while it resolves.
- **Existing tests:** every other test passes unchanged, including `TestStarterSet_Executable` with the name removed, `TestView_Case8_ExecutableFlag` (seat 0's seed-1 hand still holds unsupported cards, *Sinuous Growth* among them), `TestProgress_Case4_Coverage` and `TestEvents_Case4_EveryActionKind`.

## Helpers Needed
- None expected. The tests use existing helpers (`newScene`, `give`, `card`, `do`, `script`, `asked`), those in `maneuver_test.go` (`enactAction`, `offersEnact`, `lastDiscard`) and `enactFor`, `spFaceDown` in `cards_animus_maneuvers_test.go`; the test file may add its own setup helpers.

## Allowed Files
- `pkg/engine/cards/animus.go` (the definition)
- `pkg/engine/cards/cards.go` (remove the name from `NotYetExecutable`)

## Done When
- [ ] The card tests pass: `go test ./pkg/engine -run TestCard_RapturousSermon`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] The card loads with no unsupported lines (`TestStarterSet_Executable` passes with the name removed)

## Log
- 2026-10-31 lead-architect: wave C run, session 14. No `main` commits to sync (`git log autoloop/m2-wave-c..main` is empty). Not `WAVE-COMPLETE`: 18 deck cards are on `NotYetExecutable`, 5 of them in this wave's `Cards` (*Rapturous Sermon*, *Rhapsody's Embrace*, *Funeral Fungus*, *Ambush Tactics*, *Steel Rain*). Directives read on the run branch: no `open: before` directive. #14 (a): deck cards first. *Rapturous Sermon* is the only one of the five whose shapes are all `done`: *Rhapsody's Embrace* and *Funeral Fungus* need an additional-cost shape (and *Funeral Fungus*'s exhaust cost fires #58), *Steel Rain* needs that shape plus #20 (b) and #59 (2)'s ruling on repeated targets, and *Ambush Tactics* waits on #61's fifth question and #88's distinctness ruling. No non-deck card has identical shapes, so none joins under #14 (c); *Ensnaring Mist*, T-0119's other card, has two "chosen unit" slots whose sameness FRD 01 doesn't settle (row 5). Waiting triggers checked: #8 (no `deck.go` change), #15 (no turn-boundary or draw hook; a +2 power modifier expiring in 4c can't make an asset lethal), #58 (no exhaust in `claims:`), #59 (2) (no repeated targets), #61's fifth question (no modal card in `cards:`), #71 (acts on allied units only), #88's distinctness (no slot). None has fired. Duplicates searched: no new code; the move goes through the one `stepMoveZone` and `moveOnBattlefield`, the +2 through the one `stepModifyStat`, the selection through the one `stepAssets`. Two non-test files. Status `tests`; dispatching `rules-test-author`.
- 2026-10-31 rules-test-author: wrote the eight tests in `pkg/engine/cards_animus_maneuvers_test.go`, as planned: `TestCard_RapturousSermon_MovesAlliedScratchCompanyToTheMainTheater`, `_AlliedScratchCompanyGetPlusTwo`, `_OthersAreUntouched`, `_MovedUnitsKeepTheirState`, `_LaterScratchCompanyDoesNotGetPlusTwo`, `_PlusTwoExpiresAtEndOfTurn`, `_EnactableWithNoScratchCompany`, `_AsksNothingAndPaysFourTactics`. Test-local helpers only (`rsScratchInMain`, `rsMoveEventsNaming`, `containsAsset`); no helpers requested. They compile, `cardtown lint rules` passes, and they fail now because the card is not yet executable. Notes: move events are matched by `Target` or `Source` asset ID (the event's asset field isn't specified in `go doc`); the "later" Scratch Company unit enters through `putScratch` rather than a second Maneuver. No cases left out.
