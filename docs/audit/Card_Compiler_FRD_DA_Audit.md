# Product Manager Devil's Advocate Audit: Card Compiler & Blueprint Ingestion FRD (FRD v1.0)

- **Target Document:** [Card_Compiler_FRD.md](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/Card_Compiler_FRD.md)
- **Auditor:** PM Devil's Advocate (`pm-devils-advocate`)
- **Date:** 2026-10-13
- **Verdict:** REVISE WITH BLOCKERS

---

## 1. Executive Summary & Red-Team Assessment

The **Card Compiler & Blueprint Ingestion FRD (v1.0)** specifies an automated, multi-stage compilation pipeline (`compile_blueprints.ps1`) to ingest RFC-4180 balance spreadsheets, synthesize declarative ability ASTs, enforce schema validation, and emit cryptographically fingerprinted JSON blueprints (`blueprints.json`). The multi-stage pipeline, atomic staging rollback guarantee (`FR-CMP-003`), and explicit error taxonomy (`FR-CMP-027`) demonstrate high engineering discipline.

However, red-teaming reveals severe architectural contradictions and delivery risks:
1. **The Parity Verification Paradox**: The specification justifies the sidecar overrides registry (`FR-CMP-017`) on the premise that complex natural language text cannot be reliably parsed. Yet `FR-CMP-019` simultaneously mandates an automated forward parity checker that must map every semantic element of printed text to AST nodes. The compiler cannot verify what it cannot parse.
2. **Dual-Maintenance Desynchronization**: Isolating bespoke mechanics into `ast_overrides.json` creates a fragile dual-source model. When designers modify CSV text or numbers, the sidecar JSON remains unchanged, introducing silent desynchronization bugs between printed rules and compiled abilities.
3. **PowerShell 5.1 RFC 8785 Canonicalization Feasibility**: Mandating strict RFC 8785 canonical JSON serialization (`FR-CMP-024`) within PowerShell 5.1 is highly problematic; native `ConvertTo-Json` does not sort keys lexicographically or adhere to JCS formatting, requiring a costly custom serialization engine.
4. **Unrealistic 2,000ms Performance SLA**: Requiring a 10-stage PowerShell script performing multi-sheet CSV ingestion, regex tokenization, JSON Schema Draft 2020-12 validation, and canonical hashing to execute in under 2,000 ms (`FR-CMP-002`) sets a brittle performance trap.

**Recommendation:** Resolve the parity verification paradox, simplify the sidecar workflow to prevent dual-source drift, and relax rigid PowerShell performance assertions before hardening into code.

---

## 2. Severity-Ordered Punch-List

### 🔴 Blocker / Critical

- **[DA-RULES-001] The Parity Verification Paradox in Two-Tier Architecture**
  - **Requirement Citations:** Architectural Decision 1, `FR-CMP-017`, `FR-CMP-019` (`ERR_PARITY_UNMAPPED_TEXT`)
  - **Vulnerability / Flaw:** In Architectural Decision 1, the document states: *"highly complex mechanics (NO-AH-5, Samsara, Ambush Tactics) cannot be reliably parsed with naive regular expressions without introducing extreme parser fragility."* This is the foundational rationale for creating `ast_overrides.json`.
    However, `FR-CMP-019` (Forward Parity Verification) simultaneously mandates:
    *"The compiler inspects every card's printed rulesText and verifies that every semantic element corresponds to an active AST node... Any printed rule clause that fails to map to an AST node triggers compilation error ERR_PARITY_UNMAPPED_TEXT."*
  - **Exploit / Failure Scenario:**
    1. If the compiler's parser is not capable of understanding complex rules clauses like *Samsara*'s conditional maneuver discard damage or *Ambush Tactics*' 4-mode selection, how can the parity checker verify that every semantic element in the text maps 1-to-1 to the override AST?
    2. If the parity checker relies on naive keyword matching, it offers illusory verification (e.g. seeing the word "maneuver" and assuming the conditional trigger is sound).
    3. If the parity checker performs deep syntactic and semantic analysis to verify the mapping, then the compiler possesses the very NLP capability that Decision 1 declared impractical!
    4. This circular contradiction means either `ast_overrides.json` is redundant, or `FR-CMP-019` will throw false-positive `ERR_PARITY_UNMAPPED_TEXT` errors on every bespoke card with an override.
  - **Remedy:** Explicitly exempt cards registered in `ast_overrides.json` from automated lexical forward-parity parsing. For overridden cards, verify only top-level keyword presence and schema conformance, replacing automated semantic mapping with manual designer sign-off in the review register.

- **[DA-RULES-002] Dual-Maintenance Desynchronization Between CSVs and Sidecar Overrides**
  - **Requirement Citations:** `FR-CMP-017`, `FR-CMP-018`, `FR-DAT-001`
  - **Vulnerability / Flaw:** When a card's ability is defined in `ast_overrides.json`, the CSV `rulesText` column is ignored for ability generation (`FR-CMP-017` states: *"the compiler binds the abilities array from ast_overrides.json, bypassing automated text synthesis"*).
  - **Exploit / Failure Scenario:**
    1. A designer balances *NO-AH-5* in `units-and-structures-soulless.csv`, changing the tuck cost from `R2` to `R1` and editing `rulesText` to `"R1: kill chosen allied non HQ asset and tuck it..."`.
    2. The designer commits the CSV or synchronizes from Google Drive.
    3. The designer does not edit `data/schemas/ast_overrides.json` (either forgetting it exists or lacking JSON skills).
    4. The compiler runs. Because *NO-AH-5* has an entry in `ast_overrides.json`, it binds the stale AST where the cost is still `R2`.
    5. The client displays printed card text saying `R1`, but the game engine executes the ability requiring `R2`.
    6. Players experience a game-breaking desynchronization between what the card text claims and how the engine resolves it.
  - **Remedy:** Enforce that `compile_blueprints.ps1` extracts printed costs (`R1`, `T1`, `exhaust`) from the CSV text even for overridden cards and validates them against the `costs` array in `ast_overrides.json`, failing with `ERR_OVERRIDE_COST_MISMATCH` if numeric or cost divergence is detected.

---

### 🟠 High

- **[DA-SCOPE-001] Pure PowerShell 5.1 Implementation of RFC 8785 Canonical JSON**
  - **Requirement Citations:** `FR-CMP-024`, `FR-CMP-025`, `FR-DAT-004`
  - **Vulnerability / Flaw:** `FR-CMP-024` mandates that the compiler canonicalize the blueprint catalog strictly in accordance with **RFC 8785 (JSON Canonicalization Scheme - JCS)**: lexicographical UTF-16 key sorting, whitespace omission, no trailing zeroes on numbers, and UTF-8 encoding without BOM.
  - **Assessment:**
    PowerShell 5.1's built-in `ConvertTo-Json` cmdlet completely fails to satisfy RFC 8785:
    - It does not sort object keys alphabetically.
    - It introduces inconsistent whitespace and indentation that varies across OS updates.
    - It has known depth serialization bugs (default depth 2).
    - It formats floating point numbers with regional culture settings.
    To satisfy RFC 8785 in PowerShell 5.1 without third-party compiled binaries, the script author must implement a custom recursive JSON serializer and sorting algorithm in PowerShell. This introduces high implementation complexity and potential performance bottlenecks.
  - **Remedy:** Provide a standardized C# / .NET snippet embedded via `Add-Type` in PowerShell, or utilize a standardized Node/Python helper, rather than writing a custom JCS serializer in pure PowerShell script.

- **[DA-SCOPE-002] Unrealistic 2,000ms Execution SLA for PowerShell Pipeline**
  - **Requirement Citations:** `FR-CMP-002` (Stage 10)
  - **Vulnerability / Flaw:** The specification states: *"Total compilation time for 200 cards must not exceed 2,000 milliseconds on standard hardware."*
  - **Assessment:**
    In a typical CI environment running Windows PowerShell 5.1, parsing 10 CSV files via `Import-Csv`, performing regex matching across 200+ cards, validating deep JSON schemas against Draft 2020-12, sorting hundreds of nested objects for RFC 8785 hashing, and writing multiple JSON files will easily exceed 2,000 ms. Setting a hard execution ceiling in an FRD creates an immediate non-functional requirement failure.
  - **Remedy:** Relax the execution SLA to a pragmatic target (e.g. `< 5,000 ms` in PowerShell 7+, `< 10,000 ms` in Windows PowerShell 5.1), and explicitly state that the SLA applies to warm execution runs excluding shell cold-start initialization.

---

### 🟡 Medium

- **[DA-XP-001] Truncated 64-Bit Checksum Discards 75% of SHA-256 Entropy**
  - **Requirement Citations:** `FR-CMP-025`, Architectural Decision 2
  - **Vulnerability / Flaw:** The specification computes standard SHA-256 (256 bits) and then truncates it to the leading 64 bits (16 hex characters).
  - **Assessment:** Truncating a 256-bit hash to 64 bits increases birthday collision probability while saving only 24 bytes of string length in a JSON file. Modern network packets and JSON files comfortably accommodate a full 64-character SHA-256 hex string. Truncation adds arbitrary substring logic across all consuming tools (Godot, C#, PowerShell, Python) without meaningful storage or bandwidth benefits.
  - **Remedy:** Retain the full standard 256-bit SHA-256 hex string (64 characters) for `catalogChecksum`.

- **[DA-SCOPE-003] Redundant Requirement Duplications Across Compiler and Data Pipeline FRDs**
  - **Requirement Citations:** `FR-CMP-006` vs `FR-DAT-002`, `FR-CMP-013` vs `FR-ABL-012`
  - **Vulnerability / Flaw:** Multiple sections of `Card_Compiler_FRD.md` duplicate requirements defined in `Card_Data_Pipeline_FRD.md` almost word-for-word (e.g. slug derivation algorithm in `FR-CMP-006` vs `FR-DAT-002`, keyword AST table in `FR-CMP-013` vs `FR-ABL-012`).
  - **Assessment:** Redundant requirement duplication clutters the specification suite and risks cross-document inconsistency when updates are made.
  - **Remedy:** Cross-reference canonical requirements in `Card_Data_Pipeline_FRD.md` rather than re-declaring them.

---

### 🔵 Low / Clarification

- **[DA-CLARIFY-001] Handling of Trailing Commas in Subtypes Column**
  - **Requirement Citations:** `FR-CMP-007`
  - **Issue:** If a designer writes `"Infantry, Scratch Company,"` with a trailing comma in the CSV, does `FR-CMP-007` produce an empty subtype `""` in the array or strip it cleanly?
  - **Clarification Needed:** Explicitly mandate that empty tokens resulting from trailing or multiple consecutive commas are filtered out during subtype splitting.

- **[DA-CLARIFY-002] Staging File Collisions in Concurrent CI Builds**
  - **Requirement Citations:** `FR-CMP-003`
  - **Issue:** `FR-CMP-003` specifies writing to a fixed temporary file `<OutputDir>/blueprints.json.tmp`. If two parallel test runners or CI workers execute compilation simultaneously on the same workspace, they will collide on `blueprints.json.tmp`.
  - **Clarification Needed:** Use a process-unique temporary staging filename (e.g. `blueprints.json.tmp.<PID>`).

---

## 3. Redundancy & De-Scoping Ledger

| Requirement ID | Fully Subsumed By | Rationale / Recommendation |
| :--- | :--- | :--- |
| `FR-CMP-006` (Deterministic Slug Derivation) | `FR-DAT-002` (Card Data Pipeline FRD) | Exact algorithmic duplicate. Reference `FR-DAT-002` directly. |
| `FR-CMP-013` (Canonical Keyword AST Synthesizer) | `FR-ABL-012` (Card Data Pipeline FRD) | 100% duplicate keyword mapping table. Reference `FR-ABL-012`. |
| `FR-CMP-022` (JSON Schema Conformance) | `FR-DAT-005` & Section 8.1 (Master FRD) | Duplicates schema requirements. Collapse into cross-reference. |

---

## 4. Socratic Interrogation Register

1. *If the natural language rules text parser cannot parse complex cards like Samsara or Ambush Tactics, how can the compiler's forward parity checker possibly verify that the printed text maps 1-to-1 to the override AST?*
2. *When a game designer modifies card numbers in the balance CSV, what guarantees that the sidecar `ast_overrides.json` file is updated to match, preventing severe text-versus-engine desyncs?*
3. *Why are we truncating a cryptographically secure 256-bit SHA-256 hash to 64 bits when a full hex string costs only 24 extra bytes in a file that is already hundreds of kilobytes?*
4. *Can pure Windows PowerShell 5.1 realistically perform full RFC 8785 canonical JSON sorting and Draft 2020-12 schema validation across 200 cards in under 2,000 milliseconds?*
5. *Why is the compiler written in PowerShell if the game client is Godot (C#/GDScript) and the game server is dedicated containers—would compiling in C# or Python provide better cross-platform maintainability and performance?*
