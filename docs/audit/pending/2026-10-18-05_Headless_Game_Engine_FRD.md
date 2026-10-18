# Product Manager Devil's Advocate Audit: Headless Game Engine FRD

- **Target Document:** [docs/frd/05_Headless_Game_Engine_FRD.md](file:///c:/Users/blade/animus-mundi/animus-mundi/docs/frd/05_Headless_Game_Engine_FRD.md)
- **Auditor:** PM Devil's Advocate (`pm-devils-advocate`)
- **Evaluation Date:** 2026-10-18
- **Status:** PENDING (Active review queue)
- **Verdict:** REVISE WITH HIGH FINDINGS

## 1. Executive Summary & Red-Team Assessment

This audit evaluates `05_Headless_Game_Engine_FRD.md` (Master FRD v1.3) specifically focused on requirement duplication, copycat definitions, and domain authority boundaries across the master specification suite.

`05_Headless_Game_Engine_FRD.md` canonically owns the Go kernel runtime architecture (`cardtown/engine`), discrete integer math, isolated PRNG stream counters (`FR-ENG-002`), transactional execution with delta-rollback mementos (`FR-ENG-008`), 3-tier runtime memory structs (`CardBlueprint`, `CardInstance`, `AssetInstance`, `AssetSnapshot`), frame tick invariants and rolling state hashes (`FR-ENG-005`), stack frame taxonomy (`FR-ENG-012`), Go AST type-switch visitors (`FR-ENG-018`), CLI simulation runners, and automated benchmarks.

However, a rigorous red-team audit reveals extensive requirement duplication where FRD 05 creates independent requirement IDs that re-define fundamental game rules canonically owned by `01_Game_Rules_FRD.md`, without introducing any engine-specific execution mechanisms, data structures, or transaction logic:
1. `FR-ENG-030` (Combat Zone Eligibility Enforcement) is a single-sentence copycat of `01_Game_Rules_FRD.md` (`FR-CBT-001`, `FR-CAP-001`), while `FR-ENG-031` already performs command intake and validation.
2. `FR-ENG-034` (Canonical 5-Stage Damage Mitigation Pipeline) re-defines the 5-stage damage mitigation pipeline, mathematical formulas, and trigger conditions already fully codified in `01_Game_Rules_FRD.md` (`FR-REP-004`).
3. `FR-ENG-014` (APNAP Trigger Dispatching) and `FR-ENG-015` (Two-Stage Target Verification) re-define APNAP ordering and two-stage target fizzle rules from `01_Game_Rules_FRD.md` (`FR-TRG-003`, `FR-TGT-003`).
4. Section 4.1 contains thin passthrough requirements (`FR-ENG-019`, `FR-ENG-020`, `FR-ENG-024`) that merely state the engine must evaluate schemas from FRD 02 per the rules of FRD 01.

## 2. Severity-Ordered Punch-List

### 🔴 Blocker / Critical
_None._

### 🟠 High
- **[DA-ENG-005] Redundant Combat Zone Eligibility Requirement in FR-ENG-030**
  - **Requirement Citations:** `FR-ENG-030`
  - **Vulnerability / Flaw:** `FR-ENG-030` (Combat Zone Eligibility Enforcement) is a single sentence that re-states attacker and target eligibility rules codified in `01_Game_Rules_FRD.md` (`FR-CBT-001`, `FR-CAP-001`, `FR-ZON-002`). It specifies no Go structs, no error codes, no stack frames, and no engine mechanisms. In addition, `FR-ENG-031` ("Step 1: Attack Declaration, Upfront Exhaustion & FRAME_ATTACK Push") immediately performs the actual command intake and validation (`AttackCommand`, clean stack, readiness, `Guardian` priority, `Berserk` target binding, and pushing `FRAME_ATTACK`).
  - **Exploit / Failure Scenario:** Maintenance overhead and specification bloat without adding any architectural value to the headless engine.
  - **Remedy:** Formally mark `FR-ENG-030` as subsumed by `FR-CBT-001` (Game Rules) and `FR-ENG-031` (Step 1 Attack Declaration).

- **[DA-ENG-006] Redundant 5-Stage Damage Mitigation Pipeline in FR-ENG-034**
  - **Requirement Citations:** `FR-ENG-034`
  - **Vulnerability / Flaw:** `FR-ENG-034` (Canonical 5-Stage Damage Mitigation Pipeline) re-defines the 5-stage damage mitigation pipeline, mathematical formulas, and trigger emission criteria (`EVENT_TAKE_DAMAGE`, `EVENT_DEAL_DAMAGE`) already fully codified in `01_Game_Rules_FRD.md` (`FR-REP-004`). `FR-ENG-034` adds no Go structs, method signatures, or engine rollback invariants beyond repeating `FR-REP-004`. State-based death sweeps following damage are already handled in `FR-ENG-035`.
  - **Exploit / Failure Scenario:** Dual-spec drift when damage mitigation formulas or stage interactions are updated.
  - **Remedy:** Formally mark `FR-ENG-034` as subsumed by `FR-REP-004` (Game Rules) and `FR-ENG-035` (State-Based Death Sweep & Discard Transitions).

### 🟡 Medium
- **[DA-ENG-007] Redundant Trigger Ordering (FR-ENG-014) & Target Verification (FR-ENG-015) Requirements**
  - **Requirement Citations:** `FR-ENG-014`, `FR-ENG-015`
  - **Vulnerability / Flaw:**
    1. `FR-ENG-014` (APNAP Trigger Dispatching & Simultaneous Event Ordering) repeats the exact 3-step APNAP sequence and LIFO resolution codified in `01_Game_Rules_FRD.md` (`FR-TRG-003`). The LIFO stack data structure is already specified in `FR-ENG-012`.
    2. `FR-ENG-015` (Two-Stage Target Verification & Fizzle Execution Contract) repeats the two verification checkpoints, partial invalidation, distributed pool evaporation, and composite clause survival codified in `01_Game_Rules_FRD.md` (`FR-TGT-003`).
  - **Exploit / Failure Scenario:** Redundant requirements inflate the engine specification count and duplicate game rules without introducing engine-specific contracts.
  - **Remedy:** Formally mark `FR-ENG-014` and `FR-ENG-015` as subsumed by `FR-TRG-003`, `FR-TGT-003`, and `FR-ENG-012`.

### 🔵 Low / Clarification
- **[DA-ENG-008] Thin Passthrough Requirements in Section 4.1 (FR-ENG-019, FR-ENG-020, FR-ENG-024)**
  - **Requirement Citations:** `FR-ENG-019`, `FR-ENG-020`, `FR-ENG-024`
  - **Vulnerability / Flaw:** `FR-ENG-019` (Dynamic FilterDef Evaluator), `FR-ENG-020` (Condition Gate & ValueToken Resolver), and `FR-ENG-024` (Maneuver Card Enactment) are 1-to-2 sentence passthroughs stating that the engine must evaluate schemas from FRD 02 per the rules of FRD 01. The actual AST evaluation engine is specified in `FR-ENG-018` (Type-Switch Visitor) and stack execution is in `FR-ENG-012`.
  - **Exploit / Failure Scenario:** Unnecessary granularity in requirement taxonomy where single-sentence requirements act as mere pointers.
  - **Remedy:** Consolidate `FR-ENG-019` and `FR-ENG-020` into `FR-ENG-018`, and consolidate `FR-ENG-024` into `FR-ENG-012` and `FR-ENG-013`.

## 3. Redundancy & De-Scoping Ledger
| Requirement ID | Fully Subsumed By | Rationale / Recommendation |
| :--- | :--- | :--- |
| `FR-ENG-030` | `FR-CBT-001`, `FR-ENG-031` | Spatial combat eligibility belongs to FRD 01; attack declaration validation belongs to `FR-ENG-031`. |
| `FR-ENG-034` | `FR-REP-004`, `FR-ENG-035` | 5-stage damage mitigation belongs to FRD 01; death sweep execution belongs to `FR-ENG-035`. |
| `FR-ENG-014` | `FR-TRG-003`, `FR-ENG-012` | APNAP rules belong to FRD 01; stack architecture belongs to `FR-ENG-012`. |
| `FR-ENG-015` | `FR-TGT-003`, `FR-ENG-012` | Two-stage target verification and fizzle rules belong to FRD 01. |
| `FR-ENG-019`, `FR-ENG-020` | `FR-ENG-018`, FRD 02 | Thin passthrough evaluations subsumed into the AST type-switch interpreter. |
| `FR-ENG-024` | `FR-ENG-012`, `FR-ENG-013`, `FR-MNV-001` | Maneuver execution flow subsumed into stack frame lifecycle and cost validation. |

## 4. Socratic Interrogation Register
1. *Why does the Headless Engine FRD recreate distinct requirement numbers for fundamental game rules (combat eligibility, 5-stage mitigation, APNAP trigger ordering, two-stage targeting) instead of referencing FRD 01 and scoping FRD 05 requirements strictly to Go kernel implementation constructs (state machine transactions, memento rollback, memory layout, stack frame lifecycle, and deterministic test harnesses)?*
