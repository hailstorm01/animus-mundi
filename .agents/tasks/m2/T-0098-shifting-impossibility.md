---
id: T-0098
title: A Scratch Company filter, and Shifting Impossibility (FR-TGT-001, FR-REP-001, wave G shape, sonnet)
kind: shape
tier: sonnet
wave: G
status: done
branch: m2/T-0098-shifting-impossibility
depends_on: [T-0100]
claims: [filter:scratch, card:Shifting Impossibility]
cards: [Shifting Impossibility]
requirements: [FR-TGT-001, FR-REP-001]
---

## Goal
A filter can ask for Scratch Company units only ("an allied Scratch Company unit"), the counterpart of today's `NonScratch`. *Shifting Impossibility* runs: "If an allied Scratch Company unit would be killed, return it to its owner's hand and deal 1 damage to this unit instead." It comes off `NotYetExecutable`, so the Animus starter deck plays it.

## Inputs
- FRD 01, `FR-TGT-001` (line 330: subtype criteria in a filter, "non-Scratch" exclusions). `FR-REP-001` to `FR-REP-003` (lines 454-464): Inline Event Substitution names *Shifting Impossibility*; Consequential Event Isolation (its 1 damage is a separate child event); Replaced Actions Still Count as Done (`SVY-R15`, the *Infested Apostle* example); Per-Entity Death Replacement; Same-Moment Deaths (`SVY-R21`); Replacement Within a Death Sweep (`SVY-R20`, `SVY-R22`). `FR-ZON-005` (lines 212-216): a face-down Scratch Company returned to its owner's hand. `SVY-R23` (choosing between replacements) is the user's ruling in board directive #75, not yet in FRD 01.
- Survey: wave G (line 35), *Shifting Impossibility* (line 139): `WOULD_BE_DESTROYED` on allied Scratch Company → `RETURN_TO_HAND` + `DEAL_DAMAGE 1` self. Rulings `R5` (revised by `R20`), `R15`, `R21`, `R23`.
- Card: `data/cards/units-and-structures-animus.csv` line 4 (Unit, Horror, 3 power, 4 Health, deploys to Reserves). The Animus starter deck runs two.
- Design note: `docs/architecture/m2-ability-model.md`, Deliverable 6c (T-0098).
- Kernel today:
  - `ability.Filter.NonScratch` excludes Scratch Company units. `matchesPrinted` (`targeting.go`) reads it against the printed characteristics (`c.HasSubtype("Scratch Company")`), and every filter check (targets, picks, "all" steps, pile cards, replacement filters) goes through `matchesPrinted`. A face-down asset's printed characteristics are the Scratch Company face (`Asset.Printed`, `scratchFace`).
  - Destruction replacement is built (T-0096, T-0097, T-0101, T-0100): `ability.Replacement{Event: WouldBeDestroyed, Filter, Steps}`, `SubjectReplaced`, `SubjectSelf`, `DecideReplacement`.

### Shape (lead decisions, binding)
- **The filter (`ability/ability.go`):** add `Scratch bool` to `Filter`, right after `NonScratch`, with the comment `// only Scratch Company units`. `NonScratch` is unchanged.
- **One Scratch Company check (`targeting.go`):** `matchesPrinted` reads the subtype once and tests both flags, so what counts as a Scratch Company stays written in one place:
  ```go
  scratch := c.HasSubtype("Scratch Company")
  return !(f.NonScratch && scratch) && !(f.Scratch && !scratch)
  ```
  The doc comments of `matchesCard`, `matchesPrinted` and `matchesFilter` that say "NonScratch" say "Scratch and NonScratch".
- **The card (`cards/animus.go`):** `animusDefs["Shifting Impossibility"]`, in alphabetical place, with `Text` the exact printed line and one `Replacement{Event: ability.WouldBeDestroyed, Filter: ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit, Scratch: true}, Steps: [{Op: OpReturnToHand, Subject: SubjectReplaced}, {Op: OpDealDamage, Amount: Value{N: 1}, Subject: SubjectSelf}]}`. No `Other`: *Shifting Impossibility* is a Horror, never a Scratch Company while face-up, and face-down it prints nothing.
- **`cards/cards.go`:** remove "Shifting Impossibility" from `NotYetExecutable`.
- **Unchanged:** `replace.go`, `damage.go`, `effects.go`, `choice.go`, `doc.go` (its built list already names destruction replacement, T-0100), and the clients.
- **Sims:** *Shifting Impossibility* is a deck card, so the starter sim changes (new replacements, new `DecideReplacement` questions on the random chooser's stream). `go tool task sim` must run clean; record its result line. No `sim-compare`: this is not a pure move.
- **Reuse:** searched `pkg/engine` and the `done` tasks for a positive Scratch Company filter and found none. `NonScratch` is its negation, and both read one subtype check. A general subtype criterion (`FR-TGT-001`'s *Infantry*, *Fungus*) is not built: the only other subtype filter among Animus and Path cards is *Melodic Pummeler*'s "Allied Horror units", a non-deck card whose same-zone aura waits under #14 (b). *Animoceps Spores* ("return an allied Scratch Company unit") will reuse `Scratch`.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
Filter tests use synthetic cards. Card tests use the real *Shifting Impossibility* (`s.card`), face-down Scratch Companies (`s.putScratch`), and the replacement helpers in `replacement_test.go` (`sweep`, `wantReplaced`, `replacementAsks`, `replaceSources`, `wantOffered`); `alliedUnits` with `Scratch: true` (or a filter of its own) keeps a sweeper off *Shifting Impossibility*. No existing test is edited.

- `TestFR_TGT_001_ScratchFilterTargetsOnlyScratchCompanies` in `targeting_test.go`: a chosen-target clause with `Scratch: true` offers face-down Scratch Companies (both players', with no allegiance set) and no face-up unit.
- `TestFR_TGT_002_FilteredAllScratchHitsOnlyScratchCompanies` in `targeting_test.go`: an "all" damage step with `Kinds: KindUnit, Scratch: true` damages every Scratch Company and no other unit.
- `TestCard_ShiftingImpossibility_ReturnsAScratchCompanyKilledAtTheSweep` in `cards_animus_impossibility_test.go` (`FR-REP-001`, `FR-ZON-005`): an allied face-down Scratch Company lethal at a sweep goes to its owner's hand as its printed card, not to the discard pile; *Shifting Impossibility* has taken 1 damage; no `killed` event for the Scratch Company.
- `TestCard_ShiftingImpossibility_ReplacesOnlyAlliedScratchCompanies` in the same file: an allied face-up unit and an opposing Scratch Company, both lethal at one sweep, are destroyed; *Shifting Impossibility* takes no damage.
- `TestCard_ShiftingImpossibility_ReplacedKillStillCountsAsDone` in the same file (`SVY-R15`, FRD 01's example): *Infested Apostle* deploys, its controller says yes and picks an allied Scratch Company for the kill. The Scratch Company returns to hand, *Shifting Impossibility* takes 1 damage, and the Apostle's 2 damage to the chosen opposing unit still happens.
- `TestCard_ShiftingImpossibility_ItsOwnLethalDamageStillReplacesEveryDeathInThePass` in the same file (`SVY-R20`): *Shifting Impossibility* at 1 Current Health and two allied Scratch Companies lethal at one sweep. Both return to hand, and *Shifting Impossibility* is destroyed after them (in its owner's discard pile when the sweep ends).
- `TestCard_ShiftingImpossibility_TwoCopiesAskWhichOneReplaces` in the same file (`SVY-R23`): two *Shifting Impossibility* and one allied Scratch Company lethal at a sweep. Player 0 is asked one `"Replacement"` question offering both, picks the newer one, and only that one takes 1 damage.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected. The new test file may declare its own helpers.

## Allowed Files
- `pkg/engine/ability/ability.go` (`Filter.Scratch`)
- `pkg/engine/targeting.go` (`matchesPrinted` and the three doc comments)
- `pkg/engine/cards/animus.go` (*Shifting Impossibility*)
- `pkg/engine/cards/cards.go` (`NotYetExecutable`)

## Done When
- [x] The listed tests pass: `go test ./pkg/engine -run 'ScratchFilterTargetsOnly|FilteredAllScratchHitsOnly|TestCard_ShiftingImpossibility'`
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks
- [x] The card loads with no unsupported lines (`cardtown lint --all`, in `go tool task check`)

## Log
- 2026-10-29 lead-architect (wave G run, session 6): Wrote the task, next in design note Deliverable 6c after T-0100 (done, 4adc393). Planned by T-0096 as `Filter.Scratch` plus the card that first uses it.
  - **Sync:** `main` has nothing the run branch lacks; no merge.
  - **Directives:** read `BOARD.md` on `autoloop/m2-wave-g` fresh. #73 stays `open: before the first wave C task`; this is wave G, and T-0099 applies #73. #14 is a standing order, and *Shifting Impossibility* is a deck card (#14 (a)). Waiting triggers checked: #6 and #7 (mode is `serial`); #8 (no `deck.go` change); #15 (no turn hook); #20 (b) (no wave C shape task); #58 (no exhaust by a cost or an effect); #59 (2) (no repeated targets); #61 (no modal card); #71 (no effect on each player). None has fired.
  - **Reuse and placement:** see Shape, Reuse. Every file sits where the Boundaries table puts it (`ability` for shapes, `engine` for filter checks, `cards` for definitions).
  - **Deck cards still on `NotYetExecutable` (#14 (e)):** 23 (10 Animus, 13 Path) before this task; 22 (9 Animus, 13 Path) once it merges.
- 2026-10-29 rules-test-author: Tests written, status tests-review.
  - `targeting_test.go`: `TestFR_TGT_001_ScratchFilterTargetsOnlyScratchCompanies`, `TestFR_TGT_002_FilteredAllScratchHitsOnlyScratchCompanies`.
  - `cards_animus_impossibility_test.go`: `TestCard_ShiftingImpossibility_ReturnsAScratchCompanyKilledAtTheSweep`, `_ReplacesOnlyAlliedScratchCompanies`, `_ReplacedKillStillCountsAsDone`, `_ItsOwnLethalDamageStillReplacesEveryDeathInThePass`, `_TwoCopiesAskWhichOneReplaces`. The file declares its own helpers `alliedScratch` and `noEvent`.
  - Helpers requested: none. `go vet ./pkg/engine` fails only on `ability.Filter.Scratch`, the field this task adds. `cardtown lint rules` passes.
  - Notes: in `ReplacesOnlyAlliedScratchCompanies` a face-up allied unit cannot be swept without also sweeping *Shifting Impossibility* (it is an allied non-Scratch unit), so the sweep filter hits it for 1 and the test expects exactly 1 damage on it and no `replace` events. The "no `killed` event" check also checks `destroy`, the only death-event kind other tests use. *Shifting Impossibility* "at 1 Current Health" is set as `Damage = 3` of 4 Health. No cases left out.
- 2026-10-29 lead-architect (wave G run, session 6): Tests approved (D4). Each matches the FRD 01 line it cites: the Scratch filter offers and hits only Scratch Companies of either side; an allied Scratch Company's sweep death and an *Infested Apostle* kill are replaced (hand, printed card, 1 damage to the source, no `destroy`), and the Apostle's "if you do" damage still lands (`SVY-R15`); a face-up allied unit and an opposing Scratch Company die normally; at 1 Current Health it replaces both deaths of the pass and is destroyed after the last `leave` (`SVY-R20`); two copies ask one `"Replacement"` question offering both (`SVY-R23`). The sweep in `ReplacesOnlyAlliedScratchCompanies` hitting *Shifting Impossibility* for exactly 1 is accepted. Status `ready`; dispatching `kernel-engineer` (sonnet).
- 2026-10-29 kernel-engineer: Implemented, status review.
  - **Changed:** `ability.Filter.Scratch` (after `NonScratch`); `matchesPrinted` reads the Scratch Company subtype once and tests both flags, three doc comments say "Scratch and NonScratch"; `animusDefs["Shifting Impossibility"]` (one `Replacement`, steps `OpReturnToHand` on `SubjectReplaced`, `OpDealDamage` 1 on `SubjectSelf`); removed it from `NotYetExecutable`.
  - **Reused:** the existing `Replacement` shape and `NonScratch` check; no new shapes. No helpers added to `scenario_test.go`; no test edited.
  - **Commands:** `go test ./pkg/engine -run 'ScratchFilterTargetsOnly|FilteredAllScratchHitsOnly|TestCard_ShiftingImpossibility'` ok; `go tool task check` passed; `go tool task sim`: Animus 32/50 executable (was 31 before this card); 200 games, P0 wins 67.0%, P1 32.5%, draws 0.5%, avg 26.8 turns, "[PASS] Every match terminated with all invariants holding." No `sim-compare` (not a pure move).
- 2026-10-29 lead-architect (wave G run, session 6): Code review (D5) passed, no changes.
  - **Scope:** only the four Allowed Files and the task file changed after the test commit 7f2624d; no test edited (checked with `git diff`).
  - **Fences and reuse:** the card name appears only in `cards`. `Scratch` sits beside `NonScratch` and both read one `HasSubtype("Scratch Company")` in `matchesPrinted`, so what counts as a Scratch Company is written once. The card reuses the `Replacement` shape, `SubjectReplaced` and `SubjectSelf`.
  - **Rules fidelity:** general. Every filter check (targets, picks, "all" steps, replacement filters) goes through `matchesPrinted`, and a face-down asset's printed face is the Scratch Company face.
  - **Gates re-run in a review worktree:** the listed tests ok; `go tool task check` passes; `go tool task sim` clean, 200 games, P0 67.0% / P1 32.5% / draws 0.5% (was 63.5% / 34.5% / 2.0%; the Animus deck now plays the card), Animus 32/50 main cards executable.
  - Merged into `autoloop/m2-wave-g` with `--no-ff` as 3215a5b. Its tree is the gated task tip. Status `done`. Deck cards still on `NotYetExecutable` (#14 (e)): 22 (9 Animus, 13 Path). Next: T-0099 (#73), design note Deliverable 6c.
