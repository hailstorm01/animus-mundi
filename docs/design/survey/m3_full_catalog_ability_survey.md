# M3 Full Catalog Ability Survey (Concordia, Soulless, Rift, Encampment)

- **Date:** 2026-11-24
- **Source:** `data/cards/*.csv`, `pkg/engine/ability` and `pkg/engine` at `4fe88866`; [FRD 01](../../frd/01_Game_Rules_FRD.md) v3.31; [FRD 02](../../frd/02_Card_Data_Pipeline_FRD.md) v2.0.
- **Purpose:** map the 67 cards on `cards.NotYetExecutable` onto the `ability` vocabulary, list the shapes it lacks and the rulings needed, and order the M3 build waves. This is Deliverable 4 of the [M3 design note](../../architecture/m3-full-catalog.md).

FRD 01 is the rules source. This document names requirement IDs and never restates the rules. The `ability` package wins any disagreement about a shape.

---

## 1. Summary

- **67 cards to model:** 33 Concordia, 32 Soulless, *Rift* and *Encampment*. 42 are starter-deck cards (23 Concordia and 19 Soulless, each count including its Headquarters). 25 are not.
- **38 need no new shape** (24 deck, 14 non-deck). This was measured, not estimated. On 2026-11-24 the lead wrote draft definitions for all 38 in a throwaway worktree:
  - Every one loaded as `Supported()`, and the other 29 did not.
  - 48 matrix runs of 40 games (starter and stress decks, seed 1) finished with no invariant or monitor failure.
  - The event log named all 24 deck cards in the starter runs, and all 38 in the stress runs.

  The drafts prove that the shapes fit. They are not tests of the cards' behavior, which each card task still writes.
- **29 need something new:** 18 deck cards and 11 non-deck cards. Three of the 11 (*Command Barge*, *Rift*, *Encampment*) need only the mandatory discard cost that *Enclave* brings.
- **FRD 01 reach:** the deck waves reach all four unpinned requirements: `FR-KWD-002n` (wave B), and `FR-KWD-002p`, `FR-TRG-006` and `FR-ENT-102` (wave E).
- **Rulings:** nine questions are open (§4, R57 to R65). Eight concern deck cards. Wave A can start now without *Will Breaker* (R59).
- **Matrix:** `KIM-C01` stays unreachable. None of the 67 cards prints Cornered.
- **Waiting directives no M3 card reaches:** #71 (a step on each player), #91 (2) (a cost on a mode), and the draw-hook half of #15 (2).
- **Pile round trip (#183):** re-checked for every wave, and none reaches it (§2).

---

## 2. Build Order (Waves)

Deck cards come first (design note Deliverables 5 and 6). Waves B to E depend only on A's card files, so they can be reordered. The playtest checkpoint comes after wave E. In every wave, shape tasks come before the cards that use them.

| Wave | New shapes (§3) | FRD 01 pinned | Cards | Count | Rulings first | Directives | #183 |
| :---: | :--- | :--- | :--- | :---: | :--- | :--- | :--- |
| **A** | None | | Growing Legion, Icewind Regent, Leyline Regrowth, Seismic Instability, Soul Whip, Stormbolt Ray, Swirling Dustdevil, Tempest City Ray, Hurricane Outrigger, Echoing Nightmare, Will Breaker, Bring It Down, Colossus Platform, Conscript Rifles, Culverin Brigade, Dolus Macrocannon, Dolus Refiner, Earthshaker Colossus, Fire for Effect, Hellfire Armorers, Inspired Salvage, MAX-8 Autonomous Warsuit, Mobile Armory, Visionary's Eye | 24 | R59 (Will Breaker only) | | Not reached: its two pile cards are Maneuvers |
| **B** | G4 readying, G5 readiness filter, G6 mandatory discard, G7 "killed this way", G8 a count of what a step did | `FR-KWD-002n` | Tempest Runner, Visionary's Insight, Armored Charge, Enclave, Frostfire Breath, All-Out War | 6 | R57 (Tempest Runner, Armored Charge, Enclave), R58 (All-Out War) | #120 | Not reached: nothing leaves a pile |
| **C** | G9 current Health, G10 computed filter bound, G11 combat-kill victim, G12 after-attack hook, G13 discard in full, G14 chosen move destination | | Cathedral of Self, Consuming Earth, Dread Manifestation, Piston Fist Prototype, Redlined Brawler, Sandstorm Lightbringer | 6 | R62 (Sandstorm Lightbringer), R63 (Redlined Brawler) | #182 | Not reached: nothing leaves a pile |
| **D** | G15 partitioned looks | | Hourglass Diviner, Samsara, Living Tempest, Light of the World Soul | 3 | R60 (Diviner, Samsara), R61 (Light) | | Not reached: Samsara only adds to a pile |
| **E** | G16 Overkill, G17 delayed triggers, G18 tucking | `FR-KWD-002p`, `FR-TRG-006`, `FR-ENT-102` | Tempest Surge, Rage's Wake, NO-AH-5 | 3 | R64 (NO-AH-5) | #142 | Not reached: see *Rage's Wake*, below |
| — | **Playtest checkpoint: Concordia vs Soulless in `cardtown play`** | | | **42** | | | |
| **F** | None (*Command Barge*, *Rift* and *Encampment* reuse G6) | | Awaken Nightmare, Flickering Ember, Forgemaster's Resilience, Reverberating Terror, Screaming Palisade, Stardust Colossus, Command Barge, Comms Relay, Dolus Infusion, Dominator Warsuit, Field Mechanic, Land Fortress, Piston Fist, Popup Bunker, Railway Reinforcements, Rift, Encampment | 17 | R65 (Awaken Nightmare) | | Not reached: *Stardust Colossus* takes a card out, and nothing in the waves puts it back above a waiting item |
| **G** | G10 again (strict bound), G19 generate, G20 deals-damage hook, G21 search with criteria, G22 targeted pool, G23 filtered revealed hand, G24 subtype grant, G25 granted trigger; G17 at end of turn | | Gemstone Titan, Natural Attunement, Tidal Leviathan, World Soul Messenger, Macrocannon Barrage, Mage Seeker's Inquisition, March of Progress, Redline Infusion | 8 | None | #106, #15 (2) end-of-turn half | Not reached: nothing leaves a pile |

**Total: 67** (24 + 6 + 6 + 3 + 3 deck, 17 + 8 non-deck).

**Rage's Wake and #183:** its delayed trigger takes its own card out of the pile. For a round trip, the card would have to die again above a second waiting item that tracks it, and no Concordia or Neutral unit's deploy trigger can damage its own unit.

**Deliverable 6a** (the Headquarters-kill check) keeps its place in the design note. Directive #181 is `open: before` its task.

---

## 3. Vocabulary Gaps (kernel decides; no ruling needed unless §4 says so)

IDs continue from the M2 survey's G1 to G3. A shape is built with its first card (`FR-ABL-013`), and its task updates FRD 02. The names below are placeholders.

| ID | Gap | Cards | Proposed shape |
| :--- | :--- | :--- | :--- |
| **G4** | No step readies an asset | Tempest Runner, Visionary's Insight, Armored Charge, Enclave, All-Out War; *Redline Infusion*'s "Ready this unit" | One `Op` that changes its Subject from Exhausted to Readied (`FR-KWD-002n`). It works with `SubjectTarget`, `SubjectAll` and `SubjectSelf` |
| **G5** | `Filter` has no readiness criterion (`FR-TGT-001` lists it) | Frostfire Breath, Visionary's Insight; the G4 cards if R57 says so | A readiness field on `Filter`, whose zero value means any |
| **G6** | `CostDiscard` is always "up to N, possibly none" (#120) | Enclave, Command Barge, Rift, Encampment | The mandatory form. `costsPayable` needs the cards in hand. `chooseCostDiscards` already chooses them before the targets |
| **G7** | `IfDone` can't ask whether a step's damage was lethal | Frostfire Breath | A step gate that reads lethality as `FR-LAY-002` defines it ("killed this way") |
| **G8** | No `Value` reads what an earlier step did, and no step discards a whole hand | All-Out War | A `ValueSource` reading the count in `r.done` of an earlier step, and an "all of them" discard. The Regroup clause then finds the new Scratch Companies itself |
| **G9** | No `Value` or `Stat` reads Current Health | Cathedral of Self; Gemstone Titan (G10) | `ValueSourceHealth`, with the source's last-known value as `SVY-R10` gives for power, and `StatHealth` beside `StatMaxHealth` |
| **G10** | `Filter.AtMost` is a constant | Consuming Earth (sum of Requisition yields), Gemstone Titan ("less than this unit's power") | A computed bound, read again at the second verification stage (`FR-TGT-003`). A strict comparison for Titan |
| **G11** | `HookCombatKill` doesn't say what was killed | Dread Manifestation ("another unit") | A victim `Filter` on the trigger. *Eternal Watchman* is unchanged, because a killed Headquarters ends the match |
| **G12** | No hook after an Attack resolves | Piston Fist Prototype | A post-attack hook, placed with the combat-kill triggers (`FR-CBT-002` Step 3) |
| **G13** | `IfDone` is true after one of two discards | Redlined Brawler | Waits for R63 |
| **G14** | `OpMoveZone.To` is fixed | Sandstorm Lightbringer | A choice of Reserves or the Main Theater. Waits for R62 |
| **G15** | `OpLookTop` can only deploy, or put cards on the bottom | Hourglass Diviner, Samsara, Light of the World Soul | Partition destinations: hand, discard pile, bottom. Samsara also needs a gate on what was discarded (*"if … was a maneuver card"*), with its target locked on placement (`FR-TGT-001`). **Technical risk:** today each whole outcome is one listed option. *Light of the World Soul* over seven unit cards has 25,200 outcomes before zone choices, which a menu can't show. A structured decision changes the Chooser contract ([ADR 0003](../../architecture/decisions/0003-synchronous-chooser-for-player-choices.md)), so wave D starts with a decision record |
| **G16** | Overkill | Tempest Surge, Redline Infusion | Its own hook, and a `Value` for the excess, fixed when it fires (`FR-KWD-002p`). Not a trigger on `HookCombatKill`. The settling tests are in M2 Open Question 12 |
| **G17** | Delayed triggered abilities | Rage's Wake (on a kill), Redline Infusion (end of turn) | A registry that tracks one asset, then its card instance (`FR-TRG-006`), and expires in 4c. The end-of-turn form also needs the 4a/4b trigger step of #15 (2) |
| **G18** | No Attached sub-zone | NO-AH-5 | `FR-ENT-102`: kill and tuck in one step, the host-departure purge, the morph purge (#142), the new zone in views, events and invariants, and a static power that counts tucked cards. Split across several tasks |
| **G19** | No step generates resources | Natural Attunement | An `Op` that adds Requisition or Tactics to the controller's pool |
| **G20** | No "deals damage" hook | Tidal Leviathan | A hook on the source, per `FR-REP-004` Stage 5 |
| **G21** | `OpSearchDeck` has no criteria and goes only to hand (#106) | World Soul Messenger | `Step.Filter`, and a free-deploy destination with its zone choice (`SVY-R43`) |
| **G22** | `Split` is untargeted armor only | Macrocannon Barrage | A pool over chosen targets, split on declaration, with evaporation (`SVY-R51`, `SVY-R52`, `FR-TGT-003`). It shares the amounts check |
| **G23** | `ControllerChooses` offers every card in the hand | Mage Seeker's Inquisition | A card-level `Filter` on the discard choice (`SVY-R50`) |
| **G24** | No subtype grant, and filters read printed subtypes | March of Progress | A Layer 2 modifier, with `matchesPrinted` reading effective subtypes for assets (`FR-LAY-001`) |
| **G25** | No granted triggered ability | Redline Infusion | A modifier that carries a trigger until end of turn, fired by the same hooks as printed ones |

**Pinned by one card in M2 (write-up §4.2).** The drafts use these again, so each task checks its card against that one test case: `Dealer` (*Soul Whip*), `UntilEndOfOpponentTurn` (*Comms Relay*), `CostReduction` (*Swirling Dustdevil*), `OpGrantKeyword` (*Fire for Effect*, *Forgemaster's Resilience*, *Comms Relay*) and `HookCombatKill` (*Dread Manifestation*).

---

## 4. Ruling Questions

IDs continue from the [M2 survey](m2_starter_set_ability_survey.md) §4, which ends at R56. `pm-frd-drafter` records each answer here from its Design Outcome.

| ID | Question | Cards | Status |
| :--- | :--- | :--- | :--- |
| **R57** | Must the target of "Regroup chosen X" be Exhausted? `FR-KWD-002n` says the player chooses Exhausted assets, but only *Visionary's Insight* prints "exhausted". If it must, *Armored Charge* can't be enacted for its heal and armor on a Readied unit. | Tempest Runner, Enclave, Armored Charge; Visionary's Insight | Open. Lead's recommendation: only the printed criteria count, and regrouping a Readied asset does nothing |
| **R58** | Does "gets power equal to the number of cards discarded this way" add that much power, or set power to it? | All-Out War | Open. Lead's recommendation: it adds. Reword to "gets +X power" |
| **R59** | What is the Requisition cost of a face-down Scratch Company on the battlefield? FRD 01 gives its stats and yields but no cost. The kernel reads 0, and no M2 card could tell. | Will Breaker | Open. Lead's recommendation: 0, so it is a legal target |
| **R60** | A look that puts one card in each of two places, with only one card left in the deck: where does it go? | Hourglass Diviner, Samsara, Living Tempest | Open. Lead's recommendation: destinations fill in printed order, so it goes to hand. Samsara then discards nothing and deals no damage |
| **R61** | In "deploy up to 2 unit cards from among them for free, and put up to 1 of those cards into your hand", does "those cards" mean the unit cards or any of the 7? | Light of the World Soul | Open. Lead's recommendation: any of the 7 not deployed. Reword to "put up to 1 of the other cards into your hand" |
| **R62** | When is "the Main Theater or Reserves" chosen, may it be the zone the unit is already in, and what does a timeout take? | Sandstorm Lightbringer | Open. Lead's recommendation: chosen on declaration with the target (`SVY-R2`). Either zone is legal. A timeout takes Reserves (as `SVY-R43`) |
| **R63** | "You may discard 2 cards. If you do, …" with fewer than 2 cards in hand: is X carried out in part, does it count as done, and is the "you may" asked? | Redlined Brawler | Open. Lead's recommendation: an X that can't be carried out in full isn't carried out, so nothing is discarded. With fewer than 2 cards as it would be placed, the "you may" is not asked (extending `SVY-R48`) |
| **R64** | Can NO-AH-5 choose itself for "Kill chosen allied non-Headquarters asset and tuck it face down under this unit"? | NO-AH-5 | Open. Lead's recommendation: no. Reword to "another chosen allied non-Headquarters asset" (`SVY-R34`) |
| **R65** | "Move chosen opposing unit … to its owner's Reserves": whose Reserves, when the opponent controls a unit you own (*Rhapsody's Hollow*)? `FR-ZON-002` uses the controller's Reserves, *Dread Manifestation* prints "its controller's", and *Tempest City Ray* prints only "to Reserves". | Awaken Nightmare; Tempest City Ray (default: its controller's) | Open. Lead's recommendation: always its controller's. Reword *Awaken Nightmare* |

**Noted but unreachable** (defaults, no ruling needed for M3):
- *Piston Fist Prototype*, when its Attack fizzles and it survives: no after-attack trigger. No card lets a defender leave in Step 2 while the attacker stays.
- *Gemstone Titan*, gone before its trigger resolves: its last-known power, as `SVY-R10`.
- *Hurricane Outrigger*, deployed on an opponent's turn ("until the end of your turn"): nothing deploys a Concordia unit then. It is modeled as until end of turn.
- *NO-AH-5*, gone before its ability resolves: the asset is killed and nothing is tucked. The killed card never enters a discard pile when it is tucked, so enter-discard abilities don't see it until the host leaves.

---

## 5. Scope Questions

- **Q1 (note):** #71, #91 (2) and the draw half of #15 (2) have no M3 card, so they stay `waiting` past M3 unless the user retires them.
- **Q2 (note):** nothing here needs a new package.

---

## 6. Card-by-Card Map

"Built" means the draft definition in §1 loaded and played. Hooks and ops use the `ability` names.

### Concordia: deck

| Card | Wave | Form | Definition | Rules | Notes |
| :--- | :---: | :--- | :--- | :--- | :--- |
| Growing Legion | A | Maneuver | `OpScratch` 1 | `FR-KWD-002m` | Built. Simplest first card |
| Icewind Regent | A | Activated | exhaust: `OpReturnToHand` chosen unit | `FR-ACT-001` | Built. May choose itself |
| Leyline Regrowth | A | Maneuver | up to 1 unit card from own discard → hand; `OpDraw` 1 | `FR-TGT-002` | Built |
| Seismic Instability | A | Maneuver | `OpDealDamage` 3 to all in the Main Theater, then 4 to all structures there | `FR-REP-004` | Built. Structures take two instances |
| Soul Whip | A | Maneuver | `OpDealDamage` with `Dealer` | `SVY-R27` | Built. Second use of `Dealer` |
| Stormbolt Ray | A | Triggered + Activated | deploy → 3 to chosen asset; T2, exhaust: return self | `FR-ACT-001` | Built |
| Swirling Dustdevil | A | `CostReduction` | R1 per Maneuver card in own discard | `FR-CST-001` | Built |
| Tempest City Ray | A | Activated | T2, exhaust: `OpMoveZone` chosen opposing unit → Reserves | `FR-ZON-002` | Built. See R65 |
| Hurricane Outrigger | A | Triggered | deploy → own Headquarters +1 power until end of turn | `FR-LAY-003` | Built |
| Echoing Nightmare | A | Triggered | deploy → return chosen opposing Main Theater unit; `IfDone` → move self | `SVY-R14`, `SVY-R1` | Built. The move fires move hooks |
| Will Breaker | A | Triggered | deploy → `OpKill` chosen opposing unit, `StatRequisition` 3 | `FR-TGT-001` | Built. R59 |
| Tempest Runner | B | Activated | exhaust: regroup another chosen allied unit in the same zone | `FR-KWD-002n`, `SVY-R34`, `SVY-R44` | G4. R57 |
| Enclave (HQ) | B | Activated | discard a card, exhaust: regroup up to 2 chosen allied units | `FR-KWD-002n` | G4, G6. R57 |
| Frostfire Breath | B | Maneuver | 2 to chosen exhausted non-Scratch unit; if killed this way, draw 1 | `FR-LAY-002` | G5, G7 |
| Cathedral of Self | C | Activated | T1, exhaust: damage = own Current Health to chosen unit | `SVY-R10` | G9 |
| Consuming Earth | C | Maneuver | `OpKill` chosen opposing unit, power ≤ sum of own structures' Requisition yields | `FR-TGT-003` | G10 |
| Dread Manifestation | C | Triggered + aura | combat kill of a unit (may) → chosen Main Theater unit → its controller's Reserves; opposing units in the same zone −2 power | `FR-CBT-002`, `SVY-R56` | G11. #182 gives its trade tests. `SVY-R19` too |
| Sandstorm Lightbringer | C | Activated | exhaust: `OpMoveZone` chosen allied unit → chosen zone | `FR-ZON-002` | G14. R62 |
| Hourglass Diviner | D | Triggered | deploy → look 2: 1 hand, 1 bottom | `SVY-R41` | G15. R60 |
| Samsara, Living Tempest | D | Triggered | attack → look 2: 1 hand, 1 discard; if a Maneuver card, 2 to the chosen asset | `FR-TGT-001`, `FR-ZON-004` | G15. R60. Target locked on placement. Fires enter-discard |
| Light of the World Soul | D | Maneuver | look 7: up to 2 units deployed free, up to 1 to hand, rest bottom | `FR-MOD-001` §4, `SVY-R41`–`SVY-R43` | G15. R61. The structured decision |
| Tempest Surge | E | Triggered + Activated | Overkill: excess to chosen opposing Main Theater unit; T2: self +2 power | `FR-KWD-002p` | G16 |
| Rage's Wake | E | Maneuver | chosen unit +2 power; delayed on its kill: deploy it from own discard → Reserves | `FR-TRG-006`, `FR-ZON-005` | G17. An opposing unit's card is in the other pile, so nothing is deployed |

### Concordia: non-deck

| Card | Wave | Form | Definition | Rules | Notes |
| :--- | :---: | :--- | :--- | :--- | :--- |
| Awaken Nightmare | F | Maneuver | `OpMoveZone` chosen opposing Main Theater unit; all opposing units −1 power | `FR-LAY-001` | Built. R65 |
| Flickering Ember | F | Activated | T4: allied assets in the same zone +1 power | `SVY-R44` | Built. Repeatable; includes itself |
| Forgemaster's Resilience | F | Maneuver | `OpGrantKeyword` Stalwart | `FR-KWD-002c` | Built |
| Reverberating Terror | F | Maneuver | up to 2 chosen units and structures −1 power | `FR-TGT-002` | Built |
| Screaming Palisade | F | Activated | T3, exhaust: `OpDiscard` 1, opponents | `SVY-R16` | Built |
| Stardust Colossus | F | Triggered ×2 | deploy or attack (may) → chosen card from own discard → hand | `SVY-R47`, `SVY-R48` | Built |
| Gemstone Titan | G | Triggered | deploy → `OpKill` chosen opposing structure, Current Health < own power | `FR-TGT-001` | G9, G10 |
| Natural Attunement | G | Maneuver | generate 2 Requisition; draw 1 | `FR-RES-001` | G19 |
| Tidal Leviathan (HQ) | G | Triggered | deals damage → chosen allied unit or structure heals 1 | `FR-REP-004` | G20. Attack and retaliation each fire it |
| World Soul Messenger | G | Triggered | deploy (may) → search for an Elemental card, deploy it free, shuffle | `FR-DEK-001`, `SVY-R37`, `SVY-R38` | G21, #106 |

### Soulless: deck

| Card | Wave | Form | Definition | Rules | Notes |
| :--- | :---: | :--- | :--- | :--- | :--- |
| Bring It Down | A | Maneuver | 2 to chosen non-Scratch asset | `FR-TGT-001` | Built |
| Colossus Platform (HQ) | A | Triggered | attack → chosen allied unit gains 1 armor | `FR-CBT-002` | Built |
| Conscript Rifles | A | Triggered | deploy → 1 to chosen non-Scratch unit | `FR-TGT-003` | Built. May choose itself |
| Culverin Brigade | A | Triggered | deploy → 1 to each of up to 2 chosen opposing assets | `SVY-R11` | Built |
| Dolus Macrocannon | A | Activated | R1, exhaust: 2 to chosen non-Scratch asset | `SVY-R36` | Built |
| Dolus Refiner | A | Activated | exhaust: draw 2, discard 2 | `SVY-R16` | Built |
| Earthshaker Colossus | A | Triggered ×2 | deploy or attack → 2 to chosen non-Scratch asset | | Built |
| Fire for Effect | A | Maneuver | opposing Main Theater assets gain Resistant −1; 1 to chosen Main Theater asset | `FR-LAY-001` | Built. Named in FRD 01 |
| Hellfire Armorers | A | Ward + aura | allied units in the same zone have Aggressive 1 | `FR-KWD-002f` | Built. Includes itself |
| Inspired Salvage | A | Maneuver | up to 2 chosen Mechanized cards from own discard → hand | `SVY-R11` | Built |
| MAX-8 Autonomous Warsuit | A | Triggered | deploy → 2 to chosen asset; chosen asset heals 2 | `SVY-R35` | Built. Two slots |
| Mobile Armory | A | Activated | exhaust: chosen allied unit in the same zone +1 power | `SVY-R44` | Built. May choose itself |
| Visionary's Eye | A | Triggered | attack → draw 1 | | Built |
| Visionary's Insight | B | Maneuver | regroup up to 1 chosen exhausted allied asset | `FR-KWD-002n` | G4, G5 |
| Armored Charge | B | Maneuver | regroup chosen allied Mechanized unit; it heals 5 and gains 2 armor | `FR-KWD-002n` | G4. R57 |
| All-Out War | B | Maneuver | discard hand; chosen allied asset +X power; Scratch X; regroup all Scratch Companies in own Reserves | `FR-TGT-003`, `FR-KWD-002n` | G4, G8. R58. Only the targeted clause can fizzle |
| Piston Fist Prototype | C | Aggressive 9 + triggered | after attack → move self to own Reserves | `FR-CBT-002` | G12 |
| Redlined Brawler | C | Triggered | deploy (may) → discard 2; if done, move self to the Main Theater | `SVY-R14` | G13. R63 |
| NO-AH-5 | E | Ward + Activated + static | R2: kill chosen allied non-Headquarters asset and tuck it; +2 power per tucked card | `FR-ENT-102` | G18. R64. #142. Its filter excludes a Headquarters (`SVY-R55`) |

### Soulless: non-deck

| Card | Wave | Form | Definition | Rules | Notes |
| :--- | :---: | :--- | :--- | :--- | :--- |
| Command Barge (HQ) | F | Activated | discard a card, exhaust: draw 2, discard 1 | `FR-ACT-001` | G6 |
| Comms Relay | F | Activated | exhaust: chosen asset gains Ward until the end of the opponent's turn | `FR-LAY-003` | Built |
| Dolus Infusion | F | Maneuver | draw 2; own Headquarters heals 2 | `FR-KWD-002d` | Built |
| Dominator Warsuit | F | Activated | R1: self gains 1 armor | `FR-ACT-002` | Built. Repeatable |
| Field Mechanic | F | Triggered | deploy → chosen allied Mechanized unit gains 2 armor, and +2 power until end of turn | `SVY-R39` | Built |
| Land Fortress | F | Activated | exhaust: chosen unit in the same zone +2 power, and gains 1 armor | `SVY-R44` | Built |
| Piston Fist | F | Maneuver | 1 to chosen unit; chosen allied unit +1 power | `SVY-R35` | Built |
| Popup Bunker | F | Triggered | deploy → allied units and structures in the same zone gain 1 armor | `SVY-R44` | Built. Includes itself |
| Railway Reinforcements | F | Maneuver | up to 3 chosen allied Reserves units → Main Theater | `FR-ZON-002` | Built |
| Macrocannon Barrage | G | Maneuver | 5 damage split among up to 5 chosen units | `SVY-R51`, `SVY-R52` | G22 |
| Mage Seeker's Inquisition | G | Maneuver | opponent reveals hand; choose a unit card; discard it | `SVY-R50` | G23 |
| March of Progress | G | Maneuver | up to 2 chosen units gain Mechanized; allied Mechanized units gain 2 armor and +2 power | `FR-LAY-001` | G24. Snapshot after the grant |
| Redline Infusion | G | Maneuver | chosen allied unit +2 power and a granted Overkill; delayed at end of turn: kill it | `FR-KWD-002p`, `FR-TRG-006` | G16, G17, G25, #15 (2) |

### Headquarters from Animus and Path

| Card | Wave | Form | Definition | Rules | Notes |
| :--- | :---: | :--- | :--- | :--- | :--- |
| Rift | F | Activated | discard a card, exhaust: Scratch 2, 5 to self | `FR-CBT-003` | G6 |
| Encampment | F | Activated | discard a card, exhaust: 1 to each of up to 2 chosen non-Scratch units | `SVY-R11` | G6 |
