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

#### FR-ENG-002: Isolated Multi-Stream PRNG Architecture & Seed Derivation
- **Functional Contract**: The engine MUST isolate pseudo-random number generation across independent streams using the PCG32 or SplitMix64 algorithmic generator to prevent seed consumption side-effects from altering unrelated match events.
- **Streams & Salt Derivation**: Child stream seeds are deterministically derived from the master 64-bit seed (`MasterSeed`) using fixed, non-colliding 64-bit cryptographic salt constants:
  1. `StreamShuffle`: Dedicated exclusively to initial deck shuffling ([`FR-SET-002`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L45)) and mid-match deck shuffles ([`FR-PRM-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L304)). Derived via `SeedShuffle = MasterSeed ^ 0x9E3779B97F4A7C15`.
  2. `StreamTargeting`: Dedicated exclusively to resolving non-deterministic card targeting (e.g. *"deal 2 damage to a random enemy unit"*). Derived via `SeedTargeting = MasterSeed ^ 0xBF58476D1CE4E5B9`.
  3. `StreamMatchDecisions`: Dedicated to pre-game coin toss ([`FR-SET-003`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L54)) and tie-breaking procedures. Derived via `SeedMatchDecisions = MasterSeed ^ 0x94D049BB133111EB`.
- **Invariants**: Shuffling a deck MUST NOT advance or alter the sequence of random target selections during combat. Supplying a seed of `0` MUST be rejected with `ERR_INVALID_COMMAND_ARGUMENTS` or initialized from a non-zero cryptographic entropy source.

#### FR-ENG-003: Subsumed by FR-ENG-002
- *Deterministic seed derivation and cryptographic salt constants are formally subsumed into [`FR-ENG-002`](#fr-eng-002-isolated-multi-stream-prng-architecture--seed-derivation).*

#### FR-ENG-004: Pure Discrete Integer Arithmetic & Rounding Invariants
- **Functional Contract**: All numeric calculations within the engine MUST execute using discrete 32-bit signed integers (`int32`).
- **Rules**:
  - Division operations resulting in fractional quantities MUST perform deterministic floor integer truncation (`a / b`).
  - Damage calculation, flat resistance, armor deductions, and cost reductions MUST NEVER underflow below 0 unless an explicit signed delta contract is invoked.
  - Floating-point arithmetic (`float32`, `float64`) is strictly prohibited in rules evaluation.

#### FR-ENG-005: Frame Tick Invariants & Rolling State Hash Generation
- **Functional Contract**: The engine MUST maintain a strictly monotonic 64-bit tick counter (`TickNumber`) incremented on every committed transaction and phase transition.
- **State Checksum**: At the conclusion of every turn sub-phase and stack frame resolution, the engine MUST calculate a rolling 64-bit FNV-1a state checksum (`StateChecksum`) across all public entity states, zone contents, and resource pools in strict conformance with [`FR-PAC-004`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/04_Client_Engine_Protocol_Networking_FRD.md#L690).
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
  - `CurrentPhase`: Current match phase (`PHASE_READY`, `PHASE_DRAW`, `PHASE_MAIN`, `PHASE_END`, `STATE_CONCLUDED`).
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
  5. Determines first player decision-maker via `StreamMatchDecisions` ([`FR-SET-003`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L54)).

#### FR-ENG-008: Transaction Checkpoint, Delta-Rollback & Commit Lifecycle
- **Functional Contract**: Every player command, stack resolution step, and phase transition MUST execute inside an atomic transaction:
  - `tx := match.BeginTransaction()`
  - Before modifying state, the transaction records a lightweight delta-snapshot of touched entities, zones, resource balances, and internal monotonic sequence counters of all active PRNG streams ([`FR-ENG-002`](#fr-eng-002-isolated-multi-stream-prng-architecture--seed-derivation)).
  - **On Failure / Illegal Command**: `tx.Rollback()` restores touched entities, resource balances, and PRNG stream counters to their exact pre-transaction state, purges `EventBuffer`, and returns a structured error code ([`FR-ERR-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/04_Client_Engine_Protocol_Networking_FRD.md#L745)). The active match state and PRNG sequence determinism remain completely uncorrupted.
  - **On Success**: `tx.Commit()` applies state mutations permanently, calculates the rolling state hash ([`FR-ENG-005`](#fr-eng-005-frame-tick-invariants--rolling-state-hash-generation)), flushes `EventBuffer` to the outgoing network/replay channel, and discards the delta-snapshot.

#### FR-ENG-009: Deterministic Turn Sequencing & Sub-Phase Transitions
- **Functional Contract**: The engine MUST execute turn phases and associated sub-phase behaviors in strict sequential order conforming to [`FR-TRN-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L236):
  - **Phase 1: Ready Phase**: Unexhausts all controlled assets in Reserves, Main Theater, and HQ Zone; grants base Requisition income and resets unspent Tactics.
  - **Phase 2: Draw Phase**: Active player draws 1 card from `DeckZone` into `HandZone`. Skipped on Turn 1 for the player taking the first turn ([`FR-SET-004`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L57)).
  - **Phase 3: Main Phase**: Clean-stack action priority window for player-initiated deployments, maneuvers, inherent actions, and combat declarations.
  - **Phase 4: End Phase**:
    - *Sub-Phase 4a (Event Emission & Trigger Enqueue)*: Emits `EVENT_END_PHASE`; enqueues end-of-turn triggers via APNAP ([`FR-TRG-003`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L396)).
    - *Sub-Phase 4b (Trigger Cascade Resolution)*: Resolution stack auto-resolves to completion.
    - *Sub-Phase 4c (Resource Depletion & Expiration)*: Unspent Requisition and Tactics pools reset to 0 ([`FR-RES-002`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L225)); temporary Layer 5 modifiers expiring at end-of-turn expire ([`FR-LAY-003`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L439)).
    - *Sub-Phase 4d (Turn Handover)*: Increment `TurnNumber` and toggle `ActivePlayer`. Universal uncapped hand capacity is preserved per [`FR-ZON-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L162) with zero hand-size discard checks.

#### FR-ENG-010: Clean-Stack Manual Action Constraint & Inherent Actions
- **Functional Contract**: Manual player actions (playing a maneuver card, deploying a unit/structure, declaring combat, or executing inherent actions) MUST assert that `ResolutionStack.IsEmpty() == true` per [`FR-QUE-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L110) and [`FR-INH-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L281):
  - **Inherent March**: Enforces unit movement between Reserves and Main Theater per [`FR-INH-002`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L264), arriving `EXHAUSTED` (emitting `EVENT_MARCH` and `EVENT_EXHAUST`).
  - **Inherent Resource Operations**: Enforces resource generation per [`FR-RES-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L221) and instantaneous conversion per [`FR-RES-003`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L228) (spending 2 Requisition adds 1 Tactics without placing items on the stack).
- **Exceptions**: Reactive triggered abilities, in-resolution free deployments ([`FR-TRN-003a`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L296)), and replacement watchers execute without clean-stack constraints.
- **Violation**: If a player attempts a manual action while the stack contains pending items, the engine MUST abort the transaction with `ERR_STACK_NOT_EMPTY`.

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
- **Failure**: If any cost component cannot be paid in full, the transaction aborts with `ERR_INSUFFICIENT_RESOURCES` or `ERR_ASSET_EXHAUSTED`. Zero partial deduction permitted.

#### FR-ENG-014: APNAP Trigger Dispatching & Simultaneous Event Ordering
- **Functional Contract**: When an in-game event triggers multiple abilities simultaneously, the engine MUST queue them onto the stack using the **APNAP (Active Player, Non-Active Player)** convention in conformance with [`FR-TRG-003`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L396):
  1. All triggers controlled by the **Active Player** are pushed onto the stack first.
  2. All triggers controlled by the **Non-Active Player** are pushed onto the stack second on top of the Active Player's triggers.
- **Invariants**: Because the stack is LIFO, the Non-Active Player's triggers resolve first, preserving deterministic priority without race conditions. Where a player controls multiple simultaneous triggers, allied triggers are ordered by their controller (falling back to deterministic blueprint ID sorting for bot/automated simulation).

#### FR-ENG-015: Dynamic Target Verification & Invalidation Invariants
- **Functional Contract**: Targets selected upon action declaration MUST be re-verified for legality immediately prior to execution when the stack frame pops:
  - Target entity must still exist in its declared zone ([`FR-TGT-102`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L16)).
  - Target must still satisfy all `FilterDef` constraints ([`FR-ABL-008`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/02_Card_Data_Pipeline_FRD.md#L292)).
- **Partial Invalidation**: If an ability targets multiple entities and at least one target remains legal, the ability executes against the remaining legal targets.
- **Complete Invalidation**: If all declared targets have become invalid, the stack frame **fizzles**: effect execution is skipped, the card is sent to the discard pile, and `ERR_INVALID_TARGET` is logged.

#### FR-ENG-016: Reactive Replacement Watcher Interceptor Pipeline
- **Functional Contract**: Replacement effects ([`FR-REP-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L14) – [`FR-REP-004`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L14)) MUST intercept and modify event payloads *before* they are applied to state or pushed to the stack.
- **Application Order**:
  1. Self-replacement effects (effects on the entity taking damage or moving zones).
  2. Controller's replacement effects.
  3. Opponent's replacement effects.
- **Recursion Guard**: A replacement effect MUST NOT apply more than once to the same root event.

#### FR-ENG-017: Choice Prompt Yield & Resumption Protocol
- **Functional Contract**: When an effect requires runtime player selection (scry/look-at-top-N cards, modal choice, distributed damage allocation; [`FR-MSG-005`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/04_Client_Engine_Protocol_Networking_FRD.md#L400), [`FR-MOD-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L18)), the engine MUST suspend execution of the active stack frame:
  1. Emits `PlayerChoicePrompt` payload to the target player client with 3,000 ms grace window ([`DA-NET-004`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/04_Client_Engine_Protocol_Networking_FRD.md#L14)).
  2. Sets engine state to `STATE_AWAITING_CHOICE`.
  3. Freezes turn clock and starts the action clock (15s for active player, 10s for inactive player per [`FR-TRG-002`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L394)).
  4. Resumes stack execution immediately upon receiving a valid `ChoiceResponse` matching prompt UUID, or applies deterministic fallbacks upon expiration ([`FR-MOD-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L361)).

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

#### FR-ENG-021: Atomic Effect Primitives Execution Engine
- **Functional Contract**: The engine MUST implement discrete execution handlers dispatching 1:1 to the canonical catalog of 25 atomic effect primitives codified in [`FR-PRM-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L304).
- **Execution Model**: The polymorphic AST type-switch interpreter ([`FR-ENG-018`](#fr-eng-018-polymorphic-ability-ast-interpreter--type-switch-visitor)) maps each canonical primitive identifier (`DEAL_DAMAGE`, `HEAL`, `ADD_ARMOR`, `MODIFY_STAT`, `MOVE_ZONE`, `RETURN_TO_HAND`, `KILL`, `DRAW_CARD`, `DISCARD_CARD`, `REVEAL_HAND`, `SCRATCH`, `REGROUP`, `DEPLOY_FREE`, `SEARCH_DECK`, `LOOK_TOP_CARDS`, `TUCK_UNDER`, `GAIN_CONTROL`, `MORPH_SCRATCH`, `GENERATE_RESOURCE`, `RECYCLE_DISCARD_TO_DECK`, `REDIRECT_DAMAGE_TARGET`, `PUT_CARD_ON_DECK`, `EXHAUST`, `GRANT_SUBTYPE`, `GRANT_KEYWORD`) directly to its dedicated runtime Go execution contract within the active transaction context.
- **Safety Invariant**: Any unhandled or unrecognized primitive encountered within an ability AST payload MUST abort the transaction with structured error code `ERR_INVALID_COMMAND_ARGUMENTS`.

#### FR-ENG-022: Double-Blind Scratch Company Lifecycle Engine
- **Functional Contract**: When executing `SCRATCH` ([`FR-PRM-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L304), [`FR-KWD-002m`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L759)) or the Headquarters inherent action ([`FR-INH-004`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L283)), the engine MUST execute a direct deck-to-battlefield transition conforming to [`FR-ZON-005`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L196) and [`FR-SEC-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/04_Client_Engine_Protocol_Networking_FRD.md#L164):
  1. **Deployment (`DECK` &rarr; `RESERVES`)**: Pops the top card of the controlling player's `DeckZone` and moves it directly into `ReservesZone` wrapped in an `AssetInstance` with `isFaceDown = true`, effective base stats `1/1/1/1 Infantry Scratch Company`, and an opaque proxy ID (`proxy_scratch_*`). Does NOT emit `EVENT_DRAW_CARD`.
  2. **Double-Blind Invariant**: The server never transmits the underlying `cardInstanceId` or `blueprintId` to either client while on the battlefield ([`FR-SEC-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/04_Client_Engine_Protocol_Networking_FRD.md#L164)). Neither player may inspect the underlying card face.
  3. **Empty Deck Protection**: Executing Scratch on a 0-card deck is a safe no-op: creates 0 units and does NOT trigger empty deck loss ([`FR-INH-004`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L283)).
  4. **Zone Departures & Reveals**: Conforming to [`FR-ZON-005`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L196) and [`FR-SEC-002`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/04_Client_Engine_Protocol_Networking_FRD.md#L171):
     - *Departure to Public Zone (`DISCARD`)*: Asset deallocates; underlying card moves face-up to Discard, revealing its identity to both players.
     - *Departure to Private Zone (`HAND`, `DECK`)*: Asset deallocates; underlying card reverts to base printed card in owner's private zone ([`FR-ENT-103`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L211), [`FR-ZON-004`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L186)).
     - *Tucked (`ATTACHED`)*: Under host asset via `TUCK_UNDER`, remains a double-blind face-down sub-entity ([`FR-ENT-102`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L203)).

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
- **Functional Contract**: The engine MUST enforce continuous dynamic attribute evaluation across all battlefield entities in strict compliance with the 6-layer pipeline codified in [`FR-LAY-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L431) through [`FR-LAY-003`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L439):
  - **Layer 1 (Base Stats)**: Base Blueprint Values (or Scratch base stats `1/1/1/1` if `isFaceDown == true`).
  - **Layer 2 (Identity Grants)**: Type, Subtype & Identity Grants (`GRANT_SUBTYPE`, `MORPH_SCRATCH`).
  - **Layer 3 (Keyword Grants)**: Dynamic Keyword Grants & Suppressions (`GRANT_KEYWORD`).
  - **Layer 4 (Continuous Auras)**: Continuous static field buffs evaluated in timestamp order.
  - **Layer 5 (Temporary Modifiers)**: Dynamic turn modifiers & temporary buffs/debuffs (`MODIFY_STAT`).
  - **Layer 6 (Mathematical Clamping)**: Mathematical clamping (`EffectivePower = max(0, RawPower)`; Combat Power is the sole clamped stat per `FR-LAY-001`).
- **Timestamp & Clamping Precedence**: Modifiers within the same layer evaluate in chronological timestamp order. Base modifiers evaluate before multiplicative modifiers.

#### FR-ENG-026: Dirty-Flag Propagation & Topological Dependency Caching
- **Functional Contract**: To maintain high simulation throughput, the engine MUST implement dirty-flag caching (`isDirty` bitmask):
  - Modifying an entity, adding/removing an aura host, or transitioning zones sets `isDirty = true`.
  - Attribute queries on clean entities return cached effective stats in $O(1)$ time without re-evaluating the layer stack.
  - Recalculation executes lazily upon next attribute query or immediately before combat damage calculation.

#### FR-ENG-027: Subsumed by FR-ENG-025
- *Timestamp ordering within layers is formally subsumed into [`FR-ENG-025`](#fr-eng-025-deterministic-6-layer-attribute-pipeline-architecture).*

#### FR-ENG-028: Subsumed by FR-ENG-025 & FR-ENG-041
- *Effective attribute clamping at Layer 6 is formally subsumed into [`FR-ENG-025`](#fr-eng-025-deterministic-6-layer-attribute-pipeline-architecture) and codified as System Invariant 1 in [`FR-ENG-041`](#fr-eng-041-system-invariants--diagnostic-assertion-protocol).*

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
  - Valid Attackers: Units positioned in the **Main Theater** and Headquarters in the **HQ Zone** ([`FR-CAP-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L265)).
  - Prohibited Attackers: Units in the Reserves Zone and Structures.
  - Valid Targets: Opposing Assets in the Main Theater, or opposing Headquarters in the HQ Zone ([`FR-ZON-003`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L420)). Reserves Zone assets are strictly protected from attacks.

#### FR-ENG-031: Step 1: Declaration, Legality Gates & Exhaustion Lock
- **Functional Contract**: Upon receiving `DeclareAttackCommand`, the engine MUST validate all preconditions within an atomic transaction:
  1. Active player has priority and stack is clean (`Stack.IsEmpty() == true`, [`FR-QUE-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L110)).
  2. Attacking asset is currently **readied** (`exhausted == false`).
  3. Target is legal and satisfies Guardian priority ([`FR-KWD-002j`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L756)). Bypassing an eligible opposing Guardian in the Main Theater aborts the transaction with `ERR_GUARDIAN_PRIORITY_VIOLATION`.
- **Execution**: The attacking asset transitions immediately to `exhausted = true` upon declaration (`EVENT_EXHAUST`, `EVENT_ATTACK_DECLARE`).

> [!NOTE] System Invariant: Absence of Summoning Sickness
> Unlike other card games that require an asset to be controlled across a turn transition before acting (often termed "summoning sickness"), in Card Town any asset that deploys readied may declare attacks or exhaust on that same turn ([`FR-SET-004`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L57), [`FR-ZON-002`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L183)). Pacing is governed organically by the spatial battlefield economy: units deploy into Reserves, and marching into Main Theater arrives exhausted ([`FR-INH-002`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L264)).

#### FR-ENG-032: Step 2: Reaction Window & Combat Interruption Frames
- **Functional Contract**: Prior to damage calculation, the engine MUST emit `EVENT_ATTACK_DECLARED` and process reactive triggers ([`FR-CBT-002`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L295)):
  - Triggers responding to attack declaration are queued onto the Resolution Stack via APNAP ([`FR-ENG-014`](#fr-eng-014-apnap-trigger-dispatching--simultaneous-event-ordering)).
  - The stack auto-resolves to completion. Declaration initiates an atomic combat transaction locking out manual card plays; once Step 2 triggers resolve, the match advances immediately to Step 3 without an intervening priority window.
  - **Target Survival Check**: If the target asset is destroyed or removed from the battlefield during Step 2, combat terminates without dealing damage.

#### FR-ENG-033: Step 3: Simultaneous Combat Damage & Symmetric Retaliation
- **Functional Contract**: When Step 2 concludes with both attacker and defender surviving on the battlefield, the engine MUST resolve combat damage **simultaneously** per [`FR-CBT-002`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L295):
  - **Continuous Aura Snapshot Lock**: All continuous aura condition gates (including *Cornered* count checks; [`FR-LAY-003`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L441)) are snapshotted and locked at the entry of Step 3; all simultaneous damage and retaliation calculations evaluate against this locked state.
  - Attacker calculates incoming damage to target = `Attacker.EffectivePower`.
  - Retaliation damage = `Target.EffectivePower` (evaluated symmetrically across all defending targets with effective power > 0 per [`FR-CBT-002`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L517)).
  - Both damage payloads route through their respective 5-stage mitigation pipelines ([`FR-ENG-034`](#fr-eng-034-canonical-5-stage-damage-mitigation-pipeline)) and apply concurrently within the same frame tick.

#### FR-ENG-034: Canonical 5-Stage Damage Mitigation Pipeline
- **Functional Contract**: When applying incoming combat or effect damage to an asset, the engine MUST evaluate defensive keywords in the strict canonical order codified in [`FR-REP-004`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L455):
  1. **Stage 1 (Target Redirection)**: Evaluates uncommitted damage against redirection targets ([`FR-REP-004`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L455)).
  2. **Stage 2 (Flat Reductions)**: Subtracts `Resistant N` (`damage = max(0, damage - N)` for positive; negative Resistant amplifies damage: $0 - (-1) = 1$).
  3. **Stage 3 (Combat Damage Capping)**: If `isCombatDamage == true` and target has `Stalwart` ([`FR-KWD-002c`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L749)), combat damage is capped strictly to 1 (`min(damage, 1)`). Has zero effect on non-combat damage.
  4. **Stage 4 (Dynamic Armor)**: Dynamic Armor absorbs damage 1-to-1 from target's current Armor pool ([`FR-KWD-002k`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L757)), depleting the pool accordingly. When a target has both Stalwart and Armor, Stalwart caps incoming combat damage to 1 in Stage 3, and Stage 4 absorbs that 1 damage by depleting 1 point of Armor.
  5. **Stage 5 (Health Damage Commit)**: Remaining damage is added to `AccumulatedDamage`. `EVENT_TAKE_DAMAGE` fires if and only if actual damage is absorbed by Armor or committed to Health (`(armorAbsorbed + healthDamageCommitted) > 0`).

#### FR-ENG-035: State-Based Death Sweep & Graveyard Transition Batching
- **Functional Contract**: Immediately following damage application, the engine MUST execute a state-based death check sweep ([`FR-ENT-003`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L17)):
  - Any asset with `AccumulatedDamage >= EffectiveDefense` is marked as destroyed.
  - All marked assets are transitioned to their owners' `GraveyardZone` simultaneously in a single atomic batch (`EVENT_ENTITY_DESTROYED`).
  - If an attacking unit has `Berserk` ([`FR-KWD-002o`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L761)) and destroys its Main Theater target, its excess combat damage routes to another chosen opposing asset in the Main Theater as discrete non-combat effect damage.
  - Queues death triggers onto the stack via APNAP ([`FR-ENG-014`](#fr-eng-014-apnap-trigger-dispatching--simultaneous-event-ordering)).

#### FR-ENG-042: Authoritative Match Conclusion & Victory Condition Engine
- **Functional Contract**: The engine MUST continuously evaluate match termination conditions during state-based sweeps and draw operations:
  1. **Headquarters Destruction**: When an HQ asset reaches `AccumulatedDamage >= EffectiveDefense` (`CurrentHealth <= 0`), its controller immediately suffers defeat ([`FR-CBT-003`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L523)).
  2. **Simultaneous Defeat & Match Draw**: If both Headquarters reach Health <= 0 within the exact same atomic frame tick (e.g. simultaneous combat strike or symmetric effect), the match terminates immediately as a **Match Draw** (`RESULT_MATCH_DRAW`, [`FR-CBT-003`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L523)).
  3. **Empty Deck Fatigue**: When a player is required to draw a card from their `DeckZone` and their deck contains 0 cards (`count(DeckZone) == 0`), that player immediately suffers defeat ([`FR-WIN-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L75)).
  4. **Match Concluded Transition**: Upon any terminal condition, the engine emits `MatchConcludedEvent` (`FR-PAC-003`), halts the transaction tick loop, transitions `MatchState` to immutable `STATE_CONCLUDED`, and flushes final `.rec` replay telemetry.

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

### Section 8.1: Error Code Registry & Wire Parity
The headless engine module enforces zero-drift error parity with the client-engine networking protocol. All transaction rollbacks and validation failures emit structured error codes conforming directly to [`FR-ERR-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/04_Client_Engine_Protocol_Networking_FRD.md#L745):

| Error Code | Severity | Trigger Condition | System Action | Wire Traceability |
| :--- | :--- | :--- | :--- | :--- |
| `ERR_STACK_NOT_EMPTY` | Error | Manual action attempted while stack non-empty (`FR-QUE-001`, `FR-INH-001`). | Abort transaction; rollback delta. | [`FR-ERR-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/04_Client_Engine_Protocol_Networking_FRD.md#L746) |
| `ERR_INSUFFICIENT_RESOURCES` | Error | Player lacks Requisition or Tactics to pay cost (`FR-RES-001`). | Abort transaction; rollback delta. | [`FR-ERR-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/04_Client_Engine_Protocol_Networking_FRD.md#L747) |
| `ERR_INVALID_TARGET` | Warning | Target became illegal or moved zones prior to stack pop (`FR-TGT-001`). | Fizzle stack frame; log diagnostic. | [`FR-ERR-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/04_Client_Engine_Protocol_Networking_FRD.md#L748) |
| `ERR_GUARDIAN_PRIORITY_VIOLATION` | Error | Combat attack bypassed an eligible opposing Guardian in Main Theater (`FR-KWD-002j`). | Abort transaction; flash Guardian UI. | [`DA-NET-005`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/04_Client_Engine_Protocol_Networking_FRD.md#L749) |
| `ERR_INVALID_COMMAND_ARGUMENTS` | Error | Command arguments malformed or out of bounds (`FR-MSG-003`). | Abort transaction; rollback delta. | [`DA-NET-003`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/04_Client_Engine_Protocol_Networking_FRD.md#L750) |
| `ERR_ASSET_EXHAUSTED` | Error | Attempted action requiring Readied state (`FR-INH-001`, `FR-CAP-001`). | Abort transaction; rollback delta. | [`FR-ERR-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/04_Client_Engine_Protocol_Networking_FRD.md#L751) |
| `ERR_INVALID_RESOURCE_CHOICE` | Error | Invalid resource generation for current zone or keyword (`FR-INH-002`). | Abort transaction; rollback delta. | [`FR-ERR-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/04_Client_Engine_Protocol_Networking_FRD.md#L752) |
| `ERR_ACTION_CLOCK_EXPIRED` | Warning | Choice prompt timed out without player selection (`FR-TRG-002`). | Apply deterministic fallback (`FR-MOD-001`). | [`FR-ERR-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/04_Client_Engine_Protocol_Networking_FRD.md#L753) |
| `ERR_DESYNC_DETECTED` | Fatal | StateChecksum mismatch between client and server (`FR-PAC-004`). | Terminate match; emit replay dump. | [`FR-PAC-004`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/04_Client_Engine_Protocol_Networking_FRD.md#L690) |
| `ERR_INFINITE_LOOP` | Fatal | Stack execution depth exceeds 100 frames. | Force-clear stack; emit fatal log. | [`FR-ERR-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/04_Client_Engine_Protocol_Networking_FRD.md#L745) |

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
| `FR-ENG-002` | Isolated multi-stream PRNG generator & seed derivation | [`FR-SET-002`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L45), [`FR-SET-003`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L54) |
| `FR-ENG-003` | Deterministic seed derivation with cryptographic salts | Subsumed by [`FR-ENG-002`](#fr-eng-002-isolated-multi-stream-prng-architecture--seed-derivation) |
| `FR-ENG-004` | Pure discrete integer arithmetic | [`FR-LAY-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L431), [`FR-CBT-002`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L504) |
| `FR-ENG-005` | Monotonic tick counter & 64-bit FNV-1a state checksum | [`FR-PAC-004`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/04_Client_Engine_Protocol_Networking_FRD.md#L690) |
| `FR-ENG-006` | Transactional event-emitting match state model & CONCLUDED state | [`FR-ENT-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L29), [`FR-ZON-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L398) |
| `FR-ENG-007` | Match setup, zone initialization & mulligan execution | [`FR-SET-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L29), [`FR-SET-002`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L45), [`FR-SES-004`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/04_Client_Engine_Protocol_Networking_FRD.md#L94) |
| `FR-ENG-008` | Transaction checkpoint & delta-rollback engine with PRNG restore | [`FR-QUE-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L110), [`FR-ERR-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/04_Client_Engine_Protocol_Networking_FRD.md#L745) |
| `FR-ENG-009` | Turn sub-phase transitions conforming to FR-TRN-001 | [`FR-TRN-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L236), [`FR-SET-004`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L57) |
| `FR-ENG-010` | Clean-stack constraint enforcement & inherent actions | [`FR-QUE-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L110), [`FR-INH-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L281), [`FR-INH-002`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L264), [`FR-RES-003`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L228) |
| `FR-ENG-011` | Turn-clock timeout & Reserve Time Bank exhaustion | [`FR-TRN-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/04_Client_Engine_Protocol_Networking_FRD.md#L508), [`FR-SES-005`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/04_Client_Engine_Protocol_Networking_FRD.md#L100) |
| `FR-ENG-012` | LIFO resolution stack & frame taxonomy | [`FR-QUE-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L110) |
| `FR-ENG-013` | Atomic cost validation, reservation & deduction | [`FR-SET-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L29), [`FR-ABL-007`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/02_Card_Data_Pipeline_FRD.md#L273) |
| `FR-ENG-014` | APNAP simultaneous trigger dispatching order | [`FR-TRG-003`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L396) |
| `FR-ENG-015` | Target verification immediately prior to stack pop | [`FR-TGT-102`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L16), [`FR-ABL-008`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/02_Card_Data_Pipeline_FRD.md#L292) |
| `FR-ENG-016` | Reactive replacement watcher interceptor pipeline | [`FR-REP-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L14), [`FR-REP-004`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L455) |
| `FR-ENG-017` | Choice prompt modal yield & action clock execution | [`FR-MSG-005`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/04_Client_Engine_Protocol_Networking_FRD.md#L400), [`FR-MOD-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L361) |
| `FR-ENG-018` | Polymorphic ability AST type-switch visitor | [`FR-ABL-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/02_Card_Data_Pipeline_FRD.md#L154) – [`FR-ABL-005`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/02_Card_Data_Pipeline_FRD.md#L236) |
| `FR-ENG-019` | Dynamic filter evaluator against effective attributes | [`FR-ABL-008`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/02_Card_Data_Pipeline_FRD.md#L292) – [`FR-ABL-011`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/02_Card_Data_Pipeline_FRD.md#L368) |
| `FR-ENG-020` | ValueToken expression resolver & condition gate | [`FR-ABL-006`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/02_Card_Data_Pipeline_FRD.md#L254), [`FR-ABL-012`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/02_Card_Data_Pipeline_FRD.md#L391) |
| `FR-ENG-021` | Canonical atomic effect primitives execution engine | [`FR-PRM-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L304) |
| `FR-ENG-022` | Double-blind Scratch Company deck-to-battlefield lifecycle | [`FR-ZON-005`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L196), [`FR-SEC-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/04_Client_Engine_Protocol_Networking_FRD.md#L164) |
| `FR-ENG-023` | Continuous aura registration & revocation engine | [`FR-ABL-004`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/02_Card_Data_Pipeline_FRD.md#L217), [`FR-TXT-004`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/02_Card_Data_Pipeline_FRD.md#L457) |
| `FR-ENG-024` | Maneuver card resolution & graveyard transition | [`FR-QUE-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L110) |
| `FR-ENG-025` | 6-layer attribute calculation pipeline execution | [`FR-LAY-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L431) – [`FR-LAY-003`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L439) |
| `FR-ENG-026` | Dirty-flag caching & topological dependency invalidation | [`FR-LAY-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L431) |
| `FR-ENG-027` | Timestamp precedence within attribute calculation layers | Subsumed by [`FR-ENG-025`](#fr-eng-025-deterministic-6-layer-attribute-pipeline-architecture) |
| `FR-ENG-028` | Effective attribute clamping and non-negative invariants | Subsumed by [`FR-ENG-025`](#fr-eng-025-deterministic-6-layer-attribute-pipeline-architecture) & [`FR-ENG-041`](#fr-eng-041-system-invariants--diagnostic-assertion-protocol) |
| `FR-ENG-029` | Round-duration temporary modifier cleanup | [`FR-LAY-003`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L439) |
| `FR-ENG-030` | Spatial zone combat restrictions (Main Theater and HQ Zone) | [`FR-CBT-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L282), [`FR-CAP-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L265) |
| `FR-ENG-031` | Step 1 attack declaration, validation & readiness lock | [`FR-CBT-002`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L504), [`FR-QUE-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L110) |
| `FR-ENG-032` | Step 2 reaction window & trigger queue | [`FR-CBT-002`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L504) |
| `FR-ENG-033` | Step 3 simultaneous combat damage & symmetric retaliation | [`FR-CBT-002`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L504) |
| `FR-ENG-034` | Canonical 5-stage damage mitigation pipeline (Redirection/Resistant/Stalwart/Armor) | [`FR-KWD-002c`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L749), [`FR-REP-004`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L455) |
| `FR-ENG-035` | State-based death check sweep & graveyard transition | [`FR-ENT-003`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L17) |
| `FR-ENG-036` | Headless engine CLI test runner architecture | [`FR-SES-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/04_Client_Engine_Protocol_Networking_FRD.md#L68) |
| `FR-ENG-037` | Scripted scenario fixture DSL & assertion framework | [`FR-PAC-004`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/04_Client_Engine_Protocol_Networking_FRD.md#L690) |
| `FR-ENG-038` | Deterministic .rec match replay recording & validation | [`FR-SPC-002`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/04_Client_Engine_Protocol_Networking_FRD.md#L735) |
| `FR-ENG-039` | Combinatorial monkey/fuzz testing orchestrator | [`FR-PAC-004`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/04_Client_Engine_Protocol_Networking_FRD.md#L690) |
| `FR-ENG-040` | Engine performance benchmarks & throughput gates | [`FR-ENG-001`](#fr-eng-001-engine-modularity--zero-engine-dependency-invariant) |
| `FR-ENG-041` | System invariants & diagnostic assertions | [`FR-ENT-003`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L17), [`FR-QUE-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L110) |
| `FR-ENG-042` | Authoritative match conclusion & victory condition engine | [`FR-WIN-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L75), [`FR-CBT-003`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L523), [`FR-PAC-003`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/04_Client_Engine_Protocol_Networking_FRD.md#L430) |
