// Commands: hook, tdd, whats-up.

package main

import (
	"flag"
	"fmt"
	"os"

	"animus-mundi/pkg/toolchain/docs"
	"animus-mundi/pkg/toolchain/hook"
	"animus-mundi/pkg/toolchain/tddslice"
)

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
