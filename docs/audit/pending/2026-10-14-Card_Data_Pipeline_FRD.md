# Product Manager Devil's Advocate Audit: Card Data Pipeline & Ability Authoring FRD (FRD v1.0)

- **Target Document:** [Card_Data_Pipeline_FRD.md](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/Card_Data_Pipeline_FRD.md)
- **Auditor:** PM Devil's Advocate (`pm-devils-advocate`)
- **Evaluation Date:** 2026-10-14
- **Status:** PENDING (Active review queue)
- **Verdict:** REVISE WITH BLOCKERS

---

## 1. Executive Summary & Red-Team Assessment

The **Card Data Pipeline & Ability Authoring FRD (v1.0)** creates a rigorous bridge between design spreadsheets, natural language rules text, declarative ASTs, and production blueprints. The adoption of MTG-aligned sentence grammar (`FR-TXT-001` - `FR-TXT-006`), strict target enum enforcement without runtime aliasing (`FR-DAT-006`), and deterministic slug derivation (`FR-DAT-002`) represents a clean and standardized data architecture.

However, red-teaming identifies serious workflow and architectural risks that will disrupt production velocity:
1. **Three-Tier Authority Model Sync Conflicts**: The split between Tier 1 (Cloud Google Sheets) and Tier 2 (Git CSVs) lacks automated two-way conflict detection (`FR-DAT-001`), creating a silent data-loss hazard when designers edit spreadsheets while engineers commit text corrections in git.
2. **Build Deadlock via 52 Flagged Review Items**: Domain 7 logs 52 unresolved design review cards (over 31% of the 167-card catalog). Coupled with the "Must Fit All" zero-tolerance validation policy (`FR-VAL-003`) and Strict Mode compiler gates, the CI/CD pipeline is virtually guaranteed to deadlock.
3. **Over-Engineered Declarative AST vs Parser Feasibility**: The granular polymorphic AST specification (`FR-ABL-001` - `FR-ABL-007`) models a level of syntactic complexity suitable for MTG Arena (thousands of cards), introducing massive cognitive and authoring overhead for a 167-card base set where direct code hooks or scriptable behaviors would be vastly simpler to deliver.

**Recommendation:** Resolve the 52 designer review items, implement safe 2-way sync conflict guards between Google Drive and Git, and establish clear migration tiers for complex ability ASTs.

---

## 2. Severity-Ordered Punch-List

### 🔴 Blocker / Critical

- **[DA-RULES-001] Three-Tier Data Authority Synchronization Race & Silent Overwrites**
  - **Requirement Citations:** `FR-DAT-001`, `FR-DAT-003`
  - **Vulnerability / Flaw:** `FR-DAT-001` defines Tier 1 (Google Sheets) as the collaborative design interface and Tier 2 (Git CSVs in `data/cards/`) as the repository source of truth, ruling that *"any discrepancy between cloud spreadsheets and repository CSVs resolves in favor of git-committed CSVs."* However, the sync tool `publish_card_sheets.ps1` executes batch downloads from Google Drive without 3-way merge capabilities or git conflict detection.
  - **Exploit / Failure Scenario:**
    1. A systems designer adjusts card stats (mana costs, health values) in the master Google Spreadsheet.
    2. Concurrently, an engineer fixes card rules text grammar or enum headers directly in `data/cards/*.csv` via a git pull request.
    3. Another team member runs `publish_card_sheets.ps1` to pull the latest spreadsheet changes.
    4. The script unconditionally overwrites the local CSV files with cloud data, silently obliterating the git-committed text and header corrections.
    5. Conversely, if an engineer commits changes to git, those changes are never propagated back to Google Sheets, causing immediate drift between what designers see in Drive and what compiles in engine builds.
  - **Remedy:** Implement a cryptographic revision handshake and bidirectional change-detection in `publish_card_sheets.ps1`. If local git CSVs have unmerged modifications relative to the last pulled Google Drive state, the script must abort with a merge conflict alert rather than blindly overwriting files.

- **[DA-SCOPE-001] 52 Unresolved Design Review Cards Paralyze Strict Pipeline Verification**
  - **Requirement Citations:** Domain 7 (`Section 7.1`, `Section 7.2`), `FR-VAL-003`, `FR-VAL-001`
  - **Vulnerability / Flaw:** `FR-VAL-003` mandates a strict "Must Fit All" validation policy where incomplete or unverified cards are rejected. Section 7.2 enumerates 52 cards with explicit open design flags (e.g. *Animoceps Rager* trigger frequency, *Mutation Swell* integer rounding, *Steel Rain* target selection, *Consuming Earth* structure yield calculations).
  - **Exploit / Failure Scenario:**
    1. Engineering enables strict compilation and linting (`lint_cardtown.ps1`, `compile_blueprints.ps1 -Strict`) in the CI/CD pull request gate.
    2. Over 31% of the entire card set (52 of 167 cards) contains pending design ambiguities flagged for review.
    3. Any commit triggering strict verification immediately fails, blocking pull requests and staging deployments across all disciplines until designers formally resolve all 52 open items.
    4. Development velocity halts, forcing engineers to disable `-Strict` mode, which in turn defeats the entire purpose of automated gating.
  - **Remedy:** Partition the card catalog into `Release-Candidate` and `Under-Review` tiers. Allow standard CI builds to compile and validate the resolved cards while quarantining the 52 review cards into a non-blocking diagnostic registry until designer sign-off is completed.

---

### 🟠 High

- **[DA-XP-001] AST Over-Engineering Imposes Massive Authoring Overhead for Base Set**
  - **Requirement Citations:** `FR-ABL-001` through `FR-ABL-007`, `FR-ABL-011`
  - **Vulnerability / Flaw:** The declarative ability AST introduces an exhaustive JSON schema featuring polymorphic ability types (`TRIGGERED`, `ACTIVATED`, `CONTINUOUS_AURA`, `REPLACEMENT_WATCHER`, `MANEUVER_EFFECT`, `KEYWORD_GRANT`), nested `FilterDef` objects, `ValueToken` dynamic property bindings, multi-stage cost arrays (`CostDef`), and distributed pools (`TargetRef`).
  - **Player & Engineering Experience Breakdown:** For a digital card game launching with a 167-card base set, maintaining a massive, deeply nested JSON AST for every complex card (e.g. *Samsara, Living Tempest*, *Light of the World Soul*, *Ambush Tactics*) creates extreme cognitive friction. Digital card games of this scale (e.g. Slay the Spire, Legends of Runeterra) succeed with lightweight behavior scripts or modular composition, whereas building an enterprise-grade AST engine introduces hundreds of potential serialization bugs without demonstrable player value.
  - **Remedy:** Restrict the declarative AST to standard keywords, basic triggers, and numeric modifiers. For complex multi-step cards (the top ~15% of cards), permit modular behavior scripts or simplified atomic action pipelines rather than trying to model full MTG Turing-complete declarative semantics.

- **[DA-RULES-002] Causality Ambiguity in Multi-Target Maneuvers ("Killed This Way")**
  - **Requirement Citations:** `FR-TXT-002`, `FR-ABL-007`, Section 7.2 (*Mutation Swell*)
  - **Vulnerability / Flaw:** *Mutation Swell* specifies: *"Kill up to 4 allied Scratch Company units. For every 2 units killed this way, deploy a unit card from your discard pile for free."*
  - **Exploit / Failure Scenario:**
    1. Player controls 3 Scratch Company units and 1 *Shifting Impossibility* ("If an allied Scratch Company unit would be killed, return it to its owner's hand and deal 1 damage to this unit instead.").
    2. Player casts *Mutation Swell*, targeting all 3 Scratch Company units.
    3. Instead of dying, 2 Scratch units are bounced to hand by *Shifting Impossibility*, and 1 dies.
    4. How many units were "killed this way"? Exactly 1. Integer division of $1 / 2$ yields 0.
    5. What if 3 units die? The spec notes: *"Designer should confirm rounding if an odd number of units is killed (integer division defaults to floor)."*
    6. If integer division floors, killing 3 units yields 1 deploy; killing 1 unit yields 0 deploys. If a player targeted 3 units expecting 2 deploys, the unexpected floor rounding leads to severe player frustration and perceived UI bugs.
  - **Remedy:** Eliminate fractional ratios from card designs. Rephrase card rules text to explicitly avoid fractional casualties: *"Kill 2 or 4 allied Scratch Company units. For every 2 units killed..."*, or state the exact payout per unit sacrificed.

---

### 🟡 Medium

- **[DA-SCOPE-002] Redundant Dual Schema Specifications Across FRD Suite**
  - **Requirement Citations:** `FR-DAT-005`, `FR-CRD-101`, Section 8.1 of Master FRD v3.0
  - **Vulnerability / Flaw:** `FR-DAT-005` ("Canonical Blueprint Schema Specification") provides a full markdown property table and constraints that duplicate `FR-CRD-101` in the Master Rules FRD.
  - **Assessment:** Dual maintenance of the canonical blueprint schema across two separate documents inevitably leads to version drift when properties (e.g. new cost types or deployment zone rules) are modified in one document and forgotten in the other.
  - **Remedy:** Designate `Card_Data_Pipeline_FRD.md` as the exclusive home of the data schema, and replace the schema table in `Game_Rules_and_Card_Modeling_FRD.md` with a direct hyperlink reference.

- **[DA-XP-002] Plural vs Singular Phrasing Ambiguities in Natural Language Grammar**
  - **Requirement Citations:** `FR-TXT-004`, `FR-TXT-005`
  - **Vulnerability / Flaw:** `FR-TXT-005` mandates using the verb "have" for continuous auras and "gets" for temporary buffs.
  - **Player Experience Breakdown:** Grammatical inconsistency arises when describing singular vs plural subjects: *"Other allied assets in the same zone have +1 power"* vs *"This unit has +1 power"* vs *"Chosen unit gets +1 power"*. When designers write cards, subtle grammatical differences between "have", "has", and "gets" create confusion for players translating between static continuous auras and discrete one-turn buffs.
  - **Remedy:** Provide an explicit grammar cheat-sheet in the pipeline documentation standardizing singular subject agreements (`has`) vs plural subject agreements (`have`).

---

### 🔵 Low / Clarification

- **[DA-CLARIFY-001] Sentinel Value Standardization for Null Subtypes**
  - **Requirement Citations:** `FR-DAT-007`
  - **Issue:** `FR-DAT-007` mandates the literal string `"--"` for empty subtypes. In some spreadsheets, designers leave cells completely blank or use `N/A`.
  - **Clarification Needed:** Ensure the compilation pipeline treats empty whitespace cells, `""`, and `"--"` identically as empty subtype arrays `[]` rather than failing validation.

- **[DA-CLARIFY-002] Target Invalidation Semantics for Dynamic Scaling Cards**
  - **Requirement Citations:** `FR-TGT-003`, Section 7.2 (*Consuming Earth*)
  - **Issue:** *Consuming Earth* targets an opposing unit with power less than or equal to the total Requisition yield of allied structures. If an allied structure is destroyed while *Consuming Earth* is on the stack, does the target eligibility check re-evaluate against the *new* structure yield, causing the spell to fizzle?
  - **Clarification Needed:** Explicitly affirm that two-stage verification (`FR-TGT-103`) re-evaluates dynamic filter clauses at resolution.

---

## 3. Redundancy & De-Scoping Ledger

| Requirement ID | Fully Subsumed By | Rationale / Recommendation |
| :--- | :--- | :--- |
| `FR-DAT-005` (Canonical Blueprint Schema) | `FR-CRD-101` (Master Rules FRD) | 100% duplicate property table. Unify into `FR-DAT-005` and cross-reference. |
| `FR-DAT-008` (Deployment Zone Array Schema) | `FR-DAT-005` & `FR-CRD-101` | Array schema constraints (`["RESERVES"]`, `["MAIN_THEATER"]`) are already defined in the blueprint schema. Merge into `FR-DAT-005`. |
| `FR-VAL-001` (Linter Specification) | `tools/linter/lint_cardtown.ps1` | The linter specification simply describes existing script behavior. De-scope narrative prose in favor of the executable script contract. |

---

## 4. Socratic Interrogation Register

1. *How can engineers guarantee data integrity when designers are actively modifying cloud spreadsheets while developers are committing rules text fixes in git?*
2. *Why are 52 cards (31% of the entire game) allowed to remain in an unresolved review state while the pipeline enforces a zero-tolerance "Must Fit All" validation policy?*
3. *Is a full MTG-style declarative AST with 6 polymorphic schemas genuinely necessary for a 167-card game, or is it introducing unneeded engineering complexity?*
4. *What happens to player trust when an interaction like Mutation Swell produces 0 deployments due to unstated integer division floor rounding?*
5. *Why are card blueprints defined in both Master FRD v3.0 and Card Data Pipeline FRD v1.0, creating an inevitable risk of documentation drift?*
