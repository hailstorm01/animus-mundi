---
id: T-0086
title: One untargeted pick body, pickFrom, and one predicate for "this filter chooses cards in a pile", Filter.InPile (directives #63 and #66, harness, flash, pure move)
kind: harness
tier: flash
wave: B.1 (board directives #63 and #66, structural review 4 findings 2 and 5)
status: ready
branch: m2/T-0086-one-pick
depends_on: [T-0085]
claims: [func:pickFrom, func:ability.Filter.InPile]
cards: []
requirements: []
---

## Goal
An untargeted pick has one body, `pickFrom`, whether it picks assets on the battlefield or cards in a discard pile; `pickAssets` and `pickCards` only build their decision and call it. "This filter chooses cards in a discard pile, not assets" is one documented predicate, `ability.Filter.InPile`, in place of six zone comparisons. The pick functions move from `effects.go` to `targeting.go`, beside `assetDecision` and `pileDecision`. Directive #66's comments say that picks are not battlefield-only. This is a pure move: same tests, same sim output.

## Inputs
- Board directives #63 and #66 (structural review 4, findings 2 and 5).
- Design note: `docs/architecture/m2-ability-model.md`, Deliverable 6a (wave B.1).
- Rules the moved code implements, unchanged: `SVY-R3`, `SVY-R14`, `SVY-R16`, `FR-MOD-001` §6, `FR-TGT-001`, `FR-TGT-003`.
- Why it is a pure move: `pickFrom` runs the same statements, in the same order, as the bodies it replaces (the empty check, one `g.choose(d)`, the index mapping), so every decision and every random draw is the same. `InPile` is exactly `Zone == ZoneDiscard`. `git grep` over `pkg` and `cmd` finds no test that names `pickAssets`, `pickCards` or `pickPrompt`.
- Size: five non-test files, one over the four of D2.4. Directive #63 puts #66's comment lines in this task, and `choice.go` gets one comment line only, so the code change itself touches four files (`ability.go`, `effects.go`, `targeting.go`, `stack.go`). Not split.

### Where it is today (at `autoloop/m2-wave-b.1` 2e6fa94)
| Where | Today |
| :--- | :--- |
| `effects.go:236` (`stepCards`) | `if st.Filter.Zone == ability.ZoneDiscard {` |
| `effects.go:296` (`stepAssets`) | `if st.Filter.Zone != ability.ZoneDiscard {` |
| `effects.go:303-347` | `pickAssets` (with doc comment), `pickPrompt`, `pickCards`, each body: build the decision, `if len(opts) == 0 { return nil }`, loop over `g.choose(d)` mapping to `opts[k]` |
| `targeting.go:47` (`targetDecision`) | `if t.Filter.Zone == ability.ZoneDiscard {` |
| `targeting.go:145` (`matchesCard`) | `if f.Zone != ability.ZoneDiscard \|\| (f.Other && inst == src) {` |
| `targeting.go:189` (`matchesFilter`) | `if f.Zone == ability.ZoneDiscard \|\| !deployedTo(a, f.Zone) {` |
| `stack.go:40` (`PileSlot`) | `… && it.Effect.Targets[i].Filter.Zone == ability.ZoneDiscard` |

## Edits (lead decisions; copy exactly)

### 1. `pkg/engine/ability/ability.go`
- Right after the `Filter` struct's closing brace, add:
  ```go

  // InPile reports whether f chooses cards in a discard pile rather than assets
  // on the battlefield (Zone is ZoneDiscard). Every target, pick and filter check
  // that tells the two apart asks this, not the zone.
  func (f Filter) InPile() bool { return f.Zone == ZoneDiscard }
  ```
- Directive #66, comments only:
  - `SubjectPick`'s trailing comment becomes: `// the entities matching Step.Filter, picked by the ability's controller as the step resolves: battlefield assets, or cards in a discard pile if Filter.InPile. One, or with Step.UpTo up to N, possibly none. Not a target (SVY-R3, SVY-R14)`
  - `Step.Filter`'s trailing comment becomes: `// SubjectAll: which assets. SubjectPick: which assets, or which cards in a discard pile (Filter.InPile)`
  - `Step.UpTo`'s trailing comment becomes: `// SubjectPick: 0 picks exactly one; N > 0 picks up to N distinct entities (assets or cards), possibly none`

### 2. `pkg/engine/choice.go`
Directive #66, a comment only. `DecidePick`'s trailing comment becomes:
`// untargeted pick from the battlefield or a discard pile as a step resolves (SVY-R3, SVY-R14): not a target, so Ward and the targeting checks skip it; same fallback as DecideTarget`

### 3. `pkg/engine/stack.go`
In `PileSlot`, `it.Effect.Targets[i].Filter.Zone == ability.ZoneDiscard` becomes `it.Effect.Targets[i].Filter.InPile()`. Nothing else changes.

### 4. `pkg/engine/effects.go`
- The file header (line 1) becomes:
  `// Effects: resolving a stack item's effect, its steps in order and their values (FR-TGT-003). Untargeted picks are targeting.go's.`
- In `stepCards`, `if st.Filter.Zone == ability.ZoneDiscard {` becomes `if st.Filter.InPile() {`. In its doc comment, `SubjectPick with ZoneDiscard is the` becomes `SubjectPick with a pile filter is the`.
- In `stepAssets`, `if st.Filter.Zone != ability.ZoneDiscard {` becomes `if !st.Filter.InPile() {`.
- Remove `pickAssets`, `pickPrompt` and `pickCards` with their doc comments (they move, step 5). Keep the `fmt` and `slices` imports: `stepDiscard` and `stepCards` still use them.

### 5. `pkg/engine/targeting.go`
- The file header (line 1) becomes:
  `// Targeting: choosing targets on placement, untargeted picks on resolution, and checking entities against filters (FR-TGT-001, FR-TGT-003, SVY-R14).`
- In `targetDecision`, `if t.Filter.Zone == ability.ZoneDiscard {` becomes `if t.Filter.InPile() {`. Rewrap its doc comment (the T-0085 review nit) to:
  ```go
  // targetDecision builds the decision for target slot i (clause t) of the
  // effect pl places: the legal entities of the slot, as pileDecision or
  // assetDecision list them. chooseTargets asks it, and legalModes reads its
  // Options (FR-MOD-001 §5), so both use the same filters.
  ```
- In `matchesCard`, `if f.Zone != ability.ZoneDiscard || (f.Other && inst == src) {` becomes `if !f.InPile() || (f.Other && inst == src) {`. In its doc comment, `Zone must be ZoneDiscard.` becomes `f must be InPile.`
- In `matchesFilter`, `if f.Zone == ability.ZoneDiscard || !deployedTo(a, f.Zone) {` becomes `if f.InPile() || !deployedTo(a, f.Zone) {`.
- Right after `pileDecision` (before `pileTarget`), add the moved functions in this order: `pickAssets`, `pickPrompt`, `pickCards`, each with its doc comment exactly as in `effects.go` today, then the new `pickFrom`. The bodies of `pickAssets` and `pickCards` become:
  ```go
  func (g *Game) pickAssets(r *resolution, st ability.Step) []*Asset {
  	d, opts := g.assetDecision(r.it.Controller, DecidePick, pickPrompt(r.it, st), st.UpTo, func(a *Asset) bool {
  		return g.matchesFilter(r.it.Controller, r.it.Source, st.Filter, a)
  	})
  	return pickFrom(g, d, opts)
  }
  ```
  ```go
  func (g *Game) pickCards(r *resolution, st ability.Step) []*Instance {
  	d, opts := g.pileDecision(r.it.Controller, DecidePick, pickPrompt(r.it, st), st.UpTo, func(inst *Instance) bool {
  		return g.matchesCard(r.it.Controller, r.it.Inst, st.Filter, inst)
  	})
  	return pickFrom(g, d, opts)
  }
  ```
  `pickPrompt` moves unchanged. Then:
  ```go
  // pickFrom asks d and returns the options picked, in pick order: the body of
  // every untargeted pick, over assets or cards. opts are d's options, in order
  // (as assetDecision and pileDecision return them). With no option nothing is
  // asked.
  func pickFrom[T any](g *Game, d Decision, opts []T) []T {
  	if len(opts) == 0 {
  		return nil
  	}
  	var out []T
  	for _, k := range g.choose(d) {
  		out = append(out, opts[k])
  	}
  	return out
  }
  ```
- `targeting.go` already imports `fmt`, which `pickPrompt` needs.

No other file changes.

## Acceptance Cases (harness tasks only; written by the lead)
No new tests: a pure move. Run each command on its own, pointed at your worktree (W). `git grep` ignores untracked files, so stage your edits with `git -C W add` before cases 3–5. Record each command and its result in the Log. The reviewer re-runs them.
1. **Existing tests unchanged:** `git -C W diff autoloop/m2-wave-b.1 -- '*_test.go'` prints nothing, and `go -C W test ./...` passes.
2. **Sims byte-identical:** `go -C W tool task sim-compare -- autoloop/m2-wave-b.1` prints `[PASS]`.
3. **One pile predicate** (directive #63 case (c)): `git -C W grep -n "Zone [!=]= ability.ZoneDiscard" -- pkg/engine` prints nothing, and `git -C W grep -n "== ZoneDiscard" -- pkg/engine/ability` prints exactly one line, `InPile`'s body.
4. **One pick body:** `git -C W grep -n "pickFrom" -- pkg/engine` prints the `func pickFrom` line, its doc comment's first line, and exactly two calls, one in `pickAssets` and one in `pickCards`.
5. **Picks live in targeting.go:** `git -C W grep -n "func .*pick" -- pkg/engine/effects.go` prints nothing, and `git -C W grep -n "func .*pick" -- pkg/engine/targeting.go` prints the five lines of `pickBounds` (already there), `pickAssets`, `pickPrompt`, `pickCards` and `pickFrom`.
6. **Doc lines (#66):** `go -C W doc ./pkg/engine/ability Filter.InPile` prints `InPile`'s doc comment, and `go -C W doc -all ./pkg/engine/ability` shows the new `SubjectPick`, `Step.Filter` and `Step.UpTo` comments.
7. **Only these edits:** `git -C W diff autoloop/m2-wave-b.1 --stat` lists only `pkg/engine/ability/ability.go`, `pkg/engine/choice.go`, `pkg/engine/stack.go`, `pkg/engine/effects.go`, `pkg/engine/targeting.go` and this task file.
8. `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/engine/ability/ability.go`
- `pkg/engine/choice.go` (one comment)
- `pkg/engine/stack.go`
- `pkg/engine/effects.go`
- `pkg/engine/targeting.go`

## Done When
- [ ] Cases 1–7 hold, with their commands and results in the Log
- [ ] `go tool task sim-compare -- autoloop/m2-wave-b.1` passes (starter and stress sims byte-identical to the run branch)
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks

## Log
- 2026-10-28, lead-architect (wave B.1 run, session 2): wrote the task from board directives #63 and #66 (#66 goes in the task for #63). Harness, so it skips `tests` and goes straight to `ready`. Tier `flash` as #63 gives. The edits are spelled out so the engineer copies them. They also take the T-0085 review nit, rewrapping `targetDecision`'s doc comment. `main` has no commits the run branch lacks, so no sync. Directives #64 and #65 (with #67) stay open for the next sessions; #61 stays `waiting`, because no task's `cards:` has a modal card. Next: `kernel-engineer` (Flash, headless).
