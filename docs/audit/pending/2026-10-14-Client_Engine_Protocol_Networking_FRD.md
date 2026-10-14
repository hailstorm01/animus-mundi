# Product Manager Devil's Advocate Audit: Client-Engine Protocol & Networking FRD (FRD v1.1)

- **Target Document:** [docs/frd/Client_Engine_Protocol_Networking_FRD.md](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/Client_Engine_Protocol_Networking_FRD.md)
- **Auditor:** PM Devil's Advocate (`pm-devils-advocate`)
- **Evaluation Date:** 2026-10-14
- **Status:** PENDING (Active review queue)
- **Verdict:** REVISE WITH BLOCKERS

---

## 1. Executive Summary & Red-Team Assessment

The **Client-Engine Protocol & Networking FRD (v1.1)** presents an authoritative client-server architecture built on WebSockets over TLS, binary Protocol Buffers v3, and double-blind Fog-of-War opaque proxy IDs (`FR-SEC-001`). The recent addition of a 60-second Reserve Time Bank (`FR-SES-005`), per-batch public state hash verification (`FR-PAC-004`), and turn handover animation grace buffers (`FR-PAC-001`) addresses previous network pacing and desync concerns.

However, deep interrogation reveals critical operational vulnerabilities and UX failure modes:
1. **Unintended Forfeiture During Opponent Turns**: `FR-SES-005` specifies a 60-second connection timeout countdown that forfeits a disconnected player even when it is *not their turn* and they face *no pending action prompts*. In mobile or jittery network environments, a player whose connection drops for 61 seconds while their opponent is thinking on a 75-second turn will be erroneously forfeited!
2. **State Snapshot DOS Attack Vector**: When a client detects a public state hash mismatch (`FR-PAC-004`), it automatically requests a `FullStateSnapshot` on that exact frame. Because there is no rate limiting, cooldown, or backoff policy, a desynced or malicious client will flood the server with expensive snapshot requests, causing severe CPU and bandwidth exhaustion.
3. **Unvalidated Command Payloads & Integer Exploits**: `ConvertResourceCommand` (`FR-MSG-003`) accepts an unvalidated raw integer (`tactics_requested`). Negative values or extreme integers create severe underflow exploit risks.

**Recommendation:** Revise `FR-SES-005` so connection timeouts only run when a player has active priority or a pending prompt, rate-limit state snapshot recovery requests, and enforce strict parameter constraints on wire commands.

---

## 2. Severity-Ordered Punch-List

### 🔴 Blocker / Critical

- **[DA-NET-001] Unintended Forfeiture During Opponent's Turn Due to Offline Connection Countdown**
  - **Requirement Citations:** `FR-SES-005`, `FR-TRN-001`, `FR-MOD-001`
  - **Vulnerability / Flaw:** `FR-SES-005` states:
    *"If a disconnected player remains offline with `timeBankMs == 0` and is not currently facing an active prompt or turn, a 60-second connection timeout countdown initiates. If the player fails to reconnect before this timeout expires, the server declares match forfeiture (`FORFEIT_TIMEOUT`)."*
    This rule triggers active forfeiture during the opponent's turn. If an offline player has no remaining reserve bank, but their opponent is currently using their 75-second Main Phase Turn Clock, the offline player will forfeit at second 60, even though they had no priority, no choices to make, and could easily have reconnected before their own turn started.
  - **Exploit / Failure Scenario:**
    1. Player 1 exhausts their 60-second reserve bank earlier in a tightly contested match.
    2. Player 1 ends their turn cleanly. Priority passes to Player 2.
    3. Player 2 has 75 seconds on their Main Phase Turn Clock.
    4. Player 1 is on a mobile device and enters a brief subway tunnel, dropping WebSocket connectivity.
    5. The server notes Player 1 is offline with `timeBankMs == 0` and begins the 60-second timeout countdown.
    6. Player 2 takes 68 seconds to carefully calculate lethal combat math.
    7. At second 60 of Player 1's disconnection (while Player 2 still has 15 seconds remaining on their turn clock), the server triggers `FORFEIT_TIMEOUT` and declares Player 2 the winner.
    8. Player 1 reconnects at second 65—before Player 2's turn has even ended—only to find the match was abruptly forfeited.
  - **Remedy:** Eliminate the 60-second offline countdown during opponent turns. A disconnected player should never forfeit while their opponent has priority unless an interactive choice prompt is actively waiting for them. If a player is offline when their own turn begins, their turn clock and reserve bank drain normally; forfeiture occurs only when their own active time is completely exhausted.

---

### 🟠 High

- **[DA-NET-002] Client Snapshot Request Storm DOS Vulnerability on State Desync**
  - **Requirement Citations:** `FR-PAC-004`, `FR-SES-005`
  - **Vulnerability / Flaw:** `FR-PAC-004` dictates that when the Godot client detects a hash mismatch across its mirrored public state cache, *"the client flags a desync immediately on that frame and automatically requests a `FullStateSnapshot` (`FR-SES-005`)"*. The specification defines zero throttling, minimum request intervals, or maximum snapshot quotas per match.
  - **Exploit / Failure Scenario:**
    1. A subtle floating-point rounding issue or edge-case discrepancy causes the client's local Layer 6 stat calculation to disagree with the server's hash by 1 bit.
    2. The client detects a hash mismatch upon draining the event batch.
    3. On that frame, the client sends a `ReconnectRequest` / `StateSyncRequest` for a full snapshot.
    4. The server serializes and transmits a heavy `FullStateSnapshot` packet.
    5. Because the underlying calculation bug remains in client code, the subsequent frame also evaluates as a hash mismatch.
    6. The client fires 60 full snapshot requests per second over the WebSocket, overwhelming the dedicated server container with serialization load, saturating bandwidth, and crashing the match daemon.
  - **Remedy:** Impose an authoritative rate limiter on `FullStateSnapshot` requests: maximum 1 snapshot per 3,000 ms. If a client records 3 consecutive state hash failures, suppress further automatic snapshots and present the player with a manual "Re-synchronizing match..." overlay.

- **[DA-NET-003] Unchecked Integer Underflow & Validation Gaps in Client Commands**
  - **Requirement Citations:** `FR-MSG-003`, `FR-RES-003`, `FR-ERR-001`
  - **Vulnerability / Flaw:** In `FR-MSG-003`, `ConvertResourceCommand` specifies:
    ```protobuf
    message ConvertResourceCommand {
      int32 tactics_requested = 1;
    }
    ```
    The protobuf definition uses a signed 32-bit integer and specifies no validation invariant requiring `tactics_requested > 0`. If a hacked client transmits a negative value (e.g. `-100`), an improperly guarded server calculation (`requisition_deducted = tactics_requested * 2`) could deduct negative Requisition, effectively generating 200 free Requisition. Similarly, `AttackCommand.target_asset_ids` has no upper bound constraint.
  - **Remedy:** Add strict range validation invariants: `tactics_requested` must be strictly `>= 1` and `<= 50`. Clamp `AttackCommand.target_asset_ids` to maximum 1 target (or maximum 2 for Berserk units). Return `ERR_INVALID_COMMAND_ARGUMENTS` on violation.

---

### 🟡 Medium

- **[DA-NET-004] Mid-Phase Animation Priority Delay & Action Clock Truncation on Inbound Prompts**
  - **Requirement Citations:** `FR-PAC-001`, `FR-PAC-002`, `FR-MOD-001`
  - **Vulnerability / Flaw:** `FR-PAC-001` provides a 5,000 ms Animation Grace Window strictly upon entering the Main Phase. However, during the Main Phase, an active player can execute an extensive sequence of board-clearing maneuvers that triggers 6 to 8 seconds of cascading death and damage animations. If this sequence concludes with a card that prompts the inactive player with an interactive choice (`PlayerChoicePrompt`), the inactive player's 10-second action clock begins ticking down on the server while the client is still rendering prior animations, severely eating into their decision time.
  - **Remedy:** Mandate that when a `PlayerChoicePrompt` is issued, the server action clock countdown does not begin until the client acknowledges that the prompt UI is visually interactive, or grant an automatic 3,000 ms animation buffer on all choice prompts.

- **[DA-NET-005] Missing Guardian Priority Validation Error Code in Wire Protocol Rejection Taxonomy**
  - **Requirement Citations:** `FR-ERR-001`, `FR-CBT-001`, `FR-KWD-002j`
  - **Vulnerability / Flaw:** `FR-ERR-001` details error codes for failed client commands (`ERR_STACK_NOT_EMPTY`, `ERR_INSUFFICIENT_RESOURCES`, `ERR_INVALID_TARGET`, `ERR_ASSET_EXHAUSTED`, `ERR_INVALID_RESOURCE_CHOICE`, `ERR_ACTION_CLOCK_EXPIRED`). However, when an attack declaration illegally bypasses a Main Theater asset with Guardian (`FR-CBT-001`), the rejection returns generic `ERR_INVALID_TARGET`. The client UI cannot distinguish between an invalid target filter vs a mandatory Guardian redirection requirement.
  - **Remedy:** Introduce explicit error code `ERR_GUARDIAN_PRIORITY_VIOLATION` to `FR-ERR-001` so the client UI can flash the defending Guardian asset and guide the player.

---

### 🔵 Low / Clarification

- **[DA-NET-006] Missing Replay Distribution Access Control & Anti-Scouting Policy**
  - **Requirement Citations:** `FR-SPC-002`
  - **Issue:** `FR-SPC-002` specifies that complete `.rec` replay files contain 100% unmasked information (including all opponent deck orders, hand cards, and face-down Scratch identities). The specification does not define access control rules or download availability. In competitive tournament settings (best-of-three matches), an opponent or their coach could fetch the `.rec` file of Game 1 via API immediately after game conclusion, exposing the player's entire 50-card deck list before Game 2 begins.
  - **Clarification Needed:** Explicitly mandate that unmasked `.rec` replays are quarantined and inaccessible via client API until the entire match match-set concludes.

---

## 3. Redundancy & De-Scoping Ledger

| Requirement ID | Fully Subsumed By | Rationale / Recommendation |
| :--- | :--- | :--- |
| `FR-MSG-002` (Client Command Taxonomy) | `FR-MSG-003` (Action Payload Serialization) | `FR-MSG-002` simply declares the `oneof action` wrapper for the concrete protobuf schemas defined in `FR-MSG-003`. Collapse into a single Protobuf specification section. |
| `FR-SES-003` (First Player Determination) | Server Coin Toss | Allowing the winner to select `PLAY_FIRST` or `DRAW_FIRST` requires an extra interactive roundtrip and a 15-second timer before Turn 1. Standardize on automated coin-flip assignment to accelerate match startup. |

---

## 4. Socratic Interrogation Register

1. *Why should a player forfeit a match during their opponent's turn just because their mobile internet dropped for 60 seconds when they had no actions to take?*
2. *What prevents a malicious client from abusing automatic FullStateSnapshot requests to crash dedicated game server containers via packet floods?*
3. *Why does ConvertResourceCommand accept raw integers without protocol-level non-negative validation?*
4. *What prevents tournament players from inspecting unmasked `.rec` replay files between games in a best-of-three match to scout opponent deck lists?*
