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
	"animus-mundi/pkg/toolchain/linter/tdd"
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
		case "tdd":
			ok, err := tdd.Lint(repoRoot, remaining)
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
	tddFlag := fs.Bool("tdd", false, "Run TDD linter")
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

	runFrd := false
	runTdd := false
	runRules := false
	runSheets := false
	runSyntax := false
	runFences := false

	var changedFrds []string
	var changedTdds []string
	var changedCsvs []string

	if *allFlag {
		runFrd = true
		runTdd = true
		runRules = true
		runSheets = true
		runSyntax = true
		runFences = true
	} else if *frdFlag || *tddFlag || *rulesFlag || *cardsFlag || *sheetsFlag || *syntaxFlag || *fencesFlag {
		if *frdFlag {
			runFrd = true
		}
		if *tddFlag {
			runTdd = true
		}
		if *rulesFlag {
			runRules = true
		}
		if *cardsFlag || *sheetsFlag {
			runSheets = true
		}
		if *cardsFlag || *syntaxFlag {
			runSyntax = true
		}
		if *fencesFlag {
			runFences = true
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
				runRules = true
				runSheets = true
				runSyntax = true
			} else {
				for _, path := range staged {
					pLower := strings.ToLower(filepath.ToSlash(path))
					if isRulesPath(pLower) {
						runRules = true
					}
					if isFencesPath(pLower) {
						runFences = true
					}
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
				runRules = true
				runSheets = true
				runSyntax = true
			} else {
				for _, sf := range statusFiles {
					pLower := strings.ToLower(filepath.ToSlash(sf.Path))
					if isRulesPath(pLower) {
						runRules = true
					}
					if isFencesPath(pLower) {
						runFences = true
					}
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

		if !runFrd && !runTdd && !runRules && !runSheets && !runSyntax && !runFences {
			fmt.Println("  No changes detected in docs/frd/, docs/tdd/, pkg/, or data/cards/.")
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

	// 2b. Executable Rules Traceability Linter
	if runRules {
		fmt.Println("\n>>> Running Executable Rules Traceability Linter...")
		ok, err := rules.Lint(repoRoot, nil)
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

	// 5. Engine Architecture Fences Linter
	if runFences {
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

// isFencesPath reports whether a change can affect the package fences.
func isFencesPath(pLower string) bool {
	return (strings.HasPrefix(pLower, "pkg/") && !strings.HasPrefix(pLower, "pkg/toolchain/") && strings.HasSuffix(pLower, ".go")) ||
		(strings.HasPrefix(pLower, "data/cards/") && strings.HasSuffix(pLower, ".csv")) ||
		pLower == "docs/frd/keyword_interaction_matrix.md" ||
		pLower == "docs/design/m2_starter_set_ability_survey.md"
}
