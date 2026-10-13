# Card Compiler & Blueprint Ingestion - Functional Requirements Document (FRD v1.0)
**Automated CSV Ingestion, Natural Language AST Synthesizer, Schema Validation & Cryptographic Fingerprinting Specification**

---

## EXECUTIVE SUMMARY & REVISION HISTORY

This Functional Requirements Document (FRD) establishes the authoritative technical, algorithmic, and architectural specification for the **Card Compiler Pipeline (`tools/importer/compile_blueprints.ps1`)**. The Card Compiler serves as the automated ingestion engine that transforms human-authored RFC-4180 balance spreadsheets into immutable, machine-executable JSON blueprint assets (`data/schemas/blueprints.json`).

The Card Compiler operates as the fourth core architectural pillar of the Animus Mundi specification suite, directly realizing the pipeline defined across:
1. **[Master Game Rules & Card Modeling FRD](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/Game_Rules_and_Card_Modeling_FRD.md) (Master FRD v3.0)** — Defines the entity taxonomy (`CardBlueprint`, `CardInstance`, `AssetInstance`), 6-layer attribute pipeline (`FR-LAY-001` - `FR-LAY-003`), combat lifecycle (`FR-CBT-002`), and keyword mechanics (`FR-KWD-002a` - `FR-KWD-002o`).
2. **[Client-Engine Protocol & Networking FRD](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/Client_Engine_Protocol_Networking_FRD.md) (FRD v1.1)** — Defines connection handshakes (`FR-SES-001`), server-side deck validation (`FR-SES-002`), public state checksum sync (`FR-PAC-004`), and replay header verification (`FR-SPC-002`).
3. **[Card Data Pipeline & Ability Authoring FRD](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/Card_Data_Pipeline_FRD.md) (FRD v1.0)** — Codifies the Tier 1 &rarr; Tier 2 &rarr; Tier 3 data authority model (`FR-DAT-001`), slug derivation (`FR-DAT-002`), ability AST definitions (`FR-ABL-001` - `FR-ABL-012`), and rules text grammar (`FR-TXT-001` - `FR-TXT-006`).

```text
┌────────────────────────────────────────────────────────────────────────┐
│              Tier 2: Version-Controlled Authoring CSVs                │
│                         (data/cards/*.csv)                             │
│   - 10 UTF-8 RFC-4180 CSV Balance Sheets                               │
│   - Direct Schema Header Parity (FR-DAT-006)                           │
│   - MTG-Aligned Rules Text (FR-TXT-001 - 006)                          │
└───────────────────────────────────┬────────────────────────────────────┘
                                    │
                      compile_blueprints.ps1 (FR-CMP-001)
                                    │
                                    ├─── [Stage 1-3] Atomic Multi-Sheet Ingestion
                                    ├─── [Stage 4]   Deterministic Slug Derivation
                                    ├─── [Stage 5-6] Natural Language AST Synthesis & Overrides
                                    ├─── [Stage 7]   Bidirectional Text-to-AST Parity Check
                                    ├─── [Stage 8]   JSON Schema Draft 2020-12 Validation
                                    └─── [Stage 9-10] Cryptographic Fingerprint & Atomic Emission
                                    │
                                    ▼
┌────────────────────────────────────────────────────────────────────────┐
│             Tier 3: Compiled Production Master Catalog                 │
│                     (data/schemas/blueprints.json)                     │
│   - Immutable Singleton Catalog (FR-ENT-001, FR-DAT-005)               │
│   - Full Declarative Ability AST (FR-ABL-001 - 012)                    │
│   - 64-Bit Cryptographic catalogChecksum (FR-DAT-004, FR-CMP-025)      │
│   - Semantic cardDataVersion (FR-DAT-004, FR-CMP-026)                  │
└───────────────────────────────────┬────────────────────────────────────┘
                                    │
                  ┌─────────────────┴─────────────────┐
                  ▼                                   ▼
        ┌──────────────────┐                ┌──────────────────┐
        │  Server Engine   │                │   Godot Client   │
        │ - Authoritative  │                │ - Visual UI / Card│
        │   State Machine  │                │   Renderer       │
        │ - Stack Resolver │                │ - Local Tooltips │
        │ - Rule Enforcer  │                │ - Action Prompts │
        └──────────────────┘                └──────────────────┘
```

### Initial Specification Scope (v1.0):
1. **Automated CLI Command Interface (`FR-CMP-001` - `FR-CMP-003`)**: Standardized parameter contracts, atomic multi-stage execution lifecycle, and zero-corrupt staging rollback guarantee.
2. **Atomic Multi-Sheet Ingestion & Type Coercion (`FR-CMP-004` - `FR-CMP-007`)**: Strict RFC-4180 parsing, whitespace stripping, prototype/proxy row filtering, and target enum enforcement.
3. **Natural Language Rules Text Tokenizer & Grammar Parser (`FR-CMP-008` - `FR-CMP-012`)**: Lexical analysis decomposing MTG-aligned sentences into triggers, costs, targets, stat modifications, and conditions.
4. **Declarative Ability AST Synthesizer (`FR-CMP-013` - `FR-CMP-016`)**: Translation engine compiling parsed clauses into polymorphic `AbilityDef` AST structures (`FR-ABL-001` - `FR-ABL-012`).
5. **Two-Tier Hybrid Architecture & Sidecar Overrides Registry (`FR-CMP-017` - `FR-CMP-018`)**: Production-grade architecture combining automated text synthesis with an explicit sidecar overrides registry (`data/schemas/ast_overrides.json`) for complex multi-clause mechanics.
6. **Bidirectional Parity Verification Engine (`FR-CMP-019` - `FR-CMP-021`)**: Implementation of `FR-VAL-002` guaranteeing zero dangling printed rules and zero unexpressed AST abilities.
7. **JSON Schema Emission & Structural Validation (`FR-CMP-022` - `FR-CMP-023`)**: Validation against `CardBlueprint.schema.json` and `AbilityDef.schema.json` conforming to JSON Schema Draft 2020-12.
8. **Cryptographic Fingerprinting & Semantic Versioning (`FR-CMP-024` - `FR-CMP-026`)**: RFC 8785 canonical serialization, 64-bit truncated SHA-256 `catalogChecksum`, and semantic dataset versioning.
9. **Structured Error Diagnostics & CI/CD Tooling (`FR-CMP-027` - `FR-CMP-028`)**: Actionable error taxonomy (`ERR_*`), line/cell references, and standardized exit codes.

---

## ARCHITECTURAL DECISIONS FLAGGED FOR REVIEW

The following significant design and architectural decisions represent inferences drawn from the foundational requirements to establish a production-grade compilation pipeline. Each decision is flagged for product and engineering sign-off:

> [!IMPORTANT]
> **DECISION 1: Two-Tier Hybrid Architecture (Parser + Sidecar Overrides Registry) (`FR-CMP-017`)**
> - **Inference**: MTG-style natural language text covers 85%+ of card behaviors (keywords, stat buffs, standard costs, simple triggers). However, highly complex mechanics (*NO-AH-5* card tucking `FR-ENT-102`, *Samsara, Living Tempest* dual top-deck inspection and conditional maneuver damage, *Ambush Tactics* 4-option modal targeting `FR-ABL-010`) cannot be reliably parsed with naive regular expressions without introducing extreme parser fragility.
> - **Decision**: We adopt an industry-standard two-tier compiler architecture: (1) An automated grammar synthesizer parses standard patterns; (2) A sidecar overrides registry (`data/schemas/ast_overrides.json`) supplies authoritative declarative ASTs for bespoke cards. When an override is present, the compiler validates that the card exists, that the override conforms to the JSON schema, and that bidirectional parity (`FR-VAL-002`) holds against the printed text.
> - **Alternative Considered**: Pure 100% regex parsing (causes fragile edge-case bugs when designers adjust wording) vs 100% manual JSON authoring (burdens designers with raw JSON).

> [!IMPORTANT]
> **DECISION 2: 64-Bit Cryptographic Checksum via Truncated SHA-256 (`FR-CMP-025`)**
> - **Inference**: `FR-DAT-004` requires a 64-bit cryptographic checksum (`catalogChecksum`), but Master FRD v3.0 does not specify the exact hash algorithm or normalization standard.
> - **Decision**: We specify canonical JSON serialization (RFC 8785: alphabetically sorted keys, no formatting whitespace, UTF-8 encoded without BOM) hashed via standard SHA-256, taking the leading 16 hexadecimal characters (64 bits). This guarantees 100% platform-independent parity across PowerShell 5.1/7+, Godot 4.x (GDScript/C#), Python, and server daemons (Go/C#) without requiring external native hashing libraries.
> - **Alternative Considered**: xxHash64 (requires native compiled DLLs/extensions in PowerShell and Godot) or CRC64 (insufficient collision resistance).

> [!IMPORTANT]
> **DECISION 3: Indented Repository Output with Canonical Hashing (`FR-CMP-023`)**
> - **Inference**: Digital card game repositories benefit from human-readable git diffs, while runtime engines benefit from minified files.
> - **Decision**: The compiler outputs formatted, 2-space indented JSON to `data/schemas/blueprints.json` for clear git diff review during pull requests. The `catalogChecksum` is computed strictly over the canonical RFC 8785 bytes, ensuring that whitespace formatting in the committed file has zero impact on the cryptographic signature. An optional `-Minify` switch emits `blueprints.min.json` for engine distribution bundles.
> - **Alternative Considered**: Emitting only minified JSON to the repository (makes git diffs unreadable for review).

> [!IMPORTANT]
> **DECISION 4: Blank Row Sanitation vs Incomplete Prototype Policy (`FR-CMP-005`)**
> - **Inference**: `FR-VAL-003` enforces a strict "must fit all" validation policy. Designers occasionally leave trailing blank rows or mark in-development rows with `needsRework == TRUE` or `currentProxy == TRUE`.
> - **Decision**: The compiler completely ignores blank rows where all cells are empty or whitespace. Any row with a non-empty `name` must fully conform to all target enums and cost constraints. Cards flagged with `needsRework == TRUE` or `currentProxy == TRUE` compile normally into blueprints with `"proxy": true` and `"needsRework": true` metadata flags, allowing QA and playtest builds to filter or identify them. Under `-Strict` mode, compilation fails if unreviewed placeholder rows exist.
> - **Alternative Considered**: Aborting on any proxy row (blocks development testing of WIP cards).

> [!IMPORTANT]
> **DECISION 5: CLI Invocation Interface & Parameter Contract (`FR-CMP-001`)**
> - **Inference**: The compilation tool must integrate into developer loops, pre-commit git hooks, and headless CI/CD build pipelines.
> - **Decision**: The compiler CLI script `tools/importer/compile_blueprints.ps1` supports explicit parameters: `-SourceDir`, `-OutputDir`, `-SchemaDir`, `-ValidateOnly` (dry run), `-EmitSchemas`, `-Strict`, `-Minify`, and `-Version`.
> - **Alternative Considered**: Zero-argument hardcoded script (inflexible for testing alternative paths and CI workflows).

---

## DOMAIN 1: COMPILER CLI SPECIFICATION, PARAMETERS & EXECUTION LIFECYCLE

```text
                                CLI INVOCATION
                compile_blueprints.ps1 [-Flags & Paths]
                                      │
                                      ▼
                        ┌───────────────────────────┐
                        │ Stage 1: Environment Init │
                        └─────────────┬─────────────┘
                                      ▼
                        ┌───────────────────────────┐
                        │ Stage 2: Schema Discovery │
                        └─────────────┬─────────────┘
                                      ▼
                        ┌───────────────────────────┐
                        │ Stage 3: CSV Multi-Parse  │
                        └─────────────┬─────────────┘
                                      ▼
                        ┌───────────────────────────┐
                        │ Stage 4: Slug Derivation  │
                        └─────────────┬─────────────┘
                                      ▼
                        ┌───────────────────────────┐
                        │ Stage 5: Lexer & Parsing  │
                        └─────────────┬─────────────┘
                                      ▼
                        ┌───────────────────────────┐
                        │ Stage 6: AST Overrides    │
                        └─────────────┬─────────────┘
                                      ▼
                        ┌───────────────────────────┐
                        │ Stage 7: Parity Checker   │
                        └─────────────┬─────────────┘
                                      ▼
                        ┌───────────────────────────┐
                        │ Stage 8: Schema Validate  │
                        └─────────────┬─────────────┘
                                      ▼
                        ┌───────────────────────────┐
                        │ Stage 9: Hash Fingerprint │
                        └─────────────┬─────────────┘
                                      ▼
                        ┌───────────────────────────┐
                        │ Stage 10: Atomic Emission │
                        └───────────────────────────┘
```

### Section 1.1: CLI Invocation Contract

#### FR-CMP-001: Compiler CLI Command Interface & Parameter Contract
The Card Compiler CLI script must be located at `tools/importer/compile_blueprints.ps1` and must support execution via PowerShell 5.1 and PowerShell 7+ on Windows, Linux, and macOS. The command interface conforms to the following parameter specification:

```powershell
param(
    [string]$SourceDir = "$PSScriptRoot\..\..\data\cards",
    [string]$OutputDir = "$PSScriptRoot\..\..\data\schemas",
    [string]$SchemaDir = "$PSScriptRoot\..\..\data\schemas",
    [string]$OverridesFile = "$PSScriptRoot\..\..\data\schemas\ast_overrides.json",
    [switch]$ValidateOnly,
    [switch]$EmitSchemas,
    [switch]$Strict,
    [switch]$Minify,
    [string]$Version = "1.0.0",
    [switch]$VerboseOutput
)
```

1. **`-SourceDir`**: Path to directory containing the 10 authoritative RFC-4180 balance CSV sheets (`FR-DAT-001`). Defaults to `<repoRoot>/data/cards`.
2. **`-OutputDir`**: Target directory where compiled JSON catalogs (`blueprints.json`, `blueprints.min.json`) are written. Defaults to `<repoRoot>/data/schemas`.
3. **`-SchemaDir`**: Directory housing canonical JSON schemas (`CardBlueprint.schema.json`, `AbilityDef.schema.json`). Defaults to `<repoRoot>/data/schemas`.
4. **`-OverridesFile`**: Path to the sidecar AST overrides registry (`FR-CMP-017`). Defaults to `<repoRoot>/data/schemas/ast_overrides.json`.
5. **`-ValidateOnly`**: Dry-run mode. Executes full ingestion, AST synthesis, parity checking, and schema validation without writing output files to disk. Returns exit code 0 on success.
6. **`-EmitSchemas`**: Emits or refreshes the standalone JSON Schema validation files (`CardBlueprint.schema.json`, `AbilityDef.schema.json`) into `-SchemaDir` during compilation.
7. **`-Strict`**: Enables strict validation mode. Treats all designer review items (`Section 7.2 of Card Data Pipeline FRD`), prototype proxies, and syntax warnings as fatal compilation errors.
8. **`-Minify`**: Emits an additional minified production artifact (`blueprints.min.json`) alongside the formatted `blueprints.json`.
9. **`-Version`**: Overrides the semantic `cardDataVersion` string embedded in the catalog header (`FR-DAT-004`). Must match Semantic Versioning 2.0 (e.g., `"1.0.0"`).
10. **`-VerboseOutput`**: Prints comprehensive per-card compilation logs, AST token streams, and parity reconciliation steps to standard output.

#### FR-CMP-002: Multi-Stage Compilation Pipeline & Lifecycle Invariants
The compilation process executes in 10 strictly sequential, deterministic stages. If any stage encounters a fatal error, execution halts immediately, and subsequent stages are suppressed:
- **Stage 1 (Environment Initialization)**: Validates directory paths, resolves absolute paths, and initializes execution timers.
- **Stage 2 (Schema Discovery & Loading)**: Loads canonical JSON schemas and the sidecar overrides registry (`FR-CMP-017`).
- **Stage 3 (Atomic Multi-Sheet Ingestion)**: Ingests all 10 CSV sheets atomically (`FR-CMP-004`), validating RFC-4180 parsing and header schemas (`FR-DAT-006`).
- **Stage 4 (Slug Derivation & Identity Checking)**: Computes canonical slug IDs (`FR-DAT-002`, `FR-CMP-006`) and verifies zero ID collisions.
- **Stage 5 (Rules Text Lexical Analysis & AST Synthesis)**: Decomposes rules text into syntactic clauses and generates declarative AST structures (`FR-CMP-008` - `FR-CMP-016`).
- **Stage 6 (Sidecar AST Overrides Binding)**: Binds bespoke AST definitions from the overrides registry for complex cards (`FR-CMP-017`).
- **Stage 7 (Bidirectional Parity Verification)**: Verifies 1-to-1 parity between printed rules text and compiled AST abilities (`FR-VAL-002`, `FR-CMP-019` - `FR-CMP-020`).
- **Stage 8 (Formal JSON Schema Validation)**: Validates all generated blueprint objects against `CardBlueprint.schema.json` and `AbilityDef.schema.json` (`FR-CMP-022`).
- **Stage 9 (Catalog Assembly & Fingerprinting)**: Assembles the master catalog, applies RFC 8785 canonical serialization, and computes the 64-bit `catalogChecksum` (`FR-DAT-004`, `FR-CMP-025`).
- **Stage 10 (Atomic Emission & Reporting)**: Writes output files via atomic file replacement (`FR-CMP-003`) and prints summary telemetry. Total compilation time for 200 cards must not exceed 2,000 milliseconds on standard hardware.

#### FR-CMP-003: Atomic Emission, File Staging & Rollback Guarantee
In fulfillment of **FR-DAT-003**:
1. The compiler must **never** write directly to the target output file (`blueprints.json`).
2. All compiled output is written to a temporary staging file located in the target directory: `<OutputDir>/blueprints.json.tmp`.
3. Only after Stage 9 completes with zero errors and 100% schema validation is the staging file atomically renamed or replaced over the production target file (`blueprints.json`).
4. If an error occurs during any compilation stage, the temporary staging file is deleted immediately, leaving the existing production `blueprints.json` completely unmodified. Partial compilation or corrupted catalog emissions are strictly impossible.

---

## DOMAIN 2: MULTI-SHEET INGESTION GRAMMAR & BLUEPRINT NORMALIZATION

### Section 2.1: CSV Parsing & Sanitation

#### FR-CMP-004: Atomic Multi-Sheet CSV Parsing & Encoding Contract
1. The compiler ingests the 10 authoritative CSV files defined in `FR-DAT-001`:
   - `maneuvers-animus.csv`, `maneuvers-concordia.csv`, `maneuvers-path.csv`, `maneuvers-soulless.csv`
   - `units-and-structures-animus.csv`, `units-and-structures-common.csv`, `units-and-structures-concordia.csv`
   - `units-and-structures-HQ.csv`, `units-and-structures-path.csv`, `units-and-structures-soulless.csv`
2. All CSV files must be encoded in **UTF-8** (RFC-4180 compliant). The parser must cleanly handle files with or without a UTF-8 Byte Order Mark (BOM).
3. The parser must normalize Windows (`CRLF`) and Unix (`LF`) line endings interchangeably.
4. Quoted fields containing escaped quotes (`""`) or embedded newlines (common in multi-ability rules text) must parse with full fidelity without splitting rows.
5. Ingestion is atomic: If any sheet is missing, unreadable, or corrupted, the compiler aborts immediately with exit code 1 (`ERR_CSV_IO`).

#### FR-CMP-005: Dataset Sanitation, Empty Rows & Prototype Filtering
1. **Empty Row Sanitation**: Trailing or intermediate rows where all columns are empty or contain only whitespace are cleanly ignored and excluded from the catalog.
2. **Mandatory Row Identification**: Any row possessing a non-empty `name` column is treated as a card declaration and must satisfy all schema requirements.
3. **Proxy & Rework Handling**:
   - The CSV columns `currentProxy` and `needsRework` are evaluated as boolean flags (`TRUE`/`FALSE`).
   - If `currentProxy` is `TRUE`, the compiled blueprint sets `"proxy": true` in its metadata block.
   - If `needsRework` is `TRUE`, the compiled blueprint sets `"needsRework": true` in its metadata block.
   - In standard mode, proxies compile normally to facilitate gameplay testing. Under `-Strict` mode, any card with `needsRework == TRUE` or `currentProxy == TRUE` triggers fatal compilation error `ERR_STRICT_PROTOTYPE_REJECTED`.

#### FR-CMP-006: Deterministic Blueprint ID Derivation & Slug Collision Invariant
The compiler executes the deterministic slug derivation algorithm codified in **FR-DAT-002**:
1. Extract card title (`name`) and faction (`faction`).
2. Strip all punctuation: remove apostrophes (`'`), commas (`,`), hyphens (`-`), periods (`.`), exclamation marks (`!`), question marks (`?`), and quotation marks (`"`).
3. Convert all uppercase letters to lowercase.
4. Replace whitespace sequences with a single underscore (`_`).
5. Prefix with faction slug:
   - For `NEUTRAL` cards: prefix with `neutral_` (e.g., `Scratch Company` &rarr; `neutral_scratch_company`).
   - For Headquarters cards: prefix with the faction slug (e.g., `Animus Rift` &rarr; `animus_rift`).
   - For all other faction cards: prefix with `<faction>_` (e.g., `CONCORDIA` + `Furnace Golem` &rarr; `concordia_furnace_golem`).
6. **Zero-Collision Invariant**: The compiler maintains a global lookup table of all derived slug IDs. If two rows produce identical IDs, compilation terminates immediately with `ERR_SLUG_COLLISION`, reporting both card titles and source files.

#### FR-CMP-007: Attribute Type Coercion & Range Clamping Invariants
In fulfillment of **FR-DAT-005** and **FR-DAT-006**, all CSV string values are coerced to strictly typed JSON fields:

| Field | Coercion Rule | Constraint / Invariant | Error Code on Failure |
| :--- | :--- | :--- | :--- |
| `name` | String | Non-empty, trimmed | `ERR_NAME_EMPTY` |
| `cardType` | String Enum | Strictly `UNIT`, `STRUCTURE`, `MANEUVER`, or `HEADQUARTERS` (`FR-DAT-006`) | `ERR_ENUM_INVALID_CARD_TYPE` |
| `faction` | String Enum | Strictly `CONCORDIA`, `SOULLESS`, `PATH`, `ANIMUS`, or `NEUTRAL` (`FR-DAT-006`) | `ERR_ENUM_INVALID_FACTION` |
| `rarity` | String Enum | Strictly `COMMON`, `UNCOMMON`, `RARE`, `MYTHIC`, or `HQ` (`FR-DAT-006`) | `ERR_ENUM_INVALID_RARITY` |
| `requisitionCost` | Integer | `>= 0` | `ERR_NUM_NEGATIVE_COST` |
| `tacticsCost` | Integer | `>= 0` | `ERR_NUM_NEGATIVE_COST` |
| `power` | Integer / Null | For Maneuvers: must be `null` or empty string.<br>For Units/Structures/HQs: integer `>= 0`. | `ERR_NUM_INVALID_POWER` |
| `health` | Integer / Null | For Maneuvers: must be `null` or empty string.<br>For Units/Structures/HQs: integer `>= 1`. | `ERR_NUM_INVALID_HEALTH` |
| `requisitionYield` | Integer | For Maneuvers: must be `0`.<br>For Assets: integer `>= 0`. | `ERR_NUM_NEGATIVE_YIELD` |
| `tacticsYield` | Integer | For Maneuvers: must be `0`.<br>For Assets: integer `>= 0`. | `ERR_NUM_NEGATIVE_YIELD` |
| `deploymentZones` | Array of Enums | Units/Structures: at least one of `RESERVES`, `MAIN_THEATER`.<br>HQs: strictly `["HEADQUARTERS"]`.<br>Maneuvers: strictly empty array `[]` (`FR-DAT-008`). | `ERR_ENUM_INVALID_ZONE` |
| `subtypes` | Array of Strings | Split on commas, trim whitespace. If value is `"--"`, emits empty array `[]` (`FR-DAT-007`). | `ERR_SUBTYPES_INVALID` |
| `rulesText` | String | Trimmed. If empty, emits empty string `""`. | N/A |

---

## DOMAIN 3: NATURAL LANGUAGE RULES TEXT PARSER & TOKENIZER

### Section 3.1: Lexical Analysis & Clause Segmentation

#### FR-CMP-008: Rules Text Lexer & Clause Segmentation
The rules text tokenizer breaks a card's `rulesText` string into discrete syntactic clauses:
1. **Reminder Text Stripping**: In fulfillment of `FR-TXT-001`, any parenthetical reminder text (e.g., `"(cannot be chosen by opponents except to attack)"`, `"(all combat damage taken is reduced to 1)"`) is stripped before tokenization.
2. **Line & Sentence Splitting**:
   - Newline characters (`\n`) define primary clause boundaries.
   - Within each line, sentences ending with a period (`.`) are partitioned into sequential syntactic clauses.
   - Bulleted lines (e.g., `Choose one: - Mode 1 - Mode 2`) are recognized as modal branching structures (`FR-ABL-010`, `FR-CMP-016`).
3. **Keyword Line Detection**: Lines consisting solely of a canonical keyword or keyword with numeric rank (e.g., `Stalwart`, `Aggressive 2`, `Deploys Exhausted`) are classified as standalone keyword declarations (`FR-CMP-013`).

#### FR-CMP-009: Trigger Preamble Parser & Temporal Hook Mapping
In fulfillment of **FR-TXT-002** and **FR-TRG-001**, clauses beginning with trigger preambles are parsed into `TriggeredAbilityDef` hook bindings:
1. **`When` Clauses (One-Shot Triggers)**:
   - Pattern: `^When this (unit|structure|Headquarters) deploys,?\s*(.+)$` &rarr; Binds hook `EVENT_DEPLOY`.
   - Pattern: `^When this (unit|structure) is killed,?\s*(.+)$` &rarr; Binds hook `EVENT_DESTROY`.
   - Pattern: `^When this card is put into your discard pile from your hand,?\s*(.+)$` &rarr; Binds hook `EVENT_ZONE_CHANGE` with filter `from: HAND, to: DISCARD`.
2. **`Whenever` Clauses (Recurring Triggers)**:
   - Pattern: `^Whenever this unit attacks,?\s*(.+)$` &rarr; Binds hook `EVENT_ATTACK_DECLARE`.
   - Pattern: `^Whenever an opposing unit attacks,?\s*(.+)$` &rarr; Binds hook `EVENT_ATTACK_DECLARE` with filter `allegiance: OPPOSING`.
   - Pattern: `^Whenever this (unit|structure|Headquarters) is dealt damage,?\s*(.+)$` &rarr; Binds hook `EVENT_TAKE_DAMAGE`.
   - Pattern: `^Whenever this unit exhausts,?\s*(.+)$` &rarr; Binds hook `EVENT_EXHAUST`.
3. **`At` Clauses (Phase / Turn Boundary Triggers)**:
   - Pattern: `^At the start of your turn,?\s*(.+)$` &rarr; Binds hook `EVENT_PHASE_START` with condition `turnOwner: SELF, phase: READY`.
   - Pattern: `^At the end of your turn,?\s*(.+)$` &rarr; Binds hook `EVENT_PHASE_START` with condition `turnOwner: SELF, phase: END`.
4. **Per-Turn Execution Limit Parser**:
   - If a sentence contains `"This ability triggers only once each turn."` or `"This effect happens once per turn."`, the parser binds property `maxPerTurn: 1` on the enclosing `TriggeredAbilityDef` (`FR-ABL-008`).

#### FR-CMP-010: Ability Cost Expression Parser
In fulfillment of **FR-TXT-003**, **FR-ACT-001**, and **FR-ABL-009**, activated ability cost headers (`[Cost], [Cost]: [Effect].`) are parsed into an array of `CostDef` structures:
1. **Resource Payments**:
   - `R(\d+)` &rarr; `costType: "RESOURCE", resource: "REQUISITION", amount: N`.
   - `T(\d+)` &rarr; `costType: "RESOURCE", resource: "TACTICS", amount: N`.
2. **Self-Exhaustion**:
   - `exhaust` &rarr; `costType: "EXHAUST_SELF"` (`FR-ABL-009`).
3. **Sacrifice / Kill Costs**:
   - `kill this (unit|structure)` &rarr; `costType: "KILL_ASSET", filter: { selector: "SELF" }`.
   - `kill chosen allied (.+)` &rarr; `costType: "KILL_ASSET", filter: <ParsedFilter>`.
4. **Discard Costs**:
   - `discard a card` &rarr; `costType: "DISCARD_CARD", amount: 1`.
   - `discard (\d+) cards` &rarr; `costType: "DISCARD_CARD", amount: N`.
5. **Additional Enactment Costs (Maneuvers)**:
   - Sentences matching `^As an additional cost to enact this maneuver,\s*(.+)\.$` are compiled into an `additionalCosts` array attached to `ManeuverEffectDef`.

#### FR-CMP-011: Declarative Target Filter & Selector Synthesizer
In fulfillment of **FR-TXT-004** and **FR-TGT-001**, targeting phrases are translated into canonical `TargetRef` and `FilterDef` AST nodes:
1. **Target Verb**: The parser matches canonical targeting verbs (`chosen`, `each of up to N chosen`). Non-standard verbs like `target` trigger a warning under normal mode and fatal error `ERR_PARITY_VERB` under `-Strict`.
2. **Target Selectors**:
   - `chosen (.+)` &rarr; `selector: "CHOSEN", count: 1`.
   - `up to (\d+) chosen (.+)` &rarr; `selector: "CHOSEN", count: N, optional: true`.
   - `each of up to (\d+) chosen (.+)` &rarr; `selector: "CHOSEN", count: N`.
   - `distribute (\d+) (damage|armor) among up to (\d+) chosen (.+)` &rarr; `selector: "DISTRIBUTED_POOL", poolValue: N, count: M` (`FR-ABL-011`).
   - `all (.+)` &rarr; `selector: "ALL"`.
3. **Filter Conjunctions**:
   - `allied` &rarr; `allegiance: "ALLIED"`.
   - `opposing` &rarr; `allegiance: "OPPOSING"`.
   - `unit` &rarr; `entityType: "UNIT"`.
   - `structure` &rarr; `entityType: "STRUCTURE"`.
   - `asset` &rarr; `entityType: "ASSET"`.
   - `non-Scratch` &rarr; `excludeScratch: true`.
   - `non-Headquarters` / `non-HQ` &rarr; `excludeHQ: true`.
   - `in the Main Theater` &rarr; `zone: "MAIN_THEATER"`.
   - `in the same zone` &rarr; `zone: "SAME_ZONE"`.

#### FR-CMP-012: Numerical & Value Token Synthesizer
In fulfillment of **FR-ABL-007**, stat expressions and numbers are compiled into `ValueToken` structures:
1. **Static Integers**: Standard numerals (`1`, `2`, `5`) &rarr; `type: "STATIC", value: N`.
2. **Dynamic Game State Counters**:
   - `"number of cards tucked under it"` &rarr; `type: "DYNAMIC", source: "SELF", property: "COUNT_TUCKED_CARDS"`.
   - `"total Requisition yield of all structures you control"` &rarr; `type: "DYNAMIC", source: "GAME_STATE", property: "COUNT_STRUCTURES_REQUISITION_YIELD"`.
   - `"number of cards discarded this way"` &rarr; `type: "DYNAMIC", source: "CAUSALITY_CONTEXT", property: "COUNT_DISCARD_CARDS"`.
3. **Variable X**: Phrasings like `Scratch X, where X is...` compile to `type: "VARIABLE_X"` bound to the parsed causality expression.

---

## DOMAIN 4: DECLARATIVE ABILITY AST SYNTHESIS ENGINE

### Section 4.1: Polymorphic Ability AST Generation

#### FR-CMP-013: Canonical Keyword AST Synthesizer
The compiler translates keyword declarations into declarative AST nodes matching **FR-ABL-012**:

| Keyword Syntax | Emitted AST Archetype | Emitted Node Payload |
| :--- | :--- | :--- |
| `Unbowed` | `TRIGGERED` | `triggerHook: "EVENT_TAKE_DAMAGE"`, `condition: { property: "damageTaken", operator: "GT", value: 0 }` |
| `Resistant N` | `REPLACEMENT_WATCHER` | `interceptHook: "WOULD_TAKE_DAMAGE"`, `replacementAction: "STAGE_2_FLAT_REDUCTION"`, `amount: N` |
| `Stalwart` | `REPLACEMENT_WATCHER` | `interceptHook: "WOULD_TAKE_DAMAGE"`, `replacementAction: "STAGE_3_COMBAT_CAP"`, `cap: 1` |
| `Aggressive N` | `TRIGGERED` | `triggerHook: "EVENT_ATTACK_DECLARE"`, `effect: { primitive: "MODIFY_STAT", attribute: "COMBAT_POWER", value: N, duration: "UNTIL_END_OF_TURN" }` |
| `Ward` | `STATIC_RESTRICTION` | `restrictionType: "TARGETING_IMMUNITY"`, `scope: "OPPOSING_NON_COMBAT"` |
| `Ethereal` | `STATIC_RESTRICTION` | `restrictionType: "ATTACK_IMMUNITY"`, `scope: "ALL_ATTACKS"` |
| `Versatile` | `CAPABILITY_FLAG` | `capability: "VERSATILE_GENERATION"` |
| `Guardian` | `STATIC_RESTRICTION` | `restrictionType: "MANDATORY_COMBAT_TARGET"` |
| `Armor N` | `STATIC_ATTRIBUTE` | `attribute: "INITIAL_ARMOR"`, `value: N` |
| `Deploys Exhausted`| `REPLACEMENT_WATCHER` | `interceptHook: "WOULD_DEPLOY"`, `replacementAction: "SET_INITIAL_READINESS"`, `readiness: "EXHAUSTED"` |
| `Scratch N` | `ATOMIC_PRIMITIVE` | `primitive: "DEPLOY_SCRATCH"`, `count: N` |
| `Regroup N` | `ATOMIC_PRIMITIVE` | `primitive: "REGROUP"`, `count: N` |
| `Heal N` | `ATOMIC_PRIMITIVE` | `primitive: "HEAL"`, `amount: N` |
| `Berserk` | `STATIC_RESTRICTION` | `restrictionType: "MULTI_ATTACK_DECLARATION"`, `maxTargets: 2` |

- **Conditional Keywords (Cornered)**: Lines matching `Cornered: [Keyword/Buff]` are compiled with an enclosing `conditionGate: { type: "CORNERED" }` (`FR-KWD-002e`). For example:
  - `Cornered: Resistant 1` &rarr; `REPLACEMENT_WATCHER` with `conditionGate: { type: "CORNERED" }`, flat reduction 1.
  - `Cornered: This unit has +2 power.` &rarr; `CONTINUOUS_AURA` evaluated in Layer 4 with `conditionGate: { type: "CORNERED" }`, power `+2`.

#### FR-CMP-014: Triggered & Activated Ability AST Synthesizer
1. **Triggered Abilities (`FR-ABL-002`)**:
   - Assembled from parsed preamble (`FR-CMP-009`), optional condition gate, targeting ref (`FR-CMP-011`), and effect primitive (`FR-PRM-001`).
   - If an effect specifies upfront target selection (*Samsara, Living Tempest*), targets are flagged with `bindUpfront: true` in fulfillment of `FR-TGT-101`.
2. **Activated Abilities (`FR-ABL-003`)**:
   - Assembled from parsed cost array (`FR-CMP-010`), targeting declaration, and effect payload.
   - If costs include `EXHAUST_SELF`, sets `requiresReady: true`.

#### FR-CMP-015: Continuous Aura & Replacement Watcher Synthesizer
1. **Continuous Auras (`FR-ABL-004`)**:
   - Sentences using `"have"` (`FR-TXT-005`): e.g. `"Other allied assets in the same zone have +1 power."`
   - Compiles to `abilityType: "CONTINUOUS_AURA"`, `layer: 4`, `scope: "ALLIED_IN_ZONE"`, `excludeSelf: true`, `modifiers: [{ attribute: "COMBAT_POWER", op: "ADD", value: 1 }]`.
   - Ability Grants: E.g., `"Other allied assets in the same zone have Cornered: Resistant 1."` compiles to `abilityGrants: [<CompiledCorneredResistantAST>]`.
2. **Replacement Watchers (`FR-ABL-005`)**:
   - Sentences matching `^If (.+) would (.+), (.+) instead\.$` (`FR-REP-001`).
   - Compiles to `abilityType: "REPLACEMENT_WATCHER"` with specified `interceptHook` and replacement action mutation.

#### FR-CMP-016: Maneuver Spell Payload & Modal Choice Synthesizer
1. **Maneuver Effects (`FR-ABL-006`)**:
   - Maneuver cards compile their main effect text into `abilityType: "MANEUVER_EFFECT"`, executing on the Resolution Stack (`FR-MNV-001`, `FR-QUE-001`).
2. **Modal Selections (`FR-ABL-010`)**:
   - Preamble matching `^Choose (one|two):\s*$` parses subsequent bulleted lines into discrete effect payloads.
   - Sets `modalCount: 1` (or `2`) and populates the `options` array.

---

## DOMAIN 5: TWO-TIER AST ARCHITECTURE & SIDECAR OVERRIDES REGISTRY

### Section 5.1: Hybrid Architecture & Bespoke Mechanics Handling

#### FR-CMP-017: Sidecar AST Overrides Extension Contract
To guarantee 100% engine reliability without creating an unmaintainable natural language parser (see Architectural Decision 1), the compiler supports an authoritative sidecar overrides file at `data/schemas/ast_overrides.json`:
1. **Schema of Overrides File**:
   ```json
   {
     "$schema": "https://json-schema.org/draft/2020-12/schema",
     "overrides": {
       "<blueprint_id>": {
         "rationale": "Description of why this card requires an explicit AST definition",
         "abilities": [ ... AbilityDef structures ... ]
       }
     }
   }
   ```
2. **Binding Priority**:
   - For every card ingested from CSV, the compiler checks if its derived `id` exists in `ast_overrides.json`.
   - If present, the compiler binds the `abilities` array from `ast_overrides.json`, bypassing automated text synthesis for that card.
   - If absent, the compiler uses the automated natural language synthesizer (`FR-CMP-013` - `FR-CMP-016`).
3. **Override Integrity Checks**:
   - Every card referenced in `ast_overrides.json` **must** exist in at least one CSV sheet. Orphaned overrides trigger fatal error `ERR_ORPHAN_OVERRIDE`.
   - Every override's `abilities` payload is strictly validated against `AbilityDef.schema.json`. Invalid schemas trigger fatal error `ERR_OVERRIDE_SCHEMA_INVALID`.
   - Every override must pass bidirectional parity verification (`FR-CMP-019` - `FR-CMP-020`) against the card's printed `rulesText`.

#### FR-CMP-018: Bespoke Mechanics Integration & Validation
The sidecar overrides registry houses formal AST definitions for cards featuring multi-clause or specialized engine interactions, including:
1. **NO-AH-5 (`soulless_no_ah_5`)**: Sub-entity tucking mechanism (`FR-ENT-102`) and dynamic power scaling based on tucked card count.
2. **Samsara, Living Tempest (`concordia_samsara_living_tempest`)**: Multi-card inspection (`FR-MOD-001`), selective hand/discard routing, and conditional maneuver discard causality trigger.
3. **Ambush Tactics (`path_ambush_tactics`)**: 4-option modal choice list (`FR-ABL-010`) with distinct targeting roles (`FR-TGT-001`).
4. **Blind Defile (`path_blind_defile`)**: Off-turn reactive combat interrupt sacrificing itself and the attacker during Step 2 of combat (`FR-CBT-002`).
5. **Forced Mutation (`animus_forced_mutation`)**: Face-down proxy transformation into a Scratch Company token (`FR-ENT-103`).

---

## DOMAIN 6: BIDIRECTIONAL RULES-TO-AST PARITY VERIFICATION ENGINE

### Section 6.1: Parity Checker Implementation (`FR-VAL-002`)

#### FR-CMP-019: Forward Parity Verification (Text-to-AST)
The compiler inspects every card's printed `rulesText` and verifies that every semantic element corresponds to an active AST node:
1. **Keyword Coverage**: If `rulesText` contains a recognized keyword (e.g., `Stalwart`, `Ward`, `Aggressive 2`), the card's `abilities` array must contain the corresponding AST node.
2. **Trigger Coverage**: If `rulesText` contains a trigger preamble (`When`, `Whenever`, `At`), the card must declare a `TRIGGERED` ability with the matching `triggerHook`.
3. **Activated Ability Coverage**: If `rulesText` contains a colon (`:`), the card must declare an `ACTIVATED` ability with matching costs.
4. **Failure Condition**: Any printed rule clause that fails to map to an AST node triggers compilation error `ERR_PARITY_UNMAPPED_TEXT`.

#### FR-CMP-020: Reverse Parity Verification (AST-to-Text)
The compiler verifies that no blueprint possesses "zombie" or "ghost" abilities absent from its printed text:
1. Every `AbilityDef` in a blueprint's `abilities` array must correspond to a recognizable clause in `rulesText`.
2. Blueprints are prohibited from declaring hidden, silent, or unprinted capabilities (excluding universal inherent actions: March, Attack, Generate; `FR-INH-001` - `FR-INH-004`).
3. Failure Condition: Any unrepresented AST node triggers compilation error `ERR_PARITY_ORPHAN_AST`.

#### FR-CMP-021: Strict Mode Enforcement & Designer Review Flags
In fulfillment of **Domain 7 of Card Data Pipeline FRD**:
1. When invoked with `-Strict`, the compiler verifies that all 52 cards listed in the Master Review Register have resolved design approvals.
2. If a card is flagged with unresolved design questions and `-Strict` is active, compilation fails with `ERR_STRICT_UNRESOLVED_REVIEW`.
3. In standard mode, review warnings are printed to standard error as non-fatal diagnostic notes.

---

## DOMAIN 7: JSON SCHEMA GENERATION & PRODUCTION BUNDLE PACKAGING

### Section 7.1: Production Schemas & Catalog Generation

#### FR-CMP-022: Formal JSON Schema Generation & Conformance
The compiler validates every blueprint against canonical JSON schemas conforming to **JSON Schema Draft 2020-12**:
1. **`CardBlueprint.schema.json`**: Located at `data/schemas/CardBlueprint.schema.json`. Validates top-level card metadata, stats, costs, enums, and the `abilities` array.
2. **`AbilityDef.schema.json`**: Located at `data/schemas/AbilityDef.schema.json`. Validates polymorphic ability types, condition gates, cost arrays, filter criteria, and targeting selectors (`FR-ABL-001` - `FR-ABL-007`).
3. **`-EmitSchemas` Flag**: When executed with `-EmitSchemas`, the compiler regenerates and overwrites both schema files with authoritative schema definitions, guaranteeing synchronization between code and schemas.

#### FR-CMP-023: Production Catalog Emission
The compiled production bundle is emitted to `data/schemas/blueprints.json`:
1. **Envelope Structure**:
   ```json
   {
     "$schema": "./CardBlueprint.schema.json",
     "catalogChecksum": "0x4f8a1c9e2b3d7a01",
     "cardDataVersion": "1.0.0",
     "generatedAt": "2026-10-13T13:30:00Z",
     "totalCardCount": 167,
     "blueprints": [
       { ... CardBlueprint ... }
     ]
   }
   ```
2. **Formatting**: Formatted with 2-space indentation and newline terminators to provide clear git diffs in source control.
3. **Minification**: If `-Minify` is specified, an additional unformatted file `blueprints.min.json` is generated for distribution to production servers and clients.

---

## DOMAIN 8: CRYPTOGRAPHIC CATALOG FINGERPRINTING & VERSIONING

### Section 8.1: Deterministic Hashing & Replay Binding

#### FR-CMP-024: Canonical JSON Normalization & Deterministic Serialization
To guarantee that the catalog checksum is 100% reproducible across different platforms, shells, and operating systems:
1. The compiler canonicalizes the blueprint catalog in accordance with **RFC 8785 (JSON Canonicalization Scheme - JCS)**:
   - Object property keys are sorted lexicographically by UTF-16 code units.
   - All insignificant whitespace (spaces, tabs, newlines) outside string literals is omitted.
   - Character strings are normalized to UTF-8 without byte order marks (BOM).
   - Numeric values are formatted without trailing zeroes (e.g. `1` rather than `1.0`).
2. Catalog serialization is strictly independent of host operating system line endings (`CRLF` vs `LF`).

#### FR-CMP-025: 64-Bit Cryptographic Checksum Computation (`catalogChecksum`)
In fulfillment of **FR-DAT-004**, **FR-SES-001**, and **FR-SPC-002**:
1. The compiler computes the cryptographic hash of the RFC 8785 canonical bytes using **SHA-256**.
2. The resulting 256-bit hash is truncated to its leading 64 bits (first 16 hexadecimal characters).
3. The checksum is formatted as a 0x-prefixed 16-character hexadecimal string: e.g. `"0x4f8a1c9e2b3d7a01"`.
4. This checksum is embedded in the catalog envelope (`blueprints.json`), checked during server session handshakes (`FR-SES-001`), and recorded in historical match replay headers (`ReplayHeader`; `FR-SPC-002`).

#### FR-CMP-026: Semantic Dataset Versioning Contract (`cardDataVersion`)
1. The `cardDataVersion` string conforms to Semantic Versioning 2.0 (`MAJOR.MINOR.PATCH`):
   - **MAJOR**: Incremented when the `CardBlueprint` or `AbilityDef` schema introduces breaking structural changes (e.g., removing a field, changing enum values).
   - **MINOR**: Incremented for balance patches that alter card stats, text, costs, or add new cards.
   - **PATCH**: Incremented for typographical, grammatical, or tooling fixes that do not alter card gameplay mechanics.
2. Connection negotiation (`FR-SES-001`) verifies that client and server share identical `catalogChecksum` values regardless of minor version labels.

---

## DOMAIN 9: COMPILER DIAGNOSTICS, ERROR TAXONOMY & OPERATIONAL ERGONOMICS

### Section 9.1: Diagnostic Reporting & Error Codes

#### FR-CMP-027: Structured Diagnostic Reporting & Error Taxonomy
When compilation fails, the compiler prints a structured, actionable error message to standard error:

```text
================================================================================
  COMPILATION ERROR [ERR_ENUM_INVALID_FACTION]
--------------------------------------------------------------------------------
  Source File : data/cards/units-and-structures-animus.csv
  Row Number  : 14
  Card Name   : Spoil Forest Haunt
  Column Name : faction
  Found Value : "Animus"
  Requirement : Must strictly equal one of [CONCORDIA, SOULLESS, PATH, ANIMUS, NEUTRAL]
  Remediation : Update "Animus" to uppercase "ANIMUS" (FR-DAT-006).
================================================================================
```

The compiler implements the following standardized error taxonomy:

| Category | Error Code | Description |
| :--- | :--- | :--- |
| **I/O & Environment** | `ERR_CSV_IO` | Missing, unreadable, or locked CSV sheet. |
| | `ERR_SCHEMA_IO` | Unable to read or write JSON schema files. |
| | `ERR_OUTPUT_IO` | Unable to write staging or production JSON artifacts. |
| **Parsing & Formatting**| `ERR_CSV_MALFORMED` | Malformed CSV row, mismatched quote, or bad delimiter. |
| | `ERR_NAME_EMPTY` | Row declares missing or whitespace-only card title. |
| | `ERR_NUM_PARSE` | Non-integer value in integer column (`power`, `health`, `cost`). |
| **Enums & Schema** | `ERR_ENUM_INVALID_CARD_TYPE` | `cardType` does not match `UNIT`, `STRUCTURE`, `MANEUVER`, `HEADQUARTERS`. |
| | `ERR_ENUM_INVALID_FACTION` | `faction` does not match canonical uppercase target enum (`FR-DAT-006`). |
| | `ERR_ENUM_INVALID_RARITY` | `rarity` does not match `COMMON`, `UNCOMMON`, `RARE`, `MYTHIC`, `HQ`. |
| | `ERR_ENUM_INVALID_ZONE` | `deploymentZones` contains illegal zone identifiers. |
| **Identity & Slugs** | `ERR_SLUG_COLLISION` | Two cards derive identical slug IDs (`FR-DAT-002`, `FR-CMP-006`). |
| **AST Synthesis** | `ERR_SYNTAX_UNPARSEABLE` | Rules text sentence cannot be parsed by lexer or grammar engine. |
| | `ERR_OVERRIDE_SCHEMA_INVALID`| Sidecar override entry fails schema validation. |
| | `ERR_ORPHAN_OVERRIDE` | Override entry defined for non-existent blueprint slug. |
| **Parity Validation** | `ERR_PARITY_UNMAPPED_TEXT` | Printed rules clause has no corresponding AST node (`FR-VAL-002`). |
| | `ERR_PARITY_ORPHAN_AST` | AST ability node has no corresponding printed rules text. |
| | `ERR_PARITY_VERB` | Non-canonical targeting verb (e.g. `target` instead of `chosen`). |
| **Strict Mode** | `ERR_STRICT_PROTOTYPE_REJECTED`| Proxy or rework card encountered while `-Strict` is enabled. |
| | `ERR_STRICT_UNRESOLVED_REVIEW` | Flagged review item encountered while `-Strict` is enabled. |

#### FR-CMP-028: Non-Zero Exit Code Contract & CI/CD Integration
The compiler process terminates with standardized process exit codes to enable seamless integration into CI/CD pipelines and pre-commit hooks:
- **`0`**: Compilation succeeded completely; output artifact verified and emitted.
- **`1`**: Environment, file I/O, or command-line parameter error (`ERR_*_IO`).
- **`2`**: CSV syntax, column structure, or parsing failure (`ERR_CSV_*`).
- **`3`**: Enum, type, or constraint validation failure (`ERR_ENUM_*`, `ERR_NUM_*`).
- **`4`**: AST synthesis, overrides, or schema conformance failure (`ERR_SYNTAX_*`, `ERR_OVERRIDE_*`).
- **`5`**: Text-to-AST parity check failure or strict mode violation (`ERR_PARITY_*`, `ERR_STRICT_*`).

---

## DOMAIN 10: COMPLETE CROSS-SPEC TRACEABILITY MATRIX

### Section 10.1: Traceability & Requirements Index

| Requirement ID | Canonical Requirement Title | Cross-Reference & Traceability Dependencies |
| :--- | :--- | :--- |
| **FR-CMP-001** | Compiler CLI Command Interface & Parameters | `FR-DAT-001`, `FR-DAT-003`, `FR-VAL-001` |
| **FR-CMP-002** | Multi-Stage Compilation Pipeline & Invariants| `FR-DAT-001`, `FR-DAT-003`, `FR-VAL-001` |
| **FR-CMP-003** | Atomic Emission, File Staging & Rollback | `FR-DAT-003`, `FR-ENT-001` |
| **FR-CMP-004** | Atomic Multi-Sheet CSV Parsing & Encoding | `FR-DAT-001`, `FR-DAT-006` |
| **FR-CMP-005** | Dataset Sanitation, Empty Rows & Proxies | `FR-VAL-003`, `FR-SET-001` |
| **FR-CMP-006** | Deterministic Slug Derivation & Collision Invariant| `FR-DAT-002`, `FR-ENT-001`, `FR-SET-001` |
| **FR-CMP-007** | Attribute Type Coercion & Range Clamping | `FR-DAT-005`, `FR-DAT-006`, `FR-CRD-101` |
| **FR-CMP-008** | Rules Text Lexer & Clause Segmentation | `FR-TXT-001`, `FR-TXT-006`, `FR-VAL-002` |
| **FR-CMP-009** | Trigger Preamble Parser & Temporal Hook Mapping | `FR-TXT-002`, `FR-TRG-001`, `FR-TRG-007` |
| **FR-CMP-010** | Ability Cost Expression Parser | `FR-TXT-003`, `FR-ACT-001`, `FR-ABL-009` |
| **FR-CMP-011** | Declarative Target Filter & Selector Synthesizer | `FR-TXT-004`, `FR-TGT-001`, `FR-TGT-102` |
| **FR-CMP-012** | Numerical & Value Token Synthesizer | `FR-ABL-007`, `FR-TXT-006` |
| **FR-CMP-013** | Canonical Keyword AST Synthesizer | `FR-ABL-012`, `FR-KWD-002a` through `FR-KWD-002o` |
| **FR-CMP-014** | Triggered & Activated Ability AST Synthesizer| `FR-ABL-002`, `FR-ABL-003`, `FR-QUE-001` |
| **FR-CMP-015** | Continuous Aura & Replacement Watcher Synthesizer | `FR-ABL-004`, `FR-ABL-005`, `FR-LAY-001`, `FR-REP-001` |
| **FR-CMP-016** | Maneuver Spell Payload & Modal Choice Synthesizer | `FR-ABL-006`, `FR-ABL-010`, `FR-MNV-001` |
| **FR-CMP-017** | Sidecar AST Overrides Extension Contract | `FR-DAT-003`, `FR-ABL-001`, `FR-VAL-002` |
| **FR-CMP-018** | Bespoke Mechanics Integration & Validation | `FR-ENT-102`, `FR-ENT-103`, `FR-MOD-001`, `FR-CBT-002` |
| **FR-CMP-019** | Forward Parity Verification (Text-to-AST) | `FR-VAL-002`, `FR-TXT-001`, `FR-ABL-001` |
| **FR-CMP-020** | Reverse Parity Verification (AST-to-Text) | `FR-VAL-002`, `FR-ENT-001` |
| **FR-CMP-021** | Strict Mode Enforcement & Designer Review Flags| `FR-VAL-003`, Section 7.2 of Card Data Pipeline FRD |
| **FR-CMP-022** | Formal JSON Schema Generation & Conformance | `FR-DAT-005`, `FR-ABL-001`, `FR-VAL-001` |
| **FR-CMP-023** | Production Catalog Emission | `FR-DAT-003`, `FR-ENT-001` |
| **FR-CMP-024** | Canonical JSON Normalization (RFC 8785) | `FR-DAT-004`, `FR-PAC-004` |
| **FR-CMP-025** | 64-Bit Cryptographic Checksum (`catalogChecksum`)| `FR-DAT-004`, `FR-SES-001`, `FR-SPC-002` |
| **FR-CMP-026** | Semantic Dataset Versioning (`cardDataVersion`)| `FR-DAT-004`, `FR-SES-001` |
| **FR-CMP-027** | Structured Diagnostic Reporting & Error Taxonomy | `FR-VAL-001`, `FR-ERR-001` |
| **FR-CMP-028** | Non-Zero Exit Code Contract & CI/CD Integration | `FR-VAL-001`, `FR-NET-001` |

---
*End of Functional Requirements Document (Card Compiler FRD v1.0)*
