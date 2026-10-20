package fences_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"animus-mundi/pkg/toolchain/linter/fences"
)

func setupFixtureRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	cardsDir := filepath.Join(dir, "data", "cards")
	if err := os.MkdirAll(cardsDir, 0755); err != nil {
		t.Fatalf("mkdir cards: %v", err)
	}

	unitsCSV := `name,cardType,faction,subtypes,requisitionCost,tacticsCost,deploymentZones,rulesText,power,health,requisitionYield,tacticsYield,rarity,needsRework
Chainbreaker Phalanx,UNIT,PATH,Soldier,3,0,MAIN_THEATER,"Rules",3,3,0,0,COMMON,false
Rift,HEADQUARTERS,ANIMUS,--,0,0,HEADQUARTERS,"Rules",0,20,0,0,HQ,false
Scratch Company,UNIT,NEUTRAL,Token,0,0,MAIN_THEATER,"Rules",1,1,0,0,COMMON,false
`
	if err := os.WriteFile(filepath.Join(cardsDir, "units.csv"), []byte(unitsCSV), 0644); err != nil {
		t.Fatalf("write units.csv: %v", err)
	}

	maneuversCSV := `name,cardType,faction,requisitionCost,tacticsCost,rulesText,rarity,needsRework
Steel Rain,MANEUVER,PATH,2,0,"Rules",COMMON,false
Chainbreaker's Tenacity,MANEUVER,PATH,1,0,"Rules",COMMON,false
`
	if err := os.WriteFile(filepath.Join(cardsDir, "maneuvers.csv"), []byte(maneuversCSV), 0644); err != nil {
		t.Fatalf("write maneuvers.csv: %v", err)
	}

	engineDir := filepath.Join(dir, "pkg", "engine")
	if err := os.MkdirAll(engineDir, 0755); err != nil {
		t.Fatalf("mkdir engine: %v", err)
	}
	cleanGo := `package engine

func Clean() {}
`
	if err := os.WriteFile(filepath.Join(engineDir, "clean.go"), []byte(cleanGo), 0644); err != nil {
		t.Fatalf("write clean.go: %v", err)
	}

	frdDir := filepath.Join(dir, "docs", "frd")
	if err := os.MkdirAll(frdDir, 0755); err != nil {
		t.Fatalf("mkdir frd: %v", err)
	}
	matrixContent := `# Keyword Interaction Matrix

## Open Items (need a ruling)

| ID | Status | Reach | Question | Kernel today |
| :-- | :-: | :-: | :-- | :-- |
| KIM-B07 | ? | U | **Berserk stacking.** | n/a |

## Rulings

| ID | Ruling (2026-10-19) | FRD 01 | Pinned by |
| :-- | :-- | :-- | :-- |
| KIM-B01 | A Ward asset cannot be the Berserk secondary. | FR-KWD-002o | TestFR_KWD_002o |
`
	if err := os.WriteFile(filepath.Join(frdDir, "keyword_interaction_matrix.md"), []byte(matrixContent), 0644); err != nil {
		t.Fatalf("write matrix: %v", err)
	}

	designDir := filepath.Join(dir, "docs", "design")
	if err := os.MkdirAll(designDir, 0755); err != nil {
		t.Fatalf("mkdir design: %v", err)
	}
	surveyContent := `# Ability Survey

## 4. Ruling Questions

| ID | Question | Cards | Status |
| :--- | :--- | :--- | :--- |
| **R4a** | Repeated targets | Steel Rain | **Ruled: yes.** Each choice is independent. |
| **R7** | Open question | Future Card | Open |
`
	if err := os.WriteFile(filepath.Join(designDir, "m2_starter_set_ability_survey.md"), []byte(surveyContent), 0644); err != nil {
		t.Fatalf("write survey: %v", err)
	}

	return dir
}

func captureOutput(f func()) string {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		outC <- buf.String()
	}()

	f()

	_ = w.Close()
	os.Stdout = oldStdout
	return <-outC
}

// 1. Engine files contain no card names and no markers -> Lint returns (true, nil) and prints [PASS].
func TestFences_Case01_NoCardNamesOrMarkers(t *testing.T) {
	repo := setupFixtureRepo(t)
	var ok bool
	var err error
	out := captureOutput(func() {
		ok, err = fences.Lint(repo, nil)
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !ok {
		t.Fatalf("expected pass, got false")
	}
	if !strings.Contains(out, "[PASS]") {
		t.Fatalf("expected [PASS] in output, got:\n%s", out)
	}
}

// 2. pkg/engine/x.go has if c.Name == "Chainbreaker Phalanx" { -> (false, nil), with one [ERROR] line naming x.go:<line> and Chainbreaker Phalanx.
func TestFences_Case02_CardNameInCode(t *testing.T) {
	repo := setupFixtureRepo(t)
	code := `package engine

func Check(c struct{ Name string }) {
	if c.Name == "Chainbreaker Phalanx" {
	}
}
`
	xPath := filepath.Join(repo, "pkg", "engine", "x.go")
	if err := os.WriteFile(xPath, []byte(code), 0644); err != nil {
		t.Fatalf("write x.go: %v", err)
	}

	var ok bool
	var err error
	out := captureOutput(func() {
		ok, err = fences.Lint(repo, nil)
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if ok {
		t.Fatalf("expected failure, got true")
	}
	if !strings.Contains(out, "[ERROR]") || !strings.Contains(out, "x.go:4") || !strings.Contains(out, "Chainbreaker Phalanx") {
		t.Fatalf("expected [ERROR] with x.go:4 and Chainbreaker Phalanx, got:\n%s", out)
	}
	if n := strings.Count(out, "[ERROR]"); n != 1 {
		t.Fatalf("expected exactly one [ERROR] line, got %d:\n%s", n, out)
	}
}

// 3. A card name inside a comment (// mirrors Steel Rain) -> fails, same as case 2.
func TestFences_Case03_CardNameInComment(t *testing.T) {
	repo := setupFixtureRepo(t)
	code := `package engine

// mirrors Steel Rain
func Foo() {}
`
	xPath := filepath.Join(repo, "pkg", "engine", "x.go")
	if err := os.WriteFile(xPath, []byte(code), 0644); err != nil {
		t.Fatalf("write x.go: %v", err)
	}

	var ok bool
	var err error
	out := captureOutput(func() {
		ok, err = fences.Lint(repo, nil)
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if ok {
		t.Fatalf("expected failure, got true")
	}
	if !strings.Contains(out, "[ERROR]") || !strings.Contains(out, "x.go:3") || !strings.Contains(out, "Steel Rain") {
		t.Fatalf("expected [ERROR] with x.go:3 and Steel Rain, got:\n%s", out)
	}
}

// 4. A card name as a camelCase identifier (func chainbreakerPhalanxBonus()) -> fails.
func TestFences_Case04_CardNameInCamelCaseIdentifier(t *testing.T) {
	repo := setupFixtureRepo(t)
	code := `package engine

func chainbreakerPhalanxBonus() {}
`
	xPath := filepath.Join(repo, "pkg", "engine", "x.go")
	if err := os.WriteFile(xPath, []byte(code), 0644); err != nil {
		t.Fatalf("write x.go: %v", err)
	}

	var ok bool
	var err error
	out := captureOutput(func() {
		ok, err = fences.Lint(repo, nil)
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if ok {
		t.Fatalf("expected failure, got true")
	}
	if !strings.Contains(out, "[ERROR]") || !strings.Contains(out, "Chainbreaker Phalanx") {
		t.Fatalf("expected error naming Chainbreaker Phalanx, got:\n%s", out)
	}
}

// 5. A card name in a different case ("steel rain") -> fails.
func TestFences_Case05_CardNameDifferentCase(t *testing.T) {
	repo := setupFixtureRepo(t)
	code := `package engine

var name = "steel rain"
`
	xPath := filepath.Join(repo, "pkg", "engine", "x.go")
	if err := os.WriteFile(xPath, []byte(code), 0644); err != nil {
		t.Fatalf("write x.go: %v", err)
	}

	var ok bool
	var err error
	out := captureOutput(func() {
		ok, err = fences.Lint(repo, nil)
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if ok {
		t.Fatalf("expected failure, got true")
	}
	if !strings.Contains(out, "[ERROR]") || !strings.Contains(out, "Steel Rain") {
		t.Fatalf("expected error naming Steel Rain, got:\n%s", out)
	}
}

// 6. An apostrophe name written without its apostrophe ("Chainbreakers Tenacity" for Chainbreaker's Tenacity) -> fails.
func TestFences_Case06_CardNameMissingApostrophe(t *testing.T) {
	repo := setupFixtureRepo(t)
	code := `package engine

var name = "Chainbreakers Tenacity"
`
	xPath := filepath.Join(repo, "pkg", "engine", "x.go")
	if err := os.WriteFile(xPath, []byte(code), 0644); err != nil {
		t.Fatalf("write x.go: %v", err)
	}

	var ok bool
	var err error
	out := captureOutput(func() {
		ok, err = fences.Lint(repo, nil)
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if ok {
		t.Fatalf("expected failure, got true")
	}
	if !strings.Contains(out, "[ERROR]") || !strings.Contains(out, "Chainbreaker's Tenacity") {
		t.Fatalf("expected error naming Chainbreaker's Tenacity, got:\n%s", out)
	}
}

// 7. A single-word card name inside a longer word (drift := 0 with a card named Rift) -> passes.
func TestFences_Case07_SingleWordCardNameInsideLongerWord(t *testing.T) {
	repo := setupFixtureRepo(t)
	code := `package engine

func DriftCalc() {
	drift := 0
	_ = drift
}
`
	xPath := filepath.Join(repo, "pkg", "engine", "x.go")
	if err := os.WriteFile(xPath, []byte(code), 0644); err != nil {
		t.Fatalf("write x.go: %v", err)
	}

	var ok bool
	var err error
	out := captureOutput(func() {
		ok, err = fences.Lint(repo, nil)
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !ok {
		t.Fatalf("expected pass for 'drift', got false; output:\n%s", out)
	}
}

// 8. A single-word card name as a whole word (case "rift":) -> fails.
func TestFences_Case08_SingleWordCardNameWholeWord(t *testing.T) {
	repo := setupFixtureRepo(t)
	code := `package engine

func CheckRift(s string) {
	switch s {
	case "rift":
	}
}
`
	xPath := filepath.Join(repo, "pkg", "engine", "x.go")
	if err := os.WriteFile(xPath, []byte(code), 0644); err != nil {
		t.Fatalf("write x.go: %v", err)
	}

	var ok bool
	var err error
	out := captureOutput(func() {
		ok, err = fences.Lint(repo, nil)
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if ok {
		t.Fatalf("expected failure for whole word 'rift', got true")
	}
	if !strings.Contains(out, "[ERROR]") || !strings.Contains(out, "Rift") {
		t.Fatalf("expected error naming Rift, got:\n%s", out)
	}
}

// 9. "Scratch Company" in a string literal, a comment and HasSubtype("Scratch Company") -> passes (rules-object exemption).
func TestFences_Case09_RulesObjectExemption(t *testing.T) {
	repo := setupFixtureRepo(t)
	code := `package engine

// Scratch Company token
func HasSubtype(s string) bool { return s == "Scratch Company" }
`
	xPath := filepath.Join(repo, "pkg", "engine", "x.go")
	if err := os.WriteFile(xPath, []byte(code), 0644); err != nil {
		t.Fatalf("write x.go: %v", err)
	}

	var ok bool
	var err error
	out := captureOutput(func() {
		ok, err = fences.Lint(repo, nil)
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !ok {
		t.Fatalf("expected Scratch Company to be exempted, but failed; output:\n%s", out)
	}
}

// 10. A card name in pkg/engine/x_test.go -> passes (test files are out of scope).
func TestFences_Case10_TestFilesOutOfScope(t *testing.T) {
	repo := setupFixtureRepo(t)
	testCode := `package engine_test

import "testing"

func TestCard(t *testing.T) {
	name := "Chainbreaker Phalanx"
	_ = name
}
`
	xTestPath := filepath.Join(repo, "pkg", "engine", "x_test.go")
	if err := os.WriteFile(xTestPath, []byte(testCode), 0644); err != nil {
		t.Fatalf("write x_test.go: %v", err)
	}

	var ok bool
	var err error
	out := captureOutput(func() {
		ok, err = fences.Lint(repo, nil)
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !ok {
		t.Fatalf("expected test files to be out of scope, but failed; output:\n%s", out)
	}
}

// 11. A card name in pkg/engine/cards/y.go -> passes (subdirectories are out of scope).
func TestFences_Case11_SubdirectoriesOutOfScope(t *testing.T) {
	repo := setupFixtureRepo(t)
	subDir := filepath.Join(repo, "pkg", "engine", "cards")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("mkdir engine/cards: %v", err)
	}
	yCode := `package cards

var Name = "Chainbreaker Phalanx"
`
	yPath := filepath.Join(subDir, "y.go")
	if err := os.WriteFile(yPath, []byte(yCode), 0644); err != nil {
		t.Fatalf("write cards/y.go: %v", err)
	}

	var ok bool
	var err error
	out := captureOutput(func() {
		ok, err = fences.Lint(repo, nil)
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !ok {
		t.Fatalf("expected subdirectories to be out of scope, but failed; output:\n%s", out)
	}
}

// 12. // OPEN(KIM-B07), where KIM-B07 is in ## Open Items -> passes.
func TestFences_Case12_MarkerOpenMatrixItem(t *testing.T) {
	repo := setupFixtureRepo(t)
	code := `package engine

// OPEN(KIM-B07)
func BerserkLogic() {}
`
	xPath := filepath.Join(repo, "pkg", "engine", "x.go")
	if err := os.WriteFile(xPath, []byte(code), 0644); err != nil {
		t.Fatalf("write x.go: %v", err)
	}

	var ok bool
	var err error
	out := captureOutput(func() {
		ok, err = fences.Lint(repo, nil)
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !ok {
		t.Fatalf("expected OPEN(KIM-B07) to pass, but failed; output:\n%s", out)
	}
}

// 13. // OPEN(KIM-B01), where KIM-B01 is only in ## Rulings -> fails, and the error says the item is ruled and the marker should be removed.
func TestFences_Case13_MarkerRuledMatrixItem(t *testing.T) {
	repo := setupFixtureRepo(t)
	code := `package engine

// OPEN(KIM-B01)
func WardLogic() {}
`
	xPath := filepath.Join(repo, "pkg", "engine", "x.go")
	if err := os.WriteFile(xPath, []byte(code), 0644); err != nil {
		t.Fatalf("write x.go: %v", err)
	}

	var ok bool
	var err error
	out := captureOutput(func() {
		ok, err = fences.Lint(repo, nil)
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if ok {
		t.Fatalf("expected failure for ruled KIM-B01, got true")
	}
	if !strings.Contains(out, "ruled") || !strings.Contains(out, "marker should be removed") {
		t.Fatalf("expected error saying item is ruled and marker should be removed, got:\n%s", out)
	}
}

// 14. // OPEN(KIM-Z99), which appears nowhere -> fails, and the error says the item is not in the matrix.
func TestFences_Case14_MarkerMissingMatrixItem(t *testing.T) {
	repo := setupFixtureRepo(t)
	code := `package engine

// OPEN(KIM-Z99)
func Unknown() {}
`
	xPath := filepath.Join(repo, "pkg", "engine", "x.go")
	if err := os.WriteFile(xPath, []byte(code), 0644); err != nil {
		t.Fatalf("write x.go: %v", err)
	}

	var ok bool
	var err error
	out := captureOutput(func() {
		ok, err = fences.Lint(repo, nil)
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if ok {
		t.Fatalf("expected failure for unknown KIM-Z99, got true")
	}
	if !strings.Contains(out, "not in the matrix") {
		t.Fatalf("expected error saying item is not in the matrix, got:\n%s", out)
	}
}

// 15. // OPEN(SVY-R7) with an open survey row **R7** -> passes. // OPEN(SVY-R4a) where **R4a**'s status is **Ruled: yes.** -> fails as ruled.
func TestFences_Case15_MarkerSurveyItems(t *testing.T) {
	repo := setupFixtureRepo(t)
	// Part 1: open survey item passes
	codeOpen := `package engine

// OPEN(SVY-R7)
func OpenSurvey() {}
`
	xPath := filepath.Join(repo, "pkg", "engine", "x.go")
	if err := os.WriteFile(xPath, []byte(codeOpen), 0644); err != nil {
		t.Fatalf("write x.go: %v", err)
	}

	var ok bool
	var err error
	out := captureOutput(func() {
		ok, err = fences.Lint(repo, nil)
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !ok {
		t.Fatalf("expected OPEN(SVY-R7) to pass, but failed; output:\n%s", out)
	}

	// Part 2: ruled survey item fails
	codeRuled := `package engine

// OPEN(SVY-R4a)
func RuledSurvey() {}
`
	if err := os.WriteFile(xPath, []byte(codeRuled), 0644); err != nil {
		t.Fatalf("write x.go: %v", err)
	}
	out = captureOutput(func() {
		ok, err = fences.Lint(repo, nil)
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if ok {
		t.Fatalf("expected failure for ruled SVY-R4a, got true")
	}
	if !strings.Contains(out, "ruled") {
		t.Fatalf("expected error saying item is ruled, got:\n%s", out)
	}
}

// 16. // OPEN(ABC-X01) -> fails, and the error says markers must name a KIM- or SVY- item.
func TestFences_Case16_MarkerInvalidPrefix(t *testing.T) {
	repo := setupFixtureRepo(t)
	code := `package engine

// OPEN(ABC-X01)
func BadPrefix() {}
`
	xPath := filepath.Join(repo, "pkg", "engine", "x.go")
	if err := os.WriteFile(xPath, []byte(code), 0644); err != nil {
		t.Fatalf("write x.go: %v", err)
	}

	var ok bool
	var err error
	out := captureOutput(func() {
		ok, err = fences.Lint(repo, nil)
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if ok {
		t.Fatalf("expected failure for ABC-X01, got true")
	}
	if !strings.Contains(out, "markers must name a KIM- or SVY- item") {
		t.Fatalf("expected error saying markers must name a KIM- or SVY- item, got:\n%s", out)
	}
}

// 17. // it with an "OPEN(KIM-...)" comment -> passes (not a marker).
func TestFences_Case17_MarkerPlaceholderIgnored(t *testing.T) {
	repo := setupFixtureRepo(t)
	code := `package engine

// it with an "OPEN(KIM-...)" comment
func DocExample() {}
`
	xPath := filepath.Join(repo, "pkg", "engine", "x.go")
	if err := os.WriteFile(xPath, []byte(code), 0644); err != nil {
		t.Fatalf("write x.go: %v", err)
	}

	var ok bool
	var err error
	out := captureOutput(func() {
		ok, err = fences.Lint(repo, nil)
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !ok {
		t.Fatalf("expected placeholder OPEN(KIM-...) to pass, but failed; output:\n%s", out)
	}
}

// 18. Two violations in different files -> both are reported, in a stable order (file name, then line). The summary counts 2.
func TestFences_Case18_TwoViolationsInDifferentFilesStableOrderAndCount(t *testing.T) {
	repo := setupFixtureRepo(t)
	aCode := `package engine

var _ = "Steel Rain"
`
	bCode := `package engine

var _ = "Chainbreaker Phalanx"
`
	if err := os.WriteFile(filepath.Join(repo, "pkg", "engine", "a.go"), []byte(aCode), 0644); err != nil {
		t.Fatalf("write a.go: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repo, "pkg", "engine", "b.go"), []byte(bCode), 0644); err != nil {
		t.Fatalf("write b.go: %v", err)
	}

	var ok bool
	var err error
	out := captureOutput(func() {
		ok, err = fences.Lint(repo, nil)
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if ok {
		t.Fatalf("expected failure, got true")
	}

	idxA := strings.Index(out, "a.go:3")
	idxB := strings.Index(out, "b.go:3")
	if idxA == -1 || idxB == -1 {
		t.Fatalf("expected both a.go:3 and b.go:3 in output, got:\n%s", out)
	}
	if idxA > idxB {
		t.Fatalf("expected a.go before b.go, got:\n%s", out)
	}
	if !strings.Contains(out, "2 fence violation(s) found") && !strings.Contains(out, "2 violation") {
		t.Fatalf("expected summary count 2 in output, got:\n%s", out)
	}
}

// 19. data/cards/ is missing -> Lint returns a non-nil error, not a pass.
func TestFences_Case19_MissingCardsDirectoryError(t *testing.T) {
	repo := setupFixtureRepo(t)
	cardsDir := filepath.Join(repo, "data", "cards")
	if err := os.RemoveAll(cardsDir); err != nil {
		t.Fatalf("remove cards dir: %v", err)
	}

	ok, err := fences.Lint(repo, nil)
	if err == nil {
		t.Fatalf("expected non-nil error when data/cards is missing, got nil (ok=%v)", ok)
	}
	if !strings.Contains(err.Error(), "cards") {
		t.Fatalf("expected error mentioning cards, got: %v", err)
	}
	if ok {
		t.Fatalf("expected ok=false when error occurs")
	}
}

// 20. The real repository (repoRoot = ../../../../, as in rules_test.go) -> passes. This pins that today's engine is clean once the Scratch Company exemption applies.
func TestFences_Case20_RealRepositoryClean(t *testing.T) {
	repoRoot, err := filepath.Abs("../../../../")
	if err != nil {
		t.Fatalf("failed to resolve repo root: %v", err)
	}

	ok, err := fences.Lint(repoRoot, nil)
	if err != nil {
		t.Fatalf("unexpected error during fences lint on real repo: %v", err)
	}
	if !ok {
		t.Fatalf("expected real repo to pass fences check, but failed")
	}
}

func writeEngineFile(t *testing.T, repo, rel, code string) {
	t.Helper()
	p := filepath.Join(repo, "pkg", "engine", filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		t.Fatalf("mkdir for %s: %v", rel, err)
	}
	if err := os.WriteFile(p, []byte(code), 0644); err != nil {
		t.Fatalf("write %s: %v", rel, err)
	}
}

func lintFixture(t *testing.T, repo string) (bool, error, string) {
	t.Helper()
	var ok bool
	var err error
	out := captureOutput(func() {
		ok, err = fences.Lint(repo, nil)
	})
	return ok, err, out
}

// 21. A card name in pkg/engine/ability/z.go -> fails, naming ability/z.go:<line>.
func TestFences_Case21_AbilityPackageInScope(t *testing.T) {
	repo := setupFixtureRepo(t)
	writeEngineFile(t, repo, "ability/z.go", "package ability\n\n// Steel Rain\nvar _ = 1\n")
	ok, err, out := lintFixture(t, repo)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if ok {
		t.Fatalf("expected failure, got true")
	}
	if !strings.Contains(out, "[ERROR] ability/z.go:3") {
		t.Fatalf("expected [ERROR] ability/z.go:3, got:\n%s", out)
	}
}

// 22. A card name in a new subfolder pkg/engine/foo/w.go -> fails.
func TestFences_Case22_NewSubfolderInScope(t *testing.T) {
	repo := setupFixtureRepo(t)
	writeEngineFile(t, repo, "foo/w.go", "package foo\n\nvar _ = \"Steel Rain\"\n")
	ok, err, out := lintFixture(t, repo)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if ok || !strings.Contains(out, "foo/w.go:3") {
		t.Fatalf("expected failure naming foo/w.go:3 (ok=%v), got:\n%s", ok, out)
	}
}

// 23. Cards subtree files and test files at any depth -> pass.
func TestFences_Case23_CardsSubtreeAndTestsExempt(t *testing.T) {
	repo := setupFixtureRepo(t)
	writeEngineFile(t, repo, "cards/sub/w.go", "package sub\n\nvar _ = \"Steel Rain\"\n")
	writeEngineFile(t, repo, "ability/z_test.go", "package ability\n\nvar _ = \"Steel Rain\"\n")
	ok, err, out := lintFixture(t, repo)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !ok {
		t.Fatalf("expected pass, got:\n%s", out)
	}
}

// 24. Missing matrix or survey -> Lint errors, naming the file.
func TestFences_Case24_MissingDocsError(t *testing.T) {
	for _, rel := range []string{
		filepath.Join("docs", "frd", "keyword_interaction_matrix.md"),
		filepath.Join("docs", "design", "m2_starter_set_ability_survey.md"),
	} {
		repo := setupFixtureRepo(t)
		if err := os.Remove(filepath.Join(repo, rel)); err != nil {
			t.Fatalf("remove %s: %v", rel, err)
		}
		ok, err, _ := lintFixture(t, repo)
		if err == nil {
			t.Fatalf("missing %s: expected error, got nil (ok=%v)", rel, ok)
		}
		if ok {
			t.Fatalf("missing %s: expected ok=false", rel)
		}
		if !strings.Contains(err.Error(), filepath.Base(rel)) {
			t.Fatalf("missing %s: error %q should name the file", rel, err)
		}
	}
}

// 25. Violations in x.go and ability/z.go -> both reported, sorted by path.
func TestFences_Case25_SortedByRelativePath(t *testing.T) {
	repo := setupFixtureRepo(t)
	writeEngineFile(t, repo, "x.go", "package engine\n\nvar _ = \"Steel Rain\"\n")
	writeEngineFile(t, repo, "ability/z.go", "package ability\n\nvar _ = \"Steel Rain\"\n")
	ok, err, out := lintFixture(t, repo)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if ok {
		t.Fatalf("expected failure")
	}
	iz, ix := strings.Index(out, "ability/z.go:3"), strings.Index(out, "x.go:3")
	if iz == -1 || ix == -1 || iz > ix {
		t.Fatalf("expected ability/z.go:3 before x.go:3, got:\n%s", out)
	}
}
