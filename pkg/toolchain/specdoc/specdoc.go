// Package specdoc holds what the toolchain shares when it scans a file: where the FRDs are, requirement IDs, the lines that define them, reading a file as lines, and the keyword matrix's items.
package specdoc

import (
	"bufio"
	"os"
	"regexp"
)

const (
	// FRDDir is the directory that holds the FRDs and the keyword matrix,
	// relative to the repository root and slash-separated.
	FRDDir = "docs/frd"
	// RulesFRD is FRD 01's file name, in FRDDir: the single source of truth for
	// game rules, which owns every game rule and the keyword dictionary.
	RulesFRD = "01_Game_Rules_FRD.md"
)

var (
	// ReqRegex matches a requirement ID.
	ReqRegex = regexp.MustCompile(`FR-[A-Z]{3}-\d{3}[a-z]?`)
	// DefLineRegex matches a line that can define a requirement: a heading, or a line, bullet or table row that opens with a bold ID.
	DefLineRegex = regexp.MustCompile(`^(####|###|\*\*FR-|- \*\*FR-|\| \*\*FR-)`)
)

// ReadLines returns the lines of the file at path.
func ReadLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	// Accommodate very long lines if needed (up to 1MB)
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, 1024*1024)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}
