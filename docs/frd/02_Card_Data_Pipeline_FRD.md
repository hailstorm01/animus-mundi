# Card Data Pipeline & Ability Authoring - Functional Requirements Document (FRD v1.0)
**Master Card Ingestion Pipeline, Declarative Ability AST, MTG-Aligned Text Standard & Schema Specification**

---

## EXECUTIVE SUMMARY & REVISION HISTORY

This document establishes the authoritative technical specification for the card data pipeline, declarative ability abstract syntax tree (AST), trigger/replacement event hook taxonomy, atomic effect primitives catalog, rules text templating standards, and dataset validation for Card Town. It serves as the authoritative **modeling bridge** between the pure tabletop game rules ([`01_Game_Rules_FRD.md`](01_Game_Rules_FRD.md)), the card compiler ([`03_Card_Compiler_FRD.md`](03_Card_Compiler_FRD.md)), the network protocol ([`04_Client_Engine_Protocol_Networking_FRD.md`](04_Client_Engine_Protocol_Networking_FRD.md)), and the headless execution engine ([`05_Headless_Game_Engine_FRD.md`](05_Headless_Game_Engine_FRD.md)).

### Specification Scope (v1.1):
1. **Zero-Aliasing Data Contract (`FR-DAT-006`)**: Codified strict enum target conformance across all authoring CSV balance sheets. Prohibited runtime shorthand aliasing; all raw card sheets must directly serialize canonical target enum identifiers (`UNIT`, `STRUCTURE`, `MANEUVER`, `HEADQUARTERS`; `CONCORDIA`, `SOULLESS`, `PATH`, `ANIMUS`, `NEUTRAL`; `COMMON`, `UNCOMMON`, `RARE`, `MYTHIC`, `HQ`; `RESERVES`, `MAIN_THEATER`, `HEADQUARTERS`).
2. **Deterministic Blueprint ID Derivation (`FR-DAT-002`)**: Formalized a deterministic, immutable blueprint slug derivation algorithm (`<faction>_<snake_case_title>`), decoupling human display titles from engine identity keys while guaranteeing 100% stable serialization.
3. **Production Blueprint JSON Schema & Ability AST (`FR-DAT-005`, `FR-ABL-001` - `FR-ABL-007`)**: Formalized the production `CardBlueprintCollection` JSON schema and polymorphic declarative schemas (`AbilityDef`, `FilterDef`, `ValueToken`, `CostDef`, `TargetRef`) across all six ability archetypes (`TRIGGERED`, `ACTIVATED`, `CONTINUOUS_AURA`, `REPLACEMENT_WATCHER`, `MANEUVER_EFFECT`, `KEYWORD_GRANT`).
4. **Event Hook Taxonomy, Atomic Primitives & Bound Payloads (`FR-TRG-001`, `FR-TRG-007`, `FR-PRM-001`, `FR-TGT-101`)**: Consolidated the canonical engine event/watcher hook catalog, the 26 atomic effect primitives and parameter schemas, bound action payload bindings, and canonical benchmark card AST implementations into the data pipeline bridge (`SOC-002`).
5. **Mechanics Extensions & Resolution Semantics (`FR-ABL-008` - `FR-ABL-012`)**: Formalized engine support for per-turn execution gates (`maxPerTurn`), exhaust-as-cost activation mechanics (`costType: EXHAUST_SELF`), modal branch definitions (`CHOOSE_N`), and distributed pool damage allocations (`DISTRIBUTED_POOL`).
6. **MTG-Aligned Rules Text Templating Grammar (`FR-TXT-001` - `FR-TXT-006`)**: Established formal English templating standards modeled on Magic: The Gathering (MTG) syntax, standardizing trigger preambles (`When`, `Whenever`, `At`), cost sequencing (`[Cost]: [Effect].`), passive auras (`have`), temporary modifiers (`gets +N power until end of turn`), and canonical keyword capitalization.
7. **Automated Verification & Linter Architecture (`FR-VAL-001` - `FR-VAL-003`) & Card Review Register (Domain 7)**: Expanded the automated verification suite (`tools/linter/lint_cardtown.ps1`) and audited all 167 cards across the 10 balance sheets.

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
- **Tier 2 (Cloud Inspection & Mobile Review Mirror)**: The Google Sheets workbooks located in the team drive (`maneuvers` and `units-and-structures`) serve purely as an outward-facing mirror and convenience viewer for designers to inspect card catalogs on mobile devices and explore balance experiments. The Google Sheets are downstream representations; any edits in Google Drive do not affect game builds and must be brought into Git CSVs via the ingestion workflow to become authoritative.
- **Tier 3 (Compiled Production Catalog)**: The compiled `blueprints.json` artifact (generated into `data/schemas/` or embedded in the game engine runtime) represents the authoritative machine-readable catalog consumed by the server engine process and Godot client builds.
- **Automated Pipeline Flow**:
  1. Local `data/cards/*.csv` files are directly edited, reviewed, and committed into Git.
  2. `tools/sheets/publish_card_sheets.ps1` publishes local repository CSV content to Google Sheets. To prevent accidental loss of remote designer work, `publish_card_sheets.ps1` executes a pre-publish freshness check comparing remote Google Drive revision timestamps against local Git commit timestamps. If un-synced remote edits exist, publication halts immediately with a collision warning unless the `--Force` switch is explicitly supplied.
  3. **Bidirectional Ingestion Tooling**: The pipeline provides a companion ingestion script (`tools/sheets/import_card_sheets.ps1`) to safely pull remote Google Sheets modifications into local `data/cards/*.csv` files, rendering an interactive terminal visual diff for designer confirmation before staging.
  4. `tools/linter/lint_cardtown.ps1` audits schema integrity, requirement links, and text terminology.
  5. The compilation pipeline (`tools/importer/compile_blueprints.ps1`) parses CSVs, validates values against target enums, binds declarative ability ASTs, and emits the immutable production JSON bundle.

> [!NOTE]
> **Design Rationale (DA-DAT-001)**: Mandating pre-publish freshness checks and introducing `import_card_sheets.ps1` prevents unconditional overwrites of remote designer balance work in Google Drive, establishing a safe bidirectional workflow.

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

#### FR-DAT-005: Canonical Blueprint JSON Schema Specification
In compliance with `FR-CRD-101` ([`01_Game_Rules_FRD.md`](01_Game_Rules_FRD.md)), every compiled `CardBlueprint` in `blueprints.json` validates against the following property table and production JSON Schema:

| Field | Type | Constraint | Description |
| :--- | :--- | :--- | :--- |
| `id` | `String` | Pattern: `^[a-z]+_[a-z0-9_]+$` | Deterministic alphanumeric slug (`FR-DAT-002`). |
| `name` | `String` | Non-empty | Title Case card title matching design sheet. |
| `cardType` | `CardTypeEnum` | Valid Enum | `HEADQUARTERS`, `UNIT`, `STRUCTURE`, `MANEUVER`. |
| `faction` | `FactionEnum` | Valid Enum | `CONCORDIA`, `SOULLESS`, `PATH`, `ANIMUS`, `NEUTRAL`. |
| `subtypes` | `Array<String>` | Array of Title Case tags | Normalized subtypes (`FR-DAT-007`). |
| `requisitionCost` | `Integer` | `>= 0` | Base Requisition deployment/casting cost. |
| `tacticsCost` | `Integer` | `>= 0` | Base Tactics deployment/casting cost. |
| `additionalDeploymentCosts` | `Array<CostDef>` | Optional CostDef Array | Non-resource deployment/enactment costs (sacrifices, discards; `FR-TRN-003a`). |
| `power` | `Integer` or `null` | `>= 0` (null for Maneuvers) | Base Combat Power (`FR-LAY-001`). |
| `health` | `Integer` or `null` | `>= 1` (null for Maneuvers) | Base Maximum Health (`FR-LAY-002`). |
| `requisitionYield` | `Integer` | `>= 0` (0 for Maneuvers) | Resources generated upon exhausting for Requisition. |
| `tacticsYield` | `Integer` | `>= 0` (0 for Maneuvers) | Resources generated upon exhausting for Tactics. |
| `deploymentZones` | `Array<DeploymentZoneEnum>` | Valid Enum Array | `["RESERVES"]`, `["MAIN_THEATER"]`, `["HEADQUARTERS"]`, or both. |
| `rarity` | `RarityEnum` | Valid Enum | `COMMON`, `UNCOMMON`, `RARE`, `MYTHIC`, `HQ`. |
| `rulesText` | `String` | MTG-Aligned Formatting | Standardized printed rules text (`FR-TXT-001`). |
| `currentProxy` | `Boolean` | Optional (`true` / `false`) | Development metadata flag indicating placeholder proxy art/design (`FR-VAL-003`). |
| `needsRework` | `Boolean` | Optional (`true` / `false`) | Development metadata flag indicating pending design rework (`FR-VAL-003`). |
| `abilities` | `Array<AbilityDef>` | Valid AST Array | Declarative definitions of abilities (`FR-ABL-001`). |

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
          "currentProxy": { "type": "boolean" },
          "needsRework": { "type": "boolean" },
          "abilities": { "type": "array", "items": { "type": "object" } }
        }
      }
    }
  }
}
```

#### FR-DAT-006: Direct Schema Header Parity & Zero-Aliasing Enums in CSVs
To eliminate translation layers and ensure absolute 1-to-1 schema fidelity with zero header aliasing, all authoring CSV files in `data/cards/` use camelCase column headers that directly mirror canonical `CardBlueprint` JSON property names (`FR-DAT-005`):
- **Unit, Structure & Headquarters Sheets (`units-and-structures-*.csv` — 15 columns)**:
  `"name","cardType","faction","subtypes","requisitionCost","tacticsCost","deploymentZones","rulesText","power","health","requisitionYield","tacticsYield","rarity","currentProxy","needsRework"`
- **Maneuver Sheets (`maneuvers-*.csv` — 9 columns)**:
  `"name","cardType","faction","requisitionCost","tacticsCost","rulesText","rarity","currentProxy","needsRework"`
  *(Maneuver sheets omit the 6 battlefield-asset columns, which deterministically compile into `subtypes: []`, `deploymentZones: []`, `power: null`, `health: null`, `requisitionYield: 0`, and `tacticsYield: 0`.)*

Furthermore, enum and boolean cells must strictly serialize canonical target strings with zero aliases:
1. **`cardType`**: Must equal strictly `UNIT`, `STRUCTURE`, `MANEUVER`, or `HEADQUARTERS`. Values such as `"Unit"`, `"Structure"`, `"Maneuver"`, and `"HQ"` are illegal.
2. **`faction`**: Must equal strictly `CONCORDIA`, `SOULLESS`, `PATH`, `ANIMUS`, or `NEUTRAL`. Title-cased values (e.g. `"Animus"`, `"Concordia"`) are illegal.
3. **`rarity`**: Must equal strictly `COMMON`, `UNCOMMON`, `RARE`, `MYTHIC`, or `HQ`. Single-letter abbreviations (`"C"`, `"U"`, `"R"`, `"M"`) are illegal.
4. **`deploymentZones`**: Must equal strictly `RESERVES`, `MAIN_THEATER`, `HEADQUARTERS`, or comma-separated combinations thereof (e.g., `"MAIN_THEATER, RESERVES"`). Informal labels such as `"Main"`, `"Reserves"`, or `"Headquarters"` are illegal.
5. **`currentProxy` & `needsRework`**: Must equal strictly lowercase `"true"` or `"false"`. Informal boolean strings such as `"Yes"`, `"No"`, `"TRUE"`, or `"FALSE"` are illegal.

#### FR-DAT-007: Subtype & Vanilla Ability Sentinel Normalization
- **Subtype Taxonomy**: Card subtypes in the `subtypes` CSV column must be serialized as a comma-separated list of Title Case strings (e.g., `"Infantry, Scratch Company"`, `"Specialist, Chainbreaker"`, `"Forge, Construct"`).
- **Null Subtypes Sentinel**: Where a card possesses no subtypes (e.g. Headquarters), the field typically contains the literal sentinel string `"--"`. The ingestion pipeline treats empty strings (`""`), whitespace-only cells, `"N/A"`, and `"--"` identically, normalizing them into an empty array `[]`.
- **Vanilla Ability Sentinel**: In the `rulesText` column, the sentinel string `"--"` is explicitly recognized as representing a vanilla card with no printed abilities (e.g., *Forest Heavies*, *Old Growth Elders*, *Spoil Forest Irregulars*), compiling into an empty `abilities: []` array.
- **Normalization Engine & Casing Semantics**: The compilation pipeline splits comma-separated strings, strips extraneous whitespace, and normalizes values to an array of unique Title Case strings for serialization and card frame rendering. During engine execution, all subtype filter comparisons (`FilterDef.subtypes`, `FR-TGT-001`) evaluate criteria using case-insensitive string matching to prevent casing mismatches from causing targeting failures.

> [!NOTE]
> **Design Rationale (DA-DAT-006)**: Authoring CSVs enforce Title Case for human readability and UI presentation, while the runtime engine performs case-insensitive subtype comparison to guarantee robust filter matching across platforms.

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
  - `triggerHook`: String enum matching `FR-TRG-001` (e.g., `EVENT_DEPLOY`, `EVENT_DESTROY`, `EVENT_ATTACK_DECLARE`, `EVENT_ATTACK_RESOLVED`, `EVENT_COMBAT_KILL`, `EVENT_TAKE_DAMAGE`, `EVENT_ZONE_CHANGE`, `EVENT_EXHAUST`, `EVENT_TURN_START`, `EVENT_TURN_END`).
  - `condition`: Declarative condition object (`ConditionDef`, `FR-ABL-007`) evaluating source, turn context, zone transition, or Cornered gate (`FR-KWD-002e`).
  - `optional`: Optional boolean (`default: false`) indicating `"you may"` trigger resolution (*Craterfall Ancient*, *Blind Defile*, *Eerie Remnant*).
  - `optionalCosts`: Optional `Array<CostDef>` for triggers requiring an optional non-resource cost to execute their effect (`"you may [Cost]. If you do, [Effect]"`, e.g., *Infested Apostle*, *Arkha, Sun's Fury*, *Redlined Brawler*, *Blind Defile*).
  - `targeting`: Optional targeting declaration (`TargetRef`) if the trigger requires player target selection.
  - `effect`: Execution payload (`EffectDef`, `FR-ABL-007`) specifying an atomic primitive (`FR-PRM-001`) or composite control-flow container (`COMPOSITE_SEQUENCE`, `CONDITIONAL_CLAUSE`).
  - `maxPerTurn`: Optional integer (`>= 1`) restricting maximum trigger frequency per turn (`FR-ABL-008`).
- **Upfront Target Binding for Conditional Resolution**: For composite triggers where an effect's magnitude or occurrence depends on an action resolved during execution (*Samsara, Living Tempest*, *Infested Apostle*), all external targets (e.g. `chosen asset`, `chosen opposing unit`) are designated and locked upfront when the trigger is enqueued onto the stack (`FR-TGT-101`). During resolution, if the condition or optional cost is not met, the conditional clause simply does not execute against the pre-locked target, avoiding secondary interactive prompts.

#### FR-ABL-003: ActivatedAbilityDef Schema & Cost Models
In compliance with **FR-ACT-001**, **FR-ACT-002**, and **FR-QUE-001**:
- Declared manually by the active player during the Main Phase while `Stack.isEmpty() == true`.
- Schema properties:
  - `costs`: Array of `CostDef` structures requiring simultaneous payment upon declaration (`FR-ABL-007`, `FR-ABL-009`).
  - `targeting`: Universal targeting filter and quantifier (`FR-TGT-001`, `FR-TGT-002`).
  - `effect`: Execution payload (`EffectDef`) specifying atomic primitives.
  - `repeatable`: Boolean indicating whether the ability may be activated multiple times per turn (default `true` unless constrained by costs or `maxPerTurn`).

#### FR-ABL-004: ContinuousAuraDef Schema & 6-Layer Interfacing
In compliance with **Section 6.2 of Master FRD v3.0 (`FR-LAY-001` - `FR-LAY-003`)**, **FR-CST-001**, and **FR-TRN-003a**:
- Evaluates continuously across board assets (or in `HAND` / `DISCARD` when `activeZone` is explicitly declared), registering static modifiers into Layer 4 (Auras) of the 6-layer attribute pipeline or granting static zone deployment permissions.
- Schema properties:
  - `activeZone`: Optional string enum (`"BATTLEFIELD"` default, `"HAND"`, or `"DISCARD"`). Specifies the zone in which the host card's continuous effect is active (e.g., `"HAND"` for self cost-reduction modifiers on *Swirling Dustdevil* and *Boneyard Amalgam*; `"DISCARD"` for alternate-zone deployment permission on *Veil Render*).
  - `scope`: `SELF`, `ALLIED_IN_ZONE`, `OPPOSING_IN_ZONE`, or `ALL_IN_ZONE`.
  - `filter`: Optional `FilterDef` narrowing which entities within `scope` receive the modifier (e.g., `{ "excludeSelf": true }` on *Animoceps Stalk*, or `{ "subtypes": ["Horror"] }` on *Hulking Incongruity*).
  - `modifiers`: Array of `AttributeModifierDef` specifying target attribute (`COMBAT_POWER`, `MAX_HEALTH`, `REQUISITION_COST`, `TACTICS_COST`, `KEYWORD_GRANT`, `SUBTYPE_GRANT`), mathematical operation (`ADD`, `SUBTRACT`, `SET`), and value expression (`ValueToken`). Dynamic cost reductions (`REQUISITION_COST`, `TACTICS_COST`) obey a minimum floor of `0` (`FR-CST-001`). (Note: Armor cannot be granted continuously by an aura; it exists strictly as a dynamic entity pool modified via one-shot events).
  - `abilityGrants`: Optional `Array<AbilityDef>` allowing continuous auras to project compound abilities (e.g. conditional keywords such as `Cornered: Resistant 1` on *Ancestral Monument*) onto scoped assets.
  - `staticPermission`: Optional object `{ "permissionType": "ALLOW_DEPLOY_FROM_ZONE", "sourceZone": "DISCARD", "additionalCosts": Array<CostDef> }` granting permission to deploy the host card from a non-hand zone by paying additional costs (`FR-TRN-003a`, *Veil Render*).

> [!NOTE]
> **Design Rationale (DA-DAT-004)**: Armor is an ephemeral consumable damage buffer (absorbed 1-for-1 before health deduction in Stage 4 of `FR-REP-004`). It exists strictly as a stateful pool modified via one-shot discrete events (e.g. *"When this unit deploys"*, *"Chosen unit gains 2 armor"*, *"Distribute 2 armor"*). A continuous aura granting "+1 Armor" is semantically equivalent to granting `Resistant 1` (`FR-KWD-002b`); granting dynamic Armor continuously would create undefined state loops (either continually regenerating depleted armor after each combat hit, creating an unintended infinite damage barrier, or requiring complex stateful damage tracking in a stateless aura layer). Defensive continuous auras must instead grant stat modifiers or damage reduction keywords (`Resistant N`, `Stalwart`).

#### FR-ABL-005: ReplacementWatcherDef Schema & Event Interception
In compliance with **FR-REP-001** through **FR-REP-004**:
- Intercepts prospective state events inline (`WOULD_TAKE_DAMAGE`, `WOULD_BE_DESTROYED`, `WOULD_DEPLOY`) without utilizing the Resolution Stack.
- Schema properties:
  - `interceptHook`: Event hook to intercept.
  - `condition`: Conjunction condition filter (`ConditionDef`).
  - `replacementAction`: Atomic mutation (`EffectDef`) substituting the prospective event payload (e.g., `REDIRECT_DAMAGE_TARGET`, `SET_INITIAL_READINESS`, `CANCEL_AND_BOUNCE`).

#### FR-ABL-006: ManeuverEffectDef Schema & Stack Resolution
In compliance with **FR-MNV-001** and **FR-TRN-003**:
- Represents the core spell payload executed when a Maneuver card resolves from the top of the Resolution Stack.
- Supports `targeting` (`TargetRef` or `Array<TargetRef>`), `effect` (`EffectDef`), and modal payloads (`modalCount`, `options`; `FR-ABL-010`).

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
        "cardType": { "type": "string", "enum": ["UNIT", "STRUCTURE", "MANEUVER", "HEADQUARTERS"] },
        "subtypes": { "type": "array", "items": { "type": "string" } },
        "negateSubtypes": { "type": "boolean" },
        "excludeSelf": { "type": "boolean" },
        "excludeScratch": { "type": "boolean" },
        "excludeHQ": { "type": "boolean" },
        "excludeUnits": { "type": "boolean" },
        "readiness": { "type": "string", "enum": ["READIED", "EXHAUSTED", "ANY"] },
        "zone": { "type": "string", "enum": ["SAME_ZONE", "RESERVES", "MAIN_THEATER", "HEADQUARTERS", "DISCARD", "HAND", "DECK"] },
        "statComparison": {
          "type": "object",
          "required": ["property", "operator", "value"],
          "properties": {
            "property": { "type": "string", "enum": ["COMBAT_POWER", "CURRENT_HEALTH", "MAX_HEALTH", "REQUISITION_COST", "TACTICS_COST"] },
            "operator": { "type": "string", "enum": ["EQ", "NEQ", "LT", "LTE", "GT", "GTE"] },
            "value": { "$ref": "#/definitions/ValueToken" }
          }
        }
      }
    },
    "ValueToken": {
      "type": "object",
      "required": ["type"],
      "properties": {
        "type": { "type": "string", "enum": ["STATIC", "DYNAMIC", "QUERY", "VARIABLE_X"] },
        "value": { "type": "integer" },
        "source": { "type": "string", "enum": ["SELF", "TARGET", "GAME_STATE", "CAUSALITY_CONTEXT"] },
        "property": { "type": "string", "enum": ["COMBAT_POWER", "CURRENT_HEALTH", "MAX_HEALTH", "BASE_REQUISITION_COST", "BASE_TACTICS_COST", "KILLED_COUNT", "DISCARDED_COUNT"] },
        "query": {
          "type": "object",
          "properties": {
            "zone": { "type": "string", "enum": ["ATTACHED", "DISCARD", "BATTLEFIELD", "HAND", "RESERVES", "MAIN_THEATER"] },
            "filter": { "$ref": "#/definitions/FilterDef" },
            "aggregate": { "type": "string", "enum": ["COUNT", "SUM_ATTRIBUTE"] },
            "targetAttribute": { "type": "string", "enum": ["REQUISITION_YIELD", "TACTICS_YIELD", "COMBAT_POWER", "HEALTH"] }
          }
        },
        "filter": { "$ref": "#/definitions/FilterDef" },
        "multiplier": { "type": "integer" },
        "divisor": { "type": "integer", "minimum": 1 }
      }
    },
    "CostDef": {
      "type": "object",
      "required": ["costType"],
      "properties": {
        "costType": { "type": "string", "enum": ["RESOURCE", "EXHAUST_SELF", "EXHAUST_ASSET", "KILL_ASSET", "DISCARD_CARD", "SACRIFICE_ALLY", "MOVE_CARD_ZONE"] },
        "resource": { "type": "string", "enum": ["REQUISITION", "TACTICS"] },
        "amount": { "type": "integer", "minimum": 1 },
        "filter": { "$ref": "#/definitions/FilterDef" },
        "destinationPosition": { "type": "string", "enum": ["TOP", "BOTTOM", "SHUFFLED_IN"] }
      }
    },
    "TargetRef": {
      "type": "object",
      "required": ["selector"],
      "properties": {
        "selector": { "type": "string", "enum": ["SELF", "CHOSEN", "ALL", "DISTRIBUTED_POOL", "RANDOM", "BOUND_TARGET", "TRIGGER_EVENT_ENTITY"] },
        "filter": { "$ref": "#/definitions/FilterDef" },
        "count": { "type": "integer", "minimum": 1 },
        "minCount": { "type": "integer", "minimum": 0 },
        "maxCount": { "type": "integer", "minimum": 1 },
        "allowedCounts": { "type": "array", "items": { "type": "integer", "minimum": 1 } },
        "poolValue": { "type": "integer", "minimum": 1 },
        "targetIndex": { "type": "integer", "minimum": 0 },
        "bindingPolicy": { "type": "string", "enum": ["UPFRONT_LOCK", "ON_RESOLUTION"] }
      }
    },
    "ConditionDef": {
      "type": "object",
      "properties": {
        "source": { "type": "string", "enum": ["SELF", "ALLIED", "OPPOSING", "ANY"] },
        "turnOwner": { "type": "string", "enum": ["SELF", "OPPONENT", "ANY"] },
        "cornered": { "type": "boolean" },
        "allegiance": { "type": "string", "enum": ["ALLIED", "OPPOSING", "ANY"] },
        "entityType": { "type": "string", "enum": ["UNIT", "STRUCTURE", "HEADQUARTERS", "ASSET", "CARD"] },
        "zone": { "type": "string", "enum": ["SAME_ZONE", "RESERVES", "MAIN_THEATER", "HEADQUARTERS", "DISCARD", "HAND", "DECK"] },
        "deployedZone": { "type": "string", "enum": ["RESERVES", "MAIN_THEATER"] },
        "fromZone": { "type": "string", "enum": ["HAND", "DECK", "BATTLEFIELD", "DISCARD", "RESERVES", "MAIN_THEATER"] },
        "toZone": { "type": "string", "enum": ["HAND", "DECK", "BATTLEFIELD", "DISCARD", "RESERVES", "MAIN_THEATER"] },
        "targetRef": { "type": "string" },
        "cardType": { "type": "string", "enum": ["UNIT", "STRUCTURE", "MANEUVER", "HEADQUARTERS"] },
        "handEmpty": { "type": "boolean" },
        "filter": { "$ref": "#/definitions/FilterDef" }
      }
    },
    "EffectDef": {
      "type": "object",
      "required": ["primitive"],
      "properties": {
        "primitive": { "type": "string" },
        "target": { "$ref": "#/definitions/TargetRef" },
        "steps": { "type": "array", "items": { "$ref": "#/definitions/EffectDef" } },
        "condition": { "$ref": "#/definitions/ConditionDef" },
        "optionalCost": { "$ref": "#/definitions/CostDef" },
        "effect": { "$ref": "#/definitions/EffectDef" },
        "ifTrue": { "$ref": "#/definitions/EffectDef" },
        "ifFalse": { "$ref": "#/definitions/EffectDef" }
      }
    }
  }
}
```

> [!NOTE]
> **Design Rationale (DA-DAT-002: Composable Query Primitives)**:
> Replaces ad-hoc, card-specific property enums (`COUNT_TUCKED_CARDS`, `COUNT_DISCARD_CARDS`, `COUNT_STRUCTURES_REQUISITION_YIELD`) with a generic, composable `query` object within `ValueToken`. This fulfills the core declarative AST architectural mandate (`Design Rationale DA-XP-001`), allowing game designers to introduce arbitrary entity counts and attribute aggregates (e.g. cards tucked under a host, cards in discard piles, or cumulative yields of controlled assets) without requiring engine recompilation or schema enum expansions.

---

### Section 3.3: Event Hook Taxonomy, Atomic Effect Primitives & Bound Payloads

#### FR-TRG-001: Trigger & Replacement Event Hook Taxonomy
`TriggeredAbilityDef` (`FR-ABL-002`) and `ReplacementWatcherDef` (`FR-ABL-005`) bind strictly to the following canonical event and prospective replacement hook identifiers:
- **Asset Lifecycle & Zones**: `EVENT_DEPLOY`, `EVENT_DESTROY`, `EVENT_ZONE_CHANGE`, `EVENT_MARCH`, `EVENT_EXHAUST`, `EVENT_READY`
- **Combat & Damage**: `EVENT_ATTACK_DECLARE`, `EVENT_DEAL_DAMAGE`, `EVENT_TAKE_DAMAGE`, `EVENT_ATTACK_RESOLVED`, `EVENT_COMBAT_KILL`
  - **Attacker-Only Combat Kill Semantics (`EVENT_COMBAT_KILL`)**: `EVENT_COMBAT_KILL` is emitted strictly when an attacking unit or Headquarters deals lethal combat damage that destroys a defending unit or structure (`"Whenever this [unit/Headquarters] attacks and kills..."`, e.g., *Eternal Watchman*, *Dread Manifestation*). Mutual combat destruction where a defender kills an attacker does not emit `EVENT_COMBAT_KILL` for the defender.
- **Economy & Resources**: `EVENT_GENERATE_REQUISITION`, `EVENT_SPEND_REQUISITION`, `EVENT_CONVERT_REQUISITION`, `EVENT_GENERATE_TACTICS`, `EVENT_SPEND_TACTICS`
- **Hand & Deck Operations**: `EVENT_DRAW_CARD`, `EVENT_DISCARD_CARD`, `EVENT_MULLIGAN`
- **Actions, Maneuvers & Abilities**: `EVENT_ENACT_MANEUVER`, `EVENT_ACTIVATE_ABILITY`, `EVENT_SCRATCH_COMPANY`
- **Phase & Turn Lifecycle**: `EVENT_READY_PHASE`, `EVENT_DRAW_PHASE`, `EVENT_MAIN_PHASE`, `EVENT_END_PHASE`, `EVENT_TURN_START`, `EVENT_TURN_END`
  - **Canonical Turn-Boundary Trigger Binding**: Card abilities templated with *"At the start of your turn, ..."* (*Hammerhand Bivouac*) bind strictly to `EVENT_TURN_START` (emitted once in Sub-Phase 1b of the Ready Phase after asset readying), and abilities templated with *"At the end of your turn, ..."* (*Redline Infusion*) bind strictly to `EVENT_TURN_END` (emitted once in Sub-Phase 4a of the End Phase). The phase-entry hooks (`EVENT_READY_PHASE`, `EVENT_DRAW_PHASE`, `EVENT_MAIN_PHASE`, `EVENT_END_PHASE`) are reserved for internal engine sub-phase transitions and are never bound simultaneously with `EVENT_TURN_START` / `EVENT_TURN_END` on card triggers.
- **Prospective Replacement Interception Hooks (`FR-REP-001`)**: `WOULD_TAKE_DAMAGE`, `WOULD_BE_DESTROYED`, `WOULD_DEPLOY`

#### FR-TRG-007: Event Causality Context Lineage
- Every emitted event and prospective replacement hook carries causal lineage metadata (`parentEventId`, `rootActionId`, and `sourceEntitySnapshot` per `FR-SNP-101`) to enable deterministic replacement evaluation, `CAUSALITY_CONTEXT` `ValueToken` resolution (e.g., *"for each card discarded this way"*, *"for every 2 units killed this way"*), and trigger scope validation.

#### FR-PRM-001: Standard Atomic Effect Primitives Catalog
Every `AbilityDef` effect node (`FR-ABL-001`, `EffectDef` in `FR-ABL-007`) dispatches to one of the 26 canonical atomic effect primitives (or composite control-flow containers `COMPOSITE_SEQUENCE` / `CONDITIONAL_CLAUSE` wrapping these primitives):

| Primitive | Parameter Schema | Rules & Engine Traceability |
| :--- | :--- | :--- |
| `DEAL_DAMAGE` | `amount: ValueToken, damageSource: TargetRef (optional, defaults to SELF), target: TargetRef (optional)` | Deals damage from `damageSource` routed through the 5-stage mitigation sequence (`FR-REP-004`). |
| `HEAL` | `amount: ValueToken` | Removes accumulated damage counters up to `amount` (`FR-KWD-002d`). |
| `ADD_ARMOR` | `amount: ValueToken` | Increments dynamic Armor pool (`FR-KWD-002k`). |
| `MODIFY_STAT` | `attribute: Enum, amount: ValueToken, duration: Enum` | Modifies stat (`END_OF_TURN`, `END_OF_OPPONENT_TURN`, `persistent`; `FR-LAY-003`). |
| `MOVE_ZONE` | `destinationZone: Enum` | Transitions unit between `RESERVES` and `MAIN_THEATER` (`FR-ZON-002`). |
| `RETURN_TO_HAND` | `none` | Returns asset to owner's hand and resets dynamic state (`FR-ZON-004`). |
| `KILL` | `none` | Destroys target asset immediately and routes to owner's `DISCARD` (`FR-ZON-002`). |
| `DRAW_CARD` | `count: ValueToken` | Draws cards from owner's deck (triggers `FR-WIN-001` on empty deck). |
| `DISCARD_CARD` | `count: ValueToken, targetPlayer: Enum` | Causes target player to discard cards from hand. |
| `REVEAL_HAND` | `targetPlayer: Enum, selector: Enum` | Reveals player's hand for inspection and targeted discard (`FR-MOD-001`). |
| `SCRATCH` | `count: ValueToken` | Sequentially executes `Deploy Scratch Company` (`FR-INH-004`, `FR-KWD-002m`). |
| `REGROUP` | `count: ValueToken` | Transitions up to `count` chosen exhausted allied assets to `READIED` (`FR-KWD-002n`). |
| `DEPLOY_FREE` | `sourceZone: Enum, filter: FilterDef, destinationZone: Enum (optional), count: ValueToken (optional)` | Deploys matching card suppressing resource costs, optionally forcing deployment into `destinationZone` (`FR-TRN-003a`). |
| `SEARCH_DECK` | `filter: FilterDef, destination: Enum` | Searches deck and executes mandatory shuffle (`FR-DEK-001`, `FR-DEK-002`). |
| `LOOK_TOP_CARDS` | `count: Integer, operations: Array<OpDef>` | Inspects top `count` cards and partitions across destinations (`FR-MOD-001`). |
| `TUCK_UNDER` | `hostTarget: TargetRef, faceDown: Boolean` | Transitions target card/asset into `ATTACHED` sub-zone under host asset (`FR-ENT-102`). |
| `GAIN_CONTROL` | `targetPlayer: Enum` | Transfers controller for an opposing `MAIN_THEATER` unit (`FR-ZON-002`). |
| `MORPH_SCRATCH` | `none` | Transforms unit face-down into a `1/1/1/1` Scratch Company (`FR-ENT-103`). |
| `GENERATE_RESOURCE` | `resourceType: Enum, amount: ValueToken` | Adds resources to active player's temporary pool (`FR-RES-001`). |
| `RECYCLE_DISCARD_TO_DECK` | `targetPlayer: Enum` | Transfers `DISCARD` pile into `DECK` with mandatory shuffle (`FR-DEK-002`). |
| `REDIRECT_DAMAGE_TARGET` | `newTarget: TargetRef` | Intercepts incoming damage in Stage 1 and re-routes to `newTarget` (`FR-REP-004`). |
| `PUT_CARD_ON_DECK` | `destinationPosition: Enum [TOP, BOTTOM, SHUFFLED_IN], targetCard: TargetRef` | Places target card at specified deck position (`FR-DEK-001`). |
| `EXHAUST` | `targetAsset: TargetRef` | Transitions target asset readiness to `EXHAUSTED` (emitting `EVENT_EXHAUST`). |
| `GRANT_SUBTYPE` | `subtype: String, duration: Enum` | Temporarily grants subtype to target asset (`END_OF_TURN`, `END_OF_OPPONENT_TURN`). |
| `GRANT_KEYWORD` | `keyword: Enum, parameter: ValueToken (optional), duration: Enum` | Temporarily grants keyword to target asset (`END_OF_TURN`, `END_OF_OPPONENT_TURN`). |
| `REGISTER_DELAYED_TRIGGER` | `triggerHook: Enum, duration: Enum [END_OF_TURN, NEXT_OCCURRENCE], condition: ConditionDef, effect: EffectDef` | Registers a turn-scoped delayed triggered ability on a bound target or turn boundary (*Rage's Wake*, *Redline Infusion*). |

#### FR-TGT-101: Bound Action Payload & Upfront Target Locking Model
- Action declarations and trigger target selections bind against the AST via structured payloads specifying `actionSourceId`, `blueprintAbilityIndex`, `selectedModalIndices` (`FR-ABL-010`), `targetBindings` (with `TargetRole`: `SOURCE`, `TARGET`, `SECONDARY_BERSERK_TARGET`, `SACRIFICE_COST`, `DISCARD_COST`, `EXHAUST_COST`, `DESTINATION_ZONE`, and optional `allocatedValue` per `FR-ABL-011`), and `deckSelectionIndices`.
- **Upfront Target Locking (`bindingPolicy: "UPFRONT_LOCK"`)**: For composite abilities or attacks where a secondary effect depends on an outcome determined during resolution (*Samsara, Living Tempest*, or **Berserk** secondary Main Theater targets in `FR-KWD-002o`), external targets are bound upfront when the action or trigger is placed onto the Resolution Stack (`FR-TGT-001`), completing execution atomically without a mid-resolution target prompt.

---

## DOMAIN 4: EXTENDED GAME MECHANICS & EXECUTION SEMANTICS

### Section 4.1: Extended Mechanics Formalization

#### FR-ABL-008: Per-Turn Trigger & Execution Frequency Limits
- To model cards with explicit frequency caps (*Animoceps Rager*: *"This ability triggers only once each turn"*), `TriggeredAbilityDef` and `ActivatedAbilityDef` support an optional `maxPerTurn: Integer` property.
- **Engine Execution**: The engine tracks an instance execution counter (`turnActivationCount`) on each `AssetInstance`. When the counter reaches `maxPerTurn`, the ability gate fails and suppresses trigger enqueuing or activated command validation. All per-turn counters reset to zero during the Ready Phase (`FR-TRN-001`).

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
  - Cards distributing a pool of values across multiple targets (*Macrocannon Barrage*: *"Distribute 5 damage among up to 5 chosen units"*, *Hammerhand Bivouac*: *"At the start of your turn, distribute 2 armor among allied units in the same zone."*) specify `selector: "DISTRIBUTED_POOL"`.
  - The client provides `TargetBinding` entries where `allocated_value` sums exactly to `poolValue`.

#### FR-ABL-012: Canonical Keyword AST Parameterization
All 15 engine keywords defined in [`01_Game_Rules_FRD.md`](01_Game_Rules_FRD.md) Appendix A map deterministically between their `CardBlueprint` JSON serialization (`FR-ABL-001`) and their runtime engine execution semantics:

| Keyword | Blueprint Serialization Form (`FR-ABL-001`) | Runtime Engine Hook / Execution Semantics |
| :--- | :--- | :--- |
| **Unbowed** | `TRIGGERED` (`triggerHook: EVENT_TAKE_DAMAGE`) | Evaluates on `EVENT_TAKE_DAMAGE` when `damageTaken > 0` (`FR-KWD-002a`). |
| **Resistant N** | `KEYWORD_GRANT` (`keyword: "RESISTANT", parameter: N`) | Registers inline `WOULD_TAKE_DAMAGE` Stage 2 flat reduction `N` (`FR-KWD-002b`). |
| **Stalwart** | `KEYWORD_GRANT` (`keyword: "STALWART"`) | Registers inline `WOULD_TAKE_DAMAGE` Stage 3 combat damage cap `min(dmg, 1)` (`FR-KWD-002c`). |
| **Heal N** | `EffectDef` (`primitive: "HEAL", amount: N`) | Atomic effect primitive inside `TRIGGERED`, `ACTIVATED`, or `MANEUVER_EFFECT` (`FR-KWD-002d`). |
| **Cornered** | `ConditionDef` (`cornered: true`) | Conditional gate (`count(opposingAssets) > count(alliedAssets)`) on `TRIGGERED` or `CONTINUOUS_AURA` (`FR-KWD-002e`). |
| **Aggressive N** | `KEYWORD_GRANT` (`keyword: "AGGRESSIVE", parameter: N`) | Registers `EVENT_ATTACK_DECLARE` trigger applying `MODIFY_STAT(+N Power)` until end of turn (`FR-KWD-002f`). |
| **Ethereal** | `KEYWORD_GRANT` (`keyword: "ETHEREAL"`) | Static targeting restriction: ineligible defender for attack declarations (`FR-KWD-002g`, `FR-CBT-001`). |
| **Ward** | `KEYWORD_GRANT` (`keyword: "WARD"`) | Static targeting restriction: ineligible target for non-combat enemy selections (`FR-KWD-002h`). |
| **Versatile** | `KEYWORD_GRANT` (`keyword: "VERSATILE"`) | Capability flag: unit inherent generator yields either Requisition or Tactics (`FR-KWD-002i`, `FR-CAP-001`). |
| **Guardian** | `KEYWORD_GRANT` (`keyword: "GUARDIAN"`) | Static combat restriction: mandatory attack target priority when able (`FR-KWD-002j`, `FR-CBT-001`). |
| **Armor N** | `KEYWORD_GRANT` (`keyword: "ARMOR", parameter: N`) | Initializes dynamic Armor pool to `N` upon deployment (`FR-KWD-002k`, `FR-REP-004`). |
| **Deploys Exhausted**| `KEYWORD_GRANT` (`keyword: "DEPLOYS_EXHAUSTED"`) | Registers `WOULD_DEPLOY` replacement setting initial readiness to `EXHAUSTED` (`FR-KWD-002l`). |
| **Scratch N** | `EffectDef` (`primitive: "SCRATCH", count: N`) | Atomic effect primitive executing `Deploy Scratch Company` `N` times sequentially (`FR-KWD-002m`, `FR-INH-004`). |
| **Regroup N** | `EffectDef` (`primitive: "REGROUP", count: N`) | Atomic effect primitive transitioning up to `N` exhausted allied assets to `READIED` (`FR-KWD-002n`). |
| **Berserk** | `KEYWORD_GRANT` (`keyword: "BERSERK"`) | Static combat capability enabling attack declarations targeting up to 2 Main Theater assets (`FR-KWD-002o`). |

- **Dynamic & Continuous Grants**: Keywords and subtypes may also be granted dynamically via atomic primitives `GRANT_KEYWORD` and `GRANT_SUBTYPE` (`FR-PRM-001`) or continuously via `ContinuousAuraDef` modifiers (`KEYWORD_GRANT`, `SUBTYPE_GRANT`) and compound `abilityGrants` (Layer 6; `FR-ABL-004`).

---

### Section 4.2: Canonical Benchmark Blueprint Implementations

#### 1. `NO-AH-5` (Soulless Mythic Unit - Tucking & Composable Query Scaling)
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
  "rulesText": "Ward\nR2: Kill chosen allied non-Headquarters asset and tuck it face down under this unit. This unit has +2 power for each card tucked under it.",
  "abilities": [
    { "abilityType": "KEYWORD_GRANT", "keyword": "WARD" },
    {
      "abilityType": "ACTIVATED",
      "costs": [
        { "costType": "RESOURCE", "resource": "REQUISITION", "amount": 2 },
        { "costType": "KILL_ASSET", "filter": { "allegiance": "ALLIED", "entityType": "ASSET", "excludeHQ": true } }
      ],
      "effect": { "primitive": "TUCK_UNDER", "hostTarget": { "selector": "SELF" }, "faceDown": true }
    },
    {
      "abilityType": "CONTINUOUS_AURA",
      "scope": "SELF",
      "modifiers": [
        {
          "attribute": "COMBAT_POWER",
          "operation": "ADD",
          "value": {
            "type": "QUERY",
            "source": "SELF",
            "query": { "zone": "ATTACHED", "aggregate": "COUNT" },
            "multiplier": 2
          }
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
    { "abilityType": "KEYWORD_GRANT", "keyword": "AGGRESSIVE", "parameter": 9 },
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
      "replacementAction": { "primitive": "REDIRECT_DAMAGE_TARGET", "newTarget": { "selector": "SELF" } }
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

#### 5. `Swirling Dustdevil` (Concordia Common Unit - Hand-Active Self Cost Reduction)
```json
{
  "id": "concordia_swirling_dustdevil",
  "name": "Swirling Dustdevil",
  "cardType": "UNIT",
  "faction": "CONCORDIA",
  "subtypes": ["Elemental"],
  "requisitionCost": 8,
  "tacticsCost": 0,
  "power": 6,
  "health": 5,
  "requisitionYield": 2,
  "tacticsYield": 2,
  "deploymentZones": ["RESERVES"],
  "rarity": "COMMON",
  "rulesText": "This unit costs R1 less to deploy for each maneuver card in your discard pile.",
  "abilities": [
    {
      "abilityType": "CONTINUOUS_AURA",
      "activeZone": "HAND",
      "scope": "SELF",
      "modifiers": [
        {
          "attribute": "REQUISITION_COST",
          "operation": "SUBTRACT",
          "value": {
            "type": "QUERY",
            "source": "GAME_STATE",
            "query": {
              "zone": "DISCARD",
              "filter": { "allegiance": "ALLIED", "entityType": "CARD", "cardType": "MANEUVER" },
              "aggregate": "COUNT"
            },
            "multiplier": 1
          }
        }
      ]
    }
  ]
}
```

#### 6. `Veil Render` (Animus Mythic Unit - Discard-Pile Static Deployment Permission)
```json
{
  "id": "animus_veil_render",
  "name": "Veil Render",
  "cardType": "UNIT",
  "faction": "ANIMUS",
  "subtypes": ["Spirit", "Horror"],
  "requisitionCost": 3,
  "tacticsCost": 3,
  "power": 5,
  "health": 4,
  "requisitionYield": 1,
  "tacticsYield": 2,
  "deploymentZones": ["RESERVES"],
  "rarity": "MYTHIC",
  "rulesText": "You may deploy this card from your discard pile by killing an allied asset in addition to paying its deployment costs.",
  "abilities": [
    {
      "abilityType": "CONTINUOUS_AURA",
      "activeZone": "DISCARD",
      "scope": "SELF",
      "staticPermission": {
        "permissionType": "ALLOW_DEPLOY_FROM_ZONE",
        "sourceZone": "DISCARD",
        "additionalCosts": [
          { "costType": "KILL_ASSET", "filter": { "allegiance": "ALLIED", "entityType": "ASSET" } }
        ]
      }
    }
  ]
}
```

#### 7. `Rage's Wake` (Concordia Common Maneuver - Delayed Trigger & Forced Free-Deploy Zone)
```json
{
  "id": "concordia_rages_wake",
  "name": "Rage's Wake",
  "cardType": "MANEUVER",
  "faction": "CONCORDIA",
  "subtypes": [],
  "requisitionCost": 0,
  "tacticsCost": 3,
  "power": null,
  "health": null,
  "requisitionYield": 0,
  "tacticsYield": 0,
  "deploymentZones": [],
  "rarity": "COMMON",
  "rulesText": "Chosen unit gets +2 power until end of turn. When that unit is killed this turn, deploy it from your discard pile to your Reserves for free.",
  "abilities": [
    {
      "abilityType": "MANEUVER_EFFECT",
      "targeting": {
        "selector": "CHOSEN",
        "filter": { "entityType": "UNIT" },
        "count": 1
      },
      "effect": {
        "primitive": "COMPOSITE_SEQUENCE",
        "steps": [
          {
            "primitive": "MODIFY_STAT",
            "target": { "selector": "BOUND_TARGET", "targetIndex": 0 },
            "attribute": "COMBAT_POWER",
            "amount": { "type": "STATIC", "value": 2 },
            "duration": "END_OF_TURN"
          },
          {
            "primitive": "REGISTER_DELAYED_TRIGGER",
            "target": { "selector": "BOUND_TARGET", "targetIndex": 0 },
            "triggerHook": "EVENT_DESTROY",
            "duration": "END_OF_TURN",
            "condition": { "source": "SELF" },
            "effect": {
              "primitive": "DEPLOY_FREE",
              "sourceZone": "DISCARD",
              "destinationZone": "RESERVES",
              "target": { "selector": "SELF" }
            }
          }
        ]
      }
    }
  ]
}
```

#### 8. `Funeral Fungus` (Animus Uncommon Maneuver - Additional Enactment Exhaust Cost)
```json
{
  "id": "animus_funeral_fungus",
  "name": "Funeral Fungus",
  "cardType": "MANEUVER",
  "faction": "ANIMUS",
  "subtypes": [],
  "requisitionCost": 0,
  "tacticsCost": 1,
  "additionalDeploymentCosts": [
    {
      "costType": "EXHAUST_ASSET",
      "filter": { "allegiance": "ALLIED", "entityType": "UNIT", "readiness": "READIED" }
    }
  ],
  "power": null,
  "health": null,
  "requisitionYield": 0,
  "tacticsYield": 0,
  "deploymentZones": [],
  "rarity": "UNCOMMON",
  "rulesText": "As an additional cost to enact this maneuver, exhaust a readied allied unit.\nReturn chosen unit card from your discard pile to your hand.",
  "abilities": [
    {
      "abilityType": "MANEUVER_EFFECT",
      "targeting": {
        "selector": "CHOSEN",
        "filter": { "allegiance": "ALLIED", "entityType": "CARD", "cardType": "UNIT", "zone": "DISCARD" },
        "count": 1
      },
      "effect": {
        "primitive": "RETURN_TO_HAND",
        "target": { "selector": "BOUND_TARGET", "targetIndex": 0 }
      }
    }
  ]
}
```

#### 9. `Infested Apostle` (Animus Common Unit - Optional Non-Resource Trigger Cost)
```json
{
  "id": "animus_infested_apostle",
  "name": "Infested Apostle",
  "cardType": "UNIT",
  "faction": "ANIMUS",
  "subtypes": ["Fungus", "Mage"],
  "requisitionCost": 4,
  "tacticsCost": 0,
  "power": 4,
  "health": 5,
  "requisitionYield": 1,
  "tacticsYield": 2,
  "deploymentZones": ["RESERVES"],
  "rarity": "COMMON",
  "rulesText": "When this unit deploys, you may kill an allied unit. If you do, deal 2 damage to chosen opposing unit.",
  "abilities": [
    {
      "abilityType": "TRIGGERED",
      "triggerHook": "EVENT_DEPLOY",
      "condition": { "source": "SELF" },
      "optional": true,
      "optionalCosts": [
        { "costType": "KILL_ASSET", "filter": { "allegiance": "ALLIED", "entityType": "UNIT" } }
      ],
      "targeting": {
        "selector": "CHOSEN",
        "filter": { "allegiance": "OPPOSING", "entityType": "UNIT" },
        "count": 1,
        "bindingPolicy": "UPFRONT_LOCK"
      },
      "effect": {
        "primitive": "DEAL_DAMAGE",
        "target": { "selector": "BOUND_TARGET", "targetIndex": 0 },
        "amount": { "type": "STATIC", "value": 2 }
      }
    }
  ]
}
```

#### 10. `Soul Whip` (Concordia Uncommon Maneuver - Delegated Damage Source)
```json
{
  "id": "concordia_soul_whip",
  "name": "Soul Whip",
  "cardType": "MANEUVER",
  "faction": "CONCORDIA",
  "subtypes": [],
  "requisitionCost": 0,
  "tacticsCost": 2,
  "power": null,
  "health": null,
  "requisitionYield": 0,
  "tacticsYield": 0,
  "deploymentZones": [],
  "rarity": "UNCOMMON",
  "rulesText": "Chosen allied unit deals damage equal to its power to chosen opposing unit.",
  "abilities": [
    {
      "abilityType": "MANEUVER_EFFECT",
      "targeting": [
        { "selector": "CHOSEN", "filter": { "allegiance": "ALLIED", "entityType": "UNIT" }, "count": 1 },
        { "selector": "CHOSEN", "filter": { "allegiance": "OPPOSING", "entityType": "UNIT" }, "count": 1 }
      ],
      "effect": {
        "primitive": "DEAL_DAMAGE",
        "damageSource": { "selector": "BOUND_TARGET", "targetIndex": 0 },
        "target": { "selector": "BOUND_TARGET", "targetIndex": 1 },
        "amount": {
          "type": "DYNAMIC",
          "source": "TARGET",
          "property": "COMBAT_POWER"
        }
      }
    }
  ]
}
```

#### 11. `Allshan's Battle Trance` (Path Mythic Maneuver - Forced Free-Deploy Destination Zone)
```json
{
  "id": "path_allshans_battle_trance",
  "name": "Allshan's Battle Trance",
  "cardType": "MANEUVER",
  "faction": "PATH",
  "subtypes": [],
  "requisitionCost": 0,
  "tacticsCost": 4,
  "power": null,
  "health": null,
  "requisitionYield": 0,
  "tacticsYield": 0,
  "deploymentZones": [],
  "rarity": "MYTHIC",
  "rulesText": "Deploy chosen unit card from your discard pile to the Main Theater. It gains Stalwart and gets +2 power until end of turn.",
  "abilities": [
    {
      "abilityType": "MANEUVER_EFFECT",
      "targeting": {
        "selector": "CHOSEN",
        "filter": { "allegiance": "ALLIED", "entityType": "CARD", "cardType": "UNIT", "zone": "DISCARD" },
        "count": 1
      },
      "effect": {
        "primitive": "COMPOSITE_SEQUENCE",
        "steps": [
          {
            "primitive": "DEPLOY_FREE",
            "sourceZone": "DISCARD",
            "destinationZone": "MAIN_THEATER",
            "target": { "selector": "BOUND_TARGET", "targetIndex": 0 }
          },
          {
            "primitive": "GRANT_KEYWORD",
            "target": { "selector": "BOUND_TARGET", "targetIndex": 0 },
            "keyword": "STALWART",
            "duration": "END_OF_TURN"
          },
          {
            "primitive": "MODIFY_STAT",
            "target": { "selector": "BOUND_TARGET", "targetIndex": 0 },
            "attribute": "COMBAT_POWER",
            "amount": { "type": "STATIC", "value": 2 },
            "duration": "END_OF_TURN"
          }
        ]
      }
    }
  ]
}
```

#### 12. `Consuming Earth` (Concordia Rare Maneuver - Dynamic Query Stat Comparison Filter)
```json
{
  "id": "concordia_consuming_earth",
  "name": "Consuming Earth",
  "cardType": "MANEUVER",
  "faction": "CONCORDIA",
  "subtypes": [],
  "requisitionCost": 0,
  "tacticsCost": 3,
  "power": null,
  "health": null,
  "requisitionYield": 0,
  "tacticsYield": 0,
  "deploymentZones": [],
  "rarity": "RARE",
  "rulesText": "Kill chosen opposing unit with power less than or equal to the total Requisition yield of all structures you control.",
  "abilities": [
    {
      "abilityType": "MANEUVER_EFFECT",
      "targeting": {
        "selector": "CHOSEN",
        "filter": {
          "allegiance": "OPPOSING",
          "entityType": "UNIT",
          "statComparison": {
            "property": "COMBAT_POWER",
            "operator": "LTE",
            "value": {
              "type": "QUERY",
              "source": "GAME_STATE",
              "query": {
                "zone": "BATTLEFIELD",
                "filter": { "allegiance": "ALLIED", "entityType": "STRUCTURE" },
                "aggregate": "SUM_ATTRIBUTE",
                "targetAttribute": "REQUISITION_YIELD"
              }
            }
          }
        },
        "count": 1
      },
      "effect": {
        "primitive": "KILL",
        "target": { "selector": "BOUND_TARGET", "targetIndex": 0 }
      }
    }
  ]
}
```

#### 13. `Tempest Runner` (Concordia Uncommon Unit - `excludeSelf` Targeting Filter)
```json
{
  "id": "concordia_tempest_runner",
  "name": "Tempest Runner",
  "cardType": "UNIT",
  "faction": "CONCORDIA",
  "subtypes": ["Cavalry", "Tempest"],
  "requisitionCost": 3,
  "tacticsCost": 0,
  "power": 3,
  "health": 2,
  "requisitionYield": 1,
  "tacticsYield": 2,
  "deploymentZones": ["MAIN_THEATER", "RESERVES"],
  "rarity": "UNCOMMON",
  "rulesText": "Exhaust: Regroup another chosen allied unit in the same zone.",
  "abilities": [
    {
      "abilityType": "ACTIVATED",
      "costs": [
        { "costType": "EXHAUST_SELF" }
      ],
      "targeting": {
        "selector": "CHOSEN",
        "filter": {
          "allegiance": "ALLIED",
          "entityType": "UNIT",
          "zone": "SAME_ZONE",
          "readiness": "EXHAUSTED",
          "excludeSelf": true
        },
        "count": 1
      },
      "effect": {
        "primitive": "REGROUP",
        "target": { "selector": "BOUND_TARGET", "targetIndex": 0 },
        "count": { "type": "STATIC", "value": 1 }
      }
    }
  ]
}
```

#### 14. `Mutation Swell` (Animus Rare Maneuver - Discrete Quantifier `allowedCounts` & Causality Divisor)
```json
{
  "id": "animus_mutation_swell",
  "name": "Mutation Swell",
  "cardType": "MANEUVER",
  "faction": "ANIMUS",
  "subtypes": [],
  "requisitionCost": 0,
  "tacticsCost": 3,
  "power": null,
  "health": null,
  "requisitionYield": 0,
  "tacticsYield": 0,
  "deploymentZones": [],
  "rarity": "RARE",
  "rulesText": "Kill 2 or 4 allied Scratch Company units. For every 2 units killed this way, deploy a unit card from your discard pile for free.",
  "abilities": [
    {
      "abilityType": "MANEUVER_EFFECT",
      "targeting": {
        "selector": "CHOSEN",
        "filter": {
          "allegiance": "ALLIED",
          "entityType": "UNIT",
          "subtypes": ["Scratch Company"]
        },
        "allowedCounts": [2, 4]
      },
      "effect": {
        "primitive": "COMPOSITE_SEQUENCE",
        "steps": [
          {
            "primitive": "KILL",
            "target": { "selector": "BOUND_TARGET", "targetIndex": 0 }
          },
          {
            "primitive": "DEPLOY_FREE",
            "sourceZone": "DISCARD",
            "filter": { "allegiance": "ALLIED", "entityType": "CARD", "cardType": "UNIT", "zone": "DISCARD" },
            "count": {
              "type": "DYNAMIC",
              "source": "CAUSALITY_CONTEXT",
              "property": "KILLED_COUNT",
              "divisor": 2
            }
          }
        ]
      }
    }
  ]
}
```

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
*(Subsumed by FR-CMP-001 & FR-CMP-022)*
The automated linter (`tools/linter/lint_cardtown.ps1`) and compiler pre-flight validation phases (`FR-CMP-001`) enforce requirement cross-reference integrity across all FRDs, CSV RFC-4180 parsing, and canonical schema conformance (`FR-CMP-022`).

#### FR-VAL-002: Rules Text vs Declarative Ability Parity Checker
- The compilation pipeline executes automated parity checks ensuring that card abilities defined in text correspond 1-to-1 with compiled AST structures.
- A card declaring `Stalwart` in rules text must contain a `KEYWORD_GRANT` (`keyword: "STALWART"`) AST entry; a card declaring an attack trigger must bind to `EVENT_ATTACK_DECLARE`, `EVENT_ATTACK_RESOLVED`, or `EVENT_COMBAT_KILL`.

#### FR-VAL-003: Uniform Dataset Validation Policy ("Must Fit All") & Review Card Compilability
- **Development & Prototyping Flexibility**: To support rapid game balance exploration and branch prototyping, placeholder or exploratory cards explicitly marked with `currentProxy == "true"` or `needsRework == "true"` in `data/cards/*.csv` compile normally into blueprints with `"currentProxy": true` / `"needsRework": true` metadata flags intact (`FR-CMP-005`). This enables designers to test rough mechanics in development sandboxes without failing local toolchains.
- **Universal Production Gating**: Release builds and CI production pipelines operating under strict mode (`-Strict`) enforce a zero-tolerance validation gate. Any card possessing incomplete ability logic, unreviewed placeholder text, or marked with `currentProxy == "true"` / `needsRework == "true"` triggers compilation error `ERR_STRICT_PROTOTYPE_REJECTED`, guaranteeing zero prototype leakage into release bundles.
- **Review Card Compilability Contract**: Cards undergoing active game balance or rules text review must maintain complete, functional ability bindings in development builds. Production releases require formal designer sign-off clearing all pending review flags from Domain 7.

> [!NOTE]
> **Design Rationale (DA-DAT-005)**: Harmonizes `FR-VAL-003` with `FR-CMP-005`. By permitting prototype cards to compile with explicit metadata flags during local development while blocking them under `-Strict`, designers can explore experimental concepts without compromising production build integrity.

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
   - *Design Call*: Confirmed self-damage interaction on Headquarters; emits `EVENT_TAKE_DAMAGE` and routes through damage mitigation (`FR-REP-004`). Damage-reactive triggers (e.g. Unbowed) trigger normally.
4. **Mutation Swell (Maneuver - Animus)**:
   - *Previous*: `"kill up to 4 allied scratch company units. for every 2 units killed, deploy 1 unit from your discard pile for free."`
   - *Standardized*: `"Kill 2 or 4 allied Scratch Company units. For every 2 units killed this way, deploy a unit card from your discard pile for free."`
   - *Design Call*: Clarified MTG causality clause ("killed this way") and restricted targets to discrete multiples (2 or 4) to eliminate fractional casualty floor ambiguity.
5. **Steel Rain & Macrocannon Barrage (Path / Soulless)**:
   - *Steel Rain*: Confirmed multi-targeting. Discarding up to 3 cards grants that many discrete instances of "deal 2 damage to chosen asset", allowing players to select up to 3 distinct assets or concentrate multiple instances on a single asset.
   - *Macrocannon Barrage*: Formalized distributed damage allocation syntax matching `DISTRIBUTED_POOL` (`FR-ABL-011`).
6. **Title Spellings & Renames**:
   - `reverbating terror` corrected to **Reverberating Terror**.
   - `Devilhorn Guerillas` corrected to **Devilhorn Guerrillas**.
   - `All out War` hyphenated to **All-Out War**.
   - `Daggerspine Rangers` renamed to **Spoil Forest Irregulars** (eliminating Warcraft "Daggerspine" prior art / IP collision).

---

### Section 7.2: Master Review Register

The following register details all cards where rules text was modified or requires designer review:

| Card Title | Sheet | Previous Text | Standardized Rules Text | Designer Review Flag / Rationale |
| :--- | :--- | :--- | :--- | :--- |
| **Rift** | `units-and-structures-HQ.csv` | `discard a card, exhaust: scratch 2, deal 5 damage to this asset` | `Discard a card, exhaust: Scratch 2. Deal 5 damage to this Headquarters.` | Resolved: Confirmed Headquarters self-damage emits `EVENT_TAKE_DAMAGE` and routes through damage mitigation pipeline (`FR-REP-004`); damage-reactive keywords evaluate normally. |
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
| **Steel Rain** | `maneuvers-path.csv` | `discard up to 3 cards. For each card discarded this way, deal 2 damage to chosen asset. then, if you have no cards left in your hand, draw 2 cards.` | `Discard up to 3 cards. For each card discarded this way, deal 2 damage to chosen asset. Then, if you have no cards in hand, draw 2 cards.` | Resolved: Multi-targeting confirmed. Resolves as iterative targeting instances per card discarded (e.g. discarding 3 cards yields 3 distinct choices of "chosen asset" dealing 2 damage each). |
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
| **Dread Manifestation** | `units-and-structures-concordia.csv` | `Whenever this unit kills another unit in combat, you may move chosen unit in the Main Theater to its controller's Reserves. Opposing units in the same zone have -2 power.` | `Whenever this unit attacks and kills another unit, you may move chosen unit in the Main Theater to its controller's Reserves. Opposing units in the same zone have -2 power.` | Restricted trigger to attacker-only combat kills (`"attacks and kills"`) matching `EVENT_COMBAT_KILL` (`FR-TRG-001`). |
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
| **Hammerhand Bivouac** | `units-and-structures-path.csv` | `At the start of your turn, distribute 1 armor among allied units in the same zone.` | `At the start of your turn, distribute 2 armor among allied units in the same zone.` | Rebalanced armor pool distribution from 1 to 2 (`FR-ABL-011`). |
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
| **Spoil Forest Irregulars** | `units-and-structures-concordia.csv` | Previous name: `Daggerspine Rangers`, Subtypes: `Mending, Specialists` | Subtypes: `Mending, Specialist` | Resolved: Renamed to Spoil Forest Irregulars to resolve Warcraft IP collision. Normalized plural subtype "Specialists" to canonical singular "Specialist" (`FR-DAT-007`). |

> [!NOTE]
> **Design Rationale (DA-DAT-003: Core Set Mechanical Triage)**:
> Resolves open ambiguities identified in audit:
> - **Steel Rain**: Multi-targeting confirmed; resolves as iterative targeting instances per card discarded.
> - **Rift**: Headquarters self-damage emits `EVENT_TAKE_DAMAGE` and routes through damage mitigation (`FR-REP-004`).
> - **Card Rename**: Renamed *Daggerspine Rangers* to *Spoil Forest Irregulars* to eliminate Blizzard IP collision risk.

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
| **FR-TRG-001** | Trigger & Replacement Event Hook Taxonomy | `FR-ABL-002`, `FR-ABL-005`, `FR-TRG-003`, `FR-REP-001` |
| **FR-TRG-007** | Event Causality Context Lineage | `FR-TRG-001`, `FR-SNP-101`, `FR-ABL-007` |
| **FR-PRM-001** | Standard Atomic Effect Primitives Catalog (26 Opcodes) | `FR-ABL-001`, `FR-ENG-021` |
| **FR-TGT-101** | Bound Action Payload & Upfront Target Locking Model | `FR-TGT-001`, `FR-ABL-007`, `FR-MSG-003` |
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
*End of Functional Requirements Document (Card Data Pipeline FRD v1.1)*
