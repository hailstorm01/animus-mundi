// Package board reads milestone build boards (.agents/tasks/m<N>/BOARD.md)
// from git, for the agent scripts.
//
// A closed milestone's board drops its Last directive line, so the counter
// moves to the next board (M2 closed on 2026-11-23 and M3 continues from
// #179). Current finds the board that still counts, which lets the scripts run
// without -Milestone in every milestone.
package board

import (
	"fmt"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Path is a milestone's board, relative to the repository root.
func Path(milestone string) string {
	return ".agents/tasks/" + milestone + "/BOARD.md"
}

var lastDirective = regexp.MustCompile("(?m)^- \\*\\*Last directive:\\*\\* `#(\\d+)`")

// LastDirective returns the board's Last directive number, or false if the
// board has no such line.
func LastDirective(board string) (int, bool) {
	m := lastDirective.FindStringSubmatch(board)
	if m == nil {
		return 0, false
	}
	n, err := strconv.Atoi(m[1])
	return n, err == nil
}

// Current returns the current milestone: the highest-numbered m<N> folder
// under .agents/tasks whose BOARD.md on base has a Last directive line.
func Current(base string) (string, error) {
	out, err := git("ls-tree", "-d", "--name-only", base, "--", ".agents/tasks/")
	if err != nil {
		return "", fmt.Errorf("git ls-tree %s .agents/tasks/: %w", base, err)
	}
	var folders []string
	for _, p := range strings.Fields(out) {
		folders = append(folders, strings.TrimPrefix(p, ".agents/tasks/"))
	}
	m, ok := current(folders, func(milestone string) bool {
		text, err := git("show", base+":"+Path(milestone))
		if err != nil {
			return false
		}
		_, ok := LastDirective(text)
		return ok
	})
	if !ok {
		return "", fmt.Errorf("no .agents/tasks/m<N>/BOARD.md on %s has a Last directive line", base)
	}
	return m, nil
}

var milestoneFolder = regexp.MustCompile(`^m(\d+)$`)

// current returns the highest-numbered m<N> folder that counts, trying the
// highest first.
func current(folders []string, counts func(milestone string) bool) (string, bool) {
	type folder struct {
		name string
		n    int
	}
	var ms []folder
	for _, f := range folders {
		if m := milestoneFolder.FindStringSubmatch(f); m != nil {
			n, _ := strconv.Atoi(m[1])
			ms = append(ms, folder{f, n})
		}
	}
	sort.Slice(ms, func(i, j int) bool { return ms[i].n > ms[j].n })
	for _, m := range ms {
		if counts(m.name) {
			return m.name, true
		}
	}
	return "", false
}

func git(args ...string) (string, error) {
	out, err := exec.Command("git", args...).Output()
	return string(out), err
}
