---
id: T-0264
title: "A kill cost with a choice of counts: Cost.Counts on CostKill, the count chosen on declaration among the payable ones and paid in full by one kill (wave H shape, sonnet)"
kind: shape
tier: sonnet
wave: H
status: tests-review
branch: m2/autoloop-wave-h/T-0264-kill-count-cost
depends_on: []
claims: [shape:counted kill cost (Cost.Counts on CostKill: "kill 2 or 4 allied X" as an additional cost, the count chosen on declaration from the counts that can be paid, placement.count, and the counted kills made by one kill call)]
cards: []
requirements: [FR-TRN-003a, FR-TRN-003, FR-MNV-001, FR-MOD-001, FR-REP-001, SVY-R2, SVY-R15, SVY-R16, SVY-R29, SVY-R53, SVY-R54]
---

## Goal
A Maneuver's additional kill cost can offer a choice of counts: *"As an additional cost to enact this maneuver, kill 2 or 4 allied Scratch Company units."* (*Mutation Swell*, directive #158 (1)). `ability.Cost` gains `Counts []int`, read only with `CostKill`: nil keeps today's "kill one", and a non-nil list is the printed counts in printed order. On declaration, before the targets, the payer chooses one count among those that can be paid now (at least that many assets match the cost's filter), and after the push the payer picks exactly that many different matching assets, which are killed together (`SVY-R53`, `FR-TRN-003a`). The card can be enacted only if some count can be paid (`SVY-R29`). A timeout takes the first listed count that can be paid and the first assets in fixed order (`SVY-R54`, `FR-MOD-001` §6). A replaced kill still pays (`SVY-R15`).

This is directive #158 part (1). The targets "for every 2 units killed this way, … a different chosen unit card" (#158 (2)) are T-0265, queued, which also narrows the counts offered to those whose targets can be filled. *Mutation Swell* itself (#158 (3), (4)) is T-0266, queued. Both are split off under the D2 size rule: the three together touch `ability.go`, `costs.go`, `placement.go`, `targeting.go`, `doc.go` and the two card files.

## Inputs
- FRD 01 (`docs/frd/01_Game_Rules_FRD.md`, v3.29): revision summary (`:17-19`); *A Maneuver Needs Its Required Targets* (`:378`); `FR-TRN-003a` *Additional Sacrifice, Discard & Exhaust Costs* (`:379`: a cost may offer a choice of counts, chosen on declaration with the cost and the targets, paid in full); *Mutation Swell*'s sub-bullet (`:401`); `FR-MOD-001` §6 *Choice of Cost Counts* (`:442`) and *Mandatory Targets and Untargeted Selections* (`:441`, the fixed order); *Replaced Actions Still Count as Done* (`:534`, `SVY-R15`, with the Swell example).
- Survey (`docs/design/m2_starter_set_ability_survey.md`): R53 (`:112`), R54 (`:113`), *Mutation Swell*'s row (`:169`).
- Outcome `docs/design/outcomes/2026-11-16-wave-h-mutation-swell-kill-count.md`; board directive #158.
- Design note `docs/architecture/m2-ability-model.md`: `Def.ExtraCosts` (`:31`), Deliverable 6m's wave F bullet (`:210`), Open Question 25 (answered).
- Code of the day (run branch at 5ba2588):
  - `pkg/engine/ability/ability.go`: `CostKind` (`:131-140`), `Cost` (`:142-154`).
  - `pkg/engine/costs.go`: `costDecision` (`:49-68`, the one builder of a kill or exhaust cost's pick, for legality and payment), `chooseCostDiscards` (`:70-89`, the discard cost's picks made before the targets), `costsPayable` (`:91-102`), `otherCostsPayable` (`:160-182`: a kill cost needs one option), `payCosts` (`:184-232`: one `g.kill` per picked asset).
  - `pkg/engine/placement.go`: `placement` (`:16-25`, `discards`), `declare` (`:44-55`: discards, then modes and targets, push, pay), `placeable` (`:176-191`).
  - `pkg/engine/choice.go`: `DecideMode` (`:15`, fallback first listed), `soleAnswer` (`:90-101`).
  - `pkg/engine/doc.go`: the Maneuver bullet (`:56-57`, additional costs).
  - Searched `pkg/engine` (non-test) and the task archive for a cost or pick with a choice of counts: none. `CostDiscard`'s `UpTo` is "up to N", not a choice of exact counts; `Step.UpTo` and `Target.UpTo` are "up to N" too.

## Shape (the lead's plan)
- `pkg/engine/ability/ability.go`:
  - `Cost` gains `Counts []int`, read only with `CostKill`: nil kills one asset (today's cost); otherwise the payer kills N different assets matching `Filter`, N chosen on declaration from `Counts` (printed order, each > 0) among the counts that can be paid (`SVY-R53`). Update `CostKill`'s line and `Cost`'s comment ("no other cost reads it", as for `UpTo`).
- `pkg/engine/costs.go`:
  - `payableCounts(pl placement) []int` (name is the engineer's): for pl's counted kill cost, the counts in `Counts` order with at least that many options in `costDecision`; nil without a counted kill cost. It is the one judge of which counts can be paid; T-0265 narrows it by the targets.
  - `chooseCostCount(pl placement, counts []int) int`: asks pl's controller a `DecideMode` (the §6 fallback "first listed" is `SVY-R54`'s, and §1's single option is taken unasked) whose options are `counts`, labelled `kill N`, `Ref` 0, `Min` = `Max` = 1, and returns the count chosen; 0 when `counts` is empty, asking nothing. The caller passes the counts to offer, so T-0265 can narrow them without changing it. A placement has at most one counted kill cost (panic on two, as `chooseCostDiscards` does for discard costs).
  - `otherCostsPayable`: a counted kill cost is payable when `payableCounts` is non-empty; a kill cost without `Counts` still needs one option.
  - `costDecision`: for a counted kill cost with `pl.count > 0`, `Min` = `Max` = `pl.count` (different assets, no repeats), and the prompt names the count. Legality calls it with `pl.count` 0 and reads only its options.
  - `payCosts`: a kill cost's picked assets are killed by one `g.kill(picked)` call, so a counted cost's kills happen together, as a kill step's do (`FR-REP-001`); for a cost without `Counts` that is the one asset, as now. Exhaust stays one call per asset.
- `pkg/engine/placement.go`:
  - `placement` gains `count int`: the count its counted kill cost pays, chosen on declaration (`SVY-R53`); 0 without one.
  - `declare`: after `chooseCostDiscards`, `pl.count = g.chooseCostCount(pl, g.payableCounts(pl))`, before the modes and targets, so T-0265's per-kill slots know it. `placeable` needs no new term (`costsPayable` reads `otherCostsPayable`).
- `pkg/engine/doc.go`: the Maneuver bullet adds that a kill cost may offer a choice of counts (`Cost.Counts`), chosen on declaration before the targets and paid in full (`SVY-R53`).
- No new `DecisionKind`: the count is a choice between printed alternatives whose timeout is "the first listed", which is `DecideMode`'s.
- Not this task: target slots that repeat per kill (T-0265), *Mutation Swell* (T-0266), counted exhaust or discard costs (no M2 card has one).

## Tests (shape tasks: written by rules-test-author; engineers never edit these)
New tests go in `pkg/engine/kill_count_cost_test.go`. Maneuvers are synthetic: `maneuver(...)` (as `killCostManeuver` in `kill_cost_test.go` builds one) with `c.Ability.ExtraCosts = []ability.Cost{{Kind: ability.CostKill, Filter: f, Counts: []int{2, 4}}}`, where f is allied Scratch Company units (`ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit, Scratch: ability.ScratchOnly}`), and an effect of one draw unless a test says otherwise. Allied Scratch Companies come from `s.putScratch(0)`. Player 0 is the active player and enacts. Through a local chooser that records every Decision, as `kcScript` does.

- `TestFR_TRN_003a_KillCountCostNeedsTheLowestCount`: with one allied Scratch Company the Maneuver isn't among player 0's legal actions; a non-Scratch allied unit and player 1's Scratch Companies don't count toward the count (still not legal with one allied Scratch Company, one allied non-Scratch unit and two of player 1's Scratch Companies). With two allied Scratch Companies it is legal.
- `TestFR_TRN_003a_KillCountCostOffersOnlyPayableCounts`: three allied Scratch Companies: the Maneuver is legal, no count decision reaches the chooser (count 2 is the only payable one, §1), the kill pick is a `DecidePick` with the three as options and `Min` = `Max` = 2; the two picked die, the third lives, and the Maneuver resolves (player 0 draws one).
- `TestFR_TRN_003a_KillCountCostChoosesTheCount`: four allied Scratch Companies: a `DecideMode` count decision reaches the chooser before any target or kill pick, with two options in printed order (2 first) and `Min` = `Max` = 1. Answering 4: all four die with no kill pick asked (§1). Answering 2 (a fresh scene): the kill pick has four options, `Min` = `Max` = 2, and only the two picked die.
- `TestFR_MOD_001_KillCountCostTimeout`: four allied Scratch Companies, two in Reserves and two in the Main Theater, and a chooser that times out on everything: the count is 2 and the two in Reserves die (the fixed order, `FR-MOD-001` §6 at `:441`), the two in the Main Theater live.
- `TestFR_TRN_003a_KillCountCostPaidAfterTargets`: the Maneuver also has an exactly-one target slot on an allied unit card in the discard pile (`ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit, Zone: ability.ZoneDiscard}`) and an empty step list; player 0's discard pile holds one unit card X; two allied Scratch Companies. The target decision offers only X (the cost-killed units' cards aren't in the pile yet, `SVY-R2`), and after the Maneuver resolves both killed units' cards are in player 0's discard pile.
- `TestFR_TRN_003a_KillCountCostTriggersGoAboveTheManeuver`: as `TestFR_TRN_003a_CostTriggersResolveBeforeTheManeuver` does for one kill, with count 2: each cost kill's death trigger is on the stack above the Maneuver and resolves before it (`FR-TRN-003a`, `FR-TRN-003`).
- `TestFR_TRN_003a_ReplacedKillCountCostStaysPaid`: as `TestFR_TRN_003a_ReplacedKillCostStaysPaid` does for one kill, with count 2 and a destruction replacement that applies to both picked units: neither dies, the cost counts as paid and the Maneuver resolves (`SVY-R15`, `SVY-R21`).

Existing tests stay as they are and must still pass (`kill_cost_test.go` pins the cost without `Counts`).

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- none expected. Helpers local to the new file live in it.

## Allowed Files
- `pkg/engine/ability/ability.go`
- `pkg/engine/costs.go`
- `pkg/engine/placement.go`
- `pkg/engine/doc.go`

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run "KillCountCost"`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-compare -- autoloop/m2-wave-h` passes (no starter or stress card prints `Counts`, and a one-asset kill through one `g.kill` call is the same kill, so every log is byte-identical)

## Log
- 2026-11-17 lead-architect (wave H run, session 26): written against the run branch at 5ba2588, after syncing `main` (aa105b0 added #158; 0fbc984 applied its outcome, FRD 01 v3.29), so #158 is `open: before T-0264`. *Mutation Swell* is next in Deliverable 6m's order (wave F bullet, after T-0263's *Ancestral Monument*). No ruling needed: `SVY-R53` makes the kills an additional cost chosen on declaration and paid in full, and `SVY-R54` gives the timeout. Reuse: `costDecision` stays the one builder of a kill cost's pick, for legality and payment, `payableCounts` the one judge of the counts, and `DecideMode`'s "first listed" fallback is the count's; no new decision kind. Split under the D2 size rule: T-0265 (the per-kill "different" targets, #158 (2)) and T-0266 (*Mutation Swell*, #158 (3), (4)) queued in this commit. #158 (5) applied in this commit (design note Open Question 25 answered, Deliverable 6m's line 210 rewritten). Four non-test files. Directives: #158 `open: before T-0264`, applied by T-0264 to T-0266; each `waiting` one was checked against the repo and none has fired (#6, #7: mode is `serial`; #15 (2): no draw or end-of-turn hook; #71: no each-player effect; #91 (2): no cost on a mode; #106: no deck search; #120: no mandatory discard; #142: no tuck). Status `tests`.
- 2026-11-17 rules-test-author: tests written in `pkg/engine/kill_count_cost_test.go`: `TestFR_TRN_003a_KillCountCostNeedsTheLowestCount`, `TestFR_TRN_003a_KillCountCostOffersOnlyPayableCounts`, `TestFR_TRN_003a_KillCountCostChoosesTheCount` (two subtests: count 4, count 2), `TestFR_MOD_001_KillCountCostTimeout`, `TestFR_TRN_003a_KillCountCostPaidAfterTargets`, `TestFR_TRN_003a_KillCountCostTriggersGoAboveTheManeuver`, `TestFR_TRN_003a_ReplacedKillCountCostStaysPaid`. Helpers requested: none (local helpers in the file). `go vet` fails only on the missing `ability.Cost.Counts` field, which the engineer adds; `cardtown lint rules` passes. Notes: the trigger test uses `Counts: 2` with non-Scratch units (a Scratch Company has no death trigger), and the replaced test uses the real *Shifting Impossibility* (2 damage after two replacements). The paid-after-targets test tolerates a sole target being taken unasked, and asserts that any target decision offers only X. Nothing left out. Status `tests-review`.
