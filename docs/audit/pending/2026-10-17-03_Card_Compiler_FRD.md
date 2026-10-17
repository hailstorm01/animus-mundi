# Product Manager Devil's Advocate Audit: Card Compiler & Blueprint Ingestion FRD (v1.0)

- **Target Document:** [docs/frd/03_Card_Compiler_FRD.md](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/03_Card_Compiler_FRD.md)
- **Auditor:** PM Devil's Advocate (`pm-devils-advocate`)
- **Evaluation Date:** 2026-10-17
- **Status:** PENDING (Active review queue)
- **Verdict:** REVISE WITH BLOCKERS

## 1. Executive Summary & Red-Team Assessment
`03_Card_Compiler_FRD.md` establishes a solid two-tier hybrid compilation architecture (`compile_blueprints.ps1` + `ast_overrides.json`). However, subsequent updates to `02_Card_Data_Pipeline_FRD.md` (`DA-DAT-002`, `SOC-002`) and `05_Headless_Game_Engine_FRD.md` were not synchronized into `03_Card_Compiler_FRD.md`. As a result, `03_Card_Compiler_FRD.md` currently instructs the compiler to synthesize **deprecated `ValueToken` properties** that were explicitly removed from `02`, **non-existent event hook names** (`EVENT_PHASE_START`), a **conflicting root JSON envelope schema** for `blueprints.json`, and **mismatched AST property names** that would cause Stage 8 JSON Schema Validation (`FR-CMP-022`) to fail on every run.

## 2. Severity-Ordered Punch-List

### 🔴 Blocker / Critical
- **[DA-CMP-001] `FR-CMP-012` Synthesizes Deprecated Ad-Hoc `ValueToken` Properties Removed by `DA-DAT-002` in `02_Card_Data_Pipeline_FRD.md`**
  - **Requirement Citations:** `FR-CMP-012`, `FR-CMP-022`, `FR-ABL-007`
  - **Vulnerability / Flaw:** `FR-CMP-012` item 2 (lines 333–336) requires the compiler to synthesize dynamic game state counters into `type: "DYNAMIC"` `ValueToken` objects with `property: "COUNT_TUCKED_CARDS"`, `property: "COUNT_STRUCTURES_REQUISITION_YIELD"`, and `property: "COUNT_DISCARD_CARDS"`. However, `02_Card_Data_Pipeline_FRD.md` (`FR-ABL-007`, `Design Rationale DA-DAT-002`) explicitly abolished those three ad-hoc property enums and replaced them with the composable `type: "QUERY"` `ValueToken` schema (`query: { zone, filter, aggregate, targetAttribute }`).
  - **Exploit / Failure Scenario:**
    1. During Stage 5 (`FR-CMP-012`), the compiler parses *Consuming Earth* (*"total Requisition yield of all structures you control"*) or *All-Out War* (*"number of cards discarded this way"*) and emits `type: "DYNAMIC", property: "COUNT_STRUCTURES_REQUISITION_YIELD"`.
    2. During Stage 8 (`FR-CMP-022`), the compiler validates the generated AST against `AbilityDef.schema.json` (`FR-ABL-007`), where `ValueToken.property` only permits `["COMBAT_POWER", "CURRENT_HEALTH", "MAX_HEALTH", "BASE_REQUISITION_COST", "BASE_TACTICS_COST"]`.
    3. Compilation aborts at Stage 8 with exit code 4 (`ERR_SYNTAX_*`) on every execution.
  - **Remedy:** Update `FR-CMP-012` item 2 to emit composable `type: "QUERY"` `ValueToken` structures (and `CAUSALITY_CONTEXT` references) conforming to `FR-ABL-007` in `02_Card_Data_Pipeline_FRD.md`.

- **[DA-CMP-002] `FR-CMP-009` Binds `"At the start/end of your turn"` to Non-Existent Event Hook `EVENT_PHASE_START`**
  - **Requirement Citations:** `FR-CMP-009`, `FR-TRG-001`, `FR-ENG-009`
  - **Vulnerability / Flaw:** `FR-CMP-009` item 3 (lines 305–306) specifies that `^At the start of your turn,?\s*(.+)$` and `^At the end of your turn,?\s*(.+)$` bind to `triggerHook: "EVENT_PHASE_START"` with `condition: { turnOwner: "SELF", phase: "READY" | "END" }`. However, `EVENT_PHASE_START` does not exist in the canonical Trigger & Replacement Event Hook Taxonomy (`FR-TRG-001` in `02_Card_Data_Pipeline_FRD.md`, which defines `EVENT_READY_PHASE`, `EVENT_DRAW_PHASE`, `EVENT_MAIN_PHASE`, `EVENT_END_PHASE`, `EVENT_TURN_START`, `EVENT_TURN_END`) nor in `05_Headless_Game_Engine_FRD.md` (`FR-ENG-009`).
  - **Exploit / Failure Scenario:**
    1. The compiler parses *Hammerhand Bivouac* (*"At the start of your turn, distribute 1 armor among allied units in the same zone."*) via `FR-CMP-009` and emits `triggerHook: "EVENT_PHASE_START"`.
    2. Either Stage 8 schema validation (`FR-CMP-022`) rejects `"EVENT_PHASE_START"` as an illegal enum value, or the headless engine (`FR-ENG-009`) emits `EVENT_TURN_START` during Sub-Phase 1b and *Hammerhand Bivouac* never triggers.
  - **Remedy:** Update `FR-CMP-009` item 3 to bind `^At the start of your turn,?\s*(.+)$` to `EVENT_TURN_START` and `^At the end of your turn,?\s*(.+)$` to `EVENT_TURN_END` with `condition: { turnOwner: "SELF" }`, matching `FR-TRG-001` and `FR-ENG-009`.

### 🟠 High
- **[DA-CMP-003] Root `blueprints.json` Envelope Schema Conflict Between `FR-CMP-023` and `FR-DAT-005`**
  - **Requirement Citations:** `FR-CMP-022`, `FR-CMP-023`, `FR-DAT-005`
  - **Vulnerability / Flaw:** `FR-CMP-023` (lines 480–490) defines the emitted `blueprints.json` root envelope with properties `"catalogChecksum"`, `"cardDataVersion"`, `"generatedAt"`, `"totalCardCount"`, and `"blueprints": [ ... ]`. In contrast, `FR-DAT-005` in `02_Card_Data_Pipeline_FRD.md` (lines 130–168) defines the JSON Schema for `CardBlueprintCollection` with root properties `"version"` and `"cards": [ ... ]`.
  - **Exploit / Failure Scenario:** Stage 8 (`FR-CMP-022`) validates the catalog assembled in Stage 9/10 (`FR-CMP-023`) against `CardBlueprint.schema.json` (`FR-DAT-005`). Because `blueprints` vs. `cards` and `cardDataVersion` vs. `version` conflict, schema validation or engine ingestion (`FR-ENT-001`) fails.
  - **Remedy:** Align `FR-DAT-005` (`02_Card_Data_Pipeline_FRD.md`) and `FR-CMP-023` (`03_Card_Compiler_FRD.md`) on the single canonical envelope schema (`catalogChecksum`, `cardDataVersion`, `generatedAt`, `totalCardCount`, and `blueprints`).

- **[DA-CMP-004] Additional Enactment Cost AST Path Conflict & Missing Allied-Exhaust Cost Pattern in `FR-CMP-010`**
  - **Requirement Citations:** `FR-CMP-010`, `FR-DAT-005`, `FR-ABL-006`
  - **Vulnerability / Flaw:**
    1. `FR-CMP-010` item 5 (line 324) compiles `"As an additional cost to enact this maneuver, ..."` into an `additionalCosts` array inside `ManeuverEffectDef`, whereas `FR-DAT-005` (`02_Card_Data_Pipeline_FRD.md` lines 119, 152) places additional deployment/enactment costs on the top-level `CardBlueprint` under `additionalDeploymentCosts`.
    2. `FR-CMP-010` defines cost patterns for `R(\d+)`, `T(\d+)`, self-`exhaust`, `kill`, and `discard`, but omits `exhaust a readied allied unit` required by *Funeral Fungus* (*"As an additional cost to enact this maneuver, exhaust a readied allied unit."*).
  - **Exploit / Failure Scenario:** *Funeral Fungus* fails Stage 5 lexical cost parsing (`FR-CMP-010`), and additional costs on *Feed the Despoiler* and *Rhapsody's Embrace* are serialized into `ManeuverEffectDef.additionalCosts` instead of `CardBlueprint.additionalDeploymentCosts`.
  - **Remedy:** Update `FR-CMP-010` item 5 to populate `CardBlueprint.additionalDeploymentCosts` per `FR-DAT-005`, and add the cost pattern `exhaust a readied allied (.+)` compiling to `costType: "EXHAUST_ALLY", filter: <ParsedFilter>`.

### 🟡 Medium
- **[DA-CMP-005] AST Field Name & CSV Header/Boolean Coercion Drift (`FR-CMP-004`, `FR-CMP-005`, `FR-CMP-014`, `FR-CMP-015`)**
  - **Requirement Citations:** `FR-CMP-004`, `FR-CMP-005`, `FR-CMP-007`, `FR-CMP-014`, `FR-CMP-015`
  - **Vulnerability / Flaw:**
    1. `FR-CMP-014` item 1 (line 355) emits `bindUpfront: true`, whereas `FR-TGT-101` (`02_Card_Data_Pipeline_FRD.md` line 415) specifies `bindingPolicy: "UPFRONT_LOCK"`.
    2. `FR-CMP-015` item 1 (line 363) emits `op: "ADD"` in `ContinuousAuraDef.modifiers`, whereas `FR-ABL-004` (`02_Card_Data_Pipeline_FRD.md` line 505) specifies `operation: "ADD"`.
    3. `FR-CMP-005` item 3 (line 240) and `FR-CMP-007` assume CSV headers are camelCase (`name`, `currentProxy`, `needsRework`) with boolean strings `"TRUE"` / `"FALSE"` and 15 columns on all sheets, whereas `data/cards/*.csv` uses Title Case headers (`"Name"`, `"Current Proxy?"`, `"Needs Rework?"`) with `"Yes"` / `"No"` values and 9 columns on `maneuvers-*.csv` (which omit `Types`, `Deployment Zones`, `Power`, `Health`, `Requisition Value`, and `Tactics Value`).
  - **Exploit / Failure Scenario:** Synthesized `TriggeredAbilityDef` and `ContinuousAuraDef` objects fail Stage 8 schema validation due to `bindUpfront` vs. `bindingPolicy` and `op` vs. `operation`, and `maneuvers-*.csv` files fail Stage 3 ingestion due to omitted stat/zone columns.
  - **Remedy:** Update `FR-CMP-014` to emit `bindingPolicy: "UPFRONT_LOCK"`, update `FR-CMP-015` to emit `operation: "ADD"`, and update `FR-CMP-004`/`FR-CMP-005`/`FR-CMP-007` to map the 15-column unit/structure/HQ CSV headers and 9-column maneuver CSV headers (with `"Yes"`/`"No"` boolean coercion) to the canonical `CardBlueprint` properties.

## 3. Redundancy & De-Scoping Ledger
| Requirement ID | Fully Subsumed By | Rationale / Recommendation |
| :--- | :--- | :--- |
| `FR-CMP-011` | `FR-TXT-004` & `FR-TGT-001` | Already marked `*(Subsumed by FR-TXT-004 & FR-TGT-001)*` in Section 3.1. |
| `FR-CMP-028` | `FR-CMP-001` & `FR-CMP-027` | Already marked `*(Subsumed by FR-CMP-001 & FR-CMP-027)*` in Section 9.1. |

## 4. Socratic Interrogation Register
1. *Since `DECISION 1` (`FR-CMP-017`) scopes automated PowerShell regex synthesis strictly to standalone keywords, standard cost headers, and simple single-clause triggers—routing all composite/bespoke cards to `data/schemas/ast_overrides.json`—should `FR-CMP-012` (dynamic `ValueToken` synthesis for tucked cards, structure yields, and variable X) be removed from the PowerShell regex parser entirely and handled exclusively via `ast_overrides.json`?*
