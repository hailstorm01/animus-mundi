# Product Manager Devil's Advocate Audit: Client-Engine Protocol & Networking FRD (v1.2)

- **Target Document:** [docs/frd/04_Client_Engine_Protocol_Networking_FRD.md](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/04_Client_Engine_Protocol_Networking_FRD.md)
- **Auditor:** PM Devil's Advocate (`pm-devils-advocate`)
- **Evaluation Date:** 2026-10-17
- **Status:** PENDING (Active review queue)
- **Verdict:** REVISE WITH BLOCKERS

## 1. Executive Summary & Red-Team Assessment
`04_Client_Engine_Protocol_Networking_FRD.md` provides a strong security boundary for double-blind Scratch Companies, fog-of-war sanitization, and disconnection reserve time banks. However, auditing `04` against recent changes in `01_Game_Rules_FRD.md`, `02_Card_Data_Pipeline_FRD.md`, `03_Card_Compiler_FRD.md`, and `05_Headless_Game_Engine_FRD.md` exposes **two critical blockers**: (1) `ClientCommand` in `FR-MSG-002` contains duplicate Protobuf field tags (`1` and `2`), making the `.proto` specification syntactically uncompilable, and (2) `FR-MSG-002`, `FR-MSG-003`, `FR-MSG-004`, and `FR-PAC-004` still classify `March`, `Attack`, and `Deploy Scratch Company` as off-stack actions that bypass the Resolution Stack, directly contradicting `FR-CAP-001` / `FR-INH-001` in `01` and `FR-ENG-010` / `FR-ENG-012` in `05`.

## 2. Severity-Ordered Punch-List

### 🔴 Blocker / Critical
- **[DA-NET-001] Uncompilable Protobuf Schema in `FR-MSG-002`: Duplicate Field Tags `1` and `2` in `ClientCommand`**
  - **Requirement Citations:** `FR-NET-003`, `FR-MSG-002`
  - **Vulnerability / Flaw:** In `FR-MSG-002` (lines 278–288), `message ClientCommand` assigns field numbers `1` and `2` twice within the same message scope:
    - Top-level fields: `string command_id = 1;` and `uint64 ack_sequence_id = 2;`
    - Inside `oneof action`: `MatchJoinCommand match_join = 1;`, `FirstPlayerChoiceCommand first_player_choice = 2;`, `MulliganCommitCommand mulligan_commit = 3;`
  - **Exploit / Failure Scenario:** In Protocol Buffers (`proto3`), fields inside a `oneof` block share the field-number namespace of their enclosing `message`. Compiling `FR-MSG-002` with `protoc` fails immediately with a fatal syntax error (`Field number 1 has already been used in "ccg.network.ClientCommand"`).
  - **Remedy:** Renumber the Match Orchestration commands inside `ClientCommand.action` in `FR-MSG-002` to non-colliding field numbers (e.g., `match_join = 3`, `first_player_choice = 4`, `mulligan_commit = 5`).

- **[DA-NET-002] Contradiction With Stack-Pushed Tactical Inherent Actions (`FR-CAP-001`, `FR-INH-001`, `FR-ENG-010`, `FR-ENG-012`)**
  - **Requirement Citations:** `FR-MSG-002`, `FR-MSG-003`, `FR-MSG-004`, `FR-PAC-004`
  - **Vulnerability / Flaw:** In `01_Game_Rules_FRD.md` (`FR-CAP-001`, `FR-INH-001`) and `05_Headless_Game_Engine_FRD.md` (`FR-ENG-010`, `FR-ENG-012`, `FR-ENG-031`), inherent capabilities are divided into **Instantaneous Economic Operations** (`Generate Resources` and `Convert Requisition`, which resolve immediately off-stack) and **Stack-Pushed Tactical Actions** (`March`, `Attack`, and `Deploy Scratch Company`, which require an empty stack and push frames onto the Resolution Stack). However, `04_Client_Engine_Protocol_Networking_FRD.md` was never updated to reflect this rule:
    1. `FR-MSG-002` (line 289) and `FR-MSG-003` (line 341) label `MarchCommand`, `AttackCommand`, and `DeployScratchCommand` under `// Inherent Actions (FR-INH-001 - Bypasses Resolution Stack)`.
    2. `StackPushedEvent.ability_type` (`FR-MSG-004`, line 491) enumerates only `"MANEUVER_EFFECT"`, `"ACTIVATED"`, `"TRIGGERED"`, and `"DEPLOYMENT"`, omitting `"ATTACK"` (`FRAME_ATTACK`), `"MARCH"`, and `"DEPLOY_SCRATCH"`.
    3. `FR-PAC-004` (line 691) groups March, Attack, and Scratch Deploy with off-stack resource actions as separate from stack unwinding sequences.
  - **Exploit / Failure Scenario:** When the headless engine (`FR-ENG-012`, `FR-ENG-031`) processes `MarchCommand`, `AttackCommand`, or `DeployScratchCommand`, it pushes a frame onto the Resolution Stack and emits `StackPushedEvent` before resolving triggers in Step 2. A Godot client built to `04` will either fail to parse the `StackPushedEvent.ability_type` or animate the March/Attack immediately before pre-combat triggers (*Blind Defile*, *Aggressive N*) resolve!
  - **Remedy:** Update `FR-MSG-002`, `FR-MSG-003`, `FR-MSG-004`, and `FR-PAC-004` to separate **Off-Stack Economic Operations** (`GenerateResourceCommand`, `ConvertResourceCommand`) from **Stack-Pushed Tactical Inherent Actions** (`MarchCommand`, `AttackCommand`, `DeployScratchCommand`), and add `"ATTACK"`, `"MARCH"`, and `"DEPLOY_SCRATCH"` to `StackPushedEvent.ability_type`.

### 🟠 High
- **[DA-NET-003] Missing `catalogChecksum` in `FR-SES-001` Handshake & Missing Wire Messages for Handshake/Rehydration (`FR-MSG-001`, `FR-MSG-002`)**
  - **Requirement Citations:** `FR-SES-001`, `FR-SES-005`, `FR-PAC-004`, `FR-MSG-001`, `FR-MSG-002`
  - **Vulnerability / Flaw:**
    1. `02_Card_Data_Pipeline_FRD.md` (`FR-DAT-004`) and `03_Card_Compiler_FRD.md` (`FR-CMP-025`, `FR-CMP-026`) mandate that during `FR-SES-001` connection negotiation, the server validates that the client's 256-bit `catalogChecksum` matches the server's authoritative card catalog. Yet `FR-SES-001` in `04_Client_Engine_Protocol_Networking_FRD.md` (lines 75–79) lists only `clientEngineVersion`, `protocolVersion`, and `matchAuthToken`—omitting `catalogChecksum`.
    2. Furthermore, `ClientHandshake`, `HandshakeAck`, `HandshakeReject` (`FR-SES-001`), `ReconnectRequest` (`FR-SES-005` line 125), and `RequestStateSnapshot` (`FR-PAC-004` line 701) are missing from the `GameEnvelope` (`FR-MSG-001`) and `ClientCommand` (`FR-MSG-002`) Protobuf definitions.
  - **Exploit / Failure Scenario:** A client with a stale `blueprints.json` (`catalogChecksum` mismatch) under the same `clientEngineVersion` passes `FR-SES-001` handshake validation and immediately desyncs during match play, and has no defined Protobuf `oneof` payload in `GameEnvelope` or `ClientCommand` to send `ClientHandshake`, `ReconnectRequest`, or `RequestStateSnapshot`.
  - **Remedy:** Add `catalogChecksum` (`string catalog_checksum`) to `ClientHandshake` in `FR-SES-001`, and add `ClientHandshake`, `HandshakeAck`, `HandshakeReject`, `ReconnectRequest`, and `RequestStateSnapshot` to the Protobuf message definitions in Domain 4 (`FR-MSG-001` / `FR-MSG-002`).

- **[DA-NET-004] Missing `TARGET_ROLE_SECONDARY_BERSERK_TARGET` in `FR-MSG-003` & Hardcoded 10s Prompt Clock in `FR-PAC-001` vs. 15s Active-Player Clock (`FR-MOD-001`)**
  - **Requirement Citations:** `FR-MSG-003`, `FR-MSG-005`, `FR-PAC-001`
  - **Vulnerability / Flaw:**
    1. `02_Card_Data_Pipeline_FRD.md` (`FR-TGT-101`) and `05_Headless_Game_Engine_FRD.md` (`FR-ENG-031`) reference `TargetRole: SECONDARY_BERSERK_TARGET` for upfront Berserk secondary target binding, but `enum TargetRole` in `FR-MSG-003` (lines 384–392) omits `TARGET_ROLE_SECONDARY_BERSERK_TARGET` (or relies on positional `AttackCommand.target_asset_ids` without aligning with `FR-TGT-101` / `FR-ENG-031`).
    2. `01_Game_Rules_FRD.md` (`FR-MOD-001` item 2) and `05_Headless_Game_Engine_FRD.md` (`FR-ENG-017`) mandate a **15-second** action clock for the **active player** and a **10-second** action clock for the **non-active player** on interactive choice prompts during stack resolution. However, `FR-PAC-001` (line 638) hardcodes a 10-second clock (`deadlineEpochMs = currentServerTimeMs() + 3000 + 10000`) for all `PlayerChoicePrompt` issuances.
  - **Exploit / Failure Scenario:** When the active player resolves *Light of the World Soul* (inspecting and partitioning 7 cards) or *Cliff of Portents* (searching a 40-card deck), `FR-PAC-001` gives them only 10 seconds instead of the 15 seconds mandated by `FR-MOD-001` and `FR-ENG-017`.
  - **Remedy:** Add `TARGET_ROLE_SECONDARY_BERSERK_TARGET = 7;` to `enum TargetRole` in `FR-MSG-003`, and update `FR-PAC-001` so `PlayerChoicePrompt` deadlines grant `15000` ms when prompting the active player and `10000` ms when prompting the non-active player (in addition to the 3,000 ms animation grace buffer).

### 🟡 Medium
- **[DA-NET-005] Missing `blueprint_id` in `ChoiceOption` & Missing Prompt Types in `PlayerChoicePrompt` (`FR-MSG-005`)**
  - **Requirement Citations:** `FR-SEC-004`, `FR-MSG-005`
  - **Vulnerability / Flaw:** `FR-SEC-004` states that during `SEARCH_DECK` or `LOOK_TOP_CARDS`, the server sends the card blueprint array to the searching player via `DeckInspectionPrompt`, while `FR-MSG-005` models all interactive prompts through `PlayerChoicePrompt` with `prompt_type` values `"LOOK_TOP_CARDS"`, `"MODAL_SELECT"`, `"TARGET_SELECT"`, and `"TRIGGER_ORDER_SELECT"`. However:
    1. `ChoiceOption` (lines 582–586) contains only `option_index`, `entity_or_card_id`, and `description`—omitting `blueprint_id`! Because cards in a player's Deck or an opponent's Hand (`REVEAL_HAND`) were previously hidden under fog of war (`FR-SEC-003`, `FR-SEC-004`), the client does not know which `blueprint_id` corresponds to each `entity_or_card_id`.
    2. `PlayerChoicePrompt.prompt_type` omits `"SEARCH_DECK"`, `"REVEAL_HAND"`, `"REPLACEMENT_CONFLICT_SELECT"` (`FR-REP-001`), and `"OPTIONAL_TRIGGER_CONFIRM"` (`FR-MOD-001`).
  - **Exploit / Failure Scenario:** When a player activates *Cliff of Portents* (`SEARCH_DECK`), resolves *Hourglass Diviner* (`LOOK_TOP_CARDS`), or deploys *Drifting Concept* (`REVEAL_HAND`), the client receives `ChoiceOption` entries with opaque `entity_or_card_id` strings and no `blueprint_id` to render the card art, stats, or rules text in the choice modal.
  - **Remedy:** Add `string blueprint_id = 4;` to `ChoiceOption` in `FR-MSG-005` (populated when the prompt reveals cards to the choosing player), and expand `PlayerChoicePrompt.prompt_type` to include `"SEARCH_DECK"`, `"REVEAL_HAND"`, `"REPLACEMENT_CONFLICT_SELECT"`, and `"OPTIONAL_TRIGGER_CONFIRM"`.

## 3. Redundancy & De-Scoping Ledger
| Requirement ID | Fully Subsumed By | Rationale / Recommendation |
| :--- | :--- | :--- |
| `FR-MSG-002` | `FR-MSG-003` | Already marked `*(Subsumed by FR-MSG-003)*` in Section 4.2. |

## 4. Socratic Interrogation Register
*None — all findings in `04_Client_Engine_Protocol_Networking_FRD.md` are concrete protocol schema and cross-FRD synchronization fixes.*
