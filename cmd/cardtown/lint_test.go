package main

import "testing"

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
