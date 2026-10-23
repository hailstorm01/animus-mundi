---
id: T-0048
title: Split cmd/cardtown/play.go into the loop, the screen and input (directive #42, harness, flash, pure move)
kind: harness
tier: flash
wave: none (board directive #42, structural review 2 cleanup 4 of 5, before the next shape task)
status: in-progress
branch: m2/T-0048-split-play-go
depends_on: [T-0047]
claims: [file:cmd/cardtown/play.go, file:cmd/cardtown/play_screen.go, file:cmd/cardtown/play_input.go]
cards: []
requirements: []
---

## Goal
`cmd/cardtown/play.go` (423 lines) holds only `runPlay`: flags, setup and the loop. The screen goes to `play_screen.go`, where it will grow with new state (auras, durations). Input goes to `play_input.go`, where it will grow with new decision kinds. Nothing is renamed, and no behavior or output changes. The README's CLI section also gains `play`, which it lacks.

## Inputs
- Board directive #42 (structural review 2, finding 5). T-0045 and T-0047 have already made their changes to this file.
- Pattern to follow: T-0035 (the pure-move split of `main.go`) and T-0041.
- Rulings that apply: none.

### Target layout (lead decision)
| File | Holds (verbatim, with the comments directly above each) |
| :--- | :--- |
| `play.go` (stays) | `runPlay` |
| `play_screen.go` (new) | `printScreen`, `formatAsset`, `actionLabel` |
| `play_input.go` (new) | `errQuit`, `promptAction`, `terminalChooser` |

Further rules:
- **Order** within each file is the order the declarations had in `play.go`.
- **File headers**, like the other command files (one comment line, a blank line, then `package main`):
  - `play_screen.go`: `// Command: play, the screen: what the human sees, drawn only from view.PlayerView.`
  - `play_input.go`: `// Command: play, input: action prompts and the terminal Chooser.`
  - `play.go` keeps its `// Command: play.` header.
- **Imports:** each file imports only what it uses. Run `go tool task fmt`, then `go vet ./cmd/cardtown`.

### README (lead decision; copy exactly)
In `## The \`cardtown\` CLI`:
1. In the code block, insert this line directly before the `cardtown lint --all` line:
   ```text
   cardtown play -deck starter -p0 ANIMUS -p1 PATH                # play a match against the bot (go tool task play)
   ```
2. Insert this bullet directly after the `replay` bullet:
   ```markdown
   - **`play`** is a text-mode match against a sim policy. The screen and the event log come only from one seat's view (`pkg/view`), so the human never sees the bot's hand, its deck or a face-down card. Actions and decision options are numbered picks; `q` quits. Cards the kernel cannot run yet are marked `[not executable yet]`.
   ```
3. In the `lint` bullet, change `engine fences (no card names in engine code)` to `package fences (no card names in package code, and import direction between packages)`.

No other README change.

## Acceptance Cases (harness tasks only; written by the lead)
No new tests. Build a baseline CLI from `rework/executable-rules` before any edit, and record each check's commands and results in the Log. The reviewer re-runs them.
1. **Same declarations:** the sorted output of `git grep -h -E '^(func|type|var|const) ' -- 'cmd/cardtown/*.go' ':!cmd/cardtown/*_test.go'` is identical before and after.
2. **Placement:** `git grep -n -E '^(func|var) ' -- cmd/cardtown/play.go cmd/cardtown/play_screen.go cmd/cardtown/play_input.go` matches the target layout. Record the output.
3. **Only moves:** every line removed from `play.go` reappears in one of the two new files, apart from import lines. The only other added lines are the two headers, the `package` lines, the import blocks and blank lines.
4. **Tests unchanged:** `cmd/cardtown/play_test.go` has no diff, and `go test ./cmd/cardtown` passes.
5. **Outputs unchanged:** the baseline and new binaries give byte-identical stdout+stderr and exit codes (`cmp`) for:
   - `cardtown play -deck starter -p0 ANIMUS -p1 PATH -seat 0 -seed 3`, fed a file of a few thousand `1` lines;
   - the same with `-seat 1 -policy random`;
   - `cardtown play -h`;
   - `cardtown` with no arguments.
6. `go tool task check` passes.

## Allowed Files
This is a pure move: the size rule is met by cases 1–3, as in T-0035.
- `cmd/cardtown/play.go`
- new: `cmd/cardtown/play_screen.go`, `cmd/cardtown/play_input.go`
- `README.md` (the three README edits only)

## Done When
- [ ] Cases 1–5 hold, with their commands and results in the Log
- [ ] `go tool task check` passes

## Log
- 2026-10-22 lead-architect: Wrote the task from board directive #42. It is a harness task, tier flash, a pure move like T-0035. The README edits fold in two pieces of drift: the CLI section lacked `play`, and still called the fence "engine fences" after T-0038 widened it. T-0047 is done, so the task is `ready` on branch `m2/T-0048-split-play-go`.
  - Next: **the user runs `@kernel-engineer T-0048` in the Antigravity agent (Gemini Flash)**.
