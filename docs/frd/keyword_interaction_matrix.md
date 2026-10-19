# Keyword Interaction Matrix

Companion to the [Game Rules FRD](01_Game_Rules_FRD.md). It answers one bounded question: **for every pair of keywords, and for each keyword against the core rules, does FRD 01 say what happens?**

Unlike an open-ended red-team pass, this document can be finished. There are 15 keywords, so there are 105 pairs plus 15 self-stacking cases. When every cell has a status and every open item has a ruling, the keyword surface is complete. A new keyword adds one row and one column, nothing more.

## How to use it

- **Status codes:**
  - `·` no interaction (the keywords touch independent parts of the rules).
  - `D` defined by FRD 01 (citation in the Defined Interactions list).
  - `I` inferred: FRD 01 is silent; the kernel or a reasonable reading picks a behavior that still needs a ruling.
  - `?` undefined: FRD 01 is silent or self-contradictory and plausible readings give different game outcomes.
- **Reachability:** `R` means a legal deck can produce the situation against some opponent with today's cards. `U` means it cannot (yet).
- **Closing an item:** a designer ruling goes into FRD 01, and where the keyword is kernel-executable, a `TestFR_*` scenario test in `pkg/engine` pins it. Then the cell flips to `D`. Physical playtest rulings close items the same way.
- **The kernel:** wherever an `I` or `?` item affects executable rules, `pkg/engine` picks a behavior and marks it with an `OPEN(KIM-...)` comment, so changing the ruling means changing one marked spot.
- **For `pm-devils-advocate`:** keyword findings are raised against this matrix (a cell that is wrong or missing), not as fresh prose findings.

## Matrix

Abbreviations: Unb = Unbowed (`FR-KWD-002a`), Res = Resistant N (`FR-KWD-002b`), Stw = Stalwart (`FR-KWD-002c`), Heal = Heal N (`FR-KWD-002d`), Cor = Cornered (`FR-KWD-002e`), Agg = Aggressive N (`FR-KWD-002f`), Eth = Ethereal (`FR-KWD-002g`), Ward = Ward (`FR-KWD-002h`), Ver = Versatile (`FR-KWD-002i`), Grd = Guardian (`FR-KWD-002j`), Arm = Armor N (`FR-KWD-002k`), DEx = Deploys Exhausted (`FR-KWD-002l`), Scr = Scratch N (`FR-KWD-002m`), Rgp = Regroup N (`FR-KWD-002n`), Bsk = Berserk N (`FR-KWD-002o`). The diagonal is self-stacking (two instances of the same keyword on one asset).

|     | Unb | Res | Stw | Heal | Cor | Agg | Eth | Ward | Ver | Grd | Arm | DEx | Scr | Rgp | Bsk |
| :-- | :-: | :-: | :-: | :-: | :-: | :-: | :-: | :-: | :-: | :-: | :-: | :-: | :-: | :-: | :-: |
| **Unb**  | · | D | D | D | D, I C01 | · | · | · | · | · | D | · | D | D | D |
| **Res**  |   | D | D | · | D | · | · | · | · | · | D | · | · | · | D |
| **Stw**  |   |   | · | · | · | · | · | · | · | · | D | · | · | · | D |
| **Heal** |   |   |   | · | · | · | · | · | · | · | D | · | · | · | · |
| **Cor**  |   |   |   |   | · | D | D | · | · | · | · | · | D | · | D |
| **Agg**  |   |   |   |   |   | · | · | · | · | · | · | · | · | D | D |
| **Eth**  |   |   |   |   |   |   | · | · | · | D | · | · | · | · | D |
| **Ward** |   |   |   |   |   |   |   | · | · | D | · | · | · | · | D |
| **Ver**  |   |   |   |   |   |   |   |   | · | · | · | · | · | · | · |
| **Grd**  |   |   |   |   |   |   |   |   |   | · | · | D | · | · | D |
| **Arm**  |   |   |   |   |   |   |   |   |   |   | D | · | · | · | D |
| **DEx**  |   |   |   |   |   |   |   |   |   |   |   | · | · | D | · |
| **Scr**  |   |   |   |   |   |   |   |   |   |   |   |   | · | D | · |
| **Rgp**  |   |   |   |   |   |   |   |   |   |   |   |   |   | · | · |
| **Bsk**  |   |   |   |   |   |   |   |   |   |   |   |   |   |   | ? B07 |

**Tally (120 cells):** 32 defined, 1 undefined, 87 no interaction. One defined cell (Unb × Cor) also carries an inferred sub-case, KIM-C01. All keyword-vs-core-rules items are ruled. The two remaining open items are unreachable with current cards. Closed items are listed under Rulings.

## Open Items (need a ruling)

Ordered by reachability, then by how much the outcome changes.

| ID | Status | Reach | Question | Kernel today |
| :-- | :-: | :-: | :-- | :-- |
| KIM-C01 | I | U | **Cornered-gated triggers inside a simultaneous step.** Cornered triggers are gated "at the moment the trigger event occurs"; during a simultaneous step Cornered is locked at the start of the step (`FR-LAY-003`). A damage-taken trigger behind a Cornered gate fires mid-step: is it gated by the locked value? No current card has a Cornered-gated damage trigger. | n/a |
| KIM-B07 | ? | U | **Berserk stacking.** Printed Berserk plus granted Berserk: one secondary with summed N, or two secondaries? Only *Tempest Surge* (Concordia) and *Redline Infusion* (Soulless) carry Berserk, so no legal deck has both. | n/a |

## Rulings

Closed items, with the designer ruling, where it lives in FRD 01, and the scenario test that pins it.

| ID | Ruling (2026-10-19) | FRD 01 | Pinned by |
| :-- | :-- | :-- | :-- |
| KIM-B01 | A **Ward** asset cannot be the Berserk secondary: the secondary follows normal targeting rules. | `FR-KWD-002o` | `TestFR_KWD_002o_SecondaryCannotBeWarded`, `TestFR_KWD_002o_SecondaryThatGainsWardBeforeStep3Dissipates` |
| KIM-B02 | An **Ethereal** asset may be the secondary; Ethereal restricts only attacks. | `FR-KWD-002o` | `TestFR_KWD_002o_SecondaryMayBeEthereal` |
| KIM-B03 | **Guardian** forces the primary target only; the secondary may be any other legal target. The primary is excluded from secondary eligibility. | `FR-KWD-002o`, `FR-CBT-002` | `TestFR_KWD_002o_GuardianDoesNotRestrictSecondary`, `TestFR_KWD_002o_SecondaryExcludesPrimary` |
| KIM-B04 | Printed Berserk always has N. *Tempest Surge* and *Redline Infusion* read **Berserk 2** for now. | Card data | `TestLoadCardsFromRepositoryData` |
| KIM-S01 | A new face-down Scratch Company enters **Exhausted**. | `FR-ZON-005`, `FR-SET-004`, `FR-KWD-002m` | `TestFR_ZON_005_ScratchCompanyEntersExhausted` |
| KIM-S02 | HQ defeat is checked at the next state-based death sweep (the next time no item is resolving). Both HQs at ≤ 0 at the same sweep is a Match Draw. | `FR-CBT-003` | `TestFR_CBT_003_HeadquartersDefeatIsCheckedAtTheDeathSweep`, `TestFR_CBT_003_BothHeadquartersLethalAtOneSweepIsMatchDraw` |
| KIM-B05 | The Berserk secondary is **mandatory** whenever a legal secondary target exists. | `FR-KWD-002o`, `FR-CBT-002` | `TestFR_KWD_002o_SecondaryIsMandatoryWhenALegalOneExists`, `TestFR_KWD_002o_NoSecondaryRequiredWithoutALegalTarget` |
| KIM-J01 | Once an attack is on the stack with its target, later eligibility changes (a Guardian arriving, the defender gaining Ethereal) neither redirect nor fizzle it, much as granting "unblockable" after blockers are declared doesn't unblock a creature in MTG. | `FR-CBT-002` | `TestFR_CBT_002_CommittedAttackIgnoresLaterGuardianAndEthereal` |
| KIM-B06 | If the attack fizzles because of the attacker (killed, removed, or moved out of a legal combat zone), Berserk cleave still happens from its last-known characteristics. If the defender is gone, there is no cleave, whatever happened to the attacker. | `FR-CBT-002`, `FR-KWD-002o` | `TestFR_KWD_002o_CleaveSurvivesAnAttackerSideFizzle`, `TestFR_KWD_002o_NoCleaveWhenTheDefenderIsGone` |
| KIM-S03 | Exhausting for a zero yield is legal; it leaves room for generation-replacement and exhaust-watching designs. | `FR-CAP-001` (Design Rationale KIM-S03) | `TestFR_CAP_001_ZeroYieldGenerationIsLegal` |
| KIM-S04 | Empty-deck loss uses the same timing as HQ defeat: the next death sweep. Any two losses at one sweep, by HQ or by deck, are a Match Draw. | `FR-WIN-001`, `FR-CBT-003`, `FR-LAY-002` | `TestFR_WIN_001_EmptyDeckLossWaitsForTheDeathSweep`, `TestFR_CBT_003_HeadquartersDefeatAndEmptyDeckAtOneSweepIsMatchDraw` |

## Defined Interactions

Each `D` cell above, with where FRD 01 settles it. A test marker means a `pkg/engine` scenario test pins it.

- **Unb × Res:** Resistant reducing damage to 0 means Unbowed does not trigger (`FR-KWD-002a`, `FR-REP-004` Stage 5).
- **Unb × Stw:** damage capped to 1 still counts as taken, so Unbowed triggers (`FR-REP-004` Stage 5).
- **Unb × Heal:** "Unbowed: Heal 2" on a lethally damaged asset fizzles after it is destroyed (`FR-KWD-002a`, *Macabre Dancers*).
- **Unb × Cor:** *Macabre Dancers*' continuous "Cornered: Resistant 1" is locked for the whole simultaneous combat step (`FR-LAY-003`). See KIM-C01 for gated triggers.
- **Unb × Arm:** damage absorbed by Armor counts as taken (`FR-KWD-002k`, `FR-KWD-002a`).
- **Unb × Scr:** "Unbowed: Scratch 1" still resolves after the asset is destroyed, since it does not modify the asset (`FR-KWD-002a`, *Hammerhand Brigands*).
- **Unb × Rgp:** a Regroup clause aimed at the destroyed asset itself fizzles; one choosing other allied assets resolves (`FR-KWD-002a`, `FR-TGT-003`).
- **Unb × Bsk:** Berserk cleave that reaches Armor or Health on the secondary triggers its Unbowed; retaliation on a Berserk attacker triggers the attacker's (`FR-REP-004`, `FR-KWD-002o`).
- **Res × Res:** multiple Resistant values sum ("the target's total Resistant N", `FR-REP-004` Stage 2).
- **Res × Stw:** Resistant (Stage 2) applies before Stalwart (Stage 3), so Resistant 1 + Stalwart takes 0 from a 1-power attack and 1 from a 5-power attack (`FR-REP-004`). Test: `TestFR_REP_004_ResistantAppliesBeforeStalwart`.
- **Res × Cor:** a Cornered-gated Resistant aura (*Ancestral Monument*) is locked for a simultaneous step and re-evaluated after the death sweep (`FR-LAY-003`).
- **Res × Arm:** Resistant reduces damage before Armor absorbs it (`FR-REP-004`).
- **Res × Bsk:** Resistant applies to Berserk's non-combat cleave (`FR-REP-004` Stage 2). Test: `TestFR_KWD_002o_BerserkCleaveIsNonCombatDamage`.
- **Stw × Arm:** Stalwart caps combat damage to 1, then Armor absorbs that 1 (`FR-REP-004` Stage 4). Test: `TestFR_REP_004_StalwartCappedDamageIsAbsorbedByArmor`.
- **Stw × Bsk:** Berserk cleave is non-combat damage and bypasses Stalwart (`FR-REP-004` Stage 3, `FR-KWD-002o`). Test: `TestFR_KWD_002o_BerserkCleaveIsNonCombatDamage`.
- **Heal × Arm:** Heal removes accumulated damage only; it never restores Armor (`FR-KWD-002d`, `FR-LAY-002`).
- **Cor × Agg:** "Cornered: Whenever this unit attacks, ..." is gated when the attack is declared (`FR-LAY-003` triggered-ability rule).
- **Cor × Eth:** Ethereal assets count toward Main Theater asset totals; Cornered has no exclusion (`FR-LAY-003`).
- **Cor × Scr:** face-down Scratch Companies are Units, so they count once in the Main Theater (`FR-ZON-005`, `FR-LAY-003`).
- **Cor × Bsk:** a secondary killed by cleave changes Main Theater counts only after the Step 3 death sweep (`FR-LAY-003`).
- **Agg × Agg (marked `·`):** two Aggressive instances give the same total power whether treated as one trigger or two, so no ruling is needed unless a card starts counting triggers.
- **Agg × Rgp:** each attack triggers Aggressive again and the modifiers last until end of turn, so a regrouped attacker grows with each attack (`FR-KWD-002f`, `FR-KWD-002n`). Reachable: *Furnace Golem* plus *Enclave*. **Balance note, not a rules hole.**
- **Agg × Bsk:** Aggressive raises Combat Power; Berserk deals flat N regardless of power (`FR-KWD-002o`).
- **Eth × Grd:** an Ethereal Guardian cannot be attacked and does not restrict attackers (`FR-CBT-001`, `FR-KWD-002g`). Test: `TestFR_KWD_002g_EtherealCannotBeAttackedAndOverridesGuardian`.
- **Eth × Bsk:** an Ethereal asset may be the Berserk secondary target; Ethereal restricts only attacks (`FR-KWD-002o`, ruling KIM-B02). Test: `TestFR_KWD_002o_SecondaryMayBeEthereal`.
- **Ward × Grd:** a Ward Guardian can still be attacked, so it still restricts attackers (`FR-KWD-002h`, `FR-KWD-002j`).
- **Ward × Bsk:** a Ward asset cannot be chosen as the Berserk secondary, and a secondary that gains Ward before Step 3 makes the cleave dissipate (`FR-KWD-002o`, `FR-TGT-003`, ruling KIM-B01). Tests: `TestFR_KWD_002o_SecondaryCannotBeWarded`, `TestFR_KWD_002o_SecondaryThatGainsWardBeforeStep3Dissipates`.
- **Grd × Bsk:** Guardian forces the primary target only; the secondary may be any other legal target (`FR-KWD-002o`, ruling KIM-B03). Test: `TestFR_KWD_002o_GuardianDoesNotRestrictSecondary`.
- **Grd × DEx:** Guardian has no readiness condition; an Exhausted Guardian still forces attacks onto it (`FR-CBT-001`). Worth confirming against prototype habit.
- **Arm × Arm:** Armor stacks additively and never expires (`FR-LAY-002`).
- **Arm × Bsk:** Armor absorbs Berserk cleave (`FR-REP-004` Stage 4).
- **DEx × Rgp:** Regroup readies any exhausted allied asset, including one that just deployed exhausted (`FR-KWD-002n`).
- **Scr × Rgp:** "Regroup all Scratch Company units in your Reserves" readies them all (`FR-KWD-002n`).
