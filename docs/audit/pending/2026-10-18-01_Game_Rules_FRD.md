# Product Manager Devil's Advocate Audit: Game Rules FRD

- **Target Document:** [docs/frd/01_Game_Rules_FRD.md](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_FRD.md)
- **Auditor:** PM Devil's Advocate (`pm-devils-advocate`)
- **Evaluation Date:** 2026-10-18
- **Status:** PENDING (Active review queue)
- **Verdict:** REVISE WITH HIGH FINDINGS

## 1. Executive Summary & Red-Team Assessment

This audit evaluates `01_Game_Rules_FRD.md` (Master FRD v3.2) across the core red-team lenses: redundancies, cross-spec conflicts, mechanical ambiguities, underdeveloped specifications, and rules gaps.

`01_Game_Rules_FRD.md` serves as the foundational root specification for Card Town, establishing pure game rules, card characteristics, zone topologies, dual-resource economics, turn sequencing, combat, 5-stage damage mitigation, continuous 6-layer attribute recalculation, and canonical keywords.

While the document maintains strong canonical ownership over core rules mechanics, deep interrogation reveals critical integration ambiguities and specification gaps:
1. **Omission of Hand/Discard Continuous Evaluation**: Section 6.2 restricts the 6-layer continuous recalculation pipeline strictly to "battlefield Asset characteristics" (`FR-LAY-001`), while `FR-ZON-004` states that cards leaving the battlefield revert to printed characteristics. However, downstream specifications (FRD 02 `FR-ABL-004`, FRD 05 `FR-ENG-013`) mandate dynamic cost reductions operating continuously on cards in **Hand** (*Swirling Dustdevil*, *Boneyard Amalgam*) and alternate deployment permissions in **Discard** (*Veil Render*). FRD 01 provides zero rules mechanics governing non-battlefield dynamic modifier evaluation.
2. **Ambiguous Delayed Trigger Instance Tracking**: When a face-down Scratch Company with a registered delayed trigger (*Rage's Wake*) is destroyed and enters Discard, `FR-TRG-006` tracks the entity by `(entityId)`—which is deallocated on battlefield departure (`FR-ENT-003`). If multiple cards of the same blueprint occupy Discard, the target card instance is ambiguous.
3. **Excess Damage from 0-Power Berserk Attackers**: When an attacker with effective Combat Power $\le 0$ attacks a target with negative `Resistant`, dealing positive damage through mitigation amplification and destroying the defender, it is ambiguous whether `FR-KWD-002o` permits excess damage to trample onto a secondary defender.
4. **Tucked Cards Fate During Face-Down Transformation**: `FR-ENT-103` specifies that `Forced Mutation` transforms a unit into a 1/1 Scratch Company in place without leaving the battlefield, leaving cards tucked in its `Attached` sub-zone in an undefined state.

## 2. Severity-Ordered Punch-List

### 🔴 Blocker / Critical
_None._

### 🟠 High
- **[DA-RULES-006] Omission of Hand and Discard Cost Reduction Rules in 6-Layer Architecture**
  - **Requirement Citations:** `FR-LAY-001`, `FR-CST-001`, `FR-ZON-004` (cross-ref `FR-ABL-004`, `FR-ENG-013`)
  - **Vulnerability / Flaw:** Section 6.2 explicitly restricts the 6-layer continuous recalculation pipeline strictly to "battlefield Asset characteristics" (`FR-LAY-001`), while `FR-ZON-004` dictates that cards in non-battlefield zones revert to printed characteristics. However, downstream specifications (FRD 02 `FR-ABL-004`, FRD 03, FRD 05 `FR-ENG-013`, `FR-ENG-025`) implement dynamic cost reductions operating continuously on `CardInstance` cards residing in **Hand** (*Swirling Dustdevil*, *Boneyard Amalgam*) and alternate deployment static permissions in **Discard** (*Veil Render*). FRD 01 contains zero rules describing how non-battlefield cards receive continuous or static modifiers, creating an unresolvable rules gap where cards in hand either cannot receive Layer 4 cost reductions or violate the "battlefield assets only" definition of the 6-layer pipeline.
  - **Exploit / Failure Scenario:** A player holds *Swirling Dustdevil* with 4 cards in Discard. Opponent argues that under `FR-LAY-001` and `FR-ZON-004`, cards in Hand are non-battlefield cards with no active Layer 4 continuous evaluation, meaning its printed cost of 4 cannot be modified by its continuous aura while in Hand.
  - **Remedy:** Explicitly formalize in `FR-LAY-001` and `FR-CST-001` how static cost-reduction abilities and non-battlefield auras (`activeZone: HAND / DISCARD`) evaluate on `CardInstance` entities on-demand during cost calculation without requiring full 6-layer battlefield instantiation.

### 🟡 Medium
- **[DA-RULES-007] Ambiguous Delayed Trigger Instance Tracking Across Zone Transitions**
  - **Requirement Citations:** `FR-TRG-006`, `FR-ZON-004`, `FR-ZON-005`
  - **Vulnerability / Flaw:** `FR-TRG-006` states that delayed triggered abilities track a watched asset `(entityId)` for the remainder of the turn and track its resulting card into Discard upon destruction (*Rage's Wake*). However, upon destruction, `AssetInstance` is deallocated and its `AssetID` is destroyed (`FR-ENT-003`). If multiple cards of the same blueprint occupy the Discard pile, tracking by `entityId` or blueprint slug is ambiguous. FRD 01 does not specify that the delayed trigger must bind to the unique `CardInstanceId` of the underlying card.
  - **Exploit / Failure Scenario:** Player 1 casts *Rage's Wake* on a face-down Scratch Company unit whose underlying card is a Unit card. The unit is destroyed and sent to Discard. A delayed trigger enqueues to deploy it from Discard. If another copy of that same card was already in Discard, which instance is deployed? If the destroyed Scratch Company turned out to be a Maneuver card, but a Unit card of the same name was already in Discard, does the delayed trigger mistakenly deploy the existing unit card?
  - **Remedy:** Specify in `FR-TRG-006` that delayed triggered abilities that track battlefield assets into non-battlefield zones bind to the persistent underlying `CardInstanceId`, disambiguating individual card instances upon asset wrapper deallocation.

- **[DA-RULES-008] Excess Damage Calculation for Non-Positive Power Berserk Attackers**
  - **Requirement Citations:** `FR-KWD-002o`, `FR-CBT-001`, `FR-REP-004`
  - **Vulnerability / Flaw:** `FR-CBT-001` allows units with effective Combat Power $\le 0$ to declare attacks, and `FR-REP-004` allows negative `Resistant` to amplify 0 raw damage into positive damage. `FR-KWD-002o` defines Berserk excess damage as "effective combat damage dealt by this unit minus the remaining damage needed to reduce the primary target's Armor and Current Health to 0". If a 0-power Berserk attacker strikes a 1-health defender with `Resistant -2` (taking 2 damage and dying), the damage dealt was 2, leaving 1 "excess damage". It is ambiguous whether Berserk excess damage can be generated by a 0-power attacker through target damage amplification.
  - **Exploit / Failure Scenario:** A player attacks with a 0-power Berserk unit against a 1-health unit debuffed with negative Resistant. The primary defender dies, and the attacker claims 1 excess damage against a secondary defender, dealing unblockable non-combat damage from an attacker with 0 offensive power.
  - **Remedy:** Clarify in `FR-KWD-002o` that Berserk excess damage requires the attacker to have effective Combat Power $> 0$, or explicitly affirm that negative Resistant amplification can generate Berserk excess damage.

- **[DA-RULES-009] Tucked Cards Fate During Face-Down Transformation (`Forced Mutation`)**
  - **Requirement Citations:** `FR-ENT-103`, `FR-ENT-102`, `FR-ZON-002`
  - **Vulnerability / Flaw:** `FR-ENT-103` specifies that turning a unit face-down transforms it into a 1/1 Scratch Company with no printed abilities while retaining damage, armor, and temporary modifiers. It does not state what occurs to cards currently tucked under it in its `Attached` sub-zone (`FR-ENT-102`). Because the unit does not leave the battlefield, `FR-ENT-102`'s host-departure discard trigger is not tripped, leaving tucked cards in an undefined state under a host that no longer possesses the ability to interact with them.
  - **Exploit / Failure Scenario:** *NO-AH-5* has 3 cards tucked under it. An opponent casts *Forced Mutation*. *NO-AH-5* becomes a 1/1 Scratch Company. Do the 3 tucked cards remain trapped in the Attached sub-zone forever without an aura referencing them? Or are they immediately purged to Discard? If purged, when do they purge and do they trigger discard-watchers?
  - **Remedy:** Add explicit clause to `FR-ENT-103` specifying that turning a host asset face-down immediately purges all cards in its `Attached` sub-zone face-up to their owners' Discard piles.

### 🔵 Low / Clarification
- **[DA-RULES-010] Redundant Inherent Action Capabilities Matrix Duplication**
  - **Requirement Citations:** `FR-CAP-001`, `FR-INH-001` through `FR-INH-004`
  - **Vulnerability / Flaw:** `FR-CAP-001` in Section 3.3 authoritatively establishes the capabilities matrix for March, Attack, Deploy Scratch Company, and Inherent Resource Generation. Section 4.1 (`FR-INH-001` through `FR-INH-004`) re-declares the identical rules across four redundant sub-bullets without adding new mechanics.
  - **Exploit / Failure Scenario:** Dual-definition drift when capabilities are modified.
  - **Remedy:** Mark `FR-INH-001` through `FR-INH-004` as formally subsumed by `FR-CAP-001`, consolidating inherent action rules into a single authoritative location.

## 3. Redundancy & De-Scoping Ledger
| Requirement ID | Fully Subsumed By | Rationale / Recommendation |
| :--- | :--- | :--- |
| `FR-INH-001` – `FR-INH-004` | `FR-CAP-001` | Inherent asset capabilities, tactical actions, and economic operations are fully codified in `FR-CAP-001`. |
| `FR-ZON-101` | `FR-ZON-004` | Subsumed into Non-Battlefield Zone Transitions & Last-Known State. |
| `FR-ENT-101` | `FR-ZON-005` | Subsumed into Scratch Company Fog of War & Double-Blind Lifecycle. |
| `FR-TGT-102` | `FR-TGT-001` | Subsumed into Target Eligibility & Conjunction Filtering. |
| `FR-TGT-103` | `FR-TGT-003` | Subsumed into Two-Stage Target Verification, Fizzle & Evaporation. |

## 4. Socratic Interrogation Register
1. *If an asset with 0 Combat Power attacks a defender with negative Resistant, causing positive damage to be dealt through mitigation amplification, should that 0-power unit be allowed to trample excess damage onto a secondary defender via Berserk?*
2. *Why are Headquarters permitted to attack enemy Headquarters directly on Turn 1 when the Main Theater is empty, rather than requiring combat engagement to occur within the shared Main Theater zone? Does this encourage non-interactive base racing?*
