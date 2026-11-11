# M2 Build Board — Done Tasks Archive

Completed task rows moved from [`../../BOARD.md`](../../BOARD.md). Each task's file (`T-NNNN-<slug>.md`) sits alongside this index in `.agents/tasks/m2/archive/tasks/`.

| id | title | status |
| :--- | :--- | :--- |
| T-0001 | Add a fences check to cardtown lint | done |
| T-0002 | Scaffold the ability and cards packages | done |
| T-0003 | Validate deck construction (FR-SET-001) | done |
| T-0004 | Load starter decklists and run them in cardtown sim | done |
| T-0005 | Make the pre-commit hook rebuild a stale CLI | done |
| T-0006 | Ask players through a per-seat Chooser (FR-MOD-001) | done |
| T-0007 | Run self-deploy triggers through a pending buffer (FR-TRG-003, FR-TRG-004) | done |
| T-0008 | Fire death triggers from the death sweep in APNAP order (FR-TRG-002, FR-TRG-003) | done |
| T-0009 | Let death triggers move the card itself, tracked in discard (FR-ZON-004) | done |
| T-0010 | Attack triggers with Aggressive at Step 2, "on your turn" gate, opponents subject (FR-CBT-002) | done |
| T-0011 | Make Spoil Forest Creeper executable (A1 card, flash) | done |
| T-0012 | Make Despoiler Larva, Fungal Evangelist and Titanic Spore Carrier executable (A1 cards, flash) | done |
| T-0013 | Make Portent Scrivener and Despoiler Worm executable (A1 cards, flash) | done |
| T-0014 | Make Stormfront of Ideals and Unceasing Tide executable (A1 cards, flash) | done |
| T-0015 | Pin FR-SET-003, FR-CRD-101, FR-ZON-003 and FR-ZON-001 with scenario tests (directive #16) | done |
| T-0016 | Refresh stale milestone notes in engine comments and messages (directive #17, harness, flash) | done |
| T-0017 | Chosen targets with conjunction filters, two-stage verification, and DealDamage (A2 shape, sonnet) | done |
| T-0018 | Make Saboteur and Chainbreaker Sappers executable (A2 cards, flash) | done |
| T-0019 | Gate deploy triggers on the deployed zone, and add Heal and AddArmor steps (A2 shape, sonnet) | done |
| T-0020 | Structured game events, one observer, settled notifications and match records (directive #18 part 1, harness, sonnet) | done |
| T-0021 | cardtown sim -log and -events (JSON Lines), with sorted end reasons (directive #18 part 2 and #12, harness, sonnet) | done |
| T-0022 | A random sim policy that plays expensive cards (directive #21, harness, sonnet) | done |
| T-0023 | Replay a recorded game to any event, as a function and a cardtown replay command (directive #19, harness, sonnet) | done |
| T-0024 | Move the sim and event-log code out of package engine (directive #22, harness, sonnet; runs before T-0023) | done |
| T-0025 | Rule monitors in every sim game, with a coverage table: framework and three monitors (directive #23 part 1, harness, sonnet) | done |
| T-0026 | Damage-stage and stack-placement event facts, with the damage-sequence and targeting monitors (directive #23 part 2, harness, sonnet) | done |
| T-0027 | Make Devilhorn Guerrillas, Spoil Forest Colossus and Ironclad Colossus executable (A2 cards, flash) | done |
| T-0028 | Damage equal to the source's power, last-known effective state, and steps on all matching assets (A2 shape, opus) | done |
| T-0029 | Make Necropolis Hunt executable (A2 card, flash) | done |
| T-0030 | Optional ("you may") triggers, "up to N" chosen targets, and "other" (A2 shape, opus) | done |
| T-0031 | Make Craterfall Ancient executable (A2 card, flash; last of the wave) | done |
| T-0032 | Split package engine's files by responsibility, pure move (directive #26, harness, sonnet) | done |
| T-0033 | One deck-naming scheme and one build-revision reader (directive #26, harness, sonnet) | done |
| T-0034 | Remove duplicated rule code: Berserk targeting, combat zones, Scratch Company face, monitor lookups (directive #26, harness, sonnet) | done |
| T-0035 | Split cmd/cardtown/main.go into one file per command, pure move (directive #26, harness, flash) | done |
| T-0036 | PlayerView, a per-seat view of the game and its event log (directive #24; Deliverable 5 part 1; harness, opus high) | done |
| T-0037 | cardtown play, a text-mode match against a sim policy, drawn only from PlayerView (directive #24; Deliverable 5 part 2; harness, flash, escalated to sonnet in review round 2) | done |
| T-0045 | Export Decision.Valid and use it in cardtown play's terminal chooser (directive #38, harness, sonnet) | done |
| T-0046 | One hidden-information oracle for the leak tests, in pkg/view/viewtest (directive #39, harness, flash) | done |
| T-0047 | One match-setup scheme, sim.MatchDecks, sim.SeatSeed and sim.DefaultEndBias (directive #41, harness, sonnet high) | done |
| T-0048 | Split cmd/cardtown/play.go into the loop, the screen and input (directive #42, harness, flash, pure move) | done |
| T-0049 | normalizeArgs as a small alias map (directive #44, harness, flash) | done |
| T-0050 | Damage-taken hook for Unbowed (FR-KWD-002a), with step<Op> functions for effect ops (wave B shape 1, sonnet) | done |
| T-0051 | Make Hammerhand Brigands executable (wave B card, flash) | done |
| T-0052 | A zone criterion on Filter ("in the Main Theater", FR-TGT-001), with Devilhorn Captain (wave B shape 2, sonnet) | done |
| T-0053 | Re-verify an attack's defender against the attacker's current controller (SVY-R8, directive #29, shape, sonnet) | done |
| T-0054 | A once-per-turn limit on triggers (Trigger.MaxPerTurn), for Animoceps Rager (wave B shape 3, sonnet) | done |
| T-0055 | Make Animoceps Rager executable, the first Headquarters ability (wave B card, flash) | done |
| T-0056 | No random fallbacks: a timed-out target picks the first legal options in fixed order (SVY-R16, directive #47, shape, sonnet) | done |
| T-0057 | The Scratch Company face prints Deploys Exhausted, and a scratch is a unit deploying (directive #48, shape, sonnet) | done |
| T-0058 | One place exhausts an asset, (g *Game) exhaust (directive #49, harness, flash, pure move) | done |
| T-0059 | One place moves a card between deck, hand and discard, takeCard and putCard (directive #52, harness, sonnet, pure move) | done |
| T-0060 | A stack item carries an effect, not a whole trigger (directive #53, harness, flash, pure move) | done |
| T-0061 | One place for the battlefield orders, in state.go (directive #54, harness, flash, pure move) | done |
| T-0062 | Fired triggers are placed when the stack starts resolving, not only after an item (directive #50 part 1, harness, flash, pure move) | done |
| T-0063 | An exhaust hook, fired from (g *Game) exhaust on every change from Readied to Exhausted (SVY-R17, directive #50 part 2, wave B shape, sonnet) | done |
| T-0064 | Make Animoceps Shambler executable (wave B card, flash) | done |
| T-0065 | One place builds an asset entering the battlefield, (g *Game) newAsset (directive #56, harness, flash, pure move) | done |
| T-0066 | Simultaneous triggers fall back to source position, not step order (directive #57, shape, sonnet) | done |
| T-0067 | A discard step that picks from your own hand, with the most-recently-added timeout (SVY-R16, wave B shape, sonnet) | done |
| T-0068 | Make Chainbreaker Irregulars executable (wave B card, flash) | done |
| T-0069 | A Cornered gate on triggers, and a step that gives power until end of turn (FR-KWD-002e, FR-LAY-003, wave B shape, sonnet) | done |
| T-0070 | Make Allshan, Chainbreaker Revenant executable (wave B card, flash) | done |
| T-0071 | Effect steps take one per-resolution struct (directive #51 part 1, harness, flash, pure move) | done |
| T-0072 | "You may X. If you do, Y": an untargeted pick on resolution, a kill step, and steps that report what they did (SVY-R14, directive #51 part 2, wave B shape, sonnet) | done |
| T-0074 | Aggressive is an ability stack item, marked as coming from the keyword (directive #60 part 1, harness, sonnet) | done |
| T-0075 | Allshan reads "Cornered: Aggressive 3." (directive #60 part 2, card, flash) | done |
| T-0073 | Make Infested Apostle executable (wave B card, flash) | done |
| T-0076 | An untargeted pick is its own decision kind, DecidePick, not a Target (SVY-R3, SVY-R14, directive #20 (a), wave B shape, sonnet) | done |
| T-0077 | An untargeted pick from your discard pile, which moves the picked card, with a top-down timeout (SVY-R14, SVY-R16, directive #59 (1), wave B shape, sonnet) | done |
| T-0078 | Make Arkha, Sun's Fury executable (wave B card, flash) | done |
| T-0079 | The view leak check accepts a name that was public when its event happened (harness, sonnet) | done |
| T-0080 | A combat-kill hook, fired when an attacker's combat damage destroys its defender (FR-CBT-002, wave B shape, sonnet) | done |
| T-0081 | A chosen target in a discard pile, chosen on placement and verified at resolution (FR-TGT-001, FR-TGT-003, wave B shape, sonnet) | done |
| T-0082 | Make Eternal Watchman executable (wave B card, flash) | done |
| T-0083 | Modal triggers, "choose one" of their modes on placement, before targets (FR-MOD-001 §5, §6, wave B shape, sonnet) | done |
| T-0084 | Make Rogue Fleshshaper executable (wave B card, flash) | done |
| T-0085 | Placing an effect chooses its modes and targets from a placement, not a pending trigger (directive #62, harness, sonnet, pure move) | done |
| T-0086 | One untargeted pick body, pickFrom, and one predicate for "this filter chooses cards in a pile", Filter.InPile (directives #63 and #66, harness, flash, pure move) | done |
| T-0087 | One place discards from hand, (g *Game) discardFromHand (directive #64, harness, flash, pure move) | done |
| T-0088 | The view asks the engine whether a chosen card is still in a pile, (g *Game) DiscardCard (directives #65 and #67, harness, flash, pure move) | done |
| T-0089 | Berserk's rules leave the kernel, so an Attack has no secondary target and nothing cleaves (directive #68 part (a), harness, sonnet) | done |
| T-0090 | The Berserk keyword leaves the kernel, so a "Berserk N" line loads as unsupported (directive #68 part (b), harness, flash) | done |
| T-0091 | The clients stop reading an Attack's secondary target (directive #68 part (c), harness, flash) | done |
| T-0092 | Events carry no secondary target, and doc.go stops listing Berserk (directive #68 part (d1), harness, flash) | done |
| T-0093 | An Attack has one target, so Action and StackItem lose Secondary (directive #68 part (d2), harness, flash) | done |
| T-0094 | Attack legality lives in combat.go, (g *Game) checkAttack (directive #69, harness, flash, pure move) | done |
| T-0095 | The enum convention is written down, and ValueFixed becomes ValueSourceFixed (directive #70, harness, flash, comments and a rename) | done |
| T-0096 | Destruction replacement at the death sweep, one death at a time in its controller's order (FR-REP-001, SVY-R5, directive #73 part 1, wave G shape, sonnet) | done |
| T-0097 | Kill steps go through destruction replacement, and a replaced kill still counts as done (FR-REP-001, SVY-R15, directive #73 part 2, wave G shape, sonnet) | done |
| T-0101 | A replacement sees every death at the same moment, in a fixed order, with no destruction partway through a sweep (FR-REP-001, SVY-R20, SVY-R21, SVY-R22, directive #74, wave G shape, sonnet) | done |
| T-0100 | When several replacements apply to one death, its controller chooses which, death by death (FR-REP-001, FR-MOD-001 §6, SVY-R23, directive #75, wave G shape, sonnet) | done |
| T-0098 | A Scratch Company filter, and Shifting Impossibility (FR-TGT-001, FR-REP-001, wave G shape, sonnet) | done |
| T-0099 | A combat-kill trigger fires only when the sweep destroys the defender, so a replaced death is no kill (FR-CBT-002, SVY-R19, directive #73, wave G shape, sonnet) | done |
| T-0102 | A step returns or puts any asset it acts on, not only a replaced one (FR-ZON-004, FR-TGT-003, directive #76, wave G.1 shape, sonnet) | done |
| T-0103 | One APNAP order, (g *Game) apnap, and replacedQueue takes "oldest first" from assetsByID (directive #77, harness, flash, pure move) | done |
| T-0104 | Scratch Company in a filter is one three-way enum, ability.ScratchCompany (directive #78, harness, flash, no behavior change) | done |
| T-0105 | The death sweep gets its own file, sweep.go (directive #79, harness, flash, pure move) | done |
| T-0106 | A destruction replacement saves an asset at most once per sweep, and the next pass takes it up again (FR-REP-003, SVY-R24, directive #81, wave G.1 shape, sonnet) | done |
| T-0107 | A step acts only on chosen assets still on the battlefield, so a later step does nothing to a target an earlier one removed (FR-ZON-004, FR-TGT-003, directive #82, wave G.2 shape, sonnet) | done |
| T-0108 | Killing is one function, (g *Game) kill, and moving between battlefield zones is one, (g *Game) moveOnBattlefield (directives #83 and #84, harness, flash, pure move) | done |
| T-0109 | Keyword names live in one table, keywordDefs, that ParseKeywordLine and Keywords.Lines both walk (directive #85, harness, flash, pure move) | done |
| T-0110 | Event kinds are named constants in events.go, and events.go, choice.go, sweep.go and zones.go emit with them (directive #86 part 1 of 5, harness, flash, pure move) | done |
| T-0111 | actions.go, combat.go and damage.go emit with the event kind constants (directive #86 part 2 of 5, harness, flash, pure move) | done |
| T-0112 | effects.go, replace.go and stack.go emit with the event kind constants (directive #86 part 3 of 5, harness, flash, pure move) | done |
| T-0113 | pkg/sim matches events by the event kind constants, and actions by engine.ActAttack.String() and its siblings (directive #86 part 4 of 5, harness, flash, pure move) | done |
| T-0114 | pkg/view and pkg/eventlog match events by the event kind constants, so no event kind is a literal outside events.go (directive #86 part 5 of 5, harness, flash, pure move) | done |
| T-0115 | File comments on damage.go, choice.go, invariants.go and layers.go, so every engine file says its job above the package line (directive #87 (c), harness, flash, comments only) | done |
| T-0038 | Fence every package under pkg/, and check the client packages' imports (directive #30, harness, sonnet) | done |
| T-0039 | Move sim, eventlog and view to the top of pkg/ (directive #31, harness, flash, pure move) | done |
| T-0040 | Package comments for sim and eventlog (directive #33, harness, flash, comments only) | done |
| T-0041 | Split pkg/engine/events.go into events, snapshots and build info (directive #34, harness, flash, pure move) | done |
| T-0042 | Split pkg/sim/monitor.go into the framework and one file per monitor (directive #35, harness, flash, pure move) | done |
| T-0043 | One card-sheet reader for the sheets, syntax and fences linters (directive #36, harness, sonnet) | done |
| T-0044 | Run git through pkg/toolchain/git in audit, docs and tddslice (directive #37, harness, sonnet) | done |
| T-0116 | Enact a Maneuver from hand, resolve its effect and put it into its owner's discard pile (FR-MNV-001, FR-TRN-003, directive #88 part 1, wave C shape, sonnet) | done |
| T-0117 | A Maneuver can be enacted only if each exactly-one target slot has a legal entity, and a "Choose N" one only with N legal modes (FR-MNV-001, SVY-R29, directive #88 part 2, wave C shape, sonnet) | done |
| T-0118 | A chosen asset can deal a step's damage, equal to its current power, and deals none once it is gone (FR-TGT-003, SVY-R27, SVY-R31, directive #88 part 3, wave C shape, sonnet) | done |
| T-0119 | An effect can move a unit to its controller's Reserves or the Main Theater (`MOVE_ZONE`), keeping its state (FR-ZON-002, FR-ZON-004, SVY-R26, directive #88 part 4, wave C shape, sonnet) | done |
| T-0120 | Make Despoiler's Emergence and Deafening Aria executable (damage to all opposing assets in a zone, wave C card, flash) | done |
| T-0121 | The progress policy enacts Maneuvers, and the coverage tests count them (D6 for T-0120, harness, sonnet) | done |
| T-0122 | The view, viewtest and the play screen show an enacted Maneuver (directive #89, harness, flash) | done |
| T-0123 | Make Devilhorn Sanction executable (kill a chosen opposing unit, wave C card, flash) | done |
| T-0124 | Remove the retired cardtown audit command and pkg/toolchain/audit (directive #90, harness, sonnet) | done |
| T-0125 | Make Improvised Orders executable (draw 3, then discard a card, wave C card, flash) | done |
| T-0126 | Make Animoceps Spores executable (Scratch 2, then return an allied Scratch Company unit, wave C card, flash) | done |
| T-0127 | Make Rapturous Sermon executable (move all allied Scratch Company units to the Main Theater, then +2 power until end of turn, wave C card, flash) | done |
| T-0128 | A Maneuver can have an additional cost that kills a chosen allied unit, paid on declaration, and Rhapsody's Embrace and Feed the Despoiler use it (FR-TRN-003a, SVY-R2, SVY-R15, wave C shape, sonnet) | done |
| T-0129 | An additional cost can exhaust a readied allied unit, and exhausting an Exhausted asset does nothing (FR-TRN-003a, SVY-R17, directive #58, wave C shape, sonnet) | done |
| T-0130 | Make Funeral Fungus executable (exhaust a readied allied unit as an additional cost, then return chosen unit card from your discard pile to your hand, wave C card, flash) | done |
| T-0131 | An additional cost can discard up to N cards from the payer's hand (FR-TRN-003a, SVY-R3, directive #20 (b), wave C shape, sonnet) | done |
| T-0132 | A step can run only if its controller has no cards in hand (FR-MNV-001, FR-TRN-003a, SVY-R3, SVY-R14, directive #20 (b), wave C shape, sonnet) | done |
| T-0133 | Make Disorienting Pulse and Tactical Withdrawal executable (return a chosen unit to its owner's hand, wave C card, flash) | done |
| T-0134 | Make Conversion Spike executable (damage a chosen opposing unit, heal a chosen allied unit or structure, wave C card, flash) | done |
| T-0135 | Make Chainbreaker's Tenacity and Fortify executable (a chosen allied unit gains armor, Tenacity also +1 power until end of turn, wave C card, flash) | done |
| T-0136 | A timed-out repeated choice puts every repetition on the first option (SVY-R32, #91 (1), wave C shape, sonnet) | done |
| T-0137 | A discard cost's cards are chosen before the targets and discarded after the push (#20 (b), wave C shape, sonnet) | done |
| T-0138 | A target slot can choose once per card the discard cost discarded (SVY-R4, #20 (b), wave C shape, sonnet) | done |
| T-0139 | Make *Steel Rain* executable (discard up to 3, 2 damage to a chosen asset per card discarded, draw 2 if the hand is empty; #20 (b), wave C card, flash) | done |
| T-0140 | An "another chosen" slot must pick a different entity from the earlier slot of its clause (SVY-R34, #88, #91 (3), wave C shape, sonnet) | done |
| T-0141 | Make *Ambush Tactics* executable (choose two of kill, return to hand, a chosen unit deals its power to another chosen unit, put on top of its owner's deck; #14 (a), wave C card, flash) | done |
| T-0142 | Make *Ensnaring Mist* executable (move a chosen unit to its controller's Reserves, then a chosen unit gets -1 power until end of turn; #91 (4), #14 (c), wave C card, flash) | done |
| T-0143 | Make *Pincer Tactics* executable (1 damage to each of up to 2 chosen allied units, 2 to a chosen non-Scratch asset; #91 (4), #14 (c), wave C card, flash) | done |
| T-0144 | One declare sequence, (g *Game) declare, and one read of a card's cost, (g *Game) EffectiveCost (directives #92 and #93, harness, flash, pure move) | done |
| T-0145 | Additional costs and a card's cost get their own file, costs.go (directive #94, harness, flash, pure move) | done |
| T-0146 | One function deploys a card onto the battlefield, deployCard (directive #95, harness, flash, pure move) | done |
| T-0147 | One function buffers a fired trigger, bufferTrigger (directive #96, harness, flash, pure move) | done |
| T-0148 | One read of "another", anotherOf, and modes joined in one place, joinModes (directive #97, harness, flash, pure move) | done |
| T-0149 | A resource's name lives in one place, Resource.String (directive #98, part 1 of 3, harness, flash, pure move) | done |
| T-0150 | Event readers compare with Resource.String, not typed-out names (directive #98, part 2 of 3, harness, flash, pure move) | done |
| T-0151 | One check for a Scratch Company card, IsScratchCompany (directive #98, part 3 of 3, harness, flash, pure move) | done |
| T-0152 | One table of policy names, sim.PolicyByName (directive #99, harness, flash, pure move) | done |
| T-0153 | The linters share one scan, pkg/toolchain/specdoc (directive #100, harness, flash, pure move) | done |
| T-0154 | Zone helpers live with the zones, deployedTo and battlefieldZone move to zones.go (directive #101, harness, flash, pure move) | done |
| T-0155 | The docs say what wave C built, doc.go, sim.go, ability.go and zones.go comments (directive #102, harness, flash, comments only) | done |
| T-0156 | Remove the TDD commands from cardtown and whats-up (directive #103 (b)-(e), harness, flash) | done |
| T-0157 | Delete the TDD packages and their git helpers, and finish directive #103 (harness, sonnet) | done |
| T-0158 | Activate a printed ability of a battlefield asset, its costs paid and its targets chosen on declaration (FR-ACT-001, FR-ACT-002, FR-TRN-003a, SVY-R2, wave D shape, sonnet) | done |
| T-0159 | The event log, replay and sim policy know an activation (harness, sonnet) | done |
| T-0160 | The docs say activated abilities are built, stack.go, doc.go, costs.go and effects.go comments (harness, flash, comments only) | done |
| T-0161 | The play screen names an activation, its asset and printed text (harness, flash) | done |
| T-0162 | A filter can ask for any of several subtypes, and for a Maneuver card (wave D shape, sonnet) | done |
| T-0163 | An activated ability can be activated only if its required targets exist (wave D shape, sonnet) | done |
| T-0164 | The play screen's not-executable tag is pinned on its own, not through one seat's starter session (harness, flash, tests only) | done |
| T-0165 | Make Sinuous Growth executable (wave D card, flash) | done |
| T-0166 | Make Boneyard Scavenger and Ruin Partisans executable (wave D card, flash) | done |
| T-0167 | A step can search its controller's deck for a card (wave D shape, sonnet) | done |
| T-0168 | The play log hides an opponent's deck search pick, and a search replays (harness, flash) | done |
| T-0169 | Make Cliff of Portents executable (wave D card, flash) | done |
| T-0170 | One read of "can pay" and one payment, `(*Player).canPay` and `(*Player).pay` (directive #107, harness, flash, pure move) | done |
| T-0171 | One function plays a card from hand, `playFromHand`, and one `emitAction` case (directive #108, harness, flash, pure move) | done |
| T-0172 | The placement builds its stack item, `placement.stackItem`, and `declare` takes a kind (directive #109, harness, flash, pure move) | done |
| T-0173 | One "first picks in this order", `firstPicks`, and one `listedOrder` (directive #110, harness, flash, pure move) | done |
| T-0174 | One "this asset is yours" check, `ownAsset`, and "Aggressive" spelled once (directives #111 and #112, harness, flash, pure move) | done |
| T-0175 | Filters get their own file, `filter.go` (directive #113, harness, flash, pure move) | done |
| T-0176 | The resource economy gets its own file, `resources.go`, and `exhaust` moves to `state.go` (directive #114, harness, flash, pure move) | done |
| T-0177 | The fences check the engine's imports (directive #115, harness, flash) | done |
| T-0178 | Snapshots and damage events report effective keywords (directive #116 (a), harness, flash) | done |
| T-0179 | Rule monitors read effective keywords (directive #116 (b), harness, flash) | done |
| T-0180 | A keyword behind Cornered is one continuous shape (directive #117, shape, sonnet) | done |
| T-0181 | The view shows a keyword behind Cornered, and the monitors stay clean on it (directive #117, harness, flash) | done |
| T-0182 | The hand view shows what a card costs now (directive #118 (a), harness, flash) | done |
| T-0183 | A hidden decision stays hidden by its flag (directive #119, harness, flash) | done |
| T-0184 | The docs say what wave D built, doc.go, placement.go and ability.go comments (directive #121, harness, flash, comments only) | done |
| T-0185 | Unstarve ProgressPolicy without scenario seeding (directive #122, harness, flash) | done |
| T-0186 | Drop the search clause from the hidden-decision check (directive #123, harness, flash) | done |
| T-0187 | Decklist parse errors name the file (directives #124 and #8, harness, flash) | done |
| T-0188 | A value doesn't need a stack item, `value` takes the controller, source and instance (directive #125, harness, flash, pure move) | done |
| T-0189 | Effective reads live in `layers.go`, `cornered`, `EffectiveState` and `effectiveState` move there (directive #126, harness, flash, pure move) | done |
| T-0190 | The policy asks the engine what an activated ability costs (directive #127, harness, flash) | done |
| T-0191 | Effects log what they do (directive #129, harness, flash) | done |
| T-0192 | One read of an action event's kind, `(Event).ActionKind` read by `actionOf` and `Monitors.Event` (directive #130, harness, flash) | done |
| T-0193 | One read of "the source, or its last-known record", `resolveAttack` calls `itemSource` (directive #131, harness, flash, pure move) | done |
| T-0194 | Conversion and deck names, each spelled once (directive #132 (a), (b), harness, flash, pure move) | done |
| T-0195 | One read of "this faction or Neutral", and no face-down guard that can't fire (directive #132 (c), (d), harness, flash, pure move) | done |
| T-0196 | One reader of the keyword matrix, and `Analyze` split by job (directive #133, harness, flash, pure move) | done |
| T-0197 | One classifier of changed files, and "all linters" means all (directive #134, harness, flash) | done |
| T-0198 | The text says what the code does (directive #135, harness, flash) | done |
| T-0199 | Make Macabre Dancers executable (Unbowed heals 2, Cornered Resistant 1, wave E card, flash) | done |
| T-0200 | Hand-active cost reduction, first card Boneyard Amalgam (wave E shape, sonnet) | done |
| T-0201 | `value` takes only what it reads, and the docs say what wave E built (directives #138 and #139, harness, flash, pure move) | done |
| T-0202 | *Chainbreaker's Tenacity*'s armor grant is its own sentence (directive #140 (1), harness, flash, text only) | done |
| T-0203 | The moves-zone hook, for Chainbreaker Phalanx and Broken Horn Armorer (wave F shape, sonnet) | done |
| T-0204 | Chainbreaker Phalanx (wave F card, flash) | done |
| T-0205 | A zone locked on the stack item, first a move's destination (wave F shape, sonnet; #141 (1)) | done |
| T-0206 | The "in that zone" filter criterion, read against the locked zone (wave F shape, sonnet; #141 (1)) | done |
| T-0207 | Broken Horn Armorer (wave F deck card, flash; #141 (2)) | done |
| T-0208 | OpDeployFree, a free deploy from a discard pile during resolution (wave F shape, sonnet) | done |
| T-0209 | A stat criterion in filters, "with power N or less" (wave F shape, sonnet) | done |
| T-0210 | Reaver's Rush (wave F deck card, flash; #140 (4)) | done |
| T-0211 | OpMorphScratch, turning a battlefield unit face down in place (wave F shape, sonnet) | done |
| T-0212 | Forced Mutation (wave F deck card, flash) | done |
| T-0213 | Deploy goes through declare, so a deployed card's additional costs are paid (wave F shape, sonnet; #128) | done |
