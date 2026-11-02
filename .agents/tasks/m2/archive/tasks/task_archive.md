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
