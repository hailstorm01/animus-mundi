package docs

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"animus-mundi/pkg/toolchain/audit"
)

var (
	frontmatterRegex = regexp.MustCompile(`(?s)^---\r?\n(.*?)\r?\n---`)
	nameRegex        = regexp.MustCompile(`(?m)^name:\s*(.+)$`)
	descBlockRegex   = regexp.MustCompile(`(?ms)^description:\s*[>|]-?[ \t]*\r?\n((?:[ \t]+[^\r\n]*(?:\r?\n|$))+)`)
	descOneRegex     = regexp.MustCompile(`(?m)^description:\s*(.+)$`)
	synopsisRegex    = regexp.MustCompile(`(?ms)\.SYNOPSIS\s*\r?\n\s*(\S[^\r\n]*)`)
)

func limitText(text string, max int) string {
	if len(text) <= max {
		return text
	}
	cut := text[:max]
	if space := strings.LastIndex(cut, " "); space > 0 {
		cut = cut[:space]
	}
	return cut + " ..."
}

func getGit(repoRoot string, args ...string) []string {
	cmdArgs := append([]string{"-C", repoRoot}, args...)
	cmd := exec.Command("git", cmdArgs...)
	out, err := cmd.Output()
	if err != nil {
		return nil
	}
	var lines []string
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		lines = append(lines, strings.TrimSpace(scanner.Text()))
	}
	return lines
}

// ShowSkills displays skills and personas in .agents/skills.
func ShowSkills(repoRoot string) {
	fmt.Println("== SKILLS / PERSONAS (.agents/skills) ==")
	skillsDir := filepath.Join(repoRoot, ".agents", "skills")

	var files []string
	_ = filepath.Walk(skillsDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && info.Name() == "SKILL.md" {
			files = append(files, path)
		}
		return nil
	})
	sort.Strings(files)

	for _, fPath := range files {
		content, err := os.ReadFile(fPath)
		if err != nil {
			continue
		}
		fmMatch := frontmatterRegex.FindSubmatch(content)
		if len(fmMatch) < 2 {
			continue
		}
		fm := string(fmMatch[1])
		name := ""
		if m := nameRegex.FindStringSubmatch(fm); len(m) > 1 {
			name = strings.TrimSpace(m[1])
		} else {
			name = filepath.Base(filepath.Dir(fPath))
		}

		desc := ""
		if m := descBlockRegex.FindStringSubmatch(fm); len(m) > 1 {
			lines := strings.Split(m[1], "\n")
			var parts []string
			for _, l := range lines {
				trimmed := strings.TrimSpace(l)
				if trimmed != "" {
					parts = append(parts, trimmed)
				}
			}
			desc = strings.Join(parts, " ")
		} else if m := descOneRegex.FindStringSubmatch(fm); len(m) > 1 {
			desc = strings.TrimSpace(m[1])
		}

		rel, _ := filepath.Rel(repoRoot, fPath)
		rel = filepath.ToSlash(rel)
		fmt.Printf("- %s (%s)\n", name, rel)
		fmt.Printf("    %s\n", limitText(desc, 220))
	}
	fmt.Println()
}

// ShowTools displays available CLI commands and tools.
func ShowTools(repoRoot string) {
	fmt.Println("== TOOLS ==")
	fmt.Println("- cardtown lint [flags] [files...] : Unified developer linter suite (FRD, TDD, CSV sheets & syntax)")
	fmt.Println("    cardtown lint [--all|--frd|--tdd|--cards|--sheets|--syntax|--staged] [--files <list>]")
	fmt.Println("- cardtown hook install : Installs native Git pre-commit hook into .git/hooks/pre-commit")
	fmt.Println("    cardtown hook install")
	fmt.Println("- cardtown audit status : Read-only status report for pending Card Town audits")
	fmt.Println("    cardtown audit status")
	fmt.Println("- cardtown audit resolve <AuditFile> : Resolves pending audit, updates registry, and commits")
	fmt.Println("    cardtown audit resolve [-dry-run] [-push] <AuditFile>")
	fmt.Println("- cardtown tdd slice [flags] : Topologically slice approved TDD into GitHub Milestone & Issues")
	fmt.Println("    cardtown tdd slice [-dry-run] [-publish] [-update-existing] <TddFile>")
	fmt.Println("- cardtown whats-up [flags] : Live capability and state snapshot for the docs workflow")
	fmt.Println("    cardtown whats-up [-section all|skills|tools|state]")
	fmt.Println()
}

// ShowState displays repository git status, hook status, and pending audits.
func ShowState(repoRoot string) {
	fmt.Println("== LIVE STATE ==")

	branchLines := getGit(repoRoot, "rev-parse", "--abbrev-ref", "HEAD")
	branch := "unknown"
	if len(branchLines) > 0 {
		branch = branchLines[0]
	}

	aheadLines := getGit(repoRoot, "rev-list", "--count", "@{u}..HEAD")
	ahead := "?"
	if len(aheadLines) > 0 {
		ahead = aheadLines[0]
	}
	fmt.Printf("Git: branch '%s', %s unpushed commit(s)\n", branch, ahead)

	changed := getGit(repoRoot, "status", "--porcelain", "--", "docs", "data/cards", "tools", ".agents", "cmd", "pkg")
	if len(changed) == 0 {
		fmt.Println("Uncommitted changes: none")
	} else {
		fmt.Printf("Uncommitted changes: %d\n", len(changed))
		limit := 8
		if len(changed) < limit {
			limit = len(changed)
		}
		for i := 0; i < limit; i++ {
			fmt.Printf("    %s\n", changed[i])
		}
		if len(changed) > 8 {
			fmt.Printf("    ... and %d more\n", len(changed)-8)
		}
	}

	hookDst := filepath.Join(repoRoot, ".git", "hooks", "pre-commit")
	if _, err := os.Stat(hookDst); os.IsNotExist(err) {
		fmt.Println("Pre-commit hook: NOT INSTALLED (run 'cardtown hook install')")
	} else {
		fmt.Println("Pre-commit hook: installed and current")
	}

	fmt.Println()
	fmt.Println("Pending audits:")
	_ = audit.Status(repoRoot)

	// Portfolio sync staleness check
	sumFile := filepath.Join(repoRoot, "docs", "PRODUCT_SUMMARY.md")
	if _, err := os.Stat(sumFile); err == nil {
		sumLog := getGit(repoRoot, "log", "-1", "--format=%ct", "--", "docs/PRODUCT_SUMMARY.md")
		if len(sumLog) > 0 {
			if sumTime, err := strconv.ParseInt(sumLog[0], 10, 64); err == nil {
				frdDir := filepath.Join(repoRoot, "docs", "frd")
				var staleFrds []string
				if frdEntries, err := os.ReadDir(frdDir); err == nil {
					for _, f := range frdEntries {
						if !f.IsDir() && strings.HasSuffix(f.Name(), ".md") {
							relFrd := "docs/frd/" + f.Name()
							frdLog := getGit(repoRoot, "log", "-1", "--format=%ct", "--", relFrd)
							if len(frdLog) > 0 {
								if frdTime, err := strconv.ParseInt(frdLog[0], 10, 64); err == nil && frdTime > sumTime {
									staleFrds = append(staleFrds, f.Name())
								}
							}
						}
					}
				}
				if len(staleFrds) > 0 {
					fmt.Printf("Portfolio roadmap (docs/PRODUCT_SUMMARY.md): OUT OF SYNC\n  Newer commits detected in: %s\n", strings.Join(staleFrds, ", "))
				} else {
					fmt.Println("Portfolio roadmap (docs/PRODUCT_SUMMARY.md): in sync with FRDs")
				}
			}
		}
	}

	tddDir := filepath.Join(repoRoot, "docs", "tdd")
	var tddFiles []string
	if entries, err := os.ReadDir(tddDir); err == nil {
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") && e.Name() != "README.md" {
				tddFiles = append(tddFiles, e.Name())
			}
		}
	}
	sort.Strings(tddFiles)

	if len(tddFiles) == 0 {
		fmt.Println("Technical Design Documents (docs/tdd/): none authored yet")
	} else {
		fmt.Printf("Technical Design Documents (docs/tdd/): %d authored\n", len(tddFiles))
		srcRe := regexp.MustCompile(`^\s*-\s+\*\*Source FRD:\*\*\s*(.+)$`)
		frdFileRe := regexp.MustCompile(`(\d{2}_[A-Za-z0-9_]+_FRD\.md)`)
		hashRe := regexp.MustCompile(`^\s*-\s+\*\*FRD Commit Hash:\*\*\s*` + "`" + `?([^` + "`" + `\s]+)` + "`" + `?`)
		sliceHeadingRe := regexp.MustCompile(`^###\s+\[SLICE-\d{2}-\d{2}\]`)

		for _, tf := range tddFiles {
			p := filepath.Join(tddDir, tf)
			content, err := os.ReadFile(p)
			if err != nil {
				continue
			}
			tLines := strings.Split(string(content), "\n")
			srcFrd := "?"
			tHash := "?"
			sliceCnt := 0

			for idx, tl := range tLines {
				if idx < 30 {
					if m := srcRe.FindStringSubmatch(tl); len(m) > 1 {
						if mf := frdFileRe.FindStringSubmatch(m[1]); len(mf) > 1 {
							srcFrd = mf[1]
						}
					}
					if m := hashRe.FindStringSubmatch(tl); len(m) > 1 {
						tHash = strings.TrimSpace(m[1])
					}
				}
				if sliceHeadingRe.MatchString(tl) {
					sliceCnt++
				}
			}

			tStale := "CURRENT"
			if srcFrd != "?" && len(tHash) >= 7 {
				newer := getGit(repoRoot, "log", "--format=%h", tHash+"..HEAD", "--", "docs/frd/"+srcFrd)
				if len(newer) > 0 {
					tStale = fmt.Sprintf("STALE (%d newer FRD commit(s))", len(newer))
				}
			}
			fmt.Printf("  - %s: Source=%s | Audited=%s | Freshness=%s | Slices=%d\n", tf, srcFrd, tHash, tStale, sliceCnt)
		}
	}
	fmt.Println()
}

// WhatsUp coordinates the snapshot report.
func WhatsUp(repoRoot string, section string) {
	switch strings.ToLower(section) {
	case "skills":
		ShowSkills(repoRoot)
	case "tools":
		ShowTools(repoRoot)
	case "state":
		ShowState(repoRoot)
	default:
		ShowSkills(repoRoot)
		ShowTools(repoRoot)
		ShowState(repoRoot)
	}
}
