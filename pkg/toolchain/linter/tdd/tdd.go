package tdd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"animus-mundi/pkg/toolchain/git"
)

var (
	reqRegex        = regexp.MustCompile(`FR-[A-Z]{3}-\d{3}[a-z]?`)
	defLineRegex    = regexp.MustCompile(`^(####|###|\*\*FR-|- \*\*FR-|\| \*\*FR-)`)
	headingReqRegex = regexp.MustCompile(`^(?:####|###)\s+(FR-[A-Z]{3}-\d{3}[a-z]?)`)
	bulletReqRegex  = regexp.MustCompile(`^-\s+\*\*(FR-[A-Z]{3}-\d{3}[a-z]?)`)
	tableReqRegex   = regexp.MustCompile(`^\|\s*(?:\*\*|` + "`" + `)*(FR-[A-Z]{3}-\d{3}[a-z]?)(?:\*\*|` + "`" + `)*\s*\|`)
	subsumedRegex   = regexp.MustCompile(`(?i)subsumed\s+by`)

	sourceFrdHeaderRegex   = regexp.MustCompile(`^\s*-\s+\*\*Source FRD:\*\*\s*(.+)$`)
	commitHashHeaderRegex  = regexp.MustCompile(`^\s*-\s+\*\*FRD Commit Hash:\*\*\s*` + "`" + `?([^` + "`" + `\s]+)` + "`" + `?`)
	tddVersionHeaderRegex  = regexp.MustCompile(`^\s*-\s+\*\*TDD Version:\*\*\s*` + "`" + `?([^` + "`" + `\r\n]+)` + "`" + `?`)
	ghMilestoneHeaderRegex = regexp.MustCompile(`^\s*-\s+\*\*GitHub Milestone:\*\*\s*` + "`" + `?([^` + "`" + `\r\n]+)` + "`" + `?`)
	frdFileExtractRegex    = regexp.MustCompile(`(\d{2}_[A-Za-z0-9_]+_FRD\.md)`)

	sliceHeadingRegex   = regexp.MustCompile(`^###\s+\[(SLICE-\d{2}-\d{2})\]\s+(.+)$`)
	sliceSubHeaderRegex = regexp.MustCompile(`^####\s+(.+)$`)
	sliceIdRegex        = regexp.MustCompile(`^\s*-\s+\*\*Slice ID:\*\*\s*` + "`" + `?([^` + "`" + `\r\n]+)` + "`" + `?`)
	labelsRegex         = regexp.MustCompile(`^\s*-\s+\*\*Labels:\*\*\s*(.+)$`)
	dependsOnRegex      = regexp.MustCompile(`^\s*-\s+\*\*Depends On:\*\*\s*(.+)$`)
	reqCoveredRegex     = regexp.MustCompile(`^\s*-\s+\*\*Requirements Covered:\*\*\s*(.+)$`)
	tddSectionsRegex    = regexp.MustCompile(`^\s*-\s+\*\*TDD Sections:\*\*\s*(.+)$`)
	targetFilesRegex    = regexp.MustCompile(`^\s*-\s+\*\*Target Files:\*\*\s*(.*)$`)
	targetFileSubBullet = regexp.MustCompile(`^\s+-\s+(.+)$`)
	depSliceExtract     = regexp.MustCompile(`SLICE-\d{2}-\d{2}`)
	checkboxRegex       = regexp.MustCompile(`^\s*-\s+\[\s*[xX ]?\s*\]\s+\S+`)
	noneRegex           = regexp.MustCompile(`(?i)^` + "`" + `?None` + "`" + `?$`)
)

type SliceInfo struct {
	ID                   string
	Title                string
	Line                 int
	DeclaredID           string
	Labels               string
	DependsOnRaw         string
	DependsOnList        []string
	RequirementsCovered  []string
	TddSections          string
	TargetFiles          []string
	InTargetFiles        bool
	HasBlueprintSection  bool
	BlueprintLines       int
	HasAcceptanceSection bool
	AcceptanceCheckboxes int
	HasVerifySection     bool
	VerifyLines          int
}

// Lint runs the TDD requirement coverage, cross-references, and slice DAG linter.
func Lint(repoRoot string, targetFilePaths []string) (bool, error) {
	frdDir := filepath.Join(repoRoot, "docs", "frd")
	tddDir := filepath.Join(repoRoot, "docs", "tdd")

	fmt.Println("==========================================================")
	fmt.Println("  CARD TOWN LINTER: TDD COVERAGE, REFS & SLICE MANIFEST   ")
	fmt.Println("==========================================================")
	fmt.Println("Repo Root:", repoRoot)
	fmt.Printf("TDD Directory: %s\n\n", tddDir)

	hasErrors := false

	// Collect target TDD files
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
				fmt.Printf("  [ERROR] Specified TDD file not found: %s\n", trimmed)
				hasErrors = true
			}
		}
	} else {
		if _, err := os.Stat(tddDir); err == nil {
			entries, _ := os.ReadDir(tddDir)
			for _, e := range entries {
				if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") && e.Name() != "README.md" {
					targetFiles = append(targetFiles, filepath.Join(tddDir, e.Name()))
				}
			}
		}
	}
	sort.Strings(targetFiles)

	if len(targetFiles) == 0 && !hasErrors {
		fmt.Println("  No TDD documents found in docs/tdd/ yet. Skipping TDD lint.")
		fmt.Println("==========================================================")
		return true, nil
	}

	// Step 1: Scan all FRDs in docs/frd/ to build master requirement catalog & per-FRD ownership
	frdEntries, err := os.ReadDir(frdDir)
	if err != nil {
		return false, fmt.Errorf("reading frd directory: %w", err)
	}

	var existingFrds []string
	for _, e := range frdEntries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			existingFrds = append(existingFrds, filepath.Join(frdDir, e.Name()))
		}
	}
	sort.Strings(existingFrds)

	allReqDefs := make(map[string]bool)
	frdOwnedReqs := make(map[string][]string)
	subsumedReqs := make(map[string]bool)

	for _, frdPath := range existingFrds {
		fn := filepath.Base(frdPath)
		frdOwnedReqs[fn] = []string{}
		lines, err := readLines(frdPath)
		if err != nil {
			continue
		}

		for i, line := range lines {
			// 1a. Master catalog matching
			if defLineRegex.MatchString(line) {
				for _, m := range reqRegex.FindAllString(line, -1) {
					allReqDefs[m] = true
				}
			}

			// 1b. Per-FRD ownership & Active vs Subsumed classification
			var ownedID string
			isSub := false

			if m := headingReqRegex.FindStringSubmatch(line); len(m) > 1 {
				ownedID = m[1]
				next1 := ""
				next2 := ""
				if i+1 < len(lines) {
					next1 = lines[i+1]
				}
				if i+2 < len(lines) {
					next2 = lines[i+2]
				}
				if subsumedRegex.MatchString(line) || subsumedRegex.MatchString(next1) || subsumedRegex.MatchString(next2) {
					isSub = true
				}
			} else if m := bulletReqRegex.FindStringSubmatch(line); len(m) > 1 {
				ownedID = m[1]
				if subsumedRegex.MatchString(line) {
					isSub = true
				}
			} else if m := tableReqRegex.FindStringSubmatch(line); len(m) > 1 {
				candidate := m[1]
				alreadyOwnedElsewhere := false
				for k, list := range frdOwnedReqs {
					if k != fn {
						for _, item := range list {
							if item == candidate {
								alreadyOwnedElsewhere = true
								break
							}
						}
					}
					if alreadyOwnedElsewhere {
						break
					}
				}
				if !alreadyOwnedElsewhere {
					ownedID = candidate
					if subsumedRegex.MatchString(line) {
						isSub = true
					}
				}
			}

			if ownedID != "" {
				alreadyInThisFRD := false
				for _, item := range frdOwnedReqs[fn] {
					if item == ownedID {
						alreadyInThisFRD = true
						break
					}
				}
				if !alreadyInThisFRD {
					frdOwnedReqs[fn] = append(frdOwnedReqs[fn], ownedID)
				}
				if isSub {
					subsumedReqs[ownedID] = true
				}
			}
		}
	}

	fmt.Printf("  Master FRD Catalog: %d defined requirements across %d FRDs.\n\n", len(allReqDefs), len(existingFrds))

	// Step 2: Validate each target TDD
	for _, tddPath := range targetFiles {
		tddName := filepath.Base(tddPath)
		fmt.Println("----------------------------------------------------------")
		fmt.Printf("  Auditing TDD: %s\n", tddName)
		fmt.Println("----------------------------------------------------------")

		lines, err := readLines(tddPath)
		if err != nil {
			fmt.Printf("  [ERROR] Reading file %s: %v\n", tddName, err)
			hasErrors = true
			continue
		}

		// 2.1 Header Metadata Validation
		var sourceFrdRaw, frdCommitHash, tddVersion, ghMilestone string
		limit := 40
		if len(lines) < limit {
			limit = len(lines)
		}
		for i := 0; i < limit; i++ {
			line := lines[i]
			if m := sourceFrdHeaderRegex.FindStringSubmatch(line); len(m) > 1 {
				sourceFrdRaw = strings.TrimSpace(m[1])
			}
			if m := commitHashHeaderRegex.FindStringSubmatch(line); len(m) > 1 {
				frdCommitHash = strings.TrimSpace(m[1])
			}
			if m := tddVersionHeaderRegex.FindStringSubmatch(line); len(m) > 1 {
				tddVersion = strings.TrimSpace(m[1])
			}
			if m := ghMilestoneHeaderRegex.FindStringSubmatch(line); len(m) > 1 {
				ghMilestone = strings.TrimSpace(m[1])
			}
		}

		if sourceFrdRaw == "" {
			fmt.Println("  [ERROR] Missing required header bullet: '- **Source FRD:**'")
			hasErrors = true
		}
		if frdCommitHash == "" {
			fmt.Println("  [ERROR] Missing required header bullet: '- **FRD Commit Hash:**'")
			hasErrors = true
		}
		if tddVersion == "" {
			fmt.Println("  [ERROR] Missing required header bullet: '- **TDD Version:**'")
			hasErrors = true
		}
		if ghMilestone == "" {
			fmt.Println("  [ERROR] Missing required header bullet: '- **GitHub Milestone:**'")
			hasErrors = true
		}

		var sourceFrdFileName string
		if sourceFrdRaw != "" {
			if m := frdFileExtractRegex.FindStringSubmatch(sourceFrdRaw); len(m) > 1 {
				sourceFrdFileName = m[1]
				sourceFrdFullPath := filepath.Join(frdDir, sourceFrdFileName)
				if _, err := os.Stat(sourceFrdFullPath); os.IsNotExist(err) {
					fmt.Printf("  [ERROR] Source FRD file does not exist in docs/frd/: %s\n", sourceFrdFileName)
					hasErrors = true
					sourceFrdFileName = ""
				} else {
					fmt.Printf("  Source FRD: %s (TDD %s | Audited FRD Commit: %s)\n", sourceFrdFileName, tddVersion, frdCommitHash)
					if frdCommitHash != "" && git.CommitExists(repoRoot, frdCommitHash) {
						newer := git.GetNewerCommits(repoRoot, frdCommitHash, "docs/frd/"+sourceFrdFileName)
						if len(newer) > 0 {
							fmt.Printf("  [WARNING] Source FRD (%s) has %d newer commit(s) since TDD's FRD Commit Hash (%s):\n", sourceFrdFileName, len(newer), frdCommitHash)
							for idx, c := range newer {
								if idx >= 3 {
									break
								}
								fmt.Printf("    -> %s\n", c)
							}
						}
					}
				}
			} else {
				fmt.Printf("  [ERROR] Could not parse valid FRD filename (XX_*_FRD.md) from Source FRD header: %s\n", sourceFrdRaw)
				hasErrors = true
			}
		}

		// 2.2 Zero Dangling FR-XXX-YYY References Check
		type tddRef struct {
			Req  string
			Line int
		}
		var tddRefs []tddRef
		uniqueTddRefMap := make(map[string]bool)
		for i, line := range lines {
			for _, m := range reqRegex.FindAllString(line, -1) {
				tddRefs = append(tddRefs, tddRef{Req: m, Line: i + 1})
				uniqueTddRefMap[m] = true
			}
		}

		var uniqueTddRefs []string
		for r := range uniqueTddRefMap {
			uniqueTddRefs = append(uniqueTddRefs, r)
		}
		sort.Strings(uniqueTddRefs)

		danglingCount := 0
		for _, r := range uniqueTddRefs {
			if !allReqDefs[r] {
				danglingCount++
				hasErrors = true
				var lineNums []string
				for _, tr := range tddRefs {
					if tr.Req == r {
						lineNums = append(lineNums, fmt.Sprintf("%d", tr.Line))
					}
				}
				fmt.Printf("  [DANGLING REFERENCE] %s (lines: %s) is not defined in any FRD.\n", r, strings.Join(lineNums, ", "))
			}
		}
		if danglingCount == 0 {
			fmt.Printf("  [OK] Cross-References: %d unique FR-XXX-YYY IDs referenced, 0 dangling.\n", len(uniqueTddRefs))
		}

		// 2.3 Parse Deliverable Slices (### [SLICE-XX-YY] Title)
		slices := make(map[string]*SliceInfo)
		var sliceOrder []string
		var currentSlice *SliceInfo
		currentSubSection := ""

		topHeadingRegex := regexp.MustCompile(`^##\s+[^#]`)

		for i, line := range lines {
			if m := sliceHeadingRegex.FindStringSubmatch(line); len(m) > 2 {
				sid := strings.TrimSpace(m[1])
				stitle := strings.TrimSpace(m[2])
				currentSubSection = ""
				if _, exists := slices[sid]; exists {
					fmt.Printf("  [ERROR] Duplicate Deliverable Slice heading [%s] at line %d.\n", sid, i+1)
					hasErrors = true
				} else {
					sliceOrder = append(sliceOrder, sid)
					currentSlice = &SliceInfo{
						ID:    sid,
						Title: stitle,
						Line:  i + 1,
					}
					slices[sid] = currentSlice
				}
				continue
			}

			// Exit slice parsing if a new top-level ## section starts after Section 6
			if currentSlice != nil && topHeadingRegex.MatchString(line) {
				currentSlice = nil
				currentSubSection = ""
				continue
			}

			if currentSlice != nil {
				if m := sliceSubHeaderRegex.FindStringSubmatch(line); len(m) > 1 {
					subTitle := strings.ToLower(strings.TrimSpace(m[1]))
					currentSlice.InTargetFiles = false
					if strings.HasPrefix(subTitle, "implementation blueprint") {
						currentSlice.HasBlueprintSection = true
						currentSubSection = "blueprint"
					} else if strings.HasPrefix(subTitle, "acceptance criteria") {
						currentSlice.HasAcceptanceSection = true
						currentSubSection = "acceptance"
					} else if strings.HasPrefix(subTitle, "verification commands") {
						currentSlice.HasVerifySection = true
						currentSubSection = "verify"
					} else {
						currentSubSection = "other"
					}
					continue
				}

				if currentSubSection == "" {
					if m := sliceIdRegex.FindStringSubmatch(line); len(m) > 1 {
						currentSlice.DeclaredID = strings.TrimSpace(m[1])
						currentSlice.InTargetFiles = false
					} else if m := labelsRegex.FindStringSubmatch(line); len(m) > 1 {
						currentSlice.Labels = strings.TrimSpace(m[1])
						currentSlice.InTargetFiles = false
					} else if m := dependsOnRegex.FindStringSubmatch(line); len(m) > 1 {
						currentSlice.DependsOnRaw = strings.TrimSpace(m[1])
						currentSlice.InTargetFiles = false
						if !noneRegex.MatchString(currentSlice.DependsOnRaw) {
							currentSlice.DependsOnList = depSliceExtract.FindAllString(currentSlice.DependsOnRaw, -1)
						}
					} else if m := reqCoveredRegex.FindStringSubmatch(line); len(m) > 1 {
						currentSlice.RequirementsCovered = reqRegex.FindAllString(m[1], -1)
						currentSlice.InTargetFiles = false
					} else if m := tddSectionsRegex.FindStringSubmatch(line); len(m) > 1 {
						currentSlice.TddSections = strings.TrimSpace(m[1])
						currentSlice.InTargetFiles = false
					} else if m := targetFilesRegex.FindStringSubmatch(line); len(m) > 1 {
						currentSlice.InTargetFiles = true
						inline := strings.TrimSpace(m[1])
						if inline != "" {
							currentSlice.TargetFiles = append(currentSlice.TargetFiles, inline)
						}
					} else if currentSlice.InTargetFiles && targetFileSubBullet.MatchString(line) {
						mSub := targetFileSubBullet.FindStringSubmatch(line)
						currentSlice.TargetFiles = append(currentSlice.TargetFiles, strings.TrimSpace(mSub[1]))
					} else if strings.TrimSpace(line) != "" {
						currentSlice.InTargetFiles = false
					}
				} else {
					trimmedLine := strings.TrimSpace(line)
					if currentSubSection == "blueprint" && trimmedLine != "" {
						currentSlice.BlueprintLines++
					} else if currentSubSection == "acceptance" && checkboxRegex.MatchString(line) {
						currentSlice.AcceptanceCheckboxes++
					} else if currentSubSection == "verify" && trimmedLine != "" && !strings.HasPrefix(trimmedLine, "```") {
						currentSlice.VerifyLines++
					}
				}
			}
		}

		if len(sliceOrder) == 0 {
			fmt.Println("  [ERROR] No Deliverable Slice blocks ('### [SLICE-XX-YY] Title') found in Section 6.")
			hasErrors = true
		} else {
			sliceErrors := 0
			for _, sid := range sliceOrder {
				s := slices[sid]
				if s.DeclaredID != sid {
					fmt.Printf("  [ERROR] Slice [%s] (line %d): '- **Slice ID:**' is '%s', expected '%s'.\n", sid, s.Line, s.DeclaredID, sid)
					sliceErrors++
					hasErrors = true
				}
				if s.Labels == "" {
					fmt.Printf("  [ERROR] Slice [%s] (line %d): Missing '- **Labels:**'.\n", sid, s.Line)
					sliceErrors++
					hasErrors = true
				}
				if s.DependsOnRaw == "" {
					fmt.Printf("  [ERROR] Slice [%s] (line %d): Missing '- **Depends On:**' (use 'None' or upstream SLICE-XX-YY IDs).\n", sid, s.Line)
					sliceErrors++
					hasErrors = true
				} else if !noneRegex.MatchString(s.DependsOnRaw) && len(s.DependsOnList) == 0 {
					fmt.Printf("  [ERROR] Slice [%s] (line %d): '- **Depends On:**' must be 'None' or valid SLICE-XX-YY ID(s), got '%s'.\n", sid, s.Line, s.DependsOnRaw)
					sliceErrors++
					hasErrors = true
				}
				if len(s.RequirementsCovered) == 0 {
					fmt.Printf("  [ERROR] Slice [%s] (line %d): '- **Requirements Covered:**' must cite at least one FR-XXX-YYY.\n", sid, s.Line)
					sliceErrors++
					hasErrors = true
				}
				if s.TddSections == "" {
					fmt.Printf("  [ERROR] Slice [%s] (line %d): Missing '- **TDD Sections:**'.\n", sid, s.Line)
					sliceErrors++
					hasErrors = true
				}
				if len(s.TargetFiles) == 0 {
					fmt.Printf("  [ERROR] Slice [%s] (line %d): '- **Target Files:**' must list at least one target file bullet.\n", sid, s.Line)
					sliceErrors++
					hasErrors = true
				}
				if !s.HasBlueprintSection || s.BlueprintLines < 2 {
					fmt.Printf("  [ERROR] Slice [%s] (line %d): Missing or empty '#### Implementation Blueprint & Signatures' section.\n", sid, s.Line)
					sliceErrors++
					hasErrors = true
				}
				if !s.HasAcceptanceSection || s.AcceptanceCheckboxes == 0 {
					fmt.Printf("  [ERROR] Slice [%s] (line %d): '#### Acceptance Criteria' must include at least one '- [ ]' checkbox.\n", sid, s.Line)
					sliceErrors++
					hasErrors = true
				}
				if !s.HasVerifySection || s.VerifyLines == 0 {
					fmt.Printf("  [ERROR] Slice [%s] (line %d): Missing or empty '#### Verification Commands' section.\n", sid, s.Line)
					sliceErrors++
					hasErrors = true
				}

				for _, dep := range s.DependsOnList {
					if dep == sid {
						fmt.Printf("  [ERROR] Slice [%s] depends on itself.\n", sid)
						sliceErrors++
						hasErrors = true
					} else if _, exists := slices[dep]; !exists {
						fmt.Printf("  [ERROR] Slice [%s] depends on unknown slice '%s'.\n", sid, dep)
						sliceErrors++
						hasErrors = true
					}
				}
			}

			// Check DAG for cycles via DFS (0=unvisited, 1=visiting, 2=visited)
			visitState := make(map[string]int)
			hasCycle := false

			var testSliceCycle func(string)
			testSliceCycle = func(nodeID string) {
				visitState[nodeID] = 1
				if s, ok := slices[nodeID]; ok {
					for _, dep := range s.DependsOnList {
						if _, exists := slices[dep]; !exists {
							continue
						}
						if visitState[dep] == 1 {
							fmt.Printf("  [ERROR] Dependency cycle detected involving [%s] -> [%s].\n", nodeID, dep)
							hasCycle = true
							hasErrors = true
						} else if visitState[dep] == 0 {
							testSliceCycle(dep)
						}
					}
				}
				visitState[nodeID] = 2
			}

			for _, sid := range sliceOrder {
				if visitState[sid] == 0 {
					testSliceCycle(sid)
				}
			}

			if sliceErrors == 0 && !hasCycle {
				fmt.Printf("  [OK] Deliverable Slice Manifest: %d valid slices with acyclic dependency DAG.\n", len(sliceOrder))
			}
		}

		// 2.4 Validate 100% Source FRD Requirement Coverage
		if sourceFrdFileName != "" {
			if ownedList, ok := frdOwnedReqs[sourceFrdFileName]; ok {
				var activeReqs []string
				var subsumedList []string
				for _, r := range ownedList {
					if subsumedReqs[r] {
						subsumedList = append(subsumedList, r)
					} else {
						activeReqs = append(activeReqs, r)
					}
				}

				sliceCoveredSet := make(map[string]bool)
				for _, sid := range sliceOrder {
					for _, r := range slices[sid].RequirementsCovered {
						sliceCoveredSet[r] = true
					}
				}

				matrixCoveredSet := make(map[string]bool)
				for _, line := range lines {
					if strings.HasPrefix(strings.TrimSpace(line), "|") {
						for _, m := range reqRegex.FindAllString(line, -1) {
							matrixCoveredSet[m] = true
						}
					}
				}

				var missingInMatrix []string
				var missingInSlices []string
				var missingSubsumed []string

				for _, r := range activeReqs {
					if !matrixCoveredSet[r] {
						missingInMatrix = append(missingInMatrix, r)
					}
					if !sliceCoveredSet[r] {
						missingInSlices = append(missingInSlices, r)
					}
				}

				for _, r := range subsumedList {
					if !uniqueTddRefMap[r] {
						missingSubsumed = append(missingSubsumed, r)
					}
				}

				if len(missingInMatrix) > 0 {
					fmt.Printf("  [ERROR] Active FRD requirements missing from TDD Traceability Matrix (%d):\n", len(missingInMatrix))
					fmt.Printf("    -> %s\n", strings.Join(missingInMatrix, ", "))
					hasErrors = true
				}
				if len(missingInSlices) > 0 {
					fmt.Printf("  [ERROR] Active FRD requirements not assigned to any Deliverable Slice (%d):\n", len(missingInSlices))
					fmt.Printf("    -> %s\n", strings.Join(missingInSlices, ", "))
					hasErrors = true
				}
				if len(missingSubsumed) > 0 {
					fmt.Printf("  [ERROR] Subsumed FRD requirements not accounted for in TDD Traceability Matrix (%d):\n", len(missingSubsumed))
					fmt.Printf("    -> %s\n", strings.Join(missingSubsumed, ", "))
					hasErrors = true
				}

				if len(missingInMatrix) == 0 && len(missingInSlices) == 0 && len(missingSubsumed) == 0 {
					fmt.Printf("  [OK] 100%% Source FRD Coverage: %d/%d Active requirements sliced & traced; %d/%d Subsumed accounted for.\n",
						len(activeReqs), len(activeReqs), len(subsumedList), len(subsumedList))
				}
			}
		}
		fmt.Println()
	}

	if !hasErrors {
		fmt.Println("  [PASS] All TDD coverage, cross-reference, and slice DAG checks passed!")
		fmt.Println("==========================================================")
		return true, nil
	}

	fmt.Println("  [FAIL] TDD linting failed. Resolve the errors above before committing or publishing issues.")
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
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, 1024*1024)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}
