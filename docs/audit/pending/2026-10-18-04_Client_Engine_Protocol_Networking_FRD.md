# Product Manager Devil's Advocate Audit: Client-Engine Protocol & Networking FRD

- **Target Document:** [docs/frd/04_Client_Engine_Protocol_Networking_FRD.md](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/04_Client_Engine_Protocol_Networking_FRD.md)
- **Auditor:** PM Devil's Advocate (`pm-devils-advocate`)
- **Evaluation Date:** 2026-10-18
- **Status:** PENDING (Active review queue)
- **Verdict:** CONDITIONAL PASS

## 1. Executive Summary & Red-Team Assessment

This audit evaluates `04_Client_Engine_Protocol_Networking_FRD.md` (FRD v1.2) specifically focusing on requirement duplication, copycat definitions, and domain authority boundaries across the master specification suite.

`04_Client_Engine_Protocol_Networking_FRD.md` canonically owns TLS WebSocket transport, binary Protocol Buffers wire schemas (`ccg.network`), session handshake tokens, connection liveness, double-blind Scratch Company proxy scrubbing on the wire, client-side animation FIFO queue pacing, desync hash reporting, and `.rec` replay wire formatting.

However, red-teaming reveals two requirements in FRD 04 that re-define rules and architecture canonically owned by `01_Game_Rules_FRD.md` and `05_Headless_Game_Engine_FRD.md` without adding transport or wire protocol functionality:
1. `FR-SES-003` (First-Player Determination Protocol) clones the game rule from `FR-SET-003` without adding unique session contracts beyond referencing wire message schemas already canonically codified in Section 4.1 (`FR-MSG-001`).
2. `FR-SEC-005` (Authoritative Server-Side RNG) re-defines server randomness, creating an architectural contradiction by demanding CSPRNG for all match permutations when `05_Headless_Game_Engine_FRD.md` (`FR-ENG-002`) mandates deterministic, seeded pseudo-random streams (PCG32/SplitMix64) for replay reproduction.

## 2. Severity-Ordered Punch-List

### 🔴 Blocker / Critical
_None._

### 🟠 High
_None._

### 🟡 Medium
- **[DA-NET-003] Duplicate Setup Rule in FR-SES-003 (First-Player Determination Protocol)**
  - **Requirement Citations:** `FR-SES-003`
  - **Vulnerability / Flaw:** `FR-SES-003` re-states the game rule for First Player Determination (random 50/50 roll, 15s timer, `PLAY_FIRST` / `DRAW_FIRST` choices, default to `PLAY_FIRST`), which is canonically owned by `01_Game_Rules_FRD.md` (`FR-SET-003`). The Protobuf wire messages (`FirstPlayerChoicePrompt`, `FirstPlayerDetermined`) are already canonically defined in `FR-MSG-001` (Section 4.1). `FR-SES-003` adds no unique transport layer or session state logic.
  - **Exploit / Failure Scenario:** If game rules adjust first-player selection timings or defaults in `FR-SET-003`, `FR-SES-003` creates a dual-maintenance hazard.
  - **Remedy:** Mark `FR-SES-003` as subsumed by `FR-SET-003` (Game Rules) and `FR-MSG-001` / `FR-MSG-004` (Wire Schemas).

- **[DA-NET-004] Redundant Server RNG Invariant in FR-SEC-005 and Architectural CSPRNG Conflict**
  - **Requirement Citations:** `FR-SEC-005`
  - **Vulnerability / Flaw:** `FR-SEC-005` mandates that all random permutations are executed on the server using CSPRNG. However, randomness rules are canonically owned by `01_Game_Rules_FRD.md` (`FR-DEK-002: Deck Shuffling Rules`), and server PRNG architecture is canonically owned by `05_Headless_Game_Engine_FRD.md` (`FR-ENG-002: Isolated Multi-Stream PRNG Architecture & Seed Derivation`). Moreover, `FR-SEC-005`'s requirement for CSPRNG directly contradicts `FR-ENG-002`'s requirement for deterministic seeded SplitMix64/PCG32 PRNG streams necessary for bit-for-bit replay validation (`FR-ENG-038`, `FR-SPC-002`). The wire protocol (`ClientCommand` in `FR-MSG-002`) already precludes client-provided seeds.
  - **Exploit / Failure Scenario:** Engine implementers reading `FR-SEC-005` might use a cryptographic non-seeded RNG for in-game rolls, breaking deterministic replay validation and desync detection.
  - **Remedy:** Mark `FR-SEC-005` as subsumed by `FR-ENG-002` and `FR-DEK-002`, eliminating the CSPRNG contradiction and redundant requirement.

### 🔵 Low / Clarification
_None._

## 3. Redundancy & De-Scoping Ledger
| Requirement ID | Fully Subsumed By | Rationale / Recommendation |
| :--- | :--- | :--- |
| `FR-SES-003` | `FR-SET-003`, `FR-MSG-001` | First-player choice rule belongs to FRD 01; wire schemas belong to FRD 04 Section 4.1. |
| `FR-SEC-005` | `FR-ENG-002`, `FR-DEK-002` | PRNG engine architecture belongs to FRD 05; shuffling rules belong to FRD 01. |

## 4. Socratic Interrogation Register
1. *Should `04_Client_Engine_Protocol_Networking_FRD.md` focus strictly on transport, Protobuf schemas, wire serialization, pacing, and packet sanitization, leaving all match initialization rules and RNG algorithms entirely to FRD 01 and FRD 05?*
