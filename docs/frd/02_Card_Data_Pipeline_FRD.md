# Card Data & Ability Vocabulary - Functional Requirements Document (FRD v2.0)
**The Card Sheet Schema and the Ability Vocabulary of `pkg/engine/ability`**

> [!NOTE]
> **Status: Active, rewritten at the start of M3 (2026-11-23).** This replaces the v1.3 design-intent text, which is kept unedited as [`reference/02_Card_Data_Pipeline_FRD_v1.3_frozen.md`](reference/02_Card_Data_Pipeline_FRD_v1.3_frozen.md) for FRD 08 (M6). The file name is unchanged because `cardtown lint frd` expects it. FRD 01 and its `pkg/engine` scenario tests still win every rules question. For the exact fields of every shape, the authority is `go doc -all ./pkg/engine/ability`: where this document and that package disagree, the package wins and this document is corrected.

---

## EXECUTIVE SUMMARY & REVISION HISTORY

This document says what a card is as data: the CSV sheets in `data/cards/` that hold every card's printed characteristics and rules text, and the vocabulary in which a card's non-keyword abilities are written, which `pkg/engine/ability` defines and `pkg/engine` interprets. It exists so that three readers agree on one vocabulary: the kernel that executes a card, the designer who maps a new card onto it (the M3 ability survey), and the catalog export that serializes it mechanically. The export is built in M5, and its FRD ([`03_Card_Compiler_FRD.md`](03_Card_Compiler_FRD.md)) is frozen as design intent until then. It states no game rule. Each shape cites the `FR-*` and `SVY-R*` IDs of [FRD 01](01_Game_Rules_FRD.md) that give its behavior, and FRD 01 owns them.

It is written from what M2 built, not from what v1.3 proposed. M2 made 75 Animus and Path cards executable with 73 ability definitions as Go values (ADR-0004), with no compiler, and the [ability shapes write-up](../design/survey/m2_ability_shapes.md) records how the kernel's vocabulary differs from v1.3's. The sheets hold 169 cards in 10 files.

```text
data/cards/*.csv ── FR-DAT-006..008 ──> printed characteristics + rules-text lines
                                              │
                         standalone keyword line ── FR-ABL-012 ──> keyword set
                         other line ── named in Def.Text (FR-ABL-014) ──> ability.Def
                                              │                     (one per card name, FR-ABL-001)
pkg/engine/cards.Defs ────────────────────────┘
                                              │
                          ┌───────────────────┴───────────────────┐
                          ▼                                       ▼
                 pkg/engine interprets (FRD 01)         catalog export (M5; FRD 03, FR-CMP-023)
```

### Revision Summary (v1.3 -> v2.0):
- **Rewritten short at the M3 start.** The CSV schema (Domain 1) and the ability vocabulary (Domain 2) are kept and brought in line with what is built. The rules-text conventions (Domain 3) are kept. The authoring pipeline moved to M6 (`PRODUCT_SUMMARY.md` §4, 2026-11-21): the compiler, `ast_overrides.json`, the rules-text parity checker, designer sign-off and the strict prototype gate are not specified here. They stay in the frozen copy for FRD 08.
- **Names are the package's.** The polymorphic `AbilityDef` AST, `FilterDef`, `ValueToken`, `TargetRef`, `ConditionDef` and the 26 primitives are replaced by `Def`, `Trigger`, `Filter`, `Value`, `Target`, `Step`, `Op` and the other Go names. Where a v1.3 construct has a counterpart, the tables below say which.
- **Only what is built is specified.** A shape the kernel does not have waits for its first card (`FR-ABL-013`).
- **Retired IDs became Subsumed stubs** pointing to their new owner (the register at the end). No ID was deleted.
- **Amended 2026-11-23: the catalog export moved from M3 to M5** (`PRODUCT_SUMMARY.md` §4). Two sheet checks that FRD 03 had in the export's pre-flight (`FR-CMP-027`) stay in M3 and are owned here: a duplicate card name (`FR-DAT-002`), and a non-numeric value in a numeric column or a Unit or Structure with no zone (`FR-DAT-006`). Citations of FRD 03 below point at frozen design intent for M5 and imply no M3 work.

### Specification Scope (v2.0):
1. **The Card Sheets (`FR-DAT-001` - `FR-DAT-009`)**: source of truth, identity, the card record, headers and enums, subtypes and sentinels, deployment zones, invariants.
2. **The Ability Vocabulary (`FR-ABL-001` - `FR-ABL-014`, `FR-TRG-001`, `FR-TRG-007`, `FR-PRM-001`, `FR-TGT-101`)**: the definition, hooks, triggered, activated and static abilities, replacements, Maneuver effects, modes, steps, targets, filters, costs, keywords, resolution context, text identity, and how new vocabulary arrives.
3. **Rules-Text Conventions (`FR-TXT-001` - `FR-TXT-006`)** and **Sheet Checks (`FR-VAL-001`)**.

---

## PRODUCT DECISIONS FLAGGED FOR REVIEW

> [!IMPORTANT]
> **DECISION 1: Keep two FRDs, not one (`FR-DAT-001`, `FR-CMP-023`)**
> - **Inference**: The product summary lets FRD 02 and FRD 03 merge. The export moved to M5 (2026-11-23) and is rewritten then, on a different trigger from the vocabulary, which the M3 ability survey and every card task read first. FRD 04 and FRD 05 cite the `FR-CMP-*` IDs, and `cardtown lint frd` requires both file names to exist.
> - **Decision**: Two files. FRD 02 owns card data, the vocabulary and the sheet checks. [FRD 03](03_Card_Compiler_FRD.md) owns the catalog export, frozen as design intent until its rewrite at the start of M5.
> - **Alternative Considered**: One merged FRD. Rejected: the `FR-CMP-*` IDs would lose their home, and the lint check would need a toolchain change for no gain in clarity. FRD 03's file name still says "Compiler" although it no longer specifies one. Renaming it is a toolchain change (an open technical question there).

> [!IMPORTANT]
> **DECISION 2: The frozen v1.3 text is preserved as an unscanned copy (`reference/`)**
> - **Inference**: FRD 08 (M6) is drafted "drawing on frozen FRD 02/03", so the authoring text must stay readable. `cardtown lint frd` reads only the top level of `docs/frd/`, so it never scans `docs/frd/reference/`. A copy there defines no requirement twice, and its references are not checked. An archive left in the live file, or a second copy beside it, would define the same IDs twice and fail the ownership check.
> - **Decision**: The v1.3 and v1.2 text moved unedited (relative links fixed, a banner added) to `docs/frd/reference/02_Card_Data_Pipeline_FRD_v1.3_frozen.md` and `docs/frd/reference/03_Card_Compiler_FRD_v1.2_frozen.md`, beside the existing benchmark blueprints reference. Git history also keeps them.
> - **Alternative Considered**: Rely on git history alone. Rejected: a drafter has to know which commit to read. A hand-curated extract. Rejected: it would be a third text to keep right, and the point is that the original stays intact.

> [!IMPORTANT]
> **DECISION 3: A card is identified by its printed name until something stores identifiers (`FR-DAT-002`)**
> - **Inference**: v1.3 derived a blueprint ID slug from faction and name. The kernel has no blueprint ID: decks, ability definitions and the loader all key by card name, and the loader keeps the later of two same-named rows without complaint.
> - **Decision**: The name is the card's identity and must be unique. The catalog export, when it is built in M5, identifies a card by that name too, and no slug is derived ([ADR 0009](../architecture/decisions/0009-cards-identified-by-printed-name.md)).
> - **Alternative Considered**: Adopt the frozen slug rule now. Rejected: nothing reads it, and a rename policy (the frozen "permanently immutable" rule) is a product decision best made when something stores IDs.

> [!IMPORTANT]
> **DECISION 4: Rules-text conventions stay in FRD 02, condensed (`FR-TXT-001` - `FR-TXT-006`)**
> - **Inference**: The templating standard is cited as live guidance (the drafter and designer personas, `cardtown lint syntax`), and the ability text a `Def` names must match it. Domain 3 of v1.3 is not part of the authoring pipeline.
> - **Decision**: Keep the six IDs with their substance, trimmed to what the printed cards use. Only `cardtown lint syntax` enforces any of it (three forbidden terms), and the rest is reviewed by the designer.
> - **Alternative Considered**: Freeze them with the pipeline for FRD 08. Rejected: the cards in M3 are written to this standard now.

## OPEN TECHNICAL QUESTIONS

Questions this FRD raises but deliberately does not answer. `lead-architect` resolves them as decision records in `docs/architecture/decisions/` when the code needs them.

- How does a `Def` serialize: field names, enum spellings (the Go names, or a mapping to them), and whether a zero value is omitted? Only `Keywords` has JSON tags today (camelCase, omitted when zero).
- ~~What form does a card's catalog identifier take (`FR-DAT-002`)?~~ Answered: the printed name, exactly as the sheet spells it ([ADR 0009](../architecture/decisions/0009-cards-identified-by-printed-name.md)). A rename policy and a stable identifier column are decided at the start of M5.
- ~~Where are the duplicate-name check (`FR-DAT-002`) and the numeric-column and zone checks (`FR-DAT-006`) enforced: the loader or `cardtown lint sheets`?~~ Answered by `docs/architecture/m3-full-catalog.md` (Decisions): the sheet loader (`LoadCards`) reports them, not `cardtown lint sheets`.
- Does `needsRework` ever reach the kernel? It is read by no engine code today.
- Is a card's rules text kept per line in the catalog, or as the cell it came from?

---

## DOMAIN 1: THE CARD SHEETS

```text
data/cards/
  maneuvers-<faction>.csv              8 columns   (animus, concordia, path, soulless)
  units-and-structures-<group>.csv    14 columns   (animus, concordia, path, soulless, HQ, common)
```

### Section 1.1: Source of Truth, Identity and Record

#### FR-DAT-001: The Card Sheets Are the Source of Truth
- **Contract**: The CSV files in `data/cards/` hold every card's printed characteristics and rules text. Anything derived from them (the cards the kernel loads and, from M5, the catalog of `FR-CMP-023`) is regenerated and never edited. A sheet whose file name starts with `maneuvers-` uses the Maneuver columns of `FR-DAT-006`, and every other sheet uses the asset columns. A new CSV placed in the folder is read like the rest. The Google Sheets that the publisher writes are a downstream view: an edit there changes no game.
- **Invariants**: One printed card is one non-blank row, in exactly one sheet.
- **Edge Cases**: A sheet that is empty or has the wrong header fails `cardtown lint sheets` and names the file.

#### FR-DAT-002: Card Identity
- **Contract**: A card is identified by its printed name, unique across every sheet. Deck lists ([`FR-SET-001`](01_Game_Rules_FRD.md)) and ability definitions (`FR-ABL-001`) name cards by it. The printed name is the identity through M4 ([ADR 0009](../architecture/decisions/0009-cards-identified-by-printed-name.md)): the catalog export built in M5 (`FR-CMP-023`) identifies a card by it, and no slug is derived.
- **Invariants**: Two rows never share a name. Renaming a card renames it everywhere that names it (the sheet row, the definition's key, deck lists) in one change.
- **Edge Cases**: A duplicate name is a data error, found when the sheets are loaded. The report names the file, the later row and the card, and names the first row that holds the name.

#### FR-DAT-003: Blueprint Compilation & Asset Bundling
*(Subsumed by [`FR-CMP-002`](03_Card_Compiler_FRD.md) & [`FR-CMP-023`](03_Card_Compiler_FRD.md); runtime bundling is [`FR-ENT-001`](05_Headless_Game_Engine_FRD.md))*

#### FR-DAT-004: Engine Version Binding & Replay Compatibility
*(Subsumed by [`FR-CMP-024`](03_Card_Compiler_FRD.md)–[`FR-CMP-026`](03_Card_Compiler_FRD.md); the handshake and replay header stay with [`FR-SES-001`](04_Client_Engine_Protocol_Networking_FRD.md) & [`FR-SPC-002`](04_Client_Engine_Protocol_Networking_FRD.md))*

#### FR-DAT-005: The Card Record
- **Contract**: A card's record holds its printed characteristics (`FR-CRD-101`): name, card type, faction, subtypes, Requisition cost and Tactics cost, rarity and rules text, and for Units, Structures and Headquarters also deployment zones, power, health, Requisition yield and Tactics yield. It also holds the `needsRework` development flag, and, from the loaders, the card's standalone keyword set (`FR-ABL-012`) and its ability definition (`FR-ABL-001`) if it has one.
- **Invariants**: The record holds printed values only. A card's effective values come from the layers of `FR-LAY-001` at play time.
- **Edge Cases**: A Maneuver has no zones, subtypes, power, health or yields: they read as empty or zero.

### Section 1.2: Columns, Enums and Text

#### FR-DAT-006: Sheet Headers and Zero-Aliasing Enums
- **Contract**: Column headers are fixed, camelCase and in this order.
  - Asset sheets (14): `name`, `cardType`, `faction`, `subtypes`, `requisitionCost`, `tacticsCost`, `deploymentZones`, `rulesText`, `power`, `health`, `requisitionYield`, `tacticsYield`, `rarity`, `needsRework`.
  - Maneuver sheets (8): `name`, `cardType`, `faction`, `requisitionCost`, `tacticsCost`, `rulesText`, `rarity`, `needsRework`.
- Enum cells use the canonical spelling with no aliases. `cardType`: `UNIT`, `STRUCTURE`, `MANEUVER`, `HEADQUARTERS`. `faction`: `CONCORDIA`, `SOULLESS`, `PATH`, `ANIMUS`, `NEUTRAL`. `rarity`: `COMMON`, `UNCOMMON`, `RARE`, `MYTHIC`, `HQ`. `deploymentZones`: `RESERVES`, `MAIN_THEATER`, `HEADQUARTERS` or the one two-zone spelling `MAIN_THEATER, RESERVES`. `needsRework`: lowercase `true` or `false`.
- Cells are RFC 4180 quoted text, and a rules-text cell holds one printed line per line break. The kernel's loader treats LF and CRLF line endings the same and accepts a byte-order mark before the first header. `cardtown lint sheets` is stricter and wants the headers exactly.
- **Invariants**: A header or enum that differs fails `cardtown lint sheets`, with the expected and found headers shown. The kernel's loader reports an unknown `cardType` with the other data errors below, and never ends the load at the first one.
- **Edge Cases**: Two things the sheet linter does not check are data errors that loading the sheets reports, naming the file, row (the header is row 1), card and column: a numeric cell that is not a whole number (an empty cell or `--` included), and a Unit or Structure with no deployment zone. Requisition cost and Tactics cost are numeric on every card. Power, health and both yields are numeric on every card except a Maneuver. Loading reports every problem at once, in file then row order, and returns no cards when there is any. A card is never loaded with a guessed value.

#### FR-DAT-007: Subtypes, Sentinels and Blank Rows
- **Contract**: `subtypes` is a comma-separated list. Spaces around an entry are dropped, and an empty cell or `--` means none. Subtype comparisons are case-insensitive (the subtype criterion of `FR-TGT-001`). In `rulesText`, each non-empty line is one printed line, and a cell that is `--` or empty prints nothing (a vanilla card). A row with an empty `name` is a placeholder and is skipped everywhere.
- **Invariants**: Each printed line is exactly one of: a standalone keyword line (`FR-ABL-012`), a line a definition claims (`FR-ABL-014`), or an unsupported line.
- **Edge Cases**: A card with an unsupported line cannot be deployed or enacted until a definition claims it. This is how a card with unbuilt text is held back without breaking the sheet.

#### FR-DAT-008: Deployment Zones
- **Contract**: A Unit or Structure lists Reserves, Main Theater or both, with Main Theater first (`MAIN_THEATER, RESERVES`). A Headquarters lists `HEADQUARTERS`. A Maneuver has no zone column. What the zones mean is `FR-ZON-001` and `FR-TRN-003`.
- **Edge Cases**: A card that lists both lets its controller choose where it deploys (`FR-TRN-003a`).

#### FR-DAT-009: Card Data Invariants
- Loading is a pure function of the sheets and the definitions: the same inputs give the same cards, with no clock, no randomness and no dependence on the order files are listed in.
- Nothing in the vocabulary or the interpreter names a card. A card's behavior is data (`FR-ABL-001`).
- A printed line and its definition cannot drift apart silently (`FR-ABL-014`).

---

## DOMAIN 2: THE ABILITY VOCABULARY

```text
Def (one per card name)
 ├─ Triggers[]      Hook + gates + Effect                    FR-TRG-001, FR-ABL-002
 ├─ Activated[]     Costs[] + Effect                          FR-ABL-003, FR-ABL-009
 ├─ Statics[]       aura | deploy permission                  FR-ABL-004
 ├─ Cornered        a grant while Cornered                    FR-ABL-004
 ├─ CostReduction   hand-active Requisition reduction         FR-ABL-004
 ├─ Replacements[]  Would + Filter + Steps                    FR-ABL-005
 ├─ Effect          a Maneuver's                              FR-ABL-006
 ├─ ExtraCosts[]    declaration costs                         FR-ABL-006, FR-ABL-009
 └─ Text[]          the printed lines it implements           FR-ABL-014

Effect = Targets[] + Steps[]   or   Modes[] + ChooseN         FR-ABL-010
Step   = Op + Subject (+ Target, Filter, Amount, ...)          FR-PRM-001, FR-TGT-101
```

### Section 2.1: The Definition and Its Parts

#### FR-ABL-001: One Definition Per Card
- **Contract**: A card's non-keyword abilities are one `ability.Def`, keyed by the card's name in `pkg/engine/cards`. A `Def` has the parts drawn above, each a list in printed order or a single value, and a part a card does not print is empty. A card may print several kinds at once (a trigger and a Cornered grant). Standalone keywords are not a part of a `Def` (`FR-ABL-012`), so a card with only keyword lines, or a vanilla card, has none.
- **Invariants**: The kernel interprets definitions and holds no card-specific logic (ADR-0004). A definition is data, so the catalog export (M5) reads the same value the kernel runs.
- **Edge Cases**: The v1.3 polymorphic `AbilityDef` with an `abilityType` is gone: the part a value sits in says what it is.

#### FR-TRG-001: Hook and Replacement Event Names
- **Contract**: A triggered ability names one `Hook`, and a replacement names one `Would`. These are the whole set:

| Hook | Names | Rules IDs |
| :--- | :--- | :--- |
| `HookDeploy` | this asset was deployed | `FR-TRN-003` |
| `HookKilled` | this asset was destroyed from the battlefield | `FR-ZON-004` |
| `HookAttacks` | this asset attacked (with `Watches`, an asset matching `Watch` attacked) | `FR-CBT-002` |
| `HookDamaged` | this asset took damage | `FR-REP-004`, `FR-KWD-002a` |
| `HookExhausted` | this asset changed from Readied to Exhausted | `SVY-R17`, `FR-CAP-001` |
| `HookCombatKill` | this asset attacked and its combat damage destroyed the defender | `FR-CBT-002`, `SVY-R19` |
| `HookMoves` | this unit moved to a new battlefield zone | `FR-ZON-002`, `SVY-R1` |
| `HookDiscardedFromHand` | this card went from its owner's hand to the discard pile | `FR-ZON-004` |
| `HookTurnStart` | a turn began | `FR-TRN-001` |
| `HookEntersDiscard` | this card entered its owner's discard pile from any zone | `FR-ZON-004`, `FR-ZON-005` |

| Would | Names | Rules IDs |
| :--- | :--- | :--- |
| `WouldBeDestroyed` | an asset matching the replacement's filter would be destroyed | `FR-REP-001`, `SVY-R15` |
| `WouldBeDealtDamage` | an asset matching the filter would be dealt damage; it goes to the source | `FR-REP-004` |

- **Invariants**: Each `Hook` fires on its own, gates on its own and counts on its own. A compound preamble ("deploys or attacks") is two triggers sharing one effect. The frozen taxonomy's other hooks (v1.3 `FR-TRG-001` had about 30) are not part of the vocabulary until a card needs one (`FR-ABL-013`).

### Section 2.2: The Parts

#### FR-ABL-002: Triggered Abilities
- **Contract**: A `Trigger` has one `Hook`, an `Effect`, and these gates and options, all optional.

| Field | Names | Rules IDs |
| :--- | :--- | :--- |
| `Turn` | any turn (`TurnAny`) or only the controller's own (`TurnYours`) | `FR-TRN-001` |
| `ToZone` | for a deploy or a move, only into this zone | `FR-ZON-002` |
| `ActiveZone` | the battlefield (`ZoneAny`) or the owner's discard pile (`ZoneDiscard`) | `FR-ZON-004` |
| `Cornered` | fires only while its controller is Cornered | `FR-KWD-002e`, `FR-LAY-003` |
| `Optional` | "you may": asked as the trigger is placed | `SVY-R9`, `SVY-R47`, `SVY-R48` |
| `FromKeyword` | the keyword this trigger comes from (only `KeywordTriggerAggressive`) | `FR-KWD-002f` |
| `Watches`, `Watch` | fire on the event of any asset matching a filter (only for `HookAttacks`) | `FR-CBT-002` |
| `MaxPerTurn` | at most this many firings each turn, per asset (`FR-ABL-008`) | `FR-TRN-001` |

- **Invariants**: A trigger is placed and ordered under `FR-TRG-003` and `FR-TRG-004`, and carries no resource cost (`FR-TRG-005`). A trigger never works from a hand.
- **Edge Cases**: An effect with "you may X. If you do, Y" is a step list with `IfDone` (`FR-TGT-001`, `SVY-R14`), not a cost.

#### FR-ABL-003: Activated Abilities
- **Contract**: An `Activated` is a list of `Costs`, paid in printed order, and an `Effect` (`FR-ACT-001`, `FR-ACT-002`).
- **Invariants**: Its targets are chosen when it is activated, as for a Maneuver (`SVY-R36`).
- **Edge Cases**: No limit on use and no "repeatable" flag exist beyond what the costs impose (`FR-ABL-013`).

#### FR-ABL-004: Static Abilities, Cornered Grants and Cost Reduction
- **Contract**: A `Static` has a `Kind`.
  - `StaticAura`: every battlefield asset matching its `Filter`, relative to the source's controller, has a power (Layer 4) and a keyword set (Layer 3) (`FR-LAY-001`), and a Cornered grant (`WhileCornered`: keywords and power) while its own controller is Cornered (`FR-KWD-002e`, `SVY-R18`). "In the same zone" reads the source's current zone (`SVY-R44`).
  - `StaticDeployPermission`: the card may be deployed from `ActiveZone` (its discard pile) by paying `Costs` in addition (`FR-TRN-003`, `FR-CST-001`).
- A card's own Cornered grant is `Def.Cornered`. Its hand-active Requisition reduction is `Def.CostReduction`: `Req` off for each discard-pile card matching a pile `Filter` (`FR-CST-001`).
- **Invariants**: An aura's filter reads no stat and no keyword (a Layer 3 or 4 value cannot be read by its own filter).

#### FR-ABL-005: Replacement Effects
- **Contract**: A `Replacement` has a `Would` event, a `Filter` and `Steps`. The steps run inline, never on the stack, with `SubjectSelf` the replacement's source and `SubjectReplaced` the asset whose event is replaced (`FR-REP-001`, `FR-REP-004`). `WouldBeDealtDamage` has no steps, because the damage goes to the source.
- **Edge Cases**: The order and choice among several replacements is `SVY-R20` to `SVY-R24`.

#### FR-ABL-006: Maneuver Effects and Declaration Costs
- **Contract**: A Maneuver's resolving effect is `Def.Effect` (`FR-MNV-001`). Costs paid when any card is declared, beyond its Requisition and Tactics, are `Def.ExtraCosts` in printed order (`FR-TRN-003a`).
- **Invariants**: `Def.Effect` is empty for every card that is not a Maneuver.

#### FR-ABL-008: Per-Turn Trigger Limit
- **Contract**: `Trigger.MaxPerTurn` limits how often a battlefield trigger fires each turn, per asset. Zero means no limit (`FR-TRN-001`).
- **Edge Cases**: No other part has a per-turn limit (`FR-ABL-013`).

#### FR-ABL-009: Costs
- **Contract**: A `Cost` has a `CostKind`. The vocabulary is:

| CostKind | Names | Rules IDs |
| :--- | :--- | :--- |
| `CostResources` | pay `Req` Requisition and `Tac` Tactics | `FR-ACT-001` |
| `CostExhaustSelf` | exhaust the ability's source, which must be Readied | `FR-ACT-002`, `SVY-R17` |
| `CostKillSelf` | kill the ability's source through the kill path | `FR-ACT-001`, `SVY-R15` |
| `CostKill` | kill one allied asset matching `Filter`, or with `Counts` a chosen number of different ones | `FR-TRN-003a`, `SVY-R53` |
| `CostExhaust` | exhaust one Readied allied asset matching `Filter` | `FR-TRN-003a`, `SVY-R17` |
| `CostDiscard` | discard up to `UpTo` cards from hand, possibly none | `FR-TRN-003a` |

- **Invariants**: Choosing what pays a cost is not targeting, so Ward does not restrict it (`FR-TGT-003`).

#### FR-ABL-010: Modes
- **Contract**: An `Effect` with `Modes` lists each mode as an `Effect` in printed order, and `ChooseN` says how many different modes are chosen ("choose one" is 1). The modes are chosen as the item is placed, before any target (`FR-MOD-001` §5, `SVY-R6`, `SVY-R28`–`SVY-R31`).
- **Invariants**: An effect with modes has no `Targets` or `Steps` of its own. A mode has no cost of its own (`FR-ABL-013`, `SVY-R33`).

#### FR-ABL-011: Pools
- **Contract**: A pool the controller splits among assets as a step resolves is `Step.Split`, on a step that adds Armor to an untargeted pick (`SVY-R51`, `SVY-R52`). It is not a target (`FR-TGT-002`).
- **Edge Cases**: A pool split among chosen targets is not part of the vocabulary yet (`FR-ABL-013`).

### Section 2.3: Steps, Targets and Shared Vocabulary

#### FR-PRM-001: Steps and Their Ops
- **Contract**: An `Effect`'s `Steps` run in order. A `Step` has an `Op`, a `Subject` (`FR-TGT-101`) and the parameters its `Op` reads. This is the whole set. The frozen name is v1.3's primitive.

| Op | Frozen name | Acts on | Rules IDs |
| :--- | :--- | :--- | :--- |
| `OpScratch` | `SCRATCH` | a player | `FR-ZON-005` |
| `OpDraw` | `DRAW_CARD` | a player | `FR-WIN-001` |
| `OpDiscard` | `DISCARD_CARD` (and `REVEAL_HAND`, folded in with `ControllerChooses`) | a player | `FR-MOD-001` §1, §3, §6; `SVY-R50` |
| `OpSearchDeck` | `SEARCH_DECK` | a player | `FR-DEK-001`, `SVY-R37`, `SVY-R38` |
| `OpLookTop` | `LOOK_TOP_CARDS` | a player | `FR-MOD-001` §4; `SVY-R41`–`SVY-R43` |
| `OpDealDamage` | `DEAL_DAMAGE` | an asset | `FR-REP-004`, `SVY-R27` |
| `OpHeal` | `HEAL` | an asset | `FR-KWD-002d` |
| `OpAddArmor` | `ADD_ARMOR` | an asset | `FR-LAY-002` |
| `OpModifyStat` | `MODIFY_STAT` | an asset | `FR-LAY-001`–`FR-LAY-003` |
| `OpGrantKeyword` | `GRANT_KEYWORD` | an asset | `FR-LAY-001`, `FR-LAY-003` |
| `OpKill` | `KILL` | an asset | `FR-LAY-002`, `FR-ZON-002`, `SVY-R46` |
| `OpMoveZone` | `MOVE_ZONE` | a unit | `FR-ZON-002`, `FR-ZON-004` |
| `OpGainControl` | `GAIN_CONTROL` | a unit | `FR-ZON-002`, `FR-ZON-003` |
| `OpMorphScratch` | `MORPH_SCRATCH` | a unit | `FR-ENT-103` |
| `OpReturnToHand` | `RETURN_TO_HAND` | an asset or card | `FR-ZON-004` |
| `OpPutOnDeck` | `PUT_CARD_ON_DECK` | an asset or card | `FR-DEK-001` |
| `OpDeployFree` | `DEPLOY_FREE` | a card in a discard pile | `FR-TRN-003a`, `FR-CST-001`, `SVY-R43` |

- The parameters a step may carry are `Amount` (a `Value`), `Target` (an index into the effect's targets), `Filter`, `UpTo`, `Dealer`, `To` (a zone), `Keywords`, `Stat`, `Until` (a `Duration`), `Bottom`, `IfDone`, `IfHandEmpty`, `ControllerChooses` and `Split`. Each is read only by the ops its Go comment names, and loading rejects a definition that sets one where nothing reads it (`FR-ABL-014`). The Go comment on each `Op` is the authority for what it skips (for example, which cards `OpKill` or `OpMoveZone` leave alone).
- **Invariants**: `IfDone` and `IfHandEmpty` are the only conditions between steps. There is no else branch.
- **Edge Cases**: The v1.3 `REDIRECT_DAMAGE_TARGET` is folded into `WouldBeDealtDamage` (`FR-ABL-005`). The other seven frozen primitives have no `Op` (`FR-ABL-013`).

#### FR-TGT-101: Targets and Subjects
- **Contract**: A `Target` is one "chosen" clause of an `Effect`: exactly one entity, or with `UpTo` "up to N" different ones (`SVY-R11`). It is locked when the item is placed (`SVY-R3`, `FR-TGT-001`, `FR-TGT-003`). Its repetition and distinctness options are `PerDiscard` (`SVY-R4`), `PerKilled` with `Different` (`SVY-R53`, `SVY-R54`) and `Another` ("another chosen", `SVY-R34`). A `Subject` says what a step acts on:

| Subject | Names |
| :--- | :--- |
| `SubjectController` (the zero value) | the ability's controller |
| `SubjectOpponents` | each opponent of the controller |
| `SubjectSelf` | the ability's own source card, tracked across zones |
| `SubjectTarget` | each entity chosen for the target slot the step names |
| `SubjectAll` | every battlefield asset matching the step's filter: not a target (`FR-TGT-002`) |
| `SubjectPick` | entities the controller picks as the step resolves, among assets or discard-pile cards: not a target (`SVY-R3`, `SVY-R14`) |
| `SubjectReplaced` | the asset a replacement replaces (replacements only) |
| `SubjectWatched` | the asset whose event fired a watching trigger: not a target |

- **Invariants**: Ward and `FR-TGT-003` apply to "chosen" only, so `SubjectAll`, `SubjectPick` and `SubjectWatched` never fall under them.
- **Edge Cases**: The wire form of a target choice is M5 work (FRD 04). It was `FR-TGT-101`'s payload in v1.3.

#### FR-ABL-007: Filters, Values, Durations and Zones
- **Contract**: A `Filter` is a conjunction of criteria (`FR-TGT-001`), and every criterion that is set must hold. A criterion left at its zero value is not applied.

| Criterion | Values |
| :--- | :--- |
| `Allegiance` | `AllegianceAny`, `AllegianceAllied`, `AllegianceOpposing`, relative to the ability's controller |
| `Kinds` | a set of `KindUnit`, `KindStructure`, `KindHeadquarters` and, for discard-pile filters only, `KindManeuver` |
| `Subtypes` | at least one of a list of printed subtypes |
| `Scratch` | `ScratchAny`, `ScratchOnly`, `ScratchExcluded` (Scratch Company units) |
| `Other` | not the ability's own source |
| `Zone` | `ZoneAny` (any battlefield zone), `ZoneReserves`, `ZoneMainTheater`, `ZoneDiscard` (cards in a discard pile) |
| `ZoneRef` | `ZoneRefSame`, the zone the source was in when the item was placed (`SVY-R44`), or `ZoneRefThat`, the zone locked as a move's destination (`SVY-R45`) |
| `Stat`, `AtMost` | the entity's `StatPower`, `StatRequisition` or `StatMaxHealth` is `AtMost` or less. A pile card reads printed, an asset effective |

- A `Value` is a number a step reads as it resolves: fixed (`ValueSourceFixed`), or the effective power of the step's dealer (`ValueSourcePower`, `SVY-R10`, `SVY-R27`). A `Duration` is `UntilEndOfTurn` or `UntilEndOfOpponentTurn` (`FR-LAY-003`). `OpAddArmor` has no duration (`SVY-R39`).
- **Invariants**: The frozen filter and value options with no counterpart here (negated subtypes, readiness, other stat operators, aggregates and counts) are not part of the vocabulary (`FR-ABL-013`).

#### FR-TRG-007: What a Resolution Can Refer To
- **Contract**: An effect may refer back to what happened earlier in the same placement or resolution, and only in these ways: a step that runs only if the step before it did something (`IfDone`, `SVY-R14`, `SVY-R15`), a target slot repeated once per discarded or killed card (`PerDiscard`, `PerKilled`), a chosen asset as the damage dealer (`Dealer`), an earlier slot another must differ from (`Another`), the asset whose event fired a watching trigger (`SubjectWatched`) and a source's last-known power once it has left (`SVY-R10`).
- **Invariants**: The frozen causality lineage (event parents, root actions) is not part of the vocabulary. The choice a player makes goes through the kernel's decision interface (ADR-0003), and its wire form is M5 work.

### Section 2.4: Keywords, Text Identity and Growth

#### FR-ABL-012: Standalone Keywords
- **Contract**: A rules-text line that is exactly one keyword, with its number where it has one (`Resistant 1`, `Armor 2`, `Ward`), is a standalone keyword line. Loading folds these into the card's `ability.Keywords` set, never into a `Def`. The set has nine members, each defined by its keyword entry in FRD 01 Appendix A: `Resistant` (`FR-KWD-002b`), `Stalwart` (`002c`), `Aggressive` (`002f`), `Ethereal` (`002g`), `Ward` (`002h`), `Versatile` (`002i`), `Guardian` (`002j`), `Armor` (`002k`) and `DeploysExhausted` (`002l`). The same set describes keywords granted by an aura, by Cornered, or by `OpGrantKeyword`, and the one fold of two sets adds numbers and sets flags.
- The other keywords are written in the vocabulary, not in the set:

| Keyword (FRD 01) | Written as |
| :--- | :--- |
| `FR-KWD-002a` | a `Trigger` on `HookDamaged` |
| `FR-KWD-002e` | `Trigger.Cornered`, `Def.Cornered` or `Static.Cornered` |
| `FR-KWD-002f` | `ability.Aggressive(n)`: a `Trigger` with `FromKeyword` set, plus the set member |
| `FR-KWD-002d` | `OpHeal` |
| `FR-KWD-002m` | `OpScratch` |

- **Invariants**: This document names keywords and cites their IDs. Their behavior is in FRD 01 Appendix A only.
- **Edge Cases**: `FR-KWD-002n` and `FR-KWD-002p` have no shape yet (`FR-ABL-013`). `FR-KWD-002o` is retired (`SVY-R12`). An unknown line such as "Overkill" is not a keyword line: it stays an unsupported line until a definition claims it.

#### FR-ABL-013: New Vocabulary Arrives With Its First Card
- **Contract**: A shape the kernel does not have is added only when a card needs it, in a `shape` task that changes the package, its interpreter and its tests together (ADR-0004). A `card` task only combines shapes that exist. This document is updated in the same milestone to name the new value.
- **Invariants**: Nothing is specified ahead of its card. The shapes still waiting are tracked in the [ability shapes write-up](../design/survey/m2_ability_shapes.md) §4.1 and in the board directives. The frozen v1.3 text lists what a past design wanted, and is a source to read, not a requirement.
- **Edge Cases**: A card that needs a shape the vocabulary lacks stays held back (`FR-DAT-007`) until the shape task lands.

#### FR-ABL-014: A Definition Matches the Printed Text
- **Contract**: `Def.Text` lists the exact non-keyword lines the definition implements, and each must equal a line printed on the card, once. Loading removes each from the card's unsupported lines. A card with a line left over cannot be deployed or enacted.
- **Invariants**: Loading rejects, and names the card, (a) a `Text` line missing from the sheet, so a reworded card stops running instead of running stale logic, and (b) a definition that sets a field in a context nothing reads it in (a watching trigger on any hook but `HookAttacks`, a pool split anywhere but an untargeted Armor pick, `ControllerChooses` off `OpDiscard`, a stat criterion on an aura's filter).
- **Edge Cases**: This checks identity of text only. Full rules-text parity, comparing what a line says with what its definition does, stays frozen for FRD 08 (v1.3 `FR-VAL-002`).

---

## DOMAIN 3: RULES-TEXT CONVENTIONS & SHEET CHECKS

#### FR-TXT-001: Sentence Grammar and Capitalization
Every printed clause is a complete sentence that starts with a capital letter and ends with a period, unless it is a standalone keyword line (`FR-ABL-012`). System names are capitalized: `Headquarters`, `Reserves`, `Main Theater`, `Scratch Company`, `Requisition`, `Tactics`, `Health`. Reminder text is not part of `rulesText`.

#### FR-TXT-002: Trigger Preambles
`When` opens an event that happens once in an object's life ("When this unit deploys"). `Whenever` opens one that can repeat ("Whenever this unit attacks"). `At` opens a turn boundary ("At the start of your turn").

#### FR-TXT-003: Cost Notation
An activated ability prints `[Cost], [Cost]: [Effect].`, with resource costs first (`FR-ABL-003`). An additional cost to enact or deploy prints as its own leading line ("As an additional cost to enact this maneuver, kill an allied unit.", `FR-ABL-006`).

#### FR-TXT-004: Targeting Phrasing
Text says `chosen`, never `target` ("chosen allied unit"), and states the eligibility filter (`FR-TGT-001`). Verification and fizzling are `FR-TGT-003`.

#### FR-TXT-005: Stat and Defense Phrasing
A temporary modifier says "gets" (singular) or "get" (plural). A continuous aura says "has" or "have". A keyword or Armor grant says "gains" or "gain". A modifier states its power with a sign ("+1 power").

#### FR-TXT-006: Zone Names and Numbers
Zone names are capitalized and complete (`Main Theater`, `Reserves`, `Headquarters Zone`) except `discard pile` and `deck`. A card outside the battlefield is a card, and on it a unit, structure or asset. Cards drawn or discarded are counted in words ("Draw 2 cards"), and damage, Armor and keyword numbers in numerals ("Deal 2 damage", "Scratch 1").

#### FR-VAL-001: Sheet Checks
- **Contract**: `cardtown lint sheets` checks headers, enums and the `needsRework` boolean (`FR-DAT-006`). `cardtown lint syntax` rejects three forms in rules text: "friendly" (write "allied"), raw `~`, and "+N attack" (write "+N power"). Loading checks the identity and shape rules of `FR-ABL-014`. A duplicate card name (`FR-DAT-002`), a non-numeric value in a numeric column and a Unit or Structure with no zone (`FR-DAT-006`) are also reported when the sheets are loaded.
- **Invariants**: A sheet change that breaks a check blocks the commit (the pre-commit hook runs both).

#### FR-VAL-002: Rules Text vs Declarative Ability Parity Checker
*(Subsumed by `FR-ABL-014` for text identity. Full parity, and the compiler checks that carried it, stay frozen for FRD 08 (M6): [`FR-CMP-019`](03_Card_Compiler_FRD.md), [`FR-CMP-020`](03_Card_Compiler_FRD.md))*

#### FR-VAL-003: Uniform Dataset Validation Policy & Review Card Compilability
*(Subsumed by `FR-DAT-006`, which holds the `needsRework` column. Its strict production gate stays frozen for FRD 08 (M6): [`FR-CMP-005`](03_Card_Compiler_FRD.md), [`FR-CMP-021`](03_Card_Compiler_FRD.md))*

---

## DOMAIN 4: FAILURE BEHAVIOR & INVARIANTS

### Section 4.1: When Things Go Wrong

| Situation | What the Operator or Designer Experiences | Rule or Invariant Preserved |
| :--- | :--- | :--- |
| A `Def` lists a text line the sheet does not print | Loading fails and names the card and the line | A definition never runs against text it was not written for (`FR-ABL-014`) |
| A printed line no definition claims | The card loads but cannot be deployed or enacted, and the sim header counts it as not executable | A card never runs with text the kernel cannot execute (`FR-DAT-007`) |
| A definition sets a field nothing reads | Loading fails and names the card | A definition never silently does nothing or the wrong thing (`FR-ABL-014`) |
| A sheet has a wrong header or an illegal enum | `cardtown lint sheets` fails with the file, row and value | Zero aliasing (`FR-DAT-006`) |
| A card needs a shape the vocabulary lacks | The card stays unsupported and is listed for a shape task | Nothing is specified ahead of its card (`FR-ABL-013`) |
| Two rows share a name | Loading reports a data error naming both rows | One card, one identity (`FR-DAT-002`) |
| A numeric column holds a non-number, or a Unit or Structure has no zone | Loading reports the file, row, card and column | A card is never loaded with a guessed value (`FR-DAT-006`) |

### Section 4.2: Invariants

`FR-DAT-009` holds for this whole document.

---

## MASTER CROSS-REFERENCE REGISTER

| Requirement ID | Scope Summary | Interoperability Target |
| :--- | :--- | :--- |
| `FR-DAT-001` | The sheets are the source of truth | `FR-CMP-029` |
| `FR-DAT-002` | Card identity | `FR-SET-001`, `FR-CMP-023` (M5) |
| `FR-DAT-003` | Subsumed stub | `FR-CMP-002`, `FR-CMP-023` |
| `FR-DAT-004` | Subsumed stub | `FR-CMP-024`–`FR-CMP-026` |
| `FR-DAT-005` | The card record | `FR-CRD-101`, `FR-CMP-023` |
| `FR-DAT-006` | Headers, zero-aliasing enums, numeric and zone checks | `FR-VAL-001` |
| `FR-DAT-007` | Subtypes, sentinels, blank rows | `FR-TGT-001`, `FR-ABL-012` |
| `FR-DAT-008` | Deployment zones | `FR-ZON-001`, `FR-TRN-003` |
| `FR-DAT-009` | Card data invariants | `FR-ABL-014` |
| `FR-ABL-001` | One definition per card | `FR-PRM-001`, `FR-CMP-023` |
| `FR-ABL-002` | Triggered abilities | `FR-TRG-001`, `FR-TRG-003`–`FR-TRG-005` |
| `FR-ABL-003` | Activated abilities | `FR-ACT-001`, `FR-ACT-002` |
| `FR-ABL-004` | Statics, Cornered grants, cost reduction | `FR-LAY-001`, `FR-KWD-002e`, `FR-CST-001` |
| `FR-ABL-005` | Replacement effects | `FR-REP-001`, `FR-REP-004` |
| `FR-ABL-006` | Maneuver effects and declaration costs | `FR-MNV-001`, `FR-TRN-003a` |
| `FR-ABL-007` | Filters, values, durations, zones | `FR-TGT-001`, `FR-LAY-003` |
| `FR-ABL-008` | Per-turn trigger limit | `FR-TRN-001` |
| `FR-ABL-009` | Costs | `FR-ACT-001`, `FR-TRN-003a` |
| `FR-ABL-010` | Modes | `FR-MOD-001` |
| `FR-ABL-011` | Pools | `FR-TGT-002` |
| `FR-ABL-012` | Standalone keywords | `FR-KWD-002b`–`FR-KWD-002l` |
| `FR-ABL-013` | New vocabulary with its first card | ADR-0004 |
| `FR-ABL-014` | A definition matches the printed text | `FR-DAT-007` |
| `FR-TRG-001` | Hook and replacement event names | `FR-ABL-002`, `FR-ABL-005`, `FR-REP-001` |
| `FR-TRG-007` | What a resolution can refer to | `FR-TGT-001`, `FR-ABL-002` |
| `FR-PRM-001` | Steps and their ops | `FR-ABL-001`, `FR-TGT-101` |
| `FR-TGT-101` | Targets and subjects | `FR-TGT-001`, `FR-TGT-003` |
| `FR-TXT-001`–`FR-TXT-006` | Rules-text conventions | `FR-ABL-014`, `FR-VAL-001` |
| `FR-VAL-001` | Sheet checks | `FR-DAT-006` |
| `FR-VAL-002`, `FR-VAL-003` | Subsumed stubs (frozen for FRD 08) | `FR-ABL-014`, `FR-DAT-006` |

---
*End of Functional Requirements Document (Card Data & Ability Vocabulary FRD v2.0)*
