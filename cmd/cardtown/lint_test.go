package main

import "testing"

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
)

// 37. isFencesPath covers pkg/ Go files except pkg/toolchain/, plus the data inputs.
func TestIsFencesPath(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		{"pkg/sim/x.go", true},
		{"pkg/engine/x.go", true},
		{"pkg/engine/cards/animus.go", true},
		{"pkg/foo/bar/x.go", true},
		{"pkg/sim/x_test.go", true},
		{"data/cards/units.csv", true},
		{"docs/frd/keyword_interaction_matrix.md", true},
		{"docs/design/m2_starter_set_ability_survey.md", true},
		{"pkg/toolchain/linter/fences/fences.go", false},
		{"cmd/cardtown/main.go", false},
		{"pkg/sim/readme.md", false},
		{"data/decks/x.txt", false},
	}
	for _, c := range cases {
		if got := isFencesPath(c.path); got != c.want {
			t.Errorf("isFencesPath(%q) = %v, want %v", c.path, got, c.want)
		}
	}
}

func TestSelectChanged_Case1_Paths(t *testing.T) {
	cases := []struct {
		name  string
		paths []string
		want  lintSelection
	}{
		{
			name:  "nil",
			paths: nil,
			want:  lintSelection{},
		},
		{
			name:  "engine file",
			paths: []string{"pkg/engine/x.go"},
			want:  lintSelection{rules: true, fences: true},
		},
		{
			name:  "FRD 01",
			paths: []string{"docs/frd/01_Game_Rules_FRD.md"},
			want:  lintSelection{frd: true, rules: true, changedFrds: []string{"docs/frd/01_Game_Rules_FRD.md"}},
		},
		{
			name:  "card units",
			paths: []string{"data/cards/units.csv"},
			want:  lintSelection{sheets: true, syntax: true, fences: true, changedCsvs: []string{"data/cards/units.csv"}},
		},
		{
			name:  "keyword interaction matrix",
			paths: []string{"docs/frd/keyword_interaction_matrix.md"},
			want:  lintSelection{frd: true, fences: true, changedFrds: []string{"docs/frd/keyword_interaction_matrix.md"}},
		},
		{
			name:  "ability survey",
			paths: []string{"docs/design/m2_starter_set_ability_survey.md"},
			want:  lintSelection{fences: true},
		},
		{
			name:  "toolchain and readme",
			paths: []string{"pkg/toolchain/x.go", "README.md"},
			want:  lintSelection{},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := selectChanged(c.paths, nil)
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("selectChanged(%v, nil) = %+v, want %+v", c.paths, got, c.want)
			}
		})
	}
}

func TestSelectChanged_Case2_BothModes(t *testing.T) {
	root := t.TempDir()
	out, err := exec.Command("git", "-C", root, "init").CombinedOutput()
	if err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}

	for _, rel := range []string{"docs/frd/02_x.md", "data/cards/units.csv", "pkg/toolchain/x.go"} {
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("content\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	out, err = exec.Command("git", "-C", root, "add", ".").CombinedOutput()
	if err != nil {
		t.Fatalf("git add: %v\n%s", err, out)
	}

	want := lintSelection{
		frd:         true,
		sheets:      true,
		syntax:      true,
		fences:      true,
		changedFrds: []string{"docs/frd/02_x.md"},
		changedCsvs: []string{"data/cards/units.csv"},
	}

	for _, staged := range []bool{true, false} {
		paths, err := changedFiles(root, staged)
		if err != nil {
			t.Fatalf("changedFiles(root, %v) error: %v", staged, err)
		}
		got := selectChanged(paths, nil)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("staged=%v: selectChanged = %+v, want %+v", staged, got, want)
		}
	}
}

func TestSelectChanged_Case3_GitFailure(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing")
	for _, staged := range []bool{true, false} {
		paths, err := changedFiles(missing, staged)
		if err == nil {
			t.Fatalf("changedFiles(%q, %v) expected error, got nil", missing, staged)
		}
		got := selectChanged(paths, err)
		if !reflect.DeepEqual(got, allLinters()) {
			t.Errorf("staged=%v: selectChanged(paths, err) = %+v, want %+v", staged, got, allLinters())
		}
	}

	got := selectChanged([]string{"README.md"}, errors.New("git failed"))
	want := lintSelection{
		frd:    true,
		rules:  true,
		sheets: true,
		syntax: true,
		fences: true,
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("selectChanged with error = %+v, want %+v", got, want)
	}
	if got.empty() {
		t.Errorf("got.empty() = true, want false")
	}
}
