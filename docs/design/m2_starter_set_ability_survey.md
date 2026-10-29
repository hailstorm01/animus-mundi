# M2 Starter Set Ability Survey (Animus vs Path)

- **Date:** 2026-10-19
- **Source:** `data/cards/*.csv` at `8cb8bc2`; [FRD 02](../frd/02_Card_Data_Pipeline_FRD.md) Domains 3–4 (frozen design intent); [FRD 01](../frd/01_Game_Rules_FRD.md) v3.4.
- **Purpose:** map every Animus and Path card onto FRD 02's ability model, find what FRD 02 cannot express, and order the M2 kernel work by how many cards each step unlocks. This is the first draft of M2 deliverable 5 (the ability write-up).

FRD 01 is the rules source. This document names requirement IDs and never restates the rules.

---

## 1. Summary

- **82 cards in M2:** 40 Animus, 40 Path, and the 2 starter Headquarters (Animoceps Rager, Eternal Watchman). Rift and Encampment moved to M3 on 2026-10-19; their rows stay in §6 for reference. **9 already execute** (standalone keywords only), leaving **73 to model**.
- **By FRD 02 ability type:** 33 triggered, 25 Maneuver effects, 5 activated, 8 continuous-only, 2 replacement watchers.
- **FRD 02 fit:** all 73 cards map onto FRD 02's six ability types and existing primitives. There is **one real schema gap** (triggers that work off the battlefield, §3). The other problems were about *when* targets are chosen (§4, R3–R4); since 2026-10-20 every target is locked on placement.
- **Ready-made test cases:** 6 starter cards already have full FRD 02 blueprints in [the benchmark reference](../frd/reference/02_Benchmark_Card_Blueprints.md): Infested Apostle, Funeral Fungus, Veil Render, Mutation Swell, Allshan's Battle Trance and Animoceps Spore Cloud.
- **FRD 01 reach:** the starter set reaches **27 of the 30** unpinned FRD 01 requirements, including `FR-SET-001` because starter decks must be construction-legal (Q1). It does not reach `FR-ENT-102` (tucking), `FR-KWD-002n` (Regroup) or `FR-TRG-006` (delayed triggers).
- **Rulings:** R1–R6 (plus the R4a follow-up) were ruled on 2026-10-19, R7 on 2026-10-20, R8–R11 on 2026-10-21, R13–R17 on 2026-10-23, R18 on 2026-10-26, R12 and R19 on 2026-10-28, and R20–R22 on 2026-10-29; R3 was revised on 2026-10-20 (targets locked on placement, with new text for Animoceps Spores and Steel Rain), and R5 on 2026-10-29. They are recorded in FRD 01 as `SVY-R1`–`SVY-R22`; R12 (Berserk is replaced by Overkill) is in v3.14. Nothing in the starter set is blocked.

---

## 2. Build Order (Shape Waves)

Each wave adds the shapes its cards need. A card is ready once every earlier wave it depends on has landed. Waves C–G only depend on A and B, so they can be reordered.

| Wave | Shapes added | FRD 01 rules reached | Cards | Count |
| :---: | :--- | :--- | :--- | :---: |
| **A1** | Triggered-ability framework: self hooks for deploy, death and attack, APNAP placement and cascading. Untargeted primitives: `SCRATCH`, `DRAW_CARD`, `RETURN_TO_HAND` (self, discard → hand), `PUT_CARD_ON_DECK`. `turnOwner` condition. Last-known state for death triggers. | `FR-TRG-003`, `FR-TRG-004`, `FR-ZON-004` | Despoiler Larva, Spoil Forest Creeper, Fungal Evangelist, Titanic Spore Carrier, Portent Scrivener, Despoiler Worm, Stormfront of Ideals, Unceasing Tide | 8 |
| **A2** | Targeting with filters and `up to N`, optional ("you may") triggers. Primitives `DEAL_DAMAGE` (static or from own power), `HEAL`, `ADD_ARMOR`. `toZone` deploy condition. | `FR-TGT-001`, `FR-TGT-002`, `FR-MOD-001`, `FR-KWD-002d` | Chainbreaker Sappers, Saboteur, Devilhorn Guerrillas, Ironclad Colossus, Spoil Forest Colossus, Necropolis Hunt, Craterfall Ancient | 7 |
| **B** | Damage-taken hook (Unbowed and HQ "is dealt damage", `maxPerTurn`). Cornered gate on triggers. Exhaust, turn-start, combat-kill and opposing-attack hooks. 'You may X. If you do, Y' actions on resolution (`SVY-R14`). Distributed pools. Modal choice. `REVEAL_HAND`, `DISCARD_CARD`. | `FR-KWD-002a`, `FR-KWD-002e`, `FR-TRG-002`, `FR-TRG-005`, `FR-LAY-003`, `FR-CAP-001` | Hammerhand Brigands, Devilhorn Captain, Chainbreaker Irregulars, Allshan Chainbreaker Revenant, Animoceps Rager, Animoceps Shambler, Hammerhand Bivouac, Eternal Watchman, Blind Defile, Infested Apostle ★, Arkha Sun's Fury, Drifting Concept, Rogue Fleshshaper | 13 |
| **C** | Maneuvers and additional enactment costs. `KILL`, `MOVE_ZONE`, `MODIFY_STAT` (power and max health, until end of turn or end of opponent's turn), "all X" selectors, delegated damage source. | `FR-MNV-001`, `FR-TRN-003a`, `FR-LAY-003` | Disorienting Pulse, Feed the Despoiler, Funeral Fungus ★, Conversion Spike, Deafening Aria, Ensnaring Mist, Rhapsody's Embrace, Despoiler's Emergence, Rapturous Sermon, Chainbreaker's Tenacity, Fortify, Tactical Withdrawal, Blindside, Improvised Orders, Devilhorn Sanction, Pincer Tactics, Poisoned Rations, Ambush Tactics, Animoceps Spores, Steel Rain. Spores and Steel Rain also need resolution-time targeting (`SVY-R3`, `SVY-R4`). | 20 |
| **D** | Activated abilities with resource, exhaust, self-kill and discard costs. `GAIN_CONTROL`, `SEARCH_DECK`. | `FR-ACT-001`, `FR-ACT-002`, `FR-DEK-001`, `FR-DEK-002`, `FR-ZON-003` | Rhapsody's Hollow, Sinuous Growth (R2), Cliff of Portents, Boneyard Scavenger, Ruin Partisans (R2) | 5 |
| **E** | Continuous auras in Layer 4: power, keyword grants and granted Cornered abilities. Hand-active cost reduction. | `FR-LAY-001`, `FR-LAY-003`, `FR-CST-001` | Animoceps Stalk, Animoceps Node, Melodic Pummeler, Hulking Incongruity, Brigand Highlander, Macabre Dancers, Ancestral Monument, Boneyard Amalgam | 8 |
| **F** | Free deploys and zone crossing: deploy permission from discard, `DEPLOY_FREE`, `LOOK_TOP_CARDS` partition, `GRANT_KEYWORD`, `MORPH_SCRATCH`, zone-move hook, off-battlefield triggers (§3 G1). | `FR-TRN-003a`, `FR-CST-001`, `FR-ENT-103` | Veil Render ★, Mutation Swell ★, Reaver's Rush, Allshan's Battle Trance ★, Devilhorn Scouting Party, Forced Mutation, Eerie Remnant, Solarheart Monk, Chainbreaker Phalanx (R1), Broken Horn Armorer (R1) | 10 |
| **G** | Replacement watchers: destruction replacement and damage redirection. | `FR-REP-001`–`FR-REP-003` | Animoceps Spore Cloud ★, Shifting Impossibility (R5) | 2 |

★ = has an FRD 02 benchmark blueprint. Baseline requirements every match exercises (`FR-CRD-101`, `FR-SET-003`, `FR-ZON-001`) need only a named test, so they can be added in any wave.

**Already executable (9):** Spoil Forest Haunt, Flickering Memory, Sigiled Stoneform, Despoiler Spawn, Spoil Forest Amalgam, Chainbreaker Sneak, Trailblazer, Improvised Bastion, Hammerhand Brigade.

---

## 3. FRD 02 Gaps (kernel decides; no ruling needed)

| ID | Gap | Cards | Proposed kernel behavior |
| :--- | :--- | :--- | :--- |
| **G1** | FRD 02 gives an `activeZone` only to continuous auras. A **triggered** ability that works from hand or discard has no representation, although FRD 01 `FR-ZON-004` defines enter-discard triggers. | Eerie Remnant, Solarheart Monk | Give triggered abilities an active zone (battlefield by default). Feed this into the FRD 02 rewrite. |
| **G2** | Aura-granted keywords sit in **Layer 4** in FRD 02 `FR-ABL-004`, but FRD 01 `FR-LAY-001` puts keyword grants in **Layer 3**. FRD 01 wins. No starter card can currently tell the difference. | Hulking Incongruity, Macabre Dancers, Ancestral Monument | Follow FRD 01 (Layer 3). Record the difference for the rewrite. |
| **G3** | Self-sacrifice costs ("kill this unit") have no dedicated cost type, unlike `EXHAUST_SELF`. They are expressible as `KILL_ASSET` with selector `SELF`. | Rhapsody's Hollow, Sinuous Growth, Boneyard Scavenger, Ruin Partisans | Model as a kill cost on self. Rename only if the code reads badly. |

---

## 4. Ruling Questions

Rulings follow the MTG-like intuitions used in earlier rulings. Ruled items are recorded in FRD 01 (v3.5 onward) under the `SVY-` IDs, and FRD 01 is the authority for them. Any future open item gets an `OPEN(SVY-…)` marker on the kernel's chosen default until it is ruled.

| ID | Question | Cards | Status |
| :--- | :--- | :--- | :--- |
| **R1** | Does "moves to a new zone" or "moves to the Main Theater" include **entering** the battlefield? | Broken Horn Armorer, Chainbreaker Phalanx | **Ruled: no.** Only March and effects that say "move" (`SVY-R1`, `FR-ZON-002`). |
| **R2** | Can an activated ability target the card it sacrificed as its cost? | Sinuous Growth, Ruin Partisans | **Ruled: no.** Targets and costs are committed together, before the card reaches the discard pile (`SVY-R2`, `FR-TGT-001`, `FR-ACT-001`). |
| **R3** | Is "then return chosen allied Scratch Company unit" targeted before or after the Scratch? | Animoceps Spores, Steel Rain | **Ruled (revised 2026-10-20): neither; targets are locked on placement.** No target is chosen during resolution. Spores now reads "return an allied Scratch Company unit", an untargeted pick made after the Scratch from every allied Scratch Company unit. Steel Rain's discard is now an additional cost, so its targets are chosen with it (`SVY-R3`, `FR-TGT-001`). Originally ruled "after" on 2026-10-19. Wording narrowed 2026-10-23 (`SVY-R14`): an earlier clause may decide whether a targeted clause applies, but not which or how many targets. |
| **R4** | One target taking 2 per discarded card, or a separate choice per card? | Steel Rain | **Ruled: a separate choice per card**, made on declaration together with the discard, which is now an additional cost (`SVY-R4`, revised with R3; `FR-TGT-001`). |
| **R4a** | Follow-up to R4: may the same asset be chosen for more than one discarded card (for example 3 discards → 6 damage to one asset)? | Steel Rain | **Ruled: yes.** Each choice is independent, and each is its own damage instance (`SVY-R4`, `FR-TGT-001`). |
| **R5** | Shifting Impossibility's replacement deals 1 damage to itself **during** a death sweep. If that is lethal, does it die in the same sweep (the sweep repeats until nothing changes) or only after the next stack item? | Shifting Impossibility | **Revised (user, 2026-10-29) by R20.** A replacement applies to every death in the sweep, and the source dies afterwards, so it is never destroyed partway through a sweep. The controller no longer orders the deaths. Originally ruled 2026-10-19, with mid-sweep destruction and a controller-chosen order. |
| **R6** | Can "Choose two" pick the same mode twice? | Ambush Tactics | **Ruled: no**, unless the card explicitly allows it (`SVY-R6`, `FR-MOD-001`). |
| **R7** | When a player times out (or has no chooser) while ordering their own simultaneous triggers, what order are they placed in? `FR-MOD-001` §6 had no fallback for this. | Any two simultaneous triggers of one player | **Ruled: a fixed default order**: by source position (Headquarters, Reserves, Main Theater, each in entry order), then keyword triggers before printed ones (`SVY-R7`, `FR-MOD-001`). Accepted 2026-10-19 as the lead's recommendation, confirmed 2026-10-20. |
| **R8** | If the defender or the attacker of a declared attack changes controller before the attack resolves (so both are now under one player), does the attack fizzle? `FR-CBT-002` did not say. The kernel compares the defender's controller with the declaring player (`resolveAttack`), not with the attacker's current controller. | No current card. *Rhapsody's Hollow* is an activated ability, so it can't be used while an Attack is on the stack (`FR-ACT-001`, `FR-QUE-001`). | **Ruled 2026-10-21** (user via game-designer): the defender is re-verified as a target, including allegiance to the attacker's current controller; a side change on either side fizzles the attack with no Berserk cleave. Recorded in FRD 01 v3.8 as `SVY-R8`. *The Berserk clause is superseded by `SVY-R12`.* |
| **R9** | When does a player answer an optional ("you may") trigger: when it is placed on the stack, or when it resolves? `FR-MOD-001` §6 fixes only the timeout (decline). | Craterfall Ancient; later Arkha, Sun's Fury and Blind Defile | **Ruled (user, 2026-10-21): on placement**, before its targets are chosen, matching `SVY-R2`/`SVY-R3`: every commitment is made on placement. A declined trigger chooses no targets and does nothing. A timeout declines. Recorded in FRD 01 v3.9 as `SVY-R9`. |
| **R10** | "Damage equal to its power": which power, if the source is still on the battlefield when the trigger is placed but gone when it resolves? `FR-ZON-004` covers only leaves-battlefield triggers. | Craterfall Ancient, Necropolis Hunt | **Ruled (user, 2026-10-21): the source's current effective power when the step resolves; if the source has left the battlefield, its last-known effective power** (`FR-ZON-004`). Recorded in FRD 01 v3.9 as `SVY-R10`. |
| **R11** | "Each of up to N chosen assets": may the same asset fill more than one slot? | Craterfall Ancient; later Pincer Tactics, Encampment | **Ruled (user, 2026-10-21): no, the assets are distinct.** Repeats stay a "for each"-only feature (`SVY-R4`). Recorded in FRD 01 v3.9 as `SVY-R11`. |
| **R12** | Berserk N needs a design review. Its secondary-target rules have generated repeated edge-case rulings (`KIM-B01`–`KIM-B07`, `SVY-R8`): mandatory secondary, cleave on attacker-side fizzle but not defender-side, re-verification and Ward on the secondary. Should Berserk be simplified or redesigned? | No M2 card. *Tempest Surge* (Concordia) has it printed and *Redline Infusion* (Soulless) grants it; both are M3. | **Ruled (user, 2026-10-28): Berserk is retired and replaced by Overkill (`FR-KWD-002p`).** Recorded in FRD 01 v3.14 as `SVY-R12`. |
| **R13** | When an Unbowed asset takes lethal damage, does its Unbowed still trigger, and when is the trigger placed? `FR-KWD-002a` said "before the asset is destroyed", which conflicted with `FR-TRG-004` (triggers are placed after the resolving item and its death sweeps). | Hammerhand Brigands, Devilhorn Captain, Chainbreaker Irregulars, Macabre Dancers, Allshan, Chainbreaker Revenant | **Ruled (user, 2026-10-23): it still triggers, even if the asset is destroyed, for damage from any source.** `FR-TRG-004` governs: the trigger is placed after the death sweep, in the same APNAP window (`FR-TRG-003`) as any death triggers from the same event. Recorded in FRD 01 v3.10 as `SVY-R13` (`FR-KWD-002a`, `FR-TRG-004`). |
| **R14** | "You may X. If you do, Y": when is X carried out, and may X select the trigger's own source? (design note Open Question 8(a)) | Infested Apostle, Arkha, Sun's Fury; later Redlined Brawler (M3) | **Ruled (user, 2026-10-23): X is carried out on resolution.** "You may" is answered and Y's target locked on placement (`SVY-R9`, `SVY-R3`). X is an untargeted selection on resolution, not a cost, and the Apostle may kill itself. Y applies only if X was carried out. With no match or a timeout on X, Y does nothing. This overrides the lead's recommendation (a kill on placement, as a cost). Recorded in FRD 01 v3.11 as `SVY-R14` (`FR-TGT-001`, `FR-TRG-005`, `FR-MOD-001`). |
| **R15** | Does an action or cost still count as done when a replacement effect replaces its event? (knock-on of Open Question 8(a)) | Infested Apostle with Shifting Impossibility; the kill costs of Veil Render, Rhapsody's Embrace and Feed the Despoiler | **Ruled (user, 2026-10-23): yes.** The player carried out the action, so "if you do" is met and a cost stays paid, even though the replaced event never occurs and triggers watching for it do not fire. Recorded in FRD 01 v3.11 as `SVY-R15` (`FR-REP-001`, `FR-TRN-003a`). |
| **R16** | How do the fast path and the timeout work for choosing a card from your own hand, and may any timeout fallback be random? (Open Question 8(b)) | Chainbreaker Irregulars; every mandatory target and untargeted selection (Animoceps Spores, Mutation Swell, Infested Apostle's kill) | **Ruled (user, 2026-10-23): no fallback is random.** Mandatory targets and untargeted selections time out to the first legal options in a fixed order: the choosing player's assets first, then Headquarters, Reserves, Main Theater by arrival, and piles from the top. Your own Hand is not an inspection: one card is selected automatically, and a timeout selects the card most recently added (for Irregulars, the card just drawn). Replaces the random mandatory-target fallback. Recorded in FRD 01 v3.11 as `SVY-R16` (`FR-MOD-001`). |
| **R17** | What counts as "this unit exhausts", and when are its triggers placed? (Open Question 8(c)) | Animoceps Shambler | **Ruled (user, 2026-10-23): every change from Readied to Exhausted, for any reason.** Entering the battlefield Exhausted is not one. Triggers from March, Attack or Deploy Scratch Company are placed on top of the action. Triggers from resource generation are placed after it completes. Recorded in FRD 01 v3.11 as `SVY-R17` (`FR-CAP-001`). |
| **R18** | When a keyword sits behind Cornered ("Cornered: Aggressive 3."), does the asset have that keyword while Cornered is not active? (review of wave B run 1, board directive #60) | Allshan, Chainbreaker Revenant; Macabre Dancers; Ancestral Monument | **Ruled (user, 2026-10-26): no.** The asset has the keyword only while Cornered is active. While it isn't, an effect that refers to the keyword, such as an aura giving "units with Aggressive" +2 power, does not apply to it. For a triggered keyword, Cornered is still checked when the trigger event happens (`FR-KWD-002e`). Not yet in FRD 01. |
| **R19** | Does "attacks and kills" trigger when a replacement effect keeps the defender from being destroyed? | Eternal Watchman with Shifting Impossibility; later Dread Manifestation | **Ruled (user, 2026-10-28): no.** The defender was not killed (`SVY-R15`). Recorded in FRD 01 v3.15 as `SVY-R19` (`FR-CBT-002`). |
| **R20** | When a destruction replacement's source is lethal in the same death sweep as the deaths it applies to, or a byproduct leaves it lethal partway through, does it still replace them? (Open Question 13, T-0096) | Shifting Impossibility; Craterfall Ancient | **Ruled (user, 2026-10-29): yes, all of them.** "All three, then it dies." The source dies after the sweep's replaced deaths. One left lethal by a byproduct dies in the sweep's next pass. Replaced deaths go in a fixed order, oldest asset first, with no prompt. Revises R5. Recorded in FRD 01 v3.16 as `SVY-R20` (`FR-REP-001`, `FR-LAY-002`). |
| **R21** | Does a replacement see several kills from one kill step, or kills made while its source is already lethal, before the next sweep? (Open Question 13, T-0097) | Shifting Impossibility with Mutation Swell; Infested Apostle; the kill costs of Rhapsody's Embrace, Feed the Despoiler and Veil Render | **Ruled (user, 2026-10-29): yes.** A replacement sees every death at the same moment as its source's own, after Harvester of Souls. Deaths during one stack item and the sweep right after it are the same moment. *Mutation Swell*'s replaced kills still count toward its deploys (`SVY-R15`). Recorded in FRD 01 v3.16 as `SVY-R21` (`FR-REP-001`). |
| **R22** | When two players' destruction replacements apply in one sweep, whose replaced deaths go first? (Open Question 13, T-0096) | None in M2. Only Animus prints a destruction replacement. | **Ruled (user, 2026-10-29): the active player's first**, as for triggers (`FR-TRG-003`). Recorded in FRD 01 v3.16 as `SVY-R22` (`FR-REP-001`). |
| **R23** | When several replacements apply to one death, does the affected asset's controller choose which applies (a prompt, as `FR-REP-001` reads), or do they go in a fixed order like replaced deaths (`SVY-R20`)? If a prompt, what does a timeout pick? `FR-MOD-001` §6 has no fallback for it. (Open Question 14, T-0100) | Shifting Impossibility (2 copies in the Animus starter deck) | **Open.** Lead's recommendation: a prompt, timing out to §6's fixed order for untargeted selections applied to the replacements' sources, a source's own replacements in printed order. |
| **R24** | If a destruction replacement's steps leave the asset on the battlefield and still lethal (for example "exhaust it instead"), what happens? Does the same sweep take it up again, may the same replacement apply to it again, and if so, how does the loop end? Today the kernel takes each asset up once per sweep, so the asset stays in play while lethal until the next sweep, where the replacement applies again. (structural review 5, design note Open Question 15) | None. *Shifting Impossibility*, the only destruction replacement printed in any faction, always moves the asset to its owner's hand. | **Open.** Lead's recommendation: a replacement applies to a given asset at most once per death sweep, or per kill step. An asset still lethal afterwards is taken up again in the sweep's next pass, where that replacement no longer applies to it, so another replacement applies or it is destroyed. This ends every loop and keeps "the sweep repeats until nothing changes" (`SVY-R20`). |

**Noted but unreachable:** whether "this Headquarters is dealt damage" (Animoceps Rager) uses the same threshold as Unbowed. No starter card can give a Headquarters Armor or Resistant. Default: same as `FR-KWD-002a`.

---

## 5. Scope Questions

- **Q1 (answered: yes):** starter decks must be construction-legal under `FR-SET-001`, so `cardtown sim` will play legal Animus and Path decks rather than cycling each faction's executable cards.
- **Q2 (note):** the relaxed M2 Done-when only moves 3 requirements to M3 (`FR-ENT-102`, `FR-KWD-002n`, `FR-TRG-006`). That is still the honest boundary, but it means M2 is nearly the whole of FRD 01.

---

## 6. Card-by-Card Map

Hooks and primitives use FRD 02 names. "Rules" lists the distinguishing FRD 01 requirements beyond the wave's baseline.

### Animus

| Card | Wave | FRD 02 form | Hook → primitives | Rules | Notes |
| :--- | :---: | :--- | :--- | :--- | :--- |
| Despoiler Larva | A1 | Triggered | attack → `SCRATCH 1` | `FR-KWD-002m` | |
| Portent Scrivener | A1 | Triggered | death (on your turn) → `DRAW_CARD 1` | `FR-ZON-004` | `turnOwner: SELF` |
| Spoil Forest Creeper | A1 | Triggered | deploy → `SCRATCH 1` | | Simplest first card |
| Stormfront of Ideals | A1 | Triggered | death → `PUT_CARD_ON_DECK` bottom (self) | `FR-ZON-004` | Self-tracking fizzle |
| Unceasing Tide | A1 | Triggered | death → `RETURN_TO_HAND` (self from discard) | `FR-ZON-004` | Self-tracking fizzle |
| Fungal Evangelist | A1 | Triggered | deploy or death → `SCRATCH 2` | | Compound hook |
| Despoiler Worm | A1 | Triggered + Resistant 2 | death → `DRAW_CARD 2` (all opponents) | `FR-WIN-001` | Can deck the opponent |
| Titanic Spore Carrier | A1 | Triggered | death → `SCRATCH 1`, `DRAW_CARD 1` | | Composite sequence |
| Spoil Forest Colossus | A2 | Triggered | deploy or attack → `HEAL 2` chosen asset | `FR-KWD-002d` | |
| Craterfall Ancient | A2 | Triggered ×2 | attack (may) → `DEAL_DAMAGE` = own power to up to 2 other chosen assets; death → `DEAL_DAMAGE 4` all assets | `FR-TGT-002`, `FR-CBT-003` | Death trigger can destroy both HQs (Match Draw) |
| Animoceps Shambler | B | Triggered | exhaust → `HEAL 1` chosen allied asset in Main Theater | `FR-CAP-001` | Every Readied→Exhausted change; generation triggers are placed after it (`SVY-R17`, KIM-S03 rationale) |
| Drifting Concept | B | Triggered + Ethereal | deploy → `REVEAL_HAND` opponent, `DISCARD_CARD` chosen | `FR-MOD-001` §3 | |
| Infested Apostle ★ | B | Triggered | deploy (may) → `KILL` an allied unit, self allowed (untargeted, on resolution); if done → `DEAL_DAMAGE 2` chosen opposing unit | `FR-TRG-005`, `FR-TGT-001` (`SVY-R14`) | Benchmark #9, but its `optionalCosts` is superseded: the kill is on resolution (`SVY-R14`) |
| Animoceps Rager (HQ) | B | Triggered | HQ dealt damage → `SCRATCH 1`, once per turn | `FR-KWD-002a` (by analogy) | `maxPerTurn: 1` |
| Disorienting Pulse | C | Maneuver | `RETURN_TO_HAND` chosen opposing unit in Main Theater | `FR-MNV-001`, `FR-ZON-005` | |
| Feed the Despoiler | C | Maneuver + additional cost | kill allied unit; `DRAW_CARD 2` | `FR-TRN-003a` | Replaced kill cost stays paid (`SVY-R15`) |
| Funeral Fungus ★ | C | Maneuver + additional cost | exhaust readied allied unit; `RETURN_TO_HAND` chosen unit card from discard | `FR-TRN-003a` | Benchmark #8 |
| Conversion Spike | C | Maneuver | `DEAL_DAMAGE 1` chosen opposing unit; `HEAL 1` chosen allied unit or structure | `FR-TGT-003` | Two independent targets |
| Deafening Aria | C | Maneuver | `DEAL_DAMAGE 1` all opposing units in Main Theater | | |
| Ensnaring Mist | C | Maneuver | `MOVE_ZONE` chosen unit → its Reserves; `MODIFY_STAT` −1 power until end of turn to a second chosen unit | `FR-ZON-002`, `FR-LAY-001` | |
| Rhapsody's Embrace | C | Maneuver + additional cost | kill allied unit; `KILL` chosen opposing unit | `FR-TRN-003a` | Replaced kill cost stays paid (`SVY-R15`) |
| Despoiler's Emergence | C | Maneuver | `DEAL_DAMAGE 4` all opposing assets in Main Theater; `DEAL_DAMAGE 2` all opposing non-Scratch assets in Reserves | | |
| Rapturous Sermon | C | Maneuver | `MOVE_ZONE` all allied Scratch Company → Main Theater; `MODIFY_STAT` +2 until end of turn | `FR-LAY-001` snapshot locking | Named in FRD 01 |
| Animoceps Spores | C | Maneuver | `SCRATCH 2`, then `RETURN_TO_HAND` an allied Scratch Company (untargeted) | `FR-TGT-001` (`SVY-R3`) | Untargeted pick after the Scratch; any allied Scratch Company; timeout picks the first in fixed order (`SVY-R16`) |
| Rhapsody's Hollow | D | Activated | T3 + exhaust + kill self → `GAIN_CONTROL` chosen opposing unit in Main Theater | `FR-ACT-001`, `FR-ZON-002`, `FR-ZON-003` | |
| Sinuous Growth | D | Activated | T1 + kill self → `RETURN_TO_HAND` chosen Fungus or Plant card from discard | `FR-ACT-001` | R2 |
| Cliff of Portents | D | Activated | R1 + T1 + exhaust → `SEARCH_DECK` any → hand, shuffle | `FR-DEK-001`, `FR-DEK-002` | |
| Rift (HQ) | M3 | Activated | discard a card + exhaust → `SCRATCH 2`, `DEAL_DAMAGE 5` to self HQ | `FR-ACT-001`, `FR-CBT-003` | |
| Animoceps Stalk | E | Continuous | other allied assets in same zone +1 power | `FR-LAY-001` | `excludeSelf` |
| Animoceps Node | E | Continuous | allied units in same zone +1 power | `FR-LAY-001` | Structure; does not buff itself |
| Melodic Pummeler | E | Continuous | allied Horror units in same zone +1 power | `FR-LAY-001` | Includes itself |
| Hulking Incongruity | E | Continuous | allied Horror units in same zone gain Resistant 1 | `FR-LAY-001` | G2; includes itself |
| Veil Render ★ | F | Continuous (discard-active permission) | deploy from discard + kill allied asset | `FR-TRN-003`, `FR-TRN-003a`, `FR-CST-001` | Benchmark #6; replaced kill cost stays paid (`SVY-R15`) |
| Mutation Swell ★ | F | Maneuver | `KILL` 2 or 4 allied Scratch Company (untargeted, picked on resolution); `DEPLOY_FREE` per 2 killed | `FR-TGT-002`, `FR-TRN-003a` (`SVY-R3`) | Benchmark #14, but its `CHOSEN` targeting is superseded: the kill is untargeted (2026-10-20); timeout picks the first in fixed order (`SVY-R16`); replaced kills count toward its deploys (`SVY-R21`, `SVY-R15`) |
| Forced Mutation | F | Maneuver | `MORPH_SCRATCH` chosen opposing unit | `FR-ENT-103` | |
| Animoceps Spore Cloud ★ | G | Replacement | `WOULD_TAKE_DAMAGE` on allied units in same zone → `REDIRECT_DAMAGE_TARGET` self | `FR-REP-001`–`FR-REP-004` | Benchmark #3 |
| Shifting Impossibility | G | Replacement | `WOULD_BE_DESTROYED` on allied Scratch Company → `RETURN_TO_HAND` + `DEAL_DAMAGE 1` self | `FR-REP-001`–`FR-REP-003`, `FR-LAY-002` | R5 (revised by R20), R21; a replaced kill still counts as done (`SVY-R15`) |

### Path

| Card | Wave | FRD 02 form | Hook → primitives | Rules | Notes |
| :--- | :---: | :--- | :--- | :--- | :--- |
| Chainbreaker Sappers | A2 | Triggered | deploy → `DEAL_DAMAGE 1` chosen structure or HQ | `FR-TGT-001` | |
| Saboteur | A2 | Triggered | deploy → `DEAL_DAMAGE 2` chosen opposing non-Scratch unit; `DEAL_DAMAGE 1` chosen allied unit | `FR-TGT-003` | Named in FRD 01 |
| Devilhorn Guerrillas | A2 | Triggered | deploy into Main Theater → `DEAL_DAMAGE 2` chosen opposing unit | `FR-TRN-003` | `toZone` condition; two deploy zones |
| Ironclad Colossus | A2 | Triggered | deploy or attack → `ADD_ARMOR 2` chosen unit | `FR-LAY-002` | |
| Necropolis Hunt | A2 | Triggered | death → `DEAL_DAMAGE` = last-known power to chosen non-HQ asset | `FR-ZON-004` | |
| Hammerhand Brigands | B | Unbowed | → `SCRATCH 1` | `FR-KWD-002a` | Named in FRD 01 |
| Devilhorn Captain | B | Unbowed | → `DEAL_DAMAGE 1` chosen asset in Main Theater | `FR-KWD-002a` | Named in FRD 01 |
| Chainbreaker Irregulars | B | Unbowed | → `DRAW_CARD 1`, `DISCARD_CARD 1` | `FR-KWD-002a` | Discard can trigger Eerie Remnant; one card fast-paths, timeout discards the card just drawn (`SVY-R16`) |
| Allshan, Chainbreaker Revenant | B | Unbowed + Cornered Aggressive | Unbowed → `ADD_ARMOR 2` self; "Cornered: Aggressive 3." → the Aggressive keyword trigger (attack → `MODIFY_STAT` +3 power until end of turn), gated on Cornered | `FR-KWD-002a`, `FR-KWD-002e`, `FR-KWD-002f`, `FR-LAY-003` | text reworded by the user, 2026-10-26 (board #60); R18 |
| Hammerhand Bivouac | B | Triggered | turn start → distribute 2 armor among allied units in same zone | `FR-TGT-002`, `FR-TRN-001` | `DISTRIBUTED_POOL` |
| Blind Defile | B | Triggered (opposing) | opposing unit attacks (may, kill self) → `KILL` attacking unit | `FR-TRG-002`, `FR-TRG-005`, `FR-CBT-002` | Non-active player's trigger |
| Arkha, Sun's Fury | B | Triggered | attack (may) → `MOVE_CARD_ZONE` a unit card from discard to deck bottom (untargeted, on resolution); if done → `DEAL_DAMAGE 3` chosen opposing HQ | `FR-TRG-005`, `FR-TGT-001` (`SVY-R14`) | `SVY-R14`: not a cost |
| Rogue Fleshshaper | B | Triggered, modal | deploy → choose one: `HEAL 1` chosen asset, or `DEAL_DAMAGE 1` chosen non-Scratch asset | `FR-MOD-001` | `FR-ABL-010` |
| Eternal Watchman (HQ) | B | Triggered | `EVENT_COMBAT_KILL` as attacker (unit or structure) → `RETURN_TO_HAND` chosen unit card from discard | `FR-CBT-002` | |
| Chainbreaker's Tenacity | C | Maneuver | chosen allied unit: `MODIFY_STAT` +1 until end of turn, `ADD_ARMOR 2` | | |
| Fortify | C | Maneuver | `ADD_ARMOR 3` chosen allied unit | | |
| Tactical Withdrawal | C | Maneuver | `RETURN_TO_HAND` up to one chosen allied unit; `DRAW_CARD 1` | `FR-TGT-002` | `minCount: 0` |
| Blindside | C | Maneuver | `MODIFY_STAT` −3 until end of opponent's turn | `FR-LAY-003` | Full-round duration |
| Improvised Orders | C | Maneuver | `DRAW_CARD 3`, `DISCARD_CARD 1` | | |
| Devilhorn Sanction | C | Maneuver | `KILL` chosen opposing unit | | |
| Pincer Tactics | C | Maneuver | `DEAL_DAMAGE 1` up to 2 chosen allied units; `DEAL_DAMAGE 2` chosen non-Scratch asset | `FR-TGT-002` | Enables Unbowed |
| Poisoned Rations | C | Maneuver | `MODIFY_STAT` −3 max health until end of turn | `FR-LAY-002` | |
| Ambush Tactics | C | Maneuver, modal | choose two of: `KILL`, `RETURN_TO_HAND`, delegated `DEAL_DAMAGE` = power, `PUT_CARD_ON_DECK` top | `FR-MOD-001`, `FR-TGT-003` | R6; delegated source as on Soul Whip |
| Steel Rain | C | Maneuver + additional cost | discard up to 3; `DEAL_DAMAGE 2` chosen asset per discarded card; if hand empty, `DRAW_CARD 2` | `FR-TRN-003a`, `FR-TGT-001` (`SVY-R3`, `SVY-R4`) | Targets chosen with the discard; repeats allowed |
| Boneyard Scavenger | D | Activated | T1 + exhaust + kill self → `RETURN_TO_HAND` chosen Maneuver card from discard | `FR-ACT-001` | |
| Ruin Partisans | D | Activated | T2 + exhaust + kill self → `RETURN_TO_HAND` chosen unit card from discard | `FR-ACT-001` | R2 |
| Encampment (HQ) | M3 | Activated | discard a card + exhaust → `DEAL_DAMAGE 1` to each of up to 2 chosen non-Scratch units | `FR-ACT-001`, `FR-TGT-002` | |
| Brigand Highlander | E | Continuous (Cornered) | self +2 power while Cornered | `FR-KWD-002e`, `FR-LAY-003` | |
| Macabre Dancers | E | Unbowed + continuous (Cornered) | Unbowed → `HEAL 2` self; Cornered → Resistant 1 | `FR-KWD-002a`, `FR-KWD-002e` | Heal fizzles if lethal (FRD 01) |
| Ancestral Monument | E | Continuous | other allied assets in same zone gain "Cornered: Resistant 1" | `FR-LAY-001`, `FR-LAY-003` | `abilityGrants`; G2 |
| Boneyard Amalgam | E | Continuous (hand-active) | costs R1 less per unit card in your discard | `FR-CST-001` | `QUERY` count |
| Eerie Remnant | F | Triggered (discard-active) + Deploys Exhausted | hand → discard (may) → `DEPLOY_FREE` self | `FR-ZON-004`, `FR-TRN-003a` | G1 |
| Solarheart Monk | F | Triggered (discard-active) | enters discard → `HEAL 2` chosen asset | `FR-ZON-004` | G1; any route into discard |
| Broken Horn Armorer | F | Triggered | self moves zone → `ADD_ARMOR 1` chosen asset in same zone | `FR-ZON-002` | R1 |
| Chainbreaker Phalanx | F | Triggered | self moves to Main Theater (may) → `MOVE_ZONE` chosen allied unit from Reserves, `ADD_ARMOR 2` to it | `FR-ZON-002` | R1; can chain into another Phalanx |
| Devilhorn Scouting Party | F | Maneuver | `LOOK_TOP_CARDS 3` → `DEPLOY_FREE` up to 1 unit with cost ≤ R3, rest to bottom in any order; `SCRATCH 1` | `FR-MOD-001` §4, `FR-TRN-003a` | |
| Reaver's Rush | F | Maneuver | `DEPLOY_FREE` chosen unit card with power ≤ 3 from discard | `FR-TRN-003a` | |
| Allshan's Battle Trance ★ | F | Maneuver | `DEPLOY_FREE` from discard to Main Theater; `GRANT_KEYWORD` Stalwart; `MODIFY_STAT` +2 | `FR-TRN-003a` | Benchmark #11 |
