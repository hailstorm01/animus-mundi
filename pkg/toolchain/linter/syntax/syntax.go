package syntax

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"animus-mundi/pkg/toolchain/linter/cardsheet"
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

	csvFiles, missing, err := cardsheet.List(repoRoot, targetFilePaths)
	if err != nil {
		return false, err
	}
	for _, m := range missing {
		fmt.Printf("  [ERROR] Specified file not found: %s\n", m)
	}

	if len(csvFiles) == 0 {
		fmt.Println("  [WARN] No CSV card sheets found to audit.")
		return true, nil
	}

	fmt.Printf("  Auditing rules text across %d CSV sheet(s)...\n\n", len(csvFiles))

	terminologyIssues := 0

	for _, csvPath := range csvFiles {
		fileName := filepath.Base(csvPath)
		sheet, err := cardsheet.Open(csvPath, false)
		if err != nil {
			fmt.Printf("  [ERROR] %s: Could not read for syntax check: %v\n", fileName, err)
			terminologyIssues++
			continue
		}

		headers, err := sheet.Header()
		if err != nil {
			sheet.Close()
			fmt.Printf("  [ERROR] %s: Could not read header: %v\n", fileName, err)
			terminologyIssues++
			continue
		}

		nameCol := cardsheet.Column(headers, "name")
		rulesCol := cardsheet.Column(headers, "rulestext", "text")

		records, err := sheet.Rows()
		sheet.Close()
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
