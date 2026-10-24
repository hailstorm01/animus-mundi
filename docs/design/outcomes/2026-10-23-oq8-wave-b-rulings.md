Status: applied: 652717e
Directive: #47
Rulings: SVY-R14, SVY-R15, SVY-R16, SVY-R17

### 🎲 Design Outcome: Open Question 8, wave B rulings SVY-R14 to SVY-R17 (2026-10-23)

Re-issued 2026-10-23 for `pm-frd-drafter`, which runs in a separate session. It replaces the chat-only report "Open Question 8, three wave B rulings". The decisions are the same, and the FRD 01 wording is now complete. The board directive for this outcome is #47 in `.agents/tasks/m2/BOARD.md` (commit `5e81a65`, status `waiting: FRD 01 v3.11`).

- **Question:** Design note `docs/architecture/m2-ability-model.md`, Open Question 8. For Infested Apostle (a), Chainbreaker Irregulars (b) and Animoceps Shambler (c): when are their choices made and their triggers placed, and what happens on a timeout?
- **Ruling count:** Open Question 8 asked three questions. They produced four rulings:

  | Ruling | Comes from | Cards |
  | :--- | :--- | :--- |
  | `SVY-R14` | Question (a), plus the Arkha knock-on (same "you may X. If you do, Y" wording) | Infested Apostle, Arkha, Sun's Fury |
  | `SVY-R15` | Knock-on of (a): Infested Apostle killing a Scratch Company unit while Shifting Impossibility is out | Infested Apostle with Shifting Impossibility; the kill costs of Veil Render, Rhapsody's Embrace and Feed the Despoiler |
  | `SVY-R16` | Question (b), widened by the user from Irregulars' discard to every timeout default | Chainbreaker Irregulars; every mandatory target and untargeted pick |
  | `SVY-R17` | Question (c) | Animoceps Shambler |

  Narrowing `SVY-R3`'s wording is a consequence of `SVY-R14`. It is not a new ruling.

- **Decisions (user, 2026-10-23), in the user's own terms:**
  - **`SVY-R14`:** "Self-kill is intentional. Apostle deploys, puts its deploy trigger on the stack, the trigger resolves and the player picks a unit to kill, which can include the Apostle itself. The damage is tied to the trigger's 'if you do', not the actual deployed apostle." On a timeout while picking the unit to kill: "it's an optional trigger, and I believe default behavior is 'no' for optionals." Arkha: "Arkha follows."
  - **`SVY-R15`:** "'If you do' … I'm leaning towards 'I said I was killing this thing, not my fault it got replaced, my damage still happens'." On additional costs (Veil Render, Feed the Despoiler): "from a player perspective you paid the cost (killed a unit) to commit the deploy to the stack and then after that something happened."
  - **`SVY-R16`:** Agreed that one card in hand is discarded without a question. "If your hand is empty when the trigger happens, you draw a card then discard it again." On the timeout: "discarding the one you just drew is fine. Randomness for default choices is something we should avoid." Asked whether that is a general rule: "I think so, knowing what'll happen if you time out a choice feels like a more consistent UX than 'oh it happened to do the thing that I wanted'." The user approved the fixed order in change 5 below.
  - **`SVY-R17`:** "'Deploys exhausted' means your starting state is exhausted as opposed to readied. No transition happened." On placement: "agreed on both counts. You pay a cost (exhaust) to put march / attack on the stack, which causes the shambler ability to trigger and be put on top of it. Generate doesn't use the stack, so it doesn't have a chance to trigger until after generate finished."
  - The user approved the narrower `SVY-R3` wording (change 1).

- **Ruling IDs:** new `SVY-R14`, `SVY-R15`, `SVY-R16` and `SVY-R17`; `SVY-R3`'s wording narrowed (its ruling is unchanged).
- **Cards affected** (all in M2 scope):
  - Infested Apostle (Animus deck ×4)
  - Arkha, Sun's Fury (Path deck ×1)
  - Chainbreaker Irregulars (Path deck ×4)
  - Animoceps Shambler (Animus deck ×3)
  - Shifting Impossibility (Animus deck ×2, interaction only)
  - Veil Render (Animus deck ×1), Rhapsody's Embrace (Animus deck ×2) and Feed the Despoiler (no deck): their kill costs (`SVY-R15`)
  - Animoceps Spores (Animus deck ×3) and Mutation Swell (no deck): timeout of their untargeted picks (`SVY-R16`)
  - Out of scope: Redlined Brawler (Soulless, M3) follows `SVY-R14` when it is built.
- **Alternatives rejected:**
  - *Kill on placement, as an optional trigger cost* (the lead's recommendation in Open Question 8(a)): the user wants the kill picked on resolution, from the units on the battlefield at that moment.
  - *The Apostle may not kill itself:* self-kill is intended.
  - *A replaced kill means "you didn't"*, so no damage: the player carried out the instruction, and the replacement isn't their doing.
  - *Random timeout defaults* (today's `FR-MOD-001` §6 rule for mandatory targets): players should be able to predict what a timeout does.
  - *Entering the battlefield exhausted counts as exhausting:* the asset starts Exhausted, so there is no change from Readied.

#### → `pm-frd-drafter` (FRD 01 and the ruling register)

**Wording that this outcome supersedes.** The cost reading of Apostle and Arkha appears only in the places listed here. No FRD 01 revision summary implies it. `SVY-R9` (v3.9 summary item 1 and `FR-MOD-001` §7) says only that "you may" is answered on placement, before targets are chosen. That still holds and needs no change.
- Survey line 30, wave B shapes: "Optional non-resource trigger costs." Change it (register edits below).
- Survey line 105 (Infested Apostle) and line 146 (Arkha: "Optional `MOVE_CARD_ZONE` cost"). Change them (register edits below).
- FRD 01 `FR-TRG-005` (line 369): "optional non-resource actions (such as … *Blind Defile*)". Its title is "Trigger Costs Prohibition" (change 7).
- `pm-frd-drafter` does not edit these:
  - the design note, line 31 (`Trigger` has "an optional non-resource `Cost` (`FR-TRG-005`)") and Open Question 8. The lead updates these under directive #47.
  - FRD 02 benchmark #9 (`optionalCosts`), which is frozen design intent.

**FRD 01 changes (v3.10 → v3.11).**

1. **`FR-TGT-001`, the bullet "Targets Are Locked on Placement (`SVY-R3`)" (line 307).** Replace the sentences *"Card text never makes a targeted clause depend on the outcome of an earlier clause of the same effect. Where that dependency is intended, the later clause is an untargeted selection (below), or the earlier clause is an additional cost so that its outcome is known on declaration (`FR-TRN-003a`)."* with:
   > An earlier clause of the same effect may decide **whether** a targeted clause applies (see Upfront Target Selection for Conditional Triggers above, and *"You May X. If You Do, Y"* below), but never **which** entities or **how many** it targets. Where the number or identity of targets is meant to depend on an earlier clause, that clause is an additional cost so that its outcome is known on declaration (`FR-TRN-003a`), or the later clause is an untargeted selection (below).

   The examples that follow (Animoceps Spores, Steel Rain) stay as they are.

2. **`FR-TGT-001`, new bullet after "Untargeted Selections (`SVY-R3`)":**
   > - **"You May X. If You Do, Y" (`SVY-R14`)**: For a triggered ability of the form *"you may [X]. If you do, [Y]"* (*Infested Apostle*: *"When this unit deploys, you may kill an allied unit. If you do, deal 2 damage to chosen opposing unit."*; *Arkha, Sun's Fury*), the controller answers *"you may"* as the trigger is placed (`SVY-R9`, `FR-MOD-001` §7), and Y's targets are chosen at the same moment (`SVY-R3`). X is carried out when the trigger resolves, as an untargeted selection among the entities that match it at that moment. X is not a cost (`FR-TRG-005`). Y applies only if X was carried out (`SVY-R15`). If nothing matches X on resolution, or the controller does not answer X's selection in time (`FR-MOD-001` §6), X is not carried out and Y does nothing. The trigger's own source may be selected when it matches X: *Infested Apostle* may kill itself, and its trigger still resolves. Assets destroyed by X are destroyed while the trigger resolves, so their death triggers are placed after it has finished (`FR-TRG-004`), and Y resolves before them.

3. **`FR-REP-001` through `FR-REP-003`, new bullet after "Consequential Event Isolation (`FR-REP-001`)" (line 425):**
   > - **Replaced Actions Still Count as Done (`SVY-R15`, `FR-REP-001`)**: An action that a card instructs a player to take, or that a player pays as a cost, counts as carried out once the player has carried it out, even if a replacement effect changes or completely replaces the event it would cause. Completely replacing an event still means the original event never occurs, and triggers that watch for it do not fire (Inline Event Substitution, above). What still counts is the player's action, for *"if you do"* (`SVY-R14`) and for costs (`FR-TRN-003a`). Example: *Infested Apostle* kills an allied Scratch Company unit, and *Shifting Impossibility* returns that unit to its owner's hand instead. The unit is not destroyed and no death trigger fires, but the kill was carried out, so the Apostle deals its 2 damage.

4. **`FR-TRN-003a`, the bullet "Additional Sacrifice, Discard & Exhaust Costs" (line 288).** Add at the end:
   > A cost that a replacement effect changes or replaces still counts as paid (`SVY-R15`, `FR-REP-001`): if *Shifting Impossibility* returns a Scratch Company unit killed to pay *Rhapsody's Embrace*'s additional cost, the Maneuver still resolves.

5. **`FR-MOD-001`, §1 and §6 (lines 330 and 335–340).**
   - **§1 Single-Option Fast-Path.** Add at the end:
     > Choosing a card from your own Hand (*Chainbreaker Irregulars*: *"discard a card"*) is not an inspection of hidden cards. If only one card qualifies, it is selected automatically (`SVY-R16`).
   - **§6 Deterministic Fallbacks on Timeout.** Add after the opening sentence:
     > No fallback is random, so a player can always predict what a timeout does (`SVY-R16`).
   - **§6, replace the bullet** *"Mandatory Multi-Option Targets: Automatically selects a random legal target."* with:
     > - *Mandatory Targets and Untargeted Selections (`SVY-R16`)*: Selects the required number of legal options, the first ones in **fixed order**. Assets: the choosing player's own assets first, then the opponent's. Within each player, the Headquarters first, then Reserves, then the Main Theater, each zone in the order its assets entered it (as `SVY-R7`). Cards in a pile such as the Discard pile: from the top down. An *"up to N"* selection selects none. The selection for X in *"you may X. If you do, Y"* is optional and also selects none, so X is not carried out (`SVY-R14`).
   - **§6, new bullet:**
     > - *Choosing Cards from Your Own Hand (`SVY-R16`)*: Selects the card most recently added to the Hand. For *"Draw a card, then discard a card"* this is the card just drawn, including when the Hand was empty before the draw.

   **Other "random" wording in FRD 01:** only line 339 (the bullet replaced above) is a timeout default. Line 89 (choosing who decides the first player) and line 115 (shuffling) are intentionally random and stay unchanged.

6. **`FR-CAP-001`, new bullet after the intro paragraph (line 256), before "March":**
   > - **What Counts as Exhausting (`SVY-R17`)**: An asset *exhausts* when it changes from **Readied** to **Exhausted** for any reason: March, Attack, Deploy Scratch Company, generating resources, an exhaust or additional cost (`FR-ACT-001`, `FR-TRN-003a`), or an effect. An asset that enters the battlefield Exhausted (**Deploys Exhausted**, `FR-KWD-002l`; a face-down **Scratch Company**, `FR-ZON-005`) has not exhausted: it never was Readied. Triggers caused by exhausting are placed as follows:
   >   - *Stack-pushed actions* (March, Attack, Deploy Scratch Company): the asset exhausts as the action's cost, so the triggers are placed **on top of** the action and resolve before it (as `FR-CBT` Step 1 already says for attacks).
   >   - *Generating resources*: generation is off-stack (`FR-QUE-001`), so the triggers are placed after the generation has completed, and resolve before the player takes another action.
   >   - *Costs*: as `FR-TRN-003a`, on top of the card or ability paid for.
   >   - *Effects*: buffered and placed after the resolving item finishes (`FR-TRG-004`).

   Line 487 (Combat Step 1, *"abilities triggered by the attacker exhausting … enqueue … on top of the Attack action"*) already agrees, and needs no change.

7. **`FR-TRG-005` (line 369).** Replace the bullet with:
   > - Triggered abilities never require paying Requisition or Tactics resources to trigger or resolve. They may require game-state conditions (such as **Cornered**) or offer optional non-resource actions (*"you may kill this structure"* on *Blind Defile*; *"you may kill an allied unit"* on *Infested Apostle*). Such an action is not a cost: it is carried out when the trigger resolves (`SVY-R14`).

**Revision summary**, a new block above "Revision Summary (v3.9 -> v3.10)":
```markdown
### Revision Summary (v3.10 -> v3.11): "If You Do" Actions, Replaced Actions, Predictable Timeouts and Exhausting (2026-10-23)
Rulings made by the user on 2026-10-23 for the wave B cards *Infested Apostle*, *Arkha, Sun's Fury*, *Chainbreaker Irregulars* and *Animoceps Shambler* (design note Open Question 8):
1. **"You May X. If You Do, Y" (`SVY-R14`; `FR-TGT-001`, `FR-TRG-005`, `FR-MOD-001`)**: "you may" is answered and Y's targets are chosen on placement. X is an untargeted selection carried out on resolution, not a cost, and Y applies only if X was carried out. A trigger's source may select itself: *Infested Apostle* may kill itself.
2. **Replaced Actions Still Count as Done (`SVY-R15`; `FR-REP-001`, `FR-TRN-003a`)**: an action or cost the player carried out counts as done even when a replacement effect changes or replaces its event. *Infested Apostle*'s kill replaced by *Shifting Impossibility* still deals its damage.
3. **No Random Timeout Fallbacks (`SVY-R16`; `FR-MOD-001`)**: mandatory targets and untargeted selections time out to the first legal options in a fixed order, not random ones. Choosing from your own Hand is not an inspection, so one card is selected automatically, and a timeout selects the card most recently added.
4. **What Counts as Exhausting (`SVY-R17`; `FR-CAP-001`)**: every change from Readied to Exhausted counts, and entering the battlefield Exhausted does not. Triggers from an action's exhaust cost go on top of the action, and triggers from resource generation are placed after it completes.
5. **Narrower Wording for Locked Targets (`SVY-R3`; `FR-TGT-001`)**: an earlier clause may decide whether a targeted clause applies, but never which entities or how many it targets. The ruling is unchanged.
```
Also update the document title line from `Master FRD v3.10` to `Master FRD v3.11`.

**Register edits** (`docs/design/m2_starter_set_ability_survey.md`):

- **§1 Summary, the "Rulings" bullet (line 18).** Replace *"R8–R11 on 2026-10-21 and R13 on 2026-10-23; … They are recorded in FRD 01 as `SVY-R1`–`SVY-R11` and `SVY-R13`."* with *"R8–R11 on 2026-10-21, and R13–R17 on 2026-10-23; … They are recorded in FRD 01 as `SVY-R1`–`SVY-R11` and `SVY-R13`–`SVY-R17`."* Leave the rest of the bullet as it is.
- **§4, new rows after R13:**

| ID | Question | Cards | Status |
| :--- | :--- | :--- | :--- |
| **R14** | "You may X. If you do, Y": when is X carried out, and may X select the trigger's own source? (design note Open Question 8(a)) | Infested Apostle, Arkha, Sun's Fury; later Redlined Brawler (M3) | **Ruled (user, 2026-10-23): X is carried out on resolution.** "You may" is answered and Y's target locked on placement (`SVY-R9`, `SVY-R3`). X is an untargeted selection on resolution, not a cost, and the Apostle may kill itself. Y applies only if X was carried out. With no match or a timeout on X, Y does nothing. This overrides the lead's recommendation (a kill on placement, as a cost). Recorded in FRD 01 v3.11 as `SVY-R14` (`FR-TGT-001`, `FR-TRG-005`, `FR-MOD-001`). |
| **R15** | Does an action or cost still count as done when a replacement effect replaces its event? (knock-on of Open Question 8(a)) | Infested Apostle with Shifting Impossibility; the kill costs of Veil Render, Rhapsody's Embrace and Feed the Despoiler | **Ruled (user, 2026-10-23): yes.** The player carried out the action, so "if you do" is met and a cost stays paid, even though the replaced event never occurs and triggers watching for it do not fire. Recorded in FRD 01 v3.11 as `SVY-R15` (`FR-REP-001`, `FR-TRN-003a`). |
| **R16** | How do the fast path and the timeout work for choosing a card from your own hand, and may any timeout fallback be random? (Open Question 8(b)) | Chainbreaker Irregulars; every mandatory target and untargeted selection (Animoceps Spores, Mutation Swell, Infested Apostle's kill) | **Ruled (user, 2026-10-23): no fallback is random.** Mandatory targets and untargeted selections time out to the first legal options in a fixed order: the choosing player's assets first, then Headquarters, Reserves, Main Theater by arrival, and piles from the top. Your own Hand is not an inspection: one card is selected automatically, and a timeout selects the card most recently added (for Irregulars, the card just drawn). Replaces the random mandatory-target fallback. Recorded in FRD 01 v3.11 as `SVY-R16` (`FR-MOD-001`). |
| **R17** | What counts as "this unit exhausts", and when are its triggers placed? (Open Question 8(c)) | Animoceps Shambler | **Ruled (user, 2026-10-23): every change from Readied to Exhausted, for any reason.** Entering the battlefield Exhausted is not one. Triggers from March, Attack or Deploy Scratch Company are placed on top of the action. Triggers from resource generation are placed after it completes. Recorded in FRD 01 v3.11 as `SVY-R17` (`FR-CAP-001`). |

- **R3 row (line 61).** Add at the end of its status cell: *"Wording narrowed 2026-10-23 (`SVY-R14`): an earlier clause may decide whether a targeted clause applies, but not which or how many targets."*
- **Wave B row (line 30).** In "Shapes added", replace "Optional non-resource trigger costs." with "'You may X. If you do, Y' actions on resolution (`SVY-R14`)." In "FRD 01 rules reached", add `FR-CAP-001` if it is not listed.
- **Card rows:**

| Line | Card | Column | Old | New |
| :--- | :--- | :--- | :--- | :--- |
| 103 | Animoceps Shambler | Notes | Triggers from off-stack generation (KIM-S03 rationale) | Every Readied→Exhausted change; generation triggers are placed after it (`SVY-R17`, KIM-S03 rationale) |
| 105 | Infested Apostle ★ | Shape | deploy (may, kill allied unit) → `DEAL_DAMAGE 2` chosen opposing unit | deploy (may) → `KILL` an allied unit, self allowed (untargeted, on resolution); if done → `DEAL_DAMAGE 2` chosen opposing unit |
| 105 | Infested Apostle ★ | Notes | Benchmark #9 | Benchmark #9, but its `optionalCosts` is superseded: the kill is on resolution (`SVY-R14`) |
| 116 | Animoceps Spores | Notes | (as now) | add: "timeout picks the first in fixed order (`SVY-R16`)" |
| 126 | Mutation Swell ★ | Notes | (as now) | add: "timeout picks the first in fixed order (`SVY-R16`)" |
| 129 | Shifting Impossibility | Notes | R5 | R5; a replaced kill still counts as done (`SVY-R15`) |
| 142 | Chainbreaker Irregulars | Notes | Discard can trigger Eerie Remnant | Discard can trigger Eerie Remnant; one card fast-paths, timeout discards the card just drawn (`SVY-R16`) |
| 146 | Arkha, Sun's Fury | Shape | attack (may, put a unit card from discard on bottom of deck) → `DEAL_DAMAGE 3` chosen opposing HQ | attack (may) → `MOVE_CARD_ZONE` a unit card from discard to deck bottom (untargeted, on resolution); if done → `DEAL_DAMAGE 3` chosen opposing HQ |
| 146 | Arkha, Sun's Fury | Notes | Optional `MOVE_CARD_ZONE` cost | `SVY-R14`: not a cost |

  The Rules column for lines 105 and 146 also gets `FR-TGT-001` (`SVY-R14`) next to `FR-TRG-005`.
  Lines 108 (Feed the Despoiler), 113 (Rhapsody's Embrace) and 125 (Veil Render) may optionally add "replaced kill cost stays paid (`SVY-R15`)" to their notes.

**Card text: none.** The current `data/cards/*.csv` text already fits every ruling. An untargeted selection reads "a" or "an", and a target reads "chosen" (`SVY-R3`):
- Infested Apostle: "When this unit deploys, you may kill an allied unit. If you do, deal 2 damage to chosen opposing unit." "An allied unit" is the untargeted X, and "chosen opposing unit" is the locked target Y.
- Arkha, Sun's Fury: "Whenever this unit attacks, you may put a unit card from your discard pile on the bottom of your deck. If you do, deal 3 damage to chosen opposing Headquarters." "A unit card" is the untargeted X.
- Chainbreaker Irregulars: "Unbowed: Draw a card, then discard a card." "A card" is untargeted.
- Animoceps Shambler: "Whenever this unit exhausts, chosen allied asset in the Main Theater heals 1." The target is locked on placement.

**Audit to close:** none. This outcome comes from a design note Open Question, not a DA audit.

**Prompt to run:** `@pm-frd-drafter apply Design Outcome "Open Question 8, wave B rulings SVY-R14 to SVY-R17" (2026-10-23) from <path to this file>`

#### → `lead-architect` (build board directive)

**Already recorded** as directive #47 in `.agents/tasks/m2/BOARD.md` (commit `5e81a65`, status `waiting: FRD 01 v3.11`). The user approved its text on 2026-10-23, and this re-issue doesn't change it. It covers:
- the resolution-time X step for the Apostle and Arkha, with no trigger-cost shape;
- the named future cases for replaced kills;
- the fixed-order `DecideTarget` fallback, replacing `TestFR_MOD_001_TargetFallbackIsRandomNotFixed` (`pkg/engine/choice_test.go:261`) as an authorized edit;
- the own-hand fast path and fallback;
- the exhaust hook and where its triggers are placed;
- marking design note Open Question 8 answered.

When the lead applies it, the lead also corrects the design note's `Trigger` line (line 31), which describes "an optional non-resource `Cost` (`FR-TRG-005`)". After `SVY-R14` that cost becomes a resolution-time action.

#### → `pm-lead` (portfolio)

- **Milestone or scope impact:** none. No card moves milestone, and the M2 done-when is unchanged. The no-random-fallback rule changes FRD 01 and existing kernel fallback code, but that work is part of directive #47 within M2.
- **Prompt to run:** none.
