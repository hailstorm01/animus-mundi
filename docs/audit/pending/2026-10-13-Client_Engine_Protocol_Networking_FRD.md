# Product Manager Devil's Advocate Audit: Client-Engine Protocol & Networking FRD (FRD v1.1)

- **Target Document:** [Client_Engine_Protocol_Networking_FRD.md](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/Client_Engine_Protocol_Networking_FRD.md)
- **Auditor:** PM Devil's Advocate (`pm-devils-advocate`)
- **Date:** 2026-10-13
- **Verdict:** REVISE WITH BLOCKERS

---

## 1. Executive Summary & Red-Team Assessment

The **Client-Engine Protocol & Networking FRD (v1.1)** presents a well-specified, authoritative client-server architecture. The adoption of WebSockets over TLS via Godot's native `WebSocketPeer`, binary Protocol Buffers v3 serialization, and double-blind Fog-of-War opaque proxy IDs for Scratch Company units (`FR-SEC-001`) demonstrates strong security and architectural hygiene.

However, deep interrogation reveals critical operational vulnerabilities and UX failure modes:
1. **Disconnect Stalling & Griefing Exploits**: Retrospective prompt clock compensation (`FR-SES-005`) lacks a cumulative per-match disconnect quota, enabling malicious players to freeze timers and stall matches indefinitely via deliberate micro-disconnections.
2. **Animation Desynchronization & Turn Theft**: Decoupling the client animation queue (`FR-PAC-001`) from an immediate, unyielding server-side 75-second turn clock (`FR-TRN-001`) leads to "turn theft," where players lose their turn while locked in mandatory animation playback.
3. **Mid-Phase Desync Blindspot**: Public state checksum verification (`FR-PAC-004`) executes only at phase transitions and stack unwinding. Because inherent actions (March, Attack, Resource Generation) bypass the stack, visual desynchronizations can persist undetected throughout the entire Main Phase.
4. **Premature Spectator Scope**: Full real-time live spectator streaming with dynamic Fog-of-War masking (`FR-SPC-001`) represents significant over-engineering and delivery risk for an initial game release when deterministic file replays (`FR-SPC-002`) already fulfill broadcast needs.

**Recommendation:** Revise network clock management, implement a strict disconnect quota, bind animation playback to turn timers, and defer live spectator infrastructure to v2.0.

---

## 2. Severity-Ordered Punch-List

### 🔴 Blocker / Critical

- **[DA-RULES-001] Disconnect Retrospective Compensation Enables Infinite Stalling & Griefing**
  - **Requirement Citations:** `FR-SES-005`, `FR-NET-004`, `FR-MOD-001`
  - **Vulnerability / Flaw:** `FR-SES-005` specifies that upon detecting a client disconnection (within 3,000 ms), the server pauses choice prompt action clocks and retrospectively restores undetected silent time:
    ```text
    remainingClockMs = max(1000, deadlineEpochMs - lastReceivedClientPacketEpochMs)
    ```
    Critically, the specification defines **no cumulative disconnect budget** per player, **no limit on disconnections per turn**, and does not define whether the 75-second Main Phase Turn Clock pauses during disconnections.
  - **Exploit / Failure Scenario:**
    1. A losing player faces an opponent's lethal board state or is prompted with an interactive choice (`PlayerChoicePrompt`).
    2. At second 14 of their 15-second action clock, the player terminates their network connection (or throttles WebSocket frames).
    3. The server waits 3,000 ms, detects silence, pauses the prompt clock, and enters the 10-second Disconnect Grace Window.
    4. The player reconnects at second 9 of the grace window, rehydrates state, and receives their prompt with the remaining clock restored.
    5. The player repeats this cycle indefinitely (disconnecting every 14 seconds), holding the winning opponent hostage in an endless, unfinishable match without triggering forfeit.
  - **Remedy:** Implement a strict cumulative match disconnect bank (e.g. maximum 60 seconds total disconnect time per match across all occurrences). Once exhausted, any further dropped connection results in immediate match forfeit.

- **[DA-RULES-002] Decoupled Animation Queue Causes Turn Theft by Server Turn Clock**
  - **Requirement Citations:** `FR-PAC-001`, `FR-PAC-002`, `FR-TRN-001`, `FR-MSG-004`
  - **Vulnerability / Flaw:** Server state is committed instantaneously, and atomic event streams are broadcast to clients (`FR-PAC-001`). The Godot client enqueues events into a FIFO `AnimationEventQueue`. Meanwhile, the server enforces an authoritative, unyielding 75-second Main Phase Turn Clock (`FR-TRN-001`).
  - **Exploit / Failure Scenario:**
    1. Player 1 executes a complex sequence involving 4 inherent actions, 2 Maneuvers, and multiple cascading triggers.
    2. The server processes this entire transaction in under 5 milliseconds and starts Player 2's turn clock.
    3. Player 2's client is still playing 8-12 seconds of visual animations (card draws, board marches, 5-stage damage flashes, death dissolves).
    4. Even with 3.0x animation fast-forwarding (`FR-PAC-002`), Player 2's effective Main Phase input window is significantly truncated.
    5. In worst-case scenarios with extensive trigger chains, Player 2's turn timer expires on the server (`AutoEndTurnCommand`) before the client finishes rendering the previous turn's combat outcomes, effectively skipping Player 2's turn ("turn theft").
  - **Remedy:** Mandate that the server pause the active player's Main Phase turn clock until both clients have acknowledged receipt and completion of the preceding phase transition events, or establish a maximum animation queue buffer cap that auto-extends the turn clock.

---

### 🟠 High

- **[DA-XP-001] Inherent Action Desync Blindspot During Main Phase**
  - **Requirement Citations:** `FR-PAC-004`, `FR-CAP-001`, `FR-INH-001`
  - **Vulnerability / Flaw:** Public State Hash verification (`FR-PAC-004`) runs strictly *"at the conclusion of every phase transition and stack unwinding sequence"*.
  - **Player Experience Breakdown:**
    Inherent actions (Marching units, generating Requisition/Tactics, converting resources, executing Scratch 1) execute directly as state operations and **never touch the stack** (`FR-CAP-001`, `FR-INH-001`).
    If a packet drop, race condition, or local GDScript simulation glitch occurs during a rapid sequence of inherent actions in the Main Phase, the client's local board cache desynchronizes from the server.
    Because no stack unwinding or phase transition occurs, the desync is **never detected** by `FR-PAC-004` until the player ends their turn. During this entire 75-second window, the player may see wrong unit readiness, incorrect resource pools, or phantom positions, leading to rejected commands (`CommandRejection`) with zero automatic snapshot recovery.
  - **Remedy:** Include `PublicStateHash` inside the batch trailer of every `ServerEventStream` packet resulting from inherent actions, guaranteeing instant desync detection within a single frame of any state divergence.

- **[DA-XP-002] Plaintext Deck List Submission Exposes Strategic Meta in Handshake**
  - **Requirement Citations:** `FR-SES-002`, `FR-MSG-003` (`MatchJoinCommand`)
  - **Vulnerability / Flaw:** The client submits its complete list of 50+ card blueprint IDs directly within the `MatchJoinCommand` packet (`FR-MSG-003`).
  - **Player Experience Breakdown:** Transmitting full decklists over client-initiated match joins allows client-side proxy sniffing or rogue API interception during connection negotiation. In competitive ranked or tournament environments, players running local packet inspection can inspect the raw payload or exploit matchmaking proxies to identify opponents' exact deck contents prior to the first turn.
  - **Remedy:** Shift deck registration to the backend matchmaking/session service. Clients should submit only a cryptographically signed `deckRegistrationToken` or hash; the dedicated match container retrieves the validated decklist directly from internal backend storage, concealing it from client-side network exposure.

---

### 🟡 Medium

- **[DA-SCOPE-001] Premature Real-Time Live Spectator Architecture**
  - **Requirement Citations:** Domain 6 (`FR-SPC-001`)
  - **Vulnerability / Flaw:** `FR-SPC-001` specifies a complex live spectator protocol with real-time Fog-of-War scrubbing, opaque proxy ID synthesis, and dedicated spectator WebSocket tokens to prevent ghosting and sniping.
  - **Assessment:** For a v1.0 / v1.1 digital CCG release, architecting, load-testing, and securing a live spectator WebSocket distribution tier across dedicated server containers is major scope creep. Digital card games typically launch with client-side match replays (`FR-SPC-002`), deferring live spectator infrastructure until tournament and esports features are required.
  - **Remedy:** De-scope `FR-SPC-001` from the core engine networking milestone. Rely on historical replay files (`.rec` files via `FR-SPC-002`) for tournament broadcasting, delayed spectating, and anti-cheat analysis.

- **[DA-XP-003] Rigid ConvertResourceCommand Parameter Contract**
  - **Requirement Citations:** `FR-MSG-003`, `FR-RES-003`
  - **Vulnerability / Flaw:** `ConvertResourceCommand` takes `int32 requisition_spent = 1;` and requires it to be a positive multiple of 2.
  - **Player Experience Breakdown:** If a client UI bug or user macro submits an odd number (e.g. 3) or negative value, the command is rejected with an error. From a UX standpoint, the player's intent is almost always *"I want N Tactics"*. Passing Requisition spent rather than Tactics desired creates an unnecessary validation failure mode.
  - **Remedy:** Refactor `ConvertResourceCommand` to specify `int32 tactics_requested = 1;`, with the server deterministically deducting `tactics_requested * 2` Requisition.

---

### 🔵 Low / Clarification

- **[DA-CLARIFY-001] Lack of Ping Frequency Throttling & DDoS Protection**
  - **Requirement Citations:** `FR-NET-004`
  - **Issue:** `FR-NET-004` states clients transmit empty `Ping` frames every 1,000 ms. Does the server enforce a rate limiter on ping frames? A hacked or malfunctioning client could flood the server process with 10,000 pings per second over the WebSocket, causing memory and CPU exhaustion on the dedicated container.
  - **Clarification Needed:** Specify a rate-limiting policy (e.g. drop connections exceeding 5 pings per second).

- **[DA-CLARIFY-002] Disconnect Grace Period Interaction with Pre-Game Mulligan**
  - **Requirement Citations:** `FR-SES-004`, `FR-SES-005`
  - **Issue:** If a player disconnects during the 25-second Mulligan countdown (`FR-SES-004`), does the server pause the countdown, commit zero discards immediately, or wait for the 10-second grace window?
  - **Clarification Needed:** Explicitly state the pre-game disconnect fallback policy (recommend defaulting to keep all 6 cards upon mulligan timeout).

---

## 3. Redundancy & De-Scoping Ledger

| Requirement ID | Fully Subsumed By | Rationale / Recommendation |
| :--- | :--- | :--- |
| `FR-MSG-002` (Client Command Taxonomy) | `FR-MSG-003` (Action Payload Serialization) | `FR-MSG-002` simply declares the `oneof action` wrapper for the concrete protobuf schemas defined in `FR-MSG-003`. Collapse into a single Protobuf specification section. |
| `FR-SPC-001` (Live Spectator Fog-of-War) | *De-Scope from v1.1* | Live spectator distribution is premature. Offline `.rec` replay architecture (`FR-SPC-002`) fully addresses analytical and review needs. |
| `FR-SES-002` (Deck Registration) | `FR-SET-001` (Deck Construction Rules) | Server deck legality validation simply mirrors `FR-SET-001`. Cross-reference `FR-SET-001` instead of duplicating validation rules. |

---

## 4. Socratic Interrogation Register

1. *Why does the client send the entire deck blueprint list over the WebSocket in `MatchJoinCommand` rather than passing a secure server-validated deck ID from matchmaking?*
2. *What prevents a bad actor from using deliberate micro-disconnects to stall interactive choice prompts indefinitely and force an opponent to concede out of boredom?*
3. *If an animation sequence takes 12 seconds to play on the client, how does the player retain their full 75-second turn when the server clock has already begun ticking down?*
4. *Why is real-time live spectator streaming included in the core networking specification before basic match synchronization and disconnect recovery are battle-tested in production?*
5. *Why are state hashes computed only during stack unwinding when the vast majority of turns consist primarily of stack-bypassing inherent actions?*
