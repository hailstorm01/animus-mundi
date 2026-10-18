# Product Manager Devil's Advocate Audit: Card Data Pipeline & Ability Authoring FRD

- **Target Document:** [docs/frd/02_Card_Data_Pipeline_FRD.md](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/02_Card_Data_Pipeline_FRD.md)
- **Auditor:** PM Devil's Advocate (`pm-devils-advocate`)
- **Evaluation Date:** 2026-10-18
- **Status:** PENDING (Active review queue)
- **Verdict:** CONDITIONAL PASS

## 1. Executive Summary & Red-Team Assessment

This audit evaluates `02_Card_Data_Pipeline_FRD.md` (FRD v1.1) specifically focusing on requirement ownership boundaries and whether requirements re-define concepts canonically owned by other specifications without adding unique domain functionality.

`02_Card_Data_Pipeline_FRD.md` is the canonical authority for Card Town's 3-tier authoring lifecycle (CSV balance sheets in `data/cards/`, Google Sheets mirror, production schema), target enum enforcement, declarative ability AST schemas (`AbilityDef`), event hook taxonomy (`EVENT_*` and `WOULD_*`), 26 atomic effect primitives (`FR-PRM-001`), and MTG-style card text standards (`FR-TXT-001`–`FR-TXT-006`).

However, red-teaming reveals that Section 1.1 retains two vestigial compilation requirements (`FR-DAT-003` and `FR-DAT-004`) written prior to the formalization of `03_Card_Compiler_FRD.md` and `04_Client_Engine_Protocol_Networking_FRD.md`. These requirements re-define atomic multi-sheet compilation invariants, RFC 8785 canonical checksum hashing, SemVer versioning, wire handshake verification, and replay header archiving—responsibilities that are now canonically owned, parameterized, and enforced by FRD 03 and FRD 04. In addition, `FR-TXT-004` includes rules text re-defining two-stage dynamic filter verification canonically owned by `01_Game_Rules_FRD.md` (`FR-TGT-003`).

## 2. Severity-Ordered Punch-List

### 🔴 Blocker / Critical
_None._

### 🟠 High
_None._

### 🟡 Medium
- **[DA-DAT-004] Vestigial Compilation, Checksum & Wire Handshake Requirements in FR-DAT-003 and FR-DAT-004**
  - **Requirement Citations:** `FR-DAT-003`, `FR-DAT-004`
  - **Vulnerability / Flaw:** `FR-DAT-003` (Blueprint Compilation & Asset Bundling) and `FR-DAT-004` (Engine Version Binding & Replay Compatibility) specify atomic compilation invariants, schema abort gates, RFC 8785 256-bit `catalogChecksum` computation, SemVer `cardDataVersion`, wire session handshake validation, and replay header archiving. These responsibilities are canonically owned and fully specified in `03_Card_Compiler_FRD.md` (`FR-CMP-002`, `FR-CMP-003`, `FR-CMP-024`, `FR-CMP-025`, `FR-CMP-026`, `FR-CMP-027`), `04_Client_Engine_Protocol_Networking_FRD.md` (`FR-SES-001`, `FR-SPC-002`), and `05_Headless_Game_Engine_FRD.md` (`FR-ENT-001`, `FR-ENG-038`). Retaining these duplicate requirements in FRD 02 bloats the authoring specification and creates synchronization drift risk across documents without adding any functionality needed for card authoring or AST modeling.
  - **Exploit / Failure Scenario:** If the compiler team refactors blueprint staging, exit codes, or checksum canonicalization in FRD 03, FRD 02's `FR-DAT-003` and `FR-DAT-004` will drift out of sync, presenting competing definitions of compilation and versioning contracts.
  - **Remedy:** Formally mark `FR-DAT-003` and `FR-DAT-004` as subsumed by `FR-CMP-002`, `FR-CMP-003`, `FR-CMP-025`, `FR-CMP-026`, `FR-SES-001`, and `FR-SPC-002`, retaining requirement ID placeholders with traceability pointers to their canonical owners.

### 🔵 Low / Clarification
- **[DA-DAT-005] Two-Stage Target Verification Phrasing in Rules Text Standard FR-TXT-004**
  - **Requirement Citations:** `FR-TXT-004`
  - **Vulnerability / Flaw:** `FR-TXT-004` is titled *"Conjunction Targeting Phrasing & Two-Stage Verification"*, and bullet 2 explicitly defines two-stage dynamic filter invalidation. However, dynamic targeting verification is a core game rule canonically owned by `01_Game_Rules_FRD.md` (`FR-TGT-003`). `FR-TXT-004`'s actual function in FRD 02 is solely to enforce the authoring syntax rule that card rules text must use the verb *"chosen"* rather than *"target"*.
  - **Exploit / Failure Scenario:** Editors updating targeting invalidation mechanics might edit FRD 02 instead of or in addition to FRD 01, fragmenting game rules governance.
  - **Remedy:** Retitle `FR-TXT-004` to *"Conjunction Targeting Phrasing (`chosen`)"* and replace bullet 2 with an explicit cross-reference to `FR-TGT-003` without restating the game rule.

## 3. Redundancy & De-Scoping Ledger
| Requirement ID | Fully Subsumed By | Rationale / Recommendation |
| :--- | :--- | :--- |
| `FR-DAT-003` | `FR-CMP-002`, `FR-CMP-003`, `FR-CMP-027`, `FR-ENT-001` | Atomic compilation and memory bundling belong canonically to FRD 03 and FRD 05. |
| `FR-DAT-004` | `FR-CMP-024`–`FR-CMP-026`, `FR-SES-001`, `FR-SPC-002` | Checksum computation, SemVer versioning, and wire handshakes belong canonically to FRD 03 and FRD 04. |

## 4. Socratic Interrogation Register
1. *Should `02_Card_Data_Pipeline_FRD.md` restrict its scope strictly to authoring datasets, CSV normalization, declarative AST schemas, and text standards, delegating all build-time CLI and runtime ingestion contracts entirely to `03_Card_Compiler_FRD.md`?*
