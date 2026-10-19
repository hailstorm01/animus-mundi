# Product Manager Devil's Advocate Audit: Game Rules FRD (Master FRD v3.2)

- **Target Document:** [docs/frd/01_Game_Rules_FRD.md](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_FRD.md)
- **Auditor:** PM Devil's Advocate (`pm-devils-advocate`)
- **Evaluation Date:** 2026-10-19
- **Status:** PENDING (Active review queue)
- **Verdict:** REVISE WITH BLOCKERS

## 1. Executive Summary & Red-Team Assessment

The Game Rules specification (`01_Game_Rules_FRD.md`) demonstrates high structural maturity and rigorous separation of concerns (`SOC-002`), having successfully offloaded compiler/pipeline schemas to FRD 02 and runtime memory models to FRD 05. However, a critical cross-specification conflict has emerged around **Berserk N**: FRD 01 explicitly establishes that Berserk cleave deals flat $N$ damage *unconditionally* regardless of primary combat outcome, whereas downstream engine specification FRD 05 (`FR-ENG-035`) implements Berserk as a *kill-contingent trample* that only cleaves if the primary defender is destroyed. Furthermore, simultaneous empty deck draw scenarios remain unaddressed in the terminal loss rules, and attacker death timing during Berserk cleave execution requires formal clarification.

## 2. Severity-Ordered Punch-List

### 🔴 Blocker / Critical
- **[DA-RULES-011] Direct Cross-Spec Contradiction: Berserk N Secondary Cleave Condition (Unconditional vs Primary Kill)**
  - **Requirement Citations:** `FR-CBT-002`, `FR-KWD-002o` (cross-ref `05_Headless_Game_Engine_FRD.md` `FR-ENG-035`)
  - **Vulnerability / Flaw:** `FR-CBT-002` (Step 3) and `FR-KWD-002o` authoritatively state that whenever an attacking Unit with Berserk N attacks an opposing asset in the Main Theater, in addition to resolving combat damage against its primary combat target, "it always deals $N$ non-combat effect damage... to the pre-selected secondary opposing Asset in the Main Theater... Berserk deals flat $N$ damage regardless of the primary combat damage outcome or the attacker's Combat Power." Conversely, FRD 05 `FR-ENG-035` item 3 implements the exact opposite requirement: *"If an attacking Berserk N unit (FR-KWD-002o) destroyed its primary Main Theater target during Step 3 combat damage and a legal pre-bound SECONDARY_BERSERK_TARGET remains in MainTheaterZone, the engine routes flat N non-combat effect damage..."*. This is a direct, irreconcilable contradiction: does Berserk cleave unconditionally, or is it gated on the primary target's lethal destruction?
  - **Exploit / Failure Scenario:** A player attacks with a 4/4 Berserk 2 unit into an 8/8 defender while an enemy 1/1 occupies the secondary Main Theater slot. Under FRD 01 rules, the 1/1 secondary defender takes 2 cleave damage and dies even though the 8/8 survived. Under FRD 05 engine code, zero cleave damage is dealt because the 8/8 did not die. Tournament rulings and automated engine validation cannot co-exist with this divergence.
  - **Remedy:** Harmonize FRD 01 and FRD 05 authoritatively. If Berserk is intended to deal unconditional sweep damage, excise the destruction check from FRD 05 `FR-ENG-035`. If Berserk is intended as a trample mechanic that only triggers upon destroying the primary defender, update `FR-CBT-002` and `FR-KWD-002o` to explicitly require lethal destruction of the primary target.

### 🟠 High
_None._

### 🟡 Medium
- **[DA-RULES-012] Simultaneous Empty Deck Defeat Rule Ambiguity**
  - **Requirement Citations:** `FR-WIN-001`, `FR-CBT-003`, `FR-QUE-001` (cross-ref `04_Client_Engine_Protocol_Networking_FRD.md` `FR-MSG-004`, `05_Headless_Game_Engine_FRD.md` `FR-ENG-042`)
  - **Vulnerability / Flaw:** `FR-WIN-001` specifies that drawing while a deck has 0 cards causes immediate match loss, while `FR-CBT-003` / `FR-ENG-042` specify that simultaneous reduction of both Headquarters to $\le 0$ health terminates in a Match Draw (`RESULT_MATCH_DRAW`). However, the rules fail to specify the outcome when a symmetrical card draw effect (e.g., *"Each player draws 2 cards"*) forces both players to draw from empty decks during the exact same stack resolution step, or when mixed simultaneous terminal events occur (e.g., combat reduces defender HQ to 0 while an Unbowed trigger simultaneously forces the attacker to draw from an empty deck).
  - **Exploit / Failure Scenario:** Both players have empty decks. A player casts a symmetrical draw spell. Does the active player lose first due to APNAP resolution sequencing, or does the match conclude in a Match Draw?
  - **Remedy:** Formalize in `FR-WIN-001` and `FR-CBT-003` that if both players are required to draw from empty decks during the exact same resolution step or simultaneous event batch, the match terminates immediately as a Match Draw (`RESULT_MATCH_DRAW`).

- **[DA-RULES-013] Ambiguous Attacker Death Effect on Berserk Secondary Cleave Execution**
  - **Requirement Citations:** `FR-CBT-002`, `FR-KWD-002o`
  - **Vulnerability / Flaw:** Step 3 of `FR-CBT-002` resolves primary combat strike damage and retaliation simultaneously. If the attacking Berserk unit suffers lethal retaliation damage from the defender, it dies during Step 3. While `FR-KWD-002o` states Berserk deals flat $N$ damage "regardless of the primary combat damage outcome", the specification does not state whether the attacker must survive combat retaliation to emit the secondary cleave damage, or if secondary cleave damage is committed concurrently with primary combat damage before state-based death sweeps deallocate the attacker.
  - **Exploit / Failure Scenario:** A 3/1 Berserk 2 attacker strikes a 4/4 defender. The attacker takes 4 retaliation damage and dies. The attacker claims 2 secondary cleave damage against another defender. The opponent argues that dead attackers cannot deal damage after leaving the battlefield.
  - **Remedy:** Clarify in `FR-CBT-002` Step 3 that secondary Berserk cleave damage is committed concurrently with primary combat strike damage and retaliation within Step 3, prior to the post-combat state-based death sweep, ensuring it executes even if the attacker takes lethal retaliation damage.

### 🔵 Low / Clarification
- **[DA-RULES-014] Redundant Subsumed Inherent Action Requirement Stubs**
  - **Requirement Citations:** `FR-INH-001` through `FR-INH-004` (cross-ref `FR-CAP-001`)
  - **Vulnerability / Flaw:** Section 4.1 preserves four distinct requirement headers (`FR-INH-001` through `FR-INH-004`) solely to state that they are completely subsumed by `FR-CAP-001`.
  - **Exploit / Failure Scenario:** Cross-spec references may cite legacy `FR-INH-*` stubs rather than canonical `FR-CAP-001`.
  - **Remedy:** Maintain them in the Redundancy Ledger and ensure companion specifications (FRD 04 and FRD 05) refer authoritatively to `FR-CAP-001`.

## 3. Redundancy & De-Scoping Ledger
| Requirement ID | Fully Subsumed By | Rationale / Recommendation |
| :--- | :--- | :--- |
| `FR-INH-001` – `FR-INH-004` | `FR-CAP-001` | Inherent asset capabilities, tactical actions, and economic operations are fully codified in `FR-CAP-001`. |
| `FR-ZON-101` | `FR-ZON-004` | Subsumed into Non-Battlefield Zone Transitions & Last-Known State. |
| `FR-ENT-101` | `FR-ZON-005` | Subsumed into Scratch Company Fog of War & Double-Blind Lifecycle. |
| `FR-TGT-102` | `FR-TGT-001` | Subsumed into Target Eligibility & Conjunction Filtering. |
| `FR-TGT-103` | `FR-TGT-003` | Subsumed into Two-Stage Target Verification, Fizzle & Evaporation. |

## 4. Socratic Interrogation Register
1. *If a Berserk attacker deals secondary cleave damage regardless of whether the primary defender survives or dies, does this decouple Berserk from the genre-standard trample fantasy and effectively turn every Berserk attack declaration into an unconditional multi-target spell?*
2. *With uncapped Reserves capacity and complete immunity to combat attacks, what mechanical disincentive exists to prevent a control deck from turtling 10+ assets in Reserves and converting excess Requisition to Tactics (2R -> 1T) every turn to cast board-wiping Maneuvers without ever placing a unit in the Main Theater?*
