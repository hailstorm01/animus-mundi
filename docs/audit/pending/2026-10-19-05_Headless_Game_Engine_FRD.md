# Product Manager Devil's Advocate Audit: Headless Game Engine FRD (Master FRD v1.4)

- **Target Document:** [docs/frd/05_Headless_Game_Engine_FRD.md](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/05_Headless_Game_Engine_FRD.md)
- **Auditor:** PM Devil's Advocate (`pm-devils-advocate`)
- **Evaluation Date:** 2026-10-19
- **Status:** PENDING (Active review queue)
- **Verdict:** REVISE WITH BLOCKERS

## 1. Executive Summary & Red-Team Assessment

The Headless Game Engine specification (`05_Headless_Game_Engine_FRD.md`) provides a comprehensive blueprint for the Go-native rules kernel, establishing transactional event boundaries, PRNG multi-stream isolation, 3-tier runtime entity modeling, and clean stack execution. However, a critical Blocker finding undermines engine-to-rules fidelity: `FR-ENG-035` restricts Berserk N secondary cleave damage strictly to cases where the attacker *"destroyed its primary Main Theater target"*, in direct contradiction to FRD 01 (`FR-CBT-002`, `FR-KWD-002o`), which establishes that Berserk cleave deals flat $N$ damage unconditionally. Furthermore, the Section 8.1 Error Code Registry has drifted significantly from FRD 04 `FR-ERR-001` (omitting 6 active error codes and listing non-existent `ERR_INVALID_TARGET`), `FR-ENG-042` attempts to serialize a non-existent `outcome` field on `MatchConcludedEvent`, and `CurrentPhase` conflates match suspension states with turn phases.

## 2. Severity-Ordered Punch-List

### 🔴 Blocker / Critical
- **[DA-ENG-014] Direct Specification Conflict: Berserk N Kill Requirement in `FR-ENG-035`**
  - **Requirement Citations:** `FR-ENG-035` item 3 (cross-ref `01_Game_Rules_FRD.md` `FR-CBT-002`, `FR-KWD-002o`)
  - **Vulnerability / Flaw:** `FR-ENG-035` item 3 states: *"If an attacking Berserk N unit (FR-KWD-002o) **destroyed its primary Main Theater target during Step 3 combat damage** and a legal pre-bound SECONDARY_BERSERK_TARGET (FR-ENG-031) remains in MainTheaterZone, the engine routes flat N non-combat effect damage..."*. This directly contradicts canonical game rules in FRD 01 `FR-CBT-002` Step 3 and `FR-KWD-002o`, which explicitly mandate that Berserk deals flat $N$ damage **"regardless of the primary combat damage outcome"**. This is a game-breaking rule divergence: implementing `FR-ENG-035` will cause the engine to suppress Berserk cleave whenever the primary defender survives, violating the tabletop rules contract.
  - **Exploit / Failure Scenario:** A player attacks with a Berserk 2 unit against a high-health defender. Both combatants survive. The engine suppresses secondary cleave damage because the primary target was not destroyed. The attacking player files a bug citing `FR-KWD-002o`, where Berserk deals flat damage regardless of primary combat outcome.
  - **Remedy:** Remove the condition *"destroyed its primary Main Theater target during Step 3 combat damage and"* from `FR-ENG-035` item 3, aligning engine execution with FRD 01's unconditional cleave damage rule.

### 🟠 High
- **[DA-ENG-015] Critical Drift in Error Code Registry and Missing Error Codes**
  - **Requirement Citations:** `FR-ENG-041` Section 8.1 (cross-ref `04_Client_Engine_Protocol_Networking_FRD.md` `FR-ERR-001`)
  - **Vulnerability / Flaw:** Section 8.1 claims *"The headless engine module enforces zero-drift error parity with the client-engine networking protocol"*, but contains major discrepancies:
    1. Line 631 lists `ERR_INVALID_TARGET` as a Warning for stack pop invalidation, but in FRD 04 `FR-ERR-001`, this code does not exist: upfront invalid selection is `ERR_INVALID_TARGET_SELECTION = 3` and pop-time invalidation is `WARN_TARGET_INVALIDATED_FIZZLE = 30`.
    2. The Section 8.1 registry completely omits 6 active error codes defined in FRD 04 `enum ErrorCode`: `ERR_NOT_ACTIVE_PLAYER = 9`, `ERR_CARD_NOT_IN_HAND = 10`, `ERR_ILLEGAL_ZONE_TARGET = 11`, `ERR_MAX_PER_TURN_EXCEEDED = 12`, `ERR_VERSION_MISMATCH = 23`, and `ERR_CATALOG_CHECKSUM_MISMATCH = 24`.
  - **Exploit / Failure Scenario:** The engine rejects a command for `ERR_MAX_PER_TURN_EXCEEDED` or `ERR_NOT_ACTIVE_PLAYER`, but because the engine table lacks these codes, it emits an unmapped error or falls back to `ERR_INVALID_COMMAND_ARGUMENTS`, degrading client UI feedback.
  - **Remedy:** Update Section 8.1 of FRD 05 to replace `ERR_INVALID_TARGET` with `WARN_TARGET_INVALIDATED_FIZZLE` and `ERR_INVALID_TARGET_SELECTION`, and add all 6 missing error codes to achieve genuine zero-drift wire parity.

- **[DA-ENG-016] Wire Schema Contradiction in `MatchConcludedEvent` Fields**
  - **Requirement Citations:** `FR-ENG-042` item 2 (cross-ref `04_Client_Engine_Protocol_Networking_FRD.md` `FR-MSG-004`)
  - **Vulnerability / Flaw:** `FR-ENG-042` item 2 specifies that upon a Match Draw: *"Conforming to FR-MSG-004, the engine emits `MatchConcludedEvent` with `conclusion_reason = "HQ_DESTROYED"`, `is_match_draw = true`, and **`outcome = RESULT_MATCH_DRAW`**."* However, `MatchConcludedEvent` in FRD 04 `FR-MSG-004` (lines 662-667) contains **no `outcome` field** (it contains only `winner_player_id`, `is_match_draw`, and `conclusion_reason`). `outcome` exists only on `ReplayHeader.match_outcome`. Attempting to serialize `outcome` on `MatchConcludedEvent` will fail Protobuf schema validation.
  - **Exploit / Failure Scenario:** The engine attempts to emit `MatchConcludedEvent` with an `outcome` property; serialization fails or the field is dropped by the client protobuf decoder.
  - **Remedy:** Remove `outcome = RESULT_MATCH_DRAW` from `FR-ENG-042` item 2, noting that `is_match_draw = true` and `winner_player_id = ""` are the canonical wire representations for a Match Draw.

### 🟡 Medium
- **[DA-ENG-017] Ambiguous `CurrentPhase` State Serialization in `StateSyncSnapshot`**
  - **Requirement Citations:** `FR-ENG-006`, `FR-ENG-008` (cross-ref `04_Client_Engine_Protocol_Networking_FRD.md` `FR-MSG-001`, `FR-MSG-004`)
  - **Vulnerability / Flaw:** `FR-ENG-006` defines `CurrentPhase` with values `PHASE_READY`, `PHASE_DRAW`, `PHASE_MAIN`, `PHASE_END`, `STATE_AWAITING_CHOICE`, and `STATE_CONCLUDED`. It mixes turn phases with match suspension states (`STATE_AWAITING_CHOICE`). When serializing `StateSyncSnapshot.current_phase` (`FR-MSG-001`), if `CurrentPhase == STATE_AWAITING_CHOICE`, it is ambiguous whether `current_phase` transmits `"STATE_AWAITING_CHOICE"` (which is not a legal phase in `PhaseTransitionEvent`) or retains the underlying phase (e.g. `"MAIN_PHASE"`) while populating `pending_prompt`.
  - **Exploit / Failure Scenario:** A client reconnects during a choice prompt; `StateSyncSnapshot.current_phase` arrives as `"STATE_AWAITING_CHOICE"`, causing the client UI phase indicator to display an unmapped phase string.
  - **Remedy:** Separate match lifecycle execution state (`STATE_RUNNING`, `STATE_AWAITING_CHOICE`, `STATE_CONCLUDED`) from turn phase (`PHASE_READY`, `PHASE_DRAW`, `PHASE_MAIN`, `PHASE_END`) in `MatchState`.

### 🔵 Low / Clarification
_None._

## 3. Redundancy & De-Scoping Ledger
| Requirement ID | Fully Subsumed By | Rationale / Recommendation |
| :--- | :--- | :--- |
| `FR-ENG-003` | `FR-ENG-002` | Seed derivation with salts is fully incorporated into the PRNG architecture in FR-ENG-002. |
| `FR-ENG-014` & `FR-ENG-015` | `FR-TRG-003`, `FR-TGT-003`, `FR-ENG-012` | APNAP ordering and two-stage target verification are canonically owned by FRD 01. |
| `FR-ENG-019` & `FR-ENG-020` | `FR-ENG-018` | Dynamic filter predicates and ValueToken evaluation are handled by the AST interpreter in FR-ENG-018. |
| `FR-ENG-024` | `FR-ENG-012`, `FR-ENG-013`, `FR-MNV-001` | Maneuver execution and STACK zone transition are owned by stack and cost engines. |
| `FR-ENG-026`–`FR-ENG-028` | `FR-ENG-025`, `FR-ENG-041` | Stateless evaluation, timestamp ordering, and clamping invariants are owned by FR-ENG-025 and FR-ENG-041. |
| `FR-ENG-030` | `FR-CBT-001`, `FR-ENG-031` | Spatial combat restrictions are owned by FRD 01 and Step 1 attack validation in FR-ENG-031. |
| `FR-ENG-034` | `FR-REP-004`, `FR-ENG-035` | 5-stage damage mitigation formulas and state sweeps are owned by FRD 01 and FR-ENG-035. |

## 4. Socratic Interrogation Register
1. *If `MatchState` uses an in-place mutable architecture with delta-rollback mementos, why does `BeginTransaction()` record delta-snapshots of touched PRNG stream states rather than deriving PRNG call sequences from deterministic transaction sequence counters? Could an un-tracked PRNG consumption inside an ability AST visitor permanently desync a match rollback?*
