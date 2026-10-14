# Product Manager Devil's Advocate Audit: Game Rules & Card Modeling FRD (Master FRD v3.0)

- **Target Document:** [docs/frd/Game_Rules_and_Card_Modeling_FRD.md](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/Game_Rules_and_Card_Modeling_FRD.md)
- **Auditor:** PM Devil's Advocate (`pm-devils-advocate`)
- **Evaluation Date:** 2026-10-14
- **Status:** PENDING (Active review queue)
- **Verdict:** REVISE WITH BLOCKERS

---

## 1. Executive Summary & Red-Team Assessment

The **Master Game Rules & Card Modeling FRD (v3.0)** provides a comprehensive, rigorous foundation for Card Town, consolidating entity hierarchies, zone rules, the 6-layer attribute pipeline (`FR-LAY-001` - `FR-LAY-003`), the 5-stage damage mitigation pipeline (`FR-REP-004`), and the 3-step combat lifecycle (`FR-CBT-002`). The resolution of the clean-stack invariant (`FR-QUE-001`) and Stalwart combat capping has substantially elevated mechanical consistency.

However, deep stress-testing across gameplay dynamics, rules integrity, and degenerate play patterns reveals critical systemic issues:
1. **Degenerate Turn 1 Base Racing**: `FR-CAP-001`, `FR-CBT-001`, and `FR-CBT-002` explicitly permit Headquarters (HQs) to declare attacks across the board directly against the opposing HQ starting on Turn 1. Because all HQs possess 2-3 Combat Power and 20-25 Health, both players can simply trade direct HQ attacks back and forth from their backlines, ignoring unit deployment, theater maneuvering, and card play entirely to rush the opponent down by Turn 6.
2. **Dynamic Keyword Mid-Combat Desynchronization**: Continuous auras conditioned on dynamic asset counts (`Cornered: count(opposing) > count(allied)`) have no defined snapshotting rule during simultaneous combat damage resolution (`Step 3`). If multiple combatants die simultaneously, fluctuating Cornered states will cause mid-step stat recalculations that produce unresolvable rules states and cascading deaths.
3. **Uncapped Reserves Infinite Turtling**: `FR-ZON-001` specifies that Reserves has zero capacity limits, and `FR-CBT-001` protects Reserves assets from all combat attacks. A player can safely amass dozens of units in Reserves, generate astronomical Requisition, hoard unlimited Armor on *Dominator Warsuit*, and stall matches indefinitely without ever contesting the Main Theater.

**Recommendation:** Strictly restrict Headquarters combat declarations (HQ should not attack HQ across the board), snapshot continuous condition gates during Step 3 damage calculation, and implement a reasonable capacity cap on the Reserves Zone.

---

## 2. Severity-Ordered Punch-List

### 🔴 Blocker / Critical

- **[DA-RULES-001] Direct Headquarters Cross-Board Combat Racing Bypasses Core Unit Gameplay**
  - **Requirement Citations:** `FR-CAP-001`, `FR-CBT-001`, `FR-CBT-002`, `FR-SET-004`
  - **Vulnerability / Flaw:** `FR-CAP-001` explicitly grants `canAttack()` to `HEADQUARTERS`. `FR-CBT-001` specifies: *"Direct cross-board attacks against opposing HQs are permitted from Turn 1... Attack declarations may target the opposing Headquarters directly regardless of Main Theater occupants, subject only to Guardian priority."* `FR-CBT-002` states: *"Retaliation is 100% symmetric across all combat pairings (Unit vs Unit, Unit vs Structure, Unit vs HQ, HQ vs HQ)."* Every Headquarters in `units-and-structures-HQ.csv` has 2 or 3 Combat Power and 20 to 25 Health.
  - **Exploit / Failure Scenario:**
    1. On Turn 1, Player 1 has an empty Main Theater. Player 1 declares an attack with their Headquarters (*Animoceps Rager*: 3 Power, 20 Health) directly targeting Player 2's Headquarters (*Enclave*: 2 Power, 25 Health).
    2. Because there are no Main Theater units, Guardian priority does not apply.
    3. Step 3 damage resolves: Player 1 deals 3 damage to Player 2's HQ; Player 2's HQ retaliates with 2 damage to Player 1's HQ.
    4. Taking 2 damage triggers *Animoceps Rager*'s rule (*"Whenever this Headquarters is dealt damage, Scratch 1"*), deploying a free Scratch Company unit into Reserves!
    5. On Turn 2, Player 2 retaliates by attacking Player 1's HQ with their HQ.
    6. Both players realize the fastest, most deterministic strategy to win is to simply bash each other's Headquarters every single turn, reducing HQ health by 2-3 per turn while bypassing the entire Main Theater board game. By Turn 7, the game is decided by raw base racing without either player needing to contest the theater or cast tactical maneuvers.
  - **Remedy:** Remove `canAttack()` from Headquarters by default in `FR-CAP-001`. Headquarters should be stationary command centers and defensive anchors, not mobile attackers. Mandate that Headquarters may only deal retaliation damage when attacked, or restrict attacks strictly to units occupying the Main Theater.

---

### 🟠 High

- **[DA-RULES-002] Continuous Aura Fluctuations from Dynamic Cornered Gate Mid-Combat Resolution**
  - **Requirement Citations:** `FR-KWD-002e`, `FR-LAY-003`, `FR-CBT-002`, `FR-LAY-002`
  - **Vulnerability / Flaw:** `FR-KWD-002e` defines Cornered as `count(opponentMainTheaterAssets) > count(playerMainTheaterAssets)`. `FR-LAY-003` notes that for *triggered* abilities, Cornered evaluates at enqueue time. However, for *Continuous Auras* (e.g. *Brigand Highlander*: `Cornered: This unit has +2 power`; *Ancestral Monument*: `Other allied assets in the same zone have Cornered: Resistant 1`), the bonus is applied continuously in Layer 4. In Step 3 of combat (`FR-CBT-002`), combat damage and retaliation occur simultaneously. If a combat exchange destroys multiple units, the Main Theater asset count changes instantaneously. If an asset survives damage solely because of `Cornered: Resistant 1`, does the simultaneous death of an enemy unit cause it to lose Cornered retroactively during damage calculation, causing its own destruction?
  - **Exploit / Failure Scenario:**
    1. Player 1 has 1 unit in Main Theater with `Cornered: Resistant 1`. Player 2 has 2 units in Main Theater. Player 1 is Cornered.
    2. Combat occurs. Damage is dealt simultaneously: Player 1 takes 2 damage (mitigated to 1 by Resistant 1, leaving 1 Health). Player 1's strike kills one of Player 2's units.
    3. The moment Player 2's unit is marked for death, Player 2 now has 1 unit in Main Theater.
    4. Player 1 is no longer Cornered! Layer 4 recalculates.
    5. If Resistant 1 vanishes before Step 3 damage is fully committed, Player 1 takes the full 2 damage and dies. If it recalculates after, Player 1 survives.
    6. Digital engine implementations will produce divergent outcomes depending on internal execution order.
  - **Remedy:** Mandate in `FR-LAY-003` and `FR-CBT-002` that all Continuous Auras with dynamic condition gates (like Cornered) snapshot their state at the entry of Step 3 (Damage Resolution) and freeze throughout all damage, armor, and retaliation calculations, updating only during post-combat state-based destruction cleanup.

- **[DA-RULES-003] Uncapped Reserves Zone Enables Non-Interactive Turtling and OTK Armor Hoarding**
  - **Requirement Citations:** `FR-ZON-001`, `FR-CBT-001`, `FR-ACT-001`, `FR-LAY-002`
  - **Vulnerability / Flaw:** `FR-ZON-001` specifies that the Reserves Zone has no upper capacity limit. `FR-CBT-001` mandates that assets in Reserves are strictly immune to combat attacks. Furthermore, `FR-ACT-001` confirms that activated abilities like *Dominator Warsuit*'s `R1: gain 1 armor` are repeatable Requisition sinks, and `FR-LAY-002` confirms that dynamic Armor never decays or expires.
  - **Exploit / Failure Scenario:**
    1. A defensive player deploys structures and units exclusively into the Reserves Zone.
    2. Because Reserves cannot be attacked, the opposing player in the Main Theater has no way to interact with or destroy these assets through combat.
    3. By Turn 6, the defensive player has 12 assets in Reserves generating 15 Requisition per turn.
    4. The player deploys *Dominator Warsuit* to Reserves and sinks 15 Requisition into its activated ability every turn, banking 45+ dynamic Armor.
    5. The opponent is powerless to stop the buildup without rare, specific direct-damage board wipes.
    6. Once fully armored, the player marches out and executes an unstoppable One-Turn-Kill (OTK).
  - **Remedy:** Implement a reasonable capacity cap on the Reserves Zone (e.g. maximum 6 assets) in `FR-ZON-001`, or establish that exceeding reserve capacity forces assets to march into the Main Theater or lose combat immunity.

---

### 🟡 Medium

- **[DA-RULES-004] Ambiguous Precedence Hierarchy Between Damage Redirection and Entity Death Replacement Watchers**
  - **Requirement Citations:** `FR-REP-001`, `FR-REP-003`, `FR-REP-004`
  - **Vulnerability / Flaw:** `FR-REP-003` states that when multiple replacement watchers intercept an event, the affected asset's controller designates the order. However, there is an architectural clash between Stage 1 Damage Redirection (*Animoceps Spore Cloud*: *"Redirect all damage that would be dealt to allied units in the same zone to this structure"*) and Entity Destruction Replacement (*Shifting Impossibility*: *"If an allied Scratch Company unit would be killed, return it to its owner's hand and deal 1 damage to this unit instead"*). If a Scratch Company takes lethal damage, does Spore Cloud intercept the prospective damage at Stage 1, or can the player choose to let the unit "die" and trigger Shifting Impossibility at Stage 5?
  - **Remedy:** Explicitly define an execution hierarchy in `FR-REP-001`: Stage 1 Damage Redirection operates on prospective *damage instances* and executes before any Stage 5 *destruction events* can be evaluated.

- **[DA-RULES-005] Zero-Power Combat Declarations Generate Risk-Free Attack Triggers**
  - **Requirement Citations:** `FR-CBT-001`, `FR-KWD-002f`, `FR-CBT-002`
  - **Vulnerability / Flaw:** `FR-CBT-001` explicitly allows units with Combat Power <= 0 to declare attacks (firing `EVENT_ATTACK_DECLARE`). While intended to allow negative Resistant amplification, it permits 0-power utility assets or tokens with on-attack triggers (or Aggressive N) to declare attacks risk-free against 0-power structures or defenders purely to harvest triggers, muddying the combat priority model.
  - **Remedy:** Clarify in `FR-CBT-001` whether declaring an attack requires positive base combat power or whether 0-power attacks are an intentional feature of the combat engine.

---

### 🔵 Low / Clarification

- **[DA-RULES-006] Unresolved Cleanup of Subsumed Requirement IDs in Master Traceability Matrix**
  - **Requirement Citations:** `FR-ZON-101`, `FR-ENT-101`, `FR-TGT-102`, `FR-TGT-103`, Appendix B
  - **Issue:** Appendix B lists `FR-ZON-101` (subsumed by `FR-ZON-004`), `FR-ENT-101` (subsumed by `FR-ZON-005`), `FR-TGT-102` (subsumed by `FR-TGT-001`), and `FR-TGT-103` (subsumed by `FR-TGT-003`). While the text indicates they are subsumed, retaining them in the active requirements matrix creates ambiguity for external tooling.
  - **Clarification Needed:** Move subsumed requirement IDs to an explicit historical archive ledger in the appendix.

---

## 3. Redundancy & De-Scoping Ledger

| Requirement ID | Fully Subsumed By | Rationale / Recommendation |
| :--- | :--- | :--- |
| `FR-ZON-101` (Clean Asset Deallocation) | `FR-ZON-004` (Non-Battlefield Reset) | Completely covered by `FR-ZON-004` steps 1-3. Deprecate ID. |
| `FR-ENT-101` (Scratch Double-Blind Proxy Wrapper) | `FR-ZON-005` (Scratch Fog of War Lifecycle) | Fully articulated in `FR-ZON-005`. Deprecate ID. |
| `FR-TGT-102` (Declarative Targeting Filters) | `FR-TGT-001` (Universal Modular Targeting) | Subsumed entirely by `FR-TGT-001`. Deprecate ID. |
| `FR-TGT-103` (Two-Stage Target Verification) | `FR-TGT-003` (Target Validation & Fizzle Rules) | Subsumed entirely by `FR-TGT-003`. Deprecate ID. |

---

## 4. Socratic Interrogation Register

1. *Why are Headquarters permitted to declare direct cross-board attacks against each other on Turn 1 before a single unit has entered combat?*
2. *How does the game prevent degenerate defensive turtling when the Reserves zone has infinite capacity and complete immunity to combat attacks?*
3. *If an asset with a continuous aura loses Cornered in the middle of simultaneous combat damage resolution, at what exact micro-step does its buff expire?*
4. *Why should a unit with 0 combat power be permitted to declare an attack solely to trigger attack-declaration abilities without dealing any combat damage?*
