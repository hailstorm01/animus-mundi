# Client-Engine Protocol & Networking - Functional Requirements Document (FRD v1.1)
**Master Network Topology, Message Taxonomy, Fog-of-War & Wire Protocol Specification**

---

## EXECUTIVE SUMMARY & REVISION HISTORY

This document defines the network architecture, communication transport, client-server message protocol, state synchronization model, and fog-of-war security boundary for the digital Collectible Card Game (CCG). It serves as the companion networking specification to the **Master Game Rules & Card Modeling FRD (v3.0 - `Game_Rules_and_Card_Modeling_FRD.md`)**.

### Revision Summary (v1.0 -> v1.1):
1. **Client Command Completion**: Added missing match orchestration commands (`MatchJoinCommand`, `FirstPlayerChoiceCommand`, `MulliganCommitCommand`) to `ClientCommand` (`FR-MSG-002`).
2. **Event Stream Expansion & Complete Schemas**: Expanded `EngineEvent` to include full Protobuf schemas for all 15 core engine events (`CardDrawnEvent`, `CombatResolvedEvent`, `AssetDeployedEvent`, `ControlChangedEvent`, etc.; `FR-MSG-004`).
3. **Heartbeat & Grace Window De-racing**: Tuned heartbeat cadence to 1,000 ms and disconnect threshold to 3,000 ms (`FR-NET-004`); configured prompt action clock (15s) with retrospective compensation based on last-received packet timestamp during Disconnect Grace Window (`FR-SES-005`).
4. **Target Roles & Cost Serialization**: Enriched `TargetBinding` with `TARGET_ROLE_SACRIFICE_COST` and `TARGET_ROLE_DISCARD_COST`; added `source_zone` to `DeployCardCommand` (`FR-MSG-003`).
5. **State Hash Mask Synchronization**: Explicitly mandated that public checksum hashing consumes effective Layer 6 values to eliminate Scratch false desyncs (`FR-PAC-004`).
6. **Deck Legality & Cross-Spec Alignment**: Formally enforced starting deck prohibition for `common_scratch_company` (`FR-SES-002`); aligned sequential `FR-SET-00X` and `FR-REP-004` references. Strict version matching enforced on handshake (`FR-SES-001`).
7. **Inherent Action Wire Schemas & Resource Choice**: Defined explicit Protobuf schemas for `GenerateResourceCommand`, `MarchCommand`, `AttackCommand` (supporting Berserk multi-target attacks), `DeployScratchCommand`, and `ConvertResourceCommand` (`FR-MSG-003`). Added `ChoiceResponse` schema with bottom-deck ordering (`FR-MSG-005`).

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
- **Round-Trip Time (RTT)**: The client computes rolling smoothed RTT to dynamically adjust local animation playback rates.
- **Connection Timeout**: If no frames (data or heartbeat) are received from a client for 3,000 ms, the server flags the connection as severed and initiates the Disconnect Grace Window (`FR-SES-005`).

---

## DOMAIN 2: SESSION LIFECYCLE & CONNECTION NEGOTIATION

```text
┌──────────────┐                               ┌──────────────┐
│ Godot Client │                               │  GCP Server  │
└──────┬───────┘                               └──────┬───────┘
       │                                              │
       │ 1. Connect (WSS Handshake)                   │
       ├─────────────────────────────────────────────►│
       │ 2. HandshakeAck (Protocol & Engine Version)  │
       │◄─────────────────────────────────────────────┤
       │ 3. MatchJoinRequest (MatchToken + DeckList)  │
       ├─────────────────────────────────────────────►│
       │ 4. MatchValidate & DeckLegality (FR-SET-001) │
       │    [Server validates HQ, neutral, 50+ cards] │
       │ 5. MatchStartNotification (Coin Toss/P1)     │
       │◄─────────────────────────────────────────────┤
       │ 6. MulliganPrompt (6 Initial Cards)          │
       │◄─────────────────────────────────────────────┤
       │ 7. MulliganCommit (Selected indices set aside)│
       ├─────────────────────────────────────────────►│
       │ 8. Turn 1 Setup (Draws, HQ placement)        │
       │◄─────────────────────────────────────────────┤
```

### Section 2.1: Handshake & Match Entry

#### FR-SES-001: Connection Handshake & Version Verification
- Upon WebSocket connection establishment, the client transmits a `ClientHandshake`:
  - `clientEngineVersion`: String (e.g., `"2.2.0"`). Must **strictly match** the server's engine version.
  - `protocolVersion`: Integer (e.g., `1`).
  - `matchAuthToken`: Cryptographically signed JWT issued by the matchmaking API.
- If the protocol version or engine version does not strictly match, the server immediately returns a `HandshakeReject` packet with an error code and closes the connection.

#### FR-SES-002: Deck Registration & Server Validation
- In compliance with **FR-SET-001**:
  - The client submits its complete deck payload: 1 designated Headquarters card ID and 50+ main deck card blueprint IDs.
  - The server verifies:
    1. The HQ card belongs strictly to Concordia, Soulless, Path, or Animus (cannot be Neutral).
    2. All main deck cards match the HQ's faction or the Neutral faction.
    3. Main deck contains >= 50 cards.
    4. Rarity copy limits are strictly adhered to: Common (max 4), Uncommon (max 3), Rare (max 2), Mythic (max 1).
    5. The card `common_scratch_company` is strictly excluded (match-generated tokens are prohibited from deck construction).
  - If validation fails, the match aborts with `DECK_LEGALITY_VIOLATION`.

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
  3. Players have a 25-second countdown to return `MulliganCommitRequest` containing an array of zero or more card instance IDs to set aside.
  4. Upon receiving both commits (or timer expiration):
     - The server draws replacement cards for the set-aside cards.
     - The set-aside cards are shuffled back into the deck.
     - The server transmits `MulliganResultNotification` containing the finalized opening hand to each respective player and notifying opponents of the number of cards swapped.

#### FR-SES-005: Disconnect Grace Window & State Rehydration
- In compliance with **FR-TRG-002**:
  - If a player disconnects mid-match (detected within 3,000 ms of silence via `FR-NET-004`), the server enters a **10-second Disconnect Grace Window**.
  - **Prompt Action Clock Pausing & Retrospective Compensation**: If the disconnect occurs while an interactive choice prompt is active, the server pauses the action clock and compensates for undetected silent time:
    ```text
    remainingClockMs = max(1000, deadlineEpochMs - lastReceivedClientPacketEpochMs)
    ```
    This completely eliminates network silent-countdown race conditions where a prompt could expire before disconnection was detected.
  - If the disconnected player reconnects within the 10-second window:
    1. The client sends `ReconnectRequest` with the active `matchAuthToken`.
    2. The server responds with a complete `FullStateSnapshot` containing:
       - Public board state (all assets across HQ, Reserves, Main Theater).
       - Controller's private hand instances.
       - Discard pile public contents.
       - Current phase, active player, and resource pools.
       - Active stack items and pending prompts (with remaining paused `remainingClockMs` restored).
       - Last processed `sequenceId`.
    3. The client renders the state instantaneously, clears its animation queue, unpauses the prompt clock, and resumes play.
  - If the player fails to reconnect within the 10-second grace window, the server commits deterministic fallbacks for any active choices and begins a 60-second match forfeit countdown.

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
  - When `SEARCH_DECK` or `LOOK_TOP_CARDS` executes, the server sends the card blueprint array **exclusively** to the searching player's client via `DeckInspectionPrompt`.
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
    
    // Server-to-Client
    ServerEventStream event_stream = 20;
    PlayerChoicePrompt choice_prompt = 21;
    StateSyncSnapshot state_snapshot = 22;
    CommandRejection rejection = 23;
    HeartbeatMessage heartbeat = 24;
  }
}
```

---

### Section 4.2: Client Command Taxonomy (Client -> Server)

#### FR-MSG-002: Client Commands
Clients issue explicit commands corresponding directly to Engine FRD actions:

```protobuf
message ClientCommand {
  string command_id = 1;
  uint64 ack_sequence_id = 2; // Last sequence_id observed by client
  
  oneof action {
    // Match Orchestration & Entry (FR-SES-001 - FR-SES-004)
    MatchJoinCommand match_join = 1;
    FirstPlayerChoiceCommand first_player_choice = 2;
    MulliganCommitCommand mulligan_commit = 3;

    // Inherent Actions (FR-INH-001 - Bypasses Stack)
    MarchCommand march = 10;
    AttackCommand attack = 11;
    GenerateResourceCommand generate_resource = 12;
    DeployScratchCommand deploy_scratch = 13;
    ConvertResourceCommand convert_resource = 14;
    
    // Stack Actions (FR-QUE-001, FR-TGT-101)
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
In fulfillment of **FR-INH-001** through **FR-INH-004** and **FR-TGT-101**:

```protobuf
// ============================================================================
// Match Orchestration Commands (FR-SES-001 - FR-SES-004)
// ============================================================================

message MatchJoinCommand {
  string match_token = 1;
  string hq_blueprint_id = 2;
  repeated string deck_blueprint_ids = 3;
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

// ============================================================================
// Inherent Action Commands (FR-INH-001 - FR-INH-004, Bypasses Resolution Stack)
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

message MarchCommand {
  string asset_instance_id = 1;
}

message AttackCommand {
  string attacker_asset_id = 1;
  // Standard attack targets 1 asset or HQ; Berserk units (FR-KWD-002o) may target up to 2 assets
  repeated string target_asset_ids = 2; 
}

message DeployScratchCommand {
  string hq_asset_id = 1; // Exhausts HQ to execute Scratch 1 (FR-INH-004)
}

message ConvertResourceCommand {
  int32 requisition_spent = 1; // Positive multiple of 2 (converts 2R -> 1T instantaneously, FR-RES-003)
}

// ============================================================================
// Stack Action Payloads (FR-TGT-101, FR-QUE-001)
// ============================================================================

enum TargetRole {
  TARGET_ROLE_UNSPECIFIED = 0;
  TARGET_ROLE_SOURCE = 1;
  TARGET_ROLE_TARGET = 2;
  TARGET_ROLE_SACRIFICE_COST = 3; // For additional deployment/activation sacrifice costs
  TARGET_ROLE_DISCARD_COST = 4;   // For activated ability hand discard costs
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
  TargetRole target_role = 4;// Role qualifier for multi-target actions (SOURCE, TARGET, SACRIFICE_COST, DISCARD_COST)
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
  string ability_type = 3; // "MANEUVER_EFFECT", "ACTIVATED", "TRIGGERED", "DEPLOYMENT"
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
  string effective_blueprint_id = 3; // "common_scratch_company"
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
  string destination_zone = 3; // "DISCARD"
}
```

---

### Section 4.4: Synchronous Choice Prompt Protocol

#### FR-MSG-005: Choice Prompts & Server Authority
In compliance with **FR-TRG-002** and **FR-MOD-001**:
- When stack execution requires an interactive decision, the server issues a `PlayerChoicePrompt`:
  ```protobuf
  message PlayerChoicePrompt {
    string prompt_id = 1;
    string prompt_type = 2; // "LOOK_TOP_CARDS", "MODAL_SELECT", "TARGET_SELECT"
    uint64 deadline_epoch_ms = 3; // Exact UTC millisecond deadline
    int32 min_selections = 4;
    int32 max_selections = 5;
    repeated ChoiceOption options = 6;
  }

  message ChoiceOption {
    int32 option_index = 1;
    string entity_or_card_id = 2;
    string description = 3;
  }

  message ChoiceResponse {
    string prompt_id = 1;
    repeated int32 selected_option_indices = 2;
    repeated int32 bottom_deck_ordered_indices = 3; // Ordered card indices returned to bottom of deck
    bool pass_or_decline = 4; // True to decline optional prompt
  }
  ```
- **Deadline Enforcement**: The client UI displays an active countdown based on `deadline_epoch_ms - currentServerTimeMs()`.
- If the client's `ChoiceResponse` is received after `deadline_epoch_ms`, the server rejects it and commits the deterministic fallback (`FR-MOD-001`).

---

## DOMAIN 5: STATE SYNCHRONIZATION, PACING & ANIMATION FLOW

```text
[ Server Engine Thread ]                   [ Godot Client ]
         │                                         │
         │ 1. EventBatch (Seq 42-45)               │
         ├────────────────────────────────────────►│ [ Receives Batch ]
         │    - March (Seq 42)                     │ ├── Enqueues to FIFO Queue
         │    - Attack (Seq 43)                    │ ├── Pop 42: Animate March
         │    - Damage Stage 1-5 (Seq 44)          │ ├── Pop 43: Animate Attack
         │    - Asset Destroyed (Seq 45)           │ ├── Pop 44: Animate 5-Stage Mit.
         │                                         │ └── Pop 45: Animate Death/Reveal
         │ 2. Ready for Next Command               │
         │    (Server state committed immediately) │
```

### Section 5.1: Decoupled Animation Queue

#### FR-PAC-001: Client-Side FIFO Event Queue
- The Godot client maintains a decoupled, local `AnimationEventQueue`.
- When a `ServerEventStream` packet arrives, the client immediately stores the latest state in an unrendered mirror state cache, while passing the atomic event structs into the queue.
- Visual components consume the queue sequentially:
  - `EVENT_MARCH` triggers sprite translation across the board.
  - `EVENT_DAMAGE_APPLIED` spawns floating combat text and flashes defensive shields.
  - `EVENT_DESTROY` plays death dissolve and triggers the reveal flip if a Scratch Company proxy is destroyed.

#### FR-PAC-002: Animation Fast-Forwarding on Inbound Prompts
- If an interactive `PlayerChoicePrompt` arrives while the client's `AnimationEventQueue` has pending background visual items:
  - The client automatically accelerates animation playback rates (e.g., 3.0x speed) or collapses non-essential visual transitions to guarantee that the player is presented with the prompt UI with maximum remaining time on their action clock.

---

### Section 5.2: Damage Mitigation Pipeline Pacing

#### FR-PAC-003: 5-Stage Mitigation Event Batching
- In compliance with **FR-REP-004**:
  - The 5-stage automated damage mitigation pipeline executes instantaneously as a single atomic transaction on the server.
  - To enable rich visual presentation without roundtrip stalls, the server outputs an array of discrete sub-stage results in a single event packet:
    ```protobuf
    message DamagePipelineStageEvent {
      string source_asset_id = 1;
      string initial_target_asset_id = 2;
      int32 raw_damage_amount = 3;
      
      // Stage 1: Redirection
      string final_target_asset_id = 4;
      // Stage 2: Resistant Flat Reduction
      int32 resistant_absorbed = 5;
      // Stage 3: Stalwart Capping
      int32 stalwart_capped_amount = 6;
      // Stage 4: Dynamic Armor Absorption
      int32 armor_absorbed = 7;
      int32 remaining_armor = 8;
      // Stage 5: Health Damage Committed
      int32 health_damage_committed = 9;
      int32 final_current_health = 10;
    }
    ```
  - The Godot client sequences this single packet into a fluid visual cascade (Shield Deflect &rarr; Stalwart Flash &rarr; Armor Cracking &rarr; Health Bar Drop) without requiring server roundtrips.

---

### Section 5.3: Desync Verification via Public State Hashing

#### FR-PAC-004: Public State Checksum Calculation
- At the conclusion of every phase transition and stack unwinding sequence, the server computes a 64-bit CRC/FNV-1a hash of the **Public Board State**:
  - Combined Health and Combat Power of all assets in HQ, Reserves, and Main Theater.
  - Readiness states (`READIED` / `EXHAUSTED`) of all assets.
  - Total card counts in Hand, Deck, and Attached zones for each player.
  - Complete list of card blueprint IDs in both public Discard piles.
  - Current Requisition and Tactics pools.
  - *Secret card identities in hands, decks, and face-down Scratch proxies are strictly excluded from the hash.*
- **Effective Layer 6 Attribute Rule**: To eliminate desync vulnerabilities with masked entities, the state hasher calculates asset Health and Combat Power strictly using post-mask effective values from Layer 6 of the Attribute Pipeline (`FR-LAY-001`, `FR-ENT-101`), guaranteeing 100% hash parity with client-visible proxy states.
- The hash is transmitted via `PublicStateHash`.
- The Godot client computes the identical hash across its mirrored public state cache:
  - If a hash mismatch is detected, the client flags a desync and automatically requests a `FullStateSnapshot` (`FR-SES-005`), seamlessly recovering the board without disconnecting.

---

## DOMAIN 6: SPECTATOR ARCHITECTURE & REPLAY SYSTEM

### Section 6.1: Live Spectator Protocol

#### FR-SPC-001: Spectator Fog-of-War Enforcement
- Live spectators connecting via a `SpectatorToken` are treated by the server as a **neutral 3rd-party observer**.
- The server applies full Fog-of-War sanitization (`FR-SEC-001`, `FR-SEC-002`):
  - Spectators **cannot** view either player's private hand contents.
  - Spectators receive the identical double-blind `scratch_proxy_id` representation for all Scratch Companies on the battlefield and in attached zones.
  - Spectators only see underlying card identities when an asset enters a public Discard pile.
- **Anti-Sniping Protection**: This eliminates any incentive or possibility of players using live spectator feeds to cheat or gain hidden information.

---

### Section 6.2: Match Replay File Format (`.rec`)

#### FR-SPC-002: Master Replay Data Architecture
- Upon match conclusion, the server compiles and archives a complete match recording (`.rec` file).
- **Zero Information Hiding in Replays**:
  - Unlike live broadcasts, the replay file contains the **100% unmasked historical truth**.
  - Includes initial shuffled deck order arrays, exact card IDs of all drawn cards, and the real underlying card identities of all Scratch Companies.
- **File Structure**:
  - `ReplayHeader`: Match GUID, Engine Version, Date, Player Metadata, Deck Lists.
  - `InitialState`: Shuffled deck states, HQ designations.
  - `EventLog`: Monotonic array of all `GameEnvelope` packets exchanged, enriched with unmasked underlying entity IDs.
- **Client Playback**: The Godot client's Replay Viewer reads the `.rec` file, offering analytical toggles:
  - *Fog-of-War Mode*: Simulates what Player 1 or Player 2 saw in real time.
  - *Omniscient Mode*: Unmasks all face-down Scratch Companies, tucked sub-cards, and hand contents for tournament broadcasting, coaching, and balance analysis.

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
  - `ERR_ASSET_EXHAUSTED`: Attempted action requiring Readied state (`FR-INH-001`).
  - `ERR_INVALID_RESOURCE_CHOICE`: Attempted resource generation invalid for current zone/keywords (e.g. non-Versatile unit generating Tactics in Reserves, or selecting an unsupported resource type; `FR-INH-002`, `FR-KWD-002i`).
  - `ERR_ACTION_CLOCK_EXPIRED`: Choice response received past deadline (`FR-TRG-002`).
- Upon receiving a rejection, the Godot client rolls back any local optimistic drag-and-drop animation and plays an invalid-action audio cue.

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
