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
