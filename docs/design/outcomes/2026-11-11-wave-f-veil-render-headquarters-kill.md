Status: applied: FRD 01 v3.24, 2026-11-11
Directive: #143
Rulings: SVY-R46

### 🎲 Design Outcome: Veil Render's discard-deploy cost kills an allied unit (2026-11-11)

- **Question:** Can *Veil Render*'s discard-deploy cost, "killing an allied asset", be paid by killing its controller's own Headquarters? If it can, what does a Headquarters killed outright do, and what does a timed-out cost pick take, given that the fixed order puts the Headquarters first? Asked by T-0216 (wave F run, `autoloop/m2-wave-f`, session 17) before the task was written; design note Open Question 21.
- **Rulings:**

  | Ruling | Settles | Cards |
  | :--- | :--- | :--- |
  | `SVY-R46` | *Veil Render*'s discard-deploy cost kills an allied **unit**, Scratch Company units included. A Headquarters or a Structure can't be killed to pay it, so a timed-out pick takes the first allied unit in fixed order (`FR-MOD-001` §6) and needs no exception. No general rule stops a Headquarters from being killed. What a Headquarters killed outright does is left unruled on purpose, for the first card that can kill one (for example a "destroy a player" win condition). | Veil Render (reworded) |

- **Decisions (user, 2026-11-11):**
  - Wording (option D): "the concept of a veil render is something that swallows souls to cross through a veil, and while animus structures might blur that line, that means units."
  - Design space: "door to nothingness was a wild card in MTG that was originally worded as "destroy target player", so I want to leave room for a weird win con like that." So there is no general rule that a Headquarters can't be killed.
  - Scratch Company units: "scratches count, they're the definition of fodder."
  - A killed Headquarters (question 2) stays unruled until a card needs it: "that's fine."
- **Cards affected:**
  - In M2 and in the Animus starter deck (1 copy): *Veil Render* (reworded; no engine definition yet, still in `cards.NotYetExecutable`).
  - In M2 and in the Animus starter deck, no text change: *Cliff of Portents*, *Rhapsody's Hollow*, *Animoceps Spore Cloud*, *Animoceps Node*. These Structures are no longer legal fodder for *Veil Render*. No other card relies on them being fodder.
  - Outside M2, no change: *NO-AH-5* (Soulless), which already says "allied non-Headquarters asset".
- **Alternatives rejected:**
  - A. "An allied non-Headquarters asset" (the lead's recommendation, *NO-AH-5*'s wording): Structures aren't souls, and the card's concept is swallowing souls.
  - B. A general FRD 01 rule that a Headquarters can't be killed: it would close off a future "destroy a player" style win condition.
  - C. The literal reading (the Headquarters is legal, killing it loses the match, timeouts skip it): this adds a concede button, a timeout exception and a new defeat path, all for a choice nobody should make.
  - "An allied non-Scratch unit": Scratch Company units are meant to be fodder.

#### → `pm-frd-drafter` (FRD 01 and the ruling register)
- **Superseded wording:**
  - `docs/frd/01_Game_Rules_FRD.md:360` (`FR-TRN-003a` *In-Resolution Free Deployments*): "waives its additional "kill an allied asset" cost".
  - `docs/design/m2_starter_set_ability_survey.md:157` (*Veil Render* row): "deploy from discard + kill allied asset".
  - Outside pm-frd-drafter's files: FRD 02 benchmark #6 (`docs/frd/reference/02_Benchmark_Card_Blueprints.md:218`, `entityType: ASSET`) is frozen design intent and is not changed. The card definition follows `SVY-R46` instead. Design note Open Question 21 and T-0216's Goal and Covers on `autoloop/m2-wave-f` are updated by the lead (directive #142 (3)).
- **FRD 01 changes** (v3.23 → v3.24):
  - `FR-TRN-003a` *In-Resolution Free Deployments* (line 360): `"kill an allied asset"` → `"kill an allied unit"`.
  - `FR-TRN-003a` *Additional Sacrifice, Discard & Exhaust Costs* (line 359): add at the end: *"A kill cost names the kind of asset it kills. *Veil Render*'s kills an allied unit, Scratch Company units included (`SVY-R46`), so neither a Headquarters nor a Structure can pay it."*
  - `FR-CBT-003` *Headquarters Destruction Defeat* (line 585): add at the end: *"No card can kill a Headquarters outright, as opposed to reducing its health to 0 or below. What a Headquarters killed outright does is left for the first card that can kill one (`SVY-R46`)."*
  - Revision summary `(v3.23 -> v3.24): Veil Render's Kill Cost (2026-11-11)`:
    1. **A Kill Cost Names a Unit (`SVY-R46`; `FR-TRN-003a`, `FR-CBT-003`)**: *Veil Render*'s discard-deploy cost kills an allied unit, Scratch Company units included, and never a Headquarters or Structure. A Headquarters killed outright stays unruled until a card can kill one. Card text: *Veil Render*.
- **Register rows:**
  - Survey §4: add `R46` with the question "May *Veil Render*'s "kill an allied asset" kill its own Headquarters, and what does a killed Headquarters do? (Open Question 21, T-0216)", the cards "Veil Render (reworded)", and status **"Ruled (user, 2026-11-11).** Recorded in FRD 01 v3.24 as `SVY-R46` (`FR-TRN-003a`, `FR-CBT-003`)."
  - Survey card row *Veil Render* (`:157`): effect `deploy from discard + kill allied unit`; note `Benchmark #6 (cost narrowed to a unit, SVY-R46); replaced kill cost stays paid (SVY-R15)`.
  - Survey "Noted but unreachable" (after `:106`): *"What a Headquarters killed outright (not damaged to 0) does. No M2 card can kill one (`SVY-R46`). Left for the first card that can, such as a "destroy a player" win condition."*
  - No keyword-matrix row changes.
- **Card text** (`rulesText` cell only; *Veil Render* has no definition yet, so `TestDefs_MatchRulesText` doesn't pin it):
  - *Veil Render* (`data/cards/units-and-structures-animus.csv:17`): `You may deploy this card from your discard pile by killing an allied asset in addition to paying its deployment costs.` → `You may deploy this card from your discard pile by killing an allied unit in addition to paying its deployment costs.`
- **Prompt to run:** `@pm-frd-drafter apply Design Outcome docs/design/outcomes/2026-11-11-wave-f-veil-render-headquarters-kill.md`

#### → `lead-architect` (build board directive)
- **Directive row:** #143 in [`.agents/tasks/m2/BOARD.md`](../../../.agents/tasks/m2/BOARD.md), status `waiting: docs/design/outcomes/2026-11-11-wave-f-veil-render-headquarters-kill.md applied (then open: before T-0216)`.
  - **(1) T-0216's definition.** The cost is one `CostKill` with `Filter{Allegiance: AllegianceAllied, Kinds: KindUnit}`, the same filter as *Feed the Despoiler* (`pkg/engine/cards/animus.go:291`), not the benchmark's `entityType: ASSET`. `Text` is the reworded CSV line.
  - **(2) T-0216's Covers.** Rewrite them for a unit. Pin these: killing an allied Scratch Company unit pays the cost. With only allied Structures and the Headquarters on the battlefield, the discard deploy isn't legal. A timed-out cost pick kills the first allied unit in fixed order (Reserves before the Main Theater, oldest first) and never the Headquarters. Keep the existing Covers (death triggers resolve above it, a replaced kill still pays under `SVY-R15`, the hand deploy has no kill, free deploys waive the kill, not legal from the opponent's pile).
  - **(3)** Mark the design note's Open Question 21 answered, and fix T-0216's title and Goal ("by killing an allied asset" → "an allied unit").
  - **(4) Leave as it is:** `kill` still skips every Headquarters (`pkg/engine/sweep.go:31`, pinned by `TestFR_LAY_002_KillSkipsAHeadquarters`). That remains a documented M2 limit, not a rule. A Headquarters killed outright is ruled by the first card that can kill one.
  - **Checked for contradictions:** none found. `git grep "Veil Render"` in `pkg` and `cmd` on `main` and `autoloop/m2-wave-f` finds only the `cards.NotYetExecutable` entry, the deck count in `pkg/engine/deck_load_test.go:277`, and a comment in `pkg/view/executable_test.go:14`. There's no definition, and T-0216 is still `queued`. This ruling doesn't change `cardtown sim` output on its own. T-0216 already changes it by making the card executable.

#### → `pm-lead` (portfolio)
- **Milestone or scope impact:** None.
