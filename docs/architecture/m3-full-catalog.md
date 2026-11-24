# M3: Full Catalog

- **Milestone:** M3 | **Source spec:** [FRD 01](../frd/01_Game_Rules_FRD.md) v3.29 (`FR-ENT-102`, `FR-KWD-002n`, `FR-TRG-006`, `FR-KWD-002p` still to pin; `FR-CBT-003`), [FRD 02](../frd/02_Card_Data_Pipeline_FRD.md) v2.0 (`FR-DAT-*`, `FR-ABL-*`), build order in [`PRODUCT_SUMMARY.md`](../PRODUCT_SUMMARY.md) §4 M3, inputs from the pm-lead review of 2026-11-23 | **Date:** 2026-11-23, revised 2026-11-23 (the catalog export moved to M5)

M3 builds on the [M2 design note](m2-ability-model.md): its Boundaries, key types, task kinds and test conventions all still hold, and this note gives only what M3 changes. The ability vocabulary is [`go doc ./pkg/engine/ability`](../design/m2_ability_shapes.md), which FRD 02 v2.0 specifies.

**The catalog export is not M3 work.** The stakeholder moved it to M5 on 2026-11-23, because nothing reads it before the M5 session checksum handshake. [FRD 03](../frd/03_Card_Compiler_FRD.md) is frozen as design intent until then, and this note designs nothing from it. In this note "the catalog" means the card set: the sheets in `data/cards/` and the definitions in `pkg/engine/cards`. It does not mean a generated file.

## Goal

Every card in `data/cards/` without a `needsRework` flag executes in the kernel as a Go value in `pkg/engine/cards`. We know it works when the M3 Done-when in `PRODUCT_SUMMARY.md` §4 holds:

- `cardtown lint rules` reports 62 of 62 active FRD 01 requirements pinned.
- `cards.NotYetExecutable` is empty with the coverage test widened to the whole catalog (169 cards, none flagged today).
- No reachable matrix item or survey ruling is open.
- The Concordia and Soulless starter decks validate, and the starter sim matrix (Deliverable 3) runs clean on every pairing of the four starter decks.
- A human finishes a Concordia vs Soulless `cardtown play` match.

FRD 02 also adds three sheet checks that the export's pre-flight used to hold: a duplicate card name (`FR-DAT-002`), and a non-numeric value in a numeric column or a Unit or Structure with no zone (`FR-DAT-006`). Deliverable 1 builds them.

## Boundaries

The M2 package table stands. M3 changes:

| Package | M3 change | Card names? | Changed by |
| :--- | :--- | :--- | :--- |
| `pkg/engine/ability` | New shapes only as their first card needs them (`FR-ABL-013`), each in a `shape` task that updates FRD 02 in step. Known ones: the Overkill hook with its excess value (`FR-KWD-002p`, M2 Open Question 12), a granted quoted triggered ability (*Redline Infusion*), delayed triggers (`FR-TRG-006`), tucking and the Attached sub-zone (`FR-ENT-102`, #142), Regroup (`FR-KWD-002n`), and the waiting M2 directives (#15 (2), #71, #91 (2), #106, #120). A small reader for Deliverable 6a: whether a `Filter` can match a Headquarters. | No (fenced) | `shape` tasks |
| `pkg/engine/cards` | `concordia.go`, `soulless.go`, and *Rift* and *Encampment* in `headquarters.go`. `NotYetExecutable` lists every unflagged card that can't run, not only starter-set cards. Deliverable 6a adds an allowlist of cards allowed to kill a Headquarters. | Yes (only place) | `card` tasks |
| `pkg/engine` | The interpreter, as in M2. The sheet loader reports the three data errors above (Decisions). | No (fenced) | `shape` and `harness` tasks |
| `pkg/sim` | The starter sim matrix: every pairing of the starter decks, both seat orders, both policies, with one combined coverage table (Deliverable 3). | No (fenced) | `harness` tasks |
| `pkg/cardcheck` (**proposed**, Open Question 2) | Checks over the whole card set that are not rules: the executable coverage checker and the definition-to-text checker (moved from `defs_test.go`, Deliverable 1), and the Headquarters-kill check (Deliverable 6a). It imports `engine`, `cards` and `ability`, and nothing imports it. | No (fenced) | `harness` tasks |
| `cmd/cardtown` | `cardtown sim -matrix` (Deliverable 3). | — | `harness` tasks |

## Decisions

- **The sheet loader reports the FRD 02 data errors, not `cardtown lint sheets`.** This answers FRD 02's open question for `FR-DAT-002` and `FR-DAT-006`. Every game, sim, play and test loads through `LoadCards`, so a bad cell can't reach a match, and `TestDefs_MatchRulesText` loads the real sheets in every `go tool task check`. The sheet linter reads CSVs as text and imports no game package. A second copy of the checks there would be the same rule in two places (D2.0). `LoadCards` returns one error that lists every problem it found, in file then row order. Each problem names the file, row, card and column, and a duplicate name names both rows. An empty cell in an asset sheet's numeric column counts as non-numeric, because a card is never loaded with a guessed value. Today's sheets have none of the three errors (checked 2026-11-23). No decision record: nothing about this is hard to reverse.
- [ADR-0009](decisions/0009-cards-identified-by-printed-name.md): a card is identified by its printed name through M4, and `FR-DAT-002` cites it. Its catalog-format parts (a list of entries sorted by name, the checksum covering the name) are direction for the M5 export, not M3 work. The card rename policy is a required item at M5 start (pm-lead, 2026-11-23).
- **Moved to M5 with the export:** every FRD 03 technical question (serialization and enum spellings, canonical form and digest, file name, location and whether it is committed, the command, a published schema, the FRD file rename). They are decided when the export is built, with the handshake as their first reader.
- ADR-0002, 0003, 0004, 0007 and 0008 still hold.

## Test Plan

- **Rules tests:** `TestFR_ENT_102_*`, `TestFR_KWD_002n_*`, `TestFR_TRG_006_*`, and the Overkill tests already named in M2 Open Question 12 (`TestFR_KWD_002p_*`). With them `cardtown lint rules` reaches 62 of 62.
- **Card tests:** at least one `TestCard_<CardSlug>_<Behavior>` per card, written from the CSV text, as in M2.
- **Coverage:** `TestStarterSet_Executable` becomes `TestCards_Executable`, a catalog-wide check: every unflagged card that can't run is on `NotYetExecutable`, and nothing else is. Its checkers move out of `defs_test.go` into a non-test file in `pkg/cardcheck` in the same task (the "test files hold only tests" rule, which `defs_test.go` predates), with `TestDefs_MatchRulesText` and its checker beside them, unchanged in behavior. A new test validates every `data/decks/*-starter.txt` under `FR-SET-001`.
- **Sheet checks:** loader tests on fixture sheets for each data error: a duplicate name across two files, a non-numeric and an empty numeric cell, a Unit and a Structure with no zone, and several problems in one load, reported together in file then row order.
- **Headquarters kills:** `TestCards_HeadquartersKills` (Deliverable 6a).
- **Sim:** `go tool task sim` stays the Animus vs Path starter run that every review runs (Runbook D5), and `go tool task sim-compare` is unchanged. The matrix (Deliverable 3) is the Done-when's "every pairing" check, run in the review of every card task once the Concordia and Soulless decks are committed.
- **Monitor coverage, tracked, not a gate:** the matrix prints which monitor rule IDs no run reached. Measured 2026-11-23 (200 games, `-seed 1`):

  | Rule | Starter, progress | Starter, random | Stress A v P, progress | Why |
  | :--- | :---: | :---: | :---: | :--- |
  | `FR-KWD-002g` Ethereal | not reached | not reached | 292 | Printed only by *Flickering Memory* and *Drifting Concept*, neither in a starter deck |
  | `FR-KWD-002j` Guardian | not reached | not reached | 165 | Printed only by *Ego Spark* and *Improvised Bastion*, neither in a starter deck |
  | `FR-KWD-002h` Ward | not reached | not reached | not reached | Printed only by Soulless cards (granted by *Comms Relay*), none executable yet |
  | `KIM-S03` zero yield | not reached | 136 | not reached | `ProgressPolicy` never generates a zero yield. The random policy does |

  So seeds and seat swaps alone reach none of the four. The policy axis reaches `KIM-S03`, and the keywords need decks that hold the cards. The matrix therefore also runs the stress decks (every executable faction card, `sim.FactionDeck`) over the same pairings. No starter deck variants are built: they would be a second set of decklists to keep in step with the stakeholder's, and the stress decks already hold every card. Ward is expected once Soulless cards execute. If a rule is still unreached when the playtest checkpoint is reached, or at M3's end, the lead reports it to the user rather than adding a variant.

## Deliverables (in order)

1. **M3 scaffold (harness).** The coverage test covers the whole catalog, and `NotYetExecutable` holds the 67 cards that can't run (each a name, sorted). The coverage checkers move to `pkg/cardcheck` (Open Question 2). `concordia.go` and `soulless.go` exist, empty. `LoadCards` reports a duplicate name (`FR-DAT-002`), a non-numeric or empty numeric cell and a Unit or Structure with no zone (`FR-DAT-006`), as Decisions gives. The deck test validates every starter deck. Done when `go tool task check` passes, `NotYetExecutable` has 67 entries, the sheet-check tests pass, and `go tool task sim-compare` passes. This task can be split: the sheet checks don't depend on the coverage work.
2. **Concordia and Soulless starter decks** (the stakeholder with `game-designer`, build order step 3). Done when both files are committed and Deliverable 1's deck test passes on them. The lead doesn't write decklists.
3. **Starter sim matrix (harness).** `cardtown sim -deck starter -matrix` plays every pairing of the factions that have a `<faction>-starter.txt`, in both seat orders, with the `progress` and `random` policies, 200 games each from the same seed, in a fixed order (factions sorted, then seat order, then policy). Without `-deck` it does the same with the stress decks. It prints one summary line per run and one coverage table summed over all runs (`Coverage.add`), with how many runs reached each rule ID. It exits non-zero on any invariant or monitor violation, naming the run and its replay command, and never on an unreached rule. A `go tool task sim-matrix` entry runs both deck sets. With four starter decks that is 24 starter runs, about two minutes at today's speed. Done when its acceptance cases pass and the two-deck matrix runs clean today. It can be built before Deliverable 2, because it finds the decks by file.
4. **M3 ability survey** (`docs/design/m3_<topic>.md`, build order step 4; owner in Open Question 1). Every one of the 67 cards mapped onto the vocabulary, its gaps and ruling questions, and the build waves, deck cards first. The *Dread Manifestation* question (Open Question 4) is asked here, because the card is in the Concordia draft.
5. **Deck-card waves**, then the playtest checkpoint: the stakeholder plays Concordia vs Soulless in `cardtown play`. Each wave is done when its survey rows pass, the matrix runs clean, and `go tool task check` passes.
6. **Non-deck cards, and the *Rift* and *Encampment* Headquarters**, wave by wave. Done when `NotYetExecutable` is empty and the four unpinned FRD 01 requirements are pinned.
6a. **Headquarters-kill check (harness), after `game-designer`'s ruling on a Headquarters killed outright** (Open Question 3). It is recorded on the M3 board at D1 as a `waiting:` directive whose trigger is checkable: FRD 01 `FR-CBT-003` no longer says "What a Headquarters killed outright does is left for the first card that can kill one". When it fires, it blocks the next card task. The check:
   - **What it flags:** a card whose definition can kill a Headquarters. That is an `OpKill` step whose subject can be a Headquarters (`SubjectTarget` with a target filter that admits one, `SubjectAll` or `SubjectPick` with such a filter, `SubjectWatched` with a watch filter that admits one, `SubjectSelf` or `SubjectReplaced` on a Headquarters), or a `CostKill` whose filter admits one. A filter admits a Headquarters when its `Kinds` is zero or includes `KindHeadquarters`, its `Zone` is `ZoneAny`, and nothing else in it rules a Headquarters out (for example a Scratch Company criterion). Every part of an effect is walked with `Effect.Parts`, including modes, triggers, activated abilities, replacements and declaration costs.
   - **What it compares against:** an allowlist in `pkg/engine/cards` of the cards the ruling lets kill a Headquarters, empty if the ruling allows none. A card that can do so and isn't listed fails the test, and so does a listed card that can't. Adding a card to the list is the deliberate step.
   - **Where it lives: a cards coverage test, not `cardtown lint`.** What a kill can match is decided by the definition's `Filter`, a Go value, not by the rules text. Reading the text for it would bring back the rules-text synthesis that FRD 03 v1.2 retired. `pkg/toolchain` imports no game package, and the lint runs only on changed files, so it would miss a change in `cards/*.go` or in a filter reader in `ability`. The coverage tests already run in `go tool task check`, so it fails in every review. The filter reader goes in `ability` beside `Filter.InPile`. The walk and the test go in `pkg/cardcheck` beside the coverage checkers (Open Question 2), because package `engine` gains nothing the rules don't need. It needs the loaded sheets only to tell which cards are Headquarters, for `SubjectSelf` and `SubjectReplaced`.
   - **Kernel change, if any:** `kill` skips a Headquarters today ([sweep.go:34](../../pkg/engine/sweep.go)). If the ruling makes a killed Headquarters mean something, the same task or a `shape` task before it implements that in the kernel, with its `TestFR_CBT_003_*` test.

   Done when `TestCards_HeadquartersKills` passes against the allowlist and fails on a synthetic definition of each flagged kind.

## Open Questions

1. **Who writes the M3 ability survey?** `docs/design/` is outside `lead-architect`'s write list (only `Status: open` stubs in `outcomes/`). The M2 survey and write-up were written there anyway. The user decides: the lead, with a one-line addition to its skill through `workflow-steward`, or another persona.
2. **The `pkg/cardcheck` package** (Runbook D2.3: a responsibility no package has). The proposed `pkg/catalog` lost its first job when the export moved to M5. What is left is checks over the whole card set that are not rules: the coverage and definition-to-text checkers (which the "test files hold only tests" rule has to move out of `defs_test.go`) and the Headquarters-kill check. None of them belongs in `engine`, and `cards` can't import `engine`. Where the M5 export lives is decided at M5. Needs the user's approval before Deliverable 1's task is written. If the user declines it, the checkers stay in `pkg/engine/defs_test.go` as M2 left them, the test-files rule waits for a later home, and Deliverable 6a's walk and test go beside them.
3. **A Headquarters killed outright** (`FR-CBT-003`, `SVY-R46`; M2 Open Question 21). Waiting on `game-designer`'s outcome. No card's text kills one today: every kill names a unit, a structure or a "non-Headquarters asset". Until the ruling, `kill` keeps skipping a Headquarters, and a definition whose filter admits one would silently kill nothing. Deliverable 6a catches that once it exists.
4. **A combat-kill attacker that dies in the same sweep** (carried from M2 Open Question 9, `FR-CBT-002`). *Dread Manifestation* is the first unit with the hook and is in the Concordia draft, so the survey asks it before the card's wave.
5. **A pile card that leaves and returns while targeted** (carried from M2 Open Question 10, `FR-ZON-004`, `FR-TGT-003`). Re-checked against each M3 wave's cards in its survey row. The first card that can do this gets a zone-change count on `Instance`.
6. **A seat that concedes or leaves** (carried from M2 Open Question 7). M4/M5 work, not M3.
