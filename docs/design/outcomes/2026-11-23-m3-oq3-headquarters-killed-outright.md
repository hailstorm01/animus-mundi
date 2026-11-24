Status: approved (user, 2026-11-23)
Directive: #181
Rulings: SVY-R55

### 🎲 Design Outcome: Kill never affects a Headquarters (2026-11-23)

- **Question:** What does a Headquarters killed outright (not damaged to 0) do? `FR-CBT-003` left it "for the first card that can kill one" (`SVY-R46`). Asked by the lead as M3 design note Open Question 3 (carried from M2 Open Question 21), because M3 Deliverable 6a (the Headquarters-kill check) waits on it.
- **Rulings:**

  | Ruling | Settles | Cards |
  | :--- | :--- | :--- |
  | `SVY-R55` | A kill effect or kill cost never affects a Headquarters. A Headquarters is defeated only through Current Health or Effective Maximum Health at the death sweep (`FR-CBT-003`). A kill that reaches nothing but a Headquarters kills nothing, so a step that applies only "if you do" does not apply (`SVY-R14`). Card text does not need "non-Headquarters" on a kill. A "destroy a player" win condition is written as that player losing the match, not as a kill, and is ruled with the first such card. **Provisional.** | None reworded. Checked: every kill in `data/cards/` (*Consuming Earth*, *Will Breaker*, *Gemstone Titan*, *Redline Infusion*, *NO-AH-5*, *Devilhorn Sanction*, *Ambush Tactics*, *Rhapsody's Embrace*, *Feed the Despoiler*, *Mutation Swell*, *Veil Render*, *Infested Apostle*, *Blind Defile*, *Rhapsody's Hollow*, *Sinuous Growth*, *Boneyard Scavenger*, *Ruin Partisans*), and the Headquarters *Rift*, *Encampment* and *Eternal Watchman* |

- **Decisions (user, 2026-11-23):**
  - Direction: option 1, a general rule that kill never affects a Headquarters: "go with 1 for now". Recorded as **provisional**. What reopens it: a card whose effect has to be a kill of a Headquarters specifically, as opposed to "chosen opponent loses the match".
  - This narrows the 2026-11-11 decision (`SVY-R46`) not to have a general "a Headquarters can't be killed" rule. The design room the user wanted then ("door to nothingness … 'destroy target player' … leave room for a weird win con like that") stays open through a player-loss effect. *Door to Nothingness*'s current wording is itself "target player loses the game".
  - Deliverable 6a: "b". It stays in M3 as a check that no card definition's kill can match a Headquarters. Its allowlist stays empty, and only a reopened ruling can add a card to it.
- **Cards affected:** none reworded and no definition changed. No card in `data/cards/` can kill a Headquarters today. Every kill names a unit or a structure, or says "allied non-Headquarters asset" (*NO-AH-5*). *Rift* and *Encampment* only deal damage. *Eternal Watchman* is a Headquarters that kills, and nothing kills it. *NO-AH-5*'s "non-Headquarters" stays as printed: it is now redundant but harmless.
- **Alternatives rejected:**
  - A killed Headquarters is destroyed and its controller loses at the next sweep: it needs a new defeat path in the kernel, a ruling on replacements of a Headquarters' death, and a timeout guard so that a timed-out pick never kills your own Headquarters. That is the problem `SVY-R46` avoided for *Veil Render*.
  - Make it a card-writing constraint (kill text must exclude a Headquarters) and keep the outcome unruled: `FR-CBT-003` would keep an open sentence, and 6a's trigger would need rewording.
  - Drop Deliverable 6a: it is cheap, and it is the check that would catch a card that needs this ruling reopened.

#### → `pm-frd-drafter` (FRD 01 and the ruling register)
- **Superseded wording:**
  - `docs/frd/01_Game_Rules_FRD.md:610` (`FR-CBT-003` *Headquarters Destruction Defeat*): the last two sentences, "No card can kill a Headquarters outright, as opposed to reducing its health to 0 or below. What a Headquarters killed outright does is left for the first card that can kill one (`SVY-R46`)."
  - `docs/design/m2_starter_set_ability_survey.md:117` (*Noted but unreachable*): "What a Headquarters killed outright (not damaged to 0) does. … Left for the first card that can, such as a "destroy a player" win condition."
  - Left as they are: `docs/frd/01_Game_Rules_FRD.md:35` (the v3.24 revision summary, which is history); `:379` (`FR-TRN-003a`: *Veil Render*'s cost still names a unit, still true). The lead updates the M3 design note (Open Question 3, Deliverable 6a) and `docs/design/m2_ability_shapes.md:152,188` (directive (1), (4)).
- **FRD 01 changes** (v3.29 → v3.30):
  - `FR-CBT-003` *Headquarters Destruction Defeat* (line 610): replace the two sentences above with: *"A kill effect or kill cost never affects a Headquarters (`SVY-R55`). A Headquarters is defeated only through its Current Health or Effective Maximum Health at that sweep. A kill that reaches nothing but a Headquarters kills nothing, so a step that applies only if it was done does not apply (`SVY-R14`). Card text does not need to say "non-Headquarters" on a kill. A card meant to defeat a player outright says that player loses the match. It is not a kill, and it is ruled with the first such card."*
  - Revision summary `(v3.29 -> v3.30): Kill Never Affects a Headquarters (2026-11-23)`:
    1. **Kill Never Affects a Headquarters (`SVY-R55`; `FR-CBT-003`)**: kill effects and kill costs skip a Headquarters, which is defeated only through Health. A "destroy a player" win condition would be a player-loss effect, not a kill. Provisional. Card text: none.
- **Register rows:** in the survey (`docs/design/m2_starter_set_ability_survey.md`, or its path under `docs/design/survey/` if Deliverable 3a has merged):
  - §4: add `R55` with the question "What does a Headquarters killed outright do? (M3 design note Open Question 3, M2 Open Question 21)", the cards "None reworded (every kill checked; Rift, Encampment, Eternal Watchman)", and the status **"Ruled (user, 2026-11-23), provisional.** Kill never affects a Headquarters; a 'destroy a player' win condition is a player-loss effect. Recorded in FRD 01 v3.30 as `SVY-R55` (`FR-CBT-003`)."
  - §4 `R46` row (line 105), status cell: append " A killed Headquarters is ruled by `SVY-R55`."
  - *Noted but unreachable* (line 117): delete the paragraph about a Headquarters killed outright (now `SVY-R55`).
  - No keyword-matrix row changes.
- **Card text:** None.
- **Prompt to run:** `@pm-frd-drafter apply Design Outcome docs/design/outcomes/2026-11-23-m3-oq3-headquarters-killed-outright.md`

#### → `lead-architect` (build board directive)
- **Directive row:** #181 in [`.agents/tasks/m3/BOARD.md`](../../../.agents/tasks/m3/BOARD.md), status `waiting: docs/design/outcomes/2026-11-23-m3-oq3-headquarters-killed-outright.md applied (then open: before Deliverable 6a's task)`.
  - **(1) Design note.** Mark M3 design note Open Question 3 answered (`SVY-R55`, provisional). Deliverable 6a stands as written, with an allowlist that stays empty. A card that the check flags gets its definition's `Filter` narrowed to exclude a Headquarters, and its card text does not change. Only a reopened ruling adds a card to the allowlist. Its 6a trigger ("`FR-CBT-003` no longer says 'What a Headquarters killed outright does is left for the first card that can kill one'") fires when `pm-frd-drafter` applies this outcome. In 6a, the "Kernel change, if any" bullet becomes "none". Leave it to the 6a task how a narrowed filter stays consistent with `TestDefs_MatchRulesText`, given that the card text has no "non-Headquarters".
  - **(2) Kernel.** No behavior change: `kill` already skips a Headquarters (`pkg/engine/sweep.go:25-34`). Its doc comment cites `SVY-R55` as the rule.
  - **(3) Tests.** `TestFR_LAY_002_KillSkipsAHeadquarters` (`pkg/engine/if_you_do_test.go:243`) says "Pins the documented M2 limit, not a rule", which now contradicts the ruling. In the 6a task (an authorized test edit), rename it to `TestFR_CBT_003_KillNeverAffectsAHeadquarters` with `// Covers: SVY-R55`, and add a case where a kill whose only match is a Headquarters counts as not done for "if you do" (`SVY-R14`).
  - **(4)** In `docs/design/m2_ability_shapes.md`, change line 188 ("A Headquarters killed outright | Unruled (`SVY-R46`)") to "Kill never affects a Headquarters (`SVY-R55`)", and line 152's note to cite `SVY-R55`.
  - **Checked for contradictions:** on `main`, `git grep -i headquarters -- 'pkg/**/*_test.go'` filtered for "kill" finds the test in (3) and *Veil Render*'s tests (`pkg/engine/cards_animus_veil_test.go:101,137`), which agree with the ruling: a Headquarters is never killed. The `KindHeadquarters` filters in `kill_cost_test.go:112` and `kill_count_cost_test.go:228` are on damage steps, not kills. No card definition kills a Headquarters (`data/cards/` text checked above). `cardtown sim` output doesn't change.

#### → `pm-lead` (portfolio)
- **Milestone or scope impact:** None. Deliverable 6a stays in M3 as the design note gives it, with an allowlist that stays empty.
