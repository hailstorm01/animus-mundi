# Product Manager Devil's Advocate Audit: Client-Engine Protocol & Networking FRD

- **Target Document:** [docs/frd/04_Client_Engine_Protocol_Networking_FRD.md](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/04_Client_Engine_Protocol_Networking_FRD.md)
- **Auditor:** PM Devil's Advocate (`pm-devils-advocate`)
- **Evaluation Date:** 2026-10-18
- **Status:** PENDING (Active review queue)
- **Verdict:** REVISE WITH BLOCKERS

## 1. Executive Summary & Red-Team Assessment

This audit evaluates `04_Client_Engine_Protocol_Networking_FRD.md` (FRD v1.1) across wire protocol taxonomy, event stream completeness, error handling, state synchronization, and cross-spec boundaries.

`04_Client_Engine_Protocol_Networking_FRD.md` canonically governs WebSocket networking, Protobuf message framing (`GameEnvelope`), client command taxonomy, server event streaming (`ServerEventStream`), double-blind information sanitization, animation queue pacing, public state checksums (`FNV-1a`), and match replay architecture (`.rec`).

The audit reveals a critical wire-level blocker and severe protocol contradictions:
1. **CRITICAL BLOCKER: Omission of `EVENT_COMBAT_KILL` from `EngineEvent`**: `01_Game_Rules_FRD.md` (`FR-CBT-002`), `02_Card_Data_Pipeline_FRD.md` (`FR-TRG-001`), and `05_Headless_Game_Engine_FRD.md` (`FR-ENG-035`) explicitly require the engine to emit `EVENT_COMBAT_KILL` carrying `killer_asset_id`, `victim_asset_id`, and `is_attacker: bool` when an asset dies from combat damage. However, in FRD 04 `FR-MSG-004`, `EngineEvent`'s `oneof detail` contains NO message definition for `combat_kill`! The headless engine cannot serialize this event over the wire, breaking client animation and desyncing client-side trigger execution for cards watching for combat kills.
2. **Contradictory `ERR_INVALID_TARGET` Severity & Purpose**: In FRD 04 line 945, `ERR_INVALID_TARGET` is a fatal command rejection error returned when a command targets an illegal/protected entity upon declaration. In FRD 05 Section 8.1 (line 629), `ERR_INVALID_TARGET` is a non-fatal Warning emitted during stack pop when a target becomes illegal prior to resolution, triggering a fizzle.
3. **Missing Protobuf `enum ErrorCode` and Undefined Error Codes**: `CommandRejection` defines `uint32 error_code = 2`, but FRD 04 provides no `enum ErrorCode` mapping error names to numeric values. Furthermore, `ERR_DESYNC_DETECTED` and `ERR_INFINITE_LOOP` (defined in FRD 05 Section 8.1 with citations to `FR-ERR-001`) are completely omitted from FRD 04, as are `ERR_VERSION_MISMATCH` and `ERR_CATALOG_CHECKSUM_MISMATCH` (used in `HandshakeReject`).
4. **Zone Naming Conflict (`"HQ"` vs `"HEADQUARTERS"`)**: `AssetSnapshot.zone` (line 343) specifies `"HQ"`, whereas FRD 01 (`FR-ZON-001`), FRD 02 (`FR-DAT-006`), and FRD 05 (`FR-ENT-002`) strictly prohibit `"HQ"` and enforce `"HEADQUARTERS"`.
5. **Event Hook Naming Contradiction**: `DelayedTriggerRegisteredEvent` (line 748) gives `"EVENT_END_PHASE"` as an example hook, directly contradicting FRD 02 line 247 which explicitly mandates that end-of-turn triggers bind to `"EVENT_TURN_END"` and forbids binding to `"EVENT_END_PHASE"`.

## 2. Severity-Ordered Punch-List

### 🔴 Blocker / Critical
- **[DA-NET-007] Omission of EVENT_COMBAT_KILL from Protobuf EngineEvent Stream**
  - **Requirement Citations:** `FR-MSG-004`, `FR-CBT-002`, `FR-TRG-001` (cross-ref `FR-ENG-035`)
  - **Vulnerability / Flaw:** `01_Game_Rules_FRD.md` (`FR-CBT-002`), `02_Card_Data_Pipeline_FRD.md` (`FR-TRG-001`), and `05_Headless_Game_Engine_FRD.md` (`FR-ENG-035`) explicitly mandate that when an asset is destroyed by combat damage, the engine emits `EVENT_COMBAT_KILL` carrying `killer_asset_id`, `victim_asset_id`, and `is_attacker: bool` so that combat-kill triggers (*Dread Manifestation*, *Eternal Watchman*) can evaluate killer identity. However, in FRD 04 `FR-MSG-004` (lines 582-604), `EngineEvent`'s `oneof detail` contains NO message definition for `combat_kill`!
  - **Exploit / Failure Scenario:** When combat damage kills an asset, the headless engine attempts to emit `EVENT_COMBAT_KILL` over the wire, but the Protobuf serialization layer cannot serialize the event because no matching field exists in `EngineEvent`. The server either crashes or drops the combat kill event, breaking client animation and desyncing client-side trigger resolution for cards that watch for combat kills.
  - **Remedy:** Add `CombatKillEvent combat_kill = 31;` to `EngineEvent`'s `oneof detail` in `FR-MSG-004` containing `string killer_asset_id = 1; string victim_asset_id = 2; bool is_attacker = 3;`.

### 🟠 High
- **[DA-NET-008] Semantic Contradiction & Severity Inversion for ERR_INVALID_TARGET**
  - **Requirement Citations:** `FR-ERR-001` (cross-ref `FR-ENG-041`, Section 8.1 in FRD 05)
  - **Vulnerability / Flaw:** In FRD 04 line 945, `ERR_INVALID_TARGET` is defined as a fatal `CommandRejection` error code transmitted to clients when a command targets an illegal entity upon declaration. In FRD 05 Section 8.1 (line 629), `ERR_INVALID_TARGET` is defined as a non-fatal `Severity: Warning` emitted during stack pop when a target becomes illegal prior to resolution, triggering a fizzle.
  - **Exploit / Failure Scenario:** When a target fizzles on the stack in FRD 05, the engine logs `ERR_INVALID_TARGET` as a warning while `StackPoppedEvent` notifies the client with `did_fizzle = true`. But if a client attempts to target an Ethereal unit upon declaration, FRD 04 expects `CommandRejection` with `ERR_INVALID_TARGET`. The dual usage creates contradictory wire semantics.
  - **Remedy:** Split `ERR_INVALID_TARGET` into two explicit codes: `ERR_INVALID_TARGET_SELECTION` (Error, upfront command rejection) and `WARN_TARGET_INVALIDATED_FIZZLE` (Warning, resolution-time fizzle).

- **[DA-NET-009] Missing Protobuf enum ErrorCode and Undefined Wire Error Codes**
  - **Requirement Citations:** `FR-ERR-001`, `FR-NET-003` (cross-ref Section 8.1 in FRD 05)
  - **Vulnerability / Flaw:** `CommandRejection` defines `uint32 error_code = 2`, but FRD 04 provides no `enum ErrorCode` mapping error names to numeric values. Furthermore, `ERR_DESYNC_DETECTED` and `ERR_INFINITE_LOOP` (defined in FRD 05 Section 8.1 with citations to `FR-ERR-001`) are completely omitted from FRD 04, as are `ERR_VERSION_MISMATCH` and `ERR_CATALOG_CHECKSUM_MISMATCH` (used in `HandshakeReject`).
  - **Exploit / Failure Scenario:** Clients receiving a `CommandRejection` packet receive an undocumented integer in `error_code`, preventing client-side UI from displaying appropriate error messages or routing to recovery workflows.
  - **Remedy:** Define an explicit `enum ErrorCode` in `FR-ERR-001` assigning distinct integer constants to all error codes across the system, including `ERR_DESYNC_DETECTED` and `ERR_INFINITE_LOOP`.

### 🟡 Medium
- **[DA-NET-010] Inconsistent Zone Naming in AssetSnapshot (HQ vs HEADQUARTERS)**
  - **Requirement Citations:** `FR-MSG-001`, `FR-ZON-001`, `FR-DAT-006`
  - **Vulnerability / Flaw:** `AssetSnapshot.zone` (line 343) uses string value `"HQ"`. However, FRD 01 (`FR-ZON-001`), FRD 02 (`FR-DAT-006`), and FRD 05 (`FR-ENT-002`) strictly prohibit `"HQ"` and mandate `"HEADQUARTERS"`.
  - **Exploit / Failure Scenario:** Client-side zone routing parses `zone == "HEADQUARTERS"` per FRD 01/02 enums and fails to identify the Headquarters asset because the snapshot sends `"HQ"`.
  - **Remedy:** Update `AssetSnapshot.zone` comment and implementation in `FR-MSG-001` to use `"HEADQUARTERS"`.

- **[DA-NET-011] Contradictory Event Hook Example in DelayedTriggerRegisteredEvent**
  - **Requirement Citations:** `FR-MSG-004`, `FR-TRG-001` (FRD 02)
  - **Vulnerability / Flaw:** In line 748, `DelayedTriggerRegisteredEvent.event_hook` comments: `// e.g. "EVENT_TURN_START", "EVENT_END_PHASE"`. This directly contradicts FRD 02 line 247, which establishes that card abilities bind strictly to `"EVENT_TURN_END"` and explicitly forbids binding to `"EVENT_END_PHASE"`.
  - **Exploit / Failure Scenario:** Developers implement delayed triggers bound to phase transitions rather than canonical turn boundaries, breaking APNAP trigger ordering during End Phase Sub-Phase 4a.
  - **Remedy:** Correct comment in line 748 to `// e.g. "EVENT_TURN_START", "EVENT_TURN_END"`.

### 🔵 Low / Clarification
- **[DA-NET-012] Underdeveloped .rec Replay File Protobuf Schema**
  - **Requirement Citations:** `FR-SPC-002`, `FR-NET-003`
  - **Vulnerability / Flaw:** Section 6.1 specifies `.rec` replay architecture in prose bullet points (`ReplayHeader`, `InitialState`, `EventLog`), but provides no formal Protobuf message schema for `ReplayFile`.
  - **Exploit / Failure Scenario:** Lack of canonical Protobuf schema allows replay format drift between server archiving and client viewer implementations.
  - **Remedy:** Provide a formal `message ReplayFile` Protobuf schema in `FR-SPC-002`.

## 3. Redundancy & De-Scoping Ledger
| Requirement ID | Fully Subsumed By | Rationale / Recommendation |
| :--- | :--- | :--- |
| `FR-MSG-002` | `FR-MSG-003` | Top-level command wrapper fully unified with concrete action payload schemas in `FR-MSG-003`. |
| `FR-SES-003` | `FR-SET-003`, `FR-MSG-001`, `FR-MSG-004` | First player rules owned by FRD 01; wire schemas owned by `FR-MSG-001`. |
| `FR-SEC-005` | `FR-ENG-002`, `FR-DEK-002` | Shuffling rules owned by FRD 01; PRNG streams owned by FRD 05. |

## 4. Socratic Interrogation Register
1. *Why does the wire protocol use raw string literals for critical state attributes like prompt_type, zone, and event_type instead of Protobuf enums, when the rest of the architecture strictly enforces zero-aliasing enums?*
2. *If a client command is rejected, how can client UI provide localized, user-friendly feedback if error_code is an undocumented integer without an authoritative Protobuf enum?*
