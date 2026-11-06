Status: applied: no FRD changes, 2026-11-05
Directive: #122
Rulings: none

### 🎲 Design Outcome: Close Open Question 18 via ProgressPolicy Unstarving (2026-11-05)

- **Question:** How should `cardtown sim` reach rare shapes like activated abilities and deck searches without building a redundant scenario-seeding harness or turning `ProgressPolicy` into a rat's nest of per-wave rules? ([`docs/architecture/m2-ability-model.md`](../../architecture/m2-ability-model.md) Open Question 18, from structural review 8 finding 15)
- **Rulings:** none (simulation harness and bot policy only; no FRD 01 rule changes).
- **Decisions (user, 2026-11-05):**
  - Keep deterministic, step-by-step scenario testing in `pkg/engine` (`scenario_test.go`) rather than adding deck-stacking, mid-game state seeding, or scripted-move files to `cardtown sim` or `MatchStart`.
  - Fix the 100%-waterfall starvation in `ProgressPolicy` (`pkg/sim/policy.go`) once, using general action buckets rather than card- or wave-specific priority rules, so unplayable cards in hand do not force the bot to tap its entire board for mana into End Phase 4c and starve `ActActivate`, `ActAttack`, and `ActMarch`.
  - Close Design Note Open Question 18.
- **Cards affected:** None (card text and rules unchanged; increases `cardtown sim` fuzz coverage of *Cliff of Portents*, *Sinuous Growth*, *Boneyard Scavenger*, *Ruin Partisans*, and future Wave F cards).
- **Alternatives rejected:**
  - *Deck-stacking / scenario-seeding in `cardtown sim` and `MatchStart`:* rejected as redundant with `pkg/engine/scenario_test.go` (`TestFR_*` and `TestCard_*`), while 1-copy Mythics are already drawn in ~70 of 200 starter games (~130+ in stress games).
  - *Scripted `"player X does Y"` move files for bots:* rejected because `scenario_test.go` and `eventlog.Replay` already cover deterministic scripts without Turn 1 resource-ramp ceremony.
  - *Low-tier LLM persona driving `sim` moves:* rejected due to ~65,000 decisions per 200-game run (prohibitive token spend/latency and non-determinism breaking `sim-compare` and `replay`).

#### → `pm-frd-drafter` (FRD 01 and the ruling register)
- **Superseded wording:** None.
- **FRD 01 changes:** None.
- **Register rows:** None.
- **Card text:** None.

#### → `lead-architect` (build board directive)
- **Directive row:**
  | 122 | **Unstarve `ProgressPolicy` without scenario-seeding (design note Open Question 18).** **After Wave D.1 (#121), before Wave E; a `harness` task.** Keep deterministic scenarios in `pkg/engine/scenario_test.go`; `pkg/engine` and `MatchStart` gain no setup or deck-stacking fields. Fix the 100%-waterfall starvation in `ProgressPolicy` ([policy.go:47-91](../../../pkg/sim/policy.go)) with general rules that need no per-wave maintenance:<br>(a) Treat `ActActivate` as a `play` (Tier 1, alongside `ActDeploy` and `ActEnact`), and include affordable/reachable on-board `ActActivate` resource costs (and discard-pile deploy costs when Wave F adds them) when computing `needReq` and `needTac`.<br>(b) Stop `closers` (`ActGenerate` and `ActConvert`) from 100%-blocking `acts` when holding an unaffordable card: only populate `closers` toward a needed `(req, tac)` cost when the active player's current pool plus readied battlefield yields can actually reach at least one needed cost this turn, and do not exhaust an asset for mana when its own `ActActivate` is the only needed cost it could fund (or soften the `closers` vs. `acts` wall so no legal action in `LegalActions` is starved).<br>(c) Close Open Question 18 in `docs/architecture/m2-ability-model.md`, citing this outcome.<br>**Acceptance:** a `policy_test.go` test pins that an unaffordable high-cost hand card does not force readied units to tap for mana instead of attacking/activating, and that a readied asset with an `R`/`T` activated ability budgets and activates; `cardtown sim -deck starter -p0 ANIMUS -p1 PATH` runs clean with activations and deck searches (*Cliff of Portents*) firing regularly; `go tool task check` passes. | User via game-designer, 2026-11-05 ([outcome](../../../docs/design/outcomes/2026-11-05-oq18-sim-policy-without-scenario-seeding.md)) | open: after Wave D.1 (#121), before Wave E |

#### → `pm-lead` (portfolio)
- **Milestone or scope impact:** None.

---

#### Revision (2026-11-05, user via game-designer, from the workflow-steward review)

The body above is frozen. Where they differ, these corrections replace it. Directive #122 on `.agents/tasks/m2/BOARD.md` carries the same corrections.

1. **Measured figures replace the estimates.** All figures were measured on 2026-11-05 with `go tool task sim-events -- <scratch>/out.jsonl` (`cardtown sim -deck starter -p0 ANIMUS -p1 PATH`, 200 games, default `-seed 1`, `ProgressPolicy`), then counted over the JSONL file. A card counts as having left its owner's deck when a `turn-start` snapshot no longer lists it in that deck. Deploys, resource taps and activations come from the `deploy`, `generate` and `action`/`Activate` events, matched to the asset.
   - *Cliff of Portents* left its owner's deck in **99** of 200 games, not "~70". It was deployed in **80**, tapped for resources **338** times, and activated **once**.
   - The run had **99** activations in total: *Sinuous Growth* 96, *Boneyard Scavenger* 2, *Cliff of Portents* 1.
   - A run makes **49,665** actions plus **4,654** decision records, **54,319** picks (about 54,000), not "~65,000". The rejection of the low-tier LLM persona still holds at that figure.
   - **Withdrawn:** "~130+ in stress games". The stress sim (and the stress half of `sim-compare`) runs SOULLESS vs CONCORDIA. It holds no Animus or Path cards, so it never draws *Cliff of Portents*.
   - **Cards affected, corrected:** *Ruin Partisans* is in neither starter deck, so its starter-sim coverage does not improve. Drop it from the list. The cards whose starter-sim coverage improves are *Cliff of Portents*, *Sinuous Growth* and *Boneyard Scavenger* (all in a starter deck), plus future Wave F cards.
2. **Tests the directive contradicts** (`pkg/sim/policy_test.go`, read against the scenes in `pkg/sim/helpers_test.go` and `canYield` in `pkg/engine/actions.go`):
   - **`TestProgress_ActivationIsAnAct`:** #122 (a) reverses its "a play from hand comes first" check (the `scene(true)` loop). That check wants `ActDeploy` on all 16 seeds when a 0-cost unit and an affordable activation are both legal. Under (a) they share one uniform tier, so some seeds pick `ActActivate`. Its `scene(false)` check (Activate or EndTurn) still holds.
   - **`TestProgress_Case2_Preferences` and `TestProgress_Case6_EnactsManeuvers`, under (b)'s reachability rule: no change.** Neither scene has an activated ability, so (a) does not touch them either. Case2's second scene (a 2R card, pool 0) reaches 2 Requisition from the HQ (yield R1) plus the Worker in Reserves (R1). Case6's Case 3 (a 2T Maneuver, pool 0/0) reaches 2 Tactics from the HQ (T2) alone, because a unit in Reserves yields only Requisition (FR-CAP-001). So Generate stays the expected pick. **Both scenes sit exactly at reach = cost**, so the rule must count the HQ's yield, respect the zone limit on unit yields, and use reach ≥ cost. A stricter rule fails both tests.
   - **The "soften the wall" alternative in (b) does contradict them.** In both scenes the readied Worker's March is legal, so `acts` is not empty. Each test wants Generate on all 8 seeds. If a softening lets an act win while a reachable closer exists, it fails Case2's second scene and Case6's Case 3. The softening keeps both tests valid only if closers keep priority whenever the needed cost is reachable this turn. Otherwise the task reverses those checks.
3. **Win-rate baseline shift: accepted.** The current baseline, measured with the run in item 1, is P0 84.0%, P1 15.0%, draws 1.0%, first player 46.5%, average 30.6 turns. Decision (user, 2026-11-05): "Yes, I accept the baseline shift; sim-compare is expected to differ for this task." The #122 task's `sim-compare` difference is expected and does not block review.
4. **Open Question 18 is closed for now (provisional).** The user approved closing it "for now". Seeding or scripted scenario files come back if the policy fix still leaves shapes unreached. **What reopens it** (user, 2026-11-05): "After the #122 task merges, a starter sim run (go tool task sim-events) still shows an activated ability or deck-search shape in the starter decks firing fewer than ~10 times in 200 games." The pre-fix baseline is in item 1: *Boneyard Scavenger* 2 and *Cliff of Portents* 1 (its deck search) are below that line, and *Sinuous Growth* 96 is above it. Item (c) of #122 records the closure in the design note as provisional, with this trigger.
5. **(a) vs. "soften the wall": soften preferred.** If every activation is as eager as a play, *Sinuous Growth*, *Boneyard Scavenger* and *Ruin Partisans* (each with a "kill this unit" cost) kill themselves whenever their activation is affordable. Decision (user, 2026-11-05): "Prefer the 'soften the wall' alternative in (b): Activate stays in acts but closers no longer fully block acts; (a) becomes optional." If the task takes (a) anyway, it must reverse the `TestProgress_ActivationIsAnAct` check named in item 2. The softening must respect the constraint on Case2 and Case6 in item 2.
