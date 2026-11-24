---
id: T-0296
title: "LoadCards reports duplicate names, non-numeric or empty numeric cells, and Units or Structures with no zone (M3 Deliverable 1b; harness, flash)"
kind: harness
tier: flash
wave: none (M3 design note Deliverable 1, part b)
status: in-progress
branch: m3/T-0296-loader-sheet-checks
depends_on: [T-0295]
claims: [func:engine.loadSheets, func:engine.readCards, type:engine.sheetErrors]
cards: []
requirements: [FR-DAT-002, FR-DAT-006]
---

## Goal
`LoadCards` refuses card sheets that have any of the three data errors FRD 02 v2.0 names. It returns no cards and one error listing every problem it found, in file then row order. Each problem names the file, the row, the card and the column. Today the loader silently keeps the later of two rows with the same name ([cardload.go:146-148](../../../pkg/engine/cardload.go)) and reads a non-number as 0 ([cardload.go:177-180](../../../pkg/engine/cardload.go)).

## Inputs
- FRD 02: `FR-DAT-002` (lines 95-98: a duplicate name is a data error that names both rows), `FR-DAT-005` (line 109: a Maneuver has no power, health or yields), `FR-DAT-006` (line 120: a non-numeric value in a numeric column, or a Unit or Structure with no zone, is a data error naming the file, row, card and column; a card is never loaded with a guessed value).
- Design note: `docs/architecture/m3-full-catalog.md`, Decisions (the loader reports these errors, not `cardtown lint sheets`), Test Plan (Sheet checks), Deliverable 1 (b).
- Board: no open directive applies.
- Where it is today (at `main` f2c83b7c):
  - `loadSheets` ([cardload.go:129-151](../../../pkg/engine/cardload.go)) reads each `*.csv` in name order and keeps the last row for a name.
  - `readCards` ([cardload.go:153-239](../../../pkg/engine/cardload.go)) reads every numeric column with `strconv.Atoi`, ignoring the error. It reads zones but never checks for none, and returns at the first unknown `cardType`.
  - The only test that writes card sheets is `pkg/engine/card_test.go` (`fixtureDir`, `unitHeader`, `maneuverHeader`). Its rows have every numeric cell and zone. Its Maneuver sheet is named `maneuvers.csv`, which doesn't match the `maneuvers-` prefix that `FR-DAT-001` uses for real sheets.
  - The real sheets have none of the three errors. The lead scanned all 10 sheets on 2026-11-24: every cost cell, and every power, health and yield cell of a Unit, Structure or Headquarters, is a whole number, and every Unit and Structure has a zone.
- Design (lead decisions):
  - **Which columns are numeric goes by the row's card type, not the file name.** `requisitionCost` and `tacticsCost` are checked on every card, and `power`, `health`, `requisitionYield` and `tacticsYield` on every card but a Maneuver (`FR-DAT-005`). Going by file name would read the fixture's `maneuvers.csv` as an asset sheet.
  - An empty cell and a `--` cell in a numeric column are both errors (`FR-DAT-006`: no guessed value). The loader already reads `--` as empty.
  - "No zone" means no recognized zone was parsed, for a Unit or a Structure (`Card.Deployable`). A Headquarters needs none.
  - A duplicate is reported at the later row and names the first: `b.csv row 3 (Twice): name: the same name as a.csv row 2 (FR-DAT-002)`. Rows are numbered as a spreadsheet shows them: the header is row 1.
  - An unknown `cardType` joins the list instead of ending the load, so one run reports everything. Its row is skipped, because its numeric columns can't be known. CSV syntax errors and file errors stay fatal, as they are today.
  - The list is an unexported `sheetErrors []string`, whose `Error()` starts `card sheets: N data errors:` and puts one problem per line. Tests read it with `errors.As`. Nothing outside the package needs the type.
  - Deck loading, `loadCards`' definition checks and every caller are unchanged.
  - **Expected output difference:** none. The real sheets load as before, and `go tool task sim-compare` passes.
- Size (D2.4): one non-test file.
- The lead dry-ran these exact files in a throwaway worktree of `main` at f2c83b7c. On the old `cardload.go`, the new test file failed to build (`undefined: sheetErrors`). With the new one, every `TestLoadCards_` test passed, old and new, and `gofmt -l pkg/engine` was clean. `go vet ./pkg/engine` was clean, `go tool task check` exited 0, and `go tool task sim-compare` printed `[PASS]`. Blobs: `cardload.go` 1645f52, `cardload_sheet_test.go` da41ffe.

## Acceptance Cases (harness tasks only; written by the lead)
1. The same name in two sheets → one problem at the later row, naming both rows, the card and the column `name` (`TestLoadCards_DuplicateNameAcrossSheets`).
2. The same name twice in one sheet → one problem naming both rows (`TestLoadCards_DuplicateNameInOneSheet`).
3. `x` in a Unit's `power` → one problem naming the file, the row, the card, `power` and `"x"` (`TestLoadCards_NonNumericCell`).
4. An empty `health` and a `--` `tacticsYield` → two problems, each saying `empty` (`TestLoadCards_EmptyNumericCell`).
5. A Maneuver with whole-number costs and no power, health or yield columns → it loads. `two` in its `tacticsCost` → one problem (`TestLoadCards_ManeuverNumericCells`).
6. A Headquarters with no zone → it loads. Its empty `health` → one problem (`TestLoadCards_HeadquartersNumericCells`).
7. A Unit with an empty zone and a Structure with `--` → one problem each, naming `deploymentZones` (`TestLoadCards_DeployableWithNoZone`).
8. Problems in two sheets, including an unknown `cardType` → one error holding all four, in file then row order, whose message counts `4 data errors` (`TestLoadCards_AllProblemsTogetherInOrder`).
9. The real `data/cards/` → loads with no error (`TestLoadCards_RealSheetsHaveNoDataErrors`).
10. With any data error, `LoadCards` returns no cards (every case above, through `sheetProblems`).
11. Every existing `TestLoadCards_` test in `card_test.go` passes unedited.

## Tests (copy exactly; a new file, tabs as shown)
`pkg/engine/cardload_sheet_test.go`:
```go
package engine

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// T-0296: data errors the sheet loader reports (FR-DAT-002, FR-DAT-006).
// unitHeader and maneuverHeader come from card_test.go.

// sheetDir writes each file name → body pair into a fresh directory.
func sheetDir(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// assetRow is one asset-sheet row with the given type, zones and numeric cells
// (requisitionCost, tacticsCost, power, health, requisitionYield, tacticsYield).
func assetRow(name, typ, zones string, nums [6]string) string {
	return `"` + name + `","` + typ + `","PATH","Infantry","` + nums[0] + `","` + nums[1] + `","` + zones + `","--","` +
		nums[2] + `","` + nums[3] + `","` + nums[4] + `","` + nums[5] + `","COMMON","false"` + "\n"
}

// maneuverRow is one Maneuver-sheet row with the given costs.
func maneuverRow(name, req, tac string) string {
	return `"` + name + `","MANEUVER","PATH","` + req + `","` + tac + `","Draw a card.","COMMON","false"` + "\n"
}

var okNums = [6]string{"1", "0", "2", "2", "0", "0"}

// sheetProblems loads dir and returns the data errors, failing unless the load
// returned a sheetErrors and no cards.
func sheetProblems(t *testing.T, dir string) sheetErrors {
	t.Helper()
	cs, err := LoadCards(dir)
	var se sheetErrors
	if !errors.As(err, &se) {
		t.Fatalf("LoadCards error = %v, want a sheetErrors", err)
	}
	if cs != nil {
		t.Fatalf("LoadCards returned %d cards with data errors, want none", len(cs))
	}
	return se
}

func wantProblems(t *testing.T, got sheetErrors, want ...[]string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %d problems %q, want %d", len(got), got, len(want))
	}
	for i, parts := range want {
		for _, p := range parts {
			if !strings.Contains(got[i], p) {
				t.Fatalf("problem %d %q should contain %q (all: %q)", i, got[i], p, got)
			}
		}
	}
}

// 1. A name on two rows in two sheets: one problem, at the later row, naming both.
func TestLoadCards_DuplicateNameAcrossSheets(t *testing.T) {
	dir := sheetDir(t, map[string]string{
		"a.csv": unitHeader + "\n" + assetRow("Twice", "UNIT", "RESERVES", okNums),
		"b.csv": maneuverHeader + "\n" + maneuverRow("Once", "0", "1") + maneuverRow("Twice", "0", "1"),
	})
	wantProblems(t, sheetProblems(t, dir), []string{"b.csv row 3", "(Twice)", "name", "a.csv row 2", "FR-DAT-002"})
}

// 2. A name on two rows of one sheet.
func TestLoadCards_DuplicateNameInOneSheet(t *testing.T) {
	dir := sheetDir(t, map[string]string{
		"a.csv": unitHeader + "\n" + assetRow("Twice", "UNIT", "RESERVES", okNums) + assetRow("Twice", "STRUCTURE", "RESERVES", okNums),
	})
	wantProblems(t, sheetProblems(t, dir), []string{"a.csv row 3", "(Twice)", "a.csv row 2"})
}

// 3. A non-numeric cell in an asset's numeric column names the value.
func TestLoadCards_NonNumericCell(t *testing.T) {
	nums := okNums
	nums[2] = "x"
	dir := sheetDir(t, map[string]string{"a.csv": unitHeader + "\n" + assetRow("Bad", "UNIT", "RESERVES", nums)})
	wantProblems(t, sheetProblems(t, dir), []string{"a.csv row 2", "(Bad)", "power", `"x"`, "FR-DAT-006"})
}

// 4. An empty cell, and a "--" cell, in a numeric column count as non-numeric.
func TestLoadCards_EmptyNumericCell(t *testing.T) {
	nums := okNums
	nums[3], nums[5] = "", "--"
	dir := sheetDir(t, map[string]string{"a.csv": unitHeader + "\n" + assetRow("Blank", "UNIT", "RESERVES", nums)})
	wantProblems(t, sheetProblems(t, dir),
		[]string{"(Blank)", "health", "empty"},
		[]string{"(Blank)", "tacticsYield", "empty"})
}

// 5. A Maneuver needs whole-number costs and nothing else: a Maneuver sheet has
// no power, health or yield columns.
func TestLoadCards_ManeuverNumericCells(t *testing.T) {
	ok := sheetDir(t, map[string]string{"m.csv": maneuverHeader + "\n" + maneuverRow("Fine", "0", "2")})
	if _, err := LoadCards(ok); err != nil {
		t.Fatalf("a Maneuver with whole-number costs: %v", err)
	}
	bad := sheetDir(t, map[string]string{"m.csv": maneuverHeader + "\n" + maneuverRow("Costly", "0", "two")})
	wantProblems(t, sheetProblems(t, bad), []string{"m.csv row 2", "(Costly)", "tacticsCost", `"two"`})
}

// 6. A Headquarters is checked like any asset, and needs no deployment zone.
func TestLoadCards_HeadquartersNumericCells(t *testing.T) {
	ok := sheetDir(t, map[string]string{"h.csv": unitHeader + "\n" + assetRow("Home", "HEADQUARTERS", "", [6]string{"0", "0", "0", "30", "1", "1"})})
	if _, err := LoadCards(ok); err != nil {
		t.Fatalf("a Headquarters with no zone: %v", err)
	}
	bad := sheetDir(t, map[string]string{"h.csv": unitHeader + "\n" + assetRow("Home", "HEADQUARTERS", "", [6]string{"0", "0", "0", "", "1", "1"})})
	wantProblems(t, sheetProblems(t, bad), []string{"(Home)", "health", "empty"})
}

// 7. A Unit or a Structure with no deployment zone ("" or "--").
func TestLoadCards_DeployableWithNoZone(t *testing.T) {
	dir := sheetDir(t, map[string]string{
		"a.csv": unitHeader + "\n" + assetRow("Nowhere", "UNIT", "", okNums) + assetRow("Unbuilt", "STRUCTURE", "--", okNums),
	})
	wantProblems(t, sheetProblems(t, dir),
		[]string{"a.csv row 2", "(Nowhere)", "deploymentZones", "FR-DAT-006"},
		[]string{"a.csv row 3", "(Unbuilt)", "deploymentZones"})
}

// 8. Several problems across two sheets come back in one error, in file then
// row order, with an unknown cardType among them.
func TestLoadCards_AllProblemsTogetherInOrder(t *testing.T) {
	nums := okNums
	nums[0] = "?"
	dir := sheetDir(t, map[string]string{
		"b.csv": unitHeader + "\n" + assetRow("Second", "UNIT", "", okNums) + assetRow("First", "UNIT", "RESERVES", okNums),
		"a.csv": unitHeader + "\n" + assetRow("First", "UNIT", "RESERVES", okNums) + assetRow("Odd", "WIDGET", "RESERVES", okNums) + assetRow("Pricey", "UNIT", "RESERVES", nums),
	})
	se := sheetProblems(t, dir)
	wantProblems(t, se,
		[]string{"a.csv row 3", "(Odd)", "cardType", `"WIDGET"`},
		[]string{"a.csv row 4", "(Pricey)", "requisitionCost"},
		[]string{"b.csv row 2", "(Second)", "deploymentZones"},
		[]string{"b.csv row 3", "(First)", "a.csv row 2"})
	if msg := se.Error(); !strings.Contains(msg, "4 data errors") {
		t.Fatalf("Error() = %q, want it to count 4 data errors", msg)
	}
}

// 9. The real sheets have none of these errors.
func TestLoadCards_RealSheetsHaveNoDataErrors(t *testing.T) {
	if _, err := LoadCards("../../data/cards"); err != nil {
		t.Fatal(err)
	}
}
```

## Code (copy exactly)
In `pkg/engine/cardload.go`, replace everything from the line `func loadSheets(dir string) (map[string]*Card, error) {` (line 129) to the end of the file (the new text starts with the `sheetErrors` type, which goes before `loadSheets`) with the text below. Lines 1-128 stay as they are. The imports don't change.

```go
// sheetErrors lists every data error one load of the card sheets found
// (FR-DAT-002, FR-DAT-006), in file then row order. Each names the file, the
// row (the header is row 1), the card and the column.
type sheetErrors []string

func (e sheetErrors) Error() string {
	return fmt.Sprintf("card sheets: %d data errors:\n  %s", len(e), strings.Join(e, "\n  "))
}

// loadSheets reads every CSV in dir, in file name order. A duplicate name, a
// numeric cell that is not a whole number, a Unit or Structure with no zone,
// or an unknown cardType is a data error: loading goes on through every sheet
// and returns them all together, and no card is returned.
func loadSheets(dir string) (map[string]*Card, error) {
	paths, err := filepath.Glob(filepath.Join(dir, "*.csv"))
	if err != nil {
		return nil, err
	}
	sort.Strings(paths)
	cards := make(map[string]*Card)
	seen := make(map[string]string) // card name → "file row N" of its first row
	var problems sheetErrors
	for _, p := range paths {
		f, err := os.Open(p)
		if err != nil {
			return nil, err
		}
		loaded, found, err := readCards(f, filepath.Base(p), seen)
		f.Close()
		if err != nil {
			return nil, fmt.Errorf("%s: %w", filepath.Base(p), err)
		}
		problems = append(problems, found...)
		for _, c := range loaded {
			cards[c.Name] = c
		}
	}
	if len(problems) > 0 {
		return nil, problems
	}
	return cards, nil
}

// readCards reads one sheet. file names it in data errors, and seen maps each
// card name already read, in this sheet or an earlier one, to its row.
func readCards(r io.Reader, file string, seen map[string]string) ([]*Card, sheetErrors, error) {
	cr := csv.NewReader(r)
	cr.FieldsPerRecord = -1
	rows, err := cr.ReadAll()
	if err != nil {
		return nil, nil, err
	}
	if len(rows) == 0 {
		return nil, nil, nil
	}
	col := make(map[string]int)
	for i, h := range rows[0] {
		col[strings.TrimPrefix(strings.TrimSpace(h), string(rune(0xFEFF)))] = i
	}
	get := func(row []string, name string) string {
		if i, ok := col[name]; ok && i < len(row) {
			v := strings.TrimSpace(row[i])
			if v == "--" {
				return ""
			}
			return v
		}
		return ""
	}

	var out []*Card
	var problems sheetErrors
	for i, row := range rows[1:] {
		name := get(row, "name")
		if name == "" {
			continue
		}
		at := fmt.Sprintf("%s row %d", file, i+2)
		bad := func(column, format string, args ...any) {
			problems = append(problems, fmt.Sprintf("%s (%s): %s: %s", at, name, column, fmt.Sprintf(format, args...)))
		}
		if first, dup := seen[name]; dup {
			bad("name", "the same name as %s (FR-DAT-002)", first)
		} else {
			seen[name] = at
		}
		c := &Card{
			Name:      name,
			Faction:   get(row, "faction"),
			Rarity:    get(row, "rarity"),
			RulesText: get(row, "rulesText"),
		}
		switch get(row, "cardType") {
		case "UNIT":
			c.Type = TypeUnit
		case "STRUCTURE":
			c.Type = TypeStructure
		case "MANEUVER":
			c.Type = TypeManeuver
		case "HEADQUARTERS":
			c.Type = TypeHeadquarters
		default:
			bad("cardType", "unknown cardType %q", get(row, "cardType"))
			continue
		}
		num := func(column string) int {
			v := get(row, column)
			n, err := strconv.Atoi(v)
			switch {
			case err == nil:
				return n
			case v == "":
				bad(column, "empty, want a whole number (FR-DAT-006)")
			default:
				bad(column, "%q is not a whole number (FR-DAT-006)", v)
			}
			return 0
		}
		// Costs are numeric on every card; power, health and yields on every
		// card but a Maneuver, which has none (FR-DAT-005).
		c.ReqCost, c.TacCost = num("requisitionCost"), num("tacticsCost")
		if c.Type != TypeManeuver {
			c.Power, c.Health = num("power"), num("health")
			c.ReqYield, c.TacYield = num("requisitionYield"), num("tacticsYield")
		}
		for _, s := range strings.Split(get(row, "subtypes"), ",") {
			if s = strings.TrimSpace(s); s != "" {
				c.Subtypes = append(c.Subtypes, s)
			}
		}
		for _, z := range strings.Split(get(row, "deploymentZones"), ",") {
			switch strings.TrimSpace(z) {
			case "RESERVES":
				c.DeployZones = append(c.DeployZones, ZoneReserves)
			case "MAIN_THEATER":
				c.DeployZones = append(c.DeployZones, ZoneMainTheater)
			case "HEADQUARTERS":
				c.DeployZones = append(c.DeployZones, ZoneHeadquarters)
			}
		}
		if c.Deployable() && len(c.DeployZones) == 0 {
			bad("deploymentZones", "a Unit or Structure needs a deployment zone (FR-DAT-006)")
		}
		for _, line := range strings.Split(strings.ReplaceAll(c.RulesText, "\r", ""), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || line == "--" {
				continue
			}
			if !ability.ParseKeywordLine(line, &c.Keywords) {
				c.Unsupported = append(c.Unsupported, line)
			}
		}
		out = append(out, c)
	}
	return out, problems, nil
}
```

## Commits
1. `test(T-0296): the sheet loader reports duplicate names, bad numeric cells and missing zones` holds only the new test file. On this commit `go test ./pkg/engine -run TestLoadCards_` fails to build (`undefined: sheetErrors`).
2. `feat(engine): LoadCards reports every duplicate name, non-numeric cell and missing zone at once (T-0296)` holds `cardload.go`.

No double quotes in either message: the allow-list refuses an escaped one.

## Allowed Files
- `pkg/engine/cardload.go`

## Done When
- [ ] `go test ./pkg/engine -run TestLoadCards_ -count=1` passes
- [ ] `gofmt -l pkg/engine` prints nothing
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim-compare` passes (starter and stress sims byte-identical to `main`)

## Log
- 2026-11-23 lead-architect: queued with T-0295 as M3 Deliverable 1 (b). It doesn't depend on T-0295's code, and follows it in serial mode.
- 2026-11-24 lead-architect: written in full against `main` f2c83b7c. Status `ready` (a harness task skips `tests`), tier `flash`. The FRD 02 lines that say "today the loader..." (98, 120) follow by sync handoff once this merges.
