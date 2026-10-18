# Product Manager Devil's Advocate Audit: Card Compiler & Blueprint Bundler FRD (FRD v1.1)

- **Target Document:** [docs/frd/03_Card_Compiler_FRD.md](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/03_Card_Compiler_FRD.md)
- **Auditor:** PM Devil's Advocate (`pm-devils-advocate`)
- **Evaluation Date:** 2026-10-18
- **Status:** PENDING (Active review queue)
- **Verdict:** CONDITIONAL PASS

## 1. Executive Summary & Red-Team Assessment

`03_Card_Compiler_FRD.md` (v1.1) cleanly scopes `tools/importer/compile_blueprints.ps1` as a deterministic Catalog Linker, Validator, and RFC 8785 Checksum Bundler delegating CSV and AST schemas to `02_Card_Data_Pipeline_FRD.md`. Cross-checking its linking, validation, and hashing contracts against `01_Game_Rules_FRD.md` and `02_Card_Data_Pipeline_FRD.md` identifies **two High-severity cross-spec conflicts** and **two Medium-severity compiler ambiguities**:
1. **High Conflicts (`DA-CMP-001`, `DA-CMP-002`):** `FR-CMP-021` ("Strict Mode Review Gate") instructs the compiler to inspect a per-card review register in Domain 7 of `02_Card_Data_Pipeline_FRD.md` that no longer exists (and duplicates `FR-CMP-005` `needsRework` and `FR-CMP-020` `parityVerified`); `FR-CMP-013` includes `Unbowed` in its standalone parameterless keyword lookup list, conflicting with `FR-KWD-002a` and `FR-ABL-012` where `Unbowed: [Effect].` is a triggered ability preamble requiring a card-specific `EffectDef` AST.
2. **Medium Ambiguities (`DA-CMP-003`, `DA-CMP-004`):** `FR-CMP-024` and `FR-CMP-013` leave the exact JSON structure hashed for `catalogChecksum` (the `blueprints` array vs. outer envelope fields, vanilla `rulesText` normalization `""` vs. `"--"`, and `currentProxy`/`needsRework` inclusion) open to competing interpretations that would produce incompatible SHA-256 digests; `FR-CMP-017` and `FR-CMP-019` do not clarify whether hybrid cards (containing both a standalone keyword line and non-keyword ability text) require the standalone `KEYWORD_GRANT` node to be explicitly included inside `ast_overrides.json`.

*(Note: Game and rule balance design questions are excluded per audit scope.)*

## 2. Severity-Ordered Punch-List

### 🔴 Blocker / Critical
_None._

### 🟠 High
- **[DA-CMP-001] [Conflict / Redundancy] FR-CMP-021 References a Non-Existent Review Register in Domain 7 of FRD 02 and Duplicates FR-CMP-005 & FR-CMP-020**
  - **Requirement Citations:** `FR-CMP-021`, `FR-CMP-005`, `FR-CMP-020`, `FR-CMP-027`
  - **Vulnerability / Flaw:** `FR-CMP-021` ("Strict Mode Review Gate") requires the compiler under `-Strict` to check *"whether any card carries an unresolved review flag in Domain 7 of `02_Card_Data_Pipeline_FRD.md` and halts with `ERR_STRICT_UNRESOLVED_REVIEW` if any remain open."* However, Domain 7 of `02_Card_Data_Pipeline_FRD.md` (line 437) was refactored to remove the inline markdown review table and explicitly states that ongoing per-card review and parity status are tracked live via the CSV `needsRework` column (`FR-VAL-003`, already enforced by `FR-CMP-005` via `ERR_STRICT_PROTOTYPE_REJECTED`) and `data/schemas/ast_overrides.json` `parityVerified` (`FR-CMP-020`, already enforced via `ERR_OVERRIDE_UNVERIFIED`).
  - **Exploit / Failure Scenario:** Implementing `FR-CMP-021` as written requires parsing a markdown table in `02_Card_Data_Pipeline_FRD.md` that does not exist, while the actual review gates (`needsRework` and `parityVerified`) are already enforced by `FR-CMP-005` and `FR-CMP-020`.
  - **Remedy:** Subsume `FR-CMP-021` (and `ERR_STRICT_UNRESOLVED_REVIEW` in `FR-CMP-027`) into `FR-CMP-005` (`ERR_STRICT_PROTOTYPE_REJECTED` on `currentProxy == "true"` or `needsRework == "true"`) and `FR-CMP-020` (`ERR_OVERRIDE_UNVERIFIED` on `parityVerified != true`), removing the obsolete reference to Domain 7 of `02_Card_Data_Pipeline_FRD.md`.

- **[DA-CMP-002] [Conflict] FR-CMP-013 Includes Parameterized Trigger Preamble "Unbowed" in the Standalone Keyword Lookup Table**
  - **Requirement Citations:** `FR-CMP-013`, `FR-CMP-019`
  - **Vulnerability / Flaw:** `FR-CMP-013` bullet 2 states that for cards not present in `ast_overrides.json`, if every line in `rulesText` matches a standalone canonical keyword from `FR-ABL-012`—explicitly listing `Unbowed` alongside `Stalwart`, `Resistant N`, `Ward`, `Aggressive N`, `Guardian`, `Armor N`, `Versatile`, `Ethereal`, `Deploys Exhausted`, and `Berserk`—the compiler emits the corresponding `AbilityDef` directly from the `FR-ABL-012` lookup table. This conflicts with `FR-KWD-002a` (`01_Game_Rules_FRD.md`) and `FR-ABL-012` (`02_Card_Data_Pipeline_FRD.md`), where `Unbowed` is always templated as `"Unbowed: [Effect]."` and maps to a `TRIGGERED` ability (`triggerHook: "EVENT_TAKE_DAMAGE"`) requiring an arbitrary card-specific `EffectDef` payload that cannot be synthesized from a static keyword lookup table.
  - **Exploit / Failure Scenario:** A card with `"Unbowed: Scratch 1."` (*Hammerhand Brigands*) or `"Unbowed: Heal 2."` (*Macabre Dancers*) cannot be compiled without an `ast_overrides.json` entry (`FR-CMP-017`) because `Unbowed` alone does not define the triggered `EffectDef`.
  - **Remedy:** Remove `Unbowed` from the standalone keyword lookup list in `FR-CMP-013` bullet 2 so that standalone keyword auto-synthesis applies strictly to the 10 `KEYWORD_GRANT` keywords (`Stalwart`, `Resistant N`, `Ward`, `Aggressive N`, `Guardian`, `Armor N`, `Versatile`, `Ethereal`, `Deploys Exhausted`, `Berserk`), and explicitly state that all `"Unbowed: [Effect]."` cards and all Maneuver cards (including `"Scratch 1."` on *Growing Legion*) link their `AbilityDef` ASTs from `ast_overrides.json` (`FR-CMP-017`).

### 🟡 Medium
- **[DA-CMP-003] [Ambiguity] Ambiguous RFC 8785 Checksum Input Scope (FR-CMP-024) and Vanilla rulesText Normalization (FR-CMP-013)**
  - **Requirement Citations:** `FR-CMP-005`, `FR-CMP-013`, `FR-CMP-023`, `FR-CMP-024`, `FR-CMP-025`, `FR-CMP-026`
  - **Vulnerability / Flaw:** Three serialization details that directly affect the 256-bit `catalogChecksum` (`FR-CMP-025`) are underspecified or ambiguous:
    1. *Hashed Root Value*: `FR-CMP-024` states that before computing `catalogChecksum`, *"the compiler serializes the `blueprints` payload in accordance with RFC 8785"*. It is ambiguous whether SHA-256 is computed strictly over the RFC 8785 canonical JSON bytes of the `blueprints` **array** (`[ { ... }, ... ]`) or over a subset of the root envelope object (`FR-CMP-023`).
    2. *Vanilla `rulesText` String*: `FR-CMP-013` bullet 1 states that when a CSV row's `rulesText` is empty (`""`) or equals `"--"`, the compiler emits `abilities: []` and *"normalizes `rulesText` per `FR-DAT-007`"*. However, `FR-DAT-007` only specifies that `"--"` compiles into `abilities: []` without stating whether the `CardBlueprint.rulesText` string field in `blueprints.json` is serialized as `""` or `"--"`.
    3. *Prototype Metadata in Checksum*: `FR-CMP-005` states that `currentProxy` and `needsRework` serialize directly onto `CardBlueprint`, whereas `FR-CMP-026` states that non-mechanical metadata changes are `PATCH` versions that should not break session/replay compatibility, and `02_Benchmark_Card_Blueprints.md` omits `currentProxy` and `needsRework`.
  - **Exploit / Failure Scenario:** If the PowerShell compiler (`compile_blueprints.ps1`) and the Go headless engine (`FR-ENG-001`) differ on whether `catalogChecksum` hashes the `blueprints` array vs. envelope, or whether vanilla `rulesText` is `""` vs. `"--"`, their computed checksums will never match.
  - **Remedy:** Specify in `FR-CMP-013` and `FR-CMP-024` that: (1) `catalogChecksum` is computed strictly over the RFC 8785 canonical JSON bytes of the `blueprints` **array** (`[ ... ]`); (2) vanilla `rulesText` (`"--"` or `""`) normalizes to `""` (or `"--"`) in `CardBlueprint.rulesText`; and (3) clarify whether `currentProxy` and `needsRework` are included in or excluded from the hashed `CardBlueprint` objects.

- **[DA-CMP-004] [Ambiguity] Standalone Keyword Linking Responsibility on Hybrid Cards Present in ast_overrides.json**
  - **Requirement Citations:** `FR-CMP-013`, `FR-CMP-017`, `FR-CMP-019`, `FR-CMP-027`
  - **Vulnerability / Flaw:** `FR-CMP-013` applies to *"cards not present in `ast_overrides.json`"*, `FR-CMP-017` bullet 2 states that when a card's `id` exists in `ast_overrides.json` the compiler binds `abilities` from the registry entry, and `FR-CMP-019` bullet 2 requires that if a card's `rulesText` includes a standalone keyword line, its compiled `abilities` array must include the corresponding keyword AST node. For hybrid cards containing both a standalone keyword line and non-keyword rules text (e.g., *NO-AH-5* with `Ward`, *Piston Fist Prototype* with `Aggressive 9`, *Drifting Concept* with `Ethereal`), it is ambiguous whether `overrides[id].abilities` in `ast_overrides.json` must manually include the `KEYWORD_GRANT` node (and which error code in `FR-CMP-027` fires if it is omitted) or whether the compiler auto-merges standalone keyword lines with `overrides[id].abilities`.
  - **Exploit / Failure Scenario:** If `FR-CMP-017` replaces `abilities` wholesale with `overrides[id].abilities` while `FR-CMP-027` defines `ERR_PARITY_UNMAPPED_TEXT` as *"Non-keyword `rulesText` has no `ast_overrides.json` entry"*, a missing `KEYWORD_GRANT` node inside an `ast_overrides.json` entry has no clearly mapped diagnostic error code.
  - **Remedy:** Clarify in `FR-CMP-017`, `FR-CMP-019`, and `FR-CMP-027` that for any card in `ast_overrides.json`, `overrides[id].abilities` must contain the complete `abilities` array (including `KEYWORD_GRANT` nodes for any standalone keyword lines in `rulesText`), and that omitting a printed standalone keyword line's `KEYWORD_GRANT` node from `overrides[id].abilities` halts compilation with `ERR_PARITY_UNMAPPED_TEXT`.

### 🔵 Low / Clarification
_None._

## 3. Redundancy & De-Scoping Ledger
| Requirement ID | Fully Subsumed By | Rationale / Recommendation |
| :--- | :--- | :--- |
| `FR-CMP-021` (`ERR_STRICT_UNRESOLVED_REVIEW`) | `FR-CMP-005` (`ERR_STRICT_PROTOTYPE_REJECTED`) & `FR-CMP-020` (`ERR_OVERRIDE_UNVERIFIED`) | References a deleted markdown table in Domain 7 of FRD 02; per-card rework and AST review sign-off are already enforced via `needsRework` (`FR-CMP-005`) and `parityVerified` (`FR-CMP-020`). |

## 4. Socratic Interrogation Register
_None — game and rule balance design questions are excluded per audit scope; all specification conflicts, gaps, ambiguities, underdeveloped schemas, and redundancies are itemized with concrete remedies in Sections 2 and 3._
