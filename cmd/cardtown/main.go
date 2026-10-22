package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
	case "play":
		os.Exit(runPlay(os.Args[2:], os.Stdin, os.Stdout))
	case "lint":
		runLint(os.Args[2:])
	case "hook":
		runHook(os.Args[2:])
	case "audit":
		runAudit(os.Args[2:])
	case "tdd":
		runTdd(os.Args[2:])
	case "replay":
		runReplay(os.Args[2:])
	case "sim":
		runSim(os.Args[2:])
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
	fmt.Println("  play      Play a text-mode match against a sim policy, drawn only from PlayerView")
	fmt.Println("  lint      Unified developer linter suite (FRD, TDD, executable rules, CSV sheets, rules syntax, fences)")
	fmt.Println("  hook      Manage Git lifecycle hooks (install pre-commit)")
	fmt.Println("  audit     Manage PM Devil's Advocate audit lifecycle (status, resolve)")
	fmt.Println("  tdd       Technical Design Document tools (slicing & GitHub issue publishing)")
	fmt.Println("  sim       Random-policy playouts on the rules kernel (termination, invariants, win rates)")
	fmt.Println("  replay    Rebuild a recorded game from a sim -events file and print its state at any event")
	fmt.Println("  whats-up  Live capability and state snapshot for the docs workflow")
	fmt.Println()
	fmt.Println("Run 'cardtown <command> -h' for command-specific flags and options.")
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
			case "fences":
				lower = "fences"
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
			case "rules":
				lower = "rules"
			case "deck":
				lower = "deck"
			case "games":
				lower = "games"
			case "seed":
				lower = "seed"
			case "p0":
				lower = "p0"
			case "p1":
				lower = "p1"
			case "seat":
				lower = "seat"
			case "policy":
				lower = "policy"
			case "maxturns", "max-turns":
				lower = "max-turns"
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
