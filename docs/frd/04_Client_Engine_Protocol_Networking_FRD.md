# Client-Engine Protocol & Networking - Functional Requirements Document (FRD v1.3)
**Master Network Topology, Message Taxonomy, Fog-of-War & Wire Protocol Specification**

---

## EXECUTIVE SUMMARY & REVISION HISTORY

This document defines the network architecture, communication transport, client-server message protocol, state synchronization model, and fog-of-war security boundary for the digital Collectible Card Game (CCG). It serves as the companion networking specification to the **Game Rules FRD (`01_Game_Rules_FRD.md`)**, **Card Data Pipeline FRD (`02_Card_Data_Pipeline_FRD.md`)**, **Card Compiler FRD (`03_Card_Compiler_FRD.md`)**, and **Headless Game Engine FRD (`05_Headless_Game_Engine_FRD.md`)**.

### Revision Summary (v1.2 -> v1.3):
1. **Protobuf Tag Collision Fix**: Renumbered `ClientCommand.action` match orchestration and session recovery tags starting at `3` (`match_join = 3` through `request_state_snapshot = 7`) to eliminate duplicate field tags `1` and `2` in `FR-MSG-002` (`DA-NET-001`).
2. **Stack-Pushed Tactical Inherent Actions Alignment**: Separated Off-Stack Economic Operations (`GenerateResourceCommand`, `ConvertResourceCommand`) from Stack-Pushed Tactical Inherent Actions (`MarchCommand`, `AttackCommand`, `DeployScratchCommand`) across `FR-MSG-002`, `FR-MSG-003`, and `FR-PAC-004`, and expanded `StackPushedEvent.ability_type` in `FR-MSG-004` to include `"ATTACK"`, `"MARCH"`, and `"DEPLOY_SCRATCH"` (`FR-CAP-001`, `FR-INH-001`, `DA-NET-002`).
3. **Handshake Catalog Checksum & Session Wire Schemas**: Added 256-bit `catalogChecksum` verification (`FR-DAT-004`, `FR-CMP-025`, `FR-CMP-026`) to `FR-SES-001`, and defined explicit Protobuf schemas and envelope/command tags for `ClientHandshake`, `HandshakeAck`, `HandshakeReject`, `ReconnectRequest`, and `RequestStateSnapshot` in `FR-MSG-001`, `FR-MSG-002`, and `FR-MSG-003` (`DA-NET-003`).
4. **Berserk Secondary Target Role & Active/Non-Active Prompt Clocks**: Added `TARGET_ROLE_SECONDARY_BERSERK_TARGET = 7` to `enum TargetRole` in `FR-MSG-003` (`FR-TGT-101`, `FR-KWD-002o`) and aligned `FR-PAC-001` `PlayerChoicePrompt` deadlines with `FR-MOD-001` (15,000 ms for the active player, 10,000 ms for the non-active player, plus the 3,000 ms animation buffer; `DA-NET-004`).
5. **Choice Option Blueprint Identity & Full Prompt Taxonomy**: Added `string blueprint_id = 4` to `ChoiceOption` in `FR-MSG-005` for hidden-zone reveals (`FR-SEC-004`) and expanded `PlayerChoicePrompt.prompt_type` to include `"SEARCH_DECK"`, `"REVEAL_HAND"`, `"REPLACEMENT_CONFLICT_SELECT"` (`FR-REP-001`), and `"OPTIONAL_TRIGGER_CONFIRM"` (`FR-MOD-001`, `DA-NET-005`).

### Revision Summary (v1.1 -> v1.2):
1. **Opponent-Turn Forfeiture Protection**: Eliminated erroneous 60s forfeiture timers during opponent turns; disconnection counters draw down strictly when the offline player holds active priority or an unanswered prompt (`FR-SES-005`, `DA-NET-001`).
2. **State Snapshot Flood Protection**: Imposed authoritative rate-limiting of 1 `FullStateSnapshot` per 3,000 ms with backoff after 3 consecutive failures to neutralize DOS packet storms (`FR-SES-005`, `FR-PAC-004`, `DA-NET-002`).
3. **Wire Command Boundary Clamping**: Mandated range validation for `tactics_requested` [1..50] and clamped `AttackCommand.target_asset_ids` to 1 (or 2 for Berserk), rejecting violations with `ERR_INVALID_COMMAND_ARGUMENTS` (`FR-MSG-003`, `FR-ERR-001`, `DA-NET-003`).
4. **Choice Prompt Animation Buffering**: Added an automatic 3,000 ms grace window on interactive `PlayerChoicePrompt` issuances before the 15s (active player) / 10s (non-active player) decision clock initiates, preventing cascading animation truncation (`FR-PAC-001`, `DA-NET-004`).
5. **Guardian Priority Wire Error Code**: Introduced explicit `ERR_GUARDIAN_PRIORITY_VIOLATION` to `FR-ERR-001`, enabling client UI to flash the mandatory Guardian target asset (`DA-NET-005`).
6. **Replay Anti-Scouting Quarantine**: Quarantined unmasked `.rec` replay telemetry on the server until match-set (e.g. Best-of-3) conclusion, preventing opponent deck scouting during tournament sets (`FR-SPC-002`, `DA-NET-006`).
7. **Redundancy Cleanup**: Folded `FR-MSG-002` into unified `FR-MSG-003` schema specification.

---

## DOMAIN 1: NETWORK TOPOLOGY & TRANSPORT ARCHITECTURE

### Section 1.1: Hosting & Transport Infrastructure

#### FR-NET-001: Cloud Infrastructure & Server Runtime
- **Hosting Environment**: Dedicated match instances hosted on Google Cloud Platform (GCP) utilizing Google Kubernetes Engine (GKE) managed via Agones game server orchestration, or serverless autoscaled containers via GCP Cloud Run for turn-based match loops.
- **Stateful Match Isolation**: Each active match is allocated to an isolated, stateful engine container process managing a single match pair. Match memory is strictly isolated.

#### FR-NET-002: Transport Layer Protocol
- **Transport Standard**: Secure WebSockets over TLS (`wss://`).
- **Client Integration**: Utilizes Godot Engine 4.x's built-in `WebSocketPeer` interface, guaranteeing cross-platform compatibility across Windows, Linux, macOS, iOS, Android, and WebGL/WebAssembly exports without third-party native socket wrappers.
- **Port & Firewall Traversal**: Operates over standard HTTPS/WSS port `443` to ensure zero friction with corporate, campus, and residential firewalls.

#### FR-NET-003: Wire Serialization Format
- **Payload Encoding**: Protocol Buffers v3 (Protobuf) binary encoding as the production standard.
- **Development Fallback**: The server supports a content-negotiation handshake permitting uncompressed JSON payloads for development, debugging, and mock test harnesses.
- **Binary Endianness**: All numeric fields conform to standard Network Byte Order (Big-Endian) or Protobuf varint standards.

#### FR-NET-004: Connection Liveness, Ping/Pong & Latency Tracking
- **Heartbeat Cadence**: The client transmits an empty `Ping` frame every 1,000 ms. The server responds immediately with a `Pong` frame echoing the client's local timestamp and appending the current server timestamp (`serverEpochMs`).
- **Heartbeat Rate-Limiting & DoS Protection**: The server enforces a strict sliding-window rate ceiling permitting a maximum of **5 `Ping` frames per 1,000 ms window** (absorbing normal network jitter/packet bunching). If a client exceeds 5 pings in a rolling 1-second interval, the server terminates the WebSocket connection with close code `1008` (`POLICY_VIOLATION`) and flags the session for forfeit due to protocol abuse.
- **Round-Trip Time (RTT)**: The client computes rolling smoothed RTT to dynamically adjust local animation playback rates.
- **Connection Timeout**: If no frames (data or heartbeat) are received from a client for 3,000 ms, the server flags the connection as severed and initiates disconnection handling and reserve bank tracking (`FR-SES-005`).

---

## DOMAIN 2: SESSION LIFECYCLE & CONNECTION NEGOTIATION

```text
┌──────────────┐                               ┌──────────────┐
│ Godot Client │                               │  GCP Server  │
└──────┬───────┘                               └──────┬───────┘
       │                                              │
       │ 1. Connect (WSS Handshake + Catalog Hash)    │
       ├─────────────────────────────────────────────►│
       │ 2. HandshakeAck (Protocol, Engine & Catalog) │
       │◄─────────────────────────────────────────────┤
       │ 3. MatchJoinRequest (MatchToken + Token)     │
       ├─────────────────────────────────────────────►│
       │ 4. MatchValidate & DeckLegality (FR-SET-001) │
       │    [Server fetches deck from backend storage]│
       │ 5. MatchStartNotification (Coin Toss/P1)     │
       │◄─────────────────────────────────────────────┤
       │ 6. MulliganPrompt (6 Initial Cards)          │
       │◄─────────────────────────────────────────────┤
       │ 7. MulliganCommit (Selected indices set aside)│
       │├─────────────────────────────────────────────►│
       │ 8. Turn 1 Setup (Draws, HQ placement)        │
       │◄─────────────────────────────────────────────┤
```

### Section 2.1: Handshake & Match Entry

#### FR-SES-001: Connection Handshake & Version Verification
- Upon WebSocket connection establishment, the client transmits a `ClientHandshake` (`FR-MSG-001`):
  - `clientEngineVersion`: String (e.g., `"3.0.0"`). Must **strictly match** the server's engine version.
  - `protocolVersion`: Integer (e.g., `1`). Must **strictly match** the server's wire protocol version.
  - `catalogChecksum`: String (64-character lowercase hexadecimal SHA-256 digest of the canonical `blueprints.json` card catalog; `FR-DAT-004`, `FR-CMP-025`, `FR-CMP-026`). Must **strictly match** the server's authoritative card catalog checksum.
  - `matchAuthToken`: Cryptographically signed JWT issued by the matchmaking API.
- If the `protocolVersion`, `clientEngineVersion`, or `catalogChecksum` does not strictly match, the server immediately returns a `HandshakeReject` packet (`FR-MSG-001`) specifying the rejection error code (`ERR_VERSION_MISMATCH` or `ERR_CATALOG_CHECKSUM_MISMATCH`) and closes the connection.

#### FR-SES-002: Secure Match Entry & Server-Side Deck Validation
- In compliance with **FR-SET-001**:
  - To prevent client-side packet sniffing and strategic meta disclosure during connection negotiation, clients **never** transmit card blueprint arrays over the game WebSocket.
  - The client submits `MatchJoinCommand` containing strictly the `match_token` (JWT issued by matchmaking) and an optional signed `deck_registration_token`.
  - The authoritative match container retrieves the player's registered deck directly from secure backend session storage via internal match coordinator APIs.
  - The server validates the retrieved deck strictly against canonical deck construction legality in **FR-SET-001** (HQ faction designation, main deck faction purity, minimum 50 cards, rarity copy limits, and prohibition of `neutral_scratch_company`).
  - If validation fails, match initialization aborts with `DECK_LEGALITY_VIOLATION`.

#### FR-SES-003: First-Player Determination Protocol
- In compliance with **FR-SET-003**:
  - The server executes a secure random roll (0 or 1) to determine the initial decision-maker.
  - The server transmits a `FirstPlayerChoicePrompt` to the winning player with a 15-second timer (`deadlineEpochMs`).
  - The player selects `PLAY_FIRST` or `DRAW_FIRST`.
  - If the timer expires, the server defaults to `PLAY_FIRST`.
  - The server broadcasts `FirstPlayerDetermined` to both clients.

#### FR-SES-004: Mulligan Negotiation Protocol
- In compliance with **FR-SET-002**:
  1. The server initializes each player's deck, executes a cryptographically secure shuffle (`FR-DEK-002`), and generates an initial hand of 6 cards for each player.
  2. The server sends `MulliganPrompt` to each player containing only their own 6 cards.
  3. Players have a 25-second countdown to return `MulliganCommitCommand` containing an array of zero or more card instance IDs to set aside.
     - **Continuous Progression & Disconnect Fallback**: Conforming to the continuous timer progression in `FR-SES-005`, the 25-second mulligan countdown never pauses upon disconnect. If the 25-second countdown expires without a commit (whether due to client timeout or disconnection), the server deterministically defaults to **keep all 6 cards** (zero cards swapped).
     - **Pre-Game Reserve Preservation**: The player's 60-second Reserve Time Bank (`FR-SES-005`) is not drawn down during the pre-game mulligan. The match transitions directly to Turn 1 setup without stalling the opponent, preserving the player's full 60s reserve bank for in-match rehydration.
  4. Upon receiving both commits (or timer expiration):
     - The server draws replacement cards for the set-aside cards.
     - The set-aside cards are shuffled back into the deck.
     - The server transmits `MulliganResultNotification` containing the finalized opening hand to each respective player and notifying opponents of the number of cards swapped.

#### FR-SES-005: Disconnection Handling, Unified Reserve Time Bank & State Rehydration
- In compliance with **FR-MOD-001** and **FR-NET-004**:
  - **Unified Reserve Time Bank**: Rather than pausing clocks or maintaining an opaque disconnect mode, the server allocates each player a cumulative **60,000 ms (60-second) Reserve Time Bank** (`timeBankMs`) per match.
  - **Authoritative Continuous Timer Progression**: Neither the 75-second Main Phase Turn Clock (`FR-TRN-001`) nor the choice prompt clocks (`FR-MSG-005`) pause when a player disconnects or encounters network silence.
  - **No Opponent-Turn Disconnection Forfeiture (DA-NET-001)**:
    - A disconnected player **never** forfeits while their opponent has priority and they face no pending interactive prompts. The server does not run an offline forfeiture countdown during opponent turns.
    - If a player is offline when their own turn begins, their 75-second Main Phase Turn Clock runs normally. Upon expiration, their remaining Reserve Time Bank (`timeBankMs`) drains millisecond-for-millisecond. Match forfeiture (`FORFEIT_TIMEOUT`) occurs strictly if their own turn clock AND full reserve time bank are completely exhausted to 0 without reconnection.
    - If an offline player is presented with a `PlayerChoicePrompt` during an opponent's turn, the prompt action clock (10 seconds for the non-active player, plus the 3,000 ms animation buffer; `FR-MOD-001`, `FR-PAC-001`) runs first, followed by Reserve Time Bank depletion. If `timeBankMs` hits 0 before a response is received, the server commits the deterministic fallback immediately (`FR-MOD-001`), and the match continues without declaring forfeiture.
  - **Reserve Bank Depletion & Deterministic Fallbacks**:
    - When any active turn clock or choice prompt expires (whether the player is connected or offline):
      - If the player's `timeBankMs > 0`, the match timer transitions into Reserve Time, drawing down `timeBankMs` millisecond-for-millisecond while displaying an active rope/fuse burning effect to both clients.
      - If `timeBankMs == 0` (or burns to 0):
        - For an active `PlayerChoicePrompt`: The server commits the deterministic fallback immediately (`FR-MOD-001`) and execution advances without stalling.
        - For the Main Phase Turn Clock: The server commits an automated `AutoEndTurnCommand`, immediately handing turn priority to the opponent.
  - **State Rehydration & Snapshot Rate-Limiting (DA-NET-002)**:
    - If the disconnected player reconnects:
      1. The client sends `ReconnectRequest` with the active `matchAuthToken`.
      2. The server responds with a complete `FullStateSnapshot` containing:
         - Match stage enum (`STAGE_FIRST_PLAYER_CHOICE`, `STAGE_MULLIGAN`, `STAGE_MAIN_MATCH`).
         - Setup stage context if reconnecting during pre-game (initial 6-card hand instances, pending `MulliganPrompt` or `FirstPlayerChoicePrompt` with active deadlines).
         - Public board state (all assets across HQ, Reserves, Main Theater).
         - Controller's private hand instances.
         - Discard pile public contents.
         - Current phase, active player, resource pools, remaining `timeBankMs`, and Main Phase Turn Clock deadline (`FR-TRN-001`).
         - Active stack items and pending prompts (with remaining deadline).
         - Last processed `sequenceId`.
      3. The client renders the state instantaneously, clears its animation queue, and synchronizes its local timer displays with the server deadlines.
    - **Snapshot Throttling**: The server authoritatively enforces a rate limit of at most 1 `FullStateSnapshot` per 3,000 ms per client. If a client triggers 3 consecutive desync mismatches within 15 seconds, automated requests are suppressed and the client is placed in a throttled recovery state with a user-facing re-synchronization prompt, mitigating snapshot request storm denial-of-service risks (`DA-NET-002`, `FR-PAC-004`).

> [!NOTE]
> **Design Rationale (DA-NET-001 & DA-NET-002)**:
> - **DA-NET-001**: Eliminates erroneous forfeitures while an opponent is thinking on their turn. Disconnection timers only drain when the offline player has active priority or an unanswered prompt.
> - **DA-NET-002**: Throttles state snapshots to 1 per 3,000 ms with backoff after 3 consecutive failures, eliminating potential denial-of-service flooding from desynced clients.

---

## DOMAIN 3: INFORMATION HIDING, FOG-OF-WAR & SECURITY BOUNDARY

```text
[ Authoritative Server Game State ]
├── Player 1 Deck: [CardInstance 101, 102, 103...]
├── Player 1 Hand: [CardInstance 104, 105, 106...]
└── Reserves: Scratch Company (Underlying Card: CardInstance 107 -> "Veil Render")
        │
        ├── Filters through Fog-of-War Serializer ────────┐
        ▼                                                 ▼
[ Player 1 Client Packet ]                       [ Player 2 Client Packet ]
- Hand: Full card details (104, 105, 106)       - Hand: Count = 3 (No card IDs/Faces)
- Scratch Unit: proxy_scratch_99                - Scratch Unit: proxy_scratch_99
  Stats: 1/1/1/1, Name: "Scratch Company"         Stats: 1/1/1/1, Name: "Scratch Company"
  Real Face: [SCRUBBED - OPAQUE TOKEN]            Real Face: [SCRUBBED - OPAQUE TOKEN]
```

### Section 3.1: Double-Blind Scratch Company Architecture

#### FR-SEC-001: Absolute Double-Blind Invariant
- In strict fulfillment of **FR-ZON-005**:
  - When a Scratch Company is deployed via `Deploy Scratch Company` (`FR-INH-004`) or `Scratch N` (`FR-KWD-002m`), the server **never transmits** the underlying `blueprintId` or real `cardInstanceId` to **either** client.
  - The server generates an ephemeral **Opaque Proxy ID** (e.g., `proxy_scratch_8841`) that represents the asset in all client-bound messages.
  - The authoritative server retains the internal binding: `proxy_scratch_8841 -> underlyingCardInstanceId`.
  - **Memory & Packet Sniffing Immunity**: Because the real card data exists strictly in server RAM and is scrubbed before packet serialization, neither player can discern the underlying card face via network interception or memory inspection.

#### FR-SEC-002: Zone Departure Reveal Protocols
In accordance with **FR-ZON-005**, when a Scratch Company leaves a battlefield zone, the server serializes distinct network packets based on the destination zone type:

1. **Departure to Public Zone (`DISCARD`)**:
   - The entity transitions face-up.
   - The server broadcasts an unmasked `AssetDestroyedEvent` or `ZoneTransitionEvent` to **both** players:
     ```protobuf
     message AssetToDiscardEvent {
       string proxy_asset_id = 1;
       string destination_zone = 2; // "DISCARD"
       string revealed_blueprint_id = 3; // e.g. "animus_veil_render"
       string owner_player_id = 4;
     }
     ```
   - Both clients play the flip/reveal animation and place the revealed card into the owner's public Discard pile.

2. **Departure to Private Zone (`HAND`, `DECK`)**:
   - When returned/bounced to a private zone (e.g., via *Ambush Tactics*):
     - **To the Owner's Client**: The server transmits an unmasked transition revealing the card identity and provisioning a new client-visible `cardInstanceId`:
       ```protobuf
       message ScratchReturnToOwnerHandEvent {
         string proxy_asset_id = 1;
         string new_card_instance_id = 2;
         string revealed_blueprint_id = 3; // e.g. "animus_veil_render"
       }
       ```
     - **To the Opponent's Client**: The server transmits an opaque transition with zero card identity:
       ```protobuf
       message ScratchReturnToOpponentHandEvent {
         string proxy_asset_id = 1;
         string destination_zone = 2; // "HAND"
         int32 opponent_hand_count = 3; // Incremented total
       }
       ```

3. **Transition to Attached Zone (`ATTACHED` / `NO-AH-5`)**:
   - When tucked under a host entity, the asset has not departed the battlefield context.
   - The server broadcasts an opaque attachment event to **both** clients:
     ```protobuf
     message AssetTuckedEvent {
       string host_asset_id = 1;
       string proxy_asset_id = 2;
       bool is_face_down = 3; // true
       // revealed_blueprint_id is STRICTLY OMITTED
     }
     ```
   - Both clients render a face-down generic card tucked beneath the host. The underlying card identity remains completely concealed from both players.

---

### Section 3.2: General Fog-of-War & Information Sanitization

#### FR-SEC-003: Player Hand Sanitization
- Clients receive full card details (blueprint, instance modifiers, cost reductions) strictly for cards in their own hand.
- For the opponent's hand, the server transmits only the array size and positional slot indices (`slot_0`, `slot_1`, ...). Card identities, types, and costs are strictly excluded.

#### FR-SEC-004: Deck Search & Inspection Sanitization
- In compliance with **FR-DEK-001**:
  - When `SEARCH_DECK` or `LOOK_TOP_CARDS` executes, the server sends the card blueprint array **exclusively** to the searching player's client via `PlayerChoicePrompt` (`prompt_type = "SEARCH_DECK"` or `"LOOK_TOP_CARDS"`, populating `ChoiceOption.blueprint_id` per `FR-MSG-005`).
  - The opponent's client receives an opaque notification indicating that a deck search is underway, without card list data.
  - Cards currently drafted as Scratch Companies on the battlefield do not reside in the deck; deduction by inspection of remaining cards is fully permitted.

#### FR-SEC-005: Authoritative Server-Side RNG
- In compliance with **FR-DEK-002**:
  - All random permutations (deck shuffling, digital coin tosses, random target selections during fallbacks) are executed exclusively on the server using cryptographically secure pseudorandom number generators (CSPRNG).
  - The client is prohibited from submitting random seeds or RNG outcomes.

---

## DOMAIN 4: MESSAGE TAXONOMY & WIRE PROTOCOL

### Section 4.1: Message Framing & Envelope

#### FR-MSG-001: Universal Message Envelope
Every message exchanged over the WebSocket connection is encapsulated in a standard binary protobuf envelope:

```protobuf
syntax = "proto3";
package ccg.network;

message GameEnvelope {
  uint64 sequence_id = 1;       // Monotonically increasing event/packet sequence
  uint64 timestamp_epoch_ms = 2;// Unix timestamp in milliseconds
  string match_id = 3;          // Active match GUID
  
  oneof payload {
    // Client-to-Server
    ClientCommand command = 10;
    ClientHandshake client_handshake = 11;
    
    // Server-to-Client
    ServerEventStream event_stream = 20;
    PlayerChoicePrompt choice_prompt = 21;
    StateSyncSnapshot state_snapshot = 22;
    CommandRejection rejection = 23;
    HeartbeatMessage heartbeat = 24;
    HandshakeAck handshake_ack = 25;
    HandshakeReject handshake_reject = 26;
  }
}

message ClientHandshake {
  string client_engine_version = 1; // e.g. "3.0.0" (FR-SES-001)
  uint32 protocol_version = 2;      // e.g. 1 (FR-SES-001)
  string catalog_checksum = 3;      // 64-char hex SHA-256 of canonical blueprints.json (FR-DAT-004, FR-CMP-025)
  string match_auth_token = 4;      // Matchmaking JWT (FR-SES-001)
}

message HandshakeAck {
  string server_engine_version = 1;
  uint32 protocol_version = 2;
  string catalog_checksum = 3;
  uint64 server_epoch_ms = 4;
}

message HandshakeReject {
  uint32 error_code = 1;     // ERR_VERSION_MISMATCH or ERR_CATALOG_CHECKSUM_MISMATCH
  string reason_message = 2;
  string expected_engine_version = 3;
  uint32 expected_protocol_version = 4;
  string expected_catalog_checksum = 5;
}
```

---

### Section 4.2: Client Command Taxonomy (Client -> Server)

#### FR-MSG-002: Client Commands
*(Subsumed by FR-MSG-003)*
The top-level `ClientCommand` message wrapper encapsulates client intent and is unified directly with the concrete action schemas specified in **FR-MSG-003**:

```protobuf
message ClientCommand {
  string command_id = 1;
  uint64 ack_sequence_id = 2; // Last sequence_id observed by client
  
  oneof action {
    // Match Orchestration & Session Recovery (FR-SES-001 - FR-SES-005, FR-PAC-004)
    MatchJoinCommand match_join = 3;
    FirstPlayerChoiceCommand first_player_choice = 4;
    MulliganCommitCommand mulligan_commit = 5;
    ReconnectRequest reconnect_request = 6;
    RequestStateSnapshot request_state_snapshot = 7;

    // Off-Stack Economic Operations (FR-CAP-001, FR-INH-001 - Require Empty Stack, Resolve Immediately Off-Stack)
    GenerateResourceCommand generate_resource = 12;
    ConvertResourceCommand convert_resource = 14;

    // Stack-Pushed Tactical Inherent Actions (FR-CAP-001, FR-INH-001 - Require Empty Stack, Push Frame onto Resolution Stack)
    MarchCommand march = 10;
    AttackCommand attack = 11;
    DeployScratchCommand deploy_scratch = 13;
    
    // Stack Card & Ability Actions (FR-QUE-001, FR-TGT-101)
    DeployCardCommand deploy_card = 20;
    EnactManeuverCommand enact_maneuver = 21;
    ActivateAbilityCommand activate_ability = 22;
    
    // Turn & Phase Flow (FR-TRN-001)
    EndPhaseCommand end_phase = 30;
    
    // Choice Prompt Responses (FR-TRG-002, FR-MOD-001)
    ChoiceResponse choice_response = 40;
    
    // Match Administration
    ConcedeCommand concede = 50;
  }
}
```

#### FR-MSG-003: Inherent & Bound Action Payload Serialization
In fulfillment of **FR-CAP-001**, **FR-INH-001** through **FR-INH-004**, and **FR-TGT-101**:

```protobuf
// ============================================================================
// Match Orchestration & Session Recovery Commands (FR-SES-001 - FR-SES-005, FR-PAC-004)
// ============================================================================

message MatchJoinCommand {
  string match_token = 1;
  string deck_registration_token = 2; // Signed token bound to backend deck registration
}

enum FirstPlayerChoice {
  FIRST_PLAYER_UNSPECIFIED = 0;
  FIRST_PLAYER_PLAY_FIRST = 1;
  FIRST_PLAYER_DRAW_FIRST = 2;
}

message FirstPlayerChoiceCommand {
  FirstPlayerChoice choice = 1;
}

message MulliganCommitCommand {
  repeated string set_aside_card_instance_ids = 1;
}

message ReconnectRequest {
  string match_auth_token = 1;          // Active matchmaking JWT (FR-SES-005)
  uint64 last_received_sequence_id = 2; // Last sequence_id processed prior to disconnect
}

message RequestStateSnapshot {
  uint64 desync_detected_at_seq_id = 1; // Sequence ID where state_hash mismatch occurred (FR-PAC-004)
  uint64 client_computed_state_hash = 2;// Client's local Layer-6 FNV-1a public board state hash
}

// ============================================================================
// Off-Stack Economic Operations (FR-CAP-001, FR-INH-001 - Require Empty Stack, Resolve Immediately Off-Stack)
// ============================================================================

enum ResourceType {
  RESOURCE_TYPE_UNSPECIFIED = 0;
  RESOURCE_TYPE_REQUISITION = 1;
  RESOURCE_TYPE_TACTICS = 2;
}

message GenerateResourceCommand {
  string asset_instance_id = 1;
  // Explicit resource type selection:
  // - Required for Versatile units, Structures, and HQs (which can yield either).
  // - Validated against zone default for standard units (Reserves -> REQUISITION, Main Theater -> TACTICS).
  ResourceType chosen_resource = 2;
}

message ConvertResourceCommand {
  // Number of Tactics desired; strictly validated to range 1 <= tactics_requested <= 50.
  // Non-positive integers (<= 0) or values > 50 reject with ERR_INVALID_COMMAND_ARGUMENTS (DA-NET-003).
  // Server deterministically deducts (tactics_requested * 2) Requisition (FR-RES-003).
  int32 tactics_requested = 1; 
}

// ============================================================================
// Stack-Pushed Tactical Inherent Actions (FR-CAP-001, FR-INH-001 - Require Empty Stack, Push onto Resolution Stack)
// ============================================================================

message MarchCommand {
  string asset_instance_id = 1; // Exhausts asset and pushes MARCH frame onto Resolution Stack
}

message AttackCommand {
  string attacker_asset_id = 1;
  // Standard attack targets strictly 1 asset or HQ; Berserk units (FR-KWD-002o) may target up to 2 assets in designated sequential retaliation order (FR-CBT-002).
  // Clamped to max 1 (or max 2 for Berserk); excess targets rejected with ERR_INVALID_COMMAND_ARGUMENTS (DA-NET-003).
  repeated string target_asset_ids = 2; 
}

message DeployScratchCommand {
  string hq_asset_id = 1; // Exhausts HQ and pushes DEPLOY_SCRATCH frame onto Resolution Stack (FR-INH-004)
}

// ============================================================================
// Stack Card & Ability Action Payloads (FR-TGT-101, FR-QUE-001)
// ============================================================================

enum TargetRole {
  TARGET_ROLE_UNSPECIFIED = 0;
  TARGET_ROLE_SOURCE = 1;
  TARGET_ROLE_TARGET = 2;
  TARGET_ROLE_SACRIFICE_COST = 3;           // For additional deployment/activation sacrifice costs
  TARGET_ROLE_DISCARD_COST = 4;             // For activated ability hand discard costs
  TARGET_ROLE_EXHAUST_COST = 5;             // For additional enactment exhaust costs (Funeral Fungus)
  TARGET_ROLE_DESTINATION_ZONE = 6;         // For target destination zone selections (Sandstorm Lightbringer)
  TARGET_ROLE_SECONDARY_BERSERK_TARGET = 7; // Upfront secondary defender binding for Berserk attacks (FR-KWD-002o, FR-TGT-101)
}

message DeployCardCommand {
  string card_instance_id = 1;
  string target_zone = 2; // "RESERVES" or "MAIN_THEATER"
  repeated TargetBinding target_bindings = 3;
  repeated int32 selected_modal_indices = 4;
  string source_zone = 5; // "HAND" (default) or "DISCARD" (FR-TRN-003a)
}

message EnactManeuverCommand {
  string card_instance_id = 1;
  repeated TargetBinding target_bindings = 2;
  repeated int32 selected_modal_indices = 3;
}

message ActivateAbilityCommand {
  string asset_instance_id = 1;
  int32 ability_index = 2;
  repeated TargetBinding target_bindings = 3;
  repeated int32 selected_modal_indices = 4;
}

message TargetBinding {
  string target_id = 1;      // AssetInstance ID, CardInstance ID (for discard costs), Player ID, or Zone
  int32 allocated_value = 2; // For DISTRIBUTED_POOL (e.g. Macrocannon Barrage)
  int32 modal_index = 3;     // Associated mode index for modal spells (e.g. Ambush Tactics)
  TargetRole target_role = 4;// Role qualifier for multi-target actions (SOURCE, TARGET, SACRIFICE_COST, DISCARD_COST, SECONDARY_BERSERK_TARGET)
}
```

---

### Section 4.3: Server Event Stream Taxonomy (Server -> Client)

#### FR-MSG-004: Atomic Event Stream Architecture
The server transmits state mutations as a continuous stream of atomic events packaged within `ServerEventStream`:

```protobuf
message ServerEventStream {
  uint64 batch_start_seq = 1;
  uint64 batch_end_seq = 2;
  repeated EngineEvent events = 3;
  PublicStateHash state_hash = 4; // Checksum for desync verification
}

message EngineEvent {
  uint64 seq_id = 1;
  string event_type = 2; // e.g. "EVENT_MARCH", "EVENT_DAMAGE_APPLIED"
  
  oneof detail {
    AssetMovedEvent asset_moved = 10;
    CombatResolvedEvent combat_resolved = 11;
    DamagePipelineStageEvent damage_stage = 12;
    ResourcePoolChangedEvent resource_changed = 13;
    StackPushedEvent stack_pushed = 14;
    StackPoppedEvent stack_popped = 15;
    AssetDestroyedEvent asset_destroyed = 16;
    PhaseTransitionEvent phase_transition = 17;
    MatchConcludedEvent match_concluded = 18;
    CardDrawnEvent card_drawn = 20;
    CardDiscardedEvent card_discarded = 21;
    AssetDeployedEvent asset_deployed = 22;
    AssetMorphedEvent asset_morphed = 23;
    ControlChangedEvent control_changed = 24;
    TuckedCardsPurgedEvent tucked_cards_purged = 25;
  }
}

// ============================================================================
// Core Engine Event Detail Schemas
// ============================================================================

message AssetMovedEvent {
  string asset_id = 1;
  string source_zone = 2;
  string destination_zone = 3;
}

message CombatResolvedEvent {
  string attacker_asset_id = 1;
  repeated string defender_asset_ids = 2;
  int32 attacker_damage_dealt = 3;
  repeated int32 defender_retaliation_damage = 4;
  bool is_simultaneous_hq_defeat = 5;
}

message ResourcePoolChangedEvent {
  string player_id = 1;
  int32 requisition_delta = 2;
  int32 current_requisition = 3;
  int32 tactics_delta = 4;
  int32 current_tactics = 5;
  string source_reason = 6; // e.g. "INHERENT_GENERATE", "CONVERT", "MANEUVER"
}

message StackPushedEvent {
  string stack_item_id = 1;
  string source_id = 2;
  string ability_type = 3; // "MANEUVER_EFFECT", "ACTIVATED", "TRIGGERED", "DEPLOYMENT", "ATTACK", "MARCH", "DEPLOY_SCRATCH"
  string controller_player_id = 4;
  int32 current_stack_depth = 5;
}

message StackPoppedEvent {
  string stack_item_id = 1;
  bool did_resolve = 2;
  bool did_fizzle = 3;
}

message AssetDestroyedEvent {
  string proxy_asset_id = 1;
  string owner_player_id = 2;
  string destination_zone = 3; // "DISCARD"
  string revealed_blueprint_id = 4; // Unmasked card identity (FR-SEC-002)
}

message PhaseTransitionEvent {
  string new_phase = 1; // "READY_PHASE", "DRAW_PHASE", "MAIN_PHASE", "END_PHASE"
  string active_player_id = 2;
  int32 turn_number = 3;
  uint64 phase_deadline_epoch_ms = 4; // Deadline for 75s Main Phase Turn Clock; auto-ends turn on timeout (FR-TRN-001)
}

message MatchConcludedEvent {
  string winner_player_id = 1;
  bool is_match_draw = 2; // True if simultaneous HQ defeat (FR-CBT-003)
  string conclusion_reason = 3; // "HQ_DESTROYED", "EMPTY_DECK_DRAW", "CONCEDE"
}

message CardDrawnEvent {
  string player_id = 1;
  string card_instance_id = 2; // Masked for opponent
  string revealed_blueprint_id = 3; // Sent strictly to owner; scrubbed for opponent (FR-SEC-003)
  int32 remaining_deck_count = 4;
}

message CardDiscardedEvent {
  string player_id = 1;
  string card_instance_id = 2;
  string blueprint_id = 3; // Public in discard
  string source_zone = 4; // "HAND", "DECK", "ATTACHED"
}

message AssetDeployedEvent {
  string asset_instance_id = 1;
  string card_blueprint_id = 2; // Generic proxy if is_face_down
  string controller_player_id = 3;
  string destination_zone = 4; // "RESERVES", "MAIN_THEATER"
  bool is_face_down = 5;
}

message AssetMorphedEvent {
  string asset_instance_id = 1;
  bool is_face_down = 2;
  string effective_blueprint_id = 3; // "neutral_scratch_company"
}

message ControlChangedEvent {
  string asset_instance_id = 1;
  string previous_controller_id = 2;
  string new_controller_id = 3;
  string current_zone = 4; // Zone remains unchanged (FR-ZON-002, FR-ENT-003)
}

message TuckedCardsPurgedEvent {
  string host_asset_id = 1;
  repeated string purged_card_instance_ids = 2;
  repeated string revealed_blueprint_ids = 3;
  string destination_zone = 4; // "DISCARD"
}
```

---

### Section 4.4: Synchronous Choice Prompt Protocol

#### FR-MSG-005: Choice Prompts & Server Authority
In compliance with **FR-TRG-002**, **FR-REP-001**, and **FR-MOD-001**:
- When stack execution requires an interactive decision, the server issues a `PlayerChoicePrompt`:
  ```protobuf
  message PlayerChoicePrompt {
    string prompt_id = 1;
    // "LOOK_TOP_CARDS", "SEARCH_DECK", "REVEAL_HAND", "MODAL_SELECT",
    // "TARGET_SELECT", "TRIGGER_ORDER_SELECT", "REPLACEMENT_CONFLICT_SELECT", "OPTIONAL_TRIGGER_CONFIRM"
    string prompt_type = 2;
    uint64 deadline_epoch_ms = 3; // Exact UTC millisecond deadline (FR-MOD-001, FR-PAC-001)
    int32 min_selections = 4;
    int32 max_selections = 5;
    repeated ChoiceOption options = 6;
  }

  message ChoiceOption {
    int32 option_index = 1;
    string entity_or_card_id = 2;
    string description = 3;
    string blueprint_id = 4; // Populated when revealing/inspecting cards to choosing player (SEARCH_DECK, LOOK_TOP_CARDS, REVEAL_HAND; FR-SEC-004)
  }

  message PartitionedSelection {
    string category_or_zone = 1; // e.g. "HAND", "DISCARD", "TOP_OF_DECK", "BOTTOM_OF_DECK"
    repeated int32 option_indices = 2;
  }

  message ChoiceResponse {
    string prompt_id = 1;
    repeated int32 selected_option_indices = 2;
    repeated int32 bottom_deck_ordered_indices = 3; // Ordered card indices returned to bottom of deck
    repeated PartitionedSelection partitioned_selections = 4; // Multi-destination partitioned choices
    bool pass_or_decline = 5; // True to decline optional prompt
    repeated int32 ordered_trigger_indices = 6; // Stack placement order for simultaneous allied triggers (FR-MSG-005, FR-TRG-002)
  }
  ```
- **Deadline Enforcement & Deterministic Fallbacks**: The client UI displays an active countdown based on `deadline_epoch_ms - currentServerTimeMs()`.
- If the client's `ChoiceResponse` is received after `deadline_epoch_ms`, the server rejects it and commits the deterministic fallback (`FR-MOD-001`). For `TRIGGER_ORDER_SELECT`, the server defaults to the natural order of trigger registration.

---

## DOMAIN 5: STATE SYNCHRONIZATION, PACING & ANIMATION FLOW

```text
[ Server Engine Thread ]                   [ Godot Client ]
         │                                         │
         │ 1. EventBatch (Seq 42-46)               │
         ├────────────────────────────────────────►│ [ Receives Batch ]
         │    - March (Seq 42: Unit Exhausts)      │ ├── Enqueues to FIFO Queue
         │    - Regroup (Seq 43: Unit Readies)     │ ├── Pop 42: Animate March (Exhausts)
         │    - Attack (Seq 44)                    │ ├── Pop 43: Animate Regroup (Readies)
         │    - Damage Stage 1-5 (Seq 45)          │ ├── Pop 44: Animate Attack
         │    - Asset Destroyed (Seq 46)           │ ├── Pop 45: Animate 5-Stage Mit.
         │                                         │ └── Pop 46: Animate Death/Reveal
         │ 2. Ready for Next Command               │
         │    (Server state committed immediately) │
```

### Section 5.1: Decoupled Animation Queue

#### FR-PAC-001: Client-Side FIFO Event Queue & Server Animation Grace Buffer
- The Godot client maintains a decoupled, local `AnimationEventQueue`.
- When a `ServerEventStream` packet arrives, the client immediately stores the latest state in an unrendered mirror state cache, while passing the atomic event structs into the queue.
- Visual components consume the queue sequentially:
  - `EVENT_MARCH` triggers sprite translation across the board and transitions the unit to Exhausted (`FR-INH-001`, `FR-CAP-001`). Attacks require a readied asset, so marching units cannot attack without an intervening readying effect (e.g. `Regroup`).
  - `EVENT_DAMAGE_APPLIED` spawns floating combat text and flashes defensive shields.
  - `EVENT_DESTROY` plays death dissolve and triggers the reveal flip if a Scratch Company proxy is destroyed.
- **Server Turn Handover Animation Grace Buffer**:
  - Upon broadcasting a `PhaseTransitionEvent` entering a player's Main Phase (`FR-TRN-001`), the server grants an authoritative **5,000 ms Animation Grace Window** before the 75-second Main Phase Turn Clock countdown begins (`turnClockDeadlineEpochMs = currentServerTimeMs() + 5000 + 75000`).
  - If the active player submits any valid command during this 5,000 ms window, the grace period ends immediately and standard turn execution begins without delay.
  - This bounded server-side buffer ensures incoming players are never deprived of playable turn time by complex prior-turn animation cascades ("turn theft"), while preventing malicious clients from withholding completion ACKs.
- **Interactive Choice Prompt Animation Buffer (DA-NET-004)**:
  - In compliance with **FR-MOD-001**, when issuing any `PlayerChoicePrompt` (`FR-MSG-005`), the server grants an automatic **3,000 ms Animation Buffer** before the player's action clock begins ticking down:
    - **Active Player Prompt Clock (15,000 ms)**: `deadlineEpochMs = currentServerTimeMs() + 3000 + 15000`.
    - **Non-Active Player Prompt Clock (10,000 ms)**: `deadlineEpochMs = currentServerTimeMs() + 3000 + 10000`.
  - If the client submits a valid `ChoiceResponse` during this buffer, it is processed immediately.
  - This bounded buffer prevents cascading animations from prior actions (such as combat resolutions or complex maneuver board wipes) from truncating either player's decision window.

> [!NOTE]
> **Design Rationale (DA-NET-004)**: The 3,000 ms prompt animation buffer guarantees that players always receive their full intended decision time (15 seconds for the active player, 10 seconds for the non-active player per `FR-MOD-001`), even after extensive visual death and combat animations.

#### FR-PAC-002: Animation Fast-Forwarding on Inbound Prompts & Phase Transitions
- If an interactive `PlayerChoicePrompt` arrives, or if a `PhaseTransitionEvent` entering the player's Main Phase is received while the client's `AnimationEventQueue` has pending background visual items:
  - If the animation queue backlog exceeds 2,000 ms of playback time, the Godot client automatically accelerates animation playback rates (up to 3.0x speed) or collapses non-essential visual transitions (e.g., snapping floating combat numbers and death dissolves to finalized states) to guarantee that the player is presented with the prompt or turn UI with maximum remaining time on their action clock.

---

### Section 5.2: Damage Mitigation Pipeline Pacing

#### FR-PAC-003: 5-Stage Mitigation Event Batching
- In compliance with **FR-REP-004**:
  - The 5-stage automated damage mitigation pipeline executes instantaneously as a single atomic transaction on the server.
  - To enable rich visual presentation without roundtrip stalls, the server outputs an array of discrete sub-stage results in a single event packet:
    ```protobuf
    enum SourceEntityType {
      SOURCE_ENTITY_TYPE_UNSPECIFIED = 0;
      SOURCE_ENTITY_TYPE_ASSET = 1; // Unit, Structure, or HQ on battlefield (AssetInstance ID)
      SOURCE_ENTITY_TYPE_CARD = 2;  // Maneuver or CardInstance from Hand/Discard/Stack
    }

    message DamagePipelineStageEvent {
      SourceEntityType source_entity_type = 1;
      string source_entity_id = 2; // AssetInstance ID or CardInstance ID
      string initial_target_asset_id = 3;
      int32 raw_damage_amount = 4;
      
      // Stage 1: Redirection
      string final_target_asset_id = 5;
      // Stage 2: Resistant Flat Reduction
      int32 resistant_absorbed = 6;
      // Stage 3: Stalwart Capping
      int32 stalwart_capped_amount = 7;
      // Stage 4: Dynamic Armor Absorption
      int32 armor_absorbed = 8;
      int32 remaining_armor = 9;
      // Stage 5: Health Damage Committed
      int32 health_damage_committed = 10;
      int32 final_current_health = 11;
    }
    ```
  - The Godot client sequences this single packet into a fluid visual cascade (Shield Deflect &rarr; Stalwart Flash &rarr; Armor Cracking &rarr; Health Bar Drop) without requiring server roundtrips.

---

### Section 5.3: Desync Verification via Public State Hashing

#### FR-PAC-004: Public State Checksum Calculation & Per-Batch Verification
- At the conclusion of every event stream batch transmitted to clients—including off-stack economic operations (Resource Generation, Conversion), stack-pushed tactical inherent actions (March, Attack, Scratch Deploy), phase transitions, and stack unwinding sequences (`FR-CAP-001`, `FR-INH-001`)—the server computes a 64-bit CRC/FNV-1a hash of the **Public Board State**:
  - Combined Health, dynamic Armor pools, and Combat Power of all assets in HQ, Reserves, and Main Theater.
  - Readiness states (`READIED` / `EXHAUSTED`) of all assets.
  - Total card counts in Hand, Deck, and Attached zones for each player.
  - Complete list of card blueprint IDs in both public Discard piles.
  - Current Requisition and Tactics pools.
  - *Secret card identities in hands, decks, and face-down Scratch proxies are strictly excluded from the hash.*
- **Effective Layer 6 Attribute Rule**: To eliminate desync vulnerabilities with masked entities, the state hasher calculates asset Health and Combat Power strictly using post-mask effective values from Layer 6 of the Attribute Pipeline (`FR-LAY-001`, `FR-ENT-101`), guaranteeing 100% hash parity with client-visible proxy states.
- **Per-Batch Emission**: The hash is transmitted via `ServerEventStream.state_hash` (`FR-MSG-004`).
- **Immediate Single-Frame Recovery & Rate-Limiting**: The Godot client computes the identical hash across its mirrored public state cache upon draining each event stream batch:
  - If a hash mismatch is detected after any batch, the client flags a desync immediately on that frame and transmits `RequestStateSnapshot` (`FR-MSG-003`) to receive a `FullStateSnapshot` (`FR-SES-005`), seamlessly recovering the board cache before any subsequent player commands can fail validation.
  - **Snapshot Flood Protection**: In accordance with `FR-SES-005` (DA-NET-002), snapshot requests are authoritatively throttled to at most 1 per 3,000 ms, with automated backoff after 3 consecutive failures to eliminate denial-of-service packet storms.

---

## DOMAIN 6: MATCH REPLAY ARCHITECTURE

### Section 6.1: Match Replay File Format (`.rec`)

#### FR-SPC-002: Master Replay Data Architecture
- Upon match conclusion, the server compiles and archives a complete match recording (`.rec` file).
- **Zero Information Hiding in Replays**:
  - The replay file contains the **100% unmasked historical truth**.
  - Includes initial shuffled deck order arrays, exact card IDs of all drawn cards, and the real underlying card identities of all Scratch Companies.
- **File Structure**:
  - `ReplayHeader`: Match GUID, Engine Version, Date, Player Metadata, Deck Lists.
  - `InitialState`: Shuffled deck states, HQ designations.
  - `EventLog`: Monotonic array of all `GameEnvelope` packets exchanged, enriched with unmasked underlying entity IDs.
- **Client Playback**: The Godot client's Replay Viewer reads the `.rec` file, offering analytical toggles:
  - *Fog-of-War Mode*: Simulates what Player 1 or Player 2 saw in real time.
  - *Omniscient Mode*: Unmasks all face-down Scratch Companies, tucked sub-cards, and hand contents for tournament broadcasting, coaching, and balance analysis.
- **Replay Distribution & Anti-Scouting Quarantine (DA-NET-006)**:
  - Unmasked `.rec` replay files are strictly quarantined on the authoritative game server until the entire match-set (e.g. Best-of-3 or Best-of-5 tournament series) concludes.
  - Clients and external API consumers cannot access unmasked replay telemetry between individual games of an active series, eliminating competitive deck scouting and information leaks prior to series completion.

> [!NOTE]
> **Design Rationale (DA-NET-006)**: Quarantining unmasked match replays until tournament series conclusion ensures competitive integrity while preserving rich omniscient replay analytics post-match.

---

## DOMAIN 7: ERROR HANDLING & COMMAND REJECTION

### Section 7.1: Rejection Protocol

#### FR-ERR-001: Rejection Packet & Clean Stack Invariants
- If a client command fails server validation (e.g., attempting an inherent action while `Stack.isEmpty() == false`, declaring an attack with an exhausted asset, or targeting an Ethereal unit), the server transmits a `CommandRejection`:
  ```protobuf
  message CommandRejection {
    string command_id = 1;
    uint32 error_code = 2;
    string reason_message = 3;
    uint64 current_valid_sequence_id = 4;
  }
  ```
- **Error Codes**:
  - `ERR_STACK_NOT_EMPTY`: Manual action declared while resolution stack is active (`FR-QUE-001`).
  - `ERR_INSUFFICIENT_RESOURCES`: Requisition/Tactics pool unable to pay cost (`FR-RES-001`).
  - `ERR_INVALID_TARGET`: Target failed conjunction filter or is protected by Ward/Ethereal (`FR-TGT-001`, `FR-KWD-002g`, `FR-KWD-002h`).
  - `ERR_GUARDIAN_PRIORITY_VIOLATION`: Attack declaration illegally bypassed an eligible enemy Guardian asset in Main Theater (`FR-CBT-001`, `FR-KWD-002j`, `DA-NET-005`). Allows the Godot client UI to flash the mandatory Guardian target and guide the player.
  - `ERR_INVALID_COMMAND_ARGUMENTS`: Command parameters out of bounds or malformed (e.g. non-positive tactics conversion value, invalid target count; `FR-MSG-003`, `DA-NET-003`).
  - `ERR_ASSET_EXHAUSTED`: Attempted action requiring Readied state (`FR-INH-001`).
  - `ERR_INVALID_RESOURCE_CHOICE`: Attempted resource generation invalid for current zone/keywords (e.g. non-Versatile unit generating Tactics in Reserves, or selecting an unsupported resource type; `FR-INH-002`, `FR-KWD-002i`).
  - `ERR_ACTION_CLOCK_EXPIRED`: Choice response received past deadline (`FR-TRG-002`).
- Upon receiving a rejection, the Godot client rolls back any local optimistic drag-and-drop animation and plays an invalid-action audio cue.

> [!NOTE]
> **Design Rationale (DA-NET-003 & DA-NET-005)**:
> - **DA-NET-003**: `ERR_INVALID_COMMAND_ARGUMENTS` guarantees wire-level protection against malformed or negative payload attacks.
> - **DA-NET-005**: Distinct `ERR_GUARDIAN_PRIORITY_VIOLATION` empowers client UI to give direct tactical feedback pointing to the blocking Guardian rather than a generic targeting failure.

---

## APPENDIX: COMPLETE CROSS-SPEC TRACEABILITY MATRIX

| Engine / Modeling Requirement | Requirement Scope | Client-Engine Protocol & Networking FRD v1.1 Fulfillment |
| :--- | :--- | :--- |
| **FR-SET-001** | Deck Legality & Rarity Limits | Section 2.1 (`FR-SES-002` Server-side deck registration & validation) |
| **FR-SET-002** | Match Init & Partial Mulligan | Section 2.1 (`FR-SES-004` Mulligan negotiation protocol) |
| **FR-SET-003** | First Player Determination | Section 2.1 (`FR-SES-003` First-player prompt protocol) |
| **FR-SET-004** | First-Turn Rules | Section 4.2 (`ClientCommand`), Section 4.3 (`PhaseTransitionEvent`) |
| **FR-DEK-001** | Deck Search Engine Primitive | Section 3.2 (`FR-SEC-004` Private deck inspection stream) |
| **FR-DEK-002** | Deck Shuffling Primitive | Section 3.2 (`FR-SEC-005` Authoritative server CSPRNG) |
| **FR-ZON-001** | Zone Definitions | Section 4.2 (`FR-MSG-003`), Section 5.3 (`FR-PAC-004`) |
| **FR-ZON-002** | Movement & Destruction | Section 4.3 (`AssetMovedEvent`, `AssetDestroyedEvent`) |
| **FR-ZON-004** | Non-Battlefield State Reset | Section 3.1 (`FR-SEC-002` Zone transition wire serialization) |
| **FR-ZON-005** | Scratch Fog of War Lifecycle | Section 3.1 (`FR-SEC-001` Double-blind proxy, `FR-SEC-002` Zone departure reveals) |
| **FR-RES-001** | Dual Resource Accumulation | Section 4.2 (`GenerateResourceCommand`, `ConvertResourceCommand`) |
| **FR-RES-003** | Requisition to Tactics Conversion| Section 4.2 (`ConvertResourceCommand` instantaneous non-stack execution) |
| **FR-TRN-001** | Phase Sequencing | Section 4.2 (`EndPhaseCommand`), Section 4.3 (`PhaseTransitionEvent`) |
| **FR-TRN-002** | Turn-Bound Structure | Section 4.2 (`ClientCommand` off-turn command suppression) |
| **FR-INH-001** | Inherent Action Clean Stack | Section 4.2 (`FR-MSG-002`), Section 7.1 (`ERR_STACK_NOT_EMPTY`) |
| **FR-INH-002** | Unit Inherent Actions | Section 4.2 (Zone-bound `GenerateResourceCommand`, `MarchCommand`, `AttackCommand`) |
| **FR-INH-003** | Structure Inherent Actions | Section 4.2 (`GenerateResourceCommand.chosen_resource` choice protocol) |
| **FR-INH-004** | HQ Inherent Actions | Section 4.2 (`DeployScratchCommand`), Section 3.1 (`FR-SEC-001`) |
| **FR-KWD-002i**| Versatile Keyword | Section 4.2 (`GenerateResourceCommand.chosen_resource` choice protocol) |
| **FR-QUE-001** | Centralized LIFO Resolution Stack| Section 4.3 (`StackPushedEvent`, `StackPoppedEvent`), Section 7.1 |
| **FR-TGT-101** | Bound Action Payload Model | Section 4.2 (`DeployCardCommand`, `TargetBinding` upfront serialization) |
| **FR-TRG-002** | Choice Timeouts & Action Clock | Section 4.4 (`PlayerChoicePrompt` with `deadline_epoch_ms` & fallbacks) |
| **FR-REP-004** | 5-Stage Damage Mitigation | Section 5.2 (`DamagePipelineStageEvent` atomic sub-stage streaming) |
| **FR-CBT-002** | Symmetric Retaliation | Section 4.3 (`CombatResolvedEvent` unified simultaneous strike packet) |
| **FR-CBT-003** | Simultaneous Defeat & Match Draw| Section 4.3 (`MatchConcludedEvent` with `RESULT_MATCH_DRAW`) |
| **FR-ENT-101** | Scratch Face-Down Wrapper | Section 3.1 (`FR-SEC-001` Opaque proxy ID masking) |
| **FR-ENT-102** | Tucked Sub-Entities (`NO-AH-5`)| Section 3.1 (`FR-SEC-002` Double-blind attached zone preservation) |
| **FR-KWD-002o**| Berserk Keyword | Section 4.2 (`AttackCommand.target_asset_ids` multi-target support) |
