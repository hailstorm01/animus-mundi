Status: applied: FRD 01 v3.21, 2026-11-04
Directive: #105
Rulings: SVY-R36, SVY-R37, SVY-R38

### 🎲 Design Outcome: activating with no legal target, and a deck search's choice and timeout (2026-11-04)

- **Question:** Three questions from lead-architect on the wave D run (`autoloop/m2-wave-d`, directive #104): whether an activated ability can be activated with no legal entity for an exactly-one target slot, whether a search may choose no card, and what a search's timeout does.
- **Rulings:**

  | Ruling | Settles | Cards |
  | :--- | :--- | :--- |
  | `SVY-R36` | An activated ability can be activated only if its required targets exist, under the same conditions as a Maneuver | Sinuous Growth, Boneyard Scavenger, Ruin Partisans, Rhapsody's Hollow; Dolus Macrocannon, Cathedral of Self, Tempest City Ray (outside M2) |
  | `SVY-R37` | A search for "a card" must choose one if the deck has any. A search with criteria may choose no card, even when a match exists | Cliff of Portents; World Soul Messenger (outside M2) |
  | `SVY-R38` | A search shows the deck in deck order, top first, and a timeout takes the first valid card | Cliff of Portents; World Soul Messenger (outside M2) |

- **Decisions (user, 2026-11-04):**
  - `SVY-R36`: "no, you need a valid target for it to be activated in the first place."
  - `SVY-R37`: "cliff of portents has no restrictions on its search … If a card is worded 'search your deck for a unit card, reveal it', you can fail to find even if you have valid choices (following MTG's standard)". The user confirmed that an unrestricted search must take a card if the deck has one.
  - `SVY-R38`: "for all flavors of find, you get the first valid option - cliff of portents would be top card, something like 'search your deck for a unit' picks the first valid unit". The user chose to list the search window in deck order.
- **Cards affected:**
  - *Sinuous Growth*: M2, wave D, Animus starter deck (4 copies).
  - *Boneyard Scavenger*: M2, wave D, Path starter deck (2 copies).
  - *Cliff of Portents*: M2, wave D, Animus starter deck (1 copy).
  - *Ruin Partisans*: M2, wave D, not in a starter deck.
  - *Rhapsody's Hollow*: M2, not in a starter deck, deferred by #14 (b).
  - *Dolus Macrocannon*, *Cathedral of Self*, *Tempest City Ray*, *World Soul Messenger*: outside M2.
  - Checked and still activatable with no targets: *Encampment* and *Enclave* (both outside M2), whose "up to 2" slots never stop activation.
- **Alternatives rejected:**
  - Paying costs for an ability with no legal target: it lets a player sacrifice a unit for nothing, and the deliberate uses are narrow.
  - Letting an unrestricted search choose no card: activating is already optional, so this adds nothing.
  - A timeout that finds no card: the player would lose a card they paid for.
  - A sorted search window: players couldn't see which card a timeout takes, so `SVY-R16` would need an exception.

#### → `pm-frd-drafter` (FRD 01 and the ruling register)
- **Superseded wording:**
  - `FR-DEK-001` (FRD 01 line 158): "any card in the deck may be chosen" is now mandatory if the deck has a card, and "if no matching card exists in the deck, the search completes with no card chosen" no longer states the only way a restricted search finds nothing.
  - Deployment Triggers Without Legal Targets (line 381): "Unlike enacting a Maneuver (`SVY-R29`, `FR-MNV-001`)" must also name activation.
- **FRD 01 changes** (v3.20 → v3.21):
  1. **`FR-ACT-001`, new bullet after line 335:** "**An Activated Ability Needs Its Required Targets (`SVY-R36`, `FR-ACT-001`)**: An activated ability can be activated only if each of its exactly-one target slots has a legal entity when it is declared (`FR-TGT-001`), under the same conditions as a Maneuver (`SVY-R29`, `FR-MNV-001`). Slots that must differ need different entities (`SVY-R34`), a *"Choose N"* ability needs at least N legal modes, and an *"up to N"* slot never stops it. An ability that can't be activated pays nothing, and its asset stays where it is. For example, *Sinuous Growth* can't be activated unless its controller's discard pile holds a Fungus or Plant card, because its own card can't be chosen (`SVY-R2`). An ability with no target slot needs only its costs: *Cliff of Portents* can be activated with an empty deck, finds nothing and still shuffles. Once declared, the ability resolves under the usual two-stage verification (`FR-TGT-003`), and its costs are not refunded if a target becomes illegal."
  2. **Line 381:** replace "Unlike enacting a Maneuver (`SVY-R29`, `FR-MNV-001`)" with "Unlike enacting a Maneuver (`SVY-R29`, `FR-MNV-001`) or activating an ability (`SVY-R36`, `FR-ACT-001`)".
  3. **`FR-DEK-001`, line 158:** replace the bullet with: "**Unrestricted vs. Restricted Search (`SVY-R37`)**: If an effect instructs a player to search for *"a card"*, any card in the deck may be chosen, and the player must choose one if the deck has any. If filtering criteria are specified (e.g., *"an Elemental card"*), only cards matching those criteria may be chosen, and the player may choose no card even when a match exists. A search of a deck with no eligible card completes with no card chosen."
  4. **`FR-DEK-001`, new bullet after line 159:** "**Search Order & Timeout (`SVY-R38`)**: A search shows the remaining deck in deck order, top first. If the player doesn't answer in time, the search takes the first eligible card in that order: the top card for *"a card"*, or the first match from the top for a search with criteria (`FR-MOD-001` §6). The shuffle then follows as usual."
  5. **`FR-MOD-001` §6, new item after *Top-of-Deck Inspections* (line 394):** "*Deck Searches (`SVY-R38`)*: Selects the first eligible card in deck order, before the shuffle, which is the first eligible card the player sees listed. A search with criteria does this too, although the player could have chosen no card (`SVY-R37`). A *"you may search"* times out earlier, at the *"you may"*, so no search happens."
  6. **Revision summary, v3.21 block** (in the format of the v3.20 block):
     - "1. **An Activated Ability Needs Its Required Targets (`SVY-R36`; `FR-ACT-001`)**: an activated ability can be activated only if each of its exactly-one target slots has a legal entity, under the same conditions as a Maneuver."
     - "2. **Finding a Card (`SVY-R37`; `FR-DEK-001`)**: a search for *"a card"* must choose one if the deck has any, and a search with criteria may choose no card even when a match exists."
     - "3. **Search Order & Timeout (`SVY-R38`; `FR-DEK-001`, `FR-MOD-001` §6)**: a search shows the deck in deck order, and a timeout takes the first eligible card."
- **Register rows:**
  - Survey §4: add rows **R36**–**R38**, one per ruling above, with the Settles text and Cards from the Rulings table. Status cell, each: "**Ruled (user, 2026-11-04).** Recorded in FRD 01 v3.21 as `SVY-R3x` (`FR-…`)", with the requirement IDs from the summary block.
  - Survey Rulings line (line 18): add "R36–R38 on 2026-11-04" and extend the recorded range to `SVY-R38`.
  - Survey card rows: add R36 to *Sinuous Growth*, *Boneyard Scavenger*, *Ruin Partisans* and *Rhapsody's Hollow*, and R37 and R38 to *Cliff of Portents*.
  - Keyword matrix: no rows affected.
- **Card text:** None.
- **Prompt to run:** `@pm-frd-drafter apply Design Outcome docs/design/outcomes/2026-11-04-wave-d-activation-without-target-and-deck-search.md`

#### → `lead-architect` (build board directive)
- **Directive row** #105 in `.agents/tasks/m2/BOARD.md`, status `waiting: docs/design/outcomes/2026-11-04-wave-d-activation-without-target-and-deck-search.md applied (then open: before the first task whose cards: include Sinuous Growth, Boneyard Scavenger, Ruin Partisans, Rhapsody's Hollow or Cliff of Portents)`:

  **Wave D: activating with no legal target, and deck search (`SVY-R36`–`SVY-R38`; `FR-ACT-001`, `FR-DEK-001`, `FR-MOD-001` §6).** Answers #104, which exists only on `autoloop/m2-wave-d`. When that branch syncs, mark #104 "answered by #105".
  **(1) `SVY-R36`, a small shape task before the card tasks.** The `ActActivate` case of `check` (`pkg/engine/actions.go:195` on the run branch) calls `placeable` with the ability's effect instead of `costsPayable`, and `placeable`'s comment (`pkg/engine/placement.go:108`) stops saying it applies to Maneuvers only. Triggers keep placing whatever they find (`SVY-R30`, `FR-TGT-003`). Tests: (a) `TestFR_ACT_001_NoLegalTargetCannotActivate`: with a kill-self ability and an empty pile, the action isn't offered, `check` rejects it, nothing is paid and the unit stays. (b) `TestFR_ACT_001_UpToSlotStillActivates`.
  **(2) `SVY-R37`, in the *Cliff of Portents* task.** `OpSearchDeck` requires one pick when the deck has a card and asks even when there is one option (`FR-MOD-001` §1 excludes deck inspections from the fast path). An empty deck asks nothing and still shuffles. A search with criteria, which offers choosing no card, arrives with its first card (*World Soul Messenger*, outside M2). Tests: `TestFR_DEK_001_UnrestrictedSearchMustChoose` and `TestFR_DEK_001_EmptyDeckStillShuffles`.
  **(3) `SVY-R38`, in the *Cliff of Portents* task.** The search decision lists its options in deck order, top first, so the existing first-option timeout takes the top card before the shuffle. Test: `TestFR_MOD_001_SearchTimeoutTakesTopCard`.
  **Merged code that contradicts the rulings:** on `autoloop/m2-wave-d`, `check` uses only `costsPayable`, which contradicts `SVY-R36`. No test pins it: the header of `pkg/engine/activated_test.go` says every exactly-one slot there has a legal entity. Activated abilities aren't on `main` yet. `OpSearchDeck` doesn't exist on the run branch, so nothing contradicts `SVY-R37` or `SVY-R38`. The design note's Actions paragraph (`docs/architecture/m2-ability-model.md:53`) already plans `SVY-R36`.

#### → `pm-lead` (portfolio)
- **Milestone or scope impact:** None.
