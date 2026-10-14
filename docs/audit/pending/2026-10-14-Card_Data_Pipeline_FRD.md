# Product Manager Devil's Advocate Audit: Card Data Pipeline & Ability Authoring FRD (FRD v1.0)

- **Target Document:** [docs/frd/Card_Data_Pipeline_FRD.md](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/Card_Data_Pipeline_FRD.md)
- **Auditor:** PM Devil's Advocate (`pm-devils-advocate`)
- **Evaluation Date:** 2026-10-14
- **Status:** PENDING (Active review queue)
- **Verdict:** REVISE WITH BLOCKERS

---

## 1. Executive Summary & Red-Team Assessment

The **Card Data Pipeline & Ability Authoring FRD (v1.0)** establishes the authoritative bridge between RFC-4180 balance spreadsheets, declarative ability Abstract Syntax Trees (ASTs), MTG-aligned rules text grammar, and dataset validation. The standardization of 127 card texts, elimination of shorthand aliases in CSVs (`FR-DAT-006`), and formalization of deterministic slug derivation (`FR-DAT-002`) are significant steps forward in project stability.

However, a rigorous red-team interrogation reveals critical vulnerabilities in tooling architecture, schema design, and production readiness:
1. **Destructive Tooling Workflow**: `FR-DAT-001` designates Google Sheets as the designers' workbench for balance exploration, yet `publish_card_sheets.ps1` operates strictly as a one-way export from Git CSVs that *unconditionally overwrites Google Sheets*. Because no reverse CSV importer exists (`tools/importer/` contains only an empty `.gitkeep`), running the publish tool will permanently destroy unsynced designer work in Google Drive.
2. **Pseudo-Declarative AST Contradiction**: Despite asserting that declarative ASTs eliminate engine recompilation (`Design Rationale DA-XP-001`), `FR-ABL-007` hardcodes one-off card mechanics directly into core schema enums (`COUNT_TUCKED_CARDS` for *NO-AH-5*, `COUNT_DISCARD_CARDS` for *All-Out War*, `COUNT_STRUCTURES_REQUISITION_YIELD` for *Consuming Earth*). This is brittle ad-hoc hacking disguised as declarative architecture.
3. **Unresolved Core Rules Ambiguities in Production Register**: Section 7.2 leaves critical targeting and trigger mechanics (*Steel Rain*, *Rift*, *Daggerspine Rangers*) open as unresolved designer review items, leaving engine programmers to guess at intended rules behavior.

**Recommendation:** Halt unconditional Google Sheets publishing until a safe bidirectional diff/import tool exists, refactor `ValueToken` into composable query primitives, and resolve all pending core card mechanics before freezing the pipeline.

---

## 2. Severity-Ordered Punch-List

### 🔴 Blocker / Critical

- **[DA-DAT-001] One-Way Google Sheets Pipeline Destroys Remote Game Designer Work**
  - **Requirement Citations:** `FR-DAT-001`, Domain 1
  - **Vulnerability / Flaw:** `FR-DAT-001` establishes a three-tier authority model where local Git CSVs are Tier 1 (Authoritative Truth) and Google Sheets are Tier 2 (Designer Inspection & Experimentation Mirror). The specification dictates:
    *"tools/sheets/publish_card_sheets.ps1 publishes local repository CSV content to Google Sheets. It treats local CSVs as authoritative input and unconditionally updates Google Sheets to reflect Git state without depending on or checking remote Sheet modifications."*
    Crucially, the repository contains **no automated importer** to pull Google Sheets edits back into Git CSVs (`tools/importer/` is completely empty).
  - **Exploit / Failure Scenario:**
    1. A game designer spends an entire afternoon rebalancing mana curves, adjusting stats, and tweaking keywords directly in the Google Sheets workbook.
    2. An engineer fixes a minor typo in a local CSV file and executes `publish_card_sheets.ps1` (or CI executes it automatically).
    3. The script unconditionally blasts the local Git state across all Google Sheets tabs.
    4. All of the designer's uncommitted balance work in Google Drive is permanently overwritten and lost with zero backup, rollback history, or collision warning.
  - **Remedy:** Mandate that `publish_card_sheets.ps1` perform a pre-publish freshness check (comparing Google Drive last-modified timestamps against local Git commit timestamps) and abort if remote edits exist. Furthermore, prioritize authoring a companion `import_card_sheets.ps1` script to safely ingest remote designer edits with a visual diff confirmation.

---

### 🟠 High

- **[DA-DAT-002] Ad-Hoc Hardcoded Enums in ValueToken AST Break Declarative Architecture**
  - **Requirement Citations:** `FR-ABL-007`, `FR-ABL-001`, `Design Rationale (DA-XP-001)`
  - **Vulnerability / Flaw:** `FR-ABL-007` defines `ValueToken.property` with the following enum:
    ```json
    "property": { 
      "type": "string", 
      "enum": [
        "COMBAT_POWER", "CURRENT_HEALTH", "MAX_HEALTH", 
        "COUNT_TUCKED_CARDS", "COUNT_DISCARD_CARDS", 
        "COUNT_STRUCTURES_REQUISITION_YIELD", "COUNT_STRUCTURES_TACTICS_YIELD"
      ] 
    }
    ```
    This directly contradicts the non-negotiable architectural justification in `Design Rationale (DA-XP-001)`, which claims that a declarative AST ensures *"long-term extensibility without engine recompilation"*. Every time designers invent a card that counts game objects (e.g. *NO-AH-5* tucked cards, *All-Out War* discarded cards, *Consuming Earth* structure yield), engineers must add a bespoke enum value, update schemas, and recompile engine match resolution code.
  - **Exploit / Failure Scenario:**
    1. Set 2 designers design a card that scales power based on the number of cards in the opponent's discard pile, or the number of damaged units in the Main Theater.
    2. The declarative AST cannot represent this effect because `ValueToken.property` does not contain `COUNT_OPPONENT_DISCARD` or `COUNT_DAMAGED_MAIN_THEATER_UNITS`.
    3. The card cannot compile (`ERR_ENUM_INVALID`), blocking designer testing until engine engineers write custom backend code.
  - **Remedy:** Replace bespoke card-specific property enums with generic composable query primitives:
    ```json
    {
      "type": "QUERY",
      "query": {
        "zone": "ATTACHED | DISCARD | BATTLEFIELD",
        "filter": { ... FilterDef ... },
        "aggregate": "COUNT | SUM_ATTRIBUTE",
        "targetAttribute": "REQUISITION_YIELD | POWER"
      }
    }
    ```

- **[DA-DAT-003] Unresolved Mechanical Ambiguities in Master Review Register Threaten Engine Implementation**
  - **Requirement Citations:** `FR-TXT-004`, `FR-ABL-011`, Section 7.2
  - **Vulnerability / Flaw:** Section 7.2 catalogs 52 cards requiring designer sign-off, but several include fundamental rules and architecture questions that directly dictate backend schema design:
    - *Steel Rain* (#487): *"Confirm if 'chosen asset' is selected once or independently per card discarded."* If chosen once, it compiles as a single damage event with a dynamic multiplier (`FR-ABL-007`). If chosen independently, it compiles as an iterative targeting loop requiring multiple sequential client prompts!
    - *Rift* (#475): *"Review Headquarters self-damage keyword interaction (Unbowed)."* Does self-inflicted damage on Headquarters trigger reactive card keywords or damage mitigation?
    - *Daggerspine Rangers* (#529): *"Flagged for card rename review (prior art / IP collision with Warcraft 'Daggerspine')".* Intellectual property and trademark risk remains unaddressed in a core set asset.
  - **Exploit / Failure Scenario:**
    Engineers implement *Steel Rain* assuming single-target damage. Designers later review playtests and insist it was intended to snipe multiple damaged targets. The entire compiled AST schema, network message payload, and client prompt flow must be refactored.
  - **Remedy:** Hold an explicit game design triage session to close the open mechanical questions in Section 7.2 before finalizing the AST compilation schemas.

---

### 🟡 Medium

- **[DA-DAT-004] Arbitrary Prohibition of Dynamic Armor in Continuous Auras**
  - **Requirement Citations:** `FR-ABL-004`, `FR-LAY-002`
  - **Vulnerability / Flaw:** `FR-ABL-004` and `FR-LAY-002` state: *"Armor cannot be granted continuously by an aura; it exists strictly as a dynamic entity pool modified via one-shot events"*. This arbitrary mechanical prohibition prevents designers from creating standard defensive aura structures (e.g., *"Allied units in this zone have +1 Armor"* or *"While in Reserves, allied assets have 1 Armor"*). Forcing designers to implement defensive auras via complex trigger-and-leave watchers (e.g. `When this unit enters zone, add 1 armor; when it leaves, remove 1 armor`) introduces massive trigger clutter and desync potential.
  - **Remedy:** Either provide a clear game balance justification for banning continuous armor auras or permit Layer 4 auras to grant dynamic armor absorption buffers cleanly.

- **[DA-DAT-005] Uniform "Must Fit All" Compilability Policy Strangles Early Design Prototyping**
  - **Requirement Citations:** `FR-VAL-003`, `FR-DAT-005`
  - **Vulnerability / Flaw:** `FR-VAL-003` establishes that every single row in an authoring CSV must possess complete, fully bound declarative ASTs, with zero exceptions for incomplete prototype rows. In active game development, designers frequently pitch new card concepts with draft names and rough text before engineering has implemented the underlying ability primitives. Forcing every exploratory row to be 100% compilable blocks designers from checking work into git branches.
  - **Remedy:** Relax `FR-VAL-003` to allow exploratory rows if explicitly flagged with `currentProxy == TRUE` or `needsRework == TRUE`, preventing them from blocking local test builds while strictly excluding them from production releases via `-Strict`.

---

### 🔵 Low / Clarification

- **[DA-DAT-006] Subtype Casing Normalization Ambiguity**
  - **Requirement Citations:** `FR-DAT-006`, `FR-DAT-007`
  - **Issue:** `FR-DAT-006` strictly mandates uppercase enums (`UNIT`, `CONCORDIA`, `COMMON`), but `FR-DAT-007` mandates Title Case strings for subtypes (`["Infantry", "Scratch Company"]`). The schema validation rules should explicitly state whether subtype comparisons in filters (`FilterDef`) are case-insensitive or strictly Title Case.
  - **Clarification Needed:** Explicitly mandate case-insensitive string normalization during pipeline ingestion.

---

## 3. Redundancy & De-Scoping Ledger

| Requirement ID | Fully Subsumed By | Rationale / Recommendation |
| :--- | :--- | :--- |
| `FR-DAT-005` (Canonical Blueprint Schema) | `FR-CRD-101` | Blueprint property specifications in `Card_Data_Pipeline_FRD` duplicate `FR-CRD-101` in the Master Rules FRD. Cross-reference `FR-CRD-101` as the single canonical schema definition. |
| `FR-VAL-001` (Ingestion Pipeline Linter) | `FR-CMP-001` & `FR-CMP-022` | Linter verification phases duplicate compilation stages in `Card_Compiler_FRD.md`. Consolidate linter rules directly into compiler pre-flight checks. |

---

## 4. Socratic Interrogation Register

1. *How can Google Sheets be designated as the designers' primary balance workbench when running the publish script wipes out all remote edits without a diff check or backup?*
2. *If our declarative AST is supposed to eliminate engine recompilation for new cards, why does every new unique card effect add a bespoke hardcoded enum to `ValueToken.property`?*
3. *Why are unresolved game balance and targeting ambiguities for core set cards still open in a v1.0 specification?*
4. *Why is Armor strictly forbidden from being granted by Continuous Auras when temporary health and defense buffs are standard mechanics across the CCG genre?*
