---
id: T-0187
title: Decklist parse errors name the file (directives #124 and #8, harness, flash)
kind: harness
tier: flash
wave: D.2
status: in-progress
branch: m2/autoloop-wave-d.2/T-0187-decklist-errors-name-file
depends_on: []
claims: [engine:LoadDeck line errors, test:TestLoadDeck_Case13_ParseErrorsNameTheFile]
cards: []
requirements: [FR-SET-001]
---

## Goal
Every `line N` error from `LoadDeck` also names the decklist, as `<path>: line N: <detail>`, in the form its `missing HQ line` error already uses. The sims load legal decklists, so no game output changes.

## Inputs
- Board directive #124, which replaces #8's trigger. Both are archived when this task merges.
- `pkg/engine/deck.go:124-161`: the eight `line %d` errors. The read error and `missing HQ line` already name the path and stay as they are.
- Reuse search: the only other deck-error helper is `illegalDeck`, for construction violations, which wraps `ErrIllegalDeck`. A parse error must not wrap it, because `TestLoadDeck_Case7` and `SimDecks` keep the two apart. No caller of `LoadDeck` adds the path (`pkg/sim/decks.go:27` and `:36-39` return its error as is), so it doesn't print twice.
- `missing card name` can't fire: `line` is trimmed, so whitespace always has text after it. It takes the same helper, because #124 says each, and the test has no case for it. Removing it is not this task.
- `TestLoadDeck_Case3`, `Case5` and `Case9` check the line number with `strings.Contains`. Once the message holds a temp path, which can contain those digits, they could pass on the path alone. They stay unedited, and the new test compares the whole prefix, including the line number, for every error that can fire.
- The lead dry-ran these exact edits in a throwaway worktree of the run branch at 5e9dc90. On the unchanged `deck.go`, all seven subtests of `TestLoadDeck_Case13_ParseErrorsNameTheFile` fail and nothing else does. With both edits: `gofmt -l` is clean, and `go tool task check`, `go tool task sim` and `go tool task sim-compare` pass.

## Acceptance Cases
1. A decklist with two `HQ:` lines → the error starts `<path>: line 2: multiple HQ lines`.
2. `HQ:` with no name → `<path>: line 1: missing HQ card name`.
3. `HQ:` naming an unknown card → `<path>: line 1: unknown card name "Ghost"`.
4. An entry with no count → `<path>: line 2: invalid entry format "Trooper"`.
5. A count that isn't a positive integer → `<path>: line 2: invalid copy count "two"`.
6. The same card on two lines → `<path>: line 3: duplicate card "Trooper"`.
7. A main-deck line naming an unknown card → `<path>: line 2: unknown card name "Ghost"`.
8. No `line %d` format string is left in `LoadDeck`: the helper holds the only one.
9. `go tool task sim-compare -- autoloop/m2-wave-d.2` passes: no sim loads a decklist that fails to parse.

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)

### 1. `pkg/engine/deck_load_test.go`, in the `test(T-0187)` commit
Add this at the end of the file, with one blank line before it:
```go
// T-0187 (directive #124): a parse error names the decklist and its line. The
// whole prefix is compared, because a temp path can hold the line's digits.
func TestLoadDeck_Case13_ParseErrorsNameTheFile(t *testing.T) {
	cards := map[string]*Card{
		"Commander": deckHQ("ANIMUS"),
		"Trooper":   deckCard("Trooper", "ANIMUS", "COMMON"),
	}
	tests := []struct {
		name    string
		content string
		line    int
		detail  string
	}{
		{"multiple HQ lines", "HQ: Commander\nHQ: Commander\n", 2, "multiple HQ lines"},
		{"missing HQ card name", "HQ:\n", 1, "missing HQ card name"},
		{"unknown HQ card", "HQ: Ghost\n", 1, `unknown card name "Ghost"`},
		{"invalid entry format", "HQ: Commander\nTrooper\n", 2, `invalid entry format "Trooper"`},
		{"invalid copy count", "HQ: Commander\ntwo Trooper\n", 2, `invalid copy count "two"`},
		{"duplicate card", "HQ: Commander\n1 Trooper\n1 Trooper\n", 3, `duplicate card "Trooper"`},
		{"unknown main card", "HQ: Commander\n1 Ghost\n", 2, `unknown card name "Ghost"`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := writeDeckFile(t, t.TempDir(), "deck.txt", tc.content)
			_, err := LoadDeck(p, cards)
			if err == nil {
				t.Fatalf("%s: expected error, got nil", tc.name)
			}
			want := p + ": line " + strconv.Itoa(tc.line) + ": " + tc.detail
			if !strings.HasPrefix(err.Error(), want) {
				t.Errorf("%s: got %q, want it to start with %q", tc.name, err, want)
			}
		})
	}
}
```
Commit only this file as `test(T-0187): ...`, after confirming that all seven subtests fail.

### 2. `pkg/engine/deck.go`
1. Right before the comment `// LoadDeck reads a decklist file and resolves its names against cards.`, add this, with one blank line after it:
   ```go
   // deckLineError reports a parse error on one line of the decklist at path.
   func deckLineError(path string, line int, format string, args ...any) error {
   	return fmt.Errorf("%s: line %d: %s", path, line, fmt.Sprintf(format, args...))
   }
   ```
   Remove the three leading spaces of the block's indentation when you copy it.
2. In `LoadDeck`, change each of the eight `fmt.Errorf("line %d: <text>", lineNum` calls to `deckLineError(path, lineNum, "<text>"`, keeping `<text>` and any arguments after `lineNum` as they are. For example, `fmt.Errorf("line %d: unknown card name %q", lineNum, cardName)` becomes `deckLineError(path, lineNum, "unknown card name %q", cardName)`. Change nothing else.

## Allowed Files
- `pkg/engine/deck.go`
- `pkg/engine/deck_load_test.go` (Edit 1 only, in the test commit)

## Done When
- [ ] `go test ./pkg/engine -run TestLoadDeck` passes
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-compare -- autoloop/m2-wave-d.2` passes (starter and stress sims byte-identical to the run branch)

## Log
- 2026-11-06 (lead-architect): Written for #124 and #8, `ready`. First task of wave D.2.
