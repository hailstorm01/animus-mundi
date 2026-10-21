---
id: T-0022
title: A random sim policy that plays expensive cards
kind: harness
tier: sonnet
wave: none (board directive #21, ahead of the next A2 task)
status: ready
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
4. **Coverage (the point of the task):** in 200 starter games with `ProgressPolicy`, count `deploy` events by `Card`. Every card in either starter decklist that is `Supported()` is deployed at least 25 times. Record the per-card counts in the Log. If a card falls short, report the counts and stop at `blocked` rather than tuning the policy toward that card.
5. **Terminates and holds:** 200 starter games and 200 default stress-deck games under `ProgressPolicy` all finish within 400 turns with `CheckInvariants` holding (`Playout` errors otherwise).
6. **CLI default:** `go run ./cmd/cardtown sim -deck starter -p0 ANIMUS -p1 PATH` uses `progress`, exits 0 and is byte-identical across two runs. Record its summary next to the old one (the `random` baseline) in the Log.
7. **CLI `-policy random`:** `go run ./cmd/cardtown sim -deck starter -p0 ANIMUS -p1 PATH -policy random` and `go run ./cmd/cardtown sim -policy random` are byte-identical to `rework/executable-rules`' `sim` output for the same flags. An unknown `-policy` value prints `[ERROR]` and exits 1.

## Allowed Files
- `pkg/engine/playout.go`
- `pkg/engine/sim.go`
- `cmd/cardtown/main.go`
- `pkg/engine/policy_test.go` (new, the `test(T-0022)` commit)

## Done When
- [ ] Cases 1 to 5 have tests in `policy_test.go`, committed before the implementation, and all pass
- [ ] Cases 6 and 7 hold, with their output and the per-card deploy counts in the Log
- [ ] `go tool task check` passes
- [ ] Existing tests pass unchanged

## Log
- 2026-10-20 lead-architect: Wrote the task from board directive #21, with status `ready` (harness, sonnet). The user set its priority ahead of the next A2 task.
