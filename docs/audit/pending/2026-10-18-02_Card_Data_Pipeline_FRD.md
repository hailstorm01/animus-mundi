# Product Manager Devil's Advocate Audit: Card Data Pipeline & Ability Authoring FRD

- **Target Document:** [docs/frd/02_Card_Data_Pipeline_FRD.md](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/02_Card_Data_Pipeline_FRD.md)
- **Auditor:** PM Devil's Advocate (`pm-devils-advocate`)
- **Evaluation Date:** 2026-10-18
- **Status:** PENDING (Active review queue)
- **Verdict:** REVISE WITH HIGH FINDINGS

## 1. Executive Summary & Red-Team Assessment

This audit evaluates `02_Card_Data_Pipeline_FRD.md` (FRD v1.2) focusing on cross-spec consistency, AST schema completeness, targeting contracts, and domain boundary integrity.

`02_Card_Data_Pipeline_FRD.md` canonically governs CSV normalization, blueprint JSON schema enforcement, declarative AbilityDef AST polymorphic modeling, atomic effect primitives, and MTG-aligned rules text standards.

The audit identifies several high-severity schema contradictions and gaps:
1. **Contradictory `ConditionDef.cornered` Gate Scope**: In `FR-ABL-012`, the runtime execution semantics of `Cornered` are defined as `count(opposingAssets) > count(alliedAssets)`, completely omitting the spatial qualification `in the Main Theater` established in `01_Game_Rules_FRD.md` (`FR-KWD-002e`, `FR-LAY-003`).
2. **`CostDef` Action Naming Taxonomy Contradiction**: `FR-ABL-007` restricts `CostDef.costType` strictly to `[RESOURCE, EXHAUST, KILL, DISCARD_CARD, MOVE_CARD_ZONE]`. However, benchmark card #8 (*Funeral Fungus*) specifies `additionalDeploymentCosts: [EXHAUST_ASSET]`, and `05_Headless_Game_Engine_FRD.md` (`FR-ENG-013`) implements `EXHAUST_SELF`, `EXHAUST_ASSET`, and `KILL_ASSET`.
3. **Asymmetric Zone Taxonomy in `FilterDef` vs `ValueToken.query`**: `FilterDef.zone` omits `ATTACHED` and `BATTLEFIELD`, while `ValueToken.query.zone` omits `HEADQUARTERS` and `DECK`. Furthermore, `targetAttribute` in `query` specifies ambiguous `HEALTH` instead of distinguishing between `CURRENT_HEALTH` and `MAX_HEALTH`.
4. **Spurious `deckSelectionIndices` in Action Payload Model**: `FR-TGT-101` declares that action payloads bind `deckSelectionIndices`, whereas `04_Client_Engine_Protocol_Networking_FRD.md` (`FR-MSG-003`) omits this field from command schemas because deck partition choices are conducted mid-resolution via `PlayerChoicePrompt`.

## 2. Severity-Ordered Punch-List

### 🔴 Blocker / Critical
_None._

### 🟠 High
- **[DA-DAT-007] Contradictory ConditionDef Cornered Gate Scope**
  - **Requirement Citations:** `FR-ABL-012`, `FR-ABL-007`, `FR-KWD-002e`, `FR-LAY-003`
  - **Vulnerability / Flaw:** In `FR-ABL-012` (line 316), the keyword parameterization table defines the runtime engine execution semantics of `Cornered` as `Conditional gate (count(opposingAssets) > count(alliedAssets)) on TRIGGERED or CONTINUOUS_AURA`. This definition completely omits the spatial constraint `in the Main Theater` canonically established in `FR-KWD-002e` and `FR-LAY-003` (`docs/frd/01_Game_Rules_FRD.md`).
  - **Exploit / Failure Scenario:** If an AST compiler or engine interpreter implements the schema definition in `FR-ABL-012`, assets in Reserves and Headquarters zones will be counted toward Cornered evaluation, causing Cornered abilities to trigger or activate in violation of the core rules.
  - **Remedy:** Update `FR-ABL-012` line 316 to explicitly specify `count(opposingAssets in MAIN_THEATER) > count(alliedAssets in MAIN_THEATER)`.

- **[DA-DAT-008] Inconsistent CostDef Action Naming Taxonomy**
  - **Requirement Citations:** `FR-ABL-007`, `FR-ABL-009`, `FR-TRN-003a` (cross-ref `FR-ENG-013`)
  - **Vulnerability / Flaw:** `FR-ABL-007` (line 225) defines `CostDef` `costType` strictly as `[RESOURCE, EXHAUST, KILL, DISCARD_CARD, MOVE_CARD_ZONE]`. However, benchmark card #8 (*Funeral Fungus*, line 345) specifies `additionalDeploymentCosts: [EXHAUST_ASSET]`, while FRD 05 (`FR-ENG-013`) evaluates `EXHAUST_SELF`, `EXHAUST_ASSET`, and `KILL_ASSET`.
  - **Exploit / Failure Scenario:** The compiler's `AbilityDef.schema.json` rejects benchmark card ASTs containing `EXHAUST_ASSET` because `CostDef.costType` only permits `EXHAUST`, or the engine type-switch in `FR-ENG-013` fails to recognize `costType: EXHAUST` with a target selector.
  - **Remedy:** Harmonize `CostDef` schema in `FR-ABL-007` to canonical enum values `[RESOURCE, EXHAUST_SELF, EXHAUST_ASSET, KILL_ASSET, DISCARD_CARD, MOVE_CARD_ZONE]`, matching FRD 05 implementation.

### 🟡 Medium
- **[DA-DAT-009] Asymmetric Zone Taxonomy in FilterDef vs ValueToken.query**
  - **Requirement Citations:** `FR-ABL-007`, `FR-PRM-001`, `FR-ENT-102`
  - **Vulnerability / Flaw:** In `FR-ABL-007`, `FilterDef.zone` permits `[SAME_ZONE, RESERVES, MAIN_THEATER, HEADQUARTERS, DISCARD, HAND, DECK]`, omitting `ATTACHED` and `BATTLEFIELD`. Conversely, `ValueToken.query.zone` permits `[ATTACHED, DISCARD, BATTLEFIELD, HAND, RESERVES, MAIN_THEATER]`, omitting `HEADQUARTERS` and `DECK`. Furthermore, `targetAttribute` in `query` specifies ambiguous `HEALTH` instead of distinguishing between `CURRENT_HEALTH` and `MAX_HEALTH`.
  - **Exploit / Failure Scenario:** A designer creates a card querying the total health or power of assets in the Headquarters Zone or Deck, but `query.zone` rejects `HEADQUARTERS` and `DECK`. A targeting ability attempts to filter entities in the `ATTACHED` sub-zone, but `FilterDef.zone` rejects `ATTACHED`.
  - **Remedy:** Align zone enums across `FilterDef` and `ValueToken.query` to support all valid zones (`HEADQUARTERS`, `RESERVES`, `MAIN_THEATER`, `BATTLEFIELD`, `ATTACHED`, `HAND`, `DECK`, `DISCARD`), and split `HEALTH` in `query.targetAttribute` into `CURRENT_HEALTH` and `MAX_HEALTH`.

- **[DA-DAT-010] Spurious deckSelectionIndices in Action Payload Model**
  - **Requirement Citations:** `FR-TGT-101` (cross-ref `FR-MSG-003`, `FR-MSG-005`)
  - **Vulnerability / Flaw:** `FR-TGT-101` states that action declarations bind against the AST via structured payloads specifying `deckSelectionIndices`. However, in FRD 04 `FR-MSG-003`, no client action command supports deck selection indices upon declaration; all deck partition choices are conducted mid-resolution via `PlayerChoicePrompt`.
  - **Exploit / Failure Scenario:** Specification drift between data pipeline payload contracts and client-engine network commands.
  - **Remedy:** Remove `deckSelectionIndices` from `FR-TGT-101`, clarifying that deck inspections and choices are executed strictly via mid-resolution interactive choice prompts (`FR-MSG-005`).

### 🔵 Low / Clarification
- **[DA-DAT-011] Missing PERSISTENT Duration in Keyword and Subtype Grants**
  - **Requirement Citations:** `FR-PRM-001`, `FR-LAY-001`
  - **Vulnerability / Flaw:** `MODIFY_STAT` in `FR-PRM-001` allows `duration: [END_OF_TURN, END_OF_OPPONENT_TURN, PERSISTENT]`, but `GRANT_KEYWORD` and `GRANT_SUBTYPE` only allow `[END_OF_TURN, END_OF_OPPONENT_TURN]`. One-shot effects cannot permanently grant keywords or subtypes to assets.
  - **Exploit / Failure Scenario:** A design card attempting to permanently grant *Stalwart* or subtype *Chainbreaker* cannot be expressed in declarative AST without creating a continuous aura.
  - **Remedy:** Add `PERSISTENT` to the allowed durations of `GRANT_KEYWORD` and `GRANT_SUBTYPE` in `FR-PRM-001`.

## 3. Redundancy & De-Scoping Ledger
| Requirement ID | Fully Subsumed By | Rationale / Recommendation |
| :--- | :--- | :--- |
| `FR-DAT-003` | `FR-CMP-002`, `FR-CMP-003`, `FR-CMP-027`, `FR-ENT-001` | Atomic compilation and memory bundling are canonically specified in FRD 03 and FRD 05. |
| `FR-DAT-004` | `FR-CMP-024`–`FR-CMP-026`, `FR-SES-001`, `FR-SPC-002` | Canonical hashing, semantic versioning, and connection verification are owned by FRD 03 and FRD 04. |

## 4. Socratic Interrogation Register
1. *Why does ValueToken query permit aggregating attributes over ATTACHED cards, but FilterDef lacks an ATTACHED zone filter to allow abilities to target or interact with attached cards directly?*
2. *If CostDef uses EXHAUST with selector: SELF, why do benchmark cards and engine requirements use EXHAUST_ASSET and EXHAUST_SELF? What is the single canonical representation?*
