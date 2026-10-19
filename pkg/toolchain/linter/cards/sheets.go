package cards

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var (
	legalCardTypes = map[string]bool{
		"UNIT":         true,
		"STRUCTURE":    true,
		"HEADQUARTERS": true,
		"MANEUVER":     true,
	}
	legalFactions = map[string]bool{
		"CONCORDIA": true,
		"SOULLESS":  true,
		"PATH":      true,
		"ANIMUS":    true,
		"NEUTRAL":   true,
	}
	legalRarities = map[string]bool{
		"COMMON":   true,
		"UNCOMMON": true,
		"RARE":     true,
		"MYTHIC":   true,
		"HQ":       true,
	}
	legalZones = map[string]bool{
		"RESERVES":               true,
		"MAIN_THEATER":           true,
		"HEADQUARTERS":           true,
		"MAIN_THEATER, RESERVES": true,
	}
	legalBooleans = map[string]bool{
		"true":  true,
		"false": true,
	}

	expectedAssetHeaders = []string{
		"name", "cardType", "faction", "subtypes", "requisitionCost", "tacticsCost",
		"deploymentZones", "rulesText", "power", "health", "requisitionYield",
		"tacticsYield", "rarity", "currentProxy", "needsRework",
	}
	expectedManeuverHeaders = []string{
		"name", "cardType", "faction", "requisitionCost", "tacticsCost",
		"rulesText", "rarity", "currentProxy", "needsRework",
	}
)

// Lint audits CSV card sheet files for schema and enum compliance.
func Lint(repoRoot string, targetFilePaths []string) (bool, error) {
	cardsDir := filepath.Join(repoRoot, "data", "cards")

	fmt.Println("==========================================================")
	fmt.Println("   CARD TOWN LINTER: CARD SHEET CSV INTEGRITY & SCHEMA    ")
	fmt.Println("==========================================================")
	fmt.Println("Repo Root:", repoRoot)
	fmt.Printf("Cards Directory: %s\n\n", cardsDir)

	hasErrors := false

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
				hasErrors = true
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

	fmt.Printf("  Found %d CSV sheet(s) to audit.\n", len(csvFiles))

	for _, csvPath := range csvFiles {
		fileName := filepath.Base(csvPath)
		file, err := os.Open(csvPath)
		if err != nil {
			fmt.Printf("  [FAIL] %s: Could not open file: %v\n", fileName, err)
			hasErrors = true
			continue
		}

		r := csv.NewReader(file)
		// Read header line
		headerParts, err := r.Read()
		if err == io.EOF {
			file.Close()
			fmt.Printf("  [FAIL] %s: File is empty!\n", fileName)
			hasErrors = true
			continue
		} else if err != nil {
			file.Close()
			fmt.Printf("  [FAIL] %s: Header parse error: %v\n", fileName, err)
			hasErrors = true
			continue
		}

		// Check for duplicate headers
		counts := make(map[string]int)
		for _, h := range headerParts {
			counts[strings.TrimSpace(h)]++
		}
		var dupeHeaders []string
		for h, c := range counts {
			if c > 1 {
				dupeHeaders = append(dupeHeaders, h)
			}
		}
		if len(dupeHeaders) > 0 {
			sort.Strings(dupeHeaders)
			fmt.Printf("  [FAIL] %s: Duplicate column headers found: %s\n", fileName, strings.Join(dupeHeaders, ", "))
			hasErrors = true
			file.Close()
			continue
		}

		// Check for corrupted 'xz' column
		hasXZ := false
		for _, h := range headerParts {
			if strings.TrimSpace(h) == "xz" {
				hasXZ = true
				break
			}
		}
		if hasXZ {
			fmt.Printf("  [FAIL] %s: Corrupted column header 'xz' detected!\n", fileName)
			hasErrors = true
			file.Close()
			continue
		}

		// Enforce exact header schema
		expectedHeaders := expectedAssetHeaders
		if strings.HasPrefix(fileName, "maneuvers-") {
			expectedHeaders = expectedManeuverHeaders
		}

		joinedActual := strings.Join(headerParts, ",")
		joinedExpected := strings.Join(expectedHeaders, ",")
		if joinedActual != joinedExpected {
			fmt.Printf("  [HEADER ERROR] %s: Headers do not match canonical camelCase schema.\n", fileName)
			fmt.Printf("    Expected: %s\n", strings.Join(expectedHeaders, ", "))
			fmt.Printf("    Found:    %s\n", strings.Join(headerParts, ", "))
			hasErrors = true
			file.Close()
			continue
		}

		// Build header-to-index lookup
		colIdx := make(map[string]int)
		for idx, h := range headerParts {
			colIdx[h] = idx
		}

		// Read all remaining rows
		records, err := r.ReadAll()
		file.Close()
		if err != nil {
			fmt.Printf("  [FAIL] %s: Import error: %v\n", fileName, err)
			hasErrors = true
			continue
		}

		csvErrors := 0
		for i, rec := range records {
			rowNum := i + 2
			getVal := func(col string) string {
				if idx, ok := colIdx[col]; ok && idx < len(rec) {
					return rec[idx]
				}
				return ""
			}

			name := getVal("name")
			if strings.TrimSpace(name) == "" {
				// Skip empty placeholder rows awaiting final design
				continue
			}

			cType := getVal("cardType")
			if cType != "" && !legalCardTypes[cType] {
				fmt.Printf("  [ENUM ERROR] %s:%d -> Invalid cardType '%s'\n", fileName, rowNum, cType)
				csvErrors++
			}

			faction := getVal("faction")
			if faction != "" && !legalFactions[faction] {
				fmt.Printf("  [ENUM ERROR] %s:%d -> Invalid faction '%s'\n", fileName, rowNum, faction)
				csvErrors++
			}

			rarity := getVal("rarity")
			if rarity != "" && !legalRarities[rarity] {
				fmt.Printf("  [ENUM ERROR] %s:%d -> Invalid rarity '%s'\n", fileName, rowNum, rarity)
				csvErrors++
			}

			if _, hasZone := colIdx["deploymentZones"]; hasZone {
				zones := getVal("deploymentZones")
				if zones != "" && !legalZones[zones] {
					fmt.Printf("  [ENUM ERROR] %s:%d -> Invalid deploymentZones '%s'\n", fileName, rowNum, zones)
					csvErrors++
				}
			}

			proxy := getVal("currentProxy")
			if !legalBooleans[proxy] {
				fmt.Printf("  [BOOL ERROR] %s:%d -> Invalid currentProxy '%s' (must be 'true' or 'false')\n", fileName, rowNum, proxy)
				csvErrors++
			}

			rework := getVal("needsRework")
			if !legalBooleans[rework] {
				fmt.Printf("  [BOOL ERROR] %s:%d -> Invalid needsRework '%s' (must be 'true' or 'false')\n", fileName, rowNum, rework)
				csvErrors++
			}
		}

		if csvErrors == 0 {
			fmt.Printf("  [PASS] %s (%d rows, %d columns)\n", fileName, len(records), len(headerParts))
		} else {
			hasErrors = true
		}
	}

	fmt.Println()
	if !hasErrors {
		fmt.Println("  [PASS] All card sheet CSVs parsed cleanly with valid schemas!")
		fmt.Println("==========================================================")
		return true, nil
	}

	fmt.Println("  [FAIL] Card sheet CSV audit failed. Resolve highlighted errors above.")
	fmt.Println("==========================================================")
	return false, nil
}
