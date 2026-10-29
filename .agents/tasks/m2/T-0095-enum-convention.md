---
id: T-0095
title: The enum convention is written down, and ValueFixed becomes ValueSourceFixed (directive #70, harness, flash, comments and a rename)
kind: harness
tier: flash
wave: B.2 (board directive #70, the user's code review, before wave C)
status: ready
branch: m2/T-0095-enum-convention
depends_on: [T-0094]
claims: [const:ValueSourceFixed]
cards: []
requirements: []
---

## Goal
A reader of `ability` learns from its package comment which of the three enum patterns a new shape enum uses (`1 << iota` for flags that combine, `iota + 1` where an unset value must be invalid, `iota` where the zero value is a meaningful default), and a reader of `engine` learns from `doc.go` that its own enums are plain `iota` lists whose zero is an ordinary member. `ValueFixed` is renamed `ValueSourceFixed`, matching `ValueSourcePower`, and `Value`'s doc comment says why it is a struct. Comments and one unreferenced rename: same tests, same sim output.

## Inputs
- Board directive #70 (user code review, 2026-10-28: "why are iotas sometimes declared just as iota, and others as +1, and others with <<?", "Why the split between Value and ValueSource").
- Design note: `docs/architecture/m2-ability-model.md`, Deliverable 6b (wave B.2).
- Rules: none change.
- Why the sims can't change: the rename keeps the constant's value (0, the first `iota`), and `git grep -n -w ValueFixed -- pkg cmd tools` finds it only at its declaration, so no code names it (every fixed value is the zero `Value.From`). Every other edit is a comment.

### Where it is today (at `autoloop/m2-wave-b.2` 8f30df8)
| Where | Today |
| :--- | :--- |
| `pkg/engine/ability/ability.go:1-4` | the package comment (three lines) and `package ability` |
| `pkg/engine/ability/ability.go:173-186` | `ValueSource`, its constants `ValueFixed` and `ValueSourcePower`, and `Value` with its two-line doc comment |
| `pkg/engine/doc.go`, near the end | the "Package boundary" paragraph, a blank `//` line, then the paragraph starting "Where FRD 01 leaves a case undefined" |
| `pkg/engine` enums | `ActionKind` (`actions.go:22`), `Resource` (`actions.go:39`), `CardType` (`card.go:16`), `DecisionKind` (`choice.go:12`), `StackKind` (`stack.go:11`), `Zone` (`state.go:21`), `Phase` (`state.go:43`); all start at `iota`, unchanged by this task |

## Edits (lead decisions; copy exactly)

### 1. `pkg/engine/ability/ability.go`, the package comment
Replace the three comment lines above `package ability` with:
```go
// Package ability is the shape vocabulary for printed card abilities: plain
// data, no game logic and no card names (the fences linter enforces the
// latter). New types here arrive only through a `shape` task.
//
// A new enum here follows one of three rules, chosen by what its zero value
// means:
//   - A set of flags that combine is declared with 1 << iota, one bit per
//     member, so a value may hold several (Kinds).
//   - An enum with no default starts at iota + 1, so an unset field (the zero
//     value) is invalid and never passes for the first member (Hook, Op).
//   - An enum whose zero value is a meaningful default starts at iota, and the
//     first member's comment says "the zero value" (TurnOwner, Keyword, Zone,
//     Allegiance, Subject, ValueSource).
```
The list lines start `//   - ` (three spaces) and their continuation lines `//     ` (five spaces), as in `engine`'s `doc.go`; gofmt keeps them that way.

### 2. `pkg/engine/ability/ability.go`, `ValueSource` and `Value`
Replace the `const (` block of `ValueSource` and `Value`'s doc comment, so the code from `// ValueSource is where` to the end of `Value` reads:
```go
// ValueSource is where a Value's number comes from.
type ValueSource int

const (
	ValueSourceFixed ValueSource = iota // the zero value: N
	ValueSourcePower                    // the source's effective power when the step resolves, or its last-known effective power (SVY-R10); N is ignored
)

// Value is a number a step uses, read when the step resolves; From says where
// it comes from. It is a struct, not a bare int, so that each source can carry
// what it needs: N is read only for ValueSourceFixed (the zero source), and a
// later source, such as "the number of X", adds its own field here without
// changing Step.
type Value struct {
	N    int
	From ValueSource
}
```
`ValueSourcePower`'s line and `Value`'s fields are unchanged. Run `gofmt -w` on the file afterwards.

### 3. `pkg/engine/doc.go`
Between the "Package boundary" paragraph and the paragraph starting `// Where FRD 01 leaves a case undefined`, insert this paragraph, with a blank `//` line before it and one after it:
```go
// Enums: this package's enums (ActionKind, Resource, CardType, DecisionKind,
// StackKind, Zone, Phase) are plain iota lists. Their zero value is an
// ordinary first member (ActGenerate, ZoneDeck, PhaseSetup), not a default or
// an unset marker, and ActionKind.String indexes an array by value. The shape
// enums in pkg/engine/ability follow that package's three rules instead.
```

No other file changes. No enum's values change.

## Acceptance Cases (harness tasks only; written by the lead)
No new tests: comments and an unreferenced rename. Run each command on its own, pointed at your worktree (W). Record each command and its result in the Log. The reviewer re-runs them.
1. **Existing tests unchanged** (#70 case (a)): `git -C W diff autoloop/m2-wave-b.2 -- '*_test.go'` prints nothing, and `go -C W test ./...` passes.
2. **Sims byte-identical** (#70 case (b)): `go -C W tool task sim-compare -- autoloop/m2-wave-b.2` prints `[PASS]`.
3. **The three rules show in `go doc`** (#70 case (c)): `go -C W doc ./pkg/engine/ability` prints the package comment above, with its three list items naming `1 << iota`, `iota + 1` and `iota`.
4. **Nothing names `ValueFixed`** (#70 case (c)): `git -C W grep -n -w ValueFixed -- pkg cmd tools` prints nothing (exit code 1). `git -C W grep -n ValueSourceFixed -- pkg cmd tools` prints exactly two lines, both in `pkg/engine/ability/ability.go`: the constant and the line of `Value`'s doc comment that names it.
5. **`Value` says why it is a struct:** `go -C W doc ./pkg/engine/ability Value` prints the doc comment above, which names `ValueSourceFixed` and "the number of X".
6. **`engine` says what its enums are:** `go -C W doc ./pkg/engine` prints the `Enums:` paragraph above.
7. **gofmt-clean:** `gofmt -l W/pkg/engine` prints nothing.
8. **Only these edits:** `git -C W diff autoloop/m2-wave-b.2 --stat` lists only `pkg/engine/ability/ability.go`, `pkg/engine/doc.go` and this task file.
9. `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/engine/ability/ability.go`
- `pkg/engine/doc.go`

## Done When
- [ ] Cases 1–9 hold, with their commands and results in the Log
- [ ] `go tool task sim-compare -- autoloop/m2-wave-b.2` passes (starter and stress sims byte-identical to the run branch)
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks

## Log
- 2026-10-28, lead-architect (wave B.2 run, session 7): wrote the task from board directive #70, the last open directive of wave B.2 now that #69 is applied (T-0094). Harness, so it skips `tests` and goes straight to `ready`; tier `flash` as #70 gives; every command here is on the headless allow-list. The comment text was checked with `gofmt -d` and `go doc` before it went into the task. #70 case (c) says "nothing in the repository names `ValueFixed`"; done task files (T-0028), the board and the design note's Deliverable 6b keep the old name as history, so case 4 checks `pkg`, `cmd` and `tools`. `doc.go` lists `engine`'s seven enums and why they stay plain `iota` (#70: `ActionKind.String` indexes an array by value); no enum's values change. `main` has no commits the run branch lacks, so no sync. Waiting directives checked: none has triggered (#8: no `deck.go`; #15, #58, #61, #71: no such hook, claim or card here; #6, #7: mode is `serial`). Next: `kernel-engineer` on Flash.
