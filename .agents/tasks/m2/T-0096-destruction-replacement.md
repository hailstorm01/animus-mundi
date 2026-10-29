---
id: T-0096
title: Destruction replacement at the death sweep, one death at a time in its controller's order (FR-REP-001, SVY-R5, wave G shape, sonnet)
kind: shape
tier: sonnet
wave: G
status: in-progress
branch: m2/T-0096-destruction-replacement
depends_on: []
claims: [replacement:would-be-destroyed, subject:replaced, sweep:per-entity, event:replace, event:leave]
cards: []
requirements: [FR-REP-001, FR-LAY-002]
---

## Goal
A card can print a replacement: "If an allied unit would be destroyed, do these steps instead." At the death sweep, every lethal asset that such a replacement applies to is handled one at a time, in an order its controller chooses (`SVY-R5`). Its destruction is replaced in full: the asset is not destroyed, its death triggers do not fire, and the replacement's steps run inline, off the stack. A step can return the replaced asset to its owner's hand. Between two replaced deaths, the sweep destroys anything a byproduct newly left lethal, so a source destroyed that way stops applying. The other deaths of the sweep then happen normally, together. This is the first of four tasks for board directive #73 (see Log). *Shifting Impossibility* comes in T-0098, once kill steps (T-0097) and a Scratch Company filter (T-0098) exist.

## Inputs
- FRD 01:
  - `FR-REP-001` to `FR-REP-003` (lines 449-457), especially: Inline Event Substitution (line 450: a completely replaced event never occurs, and triggers watching for it do not fire); Consequential Event Isolation (line 453: the byproduct is a separate child event); Sequential Per-Entity Death Replacement (line 455); Ordering and Mid-Sweep Destruction (line 456, `SVY-R5`: the replacement's controller orders the events, each is replaced in full before the next, the sweep runs between them, and a source destroyed there stops applying).
  - `FR-LAY-002` (the death sweep), and line 196: a destroyed asset goes to its owner's discard pile.
  - Line 218 (`FR-ZON-004`, `FR-ZON-005`): a face-down unit that leaves the battlefield for Hand or Deck reverts to its printed card.
- Survey: wave G (line 35), Shifting Impossibility (line 135), `R5` (line 64).
- **User ruling, 2026-10-29 (not yet in FRD 01; recorded as design note Open Question 13 for `pm-frd-drafter`):** a replacement source that is itself lethal in the same sweep stays on the battlefield until the sweep's unreplaced deaths happen, so it replaces every death it applies to, then is destroyed. Its own death is one of the sweep's simultaneous events. Between replaced deaths, the sweep destroys only assets a byproduct *newly* left lethal: lethal now, but not when this sweep began. (Asked as: *Craterfall Ancient* deals 4 damage to all assets; *Shifting Impossibility* and three allied Scratch Companies are all lethal. Answer: it replaces all three, then dies.)
- Design note: `docs/architecture/m2-ability-model.md`, Key types (`Replacement`, line 42) and Kernel changes (line 54: "The death sweep becomes per-entity, with the controller ordering replaced deaths and a sweep between them").
- Kernel today:
  - `deathSweep` (`damage.go`) collects every lethal asset, `destroyAll`s them as one simultaneous event, then checks losses. `destroyAll` fires each asset's death triggers, records its last-known state, moves it to its owner's discard pile, and logs `destroy`.
  - `resolveAbility` (`effects.go`) verifies targets, then runs the step loop with `IfDone`. `stepAssets` and `stepCards` pick each step's subjects. `stepRelocate` moves only cards that are in a pile.
  - `flushTriggers` (`trigger.go`) asks for an order with one `DecideOrder` per player who has two or more triggers. Its fallback is collection order.
  - No replacement exists anywhere in `pkg/engine` (searched for `Replacement` and `WouldBeDestroyed`).

### Shape (lead decisions, binding)
- **Vocabulary (`ability.go`):**
  - `Def.Replacements []Replacement`, after `Triggers`: the printed replacement effects, in printed order.
  - `type Would int`, starting at `iota + 1` (an unset event is invalid). Its one member is `WouldBeDestroyed // an asset matching the replacement's Filter would be destroyed by the death sweep, being lethal (FR-REP-001, FR-LAY-002)`. The type name is the members' prefix, as #70 asks. A later `WouldTakeDamage` joins it.
  - `type Replacement struct { Event Would; Filter Filter; Steps []Step }`. The doc comment: "If an asset matching Filter would ..., do Steps instead (FR-REP-001). Filter is relative to the controller of the replacement's source. Steps run inline, never on the stack, with that source as their source: SubjectSelf is the source, and SubjectReplaced is the asset whose event is replaced."
  - `SubjectReplaced`, appended to `Subject`: the asset whose event a replacement replaces (`Replacement.Steps` only).
- **Finding replacements (`replace.go`, new; file comment `// Replacement effects: destruction replacement at the death sweep (FR-REP-001, SVY-R5).`):**
  - A replacement applies to asset `a` when its source is a face-up asset on the battlefield, the replacement's `Event` is the event, and `matchesFilter(source.Controller, source.ID, r.Filter, a)` holds. Collect the candidates in battlefield order, oldest asset first (as `assetsByID`).
  - A source that is lethal in this sweep still applies, until the sweep destroys it (user ruling above).
  - With several applicable replacements for one death, apply the first. Letting the affected asset's controller choose (`FR-REP-001`, "Simultaneous Replacement Conflicts") is T-0097. No synthetic test needs two here.
- **Replacing a death (`replace.go`):** log `Event{Kind: "replace", Player: <source's controller>, Source: <source ID>, Target: <replaced asset ID>}` with the text `"%s replaces the destruction of %s"`, using each asset's `String()`. Then run the steps through a `resolution` whose item is `&StackItem{Kind: StackAbility, Controller: <source's controller>, Source: <source ID>, Inst: <source's Inst>, Effect: ability.Effect{Steps: r.Steps}}`, with the new field `resolution.replaced` set to the asset. Nothing is pushed and nothing is placed. The replaced asset is not destroyed: no `fireKilled`, no `destroy` event and no discard. Byproduct damage goes through `dealDamage` as usual, so a damage trigger it causes is buffered like any other, and it is placed at the flush after the sweep.
- **One step loop:** move `resolveAbility`'s step loop (the `IfDone` check and `r.done`) into `(g *Game) runSteps(r *resolution)` in `effects.go`. `resolveAbility` and the replacement both call it, so the loop is not copied (D5.3).
- **`SubjectReplaced` (`effects.go`):**
  - `stepAssets` returns `r.replaced` while it is still on the battlefield, so damage, heal, armor and power steps can act on it.
  - `stepRelocate` moves it from the battlefield. It calls `leaveBattlefield(a, g.effectiveState(a))`, which keeps the last-known record (`FR-ZON-004`), then `putCard(a.Inst, ZoneHand, false)` (or `ZoneDeck` for `OpPutOnDeck`), and logs `Event{Kind: "leave", Player: -1, Target: a.ID, Zone: <destination>.String()}` with the text `"%s returns to its owner's hand"` (or `"... is put into its owner's deck"`). The event carries **no `Card`**, because a face-down asset's card is private in Hand or Deck. Its `String()` already says "face-down Scratch Company". It counts as one thing done.
- **The per-entity sweep (`damage.go`, calling `replace.go`):**
  1. `dying` is every lethal asset, as now. These are the sweep's simultaneous deaths.
  2. `queue` is the assets in `dying` that a replacement applies to. If it is empty, the sweep does exactly what it does today: same `destroyAll`, same order, same events. No executable card prints a replacement, so every existing test and both sims stay byte-identical.
  3. **Order:** each player orders the queued deaths of the assets they control, the active player's first. Use one `DecideOrder` per player with two or more queued deaths (prompt `"order the deaths your replacement effects apply to (first pick is replaced first)"`, one option per asset labelled with its `String()`, `Ref` its asset ID). The fallback is collection order, oldest asset first. In M2, every destruction replacement is allied, so the affected asset's controller is also the replacement's controller (`SVY-R5`). Ordering between two players' replacements is the lead's reading, recorded as design note Open Question 13.
  4. **Process each queued asset in order.** Skip it if it has left the battlefield, is no longer lethal, or no replacement applies any longer (its source has left), so it stays among the normal deaths. Otherwise replace its death, then run the **interim sweep:** the assets lethal now that were not in `dying`, newly left lethal by a byproduct, go through steps 2 to 5 as a `dying` set of their own (recursively). That destroys them, or replaces their deaths, before the next queued asset. The outer sweep's remaining assets are not part of it.
  5. **The normal deaths:** `destroyAll` every asset of `dying` still on the battlefield and still lethal, as one simultaneous event. That covers the deaths no replacement applied to, queued deaths whose replacement stopped applying, and a source that was lethal from the start.
  6. **Losses:** a Headquarters destroyed anywhere in this sweep, interim sweeps included, loses at this sweep. Empty-deck draws are checked as now, once, at the outermost sweep.
- **Left for later:** kill steps (`stepKill`) go through replacement in T-0097. T-0097 also adds the choice between several replacements for one death, and the `doc.go` update. A Scratch Company filter and *Shifting Impossibility* come in T-0098. The combat-kill fix (#73) is T-0099.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
To be written in `pkg/engine/replacement_test.go`, with synthetic cards only. The file won't compile until `ability.Replacement`, `ability.WouldBeDestroyed` and `ability.SubjectReplaced` exist. The suggested synthetic source is a unit printing `Replacement{Event: WouldBeDestroyed, Filter: {Allegiance: Allied, Kinds: KindUnit, Other: true}, Steps: [{Op: OpReturnToHand, Subject: SubjectReplaced}, {Op: OpDealDamage, Amount: {N: 1}, Subject: SubjectSelf}]}`, "if another allied unit would be destroyed, return it to its owner's hand and deal 1 damage to this unit instead". Its Maximum Health sets how many deaths it can absorb. Required cases:
- **Replaced:** `TestFR_REP_001_ReplacedDeathReturnsTheUnitToItsOwnersHand`. An allied unit becomes lethal (for example from a damage step). At the sweep it goes to its owner's hand, not the discard pile, and the source takes 1 damage. The log has a `replace` event and a `leave` event for it, and no `destroy` event.
- **No death trigger:** `TestFR_REP_001_ReplacedDeathFiresNoDeathTrigger`. The replaced unit has a death trigger, and it does not fire.
- **Filter:** `TestFR_REP_001_ReplacementDoesNotApplyOutsideItsFilter`. An opposing unit lethal in the same sweep is destroyed normally.
- **Order (`SVY-R5`):** `TestFR_REP_001_ControllerOrdersTheReplacedDeaths`. The source has 2 Health, and three allied units are lethal at once. The scripted order answer decides which two return to hand. After the second byproduct, the source is lethal and is destroyed before the third death. The third unit is then destroyed normally: its `destroy` event comes after the source's.
- **Order fallback:** `TestFR_REP_001_ReplacedDeathOrderFallsBackToOldestFirst`. Same board, with no answer or a timeout: the two oldest units are replaced.
- **Source destroyed between deaths (`SVY-R5`):** `TestFR_REP_001_SourceLeftLethalByAByproductStopsApplying`. The source has 1 Health and two allied units are lethal. The first is replaced, the source is destroyed before the second, and the second goes to the discard pile.
- **Source lethal in the same sweep (user ruling, 2026-10-29):** `TestFR_REP_001_SourceLethalInTheSameSweepReplacesEveryDeath`. One "all assets" damage step leaves the source and three allied units lethal. All three units return to hand, then the source is destroyed.
- **Byproduct is a child event:** `TestFR_REP_001_ByproductDamageFiresTheSourcesDamageTrigger`. The source has a `HookDamaged` trigger. The byproduct damage fires it, and it is placed after the sweep.
- **Face-down unit:** `TestFR_ZON_004_FaceDownUnitReturnedToHandIsItsPrintedCard`. A face-down Scratch Company unit whose death is replaced goes to its owner's hand as its printed card instance. The `leave` event has no `Card`.
- **Non-active player:** `TestFR_REP_001_TheAffectedAssetsControllerOrdersTheirDeaths`. The non-active player's source and two of their units: the `DecideOrder` is asked of that player.
- **Existing tests:** every test in `pkg/engine` passes unchanged.
- **Left out:** a kill step's destruction (T-0097), two replacements for one death (T-0097), a Scratch Company filter (T-0098), and two players' replacements in one sweep (no M2 matchup reaches it: only Animus prints one).

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None. The file declares its own helpers (`guardianStructure`, `guardianUnit`, `(*scene).sweep`, `eventSeq`, `wantReplaced`) and reuses `mortal`, `killedTrigger`, `unbowed`, `healSelf`, `allDamage`, `trigUnit`, `deployFromHand`, `inHand`, `inDiscard`.

## Allowed Files
- `pkg/engine/ability/ability.go` (`Def.Replacements`, `Would`, `Replacement`, `SubjectReplaced`)
- `pkg/engine/replace.go` (create: finding, ordering and applying destruction replacements)
- `pkg/engine/damage.go` (`deathSweep` becomes per-entity)
- `pkg/engine/effects.go` (`runSteps`, `resolution.replaced`, `SubjectReplaced` in `stepAssets` and `stepRelocate`)

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run 'TestFR_REP_001|TestFR_ZON_004_FaceDownUnitReturnedToHand'`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] No executable card prints a replacement yet, so the sims are unchanged: `go tool task sim-compare` passes

## Log
- 2026-10-29 lead-architect: Wrote the task. Board directive #73 is `open`. Its named test needs *Shifting Impossibility*, and the kernel has no replacement machinery (`doc.go`: "replacement effects beyond the damage pipeline" are not built; `cardtown lint rules` lists `FR-REP-001` to `FR-REP-003` as not yet executable). The user chose to pull wave G forward (2026-10-29), ahead of wave C. The design note allows C to G in any order, and Impossibility is a deck card (#14 (a)). Asked the user one rules question for the tests (the source lethal in the same sweep). Answer: it replaces every death, then dies (Inputs; design note Open Question 13).
  - **The plan for #73, under the D2 size rule (four non-test files each), all serial:**
    - T-0096, this task: replacement at the death sweep.
    - T-0097 (shape, sonnet): kill steps go through replacement (`SVY-R15`, so a replaced kill still counts as done), and the affected asset's controller chooses when several replacements apply (`FR-REP-001`, `FR-REP-003`). Files: `effects.go`, `replace.go`, `choice.go` (the `DecideOrder` and `DecidePick` comments), `doc.go`.
    - T-0098 (shape, sonnet): `Filter.Scratch` ("Scratch Company unit") plus *Shifting Impossibility*, the card that first uses it. Files: `ability.go`, `targeting.go`, `cards/animus.go`, `cards/cards.go`.
    - T-0099 (sonnet, #73 itself): the combat-kill hook fires only when the sweep destroys the defender. `rules-test-author` pins #73's named test with *Eternal Watchman* and *Shifting Impossibility*. Files: `combat.go`, `damage.go` or `replace.go`, `trigger.go`, `ability.go` (the `HookCombatKill` comment).
  - **Directives:** read `BOARD.md` on `main` (d015ef4) fresh. #14 is a standing order, and Impossibility is a deck card. #73 is `open`, and this task starts it. Waiting triggers checked: #6 and #7 (mode is `serial`); #8 (no `deck.go` change); #15 (no turn hook); #58 (no exhaust by a cost or an effect); #61 (no modal card); #71 (no effect on each player). None has fired.
  - **Reuse and placement:** searched `pkg/engine` and the `done` tasks for a replacement, a replaced-asset subject or a battlefield-to-hand move, and found none. The Boundaries table gives replacements to `pkg/engine`, so a new `replace.go` there holds them. The step loop is shared (`runSteps`), not copied.
  - **Deck cards still on `NotYetExecutable` (#14 (e)):** 23 (10 Animus, 13 Path).
- 2026-10-29 rules-test-author: Wrote `pkg/engine/replacement_test.go` (10 tests, no helpers requested). The file does not compile until `ability.Replacement`, `ability.WouldBeDestroyed`, `ability.SubjectReplaced` and `Def.Replacements` exist (`go vet` reports only these). `cardtown lint rules` shows `FR-REP-001` traced.
  - Tests: `TestFR_REP_001_ReplacedDeathReturnsTheUnitToItsOwnersHand`, `..._ReplacedDeathFiresNoDeathTrigger`, `..._ReplacementDoesNotApplyOutsideItsFilter`, `..._ControllerOrdersTheReplacedDeaths`, `..._ReplacedDeathOrderFallsBackToOldestFirst`, `..._SourceLeftLethalByAByproductStopsApplying`, `..._SourceLethalInTheSameSweepReplacesEveryDeath` (cites FR-REP-001 and the 2026-10-29 user ruling), `..._ByproductDamageFiresTheSourcesDamageTrigger`, `..._TheAffectedAssetsControllerOrdersTheirDeaths`, and `TestFR_ZON_004_FaceDownUnitReturnedToHandIsItsPrintedCard`.
  - Deviation from the suggested source: the guardian in the order, byproduct, filter, face-down and non-active tests is a structure, so the unit sweeper does not also damage it (the task's own order case needs the source to take only the byproduct damage). The same-sweep ruling test uses a unit guardian. The sweeper is a deployed unit whose deploy trigger deals 1 damage to a filter.
  - Order tests expect one `DecideOrder` per player with the queued deaths as options in oldest-first order, so picks `[2,0,1]` mean the third unit is replaced first. The unscripted-seat cases rely on the default chooser's timeout fallback.
  - Left out, as the task says: kill steps, two replacements for one death, a Scratch Company filter, two players' replacements in one sweep.
- 2026-10-29 lead-architect (D4): **Tests approved, `ready`.** Checked the 10 tests in 935e42b against `FR-REP-001` (lines 450, 453, 455, 456), `SVY-R5`, line 218 (`FR-ZON-004`, `FR-ZON-005`) and the user ruling of 2026-10-29. Each expected outcome follows from the cited text or the task's binding Shape:
  - the replaced unit goes to hand, not the discard pile, with no death trigger and no `destroy` event;
  - an opposing unit outside the filter dies normally;
  - the order answer, and the oldest-first fallback, decide which deaths a 2-Health source absorbs. The source is destroyed before the unreplaced death, at the interim sweep;
  - a source lethal in the same sweep replaces all three deaths, then dies;
  - the byproduct's damage trigger is placed after the sweep;
  - a face-down unit returns as its printed instance, with no `Card` on its `leave` event;
  - the non-active player orders their own deaths.
  - **Structure source:** using a structure in most tests is sound. It keeps the sweeper's damage off the source, so only byproducts count.
  - **Fallback:** the fallback cases rely on a seat with no chooser. `choose` applies the `FR-MOD-001` §6 fallback there (collection order for `DecideOrder`), so nothing new is needed.
  - **Helpers:** every helper used exists (`putScratch`, `allDamage`, `unbowed`, `healSelf`, `mortal`, `killedTrigger`, `drawN`, `deployTrigger`, `trigUnit`, `structure`, `inHand`, `inDiscard`, `eventsOfKind`, `asked`, `alive`, `wantDamage`, `script`, `deployFromHand`).
  - **Events:** the `replace` and `leave` kinds are a public contract (ADR-0005) fixed by the Shape, so asserting on them is not an implementation detail.
