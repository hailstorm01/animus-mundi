# Product Manager Devil's Advocate Audit: Card Compiler & Blueprint Bundler FRD

- **Target Document:** [docs/frd/03_Card_Compiler_FRD.md](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/03_Card_Compiler_FRD.md)
- **Auditor:** PM Devil's Advocate (`pm-devils-advocate`)
- **Evaluation Date:** 2026-10-18
- **Status:** PENDING (Active review queue)
- **Verdict:** REVISE WITH MEDIUM FINDINGS

## 1. Executive Summary & Red-Team Assessment

This audit evaluates `03_Card_Compiler_FRD.md` (FRD v1.1) focusing on build-time invariants, schema validation gates, keyword catalog resolution, and specification bloat.

`03_Card_Compiler_FRD.md` canonically governs the compilation pipeline CLI (`compile_blueprints.ps1`), RFC-4180 CSV ingestion, deterministic slug derivation, standalone keyword linking, sidecar AST registry joining (`ast_overrides.json`), RFC 8785 canonical JSON normalization, and SHA-256 fingerprinting (`catalogChecksum`).

The audit reveals significant specification clutter from legacy placeholder requirements and edge-case gaps in keyword cataloging and vanilla schema validation:
1. **Incomplete Standalone Keyword Catalog in `FR-CMP-013`**: Line 169 lists only 10 keywords (`Stalwart, Resistant N, Ward, Aggressive N, Guardian, Armor N, Versatile, Ethereal, Deploys Exhausted, Berserk`), while `01_Game_Rules_FRD.md` and `02_Card_Data_Pipeline_FRD.md` define 15 canonical keywords. The specification fails to explicitly explain why the other 5 keywords (`Unbowed, Cornered, Heal N, Scratch N, Regroup N`) cannot be compiled standalone.
2. **Vanilla `rulesText` Serialization Parity Ambiguity**: `FR-CMP-013` specifies that vanilla cards with `rulesText: "--"` compile into `blueprints.json` with `rulesText: ""` (empty string). However, `CardBlueprint.schema.json` referenced in `FR-DAT-005` specifies `rulesText: String, Non-empty`. If the schema enforces non-empty strings, compiling vanilla cards with `""` fails JSON schema validation in Stage 4 (`FR-CMP-022`).
3. **Heavy Proliferation of Retired Regex NLP Requirements**: 11 out of 28 requirement IDs in FRD 03 are retired or subsumed placeholders. Section 3.1 is an entirely retired section of legacy NLP regex requirements.

## 2. Severity-Ordered Punch-List

### 🔴 Blocker / Critical
_None._

### 🟠 High
_None._

### 🟡 Medium
- **[DA-CMP-005] Incomplete Standalone Keyword Catalog in FR-CMP-013**
  - **Requirement Citations:** `FR-CMP-013`, `FR-ABL-012`
  - **Vulnerability / Flaw:** `FR-CMP-013` states that the compiler resolves standalone canonical keywords from `FR-ABL-012`, but lists only 10 keywords (`Stalwart, Resistant N, Ward, Aggressive N, Guardian, Armor N, Versatile, Ethereal, Deploys Exhausted, Berserk`). It omits `Unbowed, Cornered, Heal N, Scratch N, Regroup N` without documenting why they cannot be resolved standalone.
  - **Exploit / Failure Scenario:** A designer authors a card with standalone rules text line `Scratch 2.` or `Regroup 1.`, expecting compiler keyword lookup, but the compiler fails with `ERR_PARITY_UNMAPPED_TEXT` because these keywords are atomic primitives requiring overrides.
  - **Remedy:** Explicitly document in `FR-CMP-013` the taxonomic distinction between the 10 pure static keywords (auto-compiled) and the 5 parameterized/primitive keywords (`Unbowed, Cornered, Heal N, Scratch N, Regroup N`) which require `ast_overrides.json` linking.

- **[DA-CMP-006] Vanilla rulesText Serialization Parity Ambiguity**
  - **Requirement Citations:** `FR-CMP-013`, `FR-DAT-005`, `FR-DAT-007`
  - **Vulnerability / Flaw:** In `FR-CMP-013`, vanilla cards with `rulesText: "--"` compile into `blueprints.json` with `rulesText: ""` (empty string). However, `CardBlueprint.schema.json` referenced in `FR-DAT-005` specifies `rulesText: String, Non-empty`. If the schema enforces non-empty strings, compiling vanilla cards with `""` fails JSON schema validation in Stage 4 (`FR-CMP-022`).
  - **Exploit / Failure Scenario:** Invoking `compile_blueprints.ps1 -Strict` fails schema validation for vanilla cards (`ERR_SCHEMA_VALIDATION_FAILED`) due to `rulesText` being normalized to an empty string.
  - **Remedy:** Clarify in `FR-CMP-013` and `FR-DAT-005` that `CardBlueprint.schema.json` permits `minLength: 0` for `rulesText` when `abilities` is an empty array (`[]`).

### 🔵 Low / Clarification
- **[DA-CMP-007] Heavy Proliferation of Retired Regex NLP Requirements**
  - **Requirement Citations:** `FR-CMP-008` through `FR-CMP-012`, `FR-CMP-014` through `FR-CMP-016`, `FR-CMP-018`, `FR-CMP-021`, `FR-CMP-028`
  - **Vulnerability / Flaw:** 11 out of 28 requirements in FRD 03 are retired or subsumed placeholders. Section 3.1 is entirely subsumed regex NLP text parsing requirements.
  - **Exploit / Failure Scenario:** Maintenance overhead and cognitive clutter for developers reading the specification.
  - **Remedy:** Consolidate the retired Section 3.1 placeholders into a concise archival callout and collapse subsumed requirements into the traceability matrix to reduce specification clutter.

- **[DA-CMP-008] Underdeveloped -EmitSchemas Generation Contract**
  - **Requirement Citations:** `FR-CMP-001`, `FR-CMP-022`
  - **Vulnerability / Flaw:** `FR-CMP-022` specifies that `-EmitSchemas` regenerates all JSON schema files, but provides no architectural details on the source of truth for schema generation (e.g. embedded C# models, PowerShell hash-tables, or template files).
  - **Exploit / Failure Scenario:** Unclear developer contract for how schemas are updated when AST definitions change.
  - **Remedy:** Document the exact generation mechanism and source templates used by `-EmitSchemas` in `compile_blueprints.ps1`.

## 3. Redundancy & De-Scoping Ledger
| Requirement ID | Fully Subsumed By | Rationale / Recommendation |
| :--- | :--- | :--- |
| `FR-CMP-008` – `FR-CMP-012` | FRD 02 & `FR-CMP-017` | Retired regex NLP requirements replaced by declarative AST overrides. |
| `FR-CMP-014` – `FR-CMP-016` | FRD 02 & `FR-CMP-017` | Replaced by declarative AST schema models in FRD 02. |
| `FR-CMP-018` | `FR-CMP-017` & FRD 02 Sec 4.2 | Bespoke mechanics covered by sidecar overrides registry and benchmark tests. |
| `FR-CMP-021` | `FR-CMP-005` & `FR-CMP-020` | Strict review gates enforced via CSV prototype flags and override sign-offs. |
| `FR-CMP-028` | `FR-CMP-001` & `FR-CMP-027` | Exit codes fully codified in CLI contract table and error taxonomy. |

## 4. Socratic Interrogation Register
1. *Why does the compiler maintain 11 retired requirement placeholders for legacy regex NLP parsers instead of archiving them in revision history and keeping the active requirement set clean?*
2. *If vanilla cards use the sentinel string '--' in authoring CSVs for human readability, why normalize them to an empty string in the compiled production bundle instead of preserving the explicit sentinel?*
