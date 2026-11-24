Status: applied: no FRD changes, 2026-11-23
Directive: #183
Rulings: none

### 🎲 Design Outcome: A pile card that leaves and returns is a new object (2026-11-23)

- **Question:** When a card leaves a discard pile and comes back while a stack item targets it or a trigger tracks it, does the item still find it? And does *Rage's Wake*'s tracking follow the same rule as M2's tracked death triggers, ending at a second move? Asked by the lead as M3 design note Open Question 5 (carried from M2 Open Question 10; `FR-ZON-004`, `FR-TGT-003`).
- **Rulings:** none new. The user confirmed FRD 01 as written:
  - `FR-TGT-003`: "If a target left its zone … that target becomes illegal". A target that left and returned did leave.
  - `FR-ZON-004` *Self-Relocating Trigger Tracking*: the trigger "fizzles (does nothing) if the card has already departed the Discard pile".
  - `FR-TRG-006` *Cross-Zone Entity Tracking*: *Rage's Wake* does nothing "if the card left the Discard pile before the delayed trigger resolves", and "Self-relocating death triggers on printed cards follow the identical invariant". So yes, it is the same convention as *Unceasing Tide*, *Stormfront of Ideals* and *Eerie Remnant*, and any move after the card reaches the pile ends the tracking, even if the card comes back.
- **Decisions (user, 2026-11-23):**
  - "1 - it works like overkill and unbowed do". Asked to choose between the two readings, the user picked "Effect loses X (option 1)": after a round trip the card is a new object, and the waiting effect does nothing to it. No FRD change.
- **Cards affected:** none can reach the case in M3 (checked 2026-11-23 against `data/cards/` and the four starter decks):
  - Declaring anything needs an empty stack (`pkg/engine/actions.go:115`), so *Leyline Regrowth*, *Inspired Salvage* (both in decks), *Mage Seeker's Inquisition* and *All-Out War* are always the bottom item.
  - Concordia's only triggers that take a card out of a pile are *Stardust Colossus* (to hand; not in a deck) and *Rage's Wake* (to Reserves; in deck). No Concordia trigger discards from hand, and *Samsara*'s trigger moves a card from the deck into the pile, not out. *Redlined Brawler*'s discard trigger is Soulless, and Soulless takes cards out of a pile only with a Maneuver.
  - Every pile effect acts on "your" pile, so an opponent's effect can't move the card.
  - *Rage's Wake* on *Stardust Colossus* killed by *Blind Defile*: Stardust's deploy trigger returns the card to hand, and nothing puts it back in the pile before the lower item resolves.
- **Alternatives rejected:**
  - Add a sentence to `FR-ZON-004` stating the rule for hands, decks and piles: the existing past-tense wording already covers it, and no M3 card reaches the case.
  - A card keeps its identity across pile moves (the kernel's current `*Instance` behavior): it contradicts the battlefield rule and FRD 01's "has departed" wording.

#### → `pm-frd-drafter` (FRD 01 and the ruling register)
- **Superseded wording:** None.
- **FRD 01 changes:** None.
- **Register rows:** None.
- **Card text:** None.

#### → `lead-architect` (build board directive)
- **Directive row:** #183 in [`.agents/tasks/m3/BOARD.md`](../../../.agents/tasks/m3/BOARD.md), status `waiting: the first card task whose card can move a card out of a discard pile and back while an item below it targets or tracks that card`.
  - **(1) Design note.** Mark M3 design note Open Question 5 answered: `FR-ZON-004`, `FR-TGT-003` and `FR-TRG-006` as written. Keep the per-wave re-check in each M3 survey row, using the reachability conditions above: a trigger that takes a card out of a pile, plus an effect that puts the same card back in its owner's pile, both resolving above a waiting item.
  - **(2) When the trigger fires:** give `Instance` a zone-change count. Record it with each pile target (`FR-TGT-003`), each tracked card (self-relocating death and enter-discard triggers, `FR-ZON-004`) and each delayed trigger's tracked card (*Rage's Wake*, `FR-TRG-006`), and compare it at resolution. A changed count means the card left: the target is illegal and the tracked effect does nothing. Pin it with `TestFR_TGT_003_PileTargetThatLeftAndReturnedIsIllegal`, `TestFR_ZON_004_PileTriggerDoesNothingIfItsCardLeftAndReturned` and `TestFR_TRG_006_DelayedTriggerDoesNothingIfItsCardLeftAndReturned`.
  - **Checked for contradictions:** none found. `TestFR_TGT_003_PileTargetThatLeftThePileIsIllegal` (`pkg/engine/pile_target_test.go:132`), `TestFR_ZON_004_PileTriggerFizzlesIfItsCardLeftThePile` (`pkg/engine/pile_trigger_test.go:295`) and `TestCard_AllshansBattleTrance_TargetLeftThePile` pin only the one-way case, which the ruling agrees with. No test name in `pkg/` pins a round trip as legal. The kernel's `*Instance` persistence is a known gap (M2 Open Question 10), not a pinned rule. `cardtown sim` output doesn't change.

#### → `pm-lead` (portfolio)
- **Milestone or scope impact:** None.
