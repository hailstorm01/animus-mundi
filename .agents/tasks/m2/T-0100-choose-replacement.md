---
id: T-0100
title: When several replacements apply to one death, its controller chooses which, death by death (FR-REP-001, FR-MOD-001 §6, SVY-R23, directive #75, wave G shape, sonnet)
kind: shape
tier: sonnet
wave: G
status: tests
branch: m2/T-0100-choose-replacement
depends_on: [T-0101]
claims: [decision:replacement, replacement:choice]
cards: []
requirements: [FR-REP-001, FR-MOD-001]
---

## Goal
When two or more replacements apply to one death, the affected asset's controller is asked which one applies, once for each death, when that death is replaced: in the fixed replaced-death order (`SVY-R20`, `SVY-R22`), at the death sweep and for a kill step alike (`SVY-R21`). One applicable replacement applies with no question. A timeout picks the first source in `FR-MOD-001` §6's fixed order for untargeted selections, and a source's own replacements in printed order. `doc.go` lists destruction replacement as built. This applies board directive #75.

## Inputs
- FRD 01, `FR-REP-001` (lines 454-462): Simultaneous Replacement Conflicts & Controller Choice (line 457), Per-Entity Death Replacement (line 460), Same-Moment Deaths and Replacement Within a Death Sweep (`SVY-R20` to `SVY-R22`). `FR-REP-003` (line 456) holds trivially: a destruction replacement replaces the whole event, so no second one applies to it. `FR-MOD-001` §6 (the fixed order for untargeted selections, `SVY-R16`). FRD 01 v3.17 will add `SVY-R23`'s wording; this task doesn't wait for it, because the ruling is the user's.
- Survey §4 row R23 (asked; `pm-frd-drafter` records the ruling).
- Design Outcome: `docs/design/outcomes/2026-10-29-choosing-between-replacements.md`.
- Board directive #75.
- Design note: `docs/architecture/m2-ability-model.md`, Deliverable 6c and Open Question 14 (answered, this task's commit).
- Kernel today:
  - `findReplacement` (`replace.go`) returns the first applicable replacement in battlefield order, oldest source first, and asks nothing.
  - `replacedQueue` asks `findReplacement` only whether a replacement applies. `replaceDeath` asks it again, when the death is replaced.
  - `choose` (`choice.go`) asks a seat's Chooser and falls back per `fallbackPicks`. `fixedOrderPicks` sorts options stably by their `Ref`'s position in `fixedOrderPositions(d.Player)`, so options with the same `Ref` keep their listed order.

### Shape (lead decisions, binding)
- **All applicable replacements (`replace.go`):** `findReplacement` becomes `func (g *Game) replacementsFor(a *Asset) []replacementOption`, where `type replacementOption struct { src *Asset; rep *ability.Replacement }` (unexported, in `replace.go`). Same loop and tests as today's `findReplacement` (face-up, printed `WouldBeDestroyed`, filter matched), but it returns every match: sources in `assetsByID` order (option lists use it, `SVY-R11`), each source's replacements in printed order. Doc comment: "replacementsFor returns every replacement that applies to a's destruction now, oldest source first, each source's replacements in printed order: those printed by a face-up asset on the battlefield whose filter a matches. A source that is itself lethal still applies until the sweep destroys it (SVY-R20)."
- **The queue asks nothing (`replace.go`):** `replacedQueue` keeps an asset when `len(g.replacementsFor(a)) > 0`. Its doc comment is unchanged.
- **The choice, once per death (`replace.go`):** `replaceDeath(a)` takes `opts := g.replacementsFor(a)`. None: return false. One: apply it, with no `Decision` built and no decision event (`FR-REP-001` makes it a choice only when several apply). Two or more: `g.chooseReplacement(a, opts)`. Then the `replace` event and `runSteps` as today, with the chosen option's source and replacement. Doc comment gains: "When several apply, a's controller chooses one (chooseReplacement, SVY-R23)."
- **`chooseReplacement(a *Asset, opts []replacementOption) replacementOption` (`replace.go`):** asks `a.Controller` one `Decision{Kind: DecideReplacement, Prompt: fmt.Sprintf("choose which replacement applies to the destruction of %s", a), Min: 1, Max: 1}` through `g.choose`, with one `Option{Label: src.String(), Ref: src.ID}` per entry of `opts`, in `opts`' order, and returns `opts[picks[0]]`. When a source prints several applicable replacements, its options' labels add the replacement's printed position (`fmt.Sprintf("%s (replacement %d)", src, n)`, `n` from 1), so a player can tell them apart. `assetDecision` isn't reused: it lists each asset once, and one source can offer several replacements. Doc comment: "chooseReplacement asks a's controller which of several replacements applies to its destruction (FR-REP-001, SVY-R23). A timeout picks the first source in the FR-MOD-001 §6 fixed order, a source's own replacements in printed order (DecideReplacement's fallback)."
- **The decision kind (`choice.go`):** append `DecideReplacement` after `DecidePick`, so existing values don't change, with the comment "choose which of several replacements applies to one event (FR-REP-001, SVY-R23), asked of the affected asset's controller: same fallback as DecideTarget, the first source in fixed order (SVY-R16), a source's own replacements in listed (printed) order". `String` returns `"Replacement"`. `fallbackPicks`' `DecideTarget, DecidePick` case adds `DecideReplacement`. Nothing else in `choice.go` changes: `fixedOrderPicks`' stable sort already keeps a source's options in listed order.
- **`doc.go`:** in the built list, after the Ward/last-known bullet, add: "Destruction replacement (FR-REP-001): a face-up asset's printed "would be destroyed" replacement replaces a death at the death sweep or in a kill step, every death of one moment seeing its source (SVY-R20, SVY-R21), in a fixed order (SVY-R22). When several apply to one death, its controller chooses (DecideReplacement, SVY-R23)." In "Not built yet", "replacement effects beyond the damage pipeline" becomes "replacement effects other than destruction replacement and the damage pipeline (damage redirection)". In the Files list, add "replace.go: destruction replacement." after `damage.go`'s entry.
- **Unchanged:** `replaceDeaths`, `relocateReplaced`, `sweepDeaths`, `stepKill`, `fixedOrderPicks`, `fixedOrderPositions`, `choose`, and the clients: `cardtown play` asks a `Min == Max == 1` decision of any kind generically, the targeting monitor reads only `"Target"` decisions, and replay compares the kind's string.
- **Sims:** no executable card prints a replacement yet (*Shifting Impossibility* is T-0098), so no game reaches the new code, and the sims should stay byte-identical. If `sim-compare` differs, don't change code to make it match: report the first differing game in the Log and hand off.
- **Reuse:** no duplicate. `replacementsFor` replaces `findReplacement` (one loop over sources), and the timeout reuses `fixedOrderPicks` rather than a second fixed-order rule.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
Synthetic cards only, reusing the helpers in `replacement_test.go` (`guardianStructure`, `guardianUnit`, `guardianReplacement`, `eventSeq`, `wantReplaced`, `sweep`, `noOrderAsked`) and `replacement_kill_test.go` (`killStep`, `killer`). Assert on the `replace` events' `Source` and on each guardian's damage, and read the questions asked from the decision events (kind `"Replacement"`). No existing test is edited.

- `TestFR_REP_001_ControllerChoosesBetweenReplacementsForEachDeath` in `replacement_test.go` (`SVY-R23`): two guardian structures A (older) and B, 4 Health each, and two allied 1-Health units lethal at one sweep. Player 0 is scripted to pick A for the first death and B for the second. Two `DecideReplacement` questions are asked of player 0, one per death, each offering both guardians. Both units return to hand; the first unit's `replace` event has A as its source and the second's has B; A and B have each taken 1 damage.
- `TestFR_REP_001_ReplacementChoiceTimesOutToFixedOrder` in `replacement_test.go` (`SVY-R23`, `SVY-R16`): two guardian units, an older one in the Main Theater and a newer one in Reserves, and two allied 1-Health units lethal at one sweep (place the victims so the guardians aren't hit, and give the guardians enough Health to survive two byproducts). Player 0 times out on every question. Every death is replaced by the Reserves guardian (Reserves comes before the Main Theater in §6's order, though the Main Theater guardian is older), which takes 2 damage; the other takes none.
- `TestFR_REP_001_ReplacementChoiceTimesOutToPrintedOrderWithinASource` in `replacement_test.go` (`SVY-R23`): one source prints two replacements that both apply to an allied unit's death, the first returning it to its owner's hand and the second putting it on its owner's deck (or any two with visibly different results). Player 0 times out: the first printed replacement applies.
- `TestFR_REP_001_OneApplicableReplacementAsksNothing` in `replacement_test.go` (`SVY-R23`): one guardian and one allied unit lethal at a sweep, with player 0 scripted with an answer that would be wrong if asked. No `DecideReplacement` is asked of anyone, and the death is replaced.
- `TestFR_REP_001_LethalSourceIsStillAChoice` in `replacement_test.go` (`SVY-R20`, `SVY-R23`): one allied 1-Health unit, put first so it is the oldest, then guardian A (`guardianUnit(1)`) and guardian B (`guardianUnit(4)`). One damage step deals 1 damage to each of them, leaving the unit and A lethal at the same sweep. The unit's death is replaced first (oldest first, `SVY-R20`), and its question offers both guardians, A included though A is lethal. Player 0 is scripted to pick A: the unit returns to hand with A as the `replace` event's source. A's own death comes next: a guardian's filter is "another allied unit", so only B applies, and no question is asked for it (A returns to hand, B's source).
- `TestFR_REP_001_KillStepAsksForEachKill` in `replacement_kill_test.go` (`SVY-R21`, `SVY-R23`): two guardian structures and one kill step that kills two allied units. Two `DecideReplacement` questions are asked of player 0, one per kill, and both kills are replaced.
- **Existing tests:** every other test in `pkg/engine` passes unchanged. None puts two applicable replacements on one death (directive #75 checked `replacement_test.go` and `replacement_kill_test.go`).

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected. The test files may declare their own helpers (for example, the replacement questions asked of a player, read from the decision events).

## Allowed Files
- `pkg/engine/replace.go` (`replacementOption`, `replacementsFor` in place of `findReplacement`, `replacedQueue`, `replaceDeath`, `chooseReplacement`)
- `pkg/engine/choice.go` (`DecideReplacement`, its `String`, its fallback case)
- `pkg/engine/doc.go` (built list, "Not built yet", Files list)

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run 'TestFR_REP_001'`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-compare -- autoloop/m2-wave-g` passes (no executable card prints a replacement yet, so the sims are unchanged; see Shape, Sims)

## Log
- 2026-10-29 lead-architect (wave G run, session 5): Wrote the task for board directive #75 (`SVY-R23`, the user's ruling on design note Open Question 14 and survey R23, which stopped session 4). Design note: Deliverable 6c's T-0100 line no longer waits on a ruling, and Open Question 14 is rewritten as answered (this commit).
  - **Sync:** merged `main` (f8f9dbb, directive #75 and its outcome) into the run branch with no conflict (f0ba976).
  - **Directives:** read `BOARD.md` on `autoloop/m2-wave-g` fresh after the sync. #75 is applied by this task. #73 stays `open: before the first wave C task` (T-0098, T-0099 still to come); its status cell now names #75. #14 is a standing order. Waiting triggers checked: #6 and #7 (mode is `serial`); #8 (no `deck.go` change); #15 (no turn hook); #20 (b) (no wave C shape task); #58 (no exhaust by a cost or an effect); #59 (2) (no repeated targets); #61 (no modal card); #71 (no effect on each player). None has fired.
  - **Tests beyond the directive's five:** `TestFR_REP_001_ReplacementChoiceTimesOutToPrintedOrderWithinASource`, because `SVY-R23`'s timeout names a source's own replacements in printed order.
  - **Reuse:** no duplicate. `replacementsFor` takes `findReplacement`'s place rather than sitting beside it, and the timeout is `fixedOrderPicks` (the `DecideTarget`/`DecidePick` fallback), not a new order.
