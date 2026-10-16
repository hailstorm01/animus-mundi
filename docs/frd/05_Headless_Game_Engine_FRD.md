# Headless Game Engine - Functional Requirements Document (Master FRD v1.1)
**Authoritative Headless Match Kernel, Runtime Entity Architecture, Deterministic State Machine & Execution Engine**

---

## EXECUTIVE SUMMARY & REVISION HISTORY

This document establishes the authoritative technical specification for the **Card Town Headless Game Engine**, governing **Phase 3 (Core Headless Game Engine Implementation)** of the product roadmap in [`docs/PRODUCT_SUMMARY.md`](../PRODUCT_SUMMARY.md#L198-L212).

To maintain strict separation of concerns across the specification suite (`SOC-002`):
- [`01_Game_Rules_FRD.md`](01_Game_Rules_FRD.md) codifies **what the game rules are** in pure tabletop domain terminology.
- [`02_Card_Data_Pipeline_FRD.md`](02_Card_Data_Pipeline_FRD.md) and [`03_Card_Compiler_FRD.md`](03_Card_Compiler_FRD.md) define the **declarative modeling bridge** (`blueprints.json`, `AbilityDef` AST, event hooks `FR-TRG-001`, and the 25 atomic primitives `FR-PRM-001`).
- **This document (`05_Headless_Game_Engine_FRD.md`)** specifies **how** the standalone Go match kernel models entities in memory (`FR-ENT-001`–`FR-ENT-003`, `FR-SNP-101`) and executes the rules of `01` and AST primitives of `02` with **100% server-authoritative determinism**.

The engine is engineered as a standalone, dependency-free **Go (Golang)** module with zero graphics, UI, or game engine libraries. It operates as the core match runner within the dedicated server daemon on GCP Agones ([`FR-SES-001`](04_Client_Engine_Protocol_Networking_FRD.md#L68)) and as a bundled local sidecar executable for offline play and bot testing, preserving complete decoupling from the Godot 4.x presentation client ([`FR-MSG-001`](04_Client_Engine_Protocol_Networking_FRD.md#L248)).

```text
+-----------------------------------------------------------------------------------+
|                           CARD TOWN HEADLESS ENGINE                               |
|                                                                                   |
|  +-------------------------+   +-----------------------------------------------+  |
|  |   Deterministic PRNG    |   |     Transactional Event-Emitting Match        |  |
|  |  Multi-Stream Generator |   |                  State                        |  |
|  |   (PCG32 / SplitMix64)  |   |  (Zones, 3-Tier Entities, Resources, Clocks)  |  |
|  +------------+------------+   +-----------------------+-----------------------+  |
|               |                                        |                          |
|               v                                        v                          |
|  +-------------------------+   +-----------------------+-----------------------+  |
|  |  6-Layer Attribute Calc |   |            Resolution Stack                   |  |
|  |   Stateless On-Demand   |   |       LIFO Frames & Trigger Queue             |  |
|  |  (Layers 1 to 6 / Clamp)|   |        (APNAP Dispatcher / Rollback)          |  |
|  +------------+------------+   +-----------------------+-----------------------+  |
|               |                                        |                          |
|               +-------------------+--------------------+                          |
|                                   |                                               |
|                                   v                                               |
|  +-----------------------------------------------------------------------------+  |
|  |                     Declarative Ability AST Interpreter                     |  |
|  |            (Type-Switch Visitor & 25 Atomic Primitives Execution)           |  |
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

### Specification Scope & Revision Summary (v1.0 -> v1.1):
1. **Deterministic Engine Kernel & PRNG Subsystem (`FR-ENG-001` - `FR-ENG-005`)**: Multi-stream pseudo-random number generator, pure integer arithmetic, and rolling state hash generation.
2. **Authoritative Match State Machine, 3-Tier Entity Architecture & Phase Engine (`FR-ENG-006` - `FR-ENG-011`, `FR-ENT-001` - `FR-ENT-003`, `FR-SNP-101`)**: Relocated runtime memory struct definitions (`CardBlueprint`, `CardInstance`, `AssetInstance`, `AssetSnapshot`) from `01` (`SOC-002`); aligned turn sub-phases and pre-game setup with [`01_Game_Rules_FRD.md`](01_Game_Rules_FRD.md) (`DA-ENG-001`, `DA-ENG-004`).
3. **Resolution Stack & APNAP Trigger Dispatcher (`FR-ENG-012` - `FR-ENG-017`)**: LIFO stack runner, atomic cost deduction, APNAP simultaneous trigger ordering, inline replacement watcher execution (`DA-ENG-002`), two-stage target verification (`DA-ENG-005`), and interactive choice fast-pathing (`DA-ENG-008`).
4. **Declarative Ability AST Interpreter & Atomic Primitives (`FR-ENG-018` - `FR-ENG-024`)**: Polymorphic AST type-switch evaluator and execution contracts referencing [`02_Card_Data_Pipeline_FRD.md`](02_Card_Data_Pipeline_FRD.md) without schema duplication (`DA-ENG-006`).
5. **6-Layer Attribute Recalculation Pipeline (`FR-ENG-025` - `FR-ENG-029`)**: Stateless continuous on-demand evaluation engine, exclusive Layer 6 Combat Power clamping (`DA-ENG-003`), and round-duration cleanup.
6. **Combat Resolution & Retaliation Engine (`FR-ENG-030` - `FR-ENG-035`, `FR-ENG-042`)**: 3-step combat cycle orchestration, upfront `Berserk` secondary target binding (`SOC-001`), 5-stage damage mitigation with `(armorAbsorbed + healthDamageCommitted) > 0` trigger gate, `ATTACHED` sub-zone host-departure cleanup, and recursive post-`Berserk` death sweep (`DA-ENG-007`).
7. **Headless Test Harness, Replay & Simulation Suite (`FR-ENG-036` - `FR-ENG-040`)**: Headless CLI runner, scripted scenario DSL, deterministic `.rec` replay validator, and combinatorial fuzz testing.
8. **Error Taxonomy, Resilience & Diagnostics (`FR-ENG-041`)**: Structured error code catalog and parity assertion rules.

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
> - **Inference**: The card catalog is loaded dynamically from `blueprints.json` ([`FR-CMP-025`](03_Card_Compiler_FRD.md#L457)), requiring flexible ability execution that matches the declarative AST schema ([`FR-ABL-001`](02_Card_Data_Pipeline_FRD.md#L222)).
> - **Decision**: Implement a direct tree-walking AST interpreter using Go type-switches across sealed interface structs. Each AST node evaluates directly against an execution context and dispatches to atomic effect primitives.
> - **Alternative Considered**: JIT bytecode compilation. Unnecessary complexity; benchmarks indicate direct tree-walking in Go evaluates card abilities in <5 microseconds per invocation, exceeding target throughput.

> [!IMPORTANT]
> **DECISION 4: Cross-Platform Pure Integer Arithmetic (`FR-ENG-004`)**
> - **Inference**: Cross-platform simulation parity and replay reproduction ([`FR-SPC-002`](04_Client_Engine_Protocol_Networking_FRD.md#L735)) fail if floating-point calculations produce bitwise divergence across CPU architectures (x86_64 vs. ARM64).
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
       Deck Shuffling             Random Targeting           First-Player /
       (Fisher-Yates)             Selection Pool             Tie-Breakers
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
  1. `StreamShuffle`: Dedicated exclusively to initial deck shuffling ([`FR-SET-002`](01_Game_Rules_FRD.md#L48)) and mid-match deck shuffles ([`FR-DEK-002`](01_Game_Rules_FRD.md#L70), [`FR-PRM-001`](02_Card_Data_Pipeline_FRD.md#L382)). Derived via `SeedShuffle = MasterSeed ^ 0x9E3779B97F4A7C15`.
  2. `StreamTargeting`: Dedicated exclusively to resolving random target selection (`selector: "RANDOM"`, [`FR-ABL-007`](02_Card_Data_Pipeline_FRD.md#L297)) and timeout target fallbacks ([`FR-MOD-001`](01_Game_Rules_FRD.md#L267)). Derived via `SeedTargeting = MasterSeed ^ 0xBF58476D1CE4E5B9`.
  3. `StreamMatchDecisions`: Dedicated to pre-game first-player determination ([`FR-SET-003`](01_Game_Rules_FRD.md#L44)). Derived via `SeedMatchDecisions = MasterSeed ^ 0x94D049BB133111EB`.
- **Invariants**: Shuffling a deck MUST NOT advance or alter the sequence of random target selections during combat. Supplying a seed of `0` MUST be rejected with `ERR_INVALID_COMMAND_ARGUMENTS` or initialized from a non-zero cryptographic entropy source.

#### FR-ENG-003: Subsumed by FR-ENG-002
- *Deterministic seed derivation and cryptographic salt constants are formally subsumed into [`FR-ENG-002`](#fr-eng-002-isolated-multi-stream-prng-architecture--seed-derivation).*

#### FR-ENG-004: Pure Discrete Integer Arithmetic & Clamping Invariants
- **Functional Contract**: All numeric calculations within the engine MUST execute using discrete 32-bit signed integers (`int32`).
- **Rules**:
  - Division operations resulting in fractional quantities MUST perform deterministic floor integer truncation (`a / b`).
  - Clamping obeys the domain-specific rules of [`01_Game_Rules_FRD.md`](01_Game_Rules_FRD.md): **Combat Power** clamps at `0` in Layer 6 ([`FR-LAY-001`](01_Game_Rules_FRD.md#L329)), effective Requisition and Tactics costs clamp at `0` ([`FR-CST-001`](01_Game_Rules_FRD.md#L196)), dynamic **Armor** and **Accumulated Damage** counters clamp at `0` ([`FR-LAY-002`](01_Game_Rules_FRD.md#L333)), and post-Stage-2 damage clamps at `0` ([`FR-REP-004`](01_Game_Rules_FRD.md#L356)). Conversely, **Resistant** is signed (negative values amplify incoming damage in Stage 2 of `FR-REP-004`), and **Maximum Health** / **Current Health** do not clamp at `0` because values $\le 0$ trigger immediate state-based destruction (`FR-LAY-002`).
  - Floating-point arithmetic (`float32`, `float64`) is strictly prohibited in rules evaluation.

#### FR-ENG-005: Frame Tick Invariants & Rolling State Hash Generation
- **Functional Contract**: The engine MUST maintain a strictly monotonic 64-bit tick counter (`TickNumber`) incremented on every committed transaction and phase transition.
- **State Checksum**: At the conclusion of every turn sub-phase and stack frame resolution, the engine MUST calculate a rolling 64-bit FNV-1a state checksum (`StateChecksum`) across all public entity states, zone contents, and resource pools in strict conformance with [`FR-PAC-004`](04_Client_Engine_Protocol_Networking_FRD.md#L690).
- **Output**: Transmitted to clients within periodic synchronization frames to detect desync divergence immediately.

---

## DOMAIN 2: AUTHORITATIVE MATCH STATE MACHINE, ENTITY ARCHITECTURE & PHASE ENGINE

```text
                          +------------------------------+
                          |          MATCH SETUP         |
                          | 1. Place HQs in HQ Zones     |
                          | 2. First Player Choice (15s) |
                          | 3. Shuffle & Deal 6 Cards    |
                          | 4. Partial Mulligan (25s)    |
                          +--------------+---------------+
                                         |
                                         v
                      +--------------------------------------+
                      |             READY PHASE              |
                      | - Sub-Phase 1a: Unexhaust Controlled |
                      |   Assets & Reset Per-Turn Counters   |
                      |   (Zero Passive Resource Income)     |
                      | - Sub-Phase 1b: Start-of-Turn        |
                      |   Triggers Enqueue & Resolve (APNAP) |
                      +------------------+-------------------+
                                         |
                                         v
                      +--------------------------------------+
                      |              DRAW PHASE              |
                      | - Sub-Phase 2a: Draw 1 Card          |
                      |   (Skipped on Player 1 Turn 1)       |
                      | - Sub-Phase 2b: Draw Triggers        |
                      |   Enqueue & Resolve (APNAP)          |
                      +------------------+-------------------+
                                         |
                                         v
                      +--------------------------------------+ <------------------+
                      |              MAIN PHASE              |                    |
                      | - Clean-Stack Manual Actions         |  Action Execution  |
                      | - Off-Stack Resource Generation/Conv |  & Stack Cleared   |
                      | - Deployments / Maneuvers / Combat   |                    |
                      +------------------+-------------------+ -------------------+
                                         |
                                         v
                      +--------------------------------------+
                      |              END PHASE               |
                      | - Sub-Phase 4a: End Trigger Enqueue  |
                      | - Sub-Phase 4b: Cascade Resolution   |
                      | - Sub-Phase 4c: Pool Depletion (0)   |
                      |   & Temporary Modifier Expiration    |
                      | - Sub-Phase 4d: Turn Handover        |
                      |   (Universal Uncapped Hand Size)     |
                      +------------------+-------------------+
                                         |
                                         v
                      +--------------------------------------+
                      |         HANDOVER / NEXT TURN         |
                      +--------------------------------------+
```

### Section 2.1: Transactional Match State Model

#### FR-ENG-006: Transactional Event-Emitting State Model
- **Functional Contract**: The authoritative match state (`MatchState`) MUST be held in memory as an in-place mutable object graph, guarded by an active `Transaction` context.
- **Structure**:
  - `MatchID`: Unique string identifier.
  - `TurnNumber`: Monotonically increasing turn counter (starts at 1).
  - `ActivePlayer`: ID of the player currently taking their turn (0 or 1).
  - `CurrentPhase`: Current match phase (`PHASE_READY`, `PHASE_DRAW`, `PHASE_MAIN`, `PHASE_END`, `STATE_CONCLUDED`).
  - `Zones`: Zone topology container (`HeadquartersZone`, `ReservesZone`, `MainTheaterZone`, `HandZone`, `DeckZone`, `DiscardZone`, and per-asset `AttachedZone`; [`FR-ZON-001`](01_Game_Rules_FRD.md#L112) – [`FR-ZON-004`](01_Game_Rules_FRD.md#L131)).
  - `ResolutionStack`: The LIFO execution stack ([`FR-QUE-001`](01_Game_Rules_FRD.md#L284)).
  - `ResourcePools`: Requisition and Tactics integer pools per player.
  - `EventBuffer`: In-memory slice of pending `EngineEvent` objects emitted during the transaction.

#### FR-ENG-007: Match Initialization, Setup Handshake & Zone Instantiation
- **Functional Contract**: Upon receiving valid deck payloads conforming to [`FR-SET-001`](01_Game_Rules_FRD.md#L28), the engine MUST execute match initialization in the exact chronological sequence mandated by [`FR-SET-003`](01_Game_Rules_FRD.md#L44), [`FR-SET-002`](01_Game_Rules_FRD.md#L48), and [`FR-SES-004`](04_Client_Engine_Protocol_Networking_FRD.md#L94):
  1. **Headquarters Instantiation**: Instantiates `HeadquartersZone` containing each player's designated HQ `AssetInstance` visible to both players ([`FR-ZON-001`](01_Game_Rules_FRD.md#L112), [`FR-ZON-003`](01_Game_Rules_FRD.md#L102)).
  2. **First-Player Determination (15s Timer)**: Selects the decision-making player via `StreamMatchDecisions` and issues a 15-second choice prompt to **Play First** (take Turn 1) or **Draw First** (take Turn 2), defaulting to **Play First** on timeout ([`FR-SET-003`](01_Game_Rules_FRD.md#L44), [`FR-SES-004`](04_Client_Engine_Protocol_Networking_FRD.md#L94)).
  3. **Deck Shuffle & Opening Hand Deal**: Populates each player's `DeckZone` with their 50+ `CardInstance` structs, performs an initial Fisher-Yates shuffle via `StreamShuffle`, and deals an initial 6-card hand into each player's `HandZone` ([`FR-SET-002`](01_Game_Rules_FRD.md#L48)).
  4. **Simultaneous 4-Step Partial Mulligan (25s Timer)**: Opens a simultaneous 25-second partial mulligan window ([`FR-SES-004`](04_Client_Engine_Protocol_Networking_FRD.md#L94), defaulting to keeping all 6 cards on timeout). For each player who submits $N$ selected hand cards ($0 \le N \le 6$), the engine executes the 4-step set-aside algorithm of [`FR-SET-002`](01_Game_Rules_FRD.md#L48): (a) sets the $N$ selected cards aside in a temporary buffer, (b) draws $N$ replacement cards from the remaining `DeckZone`, and (c) returns the $N$ set-aside cards to `DeckZone` and shuffles via `StreamShuffle` (guaranteeing set-aside cards cannot be redrawn during the mulligan).

#### FR-ENG-008: Transaction Checkpoint, Delta-Rollback & Commit Lifecycle
- **Functional Contract**: Every player command, stack resolution step, and phase transition MUST execute inside an atomic transaction:
  - `tx := match.BeginTransaction()`
  - Before modifying state, the transaction records a lightweight delta-snapshot of touched entities, zones, resource balances, and internal monotonic sequence counters of all active PRNG streams ([`FR-ENG-002`](#fr-eng-002-isolated-multi-stream-prng-architecture--seed-derivation)).
  - **On Failure / Illegal Command**: `tx.Rollback()` restores touched entities, resource balances, and PRNG stream counters to their exact pre-transaction state, purges `EventBuffer`, and returns a structured error code ([`FR-ERR-001`](04_Client_Engine_Protocol_Networking_FRD.md#L745)). The active match state and PRNG sequence determinism remain completely uncorrupted.
  - **On Success**: `tx.Commit()` applies state mutations permanently, calculates the rolling state hash ([`FR-ENG-005`](#fr-eng-005-frame-tick-invariants--rolling-state-hash-generation)), flushes `EventBuffer` to the outgoing network/replay channel, and discards the delta-snapshot.

---

### Section 2.2: 3-Tier Runtime Entity Memory Architecture & Snapshot Model

To enforce zero mutation of compiled card data and clean state resets across zone transitions ([`FR-ZON-004`](01_Game_Rules_FRD.md#L131)), the engine models cards and battlefield entities across a strict **3-Tier Runtime Entity Architecture** plus an immutable **Last-Known State Snapshot** struct:

```text
+-----------------------------------------------------------------------+
| Tier 1: CardBlueprint (Immutable Process Singleton - FR-ENT-001)      |
| - id, name, cardType, faction, subtypes, base costs, base stats, AST  |
+-----------------------------------+-----------------------------------+
                                    | (1:N read-only pointer)
                                    v
+-----------------------------------------------------------------------+
| Tier 2: CardInstance (Session Card Struct - FR-ENT-002)               |
| - instanceId, blueprintPtr, ownerId, currentZone, isKnownTo[2]        |
+-----------------------------------+-----------------------------------+
                                    | (Wrapped while on Battlefield)
                                    v
+-----------------------------------------------------------------------+
| Tier 3: AssetInstance (Battlefield Entity Struct - FR-ENT-003)        |
| - assetId, underlyingCardPtr, controllerId, battlefieldZone,          |
|   readiness, accumulatedDamage, currentArmor, isFaceDown,             |
|   attachedCards[], activeModifiers[], turnActivationCount[]           |
+-----------------------------------+-----------------------------------+
                                    | (Captured on Zone Departure)
                                    v
+-----------------------------------------------------------------------+
| Snapshot: AssetSnapshot (Last-Known State Struct - FR-SNP-101)        |
| - effectivePower, effectiveMaxHealth, currentHealth, currentArmor,    |
|   controllerId, lastBattlefieldZone, effectiveSubtypes/Keywords       |
+-----------------------------------------------------------------------+
```

#### FR-ENT-001: CardBlueprint Immutable Runtime Singleton
- **Functional Contract**: At process startup, the engine loads `blueprints.json` ([`FR-DAT-003`](02_Card_Data_Pipeline_FRD.md#L90), [`FR-DAT-005`](02_Card_Data_Pipeline_FRD.md#L107)) into an immutable, read-only in-memory lookup table of `CardBlueprint` structs indexed by canonical blueprint `id` ([`FR-DAT-002`](02_Card_Data_Pipeline_FRD.md#L79)).
- **Invariants**: `CardBlueprint` structs are shared across all concurrent match goroutines and MUST NEVER be mutated at runtime.

#### FR-ENT-002: CardInstance Session Card Struct
- **Functional Contract**: For every physical card entering a match (each player's 1 Headquarters card and 50+ Main Deck cards), the engine instantiates a `CardInstance` struct:
  - `InstanceID`: Match-unique integer/string identifier.
  - `Blueprint`: Read-only pointer to its `CardBlueprint` (`FR-ENT-001`).
  - `OwnerID`: Immutable player ID (`0` or `1`) of the deck owner ([`FR-ZON-003`](01_Game_Rules_FRD.md#L102)).
  - `CurrentZone`: Current zone enum (`HEADQUARTERS`, `RESERVES`, `MAIN_THEATER`, `HAND`, `DECK`, `DISCARD`, `ATTACHED`, or `STACK`).
  - `VisibilityMask`: Per-player boolean visibility tracking whether the card's printed face has been revealed in a private zone ([`FR-MOD-001`](01_Game_Rules_FRD.md#L267)).

#### FR-ENT-003: AssetInstance Battlefield Entity Struct
- **Functional Contract**: While a Unit, Structure, or Headquarters occupies a battlefield zone (`HEADQUARTERS`, `RESERVES`, or `MAIN_THEATER`), it is represented by a stateful `AssetInstance` wrapping its underlying `CardInstance` (`FR-ENT-002`):
  - `AssetID`: Unique runtime entity identifier (or opaque `proxy_scratch_*` identifier while `IsFaceDown == true`; [`FR-ZON-005`](01_Game_Rules_FRD.md#L136), [`FR-SEC-001`](04_Client_Engine_Protocol_Networking_FRD.md#L164)).
  - `UnderlyingCard`: Pointer to the wrapped `CardInstance`.
  - `ControllerID`: Player ID (`0` or `1`) currently commanding the asset ([`FR-ZON-002`](01_Game_Rules_FRD.md#L125), [`FR-ZON-003`](01_Game_Rules_FRD.md#L102)).
  - `BattlefieldZone`: `HEADQUARTERS`, `RESERVES`, or `MAIN_THEATER`.
  - `Readiness`: `READIED` or `EXHAUSTED` ([`FR-ZON-003`](01_Game_Rules_FRD.md#L102)).
  - `AccumulatedDamage`: Integer damage counters marked on the asset ($\ge 0$; [`FR-LAY-002`](01_Game_Rules_FRD.md#L333), [`FR-CBT-003`](01_Game_Rules_FRD.md#L422)).
  - `CurrentArmor`: Persistent consumable integer Armor pool ($\ge 0$; [`FR-LAY-002`](01_Game_Rules_FRD.md#L333), [`FR-KWD-002k`](01_Game_Rules_FRD.md#L443)).
  - `IsFaceDown`: Boolean flag indicating whether the asset is a `1/1/1/1 Infantry Scratch Company` ([`FR-ZON-005`](01_Game_Rules_FRD.md#L136), [`FR-ENT-103`](01_Game_Rules_FRD.md#L146)).
  - `AttachedCards`: Slice of `CardInstance` pointers tucked under this host asset (`ATTACHED` sub-zone; [`FR-ENT-102`](01_Game_Rules_FRD.md#L142)).
  - `ActiveModifiers`: Timestamped slice of Layer 2, 3, and 5 modifiers applied to this asset ([`FR-LAY-001`](01_Game_Rules_FRD.md#L329)).
  - `TurnActivationCounts`: Per-ability integer execution counters for `maxPerTurn` gates ([`FR-ABL-008`](02_Card_Data_Pipeline_FRD.md#L423)).
- **Zone Departure Deallocation & Attached Cleanup**: Whenever an `AssetInstance` leaves the battlefield for a non-battlefield zone (`HAND`, `DISCARD`, `DECK`, or `ATTACHED`), the engine first captures an immutable `AssetSnapshot` (`FR-SNP-101`), routes any `AttachedCards` face-up to their respective owners' `DiscardZone` (emitting `EVENT_DISCARD_CARD` per [`FR-ENT-102`](01_Game_Rules_FRD.md#L142)), deallocates the `AssetInstance` wrapper, and places the clean `CardInstance` into its destination zone ([`FR-ZON-004`](01_Game_Rules_FRD.md#L131)).

#### FR-SNP-101: AssetSnapshot Last-Known State Snapshot Struct
- **Functional Contract**: Immediately prior to deallocating an `AssetInstance` upon destruction or battlefield departure ([`FR-ZON-004`](01_Game_Rules_FRD.md#L131)), the engine MUST capture an immutable `AssetSnapshot` struct and attach it to the resulting `EVENT_DESTROY` / `EVENT_ZONE_CHANGE` causality context ([`FR-TRG-007`](02_Card_Data_Pipeline_FRD.md#L379)):
  - Captures Layer 6 `EffectivePower`, `EffectiveMaxHealth`, `CurrentHealth`, `CurrentArmor`, `ControllerID`, `OwnerID`, `LastBattlefieldZone`, `Readiness`, `IsFaceDown`, `EffectiveSubtypes`, `EffectiveKeywords`, and `AttachedCardCount`.
  - Any reactive triggers or `ValueToken` queries evaluating the departed asset (e.g., *"When this unit dies, deal damage equal to its power"*) read directly from `AssetSnapshot`.

---

### Section 2.3: Turn Phase Engine & Action Routing

#### FR-ENG-009: Deterministic Turn Sequencing & Sub-Phase Transitions
- **Functional Contract**: The engine MUST execute turn phases and associated sub-phases in strict sequential order conforming to [`FR-TRN-001`](01_Game_Rules_FRD.md#L170):
  - **Phase 1: Ready Phase**:
    - *Sub-Phase 1a (Unexhaust & Counter Reset)*: Transitions all `EXHAUSTED` assets controlled by `ActivePlayer` across `ReservesZone`, `MainTheaterZone`, and `HeadquartersZone` to `READIED`, and resets all `TurnActivationCounts` to `0` ([`FR-ABL-008`](02_Card_Data_Pipeline_FRD.md#L423)). Conforming to [`FR-RES-001`](01_Game_Rules_FRD.md#L155) and [`FR-TRN-001`](01_Game_Rules_FRD.md#L170), the engine grants **zero** passive/base Requisition or Tactics during the Ready Phase.
    - *Sub-Phase 1b (Start-of-Turn Triggers)*: Emits `EVENT_READY_PHASE` / `EVENT_TURN_START`, enqueues all *"At the start of your turn"* triggered abilities (*Hammerhand Bivouac*) onto `ResolutionStack` in APNAP order ([`FR-TRG-003`](01_Game_Rules_FRD.md#L294)), and auto-resolves the stack to completion before advancing to the Draw Phase.
  - **Phase 2: Draw Phase**:
    - *Sub-Phase 2a (Turn Draw)*: `ActivePlayer` draws 1 card from `DeckZone` into `HandZone` (emitting `EVENT_DRAW_PHASE` and `EVENT_DRAW_CARD`, or triggering Empty Deck Loss if `DeckZone` is empty per [`FR-WIN-001`](01_Game_Rules_FRD.md#L73)). Skipped on Turn 1 for the player taking the first turn ([`FR-SET-004`](01_Game_Rules_FRD.md#L57)).
    - *Sub-Phase 2b (Draw Phase Triggers)*: Enqueues any abilities triggered by the draw (*Portent Scrivener*) onto `ResolutionStack` in APNAP order and auto-resolves the stack to completion before advancing to the Main Phase.
  - **Phase 3: Main Phase**: Emits `EVENT_MAIN_PHASE` and opens the clean-stack action priority window for `ActivePlayer` deployments, maneuvers, custom activated abilities, and inherent actions.
  - **Phase 4: End Phase**:
    - *Sub-Phase 4a (Event Emission & Trigger Enqueue)*: Emits `EVENT_END_PHASE` / `EVENT_TURN_END`; enqueues *"At the end of your turn"* triggers via APNAP ([`FR-TRG-003`](01_Game_Rules_FRD.md#L294)).
    - *Sub-Phase 4b (Trigger Cascade Resolution)*: `ResolutionStack` auto-resolves to completion.
    - *Sub-Phase 4c (Resource Depletion & Modifier Expiration)*: Unspent Requisition and Tactics pools reset to `0` ([`FR-RES-002`](01_Game_Rules_FRD.md#L160)); temporary Layer 5 modifiers expiring at the end of the current turn expire ([`FR-LAY-003`](01_Game_Rules_FRD.md#L338)).
    - *Sub-Phase 4d (Turn Handover)*: Asserts `ResolutionStack.IsEmpty() == true`, increments `TurnNumber`, and toggles `ActivePlayer`. Universal uncapped hand capacity is preserved per [`FR-ZON-001`](01_Game_Rules_FRD.md#L112) with zero hand-size discard checks.

#### FR-ENG-010: Clean-Stack Manual Action Constraint & Inherent Action Routing
- **Functional Contract**: The engine MUST enforce the action routing and clean-stack invariants codified in [`FR-QUE-001`](01_Game_Rules_FRD.md#L284), [`FR-CAP-001`](01_Game_Rules_FRD.md#L203), and [`FR-INH-001`–`FR-INH-004`](01_Game_Rules_FRD.md#L223):
  - **Stack-Pushed Manual Actions (`ResolutionStack.IsEmpty() == true` Required)**: Deploying asset cards (`FR-TRN-003`), enacting maneuvers (`FR-MNV-001`), activating custom abilities (`FR-ACT-001`), and declaring tactical inherent actions (`March` [`FR-INH-002`], `Attack` [`FR-CBT-002`], and `Deploy Scratch Company` [`FR-INH-004`]) push frames onto `ResolutionStack`. Attempting any of these while `ResolutionStack.IsEmpty() == false` aborts the transaction with `ERR_STACK_NOT_EMPTY`.
  - **Stack-Exempt Operations (No Clean-Stack Requirement)**: Economic resource operations—Inherent Resource Generation ([`FR-RES-001`](01_Game_Rules_FRD.md#L155), [`FR-CAP-001`](01_Game_Rules_FRD.md#L203)) and 2R&rarr;1T Resource Conversion ([`FR-RES-003`](01_Game_Rules_FRD.md#L163))—execute instantaneously off-stack without requiring `ResolutionStack.IsEmpty() == true`, alongside triggered abilities (`FR-TRG-001`), in-resolution free deployments ([`FR-TRN-003a`](01_Game_Rules_FRD.md#L233)), and inline replacement watchers ([`FR-REP-001`](01_Game_Rules_FRD.md#L348)).

#### FR-ENG-011: Turn-Clock Timeout Resolution & Reserve Time Bank Invariants
- **Functional Contract**: The engine MUST enforce the match timer hierarchy specified in [`FR-TRN-001`](01_Game_Rules_FRD.md#L170), [`FR-SES-004`](04_Client_Engine_Protocol_Networking_FRD.md#L94), and [`FR-SES-005`](04_Client_Engine_Protocol_Networking_FRD.md#L100):
  1. Tracks the active player's **75-second Main Phase turn clock** (emitting a warning flag at 20 seconds remaining) and each player's **60-second match Reserve Time Bank**.
  2. When the 75-second turn clock (or a 15s/10s interactive action clock in [`FR-ENG-017`](#fr-eng-017-choice-prompt-yield-fast-path--resumption-protocol)) reaches `0`, the engine immediately begins depleting that player's remaining Reserve Time Bank.
  3. When both the active timer and the player's Reserve Time Bank reach `0`, the engine automatically resolves any pending choice prompt using the deterministic fallback rules of [`FR-MOD-001`](01_Game_Rules_FRD.md#L267) or issues an automated Main Phase pass transitioning the turn to the End Phase.

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
                  |  [Bot: Root Action (Deploy / Maneuver / |
                  |        Ability / March / Attack / HQ)]  |
                  +--------------------+--------------------+
                                       |
                                       v
                  +-----------------------------------------+
                  |   Pop Top Frame & Re-Verify Targets     |
                  |             (FR-TGT-003)                |
                  +--------------------+--------------------+
                                       |
                        +--------------+--------------+
                        | At Least 1 Legal Target     |
                        | (or Un-Targeted Clause)?    |
                        v                             v
                     [ YES ]                       [ NO ]
                        |                             |
                        v                             v
           Execute Legal Clauses via         Fizzle Targeted Clause &
           FR-ENG-021 (Evaporate Dead        Log ERR_INVALID_TARGET
           Pool Allocations)                 (FR-TGT-003)
```

### Section 3.1: Stack Execution & Trigger Ordering

#### FR-ENG-012: LIFO Stack Architecture & Stack Frame Taxonomy
- **Functional Contract**: The engine MUST implement the Resolution Stack as a Last-In, First-Out (`LIFO`) data structure storing strongly-typed `StackFrame` instances conforming to [`FR-QUE-001`](01_Game_Rules_FRD.md#L284):
  - `FRAME_DEPLOYMENT`: A played Unit or Structure card awaiting battlefield instantiation in its declared zone (`RESERVES` or `MAIN_THEATER`; [`FR-TRN-003`](01_Game_Rules_FRD.md#L233)).
  - `FRAME_MANEUVER`: A played Maneuver card awaiting effect resolution ([`FR-MNV-001`](01_Game_Rules_FRD.md#L233)).
  - `FRAME_ACTIVATED_ABILITY`: A custom activated ability ([`FR-ACT-001`](01_Game_Rules_FRD.md#L229)) or tactical inherent ability (`March` [`FR-INH-002`] or `Deploy Scratch Company` [`FR-INH-004`]) queued by paying its upfront `CostDef` array.
  - `FRAME_ATTACK`: An inherent combat attack declaration awaiting simultaneous damage and retaliation resolution ([`FR-CBT-002`](01_Game_Rules_FRD.md#L405)).
  - `FRAME_TRIGGERED_ABILITY`: A reactive ability queued in response to an emitted engine event ([`FR-TRG-001`](02_Card_Data_Pipeline_FRD.md#L369), [`FR-ABL-002`](02_Card_Data_Pipeline_FRD.md#L250)).
- **Invariants**: Once a root action and any resulting APNAP triggers are pushed, the stack auto-resolves top-down (LIFO), pausing only for interactive choice prompts ([`FR-ENG-017`](#fr-eng-017-choice-prompt-yield-fast-path--resumption-protocol)). Replacement watchers intercept prospective events inline via [`FR-ENG-016`](#fr-eng-016-reactive-replacement-watcher-interceptor-pipeline) without pushing stack frames.

#### FR-ENG-013: Atomic Cost Validation, Resource Reservation & Deduction
- **Functional Contract**: Prior to executing an economic operation or pushing an action frame onto the stack, the engine MUST atomically validate and deduct all required `CostDef` components per [`FR-CST-001`](01_Game_Rules_FRD.md#L196), [`FR-INH-001`](01_Game_Rules_FRD.md#L223), [`FR-ACT-002`](01_Game_Rules_FRD.md#L229), [`FR-TRN-003a`](01_Game_Rules_FRD.md#L233), [`FR-ABL-007`](02_Card_Data_Pipeline_FRD.md#L297), and [`FR-ABL-009`](02_Card_Data_Pipeline_FRD.md#L427):
  - **Resource Costs (`RESOURCE`)**: Deducts Effective Requisition and/or Tactics costs (`FR-CST-001`) from the player's active resource pools.
  - **Conditional Physical Exhaustion (`EXHAUST_SELF`)**: Enforces `Readiness == READIED` and immediately transitions the source asset to `EXHAUSTED` (emitting `EVENT_EXHAUST`) **if and only if** the action requires exhaustion—specifically all inherent capabilities (`FR-CAP-001`, `FR-INH-001`) and custom activated abilities whose `CostDef` array includes `costType: "EXHAUST_SELF"` (`FR-ABL-009`). Custom activated abilities without `EXHAUST_SELF` do not require the source asset to be Readied (`FR-ACT-002`).
  - **Non-Resource Costs (`KILL_ASSET`, `SACRIFICE_ALLY`, `DISCARD_CARD`)**: Validates and executes required sacrifices or discards upfront upon declaration (`FR-TRN-003a`). Any triggers caused by paying those upfront costs enqueue on top of the declared action frame via APNAP (`FR-ENG-014`).
- **Failure**: If any cost component cannot be paid in full, the transaction aborts with `ERR_INSUFFICIENT_RESOURCES` or `ERR_ASSET_EXHAUSTED` with zero partial deduction.

#### FR-ENG-014: APNAP Trigger Dispatching & Simultaneous Event Ordering
- **Functional Contract**: When an in-game event triggers multiple abilities simultaneously, the engine MUST queue them onto the stack using the **APNAP (Active Player, Non-Active Player)** convention in conformance with [`FR-TRG-003`](01_Game_Rules_FRD.md#L294) and [`FR-TRG-004`](01_Game_Rules_FRD.md#L300):
  1. All triggers controlled by the **Active Player** are pushed onto the stack first.
  2. All triggers controlled by the **Non-Active Player** are pushed onto the stack second on top of the Active Player's triggers.
- **Invariants**: Because the stack is LIFO, the Non-Active Player's triggers resolve first. Where a player controls multiple simultaneous triggers, allied triggers are ordered by their controller (falling back to deterministic blueprint ID sorting for bot/automated simulation).

#### FR-ENG-015: Two-Stage Target Verification & Fizzle Execution Contract
- **Functional Contract**: The engine MUST enforce the two-stage target verification, partial invalidation, distributed pool value evaporation, composite targeted/non-targeted clause isolation, and non-fizzling asset deployment rules codified in [`FR-TGT-001`–`FR-TGT-003`](01_Game_Rules_FRD.md#L244-L262) against the declarative `TargetRef` and `FilterDef` schemas of [`FR-ABL-007`](02_Card_Data_Pipeline_FRD.md#L297) and [`FR-TGT-101`](02_Card_Data_Pipeline_FRD.md#L413). When all targets of a targeted clause have become illegal at resolution, the engine skips that targeted clause and logs diagnostic warning `ERR_INVALID_TARGET`.

#### FR-ENG-016: Reactive Replacement Watcher Interceptor Pipeline
- **Functional Contract**: The engine MUST evaluate active `ReplacementWatcherDef` interceptors ([`FR-ABL-005`](02_Card_Data_Pipeline_FRD.md#L282)) inline on prospective events (`WOULD_TAKE_DAMAGE`, `WOULD_BE_DESTROYED`, `WOULD_DEPLOY`; [`FR-TRG-001`](02_Card_Data_Pipeline_FRD.md#L369)) *before* state mutations commit or triggers enqueue.
- **Rules Conformance**: Inline interception, affected-asset controller conflict selection, single-application tracking per discrete event ID (`FR-TRG-007`), consequential child-event isolation, sequential per-entity death replacement, and 5-stage damage mitigation precedence MUST execute in strict conformance with [`FR-REP-001`–`FR-REP-003`](01_Game_Rules_FRD.md#L348) and [`FR-REP-004`](01_Game_Rules_FRD.md#L356).

#### FR-ENG-017: Choice Prompt Yield, Fast-Path & Resumption Protocol
- **Functional Contract**: The engine distinguishes between **upfront bound command payloads** (modal indices, targets, and distributed pool allocations submitted upon declaration per [`FR-TGT-101`](02_Card_Data_Pipeline_FRD.md#L413) and [`FR-MSG-003`](04_Client_Engine_Protocol_Networking_FRD.md#L295)) and **mid-resolution interactive choice prompts** (triggered target selections, replacement conflict choices, top-of-deck inspections `LOOK_TOP_CARDS`, deck searches `SEARCH_DECK`, and revealed hand inspections `REVEAL_HAND`; [`FR-MSG-005`](04_Client_Engine_Protocol_Networking_FRD.md#L400)).
- **Execution Sequence**: Conforming to [`FR-MOD-001`](01_Game_Rules_FRD.md#L267) and [`FR-TRG-002`](01_Game_Rules_FRD.md#L291):
  1. **Single-Option Fast-Path**: If a required target or choice has only 1 legal option and does not involve inspecting hidden cards (`LOOK_TOP_CARDS`, `SEARCH_DECK`, or `REVEAL_HAND`), the engine selects that sole option automatically without yielding or emitting `PlayerChoicePrompt`.
  2. **Yield & Inspection Window**: When multiple legal options exist—or when executing a hidden/revealed card inspection (`LOOK_TOP_CARDS`, `SEARCH_DECK`, or `REVEAL_HAND`)—the engine suspends the active stack frame, transitions to `STATE_AWAITING_CHOICE`, freezes the Main Phase turn clock, and emits `PlayerChoicePrompt` with the **15-second** active-player or **10-second** inactive-player action clock ([`FR-SES-004`](04_Client_Engine_Protocol_Networking_FRD.md#L94), [`FR-SES-005`](04_Client_Engine_Protocol_Networking_FRD.md#L100)) plus the 3,000 ms network grace window ([`DA-NET-004`](04_Client_Engine_Protocol_Networking_FRD.md#L56)).
  3. **Resumption or Deterministic Fallback**: Resumes stack execution immediately upon receiving a valid `SubmitChoiceCommand` matching `prompt_id`, or applies the deterministic fallback of [`FR-MOD-001`](01_Game_Rules_FRD.md#L267) (logging `ERR_ACTION_CLOCK_EXPIRED`) if both the action clock and the player's Reserve Time Bank expire.

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
    +-------------------------------------------------------------------+
    |              25 Canonical Atomic Primitives (FR-PRM-001)          |
    |  DEAL_DAMAGE     | HEAL            | ADD_ARMOR     | MODIFY_STAT  |
    |  MOVE_ZONE       | RETURN_TO_HAND  | KILL          | DRAW_CARD    |
    |  DISCARD_CARD    | REVEAL_HAND     | SCRATCH       | REGROUP      |
    |  DEPLOY_FREE     | SEARCH_DECK     | LOOK_TOP_CARDS| TUCK_UNDER   |
    |  GAIN_CONTROL    | MORPH_SCRATCH   | EXHAUST       | GRANT_SUBTYPE|
    |  GRANT_KEYWORD   | PUT_CARD_ON_DECK| GENERATE_RESOURCE            |
    |  REDIRECT_DAMAGE_TARGET            | RECYCLE_DISCARD_TO_DECK      |
    +-------------------------------------------------------------------+
```

### Section 4.1: Ability Evaluation & Execution Engine

#### FR-ENG-018: Polymorphic Ability AST Interpreter & Type-Switch Visitor
- **Functional Contract**: The engine MUST evaluate declarative `AbilityDef` nodes using an exhaustive Go type-switch visitor dispatching across all six canonical `abilityType` archetypes (`TRIGGERED`, `ACTIVATED`, `CONTINUOUS_AURA`, `REPLACEMENT_WATCHER`, `MANEUVER_EFFECT`, and `KEYWORD_GRANT`) and composite control-flow containers (`COMPOSITE_SEQUENCE`, `CONDITIONAL_CLAUSE`) defined in [`FR-ABL-001`–`FR-ABL-006`](02_Card_Data_Pipeline_FRD.md#L222-L294) and [`FR-ABL-012`](02_Card_Data_Pipeline_FRD.md#L445).

#### FR-ENG-019: Dynamic FilterDef & Target Resolution Evaluator
- **Functional Contract**: The engine MUST evaluate `FilterDef` predicates and `TargetRef` selectors exactly as specified in [`FR-ABL-007`](02_Card_Data_Pipeline_FRD.md#L297) and [`FR-TGT-001`–`FR-TGT-002`](01_Game_Rules_FRD.md#L244-L255), performing case-insensitive subtype comparisons ([`FR-DAT-007`](02_Card_Data_Pipeline_FRD.md#L178)) and evaluating stat comparisons against Layer 6 effective attributes (`EffectivePower`, `EffectiveMaxHealth`, and `CurrentHealth`).

#### FR-ENG-020: Condition Gate & ValueToken Expression Resolver
- **Functional Contract**: The engine MUST resolve dynamic numeric expressions (`ValueToken`: `STATIC`, `DYNAMIC`, `QUERY`, `VARIABLE_X`) and boolean gates (`ConditionDef`, including `Cornered` and `CAUSALITY_CONTEXT` lineage) at runtime in strict conformance with [`FR-ABL-007`](02_Card_Data_Pipeline_FRD.md#L297), [`FR-TRG-007`](02_Card_Data_Pipeline_FRD.md#L379), and [`FR-LAY-003`](01_Game_Rules_FRD.md#L338).

#### FR-ENG-021: Atomic Effect Primitives Execution Engine
- **Functional Contract**: The engine MUST implement discrete Go execution handlers dispatching 1:1 to the canonical catalog of 25 atomic effect primitives codified in [`FR-PRM-001`](02_Card_Data_Pipeline_FRD.md#L382).
- **Safety Invariant**: Any unhandled or unrecognized primitive encountered within an ability AST payload MUST abort the transaction with structured error code `ERR_INVALID_COMMAND_ARGUMENTS`.

#### FR-ENG-022: Double-Blind Scratch Company Lifecycle Engine
- **Functional Contract**: When executing `SCRATCH` ([`FR-PRM-001`](02_Card_Data_Pipeline_FRD.md#L382), [`FR-KWD-002m`](01_Game_Rules_FRD.md#L445)) or the Headquarters inherent action ([`FR-INH-004`](01_Game_Rules_FRD.md#L223)), the engine MUST execute the double-blind deck-to-battlefield transition and zone-departure reveal rules of [`FR-ZON-005`](01_Game_Rules_FRD.md#L136), [`FR-ENT-103`](01_Game_Rules_FRD.md#L146), [`FR-SEC-001`](04_Client_Engine_Protocol_Networking_FRD.md#L164), and [`FR-SEC-002`](04_Client_Engine_Protocol_Networking_FRD.md#L171), assigning an opaque `proxy_scratch_*` `AssetID` and suppressing `cardInstanceId` / `blueprintId` from all network payloads while on the battlefield or tucked face-down in an `ATTACHED` sub-zone.

#### FR-ENG-023: Continuous Aura Grant & Revocation Engine
- **Functional Contract**: Continuous auras (`CONTINUOUS_AURA`, [`FR-ABL-004`](02_Card_Data_Pipeline_FRD.md#L271)) MUST continuously apply in Layer 4 (and Layer 3 for compound `abilityGrants`) while the host `AssetInstance` occupies a battlefield zone and immediately cease when the host leaves the battlefield.

#### FR-ENG-024: Maneuver Card Enactment & Discard Transition
- **Functional Contract**: When a Maneuver card is enacted from `HandZone`, the engine validates and deducts costs upfront (`FR-ENG-013`), pushes `FRAME_MANEUVER` onto `ResolutionStack`, executes its `ManeuverEffectDef` payload when popped, and moves the `CardInstance` into its owner's `DiscardZone` in conformance with [`FR-MNV-001`](01_Game_Rules_FRD.md#L233).

---

## DOMAIN 5: 6-LAYER ATTRIBUTE RECALCULATION PIPELINE

### Section 5.1: Stateless Layer Evaluation & Cleanup

#### FR-ENG-025: Deterministic 6-Layer Attribute Pipeline Architecture
- **Functional Contract**: The engine MUST enforce stateless, on-demand dynamic attribute evaluation across all battlefield `AssetInstance` structs in strict compliance with the canonical 6-layer pipeline, timestamp ordering, and exclusive Layer 6 Combat Power clamping (`EffectivePower = max(0, RawPower)`) codified in Section 6.2 of [`01_Game_Rules_FRD.md`](01_Game_Rules_FRD.md) ([`FR-LAY-001`](01_Game_Rules_FRD.md#L329) through [`FR-LAY-003`](01_Game_Rules_FRD.md#L338)).
- **Stateless On-Demand Evaluation**: Because battlefield entity counts are bounded and `ValueToken` / `Cornered` predicates depend on cross-zone state (`FR-LAY-003`), effective attributes are computed on-demand directly from the 6-layer pipeline without stateful `isDirty` flag caching, subject only to explicit snapshot locking during simultaneous resolution sequences ([`FR-ENG-033`](#fr-eng-033-step-3-frame_attack-resolution-simultaneous-combat-damage--retaliation)).

#### FR-ENG-026: Subsumed by FR-ENG-025
- *Dirty-flag caching (`isDirty`) and topological dependency invalidation are formally de-scoped and subsumed into stateless on-demand evaluation in [`FR-ENG-025`](#fr-eng-025-deterministic-6-layer-attribute-pipeline-architecture).*

#### FR-ENG-027: Subsumed by FR-ENG-025
- *Timestamp ordering within layers is formally subsumed into [`FR-ENG-025`](#fr-eng-025-deterministic-6-layer-attribute-pipeline-architecture).*

#### FR-ENG-028: Subsumed by FR-ENG-025 & FR-ENG-041
- *Exclusive Layer 6 Combat Power clamping (`EffectivePower = max(0, RawPower)`) is formally subsumed into [`FR-ENG-025`](#fr-eng-025-deterministic-6-layer-attribute-pipeline-architecture) and codified as System Invariant 1 in [`FR-ENG-041`](#fr-eng-041-system-invariants--diagnostic-assertion-protocol).*

#### FR-ENG-029: Round-Duration Cleanup Sub-Phase Execution
- **Functional Contract**: During Sub-Phase 4c of the End Phase, the engine MUST iterate all active `AssetInstance` structs and purge temporary Layer 5 modifiers (`END_OF_TURN` on every turn; `END_OF_OPPONENT_TURN` when the ending turn belongs to the modifier creator's opponent) per [`FR-LAY-003`](01_Game_Rules_FRD.md#L338).

---

## DOMAIN 6: COMBAT RESOLUTION & RETALIATION ENGINE

### Section 6.1: Deterministic Stack-Based Combat Resolution

#### FR-ENG-030: Combat Zone Eligibility Enforcement
- **Functional Contract**: The engine MUST enforce the spatial combat attacker and target eligibility rules codified in [`FR-CBT-001`](01_Game_Rules_FRD.md#L390), [`FR-CAP-001`](01_Game_Rules_FRD.md#L203), and [`FR-ZON-002`](01_Game_Rules_FRD.md#L125) (restricting attackers to Main Theater Units and HQ Zone Headquarters, permitting attacks with `EffectivePower <= 0`, and protecting Reserves Zone assets and `Ethereal` assets from attack declarations).

#### FR-ENG-031: Step 1: Attack Declaration, Upfront Exhaustion & `FRAME_ATTACK` Push
- **Functional Contract**: Upon receiving `DeclareAttackCommand`, the engine MUST validate clean-stack priority (`ResolutionStack.IsEmpty() == true`, [`FR-QUE-001`](01_Game_Rules_FRD.md#L284)), `Readiness == READIED`, and `Guardian` target priority ([`FR-KWD-002j`](01_Game_Rules_FRD.md#L442); aborting with `ERR_GUARDIAN_PRIORITY_VIOLATION` if bypassed). If the attacking unit possesses `Berserk` ([`FR-KWD-002o`](01_Game_Rules_FRD.md#L447)), the command MAY bind an optional secondary opposing Main Theater target upfront (`TargetRole: SECONDARY_BERSERK_TARGET`, [`FR-TGT-101`](02_Card_Data_Pipeline_FRD.md#L413)). Upon validation, the engine atomically transitions the attacker to `EXHAUSTED` (`FR-ENG-013`), pushes a root `FRAME_ATTACK` onto `ResolutionStack` (`FR-ENG-012`), and emits `EVENT_EXHAUST` and `EVENT_ATTACK_DECLARE` per [`FR-CBT-002`](01_Game_Rules_FRD.md#L405).

#### FR-ENG-032: Step 2: Pre-Combat Trigger Enqueue & LIFO Resolution
- **Functional Contract**: Conforming to Step 2 of [`FR-CBT-002`](01_Game_Rules_FRD.md#L405), all reactive triggers responding to `EVENT_EXHAUST` and `EVENT_ATTACK_DECLARE` enqueue **on top of `FRAME_ATTACK`** via APNAP ([`FR-ENG-014`](#fr-eng-014-apnap-trigger-dispatching--simultaneous-event-ordering)) and auto-resolve top-down before `FRAME_ATTACK` pops.
- **Combatant Survival & Zone Legality Check**: When `FRAME_ATTACK` pops from the stack at the conclusion of Step 2, if either the **attacking asset** (e.g., destroyed by *Blind Defile*) or the **target asset** was destroyed, removed from the battlefield, or moved out of a legal combat zone during Step 2, `FRAME_ATTACK` fizzles immediately with zero combat or retaliation damage dealt ([`FR-CBT-002`](01_Game_Rules_FRD.md#L405)).

#### FR-ENG-033: Step 3: `FRAME_ATTACK` Resolution: Simultaneous Combat Damage & Retaliation
- **Functional Contract**: When `FRAME_ATTACK` pops with both attacker and defender surviving in legal combat zones, the engine MUST snapshot and lock continuous aura condition gates (`Cornered`; [`FR-LAY-003`](01_Game_Rules_FRD.md#L338)) and resolve simultaneous combat damage (`Attacker.EffectivePower`) and 100% symmetric retaliation (`Defender.EffectivePower`) across all asset combinations in strict conformance with Step 3 of [`FR-CBT-002`](01_Game_Rules_FRD.md#L405).

#### FR-ENG-034: Canonical 5-Stage Damage Mitigation Pipeline
- **Functional Contract**: All incoming combat and effect damage instances MUST route sequentially through the canonical 5-stage mitigation pipeline (**Stage 1: Target Redirection** &rarr; **Stage 2: Flat `Resistant N` Modification** &rarr; **Stage 3: `Stalwart` Combat Damage Capping** &rarr; **Stage 4: Dynamic `Armor` Absorption** &rarr; **Stage 5: Health Commit & `EVENT_TAKE_DAMAGE` Emission**) codified in [`FR-REP-004`](01_Game_Rules_FRD.md#L356). Conforming to Stage 5 of `FR-REP-004` and [`FR-KWD-002a`](01_Game_Rules_FRD.md#L431), the engine emits `EVENT_TAKE_DAMAGE` (and triggers `Unbowed`) **if and only if** `(armorAbsorbed + healthDamageCommitted) > 0`.

#### FR-ENG-035: State-Based Death Sweep, `Berserk` Trample & Discard Transition Batching
- **Functional Contract**: Immediately following damage commitment, the engine MUST execute a state-based death sweep in conformance with [`FR-LAY-002`](01_Game_Rules_FRD.md#L333), [`FR-ZON-004`](01_Game_Rules_FRD.md#L131), [`FR-ENT-102`](01_Game_Rules_FRD.md#L142), and [`FR-CBT-002`](01_Game_Rules_FRD.md#L405):
  1. **Death Check & Destruction Replacement**: Identifies all battlefield `AssetInstance` structs with `CurrentHealth <= 0` (`AccumulatedDamage >= EffectiveMaxHealth`) or `EffectiveMaxHealth <= 0`. Evaluates prospective `WOULD_BE_DESTROYED` replacement watchers sequentially per entity ([`FR-REP-001`](01_Game_Rules_FRD.md#L348), [`FR-ENG-016`](#fr-eng-016-reactive-replacement-watcher-interceptor-pipeline)).
  2. **Snapshot Capture, Attached Cleanup & Discard Transition**: For each un-replaced destroyed asset, captures `AssetSnapshot` ([`FR-SNP-101`](#fr-snp-101-assetsnapshot-last-known-state-snapshot-struct)), routes any tucked `AttachedCards` face-up to their owners' `DiscardZone` (`EVENT_DISCARD_CARD`, [`FR-ENT-102`](01_Game_Rules_FRD.md#L142)), moves the destroyed card face-up to its owner's `DiscardZone`, and emits `EVENT_DESTROY` (and `EVENT_COMBAT_KILL` if destroyed by combat damage).
  3. **`Berserk` Excess Damage & Recursive Death Sweep**: If an attacking `Berserk` unit ([`FR-KWD-002o`](01_Game_Rules_FRD.md#L447)) destroyed its primary Main Theater target during Step 3 combat damage and a legal pre-bound `SECONDARY_BERSERK_TARGET` (`FR-ENG-031`) remains in `MainTheaterZone`, the engine routes the excess combat damage as non-combat effect damage through `FR-ENG-034` against that secondary target and immediately executes a recursive state-based death sweep before emitting `EVENT_ATTACK_RESOLVED` and queueing all resulting triggers via APNAP (`FR-ENG-014`).

#### FR-ENG-042: Authoritative Match Conclusion & Victory Condition Engine
- **Functional Contract**: The engine MUST continuously evaluate match termination conditions during state-based sweeps and draw operations:
  1. **Headquarters Destruction**: When a player's Headquarters `AssetInstance` reaches `CurrentHealth <= 0` (`AccumulatedDamage >= EffectiveMaxHealth`), that player immediately suffers defeat ([`FR-CBT-003`](01_Game_Rules_FRD.md#L422)).
  2. **Simultaneous Defeat & Match Draw**: If both Headquarters reach `CurrentHealth <= 0` within the exact same simultaneous damage step, the match terminates immediately as a **Match Draw** (`RESULT_MATCH_DRAW`, [`FR-CBT-003`](01_Game_Rules_FRD.md#L422)).
  3. **Empty Deck Defeat**: When a player is required to draw a card from their `DeckZone` and `count(DeckZone) == 0`, that player immediately suffers defeat ([`FR-WIN-001`](01_Game_Rules_FRD.md#L73)).
  4. **Match Concluded Transition**: Upon any terminal condition, the engine emits `MatchConcludedEvent` ([`FR-PAC-003`](04_Client_Engine_Protocol_Networking_FRD.md#L430)), halts the transaction tick loop, transitions `MatchState` to immutable `STATE_CONCLUDED`, and flushes final `.rec` replay telemetry.

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
- **Functional Contract**: The engine MUST record all committed transactions, inputs, and seeds into the binary/JSON `.rec` match replay format specified in [`FR-SPC-002`](04_Client_Engine_Protocol_Networking_FRD.md#L735).
- **Validation Mode**: Given a `.rec` file, the engine MUST re-simulate the entire match from the initial seed and assert that every intermediate `StateChecksum` matches bit-for-bit.

#### FR-ENG-039: Combinatorial Monkey/Fuzz Testing Runner
- **Functional Contract**: The CLI suite MUST include an automated monkey testing runner executing randomized, legal player actions across random seeds:
  - Detects infinite trigger loops, stack overflows, and invariant violations (`FR-ENG-041`).
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
The headless engine module enforces zero-drift error parity with the client-engine networking protocol. All transaction rollbacks and validation failures emit structured error codes conforming directly to [`FR-ERR-001`](04_Client_Engine_Protocol_Networking_FRD.md#L745):

| Error Code | Severity | Trigger Condition | System Action | Wire Traceability |
| :--- | :--- | :--- | :--- | :--- |
| `ERR_STACK_NOT_EMPTY` | Error | Manual action attempted while stack non-empty (`FR-QUE-001`, `FR-INH-001`). | Abort transaction; rollback delta. | [`FR-ERR-001`](04_Client_Engine_Protocol_Networking_FRD.md#L746) |
| `ERR_INSUFFICIENT_RESOURCES` | Error | Player lacks Requisition or Tactics to pay cost (`FR-RES-001`). | Abort transaction; rollback delta. | [`FR-ERR-001`](04_Client_Engine_Protocol_Networking_FRD.md#L747) |
| `ERR_INVALID_TARGET` | Warning | Target became illegal or moved zones prior to stack pop (`FR-TGT-001`). | Fizzle targeted clause; log diagnostic. | [`FR-ERR-001`](04_Client_Engine_Protocol_Networking_FRD.md#L748) |
| `ERR_GUARDIAN_PRIORITY_VIOLATION` | Error | Combat attack bypassed an eligible opposing Guardian in Main Theater (`FR-KWD-002j`). | Abort transaction; flash Guardian UI. | [`DA-NET-005`](04_Client_Engine_Protocol_Networking_FRD.md#L749) |
| `ERR_INVALID_COMMAND_ARGUMENTS` | Error | Command arguments malformed or out of bounds (`FR-MSG-003`). | Abort transaction; rollback delta. | [`DA-NET-003`](04_Client_Engine_Protocol_Networking_FRD.md#L750) |
| `ERR_ASSET_EXHAUSTED` | Error | Attempted action requiring Readied state (`FR-INH-001`, `FR-CAP-001`). | Abort transaction; rollback delta. | [`FR-ERR-001`](04_Client_Engine_Protocol_Networking_FRD.md#L751) |
| `ERR_INVALID_RESOURCE_CHOICE` | Error | Invalid resource generation for current zone or keyword (`FR-INH-002`). | Abort transaction; rollback delta. | [`FR-ERR-001`](04_Client_Engine_Protocol_Networking_FRD.md#L752) |
| `ERR_ACTION_CLOCK_EXPIRED` | Warning | Choice prompt timed out without player selection (`FR-TRG-002`). | Apply deterministic fallback (`FR-MOD-001`). | [`FR-ERR-001`](04_Client_Engine_Protocol_Networking_FRD.md#L753) |
| `ERR_DESYNC_DETECTED` | Fatal | StateChecksum mismatch between client and server (`FR-PAC-004`). | Terminate match; emit replay dump. | [`FR-PAC-004`](04_Client_Engine_Protocol_Networking_FRD.md#L690) |
| `ERR_INFINITE_LOOP` | Fatal | Stack execution depth exceeds 100 frames. | Force-clear stack; emit fatal log. | [`FR-ERR-001`](04_Client_Engine_Protocol_Networking_FRD.md#L745) |

### Section 8.2: Parity & State Invariants
#### FR-ENG-041: System Invariants & Diagnostic Assertion Protocol
- **Invariant 1 (Exclusive Layer 6 Clamping & Signed Attributes)**: Conforming to [`FR-LAY-001`](01_Game_Rules_FRD.md#L329), [`FR-LAY-002`](01_Game_Rules_FRD.md#L333), and [`FR-CST-001`](01_Game_Rules_FRD.md#L196), Layer 6 `EffectivePower`, effective resource costs, `CurrentArmor`, and `AccumulatedDamage` MUST never evaluate to negative integers ($\ge 0$). `Resistant` is signed, and any battlefield `AssetInstance` reaching `EffectiveMaxHealth <= 0` or `CurrentHealth <= 0` MUST be destroyed immediately during the next state-based death sweep (`FR-ENG-035`).
- **Invariant 2 (Zone Conservation)**: Every `CardInstance` MUST reside in exactly one zone at any given moment, and an `AssetInstance` wrapper MUST exist if and only if the card occupies `HeadquartersZone`, `ReservesZone`, or `MainTheaterZone`.
- **Invariant 3 (Clean Stack at Phase & Turn Boundaries)**: Every phase transition and turn handover MUST assert `ResolutionStack.IsEmpty() == true`.

---

## MASTER CROSS-REFERENCE REGISTER

| Requirement ID | Scope Summary | Interoperability Target |
| :--- | :--- | :--- |
| `FR-ENT-001` | `CardBlueprint` immutable runtime singleton | [`FR-CRD-101`](01_Game_Rules_FRD.md#L82), [`FR-DAT-005`](02_Card_Data_Pipeline_FRD.md#L107) |
| `FR-ENT-002` | `CardInstance` session card struct | [`FR-ZON-003`](01_Game_Rules_FRD.md#L102), [`FR-ZON-004`](01_Game_Rules_FRD.md#L131) |
| `FR-ENT-003` | `AssetInstance` battlefield entity struct & attached cleanup | [`FR-ZON-003`](01_Game_Rules_FRD.md#L102), [`FR-ENT-102`](01_Game_Rules_FRD.md#L142), [`FR-LAY-002`](01_Game_Rules_FRD.md#L333) |
| `FR-SNP-101` | `AssetSnapshot` last-known battlefield state struct | [`FR-ZON-004`](01_Game_Rules_FRD.md#L131), [`FR-TRG-007`](02_Card_Data_Pipeline_FRD.md#L379) |
| `FR-ENG-001` | Pure Go engine modularity & zero graphics dependency | [`FR-ENT-001`](#fr-ent-001-cardblueprint-immutable-runtime-singleton), [`FR-SES-001`](04_Client_Engine_Protocol_Networking_FRD.md#L68) |
| `FR-ENG-002` | Isolated multi-stream PRNG generator & seed derivation | [`FR-SET-002`](01_Game_Rules_FRD.md#L48), [`FR-SET-003`](01_Game_Rules_FRD.md#L44) |
| `FR-ENG-003` | Deterministic seed derivation with cryptographic salts | Subsumed by [`FR-ENG-002`](#fr-eng-002-isolated-multi-stream-prng-architecture--seed-derivation) |
| `FR-ENG-004` | Pure discrete integer arithmetic & domain clamping | [`FR-LAY-001`](01_Game_Rules_FRD.md#L329), [`FR-REP-004`](01_Game_Rules_FRD.md#L356) |
| `FR-ENG-005` | Monotonic tick counter & 64-bit FNV-1a state checksum | [`FR-PAC-004`](04_Client_Engine_Protocol_Networking_FRD.md#L690) |
| `FR-ENG-006` | Transactional event-emitting match state model | [`FR-ENT-001`](#fr-ent-001-cardblueprint-immutable-runtime-singleton), [`FR-ZON-001`](01_Game_Rules_FRD.md#L112) |
| `FR-ENG-007` | Pre-game HQ reveal, first-player choice & 4-step mulligan | [`FR-SET-001`](01_Game_Rules_FRD.md#L28), [`FR-SET-002`](01_Game_Rules_FRD.md#L48), [`FR-SET-003`](01_Game_Rules_FRD.md#L44), [`FR-SES-004`](04_Client_Engine_Protocol_Networking_FRD.md#L94) |
| `FR-ENG-008` | Transaction checkpoint & delta-rollback engine with PRNG restore | [`FR-QUE-001`](01_Game_Rules_FRD.md#L284), [`FR-ERR-001`](04_Client_Engine_Protocol_Networking_FRD.md#L745) |
| `FR-ENG-009` | Turn sub-phase transitions (Ready, Draw, Main, End 4a–4d) | [`FR-TRN-001`](01_Game_Rules_FRD.md#L170), [`FR-SET-004`](01_Game_Rules_FRD.md#L57) |
| `FR-ENG-010` | Clean-stack constraint enforcement & inherent actions | [`FR-QUE-001`](01_Game_Rules_FRD.md#L284), [`FR-CAP-001`](01_Game_Rules_FRD.md#L203), [`FR-INH-001`](01_Game_Rules_FRD.md#L223), [`FR-RES-003`](01_Game_Rules_FRD.md#L163) |
| `FR-ENG-011` | Turn-clock timeout & Reserve Time Bank exhaustion | [`FR-TRN-001`](01_Game_Rules_FRD.md#L170), [`FR-SES-004`](04_Client_Engine_Protocol_Networking_FRD.md#L94), [`FR-SES-005`](04_Client_Engine_Protocol_Networking_FRD.md#L100) |
| `FR-ENG-012` | LIFO resolution stack & frame taxonomy | [`FR-QUE-001`](01_Game_Rules_FRD.md#L284) |
| `FR-ENG-013` | Atomic cost validation, reservation & deduction | [`FR-CST-001`](01_Game_Rules_FRD.md#L196), [`FR-TRN-003a`](01_Game_Rules_FRD.md#L233), [`FR-ABL-007`](02_Card_Data_Pipeline_FRD.md#L297) |
| `FR-ENG-014` | APNAP simultaneous trigger dispatching order | [`FR-TRG-003`](01_Game_Rules_FRD.md#L294), [`FR-TRG-004`](01_Game_Rules_FRD.md#L300) |
| `FR-ENG-015` | Two-stage target verification, evaporation & fizzle | [`FR-TGT-001`](01_Game_Rules_FRD.md#L244) – [`FR-TGT-003`](01_Game_Rules_FRD.md#L256) |
| `FR-ENG-016` | Reactive replacement watcher interceptor pipeline | [`FR-REP-001`](01_Game_Rules_FRD.md#L348) – [`FR-REP-004`](01_Game_Rules_FRD.md#L356) |
| `FR-ENG-017` | Choice prompt fast-path, inspection windows & action clocks | [`FR-MOD-001`](01_Game_Rules_FRD.md#L267), [`FR-MSG-005`](04_Client_Engine_Protocol_Networking_FRD.md#L400) |
| `FR-ENG-018` | Polymorphic ability AST type-switch visitor | [`FR-ABL-001`](02_Card_Data_Pipeline_FRD.md#L222) – [`FR-ABL-006`](02_Card_Data_Pipeline_FRD.md#L290) |
| `FR-ENG-019` | Dynamic filter evaluator against effective attributes | [`FR-ABL-007`](02_Card_Data_Pipeline_FRD.md#L297), [`FR-TGT-001`](01_Game_Rules_FRD.md#L244) |
| `FR-ENG-020` | ValueToken expression resolver & condition gate | [`FR-ABL-007`](02_Card_Data_Pipeline_FRD.md#L297), [`FR-TRG-007`](02_Card_Data_Pipeline_FRD.md#L379) |
| `FR-ENG-021` | Canonical atomic effect primitives execution engine | [`FR-PRM-001`](02_Card_Data_Pipeline_FRD.md#L382) |
| `FR-ENG-022` | Double-blind Scratch Company deck-to-battlefield lifecycle | [`FR-ZON-005`](01_Game_Rules_FRD.md#L136), [`FR-SEC-001`](04_Client_Engine_Protocol_Networking_FRD.md#L164) |
| `FR-ENG-023` | Continuous aura registration & revocation engine | [`FR-ABL-004`](02_Card_Data_Pipeline_FRD.md#L271) |
| `FR-ENG-024` | Maneuver card resolution & Discard transition | [`FR-MNV-001`](01_Game_Rules_FRD.md#L233) |
| `FR-ENG-025` | Stateless 6-layer attribute calculation pipeline execution | [`FR-LAY-001`](01_Game_Rules_FRD.md#L329) – [`FR-LAY-003`](01_Game_Rules_FRD.md#L338) |
| `FR-ENG-026` | Dirty-flag caching & topological dependency invalidation | Subsumed by [`FR-ENG-025`](#fr-eng-025-deterministic-6-layer-attribute-pipeline-architecture) |
| `FR-ENG-027` | Timestamp precedence within attribute calculation layers | Subsumed by [`FR-ENG-025`](#fr-eng-025-deterministic-6-layer-attribute-pipeline-architecture) |
| `FR-ENG-028` | Effective attribute clamping and non-negative invariants | Subsumed by [`FR-ENG-025`](#fr-eng-025-deterministic-6-layer-attribute-pipeline-architecture) & [`FR-ENG-041`](#fr-eng-041-system-invariants--diagnostic-assertion-protocol) |
| `FR-ENG-029` | Round-duration temporary modifier cleanup | [`FR-LAY-003`](01_Game_Rules_FRD.md#L338) |
| `FR-ENG-030` | Spatial zone combat restrictions (Main Theater and HQ Zone) | [`FR-CBT-001`](01_Game_Rules_FRD.md#L390), [`FR-CAP-001`](01_Game_Rules_FRD.md#L203) |
| `FR-ENG-031` | Step 1 attack declaration, Berserk target binding & readiness lock | [`FR-CBT-002`](01_Game_Rules_FRD.md#L405), [`FR-KWD-002o`](01_Game_Rules_FRD.md#L447), [`FR-TGT-101`](02_Card_Data_Pipeline_FRD.md#L413) |
| `FR-ENG-032` | Step 2 reaction window & trigger queue | [`FR-CBT-002`](01_Game_Rules_FRD.md#L405) |
| `FR-ENG-033` | Step 3 simultaneous combat damage & symmetric retaliation | [`FR-CBT-002`](01_Game_Rules_FRD.md#L405) |
| `FR-ENG-034` | Canonical 5-stage damage mitigation pipeline & damage-taken gate | [`FR-REP-004`](01_Game_Rules_FRD.md#L356), [`FR-KWD-002a`](01_Game_Rules_FRD.md#L431) |
| `FR-ENG-035` | State-based death check sweep, Berserk trample & Discard transition | [`FR-ENT-003`](#fr-ent-003-assetinstance-battlefield-entity-struct), [`FR-SNP-101`](#fr-snp-101-assetsnapshot-last-known-state-snapshot-struct), [`FR-CBT-002`](01_Game_Rules_FRD.md#L405) |
| `FR-ENG-036` | Headless engine CLI test runner architecture | [`FR-SES-001`](04_Client_Engine_Protocol_Networking_FRD.md#L68) |
| `FR-ENG-037` | Scripted scenario fixture DSL & assertion framework | [`FR-PAC-004`](04_Client_Engine_Protocol_Networking_FRD.md#L690) |
| `FR-ENG-038` | Deterministic .rec match replay recording & validation | [`FR-SPC-002`](04_Client_Engine_Protocol_Networking_FRD.md#L735) |
| `FR-ENG-039` | Combinatorial monkey/fuzz testing orchestrator | [`FR-PAC-004`](04_Client_Engine_Protocol_Networking_FRD.md#L690) |
| `FR-ENG-040` | Engine performance benchmarks & throughput gates | [`FR-ENG-001`](#fr-eng-001-engine-modularity--zero-engine-dependency-invariant) |
| `FR-ENG-041` | System invariants & diagnostic assertions | [`FR-ENT-003`](#fr-ent-003-assetinstance-battlefield-entity-struct), [`FR-LAY-001`](01_Game_Rules_FRD.md#L329), [`FR-QUE-001`](01_Game_Rules_FRD.md#L284) |
| `FR-ENG-042` | Authoritative match conclusion & victory condition engine | [`FR-WIN-001`](01_Game_Rules_FRD.md#L73), [`FR-CBT-003`](01_Game_Rules_FRD.md#L422), [`FR-PAC-003`](04_Client_Engine_Protocol_Networking_FRD.md#L430) |
