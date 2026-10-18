# Card Compiler & Blueprint Bundler - Functional Requirements Document (FRD v1.1)
**CLI Contract, CSV-to-AST Catalog Linker, Schema Validation Gate & RFC 8785 Cryptographic Fingerprinting Specification**

---

## EXECUTIVE SUMMARY & REVISION HISTORY

This Functional Requirements Document (FRD) establishes the authoritative technical and operational specification for the **Card Compiler & Blueprint Bundler (`tools/importer/compile_blueprints.ps1`)**. Whereas **[Card Data Pipeline & Ability Authoring FRD](02_Card_Data_Pipeline_FRD.md)** (`FR-DAT-001` – `FR-DAT-008`, `FR-ABL-001` – `FR-ABL-012`, `FR-TXT-001` – `FR-TXT-006`) defines all card CSV schemas, target enums, slug derivation rules, and declarative Ability AST models, this document specifies the build-time compiler toolchain that links those authored artifacts into the immutable production catalog (`data/schemas/blueprints.json`) consumed by **[Client-Engine Protocol & Networking FRD](04_Client_Engine_Protocol_Networking_FRD.md)** and **[Headless Game Engine FRD](05_Headless_Game_Engine_FRD.md)**.

```text
┌──────────────────────────────────┐   ┌──────────────────────────────────┐
│  Authoring Balance CSV Sheets    │   │  Declarative Ability AST Registry│
│       (data/cards/*.csv)         │   │  (data/schemas/ast_overrides.json│
│  - 10 RFC-4180 Sheets            │   │   + FR-ABL-012 Keyword Lookup)   │
│  - Defined in FR-DAT-001 - 008   │   │  - Modeled per FR-ABL-001 - 012  │
└─────────────────┬────────────────┘   └─────────────────┬────────────────┘
                  │                                      │
                  └──────────────────┬───────────────────┘
                                     │
                       compile_blueprints.ps1 (FR-CMP-001)
                                     │
                                     ├─── [1] CLI Init & Atomic Temp Staging (FR-CMP-001 - 003)
                                     ├─── [2] CSV Reader & Collision Gate (FR-CMP-004 - 007)
                                     ├─── [3] Catalog Linking & Keyword Lookup (FR-CMP-013, 017)
                                     ├─── [4] Parity, Strict & Schema Gates (FR-CMP-019 - 022)
                                     └─── [5] RFC 8785 Hash & Envelope Emission (FR-CMP-023 - 027)
                                     │
                                     ▼
┌─────────────────────────────────────────────────────────────────────────┐
│             Compiled Production Master Catalog                          │
│                  (data/schemas/blueprints.json)                         │
│  - Canonical Envelope (catalogChecksum, cardDataVersion, blueprints)    │
│  - 256-Bit SHA-256 over RFC 8785 Canonical Bytes (FR-CMP-024 - 025)     │
└─────────────────────────────────────────────────────────────────────────┘
```

### Specification Scope (v1.1):
1. **CLI Contract, Execution Lifecycle & Atomic Staging (`FR-CMP-001` – `FR-CMP-003`)**: Command-line parameters, deterministic pipeline stages, process exit codes (`0`–`5`), and atomic temp-file staging (`blueprints.json.tmp.$PID`) with rollback guarantee.
2. **CSV Reader Mechanics & Ingestion Edge Cases (`FR-CMP-004` – `FR-CMP-007`)**: Low-level file reader edge cases (BOM/line-ending tolerance, blank row skipping, duplicate slug collision detection, and diagnostic error mapping) delegating all column schemas and type rules to `02_Card_Data_Pipeline_FRD.md` (`FR-DAT-001` – `FR-DAT-008`).
3. **Catalog Linking, Keyword Lookup & AST Registry Binding (`FR-CMP-013`, `FR-CMP-017`)**: Joining ingested CSV rows with standalone keyword definitions (`FR-ABL-012`) and modeled ability ASTs (`data/schemas/ast_overrides.json`), enforcing `rulesText` synchronization (`ERR_OVERRIDE_TEXT_MISMATCH`) and orphan detection (`ERR_ORPHAN_OVERRIDE`).
4. **Build-Time Parity, Review & JSON Schema Validation Gates (`FR-CMP-019` – `FR-CMP-022`)**: Designer parity sign-off verification (`parityVerified`), `-Strict` production gating (`FR-VAL-003`), and JSON Schema Draft 2020-12 validation.
5. **RFC 8785 Canonical Hashing, Envelope Emission & CLI Diagnostics (`FR-CMP-023` – `FR-CMP-027`)**: Production `blueprints.json` envelope assembly, RFC 8785 JCS normalization, 256-bit SHA-256 `catalogChecksum` computation, semantic dataset versioning, and structured `ERR_*` CLI reporting.

---

## ARCHITECTURAL DECISIONS

> [!IMPORTANT]
> **DECISION 1: Pure Catalog Linker Architecture — Delegating CSV & AST Modeling to FRD 02 (`FR-CMP-017`)**
> - **Context**: All card data schemas (`FR-DAT-001` – `FR-DAT-008`), rules text grammar (`FR-TXT-001` – `FR-TXT-006`), and declarative Ability AST structures (`FR-ABL-001` – `FR-ABL-012`, `FR-TRG-001`, `FR-PRM-001`) are canonically specified in [`02_Card_Data_Pipeline_FRD.md`](02_Card_Data_Pipeline_FRD.md).
> - **Decision**: `compile_blueprints.ps1` operates strictly as a deterministic **Catalog Linker, Validator, and Checksum Bundler** rather than a natural-language regex parser:
>   1. **Single Source of Schema Truth**: CSV column layouts (15-column asset sheets and 9-column maneuver sheets), enum sets, boolean literals (`"true"` / `"false"`), and slug derivation rules are referenced directly from `FR-DAT-001` – `FR-DAT-008`.
>   2. **Keyword Lookup + Declarative AST Registry**: Vanilla cards (`rulesText` equal to `"--"` or `""`) emit `abilities: []` (`FR-DAT-007`), and standalone keyword lines map via the 15-keyword lookup table in `FR-ABL-012` (`FR-CMP-013`). All non-keyword card abilities and additional deployment/enactment costs (`additionalDeploymentCosts`, `FR-DAT-005`) are modeled declaratively per `02_Card_Data_Pipeline_FRD.md` and linked from `data/schemas/ast_overrides.json` (`FR-CMP-017`). Sentence-level regex NLP synthesis (`FR-CMP-008` – `FR-CMP-012`, `FR-CMP-014` – `FR-CMP-016`) is retired and subsumed.

> [!IMPORTANT]
> **DECISION 2: 256-Bit Cryptographic Checksum via RFC 8785 & Full SHA-256 (`FR-CMP-024`, `FR-CMP-025`)**
> - **Decision**: `blueprints.json` is formatted with 2-space indentation in Git for readable diffs (`FR-CMP-023`), while `catalogChecksum` is computed via SHA-256 (`0x`-prefixed 64-char lowercase hex) over the RFC 8785 Canonical JSON bytes (sorted keys, no whitespace, UTF-8 without BOM), guaranteeing cross-platform parity independent of file formatting.

---

## DOMAIN 1: CLI CONTRACT, EXECUTION LIFECYCLE & ATOMIC STAGING

### Section 1.1: CLI Invocation & Staging Invariants

#### FR-CMP-001: Compiler CLI Command Interface & Parameter Contract
The Card Compiler CLI script resides at `tools/importer/compile_blueprints.ps1` and supports PowerShell 5.1 and PowerShell 7+ across Windows, Linux, and macOS:

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

1. **`-SourceDir`**: Directory containing the 10 authoritative CSV sheets defined in [`FR-DAT-001`](02_Card_Data_Pipeline_FRD.md).
2. **`-OutputDir`**: Target directory for `blueprints.json` (and `blueprints.min.json` when `-Minify` is set).
3. **`-SchemaDir`**: Directory containing `CardBlueprint.schema.json`, `AbilityDef.schema.json`, and `ast_overrides.schema.json`.
4. **`-OverridesFile`**: Path to the declarative AST registry (`data/schemas/ast_overrides.json`; `FR-CMP-017`).
5. **`-ValidateOnly`**: Dry-run mode executing full ingestion, linking, parity verification, and schema validation without writing output files.
6. **`-EmitSchemas`**: Regenerates the standalone JSON Schema files in `-SchemaDir` (`FR-CMP-022`).
7. **`-Strict`**: Treats prototype/rework flags (`currentProxy`, `needsRework`; [`FR-VAL-003`](02_Card_Data_Pipeline_FRD.md)) and unresolved review items (`FR-CMP-021`) as fatal errors.
8. **`-Minify`**: Emits `blueprints.min.json` alongside the 2-space indented `blueprints.json`.
9. **`-Version`**: Sets `cardDataVersion` (SemVer 2.0, e.g., `"1.0.0"`; `FR-CMP-026`).
10. **`-VerboseOutput`**: Emits per-card linking and validation telemetry to stdout.

##### Process Exit Code Contract (Subsumes FR-CMP-028)
| Exit Code | Classification | Trigger Scenarios & Error Taxonomy (`FR-CMP-027`) |
| :---: | :--- | :--- |
| **`0`** | **Success** | Compilation, validation, and emission succeeded. |
| **`1`** | **Environment / I/O** | Missing, unreadable, or locked CSV, override, or schema file (`ERR_*_IO`). |
| **`2`** | **CSV Syntax / Parsing** | Malformed RFC-4180 CSV row, header mismatch, or empty card name (`ERR_CSV_*`, `ERR_NAME_EMPTY`). |
| **`3`** | **Type / Enum / Slug** | Type coercion failure, illegal enum value, or duplicate slug ID (`ERR_ENUM_*`, `ERR_NUM_*`, `ERR_SLUG_COLLISION`). |
| **`4`** | **AST Registry / Schema** | Override schema failure, orphan override, text mismatch, or blueprint schema failure (`ERR_OVERRIDE_*`, `ERR_ORPHAN_OVERRIDE`). |
| **`5`** | **Parity / Strict Gate** | Unmapped rules text, unverified override, or proxy/review card rejected under `-Strict` (`ERR_PARITY_*`, `ERR_STRICT_*`). |

#### FR-CMP-002: Multi-Stage Compilation Pipeline & Lifecycle Invariants
The compiler executes deterministically across five sequential stages, halting immediately on any fatal error:
1. **Stage 1 (Environment & Schema Initialization)**: Resolves paths and loads JSON schemas plus `ast_overrides.json`.
2. **Stage 2 (Atomic CSV Ingestion & Slug Derivation)**: Parses all 10 CSV sheets (`FR-CMP-004` – `FR-CMP-007`) in accordance with [`FR-DAT-001`](02_Card_Data_Pipeline_FRD.md) – [`FR-DAT-008`](02_Card_Data_Pipeline_FRD.md) and checks for duplicate derived IDs.
3. **Stage 3 (Catalog Linking & Keyword Lookup)**: Binds standalone keyword ASTs (`FR-CMP-013`) and declarative AST overrides (`FR-CMP-017`).
4. **Stage 4 (Build-Time Parity, Strict & Schema Validation)**: Verifies text-to-AST synchronization (`FR-CMP-019` – `FR-CMP-020`), `-Strict` gates (`FR-CMP-021`), and JSON Schema conformance (`FR-CMP-022`).
5. **Stage 5 (RFC 8785 Fingerprinting & Atomic Emission)**: Computes `catalogChecksum` (`FR-CMP-024` – `FR-CMP-025`) and atomically writes `blueprints.json` (`FR-CMP-003`, `FR-CMP-023`). Warm execution for ~200 cards must complete within 5,000 ms on PowerShell 7+ (< 10,000 ms on Windows PowerShell 5.1).

#### FR-CMP-003: Atomic Emission, File Staging & Rollback Guarantee
In fulfillment of [`FR-DAT-003`](02_Card_Data_Pipeline_FRD.md):
1. The compiler never writes directly to `<OutputDir>/blueprints.json`.
2. Output is written first to a process-unique staging file `<OutputDir>/blueprints.json.tmp.$PID`.
3. Only after all validation and hashing stages succeed is the staging file atomically renamed/replaced over `blueprints.json`.
4. On any error, `<OutputDir>/blueprints.json.tmp.$PID` is deleted immediately, leaving any pre-existing `blueprints.json` untouched.

---

## DOMAIN 2: CSV READER MECHANICS & INGESTION EDGE CASES

### Section 2.1: Low-Level CSV Reader & Coercion Diagnostics

#### FR-CMP-004: Low-Level RFC-4180 Reader & Encoding Edge Cases
The compiler ingests the 10 CSV sheets and column layouts defined in [`FR-DAT-001`](02_Card_Data_Pipeline_FRD.md) and [`FR-DAT-006`](02_Card_Data_Pipeline_FRD.md) (15-column `units-and-structures-*.csv` and 9-column `maneuvers-*.csv`) with the following reader mechanics:
1. Accepts UTF-8 files with or without a Byte Order Mark (BOM).
2. Normalizes Windows (`CRLF`) and Unix (`LF`) line endings interchangeably.
3. Preserves embedded newlines and escaped double-quotes (`""`) inside quoted `rulesText` cells without splitting rows.
4. Aborts atomically with exit code `1` (`ERR_CSV_IO`) if any of the 10 files is missing or unreadable, or exit code `2` (`ERR_CSV_MALFORMED`) if a row violates RFC-4180 framing or column headers in [`FR-DAT-006`](02_Card_Data_Pipeline_FRD.md).

#### FR-CMP-005: Blank Row Skipping & Prototype Strict Gate
1. **Blank Row Sanitation**: Trailing or intermediate CSV rows where all cells are empty or whitespace-only are silently skipped.
2. **Mandatory Card Name**: Any non-blank row with an empty or whitespace-only `name` cell halts compilation with `ERR_NAME_EMPTY`.
3. **Prototype & Rework Flag Enforcement ([`FR-VAL-003`](02_Card_Data_Pipeline_FRD.md))**: Boolean columns `currentProxy` and `needsRework` (`"true"` / `"false"` per [`FR-DAT-006`](02_Card_Data_Pipeline_FRD.md)) serialize directly to `"currentProxy"` and `"needsRework"` on the `CardBlueprint` (`FR-DAT-005`). When `-Strict` is active, any row with `currentProxy == "true"` or `needsRework == "true"` halts compilation with `ERR_STRICT_PROTOTYPE_REJECTED`.

#### FR-CMP-006: Slug Collision Detection Invariant
The compiler derives each card's `id` using the deterministic `<faction>_<slug>` algorithm defined in [`FR-DAT-002`](02_Card_Data_Pipeline_FRD.md) and maintains a global ID lookup table across all 10 sheets. If two rows produce an identical `id`, compilation halts with `ERR_SLUG_COLLISION`, reporting both card names and source files.

#### FR-CMP-007: Coercion Failure Diagnostic Mapping
All CSV field coercions, maneuver default values, subtype/sentinel normalizations (`"--"`), and deployment zone constraints follow [`FR-DAT-005`](02_Card_Data_Pipeline_FRD.md) through [`FR-DAT-008`](02_Card_Data_Pipeline_FRD.md). Any violation maps deterministically to the error codes in `FR-CMP-027` (`ERR_ENUM_INVALID_CARD_TYPE`, `ERR_ENUM_INVALID_FACTION`, `ERR_ENUM_INVALID_RARITY`, `ERR_ENUM_INVALID_ZONE`, `ERR_NUM_PARSE`, `ERR_NUM_NEGATIVE_COST`, `ERR_NUM_INVALID_POWER`, `ERR_NUM_INVALID_HEALTH`, `ERR_NUM_NEGATIVE_YIELD`, `ERR_SUBTYPES_INVALID`).

---

## DOMAIN 3: CATALOG LINKING, KEYWORD LOOKUP & AST REGISTRY BINDING

### Section 3.1: Retired Regex NLP Requirements (Subsumed by FRD 02 & FR-CMP-017)

#### FR-CMP-008: Rules Text Lexer & Clause Segmentation
*(Subsumed by [`FR-TXT-001`](02_Card_Data_Pipeline_FRD.md), `FR-CMP-013` & `FR-CMP-017`)*

#### FR-CMP-009: Trigger Preamble Parser & Temporal Hook Mapping
*(Subsumed by [`FR-TXT-002`](02_Card_Data_Pipeline_FRD.md), [`FR-TRG-001`](02_Card_Data_Pipeline_FRD.md), [`FR-ABL-002`](02_Card_Data_Pipeline_FRD.md) & `FR-CMP-017`)*

#### FR-CMP-010: Ability Cost Expression Parser
*(Subsumed by [`FR-TXT-003`](02_Card_Data_Pipeline_FRD.md), [`FR-DAT-005`](02_Card_Data_Pipeline_FRD.md), [`FR-ABL-007`](02_Card_Data_Pipeline_FRD.md), [`FR-ABL-009`](02_Card_Data_Pipeline_FRD.md) & `FR-CMP-017`)*

#### FR-CMP-011: Declarative Target Filter & Selector Synthesizer
*(Subsumed by [`FR-TXT-004`](02_Card_Data_Pipeline_FRD.md) & [`FR-TGT-001`](01_Game_Rules_FRD.md))*

#### FR-CMP-012: Numerical & Value Token Synthesizer
*(Subsumed by [`FR-ABL-007`](02_Card_Data_Pipeline_FRD.md) & `FR-CMP-017`)*

### Section 3.2: Keyword Lookup & Declarative AST Linking

#### FR-CMP-013: Vanilla Sentinel & Standalone Keyword Lookup
For cards not present in `ast_overrides.json` (`FR-CMP-017`), the compiler resolves `rulesText` strictly for vanilla sentinels and standalone canonical keywords:
1. **Vanilla Cards ([`FR-DAT-007`](02_Card_Data_Pipeline_FRD.md))**: If `rulesText` is empty (`""`) or equals the vanilla sentinel `"--"`, the compiler emits `abilities: []` andnormalizes `rulesText` per `FR-DAT-007`.
2. **Standalone Keyword Lines ([`FR-ABL-012`](02_Card_Data_Pipeline_FRD.md))**: If every line in `rulesText` matches a standalone canonical keyword from `FR-ABL-012` (`Stalwart`, `Resistant N`, `Ward`, `Unbowed`, `Aggressive N`, `Guardian`, `Armor N`, `Versatile`, `Ethereal`, `Deploys Exhausted`, `Berserk`), the compiler emits the corresponding `AbilityDef` entries directly from the `FR-ABL-012` lookup table.
3. Any card containing non-keyword rules text sentences MUST have a corresponding entry in `data/schemas/ast_overrides.json` (`FR-CMP-017`) or compilation fails with `ERR_PARITY_UNMAPPED_TEXT` (`FR-CMP-019`).

#### FR-CMP-014: Triggered & Activated Ability AST Synthesizer
*(Subsumed by [`FR-ABL-002`](02_Card_Data_Pipeline_FRD.md), [`FR-ABL-003`](02_Card_Data_Pipeline_FRD.md), [`FR-TGT-101`](02_Card_Data_Pipeline_FRD.md) & `FR-CMP-017`)*

#### FR-CMP-015: Continuous Aura & Replacement Watcher Synthesizer
*(Subsumed by [`FR-ABL-004`](02_Card_Data_Pipeline_FRD.md), [`FR-ABL-005`](02_Card_Data_Pipeline_FRD.md) & `FR-CMP-017`)*

#### FR-CMP-016: Maneuver Spell Payload & Modal Choice Synthesizer
*(Subsumed by [`FR-ABL-006`](02_Card_Data_Pipeline_FRD.md), [`FR-ABL-010`](02_Card_Data_Pipeline_FRD.md) & `FR-CMP-017`)*

#### FR-CMP-017: Sidecar AST Overrides Registry & Catalog Linking Contract
The compiler joins ingested CSV card rows with their declarative AST definitions modeled per [`02_Card_Data_Pipeline_FRD.md`](02_Card_Data_Pipeline_FRD.md) (`FR-ABL-001` – `FR-ABL-012`) from `data/schemas/ast_overrides.json`:
1. **Overrides Registry Envelope**:
   ```json
   {
     "$schema": "./ast_overrides.schema.json",
     "overrides": {
       "<blueprint_id>": {
         "rulesText": "Exact rulesText string from source CSV",
         "parityVerified": true,
         "verifiedBy": "designer-alias",
         "verifiedAt": "2026-10-13T00:00:00Z",
         "rationale": "Optional authoring note",
         "additionalDeploymentCosts": [ ... optional CostDef array per FR-DAT-005 ... ],
         "abilities": [ ... AbilityDef array per FR-ABL-001 ... ]
       }
     }
   }
   ```
2. **Linking by Derived `id`**: When a card's derived `id` (`FR-CMP-006`) exists in `overrides`, the compiler binds `abilities` (and optional `additionalDeploymentCosts`) from the registry entry onto the compiled `CardBlueprint`.
3. **Orphan Override Detection**: Every `<blueprint_id>` key in `ast_overrides.json` must match an ingested CSV row; any unmatched key halts compilation with `ERR_ORPHAN_OVERRIDE`.
4. **CSV-to-AST `rulesText` Synchronization Check**: The compiler performs a whitespace-normalized string comparison between `overrides[id].rulesText` and the CSV row's `rulesText`. Any divergence halts compilation immediately with `ERR_OVERRIDE_TEXT_MISMATCH`.

#### FR-CMP-018: Bespoke Mechanics Integration & Validation
*(Subsumed by `FR-CMP-017` and Section 4.2 of [`02_Card_Data_Pipeline_FRD.md`](02_Card_Data_Pipeline_FRD.md))*

---

## DOMAIN 4: BUILD-TIME PARITY, REVIEW & JSON SCHEMA VALIDATION GATES

### Section 4.1: Parity, Strict & Schema Gates

#### FR-CMP-019: Forward Parity Verification (Text-to-AST)
In fulfillment of [`FR-VAL-002`](02_Card_Data_Pipeline_FRD.md), the compiler verifies that every card's printed `rulesText` is accounted for in its compiled blueprint:
1. **Unmapped Non-Keyword Text**: Any card with non-vanilla, non-standalone-keyword `rulesText` that lacks an entry in `ast_overrides.json` fails compilation with `ERR_PARITY_UNMAPPED_TEXT`.
2. **Standalone Keyword Coverage**: If a card's `rulesText` includes a standalone keyword line (per `FR-ABL-012`), its compiled `abilities` array must include the corresponding keyword AST node.
3. **Canonical Targeting Verb Check**: Under `-Strict`, `rulesText` using non-canonical targeting terminology (e.g., `"target"` instead of `"chosen"` per [`FR-TXT-004`](02_Card_Data_Pipeline_FRD.md)) fails with `ERR_PARITY_VERB`.

#### FR-CMP-020: Reverse Parity & Override Sign-Off Verification
To prevent unprinted abilities or unreviewed AST edits ([`FR-VAL-002`](02_Card_Data_Pipeline_FRD.md)):
1. **Keyword-Only / Vanilla Cards**: May only contain `AbilityDef` nodes directly mapped from their standalone keyword lines (`FR-CMP-013`); any extraneous AST node fails with `ERR_PARITY_ORPHAN_AST`.
2. **Registry-Linked Cards (`ast_overrides.json`)**: Every entry in `ast_overrides.json` must declare `"parityVerified": true`, a non-empty `verifiedBy`, and an ISO-8601 `verifiedAt` timestamp, in addition to passing the `ERR_OVERRIDE_TEXT_MISMATCH` check (`FR-CMP-017`) and `AbilityDef.schema.json` validation (`FR-CMP-022`). Missing or `false` `"parityVerified"` metadata halts compilation with `ERR_OVERRIDE_UNVERIFIED`.

#### FR-CMP-021: Strict Mode Review Gate
When invoked with `-Strict`, the compiler checks whether any card carries an unresolved review flag in Domain 7 of [`02_Card_Data_Pipeline_FRD.md`](02_Card_Data_Pipeline_FRD.md) and halts with `ERR_STRICT_UNRESOLVED_REVIEW` if any remain open. In standard mode, open review flags emit non-fatal warnings to stderr.

#### FR-CMP-022: Formal JSON Schema Validation & `-EmitSchemas`
In fulfillment of [`FR-DAT-005`](02_Card_Data_Pipeline_FRD.md) and [`FR-VAL-001`](02_Card_Data_Pipeline_FRD.md), the compiler validates all inputs and outputs against JSON Schema Draft 2020-12 definitions in `data/schemas/`:
1. **`CardBlueprint.schema.json`**: Validates the root `blueprints.json` envelope (`catalogChecksum`, `cardDataVersion`, `generatedAt`, `totalCardCount`, `blueprints`) and each `CardBlueprint` object per [`FR-DAT-005`](02_Card_Data_Pipeline_FRD.md).
2. **`AbilityDef.schema.json`**: Validates all polymorphic `AbilityDef`, `FilterDef`, `ValueToken`, `CostDef`, `TargetRef`, `ConditionDef`, and `EffectDef` nodes per [`FR-ABL-001`](02_Card_Data_Pipeline_FRD.md) – [`FR-ABL-012`](02_Card_Data_Pipeline_FRD.md).
3. **`ast_overrides.schema.json`**: Validates `data/schemas/ast_overrides.json` (`FR-CMP-017`); failures halt compilation with `ERR_OVERRIDE_SCHEMA_INVALID`.
4. **`-EmitSchemas` Flag**: Regenerates all three `.schema.json` files from the canonical definitions when passed on the CLI.

---

## DOMAIN 5: RFC 8785 CANONICAL HASHING, ENVELOPE EMISSION & CLI DIAGNOSTICS

### Section 5.1: Production Envelope, Fingerprinting & Error Taxonomy

#### FR-CMP-023: Production Catalog Envelope Emission
The compiled catalog is emitted to `data/schemas/blueprints.json` matching the root envelope schema in [`FR-DAT-005`](02_Card_Data_Pipeline_FRD.md):
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
1. **Repository Formatting**: Emitted with 2-space indentation and `LF` newlines for readable Git diffs, with `blueprints` sorted lexicographically by `id`.
2. **Minified Artifact**: When `-Minify` is specified, an unformatted `blueprints.min.json` is also emitted.

#### FR-CMP-024: Canonical JSON Normalization (RFC 8785 JCS)
Before computing `catalogChecksum`, the compiler serializes the `blueprints` payload in accordance with **RFC 8785 (JSON Canonicalization Scheme)**:
1. Object keys are sorted lexicographically by UTF-16 code units.
2. All insignificant whitespace is omitted, and strings are encoded as UTF-8 without BOM.
3. **Session Caching**: To avoid repeated `csc.exe` startup overhead in PowerShell 5.1/7+, the canonical serializer uses session-cached types or native .NET `System.Text.Json` / `System.Security.Cryptography.SHA256` APIs so pre-commit hook runs stay well within the 5,000 ms budget (`FR-CMP-002`).

#### FR-CMP-025: 256-Bit Cryptographic Checksum Computation (`catalogChecksum`)
In fulfillment of [`FR-DAT-004`](02_Card_Data_Pipeline_FRD.md), [`FR-SES-001`](04_Client_Engine_Protocol_Networking_FRD.md), and [`FR-SPC-002`](04_Client_Engine_Protocol_Networking_FRD.md):
1. Computes the SHA-256 digest over the RFC 8785 canonical bytes (`FR-CMP-024`).
2. Formats `catalogChecksum` as a `0x`-prefixed 64-character lowercase hexadecimal string.

#### FR-CMP-026: Semantic Dataset Versioning Contract (`cardDataVersion`)
`cardDataVersion` follows Semantic Versioning 2.0 (`MAJOR.MINOR.PATCH`): **MAJOR** for breaking schema changes, **MINOR** for card balance/text/catalog changes, and **PATCH** for non-mechanical tooling or typo fixes. Session handshakes (`FR-SES-001`) verify exact `catalogChecksum` equality.

#### FR-CMP-027: Structured Diagnostic Reporting & Error Taxonomy
On failure, the compiler prints a structured error block to stderr identifying the error code, source file, row/card identifier, offending field/value, and violated requirement:

| Category | Error Code | Description |
| :--- | :--- | :--- |
| **I/O & Environment** | `ERR_CSV_IO` | Missing, unreadable, or locked CSV sheet. |
| | `ERR_SCHEMA_IO` | Unable to read or write JSON schema or overrides file. |
| | `ERR_OUTPUT_IO` | Unable to write staging or production JSON artifacts. |
| **CSV & Coercion** | `ERR_CSV_MALFORMED` | Malformed RFC-4180 CSV row or header mismatch (`FR-DAT-006`). |
| | `ERR_NAME_EMPTY` | Non-blank row declares an empty card `name`. |
| | `ERR_NUM_PARSE` | Non-integer value in an integer column (`FR-DAT-005`). |
| | `ERR_NUM_NEGATIVE_COST` | Negative `requisitionCost` or `tacticsCost`. |
| | `ERR_NUM_INVALID_POWER` | Invalid `power` value (`< 0` for assets or non-null for maneuvers). |
| | `ERR_NUM_INVALID_HEALTH` | Invalid `health` value (`< 1` for assets or non-null for maneuvers). |
| | `ERR_NUM_NEGATIVE_YIELD`| Negative yield or non-zero yield on a maneuver. |
| | `ERR_ENUM_INVALID_CARD_TYPE` | `cardType` violates `FR-DAT-006`. |
| | `ERR_ENUM_INVALID_FACTION` | `faction` violates `FR-DAT-006`. |
| | `ERR_ENUM_INVALID_RARITY` | `rarity` violates `FR-DAT-006`. |
| | `ERR_ENUM_INVALID_ZONE` | `deploymentZones` violates `FR-DAT-006` / `FR-DAT-008`. |
| | `ERR_SUBTYPES_INVALID` | Malformed `subtypes` value (`FR-DAT-007`). |
| | `ERR_SLUG_COLLISION` | Two rows derive identical `id` slugs (`FR-CMP-006`). |
| **Linking & Schemas** | `ERR_OVERRIDE_SCHEMA_INVALID` | Entry in `ast_overrides.json` fails `AbilityDef.schema.json`. |
| | `ERR_OVERRIDE_TEXT_MISMATCH` | `ast_overrides.json` `rulesText` differs from CSV `rulesText` (`FR-CMP-017`). |
| | `ERR_OVERRIDE_UNVERIFIED` | Override entry missing `"parityVerified": true` metadata (`FR-CMP-020`). |
| | `ERR_ORPHAN_OVERRIDE` | Override key does not match any CSV card `id` (`FR-CMP-017`). |
| **Parity & Strict Gate**| `ERR_PARITY_UNMAPPED_TEXT` | Non-keyword `rulesText` has no `ast_overrides.json` entry (`FR-CMP-019`). |
| | `ERR_PARITY_ORPHAN_AST` | Keyword-only/vanilla card contains extraneous AST nodes (`FR-CMP-020`). |
| | `ERR_PARITY_VERB` | Non-canonical targeting verb under `-Strict` (`FR-CMP-019`). |
| | `ERR_STRICT_PROTOTYPE_REJECTED`| Card has `currentProxy` or `needsRework` set to `"true"` under `-Strict` (`FR-CMP-005`). |
| | `ERR_STRICT_UNRESOLVED_REVIEW` | Card has an open review item under `-Strict` (`FR-CMP-021`). |

#### FR-CMP-028: Non-Zero Exit Code Contract & CI/CD Integration
*(Subsumed by `FR-CMP-001` & `FR-CMP-027`)*

---

## DOMAIN 6: CROSS-SPEC TRACEABILITY MATRIX

### Section 6.1: Traceability & Requirements Index

| Requirement ID | Canonical Requirement Title | Companion FRD Traceability |
| :--- | :--- | :--- |
| **FR-CMP-001** | Compiler CLI Command Interface & Exit Codes | `FR-DAT-001`, `FR-DAT-003`, `FR-VAL-001` |
| **FR-CMP-002** | Multi-Stage Compilation Pipeline & Invariants | `FR-DAT-001`, `FR-DAT-003`, `FR-VAL-001` |
| **FR-CMP-003** | Atomic Emission, File Staging & Rollback | `FR-DAT-003`, `FR-ENG-001` |
| **FR-CMP-004** | Low-Level RFC-4180 Reader & Encoding Edge Cases | `FR-DAT-001`, `FR-DAT-006` |
| **FR-CMP-005** | Blank Row Skipping & Prototype Strict Gate | `FR-VAL-003`, `FR-DAT-006` |
| **FR-CMP-006** | Slug Collision Detection Invariant | `FR-DAT-002`, `FR-SET-001` |
| **FR-CMP-007** | Coercion Failure Diagnostic Mapping | `FR-DAT-005` – `FR-DAT-008` |
| **FR-CMP-008** – **FR-CMP-012** | *(Subsumed by FRD 02 & `FR-CMP-017`)* | `FR-TXT-001` – `FR-TXT-006`, `FR-ABL-001` – `FR-ABL-012` |
| **FR-CMP-013** | Vanilla Sentinel & Standalone Keyword Lookup | `FR-DAT-007`, `FR-ABL-012` |
| **FR-CMP-014** – **FR-CMP-016** | *(Subsumed by FRD 02 & `FR-CMP-017`)* | `FR-ABL-002` – `FR-ABL-010`, `FR-TGT-101` |
| **FR-CMP-017** | Sidecar AST Overrides Registry & Catalog Linking | `FR-DAT-003`, `FR-DAT-005`, `FR-ABL-001` |
| **FR-CMP-018** | *(Subsumed by `FR-CMP-017` & FRD 02 Section 4.2)* | `FR-ABL-001` – `FR-ABL-012` |
| **FR-CMP-019** | Forward Parity Verification (Text-to-AST) | `FR-VAL-002`, `FR-TXT-004` |
| **FR-CMP-020** | Reverse Parity & Override Sign-Off Verification | `FR-VAL-002`, `FR-CMP-017` |
| **FR-CMP-021** | Strict Mode Review Gate | `FR-VAL-003` |
| **FR-CMP-022** | Formal JSON Schema Validation & `-EmitSchemas` | `FR-DAT-005`, `FR-ABL-001`, `FR-VAL-001` |
| **FR-CMP-023** | Production Catalog Envelope Emission | `FR-DAT-003`, `FR-DAT-005`, `FR-ENG-001` |
| **FR-CMP-024** | Canonical JSON Normalization (RFC 8785 JCS) | `FR-DAT-004`, `FR-PAC-004` |
| **FR-CMP-025** | 256-Bit Cryptographic Checksum (`catalogChecksum`) | `FR-DAT-004`, `FR-SES-001`, `FR-SPC-002` |
| **FR-CMP-026** | Semantic Dataset Versioning (`cardDataVersion`) | `FR-DAT-004`, `FR-SES-001` |
| **FR-CMP-027** | Structured Diagnostic Reporting & Error Taxonomy | `FR-VAL-001`, `FR-ERR-001` |
| **FR-CMP-028** | *(Subsumed by `FR-CMP-001` & `FR-CMP-027`)* | `FR-VAL-001` |

---
*End of Functional Requirements Document (Card Compiler & Blueprint Bundler FRD v1.1)*
