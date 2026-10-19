package frd

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
	reqRegex     = regexp.MustCompile(`FR-[A-Z]{3}-\d{3}[a-z]?`)
	defLineRegex = regexp.MustCompile(`^(####|###|\*\*FR-|- \*\*FR-|\| \*\*FR-)`)

	ExpectedFiles = []string{
		"01_Game_Rules_FRD.md",
		"02_Card_Data_Pipeline_FRD.md",
		"03_Card_Compiler_FRD.md",
		"04_Client_Engine_Protocol_Networking_FRD.md",
		"05_Headless_Game_Engine_FRD.md",
	}
)

type ReqDef struct {
	File string
	Line int
	Text string
}

type ReqRef struct {
	Req  string
	File string
	Line int
	Text string
}

// Lint runs the FRD requirement cross-reference audit.
func Lint(repoRoot string, targetFilePaths []string) (bool, error) {
	frdDir := filepath.Join(repoRoot, "docs", "frd")

	fmt.Println("==========================================================")
	fmt.Println("  CARD TOWN LINTER: FRD REQUIREMENTS & CROSS-REFERENCES   ")
	fmt.Println("==========================================================")
	fmt.Println("Repo Root:", repoRoot)
	fmt.Printf("FRD Directory: %s\n\n", frdDir)

	hasErrors := false

	// Check expected core FRD files exist
	for _, ef := range ExpectedFiles {
		p := filepath.Join(frdDir, ef)
		if _, err := os.Stat(p); os.IsNotExist(err) {
			fmt.Printf("  [ERROR] Missing expected core FRD file: %s\n", ef)
			hasErrors = true
		}
	}

	// Discover all existing FRD md files
	entries, err := os.ReadDir(frdDir)
	if err != nil {
		return false, fmt.Errorf("reading frd directory: %w", err)
	}

	var existingFrds []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			existingFrds = append(existingFrds, filepath.Join(frdDir, e.Name()))
		}
	}
	sort.Strings(existingFrds)

	// Determine target files to inspect
	var targetFiles []string
	if len(targetFilePaths) > 0 {
		for _, f := range targetFilePaths {
			trimmed := strings.TrimSpace(f)
			if trimmed == "" {
				continue
			}
			p := trimmed
			if !filepath.IsAbs(p) {
				p = filepath.Join(repoRoot, p)
			}
			if _, err := os.Stat(p); err == nil {
				targetFiles = append(targetFiles, p)
			} else {
				fmt.Printf("  [ERROR] Specified file not found: %s\n", trimmed)
				hasErrors = true
			}
		}
	} else {
		targetFiles = existingFrds
	}

	// Step 1: Scan ALL existing FRDs to build the complete requirement definition catalog
	reqDefs := make(map[string]ReqDef)
	for _, filePath := range existingFrds {
		fileName := filepath.Base(filePath)
		lines, err := readLines(filePath)
		if err != nil {
			fmt.Printf("  [ERROR] Reading file %s: %v\n", fileName, err)
			hasErrors = true
			continue
		}

		for i, line := range lines {
			matches := reqRegex.FindAllString(line, -1)
			if len(matches) > 0 && defLineRegex.MatchString(line) {
				for _, val := range matches {
					if _, exists := reqDefs[val]; !exists {
						reqDefs[val] = ReqDef{
							File: fileName,
							Line: i + 1,
							Text: strings.TrimSpace(line),
						}
					}
				}
			}
		}
	}

	fmt.Printf("  Discovered %d defined requirement IDs across %d FRD files.\n", len(reqDefs), len(existingFrds))

	// Step 2: Scan target file(s) for references to validate against the master definitions
	var reqRefs []ReqRef
	for _, filePath := range targetFiles {
		fileName := filepath.Base(filePath)
		lines, err := readLines(filePath)
		if err != nil {
			fmt.Printf("  [ERROR] Reading file %s: %v\n", fileName, err)
			hasErrors = true
			continue
		}

		for i, line := range lines {
			matches := reqRegex.FindAllString(line, -1)
			for _, val := range matches {
				reqRefs = append(reqRefs, ReqRef{
					Req:  val,
					File: fileName,
					Line: i + 1,
					Text: strings.TrimSpace(line),
				})
			}
		}
	}

	fmt.Printf("  Checking %d requirement references across %d target FRD file(s)...\n\n", len(reqRefs), len(targetFiles))

	// Step 3: Check for dangling references
	uniqueRefMap := make(map[string]bool)
	for _, r := range reqRefs {
		uniqueRefMap[r.Req] = true
	}
	var uniqueRefs []string
	for r := range uniqueRefMap {
		uniqueRefs = append(uniqueRefs, r)
	}
	sort.Strings(uniqueRefs)

	var dangling []string
	for _, r := range uniqueRefs {
		if _, exists := reqDefs[r]; !exists {
			dangling = append(dangling, r)
			hasErrors = true
			fmt.Printf("  [DANGLING REFERENCE] %s\n", r)
			for _, ref := range reqRefs {
				if ref.Req == r {
					fmt.Printf("    -> Referenced in %s:%d\n", ref.File, ref.Line)
				}
			}
		}
	}

	fmt.Println()
	if len(dangling) == 0 && !hasErrors {
		fmt.Println("  [PASS] Zero dangling references! All requirement cross-references resolve cleanly.")
		fmt.Println("==========================================================")
		return true, nil
	}

	fmt.Printf("  [FAIL] Found %d dangling references or missing core FRD files.\n", len(dangling))
	fmt.Println("==========================================================")
	return false, nil
}

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	// Accommodate very long lines if needed (up to 1MB)
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, 1024*1024)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}
