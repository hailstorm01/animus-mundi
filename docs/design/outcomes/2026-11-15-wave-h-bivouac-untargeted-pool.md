Status: applied: FRD 01 v3.28, 2026-11-15
Directive: #157
Rulings: SVY-R51, SVY-R52

### 🎲 Design Outcome: an untargeted pool is split as it resolves (2026-11-15)

- **Question:** Is *Hammerhand Bivouac*'s "distribute 2 armor among allied units in the same zone" (no "chosen") split on placement, as targets, or as it resolves, as an untargeted selection? And what does a timeout do with a pool split (`FR-MOD-001` §6)? Raised by lead-architect in the wave H run (`autoloop/m2-wave-h`, session 6) before the *Bivouac* pool and card tasks were written, as a `Status: open` stub at this path. Design note Open Question 24.
- **Rulings:**

  | Ruling | Settles | Cards |
  | :--- | :--- | :--- |
  | `SVY-R51` | A pool distributed among *chosen* targets is split on declaration (`SVY-R3`), and its evaporation rule is unchanged. A pool distributed among entities named without *"chosen"* is an untargeted selection. The acting player splits it as the clause resolves, among the entities that match at that moment, so Ward and `FR-TGT-003` don't apply and nothing evaporates. Either way, each entity that receives part of the pool gets a positive whole amount, and the amounts add up to exactly the pool. With one match it takes the whole pool and nothing is asked (§1). With none, the clause does nothing. | Hammerhand Bivouac; Macrocannon Barrage (outside M2, targeted, unchanged) |
  | `SVY-R52` | A timed-out pool split puts the whole pool on the first legal entity in fixed order (§6), as a timed-out repeated target does (`SVY-R32`). For *Bivouac*, that's the allied unit that entered the Main Theater earliest. | Hammerhand Bivouac |

- **Decisions (user, 2026-11-15):** "A and a". That's alternative A for question 1 (split on resolution, no card text change) and (a) for question 2 (the whole pool on the first legal entity in fixed order).
- **Cards affected:**
  - *Hammerhand Bivouac* (Path structure, `data/cards/units-and-structures-path.csv:17`): in M2 scope (wave H, on `cards.NotYetExecutable`), in neither starter deck. Text unchanged.
  - *Macrocannon Barrage* (Soulless maneuver, `data/cards/maneuvers-soulless.csv:15`): outside M2. It is targeted, so it is split on declaration and evaporation applies, both unchanged. A timeout never reaches its split, because "up to 5 chosen" selects none on a timeout (§6).
- **Alternatives rejected:**
  - B, reword *Bivouac* to "among up to 2 chosen allied units": one timing for every pool, but it adds "chosen" to a card that never needed it, and it breaks the reading from `SVY-R3` and `SVY-R44` that untargeted picks happen on resolution.
  - C, remove the pool ("Up to 2 allied units … each gain 1 armor"): changes how the card plays (no more 2 armor on one unit).
  - (b), give one point at a time on a timeout: needs a second fallback rule, and would spread damage thinly in a future untargeted damage pool.
  - Note: with the clean-stack rule and no other start-of-turn trigger in the card data, A and B play the same for every current card. The ruling settles which rule is right and which engine model to build.

#### → `pm-frd-drafter` (FRD 01 and the ruling register)
- **Superseded wording:**
  - `docs/frd/01_Game_Rules_FRD.md:406` (`FR-TGT-002`, Distributed Pool Allocations): names *Bivouac* as split "to each chosen target upon declaration".
  - `docs/frd/01_Game_Rules_FRD.md:413` (`FR-TGT-003`, Distributed Pool Evaporation): correct for targeted pools, but doesn't say untargeted pools don't evaporate.
- **FRD 01 changes (v3.27 → v3.28):**
  1. `FR-TGT-002` (line 406): replace the bullet with: "**Distributed Pool Allocations (`SVY-R51`)**: An effect that distributes a fixed pool of damage or armor gives each entity that receives part of it a positive whole amount, and the amounts add up to exactly the pool. A pool distributed among *chosen* targets (*Macrocannon Barrage*: *"Distribute 5 damage among up to 5 chosen units"*) is split on declaration, as its targets are chosen (`SVY-R3`). A pool distributed among entities named without *"chosen"* (*Hammerhand Bivouac*: *"distribute 2 armor among allied units in the same zone"*) is an untargeted selection (`FR-TGT-001`): the acting player splits it as the clause resolves, among the entities that match at that moment, so Ward and two-stage verification (`FR-TGT-003`) don't apply to it. With one match it takes the whole pool and nothing is asked (`FR-MOD-001` §1). With none, the clause does nothing. On a timeout the whole pool goes on one entity (`FR-MOD-001` §6, `SVY-R52`)."
  2. `FR-TGT-003` (line 413): replace the bullet with: "**Distributed Pool Evaporation**: For a pool distributed among chosen targets (`FR-TGT-002`), any value allocated to a target that becomes illegal prior to resolution **evaporates completely** and is **not** redistributed to the remaining legal targets. A pool distributed without *"chosen"* is split as it resolves, so none of it evaporates (`SVY-R51`)."
  3. `FR-MOD-001` §6: add a bullet directly after *Mandatory Targets and Untargeted Selections*: "*Splitting a Pool (`SVY-R52`, `FR-TGT-002`)*: The whole pool goes on the first legal entity in the fixed order above, as a timed-out repeated target does (`SVY-R32`). A timed-out *Hammerhand Bivouac* gives both armor to the allied unit that entered the Main Theater earliest. A pool among *"up to N chosen"* targets selects no target on a timeout, so it is never split."
  4. Revision summary above v3.26 → v3.27, and bump the title to v3.28: `### Revision Summary (v3.27 -> v3.28): Untargeted Pools (2026-11-15)` with the item `1. **An Untargeted Pool Is Split as It Resolves (`SVY-R51`, `SVY-R52`; `FR-TGT-002`, `FR-TGT-003`, `FR-MOD-001` §6)**: a pool distributed without "chosen" is split on resolution among what matches then, and nothing of it evaporates; a timed-out pool split puts the whole pool on the first legal entity in fixed order. Card text: none.`
- **Register rows:**
  - Survey §4, add after R50:
    - `| **R51** | Is an untargeted pool ("distribute 2 armor among allied units in the same zone") split on placement or as it resolves? (Open Question 24, before the *Hammerhand Bivouac* pool task) | Hammerhand Bivouac; Macrocannon Barrage (outside M2, targeted, unchanged) | **Ruled (user, 2026-11-15).** A pool among chosen targets is split on declaration, and one without "chosen" is split as it resolves, untargeted, with nothing evaporating. Recorded in FRD 01 v3.28 as `SVY-R51` (`FR-TGT-002`, `FR-TGT-003`). |`
    - `| **R52** | What does a timeout do with a pool split? (Open Question 24) | Hammerhand Bivouac | **Ruled (user, 2026-11-15).** The whole pool goes on the first legal entity in fixed order, as `SVY-R32`. Recorded in FRD 01 v3.28 as `SVY-R52` (`FR-MOD-001` §6). |`
  - Survey card row *Hammerhand Bivouac* (line 183), notes column: `` `DISTRIBUTED_POOL`, untargeted: split as it resolves among allied units in the locked zone (`SVY-R44`, `SVY-R51`); timeout puts both armor on the first in fixed order (`SVY-R52`) ``.
  - Keyword interaction matrix: none.
- **Card text:** None.
- **Prompt to run:** `@pm-frd-drafter apply Design Outcome docs/design/outcomes/2026-11-15-wave-h-bivouac-untargeted-pool.md`

#### → `lead-architect` (build board directive)
- **Directive row:**

  | # | Directive | Source | Status |
  | :---: | :--- | :--- | :--- |
  | 157 | **An untargeted pool is split as it resolves (`SVY-R51`, `SVY-R52`; `FR-TGT-002`, `FR-TGT-003`, `FR-MOD-001` §1, §6; design note Open Question 24).** In the *Hammerhand Bivouac* pool and card tasks: *Bivouac*'s pool has no "chosen", so it is an untargeted selection made as its step resolves, like `SubjectPick` (`DecidePick`, [targeting.go:154-176](../../../pkg/engine/targeting.go)), among allied units in the zone locked on placement (`StackItem.Locked`, `SVY-R44`). It is not chosen on placement: no Ward, no `FR-TGT-003` re-check, no evaporation. The pool decision assigns positive whole amounts that add up to exactly the pool. One match takes the whole pool with nothing asked (§1), and none does nothing. A timeout puts the whole pool on the first option in the existing fixed order (§6, the same order `SubjectPick` uses), not one point at a time. Build the pool so a later targeted pool (*Macrocannon Barrage*, outside M2) can split on declaration through the same amounts check, but don't build the targeted path now. Tests: (a) two allied units in the Main Theater: the controller may put 2 on one or 1 on each, and splits like 0+2 or 3 are rejected; (b) one allied unit: it gains 2 with nothing asked; (c) no allied unit in the zone: nothing happens and nothing is asked; (d) timeout with three allied units: both armor go on the one that entered the Main Theater earliest; (e) allied units in Reserves or the Headquarters aren't offered; (f) armor gained this way persists into later turns. Mark design note Open Question 24 answered, citing this outcome. **Checked for contradictions:** none. On `autoloop/m2-wave-h`, `pkg` and `cmd` have no pool code or test (`git grep -i 'pool\|distribut'`, resource pools aside). `OpAddArmor` adds a fixed amount and stays as it is for fixed grants. *Bivouac* is only in the `cards.NotYetExecutable` entry and in neither starter deck, so the sim baseline doesn't move. | User via game-designer, 2026-11-15 ([outcome](../../../docs/design/outcomes/2026-11-15-wave-h-bivouac-untargeted-pool.md)) | waiting: docs/design/outcomes/2026-11-15-wave-h-bivouac-untargeted-pool.md applied (then open: before the *Hammerhand Bivouac* pool task) |

#### → `pm-lead` (portfolio)
- **Milestone or scope impact:** None.
