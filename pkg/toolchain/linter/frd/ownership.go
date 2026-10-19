package frd

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

const (
	// RulesFRD owns every game rule and the keyword dictionary.
	RulesFRD = "01_Game_Rules_FRD.md"
	// KeywordMatrix quotes keyword rules on purpose to pose questions about them.
	KeywordMatrix = "keyword_interaction_matrix.md"
	// frozenMarker identifies FRDs frozen as design intent.
	frozenMarker = "**Status: FROZEN"
)

var (
	ownerRegex      = regexp.MustCompile(`^(?:#### |- \*\*)(FR-[A-Z]{3}-\d{3}[a-z]?)`)
	keywordDictLine = regexp.MustCompile(`^- \*\*FR-KWD-\d{3}[a-z]: ([A-Za-z ]+?)(?: N)?\*\*:`)
)

// CheckOwnership enforces that each requirement has exactly one owning FRD
// and that keyword behavior is defined only in FRD 01. contents maps FRD file
// name to lines. It returns the violations and how many keyword restatements
// it skipped because they sit in frozen FRDs.
func CheckOwnership(contents map[string][]string) (problems []string, frozenRestatements int) {
	names := make([]string, 0, len(contents))
	for name := range contents {
		names = append(names, name)
	}
	sort.Strings(names)

	// A requirement is owned by the file whose heading or dictionary entry
	// leads with its ID. Mentions elsewhere are references, not ownership.
	owners := make(map[string][]string)
	for _, name := range names {
		seen := make(map[string]bool)
		for _, line := range contents[name] {
			if m := ownerRegex.FindStringSubmatch(line); m != nil && !seen[m[1]] {
				seen[m[1]] = true
				owners[m[1]] = append(owners[m[1]], name)
			}
		}
	}
	ids := make([]string, 0, len(owners))
	for id := range owners {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		if files := owners[id]; len(files) > 1 {
			problems = append(problems, fmt.Sprintf("%s is defined in more than one FRD: %s", id, strings.Join(files, ", ")))
		}
	}

	restatement := keywordRestatementRegex(contents[RulesFRD])
	if restatement == nil {
		return problems, 0
	}
	for _, name := range names {
		if name == RulesFRD || name == KeywordMatrix {
			continue
		}
		frozen := false
		for _, line := range contents[name] {
			if strings.Contains(line, frozenMarker) {
				frozen = true
				break
			}
		}
		for i, line := range contents[name] {
			m := restatement.FindStringSubmatch(line)
			if m == nil {
				continue
			}
			if frozen {
				frozenRestatements++
				continue
			}
			problems = append(problems, fmt.Sprintf("%s:%d restates keyword %q; define keywords only in %s Appendix A and cite their FR-KWD ID", name, i+1, m[1], RulesFRD))
		}
	}
	return problems, frozenRestatements
}

// keywordRestatementRegex matches a bolded keyword name used as a definition
// term ("**Berserk N**:" or "| **Berserk N** |"), with names taken from FRD 01.
func keywordRestatementRegex(rules []string) *regexp.Regexp {
	var names []string
	for _, line := range rules {
		if m := keywordDictLine.FindStringSubmatch(line); m != nil {
			names = append(names, regexp.QuoteMeta(strings.TrimSpace(m[1])))
		}
	}
	if len(names) == 0 {
		return nil
	}
	return regexp.MustCompile(`(?:^|\| )\*\*(` + strings.Join(names, "|") + `)(?: N)?\*\* ?(?:\||:)`)
}
