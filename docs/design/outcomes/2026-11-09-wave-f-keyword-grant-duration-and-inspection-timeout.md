Status: open (lead-architect, 2026-11-09)
Task / Directive: none written yet / none (branch: main)
Cards: Allshan's Battle Trance; Devilhorn Scouting Party

### ❓ Open Ruling Question: how long a keyword grant lasts, and a timed-out look at the top cards (2026-11-09)

- **Question:**
  1. **How long a keyword grant lasts.** *Allshan's Battle Trance* says *"It gains Stalwart and gets +2 power until end of turn."* Does "until end of turn" cover Stalwart too?
     - (a) Yes. The unit has Stalwart and +2 power until Sub-Phase 4c, and both expire then.
     - (b) No. Only the +2 power expires. The unit keeps Stalwart while it stays on the battlefield, and loses it only when it leaves (`FR-ZON-004`).
  2. **A look at the top cards that times out.** *Devilhorn Scouting Party* says *"Look at the top 3 cards of your deck. You may deploy a unit card with Requisition cost 3 or less from among them for free. Put the rest on the bottom of your deck in any order."* If the player doesn't answer that one prompt (`FR-MOD-001` §4) in time:
     - (a) **The "you may":** does the timeout deploy the first eligible unit card from the top, as §6's top-of-deck fallback says (*"selects the first eligible card(s) in order"*)? Or does it deploy nothing, as a *"you may search"* times out at the "you may" and doesn't search?
     - (b) **The order on the bottom:** in what order do the rest go to the bottom? (i) In the order they were in, so the higher of two cards stays higher. (ii) Reversed, as if they were put on the bottom one at a time from the top.
- **Source:** lead-architect's readiness check of wave F, 2026-11-09, before the first wave F task is written. Asked now so a wave F run doesn't stop at either card (D7.6).
- **Cards affected** (`data/cards/maneuvers-path.csv`; both M2, wave F, Path **starter deck**):
  - *Allshan's Battle Trance* (Path Maneuver, T4, 1 copy): *"Deploy chosen unit card from your discard pile to the Main Theater. It gains Stalwart and gets +2 power until end of turn."* Question 1.
  - *Devilhorn Scouting Party* (Path Maneuver, T2, 2 copies): text above. Question 2.
  - Outside M2, for the same readings: *Field Mechanic* (Soulless Unit): *"chosen allied Mechanized unit gains 2 armor and gets +2 power until end of turn"*. Armor is a persistent pool and "never expires at end of turn" (`FR-LAY-002`, `FR-KWD-002k`), so on that card "until end of turn" can only cover the power. *Light of the World Soul* (Concordia Maneuver) looks at the top 7, may deploy up to 2 for free and put up to 1 into hand, and puts the rest on the bottom "in any order", so question 2 applies to it as well.
- **Current FRD 01 & engine state** (FRD 01 v3.21; `pkg/engine` on `main` at `2ff3bb2`):
  - Question 1: `FR-TRN-003a` 2 (line 351) quotes the card text only to show that an effect's own zone wins over the printed Deployment Zones. Layers 3 and 5 (lines 455, 461) and the snapshot-locking rule (line 471) speak of "Layer 3 keyword grants … lasting *until end of turn*", and Sub-Phase 4c (line 290) expires every such modifier. None of them says how far one duration phrase reaches across "gains X and gets Y". The engine has no keyword grant yet. Its only temporary modifier is `Asset.TempPower` (`pkg/engine/state.go:76`), cleared in 4c (`pkg/engine/turn.go:36`), and `OpModifyStat` grants power only (`pkg/engine/ability/ability.go:271`). Board directive #137 will make that one list of modifiers before this card, and the list can hold a grant with either duration, so the ruling decides only the card's data and its tests.
  - Question 2 (a): §4 (line 398) makes the whole look one choice prompt. §6's top-of-deck fallback (line 401) selects the first eligible card(s) in order, and it doesn't say whether that applies when the card says "you may". §6's deck-search fallback says a *"you may search"* times out at the "you may", so no search happens. The two readings disagree for *Scouting Party*.
  - Question 2 (b): the same fallback "places the remainder on the bottom of the deck" and gives no order. There is no look-at-top-cards op yet, so nothing in the engine or its tests settles either part.
- **Lead's recommendation:**
  1. (a). One duration phrase at the end covers every temporary clause before it, which is also MTG's reading. On *Field Mechanic* it still covers only the power, because armor can't expire. A permanent grant from a one-shot Maneuver would be the only one of its kind in M2.
  2. (a): the first eligible unit is deployed. The "may" is part of the one §4 prompt, not a question asked before it, so §6's top-of-deck fallback applies as written, and a timeout still uses the card. (b) (i): the rest keep their order. The player then knows the bottom order from the top order they saw, which keeps every timeout predictable (`SVY-R16`).
