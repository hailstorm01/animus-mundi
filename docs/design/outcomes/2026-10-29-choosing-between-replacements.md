Status: applied: FRD 01 v3.17, 2026-10-29
Directive: #75
Rulings: SVY-R23
Audit to close: none

### 🎲 Design Outcome: Choosing between several replacements for one death (2026-10-29)

- **Question:** When several replacements apply to one death, does the affected asset's controller choose which applies (a prompt, as `FR-REP-001` "Simultaneous Replacement Conflicts" reads), or do they go in a fixed order, as replaced deaths now do (`SVY-R20`)? If it is a prompt, what does a timeout pick? `FR-MOD-001` §6 has no fallback for it. The wave G run raised this before writing T-0100 (survey R23, design note [Open Question 14](../../architecture/m2-ability-model.md#open-questions)). M2 reaches it: the Animus starter deck runs two *Shifting Impossibility*, and with both in play every allied Scratch Company death has two replacements. The choice decides which copy takes the 1 damage. With two undamaged copies (4 Health each) and four deaths at once, putting every death on one copy kills it, while splitting them two and two keeps both alive at 2 Health.
- **Rulings:**

  | Ruling | Settles | Cards |
  | :--- | :--- | :--- |
  | `SVY-R23` | The affected asset's controller chooses, death by death, which of several applicable replacements applies. A timeout picks the first source in `FR-MOD-001` §6's fixed order for untargeted selections, and a source's own replacements in printed order. | Shifting Impossibility (two copies) |

- **Decisions (user, 2026-10-29):**
  - How the physical table plays it: "Never come up." The table has not had two copies replacing the same death, so there is no table ruling to default to.
  - Direction: chose "A. Prompt per death". The player chooses for each death, as `FR-REP-001` reads, over one prompt per moment (B), the healthiest copy without a prompt (C) and the oldest copy without a prompt (D, the kernel today).
  - Timeout: chose "§6 fixed order": the same fixed order as untargeted selections (`SVY-R16`), applied to the replacements' sources, with a source's own replacements in printed order.
- **Cards affected:**
  - *Shifting Impossibility*: M2, two copies in the Animus starter deck. Card text is unchanged.
  - Every case that kills several allied Scratch Companies at one moment with both copies in play asks once per death: *Craterfall Ancient*'s death trigger (starter deck) and *Mutation Swell* (M2, not in a starter deck).
- **Alternatives rejected:**
  - B, one assignment prompt per moment: it needs a new decision shape. The user chose the prompt FRD 01 already describes.
  - C, the healthiest copy without a prompt: it removes a choice that changes outcomes, such as putting damage on a copy that is already doomed.
  - D, the oldest copy without a prompt (the kernel today): it piles every death onto one copy, which the player cannot play around.
  - The healthiest source as the timeout: it would be a second fallback rule to learn. A player who keeps timing out piles the damage on one copy, which is acceptable for a timeout.

#### → `pm-frd-drafter` (FRD 01 and the ruling register)
- **Superseded wording:** None. `FR-REP-001` "Simultaneous Replacement Conflicts & Controller Choice" (line 457) already has the controller choose. This ruling makes it a per-event prompt and adds the missing timeout.
- **FRD 01 changes (v3.16 → v3.17):**
  - **`FR-REP-001`, line 457:** replace the "Simultaneous Replacement Conflicts & Controller Choice" bullet with:
    > - **Simultaneous Replacement Conflicts & Controller Choice (`FR-REP-001`, `SVY-R23`)**: If multiple replacement effects at the same stage could apply to the exact same event on an asset, the **controller of the affected asset** chooses which replacement effect applies first. They choose separately for each event, including each of several deaths at the same moment (`SVY-R20`, `SVY-R21`), and a source that is itself lethal is still a legal choice. When only one replacement applies, nothing is asked. Once a replacement effect modifies or replaces the event such that another replacement effect's condition is no longer applicable, the remaining replacement effect does nothing. A timeout follows `FR-MOD-001` §6. *Example:* two *Shifting Impossibility* are in play, and three allied Scratch Companies are lethal at one sweep. Their controller chooses, for each death, which *Shifting Impossibility* returns that unit and takes the 1 damage.
  - **`FR-MOD-001` §6 (after line 372, "Ordering Own Simultaneous Triggers"):** add a bullet:
    > - *Choosing Between Replacement Effects (`SVY-R23`, `FR-REP-001`)*: Applies the replacement whose source comes first in the fixed order for untargeted selections above: the choosing player's own assets first, then the opponent's; within each player, the Headquarters first, then Reserves, then the Main Theater, each zone in the order its assets entered it. A single source's replacements go in printed order.
  - **Revision summary:** `### Revision Summary (v3.16 -> v3.17): Choosing Between Replacements for One Death (2026-10-29)`
    1. **Choosing Between Replacements (`SVY-R23`; `FR-REP-001`, `FR-MOD-001`)**: when several replacements apply to one death, the affected asset's controller chooses one for each death. A timeout picks the first source in the fixed order for untargeted selections. With two *Shifting Impossibility* in play, the player decides which copy takes each 1 damage.
  - **Title line:** `Master FRD v3.16` becomes `Master FRD v3.17`.
- **Register rows:**
  - **Survey §4, row R23, status cell:** **Ruled (user, 2026-10-29): a prompt for each death.** The affected asset's controller chooses which replacement applies, death by death, and nothing is asked when only one applies. A timeout picks the first source in `FR-MOD-001` §6's fixed order for untargeted selections, and a source's own replacements in printed order. Recorded in FRD 01 v3.17 as `SVY-R23` (`FR-REP-001`, `FR-MOD-001`).
  - **Survey line 18:** add "R23 on 2026-10-29" (it becomes "R20–R23 on 2026-10-29"), and change the range to `SVY-R1`–`SVY-R23`.
  - **Survey §6, Shifting Impossibility (line 139), last cell:** `R5 (revised by R20), R21, R23; a replaced kill still counts as done (SVY-R15)`.
- **Card text:** None.
- **Audit to close:** None.
- **Prompt to run:** `@pm-frd-drafter apply Design Outcome docs/design/outcomes/2026-10-29-choosing-between-replacements.md`

#### → `lead-architect` (build board directive)
- **Directive row:**

  | 75 | **Choosing between several replacements for one death: a prompt for each death (`SVY-R23`; `FR-REP-001`, `FR-MOD-001` §6; design note Open Question 14, survey R23).** T-0100 goes ahead as planned in Deliverable 6c, with the choice as a prompt. When two or more replacements apply to one death, `findReplacement` ([replace.go:11-30](../../../pkg/engine/replace.go)) asks the affected asset's controller which one applies, once per death, in the replaced-death order `SVY-R20` fixes. This holds both at the sweep and for a kill step during resolution (`SVY-R21`). A single applicable replacement applies with no question. A source that is lethal but not yet swept is a legal option. Timeout: the first source in §6's fixed order for untargeted selections (`SVY-R16`: the chooser's own assets first; within a player the Headquarters, Reserves, then Main Theater, each zone in entry order), and a source's own replacements in printed order. Today's choice, the first in battlefield order with the oldest source first, can differ from that order when sources sit in different zones. `replacedQueue` asks `findReplacement` only whether a replacement applies, so it must not ask the question there. Each death is asked about once, when it is replaced.<br>Settling tests: `TestFR_REP_001_ControllerChoosesBetweenReplacementsForEachDeath` (two guardians with 4 Health, two allied units lethal at one sweep; the controller picks guardian A for the first death and guardian B for the second, and each guardian takes 1 damage); `TestFR_REP_001_ReplacementChoiceTimesOutToFixedOrder` (two guardians, one in Reserves and one older one in the Main Theater; a timeout picks the Reserves guardian for every death); `TestFR_REP_001_OneApplicableReplacementAsksNothing`; `TestFR_REP_001_LethalSourceIsStillAChoice` (a guardian already lethal in the same pass is offered and can take the death, `SVY-R20`); `TestFR_REP_001_KillStepAsksForEachKill` (one kill step kills two allied units with two guardians in play: two questions, `SVY-R21`).<br>Contradicting merged code: none. `findReplacement`'s doc comment already leaves the choice to T-0100. No merged test contradicts it: `TestFR_REP_001_ReplacementDoesNotApplyOutsideItsFilter` and `TestFR_REP_001_ActivePlayersReplacedDeathsGoFirst` put two guardians in play, but no death has two that apply (I checked every test in `replacement_test.go` and `replacement_kill_test.go` that puts a guardian in play). A new question kind may change `cardtown sim`'s random chooser stream once *Shifting Impossibility* (T-0098) is in the deck; record `go tool task sim` as clean.<br>**Design note:** mark Open Question 14 answered, pointing at this outcome. It does not wait for FRD 01 v3.17, because the ruling is the user's. | User via game-designer, 2026-10-29 (`docs/design/outcomes/2026-10-29-choosing-between-replacements.md`) | open |

#### → `pm-lead` (portfolio)
- **Milestone or scope impact:** None. T-0100 was already planned in wave G's destruction replacement (Deliverable 6c).
- **Prompt to run:** None.
