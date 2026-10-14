# Product Manager Devil's Advocate Audit: Card Compiler & Blueprint Ingestion FRD (FRD v1.0)

- **Target Document:** [docs/frd/Card_Compiler_FRD.md](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/Card_Compiler_FRD.md)
- **Auditor:** PM Devil's Advocate (`pm-devils-advocate`)
- **Evaluation Date:** 2026-10-14
- **Status:** PENDING (Active review queue)
- **Verdict:** REVISE WITH BLOCKERS

---

## 1. Executive Summary & Red-Team Assessment

The **Card Compiler & Blueprint Ingestion FRD (v1.0)** details a robust, multi-stage compilation pipeline (`compile_blueprints.ps1`) translating human-authored balance CSVs into an authoritative, cryptographically fingerprinted JSON catalog (`blueprints.json`). The inclusion of RFC 8785 canonical serialization, 256-bit SHA-256 fingerprinting, atomic staging file replacement, and strict enum validation demonstrates commendable engineering rigor.

However, deep red-team interrogation across the three devil's advocate lenses reveals critical systemic flaws:
1. **Architectural Parity Deadlock**: `FR-CMP-020` mandates that *every* compiled `AbilityDef` AST node must map to a recognizable printed clause in rules text, yet `FR-CMP-017` introduces `ast_overrides.json` specifically for cards whose mechanics cannot be parsed by regex. Because the compiler possesses no reverse Natural Language Generation (NLG) engine for bespoke override ASTs, Stage 7 parity checking will throw fatal `ERR_PARITY_ORPHAN_AST` errors on every single overridden card, making successful compilation mathematically impossible.
2. **Brittle NLP Architecture in PowerShell**: Implementing a full Natural Language Processing (NLP) tokenizer, lexical analyzer, and clause parser via PowerShell regular expressions (`FR-CMP-008` - `FR-CMP-016`) introduces massive delivery and operational risk. Minor prose variations, clause re-orderings, or punctuation quirks will cause silent miscompilations or catastrophic pipeline aborts.
3. **Pre-Commit Hook Latency Violation**: Compiling C# helpers in-memory on every compiler execution via `Add-Type` (`FR-CMP-024`) invokes `csc.exe` out-of-process, adding 1,500 to 3,500 ms of overhead. This directly violates the compiler latency budget (`FR-CMP-002`) and will make git pre-commit hooks intolerably sluggish for the development team.

**Recommendation:** Revise `FR-CMP-020` to formally exempt sidecar overrides from automated reverse AST-to-text NLP mapping, decouple complex card authoring from regex heuristics, and eliminate runtime `Add-Type` compilation in favor of pre-compiled or native BCL serialization.

---

## 2. Severity-Ordered Punch-List

### 🔴 Blocker / Critical

- **[DA-CMP-001] Reverse Parity Verification Deadlocks on Sidecar Overridden Cards**
  - **Requirement Citations:** `FR-CMP-020`, `FR-CMP-017`, `FR-CMP-019`, `FR-VAL-002`
  - **Vulnerability / Flaw:** `FR-CMP-020` (Reverse Parity Verification) dictates:
    *"Every `AbilityDef` in a blueprint's `abilities` array must correspond to a recognizable clause in `rulesText`... Any unrepresented AST node triggers compilation error `ERR_PARITY_ORPHAN_AST`."*
    Meanwhile, `FR-CMP-017` establishes that complex cards (*NO-AH-5*, *Samsara*, *Ambush Tactics*, *Blind Defile*) use `ast_overrides.json` *precisely because* the automated grammar engine cannot parse or generate their ASTs from text. While `FR-CMP-019` clause 2 exempts overrides from *forward* NLP parsing, `FR-CMP-020` contains no such exemption for *reverse* verification. The compiler has no algorithm to prove that an arbitrary, hand-crafted polymorphic AST structure in `ast_overrides.json` maps back to English prose.
  - **Exploit / Failure Scenario:**
    1. A designer configures *NO-AH-5* in `ast_overrides.json` with an activated ability executing `{ "primitive": "TUCK_UNDER", "hostTarget": "SELF" }`.
    2. The compiler runs Stage 6 and successfully binds the override AST.
    3. The compiler enters Stage 7 and executes `FR-CMP-020` reverse parity verification.
    4. The reverse parity engine inspects the AST's `TUCK_UNDER` primitive and scans the card's rules text: *"Ward\nR2: kill chosen allied non HQ asset and put it face down under this unit..."*.
    5. Because there is no hardcoded reverse-prose dictionary mapping `TUCK_UNDER` to *"put it face down under this unit"*, the engine flags `TUCK_UNDER` as an unrepresented "ghost" ability.
    6. Compilation halts with fatal error `ERR_PARITY_ORPHAN_AST`, failing every build that includes bespoke cards.
  - **Remedy:** Amend `FR-CMP-020` to explicitly scope automated reverse AST-to-text mapping to standard synthesized cards. For cards in `ast_overrides.json`, reverse parity must be satisfied strictly via the explicit designer verification block (`"parityVerified": true`, `verifiedBy`, `verifiedAt`) and CSV text string matching specified in `FR-CMP-017`.

---

### 🟠 High

- **[DA-CMP-002] High-Risk PowerShell Regex NLP Architecture for Context-Sensitive Rules Text**
  - **Requirement Citations:** `FR-CMP-008`, `FR-CMP-009`, `FR-CMP-010`, `FR-CMP-011`, `FR-CMP-012`, `FR-CMP-013`, `FR-CMP-014`
  - **Vulnerability / Flaw:** The specification tasks a PowerShell script (`compile_blueprints.ps1`) with acting as a full lexical analyzer, tokenizer, and grammar parser for English rules text using regular expressions. Card game rules text is inherently context-sensitive (e.g. conditional clauses, compound costs like `"T1, kill this unit:"`, modal branches, dynamic targeting filters). Regex-based grammar engines inevitably devolve into unmaintainable, brittle webs of pattern matching where slight changes in card wording break unrelated cards.
  - **Exploit / Failure Scenario:**
    1. A designer edits a card text from *"Deal 2 damage to chosen asset."* to *"Deal 2 damage to up to 1 chosen asset."*
    2. The regex in `FR-CMP-011` matches `chosen asset` but fails on `up to 1 chosen asset` due to unanticipated phrasing variation.
    3. The compiler emits `ERR_SYNTAX_UNPARSEABLE` (exit code 4), blocking the entire pre-commit hook and halting all engineering commits across the team until a tool programmer updates the PowerShell regex suite.
  - **Remedy:** Simplify Domain 3 & 4 scope. Explicitly limit automated regex synthesis to standalone keywords (`FR-CMP-013`), standard activated costs (`FR-CMP-010`), and simple single-clause triggers. Mandate that any card with composite logic, modal branches, or complex conditionals bypass regex parsing and reside directly in `ast_overrides.json` without failing the build.

- **[DA-CMP-003] Pre-Commit CLI Latency Violation from Dynamic In-Memory C# Compilation**
  - **Requirement Citations:** `FR-CMP-002`, `FR-CMP-024`, `FR-CMP-028`
  - **Vulnerability / Flaw:** `FR-CMP-024` dictates that RFC 8785 canonical serialization be performed via an embedded C# JCS helper compiled in-memory via `Add-Type`. In Windows PowerShell 5.1 (standard on many Windows developer workstations), `Add-Type` dynamically invokes `csc.exe` on disk, creating a 1,500 ms to 3,500 ms process startup penalty on every invocation. When added to CSV parsing, schema validation, and git hook execution, warm compilation for 200 cards will exceed the 5,000 ms SLA defined in `FR-CMP-002`.
  - **Exploit / Failure Scenario:**
    1. A developer commits a single-line balance change in `data/cards/maneuvers-animus.csv`.
    2. Git pre-commit hook invokes `compile_blueprints.ps1`.
    3. Windows PowerShell 5.1 initializes and invokes `Add-Type`, stalling for 3 seconds while `csc.exe` compiles the C# class.
    4. Combined with JSON Schema Draft 2020-12 evaluation across 200 cards, the pre-commit hook takes 12-15 seconds to execute.
    5. Developers routinely bypass the hook (`git commit --no-verify`), defeating the entire automated ingestion safety net.
  - **Remedy:** Replace dynamic in-memory `Add-Type` compilation with a pre-compiled standalone .NET assembly (DLL) or implement canonical JSON normalization using pure PowerShell/.NET BCL features (`System.Security.Cryptography.SHA256` and ordered dictionary key sorting).

---

### 🟡 Medium

- **[DA-CMP-004] Dataset Sanitation and Prototype Linter vs Compiler Discrepancy**
  - **Requirement Citations:** `FR-CMP-005`, `FR-VAL-003`, `FR-CMP-027`
  - **Vulnerability / Flaw:** `FR-CMP-005` specifies that the compiler completely ignores blank rows and permits cards with `needsRework == TRUE` or `currentProxy == TRUE` to compile normally in standard mode (setting metadata flags). However, `Card_Data_Pipeline_FRD` `FR-VAL-003` defines a strict "Must Fit All" validation policy where any row lacking complete ability logic is rejected. If `lint_cardtown.ps1` runs in pre-commit and rejects WIP rows while `compile_blueprints.ps1` allows them, the toolchain gives contradictory feedback.
  - **Remedy:** Harmonize `FR-CMP-005` with `FR-VAL-003`. Clarify that WIP/prototype rows are valid in development builds under both linter and compiler, but strictly blocked when `-Strict` is enabled.

- **[DA-CMP-005] Hardcoded Static Count in Master Review Register Gate**
  - **Requirement Citations:** `FR-CMP-021`, `FR-VAL-003`
  - **Vulnerability / Flaw:** `FR-CMP-021` states: *"When invoked with `-Strict`, the compiler verifies that all 52 cards listed in the Master Review Register have resolved design approvals."* Hardcoding the literal number "52" into a technical requirement is fragile; as cards are approved or new cards flagged, the register size changes, instantly obsoleting the requirement text.
  - **Remedy:** Remove the hardcoded "52" and rephrase to dynamically check that *"all cards flagged with unresolved review items in the Master Review Register have resolved design approvals"*.

---

### 🔵 Low / Clarification

- **[DA-CMP-006] Missing JSON Schema Definition for `ast_overrides.json` Sidecar Registry**
  - **Requirement Citations:** `FR-CMP-017`, `FR-CMP-022`
  - **Issue:** `FR-CMP-022` explicitly provisions `CardBlueprint.schema.json` and `AbilityDef.schema.json`, but does not provide an authoritative schema file for `ast_overrides.json` itself. Without a dedicated schema, syntax errors in the overrides envelope (e.g. missing `parityVerified`, malformed `rulesText`) are caught only via runtime compiler exceptions rather than IDE schema intellisense.
  - **Clarification Needed:** Author and commit `data/schemas/ast_overrides.schema.json` conforming to Draft 2020-12.

---

## 3. Redundancy & De-Scoping Ledger

| Requirement ID | Fully Subsumed By | Rationale / Recommendation |
| :--- | :--- | :--- |
| `FR-CMP-011` (Declarative Target Filter Synthesizer) | `FR-TXT-004` & `FR-TGT-001` | Targeting grammar and selector definitions are completely detailed in `Card_Data_Pipeline_FRD` and `Game_Rules_FRD`. Reference existing specs rather than duplicating filter conjunctions. |
| `FR-CMP-028` (Non-Zero Exit Code Contract) | `FR-CMP-001` & `FR-CMP-027` | Exit codes can be cleanly incorporated into the CLI Invocation Contract table (`FR-CMP-001`) alongside diagnostic categories. |

---

## 4. Socratic Interrogation Register

1. *Why are we attempting to write an entire Natural Language Processing compiler in PowerShell regexes when digital card games universally adopt structured DSLs or declarative schema authoring to avoid catastrophic card behavior bugs?*
2. *If reverse parity verification (`FR-CMP-020`) fails on every card in `ast_overrides.json` because regexes cannot map custom ASTs back to prose, how can this compiler ever pass a CI build?*
3. *Why should every developer pre-commit hook incur a multi-second C# compilation penalty via `Add-Type` when pre-compiled binaries or standard serializers are readily available?*
4. *What happens when a game designer introduces a new keyword or modal variant—are they blocked from playtesting until a programmer writes both an AST override and a new regex parser hook?*
