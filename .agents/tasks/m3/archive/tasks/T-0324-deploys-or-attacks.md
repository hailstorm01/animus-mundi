---
id: T-0324
title: "\"Whenever this unit deploys or attacks\" is built in one place, `deploysOrAttacks` (directive #191; wave A.1 harness, flash)"
kind: harness
tier: flash
wave: A.1 (board directive #191, design note Deliverable 5a)
status: done
branch: m3/autoloop-wave-a.1/T-0324-deploys-or-attacks
depends_on: [T-0323]
claims: [func:cards.deploysOrAttacks]
cards: []
requirements: []
---

## Goal
"Whenever this unit deploys or attacks, …" is one printed ability that the kernel runs as two triggers, one on `HookDeploy` and one on `HookAttacks`, with the same `Effect`. Today each of the three cards with it writes that `Effect` out twice. After this task, one helper in `pkg/engine/cards/cards.go`, `deploysOrAttacks(t ability.Trigger) []ability.Trigger`, builds the pair, and *Earthshaker Colossus*, *Spoil Forest Colossus* and *Ironclad Colossus* each write their `Effect` once. Their definitions are unchanged in value, no test changes, and the sims are byte-identical. The *Stardust Colossus* task (wave F) uses the helper too.

## Inputs
- Board directive #191 (structural review 16, finding 2), approved by the user 2026-11-26, `open: before the next Deliverables task`. Design note `docs/architecture/m3-full-catalog.md` Deliverable 5a. Directive #190 (T-0323) is `applied`, so #191 is next in board order.
- The three copies, at `autoloop/m3-wave-a.1` fe4a56b3. Each card's `Triggers` is two literals that differ only in `Hook` (the lead checked each pair is identical text):
  - *Earthshaker Colossus*, [soulless.go:161-205](../../../../../pkg/engine/cards/soulless.go): deal 2 damage to a chosen non-Scratch asset;
  - *Spoil Forest Colossus*, [animus.go:703-743](../../../../../pkg/engine/cards/animus.go): the chosen asset heals 2;
  - *Ironclad Colossus*, [path.go:590-634](../../../../../pkg/engine/cards/path.go): the chosen unit gains 2 armor.
- No other card has the pattern: `git grep -n -i "deploys or attacks" -- data pkg/engine/cards` finds only these three definitions, and *Stardust Colossus* in `data/cards/units-and-structures-concordia.csv` (not executable yet).
- Not a `Hooks []Hook` field on `Trigger`: that is a shape change FRD 02 would have to follow, and the survey records two triggers (directive #191).
- The two triggers may share their `Effect` slices: definitions are read-only package values (directive #191).
- Rulings that apply: none. No behavior changes.
- **Expected output difference:** none. The three `Defs` entries are equal in value to the old ones, so `go tool task sim-compare -- autoloop/m3-wave-a.1` passes.
- The lead dry-ran these exact edits in a throwaway worktree of the run branch at fe4a56b3: `gofmt -l pkg cmd` clean, `go tool task check` exit 0, `go tool task sim-compare -- autoloop/m3-wave-a.1` `[PASS]`. A throwaway test dumping the three `Defs` entries (JSON and `%#v`) printed the same bytes before and after the edits. Resulting blobs (`git hash-object`): `cards.go` a585473, `soulless.go` 25eaf13, `animus.go` e64c9ef, `path.go` 4f29a30.

## Tests
No test is added, removed or edited. The existing card tests pin the behavior, and they must pass unchanged: every `TestCard_EarthshakerColossus_*` (`pkg/engine/cards_soulless_earthshaker_test.go`), `TestCard_SpoilForestColossus_*` (`pkg/engine/cards_animus_test.go`) and `TestCard_IroncladColossus_*` (`pkg/engine/cards_path_test.go`), deploy and attack cases both.

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)
1. **`pkg/engine/cards/cards.go`:** right after the line `var Defs = map[string]*ability.Def{}`, insert a blank line and then:
   ```go
   // deploysOrAttacks builds a "Whenever this unit deploys or attacks" ability:
   // one printed ability, two triggers, t on HookDeploy and t on HookAttacks.
   // t's own Hook is not used. The two copies share t's Effect slices, which is
   // safe because definitions are read-only package values.
   func deploysOrAttacks(t ability.Trigger) []ability.Trigger {
   	deploys, attacks := t, t
   	deploys.Hook = ability.HookDeploy
   	attacks.Hook = ability.HookAttacks
   	return []ability.Trigger{deploys, attacks}
   }
   ```
   The blank line that already follows `var Defs` then separates the function from the `// NotYetExecutable …` comment.
2. **The three cards** (`soulless.go` *Earthshaker Colossus*, `animus.go` *Spoil Forest Colossus*, `path.go` *Ironclad Colossus*). In each, replace the whole `Triggers: []ability.Trigger{ … },` value, both trigger literals, with one call that holds the `Effect` once and no `Hook` line, one tab less indented than before:
   ```go
   		Triggers: deploysOrAttacks(ability.Trigger{
   			Effect: ability.Effect{
   				… the card's Targets and Steps, exactly as in its old HookDeploy literal, one tab less …
   			},
   		}),
   ```
   *Earthshaker Colossus* afterwards reads, in full:
   ```go
   	"Earthshaker Colossus": {
   		Text: []string{"Whenever this unit deploys or attacks, deal 2 damage to chosen non-Scratch asset."},
   		Triggers: deploysOrAttacks(ability.Trigger{
   			Effect: ability.Effect{
   				Targets: []ability.Target{
   					{
   						Filter: ability.Filter{
   							Scratch: ability.ScratchExcluded,
   						},
   					},
   				},
   				Steps: []ability.Step{
   					{
   						Op:      ability.OpDealDamage,
   						Amount:  ability.Value{N: 2},
   						Subject: ability.SubjectTarget,
   						Target:  0,
   					},
   				},
   			},
   		}),
   	},
   ```
   The other two keep their own `Targets` and `Steps` (`Filter: ability.Filter{}` with `OpHeal`; `Kinds: ability.KindUnit` with `OpAddArmor`), each `Amount: ability.Value{N: 2}`. Leave `Text` lines as they are. Run `gofmt -w` on the four files, then check the blobs against the Inputs.

## Commits
1. `refactor(engine): one deploysOrAttacks builds the deploy-or-attack trigger pair, for directive 191 (T-0324)` with Edits 1-2 (no double quotes and no `#` in the message). No test changes, so there is no `test(T-0324)` commit.

## Acceptance Cases (harness tasks only; written by the lead)
These are checks recorded in the Log, and the reviewer re-runs them. No new tests.
1. `go test -list . ./pkg/engine/...` gives the same test names before and after (record the count), and `go test ./pkg/engine/...` passes.
2. `git grep -n "deploysOrAttacks" -- pkg` finds exactly five lines: in `cards.go`, the doc comment's first line and the `func deploysOrAttacks(` line; and one `Triggers: deploysOrAttacks(` line in each of `soulless.go`, `animus.go` and `path.go`.
3. `git grep -n -c "ability.HookAttacks" -- pkg/engine/cards/soulless.go pkg/engine/cards/animus.go pkg/engine/cards/path.go` shows one fewer `HookAttacks` in each file than on `autoloop/m3-wave-a.1`, and the same holds for `ability.HookDeploy` (compare with `git grep -n -c "ability.HookAttacks" autoloop/m3-wave-a.1 -- <the same files>`).
4. `git diff --stat autoloop/m3-wave-a.1 HEAD` lists only the four files under Allowed Files and this task file.
5. `go tool task sim-compare -- autoloop/m3-wave-a.1` passes.
6. `go tool task check` passes.

## Allowed Files
- `pkg/engine/cards/cards.go`, `pkg/engine/cards/soulless.go`, `pkg/engine/cards/animus.go`, `pkg/engine/cards/path.go`.
- No test file.

## Done When
- [x] Edits 1-2 are committed as `refactor(engine): … (T-0324)`
- [x] Cases 1-5 hold, with their commands and results in the Log
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim-compare -- autoloop/m3-wave-a.1` passes (starter and stress sims byte-identical to the run branch)

## Log
- 2026-11-26 lead-architect: Wrote the task from board directive #191, the second of wave A.1's four directives (design note Deliverable 5a). Harness, tier flash: a pure move inside `pkg/engine/cards`, one helper and three definitions, with no shape change. The lead dry-ran the edits (Inputs). Status `ready` on `m3/autoloop-wave-a.1/T-0324-deploys-or-attacks`, from `autoloop/m3-wave-a.1`. #192 and #193 follow, one task each.
- 2026-11-26 kernel-engineer: Implemented directive #191. Added `deploysOrAttacks(t ability.Trigger) []ability.Trigger` in `pkg/engine/cards/cards.go` to construct the HookDeploy and HookAttacks trigger pair sharing t's Effect slices. Refactored Earthshaker Colossus (`soulless.go`), Spoil Forest Colossus (`animus.go`), and Ironclad Colossus (`path.go`) to use `Triggers: deploysOrAttacks(...)`, eliminating duplicate trigger literals. Reused existing ability primitives with no new shapes or type additions. Resulting blob hashes match exact expected hashes from Inputs (`cards.go` a585473, `soulless.go` 25eaf13, `animus.go` e64c9ef, `path.go` 4f29a30). Verified all acceptance cases:
  1. `go test -list . ./pkg/engine/...` returned 1725 test names before and after; `go test ./pkg/engine/...` passed (0.840s).
  2. `git grep -n "deploysOrAttacks" -- pkg` found exactly 5 lines (`cards.go:11,15`, `animus.go:705`, `path.go:592`, `soulless.go:163`).
  3. `git grep -n -c "ability.HookAttacks"` and `HookDeploy` confirmed exactly 1 fewer hook count per file vs `autoloop/m3-wave-a.1` (HookAttacks: animus 3->2, path 3->2, soulless 3->2; HookDeploy: animus 5->4, path 5->4, soulless 4->3).
  4. `git diff --stat autoloop/m3-wave-a.1 HEAD` shows only the 4 Allowed Files and the task file.
  5. `go tool task sim-compare -- autoloop/m3-wave-a.1` passed: starter and stress sims match byte for byte, exit codes included.
  6. `go tool task check` passed cleanly (gofmt, go vet, all package tests, cardtown lint --all).
  Committed implementation as 15f41b80. Moved status to `review`.
- 2026-11-26 lead-architect: Review: **approved**, merged into `autoloop/m3-wave-a.1` as 1e4b5a8e, status `done`. Directive #191 is `applied`.
  - Scope: 487d7e17 and 30d0b18a change only this file. 15f41b80 changes only the four files under Allowed Files, and its blobs are the lead's dry-run blobs (`cards.go` a585473, `soulless.go` 25eaf13, `animus.go` e64c9ef, `path.go` 4f29a30). No test file changes. `deploysOrAttacks` copies `t` and sets only `Hook`, so it adds no shape, alias or wrapper.
  - Checks, re-run in a scratch worktree of the task branch: `go test -list . ./pkg/engine/...` gives 1725 tests, as on the run branch; case 2's grep finds the five lines; case 3's counts are one lower per file. The three `Defs` entries are equal in value to the old ones (the dry run's JSON and `%#v` dump, same blobs). A lead mutant (`deploysOrAttacks` returning only the deploy trigger) fails the attack tests of all three cards, among them `TestCard_EarthshakerColossus_AttackDealsTwoToTheChosenAsset` and `TestCard_IroncladColossus_AttackGrantResolvesBeforeCombatDamage`.
  - Gates: `go tool task check` exit 0; `go tool task sim` clean (P0 58.5% / P1 39.5% / draws 2.0%, unchanged); `go tool task sim-compare -- autoloop/m3-wave-a.1` `[PASS]`. `sim-matrix` is not this harness task's gate (D5.5); it runs with the wave's completion check.
