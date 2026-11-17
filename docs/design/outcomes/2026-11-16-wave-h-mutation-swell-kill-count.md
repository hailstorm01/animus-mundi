Status: approved (user, 2026-11-17)
Directive: #158
Rulings: SVY-R53, SVY-R54

### 🎲 Design Outcome: Mutation Swell's kills become an additional cost (2026-11-17)

- **Question:** For *Mutation Swell*'s "Kill 2 or 4 allied Scratch Company units", when is the count chosen, what does a shortfall (1 or 3 matches) do, and which count does a timeout take? Asked by lead-architect in the wave H run (`autoloop/m2-wave-h`, session 25) before the *Mutation Swell* tasks were written, as a `Status: open` stub at this path. Design note Open Question 25.
- **Rulings:**

  | Ruling | Settles | Cards |
  | :--- | :--- | :--- |
  | `SVY-R53` | *Mutation Swell*'s kills are an additional cost: its controller chooses 2 or 4 on declaration and kills that many allied Scratch Company units. Then, for every 2 units killed, they choose a different unit card in their own discard pile as a target. Count, kills and targets are one commitment on declaration (`SVY-R2`, `SVY-R3`, `FR-TRN-003a`). The cards of the units killed for the cost are not yet in the discard pile, so they can't be chosen. Swell can be enacted only if a count can be paid and its targets filled: 2 needs 2 allied Scratch Company units and 1 unit card in the pile; 4 needs 4 units and 2 different unit cards (`SVY-R29`). A kill that a replacement changes still counts as paid (`SVY-R15`, `SVY-R21`). A target that leaves the pile before resolution is skipped and the other still deploys (`FR-TGT-003`). The deploys use printed zones (`FR-TRN-003a` 1). | Mutation Swell (reworded) |
  | `SVY-R54` | A repeated target clause that says *"different"* picks a different entity for each repetition, an exception to `SVY-R4`. Legality needs that many different legal entities. On a timeout (`FR-MOD-001` §6), the choice of how many to kill takes the first listed count that can be paid (2 for *Mutation Swell*), the cost kills take the first allied Scratch Company units in fixed order, and the targets take the first different legal cards from the top of the pile. | Mutation Swell |

- **Decisions (user, 2026-11-17):**
  - How the physical table plays it: "usually it's "swell, killing 2/4"". A response that leaves fewer Scratch Companies than the count is "not possible".
  - Direction: "make the kill 2/4 an additional cost, and make the card itself target two units in discard. Then there's still a window for future instant speed responses to get rid of the desired reanimate targets".
  - The killed units' own cards can't be targeted (`SVY-R2`): accepted.
  - Deploy count: one target per 2 killed (1 or 2).
  - Card text: as drafted, with "different chosen unit card".
  - Timeout fallback (`SVY-R54`): approved with this report.
- **Cards affected:**
  - *Mutation Swell* (reworded): in M2 (wave H, on `cards.NotYetExecutable`), in neither starter deck.
  - *Shifting Impossibility* (M2, no text change): still replaces Swell's kills, which now are cost kills, with the same result (`SVY-R15`).
  - No other card offers a choice of counts or a "different chosen" repeated target.
- **Alternatives rejected:**
  - The count as a cap, chosen and picked on resolution (the lead's recommendation): the table declares "2/4" on play, and the opponent couldn't answer the reanimation targets.
  - Rewording to "Kill up to 4": a timed-out Swell would do nothing, and 0 or 1 kills would be legal but pointless.
  - Strict exact counts, picked on resolution: needs a new check whether a count can be met, and fizzles on edge cases.
  - Cost kills with untargeted deploys picked on resolution: keeps the reanimation of the killed units but loses the response window the user wants.
  - An exception to `SVY-R2` so cost-killed cards can be targeted: a band-aid that would reopen *Sinuous Growth*'s ruling.
  - Always two targets: makes the choice of 4 pointless.

#### → `pm-frd-drafter` (FRD 01 and the ruling register)
- **Superseded wording:**
  - `docs/frd/01_Game_Rules_FRD.md:404` (*Quantifiers*): Swell as the example of an untargeted count.
  - `docs/frd/01_Game_Rules_FRD.md:528` (`SVY-R15` example): "if *Mutation Swell* kills four allied Scratch Company units … deploys two unit cards from their discard pile".
  - `docs/design/m2_starter_set_ability_survey.md:167` (*Mutation Swell* card row).
  - Left as they are: `docs/frd/01_Game_Rules_FRD.md:111` (an old revision summary, i.e. history); `:376–377` (Swell still deploys for free, in printed zones); FRD 02 benchmark #14 (frozen). The lead updates design note Open Question 25 and its description of the deliverable (line 210) on `autoloop/m2-wave-h` (directive #158 (5)).
- **FRD 01 changes** (v3.28 → v3.29):
  - *Quantifiers* (line 404): replace `*Mutation Swell*'s *"Kill 2 or 4 allied Scratch Company units"* picks its units as it resolves.` with `*Animoceps Spores*' *"return an allied Scratch Company unit"* picks its unit as it resolves.`
  - `FR-TRN-003a` *Additional Sacrifice, Discard & Exhaust Costs* (line 375): add *Mutation Swell* to the list of cards; add at the end: *"A cost may offer a choice of counts (*Mutation Swell*: "kill 2 or 4 allied Scratch Company units"). The count is chosen on declaration together with the cost and the targets, and it must be paid in full (`SVY-R53`)."*
  - After the *Steel Rain* sub-bullet (line 396), add: *"*Mutation Swell* (*"As an additional cost to enact this maneuver, kill 2 or 4 allied Scratch Company units. For every 2 units killed this way, deploy a different chosen unit card from your discard pile for free."*): the count, the kills and one different target for every 2 units killed are chosen on declaration. The killed units' cards are not yet in the discard pile, so they can't be chosen (`SVY-R2`). A target that has left the pile by resolution is skipped, and the other still deploys (`SVY-R53`)."*
  - *A Maneuver Needs Its Required Targets* (line 374): add after the *Reaver's Rush* example: *"*Mutation Swell* can't be enacted without 2 allied Scratch Company units and a unit card already in its controller's discard pile, and its count of 4 also needs 4 such units and 2 different unit cards there (`SVY-R53`)."*
  - *Repeated Targets* (line 401): add at the end: *"A repeated clause that says *"different"* must pick a different entity for each repetition, and is legal only if enough different legal entities exist (`SVY-R54`)."*
  - `FR-MOD-001` §6 (after *Mandatory Targets and Untargeted Selections*, line 436): add *"*Choice of Cost Counts (`SVY-R54`)*: Selects the first listed count that can be paid: a timed-out *Mutation Swell* kills 2. Its kills take the first allied Scratch Company units in fixed order, and its *"different"* targets take the first different legal cards from the top of the discard pile."*
  - `SVY-R15` example (line 528): replace `if *Mutation Swell* kills four allied Scratch Company units and *Shifting Impossibility* replaces all four, four units were killed this way, and its controller deploys two unit cards from their discard pile (`SVY-R21`).` with `if four allied Scratch Company units are killed to pay *Mutation Swell*'s additional cost and *Shifting Impossibility* replaces all four, the cost is paid, and both chosen unit cards are deployed (`SVY-R21`, `SVY-R53`).`
  - Revision summary `(v3.28 -> v3.29): Mutation Swell's Kill Cost (2026-11-17)`:
    1. **A Choice of Counts Is Part of a Cost (`SVY-R53`; `FR-TRN-003a`, `FR-MNV-001`, `FR-TGT-001`)**: *Mutation Swell* kills 2 or 4 allied Scratch Company units as an additional cost, chosen on declaration, and targets one different unit card already in its controller's discard pile for every 2 killed. Card text: *Mutation Swell*.
    2. **"Different" Repeated Targets (`SVY-R54`; `FR-TGT-001`, `FR-MOD-001` §6)**: a repeated clause that says "different" picks a distinct entity each time. A timeout takes the first count that can be paid and the first different legal cards from the top.
- **Register rows:**
  - Survey §4: add `R53` with the question "When does *Mutation Swell* choose 2 or 4, what does a shortfall do, and which count does a timeout take? (Open Question 25, wave H)", the cards "Mutation Swell (reworded); Shifting Impossibility", and the status **"Ruled (user, 2026-11-17).** The kills are an additional cost chosen on declaration; one different chosen unit card in the discard pile per 2 killed. Recorded in FRD 01 v3.29 as `SVY-R53` (`FR-TRN-003a`, `FR-MNV-001`, `FR-TGT-001`)."
  - Survey §4: add `R54` with the question "Does a repeated *"different"* target allow the same entity twice, and what does a timeout take?", the cards "Mutation Swell", and the status **"Ruled (user, 2026-11-17).** Different entities each time; the timeout takes the first count that can be paid and the first different legal cards. Recorded in FRD 01 v3.29 as `SVY-R54` (`FR-TGT-001`, `FR-MOD-001` §6)."
  - Survey `R16` row (line 75), cards cell: `Mutation Swell` → `Mutation Swell (now a cost, SVY-R53)`.
  - Survey card row *Mutation Swell* (line 167): effect `additional cost: KILL 2 or 4 allied Scratch Company (chosen on declaration); DEPLOY_FREE a different chosen unit card from discard per 2 killed`; rules `FR-TRN-003a`, `FR-TGT-001` (`SVY-R53`, `SVY-R54`); note `Benchmark #14 superseded: kills are a cost, deploys target cards already in the pile (SVY-R2); replaced kills still pay (SVY-R15, SVY-R21); timeout kills 2 (SVY-R54)`.
  - No keyword-matrix row changes.
- **Card text** (`rulesText` cell; Swell has no engine definition yet):
  - *Mutation Swell* (`data/cards/maneuvers-animus.csv:12`): `Kill 2 or 4 allied Scratch Company units. For every 2 units killed this way, deploy a unit card from your discard pile for free.` → `As an additional cost to enact this maneuver, kill 2 or 4 allied Scratch Company units. For every 2 units killed this way, deploy a different chosen unit card from your discard pile for free.`
- **Prompt to run:** `@pm-frd-drafter apply Design Outcome docs/design/outcomes/2026-11-16-wave-h-mutation-swell-kill-count.md`

#### → `lead-architect` (build board directive)
- **Directive row:** #158 in [`.agents/tasks/m2/BOARD.md`](../../../.agents/tasks/m2/BOARD.md), status `waiting: docs/design/outcomes/2026-11-16-wave-h-mutation-swell-kill-count.md applied (then open: before the first Mutation Swell task)`.
  - **(1) The cost.** `CostKill` kills one asset (`pkg/engine/ability/ability.go:97`). Swell needs a kill cost whose count is chosen on declaration from {2, 4}, filtered to allied Scratch Company units, and only offering a count that can be paid and whose targets can be filled (`SVY-R53`, `SVY-R29`).
  - **(2) The targets.** One slot per 2 units killed, filtered like *Reaver's Rush*'s without the power limit (`pkg/engine/cards/path.go:643-649`: allied, unit, `ZoneDiscard`). This is like `PerDiscard` (*Steel Rain*), except the picks must be different (`SVY-R54`, as `UpTo`'s distinct picks under `SVY-R11`), not independent (`SVY-R4`). The killed units' cards are never offered (`SVY-R2`).
  - **(3) The step.** `OpDeployFree` on each target still in the discard pile at resolution; a gone target is skipped (`FR-TGT-003`).
  - **(4) Tests to pin:** not enactable with 1 allied Scratch Company unit, or with 2 and no unit card in the pile; with 3 units, or with 4 and only 1 unit card, only 2 is offered; the killed units' cards aren't offered as targets; *Shifting Impossibility* replacing every cost kill still deploys every target; one target removed before resolution and the other still deploys; a timeout kills 2 (the first Scratch Companies in fixed order) and takes the first different unit cards from the top of the pile.
  - **(5)** On `autoloop/m2-wave-h`, mark design note Open Question 25 answered and rewrite the line 210 description ("'2 or 4' target counts for an untargeted kill picked on resolution") to describe the cost and its different targets.
  - **Checked for contradictions:** none found. `git grep -i "mutation swell"` in `pkg` and `cmd` on `main` and `autoloop/m2-wave-h` finds only the `cards.NotYetExecutable` entry (`pkg/engine/cards/cards.go`). No task for Swell has been written yet. Swell isn't in either starter deck, so this ruling doesn't change `cardtown sim` output.

#### → `pm-lead` (portfolio)
- **Milestone or scope impact:** None.
