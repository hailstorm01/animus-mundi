package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"animus-mundi/pkg/toolchain/audit"
	"animus-mundi/pkg/toolchain/docs"
	"animus-mundi/pkg/toolchain/git"
	"animus-mundi/pkg/toolchain/hook"
	"animus-mundi/pkg/toolchain/linter/cards"
	"animus-mundi/pkg/toolchain/linter/frd"
	"animus-mundi/pkg/toolchain/linter/syntax"
	"animus-mundi/pkg/toolchain/linter/tdd"
	"animus-mundi/pkg/toolchain/tddslice"
)

func findRepoRoot() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "."
	}
	dir := cwd
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return cwd
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	switch command {
	case "lint":
		runLint(os.Args[2:])
	case "hook":
		runHook(os.Args[2:])
	case "audit":
		runAudit(os.Args[2:])
	case "tdd":
		runTdd(os.Args[2:])
	case "whats-up", "whatsup", "docs":
		runWhatsUp(os.Args[2:])
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Printf("Unknown command %q\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Card Town Unified Developer CLI (cardtown)")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  cardtown <command> [subcommand] [flags] [arguments]")
	fmt.Println()
	fmt.Println("Available Commands:")
	fmt.Println("  lint      Unified developer linter suite (FRD, TDD, CSV sheets, rules syntax)")
	fmt.Println("  hook      Manage Git lifecycle hooks (install pre-commit)")
	fmt.Println("  audit     Manage PM Devil's Advocate audit lifecycle (status, resolve)")
	fmt.Println("  tdd       Technical Design Document tools (slicing & GitHub issue publishing)")
	fmt.Println("  whats-up  Live capability and state snapshot for the docs workflow")
	fmt.Println()
	fmt.Println("Run 'cardtown <command> -h' for command-specific flags and options.")
}

func runHook(args []string) {
	if len(args) == 0 || args[0] != "install" {
		fmt.Println("Usage: cardtown hook install")
		os.Exit(1)
	}
	repoRoot := findRepoRoot()
	if err := hook.Install(repoRoot); err != nil {
		fmt.Printf("[ERROR] %v\n", err)
		os.Exit(1)
	}
}

func runAudit(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: cardtown audit [status | resolve]")
		os.Exit(1)
	}
	repoRoot := findRepoRoot()
	sub := args[0]
	switch sub {
	case "status":
		if err := audit.Status(repoRoot); err != nil {
			fmt.Printf("[ERROR] %v\n", err)
			os.Exit(1)
		}
	case "resolve":
		fs := flag.NewFlagSet("audit resolve", flag.ExitOnError)
		source := fs.String("source", "Agentic (pm-audit-reviewer)", "Resolution source label")
		msg := fs.String("msg", "", "Custom commit message for remediation")
		push := fs.Bool("push", false, "Run git push after resolution")
		dryRun := fs.Bool("dry-run", false, "Validate without committing or modifying files")
		fs.Parse(normalizeArgs(args[1:]))

		if fs.NArg() == 0 {
			fmt.Println("Usage: cardtown audit resolve [-dry-run] [-push] [-source <label>] [-msg <msg>] <auditFile>")
			os.Exit(1)
		}
		auditFile := fs.Arg(0)
		opts := audit.ResolveOptions{
			AuditFile:     auditFile,
			Source:        *source,
			CommitMessage: *msg,
			Push:          *push,
			DryRun:        *dryRun,
		}
		if err := audit.Resolve(repoRoot, opts); err != nil {
			fmt.Printf("[ERROR] %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Printf("Unknown audit subcommand %q. Use 'status' or 'resolve'.\n", sub)
		os.Exit(1)
	}
}

func runTdd(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: cardtown tdd [slice]")
		os.Exit(1)
	}
	repoRoot := findRepoRoot()
	sub := args[0]
	switch sub {
	case "slice":
		fs := flag.NewFlagSet("tdd slice", flag.ExitOnError)
		repo := fs.String("repo", "", "Optional GitHub owner/repo override")
		tddFileFlag := fs.String("tdd-file", "", "Path to target TDD markdown file")
		dryRun := fs.Bool("dry-run", true, "Preview mode (default)")
		publish := fs.Bool("publish", false, "Execute live creation/synchronization to GitHub")
		updateExisting := fs.Bool("update-existing", false, "Update existing issues instead of skipping")
		showBodies := fs.Bool("show-bodies", false, "Print rendered Markdown body for all slices")
		fs.Parse(normalizeArgs(args[1:]))

		tddFile := *tddFileFlag
		if tddFile == "" && fs.NArg() > 0 {
			tddFile = fs.Arg(0)
		}
		if tddFile == "" {
			fmt.Println("Usage: cardtown tdd slice [-dry-run] [-publish] [-update-existing] [-show-bodies] [-repo <owner/repo>] <TddFile>")
			os.Exit(1)
		}
		opts := tddslice.SliceOptions{
			TddFile:        tddFile,
			Repo:           *repo,
			DryRun:         *dryRun && !*publish,
			Publish:        *publish,
			UpdateExisting: *updateExisting,
			ShowBodies:     *showBodies,
		}
		if err := tddslice.Slice(repoRoot, opts); err != nil {
			fmt.Printf("[ERROR] %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Printf("Unknown tdd subcommand %q. Use 'slice'.\n", sub)
		os.Exit(1)
	}
}

func runWhatsUp(args []string) {
	fs := flag.NewFlagSet("whats-up", flag.ExitOnError)
	section := fs.String("section", "all", "Section to show: all, skills, tools, or state")
	fs.Parse(normalizeArgs(args))

	repoRoot := findRepoRoot()
	docs.WhatsUp(repoRoot, *section)
}

func runLint(args []string) {
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		sub := args[0]
		remaining := args[1:]
		repoRoot := findRepoRoot()

		switch sub {
		case "frd":
			ok, err := frd.Lint(repoRoot, remaining)
			if err != nil || !ok {
				os.Exit(1)
			}
			os.Exit(0)
		case "tdd":
			ok, err := tdd.Lint(repoRoot, remaining)
			if err != nil || !ok {
				os.Exit(1)
			}
			os.Exit(0)
		case "sheets":
			ok, err := cards.Lint(repoRoot, remaining)
			if err != nil || !ok {
				os.Exit(1)
			}
			os.Exit(0)
		case "syntax":
			ok, err := syntax.Lint(repoRoot, remaining, true)
			if err != nil || !ok {
				os.Exit(1)
			}
			os.Exit(0)
		case "cards":
			ok1, err1 := cards.Lint(repoRoot, remaining)
			fmt.Println()
			ok2, err2 := syntax.Lint(repoRoot, remaining, true)
			if err1 != nil || err2 != nil || !ok1 || !ok2 {
				os.Exit(1)
			}
			os.Exit(0)
		}
	}

	fs := flag.NewFlagSet("lint", flag.ExitOnError)
	allFlag := fs.Bool("all", false, "Force full audit across all files")
	frdFlag := fs.Bool("frd", false, "Run FRD linter")
	tddFlag := fs.Bool("tdd", false, "Run TDD linter")
	cardsFlag := fs.Bool("cards", false, "Run sheet integrity & syntax linters")
	sheetsFlag := fs.Bool("sheets", false, "Run sheet integrity linter only")
	syntaxFlag := fs.Bool("syntax", false, "Run rules syntax linter only")
	stagedFlag := fs.Bool("staged", false, "Smart-run only on Git staged files")
	filesFlag := fs.String("files", "", "Comma-separated list of target files")
	repoRootFlag := fs.String("repo-root", "", "Path to repository root")
	normalizedArgs := normalizeArgs(args)
	fs.Parse(normalizedArgs)

	repoRoot := *repoRootFlag
	if repoRoot == "" {
		repoRoot = findRepoRoot()
	}

	var parsedFiles []string
	if *filesFlag != "" {
		for _, part := range strings.Split(*filesFlag, ",") {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				parsedFiles = append(parsedFiles, trimmed)
			}
		}
	}
	parsedFiles = append(parsedFiles, fs.Args()...)

	fmt.Println("==========================================================")
	fmt.Println("               CARD TOWN LINTER DISPATCHER                ")
	fmt.Println("==========================================================")
	fmt.Println("Repo Root:", repoRoot)
	fmt.Println()

	runFrd := false
	runTdd := false
	runSheets := false
	runSyntax := false

	var changedFrds []string
	var changedTdds []string
	var changedCsvs []string

	if *allFlag {
		runFrd = true
		runTdd = true
		runSheets = true
		runSyntax = true
	} else if *frdFlag || *tddFlag || *cardsFlag || *sheetsFlag || *syntaxFlag {
		if *frdFlag {
			runFrd = true
		}
		if *tddFlag {
			runTdd = true
		}
		if *cardsFlag || *sheetsFlag {
			runSheets = true
		}
		if *cardsFlag || *syntaxFlag {
			runSyntax = true
		}
	} else {
		// Smart change detection
		if *stagedFlag {
			fmt.Println("[Detecting Staged Files via Git...]")
			staged, err := git.GetStagedFiles(repoRoot)
			if err != nil {
				fmt.Printf("Git staged diff unavailable: %v. Running all linters...\n", err)
				runFrd = true
				runTdd = true
				runSheets = true
				runSyntax = true
			} else {
				for _, path := range staged {
					pLower := strings.ToLower(filepath.ToSlash(path))
					if strings.HasPrefix(pLower, "docs/frd/") && strings.HasSuffix(pLower, ".md") {
						changedFrds = append(changedFrds, path)
					} else if strings.HasPrefix(pLower, "docs/tdd/") && strings.HasSuffix(pLower, ".md") && filepath.Base(pLower) != "readme.md" {
						changedTdds = append(changedTdds, path)
					} else if strings.HasPrefix(pLower, "data/cards/") && strings.HasSuffix(pLower, ".csv") {
						changedCsvs = append(changedCsvs, path)
					}
				}
			}
		} else {
			fmt.Println("[Detecting Changed Files via Git...]")
			statusFiles, err := git.GetStatus(repoRoot)
			if err != nil {
				fmt.Println("  Git change detection unavailable. Running all linters...")
				runFrd = true
				runTdd = true
				runSheets = true
				runSyntax = true
			} else {
				for _, sf := range statusFiles {
					pLower := strings.ToLower(filepath.ToSlash(sf.Path))
					if strings.HasPrefix(pLower, "docs/frd/") && strings.HasSuffix(pLower, ".md") {
						changedFrds = append(changedFrds, sf.Path)
					} else if strings.HasPrefix(pLower, "docs/tdd/") && strings.HasSuffix(pLower, ".md") && filepath.Base(pLower) != "readme.md" {
						changedTdds = append(changedTdds, sf.Path)
					} else if strings.HasPrefix(pLower, "data/cards/") && strings.HasSuffix(pLower, ".csv") {
						changedCsvs = append(changedCsvs, sf.Path)
					}
				}
			}
		}

		if len(changedFrds) > 0 {
			fmt.Printf("  Detected %d modified/added FRD document(s).\n", len(changedFrds))
			runFrd = true
		}
		if len(changedTdds) > 0 {
			fmt.Printf("  Detected %d modified/added TDD document(s).\n", len(changedTdds))
			runTdd = true
		}
		if len(changedCsvs) > 0 {
			fmt.Printf("  Detected %d modified/added Card CSV sheet(s).\n", len(changedCsvs))
			runSheets = true
			runSyntax = true
		}

		if !runFrd && !runTdd && !runSheets && !runSyntax {
			fmt.Println("  No changes detected in docs/frd/, docs/tdd/, or data/cards/.")
			fmt.Println("  Nothing to lint! (Use --all to force full audit across all files)")
			fmt.Println()
			fmt.Println("==========================================================")
			os.Exit(0)
		}
	}

	anyFailed := false

	// 1. FRD Linter
	if runFrd {
		fmt.Println("\n>>> Running FRD Requirements Linter...")
		files := parsedFiles
		if len(files) == 0 && len(changedFrds) > 0 && !*allFlag {
			files = changedFrds
		}
		ok, err := frd.Lint(repoRoot, files)
		if err != nil || !ok {
			anyFailed = true
		}
	}

	// 2. TDD Linter
	if runTdd {
		fmt.Println("\n>>> Running TDD Coverage & Slice Manifest Linter...")
		files := parsedFiles
		if len(files) == 0 && len(changedTdds) > 0 && !*allFlag {
			files = changedTdds
		}
		ok, err := tdd.Lint(repoRoot, files)
		if err != nil || !ok {
			anyFailed = true
		}
	}

	// 3. Card Sheet CSV Integrity Linter
	if runSheets {
		fmt.Println("\n>>> Running Card Sheet CSV Integrity Linter...")
		files := parsedFiles
		if len(files) == 0 && len(changedCsvs) > 0 && !*allFlag {
			files = changedCsvs
		}
		ok, err := cards.Lint(repoRoot, files)
		if err != nil || !ok {
			anyFailed = true
		}
	}

	// 4. Card Rules Syntax & Terminology Linter
	if runSyntax {
		fmt.Println("\n>>> Running Card Rules Syntax & Terminology Linter...")
		files := parsedFiles
		if len(files) == 0 && len(changedCsvs) > 0 && !*allFlag {
			files = changedCsvs
		}
		ok, err := syntax.Lint(repoRoot, files, true)
		if err != nil || !ok {
			anyFailed = true
		}
	}

	fmt.Println("\n==========================================================")
	if anyFailed {
		fmt.Println("  LINTER SUMMARY: One or more checks FAILED!             ")
		fmt.Println("==========================================================")
		os.Exit(1)
	} else {
		fmt.Println("  LINTER SUMMARY: All executed checks PASSED!             ")
		fmt.Println("==========================================================")
		os.Exit(0)
	}
}

func normalizeArgs(args []string) []string {
	var normalized []string
	for _, a := range args {
		if strings.HasPrefix(a, "-") {
			parts := strings.SplitN(a, "=", 2)
			flagName := strings.TrimLeft(parts[0], "-")
			lower := strings.ToLower(flagName)
			switch lower {
			case "all":
				lower = "all"
			case "frd":
				lower = "frd"
			case "tdd":
				lower = "tdd"
			case "cards":
				lower = "cards"
			case "sheets":
				lower = "sheets"
			case "syntax":
				lower = "syntax"
			case "staged":
				lower = "staged"
			case "files":
				lower = "files"
			case "reporoot", "repo-root":
				lower = "repo-root"
			case "dryrun", "dry-run":
				lower = "dry-run"
			case "publish":
				lower = "publish"
			case "updateexisting", "update-existing":
				lower = "update-existing"
			case "showbodies", "show-bodies":
				lower = "show-bodies"
			case "tddfile", "tdd-file":
				lower = "tdd-file"
			case "source":
				lower = "source"
			case "msg", "commitmessage", "commit-message":
				lower = "msg"
			case "push":
				lower = "push"
			case "section":
				lower = "section"
			}
			newFlag := "--" + lower
			if len(parts) == 2 {
				normalized = append(normalized, newFlag+"="+parts[1])
			} else {
				normalized = append(normalized, newFlag)
			}
		} else {
			normalized = append(normalized, a)
		}
	}
	return normalized
}

