package engine

import (
	"animus-mundi/pkg/engine/ability"
	"slices"
	"testing"
)

// Directive #85 (c): each standalone keyword, alone, prints one line, and
// parsing that line gives the same Keywords.
func TestKeywords_LinesRoundTrip(t *testing.T) {
	cases := []struct {
		kw   ability.Keywords
		line string
	}{
		{ability.Keywords{Resistant: 2}, "Resistant 2"},
		{ability.Keywords{Stalwart: true}, "Stalwart"},
		{ability.Keywords{Aggressive: 2}, "Aggressive 2"},
		{ability.Keywords{Ethereal: true}, "Ethereal"},
		{ability.Keywords{Ward: true}, "Ward"},
		{ability.Keywords{Versatile: true}, "Versatile"},
		{ability.Keywords{Guardian: true}, "Guardian"},
		{ability.Keywords{Armor: 2}, "Armor 2"},
		{ability.Keywords{DeploysExhausted: true}, "Deploys Exhausted"},
	}
	for _, c := range cases {
		if got := c.kw.Lines(); !slices.Equal(got, []string{c.line}) {
			t.Errorf("%+v.Lines() = %q, want [%q]", c.kw, got, c.line)
		}
		var parsed ability.Keywords
		if !ability.ParseKeywordLine(c.line, &parsed) || parsed != c.kw {
			t.Errorf("ParseKeywordLine(%q) = %+v, want %+v", c.line, parsed, c.kw)
		}
	}
}

// Lines is never nil, so a keywordless card's view encodes [] rather than null.
func TestKeywords_LinesEmpty(t *testing.T) {
	if got := (ability.Keywords{}).Lines(); got == nil || len(got) != 0 {
		t.Errorf("Keywords{}.Lines() = %#v, want an empty non-nil slice", got)
	}
}
