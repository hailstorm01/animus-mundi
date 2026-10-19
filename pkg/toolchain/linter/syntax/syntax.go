package syntax

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type SyntaxRule struct {
	Regex   *regexp.Regexp
	Message string
}

var forbiddenRules = []SyntaxRule{
	{Regex: regexp.MustCompile(`(?i)\bfriendly\b`), Message: "Use 'allied' instead of 'friendly'"},
	{Regex: regexp.MustCompile(`~\b`), Message: "Raw '~' MTG shorthand detected"},
	{Regex: regexp.MustCompile(`(?i)\+\d+\s+attack\b`), Message: "Use '+N power' instead of '+N attack'"},
}

// Lint audits card rules text for forbidden terminology and syntax shorthand.
func Lint(repoRoot string, targetFilePaths []string, treatWarningsAsErrors bool) (bool, error) {
	cardsDir := filepath.Join(repoRoot, "data", "cards")

	fmt.Println("==========================================================")
	fmt.Println("   CARD TOWN LINTER: CARD RULES SYNTAX & TERMINOLOGY      ")
	fmt.Println("==========================================================")
	fmt.Println("Repo Root:", repoRoot)
	fmt.Printf("Cards Directory: %s\n\n", cardsDir)

	var csvFiles []string
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
				csvFiles = append(csvFiles, p)
			} else {
				fmt.Printf("  [ERROR] Specified file not found: %s\n", trimmed)
			}
		}
	} else {
		entries, err := os.ReadDir(cardsDir)
		if err != nil {
			return false, fmt.Errorf("reading cards directory: %w", err)
		}
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".csv") {
				csvFiles = append(csvFiles, filepath.Join(cardsDir, e.Name()))
			}
		}
	}
	sort.Strings(csvFiles)

	if len(csvFiles) == 0 {
		fmt.Println("  [WARN] No CSV card sheets found to audit.")
		return true, nil
	}

	fmt.Printf("  Auditing rules text across %d CSV sheet(s)...\n\n", len(csvFiles))

	terminologyIssues := 0

	for _, csvPath := range csvFiles {
		fileName := filepath.Base(csvPath)
		file, err := os.Open(csvPath)
		if err != nil {
			fmt.Printf("  [ERROR] %s: Could not read for syntax check: %v\n", fileName, err)
			terminologyIssues++
			continue
		}

		r := csv.NewReader(file)
		headers, err := r.Read()
		if err != nil {
			file.Close()
			fmt.Printf("  [ERROR] %s: Could not read header: %v\n", fileName, err)
			terminologyIssues++
			continue
		}

		nameCol := -1
		rulesCol := -1
		for idx, h := range headers {
			hLower := strings.ToLower(strings.TrimSpace(h))
			if hLower == "name" {
				nameCol = idx
			} else if hLower == "rulestext" || hLower == "text" {
				rulesCol = idx
			}
		}

		records, err := r.ReadAll()
		file.Close()
		if err != nil {
			fmt.Printf("  [ERROR] %s: CSV read error: %v\n", fileName, err)
			terminologyIssues++
			continue
		}

		for i, rec := range records {
			cardName := fmt.Sprintf("Row %d", i+2)
			if nameCol != -1 && nameCol < len(rec) && strings.TrimSpace(rec[nameCol]) != "" {
				cardName = strings.TrimSpace(rec[nameCol])
			}

			rulesText := ""
			if rulesCol != -1 && rulesCol < len(rec) {
				rulesText = rec[rulesCol]
			}

			if strings.TrimSpace(rulesText) != "" {
				for _, rule := range forbiddenRules {
					if rule.Regex.MatchString(rulesText) {
						fmt.Printf("  [SYNTAX ISSUE] %s -> %s: %s (Text: %q)\n", fileName, cardName, rule.Message, rulesText)
						terminologyIssues++
					}
				}
			}
		}
	}

	fmt.Println()
	if terminologyIssues == 0 {
		fmt.Println("  [PASS] All card rules text conforms to standard engine terminology!")
		fmt.Println("==========================================================")
		return true, nil
	}

	fmt.Printf("  [WARN] Found %d terminology / syntax issue(s).\n", terminologyIssues)
	fmt.Println("==========================================================")
	if treatWarningsAsErrors {
		return false, nil
	}
	return true, nil
}
