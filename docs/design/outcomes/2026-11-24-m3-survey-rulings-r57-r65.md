Status: approved (user, 2026-11-24)
Directive: #185, #186, #187, #188, #189
Rulings: SVY-R57, SVY-R58, SVY-R59, SVY-R60, SVY-R61, SVY-R62, SVY-R63, SVY-R64, SVY-R65

### 🎲 Design Outcome: M3 survey rulings R57 to R65 (2026-11-24)

- **Question:** The nine open ruling questions of the M3 ability survey (`docs/design/survey/m3_full_catalog_ability_survey.md` §4, R57 to R65).
- **Rulings:**

  | Ruling | Settles | Cards |
  | :--- | :--- | :--- |
  | `SVY-R57` | A target needs only the criteria its text prints. Regroup may choose a Readied asset and does nothing to it. | Tempest Runner, Enclave, Armored Charge, Visionary's Insight; Tempest City Ray |
  | `SVY-R58` | All-Out War adds power. | All-Out War (reworded) |
  | `SVY-R59` | A face-down Scratch Company costs 0 Requisition and 0 Tactics. | Will Breaker |
  | `SVY-R60` | A look with too few cards fills its destinations in printed order. | Hourglass Diviner, Samsara |
  | `SVY-R61` | The card put into hand is any of the looked-at cards not deployed. | Light of the World Soul (reworded) |
  | `SVY-R62` | "The Main Theater or Reserves" is chosen on declaration with the target. Either zone is legal. A timeout takes Reserves. | Sandstorm Lightbringer |
  | `SVY-R63` | "You may discard N. If you do": all or nothing, and not asked with fewer than N cards. | Redlined Brawler |
  | `SVY-R64` | NO-AH-5 can't choose itself. | NO-AH-5 (reworded) |
  | `SVY-R65` | A unit moved to Reserves goes to its controller's Reserves. | Awaken Nightmare, Tempest City Ray (reworded) |

- **Decisions (user, 2026-11-24):**
  - R57: "yep, regrouping a readied unit is still a valid game action, it just does nothing. If something specifically requires a target to be exhausted, it'd say "chosen exhausted asset"."
  - R58: "agreed, old wording that needs updating."
  - R59: "yep, scratches have a cost of 0/0. Add an explicit mention of it."
  - R60: "hand, this follows MTG's behavior of "do as many things as you can in order" and the first thing of the two part effect is putting one into your hand. See Sleight of Hand."
  - R61: "any of the remaining N cards is the intent. Probably needs a reword."
  - R62: "lead's choice is fine, and doesn't require rewording the card or more intricate default behavior."
  - R63: "yep, you need to have 2 cards to discard in order to pay for it."
  - R64: "yeah, reword. Eating itself is dumb."
  - R65: "reword the two cards as described, yep"
  - The whole report: "approved". It was shown with these points flagged as going past the answers above: `SVY-R57` covers every target, not only Regroup. `SVY-R59` covers a unit turned face down. `SVY-R60` covers looks only. `SVY-R63`'s discard stays an effect on resolution, its "not asked" covers only discards from hand, and *Dolus Refiner* discards what it can. `SVY-R65` covers any card that moves a unit to Reserves.
- **Cards affected:** all in M3 scope. All are in a starter deck except Awaken Nightmare.
- **Alternatives rejected:**
  - R59, no cost, or a "non-Scratch" rider on Will Breaker: 0/0 is the rule.
  - R60, the player chooses the destination: printed order.
  - R62, always the other zone: needs a reword and more intricate default behavior.
  - R65, the owner's Reserves: the move would also have to change control.

#### → `pm-frd-drafter` (FRD 01 and the ruling register)
- **Superseded wording:**
  - `docs/frd/01_Game_Rules_FRD.md:638` (`FR-KWD-002n`): "chooses up to $N$ **Exhausted** allied assets (or all matching exhausted allied assets …)".
  - `:283` (`FR-ENT-102`): quotes NO-AH-5's old text.
  - Left as it is: `docs/frd/reference/02_Benchmark_Card_Blueprints.md:24` (old NO-AH-5 text in an FRD 02 reference).
- **FRD 01 changes** (v3.31 → v3.32):
  - `FR-KWD-002n`: *"One-shot effect action. Each asset it names changes from **Exhausted** to **Readied**. An asset that is already Readied may be chosen and stays Readied (`SVY-R57`). Text that needs an Exhausted target says so ("chosen exhausted allied asset", *Visionary's Insight*)."*
  - `FR-TGT-001`, new bullet after Conjunction Eligibility: *"**Only Stated Criteria Count (`SVY-R57`)**: An entity that meets every stated criterion is a legal target even if the effect would not change it. *Tempest Runner* may choose a Readied unit, and *Ensnaring Mist* a unit already in its controller's Reserves. That instruction does nothing to it, and the rest of the effect applies: *Armored Charge* on a Readied unit still heals it and gives it armor. Targeting restrictions such as Ward (`FR-KWD-002h`) and `FR-ZON-002`'s movement immunity still apply."*
  - `FR-ZON-005` (Double-Blind Deployment), after the yields: *"Its Requisition cost and Tactics cost are both 0 (`SVY-R59`), so *Will Breaker* ("Requisition cost 3 or less") may choose it."* `FR-ENT-103`: the transformed unit's costs are also 0 and 0.
  - `FR-MOD-001` §4: *"**Too Few Cards (`SVY-R60`)**: When a look at the top cards finds fewer cards than its required destinations need, the destinations fill in printed order. With one card left, *Hourglass Diviner* and *Samsara* put it into hand. Samsara then puts no card into the discard pile, so it deals no damage."* And: *"*Light of the World Soul*'s card for the hand is any of the looked-at cards not deployed (`SVY-R61`)."*
  - `FR-TGT-001`, after "Another" Targets: *"**Choosing a Destination Zone (`SVY-R62`)**: When an effect moves a chosen unit to one of two zones (*Sandstorm Lightbringer*), the zone is chosen on placement with the target (`SVY-R3`). Either zone may be chosen. A unit already there does not move."*
  - `FR-MOD-001` §6, beside Free-Deploy Zone: *"*Destination Zone (`SVY-R62`)*: a choice between Reserves and the Main Theater takes Reserves."*
  - `FR-TGT-001` (`SVY-R14` bullet), append: *"An X of "discard N cards" is carried out in full or not at all (`SVY-R63`): with fewer than N cards in hand on resolution, nothing is discarded and Y does nothing. X is still not a cost. A mandatory discard that is no such X discards as many cards as it can (*Dolus Refiner*)."*
  - `FR-MOD-001` §7 (`SVY-R48`), append: *"Nor is it asked, or its trigger placed, when its X is "discard N cards" and its controller has fewer than N cards in hand as it would be placed (`SVY-R63`)."*
  - `FR-ENT-102`: NO-AH-5's quote becomes the new text (`SVY-R64`).
  - `FR-ZON-002` (Unit Battlefield Movement), append: *"An effect that moves a unit to Reserves moves it to its controller's Reserves, whoever owns it (`SVY-R65`)."*
  - Revision summary `(v3.31 -> v3.32): M3 Survey Rulings (2026-11-24)`, one line per ruling as in the table above, naming its requirements. `SVY-R58`'s line names only the card text.
- **Register rows** (M3 survey):
  - §4, status cells R57 to R65: **"Ruled (user, 2026-11-24).** [its "Settles" text above] Recorded in FRD 01 v3.32 as `SVY-R<n>`."
  - §6 card rows: cite `SVY-R<n>` in place of each "R57" to "R65". NO-AH-5's definition reads "kill another chosen …", and Tempest City Ray's "→ its controller's Reserves".
  - Keyword matrix: no row changes.
- **Card text:**
  - *All-Out War*, bullet 1. Old: "- Chosen allied asset gets power equal to the number of cards discarded this way until end of turn." New: "- Chosen allied asset gets +X power until end of turn, where X is the number of cards discarded this way."
  - *All-Out War*, bullet 2. Old: "- Scratch X, where X is the number of cards discarded this way." New: "- Scratch X."
  - *Light of the World Soul*. Old: "put up to 1 of those cards into your hand". New: "put up to 1 of the other cards into your hand".
  - *NO-AH-5*. Old: "R2: Kill chosen allied non-Headquarters asset and tuck it face down under this unit." New: "R2: Kill another chosen allied non-Headquarters asset and tuck it face down under this unit."
  - *Awaken Nightmare*. Old: "to its owner's Reserves". New: "to its controller's Reserves".
  - *Tempest City Ray*. Old: "T2, exhaust: Move chosen opposing unit to Reserves." New: "T2, exhaust: Move chosen opposing unit to its controller's Reserves."
- **Prompt to run:** `@pm-frd-drafter apply Design Outcome docs/design/outcomes/2026-11-24-m3-survey-rulings-r57-r65.md`

#### → `lead-architect` (build board directives)
Each row is in [`.agents/tasks/m3/BOARD.md`](../../../.agents/tasks/m3/BOARD.md), with status `waiting: docs/design/outcomes/2026-11-24-m3-survey-rulings-r57-r65.md applied (then open: before <below>)`.
- **#185, wave A (`SVY-R59`, `SVY-R65`)**, before the first wave A card task.
  (1) Update survey §1 to §3 for all nine rulings: the open counts, §2's "Rulings first" column, and G5, G13 and G14.
  (2) Kernel: no behavior change. `assetStat` and `scratchFace` already give 0. `scratchFace`'s doc comment cites `SVY-R59`.
  (3) *Will Breaker* card tests: it may choose a Scratch Company, it must when that is the only legal target, and it may choose a unit transformed face down whose printed cost is over 3.
  (4) `stepMoveZone` already uses the controller's Reserves. *Tempest City Ray* and *Awaken Nightmare* definitions use the new text. A card test moves a unit controlled through *Rhapsody's Hollow* to its controller's Reserves, where that is legal.
- **#186, wave B (`SVY-R57`, `SVY-R58`)**, before the first wave B shape task.
  (1) G4's readying step takes no readiness filter and skips Readied assets. G5's filter is only for printed criteria.
  (2) Tests: `TestFR_KWD_002n_ReadiedAssetMayBeChosenAndStaysReadied`, and an *Armored Charge* card test on a Readied unit.
  (3) *All-Out War* adds power (Layer 5).
  (4) Not a ruling: two *Tempest Runners* can ready each other endlessly. Check that the sim policies and the turn clock end such a turn.
- **#187, wave C (`SVY-R62`, `SVY-R63`)**, before the *Sandstorm Lightbringer* and *Redlined Brawler* tasks.
  (1) G14: the zone is chosen on placement with the target, both zones are always offered, and a timeout takes Reserves. Choosing the current zone moves nothing and logs no move.
  (2) G13: a discard X is all or nothing, and `couldDoNothing` (`pkg/engine/placement.go:264`) also covers a hand with fewer than N cards.
  (3) Authorized edit: the "an empty hand" case of `TestFR_TRG_005_IfYouDoFollowsADiscardCount` (`pkg/engine/if_you_do_test.go:320`) scripts a "yes" that is no longer asked. It should assert that nothing is asked or placed.
  (4) Tests: one card in hand asks nothing. A hand that shrinks before resolution discards nothing. A timed-out discard choice discards nothing. A mandatory discard 2 with one card discards it.
- **#188, wave D (`SVY-R60`, `SVY-R61`)**, before wave D's decision record.
  (1) G15 fills destinations in printed order.
  (2) Tests: a one-card look puts it into hand for both cards, and *Samsara* then deals nothing. `TestFR_MOD_001_LookTopIsHiddenSoNoFastPath` still applies, so a one-card look is still shown.
  (3) The hand pick offers every card not deployed, units included.
- **#189, wave E (`SVY-R64`)**, before the *NO-AH-5* shape tasks.
  Its slot excludes its source (`SVY-R34`). Tests: it can't choose itself. A second *NO-AH-5* is legal, and the cards tucked under it go to their owners' discard piles.
- **Checked for contradictions:** test names and bodies under `pkg/` for moves, Requisition criteria, looks and "if you do". Only #187 (3) needs an edit. `TestCard_EnsnaringMist_UnitAlreadyInReservesStays` and `TestFR_MOD_001_LookTopShortDeckLooksAtWhatIsThere` agree. The reworded cards have no definitions yet. The new card text passes `cardtown lint --all` and `go test ./...` (scratch worktree, 2026-11-24). `cardtown sim` output doesn't change.

#### → `pm-lead` (portfolio)
- **Milestone or scope impact:** None.
