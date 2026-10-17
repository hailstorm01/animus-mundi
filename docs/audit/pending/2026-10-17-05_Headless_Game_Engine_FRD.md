# Product Manager Devil's Advocate Audit: Headless Game Engine FRD (v1.1)

- **Target Document:** [docs/frd/05_Headless_Game_Engine_FRD.md](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/05_Headless_Game_Engine_FRD.md)
- **Auditor:** PM Devil's Advocate (`pm-devils-advocate`)
- **Evaluation Date:** 2026-10-17
- **Status:** PENDING (Active review queue)
- **Verdict:** CONDITIONAL PASS

## 1. Executive Summary & Red-Team Assessment
Following the v1.1 remediation pass (`a46120b`), `05_Headless_Game_Engine_FRD.md` is well-aligned with `01_Game_Rules_FRD.md` across the 3-tier entity architecture (`FR-ENT-001`–`FR-ENT-003`, `FR-SNP-101`), 6-layer attribute pipeline (`FR-ENG-025`), and 5-stage damage mitigation (`FR-ENG-034`). However, auditing the engine specification against the full 167-card catalog (`data/cards/*.csv`) and companion FRDs (`02` and `04`) reveals **three high-severity gaps**: (1) `MatchState` (`FR-ENG-006`) and the APNAP dispatcher (`FR-ENG-014`) have no runtime registry or lifecycle for **delayed triggered abilities** (*Rage's Wake*, *Redline Infusion*) that must survive `AssetInstance` deallocation upon death, (2) `FR-ENG-010`, `FR-ENG-013`, and `FR-ENG-025` omit execution rules for **hand-active cost reductions** (*Swirling Dustdevil*, *Boneyard Amalgam*), **discard-pile deployments** (*Veil Render*), and **allied-unit exhaustion costs** (*Funeral Fungus*), and (3) `FR-ENG-008` and `FR-ENG-017` leave the **transaction commit boundary across suspended choice prompts** ambiguous.

## 2. Severity-Ordered Punch-List

### 🔴 Blocker / Critical
*None identified.*

### 🟠 High
- **[DA-ENG-001] Missing Runtime Registry & Lifecycle in `MatchState` for Delayed Triggered Abilities (*Rage's Wake*, *Redline Infusion*)**
  - **Requirement Citations:** `FR-ENG-006`, `FR-ENG-014`, `FR-ENG-029`, `FR-ENG-035`, `FR-ENT-003`
  - **Vulnerability / Flaw:** `MatchState` (`FR-ENG-006`) tracks `Zones`, `ResolutionStack`, `ResourcePools`, and `EventBuffer`, while `AssetInstance` (`FR-ENT-003`) tracks `ActiveModifiers` (Layers 2, 3, and 5) and is immediately deallocated when the asset leaves the battlefield (`FR-ENT-003` line 302). Neither `MatchState` nor `FR-ENG-014` (APNAP Trigger Dispatcher) defines where **delayed triggered abilities** created by resolving Maneuvers (*Rage's Wake*: *"When that unit is killed this turn, deploy it from your discard pile to your Reserves for free"*; *Redline Infusion*: *"At the end of your turn, kill that unit"*) are stored, how they survive `AssetInstance` deallocation when the watched unit dies in `FR-ENG-035`, or how they track the underlying `CardInstance` into `DiscardZone`.
  - **Exploit / Failure Scenario:**
    1. Player A enacts *Rage's Wake* targeting an allied unit.
    2. Later that turn, that unit takes lethal damage (`FR-ENG-035`).
    3. Step 2 of `FR-ENG-035` captures `AssetSnapshot`, deallocates the `AssetInstance` wrapper (`FR-ENT-003`), moves the underlying `CardInstance` to `DiscardZone`, and emits `EVENT_DESTROY`.
    4. Because `AssetInstance` is already deallocated and `MatchState` has no `DelayedTriggers` registry watching `EVENT_DESTROY` for that entity, *Rage's Wake*'s delayed re-deployment trigger either fails to fire or cannot resolve the underlying `CardInstance` from `DiscardZone` (especially if the killed unit was a face-down Scratch Company whose underlying card turned face-up into a non-unit card in `DiscardZone`).
  - **Remedy:** Add a turn-scoped `DelayedTriggers` slice to `MatchState` (`FR-ENG-006`) evaluated by the APNAP Trigger Dispatcher (`FR-ENG-014`) on `EVENT_DESTROY` and `EVENT_TURN_END` (recording the watched `AssetID` and underlying `CardInstance.InstanceID`, and verifying in `DEPLOY_FREE` that the card in `DiscardZone` still satisfies the required `cardType`), and purge expired turn-scoped delayed triggers during Sub-Phase 4c (`FR-ENG-029`).

- **[DA-ENG-002] Missing Engine Execution Contracts for Non-Battlefield Cost Modifiers (`FR-CST-001`), Discard-Zone Deployments (*Veil Render*), and Allied-Unit Exhaustion Costs (*Funeral Fungus*)**
  - **Requirement Citations:** `FR-ENG-010`, `FR-ENG-013`, `FR-ENG-025`
  - **Vulnerability / Flaw:**
    1. `FR-ENG-025` specifies that the 6-layer attribute pipeline evaluates strictly *"across all battlefield `AssetInstance` structs"*, leaving no evaluation contract in `FR-ENG-013` or `FR-ENG-025` for dynamic deployment cost reductions (`FR-CST-001`) printed on `CardInstance` structs residing in `HandZone` (*Swirling Dustdevil*, *Boneyard Amalgam*).
    2. `FR-ENG-013` defines upfront cost validation and deduction for `RESOURCE`, `EXHAUST_SELF`, `KILL_ASSET`, `SACRIFICE_ALLY`, and `DISCARD_CARD`, but omits `EXHAUST_ALLY` (`TARGET_ROLE_EXHAUST_COST`) required when a Maneuver card in `HandZone` requires exhausting a Readied allied unit on the battlefield as an additional enactment cost (*Funeral Fungus*).
    3. `FR-ENG-010` and `FR-ENG-012` do not specify validating static alternate-zone deployment permissions (*Veil Render* deployed from `DiscardZone` by killing an allied asset in addition to paying its deployment costs).
  - **Exploit / Failure Scenario:** Attempting to enact *Funeral Fungus*, deploy *Veil Render* from `DiscardZone`, or deploy *Swirling Dustdevil* / *Boneyard Amalgam* using reduced Requisition cost aborts in `FR-ENG-013` with `ERR_INSUFFICIENT_RESOURCES` or `ERR_INVALID_COMMAND_ARGUMENTS`.
  - **Remedy:** Update `FR-ENG-013` to: (1) evaluate card-level `COST_MODIFIER` queries on `CardInstance` structs in `HandZone`/`DiscardZone` when computing effective Requisition and Tactics costs (`FR-CST-001`), (2) validate `Readiness == READIED` and transition target allied assets bound via `TARGET_ROLE_EXHAUST_COST` (`EXHAUST_ALLY`) to `EXHAUSTED` (emitting `EVENT_EXHAUST`), and (3) allow `DeployCardCommand` from `DiscardZone` when permitted by the card's blueprint and its additional `KILL_ASSET` cost is paid.

- **[DA-ENG-003] Ambiguous Transaction Commit & Rollback Boundary Across Suspended Choice Prompts (`FR-ENG-008` vs. `FR-ENG-017`)**
  - **Requirement Citations:** `FR-ENG-008`, `FR-ENG-017`
  - **Vulnerability / Flaw:** `FR-ENG-008` states that every stack resolution step executes inside an atomic transaction (`tx.BeginTransaction()` ... `tx.Commit()`) that flushes `EventBuffer` only on `tx.Commit()`. Meanwhile, `FR-ENG-017` states that when a stack frame requires an interactive choice (`LOOK_TOP_CARDS`, `SEARCH_DECK`, `REVEAL_HAND`, or multi-target trigger selection), the engine *"suspends the active stack frame, transitions to `STATE_AWAITING_CHOICE`, freezes the Main Phase turn clock, and emits `PlayerChoicePrompt`."* It is unspecified whether yielding for a prompt commits the pre-prompt transaction immediately (flushing pre-prompt events and `PlayerChoicePrompt` to the wire) and opens a separate transaction when `ChoiceResponse` arrives.
  - **Exploit / Failure Scenario:**
    1. If the transaction is held open across network I/O while waiting for `ChoiceResponse`, `EventBuffer` is never flushed and the client never receives the events or prompt.
    2. Worse, if `tx.Rollback()` on a malformed `ChoiceResponse` rolls back the pre-prompt state and PRNG stream counters (`FR-ENG-008`), a player who sees an unfavorable top-deck reveal in `LOOK_TOP_CARDS` or `REVEAL_HAND` could intentionally submit a malformed `ChoiceResponse` to roll back the entire stack frame!
  - **Remedy:** Specify in `FR-ENG-008` and `FR-ENG-017` that suspending execution at an interactive choice boundary **commits** the current transaction (persisting `STATE_AWAITING_CHOICE` and flushing `EventBuffer` and `PlayerChoicePrompt`), and that each subsequent `ChoiceResponse` executes inside a new transaction where validation failure rolls back *only* the invalid response attempt (emitting `CommandRejection` with `ERR_INVALID_COMMAND_ARGUMENTS` while remaining in `STATE_AWAITING_CHOICE` until a valid response is received or the action clock expires).

### 🟡 Medium
- **[DA-ENG-004] Contradiction Between `FR-ENG-041` (`ERR_DESYNC_DETECTED` Terminates Match) and `04_Client_Engine_Protocol_Networking_FRD.md` (`FR-PAC-004` Snapshot Recovery)**
  - **Requirement Citations:** `FR-ENG-005`, `FR-ENG-041`, `FR-PAC-004`, `FR-SES-005`
  - **Vulnerability / Flaw:** The error table in Section 8.1 of `05_Headless_Game_Engine_FRD.md` defines `ERR_DESYNC_DETECTED` with severity `Fatal`, trigger condition `StateChecksum mismatch between client and server (FR-PAC-004)`, and system action `Terminate match; emit replay dump.` This directly contradicts `04_Client_Engine_Protocol_Networking_FRD.md` (`FR-PAC-004` and `FR-SES-005`), where `StateChecksum` comparison occurs *client-side* and a mismatch triggers a non-destructive `FullStateSnapshot` rehydration request (rate-limited to 1 per 3,000 ms) without terminating the match.
  - **Exploit / Failure Scenario:** If the server terminated a live match whenever a client reported a `StateChecksum` mismatch, any losing player could send a bogus desync report to force-terminate the match and avoid defeat.
  - **Remedy:** Update the `ERR_DESYNC_DETECTED` entry in Section 8.1 to specify that in live networked matches a client checksum mismatch triggers rate-limited `FullStateSnapshot` rehydration per `FR-SES-005`/`FR-PAC-004`, while fatal match termination applies strictly to offline `.rec` replay validation (`FR-ENG-038`) and fuzz testing (`FR-ENG-039`).

- **[DA-ENG-005] Misattributed *Portent Scrivener* Citation, Dual Phase/Turn Event Hooks in `FR-ENG-009`, and Wire Command Name Drift (`FR-ENG-017`, `FR-ENG-031`)**
  - **Requirement Citations:** `FR-ENG-009`, `FR-ENG-017`, `FR-ENG-031`
  - **Vulnerability / Flaw:**
    1. `FR-ENG-009` Sub-Phase 2b (line 320) cites *Portent Scrivener* as a Draw Phase card-draw trigger, whereas *Portent Scrivener* (`units-and-structures-animus.csv`) is `"When this unit is killed on your turn, draw a card."` (`EVENT_DESTROY`).
    2. `FR-ENG-009` Sub-Phase 1b and Sub-Phase 4a state that the engine emits `EVENT_READY_PHASE` / `EVENT_TURN_START` and `EVENT_END_PHASE` / `EVENT_TURN_END` without distinguishing whether both hooks are emitted or which one triggers `"At the start/end of your turn"` abilities.
    3. `FR-ENG-017` and `FR-ENG-031` reference `SubmitChoiceCommand` and `DeclareAttackCommand`, whereas `04_Client_Engine_Protocol_Networking_FRD.md` (`FR-MSG-002`, `FR-MSG-003`, `FR-MSG-005`) names those wire messages `ChoiceResponse` and `AttackCommand`.
  - **Exploit / Failure Scenario:** Mismatched event hook emissions (`EVENT_READY_PHASE` vs. `EVENT_TURN_START`) and wire message names (`SubmitChoiceCommand` vs. `ChoiceResponse`) cause integration failures between the Go engine (`05`), compiler (`03`), and Protobuf wire layer (`04`).
  - **Remedy:** Remove the *Portent Scrivener* reference from `FR-ENG-009` Sub-Phase 2b, standardize on `EVENT_TURN_START` (Sub-Phase 1b) and `EVENT_TURN_END` (Sub-Phase 4a) for turn-boundary triggers, and align command struct names in `FR-ENG-017` (`ChoiceResponse`) and `FR-ENG-031` (`AttackCommand`) with `04_Client_Engine_Protocol_Networking_FRD.md`.

## 3. Redundancy & De-Scoping Ledger
| Requirement ID | Fully Subsumed By | Rationale / Recommendation |
| :--- | :--- | :--- |
| `FR-ENG-003` | `FR-ENG-002` | Already marked `Subsumed by FR-ENG-002` in Section 1.1. |
| `FR-ENG-026`, `FR-ENG-027`, `FR-ENG-028` | `FR-ENG-025` & `FR-ENG-041` | Already marked `Subsumed by FR-ENG-025` in Section 5.1. |

## 4. Socratic Interrogation Register
*None — all findings in `05_Headless_Game_Engine_FRD.md` are concrete architectural and cross-FRD synchronization fixes.*
