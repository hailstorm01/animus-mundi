Status: applied: FRD 01 v3.27, 2026-11-15
Directive: #156
Rulings: SVY-R50

### 🎲 Design Outcome: one timeout rule for every hand (2026-11-15)

- **Question:** When the player choosing a card from an opponent's revealed hand (*Drifting Concept*: "You choose a card from it") times out (`FR-MOD-001` §6), which card is picked? `SVY-R16` covers only choosing from your own hand. Raised by lead-architect in the wave H run (`autoloop/m2-wave-h`, session 3) before the *Drifting Concept* card task was written, as a `Status: open` stub at this path; design note Open Question 23.
- **Rulings:**

  | Ruling | Settles | Cards |
  | :--- | :--- | :--- |
  | `SVY-R50` | A timed-out choice of cards from any Hand selects the card(s) most recently added to that Hand, whoever is choosing. `SVY-R16`'s own-hand rule becomes the case where the chooser is the owner. A choice from an opponent's revealed hand is still an inspection, so it is always asked, even with one card (`FR-MOD-001` §1, §3, unchanged). The client may show the chooser which card a timeout would take, even though that reveals which card the opponent added last. | Drifting Concept; Mage Seeker's Inquisition (outside M2) |

- **Decisions (user, 2026-11-15):**
  - "go with alternative 1, one rule for all hands."
  - On hidden information: "Leaking a bit of info about which is the most recently drawn card is acceptable."
- **Cards affected:**
  - *Drifting Concept* (Animus unit, `data/cards/units-and-structures-animus.csv:19`): in M2 scope (wave H, on `cards.NotYetExecutable`), in neither starter deck.
  - *Mage Seeker's Inquisition* (Soulless maneuver, `data/cards/maneuvers-soulless.csv:3`): outside M2. A timeout takes the unit card most recently added to that hand. With no unit card in the hand, it discards nothing.
  - Checked and unchanged: *Screaming Palisade* ("Each opponent discards a card"). Each player chooses from their own hand, which `SVY-R16` already covers.
- **Alternatives rejected:**
  - The oldest card first: a second order for the same zone, and it leaks the same kind of information.
  - Order by printed attributes (highest Cost, then name): hides draw timing, but adds a new ordering scheme no other rule uses, for a case that almost never happens.

#### → `pm-frd-drafter` (FRD 01 and the ruling register)
- **Superseded wording:**
  - `docs/frd/01_Game_Rules_FRD.md:431` (`FR-MOD-001` §6): the bullet *"Choosing Cards from Your Own Hand (`SVY-R16`)"*.
  - `docs/frd/01_Game_Rules_FRD.md:86` (executive summary item 3): "a timeout selects the card most recently added" (still true, but it implies only your own hand).
- **FRD 01 changes (v3.26 → v3.27):**
  1. `FR-MOD-001` §6 (line 431): replace the bullet with: "*Choosing Cards from a Hand (`SVY-R16`, `SVY-R50`)*: Selects the card most recently added to that Hand, whoever is choosing. For *"Draw a card, then discard a card"* this is the card just drawn, including when the Hand was empty before the draw. When the chooser picks from an opponent's revealed hand (*Drifting Concept*, *Mage Seeker's Inquisition*, §3), it is the eligible card most recently added to the opponent's Hand. If no card is eligible, none is selected."
  2. Executive summary item 3 (line 86): replace "Choosing from your own Hand is not an inspection, so one card is selected automatically, and a timeout selects the card most recently added." with "Choosing from your own Hand is not an inspection, so one card is selected automatically. A timeout choosing from any Hand, your own or an opponent's revealed one, selects the card most recently added to it (`SVY-R50`)."
  3. Revision summary above v3.25 → v3.26, and bump the title to v3.27: `### Revision Summary (v3.26 -> v3.27): Timeouts When Choosing From a Revealed Hand (2026-11-15)` with the item `1. **One Timeout Rule for Every Hand (`SVY-R50`; `FR-MOD-001` §6)**: a timed-out choice from any Hand selects the card most recently added to that Hand, whoever is choosing. Card text: none.`
- **Register rows:**
  - Survey §4, add after R49: `| **R50** | What does a timeout pick when you choose from an opponent's revealed hand? (Open Question 23, before the *Drifting Concept* task) | Drifting Concept; Mage Seeker's Inquisition (outside M2) | **Ruled (user, 2026-11-15).** The card most recently added to that Hand, whoever is choosing; extends `SVY-R16`. Recorded in FRD 01 v3.27 as `SVY-R50` (`FR-MOD-001` §6). |`
  - Survey card row *Drifting Concept* (line 142), notes column: `Always asked, even with one card (§1, §3); timeout picks the card most recently added to the opponent's hand (`SVY-R50`)`.
  - Keyword interaction matrix: none.
- **Card text:** None.
- **Prompt to run:** `@pm-frd-drafter apply Design Outcome docs/design/outcomes/2026-11-15-wave-h-drifting-concept-revealed-hand-timeout.md`

#### → `lead-architect` (build board directive)
- **Directive row:**

  | # | Directive | Source | Status |
  | :---: | :--- | :--- | :--- |
  | 156 | **One timeout rule for every hand (`SVY-R50`; `FR-MOD-001` §1, §3, §6; design note Open Question 23).** In the *Drifting Concept* task: the choice from the opponent's revealed hand is a hand decision asked of *Drifting Concept*'s controller, and its options are cards in the **opponent's** hand. `handPicks` ([choice.go:188-204](../../../pkg/engine/choice.go)) ranks by the hand that holds the options (the hand's owner), not by `d.Player`, through the one ranking function: no second copy for opponents' hands. This choice is an inspection, so it never fast-paths, even with one card (§1, §3), while choosing from your own hand still does. Update `DecideHand`'s comment (`choice.go:18`, "your own hand") and `doc.go:36` to match. Tests: (a) the chooser times out on a 3-card opposing hand: the card most recently added to that hand is discarded, from its owner's hand, to its owner's discard pile; (b) a 1-card opposing hand is still asked; (c) an empty hand: nothing is asked or discarded; (d) the discard fires the owner's put-into-discard-from-hand trigger (*Eerie Remnant*). Mark design note Open Question 23 answered, citing this outcome. **Checked for contradictions:** none. On `autoloop/m2-wave-h`, `handPicks`' only caller is the `DecideHand` fallback, and every `DecideHand` decision comes from `chooseDiscards` ([zones.go:90](../../../pkg/engine/zones.go)), which always asks the hand's owner, so owner and decider are the same in every existing test (`git grep DecideHand` in `pkg`). *Drifting Concept* is only in the `cards.NotYetExecutable` entry and in neither starter deck, so the sim baseline doesn't move. | User via game-designer, 2026-11-15 ([outcome](../../../docs/design/outcomes/2026-11-15-wave-h-drifting-concept-revealed-hand-timeout.md)) | waiting: docs/design/outcomes/2026-11-15-wave-h-drifting-concept-revealed-hand-timeout.md applied (then open: before the *Drifting Concept* task) |

#### → `pm-lead` (portfolio)
- **Milestone or scope impact:** None.
