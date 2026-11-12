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
//	go tool task next-directive
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

var (
	milestone = flag.String("Milestone", "m2", "milestone folder under .agents/tasks")
	base      = flag.String("Base", "main", "integration branch")
)

var lastDirective = regexp.MustCompile("(?m)^- \\*\\*Last directive:\\*\\* `#(\\d+)`")

// parseLast returns the board's Last directive number, or false if the board
// has no such line.
func parseLast(board string) (int, bool) {
	m := lastDirective.FindStringSubmatch(board)
	if m == nil {
		return 0, false
	}
	n, err := strconv.Atoi(m[1])
	return n, err == nil
}

func git(args ...string) (string, error) {
	out, err := exec.Command("git", args...).Output()
	return string(out), err
}

func main() {
	flag.Parse()
	path := ".agents/tasks/" + *milestone + "/BOARD.md"

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
		board, err := git("show", r+":"+path)
		if err != nil {
			if r == *base {
				fmt.Fprintf(os.Stderr, "nextdirective: no %s on %s\n", path, r)
				os.Exit(1)
			}
			continue // a run branch from before this milestone's board
		}
		n, ok := parseLast(board)
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
	fmt.Printf("(one after #%d on %s; Last directive by branch:)\n", highest, from)
	for _, l := range lines {
		fmt.Println(l)
	}
}
