// Package cardsheet finds and reads the data/cards CSV sheets for the toolchain linters. pkg/engine has its own loader.
package cardsheet

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// List returns the sorted card sheet files to audit. With paths, each trimmed non-blank
// path (joined to repoRoot when relative) that exists goes in files; the others go in
// missing, as trimmed, in input order. Without paths it lists data/cards/*.csv.
func List(repoRoot string, paths []string) (files, missing []string, err error) {
	if len(paths) > 0 {
		for _, f := range paths {
			trimmed := strings.TrimSpace(f)
			if trimmed == "" {
				continue
			}
			p := trimmed
			if !filepath.IsAbs(p) {
				p = filepath.Join(repoRoot, p)
			}
			if _, statErr := os.Stat(p); statErr == nil {
				files = append(files, p)
			} else {
				missing = append(missing, trimmed)
			}
		}
	} else {
		cardsDir := filepath.Join(repoRoot, "data", "cards")
		entries, readErr := os.ReadDir(cardsDir)
		if readErr != nil {
			return nil, nil, fmt.Errorf("reading cards directory: %w", readErr)
		}
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".csv") {
				files = append(files, filepath.Join(cardsDir, e.Name()))
			}
		}
	}
	sort.Strings(files)
	return files, missing, nil
}

// Sheet is an open card sheet. Read the header first, then the rows.
type Sheet struct {
	file *os.File
	r    *csv.Reader
}

// Open opens a card sheet. With ragged, rows may have any number of fields.
func Open(path string, ragged bool) (*Sheet, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	r := csv.NewReader(f)
	if ragged {
		r.FieldsPerRecord = -1
	}
	return &Sheet{file: f, r: r}, nil
}

// Header returns the first record, untrimmed. An empty file gives io.EOF.
func (s *Sheet) Header() ([]string, error) {
	return s.r.Read()
}

// Rows returns the records after the header.
func (s *Sheet) Rows() ([][]string, error) {
	return s.r.ReadAll()
}

// Close closes the underlying file.
func (s *Sheet) Close() error {
	return s.file.Close()
}

// Column returns the index of the first header cell matching any of names, ignoring
// case, surrounding spaces and a leading byte-order mark. It returns -1 if none match.
func Column(header []string, names ...string) int {
	for i, h := range header {
		clean := strings.TrimSpace(strings.TrimPrefix(h, "\ufeff"))
		for _, n := range names {
			if strings.EqualFold(clean, n) {
				return i
			}
		}
	}
	return -1
}
