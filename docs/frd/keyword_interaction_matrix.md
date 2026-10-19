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
| **Eth**  |   |   |   |   |   |   | · | · | · | D | · | · | · | · | ? B02 |
| **Ward** |   |   |   |   |   |   |   | · | · | D | · | · | · | · | ? B01 |
| **Ver**  |   |   |   |   |   |   |   |   | · | · | · | · | · | · | · |
| **Grd**  |   |   |   |   |   |   |   |   |   | · | · | D | · | · | ? B03 |
| **Arm**  |   |   |   |   |   |   |   |   |   |   | D | · | · | · | D |
| **DEx**  |   |   |   |   |   |   |   |   |   |   |   | · | · | D | · |
| **Scr**  |   |   |   |   |   |   |   |   |   |   |   |   | · | D | · |
| **Rgp**  |   |   |   |   |   |   |   |   |   |   |   |   |   | · | · |
| **Bsk**  |   |   |   |   |   |   |   |   |   |   |   |   |   |   | ? B07 |

**Tally (120 cells):** 29 defined, 4 undefined, 87 no interaction. One defined cell (Unb × Cor) also carries an inferred sub-case, KIM-C01. The keyword-vs-core-rules items add 3 more undefined cases (KIM-B04, KIM-B05, KIM-J01) and 4 more inferred ones (KIM-B06, KIM-S01, KIM-S02, KIM-S03).

## Open Items (need a ruling)

Ordered by reachability, then by how much the outcome changes.

| ID | Status | Reach | Question | Kernel today |
| :-- | :-: | :-: | :-- | :-- |
| KIM-B04 | ? | R | **Printed "Berserk" has no N.** *Tempest Surge* prints `Berserk` and *Redline Infusion* grants `Berserk`, but `FR-KWD-002o` only defines **Berserk N**. What is N for these cards, or was the physical-prototype Berserk a different mechanic (for example trample)? | Bare `Berserk` is not executable; *Tempest Surge* loads as unsupported. |
| KIM-B05 | ? | R | **Is the secondary target optional?** `FR-CBT-002` Step 1 says "an optional secondary opposing Main Theater target"; `FR-KWD-002o` says it is "selected upfront upon attack declaration if another legal opposing Main Theater asset exists", which reads as mandatory. Matters when every other opposing asset is one you would rather not damage (for example one with a damage-taken trigger such as **Unbowed**). | Optional. |
| KIM-B01 | ? | R | **Can a Ward asset be the Berserk secondary?** **Ward** forbids opponents choosing it "except to attack". The secondary is chosen by the opponent as part of an attack declaration, but it is not the attack target and takes non-combat damage. (*Tempest Surge* vs Soulless Ward units, or anything under *Comms Relay*.) | Allowed. |
| KIM-B02 | ? | R | **Can an Ethereal asset be the Berserk secondary?** **Ethereal** says it "cannot be attacked"; the secondary is not attacked, but the damage only exists because of an attack. (*Tempest Surge* vs *Flickering Memory* / *Drifting Concept*.) | Allowed. |
| KIM-B03 | ? | R | **Does Guardian restrict the secondary?** With a Guardian in the Main Theater the primary must be a Guardian. May the secondary be a non-Guardian asset? (*Tempest Surge* vs *Improvised Bastion*.) | Yes, any other opposing Main Theater asset. |
| KIM-J01 | ? | R | **Combat eligibility changing between declaration and Step 3.** `FR-CBT-002` Step 3 fizzles the attack only if a combatant was destroyed, removed, or moved out of a legal combat zone. `FR-TGT-003` re-verifies targets against "dynamic stat/filter criteria" before resolution. If a Guardian enters the Main Theater, or the defender gains Ethereal, during the Step 2 trigger window, does the attack fizzle? | Does not fizzle (Step 3 survival/zone check only). |
| KIM-B06 | I | R | **Berserk when the attack fizzles.** If the primary defender is removed in Step 2, the attack fizzles with "zero combat or retaliation damage". Berserk cleave is non-combat damage. Is it also cancelled? | Cancelled with the attack. |
| KIM-S01 | I | R | **Readiness of a new Scratch Company.** `FR-ZON-005` does not say whether it enters Readied or Exhausted. `FR-SET-004`'s "Readied by default" suggests Readied, which lets a Headquarters scratch and then the new unit immediately generate 1 resource. | Readied. |
| KIM-S02 | I | R | **HQ defeat timing.** `FR-CBT-003` says a player loses "the moment" their HQ reaches ≤ 0 Current Health, while `FR-LAY-002` runs destruction in the state-based sweep after the resolving item finishes. With *Rift* ("Scratch 2. Deal 5 damage to this Headquarters."), or any mid-resolution HQ damage followed by more clauses, which applies? | Checked at the death sweep. Equivalent in M1, where HQs only take combat damage. |
| KIM-S03 | I | R | **Generating 0.** May an asset exhaust to generate a resource it yields 0 of (e.g. a 0-Tactics-yield unit in the Main Theater)? Matters for exhaust-watching triggers. | Allowed. |
| KIM-C01 | I | U | **Cornered-gated triggers inside a simultaneous step.** Cornered triggers are gated "at the moment the trigger event occurs"; during a simultaneous step Cornered is locked at the start of the step (`FR-LAY-003`). A damage-taken trigger behind a Cornered gate fires mid-step: is it gated by the locked value? No current card has a Cornered-gated damage trigger. | n/a |
| KIM-B07 | ? | U | **Berserk stacking.** Printed Berserk plus granted Berserk: one secondary with summed N, or two secondaries? Only *Tempest Surge* (Concordia) and *Redline Infusion* (Soulless) carry Berserk, so no legal deck has both. | n/a |

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
- **Ward × Grd:** a Ward Guardian can still be attacked, so it still restricts attackers (`FR-KWD-002h`, `FR-KWD-002j`).
- **Grd × DEx:** Guardian has no readiness condition; an Exhausted Guardian still forces attacks onto it (`FR-CBT-001`). Worth confirming against prototype habit.
- **Arm × Arm:** Armor stacks additively and never expires (`FR-LAY-002`).
- **Arm × Bsk:** Armor absorbs Berserk cleave (`FR-REP-004` Stage 4).
- **DEx × Rgp:** Regroup readies any exhausted allied asset, including one that just deployed exhausted (`FR-KWD-002n`).
- **Scr × Rgp:** "Regroup all Scratch Company units in your Reserves" readies them all (`FR-KWD-002n`).
