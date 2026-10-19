# Product Manager Devil's Advocate Audit: Card Compiler FRD (Master FRD v1.2)

- **Target Document:** [docs/frd/03_Card_Compiler_FRD.md](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/03_Card_Compiler_FRD.md)
- **Auditor:** PM Devil's Advocate (`pm-devils-advocate`)
- **Evaluation Date:** 2026-10-19
- **Status:** PENDING (Active review queue)
- **Verdict:** CONDITIONAL PASS

## 1. Executive Summary & Red-Team Assessment

The Card Compiler specification (`03_Card_Compiler_FRD.md`) successfully completed its pure Go architecture pivot, retiring legacy PowerShell scripts and establishing strict CLI contracts, atomic file staging, RFC 8785 canonical hashing, and rigorous text-to-AST parity gates. However, a major omission exists in the error taxonomy table: `FR-CMP-001` specifies that Exit Code 4 triggers on "blueprint schema failure", yet `FR-CMP-027` completely lacks an error code for `CardBlueprint.schema.json` validation failure (such as `ERR_BLUEPRINT_SCHEMA_INVALID`). In addition, the document exhibits flag naming drift between POSIX flags (`--strict`, `--emit-schemas`) and legacy PascalCase text references (`-Strict`, `-EmitSchemas`), and the standalone execution behavior of `--emit-schemas` remains underdeveloped.

## 2. Severity-Ordered Punch-List

### 🔴 Blocker / Critical
_None._

### 🟠 High
- **[DA-CMP-009] Omission of Blueprint Schema Validation Error Code in Error Taxonomy**
  - **Requirement Citations:** `FR-CMP-001`, `FR-CMP-022`, `FR-CMP-027`
  - **Vulnerability / Flaw:** In `FR-CMP-001` (Process Exit Code Contract, line 110), Exit Code 4 is explicitly defined to trigger on: *"Override schema failure, orphan override, text mismatch, or **blueprint schema failure** (`ERR_OVERRIDE_*`, `ERR_ORPHAN_OVERRIDE`)"*. Furthermore, `FR-CMP-022` mandates validating compiled output against `CardBlueprint.schema.json`. However, in the canonical error taxonomy table of `FR-CMP-027` (lines 280-307), there is **no error code** defined for blueprint schema validation failure. The table defines `ERR_OVERRIDE_SCHEMA_INVALID` for `ast_overrides.json`, but completely omits an error identifier for `CardBlueprint.schema.json` validation failure (e.g. `ERR_BLUEPRINT_SCHEMA_INVALID`).
  - **Exploit / Failure Scenario:** When compiled `blueprints.json` fails `CardBlueprint.schema.json` validation, the compiler lacks a canonical error code to emit, leading to inconsistent error logging or miscategorization under `ERR_SCHEMA_IO` (which is an I/O Exit Code 1 error, directly contradicting Exit Code 4).
  - **Remedy:** Add `ERR_BLUEPRINT_SCHEMA_INVALID` to the error taxonomy in `FR-CMP-027` under the "Linking & Schemas" category mapped to Exit Code 4.

### 🟡 Medium
- **[DA-CMP-010] CLI Parameter Syntax Drift (POSIX `--flag` vs PascalCase `-Flag`)**
  - **Requirement Citations:** `FR-CMP-001`, `FR-CMP-005`, `FR-CMP-019`, `FR-CMP-021`, `FR-CMP-022`, `FR-CMP-023`
  - **Vulnerability / Flaw:** Section 1.1 (`FR-CMP-001`, lines 79-90) explicitly updated the CLI contract to POSIX/GNU lowercase kebab-case flags (`--source-dir`, `--output-dir`, `--strict`, `--emit-schemas`, `--minify`). However, throughout the rest of the document (lines 144, 220, 229, 236, 260), the prose repeatedly refers to legacy PowerShell single-dash PascalCase flags: `-Strict`, `-EmitSchemas`, and `-Minify`.
  - **Exploit / Failure Scenario:** Developers or CI/CD pipelines invoking `compile-blueprints -Strict` or `-EmitSchemas` on Linux/macOS may encounter unrecognized flag errors if the Go CLI flag parser expects `--strict` or `--emit-schemas`.
  - **Remedy:** Standardize all prose citations throughout FRD 03 to canonical POSIX flags (`--strict`, `--emit-schemas`, `--minify`).

- **[DA-CMP-011] Underdeveloped `--emit-schemas` Schema Generation Contract**
  - **Requirement Citations:** `FR-CMP-022`
  - **Vulnerability / Flaw:** `FR-CMP-022` specifies that passing `--emit-schemas` exports `CardBlueprint.schema.json`, `AbilityDef.schema.json`, and `ast_overrides.schema.json` from Go AST structs. However, it does not specify whether `--emit-schemas` can be run standalone (without requiring valid CSV files in `--source-dir`) or how validation behaves if both `--emit-schemas` and `--strict` are passed.
  - **Exploit / Failure Scenario:** A developer runs `compile-blueprints --emit-schemas` to bootstrap schema files in a fresh workspace, but the compiler aborts with `ERR_CSV_IO` because CSV files have not yet been placed or compiled.
  - **Remedy:** Clarify in `FR-CMP-022` that `--emit-schemas` operates as an independent utility mode that exports schemas and exits with code 0 without executing the CSV ingestion or linking pipeline.

### 🔵 Low / Clarification
_None._

## 3. Redundancy & De-Scoping Ledger
| Requirement ID | Fully Subsumed By | Rationale / Recommendation |
| :--- | :--- | :--- |
| `FR-CMP-008` – `FR-CMP-012` | `FR-TXT-001`–`FR-TXT-006`, `FR-ABL-001`–`FR-ABL-012`, `FR-CMP-013`, `FR-CMP-017` | Retired regex natural-language parsers are fully subsumed by declarative AST linking. |
| `FR-CMP-014` – `FR-CMP-016` | `FR-ABL-002`–`FR-ABL-010`, `FR-CMP-017` | Ability synthesis is handled declaratively in `ast_overrides.json`. |
| `FR-CMP-018` | `FR-CMP-017` & FRD 02 Section 4.2 | Bespoke mechanics modeling is specified in `ast_overrides.json` and benchmark catalog. |
| `FR-CMP-021` | `FR-CMP-005` & `FR-CMP-020` | Strict mode review gates are fully enforced by `currentProxy` / `needsRework` and `parityVerified`. |
| `FR-CMP-028` | `FR-CMP-001` & `FR-CMP-027` | Exit code table and error taxonomy are fully codified in `FR-CMP-001` and `FR-CMP-027`. |

## 4. Socratic Interrogation Register
1. *If `cmd/compile-blueprints` is an offline build-time code and schema generator, why does it maintain runtime JSON Schema Draft 2020-12 validation rather than relying on Go native static type-checking and struct deserialization, which runs 10x faster and requires no external schema files?*
