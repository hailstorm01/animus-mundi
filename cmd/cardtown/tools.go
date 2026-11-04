// Commands: hook, whats-up.

package main

import (
	"flag"
	"fmt"
	"os"

	"animus-mundi/pkg/toolchain/docs"
	"animus-mundi/pkg/toolchain/hook"
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

func runWhatsUp(args []string) {
	fs := flag.NewFlagSet("whats-up", flag.ExitOnError)
	section := fs.String("section", "all", "Section to show: all, skills, tools, or state")
	fs.Parse(normalizeArgs(args))

	repoRoot := findRepoRoot()
	docs.WhatsUp(repoRoot, *section)
}
