package rules

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var (
	reqRegex      = regexp.MustCompile(`FR-[A-Z]{3}-\d{3}[a-z]?`)
	defLineRegex  = regexp.MustCompile(`^(####|###|\*\*FR-|- \*\*FR-|\| \*\*FR-)`)
	subsumedRow   = regexp.MustCompile(`^\| \*\*(FR-[A-Z]{3}-\d{3}[a-z]?)\*\*.*Subsumed by`)
	testNameRegex = regexp.MustCompile(`^func (TestFR_([A-Z]{3})_(\d{3}[a-z]?)\w*)\(`)
	coversRegex   = regexp.MustCompile(`^// Covers: (.+)$`)

	// RulesFRD is the single source of truth for game rules.
	RulesFRD = filepath.Join("docs", "frd", "01_Game_Rules_FRD.md")
	// EngineDir holds the executable rules and their scenario tests.
	EngineDir = filepath.Join("pkg", "engine")
)

// Coverage maps each active FRD 01 requirement to the scenario tests that exercise it.
type Coverage struct {
	Active   []string
	Tests    map[string][]string
	Problems []string
}

// Analyze reads FRD 01 and the engine's TestFR_* scenario tests.
func Analyze(repoRoot string) (*Coverage, error) {
	defined, subsumed, err := frdRequirements(filepath.Join(repoRoot, RulesFRD))
	if err != nil {
		return nil, err
	}
	cov := &Coverage{Tests: map[string][]string{}}
	for id := range defined {
		if !subsumed[id] {
			cov.Active = append(cov.Active, id)
		}
	}
	sort.Strings(cov.Active)

	files, err := filepath.Glob(filepath.Join(repoRoot, EngineDir, "*_test.go"))
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	for _, f := range files {
		lines, err := readLines(f)
		if err != nil {
			return nil, err
		}
		// A "// Covers: FR-..., FR-..." comment directly above a test adds
		// the requirements it also exercises beyond the one in its name.
		var covers []string
		for i, line := range lines {
			if c := coversRegex.FindStringSubmatch(line); c != nil {
				covers = append(covers, reqRegex.FindAllString(c[1], -1)...)
				continue
			}
			m := testNameRegex.FindStringSubmatch(line)
			if m == nil {
				if !strings.HasPrefix(line, "//") {
					covers = nil
				}
				continue
			}
			where := fmt.Sprintf("%s:%d", filepath.Base(f), i+1)
			for _, id := range append([]string{fmt.Sprintf("FR-%s-%s", m[2], m[3])}, covers...) {
				switch {
				case !defined[id]:
					cov.Problems = append(cov.Problems, fmt.Sprintf("%s (%s) names %s, which FRD 01 does not define", m[1], where, id))
				case subsumed[id]:
					cov.Problems = append(cov.Problems, fmt.Sprintf("%s (%s) names %s, which is subsumed; use its canonical owner", m[1], where, id))
				default:
					cov.Tests[id] = append(cov.Tests[id], m[1])
				}
			}
			covers = nil
		}
	}
	return cov, nil
}

// Lint fails when a scenario test names a requirement FRD 01 does not own as
// active, and reports which active requirements have no executable test yet.
func Lint(repoRoot string, _ []string) (bool, error) {
	fmt.Println("==========================================================")
	fmt.Println("  CARD TOWN LINTER: EXECUTABLE RULES TRACEABILITY          ")
	fmt.Println("==========================================================")
	cov, err := Analyze(repoRoot)
	if err != nil {
		return false, err
	}

	var untested []string
	for _, id := range cov.Active {
		if len(cov.Tests[id]) == 0 {
			untested = append(untested, id)
		}
	}
	covered := len(cov.Active) - len(untested)
	fmt.Printf("  %d of %d active FRD 01 requirements have a TestFR_* scenario test.\n", covered, len(cov.Active))
	if len(untested) > 0 {
		fmt.Printf("  Not yet executable: %s\n", strings.Join(untested, ", "))
	}
	fmt.Println()

	if len(cov.Problems) > 0 {
		for _, p := range cov.Problems {
			fmt.Printf("  [ERROR] %s\n", p)
		}
		fmt.Printf("\n  [FAIL] %d scenario test name(s) do not trace to an active FRD 01 requirement.\n", len(cov.Problems))
		fmt.Println("==========================================================")
		return false, nil
	}
	fmt.Println("  [PASS] Every scenario test traces to an active FRD 01 requirement.")
	fmt.Println("==========================================================")
	return true, nil
}

func frdRequirements(path string) (defined, subsumed map[string]bool, err error) {
	lines, err := readLines(path)
	if err != nil {
		return nil, nil, err
	}
	defined, subsumed = map[string]bool{}, map[string]bool{}
	for _, line := range lines {
		if m := subsumedRow.FindStringSubmatch(line); m != nil {
			subsumed[m[1]] = true
		}
		if defLineRegex.MatchString(line) {
			for _, id := range reqRegex.FindAllString(line, -1) {
				defined[id] = true
			}
		}
	}
	return defined, subsumed, nil
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
