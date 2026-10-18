# Product Manager Devil's Advocate Audit: Headless Game Engine FRD

- **Target Document:** [docs/frd/05_Headless_Game_Engine_FRD.md](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/05_Headless_Game_Engine_FRD.md)
- **Auditor:** PM Devil's Advocate (`pm-devils-advocate`)
- **Evaluation Date:** 2026-10-18
- **Status:** PENDING (Active review queue)
- **Verdict:** REVISE WITH HIGH FINDINGS

## 1. Executive Summary & Red-Team Assessment

This audit evaluates `05_Headless_Game_Engine_FRD.md` (FRD v1.4) across engine runtime contracts, stack architecture, transaction rollback lifecycle, error code parity, and cross-spec integration.

`05_Headless_Game_Engine_FRD.md` canonically governs the headless Go kernel architecture (`cardtown/engine`), discrete integer math, isolated PRNG stream counters (`FR-ENG-002`), transactional execution with delta-rollback mementos (`FR-ENG-008`), 3-tier runtime memory structs (`CardBlueprint`, `CardInstance`, `AssetInstance`, `AssetSnapshot`), frame tick invariants and rolling state hashes (`FR-ENG-005`), stack frame taxonomy (`FR-ENG-012`), Go AST type-switch visitors (`FR-ENG-018`), CLI simulation runners, and automated benchmarks.

The audit identifies significant architectural gaps and cross-spec discrepancies:
1. **Stack Frame Taxonomy Gap for Tactical Inherent Actions (`FR-ENG-012`)**: `FR-ENG-012` defines only 5 stack frames (`FRAME_DEPLOYMENT`, `FRAME_MANEUVER`, `FRAME_ACTIVATED_ABILITY`, `FRAME_ATTACK`, `FRAME_TRIGGERED_ABILITY`), forcing `March` and `Deploy Scratch Company` into `FRAME_ACTIVATED_ABILITY`. However, `04_Client_Engine_Protocol_Networking_FRD.md` (`FR-MSG-004`) defines discrete wire types `"MARCH"` and `"DEPLOY_SCRATCH"`, and FRD 05's own ASCII architecture diagram (line 381) depicts `March` and `HQ` as distinct frames.
2. **Missing `TRIGGER_ORDER_SELECT` Implementation in Choice Prompt Engine**: FRD 01 `FR-TRG-003` mandates that when multiple triggers fire simultaneously, players place them onto the stack "in the order of their choosing", and FRD 04 `FR-MSG-005` defines `prompt_type = "TRIGGER_ORDER_SELECT"` with `ordered_trigger_indices`. However, FRD 05's choice prompt yield protocol (`FR-ENG-017`) completely omits `TRIGGER_ORDER_SELECT`, and `FR-ENG-014` was subsumed into natural LIFO stack ordering.
3. **Wire Error Traceability Drift in Error Code Registry**: Section 8.1 defines `ERR_DESYNC_DETECTED` and `ERR_INFINITE_LOOP`, claiming wire traceability to `FR-ERR-001 (04_Client_Engine_Protocol_Networking_FRD.md)`. However, `FR-ERR-001` in FRD 04 does not define either error code. Furthermore, Section 8.1 lacks an error code for deck validation failures (`ERR_DECK_LEGALITY_VIOLATION`, referenced in FRD 04 line 105).
4. **Zone Identifier Discrepancy in `AssetSnapshot` (`"HQ"` vs `"HEADQUARTERS"`)**: `FR-SNP-101` and Section 2.2 specify `AssetSnapshot.LastBattlefieldZone` as `"HQ"`, whereas `CardInstance.CurrentZone` in `FR-ENT-002` and all rules specifications in FRD 01/02 enforce `"HEADQUARTERS"`.

## 2. Severity-Ordered Punch-List

### 🔴 Blocker / Critical
_None._

### 🟠 High
- **[DA-ENG-009] Stack Frame Taxonomy Gap for Tactical Inherent Actions (FR-ENG-012)**
  - **Requirement Citations:** `FR-ENG-012`, `FR-INH-002`, `FR-INH-004` (cross-ref `FR-MSG-004` in FRD 04)
  - **Vulnerability / Flaw:** `FR-ENG-012` defines only 5 stack frames: `FRAME_DEPLOYMENT`, `FRAME_MANEUVER`, `FRAME_ACTIVATED_ABILITY`, `FRAME_ATTACK`, `FRAME_TRIGGERED_ABILITY`. It forces `March` and `Deploy Scratch Company` into `FRAME_ACTIVATED_ABILITY`. However, FRD 04 `StackPushedEvent` defines discrete wire types `"MARCH"` and `"DEPLOY_SCRATCH"`, and FRD 05's own ASCII diagram (line 381) depicts `March` and `HQ` as distinct frames.
  - **Exploit / Failure Scenario:** When an asset marches or deploys a scratch company, the engine pushes a `FRAME_ACTIVATED_ABILITY`. When emitting `StackPushedEvent`, the engine either emits `ability_type = "ACTIVATED"`, causing the Godot client UI to fail to display tactical March/Scratch indicators, or must inspect the internal action payload to synthesize a non-standard frame type.
  - **Remedy:** Formalize `FRAME_MARCH` and `FRAME_DEPLOY_SCRATCH` as first-class `StackFrame` variants in `FR-ENG-012`, aligning engine memory architecture 1:1 with wire protocol `StackPushedEvent.ability_type`.

- **[DA-ENG-010] Missing TRIGGER_ORDER_SELECT Implementation in Choice Prompt Engine**
  - **Requirement Citations:** `FR-ENG-017`, `FR-TRG-003`, `FR-MSG-005`
  - **Vulnerability / Flaw:** FRD 01 `FR-TRG-003` mandates that when multiple triggers fire simultaneously, players place them onto the stack "in the order of their choosing", and FRD 04 `FR-MSG-005` defines `prompt_type = "TRIGGER_ORDER_SELECT"` with `ordered_trigger_indices`. However, FRD 05's choice prompt yield protocol (`FR-ENG-017`) completely omits `TRIGGER_ORDER_SELECT`, and `FR-ENG-014` was subsumed into natural LIFO stack ordering.
  - **Exploit / Failure Scenario:** When a player triggers multiple abilities on the same event (e.g. *Hammerhand Brigands* and *Eternal Watchman* on combat damage), the headless engine auto-queues them in arbitrary memory order without yielding `TRIGGER_ORDER_SELECT` to the player, violating `FR-TRG-003` and depriving players of tactical ordering agency.
  - **Remedy:** Formalize `TRIGGER_ORDER_SELECT` yield protocol in `FR-ENG-017`, specifying that when multiple triggers belong to the same player, the engine yields a `PlayerChoicePrompt` allowing the player to submit `ordered_trigger_indices`.

### 🟡 Medium
- **[DA-ENG-011] Wire Error Traceability Drift in Error Code Registry**
  - **Requirement Citations:** Section 8.1, `FR-ENG-041` (cross-ref `FR-ERR-001` in FRD 04)
  - **Vulnerability / Flaw:** Section 8.1 defines `ERR_DESYNC_DETECTED` and `ERR_INFINITE_LOOP`, claiming wire traceability to `FR-ERR-001 (04_Client_Engine_Protocol_Networking_FRD.md#L745)`. However, `FR-ERR-001` in FRD 04 does not define either error code. Furthermore, Section 8.1 lacks an error code for deck validation failures (`ERR_DECK_LEGALITY_VIOLATION`, referenced in FRD 04 line 105).
  - **Exploit / Failure Scenario:** Engine emits `ERR_DESYNC_DETECTED` or `ERR_INFINITE_LOOP`, but the client networking layer rejects the packet as an unmapped error code.
  - **Remedy:** Add missing error codes to FRD 04 `FR-ERR-001` and add `ERR_DECK_LEGALITY_VIOLATION` to Section 8.1.

- **[DA-ENG-012] Zone Identifier Discrepancy in AssetSnapshot (HQ vs HEADQUARTERS)**
  - **Requirement Citations:** `FR-SNP-101`, `FR-ENT-002`, `FR-ZON-001`
  - **Vulnerability / Flaw:** `FR-SNP-101` and Section 2.2 specify `AssetSnapshot.LastBattlefieldZone` as `"HQ"`, whereas `CardInstance.CurrentZone` in `FR-ENT-002` and all rules specifications in FRD 01/02 enforce `"HEADQUARTERS"`.
  - **Exploit / Failure Scenario:** State comparison logic fails when comparing a `CardInstance`'s zone (`HEADQUARTERS`) against an `AssetSnapshot`'s zone (`HQ`).
  - **Remedy:** Standardize `AssetSnapshot` zone field to `"HEADQUARTERS"`.

### 🔵 Low / Clarification
- **[DA-ENG-013] Match Draw Conclusion Reason Ambiguity**
  - **Requirement Citations:** `FR-ENG-042`, `FR-CBT-003` (cross-ref `FR-MSG-004` in FRD 04)
  - **Vulnerability / Flaw:** In `FR-ENG-042`, simultaneous HQ defeat terminates as a Match Draw (`RESULT_MATCH_DRAW`). But FRD 04 `MatchConcludedEvent` lists `conclusion_reason` values: `"HQ_DESTROYED", "EMPTY_DECK_DRAW", "CONCEDE", "FORFEIT_TIMEOUT", "FORFEIT_PROTOCOL_ABUSE"`. It is ambiguous whether the engine emits `conclusion_reason: "HQ_DESTROYED"` with `is_match_draw: true`, or a distinct reason string.
  - **Exploit / Failure Scenario:** Client replay parser expects a specific draw conclusion string, but receives `"HQ_DESTROYED"`.
  - **Remedy:** Clarify in `FR-ENG-042` that simultaneous HQ defeat emits `conclusion_reason = "HQ_DESTROYED"` with `is_match_draw = true`.

## 3. Redundancy & De-Scoping Ledger
| Requirement ID | Fully Subsumed By | Rationale / Recommendation |
| :--- | :--- | :--- |
| `FR-ENG-003` | `FR-ENG-002` | PRNG seed derivation unified with multi-stream architecture in `FR-ENG-002`. |
| `FR-ENG-014` | `FR-TRG-003`, `FR-TRG-004`, `FR-ENG-012` | APNAP rules owned by FRD 01; stack architecture owned by `FR-ENG-012`. |
| `FR-ENG-015` | `FR-TGT-001`–`FR-TGT-003`, `FR-ENG-012` | Target verification and fizzle rules owned by FRD 01. |
| `FR-ENG-019`, `FR-ENG-020` | `FR-ENG-018` | Dynamic filter and value evaluation subsumed into AST type-switch interpreter. |
| `FR-ENG-024` | `FR-ENG-012`, `FR-ENG-013`, `FR-MNV-001` | Maneuver execution flow subsumed into stack frame lifecycle and cost validation. |
| `FR-ENG-026`, `FR-ENG-027`, `FR-ENG-028` | `FR-ENG-025`, `FR-ENG-041` | Caching de-scoped; timestamp ordering and power clamping subsumed into `FR-ENG-025` and Invariant 1. |
| `FR-ENG-030` | `FR-CBT-001`, `FR-ENG-031` | Spatial combat eligibility belongs to FRD 01; attack declaration validation belongs to `FR-ENG-031`. |
| `FR-ENG-034` | `FR-REP-004`, `FR-ENG-035` | 5-stage damage mitigation belongs to FRD 01; death sweep execution belongs to `FR-ENG-035`. |

## 4. Socratic Interrogation Register
1. *Why does the engine allow transaction delta-rollback for player commands, but commits transactions permanently before waiting for player choice responses (FR-ENG-008)? Does this asymmetry allow malicious clients to exploit two-phase choice prompts by inspecting intermediate state before responding?*
2. *If stack depth exceeds 100 frames, triggering ERR_INFINITE_LOOP and force-clearing the stack, why does this not terminate the match or assign a forfeit to the player whose recursive trigger created the infinite loop?*
