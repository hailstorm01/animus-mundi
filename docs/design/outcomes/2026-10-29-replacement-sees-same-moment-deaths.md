Status: approved (user, 2026-10-29)
Directive: #74
Rulings: SVY-R20, SVY-R21, SVY-R22; SVY-R5 (revised)
Audit to close: none

### 🎲 Design Outcome: A replacement sees every death at the same moment (2026-10-29)

- **Question:** Two destruction-replacement cases came up while T-0096 and T-0097 were being planned (design note [Open Question 13](../../architecture/m2-ability-model.md#open-questions)). (1) When *Shifting Impossibility* is lethal in the same death sweep as the deaths it replaces, does it still replace them? (2) When one kill step kills several allied Scratch Companies mid-resolution, does Impossibility replace all of them? The user ruled on both in a lead-architect session. Recording them here led to a revision of `SVY-R5` and two knock-on rulings.
- **Rulings:**

  | Ruling | Settles | Cards |
  | :--- | :--- | :--- |
  | `SVY-R20` (revises `SVY-R5`) | A destruction replacement applies to every death in a death sweep, including deaths whose byproducts leave its source lethal. The source is never swept partway through a sweep, and no player orders the replaced deaths. | Shifting Impossibility, Craterfall Ancient |
  | `SVY-R21` | A replacement sees every death at the same moment as its source's own death. Deaths caused while one stack item resolves, and the death sweep right after it, are the same moment. | Shifting Impossibility, Mutation Swell, Infested Apostle; the kill costs of Rhapsody's Embrace, Feed the Despoiler and Veil Render |
  | `SVY-R22` | When two players' destruction replacements apply in one sweep, the active player's replaced deaths are processed first, then the non-active player's. | none in M2 (only Animus prints a destruction replacement) |

- **Decisions (user, 2026-10-29):**
  - Source lethal in the same sweep (the *Craterfall Ancient* case), chosen from "all three, one, none": "All three, then it dies."
  - Several kills from one kill step: "I think we use this general principle from MTG's harvester of souls: 'If another nontoken creature dies at the same time as Harvester of Souls, its ability triggers for that creature.' So if the 3 scratch companies and shifting are all dying at the same time, shifting impossibility's replacement sees them."
  - Principle wording: chose "same-moment deaths" (a replacement sees every death at the same moment as its source's own) over "a lethal asset stays until it is swept".
  - "Same moment" means everything up to the next sweep: deaths caused while one item resolves and the sweep right after it. Assets that a byproduct newly leaves lethal die in the sweep's next pass, which is a new moment.
  - Asymmetry (an Impossibility that is lethal at the start of a sweep saves more deaths than a healthy one): chose "Revisit `SVY-R5`". Impossibility is never swept partway through a sweep. It replaces every death in the sweep, then dies.
  - Order prompt: chose "Drop the prompt". Replaced deaths are processed oldest first, which was T-0096's timeout order.
  - *Mutation Swell* with Impossibility at 1 Health, killing 4 allied Scratch Companies: chose "Confirm". All 4 return to hand, Impossibility dies at the sweep after the Maneuver, and the player still deploys 2 units from their discard pile (`SVY-R15`).
  - Two players' replacements in one sweep: chose "Rule now, as built", with the active player's deaths first. The order prompt decision above replaces the "each player orders their own" half.
- **Cards affected:**
  - *Shifting Impossibility*: M2, in the Animus starter deck.
  - *Craterfall Ancient*: M2, in the Animus starter deck. Its death trigger, 4 damage to all assets, is the case that reaches `SVY-R20`.
  - *Infested Apostle*, *Rhapsody's Embrace*, *Veil Render*: M2, in the Animus starter deck. *Feed the Despoiler*: M2, not in a starter deck. Each kills one unit, so `SVY-R21` changes nothing for them.
  - *Mutation Swell*: M2 (wave F), not in a starter deck. It is the only M2 card whose kill step kills several assets. No Animus or Path card says "kill all X".
  - Card text is unchanged.
- **Alternatives rejected:**
  - "A lethal asset stays until it is swept" as the principle's wording: the user preferred the Harvester of Souls form. "Same moment" was then defined so that it gives the same answers in M2.
  - Keeping `SVY-R5`'s mid-sweep destruction: it made a damaged Impossibility save more deaths than a healthy one.
  - Keeping the order prompt: under `SVY-R20` the order changes no outcome in M2, so the prompt only added friction.
  - Counting only real deaths for *Mutation Swell*: it would carve an exception out of `SVY-R15`.
  - Leaving the two-player order open: T-0096 already does it this way, so ruling now keeps FRD 01 and the code in step.

#### → `pm-frd-drafter` (FRD 01 and the ruling register)
- **Superseded wording:**
  - FRD 01 v3.15, `FR-REP-001` "Sequential Per-Entity Death Replacement" (line 455): "replace each allied Scratch Company's death sequentially until *Shifting Impossibility* itself takes lethal damage."
  - FRD 01 v3.15, `FR-REP-001` "Ordering and Mid-Sweep Destruction" (line 456): the whole bullet.
  - Survey §4 row R5 (line 64): the status cell.
  - Leave FRD 01's older revision summaries as they are (line 65, "5. Ordering and Mid-Sweep Destruction"). They are history.
- **FRD 01 changes (v3.15 → v3.16):**
  - **`FR-REP-001`, line 455:** replace the "Sequential Per-Entity Death Replacement" bullet with:
    > - **Per-Entity Death Replacement (`FR-REP-001`)**: When one effect or one death sweep would destroy several assets at once, each asset's prospective destruction is a separate death event, and a replacement effect such as *Shifting Impossibility* applies to each one it matches.
  - **`FR-REP-001`, line 456:** replace the "Ordering and Mid-Sweep Destruction" bullet with:
    > - **Same-Moment Deaths (`SVY-R21`, `SVY-R20`, `FR-REP-001`, `FR-LAY-002`)**: A destruction replacement applies to every death that happens at the same moment as its source's own death. Deaths caused while one stack item resolves, and the death sweep right after it, happen at the same moment. Kills during resolution happen inline (`FR-LAY-002`), and lethal damage waits for that sweep, so the source is still on the battlefield for every one of them, even after a byproduct has left it lethal. *Example:* one kill step kills two allied Scratch Companies while *Shifting Impossibility* has 1 Current Health. Both kills are replaced (returned to hand, 1 damage each to *Shifting Impossibility*), and *Shifting Impossibility* is destroyed at the sweep after the item.
    > - **Replacement Within a Death Sweep (`SVY-R20`, `SVY-R22`, `FR-REP-001`, `FR-LAY-002`)**: The assets that are lethal when a death sweep starts die simultaneously. Every death among them that a replacement applies to is replaced, one at a time, including its byproduct events, whether or not the replacement's source is among the lethal assets or a byproduct leaves it lethal. Replaced deaths are processed in a fixed order: the active player's assets first, then the non-active player's, and each player's oldest asset first. Then the remaining lethal assets are destroyed together. The sweep then repeats: assets that a byproduct newly left lethal die in the next pass, which is a new moment. *Example:* *Craterfall Ancient* is killed and deals 4 damage to all assets. *Shifting Impossibility* (Maximum Health 4) and three allied Scratch Companies are lethal at the same sweep. All three deaths are replaced, and *Shifting Impossibility* is then destroyed with the sweep's other deaths. *Example:* *Shifting Impossibility* (Maximum Health 4, undamaged) and five allied Scratch Companies: the five Scratch Companies are lethal at a sweep, and *Shifting Impossibility* is not. All five deaths are replaced, leaving *Shifting Impossibility* at −1 Current Health, and the sweep's next pass destroys it.
  - **`FR-REP-001`, "Replaced Actions Still Count as Done":** add at the end:
    > This applies to every kill an effect makes: if *Mutation Swell* kills four allied Scratch Company units and *Shifting Impossibility* replaces all four, four units were killed this way, and its controller deploys two unit cards from their discard pile (`SVY-R21`).
  - **Revision summary:** `### Revision Summary (v3.15 -> v3.16): A Replacement Sees Every Death at the Same Moment (2026-10-29)`
    1. **Same-Moment Deaths (`SVY-R21`; `FR-REP-001`, `FR-LAY-002`)**: a destruction replacement applies to every death at the same moment as its source's own, meaning deaths during one stack item and the sweep right after it. One kill step that kills two Scratch Companies while *Shifting Impossibility* has 1 Health has both replaced.
    2. **No Destruction Partway Through a Sweep (`SVY-R20`, revising `SVY-R5`; `FR-REP-001`, `FR-LAY-002`)**: a replacement applies to every death in a sweep, even when its source is lethal too or a byproduct leaves it lethal, and the source dies afterwards. The player no longer orders the replaced deaths: they go in a fixed order.
    3. **Two Players' Replacements (`SVY-R22`; `FR-REP-001`)**: the active player's replaced deaths are processed first.
  - **Title line:** `Master FRD v3.15` becomes `Master FRD v3.16`.
- **Register rows:**
  - **Survey §4, row R5, status cell:** **Revised (user, 2026-10-29) by R20.** A replacement applies to every death in the sweep, and the source dies afterwards, so it is never destroyed partway through a sweep. The controller no longer orders the deaths. Originally ruled 2026-10-19, with mid-sweep destruction and a controller-chosen order.
  - **Survey §4, new row R20:**
    - Question: when a destruction replacement's source is lethal in the same death sweep as the deaths it applies to, or a byproduct leaves it lethal partway through, does it still replace them? (Open Question 13, T-0096)
    - Cards: Shifting Impossibility; Craterfall Ancient.
    - Status: **Ruled (user, 2026-10-29): yes, all of them.** "All three, then it dies." The source dies after the sweep's replaced deaths. One left lethal by a byproduct dies in the sweep's next pass. Replaced deaths go in a fixed order, oldest asset first, with no prompt. Revises R5. Recorded in FRD 01 v3.16 as `SVY-R20` (`FR-REP-001`, `FR-LAY-002`).
  - **Survey §4, new row R21:**
    - Question: does a replacement see several kills from one kill step, or kills made while its source is already lethal, before the next sweep? (Open Question 13, T-0097)
    - Cards: Shifting Impossibility with Mutation Swell; Infested Apostle; the kill costs of Rhapsody's Embrace, Feed the Despoiler and Veil Render.
    - Status: **Ruled (user, 2026-10-29): yes.** A replacement sees every death at the same moment as its source's own, after Harvester of Souls. Deaths during one stack item and the sweep right after it are the same moment. *Mutation Swell*'s replaced kills still count toward its deploys (`SVY-R15`). Recorded in FRD 01 v3.16 as `SVY-R21` (`FR-REP-001`).
  - **Survey §4, new row R22:**
    - Question: when two players' destruction replacements apply in one sweep, whose replaced deaths go first? (Open Question 13, T-0096)
    - Cards: none in M2. Only Animus prints a destruction replacement.
    - Status: **Ruled (user, 2026-10-29): the active player's first**, as for triggers (`FR-TRG-003`). Recorded in FRD 01 v3.16 as `SVY-R22` (`FR-REP-001`).
  - **Survey line 18:** add "R20–R22 on 2026-10-29", note that R5 was revised on 2026-10-29, and change the range to `SVY-R1`–`SVY-R22`.
  - **Survey §6 card rows:**
    - Shifting Impossibility (line 135), last cell: `R5 (revised by R20), R21; a replaced kill still counts as done (SVY-R15)`.
    - Mutation Swell (line 132), last cell: append `; replaced kills count toward its deploys (SVY-R21, SVY-R15)`.
- **Card text:** None.
- **Audit to close:** None.
- **Prompt to run:** `@pm-frd-drafter apply Design Outcome docs/design/outcomes/2026-10-29-replacement-sees-same-moment-deaths.md`

#### → `lead-architect` (build board directive)
- **Directive row:**

  | 74 | **A replacement sees every death at the same moment, and nothing is swept partway through a sweep (`SVY-R20` revising `SVY-R5`, `SVY-R21`, `SVY-R22`; `FR-REP-001`).** This changes merged T-0096 code. `sweepDeaths` runs an interim sweep after each replaced death ([damage.go:143-171](../../../pkg/engine/damage.go)), so a source that a byproduct leaves lethal is destroyed before the next death. Under `SVY-R20`, every queued death in the pass is replaced, the pass's unreplaced deaths are destroyed together, and only then does the sweep take up what byproducts newly left lethal, as its next pass. `replacedQueue` asks each player to order their replaced deaths ([replace.go:32-60](../../../pkg/engine/replace.go)). Under `SVY-R20` there is no prompt: the active player's deaths first (`SVY-R22`), then each player's oldest asset first. `findReplacement`'s "a source that is itself lethal still applies" stays. Its doc comment and `deathSweep`'s should drop "in their controller's order".<br>**Tests contradicted** (authorized rewrites for the task that applies this, since they are T-0096's protected tests): `TestFR_REP_001_SourceLeftLethalByAByproductStopsApplying` (the second unit is now replaced, and the guardian dies in the next pass); `TestFR_REP_001_ControllerOrdersTheReplacedDeaths` (no order decision, and all three are replaced); `TestFR_REP_001_ReplacedDeathOrderFallsBackToOldestFirst` (all three replaced, and the order is oldest first without a prompt); `TestFR_REP_001_TheAffectedAssetsControllerOrdersTheirDeaths` (no order decision for either player, and player 1's deaths still go after player 0's). `TestFR_REP_001_SourceLethalInTheSameSweepReplacesEveryDeath` and `TestFR_REP_001_ByproductDamageFiresTheSourcesDamageTrigger` still hold. I read every test in `replacement_test.go`. Settling tests: `TestFR_REP_001_ByproductLethalSourceStillReplacesEveryDeathInThePass` (guardian with 4 Health and five allied units at one sweep: all five are replaced, and the guardian is destroyed after the fifth replaced death); `TestFR_REP_001_ReplacedDeathsAreNotOrderedByThePlayer` (no `DecideOrder`, processed oldest first); `TestFR_REP_001_ActivePlayersReplacedDeathsGoFirst` (`SVY-R22`).<br>**T-0097 (kill steps, `SVY-R21`):** its scope stands. Add `TestFR_REP_001_OneKillStepReplacesEveryKill` (two allied Scratch Companies killed by one step while *Shifting Impossibility* has 1 Health: both are replaced, and Impossibility is destroyed at the sweep after the item). Once T-0098 lands, add a *Mutation Swell* scenario to wave F's task: four kills replaced, and two units still deployed from the discard pile (`SVY-R15`).<br>**Design note:** rewrite Open Question 13 as answered, pointing at this outcome, and update Deliverable 6c's T-0096 line ("in its controller's order" is gone). This can go ahead of FRD 01 v3.16, because the rulings are the user's. Removing a decision changes the random chooser's stream, so `cardtown sim` output may change: run `go tool task sim` and record that it is clean. Tier `sonnet`, before T-0097 (one task, or folded into T-0097: the lead's call). | User via game-designer, 2026-10-29 (`docs/design/outcomes/2026-10-29-replacement-sees-same-moment-deaths.md`) | open |

#### → `pm-lead` (portfolio)
- **Milestone or scope impact:** None. It is a rules revision inside wave G's destruction replacement, which is already planned before wave C (Deliverable 6c).
- **Prompt to run:** None.
