# Game Rules - Functional Requirements Document (Master FRD v3.19)
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

### Revision Summary (v3.18 -> v3.19): Targets Removed During Resolution, and Modal Legality (2026-10-30)
1. **Removed Targets (`SVY-R25`; `FR-TGT-003`, `FR-ZON-004`)**: a chosen asset that an earlier step or mode of the same effect removed from the battlefield is no longer that target. Later steps skip it and still act on the rest of its slot. An asset that leaves and returns is a new asset. A step whose targets are all gone did nothing (`SVY-R14`).
2. **Presence Only Between Stages (`SVY-R26`; `FR-TGT-003`)**: a step checks only that each chosen asset is still on the battlefield. Zone, stat and Ward criteria are checked only at the two verification stages.
3. **A Gone Dealer Deals Nothing (`SVY-R27`; `FR-TGT-003`, `FR-ZON-004`)**: a chosen asset that deals the damage and is gone deals none. Last-known values are only for the effect's own source (`SVY-R10`).
4. **Legal Modes (`SVY-R28`; `FR-MOD-001` §5)**: a mode is legal only if each of its exactly-one target slots has a legal entity on placement.
5. **A Maneuver Needs Its Required Targets (`SVY-R29`; `FR-MNV-001`, `FR-MOD-001` §5)**: a Maneuver can be enacted only if each of its exactly-one target slots has a legal entity, and a "Choose N" Maneuver only if N of its modes are legal.
6. **Modal Triggers With Too Few Legal Modes (`SVY-R30`; `FR-MOD-001` §5)**: the trigger uses every legal mode, and with none it does nothing.
7. **Printed Order (`SVY-R31`; `FR-MOD-001` §5)**: chosen modes resolve in printed order.

### Revision Summary (v3.17 -> v3.18): A Replacement Saves an Asset at Most Once per Sweep (2026-10-29)
1. **Once per Sweep (`SVY-R24`; `FR-REP-003`, `FR-REP-001`, `FR-LAY-002`)**: within one death sweep, a destruction replacement applies at most once to a given asset. An asset its replacement leaves on the battlefield and still lethal is taken up again in the sweep's next pass, where a different replacement applies or it is destroyed. A replacement that keeps its asset on the battlefield saves it from lethal damage only if its own steps remove that damage.

### Revision Summary (v3.16 -> v3.17): Choosing Between Replacements for One Death (2026-10-29)
1. **Choosing Between Replacements (`SVY-R23`; `FR-REP-001`, `FR-MOD-001`)**: when several replacements apply to one death, the affected asset's controller chooses one for each death. A timeout picks the first source in the fixed order for untargeted selections. With two *Shifting Impossibility* in play, the player decides which copy takes each 1 damage.

### Revision Summary (v3.15 -> v3.16): A Replacement Sees Every Death at the Same Moment (2026-10-29)
1. **Same-Moment Deaths (`SVY-R21`; `FR-REP-001`, `FR-LAY-002`)**: a destruction replacement applies to every death at the same moment as its source's own, meaning deaths during one stack item and the sweep right after it. One kill step that kills two Scratch Companies while *Shifting Impossibility* has 1 Health has both replaced.
2. **No Destruction Partway Through a Sweep (`SVY-R20`, revising `SVY-R5`; `FR-REP-001`, `FR-LAY-002`)**: a replacement applies to every death in a sweep, even when its source is lethal too or a byproduct leaves it lethal, and the source dies afterwards. The player no longer orders the replaced deaths: they go in a fixed order.
3. **Two Players' Replacements (`SVY-R22`; `FR-REP-001`)**: the active player's replaced deaths are processed first.

### Revision Summary (v3.14 -> v3.15): Combat Kills and Replaced Deaths (2026-10-28)
1. **A Replaced Death Is Not a Combat Kill (`SVY-R19`; `FR-CBT-002`, `FR-REP-001`)**: "attacks and kills" does not trigger when a replacement effect saves the defender. *Eternal Watchman* attacking a Scratch Company unit that *Shifting Impossibility* returns to hand gets no card back.
2. **Overkill Does Not Depend on the Kill (`SVY-R12`; `FR-CBT-002`, `FR-KWD-002p`)**: a wording fix. Overkill triggers from the attacker's combat damage, not from the combat-kill event.

### Revision Summary (v3.13 -> v3.14): Berserk Is Replaced by Overkill (2026-10-28)
1. **Overkill Replaces Berserk (`SVY-R12`, `KIM-O01`; `FR-KWD-002o`, `FR-KWD-002p`, `FR-CBT-002`, `FR-CAP-001`)**: Berserk N and its secondary target are retired. Overkill triggers when an attacker's combat damage leaves the defender below 0 Current Health, and its effect may use the excess. *Tempest Surge* now reads *"Overkill: Deal damage equal to the excess to chosen opposing unit in the Main Theater."*, and *Redline Infusion* grants *"Overkill: Ready this unit."* until end of turn.

### Revision Summary (v3.12 -> v3.13): A Keyword Behind Cornered Counts Only While Cornered (2026-10-26)
1. **A Keyword Behind Cornered (`SVY-R18`; `FR-KWD-002e`, `FR-KWD-002f`)**: an asset with *"Cornered: [Keyword]."* has the keyword only while its controller is Cornered, so effects that refer to the keyword ignore it otherwise. A triggered keyword is still checked for Cornered when its trigger event occurs. *Allshan, Chainbreaker Revenant* is reworded to *"Cornered: Aggressive 3."* (build board directive #60).

### Revision Summary (v3.11 -> v3.12): Scratch Company Prints Deploys Exhausted (2026-10-23)
1. **Scratch Company Prints Deploys Exhausted (`KIM-S01` revised, `KIM-S05`, `KIM-S06`; `FR-ZON-005`, `FR-KWD-002l`, `FR-KWD-002m`, `FR-SET-004`, `FR-ENT-103`)**: the face-down Scratch Company face prints Deploys Exhausted in place of the separate game rule. A scratch is a unit deploying, but it is not the Deploy action. Deploys Exhausted is decided as the asset enters, so an effect in force that removes it lets Scratch Companies enter Readied.

### Revision Summary (v3.10 -> v3.11): "If You Do" Actions, Replaced Actions, Predictable Timeouts and Exhausting (2026-10-23)
Rulings made by the user on 2026-10-23 for the wave B cards *Infested Apostle*, *Arkha, Sun's Fury*, *Chainbreaker Irregulars* and *Animoceps Shambler* (design note Open Question 8):
1. **"You May X. If You Do, Y" (`SVY-R14`; `FR-TGT-001`, `FR-TRG-005`, `FR-MOD-001`)**: "you may" is answered and Y's targets are chosen on placement. X is an untargeted selection carried out on resolution, not a cost, and Y applies only if X was carried out. A trigger's source may select itself: *Infested Apostle* may kill itself.
2. **Replaced Actions Still Count as Done (`SVY-R15`; `FR-REP-001`, `FR-TRN-003a`)**: an action or cost the player carried out counts as done even when a replacement effect changes or replaces its event. *Infested Apostle*'s kill replaced by *Shifting Impossibility* still deals its damage.
3. **No Random Timeout Fallbacks (`SVY-R16`; `FR-MOD-001`)**: mandatory targets and untargeted selections time out to the first legal options in a fixed order, not random ones. Choosing from your own Hand is not an inspection, so one card is selected automatically, and a timeout selects the card most recently added.
4. **What Counts as Exhausting (`SVY-R17`; `FR-CAP-001`)**: every change from Readied to Exhausted counts, and entering the battlefield Exhausted does not. Triggers from an action's exhaust cost go on top of the action, and triggers from resource generation are placed after it completes.
5. **Narrower Wording for Locked Targets (`SVY-R3`; `FR-TGT-001`)**: an earlier clause may decide whether a targeted clause applies, but never which entities or how many it targets. The ruling is unchanged.

### Revision Summary (v3.9 -> v3.10): Lethal Unbowed Is Placed After the Death Sweep (2026-10-23)
1. **Lethal Unbowed Is Placed After the Death Sweep (`SVY-R13`; `FR-KWD-002a`, `FR-TRG-004`)**: damage from any source, combat or non-combat, triggers Unbowed, even if the asset is destroyed by it. The trigger is buffered like any other and placed after the death sweep, in the same APNAP window (`FR-TRG-003`) as any death triggers from the same event. This replaces the earlier wording that the trigger was placed "before the asset is destroyed", which conflicted with `FR-TRG-004`. What resolves or fizzles from the discard pile is unchanged.

### Revision Summary (v3.8 -> v3.9): Optional Triggers, Source Power and Distinct "Up To" Targets (2026-10-21)
Rulings made by the user on 2026-10-21 for *Craterfall Ancient* and *Necropolis Hunt*, already pinned by the kernel's scenario tests:
1. **Optional Triggers Are Answered on Placement (`SVY-R9`, `FR-MOD-001`)**: a *"you may"* trigger asks its controller whether to use it as it is placed on the Resolution Stack, before its targets are chosen. A declined trigger is not placed: it chooses no targets and does nothing. A timeout declines.
2. **Source Power Is Read When the Step Resolves (`SVY-R10`, `FR-ZON-004`)**: *"damage equal to its power"* uses the source's current effective power when that step resolves, or its last-known effective power if it has left the battlefield.
3. **"Up To N" Targets Are Distinct (`SVY-R11`, `FR-TGT-002`)**: an *"up to N chosen"* clause cannot choose the same entity twice. Repeated choices stay a *"for each"*-only feature (`SVY-R4`).

### Revision Summary (v3.7 -> v3.8): An Attack's Targets Are Re-verified at Resolution (2026-10-21)
1. **An Attack's Targets Are Re-verified at Resolution (`SVY-R8`; `FR-CAP-001`, `FR-CBT-001`, `FR-CBT-002`, ~~`FR-KWD-002o`~~)**: the defender ~~and the Berserk secondary~~ are targets and are re-checked when the Attack resolves, including allegiance to the attacker's current controller. If either combatant changes sides, the attack fizzles ~~with no cleave~~. Guardian and Ethereal are still checked only at declaration. Ward never applies to the defender~~, and a secondary that gains Ward before Step 3 still makes the cleave dissipate (`KIM-B01`)~~. *Superseded in v3.14 (`SVY-R12`): Berserk was replaced by Overkill (`FR-KWD-002p`).*
2. **Ward Is Checked on Placement and at Resolution (`FR-TGT-003`, `FR-KWD-002h`)**: an opposing target that gains Ward before its card or ability resolves is no longer a legal target, so temporary Ward works as a response. This states ~~what `KIM-B01` already ruled for the Berserk secondary and~~ what the kernel already does for every target; Ward still never restricts attacks. *Superseded in v3.14 (`SVY-R12`): Berserk was replaced by Overkill (`FR-KWD-002p`).*

### Revision Summary (v3.6 -> v3.7): Targets Are Locked on Placement (2026-10-20)
1. **Targets Are Locked on Placement (`SVY-R3`, revised; `FR-TGT-001`)**: every target is chosen when the card or ability is placed on the Resolution Stack. This replaces v3.5's resolution-time targets for dependent clauses.
2. **Untargeted Selections (`SVY-R3`, `FR-TGT-001`, `FR-TGT-002`)**: a clause that names its entities without *"chosen"* (with *"a"*, *"an"* or a count, such as *Mutation Swell*'s *"Kill 2 or 4 allied Scratch Company units"*) does not target. Its pick is made as the clause resolves, from everything that matches then.
3. **Card text:** *Animoceps Spores* now reads *"Scratch 2, then return an allied Scratch Company unit to its owner's hand."* *Steel Rain* now makes its discard an additional cost: *"As an additional cost to enact this maneuver, discard up to 3 cards. For each card discarded this way, deal 2 damage to chosen asset. Then, if you have no cards in hand, draw 2 cards."* Repeated targets (`SVY-R4`) are unchanged, except that they are now chosen on placement.

### Revision Summary (v3.5 -> v3.6): Trigger-Order Timeout Fallback (2026-10-20)
1. **Default Order for Unanswered Trigger Ordering (`SVY-R7`, `FR-MOD-001`, `FR-TRG-003`)**: when a player does not order their own simultaneous triggers in time, they are placed in a fixed default order. This records the behavior the kernel already pins (`TestFR_TRG_003_FallbackIsCollectionOrder`), accepted by the user on 2026-10-19 and confirmed on 2026-10-20.

### Revision Summary (v3.4 -> v3.5): Starter Set Survey Rulings (2026-10-19)
Rulings on the questions raised by the [M2 starter set ability survey](../design/m2_starter_set_ability_survey.md):
1. **Deploying Is Not Moving (`SVY-R1`, `FR-ZON-002`)**: only the March action and effects that say to move something to a zone count as moving it.
2. **Targets and Costs Are Committed Together (`SVY-R2`, `FR-TGT-001`, `FR-ACT-001`)**: an asset sacrificed as a cost is not yet in the Discard pile when targets are chosen, so it cannot target itself there.
3. **Resolution-Time Targets for Dependent Clauses (`SVY-R3`, `SVY-R4`, `FR-TGT-001`)**: a targeted clause whose possible targets, or number of targets, depend on an earlier clause of the same effect is targeted as it resolves. Per-card repeats are chosen separately. *Superseded in v3.7: targets are locked on placement.*
4. **Distinct Modes (`SVY-R6`, `FR-MOD-001`)**: "Choose N" requires N different modes unless the card explicitly allows repeats.
5. **Ordering and Mid-Sweep Destruction (`SVY-R5`, `FR-REP-001`, `FR-LAY-002`)**: a replacement effect's controller orders the simultaneous events it applies to. Between events, the death sweep can destroy the replacement's source, and the remaining events then happen unreplaced.
6. **Repeated Targets (`SVY-R4`, `FR-TGT-001`)**: per-repetition target choices are independent and may pick the same entity.

### Revision Summary (v3.3 -> v3.4): Remaining Reachable Matrix Rulings (2026-10-19)
1. ~~**Berserk Secondary Is Mandatory (`KIM-B05`, `FR-KWD-002o`, `FR-CBT-002`)**: required whenever a legal secondary target exists.~~ *Superseded in v3.14 (`SVY-R12`): Berserk was replaced by Overkill (`FR-KWD-002p`).*
2. **Committed Combat Targets (`KIM-J01`, `FR-CBT-002`)**: eligibility changes after an attack is on the stack (Guardian arriving, defender gaining Ethereal) do not redirect or fizzle it.
3. ~~**Berserk on a Fizzled Attack (`KIM-B06`, `FR-CBT-002`, `FR-KWD-002o`)**: cleave still happens when the attacker alone caused the fizzle, and not when the defender is gone. *Extended in v3.8 (`SVY-R8`): nor when the defender is no longer a legal target, including no longer opposing.*~~ *Superseded in v3.14 (`SVY-R12`): Berserk was replaced by Overkill (`FR-KWD-002p`).*
4. **Zero-Yield Generation Is Legal (`KIM-S03`, `FR-CAP-001`)**.
5. **Empty-Deck Loss at the Death Sweep (`KIM-S04`, `FR-WIN-001`, `FR-CBT-003`, `FR-LAY-002`)**: same timing as HQ defeat; any two losses at one sweep are a Match Draw.

### Revision Summary (v3.2 -> v3.3): Designer Rulings on Keyword Matrix Items (2026-10-19)
1. ~~**Berserk Secondary Targeting (`KIM-B01`–`KIM-B03`, `FR-KWD-002o`, `FR-CBT-002`)**: The secondary target follows normal targeting rules with the primary target excluded. Ward blocks it; Ethereal and Guardian do not; it is re-verified before Step 3.~~ *Superseded in v3.14 (`SVY-R12`): Berserk was replaced by Overkill (`FR-KWD-002p`).*
2. ~~**Printed Berserk Always Has N (`KIM-B04`)**: *Tempest Surge* and *Redline Infusion* now read **Berserk 2** in the card data.~~ *Superseded in v3.14 (`SVY-R12`): Berserk was replaced by Overkill (`FR-KWD-002p`).*
3. **Scratch Companies Enter Exhausted (`KIM-S01`, `FR-ZON-005`, `FR-SET-004`, `FR-KWD-002m`)**.
4. **Headquarters Defeat at the Death Sweep (`KIM-S02`, `FR-CBT-003`)**: HQ defeat is checked at the next state-based death sweep rather than mid-resolution. Both HQs at $\le 0$ at the same sweep is a Match Draw.

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
- **Immediate Readiness**: All action declarations, deployments, inherent actions, and combat declarations are permitted from Turn 1. Newly deployed assets enter the battlefield **Readied** by default (unless modified by **Deploys Exhausted**, `FR-KWD-002l`, which every face-down **Scratch Company** has, `FR-ZON-005`) and may immediately attack, march, generate resources, or activate abilities on the turn they enter play.

---

### Section 1.3: Deck Rules & Empty Deck Defeat

#### FR-DEK-001: Deck Search Rules
- **Visibility Scope**: Searching a deck grants the searching player full visibility of all remaining cards in that deck.
- **Unrestricted vs. Restricted Search**: If an effect instructs a player to search for *"a card"*, any card in the deck may be chosen. If filtering criteria are specified (e.g., *"an Elemental card"*), only cards matching those criteria may be chosen; if no matching card exists in the deck, the search completes with no card chosen.
- **Sequence & Mandatory Shuffle**: When a player searches a deck, they select the eligible card (if any), shuffle the remaining deck (`FR-DEK-002`), and move the selected card to its specified destination zone. Every deck search requires shuffling the searched deck afterward, even if no card was chosen.

#### FR-DEK-002: Deck Shuffling Rules
- Shuffling randomizes the order of all cards currently in the specified deck zone. Shuffling occurs as an immediate game state action and does not use the Resolution Stack (`FR-QUE-001`).

#### FR-WIN-001: Empty Deck Loss
- If a player is required to draw a card while their deck contains 0 cards, the draw does nothing and that player loses the match at the next state-based death sweep (`FR-LAY-002`), the next time no item is resolving on the Resolution Stack. Empty-deck loss and Headquarters defeat are checked at the same sweep, so any losses detected at one sweep are simultaneous (`FR-CBT-003`).

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
- **What Counts as Moving (`SVY-R1`)**: A unit moves to a zone only through the **March** action or an effect that explicitly says to move it to a zone. Deploying a card, including a free deployment into a specified zone (`FR-TRN-003a`), means it enters the battlefield and is not a move. Abilities that trigger when a unit *"moves to a new zone"* or *"moves to the Main Theater"* (*Broken Horn Armorer*, *Chainbreaker Phalanx*) therefore do not trigger on deployment.
- **Control Transfer Scope & Spatial Invariance**: Effects that transfer control of an opposing asset (*Rhapsody's Hollow*) apply strictly to eligible opposing units in the **Main Theater**, preventing opposing assets from occupying a player's private Reserves backline. Changing control does **not** change the asset's zone or readiness state (**Readied** remains **Readied**; **Exhausted** remains **Exhausted**). Because there is no summoning sickness (`FR-SET-004`), a Readied unit whose control is gained may immediately attack, march, or activate abilities under its new controller. If its new controller subsequently marches that unit out of the Main Theater, it moves into its **current controller's** Reserves (arriving Exhausted). If destroyed or returned to a non-battlefield zone, an asset always goes to its **owner's** Discard, Hand, or Deck.
- **Destruction**: When an Asset's Current Health drops to 0 or below (or its Maximum Health drops to 0 or below; `FR-LAY-002`), or when it is killed by an effect or cost, it is destroyed and placed face-up in its **owner's** Discard pile.

#### FR-ZON-004 & FR-ZON-101: Non-Battlefield Zone Transitions & Last-Known State
- **State Reset on Leaving the Battlefield (`FR-ZON-004`, subsumes `FR-ZON-101`)**: Whenever an Asset moves from a battlefield zone (**Headquarters Zone**, **Reserves**, or **Main Theater**) to a non-battlefield zone (**Hand**, **Discard**, or **Deck**), all accumulated damage, dynamic Armor, temporary stat modifiers, continuous aura bonuses, readiness state, and granted subtypes or keywords are immediately removed, and the card reverts to its printed characteristics. An asset that leaves the battlefield and later re-enters it is a new asset. It is not the target of any card or ability that chose it before it left, including one still resolving (`SVY-R25`, `FR-TGT-003`).
- **Intra-Battlefield Movement Preservation**: Moving a Unit between **Reserves** and the **Main Theater** does **not** reset its state; all accumulated damage, dynamic Armor, and active modifiers are preserved across battlefield zones.
- **Last-Known Battlefield State & Trigger Scope**:
  - *Leaves-Battlefield Triggers (`"When this unit/structure is killed"`, `"When this asset leaves the battlefield"`)*: Evaluated strictly against the asset's last-known characteristics on the battlefield immediately prior to departure (`FR-ZON-004`). Because a face-down Scratch Company had no triggered abilities on the battlefield (its only ability is **Deploys Exhausted**) and is not revealed until it enters Discard (after leaving the battlefield), it never triggers printed leaves-battlefield death abilities.
  - *Values Read From the Source (`SVY-R10`)*: An effect that uses its source's stats (*"deal damage equal to its power"*: *Craterfall Ancient*, *Necropolis Hunt*) reads the source's current effective value (Layer 6, `FR-LAY-001`) when that step resolves, not when the card or ability was placed on the Resolution Stack. If the source has left the battlefield by then, the step reads the source's last-known effective value from immediately before it departed (`FR-ZON-004`). For example, *Necropolis Hunt* killed while it has +2 power until end of turn deals 7 damage, not 5. This applies only to the effect's own source. A chosen asset that would deal damage and has left the battlefield deals none (`SVY-R27`, `FR-TGT-003`).
  - *Enter-Discard Triggers (`"When this card is put into your discard pile"`, e.g., *Solarheart Monk*)*: Evaluated against the card's revealed face-up printed characteristics upon entering the Discard pile (`FR-ZON-005`, `FR-ENT-102`). Tucked or destroyed Scratch Companies entering Discard face-up do trigger printed enter-discard abilities (provided explicit condition gates are satisfied; e.g., *Eerie Remnant* requires discarding from hand).
  - *Self-Relocating Trigger Tracking*: Any death trigger or enter-discard trigger that moves `"it"` (the triggering card itself) out of the Discard pile (such as *Unceasing Tide*: *"When this unit is killed, return it to its owner's hand"*; *Stormfront of Ideals*: *"When this unit is killed, put it on the bottom of its owner's deck"*) tracks that card in the Discard pile and fizzles (does nothing) if the card has already departed the Discard pile prior to the trigger's resolution.

#### FR-ZON-005 & FR-ENT-101: Scratch Company Fog of War & Double-Blind Lifecycle
- **Double-Blind Deployment (`FR-ZON-005`, subsumes `FR-ENT-101`)**: Deploying a **Scratch Company** (**Scratch**, `FR-KWD-002m`) deploys the top card of the player's Deck directly into their **Reserves** face-down as a `1 Power / 1 Health` (`1 Requisition Yield / 1 Tactics Yield`) Unit with subtypes **Infantry** and **Scratch Company**, whose only printed ability is **Deploys Exhausted** (`FR-KWD-002l`, `KIM-S01`), so it enters **Exhausted**. A scratch is a unit deploying (`KIM-S05`): abilities that watch for a unit deploying see it. It is not the Deploy action of `FR-TRN-003`: no card is declared, no cost is paid, and no deploy item is placed on the Resolution Stack. It readies normally during its controller's next Ready Phase (`FR-TRN-001`) or when an effect readies it (`FR-KWD-002n`). Its underlying card face is strictly **unknown to both players** while on the battlefield; neither player may look at its front face. Moving the top card of the deck to Reserves this way is **not** drawing a card, and attempting to deploy a Scratch Company from a 0-card deck simply creates 0 units without causing an Empty Deck Loss (`FR-WIN-001`).
- **Departure to a Public Zone (Discard)**: When a face-down Scratch Company unit is destroyed or otherwise enters a **Discard** pile, it is turned face-up, revealing its printed card identity to **both players** and immediately reverting to its printed card characteristics (`FR-ZON-004`). If a delayed triggered ability or effect attempts to deploy that card from the Discard pile (e.g., *Rage's Wake*: *"Chosen unit gets +2 power until end of turn. When that unit is killed this turn, deploy it from your discard pile to your Reserves for free."*), the engine validates whether the revealed card's printed Card Type is legally deployable to that destination zone. If the revealed card is a Unit card, it is deployed to the specified zone face-up with its printed characteristics; if the revealed card is not a Unit card (e.g., a Structure or Maneuver card), the deployment fails and the card remains in the Discard pile.
- **Departure to a Private Zone (Hand or Deck)**: When a face-down Scratch Company unit is returned to its owner's **Hand** or placed into a **Deck**, its printed card identity is revealed **strictly to its owner**, and it reverts to its normal printed card in that zone (`FR-ZON-004`).
- **Tucked Under a Host Asset (Attached Zone)**: When a face-down Scratch Company unit is tucked face-down under a host asset (`FR-ENT-102`), its underlying identity remains hidden from **both players** until it later enters a Discard pile.

#### FR-ENT-102: Tucked Cards (`Attached` Sub-Zone)
- **Tucking Rules**: Card effects may tuck a card or battlefield asset under a host battlefield Asset (into that host's **Attached** sub-zone), either face-down (as on *NO-AH-5*) or face-up. Tucking an asset removes it from the battlefield; it counts as being destroyed if and only if the effect explicitly kills it as part of the action (e.g., *NO-AH-5*: *"Kill chosen allied non-Headquarters asset and tuck it face down under this unit"*).
- **Host Departure & Discard Routing**: If a host Asset leaves the battlefield for any reason (destroyed, returned to hand, or placed into a deck), all cards tucked in its **Attached** sub-zone are immediately placed face-up into their respective **owners'** Discard piles, revealing their identities to both players. Placing tucked cards into the Discard pile this way triggers any abilities that watch for cards entering a Discard pile (*Solarheart Monk*), but does **not** count as destroying battlefield assets.

#### FR-ENT-103: Face-Down Transformation (*Forced Mutation*)
- Effects that turn a battlefield Unit face-down (*Forced Mutation*) transform that unit in place into a `1 Power / 1 Health` (`1 Requisition Yield / 1 Tactics Yield`) **Infantry Scratch Company** unit whose only printed ability is **Deploys Exhausted**, while retaining its existing readiness state, accumulated damage, dynamic Armor, and temporary turn modifiers. Deploys Exhausted does nothing here: the unit is already on the battlefield and does not enter it (`FR-KWD-002l`). Because its base Maximum Health becomes 1, if its accumulated damage is greater than or equal to its new effective Maximum Health, it is immediately destroyed (`FR-LAY-002`). If it later leaves the battlefield for a private zone (**Hand** or **Deck**), its owner sees its printed face and it reverts to its printed card (`FR-ZON-004`, `FR-ZON-005`).
- **Attached Cards Purge on Mutation**: Upon face-down transformation, all cards currently tucked under the host asset in its **Attached** sub-zone (`FR-ENT-102`) are immediately purged face-up to their respective owners' Discard piles. Placing tucked cards into Discard this way emits `EVENT_DISCARD_CARD` (triggering enter-discard watchers such as *Solarheart Monk* per `FR-ENT-102`), but does not count as destroying battlefield assets.

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
   - The active player draws 1 card from their Deck (skipped on Turn 1 for the player taking the first turn; `FR-SET-004`). Because nothing is resolving, a draw from an empty deck is settled by the state-based death sweep that immediately follows (`FR-WIN-001`).
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
- **Non-Battlefield Cost Evaluation (`CardInstance` Scope)**: While the 6-layer continuous recalculation pipeline (`FR-LAY-001`) applies exclusively to battlefield Assets, continuous cost-reduction abilities and static deployment permissions operating on cards residing in non-battlefield zones (such as **Hand** on *Swirling Dustdevil* and *Boneyard Amalgam*, or **Discard** on *Veil Render*; `activeZone: HAND` or `activeZone: DISCARD`) evaluate on-demand directly on the `CardInstance` in its current zone during cost calculation and deployment legality checks without requiring full 6-layer battlefield instantiation.

#### FR-CAP-001: Inherent Asset Capabilities Matrix
Every battlefield Asset possesses built-in capabilities governed by its Card Type and current battlefield zone. All inherent capabilities require the Asset to be **Readied** and transition that Asset to **Exhausted** upfront as their activation cost. Inherent capabilities divide into **Stack-Pushed Tactical Actions** (which require the Resolution Stack to be empty and push an action onto the stack) and **Instantaneous Economic Operations** (which resolve immediately off-stack):

- **What Counts as Exhausting (`SVY-R17`)**: An asset *exhausts* when it changes from **Readied** to **Exhausted** for any reason: March, Attack, Deploy Scratch Company, generating resources, an exhaust or additional cost (`FR-ACT-001`, `FR-TRN-003a`), or an effect. An asset that enters the battlefield Exhausted (**Deploys Exhausted**, `FR-KWD-002l`, which every face-down **Scratch Company** has, `FR-ZON-005`) has not exhausted: it never was Readied. Triggers caused by exhausting are placed as follows:
  - *Stack-pushed actions* (March, Attack, Deploy Scratch Company): the asset exhausts as the action's cost, so the triggers are placed **on top of** the action and resolve before it (as `FR-CBT-002` Step 2 already says for attacks).
  - *Generating resources*: generation is off-stack (`FR-QUE-001`), so the triggers are placed after the generation has completed, and resolve before the player takes another action.
  - *Costs*: as `FR-TRN-003a`, on top of the card or ability paid for.
  - *Effects*: buffered and placed after the resolving item finishes (`FR-TRG-004`).
- **March (Stack-Pushed Tactical Action — Units Only)**: Requires the Resolution Stack to be empty. The Unit exhausts upfront and places a March action onto the Resolution Stack. When the March action resolves, the Unit moves from its controller's **Reserves** to the **Main Theater**, or from the **Main Theater** to its controller's **Reserves**. Headquarters and Structures cannot March (`FR-ZON-002`).
- **Attack (Stack-Pushed Tactical Action — Main Theater Units & Headquarters Only)**: Requires the Resolution Stack to be empty. Permitted for **Units** currently in the **Main Theater** and **Headquarters** in the **Headquarters Zone**; prohibited for Units in Reserves and prohibited for **Structures**. The attacker exhausts upfront, declares a legal opposing combat target (`FR-CBT-001`), and places an Attack action onto the Resolution Stack (`FR-CBT-002`). The target is re-verified when the Attack resolves (`FR-TGT-003`, `FR-CBT-002` Step 3). Assets with effective Combat Power $\le 0$ may legally declare attacks.
- **Deploy Scratch Company (Stack-Pushed Tactical Action — Headquarters Only)**: Requires the Resolution Stack to be empty. The Headquarters exhausts upfront and places a Deploy Scratch Company action onto the Resolution Stack. When it resolves, the controller performs **Scratch 1** (`FR-ZON-005`, `FR-KWD-002m`).
- **Generate Resources (Instantaneous Economic Operation — All Assets)**: Resolves immediately without using the Resolution Stack and does not require an empty stack:
  - *Standard Units*: In **Reserves**, exhausts to generate its **Requisition Yield**; in the **Main Theater**, exhausts to generate its **Tactics Yield**.
  - *Versatile Units (`FR-KWD-002i`)*: In either **Reserves** or **Main Theater**, exhausts to generate either its **Requisition Yield** or its **Tactics Yield** (controller's choice).
  - *Structures & Headquarters*: In their battlefield zone, exhausts to generate either its **Requisition Yield** or its **Tactics Yield** (controller's choice).
  - *Zero Yield*: An asset may exhaust to generate a resource even when its yield for that resource is `0`. It still exhausts, and any effects that modify or replace resource generation still apply.

> [!NOTE] Design Rationale (KIM-S03)
> Exhausting for a zero yield is intentionally legal. It leaves design space for replacement effects such as *"if you would generate Tactics, generate 1 instead"*, and for exhaust-watching abilities, which turn exhausting an asset with no yield into a real plan.

> [!NOTE] Design Rationale (DA-RULES-001)
> Permitting Headquarters to attack directly (including on Turn 1 when the Main Theater is empty) is intentional and fully integrated into the game's resource economy. Attacking with a Headquarters requires the HQ to be Readied and immediately transitions it to Exhausted (`FR-CAP-001`). Because an exhausted HQ cannot generate resources or deploy a Scratch Company, spending turns attacking with an HQ directly deprives the player of basic economic generation, completely stalling their early board development and card deployment. Any opposing player can capitalize on this tempo loss by generating resources, deploying assets, building an economy, and winning the damage and board race. Resource development and combat declarations share the identical readiness economy; assuming a player can race with HQ attacks without suffering catastrophic board loss is fundamentally flawed.

---

## DOMAIN 4: ACTION DECLARATIONS, DEPLOYMENTS & MANEUVERS

### Section 4.1: Inherent Actions & Activated Abilities

#### FR-CAP-001 & FR-INH-001 through FR-INH-004: Inherent Actions Subsumption
- **Subsumption under `FR-CAP-001`**: All inherent asset capabilities and execution rules (`FR-INH-001` Upfront Exhaustion & Timing Split, `FR-INH-002` Unit March, Attack & Generate, `FR-INH-003` Structure Generate, and `FR-INH-004` Headquarters Attack, Deploy Scratch Company & Generate) are canonically and authoritatively defined in `FR-CAP-001` (Section 3.3). Section 4.1 defers completely to `FR-CAP-001`; `FR-INH-001` through `FR-INH-004` are preserved as backward-compatible requirement identifiers fully subsumed by `FR-CAP-001`.

#### FR-ACT-001 & FR-ACT-002: Custom Activated Abilities
- **Stack Execution & Clean-Stack Requirement (`FR-ACT-001`)**: Activating a custom ability printed on a battlefield Asset (`[Cost]: [Effect].`) requires the Resolution Stack to be empty. All resource costs, exhaust costs, and sacrifice/discard costs are paid upfront upon declaration, placing the ability onto the Resolution Stack. Targets are chosen at the same time the costs are paid (`FR-TGT-001`), so an asset sacrificed by its own ability's cost cannot be chosen as that ability's target in the Discard pile (*Sinuous Growth*, *Ruin Partisans*).
- **Repeatability & Exhaust Independence (`FR-ACT-002`)**: Unless an activated ability explicitly includes exhausting the asset in its cost (`exhaust:`) or specifies a per-turn frequency limit, it does **not** require the asset to be Readied and may be activated repeatedly during the Main Phase as long as the player can pay its costs (e.g., *Dominator Warsuit*'s `R1: This unit gains 1 armor.`).

#### FR-TRN-003, FR-TRN-003a & FR-MNV-001: Card Deployments & Maneuver Enactment
- **Deploying Assets & Enacting Maneuvers (`FR-TRN-003`, `FR-MNV-001`)**: Playing a Unit or Structure card (choosing a legal deployment zone from its printed characteristics) or enacting a Maneuver card from **Hand**—or deploying a card from a non-Hand zone such as the **Discard** pile when explicitly permitted by a static ability on that card (*Veil Render*)—requires the Resolution Stack to be empty. Upon declaration, the card moves immediately from its origin zone (**Hand** or **Discard**) into the **Resolution Stack** zone (`FR-ZON-001`), and the player pays all effective resource and additional costs upfront. Because the declared card is already on the Resolution Stack before cost-triggered abilities enqueue, any abilities triggered by paying an additional cost (such as *Veil Render* itself no longer being in the Discard pile when its sacrificed ally dies) evaluate with the declared card already on the Resolution Stack rather than in its origin zone. When a Unit or Structure resolves from the stack, it enters its declared battlefield zone (**Reserves** or **Main Theater**); when a Maneuver resolves from the stack, its effects execute and the card is placed into its owner's **Discard** pile.
- **A Maneuver Needs Its Required Targets (`SVY-R29`, `FR-MNV-001`)**: A Maneuver can be enacted only if each of its exactly-one target slots has a legal entity when it is declared (`FR-TGT-001`). A *"Choose N"* Maneuver also needs at least N legal modes (`FR-MOD-001` §5). An *"up to N"* slot never stops it. A Maneuver that can't be enacted stays in hand. For example, *"Kill chosen unit"* can't be enacted with no unit on the battlefield, and *Reaver's Rush* can't be enacted without a unit card of power 3 or less in its controller's discard pile. Once declared, the Maneuver resolves under the usual two-stage verification (`FR-TGT-003`). A target that becomes illegal before resolution is skipped, and the card still goes to its owner's Discard pile.
- **Additional Sacrifice, Discard & Exhaust Costs (`FR-TRN-003a`)**: When a card or activated ability manually declared from Hand or Discard requires sacrificing/killing an allied asset, discarding a card, or exhausting an allied unit as an additional cost (*Feed the Despoiler*, *Rhapsody's Embrace*, *Veil Render*, *Funeral Fungus*, *Steel Rain*), that cost is paid immediately upon declaration. Any abilities triggered by paying that additional cost enqueue onto the Resolution Stack **on top of** the declared card or ability, resolving before it. A cost that a replacement effect changes or replaces still counts as paid (`SVY-R15`, `FR-REP-001`): if *Shifting Impossibility* returns a Scratch Company unit killed to pay *Rhapsody's Embrace*'s additional cost, the Maneuver still resolves.
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
- **Targets and Costs Are Committed Together (`SVY-R2`)**: Choosing targets and paying costs happen at the same moment of declaration, and target eligibility is judged against the game state at that moment. A card that moves to a new zone by paying a cost (for example, an asset sacrificed into the Discard pile) is not yet in that zone when targets are chosen, so it cannot be chosen there as a target of the card or ability it paid for.
- **Targets Are Locked on Placement (`SVY-R3`)**: Every target of a card or ability is chosen when it is placed on the Resolution Stack, together with its costs (`SVY-R2`). No target is chosen while an item resolves. An earlier clause of the same effect may decide **whether** a targeted clause applies (see Upfront Target Selection for Conditional Triggers above, and *"You May X. If You Do, Y"* below), but never **which** entities or **how many** it targets. Where the number or identity of targets is meant to depend on an earlier clause, that clause is an additional cost so that its outcome is known on declaration (`FR-TRN-003a`), or the later clause is an untargeted selection (below). Examples:
  - *Animoceps Spores* (*"Scratch 2, then return an allied Scratch Company unit to its owner's hand"*): the return is an untargeted selection, made after both new units are deployed. It may pick one of them or any other allied Scratch Company unit.
  - *Steel Rain* (*"As an additional cost to enact this maneuver, discard up to 3 cards. For each card discarded this way, deal 2 damage to chosen asset."*): the discard is paid on declaration, so the number of targets is known then, and one asset is chosen for each card discarded at that same moment.
- **Untargeted Selections (`SVY-R3`)**: A clause that names its entities without *"chosen"*, with *"a"*, *"an"* or a count (*"return an allied Scratch Company unit"*, *"Kill 2 or 4 allied Scratch Company units"*), does not target. The acting player picks among every entity that matches when the clause resolves, after the earlier clauses of the same effect have finished, and no other item resolves in between. Because the pick is not a target, targeting restrictions such as Ward (`FR-KWD-002h`) and the two-stage verification of `FR-TGT-003` do not apply to it. If nothing matches, only that clause does nothing.
- **"You May X. If You Do, Y" (`SVY-R14`)**: For a triggered ability of the form *"you may [X]. If you do, [Y]"* (*Infested Apostle*: *"When this unit deploys, you may kill an allied unit. If you do, deal 2 damage to chosen opposing unit."*; *Arkha, Sun's Fury*), the controller answers *"you may"* as the trigger is placed (`SVY-R9`, `FR-MOD-001` §7), and Y's targets are chosen at the same moment (`SVY-R3`). X is carried out when the trigger resolves, as an untargeted selection among the entities that match it at that moment. X is not a cost (`FR-TRG-005`). Y applies only if X was carried out (`SVY-R15`). If nothing matches X on resolution, or the controller does not answer X's selection in time (`FR-MOD-001` §6), X is not carried out and Y does nothing. The trigger's own source may be selected when it matches X: *Infested Apostle* may kill itself, and its trigger still resolves. Assets destroyed by X are destroyed while the trigger resolves, so their death triggers are placed after it has finished (`FR-TRG-004`), and Y resolves before them.
- **Repeated Targets (`SVY-R4`)**: A targeted clause repeated *"for each"* outcome makes a separate target choice for each repetition, all on placement. The choices are independent, so the same entity may be chosen more than once, and each repetition is its own damage instance (`FR-REP-004`). For example, *Steel Rain* with 3 cards discarded may put all three 2-damage instances on one asset.

#### FR-TGT-002: Target Quantifiers & Distributed Value Allocations
- **Quantifiers**: Targeting clauses may require a single mandatory target (*"chosen unit"*), an optional or bounded number of targets (*"up to 2 chosen units"*), or apply globally without targeting (*"all assets"*). A count without *"chosen"* is an untargeted selection, not a target (`SVY-R3`): *Mutation Swell*'s *"Kill 2 or 4 allied Scratch Company units"* picks its units as it resolves.
- **"Up To N" Targets Are Distinct (`SVY-R11`)**: An *"up to N chosen"* clause (*Craterfall Ancient*: *"each of up to 2 other chosen assets"*) chooses N or fewer **different** entities. The same entity cannot be chosen twice. Only a clause repeated *"for each"* outcome may choose the same entity more than once (`SVY-R4`, `FR-TGT-001`).
- **Distributed Pool Allocations**: Effects that distribute a fixed pool of damage or armor across multiple targets (*Macrocannon Barrage*: *"Distribute 5 damage among up to 5 chosen units"*; *Hammerhand Bivouac*: *"Distribute 2 armor among allied units in the same zone"*) require the acting player to assign positive integer amounts to each chosen target upon declaration such that the assigned amounts sum exactly to the total pool value.

#### FR-TGT-003 & FR-TGT-103: Two-Stage Target Verification, Fizzle & Evaporation
- **Two-Stage Verification (`FR-TGT-003`, subsumes `FR-TGT-103`)**: Target legality is verified twice: first when the card or ability is placed on the Resolution Stack, and again immediately before the card or ability resolves from the top of the stack. If a target left its zone (or moved to a battlefield zone that no longer satisfies the ability's zone restriction) or no longer satisfies dynamic stat/filter criteria (e.g., *Consuming Earth* when allied structures are destroyed before resolution), that target becomes illegal. A target that has gained **Ward** (`FR-KWD-002h`) against the acting player since placement is likewise illegal; Ward is checked at both stages.
- **Partial Invalidation**: If an effect targets multiple entities and at least one target remains legal at resolution, the effect resolves normally for the remaining legal targets and does nothing to the illegal targets.
- **Targets Removed During Resolution (`SVY-R25`, `SVY-R26`)**: A chosen asset that an earlier step or mode of the same effect removed from the battlefield is no longer that target. Each later step does nothing to it and still acts on the rest of its slot (Partial Invalidation, above). An asset that leaves the battlefield and returns during the same resolution is a new asset (`FR-ZON-004`) and is not that target either. Between the two verification stages, a step checks only that each chosen asset is still on the battlefield. It does not re-check the slot's zone, stat or Ward criteria, so an earlier step that moves a target to another battlefield zone or changes its stats does not stop a later step from acting on it. *Example:* *Ambush Tactics* choosing *"Kill chosen unit"* and *"Put chosen unit on top of its owner's deck"* for the same unit kills it. The put-on-deck mode does nothing.
- **A Chosen Asset That Deals Damage (`SVY-R27`)**: When a chosen asset deals the damage (*Ambush Tactics*: *"Chosen unit deals damage equal to its power to another chosen unit"*), it is a target like any other. If it was illegal at the second verification stage, or an earlier step removed it, it deals no damage. Its last-known power is not used, because last-known values are read only for the effect's own source (`SVY-R10`, `FR-ZON-004`). Otherwise it deals damage equal to its current effective power (Layer 6, `FR-LAY-001`) as the step resolves. *Example:* *Ambush Tactics* choosing *"Return chosen unit to its owner's hand"* for unit X and *"Chosen unit deals damage equal to its power to another chosen unit"* with X dealing: X returns to its owner's hand and deals no damage.
- **Distributed Pool Evaporation**: For effects that distribute a pool of values across multiple targets (`FR-TGT-002`), any value allocated to a target that becomes illegal prior to resolution **evaporates completely** and is **not** redistributed to the remaining legal targets.
- **Complete Target Invalidation & Composite Clauses**: If all targets of a targeted clause have become illegal at resolution, that targeted clause **fizzles** (does nothing). A targeted step whose chosen assets are all gone when it resolves, whether at the second verification stage or removed by an earlier step (`SVY-R25`), does nothing, so a step that applies only if it was done (*"If you do"*, `SVY-R14`) does not apply. If a card or ability is a composite effect containing both a targeted clause and a non-targeted clause (such as *All-Out War*), the invalidation of all targets fizzles **only** the targeted clause; all non-targeted clauses still resolve in full. When a Maneuver card finishes resolving (even if all of its clauses fizzled), it is placed in its owner's Discard pile; when an ability on a battlefield Asset fizzles, the Asset remains on the battlefield.
- **Deployment Triggers Without Legal Targets**: Unlike enacting a Maneuver (`SVY-R29`, `FR-MNV-001`), deploying a Unit or Structure from Hand does not itself require targets on the stack; if the deployed asset has a *"When this unit/structure deploys"* triggered ability (*Saboteur*) and no legal targets exist for a targeted clause when that trigger is placed on the stack or resolves, the asset still enters the battlefield normally and only the targetless clause fizzles.

---

### Section 5.2: Interactive Choices, Fast-Path & Action Clocks

#### FR-MOD-001: Interactive Choice Rules & Deterministic Fallbacks
1. **Single-Option Fast-Path**: Whenever a required target selection or choice has **only one legal option** (and does not involve inspecting hidden cards in a Hand or Deck), that sole legal option is selected automatically without pausing the match for player confirmation. Choosing a card from your own Hand (*Chainbreaker Irregulars*: *"discard a card"*) is not an inspection of hidden cards. If only one card qualifies, it is selected automatically (`SVY-R16`).
2. **Action Clocks**: Interactive choices during stack resolution enforce a **15-second** action clock for the active player and a **10-second** action clock for the non-active player (`FR-TRG-002`), backed by the player's remaining 60-second Reserve Time Bank.
3. **Revealed Hand Inspection Window**: When an effect reveals an opponent's hand (*Drifting Concept*, *Mage Seeker's Inquisition*), the choosing player is always granted an active inspection window to view the revealed hand before the chosen card is discarded.
4. **Partitioned Deck Inspections**: When an effect inspects the top $N$ cards of a deck and partitions them across multiple destinations (*Light of the World Soul*, *Devilhorn Scouting Party*), the player assigns which inspected cards are deployed for free, placed into hand, or placed on the bottom of the deck in their chosen order within a single choice prompt.
5. **Distinct Modes (`SVY-R6`)**: A modal card or ability that says *"Choose N"* (*Ambush Tactics*) requires choosing N **different** modes. The same mode may be chosen more than once only if the card's text explicitly allows it. A mode is **legal** only if each of its exactly-one target slots has a legal entity when the effect is placed. An *"up to N"* slot never makes a mode illegal (`SVY-R28`). Only legal modes may be chosen. A *"Choose N"* Maneuver can be enacted only if at least N of its modes are legal (`SVY-R29`, `FR-MNV-001`). A triggered ability must still be placed. With fewer than N legal modes it uses every legal one, and with none it is placed with no mode and does nothing, as a deployment trigger without legal targets does (`SVY-R30`, `FR-TGT-003`). The chosen modes resolve in printed order, whatever order they were picked in (`SVY-R31`). So *Ambush Tactics*' kill and return modes run before its damage mode, and its put-on-deck mode runs after it.
6. **Deterministic Fallbacks on Timeout**: No fallback is random, so a player can always predict what a timeout does (`SVY-R16`). If both the action clock and the player's Reserve Time Bank expire on a pending choice:
   - *Top-of-Deck Inspections*: Automatically selects the first eligible card(s) in order and places the remainder on the bottom of the deck.
   - *Modal Choices*: Automatically selects the first listed legal mode(s).
   - *Optional Triggers ("you may")*: Automatically declines the optional effect, so the trigger is not placed (§7, `SVY-R9`).
   - *Mandatory Targets and Untargeted Selections (`SVY-R16`)*: Selects the required number of legal options, the first ones in **fixed order**. Assets: the choosing player's own assets first, then the opponent's. Within each player, the Headquarters first, then Reserves, then the Main Theater, each zone in the order its assets entered it (as `SVY-R7`). Cards in a pile such as the Discard pile: from the top down. An *"up to N"* selection selects none. The selection for X in *"you may X. If you do, Y"* is optional and also selects none, so X is not carried out (`SVY-R14`).
   - *Choosing Cards from Your Own Hand (`SVY-R16`)*: Selects the card most recently added to the Hand. For *"Draw a card, then discard a card"* this is the card just drawn, including when the Hand was empty before the draw.
   - *Ordering Own Simultaneous Triggers (`SVY-R7`, `FR-TRG-003`)*: Places that player's triggers in a fixed default order, not a random one. Triggers from different sources go by source position: the Headquarters first, then Reserves, then the Main Theater, each zone in the order its assets entered it. A single source's triggers go keyword triggers (such as **Aggressive N**) first, then its printed triggers in printed order. The first trigger in this order is placed first, so it resolves last.
   - *Choosing Between Replacement Effects (`SVY-R23`, `FR-REP-001`)*: Applies the replacement whose source comes first in the fixed order for untargeted selections above: the choosing player's own assets first, then the opponent's; within each player, the Headquarters first, then Reserves, then the Main Theater, each zone in the order its assets entered it. A single source's replacements go in printed order.
7. **Optional Triggers Are Answered on Placement (`SVY-R9`)**: A triggered ability that says *"you may"* (*Craterfall Ancient*) asks its controller whether to use it as it is placed on the Resolution Stack, before its targets are chosen. This matches `SVY-R2` and `SVY-R3`: every commitment is made on placement (`FR-TGT-001`). If the controller accepts, the trigger's targets are chosen and it is placed as usual. If they decline, it is not placed: it chooses no targets and does nothing. A timeout declines (§6).

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
- Triggered abilities never require paying Requisition or Tactics resources to trigger or resolve. They may require game-state conditions (such as **Cornered**) or offer optional non-resource actions (*"you may kill this structure"* on *Blind Defile*; *"you may kill an allied unit"* on *Infested Apostle*). Such an action is not a cost: it is carried out when the trigger resolves (`SVY-R14`).

#### FR-TRG-006: Delayed Triggered Abilities
- **Creation & Scope**: A resolving Maneuver or ability may register a turn-scoped **delayed triggered ability** that watches for a specified event later in the current turn (such as *Rage's Wake*: *"Chosen unit gets +2 power until end of turn. When that unit is killed this turn, deploy it from your discard pile to your Reserves for free."* or *Redline Infusion*: *"At the end of your turn, kill that unit."*).
- **Cross-Zone Entity Tracking & Single Firing**: A delayed triggered ability tracks the specific watched asset (`entityId`) while on the battlefield and binds to the persistent underlying `CardInstanceId` upon battlefield departure or asset wrapper deallocation (`FR-ENT-003`). It triggers at most **once** when its trigger condition is met (enqueuing onto the Resolution Stack in APNAP order per `FR-TRG-003`). If the watched event is the destruction of that asset (*Rage's Wake*), the delayed trigger tracks that specific `CardInstanceId` into its owner's **Discard** pile (disambiguating it from any other identical blueprint cards already in Discard). When the delayed trigger resolves from the stack, it verifies whether that exact `CardInstanceId` still resides in the Discard pile and can be legally deployed as a unit (`FR-ZON-005`). If the card left the Discard pile before the delayed trigger resolves, or if engine card-type validation fails (e.g., a destroyed face-down Scratch Company was a non-Unit card), the delayed trigger's deployment effect does nothing. (Self-relocating death triggers on printed cards follow the identical invariant; `FR-ZON-004`).
- **End-of-Turn Enqueuing & Expiration**: End-of-turn delayed triggers (*"At the end of your turn, ..."* on *Redline Infusion*) enqueue during Sub-Phase 4a (`FR-TRN-001`) alongside standard end-of-turn triggers. Any delayed triggered ability whose trigger event did not occur during the turn expires automatically during Sub-Phase 4c (`FR-TRN-001`).

---

### Section 6.2: Six-Layer Continuous Attribute Pipeline

```text
Layer 1: Base Printed Values (or 1/1/1/1 Infantry Scratch Company with Deploys Exhausted if Face-Down)
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
- **Non-Battlefield Evaluation Boundary**: Non-battlefield cards (`CardInstance` in **Hand** or **Discard**) do not participate in continuous 6-layer recalculation; static cost reductions and zone-play permissions (`activeZone: HAND / DISCARD`) evaluate on-demand directly during action/cost declaration (`FR-CST-001`, `FR-TRN-003`).
- **Exclusive Combat Power Clamping**: In Layer 6, **Combat Power is the only stat that clamps at zero** (`Effective Power = max(0, Raw Power)`). Negative **Resistant** values do **not** clamp at zero (`FR-REP-004`), and **Maximum Health** / **Current Health** do **not** clamp at zero (`FR-LAY-002`).
- **One-Shot Modifier Snapshot Locking**: Instantaneous / one-shot resolving effects (from Maneuvers or abilities) that apply persistent temporary modifiers (such as Layer 2 identity grants, Layer 3 keyword grants, or Layer 5 stat modifications lasting *"until end of turn"* or *"until the end of your opponent's turn"*) based on a filtering condition (e.g., *Fire for Effect*, *March of Progress*, *Rapturous Sermon*) evaluate legal targets strictly at the moment of resolution and attach the modifier instances directly to those matching entities. Targets that become legal or enter the filtered zone after the effect resolves do **not** receive the modifier; entities that received the modifier retain it for its full duration even if they subsequently move zones (unless they leave the battlefield entirely to Hand, Discard, or Deck, which clears all modifiers under `FR-ZON-004`). In contrast, continuous static abilities (Layer 4 auras) evaluate dynamically across board state changes.

#### FR-LAY-002: Health, Damage, Dynamic Armor & State-Based Destruction Timing
- **Current Health & Lethal Destruction**: An Asset's `Current Health = Effective Maximum Health - Accumulated Damage`. Whenever an Asset's `Current Health <= 0` (`Accumulated Damage >= Effective Maximum Health`) or its `Effective Maximum Health <= 0` (from a Layer 5 debuff such as *Poisoned Rations*), that Asset is destroyed during the state-based death sweep.
- **State-Based Death Sweep Timing**: State-based death sweeps (which also apply the match-loss checks of `FR-CBT-003` and `FR-WIN-001`) execute **strictly when there is no currently resolving item on the Resolution Stack** (i.e. immediately after a resolving card, ability, or combat action completely finishes all of its clauses, prior to enqueuing buffered triggers or resolving the next stack item), and **not** on intermediate clauses or state mutations within a resolving effect. Direct effect-based destruction (e.g., effects explicitly stating *"kill chosen asset"*) executes immediately inline as instructed by the effect. When an effect contains a conditional clause checking whether an entity was killed by that effect (e.g., *Frostfire Breath*: *"If the chosen unit is killed this way, draw a card"*), the condition evaluates whether the damage dealt by the effect reduced the asset's current health to 0 or below (lethal damage); if lethal, the asset is flagged for post-resolution state-based destruction and the conditional clause executes.
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
- **Single Application per Discrete Event (`FR-REP-003`, `SVY-R24`)**: A given replacement effect may apply at most once to each discrete event. For a destruction replacement, an asset's death is one event for each kill step and for each death sweep. Within one sweep, a replacement applies at most once to a given asset, however many of the sweep's passes take that asset up (`FR-LAY-002`).
- **Simultaneous Replacement Conflicts & Controller Choice (`FR-REP-001`, `SVY-R23`)**: If multiple replacement effects at the same stage could apply to the exact same event on an asset, the **controller of the affected asset** chooses which replacement effect applies first. They choose separately for each event, including each of several deaths at the same moment (`SVY-R20`, `SVY-R21`), and a source that is itself lethal is still a legal choice. When only one replacement applies, nothing is asked. Once a replacement effect modifies or replaces the event such that another replacement effect's condition is no longer applicable, the remaining replacement effect does nothing. A timeout follows `FR-MOD-001` §6. *Example:* two *Shifting Impossibility* are in play, and three allied Scratch Companies are lethal at one sweep. Their controller chooses, for each death, which *Shifting Impossibility* returns that unit and takes the 1 damage.
- **Consequential Event Isolation (`FR-REP-001`)**: Any damage or state change caused as a byproduct of a replacement effect (such as the 1 damage dealt to *Shifting Impossibility* when it replaces a Scratch Company's death) is a separate child event and cannot recursively replace its own parent event.
- **Replaced Actions Still Count as Done (`SVY-R15`, `FR-REP-001`)**: An action that a card instructs a player to take, or that a player pays as a cost, counts as carried out once the player has carried it out, even if a replacement effect changes or completely replaces the event it would cause. Completely replacing an event still means the original event never occurs, and triggers that watch for it do not fire (Inline Event Substitution, above). What still counts is the player's action, for *"if you do"* (`SVY-R14`) and for costs (`FR-TRN-003a`). Example: *Infested Apostle* kills an allied Scratch Company unit, and *Shifting Impossibility* returns that unit to its owner's hand instead. The unit is not destroyed and no death trigger fires, but the kill was carried out, so the Apostle deals its 2 damage. This applies to every kill an effect makes: if *Mutation Swell* kills four allied Scratch Company units and *Shifting Impossibility* replaces all four, four units were killed this way, and its controller deploys two unit cards from their discard pile (`SVY-R21`).
- **Per-Entity Death Replacement (`FR-REP-001`)**: When one effect or one death sweep would destroy several assets at once, each asset's prospective destruction is a separate death event, and a replacement effect such as *Shifting Impossibility* applies to each one it matches.
- **Same-Moment Deaths (`SVY-R21`, `SVY-R20`, `FR-REP-001`, `FR-LAY-002`)**: A destruction replacement applies to every death that happens at the same moment as its source's own death. Deaths caused while one stack item resolves, and the death sweep right after it, happen at the same moment. Kills during resolution happen inline (`FR-LAY-002`), and lethal damage waits for that sweep, so the source is still on the battlefield for every one of them, even after a byproduct has left it lethal. *Example:* one kill step kills two allied Scratch Companies while *Shifting Impossibility* has 1 Current Health. Both kills are replaced (returned to hand, 1 damage each to *Shifting Impossibility*), and *Shifting Impossibility* is destroyed at the sweep after the item.
- **Replacement Within a Death Sweep (`SVY-R20`, `SVY-R22`, `FR-REP-001`, `FR-LAY-002`)**: The assets that are lethal when a death sweep starts die simultaneously. Every death among them that a replacement applies to is replaced, one at a time, including its byproduct events, whether or not the replacement's source is among the lethal assets or a byproduct leaves it lethal. Replaced deaths are processed in a fixed order: the active player's assets first, then the non-active player's, and each player's oldest asset first. Then the remaining lethal assets are destroyed together. The sweep then repeats with every asset that is still lethal, which is a new moment. That covers an asset a byproduct newly left lethal, and an asset a replacement's steps left on the battlefield and still lethal. A replacement that has already applied to an asset in this sweep does not apply to it again (`FR-REP-003`, `SVY-R24`), so a different replacement applies or the asset is destroyed. *Example:* *Craterfall Ancient* is killed and deals 4 damage to all assets. *Shifting Impossibility* (Maximum Health 4) and three allied Scratch Companies are lethal at the same sweep. All three deaths are replaced, and *Shifting Impossibility* is then destroyed with the sweep's other deaths. *Example:* *Shifting Impossibility* (Maximum Health 4, undamaged) and five allied Scratch Companies: the five Scratch Companies are lethal at a sweep, and *Shifting Impossibility* is not. All five deaths are replaced, leaving *Shifting Impossibility* at −1 Current Health, and the sweep's next pass destroys it. *Example:* a replacement reading *"If an allied unit would be killed, deal 1 damage to this instead"* applies to an allied unit at −2 Current Health. The unit stays on the battlefield and stays lethal, and the sweep's next pass destroys it. To save a unit from lethal damage, a replacement's steps must also leave it non-lethal, for example *"and remove all damage from it"*. The same replacement does save a unit from a kill effect, because a killed unit with Health left is not lethal afterwards. If one stack item both kills that unit and leaves it at 0 Current Health, the replacement applies to the kill and once more at the sweep after the item, and then the sweep's next pass destroys the unit.
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
- **Eligible Defenders**: An attack may target only an opposing Asset (Unit or Structure) in the **Main Theater** or the opposing **Headquarters** in its **Headquarters Zone**. *Opposing* means controlled by an opponent of the attacker's controller.
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
   - During their Main Phase with an empty Resolution Stack, the active player chooses a Readied eligible attacker, exhausts it upfront, chooses a single legal opposing combat target (`FR-CBT-001`), and places the Attack action onto the bottom of the Resolution Stack.
2. **Step 2 (Pre-Combat Trigger & Reaction Window)**:
   - All abilities triggered by the attacker exhausting or by the attack being declared (such as **Aggressive N** [`FR-KWD-002f`], *Samsara, Living Tempest*, *Arkha, Sun's Fury*, or opposing defensive triggers like *Blind Defile*) enqueue onto the Resolution Stack **on top of** the Attack action in APNAP order (`FR-TRG-003`).
   - Because the Attack action is on the stack, no manual card plays or actions may be declared (`FR-QUE-001`). All triggers above the Attack action resolve top-down until the Attack action itself reaches the top of the stack.
3. **Step 3 (Simultaneous Combat Damage, Symmetric Retaliation & Post-Combat Resolution)**:
   - **Committed Combat Target**: Once the Attack action is on the stack with its target, the attack restrictions of `FR-CBT-001` (Guardian priority and Ethereal) are not checked again. A Guardian entering the Main Theater, or the defender gaining Ethereal, during Step 2 does not redirect or fizzle the attack. Ward never applies to attacks (`FR-KWD-002h`). Only the target check below can stop it.
   - **Combatant Target Check (`SVY-R8`)**: When the Attack action resolves, its target is re-verified like any other target (`FR-TGT-003`). The defender must still be on the battlefield, in a legal combat zone, and an opposing asset: controlled by an opponent of the attacker's current controller, or of its last-known controller if the attacker has left the battlefield. The attacker must still be on the battlefield in a legal combat zone. If either check fails, the Attack **fizzles immediately** and zero combat or retaliation damage is dealt. A control change on either side leaves both combatants under one player, so the defender check fails.
   - **Simultaneous Damage & 100% Symmetric Retaliation**: If both combatants pass the target check, continuous condition gates (including **Cornered**, `FR-LAY-003`) are locked, and both assets deal combat damage to each other **simultaneously** equal to their respective effective **Combat Power**:
      - The attacker deals combat damage equal to its effective Combat Power to the defender.
      - The defender simultaneously deals combat retaliation damage equal to its effective Combat Power back to the attacker. Retaliation is **100% symmetric across all asset types** (Unit vs. Unit, Unit vs. Structure, Unit vs. Headquarters, and Headquarters vs. Headquarters).
   - **Mitigation & State-Based Destruction**: Both combat damage instances route through the 5-stage damage mitigation sequence (`FR-REP-004`; **Stalwart** applies to both strike and retaliation damage). Any combatants reduced to $\le 0$ Current Health are destroyed simultaneously (`FR-ZON-002`).
    - **Post-Attack & Combat-Kill Triggers**: After combat damage and state-based destructions complete, the engine emits `EVENT_COMBAT_KILL` carrying `killer_asset_id` (the asset that dealt the lethal combat damage), `victim_asset_id` (the destroyed asset), and `is_attacker` (true if the killer was the declared attacker, false if defender retaliation). Post-attack abilities (*"After this unit attacks"*, *Piston Fist Prototype*) and combat-kill abilities (*"Whenever [this asset] attacks and kills..."*, *Eternal Watchman*, *Dread Manifestation*) trigger and enqueue onto the Resolution Stack in APNAP order (`FR-TRG-003`). Combat-kill triggers evaluate `killer_asset_id` and `is_attacker` to restrict execution strictly to when that asset was the declared attacker destroying an opposing defender. If a replacement effect keeps the defender from being destroyed, it was not killed, and combat-kill abilities do not trigger (`SVY-R19`, `SVY-R15`). **Overkill** (`FR-KWD-002p`) is not a combat-kill trigger. It triggers from the attacker's Step 3 combat damage when that damage leaves the defender's Current Health below 0, whether or not the defender is then destroyed.

#### FR-CBT-003: Persistent Damage, Headquarters Defeat & Match Draws
- **Persistent Damage**: Damage marked on an Asset does **not** heal or clear at the end of the turn; it remains on the Asset until healed (`FR-KWD-002d`) or until the Asset leaves the battlefield (`FR-ZON-004`).
- **Headquarters Destruction Defeat**: Headquarters defeat is checked in the state-based death sweep (`FR-LAY-002`), the next time no item is resolving on the Resolution Stack. If a player's **Headquarters** has $\le 0$ Current Health (or $\le 0$ Effective Maximum Health) at that sweep, that player loses the match. Damage taken mid-resolution does not end the match by itself: a later clause of the same resolving item that heals the Headquarters back above `0` prevents the loss.
- **Simultaneous Defeat (Match Draw)**: If both players would lose at the same state-based death sweep, by any combination of Headquarters at $\le 0$ Current Health and required draws from an empty deck (`FR-WIN-001`), the match ends in a **Match Draw**.

---

## APPENDIX A: KEYWORD MECHANICS DICTIONARY

- **FR-KWD-002a: Unbowed**: Triggered ability (`"Unbowed: [Effect]."`). Triggers whenever this asset takes actual damage (`(armorAbsorbed + healthDamageCommitted) > 0` in `FR-REP-004`). If incoming damage is reduced to `0` in Stage 2 (**Resistant**), Unbowed does not trigger. If an asset with Unbowed takes lethal damage, its Unbowed ability still triggers, even if the asset is destroyed. Like any trigger, it is placed after the death sweep (`FR-TRG-004`), together with any death triggers from the same event (`FR-TRG-003`). When that Unbowed trigger resolves after the asset has been destroyed:
  - Effect clauses that do **not** mutate the destroyed asset itself (e.g., *Hammerhand Brigands* `Scratch 1`, *Devilhorn Captain* `Deal 1 damage to chosen asset in the Main Theater`, *Chainbreaker Irregulars* `Draw a card, then discard a card`) resolve completely from the discard pile.
  - Effect clauses that attempt to heal or modify the destroyed asset itself (e.g., *Macabre Dancers* `Heal 2`, *Allshan, Chainbreaker Revenant* `This unit gains 2 armor`) fizzle because the asset is no longer on the battlefield (`FR-TGT-003`, `FR-ZON-004`).
- **FR-KWD-002b: Resistant N**: Static damage replacement ability (`FR-REP-004` Stage 2). Reduces each discrete instance of incoming combat or non-combat damage dealt to this asset by $N$ (to a minimum of `0` for positive $N$). If $N$ is negative (`Resistant -1`), it amplifies each incoming damage instance by $|N|$ (even amplifying `0` raw damage to $|N|$).
- **FR-KWD-002c: Stalwart**: Static combat damage replacement ability (`FR-REP-004` Stage 3). Canonical rules meaning: *"All combat damage taken is reduced to 1."* Caps any positive combat damage dealt to this asset from an Attack or combat retaliation (`FR-CBT-002`) after Stage 2 (`Resistant`) to a maximum of `1`. Stalwart has no effect on non-combat damage (from Maneuvers or abilities).
- **FR-KWD-002d: Heal N**: One-shot effect action (`"[Target] heals N."`). Removes up to $N$ accumulated damage from the target asset (cannot reduce accumulated damage below `0` or raise Current Health above Effective Maximum Health).
- **FR-KWD-002e: Cornered**: Conditional state gate (`"Cornered: [Ability]."`). Active for a player whenever their opponent controls strictly more battlefield Assets (Units and Structures combined) in the **Main Theater** than that player controls in the **Main Theater** (`FR-LAY-003`). For triggered abilities, Cornered is checked strictly when the trigger event occurs; for continuous effects during simultaneous combat/damage steps, Cornered is locked at the start of the simultaneous step. **A keyword behind Cornered (`SVY-R18`)**: an asset with *"Cornered: [Keyword]."* (*Allshan, Chainbreaker Revenant*: *"Cornered: Aggressive 3."*) has that keyword only while its controller is Cornered. While Cornered is not active, the asset does not have the keyword, so an effect that refers to the keyword (such as an aura giving *"units with Aggressive"* +2 power) does not apply to it. For a triggered keyword such as **Aggressive N** (`FR-KWD-002f`), Cornered is still checked when the trigger event occurs, as above.
- **FR-KWD-002f: Aggressive N**: Triggered combat ability. Meaning: *"Whenever this unit attacks, it gets +N power until end of turn."* Triggers in Step 2 of combat (`FR-CBT-002`) and resolves before Step 3 combat damage is dealt.
- **FR-KWD-002g: Ethereal**: Static combat restriction. Meaning: *"This asset cannot be attacked."* Opposing attackers cannot declare this asset as a combat target (`FR-CBT-001`). Ethereal overrides **Guardian**.
- **FR-KWD-002h: Ward**: Static targeting restriction. Meaning: *"This asset cannot be chosen by opponents except to attack."* Opponents cannot choose this asset as a target for Maneuvers, activated abilities, or triggered abilities, though it may still be attacked in combat and is still affected by non-targeted opposing effects (such as *"deal 4 damage to all assets"*). Ward is checked when targets are chosen and again at resolution (`FR-TGT-003`): an opposing target that gains Ward before its card or ability resolves becomes illegal.
- **FR-KWD-002i: Versatile**: Inherent resource capability modifier for Units (`FR-CAP-001`). When exhausting to generate resources, a Unit with **Versatile** may generate either its **Requisition Yield** or its **Tactics Yield** (controller's choice), regardless of whether it is in **Reserves** or the **Main Theater**.
- **FR-KWD-002j: Guardian**: Static combat restriction (`FR-CBT-001`). Meaning: *"Opposing attacks must target an asset with Guardian in the Main Theater if able."* While a player controls one or more non-Ethereal assets with **Guardian** in the **Main Theater**, opposing attackers cannot attack that player's Headquarters or non-Guardian Main Theater assets. Guardian has no effect while in **Reserves**.
- **FR-KWD-002k: Armor N**: Persistent damage-absorption pool (`FR-LAY-002`, `FR-REP-004` Stage 4). An asset with printed **Armor N** enters the battlefield with $N$ Armor counters; effects may also grant additional Armor. Armor never expires at end of turn and absorbs incoming damage 1-for-1 before Health is reduced. Damage absorbed by Armor still counts as damage taken for triggers (`FR-KWD-002a`).
- **FR-KWD-002l: Deploys Exhausted**: Static deployment replacement ability (`FR-REP-001`). This asset enters the battlefield **Exhausted** instead of **Readied**, however it is deployed: from Hand, from the Discard pile, for free (`FR-TRN-003a`), or face-down by **Scratch** (`FR-ZON-005`). Whether the asset has Deploys Exhausted is decided as it enters, counting effects already in force that would grant or remove it once it is on the battlefield (`KIM-S06`). An effect that removes Deploys Exhausted from Scratch Companies therefore lets them enter **Readied**, and so does an effect that makes them deploy **Readied**, which applies on top of Deploys Exhausted. It does nothing for an asset that is already on the battlefield (*Forced Mutation*, `FR-ENT-103`).
- **FR-KWD-002m: Scratch N**: One-shot effect action. $N$ times in sequence, the player deploys the top card of their Deck face-down into their **Reserves** as a double-blind `1/1/1/1` **Infantry Scratch Company** unit with **Deploys Exhausted** (`FR-ZON-005`, `FR-KWD-002l`). Scratch is not the **Deploy Scratch Company** inherent action (`FR-CAP-001`): no Headquarters exhausts and nothing is placed on the Resolution Stack. That action performs Scratch 1 when it resolves. Safe on an empty deck (creates no units and does not cause Empty Deck Loss).
- **FR-KWD-002n: Regroup N**: One-shot effect action. The acting player chooses up to $N$ **Exhausted** allied assets (or all matching exhausted allied assets if specified, e.g., *"Regroup all Scratch Company units in your Reserves"*) and transitions them to **Readied**.
- **FR-KWD-002o: Berserk N**: *Retired in v3.14 (`SVY-R12`); superseded by `FR-KWD-002p` (Overkill).*
- **FR-KWD-002p: Overkill**: Triggered combat ability (`"Overkill: [Effect]."`). Triggers when this unit is the declared attacker and its combat damage in Step 3 of `FR-CBT-002` leaves the defender with Current Health below 0 (`FR-LAY-002`). Exactly lethal damage (Current Health 0) is not overkill. The **excess** is how far below 0 the defender's Current Health is after Step 3; it is fixed when the trigger fires. Damage reduced by **Resistant** or capped by **Stalwart** (Stages 2–3) or absorbed by **Armor** (Stage 4) never reaches Health, so it never counts toward the excess, and an asset with Stalwart can never be overkilled (`FR-REP-004`). Overkill still triggers if a destruction replacement (`FR-REP-001`) keeps the defender on the battlefield, and if this unit is destroyed in the same death sweep. Like any trigger it is placed after that sweep (`FR-TRG-004`), in APNAP order with death triggers from the same event (`FR-TRG-003`). Retaliation damage never causes Overkill. Damage dealt by an Overkill effect is non-combat effect damage (`FR-REP-004`), so it cannot cause another Overkill. Two Overkill abilities on one unit trigger separately, each with the same excess (`KIM-O01`). An Overkill effect that resolves after this unit has left the battlefield uses its last-known characteristics (`FR-ZON-004`, `SVY-R10`). A Headquarters defender left lethal ends the match at the death sweep, so its Overkill never resolves.

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
| **FR-CAP-001** | Inherent Asset Capabilities Matrix | Domain 3, Section 3.3 (`FR-CAP-001`; subsumes `FR-INH-001`–`FR-INH-004`) |
| **FR-INH-001** | Inherent Action Upfront Exhaustion & Timing Split | Domain 3, Section 3.3 (Subsumed by `FR-CAP-001`) |
| **FR-INH-002** | Unit Inherent Actions (March, Attack, Generate) | Domain 3, Section 3.3 (Subsumed by `FR-CAP-001`) |
| **FR-INH-003** | Structure Inherent Actions (Generate Resources) | Domain 3, Section 3.3 (Subsumed by `FR-CAP-001`) |
| **FR-INH-004** | Headquarters Inherent Actions (Attack, Scratch, Generate) | Domain 3, Section 3.3 (Subsumed by `FR-CAP-001`) |
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
| **FR-KWD-002a**–**002p** | Canonical Keyword Mechanics Dictionary (`002o` retired) | Appendix A (`FR-KWD-002a` through `FR-KWD-002p`) |
| **FR-KWD-002o** | Berserk N (retired in v3.14, `SVY-R12`) | Appendix A (Subsumed by `FR-KWD-002p`) |
| *Relocated to `02`* | `FR-PRM-001`, `FR-TRG-001`, `FR-TRG-007`, `FR-TGT-101` | [`02_Card_Data_Pipeline_FRD.md`](02_Card_Data_Pipeline_FRD.md) (Modeling Bridge) |
| *Relocated to `05`* | `FR-ENT-001`, `FR-ENT-002`, `FR-ENT-003`, `FR-SNP-101` | [`05_Headless_Game_Engine_FRD.md`](05_Headless_Game_Engine_FRD.md) (Runtime Kernel) |