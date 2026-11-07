package specdoc

import (
	"path/filepath"
	"strings"
)

// MatrixFile is the keyword interaction matrix's file name, in docs/frd.
const MatrixFile = "keyword_interaction_matrix.md"

// MatrixItem is a row of the keyword matrix's Open Items or Rulings table.
type MatrixItem struct {
	ID    string // "KIM-C01"
	Open  bool   // in "## Open Items"; false for a row of "## Rulings"
	Reach string // an open item's Reach cell: "R" if reachable with current cards, "U" if not
}

// ReadMatrix returns the KIM- rows of the keyword matrix's Open Items and Rulings tables, in file order.
func ReadMatrix(repoRoot string) ([]MatrixItem, error) {
	lines, err := ReadLines(filepath.Join(repoRoot, "docs", "frd", MatrixFile))
	if err != nil {
		return nil, err
	}
	var items []MatrixItem
	section := ""
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "## ") {
			section = trimmed
			continue
		}
		open := strings.HasPrefix(section, "## Open Items")
		if !open && !strings.HasPrefix(section, "## Rulings") || !strings.HasPrefix(trimmed, "|") {
			continue
		}
		cells := strings.Split(trimmed, "|")
		if len(cells) < 3 {
			continue
		}
		id := strings.TrimSpace(cells[1])
		if !strings.HasPrefix(id, "KIM-") {
			continue
		}
		item := MatrixItem{ID: id, Open: open}
		if open && len(cells) > 3 {
			item.Reach = strings.TrimSpace(cells[3])
		}
		items = append(items, item)
	}
	return items, nil
}
