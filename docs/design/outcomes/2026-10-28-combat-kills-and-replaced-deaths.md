Status: approved (user, 2026-10-28)
Directive: #73
Rulings: SVY-R19; SVY-R12 (wording)
Audit to close: none

### 🎲 Design Outcome: Combat kills and replaced deaths (2026-10-28)

- **Question:** The user flagged that FRD 01 v3.14 says Overkill "triggers from the same event" as combat-kill triggers (`FR-CBT-002`). Yet `FR-KWD-002p` says it fires even when a replacement keeps the defender alive, when there is no combat-kill event. Checking that raised a second question, reachable in M2: does "attacks and kills" fire when a replacement saves the defender? The kernel fires it when the defender is lethal, before the death sweep ([combat.go:114](../../../pkg/engine/combat.go)). FRD 01 requires the defender to be destroyed, and `SVY-R15` says triggers watching for a replaced event do not fire.
- **Rulings:**

  | Ruling | Settles | Cards |
  | :--- | :--- | :--- |
  | `SVY-R19` | "Attacks and kills" does not trigger when a replacement effect keeps the defender from being destroyed. | Eternal Watchman, Shifting Impossibility; later Dread Manifestation |
  | `SVY-R12` (wording) | Overkill depends only on damage: at least 1 more than lethal, whether or not the defender dies. It is not a combat-kill trigger. | Tempest Surge, Redline Infusion |

- **Decisions (user, 2026-10-28):**
  - `SVY-R12` wording: "overkill should be based on dealing at least 1 more than lethal damage, not 'lethal + 1 AND it died'."
  - `SVY-R19`: "agreed on Watchman, since it's worded as 'attacks and kills' it doesn't get its return trigger if shifting impossibility saves the scratch company".
- **Cards affected:**
  - *Eternal Watchman*: M2, the Path starter Headquarters.
  - *Shifting Impossibility*: M2, in the Animus starter deck.
  - *Dread Manifestation*: Concordia, M3.
  - *Tempest Surge* and *Redline Infusion*: M3. Their text is unchanged.
- **Alternatives rejected:**
  - Lethal combat damage counts as a kill: it contradicts the card wording and `SVY-R15`.
  - Overkill requires the defender to die: the user ruled it depends on damage only.

#### → `pm-frd-drafter` (FRD 01 and the ruling register)
- **Superseded wording:** FRD 01 v3.14, `FR-CBT-002` "Post-Attack & Combat-Kill Triggers" (line 521), last sentence: "**Overkill** (`FR-KWD-002p`) triggers from the same event when the defender's Current Health is below 0."
- **FRD 01 changes (v3.14 → v3.15):**
  - **`FR-CBT-002`, line 521:** replace that sentence with:
    > If a replacement effect keeps the defender from being destroyed, it was not killed, and combat-kill abilities do not trigger (`SVY-R19`, `SVY-R15`). **Overkill** (`FR-KWD-002p`) is not a combat-kill trigger. It triggers from the attacker's Step 3 combat damage when that damage leaves the defender's Current Health below 0, whether or not the defender is then destroyed.
  - **`FR-KWD-002p`:** unchanged.
  - **Revision summary:** `### Revision Summary (v3.14 -> v3.15): Combat Kills and Replaced Deaths (2026-10-28)`
    1. **A Replaced Death Is Not a Combat Kill (`SVY-R19`; `FR-CBT-002`, `FR-REP-001`)**: "attacks and kills" does not trigger when a replacement effect saves the defender. *Eternal Watchman* attacking a Scratch Company unit that *Shifting Impossibility* returns to hand gets no card back.
    2. **Overkill Does Not Depend on the Kill (`SVY-R12`; `FR-CBT-002`, `FR-KWD-002p`)**: a wording fix. Overkill triggers from the attacker's combat damage, not from the combat-kill event.
  - **Title line:** `Master FRD v3.14` becomes `Master FRD v3.15`.
- **Register rows:**
  - **Survey §4, new row R19:**
    - Question: does "attacks and kills" trigger when a replacement effect keeps the defender from being destroyed?
    - Cards: Eternal Watchman with Shifting Impossibility; later Dread Manifestation.
    - Status: **Ruled (user, 2026-10-28): no.** The defender was not killed (`SVY-R15`). Recorded in FRD 01 v3.15 as `SVY-R19` (`FR-CBT-002`).
  - **Survey line 18:** add "R19 on 2026-10-28". Recorded as `SVY-R1`–`SVY-R19`.
- **Card text:** None.
- **Audit to close:** None.
- **Prompt to run:** `@pm-frd-drafter apply Design Outcome docs/design/outcomes/2026-10-28-combat-kills-and-replaced-deaths.md`

#### → `lead-architect` (build board directive)
- **Directive row:**

  | 73 | **A replaced death is not a combat kill, and Overkill does not need one (`SVY-R19`; `SVY-R12` wording; `FR-CBT-002`).** `resolveAttack` fires the combat-kill hook when the defender `IsLethal`, before the death sweep ([combat.go:114](../../../pkg/engine/combat.go)). Under `SVY-R19` it fires only if the death sweep after the Attack actually destroys the defender. A destruction replacement that saves it (*Shifting Impossibility* returning a Scratch Company unit to hand) means no trigger. Reachable with the M2 starter decks: Path's *Eternal Watchman* attacking an Animus Scratch Company unit while *Shifting Impossibility* is in play. Keep what T-0080 pinned: an attacker destroyed in the same sweep still fires, and the trigger is placed in APNAP order with the defender's death triggers (`TestFR_TRG_003_CombatKillTriggerAndTheDefendersDeathTriggerAreAPNAP`). How the kernel knows the death was replaced is the lead's call. One small task, tier `sonnet`. `rules-test-author` pins `TestFR_CBT_002_CombatKillTriggerDoesNotFireWhenTheDeathIsReplaced`: *Eternal Watchman* attacks and deals lethal damage to an opposing Scratch Company unit while its controller has *Shifting Impossibility*; the unit returns to its owner's hand, Impossibility takes 1 damage, Watchman's trigger does not fire, and no card returns from the discard pile. FRD 01 already requires a destroyed defender, so this does not wait for v3.15.<br>**Overkill (M3, corrects #72 and design note Open Question 12):** Overkill is not "a trigger on `HookCombatKill` with one more gate". It is its own hook, fired from Step 3 when the attacker's combat damage leaves the defender's Current Health below 0, so it still fires when the death is replaced (`_FiresWhenTheDefendersDeathIsReplaced`). Update Open Question 12 to match.<br>Contradicting merged code: `combat.go:114`. No test pins the replaced case: I checked the test names in `combat_kill_test.go`, and none of them sets up a replacement. | User via game-designer, 2026-10-28 (`docs/design/outcomes/2026-10-28-combat-kills-and-replaced-deaths.md`) | open |

#### → `pm-lead` (portfolio)
- **Milestone or scope impact:** None. It is a kernel fix within M2's existing done-when.
- **Prompt to run:** None.
