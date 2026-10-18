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

> [!NOTE] Design Rationale (DA-DAT-001)
> Mandating pre-publish freshness checks and introducing `import_card_sheets.ps1` prevents unconditional overwrites of remote designer balance work in Google Drive, establishing a safe bidirectional workflow.

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
In compliance with `FR-CRD-101` ([`01_Game_Rules_FRD.md`](01_Game_Rules_FRD.md)) and `FR-CMP-023` ([`03_Card_Compiler_FRD.md`](03_Card_Compiler_FRD.md)), `data/schemas/blueprints.json` uses the root `CardBlueprintCollection` envelope (`required: ["catalogChecksum", "cardDataVersion", "generatedAt", "totalCardCount", "blueprints"]`), where `catalogChecksum` matches `^0x[0-9a-f]{64}$` (`FR-CMP-025`) and every item in `blueprints` conforms to the following `CardBlueprint` schema (`required: ["id", "name", "cardType", "faction", "subtypes", "requisitionCost", "tacticsCost", "rarity", "rulesText", "abilities"]`):

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

> [!NOTE] Design Rationale (DA-DAT-006)
> Authoring CSVs enforce Title Case for human readability and UI presentation, while the runtime engine performs case-insensitive subtype comparison to guarantee robust filter matching across platforms.

#### FR-DAT-008: Deployment Zone Array Schema & Invariants
- **Units & Structures**: Every Unit and Structure blueprint must declare at least one legal deployment zone (`["RESERVES"]`, `["MAIN_THEATER"]`, or `["RESERVES", "MAIN_THEATER"]`). Headquarters cards declare strictly `["HEADQUARTERS"]`.
- **Maneuver Exemption**: Maneuvers resolve on the stack (`FR-QUE-001`, `FR-MNV-001`) and never occupy battlefield zones. Maneuver blueprints serialize `deploymentZones: []` (or `null`).
- **Zone Order Normalization**: In CSV authoring sheets where multi-zone deployment is permitted, zones are normalized alphabetically: `"MAIN_THEATER, RESERVES"`.

---

## DOMAIN 3: DECLARATIVE ABILITY ENGINE SCHEMA & AST SPECIFICATION

### Section 3.1: Polymorphic Ability AST Schemas

#### FR-ABL-001: Universal AbilityDef Polymorphic Schema
Every entry in a blueprint's `abilities` array is a polymorphic `AbilityDef` JSON object discriminated by a required `abilityType` enum property:
- `abilityType`: `"TRIGGERED"` (`FR-ABL-002`), `"ACTIVATED"` (`FR-ABL-003`), `"CONTINUOUS_AURA"` (`FR-ABL-004`), `"REPLACEMENT_WATCHER"` (`FR-ABL-005`), `"MANEUVER_EFFECT"` (`FR-ABL-006`), or `"KEYWORD_GRANT"` (`FR-ABL-012`).

> [!NOTE] Design Rationale (DA-XP-001)
> The declarative, polymorphic Ability AST architecture is intentional and non-negotiable. Digital card games that rely on ad-hoc imperative scripts or stopgap code hooks (such as early MTGO architecture) suffer from compounding technical debt, desync vulnerabilities, and fragile client-server state replication. A fully declarative JSON AST ensures deterministic replay verification, clean client-side visual effect binding, robust schema validation, and long-term extensibility without engine recompilation.

#### FR-ABL-002: TriggeredAbilityDef Schema & Hook Bindings
In compliance with **FR-TRG-001** through **FR-TRG-007**:
- Evaluates automatically whenever the engine emits a matching event hook and enqueues an execution frame onto the Resolution Stack (`FR-QUE-001`) during Sub-Phase 4a or immediately in context.
- **Schema Properties**:
  - `triggerHook`: String enum matching `FR-TRG-001`.
  - `condition`: Declarative condition object (`ConditionDef`, `FR-ABL-007`) evaluating source, turn context, zone transition, or Cornered gate (`FR-KWD-002e`).
  - `optional`: Optional boolean (`default: false`) indicating `"you may"` trigger resolution (*Craterfall Ancient*, *Blind Defile*, *Eerie Remnant*).
  - `optionalCosts`: Optional `Array<CostDef>` for triggers requiring an optional non-resource cost (`"you may [Cost]. If you do, [Effect]"`, e.g., *Infested Apostle*, *Arkha, Sun's Fury*, *Redlined Brawler*, *Blind Defile*).
  - `targeting`: Optional targeting declaration (`TargetRef`) if the trigger requires player target selection.
  - `effect`: Execution payload (`EffectDef`, `FR-ABL-007`) specifying an atomic primitive (`FR-PRM-001`) or composite control-flow container (`COMPOSITE_SEQUENCE`, `CONDITIONAL_CLAUSE`).
  - `maxPerTurn`: Optional integer (`>= 1`) restricting maximum trigger frequency per turn (`FR-ABL-008`).
- **Upfront Target Binding for Conditional Resolution**: For composite triggers where an effect's magnitude or occurrence depends on an action resolved during execution (*Samsara, Living Tempest*, *Infested Apostle*), all external targets are designated and locked upfront when the trigger is enqueued onto the stack (`FR-TGT-101`). If the condition or optional cost is not met during resolution, the conditional clause simply does not execute against the pre-locked target.

#### FR-ABL-003: ActivatedAbilityDef Schema & Cost Models
In compliance with **FR-ACT-001**, **FR-ACT-002**, and **FR-QUE-001**:
- Declared manually by the active player during the Main Phase while `Stack.isEmpty() == true`.
- **Schema Properties**: `costs` (`Array<CostDef>` paid simultaneously upon declaration; `FR-ABL-007`, `FR-ABL-009`), `targeting` (`TargetRef`; `FR-TGT-001`, `FR-TGT-002`), `effect` (`EffectDef`), and `repeatable` (boolean, default `true` unless constrained by costs or `maxPerTurn`).

#### FR-ABL-004: ContinuousAuraDef Schema & 6-Layer Interfacing
In compliance with **Section 6.2 of [`01_Game_Rules_FRD.md`](01_Game_Rules_FRD.md) (`FR-LAY-001` - `FR-LAY-003`)**, **FR-CST-001**, and **FR-TRN-003a**:
- Evaluates continuously across board assets (or in `HAND` / `DISCARD` when `activeZone` is explicitly declared), registering static modifiers into Layer 4 (Auras) of the 6-layer attribute pipeline or granting static zone deployment permissions.
- **Schema Properties**:
  - `activeZone`: Optional string enum (`"BATTLEFIELD"` default, `"HAND"`, or `"DISCARD"`). Specifies the zone in which the host card's continuous effect is active (e.g., `"HAND"` for self cost-reduction modifiers on *Swirling Dustdevil* and *Boneyard Amalgam*; `"DISCARD"` for alternate-zone deployment permission on *Veil Render*).
  - `scope`: `SELF`, `ALLIED_IN_ZONE`, `OPPOSING_IN_ZONE`, or `ALL_IN_ZONE`.
  - `filter`: Optional `FilterDef` narrowing which entities within `scope` receive the modifier (e.g., `{ "excludeSelf": true }` on *Animoceps Stalk*, or `{ "subtypes": ["Horror"] }` on *Hulking Incongruity*).
  - `modifiers`: Array of `AttributeModifierDef` specifying target attribute (`COMBAT_POWER`, `MAX_HEALTH`, `REQUISITION_COST`, `TACTICS_COST`, `KEYWORD_GRANT`, `SUBTYPE_GRANT`), mathematical operation (`ADD`, `SUBTRACT`, `SET`), and value expression (`ValueToken`). Dynamic cost reductions (`REQUISITION_COST`, `TACTICS_COST`) obey a minimum floor of `0` (`FR-CST-001`).
  - `abilityGrants`: Optional `Array<AbilityDef>` allowing continuous auras to project compound abilities (e.g. `Cornered: Resistant 1` on *Ancestral Monument*) onto scoped assets.
  - `staticPermission`: Optional `{ "permissionType": "ALLOW_DEPLOY_FROM_ZONE", "sourceZone": "DISCARD", "additionalCosts": Array<CostDef> }` granting permission to deploy the host card from a non-hand zone by paying additional costs (`FR-TRN-003a`, *Veil Render*).

> [!NOTE] Design Rationale (DA-DAT-004)
> Armor is an ephemeral consumable damage buffer (absorbed 1-for-1 before health deduction in Stage 4 of `FR-REP-004`) modified strictly via one-shot discrete events. A continuous aura granting "+1 Armor" is semantically equivalent to `Resistant 1` (`FR-KWD-002b`) and would otherwise create undefined state regeneration loops; defensive continuous auras must grant stat modifiers or damage reduction keywords (`Resistant N`, `Stalwart`).

#### FR-ABL-005: ReplacementWatcherDef Schema & Event Interception
In compliance with **FR-REP-001** through **FR-REP-004**:
- Intercepts prospective state events inline (`WOULD_TAKE_DAMAGE`, `WOULD_BE_DESTROYED`, `WOULD_DEPLOY`) without utilizing the Resolution Stack.
- **Schema Properties**: `interceptHook` (event hook to intercept), `condition` (`ConditionDef`), and `replacementAction` (`EffectDef` substituting the prospective event payload, e.g., `REDIRECT_DAMAGE_TARGET`, `SET_INITIAL_READINESS`, `CANCEL_AND_BOUNCE`).

#### FR-ABL-006: ManeuverEffectDef Schema & Stack Resolution
In compliance with **FR-MNV-001** and **FR-TRN-003**:
- Represents the core spell payload executed when a Maneuver card resolves from the top of the Resolution Stack.
- **Schema Properties**: `targeting` (`TargetRef` or `Array<TargetRef>`), `effect` (`EffectDef`), and modal payloads (`modalCount`, `options`; `FR-ABL-010`).

### Section 3.2: Declarative AST Sub-Schemas

#### FR-ABL-007: Common Declarative AST Types
All `AbilityDef` nodes compose the following six canonical sub-schemas (validated by `AbilityDef.schema.json` per `FR-CMP-022`):

| Sub-Schema | Required | Properties & Enum Constraints |
| :--- | :--- | :--- |
| **`FilterDef`** | *none* | `allegiance` (`ALLIED`, `OPPOSING`, `ANY`), `entityType` (`UNIT`, `STRUCTURE`, `HEADQUARTERS`, `ASSET`, `CARD`), `cardType` (`UNIT`, `STRUCTURE`, `MANEUVER`, `HEADQUARTERS`), `subtypes` (`Array<String>`), `negateSubtypes` (`Boolean`), `excludeSelf` / `excludeScratch` / `excludeHQ` / `excludeUnits` (`Boolean`), `readiness` (`READIED`, `EXHAUSTED`, `ANY`), `zone` (`SAME_ZONE`, `RESERVES`, `MAIN_THEATER`, `HEADQUARTERS`, `DISCARD`, `HAND`, `DECK`), `statComparison` (`{ property: [COMBAT_POWER, CURRENT_HEALTH, MAX_HEALTH, REQUISITION_COST, TACTICS_COST], operator: [EQ, NEQ, LT, LTE, GT, GTE], value: ValueToken }`). |
| **`ValueToken`** | `type` | `type` (`STATIC`, `DYNAMIC`, `QUERY`, `VARIABLE_X`), `value` (`Integer`), `source` (`SELF`, `TARGET`, `GAME_STATE`, `CAUSALITY_CONTEXT`), `property` (`COMBAT_POWER`, `CURRENT_HEALTH`, `MAX_HEALTH`, `BASE_REQUISITION_COST`, `BASE_TACTICS_COST`, `KILLED_COUNT`, `DISCARDED_COUNT`), `query` (`{ zone: [ATTACHED, DISCARD, BATTLEFIELD, HAND, RESERVES, MAIN_THEATER], filter: FilterDef, aggregate: [COUNT, SUM_ATTRIBUTE], targetAttribute: [REQUISITION_YIELD, TACTICS_YIELD, COMBAT_POWER, HEALTH] }`), `filter` (`FilterDef`), `multiplier` (`Integer`), `divisor` (`Integer >= 1`). |
| **`CostDef`** | `costType` | `costType` (`RESOURCE`, `EXHAUST_SELF`, `EXHAUST_ASSET`, `KILL_ASSET`, `DISCARD_CARD`, `SACRIFICE_ALLY`, `MOVE_CARD_ZONE`), `resource` (`REQUISITION`, `TACTICS`), `amount` (`Integer >= 1`), `filter` (`FilterDef`), `destinationPosition` (`TOP`, `BOTTOM`, `SHUFFLED_IN`). |
| **`TargetRef`** | `selector` | `selector` (`SELF`, `CHOSEN`, `ALL`, `DISTRIBUTED_POOL`, `RANDOM`, `BOUND_TARGET`, `TRIGGER_EVENT_ENTITY`), `filter` (`FilterDef`), `count` (`Integer >= 1`), `minCount` (`Integer >= 0`), `maxCount` (`Integer >= 1`), `allowedCounts` (`Array<Integer >= 1>`), `poolValue` (`Integer >= 1`), `targetIndex` (`Integer >= 0`), `bindingPolicy` (`UPFRONT_LOCK`, `ON_RESOLUTION`). |
| **`ConditionDef`** | *none* | `source` (`SELF`, `ALLIED`, `OPPOSING`, `ANY`), `turnOwner` (`SELF`, `OPPONENT`, `ANY`), `cornered` (`Boolean`), `deployedZone` (`RESERVES`, `MAIN_THEATER`), `fromZone` / `toZone` (`HAND`, `DECK`, `BATTLEFIELD`, `DISCARD`, `RESERVES`, `MAIN_THEATER`), `targetRef` (`String`), `handEmpty` (`Boolean`), plus any `FilterDef` predicate (`allegiance`, `entityType`, `cardType`, `zone`, `filter`). |
| **`EffectDef`** | `primitive` | `primitive` (`String` matching `FR-PRM-001` or `COMPOSITE_SEQUENCE` / `CONDITIONAL_CLAUSE`), `target` (`TargetRef`), `steps` (`Array<EffectDef>`), `condition` (`ConditionDef`), `optionalCost` (`CostDef`), `effect` / `ifTrue` / `ifFalse` (`EffectDef`), plus primitive-specific parameters (`FR-PRM-001`). |

> [!NOTE] Design Rationale (DA-DAT-002)
> Replaces ad-hoc, card-specific property enums (`COUNT_TUCKED_CARDS`, `COUNT_DISCARD_CARDS`, `COUNT_STRUCTURES_REQUISITION_YIELD`) with a generic, composable `query` object within `ValueToken`. This fulfills the declarative AST mandate (`DA-XP-001`), allowing designers to express entity counts and attribute aggregates without engine recompilation or schema enum expansions.

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
- `TriggeredAbilityDef` (`FR-ABL-002`) and `ActivatedAbilityDef` (`FR-ABL-003`) support `maxPerTurn: Integer` (e.g., *Animoceps Rager*: *"This ability triggers only once each turn"*). The engine increments `turnActivationCount` per `AssetInstance`, suppresses execution once `maxPerTurn` is reached, and resets counters during the Ready Phase (`FR-TRN-001`).

#### FR-ABL-009: Exhaust-Self as Ability Cost Model
- `CostDef` with `costType: "EXHAUST_SELF"` (`FR-ABL-007`) requires host `readiness == READIED`, transitions readiness to `EXHAUSTED` upon declaration (`EVENT_EXHAUST`), and may combine with resource payments (e.g., `[R1, EXHAUST_SELF]` on *Dolus Macrocannon*).

#### FR-ABL-010: Modal Selection AST Primitives
- Modal cards (*Rogue Fleshshaper*, *Ambush Tactics*) specify `modalCount: Integer` and `options: Array<EffectDef>` on `MANEUVER_EFFECT` or `TRIGGERED` nodes, validated against `selected_modal_indices` (`FR-MSG-003`).

#### FR-ABL-011: Distributed Pool Value Allocations
- In compliance with **FR-TGT-002** and **FR-MSG-003**, cards distributing a pool across targets (*Macrocannon Barrage*, *Hammerhand Bivouac*) specify `selector: "DISTRIBUTED_POOL"` with `poolValue: Integer`, requiring client `TargetBinding` entries whose `allocated_value` sum equals `poolValue`.

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

Complete JSON `CardBlueprint` reference implementations for the 14 benchmark cards are maintained in **[`reference/02_Benchmark_Card_Blueprints.md`](reference/02_Benchmark_Card_Blueprints.md)**:

| # | Benchmark Card (`id`) | Demonstrated AST Archetypes & Schema Features |
| :---: | :--- | :--- |
| **1** | `NO-AH-5` (`soulless_no_ah_5`) | `KEYWORD_GRANT` (`WARD`), `ACTIVATED` (`KILL_ASSET` + `TUCK_UNDER`), `CONTINUOUS_AURA` (`QUERY` over `ATTACHED` sub-zone; `FR-ENT-102`). |
| **2** | `Piston Fist Prototype` (`soulless_piston_fist_prototype`) | `KEYWORD_GRANT` (`AGGRESSIVE 9`) + post-attack `EVENT_ATTACK_RESOLVED` recoil (`MOVE_ZONE: RESERVES`; `FR-KWD-002f`). |
| **3** | `Animoceps Spore Cloud` (`animus_animoceps_spore_cloud`) | `REPLACEMENT_WATCHER` intercepting `WOULD_TAKE_DAMAGE` via `REDIRECT_DAMAGE_TARGET` (`FR-ABL-005`, `FR-REP-004`). |
| **4** | `Samsara, Living Tempest` (`concordia_samsara_living_tempest`) | `UPFRONT_LOCK` target binding (`FR-TGT-101`) with `LOOK_TOP_CARDS` partition tracking (`trackAs`) and `CONDITIONAL_CLAUSE`. |
| **5** | `Swirling Dustdevil` (`concordia_swirling_dustdevil`) | Hand-active `CONTINUOUS_AURA` (`activeZone: "HAND"`) reducing `REQUISITION_COST` via `DISCARD` `QUERY` (`FR-ABL-004`, `FR-CST-001`). |
| **6** | `Veil Render` (`animus_veil_render`) | Discard-active `CONTINUOUS_AURA` (`activeZone: "DISCARD"`) granting `ALLOW_DEPLOY_FROM_ZONE` with `KILL_ASSET` cost (`FR-ABL-004`, `FR-TRN-003a`). |
| **7** | `Rage's Wake` (`concordia_rages_wake`) | `COMPOSITE_SEQUENCE` combining `MODIFY_STAT` and `REGISTER_DELAYED_TRIGGER` (`EVENT_DESTROY` &rarr; `DEPLOY_FREE` to `RESERVES`). |
| **8** | `Funeral Fungus` (`animus_funeral_fungus`) | Root `additionalDeploymentCosts` (`EXHAUST_ASSET` on readied allied unit; `FR-DAT-005`) + discard `RETURN_TO_HAND`. |
| **9** | `Infested Apostle` (`animus_infested_apostle`) | `TRIGGERED` ability with `optional: true`, `optionalCosts` (`KILL_ASSET`), and `UPFRONT_LOCK` target binding (`FR-ABL-002`). |
| **10** | `Soul Whip` (`concordia_soul_whip`) | Multi-target `TargetRef` array with delegated `damageSource` (`targetIndex: 0`) and dynamic `COMBAT_POWER` scaling (`FR-PRM-001`). |
| **11** | `Allshan's Battle Trance` (`path_allshans_battle_trance`) | `DEPLOY_FREE` with forced `destinationZone: "MAIN_THEATER"`, `GRANT_KEYWORD` (`STALWART`), and `MODIFY_STAT`. |
| **12** | `Consuming Earth` (`concordia_consuming_earth`) | Dynamic `statComparison` filter (`COMBAT_POWER <= SUM_ATTRIBUTE(REQUISITION_YIELD)` of allied structures; `FR-ABL-007`). |
| **13** | `Tempest Runner` (`concordia_tempest_runner`) | `EXHAUST_SELF` activation cost (`FR-ABL-009`) with `excludeSelf: true` targeting filter (`FR-ABL-007`) and `REGROUP`. |
| **14** | `Mutation Swell` (`animus_mutation_swell`) | Discrete quantifier `allowedCounts: [2, 4]` and `CAUSALITY_CONTEXT` `KILLED_COUNT` scaling with `divisor: 2` (`FR-ABL-007`, `FR-TRG-007`). |

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
*(Subsumed by [`FR-CMP-001`](03_Card_Compiler_FRD.md) & [`FR-CMP-022`](03_Card_Compiler_FRD.md))* — `tools/linter/lint_cardtown.ps1` and compiler pre-flight stages enforce `FR-XXX-YYY` cross-reference integrity, RFC-4180 CSV syntax, and JSON Schema conformance.

#### FR-VAL-002: Rules Text vs Declarative Ability Parity Checker
- In compliance with [`FR-CMP-019`](03_Card_Compiler_FRD.md) and [`FR-CMP-020`](03_Card_Compiler_FRD.md), the compilation pipeline verifies 1-to-1 parity between printed `rulesText` and compiled `AbilityDef` AST entries (matching standalone keywords per `FR-ABL-012` and `ast_overrides.json` entries per `FR-CMP-017`).

#### FR-VAL-003: Uniform Dataset Validation Policy ("Must Fit All") & Review Card Compilability
- **Development vs. Production Gating ([`FR-CMP-005`](03_Card_Compiler_FRD.md), [`FR-CMP-021`](03_Card_Compiler_FRD.md))**: Cards flagged with `currentProxy == "true"` or `needsRework == "true"` in `data/cards/*.csv` compile with their boolean metadata flags intact during local development, while `-Strict` production builds reject any card carrying `currentProxy == "true"`, `needsRework == "true"`, or an unresolved review flag (`ERR_STRICT_PROTOTYPE_REJECTED`, `ERR_STRICT_UNRESOLVED_REVIEW`).

> [!NOTE] Design Rationale (DA-DAT-005)
> Harmonizes `FR-VAL-003` with `FR-CMP-005`: permitting prototype cards to compile with explicit metadata flags locally while blocking them under `-Strict` lets designers iterate on experimental mechanics without risking prototype leakage into release builds.

---

## DOMAIN 7: CARD AUDIT REGISTER & DESIGNER REVIEW LOG

### Section 7.1: Audit Summary & Canonical Design Rulings

All 167 cards across the 10 CSV balance sheets were audited and standardized to conform to target enums (`FR-DAT-006`) and MTG-aligned grammar (`FR-TXT-001` – `FR-TXT-006`). The complete 52-card historical text diff and standardization register is archived in **[`docs/Card_Standardization_Audit_Report.md`](../Card_Standardization_Audit_Report.md)**. Ongoing per-card review and parity status are tracked live via the CSV `needsRework` column (`FR-VAL-003`) and `data/schemas/ast_overrides.json` (`parityVerified`, `FR-CMP-020`).

> [!NOTE] Design Rationale (DA-DAT-003)
> Core Set mechanical triage rulings established during the 167-card standardization audit:
> - **Animoceps Rager**: Per-turn trigger frequency cap (`"This ability triggers only once each turn"`) maps to `maxPerTurn: 1` (`FR-ABL-008`).
> - **Funeral Fungus**: Unit exhaustion is templated as an `additionalDeploymentCosts` enactment cost (`EXHAUST_ASSET`; `FR-DAT-005`) rather than a resolution effect to prevent fizzle ambiguity.
> - **Rift**: Headquarters self-damage emits `EVENT_TAKE_DAMAGE` and routes through the 5-stage damage mitigation pipeline (`FR-REP-004`), triggering damage-reactive abilities (e.g., `Unbowed`) normally.
> - **Mutation Swell**: Target counts are restricted to discrete pairs (`allowedCounts: [2, 4]`, `divisor: 2`; `FR-ABL-007`) to eliminate odd-number floor ambiguity.
> - **Steel Rain & Macrocannon Barrage**: *Steel Rain* resolves as independent 2-damage targeted iterations per card discarded; *Macrocannon Barrage* uses `DISTRIBUTED_POOL` (`FR-ABL-011`).
> - **Title & Subtype Normalization**: Renamed *Daggerspine Rangers* to **Spoil Forest Irregulars** (resolving Warcraft IP collision), normalized plural subtype `"Specialists"` to `"Specialist"` (`FR-DAT-007`), and corrected **Reverberating Terror**, **Devilhorn Guerrillas**, and **All-Out War**.

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
