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
