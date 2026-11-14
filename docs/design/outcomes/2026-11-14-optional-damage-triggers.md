Status: approved (user, 2026-11-14)
Directive: #152
Rulings: SVY-R47, SVY-R48

### 🎲 Design Outcome: Damage triggers that could hit your own assets say "you may" (2026-11-14)

- **Question:** This came from user playtest feedback on 2026-11-13. *Devilhorn Captain* reads *"Unbowed: Deal 1 damage to chosen asset in the Main Theater."* When the only legal targets belong to its controller, it must damage one of them. That happens when the opponent has no asset in the Main Theater, or only Warded ones. A sole legal target is picked automatically (`FR-MOD-001` §1). So if the Captain is alone in the Main Theater and takes damage (from *Steel Rain*, say), it must hit itself, which fires Unbowed again, until it dies. This never came up on paper, because the current Captain has no physical proxy yet.
- **Rulings:**

  | Ruling | Settles | Cards |
  | :--- | :--- | :--- |
  | `SVY-R47` | A triggered ability is made optional with "you may", never with "up to one chosen". In a trigger, "up to N" is a count of targets. A Maneuver that does several things may use "up to" to make one of them optional. | Devilhorn Captain, Necropolis Hunt, Stardust Colossus (all reworded) |
  | `SVY-R48` | A "you may" is not asked, and its trigger is not placed, when accepting could do nothing: every step acts only on the trigger's targets, and no target slot has a legal entity at placement. | Devilhorn Captain, Necropolis Hunt, Chainbreaker Phalanx; Infested Apostle, Arkha, Eerie Remnant and Craterfall Ancient are unaffected |

- **Decisions (user, 2026-11-13 and 2026-11-14):**
  - Captain, option A ("you may"): "hitting my own Unbowed units on purpose is good, being forced to is bad"; "pinging your own guys is an intentional line of play." On being asked each time Unbowed fires: "yes, asking each time is fine." On timeouts: "When it becomes a 'may' the default behavior is 'no'."
  - *Necropolis Hunt*: "change hunt's trigger to a may as well."
  - `SVY-R47`: "'Up to one' verbiage shouldn't be applied to triggers to make them optional, only maneuvers where you're doing multiple things and want one to be optional." And: "I think this is reasonable to make templating consistent across sets. Change stardust to be a 'may' instead of up to one."
  - `SVY-R48`: "A is acceptable, it's an extension of the 'if there's only one choice, don't ask' doctrine in this case."
  - Sim baseline shift and directive placement (wave F.2, after #151): approved with the report (user, 2026-11-14).
- **Cards affected:**
  - Reworded:
    - *Devilhorn Captain*: M2, Path starter ×3.
    - *Necropolis Hunt*: M2, Path starter ×1.
    - *Stardust Colossus*: Concordia, outside M2, in no starter deck, no engine definition.
  - Text unchanged, behavior changed by `SVY-R48`:
    - *Chainbreaker Phalanx* (M2, Path starter ×2) is no longer asked when no allied unit is in your Reserves.
  - Text unchanged, behavior unchanged:
    - *Craterfall Ancient* is still asked, because an attack always leaves another asset it could target, such as a Headquarters.
    - *Infested Apostle*, *Arkha, Sun's Fury* and *Eerie Remnant* are still asked, because each has a step that isn't a target.
- **Alternatives rejected:**
  - *"Deal 1 damage to up to one chosen asset"*: "up to" doesn't make a trigger optional (`SVY-R47`).
  - *"chosen opposing asset"*: it removes the deliberate self-hit line.
  - A general rule letting a mandatory trigger be declined when only your own assets are legal: players can't see it on the card, and it would break *Saboteur*'s intended damage to your own unit.
  - Always asking the "you may": it gives a dead prompt when accepting does nothing.

#### → `pm-frd-drafter` (FRD 01 and the ruling register)
- **Superseded wording:**
  - [FRD 01:596](../../frd/01_Game_Rules_FRD.md) (`FR-KWD-002a`) quotes the Captain's clause as `Deal 1 damage to chosen asset in the Main Theater`. Change it to `You may deal 1 damage to chosen asset in the Main Theater`.
  - [FRD 01:426](../../frd/01_Game_Rules_FRD.md) (`FR-MOD-001` §7) says an optional trigger always asks. It gains the `SVY-R48` exception.
  - Survey card rows for *Necropolis Hunt* ([survey:174](../m2_starter_set_ability_survey.md)) and *Devilhorn Captain* ([survey:176](../m2_starter_set_ability_survey.md)).
  - FRD 01:87 and FRD 01:249 quote only Hunt's "deal damage equal to its power", which is still accurate, so leave them alone.
- **FRD 01 changes (v3.24 → v3.25):**
  1. **`FR-TGT-002`**: add a bullet after "Up To N" Targets Are Distinct:
     **"You May" Makes a Trigger Optional (`SVY-R47`)**: A triggered ability is made optional with *"you may"* (`SVY-R9`, `FR-MOD-001` §7), never with *"up to one chosen"*. In a trigger, *"up to N chosen"* is a count of targets, as in *Craterfall Ancient*'s *"you may deal damage equal to its power to each of up to 2 other chosen assets"*. A Maneuver that does several things may use *"up to"* to make one of them optional (*Tactical Withdrawal*: *"Return up to one chosen allied unit to its owner's hand. Draw a card."*). Card text: *Devilhorn Captain*, *Necropolis Hunt*, *Stardust Colossus*.
  2. **`FR-MOD-001` §7**: append:
     **Not Asked When It Could Do Nothing (`SVY-R48`)**: The *"you may"* is not asked, and the trigger is not placed, when accepting could do nothing: every step of the trigger acts only on its targets, and none of its target slots has a legal entity when it would be placed. This extends the Single-Option Fast-Path (§1). A trigger with a step that is not a target is still asked: the untargeted *"kill an allied unit"* of *Infested Apostle* (`SVY-R14`), or *Eerie Remnant*'s deploy of itself. *Devilhorn Captain* with no asset in the Main Theater asks nothing.
  3. **Revision summary:** add `### Revision Summary (v3.24 -> v3.25): Optional Damage Triggers (2026-11-14)` with two items:
     1. **"You May" Makes a Trigger Optional (`SVY-R47`; `FR-TGT-002`)**: triggers are made optional with "you may", never "up to one chosen". Card text: *Devilhorn Captain*, *Necropolis Hunt*, *Stardust Colossus*.
     2. **Not Asked When It Could Do Nothing (`SVY-R48`; `FR-MOD-001` §1, §7)**: a "you may" whose steps act only on targets, with no legal entity for any slot, is not asked and not placed.
- **Register rows:**
  - Survey §4 gets two new rows:
    - **R47**: the `SVY-R47` text above. Cards: "Devilhorn Captain, Necropolis Hunt, Stardust Colossus (outside M2) (reworded)". Status: **Ruled (user, 2026-11-13).** Recorded in FRD 01 v3.25 as `SVY-R47` (`FR-TGT-002`).
    - **R48**: the `SVY-R48` text above. Cards: "Devilhorn Captain, Necropolis Hunt, Chainbreaker Phalanx". Status: **Ruled (user, 2026-11-14).** Recorded in FRD 01 v3.25 as `SVY-R48` (`FR-MOD-001` §1, §7).
  - Survey card rows:
    - *Necropolis Hunt*: Hook → primitives becomes `death → (may) DEAL_DAMAGE = last-known power to chosen non-HQ asset`. Notes: `Optional (SVY-R47); not asked with no legal target (SVY-R48)`.
    - *Devilhorn Captain*: Hook → primitives becomes `→ (may) DEAL_DAMAGE 1 chosen asset in Main Theater`. Notes: `Named in FRD 01; optional (SVY-R47); not asked with an empty Main Theater (SVY-R48)`.
  - Keyword matrix: none. No `KIM-*` row names these cards (checked by grep).
- **Card text** (`rulesText` cell only). **Do not change the CSV rows for *Devilhorn Captain* or *Necropolis Hunt*.** Directive #152 changes each one together with its definition, because `TestDefs_MatchRulesText` pins it. Their new text:
  - *Devilhorn Captain* (`units-and-structures-path.csv`): `Unbowed: Deal 1 damage to chosen asset in the Main Theater.` → `Unbowed: You may deal 1 damage to chosen asset in the Main Theater.`
  - *Necropolis Hunt* (`units-and-structures-path.csv`): `When this unit is killed, deal damage equal to its power to chosen non-Headquarters asset.` → `When this unit is killed, you may deal damage equal to its power to chosen non-Headquarters asset.`

  pm-frd-drafter changes this one:
  - *Stardust Colossus* (`units-and-structures-concordia.csv`): `Whenever this unit deploys or attacks, return up to one chosen card from your discard pile to your hand.` → `Whenever this unit deploys or attacks, you may return chosen card from your discard pile to your hand.`
- **Prompt to run:** `@pm-frd-drafter apply Design Outcome docs/design/outcomes/2026-11-14-optional-damage-triggers.md`

#### → `lead-architect` (build board directive)

Directive #152 on `.agents/tasks/m2/BOARD.md`, exactly as below:

| 152 | **Devilhorn Captain and Necropolis Hunt say "you may", and a "you may" that could do nothing isn't asked (`SVY-R47`, `SVY-R48`; `FR-TGT-002`, `FR-MOD-001` §1, §7; playtest 2026-11-13).** **Wave F.2, after #151.**<br>**(1) Skipping the question, first.** [trigger.go:223](../../../pkg/engine/trigger.go) asks `accepts` for every `Optional` trigger. Don't ask, and don't place the trigger, when every step's subject is `SubjectTarget` and no target slot has a legal entity at placement. Reuse the placement legality the targeting code already computes, rather than writing a second copy. Tests: a Captain trigger with an empty Main Theater asks nothing; *Chainbreaker Phalanx* moving to the Main Theater with no allied unit in Reserves asks nothing; *Infested Apostle* with no opposing unit is still asked (its kill is a pick, `SVY-R14`); *Eerie Remnant* is still asked.<br>**(2) The cards.** In one commit, change the `rulesText` of *Devilhorn Captain* and *Necropolis Hunt* in `data/cards/units-and-structures-path.csv` (texts in the outcome file), their `Text` in [cards/path.go](../../../pkg/engine/cards/path.go) to match (`TestDefs_MatchRulesText`), and set `Optional: true` on both triggers.<br>**Merged tests this contradicts** (authorized edits for `rules-test-author`): in `cards_path_captain_test.go`, `TestCard_DevilhornCaptain_SoleLegalTargetIsNotAsked` (it now asks `DecideOptional`, then auto-picks), `TestCard_DevilhornCaptain_TimeoutDamagesExactlyOneLegalTarget` (a timeout now declines and deals nothing), and every Captain test that scripts a target answer (it now needs a "yes" first). In `cards_path_test.go`, `TestCard_NecropolisHunt_SoleLegalTargetIsNotAsked`, `TestCard_NecropolisHunt_TimeoutDamagesExactlyOneLegalTarget`, and every Hunt test that scripts a target answer. `TestCard_NecropolisHunt_NoOtherUnitOrStructureAsksNothingAndDoesNothing` stays as it is and now pins `SVY-R48` too. New tests: declining deals no damage; a lone Captain in the Main Theater that is damaged can decline, so it no longer kills itself; choosing to hit your own Unbowed unit still fires its Unbowed. To find these, I grepped both card names across `pkg/` and `cmd/` tests. No sim or `cmd/cardtown` test names either card.<br>**Sim:** `cardtown sim` uses `RandomChooser`, which now answers the new question, so the win-rate baseline shifts. The user accepted this, and `sim-compare` is expected to differ for this task. | User via game-designer, 2026-11-14 ([outcome](../../../docs/design/outcomes/2026-11-14-optional-damage-triggers.md)) | waiting: docs/design/outcomes/2026-11-14-optional-damage-triggers.md applied |

#### → `pm-lead` (portfolio)
- **Milestone or scope impact:** None. Every card stays in its milestone, and the Done-when criteria don't change.
