# Card Data Pipeline & Ability Authoring - Functional Requirements Document (FRD v1.0)
**Master Card Ingestion Pipeline, Declarative Ability AST, MTG-Aligned Text Standard & Schema Specification**

---

## EXECUTIVE SUMMARY & REVISION HISTORY

This document establishes the authoritative technical specification for the card data pipeline, declarative ability abstract syntax tree (AST), rules text templating standards, and dataset validation for the Collectible Card Game (CCG). It serves as the authoritative bridge between balance spreadsheets, compiled engine blueprint definitions, and game client presentation, operating as the third foundational pillar alongside:
1. **Master Game Rules & Card Modeling FRD (`Game_Rules_and_Card_Modeling_FRD.md` - Master FRD v3.0)**
2. **Client-Engine Protocol & Networking FRD (`Client_Engine_Protocol_Networking_FRD.md` - FRD v1.1)**

### Initial Specification Scope (v1.0):
1. **Zero-Aliasing Data Contract (`FR-DAT-006`)**: Codified strict enum target conformance across all authoring CSV balance sheets. Prohibited runtime shorthand aliasing; all raw card sheets must directly serialize canonical target enum identifiers (`UNIT`, `STRUCTURE`, `MANEUVER`, `HEADQUARTERS`; `CONCORDIA`, `SOULLESS`, `PATH`, `ANIMUS`, `NEUTRAL`; `COMMON`, `UNCOMMON`, `RARE`, `MYTHIC`, `HQ`; `RESERVES`, `MAIN_THEATER`, `HEADQUARTERS`).
2. **Deterministic Blueprint ID Derivation (`FR-DAT-002`)**: Formalized a deterministic, immutable blueprint slug derivation algorithm (`<faction>_<snake_case_title>`), decoupling human display titles from engine identity keys while guaranteeing 100% stable serialization.
3. **Comprehensive Ability AST Specification (`FR-ABL-001` - `FR-ABL-007`)**: Fully formalized the declarative schemas previously left abstract in Master FRD v3.0 (`AbilityDef`, `FilterDef`, `ValueToken`, `CostDef`, `TargetRef`, `OpDef`), establishing a unified JSON schema for all six ability archetypes (`TRIGGERED`, `ACTIVATED`, `CONTINUOUS_AURA`, `REPLACEMENT_WATCHER`, `MANEUVER_EFFECT`, `KEYWORD_GRANT`).
4. **Mechanics Extensions & Resolution Semantics (`FR-ABL-008` - `FR-ABL-012`)**: Formalized engine support for per-turn execution gates (`maxPerTurn`), exhaust-as-cost activation mechanics (`costType: EXHAUST_SELF`), modal branch definitions (`CHOOSE_N`), and distributed pool damage allocations (`DISTRIBUTED_POOL`).
5. **MTG-Aligned Rules Text Templating Grammar (`FR-TXT-001` - `FR-TXT-006`)**: Established formal English templating standards modeled on Magic: The Gathering (MTG) syntax, standardizing trigger preambles (`When`, `Whenever`, `At`), cost sequencing (`[Cost]: [Effect].`), passive auras (`have`), temporary modifiers (`gets +N power until end of turn`), and canonical keyword capitalization.
6. **Automated Verification & Linter Architecture (`FR-VAL-001` - `FR-VAL-003`)**: Expanded the automated testing suite (`tools/linter/lint_cardtown.ps1`) to audit requirement cross-references across all three FRDs, enforce zero dangling references, and guarantee 100% parsing integrity.
7. **Comprehensive Card Review Register (Domain 7)**: Audited all 167 cards across the 10 balance sheets, logging all 127 standardized card texts and categorizing 52 explicit design review items for designer sign-off.

---

## DOMAIN 1: CARD DATA ARCHITECTURE, INGESTION LIFECYCLE & AUTHORITY MODEL

```text
┌────────────────────────────────────────────────────────┐
│     Google Spreadsheets (Cloud Viewer & Mobile Mirror) │
│  - maneuvers (animus, concordia, path, soulless)       │
│  - units-and-structures (animus, concordia, etc.)      │
└───────────────────────────▲────────────────────────────┘
                            │
               tools/sheets/publish_card_sheets.ps1
               (Downstream export from repository)
                            │
┌───────────────────────────┴────────────────────────────┐
│     Authoring Balance CSV Sheets (data/cards/*.csv)    │
│  - 10 RFC-4180 UTF-8 CSV Files (Repository Truth)      │
│  - Strict Target Enum Values (FR-DAT-006)              │
│  - MTG-Aligned Standardized Rules Text (FR-TXT-001)    │
└───────────────────────────┬────────────────────────────┘
                            │
              tools/importer/compile_blueprints.ps1
                            │
                            ▼
┌────────────────────────────────────────────────────────┐
│     Compiled Master JSON Blueprints (data/schemas/)    │
│  - blueprints.json (Immutable Singleton Catalog)       │
│  - Full Declarative Ability AST (FR-ABL-001)           │
│  - Strict Schema Validation (FR-VAL-001)               │
└───────────────────────────┬────────────────────────────┘
                            │
          ┌─────────────────┴─────────────────┐
          ▼                                   ▼
┌──────────────────────┐            ┌──────────────────────┐
│ Authoritative Engine │            │ Godot Client (src/)  │
│ - Rules State Machine│            │ - UI Rendering       │
│ - Match Validation   │            │ - Card Text & Frame  │
│ - Fog of War Logic   │            │ - Visual FX Triggers │
└──────────────────────┘            └──────────────────────┘
```

### Section 1.1: Single Source of Truth & Pipeline Flow

#### FR-DAT-001: Data Ingestion Lifecycle & Absolute CSV Source of Truth
- **Tier 1 (Authoritative Repository Source of Truth)**: The 10 per-faction RFC-4180 CSV files residing in `data/cards/` constitute the absolute, definitive source of truth for all card definitions, rules text, and balance statistics. Version control (Git) governs all official card changes.
- **Tier 2 (Cloud Inspection & Mobile Review Mirror)**: The Google Sheets workbooks located in the team drive (`maneuvers` and `units-and-structures`) serve purely as an outward-facing mirror and convenience viewer for designers to inspect card catalogs on mobile devices and explore balance experiments. The Google Sheets are downstream representations; any edits in Google Drive do not affect game builds and must be manually brought into Git CSVs to become authoritative.
- **Tier 3 (Compiled Production Catalog)**: The compiled `blueprints.json` artifact (generated into `data/schemas/` or embedded in the game engine runtime) represents the authoritative machine-readable catalog consumed by the server engine process and Godot client builds.
- **Automated Pipeline Flow**:
  1. Local `data/cards/*.csv` files are directly edited, reviewed, and committed into Git.
  2. `tools/sheets/publish_card_sheets.ps1` publishes the local repository CSV content to Google Sheets. It treats local CSVs as the authoritative input and unconditionally updates Google Sheets to reflect Git state without depending on or checking remote Sheet modifications.
  3. `tools/linter/lint_cardtown.ps1` audits schema integrity, requirement links, and text terminology.
  4. The compilation pipeline (`tools/importer/compile_blueprints.ps1`) parses CSVs, validates values against target enums, binds declarative ability ASTs, and emits the immutable production JSON bundle.

> [!NOTE] Product Vision: Standalone Interactive Card & Set Designer Tool
> Long-term tooling roadmap includes a dedicated, standalone **Interactive Card & Set Designer Tool**. This application will use a structured database model (outside of the client/server runtime engines) to manage card balance schemas, keyword catalogs, set dependencies, and migration tracking, replacing manual spreadsheet authoring while serializing directly into canonical version-controlled formats.

#### FR-DAT-002: Deterministic Blueprint ID & Slug Derivation
Every `CardBlueprint` requires a globally unique, immutable string identifier (`id`) complying with `FR-ENT-001`. The blueprint ID is derived deterministically from the card's printed name and faction via the following algorithmic rules:
1. Strip all non-alphanumeric characters except spaces (apostrophes, commas, dashes, and periods are removed).
2. Convert all uppercase characters to lowercase.
3. Replace consecutive space characters with a single underscore (`_`).
4. Prefix the resulting string with the card's lowercase faction slug followed by an underscore: `<faction>_<slug>`.
5. **Neutral / HQ Exception**:
   - Neutral cards are prefixed with `neutral_` (e.g., `Scratch Company` &rarr; `neutral_scratch_company`). The match-generated token entity is universally identified by the blueprint ID `neutral_scratch_company`.
   - Headquarters cards retain their faction prefix (e.g., `Animus Rift` &rarr; `animus_rift`).
6. **Invariance Rule**: Blueprint IDs are permanently immutable. Renaming a card's display title in design balance sheets does not alter its canonical `id` unless an explicit database migration is executed.

#### FR-DAT-003: Blueprint Compilation & Asset Bundling
- The compilation pipeline evaluates all 10 CSV sheets atomically.
- If any card fails schema validation (`FR-VAL-001`), contains unparsable text syntax, or lacks a required declarative ability AST, compilation aborts immediately with a non-zero exit code. Partial compilation is strictly prohibited.
- Production engine builds bundle the compiled blueprint catalog as a read-only resource singleton instantiated into memory during engine startup (`FR-ENT-001`).

#### FR-DAT-004: Engine Version Binding & Replay Compatibility
- In fulfillment of **FR-SES-001** and **FR-SPC-002**:
  - The compiled blueprint dataset includes a 256-bit cryptographic content checksum (`catalogChecksum`) and a semantic dataset version (`cardDataVersion`, e.g., `"1.0.0"`).
  - During connection negotiation (`FR-SES-001`), the server validates that the client's `clientEngineVersion` and `catalogChecksum` match the server's authoritative version.
  - Historical replay recordings (`.rec` files; `FR-SPC-002`) archive the active `catalogChecksum` within `ReplayHeader` to ensure deterministic playback across balance patches.

---

## DOMAIN 2: BLUEPRINT DATA SCHEMA & CSV NORMALIZATION CONTRACT

### Section 2.1: Target Enum Enforcement

#### FR-DAT-005: Canonical Blueprint Schema Specification
The production blueprint data structure defined in this section serves as the canonical technical schema for the data ingestion and compilation pipeline. It incorporates and extends the fundamental entity schema defined in `FR-CRD-101` (Master Rules FRD) with explicit ingestion metadata and declarative AST fields:

| Field | Type | Constraint | Description |
| :--- | :--- | :--- | :--- |
| `id` | `String` | Pattern: `^[a-z]+_[a-z0-9_]+$` | Deterministic alphanumeric slug (`FR-DAT-002`). |
| `name` | `String` | Non-empty | Title Case card title matching design sheet. |
| `cardType` | `CardTypeEnum` | Valid Enum | `HEADQUARTERS`, `UNIT`, `STRUCTURE`, `MANEUVER`. |
| `faction` | `FactionEnum` | Valid Enum | `CONCORDIA`, `SOULLESS`, `PATH`, `ANIMUS`, `NEUTRAL`. |
| `subtypes` | `Array<String>` | Array of Title Case tags | Normalized subtypes (`FR-DAT-007`). |
| `requisitionCost` | `Integer` | `>= 0` | Base Requisition deployment/casting cost. |
| `tacticsCost` | `Integer` | `>= 0` | Base Tactics deployment/casting cost. |
| `power` | `Integer` or `null` | `>= 0` (null for Maneuvers) | Base Combat Power (`FR-LAY-001`). |
| `health` | `Integer` or `null` | `>= 1` (null for Maneuvers) | Base Maximum Health (`FR-LAY-002`). |
| `requisitionYield` | `Integer` | `>= 0` (0 for Maneuvers) | Resources generated upon exhausting for Requisition. |
| `tacticsYield` | `Integer` | `>= 0` (0 for Maneuvers) | Resources generated upon exhausting for Tactics. |
| `deploymentZones` | `Array<DeploymentZoneEnum>` | Valid Enum Array | `["RESERVES"]`, `["MAIN_THEATER"]`, `["HEADQUARTERS"]`, or both. |
| `rarity` | `RarityEnum` | Valid Enum | `COMMON`, `UNCOMMON`, `RARE`, `MYTHIC`, `HQ`. |
| `rulesText` | `String` | MTG-Aligned Formatting | Standardized printed rules text (`FR-TXT-001`). |
| `abilities` | `Array<AbilityDef>` | Valid AST Array | Declarative definitions of abilities (`FR-ABL-001`). |

#### FR-DAT-006: Direct Schema Header Parity & Zero-Aliasing Enums in CSVs
To eliminate translation layers and ensure absolute schema fidelity, all authoring CSV files in `data/cards/` use column headers that directly mirror canonical Blueprint JSON property names (`name`, `cardType`, `faction`, `subtypes`, `requisitionCost`, `tacticsCost`, `deploymentZones`, `rulesText`, `power`, `health`, `requisitionYield`, `tacticsYield`, `rarity`, `currentProxy`, `needsRework`). Furthermore, enum cells must strictly serialize canonical target enum strings with zero aliases:
1. **`cardType`**: Must equal strictly `UNIT`, `STRUCTURE`, `MANEUVER`, or `HEADQUARTERS`. Values such as `"Unit"`, `"Structure"`, `"Maneuver"`, and `"HQ"` are illegal.
2. **`faction`**: Must equal strictly `CONCORDIA`, `SOULLESS`, `PATH`, `ANIMUS`, or `NEUTRAL`. Title-cased values (e.g. `"Animus"`, `"Concordia"`) are illegal.
3. **`rarity`**: Must equal strictly `COMMON`, `UNCOMMON`, `RARE`, `MYTHIC`, or `HQ`. Single-letter abbreviations (`"C"`, `"U"`, `"R"`, `"M"`) are illegal.
4. **`deploymentZones`**: Must equal strictly `RESERVES`, `MAIN_THEATER`, `HEADQUARTERS`, or comma-separated combinations thereof (e.g., `"MAIN_THEATER, RESERVES"`). Informal labels such as `"Main"`, `"Reserves"`, or `"Headquarters"` are illegal.

#### FR-DAT-007: Subtype & Vanilla Ability Sentinel Normalization
- **Subtype Taxonomy**: Card subtypes in the `subtypes` (or `Types`) CSV column must be serialized as a comma-separated list of Title Case strings (e.g., `"Infantry, Scratch Company"`, `"Specialist, Chainbreaker"`, `"Forge, Construct"`).
- **Null Subtypes Sentinel**: Where a card possesses no subtypes (e.g. Headquarters), the field typically contains the literal sentinel string `"--"`. The ingestion pipeline treats empty strings (`""`), whitespace-only cells, `"N/A"`, and `"--"` identically, normalizing them into an empty array `[]`.
- **Vanilla Ability Sentinel**: In rules text columns (`rulesText` or `Text`), the sentinel string `"--"` is explicitly recognized as representing a vanilla card with no printed abilities (e.g., *Forest Heavies*, *Old Growth Elders*), compiling into an empty `abilities: []` array.
- **Normalization Engine**: The compilation pipeline splits comma-separated strings, strips extraneous whitespace, and normalizes values to an array of unique strings.

#### FR-DAT-008: Deployment Zone Array Schema & Invariants
- **Units & Structures**: Every Unit and Structure blueprint must declare at least one legal deployment zone (`["RESERVES"]`, `["MAIN_THEATER"]`, or `["RESERVES", "MAIN_THEATER"]`). Headquarters cards declare strictly `["HEADQUARTERS"]`.
- **Maneuver Exemption**: Maneuvers resolve on the stack (`FR-QUE-001`, `FR-MNV-001`) and never occupy battlefield zones. Maneuver blueprints serialize `deploymentZones: []` (or `null`).
- **Zone Order Normalization**: In CSV authoring sheets where multi-zone deployment is permitted, zones are normalized alphabetically: `"MAIN_THEATER, RESERVES"`.

---

## DOMAIN 3: DECLARATIVE ABILITY ENGINE SCHEMA & AST SPECIFICATION

```text
┌────────────────────────────────────────────────────────┐
│                      AbilityDef                        │
│                     (Polymorphic)                      │
└───────────────────────────┬────────────────────────────┘
                            │
     ┌──────────────────────┼──────────────────────┐
     ▼                      ▼                      ▼
┌──────────────────┐ ┌──────────────────┐ ┌──────────────────┐
│TriggeredAbilityDef│ │ActivatedAbilityDef│ │ContinuousAuraDef │
│- triggerHook     │ │- costs [CostDef] │ │- layer (Layer 4) │
│- condition       │ │- targeting       │ │- scope (Zone/Sub)│
│- effect          │ │- effect          │ │- modifiers       │
│- maxPerTurn      │ │- repeatability   │ └──────────────────┘
└──────────────────┘ └──────────────────┘
     │                      │                      │
     ▼                      ▼                      ▼
┌──────────────────┐ ┌──────────────────┐ ┌──────────────────┐
│ReplacementWatcher│ │ManeuverEffectDef │ │ KeywordGrantDef  │
│- interceptHook   │ │- targeting       │ │- keywordName     │
│- replacementAction││- effectPayload   │ │- parameter N     │
└──────────────────┘ └──────────────────┘ └──────────────────┘
```

### Section 3.1: Polymorphic Ability AST Schemas

#### FR-ABL-001: Universal AbilityDef Polymorphic Schema
Every entry in a blueprint's `abilities` array conforms to the polymorphic `AbilityDef` specification discriminated by `abilityType`:

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "AbilityDef",
  "type": "object",
  "required": ["abilityType"],
  "properties": {
    "abilityType": {
      "type": "string",
      "enum": [
        "TRIGGERED",
        "ACTIVATED",
        "CONTINUOUS_AURA",
        "REPLACEMENT_WATCHER",
        "MANEUVER_EFFECT",
        "KEYWORD_GRANT"
      ]
    }
  }
}
```

> [!NOTE] Design Rationale (DA-XP-001)
> The declarative, polymorphic Ability AST architecture is intentional and non-negotiable. Digital card games that rely on ad-hoc imperative scripts or stopgap code hooks (such as early MTGO architecture) suffer from compounding technical debt, desync vulnerabilities, and fragile client-server state replication. A fully declarative JSON AST ensures deterministic replay verification, clean client-side visual effect binding, robust schema validation, and long-term extensibility without engine recompilation.

#### FR-ABL-002: TriggeredAbilityDef Schema & Hook Bindings
In compliance with **FR-TRG-001** through **FR-TRG-007**:
- Evaluates automatically whenever the engine emits a matching event hook.
- Enqueues an execution frame onto the Resolution Stack (`FR-QUE-001`) during Sub-Phase 4a or immediately in context.
- Schema properties:
  - `triggerHook`: String enum matching `FR-TRG-001` (e.g., `EVENT_DEPLOY`, `EVENT_DESTROY`, `EVENT_ATTACK_DECLARE`, `EVENT_ATTACK_RESOLVED`, `EVENT_TAKE_DAMAGE`, `EVENT_ZONE_CHANGE`, `EVENT_EXHAUST`).
  - `condition`: Declarative condition object (`ConditionDef`) evaluating source, turn context, zone, or Cornered gate (`FR-KWD-002e`).
  - `targeting`: Optional targeting declaration (`TargetingDef`) if the trigger requires player target selection.
  - `effect`: Execution payload specifying atomic primitive (`FR-PRM-001`) and parameter bindings.
  - `maxPerTurn`: Optional integer (`>= 1`) restricting maximum trigger frequency per turn (`FR-ABL-008`).
- **Upfront Target Binding for Conditional Resolution**: For composite triggers where an effect's magnitude or occurrence depends on an action resolved during execution (*Samsara, Living Tempest*), all external targets (e.g. `chosen asset`) are designated and locked upfront when the trigger is enqueued onto the stack (`FR-TGT-101`). During resolution, if the condition is not met (e.g. no maneuver discarded), the conditional damage clause simply does not execute against the pre-locked target, avoiding secondary interactive prompts.

#### FR-ABL-003: ActivatedAbilityDef Schema & Cost Models
In compliance with **FR-ACT-001**, **FR-ACT-002**, and **FR-QUE-001**:
- Declared manually by the active player during the Main Phase while `Stack.isEmpty() == true`.
- Schema properties:
  - `costs`: Array of `CostDef` structures requiring simultaneous payment upon declaration (`FR-ABL-007`, `FR-ABL-009`).
  - `targeting`: Universal targeting filter and quantifier (`FR-TGT-001`, `FR-TGT-002`).
  - `effect`: Execution payload specifying atomic primitives.
  - `repeatable`: Boolean indicating whether the ability may be activated multiple times per turn (default `true` unless constrained by costs or `maxPerTurn`).

#### FR-ABL-004: ContinuousAuraDef Schema & 6-Layer Interfacing
In compliance with **Section 6.2 of Master FRD v3.0 (`FR-LAY-001` - `FR-LAY-003`)**:
- Evaluates continuously across board assets, registering static modifiers into Layer 4 (Auras) of the 6-layer attribute pipeline.
- Schema properties:
  - `scope`: `SELF`, `ALLIED_IN_ZONE`, `OPPOSING_IN_ZONE`, or `ALL_IN_ZONE`.
  - `modifiers`: Array of `AttributeModifierDef` specifying target attribute (`COMBAT_POWER`, `MAX_HEALTH`, `KEYWORD_GRANT`, `SUBTYPE_GRANT`), mathematical operation (`ADD`, `SUBTRACT`, `SET`), and value expression (`ValueToken`). (Note: Armor cannot be granted continuously by an aura; it exists strictly as a dynamic entity pool modified via one-shot events).
  - `abilityGrants`: Optional `Array<AbilityDef>` allowing continuous auras to project compound abilities (e.g. conditional keywords such as `Cornered: Resistant 1` on *Ancestral Monument*) onto scoped assets.

#### FR-ABL-005: ReplacementWatcherDef Schema & Event Interception
In compliance with **FR-REP-001** through **FR-REP-004**:
- Intercepts prospective state events inline (`WOULD_TAKE_DAMAGE`, `WOULD_BE_DESTROYED`, `WOULD_DEPLOY`) without utilizing the Resolution Stack.
- Schema properties:
  - `interceptHook`: Event hook to intercept.
  - `condition`: Conjunction condition filter.
  - `replacementAction`: Atomic mutation substituting the prospective event payload (e.g., `REDIRECT_DAMAGE_TARGET`, `SET_INITIAL_READINESS`, `CANCEL_AND_BOUNCE`).

#### FR-ABL-006: ManeuverEffectDef Schema & Stack Resolution
In compliance with **FR-MNV-001** and **FR-TRN-003**:
- Represents the core spell payload executed when a Maneuver card resolves from the top of the Resolution Stack.
- Supports composite, sequential, and modal payloads (`FR-ABL-010`).

### Section 3.2: Declarative AST Sub-Schemas

#### FR-ABL-007: Common Declarative AST Types

```json
{
  "definitions": {
    "FilterDef": {
      "type": "object",
      "properties": {
        "allegiance": { "type": "string", "enum": ["ALLIED", "OPPOSING", "ANY"] },
        "entityType": { "type": "string", "enum": ["UNIT", "STRUCTURE", "HEADQUARTERS", "ASSET", "CARD"] },
        "subtypes": { "type": "array", "items": { "type": "string" } },
        "negateSubtypes": { "type": "boolean" },
        "excludeScratch": { "type": "boolean" },
        "excludeHQ": { "type": "boolean" },
        "excludeUnits": { "type": "boolean" },
        "readiness": { "type": "string", "enum": ["READIED", "EXHAUSTED", "ANY"] },
        "zone": { "type": "string", "enum": ["SAME_ZONE", "RESERVES", "MAIN_THEATER", "HEADQUARTERS", "DISCARD", "HAND", "DECK"] }
      }
    },
    "ValueToken": {
      "type": "object",
      "required": ["type"],
      "properties": {
        "type": { "type": "string", "enum": ["STATIC", "DYNAMIC", "VARIABLE_X"] },
        "value": { "type": "integer" },
        "source": { "type": "string", "enum": ["SELF", "TARGET", "GAME_STATE"] },
        "property": { "type": "string", "enum": ["COMBAT_POWER", "CURRENT_HEALTH", "MAX_HEALTH", "COUNT_TUCKED_CARDS", "COUNT_DISCARD_CARDS", "COUNT_STRUCTURES_REQUISITION_YIELD", "COUNT_STRUCTURES_TACTICS_YIELD"] },
        "filter": { "$ref": "#/definitions/FilterDef" },
        "multiplier": { "type": "integer" }
      }
    },
    "CostDef": {
      "type": "object",
      "required": ["costType"],
      "properties": {
        "costType": { "type": "string", "enum": ["RESOURCE", "EXHAUST_SELF", "KILL_ASSET", "DISCARD_CARD", "SACRIFICE_ALLY"] },
        "resource": { "type": "string", "enum": ["REQUISITION", "TACTICS"] },
        "amount": { "type": "integer", "minimum": 1 },
        "filter": { "$ref": "#/definitions/FilterDef" }
      }
    },
    "TargetRef": {
      "type": "object",
      "required": ["selector"],
      "properties": {
        "selector": { "type": "string", "enum": ["SELF", "CHOSEN", "ALL", "DISTRIBUTED_POOL", "RANDOM"] },
        "filter": { "$ref": "#/definitions/FilterDef" },
        "count": { "type": "integer", "minimum": 1 },
        "poolValue": { "type": "integer", "minimum": 1 }
      }
    }
  }
}
```

---

## DOMAIN 4: EXTENDED GAME MECHANICS & EXECUTION SEMANTICS

### Section 4.1: Extended Mechanics Formalization

#### FR-ABL-008: Per-Turn Trigger & Execution Frequency Limits
- To model cards with explicit frequency caps (*Animoceps Rager*: *"This ability triggers only once each turn"*), `TriggeredAbilityDef` and `ActivatedAbilityDef` support an optional `maxPerTurn: Integer` property.
- **Engine Execution**: The engine tracks an instance execution counter (`turnActivationCount`) on each `AssetInstance`. When the counter reaches `maxPerTurn`, the ability gate fails and suppresses trigger enqueuing or activated command validation. All per-turn counters reset to zero during Ready Phase Sub-Phase 1.

#### FR-ABL-009: Exhaust-Self as Ability Cost Model
- In standard card game semantics (and MTG tap costs `{T}`), an asset may require exhausting itself as an explicit activation cost.
- Formalized via `CostDef` with `costType: "EXHAUST_SELF"`:
  - Requires the host asset to have `readiness == READIED`.
  - Transitions the asset's readiness state to `EXHAUSTED` upon declaration, emitting `EVENT_EXHAUST`.
  - May be combined with resource payments (e.g. `[R1, EXHAUST_SELF]` on *Dolus Macrocannon*).

#### FR-ABL-010: Modal Selection AST Primitives
- Cards offering modal options (*Rogue Fleshshaper*, *Ambush Tactics*) specify `abilityType: "MANEUVER_EFFECT"` or `"TRIGGERED"` with:
  - `modalCount`: Integer indicating number of modes to choose (e.g. `1` for *"Choose one"*, `2` for *"Choose two"*).
  - `options`: Array of discrete ability effect payloads.
- Validated via client `selected_modal_indices` (`FR-MSG-003`).

#### FR-ABL-011: Distributed Pool Value Allocations
- In compliance with **FR-TGT-002** and **FR-MSG-003**:
  - Cards distributing a pool of values across multiple targets (*Macrocannon Barrage*: *"Distribute 5 damage among up to 5 chosen units"*, *Hammerhand Bivouac*: *"Distribute 1 armor among allied units in the same zone"*) specify `selector: "DISTRIBUTED_POOL"`.
  - The client provides `TargetBinding` entries where `allocated_value` sums exactly to `poolValue`.

#### FR-ABL-012: Canonical Keyword AST Parameterization
All 15 engine keywords defined in Master FRD v3.0 Appendix A map directly to standardized AST structures:

| Keyword | AST Schema Archetype | Parameterization / Payload |
| :--- | :--- | :--- |
| **Unbowed** | `TRIGGERED` | `triggerHook: EVENT_TAKE_DAMAGE`, condition: `damageTaken > 0`. |
| **Resistant N** | `REPLACEMENT_WATCHER` | `interceptHook: WOULD_TAKE_DAMAGE`, Stage 2 flat reduction `N`. |
| **Stalwart** | `REPLACEMENT_WATCHER` | `interceptHook: WOULD_TAKE_DAMAGE`, Stage 3 combat cap `min(dmg, 1)`. |
| **Heal N** | `ATOMIC_PRIMITIVE` | Primitive `HEAL`, amount `N`. |
| **Cornered** | `CONDITIONAL_GATE` | Gate: `count(opposingAssets) > count(alliedAssets)`. |
| **Aggressive N** | `TRIGGERED` | `triggerHook: EVENT_ATTACK_DECLARE`, effect `MODIFY_STAT(+N Power)`. |
| **Ethereal** | `STATIC_RESTRICTION` | Ineligible target for attack declarations (`FR-CBT-001`). |
| **Ward** | `STATIC_RESTRICTION` | Ineligible target for non-combat enemy selections (`FR-KWD-002h`). |
| **Versatile** | `CAPABILITY_FLAG` | Unit inherent generator yields either Requisition or Tactics (`FR-CAP-001`). |
| **Guardian** | `STATIC_RESTRICTION` | Mandatory combat attack target priority when able (`FR-CBT-001`). |
| **Armor N** | `STATIC_ATTRIBUTE` | Dynamic Armor pool initialized to `N` upon deployment (`FR-REP-004`). |
| **Deploys Exhausted**| `REPLACEMENT_WATCHER` | `interceptHook: WOULD_DEPLOY`, sets initial readiness to `EXHAUSTED`. |
| **Scratch N** | `ATOMIC_PRIMITIVE` | Executes `Deploy Scratch Company` `N` times sequentially (`FR-INH-004`). |
| **Regroup N** | `ATOMIC_PRIMITIVE` | Transitions up to `N` exhausted assets to Readied (`FR-KWD-002n`). |
| **Berserk** | `STATIC_RESTRICTION` | Enables attack declaration targeting up to 2 assets (`FR-KWD-002o`). |

- **Dynamic & Continuous Grants**: Keywords and subtypes may also be granted dynamically via atomic primitives `GRANT_KEYWORD` and `GRANT_SUBTYPE` (`FR-PRM-001`) or continuously via `ContinuousAuraDef` modifiers (`KEYWORD_GRANT`, `SUBTYPE_GRANT`) and compound `abilityGrants` (Layer 6; `FR-ABL-004`).

---

## DOMAIN 5: MTG-ALIGNED RULES TEXT TEMPLATING STANDARD & GRAMMAR

### Section 5.1: Card Text Syntax Guidelines

#### FR-TXT-001: MTG-Aligned Sentence Grammar & Capitalization
1. Every rules text clause must form a grammatically complete sentence beginning with an uppercase letter and terminating with a period (`.`), unless formatted as a keyword on a standalone line.
2. Proper nouns and system entities are capitalized: `Headquarters`, `Reserves`, `Main Theater`, `Scratch Company`, `Requisition`, `Tactics`, `Health`, `Combat Power`.
3. Informational reminder text (e.g. `"(cannot be chosen by opponents except to attack)"`) must be stripped from authoritative rules text. Reminders are rendered via client tooltip hover systems.

#### FR-TXT-002: Trigger Preambles (`When`, `Whenever`, `At`)
To ensure total clarity regarding trigger frequency:
1. **`When`**: Used exclusively for events that occur once per object lifecycle:
   - *"When this unit deploys, ..."*
   - *"When this unit is killed, ..."*
2. **`Whenever`**: Used for recurring game events that can occur multiple times:
   - *"Whenever this unit attacks, ..."*
   - *"Whenever this unit exhausts, ..."*
   - *"Whenever an opposing unit attacks, ..."*
   - *"Whenever this Headquarters is dealt damage, ..."*
3. **`At`**: Used for phase and turn step boundaries:
   - *"At the start of your turn, ..."*
   - *"At the end of your turn, ..."*

#### FR-TXT-003: Cost Notation Syntax
1. **Activated Abilities**: Formatted strictly as `[Cost], [Cost]: [Effect].`
   - Resource costs precede exhaust and sacrifice costs: `R1, exhaust: Deal 2 damage to chosen non-Scratch asset.`
   - Non-resource sacrifices: `T1, kill this unit: Return chosen Fungus or Plant card from your discard pile to your hand.`
2. **Additional Enactment Costs**: Formatted as a leading preamble sentence on its own line:
   - *"As an additional cost to enact this maneuver, kill an allied unit."*

#### FR-TXT-004: Conjunction Targeting Phrasing & Two-Stage Verification
Card Town's targeting engine (`FR-TGT-001`) utilizes declarative conjunctions. Rules text must use the canonical targeting verb **"chosen"** rather than MTG's "target":
- *"chosen allied unit"*
- *"chosen opposing non-Scratch unit"*
- *"chosen asset in the Main Theater"*
- Multi-target distributions: *"each of up to 2 chosen opposing assets"*, *"distribute 5 damage among up to 5 chosen units"*.
- **Two-Stage Dynamic Filter Invalidation**: In compliance with `FR-TGT-103`, abilities with dynamic target filters (such as *Consuming Earth*: *"Kill chosen opposing unit with power less than or equal to the total Requisition yield of all structures you control."*) validate target legality both when placed on the stack and immediately prior to effect resolution. If the target no longer satisfies the dynamic filter criteria at resolution (e.g., allied structures were destroyed or the target's power increased), the effect fizzles against that target without executing.

#### FR-TXT-005: Stat & Defense Modification Phrasing
1. **Temporary Modifiers (Layer 5)**: Use the verb **"gets"** for singular subjects and **"get"** for plural subjects:
   - Singular: *"Chosen allied unit gets +1 power until end of turn."*
   - Plural: *"Up to 2 chosen units get -1 power until end of turn."*
   - Debuffs: *"Chosen opposing unit gets -3 power until the end of your opponent's turn."*
2. **Continuous Auras (Layer 4)**: Use the verb **"has"** for singular subjects and **"have"** for plural subjects:
   - Singular (Self / Individual): *"This unit has +2 power."*, *"Chosen asset has Stalwart."*
   - Plural (Collective / Scoped): *"Other allied assets in the same zone have +1 power."*, *"Allied Horror units in the same zone have Resistant 1."*
3. **Keyword Grants**: Use the verb **"gains"** for singular subjects and **"gain"** for plural subjects:
   - *"Chosen unit gains Stalwart until end of turn."*
   - *"This unit gains 1 armor."*
   - *"Allied units in the same zone gain 1 armor."*

#### FR-TXT-006: Zone Naming, Card Reference & Cardinality Conventions
1. **Zone Names**: Always capitalized and complete: `Main Theater`, `Reserves`, `Headquarters Zone`, `discard pile`, `deck`.
2. **Cards vs Assets**: Entities residing in non-battlefield zones are referred to as **cards**:
   - *"return chosen unit card from your discard pile to your hand."*
   - Entities on the battlefield are referred to as **units**, **structures**, or **assets**.
3. **Numerical Expressions**:
   - Cards drawn or discarded use words: *"Draw a card"*, *"Draw 2 cards"*, *"discard a card"*.
   - Damage, armor, and keyword ranks use Arabic numerals: *"Deal 2 damage"*, *"gains 3 armor"*, *"Scratch 1"*, *"Resistant 1"*.

---

## DOMAIN 6: AUTOMATED VALIDATION, INGESTION LINTER & CI/CD PIPELINE

### Section 6.1: Verification Suite Extensions

#### FR-VAL-001: Ingestion Pipeline Linter Specification
The automated linter (`tools/linter/lint_cardtown.ps1`) enforces three validation phases:
1. **Requirement Integrity**: Scans all FRDs in `docs/frd/` (`Game_Rules_and_Card_Modeling_FRD.md`, `Client_Engine_Protocol_Networking_FRD.md`, `Card_Data_Pipeline_FRD.md`), verifying that every referenced requirement ID is defined (zero dangling references).
2. **CSV Table & Enum Integrity**: Verifies that all 10 CSVs parse without errors, column counts are invariant, no duplicate headers exist, and every enum cell matches target enum values (`FR-DAT-006`).
3. **Rules Text Terminology**: Audits rules text against forbidden terminology patterns (`friendly` &rarr; `allied`, raw `~`, raw `attack` &rarr; `power`).

#### FR-VAL-002: Rules Text vs Declarative Ability Parity Checker
- The compilation pipeline executes automated parity checks ensuring that card abilities defined in text correspond 1-to-1 with compiled AST structures.
- A card declaring `Stalwart` in rules text must contain a `REPLACEMENT_WATCHER` AST entry with `min(damage, 1)` logic; a card declaring an attack trigger must bind to `EVENT_ATTACK_DECLARE` or `EVENT_ATTACK_RESOLVED`.

#### FR-VAL-003: Uniform Dataset Validation Policy ("Must Fit All") & Review Card Compilability
- **Universal Compilability Invariant**: Balance sheet validation requirements apply uniformly across all rows with zero exceptions for placeholder rows ("must fit all"). Every row in an authoring sheet must fully define all mandatory target enum fields (`Card Type`, `Faction`, `Deployment Zones`, `Rarity`), pass CSV schema validation, and possess complete, syntactically valid declarative ability bindings (`FR-ABL-001` - `FR-ABL-007`). Incomplete rows lacking required attributes or cards with incomplete/unbound ability logic are rejected during pipeline linting and compiler execution.
- **Review Card Compilability Contract**: Cards undergoing active game balance or rules text review (e.g., the 52 cards listed in Domain 7) must still compile cleanly with complete, functional ability bindings in developmental and pull-request builds. Work-in-progress or exploratory cards are not permitted to bypass compilation or remain in an unbindable state.
- **Production Gating**: While review cards must remain 100% compilable for test matches and CI validation, production release builds require formal designer sign-off clearing all pending review flags from Domain 7.

---

## DOMAIN 7: CARD AUDIT REGISTER & DESIGNER REVIEW LOG

### Section 7.1: Audit Summary & Critical Design Decisions

All 167 cards across the 10 CSV sheets were comprehensively audited and standardized to conform to target enums and MTG-style grammar. 127 cards had text standardized for punctuation, capitalization, and canonical syntax, and 52 cards were flagged for design review.

#### High-Priority Designer Review Items:
1. **Animoceps Rager (Headquarters - Animus)**:
   - *Previous*: `"When this HQ is dealt damage, scratch once. This effect happens once per turn."`
   - *Standardized*: `"Whenever this Headquarters is dealt damage, Scratch 1. This ability triggers only once each turn."`
   - *Design Call*: Formalized the per-turn trigger frequency limit (`maxPerTurn: 1`; `FR-ABL-008`).
2. **Funeral Fungus (Maneuver - Animus)**:
   - *Previous*: `"exhaust chosen ready allied unit and return 1 unit card from your discard pile to your hand."`
   - *Standardized*: `"As an additional cost to enact this maneuver, exhaust a readied allied unit. Return chosen unit card from your discard pile to your hand."`
   - *Design Call*: Templated exhausting the unit as an additional casting cost rather than an effect to eliminate spell fizzle ambiguity if the unit is exhausted prior to resolution.
3. **Rift (Headquarters - Animus)**:
   - *Previous*: `"discard a card, exhaust: scratch 2, deal 5 damage to this asset"`
   - *Standardized*: `"Discard a card, exhaust: Scratch 2. Deal 5 damage to this Headquarters."`
   - *Design Call*: Confirmed self-damage interaction on Headquarters; Headquarters is eligible target for damage mitigation (`FR-REP-004`).
4. **Mutation Swell (Maneuver - Animus)**:
   - *Previous*: `"kill up to 4 allied scratch company units. for every 2 units killed, deploy 1 unit from your discard pile for free."`
   - *Standardized*: `"Kill 2 or 4 allied Scratch Company units. For every 2 units killed this way, deploy a unit card from your discard pile for free."`
   - *Design Call*: Clarified MTG causality clause ("killed this way") and restricted targets to discrete multiples (2 or 4) to eliminate fractional casualty floor ambiguity.
5. **Steel Rain & Macrocannon Barrage (Path / Soulless)**:
   - *Standardized*: Formalized distributed damage allocation syntax matching `DISTRIBUTED_POOL` (`FR-ABL-011`).
6. **Title Spellings**:
   - `reverbating terror` corrected to **Reverberating Terror**.
   - `Devilhorn Guerillas` corrected to **Devilhorn Guerrillas**.
   - `All out War` hyphenated to **All-Out War**.

---

### Section 7.2: Master Review Register

The following register details all cards where rules text was modified or requires designer review:

| Card Title | Sheet | Previous Text | Standardized Rules Text | Designer Review Flag / Rationale |
| :--- | :--- | :--- | :--- | :--- |
| **Rift** | `units-and-structures-HQ.csv` | `discard a card, exhaust: scratch 2, deal 5 damage to this asset` | `Discard a card, exhaust: Scratch 2. Deal 5 damage to this Headquarters.` | Review Headquarters self-damage keyword interaction (Unbowed). |
| **Animoceps Rager** | `units-and-structures-HQ.csv` | `When this HQ is dealt damage, scratch once. This effect happens once per turn.` | `Whenever this Headquarters is dealt damage, Scratch 1. This ability triggers only once each turn.` | Confirm "triggers only once each turn" limit syntax (`FR-ABL-008`). |
| **Encampment** | `units-and-structures-HQ.csv` | `discard a card, exhaust: deal 1 damage to up to 2 chosen non scratch company units` | `Discard a card, exhaust: Deal 1 damage to each of up to 2 chosen non-Scratch Company units.` | Review multi-target distributed damage phrasing. |
| **Ensnaring Mist** | `maneuvers-animus.csv` | `move chosen unit to its controller's reserves. chosen unit gets -1 power until the end of your turn.` | `Move chosen unit to its controller's Reserves. Chosen unit gets -1 power until end of turn.` | Changed "until the end of your turn" to canonical "until end of turn". |
| **Mutation Swell** | `maneuvers-animus.csv` | `kill up to 4 allied scratch company units. for every 2 units killed, deploy 1 unit from your discard pile for free.` | `Kill 2 or 4 allied Scratch Company units. For every 2 units killed this way, deploy a unit card from your discard pile for free.` | Standardized to discrete target multiples (2 or 4) to eliminate fractional casualty floor ambiguity. |
| **Forced Mutation** | `maneuvers-animus.csv` | `chosen opposing unit is turned face down and becomes a scratch company unit` | `Turn chosen opposing unit face down. It becomes a Scratch Company unit.` | Review morph rules text alignment with `FR-ENT-103`. |
| **Funeral Fungus** | `maneuvers-animus.csv` | `exhaust chosen ready allied unit and return 1 unit card from your discard pile to your hand.` | `As an additional cost to enact this maneuver, exhaust a readied allied unit. Return chosen unit card from your discard pile to your hand.` | Templated exhaust as additional enactment cost rather than spell effect. |
| **Reverberating Terror** | `maneuvers-concordia.csv` | `up to 2 chosen units and structures get -1 power until end of turn.` | `Up to 2 chosen units and structures get -1 power until end of turn.` | Title spelling corrected ("Reverberating Terror"). |
| **Rage's Wake** | `maneuvers-concordia.csv` | `chosen unit gains +2 power until end of turn. When chosen unit is killed this turn, redeploy it to your reserves from your discard for free.` | `Chosen unit gets +2 power until end of turn. When that unit is killed this turn, deploy it from your discard pile to your Reserves for free.` | Clarified "redeploy" as "deploy it from your discard pile to your Reserves for free". |
| **Consuming Earth** | `maneuvers-concordia.csv` | `kill chosen opposing unit with power less than or equal to the total requisition value of all structures you control.` | `Kill chosen opposing unit with power less than or equal to the total Requisition yield of all structures you control.` | Term "requisition value" replaced with canonical "Requisition yield" (`FR-CRD-101`). |
| **Light of the World Soul** | `maneuvers-concordia.csv` | `look at the top 7 cards of your deck. you may deploy up to 2 unit cards for free, and put up to 1 card into your hand. Put the rest on the bottom of your deck in any order.` | `Look at the top 7 cards of your deck. You may deploy up to 2 unit cards from among them for free, and put up to 1 of those cards into your hand. Put the rest on the bottom of your deck in any order.` | Clarified card partitioning from top-7 inspection pool (`FR-MOD-001`). |
| **Devilhorn Scouting Party** | `maneuvers-path.csv` | `You may deploy a unit card that costs R3 or less from among the top 3 cards of your deck for free. Put the rest on the bottom of your deck in any order. Then, Scratch 1.` | `Look at the top 3 cards of your deck. You may deploy a unit card with Requisition cost 3 or less from among them for free. Put the rest on the bottom of your deck in any order. Then, Scratch 1.` | Added explicit "Look at the top 3 cards..." preamble. |
| **Steel Rain** | `maneuvers-path.csv` | `discard up to 3 cards. For each card discarded this way, deal 2 damage to chosen asset. then, if you have no cards left in your hand, draw 2 cards.` | `Discard up to 3 cards. For each card discarded this way, deal 2 damage to chosen asset. Then, if you have no cards in hand, draw 2 cards.` | Confirm if "chosen asset" is selected once or independently per card discarded. |
| **Ambush Tactics** | `maneuvers-path.csv` | `choose two: kill chosen unit. return chosen unit to its owner's hand. chosen unit deals damage equal to its power to chosen unit. put chosen unit on top of its owner's deck.` | `Choose two: - Kill chosen unit. - Return chosen unit to its owner's hand. - Chosen unit deals damage equal to its power to another chosen unit. - Put chosen unit on top of its owner's deck.` | Modal choice list formatting applied (`FR-ABL-010`); disambiguated target strike role with "another chosen unit". |
| **Mage Seeker's Inquisition** | `maneuvers-soulless.csv` | `Chosen player reveals their hand and discards a unit of your choice.` | `Chosen opponent reveals their hand. You choose a unit card from it. That player discards that card.` | Harmonized with hand inspection workflow `FR-MOD-001`. |
| **Macrocannon Barrage** | `maneuvers-soulless.csv` | `deal 5 damage distributed among up to 5 chosen units` | `Distribute 5 damage among up to 5 chosen units.` | Templated as distributed damage pool (`FR-ABL-011`). |
| **All-Out War** | `maneuvers-soulless.csv` | `Discard your hand, then: - chosen allied asset gains power... - scratch a number of times... - regroup all...` | `Discard your hand. Then: - Chosen allied asset gets power equal to the number of cards discarded this way until end of turn. - Scratch X, where X is the number of cards discarded this way. - Regroup all Scratch Company units in your Reserves.` | Title hyphenated ("All-Out War"); list templating standardized. |
| **Shifting Impossibility** | `units-and-structures-animus.csv` | `When an allied Scratch Company unit would be killed, return it to its owner's hand and deal 1 damage to this unit.` | `If an allied Scratch Company unit would be killed, return it to its owner's hand and deal 1 damage to this unit instead.` | Templated as canonical replacement watcher with "instead" (`FR-REP-001`). |
| **Spoil Forest Haunt** | `units-and-structures-animus.csv` | `this unit deploys exhausted` | `Deploys Exhausted` | Replaced sentence with standard keyword "Deploys Exhausted" (`FR-KWD-002l`). |
| **Rhapsody's Hollow** | `units-and-structures-animus.csv` | `T3, exhaust: kill this structure to gain control of chosen opposing unit in the main theater` | `T3, exhaust, kill this structure: Gain control of chosen opposing unit in the Main Theater.` | Templated structure sacrifice as an activated ability cost. |
| **Flickering Memory** | `units-and-structures-animus.csv` | `ethereal (can't be attacked)` | `Ethereal` | Reminder text stripped. |
| **Sinuous Growth** | `units-and-structures-animus.csv` | `T1, kill this unit: return chosen fungus or plant card from your discard to your hand` | `T1, kill this unit: Return chosen Fungus or Plant card from your discard pile to your hand.` | Templated unit sacrifice as cost. |
| **Cliff of Portents** | `units-and-structures-animus.csv` | `T1, R1, exhaust: search your deck for a card and put it into your hand.` | `R1, T1, exhaust: Search your deck for a card, put it into your hand, then shuffle your deck.` | Added explicit shuffle clause required by `FR-DEK-001`. |
| **Drifting Concept** | `units-and-structures-animus.csv` | `ethereal when this unit deploys, chosen opponent reveals their hand. choose a card from it. that player discards that card.` | `Ethereal When this unit deploys, chosen opponent reveals their hand. You choose a card from it. That player discards that card.` | Harmonized with hand inspection workflow `FR-MOD-001`. |
| **Spoil Forest Amalgam** | `units-and-structures-animus.csv` | `stalwart (all combat damage taken is reduced to 1)` | `Stalwart` | Reminder text stripped. |
| **Despoiler Worm** | `units-and-structures-animus.csv` | `Resistant 2 (all damage taken reduced by 2) when Despoiler Worm is killed, all opponents draw 2 cards` | `Resistant 2 When this unit is killed, all opponents draw 2 cards.` | Self-reference replaced with "this unit"; reminder text stripped. |
| **Craterfall Ancient** | `units-and-structures-animus.csv` | `when this unit attacks, you may deal damage equal to its power to up to 2 other chosen assets. when this unit is killed, deal 4 damage to all assets.` | `Whenever this unit attacks, you may deal damage equal to its power to each of up to 2 other chosen assets. When this unit is killed, deal 4 damage to all assets.` | Clarified attack damage allocation phrasing. |
| **Grove Tender** | `units-and-structures-concordia.csv` | `this unit deploys exhausted` | `Deploys Exhausted` | Replaced sentence with standard keyword "Deploys Exhausted" (`FR-KWD-002l`). |
| **Stormbolt Ray** | `units-and-structures-concordia.csv` | `when this structure deploys, it deals 3 damage to chosen asset. T2, exhaust: return this structure to its owner's hand.` | `When this structure deploys, deal 3 damage to chosen asset. T2, exhaust: Return this structure to its owner's hand.` | Separated deployment trigger from activated ability onto two lines. |
| **Furnace Golem** | `units-and-structures-concordia.csv` | `aggressive 2 (gets +2 power until end of turn when it attacks)` | `Aggressive 2` | Reminder text stripped. |
| **Samsara, Living Tempest** | `units-and-structures-concordia.csv` | `when this unit attacks, look at the top 2 cards of your deck. Put one into your hand and the other into your discard. If the discarded card was a maneuver, deal 2 damage to chosen asset.` | `Whenever this unit attacks, look at the top 2 cards of your deck. Put one into your hand and the other into your discard pile. If the card put into your discard pile this way was a maneuver card, deal 2 damage to chosen asset.` | Clarified causality of discarded maneuver condition. |
| **World Soul Messenger** | `units-and-structures-concordia.csv` | `when this unit deploys, you may search your deck for an Elemental and deploy it for free. If you do, shuffle your deck.` | `When this unit deploys, you may search your deck for an Elemental card and deploy it for free, then shuffle your deck.` | Added explicit shuffle clause required by `FR-DEK-001`. |
| **Hammerhand Brigands** | `units-and-structures-path.csv` | `unbowed: scratch once` | `Unbowed: Scratch 1.` | Standardized to "Unbowed: Scratch 1.". |
| **Eerie Remnant** | `units-and-structures-path.csv` | `This unit deploys exhausted. When this unit is put into your discard from your hand, you may deploy it for free instead.` | `Deploys Exhausted When this card is put into your discard pile from your hand, you may deploy it from your discard pile for free.` | Simplified deploy trigger removing "instead" ambiguity; deploys directly from discard pile. |
| **Brigand Highlander** | `units-and-structures-path.csv` | `cornered: +2 power` | `Cornered: This unit has +2 power.` | Standardized to "Cornered: This unit has +2 power.". |
| **Blind Defile** | `units-and-structures-path.csv` | `when an opposing unit attacks, you may kill this structure and the attacking unit` | `Whenever an opposing unit attacks, you may kill this structure and that attacking unit.` | Refined reference to "that attacking unit" (`FR-CBT-002`). |
| **Rogue Fleshshaper** | `units-and-structures-path.csv` | `when this unit deploys, choose one: chosen asset heals 1 deal 1 damage to chosen non-scratch asset` | `When this unit deploys, choose one: - Chosen asset heals 1. - Deal 1 damage to chosen non-Scratch asset.` | Modal choice list formatting applied (`FR-ABL-010`). |
| **Macabre Dancers** | `units-and-structures-path.csv` | `unbowed: heal 2 cornered: resistant 1` | `Unbowed: Heal 2. Cornered: Resistant 1.` | Standardized to "Unbowed: Heal 2." and "Cornered: Resistant 1.". |
| **Devilhorn Captain** | `units-and-structures-path.csv` | `unbowed: deal 1 damage to chosen asset in the main theater.` | `Unbowed: Deal 1 damage to chosen asset in the Main Theater.` | Standardized to "Unbowed: Deal 1 damage...". |
| **Chainbreaker Irregulars** | `units-and-structures-path.csv` | `unbowed: draw a card, then discard a card.` | `Unbowed: Draw a card, then discard a card.` | Standardized to "Unbowed: Draw a card...". |
| **Boneyard Scavenger** | `units-and-structures-path.csv` | `1T, exhaust: kill this unit and return a maneuver from your discard pile to your hand.` | `T1, exhaust, kill this unit: Return chosen maneuver card from your discard pile to your hand.` | Standardized cost ordering and "1T" -> "T1". |
| **Ancestral Monument** | `units-and-structures-path.csv` | `other allied assets in the same zone as this structure have cornered: resistant 1` | `Other allied assets in the same zone have Cornered: Resistant 1.` | Templated ability grant syntax. |
| **Devilhorn Guerrillas** | `units-and-structures-path.csv` | `when this unit deploys, if it was deployed to the main theater, deal 2 damage to chosen opposing unit.` | `When this unit deploys, if it was deployed to the Main Theater, deal 2 damage to chosen opposing unit.` | Title spelling corrected ("Devilhorn Guerrillas"). |
| **Hammerhand Bivouac** | `units-and-structures-path.csv` | `distribute 1 armor among allied units in the same zone` | `Distribute 2 armor among allied units in the same zone.` | Rebalanced armor pool distribution from 1 to 2 (`FR-ABL-011`). |
| **Ruin Partisans** | `units-and-structures-path.csv` | `T2, exhaust: kill this unit and return a unit card from your discard pile to your hand.` | `T2, exhaust, kill this unit: Return chosen unit card from your discard pile to your hand.` | Standardized cost ordering. |
| **Allshan, Chainbreaker Revenant** | `units-and-structures-path.csv` | `unbowed: this unit gains 2 armor cornered: when this unit attacks, it gets +3 power until end of turn.` | `Unbowed: This unit gains 2 armor. Cornered: Whenever this unit attacks, it gets +3 power until end of turn.` | Standardized ability words. |
| **Solarheart Monk** | `units-and-structures-path.csv` | `when this unit is put into your discard, chosen asset heals 2.` | `When this card is put into your discard pile, chosen asset heals 2.` | Clarified trigger on "this card" rather than "this unit" while entering discard (`FR-ZON-003`). |
| **Manufactus Conscripts** | `units-and-structures-soulless.csv` | `this unit deploys exhausted` | `Deploys Exhausted` | Replaced sentence with standard keyword "Deploys Exhausted" (`FR-KWD-002l`). |
| **Mage Hunter Brigade** | `units-and-structures-soulless.csv` | `Ward (cannot be chosen by opponents except to attack)` | `Ward` | Reminder text stripped. |
| **NO-AH-5** | `units-and-structures-soulless.csv` | `ward R2: kill chosen allied non HQ asset and put it face down under this unit. This unit has +2 power for each card underneath it.` | `Ward R2: Kill chosen allied non-Headquarters asset and tuck it face down under this unit. This unit has +2 power for each card tucked under it.` | Harmonized with canonical benchmark in Master FRD Section 8.2 (`FR-ENT-102`). |
| **Mobile Armory** | `units-and-structures-soulless.csv` | `Exhaust: chosen allied unit in the same zone gets +1 power this turn` | `Exhaust: Chosen allied unit in the same zone gets +1 power until end of turn.` | Changed "this turn" to "until end of turn". |
| **Piston Fist Prototype** | `units-and-structures-soulless.csv` | `After this unit attacks, move this unit to your Reserves.` | `Aggressive 9 After this unit attacks, move this unit to your Reserves.` | Rebalanced to 3R/3T 1/9 with Aggressive 9 and post-attack recoil to Reserves (`FR-KWD-002f`). |
| **MAX-8 Autonomous Warsuit** | `units-and-structures-soulless.csv` | `When this unit deploys, deal 2 damage to chosen asset and chosen asset heals 2.` | `When this unit deploys, deal 2 damage to chosen asset. Chosen asset heals 2.` | Separated damage and heal into two discrete sequential effect clauses. |
| **Gemstone Titan** | `units-and-structures-concordia.csv` | `when this unit deploys, kill chosen structure with current Health less than this unit's power` | `When this unit deploys, kill chosen opposing structure with current Health less than this unit's power.` | Added opposing qualification; filters structures by current Health less than deployer's power. |
| **Daggerspine Rangers** | `units-and-structures-concordia.csv` | Subtypes: `Mending, Specialists` | Subtypes: `Mending, Specialist` | Normalized plural subtype "Specialists" to canonical singular "Specialist" (`FR-DAT-007`). **Flagged for card rename review** (prior art / IP collision with Warcraft "Daggerspine"). |

---

## APPENDIX: COMPLETE CROSS-SPEC TRACEABILITY MATRIX

| Requirement ID | Canonical Requirement Scope | Companion FRD Cross-Reference & Traceability |
| :--- | :--- | :--- |
| **FR-DAT-001** | Data Ingestion Lifecycle & Authority Model | `FR-SET-001`, `FR-CRD-101`, `FR-NET-001` |
| **FR-DAT-002** | Deterministic Blueprint ID & Slug Derivation | `FR-ENT-001`, `FR-SET-001`, `FR-CRD-101` |
| **FR-DAT-003** | Blueprint Compilation & Asset Bundling | `FR-ENT-001`, `FR-CRD-101`, Section 8.1 |
| **FR-DAT-004** | Engine Version Binding & Replay Hash Parity | `FR-SES-001`, `FR-SPC-002`, `FR-PAC-004` |
| **FR-DAT-005** | Production Blueprint Schema Specification | `FR-CRD-101`, `FR-ENT-001`, Section 8.1 |
| **FR-DAT-006** | Zero-Aliasing Target Enum Enforcement in CSVs| `FR-SET-001`, `FR-CRD-101`, `FR-SES-002` |
| **FR-DAT-007** | Subtype Taxonomy & Normalization | `FR-CRD-101`, `FR-ZON-003`, `FR-LAY-001` |
| **FR-DAT-008** | Deployment Zone Array Schema & Invariants | `FR-ZON-001`, `FR-CRD-101`, `FR-TRN-003` |
| **FR-ABL-001** | Universal AbilityDef Polymorphic Schema | `FR-CRD-101`, `FR-PRM-001`, Section 8.1 |
| **FR-ABL-002** | TriggeredAbilityDef Schema & Hook Bindings | `FR-TRG-001`, `FR-TRG-002`, `FR-TRG-007` |
| **FR-ABL-003** | ActivatedAbilityDef Schema & Multi-Modal Costs| `FR-ACT-001`, `FR-ACT-002`, `FR-QUE-001` |
| **FR-ABL-004** | ContinuousAuraDef Schema & 6-Layer Mapping | `FR-LAY-001`, `FR-LAY-002`, `FR-LAY-003` |
| **FR-ABL-005** | ReplacementWatcherDef Schema & Event Interception | `FR-REP-001`, `FR-REP-002`, `FR-REP-004` |
| **FR-ABL-006** | ManeuverEffectDef Schema & Stack Resolution | `FR-MNV-001`, `FR-QUE-001`, `FR-TRN-003` |
| **FR-ABL-007** | Declarative AST Sub-Schemas (Filter, Value, Cost)| `FR-TGT-001`, `FR-PRM-001`, `FR-TGT-101` |
| **FR-ABL-008** | Per-Turn Execution Frequency Limits (`maxPerTurn`)| `FR-ACT-002`, `FR-TRG-001`, `FR-TRN-001` |
| **FR-ABL-009** | Exhaust-Self as Ability Cost Model | `FR-INH-001`, `FR-ACT-001`, `FR-CAP-001` |
| **FR-ABL-010** | Modal Selection AST Primitives | `FR-MOD-001`, `FR-MSG-003`, `FR-MSG-005` |
| **FR-ABL-011** | Distributed Pool Value Allocations | `FR-TGT-002`, `FR-MSG-003`, `FR-MSG-005` |
| **FR-ABL-012** | Canonical Keyword AST Parameterization | `FR-KWD-002a` through `FR-KWD-002o` |
| **FR-TXT-001** | MTG-Aligned Sentence Grammar & Punctuation | `FR-CRD-101`, Appendix A |
| **FR-TXT-002** | Trigger Preambles (`When`, `Whenever`, `At`) | `FR-TRG-001`, `FR-TRN-001`, `FR-CBT-002` |
| **FR-TXT-003** | Cost Notation Syntax (`[Cost]: [Effect].`) | `FR-ACT-001`, `FR-TRN-003a`, `FR-ABL-009` |
| **FR-TXT-004** | Conjunction Targeting Phrasing (`chosen`) | `FR-TGT-001`, `FR-TGT-102`, `FR-MSG-003` |
| **FR-TXT-005** | Stat & Defense Modification Phrasing | `FR-LAY-001`, `FR-LAY-003`, `FR-REP-004` |
| **FR-TXT-006** | Zone Naming & Cardinality Conventions | `FR-ZON-001`, `FR-ZON-003`, `FR-DEK-001` |
| **FR-VAL-001** | Ingestion Pipeline Linter Specification | `FR-SET-001`, `FR-CRD-101`, `FR-DAT-006` |
| **FR-VAL-002** | Rules Text vs Ability AST Parity Checker | `FR-CRD-101`, `FR-ABL-001`, `FR-TXT-001` |
| **FR-VAL-003** | Placeholder Row Handling & Incomplete Blueprint Policy | `FR-SET-001`, `FR-DAT-005`, `FR-DAT-006` |

---
*End of Functional Requirements Document (Card Data Pipeline FRD v1.0)*
