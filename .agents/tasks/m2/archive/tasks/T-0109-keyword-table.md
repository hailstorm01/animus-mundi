---
id: T-0109
title: Keyword names live in one table, keywordDefs, that ParseKeywordLine and Keywords.Lines both walk (directive #85, harness, flash, pure move)
kind: harness
tier: flash
wave: G.2 (board directive #85, structural review 6, finding 4)
status: done
branch: m2/T-0109-keyword-table
depends_on: [T-0108]
claims: [var:keywordDefs, func:Keywords.Lines]
cards: []
requirements: []
---

## Goal
The printed names of the standalone keywords are written once, in a table `keywordDefs` in `pkg/engine/card.go`. `ParseKeywordLine` (cardload.go) and a new `(k Keywords) Lines() []string` both walk it, and `pkg/view`'s own `keywords` goes: the view calls `Lines`. Today removing or adding a keyword edits both `ParseKeywordLine` and `pkg/view` (T-0090 did, for Berserk); keyword grants (waves E and F) and Overkill (M3) would again (#85). A pure move: same tests, same sim output.

## Inputs
- Board directive #85 (structural review 6, approved by the user, 2026-10-30).
- Design note: `docs/architecture/m2-ability-model.md`, Deliverable 6e (wave G.2).
- `ability.Keyword.String` keeps its own `Aggressive`: it names a keyword trigger, and `ability` imports nothing (#85). Do not touch `pkg/engine/ability`.
- Why it is a pure move: `Lines` prints the same lines in the same order as `view.keywords` did (numeric keywords only when N ≠ 0, as `<name> <N>`; flags by name; never nil). `ParseKeywordLine` accepts exactly the lines it accepted before: a flag keyword's exact name, or a numeric keyword's name, one space and `-?\d+` (the old regex `^(Resistant|Aggressive|Armor) (-?\d+)$`); a numeric keyword adds its N, a flag sets true; anything else returns false and leaves `kw` alone. The lead dry-ran these exact edits on the run branch at 273d3ae: `go tool task check` passed, `cardtown lint fences` passed, and `go tool task sim-compare -- autoloop/m2-wave-g.2` printed `[PASS]`.

### Where it is today (at `autoloop/m2-wave-g.2` 273d3ae)
| Where | Today |
| :--- | :--- |
| `pkg/engine/card.go:5-10` | imports `slices`, `strings` and `ability` |
| `pkg/engine/card.go:26-38` | `Keywords`; line 38 is its closing `}` |
| `pkg/engine/cardload.go:20-54` | `keywordLine` (20-21) and `ParseKeywordLine` (23-54) |
| `pkg/view/view.go:19-23` | imports `strconv` and `engine` |
| `pkg/view/view.go:140`, `:151` | `keywords(c.Keywords)`, `keywords(g.Keywords(a))` |
| `pkg/view/view.go:198-222` | a blank line, then `keywords`, the file's last function |

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)

### 1. `pkg/engine/card.go`
In the import block, add `"strconv"` between `"slices"` and `"strings"`.

After line 38 (`Keywords`' closing `}`), insert one blank line and then:
```go
// keywordDef is one standalone keyword: its printed name and its Keywords
// field. A numeric keyword (num) prints its N, as in "Resistant 2"; a flag
// keyword prints its name alone.
type keywordDef struct {
	name string
	num  func(*Keywords) *int
	flag func(*Keywords) *bool
}

// keywordDefs lists the standalone keywords in printed order. ParseKeywordLine
// and Keywords.Lines both walk it, so each keyword's name is written only here.
var keywordDefs = []keywordDef{
	{name: "Resistant", num: func(k *Keywords) *int { return &k.Resistant }},
	{name: "Stalwart", flag: func(k *Keywords) *bool { return &k.Stalwart }},
	{name: "Aggressive", num: func(k *Keywords) *int { return &k.Aggressive }},
	{name: "Ethereal", flag: func(k *Keywords) *bool { return &k.Ethereal }},
	{name: "Ward", flag: func(k *Keywords) *bool { return &k.Ward }},
	{name: "Versatile", flag: func(k *Keywords) *bool { return &k.Versatile }},
	{name: "Guardian", flag: func(k *Keywords) *bool { return &k.Guardian }},
	{name: "Armor", num: func(k *Keywords) *int { return &k.Armor }},
	{name: "Deploys Exhausted", flag: func(k *Keywords) *bool { return &k.DeploysExhausted }},
}

// Lines lists k's standalone keywords in printed form and order, never nil.
func (k Keywords) Lines() []string {
	out := []string{}
	for _, d := range keywordDefs {
		switch {
		case d.num != nil:
			if n := *d.num(&k); n != 0 {
				out = append(out, d.name+" "+strconv.Itoa(n))
			}
		case *d.flag(&k):
			out = append(out, d.name)
		}
	}
	return out
}
```
The blank line that followed line 38 now follows `Lines`' closing `}`, before `// Card holds …`. Nothing else in `card.go` changes.

### 2. `pkg/engine/cardload.go`
Replace lines 20-54 (from `// keywordLine matches a numeric keyword, which always prints its N.` through `ParseKeywordLine`'s closing `}`) with:
```go
// keywordN matches a numeric keyword's N, which always prints.
var keywordN = regexp.MustCompile(`^-?\d+$`)

// ParseKeywordLine parses one standalone keyword line from rules text
// (keywordDefs).
func ParseKeywordLine(line string, kw *Keywords) bool {
	for _, d := range keywordDefs {
		if d.flag != nil {
			if line == d.name {
				*d.flag(kw) = true
				return true
			}
			continue
		}
		if n, ok := strings.CutPrefix(line, d.name+" "); ok && keywordN.MatchString(n) {
			v, _ := strconv.Atoi(n)
			*d.num(kw) += v
			return true
		}
	}
	return false
}
```
The imports stay as they are (`regexp`, `strconv` and `strings` are all still used). Nothing else in `cardload.go` changes.

### 3. `pkg/view/view.go`
- Replace the import block (lines 19-23) with the single line `import "animus-mundi/pkg/engine"` (`strconv` is no longer used).
- Line 140: `keywords(c.Keywords)` becomes `c.Keywords.Lines()`.
- Line 151: `keywords(g.Keywords(a))` becomes `g.Keywords(a).Lines()`.
- Delete lines 198-222: the blank line before `// keywords lists k's standalone keywords in printed form, never nil.`, that comment and the function through its closing `}`. The file then ends with `stackView`'s closing `}` and a newline.

Nothing else in `view.go` changes. No other file changes, apart from the new test file below. `doc.go` needs nothing.

### 4. `pkg/engine/keywords_test.go` (new, the `test(T-0109)` commit)
Commit this first, on its own, as `test(T-0109): each standalone keyword prints one line that parses back`. It must fail before Edits 1–3 (it does not compile: `Keywords` has no `Lines`). Record that failure in the Log.
```go
package engine

import (
	"slices"
	"testing"
)

// Directive #85 (c): each standalone keyword, alone, prints one line, and
// parsing that line gives the same Keywords.
func TestKeywords_LinesRoundTrip(t *testing.T) {
	cases := []struct {
		kw   Keywords
		line string
	}{
		{Keywords{Resistant: 2}, "Resistant 2"},
		{Keywords{Stalwart: true}, "Stalwart"},
		{Keywords{Aggressive: 2}, "Aggressive 2"},
		{Keywords{Ethereal: true}, "Ethereal"},
		{Keywords{Ward: true}, "Ward"},
		{Keywords{Versatile: true}, "Versatile"},
		{Keywords{Guardian: true}, "Guardian"},
		{Keywords{Armor: 2}, "Armor 2"},
		{Keywords{DeploysExhausted: true}, "Deploys Exhausted"},
	}
	for _, c := range cases {
		if got := c.kw.Lines(); !slices.Equal(got, []string{c.line}) {
			t.Errorf("%+v.Lines() = %q, want [%q]", c.kw, got, c.line)
		}
		var parsed Keywords
		if !ParseKeywordLine(c.line, &parsed) || parsed != c.kw {
			t.Errorf("ParseKeywordLine(%q) = %+v, want %+v", c.line, parsed, c.kw)
		}
	}
}

// Lines is never nil, so a keywordless card's view encodes [] rather than null.
func TestKeywords_LinesEmpty(t *testing.T) {
	if got := (Keywords{}).Lines(); got == nil || len(got) != 0 {
		t.Errorf("Keywords{}.Lines() = %#v, want an empty non-nil slice", got)
	}
}
```

## Acceptance Cases (harness tasks only; written by the lead)
Run each command on its own, pointed at your worktree (W). Record each command and its result in the Log. The reviewer re-runs them.
1. **Existing tests unchanged** (#85 (a)): `git -C W diff autoloop/m2-wave-g.2 --stat -- '*_test.go'` lists only `pkg/engine/keywords_test.go`, and `go -C W test ./...` passes (the `pkg/view` tests included).
2. **Sims byte-identical** (#85 (b)): `go -C W tool task sim-compare -- autoloop/m2-wave-g.2` prints `[PASS]`.
3. **Round trip** (#85 (c)): `go -C W test ./pkg/engine -run TestKeywords_ -v` passes both `TestKeywords_LinesRoundTrip` and `TestKeywords_LinesEmpty`, and the `test(T-0109)` commit alone failed to build (`Lines` undefined).
4. **Names written once** (#85 (d)): `git -C W grep -n -F -e Stalwart -e "Deploys Exhausted" -- pkg ':!*_test.go'` prints exactly these 15 lines and no others: `pkg/engine/card.go:31` (the `Stalwart` field), `:54` and `:61` (the two table rows), `:128` (a comment); `pkg/engine/damage.go:7`, `:30`, `:31`; `pkg/engine/doc.go:13`, `:14`; `pkg/engine/events.go:222`; `pkg/engine/zones.go:169`; `pkg/sim/monitor_damage.go:91`, `:99`, `:104` and `:105`. Only `card.go:54` and `:61` spell a name as a string literal; the rest are comments, the field and field reads. None is in `cardload.go` or `pkg/view`.
5. **The view has no keyword list of its own:** `git -C W grep -n "func keywords" -- pkg` prints nothing (git grep exits 1 when nothing matches; that is the expected result).
6. **Moved, not changed:** `git -C W diff autoloop/m2-wave-g.2 --numstat -- pkg` prints exactly these four lines: `40	0	pkg/engine/card.go`, `17	30	pkg/engine/cardload.go`, `41	0	pkg/engine/keywords_test.go` and `3	32	pkg/view/view.go`.
7. **Fences:** `go -C W run ./cmd/cardtown lint fences` exits 0.
8. **Only these edits:** `git -C W diff autoloop/m2-wave-g.2 --stat` lists only the four files in case 6 and this task file.
9. `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/engine/card.go`
- `pkg/engine/cardload.go`
- `pkg/engine/keywords_test.go` (new)
- `pkg/view/view.go`

## Done When
- [x] Cases 1–9 hold, with their commands and results in the Log
- [x] `go tool task sim-compare -- autoloop/m2-wave-g.2` passes (starter and stress sims byte-identical to the run branch)
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks

## Log
- 2026-10-30, lead-architect (wave G.2 run, session 3): wrote the task from board directive #85, next in board order now that #83 and #84 are applied (T-0108). Harness, so it skips `tests` and goes straight to `ready`; tier `flash` as #85 gives. Three non-test files, within the size rule. Every command here is on the headless allow-list, one per line, with no `|`, `$` or unquoted parentheses. #85 (d) ("each appears in exactly one non-test Go file") is read as the names' string literals: comments and field reads (`kw.Stalwart`) also contain the words, so case 4 lists every line and shows that only `card.go` spells them as data. The table holds an accessor per field rather than a reflect lookup, so the compiler checks each field name. The edits were dry-run on the run branch (Inputs). `main` (281e7d8) has no commits the run branch lacks, so no sync. Waiting directives checked: none has triggered (#6, #7: mode is `serial`; #8: no `deck.go`; #15, #20 (b), #58, #59 (2), #61, #71: no such hook, claim or card here). #86 and #87 (c) stay `open: before the next Deliverables task`; #86 is next. Next: `kernel-engineer` on Flash.
- 2026-10-30, kernel-engineer (T-0109): implemented directive #85 in two commits:
  1. `test(T-0109): each standalone keyword prints one line that parses back` (3585287) wrote `pkg/engine/keywords_test.go` with `TestKeywords_LinesRoundTrip` and `TestKeywords_LinesEmpty`. Verified initial failure: compilation failed with `(Keywords{}).Lines undefined`.
  2. `feat(engine): keyword names live in one table keywordDefs (T-0109)` (3f60a1b) added `keywordDef`, `keywordDefs`, and `(k Keywords) Lines() []string` in `pkg/engine/card.go`; updated `ParseKeywordLine` in `pkg/engine/cardload.go` to walk `keywordDefs` using `strings.CutPrefix` and `keywordN`; updated `pkg/view/view.go` to use `Lines()` on cards and assets and removed private `keywords` helper and unused `strconv` import.
  Reused: existing `Keywords` struct fields and accessors, `strings.CutPrefix`, `strconv.Itoa`/`Atoi`.
  Acceptance cases verified on task branch:
  - Case 1: `git -C W diff autoloop/m2-wave-g.2 --stat -- '*_test.go'` listed only `pkg/engine/keywords_test.go`; `go -C W test ./...` passed.
  - Case 2: `go -C W tool task sim-compare -- autoloop/m2-wave-g.2` printed `[PASS] starter and stress sims match autoloop/m2-wave-g.2 byte for byte, exit codes included`.
  - Case 3: `go -C W test ./pkg/engine -run TestKeywords_ -v` passed both round-trip and empty tests.
  - Case 4: `git -C W grep -n -F -e Stalwart -e "Deploys Exhausted" -- pkg ':!*_test.go'` printed exactly the expected 15 lines across `pkg/engine` and `pkg/sim`, with string literals only in `card.go`.
  - Case 5: `git -C W grep -n "func keywords" -- pkg` returned no matches (exit code 1).
  - Case 6: `git -C W diff autoloop/m2-wave-g.2 --numstat -- pkg` showed exactly 40/0 `card.go`, 17/30 `cardload.go`, 41/0 `keywords_test.go`, and 3/32 `view.go`.
  - Case 7: `go -C W run ./cmd/cardtown lint fences` exited 0.
  - Case 8: `git -C W diff autoloop/m2-wave-g.2 --stat` listed only the 4 package files plus this task file.
  - Case 9: `go -C W tool task check` and `go -C W tool task sim` passed cleanly.
  Handoff: ready for lead-architect review.
- 2026-10-30, lead-architect (wave G.2 run, session 3, D5): **Approved, `done`.** Merged into `autoloop/m2-wave-g.2` as 5b09f74. Flash run `061ae9a8-759d-4848-b03b-26ce335c5b71`: exit 0, `status: SUCCESS`, 410 s, no refusals (the JSON has no `denied_actions` entry, `cli.log` has no `soft-denying` line and the transcript no `permission check failed`). Review of `git diff autoloop/m2-wave-g.2...m2/T-0109-keyword-table` (4e93608):
  - Scope: only `card.go`, `cardload.go`, `keywords_test.go`, `view.go` and this file changed. The `test(T-0109)` commit (3585287) holds only `keywords_test.go`, and on its own it fails to build (`go vet ./pkg/engine`: `c.kw.Lines undefined`); the `feat` commit (3f60a1b) touches no test file. No existing test changed.
  - Pure move, checked line by line: the code hunks are exactly Edits 1–3, identical to the lead's dry run. `Lines` prints what `view.keywords` printed, in the same order; `ParseKeywordLine` accepts the same lines as the old regex and switch (a flag's exact name; a numeric name, one space, `-?\d+`), adds N for a numeric keyword and sets a flag.
  - Fences and placement: no card name, no new rule; the table and `Lines` sit beside `Keywords` in `card.go`, as #85 places them. `pkg/view` keeps no keyword list; `ability.Keyword.String` is untouched.
  - Gates re-run in a scratch worktree of the task branch, each with exit 0: case 1 (only `keywords_test.go` among test files); case 4 (the 15 listed lines, string literals only at `card.go:54` and `:61`); case 5 (no output, exit 1 as expected); case 6 (the four numstat lines); case 7 `cardtown lint fences` `[PASS]`; `go tool task sim-compare -- autoloop/m2-wave-g.2` `[PASS]` (starter and stress byte-identical); `go tool task check` passes; `go tool task sim` clean, 200 games, P0 67.0% / P1 32.5% / draws 0.5% (unchanged), 32/50 Animus and 25/50 Path main cards executable.
  - The merge's tree equals the task tip's, so the gates hold on the run branch. Directive #85 is applied. #86 is next, then #87 (c).
