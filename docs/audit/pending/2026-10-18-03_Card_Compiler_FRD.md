# Product Manager Devil's Advocate Audit: Card Compiler & Blueprint Bundler FRD

- **Target Document:** [docs/frd/03_Card_Compiler_FRD.md](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/03_Card_Compiler_FRD.md)
- **Auditor:** PM Devil's Advocate (`pm-devils-advocate`)
- **Evaluation Date:** 2026-10-18
- **Status:** PENDING (Active review queue)
- **Verdict:** CONVERGED / PASS

## 1. Executive Summary & Red-Team Assessment

This audit evaluates `03_Card_Compiler_FRD.md` (FRD v1.1) specifically regarding requirement duplication, copycat definitions, and domain boundary compliance.

`03_Card_Compiler_FRD.md` canonically owns the parameterized PowerShell CLI (`compile_blueprints.ps1`), the 10-stage compilation pipeline, atomic `.tmp.$PID` staging, low-level RFC-4180 parsing, the sidecar overrides registry contract (`ast_overrides.json`), text-to-AST parity verification (`FR-CMP-019`, `FR-CMP-020`), JSON Schema Draft 2020-12 validation gates, RFC 8785 canonical hashing (`catalogChecksum`), and semantic dataset versioning (`cardDataVersion`).

Red-teaming confirms that in earlier revisions, FRD 03 attempted to specify natural-language rules text parsing (`FR-CMP-008`–`FR-CMP-012`, `FR-CMP-014`–`FR-CMP-016`, `FR-CMP-018`) which re-defined declarative AST modeling from `02_Card_Data_Pipeline_FRD.md`. However, these copycat requirements were previously flagged, retired, and formally marked as subsumed (`*(Subsumed by...)*`) under Direction B (Declarative Catalog Linking). Additionally, `FR-CMP-021` and `FR-CMP-028` were cleanly subsumed into `FR-CMP-005`, `FR-CMP-020`, `FR-CMP-001`, and `FR-CMP-027`. All remaining active requirements in FRD 03 define contracts specifically required for compiler execution.

## 2. Severity-Ordered Punch-List

### 🔴 Blocker / Critical
_None._

### 🟠 High
_None._

### 🟡 Medium
_None._

### 🔵 Low / Clarification
_None._

## 3. Redundancy & De-Scoping Ledger
| Requirement ID | Fully Subsumed By | Rationale / Recommendation |
| :--- | :--- | :--- |
| `FR-CMP-008`–`FR-CMP-012` | `FR-TXT-001`–`FR-TXT-004`, `FR-CMP-013`, `FR-CMP-017` | Retired regex NLP synthesizer subsumed by FRD 02 and sidecar overrides. |
| `FR-CMP-014`–`FR-CMP-016` | `FR-ABL-002`–`FR-ABL-006`, `FR-ABL-010`, `FR-CMP-017` | Retired AST synthesizers subsumed by declarative schemas and sidecar overrides. |
| `FR-CMP-018` | `FR-CMP-017`, FRD 02 Section 4.2 | Bespoke mechanics integration subsumed by sidecar overrides registry. |
| `FR-CMP-021` | `FR-CMP-005`, `FR-CMP-020` | Strict review gate subsumed by prototype flag rejection and override sign-off checks. |
| `FR-CMP-028` | `FR-CMP-001`, `FR-CMP-027` | Exit code contract subsumed into CLI interface and error taxonomy. |

## 4. Socratic Interrogation Register
_None — all historical copycat requirements in 03_Card_Compiler_FRD.md have been formally subsumed, and active requirements strictly fulfill compiler responsibilities._
