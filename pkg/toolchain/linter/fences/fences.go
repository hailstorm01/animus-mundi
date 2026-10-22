package fences

import (
	"bufio"
	"fmt"
	"go/parser"
	"go/scanner"
	"go/token"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"animus-mundi/pkg/toolchain/linter/cardsheet"
)

var (
	// Exemptions maps card names that FRD 01 also defines as game objects to their FR ID.
	Exemptions = map[string]string{
		"Scratch Company": "FR-ZON-005",
	}

	markerRegex = regexp.MustCompile(`OPEN\(([A-Z]+-[A-Z]+\d+[a-z]?)\)`)
)

// modulePath is the Go module path that import rules are written against.
const modulePath = "animus-mundi"

// allowedImports maps a package directory to the in-module import paths (relative to
// the module root) its non-test files may use. Packages not listed are not checked.
var allowedImports = map[string][]string{
	"pkg/engine/ability": {},
	"pkg/engine/cards":   {"pkg/engine/ability"},
	"pkg/sim":            {"pkg/engine"},
	"pkg/eventlog":       {"pkg/engine"},
	"pkg/view":           {"pkg/engine", "pkg/eventlog"},
}

// Report holds fence violations found in package code.
type Report struct {
	Violations []string
}

type violationItem struct {
	file string
	line int
	msg  string
}

type cardInfo struct {
	name  string
	words []string
}

// Analyze scans non-test Go files under pkg/ (except pkg/engine/cards and pkg/toolchain)
// for card names and invalid markers, and checks the imports of the packages in
// allowedImports (pkg/engine/cards included).
func Analyze(repoRoot string) (*Report, error) {
	cardsDir := filepath.Join(repoRoot, "data", "cards")
	info, err := os.Stat(cardsDir)
	if err != nil {
		return nil, fmt.Errorf("reading data/cards directory: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("data/cards is not a directory")
	}

	csvFiles, _, err := cardsheet.List(repoRoot, nil)
	if err != nil {
		return nil, err
	}

	cardNames := make(map[string]bool)
	for _, csvFile := range csvFiles {
		sheet, err := cardsheet.Open(csvFile, true)
		if err != nil {
			return nil, fmt.Errorf("opening card sheet %s: %w", filepath.Base(csvFile), err)
		}
		header, err := sheet.Header()
		var rows [][]string
		if err == nil {
			rows, err = sheet.Rows()
		}
		sheet.Close()
		if err == io.EOF {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("reading card sheet %s: %w", filepath.Base(csvFile), err)
		}
		nameCol := cardsheet.Column(header, "name")
		if nameCol == -1 {
			continue
		}
		for _, row := range rows {
			if nameCol < len(row) {
				cName := strings.TrimSpace(row[nameCol])
				if cName == "" || cName == "--" {
					continue
				}
				if _, exempt := Exemptions[cName]; exempt {
					continue
				}
				cardNames[cName] = true
			}
		}
	}

	var cards []cardInfo
	for name := range cardNames {
		words := splitWords(name)
		if len(words) > 0 {
			cards = append(cards, cardInfo{name: name, words: words})
		}
	}
	sort.Slice(cards, func(i, j int) bool {
		return cards[i].name < cards[j].name
	})

	openMatrix := make(map[string]bool)
	ruledMatrix := make(map[string]bool)
	matrixPath := filepath.Join(repoRoot, "docs", "frd", "keyword_interaction_matrix.md")
	mContent, err := os.ReadFile(matrixPath)
	if err != nil {
		return nil, fmt.Errorf("reading keyword_interaction_matrix.md: %w", err)
	}
	matrixLines := strings.Split(string(mContent), "\n")
	currentSec := ""
	for _, line := range matrixLines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "## ") {
			currentSec = trimmed
			continue
		}
		if !strings.HasPrefix(trimmed, "|") {
			continue
		}
		cells := strings.Split(trimmed, "|")
		if len(cells) < 3 {
			continue
		}
		id := strings.TrimSpace(cells[1])
		if strings.HasPrefix(id, "KIM-") {
			if strings.HasPrefix(currentSec, "## Open Items") {
				openMatrix[id] = true
			} else if strings.HasPrefix(currentSec, "## Rulings") {
				ruledMatrix[id] = true
			}
		}
	}

	openSurvey := make(map[string]bool)
	ruledSurvey := make(map[string]bool)
	surveyPath := filepath.Join(repoRoot, "docs", "design", "m2_starter_set_ability_survey.md")
	sContent, err := os.ReadFile(surveyPath)
	if err != nil {
		return nil, fmt.Errorf("reading m2_starter_set_ability_survey.md: %w", err)
	}
	surveyLines := strings.Split(string(sContent), "\n")
	inRulingQuestions := false
	for _, line := range surveyLines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "## 4. Ruling Questions") {
			inRulingQuestions = true
			continue
		}
		if inRulingQuestions && (strings.HasPrefix(trimmed, "## ") || strings.HasPrefix(trimmed, "---")) {
			inRulingQuestions = false
			continue
		}
		if !inRulingQuestions || !strings.HasPrefix(trimmed, "|") {
			continue
		}
		cells := strings.Split(trimmed, "|")
		if len(cells) < 4 {
			continue
		}
		rawID := strings.TrimSpace(cells[1])
		cleanID := strings.Trim(rawID, "* ")
		if cleanID == "" || strings.EqualFold(cleanID, "id") || strings.HasPrefix(cleanID, ":-") {
			continue
		}
		svyID := "SVY-" + cleanID
		lastCell := strings.TrimSpace(cells[len(cells)-2])
		if strings.Contains(lastCell, "Ruled") {
			ruledSurvey[svyID] = true
		} else {
			openSurvey[svyID] = true
		}
	}

	pkgDir := filepath.Join(repoRoot, "pkg")
	var files []string
	err = filepath.WalkDir(pkgDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(repoRoot, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if d.IsDir() {
			if rel == "pkg/engine/cards" || rel == "pkg/toolchain" {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(rel, ".go") && !strings.HasSuffix(rel, "_test.go") {
			files = append(files, rel)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("reading pkg directory: %w", err)
	}

	var items []violationItem

	for _, name := range files {
		lines, err := readLines(filepath.Join(repoRoot, filepath.FromSlash(name)))
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", name, err)
		}

		for idx, line := range lines {
			lineNum := idx + 1

			// Check markers
			markers := markerRegex.FindAllStringSubmatch(line, -1)
			for _, m := range markers {
				markerID := m[1]
				parts := strings.SplitN(markerID, "-", 2)
				prefix := parts[0]
				switch prefix {
				case "KIM":
					if openMatrix[markerID] {
						// Open item: valid marker
					} else if ruledMatrix[markerID] {
						items = append(items, violationItem{
							file: name,
							line: lineNum,
							msg:  fmt.Sprintf("OPEN(%s) is ruled and the marker should be removed", markerID),
						})
					} else {
						items = append(items, violationItem{
							file: name,
							line: lineNum,
							msg:  fmt.Sprintf("OPEN(%s) is not in the matrix", markerID),
						})
					}
				case "SVY":
					if openSurvey[markerID] {
						// Open survey item: valid marker
					} else if ruledSurvey[markerID] {
						items = append(items, violationItem{
							file: name,
							line: lineNum,
							msg:  fmt.Sprintf("OPEN(%s) is ruled and the marker should be removed", markerID),
						})
					} else {
						items = append(items, violationItem{
							file: name,
							line: lineNum,
							msg:  fmt.Sprintf("OPEN(%s) is not in the survey", markerID),
						})
					}
				default:
					items = append(items, violationItem{
						file: name,
						line: lineNum,
						msg:  fmt.Sprintf("OPEN(%s): markers must name a KIM- or SVY- item", markerID),
					})
				}
			}

			// Check card names
			lineWords := splitWords(line)
			if len(lineWords) > 0 {
				for _, c := range cards {
					if containsWordSequence(lineWords, c.words) {
						items = append(items, violationItem{
							file: name,
							line: lineNum,
							msg:  fmt.Sprintf("contains card name %q", c.name),
						})
					}
				}
			}
		}
	}

	// Import direction also applies to the cards package, which the scan above skips.
	importFiles := append([]string(nil), files...)
	cardsFiles, err := goFilesUnder(repoRoot, "pkg/engine/cards")
	if err != nil {
		return nil, err
	}
	importFiles = append(importFiles, cardsFiles...)
	for _, name := range importFiles {
		items = append(items, checkImports(repoRoot, name)...)
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].file != items[j].file {
			return items[i].file < items[j].file
		}
		if items[i].line != items[j].line {
			return items[i].line < items[j].line
		}
		return items[i].msg < items[j].msg
	})

	rep := &Report{}
	for _, it := range items {
		rep.Violations = append(rep.Violations, fmt.Sprintf("%s:%d: %s", it.file, it.line, it.msg))
	}
	return rep, nil
}

// Lint runs the fences check over package code under pkg/ and prints the results.
func Lint(repoRoot string, _ []string) (bool, error) {
	fmt.Println("==========================================================")
	fmt.Println("  CARD TOWN LINTER: ENGINE ARCHITECTURE FENCES            ")
	fmt.Println("==========================================================")
	rep, err := Analyze(repoRoot)
	if err != nil {
		return false, err
	}

	if len(rep.Violations) > 0 {
		for _, v := range rep.Violations {
			fmt.Printf("  [ERROR] %s\n", v)
		}
		fmt.Printf("\n  [FAIL] %d fence violation(s) found.\n", len(rep.Violations))
		fmt.Println("==========================================================")
		return false, nil
	}

	fmt.Println("  [PASS] Package code is free of card names, invalid markers and disallowed imports.")
	fmt.Println("==========================================================")
	return true, nil
}

// goFilesUnder lists non-test Go files under a repo-relative directory, slash-separated.
func goFilesUnder(repoRoot, relDir string) ([]string, error) {
	root := filepath.Join(repoRoot, filepath.FromSlash(relDir))
	if _, err := os.Stat(root); os.IsNotExist(err) {
		return nil, nil
	}
	var out []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		rel, err := filepath.Rel(repoRoot, path)
		if err != nil {
			return err
		}
		out = append(out, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", relDir, err)
	}
	return out, nil
}

// checkImports reports imports in a table-listed package's file that leave its allowed set.
func checkImports(repoRoot, rel string) []violationItem {
	pkgPath := rel[:strings.LastIndex(rel, "/")]
	allowed, listed := allowedImports[pkgPath]
	if !listed {
		return nil
	}
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filepath.Join(repoRoot, filepath.FromSlash(rel)), nil, parser.ImportsOnly)
	if err != nil {
		line := 1
		if el, ok := err.(scanner.ErrorList); ok && len(el) > 0 {
			line = el[0].Pos.Line
			return []violationItem{{file: rel, line: line, msg: "cannot read imports: " + el[0].Msg}}
		}
		return []violationItem{{file: rel, line: line, msg: fmt.Sprintf("cannot read imports: %v", err)}}
	}
	var items []violationItem
	for _, spec := range f.Imports {
		path, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			continue
		}
		if path != modulePath && !strings.HasPrefix(path, modulePath+"/") {
			continue
		}
		inModule := strings.TrimPrefix(strings.TrimPrefix(path, modulePath), "/")
		ok := false
		for _, a := range allowed {
			if inModule == a {
				ok = true
				break
			}
		}
		if ok {
			continue
		}
		rule := "may import nothing from this module"
		if len(allowed) > 0 {
			rule = "may import only " + strings.Join(allowed, ", ")
		}
		items = append(items, violationItem{
			file: rel,
			line: fset.Position(spec.Pos()).Line,
			msg:  fmt.Sprintf("imports %s; %s %s", path, pkgPath, rule),
		})
	}
	return items
}

func splitWords(s string) []string {
	var b strings.Builder
	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		b.WriteRune(runes[i])
		if i+1 < len(runes) {
			r1 := runes[i]
			r2 := runes[i+1]
			isLowerUpper := unicode.IsLower(r1) && unicode.IsUpper(r2)
			isLetterDigit := (unicode.IsLetter(r1) && unicode.IsDigit(r2)) || (unicode.IsDigit(r1) && unicode.IsLetter(r2))
			if isLowerUpper || isLetterDigit {
				b.WriteRune(' ')
			}
		}
	}
	expanded := strings.ToLower(b.String())
	expanded = strings.ReplaceAll(expanded, "'", "")
	expanded = strings.ReplaceAll(expanded, "’", "")
	return strings.FieldsFunc(expanded, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
}

func containsWordSequence(haystack, needle []string) bool {
	if len(needle) == 0 || len(needle) > len(haystack) {
		return false
	}
	for i := 0; i <= len(haystack)-len(needle); i++ {
		match := true
		for j := 0; j < len(needle); j++ {
			if haystack[i+j] != needle[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, 1024*1024)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}
