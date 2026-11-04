---
id: T-0155
title: The docs say what wave C built, doc.go, sim.go, ability.go and zones.go comments (directive #102, harness, flash, comments only)
kind: harness
tier: flash
wave: C.1 (board directive #102, structural review 7, finding 12)
status: ready
branch: m2/autoloop-wave-c.1/T-0155-docs-wave-c
depends_on: [T-0154]
claims: [comments:pkg/engine/doc.go, comments:pkg/sim/sim.go FactionDeck, comments:pkg/engine/ability/ability.go Subject, comments:pkg/engine/zones.go header]
cards: []
requirements: []
---

## Goal
Comments that wave C left stale say what the code does now (#102 (a), (b), (c)). No code changes: same declarations, same tests, same sim output. The lead made #102 (d), the design note's edits, in the commit before this task's.

## Inputs
- Board directive #102 (structural review 7, approved by the user, 2026-11-03). The table below has today's line numbers.
- Design note: `docs/architecture/m2-ability-model.md`, Deliverable 6f (wave C.1).
- Three lead decisions beyond #102's text, all comments:
  - **`sim.go`:** #102 (b) says a stress deck's Headquarters abilities run. True of *Animoceps Rager*, but `FactionDeck` takes the first Headquarters by name, which for Path is *Encampment*: it has no definition in `pkg/engine/cards/headquarters.go`, so its printed line stays `Unsupported` and does nothing. The comment says both.
  - **`zones.go` and `doc.go`'s Files entry:** T-0154's Inputs and review left these for #102. `deployCard` (T-0146) and `deployedTo` and `battlefieldZone` (T-0154) moved in, so each line now names them.
  - **`ability.go`:** `SubjectController`'s comment gets the same `OpDiscard` fix as `SubjectOpponents`' (line 238 beside #102 (c)'s line 240): `stepDiscard` reads both through `stepPlayers` (`pkg/engine/effects.go:126`, `:218`).
- Why it is comments only: the lead dry-ran these exact edits in a throwaway worktree of the run branch at cf6ab3d: `gofmt -l` clean, `go tool task check` passes, `go tool task sim` runs clean, and `go tool task sim-compare -- autoloop/m2-wave-c.1` printed `[PASS]`.

### Where it is today (at `autoloop/m2-wave-c.1`, cf6ab3d)
| Where | Today |
| :--- | :--- |
| `pkg/engine/doc.go:30-35` | the Steps bullet, without `OpMoveZone` or `Step.Dealer` |
| `pkg/engine/doc.go:47-51` | "Not built yet" still names Maneuvers and Heal N |
| `pkg/engine/doc.go:56` | `zones.go: moving cards and assets.` |
| `pkg/engine/zones.go:1` | the header line |
| `pkg/sim/sim.go:23` | `FactionDeck`'s Headquarters sentence |
| `pkg/engine/ability/ability.go:238`, `:240` | `SubjectController`'s and `SubjectOpponents`' trailing comments |

## Edits (lead decisions; copy exactly)

### 1. `pkg/engine/doc.go`, the Steps bullet
Replace lines 30-35 with this text. Line 29 (`//   - Steps OpScratch, ...`) stays, and line 36 (`//   - Ward on targets, ...`) follows the new text:
```go
//     OpDealDamage, OpHeal, OpAddArmor, OpModifyStat, OpKill (inline destruction) and OpMoveZone, with "all" steps, source-power values, the
//     deploy-zone gate, the TurnYours gate, the Cornered gate, SubjectSelf (tracked in discard,
//     FR-ZON-004) and SubjectOpponents. SubjectPick is an untargeted pick made as the
//     step resolves (SVY-R3), asked as a DecidePick and not a target, and a step with IfDone
//     runs only if the step before it did something ("if you do", SVY-R14), and a step with
//     IfHandEmpty runs only if its controller has no cards in hand when it resolves. A damage
//     step with Step.Dealer has an asset chosen for one of its slots deal the damage (SVY-R27).
//   - Target slots that choose once for each card the placement's discard cost
//     discarded (Target.PerDiscard, SVY-R4), or that must pick an entity different
//     from an earlier slot's (Target.Another, SVY-R34).
//   - Maneuvers (ActEnact, FR-MNV-001): modes and targets are chosen on placement,
//     as a trigger's are, the effect resolves as a StackManeuver item, and the card
//     then goes to its owner's discard pile. A Maneuver may print additional costs
//     to kill, exhaust or discard (Def.ExtraCosts), paid on declaration (FR-TRN-003a).
```

### 2. `pkg/engine/doc.go`, "Not built yet"
Replace the bullet under `// Not built yet:` (lines 47-51 today) with:
```go
//   - Activated and continuous abilities, continuous Cornered,
//     Regroup N, replacement effects other than destruction replacement and the
//     damage pipeline (damage redirection), and match
//     clocks. Cards whose text needs any of these load with Card.Unsupported set
//     and cannot be deployed or enacted.
```
The next line, `//     cards.NotYetExecutable lists ...`, stays.

### 3. `pkg/engine/doc.go`, the Files entry
In the line `//   - turn.go: beginTurn and endTurn. zones.go: moving cards and assets.`, replace `zones.go: moving cards and assets.` with `zones.go: moving cards and assets, deploying a card, and the zone a shape names.`

### 4. `pkg/engine/zones.go`, line 1
Replace the line with:
```go
// Zones: shuffling, drawing, deploying and moving cards and assets between zones, and the zone a shape names (FR-ZON-001, FR-ZON-004).
```

### 5. `pkg/sim/sim.go`, line 23
Replace `// The faction's first Headquarters is used with its printed abilities ignored.` with these two lines:
```go
// The faction's first Headquarters is used even if it is not Supported: the
// abilities its definition implements run, and its other lines are ignored.
```

### 6. `pkg/engine/ability/ability.go`, lines 238 and 240
Only the trailing comments change, and the code and the alignment stay:
- On line 238, replace `(the zero value; Scratch and Draw)` with `(the zero value; OpScratch, OpDraw, OpDiscard)`.
- On line 240, replace `each opponent of the ability's controller (OpScratch, OpDraw)` with `each opponent of the ability's controller (OpScratch, OpDraw, OpDiscard)`.

No other file changes.

## Acceptance Cases (harness tasks only; written by the lead)
No new tests: comments only. W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **Existing tests unchanged** (#102): `git -C W diff autoloop/m2-wave-c.1 -- '*_test.go'` prints nothing, and `go -C W test ./...` passes.
2. **Sims byte-identical** (#102): `go -C W tool task sim-compare -- autoloop/m2-wave-c.1` prints `[PASS]`.
3. **Only the listed edits:** `git -C W diff autoloop/m2-wave-c.1 --numstat -- pkg` prints exactly these four lines: `2	2	pkg/engine/ability/ability.go`, `13	5	pkg/engine/doc.go`, `1	1	pkg/engine/zones.go` and `2	1	pkg/sim/sim.go`.
4. **`doc.go` names what wave C built** (#102 (a)): `git -C W grep -n -e OpMoveZone -e Step.Dealer -e Target.PerDiscard -e Target.Another -e StackManeuver -e Def.ExtraCosts -- pkg/engine/doc.go` prints exactly six lines, at `doc.go:30`, `:36`, `:38`, `:39`, `:41` and `:43`.
5. **`doc.go` no longer calls them unbuilt** (#102 (a)): `git -C W grep -n -e 'Heal N' -e 'Maneuvers, activated' -- pkg/engine/doc.go` prints nothing (git exits 1).
6. **`sim.go`** (#102 (b)): `git -C W grep -n 'printed abilities ignored' -- pkg/sim/sim.go` prints nothing (git exits 1).
7. **`ability.go`** (#102 (c)): `git -C W grep -n -F 'OpScratch, OpDraw, OpDiscard' -- pkg/engine/ability/ability.go` prints exactly two lines, at `ability.go:238` and `:240`.
8. **Fences:** `go -C W run ./cmd/cardtown lint fences` exits 0.
9. **Nothing else changed:** `git -C W diff autoloop/m2-wave-c.1 --stat` lists only the four files in case 3 and this task file.
10. `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/engine/doc.go`
- `pkg/engine/zones.go`
- `pkg/sim/sim.go`
- `pkg/engine/ability/ability.go`

## Done When
- [ ] Cases 1–10 hold, with their commands and results in the Log
- [ ] `go tool task sim-compare -- autoloop/m2-wave-c.1` passes (starter and stress sims byte-identical to the run branch)
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-03, lead-architect (wave C.1 run, session 12): wrote this task from directive #102 against the run branch at cf6ab3d; `main` has no commits to sync. Harness, so it skips `tests` and goes straight to `ready`; tier `flash`, because the edits are exact and were dry-run. Every command here is on the headless allow-list, one per line, with no pipes, `$` or unquoted parentheses. Checked each claim against the code: `OpMoveZone`, `Step.Dealer`, `Target.PerDiscard`, `Target.Another`, `StackManeuver` and `Def.ExtraCosts` exist; `OpHeal` runs and no "Heal N" keyword is parsed; `Regroup` appears nowhere in `pkg/engine` outside `doc.go`; `stepDiscard` goes through `stepPlayers`. Waiting directives checked: none has triggered (#6, #7: mode is `serial`; #8: no `deck.go`; #15: no turn-boundary hook; #71, #91 (2): no such card). #103 is `open`, not blocking, and outside the wave's done-when. Next: `kernel-engineer` on Flash.
