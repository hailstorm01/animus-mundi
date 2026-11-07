---
id: T-0194
title: Conversion and deck names, each spelled once (directive #132 (a), (b), harness, flash, pure move)
kind: harness
tier: flash
wave: D.2
status: in-progress
branch: m2/autoloop-wave-d.2/T-0194-conversion-and-deck-names-spelled-once
depends_on: [T-0193]
claims: [engine:ConvertTac, sim:decklistDir, sim:stressPrefix, cardtown:convert-label]
cards: []
requirements: []
---

## Goal
What one conversion gives, and the two deck-name strings, are each written in one place (#132 (a), (b)). A pure move: same test names, same sim output, and the menu label prints the same text.

## Inputs
- Board directive #132 (structural review 9, approved by the user, 2026-11-06). Its four parts touch eight non-test files, more than D2.4 allows in one task, so this task takes (a) and (b) and T-0195 takes (c) and (d). The directive stays on the board until T-0195 merges.
- Design note: `docs/architecture/m2-ability-model.md`, Deliverable 6h.
- #132's `policy.go` line numbers predate T-0190. `funds` is at `pkg/sim/policy.go:195` and `reachable` at `:222`.
- Why the menu label prints the same text: `%d` of the constants 2 and 1 gives `Convert: 2 Requisition -> 1 Tactics`. No test reads this label, so case 4 checks its format string and the reviewer reads the diff.
- Why `reachable` is unchanged: `*` and `/` group left to right, so the amount is `((s.req-c.req)/ConvertReq)*ConvertTac`, and `ConvertTac` is 1.
- Beyond #132's line range: `MatchDecks` also spells `data/decks` as `filepath.Join(repoRoot, "data", "decks")`, so it reads the constant too. `filepath.Join` cleans the trailing slash, so the path is the same.
- Stay as they are: `pkg/sim/monitor_resources.go` keeps its literals, as #132 says, because it is an independent check of `FR-RES-003`. The comment at `pkg/engine/actions.go:23` and the flag help in `cmd/cardtown/sim.go` and `play.go` are text, not reads. Tests spell their own expected names, on purpose.
- Reuse search (D2.3): no constant for the Tactics side of a conversion or for either deck string exists (searched `Convert`, `stress:`, `data/decks` and `"decks"` in `pkg` and `cmd`).
- The lead dry-ran these exact edits in a throwaway worktree of the run branch at 88b5e37: `gofmt -l pkg cmd` clean, `go tool task check` and `go tool task sim` pass, and `go tool task sim-compare -- autoloop/m2-wave-d.2` printed `[PASS]`. Every expected output below is from that dry run.

## Acceptance Cases (harness tasks only; written by the lead)
No new tests: a pure move, so no `test(T-0194)` commit. W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **Existing tests unchanged** (#132): `git -C W diff autoloop/m2-wave-d.2 -- '*_test.go'` prints nothing, and `go -C W test ./...` passes.
2. **Sims byte-identical** (#132): `go -C W tool task sim-compare -- autoloop/m2-wave-d.2` prints `[PASS]`.
3. **No literal conversion is left** (#132 (a)): `git -C W grep -n -e 'tac + 1' -e 'p.Tac++' -e 'Convert: 2' -- pkg/engine/resources.go pkg/sim/policy.go cmd/cardtown/play_screen.go` prints nothing. At the run branch it prints three lines, so an exit code of 1 here is the pass.
4. **The label formats both** (#132 (a)): `git -C W grep -n -e 'Convert: %d Requisition -> %d Tactics' -- cmd/cardtown/play_screen.go` prints exactly one line, at `cmd/cardtown/play_screen.go:162`.
5. **Each deck string is spelled once** (#132 (b)): `git -C W grep -n -e 'stress:' -e 'data.*decks' -- pkg/sim/decks.go` prints exactly three lines, at `:13` (`decklistDir`), `:15` (`stressPrefix`) and `:56` (the comment on `MatchDecks`).
6. **Only the listed edits:** `git -C W diff autoloop/m2-wave-d.2 --numstat -- pkg cmd` prints exactly four lines: `1	1	cmd/cardtown/play_screen.go`, `6	3	pkg/engine/resources.go`, `12	5	pkg/sim/decks.go` and `2	2	pkg/sim/policy.go`.
7. **Nothing else changed:** `git -C W diff autoloop/m2-wave-d.2 --stat` lists only those four files and this task file.
8. `go -C W tool task check` and `go -C W tool task sim` pass.

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)
Line numbers are at 88b5e37, before any edit.

### 1. `pkg/engine/resources.go`
Replace lines 21-22 with:
```go
// One conversion takes ConvertReq Requisition and gives ConvertTac Tactics (FR-RES-003).
const (
	ConvertReq = 2
	ConvertTac = 1
)
```
In `convert`, replace line 44, `p.Tac++`, with:
```go
	p.Tac += ConvertTac
```

### 2. `pkg/sim/policy.go`
On line 195, replace `pool.tac + 1` with `pool.tac + engine.ConvertTac`. The line becomes:
```go
		return shortTac && reachable(resources{pool.req - engine.ConvertReq, pool.tac + engine.ConvertTac}, rest, c.resources)
```
On line 222, replace `s.tac+(s.req-c.req)/engine.ConvertReq >= c.tac` with `s.tac+(s.req-c.req)/engine.ConvertReq*engine.ConvertTac >= c.tac`. The line becomes:
```go
		if s.req >= c.req && s.tac+(s.req-c.req)/engine.ConvertReq*engine.ConvertTac >= c.tac {
```

### 3. `cmd/cardtown/play_screen.go`
Replace line 162 with:
```go
		return fmt.Sprintf("Convert: %d Requisition -> %d Tactics", engine.ConvertReq, engine.ConvertTac)
```

### 4. `pkg/sim/decks.go`
Replace lines 11-18, from the comment on `DecklistName` through `StressName`, with:
```go
const (
	// decklistDir holds the decklists: its path from the repo root, slash-separated.
	decklistDir = "data/decks/"
	// stressPrefix begins a stress deck's recorded name.
	stressPrefix = "stress:"
)

// DecklistName is the recorded name of a decklist: its path from the repo
// root, slash-separated.
func DecklistName(faction, deck string) string {
	return fmt.Sprintf("%s%s-%s.txt", decklistDir, strings.ToLower(faction), deck)
}

// StressName is the recorded name of a faction's stress deck.
func StressName(faction string) string { return stressPrefix + faction }
```
Then three one-word replacements:
- Line 24: `strings.CutPrefix(name, "stress:")` becomes `strings.CutPrefix(name, stressPrefix)`.
- Line 34: `strings.TrimPrefix(DecklistName(f, name), "data/decks/")` becomes `strings.TrimPrefix(DecklistName(f, name), decklistDir)`.
- Line 54: `filepath.Join(repoRoot, "data", "decks")` becomes `filepath.Join(repoRoot, filepath.FromSlash(decklistDir))`.

No other file changes.

## Allowed Files
- `pkg/engine/resources.go`
- `pkg/sim/policy.go`
- `pkg/sim/decks.go`
- `cmd/cardtown/play_screen.go`

## Done When
- [ ] Cases 1-8 hold, with their commands and results in the Log
- [ ] `go tool task sim-compare -- autoloop/m2-wave-d.2` passes (starter and stress sims byte-identical to the run branch)
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-07, lead-architect (wave D.2 run, session 8): wrote this task from board directive #132 (a), (b), next in board order, against the run branch at 88b5e37, and T-0195 as `queued` for (c), (d) (D2.4). No `main` commits to sync. Not `WAVE-COMPLETE`: #132-#135 are `open: before the next Deliverables task`. Harness, so it skips `tests` and goes straight to `ready`; tier `flash`, because the edits are exact and were dry-run. Every command here is on the headless allow-list, one per line, with no pipes, `$`, double quotes or unquoted parentheses. Waiting directives checked: none has triggered (#6, #7: mode is `serial`; #15: no turn-boundary hook; #71, #91 (2), #106, #118 (b), #120, #128: no such card). Next: `kernel-engineer` on Flash.
