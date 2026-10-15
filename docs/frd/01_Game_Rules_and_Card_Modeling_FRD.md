# Game Rules & Card Modeling - Functional Requirements Document (Master FRD v3.0)
**Unified Match Engine Architecture, Entity Modeling & Rules Specification**

---

## EXECUTIVE SUMMARY & REVISION HISTORY

This document establishes the authoritative unified specification for the Collectible Card Game (CCG) engine, combining the master state machine, zone topology, execution pipelines, entity taxonomy, declarative ability schemas, and card modeling specifications into a single cohesive architectural standard. It supersedes and consolidates **Engine FRD (v2.2)** and **Card Modeling FRD (v1.2)** into **Master FRD v3.0**, serving as the definitive game rules and data architecture reference alongside the **Client-Engine Protocol Networking FRD (`04_Client_Engine_Protocol_Networking_FRD.md` - FRD v1.2)**.

### Consolidated Revision Summary (v2.2 / v1.2 -> v3.0):
1. **Unified Architecture Consolidation**: Merged match execution state machine, three-tier entity architecture, and declarative ability engine into a contiguous master specification, eliminating cross-spec version drift.
2. **Combat Lifecycle Formalization (`FR-CBT-002`)**: Formalized a deterministic 3-step combat sequence (`Step 1: Declaration & Exhaustion`, `Step 2: Reaction & Trigger Window`, `Step 3: Simultaneous Combat Damage Resolution & Retaliation`), resolving timing ambiguities for off-turn interrupts like *blind defile*.
3. **Clean Stack & Deployment Disambiguation (`FR-QUE-001`, `FR-TRN-003a`)**: Codified the invariant that manual player-initiated actions require `Stack.isEmpty() == true`, whereas triggered abilities, in-resolution free deployments, and reactive replacements execute without a clean-stack constraint.
4. **Keyword Rigor & Stalwart Formalization (`FR-KWD-002c`, `FR-REP-004`)**: Standardized Stalwart rules text to *"All combat damage taken is reduced to 1"*. Confirmed that combat retaliation damage is combat damage (Stalwart applies), non-combat damage bypasses Stalwart, and Stage 2 flat reductions (Resistant) do not stack additively with Stalwart unless incoming damage is reduced to zero.
5. **Full-Round Duration Model (`FR-LAY-003`)**: Standardized temporary buffs/debuffs spanning player turns (*blindside*, *Comms Relay*) to *"until the end of your opponent's turn"*, executing deterministic cleanup during Sub-Phase 4c of the recipient's turn.
6. **Headquarters Asset Targeting Invariants (`FR-ZON-003`, `FR-TGT-102`)**: Reaffirmed intentional design rules: Headquarters are battlefield Assets; direct non-scratch asset burn (*bring it down*, *Dolus Macrocannon*) may legally target HQs; *Chainbreaker Sappers* explicitly undermines structures and HQs but not units; *Arkha, Sun's Fury* is scoped to opposing HQs as a dedicated finisher.
7. **Control Transfer Spatial Invariant (`FR-ZON-002`, `FR-ENT-003`)**: Scoped mind-control abilities (*Rhapsody's hollow*) to opposing assets in the Main Theater, preventing per-player backline Reserves zone contamination.
8. **Hand Inspection Protocol (`FR-MOD-001`)**: Guaranteed a client inspection window for revealed opponent hands (*drifting concept*, *mage seeker's inquisition*) before auto-discard resolution.
9. **Atomic Primitives Formalization (`FR-PRM-001`)**: Formally defined the atomic effect primitive catalog in Section 4.5.
10. **Cornered Gate Timing (`FR-KWD-002e`)**: Codified Cornered as an enqueue gate evaluated strictly at event emission for triggered abilities.
11. **Piston Fist Prototype Respecification**: Rebalanced *Piston Fist Prototype* to a 6-drop 3R/3T 1/9 statline with Aggressive 9 and pneumatic recoil (`"Aggressive 9\nAfter this unit attacks, move this unit to your Reserves."`).

---

## DOMAIN 1: MATCH ARCHITECTURE & DECK LIFECYCLE

### Section 1.1: Deck Building & Legalities

#### FR-SET-001: Deck Construction & Rarity Copy Limits
- **Main Deck Size**: A legal main deck must contain a minimum of 50 cards, with no upper maximum limit.
- **Headquarters Card**: Exactly 1 designated Headquarters (HQ) card is selected during deck construction and resides outside the main deck stack. HQ cards are strictly prohibited from being included in the main deck. An HQ card must belong to a specific faction (Concordia, Soulless, Path, Animus) and cannot be Neutral.
- **Faction Deck Legality**: A deck may only contain cards matching the faction of its designated Headquarters card and cards belonging to the **Neutral** faction.
- **Neutral Cards**: Neutral cards may be included in any deck regardless of the chosen Headquarters card's faction.
- **Scratch Company Deck Prohibition**: The card `neutral_scratch_company` is a match-generated token entity and is strictly prohibited from being included in a starting main deck.
- **Rarity Copy Limits**: The maximum permitted duplicate copies sharing the same card identity/name in a main deck are governed strictly by rarity:
  - **Common**: Maximum of 4 copies
  - **Uncommon**: Maximum of 3 copies
  - **Rare**: Maximum of 2 copies
  - **Mythic**: Maximum of 1 copy

---

### Section 1.2: Match Initialization & First-Turn Rules

#### FR-SET-002: Match Initialization & Mulligan Protocol
- **Zone Setup**: Each player's designated HQ card is placed in their respective Headquarters Zone prior to match start.
- **Opening Hand**: Each player draws an initial hand of 6 cards from their shuffled main deck.
- **Mulligan Rule**: Prior to Turn 1, each player may perform a single partial mulligan within a dedicated 25-second countdown:
  1. Select N cards from the opening hand.
  2. Set those N cards aside.
  3. Draw N replacement cards from the main deck.
  4. Shuffle the N set-aside cards back into the main deck.

#### FR-SET-003: First Player Determination
- Prior to match initialization, an authoritative server-side random roll (0 or 1) determines the initial decision-maker. The winner chooses whether to Play First (Turn 1) or Draw First (Turn 2). If the 15-second decision timer expires, the server defaults to Play First.

#### FR-SET-004: First-Turn Rules
- The player taking Turn 1 skips their Draw Phase on Turn 1. All action declarations, deployments, and combat declarations are permitted on Turn 1.

---

### Section 1.3: Deck Engine Operations

#### FR-DEK-001: Deck Search Engine Primitive
- **Execution Scope**: Searching a deck grants full visibility of all remaining ordered cards within that deck.
- **Unrestricted Search**: If the effect specifies "a card", any card in the deck is eligible.
- **Restricted Search**: If filtering criteria are defined (e.g., "an Elemental"), only matching cards are eligible. If no match exists, the search resolves with no card selected.
- **Execution Order**:
  `Search Deck` -> `Select Card` -> `Execute Shuffle (FR-DEK-002)` -> `Transition Card to Destination`
- **Mandatory Shuffling Trigger**: Executing a search automatically flags the target deck for a mandatory shuffle operation (`FR-DEK-002`), regardless of whether a card was selected.

#### FR-DEK-002: Deck Shuffling Engine Primitive
- Applies a pseudo-random permutation across all remaining cards in the specified deck zone, completely resetting card order. Shuffling executes as an instantaneous state operation and bypasses the Resolution Stack (`FR-QUE-001`).

#### FR-WIN-001: Empty Deck Loss
- If a player is required to draw a card while their deck contains 0 cards, that player loses the match immediately.

---

## DOMAIN 2: THREE-TIER ENTITY TAXONOMY & GAME ZONES

```text
â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”
â”‚                      CardBlueprint (Immutable Master)                  â”‚
â”‚  - Static JSON data definition loaded at engine startup                â”‚
â”‚  - Base printed attributes: Costs, Stats, Types, Subtypes, Abilities   â”‚
â”‚  - Optional additional deployment costs (sacrifices, discards)         â”‚
â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜
                                    â”‚ 1-to-Many
                                    â–¼
â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”
â”‚                      CardInstance (Game-Session Entity)                â”‚
â”‚  - Unique session ID (e.g. card_inst_104)                              â”‚
â”‚  - Tracks Owner, Zone (Deck, Hand, Discard, Stack, Attached)           â”‚
â”‚  - Instance-level transient modifiers (cost increases/reductions)      â”‚
â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜
                                    â”‚ 1-to-1 (When entering Battlefield)
                                    â–¼
â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”
â”‚                      AssetInstance (Battlefield Entity)                â”‚
â”‚  - Exists strictly in HQ Zone, Reserves Zone, or Main Theater          â”‚
â”‚  - Tracks Controller, Readiness State (Readied / Exhausted)            â”‚
â”‚  - Persistent Damage, Dynamic Armor, Active Power Buffs/Auras          â”‚
â”‚  - Face-Down Masks (Scratch Company), Tucked Sub-Entities (NO-AH-5)    â”‚
â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜
```

### Section 2.1: Entity Hierarchy Specifications

#### FR-ENT-001: CardBlueprint (Master Definition)
- **Immutability**: Read-only singleton data structure instantiated once during application startup.
- **Scope**: Canonical printed profile of a card identity. All runtime cards of the same identity reference the identical `CardBlueprint`.
- **Identity Key**: Unique alphanumeric slug (e.g., `animus_veil_render`, `neutral_scratch_company`).
- **Deck Building Compliance (`FR-SET-001`)**: Exposes `rarity`, `faction`, and `cardType` to enforce legal deck construction.

#### FR-ENT-002: CardInstance (Session Card Entity)
- **Lifecycle Scope**: Instantiated during match setup or mid-game card creation. Exists across non-battlefield zones (`DECK`, `HAND`, `DISCARD`, `STACK`, `ATTACHED`).
- **Identity & Ownership**: Contains a globally unique `instanceId` and immutable `ownerPlayerId`.
- **Reference**: Maintains an immutable link to its parent `blueprintId`.
- **Zone State**: Tracks its current container zone (`FR-ZON-001`).

#### FR-ENT-003: AssetInstance (Battlefield Entity)
- **Lifecycle Scope**: Instantiated exclusively when a card enters an active battlefield zone (`HEADQUARTERS`, `RESERVES`, or `MAIN_THEATER`). Destroyed/deallocated immediately upon departing the battlefield (`FR-ZON-002`, `FR-ZON-004`).
- **Reference Link**: Direct reference to its generating `underlyingCardInstance`.
- **Control vs. Ownership**: Tracks `controllerPlayerId` independently of `ownerPlayerId`. Gaining control does **not** change physical zones, subject to `FR-ZON-002`.
- **Readiness State**: Tracks physical activation status (`READIED` vs. `EXHAUSTED`). Newly deployed assets enter play `READIED` by default (`FR-INH-001`) unless intercepted by `Deploys Exhausted` (`FR-KWD-002l`).

---

### Section 2.2: Blueprint Schema & Base Attributes

#### FR-CRD-101: Blueprint Attribute Schema
Every `CardBlueprint` validates against the following property schema:

| Field | Type | Description |
| :--- | :--- | :--- |
| `id` | `String` | Unique canonical identifier (snake_case). |
| `name` | `String` | Card title matching design sheet. |
| `cardType` | `Enum` | `HEADQUARTERS`, `UNIT`, `STRUCTURE`, `MANEUVER`. |
| `faction` | `Enum` | `CONCORDIA`, `SOULLESS`, `PATH`, `ANIMUS`, `NEUTRAL`. |
| `subtypes` | `Array<String>` | Normal case tags (e.g. `["Spirit", "Horror"]`, `["Infantry"]`). |
| `requisitionCost` | `Integer` | Base Requisition cost to deploy/cast (>= 0). |
| `tacticsCost` | `Integer` | Base Tactics cost to deploy/cast (>= 0). |
| `additionalDeploymentCosts`| `Array<CostDef>` | Optional non-resource deployment costs (sacrifices, discards). |
| `power` | `Integer` | Base Combat Power (>= 0; null for Maneuvers). |
| `health` | `Integer` | Base Maximum Health (>= 1; null for Maneuvers). |
| `requisitionYield`| `Integer` | Resources generated when exhausted for Requisition (>= 0). |
| `tacticsYield` | `Integer` | Resources generated when exhausted for Tactics (>= 0). |
| `deploymentZones` | `Array<Enum>` | Valid zones for initial deployment: `RESERVES`, `MAIN_THEATER`, `HEADQUARTERS`. |
| `rarity` | `Enum` | `COMMON`, `UNCOMMON`, `RARE`, `MYTHIC`, `HQ`. |
| `rulesText` | `String` | Raw printed card rules text. |
| `abilities` | `Array<AbilityDef>` | Declarative definitions of keywords, triggers, actions, and maneuvers. |

---

### Section 2.3: Zone Definitions & Topological Rules

#### FR-ZON-001: Zone Definitions
- **Headquarters Zone**: Per-player battlefield zone housing the designated HQ card. May generate either Requisition or Tactics yield upon exhausting (`FR-INH-004`).
- **Reserves Zone**: Per-player backline battlefield zone for safe deployment and resource generation with uncapped capacity (no maximum entity limit). Protected from direct combat attacks (`FR-CBT-001`). Units occupying Reserves exhaust to generate **Requisition** (`FR-INH-002`).
- **Main Theater Zone**: Shared central battlefield zone where engagement and combat occur with uncapped capacity (no maximum entity limit). Units occupying Main Theater exhaust to generate **Tactics** (`FR-INH-002`).
- **Player Hand**: Private hand area per player with uncapped capacity (no maximum hand size).
- **Zone Capacity Universal Invariant**: All zones (Hand, Reserves, Main Theater, Discard) are uncapped with no upper maximum entity capacity.
- **Player Deck**: Ordered draw stack for remaining cards in the deck.
- **Player Discard**: Public depository for destroyed Assets and spent Maneuvers.
- **Attached Zone**: Non-battlefield sub-container zone housing cards tucked under or attached to battlefield assets (`FR-ENT-102`).

> [!NOTE] Design Rationale (DA-RULES-003)
> Uncapped Reserves capacity and combat attack immunity are foundational pillars of Card Town's backline resource economy. Turtling in Reserves without deploying units to the Main Theater leaves the player's Headquarters completely vulnerable to direct attacks from opposing Main Theater units. In the absence of Guardian units in the Main Theater, aggressive opponents can close matches by Turn 4–5 through direct HQ strikes before a turtling player can deploy high-cost assets or accumulate excessive armor. Physical UI display constraints are handled through client-side card compaction and stacking, not artificial rules-level slot caps.

#### FR-ZON-003: Entity Nomenclature & Asset Classification
- **Battlefield Entities (Assets)**: Entities occupying an active battlefield zone (HQ Zone, Reserves Zone, Main Theater Zone) are designated as **Assets** (Headquarters, Unit, or Structure). Rules text referring to "assets" or "non-scratch assets" (e.g. *bring it down*, *Dolus Macrocannon*) intentionally includes Headquarters and Structures as eligible targets in addition to non-Scratch units.
- **Non-Battlefield Entities (Cards)**: Entities residing in non-battlefield zones (Hand, Deck, Discard, Attached) carry the Card suffix (HQ Card, Unit Card, Structure Card, Maneuver Card).
- **Card Types & Subtypes (`FR-CRD-101`, `FR-ZON-003`)**: Card Types are mutually exclusive (`HEADQUARTERS`, `UNIT`, `STRUCTURE`, `MANEUVER`). Standardized Scratch Company units possess both `Infantry` and `Scratch Company` subtypes.

---

### Section 2.4: Asset Movement, Immutability & Fog of War

#### FR-ZON-002: Asset Movement, Immutability & Control Transfer
- **Headquarters & Structure Movement Immunity**: Headquarters and Structures cannot execute movement actions (March) and are permanently immune to and ineligible targets for any effect that changes battlefield zones (`MOVE_ZONE`). They are strictly bound to their initial zones. They may be destroyed (routing to Discard) or returned to Hand via card abilities.
- **Unit Battlefield Movement**: Units change battlefield zones between Reserves and Main Theater via the inherent March action or card abilities.
- **Control Transfer Scope & Zone Invariance**: When a player gains control of an opposing Asset (via `GAIN_CONTROL`), the asset does **not** change zones immediately. Mind-control effects (*Rhapsody's hollow*) are strictly scoped to eligible opposing units in the **Main Theater Zone**, preventing foreign assets from directly deploying into an opponent's private Reserves backline. When control is transferred, the asset retains its current readiness state (`READIED` remains `READIED`, `EXHAUSTED` remains `EXHAUSTED`). In accordance with the absence of summoning sickness (`FR-SET-004`), a Readied stolen unit may legally declare attacks or activate abilities on the same turn it is acquired. If the new controller subsequently commands that unit to execute an inherent March action from the Main Theater, the unit moves into its **controller's** Reserves Zone (arriving exhausted; `FR-CAP-001`). Upon destruction, the asset is deallocated and routed strictly to its **owner's** Discard zone (`EVENT_DESTROY`).
- **Destruction**: When an Asset's Health drops to 0 or below, it is destroyed immediately and routed to its owner's Discard zone (`EVENT_DESTROY`).

#### FR-ZON-004 & FR-ZON-101: Non-Battlefield Zone Transition & State Reset
- **FR-ZON-004 (Subsumes FR-ZON-101)**: Whenever an Asset transitions from a battlefield zone to a non-battlefield zone (`HAND`, `DISCARD`, `DECK`):
  1. The `AssetInstance` is deallocated (`FR-ZON-101`).
  2. All persistent state attributes—accumulated damage, dynamic Armor, temporary stat modifiers, continuous aura registrations, and keywords—are wiped.
  3. The entity reverts strictly to its printed base card statistics.
- **Intra-Battlefield Movements (March)**: Moving between Reserves and Main Theater does **not** wipe attributes or buffs; dynamic state is fully preserved.

#### FR-SNP-101: Last-Known State Snapshotting
- Immediately prior to an Asset's destruction (`EVENT_DESTROY`) or departure from an active zone, the engine captures an immutable `AssetSnapshot` (`FR-SNP-101`). Secondary triggers evaluate criteria strictly against this snapshot.

#### FR-ZON-005 & FR-ENT-101: Scratch Company Fog of War & Double-Blind Lifecycle
- **FR-ZON-005 (Subsumes FR-ENT-101)**:
  - **Double-Blind Deployment**: When a Scratch Company unit is deployed, its underlying card identity is strictly **unknown to BOTH players**. The engine enforces complete double-blind information masking—neither the controlling player nor the opponent may inspect its underlying card face or identity (`FR-ENT-101`).
  - **Departure to Public Zone (Discard)**: When a Scratch Company unit enters Discard, it transitions face-up; its underlying card identity is revealed to **both players**.
  - **Departure to Private Zone (Hand, Deck)**: When returned to Hand or Deck, it reverts to a Card; its underlying identity is revealed **strictly to the owner of that private zone**.
  - **Transition to Attached Zone (Tucked / NO-AH-5)**: When tucked into `ATTACHED` under an asset (`FR-ENT-102`), it remains a battlefield sub-entity; its identity remains strictly **hidden from BOTH players**.

#### FR-ENT-102: Tucked Sub-Entities Container (`ATTACHED` Sub-Zone)
- **Spatial Container Definition**: `TUCK_UNDER` transitions a target card from any legal source zone (`BATTLEFIELD`, `HAND`, `DECK`, `DISCARD`) into the `ATTACHED` sub-container under a specified host asset. Tucking does **not** inherently imply entity destruction.
  - If an active battlefield `AssetInstance` is tucked, its `AssetInstance` deallocates from the battlefield. `EVENT_DESTROY` emits if and only if an explicit `KILL` primitive was executed as part of the card ability (e.g. *NO-AH-5*: *"Kill chosen allied non-Headquarters asset and tuck it face down under this unit"*).
  - Tucked cards can be attached `faceDown = true` (masked identity, as in *NO-AH-5*) or `faceDown = false` as mandated by the ability.
- **Host Departure & Discard Routing**: If the host leaves the battlefield (destroyed, bounced, or exiled), all tucked cards immediately route to their respective owners' `DISCARD` piles face-up, revealing their identities to both players.
  - Each routed card emits an `EVENT_DISCARD_CARD` event. Discard-entry triggers (*Solarheart Monk*, *Eerie Remnant*) resolve normally.
  - Routing from `ATTACHED` to `DISCARD` does **not** emit `EVENT_DESTROY` (the cards were sub-entities, not battlefield combatants dying on the field).

#### FR-ENT-103: Face-Down Transformation (`forced mutation`)
- Sets `isFaceDown = true` while retaining existing `underlyingCardInstance`. Stat masking overrides values to `1/1/1/1 Scratch Company`. Evaluates damage against new Max Health of 1 (dies if accumulated damage >= 1).
- **Departure to Private Zone & Face Reversion**: If a face-down Scratch Company unit (whether match-generated or morphed via *forced mutation*) departs the battlefield to a private zone (returned to hand via *Disorienting Pulse*, *Icewind Regent*, or *Ambush Tactics*, or placed into a deck), the `AssetInstance` is deallocated and its underlying card face is revealed strictly to its owner (`FR-ZON-005`). The entity reverts to its printed base identity in the owner's hand or deck, where it may subsequently be played normally as that card (`FR-ZON-004`).

---

## DOMAIN 3: RESOURCE ENGINE, TURN STRUCTURE & CAPABILITIES

### Section 3.1: Dual-Resource Lifecycle

#### FR-RES-001: Resource Accumulation & Pool Isolation
- Generated resources accumulate in the active player's temporary Requisition and Tactics pools during the Main Phase.
- **Upfront Resource Requirement**: Players must explicitly generate resources into their pools prior to declaring card plays or ability activations. The engine **will not** auto-exhaust assets during cost payment.

#### FR-RES-002: End-of-Turn Resource Reset
- Resource pools do not bank or persist across turns. All unspent resources clear to zero during End Phase Sub-Phase 4c.

#### FR-RES-003: Requisition to Tactics Conversion
- During the Main Phase, spending 2 Requisition adds 1 Tactics to the pool (unidirectional exchange).
- **Execution Model**: Conversion executes **instantaneously as an economic state operation** bypassing the Resolution Stack. May be performed at any time during the Main Phase while `Stack.isEmpty() == true`. Emits `EVENT_CONVERT_REQUISITION`.

---

### Section 3.2: Phase Sequencing & Priority

#### FR-TRN-001: Phase Sequence
1. **Ready Phase**: Engine automatically transitions all exhausted controlled Assets to `READIED`.
2. **Draw Phase**: Active player draws exactly 1 card from their Deck (skipped on Turn 1 by Player 1; `FR-SET-004`).
3. **Main Phase**:
   - **Start-of-Turn & Draw Trigger Frame**: Prior to granting the active player priority for manual action declarations, all pending "At the start of your turn" triggers (*Hammerhand Bivouac*) and Draw Phase triggers (*Portent Scrivener*) enqueue onto the Resolution Stack and auto-resolve completely (active player orders allied triggers; APNAP applies between players; `FR-TRG-003`).
   - **Action Priority & Turn Clock**: Once the start-of-turn trigger stack clears (`Stack.isEmpty() == true`), the active player may freely deploy cards, execute inherent actions, convert resources, enact Maneuvers, and activate abilities. Enforces an authoritative server-side **75-second Main Phase Turn Clock** (`turnClockMs`). If the timer expires without an action or manual phase handover, the engine broadcasts an automated warning at 20 seconds remaining (fuse/rope visual) and deterministically commits an `AutoEndTurnCommand` upon expiration, transitioning the match directly to the End Phase.
4. **End Phase**: Executes deterministically across four sequential sub-phases:
   - **Sub-Phase 4a (Event Emission & Trigger Enqueuing)**: Emits `EVENT_END_PHASE`. All "at end of turn" triggers and pending delayed triggers (e.g. *Redline Infusion*) enqueue onto the Resolution Stack. Active player orders allied triggers; APNAP sequencing applies between players (`FR-TRG-003`).
   - **Sub-Phase 4b (Trigger & Cascade Resolution)**: Auto-resolves the stack completely, including cascading triggers. Interactive prompts enforce the 15-second action clock (`FR-TRG-002`).
   - **Sub-Phase 4c (Resource Depletion & Expiration)**: Unspent pools clear to zero (`FR-RES-002`); all temporary modifiers scheduled to expire at the end of the current turn expire (`FR-LAY-003`).
   - **Sub-Phase 4d (Turn Handover)**: Active role passes to opponent; turn counter increments.

#### FR-TRN-002: Turn-Bound Structure
- Non-active players cannot make manual off-turn card plays or actions. Off-turn interaction is restricted strictly to executing triggers prompted by the Triggered Ability Engine.

---

### Section 3.3: Dynamic Cost Evaluation & Capability Matrix

#### FR-CST-001: Dynamic Deployment Cost Evaluation
Deployment and casting costs are computed dynamically via:
```text
Effective Cost = max(0, Base Cost + Sum(Instance Modifiers) + Sum(Aura Modifiers) - Sum(Conditional Reductions))
```
Resource costs clamp at zero. Free cast effects force resource costs to 0 while requiring non-resource costs to be satisfied.

#### FR-CAP-001: Unified Asset Capability Query Matrix
Inherent actions execute directly as instantaneous state operations bypassing the Resolution Stack, requiring `readiness == READIED` and `Stack.isEmpty() == true`:
- **`canMarch()`**: Permitted strictly for `UNIT`. Moves the unit between Reserves and Main Theater. Executing March transitions the unit's readiness state from `READIED` to `EXHAUSTED` (emitting `EVENT_EXHAUST` and `EVENT_MARCH`). The unit remains exhausted until the Ready Phase of its controller's next turn. *(Sequence note: For units like Piston Fist Prototype that move back to Reserves upon attacking, they remain exhausted until readied on a subsequent turn, after which marching back to the Main Theater again arrives exhausted).*
- **`canAttack()`**: Permitted for `UNIT` in Main Theater and `HEADQUARTERS`. Prohibited for `STRUCTURE`. Exhausts the attacker upon declaration (`FR-CBT-002`). Units/HQs with effective Power <= 0 are permitted to declare attacks (dealing 0 damage). Direct cross-board attacks against opposing HQs are permitted from Turn 1.
- **`canGenerateResources()` & `getAvailableResourceTypes()`**:
  - *Standard Units*: Reserves &rarr; `["REQUISITION"]`; Main Theater &rarr; `["TACTICS"]`. Exhausts unit upon generation.
  - *Versatile Units (`FR-KWD-002i`)*: Reserves or Main Theater &rarr; `["REQUISITION", "TACTICS"]` (controller chooses). Exhausts unit upon generation.
  - *Structures & Headquarters*: Any zone &rarr; `["REQUISITION", "TACTICS"]` (controller chooses). Exhausts asset upon generation.
- **`canDeployScratch()`**: Permitted strictly for `HEADQUARTERS`. Exhausts HQ to execute Scratch 1 (`FR-INH-004`).

> [!NOTE] Design Rationale (DA-RULES-001)
> Permitting Headquarters to attack is intentional and fully integrated into the game's resource economy. Attacking with a Headquarters requires `readiness == READIED` and immediately transitions the HQ to `EXHAUSTED` (`FR-INH-001`, `FR-CAP-001`). Because an exhausted HQ cannot execute `canGenerateResources()` or `canDeployScratch()`, spending early turns attacking with an HQ starves the player of basic resource generation, freezing their board development entirely. The opposing player is free to generate resources, deploy battlefield units, establish structures, and overwhelm the attacker. Resource development and combat declarations share the identical readiness economy; assuming a player can race with HQ attacks without suffering catastrophic board loss is fundamentally flawed.

---

## DOMAIN 4: ACTION EXECUTION, DEPLOYMENTS & EFFECT PRIMITIVES

### Section 4.1: Inherent Actions

#### FR-INH-001 through FR-INH-004: Inherent Actions
- **Inherent Action Rules**: Execute instantaneously as direct state operations upon payment/exhaustion without placing items on the stack. Require clean stack (`Stack.isEmpty() == true`). March, Attack, Resource Generation, and Scratch deployments all require the source asset to be `READIED` and transition that asset to `EXHAUSTED`. Emit synchronous event hooks (`EVENT_MARCH`, `EVENT_ATTACK_DECLARE`, `EVENT_GENERATE_REQUISITION`, `EVENT_GENERATE_TACTICS`, `EVENT_SCRATCH_COMPANY`, `EVENT_EXHAUST`).
- **Deploy Scratch Company (`FR-INH-004`)**: Direct deck-to-battlefield move (`DECK` &rarr; `RESERVES`). Double-blind to both players. Does not emit `EVENT_DRAW_CARD`. Safe on 0-card deck (creates 0 units without triggering empty deck loss).

---

### Section 4.2: Activated Abilities & Maneuver Enactment

#### FR-ACT-001 & FR-ACT-002: Custom Activated Abilities
- Pushes effect payload to Resolution Stack; requires `Stack.isEmpty() == true` to initiate.
- May be activated repeatedly during Main Phase as long as costs are satisfied, unless explicitly constrained (e.g. *Dominator Warsuit*'s `R1: gain 1 armor` serves as an intentional repeatable Requisition sink).

#### FR-TRN-003 & FR-MNV-001: Deployment & Enactment onto Stack
- Playing Asset Cards or enacting Maneuvers places action payload onto the Resolution Stack. Requires `Stack.isEmpty() == true`. Maneuvers resolve payload and route immediately to Discard.

#### FR-TRN-003a: Alternative Deployment & Additional Non-Resource Costs
- **Sacrifice Costs**: When an asset is killed as an additional deployment/enactment cost (e.g. *feed the despoiler*, *rhapsody's embrace*, *Veil Render*), `EVENT_DESTROY` emits immediately upon cost payment, enqueuing on-death triggers to the stack above the pending spell.
- **In-Resolution Free Deployments**: When an effect deploys a card for free during stack resolution (e.g., *world soul messenger*, *allshan's battle trance*, *mutation swell*), the unit is deployed atomically as direct state execution without pushing an extra stack item. Bypasses the clean stack rule.

---

### Section 4.3: Standard Atomic Effect Primitives

#### FR-PRM-001: Standard Atomic Effect Primitives
The engine provides a standardized catalog of atomic effect primitives:

| Primitive | Parameter Schema | Description & Engine Traceability |
| :--- | :--- | :--- |
| `DEAL_DAMAGE` | `amount: ValueToken` | Deals damage routed through 5-stage pipeline (`FR-REP-004`). |
| `HEAL` | `amount: ValueToken` | Removes accumulated damage counters up to amount (`FR-KWD-002d`). |
| `ADD_ARMOR` | `amount: ValueToken` | Increments dynamic Armor pool (`FR-KWD-002k`). |
| `MODIFY_STAT` | `attribute: Enum, amount: ValueToken, duration: Enum` | Modifies stat (`END_OF_TURN`, `END_OF_OPPONENT_TURN`, `persistent`). |
| `MOVE_ZONE` | `destinationZone: Enum` | Transitions asset between Reserves and Main Theater (`FR-ZON-002`). |
| `RETURN_TO_HAND`| `none` | Bounces asset to owner hand; deallocates asset (`FR-ZON-004`). |
| `KILL` | `none` | Destroys target asset immediately; routes to Discard (`FR-ZON-002`). |
| `DRAW_CARD` | `count: ValueToken` | Draws cards from owner deck (triggers `FR-WIN-001` on empty deck). |
| `DISCARD_CARD` | `count: ValueToken, targetPlayer: Enum` | Causes player to discard cards from hand. |
| `REVEAL_HAND` | `targetPlayer: Enum, selector: Enum` | Reveals player hand for inspection and targeted discard (`FR-MOD-001`). |
| `SCRATCH` | `count: ValueToken` | Sequentially executes `Deploy Scratch Company` (`FR-INH-004`, `FR-KWD-002m`). |
| `REGROUP` | `count: ValueToken` | Transitions up to N chosen exhausted assets to Readied (`FR-KWD-002n`). |
| `DEPLOY_FREE` | `sourceZone: Enum, filter: FilterDef` | Deploys matching card suppressing resource costs (`FR-TRN-003a`). |
| `SEARCH_DECK` | `filter: FilterDef, destination: Enum` | Searches deck; executes mandatory shuffle (`FR-DEK-001`). |
| `LOOK_TOP_CARDS`| `count: Integer, operations: Array<OpDef>` | Inspects top N cards with 15s action clock (`FR-MOD-001`). |
| `TUCK_UNDER` | `hostTarget: TargetRef, faceDown: Boolean` | Transitions target card/asset into ATTACHED zone under host asset without implicit kill (`FR-ENT-102`). |
| `GAIN_CONTROL` | `targetPlayer: Enum` | Transfers controller for Main Theater asset (`FR-ZON-002`). |
| `MORPH_SCRATCH` | `none` | Turns asset face-down into Scratch Company (`FR-ENT-103`). |
| `GENERATE_RESOURCE`| `resourceType: Enum, amount: ValueToken` | Adds resources to active temporary pool (`FR-RES-001`). |
| `RECYCLE_DISCARD_TO_DECK`| `targetPlayer: Enum` | Bulk transfers Discard to Deck with mandatory shuffle (`FR-DEK-002`). |
| `REDIRECT_DAMAGE_TARGET`| `newTarget: TargetRef` | Intercepts incoming damage in Stage 1 and re-routes uncommitted damage to new target (`FR-REP-004`). |
| `PUT_CARD_ON_DECK` | `destinationPosition: Enum [TOP, BOTTOM, SHUFFLED_IN], targetCard: TargetRef` | Places target card at specified deck position without triggering a deck search shuffle (unless SHUFFLED_IN; `FR-DEK-001`). |
| `EXHAUST` | `targetAsset: TargetRef` | Transitions target asset readiness state to EXHAUSTED (emitting `EVENT_EXHAUST`). |
| `GRANT_SUBTYPE` | `subtype: String, duration: Enum` | Temporarily grants specified subtype to target asset until designated duration (`END_OF_TURN`, `END_OF_OPPONENT_TURN`). |
| `GRANT_KEYWORD` | `keyword: Enum, parameter: ValueToken (optional), duration: Enum` | Temporarily grants specified keyword to target asset until designated duration (`END_OF_TURN`, `END_OF_OPPONENT_TURN`). |

---

## DOMAIN 5: UNIVERSAL TARGETING, MODAL & SELECTION PROTOCOL

### Section 5.1: Declarative Targeting Filters

#### FR-TGT-001 & FR-TGT-102: Targeting Schema Definition
- **FR-TGT-001 (Subsumes FR-TGT-102)**: Evaluates eligibility via conjunction: `[Allegiance] AND [Health State] AND [Subtype Filter] AND [EntityType] AND [Zone]`.
- **Subtype & Entity Negation**: Supports `negateSubtypes: boolean`, `excludeScratch: boolean`, `excludeHQ: boolean`, and `excludeUnits: boolean` shorthand. Setting `entityType: "ASSET"` with `excludeUnits: true` targets Structures and Headquarters (*Chainbreaker Sappers*).
- **Headquarters Filtering**: By default, generic `ASSET` targets include Headquarters unless filtered via `excludeHQ: true`.
- **Health State Filters**: Health filters explicitly support `CURRENT_HEALTH` (`Current Health = Max Health - Accumulated Damage`) and `MAX_HEALTH` (*Cathedral of Self*, *Gemstone Titan*).
- **Quantifiers & Allocations (`FR-TGT-002`)**: Single Target, Bounded Multi-Target, Unbounded Global Target; Fixed Constant Allocation or Distributed Pool Allocation.

#### FR-TGT-003 & FR-TGT-103: Target Validation & Fizzle Rules
- **FR-TGT-003 (Subsumes FR-TGT-103)**: Verified at declaration and re-verified prior to resolution. If a subset of targets becomes invalid, resolves against remaining valid targets. If all targets become invalid, targeted effect clauses fizzle completely. For composite actions containing both targeted clauses and non-targeted global clauses (e.g. *All out War*), invalidation of all target bindings fizzles only the targeted clauses; non-targeted clauses resolve to the fullest extent possible.
- **Distributed Pool Invalidation**: For actions distributing a pool of values (*Macrocannon Barrage*, `DISTRIBUTED_POOL`), values allocated to any target that becomes invalid prior to resolution evaporate completely and are **not** redistributed among remaining valid targets.
- **Deployment Trigger Target Availability**: The absence of legal opposing targets for an asset's deployment trigger (e.g. *Saboteur*) does **not** cause the deployment command to fail validation. The asset deploys legally; upon resolution, any clauses with legal targets resolve (such as mandatory friendly damage, which targets the deploying unit itself if no other allied assets exist), while clauses lacking legal targets fizzle.

#### FR-TGT-101: Bound Action Payload Model
Clients submit fully bound payloads specifying `actionSourceId`, `blueprintAbilityIndex`, `selectedModalIndices`, `targetBindings` (with `TargetRole`: `SOURCE`, `TARGET`, `SACRIFICE_COST`, `DISCARD_COST`, `EXHAUST_COST`, `DESTINATION_ZONE`), and `deckSelectionIndices`.
- **Upfront Target Locking for Conditional Triggers (Blast of Genius Model)**: For composite abilities where secondary damage or effects depend on an outcome evaluated during resolution (*Samsara, Living Tempest*), all external targets (e.g. `chosen asset`) are selected and bound upfront when the trigger enqueues onto the stack. Upon resolution, the initial actions execute (e.g. inspect top cards, put one into discard); if the condition is satisfied (e.g. the discarded card was a maneuver), the damage is committed to the locked target. If the condition is not satisfied, the damage clause does not execute (no-op on the pre-locked target), completing execution in a single atomic resolution without requiring a secondary interactive prompt.

---

### Section 5.2: Synchronous Choice Workflows & Timeouts

#### FR-MOD-001: Synchronous Choice Prompts & Fallbacks
When stack execution requires interactive decisions:
1. **Single-Option Fast-Path**: If exactly one legal target or valid option exists for a required choice, the engine auto-selects it immediately and resolves without issuing an interactive client prompt or pausing match execution.
2. **Action Clocks**: Active player choice prompts enforce a 15-second action clock. Inactive player choice prompts enforce a streamlined 10-second action clock (`FR-TRG-002`) to preserve active turn momentum.
3. **Revealed Hand Inspection Window**: When `REVEAL_HAND` executes (*drifting concept*, *mage seeker's inquisition*), the choosing player is guaranteed an active inspection window to view the opponent's hand before any auto-discard completes.
4. **Partitioned Deck Inspections**: For multi-destination inspections (*light of the world soul*, *devilhorn scouting party*), choice payloads support explicit partitioning across deployment, hand, and bottom-deck ordering.
5. **Deterministic Fallbacks on Expiry**:
   - `LOOK_TOP_CARDS`: Auto-picks index 0; routes rest to bottom.
   - Modal choices: Auto-selects lowest available indices.
   - Optional triggers: Defaults to Decline / Pass.
   - Mandatory targets: Selects random legal target.
6. **Disconnect Grace Window**: Allows 10 seconds to reconnect with frozen clock before fallback commits.

---

## DOMAIN 6: EVENT PIPELINE, LAYER ENGINE & DAMAGE MITIGATION

### Section 6.1: Trigger Event Taxonomy & Stack Architecture

#### FR-TRG-001: Trigger Event Hook Taxonomy
Standardized event hooks across core categories:
- **Asset Lifecycle & Zones**: `EVENT_DEPLOY`, `EVENT_DESTROY`, `EVENT_ZONE_CHANGE`, `EVENT_MARCH`, `EVENT_EXHAUST`, `EVENT_READY`
- **Combat & Damage**: `EVENT_ATTACK_DECLARE`, `EVENT_DEAL_DAMAGE`, `EVENT_TAKE_DAMAGE`, `EVENT_ATTACK_RESOLVED`, `EVENT_COMBAT_KILL`
- **Economy & Resources**: `EVENT_GENERATE_REQUISITION`, `EVENT_SPEND_REQUISITION`, `EVENT_CONVERT_REQUISITION`, `EVENT_GENERATE_TACTICS`, `EVENT_SPEND_TACTICS`
- **Hand & Deck Operations**: `EVENT_DRAW_CARD`, `EVENT_DISCARD_CARD`, `EVENT_MULLIGAN`
- **Actions, Maneuvers & Abilities**: `EVENT_ENACT_MANEUVER`, `EVENT_ACTIVATE_ABILITY`, `EVENT_SCRATCH_COMPANY`
- **Phase & Turn Lifecycle**: `EVENT_READY_PHASE`, `EVENT_DRAW_PHASE`, `EVENT_MAIN_PHASE`, `EVENT_END_PHASE`, `EVENT_TURN_START`, `EVENT_TURN_END`

#### FR-QUE-001: Centralized LIFO Stack & Clean Stack Invariant
- **Clean Stack Rule**: Player-initiated manual actions (Deploying cards, Enacting Maneuvers, Activating abilities) require `Stack.isEmpty() == true` to declare.
- **Triggered Exception**: Triggered abilities, in-resolution free deployments, and cascading triggers bypass the clean stack requirement and push/resolve directly on the active stack.

#### FR-TRG-002: Inactive Player Agency & Streamlined Action Clocks
- When stack resolution requires interactive choices from the active player (15s clock) or inactive player (streamlined 10s clock; `FR-MOD-001`), the engine issues a choice prompt. If exactly one legal option exists, the engine fast-paths auto-selection without issuing a prompt. Prompts resolve deterministically via fallbacks if the timer expires.

#### FR-TRG-003: APNAP Push Sequencing
- Triggers occurring simultaneously push via Active Player, Non-Active Player (APNAP) order: non-active player triggers push first; active player triggers push on top and resolve first. Allied triggers are ordered by their controller.

#### FR-TRG-004: Cascading Trigger Enqueuing
- Secondary triggers generated during the resolution of a stack item push directly to the top of the stack, resolving immediately in context before downstream stack items resolve.

#### FR-TRG-005: Non-Resource Trigger Costs Prohibition
- Triggered abilities cannot demand Requisition or Tactics resource costs. Triggered abilities may demand game-state conditions or non-resource costs (e.g. sacrifices).

#### FR-TRG-007: Event Causality Context Lineage
- Each event preserves causal lineage (parent event ID, root action ID, source entity snapshot) to enable deterministic replacement resolution and trigger scope validation.

---

### Section 6.2: Six-Layer Attribute Pipeline

```text
Layer 1: Base Blueprint Values (or Scratch Blueprint if isFaceDown == true)
    â”‚
    â–¼
Layer 2: Type, Subtype & Identity Grants
    â”‚
    â–¼
Layer 3: Dynamic Keyword Grants & Suppressions
    â”‚
    â–¼
Layer 4: Continuous Auras (Static Board State)
    â”‚
    â–¼
Layer 5: Dynamic Turn Modifiers & Temporary Buffs
    â”‚
    â–¼
Layer 6: Mathematical Clamping & Effective Value (Effective Power = max(0, Raw Power))
```

#### FR-LAY-001: Combat Power Clamping
- `Effective Power = max(0, Raw Power)`. Combat Power is the only stat clamping at zero. Negative Resistant values do not clamp.

#### FR-LAY-002: Health, Damage & Armor Invariants
- `Current Health = Max Health - Accumulated Damage`. Destruction flags immediately if `Accumulated Damage >= Max Health` or if Max Health drops to <= 0.
- Temporary Max Health debuffs (*poisoned rations*) alter Max Health in Layer 5 without modifying `accumulatedDamage`. If the unit survives, restoring Max Health effectively increases Current Health.
- **Dynamic Armor Invariants**: Dynamic Armor is a persistent, consumable defensive pool initialized upon deployment (`Armor N`) or incremented via `ADD_ARMOR` effect primitives. Armor does **not** expire at end of turn, does not decay, and stacks additively from discrete addition events. Armor absorbs incoming damage 1-for-1 in Stage 4 of `FR-REP-004`. Armor cannot be granted continuously by Layer 4 static auras (it exists strictly as an entity pool modified by discrete events).

#### FR-LAY-003: Keyword Resolution & Duration Rules
- **Full-Round Durations**: Modifiers lasting *"until the end of your opponent's turn"* (*blindside*, *Comms Relay*) persist through the current turn and the opponent's subsequent turn, expiring deterministically during Sub-Phase 4c of that opponent's turn.
- **Cornered (`FR-KWD-002e`)**: Evaluates `count(opponentMainTheaterAssets) > count(playerMainTheaterAssets)` across all battlefield Assets (Units and Structures). For triggered abilities, Cornered is an enqueue condition evaluated strictly at trigger event emission; once enqueued, it resolves unconditionally. For continuous auras and static abilities during simultaneous resolution sequences (such as board wipes, multi-target maneuvers, or Step 3 combat damage and retaliation; `FR-CBT-002`), all continuous condition gates (including Cornered count checks) are snapshotted at the start of the resolution step and locked throughout all simultaneous damage and mitigation calculations. Continuous auras re-evaluate only after all simultaneous damage has been committed and state-based destruction checks have completed.

---

### Section 6.3: Deterministic Damage Mitigation Pipeline

#### FR-REP-001 through FR-REP-003: Replacement Interception
- **Inline Execution**: Replacement watchers intercept prospective hooks (`WOULD_TAKE_DAMAGE`, `WOULD_BE_DESTROYED`, `WOULD_DEPLOY`) inline, modifying or replacing event payloads prior to event emission and bypassing the Resolution Stack.
- **Single Application per Event**: A replacement effect applies at most once per discrete event.
- **Simultaneous Replacement Conflicts & Controller Choice**: If multiple replacement watchers attempt to intercept the exact same event, the controller of the affected asset designates the application order. Once a replacement effect alters the event (e.g. prospective death is replaced by return-to-hand), any competing replacement watchers whose trigger conditions are no longer met immediately fizzle without effect.
- **Consequential Event Isolation**: Any damage or state mutations occurring as a direct consequence of a replacement effect (e.g. self-inflicted damage on *Shifting Impossibility*) emit as independent, discrete child events and cannot circularly replace their parent event.
- **Entity Destruction Interception**: For replacement watchers intercepting entity destruction (*Shifting Impossibility*), each asset's prospective destruction emits an individual interceptable death hook, permitting multiple simultaneous entity destructions from a single root damage event to each be replaced sequentially until the intercepting watcher's health is depleted.
- **Pipeline Precedence Invariant**: Damage-phase replacement watchers (Stage 1 Redirection, Stage 2 Flat Reductions; `FR-REP-004`) operate strictly on prospective damage instances and evaluate prior to damage commitment. Entity-lifecycle replacement watchers (intercepting `WOULD_BE_DESTROYED`) evaluate strictly in Stage 5 after accumulated damage is verified to be lethal. Damage redirected away from an entity in Stage 1 cannot cause that entity to enter a prospective destruction state.

#### FR-REP-004: Deterministic 5-Stage Damage Mitigation Pipeline
Every incoming damage instance resolves through 5 sequential automated stages:

```text
Incoming Damage Instance
    │
    ▼
[ Stage 1: Target Redirection ] ────► Moves uncommitted damage to new target (e.g. Spore Cloud)
    │
    ▼
[ Stage 2: Flat Reductions ] ───────► Modifies damage by target's total Resistant N (negatives amplify)
    │
    ▼
[ Stage 3: Combat Damage Capping ] ─► Applies Stalwart to combat damage (min(dmg, 1); no-op on non-combat)
    │
    ▼
[ Stage 4: Dynamic Armor ] ─────────► Deducts damage 1-for-1 from target's current Armor pool
    │
    ▼
[ Stage 5: Health Counter Commit ] ─► Remaining damage added to accumulatedDamage
```

1. **Stage 1 (Redirection)**: Evaluates downstream stages against new target's defenses.
2. **Stage 2 (Resistant)**: Subtracts Resistant N. Negative Resistant amplifies damage: an incoming damage instance of 0 raw damage against a target with `Resistant -1` is amplified to 1 damage ($0 - (-1) = 1$). Floors at 0 for positive reductions.
3. **Stage 3 (Stalwart)**: Applies strictly if `isCombatDamage == true` and target has Stalwart. Evaluates as `damage = min(damage, 1)`. Does not raise damage <= 0. Non-combat damage (maneuvers, abilities like *soul whip*) bypasses Stalwart entirely. *(Design Invariant: Stage 2 flat reductions executing prior to Stage 3 combat capping is intentional by design to prevent combining Resistant and Stalwart into complete combat damage invulnerability).*
4. **Stage 4 (Armor)**: Deducts damage 1-for-1 from target's dynamic Armor pool. Damage absorbed by Armor counts as damage taken and reduces the asset's dynamic Armor pool accordingly (consuming the armor). When a target possesses both Stalwart and Armor, Stalwart caps incoming combat damage to 1 in Stage 3, and Stage 4 absorbs that 1 damage by depleting 1 point from the dynamic Armor pool.
5. **Stage 5 (Health Commit & Trigger Emission)**: Remaining damage added to `accumulatedDamage`. `EVENT_TAKE_DAMAGE` fires if and only if actual damage is absorbed by Armor or committed to Health (`(armorAbsorbed + healthDamageCommitted) > 0`).

---

## DOMAIN 7: COMBAT RESOLUTION ENGINE

### Section 7.1: Combat Target Eligibility & Invariants

#### FR-CBT-001: Target Eligibility & Reserves Protection
- Attack declarations may target only opposing Assets in the Main Theater Zone or the opposing HQ in the HQ Zone.
- Reserves Zone assets are strictly protected from combat attacks.
- Assets with `Ethereal` (`FR-KWD-002g`) cannot be targeted by attacks.
- **Base Racing & Direct HQ Attacks**: Attack declarations may target the opposing Headquarters directly regardless of Main Theater occupants, subject only to Guardian priority.
- **Guardian Priority ("If Able")**: Attack declarations must target a Main Theater asset with Guardian when able. Guardian in Reserves cannot be attacked and does not redirect attacks. Ethereal overrides Guardian.
- **Zero-Power Attacks**: Units and HQs with Combat Power <= 0 may legally declare attacks (dealing 0 raw combat damage, firing `EVENT_ATTACK_DECLARE`, routing through `FR-REP-004` where negative Resistant may amplify: $0 - (-1) = 1$).

> [!NOTE] Design Rationale (DA-RULES-005)
> Permitting attacks with Combat Power <= 0 is an intentional design rule. It enables keyword interactions such as Aggressive N (which grants power during the Step 2 trigger window prior to Step 3 damage resolution) and allows attacks against targets with negative Resistant to amplify damage. Retaliation is 100% symmetric, ensuring zero-power attacks are not risk-free against armed defenders.

---

### Section 7.2: Formalized 3-Step Combat Lifecycle

#### FR-CBT-002: Combat Resolution Sequence & Symmetric Retaliation
Combat execution proceeds through three deterministic steps:

1. **Step 1 (Declaration & Exhaustion)**:
   - Attacker exhausts and designates a single valid target (`target`).
   - Synchronous `EVENT_ATTACK_DECLARE` emits.
2. **Step 2 (Reaction & Trigger Window & Action Lock)**:
   - All `EVENT_ATTACK_DECLARE` triggers enqueue onto the Resolution Stack (e.g. *blind defile*, *Aggressive N*, *Arkha, Sun's Fury*).
   - The stack auto-resolves completely.
   - **Combat Action Lock Invariant**: Declaration of an attack initiates an atomic, uninterruptible combat transaction. Manual player action declarations (deploying cards, enacting maneuvers, activating abilities) are strictly locked out for both players; once all Step 2 triggers resolve, the match advances immediately to Step 3 damage resolution without an intervening player priority window.
   - If an off-turn trigger destroys or removes the attacker (e.g. *blind defile*) or defender prior to Step 3, combat immediately terminates for that entity with zero damage dealt.
3. **Step 3 (Simultaneous Combat Damage Resolution & Retaliation)**:
   - If both combatants remain valid, combat damage is dealt **simultaneously** equal to effective Combat Power.
   - Retaliation is 100% symmetric across all combat pairings (Unit vs Unit, Unit vs Structure, Unit vs HQ, HQ vs HQ).
   - Combat retaliation is combat damage (`isCombatDamage = true`); Stalwart applies.
   - Damage instances route through respective 5-stage pipelines (`FR-REP-004`).
   - **Continuous Aura Snapshot Lock**: Continuous aura condition gates (including Cornered count checks; `FR-LAY-003`) are snapshotted and locked at the entry of Step 3; all simultaneous damage and retaliation calculations evaluate against this locked state.
   - Following damage application, state-based destruction checks evaluate (`FR-LAY-002`, `FR-CBT-003`). If an attacking unit with `Berserk` (`FR-KWD-002o`) destroyed its Main Theater target, its excess combat damage routes to another chosen opposing asset in the Main Theater as discrete non-combat effect damage.

#### FR-CBT-003: Persistent Damage, Defeat Rules & Match Draws
- Damage remains marked across turns. When an HQ reaches <= 0 Health, its controller loses immediately.
- **Match Draw**: If both Headquarters reach <= 0 Health in the same atomic resolution step, the match concludes as a **Match Draw**.

---

## DOMAIN 8: DATA INGESTION, JSON SCHEMA & BENCHMARKS

### Section 8.1: Production JSON Blueprint Schema

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "CardBlueprintCollection",
  "type": "object",
  "properties": {
    "version": { "type": "string" },
    "cards": {
      "type": "array",
      "items": {
        "type": "object",
        "required": [
          "id", "name", "cardType", "faction", "subtypes", 
          "requisitionCost", "tacticsCost", "rarity", "rulesText", "abilities"
        ],
        "properties": {
          "id": { "type": "string" },
          "name": { "type": "string" },
          "cardType": { "type": "string", "enum": ["HEADQUARTERS", "UNIT", "STRUCTURE", "MANEUVER"] },
          "faction": { "type": "string", "enum": ["CONCORDIA", "SOULLESS", "PATH", "ANIMUS", "NEUTRAL"] },
          "subtypes": { "type": "array", "items": { "type": "string" } },
          "requisitionCost": { "type": "integer", "minimum": 0 },
          "tacticsCost": { "type": "integer", "minimum": 0 },
          "additionalDeploymentCosts": { "type": "array", "items": { "type": "object" } },
          "power": { "type": ["integer", "null"] },
          "health": { "type": ["integer", "null"], "minimum": 1 },
          "requisitionYield": { "type": "integer", "minimum": 0 },
          "tacticsYield": { "type": "integer", "minimum": 0 },
          "deploymentZones": { 
            "type": "array", 
            "items": { "type": "string", "enum": ["RESERVES", "MAIN_THEATER", "HEADQUARTERS"] } 
          },
          "rarity": { "type": "string", "enum": ["COMMON", "UNCOMMON", "RARE", "MYTHIC", "HQ"] },
          "rulesText": { "type": "string" },
          "abilities": { "type": "array", "items": { "type": "object" } }
        }
      }
    }
  }
}
```

---

### Section 8.2: Canonical Benchmark Blueprint Implementations

#### 1. `NO-AH-5` (Soulless Mythic Unit - Tucking & Scaling)
```json
{
  "id": "soulless_no_ah_5",
  "name": "NO-AH-5",
  "cardType": "UNIT",
  "faction": "SOULLESS",
  "subtypes": ["Officer", "Cybernetic"],
  "requisitionCost": 3,
  "tacticsCost": 0,
  "power": 2,
  "health": 5,
  "requisitionYield": 1,
  "tacticsYield": 2,
  "deploymentZones": ["RESERVES"],
  "rarity": "MYTHIC",
  "rulesText": "Ward\nR2: kill chosen allied non HQ asset and put it face down under this unit. This unit has +2 power for each card underneath it.",
  "abilities": [
    { "abilityType": "KEYWORD", "keyword": "WARD" },
    {
      "abilityType": "ACTIVATED",
      "costs": [{ "costType": "RESOURCE", "resource": "REQUISITION", "amount": 2 }],
      "targeting": {
        "selector": "CHOSEN",
        "filter": { "allegiance": "ALLIED", "entityType": "ASSET", "excludeHQ": true },
        "count": 1
      },
      "effect": { "primitive": "TUCK_UNDER", "hostTarget": "SELF" }
    },
    {
      "abilityType": "CONTINUOUS_AURA",
      "scope": "SELF",
      "modifiers": [
        {
          "attribute": "COMBAT_POWER",
          "operation": "ADD",
          "value": { "type": "DYNAMIC", "source": "SELF", "property": "COUNT_TUCKED_CARDS", "multiplier": 2 }
        }
      ]
    }
  ]
}
```

#### 2. `Piston Fist Prototype` (Soulless Mythic Unit - 6-Drop Pneumatic Recoil)
```json
{
  "id": "soulless_piston_fist_prototype",
  "name": "Piston Fist Prototype",
  "cardType": "UNIT",
  "faction": "SOULLESS",
  "subtypes": ["Mechanized"],
  "requisitionCost": 3,
  "tacticsCost": 3,
  "power": 1,
  "health": 9,
  "requisitionYield": 1,
  "tacticsYield": 2,
  "deploymentZones": ["RESERVES"],
  "rarity": "MYTHIC",
  "rulesText": "Aggressive 9\nAfter this unit attacks, move this unit to your Reserves.",
  "abilities": [
    { "abilityType": "KEYWORD", "keyword": "AGGRESSIVE", "parameter": 9 },
    {
      "abilityType": "TRIGGERED",
      "triggerHook": "EVENT_ATTACK_RESOLVED",
      "condition": { "source": "SELF" },
      "effect": {
        "primitive": "MOVE_ZONE",
        "destinationZone": "RESERVES"
      }
    }
  ]
}
```

#### 3. `Animoceps Spore Cloud` (Animus Rare Structure - Replacement Redirection)
```json
{
  "id": "animus_animoceps_spore_cloud",
  "name": "Animoceps Spore Cloud",
  "cardType": "STRUCTURE",
  "faction": "ANIMUS",
  "subtypes": ["Fungus"],
  "requisitionCost": 4,
  "tacticsCost": 0,
  "power": 0,
  "health": 9,
  "requisitionYield": 2,
  "tacticsYield": 2,
  "deploymentZones": ["MAIN_THEATER"],
  "rarity": "RARE",
  "rulesText": "Redirect all damage that would be dealt to allied units in the same zone to this structure.",
  "abilities": [
    {
      "abilityType": "REPLACEMENT_WATCHER",
      "interceptHook": "WOULD_TAKE_DAMAGE",
      "condition": { "allegiance": "ALLIED", "entityType": "UNIT", "zone": "SAME_ZONE" },
      "replacementAction": { "primitive": "REDIRECT_DAMAGE_TARGET", "newTarget": "SELF" }
    }
  ]
}
```

#### 4. `Samsara, Living Tempest` (Concordia Mythic Unit - Upfront Bound Trigger & Composite Partition)
```json
{
  "id": "concordia_samsara_living_tempest",
  "name": "Samsara, Living Tempest",
  "cardType": "UNIT",
  "faction": "CONCORDIA",
  "subtypes": ["Elemental", "Tempest"],
  "requisitionCost": 6,
  "tacticsCost": 0,
  "power": 6,
  "health": 7,
  "requisitionYield": 0,
  "tacticsYield": 2,
  "deploymentZones": ["MAIN_THEATER"],
  "rarity": "MYTHIC",
  "rulesText": "Whenever this unit attacks, look at the top 2 cards of your deck. Put one into your hand and the other into your discard pile. If the card put into your discard pile this way was a maneuver card, deal 2 damage to chosen asset.",
  "abilities": [
    {
      "abilityType": "TRIGGERED",
      "triggerHook": "EVENT_ATTACK_DECLARE",
      "condition": { "source": "SELF" },
      "targeting": {
        "selector": "CHOSEN",
        "filter": { "entityType": "ASSET" },
        "count": 1,
        "bindingPolicy": "UPFRONT_LOCK"
      },
      "effect": {
        "primitive": "COMPOSITE_SEQUENCE",
        "steps": [
          {
            "primitive": "LOOK_TOP_CARDS",
            "count": 2,
            "partition": {
              "destinations": [
                { "zone": "HAND", "count": 1 },
                { "zone": "DISCARD", "count": 1, "trackAs": "discardedCard" }
              ]
            }
          },
          {
            "primitive": "CONDITIONAL_CLAUSE",
            "condition": {
              "targetRef": "discardedCard",
              "cardType": "MANEUVER"
            },
            "effect": {
              "primitive": "DEAL_DAMAGE",
              "target": { "selector": "BOUND_TARGET", "targetIndex": 0 },
              "amount": { "type": "STATIC", "value": 2 }
            }
          }
        ]
      }
    }
  ]
}
```

---

## APPENDIX A: KEYWORD MECHANICS DICTIONARY

- **FR-KWD-002a: Unbowed**: Triggered ability firing on `EVENT_TAKE_DAMAGE`. Triggers whenever actual damage is absorbed by dynamic Armor or committed to Health (`(armorAbsorbed + healthDamageCommitted) > 0`). If damage was reduced to 0 in Stage 2, Unbowed does not trigger. If a unit takes lethal damage, Unbowed triggers and enqueues before state-based actions kill it. Because non-self effect clauses (e.g. *Hammerhand Brigands* `Scratch 1`, *Devilhorn Captain* `Deal 1 damage to chosen asset in the Main Theater`, *Chainbreaker Irregulars* `Draw a card, then discard a card`) do not depend on the deceased source unit remaining on the battlefield, they resolve completely from the grave. Only self-referential mutations targeting the deceased source unit itself (such as healing or buffing itself; *Macabre Dancers*, *Allshan, Chainbreaker Revenant*) fizzle upon stack resolution under `FR-TGT-103` and `FR-ZON-101`.
- **FR-KWD-002b: Resistant N**: Replacement watcher modifying incoming damage by N in Stage 2 of `FR-REP-004`. Can be negative (amplifies damage; does not clamp).
- **FR-KWD-002c: Stalwart**: Replacement watcher capping incoming combat damage at max 1 in Stage 3 of `FR-REP-004` (`min(damage, 1)`). Standardized rules text: *"All combat damage taken is reduced to 1"*. Applies strictly to combat damage resulting from an inherent Attack declaration or combat retaliation (`FR-CBT-002`). Has zero interaction with non-combat damage (maneuvers, abilities pass through unmodified).
- **FR-KWD-002d: Heal N**: Atomic primitive removing up to N accumulated damage counters.
- **FR-KWD-002e: Cornered**: Conditional gate evaluating `count(opponentMainTheaterAssets) > count(playerMainTheaterAssets)` across all battlefield Assets (Units and Structures). For triggered abilities, evaluates strictly as an enqueue check upon event emission.
- **FR-KWD-002f: Aggressive N**: Triggered ability firing on `EVENT_ATTACK_DECLARE`, granting +N Power until end of turn.
- **FR-KWD-002g: Ethereal**: Static restriction excluding asset from combat attack declarations. Overrides Guardian.
- **FR-KWD-002h: Ward**: Static combat restriction. This asset cannot be chosen by opponents except to attack in combat. Opponents cannot select this asset with non-combat abilities or Maneuvers.
- **FR-KWD-002i: Versatile**: Inherent action modifier for Units. Controller chooses whether to generate Requisition Yield or Tactics Yield upon exhausting, regardless of zone.
- **FR-KWD-002j: Guardian**: Static restriction. Opposing combat attacks must target an asset with Guardian when able. Guardian in Reserves does not redirect attacks.
- **FR-KWD-002k: Armor N**: Secondary health pool absorbing damage 1-for-1 before health deduction in Stage 4 of `FR-REP-004`. Damage absorbed counts as damage taken.
- **FR-KWD-002l: Deploys Exhausted**: Replacement watcher intercepting `WOULD_DEPLOY`, setting initial readiness to Exhausted.
- **FR-KWD-002m: Scratch N**: Sequentially executes `Deploy Scratch Company` N times. Direct deck-to-battlefield move; double-blind; safe on empty deck. Creates token units with subtypes `["Infantry", "Scratch Company"]` prohibited from starting decks.
- **FR-KWD-002n: Regroup N**: Atomic primitive transitioning up to N chosen exhausted allied assets to Readied.
- **FR-KWD-002o: Berserk**: Static combat ability. Whenever this unit attacks and destroys an opposing asset in the Main Theater, deal its excess combat damage (effective combat damage dealt minus damage required to destroy the target) to another chosen opposing asset in the Main Theater as discrete non-combat effect damage routed through `FR-REP-004`. Excess damage cannot target Reserves or Headquarters, and dissipates if no other opposing asset occupies the Main Theater.

---

## APPENDIX B: COMPLETE MASTER TRACEABILITY MATRIX

| Requirement ID | Canonical Scope | Master Specification Section |
| :--- | :--- | :--- |
| **FR-SET-001** | Deck Construction & Rarity Copy Limits | Section 1.1 (`FR-SET-001`), Section 8.1 |
| **FR-SET-002** | Match Initialization & Partial Mulligan | Section 1.2 (`FR-SET-002`) |
| **FR-SET-003** | First Player Determination Protocol | Section 1.2 (`FR-SET-003`) |
| **FR-SET-004** | First-Turn Rules & Draw Skip | Section 1.2 (`FR-SET-004`) |
| **FR-DEK-001** | Deck Search Engine Primitive | Section 1.3 (`FR-DEK-001`), Section 4.3 (`FR-PRM-001`) |
| **FR-DEK-002** | Deck Shuffling Primitive | Section 1.3 (`FR-DEK-002`), Section 4.3 (`FR-PRM-001`) |
| **FR-WIN-001** | Empty Deck Loss Condition | Section 1.3 (`FR-WIN-001`), Section 4.3 (`FR-PRM-001`) |
| **FR-ENT-001** | CardBlueprint Master Architecture | Section 2.1 (`FR-ENT-001`), Section 8.1 |
| **FR-ENT-002** | CardInstance Session Entity | Section 2.1 (`FR-ENT-002`) |
| **FR-ENT-003** | AssetInstance Battlefield Entity | Section 2.1 (`FR-ENT-003`) |
| **FR-CRD-101** | Blueprint Property Validation Schema | Section 2.2 (`FR-CRD-101`), Section 8.1 |
| **FR-ZON-001** | Battlefield & Non-Battlefield Zones | Section 2.3 (`FR-ZON-001`) |
| **FR-ZON-003** | Entity Nomenclature & Asset Taxonomy | Section 2.3 (`FR-ZON-003`) |
| **FR-ZON-002** | Movement, Immutability & Control Transfer | Section 2.4 (`FR-ZON-002`) |
| **FR-ZON-004** | Non-Battlefield Zone Transition Reset | Section 2.4 (`FR-ZON-004`; subsumes `FR-ZON-101`) |
| **FR-ZON-101** | Clean Asset Deallocation | Section 2.4 (Subsumed by `FR-ZON-004`; historical archive reference) |
| **FR-SNP-101** | Last-Known State Snapshotting | Section 2.4 (`FR-SNP-101`) |
| **FR-ZON-005** | Scratch Fog of War Lifecycle | Section 2.4 (`FR-ZON-005`; subsumes `FR-ENT-101`) |
| **FR-ENT-101** | Scratch Double-Blind Proxy Wrapper | Section 2.4 (Subsumed by `FR-ZON-005`; historical archive reference) |
| **FR-ENT-102** | Tucked Sub-Entities Container (NO-AH-5) | Section 2.4 (`FR-ENT-102`), Section 8.2 |
| **FR-ENT-103** | Face-Down Transformation (forced mutation)| Section 2.4 (`FR-ENT-103`) |
| **FR-RES-001** | Resource Accumulation & Pools | Section 3.1 (`FR-RES-001`) |
| **FR-RES-002** | End-of-Turn Resource Reset | Section 3.1 (`FR-RES-002`) |
| **FR-RES-003** | Requisition to Tactics Conversion | Section 3.1 (`FR-RES-003`) |
| **FR-TRN-001** | Phase Sequencing (4a-4d End Phase) | Section 3.2 (`FR-TRN-001`) |
| **FR-TRN-002** | Turn-Bound Structure & Off-Turn Priority| Section 3.2 (`FR-TRN-002`) |
| **FR-CST-001** | Dynamic Deployment Cost Pipeline | Section 3.3 (`FR-CST-001`) |
| **FR-CAP-001** | Unified Asset Capability Query Matrix | Section 3.3 (`FR-CAP-001`) |
| **FR-INH-001** | Universal Inherent Action Rules | Section 4.1 (`FR-INH-001`) |
| **FR-INH-002** | Unit Inherent Actions (March, Attack, Gen)| Section 4.1 (`FR-INH-002`) |
| **FR-INH-003** | Structure Inherent Actions (Generate) | Section 4.1 (`FR-INH-003`) |
| **FR-INH-004** | HQ Inherent Actions (Attack, Gen, Scratch)| Section 4.1 (`FR-INH-004`) |
| **FR-ACT-001** | Custom Activated Abilities | Section 4.2 (`FR-ACT-001`) |
| **FR-ACT-002** | Activation Frequency & Repeatability | Section 4.2 (`FR-ACT-002`) |
| **FR-TRN-003** | Deployment onto Resolution Stack | Section 4.2 (`FR-TRN-003`) |
| **FR-TRN-003a**| Alternative Deployments & Sacrifices | Section 4.2 (`FR-TRN-003a`) |
| **FR-MNV-001** | Maneuver Resolution Lifecycle | Section 4.2 (`FR-MNV-001`) |
| **FR-PRM-001** | Standard Atomic Effect Primitives Table | Section 4.3 (`FR-PRM-001`) |
| **FR-TGT-001** | Universal Modular Target Filtering | Section 5.1 (`FR-TGT-001`; subsumes `FR-TGT-102`) |
| **FR-TGT-002** | Target Quantifiers & Value Allocation | Section 5.1 (`FR-TGT-002`) |
| **FR-TGT-003** | Target Validation & Fizzle Rules | Section 5.1 (`FR-TGT-003`; subsumes `FR-TGT-103`) |
| **FR-TGT-101** | Bound Action Payload Model | Section 5.1 (`FR-TGT-101`) |
| **FR-TGT-102** | Declarative Targeting Filters | Section 5.1 (Subsumed by `FR-TGT-001`; historical archive reference) |
| **FR-TGT-103** | Two-Stage Target Verification | Section 5.1 (Subsumed by `FR-TGT-003`; historical archive reference) |
| **FR-MOD-001** | Inspection, Hand Reveals & Timeouts | Section 5.2 (`FR-MOD-001`) |
| **FR-TRG-001** | Trigger Event Hook Taxonomy | Section 6.1 (`FR-TRG-001`) |
| **FR-QUE-001** | Centralized LIFO Stack & Clean Stack Rule| Section 6.1 (`FR-QUE-001`) |
| **FR-TRG-002** | Inactive Player Agency & Action Clocks | Section 6.1 (`FR-TRG-002`) |
| **FR-TRG-003** | APNAP Push Sequencing | Section 6.1 (`FR-TRG-003`) |
| **FR-TRG-004** | Cascading Trigger Enqueuing | Section 6.1 (`FR-TRG-004`) |
| **FR-TRG-005** | Non-Resource Trigger Costs Prohibition | Section 6.1 (`FR-TRG-005`) |
| **FR-TRG-007** | Event Causality Context Lineage | Section 6.1 (`FR-TRG-007`) |
| **FR-LAY-001** | Combat Power Clamping & Layer Pipeline | Section 6.2 (`FR-LAY-001`) |
| **FR-LAY-002** | Health, Damage & Armor Invariants | Section 6.2 (`FR-LAY-002`) |
| **FR-LAY-003** | Dynamic Keywords & Duration Rules | Section 6.2 (`FR-LAY-003`) |
| **FR-REP-001** | Replacement Interception Pipeline | Section 6.3 (`FR-REP-001`) |
| **FR-REP-002** | Event Mutation & Trigger Suppression | Section 6.3 (`FR-REP-002`) |
| **FR-REP-003** | Single Application Rule | Section 6.3 (`FR-REP-003`) |
| **FR-REP-004** | 5-Stage Damage Mitigation Pipeline | Section 6.3 (`FR-REP-004`) |
| **FR-CBT-001** | Combat Eligibility & Reserves Protection | Section 7.1 (`FR-CBT-001`) |
| **FR-CBT-002** | Formalized 3-Step Combat & Retaliation | Section 7.2 (`FR-CBT-002`) |
| **FR-CBT-003** | Headquarters Defeat & Match Draws | Section 7.2 (`FR-CBT-003`) |
| **FR-KWD-002a**| Unbowed | Appendix A (`FR-KWD-002a`) |
| **FR-KWD-002b**| Resistant N | Appendix A (`FR-KWD-002b`) |
| **FR-KWD-002c**| Stalwart | Appendix A (`FR-KWD-002c`) |
| **FR-KWD-002d**| Heal N | Appendix A (`FR-KWD-002d`) |
| **FR-KWD-002e**| Cornered | Appendix A (`FR-KWD-002e`) |
| **FR-KWD-002f**| Aggressive N | Appendix A (`FR-KWD-002f`) |
| **FR-KWD-002g**| Ethereal | Appendix A (`FR-KWD-002g`) |
| **FR-KWD-002h**| Ward | Appendix A (`FR-KWD-002h`) |
| **FR-KWD-002i**| Versatile | Appendix A (`FR-KWD-002i`) |
| **FR-KWD-002j**| Guardian | Appendix A (`FR-KWD-002j`) |
| **FR-KWD-002k**| Armor N | Appendix A (`FR-KWD-002k`) |
| **FR-KWD-002l**| Deploys Exhausted | Appendix A (`FR-KWD-002l`) |
| **FR-KWD-002m**| Scratch N | Appendix A (`FR-KWD-002m`) |
| **FR-KWD-002n**| Regroup N | Appendix A (`FR-KWD-002n`) |
| **FR-KWD-002o**| Berserk | Appendix A (`FR-KWD-002o`) |