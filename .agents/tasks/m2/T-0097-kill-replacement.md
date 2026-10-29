---
id: T-0097
title: Kill steps go through destruction replacement, and a replaced kill still counts as done (FR-REP-001, SVY-R15, directive #73 part 2, wave G shape, sonnet)
kind: shape
tier: sonnet
wave: G
status: ready
branch: m2/T-0097-kill-replacement
depends_on: [T-0096]
claims: [replacement:kill-step]
cards: []
requirements: [FR-REP-001]
---

## Goal
A kill step's destruction goes through the destruction replacements T-0096 built, as the death sweep's does. Every asset a kill step kills that a replacement applies to has its death replaced in full, one at a time in its controller's order: it is not destroyed, its death triggers do not fire, and the replacement's steps run inline. The step's other assets are then destroyed together. Nothing is swept while the step's item resolves, so a replacement source left lethal by a byproduct keeps applying to every death of that step, and is destroyed at the sweep after the item (user ruling, 2026-10-29). A replaced kill still counts as done (`SVY-R15`), so an "if you do" step after it runs. This is the second of the tasks for board directive #73 (T-0096 Log; design note Deliverable 6c).

## Inputs
- FRD 01:
  - `FR-REP-001` to `FR-REP-003` (lines 449-457), especially Inline Event Substitution (line 450), Replaced Actions Still Count as Done (line 454, `SVY-R15`: the Infested Apostle example), and Ordering and Mid-Sweep Destruction (line 456, `SVY-R5`: the replacement's controller orders the events).
  - `FR-LAY-002` (the death sweep runs after an item resolves, not while it resolves).
- Survey: wave G (line 35), Shifting Impossibility (line 135: "a replaced kill still counts as done"), `R15` (line 74).
- **User ruling, 2026-10-29 (design note Open Question 13, "Kill steps"; not yet in FRD 01):** a replacement sees every death that happens at the same moment. When one kill step kills two allied Scratch Companies while *Shifting Impossibility* has 1 Health, both kills are replaced. Nothing is swept while an effect is resolving, so Impossibility is destroyed at the sweep after the item.
- Kernel today:
  - `stepKill` (`effects.go`) destroys its step assets except a Headquarters with one `destroyAll`, and returns how many it killed.
  - `sweepDeaths` (`damage.go`) queues the deaths a replacement applies to (`replacedQueue`, `replace.go`), replaces each still lethal one (`replaceDeath`) with an interim sweep after it, then destroys the rest of `dying` still on the battlefield and lethal.
  - `findReplacement` (`replace.go`) does not check whether the source is lethal, so a lethal source still applies until something destroys it.

### Shape (lead decisions, binding)
- **One replacement loop (`replace.go`):** move `sweepDeaths`' loop into `(g *Game) replaceDeaths(dying []*Asset, dies func(*Asset) bool, between func()) []*Asset`, doc comment: "replaceDeaths replaces the deaths of dying that a replacement applies to, one at a time in their controller's order (replacedQueue, SVY-R5), and calls between after each one it replaces. It returns the rest, to be destroyed together: the assets of dying still on the battlefield, for which dies still holds, whose death it did not replace."
  - For each asset of `replacedQueue(dying)`: skip it if it has left the battlefield or `dies` no longer holds; otherwise, if `replaceDeath` replaces it, record it as replaced and call `between`.
  - The rest keeps `dying`'s order. A replaced asset is never in it, even if it is still on the battlefield (`FR-REP-003`: a replacement applies once to each event, and that event is over).
  - With an empty queue it returns every asset of `dying` still on the battlefield with `dies` true, in order, and asks nothing.
- **The sweep (`damage.go`):** `sweepDeaths` becomes `g.destroyDying(g.replaceDeaths(dying, g.IsLethal, func() { g.sweepDeaths(g.newlyLethal(seen), seen, loses, reasons) }), loses, reasons)`, with its doc comment kept. Its behavior doesn't change: with no replacement, `dying` is exactly what it destroys now, in the same order, so every existing test and both sims stay byte-identical.
- **The kill step (`effects.go`):** `stepKill` collects its non-Headquarters step assets as now, then `g.destroyAll(g.replaceDeaths(dying, func(*Asset) bool { return true }, func() {}))`. No sweep between replaced kills, since nothing is swept while an item resolves (user ruling above). It still returns `len(dying)`, replaced kills included (`SVY-R15`). Doc comment: "stepKill kills each step asset except a Headquarters, as one simultaneous event. A death a replacement applies to is replaced instead, in its controller's order, before the rest are destroyed (replaceDeaths), with no sweep between: a source a byproduct leaves lethal applies to every death of this step and is destroyed at the sweep after the item (design note Open Question 13). A replaced kill still counts as done (SVY-R15). It returns how many it killed."
- **Vocabulary (`ability.go`):** comment only. `WouldBeDestroyed`'s comment becomes "an asset matching the replacement's Filter would be destroyed, by the death sweep (being lethal) or by a kill step (FR-REP-001, FR-LAY-002, SVY-R15)".
- **`findReplacement`'s comment (`replace.go`):** its last sentence becomes "Letting a's controller choose between several is T-0100's."
- **Left for later:** the choice between several replacements for one death, and the `doc.go` and `choice.go` comment updates, are T-0100. A Scratch Company filter and *Shifting Impossibility* are T-0098. The combat-kill fix (#73) is T-0099.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
To be written in a new file, `pkg/engine/replacement_kill_test.go`, with synthetic cards only. Reuse `replacement_test.go`'s helpers (`guardianReplacement`, `guardianStructure`, `guardianUnit`, `eventSeq`, `wantReplaced`, `threeVictims` where it fits) rather than copying them, and never edit `replacement_test.go`. A kill step comes from a synthetic trigger, for example a deploy trigger whose step is `{Op: OpKill, Subject: SubjectAll, Filter: ...}` or a targeted kill. Required cases:
- **Replaced kill:** `TestFR_REP_001_ReplacedKillReturnsTheUnitToItsOwnersHand`. A kill step kills an allied unit that a replacement applies to. It goes to its owner's hand, not the discard pile, and the source takes 1 damage. The log has a `replace` event and a `leave` event for it, and no `destroy` event. Its death trigger does not fire.
- **Still done (`SVY-R15`):** `TestFR_REP_001_ReplacedKillStillCountsAsDone`. A kill step followed by an `IfDone` step (for example "if you do, draw a card"). The kill is replaced, and the `IfDone` step still runs.
- **Kill steps, user ruling of 2026-10-29:** `TestFR_REP_001_OneKillStepsDeathsAreAllReplaced`. The source has 1 Health, and one kill step kills two allied units it applies to. Both return to hand, so the source took 2 damage, and it is destroyed only at the sweep after the item: its `destroy` event comes after both `leave` events.
- **Outside the filter:** `TestFR_REP_001_KillStepDestroysWhatNoReplacementAppliesTo`. One kill step kills an allied unit the replacement applies to and an opposing unit it does not. The allied unit returns to hand. The opposing unit is destroyed normally and its death trigger fires.
- **Source killed by the same step (the same-moment ruling):** `TestFR_REP_001_SourceKilledByTheSameStepReplacesTheOthersFirst`. One "kill all units" step kills the source (a unit) and two allied units. Both allied units return to hand, then the source is destroyed: its `destroy` event comes after both `leave` events.
- **Order (`SVY-R5`):** `TestFR_REP_001_ControllerOrdersTheReplacedKills`. A kill step kills two allied units the replacement applies to. One `DecideOrder` is asked of their controller, and the scripted answer decides the order of the two `replace` events.
- **Existing tests:** every test in `pkg/engine` passes unchanged, `replacement_test.go` included.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected. The new file may declare its own helpers.

## Allowed Files
- `pkg/engine/replace.go` (`replaceDeaths`; `findReplacement`'s comment)
- `pkg/engine/damage.go` (`sweepDeaths` calls `replaceDeaths`)
- `pkg/engine/effects.go` (`stepKill` calls `replaceDeaths`)
- `pkg/engine/ability/ability.go` (`WouldBeDestroyed`'s comment only)

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run 'TestFR_REP_001'`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] No executable card prints a replacement yet, so the sims are unchanged: `go tool task sim-compare -- autoloop/m2-wave-g` passes

## Log
- 2026-10-29 lead-architect (wave G run, session 1): Wrote the task, the second for board directive #73 (open: before the first wave C task).
  - **Split:** T-0096's plan gave T-0097 the kill steps and the choice between several replacements, with `effects.go`, `replace.go`, `choice.go` and `doc.go`. Done without a copied loop, the kill steps also need `damage.go` (the sweep calls the shared `replaceDeaths`) and a comment in `ability.go` (`WouldBeDestroyed` says "by the death sweep"). Six non-test files is over the D2 size rule, so the choice between replacements, with the `choice.go` and `doc.go` comments, is now T-0100. It goes after this task and before T-0098, and is numbered after T-0099 so that T-0098 and T-0099 keep the ids the board's #73 row and the design note already give them. Design note Deliverable 6c updated to match in this commit.
  - **Directives:** read `BOARD.md` on `autoloop/m2-wave-g` (294b6a2, the same as `main`) fresh. #14 is a standing order: Impossibility is a deck card. #73 is `open: before the first wave C task`, and this task continues it. Waiting triggers checked: #6 and #7 (mode is `serial`); #8 (no `deck.go` change); #15 (no turn hook); #58 (no exhaust by a cost or an effect); #61 (no modal card); #71 (no effect on each player). None has fired.
  - **Reuse:** the only kill path is `stepKill` (`destroyAll`'s callers are `stepKill` and `destroyDying`; no kill cost is built yet). The replacement loop is moved out of `sweepDeaths` into `replaceDeaths` and shared, not copied (D5.3).
  - **Deck cards still on `NotYetExecutable` (#14 (e)):** 23 (10 Animus, 13 Path).
- 2026-10-29 rules-test-author: Wrote `pkg/engine/replacement_kill_test.go` (6 tests, all fail until the engine change; vet and `cardtown lint rules` clean). Tests: `TestFR_REP_001_ReplacedKillReturnsTheUnitToItsOwnersHand`, `..._ReplacedKillStillCountsAsDone` (SVY-R15), `..._OneKillStepsDeathsAreAllReplaced` (user ruling 2026-10-29; 1-Health source), `..._KillStepDestroysWhatNoReplacementAppliesTo`, `..._SourceKilledByTheSameStepReplacesTheOthersFirst`, `..._ControllerOrdersTheReplacedKills` (SVY-R5).
  - **Helpers requested:** none. The file declares `killStep` and `(*scene).killer`, and reuses `replacement_test.go`'s helpers.
  - **Left out:** the "every existing test passes unchanged" case is the full suite, not a new test.
- 2026-10-29 lead-architect (D4): **Tests approved, `ready`.** Checked the 6 tests in f9ec020 against `FR-REP-001` (lines 450, 454, 456), `SVY-R15`, `SVY-R5` and the user ruling of 2026-10-29 (design note Open Question 13). Each expected outcome follows from the cited text or the binding Shape:
  - a replaced kill sends the unit to hand, not the discard pile, with a `replace` and a `leave` event, no `destroy` event and no death trigger;
  - the `IfDone` draw after a replaced kill runs (`SVY-R15`);
  - a 1-Health source replaces both kills of one step, so nothing was swept mid-item, and its `destroy` comes after both `leave` events;
  - an opposing unit outside the filter is destroyed normally, and its death trigger fires;
  - a source killed by the same step replaces the other two kills first;
  - one `DecideOrder` with two options orders the replaced kills.
  - **Killer:** `alliedUnits` and test 4's filter both set `Other`, so the killer unit never kills itself, and the guardian's own `Other` keeps it from replacing its own death.
  - **Helpers:** reused from `replacement_test.go` unedited (`alliedUnits`, `guardianStructure`, `guardianUnit`, `eventSeq`, `wantReplaced`), plus the file's own `killStep` and `(*scene).killer`.
