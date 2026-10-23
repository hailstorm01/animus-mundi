package engine

import "testing"

// Tests for T-0045: Decision.Valid is the kernel's answer rule, exported for clients (FR-MOD-001).

func TestDecisionValid(t *testing.T) {
	opts := []Option{{Label: "a"}, {Label: "b"}, {Label: "c"}}
	tests := []struct {
		name    string
		min     int
		max     int
		repeats bool
		picks   []int
		want    bool
	}{
		{"1-2 empty", 1, 2, false, []int{}, false},
		{"1-2 one", 1, 2, false, []int{0}, true},
		{"1-2 two", 1, 2, false, []int{0, 2}, true},
		{"1-2 three", 1, 2, false, []int{0, 1, 2}, false},
		{"1-2 out of range", 1, 2, false, []int{3}, false},
		{"1-2 negative", 1, 2, false, []int{-1}, false},
		{"2-2 repeat no repeats", 2, 2, false, []int{1, 1}, false},
		{"2-2 repeat allowed", 2, 2, true, []int{1, 1}, true},
		{"0-1 empty", 0, 1, false, []int{}, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			d := Decision{Kind: DecideTarget, Options: opts, Min: tc.min, Max: tc.max, Repeats: tc.repeats}
			if got := d.Valid(tc.picks); got != tc.want {
				t.Errorf("Valid(%v) = %v, want %v", tc.picks, got, tc.want)
			}
		})
	}
}
