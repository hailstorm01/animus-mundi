package cardsheet_test

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"animus-mundi/pkg/toolchain/linter/cardsheet"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func listFixture(t *testing.T) (root, a, b string) {
	t.Helper()
	root = t.TempDir()
	dir := filepath.Join(root, "data", "cards")
	a = filepath.Join(dir, "a.csv")
	b = filepath.Join(dir, "b.csv")
	writeFile(t, b, "name\n")
	writeFile(t, a, "name\n")
	writeFile(t, filepath.Join(dir, "notes.txt"), "x")
	if err := os.MkdirAll(filepath.Join(dir, "sub.csv"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	return root, a, b
}

// 1. Without paths, List returns the sorted non-directory csv files.
func TestList_Case1_Directory(t *testing.T) {
	root, a, b := listFixture(t)
	files, missing, err := cardsheet.List(root, nil)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !reflect.DeepEqual(files, []string{a, b}) {
		t.Fatalf("files = %v, want %v", files, []string{a, b})
	}
	if len(missing) != 0 {
		t.Fatalf("missing = %v", missing)
	}
}

// 2. With paths, List trims, skips blanks, reports missing ones and sorts.
func TestList_Case2_Paths(t *testing.T) {
	root, a, b := listFixture(t)
	files, missing, err := cardsheet.List(root, []string{"data/cards/b.csv", "  ", "data/cards/none.csv", a})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !reflect.DeepEqual(files, []string{a, b}) {
		t.Fatalf("files = %v, want %v", files, []string{a, b})
	}
	if !reflect.DeepEqual(missing, []string{"data/cards/none.csv"}) {
		t.Fatalf("missing = %v", missing)
	}
}

// 3. A missing cards directory is an error.
func TestList_Case3_NoDirectory(t *testing.T) {
	_, _, err := cardsheet.List(t.TempDir(), nil)
	if err == nil || !strings.Contains(err.Error(), "reading cards directory") {
		t.Fatalf("err = %v", err)
	}
}

// 4. Open returns the unwrapped os error for a missing file.
func TestOpen_Case4_Missing(t *testing.T) {
	_, err := cardsheet.Open(filepath.Join(t.TempDir(), "nope.csv"), false)
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("err = %v", err)
	}
}

// 5. An empty file gives io.EOF from Header.
func TestHeader_Case5_Empty(t *testing.T) {
	p := filepath.Join(t.TempDir(), "empty.csv")
	writeFile(t, p, "")
	s, err := cardsheet.Open(p, false)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer s.Close()
	if _, err := s.Header(); err != io.EOF {
		t.Fatalf("err = %v, want io.EOF", err)
	}
}

// 6. Ragged controls the field-count check.
func TestRows_Case6_Ragged(t *testing.T) {
	p := filepath.Join(t.TempDir(), "r.csv")
	writeFile(t, p, "name,power\nA,1\nB\n")

	s, err := cardsheet.Open(p, false)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if _, err := s.Header(); err != nil {
		t.Fatalf("header: %v", err)
	}
	if _, err := s.Rows(); err == nil {
		t.Fatalf("expected a field count error")
	}
	s.Close()

	s, err = cardsheet.Open(p, true)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer s.Close()
	if _, err := s.Header(); err != nil {
		t.Fatalf("header: %v", err)
	}
	rows, err := s.Rows()
	if err != nil {
		t.Fatalf("rows: %v", err)
	}
	if want := [][]string{{"A", "1"}, {"B"}}; !reflect.DeepEqual(rows, want) {
		t.Fatalf("rows = %v, want %v", rows, want)
	}
}

// 7. Column trims, strips a BOM, ignores case and takes the first match.
func TestColumn_Case7(t *testing.T) {
	h := []string{"\ufeffName", " rulesText ", "power"}
	cases := []struct {
		names []string
		want  int
	}{
		{[]string{"name"}, 0},
		{[]string{"text", "rulestext"}, 1},
		{[]string{"POWER"}, 2},
		{[]string{"health"}, -1},
	}
	for _, c := range cases {
		if got := cardsheet.Column(h, c.names...); got != c.want {
			t.Errorf("Column(%v) = %d, want %d", c.names, got, c.want)
		}
	}
}
