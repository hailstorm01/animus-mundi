// Package specdoc holds what the toolchain linters share when they scan a file: requirement IDs, the lines that define them, and reading a file as lines.
package specdoc

import (
	"bufio"
	"os"
	"regexp"
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
