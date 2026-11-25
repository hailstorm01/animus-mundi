# M2 Ability Shapes: the `ability` Vocabulary and What Changed From FRD 02

- **Milestone:** M2 → M3 (M2 Deliverable 7, carried over) | **Date:** 2026-11-22 | **Author:** `lead-architect`
- **Source:** `go doc -all ./pkg/engine/ability` and `pkg/engine/cards` at `09c7b523`; frozen [FRD 02](../../frd/reference/02_Card_Data_Pipeline_FRD_v1.3_frozen.md) v1.3 Domains 3–4; the [M2 survey](m2_starter_set_ability_survey.md) §3; the [M2 design note](../../architecture/m2-ability-model.md).
- **Purpose:** the input `pm-frd-drafter` rewrites FRD 02/03 from at M3 start (`PRODUCT_SUMMARY.md` §4, M3), and the vocabulary the M3 ability survey maps the remaining cards onto.

The package is the authority. This file indexes it and doesn't copy its field comments: for a field's exact meaning, read `go doc -all ./pkg/engine/ability` (or `ability.go` and `keywords.go`, 623 lines together). Rules behavior is cited by `FR-*` and `SVY-R*` IDs from [FRD 01](../../frd/01_Game_Rules_FRD.md), never restated.

---

## 1. The Model in Brief

- **One `ability.Def` per card**, a plain Go value in `cards.Defs`, keyed by card name (`animus.go`, `path.go`, `headquarters.go`; 73 definitions). The kernel interprets it ([ADR 0004](../../architecture/decisions/0004-abilities-as-data-in-go.md)). There's no JSON and no blueprint ID yet: the M3 catalog export writes these values out.
- **Standalone keywords aren't abilities.** `engine.LoadCards` parses each standalone keyword line in the CSV's rules text with `ability.ParseKeywordLine` into an `ability.Keywords` set ([ADR 0008](../../architecture/decisions/0008-keywords-live-in-ability.md)). A card with only keyword lines needs no `Def`.
- **`Def.Text` ties a definition to the card text.** It lists the exact non-keyword lines the `Def` implements. `LoadCards` removes each one from `Card.Unsupported`. If a `Text` line is missing from the CSV, that's a load error, so a reworded card stops running instead of running stale logic. A card with any line left in `Unsupported` can't be deployed or enacted. `cards.NotYetExecutable` lists those, and it's empty at the end of M2.
- **No field is set where nothing reads it.** `engine.checkShapeReads` (`cardload.go`) rejects these at load time: `Trigger.Watches` off `HookAttacks`, `Step.Split` anywhere but an `OpAddArmor` `SubjectPick` over assets with `UpTo` 0, `Step.ControllerChooses` off `OpDiscard`, and an aura's `Static.Filter.Stat` (#162, T-0278).
- **New vocabulary arrives only through a `shape` task.** A `card` task only combines values that already exist (design note, Boundaries).

---

## 2. Index of `go doc ./pkg/engine/ability`

### 2.1 Structs

| Type | What it is | Fields |
| :--- | :--- | :--- |
| `Def` | Everything one card prints besides standalone keywords | `Text`, `Triggers`, `Replacements`, `Effect` (a Maneuver's), `ExtraCosts` (declaration costs, `FR-TRN-003a`), `Activated`, `Cornered`, `CostReduction`, `Statics` |
| `Trigger` | One printed triggered ability | `Hook`, `Turn`, `ToZone`, `ActiveZone` (G1), `Cornered`, `Optional`, `FromKeyword`, `Watches`, `Watch`, `Effect`, `MaxPerTurn` |
| `Activated` | One printed "[Cost]: [Effect]." ability (`FR-ACT-001`, `FR-ACT-002`) | `Costs`, `Effect` |
| `Static` | One printed static ability, by `Kind` | `Kind`, `ActiveZone`, `Costs` (a deploy permission's), `Filter`, `Power`, `Keywords`, `Cornered` (an aura's) |
| `Replacement` | One printed replacement effect (`FR-REP-001`–`FR-REP-004`) | `Event`, `Filter`, `Steps` (run inline, never on the stack) |
| `Effect` | What a resolving ability does | `Targets`, `Steps`, or instead `Modes` with `ChooseN` (`FR-MOD-001` §5) |
| `Target` | One "chosen" clause, locked on placement (`SVY-R3`) | `Filter`, `UpTo` (`SVY-R11`), `PerDiscard` (`SVY-R4`), `PerKilled` with `Different` (`SVY-R53`, `SVY-R54`), `Another` (`SVY-R34`) |
| `Step` | One primitive applied to a subject | `Op`, `Amount`, `Subject`, `Target`, `Filter`, `UpTo`, `Dealer` (`SVY-R27`), `To`, `Keywords`, `Stat`, `Until`, `Bottom`, `IfDone` (`SVY-R14`), `IfHandEmpty`, `ControllerChooses` (`SVY-R50`), `Split` (`SVY-R51`) |
| `Cost` | One cost, paid on declaration | `Kind`, `Filter`, `UpTo`, `Counts` (`SVY-R53`), `Req`, `Tac` |
| `Filter` | A conjunction of criteria (`FR-TGT-001`) | `Allegiance`, `Kinds`, `Subtypes`, `Scratch`, `Other`, `Zone`, `ZoneRef`, `Stat`, `AtMost` |
| `Value` | A number a step reads when it resolves | `N`, `From` |
| `CostReduction` | The card's own hand-active Requisition reduction (`FR-CST-001`) | `Req`, `Per` (a pile filter) |
| `WhileCornered` | What a card has, or an aura grants, only while Cornered (`FR-KWD-002e`, `SVY-R18`) | `Keywords`, `Power` |
| `Keywords` | The standalone keyword set (`FR-KWD-002b`, `FR-KWD-002c` and `FR-KWD-002f` to `FR-KWD-002l`) | `Resistant`, `Stalwart`, `Aggressive`, `Ethereal`, `Ward`, `Versatile`, `Guardian`, `Armor`, `DeploysExhausted` |

### 2.2 Enums

The package doc's three rules decide where an enum starts: bit flags use `1 << iota`; an enum with no default starts at `iota + 1`, so an unset field is invalid; an enum whose zero value is a meaningful default starts at `iota`. The **Uses** column counts references in `pkg/engine/cards` at `09c7b523`. A zero-value member is usually left unwritten, so its count reads low. A count of 1 means a single card's tests pin the shape (§4.2).

| Enum | Start | Members (uses) |
| :--- | :--- | :--- |
| `Hook` | `iota + 1` | `HookDeploy` (10), `HookKilled` (8), `HookAttacks` (6), `HookDamaged` (6), `HookExhausted` (1), `HookCombatKill` (1), `HookMoves` (2), `HookDiscardedFromHand` (1), `HookTurnStart` (1), `HookEntersDiscard` (1) |
| `Op` | `iota + 1` | `OpScratch` (9), `OpDraw` (8), `OpReturnToHand` (11), `OpPutOnDeck` (3), `OpDealDamage` (20), `OpHeal` (7), `OpAddArmor` (8), `OpDiscard` (3), `OpModifyStat` (6), `OpKill` (6), `OpMoveZone` (3), `OpSearchDeck` (1), `OpDeployFree` (4), `OpMorphScratch` (1), `OpGrantKeyword` (1), `OpLookTop` (1), `OpGainControl` (1) |
| `Subject` | `iota` | `SubjectController` (zero), `SubjectSelf` (7), `SubjectOpponents` (2), `SubjectTarget` (52), `SubjectAll` (6), `SubjectPick` (4), `SubjectReplaced` (1), `SubjectWatched` (1) |
| `CostKind` | `iota + 1` | `CostKill` (4), `CostExhaust` (1), `CostDiscard` (1), `CostResources` (5), `CostExhaustSelf` (4), `CostKillSelf` (4) |
| `StaticKind` | `iota + 1` | `StaticDeployPermission` (1), `StaticAura` (5) |
| `Would` | `iota + 1` | `WouldBeDestroyed` (1), `WouldBeDealtDamage` (1) |
| `TurnOwner` | `iota` | `TurnAny` (zero), `TurnYours` (2) |
| `KeywordTrigger` | `iota` | `KeywordTriggerNone` (zero), `KeywordTriggerAggressive` (set by `Aggressive`) |
| `Duration` | `iota` | `UntilEndOfTurn` (zero), `UntilEndOfOpponentTurn` (1) |
| `Zone` | `iota` | `ZoneAny` (zero: any battlefield zone), `ZoneReserves` (3), `ZoneMainTheater` (11), `ZoneDiscard` (13) |
| `ZoneRef` | `iota` | `ZoneRefNone` (zero), `ZoneRefThat` (1, `SVY-R45`), `ZoneRefSame` (7, `SVY-R44`) |
| `Stat` | `iota` | `StatNone` (zero; Combat Power on a modify step), `StatPower` (1), `StatRequisition` (1), `StatMaxHealth` (1) |
| `Allegiance` | `iota` | `AllegianceAny` (zero), `AllegianceAllied` (35), `AllegianceOpposing` (15) |
| `Kinds` | `1 << iota` | `KindUnit` (54), `KindStructure` (3), `KindHeadquarters` (2), `KindManeuver` (1, pile filters only) |
| `ScratchCompany` | `iota` | `ScratchAny` (zero), `ScratchOnly` (5), `ScratchExcluded` (4) |
| `ValueSource` | `iota` | `ValueSourceFixed` (zero), `ValueSourcePower` (3, `SVY-R10`, `SVY-R27`) |

### 2.3 Functions and Readers

Every caller uses these, so a question about a shape is answered in one place.

| Name | Answers |
| :--- | :--- |
| `ParseKeywordLine(line, *Keywords) bool` | Is this rules-text line a standalone keyword? If so, add it to the set |
| `Keywords.Add`, `Keywords.Lines` | The one fold of two keyword sets (flags OR, numbers add), and the printed form |
| `Aggressive(n) Trigger` | The trigger the keyword "Aggressive N" means, with `FromKeyword` set (`FR-KWD-002f`) |
| `KeywordTrigger.String`, `Duration.String` | Printed name and log text |
| `Effect.Parts` | An effect and each of its modes' parts, depth first in printed order: the one walk into modes (#173 (c)) |
| `Effect.ReadsZoneRef` | Does any filter in the effect read this `ZoneRef`? |
| `Filter.InPile` | Does this filter choose cards in a discard pile rather than battlefield assets? |
| `Cost.Counted` | Is this a kill cost that offers a choice of counts? |
| `Target.AnotherSlot`, `Step.DealerSlot` | Which earlier target slot this one must differ from, and which slot deals the damage |
| `WhileCornered.Add` | The one fold of two Cornered grants |

### 2.4 How FRD 02's Keyword Table Maps

FRD 02 `FR-ABL-012` lists 14 keywords. In the kernel they take four forms:

| Form | Keywords |
| :--- | :--- |
| A field of `Keywords`, parsed from a standalone line | Resistant N, Stalwart, Ethereal, Ward, Versatile, Guardian, Armor N, Deploys Exhausted |
| A field of `Keywords` that is also a trigger | Aggressive N: the field, plus `Aggressive(n)` with `FromKeyword` |
| A printed preamble on a `Trigger` or `Def` | Unbowed: `Trigger{Hook: HookDamaged}`. Cornered: `Trigger.Cornered`, `Def.Cornered` and `Static.Cornered` |
| A step | Heal N: `OpHeal`. Scratch N: `OpScratch` |
| Not built (M3) | Regroup N (`FR-KWD-002n`), Overkill (`FR-KWD-002p`, new in FRD 01 v3.14) |

Berserk N (`FR-KWD-002o`) is retired (`SVY-R12`, board directive #68).

---

## 3. What Changed From FRD 02

FRD 02 is frozen design intent and was never built as written. The kernel kept its names where they fit (the survey and design note borrowed them for that reason) and changed them where FRD 01 or a ruling said otherwise. FRD 01 wins every conflict.

### 3.1 The Survey's Gaps (G1–G3)

Survey §3 found these gaps at M2 start. The design note's Open Question 2 decided all three in the kernel, with no ruling needed.

| ID | FRD 02 | `ability` today |
| :--- | :--- | :--- |
| **G1** | Only continuous auras have an `activeZone` | `Trigger.ActiveZone`: `ZoneAny` (the battlefield) or `ZoneDiscard`, with the discard hooks `HookEntersDiscard` and `HookDiscardedFromHand` (`FR-ZON-004`). *Eerie Remnant*, *Solarheart Monk*. No hand-active trigger exists |
| **G2** | Aura keyword grants sit in Layer 4 (`FR-ABL-004`) | Layer 3, per `FR-LAY-001`: `Static.Keywords`, `Static.Cornered.Keywords` and `OpGrantKeyword` all fold through `Keywords.Add` into `g.Keywords`. Only power is Layer 4 |
| **G3** | No self-sacrifice cost type (`KILL_ASSET` with selector `SELF`) | `CostKillSelf`, paid through the kill path (`FR-ACT-001`, `SVY-R15`) |

### 3.2 Later Changes, by FRD 02 Construct

| # | FRD 02 construct | `ability` today | Why |
| :---: | :--- | :--- | :--- |
| 1 | `AbilityDef` JSON, polymorphic on `abilityType` (`FR-ABL-001`) | One `Def` struct per card, with a typed list for each kind | ADR 0004. A card prints several kinds at once (*Macabre Dancers*: a trigger and a Cornered grant) |
| 2 | `KEYWORD_GRANT` as an ability type (`FR-ABL-012`) | Not an ability: `Keywords`, parsed from lines (§2.4) | ADR 0008, #159 |
| 3 | `triggerHook` may be an array ("deploys or attacks") | One `Hook` per `Trigger`. A compound preamble is two triggers with the same `Effect` (*Fungal Evangelist*) | Each hook fires, gates and counts on its own |
| 4 | About 30 `EVENT_*` hooks (`FR-TRG-001` in FRD 02) | 10 `Hook`s, all built for a starter card. `HookCombatKill` counts the attacker's kills only, and a replaced death isn't a kill (`SVY-R19`). `HookExhausted` is any Readied → Exhausted change (`SVY-R17`). `HookMoves` excludes entering the battlefield (`SVY-R1`). An opposing unit's attack is `HookAttacks` with `Watches` and `SubjectWatched` (`FR-CBT-002`, T-0241) | Built only what cards read. No `is_attacker` flag is needed |
| 5 | `ConditionDef` (turn owner, Cornered, from/to zone, value comparison, hand empty, filters) | Flat gates: `Trigger.Turn` (`TurnYours` only), `ToZone`, `Cornered`, `MaxPerTurn`. On steps: `IfDone` and `IfHandEmpty` | No starter card needs an opponent's-turn gate, a from-zone gate or a count comparison |
| 6 | `optionalCosts`: "you may [Cost]. If you do, [Effect]" | Not a cost. `Optional` is asked on placement (`SVY-R9`). X is a step on resolution, and Y has `IfDone` (`SVY-R14`, `SVY-R15`, `FR-TRG-005`). An optional trigger that could do nothing isn't asked (`SVY-R48`). "Up to one chosen" never makes a trigger optional (`SVY-R47`) | #47 |
| 7 | `COMPOSITE_SEQUENCE`, `CONDITIONAL_CLAUSE`, `ifTrue` / `ifFalse` | A flat `Steps` list with `IfDone` and `IfHandEmpty`. No else branch | No starter card needs one |
| 8 | `TargetRef.selector` (`SELF`, `CHOSEN`, `ALL`, `DISTRIBUTED_POOL`, `RANDOM`, `BOUND_TARGET`, `TRIGGER_EVENT_ENTITY`) | Split in two. **`Target`** is a "chosen" clause only, always locked on placement (`SVY-R3`). **`Subject`** says what a step acts on: `SubjectSelf`, `SubjectAll` (not a target, `FR-TGT-002`), `SubjectPick` (an untargeted pick on resolution, so Ward doesn't apply), `SubjectController` / `SubjectOpponents`, `SubjectTarget`, `SubjectReplaced`, `SubjectWatched` | Ward and `FR-TGT-003` apply only to "chosen". `RANDOM` isn't built: no fallback is random (`SVY-R16`) |
| 9 | `bindingPolicy: UPFRONT_LOCK / ON_RESOLUTION` | Gone. Every target locks on placement | `SVY-R3` (revised 2026-10-20) |
| 10 | `count`, `minCount`, `maxCount`, `allowedCounts` | `Target.UpTo` (distinct, `SVY-R11`). A choice of counts is a kill *cost's* `Counts` (`SVY-R53`) | *Mutation Swell* became an additional cost |
| 11 | `repeatCount` with a `CAUSALITY_CONTEXT` count (`DISCARDED_COUNT`, `KILLED_COUNT`, `divisor`) | `Target.PerDiscard` (`SVY-R4`, `SVY-R32`), and `Target.PerKilled` = k with `Different` (`SVY-R53`, `SVY-R54`) | The repetition is a target slot chosen on placement, not a value read later |
| 12 | `DISTRIBUTED_POOL` with `poolValue` (`FR-ABL-011`) | Only the untargeted pool is built: `Step.Split` on an `OpAddArmor` `SubjectPick`, split as it resolves (`SVY-R51`, `SVY-R52`). A targeted pool isn't built | *Hammerhand Bivouac*. *Macrocannon Barrage* (M3) needs the targeted form |
| 13 | `FilterDef.entityType` / `cardType` / `excludeHQ` / `excludeUnits` | One `Kinds` bit set | "Structure or Headquarters" is two bits |
| 14 | `excludeScratch` | The three-way `ScratchCompany`: any, only, excluded | "Allied Scratch Company unit" needs *only* |
| 15 | `negateSubtypes`, `readiness` | Not in `Filter`. Readiness is built into `CostExhaust` ("Readied allied asset") | No starter filter needs them |
| 16 | `zone` (`SAME_ZONE`, `RESERVES`, `MAIN_THEATER`, `HEADQUARTERS`, `BATTLEFIELD`, `ATTACHED`, `HAND`, `DECK`, `DISCARD`) | `Zone`: any battlefield zone, Reserves, Main Theater, or a discard pile (`InPile`). Source-relative zones are `ZoneRef`: `ZoneRefSame` is locked on placement for a stack item and is the current zone for an aura (`SVY-R44`); `ZoneRefThat` is a move's destination (`SVY-R45`) | "Same zone" needed a lock and a second meaning ("that zone") |
| 17 | `statComparison` (5 properties, 6 operators, a `ValueToken` bound) | `Stat` + `AtMost`: "N or less" with a fixed N, for power, Requisition cost or Maximum Health. A pile card is read printed, an asset effective | Only "N or less" is printed in M2 |
| 18 | `ValueToken` (`STATIC`, `DYNAMIC`, `QUERY`, `VARIABLE_X`, aggregates, `multiplier`, `divisor`) | `Value{N, From}`: fixed, or `ValueSourcePower` (the source's last-known power, `SVY-R10`; a chosen dealer's power while it's on the battlefield, `SVY-R27`). No count source | The one M2 count, *Boneyard Amalgam*'s, is `Def.CostReduction`. The first count `Value` reads `countPileCards` (#138) |
| 19 | `CostDef` (`RESOURCE`, `EXHAUST_SELF`, `EXHAUST_ASSET`, `KILL_ASSET`, `DISCARD_CARD`, `MOVE_CARD_ZONE`) | `CostResources` (Requisition and Tactics in one), `CostExhaustSelf`, `CostExhaust`, `CostKill`, `CostDiscard` (up to N, possibly none), and the new `CostKillSelf` (G3). `MOVE_CARD_ZONE` isn't built. Paying a cost isn't targeting, so Ward doesn't apply | A mandatory discard cost waits on #120 |
| 20 | `ActivatedAbilityDef.repeatable`, and `maxPerTurn` on activated abilities (`FR-ABL-008`) | Neither exists. An activated ability needs its targets to exist, as a Maneuver does (`SVY-R36`). `MaxPerTurn` is only on `Trigger` | No starter activated ability prints a limit |
| 21 | `ContinuousAuraDef` (`activeZone`, `scope`, `modifiers`, `abilityGrants`, `staticPermission`) | Split by job. `Static{Kind: StaticAura}`: `Filter` (with `Other` and `ZoneRefSame` for scope), `Power` (Layer 4), `Keywords` (Layer 3, G2), `Cornered`. `Static{Kind: StaticDeployPermission}`: `ActiveZone: ZoneDiscard` and `Costs`. Hand-active cost reduction is `Def.CostReduction`, not a `Static`. A card's own Cornered grant is `Def.Cornered` | Each job has one reader: `g.auras`, `deployPermission`, `Game.EffectiveCost`, `g.whileCornered` |
| 22 | Aura `modifiers` on `MAX_HEALTH`, `TACTICS_COST`, `SUBTYPE_GRANT`, with `SET` / `SUBTRACT` | Not built | No starter aura needs them |
| 23 | `abilityGrants` (any `AbilityDef`) | Only a Cornered grant (`Static.Cornered`, a `WhileCornered`). No granted trigger or activated ability | *Ancestral Monument*. *Redline Infusion* (M3) needs a granted trigger |
| 24 | `ReplacementWatcherDef.interceptHook` (`WOULD_TAKE_DAMAGE`, `WOULD_BE_DESTROYED`, `WOULD_DEPLOY`) + `replacementAction` | `Would`: `WouldBeDestroyed` with `Steps`, and `WouldBeDealtDamage` with no `Steps`, where the damage is always dealt to the source (`FR-REP-004` Stage 1). The `REDIRECT_DAMAGE_TARGET` primitive is folded in. `WOULD_DEPLOY` isn't built: Deploys Exhausted is a keyword | The sweep order and choice rulings are `SVY-R20`–`SVY-R24` |
| 25 | `ManeuverEffectDef` (`targeting`, `effect`, `modalCount`, `options`) and the root `additionalDeploymentCosts` | `Def.Effect`, with `Modes` and `ChooseN` for modal cards: chosen on placement before targets, from the legal modes, resolved in printed order (`SVY-R6`, `SVY-R28`–`SVY-R31`). Declaration costs are `Def.ExtraCosts` (`FR-TRN-003a`) | A mode with a cost waits on #91 (2) (`SVY-R33`) |
| 26 | `MODIFY_STAT.duration` / `GRANT_KEYWORD.duration` (`END_OF_TURN`, `END_OF_OPPONENT_TURN`, `PERSISTENT`) | `Duration`: the first two. A modify step is Layer 5. Armor never expires (`SVY-R39`) | No starter card grants a permanent modifier |
| 27 | `FR-TGT-101` bound payloads and `TargetRole`s, `FR-TRG-007` causality lineage | Not in `ability`. Choices go through `engine.Decision` and a per-seat `Chooser` ([ADR 0003](../../architecture/decisions/0003-synchronous-chooser-for-player-choices.md)). A resolution records what each step did, for `IfDone` | The wire format is M5 (the FRD 04 rewrite) |
| 28 | `FR-VAL-002` rules-text parity checker | `Def.Text` must equal the printed lines, or the card loads unsupported (§1) | A check on text identity only. Full parity stays frozen for FRD 08 (M6) |

### 3.3 Primitives (`FR-PRM-001`) → `Op`

FRD 02 has 26 primitives. Seventeen are `Op`s, two are folded into other shapes, and seven aren't built.

| FRD 02 primitive | `ability` | Notes |
| :--- | :--- | :--- |
| `DEAL_DAMAGE` | `OpDealDamage` | `damageSource` → `Step.Dealer` (`SVY-R27`) |
| `HEAL`, `ADD_ARMOR`, `KILL`, `RETURN_TO_HAND`, `SCRATCH`, `GAIN_CONTROL`, `MORPH_SCRATCH` | `OpHeal`, `OpAddArmor`, `OpKill`, `OpReturnToHand`, `OpScratch`, `OpGainControl`, `OpMorphScratch` | `OpKill` skips a Headquarters (`SVY-R55`) |
| `MODIFY_STAT` | `OpModifyStat` | `Step.Stat` (power or Maximum Health) and `Step.Until`. No `PERSISTENT` |
| `MOVE_ZONE` | `OpMoveZone` | `Step.To`. No `CHOSEN_ZONE` |
| `DRAW_CARD` | `OpDraw` | `SubjectController` or `SubjectOpponents`. "Each player" waits on #71 |
| `DISCARD_CARD` | `OpDiscard` | Exactly `Amount`. `ControllerChooses` covers "reveal, you choose" |
| `REVEAL_HAND` | Folded into `OpDiscard` + `ControllerChooses` | `FR-MOD-001` §3, `SVY-R50` |
| `DEPLOY_FREE` | `OpDeployFree` | From a discard pile, onto `Step.To` or the printed zone (`SVY-R43`) |
| `SEARCH_DECK` | `OpSearchDeck` | "A card" only (`SVY-R37`, `SVY-R38`). A search with criteria waits on #106 |
| `LOOK_TOP_CARDS` + `PartitionDestDef` | `OpLookTop` | One fixed partition: deploy up to `UpTo` matching `Filter`, and the rest go to the bottom (`SVY-R41`–`SVY-R43`). No general partition |
| `PUT_CARD_ON_DECK` | `OpPutOnDeck` | `Bottom`. No `SHUFFLED_IN` |
| `GRANT_KEYWORD` | `OpGrantKeyword` | Grants a whole `Keywords` set |
| `REDIRECT_DAMAGE_TARGET` | Folded into `WouldBeDealtDamage` | §3.2 row 24 |
| `REGROUP`, `TUCK_UNDER`, `REGISTER_DELAYED_TRIGGER` | Not built | `FR-KWD-002n`, `FR-ENT-102` (with #142), `FR-TRG-006`: the four M2 requirements moved to M3 include these three |
| `GENERATE_RESOURCE`, `RECYCLE_DISCARD_TO_DECK`, `EXHAUST`, `GRANT_SUBTYPE` | Not built | No starter card needs them |

---

## 4. For M3

### 4.1 Waiting Shapes

Each of these waits for the first card that needs it, and none is reachable by an Animus or Path card. Most are board directives with a checkable trigger.

| Shape | Where it waits |
| :--- | :--- |
| Draw and end-of-turn hooks, resolved at the right time | #15 (part 2) |
| A step that acts on each player | #71 |
| Costs on modes | #91 (2), `SVY-R33` |
| A deck search with criteria | #106 |
| A mandatory discard cost | #120 |
| Tucking, and the attached-card purge on a face-down transformation | `FR-ENT-102`, #142 |
| Regroup N, delayed triggers | `FR-KWD-002n`, `FR-TRG-006` |
| Overkill: its own hook, which fires from combat Step 3 and fixes the excess | `FR-KWD-002p`, design note Open Question 12 |
| A granted triggered ability ("gains 'Overkill: Ready this unit.'") | *Redline Infusion*, design note Open Question 12 |
| A count `Value`, reading `countPileCards` | #138 (applied; it names the reader the first count uses) |
| A targeted distributed pool | `FR-TGT-002`, *Macrocannon Barrage* |
| A Headquarters killed outright | Ruled, provisional (`SVY-R55`): no shape is needed |
| A combat-kill attacker that dies in the same sweep; a pile card that leaves and returns while targeted | Design note Open Questions 9 and 10 |

The rest of FRD 02's vocabulary that §3 marks *not built* (ValueToken aggregates, an opponent's-turn gate, from-zone gates, else branches, `negateSubtypes`, `readiness` filters, other stat operators, `PERSISTENT`, `SHUFFLED_IN`, `CHOSEN_ZONE`, `MOVE_CARD_ZONE`, `WOULD_DEPLOY`, subtype grants, aura modifiers other than power and keywords, and the four primitives in §3.3's last row) has no card yet. It should come back only when the M3 survey finds a card that prints it.

### 4.2 Shapes Pinned by One Card

Each of these has exactly one use in `pkg/engine/cards` (§2.2), so one card's tests are all that pin it. Before the M3 survey maps a second card onto one, it should check that card against that one test case: `HookExhausted`, `HookCombatKill`, `HookTurnStart`, `HookDiscardedFromHand`, `HookEntersDiscard`, `OpSearchDeck`, `OpMorphScratch`, `OpGrantKeyword`, `OpLookTop`, `OpGainControl`, `CostExhaust`, `CostDiscard`, `StaticDeployPermission`, both `Would`s, `SubjectReplaced`, `SubjectWatched`, `ZoneRefThat`, `UntilEndOfOpponentTurn`, and the fields `Watches`, `MaxPerTurn`, `IfHandEmpty`, `ControllerChooses`, `Split`, `Dealer`, `PerDiscard`, `PerKilled`, `Different`, `Another`, `Counts` and `CostReduction`.

### 4.3 Notes for the FRD 02/03 Rewrite

1. **Name the vocabulary as the Go names.** If FRD 02's names are `ability`'s (or a one-to-one mapping in a table), the catalog export is a mechanical serialization, and a later rename is an ADR 0007 move with one place to update.
2. **The export format is still open.** Only `Keywords` has JSON tags today. How `Def` serializes (field names, enum spellings, zero-value omission) is a decision for the M3 design note, recorded as a decision record when the export is built.
3. **Keep the rules out.** Each shape's field comment cites the `FR-*` and `SVY-R*` IDs that give its behavior. FRD 02 should cite them the same way and not restate them (FRD 01 owns them).
4. **The frozen parts stay frozen.** Authoring a second ability format, its loader, and the parity and sign-off gates (`FR-VAL-002`, `FR-CMP-020`) move to FRD 08 at M6 (`PRODUCT_SUMMARY.md` §4, 2026-11-21).
