# Game Rules - Functional Requirements Document (Master FRD v3.2)
**Authoritative Game Rules, Card Taxonomy, Zone Topology, Turn Structure & Keyword Dictionary**

---

## EXECUTIVE SUMMARY & REVISION HISTORY

This document establishes the authoritative **Comprehensive Game Rules** for Card Town. It defines the rules of play, deck construction legalities, card and battlefield asset taxonomy, zone topology, resource economy, turn phases, resolution stack timing, target eligibility and fizzle rules, the 6-layer continuous modification system, replacement effects, the 5-stage damage mitigation sequence, combat resolution, and the canonical keyword dictionary.

To maintain strict separation of concerns across the specification suite:
1. **Game Rules (`01_Game_Rules_FRD.md` - this document)**: Defines *what the game rules are* in pure domain terminology, independent of JSON schemas, compiler opcodes, or software memory structures.
2. **Card Data Pipeline (`02_Card_Data_Pipeline_FRD.md`)**: Bridges the pure rules and the software engine by defining the CSV authoring schema, `blueprints.json` schema, polymorphic `AbilityDef` AST, trigger/replacement hook catalog (`FR-TRG-001`), and the atomic effect primitives (`FR-PRM-001`).
3. **Card Compiler (`03_Card_Compiler_FRD.md`)**: Translates authoring CSVs and rules text into compiled `blueprints.json` AST payloads.
4. **Client-Engine Protocol & Networking (`04_Client_Engine_Protocol_Networking_FRD.md`)**: Defines the WebSocket/Protobuf wire protocol, session handshakes, fog-of-war serialization, and `.rec` replay format.
5. **Headless Game Engine (`05_Headless_Game_Engine_FRD.md`)**: Specifies *how* the deterministic Go match kernel models entities in memory (`FR-ENT-001`–`FR-ENT-003`, `FR-SNP-101`) and executes the rules of `01` using the AST components of `02`.

### Revision Summary (v3.0 -> v3.2):
1. **Pure Rules Separation (`SOC-002`)**: Relocated data-modeling schemas, event hook enums (`FR-TRG-001`, `FR-TRG-007`), atomic opcode tables (`FR-PRM-001`), bound payload schemas (`FR-TGT-101`), and benchmark JSON ASTs into [`02_Card_Data_Pipeline_FRD.md`](02_Card_Data_Pipeline_FRD.md), and relocated runtime memory struct definitions (`FR-ENT-001`–`FR-ENT-003`, `FR-SNP-101`) into [`05_Headless_Game_Engine_FRD.md`](05_Headless_Game_Engine_FRD.md).
2. **Ready Phase Start-of-Turn Triggers (`DA-ENG-001`, `FR-TRN-001`)**: Clarified that *"At the start of your turn"* triggered abilities (*Hammerhand Bivouac*) trigger and resolve during the **Ready Phase** after controlled assets unexhaust, and Draw Phase triggers resolve during the **Draw Phase**, prior to opening the Main Phase action priority window.
3. **Pre-Game First-Player Sequencing (`DA-ENG-004`, `FR-SET-002`, `FR-SET-003`)**: Reaffirmed that First-Player Determination occurs immediately after Headquarters reveal and prior to drawing and mulliganing opening hands.
4. **Delayed Triggers, Resolution Stack Zone, Free-Deploy Zone Rules & Combat Attribution (`DA-RULES-001`–`DA-RULES-006`)**: Added `FR-TRG-006` for turn-scoped delayed triggered abilities (*Rage's Wake*, *Redline Infusion*) and Scratch Company unit-card verification in Discard (`FR-ZON-005`); formalized the **Resolution Stack** zone and discard-pile deployment rules (`FR-ZON-001`, `FR-ZON-003`, `FR-TRN-003`, `FR-TRN-003a`); specified destination zone selection and override rules for in-resolution free deployments (`FR-TRN-003a`); defined source `"deals damage"` attribution and delegated non-combat unit damage in `FR-REP-004`; and clarified attacker-only combat-kill triggers in `FR-CBT-002`.

---

## DOMAIN 1: DECK CONSTRUCTION & MATCH SETUP

### Section 1.1: Deck Building & Legalities

#### FR-SET-001: Deck Construction & Rarity Copy Limits
- **Main Deck Size**: A legal main deck must contain a minimum of 50 cards, with no upper maximum limit.
- **Headquarters Card**: Exactly 1 designated Headquarters (HQ) card is selected during deck construction and resides outside the main deck. Headquarters cards are strictly prohibited from being included in the main deck. A Headquarters card must belong to a specific faction (**Concordia**, **Soulless**, **Path**, or **Animus**) and cannot be **Neutral**.
- **Faction Deck Legality**: A main deck may contain only cards matching the faction of its designated Headquarters card and **Neutral** cards.
- **Neutral Cards**: Neutral cards may be included in any main deck regardless of the chosen Headquarters card's faction.
- **Scratch Company Deck Prohibition**: **Scratch Company** units are match-generated token entities and are strictly prohibited from being included in a starting main deck.
- **Rarity Copy Limits**: The maximum number of copies sharing the same card name permitted in a main deck is governed by rarity:
  - **Common**: Maximum of 4 copies
  - **Uncommon**: Maximum of 3 copies
  - **Rare**: Maximum of 2 copies
  - **Mythic**: Maximum of 1 copy

---

### Section 1.2: Match Initialization, Mulligan & First-Turn Rules

#### FR-SET-003: First Player Determination
- Immediately after both players' designated Headquarters cards are placed in their respective Headquarters Zones—and prior to drawing opening hands or conducting the mulligan—a random determination selects the initial decision-maker.
- The winning player has 15 seconds to choose whether to **Play First** (take Turn 1) or **Draw First** (take Turn 2). If the 15-second decision timer expires, the choice defaults to **Play First**.

#### FR-SET-002: Match Initialization & Mulligan Protocol
- **Zone Setup**: Each player's designated Headquarters card is placed in their Headquarters Zone prior to match start.
- **Opening Hand**: Following First Player Determination (`FR-SET-003`), each player's main deck is shuffled and each player draws an initial hand of 6 cards.
- **Partial Mulligan Rule**: Prior to Turn 1, each player may perform a single partial mulligan within a 25-second countdown (defaulting to keeping all 6 cards if the timer expires):
  1. Select $N$ cards ($0 \le N \le 6$) from the opening hand.
  2. Set those $N$ cards aside.
  3. Draw $N$ replacement cards from the main deck.
  4. Shuffle the $N$ set-aside cards back into the main deck.

#### FR-SET-004: First-Turn Rules & Absence of Summoning Sickness
- **Turn 1 Draw Skip**: The player taking Turn 1 skips their Draw Phase on Turn 1.
- **Immediate Readiness**: All action declarations, deployments, inherent actions, and combat declarations are permitted from Turn 1. Newly deployed assets enter the battlefield **Readied** by default (unless modified by **Deploys Exhausted**, `FR-KWD-002l`) and may immediately attack, march, generate resources, or activate abilities on the turn they enter play.

---

### Section 1.3: Deck Rules & Empty Deck Defeat

#### FR-DEK-001: Deck Search Rules
- **Visibility Scope**: Searching a deck grants the searching player full visibility of all remaining cards in that deck.
- **Unrestricted vs. Restricted Search**: If an effect instructs a player to search for *"a card"*, any card in the deck may be chosen. If filtering criteria are specified (e.g., *"an Elemental card"*), only cards matching those criteria may be chosen; if no matching card exists in the deck, the search completes with no card chosen.
- **Sequence & Mandatory Shuffle**: When a player searches a deck, they select the eligible card (if any), shuffle the remaining deck (`FR-DEK-002`), and move the selected card to its specified destination zone. Every deck search requires shuffling the searched deck afterward, even if no card was chosen.

#### FR-DEK-002: Deck Shuffling Rules
- Shuffling randomizes the order of all cards currently in the specified deck zone. Shuffling occurs as an immediate game state action and does not use the Resolution Stack (`FR-QUE-001`).

#### FR-WIN-001: Empty Deck Loss
- If a player is required to draw a card while their deck contains 0 cards, that player immediately loses the match.

---

## DOMAIN 2: CARD TAXONOMY, ENTITIES & GAME ZONES

### Section 2.1: Card Characteristics & Entity Nomenclature

#### FR-CRD-101: Printed Card Characteristics
Every card in Card Town possesses the following canonical printed characteristics:

| Characteristic | Rules Scope & Legal Values |
| :--- | :--- |
| **Name** | Printed title identifying the card for deck construction copy limits (`FR-SET-001`). |
| **Card Type** | Mutually exclusive primary classification: **Headquarters**, **Unit**, **Structure**, or **Maneuver**. |
| **Faction** | **Concordia**, **Soulless**, **Path**, **Animus**, or **Neutral**. |
| **Subtypes** | Zero or more classification tags (e.g., *Spirit*, *Horror*, *Infantry*, *Scratch Company*). |
| **Requisition Cost** | Base Requisition required to deploy or enact the card ($\ge 0$). |
| **Tactics Cost** | Base Tactics required to deploy or enact the card ($\ge 0$). |
| **Additional Costs** | Optional non-resource costs required upon declaration (e.g., sacrificing an allied unit or discarding a card; `FR-TRN-003a`). |
| **Combat Power** | Base offensive damage dealt in combat ($\ge 0$; applicable to Headquarters, Units, and Structures; not applicable to Maneuvers). |
| **Maximum Health** | Base damage capacity before destruction ($\ge 1$; applicable to Headquarters, Units, and Structures; not applicable to Maneuvers). |
| **Requisition Yield** | Requisition added to the controller's pool when the asset exhausts to generate Requisition ($\ge 0$). |
| **Tactics Yield** | Tactics added to the controller's pool when the asset exhausts to generate Tactics ($\ge 0$). |
| **Deployment Zones** | Battlefield zones where an asset may legally be deployed from hand (**Reserves**, **Main Theater**, or **Headquarters Zone**). |
| **Rarity** | **Common**, **Uncommon**, **Rare**, **Mythic**, or **Headquarters**. |
| **Rules Text** | Printed abilities, keywords, triggers, activated abilities, continuous auras, replacement effects, or maneuver effects. |

#### FR-ZON-003: Entity Nomenclature, Ownership, Control & Readiness
- **Battlefield Entities (Assets)**: Entities occupying an active battlefield zone (**Headquarters Zone**, **Reserves**, or **Main Theater**) are collectively designated as **Assets** (**Headquarters**, **Units**, and **Structures**). Rules text referring to *"assets"* or *"non-Scratch assets"* (e.g., *Bring It Down*, *Dolus Macrocannon*) includes Headquarters and Structures as well as eligible Units.
- **Non-Battlefield Entities (Cards)**: Entities residing in non-battlefield zones (**Hand**, **Deck**, **Discard**, **Attached**, or **Resolution Stack**) are referred to as **Cards** (*Unit Card*, *Structure Card*, *Maneuver Card*).
- **Ownership vs. Control**: A card's **owner** is the player whose starting deck or Headquarters Zone it began the match in; ownership never changes during a match. An asset's **controller** is the player who currently commands it on the battlefield (by default, the player who deployed it, unless modified by a control-changing effect; `FR-ZON-002`).
- **Readiness State**: Every battlefield Asset is either **Readied** (upright and available to exhaust for inherent actions or exhaust-cost abilities) or **Exhausted** (spent and unable to pay exhaust costs until readied).

---

### Section 2.2: Zone Topology & Spatial Rules

#### FR-ZON-001: Zone Definitions & Uncapped Capacity Invariant
- **Headquarters Zone**: Per-player battlefield zone housing that player's designated Headquarters asset.
- **Reserves**: Per-player backline battlefield zone for safe deployment and Requisition generation. Assets in Reserves are protected from combat attacks (`FR-CBT-001`). Standard units in Reserves exhaust to generate **Requisition** (`FR-CAP-001`).
- **Main Theater**: Shared central battlefield zone where combat engagement occurs. Standard units in the Main Theater exhaust to generate **Tactics** (`FR-CAP-001`).
- **Player Hand**: Private zone holding drawn cards awaiting play, visible only to the hand's owner unless revealed by an effect.
- **Player Deck**: Private, ordered draw pile for remaining cards in a player's deck.
- **Player Discard**: Public pile where destroyed Assets, spent Maneuvers, and discarded Cards are placed face-up.
- **Attached Zone**: Sub-container zone beneath a specific battlefield host Asset housing cards tucked under that asset (`FR-ENT-102`).
- **Resolution Stack**: Shared, ordered public zone holding declared Card deployments, Maneuver enactments, activated abilities, tactical inherent actions, and triggered abilities awaiting LIFO resolution (`FR-QUE-001`). While a declared Unit, Structure, or Maneuver card is on the stack awaiting resolution, that card physically resides in the **Resolution Stack** zone (having immediately departed its origin zone, such as **Hand** or **Discard**, upon declaration).
- **Universal Uncapped Capacity Invariant**: All zones (**Hand**, **Reserves**, **Main Theater**, and **Discard**) have **no upper maximum limit** on the number of cards or assets they may contain, and there is no maximum hand-size discard step.

> [!NOTE] Design Rationale (DA-RULES-003)
> Uncapped Reserves capacity and combat attack immunity are foundational pillars of Card Town's backline resource economy. Turtling in Reserves without deploying units to the Main Theater leaves the player's Headquarters completely vulnerable to direct attacks from opposing Main Theater units. In the absence of Guardian units in the Main Theater, aggressive opponents can close matches by Turn 4–5 through direct HQ strikes before a turtling player can deploy high-cost assets or accumulate excessive armor. Physical UI display constraints are handled through client-side card compaction and stacking, not artificial rules-level slot caps.

#### FR-ZON-002: Asset Movement, Immutability & Control Transfer
- **Headquarters & Structure Movement Immunity**: Headquarters and Structures cannot March and are permanently immune to and ineligible targets for any effect that moves an asset between battlefield zones. They remain bound to their initial battlefield zones unless destroyed (moving to Discard) or returned to Hand/Deck by a card effect.
- **Unit Battlefield Movement**: Units move between their controller's **Reserves** and the **Main Theater** via the inherent **March** action (`FR-CAP-001`) or card effects (*Ensnaring Mist*, *Piston Fist Prototype*).
- **Control Transfer Scope & Spatial Invariance**: Effects that transfer control of an opposing asset (*Rhapsody's Hollow*) apply strictly to eligible opposing units in the **Main Theater**, preventing opposing assets from occupying a player's private Reserves backline. Changing control does **not** change the asset's zone or readiness state (**Readied** remains **Readied**; **Exhausted** remains **Exhausted**). Because there is no summoning sickness (`FR-SET-004`), a Readied unit whose control is gained may immediately attack, march, or activate abilities under its new controller. If its new controller subsequently marches that unit out of the Main Theater, it moves into its **current controller's** Reserves (arriving Exhausted). If destroyed or returned to a non-battlefield zone, an asset always goes to its **owner's** Discard, Hand, or Deck.
- **Destruction**: When an Asset's Current Health drops to 0 or below (or its Maximum Health drops to 0 or below; `FR-LAY-002`), or when it is killed by an effect or cost, it is destroyed and placed face-up in its **owner's** Discard pile.

#### FR-ZON-004 & FR-ZON-101: Non-Battlefield Zone Transitions & Last-Known State
- **State Reset on Leaving the Battlefield (`FR-ZON-004`, subsumes `FR-ZON-101`)**: Whenever an Asset moves from a battlefield zone (**Headquarters Zone**, **Reserves**, or **Main Theater**) to a non-battlefield zone (**Hand**, **Discard**, or **Deck**), all accumulated damage, dynamic Armor, temporary stat modifiers, continuous aura bonuses, readiness state, and granted subtypes or keywords are immediately removed, and the card reverts to its printed characteristics.
- **Intra-Battlefield Movement Preservation**: Moving a Unit between **Reserves** and the **Main Theater** does **not** reset its state; all accumulated damage, dynamic Armor, and active modifiers are preserved across battlefield zones.
- **Last-Known Battlefield State & Trigger Scope**:
  - *Leaves-Battlefield Triggers (`"When this unit/structure is killed"`, `"When this asset leaves the battlefield"`)*: Evaluated strictly against the asset's last-known characteristics on the battlefield immediately prior to departure (`FR-ZON-004`). Because a face-down Scratch Company had no abilities on the battlefield and is not revealed until it enters Discard (after leaving the battlefield), it never triggers printed leaves-battlefield death abilities.
  - *Enter-Discard Triggers (`"When this card is put into your discard pile"`, e.g., *Solarheart Monk*)*: Evaluated against the card's revealed face-up printed characteristics upon entering the Discard pile (`FR-ZON-005`, `FR-ENT-102`). Tucked or destroyed Scratch Companies entering Discard face-up do trigger printed enter-discard abilities (provided explicit condition gates are satisfied; e.g., *Eerie Remnant* requires discarding from hand).
  - *Self-Relocating Trigger Tracking*: Any death trigger or enter-discard trigger that moves `"it"` (the triggering card itself) out of the Discard pile (such as *Unceasing Tide*: *"When this unit is killed, return it to its owner's hand"*; *Stormfront of Ideals*: *"When this unit is killed, put it on the bottom of its owner's deck"*) tracks that card in the Discard pile and fizzles (does nothing) if the card has already departed the Discard pile prior to the trigger's resolution.

#### FR-ZON-005 & FR-ENT-101: Scratch Company Fog of War & Double-Blind Lifecycle
- **Double-Blind Deployment (`FR-ZON-005`, subsumes `FR-ENT-101`)**: Deploying a **Scratch Company** moves the top card of the player's Deck directly into their **Reserves** face-down as a `1 Power / 1 Health` (`1 Requisition Yield / 1 Tactics Yield`) Unit with subtypes **Infantry** and **Scratch Company**, and no printed abilities. Its underlying card face is strictly **unknown to both players** while on the battlefield; neither player may look at its front face. Moving the top card of the deck to Reserves this way is **not** drawing a card, and attempting to deploy a Scratch Company from a 0-card deck simply creates 0 units without causing an Empty Deck Loss (`FR-WIN-001`).
- **Departure to a Public Zone (Discard)**: When a face-down Scratch Company unit is destroyed or otherwise enters a **Discard** pile, it is turned face-up, revealing its printed card identity to **both players** and immediately reverting to its printed card characteristics (`FR-ZON-004`). If a delayed triggered ability or effect attempts to deploy that card from the Discard pile (e.g., *Rage's Wake*: *"Chosen unit gets +2 power until end of turn. When that unit is killed this turn, deploy it from your discard pile to your Reserves for free."*), the engine validates whether the revealed card's printed Card Type is legally deployable to that destination zone. If the revealed card is a Unit card, it is deployed to the specified zone face-up with its printed characteristics; if the revealed card is not a Unit card (e.g., a Structure or Maneuver card), the deployment fails and the card remains in the Discard pile.
- **Departure to a Private Zone (Hand or Deck)**: When a face-down Scratch Company unit is returned to its owner's **Hand** or placed into a **Deck**, its printed card identity is revealed **strictly to its owner**, and it reverts to its normal printed card in that zone (`FR-ZON-004`).
- **Tucked Under a Host Asset (Attached Zone)**: When a face-down Scratch Company unit is tucked face-down under a host asset (`FR-ENT-102`), its underlying identity remains hidden from **both players** until it later enters a Discard pile.

#### FR-ENT-102: Tucked Cards (`Attached` Sub-Zone)
- **Tucking Rules**: Card effects may tuck a card or battlefield asset under a host battlefield Asset (into that host's **Attached** sub-zone), either face-down (as on *NO-AH-5*) or face-up. Tucking an asset removes it from the battlefield; it counts as being destroyed if and only if the effect explicitly kills it as part of the action (e.g., *NO-AH-5*: *"Kill chosen allied non-Headquarters asset and tuck it face down under this unit"*).
- **Host Departure & Discard Routing**: If a host Asset leaves the battlefield for any reason (destroyed, returned to hand, or placed into a deck), all cards tucked in its **Attached** sub-zone are immediately placed face-up into their respective **owners'** Discard piles, revealing their identities to both players. Placing tucked cards into the Discard pile this way triggers any abilities that watch for cards entering a Discard pile (*Solarheart Monk*), but does **not** count as destroying battlefield assets.

#### FR-ENT-103: Face-Down Transformation (*Forced Mutation*)
- Effects that turn a battlefield Unit face-down (*Forced Mutation*) transform that unit in place into a `1 Power / 1 Health` (`1 Requisition Yield / 1 Tactics Yield`) **Infantry Scratch Company** unit with no printed abilities, while retaining its existing readiness state, accumulated damage, dynamic Armor, and temporary turn modifiers. Because its base Maximum Health becomes 1, if its accumulated damage is greater than or equal to its new effective Maximum Health, it is immediately destroyed (`FR-LAY-002`). If it later leaves the battlefield for a private zone (**Hand** or **Deck**), its owner sees its printed face and it reverts to its printed card (`FR-ZON-004`, `FR-ZON-005`).

---

## DOMAIN 3: RESOURCE ECONOMY, TURN STRUCTURE & CAPABILITIES

### Section 3.1: Dual-Resource Economy

#### FR-RES-001: Resource Accumulation & Explicit Generation
- Players generate and spend two resources during their **Main Phase**: **Requisition** (logistics and deployment) and **Tactics** (combat positioning and maneuvers).
- **Instantaneous Economic Action**: Exhausting a Readied asset to generate its Requisition or Tactics yield (`FR-CAP-001`) resolves **immediately** without using the Resolution Stack (`FR-QUE-001`) and does not require the Resolution Stack to be empty.
- **Explicit Pre-Generation**: Players must explicitly generate resources into their resource pools before declaring card plays or ability activations; assets are never automatically exhausted to pay costs.

#### FR-RES-002: End-of-Turn Resource Depletion
- Unspent Requisition and Tactics do not carry over between turns. Both resource pools reset to `0` during Sub-Phase 4c of the End Phase (`FR-TRN-001`).

#### FR-RES-003: Requisition-to-Tactics Conversion (`2R -> 1T`)
- During their Main Phase, the active player may spend **2 Requisition** from their pool to add **1 Tactics** to their pool. This conversion is strictly one-way (Tactics cannot be converted into Requisition), resolves **immediately** without using the Resolution Stack (`FR-QUE-001`), and does not require the Resolution Stack to be empty.

---

### Section 3.2: Turn Structure & Phase Sequence

#### FR-TRN-001: Turn Phase Sequence
Each player's turn progresses through four sequential phases:

1. **Ready Phase**:
   - All **Exhausted** assets controlled by the active player across their **Reserves**, the **Main Theater**, and their **Headquarters Zone** automatically transition to **Readied**, and all once-per-turn ability limits reset for the new turn.
   - **No Passive Resource Income**: Players do **not** receive automatic or base Requisition or Tactics during the Ready Phase (`FR-RES-001`).
   - **Start-of-Turn Triggers**: All *"At the start of your turn"* triggered abilities (*Hammerhand Bivouac*) enqueue onto the Resolution Stack in APNAP order (`FR-TRG-003`) and resolve completely before the turn advances to the Draw Phase.
2. **Draw Phase**:
   - The active player draws 1 card from their Deck (skipped on Turn 1 for the player taking the first turn; `FR-SET-004`).
   - Any abilities triggered by that card draw enqueue onto the Resolution Stack in APNAP order (`FR-TRG-003`) and resolve completely before the turn advances to the Main Phase.
3. **Main Phase**:
   - **Action Priority & Turn Clock**: Whenever the Resolution Stack is empty, the active player holds priority to deploy Unit and Structure cards, enact Maneuver cards, activate custom abilities, and execute inherent actions (**March**, **Attack**, **Deploy Scratch Company**, **Generate Resources**, and **Convert Requisition**).
   - **Main Phase Turn Clock**: The Main Phase enforces a **75-second** turn clock (with a visual warning at 20 seconds remaining), backed by the player's **60-second** match Reserve Time Bank. If both the turn clock and Reserve Time Bank expire, the turn automatically transitions to the End Phase.
4. **End Phase**: Executes across four sequential sub-phases:
   - **Sub-Phase 4a (End-of-Turn Trigger Enqueuing)**: All *"At the end of your turn"* triggered abilities (including turn-scoped delayed triggers; `FR-TRG-006`) enqueue onto the Resolution Stack in APNAP order (`FR-TRG-003`).
   - **Sub-Phase 4b (Trigger & Cascade Resolution)**: All abilities on the Resolution Stack (and any cascading triggers they cause) resolve completely.
   - **Sub-Phase 4c (Resource Depletion, Modifier & Delayed Trigger Expiration)**: Unspent Requisition and Tactics pools reset to `0` (`FR-RES-002`), all temporary modifiers lasting *"until end of turn"* (or *"until the end of your opponent's turn"* when the current turn is that opponent's turn) expire (`FR-LAY-003`), and any unconsumed turn-scoped delayed triggered abilities (`FR-TRG-006`) expire.
   - **Sub-Phase 4d (Turn Handover)**: Turn priority passes to the opponent and the turn counter increments. In accordance with `FR-ZON-001`, there is no hand-size limit and no end-of-turn hand discard check.

#### FR-TRN-002: Turn-Bound Action Constraint
- Non-active players cannot manually play cards, activate abilities, or declare inherent actions during their opponent's turn. Off-turn agency is limited strictly to making choices required by triggered abilities (`FR-TRG-002`) or replacement effects (`FR-REP-001`).

---

### Section 3.3: Dynamic Costs & Inherent Asset Capabilities

#### FR-CST-001: Dynamic Deployment & Enactment Cost Evaluation
The effective resource cost to deploy an Asset card or enact a Maneuver card is calculated dynamically:
```text
Effective Cost = max(0, Base Printed Cost + Sum(Cost Increases) - Sum(Cost Reductions))
```
Effective Requisition and Tactics costs clamp at a minimum of `0`. Effects that allow a card to be deployed or enacted *"for free"* (`DEPLOY_FREE`) set its Requisition and Tactics costs to `0` and exclude all associated costs, waiving any mandatory additional non-resource costs (such as sacrificing an asset or discarding a card; `FR-TRN-003a`).

#### FR-CAP-001: Inherent Asset Capabilities Matrix
Every battlefield Asset possesses built-in capabilities governed by its Card Type and current battlefield zone. All inherent capabilities require the Asset to be **Readied** and transition that Asset to **Exhausted** upfront as their activation cost. Inherent capabilities divide into **Stack-Pushed Tactical Actions** (which require the Resolution Stack to be empty and push an action onto the stack) and **Instantaneous Economic Operations** (which resolve immediately off-stack):

- **March (Stack-Pushed Tactical Action — Units Only)**: Requires the Resolution Stack to be empty. The Unit exhausts upfront and places a March action onto the Resolution Stack. When the March action resolves, the Unit moves from its controller's **Reserves** to the **Main Theater**, or from the **Main Theater** to its controller's **Reserves**. Headquarters and Structures cannot March (`FR-ZON-002`).
- **Attack (Stack-Pushed Tactical Action — Main Theater Units & Headquarters Only)**: Requires the Resolution Stack to be empty. Permitted for **Units** currently in the **Main Theater** and **Headquarters** in the **Headquarters Zone**; prohibited for Units in Reserves and prohibited for **Structures**. The attacker exhausts upfront, declares a legal opposing combat target (`FR-CBT-001`), and places an Attack action onto the Resolution Stack (`FR-CBT-002`). Assets with effective Combat Power $\le 0$ may legally declare attacks.
- **Deploy Scratch Company (Stack-Pushed Tactical Action — Headquarters Only)**: Requires the Resolution Stack to be empty. The Headquarters exhausts upfront and places a Deploy Scratch Company action onto the Resolution Stack. When it resolves, the controller performs **Scratch 1** (`FR-ZON-005`, `FR-KWD-002m`).
- **Generate Resources (Instantaneous Economic Operation — All Assets)**: Resolves immediately without using the Resolution Stack and does not require an empty stack:
  - *Standard Units*: In **Reserves**, exhausts to generate its **Requisition Yield**; in the **Main Theater**, exhausts to generate its **Tactics Yield**.
  - *Versatile Units (`FR-KWD-002i`)*: In either **Reserves** or **Main Theater**, exhausts to generate either its **Requisition Yield** or its **Tactics Yield** (controller's choice).
  - *Structures & Headquarters*: In their battlefield zone, exhausts to generate either its **Requisition Yield** or its **Tactics Yield** (controller's choice).

> [!NOTE] Design Rationale (DA-RULES-001)
> Permitting Headquarters to attack is intentional and fully integrated into the game's resource economy. Attacking with a Headquarters requires the HQ to be Readied and immediately transitions it to Exhausted (`FR-INH-001`, `FR-CAP-001`). Because an exhausted HQ cannot generate resources or deploy a Scratch Company, spending early turns attacking with an HQ starves the player of basic resource generation, freezing their board development entirely. The opposing player is free to generate resources, deploy battlefield units, establish structures, and overwhelm the attacker. Resource development and combat declarations share the identical readiness economy; assuming a player can race with HQ attacks without suffering catastrophic board loss is fundamentally flawed.

---

## DOMAIN 4: ACTION DECLARATIONS, DEPLOYMENTS & MANEUVERS

### Section 4.1: Inherent Actions & Activated Abilities

#### FR-INH-001 through FR-INH-004: Inherent Action Rules & Execution
- **Authoritative Capabilities Matrix (`FR-INH-001`)**: All inherent asset actions are governed by the normative matrix in `FR-CAP-001`. Every inherent action requires the asset to be **Readied** and transitions that asset to **Exhausted** upfront upon declaration.
- **Tactical Actions (`FR-INH-002` Unit March & Attack, `FR-INH-004` Headquarters Attack & Deploy Scratch Company)**: Require an empty Resolution Stack (`FR-QUE-001`) and push an action onto the stack. Any abilities triggered by exhausting the asset or declaring the action enqueue on top in APNAP order (`FR-TRG-003`) and resolve before the action itself resolves.
- **Economic Operations (`FR-INH-002` Unit, `FR-INH-003` Structure, `FR-INH-004` Headquarters Resource Generation)**: Resolve immediately off-stack (`FR-RES-001`) and do not require an empty stack.
- **Deploy Scratch Company Resolution (`FR-INH-004`)**: Resolving a Headquarters Deploy Scratch Company moves the top card of the deck directly to Reserves face-down as a double-blind Scratch Company (`FR-ZON-005`). An empty deck yields 0 units without loss (`FR-WIN-001`).

#### FR-ACT-001 & FR-ACT-002: Custom Activated Abilities
- **Stack Execution & Clean-Stack Requirement (`FR-ACT-001`)**: Activating a custom ability printed on a battlefield Asset (`[Cost]: [Effect].`) requires the Resolution Stack to be empty. All resource costs, exhaust costs, and sacrifice/discard costs are paid upfront upon declaration, placing the ability onto the Resolution Stack.
- **Repeatability & Exhaust Independence (`FR-ACT-002`)**: Unless an activated ability explicitly includes exhausting the asset in its cost (`exhaust:`) or specifies a per-turn frequency limit, it does **not** require the asset to be Readied and may be activated repeatedly during the Main Phase as long as the player can pay its costs (e.g., *Dominator Warsuit*'s `R1: This unit gains 1 armor.`).

#### FR-TRN-003, FR-TRN-003a & FR-MNV-001: Card Deployments & Maneuver Enactment
- **Deploying Assets & Enacting Maneuvers (`FR-TRN-003`, `FR-MNV-001`)**: Playing a Unit or Structure card (choosing a legal deployment zone from its printed characteristics) or enacting a Maneuver card from **Hand**—or deploying a card from a non-Hand zone such as the **Discard** pile when explicitly permitted by a static ability on that card (*Veil Render*)—requires the Resolution Stack to be empty. Upon declaration, the card moves immediately from its origin zone (**Hand** or **Discard**) into the **Resolution Stack** zone (`FR-ZON-001`), and the player pays all effective resource and additional costs upfront. Because the declared card is already on the Resolution Stack before cost-triggered abilities enqueue, any abilities triggered by paying an additional cost (such as *Veil Render* itself no longer being in the Discard pile when its sacrificed ally dies) evaluate with the declared card already on the Resolution Stack rather than in its origin zone. When a Unit or Structure resolves from the stack, it enters its declared battlefield zone (**Reserves** or **Main Theater**); when a Maneuver resolves from the stack, its effects execute and the card is placed into its owner's **Discard** pile.
- **Additional Sacrifice, Discard & Exhaust Costs (`FR-TRN-003a`)**: When a card or activated ability manually declared from Hand or Discard requires sacrificing/killing an allied asset, discarding a card, or exhausting an allied unit as an additional cost (*Feed the Despoiler*, *Rhapsody's Embrace*, *Veil Render*, *Funeral Fungus*), that cost is paid immediately upon declaration. Any abilities triggered by paying that additional cost enqueue onto the Resolution Stack **on top of** the declared card or ability, resolving before it.
- **In-Resolution Free Deployments (`DEPLOY_FREE`) & Destination Zone Rules (`FR-TRN-003a`)**: When a resolving ability or Maneuver instructs a player to deploy a card *"for free"* (*World Soul Messenger*, *Light of the World Soul*, *Devilhorn Scouting Party*, *Reaver's Rush*, *Eerie Remnant*, *Mutation Swell*, *Allshan's Battle Trance*, *Rage's Wake*), that card enters the battlefield immediately during the resolution of the current effect without pushing a separate deployment item onto the Resolution Stack, without requiring the stack to be empty, and with **all resource and additional costs excluded** (`FR-CST-001`). (For example, deploying *Veil Render* for free from Discard via *Allshan's Battle Trance* or *Mutation Swell* waives its additional "kill an allied asset" cost, which applies only when manually declaring deployment via its own static permission):
  1. *Unspecified Destination Zone*: If the effect does **not** specify a destination battlefield zone (*Mutation Swell*, *World Soul Messenger*, *Light of the World Soul*, *Devilhorn Scouting Party*, *Reaver's Rush*, *Eerie Remnant*), the card must be deployed to one of its printed **Deployment Zones** (`FR-CRD-101`). If the card's printed Deployment Zones include both **Reserves** and **Main Theater**, the deploying player chooses which of those two zones the asset enters during resolution.
  2. *Explicitly Specified Destination Zone*: If the effect explicitly specifies a destination battlefield zone (*Allshan's Battle Trance*: *"Deploy chosen unit card from your discard pile to the Main Theater. It gains Stalwart and gets +2 power until end of turn."*; *Rage's Wake*: *"deploy it from your discard pile to your Reserves for free"*), the effect's explicit zone instruction overrides the card's printed Deployment Zones, placing the asset directly into the specified zone even if that zone is not among its printed Deployment Zones. Intrinsic engine validation enforces that only Unit cards may be deployed to the battlefield as units (`FR-ZON-005`).

---

## DOMAIN 5: TARGETING, FIZZLE & INTERACTIVE CHOICES

### Section 5.1: Target Eligibility & Fizzle Rules

#### FR-TGT-001 & FR-TGT-102: Target Eligibility & Conjunction Filtering
- **Conjunction Eligibility (`FR-TGT-001`, subsumes `FR-TGT-102`)**: When a card or ability requires a *"chosen"* target, an entity is a legal target if and only if it simultaneously satisfies all stated criteria:
  - **Allegiance**: *Allied* (controlled by the acting player), *Opposing* (controlled by the opponent), or unrestricted (*any*).
  - **Entity Classification & Exclusions**: *Unit*, *Structure*, *Headquarters*, *Asset* (any Unit, Structure, or Headquarters on the battlefield; `FR-ZON-003`), or *Card* in a specified non-battlefield zone. Exclusions such as *"non-Scratch"*, *"non-Headquarters"*, or *"structure or Headquarters"* restrict eligibility accordingly.
  - **Zone**: Restricted to a specified zone (e.g., *"in the Main Theater"*, *"in the same zone"*, *"from your discard pile"*) or any battlefield zone if no zone is specified.
  - **Readiness, Subtypes & Dynamic Stats**: Any required readiness state (*Readied* or *Exhausted*), subtype (*Infantry*, *Fungus*, *non-Spirit*), or stat comparison (evaluated against current effective stats in Layer 6, distinguishing between **Current Health** [`Maximum Health - Accumulated Damage`] and **Maximum Health** as specified by the card text, e.g., *Gemstone Titan* vs. *Cathedral of Self*).
- **Upfront Target Selection for Conditional Triggers**: When a triggered ability includes a targeted clause whose execution depends on an outcome determined during resolution (such as *Samsara, Living Tempest*: *"Whenever this unit attacks, look at the top 2 cards of your deck. Put one into your hand and the other into your discard pile. If the card put into your discard pile this way was a maneuver card, deal 2 damage to chosen asset."*), the target (*chosen asset*) is selected upfront when the triggered ability is placed onto the Resolution Stack. When the ability resolves, if the condition is met, the effect applies to that pre-selected target; if the condition is not met, the conditional clause simply does nothing.

#### FR-TGT-002: Target Quantifiers & Distributed Value Allocations
- **Quantifiers**: Targeting clauses may require a single mandatory target (*"chosen unit"*), an optional or bounded number of targets (*"up to 2 chosen units"*), a discrete set of targets (*"2 or 4 allied Scratch Company units"*), or apply globally without targeting (*"all assets"*).
- **Distributed Pool Allocations**: Effects that distribute a fixed pool of damage or armor across multiple targets (*Macrocannon Barrage*: *"Distribute 5 damage among up to 5 chosen units"*; *Hammerhand Bivouac*: *"Distribute 2 armor among allied units in the same zone"*) require the acting player to assign positive integer amounts to each chosen target upon declaration such that the assigned amounts sum exactly to the total pool value.

#### FR-TGT-003 & FR-TGT-103: Two-Stage Target Verification, Fizzle & Evaporation
- **Two-Stage Verification (`FR-TGT-003`, subsumes `FR-TGT-103`)**: Target legality is verified twice: first when the card or ability is placed on the Resolution Stack, and again immediately before the card or ability resolves from the top of the stack. If a target left its zone (or moved to a battlefield zone that no longer satisfies the ability's zone restriction) or no longer satisfies dynamic stat/filter criteria (e.g., *Consuming Earth* when allied structures are destroyed before resolution), that target becomes illegal.
- **Partial Invalidation**: If an effect targets multiple entities and at least one target remains legal at resolution, the effect resolves normally for the remaining legal targets and does nothing to the illegal targets.
- **Distributed Pool Evaporation**: For effects that distribute a pool of values across multiple targets (`FR-TGT-002`), any value allocated to a target that becomes illegal prior to resolution **evaporates completely** and is **not** redistributed to the remaining legal targets.
- **Complete Target Invalidation & Composite Clauses**: If all targets of a targeted clause have become illegal at resolution, that targeted clause **fizzles** (does nothing). If a card or ability is a composite effect containing both a targeted clause and a non-targeted clause (such as *All-Out War*), the invalidation of all targets fizzles **only** the targeted clause; all non-targeted clauses still resolve in full. When a Maneuver card finishes resolving (even if all of its clauses fizzled), it is placed in its owner's Discard pile; when an ability on a battlefield Asset fizzles, the Asset remains on the battlefield.
- **Deployment Triggers Without Legal Targets**: Deploying a Unit or Structure from Hand does not itself require targets on the stack; if the deployed asset has a *"When this unit/structure deploys"* triggered ability (*Saboteur*) and no legal targets exist for a targeted clause when that trigger is placed on the stack or resolves, the asset still enters the battlefield normally and only the targetless clause fizzles.

---

### Section 5.2: Interactive Choices, Fast-Path & Action Clocks

#### FR-MOD-001: Interactive Choice Rules & Deterministic Fallbacks
1. **Single-Option Fast-Path**: Whenever a required target selection or choice has **only one legal option** (and does not involve inspecting hidden cards in a Hand or Deck), that sole legal option is selected automatically without pausing the match for player confirmation.
2. **Action Clocks**: Interactive choices during stack resolution enforce a **15-second** action clock for the active player and a **10-second** action clock for the non-active player (`FR-TRG-002`), backed by the player's remaining 60-second Reserve Time Bank.
3. **Revealed Hand Inspection Window**: When an effect reveals an opponent's hand (*Drifting Concept*, *Mage Seeker's Inquisition*), the choosing player is always granted an active inspection window to view the revealed hand before the chosen card is discarded.
4. **Partitioned Deck Inspections**: When an effect inspects the top $N$ cards of a deck and partitions them across multiple destinations (*Light of the World Soul*, *Devilhorn Scouting Party*), the player assigns which inspected cards are deployed for free, placed into hand, or placed on the bottom of the deck in their chosen order within a single choice prompt.
5. **Deterministic Fallbacks on Timeout**: If both the action clock and the player's Reserve Time Bank expire on a pending choice:
   - *Top-of-Deck Inspections*: Automatically selects the first eligible card(s) in order and places the remainder on the bottom of the deck.
   - *Modal Choices*: Automatically selects the first listed legal mode(s).
   - *Optional Triggers ("you may")*: Automatically declines the optional effect.
   - *Mandatory Multi-Option Targets*: Automatically selects a random legal target.

---

## DOMAIN 6: THE RESOLUTION STACK, 6-LAYER ENGINE & DAMAGE MITIGATION

### Section 6.1: The Resolution Stack & Trigger Ordering

#### FR-QUE-001: Centralized LIFO Resolution Stack & Clean-Stack Rule
- **Last-In, First-Out (LIFO) Resolution**: Card deployments, Maneuver enactments, custom activated abilities, tactical inherent actions (**March**, **Attack**, **Deploy Scratch Company**), and triggered abilities resolve through a single shared **Resolution Stack** in Last-In, First-Out order (the top item on the stack always resolves first).
- **Clean-Stack Rule for Manual Actions**: A player may manually declare a card deployment, Maneuver enactment, custom activated ability, or tactical inherent action **only during their Main Phase while the Resolution Stack is completely empty**. Once an action and any resulting triggered abilities are placed on the stack, the stack resolves top-down automatically (pausing only for interactive choices; `FR-MOD-001`) until it is empty again.
- **Stack-Exempt Exceptions**:
  1. **Economic Resource Operations (`FR-RES-001`, `FR-RES-003`)**: Exhausting assets to generate resources and converting 2 Requisition into 1 Tactics occur immediately off-stack and do not require the Resolution Stack to be empty.
  2. **Triggered Abilities, In-Resolution Free Deployments & Replacement Effects**: Triggered abilities (`FR-TRG-003`, `FR-TRG-004`), free deployments executed inside a resolving effect (`FR-TRN-003a`), and inline replacement effects (`FR-REP-001`) occur while the stack is non-empty.

#### FR-TRG-002: Inactive Player Agency & Streamlined Clocks
- When a triggered ability or replacement effect requires a target or choice from the non-active player during the active player's turn, the non-active player is prompted with a 10-second action clock (`FR-MOD-001`), fast-pathing automatically if only one legal non-inspection option exists.

#### FR-TRG-003: APNAP Simultaneous Trigger Ordering
- Whenever a single game event causes multiple triggered abilities to trigger simultaneously, they are placed onto the Resolution Stack in **APNAP (Active Player, Non-Active Player)** order:
  1. First, the **Active Player** places all of their triggered abilities onto the stack in the order of their choosing.
  2. Second, the **Non-Active Player** places all of their triggered abilities onto the stack on top of the Active Player's triggers in the order of their choosing.
- Because the Resolution Stack resolves top-down (LIFO), the Non-Active Player's triggers resolve first, followed by the Active Player's triggers.

#### FR-TRG-004: Cascading Trigger Enqueuing
- Any new triggered abilities that trigger while an item on the Resolution Stack is resolving are buffered and placed directly on top of the Resolution Stack (in APNAP order if multiple trigger simultaneously) only **after the currently resolving stack item has completely finished executing all of its clauses and post-resolution state-based sweeps have completed** (`FR-LAY-002`). They resolve completely before the next item deeper in the stack resolves.

#### FR-TRG-005: Non-Resource Trigger Costs Prohibition
- Triggered abilities never require paying Requisition or Tactics resources to trigger or resolve, though they may require game-state conditions (such as **Cornered**) or optional non-resource actions (such as *"you may kill this structure"* on *Blind Defile*).

#### FR-TRG-006: Delayed Triggered Abilities
- **Creation & Scope**: A resolving Maneuver or ability may register a turn-scoped **delayed triggered ability** that watches for a specified event later in the current turn (such as *Rage's Wake*: *"Chosen unit gets +2 power until end of turn. When that unit is killed this turn, deploy it from your discard pile to your Reserves for free."* or *Redline Infusion*: *"At the end of your turn, kill that unit."*).
- **Cross-Zone Entity Tracking & Single Firing**: A delayed triggered ability tracks the specific watched asset (`entityId`) for the remainder of the current turn and triggers at most **once** when its trigger condition is met (enqueuing onto the Resolution Stack in APNAP order per `FR-TRG-003`). If the watched event is the destruction of that asset (*Rage's Wake*), the delayed trigger tracks the resulting card into its owner's **Discard** pile so that when the delayed trigger resolves from the stack, it verifies whether the card still resides in the Discard pile and can be legally deployed as a unit (`FR-ZON-005`). If the card left the Discard pile before the delayed trigger resolves, or if engine card-type validation fails (e.g., a destroyed face-down Scratch Company was a non-Unit card), the delayed trigger's deployment effect does nothing. (Self-relocating death triggers on printed cards follow the identical invariant; `FR-ZON-004`).
- **End-of-Turn Enqueuing & Expiration**: End-of-turn delayed triggers (*"At the end of your turn, ..."* on *Redline Infusion*) enqueue during Sub-Phase 4a (`FR-TRN-001`) alongside standard end-of-turn triggers. Any delayed triggered ability whose trigger event did not occur during the turn expires automatically during Sub-Phase 4c (`FR-TRN-001`).

---

### Section 6.2: Six-Layer Continuous Attribute Pipeline

```text
Layer 1: Base Printed Values (or 1/1/1/1 Infantry Scratch Company if Face-Down)
    │
    ▼
Layer 2: Type, Subtype & Identity Grants
    │
    ▼
Layer 3: Dynamic Keyword Grants & Suppressions
    │
    ▼
Layer 4: Continuous Auras (Static Board State)
    │
    ▼
Layer 5: Temporary Turn Modifiers ("until end of turn" / "until end of opponent's turn")
    │
    ▼
Layer 6: Mathematical Clamping (Effective Power = max(0, Raw Power))
```

#### FR-LAY-001: Six-Layer Evaluation & Exclusive Combat Power Clamping
- All battlefield Asset characteristics are evaluated continuously through the 6-layer pipeline in the order shown above (with modifiers inside the same layer applied in timestamp order of when they entered play or resolved).
- **Exclusive Combat Power Clamping**: In Layer 6, **Combat Power is the only stat that clamps at zero** (`Effective Power = max(0, Raw Power)`). Negative **Resistant** values do **not** clamp at zero (`FR-REP-004`), and **Maximum Health** / **Current Health** do **not** clamp at zero (`FR-LAY-002`).
- **One-Shot Modifier Snapshot Locking**: Instantaneous / one-shot resolving effects (from Maneuvers or abilities) that apply persistent temporary modifiers (such as Layer 2 identity grants, Layer 3 keyword grants, or Layer 5 stat modifications lasting *"until end of turn"* or *"until the end of your opponent's turn"*) based on a filtering condition (e.g., *Fire for Effect*, *March of Progress*, *Rapturous Sermon*) evaluate legal targets strictly at the moment of resolution and attach the modifier instances directly to those matching entities. Targets that become legal or enter the filtered zone after the effect resolves do **not** receive the modifier; entities that received the modifier retain it for its full duration even if they subsequently move zones (unless they leave the battlefield entirely to Hand, Discard, or Deck, which clears all modifiers under `FR-ZON-004`). In contrast, continuous static abilities (Layer 4 auras) evaluate dynamically across board state changes.

#### FR-LAY-002: Health, Damage, Dynamic Armor & State-Based Destruction Timing
- **Current Health & Lethal Destruction**: An Asset's `Current Health = Effective Maximum Health - Accumulated Damage`. Whenever an Asset's `Current Health <= 0` (`Accumulated Damage >= Effective Maximum Health`) or its `Effective Maximum Health <= 0` (from a Layer 5 debuff such as *Poisoned Rations*), that Asset is destroyed during the state-based death sweep.
- **State-Based Death Sweep Timing**: State-based death sweeps execute **strictly when there is no currently resolving item on the Resolution Stack** (i.e. immediately after a resolving card, ability, or combat action completely finishes all of its clauses, prior to enqueuing buffered triggers or resolving the next stack item), and **not** on intermediate clauses or state mutations within a resolving effect. Direct effect-based destruction (e.g., effects explicitly stating *"kill chosen asset"*) executes immediately inline as instructed by the effect. When an effect contains a conditional clause checking whether an entity was killed by that effect (e.g., *Frostfire Breath*: *"If the chosen unit is killed this way, draw a card"*), the condition evaluates whether the damage dealt by the effect reduced the asset's current health to 0 or below (lethal damage); if lethal, the asset is flagged for post-resolution state-based destruction and the conditional clause executes.
- **Temporary Maximum Health Modifiers**: Temporary Layer 5 reductions to Maximum Health reduce `Effective Maximum Health` without adding damage counters. If the Asset survives the turn, the expiration of the reduction in Sub-Phase 4c restores its Maximum Health (and thus increases its Current Health accordingly).
- **Dynamic Armor Pool**: **Armor** is a persistent, consumable defensive buffer initialized when an Asset with **Armor N** enters the battlefield or incremented when an effect grants armor (*"gains N armor"*, *"distribute N armor"*). Armor stacks additively across discrete grants, **never expires at the end of the turn**, and absorbs incoming damage 1-for-1 in Stage 4 of the damage mitigation sequence (`FR-REP-004`). Continuous Layer 4 auras never grant Armor (continuous defensive auras grant **Resistant N** or **Stalwart** instead).

#### FR-LAY-003: Duration Rules & Cornered Gate Timing
- **Full-Round Durations**: Temporary modifiers lasting *"until the end of your opponent's turn"* (*Blindside*, *Comms Relay*) remain active through the remainder of the current turn and throughout the opponent's next turn, expiring during Sub-Phase 4c of that opponent's End Phase (`FR-TRN-001`).
- **Cornered Evaluation & Snapshot Locking (`FR-KWD-002e`)**: **Cornered** is active for a player whenever their opponent controls strictly more battlefield Assets (Units and Structures combined) in the **Main Theater** than that player controls in the **Main Theater**.
  - *For Triggered Abilities*: **Cornered** is checked strictly at the moment the trigger event occurs (as an enqueue gate); once a Cornered triggered ability is placed on the Resolution Stack, it resolves even if Main Theater asset counts change before it pops.
  - *For Continuous Auras & Simultaneous Resolution*: During any simultaneous damage or resolution step (such as Step 3 combat damage and retaliation in `FR-CBT-002` or multi-target area damage), all continuous condition gates (including **Cornered**) are locked at the start of that simultaneous step and re-evaluated only after all simultaneous damage has been committed and state-based destruction checks have completed.

---

### Section 6.3: Replacement Effects & 5-Stage Damage Mitigation

#### FR-REP-001 through FR-REP-003: Replacement Effect Rules
- **Inline Event Substitution (`FR-REP-001`, `FR-REP-002`)**: Replacement effects (*"If X would happen, do Y instead"*, damage redirection, **Resistant N**, **Stalwart**, **Deploys Exhausted**) intercept an event immediately before it occurs and replace or modify it inline without using the Resolution Stack. If an event is completely replaced (e.g., prospective destruction is replaced by returning the unit to its owner's hand on *Shifting Impossibility*), the original event never occurs and triggers that watch for the original event do not fire.
- **Single Application per Discrete Event (`FR-REP-003`)**: A given replacement effect may apply at most once to each discrete event.
- **Simultaneous Replacement Conflicts & Controller Choice (`FR-REP-001`)**: If multiple replacement effects at the same stage could apply to the exact same event on an asset, the **controller of the affected asset** chooses which replacement effect applies first. Once a replacement effect modifies or replaces the event such that another replacement effect's condition is no longer applicable, the remaining replacement effect does nothing.
- **Consequential Event Isolation (`FR-REP-001`)**: Any damage or state change caused as a byproduct of a replacement effect (such as the 1 damage dealt to *Shifting Impossibility* when it replaces a Scratch Company's death) is a separate child event and cannot recursively replace its own parent event.
- **Sequential Per-Entity Death Replacement (`FR-REP-001`)**: When a single area-damage effect would destroy multiple assets simultaneously, each asset's prospective destruction is an independent interceptable death event, allowing a replacement effect like *Shifting Impossibility* to replace each allied Scratch Company's death sequentially until *Shifting Impossibility* itself takes lethal damage.
- **Damage vs. Destruction Precedence (`FR-REP-001`)**: Damage-modification replacements (Stage 1 Redirection, Stage 2 **Resistant N**, Stage 3 **Stalwart**; `FR-REP-004`) always apply before destruction-replacement effects (which apply in Stage 5 only if the resulting damage committed to Health is lethal). Damage redirected away from an asset in Stage 1 never causes that asset to face destruction.

#### FR-REP-004: 5-Stage Damage Mitigation Sequence
Every discrete instance of combat or effect damage dealt to an Asset resolves through five sequential stages:

```text
Incoming Damage Instance
    │
    ▼
[ Stage 1: Target Redirection ] ────► Moves uncommitted damage to new target (e.g. Animoceps Spore Cloud)
    │
    ▼
[ Stage 2: Flat Reductions ] ───────► Subtracts target's total Resistant N (negative Resistant amplifies)
    │
    ▼
[ Stage 3: Combat Damage Capping ] ─► Applies Stalwart to combat damage (min(dmg, 1); no-op on non-combat)
    │
    ▼
[ Stage 4: Dynamic Armor ] ─────────► Deducts damage 1-for-1 from target's current Armor pool
    │
    ▼
[ Stage 5: Health Commit ] ─────────► Remaining damage added to accumulated damage; fires damage-taken triggers
```

1. **Stage 1 (Target Redirection)**: If an active redirection effect (*Animoceps Spore Cloud*) intercepts the incoming damage, the uncommitted damage moves to the new target, and Stages 2–5 evaluate using the **new target's** defenses (`Resistant`, `Stalwart`, `Armor`, and `Health`).
2. **Stage 2 (Flat `Resistant N` Modification)**: The target's total **Resistant N** is subtracted from the incoming damage (`damage = max(0, rawDamage - Resistant)`). **Negative Resistant amplifies incoming damage**: an incoming damage instance of `0` raw damage against a target with `Resistant -1` becomes `1` damage ($0 - (-1) = 1$). Positive Resistant floors the damage at `0`.
3. **Stage 3 (`Stalwart` Combat Damage Capping)**: Applies **only if the damage is combat damage** (from an Attack or combat retaliation in Step 3 of `FR-CBT-002`) and the target has **Stalwart** (`FR-KWD-002c`). Caps damage at a maximum of `1` (`damage = min(damage, 1)`), leaving `0` damage at `0`. Non-combat damage—including damage dealt directly by a Maneuver or ability as well as **delegated unit effect damage** where a Maneuver or ability instructs a unit to deal damage equal to its power (*Soul Whip*, *Ambush Tactics*)—bypasses Stage 3 completely (and does not trigger combat retaliation), while still crediting that unit as the damage source in Stage 5. *(Executing Stage 2 before Stage 3 is an intentional design rule so that combining Resistant and Stalwart does not grant total combat invulnerability).*
4. **Stage 4 (Dynamic `Armor` Absorption)**: Any remaining damage is deducted 1-for-1 from the target's current **Armor** pool (`FR-KWD-002k`), consuming that much Armor. When a target has both **Stalwart** and **Armor**, Stalwart caps incoming combat damage to `1` in Stage 3, and Stage 4 absorbs that `1` damage by removing `1` Armor.
5. **Stage 5 (Health Commit, Damage-Taken & Damage-Dealt Triggers)**: Any remaining unabsorbed damage is marked on the Asset as accumulated damage (reducing its Current Health):
   - **Target Takes Damage (`"takes damage"` / `"is dealt damage"`)**: Abilities that trigger when a target asset takes damage (including **Unbowed**, `FR-KWD-002a`, and *"Whenever this Headquarters is dealt damage"*) trigger **if and only if at least 1 total point of damage was absorbed by Armor in Stage 4 or marked on Health in Stage 5** (`(armorAbsorbed + healthDamageCommitted) > 0`).
   - **Source Deals Damage (`"deals damage"`)**: A source asset **deals damage** to a target (firing triggers such as *Herald of Rhapsody* and *Scavenging Pack*) **if and only if `(armorAbsorbed + healthDamageCommitted) > 0` on the final post-Stage-1 target** after Stage 2 (`Resistant`) and Stage 3 (`Stalwart`). If Stage 2 reduced the damage to `0`, neither damage-taken nor damage-dealt triggers fire. If Stage 1 redirected the damage to a different asset (e.g., *Animoceps Spore Cloud*) and `(armorAbsorbed + healthDamageCommitted) > 0` on that new target, the source asset still counts as having dealt damage (to that redirected target).

---

## DOMAIN 7: COMBAT RESOLUTION RULES

### Section 7.1: Combat Eligibility & Target Priority

#### FR-CBT-001: Combat Attacker & Target Eligibility
- **Eligible Attackers**: Only Readied **Units** in the **Main Theater** and Readied **Headquarters** in the **Headquarters Zone** may declare attacks (`FR-CAP-001`). Units in **Reserves** and **Structures** in any zone cannot attack.
- **Eligible Defenders**: An attack may target only an opposing Asset (Unit or Structure) in the **Main Theater** or the opposing **Headquarters** in its **Headquarters Zone**.
- **Reserves Protection**: Assets in **Reserves** can never be attacked in combat.
- **Ethereal Immunity (`FR-KWD-002g`)**: Assets with **Ethereal** cannot be targeted by combat attacks. Ethereal overrides Guardian.
- **Direct Headquarters Attacks & Guardian Priority (`FR-KWD-002j`)**: An attacker may attack the opposing Headquarters directly even if the opponent controls assets in the Main Theater, **unless** the opponent controls one or more non-Ethereal assets with **Guardian** in the **Main Theater**. If an opponent controls any eligible **Guardian** asset in the Main Theater, all combat attacks must target one of those Guardian assets. A unit with Guardian in **Reserves** cannot be attacked and does not restrict attackers.
- **Zero-Power Attacks**: Units and Headquarters with effective Combat Power $\le 0$ may legally declare attacks (dealing `0` raw combat damage, firing attack triggers such as **Aggressive N**, and routing through `FR-REP-004` where negative **Resistant** on the defender can amplify `0` raw damage into positive damage).

> [!NOTE] Design Rationale (DA-RULES-005)
> Permitting attacks with Combat Power <= 0 is an intentional design rule. It enables keyword interactions such as Aggressive N (which grants power during the Step 2 trigger window prior to Step 3 damage resolution) and allows attacks against targets with negative Resistant to amplify damage. Retaliation is 100% symmetric, ensuring zero-power attacks are not risk-free against armed defenders.

---

### Section 7.2: 3-Step Combat Lifecycle & Retaliation

#### FR-CBT-002: 3-Step Combat Sequence & Symmetric Retaliation
Every declared combat attack resolves on the LIFO Resolution Stack (`FR-QUE-001`) across three sequential steps:

1. **Step 1 (Attack Declaration & Upfront Exhaustion)**:
   - During their Main Phase with an empty Resolution Stack, the active player chooses a Readied eligible attacker, exhausts it upfront, chooses a single legal opposing combat target (`FR-CBT-001`; plus an optional secondary opposing Main Theater target if the attacker has **Berserk**, `FR-KWD-002o`), and places the Attack action onto the bottom of the Resolution Stack.
2. **Step 2 (Pre-Combat Trigger & Reaction Window)**:
   - All abilities triggered by the attacker exhausting or by the attack being declared (such as **Aggressive N** [`FR-KWD-002f`], *Samsara, Living Tempest*, *Arkha, Sun's Fury*, or opposing defensive triggers like *Blind Defile*) enqueue onto the Resolution Stack **on top of** the Attack action in APNAP order (`FR-TRG-003`).
   - Because the Attack action is on the stack, no manual card plays or actions may be declared (`FR-QUE-001`). All triggers above the Attack action resolve top-down until the Attack action itself reaches the top of the stack.
3. **Step 3 (Simultaneous Combat Damage, Symmetric Retaliation & Post-Combat Resolution)**:
   - **Combatant Survival & Zone Check**: When the Attack action resolves from the stack, if either the **attacker** (e.g., destroyed by *Blind Defile*) or the **defender** was destroyed, removed from the battlefield, or moved out of a legal combat zone during Step 2, the Attack **fizzles immediately** and zero combat or retaliation damage is dealt.
   - **Simultaneous Damage & 100% Symmetric Retaliation**: If both the attacker and defender remain on the battlefield in legal combat zones, continuous condition gates (including **Cornered**, `FR-LAY-003`) are locked, and both assets deal combat damage to each other **simultaneously** equal to their respective effective **Combat Power**:
     - The attacker deals combat damage equal to its effective Combat Power to the defender.
     - The defender simultaneously deals combat retaliation damage equal to its effective Combat Power back to the attacker. Retaliation is **100% symmetric across all asset types** (Unit vs. Unit, Unit vs. Structure, Unit vs. Headquarters, and Headquarters vs. Headquarters).
   - **Mitigation & State-Based Destruction**: Both combat damage instances route through the 5-stage damage mitigation sequence (`FR-REP-004`; **Stalwart** applies to both strike and retaliation damage). Any combatants reduced to $\le 0$ Current Health are destroyed simultaneously (`FR-ZON-002`).
   - **Berserk Excess Damage (`FR-KWD-002o`)**: If an attacking Unit with **Berserk** destroys an opposing Asset in the **Main Theater**, any excess combat damage is immediately dealt as non-combat effect damage (routed through `FR-REP-004` followed by a state-based destruction check) to the pre-selected secondary opposing Asset in the **Main Theater**.
   - **Post-Attack & Combat-Kill Triggers**: After combat damage and state-based destructions complete, the engine emits `EVENT_COMBAT_KILL` carrying `killer_asset_id` (the asset that dealt the lethal combat damage), `victim_asset_id` (the destroyed asset), and `is_attacker` (true if the killer was the declared attacker, false if defender retaliation). Post-attack abilities (*"After this unit attacks"*, *Piston Fist Prototype*) and combat-kill abilities (*"Whenever [this asset] attacks and kills..."*, *Eternal Watchman*, *Dread Manifestation*) trigger and enqueue onto the Resolution Stack in APNAP order (`FR-TRG-003`). Combat-kill triggers evaluate `killer_asset_id` and `is_attacker` to restrict execution strictly to when that asset was the declared attacker destroying an opposing defender.

#### FR-CBT-003: Persistent Damage, Headquarters Defeat & Match Draws
- **Persistent Damage**: Damage marked on an Asset does **not** heal or clear at the end of the turn; it remains on the Asset until healed (`FR-KWD-002d`) or until the Asset leaves the battlefield (`FR-ZON-004`).
- **Headquarters Destruction Defeat**: The moment a player's **Headquarters** reaches $\le 0$ Current Health, that player immediately loses the match.
- **Simultaneous Defeat (Match Draw)**: If both players' Headquarters reach $\le 0$ Current Health during the exact same simultaneous damage step, the match ends immediately in a **Match Draw**.

---

## APPENDIX A: KEYWORD MECHANICS DICTIONARY

- **FR-KWD-002a: Unbowed**: Triggered ability (`"Unbowed: [Effect]."`). Triggers whenever this asset takes actual damage (`(armorAbsorbed + healthDamageCommitted) > 0` in `FR-REP-004`). If incoming damage is reduced to `0` in Stage 2 (**Resistant**), Unbowed does not trigger. If an asset with Unbowed takes lethal damage, its Unbowed ability still triggers and enqueues onto the Resolution Stack before the asset is destroyed. When that Unbowed trigger resolves after the asset has been destroyed:
  - Effect clauses that do **not** mutate the destroyed asset itself (e.g., *Hammerhand Brigands* `Scratch 1`, *Devilhorn Captain* `Deal 1 damage to chosen asset in the Main Theater`, *Chainbreaker Irregulars* `Draw a card, then discard a card`) resolve completely from the discard pile.
  - Effect clauses that attempt to heal or modify the destroyed asset itself (e.g., *Macabre Dancers* `Heal 2`, *Allshan, Chainbreaker Revenant* `This unit gains 2 armor`) fizzle because the asset is no longer on the battlefield (`FR-TGT-003`, `FR-ZON-004`).
- **FR-KWD-002b: Resistant N**: Static damage replacement ability (`FR-REP-004` Stage 2). Reduces each discrete instance of incoming combat or non-combat damage dealt to this asset by $N$ (to a minimum of `0` for positive $N$). If $N$ is negative (`Resistant -1`), it amplifies each incoming damage instance by $|N|$ (even amplifying `0` raw damage to $|N|$).
- **FR-KWD-002c: Stalwart**: Static combat damage replacement ability (`FR-REP-004` Stage 3). Canonical rules meaning: *"All combat damage taken is reduced to 1."* Caps any positive combat damage dealt to this asset from an Attack or combat retaliation (`FR-CBT-002`) after Stage 2 (`Resistant`) to a maximum of `1`. Stalwart has no effect on non-combat damage (from Maneuvers or abilities).
- **FR-KWD-002d: Heal N**: One-shot effect action (`"[Target] heals N."`). Removes up to $N$ accumulated damage from the target asset (cannot reduce accumulated damage below `0` or raise Current Health above Effective Maximum Health).
- **FR-KWD-002e: Cornered**: Conditional state gate (`"Cornered: [Ability]."`). Active for a player whenever their opponent controls strictly more battlefield Assets (Units and Structures combined) in the **Main Theater** than that player controls in the **Main Theater** (`FR-LAY-003`). For triggered abilities, Cornered is checked strictly when the trigger event occurs; for continuous effects during simultaneous combat/damage steps, Cornered is locked at the start of the simultaneous step.
- **FR-KWD-002f: Aggressive N**: Triggered combat ability. Meaning: *"Whenever this unit attacks, it gets +N power until end of turn."* Triggers in Step 2 of combat (`FR-CBT-002`) and resolves before Step 3 combat damage is dealt.
- **FR-KWD-002g: Ethereal**: Static combat restriction. Meaning: *"This asset cannot be attacked."* Opposing attackers cannot declare this asset as a combat target (`FR-CBT-001`). Ethereal overrides **Guardian**.
- **FR-KWD-002h: Ward**: Static targeting restriction. Meaning: *"This asset cannot be chosen by opponents except to attack."* Opponents cannot choose this asset as a target for Maneuvers, activated abilities, or triggered abilities, though it may still be attacked in combat and is still affected by non-targeted opposing effects (such as *"deal 4 damage to all assets"*).
- **FR-KWD-002i: Versatile**: Inherent resource capability modifier for Units (`FR-CAP-001`). When exhausting to generate resources, a Unit with **Versatile** may generate either its **Requisition Yield** or its **Tactics Yield** (controller's choice), regardless of whether it is in **Reserves** or the **Main Theater**.
- **FR-KWD-002j: Guardian**: Static combat restriction (`FR-CBT-001`). Meaning: *"Opposing attacks must target an asset with Guardian in the Main Theater if able."* While a player controls one or more non-Ethereal assets with **Guardian** in the **Main Theater**, opposing attackers cannot attack that player's Headquarters or non-Guardian Main Theater assets. Guardian has no effect while in **Reserves**.
- **FR-KWD-002k: Armor N**: Persistent damage-absorption pool (`FR-LAY-002`, `FR-REP-004` Stage 4). An asset with printed **Armor N** enters the battlefield with $N$ Armor counters; effects may also grant additional Armor. Armor never expires at end of turn and absorbs incoming damage 1-for-1 before Health is reduced. Damage absorbed by Armor still counts as damage taken for triggers (`FR-KWD-002a`).
- **FR-KWD-002l: Deploys Exhausted**: Static deployment replacement ability (`FR-REP-001`). This asset enters the battlefield **Exhausted** instead of **Readied**.
- **FR-KWD-002m: Scratch N**: One-shot effect action. Instructs the player to perform **Deploy Scratch Company** $N$ times sequentially, moving the top card of their Deck directly into their **Reserves** face-down as a double-blind `1/1/1/1` **Infantry Scratch Company** unit each time (`FR-ZON-005`, `FR-INH-004`). Safe on an empty deck (creates no units and does not cause Empty Deck Loss).
- **FR-KWD-002n: Regroup N**: One-shot effect action. The acting player chooses up to $N$ **Exhausted** allied assets (or all matching exhausted allied assets if specified, e.g., *"Regroup all Scratch Company units in your Reserves"*) and transitions them to **Readied**.
- **FR-KWD-002o: Berserk**: Static combat ability (`FR-CBT-002`). Whenever this Unit attacks an opposing Asset in the **Main Theater** and destroys it during Step 3 combat damage, this Unit deals its **excess combat damage** (effective combat damage dealt by this unit minus the remaining damage needed to reduce the primary target's Armor and Current Health to `0`) to another chosen opposing Asset in the **Main Theater** (selected upfront upon attack declaration if another legal opposing Main Theater asset exists) as discrete non-combat effect damage routed through `FR-REP-004`. Excess damage cannot target assets in Reserves or an opposing Headquarters, and dissipates if no other opposing asset occupies the Main Theater.

---

## APPENDIX B: MASTER RULES TRACEABILITY MATRIX

| Requirement ID | Canonical Rules Scope | Specification Location |
| :--- | :--- | :--- |
| **FR-SET-001** | Deck Construction & Rarity Copy Limits | Domain 1, Section 1.1 (`FR-SET-001`) |
| **FR-SET-002** | Match Initialization & Partial Mulligan | Domain 1, Section 1.2 (`FR-SET-002`) |
| **FR-SET-003** | First Player Determination | Domain 1, Section 1.2 (`FR-SET-003`) |
| **FR-SET-004** | First-Turn Rules & Immediate Readiness | Domain 1, Section 1.2 (`FR-SET-004`) |
| **FR-DEK-001** | Deck Search Rules & Mandatory Shuffle | Domain 1, Section 1.3 (`FR-DEK-001`) |
| **FR-DEK-002** | Deck Shuffling Rules | Domain 1, Section 1.3 (`FR-DEK-002`) |
| **FR-WIN-001** | Empty Deck Loss Condition | Domain 1, Section 1.3 (`FR-WIN-001`) |
| **FR-CRD-101** | Printed Card Characteristics | Domain 2, Section 2.1 (`FR-CRD-101`) |
| **FR-ZON-001** | Zone Definitions & Uncapped Capacity | Domain 2, Section 2.2 (`FR-ZON-001`) |
| **FR-ZON-002** | Asset Movement, Immutability & Control Transfer | Domain 2, Section 2.2 (`FR-ZON-002`) |
| **FR-ZON-003** | Entity Nomenclature (Assets vs. Cards) | Domain 2, Section 2.1 (`FR-ZON-003`) |
| **FR-ZON-004** | Non-Battlefield Zone Reset & Last-Known State | Domain 2, Section 2.2 (`FR-ZON-004`; subsumes `FR-ZON-101`) |
| **FR-ZON-101** | Clean Asset State Reset | Domain 2, Section 2.2 (Subsumed by `FR-ZON-004`) |
| **FR-ZON-005** | Scratch Company Double-Blind Lifecycle | Domain 2, Section 2.2 (`FR-ZON-005`; subsumes `FR-ENT-101`) |
| **FR-ENT-101** | Scratch Company Fog-of-War Masking | Domain 2, Section 2.2 (Subsumed by `FR-ZON-005`) |
| **FR-ENT-102** | Tucked Cards (`Attached` Sub-Zone) & Host Departure | Domain 2, Section 2.2 (`FR-ENT-102`) |
| **FR-ENT-103** | Face-Down Transformation (*Forced Mutation*) | Domain 2, Section 2.2 (`FR-ENT-103`) |
| **FR-RES-001** | Resource Accumulation & Explicit Generation | Domain 3, Section 3.1 (`FR-RES-001`) |
| **FR-RES-002** | End-of-Turn Resource Depletion | Domain 3, Section 3.1 (`FR-RES-002`) |
| **FR-RES-003** | Requisition-to-Tactics Conversion (`2R -> 1T`) | Domain 3, Section 3.1 (`FR-RES-003`) |
| **FR-TRN-001** | Turn Phase Sequence (Ready, Draw, Main, End 4a–4d) | Domain 3, Section 3.2 (`FR-TRN-001`) |
| **FR-TRN-002** | Turn-Bound Action Constraint | Domain 3, Section 3.2 (`FR-TRN-002`) |
| **FR-CST-001** | Dynamic Deployment & Enactment Cost Formula | Domain 3, Section 3.3 (`FR-CST-001`) |
| **FR-CAP-001** | Inherent Asset Capabilities Matrix | Domain 3, Section 3.3 (`FR-CAP-001`) |
| **FR-INH-001** | Inherent Action Upfront Exhaustion & Timing Split | Domain 4, Section 4.1 (`FR-INH-001`) |
| **FR-INH-002** | Unit Inherent Actions (March, Attack, Generate) | Domain 4, Section 4.1 (`FR-INH-002`) |
| **FR-INH-003** | Structure Inherent Actions (Generate Resources) | Domain 4, Section 4.1 (`FR-INH-003`) |
| **FR-INH-004** | Headquarters Inherent Actions (Attack, Scratch, Generate) | Domain 4, Section 4.1 (`FR-INH-004`) |
| **FR-ACT-001** | Custom Activated Abilities & Clean-Stack Rule | Domain 4, Section 4.1 (`FR-ACT-001`) |
| **FR-ACT-002** | Activated Ability Repeatability | Domain 4, Section 4.1 (`FR-ACT-002`) |
| **FR-TRN-003** | Deploying Asset Cards onto the Stack | Domain 4, Section 4.1 (`FR-TRN-003`) |
| **FR-TRN-003a**| Additional Sacrifice/Discard Costs & Free Deployments | Domain 4, Section 4.1 (`FR-TRN-003a`) |
| **FR-MNV-001** | Maneuver Enactment & Discard Resolution | Domain 4, Section 4.1 (`FR-MNV-001`) |
| **FR-TGT-001** | Target Eligibility & Conjunction Filtering Rules | Domain 5, Section 5.1 (`FR-TGT-001`; subsumes `FR-TGT-102`) |
| **FR-TGT-002** | Target Quantifiers & Distributed Pool Allocations | Domain 5, Section 5.1 (`FR-TGT-002`) |
| **FR-TGT-003** | Two-Stage Target Verification, Fizzle & Evaporation | Domain 5, Section 5.1 (`FR-TGT-003`; subsumes `FR-TGT-103`) |
| **FR-TGT-102** | Declarative Target Eligibility | Domain 5, Section 5.1 (Subsumed by `FR-TGT-001`) |
| **FR-TGT-103** | Target Re-Verification & Fizzle | Domain 5, Section 5.1 (Subsumed by `FR-TGT-003`) |
| **FR-MOD-001** | Interactive Choices, Fast-Path & Timeouts | Domain 5, Section 5.2 (`FR-MOD-001`) |
| **FR-QUE-001** | Centralized LIFO Stack & Clean-Stack Rule | Domain 6, Section 6.1 (`FR-QUE-001`) |
| **FR-TRG-002** | Inactive Player Agency & Streamlined Clocks | Domain 6, Section 6.1 (`FR-TRG-002`) |
| **FR-TRG-003** | APNAP Simultaneous Trigger Ordering | Domain 6, Section 6.1 (`FR-TRG-003`) |
| **FR-TRG-004** | Cascading Trigger Enqueuing | Domain 6, Section 6.1 (`FR-TRG-004`) |
| **FR-TRG-005** | Non-Resource Trigger Costs Prohibition | Domain 6, Section 6.1 (`FR-TRG-005`) |
| **FR-TRG-006** | Delayed Triggered Abilities | Domain 6, Section 6.1 (`FR-TRG-006`) |
| **FR-LAY-001** | 6-Layer Pipeline & Exclusive Combat Power Clamping | Domain 6, Section 6.2 (`FR-LAY-001`) |
| **FR-LAY-002** | Health, Damage & Dynamic Armor Rules | Domain 6, Section 6.2 (`FR-LAY-002`) |
| **FR-LAY-003** | Full-Round Durations & Cornered Timing/Locking | Domain 6, Section 6.2 (`FR-LAY-003`) |
| **FR-REP-001** | Replacement Effects, Conflict Choice & Per-Entity Death | Domain 6, Section 6.3 (`FR-REP-001`) |
| **FR-REP-002** | Inline Event Substitution & Trigger Suppression | Domain 6, Section 6.3 (`FR-REP-002`) |
| **FR-REP-003** | Single Application per Discrete Event | Domain 6, Section 6.3 (`FR-REP-003`) |
| **FR-REP-004** | 5-Stage Damage Mitigation Sequence | Domain 6, Section 6.3 (`FR-REP-004`) |
| **FR-CBT-001** | Combat Eligibility, Reserves Protection & Guardian | Domain 7, Section 7.1 (`FR-CBT-001`) |
| **FR-CBT-002** | 3-Step Combat Lifecycle & Symmetric Retaliation | Domain 7, Section 7.2 (`FR-CBT-002`) |
| **FR-CBT-003** | Persistent Damage, Headquarters Defeat & Match Draws | Domain 7, Section 7.2 (`FR-CBT-003`) |
| **FR-KWD-002a**–**002o** | Canonical Keyword Mechanics Dictionary | Appendix A (`FR-KWD-002a` through `FR-KWD-002o`) |
| *Relocated to `02`* | `FR-PRM-001`, `FR-TRG-001`, `FR-TRG-007`, `FR-TGT-101` | [`02_Card_Data_Pipeline_FRD.md`](02_Card_Data_Pipeline_FRD.md) (Modeling Bridge) |
| *Relocated to `05`* | `FR-ENT-001`, `FR-ENT-002`, `FR-ENT-003`, `FR-SNP-101` | [`05_Headless_Game_Engine_FRD.md`](05_Headless_Game_Engine_FRD.md) (Runtime Kernel) |