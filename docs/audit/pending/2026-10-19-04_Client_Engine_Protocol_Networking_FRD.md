# Product Manager Devil's Advocate Audit: Client-Engine Protocol & Networking FRD (Master FRD v1.5)

- **Target Document:** [docs/frd/04_Client_Engine_Protocol_Networking_FRD.md](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/04_Client_Engine_Protocol_Networking_FRD.md)
- **Auditor:** PM Devil's Advocate (`pm-devils-advocate`)
- **Evaluation Date:** 2026-10-19
- **Status:** PENDING (Active review queue)
- **Verdict:** CONDITIONAL PASS

## 1. Executive Summary & Red-Team Assessment

The Client-Engine Protocol & Networking specification (`04_Client_Engine_Protocol_Networking_FRD.md`) provides an exceptionally thorough network wire contract, complete with Protobuf schemas, fog-of-war security boundaries, unified reserve time bank disconnect rules, and decoupled animation queuing. However, several critical payload discrepancies and contract gaps exist: `CombatResolvedEvent` uses a scalar `attacker_damage_dealt` alongside `repeated string defender_asset_ids`, making it impossible to express distinct damage amounts to primary vs secondary Berserk defenders; `AssetTuckedEvent` omits `revealed_blueprint_id` and `card_instance_id`, making face-up card tucking impossible to render; `AssetStatModifiedEvent` contains a contradictory reference to "armor expiration" violating the core rules invariant that Armor never expires; `DamagePipelineStageEvent` omits a combat vs non-combat discriminator needed for Stage 3 Stalwart VFX; and `ReplayEventRecord` leaves stream directionality unspecified.

## 2. Severity-Ordered Punch-List

### 🔴 Blocker / Critical
_None._

### 🟠 High
- **[DA-NET-013] Broken Damage Representation for Multi-Defender Attacks in `CombatResolvedEvent`**
  - **Requirement Citations:** `FR-MSG-004`, `FR-CBT-002` (cross-ref `FR-KWD-002o`)
  - **Vulnerability / Flaw:** `CombatResolvedEvent` (lines 617-623) represents combat outcomes using `repeated string defender_asset_ids = 2`, but pairs it with a **scalar `int32 attacker_damage_dealt = 3`**. When an attacking unit with Berserk N attacks a primary defender and cleaves a secondary defender, `defender_asset_ids` contains two entity IDs. Because `attacker_damage_dealt` is a single scalar integer, it cannot convey how much damage was dealt to defender 1 (combat damage) versus defender 2 (flat Berserk cleave damage).
  - **Exploit / Failure Scenario:** The Godot client receives `CombatResolvedEvent` with `defender_asset_ids: ["unit_primary", "unit_secondary"]` and `attacker_damage_dealt: 6`. The client cannot determine whether `unit_primary` took 4 damage and `unit_secondary` took 2, or if 6 was split evenly, resulting in desynced combat numbers and corrupted floating text.
  - **Remedy:** Refactor `CombatResolvedEvent` to serialize per-defender damage, such as `repeated int32 attacker_damage_dealt_per_defender = 3` or a nested `repeated DefenderCombatOutcome defender_outcomes = 2` struct containing `defender_asset_id`, `damage_taken`, and `retaliation_dealt`.

- **[DA-NET-014] Omission of Blueprint Identity and Card Instance for Face-Up Tucking in `AssetTuckedEvent`**
  - **Requirement Citations:** `FR-MSG-004`, `FR-SEC-002` (cross-ref `01_Game_Rules_FRD.md` `FR-ENT-102`, `02_Card_Data_Pipeline_FRD.md` `FR-PRM-001`)
  - **Vulnerability / Flaw:** In FRD 01 `FR-ENT-102` and FRD 02 `FR-PRM-001` (`TUCK_UNDER: hostTarget, faceDown: Boolean`), tucking cards into an asset's Attached sub-zone is explicitly permitted to occur **either face-down or face-up**. However, `AssetTuckedEvent` (lines 739-744) specifies:
    ```protobuf
    message AssetTuckedEvent {
      string host_asset_id = 1;
      string proxy_asset_id = 2;
      bool is_face_down = 3; // true
      // revealed_blueprint_id is STRICTLY OMITTED (FR-SEC-002, NO-AH-5)
    }
    ```
    The message explicitly omits `revealed_blueprint_id` and assumes `is_face_down` is always true. If an effect tucks a card face-up, both clients need the blueprint identity to render the face-up tucked card frame under the host. Furthermore, `proxy_asset_id` only applies to battlefield assets; if a card tucked from Hand or Deck is tucked, it has a `card_instance_id`, which `AssetTuckedEvent` does not support.
  - **Exploit / Failure Scenario:** A card effect tucks a card from Hand or Discard face-up under an allied structure. The server emits `AssetTuckedEvent`, but clients cannot display which card was tucked because `revealed_blueprint_id` is omitted.
  - **Remedy:** Add `string revealed_blueprint_id = 4` (populated when `is_face_down == false`) and `string card_instance_id = 5` to `AssetTuckedEvent`.

### 🟡 Medium
- **[DA-NET-015] Contradictory Invariant: "Armor Expiration" in `AssetStatModifiedEvent`**
  - **Requirement Citations:** `FR-MSG-004` (cross-ref `01_Game_Rules_FRD.md` `FR-LAY-002`, `FR-KWD-002k`, `02_Card_Data_Pipeline_FRD.md` `DA-DAT-004`)
  - **Vulnerability / Flaw:** `AssetStatModifiedEvent` line 721 defines `int32 armor_delta = 3; // Dynamic armor addition (ADD_ARMOR) or expiration` and line 726 lists `string reason = 8; // e.g. "HEAL", "ADD_ARMOR", "MODIFY_STAT", "STAT_EXPIRATION"`. This directly contradicts the core rules invariant codified across all other FRDs: "Armor stacks additively across discrete grants, **never expires at the end of the turn**, and absorbs incoming damage 1-for-1" (`FR-LAY-002`, `FR-KWD-002k`, `DA-DAT-004`).
  - **Exploit / Failure Scenario:** A client developer reading `AssetStatModifiedEvent` implements an end-of-turn armor expiration routine that wipes players' armor pools, breaking defensive structures.
  - **Remedy:** Remove the words "or expiration" from line 721, and remove `"STAT_EXPIRATION"` from the armor delta context, reinforcing that armor is consumable strictly via damage absorption.

- **[DA-NET-016] Missing `is_combat_damage` Discriminator in `DamagePipelineStageEvent`**
  - **Requirement Citations:** `FR-MSG-004`, `FR-PAC-003` (cross-ref `01_Game_Rules_FRD.md` `FR-REP-004`)
  - **Vulnerability / Flaw:** `DamagePipelineStageEvent` (lines 868-887) streams the 5-stage damage mitigation cascade. Stage 3 (`Stalwart`) applies *strictly* to combat damage (`FR-REP-004`). However, `DamagePipelineStageEvent` lacks a boolean `bool is_combat_damage = 12;`. Clients cannot determine whether Stalwart was bypassed due to lack of the keyword or because the incoming damage was non-combat effect damage, hindering client visual shield telemetry.
  - **Exploit / Failure Scenario:** Godot client animation sequencer cannot distinguish between a combat hit blocked by armor versus a spell hit bypassing Stalwart, causing incorrect combat shield VFX.
  - **Remedy:** Add `bool is_combat_damage = 12;` to `DamagePipelineStageEvent`.

### 🔵 Low / Clarification
- **[DA-NET-017] Unspecified Replay Stream Directionality in `ReplayEventRecord`**
  - **Requirement Citations:** `FR-SPC-002` (cross-ref `05_Headless_Game_Engine_FRD.md` `FR-ENG-038`)
  - **Vulnerability / Flaw:** `ReplayEventRecord` stores `GameEnvelope envelope = 3;`, but does not clarify whether `event_log` records strictly server-emitted envelopes (`ServerEventStream`), client command envelopes (`ClientCommand`), or both interleaved. FRD 05 `FR-ENG-038` requires re-simulating from inputs, which requires client commands.
  - **Exploit / Failure Scenario:** Replay recorder logs only server events; replay validator cannot re-simulate match from player inputs.
  - **Remedy:** Explicitly clarify in `FR-SPC-002` that `event_log` records all committed client command envelopes and server event stream envelopes interleaved in chronological order.

## 3. Redundancy & De-Scoping Ledger
| Requirement ID | Fully Subsumed By | Rationale / Recommendation |
| :--- | :--- | :--- |
| `FR-MSG-002` | `FR-MSG-003` | ClientCommand action wrappers are unified directly into FR-MSG-003 concrete command schemas. |
| `FR-SES-003` | `FR-SET-003`, `FR-MSG-001`, `FR-MSG-004` | First-player determination rules and wire schemas are canonically owned by FRD 01 and Section 4.1. |
| `FR-SEC-005` | `FR-ENG-002`, `FR-DEK-002` | Server PRNG seed derivation and deck shuffling are owned by FRD 05 and FRD 01. |

## 4. Socratic Interrogation Register
1. *If the authoritative server already executes 100% deterministic rules and broadcasts full state synchronization snapshots upon desync, why is the client required to compute and verify a 64-bit FNV-1a public board state hash after every single event stream batch rather than verifying state periodically at turn boundaries? Does per-batch hashing add unnecessary wire overhead and client CPU stutter during large animation cascades?*
