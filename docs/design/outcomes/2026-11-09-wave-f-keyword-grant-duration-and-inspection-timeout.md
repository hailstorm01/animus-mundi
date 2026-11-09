Status: applied: FRD 01 v3.22, 2026-11-09
Directive: #140
Rulings: SVY-R39, SVY-R40, SVY-R41, SVY-R42, SVY-R43

### 🎲 Design Outcome: how long a duration reaches, its wording, and a timed-out look at the top cards (2026-11-09)

- **Question:** Does *Allshan's Battle Trance*'s "until end of turn" also cover "gains Stalwart"? And when the one prompt of *Devilhorn Scouting Party*'s look at the top 3 cards times out, is a unit still deployed, and in what order do the rest go to the bottom? Asked by lead-architect's wave F readiness check (design note Open Question 19). A knock-on case came up in the discussion: the timeout for a free deploy's choice of zone.
- **Rulings:**

  | Ruling | Settles | Cards |
  | :--- | :--- | :--- |
  | `SVY-R39` | A duration phrase covers every temporary effect in its sentence. Armor never expires (`FR-LAY-002`), so a duration never covers an armor grant. | Allshan's Battle Trance, Redline Infusion, Field Mechanic, March of Progress |
  | `SVY-R40` | Card text: when two or more temporary effects share one duration, the duration starts the sentence ("Until end of turn, X and Y."). An effect that doesn't expire, such as an armor grant, gets a sentence of its own. | Allshan's Battle Trance, Redline Infusion, Field Mechanic, March of Progress, Chainbreaker's Tenacity, Land Fortress |
  | `SVY-R41` | A timeout does what the printed text says. In a look at the top cards, a pick the card makes optional ("you may", "up to N") takes no card. A pick the card requires takes the first eligible card(s) from the top. This is the same rule as `SVY-R37`/`SVY-R38`: *Cliff of Portents* says "put it into your hand", which is not optional, and a search with criteria may come up empty only because a hidden deck allows it, not because the card says "may". | Devilhorn Scouting Party; Light of the World Soul, Hourglass Diviner, Samsara (outside M2) |
  | `SVY-R42` | Cards a look puts on the bottom of the deck, when the player hasn't arranged them (a timeout), keep the order the player saw: the higher of two cards stays higher. | Devilhorn Scouting Party; Light of the World Soul (outside M2) |
  | `SVY-R43` | When a free-deployed card may enter either Reserves or the Main Theater (`FR-TRN-003a` 1), a timeout puts it in Reserves, following the fixed order of `SVY-R16`. In a look at the top cards, the zone of each deployed card is part of that look's one prompt (`FR-MOD-001` §4). | Reaver's Rush (with Devilhorn Guerrillas); Mutation Swell, Devilhorn Scouting Party, Light of the World Soul, World Soul Messenger |

- **Decisions (user, 2026-11-09):**
  - `SVY-R39`: "for question 1, A is the intention".
  - `SVY-R40`: "we should also establish that when there's two or more effects tied to a single duration qualifier, it gets worded as C proposes since the "until end of turn," more clearly applies to everything that comes after it." Armor: "It's probably worth separating the "gains 2 armor" for cards like this into its own sentence entirely, to make it crystal clear it's not at all related to a duration clause for another effect."
  - `SVY-R41`: "the "may" is not incidental to the wording of the card, so it gets treated as an optional benefit and follows the "opt not to" standard." On `SVY-R38`: "Cliff of Portents doesn't say "you may put a card into your hand", it says "and put it into your hand". It's not optional".
  - `SVY-R42`: "default to order they saw them is fine."
  - `SVY-R43`: "reserves default is good."
- **Cards affected:**
  - In M2 and in the Path starter deck: *Allshan's Battle Trance* (reworded), *Devilhorn Scouting Party* (timeout), *Chainbreaker's Tenacity* (reworded; already built).
  - In M2, not in a starter deck: *Reaver's Rush* (zone timeout; *Devilhorn Guerrillas*, 3 power, is the only M2 unit that can deploy to both zones). *Mutation Swell* is affected in principle only, since no Animus unit can deploy to both zones.
  - Outside M2: *Redline Infusion*, *Field Mechanic*, *March of Progress*, *Land Fortress* (reworded); *Light of the World Soul*, *Hourglass Diviner*, *Samsara, Living Tempest*, *World Soul Messenger* (behavior under the new rules, no text change).
- **Alternatives rejected:**
  - Stalwart lasting while the unit stays on the battlefield: a permanent grant from a one-shot Maneuver would be the only one in the card pool, and the earlier blueprint (benchmark #11) already gives it `END_OF_TURN`.
  - A timeout deploying the first eligible unit (the lead's recommendation): the "may" on the card is a real choice, and every other "you may" times out to "no".
  - Removing "may" from *Scouting Party*: the option to decline is intended.
  - Reversing the order on the bottom: it is less predictable and gains nothing.
  - A zone timeout that takes the first printed zone: it would depend on the order of the CSV column, not on a rule.

#### → `pm-frd-drafter` (FRD 01 and the ruling register)
- **Superseded wording:**
  - `docs/frd/01_Game_Rules_FRD.md:351` (`FR-TRN-003a` 2) quotes the old *Battle Trance* text.
  - `:401` (`FR-MOD-001` §6, *Top-of-Deck Inspections*) says "selects the first eligible card(s)" for every pick and gives no order for the rest.
  - `docs/design/m2_starter_set_ability_survey.md:174` (*Chainbreaker's Tenacity* row) and `:197` (*Battle Trance* row) describe the old text.
- **FRD 01 changes** (v3.21 → v3.22):
  - `FR-LAY-003`, a new first bullet: **Duration Reach & Wording (`SVY-R39`, `SVY-R40`)**: *"A duration phrase covers every temporary effect in its sentence. Armor never expires (`FR-LAY-002`), so a duration never covers an armor grant. When two or more temporary effects share one duration, the card text starts the sentence with it: Allshan's Battle Trance reads "Until end of turn, it gains Stalwart and gets +2 power." An effect that doesn't expire, such as an armor grant, is its own sentence: Chainbreaker's Tenacity reads "Chosen allied unit gets +1 power until end of turn. It gains 2 armor.""*
  - `FR-TRN-003a` 2 (line 351): replace the quote with *"Deploy chosen unit card from your discard pile to the Main Theater. Until end of turn, it gains Stalwart and gets +2 power."*
  - `FR-TRN-003a` 1 (line 350): add at the end *"If the deploying player doesn't choose before the action clock and Reserve Time Bank expire, the card enters Reserves (`SVY-R43`, `FR-MOD-001` §6)."*
  - `FR-MOD-001` §4 (line 398): add at the end *"The zone of each card deployed this way (`FR-TRN-003a` 1) is chosen in the same prompt (`SVY-R43`)."*
  - `FR-MOD-001` §6, replace *Top-of-Deck Inspections* with: *"Top-of-Deck Inspections (`SVY-R41`, `SVY-R42`): A timeout does what the printed text says. A pick the card requires takes the first eligible card(s) from the top (Hourglass Diviner puts the top card into hand). A pick the card makes optional ("you may", "up to N") takes no card: a timed-out Devilhorn Scouting Party deploys nothing and still Scratches 1. The rest go to the bottom of the deck in the order they were seen, top first. This is the deck-search rule as well (`SVY-R37`, `SVY-R38`): a search with criteria may come up empty because the deck is hidden, not because the card says "may", so its timeout still takes the first eligible card."*
  - `FR-MOD-001` §6, add a bullet after *Deck Searches*: *"Free-Deploy Zone (`SVY-R43`): a free-deployed card that may enter either Reserves or the Main Theater (`FR-TRN-003a` 1) enters Reserves, the fixed order's first battlefield zone."*
  - Revision summary `(v3.21 -> v3.22): Duration Reach and Wording, and Timed-Out Looks and Free-Deploy Zones (2026-11-09)`:
    1. **Duration Reach (`SVY-R39`; `FR-LAY-003`)**: a duration phrase covers every temporary effect in its sentence, never an armor grant.
    2. **Duration Wording (`SVY-R40`; `FR-LAY-003`)**: shared durations start the sentence, and armor grants get a sentence of their own. Card text: *Allshan's Battle Trance*, *Chainbreaker's Tenacity*, *Redline Infusion*, *Field Mechanic*, *March of Progress*, *Land Fortress*.
    3. **Timed-Out Looks (`SVY-R41`; `FR-MOD-001` §6)**: a timeout does what the printed text says. An optional pick takes no card, and a required pick takes the first eligible card.
    4. **Bottom Order (`SVY-R42`; `FR-MOD-001` §6)**: on a timeout, the rest go to the bottom in the order they were seen.
    5. **Free-Deploy Zone (`SVY-R43`; `FR-TRN-003a`, `FR-MOD-001` §4, §6)**: a timeout deploys to Reserves, and a look's zone choices are part of its one prompt.
- **Register rows:**
  - Survey §4: add `R39`–`R43`, each with status **"Ruled (user, 2026-11-09).** Recorded in FRD 01 v3.22 as `SVY-R<n>` (…)".
  - Survey card rows:
    - *Chainbreaker's Tenacity* (`:174`): note "Text reworded under `SVY-R40` (CSV changed with its definition by directive #140)".
    - *Devilhorn Scouting Party* (`:195`): note "Timeout deploys nothing, rest to bottom in seen order (`SVY-R41`, `SVY-R42`); zone in the same prompt (`SVY-R43`)".
    - *Reaver's Rush* (`:196`): note "Zone timeout: Reserves (`SVY-R43`)".
    - *Allshan's Battle Trance* (`:197`): note "Stalwart and +2 both until end of turn (`SVY-R39`); text reworded (`SVY-R40`)".
  - No keyword-matrix row changes.
- **Card text** (`rulesText` cell only). **Do not change *Chainbreaker's Tenacity*'s CSV row:** directive #140 changes it together with its definition, because `TestDefs_MatchRulesText` pins it.
  - *Allshan's Battle Trance* (`maneuvers-path.csv`): `Deploy chosen unit card from your discard pile to the Main Theater. It gains Stalwart and gets +2 power until end of turn.` → `Deploy chosen unit card from your discard pile to the Main Theater. Until end of turn, it gains Stalwart and gets +2 power.`
  - *Redline Infusion* (`maneuvers-soulless.csv`): `Chosen allied unit gets +2 power and gains "Overkill: Ready this unit." until end of turn. At the end of your turn, kill that unit.` → `Until end of turn, chosen allied unit gets +2 power and gains "Overkill: Ready this unit." At the end of your turn, kill that unit.`
  - *March of Progress* (`maneuvers-soulless.csv`): `Up to 2 chosen units gain the Mechanized subtype until end of turn. Allied Mechanized units gain 2 armor and get +2 power until end of turn.` → `Up to 2 chosen units gain the Mechanized subtype until end of turn. Allied Mechanized units gain 2 armor. They get +2 power until end of turn.`
  - *Field Mechanic* (Soulless units CSV): `When this unit deploys, chosen allied Mechanized unit gains 2 armor and gets +2 power until end of turn.` → `When this unit deploys, chosen allied Mechanized unit gains 2 armor. It gets +2 power until end of turn.`
  - *Land Fortress* (Soulless units CSV): `Exhaust: Chosen unit in the same zone gets +2 power until end of turn and gains 1 armor.` → `Exhaust: Chosen unit in the same zone gets +2 power until end of turn. It gains 1 armor.`
- **Prompt to run:** `@pm-frd-drafter apply Design Outcome docs/design/outcomes/2026-11-09-wave-f-keyword-grant-duration-and-inspection-timeout.md`

#### → `lead-architect` (build board directive)
- **Directive row:** #140 in [`.agents/tasks/m2/BOARD.md`](../../../.agents/tasks/m2/BOARD.md), status `waiting: docs/design/outcomes/2026-11-09-wave-f-keyword-grant-duration-and-inspection-timeout.md applied (then open: before the Allshan's Battle Trance card task)`.
  - **(1) *Chainbreaker's Tenacity*'s rewording, first, a `flash` harness task with no behavior change.** In one commit, change its `rulesText` in `data/cards/maneuvers-path.csv` to `Chosen allied unit gets +1 power until end of turn. It gains 2 armor.`, its `Text` in `pkg/engine/cards/path.go:229` to match (`TestDefs_MatchRulesText`), and the quote in the comment above its tests (`pkg/engine/cards_path_maneuvers_test.go:523`). pm-frd-drafter leaves this row alone.
  - **(2) *Allshan's Battle Trance*, in its card task.** Its text is now "Until end of turn, it gains Stalwart and gets +2 power." Both go on #137's list of modifiers with the end-of-turn duration. Tests: after Sub-Phase 4c the unit stays in the Main Theater without Stalwart and at printed power.
  - **(3) The look-at-the-top-cards shape (*Devilhorn Scouting Party*).** One prompt per look, holding the deploy picks, their zones and the bottom order (`SVY-R43`). Tests: a timeout deploys nothing, puts all 3 cards on the bottom in the order seen (top first, so the top card ends highest), and Scratch 1 still happens (`SVY-R41`, `SVY-R42`); a player's own bottom order is kept. A required pick (*Hourglass Diviner*, outside M2) takes the first eligible card, and no M2 card needs it.
  - **(4) Free-deploy zone timeout, in the first task that deploys a card that can enter both zones for free (*Reaver's Rush* with *Devilhorn Guerrillas*).** A timeout puts it in Reserves, so the Guerrillas' Main Theater trigger doesn't fire (`SVY-R43`).
  - **(5)** Mark the design note's Open Question 19 answered, and update #137's quote of the *Battle Trance* text.
  - **Checked for contradictions:** no keyword grant, look op or free deploy exists yet (`DeployFree` and `LookTop` aren't in `pkg/engine`; *Battle Trance*, *Scouting Party*, *Reaver's Rush*, *Mutation Swell* and *Eerie Remnant* are in `cards.NotYetExecutable`). *Tenacity*'s tests pin behavior, not text, and the behavior doesn't change. The other reworded cards are outside M2 and have no definitions.

#### → `pm-lead` (portfolio)
- **Milestone or scope impact:** None. No cards move, and no done-when criteria change.
