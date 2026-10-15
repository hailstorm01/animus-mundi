# Headless Game Engine - Functional Requirements Document (Master FRD v1.0)
**Authoritative Headless Match Kernel, Deterministic State Machine & Execution Engine**

---

## EXECUTIVE SUMMARY & REVISION HISTORY

This document establishes the authoritative technical specification for the **Card Town Headless Game Engine**, governing **Phase 3 (Core Headless Game Engine Implementation)** of the product roadmap in [`docs/PRODUCT_SUMMARY.md`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/PRODUCT_SUMMARY.md#L198-L212). 

While [`01_Game_Rules_and_Card_Modeling_FRD.md`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md) codifies abstract rules, entity taxonomies, and formulas, and [`03_Card_Compiler_FRD.md`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/03_Card_Compiler_FRD.md) produces declarative catalog assets (`blueprints.json`), this document specifies the software kernel that executes those rules with **100% server-authoritative determinism**.

The engine is engineered as a standalone, dependency-free **Go (Golang)** module with zero graphics, UI, or game engine libraries. It operates as the core match runner within the dedicated server daemon on GCP Agones ([`FR-SES-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/04_Client_Engine_Protocol_Networking_FRD.md#L68)) and as a bundled local sidecar executable for offline play and bot testing, preserving complete decoupling from the Godot 4.x presentation client ([`FR-MSG-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/04_Client_Engine_Protocol_Networking_FRD.md#L248)).

```text
+-----------------------------------------------------------------------------------+
|                           CARD TOWN HEADLESS ENGINE                               |
|                                                                                   |
|  +-------------------------+   +-----------------------------------------------+  |
|  |   Deterministic PRNG    |   |     Transactional Event-Emitting Match        |  |
|  |  Multi-Stream Generator |   |                  State                        |  |
|  |   (PCG32 / SplitMix64)  |   |    (Zones, Entities, Resources, Clocks)       |  |
|  +------------+------------+   +-----------------------+-----------------------+  |
|               |                                        |                          |
|               v                                        v                          |
|  +-------------------------+   +-----------------------+-----------------------+  |
|  |  6-Layer Attribute Calc |   |            Resolution Stack                   |  |
|  | Pipeline & Dirty Caching|   |       LIFO Frames & Trigger Queue             |  |
|  |  (Layers 1 to 6 / Clamps)   |   |        (APNAP Dispatcher / Rollback)          |  |
|  +------------+------------+   +-----------------------+-----------------------+  |
|               |                                        |                          |
|               +-------------------+--------------------+                          |
|                                   |                                               |
|                                   v                                               |
|  +-----------------------------------------------------------------------------+  |
|  |                     Declarative Ability AST Interpreter                     |  |
|  |            (Type-Switch Visitor & 14 Atomic Primitives Execution)           |  |
|  +--------------------------------+--------------------------------------------+  |
|                                   |                                               |
|                                   v                                               |
|  +-----------------------------------------------------------------------------+  |
|  |                       Combat Resolution Engine                              |  |
|  |          (Declaration -> Reaction Window -> Simultaneous Damage)            |  |
|  +--------------------------------+--------------------------------------------+  |
|                                   |                                               |
|                                   v                                               |
|  +-----------------------------------------------------------------------------+  |
|  |              Authoritative Event Buffer & Checksum Generator                |  |
|  |                  (Protobuf Output & FR-PAC-004 Parity Hash)                 |  |
|  +-----------------------------------------------------------------------------+  |
+-----------------------------------------------------------------------------------+
```

### Initial Specification Scope (v1.0):
1. **Deterministic Engine Kernel & PRNG Subsystem (`FR-ENG-001` - `FR-ENG-005`)**: Multi-stream pseudo-random number generator, pure integer arithmetic, and rolling state hash generation.
2. **Authoritative Match State Machine & Phase Engine (`FR-ENG-006` - `FR-ENG-011`)**: Transactional event-emitting state model, delta-rollback checkpoints, turn sub-phases, and clean-stack action constraints.
3. **Resolution Stack & APNAP Trigger Dispatcher (`FR-ENG-012` - `FR-ENG-017`)**: LIFO stack runner, atomic cost deduction, APNAP simultaneous trigger ordering, and reactive replacement watchers.
4. **Declarative Ability AST Interpreter & Atomic Primitives (`FR-ENG-018` - `FR-ENG-024`)**: Polymorphic AST type-switch evaluator, target filtering, condition gates, and execution contracts for the 14 atomic primitives.
5. **6-Layer Attribute Recalculation Pipeline (`FR-ENG-025` - `FR-ENG-029`)**: Continuous evaluation engine, dirty-flag propagation, aura caching, and round-duration cleanup.
6. **Combat Resolution & Retaliation Engine (`FR-ENG-030` - `FR-ENG-035`)**: 3-step combat cycle orchestration, exhaustion locks, armor absorption, Stalwart caps, and state-based death batching.
7. **Headless Test Harness, Replay & Simulation Suite (`FR-ENG-036` - `FR-ENG-040`)**: Headless CLI runner, scripted scenario DSL, deterministic `.rec` replay validator, and combinatorial fuzz testing.
8. **Error Taxonomy, Resilience & Diagnostics**: Structured error code catalog (`ERR_ENG_*`) and parity assertion rules.

---

## ARCHITECTURAL DECISIONS FLAGGED FOR REVIEW

The following foundational design decisions represent inferences drawn from product vision pillars and technical constraints. Each decision is formally flagged for engineering and game design sign-off:

> [!IMPORTANT]
> **DECISION 1: Pure Go (Golang) Headless Core Runtime (`FR-ENG-001`)**
> - **Inference**: High-density server hosting on GCP Agones requires microsecond startup times, minimal RSS memory footprint (<30 MB per match container), and massive concurrent match orchestration.
> - **Decision**: The headless engine is implemented in pure Go (1.22+) as a standalone module (`cardtown/engine`) without dependencies on Godot or Cgo.
> - **Alternative Considered**: C# (.NET 8) or Rust. While C# offers direct Godot .NET sharing and Rust provides zero-GC execution, Go provides native first-party Agones integration, lightning-fast compilation (<1s test suites), trivial cross-compilation into static binaries, and superior container packing via lightweight goroutines (~2 KB per match loop).

> [!IMPORTANT]
> **DECISION 2: Transactional Event-Emitting State Architecture (`FR-ENG-006`, `FR-ENG-008`)**
> - **Inference**: Card games require strict transaction safety (reverting partial mutations if target validation fails or a player submits an illegal command) while maintaining high CPU throughput during high-frequency layer calculations.
> - **Decision**: The engine utilizes an in-place mutable state model wrapped in explicit transaction boundaries with lightweight delta-mementos for rollback. As state mutates within a transaction, authoritative domain events are buffered. If an action fails, the delta-memento restores touched entities and the event buffer is cleared; upon success, events are committed to the Protobuf wire stream and `.rec` replay file.
> - **Alternative Considered**: Pure immutable event sourcing (`(State, Action) -> (NewState, Events)`). Rejected due to severe GC allocation churn in Go: copying slices and struct graphs across dozens of cascading triggers and continuous 6-layer updates creates unacceptable GC pause risks.

> [!IMPORTANT]
> **DECISION 3: Direct Tree-Walking AST Interpreter with Type-Switch Dispatch (`FR-ENG-018`)**
> - **Inference**: The card catalog is loaded dynamically from `blueprints.json` ([`FR-CMP-025`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/03_Card_Compiler_FRD.md#L457)), requiring flexible ability execution that matches the declarative AST schema ([`FR-ABL-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/02_Card_Data_Pipeline_FRD.md#L154)).
> - **Decision**: Implement a direct tree-walking AST interpreter using Go type-switches across sealed interface structs. Each AST node evaluates directly against an execution context and dispatches to atomic effect primitives.
> - **Alternative Considered**: JIT bytecode compilation. Unnecessary complexity; benchmarks indicate direct tree-walking in Go evaluates card abilities in <5 microseconds per invocation, exceeding target throughput.

> [!IMPORTANT]
> **DECISION 4: Cross-Platform Pure Integer Arithmetic (`FR-ENG-004`)**
> - **Inference**: Cross-platform simulation parity and replay reproduction ([`FR-SPC-002`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/04_Client_Engine_Protocol_Networking_FRD.md#L735)) fail if floating-point calculations produce bitwise divergence across CPU architectures (x86_64 vs. ARM64).
> - **Decision**: All combat calculations, damage reductions, resource adjustments, and timer values use strict 32-bit signed/unsigned integer arithmetic (`int32`, `uint32`, `int64`). Floating-point types (`float32`, `float64`) are strictly prohibited in the engine domain model.
> - **Alternative Considered**: Fixed-point arithmetic (Q16.16). Unnecessary; Card Town rules require only discrete integer quantities (health, power, cost, turn seconds).

> [!IMPORTANT]
> **DECISION 5: Decoupled Thin-Client Presentation Contract (`FR-ENG-036`)**
> - **Inference**: Maintaining duplicate rules logic inside the Godot client creates desync drift, exposes hidden information (fog of war), and requires client app updates for card balance patches.
> - **Decision**: The Godot presentation client never executes game rules logic. It consumes Protobuf events emitted by the engine. For offline play and bot testing, the client spawns the compiled headless engine as a local background sidecar process and communicates via local WebSockets (`ws://127.0.0.1:PORT`).
> - **Alternative Considered**: Embedding the engine via Cgo/GDExtension. Adds brittle C-FFI bindings and GC boundary overhead without operational benefit.

---

## DOMAIN 1: DETERMINISTIC ENGINE KERNEL & PRNG SUBSYSTEM

```text
                     +---------------------------------------+
                     |           Master Seed (int64)         |
                     +-------------------+-------------------+
                                         |
               +-------------------------+-------------------------+
               |                         |                         |
               v                         v                         v
       +---------------+         +---------------+         +---------------+
       | Shuffle Stream|         | Target Stream |         | Coin / RNG    |
       | (PCG32 Salt A)|         | (PCG32 Salt B)|         | (PCG32 Salt C)|
       +-------+-------+         +-------+-------+         +-------+-------+
               |                         |                         |
               v                         v                         v
       Deck Shuffling             Random Targeting           Mulligan / Flip
       (Fisher-Yates)             Selection Pool             Decisions
```

### Section 1.1: Core Engine Lifecycle & Seeded Determinism

#### FR-ENG-001: Engine Modularity & Zero-Engine Dependency Invariant
- **Functional Contract**: The headless engine module MUST be completely decoupled from graphic rendering, audio, window management, and client GUI frameworks (including Godot engine packages).
- **Inputs**: Card catalog payload (`blueprints.json`), match configuration parameters (`MatchConfig`), player session descriptors (`PlayerSession`), and master 64-bit integer seed (`int64`).
- **Outputs**: Instantiated authoritative `MatchEngine` instance ready for match initialization.
- **Invariants**:
  - Zero imports of `github.com/godotengine/*` or OS windowing packages.
  - Zero Cgo compilation requirements for core rules evaluation; compiles with pure `CGO_ENABLED=0`.
  - Deterministic execution: identical inputs and seeds MUST produce bit-for-bit identical state histories across Linux (x86_64/ARM64), Windows, and macOS.

#### FR-ENG-002: Isolated Multi-Stream PRNG Architecture
- **Functional Contract**: The engine MUST isolate pseudo-random number generation across independent streams using the PCG32 or SplitMix64 algorithmic generator to prevent seed consumption side-effects from altering unrelated match events.
- **Streams**:
  1. `StreamShuffle`: Dedicated exclusively to initial deck shuffling ([`FR-SET-002`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L45)) and mid-match deck shuffles ([`FR-PRM-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L525)).
  2. `StreamTargeting`: Dedicated exclusively to resolving non-deterministic card targeting (e.g. *"deal 2 damage to a random enemy unit"*).
  3. `StreamMatchDecisions`: Dedicated to pre-game coin toss ([`FR-SET-003`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L54)) and tie-breaking procedures.
- **Invariants**: Shuffling a deck MUST NOT advance or alter the sequence of random target selections during combat.

#### FR-ENG-003: Deterministic Seed Derivation & Salt Invariants
- **Functional Contract**: The engine MUST derive child PRNG stream seeds from the master 64-bit seed using fixed, non-colliding 64-bit cryptographic salt constants:
  - `SeedShuffle = MasterSeed ^ 0x9E3779B97F4A7C15`
  - `SeedTargeting = MasterSeed ^ 0xBF58476D1CE4E5B9`
  - `SeedMatchDecisions = MasterSeed ^ 0x94D049BB133111EB`
- **Output**: Each stream advances deterministically on each invocation (`NextUint32()`, `NextIntBounded(max)`).
- **Error Handling**: Supplying a seed of `0` MUST be rejected with `ERR_ENG_INVALID_SEED` or re-seeded using a non-zero cryptographic entropy source.

#### FR-ENG-004: Pure Discrete Integer Arithmetic & Rounding Invariants
- **Functional Contract**: All numeric calculations within the engine MUST execute using discrete 32-bit signed integers (`int32`).
- **Rules**:
  - Division operations resulting in fractional quantities MUST perform deterministic floor integer truncation (`a / b`).
  - Damage calculation, flat resistance, armor deductions, and cost reductions MUST NEVER underflow below 0 unless an explicit signed delta contract is invoked.
  - Floating-point arithmetic (`float32`, `float64`) is strictly prohibited in rules evaluation.

#### FR-ENG-005: Frame Tick Invariants & Rolling State Hash Generation
- **Functional Contract**: The engine MUST maintain a strictly monotonic 64-bit tick counter (`TickNumber`) incremented on every committed transaction and phase transition.
- **State Checksum**: At the conclusion of every turn sub-phase and stack frame resolution, the engine MUST calculate a rolling 64-bit MurmurHash3 or FNV-1a state checksum (`StateChecksum`) across all public entity states, zone contents, and resource pools in conformance with [`FR-PAC-004`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/04_Client_Engine_Protocol_Networking_FRD.md#L450).
- **Output**: Transmitted to clients within periodic synchronization frames to detect desync divergence immediately.

---

## DOMAIN 2: AUTHORITATIVE MATCH STATE MACHINE & PHASE ENGINE

```text
                          +-------------------+
                          |    MATCH START    |
                          +---------+---------+
                                    |
                                    v
                     +------------------------------+
                     |         READY PHASE          |
                     |  - Sub-Phase 1a: Unexhaust   |
                     |  - Sub-Phase 1b: Requisition |
                     +--------------+---------------+
                                    |
                                    v
                     +------------------------------+
                     |          DRAW PHASE          |
                     |  - Sub-Phase 2a: Draw 1 card |
                     |    (Skip on Player 1 Turn 1) |
                     +--------------+---------------+
                                    |
                                    v
                     +------------------------------+ <------------------+
                     |          MAIN PHASE          |                    |
                     |  - Clean-Stack Actions       |  Action Execution  |
                     |  - Deployments / Maneuvers   |  & Stack Cleared   |
                     |  - Combat Declarations       |                    |
                     +--------------+---------------+ -------------------+
                                    |
                                    v
                     +------------------------------+
                     |          END PHASE           |
                     |  - Sub-Phase 4a: End Trigger |
                     |  - Sub-Phase 4b: Hand Discard|
                     |  - Sub-Phase 4c: Buff Decay  |
                     +--------------+---------------+
                                    |
                                    v
                     +------------------------------+
                     |     HANDOVER / NEXT TURN     |
                     +------------------------------+
```

### Section 2.1: Transactional Match State Model

#### FR-ENG-006: Transactional Event-Emitting State Model
- **Functional Contract**: The authoritative match state (`MatchState`) MUST be held in memory as an in-place mutable object graph, guarded by an active `Transaction` context.
- **Structure**:
  - `MatchID`: Unique string identifier.
  - `TurnNumber`: Monotonically increasing turn counter (starts at 1).
  - `ActivePlayer`: ID of the player currently taking their turn (0 or 1).
  - `CurrentPhase`: Current match phase (`PHASE_READY`, `PHASE_DRAW`, `PHASE_MAIN`, `PHASE_END`).
  - `Zones`: Zone topology container ([`FR-ZON-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L398) – [`FR-ZON-004`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L430)).
  - `ResolutionStack`: The LIFO execution stack ([`FR-QUE-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L110)).
  - `ResourcePools`: Requisition and Tactics counters per player.
  - `EventBuffer`: In-memory slice of pending `EngineEvent` objects emitted during the transaction.

#### FR-ENG-007: Match Initialization, Setup Handshake & Zone Instantiation
- **Functional Contract**: Upon receiving valid deck payloads conforming to [`FR-SET-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L29), the engine MUST instantiate the formal zone topology for both players:
  1. Instantiates `HeadquartersZone` containing exactly 1 designated HQ asset for Player 1 and Player 2 ([`FR-ZON-003`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L420)).
  2. Populates `DeckZone` with the 50+ card instances and performs initial Fisher-Yates shuffle via `StreamShuffle`.
  3. Deals initial 6-card hand into `HandZone` ([`FR-SET-002`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L45)).
  4. Initiates the 25-second partial mulligan window ([`FR-SET-002`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L45), [`FR-SES-004`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/04_Client_Engine_Protocol_Networking_FRD.md#L94)).
  5. Determines first player via `StreamMatchDecisions` ([`FR-SET-003`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L54)).

#### FR-ENG-008: Transaction Checkpoint, Delta-Rollback & Commit Lifecycle
- **Functional Contract**: Every player command, stack resolution step, and phase transition MUST execute inside an atomic transaction:
  - `tx := match.BeginTransaction()`
  - Before modifying state, the transaction records a lightweight delta-snapshot of touched entities, zones, and resource balances.
  - **On Failure / Illegal Command**: `tx.Rollback()` restores touched entities to their pre-transaction values, purges `EventBuffer`, and returns a structured error code (`ERR_ENG_*`). The active state remains uncorrupted.
  - **On Success**: `tx.Commit()` applies state mutations permanently, calculates the rolling state hash ([`FR-ENG-005`](#fr-eng-005-frame-tick-invariants--rolling-state-hash-generation)), flushes `EventBuffer` to the outgoing network/replay channel, and discards the delta-snapshot.

#### FR-ENG-009: Deterministic Turn Sequencing & Sub-Phase Transitions
- **Functional Contract**: The engine MUST execute turn phases in strict sequential order:
  - **Phase 1: Ready Phase**:
    - *Sub-Phase 1a (Unexhaust)*: Unexhaust all exhausted assets in the active player's Reserves, Main Theater, and HQ Zone.
    - *Sub-Phase 1b (Requisition)*: Add base turn Requisition income and reset unspent turn Tactics.
  - **Phase 2: Draw Phase**:
    - *Sub-Phase 2a (Draw)*: Draw 1 card from `DeckZone` into `HandZone`. Skipped on Turn 1 for the player taking the first turn ([`FR-SET-004`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L57)).
  - **Phase 3: Main Phase**:
    - Window for player-initiated deployments, maneuvers, asset exhaustion, and combat declarations.
  - **Phase 4: End Phase**:
    - *Sub-Phase 4a (End-of-Turn Triggers)*: Emit `EVENT_END_OF_TURN` and resolve queued abilities.
    - *Sub-Phase 4b (Hand Size Cleanup)*: Assert hand size $\le 10$; prompt or discard excess cards.
    - *Sub-Phase 4c (Temporary Modifier Decay)*: Deterministically purge all temporary buffs/debuffs whose duration has expired ([`FR-LAY-003`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L247)).
    - *Handover*: Increment `TurnNumber` and toggle `ActivePlayer`.

#### FR-ENG-010: Clean-Stack Manual Action Constraint & Priority Windows
- **Functional Contract**: Manual player actions (playing a maneuver card, deploying a unit/structure, manually exhausting an asset for resources, or declaring combat) MUST assert that `ResolutionStack.IsEmpty() == true` per [`FR-QUE-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L110) and [`FR-TRN-003a`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L16).
- **Exceptions**: Reactive triggered abilities, in-resolution free deployments, and replacement watchers execute without clean-stack constraints.
- **Violation**: If a player attempts a manual action while the stack contains pending items, the engine MUST abort the transaction with `ERR_ENG_DIRTY_STACK_ACTION`.

#### FR-ENG-011: Turn-Clock Timeout Resolution & Auto-Pass Invariants
- **Functional Contract**: The engine MUST track active turn duration against the 75-second turn clock ([`FR-TRN-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/04_Client_Engine_Protocol_Networking_FRD.md#L508)) and 60-second Reserve Time Bank ([`FR-SES-005`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/04_Client_Engine_Protocol_Networking_FRD.md#L100)).
- **Expiration Actions**:
  1. When regular turn clock reaches 0, the engine begins depleting the player's Reserve Bank.
  2. When Reserve Bank reaches 0, the engine issues an automated force-pass command.
  3. If pending choice prompts ([`FR-MSG-005`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/04_Client_Engine_Protocol_Networking_FRD.md#L400)) are unresolved at timeout, default selections are applied automatically (e.g., keep all cards, forfeit optional effects).

---

## DOMAIN 3: RESOLUTION STACK & APNAP TRIGGER DISPATCHER

```text
                  +-----------------------------------------+
                  |         EVENT OCCURS IN MATCH           |
                  |     (e.g., Unit Enters Battlefield)     |
                  +--------------------+--------------------+
                                       |
                                       v
                  +-----------------------------------------+
                  |       APNAP Trigger Dispatcher          |
                  |  1. Collect active player triggers      |
                  |  2. Collect non-active player triggers  |
                  +--------------------+--------------------+
                                       |
                                       v
                  +-----------------------------------------+
                  |        Resolution Stack (LIFO)          |
                  |  [Top: Non-Active Player Trigger]       |  <-- Pops and executes first
                  |  [Mid: Active Player Trigger]           |
                  |  [Bot: Initiating Maneuver Card]        |
                  +--------------------+--------------------+
                                       |
                                       v
                  +-----------------------------------------+
                  |   Pop Top Frame & Validate Targets      |
                  +--------------------+--------------------+
                                       |
                        +--------------+--------------+
                        | Valid Targets?              |
                        v                             v
                     [ YES ]                       [ NO ]
                        |                             |
                        v                             v
           Execute Atomic Primitives            Fizzle Frame & Log
           via FR-ENG-021                       ERR_ENG_TARGET_INVALID
```

### Section 3.1: Stack Execution & Trigger Ordering

#### FR-ENG-012: LIFO Stack Architecture & Stack Frame Taxonomy
- **Functional Contract**: The engine MUST implement the Resolution Stack as a Last-In, First-Out (`LIFO`) data structure storing strongly-typed `StackFrame` instances.
- **Frame Types**:
  - `FRAME_MANEUVER`: A played maneuver spell awaiting resolution.
  - `FRAME_TRIGGERED_ABILITY`: An ability queued in response to an in-game event.
  - `FRAME_ACTIVATED_ABILITY`: An ability activated by paying an asset cost.
  - `FRAME_REPLACEMENT_WATCHER`: A modifier intercepting a pending effect.
- **Invariants**: When priority passes and all players pass sequentially, the engine pops and resolves the top stack frame.

#### FR-ENG-013: Atomic Cost Validation, Resource Reservation & Deduction
- **Functional Contract**: Prior to pushing an action or ability to the stack, the engine MUST atomically validate and deduct all required costs:
  - Requisition cost ([`FR-SET-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L29)).
  - Tactics cost ([`FR-SET-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L29)).
  - Physical exhaustion requirement (asset must currently be unexhausted).
  - Explicit sacrifice or discard costs ([`FR-ABL-007`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/02_Card_Data_Pipeline_FRD.md#L273)).
- **Failure**: If any cost component cannot be paid in full, the transaction aborts with `ERR_ENG_INSUFFICIENT_RESOURCE` or `ERR_ENG_ASSET_EXHAUSTED`. Zero partial deduction permitted.

#### FR-ENG-014: APNAP Trigger Dispatching & Simultaneous Event Ordering
- **Functional Contract**: When an in-game event triggers multiple abilities simultaneously, the engine MUST queue them onto the stack using the **APNAP (Active Player, Non-Active Player)** convention:
  1. All triggers controlled by the **Active Player** are pushed onto the stack in timestamp order.
  2. All triggers controlled by the **Non-Active Player** are pushed onto the stack on top of the Active Player's triggers.
- **Invariants**: Because the stack is LIFO, the Non-Active Player's triggers resolve first, preserving deterministic priority without race conditions.

#### FR-ENG-015: Dynamic Target Verification & Invalidation Invariants
- **Functional Contract**: Targets selected upon action declaration MUST be re-verified for legality immediately prior to execution when the stack frame pops:
  - Target entity must still exist in its declared zone ([`FR-TGT-102`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L16)).
  - Target must still satisfy all `FilterDef` constraints ([`FR-ABL-008`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/02_Card_Data_Pipeline_FRD.md#L292)).
- **Partial Invalidation**: If an ability targets multiple entities and at least one target remains legal, the ability executes against the remaining legal targets.
- **Complete Invalidation**: If all declared targets have become invalid, the stack frame **fizzles**: effect execution is skipped, the card is sent to the discard pile, and `ERR_ENG_TARGET_INVALID` is logged.

#### FR-ENG-016: Reactive Replacement Watcher Interceptor Pipeline
- **Functional Contract**: Replacement effects ([`FR-REP-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L14) – [`FR-REP-004`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L14)) MUST intercept and modify event payloads *before* they are applied to state or pushed to the stack.
- **Application Order**:
  1. Self-replacement effects (effects on the entity taking damage or moving zones).
  2. Controller's replacement effects.
  3. Opponent's replacement effects.
- **Recursion Guard**: A replacement effect MUST NOT apply more than once to the same root event.

#### FR-ENG-017: Choice Prompt Yield & Resumption Protocol
- **Functional Contract**: When an effect requires runtime player selection (scry/look-at-top-N cards, modal choice, distributed damage allocation; [`FR-MSG-005`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/04_Client_Engine_Protocol_Networking_FRD.md#L400), [`FR-MOD-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L18)), the engine MUST suspend execution of the active stack frame:
  1. Emits `ChoicePrompt` payload to the target player client.
  2. Sets engine state to `STATE_AWAITING_CHOICE`.
  3. Freezes turn clock and starts the 30-second choice countdown timer.
  4. Resumes stack execution immediately upon receiving a valid `ChoiceResponse` matching prompt UUID.

---

## DOMAIN 4: DECLARATIVE ABILITY AST INTERPRETER & ATOMIC PRIMITIVES

```text
                  +-----------------------------------------+
                  |         AbilityDef (JSON AST)           |
                  +--------------------+--------------------+
                                       |
                                       v
                  +-----------------------------------------+
                  |      Type-Switch AST Interpreter        |
                  +----+-------------+----------------+-----+
                       |             |                |
         +-------------+             |                +-------------+
         v                           v                              v
  [TRIGGERED]                  [ACTIVATED]                  [CONTINUOUS_AURA]
  - Event Matching             - Cost Check                 - Layer Register
  - Gate Verification          - Activation Lock            - Scope Filter
         |                           |                              |
         +---------------------------+------------------------------+
                                     |
                                     v
                  +-----------------------------------------+
                  |        14 Atomic Primitives             |
                  |  DEAL_DAMAGE   | HEAL     | DRAW_CARD   |
                  |  EXHAUST_ASSET | READY    | BUFF_STATS  |
                  |  DEPLOY_SCRATCH| DISCARD  | DESTROY     |
                  +-----------------------------------------+
```

### Section 4.1: Ability Evaluation & Execution Engine

#### FR-ENG-018: Polymorphic Ability AST Interpreter & Type-Switch Visitor
- **Functional Contract**: The engine MUST evaluate declarative `AbilityDef` nodes conforming to [`FR-ABL-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/02_Card_Data_Pipeline_FRD.md#L154) using an exhaustive type-switch visitor pattern.
- **Handled Archetypes**:
  - `TRIGGERED`: Listens for engine event matches; evaluates conditions; queues effect frames ([`FR-ABL-002`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/02_Card_Data_Pipeline_FRD.md#L174)).
  - `ACTIVATED`: Validates costs and timing; queues effect frames ([`FR-ABL-003`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/02_Card_Data_Pipeline_FRD.md#L198)).
  - `CONTINUOUS_AURA`: Registers dynamic attribute modifiers into the 6-layer pipeline ([`FR-ABL-004`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/02_Card_Data_Pipeline_FRD.md#L217)).
  - `REPLACEMENT_WATCHER`: Registers pre-event payload mutators ([`FR-ABL-005`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/02_Card_Data_Pipeline_FRD.md#L236)).
  - `MANEUVER_EFFECT`: Direct spell effect resolution ([`FR-ABL-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/02_Card_Data_Pipeline_FRD.md#L154)).

#### FR-ENG-019: Dynamic FilterDef & Target Resolution Evaluator
- **Functional Contract**: The engine MUST evaluate `FilterDef` predicates ([`FR-ABL-008`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/02_Card_Data_Pipeline_FRD.md#L292) – [`FR-ABL-011`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/02_Card_Data_Pipeline_FRD.md#L368)) against active game entities:
  - `FactionScope`: `ALLIED`, `ENEMY`, `ANY`.
  - `ZoneScope`: `MAIN_THEATER`, `RESERVES`, `HQ_ZONE`, `HAND`, `GRAVEYARD`.
  - `CardType`: `UNIT`, `STRUCTURE`, `HEADQUARTERS`, `MANEUVER`.
  - `AttributeComparisons`: Power, defense, exhaustion state, cost, subtype tags.
- **Invariants**: Targeting filters MUST evaluate against *effective* stats calculated by the 6-layer pipeline, not raw blueprint base stats.

#### FR-ENG-020: Condition Gate & ValueToken Expression Resolver
- **Functional Contract**: The engine MUST resolve dynamic numeric tokens (`ValueToken`) and boolean gates (`ConditionDef`) at runtime per [`FR-ABL-006`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/02_Card_Data_Pipeline_FRD.md#L254) and [`FR-ABL-012`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/02_Card_Data_Pipeline_FRD.md#L391).
- **Supported ValueTokens**:
  - `CONSTANT`: Fixed integer value.
  - `COUNT_MATCHING`: Number of entities matching a `FilterDef` (e.g., *"for each allied unit"*).
  - `PROPERTY_OF_TARGET`: Value derived from a target entity (e.g., target's power).
  - `EXHAUSTED_ASSET_COUNT`: Current count of exhausted assets.

#### FR-ENG-021: The 14 Atomic Primitives Execution Contracts
- **Functional Contract**: The engine MUST implement discrete execution handlers for all 14 atomic effect primitives defined in [`FR-PRM-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L525):
  1. `DEAL_DAMAGE`: Applies damage to target asset with armor absorption and reduction.
  2. `HEAL`: Removes accumulated damage counters; cannot exceed effective defense.
  3. `DRAW_CARD`: Transfers card from `DeckZone` to `HandZone`; triggers fatigue if deck empty.
  4. `DISCARD_CARD`: Transfers card from `HandZone` to `GraveyardZone`.
  5. `DEPLOY_SCRATCH`: Deploys token asset `neutral_scratch_company` to Reserves.
  6. `EXHAUST_ASSET`: Sets asset exhaustion state to `true`.
  7. `READY_ASSET`: Sets asset exhaustion state to `false`.
  8. `BUFF_STATS`: Creates temporary stat modifier in Layer 5 with duration tracker.
  9. `DEBUFF_STATS`: Creates temporary stat reduction in Layer 5.
  10. `DESTROY_ASSET`: Transitions asset directly to Graveyard regardless of damage counters.
  11. `MOVE_ZONE`: Transitions asset between Reserves and Main Theater ([`FR-ZON-002`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L407)).
  12. `ADD_REQUISITION`: Adds immediate Requisition currency to player pool.
  13. `ADD_TACTICS`: Adds immediate Tactics currency to player pool.
  14. `GAIN_KEYWORD`: Adds temporary keyword grant to entity in Layer 3 or 5.

#### FR-ENG-022: Scratch Token Deployment & Pool Constraints
- **Functional Contract**: When resolving `DEPLOY_SCRATCH`, the engine MUST instantiate a new `AssetInstance` referencing `neutral_scratch_company` ([`FR-SET-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L34)):
  - Placed directly into the controlling player's `ReservesZone`.
  - Enters the battlefield unexhausted unless modified by `Deploys Exhausted` ([`FR-KWD-002l`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L758)).
  - Token instances leaving the battlefield (destruction, bounce) are deleted permanently and never enter the graveyard deck.

#### FR-ENG-023: Continuous Aura Grant & Revocation Engine
- **Functional Contract**: Continuous auras (`CONTINUOUS_AURA`, [`FR-ABL-004`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/02_Card_Data_Pipeline_FRD.md#L217), [`FR-TXT-004`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/02_Card_Data_Pipeline_FRD.md#L457)) MUST continuously apply while the host asset is on the battlefield and immediately cease when the host leaves the battlefield.
- **Invariants**: Moving zones or changing control immediately updates aura beneficiaries without requiring manual event triggers.

#### FR-ENG-024: Maneuver Card Spell Lifecycle & Graveyard Transition
- **Functional Contract**: When a maneuver card is played from hand:
  1. Hand deductions and costs are applied.
  2. The card is placed on the Resolution Stack as `FRAME_MANEUVER`.
  3. Upon reaching the top of the stack, its effects execute.
  4. Immediately upon completion, the card instance is moved to `GraveyardZone`.

---

## DOMAIN 5: 6-LAYER ATTRIBUTE RECALCULATION PIPELINE

```text
  +-----------------------------------------------------------------------------+
  |              6-LAYER CONTINUOUS ATTRIBUTE CALCULATION PIPELINE              |
  |                                                                             |
  |  Layer 1: Base Blueprint Stats (Immutable Power / Defense from CSV)         |
  |                                |                                            |
  |                                v                                            |
  |  Layer 2: Blueprint Modifiers (Static set-wide rules adjustments)           |
  |                                |                                            |
  |                                v                                            |
  |  Layer 3: Instance Modifiers (Permanent attachments / intrinsic keywords)   |
  |                                |                                            |
  |                                v                                            |
  |  Layer 4: Continuous Auras (Dynamic passive field buffs: "Allied units have")|
  |                                |                                            |
  |                                v                                            |
  |  Layer 5: Temporary Modifiers ("Gets +2/+0 until end of turn")              |
  |                                |                                            |
  |                                v                                            |
  |  Layer 6: Counters & Effective Clamping (Floor at 0; Effective Stats)       |
  +-----------------------------------------------------------------------------+
```

### Section 5.1: Layer Evaluation & Dirty Caching

#### FR-ENG-025: Deterministic 6-Layer Attribute Pipeline Architecture
- **Functional Contract**: The engine MUST calculate all dynamic entity attributes (effective power, effective defense, keywords, armor) using the strict 6-layer pipeline codified in [`FR-LAY-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L201) – [`FR-LAY-003`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L247).
- **Layer Precedence**:
  - **Layer 1 (Base Stats)**: Raw values from `CardBlueprint` ([`FR-DAT-005`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/02_Card_Data_Pipeline_FRD.md#L107)).
  - **Layer 2 (Blueprint Modifiers)**: Global modifications modifying card definition templates.
  - **Layer 3 (Instance Modifiers)**: Card-specific permanent modifiers, intrinsic keywords.
  - **Layer 4 (Continuous Auras)**: Dynamic field buffs evaluated in timestamp order.
  - **Layer 5 (Temporary Effects)**: Turn-duration stat buffs/debuffs with expiration timestamps.
  - **Layer 6 (Counters & Final Clamping)**: `+1/+1` counters, discrete tokens, and non-negative clamping.

#### FR-ENG-026: Dirty-Flag Propagation & Topological Dependency Caching
- **Functional Contract**: To maintain high simulation throughput, the engine MUST implement dirty-flag caching (`isDirty` bitmask):
  - Modifying an entity, adding/removing an aura host, or transitioning zones sets `isDirty = true`.
  - Attribute queries on clean entities return cached effective stats in $O(1)$ time without re-evaluating the layer stack.
  - Recalculation executes lazily upon next attribute query or immediately before combat damage calculation.

#### FR-ENG-027: Layer Recalculation Ordering & Mutator Precedence
- **Functional Contract**: Modifiers within the same layer MUST evaluate in strict chronological **Timestamp Order** (the order in which the modifier or aura source entered the active zone).
- **Stat Invariants**:
  - Modifiers modifying base stats evaluate before multiplicative modifiers.
  - Keyword grants in Layer 3/4 grant full keyword functionality immediately upon layer resolution.

#### FR-ENG-028: Effective Stat Clamping & Non-Negative Invariants
- **Functional Contract**: At the conclusion of Layer 6:
  - Effective power MUST clamp to a minimum of 0 (`EffectivePower = max(0, calculatedPower)`). Negative power values are prohibited.
  - Effective defense MUST clamp to a minimum of 0.
  - Damage counters accumulate on the asset; an asset is destroyed when `AccumulatedDamage >= EffectiveDefense`.

#### FR-ENG-029: Round-Duration Cleanup Sub-Phase Execution
- **Functional Contract**: During Sub-Phase 4c of the End Phase ([`FR-LAY-003`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L247)), the engine MUST iterate all active entities and purge temporary Layer 5 modifiers whose duration criteria are met:
  - *"Until the end of turn"*: Purged at the end of the current player's turn.
  - *"Until the end of your opponent's turn"*: Purged at the end of the recipient's opponent's turn.
  - Flagging `isDirty = true` on affected entities for immediate layer recalculation.

---

## DOMAIN 6: COMBAT RESOLUTION & RETALIATION ENGINE

```text
  +-----------------------------------------------------------------------------+
  |                        3-STEP COMBAT CYCLE (FR-CBT-002)                     |
  |                                                                             |
  |  Step 1: Declaration & Exhaustion Lock                                      |
  |  - Attacker in Main Theater targets enemy asset or HQ                       |
  |  - Attacker exhausts physically                                             |
  |                                |                                            |
  |                                v                                            |
  |  Step 2: Reaction & Trigger Window                                          |
  |  - Emit EVENT_ATTACK_DECLARED                                               |
  |  - Queue triggers on Stack via APNAP (FR-ENG-014)                           |
  |  - Resolve priority passes                                                  |
  |                                |                                            |
  |                                v                                            |
  |  Step 3: Simultaneous Combat Damage Resolution & Retaliation                |
  |  - Evaluate Stalwart, Resistant, and Armor absorption                       |
  |  - Apply simultaneous damage ticks                                          |
  |  - Execute state-based death sweep batch                                    |
  +-----------------------------------------------------------------------------+
```

### Section 6.1: Deterministic Combat Resolution

#### FR-ENG-030: Combat Cycle Orchestration & Zone Restrictions
- **Functional Contract**: Combat declarations MUST strictly adhere to the spatial battlefield topology defined in [`FR-CBT-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L282) and [`FR-ZON-002`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L407):
  - Only units positioned in the **Main Theater** may declare attacks.
  - Units in the Reserves Zone are prohibited from attacking.
  - Valid targets: Enemy units/structures in the Main Theater, or enemy Headquarters in the HQ Zone ([`FR-ZON-003`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L420)).

#### FR-ENG-031: Step 1: Declaration, Legality Gates & Exhaustion Lock
- **Functional Contract**: Upon receiving `DeclareAttackCommand`, the engine MUST validate all preconditions within an atomic transaction:
  1. Active player has priority and stack is clean (`Stack.IsEmpty() == true`, [`FR-QUE-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L110)).
  2. Attacking unit is currently **unexhausted**.
  3. Attacker is not constrained by summoning sickness (unless possessing `Aggressive N`, [`FR-KWD-002f`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L752)).
  4. Target is legal and unshielded by `Guardian` ([`FR-KWD-002j`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L756)).
- **Execution**: The attacking unit is set to `exhausted = true` immediately upon declaration.

#### FR-ENG-032: Step 2: Reaction Window & Combat Interruption Frames
- **Functional Contract**: Prior to damage calculation, the engine MUST emit `EVENT_ATTACK_DECLARED` and open a priority reaction window ([`FR-CBT-002`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L295)):
  - Triggers responding to attack declaration are queued onto the Resolution Stack via APNAP ([`FR-ENG-014`](#fr-eng-014-apnap-trigger-dispatching--simultaneous-event-ordering)).
  - The stack resolves to completion.
  - **Target Survival Check**: If the target asset is destroyed or removed from the battlefield during Step 2, combat terminates without dealing damage.

#### FR-ENG-033: Step 3: Simultaneous Combat Damage & Retaliation Calculation
- **Functional Contract**: When Step 2 concludes with both attacker and defender surviving on the battlefield, the engine MUST resolve combat damage **simultaneously** per [`FR-CBT-002`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L295):
  - Attacker calculates incoming damage to target = `Attacker.EffectivePower`.
  - If target is a unit (not structure/HQ), target calculates retaliation damage = `Target.EffectivePower`.
  - Both damage payloads apply concurrently within the same frame tick.

#### FR-ENG-034: Damage Absorption Hierarchy (Armor, Stalwart, Resistant)
- **Functional Contract**: When applying incoming combat damage to an asset, the engine MUST evaluate defensive keywords in the strict canonical order codified in [`FR-KWD-002c`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L14) and [`FR-REP-004`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L14):
  1. **Armor N**: Armor points absorb incoming damage on a 1-to-1 basis until Armor reaches 0.
  2. **Stalwart**: If remaining combat damage $> 1$, damage is capped strictly to 1.
  3. **Resistant N**: Flat damage reduction subtracts $N$ from remaining damage (`damage = max(0, damage - N)`).
  4. **Base Defense**: Remaining net damage is added to `AccumulatedDamage`.

#### FR-ENG-035: State-Based Death Sweep & Graveyard Transition Batching
- **Functional Contract**: Immediately following damage application, the engine MUST execute a state-based death check sweep ([`FR-ENT-003`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L17)):
  - Any asset with `AccumulatedDamage >= EffectiveDefense` is marked as destroyed.
  - All marked assets are transitioned to `GraveyardZone` simultaneously in a single atomic batch.
  - `EVENT_ENTITY_DESTROYED` events are emitted, queuing death triggers onto the stack via APNAP.

---

## DOMAIN 7: HEADLESS TEST HARNESS, REPLAY & SIMULATION SUITE

```text
  +-----------------------------------------------------------------------------+
  |                     HEADLESS TEST & SIMULATION SUITE                        |
  |                                                                             |
  |  +--------------------+   +---------------------+   +--------------------+  |
  |  |   CLI Test Runner  |   | Scripted Scenario   |   | Deterministic .rec |  |
  |  |  (engine-sim tool) |   |    DSL (.yaml/.json)|   |  Replay Validator  |  |
  |  +---------+----------+   +----------+----------+   +---------+----------+  |
  |            |                         |                        |             |
  |            +-------------------------+------------------------+             |
  |                                      |                                      |
  |                                      v                                      |
  |                   +--------------------------------------+                  |
  |                   | Combinatorial Fuzzing & Stress Suite |                  |
  |                   | (10,000 matches in <5s, State Hashes)|                  |
  |                   +--------------------------------------+                  |
  +-----------------------------------------------------------------------------+
```

### Section 7.1: Simulation, Replay & Benchmarking Tools

#### FR-ENG-036: Headless Engine CLI & Test Runner Architecture
- **Functional Contract**: The module MUST provide a standalone CLI tool (`cmd/engine-sim`) enabling automated headless match execution without graphical dependencies.
- **Flags**:
  - `--catalog <path>`: Path to `blueprints.json`.
  - `--p1-deck <path>`, `--p2-deck <path>`: Paths to test deck manifests.
  - `--seed <int64>`: Deterministic PRNG seed.
  - `--scenario <path>`: Pre-scripted state scenario file.
  - `--fuzz <n>`: Run $N$ automated randomized bot simulations.

#### FR-ENG-037: Scripted Match Scenario DSL & Assertion Harness
- **Functional Contract**: The engine test harness MUST support JSON/YAML scenario fixtures defining starting board states and scripted player action sequences:
  - Explicit starting hands, zones, damage counters, and resource pools.
  - Step-by-step action sequence (deploy, attack, pass).
  - Post-condition assertions (`assert_zone`, `assert_stat`, `assert_stack_empty`).
- **Assertion Failure**: Any assertion failure MUST halt execution and output a detailed diff of expected versus actual state.

#### FR-ENG-038: Deterministic .rec Replay Recorder & Validator
- **Functional Contract**: The engine MUST record all committed transactions, inputs, and seeds into the binary/JSON `.rec` match replay format specified in [`FR-SPC-002`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/04_Client_Engine_Protocol_Networking_FRD.md#L735).
- **Validation Mode**: Given a `.rec` file, the engine MUST re-simulate the entire match from the initial seed and assert that every intermediate `StateChecksum` matches bit-for-bit.

#### FR-ENG-039: Combinatorial Monkey/Fuzz Testing Runner
- **Functional Contract**: The CLI suite MUST include an automated monkey testing runner executing randomized, legal player actions across random seeds:
  - Detects infinite trigger loops, stack overflows, and negative stat violations.
  - Asserts that every random match terminates cleanly within 100 turns.
  - Emits reproducible failure artifacts (`repro_seed_<seed>.json`) upon error.

#### FR-ENG-040: Simulation Benchmark Suite & Performance Thresholds
- **Functional Contract**: The engine MUST include automated Go benchmark tests (`go test -bench=. ./pkg/engine/...`) asserting the following minimum performance gates on modern x86_64/ARM64 CPUs:
  - **Turn Execution**: $\ge 10,000$ simulated player turns per second per CPU core.
  - **Layer Recalculation**: $\le 2$ microseconds per 6-layer entity evaluation.
  - **Memory Footprint**: $\le 4$ MB RSS memory per active headless match instance.

---

## DOMAIN 8: ERROR TAXONOMY, RESILIENCE & DIAGNOSTICS

### Section 8.1: Error Code Registry
| Error Code | Severity | Trigger Condition | System Action | Client Presentation |
| :--- | :--- | :--- | :--- | :--- |
| `ERR_ENG_INVALID_SEED` | Fatal | Master seed provided is 0 or malformed. | Abort match initialization. | "Initialization failed: Invalid match seed." |
| `ERR_ENG_DIRTY_STACK_ACTION` | Error | Manual action attempted while stack non-empty (`FR-QUE-001`). | Abort transaction; rollback delta. | "Action prohibited: Cannot perform manual action while stack is resolving." |
| `ERR_ENG_INSUFFICIENT_RESOURCE` | Error | Player lacks Requisition or Tactics to pay cost. | Abort transaction; rollback delta. | "Cannot play card: Insufficient resources." |
| `ERR_ENG_ASSET_EXHAUSTED` | Error | Attempted to attack or activate ability with exhausted asset. | Abort transaction; rollback delta. | "Asset is exhausted and cannot act." |
| `ERR_ENG_TARGET_INVALID` | Warning | Target became illegal or moved zones prior to stack pop. | Fizzle stack frame; log diagnostic. | "Spell fizzled: Target is no longer valid." |
| `ERR_ENG_ZONE_RESTRICTION` | Error | Action violates spatial zone rules (e.g. attacking from Reserves). | Abort transaction; rollback delta. | "Action prohibited: Unit is not in the Main Theater." |
| `ERR_ENG_DESYNC_DETECTED` | Fatal | StateChecksum mismatch between client and server (`FR-PAC-004`). | Terminate match; emit replay dump. | "Match terminated due to desynchronization." |
| `ERR_ENG_INFINITE_LOOP` | Fatal | Stack execution depth exceeds 100 frames. | Force-clear stack; emit fatal log. | "Match error: Infinite trigger loop detected." |

### Section 8.2: Parity & State Invariants
#### FR-ENG-041: System Invariants & Diagnostic Assertion Protocol
- **Invariant 1 (Non-Negative Stats)**: Effective power and effective defense MUST never evaluate to negative integers (`EffectiveStat >= 0`).
- **Invariant 2 (Zone Conservation)**: A `CardInstance` or `AssetInstance` MUST reside in exactly one zone at any given moment. Moving an entity MUST atomically remove it from the source zone before insertion into the destination zone.
- **Invariant 3 (Clean Stack at Turn Handover)**: Turn transition from active to non-active player MUST assert `ResolutionStack.IsEmpty() == true`.

---

## MASTER CROSS-REFERENCE REGISTER

| Requirement ID | Scope Summary | Interoperability Target |
| :--- | :--- | :--- |
| `FR-ENG-001` | Pure Go engine modularity & zero graphics dependency | [`FR-ENT-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L29), [`FR-SES-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/04_Client_Engine_Protocol_Networking_FRD.md#L68) |
| `FR-ENG-002` | Isolated multi-stream PRNG generator | [`FR-SET-002`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L45), [`FR-SET-003`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L54) |
| `FR-ENG-003` | Deterministic seed derivation with cryptographic salts | [`FR-SET-003`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L54), [`FR-SPC-002`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/04_Client_Engine_Protocol_Networking_FRD.md#L735) |
| `FR-ENG-004` | Pure discrete integer arithmetic | [`FR-LAY-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L201), [`FR-CBT-002`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L295) |
| `FR-ENG-005` | Monotonic tick counter & state checksum generation | [`FR-PAC-004`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/04_Client_Engine_Protocol_Networking_FRD.md#L450) |
| `FR-ENG-006` | Transactional event-emitting match state model | [`FR-ENT-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L29), [`FR-ZON-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L398) |
| `FR-ENG-007` | Match setup, zone initialization & mulligan execution | [`FR-SET-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L29), [`FR-SET-002`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L45), [`FR-SES-004`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/04_Client_Engine_Protocol_Networking_FRD.md#L94) |
| `FR-ENG-008` | Transaction checkpoint & delta-rollback engine | [`FR-QUE-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L110) |
| `FR-ENG-009` | Turn sub-phase transitions & cleanup execution | [`FR-LAY-003`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L247), [`FR-SET-004`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L57) |
| `FR-ENG-010` | Clean-stack constraint enforcement on manual actions | [`FR-QUE-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L110), [`FR-TRN-003a`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L16) |
| `FR-ENG-011` | Turn-clock timeout & Reserve Time Bank exhaustion | [`FR-TRN-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/04_Client_Engine_Protocol_Networking_FRD.md#L508), [`FR-SES-005`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/04_Client_Engine_Protocol_Networking_FRD.md#L100) |
| `FR-ENG-012` | LIFO resolution stack & frame taxonomy | [`FR-QUE-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L110) |
| `FR-ENG-013` | Atomic cost validation, reservation & deduction | [`FR-SET-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L29), [`FR-ABL-007`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/02_Card_Data_Pipeline_FRD.md#L273) |
| `FR-ENG-014` | APNAP simultaneous trigger dispatching order | [`FR-QUE-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L110) |
| `FR-ENG-015` | Target verification immediately prior to stack pop | [`FR-TGT-102`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L16), [`FR-ABL-008`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/02_Card_Data_Pipeline_FRD.md#L292) |
| `FR-ENG-016` | Reactive replacement watcher interceptor pipeline | [`FR-REP-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L14), [`FR-REP-004`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L14) |
| `FR-ENG-017` | Choice prompt modal yield & execution suspension | [`FR-MSG-005`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/04_Client_Engine_Protocol_Networking_FRD.md#L400), [`FR-MOD-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L18) |
| `FR-ENG-018` | Polymorphic ability AST type-switch visitor | [`FR-ABL-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/02_Card_Data_Pipeline_FRD.md#L154) – [`FR-ABL-005`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/02_Card_Data_Pipeline_FRD.md#L236) |
| `FR-ENG-019` | Dynamic filter evaluator against effective attributes | [`FR-ABL-008`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/02_Card_Data_Pipeline_FRD.md#L292) – [`FR-ABL-011`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/02_Card_Data_Pipeline_FRD.md#L368) |
| `FR-ENG-020` | ValueToken expression resolver & condition gate | [`FR-ABL-006`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/02_Card_Data_Pipeline_FRD.md#L254), [`FR-ABL-012`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/02_Card_Data_Pipeline_FRD.md#L391) |
| `FR-ENG-021` | The 14 atomic primitives execution contracts | [`FR-PRM-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L525) |
| `FR-ENG-022` | Token Scratch Company instantiation & pool rules | [`FR-SET-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L34), [`FR-KWD-002l`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L758) |
| `FR-ENG-023` | Continuous aura registration & revocation engine | [`FR-ABL-004`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/02_Card_Data_Pipeline_FRD.md#L217), [`FR-TXT-004`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/02_Card_Data_Pipeline_FRD.md#L457) |
| `FR-ENG-024` | Maneuver card resolution & graveyard transition | [`FR-QUE-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L110) |
| `FR-ENG-025` | 6-layer attribute calculation pipeline execution | [`FR-LAY-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L201) – [`FR-LAY-003`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L247) |
| `FR-ENG-026` | Dirty-flag caching & topological dependency invalidation | [`FR-LAY-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L201) |
| `FR-ENG-027` | Timestamp precedence within attribute calculation layers | [`FR-LAY-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L201) |
| `FR-ENG-028` | Effective attribute clamping and non-negative invariants | [`FR-LAY-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L201) |
| `FR-ENG-029` | Round-duration temporary modifier cleanup | [`FR-LAY-003`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L247) |
| `FR-ENG-030` | Spatial zone combat restrictions (Main Theater only) | [`FR-CBT-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L282), [`FR-ZON-002`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L407) |
| `FR-ENG-031` | Step 1 attack declaration, validation & exhaustion lock | [`FR-CBT-002`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L295), [`FR-QUE-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L110) |
| `FR-ENG-032` | Step 2 reaction window & trigger queue | [`FR-CBT-002`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L295) |
| `FR-ENG-033` | Step 3 simultaneous combat damage & retaliation | [`FR-CBT-002`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L295) |
| `FR-ENG-034` | Canonical damage absorption hierarchy (Armor/Stalwart/Resistant) | [`FR-KWD-002c`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L14), [`FR-REP-004`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L14) |
| `FR-ENG-035` | State-based death check sweep & graveyard transition | [`FR-ENT-003`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L17) |
| `FR-ENG-036` | Headless engine CLI test runner architecture | [`FR-SES-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/04_Client_Engine_Protocol_Networking_FRD.md#L68) |
| `FR-ENG-037` | Scripted scenario fixture DSL & assertion framework | [`FR-PAC-004`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/04_Client_Engine_Protocol_Networking_FRD.md#L450) |
| `FR-ENG-038` | Deterministic .rec match replay recording & validation | [`FR-SPC-002`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/04_Client_Engine_Protocol_Networking_FRD.md#L735) |
| `FR-ENG-039` | Combinatorial monkey/fuzz testing orchestrator | [`FR-PAC-004`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/04_Client_Engine_Protocol_Networking_FRD.md#L450) |
| `FR-ENG-040` | Engine performance benchmarks & throughput gates | [`FR-ENG-001`](#fr-eng-001-engine-modularity--zero-engine-dependency-invariant) |
| `FR-ENG-041` | System invariants & diagnostic assertions | [`FR-ENT-003`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L17), [`FR-QUE-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L110) |
