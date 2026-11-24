Status: approved (user, 2026-11-23)
Directive: #182
Rulings: SVY-R56

### 🎲 Design Outcome: A combat-kill attacker that dies in the same combat still attacks and kills (2026-11-23)

- **Question:** Does a "whenever this unit attacks and kills" ability trigger when the attacker is destroyed in the same death sweep as its defender, and does its effect still resolve? Asked by the lead as M3 design note Open Question 4 (carried from M2 Open Question 9, `FR-CBT-002` Step 3), because *Dread Manifestation* is in the Concordia starter deck.
- **Rulings:**

  | Ruling | Settles | Cards |
  | :--- | :--- | :--- |
  | `SVY-R56` | A combat-kill ability still triggers if the attacker is destroyed in the same death sweep as the defender, as Overkill (`FR-KWD-002p`) and Unbowed (`FR-KWD-002a`) do. It is placed after that sweep (`FR-TRG-004`). When it resolves after the attacker has left the battlefield, the attacker's last-known controller makes its choices. Clauses that don't change the attacker itself resolve in full, and clauses that change the attacker do nothing (`FR-ZON-004`, `FR-TGT-003`). | Dread Manifestation; Eternal Watchman (no difference: its death ends the match) |

- **Decisions (user, 2026-11-23):**
  - "option 1, same behavior as overkill and unbowed": the trigger fires and its effect resolves.
  - Overkill: already ruled the same way by `FR-KWD-002p` ("and if this unit is destroyed in the same death sweep"). No change.
  - *Dread Manifestation*'s text is unchanged. "Chosen unit in the Main Theater" can still pick an allied unit.
- **Cards affected:**
  - *Dread Manifestation* (Concordia, M3, 1 copy in `data/decks/concordia-starter.txt`): its "you may move chosen unit in the Main Theater to its controller's Reserves" resolves after a trade. Its static "-2 power" stops when it leaves the battlefield, but Step 3's retaliation was already dealt with the static applied (gates locked at Step 3, `FR-LAY-003`).
  - *Eternal Watchman* (Path Headquarters, M2): no difference, because a destroyed Headquarters ends the match at that sweep.
  - *Tempest Surge* (Concordia deck) and *Redline Infusion* (Soulless): Overkill, already ruled. *Redline Infusion*'s granted "Ready this unit." does nothing if the unit died, because it changes the attacker.
- **Alternatives rejected:**
  - The attacker must survive: combat-kill would differ from Overkill and Unbowed, and the kernel would need a second wait on the attacker's survival.
  - Rewording *Dread Manifestation* ("… and survives"): only needed with the rule above.

#### → `pm-frd-drafter` (FRD 01 and the ruling register)
- **Superseded wording:** None. FRD 01 is silent on the case. `docs/architecture/m2-ability-model.md:263` (Open Question 9) and the M3 design note's Open Question 4 and Deliverable 4 belong to the lead (directive (1)).
- **FRD 01 changes** (v3.30 → v3.31, or v3.29 → v3.30 if the Headquarters-kill outcome is applied after this one):
  - `FR-CBT-002` Step 3 *Post-Attack & Combat-Kill Triggers* (line 606): after the sentence ending "combat-kill abilities do not trigger (`SVY-R19`, `SVY-R15`).", insert: *"A combat-kill ability still triggers if the attacker is destroyed in the same death sweep as the defender, as Overkill (`FR-KWD-002p`) and Unbowed (`FR-KWD-002a`) do, and it is placed after that sweep (`FR-TRG-004`). When it resolves after the attacker has left the battlefield, the attacker's last-known controller makes its choices. Clauses that don't change the attacker itself resolve in full, and clauses that change the attacker do nothing (`FR-ZON-004`, `FR-TGT-003`): *Dread Manifestation* killed in the trade still moves its chosen unit to that unit's controller's Reserves (`SVY-R56`)."*
  - Revision summary `(… -> …): A Combat-Kill Attacker That Dies Still Triggers (2026-11-23)`:
    1. **An Attacker That Dies in the Trade Still Attacks and Kills (`SVY-R56`; `FR-CBT-002`)**: combat-kill abilities trigger even when the attacker is destroyed in the same sweep, as Overkill and Unbowed do. Clauses that don't change the attacker resolve. Card text: none.
- **Register rows:** in the survey (`docs/design/m2_starter_set_ability_survey.md`, or its path under `docs/design/survey/` if Deliverable 3a has merged):
  - §4: add `R56` with the question "Does a combat-kill attacker that dies in the same sweep still 'attack and kill'? (M3 design note Open Question 4, M2 Open Question 9)", the cards "Dread Manifestation; Eternal Watchman", and the status **"Ruled (user, 2026-11-23).** It triggers and resolves, as Overkill and Unbowed do. Recorded in FRD 01 as `SVY-R56` (`FR-CBT-002`)."
  - No keyword-matrix row changes.
- **Card text:** None.
- **Prompt to run:** `@pm-frd-drafter apply Design Outcome docs/design/outcomes/2026-11-23-m3-oq4-combat-kill-attacker-dies.md`

#### → `lead-architect` (build board directive)
- **Directive row:** #182 in [`.agents/tasks/m3/BOARD.md`](../../../.agents/tasks/m3/BOARD.md), status `waiting: docs/design/outcomes/2026-11-23-m3-oq4-combat-kill-attacker-dies.md applied (then open: before the Dread Manifestation card task)`.
  - **(1) Design notes.** Mark M3 design note Open Question 4 answered (`SVY-R56`). Deliverable 4's survey no longer needs to ask the *Dread Manifestation* question, and its row cites `SVY-R56`. Note in M2 design note Open Question 9 that it was answered in M3.
  - **(2) Kernel.** No behavior change: `HookCombatKill` already fires at the end of Step 3, before the death sweep (`pkg/engine/trigger.go:121-129`, called at `pkg/engine/combat.go:114`).
  - **(3) Tests to pin**, in the *Dread Manifestation* task or a rules task before it: `TestFR_CBT_002_CombatKillTriggerFiresWhenTheAttackerDiesInTheSameSweep` (`// Covers: SVY-R56`), with an attacker and defender that kill each other: the trigger is placed after the sweep, its non-self clause resolves, its last-known controller makes the choice, and a clause that changes the attacker does nothing. Card tests: *Dread Manifestation* killed in a trade still moves its chosen unit, an allied unit included. The Overkill test `TestFR_KWD_002p_FiresWhenTheAttackerDiesInTheSameSweep` (M2 Open Question 12) already covers the Overkill side.
  - **Checked for contradictions:** none found. `pkg/engine/combat_kill_test.go`, `combat_kill_replaced_test.go` and `redirect_test.go` have no case where the attacker dies. `Dread Manifestation` has no definition yet (on `cards.NotYetExecutable` once Deliverable 1 lands). `cardtown sim` output doesn't change.

#### → `pm-lead` (portfolio)
- **Milestone or scope impact:** None.
