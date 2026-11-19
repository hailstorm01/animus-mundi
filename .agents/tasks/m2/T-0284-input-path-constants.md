---
id: T-0284
title: "Each linter input path is written once: `specdoc.FRDDir`, `specdoc.RulesFRD`, `cardsheet.Dir` (directive #172 (a); wave H.2 harness, flash, a move)"
kind: harness
tier: flash
wave: H.2 (board directive #172 (a))
status: in-progress
branch: m2/autoloop-wave-h.2/T-0284-input-path-constants
depends_on: [T-0283]
claims: [const:specdoc.FRDDir, const:specdoc.RulesFRD, const:cardsheet.Dir]
cards: []
requirements: []
---

## Goal
The toolchain writes the repository's input paths in many places: `data/cards` is joined in four linters, `cmd/cardtown`'s `play`, `replay` and `sim`, and `whats-up`; `docs/frd` in `specdoc`, the FRD linter, the rules linter, `whats-up` and `cmd/cardtown/lint.go`; FRD 01's file name in `frd.RulesFRD`, `frd.ExpectedFiles`, `rules.RulesFRD` and `lint.go`'s `isRulesPath`. After this task, `specdoc` holds `FRDDir` (`docs/frd`) and `RulesFRD` (FRD 01's file name), `cardsheet` holds `Dir` (`data/cards`), and every one of those places reads them. `frd.RulesFRD`, `rules.RulesFRD` and `frd.KeywordMatrix` go, with every reference updated and no alias ([ADR 0007](../../../docs/architecture/decisions/0007-moves-update-every-reference.md)). It is a move: the same tests pass, each linter's output is unchanged, and match output is byte for byte the same. #172 (b) (`fences.Reads`, `rules.Reads`) is [T-0285](T-0285-linters-say-what-they-read.md), `queued` behind this task. When both merge, #172 is `applied`.

## Inputs
- Board directive #172 (structural review 14, finding 4), approved by the user 2026-11-19, `open: before the next Deliverables task`; wave H.2, after #171 (T-0283, `done`). Design note Deliverable 6o ([m2-ability-model.md:231](../../../docs/architecture/m2-ability-model.md)) names it. The directive allows (a) and (b) to be two tasks under the D2 size rule, (a) first. This is (a).
- Where it is today (at `autoloop/m2-wave-h.2` 7d092ac): a search for `data/cards`, `docs/frd`, `"data", "cards"`, `"docs", "frd"` and `01_Game_Rules_FRD` in non-test Go under `pkg/toolchain` and `cmd` finds `lint.go:218`, `:270`, `:272`, `:286`, `:287`; `play.go:53`, `replay.go:56`, `sim.go:55`; `whatsup.go:125`, `:168`, `:173`; `cards/sheets.go:58`; `cardsheet.go:34`; `fences.go:123`; `frd.go:15`, `:38`; `ownership.go:14`; `rules.go:19`; `syntax.go:25`; `specdoc/matrix.go:20`; plus comments and printed messages, which stay (the directive excepts them).
- Design (lead decisions):
  - The constants are slash-separated (`"docs/frd"`, `"data/cards"`), because `lint.go` compares slash-separated paths with them, `git status` takes them as a pathspec, and `filepath.Join` cleans the slash to the OS separator. So the linters print the same `FRD Directory:` and `Cards Directory:` lines as before (the lead compared them).
  - `specdoc.FRDDir` and `specdoc.RulesFRD` go in `specdoc.go`, the package's main file, and the package doc names them ("where the FRDs are"). `RulesFRD`'s comment keeps what `frd.RulesFRD`'s and `rules.RulesFRD`'s comments said.
  - `rules.RulesFRD` (the path `docs/frd/01_Game_Rules_FRD.md`, read only at `rules.go:33`) goes too: kept beside `specdoc.RulesFRD`, the same name would mean a path in one package and a file name in the other. `Analyze` joins `specdoc.FRDDir` and `specdoc.RulesFRD`. `rules.EngineDir` stays; #172 (b) uses it.
  - `frd.KeywordMatrix` (`= specdoc.MatrixFile`) goes too, and its two readers read `specdoc.MatrixFile`. It is a second name for a declaration in another package, which ADR 0007 rules out, and it sits in the same `const` block as `frd.RulesFRD`, so leaving it would keep the block for it alone. `frozenMarker` stays, as a single `const`.
  - `play.go`, `replay.go` and `sim.go` read `cardsheet.Dir` too: the directive's acceptance covers all non-test Go under `cmd`. `cmd/cardtown` already imports four linter packages; `cardsheet` imports nothing from the repo.
  - Printed messages and comments that say `data/cards` or `docs/frd` stay as they are, so no output changes (`lint.go:137`, `fences.go:126`, `:129`).
- **Authorized test edit** (#172; D2.4): `pkg/toolchain/linter/frd/ownership_test.go` (Edit 1), reference updates only, in the `test(T-0284)` commit: the map keys `RulesFRD` and `KeywordMatrix` in `TestCheckOwnershipFlagsDuplicateDefinitions`, `TestCheckOwnershipFlagsKeywordRestatementOutsideFRD01` and `TestCheckOwnershipSkipsFrozenFRDsAndTheMatrix` become `specdoc.RulesFRD` and `specdoc.MatrixFile`, and the file imports `specdoc`. At that commit the test references a name that doesn't exist yet, so `go vet ./pkg/toolchain/linter/frd/` fails (case 1), as in T-0283.
- Reuse search (D2.3): the three constants replace the existing equivalents `frd.RulesFRD`, `rules.RulesFRD` and the literals; nothing else declares these paths (`git grep` above; `pkg/engine` has no constant for `data/cards`, and `engine.LoadCards` takes the directory as an argument).
- Size (D2.4): two non-test files gain a declaration (`specdoc.go`, `cardsheet.go`); the other twelve change only to read the moved names (a move's reference updates don't count toward the four, ADR 0007).
- Docs: no doc names `frd.RulesFRD`, `rules.RulesFRD` or `frd.KeywordMatrix` (searched outside the board archives and run files). #176 (b) updates the docs for wave H.2 as a whole.
- **Expected output difference:** none. `go tool task sim-compare` passes, and `cardtown lint --all` prints the same lines as on the run branch apart from the worktree path (the lead compared them byte for byte with the root path masked). `cardtown whats-up` differs only in its uncommitted-changes count.
- The lead dry-ran these exact edits in a throwaway worktree of the run branch at 7d092ac: `gofmt -l pkg cmd` clean, `go vet ./pkg/toolchain/... ./cmd/...` clean, `go tool task check` exit 0, `go tool task sim-compare -- autoloop/m2-wave-h.2` `[PASS]`; with only Edit 1 applied, `go vet ./pkg/toolchain/linter/frd/` fails with `ownership_test.go:18:11: undefined: specdoc.RulesFRD`. Blobs: `lint.go` 0f89815, `play.go` d50e327, `replay.go` 341ba9b, `sim.go` a74d87f, `whatsup.go` 55dff4e, `sheets.go` 8f97c49, `cardsheet.go` 997bf21, `fences.go` 9198068, `frd.go` b3e3682, `ownership.go` 1f968f4, `ownership_test.go` 8a7a23b, `rules.go` a6aebde, `syntax.go` b29f7a8, `matrix.go` 37b1c28, `specdoc.go` e6f26d5.

## Commits
1. `test(T-0284): ownership tests read specdoc.RulesFRD and specdoc.MatrixFile` with Edit 1 only. At this commit the `frd` package's tests do not build (case 1).
2. `refactor(toolchain): each linter input path is written once, specdoc.FRDDir, specdoc.RulesFRD, cardsheet.Dir (T-0284)` with Edits 2-15.

No double quotes in either message: the allow-list refuses an escaped one.

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)
### 1. `pkg/toolchain/linter/frd/ownership_test.go` (commit 1, reference only)
```diff
@@ -3,6 +3,8 @@ package frd
 import (
 	"strings"
 	"testing"
+
+	"animus-mundi/pkg/toolchain/specdoc"
 )
 
 var rulesFixture = []string{
@@ -13,7 +15,7 @@ var rulesFixture = []string{
 
 func TestCheckOwnershipFlagsDuplicateDefinitions(t *testing.T) {
 	problems, _ := CheckOwnership(map[string][]string{
-		RulesFRD:           rulesFixture,
+		specdoc.RulesFRD:   rulesFixture,
 		"05_Engine_FRD.md": {"#### FR-CBT-002: Combat again", "See `FR-KWD-002o`."},
 	})
 	if len(problems) != 1 || !strings.Contains(problems[0], "FR-CBT-002") {
@@ -23,7 +25,7 @@ func TestCheckOwnershipFlagsDuplicateDefinitions(t *testing.T) {
 
 func TestCheckOwnershipFlagsKeywordRestatementOutsideFRD01(t *testing.T) {
 	problems, _ := CheckOwnership(map[string][]string{
-		RulesFRD:           rulesFixture,
+		specdoc.RulesFRD:   rulesFixture,
 		"06_Server_FRD.md": {"| **Berserk N** | cleaves only on a kill |"},
 	})
 	if len(problems) != 1 || !strings.Contains(problems[0], `"Berserk"`) {
@@ -33,9 +35,9 @@ func TestCheckOwnershipFlagsKeywordRestatementOutsideFRD01(t *testing.T) {
 
 func TestCheckOwnershipSkipsFrozenFRDsAndTheMatrix(t *testing.T) {
 	problems, frozen := CheckOwnership(map[string][]string{
-		RulesFRD:           rulesFixture,
+		specdoc.RulesFRD:   rulesFixture,
 		"02_Data_FRD.md":   {"> **Status: FROZEN as design intent.**", "| **Stalwart** | KEYWORD_GRANT |"},
-		KeywordMatrix:      {"- **Stalwart**: quoted to pose a question"},
+		specdoc.MatrixFile: {"- **Stalwart**: quoted to pose a question"},
 		"07_Client_FRD.md": {"Units with **Stalwart** show a shield icon."},
 	})
 	if len(problems) != 0 || frozen != 1 {
```

### 2. `cmd/cardtown/lint.go` (commit 2, reference only)
```diff
@@ -11,6 +11,7 @@ import (
 
 	"animus-mundi/pkg/toolchain/git"
 	"animus-mundi/pkg/toolchain/linter/cards"
+	"animus-mundi/pkg/toolchain/linter/cardsheet"
 	"animus-mundi/pkg/toolchain/linter/fences"
 	"animus-mundi/pkg/toolchain/linter/frd"
 	"animus-mundi/pkg/toolchain/linter/rules"
@@ -215,7 +216,7 @@ func runLint(args []string) {
 
 // isRulesPath reports whether a change can affect executable rules traceability.
 func isRulesPath(pLower string) bool {
-	return pLower == "docs/frd/01_game_rules_frd.md" || (strings.HasPrefix(pLower, "pkg/engine/") && strings.HasSuffix(pLower, ".go"))
+	return pLower == strings.ToLower(specdoc.FRDDir+"/"+specdoc.RulesFRD) || (strings.HasPrefix(pLower, "pkg/engine/") && strings.HasSuffix(pLower, ".go"))
 }
 
 // lintSelection is which linters a lint run runs, and the changed FRD documents and card
@@ -267,9 +268,9 @@ func selectChanged(paths []string, gitErr error) lintSelection {
 		if isFencesPath(pLower) {
 			s.fences = true
 		}
-		if strings.HasPrefix(pLower, "docs/frd/") && strings.HasSuffix(pLower, ".md") {
+		if strings.HasPrefix(pLower, specdoc.FRDDir+"/") && strings.HasSuffix(pLower, ".md") {
 			s.changedFrds = append(s.changedFrds, path)
-		} else if strings.HasPrefix(pLower, "data/cards/") && strings.HasSuffix(pLower, ".csv") {
+		} else if strings.HasPrefix(pLower, cardsheet.Dir+"/") && strings.HasSuffix(pLower, ".csv") {
 			s.changedCsvs = append(s.changedCsvs, path)
 		}
 	}
@@ -283,7 +284,7 @@ func selectChanged(paths []string, gitErr error) lintSelection {
 // pkg/, cmd/ or tools/ (the alias check reads them all), plus the data inputs.
 func isFencesPath(pLower string) bool {
 	return ((strings.HasPrefix(pLower, "pkg/") || strings.HasPrefix(pLower, "cmd/") || strings.HasPrefix(pLower, "tools/")) && strings.HasSuffix(pLower, ".go")) ||
-		(strings.HasPrefix(pLower, "data/cards/") && strings.HasSuffix(pLower, ".csv")) ||
-		pLower == "docs/frd/"+specdoc.MatrixFile ||
+		(strings.HasPrefix(pLower, cardsheet.Dir+"/") && strings.HasSuffix(pLower, ".csv")) ||
+		pLower == specdoc.FRDDir+"/"+specdoc.MatrixFile ||
 		pLower == "docs/design/"+fences.SurveyFile
 }
```

### 3. `cmd/cardtown/play.go` (commit 2, reference only)
```diff
@@ -12,6 +12,7 @@ import (
 
 	"animus-mundi/pkg/engine"
 	"animus-mundi/pkg/sim"
+	"animus-mundi/pkg/toolchain/linter/cardsheet"
 	"animus-mundi/pkg/view"
 )
 
@@ -50,7 +51,7 @@ func runPlay(args []string, in io.Reader, out io.Writer) (exitCode int) {
 	}
 
 	repoRoot := findRepoRoot()
-	cards, err := engine.LoadCards(filepath.Join(repoRoot, "data", "cards"))
+	cards, err := engine.LoadCards(filepath.Join(repoRoot, cardsheet.Dir))
 	if err != nil {
 		fmt.Fprintf(out, "[ERROR] loading cards: %v\n", err)
 		return 1
```

### 4. `cmd/cardtown/replay.go` (commit 2, reference only)
```diff
@@ -12,6 +12,7 @@ import (
 	"animus-mundi/pkg/engine"
 	"animus-mundi/pkg/eventlog"
 	"animus-mundi/pkg/sim"
+	"animus-mundi/pkg/toolchain/linter/cardsheet"
 )
 
 // runReplay rebuilds one recorded game from a `cardtown sim -events` file and
@@ -53,7 +54,7 @@ func runReplay(args []string) {
 	}
 
 	repoRoot := findRepoRoot()
-	cards, err := engine.LoadCards(filepath.Join(repoRoot, "data", "cards"))
+	cards, err := engine.LoadCards(filepath.Join(repoRoot, cardsheet.Dir))
 	if err != nil {
 		fail("loading cards: %v", err)
 	}
```

### 5. `cmd/cardtown/sim.go` (commit 2, reference only)
```diff
@@ -13,6 +13,7 @@ import (
 	"animus-mundi/pkg/engine"
 	"animus-mundi/pkg/eventlog"
 	"animus-mundi/pkg/sim"
+	"animus-mundi/pkg/toolchain/linter/cardsheet"
 )
 
 func runSim(args []string) {
@@ -52,7 +53,7 @@ func runSim(args []string) {
 	}
 
 	repoRoot := findRepoRoot()
-	cards, err := engine.LoadCards(filepath.Join(repoRoot, "data", "cards"))
+	cards, err := engine.LoadCards(filepath.Join(repoRoot, cardsheet.Dir))
 	if err != nil {
 		fmt.Printf("[ERROR] loading cards: %v\n", err)
 		os.Exit(1)
```

### 6. `pkg/toolchain/docs/whatsup.go` (commit 2, reference only)
```diff
@@ -10,6 +10,7 @@ import (
 	"strings"
 
 	"animus-mundi/pkg/toolchain/git"
+	"animus-mundi/pkg/toolchain/linter/cardsheet"
 	"animus-mundi/pkg/toolchain/linter/rules"
 	"animus-mundi/pkg/toolchain/specdoc"
 )
@@ -122,7 +123,7 @@ func ShowState(repoRoot string) {
 	}
 	fmt.Printf("Git: branch '%s', %s unpushed commit(s)\n", branch, ahead)
 
-	changed := git.Lines(repoRoot, "status", "--porcelain", "--", "docs", "data/cards", "tools", ".agents", "cmd", "pkg")
+	changed := git.Lines(repoRoot, "status", "--porcelain", "--", "docs", cardsheet.Dir, "tools", ".agents", "cmd", "pkg")
 	if len(changed) == 0 {
 		fmt.Println("Uncommitted changes: none")
 	} else {
@@ -165,12 +166,12 @@ func ShowState(repoRoot string) {
 		sumLog := git.Lines(repoRoot, "log", "-1", "--format=%ct", "--", "docs/PRODUCT_SUMMARY.md")
 		if len(sumLog) > 0 {
 			if sumTime, err := strconv.ParseInt(sumLog[0], 10, 64); err == nil {
-				frdDir := filepath.Join(repoRoot, "docs", "frd")
+				frdDir := filepath.Join(repoRoot, specdoc.FRDDir)
 				var staleFrds []string
 				if frdEntries, err := os.ReadDir(frdDir); err == nil {
 					for _, f := range frdEntries {
 						if !f.IsDir() && strings.HasSuffix(f.Name(), ".md") {
-							relFrd := "docs/frd/" + f.Name()
+							relFrd := specdoc.FRDDir + "/" + f.Name()
 							frdLog := git.Lines(repoRoot, "log", "-1", "--format=%ct", "--", relFrd)
 							if len(frdLog) > 0 {
 								if frdTime, err := strconv.ParseInt(frdLog[0], 10, 64); err == nil && frdTime > sumTime {
```

### 7. `pkg/toolchain/linter/cards/sheets.go` (commit 2, reference only)
```diff
@@ -55,7 +55,7 @@ var (
 
 // Lint audits CSV card sheet files for schema and enum compliance.
 func Lint(repoRoot string, targetFilePaths []string) (bool, error) {
-	cardsDir := filepath.Join(repoRoot, "data", "cards")
+	cardsDir := filepath.Join(repoRoot, cardsheet.Dir)
 
 	fmt.Println("==========================================================")
 	fmt.Println("   CARD TOWN LINTER: CARD SHEET CSV INTEGRITY & SCHEMA    ")
```

### 8. `pkg/toolchain/linter/cardsheet/cardsheet.go` (commit 2: cardsheet.Dir)
```diff
@@ -10,6 +10,10 @@ import (
 	"strings"
 )
 
+// Dir is the directory that holds the card sheets, relative to the repository
+// root and slash-separated. Every cardtown command that reads the cards joins it.
+const Dir = "data/cards"
+
 // List returns the sorted card sheet files to audit. With paths, each trimmed non-blank
 // path (joined to repoRoot when relative) that exists goes in files; the others go in
 // missing, as trimmed, in input order. Without paths it lists data/cards/*.csv.
@@ -31,7 +35,7 @@ func List(repoRoot string, paths []string) (files, missing []string, err error)
 			}
 		}
 	} else {
-		cardsDir := filepath.Join(repoRoot, "data", "cards")
+		cardsDir := filepath.Join(repoRoot, Dir)
 		entries, readErr := os.ReadDir(cardsDir)
 		if readErr != nil {
 			return nil, nil, fmt.Errorf("reading cards directory: %w", readErr)
```

### 9. `pkg/toolchain/linter/fences/fences.go` (commit 2, reference only)
```diff
@@ -120,7 +120,7 @@ func Analyze(repoRoot string) (*Report, error) {
 
 // loadCardNames reads every card name in data/cards except the exempt ones, sorted by name.
 func loadCardNames(repoRoot string) ([]cardInfo, error) {
-	cardsDir := filepath.Join(repoRoot, "data", "cards")
+	cardsDir := filepath.Join(repoRoot, cardsheet.Dir)
 	info, err := os.Stat(cardsDir)
 	if err != nil {
 		return nil, fmt.Errorf("reading data/cards directory: %w", err)
```

### 10. `pkg/toolchain/linter/frd/frd.go` (commit 2, reference only)
```diff
@@ -12,7 +12,7 @@ import (
 
 var (
 	ExpectedFiles = []string{
-		"01_Game_Rules_FRD.md",
+		specdoc.RulesFRD,
 		"02_Card_Data_Pipeline_FRD.md",
 		"03_Card_Compiler_FRD.md",
 		"04_Client_Engine_Protocol_Networking_FRD.md",
@@ -35,7 +35,7 @@ type ReqRef struct {
 
 // Lint runs the FRD requirement cross-reference audit.
 func Lint(repoRoot string, targetFilePaths []string) (bool, error) {
-	frdDir := filepath.Join(repoRoot, "docs", "frd")
+	frdDir := filepath.Join(repoRoot, specdoc.FRDDir)
 
 	fmt.Println("==========================================================")
 	fmt.Println("  CARD TOWN LINTER: FRD REQUIREMENTS & CROSS-REFERENCES   ")
```

### 11. `pkg/toolchain/linter/frd/ownership.go` (commit 2: RulesFRD and KeywordMatrix go)
```diff
@@ -9,14 +9,8 @@ import (
 	"animus-mundi/pkg/toolchain/specdoc"
 )
 
-const (
-	// RulesFRD owns every game rule and the keyword dictionary.
-	RulesFRD = "01_Game_Rules_FRD.md"
-	// KeywordMatrix quotes keyword rules on purpose to pose questions about them.
-	KeywordMatrix = specdoc.MatrixFile
-	// frozenMarker identifies FRDs frozen as design intent.
-	frozenMarker = "**Status: FROZEN"
-)
+// frozenMarker identifies FRDs frozen as design intent.
+const frozenMarker = "**Status: FROZEN"
 
 var (
 	ownerRegex      = regexp.MustCompile(`^(?:#### |- \*\*)(FR-[A-Z]{3}-\d{3}[a-z]?)`)
@@ -57,12 +51,12 @@ func CheckOwnership(contents map[string][]string) (problems []string, frozenRest
 		}
 	}
 
-	restatement := keywordRestatementRegex(contents[RulesFRD])
+	restatement := keywordRestatementRegex(contents[specdoc.RulesFRD])
 	if restatement == nil {
 		return problems, 0
 	}
 	for _, name := range names {
-		if name == RulesFRD || name == KeywordMatrix {
+		if name == specdoc.RulesFRD || name == specdoc.MatrixFile {
 			continue
 		}
 		frozen := false
@@ -81,7 +75,7 @@ func CheckOwnership(contents map[string][]string) (problems []string, frozenRest
 				frozenRestatements++
 				continue
 			}
-			problems = append(problems, fmt.Sprintf("%s:%d restates keyword %q; define keywords only in %s Appendix A and cite their FR-KWD ID", name, i+1, m[1], RulesFRD))
+			problems = append(problems, fmt.Sprintf("%s:%d restates keyword %q; define keywords only in %s Appendix A and cite their FR-KWD ID", name, i+1, m[1], specdoc.RulesFRD))
 		}
 	}
 	return problems, frozenRestatements
```

### 12. `pkg/toolchain/linter/rules/rules.go` (commit 2: RulesFRD goes)
```diff
@@ -15,8 +15,6 @@ var (
 	testNameRegex = regexp.MustCompile(`^func (TestFR_([A-Z]{3})_(\d{3}[a-z]?)\w*)\(`)
 	coversRegex   = regexp.MustCompile(`^// Covers: (.+)$`)
 
-	// RulesFRD is the single source of truth for game rules.
-	RulesFRD = filepath.Join("docs", "frd", "01_Game_Rules_FRD.md")
 	// EngineDir holds the executable rules and their scenario tests.
 	EngineDir = filepath.Join("pkg", "engine")
 )
@@ -30,7 +28,7 @@ type Coverage struct {
 
 // Analyze reads FRD 01 and the engine's TestFR_* scenario tests.
 func Analyze(repoRoot string) (*Coverage, error) {
-	defined, subsumed, err := frdRequirements(filepath.Join(repoRoot, RulesFRD))
+	defined, subsumed, err := frdRequirements(filepath.Join(repoRoot, specdoc.FRDDir, specdoc.RulesFRD))
 	if err != nil {
 		return nil, err
 	}
```

### 13. `pkg/toolchain/linter/syntax/syntax.go` (commit 2, reference only)
```diff
@@ -22,7 +22,7 @@ var forbiddenRules = []SyntaxRule{
 
 // Lint audits card rules text for forbidden terminology and syntax shorthand.
 func Lint(repoRoot string, targetFilePaths []string, treatWarningsAsErrors bool) (bool, error) {
-	cardsDir := filepath.Join(repoRoot, "data", "cards")
+	cardsDir := filepath.Join(repoRoot, cardsheet.Dir)
 
 	fmt.Println("==========================================================")
 	fmt.Println("   CARD TOWN LINTER: CARD RULES SYNTAX & TERMINOLOGY      ")
```

### 14. `pkg/toolchain/specdoc/matrix.go` (commit 2, reference only)
```diff
@@ -17,7 +17,7 @@ type MatrixItem struct {
 
 // ReadMatrix returns the KIM- rows of the keyword matrix's Open Items and Rulings tables, in file order.
 func ReadMatrix(repoRoot string) ([]MatrixItem, error) {
-	lines, err := ReadLines(filepath.Join(repoRoot, "docs", "frd", MatrixFile))
+	lines, err := ReadLines(filepath.Join(repoRoot, FRDDir, MatrixFile))
 	if err != nil {
 		return nil, err
 	}
```

### 15. `pkg/toolchain/specdoc/specdoc.go` (commit 2: FRDDir and RulesFRD)
```diff
@@ -1,4 +1,4 @@
-// Package specdoc holds what the toolchain shares when it scans a file: requirement IDs, the lines that define them, reading a file as lines, and the keyword matrix's items.
+// Package specdoc holds what the toolchain shares when it scans a file: where the FRDs are, requirement IDs, the lines that define them, reading a file as lines, and the keyword matrix's items.
 package specdoc
 
 import (
@@ -7,6 +7,15 @@ import (
 	"regexp"
 )
 
+const (
+	// FRDDir is the directory that holds the FRDs and the keyword matrix,
+	// relative to the repository root and slash-separated.
+	FRDDir = "docs/frd"
+	// RulesFRD is FRD 01's file name, in FRDDir: the single source of truth for
+	// game rules, which owns every game rule and the keyword dictionary.
+	RulesFRD = "01_Game_Rules_FRD.md"
+)
+
 var (
 	// ReqRegex matches a requirement ID.
 	ReqRegex = regexp.MustCompile(`FR-[A-Z]{3}-\d{3}[a-z]?`)
```


## Acceptance Cases (harness tasks only; written by the lead)
W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **The test edit fails before the implementation:** at the `test(T-0284)` commit, `go -C W vet ./pkg/toolchain/linter/frd/` fails, reporting `undefined: specdoc.RulesFRD` in `ownership_test.go`.
2. **Each path is written once:** `git -C W grep -n -i -e "data/cards" -e "docs/frd" -e "data., .cards" -e "docs., .frd" -e "01_game_rules_frd" -- pkg/toolchain cmd ":!*_test.go"` prints exactly ten lines: `cmd/cardtown/lint.go:138` (a printed message), `pkg/toolchain/linter/cardsheet/cardsheet.go:1`, `:15` and `:19`, `pkg/toolchain/linter/fences/fences.go:121`, `:126` and `:129` (a comment and two error messages), `pkg/toolchain/specdoc/matrix.go:8`, and `pkg/toolchain/specdoc/specdoc.go:13` and `:16`.
3. **The old names are gone:** `git -C W grep -n -w KeywordMatrix -- "*.go"` prints nothing, and `git -C W grep -n -w RulesFRD -- "*.go"` prints exactly eleven lines, each `specdoc.RulesFRD` except the declaration and its comment: `cmd/cardtown/lint.go:219`, `pkg/toolchain/linter/frd/frd.go:15`, `pkg/toolchain/linter/frd/ownership.go:54`, `:59` and `:78`, `pkg/toolchain/linter/frd/ownership_test.go:18`, `:28` and `:38`, `pkg/toolchain/linter/rules/rules.go:31`, and `pkg/toolchain/specdoc/specdoc.go:14` and `:16`.
4. **The same tests pass, with only the authorized edit:** `go -C W test ./pkg/... ./cmd/...` passes, and `git -C W diff autoloop/m2-wave-h.2 --numstat -- "*_test.go"` prints exactly one line: `6	4	pkg/toolchain/linter/frd/ownership_test.go`.
5. **The linters read the same files:** `go -C W run ./cmd/cardtown lint --all` exits 0 and ends with `LINTER SUMMARY: All executed checks PASSED!`; its `FRD Directory:` line ends in `docs\frd` and both `Cards Directory:` lines end in `data\cards`.
6. **Sims byte-identical:** `go -C W tool task sim-compare -- autoloop/m2-wave-h.2` prints `[PASS]`.
7. **Only the listed edits:** `git -C W diff autoloop/m2-wave-h.2 --numstat -- pkg cmd docs` prints exactly fifteen lines: `6	5	cmd/cardtown/lint.go`, `2	1	cmd/cardtown/play.go`, `2	1	cmd/cardtown/replay.go`, `2	1	cmd/cardtown/sim.go`, `4	3	pkg/toolchain/docs/whatsup.go`, `1	1	pkg/toolchain/linter/cards/sheets.go`, `5	1	pkg/toolchain/linter/cardsheet/cardsheet.go`, `1	1	pkg/toolchain/linter/fences/fences.go`, `2	2	pkg/toolchain/linter/frd/frd.go`, `5	11	pkg/toolchain/linter/frd/ownership.go`, `6	4	pkg/toolchain/linter/frd/ownership_test.go`, `1	3	pkg/toolchain/linter/rules/rules.go`, `1	1	pkg/toolchain/linter/syntax/syntax.go`, `1	1	pkg/toolchain/specdoc/matrix.go` and `10	1	pkg/toolchain/specdoc/specdoc.go`.
8. **Nothing else changed:** `git -C W diff autoloop/m2-wave-h.2 --stat` lists only the fifteen files of case 7 and this task file.
9. `gofmt -l W/pkg W/cmd` prints nothing; `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/toolchain/specdoc/specdoc.go` (Edit 15: `FRDDir`, `RulesFRD`)
- `pkg/toolchain/linter/cardsheet/cardsheet.go` (Edit 8: `Dir`)
- Reference only (a move's reference updates, ADR 0007): `cmd/cardtown/lint.go` (Edit 2), `cmd/cardtown/play.go` (Edit 3), `cmd/cardtown/replay.go` (Edit 4), `cmd/cardtown/sim.go` (Edit 5), `pkg/toolchain/docs/whatsup.go` (Edit 6), `pkg/toolchain/linter/cards/sheets.go` (Edit 7), `pkg/toolchain/linter/fences/fences.go` (Edit 9), `pkg/toolchain/linter/frd/frd.go` (Edit 10), `pkg/toolchain/linter/frd/ownership.go` (Edit 11: `RulesFRD` and `KeywordMatrix` go), `pkg/toolchain/linter/rules/rules.go` (Edit 12: `RulesFRD` goes), `pkg/toolchain/linter/syntax/syntax.go` (Edit 13), `pkg/toolchain/specdoc/matrix.go` (Edit 14)

## Tests
Authorized reference-only edit (#172), in the `test(T-0284)` commit; no test is added, renamed or removed:
- `pkg/toolchain/linter/frd/ownership_test.go` (Edit 1): `TestCheckOwnershipFlagsDuplicateDefinitions`, `TestCheckOwnershipFlagsKeywordRestatementOutsideFRD01`, `TestCheckOwnershipSkipsFrozenFRDsAndTheMatrix` (map keys only), and the file's imports

## Done When
- [ ] Cases 1-9 hold, with their commands and results in the Log
- [ ] `go tool task sim-compare -- autoloop/m2-wave-h.2` passes (starter and stress sims byte-identical to the run branch)
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-19, lead-architect (wave H.2 run, session 4): written in full against 7d092ac for #172 (a), next in wave H.2's order after #171 (T-0283, `done`). `main` (1859fe4) has nothing the run branch lacks, so no sync. Not `WAVE-COMPLETE` before the task: #172, #173 and #174 are `open: before the next Deliverables task`. #172 splits into two tasks as the directive allows: this one, and T-0285 (`queued`, board row and task file in the same commit, D2.4) for (b). This is a harness task, so it skips `tests` and goes straight to `ready`. Tier is `flash`, because the edits are exact and were dry-run. Every command here is on the headless allow-list: one per line, no pipes, no `$`, and patterns with special characters are quoted. Waiting directives checked: #6, #7 (mode `serial`), #15 (2) (no draw or end-of-turn hook), #71, #91 (2), #106, #120, #142 (wave H.2 has no cards): none has fired. Next: `kernel-engineer` on Flash.
