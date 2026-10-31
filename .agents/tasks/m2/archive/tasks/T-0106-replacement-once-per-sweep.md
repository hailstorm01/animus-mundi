---
id: T-0106
title: A destruction replacement saves an asset at most once per sweep, and the next pass takes it up again (FR-REP-003, SVY-R24, directive #81, wave G.1 shape, sonnet)
kind: shape
tier: sonnet
wave: G.1
status: done
branch: m2/T-0106-replacement-once-per-sweep
depends_on: [T-0100, T-0105]
claims: [rule:replacement-once-per-window, helper:used-replacements]
cards: []
requirements: [FR-REP-003, FR-REP-001, FR-LAY-002]
---

## Goal
Within one death sweep, a destruction replacement applies at most once to a given asset, and every pass of the sweep takes up every asset that is still lethal. An asset that a replacement leaves on the battlefield and still lethal is therefore taken up again in the next pass, where a different replacement applies or it is destroyed. Each kill step is its own window. This applies board directive #81 (`SVY-R24`, design note Open Question 15).

## Inputs
- FRD 01 v3.18: `FR-REP-003` "Single Application per Discrete Event" (line 463); `FR-REP-001` "Replacement Within a Death Sweep" (line 469, its last sentences and last *Example*); "Same-Moment Deaths" (line 468, `SVY-R21`); `FR-LAY-002`.
- Survey rows: none (wave G.1 has no survey row and no cards). Survey §4, row R24.
- Rulings that apply: `SVY-R24` (this task), `SVY-R20`, `SVY-R21`, `SVY-R22`, `SVY-R23` (unchanged).
- Outcome: `docs/design/outcomes/2026-10-29-replacement-once-per-sweep.md`.
- Design note: `docs/architecture/m2-ability-model.md`, Deliverable 6d; Open Question 15 (closed in this task's commit).
- Board directive #81.
- Kernel today (`sweep.go`, `replace.go`, `effects.go`):
  - `sweepDeaths` takes up `newlyLethal(seen)` in passes: an asset is taken up at most once per sweep, so one a replacement leaves on the battlefield and lethal survives the sweep (and `CheckInvariants` reports "lethal but survived the death sweep"), and the same replacement applies to it again at every later sweep.
  - `replaceDeaths` (sweep and `stepKill`) handles each asset of `dying` once; `replacementsFor` knows nothing of what has already applied.

### Shape (lead decisions, binding)
- **One record of what has applied, per window.** A set of the replacements that have applied in this window, each keyed by the asset it applied to, its source asset's ID and the `*ability.Replacement`. The source ID is part of the key because two copies of one card share the same printed `Replacement`, and a source may print several. Name the type as reads best (for example `usedReplacements`), and declare it in `replace.go`.
- **`replace.go`:**
  - `replacementsFor(a, used)` leaves out every option already used on `a` in this window. Its doc comment says so (`FR-REP-003`, `SVY-R24`).
  - `replacedQueue`, `replaceDeaths` and `replaceDeath` take the set and pass it down. `replaceDeath` adds the option it applies. Nothing else changes: the order (`SVY-R20`, `SVY-R22`), the `SVY-R23` prompt only when two or more options remain, and the timeout.
- **`sweep.go`:**
  - `deathSweep` makes one set for the whole sweep, in place of `seen`.
  - `sweepDeaths` takes the set. Each pass takes up every asset that is lethal now. `newlyLethal` goes; a function that returns every lethal asset in `g.Assets()` order replaces it (or inline it, as reads best). The loop ends when a pass finds no lethal asset. It still ends, because every pass either destroys an asset or uses up a replacement for one.
  - Rewrite `sweepDeaths`'s comment to match `FR-REP-001`'s new wording: each pass takes up every asset still lethal, a new moment; a replacement already used on an asset in this sweep does not apply to it again (`SVY-R24`).
- **`effects.go`:** `stepKill` passes a new, empty set: each kill step is its own window. Nothing else changes there.
- **`doc.go`:** the "Destruction replacement" bullet adds that a replacement applies at most once to a given asset per sweep or kill step (`SVY-R24`).
- **`CheckInvariants` stays as it is.**
- **Sims:** no printed card reaches the new behavior. *Shifting Impossibility*, the only printed destruction replacement, always moves the asset to its owner's hand, so it is never lethal on the battlefield in a later pass. So `go tool task sim-compare -- autoloop/m2-wave-g.1` should pass byte for byte. If it differs, don't change code to make it match: report the first differing game in the Log and hand off.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
Synthetic cards and test-only replacements only, in a new file `pkg/engine/replacement_once_test.go`. Reuse the helpers in `replacement_test.go` (`alliedUnits`, `(s *scene) sweep`, `eventSeq`, `replaceTargets`, `replaceSources`, `replacementAsks`) and `scenario_test.go` where they fit. A "damage-instead" replacement below is a `WouldBeDestroyed` replacement on an allied-unit filter (`Other: true`) whose only step is `OpDealDamage` 1 to `SubjectSelf`, printed by a structure with enough Health to take every point the test deals it. Every existing test passes unchanged.
- `TestFR_REP_003_ReplacementAppliesOncePerSweep` (directive #81 (a)): one damage-instead source and an allied unit at 0 Current Health at a sweep. The unit's death is replaced once (one `replace` event naming it), the unit is destroyed in the same sweep's next pass, its source took exactly 1 damage, and `CheckInvariants` passes after the sweep.
- `TestFR_REP_003_NextPassAppliesAnotherReplacement` (directive #81 (b), `SVY-R23`): two damage-instead sources and one lethal allied unit. In the first pass both apply, so its controller is asked once (`DecideReplacement`) and picks one; in the second pass only the other applies, and it applies with no second question; the third pass destroys the unit. Each source took exactly 1 damage, and there are two `replace` events naming the unit, from different sources.
- `TestFR_REP_003_ReplacementThatHealsSaves` (directive #81 (c)): a replacement whose steps include `OpHeal` on `SubjectReplaced`, large enough to clear the unit's damage. At a sweep where the unit is lethal, its death is replaced, the unit stays on the battlefield, is not lethal, and is not destroyed; `CheckInvariants` passes.
- `TestFR_REP_003_KillThenSweepAppliesAgain` (directive #81 (d), each kill step its own window): one stack item (a test ability) first damages an allied unit to 0 Current Health, then kills it with `OpKill`. The damage-instead replacement applies to the kill. At the sweep after the item it applies to the unit once more, and the sweep's next pass destroys the unit. Its source took exactly 2 damage, and there are two `replace` events naming the unit.
- `TestFR_REP_003_ReplacementSavesFromAKill` (the outcome's "does save a unit from a kill effect"): a unit with Health left is killed by an `OpKill` step, and the damage-instead replacement applies. After the item and its sweep, the unit is on the battlefield with its damage unchanged, and its source took exactly 1 damage.

Directive #81's "the sweep still ends" is covered by (a) and (b) finishing. *Shifting Impossibility*'s and every `TestFR_REP_001_*` test pass unchanged; `sim-compare` is checked by the gates.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected. The test file may declare its own helpers.

## Allowed Files
- `pkg/engine/replace.go` (the used set; `replacementsFor`, `replacedQueue`, `replaceDeaths`, `replaceDeath` take it)
- `pkg/engine/sweep.go` (`deathSweep`, `sweepDeaths`; `newlyLethal` goes)
- `pkg/engine/effects.go` (`stepKill` passes a new set; nothing else)
- `pkg/engine/doc.go` (the destruction-replacement bullet only)

## Done When
- [x] The listed tests pass: `go test ./pkg/engine -run 'TestFR_REP_|TestCard_ShiftingImpossibility|TestFR_CBT_002'`
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks
- [x] `go tool task sim-compare -- autoloop/m2-wave-g.1` passes (no printed card reaches the new behavior; see Shape, Sims)
- [x] No non-test file in `pkg/engine` still declares `newlyLethal` or a `seen` set in the sweep.

## Log
- 2026-10-29 lead-architect (wave G.1 run, session 5): Wrote the task for board directive #81 (`open: before the next Deliverables task`; #76–#79 applied, so #81 goes before the wave closes). `main` (556811f) has nothing the run branch lacks, so no sync.
  - **Directives:** read `BOARD.md` on `autoloop/m2-wave-g.1` fresh. #81 is applied by this task. #14 is a standing order. Waiting triggers checked against this task's claims and the repository: #6 and #7 (mode is `serial`); #8 (no `deck.go` change); #15 (no turn hook); #20 (b) (no wave C shape task); #58 (no exhaust by a cost or an effect); #59 (2) (no repeated targets); #61 (no modal card); #71 (no effect on each player). None has fired.
  - **Reuse:** no existing record of applied replacements exists (`grep` of `pkg/engine` for `replacementsFor`, `seen`, `used`). The set replaces `seen`, so the sweep keeps one per-sweep record, not two.
  - **Tests beyond the directive:** `TestFR_REP_003_ReplacementSavesFromAKill`, the other half of the outcome's example, so that the kill window is pinned both when the asset is lethal afterwards and when it is not.
  - **Design note:** Open Question 15 marked answered in this commit.
- 2026-10-29 rules-test-author: Wrote `pkg/engine/replacement_once_test.go`: `TestFR_REP_003_ReplacementAppliesOncePerSweep`, `_NextPassAppliesAnotherReplacement`, `_ReplacementThatHealsSaves`, `_KillThenSweepAppliesAgain`, `_ReplacementSavesFromAKill`. Helpers requested: none (the file declares `damageInsteadReplacement`, `damageInsteadStructure`, `wantInvariants`). Cases left out: none.
  - **Fail before the implementation:** `AppliesOncePerSweep`, `NextPassAppliesAnotherReplacement` and `KillThenSweepAppliesAgain` (the unit survives lethal, `CheckInvariants` reports "lethal but survived the death sweep", and the second replace event never happens, because the kernel takes an asset up once per sweep). `ReplacementThatHealsSaves` and `ReplacementSavesFromAKill` pass already: they pin behavior that must stay.
  - `go vet` is clean; `cardtown lint rules` passes.
- 2026-10-29 lead-architect (tests review, round 1): Approved, `ready` for `kernel-engineer` (sonnet). Checked each test against `FR-REP-003` (line 463), `FR-REP-001` (line 469 and its last *Example*) and `SVY-R23`: (a), (b) and (d) fail today for the reason the ruling changes (the sweep takes an asset up only once); (c) and the kill-save test pin behavior that must stay. `wantReplaced` counts every `replace` event, so (a) also pins "no second application in one sweep". Engineer: keep the replace order and the `SVY-R23` prompt exactly as they are; only the used set is new.
- 2026-10-29 kernel-engineer: Implemented the shape. Reuse: no existing record of applied replacements; the new `usedReplacements` set (key: asset ID, source ID, `*ability.Replacement`) replaces the sweep's `seen` map.
  - `replace.go`: `usedKey`/`usedReplacements`; `replacementsFor`, `replacedQueue`, `replaceDeaths`, `replaceDeath` take the set; `replaceDeath` records the option it applies. Order, SVY-R23 prompt and timeout untouched.
  - `sweep.go`: `deathSweep` makes one set; `sweepDeaths` takes it and each pass takes up `lethalAssets()` (every lethal asset, `Assets()` order); `newlyLethal` and `seen` removed; comment rewritten. `effects.go`: `stepKill` passes `usedReplacements{}`. `doc.go`: bullet extended.
  - Ran: `go test ./pkg/engine -run 'TestFR_REP_|TestCard_ShiftingImpossibility|TestFR_CBT_002'` ok; `go tool task check` passed; `go tool task sim` clean; `go tool task sim-compare -- autoloop/m2-wave-g.1` PASS byte for byte; grep of sweep.go finds no `newlyLethal`/`seen`.
- 2026-10-29 lead-architect (code review, round 1): Approved, merged into `autoloop/m2-wave-g.1` as aace7a6 (the merge tree is the gated task tip 5fd363b). Re-run in a review worktree at 5fd363b: `go tool task check` passes; `go tool task sim` clean, 200 games, P0 67.0% / P1 32.5% / draws 0.5% (unchanged); `go tool task sim-compare -- autoloop/m2-wave-g.1` byte-identical. Scope: only Allowed Files and `replacement_once_test.go` changed, and no test changed after 7a5b28d. Fences pass. Reuse: the used set replaces `seen`, so the sweep keeps one per-sweep record, and `stepKill` gets its own window. Rules fidelity: `replacementsFor` leaves out a replacement already used on the asset for every caller, so the queue, the `SVY-R23` prompt and the apply step all see the same options; each pass takes up every lethal asset, and every pass either destroys one or uses up a replacement, so the sweep ends. Directive #81 `applied`.
