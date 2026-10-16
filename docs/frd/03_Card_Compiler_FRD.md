# Card Compiler & Blueprint Ingestion - Functional Requirements Document (FRD v1.0)
**Automated CSV Ingestion, Natural Language AST Synthesizer, Schema Validation & Cryptographic Fingerprinting Specification**

---

## EXECUTIVE SUMMARY & REVISION HISTORY

This Functional Requirements Document (FRD) establishes the authoritative technical, algorithmic, and architectural specification for the **Card Compiler Pipeline (`tools/importer/compile_blueprints.ps1`)**. The Card Compiler serves as the automated ingestion engine that transforms human-authored RFC-4180 balance spreadsheets into immutable, machine-executable JSON blueprint assets (`data/schemas/blueprints.json`).

The Card Compiler operates as the fourth core architectural pillar of the Animus Mundi specification suite, directly realizing the pipeline defined across:
1. **[Master Game Rules & Card Modeling FRD](01_Game_Rules_and_Card_Modeling_FRD.md) (Master FRD v3.0)** — Defines the entity taxonomy (`CardBlueprint`, `CardInstance`, `AssetInstance`), 6-layer attribute pipeline (`FR-LAY-001` - `FR-LAY-003`), combat lifecycle (`FR-CBT-002`), and keyword mechanics (`FR-KWD-002a` - `FR-KWD-002o`).
2. **[Client-Engine Protocol & Networking FRD](04_Client_Engine_Protocol_Networking_FRD.md) (FRD v1.2)** — Defines connection handshakes (`FR-SES-001`), server-side deck validation (`FR-SES-002`), public state checksum sync (`FR-PAC-004`), and replay header verification (`FR-SPC-002`).
3. **[Card Data Pipeline & Ability Authoring FRD](02_Card_Data_Pipeline_FRD.md) (FRD v1.0)** — Codifies the Tier 1 &rarr; Tier 2 &rarr; Tier 3 data authority model (`FR-DAT-001`), slug derivation (`FR-DAT-002`), ability AST definitions (`FR-ABL-001` - `FR-ABL-012`), and rules text grammar (`FR-TXT-001` - `FR-TXT-006`).

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
│   - 256-Bit Cryptographic catalogChecksum (FR-DAT-004, FR-CMP-025)     │
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
8. **Cryptographic Fingerprinting & Semantic Versioning (`FR-CMP-024` - `FR-CMP-026`)**: RFC 8785 canonical serialization, 256-bit SHA-256 `catalogChecksum`, and semantic dataset versioning.
9. **Structured Error Diagnostics & CI/CD Tooling (`FR-CMP-027` - `FR-CMP-028`)**: Actionable error taxonomy (`ERR_*`), line/cell references, and standardized exit codes.

---

## ARCHITECTURAL DECISIONS FLAGGED FOR REVIEW

The following significant design and architectural decisions represent inferences drawn from the foundational requirements to establish a production-grade compilation pipeline. Each decision is flagged for product and engineering sign-off:

> [!IMPORTANT]
> **DECISION 1: Decoupled Declarative Catalog Model & Sidecar Overrides Registry (`FR-CMP-017`)**
> - **Inference**: Attempting to implement a full natural language processing (NLP) compiler across arbitrary English rules text in PowerShell regular expressions is fragile and introduces high operational risk (`DA-CMP-002`). Minor phrasing adjustments, modal choices, or composite triggers inevitably break regex heuristics or cause silent miscompilations.
> - **Decision**: We adopt the **Decoupled Declarative Catalog Model (Direction B)**:
>   1. **CSV Authority**: RFC-4180 balance CSVs (`data/cards/*.csv`) authoritatively own card metadata, base numerical stats, resource costs, combat zones, and the human-readable printed `rulesText` displayed on client UI tooltips and card frames.
>   2. **Declarative Ability AST Authority**: Authoritative, executable ability logic lives in declarative JSON AST format (`data/schemas/ast_overrides.json` or card JSON catalogs).
>   3. **Retired Complex Regex NLP**: Automated text synthesis in `compile_blueprints.ps1` is strictly scoped to canonical keywords (`FR-CMP-013`), standard cost headers (`FR-CMP-010`), and simple single-clause triggers (`FR-CMP-009`). Any card with composite logic, modal branches, dynamic values, or non-trivial mechanics bypasses regex parsing and is defined directly in declarative JSON.
>   4. **AI Card Compiler Assistant**: An AI card compiler authoring persona guides designers from natural language rules text intent into verified, schema-compliant JSON AST definitions.
>   5. **Reverse Parity Scoping (`FR-CMP-020`)**: Overridden cards are exempt from automated reverse AST-to-text NLP mapping. Parity is enforced strictly via explicit designer sign-off (`"parityVerified": true`, `verifiedBy`, `verifiedAt`) and exact sanitized CSV rulesText matching (`FR-CMP-017`, `FR-CMP-019`).
> - **Alternative Considered**: Full regex NLP parser in PowerShell (rejected due to fragility on context-sensitive text and high maintenance burden) vs 100% raw JSON authoring without CSV balance sheets (rejected due to designer workflow friction).

> [!IMPORTANT]
> **DECISION 2: 256-Bit Cryptographic Checksum via Full SHA-256 (`FR-CMP-025`)**
> - **Inference**: `FR-DAT-004` requires a cryptographic checksum (`catalogChecksum`), but Master FRD v3.0 does not specify the exact hash algorithm or normalization standard.
> - **Decision**: We specify canonical JSON serialization (RFC 8785: alphabetically sorted keys, no formatting whitespace, UTF-8 encoded without BOM) hashed via standard SHA-256, retaining the full 256-bit hexadecimal string (64 characters prefixed with `0x`). This guarantees 100% platform-independent parity across PowerShell 5.1/7+, Godot 4.x (GDScript/C#), Python, and server daemons (Go/C#) while eliminating birthday collision risks as the card catalog expands.
> - **Alternative Considered**: Truncating to 64 bits (discards 75% of entropy with negligible storage savings) or xxHash64 (requires native compiled DLLs/extensions).

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
3. **`-SchemaDir`**: Directory housing canonical JSON schemas (`CardBlueprint.schema.json`, `AbilityDef.schema.json`, `ast_overrides.schema.json`). Defaults to `<repoRoot>/data/schemas`.
4. **`-OverridesFile`**: Path to the sidecar AST overrides registry (`FR-CMP-017`). Defaults to `<repoRoot>/data/schemas/ast_overrides.json`.
5. **`-ValidateOnly`**: Dry-run mode. Executes full ingestion, AST synthesis, parity checking, and schema validation without writing output files to disk. Returns exit code 0 on success.
6. **`-EmitSchemas`**: Emits or refreshes the standalone JSON Schema validation files (`CardBlueprint.schema.json`, `AbilityDef.schema.json`, `ast_overrides.schema.json`) into `-SchemaDir` during compilation (`FR-CMP-022`).
7. **`-Strict`**: Enables strict validation mode. Treats all designer review items (`Section 7.2 of Card Data Pipeline FRD`), prototype proxies, and syntax warnings as fatal compilation errors.
8. **`-Minify`**: Emits an additional minified production artifact (`blueprints.min.json`) alongside the formatted `blueprints.json`.
9. **`-Version`**: Overrides the semantic `cardDataVersion` string embedded in the catalog header (`FR-DAT-004`). Must match Semantic Versioning 2.0 (e.g., `"1.0.0"`).
10. **`-VerboseOutput`**: Prints comprehensive per-card compilation logs, AST token streams, and parity reconciliation steps to standard output.

##### Process Exit Code Contract (Subsumes FR-CMP-028)
The compiler terminates with standardized process exit codes to enable seamless integration into CI/CD pipelines and pre-commit hooks:

| Exit Code | Classification | Trigger Scenarios & Error Taxonomy (`FR-CMP-027`) |
| :---: | :--- | :--- |
| **`0`** | **Success** | Compilation succeeded completely; output artifact verified and emitted. |
| **`1`** | **Environment / I/O** | Missing, unreadable, or locked CSV sheet or schema file (`ERR_*_IO`). |
| **`2`** | **CSV Syntax / Parsing** | Malformed CSV row, unclosed quote, or header mismatch (`ERR_CSV_*`, `ERR_NAME_EMPTY`). |
| **`3`** | **Type / Enum Constraint** | Illegal enum value, negative cost, or type coercion failure (`ERR_ENUM_*`, `ERR_NUM_*`). |
| **`4`** | **AST Synthesis / Schema** | Override schema invalid, override text mismatch, unverified override (`ERR_SYNTAX_*`, `ERR_OVERRIDE_*`). |
| **`5`** | **Parity / Strict Gate** | Text-to-AST parity failure, proxy/rework rejected under strict (`ERR_PARITY_*`, `ERR_STRICT_*`). |


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
- **Stage 9 (Catalog Assembly & Fingerprinting)**: Assembles the master catalog, applies RFC 8785 canonical serialization, and computes the 256-bit `catalogChecksum` (`FR-DAT-004`, `FR-CMP-025`).
- **Stage 10 (Atomic Emission & Reporting)**: Writes output files via atomic file replacement (`FR-CMP-003`) and prints summary telemetry. Total compilation time for 200 cards must not exceed 5,000 milliseconds in PowerShell 7+ (< 10,000 ms in Windows PowerShell 5.1) on standard hardware for warm execution runs.

#### FR-CMP-003: Atomic Emission, File Staging & Rollback Guarantee
In fulfillment of **FR-DAT-003**:
1. The compiler must **never** write directly to the target output file (`blueprints.json`).
2. All compiled output is written to a process-unique temporary staging file located in the target directory: `<OutputDir>/blueprints.json.tmp.$PID`.
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
3. **Harmonized Prototype & Rework Policy (FR-VAL-003 Alignment)**:
   - The CSV columns `currentProxy` and `needsRework` are evaluated as boolean flags (`TRUE`/`FALSE`).
   - If `currentProxy` is `TRUE`, the compiled blueprint sets `"proxy": true` in its metadata block.
   - If `needsRework` is `TRUE`, the compiled blueprint sets `"needsRework": true` in its metadata block.
   - **Development & Playtest Mode**: To support rapid gameplay iteration and playtesting, prototype or rework cards compile cleanly into blueprints with these metadata flags intact, enabling game clients and simulation harnesses to filter or flag them.
   - **Strict Mode Production Gate**: Under `-Strict` mode (enforced in CI/CD release builds and pre-commit production gates), any card marked with `needsRework == TRUE`, `currentProxy == TRUE`, or containing incomplete placeholder rules text triggers fatal compilation error `ERR_STRICT_PROTOTYPE_REJECTED`, guaranteeing zero prototype leakage into release bundles.

> [!NOTE]
> **Design Rationale (DA-CMP-004)**: Eliminates contradictory behavior between `lint_cardtown.ps1` and `compile_blueprints.ps1`. Both tools permit WIP prototype cards during local development, but enforce a zero-tolerance gate under `-Strict` compilation.

#### FR-CMP-006: Deterministic Blueprint ID Derivation & Slug Collision Invariant
The compiler derives canonical blueprint IDs strictly conforming to the deterministic slug derivation algorithm codified in **FR-DAT-002**:
1. Extracts `name` and `faction`, strips punctuation, converts to lowercase, and replaces whitespace sequences with a single underscore (`_`).
2. Prepends canonical faction prefixes (`neutral_` for `NEUTRAL`, faction prefix for Headquarters and faction assets) per **FR-DAT-002**.
3. **Zero-Collision Invariant**: The compiler maintains a global lookup table of all derived slug IDs. If two rows produce identical IDs, compilation terminates immediately with `ERR_SLUG_COLLISION`, reporting both card titles and source files.

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
| `subtypes` | Array of Strings | Split on commas, trim whitespace, and filter out empty or whitespace-only tokens. If value is `"--"` or empty, emits empty array `[]` (`FR-DAT-007`). | `ERR_SUBTYPES_INVALID` |
| `rulesText` | String | Trimmed. If empty, emits empty string `""`. | N/A |

---

## DOMAIN 3: NATURAL LANGUAGE RULES TEXT PARSER & TOKENIZER

> [!NOTE]
> **Design Rationale (DA-CMP-002: Decoupled Declarative Catalog Model)**:
> In accordance with Direction B, automated regex parsing within `compile_blueprints.ps1` is deliberately restricted to well-bounded, predictable syntactic patterns: standalone canonical keywords (`FR-CMP-013`), standard activated ability cost headers (`FR-CMP-010`), and simple single-clause triggers (`FR-CMP-009`). Any card with compound costs, modal branching (`Choose one:`), dynamic state variables, or complex non-combat mechanics bypasses regex parsing and is authored directly as declarative JSON in `data/schemas/ast_overrides.json` (supported by the AI Card Compiler authoring persona). This eliminates the fragility of regular-expression NLP on context-sensitive rules prose.

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
*(Subsumed by FR-TXT-004 & FR-TGT-001)*
Targeting grammar, selector parsing (`chosen`, `each of up to N chosen`, `distribute`), and filter conjunctions are canonically defined in **FR-TXT-004** and **FR-TGT-001**. The compiler maps canonical targeting verbs (`chosen`, `each of up to N chosen`) and filter conjunctions (`allied`, `opposing`, `unit`, `structure`, `asset`, `non-Scratch`, `non-Headquarters`, `in the Main Theater`, `in the same zone`) to `TargetRef` and `FilterDef` AST nodes directly per **FR-TGT-001**. Non-canonical targeting verbs (e.g., `target`) trigger fatal error `ERR_PARITY_VERB` under `-Strict`.

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
The compiler translates standalone keyword declarations and numeric keyword variants into declarative AST nodes matching the canonical keyword specifications defined in **FR-ABL-012** (and Master FRD `FR-KWD-002a` through `FR-KWD-002o`):
1. **Keyword Mapping**: Standalone keyword tokens (`Stalwart`, `Resistant N`, `Ward`, `Unbowed`, `Aggressive N`, `Guardian`, `Armor N`, `Versatile`, `Ethereal`, `Deploys Exhausted`, `Scratch N`, `Regroup N`, `Heal N`, `Berserk`) are mapped directly to their corresponding declarative `AbilityDef` AST structures (Triggered, Replacement Watcher, Static Restriction, Static Attribute, Capability Flag, or Atomic Primitive) as codified in **FR-ABL-012**.
2. **Conditional Keywords (Cornered)**: Lines matching `Cornered: [Keyword/Buff]` are compiled with an enclosing `conditionGate: { type: "CORNERED" }` (`FR-KWD-002e`). For example:
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
         "rulesText": "Exact raw rules text string matching source CSV",
         "parityVerified": true,
         "verifiedBy": "designer-alias",
         "verifiedAt": "2026-10-13T00:00:00Z",
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
4. **Desynchronization Check (Sanitized String Match)**:
   - To prevent silent divergence between design balance spreadsheets and the sidecar overrides registry, every entry in `ast_overrides.json` must declare the expected raw `rulesText`.
   - `compile_blueprints.ps1` executes a sanitized, whitespace-normalized string comparison between `overrides[id].rulesText` and the card's `rulesText` from the source CSV.
   - If the text differs, compilation halts immediately with `ERR_OVERRIDE_TEXT_MISMATCH`, forcing the designer or engineer to review and synchronize the sidecar override AST.

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
The compiler verifies that card rules text is faithfully represented in the compiled AST:
1. **Standard Cards (Automated Synthesis)**: The compiler inspects every non-overridden card's printed `rulesText` and verifies that every semantic element corresponds to an active AST node:
   - **Keyword Coverage**: If `rulesText` contains a recognized keyword (e.g., `Stalwart`, `Ward`, `Aggressive 2`), the card's `abilities` array must contain the corresponding AST node (`FR-CMP-013`).
   - **Trigger Coverage**: If `rulesText` contains a trigger preamble (`When`, `Whenever`, `At`), the card must declare a `TRIGGERED` ability with the matching `triggerHook` (`FR-CMP-009`).
   - **Activated Ability Coverage**: If `rulesText` contains a colon (`:`), the card must declare an `ACTIVATED` ability with matching costs (`FR-CMP-010`).
   - **Failure Condition**: Any printed rule clause that fails to map to an AST node triggers compilation error `ERR_PARITY_UNMAPPED_TEXT`.
2. **Overridden Cards (`ast_overrides.json`)**: Cards defined in the sidecar overrides registry are exempt from deep NLP clause segmentation. Instead, the compiler verifies:
   - **Keyword Presence**: If `rulesText` contains recognized standalone keywords (e.g., `Stalwart`, `Versatile`), the override's `abilities` array must contain the corresponding keyword AST nodes.
   - **Schema Conformance**: The override must strictly validate against `AbilityDef.schema.json`.
   - **Rules Text Parity**: Must match the source CSV `rulesText` via sanitized string comparison (`FR-CMP-017`).
   - **Designer Sign-Off Invariant**: Must declare `"parityVerified": true`, `verifiedBy`, and `verifiedAt`. If `"parityVerified"` is missing or `false`, compilation fails with `ERR_OVERRIDE_UNVERIFIED`.

#### FR-CMP-020: Reverse Parity Verification (AST-to-Text)
The compiler verifies that no blueprint possesses "zombie" or "ghost" abilities absent from its printed text:
1. **Standard Synthesized Cards**: Every `AbilityDef` in a blueprint's `abilities` array generated via automated synthesis must correspond to a recognizable clause in `rulesText`. Any unrepresented AST node triggers compilation error `ERR_PARITY_ORPHAN_AST`.
2. **Sidecar Overridden Cards**: Cards defined in `ast_overrides.json` are formally exempt from automated reverse AST-to-text NLP mapping. Because bespoke AST structures cannot be algorithmically decompiled into arbitrary English prose, reverse parity for overridden cards is strictly satisfied via:
   - Explicit designer verification metadata (`"parityVerified": true`, `verifiedBy`, `verifiedAt`) per `FR-CMP-017` and `FR-CMP-019`.
   - Sanitized, whitespace-normalized string equality between `overrides[id].rulesText` and the source CSV `rulesText`.
   - Complete schema conformance against `AbilityDef.schema.json`.
3. Blueprints are prohibited from declaring hidden, silent, or unprinted capabilities (excluding universal inherent actions: March, Attack, Generate; `FR-INH-001` - `FR-INH-004`).
4. **Failure Condition**: Any unrepresented AST node on a standard synthesized card triggers compilation error `ERR_PARITY_ORPHAN_AST`.

> [!NOTE]
> **Design Rationale (DA-CMP-001)**: Resolves the architectural deadlock where `FR-CMP-020` would otherwise fail compilation on bespoke cards in `ast_overrides.json` (such as *NO-AH-5*, *Samsara*, or *Ambush Tactics*). Sidecar overrides bypass automated reverse prose mapping in favor of explicit designer sign-off, schema validation, and CSV rulesText string matching.

#### FR-CMP-021: Strict Mode Enforcement & Designer Review Flags
In fulfillment of **Domain 7 of Card Data Pipeline FRD**:
1. When invoked with `-Strict`, the compiler verifies that all cards flagged with unresolved review items in the Master Review Register have resolved design approvals.
2. If a card is flagged with unresolved design questions and `-Strict` is active, compilation fails with `ERR_STRICT_UNRESOLVED_REVIEW`.
3. In standard mode, review warnings are printed to standard error as non-fatal diagnostic notes.

> [!NOTE]
> **Design Rationale (DA-CMP-005)**: Replaced hardcoded "52 cards" with dynamic evaluation against unresolved items in the Master Review Register, preventing requirement obsolescence as cards are reviewed and approved.

---

## DOMAIN 7: JSON SCHEMA GENERATION & PRODUCTION BUNDLE PACKAGING

### Section 7.1: Production Schemas & Catalog Generation

#### FR-CMP-022: Formal JSON Schema Generation & Conformance
The compiler validates every blueprint against canonical JSON schemas conforming to **JSON Schema Draft 2020-12** (cross-referencing **FR-DAT-005** and Section 8.1 of Master FRD):
1. **`CardBlueprint.schema.json`**: Located at `data/schemas/CardBlueprint.schema.json`. Validates top-level card metadata, stats, costs, enums, and the `abilities` array against canonical definitions in **FR-DAT-005**.
2. **`AbilityDef.schema.json`**: Located at `data/schemas/AbilityDef.schema.json`. Validates polymorphic ability types, condition gates, cost arrays, filter criteria, and targeting selectors (`FR-ABL-001` - `FR-ABL-012`).
3. **`ast_overrides.schema.json`**: Located at `data/schemas/ast_overrides.schema.json`. Validates the structure of the sidecar overrides registry envelope, ensuring mandatory presence of `rulesText`, `parityVerified`, `verifiedBy`, `verifiedAt`, and valid polymorphic `abilities` payloads (`FR-CMP-017`).
4. **`-EmitSchemas` Flag**: When executed with `-EmitSchemas`, the compiler regenerates and overwrites all canonical schema files (`CardBlueprint.schema.json`, `AbilityDef.schema.json`, `ast_overrides.schema.json`) with authoritative schema definitions, guaranteeing synchronization between code and schemas.

> [!NOTE]
> **Design Rationale (DA-CMP-006)**: Authoritative JSON schema added for `ast_overrides.json` sidecar registry, enabling IDE schema validation and upfront syntax linting for bespoke card abilities.

#### FR-CMP-023: Production Catalog Emission
The compiled production bundle is emitted to `data/schemas/blueprints.json`:
1. **Envelope Structure**:
   ```json
   {
     "$schema": "./CardBlueprint.schema.json",
     "catalogChecksum": "0x4f8a1c9e2b3d7a016f4e8b2c1d9a3e5f7a0b2c4d6e8f1a3b5c7d9e1f3a5b7c9d",
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
3. **PowerShell Implementation & Session Caching Strategy**:
   - To guarantee strict RFC 8785 compliance and avoid the non-compliant quirks of native `ConvertTo-Json` in Windows PowerShell 5.1 (unordered properties, floating-point localization, arbitrary whitespace), `compile_blueprints.ps1` executes canonical serialization via an embedded C# JCS helper or native BCL serialization.
   - **Session Caching**: To prevent `Add-Type` from invoking `csc.exe` out-of-process on repeated executions, the script checks if the type is already loaded in the active session (`if (-not ([System.Management.Automation.PSTypeName]'CardTown.CanonicalJson').Type)`). If present, compilation is bypassed entirely.
   - **Pre-Compiled Assembly / Native BCL Support**: The compiler supports loading a pre-compiled standalone assembly (`CardTown.CanonicalJson.dll`) or utilizing native PowerShell 7+ .NET Core APIs (`System.Text.Json` with custom ordered property writers and `System.Security.Cryptography.SHA256`), eliminating the 1,500–3,500 ms in-memory compilation penalty during git pre-commit hook runs.

> [!NOTE]
> **Design Rationale (DA-CMP-003)**: Session caching and native BCL / pre-compiled assembly support eliminate recurring `csc.exe` invocation latency, ensuring git pre-commit hooks easily meet the 5,000 ms SLA under `FR-CMP-002`.

#### FR-CMP-025: 256-Bit Cryptographic Checksum Computation (`catalogChecksum`)
In fulfillment of **FR-DAT-004**, **FR-SES-001**, and **FR-SPC-002**:
1. The compiler computes the cryptographic hash of the RFC 8785 canonical bytes using **SHA-256**.
2. Retains the full standard 256-bit hash formatted as a 0x-prefixed 64-character lowercase hexadecimal string: e.g. `"0x4f8a1c9e2b3d7a016f4e8b2c1d9a3e5f7a0b2c4d6e8f1a3b5c7d9e1f3a5b7c9d"`.
3. This checksum is embedded in the catalog envelope (`blueprints.json`), checked during server session handshakes (`FR-SES-001`), and recorded in historical match replay headers (`ReplayHeader`; `FR-SPC-002`).

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
| **AST Synthesis & Overrides** | `ERR_SYNTAX_UNPARSEABLE` | Rules text sentence cannot be parsed by lexer or grammar engine. |
| | `ERR_OVERRIDE_SCHEMA_INVALID`| Sidecar override entry fails schema validation. |
| | `ERR_OVERRIDE_TEXT_MISMATCH`| Sidecar override rulesText does not match source CSV rulesText (`FR-CMP-017`). |
| | `ERR_OVERRIDE_UNVERIFIED` | Sidecar override missing required designer parityVerified flag (`FR-CMP-019`). |
| | `ERR_ORPHAN_OVERRIDE` | Override entry defined for non-existent blueprint slug. |
| **Parity Validation** | `ERR_PARITY_UNMAPPED_TEXT` | Printed rules clause has no corresponding AST node (`FR-VAL-002`). |
| | `ERR_PARITY_ORPHAN_AST` | AST ability node has no corresponding printed rules text. |
| | `ERR_PARITY_VERB` | Non-canonical targeting verb (e.g. `target` instead of `chosen`). |
| **Strict Mode** | `ERR_STRICT_PROTOTYPE_REJECTED`| Proxy or rework card encountered while `-Strict` is enabled. |
| | `ERR_STRICT_UNRESOLVED_REVIEW` | Flagged review item encountered while `-Strict` is enabled. |

#### FR-CMP-028: Non-Zero Exit Code Contract & CI/CD Integration
*(Subsumed by FR-CMP-001 & FR-CMP-027)*
The standardized non-zero process exit codes (`0` through `5`) are incorporated directly into the CLI Invocation Contract in `FR-CMP-001` and the Structured Diagnostic Error Taxonomy in `FR-CMP-027`. For CI/CD and pre-commit automation, any non-zero exit code halts pipeline execution immediately.

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
| **FR-CMP-025** | 256-Bit Cryptographic Checksum (`catalogChecksum`)| `FR-DAT-004`, `FR-SES-001`, `FR-SPC-002` |
| **FR-CMP-026** | Semantic Dataset Versioning (`cardDataVersion`)| `FR-DAT-004`, `FR-SES-001` |
| **FR-CMP-027** | Structured Diagnostic Reporting & Error Taxonomy | `FR-VAL-001`, `FR-ERR-001` |
| **FR-CMP-028** | Non-Zero Exit Code Contract & CI/CD Integration | `FR-VAL-001`, `FR-NET-001` |

---
*End of Functional Requirements Document (Card Compiler FRD v1.0)*
