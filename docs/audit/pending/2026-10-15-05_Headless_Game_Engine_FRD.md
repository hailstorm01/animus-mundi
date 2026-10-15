# Product Manager Devil's Advocate Audit: Headless Game Engine FRD

- **Target Document:** [docs/frd/05_Headless_Game_Engine_FRD.md](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/05_Headless_Game_Engine_FRD.md)
- **Auditor:** PM Devil's Advocate (`pm-devils-advocate`)
- **Evaluation Date:** 2026-10-15
- **Status:** PENDING (Active review queue)
- **Verdict:** REVISE WITH BLOCKERS

---

## 1. Executive Summary & Red-Team Assessment

The **Headless Game Engine Functional Requirements Document (Master FRD v1.0)** specifies the core rules execution kernel for Card Town. Built as a pure Go module designed for server-authoritative hosting on GCP Agones and local offline execution, the document presents strong architectural foundations in its zero-graphics dependency isolation, multi-stream PRNG design, discrete integer arithmetic, and transactional delta-memento rollback concepts.

However, a rigorous red-team interrogation against the foundational rules ([`01_Game_Rules_and_Card_Modeling_FRD.md`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md)), card compiler pipeline ([`03_Card_Compiler_FRD.md`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/03_Card_Compiler_FRD.md)), and network protocol ([`04_Client_Engine_Protocol_Networking_FRD.md`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/04_Client_Engine_Protocol_Networking_FRD.md)) reveals critical design regressions and severe specification gaps that make the engine unexecutable in its current form:

1. **Catastrophic Rules Inversion in Damage Mitigation**: `FR-ENG-034` evaluates Armor *before* Stalwart and Resistant, directly inverting the canonical 5-stage pipeline in `FR-REP-004` and destroying the intentional design synergy of defensive cards.
2. **Double-Blind Fog-of-War Destruction**: `FR-ENG-022` defines Scratch Company units as generic summoned tokens deleted upon leaving the battlefield, directly violating the double-blind deck-to-battlefield lifecycle (`FR-ZON-005`, `FR-ENT-101`, `FR-SEC-001`), breaking bounce spells, graveyard triggers, and deck exhaustion rules.
3. **Missing Terminal Match State Machine**: The engine contains zero requirements specifying how a match is won, lost, or drawn. Headquarters destruction and empty deck fatigue loss are completely missing from the state machine.
4. **Card Catalog Execution Failure**: `FR-ENG-021` limits atomic effect primitives to 14 arbitrarily renamed handlers, completely omitting 11 canonical primitives (including `RETURN_TO_HAND`, `SEARCH_DECK`, `TUCK_UNDER`, and `ADD_ARMOR`), guaranteeing that compiled card blueprints (`blueprints.json`) cannot execute.
5. **Alien Mechanic Infiltration**: The specification hallucinates Magic: The Gathering conventions that do not exist in Card Town, including "summoning sickness", MTG-style LIFO priority pass rounds, and a 10-card hand discard phase that directly violates the universal uncapped hand size rule (`FR-ZON-001`).

The specification requires immediate structural revisions before implementation can proceed.

---

## 2. Severity-Ordered Punch-List

### 🔴 Blocker / Critical

- **[DA-ENG-001] Inverted 5-Stage Damage Mitigation Pipeline & Defensive Keyword Precedence**
  - **Requirement Citations:** `FR-ENG-034` (contradicts [`FR-REP-004`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L455), [`FR-KWD-002c`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L749), [`FR-KWD-002k`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L757))
  - **Vulnerability / Flaw:** `FR-ENG-034` orders defensive keyword calculation as: (1) Armor N, (2) Stalwart, (3) Resistant N, (4) Base Defense. However, `01_Game_Rules_and_Card_Modeling_FRD.md` `FR-REP-004` explicitly codifies the 5-Stage Damage Mitigation Pipeline with an intentional, non-negotiable design invariant:
    - Stage 1: Target Redirection
    - Stage 2: Flat Reductions (`Resistant N`; negatives amplify)
    - Stage 3: Combat Damage Capping (`Stalwart` caps combat damage strictly to 1)
    - Stage 4: Dynamic Armor (deducts 1-for-1 from current Armor pool)
    - Stage 5: Health Counter Commit
    `FR-REP-004` explicitly specifies: *"When a target possesses both Stalwart and Armor, Stalwart caps incoming combat damage to 1 in Stage 3, and Stage 4 absorbs that 1 damage by depleting 1 point from the dynamic Armor pool."* Furthermore: *"Stage 2 flat reductions executing prior to Stage 3 combat capping is intentional by design to prevent combining Resistant and Stalwart into complete combat damage invulnerability."*
  - **Exploit / Failure Scenario:**
    1. A player controls a defensive unit with 3 dynamic Armor and Stalwart.
    2. An opponent attacks with a 6 Power unit.
    3. Under `FR-ENG-034`, Armor is evaluated first: Armor absorbs 3 damage, wiping the Armor pool to 0. Remaining damage is 3. Stalwart then caps the remaining 3 damage to 1. The unit takes 1 damage to Health. Result: Armor completely gone, 1 Health lost.
    4. Under canonical `FR-REP-004`, Stalwart evaluates in Stage 3, capping the 6 incoming combat damage to 1. Dynamic Armor evaluates in Stage 4, absorbing that 1 damage and reducing Armor from 3 to 2. Health damage is 0. Result: Unit survives with 2 Armor remaining and 0 Health lost.
    5. This inverted evaluation severely nerfs defensive units, creates bitwise simulation desyncs against client presentation prediction (`FR-PAC-004`), and breaks fundamental game balance.
  - **Remedy:** Rewrite `FR-ENG-034` to strictly match the canonical 5-stage pipeline order of `FR-REP-004`: Stage 1 (Redirection) &rarr; Stage 2 (Resistant N flat reductions/amplifications) &rarr; Stage 3 (Stalwart combat damage capping to max 1) &rarr; Stage 4 (Dynamic Armor 1-to-1 absorption) &rarr; Stage 5 (Health damage commit & `EVENT_TAKE_DAMAGE` emission).

- **[DA-ENG-002] Destruction of Double-Blind Fog-of-War via Tokenized Scratch Company Model**
  - **Requirement Citations:** `FR-ENG-022`, `FR-ENG-021` (contradicts [`FR-ZON-005`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L196), [`FR-ENT-101`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L196), [`FR-INH-004`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L283), [`FR-SEC-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/04_Client_Engine_Protocol_Networking_FRD.md#L164), [`FR-SEC-002`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/04_Client_Engine_Protocol_Networking_FRD.md#L171))
  - **Vulnerability / Flaw:** `FR-ENG-022` defines `DEPLOY_SCRATCH` as spawning an unexhausted token instance from a `neutral_scratch_company` template, stating: *"Token instances leaving the battlefield (destruction, bounce) are deleted permanently and never enter the graveyard deck."* This completely misunderstands Card Town's flagship Scratch mechanic. Per `FR-ZON-005`, `FR-ENT-101`, and `FR-SEC-001`, Scratch Companies are NOT tokens summoned from the ether; they are a direct **deck-to-battlefield move** (`DECK` &rarr; `RESERVES`) where the top card of the player's deck is deployed face-down under an opaque proxy ID, double-blind to BOTH players. When a Scratch Company unit leaves the battlefield to Discard, its underlying card identity is revealed face-up to both players. When returned to hand or deck, it reverts to that player's hand as its real printed card (`FR-ENT-103`). If the deck is empty, Deploy Scratch is a safe no-op (deploys 0 units).
  - **Exploit / Failure Scenario:**
    1. A player casts a bounce maneuver (*Disorienting Pulse*) targeting an opponent's Scratch Company (which secretly represents an unrevealed high-impact card from the opponent's deck).
    2. Under `FR-ENG-022`, the engine permanently deletes the token. The player never receives their card into hand, permanently destroying card advantage.
    3. When Scratch units die, they are deleted from memory rather than routed to Discard, completely breaking graveyard triggers (*Solarheart Monk*, *Eerie Remnant*) and graveyard recycling (`RECYCLE_DISCARD_TO_DECK`).
    4. An HQ can continuously deploy Scratch units from an empty deck indefinitely because it does not draw from the deck, creating an infinite blocker stall engine.
  - **Remedy:** Replace the token instantiation contract in `FR-ENG-022` with the canonical double-blind deck-to-battlefield proxy lifecycle: pop top card from `DeckZone`, wrap in `AssetInstance` with face-down proxy state (`1/1/1/1 Infantry Scratch Company`), route to Discard face-up on death, and revert to owner's hand as printed base card on bounce (`FR-ZON-005`, `FR-ENT-103`). Reaffirm safe no-op execution on an empty deck (`FR-INH-004`).

- **[DA-ENG-003] Omission of Core Match Win/Loss/Draw Conditions & Concluded State Machine**
  - **Requirement Citations:** `FR-ENG-006`, `FR-ENG-009`, `FR-ENG-035` (contradicts [`FR-WIN-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L75), [`FR-CBT-003`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L523), [`FR-PAC-003`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/04_Client_Engine_Protocol_Networking_FRD.md#L430))
  - **Vulnerability / Flaw:** The engine specification contains ZERO requirements governing how a match terminates or declares a winner. There is no state transition to `STATE_CONCLUDED`, no check for Headquarters destruction (`FR-CBT-003`), no check for empty deck draw exhaustion (`FR-WIN-001`), and no handling for simultaneous defeat leading to a Match Draw (`FR-CBT-003`, `RESULT_MATCH_DRAW`).
  - **Exploit / Failure Scenario:**
    1. An attacking unit deals lethal combat damage to the opponent's Headquarters.
    2. In Step 3 state-based death sweep (`FR-ENG-035`), the HQ reaches `AccumulatedDamage >= EffectiveDefense` and is transitioned to `GraveyardZone`.
    3. Because there is no terminal match condition evaluator, the engine increments `TurnNumber`, toggles `ActivePlayer`, and continues running the match loop with an empty HQ zone.
    4. Automated monkey/fuzz testing (`FR-ENG-039`) runs aimlessly until hitting the 100-turn timeout, failing every benchmark suite.
  - **Remedy:** Add explicit match termination requirements into Domain 2 and Domain 6. Codify continuous state-based win/loss checks:
    1. *Headquarters Destruction*: When an HQ reaches Health <= 0 (`FR-CBT-003`), its controller loses immediately. If both HQs reach <= 0 in the same atomic frame tick, match concludes as `RESULT_MATCH_DRAW`.
    2. *Empty Deck Fatigue*: Drawing a card from an empty `DeckZone` emits `EVENT_MATCH_CONCLUDED` with `EMPTY_DECK_DRAW` defeat (`FR-WIN-001`).
    3. *State Finalization*: Emit `MatchConcludedEvent` and transition `MatchState` to immutable `STATE_CONCLUDED`.

- **[DA-ENG-004] Missing Atomic Primitives & Incompatible Primitive Nomenclature**
  - **Requirement Citations:** `FR-ENG-021` (contradicts [`FR-PRM-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L304), [`FR-CMP-014`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/03_Card_Compiler_FRD.md#L352))
  - **Vulnerability / Flaw:** `FR-ENG-021` claims to implement `FR-PRM-001`, but arbitrarily truncates the catalog to "14 atomic primitives" while renaming half of them. `FR-PRM-001` in the master Game Rules FRD establishes 25 canonical primitives. The engine specification completely omits:
    - `RETURN_TO_HAND` (bounce spells like *Disorienting Pulse*, *Icewind Regent*, *Ambush Tactics* cannot execute)
    - `TUCK_UNDER` (*NO-AH-5* tucking mechanic cannot execute)
    - `SEARCH_DECK` (deck tutoring cannot execute)
    - `LOOK_TOP_CARDS` (scry / top deck inspection cannot execute)
    - `GAIN_CONTROL` (*Rhapsody's Hollow* mind control cannot execute)
    - `MORPH_SCRATCH` (*Forced Mutation* face-down transformation cannot execute)
    - `ADD_ARMOR` (dynamic armor increments cannot execute)
    - `RECYCLE_DISCARD_TO_DECK` (graveyard reshuffle cannot execute)
    - `REDIRECT_DAMAGE_TARGET` (*Spore Cloud* damage interception cannot execute)
    - `PUT_CARD_ON_DECK` (top/bottom deck ordering cannot execute)
    - `REVEAL_HAND` (hand reveal and targeted discard cannot execute)
    - `DEPLOY_FREE` (free in-resolution deployments cannot execute)
    Furthermore, renaming canonical primitives (`DEPLOY_SCRATCH` vs `SCRATCH`, `READY_ASSET` vs `REGROUP`, `DESTROY_ASSET` vs `KILL`, `ADD_REQUISITION` / `ADD_TACTICS` vs `GENERATE_RESOURCE`) ensures that AST payloads compiled by `compile_blueprints.ps1` (`03_Card_Compiler_FRD.md`) will crash the Go type-switch visitor on every card.
  - **Exploit / Failure Scenario:**
    1. A player plays *Icewind Regent* with ability `{ "primitive": "RETURN_TO_HAND", "target": "CHOSEN_ENEMY_UNIT" }`.
    2. The AST visitor in `FR-ENG-018` matches `RETURN_TO_HAND`, but finding no handler in `FR-ENG-021`, panics or aborts the transaction with an unknown primitive error.
    3. Over 40% of the cards in the catalog fail to execute in the engine.
  - **Remedy:** Align `FR-ENG-021` 1:1 with the canonical 25 primitives defined in `FR-PRM-001` of `01_Game_Rules_and_Card_Modeling_FRD.md`, matching identical string enums and parameter schemas.

---

### 🟠 High

- **[DA-ENG-005] Phantom "Summoning Sickness" & Haste Conflation**
  - **Requirement Citations:** `FR-ENG-031` (contradicts [`FR-SET-004`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L57), [`FR-ZON-002`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L180), [`FR-KWD-002f`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L752))
  - **Vulnerability / Flaw:** `FR-ENG-031` Step 1 gate 3 specifies: *"Attacker is not constrained by summoning sickness (unless possessing Aggressive N, FR-KWD-002f)."* Card Town explicitly has NO summoning sickness. `FR-SET-004` explicitly states: *"All action declarations, deployments, and combat declarations are permitted on Turn 1."* And `FR-ZON-002` confirms: *"In accordance with the absence of summoning sickness (FR-SET-004), a Readied stolen unit may legally declare attacks or activate abilities on the same turn it is acquired."* Units deploy into Reserves; marching into Main Theater transitions readiness from Readied to Exhausted (`FR-INH-002`), naturally pacing attacks without an artificial summoning sickness flag. Furthermore, `FR-KWD-002f` is `Aggressive N`: *"Triggered ability firing on `EVENT_ATTACK_DECLARE`, granting +N Power until end of turn."* It is a combat power buff on attack, NOT a haste mechanic bypassing summoning sickness.
  - **Exploit / Failure Scenario:**
    1. A unit in Main Theater is readied on the turn it is deployed via an effect (or acquired via `GAIN_CONTROL`).
    2. Under `FR-ENG-031`, the engine erroneously blocks it from attacking citing "summoning sickness".
    3. Conversely, a unit with `Aggressive 1` is treated as having "haste", allowing it to attack while exhausted or from the Reserves backline.
  - **Remedy:** Remove "summoning sickness" entirely from `FR-ENG-031` and the engine state model. Attack legality checks must verify strictly: (1) asset is in Main Theater (or is HQ), (2) asset is `READIED`, (3) Guardian targeting constraints (`FR-KWD-002j`). Re-align `Aggressive N` as an `EVENT_ATTACK_DECLARE` trigger granting +N power during Step 2.

- **[DA-ENG-006] Contradiction on Headquarters Combat & Asymmetric Retaliation**
  - **Requirement Citations:** `FR-ENG-030`, `FR-ENG-033` (contradicts [`FR-CAP-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L262), [`FR-CBT-002`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L504), and Design Rationale [`DA-RULES-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L272))
  - **Vulnerability / Flaw:**
    1. `FR-ENG-030` dictates: *"Only units positioned in the Main Theater may declare attacks."* This directly contradicts `FR-CAP-001` and Design Rationale `DA-RULES-001`, which explicitly permit Headquarters to attack (`canAttack(): Permitted for UNIT in Main Theater and HEADQUARTERS`).
    2. `FR-ENG-033` states: *"If target is a unit (not structure/HQ), target calculates retaliation damage = Target.EffectivePower."* This directly contradicts `FR-CBT-002`, which mandates: *"Retaliation is 100% symmetric across all combat pairings (Unit vs Unit, Unit vs Structure, Unit vs HQ, HQ vs HQ)."*
  - **Exploit / Failure Scenario:**
    1. An attacking 2/2 unit attacks an opponent's Headquarters.
    2. The opposing HQ has 2 Base Combat Power. Under `FR-ENG-033`, the HQ does not retaliate because it is not a unit. The attacker takes 0 damage and survives.
    3. Under canonical `FR-CBT-002`, the HQ retaliates with its effective combat power, dealing 2 damage to the attacking unit and destroying it.
    4. Exempting HQs and structures from retaliating makes attacking HQs completely risk-free for low-health attackers, skewing the entire combat and tempo meta.
  - **Remedy:** Harmonize `FR-ENG-030` and `FR-ENG-033` with `FR-CAP-001` and `FR-CBT-002`: permit Headquarters in HQ Zone to declare attacks when Readied, and calculate simultaneous retaliation damage across ALL defending combatant types with effective Power > 0.

- **[DA-ENG-007] Inverted APNAP Trigger Ordering & Controller Priority**
  - **Requirement Citations:** `FR-ENG-014` (contradicts [`FR-TRG-003`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L396))
  - **Vulnerability / Flaw:** `FR-ENG-014` specifies:
    1. Active Player triggers pushed first in timestamp order.
    2. Non-Active Player triggers pushed on top.
    Invariant: *"Because the stack is LIFO, the Non-Active Player's triggers resolve first"*.
    However, `01_Game_Rules_and_Card_Modeling_FRD.md` `FR-TRG-003` explicitly codifies:
    *"Triggers occurring simultaneously push via Active Player, Non-Active Player (APNAP) order: non-active player triggers push first; active player triggers push on top and resolve first. Allied triggers are ordered by their controller."*
    Furthermore, `FR-ENG-014` forces Allied triggers to push in "timestamp order" rather than allowing their controller to order them.
  - **Exploit / Failure Scenario:**
    1. On Player 1's turn, a board wipe kills Player 1's unit (death trigger: "Draw 2 cards") and Player 2's unit (death trigger: "Opponent discards their hand").
    2. Under canonical `FR-TRG-003`, Non-Active Player (P2) pushes first, Active Player (P1) pushes on top. P1 draws 2 cards first, then P2 forces discard.
    3. Under `FR-ENG-014`, P2 pushes on top and resolves first: P1 discards their empty hand, then P1 draws 2 cards and keeps them.
    4. Game resolution is inverted 180 degrees based on stack push order.
  - **Remedy:** Align `FR-ENG-014` with `FR-TRG-003`: Non-Active Player pushes triggers first, Active Player pushes triggers on top so Active Player triggers resolve first. Where multiple allied triggers fire simultaneously, allow controller ordering (or deterministic tie-break fallback if automated).

- **[DA-ENG-008] Fabricated Hand Size Discard Step Violating Universal Uncapped Zone Rule**
  - **Requirement Citations:** `FR-ENG-009` (contradicts [`FR-ZON-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L162), [`FR-TRN-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L242))
  - **Vulnerability / Flaw:** `FR-ENG-009` Phase 4 End Phase includes:
    *"Sub-Phase 4b (Hand Size Cleanup): Assert hand size <= 10; prompt or discard excess cards."*
    Card Town rules explicitly establish in `FR-ZON-001`:
    - *"Player Hand: Private hand area per player with uncapped capacity (no maximum hand size)."*
    - *"Zone Capacity Universal Invariant: All zones (Hand, Reserves, Main Theater, Discard) are uncapped with no upper maximum entity capacity."*
    Card Town has NO hand size limit and NO end-of-turn discard step. Furthermore, `FR-ENG-009` omits End Phase Sub-Phase 4c resource depletion (`FR-RES-002`: unspent Requisition and Tactics pools reset to 0).
  - **Exploit / Failure Scenario:**
    1. A player ends their turn with 11 cards in hand from card draw maneuvers.
    2. The engine halts execution, transitions to `STATE_AWAITING_CHOICE`, and demands the player discard down to 10 cards.
    3. The Godot client has no UI screen or protocol message for "discard down to maximum hand size" because no such mechanic exists in Card Town, causing the turn to hang until the action clock expires and randomly discards cards.
  - **Remedy:** Eliminate Sub-Phase 4b hand discard entirely from `FR-ENG-009`. Re-structure End Phase to match `FR-TRN-001`: Sub-Phase 4a (End triggers enqueued), Sub-Phase 4b (Stack auto-resolution), Sub-Phase 4c (Unspent resource pool wipe & temporary modifier decay), Sub-Phase 4d (Handover).

- **[DA-ENG-009] Missing Wire Error Taxonomy Alignment & Guardian Priority Violation**
  - **Requirement Citations:** `FR-ENG-041`, Section 8.1 (contradicts [`04_Client_Engine_Protocol_Networking_FRD.md`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/04_Client_Engine_Protocol_Networking_FRD.md#L745) Section 7.1, `DA-NET-003`, `DA-NET-005`)
  - **Vulnerability / Flaw:** Section 8.1 invents an isolated `ERR_ENG_*` error code list that completely diverges from the established wire error codes in `04_Client_Engine_Protocol_Networking_FRD.md`. Specifically:
    - Missing `ERR_GUARDIAN_PRIORITY_VIOLATION` (specifically mandated by `DA-NET-005` so the Godot UI can highlight the blocking Guardian unit).
    - Missing `ERR_INVALID_COMMAND_ARGUMENTS` (mandated by `DA-NET-003` for input boundary clamping).
    - Uses `ERR_ENG_DIRTY_STACK_ACTION` instead of `ERR_STACK_NOT_EMPTY`.
    - Uses `ERR_ENG_INSUFFICIENT_RESOURCE` instead of `ERR_INSUFFICIENT_RESOURCES`.
    - Uses `ERR_ENG_TARGET_INVALID` instead of `ERR_INVALID_TARGET`.
  - **Exploit / Failure Scenario:**
    1. A player tries to attack an asset past a Guardian. The engine returns `ERR_ENG_TARGET_INVALID`.
    2. The client receives a generic failure rather than `ERR_GUARDIAN_PRIORITY_VIOLATION`, preventing the client UI from highlighting the mandatory Guardian target asset, degrading onboarding clarity.
  - **Remedy:** Unify Section 8.1 error codes with `04_Client_Engine_Protocol_Networking_FRD.md` Section 7.1.

---

### 🟡 Medium

- **[DA-ENG-010] PRNG Desync Vulnerability on Transaction Rollback**
  - **Requirement Citations:** `FR-ENG-008`, `FR-ENG-002`
  - **Vulnerability / Flaw:** `FR-ENG-008` states that `tx.Rollback()` restores touched entities, zones, and resource balances, and purges the event buffer. However, it fails to record and restore the PRNG stream counters (`StreamShuffle`, `StreamTargeting`, `StreamMatchDecisions`).
  - **Exploit Scenario:**
    1. An action or stack item consumes PRNG ticks (e.g. random targeting or draw) before failing validation.
    2. The transaction aborts and state rolls back, but PRNG stream counters are NOT rolled back.
    3. A client intentionally submits invalid commands to burn bad random seeds and manipulate upcoming random rolls, breaking seeded determinism and causing `.rec` replay desyncs.
  - **Remedy:** Mandate that `Transaction` delta-snapshots include the monotonic state/counter of all three PRNG streams, ensuring `Rollback()` restores PRNG state cleanly.

- **[DA-ENG-011] Missing Continuous Aura Snapshot Lock in Combat Step 3**
  - **Requirement Citations:** `FR-ENG-023`, `FR-ENG-033` (contradicts [`FR-LAY-003`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L441), [`FR-CBT-002`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L520))
  - **Vulnerability / Flaw:** `FR-ENG-023` specifies continuous auras re-evaluate continuously upon any zone move or death. However, `FR-LAY-003` and `FR-CBT-002` explicitly mandate the **Continuous Aura Snapshot Lock**: continuous aura condition gates (such as *Cornered* count checks) are snapshotted and locked at the start of Step 3 combat damage, remaining locked throughout simultaneous damage and retaliation calculations, re-evaluating only after state-based death checks complete.
  - **Exploit / Failure Scenario:**
    1. Player 1 has a unit buffed with +1/+1 while *Cornered*.
    2. During Step 3 simultaneous combat, Player 2's unit takes lethal damage.
    3. If continuous auras re-evaluate mid-tick, Player 1 ceases to be Cornered, loses the health buff, and dies to retaliation damage that should have been survived.
  - **Remedy:** Formally specify the Continuous Aura Snapshot Lock at the entry of Step 3 combat in `FR-ENG-033`.

- **[DA-ENG-012] Incompatible 6-Layer Attribute Pipeline Architecture**
  - **Requirement Citations:** `FR-ENG-025`, `FR-ENG-027`, `FR-ENG-028` (contradicts [`FR-LAY-001`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L431) - [`FR-LAY-003`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L439))
  - **Vulnerability / Flaw:** `FR-ENG-025` redefines the 6 layers in a way that conflicts with `01_Game_Rules_and_Card_Modeling_FRD.md`:
    - Canonical Layer 2 is "Type, Subtype & Identity Grants", but `FR-ENG-025` calls it "Blueprint Modifiers: Static set-wide rules adjustments".
    - Canonical Layer 3 is "Dynamic Keyword Grants & Suppressions", but `FR-ENG-025` calls it "Instance Modifiers: Card-specific permanent modifiers".
    - Canonical Layer 6 clamps only Combat Power at 0 (`FR-LAY-001`), whereas `FR-ENG-025` invents "+1/+1 counters and discrete tokens" and clamps defense to 0.
  - **Remedy:** Realign the layer definitions in `FR-ENG-025` to match `FR-LAY-001` through `FR-LAY-003`.

---

### 🔵 Low / Clarification

- **[DA-ENG-013] Missing Specification for Inherent March and Resource Conversion Commands**
  - **Requirement Citations:** `FR-ENG-006`, `FR-ENG-010` (contradicts [`FR-INH-002`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L264), [`FR-RES-003`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/01_Game_Rules_and_Card_Modeling_FRD.md#L228))
  - **Issue:** `05_Headless_Game_Engine_FRD.md` lists manual actions as "playing a maneuver card, deploying a unit/structure, manually exhausting an asset for resources, or declaring combat", but never formally defines the state execution rules for `March` (`FR-INH-002`) or `Requisition to Tactics Conversion` (`FR-RES-003`).
  - **Clarification Needed:** Add explicit execution requirements for the inherent March action and instantaneous resource conversion in Domain 2.

- **[DA-ENG-014] Ambiguous Public State Checksum Hash Algorithm**
  - **Requirement Citations:** `FR-ENG-005` (contradicts [`FR-PAC-004`](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/04_Client_Engine_Protocol_Networking_FRD.md#L690))
  - **Issue:** `FR-ENG-005` specifies calculating a rolling 64-bit "MurmurHash3 or FNV-1a state checksum", while `04_Client_Engine_Protocol_Networking_FRD.md` `FR-PAC-004` specifies "64-bit CRC/FNV-1a". Leaving multiple algorithm choices across different FRDs invites cross-language hash divergence between Go server and Godot client.
  - **Clarification Needed:** Standardize strictly on 64-bit FNV-1a across all documents.

---

## 3. Redundancy & De-Scoping Ledger

| Requirement ID | Fully Subsumed By | Rationale / Recommendation |
| :--- | :--- | :--- |
| `FR-ENG-003` (Deterministic Seed Derivation & Salts) | `FR-ENG-002` | Seed salting and stream derivation can be merged directly into the multi-stream PRNG architectural requirement rather than maintaining a separate requirement. |
| `FR-ENG-027` (Layer Recalculation Ordering & Mutator Precedence) | `FR-ENG-025` | Timestamp evaluation within layers is an intrinsic property of the 6-layer pipeline and does not require an independent requirement. |
| `FR-ENG-028` (Effective Stat Clamping & Non-Negative Invariants) | `FR-ENG-025` & `FR-ENG-041` | Layer 6 clamping is already the terminal stage of `FR-ENG-025` and is duplicated as Invariant 1 in `FR-ENG-041`. |

---

## 4. Socratic Interrogation Register

1. *Why does the engine specification introduce Magic: The Gathering mechanics—such as summoning sickness, MTG-style LIFO trigger resolution where the non-active player's triggers resolve first, and a 10-card hand discard phase—when Card Town was explicitly designed with zero summoning sickness, uncapped hand sizes, and active-player-first trigger resolution?*
2. *How is a game match expected to conclude when the engine state machine lacks any concept of Headquarters destruction, empty-deck fatigue, or victory/defeat conditions?*
3. *If the card compiler produces ASTs utilizing the canonical 25 effect primitives in `FR-PRM-001`, how can the engine execute the card catalog when it only recognizes 14 arbitrarily renamed primitives and omits bounce, deck search, and card tucking?*
4. *Why does the engine invert the canonical damage mitigation pipeline by evaluating Armor before Stalwart, when doing so completely destroys the intentional design synergy of defensive cards?*
5. *Why is Scratch Company specified as an ephemeral token spawned from thin air when Card Town's core design defines Scratch as a double-blind draw from the player's actual deck?*
