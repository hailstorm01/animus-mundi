package tddslice

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"animus-mundi/pkg/toolchain/git"
	"animus-mundi/pkg/toolchain/linter/tdd"
)

var (
	sourceFrdRegex   = regexp.MustCompile(`^\s*-\s+\*\*Source FRD:\*\*\s*(.+)$`)
	commitHashRegex  = regexp.MustCompile(`^\s*-\s+\*\*FRD Commit Hash:\*\*\s*` + "`" + `?([^` + "`" + `\s]+)` + "`" + `?`)
	tddVersionRegex  = regexp.MustCompile(`^\s*-\s+\*\*TDD Version:\*\*\s*` + "`" + `?([^` + "`" + `\r\n]+)` + "`" + `?`)
	ghMilestoneRegex = regexp.MustCompile(`^\s*-\s+\*\*GitHub Milestone:\*\*\s*` + "`" + `?([^` + "`" + `\r\n]+)` + "`" + `?`)
	frdFileRegex     = regexp.MustCompile(`(\d{2}_[A-Za-z0-9_]+_FRD\.md)`)

	sliceHeadingRegex   = regexp.MustCompile(`^###\s+\[(SLICE-\d{2}-\d{2})\]\s+(.+)$`)
	sliceSubHeaderRegex = regexp.MustCompile(`^####\s+(.+)$`)
	labelsRegex         = regexp.MustCompile(`^\s*-\s+\*\*Labels:\*\*\s*(.+)$`)
	dependsOnRegex      = regexp.MustCompile(`^\s*-\s+\*\*Depends On:\*\*\s*(.+)$`)
	reqCoveredRegex     = regexp.MustCompile(`^\s*-\s+\*\*Requirements Covered:\*\*\s*(.+)$`)
	tddSectionsRegex    = regexp.MustCompile(`^\s*-\s+\*\*TDD Sections:\*\*\s*(.+)$`)
	targetFilesRegex    = regexp.MustCompile(`^\s*-\s+\*\*Target Files:\*\*\s*(.*)$`)
	targetFileSubBullet = regexp.MustCompile(`^\s+-\s+(.+)$`)
	depSliceExtract     = regexp.MustCompile(`SLICE-\d{2}-\d{2}`)
	reqExtract          = regexp.MustCompile(`FR-[A-Z]{3}-\d{3}[a-z]?`)
	noneRegex           = regexp.MustCompile(`(?i)^` + "`" + `?None` + "`" + `?$`)
	topHeadingRegex     = regexp.MustCompile(`^##\s+[^#]`)
)

type SliceDef struct {
	ID                  string
	Title               string
	IssueTitle          string
	Labels              []string
	DependsOnList       []string
	RequirementsCovered []string
	TddSections         string
	TargetFiles         []string
	InTargetFiles       bool
	BlueprintLines      []string
	AcceptanceLines     []string
	VerifyLines         []string
}

type ExistingIssue struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	URL    string `json:"url"`
	State  string `json:"state"`
}

type SliceOptions struct {
	TddFile        string
	Repo           string
	DryRun         bool
	Publish        bool
	UpdateExisting bool
	ShowBodies     bool
}

// Slice processes a TDD into a GitHub Milestone and dependency-ordered deliverable issues.
func Slice(repoRoot string, opts SliceOptions) error {
	tddFullPath := opts.TddFile
	if !filepath.IsAbs(tddFullPath) {
		tddFullPath = filepath.Join(repoRoot, tddFullPath)
	}

	if _, err := os.Stat(tddFullPath); os.IsNotExist(err) {
		return fmt.Errorf("TDD file not found: %s", opts.TddFile)
	}

	tddRelPath, _ := filepath.Rel(repoRoot, tddFullPath)
	tddRelPath = filepath.ToSlash(tddRelPath)

	isLivePublish := opts.Publish
	if !isLivePublish {
		opts.DryRun = true
	}

	// Step 1: Mandatory Quality Gate via tdd.Lint
	fmt.Println(">>> Running mandatory pre-slice TDD linter gate...")
	ok, err := tdd.Lint(repoRoot, []string{tddFullPath})
	if err != nil || !ok {
		return fmt.Errorf("[ABORT] TDD failed linting. Fix lint errors before slicing or publishing GitHub issues")
	}

	// Step 2: Resolve GitHub repository
	repo := opts.Repo
	if repo == "" {
		if out, err := git.Output(repoRoot, "remote", "get-url", "origin"); err == nil {
			rem := strings.TrimSpace(string(out))
			re := regexp.MustCompile(`github\.com[:/]([^/]+/[^/.]+?)(?:\.git)?$`)
			if m := re.FindStringSubmatch(rem); len(m) > 1 {
				repo = m[1]
			}
		}
	}
	if repo == "" {
		repo = "hailstorm01/animus-mundi"
	}

	// Step 3: Parse TDD Metadata & Slices
	content, err := os.ReadFile(tddFullPath)
	if err != nil {
		return fmt.Errorf("reading TDD file: %w", err)
	}
	lines := strings.Split(string(content), "\n")

	sourceFrdRel := ""
	frdCommitHash := ""
	tddVersion := ""
	ghMilestone := ""

	limit := 40
	if len(lines) < limit {
		limit = len(lines)
	}
	for i := 0; i < limit; i++ {
		l := lines[i]
		if m := sourceFrdRegex.FindStringSubmatch(l); len(m) > 1 {
			rawFrd := strings.TrimSpace(m[1])
			if mf := frdFileRegex.FindStringSubmatch(rawFrd); len(mf) > 1 {
				sourceFrdRel = "docs/frd/" + mf[1]
			} else {
				sourceFrdRel = rawFrd
			}
		}
		if m := commitHashRegex.FindStringSubmatch(l); len(m) > 1 {
			frdCommitHash = strings.TrimSpace(m[1])
		}
		if m := tddVersionRegex.FindStringSubmatch(l); len(m) > 1 {
			tddVersion = strings.TrimSpace(m[1])
		}
		if m := ghMilestoneRegex.FindStringSubmatch(l); len(m) > 1 {
			ghMilestone = strings.TrimSpace(m[1])
		}
	}

	slices := make(map[string]*SliceDef)
	var sliceOrder []string
	var currentSlice *SliceDef
	currentSubSection := ""

	for _, rawLine := range lines {
		line := strings.TrimRight(rawLine, "\r")

		if m := sliceHeadingRegex.FindStringSubmatch(line); len(m) > 2 {
			sid := strings.TrimSpace(m[1])
			stitle := strings.TrimSpace(m[2])
			currentSubSection = ""
			sliceOrder = append(sliceOrder, sid)
			currentSlice = &SliceDef{
				ID:         sid,
				Title:      stitle,
				IssueTitle: fmt.Sprintf("[%s] %s", sid, stitle),
			}
			slices[sid] = currentSlice
			continue
		}

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
					currentSubSection = "blueprint"
				} else if strings.HasPrefix(subTitle, "acceptance criteria") {
					currentSubSection = "acceptance"
				} else if strings.HasPrefix(subTitle, "verification commands") {
					currentSubSection = "verify"
				} else {
					currentSubSection = "other"
				}
				continue
			}

			if currentSubSection == "" {
				if m := labelsRegex.FindStringSubmatch(line); len(m) > 1 {
					parts := strings.Split(m[1], ",")
					for _, p := range parts {
						clean := strings.Trim(strings.TrimSpace(p), "`")
						if clean != "" {
							currentSlice.Labels = append(currentSlice.Labels, clean)
						}
					}
					currentSlice.InTargetFiles = false
				} else if m := dependsOnRegex.FindStringSubmatch(line); len(m) > 1 {
					depRaw := strings.TrimSpace(m[1])
					if !noneRegex.MatchString(depRaw) {
						currentSlice.DependsOnList = depSliceExtract.FindAllString(depRaw, -1)
					}
					currentSlice.InTargetFiles = false
				} else if m := reqCoveredRegex.FindStringSubmatch(line); len(m) > 1 {
					currentSlice.RequirementsCovered = reqExtract.FindAllString(m[1], -1)
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
				if currentSubSection == "blueprint" {
					currentSlice.BlueprintLines = append(currentSlice.BlueprintLines, line)
				} else if currentSubSection == "acceptance" {
					currentSlice.AcceptanceLines = append(currentSlice.AcceptanceLines, line)
				} else if currentSubSection == "verify" {
					currentSlice.VerifyLines = append(currentSlice.VerifyLines, line)
				}
			}
		}
	}

	// Step 4: Topological Sort
	var topoOrder []string
	visited := make(map[string]bool)

	var addTopoSlice func(string)
	addTopoSlice = func(nodeID string) {
		if visited[nodeID] {
			return
		}
		visited[nodeID] = true
		if s, ok := slices[nodeID]; ok {
			for _, dep := range s.DependsOnList {
				if _, exists := slices[dep]; exists && !visited[dep] {
					addTopoSlice(dep)
				}
			}
		}
		topoOrder = append(topoOrder, nodeID)
	}

	for _, sid := range sliceOrder {
		addTopoSlice(sid)
	}

	var allLabels []string
	labelSeen := make(map[string]bool)
	for _, sid := range topoOrder {
		for _, lbl := range slices[sid].Labels {
			if !labelSeen[lbl] {
				labelSeen[lbl] = true
				allLabels = append(allLabels, lbl)
			}
		}
	}

	// Step 5: Check existing GitHub Issues
	existingIssuesBySlice := make(map[string]ExistingIssue)
	issueNumberMap := make(map[string]int)
	ghAvailable := false

	cmdAuth := exec.Command("gh", "auth", "status")
	if cmdAuth.Run() == nil {
		ghAvailable = true
		cmdList := exec.Command("gh", "issue", "list", "--repo", repo, "--state", "all", "--limit", "500", "--json", "number,title,url,state")
		if out, err := cmdList.Output(); err == nil {
			var parsed []ExistingIssue
			if json.Unmarshal(out, &parsed) == nil {
				sliceRe := regexp.MustCompile(`^\[(SLICE-\d{2}-\d{2})\]`)
				for _, iss := range parsed {
					if m := sliceRe.FindStringSubmatch(iss.Title); len(m) > 1 {
						sid := m[1]
						if _, ok := existingIssuesBySlice[sid]; !ok {
							existingIssuesBySlice[sid] = iss
							issueNumberMap[sid] = iss.Number
						}
					}
				}
			}
		}
	}

	fmt.Println()
	fmt.Println("==========================================================")
	if isLivePublish {
		fmt.Println("  TDD ISSUE SLICER: LIVE GITHUB PUBLICATION (-Publish)    ")
	} else {
		fmt.Println("  TDD ISSUE SLICER: DRY-RUN PREVIEW (-DryRun)             ")
	}
	fmt.Println("==========================================================")
	fmt.Printf("Repository      : %s\n", repo)
	fmt.Printf("Source TDD      : %s (%s)\n", tddRelPath, tddVersion)
	fmt.Printf("Source FRD      : %s (Commit: %s)\n", sourceFrdRel, frdCommitHash)
	fmt.Printf("GitHub Milestone: %s\n", ghMilestone)
	fmt.Printf("Labels to Ensure: %s\n", strings.Join(allLabels, ", "))
	fmt.Printf("Total Slices    : %d (Topologically Ordered)\n\n", len(topoOrder))

	formatSliceBody := func(s *SliceDef) string {
		depText := "None *(Foundational Slice -- Ready to Start Immediately)*"
		if len(s.DependsOnList) > 0 {
			var depParts []string
			for _, d := range s.DependsOnList {
				if num, ok := issueNumberMap[d]; ok {
					depParts = append(depParts, fmt.Sprintf("#%d (`%s`)", num, d))
				} else {
					depParts = append(depParts, fmt.Sprintf("`%s`", d))
				}
			}
			depText = strings.Join(depParts, ", ")
		}

		var reqParts []string
		for _, r := range s.RequirementsCovered {
			reqParts = append(reqParts, fmt.Sprintf("`%s`", r))
		}
		reqList := strings.Join(reqParts, ", ")

		var targetFileLines []string
		for _, tf := range s.TargetFiles {
			targetFileLines = append(targetFileLines, "- "+tf)
		}
		targetFilesBlock := strings.Join(targetFileLines, "\n")
		blueprintBlock := strings.TrimSpace(strings.Join(s.BlueprintLines, "\n"))
		acceptanceBlock := strings.TrimSpace(strings.Join(s.AcceptanceLines, "\n"))
		verifyBlock := strings.TrimSpace(strings.Join(s.VerifyLines, "\n"))

		return fmt.Sprintf(`## Deliverable Context & Traceability

- **Slice ID:** `+"`%s`"+`
- **Source TDD:** [%s](%s) *(Version: `+"`%s`"+`)*
- **Source FRD:** [%s](%s) *(Audited FRD Commit: `+"`%s`"+`)*
- **TDD Reference Sections:** %s
- **FRD Requirements Covered:** %s
- **Depends On:** %s

---

## Target Files

%s

---

## Implementation Blueprint & Signatures

%s

---

## Acceptance Criteria

%s

---

## Verification Commands

%s

---

> [!IMPORTANT]
> **Zero-Guesswork Execution Protocol (Engineer / Agent):**
> 1. Read the referenced **TDD Sections (%s)** in `+"`%s`"+` before writing code.
> 2. Implement strictly the types, signatures, and target files scoped to `+"`%s`"+`. Do not invent unapproved schemas or alter FRD rules.
> 3. Run the **Verification Commands** above and `+"`cardtown lint`"+` until all checks exit `+"`0`"+`.
`,
			s.ID, tddRelPath, tddRelPath, tddVersion,
			sourceFrdRel, sourceFrdRel, frdCommitHash,
			s.TddSections, reqList, depText,
			targetFilesBlock, blueprintBlock, acceptanceBlock, verifyBlock,
			s.TddSections, tddRelPath, s.ID)
	}

	if !isLivePublish {
		fmt.Println("== TOPOLOGICAL DELIVERABLE SLICE PLAN ==")
		for idx, sid := range topoOrder {
			s := slices[sid]
			deps := "None (Ready)"
			if len(s.DependsOnList) > 0 {
				deps = strings.Join(s.DependsOnList, ", ")
			}
			reqs := strings.Join(s.RequirementsCovered, ", ")
			actionTag := "[CREATE NEW]"
			if iss, ok := existingIssuesBySlice[sid]; ok {
				if opts.UpdateExisting {
					actionTag = fmt.Sprintf("[UPDATE #%d]", iss.Number)
				} else {
					actionTag = fmt.Sprintf("[SKIP EXISTING #%d]", iss.Number)
				}
			}
			fmt.Printf("  %2d. %s %s\n", idx+1, actionTag, s.IssueTitle)
			fmt.Printf("      Depends On  : %s\n", deps)
			fmt.Printf("      Requirements: %s\n", reqs)
			fmt.Printf("      Target Files: %d file(s) (%s)\n", len(s.TargetFiles), strings.Join(s.TargetFiles, "; "))
		}

		if opts.ShowBodies || len(topoOrder) > 0 {
			previewList := []string{topoOrder[0]}
			if opts.ShowBodies {
				previewList = topoOrder
			}
			for _, psid := range previewList {
				fmt.Println("\n----------------------------------------------------------")
				fmt.Printf("  SAMPLE ISSUE BODY PREVIEW: %s\n", slices[psid].IssueTitle)
				fmt.Println("----------------------------------------------------------")
				fmt.Println(formatSliceBody(slices[psid]))
			}
		}

		fmt.Printf("\n[DRY-RUN COMPLETE] No changes were made to GitHub (%s).\n", repo)
		fmt.Println("To publish live after human confirmation, run with -publish.")
		return nil
	}

	// Step 6: Live Publish Mode (-Publish)
	if !ghAvailable {
		return fmt.Errorf("GitHub CLI (gh) is not authenticated or cannot reach GitHub. Run 'gh auth login' first")
	}

	fmt.Printf(">>> Ensuring GitHub Milestone '%s' exists in %s...\n", ghMilestone, repo)
	cmdMilestones := exec.Command("gh", "api", fmt.Sprintf("repos/%s/milestones?state=all&per_page=100", repo))
	milestoneExists := false
	if out, err := cmdMilestones.Output(); err == nil {
		type milestoneObj struct {
			Number int    `json:"number"`
			Title  string `json:"title"`
		}
		var msList []milestoneObj
		_ = json.Unmarshal(out, &msList)
		for _, m := range msList {
			if m.Title == ghMilestone {
				milestoneExists = true
				fmt.Printf("  [OK] Milestone already exists (#%d: %s)\n", m.Number, ghMilestone)
				break
			}
		}
	}

	if !milestoneExists {
		desc := fmt.Sprintf("Engineering deliverables sliced from %s (%s) governing %s.", tddRelPath, tddVersion, sourceFrdRel)
		cmdCreateM := exec.Command("gh", "api", fmt.Sprintf("repos/%s/milestones", repo), "-X", "POST", "-f", fmt.Sprintf("title=%s", ghMilestone), "-f", fmt.Sprintf("description=%s", desc))
		if out, err := cmdCreateM.CombinedOutput(); err != nil {
			return fmt.Errorf("creating milestone failed: %s (%w)", string(out), err)
		}
		fmt.Printf("  [CREATED] Milestone '%s'\n", ghMilestone)
	}

	// 6.2 Ensure Labels
	fmt.Printf(">>> Ensuring Labels exist in %s...\n", repo)
	for _, lbl := range allLabels {
		color := "EDEDED"
		if strings.HasPrefix(lbl, "type:") {
			color = "0E8A16"
		} else if strings.HasPrefix(lbl, "phase:") {
			color = "1D76DB"
		} else if strings.HasPrefix(lbl, "component:") {
			color = "5319E7"
		} else if strings.HasPrefix(lbl, "tdd:") {
			color = "FBCA04"
		}
		cmdLabel := exec.Command("gh", "label", "create", lbl, "--repo", repo, "--color", color, "--force")
		_ = cmdLabel.Run()
	}
	fmt.Printf("  [OK] Verified %d label(s): %s\n", len(allLabels), strings.Join(allLabels, ", "))

	// 6.3 Publish / Update Slices
	fmt.Println(">>> Publishing Deliverable Slice Issues in topological order...")
	type resultRow struct {
		Slice  string
		Issue  string
		Action string
		URL    string
	}
	var results []resultRow

	for _, sid := range topoOrder {
		s := slices[sid]
		bodyText := formatSliceBody(s)
		tmpFile, err := os.CreateTemp("", fmt.Sprintf("tdd_issue_%s_*.md", sid))
		if err != nil {
			return fmt.Errorf("creating temp file: %w", err)
		}
		tmpPath := tmpFile.Name()
		_, _ = tmpFile.WriteString(bodyText)
		_ = tmpFile.Close()

		if iss, ok := existingIssuesBySlice[sid]; ok {
			if opts.UpdateExisting {
				args := []string{"issue", "edit", fmt.Sprintf("%d", iss.Number), "--repo", repo, "--title", s.IssueTitle, "--milestone", ghMilestone, "--body-file", tmpPath}
				for _, lbl := range s.Labels {
					args = append(args, "--add-label", lbl)
				}
				cmdEdit := exec.Command("gh", args...)
				if out, err := cmdEdit.CombinedOutput(); err != nil {
					_ = os.Remove(tmpPath)
					return fmt.Errorf("editing issue #%d failed: %s (%w)", iss.Number, string(out), err)
				}
				fmt.Printf("  [UPDATED] #%d : %s -> %s\n", iss.Number, s.IssueTitle, iss.URL)
				results = append(results, resultRow{Slice: sid, Issue: fmt.Sprintf("#%d", iss.Number), Action: "UPDATED", URL: iss.URL})
			} else {
				fmt.Printf("  [SKIPPED] #%d : %s (already exists -> %s)\n", iss.Number, s.IssueTitle, iss.URL)
				results = append(results, resultRow{Slice: sid, Issue: fmt.Sprintf("#%d", iss.Number), Action: "SKIPPED", URL: iss.URL})
			}
		} else {
			args := []string{"issue", "create", "--repo", repo, "--title", s.IssueTitle, "--milestone", ghMilestone, "--body-file", tmpPath}
			for _, lbl := range s.Labels {
				args = append(args, "--label", lbl)
			}
			cmdCreate := exec.Command("gh", args...)
			out, err := cmdCreate.CombinedOutput()
			if err != nil {
				_ = os.Remove(tmpPath)
				return fmt.Errorf("creating issue for %s failed: %s (%w)", sid, string(out), err)
			}
			urlRe := regexp.MustCompile(`https://github\.com/[^/]+/[^/]+/issues/(\d+)`)
			createdURL := strings.TrimSpace(string(out))
			newNum := "?"
			if m := urlRe.FindStringSubmatch(createdURL); len(m) > 1 {
				newNum = m[1]
				var n int
				if _, err := fmt.Sscanf(newNum, "%d", &n); err == nil {
					issueNumberMap[sid] = n
				}
			}
			fmt.Printf("  [CREATED] #%s : %s -> %s\n", newNum, s.IssueTitle, createdURL)
			results = append(results, resultRow{Slice: sid, Issue: "#" + newNum, Action: "CREATED", URL: createdURL})
		}
		_ = os.Remove(tmpPath)
	}

	fmt.Println()
	fmt.Println("==========================================================")
	fmt.Printf("  GITHUB PUBLICATION SUMMARY (%d Slices)\n", len(results))
	fmt.Println("==========================================================")
	for _, r := range results {
		fmt.Printf("  %-12s | %-6s | %-8s | %s\n", r.Slice, r.Issue, r.Action, r.URL)
	}
	fmt.Println("==========================================================")
	return nil
}
