package audit

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"animus-mundi/pkg/toolchain/git"
	"animus-mundi/pkg/toolchain/linter/frd"
)

var (
	socraticHeaderRegex = regexp.MustCompile(`^##\s+\d+\.\s+Socratic`)
	topHeaderRegex      = regexp.MustCompile(`^##\s+`)
	subHeaderRegex      = regexp.MustCompile(`^###\s+`)
	daFindingRegex      = regexp.MustCompile(`^\s*-\s+\*\*\[DA-[A-Z]+-\d+\]`)
	socraticItemRegex   = regexp.MustCompile(`^\s*\d+\.\s`)
)

type AuditCounts struct {
	Blocker  int
	High     int
	Medium   int
	Low      int
	Socratic int
}

// Status inspects docs/audit/pending and outputs staleness and severity metrics.
func Status(repoRoot string) error {
	pendingDir := filepath.Join(repoRoot, "docs", "audit", "pending")
	registryPath := filepath.Join(repoRoot, "docs", "audit", "audit_registry.md")

	var registryRows []string
	if content, err := os.ReadFile(registryPath); err == nil {
		scanner := bufio.NewScanner(strings.NewReader(string(content)))
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(strings.TrimSpace(line), "|") {
				registryRows = append(registryRows, line)
			}
		}
	}

	var files []string
	if entries, err := os.ReadDir(pendingDir); err == nil {
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
				files = append(files, e.Name())
			}
		}
	}
	sort.Strings(files)

	if len(files) == 0 {
		fmt.Println("No pending audits.")
		return nil
	}

	for _, fName := range files {
		var matchedRow string
		for _, row := range registryRows {
			if strings.Contains(row, fName) {
				matchedRow = row
				break
			}
		}

		stale := "NO REGISTRY ROW"
		detail := ""
		target := "?"
		hash := "?"

		if matchedRow != "" {
			parts := strings.Split(matchedRow, "|")
			var cols []string
			for _, p := range parts {
				cols = append(cols, strings.TrimSpace(p))
			}
			if len(cols) > 4 {
				hash = strings.Trim(cols[2], "`")
				target = strings.Trim(cols[4], "`")

				if !git.CommitExists(repoRoot, hash) {
					stale = fmt.Sprintf("UNKNOWN (commit %s not found)", hash)
				} else {
					newer := git.GetNewerCommits(repoRoot, hash, target)
					if len(newer) > 0 {
						stale = "STALE"
						limit := 5
						if len(newer) < limit {
							limit = len(newer)
						}
						detail = strings.Join(newer[:limit], "; ")
					} else {
						stale = "CURRENT"
					}
				}
			}
		}

		filePath := filepath.Join(pendingDir, fName)
		lines, err := readLines(filePath)
		if err != nil {
			fmt.Printf("  [ERROR] Reading %s: %v\n", fName, err)
			continue
		}

		counts := AuditCounts{}
		section := ""
		inSocratic := false

		for _, line := range lines {
			if socraticHeaderRegex.MatchString(line) {
				inSocratic = true
				section = ""
				continue
			}
			if topHeaderRegex.MatchString(line) && !strings.Contains(line, "Socratic") {
				inSocratic = false
			}
			if subHeaderRegex.MatchString(line) {
				section = ""
				if strings.Contains(line, "Blocker") || strings.Contains(line, "Critical") {
					section = "Blocker"
				} else if strings.Contains(line, "High") {
					section = "High"
				} else if strings.Contains(line, "Medium") {
					section = "Medium"
				} else if strings.Contains(line, "Low") {
					section = "Low"
				}
				continue
			}

			if section != "" && daFindingRegex.MatchString(line) {
				switch section {
				case "Blocker":
					counts.Blocker++
				case "High":
					counts.High++
				case "Medium":
					counts.Medium++
				case "Low":
					counts.Low++
				}
			}
			if inSocratic && socraticItemRegex.MatchString(line) {
				counts.Socratic++
			}
		}

		human := (counts.Blocker+counts.High) > 0 || counts.Socratic > 0

		fmt.Println()
		fmt.Println(fName)
		fmt.Printf("  Target: %s | Audited at: %s | Freshness: %s\n", target, hash, stale)
		if detail != "" {
			fmt.Printf("  FRD commits since audit: %s\n", detail)
		}
		fmt.Printf("  Findings: Blocker=%d High=%d Medium=%d Low=%d | Socratic questions=%d\n",
			counts.Blocker, counts.High, counts.Medium, counts.Low, counts.Socratic)
		if human {
			fmt.Println("  Human required: YES (Blocker/High findings or open Socratic questions)")
		} else {
			fmt.Println("  Human required: NO (loop-eligible)")
		}
	}
	return nil
}

// ResolveOptions defines flags for audit resolution.
type ResolveOptions struct {
	AuditFile     string
	Source        string
	CommitMessage string
	Push          bool
	DryRun        bool
}

// Resolve resolves a pending audit and updates audit_registry.md.
func Resolve(repoRoot string, opts ResolveOptions) error {
	auditFileName := filepath.Base(opts.AuditFile)
	pendingRel := filepath.Join("docs", "audit", "pending", auditFileName)
	pendingFilePath := filepath.Join(repoRoot, pendingRel)
	registryRel := filepath.Join("docs", "audit", "audit_registry.md")
	registryPath := filepath.Join(repoRoot, registryRel)

	if opts.Source == "" {
		opts.Source = "Agentic (pm-audit-reviewer)"
	}

	fmt.Println("==========================================================")
	fmt.Println("  CARD TOWN: PM AUDIT RESOLUTION ENGINE                   ")
	fmt.Println("==========================================================")
	fmt.Printf("Target Audit: %s\n", auditFileName)
	fmt.Printf("Source:       %s\n\n", opts.Source)

	if _, err := os.Stat(registryPath); os.IsNotExist(err) {
		return fmt.Errorf("audit registry not found at %s", registryPath)
	}

	lines, err := readLines(registryPath)
	if err != nil {
		return fmt.Errorf("reading registry: %w", err)
	}

	targetLineIndex := -1
	for i, l := range lines {
		if strings.HasPrefix(strings.TrimSpace(l), "|") && strings.Contains(l, auditFileName) {
			targetLineIndex = i
			break
		}
	}
	if targetLineIndex == -1 {
		return fmt.Errorf("no registry row in %s matches %s", registryRel, auditFileName)
	}

	cols := strings.Split(lines[targetLineIndex], "|")
	for i := range cols {
		cols[i] = strings.TrimSpace(cols[i])
	}
	if len(cols) < 9 {
		return fmt.Errorf("failed to parse registry row: %s", lines[targetLineIndex])
	}
	if cols[5] == "RESOLVED" {
		return fmt.Errorf("registry row for %s is already RESOLVED", auditFileName)
	}

	targetFRD := strings.Trim(cols[4], "`")
	targetFRDFull := filepath.Join(repoRoot, filepath.FromSlash(targetFRD))
	if _, err := os.Stat(targetFRDFull); os.IsNotExist(err) {
		return fmt.Errorf("target FRD does not exist: %s", targetFRDFull)
	}

	// Lint the target FRD
	fmt.Printf("Linting target FRD (%s)...\n", targetFRD)
	ok, err := frd.Lint(repoRoot, []string{targetFRD})
	if err != nil || !ok {
		return fmt.Errorf("target FRD failed linting, aborting resolution")
	}

	if opts.DryRun {
		fmt.Println("[DryRun] Validation passed. Would archive pending file and mark resolved in registry.")
		return nil
	}

	// 1. Git rm pending file
	if _, err := os.Stat(pendingFilePath); err == nil {
		if out, err := git.Combined(repoRoot, "rm", filepath.ToSlash(pendingRel)); err != nil {
			return fmt.Errorf("git rm %s failed: %s (%w)", pendingRel, string(out), err)
		}
	}

	// 2. Git add target FRD
	if out, err := git.Combined(repoRoot, "add", filepath.ToSlash(targetFRD)); err != nil {
		return fmt.Errorf("git add %s failed: %s (%w)", targetFRD, string(out), err)
	}

	// 3. Commit remediation
	commitMsg := opts.CommitMessage
	if commitMsg == "" {
		commitMsg = fmt.Sprintf("docs(audit): resolve %s findings in %s", auditFileName, filepath.Base(targetFRD))
	}
	if out, err := git.Combined(repoRoot, "commit", "-m", commitMsg); err != nil {
		return fmt.Errorf("git commit failed: %s (%w)", string(out), err)
	}

	// Get remediation commit hash
	revOut, err := git.Output(repoRoot, "rev-parse", "--short", "HEAD")
	if err != nil {
		return fmt.Errorf("getting HEAD commit: %w", err)
	}
	resCommit := strings.TrimSpace(string(revOut))
	resDate := time.Now().Format("2006-01-02")

	// Update row: cols[5] = "RESOLVED", cols[6] = resDate, cols[7] = "`"+resCommit+"`", cols[8] = opts.Source
	cols[5] = "RESOLVED"
	cols[6] = resDate
	cols[7] = "`" + resCommit + "`"
	cols[8] = opts.Source

	lines[targetLineIndex] = strings.Join(cols, " | ")
	if err := os.WriteFile(registryPath, []byte(strings.Join(lines, "\n")+"\n"), 0644); err != nil {
		return fmt.Errorf("writing updated registry: %w", err)
	}

	// Commit registry update
	_, _ = git.Combined(repoRoot, "add", filepath.ToSlash(registryRel))
	_, _ = git.Combined(repoRoot, "commit", "-m", fmt.Sprintf("docs(audit): mark %s resolved in registry", auditFileName))

	fmt.Printf("[SUCCESS] Audit %s resolved in commit %s\n", auditFileName, resCommit)

	if opts.Push {
		fmt.Println("Pushing to remote...")
		if out, err := git.Combined(repoRoot, "push"); err != nil {
			return fmt.Errorf("git push failed: %s (%w)", string(out), err)
		}
	}

	return nil
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
