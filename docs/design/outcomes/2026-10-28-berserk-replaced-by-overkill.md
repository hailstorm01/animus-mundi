Status: applied: FRD 01 v3.14, 2026-10-28
Directive: #72
Rulings: SVY-R12, KIM-O01
Audit to close: none

### 🎲 Design Outcome: Berserk is replaced by Overkill (2026-10-28)

- **Question:** Survey `R12` (opened 2026-10-21) asked whether Berserk N should be simplified or redesigned. Its secondary target needed repeated special rulings (`KIM-B01`–`KIM-B07`, `SVY-R8`). The user's code review (board directive #68, 2026-10-28) had already ordered the kernel plumbing removed pending this decision.
- **Rulings:**

  | Ruling | Settles | Cards |
  | :--- | :--- | :--- |
  | `SVY-R12` | Berserk N (`FR-KWD-002o`) is retired and every trace removed. The new keyword **Overkill** (`FR-KWD-002p`) triggers when an attacker's combat damage strictly exceeds what was needed to kill the defender. The **excess** is how far below 0 the defender's Current Health went. Exactly lethal is not overkill. It still fires when a replacement keeps the defender alive. | Tempest Surge, Redline Infusion |
  | `KIM-O01` | Two Overkill abilities on one unit (replaces `KIM-B07`): each triggers separately, with the same excess. | none yet |

- **Decisions (user, 2026-10-28):**
  - Remove Berserk: "we need to remove all trace of berserk from the FRDs and come up with something new. It's too problematic in its current form." The physical prototype dropped it: "I took it out because of how weird it was."
  - New keyword, overkill: "it's better as a 'when this unit overkills something, do this thing'" / "new keyword, it's still something I want to capture".
  - Exactly lethal: "strictly exceed, 'exactly kill' isn't 'overkill'".
  - Defender saved by a replacement: "yes, still fires".
  - *Tempest Surge* target: (b) "chosen opposing unit".
  - *Redline Infusion* granting "Overkill: Ready this unit.": "I think it's worth a shot".
  - Scope: "exception to remove them from frozen FRDs + TDD. Strike out and supersede is fine" (for revision history).
  - `KIM-O01`: "each triggers separately, not as one big trigger."
- **Cards affected:** *Tempest Surge* (Concordia) and *Redline Infusion* (Soulless). Both are M3, and neither is in an M2 starter deck. No M2 card has Berserk or Overkill.
- **Alternatives rejected:**
  - A plain "whenever this unit attacks, deal 2 damage" trigger with no keyword: the user wants overkill captured as a keyword.
  - An untargeted splash on attack: it doesn't capture overkill, and it's a board wipe on a 2-cost card.
  - A second target that can hit allied units on *Tempest Surge*: the user chose an opposing-only target.
  - Counting exactly lethal damage as overkill: "'exactly kill' isn't 'overkill'".

#### → `pm-frd-drafter` (FRD 01 and the ruling register)

**Run this only after board directive #68 parts (a) and (b) have merged.** Until then, merged tests are named after `FR-KWD-002o` and check that *Tempest Surge* parses "Berserk 2".

- **Superseded wording** (line numbers as of `fdedc9f`):
  - FRD 01 v3.13:
    - Revision history, lines 41, 42, 62, 64, 69 and 70: strike through each Berserk item and append *"Superseded in v3.14 (`SVY-R12`): Berserk was replaced by Overkill (`FR-KWD-002p`)."*
    - Line 196 (`SVY-R10` example).
    - Line 278 (`FR-CAP-001` Attack).
    - Lines 507, 513, 518 and 519 (`FR-CBT-002`).
    - Line 546 (`FR-KWD-002o`).
    - Line 609 (traceability row).
  - Keyword matrix:
    - Line 21 (abbreviations), line 39 (grid row) and the grid column header.
    - Lines 50 and 58–68 (`KIM-B01`–`KIM-B07`).
    - Cross-keyword notes at lines 83, 88, 90, 95, 98, 100, 102, 103 and 106.
  - Survey: lines 18, 67 (R8), 69 (R10) and 71 (R12).
  - **Frozen documents (user exception, 2026-10-28: remove only, add no Overkill text):**
    - FRD 02: lines 297–298 and 337.
    - FRD 03: line 175.
    - FRD 04: lines 21, 27, 512, 515–516, 536, 565, 625–627, 632 and 1073. Drop `TARGET_ROLE_SECONDARY_BERSERK_TARGET` and `is_secondary_cleave`. An attack binds exactly one target.
    - FRD 05: lines 81, 541, 551, 554, 558, 698, 701 and 702.
    - TDD 03: lines 200, 659, 1212, 1269 and 1598.
  - Left alone as history: ADR-0001 line 10, `docs/frd/README.md` line 20, and the task files.
- **FRD 01 changes (v3.13 → v3.14):**
  - **`FR-KWD-002o`:** replace the whole entry with: `- **FR-KWD-002o: Berserk N**: *Retired in v3.14 (`SVY-R12`); superseded by `FR-KWD-002p` (Overkill).*`
  - **New `FR-KWD-002p`**, after it:
    > - **FR-KWD-002p: Overkill**: Triggered combat ability (`"Overkill: [Effect]."`). Triggers when this unit is the declared attacker and its combat damage in Step 3 of `FR-CBT-002` leaves the defender with Current Health below 0 (`FR-LAY-002`). Exactly lethal damage (Current Health 0) is not overkill. The **excess** is how far below 0 the defender's Current Health is after Step 3; it is fixed when the trigger fires. Damage reduced by **Resistant** or capped by **Stalwart** (Stages 2–3) or absorbed by **Armor** (Stage 4) never reaches Health, so it never counts toward the excess, and an asset with Stalwart can never be overkilled (`FR-REP-004`). Overkill still triggers if a destruction replacement (`FR-REP-001`) keeps the defender on the battlefield, and if this unit is destroyed in the same death sweep. Like any trigger it is placed after that sweep (`FR-TRG-004`), in APNAP order with death triggers from the same event (`FR-TRG-003`). Retaliation damage never causes Overkill. Damage dealt by an Overkill effect is non-combat effect damage (`FR-REP-004`), so it cannot cause another Overkill. Two Overkill abilities on one unit trigger separately, each with the same excess (`KIM-O01`). An Overkill effect that resolves after this unit has left the battlefield uses its last-known characteristics (`FR-ZON-004`, `SVY-R10`). A Headquarters defender left lethal ends the match at the death sweep, so its Overkill never resolves.
  - **`FR-CBT-002`:**
    - Line 507: delete "; plus a secondary opposing Main Theater target … `FR-KWD-002o`".
    - Line 513: delete the two sentences beginning "**Berserk N** cleave still happens…".
    - Line 518: delete the whole "Berserk N Secondary Cleave Damage" sub-step.
    - Line 519: append: "**Overkill** (`FR-KWD-002p`) triggers from the same event when the defender's Current Health is below 0."
  - **`FR-CAP-001`, line 278:**
    - Delete ", plus a secondary opposing Main Theater target if it has **Berserk N** (`FR-KWD-002o`)".
    - Replace "Both are targets: they are re-verified" with "The target is re-verified".
  - **`SVY-R10` example, line 196:** replace "as Berserk cleave does after an attacker-side fizzle (`FR-CBT-002`)" with "(`FR-ZON-004`)".
  - **`SVY-R8`:** in the requirement text, delete "and the Berserk secondary", "with no cleave" and the `KIM-B01` Ward clause. The revision history keeps them struck through, as above.
  - **Traceability row, line 609:** `002a`–`002o` becomes `002a`–`002p`, and the description adds "(`002o` retired)".
  - **Revision summary line:** `### Revision Summary (v3.13 -> v3.14): Berserk Is Replaced by Overkill (2026-10-28)`, with item 1: **Overkill Replaces Berserk (`SVY-R12`, `KIM-O01`; `FR-KWD-002o`, `FR-KWD-002p`, `FR-CBT-002`, `FR-CAP-001`)**: Berserk N and its secondary target are retired. Overkill triggers when an attacker's combat damage leaves the defender below 0 Current Health, and its effect may use the excess.
- **Register rows:**
  - **Survey R12:** set the status to **Ruled (user, 2026-10-28): Berserk is retired and replaced by Overkill (`FR-KWD-002p`).** Recorded in FRD 01 v3.14 as `SVY-R12`.
  - **Survey R8:** append *"The Berserk clause is superseded by `SVY-R12`."*
  - **Survey R10:** replace "as Berserk cleave already does on an attacker-side fizzle" with "(`FR-ZON-004`)".
  - **Survey line 18:** R12 is ruled.
  - **Matrix `KIM-B01`–`KIM-B07`:** strike through, with the status cell *Superseded by `SVY-R12` (Berserk retired, 2026-10-28)*.
  - **Matrix abbreviation, grid row and column:** `Bsk` becomes `Ovk = Overkill (FR-KWD-002p)`. The diagonal cell reads `KIM-O01`.
  - **New matrix row `KIM-O01`:** two Overkill abilities on one unit trigger separately, each with the same excess. Settled by `FR-KWD-002p`, test `TestFR_KWD_002p_TwoInstancesTriggerSeparately` (M3).
  - **Cross-keyword notes:** replace the nine `× Bsk` notes with these:
    - **Stw × Ovk:** a Stalwart defender can never be overkilled. Overkill's own effect damage is non-combat, so it bypasses Stalwart.
    - **Arm × Ovk:** Armor absorbs combat damage before the excess is measured.
    - **Res × Ovk:** Resistant reduces combat damage before the excess, and it also reduces Overkill effect damage.
    - **Agg × Ovk:** Aggressive's power resolves before Step 3, so it raises the excess.
    - **Unb × Ovk:** an overkilled Unbowed defender still fires its Unbowed (`SVY-R13`), in the same APNAP window.
    - **Cor × Ovk:** *"Cornered: Overkill: …"* is checked when the trigger event occurs (`FR-KWD-002e`, `SVY-R18`).
    - **Ward × Ovk:** an Overkill effect's target follows normal targeting rules (`FR-TGT-003`).
- **Card text:**

  `data/cards/units-and-structures-concordia.csv`, old:
  ```
  "Tempest Surge","UNIT","CONCORDIA","Elemental, Tempest","2","0","MAIN_THEATER","Berserk 2
  T2: This unit gets +2 power until end of turn.","4","1","0","0","UNCOMMON","false"
  ```
  New:
  ```
  "Tempest Surge","UNIT","CONCORDIA","Elemental, Tempest","2","0","MAIN_THEATER","Overkill: Deal damage equal to the excess to chosen opposing unit in the Main Theater.
  T2: This unit gets +2 power until end of turn.","4","1","0","0","UNCOMMON","false"
  ```
  `data/cards/maneuvers-soulless.csv`, old:
  ```
  "Redline Infusion","MANEUVER","SOULLESS","0","2","Chosen allied unit gets +2 power and gains Berserk 2 until end of turn. At the end of your turn, kill that unit.","COMMON","false"
  ```
  New:
  ```
  "Redline Infusion","MANEUVER","SOULLESS","0","2","Chosen allied unit gets +2 power and gains ""Overkill: Ready this unit."" until end of turn. At the end of your turn, kill that unit.","COMMON","false"
  ```
  *Redline Infusion* is the first card to grant a quoted triggered ability. Run `cardtown lint syntax` and report whether it accepts the pattern.
- **Audit to close:** None.
- **Prompt to run (after #68 parts (a) and (b) merge):** `@pm-frd-drafter apply Design Outcome docs/design/outcomes/2026-10-28-berserk-replaced-by-overkill.md`

#### → `lead-architect` (build board directive)
- **Directive row:**

  | 72 | **Berserk is replaced by Overkill (`SVY-R12`; `FR-KWD-002o` retired, `FR-KWD-002p` new).** Survey R12, which #68 waits on, is ruled. #68 proceeds as written, with three changes. (1) `TestLoadCardsFromRepositoryData`: drop the Tempest Surge assertion ([playout_test.go:93-96](../../../pkg/sim/playout_test.go)) rather than rewording it to "Berserk 2 is unsupported", because the card text becomes "Overkill: …". #68 acceptance (d) (both cards load and are not `Supported()`) already holds for either text. (2) Merge parts (a) and (b) before the user runs `pm-frd-drafter` on the outcome, because the FRD retires `FR-KWD-002o` and changes the card CSVs; record the merge on this row so the user knows. (3) No #68 test is rewritten into an Overkill test.<br>**Overkill is M3 work, with no M2 kernel task.** When the M3 card tasks for Tempest Surge and Redline Infusion are planned, it is a trigger on the existing `HookCombatKill` (T-0080) with one more gate, the defender's Current Health below 0, plus an "excess" value fixed when the hook fires. Redline Infusion also needs a new shape: a granted quoted triggered ability until end of turn. Settling tests for that work: `TestFR_KWD_002p_ExactlyLethalIsNotOverkill`, `_ExcessIsHealthBelowZero`, `_ArmorAbsorbsBeforeTheExcess`, `_StalwartDefenderIsNeverOverkilled`, `_FiresWhenTheDefendersDeathIsReplaced`, `_FiresWhenTheAttackerDiesInTheSameSweep`, `_RetaliationNeverOverkills`, `_EffectDamageIsNonCombatAndDoesNotChain`, `_TwoInstancesTriggerSeparately` (`KIM-O01`). Record them as an M3 item in the design note.<br>Contradicting merged code: all of it is Berserk code that #68 already removes. Nothing else contradicts the ruling (searched `pkg`, `cmd` and `tools` for Berserk, cleave and `Secondary`). The FRD-ownership and fences linter tests keep "Berserk" as FRD text fixtures, as #68 says. | User via game-designer, 2026-10-28 (`docs/design/outcomes/2026-10-28-berserk-replaced-by-overkill.md`) | open: with #68 (wave B.2) |

#### → `pm-lead` (portfolio)
- **Milestone or scope impact:**
  - **M3 keyword list:** Berserk N is replaced by Overkill (`FR-KWD-002p`). M3 gains a shape for granting a quoted triggered ability until end of turn (*Redline Infusion*).
  - **`PRODUCT_SUMMARY.md`:** update lines 88, 109, 118, 161, 162, 170, 172, 177, 251 and 276. R12 is ruled, and the "Before Berserk cards" gate goes. The historical notes (lines 162 and 170) name the past contradiction neutrally or point to `SVY-R12`.
  - No M2 change. The M2 target of 59 of 62 is unchanged, since `FR-KWD-002o` was never counted as executable.
- **Prompt to run:** `@pm-lead Runbook A: Berserk retired for Overkill (SVY-R12); M3 keyword list and PRODUCT_SUMMARY updated, new M3 shape for granted triggered abilities`

> **Revision (2026-10-28):** The `FR-CBT-002` sentence this outcome added ("triggers from the same event") and #72's "a trigger on `HookCombatKill` with one more gate" are corrected by [2026-10-28-combat-kills-and-replaced-deaths.md](2026-10-28-combat-kills-and-replaced-deaths.md) (directive #73). Overkill depends only on damage.
