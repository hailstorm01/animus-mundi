# Full card standardizer and audit logger
param(
    [string]$RepoRoot = (Resolve-Path "$PSScriptRoot\..\..").Path
)

$cardsDir = Join-Path $RepoRoot "data\cards"
. (Join-Path $PSScriptRoot "csv_helper.ps1")

# Master lookup table of text transformations and review notes
$textRules = @{
    # =========================================================================
    # HEADQUARTERS (units-and-structures-HQ.csv)
    # =========================================================================
    "rift" = @{
        NewName = "Rift"
        NewText = "Discard a card, exhaust: Scratch 2. Deal 5 damage to this Headquarters."
        Note = "Standardized cost format; capitalized 'Scratch 2'; specified 'this Headquarters' instead of 'this asset'."
        Flag = "Review Headquarters self-damage keyword interaction (Unbowed)."
    }
    "animoceps rager" = @{
        NewName = "Animoceps Rager"
        NewText = "Whenever this Headquarters is dealt damage, Scratch 1. This ability triggers only once each turn."
        Note = "MTG trigger phrasing 'Whenever [event]'; capitalized 'Scratch 1'; MTG-standard 'This ability triggers only once each turn.'"
        Flag = "Confirm 'triggers only once each turn' limit syntax."
    }
    "enclave" = @{
        NewName = "Enclave"
        NewText = "Discard a card, exhaust: Regroup up to 2 chosen allied units."
        Note = "Standardized cost format; capitalized Regroup; added 'chosen' targeting qualifier."
        Flag = "None."
    }
    "tidal leviathan" = @{
        NewName = "Tidal Leviathan"
        NewText = "Whenever this Headquarters deals damage, chosen allied unit or structure heals 1."
        Note = "MTG trigger phrasing 'Whenever [event]'; capitalized 'Headquarters'."
        Flag = "None."
    }
    "encampment" = @{
        NewName = "Encampment"
        NewText = "Discard a card, exhaust: Deal 1 damage to each of up to 2 chosen non-Scratch Company units."
        Note = "Standardized cost format; clarified 'each of up to 2 chosen non-Scratch Company units'."
        Flag = "Review multi-target damage phrasing."
    }
    "eternal watchman" = @{
        NewName = "Eternal Watchman"
        NewText = "Whenever this Headquarters attacks and kills an opposing unit or structure, return chosen unit card from your discard pile to your hand."
        Note = "MTG trigger phrasing 'Whenever [event]'; specified 'unit card' and 'discard pile'."
        Flag = "None."
    }
    "command barge" = @{
        NewName = "Command Barge"
        NewText = "Discard a card, exhaust: Draw 2 cards, then discard a card."
        Note = "Standardized cost format; MTG sequential phrasing 'then discard a card'."
        Flag = "None."
    }
    "colossus platform" = @{
        NewName = "Colossus Platform"
        NewText = "Whenever this Headquarters attacks, chosen allied unit gains 1 armor."
        Note = "MTG trigger phrasing 'Whenever [event]'; capitalized 'Headquarters'."
        Flag = "None."
    }

    # =========================================================================
    # COMMON (units-and-structures-common.csv)
    # =========================================================================
    "Scratch Company" = @{
        NewName = "Scratch Company"
        NewText = "--"
        Note = "No rules text; token baseline."
        Flag = "None."
    }

    # =========================================================================
    # MANEUVERS - ANIMUS (maneuvers-animus.csv)
    # =========================================================================
    "disorienting pulse" = @{
        NewName = "Disorienting Pulse"
        NewText = "Return chosen opposing unit in the Main Theater to its owner's hand."
        Note = "Capitalized sentence; capitalized zone 'Main Theater'; terminal period."
        Flag = "None."
    }
    "animoceps spores" = @{
        NewName = "Animoceps Spores"
        NewText = "Scratch 2, then return chosen allied Scratch Company unit to its owner's hand."
        Note = "Standardized 'scratch twice' to canonical 'Scratch 2'; added 'chosen' targeting qualifier."
        Flag = "None."
    }
    "deafening aria" = @{
        NewName = "Deafening Aria"
        NewText = "Deal 1 damage to all opposing units in the Main Theater."
        Note = "Capitalized sentence; capitalized zone 'Main Theater'; terminal period."
        Flag = "None."
    }
    "feed the despoiler" = @{
        NewName = "Feed the Despoiler"
        NewText = "As an additional cost to enact this maneuver, kill an allied unit.`nDraw 2 cards."
        Note = "MTG additional cost formatting on dedicated line; terminal periods."
        Flag = "None."
    }
    "conversion spike" = @{
        NewName = "Conversion Spike"
        NewText = "Deal 1 damage to chosen opposing unit, and chosen allied unit or structure heals 1."
        Note = "Capitalized sentence; comma separation; terminal period."
        Flag = "None."
    }
    "ensnaring mist" = @{
        NewName = "Ensnaring Mist"
        NewText = "Move chosen unit to its controller's Reserves.`nChosen unit gets -1 power until end of turn."
        Note = "Capitalized zone 'Reserves'; standardized duration to 'until end of turn' and stat reduction to 'gets -1 power'."
        Flag = "Changed 'until the end of your turn' to canonical 'until end of turn'."
    }
    "mutation swell" = @{
        NewName = "Mutation Swell"
        NewText = "Kill up to 4 allied Scratch Company units. For every 2 units killed this way, deploy a unit card from your discard pile for free."
        Note = "Capitalized 'Scratch Company'; MTG causal phrase 'killed this way'; specified 'unit card' and 'discard pile'."
        Flag = "Confirm rounding if odd number of Scratch Company units were killed (currently 'for every 2')."
    }
    "rhapsody's embrace" = @{
        NewName = "Rhapsody's Embrace"
        NewText = "As an additional cost to enact this maneuver, kill an allied unit.`nKill chosen opposing unit."
        Note = "MTG additional cost formatting on dedicated line; terminal periods."
        Flag = "None."
    }
    "despoiler's emergence" = @{
        NewName = "Despoiler's Emergence"
        NewText = "Deal 4 damage to all opposing assets in the Main Theater.`nDeal 2 damage to all opposing non-Scratch assets in opponents' Reserves."
        Note = "Capitalized zones 'Main Theater' and 'Reserves'; standardized 'non-Scratch assets'."
        Flag = "None."
    }
    "rapturous sermon" = @{
        NewName = "Rapturous Sermon"
        NewText = "Move all allied Scratch Company units to the Main Theater. Allied Scratch Company units get +2 power until end of turn."
        Note = "Capitalized 'Scratch Company' and 'Main Theater'; standardized 'get +2 power until end of turn'."
        Flag = "None."
    }
    "forced mutation" = @{
        NewName = "Forced Mutation"
        NewText = "Turn chosen opposing unit face down. It becomes a Scratch Company unit."
        Note = "Standardized face-down morph syntax matching FR-ENT-103."
        Flag = "Review morph rules text alignment with FR-ENT-103."
    }
    "funeral fungus" = @{
        NewName = "Funeral Fungus"
        NewText = "As an additional cost to enact this maneuver, exhaust a readied allied unit.`nReturn chosen unit card from your discard pile to your hand."
        Note = "Converted exhaust requirement to MTG-aligned additional cost; specified 'readied allied unit', 'chosen unit card', and 'discard pile'."
        Flag = "Templated exhaust as additional enactment cost rather than spell resolution effect."
    }

    # =========================================================================
    # MANEUVERS - CONCORDIA (maneuvers-concordia.csv)
    # =========================================================================
    "reverbating terror" = @{
        NewName = "Reverberating Terror"
        NewText = "Up to 2 chosen units and structures get -1 power until end of turn."
        Note = "Corrected title spelling from 'reverbating' to 'Reverberating Terror'; capitalized sentence."
        Flag = "Title spelling corrected ('Reverberating Terror')."
    }
    "natural attunement" = @{
        NewName = "Natural Attunement"
        NewText = "Generate 2 Requisition. Draw a card."
        Note = "Capitalized sentence and resource name 'Requisition'; terminal period."
        Flag = "None."
    }
    "growing legion" = @{
        NewName = "Growing Legion"
        NewText = "Scratch 1."
        Note = "Capitalized canonical keyword 'Scratch 1'."
        Flag = "None."
    }
    "leyline regrowth" = @{
        NewName = "Leyline Regrowth"
        NewText = "Return up to one chosen unit card from your discard pile to your hand. Draw a card."
        Note = "Added 'chosen' targeting qualifier; specified 'unit card' and 'discard pile'."
        Flag = "None."
    }
    "forgemaster's resilience" = @{
        NewName = "Forgemaster's Resilience"
        NewText = "Chosen unit gains Stalwart until end of turn."
        Note = "Capitalized keyword 'Stalwart'; terminal period."
        Flag = "None."
    }
    "frostfire breath" = @{
        NewName = "Frostfire Breath"
        NewText = "Deal 2 damage to chosen exhausted non-Scratch asset. If that asset is killed, draw a card."
        Note = "Capitalized 'non-Scratch'; combined into clear conditional sentence."
        Flag = "None."
    }
    "soul whip" = @{
        NewName = "Soul Whip"
        NewText = "Chosen allied unit deals damage equal to its power to chosen opposing unit."
        Note = "Capitalized sentence; MTG bite template; terminal period."
        Flag = "None."
    }
    "rage's wake" = @{
        NewName = "Rage's Wake"
        NewText = "Chosen unit gets +2 power until end of turn. When that unit is killed this turn, deploy it from your discard pile to your Reserves for free."
        Note = "Standardized 'gets +2 power'; clarified delayed trigger and deployment destination zone 'Reserves'."
        Flag = "Clarified 'redeploy' as 'deploy it from your discard pile to your Reserves for free'."
    }
    "consuming earth" = @{
        NewName = "Consuming Earth"
        NewText = "Kill chosen opposing unit with power less than or equal to the total Requisition yield of all structures you control."
        Note = "Changed 'requisition value' to canonical attribute 'Requisition yield' (FR-CRD-101)."
        Flag = "Term 'requisition value' replaced with canonical 'Requisition yield'."
    }
    "Awaken Nightmare" = @{
        NewName = "Awaken Nightmare"
        NewText = "Move chosen opposing unit from the Main Theater to its owner's Reserves. All opposing units get -1 power until end of turn."
        Note = "Capitalized zones 'Main Theater' and 'Reserves'; terminal period."
        Flag = "None."
    }
    "seismic instability" = @{
        NewName = "Seismic Instability"
        NewText = "Deal 4 damage to all opposing structures."
        Note = "Capitalized sentence; terminal period."
        Flag = "None."
    }
    "light of the world soul" = @{
        NewName = "Light of the World Soul"
        NewText = "Look at the top 7 cards of your deck. You may deploy up to 2 unit cards from among them for free, and put up to 1 of those cards into your hand. Put the rest on the bottom of your deck in any order."
        Note = "MTG-standard inspection and partition syntax matching FR-MOD-001 and FR-MSG-005."
        Flag = "Clarified card partitioning from top-7 inspection pool."
    }

    # =========================================================================
    # MANEUVERS - PATH (maneuvers-path.csv)
    # =========================================================================
    "fortify" = @{
        NewName = "Fortify"
        NewText = "Chosen allied unit gains 3 armor."
        Note = "Capitalized sentence; terminal period."
        Flag = "None."
    }
    "chainbreaker's tenacity" = @{
        NewName = "Chainbreaker's Tenacity"
        NewText = "Chosen allied unit gets +1 power until end of turn and gains 2 armor."
        Note = "Standardized 'gets +1 power until end of turn'."
        Flag = "None."
    }
    "tactical withdrawal" = @{
        NewName = "Tactical Withdrawal"
        NewText = "Return up to one chosen allied unit to its owner's hand. Draw a card."
        Note = "Added 'chosen'; MTG 'Draw a card' instead of 'draw 1 card'."
        Flag = "None."
    }
    "blindside" = @{
        NewName = "Blindside"
        NewText = "Chosen opposing unit gets -3 power until the end of your opponent's turn."
        Note = "Capitalized sentence; benchmark phrasing from Master FRD Section 6.2 (FR-LAY-003)."
        Flag = "None."
    }
    "Reaver's rush" = @{
        NewName = "Reaver's Rush"
        NewText = "Deploy chosen unit card with power 3 or less from your discard pile for free."
        Note = "Capitalized sentence and name; specified 'unit card' and 'discard pile'."
        Flag = "None."
    }
    "devilhorn scouting party" = @{
        NewName = "Devilhorn Scouting Party"
        NewText = "Look at the top 3 cards of your deck. You may deploy a unit card with Requisition cost 3 or less from among them for free. Put the rest on the bottom of your deck in any order. Then, Scratch 1."
        Note = "MTG search/inspection syntax; specified 'Requisition cost 3 or less'; canonical 'Scratch 1'."
        Flag = "Added explicit 'Look at the top 3 cards...' preamble."
    }
    "improvised orders" = @{
        NewName = "Improvised Orders"
        NewText = "Draw 3 cards, then discard a card."
        Note = "MTG 'discard a card' instead of 'discard 1 card'; terminal period."
        Flag = "None."
    }
    "steel rain" = @{
        NewName = "Steel Rain"
        NewText = "Discard up to 3 cards. For each card discarded this way, deal 2 damage to chosen asset.`nThen, if you have no cards in hand, draw 2 cards."
        Note = "MTG 'discarded this way'; 'no cards in hand' instead of 'no cards left in your hand'."
        Flag = "Confirm if 'chosen asset' is selected once or independently per card discarded (currently 1 chosen target taking 2*N)."
    }
    "devilhorn sanction" = @{
        NewName = "Devilhorn Sanction"
        NewText = "Kill chosen opposing unit."
        Note = "Capitalized sentence; terminal period."
        Flag = "None."
    }
    "Pincer Tactics" = @{
        NewName = "Pincer Tactics"
        NewText = "Deal 1 damage to each of up to 2 chosen allied units. Deal 2 damage to chosen non-Scratch asset."
        Note = "Clarified distributed damage 'each of up to 2 chosen allied units'; capitalized 'non-Scratch'."
        Flag = "None."
    }
    "allshan's battle trance" = @{
        NewName = "Allshan's Battle Trance"
        NewText = "Deploy chosen unit card from your discard pile to the Main Theater. It gains Stalwart and gets +2 power until end of turn."
        Note = "Capitalized 'Main Theater' and 'Stalwart'; standardized 'gets +2 power until end of turn'."
        Flag = "Changed 'until the end of your turn' to canonical 'until end of turn'."
    }
    "ambush tactics" = @{
        NewName = "Ambush Tactics"
        NewText = "Choose two:`n- Kill chosen unit.`n- Return chosen unit to its owner's hand.`n- Chosen unit deals damage equal to its power to chosen unit.`n- Put chosen unit on top of its owner's deck."
        Note = "MTG modal standard bullet list syntax."
        Flag = "Modal choice bullet formatting applied."
    }
    "poisoned rations" = @{
        NewName = "Poisoned Rations"
        NewText = "Chosen unit gets -3 max health until end of turn."
        Note = "Standardized 'until end of turn' matching FR-LAY-002."
        Flag = "None."
    }

    # =========================================================================
    # MANEUVERS - SOULLESS (maneuvers-soulless.csv)
    # =========================================================================
    "piston fist" = @{
        NewName = "Piston Fist"
        NewText = "Deal 1 damage to chosen unit, and chosen allied unit gets +1 power until end of turn."
        Note = "Standardized 'gets +1 power until end of turn'; comma separation."
        Flag = "None."
    }
    "bring it down" = @{
        NewName = "Bring It Down"
        NewText = "Deal 2 damage to chosen non-Scratch asset."
        Note = "Capitalized 'non-Scratch'; terminal period."
        Flag = "None."
    }
    "visionary's insight" = @{
        NewName = "Visionary's Insight"
        NewText = "Regroup up to one chosen exhausted allied asset."
        Note = "Capitalized keyword 'Regroup'."
        Flag = "None."
    }
    "mage seeker's inquisition" = @{
        NewName = "Mage Seeker's Inquisition"
        NewText = "Chosen opponent reveals their hand. You choose a unit card from it. That player discards that card."
        Note = "MTG Thoughtseize template matching inspection protocol FR-MOD-001."
        Flag = "Harmonized with hand inspection workflow FR-MOD-001."
    }
    "railway reinforcements" = @{
        NewName = "Railway Reinforcements"
        NewText = "Move up to 3 chosen allied units from your Reserves to the Main Theater."
        Note = "Capitalized zones 'Reserves' and 'Main Theater'."
        Flag = "None."
    }
    "redline infusion" = @{
        NewName = "Redline Infusion"
        NewText = "Chosen allied unit gets +2 power and gains Berserk until end of turn. At the end of your turn, kill that unit."
        Note = "Capitalized keyword 'Berserk'; matches delayed trigger Sub-Phase 4a specification in FR-TRN-001."
        Flag = "None."
    }
    "dolus infusion" = @{
        NewName = "Dolus Infusion"
        NewText = "Draw 2 cards, and your Headquarters heals 2."
        Note = "Capitalized 'Headquarters'; comma separation."
        Flag = "None."
    }
    "inspired salvage" = @{
        NewName = "Inspired Salvage"
        NewText = "Return up to 2 chosen Mechanized cards from your discard pile to your hand."
        Note = "Added 'chosen'; capitalized subtype 'Mechanized'; specified 'discard pile'."
        Flag = "None."
    }
    "fire for effect" = @{
        NewName = "Fire for Effect"
        NewText = "Opposing assets in the Main Theater gain Resistant -1 until end of turn.`nDeal 1 damage to chosen asset in the Main Theater."
        Note = "Capitalized 'Main Theater' and keyword 'Resistant -1'; standardized 'until end of turn'."
        Flag = "None."
    }
    "macrocannon barrage" = @{
        NewName = "Macrocannon Barrage"
        NewText = "Distribute 5 damage among up to 5 chosen units."
        Note = "MTG distributed damage standard syntax matching FR-TGT-002 and FR-MSG-003."
        Flag = "Templated as 'Distribute 5 damage among up to 5 chosen units'."
    }
    "Armored Charge" = @{
        NewName = "Armored Charge"
        NewText = "Regroup chosen allied Mechanized unit. It heals 5 and gains 2 armor."
        Note = "Capitalized subtype 'Mechanized'; terminal period."
        Flag = "None."
    }
    "All out War" = @{
        NewName = "All-Out War"
        NewText = "Discard your hand. Then:`n- Chosen allied asset gets power equal to the number of cards discarded this way until end of turn.`n- Scratch X, where X is the number of cards discarded this way.`n- Regroup all Scratch Company units in your Reserves."
        Note = "Hyphenated title 'All-Out War'; MTG sequential list formatting; standardized 'gets power', 'Scratch X', and 'Scratch Company'."
        Flag = "Title hyphenated ('All-Out War'); list templating standardized."
    }
    "March of progress" = @{
        NewName = "March of Progress"
        NewText = "Up to 2 chosen units gain the Mechanized subtype until end of turn. Allied Mechanized units gain 2 armor and get +2 power until end of turn."
        Note = "Specified 'Mechanized subtype'; standardized 'get +2 power until end of turn'."
        Flag = "None."
    }

    # =========================================================================
    # UNITS & STRUCTURES - ANIMUS (units-and-structures-animus.csv)
    # =========================================================================
    "portent scrivener" = @{
        NewName = "Portent Scrivener"
        NewText = "When this unit is killed on your turn, draw a card."
        Note = "Capitalized sentence; MTG 'draw a card'; terminal period."
        Flag = "None."
    }
    "shifting impossibility" = @{
        NewName = "Shifting Impossibility"
        NewText = "If an allied Scratch Company unit would be killed, return it to its owner's hand and deal 1 damage to this unit instead."
        Note = "MTG replacement effect syntax 'If [event] would [happen], [action] instead' matching FR-REP-001."
        Flag = "Templated as canonical replacement watcher with 'instead'."
    }
    "Spoil Forest Haunt" = @{
        NewName = "Spoil Forest Haunt"
        NewText = "Deploys Exhausted"
        Note = "Standardized to canonical keyword 'Deploys Exhausted' (FR-KWD-002l)."
        Flag = "Replaced sentence with standard keyword 'Deploys Exhausted'."
    }
    "Rhapsody's hollow" = @{
        NewName = "Rhapsody's Hollow"
        NewText = "T3, exhaust, kill this structure: Gain control of chosen opposing unit in the Main Theater."
        Note = "MTG activated cost formatting; capitalized 'Main Theater' and title."
        Flag = "Templated structure sacrifice as an activated ability cost."
    }
    "flickering memory" = @{
        NewName = "Flickering Memory"
        NewText = "Ethereal"
        Note = "Standardized keyword 'Ethereal'; stripped informal reminder text '(can't be attacked)'."
        Flag = "Reminder text stripped."
    }
    "Animoceps Stalk" = @{
        NewName = "Animoceps Stalk"
        NewText = "Other allied assets in the same zone have +1 power."
        Note = "Terminal period added."
        Flag = "None."
    }
    "Sigiled Stoneform" = @{
        NewName = "Sigiled Stoneform"
        NewText = "Stalwart"
        Note = "Capitalized keyword."
        Flag = "None."
    }
    "Sinuous Growth" = @{
        NewName = "Sinuous Growth"
        NewText = "T1, kill this unit: Return chosen Fungus or Plant card from your discard pile to your hand."
        Note = "Standardized activated cost format; capitalized subtypes; specified 'discard pile'."
        Flag = "Templated unit sacrifice as cost."
    }
    "Cliff of Portents" = @{
        NewName = "Cliff of Portents"
        NewText = "R1, T1, exhaust: Search your deck for a card, put it into your hand, then shuffle your deck."
        Note = "Standardized resource cost ordering 'R1, T1, exhaust:'; added mandatory shuffle clause per FR-DEK-001."
        Flag = "Added explicit shuffle clause required by FR-DEK-001."
    }
    "despoiler spawn" = @{
        NewName = "Despoiler Spawn"
        NewText = "Resistant 1"
        Note = "Capitalized keyword."
        Flag = "None."
    }
    "Spoil Forest Creeper" = @{
        NewName = "Spoil Forest Creeper"
        NewText = "When this unit deploys, Scratch 1."
        Note = "Capitalized keyword 'Scratch 1'."
        Flag = "None."
    }
    "stormfront of ideals" = @{
        NewName = "Stormfront of Ideals"
        NewText = "When this unit is killed, put it on the bottom of its owner's deck."
        Note = "Capitalized sentence; terminal period."
        Flag = "None."
    }
    "Unceasing Tide" = @{
        NewName = "Unceasing Tide"
        NewText = "When this unit is killed, return it to its owner's hand."
        Note = "Terminal period added."
        Flag = "None."
    }
    "Veil Render" = @{
        NewName = "Veil Render"
        NewText = "You may deploy this card from your discard pile by killing an allied asset in addition to paying its deployment costs."
        Note = "Specified 'discard pile' and 'deployment costs' matching FR-TRN-003a."
        Flag = "None."
    }
    "infested apostle" = @{
        NewName = "Infested Apostle"
        NewText = "When this unit deploys, you may kill an allied unit. If you do, deal 2 damage to chosen opposing unit."
        Note = "Capitalized sentence; terminal period."
        Flag = "None."
    }
    "drifting concept" = @{
        NewName = "Drifting Concept"
        NewText = "Ethereal`nWhen this unit deploys, chosen opponent reveals their hand. You choose a card from it. That player discards that card."
        Note = "Capitalized keyword on line 1; MTG Thoughtseize template matching inspection protocol FR-MOD-001."
        Flag = "Harmonized with hand inspection workflow FR-MOD-001."
    }
    "hulking incongruity" = @{
        NewName = "Hulking Incongruity"
        NewText = "Allied Horror units in the same zone have Resistant 1."
        Note = "Terminal period added."
        Flag = "None."
    }
    "Animoceps Spore Cloud" = @{
        NewName = "Animoceps Spore Cloud"
        NewText = "Redirect all damage that would be dealt to allied units in the same zone to this structure."
        Note = "Harmonized with benchmark text in Master FRD Section 8.2."
        Flag = "None."
    }
    "Spoil Forest Amalgam" = @{
        NewName = "Spoil Forest Amalgam"
        NewText = "Stalwart"
        Note = "Capitalized keyword; stripped reminder text."
        Flag = "Reminder text stripped."
    }
    "Animoceps Node" = @{
        NewName = "Animoceps Node"
        NewText = "Allied units in the same zone have +1 power."
        Note = "Capitalized sentence; terminal period."
        Flag = "None."
    }
    "Spoil Forest Colossus" = @{
        NewName = "Spoil Forest Colossus"
        NewText = "Whenever this unit deploys or attacks, chosen asset heals 2."
        Note = "MTG trigger phrasing 'Whenever [event]'."
        Flag = "None."
    }
    "melodic pummeler" = @{
        NewName = "Melodic Pummeler"
        NewText = "Allied Horror units in the same zone have +1 power."
        Note = "Capitalized sentence; terminal period."
        Flag = "None."
    }
    "fungal evangelist" = @{
        NewName = "Fungal Evangelist"
        NewText = "When this unit deploys or is killed, Scratch 2."
        Note = "Capitalized keyword 'Scratch 2'."
        Flag = "None."
    }
    "Despoiler Worm" = @{
        NewName = "Despoiler Worm"
        NewText = "Resistant 2`nWhen this unit is killed, all opponents draw 2 cards."
        Note = "Stripped reminder text; MTG self-reference 'this unit' instead of card name; terminal period."
        Flag = "Self-reference replaced with 'this unit'; reminder text stripped."
    }
    "Titanic Spore Carrier" = @{
        NewName = "Titanic Spore Carrier"
        NewText = "When this unit is killed, Scratch 1 and draw a card."
        Note = "Capitalized 'Scratch 1'; MTG 'draw a card' instead of 'draw one card'."
        Flag = "None."
    }
    "craterfall ancient" = @{
        NewName = "Craterfall Ancient"
        NewText = "Whenever this unit attacks, you may deal damage equal to its power to each of up to 2 other chosen assets.`nWhen this unit is killed, deal 4 damage to all assets."
        Note = "MTG trigger phrasing 'Whenever this unit attacks'; clarified multi-target damage distribution."
        Flag = "Clarified attack damage allocation phrasing."
    }
    "despoiler larva" = @{
        NewName = "Despoiler Larva"
        NewText = "Whenever this unit attacks, Scratch 1."
        Note = "MTG trigger phrasing 'Whenever this unit attacks'; capitalized keyword 'Scratch 1'."
        Flag = "None."
    }
    "Animoceps Shambler" = @{
        NewName = "Animoceps Shambler"
        NewText = "Whenever this unit exhausts, chosen allied asset in the Main Theater heals 1."
        Note = "MTG trigger phrasing 'Whenever this unit exhausts'; capitalized zone 'Main Theater'."
        Flag = "None."
    }

    # =========================================================================
    # UNITS & STRUCTURES - CONCORDIA (units-and-structures-concordia.csv)
    # =========================================================================
    "Foundry Artisan" = @{
        NewName = "Foundry Artisan"
        NewText = "Stalwart"
        Note = "Standard keyword."
        Flag = "None."
    }
    "Grove Tender" = @{
        NewName = "Grove Tender"
        NewText = "Deploys Exhausted"
        Note = "Standardized to canonical keyword 'Deploys Exhausted' (FR-KWD-002l)."
        Flag = "Replaced sentence with standard keyword 'Deploys Exhausted'."
    }
    "sandstorm Lightbringer" = @{
        NewName = "Sandstorm Lightbringer"
        NewText = "Exhaust: Move chosen allied unit to the Main Theater or Reserves."
        Note = "Standardized activated cost format; added 'chosen'; capitalized zones."
        Flag = "None."
    }
    "Daggerspine Rangers" = @{
        NewName = "Daggerspine Rangers"
        NewText = "--"
        Note = "Vanilla unit."
        Flag = "None."
    }
    "Ego Spark" = @{
        NewName = "Ego Spark"
        NewText = "Guardian"
        Note = "Capitalized keyword."
        Flag = "None."
    }
    "tempest surge" = @{
        NewName = "Tempest Surge"
        NewText = "T2: This unit gets +2 power until end of turn."
        Note = "Standardized activated cost 'T2:'; standardized 'gets +2 power until end of turn'."
        Flag = "None."
    }
    "hourglass diviner" = @{
        NewName = "Hourglass Diviner"
        NewText = "When this unit deploys, look at the top 2 cards of your deck. Put one into your hand and the other on the bottom of your deck."
        Note = "Capitalized name and terminal period."
        Flag = "None."
    }
    "Screaming Palisade" = @{
        NewName = "Screaming Palisade"
        NewText = "T3, exhaust: Each opponent discards a card."
        Note = "Standardized activated cost format 'T3, exhaust:'."
        Flag = "None."
    }
    "Forest Heavies" = @{
        NewName = "Forest Heavies"
        NewText = "--"
        Note = "Vanilla unit."
        Flag = "None."
    }
    "Icewind Regent" = @{
        NewName = "Icewind Regent"
        NewText = "Exhaust: Return chosen unit to its owner's hand."
        Note = "Standardized activated cost format; terminal period."
        Flag = "None."
    }
    "stormbolt ray" = @{
        NewName = "Stormbolt Ray"
        NewText = "When this structure deploys, deal 3 damage to chosen asset.`nT2, exhaust: Return this structure to its owner's hand."
        Note = "Split triggered deployment from activated ability; standardized cost 'T2, exhaust:'."
        Flag = "Separated deployment trigger from activated ability onto two lines."
    }
    "frontier foundry" = @{
        NewName = "Frontier Foundry"
        NewText = "Stalwart"
        Note = "Capitalized keyword."
        Flag = "None."
    }
    "Tempest Runner" = @{
        NewName = "Tempest Runner"
        NewText = "Exhaust: Regroup another chosen allied unit in the same zone."
        Note = "Standardized activated cost; capitalized keyword 'Regroup'."
        Flag = "None."
    }
    "Old Growth Elders" = @{
        NewName = "Old Growth Elders"
        NewText = "--"
        Note = "Vanilla unit."
        Flag = "None."
    }
    "will breaker" = @{
        NewName = "Will Breaker"
        NewText = "When this unit deploys, kill chosen opposing unit with Requisition cost 3 or less."
        Note = "Capitalized 'Requisition cost'; terminal period."
        Flag = "None."
    }
    "cathedral of self" = @{
        NewName = "Cathedral of Self"
        NewText = "T1, exhaust: Deal damage equal to this structure's current Health to chosen unit."
        Note = "Standardized activated cost 'T1, exhaust:'; capitalized 'Health'."
        Flag = "None."
    }
    "Crystalborn Rampager" = @{
        NewName = "Crystalborn Rampager"
        NewText = "Resistant 1"
        Note = "Standard keyword."
        Flag = "None."
    }
    "Furnace Golem" = @{
        NewName = "Furnace Golem"
        NewText = "Aggressive 2"
        Note = "Standardized keyword 'Aggressive 2'; stripped reminder text."
        Flag = "Reminder text stripped."
    }
    "echoing nightmare" = @{
        NewName = "Echoing Nightmare"
        NewText = "Ethereal"
        Note = "Capitalized keyword."
        Flag = "None."
    }
    "Tempest City Ray" = @{
        NewName = "Tempest City Ray"
        NewText = "T2, exhaust: Move chosen opposing unit to Reserves."
        Note = "Standardized activated cost 'T2, exhaust:'; capitalized zone 'Reserves'."
        Flag = "None."
    }
    "Samsara, Living Tempest" = @{
        NewName = "Samsara, Living Tempest"
        NewText = "Whenever this unit attacks, look at the top 2 cards of your deck. Put one into your hand and the other into your discard pile. If the card put into your discard pile this way was a maneuver card, deal 2 damage to chosen asset."
        Note = "MTG trigger phrasing 'Whenever this unit attacks'; specified 'discard pile' and 'maneuver card'; MTG causality clause 'put into your discard pile this way'."
        Flag = "Clarified causality of discarded maneuver condition."
    }
    "stardust colossus" = @{
        NewName = "Stardust Colossus"
        NewText = "Whenever this unit deploys or attacks, return up to one chosen card from your discard pile to your hand."
        Note = "MTG trigger phrasing 'Whenever [event]'; added 'chosen'; specified 'discard pile'."
        Flag = "None."
    }
    "gemstone titan" = @{
        NewName = "Gemstone Titan"
        NewText = "When this unit deploys, kill chosen structure with current Health less than this unit's power."
        Note = "Capitalized 'Health'; terminal period."
        Flag = "None."
    }
    "Dread Manifestation" = @{
        NewName = "Dread Manifestation"
        NewText = "Whenever this unit kills another unit in combat, you may move chosen unit in the Main Theater to its controller's Reserves.`nOpposing units in the same zone have -2 power."
        Note = "MTG trigger phrasing 'Whenever [event]'; capitalized zones 'Main Theater' and 'Reserves'; continuous aura on line 2."
        Flag = "None."
    }
    "Swirling Dustdevil" = @{
        NewName = "Swirling Dustdevil"
        NewText = "This unit costs R1 less to deploy for each maneuver card in your discard pile."
        Note = "Capitalized sentence; specified 'maneuver card' and 'discard pile'."
        Flag = "None."
    }
    "world soul messenger" = @{
        NewName = "World Soul Messenger"
        NewText = "When this unit deploys, you may search your deck for an Elemental card and deploy it for free, then shuffle your deck."
        Note = "Specified 'Elemental card'; added mandatory shuffle clause per FR-DEK-001."
        Flag = "Added explicit shuffle clause required by FR-DEK-001."
    }
    "flickering ember" = @{
        NewName = "Flickering Ember"
        NewText = "T4: Allied assets in the same zone get +1 power until end of turn."
        Note = "Standardized activated cost 'T4:'; standardized 'get +1 power until end of turn'."
        Flag = "None."
    }

    # =========================================================================
    # UNITS & STRUCTURES - PATH (units-and-structures-path.csv)
    # =========================================================================
    "Chainbreaker Sneak" = @{
        NewName = "Chainbreaker Sneak"
        NewText = "Versatile"
        Note = "Capitalized keyword."
        Flag = "None."
    }
    "Hammerhand Brigands" = @{
        NewName = "Hammerhand Brigands"
        NewText = "Unbowed: Scratch 1."
        Note = "Standardized ability word 'Unbowed:'; canonical keyword 'Scratch 1'."
        Flag = "Standardized to 'Unbowed: Scratch 1.'."
    }
    "Eerie Remnant" = @{
        NewName = "Eerie Remnant"
        NewText = "Deploys Exhausted`nWhen this card is put into your discard pile from your hand, you may deploy it for free instead."
        Note = "Keyword 'Deploys Exhausted' on line 1; specified 'card' and 'discard pile'; MTG replacement 'instead'."
        Flag = "Specified 'card' and 'discard pile'."
    }
    "Brigand Highlander" = @{
        NewName = "Brigand Highlander"
        NewText = "Cornered: This unit has +2 power."
        Note = "Standardized ability word 'Cornered:'; continuous stat grant 'has +2 power'."
        Flag = "Standardized to 'Cornered: This unit has +2 power.'."
    }
    "blind defile" = @{
        NewName = "Blind Defile"
        NewText = "Whenever an opposing unit attacks, you may kill this structure and that attacking unit."
        Note = "MTG trigger phrasing 'Whenever [event]'; refers to 'that attacking unit'; aligns with Step 2 interrupt FR-CBT-002."
        Flag = "Refined reference to 'that attacking unit'."
    }
    "rogue fleshshaper" = @{
        NewName = "Rogue Fleshshaper"
        NewText = "When this unit deploys, choose one:`n- Chosen asset heals 1.`n- Deal 1 damage to chosen non-Scratch asset."
        Note = "MTG modal list formatting; capitalized 'non-Scratch'."
        Flag = "Modal choice list formatting applied."
    }
    "macabre dancers" = @{
        NewName = "Macabre Dancers"
        NewText = "Unbowed: Heal 2.`nCornered: Resistant 1."
        Note = "Standardized ability words for Unbowed and Cornered."
        Flag = "Standardized to 'Unbowed: Heal 2.' and 'Cornered: Resistant 1.'."
    }
    "Chainbreaker Sappers" = @{
        NewName = "Chainbreaker Sappers"
        NewText = "When this unit deploys, deal 1 damage to chosen structure or Headquarters."
        Note = "Capitalized 'Headquarters'; terminal period."
        Flag = "None."
    }
    "Broken Horn Armorer" = @{
        NewName = "Broken Horn Armorer"
        NewText = "Whenever this unit moves to a new zone, chosen asset in the same zone gains 1 armor."
        Note = "MTG trigger phrasing 'Whenever [event]'."
        Flag = "None."
    }
    "Devilhorn Captain" = @{
        NewName = "Devilhorn Captain"
        NewText = "Unbowed: Deal 1 damage to chosen asset in the Main Theater."
        Note = "Standardized ability word; capitalized zone 'Main Theater'."
        Flag = "Standardized to 'Unbowed: Deal 1 damage...'."
    }
    "Trailblazer" = @{
        NewName = "Trailblazer"
        NewText = "Versatile"
        Note = "Capitalized keyword."
        Flag = "None."
    }
    "Improvised Bastion" = @{
        NewName = "Improvised Bastion"
        NewText = "Guardian"
        Note = "Capitalized keyword."
        Flag = "None."
    }
    "Saboteur" = @{
        NewName = "Saboteur"
        NewText = "When this unit deploys, deal 2 damage to chosen opposing non-Scratch unit and 1 damage to chosen allied unit."
        Note = "Capitalized 'non-Scratch'; terminal period."
        Flag = "None."
    }
    "Hammerhand Bivouac" = @{
        NewName = "Hammerhand Bivouac"
        NewText = "At the start of your turn, distribute 1 armor among allied units in the same zone."
        Note = "Terminal period added."
        Flag = "None."
    }
    "Chainbreaker Irregulars" = @{
        NewName = "Chainbreaker Irregulars"
        NewText = "Unbowed: Draw a card, then discard a card."
        Note = "Standardized ability word; 'Draw a card, then discard a card'."
        Flag = "Standardized to 'Unbowed: Draw a card...'."
    }
    "Necropolis Hunt" = @{
        NewName = "Necropolis Hunt"
        NewText = "When this unit is killed, deal damage equal to its power to chosen non-Headquarters asset."
        Note = "Capitalized 'non-Headquarters'; terminal period."
        Flag = "None."
    }
    "boneyard scavenger" = @{
        NewName = "Boneyard Scavenger"
        NewText = "T1, exhaust, kill this unit: Return chosen maneuver card from your discard pile to your hand."
        Note = "Standardized cost format 'T1, exhaust, kill this unit:'; specified 'maneuver card' and 'discard pile'."
        Flag = "Standardized cost ordering and '1T' -> 'T1'."
    }
    "ancestral monument" = @{
        NewName = "Ancestral Monument"
        NewText = "Other allied assets in the same zone have Cornered: Resistant 1."
        Note = "Standardized continuous grant syntax; ability word format."
        Flag = "Templated ability grant syntax."
    }
    "Chainbreaker Phalanx" = @{
        NewName = "Chainbreaker Phalanx"
        NewText = "Whenever this unit moves to the Main Theater, you may move chosen allied unit from your Reserves to the Main Theater. That unit gains 2 armor."
        Note = "MTG trigger phrasing 'Whenever [event]'; capitalized zones 'Main Theater' and 'Reserves'."
        Flag = "None."
    }
    "Devilhorn Guerillas" = @{
        NewName = "Devilhorn Guerrillas"
        NewText = "When this unit deploys, if it was deployed to the Main Theater, deal 2 damage to chosen opposing unit."
        Note = "Corrected title spelling from 'Guerillas' to 'Guerrillas'; capitalized zone 'Main Theater'."
        Flag = "Title spelling corrected ('Devilhorn Guerrillas')."
    }
    "ruin partisans" = @{
        NewName = "Ruin Partisans"
        NewText = "T2, exhaust, kill this unit: Return chosen unit card from your discard pile to your hand."
        Note = "Standardized cost format; specified 'unit card' and 'discard pile'."
        Flag = "Standardized cost ordering."
    }
    "Allshan, Chainbreaker Revenant" = @{
        NewName = "Allshan, Chainbreaker Revenant"
        NewText = "Unbowed: This unit gains 2 armor.`nCornered: Whenever this unit attacks, it gets +3 power until end of turn."
        Note = "Standardized ability words for Unbowed and Cornered; standardized 'gets +3 power until end of turn'."
        Flag = "Standardized ability words."
    }
    "Hammerhand Brigade" = @{
        NewName = "Hammerhand Brigade"
        NewText = "Armor 2"
        Note = "Standard keyword."
        Flag = "None."
    }
    "Ironclad Colossus" = @{
        NewName = "Ironclad Colossus"
        NewText = "Whenever this unit deploys or attacks, chosen unit gains 2 armor."
        Note = "MTG trigger phrasing 'Whenever [event]'."
        Flag = "None."
    }
    "Arkha, Sun's Fury" = @{
        NewName = "Arkha, Sun's Fury"
        NewText = "Whenever this unit attacks, you may put a unit card from your discard pile on the bottom of your deck. If you do, deal 3 damage to chosen opposing Headquarters."
        Note = "MTG trigger phrasing 'Whenever this unit attacks'; specified 'discard pile' and 'Headquarters'."
        Flag = "None."
    }
    "Solarheart Monk" = @{
        NewName = "Solarheart Monk"
        NewText = "When this card is put into your discard pile, chosen asset heals 2."
        Note = "Specified 'card' and 'discard pile' (triggers from non-battlefield zone per FR-ZON-003)."
        Flag = "Clarified trigger on 'this card' rather than 'this unit' while entering discard."
    }
    "boneyard amalgam" = @{
        NewName = "Boneyard Amalgam"
        NewText = "This unit costs R1 less to deploy for each unit card in your discard pile."
        Note = "Capitalized sentence; specified 'unit card' and 'discard pile'."
        Flag = "None."
    }

    # =========================================================================
    # UNITS & STRUCTURES - SOULLESS (units-and-structures-soulless.csv)
    # =========================================================================
    "visionary's eye" = @{
        NewName = "Visionary's Eye"
        NewText = "Whenever this unit attacks, draw a card."
        Note = "MTG trigger phrasing 'Whenever this unit attacks'; MTG 'draw a card'."
        Flag = "None."
    }
    "Manufactus Conscripts" = @{
        NewName = "Manufactus Conscripts"
        NewText = "Deploys Exhausted"
        Note = "Standardized to canonical keyword 'Deploys Exhausted' (FR-KWD-002l)."
        Flag = "Replaced sentence with standard keyword 'Deploys Exhausted'."
    }
    "brass adjutant" = @{
        NewName = "Brass Adjutant"
        NewText = "Armor 2"
        Note = "Capitalized keyword."
        Flag = "None."
    }
    "Conscript Rifles" = @{
        NewName = "Conscript Rifles"
        NewText = "When this unit deploys, deal 1 damage to chosen non-Scratch Company unit."
        Note = "Capitalized 'non-Scratch Company'; terminal period."
        Flag = "None."
    }
    "popup bunker" = @{
        NewName = "Popup Bunker"
        NewText = "When this structure deploys, each allied unit and structure in the same zone gains 1 armor."
        Note = "Terminal period added."
        Flag = "None."
    }
    "shrike scout suits" = @{
        NewName = "Shrike Scout Suits"
        NewText = "Armor 2"
        Note = "Capitalized keyword."
        Flag = "None."
    }
    "Magic Management Seekers" = @{
        NewName = "Magic Management Seekers"
        NewText = "Ward"
        Note = "Standard keyword."
        Flag = "None."
    }
    "Smelter Stoker" = @{
        NewName = "Smelter Stoker"
        NewText = "Aggressive 2"
        Note = "Capitalized keyword."
        Flag = "None."
    }
    "hellfire armorers" = @{
        NewName = "Hellfire Armorers"
        NewText = "Ward`nAllied units in the same zone have Aggressive 1."
        Note = "Capitalized keyword on line 1; continuous grant on line 2."
        Flag = "None."
    }
    "Comms Relay" = @{
        NewName = "Comms Relay"
        NewText = "Exhaust: Chosen asset gains Ward until the end of your opponent's turn."
        Note = "Standardized activated cost; capitalized 'Ward'; benchmark phrasing FR-LAY-003."
        Flag = "None."
    }
    "Mage Hunter Brigade" = @{
        NewName = "Mage Hunter Brigade"
        NewText = "Ward"
        Note = "Standardized keyword; stripped reminder text."
        Flag = "Reminder text stripped."
    }
    "Sentinel Warsuit Core" = @{
        NewName = "Sentinel Warsuit Core"
        NewText = "Armor 2"
        Note = "Capitalized keyword."
        Flag = "None."
    }
    "NO-AH-5" = @{
        NewName = "NO-AH-5"
        NewText = "Ward`nR2: Kill chosen allied non-Headquarters asset and tuck it face down under this unit.`nThis unit has +2 power for each card tucked under it."
        Note = "Benchmark text from Master FRD Section 8.2; standardized tucking terminology (FR-ENT-102)."
        Flag = "Harmonized with canonical benchmark in Master FRD Section 8.2."
    }
    "Mobile Armory" = @{
        NewName = "Mobile Armory"
        NewText = "Exhaust: Chosen allied unit in the same zone gets +1 power until end of turn."
        Note = "Standardized activated cost; 'gets +1 power until end of turn' instead of 'this turn'."
        Flag = "Changed 'this turn' to 'until end of turn'."
    }
    "Piston Fist Prototype" = @{
        NewName = "Piston Fist Prototype"
        NewText = "After this unit attacks, move this unit to your Reserves."
        Note = "Canonical text from Master FRD Section 8.2 benchmark."
        Flag = "Harmonized with Master FRD benchmark."
    }
    "field mechanic" = @{
        NewName = "Field Mechanic"
        NewText = "When this unit deploys, chosen allied Mechanized unit gains 2 armor and gets +2 power until end of turn."
        Note = "Capitalized subtype 'Mechanized'; standardized 'gets +2 power until end of turn'."
        Flag = "None."
    }
    "Dolus Macrocannon" = @{
        NewName = "Dolus Macrocannon"
        NewText = "R1, exhaust: Deal 2 damage to chosen non-Scratch asset."
        Note = "Standardized activated cost 'R1, exhaust:'; capitalized 'non-Scratch'."
        Flag = "None."
    }
    "Dolus Refiner" = @{
        NewName = "Dolus Refiner"
        NewText = "Exhaust: Draw 2 cards, then discard 2 cards."
        Note = "Standardized activated cost; MTG sequential phrasing 'then discard 2 cards'."
        Flag = "None."
    }
    "Culverin Brigade" = @{
        NewName = "Culverin Brigade"
        NewText = "When this unit deploys, deal 1 damage to each of up to 2 chosen opposing assets."
        Note = "Clarified distributed damage phrasing 'each of up to 2 chosen opposing assets'."
        Flag = "None."
    }
    "Aggressor Assault Suit" = @{
        NewName = "Aggressor Assault Suit"
        NewText = "Armor 2"
        Note = "Capitalized keyword."
        Flag = "None."
    }
    "MAX-8 Autonomous Warsuit" = @{
        NewName = "MAX-8 Autonomous Warsuit"
        NewText = "When this unit deploys, deal 2 damage to chosen asset, and chosen asset heals 2."
        Note = "Capitalized sentence; comma separation; terminal period."
        Flag = "None."
    }
    "Repurposed Industrial Loader" = @{
        NewName = "Repurposed Industrial Loader"
        NewText = "Stalwart"
        Note = "Standard keyword."
        Flag = "None."
    }
    "Redlined Brawler" = @{
        NewName = "Redlined Brawler"
        NewText = "When this unit deploys, you may discard 2 cards. If you do, move this unit to the Main Theater."
        Note = "Capitalized zone 'Main Theater'; terminal period."
        Flag = "None."
    }
    "Earthshaker Colossus" = @{
        NewName = "Earthshaker Colossus"
        NewText = "Whenever this unit deploys or attacks, deal 2 damage to chosen non-Scratch asset."
        Note = "MTG trigger phrasing 'Whenever [event]'; capitalized 'non-Scratch'."
        Flag = "None."
    }
    "Land Fortress" = @{
        NewName = "Land Fortress"
        NewText = "Exhaust: Chosen unit in the same zone gets +2 power until end of turn and gains 1 armor."
        Note = "Standardized activated cost; standardized 'gets +2 power until end of turn'."
        Flag = "None."
    }
    "dominator warsuit" = @{
        NewName = "Dominator Warsuit"
        NewText = "R1: This unit gains 1 armor."
        Note = "Standardized activated cost 'R1:'; capitalized sentence."
        Flag = "None."
    }
}

# Subtype title case normalizer
function Normalize-Subtypes([string]$raw) {
    if ([string]::IsNullOrWhiteSpace($raw) -or $raw.Trim() -eq "--") { return "--" }
    $parts = $raw.Split(',') | ForEach-Object {
        $part = $_.Trim()
        # Custom known casing
        if ($part -match '^scratch company$') { "Scratch Company" }
        elseif ($part -match '^giant$') { "Giant" }
        else {
            (Get-Culture).TextInfo.ToTitleCase($part.ToLower())
        }
    }
    return ($parts -join ', ')
}

# Deployment zone normalizer
function Normalize-Zones([string]$raw) {
    if ([string]::IsNullOrWhiteSpace($raw)) { return "" }
    $parts = $raw.Split(',') | ForEach-Object {
        $p = $_.Trim().ToUpper()
        if ($p -eq 'MAIN') { 'MAIN_THEATER' }
        elseif ($p -eq 'RESERVES') { 'RESERVES' }
        elseif ($p -eq 'HEADQUARTERS') { 'HEADQUARTERS' }
        else { $p }
    }
    # Standard ordering
    $sorted = $parts | Sort-Object -Unique
    return ($sorted -join ', ')
}

# Process all 10 CSVs
$csvFiles = Get-ChildItem -Path $cardsDir -Filter "*.csv"
$auditLog = [System.Collections.Generic.List[PSCustomObject]]::new()

foreach ($csv in $csvFiles) {
    Write-Host "Standardizing $($csv.Name)..." -ForegroundColor Cyan
    $rows = Import-Csv -Path $csv.FullName
    $newRows = [System.Collections.Generic.List[PSCustomObject]]::new()
    
    foreach ($row in $rows) {
        $origName = $row.Name
        $isPlaceholder = [string]::IsNullOrWhiteSpace($origName)
        
        # New object
        $p = [ordered]@{}
        foreach ($prop in $row.PSObject.Properties) {
            $p[$prop.Name] = $prop.Value
        }
        $newObj = [PSCustomObject]$p
        
        if ($isPlaceholder) {
            # Normalize enum for placeholder rows
            if ($newObj.Rarity -eq 'U') { $newObj.Rarity = 'UNCOMMON' }
            elseif ($newObj.Rarity -eq 'C') { $newObj.Rarity = 'COMMON' }
            elseif ($newObj.Rarity -eq 'R') { $newObj.Rarity = 'RARE' }
            elseif ($newObj.Rarity -eq 'M') { $newObj.Rarity = 'MYTHIC' }
            $newRows.Add($newObj)
            continue
        }
        
        # Find rule
        $lookupKey = $origName
        $rule = $textRules[$lookupKey]
        if (-not $rule) {
            # Try lowercase match
            $matchedKey = $textRules.Keys | Where-Object { $_ -eq $origName.ToLower() } | Select-Object -First 1
            if ($matchedKey) { $rule = $textRules[$matchedKey] }
        }
        
        if (-not $rule) {
            Write-Warning "No rule found for '$origName' in $($csv.Name)!"
        }
        
        # Update Name
        if ($rule.NewName) { $newObj.Name = $rule.NewName }
        
        # Normalize Card Type
        $ct = $newObj.'Card Type'
        if ($ct -eq 'Unit') { $newObj.'Card Type' = 'UNIT' }
        elseif ($ct -eq 'Structure') { $newObj.'Card Type' = 'STRUCTURE' }
        elseif ($ct -eq 'Maneuver') { $newObj.'Card Type' = 'MANEUVER' }
        elseif ($ct -eq 'HQ') { $newObj.'Card Type' = 'HEADQUARTERS' }
        
        # Normalize Faction
        $fac = $newObj.Faction
        if ($fac -eq 'Animus') { $newObj.Faction = 'ANIMUS' }
        elseif ($fac -eq 'Concordia') { $newObj.Faction = 'CONCORDIA' }
        elseif ($fac -eq 'Path') { $newObj.Faction = 'PATH' }
        elseif ($fac -eq 'Soulless') { $newObj.Faction = 'SOULLESS' }
        elseif ($fac -eq 'Neutral') { $newObj.Faction = 'NEUTRAL' }
        
        # Normalize Rarity
        $rar = $newObj.Rarity
        if ($rar -eq 'C') { $newObj.Rarity = 'COMMON' }
        elseif ($rar -eq 'U') { $newObj.Rarity = 'UNCOMMON' }
        elseif ($rar -eq 'R') { $newObj.Rarity = 'RARE' }
        elseif ($rar -eq 'M') { $newObj.Rarity = 'MYTHIC' }
        elseif ($rar -eq 'HQ') { $newObj.Rarity = 'HQ' }
        
        # Normalize Deployment Zones
        if ($newObj.PSObject.Properties['Deployment Zones']) {
            $newObj.'Deployment Zones' = Normalize-Zones $newObj.'Deployment Zones'
        }
        
        # Normalize Types
        if ($newObj.PSObject.Properties['Types']) {
            $newObj.'Types' = Normalize-Subtypes $newObj.'Types'
        }
        
        # Update Rules Text
        $origText = $row.Text
        $newText = if ($rule.NewText) { $rule.NewText } else { $origText }
        $newObj.Text = $newText
        
        # Record Audit Entry
        $auditLog.Add([PSCustomObject]@{
            File = $csv.Name
            OriginalName = $origName
            StandardizedName = $newObj.Name
            Faction = $newObj.Faction
            CardType = $newObj.'Card Type'
            Rarity = $newObj.Rarity
            OriginalText = $origText
            StandardizedText = $newText
            TextChanged = ($origText -ne $newText)
            Note = if ($rule.Note) { $rule.Note } else { "No text change." }
            Flag = if ($rule.Flag) { $rule.Flag } else { "None." }
        })
        
        $newRows.Add($newObj)
    }
    
    # Write back standardized CSV
    Export-QuotedCsv -Rows $newRows -Path $csv.FullName
}

Write-Host "`nAll 10 CSVs standardized successfully!" -ForegroundColor Green

# Output summary counts
$changedCount = ($auditLog | Where-Object { $_.TextChanged }).Count
$flaggedCount = ($auditLog | Where-Object { $_.Flag -ne "None." }).Count
Write-Host "Total Cards Audited: $($auditLog.Count)"
Write-Host "Cards with Text Standardized: $changedCount"
Write-Host "Cards with Review Flags: $flaggedCount"

# Save audit log to JSON in scratch directory
$jsonPath = Join-Path $PSScriptRoot "card_standardization_audit.json"
$auditLog | ConvertTo-Json -Depth 5 | Set-Content -Path $jsonPath -Encoding UTF8
Write-Host "Audit log saved to $jsonPath"
