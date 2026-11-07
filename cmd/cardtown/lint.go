// Command: lint.

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"animus-mundi/pkg/toolchain/git"
	"animus-mundi/pkg/toolchain/linter/cards"
	"animus-mundi/pkg/toolchain/linter/fences"
	"animus-mundi/pkg/toolchain/linter/frd"
	"animus-mundi/pkg/toolchain/linter/rules"
	"animus-mundi/pkg/toolchain/linter/syntax"
	"animus-mundi/pkg/toolchain/specdoc"
)

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
		case "rules":
			ok, err := rules.Lint(repoRoot, remaining)
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
		case "fences":
			ok, err := fences.Lint(repoRoot, remaining)
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
	rulesFlag := fs.Bool("rules", false, "Run executable rules traceability linter")
	cardsFlag := fs.Bool("cards", false, "Run sheet integrity & syntax linters")
	sheetsFlag := fs.Bool("sheets", false, "Run sheet integrity linter only")
	syntaxFlag := fs.Bool("syntax", false, "Run rules syntax linter only")
	fencesFlag := fs.Bool("fences", false, "Run engine architecture fences linter")
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

	var sel lintSelection
	if *allFlag {
		sel = allLinters()
	} else if *frdFlag || *rulesFlag || *cardsFlag || *sheetsFlag || *syntaxFlag || *fencesFlag {
		sel.frd = *frdFlag
		sel.rules = *rulesFlag
		sel.sheets = *cardsFlag || *sheetsFlag
		sel.syntax = *cardsFlag || *syntaxFlag
		sel.fences = *fencesFlag
	} else {
		// Smart change detection
		if *stagedFlag {
			fmt.Println("[Detecting Staged Files via Git...]")
		} else {
			fmt.Println("[Detecting Changed Files via Git...]")
		}
		paths, err := changedFiles(repoRoot, *stagedFlag)
		if err != nil && *stagedFlag {
			fmt.Printf("Git staged diff unavailable: %v. Running all linters...\n", err)
		} else if err != nil {
			fmt.Println("  Git change detection unavailable. Running all linters...")
		}
		sel = selectChanged(paths, err)

		if len(sel.changedFrds) > 0 {
			fmt.Printf("  Detected %d modified/added FRD document(s).\n", len(sel.changedFrds))
		}
		if len(sel.changedCsvs) > 0 {
			fmt.Printf("  Detected %d modified/added Card CSV sheet(s).\n", len(sel.changedCsvs))
		}

		if sel.empty() {
			fmt.Println("  No changes detected in docs/frd/, pkg/, or data/cards/.")
			fmt.Println("  Nothing to lint! (Use --all to force full audit across all files)")
			fmt.Println()
			fmt.Println("==========================================================")
			os.Exit(0)
		}
	}

	anyFailed := false

	// 1. FRD Linter
	if sel.frd {
		fmt.Println("\n>>> Running FRD Requirements Linter...")
		files := parsedFiles
		if len(files) == 0 && len(sel.changedFrds) > 0 && !*allFlag {
			files = sel.changedFrds
		}
		ok, err := frd.Lint(repoRoot, files)
		if err != nil || !ok {
			anyFailed = true
		}
	}

	// 2. Executable Rules Traceability Linter
	if sel.rules {
		fmt.Println("\n>>> Running Executable Rules Traceability Linter...")
		ok, err := rules.Lint(repoRoot, nil)
		if err != nil || !ok {
			anyFailed = true
		}
	}

	// 3. Card Sheet CSV Integrity Linter
	if sel.sheets {
		fmt.Println("\n>>> Running Card Sheet CSV Integrity Linter...")
		files := parsedFiles
		if len(files) == 0 && len(sel.changedCsvs) > 0 && !*allFlag {
			files = sel.changedCsvs
		}
		ok, err := cards.Lint(repoRoot, files)
		if err != nil || !ok {
			anyFailed = true
		}
	}

	// 4. Card Rules Syntax & Terminology Linter
	if sel.syntax {
		fmt.Println("\n>>> Running Card Rules Syntax & Terminology Linter...")
		files := parsedFiles
		if len(files) == 0 && len(sel.changedCsvs) > 0 && !*allFlag {
			files = sel.changedCsvs
		}
		ok, err := syntax.Lint(repoRoot, files, true)
		if err != nil || !ok {
			anyFailed = true
		}
	}

	// 5. Engine Architecture Fences Linter
	if sel.fences {
		fmt.Println("\n>>> Running Engine Architecture Fences Linter...")
		ok, err := fences.Lint(repoRoot, nil)
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

// isRulesPath reports whether a change can affect executable rules traceability.
func isRulesPath(pLower string) bool {
	return pLower == "docs/frd/01_game_rules_frd.md" || (strings.HasPrefix(pLower, "pkg/engine/") && strings.HasSuffix(pLower, ".go"))
}

// lintSelection is which linters a lint run runs, and the changed FRD documents and card
// sheets that the FRD, sheet and syntax linters then read.
type lintSelection struct {
	frd, rules, sheets, syntax, fences bool
	changedFrds, changedCsvs           []string
}

// allLinters selects every linter.
func allLinters() lintSelection {
	return lintSelection{frd: true, rules: true, sheets: true, syntax: true, fences: true}
}

// empty reports whether s selects no linter.
func (s lintSelection) empty() bool {
	return !s.frd && !s.rules && !s.sheets && !s.syntax && !s.fences
}

// changedFiles lists the changed paths git reports: the staged files when staged is set,
// otherwise every path git status shows.
func changedFiles(repoRoot string, staged bool) ([]string, error) {
	if staged {
		return git.GetStagedFiles(repoRoot)
	}
	statusFiles, err := git.GetStatus(repoRoot)
	if err != nil {
		return nil, err
	}
	var paths []string
	for _, sf := range statusFiles {
		paths = append(paths, sf.Path)
	}
	return paths, nil
}

// selectChanged selects the linters that changes to paths can affect, or every linter when
// git couldn't list the changes (gitErr is not nil).
func selectChanged(paths []string, gitErr error) lintSelection {
	if gitErr != nil {
		return allLinters()
	}
	var s lintSelection
	for _, path := range paths {
		pLower := strings.ToLower(filepath.ToSlash(path))
		if isRulesPath(pLower) {
			s.rules = true
		}
		if isFencesPath(pLower) {
			s.fences = true
		}
		if strings.HasPrefix(pLower, "docs/frd/") && strings.HasSuffix(pLower, ".md") {
			s.changedFrds = append(s.changedFrds, path)
		} else if strings.HasPrefix(pLower, "data/cards/") && strings.HasSuffix(pLower, ".csv") {
			s.changedCsvs = append(s.changedCsvs, path)
		}
	}
	s.frd = len(s.changedFrds) > 0
	s.sheets = len(s.changedCsvs) > 0
	s.syntax = s.sheets
	return s
}

// isFencesPath reports whether a change can affect the package fences.
func isFencesPath(pLower string) bool {
	return (strings.HasPrefix(pLower, "pkg/") && !strings.HasPrefix(pLower, "pkg/toolchain/") && strings.HasSuffix(pLower, ".go")) ||
		(strings.HasPrefix(pLower, "data/cards/") && strings.HasSuffix(pLower, ".csv")) ||
		pLower == "docs/frd/"+specdoc.MatrixFile ||
		pLower == "docs/design/"+fences.SurveyFile
}
