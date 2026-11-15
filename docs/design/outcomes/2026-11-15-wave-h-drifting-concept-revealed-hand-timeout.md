Status: open (lead-architect, 2026-11-15)
Task / Directive: none yet, the *Drifting Concept* card task (wave H, design note Deliverable 6m) / none (branch: autoloop/m2-wave-h)
Cards: Drifting Concept

### ❓ Open Ruling Question: what a timeout picks from an opponent's revealed hand (2026-11-15)

- **Question:**
  1. When the player choosing a card from an **opponent's** revealed hand (*Drifting Concept*: "You choose a card from it") runs out of time (`FR-MOD-001` §6), which card is picked? `SVY-R16` covers only your **own** hand ("the card most recently added to the Hand"), and §6's fixed order covers assets and piles ("Cards in a pile such as the Discard pile: from the top down"). A hand has no top, and the chooser here isn't its owner.
- **Source:** wave H run (`autoloop/m2-wave-h`, session 3), before writing *Drifting Concept*'s card task. Design note Open Question 23.
- **Cards affected:**
  - *Drifting Concept* (Animus unit, `data/cards/units-and-structures-animus.csv:19`): "Ethereal / When this unit deploys, chosen opponent reveals their hand. You choose a card from it. That player discards that card." In M2 scope (wave H, on `cards.NotYetExecutable`), in neither starter deck.
  - *Mage Seeker's Inquisition* (Soulless maneuver, `data/cards/maneuvers-soulless.csv:3`): "Chosen opponent reveals their hand. You choose a unit card from it. That player discards that card." Outside M2, but the same ruling applies to it.
- **Current FRD 01 & engine state:**
  - Settled already, so not asked: "a card from it" is an untargeted selection made on resolution (`FR-TGT-001`, `SVY-R3`), and an empty hand discards nothing. A one-card hand is still asked, because §3 always grants the inspection window and §1's fast path excludes inspecting hidden cards. The card is discarded from its owner's hand, so its owner's "put into your discard pile from your hand" triggers fire (*Eerie Remnant*). In a two-player match "chosen opponent" has only one legal choice.
  - Not settled: the timeout. `FR-MOD-001` §6 (`docs/frd/01_Game_Rules_FRD.md:430-431`) gives fixed orders for assets, piles and your own hand only.
  - `pkg/engine/choice.go:18`: `DecideHand` is "choose cards from your own hand", and its fallback `handPicks` (`choice.go:188-204`) ranks options by the **deciding player's** hand. It would rank nothing correctly for an opponent's hand. `chooseDiscards` (`zones.go:84`) builds every hand-discard decision, always for the hand's owner.
- **Lead's recommendation:** use one timeout rule for every hand: a timeout picks the card most recently added to **that** hand, whoever is choosing. That extends `SVY-R16` without a second order, and in the engine `handPicks` would read the hand's owner instead of the decider. The other option, the card that has been in the hand longest, needs a second rule for the same zone.
