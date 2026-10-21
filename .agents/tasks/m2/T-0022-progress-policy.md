---
id: T-0022
title: A random sim policy that plays expensive cards
kind: harness
tier: sonnet
wave: none (board directive #21, ahead of the next A2 task)
status: done
branch: m2/T-0022-progress-policy
depends_on: [T-0021]
claims: [func:ProgressPolicy, field:SimOptions.Policy, cli:sim-policy]
cards: []
requirements: []
---

## Goal
`cardtown sim` plays its random games with a policy that makes progress. It deploys what it can afford, generates resources when that lets it afford a card in hand, and mostly ends its turn only when nothing productive is left. It's still random, deterministic for a seed, and legal-only. Every executable starter-deck card then gets played often enough that `sim` actually exercises its abilities. `RandomPolicy` stays as it is.

## Inputs
- Board directive #21 (`.agents/tasks/m2/BOARD.md`).
- The finding from the T-0021 review: in 200 starter games, `RandomPolicy` deployed Saboteur 5 times, Stormfront of Ideals 5 and Unceasing Tide 13, against 270 Portent Scriveners. It picks uniformly among non-EndTurn actions and ends the turn with probability `endBias` (0.1 in `sim`).
- `Policy`, `RandomPolicy` and `Playout` are in `pkg/engine/playout.go`. `LegalActions` (`actions.go`) lists only legal actions, so a listed Deploy is already affordable and supported, and EndTurn is always last. `SimOptions` and `SimulateWith` are in `sim.go`, and `runSim` in `cmd/cardtown/main.go`.
- Pools empty at end of turn (sub-phase 4c), so unspent resources don't carry over.
- Rulings that apply: none. A policy picks among legal actions only, and changes no rule.

### Shapes
```go
// pkg/engine/playout.go
// ProgressPolicy is a random policy that favors progress over passing. It
// picks only from legal and is deterministic for seed.
func ProgressPolicy(seed uint64) Policy

// pkg/engine/sim.go: SimOptions gains
	Policy func(seed uint64) Policy // nil means RandomPolicy(seed, 0.1), today's behavior
```

### Lead decisions (technical)
- **Order of preference each call:**
  1. If any Deploy is legal, pick one at random (card and zone).
  2. Otherwise, if a supported, deployable card in hand costs more than the current pool, and a legal Generate or Convert would close the gap in the missing resource, pick one at random among those. "Close the gap" means it yields the resource that is short. Convert counts when Tactics is short and Requisition is at least 2.
  3. Otherwise, pick at random among Attack, March and Scratch with probability 0.9, and EndTurn with probability 0.1.
  4. With nothing left but EndTurn, end the turn.
  Keep the randomness inside each step. The exact weights and tie-breaks are the engineer's choice and are recorded in the Log.
- **No card names and no card-specific rules.** The policy reads only `LegalActions`, costs, pools and yields. It lives in package `engine`, so the fences lint applies.
- **Termination:** each preferred action spends something (a cost, an exhaust, or 2 Requisition), so a turn can't loop forever. `Playout`'s `maxTurns` still guards the match.
- **`cardtown sim -policy random|progress`:** defaults to `progress`. `random` reproduces today's runs exactly. `-log` and `-events` work with both.
- **`Simulate`** keeps its behavior (`RandomPolicy`). Only `SimulateWith` with `Policy` set, and the CLI default, use the new one. `RandomPolicy` doesn't change.
- **Not in this task:** smarter targeting or choice answers (choosers stay `RandomChooser`), card evaluation, attack selection heuristics, and "play" agents for `cardtown play`.

## Acceptance Cases (harness tasks only; written by the lead)
Cases 1 to 5 go in a new `pkg/engine/policy_test.go`, committed first as `test(T-0022): …` and failing before the implementation. Cases 6 and 7 are CLI checks, run by the engineer and re-run by the reviewer. Record their output in the Log.
1. **Legal only:** across 50 starter-deck games under `ProgressPolicy`, every action it returns is one of the actions in the `legal` slice it was given. Wrap the policy to check this.
2. **Preferences, on scenes:**
   - With an affordable card in hand and a readied asset, it returns a Deploy.
   - With no affordable card, a 2-Requisition card in hand, a pool of 0 and a readied asset in Reserves, it returns a Generate for Requisition.
   - With an empty hand and no attack, march or scratch available, it returns EndTurn.
   Check each over several seeds.
3. **Deterministic:** `SimulateWith` on the starter decks, 20 games, with `Policy: ProgressPolicy`, gives equal `SimStats` and byte-identical `WriteEventsJSONL` output across two runs with one seed.
4. **Coverage (the point of the task):** in 200 starter games with `ProgressPolicy`, count `deploy` events by `Card`. Every card in either starter decklist that is `Supported()` is deployed at least 10 times (lowered from 25 at the lead's review, see Log). Record the per-card counts in the Log. If a card falls short, report the counts and stop at `blocked` rather than tuning the policy toward that card.
5. **Terminates and holds:** 200 starter games and 200 default stress-deck games under `ProgressPolicy` all finish within 400 turns with `CheckInvariants` holding (`Playout` errors otherwise).
6. **CLI default:** `go run ./cmd/cardtown sim -deck starter -p0 ANIMUS -p1 PATH` uses `progress`, exits 0 and is byte-identical across two runs. Record its summary next to the old one (the `random` baseline) in the Log.
7. **CLI `-policy random`:** `go run ./cmd/cardtown sim -deck starter -p0 ANIMUS -p1 PATH -policy random` and `go run ./cmd/cardtown sim -policy random` are byte-identical to `rework/executable-rules`' `sim` output for the same flags. An unknown `-policy` value prints `[ERROR]` and exits 1.

## Allowed Files
- `pkg/engine/playout.go`
- `pkg/engine/sim.go`
- `cmd/cardtown/main.go`
- `pkg/engine/policy_test.go` (new, the `test(T-0022)` commit)

## Done When
- [x] Cases 1 to 5 have tests in `policy_test.go`, committed before the implementation, and all pass
- [x] Cases 6 and 7 hold, with their output and the per-card deploy counts in the Log
- [x] `go tool task check` passes
- [x] Existing tests pass unchanged

## Log
- 2026-10-20 lead-architect: Wrote the task from board directive #21, with status `ready` (harness, sonnet). The user set its priority ahead of the next A2 task.
- 2026-10-20 kernel-engineer: BLOCKED on Acceptance Case 4. Tests committed first (1a3cded, failing: ProgressPolicy undefined), then ProgressPolicy, SimOptions.Policy and `sim -policy random|progress` implemented as specced. Cases 1, 2, 3, 5 pass. Case 4 fails for one card; per the process rule the test is untouched and the policy is not tuned toward it.
  - Per-card deploys, 200 starter games (seeds 1..200, progress): Chainbreaker Sneak 326, Despoiler Larva 361, Despoiler Worm 56, **Hammerhand Brigade 22 (needs 25)**, Portent Scrivener 393, Saboteur 90, Spoil Forest Amalgam 72, Spoil Forest Creeper 296, Stormfront of Ideals 155, Unceasing Tide 361. (Random baseline from T-0021 review: Saboteur 5, Stormfront 5, Unceasing Tide 13, Portent Scrivener 270.)
  - Reason: Hammerhand Brigade has 2 copies in path-starter.txt, so few are drawn; the 25 threshold is likely too high for a 2-copy card, not a policy fault. Lead decides: lower the threshold (e.g. per-copy rate), more games, or accept a policy change.
  - Policy design (exact choices): step 1 uniform over legal Deploy actions; step 2 "short" = any supported hand card with deploy zones whose ReqCost > Req (Requisition Generate with ReqYield>0 closes it) or TacCost > Tac (Tactics Generate with TacYield>0, or any legal Convert closes it), uniform over closers; step 3 uniform over Attack/March/Scratch with p=0.9 else EndTurn; step 4 EndTurn. Reads only LegalActions, hand costs, pools and printed yields. Reuse: modeled on RandomPolicy (PCG seeding, EndTurn last).
  - Case 6 (progress, `sim -deck starter -p0 ANIMUS -p1 PATH`): P0 wins 81.5% P1 15.5% draws 3.0%, first player 49.5%, avg 30.7 turns; reasons: HQ destroyed 194, both lost simultaneously 6. Old (random): P0 51.0% P1 17.0% draws 32.0%, first 35.5%, avg 27.2 turns; HQ destroyed 136, simultaneous 64. Byte-identical-across-runs not re-run by CLI diff; covered by Case 3 test.
  - Case 7: `-policy random` output byte-identical (cmp) to the pre-change binary for both `-deck starter -p0 ANIMUS -p1 PATH` and bare `sim`; `-policy x` prints `[ERROR] unknown -policy "x" ...`, exit 1.
  - `go tool task check` not run to green because policy_test Case 4 fails; no further gate run.
- 2026-10-20 lead-architect: The block is a **threshold set too high by the lead**, not a policy fault. Hammerhand Brigade (2 copies, 5 Requisition) reached 22 because Path has only 9/50 executable cards, so few Reserves units generate Requisition, and Path loses 81.5% of games. The 1-copy, 5-cost Spoil Forest Amalgam reached 72 on the Animus side. A threshold of **10** keeps the case's purpose: under `RandomPolicy`, Saboteur (5) and Stormfront (5) fail it. Every current card clears it with margin, and the case won't flake as more Path cards become executable.
  - **Authorized test edit:** in `policy_test.go`, case 4's minimum goes from 25 to 10. That constant and its message are the only change, in a new `test(T-0022)` commit on top of 785ad0e. Then run `go tool task check`, diff case 6's CLI output across two runs, and set `review`.
  - Directive #21 (a) is updated on the board to match.
- 2026-10-20 kernel-engineer: Case 4 minimum changed 25 -> 10 in policy_test.go only (comparison and message), commit f34fe2e, lead-authorized. `go tool task check` exit 0 (gofmt, vet, all tests, lint). Case 6: two runs of `sim -deck starter -p0 ANIMUS -p1 PATH` exit 0 and `cmp` byte-identical. Case 7 results as in the earlier entry. Lowest deploy count is Hammerhand Brigade at 22, above 10. Status review.
- 2026-10-20 lead-architect: Review: **approved**, status `done`.
  - Process: test commit 1a3cded comes first. The implementation is 785ad0e, and the only later test change is the lead-authorized f34fe2e (case 4's minimum constant and its message). The engineer stopped at `blocked` instead of editing the test, which is the right call.
  - Code: `ProgressPolicy` follows the preference order exactly. It picks only from `legal`, reads only costs, pools and printed yields, and names no card. `SimOptions.Policy` nil keeps `RandomPolicy`, so `Simulate` and existing tests are unchanged.
  - Gates re-run by the reviewer:
    - 5 `TestProgress_*` pass, and `go tool task check` exits 0.
    - The default starter sim (progress) is byte-identical across runs: P0 81.5%, P1 15.5%, draws 3.0%, all invariants hold.
    - `-policy random` is byte-identical to `rework/executable-rules` for the starter flags and for bare `sim`, and `-policy bogus` exits 1.
    - The default stress-deck sim under progress: P0 18.0%, P1 82.0%, draws 0.0%, all invariants hold.
  - **New baselines for later reviews:** starter progress P0 81.5 / P1 15.5 / draws 3.0, and stress progress P0 18.0 / P1 82.0 / draws 0.0. The random baselines are reproducible with `-policy random`.
