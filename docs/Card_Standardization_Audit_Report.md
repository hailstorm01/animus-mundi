# Card Town: Card Standardization Audit Report

> **Scope**: Complete audit of all 167 cards across 10 balance sheets  
> **Companion FRD**: [Card Data Pipeline FRD](file:///docs/frd/Card_Data_Pipeline_FRD.md) (Domain 7)  
> **Grammar Standard**: Magic: The Gathering (MTG) syntax & Zero-Aliasing Target Enums (FR-DAT-006)

---

## 1. Executive Summary & Audit Metrics

| Audit Metric | Count | Percentage | Description |
| :--- | :--- | :--- | :--- |
| **Total Cards Audited** | **167** | **100.0%** | All cards across all 10 balance CSV sheets |
| **Rules Text Standardized** | **127** | **76.0%** | MTG capitalization, punctuation, zones, keywords |
| **Flagged for Designer Review** | **52** | **31.1%** | Ambiguities, rounding rules, costs, requiring sign-off |
| **Rules Text Unchanged** | **40** | **24.0%** | Cards already conforming to grammar guidelines |

### Faction Breakdown

| Faction | Total Cards | Maneuvers | Units & Structures | Headquarters |
| :--- | :--- | :--- | :--- | :--- |
| **ANIMUS** |                                          .Count | 12 | 28 | 2 |
| **CONCORDIA** |                                         .Count | 12 | 27 | 2 |
| **PATH** |                                          .Count | 13 | 27 | 2 |
| **SOULLESS** |                                         .Count | 13 | 26 | 2 |
| **COMMON / NEUTRAL** |  | 0 |  | 0 |

---

## 2. High-Priority Designer Review Items (Domain 7.1)

| Card Title | Faction / Type | Previous Rules Text | Standardized Rules Text | Designer Review Flag / Rationale | Companion Requirement |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **Animoceps Rager** | ANIMUS HQ | When this HQ is dealt damage, scratch once. This effect happens once per turn. | Whenever this Headquarters is dealt damage, Scratch 1. This ability triggers only once each turn. | Formalized per-turn limit syntax (maxPerTurn: 1). | FR-ABL-008 |
| **Funeral Fungus** | ANIMUS Maneuver | exhaust chosen ready allied unit and return 1 unit card from your discard pile to your hand. | As an additional cost to enact this maneuver, exhaust a readied allied unit. Return chosen unit card from your discard pile to your hand. | Templated exhaust as additional enactment cost rather than spell effect. | FR-ABL-009 |
| **Rift** | ANIMUS HQ | discard a card, exhaust: scratch 2, deal 5 damage to this asset | Discard a card, exhaust: Scratch 2. Deal 5 damage to this Headquarters. | Confirmed Headquarters self-damage keyword interaction (Unbowed). | FR-REP-004 |
| **Mutation Swell** | ANIMUS Maneuver | kill up to 4 allied scratch company units. for every 2 units killed, deploy 1 unit from your discard pile for free. | Kill up to 4 allied Scratch Company units. For every 2 units killed this way, deploy a unit card from your discard pile for free. | Clarified causality ('killed this way'); confirm odd-number rounding. | FR-ABL-007 |
| **Steel Rain** | PATH Maneuver | discard up to 3 cards. For each card discarded this way, deal 2 damage to chosen asset. then, if you have no cards left in your hand, draw 2 cards. | Discard up to 3 cards. For each card discarded this way, deal 2 damage to chosen asset. Then, if you have no cards in hand, draw 2 cards. | Confirm if 'chosen asset' is selected once or independently per card discarded. | FR-ABL-011 |
| **Macrocannon Barrage** | SOULLESS Maneuver | deal 5 damage distributed among up to 5 chosen units | Distribute 5 damage among up to 5 chosen units. | Templated as distributed damage pool (DISTRIBUTED_POOL). | FR-ABL-011 |
| **Reverberating Terror** | CONCORDIA Maneuver | up to 2 chosen units and structures get -1 power until end of turn. | Up to 2 chosen units and structures get -1 power until end of turn. | Title spelling corrected ('Reverberating Terror'). | FR-TXT-001 |
| **Devilhorn Guerrillas** | PATH Unit | when this unit deploys, if it was deployed to the main theater, deal 2 damage to chosen opposing unit. | When this unit deploys, if it was deployed to the Main Theater, deal 2 damage to chosen opposing unit. | Title spelling corrected ('Devilhorn Guerrillas'). | FR-TXT-001 |
| **All-Out War** | SOULLESS Maneuver | Discard your hand, then: - chosen allied asset gains power... | Discard your hand. Then: - Chosen allied asset gets power equal to the number of cards discarded this way until end of turn. - Scratch X, where X is the number of cards discarded this way. - Regroup all Scratch Company units in your Reserves. | Title hyphenated ('All-Out War'); list templating standardized. | FR-ABL-010 |

---

## 3. Master Review Register (All 52 Flagged Cards)

| Card Title | Faction | Type | Rarity | Designer Review Flag | Standardization Note | Standardized Rules Text | Previous Rules Text |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **[HIGH]** **Animoceps Rager** | ANIMUS | HEADQUARTERS | HQ | Confirm 'triggers only once each turn' limit syntax. | MTG trigger phrasing 'Whenever [event]'; capitalized 'Scratch 1'; MTG-standard 'This ability triggers only once each turn.' | $stdClean | $prevClean |
| **[HIGH]** **Rift** | ANIMUS | HEADQUARTERS | HQ | Review Headquarters self-damage keyword interaction (Unbowed). | Standardized cost format; capitalized 'Scratch 2'; specified 'this Headquarters' instead of 'this asset'. | $stdClean | $prevClean |
| **[HIGH]** **Funeral Fungus** | ANIMUS | MANEUVER | UNCOMMON | Templated exhaust as additional enactment cost rather than spell resolution effect. | Converted exhaust requirement to MTG-aligned additional cost; specified 'readied allied unit', 'chosen unit card', and 'discard pile'. | $stdClean | $prevClean |
| **[HIGH]** **Mutation Swell** | ANIMUS | MANEUVER | RARE | Confirm rounding if odd number of Scratch Company units were killed (currently 'for every 2'). | Capitalized 'Scratch Company'; MTG causal phrase 'killed this way'; specified 'unit card' and 'discard pile'. | $stdClean | $prevClean |
| **[HIGH]** **Reverberating Terror** | CONCORDIA | MANEUVER | UNCOMMON | Title spelling corrected ('Reverberating Terror'). | Corrected title spelling from 'reverbating' to 'Reverberating Terror'; capitalized sentence. | $stdClean | $prevClean |
| **[HIGH]** **Steel Rain** | PATH | MANEUVER | RARE | Confirm if 'chosen asset' is selected once or independently per card discarded (currently 1 chosen target taking 2*N). | MTG 'discarded this way'; 'no cards in hand' instead of 'no cards left in your hand'. | $stdClean | $prevClean |
| **[HIGH]** **Devilhorn Guerrillas** | PATH | UNIT | UNCOMMON | Title spelling corrected ('Devilhorn Guerrillas'). | Corrected title spelling from 'Guerillas' to 'Guerrillas'; capitalized zone 'Main Theater'. | $stdClean | $prevClean |
| **[HIGH]** **All-Out War** | SOULLESS | MANEUVER | MYTHIC | Title hyphenated ('All-Out War'); list templating standardized. | Hyphenated title 'All-Out War'; MTG sequential list formatting; standardized 'gets power', 'Scratch X', and 'Scratch Company'. | $stdClean | $prevClean |
| **[HIGH]** **Macrocannon Barrage** | SOULLESS | MANEUVER | UNCOMMON | Templated as 'Distribute 5 damage among up to 5 chosen units'. | MTG distributed damage standard syntax matching FR-TGT-002 and FR-MSG-003. | $stdClean | $prevClean |
| **Ensnaring Mist** | ANIMUS | MANEUVER | UNCOMMON | Changed 'until the end of your turn' to canonical 'until end of turn'. | Capitalized zone 'Reserves'; standardized duration to 'until end of turn' and stat reduction to 'gets -1 power'. | $stdClean | $prevClean |
| **Forced Mutation** | ANIMUS | MANEUVER | COMMON | Review morph rules text alignment with FR-ENT-103. | Standardized face-down morph syntax matching FR-ENT-103. | $stdClean | $prevClean |
| **Cliff of Portents** | ANIMUS | STRUCTURE | MYTHIC | Added explicit shuffle clause required by FR-DEK-001. | Standardized resource cost ordering 'R1, T1, exhaust:'; added mandatory shuffle clause per FR-DEK-001. | $stdClean | $prevClean |
| **Rhapsody's Hollow** | ANIMUS | STRUCTURE | RARE | Templated structure sacrifice as an activated ability cost. | MTG activated cost formatting; capitalized 'Main Theater' and title. | $stdClean | $prevClean |
| **Craterfall Ancient** | ANIMUS | UNIT | RARE | Clarified attack damage allocation phrasing. | MTG trigger phrasing 'Whenever this unit attacks'; clarified multi-target damage distribution. | $stdClean | $prevClean |
| **Despoiler Worm** | ANIMUS | UNIT | MYTHIC | Self-reference replaced with 'this unit'; reminder text stripped. | Stripped reminder text; MTG self-reference 'this unit' instead of card name; terminal period. | $stdClean | $prevClean |
| **Drifting Concept** | ANIMUS | UNIT | RARE | Harmonized with hand inspection workflow FR-MOD-001. | Capitalized keyword on line 1; MTG Thoughtseize template matching inspection protocol FR-MOD-001. | $stdClean | $prevClean |
| **Flickering Memory** | ANIMUS | UNIT | UNCOMMON | Reminder text stripped. | Standardized keyword 'Ethereal'; stripped informal reminder text '(can't be attacked)'. | $stdClean | $prevClean |
| **Shifting Impossibility** | ANIMUS | UNIT | RARE | Templated as canonical replacement watcher with 'instead'. | MTG replacement effect syntax 'If [event] would [happen], [action] instead' matching FR-REP-001. | $stdClean | $prevClean |
| **Sinuous Growth** | ANIMUS | UNIT | COMMON | Templated unit sacrifice as cost. | Standardized activated cost format; capitalized subtypes; specified 'discard pile'. | $stdClean | $prevClean |
| **Spoil Forest Amalgam** | ANIMUS | UNIT | UNCOMMON | Reminder text stripped. | Capitalized keyword; stripped reminder text. | $stdClean | $prevClean |
| **Spoil Forest Haunt** | ANIMUS | UNIT | COMMON | Replaced sentence with standard keyword 'Deploys Exhausted'. | Standardized to canonical keyword 'Deploys Exhausted' (FR-KWD-002l). | $stdClean | $prevClean |
| **Consuming Earth** | CONCORDIA | MANEUVER | RARE | Term 'requisition value' replaced with canonical 'Requisition yield'. | Changed 'requisition value' to canonical attribute 'Requisition yield' (FR-CRD-101). | $stdClean | $prevClean |
| **Light of the World Soul** | CONCORDIA | MANEUVER | MYTHIC | Clarified card partitioning from top-7 inspection pool. | MTG-standard inspection and partition syntax matching FR-MOD-001 and FR-MSG-005. | $stdClean | $prevClean |
| **Rage's Wake** | CONCORDIA | MANEUVER | COMMON | Clarified 'redeploy' as 'deploy it from your discard pile to your Reserves for free'. | Standardized 'gets +2 power'; clarified delayed trigger and deployment destination zone 'Reserves'. | $stdClean | $prevClean |
| **Stormbolt Ray** | CONCORDIA | STRUCTURE | RARE | Separated deployment trigger from activated ability onto two lines. | Split triggered deployment from activated ability; standardized cost 'T2, exhaust:'. | $stdClean | $prevClean |
| **Furnace Golem** | CONCORDIA | UNIT | COMMON | Reminder text stripped. | Standardized keyword 'Aggressive 2'; stripped reminder text. | $stdClean | $prevClean |
| **Grove Tender** | CONCORDIA | UNIT | COMMON | Replaced sentence with standard keyword 'Deploys Exhausted'. | Standardized to canonical keyword 'Deploys Exhausted' (FR-KWD-002l). | $stdClean | $prevClean |
| **Samsara, Living Tempest** | CONCORDIA | UNIT | MYTHIC | Clarified causality of discarded maneuver condition. | MTG trigger phrasing 'Whenever this unit attacks'; specified 'discard pile' and 'maneuver card'; MTG causality clause 'put into your discard pile this way'. | $stdClean | $prevClean |
| **World Soul Messenger** | CONCORDIA | UNIT | RARE | Added explicit shuffle clause required by FR-DEK-001. | Specified 'Elemental card'; added mandatory shuffle clause per FR-DEK-001. | $stdClean | $prevClean |
| **Encampment** | PATH | HEADQUARTERS | HQ | Review multi-target damage phrasing. | Standardized cost format; clarified 'each of up to 2 chosen non-Scratch Company units'. | $stdClean | $prevClean |
| **Allshan's Battle Trance** | PATH | MANEUVER | MYTHIC | Changed 'until the end of your turn' to canonical 'until end of turn'. | Capitalized 'Main Theater' and 'Stalwart'; standardized 'gets +2 power until end of turn'. | $stdClean | $prevClean |
| **Ambush Tactics** | PATH | MANEUVER | RARE | Modal choice bullet formatting applied. | MTG modal standard bullet list syntax. | $stdClean | $prevClean |
| **Devilhorn Scouting Party** | PATH | MANEUVER | UNCOMMON | Added explicit 'Look at the top 3 cards...' preamble. | MTG search/inspection syntax; specified 'Requisition cost 3 or less'; canonical 'Scratch 1'. | $stdClean | $prevClean |
| **Ancestral Monument** | PATH | STRUCTURE | RARE | Templated ability grant syntax. | Standardized continuous grant syntax; ability word format. | $stdClean | $prevClean |
| **Blind Defile** | PATH | STRUCTURE | COMMON | Refined reference to 'that attacking unit'. | MTG trigger phrasing 'Whenever [event]'; refers to 'that attacking unit'; aligns with Step 2 interrupt FR-CBT-002. | $stdClean | $prevClean |
| **Allshan, Chainbreaker Revenant** | PATH | UNIT | MYTHIC | Standardized ability words. | Standardized ability words for Unbowed and Cornered; standardized 'gets +3 power until end of turn'. | $stdClean | $prevClean |
| **Boneyard Scavenger** | PATH | UNIT | UNCOMMON | Standardized cost ordering and '1T' -> 'T1'. | Standardized cost format 'T1, exhaust, kill this unit:'; specified 'maneuver card' and 'discard pile'. | $stdClean | $prevClean |
| **Brigand Highlander** | PATH | UNIT | COMMON | Standardized to 'Cornered: This unit has +2 power.'. | Standardized ability word 'Cornered:'; continuous stat grant 'has +2 power'. | $stdClean | $prevClean |
| **Chainbreaker Irregulars** | PATH | UNIT | COMMON | Standardized to 'Unbowed: Draw a card...'. | Standardized ability word; 'Draw a card, then discard a card'. | $stdClean | $prevClean |
| **Devilhorn Captain** | PATH | UNIT | UNCOMMON | Standardized to 'Unbowed: Deal 1 damage...'. | Standardized ability word; capitalized zone 'Main Theater'. | $stdClean | $prevClean |
| **Eerie Remnant** | PATH | UNIT | UNCOMMON | Specified 'card' and 'discard pile'. | Keyword 'Deploys Exhausted' on line 1; specified 'card' and 'discard pile'; MTG replacement 'instead'. | $stdClean | $prevClean |
| **Hammerhand Brigands** | PATH | UNIT | COMMON | Standardized to 'Unbowed: Scratch 1.'. | Standardized ability word 'Unbowed:'; canonical keyword 'Scratch 1'. | $stdClean | $prevClean |
| **Macabre Dancers** | PATH | UNIT | RARE | Standardized to 'Unbowed: Heal 2.' and 'Cornered: Resistant 1.'. | Standardized ability words for Unbowed and Cornered. | $stdClean | $prevClean |
| **Rogue Fleshshaper** | PATH | UNIT | COMMON | Modal choice list formatting applied. | MTG modal list formatting; capitalized 'non-Scratch'. | $stdClean | $prevClean |
| **Ruin Partisans** | PATH | UNIT | UNCOMMON | Standardized cost ordering. | Standardized cost format; specified 'unit card' and 'discard pile'. | $stdClean | $prevClean |
| **Solarheart Monk** | PATH | UNIT | COMMON | Clarified trigger on 'this card' rather than 'this unit' while entering discard. | Specified 'card' and 'discard pile' (triggers from non-battlefield zone per FR-ZON-003). | $stdClean | $prevClean |
| **Mage Seeker's Inquisition** | SOULLESS | MANEUVER | UNCOMMON | Harmonized with hand inspection workflow FR-MOD-001. | MTG Thoughtseize template matching inspection protocol FR-MOD-001. | $stdClean | $prevClean |
| **Mage Hunter Brigade** | SOULLESS | UNIT | COMMON | Reminder text stripped. | Standardized keyword; stripped reminder text. | $stdClean | $prevClean |
| **Manufactus Conscripts** | SOULLESS | UNIT | COMMON | Replaced sentence with standard keyword 'Deploys Exhausted'. | Standardized to canonical keyword 'Deploys Exhausted' (FR-KWD-002l). | $stdClean | $prevClean |
| **Mobile Armory** | SOULLESS | UNIT | UNCOMMON | Changed 'this turn' to 'until end of turn'. | Standardized activated cost; 'gets +1 power until end of turn' instead of 'this turn'. | $stdClean | $prevClean |
| **NO-AH-5** | SOULLESS | UNIT | MYTHIC | Harmonized with canonical benchmark in Master FRD Section 8.2. | Benchmark text from Master FRD Section 8.2; standardized tucking terminology (FR-ENT-102). | $stdClean | $prevClean |
| **Piston Fist Prototype** | SOULLESS | UNIT | MYTHIC | Harmonized with Master FRD benchmark. | Canonical text from Master FRD Section 8.2 benchmark. | $stdClean | $prevClean |

