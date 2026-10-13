# Product Manager Devil's Advocate Audit: Game Rules & Card Modeling FRD (Master FRD v3.0)

- **Target Document:** [Game_Rules_and_Card_Modeling_FRD.md](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/Game_Rules_and_Card_Modeling_FRD.md)
- **Auditor:** PM Devil's Advocate (`pm-devils-advocate`)
- **Date:** 2026-10-13
- **Verdict:** REVISE WITH BLOCKERS

---

## 1. Executive Summary & Red-Team Assessment

The **Master Game Rules & Card Modeling FRD (v3.0)** establishes an ambitious and highly structured foundation for Card Town, unifying match state, entity topology, execution pipelines, and card blueprints. The 5-stage damage mitigation pipeline, clean-stack execution model, and 6-layer attribute architecture represent robust systems engineering.

However, from a product, game balance, and systems red-teaming standpoint, several severe design vulnerabilities threaten core match health:
1. **Turn 1 Alpha Strike & Base Racing Dominance**: The complete absence of summoning sickness combined with permitting direct HQ attacks from Turn 1 (`FR-SET-004`, `FR-CBT-001`) creates an insurmountable first-player tempo advantage where aggressive decks can bypass interactive board play entirely.
2. **Infinite Armor Hoarding & Anti-Fun Stalemate Vectors**: Dynamic Armor is permanent and non-decaying (`FR-LAY-002`), while cards like *Dominator Warsuit* provide repeatable, uncapped resource sinks. Stall decks can generate 20+ armor, rendering units permanently immune to standard combat interaction.
3. **Sequential Retaliation Paradox in Simultaneous Combat**: `FR-CBT-002` mandates simultaneous combat resolution, yet introduces "sequential defender retaliation" for Berserk units (`FR-KWD-002o`), creating contradictory timing windows and unresolvable death trigger states.
4. **Asymmetric Interaction Overhead vs Agency**: The strict clean-stack invariant (`FR-QUE-001`) bans off-turn response actions (`FR-TRN-002`), yet imposes frequent 15-second synchronous choice prompts during trigger resolution (`FR-TRG-002`), capturing the latency and pacing penalties of MTG priority without providing responsive counterplay.

**Recommendation:** Halt production sign-off until the blockers in Turn 1 combat timing, armor decay mechanics, and simultaneous retaliation sequencing are resolved.

---

## 2. Severity-Ordered Punch-List

### 🔴 Blocker / Critical

- **[DA-RULES-001] Turn 1 First-Player Alpha Strike & Unmitigated Base-Racing**
  - **Requirement Citations:** `FR-SET-004`, `FR-CAP-001`, `FR-CBT-001`, `FR-INH-001`
  - **Vulnerability / Flaw:** Units enter the battlefield readied by default (`FR-ENT-003`, `FR-INH-001`) with no summoning sickness or deployment sickness. Turn 1 rules permit all actions, deployments, and combat declarations (`FR-SET-004`). Combat rules explicitly state: *"Direct cross-board attacks against opposing HQs are permitted from Turn 1"* (`FR-CBT-001`), and HQs themselves may attack (`FR-CAP-001`). Non-active players cannot take manual actions off-turn (`FR-TRN-002`).
  - **Exploit / Failure Scenario:**
    1. Player 1 wins the coin toss (`FR-SET-003`) and elects to Play First.
    2. Player 1 exhausts their HQ to generate 1 Requisition (`FR-INH-004`), deploys an aggressive 1-cost unit into the Main Theater (e.g. *Furnace Golem* or a cheap striker), and immediately declares an attack with that unit directly against Player 2's HQ.
    3. Player 1 additionally attacks with their own Headquarters directly into Player 2's Headquarters.
    4. Player 2 has had zero turns, holds an empty board, has no units with Guardian in play, and has no legal window to respond. Player 2's HQ takes significant unpreventable damage before Player 2 ever receives priority.
    5. In hyper-aggro matchups, Player 1 establishes an irreversible race advantage purely from the coin toss, forcing Player 2 to spend the entire match reacting to lethal pressure.
  - **Remedy:** Prohibit attack declarations on Turn 1 across all entities, or enforce that units deployed to the Main Theater cannot attack on the turn they are deployed (standard summoning sickness) unless granted an explicit *Rush/Charge* keyword.

- **[DA-RULES-002] Permanent Dynamic Armor Accumulation & Uncapped Stall Deadlocks**
  - **Requirement Citations:** `FR-LAY-002`, `FR-REP-004`, `FR-ACT-002`
  - **Vulnerability / Flaw:** `FR-LAY-002` codifies that *"Dynamic Armor is a persistent, consumable defensive pool... Armor does NOT expire at end of turn, does not decay, and stacks additively from discrete addition events."* Furthermore, `FR-ACT-002` highlights *Dominator Warsuit* (`R1: gain 1 armor`) as an intentional repeatable Requisition sink.
  - **Exploit / Failure Scenario:**
    1. A control/stall player stabilizes with *Dominator Warsuit* and defensive structures (e.g. *Animoceps Spore Cloud* or *Ancestral Monument*).
    2. During successive quiet turns, the player funnels all unspent Requisition (e.g. 6-10 Requisition per turn) into `gain 1 armor`.
    3. Over 3 turns, the asset accumulates 25+ dynamic Armor.
    4. Because Armor absorbs damage 1-for-1 before Health in Stage 4 (`FR-REP-004`) and never decays, the unit becomes completely impervious to conventional board trading.
    5. Paired with damage redirection (*Animoceps Spore Cloud*), the player creates an impenetrable damage sponge, dragging the game out to empty-deck exhaustion (`FR-WIN-001`) with zero interactive counterplay for non-hard-removal factions.
  - **Remedy:** Cap dynamic Armor to an asset's base Maximum Health (`Armor <= Max Health`), or enforce that non-inherent dynamic Armor decays to zero during End Phase Sub-Phase 4c (`FR-TRN-001`).

- **[DA-RULES-003] Berserk Multi-Target Sequential Retaliation Paradox in Simultaneous Combat**
  - **Requirement Citations:** `FR-CBT-002`, `FR-KWD-002o`, `FR-REP-004`
  - **Vulnerability / Flaw:** `FR-CBT-002` establishes that Step 3 combat damage is dealt **simultaneously** equal to effective Combat Power. However, for Berserk units (`FR-KWD-002o`), the spec mandates: *"each defender's retaliation is processed as an independent, discrete damage instance sequentially routed through `FR-REP-004`; the attacking player designates the sequential resolution order of the defender retaliations upon attack declaration."*
  - **Exploit / Failure Scenario:**
    1. Player 1 attacks with a Berserk unit against Defender A (4 Power) and Defender B (4 Power). Player 1's unit has 5 Health.
    2. Player 1 designates Defender A to resolve first, then Defender B.
    3. Defender A deals 4 damage to the Berserk unit (Health drops to 1).
    4. What if Defender A possessed lethal deathtouch or dealt 5 damage, reducing the Berserk unit to 0 Health during the first sequential retaliation?
    5. Under strict sequential execution, the Berserk unit is dead before Defender B's retaliation resolves. Does Defender B still retaliate against a dead entity? Does Defender B deal damage to nothing? Do Defender B's on-damage triggers fire?
    6. If combat is simultaneous, introducing sequential ordering creates a contradictory state machine where an entity can be simultaneously dead and resolving further combat interactions.
  - **Remedy:** Eliminate arbitrary sequential retaliation ordering. Route simultaneous combat damage from all defenders into an aggregated combat damage calculation, or process defender retaliation instances in parallel through Stage 1-5 with state-based death checks deferred strictly until all Step 3 strikes conclude.

---

### 🟠 High

- **[DA-XP-001] Direct HQ Bypassing Renders Main Theater Engagement Optional**
  - **Requirement Citations:** `FR-CBT-001`, `FR-ZON-001`
  - **Vulnerability / Flaw:** Unless an opponent has deployed an asset with `Guardian` (`FR-KWD-002j`) in the Main Theater, attack declarations may target the opposing Headquarters directly at all times.
  - **Player Experience Breakdown:** Card Town's central tactical conceit is the dual-zone battlefield (Reserves vs Main Theater). However, because any unit in the Main Theater can ignore all non-Guardian enemies and strike the HQ, board control is frequently devalued. Players who build intricate unit synergies are simply ignored while opponents execute uninteractive face attacks. Matches degrade into a coin-flip damage race rather than tactical territory control.
  - **Remedy:** Mandate that if an opponent controls *any* ready unit in the Main Theater, direct attacks on the HQ require an open lane or cost an additional movement/tactics toll, elevating Guardian to an active taunt rather than the sole protector of the base.

- **[DA-XP-002] 15-Second Synchronous Stack Prompts Destroy Match Pacing Without Providing Priority**
  - **Requirement Citations:** `FR-TRN-002`, `FR-QUE-001`, `FR-TRG-002`, `FR-MOD-001`
  - **Vulnerability / Flaw:** `FR-TRN-002` strictly bars non-active players from casting spells or activating abilities off-turn. Yet, whenever triggered abilities resolve (e.g. on attack declarations or damage), inactive players are hit with interactive choice prompts (`FR-TRG-002`) bounded by a 15-second action clock (`FR-MOD-001`).
  - **Player Experience Breakdown:** The worst of both worlds: players endure repeated pause-and-wait prompt cycles during their opponent's turn (stopping fluid digital play), but are denied the satisfying agency of playing counter-spells, combat tricks, or instant-speed cards. Pacing feels sluggish and stop-and-go, violating modern digital CCG pacing conventions (Hearthstone/Marvel Snap style flow).
  - **Remedy:** Restructure off-turn triggered choices to auto-resolve via smart deterministic defaults or client pre-targeting whenever possible. Reserve interactive prompts strictly for high-impact decision trees.

- **[DA-RULES-004] Replacement Interception Loop & Death Interception Re-Triggering**
  - **Requirement Citations:** `FR-REP-001`, `FR-REP-003`, `FR-ENT-103`
  - **Vulnerability / Flaw:** `FR-REP-001` allows replacement watchers to intercept `WOULD_BE_DESTROYED`. `FR-REP-003` specifies a single application per root event, but notes: *"for replacement watchers intercepting entity destruction (*Shifting Impossibility*), each asset's prospective destruction emits an individual interceptable death hook, permitting multiple simultaneous entity destructions from a single root damage event to each be replaced."*
  - **Exploit / Failure Scenario:**
    1. *Shifting Impossibility* states: *"If an allied Scratch Company unit would be killed, return it to its owner's hand and deal 1 damage to this unit instead."*
    2. Player controls two copies of *Shifting Impossibility* (Unit A and Unit B) and one Scratch Company.
    3. The Scratch Company is dealt lethal damage. Both Unit A and Unit B attempt to replace the death.
    4. Who replaces it? If Unit A replaces it, the Scratch is bounced and Unit A takes 1 damage. What if Unit A takes lethal damage from this self-inflicted damage? Does that emit another prospective death hook during root event resolution?
    5. The interaction between cascading replacement damage and simultaneous death events lacks a deterministic priority tie-breaker between duplicate identical replacement watchers.
  - **Remedy:** Codify an explicit controller-choice or timestamp rule for simultaneous replacement conflicts, and state that damage dealt *as a consequence* of a replacement effect is a new discrete event that cannot trigger circular replacement loops.

---

### 🟡 Medium

- **[DA-SCOPE-001] Over-Engineered 6-Layer Attribute Pipeline for a 167-Card Base Set**
  - **Requirement Citations:** Section 6.2 (`FR-LAY-001` - `FR-LAY-003`)
  - **Vulnerability / Flaw:** The architecture implements a full 6-layer continuous attribute recalculation pipeline directly modeled after Magic: The Gathering's comprehensive rules (Layer 1 Base -> Layer 2 Types -> Layer 3 Keywords -> Layer 4 Continuous Auras -> Layer 5 Turn Modifiers -> Layer 6 Mathematical Clamping).
  - **Assessment:** Card Town has exactly 167 cards, none of which feature complex continuous type-changing effects like MTG's *Blood Moon* or *Humility*. Implementing a full continuous recalculation pipeline across 6 layers introduces massive engine and serialization complexity (`FR-PAC-004` requires hash parity over Layer 6) when discrete event-driven stat modification handles 95% of use cases cleanly.
  - **Remedy:** Simplify to a 3-tier modifier model (Base -> Continuous Static Auras -> Transient Timed Buffs), eliminating formal layer dependency graphs until the card pool actually demands it.

- **[DA-XP-003] Inherent Action Clean-Stack Constraint Causes False Action Lockouts**
  - **Requirement Citations:** `FR-CAP-001`, `FR-INH-001`, `FR-QUE-001`
  - **Vulnerability / Flaw:** `FR-CAP-001` and `FR-INH-001` require `Stack.isEmpty() == true` to execute any inherent action (March, Attack, Resource Generation, Scratch).
  - **Player Experience Breakdown:** If a lingering delayed trigger or end-of-phase cascade sits on the stack, a player attempting to tap an asset for resources or move a unit receives a hard error (`ERR_STACK_NOT_EMPTY`). For new players, this feels like an unresponsive UI bug because the "stack" is an abstract engine concept that is visually invisible until cards animate.
  - **Remedy:** Allow the client to queue inherent actions optimistically, resolving them automatically the moment pending stack items clear, rather than rejecting commands with confusing stack-state errors.

---

### 🔵 Low / Clarification

- **[DA-CLARIFY-001] Ambiguity on Negative Resistant Value Interactions with Zero Raw Damage**
  - **Requirement Citations:** `FR-LAY-001`, `FR-REP-004`
  - **Issue:** `FR-REP-004` specifies: *"an incoming damage instance of 0 raw damage against a target with Resistant -1 is amplified to 1 damage ($0 - (-1) = 1$)"*. Does an attack declared by a 0-power unit (`FR-CBT-001` allows 0-power attacks) deal actual combat damage if the defender has `Resistant -1`?
  - **Clarification Needed:** Does a 0-power combat strike count as an incoming damage instance that gets amplified, or does a 0-power strike deal "no damage" and skip the pipeline?

- **[DA-CLARIFY-002] Clarification on Scratch Company Face-Down Status in Discard vs Attached**
  - **Requirement Citations:** `FR-ZON-005`, `FR-ENT-102`
  - **Issue:** When an asset tucked under *NO-AH-5* is purged because *NO-AH-5* departs the battlefield, it routes to `DISCARD` and reveals its identity. Does the act of routing from `ATTACHED` to `DISCARD` trigger `EVENT_DESTROY`? The spec states in `FR-ENT-102`: *"without emitting EVENT_DESTROY"*, but `FR-ZON-005` says departing to a public zone reveals it. Designer should clarify if graveyard-entry triggers (*Solarheart Monk*) fire when tucked cards route to Discard.

---

## 3. Redundancy & De-Scoping Ledger

| Requirement ID | Fully Subsumed By | Rationale / Recommendation |
| :--- | :--- | :--- |
| `FR-ZON-101` (Clean Asset Deallocation) | `FR-ZON-004` (Non-Battlefield Zone Transition) | Fully redundant. `FR-ZON-004` already mandates deallocating `AssetInstance` and wiping state attributes when leaving battlefield. Collapse into `FR-ZON-004`. |
| `FR-TGT-102` (Declarative Targeting Filters) | `FR-TGT-001` (Universal Modular Target Filtering) | Identical conjunction schema definitions (`[Allegiance] AND [Health] AND [Subtype]...`). Merge into a single canonical targeting requirement. |
| `FR-TGT-103` (Two-Stage Target Verification) | `FR-TGT-003` (Target Validation & Fizzle Rules) | Both define declaration verification, resolution re-verification, and fizzling rules. Collapse `FR-TGT-103` into `FR-TGT-003`. |
| `FR-ENT-101` (Scratch Double-Blind Wrapper) | `FR-ZON-005` (Scratch Fog of War Lifecycle) | `FR-ZON-005` already details double-blind deployment, masking, and public/private reveals. Merge `FR-ENT-101` into `FR-ZON-005`. |

---

## 4. Socratic Interrogation Register

1. *Why are direct HQ attacks permitted on Turn 1 if player 2 has not had an opportunity to play a single card or deploy a defender?*
2. *If the game engine strictly prohibits off-turn player actions (no instants, no off-turn activated abilities), why does the engine incur the massive networking, UI, and pacing cost of an asynchronous Resolution Stack with 15-second client timeouts?*
3. *What prevents a match between two defensive players utilizing permanent Dynamic Armor from degenerating into a 45-minute stall fest decided purely by deck-out loss?*
4. *If combat damage in Step 3 is simultaneous, why does the attacking player need to specify a sequential resolution order for Berserk defenders? What is the physical game state of an asset between sequential strikes?*
5. *Why does Headquarters have Combat Power and attack capabilities? If HQs can attack, does this not accelerate degenerate base-racing at the expense of unit-to-unit tactical combat?*
