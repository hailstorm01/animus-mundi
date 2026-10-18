# Product Manager Devil's Advocate Audit: Client-Engine Protocol & Networking FRD (FRD v1.3)

- **Target Document:** [docs/frd/04_Client_Engine_Protocol_Networking_FRD.md](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/04_Client_Engine_Protocol_Networking_FRD.md)
- **Auditor:** PM Devil's Advocate (`pm-devils-advocate`)
- **Evaluation Date:** 2026-10-18
- **Status:** PENDING (Active review queue)
- **Verdict:** REVISE WITH BLOCKERS

## 1. Executive Summary & Red-Team Assessment

`04_Client_Engine_Protocol_Networking_FRD.md` (v1.3) establishes a solid WebSocket/Protobuf session architecture, double-blind fog-of-war boundary, and animation pacing model. However, cross-referencing its Protobuf schemas (`FR-MSG-001`–`FR-MSG-005`) against the 26 atomic effect primitives in `02_Card_Data_Pipeline_FRD.md` (`FR-PRM-001`) and the rules in `01_Game_Rules_FRD.md` uncovers **two Blocker-severity protocol gaps**, **two High-severity conflicts/gaps**, and **one Medium-severity wire schema conflict**:
1. **Blocker Protocol Gaps (`DA-NET-001`, `DA-NET-002`):**
   - `FR-PRM-001` in FRD 02 defines 26 atomic effect primitives, and `FR-PAC-004` requires client and server to compute an identical 64-bit `PublicStateHash` over all assets' **Health, Armor, Combat Power, and Readiness**, plus **Hand, Deck, Attached, and Discard** contents after every `ServerEventStream` batch. Yet `EngineEvent.detail` in `FR-MSG-004` only supports a subset of those primitives—completely omitting `EngineEvent` payloads for `HEAL`, `ADD_ARMOR`, `MODIFY_STAT`, `REGROUP`/`EXHAUST` (readiness changes), `GRANT_KEYWORD`, `GRANT_SUBTYPE`, `PUT_CARD_ON_DECK`, `RECYCLE_DISCARD_TO_DECK`, and `REGISTER_DELAYED_TRIGGER`, while leaving `AssetTuckedEvent`, `ScratchReturnToOwnerHandEvent`, and `ScratchReturnToOpponentHandEvent` (defined in `FR-SEC-002`) unregistered in `EngineEvent.detail`.
   - `GameEnvelope` (`FR-MSG-001`) and `ServerEventStream` (`FR-MSG-004`) reference undefined Protobuf message types (`StateSyncSnapshot`, `HeartbeatMessage`, `PublicStateHash`), and the pre-game setup messages in `FR-SES-003` and `FR-SES-004` (`FirstPlayerChoicePrompt`, `FirstPlayerDetermined`, `MulliganPrompt`, `MulliganResultNotification`) have no Protobuf definitions or envelope tags.
2. **High Conflicts & Gaps (`DA-NET-003`, `DA-NET-004`):** `FR-MSG-002` and `FR-MSG-003` state that Off-Stack Economic Operations (`GenerateResourceCommand`, `ConvertResourceCommand`) *"Require Empty Stack"*, directly contradicting `FR-RES-001`, `FR-RES-003`, `FR-CAP-001`, `FR-QUE-001`, and `FR-ENG-010`; `PlayerChoicePrompt` and `ChoiceResponse` (`FR-MSG-005`) cannot represent triggered `DISTRIBUTED_POOL` value allocations (*Hammerhand Bivouac*), `LOOK_TOP_CARDS` free deployments (*Light of the World Soul*, *Devilhorn Scouting Party*), or in-resolution `DEPLOY_FREE` zone selections.
3. **Medium Wire Conflict (`DA-NET-005`):** `AttackCommand` (`FR-MSG-003`) uses `repeated string target_asset_ids` rather than `TargetBinding`, conflicting with `TARGET_ROLE_SECONDARY_BERSERK_TARGET` (`FR-MSG-003`, `FR-ENG-031`), and `MatchConcludedEvent.conclusion_reason` omits `FORFEIT_TIMEOUT` (`FR-SES-005`).

*(Note: Game and rule balance design questions are excluded per audit scope.)*

## 2. Severity-Ordered Punch-List

### 🔴 Blocker / Critical
- **[DA-NET-001] [Gap / Conflict] Missing EngineEvent Schemas in FR-MSG-004 for 11 Atomic Primitives & Readiness Transitions, Guaranteeing PublicStateHash Desyncs**
  - **Requirement Citations:** `FR-MSG-004`, `FR-PAC-004`, `FR-SEC-002`
  - **Vulnerability / Flaw:** `02_Card_Data_Pipeline_FRD.md` (`FR-PRM-001`) defines 26 atomic effect primitives, and `FR-PAC-004` requires the client to compute a 64-bit FNV-1a `PublicStateHash` across its local mirror state after every `ServerEventStream` batch covering all battlefield assets' **Health, dynamic Armor pools, Combat Power, and Readiness (`READIED`/`EXHAUSTED`)**, plus **Hand, Deck, Attached, and Discard** contents. However, `EngineEvent.detail` in `FR-MSG-004` (lines 494–510) defines only 15 event variants and lacks wire events for:
    1. `HEAL` (removing damage counters / increasing Current Health)
    2. `ADD_ARMOR` (incrementing an asset's dynamic Armor pool outside of damage mitigation)
    3. `MODIFY_STAT` (applying or expiring Layer 5 Combat Power or Maximum Health modifiers)
    4. `REGROUP`, `EXHAUST`, Ready Phase unexhausting, and upfront action exhaustion (transitioning asset `Readiness` between `READIED` and `EXHAUSTED`)
    5. `GRANT_KEYWORD` and `GRANT_SUBTYPE` (granting/expiring dynamic keywords and subtypes)
    6. `PUT_CARD_ON_DECK` and `RECYCLE_DISCARD_TO_DECK` (moving assets/cards to `DECK` and updating deck/discard counts)
    7. `REGISTER_DELAYED_TRIGGER`
    8. Unlinked `FR-SEC-002` messages: `FR-SEC-002` (lines 186–224) defines `AssetToDiscardEvent` (which duplicates `AssetDestroyedEvent` with conflicting field tag numbers), `ScratchReturnToOwnerHandEvent`, `ScratchReturnToOpponentHandEvent`, and `AssetTuckedEvent` (`TUCK_UNDER`), none of which are included in `EngineEvent.detail` (`FR-MSG-004`). Similarly, `AssetMovedEvent` and `AssetDeployedEvent` lack `source_zone` and `source_card_instance_id` fields needed when returning a card from `DISCARD` to `HAND` (*Funeral Fungus*) or deploying a card from `DISCARD` or `DECK` (`DEPLOY_FREE`).
  - **Exploit / Failure Scenario:** Any batch in which an asset heals (*Dolus Infusion*), gains armor (*Fortify*, *Dominator Warsuit*), receives a power/health modifier (*Chainbreaker's Tenacity*, *Poisoned Rations*), or changes readiness (*Tempest Runner*, Ready Phase, or any inherent action) updates the server's `PublicStateHash` (`FR-PAC-004`), but `ServerEventStream` has no `EngineEvent` payload to convey that state change to the client. The client's mirror state hash immediately mismatches on 100% of such batches, triggering constant `RequestStateSnapshot` desync recoveries.
  - **Remedy:** Expand `EngineEvent.detail` in `FR-MSG-004` to support all 26 primitives in `FR-PRM-001` and all hashed public state attributes in `FR-PAC-004`—specifically adding `AssetReadinessChangedEvent`, `AssetStatModifiedEvent` (`HEAL`, `ADD_ARMOR`, `MODIFY_STAT`, `GRANT_KEYWORD`, `GRANT_SUBTYPE`), `CardZoneMovedEvent` (unifying `RETURN_TO_HAND`, `PUT_CARD_ON_DECK`, `RECYCLE_DISCARD_TO_DECK`, and `ScratchReturnToOwnerHandEvent` / `ScratchReturnToOpponentHandEvent`), `AssetTuckedEvent` (from `FR-SEC-002`), and `DelayedTriggerRegisteredEvent`; add `source_zone` and `source_card_instance_id` to `AssetDeployedEvent`; and remove the redundant `AssetToDiscardEvent` in `FR-SEC-002` in favor of `AssetDestroyedEvent`.

- **[DA-NET-002] [Gap / Conflict] Missing Protobuf Message Definitions in Domain 4 for Pre-Game Setup Prompts, StateSyncSnapshot / FullStateSnapshot, PublicStateHash, and HeartbeatMessage**
  - **Requirement Citations:** `FR-SES-003`, `FR-SES-004`, `FR-SES-005`, `FR-MSG-001`, `FR-MSG-004`, `FR-PAC-004`
  - **Vulnerability / Flaw:**
    1. *Undefined Referenced Protobuf Types*: `GameEnvelope.payload` (`FR-MSG-001`, lines 272–274) references `StateSyncSnapshot state_snapshot = 22` and `HeartbeatMessage heartbeat = 24`, and `ServerEventStream` (`FR-MSG-004`, line 487) references `PublicStateHash state_hash = 4`. None of `StateSyncSnapshot`, `HeartbeatMessage`, or `PublicStateHash` has a Protobuf `message` definition anywhere in the FRD. Furthermore, `FR-SES-005` (line 134) and `FR-PAC-004` (line 758) refer to `StateSyncSnapshot` by a conflicting name (`FullStateSnapshot`).
    2. *Missing Pre-Game Setup Server Messages*: `FR-SES-003` and `FR-SES-004` mandate server-to-client messages `FirstPlayerChoicePrompt`, `FirstPlayerDetermined`, `MulliganPrompt`, and `MulliganResultNotification`, yet while their client-to-server counterparts (`FirstPlayerChoiceCommand`, `MulliganCommitCommand`) are defined in `FR-MSG-002` / `FR-MSG-003`, the server-to-client messages have no Protobuf definitions and are absent from `GameEnvelope.payload` (`FR-MSG-001`).
  - **Exploit / Failure Scenario:** The Protobuf definitions in Domain 4 cannot compile under `protoc` due to missing definitions for `StateSyncSnapshot`, `HeartbeatMessage`, and `PublicStateHash`, and client/server engineers lack wire schemas for First-Player Determination, Mulligan prompts/results, and full state snapshot rehydration.
  - **Remedy:** Add complete Protobuf `message` definitions in Domain 4 for `StateSyncSnapshot` (standardizing the name across `FR-SES-005`, `FR-MSG-001`, and `FR-PAC-004` and including all fields listed in `FR-SES-005` lines 135–142), `PublicStateHash`, `HeartbeatMessage`, and the pre-game setup messages (`FirstPlayerChoicePrompt`, `FirstPlayerDetermined`, `MulliganPrompt`, `MulliganResultNotification`), and register them in `GameEnvelope.payload` / `EngineEvent.detail`.

### 🟠 High
- **[DA-NET-003] [Conflict] FR-MSG-002 and FR-MSG-003 Falsely State That Off-Stack Economic Operations "Require Empty Stack"**
  - **Requirement Citations:** `FR-MSG-002`, `FR-MSG-003`
  - **Vulnerability / Flaw:** In `FR-MSG-002` (line 324) and `FR-MSG-003` (line 388), the wire specification comments for `GenerateResourceCommand` and `ConvertResourceCommand` state: `// Off-Stack Economic Operations (FR-CAP-001, FR-INH-001 - Require Empty Stack, Resolve Immediately Off-Stack)`. This directly contradicts `01_Game_Rules_FRD.md` (`FR-RES-001` line 159, `FR-RES-003` line 166, `FR-CAP-001` line 211, `FR-QUE-001` line 292) and `05_Headless_Game_Engine_FRD.md` (`FR-ENG-010` line 342), which all explicitly specify that exhausting assets to generate resources and converting `2R -> 1T` **do not require the Resolution Stack to be empty**.
  - **Exploit / Failure Scenario:** Implementing wire validation from `FR-MSG-002` / `FR-MSG-003` will reject `GenerateResourceCommand` and `ConvertResourceCommand` with `ERR_STACK_NOT_EMPTY` whenever the stack is non-empty, violating `FR-RES-001`, `FR-RES-003`, `FR-CAP-001`, `FR-QUE-001`, and `FR-ENG-010`.
  - **Remedy:** Correct the comments in `FR-MSG-002` (line 324) and `FR-MSG-003` (line 388) to state: `// Off-Stack Economic Operations (FR-RES-001, FR-RES-003, FR-CAP-001 - Stack-Exempt; Do NOT Require Empty Stack, Resolve Immediately Off-Stack)`.

- **[DA-NET-004] [Gap] PlayerChoicePrompt and ChoiceResponse (FR-MSG-005) Cannot Represent Distributed Pool Triggers, LOOK_TOP_CARDS Free Deployments, or In-Resolution Free-Deploy Zone Choices**
  - **Requirement Citations:** `FR-MSG-003`, `FR-MSG-005`
  - **Vulnerability / Flaw:** Three interactive choice patterns required by `01_Game_Rules_FRD.md` and `02_Card_Data_Pipeline_FRD.md` cannot be serialized in `PlayerChoicePrompt` and `ChoiceResponse` (`FR-MSG-005`):
    1. *Triggered `DISTRIBUTED_POOL` Allocations (`Hammerhand Bivouac`)*: *Hammerhand Bivouac* (`"At the start of your turn, distribute 2 armor among allied units in the same zone"`) is a `TRIGGERED` ability with `selector: "DISTRIBUTED_POOL", poolValue: 2` (`FR-ABL-011`). Because it triggers automatically on `EVENT_TURN_START` rather than via a client command (`FR-MSG-003`), the server prompts the player via `PlayerChoicePrompt`. However, `ChoiceResponse` only has `repeated int32 selected_option_indices = 2` and lacks an `allocated_values` (or `repeated TargetBinding`) field to specify how the 2 armor is allocated across the chosen units.
    2. *`LOOK_TOP_CARDS` Free-Deployment Partitioning (`Light of the World Soul`, `Devilhorn Scouting Party`)*: `PartitionedSelection.category_or_zone` in `FR-MSG-005` (line 644) lists `"HAND"`, `"DISCARD"`, `"TOP_OF_DECK"`, and `"BOTTOM_OF_DECK"`, omitting `"DEPLOY_RESERVES"` / `"DEPLOY_MAIN_THEATER"` required when partitioning inspected top-of-deck cards into free battlefield deployments (`FR-MOD-001` bullet 4, `FR-TRN-003a`).
    3. *In-Resolution `DEPLOY_FREE` Destination Zone Choice*: When a card deployed via `DEPLOY_FREE` (*Mutation Swell*, *World Soul Messenger*, *Eerie Remnant*) has printed `deploymentZones: ["MAIN_THEATER", "RESERVES"]`, `FR-TRN-003a` bullet 1 requires the deploying player to choose which zone it enters during resolution, which has no prompt type or field in `FR-MSG-005`.
  - **Exploit / Failure Scenario:** A player resolving *Hammerhand Bivouac*'s start-of-turn trigger with 2 allied units in the Main Theater cannot specify `2 armor to Unit A` vs. `1 armor to Unit A and 1 armor to Unit B` in `ChoiceResponse`; similarly, *Light of the World Soul* cannot partition inspected cards into free deployments.
  - **Remedy:** Add `repeated int32 allocated_values = 7` (parallel to `selected_option_indices`) or `repeated TargetBinding target_bindings = 7` to `ChoiceResponse` in `FR-MSG-005`, add `"DEPLOY_RESERVES"` and `"DEPLOY_MAIN_THEATER"` to `PartitionedSelection.category_or_zone`, and allow `ChoiceResponse` / `PlayerChoicePrompt` to select destination zones for multi-zone `DEPLOY_FREE` cards.

### 🟡 Medium
- **[DA-NET-005] [Conflict] AttackCommand Wire Schema vs. TARGET_ROLE_SECONDARY_BERSERK_TARGET and Missing FORFEIT_TIMEOUT in MatchConcludedEvent**
  - **Requirement Citations:** `FR-MSG-003`, `FR-MSG-004`, `FR-SES-005`, `FR-NET-004`
  - **Vulnerability / Flaw:**
    1. `FR-MSG-003` (line 443) defines `TARGET_ROLE_SECONDARY_BERSERK_TARGET = 7` inside `enum TargetRole` (a field of `TargetBinding`), and `05_Headless_Game_Engine_FRD.md` (`FR-ENG-031`, line 525) states that `AttackCommand` binds its optional secondary `Berserk` target via `TargetRole: TARGET_ROLE_SECONDARY_BERSERK_TARGET`. However, `AttackCommand` in `FR-MSG-003` (lines 420–425) defines `repeated string target_asset_ids = 2` and does not include `TargetBinding` at all.
    2. `FR-SES-005` (line 123) defines `FORFEIT_TIMEOUT` and `FR-NET-004` (line 48) defines forfeit due to protocol abuse (`POLICY_VIOLATION`), but `MatchConcludedEvent.conclusion_reason` in `FR-MSG-004` (line 571) lists only `"HQ_DESTROYED"`, `"EMPTY_DECK_DRAW"`, and `"CONCEDE"`.
  - **Exploit / Failure Scenario:** Engine and client implementations will conflict on whether `AttackCommand` passes `Berserk` secondary targets via `target_asset_ids[1]` or `TargetBinding`, and timeout forfeitures lack a valid `MatchConcludedEvent.conclusion_reason` value.
  - **Remedy:** Align `FR-MSG-003` and `FR-ENG-031` (either documenting that `AttackCommand` uses `target_asset_ids[0]` for primary and `target_asset_ids[1]` for secondary `Berserk` target, or adding `repeated TargetBinding target_bindings = 3` to `AttackCommand` for `Berserk` and upfront-locked attack triggers like *Samsara, Living Tempest*), and add `"FORFEIT_TIMEOUT"` and `"FORFEIT_PROTOCOL_ABUSE"` to `MatchConcludedEvent.conclusion_reason` in `FR-MSG-004`.

### 🔵 Low / Clarification
_None._

## 3. Redundancy & De-Scoping Ledger
| Requirement ID | Fully Subsumed By | Rationale / Recommendation |
| :--- | :--- | :--- |
| `FR-SEC-002` (`AssetToDiscardEvent` message) | `FR-MSG-004` (`AssetDestroyedEvent` / `CardZoneMovedEvent`) | Defines a duplicate Protobuf schema (`AssetToDiscardEvent`) with field tag numbers that conflict with `AssetDestroyedEvent` in `FR-MSG-004`. Consolidate all `EngineEvent` message definitions in `FR-MSG-004` and reference them from `FR-SEC-002`. |

## 4. Socratic Interrogation Register
_None — game and rule balance design questions are excluded per audit scope; all specification conflicts, gaps, ambiguities, underdeveloped schemas, and redundancies are itemized with concrete remedies in Sections 2 and 3._
