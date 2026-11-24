// Command nextdirective prints the number the next board directive takes.
//
// During a wave run, a milestone's BOARD.md lives on two branches: the
// integration branch and the autoloop/ run branch. Each has its own
// "Last directive" counter, and personas on either side add rows. Wave F
// session 18 stopped because main and the run branch each added a different
// #142. This reads the counter on the integration branch and on every local
// autoloop/ branch and prints one more than the highest, so either side picks
// a number the other hasn't used.
//
// The board is the current milestone's: the highest-numbered
// .agents/tasks/m<N>/ whose BOARD.md on the integration branch has a Last
// directive line. A closed milestone's board has none (M2's, since
// 2026-11-23). -Milestone picks another board.
//
//	go tool task next-directive
//	go tool task next-directive -- -Milestone m3
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"animus-mundi/.agents/scripts/internal/board"
)

var (
	milestone = flag.String("Milestone", "", "milestone folder under .agents/tasks (default: the highest m<N> whose BOARD.md on -Base has a Last directive line)")
	base      = flag.String("Base", "main", "integration branch")
)

func git(args ...string) (string, error) {
	out, err := exec.Command("git", args...).Output()
	return string(out), err
}

func main() {
	flag.Parse()
	if *milestone == "" {
		m, err := board.Current(*base)
		if err != nil {
			fmt.Fprintln(os.Stderr, "nextdirective:", err)
			os.Exit(1)
		}
		*milestone = m
	}
	path := board.Path(*milestone)

	refs := []string{*base}
	out, err := git("for-each-ref", "--format=%(refname:short)", "refs/heads/autoloop/")
	if err != nil {
		fmt.Fprintln(os.Stderr, "nextdirective: git for-each-ref:", err)
		os.Exit(1)
	}
	for _, r := range strings.Fields(out) {
		if r != *base {
			refs = append(refs, r)
		}
	}

	highest, from := -1, ""
	var lines []string
	for _, r := range refs {
		text, err := git("show", r+":"+path)
		if err != nil {
			if r == *base {
				fmt.Fprintf(os.Stderr, "nextdirective: no %s on %s\n", path, r)
				os.Exit(1)
			}
			continue // a run branch from before this milestone's board
		}
		n, ok := board.LastDirective(text)
		if !ok {
			fmt.Fprintf(os.Stderr, "nextdirective: no Last directive line in %s:%s\n", r, path)
			os.Exit(1)
		}
		lines = append(lines, fmt.Sprintf("  %s: #%d", r, n))
		if n > highest {
			highest, from = n, r
		}
	}

	fmt.Printf("#%d\n", highest+1)
	fmt.Printf("(%s board; one after #%d on %s; Last directive by branch:)\n", *milestone, highest, from)
	for _, l := range lines {
		fmt.Println(l)
	}
}
