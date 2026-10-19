# Product Manager Devil's Advocate Audit: Card Data Pipeline FRD (Master FRD v1.3)

- **Target Document:** [docs/frd/02_Card_Data_Pipeline_FRD.md](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/02_Card_Data_Pipeline_FRD.md)
- **Auditor:** PM Devil's Advocate (`pm-devils-advocate`)
- **Evaluation Date:** 2026-10-19
- **Status:** PENDING (Active review queue)
- **Verdict:** CONDITIONAL PASS

## 1. Executive Summary & Red-Team Assessment

The Card Data Pipeline specification (`02_Card_Data_Pipeline_FRD.md`) effectively models the authoring bridge between tabletop rules and the digital runtime engine, establishing declarative Ability AST schemas, MTG-aligned text standards, and zero-aliasing CSV contracts. However, critical modeling ambiguities and cross-spec gaps remain: `ValueToken.query` defines an ad-hoc `zone` field that collides with its embedded `FilterDef.zone` property while omitting `SAME_ZONE`; `CostDef` supports `costType: MOVE_CARD_ZONE` (*Arkha, Sun's Fury*) without a corresponding `TargetRole` wire binding in FRD 04; `FR-VAL-002` and `FR-VAL-003` duplicate compiler requirements while citing retired error codes; and `ConditionDef` contains redundant, overlapping predicates.

## 2. Severity-Ordered Punch-List

### 🔴 Blocker / Critical
_None._

### 🟠 High
- **[DA-DAT-012] Dual Zone Field Collision and Redundancy in `ValueToken.query`**
  - **Requirement Citations:** `FR-ABL-007` (cross-ref `FilterDef`)
  - **Vulnerability / Flaw:** In Section 3.2 (`FR-ABL-007`), `ValueToken` defines `query` as `{ zone: [RESERVES, MAIN_THEATER, HEADQUARTERS, BATTLEFIELD, ATTACHED, HAND, DECK, DISCARD], filter: FilterDef, aggregate: [...], targetAttribute: [...] }`. However, `FilterDef` *already* defines a canonical `zone` property (`SAME_ZONE`, `RESERVES`, `MAIN_THEATER`, `HEADQUARTERS`, `BATTLEFIELD`, `ATTACHED`, `HAND`, `DECK`, `DISCARD`). Having a top-level `zone` on `query` alongside `filter: FilterDef` creates an immediate conflict: if `query.zone` is set to `RESERVES` while `query.filter.zone` is set to `MAIN_THEATER`, engine evaluation behavior is undefined. Furthermore, `query.zone` omits `SAME_ZONE`, making it impossible for queries to filter relative to the evaluating asset's current zone without creating a field conflict.
  - **Exploit / Failure Scenario:** A card ability counts allied assets in the same zone. The author specifies `filter: { zone: "SAME_ZONE" }` and is forced to set `query.zone: "BATTLEFIELD"`. The query evaluator evaluates all battlefield zones because `query.zone` overrides `filter.zone`, miscalculating the dynamic value.
  - **Remedy:** Deprecate the redundant `zone` property directly under `ValueToken.query` and rely exclusively on `query.filter.zone` (`FilterDef`), eliminating the field collision and restoring full access to `SAME_ZONE`.

- **[DA-DAT-013] Missing Wire Mapping and Schema Gap for `MOVE_CARD_ZONE` Cost Type**
  - **Requirement Citations:** `FR-ABL-007`, `FR-TXT-003` (cross-ref `04_Client_Engine_Protocol_Networking_FRD.md` `FR-MSG-003` `TargetRole`)
  - **Vulnerability / Flaw:** `CostDef` in `FR-ABL-007` defines `costType: MOVE_CARD_ZONE` with `destinationPosition: [TOP, BOTTOM, SHUFFLED_IN]` (demonstrated on *Arkha, Sun's Fury*: *"Whenever this unit attacks, you may put a unit card from your discard pile on the bottom of your deck. If you do, deal 3 damage to chosen opposing Headquarters"*). However, FRD 04 `enum TargetRole` (`FR-MSG-003`) only provides `TARGET_ROLE_SACRIFICE_COST = 3`, `TARGET_ROLE_DISCARD_COST = 4`, and `TARGET_ROLE_EXHAUST_COST = 5`. There is no `TARGET_ROLE_MOVE_COST` or corresponding wire protocol binding role for target cards chosen to pay a `MOVE_CARD_ZONE` cost, leaving client-server action payload serialization undefined for cards like *Arkha, Sun's Fury*.
  - **Exploit / Failure Scenario:** A client attempts to declare an attack or trigger response with *Arkha, Sun's Fury*. To bind the discard card moved to the bottom of the deck, the client has no valid `TargetRole` in `TargetBinding`, causing the server to reject the command with `ERR_INVALID_COMMAND_ARGUMENTS`.
  - **Remedy:** Add `TARGET_ROLE_MOVE_COST = 8` to `enum TargetRole` in FRD 04 `FR-MSG-003`, and cross-reference it in `FR-ABL-007`.

### 🟡 Medium
- **[DA-DAT-014] Redundant Text Parity Requirements Subsumed by Compiler FRD**
  - **Requirement Citations:** `FR-VAL-002`, `FR-VAL-003` (cross-ref `03_Card_Compiler_FRD.md` `FR-CMP-019`, `FR-CMP-020`, `FR-CMP-021`)
  - **Vulnerability / Flaw:** `FR-VAL-002` in FRD 02 is a 2-line requirement stating "In compliance with FR-CMP-019 and FR-CMP-020, the compilation pipeline verifies 1-to-1 parity...". It adds zero pipeline-specific specifications beyond what `FR-CMP-019` and `FR-CMP-020` canonically define in FRD 03. Similarly, `FR-VAL-003` cites retired error code `ERR_STRICT_UNRESOLVED_REVIEW`, which FRD 03 explicitly retired and subsumed into `ERR_STRICT_PROTOTYPE_REJECTED` and `ERR_OVERRIDE_UNVERIFIED` (`FR-CMP-021`).
  - **Exploit / Failure Scenario:** Dual-spec maintenance drift occurs when compiler error codes or parity checking procedures are updated.
  - **Remedy:** Formally mark `FR-VAL-002` as subsumed by `FR-CMP-019` / `FR-CMP-020`, and purge the retired `ERR_STRICT_UNRESOLVED_REVIEW` reference from `FR-VAL-003`.

- **[DA-DAT-015] Overlapping and Redundant Predicates in `ConditionDef`**
  - **Requirement Citations:** `FR-ABL-007`
  - **Vulnerability / Flaw:** `ConditionDef` defines `source (SELF, ALLIED, OPPOSING, ANY)` and `deployedZone (RESERVES, MAIN_THEATER)`, while simultaneously embedding all predicates of `FilterDef` (which contains `allegiance` and `zone`). Having both `source` and `allegiance`, as well as `deployedZone` and `zone`, introduces ambiguous precedence and dual-definition confusion for ability authors.
  - **Exploit / Failure Scenario:** An author writes `{ allegiance: "ALLIED", source: "OPPOSING" }` or `{ deployedZone: "RESERVES", zone: "MAIN_THEATER" }`. Schema validation allows both, but engine execution behavior is undefined.
  - **Remedy:** Formalize strict field precedence in `ConditionDef`, standardizing on `FilterDef` properties (`allegiance`, `zone`) where possible and documenting that `deployedZone` specifically applies to `EVENT_DEPLOY` destination filtering.

### 🔵 Low / Clarification
_None._

## 3. Redundancy & De-Scoping Ledger
| Requirement ID | Fully Subsumed By | Rationale / Recommendation |
| :--- | :--- | :--- |
| `FR-DAT-003` | `FR-CMP-002`, `FR-CMP-003`, `FR-CMP-027`, `FR-ENT-001` | Atomic compilation and memory bundling are canonically specified in FRD 03 and FRD 05. |
| `FR-DAT-004` | `FR-CMP-024`–`FR-CMP-026`, `FR-SES-001`, `FR-SPC-002` | RFC 8785 hashing, dataset SemVer, and handshake verification are owned by FRD 03 and FRD 04. |
| `FR-VAL-001` | `FR-CMP-001`, `FR-CMP-022` | Compiler CLI preflight validation and schema verification are canonically owned by FRD 03. |
| `FR-VAL-002` | `FR-CMP-019`, `FR-CMP-020` | Parity verification between rulesText and AST is canonically owned and enforced by FRD 03. |

## 4. Socratic Interrogation Register
1. *If authoring CSVs are the absolute source of truth, why are non-keyword card abilities maintained in a separate manual JSON sidecar (`ast_overrides.json`) rather than generated directly by the compiler from structured CSV columns or embedded AST payloads? Does this split-source architecture risk silent text-to-AST drift whenever designers edit balance CSVs?*
