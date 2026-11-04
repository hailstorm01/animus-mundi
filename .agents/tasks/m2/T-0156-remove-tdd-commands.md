---
id: T-0156
title: Remove the TDD commands from cardtown and whats-up (directive #103 (b)-(e), harness, flash)
kind: harness
tier: flash
wave: none (board directive #103, after wave C.1)
status: ready
branch: m2/T-0156-remove-tdd-commands
depends_on: []
claims: [cli:tdd command removed, cli:lint tdd removed, cli:tdd flag aliases removed, whats-up:tdd lines removed]
cards: []
requirements: []
---

## Goal
`cardtown tdd`, `cardtown lint tdd` and `cardtown lint --tdd` no longer exist, the linter's change detection no longer looks at `docs/tdd/`, and `cardtown whats-up` no longer lists the TDD tools or prints a block about TDD documents. This is directive #103 (b), (c), (d) and (e). After this task nothing under `cmd/` imports `pkg/toolchain/linter/tdd` or `pkg/toolchain/tddslice`. No engine code changes.

## Inputs
- Board directive #103. Its line numbers were checked at `e64ca5f` and still hold at `16a22d3`, where the lead took the numbers below.
- The split (D2.4): #103 edits five non-test files, and this task takes four of them. **T-0157** (`queued`) deletes `pkg/toolchain/linter/tdd` and `pkg/toolchain/tddslice`, removes `CommitExists` and `GetNewerCommits` from `pkg/toolchain/git/git.go`, and finishes the directive. Those packages stay in the tree after this task, so `git grep -i tdd -- pkg` still finds them, and `go build ./...` still compiles them.
- **Why the new tests never name the command.** #103's acceptance (c) wants `git grep -i tdd -- pkg cmd` to find nothing, including in test files. So a test can't say "`tdd` is unknown". The tests below pin the exact list of live commands, tools and flag aliases instead, which fails today and stays free of the word. Cases 6 and 7 check the unknown command and flag by running them.
- `cardtown lint tdd` after this task: `tdd` is no longer a subcommand, so it is treated like any other positional argument (a file name) by the dispatcher. That is what `cardtown lint <any word>` does today. No test or special case.
- The lead dry-ran these exact edits in a throwaway worktree of `main` at `16a22d3`: the four new tests failed before the code edits and passed after, `gofmt -l` was clean, `go tool task check` and `go tool task sim` pass, and `go tool task sim-compare` printed `[PASS]`.

## Test Edits (commit these first, as `test(T-0156): …`; copy exactly, tabs as in the surrounding code)
These are the only edits to existing tests (D2.4). W is your worktree. Line numbers are from each file before any edit, so work from the bottom of a file up.

### 1. `cmd/cardtown/args_test.go`
- Delete lines 18-21 (the cases `-DryRun`, `-updateExisting`, `-SHOWBODIES`, `-tddFile=a.md`).
- Delete line 50 (`"tdd": "tdd",`) and lines 59-67 (the rows `dryrun` through `tdd-file`).
- Replace line 86 with:
```go
	oldNames := strings.Fields("all frd cards sheets syntax fences staged files reporoot repo-root source msg commitmessage commit-message push section rules deck games seed p0 p1 seat policy maxturns max-turns")
```
- Append at the end of the file:
```go

func TestFlagAliases_NameOnlyLiveFlags(t *testing.T) {
	want := map[string]string{
		"reporoot":       "repo-root",
		"commitmessage":  "msg",
		"commit-message": "msg",
		"maxturns":       "max-turns",
	}
	if !reflect.DeepEqual(flagAliases, want) {
		t.Errorf("flagAliases = %#v, want %#v", flagAliases, want)
	}
}
```
- Then run `go -C W tool task fmt`. Removing the longest keys realigns the rest of the `oldSwitchExpected` map, so that map's rows show as changed. That is expected.

### 2. `cmd/cardtown/tools_test.go`
- Add `"slices"` to the imports, between `"os"` and `"strings"`.
- Replace line 36 with:
```go
	for _, cmd := range []string{"play", "lint", "hook", "sim", "replay", "whats-up"} {
```
- Append at the end of the file:
```go

func TestUsage_ListsOnlyLiveCommands(t *testing.T) {
	lines := strings.Split(captureOutput(printUsage), "\n")
	var got []string
	inCommands := false
	for _, l := range lines {
		switch {
		case l == "Available Commands:":
			inCommands = true
		case inCommands && strings.HasPrefix(l, "  "):
			got = append(got, strings.Fields(l)[0])
		default:
			inCommands = false
		}
	}
	want := []string{"play", "lint", "hook", "sim", "replay", "whats-up"}
	if !slices.Equal(got, want) {
		t.Errorf("usage commands = %q, want %q", got, want)
	}
	const lint = "  lint      Unified developer linter suite (FRD, executable rules, CSV sheets, rules syntax, fences)"
	if !slices.Contains(lines, lint) {
		t.Errorf("usage lacks the line %q", lint)
	}
}
```

### 3. `pkg/toolchain/docs/whatsup_test.go`
- Add `"slices"` to the imports, between `"os"` and `"strings"`.
- Replace line 37 with:
```go
	for _, want := range []string{"cardtown lint", "cardtown sim", "cardtown hook install", "cardtown whats-up"} {
```
- Append at the end of the file:
```go

func TestShowTools_ListsOnlyLiveTools(t *testing.T) {
	lines := strings.Split(captureOutput(func() { ShowTools(repoRoot) }), "\n")
	var got []string
	for _, l := range lines {
		if cmd, _, ok := strings.Cut(l, " : "); ok && strings.HasPrefix(cmd, "- cardtown ") {
			got = append(got, strings.TrimPrefix(cmd, "- "))
		}
	}
	want := []string{
		"cardtown lint [flags] [files...]",
		"cardtown lint rules",
		"cardtown lint fences",
		"cardtown sim [flags]",
		"cardtown hook install",
		"cardtown whats-up [flags]",
	}
	if !slices.Equal(got, want) {
		t.Errorf("tools = %q, want %q", got, want)
	}
	for _, line := range []string{
		"- cardtown lint [flags] [files...] : Unified developer linter suite (FRD, CSV sheets & syntax)",
		"    cardtown lint [--all|--frd|--rules|--cards|--sheets|--syntax|--fences|--staged] [--files <list>]",
	} {
		if !slices.Contains(lines, line) {
			t.Errorf("tools section lacks the line %q", line)
		}
	}
}

func TestShowState_PrintsNoDesignDocumentsBlock(t *testing.T) {
	out := captureOutput(func() { ShowState(repoRoot) })
	lines := strings.Split(out, "\n")
	for _, l := range lines {
		if strings.HasPrefix(l, "Technical Design Documents") {
			t.Errorf("state section prints %q", l)
		}
	}
}
```

Run `go -C W test ./cmd/cardtown ./pkg/toolchain/docs`. Exactly four tests must fail: `TestFlagAliases_NameOnlyLiveFlags`, `TestUsage_ListsOnlyLiveCommands`, `TestShowTools_ListsOnlyLiveTools` and `TestShowState_PrintsNoDesignDocumentsBlock`. That failure is the expected result here: record it and commit the three test files.

## Code Edits (lead decisions; copy exactly, tabs as in the surrounding code)
Line numbers are from each file before any edit, so work from the bottom of a file up.

### 4. `cmd/cardtown/lint.go`
| Lines | Today | Edit |
| :--- | :--- | :--- |
| 18 | the `linter/tdd` import | delete |
| 34-39 | `case "tdd":` and its body | delete |
| 78 | `tddFlag := fs.Bool("tdd", …)` | delete |
| 113 | `runTdd := false` | delete |
| 120 | `var changedTdds []string` | delete |
| 125 | `runTdd = true` under `*allFlag` | delete |
| 130 | the flag condition | replace (below) |
| 134-136 | `if *tddFlag { … }` | delete |
| 157 | `runTdd = true` after the staged-diff error | delete |
| 172-173 | `} else if strings.HasPrefix(pLower, "docs/tdd/") …` and its `changedTdds` append | delete |
| 185 | `runTdd = true` after the status error | delete |
| 200-201 | the same `docs/tdd/` branch and its append | delete |
| 213-216 | `if len(changedTdds) > 0 { … }` | delete |
| 223 | the nothing-to-lint condition | replace (below) |
| 224 | the nothing-to-lint message | replace (below) |
| 247-259 | the `// 2. TDD Linter` block and the blank line after it | delete |
| 260 | `// 2b. Executable Rules Traceability Linter` | replace (below) |

```go
	} else if *frdFlag || *rulesFlag || *cardsFlag || *sheetsFlag || *syntaxFlag || *fencesFlag {
```
```go
		if !runFrd && !runRules && !runSheets && !runSyntax && !runFences {
```
```go
			fmt.Println("  No changes detected in docs/frd/, pkg/, or data/cards/.")
```
```go
	// 2. Executable Rules Traceability Linter
```

### 5. `cmd/cardtown/main.go`
- Delete lines 43-44 (`case "tdd":` and `runTdd(os.Args[2:])`).
- Replace line 68 with:
```go
	fmt.Println("  lint      Unified developer linter suite (FRD, executable rules, CSV sheets, rules syntax, fences)")
```
- Delete line 70 (the `tdd` usage line).
- Delete lines 82-85 (the aliases `dryrun`, `updateexisting`, `showbodies`, `tddfile`).

### 6. `cmd/cardtown/tools.go`
- Replace line 1 with `// Commands: hook, whats-up.`
- Delete line 12 (the `tddslice` import).
- Delete lines 27-70 (`runTdd` and the blank line after it). `flag` stays, because `runWhatsUp` uses it.

### 7. `pkg/toolchain/docs/whatsup.go`
- Replace lines 92-93 with:
```go
	fmt.Println("- cardtown lint [flags] [files...] : Unified developer linter suite (FRD, CSV sheets & syntax)")
	fmt.Println("    cardtown lint [--all|--frd|--rules|--cards|--sheets|--syntax|--fences|--staged] [--files <list>]")
```
- Delete lines 102-103 (the two `cardtown tdd slice` lines).
- Delete lines 193-249 (from `tddDir := filepath.Join(repoRoot, "docs", "tdd")` through the `}` that closes the `else`). The blank line at 192 and the `fmt.Println()` at 250 stay, so `ShowState` ends with those. All the imports stay: `sort` and `regexp` have other uses.

## Acceptance Cases (harness tasks only; written by the lead)
W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **The flag aliases name only live flags:** `TestFlagAliases_NameOnlyLiveFlags`. Fails before the code edits.
2. **The usage lists only live commands:** `TestUsage_ListsOnlyLiveCommands`. Fails before the code edits.
3. **`whats-up` lists only live tools:** `TestShowTools_ListsOnlyLiveTools`. Fails before the code edits.
4. **`whats-up` prints no TDD block:** `TestShowState_PrintsNoDesignDocumentsBlock`. Fails before the code edits.
5. **No other test changed:** `git -C W diff main...HEAD --stat -- "*_test.go"` lists exactly `cmd/cardtown/args_test.go`, `cmd/cardtown/tools_test.go` and `pkg/toolchain/docs/whatsup_test.go`.
6. **The command is unknown:** `go -C W run ./cmd/cardtown tdd` prints `Unknown command "tdd"`, the usage and `exit status 1`, and the command exits non-zero. That failure is the expected result here: record it and go on. Checked by this command, not a test (see Inputs).
7. **The flag is unknown:** `go -C W run ./cmd/cardtown lint --tdd` prints `flag provided but not defined: -tdd`, the flag usage and `exit status 2`, and the command exits non-zero. That failure is the expected result here: record it and go on.
8. **`lint --all` runs without the TDD linter:** `go -C W run ./cmd/cardtown lint --all` exits 0. Its `>>> Running` lines are exactly FRD Requirements, Executable Rules Traceability, Card Sheet CSV Integrity, Card Rules Syntax & Terminology and Engine Architecture Fences.
9. **`whats-up` still runs:** `go -C W run ./cmd/cardtown whats-up` exits 0. Outside its `Uncommitted changes` list (which can name this task file), it prints no `TDD`, `tdd` or `Technical Design Documents`.
10. **The word is gone from these directories:** `git -C W grep -n -i tdd -- cmd pkg/toolchain/docs` prints nothing. `git grep` exits 1 when it finds nothing, which is the expected result here.
11. **Only the listed edits:** `git -C W diff main...HEAD --numstat -- cmd pkg` prints exactly these seven lines:
    ```text
    39	41	cmd/cardtown/args_test.go
    4	41	cmd/cardtown/lint.go
    1	8	cmd/cardtown/main.go
    1	46	cmd/cardtown/tools.go
    26	1	cmd/cardtown/tools_test.go
    2	61	pkg/toolchain/docs/whatsup.go
    41	1	pkg/toolchain/docs/whatsup_test.go
    ```
12. **The sims are unchanged:** `go -C W tool task sim-compare` prints `[PASS]`.
13. `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `cmd/cardtown/lint.go`
- `cmd/cardtown/main.go`
- `cmd/cardtown/tools.go`
- `pkg/toolchain/docs/whatsup.go`

Tests, edited only as `Test Edits` says, in the `test(T-0156)` commit: `cmd/cardtown/args_test.go`, `cmd/cardtown/tools_test.go`, `pkg/toolchain/docs/whatsup_test.go`.

## Done When
- [ ] Cases 1-4 have the tests above, each of which failed before the implementation commit; cases 5-13 hold, with their commands and results in the Log
- [ ] `go tool task sim-compare` passes (starter and stress sims byte-identical to the integration branch)
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-04, lead: written from #103, with T-0157 split off and queued (D2.4). Waiting triggers #6, #7, #8, #15, #71 and #91 (2) have not fired: this task touches no `deck.go`, no turn hook, no parallel mode and no card. 13 deck cards are still not executable (#14 (e)), and this task doesn't change that. Dispatch: `ready`, tier `flash`, because it deletes no files.
