package engine

import (
	"slices"
	"testing"
)

// Directive #85 (c): each standalone keyword, alone, prints one line, and
// parsing that line gives the same Keywords.
func TestKeywords_LinesRoundTrip(t *testing.T) {
	cases := []struct {
		kw   Keywords
		line string
	}{
		{Keywords{Resistant: 2}, "Resistant 2"},
		{Keywords{Stalwart: true}, "Stalwart"},
		{Keywords{Aggressive: 2}, "Aggressive 2"},
		{Keywords{Ethereal: true}, "Ethereal"},
		{Keywords{Ward: true}, "Ward"},
		{Keywords{Versatile: true}, "Versatile"},
		{Keywords{Guardian: true}, "Guardian"},
		{Keywords{Armor: 2}, "Armor 2"},
		{Keywords{DeploysExhausted: true}, "Deploys Exhausted"},
	}
	for _, c := range cases {
		if got := c.kw.Lines(); !slices.Equal(got, []string{c.line}) {
			t.Errorf("%+v.Lines() = %q, want [%q]", c.kw, got, c.line)
		}
		var parsed Keywords
		if !ParseKeywordLine(c.line, &parsed) || parsed != c.kw {
			t.Errorf("ParseKeywordLine(%q) = %+v, want %+v", c.line, parsed, c.kw)
		}
	}
}

// Lines is never nil, so a keywordless card's view encodes [] rather than null.
func TestKeywords_LinesEmpty(t *testing.T) {
	if got := (Keywords{}).Lines(); got == nil || len(got) != 0 {
		t.Errorf("Keywords{}.Lines() = %#v, want an empty non-nil slice", got)
	}
}
