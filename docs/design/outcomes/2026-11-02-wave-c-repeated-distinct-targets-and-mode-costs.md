Status: approved (user, 2026-11-02)
Directive: #91
Rulings: SVY-R32, SVY-R33, SVY-R34, SVY-R35

### 🎲 Design Outcome: repeated-target timeouts, mode costs, "another" and separate target clauses (2026-11-02)

- **Question:** Four questions raised by lead-architect from the Wave C run (`autoloop/m2-wave-c`, session 25): how a timed-out *Steel Rain* spreads its repeated targets (board directive #59 (2)); whether a mode's costs decide whether it is legal and lock with it (#61, fifth question); what "another chosen unit" requires (*Ambush Tactics* mode 3; design note Open Question 16; #88); and whether two separate "chosen" clauses in one effect may pick the same entity (*Ensnaring Mist*, *Pincer Tactics*; Open Question 16).
- **Rulings:**

  | Ruling | Settles | Cards |
  | :--- | :--- | :--- |
  | `SVY-R32` | A timed-out repeated target choice puts every repetition on the first legal entity in fixed order | Steel Rain |
  | `SVY-R33` | A mode is legal only if its targets exist **and** its costs can be paid; the mode, its targets and its costs lock together on placement; all chosen modes are one stack item | Ambush Tactics, Rogue Fleshshaper (no M2 mode has a cost) |
  | `SVY-R34` | "Another chosen X" must be a different entity: from the earlier slot it follows in the same clause, or from the source when there is none. A mode or Maneuver is legal only if enough different entities exist; a timeout takes the next different one | Ambush Tactics; Tempest Runner (M3) |
  | `SVY-R35` | Each separate "chosen X" clause is its own target slot, and two slots may pick the same entity. Only a clause that points back ("it", "that unit", "the chosen unit") reuses an earlier target | Ensnaring Mist, Pincer Tactics; MAX-8 Autonomous Warsuit, Frostfire Breath, Piston Fist (outside M2) |

- **Decisions (user, 2026-11-02):**
  - `SVY-R32`: "concentrate, although blowing up your own HQ with your steel rain is pretty funny"
  - `SVY-R33`: "To pick a mode you have to have valid targets and pay any additional costs that may be attached to that mode (that's an interesting design space). … all modes become part of the same stack item too - IE they don't resolve as item 1 mode 1, item 2 mode 2 - it's item 1, mode 1 then mode 2"
  - `SVY-R34`: "this would be an invalid mode on a board with one unit, because the 'another' means it has to have at least 2 picks." The user approved extending "another" to mean "other than the source" when no earlier slot is in the clause (*Tempest Runner*), and approved that repeatable modes inherit `SVY-R32`'s timeout.
  - `SVY-R35`: "unless the sentences are conjoined with a comma (move chosen unit, it gets) or a subsequent statement directly refers to the first one (If you did, that unit ...), they're two distinct effects that can have their own targets." *Pincer Tactics*: "it's a valid (if greedy) line of play … if they wanted a 3rd Unbowed trigger from their own guy instead of 2 damage, that's valid." *Ensnaring Mist*: "2 for 1 … Shove one thing, make combat on another one more favorable." *MAX-8*: "can damage and heal the same guy if desired." *Piston Fist*: "should be split into a 'deal 1. pump 1.' at this point, I think. The and makes it potentially confusing."
- **Cards affected:**
  - *Steel Rain*: M2, wave C, Path starter deck.
  - *Ambush Tactics*: M2, wave C, Path starter deck. The CSV already says "another"; the physical prototype is behind it.
  - *Ensnaring Mist*: M2, wave C, not in a starter deck. No card change.
  - *Pincer Tactics*: M2, wave C, not in a starter deck. No card change.
  - *Piston Fist*: Soulless, outside M2. Card text reworded.
  - *MAX-8 Autonomous Warsuit*, *Frostfire Breath*, *Tempest Runner*: outside M2. No card change; they follow `SVY-R35` and `SVY-R34` as written.
  - Checked and unaffected: *Conversion Spike*, *Soul Whip* and *Saboteur*. Each has one allied and one opposing slot, so the two slots can never pick the same unit.
- **Alternatives rejected:**
  - `SVY-R32` (b), cycling down the fixed order: it treats the independent choices of `SVY-R4` as one linked choice, and it isn't how the table plays.
  - Checking "another" slot by slot: it would offer a mode that can't be completed, which is what `SVY-R29` rules out.
  - Treating two "chosen" clauses as one target unless the card says "another": it makes *Ensnaring Mist* a single-unit effect, which is not the intended two-for-one.
  - Leaving *Piston Fist*'s "and" in place: it reads as one target, the ambiguity `SVY-R35` removes.

#### → `pm-frd-drafter` (FRD 01 and the ruling register)
- **Superseded wording:**
  - `FR-MOD-001` §5 (FRD 01 line 384) and the v3.19 summary item `SVY-R28` (line 21): "A mode is legal only if each of its exactly-one target slots has a legal entity." It now also requires that the mode's costs can be paid.
  - Nothing else states a rule these rulings reverse; the rest of the text is silent.
- **FRD 01 changes** (v3.19 → v3.20):
  1. **`FR-MOD-001` §6, *Mandatory Targets and Untargeted Selections*** (line 389). After "…the first ones in **fixed order**." insert: "A target choice made separately for each repetition (`SVY-R4`) times out to the first legal entity for every repetition, so all of them land on one entity: a timed-out *Steel Rain* that discarded 3 cards deals all 6 damage to the first legal asset (`SVY-R32`)."
  2. **Repeated Targets (`SVY-R4`)** (line 359). Append: "On a timeout every repetition picks the same entity, the first legal one in fixed order (`FR-MOD-001` §6, `SVY-R32`)."
  3. **`FR-MOD-001` §5, Distinct Modes** (line 384). Replace "A mode is **legal** only if each of its exactly-one target slots has a legal entity when the effect is placed." with: "A mode is **legal** only if each of its exactly-one target slots has a legal entity, and every cost the mode itself carries can be paid, when the effect is placed (`SVY-R28`, `SVY-R33`). Choosing a mode, choosing its targets and paying its costs are one commitment made on placement (`SVY-R2`, `SVY-R3`), as for a Maneuver's additional cost (`FR-TRN-003a`)." After the `SVY-R31` sentence add: "All chosen modes belong to one stack item and resolve one after another as that item resolves. They are never separate items (`SVY-R33`)."
  4. **Targets Are Locked on Placement (`SVY-R3`)** (line 354). After "together with its costs (`SVY-R2`)" add: ", including the costs of its chosen modes (`SVY-R33`)".
  5. **New bullet after "Up To N" Targets Are Distinct (`SVY-R11`)** (after line 363): "**"Another" Targets (`SVY-R34`)**: A slot that says *"another chosen X"* (or *"other chosen"*) must pick an entity different from the slot it follows in the same clause (*Ambush Tactics*: *"Chosen unit deals damage equal to its power to another chosen unit"*). With no earlier slot in its clause, it must differ from the effect's source (*Tempest Runner*: *"Regroup another chosen allied unit"*). A mode or Maneuver with an *"another"* slot is legal only if enough different legal entities exist to fill every exactly-one slot at once, so *Ambush Tactics*' third mode is illegal with one unit on the battlefield (`SVY-R28`, `SVY-R29`). The *"another"* slot is offered every legal entity except the earlier pick. On a timeout, the earlier slot takes the first legal entity in fixed order and the *"another"* slot takes the first one after it that is different (`FR-MOD-001` §6)."
  6. **New bullet after that one:** "**Separate Target Clauses (`SVY-R35`)**: Each *"chosen X"* in an effect is its own target slot with its own choice, even when two clauses are joined by *"and"* or a comma. Two slots may pick the same entity unless one says *"another"* (`SVY-R34`) or their restrictions can't both match it. A clause that points back to an earlier target (*"it"*, *"that unit"*, *"the chosen unit"*) has no slot of its own and acts on that earlier target. Examples: *Ensnaring Mist* may move one unit to Reserves and give -1 power to the same unit or a different one. *Pincer Tactics*' *"chosen non-Scratch asset"* may be one of its *"up to 2 chosen allied units"*, which then takes 1 + 2 damage. *Frostfire Breath*'s *"the chosen unit"* is its one target. Different modes of one effect already work this way (`SVY-R25`)."
  7. **`FR-MNV-001`, A Maneuver Needs Its Required Targets** (line 334). After "…has a legal entity when it is declared (`FR-TGT-001`)." add: "Slots that must differ (*"another"*, `SVY-R34`) need different entities. A Maneuver whose modes carry costs counts only the modes whose costs can be paid (`SVY-R33`)."
  8. **Revision summary, v3.20 block** (in the format of the v3.19 block at lines 18-22):
     - "1. **Repeated-Target Timeouts (`SVY-R32`; `FR-MOD-001` §6, `FR-TGT-001`)**: a timed-out per-repetition target choice puts every repetition on the first legal entity in fixed order."
     - "2. **Mode Costs and One Stack Item (`SVY-R33`; `FR-MOD-001` §5)**: a mode is legal only if its costs can be paid as well. Mode, targets and costs lock as one commitment on placement, and all chosen modes are one stack item."
     - "3. **"Another" Targets (`SVY-R34`; `FR-TGT-002`, `FR-MNV-001`)**: an *"another"* slot must pick a different entity, and legality needs enough different entities."
     - "4. **Separate Target Clauses (`SVY-R35`; `FR-TGT-001`)**: each *"chosen"* clause is its own slot and may pick the same entity as another; only *"it"*, *"that unit"* or *"the chosen unit"* reuses a target. Card text: *Piston Fist* split into two sentences."
- **Register rows:**
  - Survey §4: add rows **R32**–**R35**, one per ruling above, with the Settles text and Cards from the Rulings table. Status cell, each: "**Ruled (user, 2026-11-02).** Recorded in FRD 01 v3.20 as `SVY-R3x` (`FR-…`)", with the requirement IDs from the table. Add "R32–R35 on 2026-11-02" to the survey's Rulings line (line 18).
  - Survey card rows for *Steel Rain*, *Ambush Tactics*, *Ensnaring Mist* and *Pincer Tactics*: wherever a note cites Open Question 16 or a pending ruling, replace it with the new ruling ID.
  - Keyword matrix: no rows affected.
- **Card text:** `data/cards/maneuvers-soulless.csv`, *Piston Fist*:
  - Old: `Deal 1 damage to chosen unit, and chosen allied unit gets +1 power until end of turn.`
  - New: `Deal 1 damage to chosen unit. Chosen allied unit gets +1 power until end of turn.`
- **Prompt to run:** `@pm-frd-drafter apply Design Outcome docs/design/outcomes/2026-11-02-wave-c-repeated-distinct-targets-and-mode-costs.md`

#### → `lead-architect` (build board directive)
- **Directive row** #91 in `.agents/tasks/m2/BOARD.md`, status `waiting: docs/design/outcomes/2026-11-02-wave-c-repeated-distinct-targets-and-mode-costs.md applied (then open: before the first task whose cards: include Steel Rain, Ambush Tactics, Ensnaring Mist or Pincer Tactics)`:

  **Wave C: repeated-target timeouts, mode costs, "another" and separate clauses (`SVY-R32`–`SVY-R35`; `FR-MOD-001` §5, §6, `FR-TGT-001`, `FR-TGT-002`, `FR-MNV-001`; design note Open Question 16).** Answers #59 (2) and #61's fifth question.
  **(1) `SVY-R32`, in the *Steel Rain* task.** In `fixedOrderPicks` (`pkg/engine/choice.go:203-207`), a `Repeats` decision returns `idx[0]` for every pick instead of cycling. This applies to every `Repeats` decision, so a mode that may repeat (none exists yet) also times out to the first listed legal mode each time. Test: `TestFR_MOD_001_RepeatedTargetTimeoutConcentrates`: 3 repetitions, two legal assets, all 3 picks are the first one in fixed order.
  **(2) `SVY-R33`, a convention for the first task with a cost on a mode (no M2 card has one).** `legalModes` (`pkg/engine/placement.go:36`) also drops a mode whose costs can't be paid, and those costs are paid in the same placement step as its targets. One stack item per placement already holds (`chooseModes` joins the chosen modes into one effect), so nothing changes now.
  **(3) `SVY-R34`, in the *Ambush Tactics* task.** `ability.Target` gets a marker meaning "differs from slot k" (or from the source when there is no earlier slot; *Tempest Runner*, M3, reuses it). `chooseTargets` (`pkg/engine/targeting.go:20-36`) leaves the earlier pick out of that slot's options. `legalModes` and #88 (a)'s enact check need a joint assignment of different entities, not slot-by-slot checks. On a timeout the second slot takes the first legal entity after the first slot's pick. Tests: (a) `TestFR_MOD_001_AnotherSlotNeedsTwoUnits`: with one unit on the battlefield, *Ambush Tactics* mode 3 is not offered; with two it is. (b) `TestFR_TGT_002_AnotherSlotExcludesEarlierPick`. (c) `TestFR_MOD_001_AnotherSlotTimeoutTakesNextDifferent`.
  **(4) `SVY-R35`, in the *Ensnaring Mist* and *Pincer Tactics* tasks.** No kernel change: independent slots already allow the same entity. Each card task adds a test that one entity filling both slots takes both steps (*Pincer Tactics*: 1 + 2 damage to one allied non-Scratch unit).
  Design note Open Question 16 closes, citing this outcome.
  **Merged code that contradicts the rulings:** `fixedOrderPicks`' cycling contradicts `SVY-R32`; no test pins it (searched `Repeats` in `pkg/engine/*_test.go`; the only repeats timeout test is the single-option fast path, which holds either way). `chooseTargets` lets a later slot name an earlier slot's pick, which contradicts `SVY-R34` only for "another" slots, which `ability.Target` can't express yet. Nothing contradicts `SVY-R33` or `SVY-R35`.
- Also set #59's status to "(1) applied (T-0077, f6da866); (2) answered by #91" and #61's status to "answered: questions 1–4 by #88, question 5 by #91".

#### → `pm-lead` (portfolio)
- **Milestone or scope impact:** None. *Piston Fist* is reworded but stays outside M2.
